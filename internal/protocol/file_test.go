package protocol

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gian5204/carry/internal/filetransfer"
)

type memoryDestination struct {
	bytes.Buffer
	committed bool
	aborted   bool
}

func (d *memoryDestination) Commit() error {
	d.committed = true
	return nil
}

func (d *memoryDestination) Abort() error {
	d.aborted = true
	return nil
}

func TestFileTransfer(t *testing.T) {
	sender, receiver := net.Pipe()
	defer sender.Close()
	defer receiver.Close()

	destination := &memoryDestination{}
	receiverResult := make(chan error, 1)
	go func() {
		receiverResult <- ReceiveFile(
			receiver,
			[]string{".env", "config/local.env"},
			func(path string, size int64) (FileDestination, error) {
				if path != "config/local.env" {
					return nil, fmt.Errorf("unexpected path %q", path)
				}
				if size != int64(len("secret contents")) {
					return nil, fmt.Errorf("unexpected size %d", size)
				}
				return destination, nil
			},
		)
	}()

	contents := "secret contents"
	if err := SendFile(
		sender,
		"config/local.env",
		int64(len(contents)),
		strings.NewReader(contents),
	); err != nil {
		t.Fatalf("SendFile() error = %v", err)
	}

	if err := waitForProtocolResult(t, receiverResult); err != nil {
		t.Fatalf("ReceiveFile() error = %v", err)
	}
	if destination.String() != contents {
		t.Errorf("transferred contents = %q; want %q", destination.String(), contents)
	}
	if !destination.committed {
		t.Error("destination was not committed")
	}
}

func TestFileTransferOverTCPWritesDestination(t *testing.T) {
	sourceRoot := t.TempDir()
	sourcePath := filepath.Join(sourceRoot, "config", "local.env")
	if err := os.MkdirAll(filepath.Dir(sourcePath), 0755); err != nil {
		t.Fatalf("MkdirAll(source) error = %v", err)
	}
	contents := []byte("remote file contents")
	if err := os.WriteFile(sourcePath, contents, 0600); err != nil {
		t.Fatalf("WriteFile(source) error = %v", err)
	}
	source, size, err := filetransfer.OpenSource(sourceRoot, "config/local.env")
	if err != nil {
		t.Fatalf("OpenSource() error = %v", err)
	}
	defer source.Close()

	destinationRoot := t.TempDir()
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}
	defer listener.Close()
	tcpListener := listener.(*net.TCPListener)
	if err := tcpListener.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("SetDeadline() error = %v", err)
	}

	receiverResult := make(chan error, 1)
	go func() {
		connection, err := listener.Accept()
		if err == nil {
			err = ReceiveFile(
				connection,
				[]string{"config/local.env"},
				func(path string, _ int64) (FileDestination, error) {
					return filetransfer.PrepareDestination(destinationRoot, path)
				},
			)
			connection.Close()
		}
		receiverResult <- err
	}()

	connection, err := net.DialTimeout("tcp4", listener.Addr().String(), 5*time.Second)
	if err != nil {
		t.Fatalf("net.DialTimeout() error = %v", err)
	}
	defer connection.Close()
	if err := SendFile(connection, "config/local.env", size, source); err != nil {
		t.Fatalf("SendFile() error = %v", err)
	}
	if err := waitForProtocolResult(t, receiverResult); err != nil {
		t.Fatalf("ReceiveFile() error = %v", err)
	}

	destinationPath := filepath.Join(destinationRoot, "config", "local.env")
	got, err := os.ReadFile(destinationPath)
	if err != nil {
		t.Fatalf("ReadFile(destination) error = %v", err)
	}
	if !bytes.Equal(got, contents) {
		t.Errorf("destination contents = %q; want %q", got, contents)
	}
}

func TestFileTransferDeclaredSizeSmallerThanSource(t *testing.T) {
	sender, receiver := net.Pipe()
	defer sender.Close()
	defer receiver.Close()

	destination := &memoryDestination{}
	receiverResult := make(chan error, 1)
	go func() {
		receiverResult <- ReceiveFile(
			receiver,
			[]string{".env"},
			func(string, int64) (FileDestination, error) {
				return destination, nil
			},
		)
	}()

	if err := SendFile(sender, ".env", 3, strings.NewReader("abcdef")); err != nil {
		t.Fatalf("SendFile() error = %v", err)
	}
	if err := waitForProtocolResult(t, receiverResult); err != nil {
		t.Fatalf("ReceiveFile() error = %v", err)
	}
	if destination.String() != "abc" {
		t.Errorf("transferred contents = %q; want %q", destination.String(), "abc")
	}
}

