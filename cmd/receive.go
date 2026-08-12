package cmd

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/gian5204/carry/internal/filetransfer"
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
	repository, repositoryID, err := detectRepositoryIdentity()
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
			if err := protocol.ReceiveRepositoryVerification(
				connection,
				repositoryID,
			); err != nil {
				return err
			}
			fmt.Fprintln(os.Stdout, "Repository verified")
			managedFiles, err := protocol.ReceiveManagedFiles(connection)
			if err != nil {
				return err
			}
			fmt.Fprintln(os.Stdout, "Managed files accepted")

			if err := protocol.ReceiveFile(
				connection,
				managedFiles,
				func(path string, _ int64) (protocol.FileDestination, error) {
					destination, err := filetransfer.PrepareDestination(
						repository.Root,
						path,
					)
					if errors.Is(err, filetransfer.ErrDestinationExists) {
						return nil, protocol.Reject(
							"destination file already exists",
							err,
						)
					}
					if errors.Is(err, filetransfer.ErrUnsafePath) {
						return nil, protocol.Reject("invalid file path", err)
					}
					return destination, err
				},
			); err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "Received %s\n", managedFiles[0])
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
