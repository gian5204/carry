package discovery

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIgnoreRulesShouldExclude(t *testing.T) {
	repositoryRoot := t.TempDir()
	contents := `
# Local test data
  local_testing/  

# Local databases
  *.sqlite

  config/local.json  
`
	if err := os.WriteFile(filepath.Join(repositoryRoot, ignoreFileName), []byte(contents), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	rules, err := LoadIgnoreRules(repositoryRoot)
	if err != nil {
		t.Fatalf("LoadIgnoreRules() error = %v", err)
	}

	tests := []struct {
		path     string
		excluded bool
	}{
		{path: "local_testing/test.json", excluded: true},
		{path: "local_testing/nested/test.json", excluded: true},
		{path: "frontend/local_testing/test.json", excluded: true},
		{path: `local_testing\test.json`, excluded: true},
		{path: "database.sqlite", excluded: true},
		{path: "data/database.sqlite", excluded: true},
		{path: "config/local.json", excluded: true},
		{path: "config/other.json", excluded: false},
		{path: ".env.local", excluded: false},
		{path: ignoreFileName, excluded: true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := rules.ShouldExclude(tt.path); got != tt.excluded {
				t.Errorf("ShouldExclude(%q) = %t; want %t", tt.path, got, tt.excluded)
			}
		})
	}

	if len(rules.rules) != 3 {
		t.Errorf("loaded %d rules; want 3", len(rules.rules))
	}
}

func TestLoadIgnoreRulesAllowsMissingFile(t *testing.T) {
	rules, err := LoadIgnoreRules(t.TempDir())
	if err != nil {
		t.Fatalf("LoadIgnoreRules() error = %v", err)
	}
	if rules.ShouldExclude(".env.local") {
		t.Error("empty rules excluded .env.local")
	}
}