func TestFileTransferDeclaredSizeLargerThanSource(t *testing.T) {
	sender, receiver := net.Pipe()
	destination := &memoryDestination{}
	receiverResult := make(chan error, 1)
	go func() {
		receiverResult <- ReceiveFile(
			receiver,
			[]string{".env"},
			func(string, int64) (FileDestination, error) {
				return destination, nil
			},
		)
	}()

	err := SendFile(sender, ".env", 5, strings.NewReader("abc"))
	if err == nil || !strings.Contains(err.Error(), "send file payload") {
		t.Fatalf("SendFile() error = %v; want short payload error", err)
	}
	sender.Close()

	err = waitForProtocolResult(t, receiverResult)
	receiver.Close()
	if err == nil || !strings.Contains(err.Error(), "receive file payload") {
		t.Fatalf("ReceiveFile() error = %v; want incomplete payload error", err)
	}
	if destination.committed {
		t.Error("partial destination was committed")
	}
	if !destination.aborted {
		t.Error("partial destination was not aborted")
	}
}

func TestReceiveFileDisconnectCleansTemporaryDestination(t *testing.T) {
	root := t.TempDir()
	receiver, peer := net.Pipe()

	receiverResult := make(chan error, 1)
	go func() {
		receiverResult <- ReceiveFile(
			receiver,
			[]string{"nested/.env"},
			func(path string, _ int64) (FileDestination, error) {
				return filetransfer.PrepareDestination(root, path)
			},
		)
	}()

	if _, err := peer.Write([]byte(fileBeginJSON("nested/.env", 5))); err != nil {
		t.Fatalf("Write(file_begin) error = %v", err)
	}
	ready, err := readMessage(peer)
	if err != nil {
		t.Fatalf("readMessage(file_ready) error = %v", err)
	}
	if ready.Type != TypeFileReady {
		t.Fatalf("response type = %q; want %q", ready.Type, TypeFileReady)
	}
	if _, err := peer.Write([]byte("abc")); err != nil {
		t.Fatalf("Write(payload) error = %v", err)
	}
	if err := peer.Close(); err != nil {
		t.Fatalf("peer.Close() error = %v", err)
	}

	err = waitForProtocolResult(t, receiverResult)
	receiver.Close()
	if err == nil || !strings.Contains(err.Error(), "receive file payload") {
		t.Fatalf("ReceiveFile() error = %v; want incomplete payload error", err)
	}

	final := filepath.Join(root, "nested", ".env")
	if _, err := os.Stat(final); !os.IsNotExist(err) {
		t.Fatalf("final file error = %v; want not exist", err)
	}
	temporary, err := filepath.Glob(filepath.Join(root, "nested", ".carry-transfer-*"))
	if err != nil {
		t.Fatalf("Glob() error = %v", err)
	}
	if len(temporary) != 0 {
		t.Errorf("temporary files = %v; want none", temporary)
	}
}

func TestReceiveFileRejectsInvalidMetadata(t *testing.T) {
	tests := []struct {
		name       string
		request    string
		accepted   []string
		wantReason string
	}{
		{
			name:       "malformed file begin",
			request:    "{invalid}\n",
			accepted:   []string{".env"},
			wantReason: "invalid file metadata",
		},
		{
			name:       "unexpected message",
			request:    `{"type":"ack","protocolVersion":1}` + "\n",
			accepted:   []string{".env"},
			wantReason: "unexpected file message",
		},
		{
			name:       "path not accepted",
			request:    fileBeginJSON("other.env", 1),
			accepted:   []string{".env"},
			wantReason: "file path was not accepted",
		},
		{
			name:       "unsafe path",
			request:    fileBeginJSON("../secret", 1),
			accepted:   []string{"../secret"},
			wantReason: "invalid file metadata",
		},
		{
			name:       "negative size",
			request:    fileBeginJSON(".env", -1),
			accepted:   []string{".env"},
			wantReason: "invalid file metadata",
		},
		{
			name:       "oversized",
			request:    fileBeginJSON(".env", MaxFileSize+1),
			accepted:   []string{".env"},
			wantReason: "invalid file metadata",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reason, receiverErr := receiveFileRejection(t, test.request, test.accepted, nil)
			if reason != test.wantReason {
				t.Errorf("rejection reason = %q; want %q", reason, test.wantReason)
			}
			if receiverErr == nil {
				t.Fatal("ReceiveFile() error = nil; want rejection error")
			}
		})
	}
}

