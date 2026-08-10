package cmd

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/gian5204/carry/internal/discovery"
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

	got := filterDiscoveryCandidates(discovered, managed, discovery.IgnoreRules{})
	want := []string{"a.env", "z.env"}

	if !slices.Equal(got, want) {
		t.Errorf("filterDiscoveryCandidates() = %v; want %v", got, want)
	}
}

func TestFilterDiscoveryCandidatesSortsWithoutManifestEntries(t *testing.T) {
	discovered := []string{"z.env", "a.env", "m.env"}

	got := filterDiscoveryCandidates(discovered, nil, discovery.IgnoreRules{})
	want := []string{"a.env", "m.env", "z.env"}

	if !slices.Equal(got, want) {
		t.Errorf("filterDiscoveryCandidates() = %v; want %v", got, want)
	}
}
