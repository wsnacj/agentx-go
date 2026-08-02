package tools

import (
	"encoding/base64"
	"reflect"
	"testing"
)

func TestBrowserElementRefForSnapshotElementPreservesSelectorIndex(t *testing.T) {
	ref := browserElementRefForSnapshotElement(BrowserSnapshotElement{
		Role:          "button",
		Tag:           "button",
		Label:         "Target",
		Selector:      "button",
		SelectorIndex: 2,
		FramePath:     "main/checkout",
	}, "https://93.184.216.34/frame", "Frame")
	if ref == "" {
		t.Fatalf("expected encoded element ref")
	}
	payload, err := browserDecodeElementRef(ref)
	if err != nil {
		t.Fatalf("decode element ref: %v", err)
	}
	if payload.Selector != "button" || payload.SelectorIndex != 2 || payload.FramePath != "main/checkout" {
		t.Fatalf("expected selector_index to round-trip through element ref, got %#v", payload)
	}
	hint := browserElementHintFromPayload(payload)
	if hint == nil || hint.Selector != "button" || hint.SelectorIndex != 2 || hint.FramePath != "main/checkout" {
		t.Fatalf("expected selector_index to reach browser element hint, got %#v", hint)
	}
}

func TestResolveBrowserElementTargetWithHintPreservesFramePath(t *testing.T) {
	target, err := resolveBrowserElementTargetWithHint("", "", &BrowserElementHint{
		NativeRef: "e12",
		Role:      "button",
		Label:     "Checkout",
		FramePath: "main/cart/line_items",
		PageURL:   "https://93.184.216.34/cart",
		PageTitle: "Cart",
	})
	if err != nil {
		t.Fatalf("resolve with frame-path hint: %v", err)
	}
	if target.Payload.FramePath != "main/cart/line_items" {
		t.Fatalf("expected frame_path in target payload, got %#v", target.Payload)
	}
	hint := browserElementHintForTarget(target)
	if hint == nil || hint.FramePath != "main/cart/line_items" {
		t.Fatalf("expected frame_path in target hint, got %#v", hint)
	}
	resolver := browserElementResolverRequestForTarget(target)
	if resolver == nil || resolver.FramePath != "main/cart/line_items" {
		t.Fatalf("expected frame_path in target resolver, got %#v", resolver)
	}
	if len(resolver.LocatorPlan) == 0 || resolver.LocatorPlan[0].FramePath != "main/cart/line_items" {
		t.Fatalf("expected frame_path in locator plan, got %#v", resolver)
	}
}

func TestBrowserDecodeElementRefSanitizesWrappedRefsAndCorruptedPayloadControls(t *testing.T) {
	rawPayload := "{\"selector\":\"button.signup\",\"label\":\"Sign" + string(rune(8)) + "up\"}"
	ref := "\" " + browserElementMetaRefPrefix + base64.RawURLEncoding.EncodeToString([]byte(rawPayload)) + " \n\""
	payload, err := browserDecodeElementRef(ref)
	if err != nil {
		t.Fatalf("decode sanitized element ref: %v", err)
	}
	if payload.Selector != "button.signup" {
		t.Fatalf("expected selector to survive sanitization, got %#v", payload)
	}
	if payload.Label != "Signup" {
		t.Fatalf("expected control characters to be stripped from label, got %#v", payload)
	}
}

func TestBrowserDecodeElementRefAcceptsPaddedBase64MetaRefs(t *testing.T) {
	rawPayload := "{\"selector\":\"a\",\"native_ref\":\"e20\",\"role\":\"link\",\"tag\":\"a\"}"
	ref := browserElementMetaRefPrefix + base64.StdEncoding.EncodeToString([]byte(rawPayload))
	payload, err := browserDecodeElementRef(ref)
	if err != nil {
		t.Fatalf("decode padded element ref: %v", err)
	}
	if payload.Selector != "a" || payload.NativeRef != "e20" || payload.Role != "link" {
		t.Fatalf("unexpected padded element ref payload: %#v", payload)
	}
}

func TestResolveBrowserElementTargetAcceptsPaddedCSSRefs(t *testing.T) {
	ref := browserElementRefPrefix + base64.StdEncoding.EncodeToString([]byte("button.buy"))
	target, err := resolveBrowserElementTarget("", ref)
	if err != nil {
		t.Fatalf("resolve padded css ref target: %v", err)
	}
	if target.Selector != "button.buy" || target.Ref == "" {
		t.Fatalf("expected padded css ref to resolve selector, got %#v", target)
	}
}

