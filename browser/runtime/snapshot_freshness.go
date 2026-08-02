package browserruntime

import "strings"

const (
	BrowserSnapshotFreshnessStateFresh         = "fresh"
	BrowserSnapshotFreshnessStateStale         = "stale"
	BrowserSnapshotFreshnessStateMissing       = "missing"
	BrowserSnapshotFreshnessStateNotApplicable = "not_applicable"

	BrowserSnapshotFreshnessSourceSnapshotRef      = "snapshot_ref"
	BrowserSnapshotFreshnessSourceElementHint      = "element_hint"
	BrowserSnapshotFreshnessSourcePageBinding      = "page_binding"
	BrowserSnapshotFreshnessSourceResolverOutcome  = "resolver_outcome"
	BrowserSnapshotFreshnessSourceRecoverySnapshot = "recovery_snapshot"

	BrowserSnapshotRefreshReasonActionChangedDOM     = "action_changed_dom"
	BrowserSnapshotRefreshReasonPageChanged          = "page_changed"
	BrowserSnapshotRefreshReasonFramePathChanged     = "frame_path_changed"
	BrowserSnapshotRefreshReasonSelectorIndexInvalid = "selector_index_invalid"
	BrowserSnapshotRefreshReasonAmbiguousTarget      = "ambiguous_target"
	BrowserSnapshotRefreshReasonResolverMiss         = "resolver_miss"
	BrowserSnapshotRefreshReasonSnapshotUnavailable  = "snapshot_unavailable"

	BrowserSnapshotRecoveryUseRecoverySnapshot = "use_recovery_snapshot"
)

// BrowserSnapshotFreshness describes whether an action is operating from a
// usable browser snapshot, a stale snapshot reference, or no snapshot evidence.
// It complements resolver/actionability evidence instead of replacing it.
type BrowserSnapshotFreshness struct {
	State              string   `json:"state,omitempty"`
	Source             string   `json:"source,omitempty"`
	SnapshotEpoch      string   `json:"snapshot_epoch,omitempty"`
	TargetKind         string   `json:"target_kind,omitempty"`
	Target             string   `json:"target,omitempty"`
	PageURL            string   `json:"page_url,omitempty"`
	PageTitle          string   `json:"page_title,omitempty"`
	RefreshRecommended bool     `json:"refresh_recommended,omitempty"`
	RefreshReason      string   `json:"refresh_reason,omitempty"`
	RecoveryAction     string   `json:"recovery_action,omitempty"`
	NextStepAlias      string   `json:"next_step_alias,omitempty"`
	Notes              []string `json:"notes,omitempty"`
}

// BrowserSnapshotFreshnessRequest carries action-local evidence used to build
// a backend-neutral freshness contract.
type BrowserSnapshotFreshnessRequest struct {
	Action                    string
	TargetKind                string
	Target                    string
	SnapshotEpoch             string
	SnapshotSource            string
	PageURL                   string
	PageTitle                 string
	DOMChanged                bool
	ResolverOutcome           *BrowserElementResolverOutcome
	RecoverySnapshotAvailable bool
}

