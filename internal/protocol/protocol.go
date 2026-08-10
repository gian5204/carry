package protocol

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

const ProtocolVersion = 1

const exchangeTimeout = 5 * time.Second

var ErrRepositoryMismatch = errors.New("repository mismatch")

type MessageType string

const (
	TypeHello      MessageType = "hello"
	TypeAck        MessageType = "ack"
	TypeRepository MessageType = "repository"
	TypeReject     MessageType = "reject"
)

type Message struct {
	Type            MessageType `json:"type"`
	ProtocolVersion int         `json:"protocolVersion"`
	RepositoryID    string      `json:"repositoryId,omitempty"`
	Error           string      `json:"error,omitempty"`
}

func SendHandshake(connection net.Conn) error {
	if err := setExchangeDeadline(connection); err != nil {
		return err
	}
	defer connection.SetDeadline(time.Time{})

	hello := Message{Type: TypeHello, ProtocolVersion: ProtocolVersion}
	if err := writeMessage(connection, hello); err != nil {
		return fmt.Errorf("send hello: %w", err)
	}

	response, err := readMessage(connection)
	if err != nil {
		return fmt.Errorf("read acknowledgment: %w", err)
	}
	if err := expectMessageType(response, TypeAck); err != nil {
		return fmt.Errorf("validate acknowledgment: %w", err)
	}

	return nil
}

func ReceiveHandshake(connection net.Conn) error {
	if err := setExchangeDeadline(connection); err != nil {
		return err
	}
	defer connection.SetDeadline(time.Time{})

	request, err := readMessage(connection)
	if err != nil {
		return fmt.Errorf("read hello: %w", err)
	}
	if err := expectMessageType(request, TypeHello); err != nil {
		return fmt.Errorf("validate hello: %w", err)
	}

	acknowledgment := Message{Type: TypeAck, ProtocolVersion: ProtocolVersion}
	if err := writeMessage(connection, acknowledgment); err != nil {
		return fmt.Errorf("send acknowledgment: %w", err)
	}

	return nil
}

func SendRepositoryVerification(connection net.Conn, repositoryID string) error {
	if err := setExchangeDeadline(connection); err != nil {
		return err
	}
	defer connection.SetDeadline(time.Time{})

	request := Message{
		Type:            TypeRepository,
		ProtocolVersion: ProtocolVersion,
		RepositoryID:    repositoryID,
	}
	if err := writeMessage(connection, request); err != nil {
		return fmt.Errorf("send repository identity: %w", err)
	}

	response, err := readMessage(connection)
	if err != nil {
		return fmt.Errorf("read repository verification: %w", err)
	}
	if response.Type == TypeReject {
		if response.Error == ErrRepositoryMismatch.Error() {
			return ErrRepositoryMismatch
		}
		return fmt.Errorf("repository verification rejected: %s", response.Error)
	}
	if err := expectMessageType(response, TypeRepository); err != nil {
		return fmt.Errorf("validate repository response: %w", err)
	}
	if response.RepositoryID != repositoryID {
		return ErrRepositoryMismatch
	}

	return nil
}

func ReceiveRepositoryVerification(connection net.Conn, repositoryID string) error {
	if err := setExchangeDeadline(connection); err != nil {
		return err
	}
	defer connection.SetDeadline(time.Time{})

	request, err := readMessage(connection)
	if err != nil {
		return rejectRepositoryVerification(
			connection,
			"invalid repository message",
			fmt.Errorf("read repository identity: %w", err),
		)
	}
	if err := expectMessageType(request, TypeRepository); err != nil {
		return rejectRepositoryVerification(
			connection,
			"unexpected repository message",
			fmt.Errorf("validate repository identity: %w", err),
		)
	}
	if request.RepositoryID != repositoryID {
		return rejectRepositoryVerification(
			connection,
			ErrRepositoryMismatch.Error(),
			ErrRepositoryMismatch,
		)
	}

	response := Message{
		Type:            TypeRepository,
		ProtocolVersion: ProtocolVersion,
		RepositoryID:    repositoryID,
	}
	if err := writeMessage(connection, response); err != nil {
		return fmt.Errorf("send repository identity: %w", err)
	}

	return nil
}

func rejectRepositoryVerification(
	connection net.Conn,
	reason string,
	cause error,
) error {
	rejection := Message{
		Type:            TypeReject,
		ProtocolVersion: ProtocolVersion,
		Error:           reason,
	}
	if err := writeMessage(connection, rejection); err != nil {
		return fmt.Errorf("%w (send rejection: %v)", cause, err)
	}
	return cause
}

func setExchangeDeadline(connection net.Conn) error {
	if err := connection.SetDeadline(time.Now().Add(exchangeTimeout)); err != nil {
		return fmt.Errorf("set protocol deadline: %w", err)
	}
	return nil
}

func writeMessage(output io.Writer, message Message) error {
	if err := validateMessage(message); err != nil {
		return err
	}
	if err := json.NewEncoder(output).Encode(message); err != nil {
		return fmt.Errorf("encode protocol message: %w", err)
	}
	return nil
}

func readMessage(input io.Reader) (Message, error) {
	var message Message
	if err := json.NewDecoder(input).Decode(&message); err != nil {
		return Message{}, fmt.Errorf("decode protocol message: %w", err)
	}
	if err := validateMessage(message); err != nil {
		return Message{}, err
	}
	return message, nil
}

func validateMessage(message Message) error {
	switch message.Type {
	case "":
		return fmt.Errorf("message type is required")
	case TypeHello, TypeAck, TypeRepository, TypeReject:
	default:
		return fmt.Errorf("unknown message type %q", message.Type)
	}

	switch {
	case message.ProtocolVersion == 0:
		return fmt.Errorf("protocol version is required")
	case message.ProtocolVersion < 0:
		return fmt.Errorf("invalid protocol version %d", message.ProtocolVersion)
	case message.ProtocolVersion != ProtocolVersion:
		return fmt.Errorf("unsupported protocol version %d", message.ProtocolVersion)
	}

	switch message.Type {
	case TypeRepository:
		if message.RepositoryID == "" {
			return fmt.Errorf("repository identity is required")
		}
	case TypeReject:
		if message.Error == "" {
			return fmt.Errorf("rejection reason is required")
		}
	}

	return nil
}

func expectMessageType(message Message, expected MessageType) error {
	if message.Type != expected {
		return fmt.Errorf(
			"unexpected message type %q: expected %q",
			message.Type,
			expected,
		)
	}
	return nil
}
