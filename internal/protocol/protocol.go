package protocol

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/gian5204/carry/internal/managedpath"
)

const ProtocolVersion = 1

const (
	exchangeTimeout            = 5 * time.Second
	fileTransferTimeout        = 30 * time.Second
	maxProtocolFrameSize       = 1 << 20
	MaxFileSize          int64 = 10 << 20
)

var ErrRepositoryMismatch = errors.New("repository mismatch")

type MessageType string

const (
	TypeHello           MessageType = "hello"
	TypeAck             MessageType = "ack"
	TypeRepository      MessageType = "repository"
	TypeReject          MessageType = "reject"
	TypeManagedFiles    MessageType = "managed_files"
	TypeManagedFilesAck MessageType = "managed_files_ack"
	TypeFileBegin       MessageType = "file_begin"
	TypeFileReady       MessageType = "file_ready"
	TypeFileAck         MessageType = "file_ack"
)

type Message struct {
	Type            MessageType `json:"type"`
	ProtocolVersion int         `json:"protocolVersion"`
	RepositoryID    string      `json:"repositoryId,omitempty"`
	Error           string      `json:"error,omitempty"`
	Files           []string    `json:"files,omitempty"`
	Path            string      `json:"path,omitempty"`
	Size            *int64      `json:"size,omitempty"`
}

type FileDestination interface {
	io.Writer
	Commit() error
	Abort() error
}

type PrepareFile func(path string, size int64) (FileDestination, error)

type RejectionError struct {
	Reason string
	Err    error
}

func (e *RejectionError) Error() string { return e.Err.Error() }
func (e *RejectionError) Unwrap() error { return e.Err }

