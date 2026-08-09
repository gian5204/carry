package main

import (
	"errors"
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

func (r *Repository) IsIgnored(path string) (bool, error) {
	cmd := exec.Command("git", "check-ignore", path)

	cmd.Dir = r.Root

	err := cmd.Run()
	if err == nil {
		return true, nil
	} else {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return false, nil
		}
	}
	return false, err
}