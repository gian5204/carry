package main

import (
	"os/exec"
	"strings"
)

type Repository struct {
	Root string
}

func detectRepo() (*Repository, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()

	if err != nil {
		return nil, err
	}

	root := strings.TrimSpace(string(output))

	repo := &Repository{
		Root: root,
	}
	return repo, nil
}
