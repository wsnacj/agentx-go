package browserruntime

import (
	"encoding/json"
	"strconv"
	"strings"
)

type BrowserElementRemoteProjection struct {
	ResolutionMode string
	PrimaryKind    string
	ElementRef     string
	Selector       string
	SelectorIndex  int
	FallbackPlan   []BrowserLocatorCandidate
}

type BrowserElementResolutionPlan struct {
	ResolutionMode string
	PrimaryKind    string
	ElementRef     string
	Selector       string
	SelectorIndex  int
	LocatorOrder   []string
	LocatorPlan    []BrowserLocatorCandidate
	MatchPlan      []BrowserLocatorCandidate
	PageBinding    *BrowserLocatorCandidate
}

// BrowserElementResolutionAttempt describes one ordered resolver attempt that a
// remote/CDP consumer can execute against its own element lookup primitives.
type BrowserElementResolutionAttempt struct {
	Index     int
	IsPrimary bool
	Candidate BrowserLocatorCandidate
}

// BrowserElementResolutionAttemptResult lets a concrete resolver preserve
// machine-readable miss context for an ordered attempt while still reporting
// whether that attempt matched.
type BrowserElementResolutionAttemptResult struct {
	Matched bool
	Outcome *BrowserElementResolverOutcome
}

// BrowserElementResolverCallbacks allows a concrete resolver implementation to
// reuse agentx's ordered resolver contract while keeping backend-specific
// element lookup details outside the browserruntime package.
type BrowserElementResolverCallbacks struct {
	ValidatePageBinding    func(BrowserLocatorCandidate) error
	ResolveAttempt         func(BrowserElementResolutionAttempt) (bool, error)
	ResolveAttemptDetailed func(BrowserElementResolutionAttempt) (BrowserElementResolutionAttemptResult, error)
}

// BrowserElementResolverAdapter is the managed-route facing adapter contract
// for future remote/CDP consumers. Backends only need to provide concrete
// lookup primitives for native refs, selectors, semantic locators, and page
// binding guards; browserruntime keeps ownership of the ordered fallback logic.
type BrowserElementResolverAdapter interface {
	ValidatePageBinding(BrowserLocatorCandidate) error
	ResolveNativeRef(string) (bool, error)
	ResolveSelector(string) (bool, error)
	ResolveSemanticLocator(BrowserLocatorCandidate) (bool, error)
}

// BrowserElementResolutionResult captures how a normalized resolver request was
// consumed by a callback-driven resolver implementation.
type BrowserElementResolutionResult struct {
	ResolutionMode      string
	PrimaryKind         string
	AttemptCount        int
	PageBinding         *BrowserLocatorCandidate
	MatchedAttempt      *BrowserElementResolutionAttempt
	FallbackFromAttempt *BrowserElementResolutionAttempt
	FallbackFromOutcome *BrowserElementResolverOutcome
}

// BrowserElementResolverOutcomeFromResult converts a callback-driven resolver
// result into a stable, machine-readable outcome contract that managed-route
// consumers can surface alongside backend action results.
func BrowserElementResolverOutcomeFromResult(result BrowserElementResolutionResult, err error) *BrowserElementResolverOutcome {
	outcome := &BrowserElementResolverOutcome{
		ResolutionMode: strings.TrimSpace(result.ResolutionMode),
		PrimaryKind:    strings.TrimSpace(result.PrimaryKind),
		AttemptCount:   result.AttemptCount,
	}
	switch {
	case err != nil && result.PageBinding != nil && result.MatchedAttempt == nil && result.AttemptCount == 0:
		outcome.Status = "page_binding_blocked"
		outcome.BlockedBy = browserElementResolverBlockedByForPageBinding(*result.PageBinding)
		outcome.Note = strings.TrimSpace(err.Error())
	case err != nil:
		outcome.Status = "resolution_failed"
		outcome.Note = strings.TrimSpace(err.Error())
	case result.MatchedAttempt != nil:
		outcome.Status = "matched"
		outcome.MatchedKind = strings.TrimSpace(result.MatchedAttempt.Candidate.Kind)
		outcome.MatchedIndex = result.MatchedAttempt.Index
		outcome.MatchedCandidateKind = outcome.MatchedKind
		outcome.ResolvedFramePath = strings.TrimSpace(result.MatchedAttempt.Candidate.FramePath)
		if result.MatchedAttempt.Candidate.SelectorIndex > 0 {
			outcome.ResolvedSelectorIndex = result.MatchedAttempt.Candidate.SelectorIndex
		}
		if result.FallbackFromAttempt != nil {
			outcome.FallbackFromKind = strings.TrimSpace(result.FallbackFromAttempt.Candidate.Kind)
			outcome.FallbackFromIndex = result.FallbackFromAttempt.Index
			outcome.FallbackFromCandidateStrength = browserElementResolverCandidateStrength(result.FallbackFromAttempt.Candidate)
			outcome.FallbackFromSpecificityFields = browserElementResolverCandidateSpecificityFields(result.FallbackFromAttempt.Candidate)
		}
		if result.FallbackFromOutcome != nil {
			fallbackOutcome := browserElementResolverOutcomeNormalizedClone(result.FallbackFromOutcome)
			outcome.FallbackFromBlockedBy = strings.TrimSpace(fallbackOutcome.BlockedBy)
			outcome.FallbackFromAmbiguityClass = strings.TrimSpace(fallbackOutcome.AmbiguityClass)
			if strength := strings.TrimSpace(fallbackOutcome.CandidateStrength); strength != "" {
				outcome.FallbackFromCandidateStrength = strength
			}
			outcome.FallbackFromManualRetryHint = strings.TrimSpace(fallbackOutcome.ManualRetryHint)
			if len(fallbackOutcome.SpecificityFields) > 0 {
				outcome.FallbackFromSpecificityFields = append([]string(nil), fallbackOutcome.SpecificityFields...)
			}
		}
		if outcome.MatchedKind != "" {
			outcome.Note = "resolved via " + outcome.MatchedKind
		}
	default:
		outcome.Status = "unresolved"
		outcome.Note = "no element matched resolver plan"
	}
	return outcome.Normalized()
}

func browserElementResolverBlockedByForPageBinding(candidate BrowserLocatorCandidate) string {
	candidate = browserLocatorCandidateTrimmed(candidate)
	switch {
	case candidate.PageURL != "":
		return "page_url"
	case candidate.PageOrigin != "":
		return "page_origin"
	case candidate.PagePath != "":
		return "page_path"
	case candidate.PageTitle != "":
		return "page_title"
	case candidate.TabIndex > 0:
		return "tab_index"
	case candidate.FramePath != "":
		return "frame_path"
	default:
		return strings.TrimSpace(candidate.Kind)
	}
}

func browserElementResolverOutcomeNormalizedClone(outcome *BrowserElementResolverOutcome) *BrowserElementResolverOutcome {
	normalized := outcome.Normalized()
	if normalized == nil {
		return nil
	}
	cloned := *normalized
	if len(normalized.FallbackFromSpecificityFields) > 0 {
		cloned.FallbackFromSpecificityFields = append([]string(nil), normalized.FallbackFromSpecificityFields...)
	} else {
		cloned.FallbackFromSpecificityFields = nil
	}
	if len(normalized.SpecificityFields) > 0 {
		cloned.SpecificityFields = append([]string(nil), normalized.SpecificityFields...)
	} else {
		cloned.SpecificityFields = nil
	}
	return &cloned
}

func browserElementResolverRecoveryAction(status string, blockedBy string) string {
	switch strings.TrimSpace(status) {
	case "page_binding_blocked":
		return "browser action=snapshot"
	case "unresolved":
		return "browser action=snapshot"
	case "resolution_failed":
		switch strings.TrimSpace(blockedBy) {
		case "page_binding", "page_url", "page_origin", "page_path", "page_title", "tab_index", "frame_path":
			return "browser action=snapshot"
		}
		return "browser action=refresh"
	default:
		return ""
	}
}

