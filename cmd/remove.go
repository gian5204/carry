package cmd

import (
	"fmt"

	"github.com/gian5204/carry/internal/manifest"
	"github.com/gian5204/carry/internal/repo"
	"github.com/gian5204/carry/internal/ui"
)

// removes a file from Carry's manifest for the current repository
func Remove(path string) error {
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

	removed := currentManifest.Remove(path)
	if !removed {
		return fmt.Errorf("%s %s is not managed by Carry", ui.Yellow("!"), path)
	}
	if err := manifest.Save(repository, currentManifest); err != nil {
		return err
	}

	fmt.Printf(
		"%s %s %s %s\n",
		ui.Green("✓"),
		ui.BoldGreen("Removed"),
		path,
		ui.Dim("from Carry"),
	)

	return nil
}
