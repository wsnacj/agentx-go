package productshell

import (
	"strings"

	agentxcases "github.com/wsnacj/agentx-go/runtime/cases"
	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"

	"github.com/wsnacj/agentx-go/extensions/pack"
	"github.com/wsnacj/agentx-go/extensions/skills"
)

func ApplyInputCase(input Input) Input {
	input.Case = MergeRequestedCase(input.Case, input.RequestedCaseID, input.RequestedCaseInput)
	input.RequestedCaseID = ""
	input.RequestedCaseInput = map[string]any{}
	if input.Case != nil {
		input = applyCaseToInput(input, input.Case)
	}
	input.Options = stripLegacyCaseIngressOptions(input.Options)
	return input
}

func ApplyEffectiveCase(input Input, value agentxcases.Case) Input {
	return applyCaseToInput(input, &value)
}

func ApplyResolvedWorkflow(input Input, value agentxworkflow.Spec) Input {
	spec := value
	input.WorkflowSpec = &spec
	return input
}

func MergeRequestedCase(base *agentxcases.Case, requestedCaseID string, requestedCaseInput map[string]any) *agentxcases.Case {
	current := agentxcases.Case{}
	hasCase := false
	if base != nil {
		current = agentxcases.Normalize(*base)
		hasCase = !agentxcases.IsZero(current)
	}
	if requestedCaseID = strings.TrimSpace(requestedCaseID); requestedCaseID != "" {
		current.ID = firstNonEmptyString(current.ID, requestedCaseID)
		hasCase = true
	}
	if len(requestedCaseInput) > 0 {
		current.Inputs = MergeShellBindingMaps(current.Inputs, requestedCaseInput)
		hasCase = true
	}
	if !hasCase {
		return nil
	}
	resolved := agentxcases.Normalize(current)
	if agentxcases.IsZero(resolved) {
		return nil
	}
	return &resolved
}

func ApplyCommandDispatch(input Input, prepared *PreparedCommandDispatch) Input {
	if prepared == nil || !prepared.Matched {
		return input
	}
	input.UserMessage = strings.TrimSpace(prepared.UserMessage)
	if skill := strings.TrimSpace(prepared.Skill); skill != "" {
		input.ShellOptions.RequestedSkills = uniqueLowerOptionStrings(
			append(append([]string(nil), input.ShellOptions.RequestedSkills...), skill),
		)
	}
	if semantic := prepared.RequestedSkillSemantic; semantic.Name != "" {
		input.ShellOptions.RequestedSkillSemantics = skills.MergeRequestedSkillSemantics(
			input.ShellOptions.RequestedSkillSemantics,
			[]RequestedSkillSemantic{semantic},
		)
	}
	input.Options = stripRequestedSkillOptions(input.Options)
	return input
}

// ParseRequestedSkills parses only explicit typed, option, and directive
// requests. Product-specific inferred skills remain a host policy.
func ParseRequestedSkills(input Input) ([]string, string) {
	requested := append([]string(nil), input.ShellOptions.RequestedSkills...)
	requested = uniqueLowerOptionStrings(append(requested, parseRequestedSkillsOptionValues(input.Options)...))
	directiveSkills, cleanedMessage := parseRequestedSkillsDirective(input.UserMessage)
	requested = uniqueLowerOptionStrings(append(requested, directiveSkills...))
	cleanedMessage = strings.TrimSpace(cleanedMessage)
	if cleanedMessage == "" && strings.TrimSpace(input.UserMessage) != "" {
		cleanedMessage = strings.TrimSpace(input.UserMessage)
	}
	return requested, cleanedMessage
}

func ParseSkillActivationPaths(input Input) []string {
	paths := append([]string(nil), input.ShellOptions.SkillActivationPaths...)
	return uniqueTrimmedOptionStrings(append(paths, parseSkillActivationPathsOptionValues(input.Options)...))
}

func SkillActivationPathsFromSessionInput(input map[string]any) []string {
	if explicit := ExplicitSkillActivationPathsFromSessionInput(input); len(explicit) > 0 {
		return explicit
	}
	return TouchedFileActivationPathsFromSessionInput(input)
}

func ExplicitSkillActivationPathsFromSessionInput(input map[string]any) []string {
	return activationPathsForKeys(input, []string{
		"skill_activation_path", "skill_activation_paths", "skillActivationPath", "skillActivationPaths",
		"focused_path", "focused_paths", "focusedPath", "focusedPaths",
	})
}

func TouchedFileActivationPathsFromSessionInput(input map[string]any) []string {
	return activationPathsForKeys(input, []string{"files_touched", "filesTouched"})
}

func activationPathsForKeys(input map[string]any, keys []string) []string {
	var out []string
	for _, key := range keys {
		if raw, ok := input[key]; ok {
			out = append(out, parseSkillActivationPathOptionValue(raw)...)
		}
	}
	return uniqueTrimmedOptionStrings(out)
}

func ApplyPackSelection(input Input, prepared *PreparedPackSelection) Input {
	if prepared == nil || !prepared.Applied {
		return input
	}
	input.PackID = firstNonEmptyString(input.PackID, prepared.Binding.PackID)
	input.CaseType = firstNonEmptyString(input.CaseType, prepared.Binding.CaseType)
	input.PackWorkflow = firstNonEmptyString(input.PackWorkflow, prepared.Binding.WorkflowID)
	return input
}

