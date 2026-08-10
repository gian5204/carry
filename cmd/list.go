package cmd

import (
	"fmt"

	"github.com/gian5204/carry/internal/manifest"
	"github.com/gian5204/carry/internal/repo"
)

// prints files managed by Carry for the current repository
func List() error {
	repository, err := repo.Detect()
	if err != nil {
		return err
	}

	currentManifest, err := manifest.Load(repository)
	if err != nil {
		return err
	}

	if len(currentManifest.Files) == 0 {
		return fmt.Errorf("no files are managed by Carry")
	}

	for _, file := range currentManifest.Files {
		fmt.Println(file)
	}

	return nil
}
