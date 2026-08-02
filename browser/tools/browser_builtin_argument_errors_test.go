package tools

import (
	"context"
	"reflect"
	"testing"

	types "github.com/wsnacj/agentx-go/components/llm"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

func TestBrowserLocatorRepairAdviceFromParamsUsesHintFields(t *testing.T) {
	repairable, safeAutorepair, repairs := browserLocatorRepairAdviceFromParams(map[string]any{
		"text": "Buy now",
	}, "element", "text", "label")
	if !repairable || !safeAutorepair {
		t.Fatalf("expected hint-driven repair advice to be marked safe repairable, got repairable=%v safe=%v", repairable, safeAutorepair)
	}
	if got, want := browserRepairAdviceKinds(repairs), []string{"use_alias_field", "use_declared_hint"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected repair advice: got=%#v want=%#v", got, want)
	}
}

func TestBrowserValueRepairAdviceFromParamsUsesTextAlias(t *testing.T) {
	repairable, safeAutorepair, repairs := browserValueRepairAdviceFromParams(map[string]any{
		"text": "a@example.com",
	}, "text")
	if !repairable || !safeAutorepair {
		t.Fatalf("expected value-alias repair advice to be marked safe repairable, got repairable=%v safe=%v", repairable, safeAutorepair)
	}
	if got, want := browserRepairAdviceKinds(repairs), []string{"use_alias_value"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected value repair advice: got=%#v want=%#v", got, want)
	}
}

func TestBrowserActFillFieldsTreatsStructuredMissingValueAsNoFields(t *testing.T) {
	fields, err := browserActFillFields(map[string]any{
		"selector": "input[name=email]",
	})
	if err != nil {
		t.Fatalf("browserActFillFields: %v", err)
	}
	if len(fields) != 0 {
		t.Fatalf("expected empty field set when fill input is incomplete, got %#v", fields)
	}
}

func TestBrowserActFillFieldsPreservesStructuredMissingLocatorWhenTopLevelSignalExists(t *testing.T) {
	_, err := browserActFillFields(map[string]any{
		"text": "a@example.com",
	})
	if err == nil {
		t.Fatalf("expected browserActFillFields to preserve missing locator")
	}
	argErr, ok := AsToolArgumentError(err)
	if !ok {
		t.Fatalf("expected structured argument error, got %T %v", err, err)
	}
	if argErr.Code != "missing_locator" {
		t.Fatalf("unexpected structured argument error: %#v", argErr)
	}
}

func TestRegisterBrowserTools_ActSelectReturnsStructuredMissingValueError(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		capabilities: BrowserCapabilities{
			Select: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"select","selector":"select[name=city]","text":"shanghai"}`,
	})
	if err == nil {
		t.Fatalf("expected browser_act select missing value error")
	}
	argErr, ok := AsToolArgumentError(err)
	if !ok {
		t.Fatalf("expected structured argument error, got %T %v", err, err)
	}
	if argErr.Code != "missing_value" || !argErr.Repairable || !argErr.SafeAutorepair {
		t.Fatalf("unexpected structured argument error: %#v", argErr)
	}
	if argErr.Error() != "browser_act: value or values is required for kind select" {
		t.Fatalf("unexpected error detail: %q", argErr.Error())
	}
	if !reflect.DeepEqual(argErr.MissingFields, []string{"value_or_values"}) {
		t.Fatalf("unexpected missing fields: %#v", argErr.MissingFields)
	}
	if got, want := browserRepairAdviceKinds(argErr.AllowedRepairs), []string{"use_alias_value"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected repair advice: got=%#v want=%#v", got, want)
	}
}

func TestRegisterBrowserTools_ActSelectReturnsStructuredMissingValueErrorForOptionAlias(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		capabilities: BrowserCapabilities{
			Select: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"select","selector":"select[name=city]","option":"shanghai"}`,
	})
	if err == nil {
		t.Fatalf("expected browser_act select missing value error")
	}
	argErr, ok := AsToolArgumentError(err)
	if !ok {
		t.Fatalf("expected structured argument error, got %T %v", err, err)
	}
	if argErr.Code != "missing_value" || !argErr.Repairable || !argErr.SafeAutorepair {
		t.Fatalf("unexpected structured argument error: %#v", argErr)
	}
	if argErr.Error() != "browser_act: value or values is required for kind select" {
		t.Fatalf("unexpected error detail: %q", argErr.Error())
	}
	if !reflect.DeepEqual(argErr.MissingFields, []string{"value_or_values"}) {
		t.Fatalf("unexpected missing fields: %#v", argErr.MissingFields)
	}
	if got, want := browserRepairAdviceKinds(argErr.AllowedRepairs), []string{"use_alias_value"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected repair advice: got=%#v want=%#v", got, want)
	}
}

