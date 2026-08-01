package productshell

import (
	"encoding/json"
	"strings"

	"github.com/wsnacj/agentx-go/extensions/skills"
)

var typedInputOptionKeys = stringSet(
	"auto_case_binding", "autoCaseBinding", "auto_workflow_planning", "autoWorkflowPlanning",
	"case_binding_auto", "caseBindingAuto", "workflow_planning_auto", "workflowPlanningAuto",
	"focused_path", "focused_paths", "focusedPath", "focusedPaths",
	"requestedSkillSemantics", "requested_skill_semantics", "requestedSkill", "requestedSkills",
	"requested_skill", "requested_skills", "skill", "skills",
	"skill_activation_path", "skill_activation_paths", "skillActivationPath", "skillActivationPaths",
)

var inputOptionKeys = stringSet(
	"case_type", "caseType", "pack", "pack_id", "packId", "pack_workflow_id", "packWorkflowId",
	"product_shell", "productShell", "session_input", "sessionInput", "shell_binding", "shellBinding",
	"state", "workflow", "workflow_id", "workflowId", "workflow_spec", "workflowSpec",
	"workflow_state", "workflowState",
)

func stringSet(values ...string) map[string]struct{} {
	out := make(map[string]struct{}, len(values))
	for _, value := range values {
		out[value] = struct{}{}
	}
	return out
}

func ContainsWorkflowSpecOption(options map[string]any) bool {
	for _, key := range []string{"workflow_spec", "workflowSpec", "workflow"} {
		raw, ok := options[key]
		if !ok || raw == nil {
			continue
		}
		switch value := raw.(type) {
		case string:
			if strings.TrimSpace(value) != "" {
				return true
			}
		case []byte:
			if len(value) > 0 {
				return true
			}
		default:
			return true
		}
	}
	return false
}

func ParseInputMapOption(options map[string]any, keys ...string) (map[string]any, bool, error) {
	raw := firstInputOption(options, keys...)
	if raw == nil {
		return nil, false, nil
	}
	value, err := decodeInputMapOption(raw)
	if err != nil {
		return nil, false, err
	}
	return value, true, nil
}

func ParseRequestedShellBindingOption(options map[string]any) (*RequestedShellBinding, bool, error) {
	raw := firstInputOption(options, "shell_binding", "shellBinding")
	if raw == nil {
		return nil, false, nil
	}
	binding, _, persist, err := DecodeShellBindingOption(raw)
	if err != nil {
		return nil, false, err
	}
	if !ShellBindingHasValues(binding) && !persist {
		return nil, false, nil
	}
	return &RequestedShellBinding{Binding: binding, PersistRequested: persist}, true, nil
}

func ProjectInputOptions(raw map[string]any) (InputShellOptions, *RequestedShellBinding, map[string]any, map[string]any, map[string]any) {
	if len(raw) == 0 {
		return InputShellOptions{}, nil, nil, nil, map[string]any{}
	}
	options := InputShellOptions{
		AutoCaseBinding:      parseInputBoolOption(raw, "auto_case_binding", "autoCaseBinding", "case_binding_auto", "caseBindingAuto"),
		AutoWorkflowPlanning: parseInputBoolOption(raw, "auto_workflow_planning", "autoWorkflowPlanning", "workflow_planning_auto", "workflowPlanningAuto"),
		RequestedSkills:      parseRequestedSkillsOptionValues(raw), RequestedSkillSemantics: parseRequestedSkillSemanticsOptionValues(raw),
		SkillActivationPaths: parseSkillActivationPathsOptionValues(raw),
	}
	binding, hasBinding := parseRequestedShellBindingOptionIgnoringError(raw)
	sessionInput, hasSession := parseInputMapOptionIgnoringError(raw, "session_input", "sessionInput")
	workflowState, hasState := parseInputMapOptionIgnoringError(raw, "workflow_state", "workflowState", "state")
	residual := map[string]any{}
	for key, value := range raw {
		if _, ok := typedInputOptionKeys[key]; ok {
			continue
		}
		if hasBinding && (key == "shell_binding" || key == "shellBinding") {
			continue
		}
		if hasSession && (key == "session_input" || key == "sessionInput") {
			continue
		}
		if hasState && (key == "workflow_state" || key == "workflowState" || key == "state") {
			continue
		}
		if _, ok := inputOptionKeys[key]; ok {
			residual[key] = value
		}
	}
	return options, binding, sessionInput, workflowState, residual
}

func ComposeInputOptions(shell InputShellOptions, binding *RequestedShellBinding, sessionInput, workflowState, residual map[string]any) map[string]any {
	parsedShell, parsedBinding, parsedSession, parsedState, residualOptions := ProjectInputOptions(residual)
	out := cloneKnownInputOptions(residualOptions)
	mergeTypedInputOptions(out, parsedShell)
	mergeStructuredInputOptions(out, parsedBinding, parsedSession, parsedState)
	mergeTypedInputOptions(out, shell)
	mergeStructuredInputOptions(out, binding, sessionInput, workflowState)
	return out
}

