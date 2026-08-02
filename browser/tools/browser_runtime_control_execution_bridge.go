package tools

import (
	"context"
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

func browserRuntimeMutationContext(
	watchManagerProvider agentxbrowserruntime.SharedSessionBrowserObserverManager,
	sessionRegistry *BrowserSessionRegistry,
	stateRegistry agentxbrowserruntime.SharedSessionBrowserStateRegistry,
) agentxbrowserruntime.SharedSessionBrowserMutationContext {
	return agentxbrowserruntime.SharedSessionBrowserMutationContextFor(
		watchManagerProvider,
		sessionRegistry,
		stateRegistry,
		browserRuntimeReconnectWatchdogWindow,
	)
}

type browserRuntimePrepareResult = agentxbrowserruntime.SharedSessionBrowserExecutionResult

func browserRuntimeEnsurePreparedProfile(
	ctx context.Context,
	registry agentxbrowserruntime.SharedSessionBrowserStateRegistry,
	control BrowserRuntimeControlBackend,
	requestedProfile string,
	selectedInfo BrowserRuntimeInfo,
	binding *browserRuntimeSessionBinding,
) (browserRuntimePrepareResult, error) {
	return agentxbrowserruntime.ExecuteSharedSessionBrowserEnsurePrepared(
		ctx,
		control,
		browserRuntimeExecutionRequest(ctx, registry, requestedProfile, selectedInfo, binding, false),
	)
}

func browserRuntimeStartProfile(
	ctx context.Context,
	registry agentxbrowserruntime.SharedSessionBrowserStateRegistry,
	control BrowserRuntimeControlBackend,
	requestedProfile string,
	selectedInfo BrowserRuntimeInfo,
	binding *browserRuntimeSessionBinding,
) (browserRuntimePrepareResult, error) {
	return agentxbrowserruntime.ExecuteSharedSessionBrowserStart(
		ctx,
		control,
		browserRuntimeExecutionRequest(ctx, registry, requestedProfile, selectedInfo, binding, false),
	)
}

func browserRuntimeSharedBindingEvaluationPtr(
	binding *browserRuntimeSessionBinding,
) *agentxbrowserruntime.SharedSessionBrowserBindingEvaluation {
	if binding == nil {
		return nil
	}
	evaluation := browserRuntimeSharedBindingEvaluation(*binding, nil)
	return &evaluation
}

func browserRuntimeSharedBindingEvaluationPtrFromPayload(
	payload browserRuntimePayload,
) *agentxbrowserruntime.SharedSessionBrowserBindingEvaluation {
	if payload.SessionBinding == nil &&
		payload.ProfileStatus == nil &&
		len(payload.Profiles) == 0 &&
		payload.SessionProfileSelection == nil &&
		payload.SessionTargetSelection == nil &&
		len(payload.SessionRoutes) == 0 {
		return nil
	}
	evaluation := browserRuntimeSharedBindingEvaluationForPayload(payload, payload.SessionRoutes)
	if projection := browserRuntimeSharedSelectionProjectionFromPayload(payload); projection != nil {
		projection = agentxbrowserruntime.ProjectSharedSessionBrowserSelectionProjectionFromBindingEvaluation(
			evaluation,
			projection,
		)
		evaluation = agentxbrowserruntime.ProjectSharedSessionBrowserBindingEvaluationWithSelectionProjection(
			evaluation,
			projection,
		)
	}
	return &evaluation
}

func browserRuntimeExecutionRequest(
	ctx context.Context,
	registry agentxbrowserruntime.SharedSessionBrowserStateRegistry,
	requestedProfile string,
	selectedInfo BrowserRuntimeInfo,
	binding *browserRuntimeSessionBinding,
	force bool,
) agentxbrowserruntime.SharedSessionBrowserExecutionRequest {
	sessionID, _ := browserRuntimeSessionStateRegistrySessionID(ctx, registry)
	return agentxbrowserruntime.BuildSharedSessionBrowserExecutionRequestForBindingEvaluation(
		registry,
		sessionID,
		requestedProfile,
		selectedInfo,
		force,
		browserRuntimeSharedBindingEvaluationPtr(binding),
		browserRuntimeReconnectWatchdogWindow,
	)
}

func browserRuntimeExecutionRequestFromPayload(
	ctx context.Context,
	registry agentxbrowserruntime.SharedSessionBrowserStateRegistry,
	requestedProfile string,
	selectedInfo BrowserRuntimeInfo,
	payload browserRuntimePayload,
	force bool,
) agentxbrowserruntime.SharedSessionBrowserExecutionRequest {
	sessionID, _ := browserRuntimeSessionStateRegistrySessionID(ctx, registry)
	return agentxbrowserruntime.BuildSharedSessionBrowserExecutionRequestForBindingEvaluation(
		registry,
		sessionID,
		requestedProfile,
		selectedInfo,
		force,
		browserRuntimeSharedBindingEvaluationPtrFromPayload(payload),
		browserRuntimeReconnectWatchdogWindow,
	)
}

func browserRuntimeClearRequest(
	ctx context.Context,
	sessionRegistry *BrowserSessionRegistry,
	stateRegistry agentxbrowserruntime.SharedSessionBrowserStateRegistry,
	selectedInfo BrowserRuntimeInfo,
	selectedRoute *browserRuntimeRouteDescriptor,
	binding *browserRuntimeSessionBinding,
	force bool,
) agentxbrowserruntime.SharedSessionBrowserClearRequest {
	sessionID, _ := browserRuntimeSessionStateRegistrySessionID(ctx, stateRegistry)
	return agentxbrowserruntime.BuildSharedSessionBrowserClearRequestForBindingEvaluation(
		sessionRegistry,
		stateRegistry,
		sessionID,
		selectedInfo,
		browserRuntimeSessionRouteFilter(selectedRoute),
		force,
		browserRuntimeSharedBindingEvaluationPtr(binding),
		browserRuntimeReconnectWatchdogWindow,
	)
}

func browserRuntimeClearRequestFromPayload(
	ctx context.Context,
	sessionRegistry *BrowserSessionRegistry,
	stateRegistry agentxbrowserruntime.SharedSessionBrowserStateRegistry,
	selectedInfo BrowserRuntimeInfo,
	selectedRoute *browserRuntimeRouteDescriptor,
	payload browserRuntimePayload,
	force bool,
) agentxbrowserruntime.SharedSessionBrowserClearRequest {
	sessionID, _ := browserRuntimeSessionStateRegistrySessionID(ctx, stateRegistry)
	return agentxbrowserruntime.BuildSharedSessionBrowserClearRequestForBindingEvaluation(
		sessionRegistry,
		stateRegistry,
		sessionID,
		selectedInfo,
		browserRuntimeSessionRouteFilter(selectedRoute),
		force,
		browserRuntimeSharedBindingEvaluationPtrFromPayload(payload),
		browserRuntimeReconnectWatchdogWindow,
	)
}

func browserRuntimeRestartProfile(
	ctx context.Context,
	registry agentxbrowserruntime.SharedSessionBrowserStateRegistry,
	control BrowserRuntimeControlBackend,
	requestedProfile string,
	selectedInfo BrowserRuntimeInfo,
	binding *browserRuntimeSessionBinding,
	force bool,
) (browserRuntimePrepareResult, error) {
	return agentxbrowserruntime.ExecuteSharedSessionBrowserRestart(
		ctx,
		control,
		browserRuntimeExecutionRequest(ctx, registry, requestedProfile, selectedInfo, binding, force),
	)
}

func browserRuntimeCreateProfile(
	ctx context.Context,
	manager BrowserRuntimeProfileManagementBackend,
	requestedProfile string,
	browserApp string,
	color string,
	copyFrom string,
	selectedInfo BrowserRuntimeInfo,
) (browserRuntimePrepareResult, error) {
	return agentxbrowserruntime.ExecuteSharedSessionBrowserCreateProfile(
		ctx,
		manager,
		agentxbrowserruntime.SharedSessionBrowserProfileCreateRequest{
			RequestedProfile: strings.TrimSpace(requestedProfile),
			SelectedInfo:     selectedInfo,
			BrowserApp:       strings.TrimSpace(browserApp),
			Color:            strings.TrimSpace(color),
			CopyFrom:         strings.TrimSpace(copyFrom),
		},
	)
}

func browserRuntimeStopProfile(
	ctx context.Context,
	registry agentxbrowserruntime.SharedSessionBrowserStateRegistry,
	control BrowserRuntimeControlBackend,
	requestedProfile string,
	selectedInfo BrowserRuntimeInfo,
	binding *browserRuntimeSessionBinding,
	force bool,
) (browserRuntimePrepareResult, error) {
	return agentxbrowserruntime.ExecuteSharedSessionBrowserStop(
		ctx,
		control,
		browserRuntimeExecutionRequest(ctx, registry, requestedProfile, selectedInfo, binding, force),
	)
}

func browserRuntimeDeleteProfile(
	ctx context.Context,
	registry agentxbrowserruntime.SharedSessionBrowserStateRegistry,
	manager BrowserRuntimeProfileManagementBackend,
	requestedProfile string,
	force bool,
	selectedInfo BrowserRuntimeInfo,
	binding *browserRuntimeSessionBinding,
) (browserRuntimePrepareResult, error) {
	return agentxbrowserruntime.ExecuteSharedSessionBrowserDeleteProfile(
		ctx,
		manager,
		browserRuntimeExecutionRequest(ctx, registry, requestedProfile, selectedInfo, binding, force),
	)
}

func browserRuntimeTeardownProfile(
	ctx context.Context,
	registry agentxbrowserruntime.SharedSessionBrowserStateRegistry,
	control BrowserRuntimeControlBackend,
	requestedProfile string,
	selectedInfo BrowserRuntimeInfo,
	binding *browserRuntimeSessionBinding,
) (browserRuntimePrepareResult, error) {
	return agentxbrowserruntime.ExecuteSharedSessionBrowserTeardown(
		ctx,
		control,
		browserRuntimeExecutionRequest(ctx, registry, requestedProfile, selectedInfo, binding, false),
	)
}

func browserRuntimeApplyPrepareResult(
	ctx context.Context,
	payload *browserRuntimePayload,
	watchManagerProvider agentxbrowserruntime.SharedSessionBrowserObserverManager,
	sessionRegistry *BrowserSessionRegistry,
	stateRegistry agentxbrowserruntime.SharedSessionBrowserStateRegistry,
	selectedInfo BrowserRuntimeInfo,
	result browserRuntimePrepareResult,
) {
	if payload == nil {
		return
	}
	if strings.TrimSpace(result.Profile) != "" {
		payload.PreparedProfile = strings.TrimSpace(result.Profile)
	}
	sessionID := strings.TrimSpace(payload.SessionID)
	if sessionID == "" {
		if resolvedSessionID, ok := browserRuntimeSessionStateRegistrySessionID(ctx, stateRegistry); ok {
			sessionID = resolvedSessionID
		}
	}
	application := agentxbrowserruntime.ApplySharedSessionBrowserExecutionResultWithMutationContext(
		browserRuntimeMutationContext(watchManagerProvider, sessionRegistry, stateRegistry),
		sessionID,
		selectedInfo,
		"",
		result,
	)
	browserRuntimeApplyPrepareResultSurface(
		ctx,
		payload,
		selectedInfo,
		agentxbrowserruntime.BuildSharedSessionBrowserExecutionSurfaceProjection(selectedInfo, result, application),
	)
}
