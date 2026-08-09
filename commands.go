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

    fmt.Printf("%s can be managed by Carry\n", path)
    return nil
}