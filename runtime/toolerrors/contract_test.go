package toolerrors

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestConstantsContract(t *testing.T) {
	got := []string{
		ToolArgumentErrorCodeInvalidJSON,
		ToolArgumentErrorCodeInvalidArgumentObject,
		ToolArgumentErrorCodeInvalidArgument,
		ToolArgumentErrorCodeMissingRequiredArgument,
		ToolArgumentRepairReturnValidJSONObject,
		ToolArgumentRepairProvideRequiredField,
		ToolArgumentRepairFixInvalidField,
		ToolArgumentRepairUseAliasURL,
	}
	want := []string{
		"invalid_json",
		"invalid_argument_object",
		"invalid_argument",
		"missing_required_argument",
		"return_valid_json_object",
		"provide_required_field",
		"fix_invalid_field",
		"use_alias_url",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("constants = %#v, want %#v", got, want)
	}
}

func TestToolArgumentRepairReflectContract(t *testing.T) {
	typ := reflect.TypeFor[ToolArgumentRepair]()
	wantNames := []string{"Kind", "From", "To"}
	wantTags := []string{"kind,omitempty", "from,omitempty", "to,omitempty"}
	if typ.NumField() != len(wantNames) {
		t.Fatalf("field count = %d, want %d", typ.NumField(), len(wantNames))
	}
	for i := range wantNames {
		field := typ.Field(i)
		if field.Name != wantNames[i] ||
			field.Type != reflect.TypeFor[string]() ||
			field.Tag.Get("json") != wantTags[i] {
			t.Fatalf("field[%d] = %s %s json:%q", i, field.Name, field.Type, field.Tag.Get("json"))
		}
	}
}

func TestToolArgumentErrorTextAndCauseContract(t *testing.T) {
	var nilError *ToolArgumentError
	if got := nilError.Error(); got != "invalid arguments" {
		t.Fatalf("nil Error() = %q", got)
	}
	if nilError.Unwrap() != nil {
		t.Fatal("nil Unwrap() is not nil")
	}

	cause := errors.New("private cause")
	err := &ToolArgumentError{Detail: " explicit detail ", Cause: cause}
	if got := err.Error(); got != "explicit detail" {
		t.Fatalf("detail Error() = %q", got)
	}
	if !errors.Is(err, cause) {
		t.Fatal("errors.Is() lost cause")
	}

	err.Detail = " "
	if got := err.Error(); got != cause.Error() {
		t.Fatalf("cause Error() = %q", got)
	}
	err.Cause = nil
	if got := err.Error(); got != "invalid arguments" {
		t.Fatalf("fallback Error() = %q", got)
	}
}

func TestNewToolArgumentErrorTrimsAndClones(t *testing.T) {
	cause := errors.New("cause")
	missing := []string{" a "}
	invalid := []string{"b"}
	disallowed := []string{"c"}
	repairs := []ToolArgumentRepair{{Kind: "repair", To: "a"}}
	err := NewToolArgumentError(" tool ", ToolArgumentErrorOptions{
		Code:             " code ",
		Detail:           " detail ",
		Repairable:       true,
		SafeAutorepair:   true,
		MissingFields:    missing,
		InvalidFields:    invalid,
		DisallowedFields: disallowed,
		AllowedRepairs:   repairs,
		Cause:            cause,
	})
	typed, ok := AsToolArgumentError(fmt.Errorf("outer: %w", err))
	if !ok {
		t.Fatalf("AsToolArgumentError() failed for %T", err)
	}
	if typed.Tool != "tool" || typed.Code != "code" || typed.Detail != "detail" ||
		!typed.Repairable || !typed.SafeAutorepair || !errors.Is(err, cause) {
		t.Fatalf("typed error = %#v", typed)
	}

	missing[0] = "changed"
	invalid[0] = "changed"
	disallowed[0] = "changed"
	repairs[0].Kind = "changed"
	if !reflect.DeepEqual(typed.MissingFields, []string{" a "}) ||
		!reflect.DeepEqual(typed.InvalidFields, []string{"b"}) ||
		!reflect.DeepEqual(typed.DisallowedFields, []string{"c"}) ||
		!reflect.DeepEqual(typed.AllowedRepairs, []ToolArgumentRepair{{Kind: "repair", To: "a"}}) {
		t.Fatalf("constructor retained caller slices: %#v", typed)
	}
}

