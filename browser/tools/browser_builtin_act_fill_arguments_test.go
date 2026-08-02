package tools

import (
	"context"
	"reflect"
	"testing"

	types "github.com/wsnacj/agentx-go/components/llm"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

func TestRegisterBrowserTools_ActFillReturnsStructuredMissingValueError(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		capabilities: BrowserCapabilities{
			Fill: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"fill","selector":"input[name=email]","text":"a@example.com"}`,
	})
	if err == nil {
		t.Fatalf("expected browser_act fill missing value error")
	}
	argErr, ok := AsToolArgumentError(err)
	if !ok {
		t.Fatalf("expected structured argument error, got %T %v", err, err)
	}
	if argErr.Code != "missing_value" || !argErr.Repairable || !argErr.SafeAutorepair {
		t.Fatalf("unexpected structured argument error: %#v", argErr)
	}
	if argErr.Error() != "browser_act: value or values is required for kind fill" {
		t.Fatalf("unexpected error detail: %q", argErr.Error())
	}
	if !reflect.DeepEqual(argErr.MissingFields, []string{"value_or_values"}) {
		t.Fatalf("unexpected missing fields: %#v", argErr.MissingFields)
	}
	if got, want := browserRepairAdviceKinds(argErr.AllowedRepairs), []string{"use_alias_value"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected repair advice: got=%#v want=%#v", got, want)
	}
}

func TestRegisterBrowserTools_ActFillReturnsStructuredMissingValueErrorForContentAlias(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		capabilities: BrowserCapabilities{
			Fill: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"fill","selector":"input[name=email]","content":"a@example.com"}`,
	})
	if err == nil {
		t.Fatalf("expected browser_act fill missing value error")
	}
	argErr, ok := AsToolArgumentError(err)
	if !ok {
		t.Fatalf("expected structured argument error, got %T %v", err, err)
	}
	if argErr.Code != "missing_value" || !argErr.Repairable || !argErr.SafeAutorepair {
		t.Fatalf("unexpected structured argument error: %#v", argErr)
	}
	if argErr.Error() != "browser_act: value or values is required for kind fill" {
		t.Fatalf("unexpected error detail: %q", argErr.Error())
	}
	if !reflect.DeepEqual(argErr.MissingFields, []string{"value_or_values"}) {
		t.Fatalf("unexpected missing fields: %#v", argErr.MissingFields)
	}
	if got, want := browserRepairAdviceKinds(argErr.AllowedRepairs), []string{"use_alias_value"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected repair advice: got=%#v want=%#v", got, want)
	}
}

func TestRegisterBrowserTools_ActFillFieldsReturnsStructuredMissingValueError(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		capabilities: BrowserCapabilities{
			Fill: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"fill","fields":[{"selector":"input[name=email]","text":"a@example.com"}]}`,
	})
	if err == nil {
		t.Fatalf("expected browser_act fill fields missing value error")
	}
	argErr, ok := AsToolArgumentError(err)
	if !ok {
		t.Fatalf("expected structured argument error, got %T %v", err, err)
	}
	if argErr.Code != "missing_value" || !argErr.Repairable || !argErr.SafeAutorepair {
		t.Fatalf("unexpected structured argument error: %#v", argErr)
	}
	if argErr.Error() != "browser_act: value or values is required for kind fill" {
		t.Fatalf("unexpected error detail: %q", argErr.Error())
	}
	if !reflect.DeepEqual(argErr.MissingFields, []string{"value_or_values"}) {
		t.Fatalf("unexpected missing fields: %#v", argErr.MissingFields)
	}
	if got, want := browserRepairAdviceKinds(argErr.AllowedRepairs), []string{"use_alias_value"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected repair advice: got=%#v want=%#v", got, want)
	}
}

func TestRegisterBrowserTools_ActFillReturnsStructuredInvalidFieldsShapeError(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		capabilities: BrowserCapabilities{
			Fill: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"fill","fields":{"selector":"input[name=email]","text":"a@example.com"}}`,
	})
	if err == nil {
		t.Fatalf("expected browser_act fill invalid fields shape error")
	}
	argErr, ok := AsToolArgumentError(err)
	if !ok {
		t.Fatalf("expected structured argument error, got %T %v", err, err)
	}
	if argErr.Code != "invalid_fill_fields_shape" || !argErr.Repairable || !argErr.SafeAutorepair {
		t.Fatalf("unexpected structured argument error: %#v", argErr)
	}
	if argErr.Error() != "browser_act: fields must be an array of objects for kind fill" {
		t.Fatalf("unexpected error detail: %q", argErr.Error())
	}
	if !reflect.DeepEqual(argErr.MissingFields, []string{"fields_array"}) {
		t.Fatalf("unexpected missing fields: %#v", argErr.MissingFields)
	}
	if got, want := browserRepairAdviceKinds(argErr.AllowedRepairs), []string{"wrap_singleton_field", "use_alias_value"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected repair advice: got=%#v want=%#v", got, want)
	}
}

