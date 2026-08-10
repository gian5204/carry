package main

import "testing"

func TestVersionTextUsesVersionVariable(t *testing.T) {
	originalVersion := version
	t.Cleanup(func() {
		version = originalVersion
	})

	version = "v0.1.0"
	if got := versionText(); got != "Carry v0.1.0" {
		t.Errorf("versionText() = %q; want %q", got, "Carry v0.1.0")
	}
}
