package main

import (
	"os"
	"path/filepath"
)

func carryHome() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".carry"), nil
}

func (r *Repository) StorageDir() (string, error) {
	home, err := carryHome()
	if err != nil {
		return "", err
	}

	id, err := r.ID()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, "repos", id), nil
}

func ensureStorageDir(repo *Repository) (string, error) {
	dir, err := repo.StorageDir()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}

	return dir, nil
}
