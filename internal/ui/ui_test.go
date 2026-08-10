package ui

import "testing"

func TestUsage(t *testing.T) {
	got := Usage("add", "<path...>")
	want := "\x1b[1mUsage:\x1b[0m\n  carry \x1b[36madd\x1b[0m \x1b[2m<path...>\x1b[0m"

	if got != want {
		t.Errorf("Usage() = %q; want %q", got, want)
	}
}