func browserElementResolverNextStepAlias(recoveryAction string) string {
	switch strings.TrimSpace(recoveryAction) {
	case "browser action=snapshot":
		return "snapshot"
	case "browser action=refresh":
		return "refresh"
	default:
		return ""
	}
}

func browserElementResolverManualRetryHint(blockedBy string, recoveryAction string) string {
	switch strings.TrimSpace(blockedBy) {
	case "multiple_candidates", "multiple_candidates_same_semantic":
		return "add_specificity"
	case "multiple_candidates_filtered":
		return "add_ordinal"
	case "selector_index_out_of_range", "selector_index_filtered_out",
		"page_binding", "page_url", "page_origin", "page_path", "page_title", "tab_index", "frame_path":
		return "refresh_snapshot"
	}
	if browserElementResolverNextStepAlias(recoveryAction) == "snapshot" {
		return "refresh_snapshot"
	}
	return ""
}

func browserElementResolverCandidateStrength(candidate BrowserLocatorCandidate) string {
	switch strings.TrimSpace(candidate.Kind) {
	case "href", "role_label", "tag_label":
		return "strong"
	case "label":
		return "medium"
	case "placeholder", "tag_type", "tag", "type":
		return "weak"
	default:
		return ""
	}
}

func browserElementResolverCandidateSpecificityFields(candidate BrowserLocatorCandidate) []string {
	candidate = browserLocatorCandidateTrimmed(candidate)
	baseFields := map[string]struct{}{}
	switch candidate.Kind {
	case "role_label":
		baseFields["role"] = struct{}{}
		baseFields["label"] = struct{}{}
	case "tag_label":
		baseFields["tag"] = struct{}{}
		baseFields["label"] = struct{}{}
	case "label":
		baseFields["label"] = struct{}{}
	case "placeholder":
		baseFields["placeholder"] = struct{}{}
	case "tag_type":
		baseFields["tag"] = struct{}{}
		baseFields["type"] = struct{}{}
	case "tag":
		baseFields["tag"] = struct{}{}
	case "type":
		baseFields["type"] = struct{}{}
	case "href":
		baseFields["href"] = struct{}{}
	}
	out := make([]string, 0, 4)
	appendField := func(name, value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		if _, ok := baseFields[name]; ok {
			return
		}
		for _, existing := range out {
			if existing == name {
				return
			}
		}
		out = append(out, name)
	}
	appendField("role", candidate.Role)
	appendField("label", candidate.Label)
	appendField("tag", candidate.Tag)
	appendField("type", candidate.Type)
	appendField("href", candidate.Href)
	appendField("placeholder", candidate.Placeholder)
	return out
}

// Normalized trims and stabilizes a resolver outcome so managed-route
// consumers can rely on a shared recovery contract regardless of whether the
// outcome originated locally or came back from a remote/CDP backend.
func (o *BrowserElementResolverOutcome) Normalized() *BrowserElementResolverOutcome {
	if o == nil {
		return nil
	}
	normalized := *o
	normalized.Status = strings.TrimSpace(normalized.Status)
	normalized.ResolutionMode = strings.TrimSpace(normalized.ResolutionMode)
	normalized.PrimaryKind = strings.TrimSpace(normalized.PrimaryKind)
	normalized.MatchedKind = strings.TrimSpace(normalized.MatchedKind)
	normalized.MatchedCandidateKind = strings.TrimSpace(normalized.MatchedCandidateKind)
	normalized.ResolvedFramePath = strings.TrimSpace(normalized.ResolvedFramePath)
	normalized.FallbackFromKind = strings.TrimSpace(normalized.FallbackFromKind)
	normalized.FallbackFromBlockedBy = strings.TrimSpace(normalized.FallbackFromBlockedBy)
	normalized.FallbackFromAmbiguityClass = strings.TrimSpace(normalized.FallbackFromAmbiguityClass)
	normalized.FallbackFromCandidateStrength = strings.TrimSpace(normalized.FallbackFromCandidateStrength)
	normalized.FallbackFromManualRetryHint = strings.TrimSpace(normalized.FallbackFromManualRetryHint)
	normalized.CandidateKind = strings.TrimSpace(normalized.CandidateKind)
	normalized.CandidateStrength = strings.TrimSpace(normalized.CandidateStrength)
	normalized.AmbiguityClass = strings.TrimSpace(normalized.AmbiguityClass)
	normalized.RetryDisposition = strings.TrimSpace(normalized.RetryDisposition)
	normalized.ManualRetryHint = strings.TrimSpace(normalized.ManualRetryHint)
	normalized.NextStepAlias = strings.TrimSpace(normalized.NextStepAlias)
	normalized.BlockedBy = strings.TrimSpace(normalized.BlockedBy)
	if normalized.LocatorCount < 0 {
		normalized.LocatorCount = 0
	}
	if normalized.CandidateCount < 0 {
		normalized.CandidateCount = 0
	}
	if normalized.FallbackFromIndex < 0 {
		normalized.FallbackFromIndex = 0
	}
	if normalized.ResolvedSelectorIndex < 0 {
		normalized.ResolvedSelectorIndex = 0
	}
	if normalized.PreferredOrdinal < 0 {
		normalized.PreferredOrdinal = 0
	}
	normalized.RecoveryAction = strings.TrimSpace(normalized.RecoveryAction)
	normalized.Note = strings.TrimSpace(normalized.Note)
	if normalized.Status == "matched" && normalized.MatchedCandidateKind == "" {
		normalized.MatchedCandidateKind = normalized.MatchedKind
	}
	if normalized.RecoveryAction == "" {
		normalized.RecoveryAction = browserElementResolverRecoveryAction(normalized.Status, normalized.BlockedBy)
	}
	if normalized.NextStepAlias == "" {
		normalized.NextStepAlias = browserElementResolverNextStepAlias(normalized.RecoveryAction)
	}
	if normalized.ManualRetryHint == "" {
		normalized.ManualRetryHint = browserElementResolverManualRetryHint(normalized.BlockedBy, normalized.RecoveryAction)
	}
	return &normalized
}

func browserSplitLocatorPlan(plan []BrowserLocatorCandidate) ([]BrowserLocatorCandidate, *BrowserLocatorCandidate) {
	if len(plan) == 0 {
		return nil, nil
	}
	matchPlan := make([]BrowserLocatorCandidate, 0, len(plan))
	var pageBinding *BrowserLocatorCandidate
	for _, candidate := range plan {
		candidate = browserLocatorCandidateTrimmed(candidate)
		if !browserLocatorCandidateValid(candidate) {
			continue
		}
		if candidate.Kind == "page_binding" {
			if pageBinding == nil {
				value := candidate
				pageBinding = &value
			}
			continue
		}
		matchPlan = append(matchPlan, candidate)
	}
	if len(matchPlan) == 0 {
		matchPlan = nil
	}
	return matchPlan, pageBinding
}

func browserLocatorCandidateTrimmed(candidate BrowserLocatorCandidate) BrowserLocatorCandidate {
	candidate.Kind = strings.TrimSpace(candidate.Kind)
	candidate.Selector = strings.TrimSpace(candidate.Selector)
	candidate.NativeRef = strings.TrimSpace(candidate.NativeRef)
	candidate.FramePath = strings.TrimSpace(candidate.FramePath)
	candidate.Role = strings.TrimSpace(candidate.Role)
	candidate.Tag = strings.TrimSpace(candidate.Tag)
	candidate.Label = strings.TrimSpace(candidate.Label)
	candidate.Type = strings.TrimSpace(candidate.Type)
	candidate.Href = strings.TrimSpace(candidate.Href)
	candidate.Placeholder = strings.TrimSpace(candidate.Placeholder)
	candidate.PageURL = strings.TrimSpace(candidate.PageURL)
	candidate.PageOrigin = strings.TrimSpace(candidate.PageOrigin)
	candidate.PagePath = strings.TrimSpace(candidate.PagePath)
	candidate.PageTitle = strings.TrimSpace(candidate.PageTitle)
	return candidate
}