func TestRegisterBrowserTools_ActFillReturnsStructuredInvalidFieldsShapeErrorForStringifiedFields(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		capabilities: BrowserCapabilities{
			Fill: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: "{\"kind\":\"fill\",\"fields\":\"[{\\\"selector\\\":\\\"input[name=email]\\\",\\\"text\\\":\\\"a@example.com\\\"}]\"}",
	})
	if err == nil {
		t.Fatalf("expected browser_act fill invalid fields shape error")
	}
	argErr, ok := AsToolArgumentError(err)
	if !ok {
		t.Fatalf("expected structured argument error, got %T %v", err, err)
	}
	if argErr.Code != "invalid_fill_fields_shape" || !argErr.Repairable || !argErr.SafeAutorepair {
		t.Fatalf("unexpected structured argument error: %#v", argErr)
	}
	if got, want := browserRepairAdviceKinds(argErr.AllowedRepairs), []string{"parse_stringified_fields", "use_alias_value"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected repair advice: got=%#v want=%#v", got, want)
	}
}

func TestRegisterBrowserTools_ActFillReturnsStructuredMissingFillInputErrorForSingularFieldAlias(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		capabilities: BrowserCapabilities{
			Fill: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"fill","field":{"selector":"input[name=email]","text":"a@example.com"}}`,
	})
	if err == nil {
		t.Fatalf("expected browser_act fill singular field alias error")
	}
	argErr, ok := AsToolArgumentError(err)
	if !ok {
		t.Fatalf("expected structured argument error, got %T %v", err, err)
	}
	if argErr.Code != "missing_fill_input" || !argErr.Repairable || !argErr.SafeAutorepair {
		t.Fatalf("unexpected structured argument error: %#v", argErr)
	}
	if argErr.Error() != "browser_act: fields or selector/ref plus value is required for kind fill" {
		t.Fatalf("unexpected error detail: %q", argErr.Error())
	}
	if !reflect.DeepEqual(argErr.MissingFields, []string{"fields_or_locator_plus_value"}) {
		t.Fatalf("unexpected missing fields: %#v", argErr.MissingFields)
	}
	if got, want := browserRepairAdviceKinds(argErr.AllowedRepairs), []string{"promote_singular_field", "use_alias_value"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected repair advice: got=%#v want=%#v", got, want)
	}
}

func TestRegisterBrowserTools_ActFillReturnsStructuredMissingFillInputErrorForStringifiedSingularFieldAlias(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		capabilities: BrowserCapabilities{
			Fill: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: "{\"kind\":\"fill\",\"field\":\"{\\\"selector\\\":\\\"input[name=email]\\\",\\\"text\\\":\\\"a@example.com\\\"}\"}",
	})
	if err == nil {
		t.Fatalf("expected browser_act fill stringified singular field alias error")
	}
	argErr, ok := AsToolArgumentError(err)
	if !ok {
		t.Fatalf("expected structured argument error, got %T %v", err, err)
	}
	if argErr.Code != "missing_fill_input" || !argErr.Repairable || !argErr.SafeAutorepair {
		t.Fatalf("unexpected structured argument error: %#v", argErr)
	}
	if got, want := browserRepairAdviceKinds(argErr.AllowedRepairs), []string{"promote_singular_field", "use_alias_value"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected repair advice: got=%#v want=%#v", got, want)
	}
}

func TestRegisterBrowserTools_ActFillTextOnlyReturnsStructuredMissingLocatorError(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		capabilities: BrowserCapabilities{
			Fill: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"fill","text":"a@example.com"}`,
	})
	if err == nil {
		t.Fatalf("expected browser_act fill text-only missing locator error")
	}
	argErr, ok := AsToolArgumentError(err)
	if !ok {
		t.Fatalf("expected structured argument error, got %T %v", err, err)
	}
	if argErr.Code != "missing_locator" {
		t.Fatalf("unexpected structured argument error: %#v", argErr)
	}
}

func TestRegisterBrowserTools_ActFillReturnsStructuredMissingLocatorErrorForLabelHint(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		capabilities: BrowserCapabilities{
			Fill: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"fill","label":"Email","text":"a@example.com"}`,
	})
	if err == nil {
		t.Fatalf("expected browser_act fill missing locator error")
	}
	argErr, ok := AsToolArgumentError(err)
	if !ok {
		t.Fatalf("expected structured argument error, got %T %v", err, err)
	}
	if argErr.Code != "missing_locator" || !argErr.Repairable || !argErr.SafeAutorepair {
		t.Fatalf("unexpected structured argument error: %#v", argErr)
	}
	if argErr.Error() != "browser_act: selector or ref is required for kind fill" {
		t.Fatalf("unexpected error detail: %q", argErr.Error())
	}
	if !reflect.DeepEqual(argErr.MissingFields, []string{"selector_or_ref"}) {
		t.Fatalf("unexpected missing fields: %#v", argErr.MissingFields)
	}
	if got, want := browserRepairAdviceKinds(argErr.AllowedRepairs), []string{"use_alias_field", "use_declared_hint", "use_alias_value"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected repair advice: got=%#v want=%#v", got, want)
	}
	if len(backend.fillReqs) != 0 {
		t.Fatalf("expected browser_act fill missing locator to block before backend dispatch, got %#v", backend.fillReqs)
	}
}
