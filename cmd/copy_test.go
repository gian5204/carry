package cmd

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCopyPlanCollectsDecisionsBeforeWriting(t *testing.T) {
	sourceRoot := t.TempDir()
	targetRoot := t.TempDir()

	writeTestFile(t, filepath.Join(sourceRoot, "skip.env"), "source skip")
	writeTestFile(t, filepath.Join(sourceRoot, "overwrite.env"), "source overwrite")
	writeTestFile(t, filepath.Join(sourceRoot, "nested", "new.env"), "source new")

	skipTarget := filepath.Join(targetRoot, "skip.env")
	overwriteTarget := filepath.Join(targetRoot, "overwrite.env")
	newTarget := filepath.Join(targetRoot, "nested", "new.env")
	writeTestFile(t, skipTarget, "target skip")
	writeTestFile(t, overwriteTarget, "target overwrite")

	var prompts bytes.Buffer
	plan, err := buildCopyPlan(
		sourceRoot,
		targetRoot,
		[]string{"skip.env", "overwrite.env", filepath.Join("nested", "new.env")},
		bufio.NewReader(strings.NewReader("n\ny\n")),
		&prompts,
	)
	if err != nil {
		t.Fatalf("buildCopyPlan() error = %v", err)
	}

	if len(plan) != 3 {
		t.Fatalf("len(plan) = %d; want 3", len(plan))
	}
	if !plan[0].skip || plan[0].overwrite {
		t.Errorf("first plan item = %+v; want skipped", plan[0])
	}
	if plan[1].skip || !plan[1].overwrite {
		t.Errorf("second plan item = %+v; want overwrite", plan[1])
	}
	if plan[2].skip || plan[2].overwrite {
		t.Errorf("third plan item = %+v; want normal copy", plan[2])
	}

	assertTestFile(t, skipTarget, "target skip")
	assertTestFile(t, overwriteTarget, "target overwrite")
	if _, err := os.Stat(newTarget); !os.IsNotExist(err) {
		t.Fatalf("new target exists before execution; stat error = %v", err)
	}

	copied, skipped, err := executeCopyPlan(plan)
	if err != nil {
		t.Fatalf("executeCopyPlan() error = %v", err)
	}
	if copied != 2 || skipped != 1 {
		t.Errorf("executeCopyPlan() = (%d, %d); want (2, 1)", copied, skipped)
	}

	assertTestFile(t, skipTarget, "target skip")
	assertTestFile(t, overwriteTarget, "source overwrite")
	assertTestFile(t, newTarget, "source new")
}

func TestCopySummaryUsesSingularAndPluralFileCounts(t *testing.T) {
	var output bytes.Buffer
	printCopySummary(&output, "target", 1, 2)

	if !strings.Contains(output.String(), "1 file") {
		t.Errorf("summary %q does not contain singular file count", output.String())
	}
	if !strings.Contains(output.String(), "2 files") {
		t.Errorf("summary %q does not contain plural file count", output.String())
	}
}

func writeTestFile(t *testing.T, path, contents string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(contents), 0644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

func assertTestFile(t *testing.T, path, expected string) {
	t.Helper()

	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	if string(contents) != expected {
		t.Errorf("contents of %q = %q; want %q", path, contents, expected)
	}
}
