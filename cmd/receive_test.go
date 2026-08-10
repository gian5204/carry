package cmd

import (
	"strings"
	"testing"

	"github.com/gian5204/carry/internal/transport"
)

func TestReceivePort(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		want      int
		wantError string
	}{
		{
			name: "default port",
			want: transport.DefaultPort,
		},
		{
			name: "valid custom port",
			args: []string{"5000"},
			want: 5000,
		},
		{
			name:      "non-numeric port",
			args:      []string{"invalid"},
			wantError: "must be numeric",
		},
		{
			name:      "port below valid range",
			args:      []string{"0"},
			wantError: "must be between 1 and 65535",
		},
		{
			name:      "port above valid range",
			args:      []string{"65536"},
			wantError: "must be between 1 and 65535",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := receivePort(test.args)
			if test.wantError != "" {
				if err == nil {
					t.Fatalf("receivePort() error = nil; want error containing %q", test.wantError)
				}
				if !strings.Contains(err.Error(), test.wantError) {
					t.Errorf("receivePort() error = %q; want error containing %q", err, test.wantError)
				}
				return
			}

			if err != nil {
				t.Fatalf("receivePort() error = %v", err)
			}
			if got != test.want {
				t.Errorf("receivePort() = %d; want %d", got, test.want)
			}
		})
	}
}
