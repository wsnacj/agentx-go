package browserruntime

import "strings"

// ProjectSharedSessionBrowserBindingEvaluationWithSelectionProjection overlays
// tool-facing profile/target selections onto an existing binding evaluation so
// tools callers do not need to locally restitch selected identity before
// asking browserruntime to project top-level session surfaces.
func ProjectSharedSessionBrowserBindingEvaluationWithSelectionProjection(
	evaluation SharedSessionBrowserBindingEvaluation,
	projection *SharedSessionBrowserSelectionProjection,
) SharedSessionBrowserBindingEvaluation {
	if projection == nil {
		return evaluation
	}
	out := evaluation
	if merged := mergeSharedSessionBrowserProfileSelectionProjection(
		out.Snapshot.SelectedProfileSelection,
		projection.ProfileSelection,
	); merged != nil {
		out.Snapshot.SelectedProfileSelection = merged
	}
	if merged := mergeSharedSessionBrowserTargetSelectionProjection(
		out.Snapshot.SelectedTargetSelection,
		projection.TargetSelection,
	); merged != nil {
		out.Snapshot.SelectedTargetSelection = merged
		out.Snapshot.CurrentTargetID = firstNonEmptyBindingString(
			strings.TrimSpace(out.Snapshot.CurrentTargetID),
			strings.TrimSpace(merged.ID),
		)
		out.Snapshot.Summary.CurrentTargetID = firstNonEmptyBindingString(
			strings.TrimSpace(out.Snapshot.Summary.CurrentTargetID),
			strings.TrimSpace(out.Snapshot.CurrentTargetID),
		)
	}
	return sharedSessionBrowserFinalizeBindingEvaluationHandoff(out)
}

func mergeSharedSessionBrowserProfileSelectionProjection(
	current *SharedSessionBrowserProfileSelection,
	overlay *SharedSessionBrowserProfileSelection,
) *SharedSessionBrowserProfileSelection {
	if current == nil && overlay == nil {
		return nil
	}
	if current == nil {
		cloned := *overlay
		if sharedSessionBrowserProfileSelectionEmpty(cloned) {
			return nil
		}
		return &cloned
	}
	merged := *current
	if overlay != nil {
		merged.Backend = firstNonEmptyBindingString(
			strings.TrimSpace(merged.Backend),
			strings.TrimSpace(overlay.Backend),
		)
		merged.Profile = firstNonEmptyBindingString(
			strings.TrimSpace(merged.Profile),
			strings.TrimSpace(overlay.Profile),
		)
		merged.RuntimeTarget = firstNonEmptyBindingString(
			strings.TrimSpace(merged.RuntimeTarget),
			strings.TrimSpace(overlay.RuntimeTarget),
		)
		merged.BrowserApp = firstNonEmptyBindingString(
			strings.TrimSpace(merged.BrowserApp),
			strings.TrimSpace(overlay.BrowserApp),
		)
		merged.Source = firstNonEmptyBindingString(
			strings.TrimSpace(merged.Source),
			strings.TrimSpace(overlay.Source),
		)
	}
	if sharedSessionBrowserProfileSelectionEmpty(merged) {
		return nil
	}
	return &merged
}

func mergeSharedSessionBrowserTargetSelectionProjection(
	current *BrowserSessionTargetSelection,
	overlay *BrowserSessionTargetSelection,
) *BrowserSessionTargetSelection {
	if current == nil && overlay == nil {
		return nil
	}
	if current == nil {
		cloned := *overlay
		if sharedSessionBrowserTargetSelectionEmpty(cloned) {
			return nil
		}
		return &cloned
	}
	merged := *current
	if overlay != nil {
		merged.ID = firstNonEmptyBindingString(
			strings.TrimSpace(merged.ID),
			strings.TrimSpace(overlay.ID),
		)
		if merged.TabIndex == 0 {
			merged.TabIndex = overlay.TabIndex
		}
		merged.URL = firstNonEmptyBindingString(
			strings.TrimSpace(merged.URL),
			strings.TrimSpace(overlay.URL),
		)
		merged.Title = firstNonEmptyBindingString(
			strings.TrimSpace(merged.Title),
			strings.TrimSpace(overlay.Title),
		)
		merged.Backend = firstNonEmptyBindingString(
			strings.TrimSpace(merged.Backend),
			strings.TrimSpace(overlay.Backend),
		)
		merged.Profile = firstNonEmptyBindingString(
			strings.TrimSpace(merged.Profile),
			strings.TrimSpace(overlay.Profile),
		)
		merged.RuntimeTarget = firstNonEmptyBindingString(
			strings.TrimSpace(merged.RuntimeTarget),
			strings.TrimSpace(overlay.RuntimeTarget),
		)
		merged.BrowserApp = firstNonEmptyBindingString(
			strings.TrimSpace(merged.BrowserApp),
			strings.TrimSpace(overlay.BrowserApp),
		)
		merged.Source = firstNonEmptyBindingString(
			strings.TrimSpace(merged.Source),
			strings.TrimSpace(overlay.Source),
		)
	}
	if sharedSessionBrowserTargetSelectionEmpty(merged) {
		return nil
	}
	return &merged
}
