package tools

import (
	"reflect"
	"strings"
	"sync"
	"testing"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

type testSessionRunRegistry struct {
	mu   sync.RWMutex
	runs map[string][]BrowserSessionRunInfo
}

func sliceContainsString(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}

func readStringFromMap(values map[string]any, key string) string {
	value, _ := values[key].(string)
	return value
}

func assertRequiredFields(t *testing.T, schema map[string]any, want []string) {
	t.Helper()
	got, ok := schema["required"].([]string)
	if !ok || !reflect.DeepEqual(got, want) {
		t.Fatalf("required fields = %#v, want %#v", schema["required"], want)
	}
}

func assertSchemaProperties(t *testing.T, schema map[string]any, names []string) {
	t.Helper()
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("properties = %#v", schema["properties"])
	}
	for _, name := range names {
		raw, ok := properties[name]
		if !ok {
			t.Fatalf("missing schema property %q", name)
		}
		property, ok := raw.(map[string]any)
		if !ok || strings.TrimSpace(readStringFromMap(property, "description")) == "" {
			t.Fatalf("schema property %q has no description: %#v", name, raw)
		}
	}
}

func browserBuiltinSharedFallbackOutcomeForTest() *agentxbrowserruntime.BrowserElementResolverOutcome {
	return agentxbrowserruntime.BrowserElementResolverOutcomeFromResult(agentxbrowserruntime.BrowserElementResolutionResult{
		ResolutionMode: "locator_plan_only",
		PrimaryKind:    "label",
		AttemptCount:   2,
		MatchedAttempt: &agentxbrowserruntime.BrowserElementResolutionAttempt{
			Index: 1,
			Candidate: agentxbrowserruntime.BrowserLocatorCandidate{
				Kind:        "placeholder",
				Placeholder: "Search docs",
				Tag:         "input",
				Type:        "text",
			},
		},
		FallbackFromAttempt: &agentxbrowserruntime.BrowserElementResolutionAttempt{
			Index:     0,
			IsPrimary: true,
			Candidate: agentxbrowserruntime.BrowserLocatorCandidate{
				Kind:  "label",
				Label: "Search",
				Tag:   "input",
				Type:  "text",
			},
		},
		FallbackFromOutcome: &agentxbrowserruntime.BrowserElementResolverOutcome{
			Status:            "unresolved",
			BlockedBy:         "multiple_candidates_filtered",
			AmbiguityClass:    "filtered_residual",
			CandidateKind:     "label",
			CandidateStrength: "medium",
			ManualRetryHint:   "add_ordinal",
			SpecificityFields: []string{"tag", "type"},
		},
	}, nil)
}

func newTestSessionRunRegistry() *testSessionRunRegistry {
	return &testSessionRunRegistry{runs: map[string][]BrowserSessionRunInfo{}}
}

func (registry *testSessionRunRegistry) Record(sessionID string, run BrowserSessionRunInfo) {
	if registry == nil {
		return
	}
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return
	}
	registry.mu.Lock()
	defer registry.mu.Unlock()
	registry.runs[sessionID] = append(registry.runs[sessionID], run)
}

func (registry *testSessionRunRegistry) SnapshotSessionRuns(sessionID string) []BrowserSessionRunInfo {
	if registry == nil {
		return nil
	}
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	return append([]BrowserSessionRunInfo(nil), registry.runs[strings.TrimSpace(sessionID)]...)
}
