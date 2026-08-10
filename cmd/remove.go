package cmd

import (
	"fmt"

	"github.com/gian5204/carry/internal/manifest"
	"github.com/gian5204/carry/internal/repo"
	"github.com/gian5204/carry/internal/ui"
)

// removes a file from Carry's manifest for the current repository
func Remove(paths []string) error {
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

	managedPaths := make([]string, 0, len(paths))
	notManaged := make([]string, 0)
	for _, path := range uniquePaths(paths) {
		if _, exists := managed[path]; exists {
			managedPaths = append(managedPaths, path)
		} else {
			notManaged = append(notManaged, path)
		}
	}

	for _, path := range notManaged {
		fmt.Printf("%s %s is not managed by Carry\n", ui.Yellow("!"), path)
	}

	for _, path := range managedPaths {
		currentManifest.Remove(path)
	}

	if len(managedPaths) > 0 {
		if err := manifest.Save(repository, currentManifest); err != nil {
			return err
		}
		printRemoveSuccess(managedPaths)
	}

	return nil
}

func printRemoveSuccess(paths []string) {
	if len(paths) == 1 {
		fmt.Printf(
			"%s %s %s %s\n",
			ui.Green("✓"),
			ui.BoldGreen("Removed"),
			paths[0],
			ui.Dim("from Carry"),
		)
		return
	}

	fmt.Printf(
		"%s %s %d %s %s\n",
		ui.Green("✓"),
		ui.BoldGreen("Removed"),
		len(paths),
		pluralizeFiles(len(paths)),
		ui.Dim("from Carry"),
	)
}
