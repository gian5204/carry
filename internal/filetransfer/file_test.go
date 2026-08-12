package filetransfer

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDestinationCommit(t *testing.T) {
	root := t.TempDir()
	destination, err := PrepareDestination(root, "config/local.env")
	if err != nil {
		t.Fatalf("PrepareDestination() error = %v", err)
	}
	t.Cleanup(func() { destination.Abort() })

	contents := "remote contents"
	if _, err := io.Copy(destination, strings.NewReader(contents)); err != nil {
		t.Fatalf("io.Copy() error = %v", err)
	}
	if err := destination.Commit(); err != nil {
		t.Fatalf("Commit() error = %v", err)
	}

	path := filepath.Join(root, "config", "local.env")
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != contents {
		t.Errorf("destination contents = %q; want %q", got, contents)
	}
}

func TestPrepareDestinationRefusesExistingFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".env")
	if err := os.WriteFile(path, []byte("existing"), 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := PrepareDestination(root, ".env")
	if !errors.Is(err, ErrDestinationExists) {
		t.Fatalf("PrepareDestination() error = %v; want destination exists", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != "existing" {
		t.Errorf("existing contents = %q; want untouched contents", got)
	}
}

func TestDestinationAbortRemovesTemporaryFile(t *testing.T) {
	root := t.TempDir()
	destination, err := PrepareDestination(root, "nested/local.env")
	if err != nil {
		t.Fatalf("PrepareDestination() error = %v", err)
	}
	if _, err := destination.Write([]byte("partial")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := destination.Abort(); err != nil {
		t.Fatalf("Abort() error = %v", err)
	}

	final := filepath.Join(root, "nested", "local.env")
	if _, err := os.Stat(final); !os.IsNotExist(err) {
		t.Fatalf("final file error = %v; want not exist", err)
	}
	temporary, err := filepath.Glob(filepath.Join(root, "nested", ".carry-transfer-*"))
	if err != nil {
		t.Fatalf("Glob() error = %v", err)
	}
	if len(temporary) != 0 {
		t.Errorf("temporary files = %v; want none", temporary)
	}
}

func TestOpenSource(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".env")
	if err := os.WriteFile(path, []byte("contents"), 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	file, size, err := OpenSource(root, ".env")
	if err != nil {
		t.Fatalf("OpenSource() error = %v", err)
	}
	defer file.Close()
	if size != int64(len("contents")) {
		t.Errorf("source size = %d; want %d", size, len("contents"))
	}
}