func TestReceiveFileRejectsDestinationConflict(t *testing.T) {
	reason, _ := receiveFileRejection(
		t,
		fileBeginJSON(".env", 1),
		[]string{".env"},
		func(string, int64) (FileDestination, error) {
			return nil, Reject(
				"destination file already exists",
				errors.New("destination exists"),
			)
		},
	)
	if reason != "destination file already exists" {
		t.Errorf("rejection reason = %q; want destination conflict", reason)
	}
}

func TestSendFileSurfacesReceiverRejection(t *testing.T) {
	sender, peer := net.Pipe()
	defer sender.Close()
	defer peer.Close()

	peerResult := make(chan error, 1)
	go func() {
		_, err := readMessage(peer)
		if err == nil {
			err = writeMessage(peer, Message{
				Type:            TypeReject,
				ProtocolVersion: ProtocolVersion,
				Error:           "destination file already exists",
			})
		}
		peerResult <- err
	}()

	err := SendFile(sender, ".env", 1, strings.NewReader("x"))
	if err == nil || !strings.Contains(err.Error(), "destination file already exists") {
		t.Fatalf("SendFile() error = %v; want receiver rejection", err)
	}
	if err := waitForProtocolResult(t, peerResult); err != nil {
		t.Fatalf("peer error = %v", err)
	}
}

func TestSendFileValidatesAcknowledgment(t *testing.T) {
	sender, peer := net.Pipe()
	defer sender.Close()
	defer peer.Close()

	peerResult := make(chan error, 1)
	go func() {
		begin, err := readMessage(peer)
		if err == nil {
			err = writeMessage(peer, Message{
				Type:            TypeFileReady,
				ProtocolVersion: ProtocolVersion,
				Path:            begin.Path,
			})
		}
		if err == nil {
			_, err = io.CopyN(io.Discard, peer, *begin.Size)
		}
		if err == nil {
			err = writeMessage(peer, Message{
				Type:            TypeFileAck,
				ProtocolVersion: ProtocolVersion,
				Path:            begin.Path,
			})
		}
		peerResult <- err
	}()

	if err := SendFile(sender, ".env", 1, strings.NewReader("x")); err != nil {
		t.Fatalf("SendFile() error = %v", err)
	}
	if err := waitForProtocolResult(t, peerResult); err != nil {
		t.Fatalf("peer error = %v", err)
	}
}

func receiveFileRejection(
	t *testing.T,
	request string,
	accepted []string,
	prepare PrepareFile,
) (string, error) {
	t.Helper()
	if prepare == nil {
		prepare = func(string, int64) (FileDestination, error) {
			return &memoryDestination{}, nil
		}
	}

	receiver, peer := net.Pipe()
	defer receiver.Close()
	defer peer.Close()
	if err := peer.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("SetDeadline() error = %v", err)
	}

	result := make(chan error, 1)
	go func() {
		result <- ReceiveFile(receiver, accepted, prepare)
	}()
	if _, err := peer.Write([]byte(request)); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	rejection, err := readMessage(peer)
	if err != nil {
		t.Fatalf("readMessage() error = %v", err)
	}
	if rejection.Type != TypeReject {
		t.Fatalf("response type = %q; want reject", rejection.Type)
	}

	return rejection.Error, waitForProtocolResult(t, result)
}

func fileBeginJSON(path string, size int64) string {
	return fmt.Sprintf(
		"{\"type\":\"file_begin\",\"protocolVersion\":1,\"path\":%q,\"size\":%d}\n",
		path,
		size,
	)
}

func waitForProtocolResult(t *testing.T, result <-chan error) error {
	t.Helper()
	select {
	case err := <-result:
		return err
	case <-time.After(5 * time.Second):
		t.Fatal("protocol operation did not return")
		return nil
	}
}
