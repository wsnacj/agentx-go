package tools

import (
	"context"
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

type browserRuntimeInspectionActionOptions struct {
	Action           string
	SelectedInfo     BrowserRuntimeInfo
	EffectiveProfile string
	Capabilities     BrowserCapabilities
}

func browserRuntimeApplyInspectionAction(
	ctx browserRegistrationContext,
	callCtx context.Context,
	payload *browserRuntimePayload,
	watchManager agentxbrowserruntime.SharedSessionBrowserWatchManager,
	options browserRuntimeInspectionActionOptions,
) *agentxbrowserruntime.SharedSessionBrowserBindingEvaluation {
	if payload == nil {
		return nil
	}
	action := browserRuntimeCanonicalAction(options.Action)
	observation := agentxbrowserruntime.ObserveProjectedSharedSessionBrowserInspectionAction(
		callCtx,
		watchManager,
		agentxbrowserruntime.BuildSharedSessionBrowserInspectionActionRequest(agentxbrowserruntime.SharedSessionBrowserInspectionActionInput{
			Action:                   action,
			SessionID:                payload.SessionID,
			SelectedInfo:             options.SelectedInfo,
			BindingRoute:             browserRuntimeSessionRouteFilter(payload.SelectedRoute),
			RequestedProfile:         options.EffectiveProfile,
			ExplicitRequestedProfile: strings.TrimSpace(payload.RequestedProfile) != "",
			ExplicitSessionScope:     strings.TrimSpace(payload.RequestedProfile) != "" || strings.TrimSpace(payload.RequestedRuntimeTarget) != "",
			IncludeStatus:            browserRuntimeActionSupported(options.Capabilities, "status"),
			IncludeProfiles:          browserRuntimeActionSupported(options.Capabilities, "profiles"),
			IncludeSessionView:       browserRuntimeActionSupported(options.Capabilities, "sessions"),
		}),
		ctx.sessionStateRegistry,
	)
	projection := observation.Projection
	bindingEvaluation := &observation.Watch.View.Binding
	switch action {
	case "status":
		browserRuntimeApplyInspectionProjection(callCtx, payload, action, projection, options.SelectedInfo)
	case "workbench":
		browserRuntimeApplyInspectionProjection(callCtx, payload, action, projection, options.SelectedInfo)
	case "profiles":
		browserRuntimeApplyInspectionProjection(callCtx, payload, action, projection, options.SelectedInfo)
	case "sessions":
		browserRuntimeApplyInspectionProjection(callCtx, payload, action, projection, options.SelectedInfo)
	}
	browserRuntimeMaybeApplyInspectionLaunchDiagnostics(ctx.opts.RepairScript, payload, action, projection.RuntimeStatus, options.SelectedInfo)
	return bindingEvaluation
}

func browserRuntimeDispatchInspectionAction(
	ctx browserRegistrationContext,
	callCtx context.Context,
	payload *browserRuntimePayload,
	selectedBackend BrowserBackend,
	watchManager agentxbrowserruntime.SharedSessionBrowserWatchManager,
	options browserRuntimeInspectionActionOptions,
) *agentxbrowserruntime.SharedSessionBrowserBindingEvaluation {
	action := browserRuntimeCanonicalAction(options.Action)
	switch action {
	case "status", "workbench":
		return browserRuntimeApplyInspectionAction(ctx, callCtx, payload, watchManager, options)
	case "profiles":
		if _, ok := selectedBackend.(BrowserRuntimeControlBackend); !ok || !browserRuntimeActionSupported(options.Capabilities, "profiles") {
			browserRuntimeApplyUnsupportedActionOutcome(payload, action)
			return nil
		}
		return browserRuntimeApplyInspectionAction(ctx, callCtx, payload, watchManager, options)
	case "sessions":
		if !browserRuntimeActionSupported(options.Capabilities, "sessions") {
			browserRuntimeApplyUnsupportedActionOutcome(payload, action)
			return nil
		}
		return browserRuntimeApplyInspectionAction(ctx, callCtx, payload, watchManager, options)
	default:
		return nil
	}
}

func browserRuntimeApplyInspectionProjection(
	callCtx context.Context,
	payload *browserRuntimePayload,
	action string,
	projection agentxbrowserruntime.SharedSessionBrowserInspectionProjection,
	selectedInfo BrowserRuntimeInfo,
) {
	browserRuntimeApplySharedActionSurface(
		callCtx,
		payload,
		browserRuntimeInspectionActionSurface(action, projection, selectedInfo),
	)
}

func browserRuntimeMaybeApplyInspectionLaunchDiagnostics(
	repairScript string,
	payload *browserRuntimePayload,
	action string,
	status *agentxbrowserruntime.BrowserProfileStatusResult,
	selectedInfo BrowserRuntimeInfo,
) {
	if payload == nil || status == nil {
		return
	}
	switch browserRuntimeCanonicalAction(action) {
	case "status", "workbench":
	default:
		return
	}
	browserRuntimeApplyLaunchDiagnostics(
		payload,
		action,
		browserRuntimeLaunchDiagnosticsSummaryFromStatusResult(repairScript, BrowserProfileStatusResult(*status), selectedInfo),
	)
}

func browserRuntimeInspectionActionSurface(
	action string,
	projection agentxbrowserruntime.SharedSessionBrowserInspectionProjection,
	selectedInfo BrowserRuntimeInfo,
) agentxbrowserruntime.SharedSessionBrowserActionSurfaceProjection {
	return agentxbrowserruntime.BuildSharedSessionBrowserInspectionSurfaceProjection(
		action,
		selectedInfo,
		projection,
	)
}
