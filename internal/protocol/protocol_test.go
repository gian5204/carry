package protocol

import (
	"bytes"
	"errors"
	"net"
	"strings"
	"testing"
	"time"
)

func TestWriteMessage(t *testing.T) {
	tests := []struct {
		name    string
		message Message
		want    string
	}{
		{
			name:    "hello",
			message: Message{Type: TypeHello, ProtocolVersion: ProtocolVersion},
			want:    "{\"type\":\"hello\",\"protocolVersion\":1}\n",
		},
		{
			name:    "acknowledgment",
			message: Message{Type: TypeAck, ProtocolVersion: ProtocolVersion},
			want:    "{\"type\":\"ack\",\"protocolVersion\":1}\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			if err := writeMessage(&output, test.message); err != nil {
				t.Fatalf("writeMessage() error = %v", err)
			}
			if output.String() != test.want {
				t.Errorf("writeMessage() = %q; want %q", output.String(), test.want)
			}
		})
	}
}

func TestReadMessage(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Message
	}{
		{
			name:  "hello",
			input: "{\"type\":\"hello\",\"protocolVersion\":1}\n",
			want:  Message{Type: TypeHello, ProtocolVersion: ProtocolVersion},
		},
		{
			name:  "acknowledgment",
			input: "{\"type\":\"ack\",\"protocolVersion\":1}\n",
			want:  Message{Type: TypeAck, ProtocolVersion: ProtocolVersion},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := readMessage(strings.NewReader(test.input))
			if err != nil {
				t.Fatalf("readMessage() error = %v", err)
			}
			if got != test.want {
				t.Errorf("readMessage() = %+v; want %+v", got, test.want)
			}
		})
	}
}

func TestReadMessageRejectsInvalidMessages(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError string
	}{
		{
			name:      "malformed JSON",
			input:     "{not json}\n",
			wantError: "decode protocol message",
		},
		{
			name:      "missing message type",
			input:     "{\"protocolVersion\":1}\n",
			wantError: "message type is required",
		},
		{
			name:      "unknown message type",
			input:     "{\"type\":\"unknown\",\"protocolVersion\":1}\n",
			wantError: "unknown message type",
		},
		{
			name:      "missing protocol version",
			input:     "{\"type\":\"hello\"}\n",
			wantError: "protocol version is required",
		},
		{
			name:      "invalid protocol version",
			input:     "{\"type\":\"hello\",\"protocolVersion\":-1}\n",
			wantError: "invalid protocol version",
		},
		{
			name:      "unsupported protocol version",
			input:     "{\"type\":\"hello\",\"protocolVersion\":2}\n",
			wantError: "unsupported protocol version 2",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := readMessage(strings.NewReader(test.input))
			if err == nil {
				t.Fatalf("readMessage() error = nil; want error containing %q", test.wantError)
			}
			if !strings.Contains(err.Error(), test.wantError) {
				t.Errorf("readMessage() error = %q; want error containing %q", err, test.wantError)
			}
		})
	}
}

func TestExpectMessageTypeRejectsUnexpectedType(t *testing.T) {
	message := Message{Type: TypeAck, ProtocolVersion: ProtocolVersion}
	err := expectMessageType(message, TypeHello)
	if err == nil {
		t.Fatal("expectMessageType() error = nil; want unexpected type error")
	}
	if !strings.Contains(err.Error(), "unexpected message type") {
		t.Errorf("expectMessageType() error = %q; want unexpected type error", err)
	}
}

