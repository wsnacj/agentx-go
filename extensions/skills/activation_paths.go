package skills

import (
	"path"
	"path/filepath"
	"strings"
)

const inactivePathScopeReason = "inactive_path_scope"

func NormalizeSkillPathScopes(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	out := make([]string, 0, len(items))
	seen := map[string]bool{}
	for _, raw := range items {
		normalized := normalizeSkillActivationPath(raw)
		if normalized == "" || seen[normalized] {
			continue
		}
		seen[normalized] = true
		out = append(out, normalized)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func NormalizeSkillActivationPaths(items []string) []string {
	return NormalizeSkillPathScopes(items)
}

func EvaluateSkillActivationPaths(skill Skill, activePaths []string) (bool, string, []string) {
	pathScopes := NormalizeSkillPathScopes(skill.Paths)
	if len(pathScopes) == 0 {
		return true, "", nil
	}
	normalizedActive := NormalizeSkillActivationPaths(activePaths)
	for _, candidate := range normalizedActive {
		for _, scope := range pathScopes {
			if skillActivationPathMatches(scope, candidate) {
				return true, "", nil
			}
		}
	}
	return false, inactivePathScopeReason, []string{
		"activate via workspace path matching one of: " + strings.Join(pathScopes, ", "),
	}
}

func SkillRequestedByName(skill Skill, requested []string) bool {
	if len(requested) == 0 {
		return false
	}
	key := strings.ToLower(strings.TrimSpace(ResolveSkillKey(skill)))
	name := strings.ToLower(strings.TrimSpace(skill.Name))
	for _, raw := range requested {
		value := strings.ToLower(strings.TrimSpace(raw))
		if value == "" {
			continue
		}
		if value == key || value == name {
			return true
		}
	}
	return false
}

func normalizeSkillActivationPath(raw string) string {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.Trim(trimmed, "`'\"")
	if trimmed == "" {
		return ""
	}
	cleaned := filepath.ToSlash(filepath.Clean(trimmed))
	if cleaned == "." || strings.HasPrefix(cleaned, "../") {
		return ""
	}
	cleaned = strings.TrimPrefix(cleaned, "./")
	cleaned = strings.TrimPrefix(cleaned, "/")
	if cleaned == "" {
		return ""
	}
	return cleaned
}

func skillActivationPathMatches(scope string, candidate string) bool {
	scope = normalizeSkillActivationPath(scope)
	candidate = normalizeSkillActivationPath(candidate)
	if scope == "" || candidate == "" {
		return false
	}
	if !containsPathGlob(scope) {
		return candidate == scope || strings.HasPrefix(candidate, scope+"/")
	}
	return matchSkillActivationSegments(splitSkillActivationPath(scope), splitSkillActivationPath(candidate))
}

func containsPathGlob(raw string) bool {
	return strings.ContainsAny(raw, "*?[")
}

func splitSkillActivationPath(raw string) []string {
	raw = strings.Trim(raw, "/")
	if raw == "" {
		return nil
	}
	return strings.Split(raw, "/")
}

func matchSkillActivationSegments(pattern []string, candidate []string) bool {
	if len(pattern) == 0 {
		return len(candidate) == 0
	}
	if pattern[0] == "**" {
		if matchSkillActivationSegments(pattern[1:], candidate) {
			return true
		}
		if len(candidate) == 0 {
			return false
		}
		return matchSkillActivationSegments(pattern, candidate[1:])
	}
	if len(candidate) == 0 {
		return false
	}
	ok, err := path.Match(pattern[0], candidate[0])
	if err != nil || !ok {
		return false
	}
	return matchSkillActivationSegments(pattern[1:], candidate[1:])
}
