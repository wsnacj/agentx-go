package tools

import (
	"errors"
	"sort"
	"strings"
)

type nameRepairMatch struct {
	name       string
	handler    Handler
	candidates []string
	reason     string
	ok         bool
	ambiguous  bool
}

// ToolNameRepairResolution describes deterministic normalization and candidate selection.
type ToolNameRepairResolution struct {
	Requested     string
	Name          string
	NormalizedKey string
	Candidates    []string
	Repaired      bool
	Ambiguous     bool
	Reason        string
	Confidence    string
}

// RepairToolName resolves a model-emitted name against registered names.
func RepairToolName(requested string, registered []string) ToolNameRepairResolution {
	keys := nameRepairKeys(requested)
	if len(keys) == 0 || len(registered) == 0 {
		return ToolNameRepairResolution{Requested: strings.TrimSpace(requested)}
	}
	rawRequested := requested
	exactRequested := strings.TrimSpace(requested)
	normalizedKey := keys[0]
	for _, name := range registered {
		if strings.TrimSpace(name) == exactRequested {
			return ToolNameRepairResolution{Requested: exactRequested, Name: name, NormalizedKey: normalizedKey,
				Repaired: strings.TrimSpace(name) != rawRequested, Reason: "exact_trim_match", Confidence: "high"}
		}
	}
	keySet := make(map[string]bool, len(keys))
	for _, key := range keys {
		keySet[key] = true
	}
	matches := make([]string, 0, 1)
	seen := map[string]bool{}
	for _, name := range registered {
		key := canonicalNameKey(name)
		if key == "" || !keySet[key] || seen[name] {
			continue
		}
		seen[name] = true
		matches = append(matches, name)
	}
	sort.Strings(matches)
	switch len(matches) {
	case 0:
		return ToolNameRepairResolution{Requested: exactRequested, NormalizedKey: normalizedKey,
			Candidates: suggestCandidates(registered, normalizedKey), Reason: "no_canonical_match", Confidence: "none"}
	case 1:
		return ToolNameRepairResolution{Requested: exactRequested, Name: matches[0], NormalizedKey: normalizedKey,
			Candidates: matches, Repaired: strings.TrimSpace(matches[0]) != exactRequested,
			Reason: "canonical_name_match", Confidence: "high"}
	default:
		return ToolNameRepairResolution{Requested: exactRequested, NormalizedKey: normalizedKey,
			Candidates: matches, Ambiguous: true, Reason: "ambiguous_canonical_match", Confidence: "none"}
	}
}

// ToolNameError reports a missing or ambiguous tool name.
type ToolNameError struct {
	Requested     string
	NormalizedKey string
	Candidates    []string
	Ambiguous     bool
	Reason        string
}

func (e *ToolNameError) Error() string {
	requested := strings.TrimSpace(e.Requested)
	if requested == "" {
		requested = "<empty>"
	}
	if e.Ambiguous {
		return "llmx: tool " + requested + " not registered; ambiguous candidates: " + strings.Join(e.Candidates, ", ")
	}
	return "llmx: tool " + requested + " not registered" + formatCandidates(e.Candidates)
}

// Code returns a stable error classification.
func (e *ToolNameError) Code() string {
	if e != nil && e.Ambiguous {
		return "ambiguous_tool_name"
	}
	return "invalid_tool_name"
}

// Repairable reports whether the caller can retry with one suggested candidate.
func (e *ToolNameError) Repairable() bool { return e != nil && !e.Ambiguous && len(e.Candidates) > 0 }

// NewToolNameError constructs a defensive typed error.
func NewToolNameError(requested string, resolution ToolNameRepairResolution) *ToolNameError {
	return &ToolNameError{Requested: firstNonEmpty(strings.TrimSpace(resolution.Requested), strings.TrimSpace(requested)),
		NormalizedKey: strings.TrimSpace(resolution.NormalizedKey), Candidates: append([]string(nil), resolution.Candidates...),
		Ambiguous: resolution.Ambiguous, Reason: strings.TrimSpace(resolution.Reason)}
}

// AsToolNameError extracts a typed tool-name error.
func AsToolNameError(err error) (*ToolNameError, bool) {
	var typed *ToolNameError
	if errors.As(err, &typed) && typed != nil {
		return typed, true
	}
	return nil, false
}

func resolveNameRepair(entries map[string]entry, requested string) nameRepairMatch {
	if len(entries) == 0 {
		return nameRepairMatch{}
	}
	names := make([]string, 0, len(entries))
	for name := range entries {
		names = append(names, name)
	}
	resolution := RepairToolName(requested, names)
	if resolution.Repaired {
		item := entries[resolution.Name]
		return nameRepairMatch{name: item.definition.Function.Name, handler: item.handler, candidates: resolution.Candidates, reason: resolution.Reason, ok: true}
	}
	if resolution.Ambiguous {
		return nameRepairMatch{candidates: resolution.Candidates, reason: resolution.Reason, ambiguous: true}
	}
	return nameRepairMatch{candidates: resolution.Candidates, reason: resolution.Reason}
}

func nameRepairKeys(name string) []string {
	canonical := canonicalNameKey(name)
	if canonical == "" {
		return nil
	}
	keys := []string{canonical}
	if stripped := stripRepairSuffix(canonical); stripped != "" && stripped != canonical {
		keys = append(keys, stripped)
	}
	return keys
}

func canonicalNameKey(name string) string {
	normalized := strings.ToLower(strings.TrimSpace(name))
	if normalized == "" {
		return ""
	}
	var builder strings.Builder
	lastUnderscore := false
	for _, character := range normalized {
		switch {
		case character >= 'a' && character <= 'z', character >= '0' && character <= '9':
			builder.WriteRune(character)
			lastUnderscore = false
		default:
			if !lastUnderscore && builder.Len() > 0 {
				builder.WriteByte('_')
				lastUnderscore = true
			}
		}
	}
	return strings.Trim(builder.String(), "_")
}

func stripRepairSuffix(name string) string {
	if strings.HasSuffix(name, "_tool") {
		return strings.TrimSuffix(name, "_tool")
	}
	if strings.HasSuffix(name, "tool") && len(name) > len("tool") {
		return strings.TrimSuffix(name, "tool")
	}
	return name
}

func suggestCandidates(names []string, requestedKey string) []string {
	if requestedKey == "" || len(names) == 0 {
		return nil
	}
	candidates := make([]string, 0, 5)
	for _, name := range names {
		key := canonicalNameKey(name)
		if key != "" && (strings.Contains(key, requestedKey) || strings.Contains(requestedKey, key)) {
			candidates = append(candidates, name)
		}
	}
	sort.Strings(candidates)
	if len(candidates) > 5 {
		candidates = candidates[:5]
	}
	return candidates
}

func formatCandidates(candidates []string) string {
	if len(candidates) == 0 {
		return ""
	}
	return "; candidates: " + strings.Join(candidates, ", ")
}

func firstNonEmpty(items ...string) string {
	for _, item := range items {
		if strings.TrimSpace(item) != "" {
			return strings.TrimSpace(item)
		}
	}
	return ""
}
