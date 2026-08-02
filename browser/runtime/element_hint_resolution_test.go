package browserruntime

import (
	"errors"
	"reflect"
	"testing"
)

type testBrowserElementResolverAdapter struct {
	events  []string
	match   string
	pageErr error
}

func (a *testBrowserElementResolverAdapter) ValidatePageBinding(candidate BrowserLocatorCandidate) error {
	a.events = append(a.events, "page:"+candidate.Kind)
	return a.pageErr
}

func (a *testBrowserElementResolverAdapter) ResolveNativeRef(value string) (bool, error) {
	a.events = append(a.events, "native:"+value)
	return a.match == "native_ref", nil
}

func (a *testBrowserElementResolverAdapter) ResolveSelector(value string) (bool, error) {
	a.events = append(a.events, "selector:"+value)
	return a.match == "selector", nil
}

func (a *testBrowserElementResolverAdapter) ResolveSemanticLocator(candidate BrowserLocatorCandidate) (bool, error) {
	a.events = append(a.events, "semantic:"+candidate.Kind)
	return a.match == candidate.Kind, nil
}

func TestBrowserElementHintEffectiveResolutionMode(t *testing.T) {
	hint := &BrowserElementHint{
		NativeRef: "e12",
		Role:      "button",
		Label:     "Buy",
		PageURL:   "https://example.com/cart",
		PageTitle: "Cart",
	}
	if got := hint.EffectiveResolutionMode(); got != "native_ref_first" {
		t.Fatalf("expected native_ref_first, got %q", got)
	}
	if order := hint.EffectiveLocatorOrder(); len(order) != 3 || order[0] != "native_ref" || order[1] != "role_label" || order[2] != "page_binding" {
		t.Fatalf("unexpected locator order: %#v", order)
	}

	hint = &BrowserElementHint{
		Selector:      `button.buy`,
		SelectorIndex: 2,
	}
	if got := hint.EffectiveResolutionMode(); got != "selector_first" {
		t.Fatalf("expected selector_first, got %q", got)
	}
	if plan := hint.EffectiveLocatorPlan(); len(plan) != 1 || plan[0].Kind != "selector" || plan[0].Selector != `button.buy` || plan[0].SelectorIndex != 2 {
		t.Fatalf("unexpected selector-first locator plan: %#v", plan)
	}

	hint = &BrowserElementHint{
		Role:      "button",
		Label:     "Search",
		PageURL:   "https://example.com/search?q=agentx",
		PageTitle: "Example Search",
	}
	if got := hint.EffectiveResolutionMode(); got != "locator_plan_only" {
		t.Fatalf("expected locator_plan_only, got %q", got)
	}
	if plan := hint.EffectiveLocatorPlan(); len(plan) != 2 || plan[0].Kind != "role_label" || plan[1].Kind != "page_binding" {
		t.Fatalf("unexpected locator-plan-only fallback: %#v", plan)
	}
}

func TestBrowserElementHintEffectiveResolutionModePreservesExplicitPlan(t *testing.T) {
	hint := &BrowserElementHint{
		LocatorOrder: []string{"native_ref", "selector"},
		LocatorPlan: []BrowserLocatorCandidate{
			{Kind: "native_ref", NativeRef: "e12"},
			{Kind: "selector", Selector: `button.buy`},
		},
	}
	if got := hint.EffectiveResolutionMode(); got != "native_ref_first" {
		t.Fatalf("expected native_ref_first, got %q", got)
	}
	if plan := hint.EffectiveLocatorPlan(); len(plan) != 2 || plan[0].NativeRef != "e12" || plan[1].Selector != `button.buy` {
		t.Fatalf("unexpected explicit locator plan: %#v", plan)
	}
}

func TestBrowserElementHintRemoteProjection(t *testing.T) {
	hint := &BrowserElementHint{
		NativeRef: "e12",
		Selector:  `button.buy`,
		Role:      "button",
		Label:     "Buy",
		PageURL:   "https://example.com/cart",
		PageTitle: "Cart",
	}
	projection := hint.RemoteProjection()
	if projection.ResolutionMode != "native_ref_first" || projection.PrimaryKind != "native_ref" || projection.ElementRef != "e12" || projection.Selector != `button.buy` {
		t.Fatalf("unexpected native_ref_first projection: %#v", projection)
	}
	if len(projection.FallbackPlan) != 2 || projection.FallbackPlan[0].Kind != "role_label" || projection.FallbackPlan[1].Kind != "page_binding" {
		t.Fatalf("unexpected native_ref_first fallback plan: %#v", projection)
	}

	hint = &BrowserElementHint{
		Selector: `button.buy`,
	}
	projection = hint.RemoteProjection()
	if projection.ResolutionMode != "selector_first" || projection.PrimaryKind != "selector" || projection.ElementRef != "" || projection.Selector != `button.buy` {
		t.Fatalf("unexpected selector_first projection: %#v", projection)
	}
	if len(projection.FallbackPlan) != 0 {
		t.Fatalf("expected selector_first fallback to be empty, got %#v", projection)
	}

	hint = &BrowserElementHint{
		Role:      "button",
		Label:     "Search",
		PageURL:   "https://example.com/search?q=agentx",
		PageTitle: "Example Search",
	}
	projection = hint.RemoteProjection()
	if projection.ResolutionMode != "locator_plan_only" || projection.PrimaryKind != "" || projection.ElementRef != "" || projection.Selector != "" {
		t.Fatalf("unexpected locator_plan_only projection: %#v", projection)
	}
	if len(projection.FallbackPlan) != 2 || projection.FallbackPlan[0].Kind != "role_label" || projection.FallbackPlan[1].Kind != "page_binding" {
		t.Fatalf("unexpected locator_plan_only fallback: %#v", projection)
	}
}

func TestBrowserElementHintFromLocatorPlan(t *testing.T) {
	hint := BrowserElementHintFromLocatorPlan([]BrowserLocatorCandidate{
		{Kind: "selector", Selector: `button.search`, SelectorIndex: 2},
		{Kind: "role_label", Role: "button", Label: "Search"},
		{Kind: "page_binding", PageURL: "https://example.com/search?q=agentx", PageTitle: "Example Search"},
	})
	if hint == nil {
		t.Fatalf("expected hint")
	}
	if hint.NativeRef != "" || hint.Selector != `button.search` || hint.SelectorIndex != 2 || hint.Role != "button" || hint.Label != "Search" || hint.PageURL != "https://example.com/search?q=agentx" || hint.PageTitle != "Example Search" {
		t.Fatalf("unexpected hint from locator plan: %#v", hint)
	}
	if got := hint.ResolutionMode; got != "selector_first" {
		t.Fatalf("expected selector_first, got %q", got)
	}
	if order := hint.LocatorOrder; len(order) != 3 || order[0] != "selector" || order[1] != "role_label" || order[2] != "page_binding" {
		t.Fatalf("unexpected locator order: %#v", order)
	}
}

