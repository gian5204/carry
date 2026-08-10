package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/gian5204/carry/internal/manifest"
	"github.com/gian5204/carry/internal/repo"
	"github.com/gian5204/carry/internal/ui"
)

// adds a file to Carry's manifest for the current repository
func Add(paths []string) error {
	repository, err := repo.Detect()
	if err != nil {
		return err
	}

	currentManifest, err := manifest.Load(repository)
	if err != nil {
		return err
	}

	managed := make(map[string]struct{}, len(currentManifest.Files))
	for _, path := range currentManifest.Files {
		managed[path] = struct{}{}
	}

	newPaths := make([]string, 0, len(paths))
	alreadyManaged := make([]string, 0)
	for _, path := range uniquePaths(paths) {
		if err := validateAddPath(repository, path); err != nil {
			return err
		}

		if _, exists := managed[path]; exists {
			alreadyManaged = append(alreadyManaged, path)
			continue
		}
		newPaths = append(newPaths, path)
	}

	for _, path := range alreadyManaged {
		fmt.Printf("%s %s is already managed by Carry\n", ui.Yellow("!"), path)
	}

	for _, path := range newPaths {
		currentManifest.Add(path)
	}

	if len(newPaths) > 0 {
		if err := manifest.Save(repository, currentManifest); err != nil {
			return err
		}
		printAddSuccess(newPaths)
	}

	return nil
}

func validateAddPath(repository *repo.Repository, path string) error {
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

	return nil
}

func printAddSuccess(paths []string) {
	if len(paths) == 1 {
		fmt.Printf(
			"%s %s %s %s\n",
			ui.Green("✓"),
			ui.BoldGreen("Added"),
			paths[0],
			ui.Dim("to Carry"),
		)
		return
	}

	fmt.Printf(
		"%s %s %d %s %s\n",
		ui.Green("✓"),
		ui.BoldGreen("Added"),
		len(paths),
		pluralizeFiles(len(paths)),
		ui.Dim("to Carry"),
	)
}
