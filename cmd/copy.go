package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gian5204/carry/internal/manifest"
	"github.com/gian5204/carry/internal/repo"
	"github.com/gian5204/carry/internal/ui"
)

type copyItem struct {
	sourcePath   string
	targetPath   string
	relativePath string
	overwrite    bool
	skip         bool
}

func Copy(destination string) error {
	sourceRepo, err := repo.Detect()
	if err != nil {
		return err
	}

	m, err := manifest.Load(sourceRepo)
	if err != nil {
		return err
	}

	if len(m.Files) == 0 {
		return fmt.Errorf("no files are managed by Carry")
	}

	targetRepo, err := repo.DetectAt(destination)
	if err != nil {
		return err
	}

	sourceID, err := sourceRepo.ID()
	if err != nil {
		return err
	}

	targetID, err := targetRepo.ID()
	if err != nil {
		return err
	}

	if sourceID != targetID {
		return fmt.Errorf("target is not a clone of the same repository")
	}

	plan, err := buildCopyPlan(
		sourceRepo.Root,
		targetRepo.Root,
		m.Files,
		bufio.NewReader(os.Stdin),
		os.Stdout,
	)
	if err != nil {
		return err
	}

	copied, skipped, err := executeCopyPlan(plan)
	if err != nil {
		return err
	}

	printCopySummary(os.Stdout, targetRepo.Root, copied, skipped)

	return nil
}

func buildCopyPlan(
	sourceRoot string,
	targetRoot string,
	files []string,
	reader *bufio.Reader,
	output io.Writer,
) ([]copyItem, error) {
	plan := make([]copyItem, 0, len(files))

	for _, file := range files {
		item := copyItem{
			sourcePath:   filepath.Join(sourceRoot, file),
			targetPath:   filepath.Join(targetRoot, file),
			relativePath: file,
		}

		if _, err := os.Stat(item.targetPath); err == nil {
			fmt.Fprintf(
				output,
				"%s %s already exists in target. Overwrite? [y/N] ",
				ui.Yellow("!"),
				item.relativePath,
			)

			answer, err := reader.ReadString('\n')
			if err != nil && !errors.Is(err, io.EOF) {
				return nil, err
			}
			if strings.EqualFold(strings.TrimSpace(answer), "y") {
				item.overwrite = true
			} else {
				item.skip = true
			}
		} else if !os.IsNotExist(err) {
			return nil, err
		}

		plan = append(plan, item)
	}

	return plan, nil
}

func executeCopyPlan(plan []copyItem) (int, int, error) {
	copied := 0
	skipped := 0

	for _, item := range plan {
		if item.skip {
			skipped++
			continue
		}

		if err := copyFile(item.sourcePath, item.targetPath, item.overwrite); err != nil {
			return copied, skipped, err
		}
		copied++
	}

	return copied, skipped, nil
}

func printCopySummary(output io.Writer, destination string, copied, skipped int) {
	fmt.Fprintf(
		output,
		"%s %s %d %s %s %s\n",
		ui.Green("✓"),
		ui.BoldGreen("Copied"),
		copied,
		pluralizeFiles(copied),
		ui.Dim("to"),
		destination,
	)

	if skipped > 0 {
		fmt.Fprintf(
			output,
			"  %d %s %s\n",
			skipped,
			pluralizeFiles(skipped),
			ui.Dim("skipped"),
		)
	}
}

func pluralizeFiles(count int) string {
	if count == 1 {
		return "file"
	}
	return "files"
}

func copyFile(sourcePath, targetPath string, overwrite bool) error {
	if !overwrite {
		if _, err := os.Stat(targetPath); err == nil {
			return fmt.Errorf("target file %s already exists", targetPath)
		} else if !os.IsNotExist(err) {
			return err
		}
	}

	src, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer src.Close()

	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return err
	}

	flags := os.O_WRONLY | os.O_CREATE | os.O_EXCL
	if overwrite {
		flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	}

	dst, err := os.OpenFile(targetPath, flags, 0666)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	return nil
}
