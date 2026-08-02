package browserruntime

import (
	"context"
	"strings"
)

// SharedSessionBrowserSelectionActionDispatchRequest carries the runtime-owned
// inputs needed to dispatch select_profile and select_target through one shared
// browserruntime helper.
type SharedSessionBrowserSelectionActionDispatchRequest struct {
	Action               string
	MutationContext      SharedSessionBrowserMutationContext
	SessionID            string
	SelectedInfo         BrowserRuntimeInfo
	Route                BrowserSessionRoute
	BrowserApp           string
	Control              BrowserRuntimeControlBackend
	ValidateWithProfiles bool
	Source               string
	TargetRequest        *SharedSessionBrowserSelectTargetRequest
}

// SharedSessionBrowserSelectionActionDispatchResult captures the shared
// selection outcome together with whether the helper claimed the action.
type SharedSessionBrowserSelectionActionDispatchResult struct {
	Handled          bool
	Decision         string
	Ready            bool
	Note             string
	ProfileSelection *SharedSessionBrowserProfileSelection
	TargetSelection  *BrowserSessionTargetSelection
	Err              error
}

// SharedSessionBrowserRememberProfileDispatchRequest carries the runtime-owned
// inputs needed to route remember_profile follow-up through one shared helper.
type SharedSessionBrowserRememberProfileDispatchRequest struct {
	MutationContext     SharedSessionBrowserMutationContext
	SessionID           string
	SelectedInfo        BrowserRuntimeInfo
	Route               BrowserSessionRoute
	ProfileStatus       *BrowserProfileStatusResult
	PreparedProfile     string
	RequestedProfile    string
	RequestedBrowserApp string
}

// SharedSessionBrowserRememberTargetDispatchRequest carries the runtime-owned
// inputs needed to route remember_target follow-up through one shared helper.
type SharedSessionBrowserRememberTargetDispatchRequest struct {
	MutationContext SharedSessionBrowserMutationContext
	SessionID       string
	Route           BrowserSessionRoute
	TargetID        string
	TabIndex        int
	Source          string
}

// SharedSessionBrowserRememberTargetDispatchResult captures the remembered
// route-scoped target together with any promoted profile selection.
type SharedSessionBrowserRememberTargetDispatchResult struct {
	Selection        *BrowserSessionTargetSelection
	ProfileSelection *SharedSessionBrowserProfileSelection
	Decision         string
	Ready            bool
}

// DispatchSharedSessionBrowserSelectionAction routes select_profile and
// select_target through the shared browserruntime owner so tools callers only
// keep tool-argument parsing and payload/result marshaling.
func DispatchSharedSessionBrowserSelectionAction(
	ctx context.Context,
	req SharedSessionBrowserSelectionActionDispatchRequest,
) SharedSessionBrowserSelectionActionDispatchResult {
	sessionID := strings.TrimSpace(req.SessionID)
	route := normalizeBrowserSessionRoute(req.Route)
	source := firstNonEmptyString(strings.TrimSpace(req.Source), "select_profile")
	switch strings.ToLower(strings.TrimSpace(req.Action)) {
	case "select_profile":
		selection, decision, ok, err := SelectSharedSessionBrowserProfileWithContext(
			req.MutationContext,
			ctx,
			sessionID,
			req.SelectedInfo,
			req.BrowserApp,
			req.Control,
			req.ValidateWithProfiles,
			source,
		)
		result := SharedSessionBrowserSelectionActionDispatchResult{
			Handled:  true,
			Decision: decision,
			Err:      err,
		}
		if err != nil || !ok {
			return result
		}
		result.ProfileSelection = &selection
		result.Ready = true
		if targetSelection, _, err := SyncOrClearSharedSessionBrowserCurrentTargetForProfileSelectionWithContext(
			req.MutationContext,
			sessionID,
			route,
			&selection,
			source,
		); err == nil {
			result.TargetSelection = targetSelection
		}
		return result
	case "select_target":
		result := SharedSessionBrowserSelectionActionDispatchResult{Handled: true}
		if req.TargetRequest == nil {
			return result
		}
		targetReq := *req.TargetRequest
		if strings.TrimSpace(targetReq.SessionID) == "" {
			targetReq.SessionID = sessionID
		}
		if sharedSessionBrowserRouteEmpty(targetReq.Route) {
			targetReq.Route = route
		}
		targetReq.Source = firstNonEmptyString(strings.TrimSpace(targetReq.Source), "select_target")
		target, err := SelectSharedSessionBrowserTargetWithContext(req.MutationContext, targetReq)
		result.Decision = target.Decision
		result.Note = target.Note
		result.Ready = target.Ready
		result.TargetSelection = target.Selection
		result.Err = err
		if err != nil || target.Selection == nil {
			return result
		}
		if promoted, ok := PromoteSharedSessionBrowserProfileFromTargetSelection(
			req.MutationContext.StateRegistry,
			sessionID,
			target.Selection,
		); ok {
			result.ProfileSelection = &promoted
		}
		return result
	default:
		return SharedSessionBrowserSelectionActionDispatchResult{}
	}
}

// DispatchSharedSessionBrowserRememberProfile routes remember_profile follow-up
// through the shared browserruntime owner so tools callers only keep payload
// projection and gating.
func DispatchSharedSessionBrowserRememberProfile(
	req SharedSessionBrowserRememberProfileDispatchRequest,
) SharedSessionBrowserRememberProfileResult {
	return RememberSharedSessionBrowserProfileForRouteWithContext(
		req.MutationContext,
		req.SessionID,
		req.SelectedInfo,
		req.Route,
		req.ProfileStatus,
		req.PreparedProfile,
		req.RequestedProfile,
		req.RequestedBrowserApp,
	)
}

// DispatchSharedSessionBrowserRememberTarget routes remember_target follow-up
// through the shared browserruntime owner so tools callers only keep popup
// review policy and payload/result marshaling.
func DispatchSharedSessionBrowserRememberTarget(
	req SharedSessionBrowserRememberTargetDispatchRequest,
) SharedSessionBrowserRememberTargetDispatchResult {
	result := RememberSharedSessionBrowserTargetWithContext(
		req.MutationContext,
		SharedSessionBrowserRememberTargetRequest{
			SessionID: strings.TrimSpace(req.SessionID),
			Route:     req.Route,
			TargetID:  strings.TrimSpace(req.TargetID),
			TabIndex:  req.TabIndex,
			Source:    firstNonEmptyString(strings.TrimSpace(req.Source), "remember_target"),
		},
	)
	dispatched := SharedSessionBrowserRememberTargetDispatchResult{
		Selection: result.Selection,
		Decision:  result.Decision,
		Ready:     result.Ready,
	}
	if promoted, ok := PromoteSharedSessionBrowserProfileFromTargetSelection(
		req.MutationContext.StateRegistry,
		strings.TrimSpace(req.SessionID),
		result.Selection,
	); ok {
		dispatched.ProfileSelection = &promoted
	}
	return dispatched
}
