package browserruntime

import "strings"

// SharedSessionBrowserWorkbenchActionPlan captures the shared coordination
// action plan surfaced by workbench-style payloads.
type SharedSessionBrowserWorkbenchActionPlan struct {
	PrimaryBrowserAction      string
	PrimaryNodeAction         string
	NextStep                  string
	RecommendedBrowserActions []string
	RecommendedNodeActions    []string
}

// SharedSessionBrowserWorkbenchProjection captures the shared workbench shell
// that tools can apply onto payload-specific workbench fields.
type SharedSessionBrowserWorkbenchProjection struct {
	Sections                 []string
	Ready                    bool
	ActionPlan               *SharedSessionBrowserWorkbenchActionPlan
	ClearActionPlan          bool
	CoordinationSummary      *SharedSessionBrowserCoordinationSummary
	ClearCoordinationSummary bool
}

// SharedSessionBrowserWorkbenchProjectionRequest carries the shared inputs
// needed to build the stable workbench shell and optional coordination/action
// plan surface.
type SharedSessionBrowserWorkbenchProjectionRequest struct {
	HasProfileStatus        bool
	HasProfiles             bool
	HasSessionProjection    bool
	ExtraSections           []string
	SyncCoordinationSurface bool
	ClearActionPlan         bool
	Evaluation              *SharedSessionBrowserBindingEvaluation
	CoordinationPlan        *SharedSessionBrowserCoordinationPlan
}

// BuildSharedSessionBrowserWorkbenchProjection lowers the shared
// workbench/session state into the stable tool-facing workbench shell,
// including section composition and optional coordination/action-plan sync.
func BuildSharedSessionBrowserWorkbenchProjection(
	req SharedSessionBrowserWorkbenchProjectionRequest,
) SharedSessionBrowserWorkbenchProjection {
	projection := SharedSessionBrowserWorkbenchProjection{
		ClearActionPlan: req.ClearActionPlan,
	}
	if req.SyncCoordinationSurface {
		if plan, ok := sharedSessionBrowserWorkbenchCoordinationPlan(req); ok {
			summary := SharedSessionBrowserCoordinationSummary{
				State: strings.TrimSpace(plan.State),
			}
			projection.CoordinationSummary = &summary
			projection.ActionPlan = &SharedSessionBrowserWorkbenchActionPlan{
				PrimaryBrowserAction:      strings.TrimSpace(plan.PrimaryBrowserAction),
				PrimaryNodeAction:         strings.TrimSpace(plan.PrimaryNodeAction),
				NextStep:                  strings.TrimSpace(plan.NextStep),
				RecommendedBrowserActions: cloneSharedSessionBrowserWorkbenchActions(plan.RecommendedBrowserActions),
				RecommendedNodeActions:    cloneSharedSessionBrowserWorkbenchActions(plan.RecommendedNodeActions),
			}
			req.ExtraSections = mergeSharedSessionBrowserWorkbenchSections(req.ExtraSections, []string{"coordination"})
		} else {
			projection.ClearCoordinationSummary = true
		}
	} else {
		projection.ClearCoordinationSummary = true
	}

	sections := []string{"route"}
	if req.HasProfileStatus {
		sections = append(sections, "status")
	}
	if req.HasProfiles {
		sections = append(sections, "profiles")
	}
	if req.HasSessionProjection {
		sections = append(sections, "sessions")
	}
	projection.Sections = mergeSharedSessionBrowserWorkbenchSections(sections, req.ExtraSections)
	projection.Ready = len(projection.Sections) > 0
	return projection
}

func sharedSessionBrowserWorkbenchCoordinationPlan(
	req SharedSessionBrowserWorkbenchProjectionRequest,
) (SharedSessionBrowserCoordinationPlan, bool) {
	if plan := normalizeSharedSessionBrowserWorkbenchCoordinationPlan(req.CoordinationPlan); plan != nil {
		return *plan, true
	}
	if req.Evaluation == nil {
		return SharedSessionBrowserCoordinationPlan{}, false
	}
	if plan := normalizeSharedSessionBrowserWorkbenchCoordinationPlan(&req.Evaluation.Coordination.Plan); plan != nil {
		return *plan, true
	}
	return SharedSessionBrowserCoordinationPlan{}, false
}

func normalizeSharedSessionBrowserWorkbenchCoordinationPlan(
	plan *SharedSessionBrowserCoordinationPlan,
) *SharedSessionBrowserCoordinationPlan {
	if plan == nil {
		return nil
	}
	normalized := SharedSessionBrowserCoordinationPlan{
		State:                     strings.TrimSpace(plan.State),
		BrowserOnNode:             plan.BrowserOnNode,
		HasActiveNodeRun:          plan.HasActiveNodeRun,
		HasRunningBrowserProfile:  plan.HasRunningBrowserProfile,
		NeedsSessionSync:          plan.NeedsSessionSync,
		SyncAction:                strings.TrimSpace(plan.SyncAction),
		PrepareAction:             strings.TrimSpace(plan.PrepareAction),
		RestartAction:             strings.TrimSpace(plan.RestartAction),
		TeardownAction:            strings.TrimSpace(plan.TeardownAction),
		PrimaryBrowserAction:      strings.TrimSpace(plan.PrimaryBrowserAction),
		PrimaryNodeAction:         strings.TrimSpace(plan.PrimaryNodeAction),
		NextStep:                  strings.TrimSpace(plan.NextStep),
		RecommendedBrowserActions: cloneSharedSessionBrowserWorkbenchActions(plan.RecommendedBrowserActions),
		RecommendedNodeActions:    cloneSharedSessionBrowserWorkbenchActions(plan.RecommendedNodeActions),
	}
	if normalized.State == "" &&
		normalized.SyncAction == "" &&
		normalized.PrepareAction == "" &&
		normalized.RestartAction == "" &&
		normalized.TeardownAction == "" &&
		normalized.PrimaryBrowserAction == "" &&
		normalized.PrimaryNodeAction == "" &&
		normalized.NextStep == "" &&
		len(normalized.RecommendedBrowserActions) == 0 &&
		len(normalized.RecommendedNodeActions) == 0 &&
		!normalized.BrowserOnNode &&
		!normalized.HasActiveNodeRun &&
		!normalized.HasRunningBrowserProfile &&
		!normalized.NeedsSessionSync {
		return nil
	}
	return &normalized
}

func mergeSharedSessionBrowserWorkbenchSections(current []string, next []string) []string {
	if len(next) == 0 {
		if len(current) == 0 {
			return nil
		}
		return append([]string(nil), current...)
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(current)+len(next))
	for _, raw := range current {
		value := strings.ToLower(strings.TrimSpace(raw))
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	for _, raw := range next {
		value := strings.ToLower(strings.TrimSpace(raw))
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func cloneSharedSessionBrowserWorkbenchActions(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, raw := range items {
		value := strings.TrimSpace(raw)
		if value == "" {
			continue
		}
		out = append(out, value)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
