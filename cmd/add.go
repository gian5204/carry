package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/gian5204/carry/internal/manifest"
	"github.com/gian5204/carry/internal/repo"
)

// adds a file to Carry's manifest for the current repository
func Add(path string) error {
	repository, err := repo.Detect()
	if err != nil {
		return err
	}

	exists, err := repository.FileExists(path)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf(
			"file %s not found",
			filepath.Join(repository.Root, path),
		)
	}

	tracked, err := repository.IsTracked(path)
	if err != nil {
		return err
	}

	if tracked {
		return fmt.Errorf("file %s is already tracked by Git", path)
	}

	ignored, err := repository.IsIgnored(path)
	if err != nil {
		return err
	}

	if !ignored {
		return fmt.Errorf("file %s is not ignored by Git", path)
	}

	currentManifest, err := manifest.Load(repository)
	if err != nil {
		return err
	}

	added := currentManifest.Add(path)
	if !added {
		return fmt.Errorf("path %s is already managed by Carry", path)
	}

	if err := manifest.Save(repository, currentManifest); err != nil {
		return err
	}

	return nil
}
