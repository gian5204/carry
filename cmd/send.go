package cmd

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/gian5204/carry/internal/protocol"
	"github.com/gian5204/carry/internal/transport"
)

func Send(args []string) error {
	address, err := sendAddress(args)
	if err != nil {
		return err
	}
	repositoryID, err := currentRepositoryIdentity()
	if err != nil {
		return err
	}

	return transport.SendOnce(
		address,
		os.Stdout,
		func(connection net.Conn) error {
			if err := protocol.SendHandshake(connection); err != nil {
				return err
			}
			fmt.Fprintln(os.Stdout, "Handshake complete")
			if err := protocol.SendRepositoryVerification(
				connection,
				repositoryID,
			); err != nil {
				return err
			}
			fmt.Fprintln(os.Stdout, "Repository verified")
			return nil
		},
	)
}

func sendAddress(args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("send requires an address")
	}
	if len(args) > 1 {
		return "", fmt.Errorf("send accepts exactly one address")
	}

	address := args[0]
	host, portText, err := net.SplitHostPort(address)
	if err != nil {
		return "", fmt.Errorf("invalid address %q: %w", address, err)
	}
	if host == "" {
		return "", fmt.Errorf("invalid address %q: host is required", address)
	}

	port, err := strconv.Atoi(portText)
	if err != nil {
		return "", fmt.Errorf("invalid address %q: port must be numeric", address)
	}
	if port < minPort || port > maxPort {
		return "", fmt.Errorf(
			"invalid address %q: port must be between %d and %d",
			address,
			minPort,
			maxPort,
		)
	}

	return address, nil
}