func TestBrowserElementHintFromLocatorPlanPreservesSemanticSpecificity(t *testing.T) {
	hint := BrowserElementHintFromLocatorPlan([]BrowserLocatorCandidate{
		{Kind: "role_label", Role: "link", Label: "Checkout", Tag: "a", Href: "/checkout", SelectorIndex: 1},
		{Kind: "page_binding", PageURL: "https://example.com/cart"},
	})
	if hint == nil {
		t.Fatalf("expected hint")
	}
	if hint.Role != "link" || hint.Label != "Checkout" || hint.Tag != "a" || hint.Href != "/checkout" || hint.SelectorIndex != 1 {
		t.Fatalf("expected semantic specificity to survive locator plan import, got %#v", hint)
	}
	if plan := hint.EffectiveLocatorPlan(); len(plan) == 0 {
		t.Fatalf("expected semantic selector_index to survive locator plan import/export, got %#v", plan)
	} else {
		var roleLabel *BrowserLocatorCandidate
		for idx := range plan {
			if plan[idx].Kind == "role_label" {
				roleLabel = &plan[idx]
				break
			}
		}
		if roleLabel == nil || roleLabel.SelectorIndex != 1 {
			t.Fatalf("expected role_label candidate to preserve selector_index=1, got %#v", plan)
		}
	}
}

func TestBrowserElementHintEffectiveLocatorPlanPreservesWeakSemanticSpecificity(t *testing.T) {
	hint := &BrowserElementHint{
		LocatorOrder:  []string{"placeholder", "type"},
		Tag:           "input",
		Type:          "email",
		Placeholder:   "Work email",
		SelectorIndex: 2,
		FramePath:     "main/form/frame-1",
	}
	plan := hint.EffectiveLocatorPlan()
	if len(plan) != 2 {
		t.Fatalf("expected placeholder/type locator plan, got %#v", plan)
	}
	if plan[0].Kind != "placeholder" ||
		plan[0].Tag != "input" ||
		plan[0].Type != "email" ||
		plan[0].Placeholder != "Work email" ||
		plan[0].SelectorIndex != 2 ||
		plan[0].FramePath != "main/form/frame-1" {
		t.Fatalf("expected placeholder candidate to preserve tag/type specificity, got %#v", plan[0])
	}
	if plan[1].Kind != "type" ||
		plan[1].Tag != "input" ||
		plan[1].Type != "email" ||
		plan[1].SelectorIndex != 2 ||
		plan[1].FramePath != "main/form/frame-1" {
		t.Fatalf("expected type candidate to preserve tag specificity, got %#v", plan[1])
	}
}

func TestBrowserElementHintEffectiveLocatorPlanRequiresCompleteRoleAndTagLabel(t *testing.T) {
	hint := &BrowserElementHint{
		LocatorOrder: []string{"role_label"},
		Label:        "Checkout",
	}
	if plan := hint.EffectiveLocatorPlan(); len(plan) != 0 {
		t.Fatalf("expected roleless role_label candidate to be omitted, got %#v", plan)
	}

	hint = &BrowserElementHint{
		LocatorOrder: []string{"tag_label"},
		Type:         "email",
		Label:        "Email",
	}
	if plan := hint.EffectiveLocatorPlan(); len(plan) != 0 {
		t.Fatalf("expected tagless tag_label candidate to be omitted, got %#v", plan)
	}

	hint = &BrowserElementHint{
		Label: "Email",
		Type:  "email",
	}
	order := hint.EffectiveLocatorOrder()
	if len(order) != 2 || order[0] != "label" || order[1] != "type" {
		t.Fatalf("expected label+type-only hint to use label/type order, got %#v", order)
	}
	plan := hint.EffectiveLocatorPlan()
	if len(plan) != 2 ||
		plan[0].Kind != "label" ||
		plan[0].Label != "Email" ||
		plan[0].Type != "email" ||
		plan[1].Kind != "type" ||
		plan[1].Type != "email" {
		t.Fatalf("expected label+type-only hint to avoid tag_label, got %#v", plan)
	}

	hint = &BrowserElementHint{
		LocatorOrder: []string{"role_label", "tag_label"},
		Role:         "button",
		Tag:          "button",
		Label:        "Checkout",
	}
	plan = hint.EffectiveLocatorPlan()
	if len(plan) != 2 ||
		plan[0].Kind != "role_label" ||
		plan[0].Role != "button" ||
		plan[0].Label != "Checkout" ||
		plan[1].Kind != "tag_label" ||
		plan[1].Tag != "button" ||
		plan[1].Label != "Checkout" {
		t.Fatalf("expected complete role_label/tag_label candidates to survive, got %#v", plan)
	}
}

func TestBrowserElementHintEffectiveLocatorPlanRequiresCompleteTagType(t *testing.T) {
	hint := &BrowserElementHint{
		LocatorOrder: []string{"tag_type"},
		Tag:          "input",
	}
	if plan := hint.EffectiveLocatorPlan(); len(plan) != 0 {
		t.Fatalf("expected incomplete tag_type candidate to be omitted, got %#v", plan)
	}

	hint = &BrowserElementHint{
		LocatorOrder: []string{"tag_type"},
		Type:         "email",
	}
	if plan := hint.EffectiveLocatorPlan(); len(plan) != 0 {
		t.Fatalf("expected tagless tag_type candidate to be omitted, got %#v", plan)
	}

	hint = &BrowserElementHint{
		LocatorOrder: []string{"tag_type"},
		Tag:          "input",
		Type:         "email",
	}
	plan := hint.EffectiveLocatorPlan()
	if len(plan) != 1 || plan[0].Kind != "tag_type" || plan[0].Tag != "input" || plan[0].Type != "email" {
		t.Fatalf("expected complete tag_type candidate to survive, got %#v", plan)
	}
}

func TestBrowserElementResolverRequestNormalizedKeepsDistinctTypeTags(t *testing.T) {
	resolver := (&BrowserElementResolverRequest{
		ResolutionMode: "locator_plan_only",
		MatchPlan: []BrowserLocatorCandidate{
			{Kind: "type", Tag: "button", Type: "submit"},
			{Kind: "type", Tag: "input", Type: "submit"},
		},
	}).Normalized()
	if resolver == nil {
		t.Fatalf("expected normalized resolver")
	}
	if len(resolver.MatchPlan) != 2 ||
		resolver.MatchPlan[0].Kind != "type" ||
		resolver.MatchPlan[0].Tag != "button" ||
		resolver.MatchPlan[1].Kind != "type" ||
		resolver.MatchPlan[1].Tag != "input" {
		t.Fatalf("expected distinct type candidates to survive tag-aware keying, got %#v", resolver.MatchPlan)
	}
}

