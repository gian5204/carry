package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"

	"github.com/gian5204/carry/internal/manifest"
	"github.com/gian5204/carry/internal/repo"
)

func TestAddMultipleValidFiles(t *testing.T) {
	repository := setupCommandRepository(t)
	writeCommandTestFile(t, repository.Root, "first.env")
	writeCommandTestFile(t, repository.Root, "second.env")

	if err := Add([]string{"first.env", "second.env"}); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	assertManifestFiles(t, repository, []string{"first.env", "second.env"})
}

func TestAddNewAndAlreadyManagedFiles(t *testing.T) {
	repository := setupCommandRepository(t)
	writeCommandTestFile(t, repository.Root, "first.env")
	writeCommandTestFile(t, repository.Root, "second.env")
	seedManifest(t, repository, []string{"first.env"})

	if err := Add([]string{"first.env", "second.env"}); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	assertManifestFiles(t, repository, []string{"first.env", "second.env"})
}

func TestAddDuplicateArgumentsOnce(t *testing.T) {
	repository := setupCommandRepository(t)
	writeCommandTestFile(t, repository.Root, "first.env")

	if err := Add([]string{"first.env", "first.env"}); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	assertManifestFiles(t, repository, []string{"first.env"})
}

func TestAddValidationFailureDoesNotPartiallyUpdateManifest(t *testing.T) {
	repository := setupCommandRepository(t)
	writeCommandTestFile(t, repository.Root, "first.env")
	seedManifest(t, repository, []string{"existing.env"})

	if err := Add([]string{"first.env", "missing.env"}); err == nil {
		t.Fatal("Add() error = nil; want validation error")
	}

	assertManifestFiles(t, repository, []string{"existing.env"})
}

func TestRemoveMultipleManagedFiles(t *testing.T) {
	repository := setupCommandRepository(t)
	seedManifest(t, repository, []string{"first.env", "second.env", "third.env"})

	if err := Remove([]string{"first.env", "second.env"}); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	assertManifestFiles(t, repository, []string{"third.env"})
}

func TestRemoveManagedAndNotManagedFiles(t *testing.T) {
	repository := setupCommandRepository(t)
	seedManifest(t, repository, []string{"first.env", "third.env"})

	if err := Remove([]string{"first.env", "missing.env"}); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	assertManifestFiles(t, repository, []string{"third.env"})
}

func setupCommandRepository(t *testing.T) *repo.Repository {
	t.Helper()

	root := t.TempDir()
	command := exec.Command("git", "init", "--quiet", root)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git init error = %v; output = %s", err, output)
	}

	gitignore := "first.env\nsecond.env\nthird.env\nexisting.env\n"
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte(gitignore), 0644); err != nil {
		t.Fatalf("WriteFile(.gitignore) error = %v", err)
	}

	t.Chdir(root)
	return &repo.Repository{Root: root}
}

func writeCommandTestFile(t *testing.T, root, relativePath string) {
	t.Helper()

	fullPath := filepath.Join(root, relativePath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(fullPath, []byte("test"), 0644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", relativePath, err)
	}
}

func seedManifest(t *testing.T, repository *repo.Repository, files []string) {
	t.Helper()

	currentManifest := &manifest.Manifest{Version: 1, Files: files}
	if err := manifest.Save(repository, currentManifest); err != nil {
		t.Fatalf("manifest.Save() error = %v", err)
	}
}

func assertManifestFiles(t *testing.T, repository *repo.Repository, want []string) {
	t.Helper()

	currentManifest, err := manifest.Load(repository)
	if err != nil {
		t.Fatalf("manifest.Load() error = %v", err)
	}
	if !slices.Equal(currentManifest.Files, want) {
		t.Errorf("manifest files = %v; want %v", currentManifest.Files, want)
	}
}
