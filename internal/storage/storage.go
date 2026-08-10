package storage

import (
	"os"
	"path/filepath"

	"github.com/gian5204/carry/internal/repo"
)

func carryHome() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".carry"), nil
}

func storageDir(repository *repo.Repository) (string, error) {
	home, err := carryHome()
	if err != nil {
		return "", err
	}

	id, err := repository.ID()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, "repos", id), nil
}

func ensureDir(repository *repo.Repository) (string, error) {
	dir, err := storageDir(repository)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}

	return dir, nil
}
