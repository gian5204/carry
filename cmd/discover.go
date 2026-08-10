package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

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

	fmt.Println()
	approved, err := promptAddAll(bufio.NewReader(os.Stdin), os.Stdout)
	if err != nil {
		return err
	}
	if !approved {
		fmt.Println("No files were added.")
		return nil
	}

	currentManifest, err = manifest.Load(repository)
	if err != nil {
		return err
	}

	added := 0
	for _, file := range discoveredFiles {
		if currentManifest.Add(file) {
			added++
		}
	}

	if err := manifest.Save(repository, currentManifest); err != nil {
		return err
	}

	fmt.Printf(
		"  %s %s %d %s %s\n",
		ui.Green("✓"),
		ui.BoldGreen("Added"),
		added,
		pluralizeFiles(added),
		ui.Dim("to Carry"),
	)

	return nil
}

func promptAddAll(reader *bufio.Reader, output io.Writer) (bool, error) {
	fmt.Fprint(output, "Add all discovered files to Carry? [y/N] ")

	answer, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, err
	}

	return strings.EqualFold(strings.TrimSpace(answer), "y"), nil
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