func TestInvalidJSONContract(t *testing.T) {
	t.Run("nil cause", func(t *testing.T) {
		typed := mustToolArgumentError(t, NewInvalidJSONToolArgumentError(" tool ", nil))
		if typed.Tool != "tool" ||
			typed.Code != ToolArgumentErrorCodeInvalidJSON ||
			typed.Detail != "decode tool args: invalid json" ||
			!typed.Repairable ||
			typed.SafeAutorepair ||
			typed.Cause != nil ||
			!reflect.DeepEqual(typed.AllowedRepairs, []ToolArgumentRepair{{
				Kind: ToolArgumentRepairReturnValidJSONObject,
			}}) {
			t.Fatalf("invalid JSON error = %#v", typed)
		}
	})

	t.Run("non object", func(t *testing.T) {
		cause := errors.New("decode: TOP-LEVEL JSON OBJECT IS REQUIRED")
		typed := mustToolArgumentError(t, NewInvalidJSONToolArgumentError("tool", cause))
		if typed.Code != ToolArgumentErrorCodeInvalidArgumentObject ||
			typed.Detail != cause.Error() ||
			!errors.Is(typed, cause) {
			t.Fatalf("non-object error = %#v", typed)
		}
	})
}

func TestInvalidAndMissingFieldContracts(t *testing.T) {
	fields := []string{" first ", "", "second", "first", " second "}

	invalid := mustToolArgumentError(t, NewInvalidToolArgumentError(" tool ", fields, " "))
	if invalid.Error() != "tool: invalid arguments" ||
		invalid.Code != ToolArgumentErrorCodeInvalidArgument ||
		!invalid.Repairable ||
		invalid.SafeAutorepair ||
		!reflect.DeepEqual(invalid.InvalidFields, []string{"first", "second"}) ||
		!reflect.DeepEqual(invalid.AllowedRepairs, []ToolArgumentRepair{{
			Kind: ToolArgumentRepairFixInvalidField,
			To:   "first,second",
		}}) {
		t.Fatalf("invalid error = %#v", invalid)
	}

	singular := mustToolArgumentError(t, NewInvalidToolArgumentError(" tool ", []string{"field"}, ""))
	if singular.Error() != "tool: field is invalid" {
		t.Fatalf("singular invalid Error() = %q", singular.Error())
	}

	missing := mustToolArgumentError(t, NewMissingRequiredToolArgumentError(" tool ", fields, " "))
	if missing.Error() != "tool: required arguments are missing" ||
		missing.Code != ToolArgumentErrorCodeMissingRequiredArgument ||
		!missing.Repairable ||
		missing.SafeAutorepair ||
		!reflect.DeepEqual(missing.MissingFields, []string{"first", "second"}) ||
		!reflect.DeepEqual(missing.AllowedRepairs, []ToolArgumentRepair{{
			Kind: ToolArgumentRepairProvideRequiredField,
			To:   "first,second",
		}}) {
		t.Fatalf("missing error = %#v", missing)
	}

	singular = mustToolArgumentError(t, NewMissingRequiredToolArgumentError(" tool ", []string{"field"}, ""))
	if singular.Error() != "tool: field is required" {
		t.Fatalf("singular missing Error() = %q", singular.Error())
	}
}

func mustToolArgumentError(t *testing.T, err error) *ToolArgumentError {
	t.Helper()
	typed, ok := AsToolArgumentError(err)
	if !ok {
		t.Fatalf("AsToolArgumentError(%T) = false", err)
	}
	return typed
}