func TestBrowserElementResolverRequestNormalizedDropsIncompleteRoleAndTagLabel(t *testing.T) {
	resolver := (&BrowserElementResolverRequest{
		ResolutionMode: "locator_plan_only",
		MatchPlan: []BrowserLocatorCandidate{
			{Kind: "role_label", Label: "Checkout"},
			{Kind: "tag_label", Type: "email", Label: "Email"},
			{Kind: "role_label", Role: "button", Label: "Checkout"},
			{Kind: "tag_label", Tag: "input", Type: "email", Label: "Email"},
		},
	}).Normalized()
	if resolver == nil {
		t.Fatalf("expected normalized resolver")
	}
	if len(resolver.MatchPlan) != 2 ||
		resolver.MatchPlan[0].Kind != "role_label" ||
		resolver.MatchPlan[0].Role != "button" ||
		resolver.MatchPlan[0].Label != "Checkout" ||
		resolver.MatchPlan[1].Kind != "tag_label" ||
		resolver.MatchPlan[1].Tag != "input" ||
		resolver.MatchPlan[1].Type != "email" ||
		resolver.MatchPlan[1].Label != "Email" {
		t.Fatalf("expected only complete role_label/tag_label candidates to survive, got %#v", resolver.MatchPlan)
	}
}

func TestBrowserElementResolverRequestNormalizedDropsIncompleteTagType(t *testing.T) {
	resolver := (&BrowserElementResolverRequest{
		ResolutionMode: "locator_plan_only",
		MatchPlan: []BrowserLocatorCandidate{
			{Kind: "tag_type", Tag: "input"},
			{Kind: "tag_type", Type: "email"},
			{Kind: "tag_type", Tag: "input", Type: "email"},
		},
	}).Normalized()
	if resolver == nil {
		t.Fatalf("expected normalized resolver")
	}
	if len(resolver.MatchPlan) != 1 ||
		resolver.MatchPlan[0].Kind != "tag_type" ||
		resolver.MatchPlan[0].Tag != "input" ||
		resolver.MatchPlan[0].Type != "email" {
		t.Fatalf("expected only complete tag_type candidate to survive, got %#v", resolver.MatchPlan)
	}
}

func TestBrowserElementHintRemoteHint(t *testing.T) {
	hint := &BrowserElementHint{
		NativeRef:     "e12",
		Selector:      `button.buy`,
		SelectorIndex: 1,
		Role:          "button",
		Label:         "Buy",
		PageURL:       "https://example.com/cart",
		PageTitle:     "Cart",
	}
	remoteHint := hint.RemoteHint()
	if remoteHint == nil {
		t.Fatalf("expected remote hint")
	}
	if remoteHint.NativeRef != "" || remoteHint.Selector != "" || remoteHint.Role != "button" || remoteHint.Label != "Buy" || remoteHint.PageURL != "https://example.com/cart" || remoteHint.PageTitle != "Cart" {
		t.Fatalf("unexpected remote hint: %#v", remoteHint)
	}
	if got := remoteHint.ResolutionMode; got != "locator_plan_only" {
		t.Fatalf("expected locator_plan_only remote hint, got %q", got)
	}
	if order := remoteHint.LocatorOrder; len(order) != 2 || order[0] != "role_label" || order[1] != "page_binding" {
		t.Fatalf("unexpected remote locator order: %#v", order)
	}

	selectorOnly := (&BrowserElementHint{Selector: `button.buy`}).RemoteHint()
	if selectorOnly != nil {
		t.Fatalf("expected selector-only remote hint to be empty, got %#v", selectorOnly)
	}
}

func TestBrowserElementHintRemoteResolverPreservesSemanticSpecificity(t *testing.T) {
	hint := &BrowserElementHint{
		NativeRef:     "e12",
		Selector:      `a.checkout`,
		SelectorIndex: 2,
		Role:          "link",
		Tag:           "a",
		Label:         "Checkout",
		Href:          "/checkout",
		PageURL:       "https://example.com/cart",
	}
	resolver := hint.RemoteResolver()
	if resolver == nil {
		t.Fatalf("expected remote resolver")
	}
	var roleLabel *BrowserLocatorCandidate
	for idx := range resolver.MatchPlan {
		candidate := resolver.MatchPlan[idx]
		if candidate.Kind == "role_label" {
			roleLabel = &candidate
			break
		}
	}
	if roleLabel == nil ||
		roleLabel.Role != "link" ||
		roleLabel.Tag != "a" ||
		roleLabel.Href != "/checkout" ||
		roleLabel.SelectorIndex != 2 {
		t.Fatalf("expected remote resolver to preserve role_label specificity, got %#v", resolver.MatchPlan)
	}
}

func TestBrowserElementHintRemoteResolver(t *testing.T) {
	hint := &BrowserElementHint{
		NativeRef:     "e12",
		Selector:      `button.buy`,
		SelectorIndex: 1,
		FramePath:     "main/cart/frame-2",
		Role:          "button",
		Label:         "Buy",
		PageURL:       "https://example.com/cart",
		PageTitle:     "Cart",
	}
	resolver := hint.RemoteResolver()
	if resolver == nil {
		t.Fatalf("expected remote resolver")
	}
	if resolver.ResolutionMode != "native_ref_first" || resolver.PrimaryKind != "native_ref" || resolver.ElementRef != "e12" || resolver.Selector != `button.buy` || resolver.SelectorIndex != 1 || resolver.FramePath != "main/cart/frame-2" {
		t.Fatalf("unexpected remote resolver: %#v", resolver)
	}
	if len(resolver.LocatorOrder) != 4 || resolver.LocatorOrder[0] != "native_ref" || resolver.LocatorOrder[1] != "selector" {
		t.Fatalf("unexpected remote resolver locator_order: %#v", resolver)
	}
	if len(resolver.LocatorPlan) != 4 || resolver.LocatorPlan[0].Kind != "native_ref" || resolver.LocatorPlan[0].FramePath != "main/cart/frame-2" || resolver.LocatorPlan[1].Kind != "selector" || resolver.LocatorPlan[1].SelectorIndex != 1 || resolver.LocatorPlan[1].FramePath != "main/cart/frame-2" {
		t.Fatalf("unexpected remote resolver locator_plan: %#v", resolver)
	}
}