func TestRegisterBrowserTools_ActSelectReturnsStructuredMissingValueErrorForOptionsAlias(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		capabilities: BrowserCapabilities{
			Select: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"select","selector":"select[name=city]","options":["shanghai","beijing"]}`,
	})
	if err == nil {
		t.Fatalf("expected browser_act select missing value error")
	}
	argErr, ok := AsToolArgumentError(err)
	if !ok {
		t.Fatalf("expected structured argument error, got %T %v", err, err)
	}
	if argErr.Code != "missing_value" || !argErr.Repairable || !argErr.SafeAutorepair {
		t.Fatalf("unexpected structured argument error: %#v", argErr)
	}
	if argErr.Error() != "browser_act: value or values is required for kind select" {
		t.Fatalf("unexpected error detail: %q", argErr.Error())
	}
	if !reflect.DeepEqual(argErr.MissingFields, []string{"value_or_values"}) {
		t.Fatalf("unexpected missing fields: %#v", argErr.MissingFields)
	}
	if got, want := browserRepairAdviceKinds(argErr.AllowedRepairs), []string{"use_alias_value"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected repair advice: got=%#v want=%#v", got, want)
	}
}

func TestRegisterBrowserTools_ActSelectReturnsStructuredMissingValueErrorForContentAlias(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		capabilities: BrowserCapabilities{
			Select: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"select","selector":"select[name=city]","content":"shanghai"}`,
	})
	if err == nil {
		t.Fatalf("expected browser_act select missing value error")
	}
	argErr, ok := AsToolArgumentError(err)
	if !ok {
		t.Fatalf("expected structured argument error, got %T %v", err, err)
	}
	if argErr.Code != "missing_value" || !argErr.Repairable || !argErr.SafeAutorepair {
		t.Fatalf("unexpected structured argument error: %#v", argErr)
	}
	if argErr.Error() != "browser_act: value or values is required for kind select" {
		t.Fatalf("unexpected error detail: %q", argErr.Error())
	}
	if !reflect.DeepEqual(argErr.MissingFields, []string{"value_or_values"}) {
		t.Fatalf("unexpected missing fields: %#v", argErr.MissingFields)
	}
	if got, want := browserRepairAdviceKinds(argErr.AllowedRepairs), []string{"use_alias_value"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected repair advice: got=%#v want=%#v", got, want)
	}
}

func TestRegisterBrowserTools_ActUploadReturnsStructuredMissingUploadPathsErrorForPathAlias(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		capabilities: BrowserCapabilities{
			Upload: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"upload","path":"fixtures/a.txt"}`,
	})
	if err == nil {
		t.Fatalf("expected browser_act upload missing paths error")
	}
	argErr, ok := AsToolArgumentError(err)
	if !ok {
		t.Fatalf("expected structured argument error, got %T %v", err, err)
	}
	if argErr.Code != "missing_upload_paths" || !argErr.Repairable || !argErr.SafeAutorepair {
		t.Fatalf("unexpected structured argument error: %#v", argErr)
	}
	if argErr.Error() != "browser_act: paths/files/file is required for kind upload" {
		t.Fatalf("unexpected error detail: %q", argErr.Error())
	}
	if !reflect.DeepEqual(argErr.MissingFields, []string{"file_or_paths"}) {
		t.Fatalf("unexpected missing fields: %#v", argErr.MissingFields)
	}
	if got, want := browserRepairAdviceKinds(argErr.AllowedRepairs), []string{"use_alias_upload_path"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected repair advice: got=%#v want=%#v", got, want)
	}
}

func TestRegisterBrowserTools_ActPressReturnsStructuredMissingKeyErrorForKeyAlias(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		capabilities: BrowserCapabilities{
			Press: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"press","key_name":"Enter"}`,
	})
	if err == nil {
		t.Fatalf("expected browser_act press missing key error")
	}
	argErr, ok := AsToolArgumentError(err)
	if !ok {
		t.Fatalf("expected structured argument error, got %T %v", err, err)
	}
	if argErr.Code != "missing_key" || !argErr.Repairable || !argErr.SafeAutorepair {
		t.Fatalf("unexpected structured argument error: %#v", argErr)
	}
	if argErr.Error() != "browser_act: key is required for kind press" {
		t.Fatalf("unexpected error detail: %q", argErr.Error())
	}
	if !reflect.DeepEqual(argErr.MissingFields, []string{"key"}) {
		t.Fatalf("unexpected missing fields: %#v", argErr.MissingFields)
	}
	if got, want := browserRepairAdviceKinds(argErr.AllowedRepairs), []string{"use_alias_key"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected repair advice: got=%#v want=%#v", got, want)
	}
}

func TestRegisterBrowserTools_ActEvaluateReturnsStructuredMissingScriptErrorForCodeAlias(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		capabilities: BrowserCapabilities{
			Evaluate: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"evaluate","code":"document.title"}`,
	})
	if err == nil {
		t.Fatalf("expected browser_act evaluate missing script error")
	}
	argErr, ok := AsToolArgumentError(err)
	if !ok {
		t.Fatalf("expected structured argument error, got %T %v", err, err)
	}
	if argErr.Code != "missing_script" || !argErr.Repairable || !argErr.SafeAutorepair {
		t.Fatalf("unexpected structured argument error: %#v", argErr)
	}
	if argErr.Error() != "browser_act: script is required for kind evaluate" {
		t.Fatalf("unexpected error detail: %q", argErr.Error())
	}
	if !reflect.DeepEqual(argErr.MissingFields, []string{"script"}) {
		t.Fatalf("unexpected missing fields: %#v", argErr.MissingFields)
	}
	if got, want := browserRepairAdviceKinds(argErr.AllowedRepairs), []string{"use_alias_script"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected repair advice: got=%#v want=%#v", got, want)
	}
}