func TestBrowserDecodeElementRefRecoversCorruptedTruncatedPayload(t *testing.T) {
	rawPayload := []byte("{\"selector\":\"button\",\"selector_index\":3,\"native_ref\":\"e21\",\"role\":\"button\",\"Label\":\"Submit\",\"type\":\"button\",\"selector_ynpe\":3,\"page_url\":\"https://93.184.216.34/signup/\",\"page_origin\":\"https://93.184.216.34\",\"page_path\":\"/signup/\",\"page_title\":\"Bad" + string(rune(11)) + "Title")
	ref := browserElementMetaRefPrefix + base64.RawURLEncoding.EncodeToString(rawPayload)
	payload, err := browserDecodeElementRef(ref)
	if err != nil {
		t.Fatalf("decode recovered truncated element ref: %v", err)
	}
	if payload.Selector != "button" || payload.NativeRef != "e21" || payload.Label != "Submit" || payload.SelectorIndex != 3 {
		t.Fatalf("expected partial recovery of actionable locator fields, got %#v", payload)
	}
}

func TestResolveBrowserElementTargetSanitizesWrappedSelectorAndRef(t *testing.T) {
	ref := "\"" + browserElementRefForSelector("button.buy") + "\""
	target, err := resolveBrowserElementTarget(" \n`button.buy`\t", ref)
	if err != nil {
		t.Fatalf("resolve sanitized target: %v", err)
	}
	if target.Selector != "button.buy" || target.Ref == "" {
		t.Fatalf("expected sanitized selector/ref, got %#v", target)
	}
}

func TestResolveBrowserActionElementTargetUsesElementLabelHintWhenSelectorAbsent(t *testing.T) {
	target, err := resolveBrowserActionElementTarget("", "", " \n`2025 年度报告`\t")
	if err != nil {
		t.Fatalf("resolve action target with element label: %v", err)
	}
	if target.Selector != "" || target.Ref == "" || target.Payload.Label != "2025 年度报告" {
		t.Fatalf("expected label-backed action target, got %#v", target)
	}
	payload, err := browserDecodeElementRef(target.Ref)
	if err != nil {
		t.Fatalf("decode synthesized action target ref: %v", err)
	}
	if payload.Label != "2025 年度报告" || payload.Selector != "" {
		t.Fatalf("expected synthesized action target ref to preserve label, got %#v", payload)
	}
}

func TestBrowserClickElementHintForTargetExpandsLabelOnlySemanticLocatorPlan(t *testing.T) {
	target, err := resolveBrowserActionElementTarget("", "", "2025 年度报告")
	if err != nil {
		t.Fatalf("resolve action target: %v", err)
	}
	hint := browserClickElementHintForTarget(target)
	assertBrowserClickLabelSemanticHint(t, hint, "2025 年度报告", "")
	resolver := hint.RemoteResolver()
	if resolver == nil {
		t.Fatalf("expected remote resolver")
	}
	if resolver.ResolutionMode != "locator_plan_only" || resolver.PrimaryKind != "" || resolver.Selector != "" {
		t.Fatalf("label-only click hint must stay semantic until backend resolver, got %#v", resolver)
	}
	for _, candidate := range resolver.MatchPlan {
		if candidate.Kind == "selector" || candidate.Selector != "" {
			t.Fatalf("label-only click hint must not synthesize Playwright-only selector candidates, got %#v", resolver.MatchPlan)
		}
	}
}

func TestResolveBrowserActionElementTargetTreatsContainsSelectorAsHint(t *testing.T) {
	target, err := resolveBrowserActionElementTarget(`a:contains("2025 年度报告")`, "", "")
	if err != nil {
		t.Fatalf("resolve action target with contains selector: %v", err)
	}
	if target.Selector != "" || target.Ref == "" || target.Payload.Label != "2025 年度报告" || target.Payload.Tag != "a" {
		t.Fatalf("expected contains selector to become tag+label hint, got %#v", target)
	}
	payload, err := browserDecodeElementRef(target.Ref)
	if err != nil {
		t.Fatalf("decode contains selector ref: %v", err)
	}
	if payload.Label != "2025 年度报告" || payload.Tag != "a" || payload.Selector != "" {
		t.Fatalf("expected decoded contains selector hint payload, got %#v", payload)
	}
}

