package transport

import (
	"fmt"
	"io"
	"net"
)

const DefaultPort = 4242

func ReceiveOnce(port int, output io.Writer) error {
	address := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", address, err)
	}
	defer listener.Close()

	return acceptOne(listener, address, output)
}

func acceptOne(listener net.Listener, displayAddress string, output io.Writer) error {
	fmt.Fprintf(output, "Listening on %s\n", displayAddress)

	connection, err := listener.Accept()
	if err != nil {
		return fmt.Errorf("accept connection: %w", err)
	}

	fmt.Fprintf(output, "Connected from %s\n", connection.RemoteAddr())
	if err := connection.Close(); err != nil {
		return fmt.Errorf("close connection: %w", err)
	}

	return nil
}
