package cmd

import (
	"fmt"

	"github.com/gian5204/carry/internal/manifest"
	"github.com/gian5204/carry/internal/repo"
)

func detectRepositoryIdentity() (*repo.Repository, string, error) {
	repository, err := repo.Detect()
	if err != nil {
		return nil, "", fmt.Errorf("detect repository: %w", err)
	}

	identity, err := repository.ID()
	if err != nil {
		return nil, "", fmt.Errorf("derive repository identity: %w", err)
	}

	return repository, identity, nil
}

func loadManagedFiles(repository *repo.Repository) ([]string, error) {
	currentManifest, err := manifest.Load(repository)
	if err != nil {
		return nil, err
	}
	if len(currentManifest.Files) == 0 {
		return nil, fmt.Errorf("no files are managed by Carry")
	}

	return append([]string(nil), currentManifest.Files...), nil
}
