package productshell

import (
	"encoding/json"
	"fmt"
	"strings"

	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

func DecodeWorkflowSpec(raw any) (*agentxworkflow.Spec, error) {
	switch value := raw.(type) {
	case nil:
		return nil, nil
	case *agentxworkflow.Spec:
		if value == nil {
			return nil, nil
		}
		spec := *value
		return &spec, nil
	case agentxworkflow.Spec:
		spec := value
		return &spec, nil
	case string:
		if strings.TrimSpace(value) == "" {
			return nil, nil
		}
		return decodeWorkflowJSON([]byte(value))
	case []byte:
		if len(value) == 0 {
			return nil, nil
		}
		return decodeWorkflowJSON(value)
	default:
		payload, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		return decodeWorkflowJSON(payload)
	}
}

func decodeWorkflowJSON(raw []byte) (*agentxworkflow.Spec, error) {
	var spec agentxworkflow.Spec
	if err := json.Unmarshal(raw, &spec); err != nil {
		return nil, err
	}
	return &spec, nil
}

func ExplicitRawWorkflowOptIn(input Input) bool {
	if input.RawWorkflowOptIn {
		return true
	}
	for _, key := range []string{"raw_workflow", "rawWorkflow"} {
		raw, ok := input.Options[key]
		if !ok || raw == nil {
			continue
		}
		switch value := raw.(type) {
		case bool:
			if value {
				return true
			}
		case string:
			if strings.EqualFold(strings.TrimSpace(value), "true") {
				return true
			}
		}
	}
	return false
}

func ResolveWorkflowBinding(input Input) (WorkflowBinding, bool, error) {
	casePackID, caseType, caseWorkflowID := "", "", ""
	if input.Case != nil {
		casePackID = strings.TrimSpace(input.Case.PackID)
		caseType = strings.TrimSpace(input.Case.Type)
		caseWorkflowID = strings.TrimSpace(input.Case.WorkflowID)
	}
	binding := WorkflowBinding{
		PackID:     firstNonEmptyString(casePackID, input.PackID, readWorkflowBindingOptionString(input.Options, "pack_id", "packId", "pack")),
		CaseType:   firstNonEmptyString(caseType, input.CaseType, readWorkflowBindingOptionString(input.Options, "case_type", "caseType")),
		WorkflowID: firstNonEmptyString(caseWorkflowID, input.PackWorkflow, readWorkflowBindingOptionString(input.Options, "pack_workflow_id", "packWorkflowId", "workflow_id", "workflowId")),
	}
	if binding.PackID == "" && binding.CaseType == "" && binding.WorkflowID == "" {
		return WorkflowBinding{}, false, nil
	}
	if binding.PackID == "" {
		return WorkflowBinding{}, false, fmt.Errorf("agentx: pack_id is required when binding workflow by pack")
	}
	if binding.CaseType == "" {
		return WorkflowBinding{}, false, fmt.Errorf("agentx: case_type is required when binding workflow by pack")
	}
	return binding, true, nil
}

func ExplicitWorkflowBindingCaseType(input Input, spec agentxworkflow.Spec) string {
	caseType := ""
	if input.Case != nil {
		caseType = strings.TrimSpace(input.Case.Type)
	}
	caseType = firstNonEmptyString(caseType, input.CaseType)
	if caseType == "" && len(spec.CaseTypes) == 1 {
		caseType = strings.TrimSpace(spec.CaseTypes[0])
	}
	return caseType
}

func readWorkflowBindingOptionString(options map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := firstInputOption(options, key).(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