func TestBrowserElementResolverRequestNormalized(t *testing.T) {
	resolver := (&BrowserElementResolverRequest{
		ResolutionMode: "native_ref_first",
		PrimaryKind:    "native_ref",
		ElementRef:     "e12",
		Selector:       `button.buy`,
		FramePath:      "main/cart/frame-2",
		LocatorOrder:   []string{"selector", "native_ref", "role_label", "selector", "page_binding"},
		LocatorPlan: []BrowserLocatorCandidate{
			{Kind: "selector", Selector: `button.buy`, FramePath: "main/cart/frame-2"},
			{Kind: "native_ref", NativeRef: "e12", FramePath: "main/cart/frame-2"},
			{Kind: "selector", Selector: `button.buy`, FramePath: "main/cart/frame-2"},
			{Kind: "role_label", Role: "button", Label: "Buy", FramePath: "main/cart/frame-2"},
			{Kind: "page_binding", PageURL: "https://example.com/cart", PageTitle: "Cart", TabIndex: 2},
		},
	}).Normalized()
	if resolver == nil {
		t.Fatalf("expected normalized resolver")
	}
	if resolver.ResolutionMode != "native_ref_first" || resolver.PrimaryKind != "native_ref" || resolver.ElementRef != "e12" || resolver.Selector != `button.buy` || resolver.FramePath != "main/cart/frame-2" {
		t.Fatalf("unexpected normalized resolver identity: %#v", resolver)
	}
	if len(resolver.LocatorOrder) != 4 || resolver.LocatorOrder[0] != "native_ref" || resolver.LocatorOrder[1] != "selector" || resolver.LocatorOrder[2] != "role_label" || resolver.LocatorOrder[3] != "page_binding" {
		t.Fatalf("unexpected normalized resolver locator_order: %#v", resolver)
	}
	if len(resolver.LocatorPlan) != 4 || resolver.LocatorPlan[0].Kind != "native_ref" || resolver.LocatorPlan[0].FramePath != "main/cart/frame-2" || resolver.LocatorPlan[1].Kind != "selector" || resolver.LocatorPlan[1].FramePath != "main/cart/frame-2" || resolver.LocatorPlan[2].Kind != "role_label" || resolver.LocatorPlan[2].FramePath != "main/cart/frame-2" || resolver.LocatorPlan[3].Kind != "page_binding" {
		t.Fatalf("unexpected normalized resolver locator_plan: %#v", resolver)
	}
	if len(resolver.MatchPlan) != 3 || resolver.MatchPlan[0].Kind != "native_ref" || resolver.MatchPlan[1].Kind != "selector" || resolver.MatchPlan[2].Kind != "role_label" {
		t.Fatalf("unexpected normalized resolver match_plan: %#v", resolver)
	}
	if resolver.PageBinding == nil || resolver.PageBinding.Kind != "page_binding" || resolver.PageBinding.PageURL != "https://example.com/cart" || resolver.PageBinding.PageTitle != "Cart" || resolver.PageBinding.TabIndex != 2 {
		t.Fatalf("unexpected normalized resolver page_binding: %#v", resolver)
	}
}

func TestBrowserElementResolverRequestNormalizedPreservesSelectorIndex(t *testing.T) {
	resolver := (&BrowserElementResolverRequest{
		ResolutionMode: "native_ref_first",
		PrimaryKind:    "native_ref",
		ElementRef:     "e12",
		Selector:       `button.buy`,
		SelectorIndex:  2,
		LocatorPlan: []BrowserLocatorCandidate{
			{Kind: "native_ref", NativeRef: "e12"},
			{Kind: "selector", Selector: `button.buy`, SelectorIndex: 2},
			{Kind: "selector", Selector: `button.buy`, SelectorIndex: 1},
		},
	}).Normalized()
	if resolver == nil {
		t.Fatalf("expected normalized resolver")
	}
	if resolver.SelectorIndex != 2 {
		t.Fatalf("expected normalized selector_index=2, got %#v", resolver)
	}
	if len(resolver.MatchPlan) != 3 {
		t.Fatalf("expected selector candidates with distinct selector_index to be preserved, got %#v", resolver.MatchPlan)
	}
	if resolver.MatchPlan[1].Kind != "selector" || resolver.MatchPlan[1].SelectorIndex != 2 {
		t.Fatalf("expected first selector fallback to preserve selector_index=2, got %#v", resolver.MatchPlan)
	}
	if resolver.MatchPlan[2].Kind != "selector" || resolver.MatchPlan[2].SelectorIndex != 1 {
		t.Fatalf("expected second selector fallback to preserve selector_index=1, got %#v", resolver.MatchPlan)
	}
}

func TestBrowserElementResolverRequestNormalizedPreservesSemanticSpecificity(t *testing.T) {
	resolver := (&BrowserElementResolverRequest{
		ResolutionMode: "locator_plan_only",
		LocatorPlan: []BrowserLocatorCandidate{
			{Kind: "role_label", Role: "link", Label: "Checkout", Href: "#one"},
			{Kind: "role_label", Role: "link", Label: "Checkout", Href: "#two"},
			{Kind: "tag_label", Tag: "input", Type: "email", Label: "Email", Placeholder: "Personal email"},
			{Kind: "tag_label", Tag: "input", Type: "email", Label: "Email", Placeholder: "Work email"},
			{Kind: "label", Label: "Name", Placeholder: "Personal name"},
			{Kind: "label", Label: "Name", Placeholder: "Work name"},
		},
	}).Normalized()
	if resolver == nil {
		t.Fatalf("expected normalized resolver")
	}
	if len(resolver.MatchPlan) != 6 {
		t.Fatalf("expected semantic candidates with distinct specificity to be preserved, got %#v", resolver.MatchPlan)
	}
	if resolver.MatchPlan[0].Kind != "role_label" || resolver.MatchPlan[0].Href != "#one" {
		t.Fatalf("expected first role_label candidate to preserve href specificity, got %#v", resolver.MatchPlan[0])
	}
	if resolver.MatchPlan[1].Kind != "role_label" || resolver.MatchPlan[1].Href != "#two" {
		t.Fatalf("expected second role_label candidate to preserve href specificity, got %#v", resolver.MatchPlan[1])
	}
	if resolver.MatchPlan[2].Kind != "tag_label" || resolver.MatchPlan[2].Placeholder != "Personal email" {
		t.Fatalf("expected first tag_label candidate to preserve placeholder specificity, got %#v", resolver.MatchPlan[2])
	}
	if resolver.MatchPlan[3].Kind != "tag_label" || resolver.MatchPlan[3].Placeholder != "Work email" {
		t.Fatalf("expected second tag_label candidate to preserve placeholder specificity, got %#v", resolver.MatchPlan[3])
	}
	if resolver.MatchPlan[4].Kind != "label" || resolver.MatchPlan[4].Placeholder != "Personal name" {
		t.Fatalf("expected first label candidate to preserve placeholder specificity, got %#v", resolver.MatchPlan[4])
	}
	if resolver.MatchPlan[5].Kind != "label" || resolver.MatchPlan[5].Placeholder != "Work name" {
		t.Fatalf("expected second label candidate to preserve placeholder specificity, got %#v", resolver.MatchPlan[5])
	}
}

func TestBrowserElementResolverRequestNormalizedPreservesSemanticSelectorIndex(t *testing.T) {
	resolver := (&BrowserElementResolverRequest{
		ResolutionMode: "locator_plan_only",
		LocatorPlan: []BrowserLocatorCandidate{
			{Kind: "role_label", Role: "button", Label: "Checkout", SelectorIndex: 1},
			{Kind: "role_label", Role: "button", Label: "Checkout", SelectorIndex: 2},
			{Kind: "tag_label", Tag: "input", Type: "email", Label: "Email", SelectorIndex: 1},
			{Kind: "tag_label", Tag: "input", Type: "email", Label: "Email", SelectorIndex: 2},
		},
	}).Normalized()
	if resolver == nil {
		t.Fatalf("expected normalized resolver")
	}
	if len(resolver.MatchPlan) != 4 {
		t.Fatalf("expected semantic candidates with distinct selector_index to be preserved, got %#v", resolver.MatchPlan)
	}
	if resolver.MatchPlan[0].Kind != "role_label" || resolver.MatchPlan[0].SelectorIndex != 1 {
		t.Fatalf("expected first role_label candidate to preserve selector_index=1, got %#v", resolver.MatchPlan[0])
	}
	if resolver.MatchPlan[1].Kind != "role_label" || resolver.MatchPlan[1].SelectorIndex != 2 {
		t.Fatalf("expected second role_label candidate to preserve selector_index=2, got %#v", resolver.MatchPlan[1])
	}
	if resolver.MatchPlan[2].Kind != "tag_label" || resolver.MatchPlan[2].SelectorIndex != 1 {
		t.Fatalf("expected first tag_label candidate to preserve selector_index=1, got %#v", resolver.MatchPlan[2])
	}
	if resolver.MatchPlan[3].Kind != "tag_label" || resolver.MatchPlan[3].SelectorIndex != 2 {
		t.Fatalf("expected second tag_label candidate to preserve selector_index=2, got %#v", resolver.MatchPlan[3])
	}
}