func TestSenderAndReceiverHandshake(t *testing.T) {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}
	defer listener.Close()

	tcpListener, ok := listener.(*net.TCPListener)
	if !ok {
		t.Fatalf("listener type = %T; want *net.TCPListener", listener)
	}
	if err := tcpListener.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("SetDeadline() error = %v", err)
	}

	receiverResult := make(chan error, 1)
	go func() {
		connection, err := listener.Accept()
		if err == nil {
			err = ReceiveHandshake(connection)
			if err == nil {
				err = ReceiveRepositoryVerification(connection, "matching-repository")
			}
			connection.Close()
		}
		receiverResult <- err
	}()

	connection, err := net.DialTimeout("tcp4", listener.Addr().String(), 5*time.Second)
	if err != nil {
		t.Fatalf("net.DialTimeout() error = %v", err)
	}
	defer connection.Close()

	if err := SendHandshake(connection); err != nil {
		t.Fatalf("SendHandshake() error = %v", err)
	}
	if err := SendRepositoryVerification(connection, "matching-repository"); err != nil {
		t.Fatalf("SendRepositoryVerification() error = %v", err)
	}

	select {
	case err := <-receiverResult:
		if err != nil {
			t.Fatalf("ReceiveHandshake() error = %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("ReceiveHandshake() did not return")
	}
}

func TestRepositoryVerificationRejectsMismatch(t *testing.T) {
	sender, receiver := net.Pipe()
	defer sender.Close()
	defer receiver.Close()

	receiverResult := make(chan error, 1)
	go func() {
		receiverResult <- ReceiveRepositoryVerification(
			receiver,
			"receiver-repository",
		)
	}()

	err := SendRepositoryVerification(sender, "sender-repository")
	if !errors.Is(err, ErrRepositoryMismatch) {
		t.Fatalf("SendRepositoryVerification() error = %v; want repository mismatch", err)
	}

	select {
	case err := <-receiverResult:
		if !errors.Is(err, ErrRepositoryMismatch) {
			t.Fatalf("ReceiveRepositoryVerification() error = %v; want repository mismatch", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("ReceiveRepositoryVerification() did not return")
	}
}

func TestRepositoryVerificationRejectsMalformedMessage(t *testing.T) {
	response, receiverErr := repositoryVerificationRejection(
		t,
		"{\"type\":\"repository\",\"protocolVersion\":1}\n",
	)

	if response.Type != TypeReject {
		t.Errorf("rejection type = %q; want %q", response.Type, TypeReject)
	}
	if response.Error != "invalid repository message" {
		t.Errorf("rejection reason = %q; want %q", response.Error, "invalid repository message")
	}
	if !strings.Contains(receiverErr.Error(), "repository identity is required") {
		t.Errorf("receiver error = %q; want missing identity error", receiverErr)
	}
}

func TestRepositoryVerificationRejectsUnexpectedMessage(t *testing.T) {
	response, receiverErr := repositoryVerificationRejection(
		t,
		"{\"type\":\"hello\",\"protocolVersion\":1}\n",
	)

	if response.Type != TypeReject {
		t.Errorf("rejection type = %q; want %q", response.Type, TypeReject)
	}
	if response.Error != "unexpected repository message" {
		t.Errorf("rejection reason = %q; want %q", response.Error, "unexpected repository message")
	}
	if !strings.Contains(receiverErr.Error(), "unexpected message type") {
		t.Errorf("receiver error = %q; want unexpected type error", receiverErr)
	}
}

func TestRepositoryVerificationPeerDisconnects(t *testing.T) {
	receiver, peer := net.Pipe()
	defer receiver.Close()

	result := make(chan error, 1)
	go func() {
		result <- ReceiveRepositoryVerification(receiver, "repository")
	}()

	if _, err := peer.Write([]byte("{\"type\":\"repository\"")); err != nil {
		t.Fatalf("peer.Write() error = %v", err)
	}
	if err := peer.Close(); err != nil {
		t.Fatalf("peer.Close() error = %v", err)
	}

	select {
	case err := <-result:
		if err == nil {
			t.Fatal("ReceiveRepositoryVerification() error = nil; want disconnect error")
		}
		if !strings.Contains(err.Error(), "read repository identity") {
			t.Errorf("ReceiveRepositoryVerification() error = %q; want read error", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("ReceiveRepositoryVerification() did not return")
	}
}

func repositoryVerificationRejection(
	t *testing.T,
	request string,
) (Message, error) {
	t.Helper()

	receiver, peer := net.Pipe()
	defer receiver.Close()
	defer peer.Close()

	if err := peer.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("peer.SetDeadline() error = %v", err)
	}

	result := make(chan error, 1)
	go func() {
		result <- ReceiveRepositoryVerification(receiver, "repository")
	}()

	if _, err := peer.Write([]byte(request)); err != nil {
		t.Fatalf("peer.Write() error = %v", err)
	}
	response, err := readMessage(peer)
	if err != nil {
		t.Fatalf("readMessage(rejection) error = %v", err)
	}

	select {
	case err := <-result:
		if err == nil {
			t.Fatal("ReceiveRepositoryVerification() error = nil; want rejection error")
		}
		return response, err
	case <-time.After(5 * time.Second):
		t.Fatal("ReceiveRepositoryVerification() did not return")
		return Message{}, nil
	}
}

func TestReceiveHandshakePeerDisconnectsEarly(t *testing.T) {
	receiver, peer := net.Pipe()
	defer receiver.Close()

	result := make(chan error, 1)
	go func() {
		result <- ReceiveHandshake(receiver)
	}()

	if _, err := peer.Write([]byte("{\"type\":\"hello\"")); err != nil {
		t.Fatalf("peer.Write() error = %v", err)
	}
	if err := peer.Close(); err != nil {
		t.Fatalf("peer.Close() error = %v", err)
	}

	select {
	case err := <-result:
		if err == nil {
			t.Fatal("ReceiveHandshake() error = nil; want incomplete message error")
		}
		if !strings.Contains(err.Error(), "decode protocol message") {
			t.Errorf("ReceiveHandshake() error = %q; want decode error", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("ReceiveHandshake() did not return")
	}
}
