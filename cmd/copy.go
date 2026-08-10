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
)

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

	reader := bufio.NewReader(os.Stdin)

	for _, file := range m.Files {
		sourcePath := filepath.Join(sourceRepo.Root, file)
		targetPath := filepath.Join(targetRepo.Root, file)

		overwrite := false
		if _, err := os.Stat(targetPath); err == nil {
			fmt.Printf("File %q already exists. Overwrite? (y/N) ", file)

			answer, err := reader.ReadString('\n')
			if err != nil && !errors.Is(err, io.EOF) {
				return err
			}
			if !strings.EqualFold(strings.TrimSpace(answer), "y") {
				continue
			}
			overwrite = true
		} else if !os.IsNotExist(err) {
			return err
		}

		if err := copyFile(sourcePath, targetPath, overwrite); err != nil {
			return err
		}
	}

	return nil
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

	dst, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	return nil
}