func TestBrowserElementResolverRequestEffectiveResolutionPlan(t *testing.T) {
	plan := (&BrowserElementResolverRequest{
		ResolutionMode: "native_ref_first",
		PrimaryKind:    "native_ref",
		ElementRef:     "e12",
		Selector:       `button.buy`,
		LocatorOrder:   []string{"native_ref", "selector", "role_label", "page_binding"},
		LocatorPlan: []BrowserLocatorCandidate{
			{Kind: "native_ref", NativeRef: "e12"},
			{Kind: "selector", Selector: `button.buy`},
			{Kind: "role_label", Role: "button", Label: "Buy"},
			{Kind: "page_binding", PageURL: "https://example.com/cart", PageTitle: "Cart", TabIndex: 2},
		},
	}).EffectiveResolutionPlan()
	if plan.ResolutionMode != "native_ref_first" || plan.PrimaryKind != "native_ref" || plan.ElementRef != "e12" || plan.Selector != `button.buy` {
		t.Fatalf("unexpected effective resolution plan identity: %#v", plan)
	}
	if len(plan.MatchPlan) != 3 || plan.MatchPlan[0].Kind != "native_ref" || plan.MatchPlan[1].Kind != "selector" || plan.MatchPlan[2].Kind != "role_label" {
		t.Fatalf("unexpected effective resolution plan match plan: %#v", plan)
	}
	if plan.PageBinding == nil || plan.PageBinding.Kind != "page_binding" || plan.PageBinding.PageURL != "https://example.com/cart" || plan.PageBinding.PageTitle != "Cart" || plan.PageBinding.TabIndex != 2 {
		t.Fatalf("unexpected effective resolution plan page binding: %#v", plan)
	}
}

func TestBrowserElementResolverRequestFromRemote(t *testing.T) {
	resolver := BrowserElementResolverRequestFromRemote(
		"e12",
		`button.buy`,
		&BrowserElementHint{
			Role:      "button",
			Label:     "Buy",
			PageURL:   "https://example.com/cart",
			PageTitle: "Cart",
		},
		nil,
	)
	if resolver == nil {
		t.Fatalf("expected synthesized resolver")
	}
	if resolver.ResolutionMode != "native_ref_first" || resolver.PrimaryKind != "native_ref" || resolver.ElementRef != "e12" || resolver.Selector != `button.buy` {
		t.Fatalf("unexpected synthesized resolver: %#v", resolver)
	}
	if len(resolver.LocatorPlan) != 4 || resolver.LocatorPlan[0].Kind != "native_ref" || resolver.LocatorPlan[1].Kind != "selector" || resolver.LocatorPlan[2].Kind != "role_label" || resolver.LocatorPlan[3].Kind != "page_binding" {
		t.Fatalf("unexpected synthesized locator_plan: %#v", resolver)
	}

	normalized := BrowserElementResolverRequestFromRemote(
		"e12",
		`button.buy`,
		&BrowserElementHint{
			Role:      "button",
			Label:     "Buy",
			PageURL:   "https://example.com/cart",
			PageTitle: "Cart",
		},
		&BrowserElementResolverRequest{
			ResolutionMode: "native_ref_first",
			PrimaryKind:    "native_ref",
			ElementRef:     "e12",
			Selector:       `button.buy`,
			LocatorOrder:   []string{"native_ref", "selector", "role_label", "page_binding"},
			LocatorPlan: []BrowserLocatorCandidate{
				{Kind: "native_ref", NativeRef: "e12"},
				{Kind: "selector", Selector: `button.buy`},
				{Kind: "role_label", Role: "button", Label: "Buy"},
				{Kind: "page_binding", PageURL: "https://example.com/cart", PageTitle: "Cart"},
			},
		},
	)
	if normalized == nil || normalized.PrimaryKind != "native_ref" || normalized.ElementRef != "e12" {
		t.Fatalf("unexpected normalized resolver: %#v", normalized)
	}
	if len(normalized.LocatorPlan) != 4 {
		t.Fatalf("unexpected normalized locator_plan: %#v", normalized)
	}
	if len(normalized.MatchPlan) != 3 || normalized.PageBinding == nil || normalized.PageBinding.Kind != "page_binding" {
		t.Fatalf("expected normalized resolver decomposition, got %#v", normalized)
	}

	normalized = BrowserElementResolverRequestFromRemote(
		"e12",
		`button.buy`,
		&BrowserElementHint{
			Role:      "button",
			Label:     "Buy",
			PageURL:   "https://example.com/cart",
			PageTitle: "Cart",
		},
		&BrowserElementResolverRequest{
			ResolutionMode: "native_ref_first",
			PrimaryKind:    "native_ref",
			ElementRef:     "e12",
			Selector:       `button.buy`,
			LocatorOrder:   []string{"selector", "native_ref", "role_label", "page_binding"},
			LocatorPlan: []BrowserLocatorCandidate{
				{Kind: "selector", Selector: `button.buy`},
				{Kind: "native_ref", NativeRef: "e12"},
				{Kind: "selector", Selector: `button.buy`},
				{Kind: "role_label", Role: "button", Label: "Buy"},
				{Kind: "page_binding", PageURL: "https://example.com/cart", PageTitle: "Cart"},
			},
		},
	)
	if normalized == nil || len(normalized.LocatorOrder) != 4 || normalized.LocatorOrder[0] != "native_ref" || len(normalized.LocatorPlan) != 4 {
		t.Fatalf("expected remote normalization to preserve primary-first deduped resolver, got %#v", normalized)
	}
	if len(normalized.MatchPlan) != 3 || normalized.PageBinding == nil || normalized.PageBinding.Kind != "page_binding" {
		t.Fatalf("expected remote normalization to preserve match/page decomposition, got %#v", normalized)
	}

	normalized = BrowserElementResolverRequestFromRemote(
		"",
		"",
		nil,
		&BrowserElementResolverRequest{
			ResolutionMode: "locator_plan_only",
			PrimaryKind:    "selector",
			MatchPlan: []BrowserLocatorCandidate{
				{Kind: "selector", Selector: "#missing-name"},
				{Kind: "label", Label: "Name"},
			},
			PageBinding: &BrowserLocatorCandidate{
				Kind:      "page_binding",
				PageURL:   "https://example.com/form",
				PageTitle: "Form",
			},
		},
	)
	if normalized == nil || normalized.ResolutionMode != "locator_plan_only" || normalized.PrimaryKind != "selector" {
		t.Fatalf("expected match-plan only remote resolver to normalize, got %#v", normalized)
	}
	if len(normalized.LocatorPlan) != 3 || normalized.LocatorPlan[0].Kind != "selector" || normalized.LocatorPlan[1].Kind != "label" || normalized.LocatorPlan[2].Kind != "page_binding" {
		t.Fatalf("unexpected match-plan only locator_plan: %#v", normalized)
	}
	if len(normalized.MatchPlan) != 2 || normalized.PageBinding == nil || normalized.PageBinding.Kind != "page_binding" {
		t.Fatalf("expected match-plan only decomposition to survive normalization, got %#v", normalized)
	}

	normalized = BrowserElementResolverRequestFromRemote(
		"",
		"",
		nil,
		&BrowserElementResolverRequest{
			ResolutionMode: "locator_plan_only",
			PrimaryKind:    "role_label",
			MatchPlan: []BrowserLocatorCandidate{
				{Kind: "role_label", Role: "button", Label: "Checkout"},
			},
			PageBinding: &BrowserLocatorCandidate{
				Kind:      "page_binding",
				FramePath: "main/checkout/frame-2",
			},
		},
	)
	if normalized == nil ||
		normalized.ResolutionMode != "locator_plan_only" ||
		len(normalized.LocatorPlan) != 2 ||
		len(normalized.MatchPlan) != 1 ||
		normalized.PageBinding == nil ||
		normalized.PageBinding.Kind != "page_binding" ||
		normalized.PageBinding.FramePath != "main/checkout/frame-2" ||
		normalized.LocatorPlan[1].Kind != "page_binding" ||
		normalized.LocatorPlan[1].FramePath != "main/checkout/frame-2" {
		t.Fatalf("expected frame-path page binding to survive normalization, got %#v", normalized)
	}
}

