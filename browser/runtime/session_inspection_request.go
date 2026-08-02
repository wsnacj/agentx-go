package browserruntime

import "strings"

// SharedSessionBrowserInspectionActionInput captures the minimal action-scoped
// inspection intent before browserruntime lowers it onto the canonical
// inspection request contract.
type SharedSessionBrowserInspectionActionInput struct {
	Action                   string
	SessionID                string
	SelectedInfo             BrowserRuntimeInfo
	BindingRoute             BrowserSessionRoute
	RequestedProfile         string
	ExplicitRequestedProfile bool
	ExplicitSessionScope     bool
	IncludeStatus            bool
	IncludeProfiles          bool
	IncludeSessionView       bool
}

// SharedSessionBrowserInspectionActionRequest captures the action-scoped
// inspection inputs that runtime status/workbench/profiles/sessions surfaces
// share before they are lowered onto the observer/watch lifecycle contract.
type SharedSessionBrowserInspectionActionRequest struct {
	Action                   string
	SessionID                string
	SelectedInfo             BrowserRuntimeInfo
	BindingRoute             BrowserSessionRoute
	RequestedProfile         string
	ExplicitRequestedProfile bool
	ExplicitSessionScope     bool
	IncludeStatus            bool
	IncludeProfiles          bool
	IncludeSessionView       bool
}

// BuildSharedSessionBrowserInspectionActionRequest lowers the minimal
// action-scoped inspection intent onto the canonical shared request contract.
func BuildSharedSessionBrowserInspectionActionRequest(input SharedSessionBrowserInspectionActionInput) SharedSessionBrowserInspectionActionRequest {
	return normalizeSharedSessionBrowserInspectionActionRequest(SharedSessionBrowserInspectionActionRequest{
		Action:                   input.Action,
		SessionID:                input.SessionID,
		SelectedInfo:             input.SelectedInfo,
		BindingRoute:             input.BindingRoute,
		RequestedProfile:         input.RequestedProfile,
		ExplicitRequestedProfile: input.ExplicitRequestedProfile,
		ExplicitSessionScope:     input.ExplicitSessionScope,
		IncludeStatus:            input.IncludeStatus,
		IncludeProfiles:          input.IncludeProfiles,
		IncludeSessionView:       input.IncludeSessionView,
	})
}

// BuildSharedSessionBrowserInspectionObserverRequest lowers the action-scoped
// inspection intent onto the shared observer/watch request contract owned by
// browserruntime.
func BuildSharedSessionBrowserInspectionObserverRequest(req SharedSessionBrowserInspectionActionRequest) SharedSessionBrowserObserverRequest {
	req = normalizeSharedSessionBrowserInspectionActionRequest(req)
	action := req.Action
	observerReq := SharedSessionBrowserObserverRequest{
		SessionID:        req.SessionID,
		SelectedInfo:     req.SelectedInfo,
		BindingRoute:     req.BindingRoute,
		RequestedProfile: req.RequestedProfile,
	}
	switch action {
	case "status":
		observerReq.IncludeStatus = req.IncludeStatus
		observerReq.IncludeProfiles = req.IncludeProfiles
	case "workbench":
		observerReq.IncludeStatus = req.IncludeStatus
		observerReq.IncludeProfiles = req.IncludeProfiles
		observerReq.IncludeSessionView = req.IncludeSessionView
		if observerReq.IncludeSessionView {
			observerReq.SessionViewInfo,
				observerReq.SessionViewRouteFilter,
				observerReq.SessionViewRequestedProfile = sharedSessionBrowserInspectionSessionViewScope(req)
		}
	case "profiles":
		if !req.ExplicitRequestedProfile {
			observerReq.BindingRoute.Profile = ""
		}
		observerReq.RequestedProfile = firstNonEmptyBindingString(
			observerReq.RequestedProfile,
			observerReq.SelectedInfo.Profile,
		)
		observerReq.IncludeProfiles = true
	case "sessions":
		observerReq.IncludeSessionView = true
		observerReq.SessionViewInfo,
			observerReq.SessionViewRouteFilter,
			observerReq.SessionViewRequestedProfile = sharedSessionBrowserInspectionSessionViewScope(req)
	}
	return observerReq
}

func normalizeSharedSessionBrowserInspectionActionRequest(req SharedSessionBrowserInspectionActionRequest) SharedSessionBrowserInspectionActionRequest {
	req.Action = strings.ToLower(strings.TrimSpace(req.Action))
	req.SessionID = strings.TrimSpace(req.SessionID)
	req.SelectedInfo = normalizeSharedSessionBrowserInspectionRuntimeInfo(req.SelectedInfo)
	req.BindingRoute = normalizeBrowserSessionRoute(req.BindingRoute)
	req.RequestedProfile = strings.TrimSpace(req.RequestedProfile)
	return req
}

func sharedSessionBrowserInspectionSessionViewScope(req SharedSessionBrowserInspectionActionRequest) (BrowserRuntimeInfo, BrowserSessionRoute, string) {
	sessionViewInfo := normalizeSharedSessionBrowserInspectionRuntimeInfo(req.SelectedInfo)
	sessionViewRouteFilter := normalizeBrowserSessionRoute(req.BindingRoute)
	sessionViewRequestedProfile := strings.TrimSpace(req.RequestedProfile)
	if !req.ExplicitSessionScope {
		sessionViewInfo = BrowserRuntimeInfo{}
		sessionViewRouteFilter = BrowserSessionRoute{}
		sessionViewRequestedProfile = ""
	}
	return sessionViewInfo, sessionViewRouteFilter, sessionViewRequestedProfile
}

func normalizeSharedSessionBrowserInspectionRuntimeInfo(info BrowserRuntimeInfo) BrowserRuntimeInfo {
	info.Backend = strings.TrimSpace(info.Backend)
	info.Profile = strings.TrimSpace(info.Profile)
	info.Target = strings.TrimSpace(info.Target)
	return info
}
