package cmd

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/gian5204/carry/internal/protocol"
	"github.com/gian5204/carry/internal/transport"
)

const (
	minPort = 1
	maxPort = 65535
)

func Receive(args []string) error {
	port, err := receivePort(args)
	if err != nil {
		return err
	}

	return transport.ReceiveOnce(
		port,
		os.Stdout,
		func(connection net.Conn) error {
			if err := protocol.ReceiveHandshake(connection); err != nil {
				return err
			}
			fmt.Fprintln(os.Stdout, "Handshake complete")
			return nil
		},
	)
}

func receivePort(args []string) (int, error) {
	if len(args) == 0 {
		return transport.DefaultPort, nil
	}
	if len(args) > 1 {
		return 0, fmt.Errorf("receive accepts at most one port")
	}

	port, err := strconv.Atoi(args[0])
	if err != nil {
		return 0, fmt.Errorf("invalid port %q: must be numeric", args[0])
	}
	if port < minPort || port > maxPort {
		return 0, fmt.Errorf(
			"invalid port %q: must be between %d and %d",
			args[0],
			minPort,
			maxPort,
		)
	}

	return port, nil
}
