package cmd

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/gian5204/carry/internal/repo"
)

func TestCurrentRepositoryIdentityMatchesClones(t *testing.T) {
	first := setupCommandRepository(t)
	addCommandTestOrigin(t, first, "https://github.com/example/project.git")
	firstIdentity, err := currentRepositoryIdentity()
	if err != nil {
		t.Fatalf("currentRepositoryIdentity() error = %v", err)
	}

	second := setupCommandRepository(t)
	addCommandTestOrigin(t, second, "git@github.com:example/project.git")
	secondIdentity, err := currentRepositoryIdentity()
	if err != nil {
		t.Fatalf("currentRepositoryIdentity() error = %v", err)
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

	_, err := currentRepositoryIdentity()
	if err == nil {
		t.Fatal("currentRepositoryIdentity() error = nil; want missing origin error")
	}
	if !strings.Contains(err.Error(), "derive repository identity") {
		t.Errorf("currentRepositoryIdentity() error = %q; want derivation error", err)
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