// BuildBrowserSnapshotFreshness builds a compact machine-readable contract for
// deciding whether the next browser step can use existing snapshot evidence or
// should refresh it first.
func BuildBrowserSnapshotFreshness(req BrowserSnapshotFreshnessRequest) *BrowserSnapshotFreshness {
	action := normalizeBrowserActionabilityAction(req.Action)
	targetKind := strings.TrimSpace(req.TargetKind)
	target := strings.TrimSpace(req.Target)
	snapshotEpoch := strings.TrimSpace(req.SnapshotEpoch)
	source := normalizeBrowserSnapshotFreshnessSource(req.SnapshotSource)
	outcome := browserElementResolverOutcomeNormalizedClone(req.ResolverOutcome)
	if action == "" && targetKind == "" && target == "" && snapshotEpoch == "" && source == "" && outcome == nil && !req.DOMChanged && !req.RecoverySnapshotAvailable {
		return nil
	}

	freshness := &BrowserSnapshotFreshness{
		Source:        source,
		SnapshotEpoch: snapshotEpoch,
		TargetKind:    targetKind,
		Target:        target,
		PageURL:       strings.TrimSpace(req.PageURL),
		PageTitle:     strings.TrimSpace(req.PageTitle),
	}
	if outcome != nil {
		freshness.RecoveryAction = strings.TrimSpace(outcome.RecoveryAction)
		freshness.NextStepAlias = strings.TrimSpace(outcome.NextStepAlias)
	}

	refreshReason := browserSnapshotRefreshReason(outcome)
	switch {
	case req.RecoverySnapshotAvailable:
		freshness.State = BrowserSnapshotFreshnessStateFresh
		freshness.Source = BrowserSnapshotFreshnessSourceRecoverySnapshot
		freshness.RefreshReason = refreshReason
		freshness.RefreshRecommended = false
		freshness.RecoveryAction = BrowserSnapshotRecoveryUseRecoverySnapshot
		freshness.NextStepAlias = firstNonEmptyString(freshness.NextStepAlias, "snapshot")
		freshness.Notes = append(freshness.Notes, "recovery_snapshot_available")
	case req.DOMChanged:
		freshness.State = BrowserSnapshotFreshnessStateStale
		freshness.RefreshRecommended = true
		freshness.RefreshReason = BrowserSnapshotRefreshReasonActionChangedDOM
	case browserSnapshotOutcomeRecommendsRefresh(outcome):
		freshness.State = browserSnapshotFreshnessStaleOrMissing(snapshotEpoch, source, target)
		freshness.RefreshRecommended = true
		freshness.RefreshReason = firstNonEmptyString(refreshReason, BrowserSnapshotRefreshReasonResolverMiss)
	case snapshotEpoch != "":
		freshness.State = BrowserSnapshotFreshnessStateFresh
	case source != "":
		freshness.State = BrowserSnapshotFreshnessStateFresh
	default:
		freshness.State = BrowserSnapshotFreshnessStateMissing
		freshness.RefreshRecommended = true
		freshness.RefreshReason = BrowserSnapshotRefreshReasonSnapshotUnavailable
	}

	if freshness.Source == "" {
		freshness.Source = browserSnapshotFreshnessDefaultSource(freshness.State, outcome)
	}
	if freshness.RefreshRecommended {
		freshness.RecoveryAction = firstNonEmptyString(freshness.RecoveryAction, "browser action=snapshot")
		freshness.NextStepAlias = firstNonEmptyString(freshness.NextStepAlias, browserElementResolverNextStepAlias(freshness.RecoveryAction), "snapshot")
	}
	if freshness.RefreshReason == "" && freshness.RefreshRecommended {
		freshness.RefreshReason = BrowserSnapshotRefreshReasonSnapshotUnavailable
	}
	return normalizeBrowserSnapshotFreshness(freshness)
}

func cloneBrowserSnapshotFreshness(freshness *BrowserSnapshotFreshness) *BrowserSnapshotFreshness {
	if freshness == nil {
		return nil
	}
	cloned := *freshness
	if len(freshness.Notes) > 0 {
		cloned.Notes = append([]string(nil), freshness.Notes...)
	} else {
		cloned.Notes = nil
	}
	return &cloned
}

func normalizeBrowserSnapshotFreshness(freshness *BrowserSnapshotFreshness) *BrowserSnapshotFreshness {
	if freshness == nil {
		return nil
	}
	freshness.State = strings.TrimSpace(freshness.State)
	freshness.Source = normalizeBrowserSnapshotFreshnessSource(freshness.Source)
	freshness.SnapshotEpoch = strings.TrimSpace(freshness.SnapshotEpoch)
	freshness.TargetKind = strings.TrimSpace(freshness.TargetKind)
	freshness.Target = strings.TrimSpace(freshness.Target)
	freshness.PageURL = strings.TrimSpace(freshness.PageURL)
	freshness.PageTitle = strings.TrimSpace(freshness.PageTitle)
	freshness.RefreshReason = strings.TrimSpace(freshness.RefreshReason)
	freshness.RecoveryAction = strings.TrimSpace(freshness.RecoveryAction)
	freshness.NextStepAlias = strings.TrimSpace(freshness.NextStepAlias)
	freshness.Notes = normalizeBrowserSnapshotFreshnessNotes(freshness.Notes)
	if freshness.State == "" {
		return nil
	}
	if freshness.State == BrowserSnapshotFreshnessStateNotApplicable &&
		freshness.Source == "" &&
		freshness.SnapshotEpoch == "" &&
		freshness.TargetKind == "" &&
		freshness.Target == "" &&
		freshness.PageURL == "" &&
		freshness.PageTitle == "" &&
		!freshness.RefreshRecommended &&
		freshness.RefreshReason == "" &&
		freshness.RecoveryAction == "" &&
		freshness.NextStepAlias == "" &&
		len(freshness.Notes) == 0 {
		return nil
	}
	return freshness
}

