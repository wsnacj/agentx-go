package productshell

import (
	"encoding/json"
	"strings"
)

type ShellBindingOptionPayload struct {
	PackID        string         `json:"pack_id,omitempty"`
	CaseType      string         `json:"case_type,omitempty"`
	WorkflowID    string         `json:"workflow_id,omitempty"`
	CaseID        string         `json:"case_id,omitempty"`
	CaseInput     map[string]any `json:"case_input,omitempty"`
	SessionInput  map[string]any `json:"session_input,omitempty"`
	WorkflowState map[string]any `json:"workflow_state,omitempty"`
	Persist       bool           `json:"persist,omitempty"`
}

func CloneShellBindingMap(input map[string]any) map[string]any {
	if len(input) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(input))
	for key, value := range input {
		out[key] = cloneShellBindingValue(value)
	}
	return out
}

func MergeShellBindingMaps(base, overlay map[string]any) map[string]any {
	out := CloneShellBindingMap(base)
	for key, value := range overlay {
		existingMap, existingOK := out[key].(map[string]any)
		valueMap, valueOK := value.(map[string]any)
		if existingOK && valueOK {
			out[key] = MergeShellBindingMaps(existingMap, valueMap)
			continue
		}
		out[key] = cloneShellBindingValue(value)
	}
	return out
}

func NormalizeShellBinding(binding ShellBinding) ShellBinding {
	binding.PackID = strings.TrimSpace(binding.PackID)
	binding.CaseType = strings.TrimSpace(binding.CaseType)
	binding.WorkflowID = strings.TrimSpace(binding.WorkflowID)
	binding.CaseID = strings.TrimSpace(binding.CaseID)
	binding.CaseInput = CloneShellBindingMap(binding.CaseInput)
	binding.SessionInput = CloneShellBindingMap(binding.SessionInput)
	binding.WorkflowState = CloneShellBindingMap(binding.WorkflowState)
	return binding
}

func ShellBindingHasValues(binding ShellBinding) bool {
	return strings.TrimSpace(binding.PackID) != "" || strings.TrimSpace(binding.CaseType) != "" ||
		strings.TrimSpace(binding.WorkflowID) != "" || strings.TrimSpace(binding.CaseID) != "" ||
		len(binding.CaseInput) > 0 || len(binding.SessionInput) > 0 || len(binding.WorkflowState) > 0
}

func CloneRequestedShellBinding(value *RequestedShellBinding) *RequestedShellBinding {
	if value == nil {
		return nil
	}
	cloned := *value
	cloned.Binding = NormalizeShellBinding(cloned.Binding)
	return &cloned
}

func EncodeRequestedShellBindingOption(value RequestedShellBinding) ShellBindingOptionPayload {
	binding := NormalizeShellBinding(value.Binding)
	return ShellBindingOptionPayload{
		PackID: binding.PackID, CaseType: binding.CaseType, WorkflowID: binding.WorkflowID, CaseID: binding.CaseID,
		CaseInput: CloneShellBindingMap(binding.CaseInput), SessionInput: CloneShellBindingMap(binding.SessionInput),
		WorkflowState: CloneShellBindingMap(binding.WorkflowState), Persist: value.PersistRequested,
	}
}

func DecodeShellBindingOption(raw any) (ShellBinding, bool, bool, error) {
	switch value := raw.(type) {
	case nil:
		return ShellBinding{}, false, false, nil
	case ShellBinding:
		binding := NormalizeShellBinding(value)
		return binding, ShellBindingHasValues(binding), false, nil
	case *ShellBinding:
		if value == nil {
			return ShellBinding{}, false, false, nil
		}
		binding := NormalizeShellBinding(*value)
		return binding, ShellBindingHasValues(binding), false, nil
	case ShellBindingOptionPayload:
		binding, hasValues, persist := decodeShellBindingPayload(value)
		return binding, hasValues, persist, nil
	case *ShellBindingOptionPayload:
		if value == nil {
			return ShellBinding{}, false, false, nil
		}
		binding, hasValues, persist := decodeShellBindingPayload(*value)
		return binding, hasValues, persist, nil
	case string:
		if strings.TrimSpace(value) == "" {
			return ShellBinding{}, false, false, nil
		}
		return decodeShellBindingJSON([]byte(value))
	case []byte:
		if len(value) == 0 {
			return ShellBinding{}, false, false, nil
		}
		return decodeShellBindingJSON(value)
	default:
		payload, err := json.Marshal(value)
		if err != nil {
			return ShellBinding{}, false, false, err
		}
		return decodeShellBindingJSON(payload)
	}
}

func decodeShellBindingJSON(raw []byte) (ShellBinding, bool, bool, error) {
	var payload ShellBindingOptionPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ShellBinding{}, false, false, err
	}
	binding, hasValues, persist := decodeShellBindingPayload(payload)
	return binding, hasValues, persist, nil
}

func decodeShellBindingPayload(payload ShellBindingOptionPayload) (ShellBinding, bool, bool) {
	binding := NormalizeShellBinding(payload.binding())
	return binding, ShellBindingHasValues(binding), payload.Persist
}

func (p ShellBindingOptionPayload) binding() ShellBinding {
	return ShellBinding{
		PackID: p.PackID, CaseType: p.CaseType, WorkflowID: p.WorkflowID, CaseID: p.CaseID,
		CaseInput: CloneShellBindingMap(p.CaseInput), SessionInput: CloneShellBindingMap(p.SessionInput),
		WorkflowState: CloneShellBindingMap(p.WorkflowState),
	}
}

func cloneShellBindingValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return CloneShellBindingMap(typed)
	case []any:
		out := make([]any, 0, len(typed))
		for _, item := range typed {
			out = append(out, cloneShellBindingValue(item))
		}
		return out
	default:
		return typed
	}
}
