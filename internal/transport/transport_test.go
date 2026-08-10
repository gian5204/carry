package transport

import (
	"bytes"
	"net"
	"strings"
	"testing"
	"time"
)

func TestAcceptOne(t *testing.T) {
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

	var output bytes.Buffer
	result := make(chan error, 1)
	go func() {
		result <- acceptOne(listener, ":0", &output)
	}()

	connection, err := net.DialTimeout("tcp4", listener.Addr().String(), 5*time.Second)
	if err != nil {
		t.Fatalf("net.DialTimeout() error = %v", err)
	}
	if err := connection.Close(); err != nil {
		t.Fatalf("connection.Close() error = %v", err)
	}

	select {
	case err := <-result:
		if err != nil {
			t.Fatalf("acceptOne() error = %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("acceptOne() did not return")
	}

	if !strings.Contains(output.String(), "Listening on :0\n") {
		t.Errorf("output = %q; want listening message", output.String())
	}
	if !strings.Contains(output.String(), "Connected from 127.0.0.1:") {
		t.Errorf("output = %q; want connected message", output.String())
	}
}