func MergeInputOptions(base map[string]any, shell InputShellOptions, binding *RequestedShellBinding, sessionInput, workflowState, residual map[string]any) map[string]any {
	out := cloneAnyMap(base)
	parsedShell, parsedBinding, parsedSession, parsedState, residualOptions := ProjectInputOptions(residual)
	for key, value := range cloneKnownInputOptions(residualOptions) {
		out[key] = value
	}
	mergeTypedInputOptions(out, parsedShell)
	mergeStructuredInputOptions(out, parsedBinding, parsedSession, parsedState)
	mergeTypedInputOptions(out, shell)
	mergeStructuredInputOptions(out, binding, sessionInput, workflowState)
	return out
}

func firstInputOption(options map[string]any, keys ...string) any {
	for _, key := range keys {
		if value, ok := options[key]; ok && value != nil {
			return value
		}
	}
	return nil
}

func decodeInputMapOption(raw any) (map[string]any, error) {
	switch value := raw.(type) {
	case nil:
		return map[string]any{}, nil
	case map[string]any:
		return cloneAnyMap(value), nil
	case string:
		if strings.TrimSpace(value) == "" {
			return map[string]any{}, nil
		}
		return decodeInputMapJSON([]byte(value))
	case []byte:
		if len(value) == 0 {
			return map[string]any{}, nil
		}
		return decodeInputMapJSON(value)
	default:
		payload, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		return decodeInputMapJSON(payload)
	}
}

func decodeInputMapJSON(raw []byte) (map[string]any, error) {
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return cloneAnyMap(out), nil
}

func cloneAnyMap(input map[string]any) map[string]any {
	if len(input) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}

func cloneKnownInputOptions(raw map[string]any) map[string]any {
	out := map[string]any{}
	for key, value := range raw {
		if _, ok := inputOptionKeys[key]; ok {
			out[key] = value
		}
	}
	return out
}

func mergeTypedInputOptions(options map[string]any, shell InputShellOptions) {
	if shell.AutoCaseBinding != nil {
		options["auto_case_binding"] = *shell.AutoCaseBinding
	}
	if shell.AutoWorkflowPlanning != nil {
		options["auto_workflow_planning"] = *shell.AutoWorkflowPlanning
	}
	if len(shell.RequestedSkills) > 0 {
		options["requested_skills"] = append([]string(nil), shell.RequestedSkills...)
	}
	if len(shell.RequestedSkillSemantics) > 0 {
		options["requested_skill_semantics"] = cloneRequestedSkillSemantics(shell.RequestedSkillSemantics)
	}
	if len(shell.SkillActivationPaths) > 0 {
		options["skill_activation_paths"] = append([]string(nil), shell.SkillActivationPaths...)
	}
}

func mergeStructuredInputOptions(options map[string]any, binding *RequestedShellBinding, sessionInput, workflowState map[string]any) {
	if binding != nil {
		options["shell_binding"] = EncodeRequestedShellBindingOption(*binding)
	}
	if len(sessionInput) > 0 {
		options["session_input"] = CloneShellBindingMap(sessionInput)
	}
	if len(workflowState) > 0 {
		options["workflow_state"] = CloneShellBindingMap(workflowState)
	}
}

func parseInputBoolOption(options map[string]any, keys ...string) *bool {
	for _, key := range keys {
		raw, ok := options[key]
		if !ok || raw == nil {
			continue
		}
		switch value := raw.(type) {
		case bool:
			return boolPtr(value)
		case string:
			switch strings.ToLower(strings.TrimSpace(value)) {
			case "1", "true", "yes", "on":
				return boolPtr(true)
			case "0", "false", "no", "off":
				return boolPtr(false)
			}
		}
	}
	return nil
}

func boolPtr(value bool) *bool { return &value }

func parseInputMapOptionIgnoringError(options map[string]any, keys ...string) (map[string]any, bool) {
	value, ok, err := ParseInputMapOption(options, keys...)
	return value, ok && err == nil
}

func parseRequestedShellBindingOptionIgnoringError(options map[string]any) (*RequestedShellBinding, bool) {
	value, ok, err := ParseRequestedShellBindingOption(options)
	return value, ok && err == nil
}

func parseRequestedSkillsOptionValues(options map[string]any) []string {
	var out []string
	for _, key := range []string{"skill", "skills", "requested_skill", "requested_skills", "requestedSkill", "requestedSkills"} {
		if raw, ok := options[key]; ok {
			out = append(out, parseRequestedSkillOptionValue(raw)...)
		}
	}
	return uniqueLowerOptionStrings(out)
}

func parseRequestedSkillSemanticsOptionValues(options map[string]any) []RequestedSkillSemantic {
	var out []RequestedSkillSemantic
	for _, key := range []string{"requested_skill_semantics", "requestedSkillSemantics"} {
		if raw, ok := options[key]; ok {
			out = append(out, parseRequestedSkillSemanticOptionValue(raw)...)
		}
	}
	return skills.MergeRequestedSkillSemantics(out)
}

