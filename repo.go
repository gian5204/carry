package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"crypto/sha256"
	"fmt"
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

func (r *Repository) Remote() (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = r.Root

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func (r *Repository) ID() (string, error) {
	remote, err := r.Remote()
	if err != nil {
		return "", err
	}

	normalized := normalizeRemote(remote)

	sum := sha256.Sum256([]byte(normalized))
	return fmt.Sprintf("%x", sum), nil
}

func (r *Repository) FileExists(path string) (bool, error) {
	fullPath := filepath.Join(r.Root, path)
	_, err := os.Stat(fullPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (r *Repository) IsTracked(path string) (bool, error) {
	cmd := exec.Command("git", "ls-files", "--error-unmatch", path)

	cmd.Dir = r.Root

	err := cmd.Run()
	if err == nil {
		return true, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return false, nil
	}
	return false, err
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

func normalizeRemote(remote string) string {
	remote = strings.TrimSpace(remote)

	remote = strings.TrimSuffix(remote, ".git")

	if strings.HasPrefix(remote, "git@") {
		remote = strings.TrimPrefix(remote, "git@")
		remote = strings.Replace(remote, ":", "/", 1)
	}

	remote = strings.TrimPrefix(remote, "https://")
	remote = strings.TrimPrefix(remote, "http://")
	remote = strings.TrimPrefix(remote, "ssh://git@")

	return strings.ToLower(remote)
}
