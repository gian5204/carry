package repo

import "testing"

func TestNormalizeRemote(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "https remote",
			input:    "https://github.com/gian/carry.git",
			expected: "github.com/gian/carry",
		},
		{
			name:     "ssh shorthand remote",
			input:    "git@github.com:gian/carry.git",
			expected: "github.com/gian/carry",
		},
		{
			name:     "ssh url remote",
			input:    "ssh://git@github.com/gian/carry.git",
			expected: "github.com/gian/carry",
		},
		{
			name:     "trailing whitespace",
			input:    "https://github.com/gian/carry.git\n",
			expected: "github.com/gian/carry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeRemote(tt.input)

			if got != tt.expected {
				t.Errorf("normalizeRemote(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}
