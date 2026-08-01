package skills

import "strings"

type RequestedSkillSemantic struct {
	Name             string   `json:"name,omitempty"`
	ExecutionContext string   `json:"execution_context,omitempty"`
	AllowedTools     []string `json:"allowed_tools,omitempty"`
	Effort           string   `json:"effort,omitempty"`
}

func SkillRequestedSemantic(skill Skill) RequestedSkillSemantic {
	name := strings.ToLower(strings.TrimSpace(skill.Name))
	if name == "" {
		return RequestedSkillSemantic{}
	}
	return RequestedSkillSemantic{
		Name:             name,
		ExecutionContext: NormalizeSkillExecutionContext(skill.ExecutionContext),
		AllowedTools:     NormalizeSkillAllowedTools(skill.AllowedTools),
		Effort:           NormalizeSkillExecutionEffort(skill.Effort),
	}
}

func ResolveRequestedSkillSemantics(skills []Skill, requested []string) []RequestedSkillSemantic {
	if len(skills) == 0 || len(requested) == 0 {
		return nil
	}
	byName := make(map[string]RequestedSkillSemantic, len(skills))
	for _, item := range skills {
		semantic := SkillRequestedSemantic(item)
		if semantic.Name == "" || !hasRequestedSkillSemanticDetails(semantic) {
			continue
		}
		if _, exists := byName[semantic.Name]; exists {
			continue
		}
		byName[semantic.Name] = semantic
	}
	if len(byName) == 0 {
		return nil
	}
	out := make([]RequestedSkillSemantic, 0, len(requested))
	for _, raw := range requested {
		name := strings.ToLower(strings.TrimSpace(raw))
		if name == "" {
			continue
		}
		semantic, ok := byName[name]
		if !ok {
			continue
		}
		out = append(out, semantic)
	}
	return MergeRequestedSkillSemantics(out)
}

func MergeRequestedSkillSemantics(groups ...[]RequestedSkillSemantic) []RequestedSkillSemantic {
	if len(groups) == 0 {
		return nil
	}
	out := make([]RequestedSkillSemantic, 0)
	indexByName := map[string]int{}
	for _, group := range groups {
		for _, item := range group {
			normalized := normalizeRequestedSkillSemantic(item)
			if normalized.Name == "" || !hasRequestedSkillSemanticDetails(normalized) {
				continue
			}
			if idx, exists := indexByName[normalized.Name]; exists {
				out[idx] = mergeRequestedSkillSemantic(out[idx], normalized)
				continue
			}
			indexByName[normalized.Name] = len(out)
			out = append(out, normalized)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func normalizeRequestedSkillSemantic(raw RequestedSkillSemantic) RequestedSkillSemantic {
	raw.Name = strings.ToLower(strings.TrimSpace(raw.Name))
	raw.ExecutionContext = NormalizeSkillExecutionContext(raw.ExecutionContext)
	raw.AllowedTools = NormalizeSkillAllowedTools(raw.AllowedTools)
	raw.Effort = NormalizeSkillExecutionEffort(raw.Effort)
	return raw
}

func mergeRequestedSkillSemantic(primary RequestedSkillSemantic, fallback RequestedSkillSemantic) RequestedSkillSemantic {
	primary = normalizeRequestedSkillSemantic(primary)
	fallback = normalizeRequestedSkillSemantic(fallback)
	if primary.Name == "" {
		return fallback
	}
	out := primary
	if out.ExecutionContext == "" {
		out.ExecutionContext = fallback.ExecutionContext
	}
	if len(out.AllowedTools) == 0 && len(fallback.AllowedTools) > 0 {
		out.AllowedTools = append([]string(nil), fallback.AllowedTools...)
	}
	if out.Effort == "" {
		out.Effort = fallback.Effort
	}
	return out
}

func hasRequestedSkillSemanticDetails(item RequestedSkillSemantic) bool {
	return strings.TrimSpace(item.ExecutionContext) != "" ||
		len(item.AllowedTools) > 0 ||
		strings.TrimSpace(item.Effort) != ""
}
