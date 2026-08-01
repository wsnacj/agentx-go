package cases_test

import (
	"context"
	"encoding/json"
	"testing"

	cases "github.com/wsnacj/agentx-go/runtime/cases"
)

var _ cases.Store = externalStore{}

func TestExternalCaseNormalizeCloneAndNilContract(t *testing.T) {
	original := cases.Case{
		ID:            " case-1 ",
		Type:          " research ",
		PackID:        " pack-1 ",
		WorkflowID:    " workflow-1 ",
		Source:        " user ",
		Intent:        " investigate ",
		SessionID:     " session-1 ",
		PolicyProfile: " safe ",
		MemorySchema:  " evidence ",
		Status:        " active ",
		Inputs: map[string]any{
			"query":  " raw ",
			"nested": map[string]any{"items": []any{"one", map[string]any{"value": "two"}}},
		},
		Outcome:   nil,
		CreatedAt: 10,
		UpdatedAt: 20,
	}
	normalized := cases.Normalize(original)
	if normalized.ID != "case-1" || normalized.Type != "research" || normalized.PackID != "pack-1" || normalized.Status != "active" {
		t.Fatalf("Normalize() = %#v", normalized)
	}
	if normalized.Inputs["query"] != " raw " {
		t.Fatalf("Normalize() changed opaque map value: %#v", normalized.Inputs)
	}
	if normalized.Outcome == nil || len(normalized.Outcome) != 0 {
		t.Fatalf("Normalize() nil Outcome = %#v, want non-nil empty map", normalized.Outcome)
	}
	originalNested := original.Inputs["nested"].(map[string]any)
	originalNested["items"].([]any)[1].(map[string]any)["value"] = "mutated"
	gotNested := normalized.Inputs["nested"].(map[string]any)
	if gotNested["items"].([]any)[1].(map[string]any)["value"] != "two" {
		t.Fatalf("Normalize() did not deep-clone maps/slices: %#v", normalized.Inputs)
	}

	cloned := cases.Clone(&normalized)
	if cloned == nil || cloned == &normalized {
		t.Fatalf("Clone() = %#v", cloned)
	}
	cloned.Inputs["query"] = "changed"
	if normalized.Inputs["query"] != " raw " {
		t.Fatalf("Clone() leaked map mutation: %#v", normalized.Inputs)
	}
	if cases.Clone(nil) != nil {
		t.Fatal("Clone(nil) must return nil")
	}
	if !cases.IsZero(cases.Case{ID: " ", Inputs: nil, Outcome: map[string]any{}}) {
		t.Fatal("whitespace-only identity and empty maps must be zero")
	}
	if cases.IsZero(cases.Case{CreatedAt: 1}) {
		t.Fatal("non-zero timestamp must not be zero")
	}
}

func TestExternalCaseJSONContract(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want string
	}{
		{name: "zero_case", in: cases.Case{}, want: `{}`},
		{
			name: "case",
			in: cases.Case{
				ID:         "case-1",
				Type:       "research",
				PackID:     "pack-1",
				WorkflowID: "workflow-1",
				Inputs:     map[string]any{"query": "risk"},
				CreatedAt:  10,
			},
			want: `{"id":"case-1","type":"research","pack_id":"pack-1","workflow_id":"workflow-1","inputs":{"query":"risk"},"created_at":10}`,
		},
		{
			name: "filter",
			in:   cases.Filter{PackID: "pack-1", CaseType: "research", Status: "active", Limit: 20},
			want: `{"pack_id":"pack-1","case_type":"research","status":"active","limit":20}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			payload, err := json.Marshal(tc.in)
			if err != nil {
				t.Fatalf("Marshal(): %v", err)
			}
			if string(payload) != tc.want {
				t.Fatalf("JSON = %s, want %s", payload, tc.want)
			}
		})
	}
}

type externalStore struct{}

func (externalStore) UpsertCase(context.Context, cases.Case) error {
	return nil
}

func (externalStore) GetCase(context.Context, string) (cases.Case, error) {
	return cases.Case{}, nil
}

func (externalStore) ListCases(context.Context, cases.Filter) ([]cases.Case, error) {
	return nil, nil
}