func ApplyCaseBinding(input Input, prepared *PreparedCaseBinding) Input {
	if prepared == nil || !prepared.Applied {
		return input
	}
	input.PackID = firstNonEmptyString(input.PackID, prepared.Binding.PackID)
	input.CaseType = firstNonEmptyString(input.CaseType, prepared.Binding.CaseType)
	input.PackWorkflow = firstNonEmptyString(input.PackWorkflow, prepared.Binding.WorkflowID)
	current := agentxcases.Case{}
	if input.Case != nil {
		current = agentxcases.Normalize(*input.Case)
	}
	current.PackID = firstNonEmptyString(current.PackID, input.PackID, prepared.Binding.PackID)
	current.Type = firstNonEmptyString(current.Type, input.CaseType, prepared.Binding.CaseType)
	current.WorkflowID = firstNonEmptyString(current.WorkflowID, input.PackWorkflow, prepared.Binding.WorkflowID)
	if len(prepared.Merged) > 0 {
		current.Inputs = CloneShellBindingMap(prepared.Merged)
	}
	if !agentxcases.IsZero(current) {
		resolved := agentxcases.Normalize(current)
		input.Case = &resolved
	}
	return input
}

func PackSelectionMetricsFromPrepared(prepared *PreparedPackSelection) PackSelectionMetrics {
	if prepared == nil {
		return PackSelectionMetrics{}
	}
	metrics := PackSelectionMetrics{
		Attempted: prepared.Selection.Attempted, Matched: prepared.Matched, Applied: prepared.Applied,
		Ambiguous: prepared.Selection.Ambiguous, Threshold: prepared.Selection.Threshold,
		CandidateCount: prepared.Selection.CandidateCount, Message: prepared.Selection.Message,
		SkipReason: strings.TrimSpace(prepared.SkipReason), Selected: packSelectionCandidateMetrics(prepared.Selection.Selected),
	}
	for _, candidate := range prepared.Selection.Candidates {
		metrics.Candidates = append(metrics.Candidates, packSelectionCandidateMetrics(candidate))
	}
	return metrics
}

func RequestedSkillsOrFallback(input Input, fallback []string) []string {
	if len(input.ShellOptions.RequestedSkills) > 0 {
		return uniqueLowerOptionStrings(input.ShellOptions.RequestedSkills)
	}
	return uniqueLowerOptionStrings(fallback)
}

func SkillActivationPathsOrFallback(input Input, fallback []string) []string {
	if len(input.ShellOptions.SkillActivationPaths) > 0 {
		return uniqueTrimmedOptionStrings(input.ShellOptions.SkillActivationPaths)
	}
	return uniqueTrimmedOptionStrings(fallback)
}

func applyCaseToInput(input Input, value *agentxcases.Case) Input {
	if value == nil {
		return input
	}
	resolved := agentxcases.Normalize(*value)
	input.Case = &resolved
	input.PackID = firstNonEmptyString(resolved.PackID, input.PackID)
	input.CaseType = firstNonEmptyString(resolved.Type, input.CaseType)
	input.PackWorkflow = firstNonEmptyString(resolved.WorkflowID, input.PackWorkflow)
	return input
}

func packSelectionCandidateMetrics(candidate pack.RouteSelectionCandidate) PackSelectionCandidateMetrics {
	return PackSelectionCandidateMetrics{
		PackID: strings.TrimSpace(candidate.PackID), CaseType: strings.TrimSpace(candidate.CaseType),
		WorkflowID: strings.TrimSpace(candidate.WorkflowID), WorkflowTitle: strings.TrimSpace(candidate.WorkflowTitle),
		Score: candidate.Score, MatchedHints: append([]string(nil), candidate.MatchedHints...),
		MatchedFragments: append([]string(nil), candidate.MatchedFragments...), Reasons: append([]string(nil), candidate.Reasons...),
	}
}

func parseRequestedSkillsDirective(message string) ([]string, string) {
	original := strings.TrimSpace(message)
	remaining := original
	var out []string
	for {
		if !strings.HasPrefix(strings.ToLower(remaining), "[skill:") {
			break
		}
		end := strings.Index(remaining, "]")
		if end <= len("[skill:") {
			break
		}
		out = append(out, parseRequestedSkillOptionNames(remaining[len("[skill:"):end])...)
		remaining = strings.TrimSpace(remaining[end+1:])
	}
	if len(out) == 0 {
		return nil, original
	}
	return uniqueLowerOptionStrings(out), remaining
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}

func stripLegacyCaseIngressOptions(options map[string]any) map[string]any {
	out := cloneAnyMap(options)
	for _, key := range []string{"case_id", "caseId", "case_input", "caseInput"} {
		delete(out, key)
	}
	return out
}

func stripRequestedSkillOptions(options map[string]any) map[string]any {
	out := cloneAnyMap(options)
	for _, key := range []string{"skill", "skills", "requested_skill", "requested_skills", "requestedSkill", "requestedSkills"} {
		delete(out, key)
	}
	return out
}

func stripSkillActivationPathOptions(options map[string]any) map[string]any {
	out := cloneAnyMap(options)
	for _, key := range []string{"skill_activation_path", "skill_activation_paths", "skillActivationPath", "skillActivationPaths", "focused_path", "focused_paths", "focusedPath", "focusedPaths"} {
		delete(out, key)
	}
	return out
}
