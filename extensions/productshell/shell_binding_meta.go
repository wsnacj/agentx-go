package productshell

import (
	"encoding/json"
	"sort"
	"strings"

	agentxcases "github.com/wsnacj/agentx-go/runtime/cases"

	"github.com/wsnacj/agentx-go/extensions/pack"
)

const sessionShellBindingMetaKey = "agentx_shell_binding"

func LoadSessionShellBindingMetaJSON(metaJSON string) (ShellBinding, bool, error) {
	if strings.TrimSpace(metaJSON) == "" {
		return ShellBinding{}, false, nil
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal([]byte(metaJSON), &payload); err != nil {
		return ShellBinding{}, false, err
	}
	raw, ok := payload[sessionShellBindingMetaKey]
	if !ok || len(raw) == 0 {
		return ShellBinding{}, false, nil
	}
	var binding ShellBinding
	if err := json.Unmarshal(raw, &binding); err != nil {
		return ShellBinding{}, false, err
	}
	binding = NormalizeShellBinding(binding)
	if !ShellBindingHasValues(binding) {
		return ShellBinding{}, false, nil
	}
	return binding, true, nil
}

func MergeSessionShellBindingMetaJSON(metaJSON string, binding ShellBinding) (string, error) {
	payload := map[string]json.RawMessage{}
	if strings.TrimSpace(metaJSON) != "" {
		if err := json.Unmarshal([]byte(metaJSON), &payload); err != nil {
			return "", err
		}
	}
	binding = NormalizeShellBinding(binding)
	if !ShellBindingHasValues(binding) {
		delete(payload, sessionShellBindingMetaKey)
	} else {
		raw, err := json.Marshal(binding)
		if err != nil {
			return "", err
		}
		payload[sessionShellBindingMetaKey] = raw
	}
	if len(payload) == 0 {
		return "", nil
	}
	raw, err := json.Marshal(payload)
	return string(raw), err
}

func BuildShellBindingMetrics(source string, binding ShellBinding, persist bool) ShellBindingMetrics {
	binding = NormalizeShellBinding(binding)
	return ShellBindingMetrics{
		Source: strings.TrimSpace(source), Matched: ShellBindingHasValues(binding), PersistRequested: persist,
		PackID: binding.PackID, CaseType: binding.CaseType, WorkflowID: binding.WorkflowID, CaseID: binding.CaseID,
		HasCaseInput: len(binding.CaseInput) > 0, HasSessionInput: len(binding.SessionInput) > 0,
		HasWorkflowState: len(binding.WorkflowState) > 0, CaseInputKeys: sortedShellBindingMapKeys(binding.CaseInput),
		SessionInputKeys: sortedShellBindingMapKeys(binding.SessionInput), WorkflowStateKeys: sortedShellBindingMapKeys(binding.WorkflowState),
	}
}

func PreparedShellBindingFromRequested(value *RequestedShellBinding, source string) *PreparedShellBinding {
	if value == nil {
		return nil
	}
	binding := NormalizeShellBinding(value.Binding)
	if !ShellBindingHasValues(binding) {
		return nil
	}
	return &PreparedShellBinding{
		Binding: binding, Source: strings.TrimSpace(source), Matched: true,
		PersistRequested: value.PersistRequested, Metrics: BuildShellBindingMetrics(source, binding, value.PersistRequested),
	}
}

func PreparedShellBindingFromStored(binding ShellBinding, source string) *PreparedShellBinding {
	binding = NormalizeShellBinding(binding)
	if !ShellBindingHasValues(binding) {
		return nil
	}
	return &PreparedShellBinding{Binding: binding, Source: strings.TrimSpace(source), Matched: true, Metrics: BuildShellBindingMetrics(source, binding, false)}
}

func FinalizeShellBindingMetrics(metrics ShellBindingMetrics, input Input, binding *pack.Binding, resolved *PreparedShellBinding) ShellBindingMetrics {
	if resolved != nil {
		metrics.Source = firstNonEmptyString(metrics.Source, resolved.Source)
		metrics.Matched = resolved.Matched
		metrics.PersistRequested = resolved.PersistRequested
	}
	metrics.PackID = firstNonEmptyString(input.PackID, metrics.PackID)
	metrics.CaseType = firstNonEmptyString(input.CaseType, metrics.CaseType)
	metrics.WorkflowID = firstNonEmptyString(input.PackWorkflow, metrics.WorkflowID)
	if input.Case != nil {
		metrics.PackID = firstNonEmptyString(input.Case.PackID, metrics.PackID)
		metrics.CaseType = firstNonEmptyString(input.Case.Type, metrics.CaseType)
		metrics.WorkflowID = firstNonEmptyString(input.Case.WorkflowID, metrics.WorkflowID)
		metrics.CaseID = firstNonEmptyString(input.Case.ID, metrics.CaseID)
		if len(input.Case.Inputs) > 0 {
			metrics.HasCaseInput = true
			metrics.CaseInputKeys = sortedShellBindingMapKeys(input.Case.Inputs)
		}
	}
	if binding != nil {
		metrics.PackID = firstNonEmptyString(binding.PackID, metrics.PackID)
		metrics.CaseType = firstNonEmptyString(binding.CaseType, metrics.CaseType)
		metrics.WorkflowID = firstNonEmptyString(binding.WorkflowID, metrics.WorkflowID)
	}
	return metrics
}

func EffectiveShellBindingFromInput(input Input, binding *pack.Binding) (ShellBinding, bool) {
	resolved := ShellBinding{PackID: strings.TrimSpace(input.PackID), CaseType: strings.TrimSpace(input.CaseType), WorkflowID: strings.TrimSpace(input.PackWorkflow)}
	if input.Case != nil {
		current := agentxcases.Normalize(*input.Case)
		resolved.PackID = firstNonEmptyString(current.PackID, resolved.PackID)
		resolved.CaseType = firstNonEmptyString(current.Type, resolved.CaseType)
		resolved.WorkflowID = firstNonEmptyString(current.WorkflowID, resolved.WorkflowID)
		resolved.CaseID = firstNonEmptyString(current.ID, resolved.CaseID)
		resolved.CaseInput = CloneShellBindingMap(current.Inputs)
	}
	if binding != nil {
		resolved.PackID = firstNonEmptyString(binding.PackID, resolved.PackID)
		resolved.CaseType = firstNonEmptyString(binding.CaseType, resolved.CaseType)
		resolved.WorkflowID = firstNonEmptyString(binding.WorkflowID, resolved.WorkflowID)
	}
	resolved.SessionInput = CloneShellBindingMap(input.SessionInput)
	resolved.WorkflowState = CloneShellBindingMap(input.WorkflowState)
	resolved = NormalizeShellBinding(resolved)
	return resolved, ShellBindingHasValues(resolved)
}

func ApplyShellBindingToInput(input Input, prepared *PreparedShellBinding) (Input, error) {
	if prepared == nil || !prepared.Matched || input.WorkflowSpec != nil || ContainsWorkflowSpecOption(input.Options) {
		return input, nil
	}
	binding := NormalizeShellBinding(prepared.Binding)
	if !ShellBindingHasValues(binding) {
		return input, nil
	}
	input.PackID = firstNonEmptyString(input.PackID, binding.PackID)
	input.CaseType = firstNonEmptyString(input.CaseType, binding.CaseType)
	input.PackWorkflow = firstNonEmptyString(input.PackWorkflow, binding.WorkflowID)
	current := agentxcases.Case{}
	if input.Case != nil {
		current = agentxcases.Normalize(*input.Case)
	}
	current.PackID = firstNonEmptyString(current.PackID, input.PackID, binding.PackID)
	current.Type = firstNonEmptyString(current.Type, input.CaseType, binding.CaseType)
	current.WorkflowID = firstNonEmptyString(current.WorkflowID, input.PackWorkflow, binding.WorkflowID)
	current.ID = firstNonEmptyString(current.ID, binding.CaseID)
	if len(binding.CaseInput) > 0 {
		current.Inputs = MergeShellBindingMaps(binding.CaseInput, current.Inputs)
	}
	if !agentxcases.IsZero(current) {
		resolved := agentxcases.Normalize(current)
		input.Case = &resolved
	}
	input.SessionInput = MergeShellBindingMaps(binding.SessionInput, input.SessionInput)
	input.WorkflowState = MergeShellBindingMaps(binding.WorkflowState, input.WorkflowState)
	return input, nil
}

func sortedShellBindingMapKeys(values map[string]any) []string {
	if len(values) == 0 {
		return nil
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		if key = strings.TrimSpace(key); key != "" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}