func TestResolveBrowserActionElementTargetAllowsContainsSelectorAlongsideRef(t *testing.T) {
	ref := browserElementRefForPayload(browserElementRefPayload{
		NativeRef: "e12",
		Role:      "link",
		Label:     "查看全部",
	})
	target, err := resolveBrowserActionElementTarget(`a:contains("查看全部")`, ref, "")
	if err != nil {
		t.Fatalf("resolve action target with ref plus contains selector: %v", err)
	}
	if target.Ref == "" || target.Payload.Label != "查看全部" {
		t.Fatalf("expected ref-backed target to keep label hint, got %#v", target)
	}
}

func TestBrowserClickElementHintForTagHintUsesSemanticLocatorPlan(t *testing.T) {
	target, err := resolveBrowserActionElementTarget(`a:contains("2025 年度报告")`, "", "")
	if err != nil {
		t.Fatalf("resolve action target with contains selector: %v", err)
	}
	hint := browserClickElementHintForTarget(target)
	assertBrowserClickLabelSemanticHint(t, hint, "2025 年度报告", "a")
}

func TestResolveBrowserActionElementTargetTreatsSelectorLikeElementAsSelector(t *testing.T) {
	target, err := resolveBrowserActionElementTarget("", "", `button:contains("2025 年度报告")`)
	if err != nil {
		t.Fatalf("resolve action target with selector-like element: %v", err)
	}
	if target.Selector != `button:contains("2025 年度报告")` || target.Ref == "" || target.Payload.Label != "" {
		t.Fatalf("expected selector-like element to remain a selector, got %#v", target)
	}
}

func assertBrowserClickLabelSemanticHint(t *testing.T, hint *BrowserElementHint, label string, tag string) {
	t.Helper()
	if hint == nil || hint.Label != label {
		t.Fatalf("expected click hint label %q, got %#v", label, hint)
	}
	if hint.ResolutionMode != "locator_plan_only" {
		t.Fatalf("expected locator_plan_only click hint, got %#v", hint)
	}
	if !reflect.DeepEqual(hint.LocatorOrder, []string{"role_label", "tag_label", "label"}) {
		t.Fatalf("unexpected click hint locator order: %#v", hint)
	}
	switch tag {
	case "":
		if len(hint.LocatorPlan) != 5 ||
			hint.LocatorPlan[0].Kind != "role_label" || hint.LocatorPlan[0].Role != "link" || hint.LocatorPlan[0].Label != label ||
			hint.LocatorPlan[1].Kind != "role_label" || hint.LocatorPlan[1].Role != "button" || hint.LocatorPlan[1].Label != label ||
			hint.LocatorPlan[2].Kind != "tag_label" || hint.LocatorPlan[2].Tag != "a" || hint.LocatorPlan[2].Label != label ||
			hint.LocatorPlan[3].Kind != "tag_label" || hint.LocatorPlan[3].Tag != "button" || hint.LocatorPlan[3].Label != label ||
			hint.LocatorPlan[4].Kind != "label" || hint.LocatorPlan[4].Label != label {
			t.Fatalf("unexpected generic click semantic locator plan: %#v", hint)
		}
	case "a":
		if hint.Tag != "a" ||
			len(hint.LocatorPlan) != 3 ||
			hint.LocatorPlan[0].Kind != "role_label" || hint.LocatorPlan[0].Role != "link" || hint.LocatorPlan[0].Label != label ||
			hint.LocatorPlan[1].Kind != "tag_label" || hint.LocatorPlan[1].Tag != "a" || hint.LocatorPlan[1].Label != label ||
			hint.LocatorPlan[2].Kind != "label" || hint.LocatorPlan[2].Label != label {
			t.Fatalf("unexpected link click semantic locator plan: %#v", hint)
		}
	case "button":
		if hint.Tag != "button" ||
			len(hint.LocatorPlan) != 3 ||
			hint.LocatorPlan[0].Kind != "role_label" || hint.LocatorPlan[0].Role != "button" || hint.LocatorPlan[0].Label != label ||
			hint.LocatorPlan[1].Kind != "tag_label" || hint.LocatorPlan[1].Tag != "button" || hint.LocatorPlan[1].Label != label ||
			hint.LocatorPlan[2].Kind != "label" || hint.LocatorPlan[2].Label != label {
			t.Fatalf("unexpected button click semantic locator plan: %#v", hint)
		}
	default:
		t.Fatalf("unsupported expected click hint tag %q", tag)
	}
	for _, candidate := range hint.LocatorPlan {
		if candidate.Kind == "selector" || candidate.Selector != "" {
			t.Fatalf("semantic click hint must not contain selector candidates, got %#v", hint)
		}
	}
}
