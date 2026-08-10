package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gian5204/carry/internal/repo"
)

func TestCurrentRepositoryIdentityMatchesClones(t *testing.T) {
	first := setupCommandRepository(t)
	addCommandTestOrigin(t, first, "https://github.com/example/project.git")
	_, firstIdentity, err := detectRepositoryIdentity()
	if err != nil {
		t.Fatalf("detectRepositoryIdentity() error = %v", err)
	}

	second := setupCommandRepository(t)
	addCommandTestOrigin(t, second, "git@github.com:example/project.git")
	_, secondIdentity, err := detectRepositoryIdentity()
	if err != nil {
		t.Fatalf("detectRepositoryIdentity() error = %v", err)
	}

	if firstIdentity != secondIdentity {
		t.Errorf(
			"repository identities differ: %q != %q",
			firstIdentity,
			secondIdentity,
		)
	}
}

func TestCurrentRepositoryIdentityRequiresOrigin(t *testing.T) {
	setupCommandRepository(t)

	_, _, err := detectRepositoryIdentity()
	if err == nil {
		t.Fatal("detectRepositoryIdentity() error = nil; want missing origin error")
	}
	if !strings.Contains(err.Error(), "derive repository identity") {
		t.Errorf("detectRepositoryIdentity() error = %q; want derivation error", err)
	}
}

func TestLoadManagedFilesMissingManifest(t *testing.T) {
	repository := setupCommandRepository(t)

	_, err := loadManagedFiles(repository)
	if err == nil {
		t.Fatal("loadManagedFiles() error = nil; want no managed files error")
	}
	if err.Error() != "no files are managed by Carry" {
		t.Errorf("loadManagedFiles() error = %q; want no managed files error", err)
	}
}

func TestLoadManagedFilesEmptyManifest(t *testing.T) {
	repository := setupCommandRepository(t)
	seedManifest(t, repository, nil)

	_, err := loadManagedFiles(repository)
	if err == nil {
		t.Fatal("loadManagedFiles() error = nil; want no managed files error")
	}
	if err.Error() != "no files are managed by Carry" {
		t.Errorf("loadManagedFiles() error = %q; want no managed files error", err)
	}
}

func TestLoadManagedFilesInvalidManifest(t *testing.T) {
	repository := setupCommandRepository(t)
	if err := os.WriteFile(
		filepath.Join(repository.Root, ".carry.json"),
		[]byte("{invalid"),
		0644,
	); err != nil {
		t.Fatalf("WriteFile(.carry.json) error = %v", err)
	}

	if _, err := loadManagedFiles(repository); err == nil {
		t.Fatal("loadManagedFiles() error = nil; want invalid manifest error")
	}
}

func addCommandTestOrigin(
	t *testing.T,
	repository *repo.Repository,
	remote string,
) {
	t.Helper()

	command := exec.Command("git", "remote", "add", "origin", remote)
	command.Dir = repository.Root
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git remote add origin error = %v; output = %s", err, output)
	}
}
