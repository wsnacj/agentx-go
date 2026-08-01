package skills

import (
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var resourcePathPattern = regexp.MustCompile(`(?i)(?:^|[^A-Za-z0-9_./-])((?:scripts|references|assets)/[A-Za-z0-9._/-]+)`)

func ExtractReferencedResourcePaths(content string) []string {
	if strings.TrimSpace(content) == "" {
		return nil
	}
	matches := resourcePathPattern.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}
	out := make([]string, 0, len(matches))
	seen := map[string]bool{}
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		path := normalizeResourcePath(match[1])
		if path == "" || seen[path] {
			continue
		}
		seen[path] = true
		out = append(out, path)
	}
	if len(out) == 0 {
		return nil
	}
	sort.Strings(out)
	return out
}

func MissingReferencedResourcePaths(skill Skill) []string {
	expected := ExtractReferencedResourcePaths(skill.Content)
	if len(expected) == 0 {
		return nil
	}
	available := resourcePathIndex(skill.Resources)
	missing := make([]string, 0)
	for _, item := range expected {
		if available[strings.ToLower(item)] {
			continue
		}
		missing = append(missing, item)
	}
	if len(missing) == 0 {
		return nil
	}
	sort.Strings(missing)
	return missing
}

func resourcePathIndex(resources Resources) map[string]bool {
	out := map[string]bool{}
	collect := func(items []string) {
		for _, item := range items {
			normalized := strings.ToLower(normalizeResourcePath(item))
			if normalized == "" {
				continue
			}
			out[normalized] = true
		}
	}
	collect(resources.Scripts)
	collect(resources.References)
	collect(resources.Assets)
	if len(out) == 0 {
		return nil
	}
	return out
}

func normalizeResourcePath(raw string) string {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.Trim(trimmed, "`'\"()[]{}<>.,;")
	if trimmed == "" {
		return ""
	}
	cleaned := filepath.ToSlash(filepath.Clean(trimmed))
	if cleaned == "." || strings.HasPrefix(cleaned, "../") {
		return ""
	}
	return cleaned
}