func TestBrowserElementResolverRequestResolveWithFallsBackInOrder(t *testing.T) {
	resolver := &BrowserElementResolverRequest{
		ResolutionMode: "native_ref_first",
		PrimaryKind:    "native_ref",
		ElementRef:     "e12",
		Selector:       `button.buy`,
		LocatorPlan: []BrowserLocatorCandidate{
			{Kind: "native_ref", NativeRef: "e12"},
			{Kind: "selector", Selector: `button.buy`},
			{Kind: "role_label", Role: "button", Label: "Buy"},
			{Kind: "page_binding", PageURL: "https://example.com/cart", PageTitle: "Cart"},
		},
	}
	var attempts []BrowserElementResolutionAttempt
	var pageChecks []BrowserLocatorCandidate
	result, err := resolver.ResolveWith(BrowserElementResolverCallbacks{
		ValidatePageBinding: func(candidate BrowserLocatorCandidate) error {
			pageChecks = append(pageChecks, candidate)
			return nil
		},
		ResolveAttempt: func(attempt BrowserElementResolutionAttempt) (bool, error) {
			attempts = append(attempts, attempt)
			return attempt.Candidate.Kind == "selector", nil
		},
	})
	if err != nil {
		t.Fatalf("ResolveWith returned error: %v", err)
	}
	if len(pageChecks) != 1 || pageChecks[0].Kind != "page_binding" {
		t.Fatalf("expected page binding to be checked first, got %#v", pageChecks)
	}
	if len(attempts) != 2 {
		t.Fatalf("expected fallback after native_ref miss, got %#v", attempts)
	}
	if !attempts[0].IsPrimary || attempts[0].Candidate.Kind != "native_ref" || attempts[1].IsPrimary || attempts[1].Candidate.Kind != "selector" {
		t.Fatalf("unexpected ordered attempts: %#v", attempts)
	}
	if result.AttemptCount != 2 || result.MatchedAttempt == nil || result.MatchedAttempt.Candidate.Kind != "selector" {
		t.Fatalf("unexpected resolution result: %#v", result)
	}
	if result.FallbackFromAttempt == nil || result.FallbackFromAttempt.Index != 0 || result.FallbackFromAttempt.Candidate.Kind != "native_ref" || !result.FallbackFromAttempt.IsPrimary {
		t.Fatalf("expected fallback-from attempt to preserve primary native_ref miss, got %#v", result)
	}
	if result.PageBinding == nil || result.PageBinding.Kind != "page_binding" {
		t.Fatalf("expected page binding in result, got %#v", result)
	}
}

func TestBrowserElementResolverRequestResolveWithStopsOnPageBindingError(t *testing.T) {
	resolver := &BrowserElementResolverRequest{
		ResolutionMode: "locator_plan_only",
		LocatorPlan: []BrowserLocatorCandidate{
			{Kind: "role_label", Role: "button", Label: "Checkout"},
			{Kind: "page_binding", PageURL: "https://example.com/cart", PageTitle: "Cart"},
		},
	}
	pageErr := errors.New("page_changed")
	called := false
	result, err := resolver.ResolveWith(BrowserElementResolverCallbacks{
		ValidatePageBinding: func(candidate BrowserLocatorCandidate) error {
			return pageErr
		},
		ResolveAttempt: func(attempt BrowserElementResolutionAttempt) (bool, error) {
			called = true
			return false, nil
		},
	})
	if !errors.Is(err, pageErr) {
		t.Fatalf("expected page binding error, got %v", err)
	}
	if called {
		t.Fatalf("expected page binding failure to stop match attempts")
	}
	if result.AttemptCount != 0 || result.MatchedAttempt != nil {
		t.Fatalf("unexpected result after page binding failure: %#v", result)
	}
}

func TestBrowserElementResolverRequestResolveWithReturnsUnresolvedAttempts(t *testing.T) {
	resolver := &BrowserElementResolverRequest{
		ResolutionMode: "locator_plan_only",
		LocatorPlan: []BrowserLocatorCandidate{
			{Kind: "role_label", Role: "button", Label: "Search"},
			{Kind: "href", Href: "https://example.com/search?q=agentx"},
		},
	}
	var kinds []string
	result, err := resolver.ResolveWith(BrowserElementResolverCallbacks{
		ResolveAttempt: func(attempt BrowserElementResolutionAttempt) (bool, error) {
			kinds = append(kinds, attempt.Candidate.Kind)
			return false, nil
		},
	})
	if err != nil {
		t.Fatalf("ResolveWith returned error: %v", err)
	}
	if !reflect.DeepEqual(kinds, []string{"role_label", "href"}) {
		t.Fatalf("unexpected unresolved attempt order: %#v", kinds)
	}
	if result.AttemptCount != 2 || result.MatchedAttempt != nil {
		t.Fatalf("unexpected unresolved result: %#v", result)
	}
}