func Reject(reason string, err error) error {
	return &RejectionError{Reason: reason, Err: err}
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
		return rejectProtocolMessage(
			connection,
			"invalid repository message",
			fmt.Errorf("read repository identity: %w", err),
		)
	}
	if err := expectMessageType(request, TypeRepository); err != nil {
		return rejectProtocolMessage(
			connection,
			"unexpected repository message",
			fmt.Errorf("validate repository identity: %w", err),
		)
	}
	if request.RepositoryID != repositoryID {
		return rejectProtocolMessage(
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

func SendManagedFiles(connection net.Conn, files []string) error {
	if err := setExchangeDeadline(connection); err != nil {
		return err
	}
	defer connection.SetDeadline(time.Time{})

	request := Message{
		Type:            TypeManagedFiles,
		ProtocolVersion: ProtocolVersion,
		Files:           files,
	}
	if err := writeMessage(connection, request); err != nil {
		return fmt.Errorf("send managed-file metadata: %w", err)
	}

	response, err := readMessage(connection)
	if err != nil {
		return fmt.Errorf("read managed-file acknowledgment: %w", err)
	}
	if response.Type == TypeReject {
		return fmt.Errorf("managed-file metadata rejected: %s", response.Error)
	}
	if err := expectMessageType(response, TypeManagedFilesAck); err != nil {
		return fmt.Errorf("validate managed-file acknowledgment: %w", err)
	}

	return nil
}

func ReceiveManagedFiles(connection net.Conn) ([]string, error) {
	if err := setExchangeDeadline(connection); err != nil {
		return nil, err
	}
	defer connection.SetDeadline(time.Time{})

	request, err := readMessage(connection)
	if err != nil {
		return nil, rejectProtocolMessage(
			connection,
			"invalid managed-file metadata",
			fmt.Errorf("read managed-file metadata: %w", err),
		)
	}
	if err := expectMessageType(request, TypeManagedFiles); err != nil {
		return nil, rejectProtocolMessage(
			connection,
			"unexpected managed-file message",
			fmt.Errorf("validate managed-file metadata: %w", err),
		)
	}

	files, err := managedpath.NormalizeAll(request.Files)
	if err != nil {
		return nil, rejectProtocolMessage(
			connection,
			"invalid managed-file metadata",
			fmt.Errorf("validate managed-file metadata: %w", err),
		)
	}

	acknowledgment := Message{
		Type:            TypeManagedFilesAck,
		ProtocolVersion: ProtocolVersion,
	}
	if err := writeMessage(connection, acknowledgment); err != nil {
		return nil, fmt.Errorf("send managed-file acknowledgment: %w", err)
	}

	return files, nil
}

func SendFile(
	connection net.Conn,
	path string,
	size int64,
	source io.Reader,
) error {
	if err := setFileTransferDeadline(connection); err != nil {
		return err
	}
	defer connection.SetDeadline(time.Time{})

	normalizedPath, err := managedpath.Normalize(path)
	if err != nil {
		return fmt.Errorf("validate file path: %w", err)
	}

	begin := Message{
		Type:            TypeFileBegin,
		ProtocolVersion: ProtocolVersion,
		Path:            normalizedPath,
		Size:            &size,
	}
	if err := writeMessage(connection, begin); err != nil {
		return fmt.Errorf("send file metadata: %w", err)
	}

	response, err := readMessage(connection)
	if err != nil {
		return fmt.Errorf("read file readiness: %w", err)
	}
	if response.Type == TypeReject {
		return fmt.Errorf("file transfer rejected: %s", response.Error)
	}
	if err := expectFileMessage(response, TypeFileReady, normalizedPath); err != nil {
		return fmt.Errorf("validate file readiness: %w", err)
	}

	written, err := io.CopyN(connection, source, size)
	if err != nil {
		return fmt.Errorf("send file payload after %d bytes: %w", written, err)
	}

	response, err = readMessage(connection)
	if err != nil {
		return fmt.Errorf("read file acknowledgment: %w", err)
	}
	if response.Type == TypeReject {
		return fmt.Errorf("file transfer rejected: %s", response.Error)
	}
	if err := expectFileMessage(response, TypeFileAck, normalizedPath); err != nil {
		return fmt.Errorf("validate file acknowledgment: %w", err)
	}

	return nil
}

func ReceiveFile(
	connection net.Conn,
	acceptedFiles []string,
	prepare PrepareFile,
) (returnErr error) {
	if err := setFileTransferDeadline(connection); err != nil {
		return err
	}
	defer connection.SetDeadline(time.Time{})

	reader := bufio.NewReaderSize(connection, maxProtocolFrameSize)
	begin, err := readFramedMessage(reader)
	if err != nil {
		return rejectProtocolMessage(
			connection,
			"invalid file metadata",
			fmt.Errorf("read file metadata: %w", err),
		)
	}
	if err := expectMessageType(begin, TypeFileBegin); err != nil {
		return rejectProtocolMessage(
			connection,
			"unexpected file message",
			fmt.Errorf("validate file metadata: %w", err),
		)
	}

	path, err := managedpath.Normalize(begin.Path)
	if err != nil {
		return rejectProtocolMessage(
			connection,
			"invalid file path",
			fmt.Errorf("validate file path: %w", err),
		)
	}
	if !containsManagedPath(acceptedFiles, path) {
		return rejectProtocolMessage(
			connection,
			"file path was not accepted",
			fmt.Errorf("file path was not part of accepted metadata"),
		)
	}

	destination, err := prepare(path, *begin.Size)
	if err != nil {
		reason := "receiver filesystem failure"
		var rejection *RejectionError
		if errors.As(err, &rejection) {
			reason = rejection.Reason
		}
		return rejectProtocolMessage(connection, reason, err)
	}
	defer func() {
		if err := destination.Abort(); err != nil {
			returnErr = errors.Join(
				returnErr,
				fmt.Errorf("clean up temporary file: %w", err),
			)
		}
	}()

	ready := Message{
		Type:            TypeFileReady,
		ProtocolVersion: ProtocolVersion,
		Path:            path,
	}
	if err := writeMessage(connection, ready); err != nil {
		return fmt.Errorf("send file readiness: %w", err)
	}

	written, err := io.CopyN(destination, reader, *begin.Size)
	if err != nil {
		return rejectProtocolMessage(
			connection,
			"incomplete file payload",
			fmt.Errorf("receive file payload after %d bytes: %w", written, err),
		)
	}
	if err := destination.Commit(); err != nil {
		reason := "receiver filesystem failure"
		var rejection *RejectionError
		if errors.As(err, &rejection) {
			reason = rejection.Reason
		}
		return rejectProtocolMessage(connection, reason, err)
	}

	acknowledgment := Message{
		Type:            TypeFileAck,
		ProtocolVersion: ProtocolVersion,
		Path:            path,
	}
	if err := writeMessage(connection, acknowledgment); err != nil {
		return fmt.Errorf("send file acknowledgment: %w", err)
	}

	return nil
}

func containsManagedPath(acceptedFiles []string, path string) bool {
	for _, accepted := range acceptedFiles {
		normalized, err := managedpath.Normalize(accepted)
		if err == nil && strings.EqualFold(normalized, path) {
			return true
		}
	}
	return false
}

func expectFileMessage(message Message, expected MessageType, path string) error {
	if err := expectMessageType(message, expected); err != nil {
		return err
	}
	normalized, err := managedpath.Normalize(message.Path)
	if err != nil {
		return err
	}
	if normalized != path {
		return fmt.Errorf("unexpected file path %q: expected %q", normalized, path)
	}
	return nil
}

func rejectProtocolMessage(
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

func setFileTransferDeadline(connection net.Conn) error {
	if err := connection.SetDeadline(time.Now().Add(fileTransferTimeout)); err != nil {
		return fmt.Errorf("set file-transfer deadline: %w", err)
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
	return decodeMessage(input)
}

func readFramedMessage(input *bufio.Reader) (Message, error) {
	frame, err := input.ReadSlice('\n')
	if err != nil {
		return Message{}, fmt.Errorf("read protocol frame: %w", err)
	}
	return decodeMessage(bytes.NewReader(frame))
}

func decodeMessage(input io.Reader) (Message, error) {
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
	case TypeHello,
		TypeAck,
		TypeRepository,
		TypeReject,
		TypeManagedFiles,
		TypeManagedFilesAck,
		TypeFileBegin,
		TypeFileReady,
		TypeFileAck:
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
	case TypeManagedFiles:
		if _, err := managedpath.NormalizeAll(message.Files); err != nil {
			return err
		}
	case TypeFileBegin:
		if _, err := managedpath.Normalize(message.Path); err != nil {
			return err
		}
		if message.Size == nil {
			return fmt.Errorf("file size is required")
		}
		if *message.Size < 0 {
			return fmt.Errorf("file size cannot be negative")
		}
		if *message.Size > MaxFileSize {
			return fmt.Errorf("file size exceeds maximum of %d bytes", MaxFileSize)
		}
	case TypeFileReady, TypeFileAck:
		if _, err := managedpath.Normalize(message.Path); err != nil {
			return err
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