func normalizeBrowserSnapshotFreshnessSource(source string) string {
	switch strings.ToLower(strings.TrimSpace(source)) {
	case "", "none":
		return ""
	case "snapshot", "snapshot_ref", "ref":
		return BrowserSnapshotFreshnessSourceSnapshotRef
	case "hint", "element_hint":
		return BrowserSnapshotFreshnessSourceElementHint
	case "page_binding":
		return BrowserSnapshotFreshnessSourcePageBinding
	case "resolver", "resolver_outcome":
		return BrowserSnapshotFreshnessSourceResolverOutcome
	case "recovery", "recovery_snapshot":
		return BrowserSnapshotFreshnessSourceRecoverySnapshot
	default:
		return strings.TrimSpace(source)
	}
}

func normalizeBrowserSnapshotFreshnessNotes(notes []string) []string {
	if len(notes) == 0 {
		return nil
	}
	out := make([]string, 0, len(notes))
	seen := make(map[string]struct{}, len(notes))
	for _, note := range notes {
		note = strings.TrimSpace(note)
		if note == "" {
			continue
		}
		if _, ok := seen[note]; ok {
			continue
		}
		seen[note] = struct{}{}
		out = append(out, note)
	}
	return out
}

func browserSnapshotOutcomeRecommendsRefresh(outcome *BrowserElementResolverOutcome) bool {
	if outcome == nil {
		return false
	}
	status := strings.TrimSpace(outcome.Status)
	if status != "" && status != "matched" {
		return true
	}
	if browserSnapshotRefreshReason(outcome) != "" {
		return true
	}
	return browserElementResolverNextStepAlias(strings.TrimSpace(outcome.RecoveryAction)) == "snapshot" ||
		strings.TrimSpace(outcome.NextStepAlias) == "snapshot"
}

func browserSnapshotFreshnessStaleOrMissing(snapshotEpoch string, source string, target string) string {
	if strings.TrimSpace(snapshotEpoch) == "" && strings.TrimSpace(source) == "" && strings.TrimSpace(target) == "" {
		return BrowserSnapshotFreshnessStateMissing
	}
	return BrowserSnapshotFreshnessStateStale
}

func browserSnapshotFreshnessDefaultSource(state string, outcome *BrowserElementResolverOutcome) string {
	if outcome != nil {
		if strings.TrimSpace(outcome.Status) == "page_binding_blocked" {
			return BrowserSnapshotFreshnessSourcePageBinding
		}
		if strings.TrimSpace(outcome.Status) != "" {
			return BrowserSnapshotFreshnessSourceResolverOutcome
		}
	}
	if state == BrowserSnapshotFreshnessStateFresh {
		return BrowserSnapshotFreshnessSourceSnapshotRef
	}
	return ""
}

func browserSnapshotRefreshReason(outcome *BrowserElementResolverOutcome) string {
	if outcome == nil {
		return ""
	}
	status := strings.ToLower(strings.TrimSpace(outcome.Status))
	blockedBy := strings.ToLower(strings.TrimSpace(outcome.BlockedBy))
	ambiguityClass := strings.ToLower(strings.TrimSpace(outcome.AmbiguityClass))
	manualRetryHint := strings.ToLower(strings.TrimSpace(outcome.ManualRetryHint))
	switch {
	case status == "page_binding_blocked",
		blockedBy == "page_binding",
		blockedBy == "page_url",
		blockedBy == "page_origin",
		blockedBy == "page_path",
		blockedBy == "page_title",
		blockedBy == "tab_index":
		return BrowserSnapshotRefreshReasonPageChanged
	case blockedBy == "frame_path", strings.Contains(blockedBy, "frame_path"):
		return BrowserSnapshotRefreshReasonFramePathChanged
	case blockedBy == "selector_index_out_of_range",
		blockedBy == "selector_index_filtered_out",
		strings.Contains(blockedBy, "selector_index"):
		return BrowserSnapshotRefreshReasonSelectorIndexInvalid
	case strings.Contains(blockedBy, "multiple_candidates"),
		strings.Contains(ambiguityClass, "multiple"),
		strings.Contains(manualRetryHint, "ordinal"),
		strings.Contains(manualRetryHint, "specificity"):
		return BrowserSnapshotRefreshReasonAmbiguousTarget
	case status != "" && status != "matched":
		return BrowserSnapshotRefreshReasonResolverMiss
	default:
		return ""
	}
}
