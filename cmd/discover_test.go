package cmd

import (
	"bufio"
	"bytes"
	"path/filepath"
	"slices"
	"strings"
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

func TestPromptAddAll(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		approved bool
	}{
		{name: "lowercase yes", input: "y\n", approved: true},
		{name: "uppercase yes", input: "Y\n", approved: true},
		{name: "lowercase no", input: "n\n", approved: false},
		{name: "uppercase no", input: "N\n", approved: false},
		{name: "empty response", input: "\n", approved: false},
		{name: "end of input", input: "", approved: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			approved, err := promptAddAll(
				bufio.NewReader(strings.NewReader(tt.input)),
				&output,
			)
			if err != nil {
				t.Fatalf("promptAddAll() error = %v", err)
			}
			if approved != tt.approved {
				t.Errorf("promptAddAll() = %t; want %t", approved, tt.approved)
			}
			if output.String() != "Add all discovered files to Carry? [y/N] " {
				t.Errorf("prompt = %q", output.String())
			}
		})
	}
}
