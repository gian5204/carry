package cmd

import (
	"fmt"

	"github.com/gian5204/carry/internal/repo"
)

func currentRepositoryIdentity() (string, error) {
	repository, err := repo.Detect()
	if err != nil {
		return "", fmt.Errorf("detect repository: %w", err)
	}

	identity, err := repository.ID()
	if err != nil {
		return "", fmt.Errorf("derive repository identity: %w", err)
	}

	return identity, nil
}
