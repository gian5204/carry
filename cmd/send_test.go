package cmd

import (
	"strings"
	"testing"
)

func TestSendAddress(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		want      string
		wantError string
	}{
		{
			name:      "missing address",
			wantError: "requires an address",
		},
		{
			name:      "malformed address",
			args:      []string{"127.0.0.1"},
			wantError: "invalid address",
		},
		{
			name:      "too many arguments",
			args:      []string{"127.0.0.1:4242", "extra"},
			wantError: "exactly one address",
		},
		{
			name: "valid address",
			args: []string{"127.0.0.1:4242"},
			want: "127.0.0.1:4242",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := sendAddress(test.args)
			if test.wantError != "" {
				if err == nil {
					t.Fatalf("sendAddress() error = nil; want error containing %q", test.wantError)
				}
				if !strings.Contains(err.Error(), test.wantError) {
					t.Errorf("sendAddress() error = %q; want error containing %q", err, test.wantError)
				}
				return
			}

			if err != nil {
				t.Fatalf("sendAddress() error = %v", err)
			}
			if got != test.want {
				t.Errorf("sendAddress() = %q; want %q", got, test.want)
			}
		})
	}
}