func parseSkillActivationPathsOptionValues(options map[string]any) []string {
	var out []string
	for _, key := range []string{"skill_activation_path", "skill_activation_paths", "skillActivationPath", "skillActivationPaths", "focused_path", "focused_paths", "focusedPath", "focusedPaths"} {
		if raw, ok := options[key]; ok {
			out = append(out, parseSkillActivationPathOptionValue(raw)...)
		}
	}
	return uniqueTrimmedOptionStrings(out)
}

func parseRequestedSkillOptionValue(raw any) []string {
	switch value := raw.(type) {
	case string:
		return parseRequestedSkillOptionNames(value)
	case []string:
		var out []string
		for _, item := range value {
			out = append(out, parseRequestedSkillOptionNames(item)...)
		}
		return uniqueLowerOptionStrings(out)
	case []any:
		var out []string
		for _, item := range value {
			out = append(out, parseRequestedSkillOptionValue(item)...)
		}
		return uniqueLowerOptionStrings(out)
	default:
		return nil
	}
}

func parseRequestedSkillSemanticOptionValue(raw any) []RequestedSkillSemantic {
	switch value := raw.(type) {
	case map[string]any:
		item := RequestedSkillSemantic{
			Name:             firstNonEmptyString(stringValue(value["name"]), stringValue(value["skill"])),
			ExecutionContext: skills.NormalizeSkillExecutionContext(firstNonEmptyString(stringValue(value["execution_context"]), stringValue(value["executionContext"]), stringValue(value["context"]))),
			AllowedTools:     skills.NormalizeSkillAllowedTools(firstNonEmptySlice(stringSliceValue(value["allowed_tools"]), stringSliceValue(value["allowedTools"]))),
			Effort:           skills.NormalizeSkillExecutionEffort(firstNonEmptyString(stringValue(value["effort"]), stringValue(value["reasoning_effort"]), stringValue(value["reasoningEffort"]))),
		}
		return skills.MergeRequestedSkillSemantics([]RequestedSkillSemantic{item})
	case []RequestedSkillSemantic:
		return skills.MergeRequestedSkillSemantics(value)
	case []map[string]any:
		var out []RequestedSkillSemantic
		for _, item := range value {
			out = append(out, parseRequestedSkillSemanticOptionValue(item)...)
		}
		return skills.MergeRequestedSkillSemantics(out)
	case []any:
		var out []RequestedSkillSemantic
		for _, item := range value {
			out = append(out, parseRequestedSkillSemanticOptionValue(item)...)
		}
		return skills.MergeRequestedSkillSemantics(out)
	default:
		return nil
	}
}

func parseSkillActivationPathOptionValue(raw any) []string {
	switch value := raw.(type) {
	case string:
		return parseCommaSeparated(value, false)
	case []string:
		var out []string
		for _, item := range value {
			out = append(out, parseCommaSeparated(item, false)...)
		}
		return uniqueTrimmedOptionStrings(out)
	case []any:
		var out []string
		for _, item := range value {
			out = append(out, parseSkillActivationPathOptionValue(item)...)
		}
		return uniqueTrimmedOptionStrings(out)
	default:
		return nil
	}
}

func parseRequestedSkillOptionNames(raw string) []string { return parseCommaSeparated(raw, true) }

func parseCommaSeparated(raw string, lower bool) []string {
	var out []string
	for _, item := range strings.Split(strings.TrimSpace(raw), ",") {
		item = strings.TrimSpace(item)
		if lower {
			item = strings.ToLower(item)
		}
		if item != "" {
			out = append(out, item)
		}
	}
	if lower {
		return uniqueLowerOptionStrings(out)
	}
	return uniqueTrimmedOptionStrings(out)
}

func uniqueLowerOptionStrings(values []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if value != "" && !seen[value] {
			seen[value] = true
			out = append(out, value)
		}
	}
	return out
}

func uniqueTrimmedOptionStrings(values []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" && !seen[value] {
			seen[value] = true
			out = append(out, value)
		}
	}
	return out
}

func cloneRequestedSkillSemantics(items []RequestedSkillSemantic) []RequestedSkillSemantic {
	out := make([]RequestedSkillSemantic, 0, len(items))
	for _, item := range items {
		cloned := item
		cloned.AllowedTools = append([]string(nil), item.AllowedTools...)
		out = append(out, cloned)
	}
	return out
}

func stringValue(value any) string { text, _ := value.(string); return text }

func stringSliceValue(value any) []string {
	switch typed := value.(type) {
	case string:
		return parseRequestedSkillOptionNames(typed)
	case []string:
		return append([]string(nil), typed...)
	case []any:
		var out []string
		for _, item := range typed {
			if text, ok := item.(string); ok {
				out = append(out, text)
			}
		}
		return out
	default:
		return nil
	}
}

func firstNonEmptySlice(values ...[]string) []string {
	for _, value := range values {
		if len(value) > 0 {
			return value
		}
	}
	return nil
}
