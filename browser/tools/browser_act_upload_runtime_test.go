package tools

import "testing"

func TestBrowserActUploadDefaultStatus(t *testing.T) {
	if got := browserActUploadDefaultStatus(browserActUploadLocator{}); got != "armed" {
		t.Fatalf("browserActUploadDefaultStatus(empty) = %q, want armed", got)
	}
	if got := browserActUploadDefaultStatus(browserActUploadLocator{InputRef: "ref://file"}); got != "uploaded" {
		t.Fatalf("browserActUploadDefaultStatus(input_ref) = %q, want uploaded", got)
	}
}

func TestBrowserActResolveUploadLocatorSelectorlessInputRef(t *testing.T) {
	ref := browserElementRefForPayload(browserElementRefPayload{
		NativeRef: "e12",
		Tag:       "input",
		Type:      "file",
		Label:     "Upload",
		PageURL:   "https://93.184.216.34/form",
		PageTitle: "Upload Form",
	})
	locator, err := browserActResolveUploadLocator(map[string]any{"input_ref": ref})
	if err != nil {
		t.Fatalf("browserActResolveUploadLocator: %v", err)
	}
	if locator.InputRef != ref || locator.Ref != "" || locator.Selector != "" {
		t.Fatalf("unexpected selectorless upload locator: %#v", locator)
	}
	if locator.ElementHint == nil || locator.ElementHint.NativeRef != "e12" || locator.ElementHint.Type != "file" {
		t.Fatalf("unexpected selectorless upload locator hint: %#v", locator.ElementHint)
	}
}
