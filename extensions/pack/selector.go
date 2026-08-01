package pack

import (
	"sort"
	"strings"
	"unicode"

	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

type SelectOptions struct {
	MaxCandidates int `json:"max_candidates,omitempty"`
	MinScore      int `json:"min_score,omitempty"`
}

type RouteSelectionCandidate struct {
	PackID           string   `json:"pack_id,omitempty"`
	CaseType         string   `json:"case_type,omitempty"`
	WorkflowID       string   `json:"workflow_id,omitempty"`
	WorkflowTitle    string   `json:"workflow_title,omitempty"`
	Score            int      `json:"score,omitempty"`
	MatchedHints     []string `json:"matched_hints,omitempty"`
	MatchedFragments []string `json:"matched_fragments,omitempty"`
	Reasons          []string `json:"reasons,omitempty"`
	defaultWorkflow  bool
}

type RouteSelection struct {
	Message        string                    `json:"message,omitempty"`
	Attempted      bool                      `json:"attempted,omitempty"`
	Matched        bool                      `json:"matched,omitempty"`
	Ambiguous      bool                      `json:"ambiguous,omitempty"`
	Threshold      int                       `json:"threshold,omitempty"`
	CandidateCount int                       `json:"candidate_count,omitempty"`
	Selected       RouteSelectionCandidate   `json:"selected"`
	Candidates     []RouteSelectionCandidate `json:"candidates,omitempty"`
}

func SelectBinding(reg Registry, message string, opts SelectOptions) (RouteSelection, bool) {
	selection := RouteSelection{
		Message:   message,
		Attempted: message != "",
	}
	if reg == nil || !selection.Attempted {
		return selection, false
	}
	defs := reg.List()
	if len(defs) == 0 {
		return selection, false
	}
	threshold := opts.MinScore
	if threshold <= 0 {
		threshold = 24
	}
	maxCandidates := opts.MaxCandidates
	if maxCandidates <= 0 {
		maxCandidates = 5
	}
	selection.Threshold = threshold
	normalizedMessage := normalizeRouteText(message)
	if normalizedMessage == "" {
		return selection, false
	}
	candidates := make([]RouteSelectionCandidate, 0)
	for _, def := range defs {
		candidates = append(candidates, routeCandidatesForDefinition(def, normalizedMessage)...)
	}
	if len(candidates) == 0 {
		return selection, false
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Score != candidates[j].Score {
			return candidates[i].Score > candidates[j].Score
		}
		if candidates[i].PackID == candidates[j].PackID && candidates[i].defaultWorkflow != candidates[j].defaultWorkflow {
			return candidates[i].defaultWorkflow
		}
		if candidates[i].PackID != candidates[j].PackID {
			return candidates[i].PackID < candidates[j].PackID
		}
		if candidates[i].CaseType != candidates[j].CaseType {
			return candidates[i].CaseType < candidates[j].CaseType
		}
		return candidates[i].WorkflowID < candidates[j].WorkflowID
	})
	filtered := make([]RouteSelectionCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.Score <= 0 {
			continue
		}
		filtered = append(filtered, candidate)
	}
	selection.CandidateCount = len(filtered)
	if len(filtered) == 0 {
		return selection, false
	}
	if len(filtered) > maxCandidates {
		selection.Candidates = append([]RouteSelectionCandidate(nil), filtered[:maxCandidates]...)
	} else {
		selection.Candidates = append([]RouteSelectionCandidate(nil), filtered...)
	}
	top := filtered[0]
	selection.Selected = top
	if top.Score < threshold {
		return selection, false
	}
	if routeSelectionIsAmbiguous(filtered) {
		selection.Ambiguous = true
		return selection, false
	}
	selection.Matched = true
	return selection, true
}

func routeCandidatesForDefinition(def Definition, normalizedMessage string) []RouteSelectionCandidate {
	workflows := def.Workflows
	if len(workflows) == 0 {
		return nil
	}
	candidates := make([]RouteSelectionCandidate, 0, len(workflows))
	for _, spec := range workflows {
		caseTypes := workflowRouteCaseTypes(def.Manifest, spec)
		for _, caseType := range caseTypes {
			candidate := scoreRouteCandidate(def, spec, caseType, normalizedMessage)
			if candidate.Score <= 0 {
				continue
			}
			candidates = append(candidates, candidate)
		}
	}
	return candidates
}

func workflowRouteCaseTypes(manifest Manifest, spec agentxworkflow.Spec) []string {
	if len(spec.CaseTypes) > 0 {
		out := make([]string, 0, len(spec.CaseTypes))
		for _, caseType := range spec.CaseTypes {
			if isCanonicalRouteIdentifier(caseType) {
				out = append(out, caseType)
			}
		}
		return out
	}
	out := make([]string, 0, len(manifest.SupportedCaseTypes))
	for _, caseType := range manifest.SupportedCaseTypes {
		if isCanonicalRouteIdentifier(caseType) {
			out = append(out, caseType)
		}
	}
	return out
}

