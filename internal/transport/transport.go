package transport

import (
	"fmt"
	"io"
	"net"
	"time"
)

const DefaultPort = 4242

const connectTimeout = 5 * time.Second

func ReceiveOnce(
	port int,
	output io.Writer,
	handleConnection func(net.Conn) error,
) error {
	address := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", address, err)
	}
	defer listener.Close()

	return acceptOne(listener, address, output, handleConnection)
}

func acceptOne(
	listener net.Listener,
	displayAddress string,
	output io.Writer,
	handleConnection func(net.Conn) error,
) error {
	fmt.Fprintf(output, "Listening on %s\n", displayAddress)

	connection, err := listener.Accept()
	if err != nil {
		return fmt.Errorf("accept connection: %w", err)
	}

	fmt.Fprintf(output, "Connected from %s\n", connection.RemoteAddr())
	return useConnection(connection, handleConnection)
}

func SendOnce(
	address string,
	output io.Writer,
	handleConnection func(net.Conn) error,
) error {
	connection, err := net.DialTimeout("tcp", address, connectTimeout)
	if err != nil {
		return fmt.Errorf("connect to %s: %w", address, err)
	}

	fmt.Fprintf(output, "Connected to %s\n", address)
	return useConnection(connection, handleConnection)
}

func useConnection(
	connection net.Conn,
	handleConnection func(net.Conn) error,
) (err error) {
	defer func() {
		if closeErr := connection.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close connection: %w", closeErr)
		}
	}()

	return handleConnection(connection)
}
