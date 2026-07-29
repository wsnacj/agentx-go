package safeerror

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

type contractCause struct{}

func (contractCause) Error() string {
	return "contract cause"
}

func TestIdentityContract(t *testing.T) {
	if got := Identity("  material  "); got != "40b30b4e8f0d137056ac497e859ea198c1a00db4267d1ade9c458d04024e2981" {
		t.Fatalf("Identity() = %q", got)
	}
	if got := Identity(" \t "); got != "" {
		t.Fatalf("Identity(empty) = %q", got)
	}
}

func TestWrapperContract(t *testing.T) {
	if got := Wrap(nil, "message"); got != nil {
		t.Fatalf("Wrap(nil) = %v", got)
	}

	cause := contractCause{}
	wrapped := Wrap(cause, "  display safe  ")
	if wrapped.Error() != "display safe" {
		t.Fatalf("Error() = %q", wrapped.Error())
	}
	if !errors.Is(wrapped, cause) {
		t.Fatalf("errors.Is() = false")
	}
	var target contractCause
	if !errors.As(wrapped, &target) {
		t.Fatalf("errors.As() = false")
	}

	fallback := Wrap(cause, " \t ")
	if fallback.Error() != "safe error" {
		t.Fatalf("fallback Error() = %q", fallback.Error())
	}

	explicit := WrapWithIdentity(cause, "failed", " explicit-id ")
	if got := Project(explicit, "runtime", "failed").Identity; got != "explicit-id" {
		t.Fatalf("explicit identity = %q", got)
	}

	var nilWrapper *wrappedError
	if nilWrapper.Error() != "safe error" || nilWrapper.Unwrap() != nil {
		t.Fatalf("nil wrapper contract drifted")
	}
}

func TestProjectionAndJSONContract(t *testing.T) {
	projection := Project(nil, " Runtime / Error ", " UPSTREAM//FAILED ")
	if projection != (Projection{Class: "runtime_error", Code: "upstream_failed"}) {
		t.Fatalf("Project(nil) = %#v", projection)
	}

	long := strings.Repeat("A", 80)
	if got := ProjectText("", long, "").Class; got != strings.Repeat("a", 64) {
		t.Fatalf("normalized class = %q", got)
	}
	if got := ProjectText(" material ", "event", "bad").Identity; got != Identity("material") {
		t.Fatalf("ProjectText identity = %q", got)
	}

	payload, err := json.Marshal(Projection{Class: "runtime", Code: "failed", Identity: "id-1"})
	if err != nil {
		t.Fatalf("Marshal(): %v", err)
	}
	if string(payload) != `{"class":"runtime","code":"failed","identity":"id-1"}` {
		t.Fatalf("JSON = %s", payload)
	}
	empty, err := json.Marshal(Projection{})
	if err != nil {
		t.Fatalf("Marshal(empty): %v", err)
	}
	if string(empty) != `{}` {
		t.Fatalf("empty JSON = %s", empty)
	}
}

func TestSummaryAndMapContract(t *testing.T) {
	projection := Projection{Class: "runtime", Code: "failed", Identity: "id-1"}
	if got := Summary(projection); got != "class=runtime code=failed identity=id-1" {
		t.Fatalf("Summary() = %q", got)
	}
	if got := Summary(Projection{}); got != "class=error code=unknown" {
		t.Fatalf("Summary(empty) = %q", got)
	}

	attrs := map[string]any{"keep": "value"}
	gotAttrs := AppendAttrs(attrs, " tool_", projection)
	if gotAttrs["keep"] != "value" ||
		gotAttrs["tool_error_class"] != "runtime" ||
		gotAttrs["tool_error_code"] != "failed" ||
		gotAttrs["tool_error_identity"] != "id-1" {
		t.Fatalf("AppendAttrs() = %#v", gotAttrs)
	}

	details := AppendDetails(nil, " tool_", projection)
	if details["tool_error_class"] != "runtime" ||
		details["tool_error_code"] != "failed" ||
		details["tool_error_identity"] != "id-1" {
		t.Fatalf("AppendDetails() = %#v", details)
	}
}