func browserLocatorCandidateValid(candidate BrowserLocatorCandidate) bool {
	candidate = browserLocatorCandidateTrimmed(candidate)
	switch candidate.Kind {
	case "native_ref":
		return candidate.NativeRef != ""
	case "selector":
		return candidate.Selector != ""
	case "href":
		return candidate.Href != ""
	case "label":
		return candidate.Label != ""
	case "role_label":
		return candidate.Role != "" && candidate.Label != ""
	case "tag_label":
		return candidate.Tag != "" && candidate.Label != ""
	case "placeholder":
		return candidate.Placeholder != ""
	case "tag_type":
		return candidate.Tag != "" && candidate.Type != ""
	case "tag":
		return candidate.Tag != ""
	case "type":
		return candidate.Type != ""
	case "page_binding":
		return candidate.PageURL != "" || candidate.PageOrigin != "" || candidate.PagePath != "" || candidate.PageTitle != "" || candidate.TabIndex > 0 || candidate.FramePath != ""
	default:
		return false
	}
}

func browserLocatorCandidateKey(candidate BrowserLocatorCandidate) string {
	candidate = browserLocatorCandidateTrimmed(candidate)
	switch candidate.Kind {
	case "native_ref":
		return "native_ref|" + candidate.NativeRef + "|" + candidate.FramePath
	case "selector":
		return "selector|" + candidate.Selector + "|" + strconv.Itoa(candidate.SelectorIndex) + "|" + candidate.FramePath
	case "href":
		return "href|" + candidate.Href + "|" + strconv.Itoa(candidate.SelectorIndex) + "|" + candidate.FramePath
	case "role_label":
		return "role_label|" + candidate.Role + "|" + candidate.Label + "|" + candidate.Tag + "|" + candidate.Type + "|" + candidate.Href + "|" + candidate.Placeholder + "|" + strconv.Itoa(candidate.SelectorIndex) + "|" + candidate.FramePath
	case "tag_label":
		return "tag_label|" + candidate.Tag + "|" + candidate.Type + "|" + candidate.Label + "|" + candidate.Href + "|" + candidate.Placeholder + "|" + strconv.Itoa(candidate.SelectorIndex) + "|" + candidate.FramePath
	case "label":
		return "label|" + candidate.Label + "|" + candidate.Tag + "|" + candidate.Type + "|" + candidate.Href + "|" + candidate.Placeholder + "|" + strconv.Itoa(candidate.SelectorIndex) + "|" + candidate.FramePath
	case "placeholder":
		return "placeholder|" + candidate.Placeholder + "|" + candidate.Tag + "|" + candidate.Type + "|" + strconv.Itoa(candidate.SelectorIndex) + "|" + candidate.FramePath
	case "tag_type":
		return "tag_type|" + candidate.Tag + "|" + candidate.Type + "|" + strconv.Itoa(candidate.SelectorIndex) + "|" + candidate.FramePath
	case "tag":
		return "tag|" + candidate.Tag + "|" + strconv.Itoa(candidate.SelectorIndex) + "|" + candidate.FramePath
	case "type":
		return "type|" + candidate.Tag + "|" + candidate.Type + "|" + strconv.Itoa(candidate.SelectorIndex) + "|" + candidate.FramePath
	case "page_binding":
		return "page_binding|" + candidate.PageURL + "|" + candidate.PageOrigin + "|" + candidate.PagePath + "|" + candidate.PageTitle + "|" + strconv.Itoa(candidate.TabIndex) + "|" + candidate.FramePath
	default:
		return ""
	}
}

func browserResolverPrimaryCandidate(kind string, elementRef string, selector string, selectorIndex int, framePath string) BrowserLocatorCandidate {
	switch strings.TrimSpace(kind) {
	case "native_ref":
		return BrowserLocatorCandidate{Kind: "native_ref", NativeRef: strings.TrimSpace(elementRef), FramePath: strings.TrimSpace(framePath)}
	case "selector":
		return BrowserLocatorCandidate{Kind: "selector", Selector: strings.TrimSpace(selector), SelectorIndex: selectorIndex, FramePath: strings.TrimSpace(framePath)}
	default:
		return BrowserLocatorCandidate{}
	}
}

func browserResolverPreferredFramePath(r *BrowserElementResolverRequest) string {
	if r == nil {
		return ""
	}
	if framePath := strings.TrimSpace(r.FramePath); framePath != "" {
		return framePath
	}
	for _, source := range [][]BrowserLocatorCandidate{r.MatchPlan, r.LocatorPlan} {
		for _, candidate := range source {
			if framePath := strings.TrimSpace(candidate.FramePath); framePath != "" {
				return framePath
			}
		}
	}
	return ""
}

func (r *BrowserElementResolverRequest) EffectivePrimaryKind() string {
	if r == nil {
		return ""
	}
	kind := strings.TrimSpace(r.PrimaryKind)
	switch kind {
	case "native_ref":
		if strings.TrimSpace(r.ElementRef) != "" {
			return kind
		}
	case "selector":
		if strings.TrimSpace(r.Selector) != "" {
			return kind
		}
	}
	switch strings.TrimSpace(r.ResolutionMode) {
	case "native_ref_first":
		if strings.TrimSpace(r.ElementRef) != "" {
			return "native_ref"
		}
	case "selector_first":
		if strings.TrimSpace(r.Selector) != "" {
			return "selector"
		}
	}
	if strings.TrimSpace(r.ElementRef) != "" {
		return "native_ref"
	}
	if strings.TrimSpace(r.Selector) != "" {
		return "selector"
	}
	for _, candidate := range r.LocatorPlan {
		candidate = browserLocatorCandidateTrimmed(candidate)
		switch candidate.Kind {
		case "native_ref":
			if candidate.NativeRef != "" {
				return "native_ref"
			}
		case "selector":
			if candidate.Selector != "" {
				return "selector"
			}
		}
	}
	return ""
}

func (r *BrowserElementResolverRequest) EffectiveResolutionMode() string {
	if r == nil {
		return ""
	}
	mode := strings.TrimSpace(r.ResolutionMode)
	switch mode {
	case "native_ref_first", "selector_first", "locator_plan_only":
		return mode
	}
	switch r.EffectivePrimaryKind() {
	case "native_ref":
		return "native_ref_first"
	case "selector":
		return "selector_first"
	}
	for _, candidate := range r.LocatorPlan {
		if browserLocatorCandidateValid(candidate) {
			return "locator_plan_only"
		}
	}
	return ""
}