func TestBrowserElementResolverRequestResolveWithPreservesFallbackOutcome(t *testing.T) {
	resolver := &BrowserElementResolverRequest{
		ResolutionMode: "locator_plan_only",
		LocatorPlan: []BrowserLocatorCandidate{
			{Kind: "label", Label: "Search", Tag: "input", Type: "text"},
			{Kind: "placeholder", Placeholder: "Search docs", Tag: "input", Type: "text"},
		},
	}
	result, err := resolver.ResolveWith(BrowserElementResolverCallbacks{
		ResolveAttemptDetailed: func(attempt BrowserElementResolutionAttempt) (BrowserElementResolutionAttemptResult, error) {
			switch attempt.Candidate.Kind {
			case "label":
				return BrowserElementResolutionAttemptResult{
					Outcome: &BrowserElementResolverOutcome{
						Status:            "unresolved",
						BlockedBy:         "multiple_candidates_filtered",
						AmbiguityClass:    "filtered_residual",
						CandidateKind:     "label",
						CandidateStrength: "medium",
						ManualRetryHint:   "add_ordinal",
						SpecificityFields: []string{"tag", "type"},
					},
				}, nil
			case "placeholder":
				return BrowserElementResolutionAttemptResult{Matched: true}, nil
			default:
				return BrowserElementResolutionAttemptResult{}, nil
			}
		},
	})
	if err != nil {
		t.Fatalf("ResolveWith returned error: %v", err)
	}
	if result.AttemptCount != 2 || result.MatchedAttempt == nil || result.MatchedAttempt.Candidate.Kind != "placeholder" {
		t.Fatalf("unexpected detailed resolution result: %#v", result)
	}
	if result.FallbackFromAttempt == nil || result.FallbackFromAttempt.Candidate.Kind != "label" {
		t.Fatalf("expected fallback-from attempt to preserve label miss, got %#v", result)
	}
	if result.FallbackFromOutcome == nil ||
		result.FallbackFromOutcome.BlockedBy != "multiple_candidates_filtered" ||
		result.FallbackFromOutcome.AmbiguityClass != "filtered_residual" ||
		result.FallbackFromOutcome.CandidateKind != "label" ||
		result.FallbackFromOutcome.CandidateStrength != "medium" ||
		result.FallbackFromOutcome.ManualRetryHint != "add_ordinal" ||
		!reflect.DeepEqual(result.FallbackFromOutcome.SpecificityFields, []string{"tag", "type"}) {
		t.Fatalf("expected detailed fallback outcome to survive, got %#v", result.FallbackFromOutcome)
	}
}

func TestBrowserElementResolverOutcomeFromResult(t *testing.T) {
	matched := BrowserElementResolverOutcomeFromResult(BrowserElementResolutionResult{
		ResolutionMode: "native_ref_first",
		PrimaryKind:    "native_ref",
		AttemptCount:   3,
		MatchedAttempt: &BrowserElementResolutionAttempt{
			Index: 2,
			Candidate: BrowserLocatorCandidate{
				Kind:          "role_label",
				FramePath:     "main/cart/frame-2",
				SelectorIndex: 2,
			},
		},
		FallbackFromAttempt: &BrowserElementResolutionAttempt{
			Index: 1,
			Candidate: BrowserLocatorCandidate{
				Kind:        "label",
				Label:       "Email",
				Tag:         "input",
				Type:        "email",
				Placeholder: "Work email",
			},
		},
		FallbackFromOutcome: &BrowserElementResolverOutcome{
			Status:            "unresolved",
			BlockedBy:         "multiple_candidates_filtered",
			AmbiguityClass:    "filtered_residual",
			CandidateKind:     "label",
			CandidateStrength: "medium",
			ManualRetryHint:   "add_ordinal",
			SpecificityFields: []string{"tag", "type"},
		},
	}, nil)
	if matched == nil ||
		matched.Status != "matched" ||
		matched.ResolutionMode != "native_ref_first" ||
		matched.PrimaryKind != "native_ref" ||
		matched.AttemptCount != 3 ||
		matched.MatchedKind != "role_label" ||
		matched.MatchedIndex != 2 ||
		matched.MatchedCandidateKind != "role_label" ||
		matched.ResolvedFramePath != "main/cart/frame-2" ||
		matched.ResolvedSelectorIndex != 2 ||
		matched.FallbackFromKind != "label" ||
		matched.FallbackFromIndex != 1 ||
		matched.FallbackFromBlockedBy != "multiple_candidates_filtered" ||
		matched.FallbackFromAmbiguityClass != "filtered_residual" ||
		matched.FallbackFromCandidateStrength != "medium" ||
		matched.FallbackFromManualRetryHint != "add_ordinal" ||
		!reflect.DeepEqual(matched.FallbackFromSpecificityFields, []string{"tag", "type"}) ||
		matched.BlockedBy != "" ||
		matched.RecoveryAction != "" ||
		matched.Note != "resolved via role_label" {
		t.Fatalf("unexpected matched resolver outcome: %#v", matched)
	}

	pageErr := errors.New("page_changed")
	blocked := BrowserElementResolverOutcomeFromResult(BrowserElementResolutionResult{
		ResolutionMode: "native_ref_first",
		PrimaryKind:    "native_ref",
		PageBinding: &BrowserLocatorCandidate{
			Kind:    "page_binding",
			PageURL: "https://example.com/cart",
		},
	}, pageErr)
	if blocked == nil ||
		blocked.Status != "page_binding_blocked" ||
		blocked.BlockedBy != "page_url" ||
		blocked.AttemptCount != 0 ||
		blocked.RecoveryAction != "browser action=snapshot" ||
		blocked.ManualRetryHint != "refresh_snapshot" ||
		blocked.NextStepAlias != "snapshot" ||
		blocked.Note != pageErr.Error() {
		t.Fatalf("unexpected page-binding resolver outcome: %#v", blocked)
	}

	frameBlocked := BrowserElementResolverOutcomeFromResult(BrowserElementResolutionResult{
		ResolutionMode: "locator_plan_only",
		PrimaryKind:    "role_label",
		PageBinding: &BrowserLocatorCandidate{
			Kind:      "page_binding",
			FramePath: "main/cart/frame-2",
		},
	}, errors.New("frame detached"))
	if frameBlocked == nil ||
		frameBlocked.Status != "page_binding_blocked" ||
		frameBlocked.BlockedBy != "frame_path" ||
		frameBlocked.AttemptCount != 0 ||
		frameBlocked.RecoveryAction != "browser action=snapshot" ||
		frameBlocked.ManualRetryHint != "refresh_snapshot" ||
		frameBlocked.NextStepAlias != "snapshot" ||
		frameBlocked.Note != "frame detached" {
		t.Fatalf("unexpected frame page-binding resolver outcome: %#v", frameBlocked)
	}

	unresolved := BrowserElementResolverOutcomeFromResult(BrowserElementResolutionResult{
		ResolutionMode: "locator_plan_only",
		AttemptCount:   2,
	}, nil)
	if unresolved == nil ||
		unresolved.Status != "unresolved" ||
		unresolved.AttemptCount != 2 ||
		unresolved.RecoveryAction != "browser action=snapshot" ||
		unresolved.ManualRetryHint != "refresh_snapshot" ||
		unresolved.NextStepAlias != "snapshot" ||
		unresolved.Note != "no element matched resolver plan" {
		t.Fatalf("unexpected unresolved resolver outcome: %#v", unresolved)
	}

	failed := BrowserElementResolverOutcomeFromResult(BrowserElementResolutionResult{
		ResolutionMode: "native_ref_first",
		PrimaryKind:    "native_ref",
		AttemptCount:   1,
	}, errors.New("cdp detached target"))
	if failed == nil ||
		failed.Status != "resolution_failed" ||
		failed.AttemptCount != 1 ||
		failed.RecoveryAction != "browser action=refresh" ||
		failed.Note != "cdp detached target" {
		t.Fatalf("unexpected failed resolver outcome: %#v", failed)
	}
}

