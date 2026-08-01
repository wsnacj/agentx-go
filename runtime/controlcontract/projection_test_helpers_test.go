package controlcontract

import (
	"encoding/json"
	"strings"
	"testing"
)

type testProjection[T ~string] struct {
	Name         string
	RunnerEffect string
	PromptEffect string
	Boundaries   []T
	Payload      any
}

func assertHostOwnedProjectionOnly[T ~string](t testing.TB, projection testProjection[T], requiredBoundaries ...string) {
	t.Helper()
	if strings.TrimSpace(projection.RunnerEffect) != "none" || strings.TrimSpace(projection.PromptEffect) != "none" {
		t.Fatalf("%s must not affect runner or prompt, runner_effect=%q prompt_effect=%q", projection.Name, projection.RunnerEffect, projection.PromptEffect)
	}
	want := append([]string{"display_safe_refs_only", "no_runner_dispatch"}, requiredBoundaries...)
	for _, required := range want {
		found := false
		for _, boundary := range projection.Boundaries {
			if string(boundary) == required {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("%s missing boundary %q in %#v", projection.Name, required, projection.Boundaries)
		}
	}
}

func assertNoRawPayload(t testing.TB, name string, payload any, rejected ...string) {
	t.Helper()
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("%s marshal payload: %v", name, err)
	}
	for _, value := range rejected {
		if value != "" && strings.Contains(string(raw), value) {
			t.Fatalf("%s leaked raw content %q in %s", name, value, raw)
		}
	}
}

func controlTokenListContains(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func productionAdapterStringContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func productionAdapterMissingContains(values []MissingInput, want MissingInput) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func productionAdapterBoundaryContains(values []Boundary, want Boundary) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