func (r *BrowserElementResolverRequest) EffectiveLocatorPlan() []BrowserLocatorCandidate {
	if r == nil {
		return nil
	}
	mode := strings.TrimSpace(r.ResolutionMode)
	primaryKind := r.EffectivePrimaryKind()
	framePath := browserResolverPreferredFramePath(r)
	primary := browserResolverPrimaryCandidate(primaryKind, r.ElementRef, r.Selector, r.SelectorIndex, framePath)
	secondary := make([]BrowserLocatorCandidate, 0, 2)
	if primaryKind != "native_ref" && strings.TrimSpace(r.ElementRef) != "" {
		secondary = append(secondary, BrowserLocatorCandidate{Kind: "native_ref", NativeRef: strings.TrimSpace(r.ElementRef), FramePath: framePath})
	}
	if primaryKind != "selector" && strings.TrimSpace(r.Selector) != "" {
		secondary = append(secondary, BrowserLocatorCandidate{Kind: "selector", Selector: strings.TrimSpace(r.Selector), SelectorIndex: r.SelectorIndex, FramePath: framePath})
	}
	if mode == "selector_first" && len(secondary) == 2 {
		secondary[0], secondary[1] = secondary[1], secondary[0]
	}

	plan := make([]BrowserLocatorCandidate, 0, len(r.LocatorPlan)+1+len(secondary))
	if browserLocatorCandidateValid(primary) {
		plan = append(plan, browserLocatorCandidateTrimmed(primary))
	}
	for _, candidate := range secondary {
		if browserLocatorCandidateValid(candidate) {
			plan = append(plan, browserLocatorCandidateTrimmed(candidate))
		}
	}
	for _, candidate := range r.MatchPlan {
		candidate = browserLocatorCandidateTrimmed(candidate)
		if !browserLocatorCandidateValid(candidate) {
			continue
		}
		plan = append(plan, candidate)
	}
	if r.PageBinding != nil {
		candidate := browserLocatorCandidateTrimmed(*r.PageBinding)
		if browserLocatorCandidateValid(candidate) {
			plan = append(plan, candidate)
		}
	}
	for _, candidate := range r.LocatorPlan {
		candidate = browserLocatorCandidateTrimmed(candidate)
		if !browserLocatorCandidateValid(candidate) {
			continue
		}
		plan = append(plan, candidate)
	}
	if len(plan) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	out := make([]BrowserLocatorCandidate, 0, len(plan))
	for _, candidate := range plan {
		key := browserLocatorCandidateKey(candidate)
		if key == "" {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, candidate)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func (r *BrowserElementResolverRequest) EffectiveLocatorOrder() []string {
	if r == nil {
		return nil
	}
	plan := r.EffectiveLocatorPlan()
	if len(plan) == 0 {
		return nil
	}
	order := make([]string, 0, len(plan))
	seen := map[string]struct{}{}
	for _, candidate := range plan {
		kind := strings.TrimSpace(candidate.Kind)
		if kind == "" {
			continue
		}
		if _, exists := seen[kind]; exists {
			continue
		}
		seen[kind] = struct{}{}
		order = append(order, kind)
	}
	if len(order) == 0 {
		return nil
	}
	return order
}

func (r *BrowserElementResolverRequest) Normalized() *BrowserElementResolverRequest {
	if r == nil {
		return nil
	}
	plan := r.EffectiveLocatorPlan()
	mode := r.EffectiveResolutionMode()
	primaryKind := r.EffectivePrimaryKind()
	framePath := browserResolverPreferredFramePath(r)
	if len(plan) == 0 && mode == "" && primaryKind == "" && strings.TrimSpace(r.ElementRef) == "" && strings.TrimSpace(r.Selector) == "" {
		return nil
	}
	normalized := &BrowserElementResolverRequest{
		ResolutionMode: mode,
		PrimaryKind:    primaryKind,
		FramePath:      framePath,
		LocatorOrder:   r.EffectiveLocatorOrder(),
		LocatorPlan:    append([]BrowserLocatorCandidate(nil), plan...),
	}
	if primaryKind == "native_ref" {
		normalized.ElementRef = strings.TrimSpace(r.ElementRef)
	}
	for _, candidate := range plan {
		switch strings.TrimSpace(candidate.Kind) {
		case "native_ref":
			if normalized.ElementRef == "" {
				normalized.ElementRef = strings.TrimSpace(candidate.NativeRef)
			}
		case "selector":
			if normalized.Selector == "" {
				normalized.Selector = strings.TrimSpace(candidate.Selector)
			}
			if normalized.SelectorIndex <= 0 && candidate.SelectorIndex > 0 {
				normalized.SelectorIndex = candidate.SelectorIndex
			}
		}
		if normalized.FramePath == "" {
			normalized.FramePath = strings.TrimSpace(candidate.FramePath)
		}
	}
	if primaryKind == "selector" && normalized.Selector == "" {
		normalized.Selector = strings.TrimSpace(r.Selector)
	}
	if primaryKind == "selector" && normalized.SelectorIndex <= 0 && r.SelectorIndex > 0 {
		normalized.SelectorIndex = r.SelectorIndex
	}
	if normalized.ResolutionMode == "" && len(normalized.LocatorPlan) > 0 {
		normalized.ResolutionMode = "locator_plan_only"
	}
	matchPlan, pageBinding := browserSplitLocatorPlan(normalized.LocatorPlan)
	if len(matchPlan) > 0 {
		normalized.MatchPlan = append([]BrowserLocatorCandidate(nil), matchPlan...)
	}
	if pageBinding != nil {
		value := *pageBinding
		normalized.PageBinding = &value
	}
	return normalized
}

func (r *BrowserElementResolverRequest) EffectiveResolutionPlan() BrowserElementResolutionPlan {
	normalized := r.Normalized()
	if normalized == nil {
		return BrowserElementResolutionPlan{}
	}
	plan := BrowserElementResolutionPlan{
		ResolutionMode: strings.TrimSpace(normalized.ResolutionMode),
		PrimaryKind:    strings.TrimSpace(normalized.PrimaryKind),
		ElementRef:     strings.TrimSpace(normalized.ElementRef),
		Selector:       strings.TrimSpace(normalized.Selector),
		SelectorIndex:  normalized.SelectorIndex,
		LocatorOrder:   append([]string(nil), normalized.LocatorOrder...),
		LocatorPlan:    append([]BrowserLocatorCandidate(nil), normalized.LocatorPlan...),
	}
	plan.MatchPlan, plan.PageBinding = browserSplitLocatorPlan(normalized.LocatorPlan)
	return plan
}

func browserResolutionPlanCandidateIsPrimary(plan BrowserElementResolutionPlan, candidate BrowserLocatorCandidate) bool {
	candidate = browserLocatorCandidateTrimmed(candidate)
	switch strings.TrimSpace(plan.PrimaryKind) {
	case "native_ref":
		return candidate.Kind == "native_ref" && strings.TrimSpace(candidate.NativeRef) != "" && strings.TrimSpace(candidate.NativeRef) == strings.TrimSpace(plan.ElementRef)
	case "selector":
		return candidate.Kind == "selector" &&
			strings.TrimSpace(candidate.Selector) != "" &&
			strings.TrimSpace(candidate.Selector) == strings.TrimSpace(plan.Selector) &&
			candidate.SelectorIndex == plan.SelectorIndex
	default:
		return false
	}
}

// ResolveWith walks a normalized resolver request the same way a future
// remote/CDP resolver should: page binding is validated first, then ordered
// match attempts are tried until one matches or the plan is exhausted.
func (r *BrowserElementResolverRequest) ResolveWith(callbacks BrowserElementResolverCallbacks) (BrowserElementResolutionResult, error) {
	plan := r.EffectiveResolutionPlan()
	result := BrowserElementResolutionResult{
		ResolutionMode: plan.ResolutionMode,
		PrimaryKind:    plan.PrimaryKind,
	}
	if plan.PageBinding != nil {
		pageBinding := browserLocatorCandidateTrimmed(*plan.PageBinding)
		result.PageBinding = &pageBinding
		if callbacks.ValidatePageBinding != nil {
			if err := callbacks.ValidatePageBinding(pageBinding); err != nil {
				return result, err
			}
		}
	}
	if (callbacks.ResolveAttempt == nil && callbacks.ResolveAttemptDetailed == nil) || len(plan.MatchPlan) == 0 {
		return result, nil
	}
	var fallbackFromAttempt *BrowserElementResolutionAttempt
	var fallbackFromOutcome *BrowserElementResolverOutcome
	for idx, candidate := range plan.MatchPlan {
		attempt := BrowserElementResolutionAttempt{
			Index:     idx,
			IsPrimary: browserResolutionPlanCandidateIsPrimary(plan, candidate),
			Candidate: browserLocatorCandidateTrimmed(candidate),
		}
		result.AttemptCount++
		var (
			matched bool
			err     error
			outcome *BrowserElementResolverOutcome
		)
		if callbacks.ResolveAttemptDetailed != nil {
			var attemptResult BrowserElementResolutionAttemptResult
			attemptResult, err = callbacks.ResolveAttemptDetailed(attempt)
			if err != nil {
				return result, err
			}
			matched = attemptResult.Matched
			outcome = browserElementResolverOutcomeNormalizedClone(attemptResult.Outcome)
		} else {
			matched, err = callbacks.ResolveAttempt(attempt)
			if err != nil {
				return result, err
			}
		}
		if matched {
			matchedAttempt := attempt
			result.MatchedAttempt = &matchedAttempt
			if fallbackFromAttempt != nil {
				fallbackAttempt := *fallbackFromAttempt
				result.FallbackFromAttempt = &fallbackAttempt
			}
			if fallbackFromOutcome != nil {
				result.FallbackFromOutcome = browserElementResolverOutcomeNormalizedClone(fallbackFromOutcome)
			}
			return result, nil
		}
		fallbackAttempt := attempt
		fallbackFromAttempt = &fallbackAttempt
		fallbackFromOutcome = outcome
	}
	return result, nil
}

// BrowserElementResolverCallbacksFromAdapter converts a managed-route adapter
// into the lower-level callback form used by ResolveWith.
func BrowserElementResolverCallbacksFromAdapter(adapter BrowserElementResolverAdapter) BrowserElementResolverCallbacks {
	if adapter == nil {
		return BrowserElementResolverCallbacks{}
	}
	return BrowserElementResolverCallbacks{
		ValidatePageBinding: func(candidate BrowserLocatorCandidate) error {
			return adapter.ValidatePageBinding(browserLocatorCandidateTrimmed(candidate))
		},
		ResolveAttempt: func(attempt BrowserElementResolutionAttempt) (bool, error) {
			candidate := browserLocatorCandidateTrimmed(attempt.Candidate)
			switch candidate.Kind {
			case "native_ref":
				return adapter.ResolveNativeRef(strings.TrimSpace(candidate.NativeRef))
			case "selector":
				return adapter.ResolveSelector(strings.TrimSpace(candidate.Selector))
			default:
				return adapter.ResolveSemanticLocator(candidate)
			}
		},
	}
}

func BrowserElementHintFromLocatorPlan(plan []BrowserLocatorCandidate) *BrowserElementHint {
	if len(plan) == 0 {
		return nil
	}
	hint := &BrowserElementHint{}
	for _, candidate := range plan {
		if strings.TrimSpace(hint.FramePath) == "" {
			hint.FramePath = strings.TrimSpace(candidate.FramePath)
		}
		switch strings.TrimSpace(candidate.Kind) {
		case "native_ref":
			if strings.TrimSpace(hint.NativeRef) == "" {
				hint.NativeRef = strings.TrimSpace(candidate.NativeRef)
			}
		case "selector":
			if strings.TrimSpace(hint.Selector) == "" {
				hint.Selector = strings.TrimSpace(candidate.Selector)
			}
			if hint.SelectorIndex <= 0 && candidate.SelectorIndex > 0 {
				hint.SelectorIndex = candidate.SelectorIndex
			}
		case "href":
			if strings.TrimSpace(hint.Href) == "" {
				hint.Href = strings.TrimSpace(candidate.Href)
			}
			if hint.SelectorIndex <= 0 && candidate.SelectorIndex > 0 {
				hint.SelectorIndex = candidate.SelectorIndex
			}
		case "role_label":
			if strings.TrimSpace(hint.Role) == "" {
				hint.Role = strings.TrimSpace(candidate.Role)
			}
			if strings.TrimSpace(hint.Label) == "" {
				hint.Label = strings.TrimSpace(candidate.Label)
			}
			if strings.TrimSpace(hint.Tag) == "" {
				hint.Tag = strings.TrimSpace(candidate.Tag)
			}
			if strings.TrimSpace(hint.Type) == "" {
				hint.Type = strings.TrimSpace(candidate.Type)
			}
			if strings.TrimSpace(hint.Href) == "" {
				hint.Href = strings.TrimSpace(candidate.Href)
			}
			if strings.TrimSpace(hint.Placeholder) == "" {
				hint.Placeholder = strings.TrimSpace(candidate.Placeholder)
			}
			if hint.SelectorIndex <= 0 && candidate.SelectorIndex > 0 {
				hint.SelectorIndex = candidate.SelectorIndex
			}
		case "tag_label":
			if strings.TrimSpace(hint.Tag) == "" {
				hint.Tag = strings.TrimSpace(candidate.Tag)
			}
			if strings.TrimSpace(hint.Type) == "" {
				hint.Type = strings.TrimSpace(candidate.Type)
			}
			if strings.TrimSpace(hint.Label) == "" {
				hint.Label = strings.TrimSpace(candidate.Label)
			}
			if strings.TrimSpace(hint.Href) == "" {
				hint.Href = strings.TrimSpace(candidate.Href)
			}
			if strings.TrimSpace(hint.Placeholder) == "" {
				hint.Placeholder = strings.TrimSpace(candidate.Placeholder)
			}
			if hint.SelectorIndex <= 0 && candidate.SelectorIndex > 0 {
				hint.SelectorIndex = candidate.SelectorIndex
			}
		case "label":
			if strings.TrimSpace(hint.Label) == "" {
				hint.Label = strings.TrimSpace(candidate.Label)
			}
			if strings.TrimSpace(hint.Tag) == "" {
				hint.Tag = strings.TrimSpace(candidate.Tag)
			}
			if strings.TrimSpace(hint.Type) == "" {
				hint.Type = strings.TrimSpace(candidate.Type)
			}
			if strings.TrimSpace(hint.Href) == "" {
				hint.Href = strings.TrimSpace(candidate.Href)
			}
			if strings.TrimSpace(hint.Placeholder) == "" {
				hint.Placeholder = strings.TrimSpace(candidate.Placeholder)
			}
			if hint.SelectorIndex <= 0 && candidate.SelectorIndex > 0 {
				hint.SelectorIndex = candidate.SelectorIndex
			}
		case "placeholder":
			if strings.TrimSpace(hint.Placeholder) == "" {
				hint.Placeholder = strings.TrimSpace(candidate.Placeholder)
			}
			if hint.SelectorIndex <= 0 && candidate.SelectorIndex > 0 {
				hint.SelectorIndex = candidate.SelectorIndex
			}
		case "tag_type":
			if strings.TrimSpace(hint.Tag) == "" {
				hint.Tag = strings.TrimSpace(candidate.Tag)
			}
			if strings.TrimSpace(hint.Type) == "" {
				hint.Type = strings.TrimSpace(candidate.Type)
			}
			if hint.SelectorIndex <= 0 && candidate.SelectorIndex > 0 {
				hint.SelectorIndex = candidate.SelectorIndex
			}
		case "tag":
			if strings.TrimSpace(hint.Tag) == "" {
				hint.Tag = strings.TrimSpace(candidate.Tag)
			}
			if hint.SelectorIndex <= 0 && candidate.SelectorIndex > 0 {
				hint.SelectorIndex = candidate.SelectorIndex
			}
		case "type":
			if strings.TrimSpace(hint.Type) == "" {
				hint.Type = strings.TrimSpace(candidate.Type)
			}
			if hint.SelectorIndex <= 0 && candidate.SelectorIndex > 0 {
				hint.SelectorIndex = candidate.SelectorIndex
			}
		case "page_binding":
			if strings.TrimSpace(hint.PageURL) == "" {
				hint.PageURL = strings.TrimSpace(candidate.PageURL)
			}
			if strings.TrimSpace(hint.PageOrigin) == "" {
				hint.PageOrigin = strings.TrimSpace(candidate.PageOrigin)
			}
			if strings.TrimSpace(hint.PagePath) == "" {
				hint.PagePath = strings.TrimSpace(candidate.PagePath)
			}
			if strings.TrimSpace(hint.PageTitle) == "" {
				hint.PageTitle = strings.TrimSpace(candidate.PageTitle)
			}
			if hint.TabIndex <= 0 && candidate.TabIndex > 0 {
				hint.TabIndex = candidate.TabIndex
			}
		}
	}
	hint.LocatorPlan = hint.EffectiveLocatorPlan()
	hint.LocatorOrder = hint.EffectiveLocatorOrder()
	hint.ResolutionMode = hint.EffectiveResolutionMode()
	return hint
}

func (h *BrowserElementHint) EffectiveLocatorOrder() []string {
	if h == nil {
		return nil
	}
	if len(h.LocatorOrder) > 0 {
		order := make([]string, 0, len(h.LocatorOrder))
		for _, kind := range h.LocatorOrder {
			kind = strings.TrimSpace(kind)
			if kind == "" {
				continue
			}
			order = append(order, kind)
		}
		if len(order) > 0 {
			return order
		}
	}
	order := make([]string, 0, 8)
	if strings.TrimSpace(h.NativeRef) != "" {
		order = append(order, "native_ref")
	}
	if strings.TrimSpace(h.Selector) != "" {
		order = append(order, "selector")
	}
	if strings.TrimSpace(h.Href) != "" {
		order = append(order, "href")
	}
	if strings.TrimSpace(h.Label) != "" {
		switch {
		case strings.TrimSpace(h.Role) != "":
			order = append(order, "role_label")
		case strings.TrimSpace(h.Tag) != "":
			order = append(order, "tag_label")
		default:
			order = append(order, "label")
		}
	}
	if strings.TrimSpace(h.Placeholder) != "" {
		order = append(order, "placeholder")
	}
	switch {
	case strings.TrimSpace(h.Tag) != "" && strings.TrimSpace(h.Type) != "":
		order = append(order, "tag_type")
	case strings.TrimSpace(h.Tag) != "":
		order = append(order, "tag")
	case strings.TrimSpace(h.Type) != "":
		order = append(order, "type")
	}
	if strings.TrimSpace(h.PageURL) != "" ||
		strings.TrimSpace(h.PageOrigin) != "" ||
		strings.TrimSpace(h.PagePath) != "" ||
		strings.TrimSpace(h.PageTitle) != "" ||
		h.TabIndex > 0 {
		order = append(order, "page_binding")
	}
	return order
}

func (h *BrowserElementHint) EffectiveLocatorPlan() []BrowserLocatorCandidate {
	if h == nil {
		return nil
	}
	if len(h.LocatorPlan) > 0 {
		plan := make([]BrowserLocatorCandidate, 0, len(h.LocatorPlan))
		for _, candidate := range h.LocatorPlan {
			candidate.Kind = strings.TrimSpace(candidate.Kind)
			candidate.Selector = strings.TrimSpace(candidate.Selector)
			candidate.NativeRef = strings.TrimSpace(candidate.NativeRef)
			candidate.FramePath = strings.TrimSpace(candidate.FramePath)
			candidate.Role = strings.TrimSpace(candidate.Role)
			candidate.Tag = strings.TrimSpace(candidate.Tag)
			candidate.Label = strings.TrimSpace(candidate.Label)
			candidate.Type = strings.TrimSpace(candidate.Type)
			candidate.Href = strings.TrimSpace(candidate.Href)
			candidate.Placeholder = strings.TrimSpace(candidate.Placeholder)
			candidate.PageURL = strings.TrimSpace(candidate.PageURL)
			candidate.PageOrigin = strings.TrimSpace(candidate.PageOrigin)
			candidate.PagePath = strings.TrimSpace(candidate.PagePath)
			candidate.PageTitle = strings.TrimSpace(candidate.PageTitle)
			if candidate.Kind == "" {
				continue
			}
			plan = append(plan, candidate)
		}
		if len(plan) > 0 {
			return plan
		}
	}
	order := h.EffectiveLocatorOrder()
	if len(order) == 0 {
		return nil
	}
	plan := make([]BrowserLocatorCandidate, 0, len(order))
	for _, kind := range order {
		switch kind {
		case "native_ref":
			plan = append(plan, BrowserLocatorCandidate{
				Kind:      kind,
				NativeRef: strings.TrimSpace(h.NativeRef),
				FramePath: strings.TrimSpace(h.FramePath),
			})
		case "selector":
			plan = append(plan, BrowserLocatorCandidate{
				Kind:          kind,
				Selector:      strings.TrimSpace(h.Selector),
				SelectorIndex: h.SelectorIndex,
				FramePath:     strings.TrimSpace(h.FramePath),
			})
		case "href":
			plan = append(plan, BrowserLocatorCandidate{
				Kind:          kind,
				Href:          strings.TrimSpace(h.Href),
				SelectorIndex: h.SelectorIndex,
				FramePath:     strings.TrimSpace(h.FramePath),
			})
		case "role_label":
			if strings.TrimSpace(h.Role) != "" && strings.TrimSpace(h.Label) != "" {
				plan = append(plan, BrowserLocatorCandidate{
					Kind:          kind,
					Role:          strings.TrimSpace(h.Role),
					Label:         strings.TrimSpace(h.Label),
					Tag:           strings.TrimSpace(h.Tag),
					Type:          strings.TrimSpace(h.Type),
					Href:          strings.TrimSpace(h.Href),
					Placeholder:   strings.TrimSpace(h.Placeholder),
					SelectorIndex: h.SelectorIndex,
					FramePath:     strings.TrimSpace(h.FramePath),
				})
			}
		case "tag_label":
			if strings.TrimSpace(h.Tag) != "" && strings.TrimSpace(h.Label) != "" {
				plan = append(plan, BrowserLocatorCandidate{
					Kind:          kind,
					Tag:           strings.TrimSpace(h.Tag),
					Type:          strings.TrimSpace(h.Type),
					Label:         strings.TrimSpace(h.Label),
					Href:          strings.TrimSpace(h.Href),
					Placeholder:   strings.TrimSpace(h.Placeholder),
					SelectorIndex: h.SelectorIndex,
					FramePath:     strings.TrimSpace(h.FramePath),
				})
			}
		case "label":
			plan = append(plan, BrowserLocatorCandidate{
				Kind:          kind,
				Label:         strings.TrimSpace(h.Label),
				Tag:           strings.TrimSpace(h.Tag),
				Type:          strings.TrimSpace(h.Type),
				Href:          strings.TrimSpace(h.Href),
				Placeholder:   strings.TrimSpace(h.Placeholder),
				SelectorIndex: h.SelectorIndex,
				FramePath:     strings.TrimSpace(h.FramePath),
			})
		case "placeholder":
			plan = append(plan, BrowserLocatorCandidate{
				Kind:          kind,
				Tag:           strings.TrimSpace(h.Tag),
				Type:          strings.TrimSpace(h.Type),
				Placeholder:   strings.TrimSpace(h.Placeholder),
				SelectorIndex: h.SelectorIndex,
				FramePath:     strings.TrimSpace(h.FramePath),
			})
		case "tag_type":
			if strings.TrimSpace(h.Tag) != "" && strings.TrimSpace(h.Type) != "" {
				plan = append(plan, BrowserLocatorCandidate{
					Kind:          kind,
					Tag:           strings.TrimSpace(h.Tag),
					Type:          strings.TrimSpace(h.Type),
					SelectorIndex: h.SelectorIndex,
					FramePath:     strings.TrimSpace(h.FramePath),
				})
			}
		case "tag":
			plan = append(plan, BrowserLocatorCandidate{
				Kind:          kind,
				Tag:           strings.TrimSpace(h.Tag),
				SelectorIndex: h.SelectorIndex,
				FramePath:     strings.TrimSpace(h.FramePath),
			})
		case "type":
			plan = append(plan, BrowserLocatorCandidate{
				Kind:          kind,
				Tag:           strings.TrimSpace(h.Tag),
				Type:          strings.TrimSpace(h.Type),
				SelectorIndex: h.SelectorIndex,
				FramePath:     strings.TrimSpace(h.FramePath),
			})
		case "page_binding":
			plan = append(plan, BrowserLocatorCandidate{
				Kind:       kind,
				PageURL:    strings.TrimSpace(h.PageURL),
				PageOrigin: strings.TrimSpace(h.PageOrigin),
				PagePath:   strings.TrimSpace(h.PagePath),
				PageTitle:  strings.TrimSpace(h.PageTitle),
				TabIndex:   h.TabIndex,
			})
		}
	}
	if len(plan) == 0 {
		return nil
	}
	return plan
}

func (h *BrowserElementHint) EffectiveResolutionMode() string {
	if h == nil {
		return ""
	}
	plan := h.EffectiveLocatorPlan()
	for _, candidate := range plan {
		if strings.TrimSpace(candidate.Kind) == "native_ref" && strings.TrimSpace(candidate.NativeRef) != "" {
			return "native_ref_first"
		}
	}
	for _, candidate := range plan {
		if strings.TrimSpace(candidate.Kind) == "selector" && strings.TrimSpace(candidate.Selector) != "" {
			return "selector_first"
		}
	}
	if len(plan) > 0 {
		return "locator_plan_only"
	}
	return ""
}

func (h *BrowserElementHint) RemoteProjection() BrowserElementRemoteProjection {
	if h == nil {
		return BrowserElementRemoteProjection{}
	}
	plan := h.EffectiveLocatorPlan()
	mode := h.EffectiveResolutionMode()
	projection := BrowserElementRemoteProjection{
		ResolutionMode: mode,
	}
	if len(plan) == 0 {
		return projection
	}

	nativeIdx := -1
	selectorIdx := -1
	for idx, candidate := range plan {
		switch strings.TrimSpace(candidate.Kind) {
		case "native_ref":
			if nativeIdx < 0 && strings.TrimSpace(candidate.NativeRef) != "" {
				nativeIdx = idx
			}
		case "selector":
			if selectorIdx < 0 && strings.TrimSpace(candidate.Selector) != "" {
				selectorIdx = idx
			}
		}
	}

	used := map[int]struct{}{}
	switch mode {
	case "native_ref_first":
		if nativeIdx >= 0 {
			projection.PrimaryKind = "native_ref"
			projection.ElementRef = strings.TrimSpace(plan[nativeIdx].NativeRef)
			used[nativeIdx] = struct{}{}
		}
		if selectorIdx >= 0 {
			projection.Selector = strings.TrimSpace(plan[selectorIdx].Selector)
			projection.SelectorIndex = plan[selectorIdx].SelectorIndex
			used[selectorIdx] = struct{}{}
		}
	case "selector_first":
		if selectorIdx >= 0 {
			projection.PrimaryKind = "selector"
			projection.Selector = strings.TrimSpace(plan[selectorIdx].Selector)
			projection.SelectorIndex = plan[selectorIdx].SelectorIndex
			used[selectorIdx] = struct{}{}
		}
	case "locator_plan_only":
	}

	if len(plan) == len(used) {
		return projection
	}
	fallback := make([]BrowserLocatorCandidate, 0, len(plan))
	for idx, candidate := range plan {
		if _, skip := used[idx]; skip {
			continue
		}
		fallback = append(fallback, candidate)
	}
	if len(fallback) > 0 {
		projection.FallbackPlan = fallback
	}
	return projection
}

func (h *BrowserElementHint) RemoteHint() *BrowserElementHint {
	if h == nil {
		return nil
	}
	return BrowserElementHintFromLocatorPlan(h.RemoteProjection().FallbackPlan)
}

func (h *BrowserElementHint) RemoteResolver() *BrowserElementResolverRequest {
	if h == nil {
		return nil
	}
	order := h.EffectiveLocatorOrder()
	plan := h.EffectiveLocatorPlan()
	projection := h.RemoteProjection()
	if len(order) == 0 && len(plan) == 0 && strings.TrimSpace(projection.ElementRef) == "" && strings.TrimSpace(projection.Selector) == "" {
		return nil
	}
	return (&BrowserElementResolverRequest{
		ResolutionMode: strings.TrimSpace(projection.ResolutionMode),
		PrimaryKind:    strings.TrimSpace(projection.PrimaryKind),
		ElementRef:     strings.TrimSpace(projection.ElementRef),
		Selector:       strings.TrimSpace(projection.Selector),
		SelectorIndex:  projection.SelectorIndex,
		FramePath:      strings.TrimSpace(h.FramePath),
		LocatorOrder:   append([]string(nil), order...),
		LocatorPlan:    append([]BrowserLocatorCandidate(nil), plan...),
	}).Normalized()
}

func mergeBrowserElementHint(dst *BrowserElementHint, src *BrowserElementHint) *BrowserElementHint {
	if dst == nil {
		dst = &BrowserElementHint{}
	}
	if src == nil {
		return dst
	}
	if strings.TrimSpace(dst.Selector) == "" {
		dst.Selector = strings.TrimSpace(src.Selector)
	}
	if dst.SelectorIndex <= 0 && src.SelectorIndex > 0 {
		dst.SelectorIndex = src.SelectorIndex
	}
	if strings.TrimSpace(dst.NativeRef) == "" {
		dst.NativeRef = strings.TrimSpace(src.NativeRef)
	}
	if strings.TrimSpace(dst.FramePath) == "" {
		dst.FramePath = strings.TrimSpace(src.FramePath)
	}
	if strings.TrimSpace(dst.Role) == "" {
		dst.Role = strings.TrimSpace(src.Role)
	}
	if strings.TrimSpace(dst.Tag) == "" {
		dst.Tag = strings.TrimSpace(src.Tag)
	}
	if strings.TrimSpace(dst.Label) == "" {
		dst.Label = strings.TrimSpace(src.Label)
	}
	if strings.TrimSpace(dst.Type) == "" {
		dst.Type = strings.TrimSpace(src.Type)
	}
	if strings.TrimSpace(dst.Href) == "" {
		dst.Href = strings.TrimSpace(src.Href)
	}
	if strings.TrimSpace(dst.Placeholder) == "" {
		dst.Placeholder = strings.TrimSpace(src.Placeholder)
	}
	if strings.TrimSpace(dst.PageURL) == "" {
		dst.PageURL = strings.TrimSpace(src.PageURL)
	}
	if strings.TrimSpace(dst.PageOrigin) == "" {
		dst.PageOrigin = strings.TrimSpace(src.PageOrigin)
	}
	if strings.TrimSpace(dst.PagePath) == "" {
		dst.PagePath = strings.TrimSpace(src.PagePath)
	}
	if strings.TrimSpace(dst.PageTitle) == "" {
		dst.PageTitle = strings.TrimSpace(src.PageTitle)
	}
	if dst.TabIndex <= 0 && src.TabIndex > 0 {
		dst.TabIndex = src.TabIndex
	}
	if len(dst.LocatorOrder) == 0 && len(src.LocatorOrder) > 0 {
		dst.LocatorOrder = append([]string(nil), src.LocatorOrder...)
	}
	if len(dst.LocatorPlan) == 0 && len(src.LocatorPlan) > 0 {
		dst.LocatorPlan = append([]BrowserLocatorCandidate(nil), src.LocatorPlan...)
	}
	if strings.TrimSpace(dst.ResolutionMode) == "" {
		dst.ResolutionMode = strings.TrimSpace(src.ResolutionMode)
	}
	return dst
}

func resolverPreferredFramePathFromRemoteHint(hint *BrowserElementHint, resolver *BrowserElementResolverRequest) string {
	if hint != nil {
		if framePath := strings.TrimSpace(hint.FramePath); framePath != "" {
			return framePath
		}
	}
	return browserResolverPreferredFramePath(resolver)
}

func BrowserElementResolverRequestFromRemote(elementRef string, selector string, hint *BrowserElementHint, resolver *BrowserElementResolverRequest) *BrowserElementResolverRequest {
	elementRef = strings.TrimSpace(elementRef)
	selector = strings.TrimSpace(selector)

	var synthesized *BrowserElementResolverRequest
	synthHint := &BrowserElementHint{
		Selector:  selector,
		NativeRef: elementRef,
		FramePath: strings.TrimSpace(resolverPreferredFramePathFromRemoteHint(hint, resolver)),
	}
	synthHint = mergeBrowserElementHint(synthHint, hint)
	if value := synthHint.RemoteResolver(); value != nil {
		synthesized = value
	}

	if resolver == nil {
		return synthesized
	}

	out := &BrowserElementResolverRequest{
		ResolutionMode: strings.TrimSpace(resolver.ResolutionMode),
		PrimaryKind:    strings.TrimSpace(resolver.PrimaryKind),
		ElementRef:     strings.TrimSpace(resolver.ElementRef),
		Selector:       strings.TrimSpace(resolver.Selector),
		SelectorIndex:  resolver.SelectorIndex,
		FramePath:      strings.TrimSpace(resolver.FramePath),
		LocatorOrder:   append([]string(nil), resolver.LocatorOrder...),
		LocatorPlan:    append([]BrowserLocatorCandidate(nil), resolver.LocatorPlan...),
		MatchPlan:      append([]BrowserLocatorCandidate(nil), resolver.MatchPlan...),
	}
	if resolver.PageBinding != nil {
		pageBinding := browserLocatorCandidateTrimmed(*resolver.PageBinding)
		out.PageBinding = &pageBinding
	}
	if strings.TrimSpace(out.ElementRef) == "" {
		out.ElementRef = elementRef
	}
	if strings.TrimSpace(out.Selector) == "" {
		out.Selector = selector
	}
	if synthesized != nil {
		if strings.TrimSpace(out.ResolutionMode) == "" {
			out.ResolutionMode = strings.TrimSpace(synthesized.ResolutionMode)
		}
		if strings.TrimSpace(out.PrimaryKind) == "" {
			out.PrimaryKind = strings.TrimSpace(synthesized.PrimaryKind)
		}
		if strings.TrimSpace(out.ElementRef) == "" {
			out.ElementRef = strings.TrimSpace(synthesized.ElementRef)
		}
		if strings.TrimSpace(out.Selector) == "" {
			out.Selector = strings.TrimSpace(synthesized.Selector)
		}
		if out.SelectorIndex <= 0 && synthesized.SelectorIndex > 0 {
			out.SelectorIndex = synthesized.SelectorIndex
		}
		if strings.TrimSpace(out.FramePath) == "" {
			out.FramePath = strings.TrimSpace(synthesized.FramePath)
		}
		if len(out.LocatorOrder) == 0 && len(synthesized.LocatorOrder) > 0 {
			out.LocatorOrder = append([]string(nil), synthesized.LocatorOrder...)
		}
		if len(out.LocatorPlan) == 0 && len(synthesized.LocatorPlan) > 0 {
			out.LocatorPlan = append([]BrowserLocatorCandidate(nil), synthesized.LocatorPlan...)
		}
		if len(out.MatchPlan) == 0 && len(synthesized.MatchPlan) > 0 {
			out.MatchPlan = append([]BrowserLocatorCandidate(nil), synthesized.MatchPlan...)
		}
		if out.PageBinding == nil && synthesized.PageBinding != nil {
			pageBinding := browserLocatorCandidateTrimmed(*synthesized.PageBinding)
			out.PageBinding = &pageBinding
		}
	}
	if strings.TrimSpace(out.ResolutionMode) == "" && strings.TrimSpace(out.ElementRef) != "" {
		out.ResolutionMode = "native_ref_first"
	}
	if strings.TrimSpace(out.ResolutionMode) == "" && strings.TrimSpace(out.Selector) != "" {
		out.ResolutionMode = "selector_first"
	}
	if strings.TrimSpace(out.PrimaryKind) == "" {
		switch {
		case strings.TrimSpace(out.ElementRef) != "":
			out.PrimaryKind = "native_ref"
		case strings.TrimSpace(out.Selector) != "":
			out.PrimaryKind = "selector"
		}
	}
	if len(out.LocatorPlan) == 0 {
		if len(out.MatchPlan) > 0 {
			out.LocatorPlan = append([]BrowserLocatorCandidate(nil), out.MatchPlan...)
		}
		if out.PageBinding != nil {
			pageBinding := browserLocatorCandidateTrimmed(*out.PageBinding)
			out.LocatorPlan = append(out.LocatorPlan, pageBinding)
		}
	}
	if len(out.LocatorOrder) == 0 && len(out.LocatorPlan) > 0 {
		h := BrowserElementHintFromLocatorPlan(out.LocatorPlan)
		if h != nil {
			out.LocatorOrder = append([]string(nil), h.EffectiveLocatorOrder()...)
			if strings.TrimSpace(out.ResolutionMode) == "" {
				out.ResolutionMode = strings.TrimSpace(h.EffectiveResolutionMode())
			}
		}
	}
	if len(out.LocatorOrder) == 0 && len(out.LocatorPlan) == 0 && strings.TrimSpace(out.ElementRef) == "" && strings.TrimSpace(out.Selector) == "" {
		return nil
	}
	return out.Normalized()
}

// ResolveBrowserElementFromRemote is the service-facing entrypoint for future
// remote/CDP resolvers. It normalizes mixed remote inputs into a single
// BrowserElementResolverRequest, then executes the shared ordered resolution
// logic via ResolveWith.
func ResolveBrowserElementFromRemote(elementRef string, selector string, hint *BrowserElementHint, resolver *BrowserElementResolverRequest, callbacks BrowserElementResolverCallbacks) (BrowserElementResolutionResult, error) {
	normalized := BrowserElementResolverRequestFromRemote(elementRef, selector, hint, resolver)
	if normalized == nil {
		return BrowserElementResolutionResult{}, nil
	}
	return normalized.ResolveWith(callbacks)
}

// ResolveBrowserElementFromRemoteWithAdapter is the managed-route convenience
// entrypoint for future remote/CDP services. It combines remote payload
// normalization with the shared adapter-driven resolution path.
func ResolveBrowserElementFromRemoteWithAdapter(elementRef string, selector string, hint *BrowserElementHint, resolver *BrowserElementResolverRequest, adapter BrowserElementResolverAdapter) (BrowserElementResolutionResult, error) {
	return ResolveBrowserElementFromRemote(elementRef, selector, hint, resolver, BrowserElementResolverCallbacksFromAdapter(adapter))
}

func browserDecodeRemoteValue[T any](raw any) *T {
	if raw == nil {
		return nil
	}
	switch value := raw.(type) {
	case T:
		out := value
		return &out
	case *T:
		if value == nil {
			return nil
		}
		out := *value
		return &out
	}
	blob, err := json.Marshal(raw)
	if err != nil || len(blob) == 0 || string(blob) == "null" {
		return nil
	}
	var out T
	if err := json.Unmarshal(blob, &out); err != nil {
		return nil
	}
	return &out
}

// ResolveBrowserElementFromRemotePayload is the JSON/map-facing convenience
// entrypoint for future proxy/CDP services that receive generic decoded request
// payloads rather than typed BrowserElementHint / BrowserElementResolverRequest
// values.
func ResolveBrowserElementFromRemotePayload(elementRef string, selector string, rawHint any, rawResolver any, adapter BrowserElementResolverAdapter) (BrowserElementResolutionResult, error) {
	return ResolveBrowserElementFromRemoteWithAdapter(
		elementRef,
		selector,
		browserDecodeRemoteValue[BrowserElementHint](rawHint),
		browserDecodeRemoteValue[BrowserElementResolverRequest](rawResolver),
		adapter,
	)
}
