package discovery

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

const ignoreFileName = ".carryignore"

type ruleKind int

const (
	exactPathRule ruleKind = iota
	directoryRule
	extensionRule
)

type ignoreRule struct {
	kind  ruleKind
	value string
}

type IgnoreRules struct {
	rules []ignoreRule
}

func LoadIgnoreRules(repositoryRoot string) (IgnoreRules, error) {
	data, err := os.ReadFile(filepath.Join(repositoryRoot, ignoreFileName))
	if err != nil {
		if os.IsNotExist(err) {
			return IgnoreRules{}, nil
		}
		return IgnoreRules{}, err
	}

	return parseIgnoreRules(string(data)), nil
}

func (r IgnoreRules) ShouldExclude(filePath string) bool {
	normalized := normalize(filePath)
	if normalized == ignoreFileName {
		return true
	}

	for _, rule := range r.rules {
		switch rule.kind {
		case directoryRule:
			if strings.HasPrefix(normalized, rule.value+"/") ||
				strings.Contains(normalized, "/"+rule.value+"/") {
				return true
			}
		case extensionRule:
			if strings.HasSuffix(normalized, rule.value) {
				return true
			}
		case exactPathRule:
			if normalized == rule.value {
				return true
			}
		}
	}

	return false
}

func parseIgnoreRules(contents string) IgnoreRules {
	lines := strings.Split(contents, "\n")
	rules := make([]ignoreRule, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		normalized := strings.ReplaceAll(line, "\\", "/")
		kind := exactPathRule
		if strings.HasSuffix(normalized, "/") {
			kind = directoryRule
			normalized = strings.TrimSuffix(normalized, "/")
		} else if strings.HasPrefix(normalized, "*.") && strings.Count(normalized, "*") == 1 {
			kind = extensionRule
			normalized = strings.TrimPrefix(normalized, "*")
		}

		rules = append(rules, ignoreRule{
			kind:  kind,
			value: normalize(normalized),
		})
	}

	return IgnoreRules{rules: rules}
}

func normalize(value string) string {
	return path.Clean(strings.ReplaceAll(value, "\\", "/"))
}
