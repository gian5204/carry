package managedpath

import (
	"fmt"
	"strings"
)

func Normalize(value string) (string, error) {
	if value == "" {
		return "", fmt.Errorf("managed path is empty")
	}
	if strings.ContainsRune(value, '\x00') {
		return "", fmt.Errorf("managed path contains a null byte")
	}

	normalized := strings.ReplaceAll(value, "\\", "/")
	if strings.HasPrefix(normalized, "/") {
		return "", fmt.Errorf("managed path %q is absolute", value)
	}
	if strings.Contains(normalized, ":") {
		return "", fmt.Errorf("managed path %q contains a Windows volume or stream", value)
	}

	segments := strings.Split(normalized, "/")
	for _, segment := range segments {
		switch segment {
		case "":
			return "", fmt.Errorf("managed path %q contains an empty segment", value)
		case ".":
			return "", fmt.Errorf("managed path %q is not canonical", value)
		case "..":
			return "", fmt.Errorf("managed path %q contains traversal", value)
		}
	}

	return strings.Join(segments, "/"), nil
}

func NormalizeAll(values []string) ([]string, error) {
	if len(values) == 0 {
		return nil, fmt.Errorf("managed file list is empty")
	}

	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		path, err := Normalize(value)
		if err != nil {
			return nil, err
		}

		key := strings.ToLower(path)
		if _, exists := seen[key]; exists {
			return nil, fmt.Errorf("duplicate managed path %q", value)
		}
		seen[key] = struct{}{}
		normalized = append(normalized, path)
	}

	return normalized, nil
}
