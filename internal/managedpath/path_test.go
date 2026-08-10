package managedpath

import (
	"strings"
	"testing"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      string
		wantError string
	}{
		{name: "root file", input: ".env", want: ".env"},
		{name: "nested file", input: "config/local.json", want: "config/local.json"},
		{name: "Windows separators", input: `config\local.json`, want: "config/local.json"},
		{name: "empty", input: "", wantError: "empty"},
		{name: "POSIX absolute", input: "/etc/passwd", wantError: "absolute"},
		{name: "rooted Windows path", input: `\secrets\local.env`, wantError: "absolute"},
		{name: "Windows drive absolute", input: `C:\secrets\local.env`, wantError: "Windows volume"},
		{name: "Windows drive relative", input: `C:local.env`, wantError: "Windows volume"},
		{name: "Windows UNC", input: `\\server\share\local.env`, wantError: "absolute"},
		{name: "alternate data stream", input: `local.env:secret`, wantError: "Windows volume"},
		{name: "parent traversal", input: "../local.env", wantError: "traversal"},
		{name: "nested traversal", input: "config/../../local.env", wantError: "traversal"},
		{name: "backslash traversal", input: `config\..\local.env`, wantError: "traversal"},
		{name: "current directory", input: "./local.env", wantError: "not canonical"},
		{name: "empty segment", input: "config//local.env", wantError: "empty segment"},
		{name: "trailing separator", input: "config/", wantError: "empty segment"},
		{name: "null byte", input: "local\x00.env", wantError: "null byte"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Normalize(test.input)
			if test.wantError != "" {
				if err == nil {
					t.Fatalf("Normalize() error = nil; want error containing %q", test.wantError)
				}
				if !strings.Contains(err.Error(), test.wantError) {
					t.Errorf("Normalize() error = %q; want error containing %q", err, test.wantError)
				}
				return
			}

			if err != nil {
				t.Fatalf("Normalize() error = %v", err)
			}
			if got != test.want {
				t.Errorf("Normalize() = %q; want %q", got, test.want)
			}
		})
	}
}

func TestNormalizeAllRejectsDuplicates(t *testing.T) {
	tests := []struct {
		name  string
		paths []string
	}{
		{name: "exact", paths: []string{".env", ".env"}},
		{name: "separator equivalent", paths: []string{"config/local.json", `config\local.json`}},
		{name: "case equivalent", paths: []string{"config/local.json", "CONFIG/LOCAL.JSON"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NormalizeAll(test.paths)
			if err == nil {
				t.Fatal("NormalizeAll() error = nil; want duplicate error")
			}
			if !strings.Contains(err.Error(), "duplicate managed path") {
				t.Errorf("NormalizeAll() error = %q; want duplicate error", err)
			}
		})
	}
}
