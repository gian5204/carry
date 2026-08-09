package main

import (
	"fmt"
	"path/filepath"
)

func addCommand(path string) error {
	repo, err := detectRepo()
	if err != nil {
		return err
	}

	exists, err := repo.FileExists(path)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf(
			"file %s not found",
			filepath.Join(repo.Root, path),
		)
	}

	tracked, err := repo.IsTracked(path)
	if err != nil {
		return err
	}

	if tracked {
		return fmt.Errorf("file %s is already tracked by Git", path)
	}

	ignored, err := repo.IsIgnored(path)
	if err != nil {
		return err
	}

	if !ignored {
		return fmt.Errorf("file %s is not ignored by Git", path)
	}

	manifest, err := loadManifest(repo)
	if err != nil {
		return err
	}

	added := manifest.Add(path)
	if !added {
		return fmt.Errorf("path %s is already managed by Carry", path)
	}

	if err := saveManifest(repo, manifest); err != nil {
		return err
	}

	return nil
}

func listCommand() error {
	repo, err := detectRepo()
	if err != nil {
		return err
	}

	manifest, err := loadManifest(repo)
	if err != nil {
		return err
	}

	if len(manifest.Files) == 0 {
		return fmt.Errorf("no files are managed by Carry")
	}

	for _, file := range manifest.Files {
		fmt.Println(file)
	}

	return nil
}

func removeCommand(path string) error {
	repo, err := detectRepo()
	if err != nil {
		return err
	}

	manifest, err := loadManifest(repo)
	if err != nil {
		return err
	}

	if len(manifest.Files) == 0 {
		return fmt.Errorf("no files are managed by Carry")
	}

	removed := manifest.Remove(path)
	if !removed {
		return fmt.Errorf("path %s is not managed by Carry", path)
	}
	if err := saveManifest(repo, manifest); err != nil {
		return err
	}
	return nil
}
