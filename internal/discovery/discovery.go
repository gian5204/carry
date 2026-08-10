package discovery

import (
	"path"
	"strings"
)

var excludedDirectories = map[string]struct{}{
	"node_modules": {},
	"dist":         {},
	"build":        {},
	"coverage":     {},
}

var excludedExtensions = map[string]struct{}{
	".exe":   {},
	".dll":   {},
	".so":    {},
	".dylib": {},
	".log":   {},
	".tmp":   {},
}

func ShouldExclude(filePath string) bool {
	normalized := strings.ReplaceAll(filePath, "\\", "/")
	normalized = strings.ToLower(path.Clean(normalized))

	parts := strings.Split(normalized, "/")
	for _, part := range parts[:len(parts)-1] {
		if _, excluded := excludedDirectories[part]; excluded {
			return true
		}
	}

	_, excluded := excludedExtensions[path.Ext(normalized)]
	return excluded
}
