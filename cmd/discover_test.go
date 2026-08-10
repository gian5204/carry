package cmd

import (
	"path/filepath"
	"slices"
	"testing"
)

func TestFilterDiscoveryCandidatesExcludesBuiltInsAndManagedFiles(t *testing.T) {
	discovered := []string{
		"z.env",
		".env.local",
		filepath.Join("config", "local.json"),
		"a.env",
		"carry.exe",
	}
	managed := []string{
		".env.local",
		filepath.Join("config", "local.json"),
	}

	got := filterDiscoveryCandidates(discovered, managed)
	want := []string{"a.env", "z.env"}

	if !slices.Equal(got, want) {
		t.Errorf("filterDiscoveryCandidates() = %v; want %v", got, want)
	}
}

func TestFilterDiscoveryCandidatesSortsWithoutManifestEntries(t *testing.T) {
	discovered := []string{"z.env", "a.env", "m.env"}

	got := filterDiscoveryCandidates(discovered, nil)
	want := []string{"a.env", "m.env", "z.env"}

	if !slices.Equal(got, want) {
		t.Errorf("filterDiscoveryCandidates() = %v; want %v", got, want)
	}
}
