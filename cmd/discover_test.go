package cmd

import (
	"path/filepath"
	"slices"
	"testing"
)

func TestFilterManagedFilesExcludesManagedAndSorts(t *testing.T) {
	discovered := []string{
		"z.env",
		".env.local",
		filepath.Join("config", "local.json"),
		"a.env",
	}
	managed := []string{
		".env.local",
		filepath.Join("config", "local.json"),
	}

	got := filterManagedFiles(discovered, managed)
	want := []string{"a.env", "z.env"}

	if !slices.Equal(got, want) {
		t.Errorf("filterManagedFiles() = %v; want %v", got, want)
	}
}

func TestFilterManagedFilesReturnsSortedDiscoveriesWithoutManifestEntries(t *testing.T) {
	discovered := []string{"z.env", "a.env", "m.env"}

	got := filterManagedFiles(discovered, nil)
	want := []string{"a.env", "m.env", "z.env"}

	if !slices.Equal(got, want) {
		t.Errorf("filterManagedFiles() = %v; want %v", got, want)
	}
}