func TestResolveBrowserElementFromRemoteUsesSynthesizedResolver(t *testing.T) {
	var attempts []string
	result, err := ResolveBrowserElementFromRemote(
		"e12",
		`button.buy`,
		&BrowserElementHint{
			Role:      "button",
			Label:     "Buy",
			PageURL:   "https://example.com/cart",
			PageTitle: "Cart",
		},
		nil,
		BrowserElementResolverCallbacks{
			ResolveAttempt: func(attempt BrowserElementResolutionAttempt) (bool, error) {
				attempts = append(attempts, attempt.Candidate.Kind)
				return attempt.Candidate.Kind == "role_label", nil
			},
		},
	)
	if err != nil {
		t.Fatalf("ResolveBrowserElementFromRemote returned error: %v", err)
	}
	if !reflect.DeepEqual(attempts, []string{"native_ref", "selector", "role_label"}) {
		t.Fatalf("unexpected synthesized remote attempt order: %#v", attempts)
	}
	if result.MatchedAttempt == nil || result.MatchedAttempt.Candidate.Kind != "role_label" {
		t.Fatalf("unexpected synthesized remote result: %#v", result)
	}
}

func TestResolveBrowserElementFromRemotePreservesExplicitResolver(t *testing.T) {
	var attempts []string
	result, err := ResolveBrowserElementFromRemote(
		"e12",
		`button.buy`,
		&BrowserElementHint{
			Role:      "button",
			Label:     "Buy",
			PageURL:   "https://example.com/cart",
			PageTitle: "Cart",
		},
		&BrowserElementResolverRequest{
			ResolutionMode: "selector_first",
			PrimaryKind:    "selector",
			Selector:       `button.buy`,
			LocatorPlan: []BrowserLocatorCandidate{
				{Kind: "selector", Selector: `button.buy`},
				{Kind: "role_label", Role: "button", Label: "Buy"},
				{Kind: "page_binding", PageURL: "https://example.com/cart", PageTitle: "Cart"},
			},
		},
		BrowserElementResolverCallbacks{
			ResolveAttempt: func(attempt BrowserElementResolutionAttempt) (bool, error) {
				attempts = append(attempts, attempt.Candidate.Kind)
				return attempt.Candidate.Kind == "selector", nil
			},
		},
	)
	if err != nil {
		t.Fatalf("ResolveBrowserElementFromRemote returned error: %v", err)
	}
	if !reflect.DeepEqual(attempts, []string{"selector"}) {
		t.Fatalf("unexpected explicit remote attempt order: %#v", attempts)
	}
	if result.MatchedAttempt == nil || result.MatchedAttempt.Candidate.Kind != "selector" || !result.MatchedAttempt.IsPrimary {
		t.Fatalf("unexpected explicit remote result: %#v", result)
	}
}

func TestResolveBrowserElementFromRemoteWithAdapterRoutesManagedLookupKinds(t *testing.T) {
	adapter := &testBrowserElementResolverAdapter{match: "role_label"}
	result, err := ResolveBrowserElementFromRemoteWithAdapter(
		"e12",
		`button.buy`,
		&BrowserElementHint{
			Role:      "button",
			Label:     "Buy",
			PageURL:   "https://example.com/cart",
			PageTitle: "Cart",
		},
		nil,
		adapter,
	)
	if err != nil {
		t.Fatalf("ResolveBrowserElementFromRemoteWithAdapter returned error: %v", err)
	}
	if !reflect.DeepEqual(adapter.events, []string{
		"page:page_binding",
		"native:e12",
		"selector:button.buy",
		"semantic:role_label",
	}) {
		t.Fatalf("unexpected adapter call order: %#v", adapter.events)
	}
	if result.MatchedAttempt == nil || result.MatchedAttempt.Candidate.Kind != "role_label" {
		t.Fatalf("unexpected adapter resolution result: %#v", result)
	}
}

func TestResolveBrowserElementFromRemoteWithAdapterStopsOnPageBindingError(t *testing.T) {
	adapter := &testBrowserElementResolverAdapter{pageErr: errors.New("page_changed")}
	_, err := ResolveBrowserElementFromRemoteWithAdapter(
		"",
		"",
		&BrowserElementHint{
			Role:      "button",
			Label:     "Buy",
			PageURL:   "https://example.com/cart",
			PageTitle: "Cart",
		},
		nil,
		adapter,
	)
	if !errors.Is(err, adapter.pageErr) {
		t.Fatalf("expected page binding error, got %v", err)
	}
	if !reflect.DeepEqual(adapter.events, []string{"page:page_binding"}) {
		t.Fatalf("expected adapter to stop after page binding failure, got %#v", adapter.events)
	}
}

func TestResolveBrowserElementFromRemotePayloadUsesGenericMaps(t *testing.T) {
	adapter := &testBrowserElementResolverAdapter{match: "role_label"}
	result, err := ResolveBrowserElementFromRemotePayload(
		"e12",
		`button.buy`,
		map[string]any{
			"role":       "button",
			"label":      "Buy",
			"page_url":   "https://example.com/cart",
			"page_title": "Cart",
		},
		map[string]any{
			"resolution_mode": "native_ref_first",
			"primary_kind":    "native_ref",
			"element_ref":     "e12",
			"selector":        `button.buy`,
			"locator_plan": []any{
				map[string]any{"kind": "native_ref", "native_ref": "e12"},
				map[string]any{"kind": "selector", "selector": `button.buy`},
				map[string]any{"kind": "role_label", "role": "button", "label": "Buy"},
				map[string]any{"kind": "page_binding", "page_url": "https://example.com/cart", "page_title": "Cart"},
			},
		},
		adapter,
	)
	if err != nil {
		t.Fatalf("ResolveBrowserElementFromRemotePayload returned error: %v", err)
	}
	if !reflect.DeepEqual(adapter.events, []string{
		"page:page_binding",
		"native:e12",
		"selector:button.buy",
		"semantic:role_label",
	}) {
		t.Fatalf("unexpected payload adapter call order: %#v", adapter.events)
	}
	if result.MatchedAttempt == nil || result.MatchedAttempt.Candidate.Kind != "role_label" {
		t.Fatalf("unexpected payload resolution result: %#v", result)
	}
}
