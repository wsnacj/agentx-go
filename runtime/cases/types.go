package cases

import "strings"

type Case struct {
	ID            string         `json:"id,omitempty"`
	Type          string         `json:"type,omitempty"`
	PackID        string         `json:"pack_id,omitempty"`
	WorkflowID    string         `json:"workflow_id,omitempty"`
	Source        string         `json:"source,omitempty"`
	Intent        string         `json:"intent,omitempty"`
	SessionID     string         `json:"session_id,omitempty"`
	PolicyProfile string         `json:"policy_profile,omitempty"`
	MemorySchema  string         `json:"memory_schema,omitempty"`
	Status        string         `json:"status,omitempty"`
	Inputs        map[string]any `json:"inputs,omitempty"`
	Outcome       map[string]any `json:"outcome,omitempty"`
	CreatedAt     int64          `json:"created_at,omitempty"`
	UpdatedAt     int64          `json:"updated_at,omitempty"`
}

type Filter struct {
	PackID   string `json:"pack_id,omitempty"`
	CaseType string `json:"case_type,omitempty"`
	Status   string `json:"status,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}

func Normalize(value Case) Case {
	value.ID = strings.TrimSpace(value.ID)
	value.Type = strings.TrimSpace(value.Type)
	value.PackID = strings.TrimSpace(value.PackID)
	value.WorkflowID = strings.TrimSpace(value.WorkflowID)
	value.Source = strings.TrimSpace(value.Source)
	value.Intent = strings.TrimSpace(value.Intent)
	value.SessionID = strings.TrimSpace(value.SessionID)
	value.PolicyProfile = strings.TrimSpace(value.PolicyProfile)
	value.MemorySchema = strings.TrimSpace(value.MemorySchema)
	value.Status = strings.TrimSpace(value.Status)
	value.Inputs = cloneMap(value.Inputs)
	value.Outcome = cloneMap(value.Outcome)
	return value
}

func Clone(value *Case) *Case {
	if value == nil {
		return nil
	}
	out := Normalize(*value)
	return &out
}

func IsZero(value Case) bool {
	value = Normalize(value)
	return value.ID == "" &&
		value.Type == "" &&
		value.PackID == "" &&
		value.WorkflowID == "" &&
		value.Source == "" &&
		value.Intent == "" &&
		value.SessionID == "" &&
		value.PolicyProfile == "" &&
		value.MemorySchema == "" &&
		value.Status == "" &&
		len(value.Inputs) == 0 &&
		len(value.Outcome) == 0 &&
		value.CreatedAt == 0 &&
		value.UpdatedAt == 0
}

func cloneMap(raw map[string]any) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(raw))
	for key, value := range raw {
		out[key] = cloneValue(value)
	}
	return out
}

func cloneValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneMap(typed)
	case []any:
		out := make([]any, 0, len(typed))
		for _, item := range typed {
			out = append(out, cloneValue(item))
		}
		return out
	default:
		return typed
	}
}
