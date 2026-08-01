package controlcontract

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// Projection carries the common read-only control-plane surface shared by
// engine, workflow, and scene adapter tests.
type Projection[T ~string] struct {
	Name         string
	RunnerEffect string
	PromptEffect string
	Boundaries   []T
	Payload      any
}

// AssertProjectionOnly verifies that a control-plane surface stays display-only
// and carries the expected boundary markers. It intentionally knows nothing
// about engine, workflow, scene packages, or business-specific source kinds.
func AssertProjectionOnly[T ~string](t testing.TB, projection Projection[T], requiredBoundaries ...string) {
	t.Helper()
	AssertNoRunnerPromptEffect(t, projection.Name, projection.RunnerEffect, projection.PromptEffect)
	AssertBoundaries(t, projection.Name, projection.Boundaries, requiredBoundaries...)
}

// AssertProposalOnly is the common assertion for proposal-only diagnostics.
func AssertProposalOnly[T ~string](t testing.TB, projection Projection[T], requiredBoundaries ...string) {
	t.Helper()
	required := append([]string{"proposal_only"}, requiredBoundaries...)
	AssertProjectionOnly(t, projection, required...)
}

// AssertHostOwnedProjectionOnly verifies a host-owned control-plane surface is
// still projection-only and does not dispatch runner or adapter work by itself.
func AssertHostOwnedProjectionOnly[T ~string](t testing.TB, projection Projection[T], requiredBoundaries ...string) {
	t.Helper()
	required := append([]string{"display_safe_refs_only", "no_runner_dispatch"}, requiredBoundaries...)
	AssertProjectionOnly(t, projection, required...)
}

// AssertNoCoreMutation verifies host-reported execution did not become a core
// mutation claim.
func AssertNoCoreMutation(t testing.TB, name string, coreExecuted bool, durableWriteByCore bool) {
	t.Helper()
	if coreExecuted || durableWriteByCore {
		t.Fatalf("%s must not claim core execution or durable write, core_executed=%v durable_write_by_core=%v", testName(name), coreExecuted, durableWriteByCore)
	}
}

// AssertNoObjectiveSatisfied verifies a host-facing audit or handoff surface did
// not turn host-reported work into an objective-complete claim.
func AssertNoObjectiveSatisfied(t testing.TB, name string, objectiveSatisfied bool) {
	t.Helper()
	if objectiveSatisfied {
		t.Fatalf("%s must not claim objective satisfaction", testName(name))
	}
}

// AssertNoVerificationSatisfied verifies a handoff remains audit input until a
// separate verification result marks it satisfied.
func AssertNoVerificationSatisfied(t testing.TB, name string, verificationSatisfied bool) {
	t.Helper()
	if verificationSatisfied {
		t.Fatalf("%s must not claim verification satisfaction", testName(name))
	}
}

// AssertHostOwnedAuditInputOnly verifies a host-owned completion or readback
// surface stays display-only and cannot become objective completion by itself.
func AssertHostOwnedAuditInputOnly[T ~string](t testing.TB, projection Projection[T], objectiveSatisfied, verificationSatisfied, coreExecuted, durableWriteByCore bool, requiredBoundaries ...string) {
	t.Helper()
	AssertHostOwnedProjectionOnly(t, projection, requiredBoundaries...)
	AssertNoObjectiveSatisfied(t, projection.Name, objectiveSatisfied)
	AssertNoVerificationSatisfied(t, projection.Name, verificationSatisfied)
	AssertNoCoreMutation(t, projection.Name, coreExecuted, durableWriteByCore)
}

// AssertNoRunnerPromptEffect verifies a projection did not change runner or prompt behavior.
func AssertNoRunnerPromptEffect(t testing.TB, name, runnerEffect, promptEffect string) {
	t.Helper()
	if strings.TrimSpace(runnerEffect) != "none" || strings.TrimSpace(promptEffect) != "none" {
		t.Fatalf("%s must not affect runner or prompt, runner_effect=%q prompt_effect=%q", testName(name), runnerEffect, promptEffect)
	}
}

// AssertBoundaries verifies all required boundary tokens are present.
func AssertBoundaries[T ~string](t testing.TB, name string, boundaries []T, required ...string) {
	t.Helper()
	values := StringValues(boundaries)
	for _, want := range required {
		if !contains(values, want) {
			t.Fatalf("%s missing boundary %q in %#v", testName(name), want, values)
		}
	}
}

// AssertNoRawPayload verifies sensitive or raw substrings do not appear after JSON projection.
func AssertNoRawPayload(t testing.TB, name string, payload any, rejected ...string) {
	t.Helper()
	raw, err := marshalPayload(payload)
	if err != nil {
		t.Fatalf("%s marshal payload: %v", testName(name), err)
	}
	for _, value := range rejected {
		if value != "" && strings.Contains(raw, value) {
			t.Fatalf("%s leaked raw content %q in %s", testName(name), value, raw)
		}
	}
}

// StringValues converts any string-backed slice into plain strings.
func StringValues[T ~string](values []T) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, string(value))
	}
	return out
}

func marshalPayload(payload any) (string, error) {
	switch value := payload.(type) {
	case nil:
		return "", nil
	case string:
		return value, nil
	case []byte:
		return string(value), nil
	default:
		raw, err := json.Marshal(value)
		if err != nil {
			return "", err
		}
		return string(raw), nil
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func testName(name string) string {
	if strings.TrimSpace(name) == "" {
		return "projection"
	}
	return fmt.Sprintf("%s projection", name)
}
