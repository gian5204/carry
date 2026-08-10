package discovery

import "testing"

func TestShouldExclude(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		excluded bool
	}{
		{name: "node modules", path: "node_modules/package.json", excluded: true},
		{name: "nested node modules", path: "frontend/node_modules/package.json", excluded: true},
		{name: "Windows separators", path: `frontend\node_modules\package.json`, excluded: true},
		{name: "dist directory", path: "dist/app.js", excluded: true},
		{name: "executable", path: "carry.exe", excluded: true},
		{name: "uppercase executable", path: "tool.EXE", excluded: true},
		{name: "log file", path: "debug.log", excluded: true},
		{name: "environment file", path: ".env.local", excluded: false},
		{name: "testing JSON", path: "local_testing/.test.json", excluded: false},
		{name: "config JSON", path: "config/local.json", excluded: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShouldExclude(tt.path); got != tt.excluded {
				t.Errorf("ShouldExclude(%q) = %t; want %t", tt.path, got, tt.excluded)
			}
		})
	}
}
