package cmd

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/gian5204/carry/internal/discovery"
	"github.com/gian5204/carry/internal/manifest"
	"github.com/gian5204/carry/internal/repo"
	"github.com/gian5204/carry/internal/ui"
)

func Discover() error {
	repository, err := repo.Detect()
	if err != nil {
		return err
	}

	currentManifest, err := manifest.Load(repository)
	if err != nil {
		return err
	}

	ignoredUntrackedFiles, err := repository.IgnoredUntrackedFiles()
	if err != nil {
		return err
	}

	ignoreRules, err := discovery.LoadIgnoreRules(repository.Root)
	if err != nil {
		return err
	}

	discoveredFiles := filterDiscoveryCandidates(
		ignoredUntrackedFiles,
		currentManifest.Files,
		ignoreRules,
	)
	if len(discoveredFiles) == 0 {
		fmt.Println("No unmanaged local files discovered.")
		return nil
	}

	fmt.Printf(
		"%s %s\n\n",
		ui.Bold("Discovered files"),
		ui.Dim(fmt.Sprintf("(%d)", len(discoveredFiles))),
	)

	for _, file := range discoveredFiles {
		fmt.Printf("  %s %s\n", ui.Green("●"), file)
	}

	return nil
}

func filterDiscoveryCandidates(
	discoveredFiles,
	managedFiles []string,
	ignoreRules discovery.IgnoreRules,
) []string {
	managed := make(map[string]struct{}, len(managedFiles))
	for _, file := range managedFiles {
		managed[filepath.Clean(file)] = struct{}{}
	}

	filtered := make([]string, 0, len(discoveredFiles))
	for _, file := range discoveredFiles {
		if discovery.ShouldExclude(file) {
			continue
		}
		if ignoreRules.ShouldExclude(file) {
			continue
		}

		if _, exists := managed[filepath.Clean(file)]; !exists {
			filtered = append(filtered, file)
		}
	}

	sort.Strings(filtered)
	return filtered
}