func scoreRouteCandidate(def Definition, spec agentxworkflow.Spec, caseType string, normalizedMessage string) RouteSelectionCandidate {
	candidate := RouteSelectionCandidate{
		PackID:          def.Manifest.ID,
		CaseType:        caseType,
		WorkflowID:      spec.ID,
		WorkflowTitle:   spec.Title,
		defaultWorkflow: spec.ID == def.Manifest.DefaultWorkflow,
	}
	seenReasons := map[string]bool{}
	seenHints := map[string]bool{}
	seenFragments := map[string]bool{}
	addReason := func(reason string) {
		if reason == "" || seenReasons[reason] {
			return
		}
		seenReasons[reason] = true
		candidate.Reasons = append(candidate.Reasons, reason)
	}
	addHint := func(hint string, score int, reason string) {
		if hint == "" {
			return
		}
		normalizedHint := normalizeRouteText(hint)
		if normalizedHint == "" || !strings.Contains(normalizedMessage, normalizedHint) {
			return
		}
		if !seenHints[hint] {
			seenHints[hint] = true
			candidate.MatchedHints = append(candidate.MatchedHints, hint)
		}
		candidate.Score += score
		addReason(reason)
	}
	addFragment := func(fragment string, score int, reason string) {
		fragment = normalizeRouteText(fragment)
		if fragment == "" || !strings.Contains(normalizedMessage, fragment) {
			return
		}
		if !seenFragments[fragment] {
			seenFragments[fragment] = true
			candidate.MatchedFragments = append(candidate.MatchedFragments, fragment)
			candidate.Score += score
		}
		addReason(reason)
	}
	matchExactID := func(value string, score int, reason string) {
		normalizedValue := canonicalRouteIdentifier(value)
		if normalizedValue == "" || !strings.Contains(normalizedMessage, normalizedValue) {
			return
		}
		candidate.Score += score
		addReason(reason)
	}

	matchExactID(candidate.PackID, 32, "matched pack id")
	matchExactID(candidate.CaseType, 28, "matched case type")
	matchExactID(candidate.WorkflowID, 28, "matched workflow id")
	for _, hint := range def.Manifest.RouteHints {
		addHint(hint, 20, "matched pack route hint")
	}
	if schema, ok := def.CaseSchemaByType(caseType); ok {
		for _, hint := range schema.RouteHints {
			addHint(hint, 24, "matched case route hint")
		}
		for _, fragment := range routeFragments(schema.Description) {
			addFragment(fragment, 4, "matched case description")
		}
	}
	for _, hint := range spec.RouteHints {
		addHint(hint, 26, "matched workflow route hint")
	}
	for _, fragment := range routeFragments(spec.Title) {
		addFragment(fragment, 5, "matched workflow title")
	}
	for _, fragment := range routeFragments(spec.Description) {
		addFragment(fragment, 3, "matched workflow description")
	}
	for _, fragment := range routeFragments(def.Manifest.Domain) {
		addFragment(fragment, 2, "matched pack domain")
	}
	return candidate
}

func routeSelectionIsAmbiguous(candidates []RouteSelectionCandidate) bool {
	if len(candidates) < 2 {
		return false
	}
	top := candidates[0]
	for _, candidate := range candidates[1:] {
		if candidate.Score != top.Score {
			return false
		}
		if top.PackID == candidate.PackID && top.defaultWorkflow && !candidate.defaultWorkflow {
			continue
		}
		return true
	}
	return false
}

func normalizeRouteText(raw string) string {
	trimmed := strings.TrimSpace(strings.ToLower(raw))
	if trimmed == "" {
		return ""
	}
	var b strings.Builder
	lastSpace := false
	for _, r := range trimmed {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r), unicode.Is(unicode.Han, r):
			b.WriteRune(r)
			lastSpace = false
		default:
			if !lastSpace {
				b.WriteByte(' ')
				lastSpace = true
			}
		}
	}
	return strings.TrimSpace(b.String())
}

func canonicalRouteIdentifier(raw string) string {
	if !isCanonicalRouteIdentifier(raw) {
		return ""
	}
	return normalizeRouteText(raw)
}

func isCanonicalRouteIdentifier(raw string) bool {
	return raw != "" && raw == strings.TrimSpace(raw)
}

func routeFragments(raw string) []string {
	normalized := normalizeRouteText(raw)
	if normalized == "" {
		return nil
	}
	parts := strings.Fields(normalized)
	if len(parts) == 0 {
		return nil
	}
	out := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if len([]rune(part)) <= 1 {
			continue
		}
		if seen[part] {
			continue
		}
		seen[part] = true
		out = append(out, part)
	}
	return out
}
