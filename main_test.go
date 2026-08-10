package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gian5204/carry/internal/ui"
)

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

func TestPrintHelp(t *testing.T) {
	var output bytes.Buffer
	printHelp(&output)

	wantLines := []string{
		ui.Bold("Carry") + " — move local config between Git clones",
		ui.Bold("Usage:"),
		"  carry <command> [arguments]",
		ui.Bold("Commands:"),
		"  " + ui.Cyan("add") + " " + ui.Dim("<path...>") + "         Add files to Carry",
		"  " + ui.Cyan("copy") + " " + ui.Dim("<destination>") + "    Copy managed files to another clone",
		"  " + ui.Cyan("receive") + " " + ui.Dim("[port]") + "        Listen for an incoming Carry connection",
		"  " + ui.Cyan("help") + "                  Show this help",
	}

	for _, line := range wantLines {
		if !strings.Contains(output.String(), line) {
			t.Errorf("help output does not contain %q", line)
		}
	}
}
