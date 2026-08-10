package protocol

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"
)

const ProtocolVersion = 1

const handshakeTimeout = 5 * time.Second

type MessageType string

const (
	TypeHello MessageType = "hello"
	TypeAck   MessageType = "ack"
)

type Message struct {
	Type            MessageType `json:"type"`
	ProtocolVersion int         `json:"protocolVersion"`
}

func SendHandshake(connection net.Conn) error {
	if err := setHandshakeDeadline(connection); err != nil {
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
	if err := setHandshakeDeadline(connection); err != nil {
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

func setHandshakeDeadline(connection net.Conn) error {
	if err := connection.SetDeadline(time.Now().Add(handshakeTimeout)); err != nil {
		return fmt.Errorf("set handshake deadline: %w", err)
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
	case TypeHello, TypeAck:
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
