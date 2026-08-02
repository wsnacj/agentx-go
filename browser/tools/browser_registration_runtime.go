package tools

import (
	"context"
	"fmt"
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
	types "github.com/wsnacj/agentx-go/components/llm"
	agentxmedia "github.com/wsnacj/agentx-go/runtime/mediaartifact"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

type browserRegistrationContext struct {
	reg                         *llmxtools.Registry
	opts                        BrowserToolOptions
	enabledTools                map[string]bool
	capabilities                BrowserCapabilities
	runtimeActions              []string
	managedOptInProjection      browserManagedOptInProjection
	managedOptInProjectionReady bool
	derivedCache                *browserRegistrationDerivedCache
	substrateAssessment         browserDefaultSubstrateAssessment
	substrateSummary            BrowserWorkbenchSubstrateSummary
	sessionRegistry             *BrowserSessionRegistry
	sessionRunRegistry          BrowserSessionRunRegistry
	sessionStateRegistry        agentxbrowserruntime.SharedSessionBrowserStateRegistry
	watchManagerProvider        agentxbrowserruntime.SharedSessionBrowserObserverManager
	policy                      outboundNetworkPolicy
	backend                     BrowserBackend
	timeoutMs                   int
	openWaitMs                  int
	screenshotWaitMs            int
	maxChars                    int
}

func browserRegistrationCompatToolNameForActKind(kind string) (string, error) {
	name := browserCompatToolForManagedOptInActKind(kind)
	if name == "" {
		return "", fmt.Errorf("browser compat registration: missing tool name for act kind %q", browserNormalizeToolToken(kind))
	}
	return name, nil
}

func browserRegistrationCompatToolErrorf(toolName string, format string, args ...any) error {
	prefix := NormalizeToolName(toolName)
	if prefix == "" {
		prefix = "browser_compat"
	}
	return fmt.Errorf(prefix+": "+format, args...)
}

func registerBrowserOpenTool(ctx browserRegistrationContext) {
	ctx.reg.Register(browserOpenDefinition(), func(callCtx context.Context, call types.FunctionCall) (string, error) {
		toolName, err := browserRegistrationCompatToolNameForActKind("open")
		if err != nil {
			return "", err
		}
		effectivePolicy, err := resolveToolRuntimeSharedFetchPolicy(callCtx, ctx.policy)
		if err != nil {
			return "", browserRegistrationCompatToolErrorf(toolName, "invalid runtime network policy: %w", err)
		}
		params, err := decodeArgs(call.Arguments)
		if err != nil {
			return "", err
		}
		rawURL := firstString(params, "url")
		if rawURL == "" {
			return "", browserRegistrationCompatToolErrorf(toolName, "url is required")
		}
		parsed, err := effectivePolicy.validateURL(callCtx, rawURL)
		if err != nil {
			return "", browserRegistrationCompatToolErrorf(toolName, "%w", err)
		}
		lane, dispatch, err := resolveBrowserExecutionLaneForRegistrationDirectURLAction(ctx, callCtx, params, toolName, parsed.String())
		if err != nil {
			return "", browserRegistrationCompatToolErrorf(toolName, "%w", err)
		}
		routedBackend := lane.Backend
		runtimeInfo := lane.Runtime
		hiddenImplicitHostDefaultBase := dispatch.HiddenImplicitHostDefaultBase
		identity := browserRegistrationCompatRuntimeIdentity(ctx, lane.Capabilities, toolName, "")
		if err := identity.legacyHostFallbackError(hiddenImplicitHostDefaultBase, dispatch.ExplicitRuntimeTarget, runtimeInfo, browserToolTarget{}, parsed.String()); err != nil {
			return "", browserRegistrationCompatToolErrorf(toolName, "%w", err)
		}
		browserApp := dispatch.BrowserApp
		waitMs := firstInt(params, "wait_ms")
		if waitMs <= 0 {
			waitMs = ctx.openWaitMs
		}
		req := BrowserOpenRequest{
			URL:        parsed.String(),
			BrowserApp: browserApp,
			WaitMs:     waitMs,
		}
		result, err := routedBackend.Open(callCtx, req)
		if err != nil {
			return "", err
		}
		result = browserAnnotateURLFallbackResult(result, dispatch.RouteFallback)
		targetID := agentxbrowserruntime.ApplySharedSessionBrowserOpenResultWithContext(
			browserSharedMutationContext(ctx.watchManagerProvider, ctx.sessionRegistry),
			agentxbrowserruntime.SharedSessionBrowserOpenResultEventRequest{
				SessionID: ToolSessionIDFromContext(callCtx),
				Route: browserSessionRoute(
					runtimeInfo,
					firstNonEmpty(strings.TrimSpace(result.BrowserApp), req.BrowserApp),
					strings.TrimSpace(result.Backend),
				),
				URL:    req.URL,
				Source: identity.EventSource,
			},
		).TargetID
		return marshalBrowserOpenPayload(browserOpenToolPayload{
			browserRegistrationRouteCapabilityPayload: browserRegistrationRouteCapabilityPayload{
				CapabilityMetadata: identity.CapabilityMetadata,
			},
			URL:                 req.URL,
			Backend:             strings.TrimSpace(result.Backend),
			BrowserApp:          firstNonEmpty(strings.TrimSpace(result.BrowserApp), req.BrowserApp),
			Profile:             runtimeInfo.Profile,
			RuntimeTarget:       runtimeInfo.Target,
			BrowserSurface:      identity.BrowserSurface,
			BrowserOptInTargets: append([]string(nil), identity.BrowserOptInTargets...),
			Target:              browserOpenCurrentTargetValue(targetID),
			TargetID:            targetID,
			Status:              firstNonEmpty(strings.TrimSpace(result.Status), "opened"),
			WaitMs:              req.WaitMs,
			Note:                strings.TrimSpace(result.Note),
		})
	})
}

func registerBrowserNavigateTool(ctx browserRegistrationContext) {
	ctx.reg.Register(browserNavigateDefinition(), func(callCtx context.Context, call types.FunctionCall) (string, error) {
		toolName, err := browserRegistrationCompatToolNameForActKind("navigate")
		if err != nil {
			return "", err
		}
		effectivePolicy, err := resolveToolRuntimeSharedFetchPolicy(callCtx, ctx.policy)
		if err != nil {
			return "", browserRegistrationCompatToolErrorf(toolName, "invalid runtime network policy: %w", err)
		}
		params, err := decodeArgs(call.Arguments)
		if err != nil {
			return "", err
		}
		rawURL := firstString(params, "url")
		if rawURL == "" {
			return "", browserRegistrationCompatToolErrorf(toolName, "url is required")
		}
		parsed, err := effectivePolicy.validateURL(callCtx, rawURL)
		if err != nil {
			return "", browserRegistrationCompatToolErrorf(toolName, "%w", err)
		}
		var lane BrowserExecutionLane
		var dispatch browserRegistrationPageActionDispatch
		if strings.TrimSpace(firstString(params, "target")) == "" {
			lane, dispatch, err = resolveBrowserExecutionLaneForRegistrationDirectURLAction(ctx, callCtx, params, toolName, parsed.String())
		} else {
			lane, dispatch, err = resolveBrowserExecutionLaneForRegistrationURLAction(ctx, callCtx, params, browserRegistrationPageActionDispatchOptions{})
		}
		if err != nil {
			return "", browserRegistrationCompatToolErrorf(toolName, "%w", err)
		}
		routedBackend := lane.Backend
		runtimeInfo := lane.Runtime
		hiddenImplicitHostDefaultBase := dispatch.HiddenImplicitHostDefaultBase
		target := dispatch.Target
		browserApp := dispatch.BrowserApp
		identity := browserRegistrationCompatRuntimeIdentity(ctx, lane.Capabilities, toolName, "")
		if err := identity.legacyHostFallbackError(hiddenImplicitHostDefaultBase, dispatch.ExplicitRuntimeTarget, runtimeInfo, target, parsed.String()); err != nil {
			return "", browserRegistrationCompatToolErrorf(toolName, "%w", err)
		}
		waitMs := firstInt(params, "wait_ms")
		if waitMs <= 0 {
			waitMs = ctx.openWaitMs
		}
		force := firstBool(params, "force")
		req := BrowserNavigateRequest{
			URL:              parsed.String(),
			BrowserApp:       browserApp,
			WaitMs:           waitMs,
			TabIndex:         target.TabIndex,
			Force:            force,
			ExplicitTargetID: strings.TrimSpace(target.TargetID),
			PriorSelection:   browserCurrentTargetSelectionSnapshotForRoute(callCtx, ctx.sessionRegistry, runtimeInfo, browserApp, strings.TrimSpace(runtimeInfo.Backend)),
		}
		result, err := routedBackend.Navigate(callCtx, req)
		if err != nil {
			return "", err
		}
		result = browserAnnotateURLFallbackResult(result, dispatch.RouteFallback)
		finalURL := firstNonEmpty(strings.TrimSpace(result.FinalURL), req.URL)
		finalURL, err = browserValidateFinalURL(callCtx, effectivePolicy, finalURL)
		if err != nil {
			return "", browserRegistrationCompatToolErrorf(toolName, "%w", err)
		}
		resolvedBrowserApp := firstNonEmpty(strings.TrimSpace(result.BrowserApp), req.BrowserApp)
		navigationResult := agentxbrowserruntime.ApplySharedSessionBrowserNavigationResultWithContext(
			browserSharedMutationContext(ctx.watchManagerProvider, ctx.sessionRegistry),
			agentxbrowserruntime.SharedSessionBrowserNavigationResultEventRequest{
				SessionID:        ToolSessionIDFromContext(callCtx),
				Route:            browserSessionRoute(runtimeInfo, resolvedBrowserApp, strings.TrimSpace(result.Backend)),
				ExplicitTargetID: strings.TrimSpace(target.TargetID),
				TabIndex:         req.TabIndex,
				RequestedURL:     req.URL,
				FinalURL:         finalURL,
				Title:            strings.TrimSpace(result.Title),
				Source:           identity.EventSource,
				Force:            force,
				PriorSelection:   browserCurrentTargetSelectionSnapshotForRoute(callCtx, ctx.sessionRegistry, runtimeInfo, resolvedBrowserApp, strings.TrimSpace(result.Backend)),
				Note:             strings.TrimSpace(result.Note),
			},
		)
		targetID := strings.TrimSpace(navigationResult.TargetID)
		status := browserNavigateStatus(result, navigationResult)
		note := strings.TrimSpace(navigationResult.Note)
		return marshalBrowserNavigatePayload(browserNavigateToolPayload{
			browserRegistrationRouteCapabilityPayload: browserRegistrationRouteCapabilityPayload{
				CapabilityMetadata: identity.CapabilityMetadata,
			},
			URL:                    req.URL,
			FinalURL:               finalURL,
			Backend:                strings.TrimSpace(result.Backend),
			BrowserApp:             resolvedBrowserApp,
			Profile:                runtimeInfo.Profile,
			RuntimeTarget:          runtimeInfo.Target,
			BrowserSurface:         identity.BrowserSurface,
			BrowserOptInTargets:    append([]string(nil), identity.BrowserOptInTargets...),
			Title:                  strings.TrimSpace(result.Title),
			Target:                 target.Value,
			TargetID:               targetID,
			Status:                 status,
			Force:                  force,
			ReviewDecision:         navigationResult.ReviewDecision,
			ReviewReady:            navigationResult.ReviewReady,
			PostNavigationSnapshot: browserPostNavigationSnapshotRecommendationForNavigate(status, finalURL),
			BotDetectionWarning:    browserBotDetectionWarningFromNavigation(finalURL, result.Title, note),
			TabIndex:               req.TabIndex,
			WaitMs:                 req.WaitMs,
			Note:                   note,
		})
	})
}

func registerBrowserTabsTool(ctx browserRegistrationContext) {
	ctx.reg.Register(browserTabsDefinition(), func(callCtx context.Context, call types.FunctionCall) (string, error) {
		toolName, err := browserRegistrationCompatToolNameForActKind("list_tabs")
		if err != nil {
			return "", err
		}
		params, err := decodeArgs(call.Arguments)
		if err != nil {
			return "", err
		}
		lane, dispatch, err := resolveBrowserExecutionLaneForRegistrationDispatch(ctx, callCtx, params, browserRegistrationPageActionDispatchOptions{})
		if err != nil {
			return "", browserRegistrationCompatToolErrorf(toolName, "%w", err)
		}
		routedBackend := lane.Backend
		runtimeInfo := lane.Runtime
		hiddenImplicitHostDefaultBase := dispatch.HiddenImplicitHostDefaultBase
		browserApp := dispatch.BrowserApp
		target := dispatch.Target
		action := strings.ToLower(strings.TrimSpace(firstString(params, "action")))
		if action == "" {
			action = "list"
		}
		switch action {
		case "list", "focus", "close":
		default:
			return "", browserRegistrationCompatToolErrorf(toolName, "action must be one of list, focus, close")
		}
		force := firstBool(params, "force")
		rememberTarget := firstBool(params, "remember_target", "remember")
		identity := browserRegistrationCompatRuntimeIdentity(ctx, lane.Capabilities, toolName, action)
		tabsActor := identity.Actor
		explicitTargetID := strings.TrimSpace(target.TargetID)
		priorSelection := browserCurrentTargetSelectionSnapshotForRoute(callCtx, ctx.sessionRegistry, runtimeInfo, browserApp, "")
		priorRequestedTargetID := browserTargetIDForTab(callCtx, ctx.sessionRegistry, runtimeInfo, browserApp, target.TabIndex)
		req := BrowserTabsRequest{
			BrowserApp:             browserApp,
			Action:                 action,
			TabIndex:               target.TabIndex,
			WaitMs:                 firstInt(params, "wait_ms"),
			Force:                  force,
			RememberTarget:         rememberTarget,
			Actor:                  tabsActor,
			ExplicitTargetID:       explicitTargetID,
			PriorSelection:         priorSelection,
			PriorRequestedTargetID: priorRequestedTargetID,
		}
		if req.Action != "list" && req.TabIndex <= 0 {
			return "", browserRegistrationCompatToolErrorf(toolName, "target or tab_index is required for action %s", req.Action)
		}
		if err := identity.legacyHostFallbackError(hiddenImplicitHostDefaultBase, false, runtimeInfo, target, ""); err != nil {
			return "", err
		}
		focusReview := browserPendingTargetReviewState{}
		if req.Action == "focus" {
			focusReview = browserPendingTargetReviewStateForToolTarget(
				callCtx,
				ctx.sessionRegistry,
				browserSessionRoute(runtimeInfo, browserApp, ""),
				target,
			)
			req.Review = focusReview
			if focusReview.Review != nil && !force {
				return marshalBrowserTabsReviewBlockedPayloadWithRouteSurface(
					runtimeInfo,
					browserApp,
					req.Action,
					force,
					target,
					req.WaitMs,
					focusReview,
					identity.CapabilityMetadata,
					identity.BrowserSurface,
					identity.BrowserOptInTargets,
					identity.reviewReason(focusReview, force),
				)
			}
		}
		result, err := routedBackend.Tabs(callCtx, req)
		if err != nil {
			return "", err
		}
		browserApp = firstNonEmpty(strings.TrimSpace(result.BrowserApp), req.BrowserApp)
		backendName := strings.TrimSpace(result.Backend)
		tabsResult := agentxbrowserruntime.ApplySharedSessionBrowserTabsResultWithContext(
			browserSharedMutationContext(ctx.watchManagerProvider, ctx.sessionRegistry),
			agentxbrowserruntime.SharedSessionBrowserTabsResultEventRequest{
				SessionID:              ToolSessionIDFromContext(callCtx),
				Route:                  browserSessionRoute(runtimeInfo, browserApp, backendName),
				Action:                 req.Action,
				RequestedTabIndex:      req.TabIndex,
				ActiveIndex:            result.ActiveIndex,
				Tabs:                   result.Tabs,
				ExplicitTargetID:       explicitTargetID,
				PriorSelection:         priorSelection,
				PriorRequestedTargetID: priorRequestedTargetID,
				Force:                  force,
				RememberTarget:         rememberTarget,
				Review:                 focusReview,
				Actor:                  tabsActor,
				Note:                   strings.TrimSpace(result.Note),
			},
		)
		targetID := strings.TrimSpace(tabsResult.TargetID)
		trackedTabs := tabsResult.Tabs
		rememberReview := tabsResult.RememberReview
		var sessionTargetSelection *agentxbrowserruntime.BrowserSessionTargetSelection
		var sessionProfileSelection *browserRuntimeSessionProfileSelection
		rememberDecision := strings.TrimSpace(rememberReview.Decision)
		rememberReady := rememberReview.Ready
		if rememberTarget {
			if !browserRememberDecisionRequiresPopupReview(rememberDecision, rememberReady) {
				rememberResult := applyBrowserRememberTargetSelection(callCtx, ctx, browserRememberTargetApplyOptions{
					Route:                    browserSessionRoute(runtimeInfo, browserApp, backendName),
					TargetID:                 rememberReview.RememberTargetID,
					TabIndex:                 rememberReview.RememberTabIndex,
					ExistingDecision:         rememberReview.Decision,
					ExistingReady:            rememberReview.Ready,
					PreserveExistingOnSelect: true,
				})
				sessionTargetSelection = rememberResult.Selection
				rememberDecision = rememberResult.Decision
				rememberReady = rememberResult.Ready
				sessionProfileSelection = rememberResult.ProfileSelection
			}
		}
		return marshalBrowserTabsPayload(browserTabsToolPayload{
			browserRegistrationRouteCapabilityPayload: browserRegistrationRouteCapabilityPayload{
				CapabilityMetadata: identity.CapabilityMetadata,
			},
			Backend:                 backendName,
			BrowserApp:              browserApp,
			Profile:                 runtimeInfo.Profile,
			RuntimeTarget:           runtimeInfo.Target,
			BrowserSurface:          identity.BrowserSurface,
			BrowserOptInTargets:     append([]string(nil), identity.BrowserOptInTargets...),
			Action:                  firstNonEmpty(strings.TrimSpace(result.Action), req.Action),
			Status:                  firstNonEmpty(strings.TrimSpace(result.Status), "ok"),
			Force:                   force,
			ReviewDecision:          tabsResult.ReviewDecision,
			ReviewReady:             tabsResult.ReviewReady,
			Target:                  target.Value,
			TargetID:                targetID,
			TabIndex:                req.TabIndex,
			ActiveIndex:             result.ActiveIndex,
			Tabs:                    trackedTabs,
			RememberDecision:        strings.TrimSpace(rememberDecision),
			RememberReady:           rememberReady,
			SessionProfileSelection: sessionProfileSelection,
			SessionTargetSelection:  sessionTargetSelection,
			WaitMs:                  req.WaitMs,
			Note:                    tabsResult.Note,
		})
	})
}

func registerBrowserExtractTool(ctx browserRegistrationContext) {
	ctx.reg.Register(browserExtractDefinition(), func(callCtx context.Context, call types.FunctionCall) (string, error) {
		toolName, err := browserRegistrationCompatToolNameForActKind("extract")
		if err != nil {
			return "", err
		}
		effectivePolicy, err := resolveToolRuntimeSharedFetchPolicy(callCtx, ctx.policy)
		if err != nil {
			return "", browserRegistrationCompatToolErrorf(toolName, "invalid runtime network policy: %w", err)
		}
		params, err := decodeArgs(call.Arguments)
		if err != nil {
			return "", err
		}
		lane, dispatch, err := resolveBrowserExecutionLaneForRegistrationPageAction(ctx, callCtx, params, browserRegistrationPageActionDispatchOptions{})
		if err != nil {
			return "", browserRegistrationCompatToolErrorf(toolName, "%w", err)
		}
		routedBackend := lane.Backend
		runtimeInfo := lane.Runtime
		hiddenImplicitHostDefaultBase := dispatch.HiddenImplicitHostDefaultBase
		explicitRuntimeTarget := dispatch.ExplicitRuntimeTarget
		target := dispatch.Target
		browserApp := dispatch.BrowserApp
		identity := browserRegistrationCompatRuntimeIdentity(ctx, lane.Capabilities, toolName, "")
		force := firstBool(params, "force")
		extractReview := browserPendingTargetReviewStateForRuntimeTarget(callCtx, ctx.sessionRegistry, runtimeInfo, browserApp, target)
		extractReviewTargetID := strings.TrimSpace(target.TargetID)
		if extractReview.Review != nil {
			extractReviewTargetID = firstNonEmpty(extractReviewTargetID, strings.TrimSpace(extractReview.Review.ID))
		}
		rawURL := firstString(params, "url")
		validatedURL := ""
		if rawURL != "" {
			parsed, err := effectivePolicy.validateURL(callCtx, rawURL)
			if err != nil {
				return "", browserRegistrationCompatToolErrorf(toolName, "%w", err)
			}
			validatedURL = parsed.String()
		}
		if err := identity.legacyHostFallbackError(hiddenImplicitHostDefaultBase, explicitRuntimeTarget, runtimeInfo, target, validatedURL); err != nil {
			return "", err
		}
		if validatedURL == "" {
			extractReview = browserAutoFollowPendingTargetReviewStateForRuntimeTarget(callCtx, ctx.sessionRegistry, runtimeInfo, hiddenImplicitHostDefaultBase, browserApp, target)
			if extractReview.Review != nil {
				extractReviewTargetID = firstNonEmpty(strings.TrimSpace(target.TargetID), strings.TrimSpace(extractReview.Review.ID))
			}
		}
		if extractReview.Review != nil && !force {
			return marshalBrowserPageActionReviewBlockedPayload(browserPageActionReviewBlockedPayload{
				browserRegistrationRouteCapabilityPayload: browserRegistrationRouteCapabilityPayload{
					CapabilityMetadata: identity.CapabilityMetadata,
				},
				Backend:             strings.TrimSpace(runtimeInfo.Backend),
				BrowserApp:          strings.TrimSpace(browserApp),
				Profile:             runtimeInfo.Profile,
				RuntimeTarget:       runtimeInfo.Target,
				BrowserSurface:      identity.BrowserSurface,
				BrowserOptInTargets: append([]string(nil), identity.BrowserOptInTargets...),
				Status:              "review_required",
				Force:               force,
				ReviewDecision:      browserPendingTargetReviewDecisionWithState(extractReview, force),
				ReviewReady:         false,
				Target:              target.Value,
				TargetID:            extractReviewTargetID,
				TabIndex:            target.TabIndex,
				Note:                identity.reviewReason(extractReview, force),
			})
		}
		if validatedURL == "" {
			if trackedURL := browserResolvedTargetURL(callCtx, ctx.sessionRegistry, runtimeInfo, hiddenImplicitHostDefaultBase, browserApp, target); trackedURL != "" {
				parsed, err := effectivePolicy.validateURL(callCtx, trackedURL)
				if err != nil {
					return "", browserRegistrationCompatToolErrorf(toolName, "%w", err)
				}
				validatedURL = parsed.String()
			}
		}
		waitMs := firstInt(params, "wait_ms")
		if waitMs <= 0 {
			if validatedURL != "" {
				waitMs = ctx.openWaitMs
			} else {
				waitMs = browserTabTargetWaitMs
			}
		}
		requestMaxChars := firstInt(params, "max_chars")
		if requestMaxChars <= 0 || requestMaxChars > ctx.maxChars {
			requestMaxChars = ctx.maxChars
		}
		req := BrowserExtractRequest{
			URL:               validatedURL,
			BrowserApp:        browserApp,
			WaitMs:            waitMs,
			MaxChars:          requestMaxChars,
			TabIndex:          target.TabIndex,
			PreferredTargetID: extractReviewTargetID,
			Actor:             identity.Actor,
			Force:             force,
			Review:            extractReview,
		}
		result, err := routedBackend.Extract(callCtx, req)
		if err != nil {
			return "", err
		}
		content := strings.TrimSpace(result.Content)
		truncated := false
		if trimmed, changed := trimToMaxChars(content, requestMaxChars); changed {
			content = trimmed
			truncated = true
		}
		finalURL := firstNonEmpty(strings.TrimSpace(result.FinalURL), req.URL)
		resolvedBrowserApp := firstNonEmpty(strings.TrimSpace(result.BrowserApp), req.BrowserApp)
		pageActionResult := agentxbrowserruntime.ApplySharedSessionBrowserPageActionResultWithContext(
			browserSharedMutationContext(ctx.watchManagerProvider, ctx.sessionRegistry),
			agentxbrowserruntime.SharedSessionBrowserPageActionResultEventRequest{
				SessionID:         ToolSessionIDFromContext(callCtx),
				Route:             browserSessionRoute(runtimeInfo, resolvedBrowserApp, runtimeInfo.Backend),
				PreferredTargetID: extractReviewTargetID,
				TabIndex:          req.TabIndex,
				URL:               finalURL,
				Title:             strings.TrimSpace(result.Title),
				Source:            identity.EventSource,
				Actor:             identity.Actor,
				Force:             force,
				Review:            extractReview,
			},
		)
		return marshalBrowserExtractPayload(browserExtractToolPayload{
			browserRegistrationRouteCapabilityPayload: browserRegistrationRouteCapabilityPayload{
				CapabilityMetadata: identity.CapabilityMetadata,
			},
			URL:                 req.URL,
			FinalURL:            finalURL,
			Backend:             strings.TrimSpace(result.Backend),
			BrowserApp:          resolvedBrowserApp,
			Profile:             runtimeInfo.Profile,
			RuntimeTarget:       runtimeInfo.Target,
			BrowserSurface:      identity.BrowserSurface,
			BrowserOptInTargets: append([]string(nil), identity.BrowserOptInTargets...),
			Title:               strings.TrimSpace(result.Title),
			Content:             content,
			ContentType:         strings.TrimSpace(result.ContentType),
			Status:              "extracted",
			Force:               force,
			ReviewDecision:      pageActionResult.ReviewDecision,
			ReviewReady:         pageActionResult.ReviewReady,
			Truncated:           truncated,
			Target:              target.Value,
			TargetID:            pageActionResult.TargetID,
			TabIndex:            req.TabIndex,
			WaitMs:              req.WaitMs,
			Note:                pageActionResult.Note,
		})
	})
}

func registerBrowserScreenshotTool(ctx browserRegistrationContext) {
	ctx.reg.Register(browserScreenshotDefinition(), func(callCtx context.Context, call types.FunctionCall) (string, error) {
		toolName, err := browserRegistrationCompatToolNameForActKind("screenshot")
		if err != nil {
			return "", err
		}
		effectivePolicy, err := resolveToolRuntimeSharedFetchPolicy(callCtx, ctx.policy)
		if err != nil {
			return "", browserRegistrationCompatToolErrorf(toolName, "invalid runtime network policy: %w", err)
		}
		params, err := decodeArgs(call.Arguments)
		if err != nil {
			return "", err
		}
		lane, dispatch, err := resolveBrowserExecutionLaneForRegistrationPageAction(ctx, callCtx, params, browserRegistrationPageActionDispatchOptions{
			UseManagedRoute:               true,
			UseManagedDefaultDispatchBase: true,
		})
		if err != nil {
			return "", browserRegistrationCompatToolErrorf(toolName, "%w", err)
		}
		routedBackend := lane.Backend
		runtimeInfo := lane.Runtime
		hiddenImplicitHostDefaultBase := dispatch.HiddenImplicitHostDefaultBase
		explicitRuntimeTarget := dispatch.ExplicitRuntimeTarget
		target := dispatch.Target
		browserApp := dispatch.BrowserApp
		identity := browserRegistrationCompatRuntimeIdentity(ctx, lane.Capabilities, toolName, "")
		force := firstBool(params, "force")
		screenshotReview := browserPendingTargetReviewStateForRuntimeTarget(callCtx, ctx.sessionRegistry, runtimeInfo, browserApp, target)
		screenshotReviewTargetID := strings.TrimSpace(target.TargetID)
		if screenshotReview.Review != nil {
			screenshotReviewTargetID = firstNonEmpty(screenshotReviewTargetID, strings.TrimSpace(screenshotReview.Review.ID))
		}
		rawURL := firstString(params, "url")
		validatedURL := ""
		if rawURL != "" {
			parsed, err := effectivePolicy.validateURL(callCtx, rawURL)
			if err != nil {
				return "", browserRegistrationCompatToolErrorf(toolName, "%w", err)
			}
			validatedURL = parsed.String()
		}
		if err := identity.legacyHostFallbackError(hiddenImplicitHostDefaultBase, explicitRuntimeTarget, runtimeInfo, target, validatedURL); err != nil {
			return "", err
		}
		if validatedURL == "" {
			screenshotReview = browserAutoFollowPendingTargetReviewStateForRuntimeTarget(callCtx, ctx.sessionRegistry, runtimeInfo, hiddenImplicitHostDefaultBase, browserApp, target)
			if screenshotReview.Review != nil {
				screenshotReviewTargetID = firstNonEmpty(strings.TrimSpace(target.TargetID), strings.TrimSpace(screenshotReview.Review.ID))
			}
		}
		waitMs := firstInt(params, "wait_ms")
		if waitMs <= 0 {
			if validatedURL != "" {
				waitMs = ctx.screenshotWaitMs
			} else {
				waitMs = browserTabTargetWaitMs
			}
		}
		elementTarget, err := resolveBrowserElementTarget(firstString(params, "selector"), firstString(params, "ref", "element_ref"))
		if err != nil {
			return "", browserRegistrationCompatToolErrorf(toolName, "%w", err)
		}
		if err := browserValidateElementTargetPageBinding(callCtx, ctx.sessionRegistry, runtimeInfo, hiddenImplicitHostDefaultBase, browserApp, target, validatedURL, elementTarget); err != nil {
			return "", browserRegistrationCompatToolErrorf(toolName, "%w", err)
		}
		selector := elementTarget.Selector
		fullPage := firstBool(params, "full_page")
		if selector != "" && fullPage {
			return "", browserRegistrationCompatToolErrorf(toolName, "ref/selector and full_page cannot be used together")
		}
		pathValue := firstString(params, "path", "output", "output_path")
		if strings.TrimSpace(pathValue) == "" {
			pathValue = defaultBrowserScreenshotRelPath()
		}
		resolvedPath, displayPath, err := resolvePathWithinRoot(ctx.opts.Root, pathValue)
		if err != nil {
			return "", browserRegistrationCompatToolErrorf(toolName, "%w", err)
		}
		if screenshotReview.Review != nil && !force {
			return marshalBrowserPageActionReviewBlockedPayload(browserPageActionReviewBlockedPayload{
				browserRegistrationRouteCapabilityPayload: browserRegistrationRouteCapabilityPayload{
					CapabilityMetadata: identity.CapabilityMetadata,
				},
				Path:                displayPath,
				Backend:             strings.TrimSpace(runtimeInfo.Backend),
				BrowserApp:          strings.TrimSpace(browserApp),
				Profile:             runtimeInfo.Profile,
				RuntimeTarget:       runtimeInfo.Target,
				BrowserSurface:      identity.BrowserSurface,
				BrowserOptInTargets: append([]string(nil), identity.BrowserOptInTargets...),
				Status:              "review_required",
				Force:               force,
				ReviewDecision:      browserPendingTargetReviewDecisionWithState(screenshotReview, force),
				ReviewReady:         false,
				Target:              target.Value,
				TargetID:            screenshotReviewTargetID,
				TabIndex:            target.TabIndex,
				WaitMs:              waitMs,
				Note:                identity.reviewReason(screenshotReview, force),
			})
		}
		req := BrowserScreenshotRequest{
			URL:               validatedURL,
			BrowserApp:        browserApp,
			WaitMs:            waitMs,
			OutputPath:        resolvedPath,
			ElementRef:        elementTarget.Ref,
			ElementHint:       browserElementHintForTarget(elementTarget),
			Selector:          selector,
			FullPage:          fullPage,
			TabIndex:          target.TabIndex,
			PreferredTargetID: screenshotReviewTargetID,
			Actor:             identity.Actor,
			Force:             force,
			Review:            screenshotReview,
		}
		var result BrowserScreenshotResult
		var recovery browserManagedResolverRecoveryResult
		artifactBytes, artifactPublished, err := publishBrowserArtifactOutput(
			callCtx,
			ctx.opts.PublishArtifact,
			routedBackend,
			ctx.opts.Root,
			"screenshot",
			resolvedPath,
			func(stagePath string) (string, bool, error) {
				stageReq := req
				stageReq.OutputPath = stagePath
				managed, executeErr := browserResolvedExecutionRouteExecuteManaged(
					callCtx,
					dispatch.Route,
					func(backend BrowserBackend) (BrowserScreenshotResult, error) {
						return backend.Screenshot(callCtx, stageReq)
					},
					browserManagedRouteExecutionArgs{
						URL:        stageReq.URL,
						BrowserApp: browserApp,
						WaitMs:     stageReq.WaitMs,
						TabIndex:   stageReq.TabIndex,
						Force:      force,
						FinalURL:   dispatch.Route.managedFinalURL(callCtx, browserApp, target, stageReq.URL),
					},
					func(policy browserManagedResolverFailurePolicyResult) BrowserScreenshotResult {
						return BrowserScreenshotResult{
							Backend:         policy.Backend,
							BrowserApp:      policy.BrowserApp,
							Path:            displayPath,
							FinalURL:        policy.FinalURL,
							Title:           policy.Title,
							Status:          policy.Status,
							Note:            policy.Note,
							ResolverOutcome: policy.Outcome,
						}
					},
				)
				if executeErr != nil {
					return "", false, executeErr
				}
				result = managed.Result
				recovery = managed.Recovery
				return result.Path, browserResolverOutcomeAllowsTargetTracking(result.ResolverOutcome), nil
			},
		)
		if err != nil {
			return "", err
		}
		artifactDisplayPath := displayPath
		var media *agentxmedia.Descriptor
		var artifacts []browserArtifactPayload
		if artifactPublished {
			media = browserScreenshotMediaDescriptor(artifactDisplayPath, artifactBytes, firstNonEmpty(strings.TrimSpace(result.FinalURL), req.URL), result)
			artifacts = browserScreenshotArtifacts(artifactDisplayPath, artifactBytes, firstNonEmpty(strings.TrimSpace(result.FinalURL), req.URL), result)
		}
		pageActionResult := agentxbrowserruntime.SharedSessionBrowserPageActionResultEventResult{
			TargetID: strings.TrimSpace(screenshotReviewTargetID),
		}
		if browserResolverOutcomeAllowsTargetTracking(result.ResolverOutcome) {
			pageActionResult = agentxbrowserruntime.ApplySharedSessionBrowserPageActionResultWithContext(
				browserSharedMutationContext(ctx.watchManagerProvider, ctx.sessionRegistry),
				agentxbrowserruntime.SharedSessionBrowserPageActionResultEventRequest{
					SessionID:         ToolSessionIDFromContext(callCtx),
					Route:             browserSessionRoute(runtimeInfo, firstNonEmpty(strings.TrimSpace(result.BrowserApp), req.BrowserApp), strings.TrimSpace(result.Backend)),
					PreferredTargetID: screenshotReviewTargetID,
					TabIndex:          req.TabIndex,
					URL:               firstNonEmpty(strings.TrimSpace(result.FinalURL), req.URL),
					Title:             strings.TrimSpace(result.Title),
					Source:            identity.EventSource,
					Actor:             identity.Actor,
					Force:             force,
					Review:            screenshotReview,
				},
			)
		}
		targetID := browserManagedResolverApplyTargetInvalidation(pageActionResult.TargetID, recovery)
		return marshalBrowserScreenshotPayload(browserScreenshotToolPayload{
			browserRegistrationRouteCapabilityPayload: browserRegistrationRouteCapabilityPayload{
				CapabilityMetadata: identity.CapabilityMetadata,
			},
			URL:                 req.URL,
			FinalURL:            firstNonEmpty(strings.TrimSpace(result.FinalURL), req.URL),
			Title:               strings.TrimSpace(result.Title),
			Path:                artifactDisplayPath,
			FilesTouched:        browserArtifactTouchedPaths(artifactDisplayPath),
			Bytes:               artifactBytes,
			Media:               media,
			Artifacts:           artifacts,
			Backend:             strings.TrimSpace(result.Backend),
			BrowserApp:          firstNonEmpty(strings.TrimSpace(result.BrowserApp), req.BrowserApp),
			Profile:             runtimeInfo.Profile,
			RuntimeTarget:       runtimeInfo.Target,
			BrowserSurface:      identity.BrowserSurface,
			BrowserOptInTargets: append([]string(nil), identity.BrowserOptInTargets...),
			CaptureScope:        strings.TrimSpace(result.CaptureScope),
			CaptureWidth:        result.CaptureWidth,
			CaptureHeight:       result.CaptureHeight,
			Status:              firstNonEmpty(strings.TrimSpace(result.Status), "captured"),
			ResolverOutcome:     result.ResolverOutcome,
			Actionability:       result.Actionability,
			FailureEvidence:     result.FailureEvidence,
			RecoveryAction:      browserResolverRecoveryAction(result.ResolverOutcome),
			Force:               force,
			ReviewDecision:      pageActionResult.ReviewDecision,
			ReviewReady:         pageActionResult.ReviewReady,
			Target:              target.Value,
			TargetID:            targetID,
			TabIndex:            req.TabIndex,
			Ref:                 req.ElementRef,
			Selector:            req.Selector,
			FullPage:            req.FullPage,
			WaitMs:              req.WaitMs,
			Note: func() string {
				if screenshotReview.Review != nil {
					return firstNonEmpty(strings.TrimSpace(result.Note), pageActionResult.Note)
				}
				return strings.TrimSpace(result.Note)
			}(),
		})
	})
}

func registerBrowserClickTool(ctx browserRegistrationContext) {
	ctx.reg.Register(browserClickDefinition(), func(callCtx context.Context, call types.FunctionCall) (string, error) {
		toolName, err := browserRegistrationCompatToolNameForActKind("click")
		if err != nil {
			return "", err
		}
		effectivePolicy, err := resolveToolRuntimeSharedFetchPolicy(callCtx, ctx.policy)
		if err != nil {
			return "", browserRegistrationCompatToolErrorf(toolName, "invalid runtime network policy: %w", err)
		}
		params, err := decodeArgs(call.Arguments)
		if err != nil {
			return "", err
		}
		lane, dispatch, err := resolveBrowserExecutionLaneForRegistrationPageAction(ctx, callCtx, params, browserRegistrationPageActionDispatchOptions{
			UseManagedRoute:               true,
			UseManagedDefaultDispatchBase: true,
		})
		if err != nil {
			return "", browserRegistrationCompatToolErrorf(toolName, "%w", err)
		}
		runtimeInfo := lane.Runtime
		hiddenImplicitHostDefaultBase := dispatch.HiddenImplicitHostDefaultBase
		explicitRuntimeTarget := dispatch.ExplicitRuntimeTarget
		target := dispatch.Target
		browserApp := dispatch.BrowserApp
		identity := browserRegistrationCompatRuntimeIdentity(ctx, lane.Capabilities, toolName, "")
		force := firstBool(params, "force")
		clickReview := browserPendingTargetReviewStateForRuntimeTarget(callCtx, ctx.sessionRegistry, runtimeInfo, browserApp, target)
		if strings.TrimSpace(firstString(params, "url")) == "" {
			clickReview = browserAutoFollowPendingTargetReviewStateForRuntimeTarget(callCtx, ctx.sessionRegistry, runtimeInfo, hiddenImplicitHostDefaultBase, browserApp, target)
		}
		clickReviewTargetID := strings.TrimSpace(target.TargetID)
		if clickReview.Review != nil {
			clickReviewTargetID = firstNonEmpty(clickReviewTargetID, strings.TrimSpace(clickReview.Review.ID))
		}
		elementTarget, err := resolveBrowserActionElementTarget(
			firstString(params, "selector"),
			firstString(params, "ref", "element_ref"),
			browserClickElementHintValue(params),
		)
		if err != nil {
			return "", browserRegistrationCompatToolErrorf(toolName, "%w", err)
		}
		if !browserElementTargetHasActionableLocator(elementTarget) {
			repairable, safeAutorepair, repairs := browserLocatorRepairAdviceFromParams(params, "element", "text", "label")
			return "", browserMissingLocatorError(
				toolName,
				fmt.Sprintf("%s: selector or ref is required", toolName),
				repairable,
				safeAutorepair,
				repairs,
			)
		}
		req := BrowserClickRequest{
			URL:               strings.TrimSpace(firstString(params, "url")),
			BrowserApp:        browserApp,
			WaitMs:            firstInt(params, "wait_ms"),
			PostWaitMs:        firstInt(params, "post_wait_ms"),
			ElementRef:        elementTarget.Ref,
			ElementHint:       browserClickElementHintForTarget(elementTarget),
			Selector:          elementTarget.Selector,
			TabIndex:          target.TabIndex,
			PreferredTargetID: firstNonEmpty(clickReviewTargetID, strings.TrimSpace(target.TargetID)),
			Actor:             identity.Actor,
			Force:             force,
			Review:            clickReview,
		}
		if req.URL != "" {
			parsed, err := effectivePolicy.validateURL(callCtx, req.URL)
			if err != nil {
				return "", browserRegistrationCompatToolErrorf(toolName, "%w", err)
			}
			req.URL = parsed.String()
		}
		if err := identity.legacyHostFallbackError(hiddenImplicitHostDefaultBase, explicitRuntimeTarget, runtimeInfo, target, req.URL); err != nil {
			return "", err
		}
		if err := browserValidateElementTargetPageBinding(callCtx, ctx.sessionRegistry, runtimeInfo, hiddenImplicitHostDefaultBase, browserApp, target, req.URL, elementTarget); err != nil {
			return "", browserRegistrationCompatToolErrorf(toolName, "%w", err)
		}
		if req.WaitMs <= 0 {
			req.WaitMs = browserActInteractivePageActionWaitMs(req.URL, 0, ctx.openWaitMs, target)
		}
		if req.PostWaitMs <= 0 {
			req.PostWaitMs = 750
		}
		if clickReview.Review != nil && !force {
			return marshalBrowserPageActionReviewBlockedPayload(browserPageActionReviewBlockedPayload{
				browserRegistrationRouteCapabilityPayload: browserRegistrationRouteCapabilityPayload{
					CapabilityMetadata: identity.CapabilityMetadata,
				},
				Backend:             strings.TrimSpace(runtimeInfo.Backend),
				BrowserApp:          strings.TrimSpace(browserApp),
				Profile:             runtimeInfo.Profile,
				RuntimeTarget:       runtimeInfo.Target,
				BrowserSurface:      identity.BrowserSurface,
				BrowserOptInTargets: append([]string(nil), identity.BrowserOptInTargets...),
				Ref:                 req.ElementRef,
				Selector:            req.Selector,
				Status:              "review_required",
				Force:               force,
				ReviewDecision:      browserPendingTargetReviewDecisionWithState(clickReview, force),
				ReviewReady:         false,
				Target:              target.Value,
				TargetID:            clickReviewTargetID,
				TabIndex:            req.TabIndex,
				WaitMs:              req.WaitMs,
				PostWaitMs:          req.PostWaitMs,
				Note:                identity.reviewReason(clickReview, force),
			})
		}
		managed, err := browserResolvedExecutionRouteExecuteManaged(
			callCtx,
			dispatch.Route,
			func(backend BrowserBackend) (BrowserClickResult, error) {
				return backend.Click(callCtx, req)
			},
			browserManagedRouteExecutionArgs{
				URL:        req.URL,
				BrowserApp: browserApp,
				WaitMs:     req.WaitMs,
				TabIndex:   req.TabIndex,
				Force:      force,
				FinalURL:   dispatch.Route.managedFinalURL(callCtx, browserApp, target, req.URL),
			},
			func(policy browserManagedResolverFailurePolicyResult) BrowserClickResult {
				return BrowserClickResult{
					Backend:         policy.Backend,
					BrowserApp:      policy.BrowserApp,
					FinalURL:        policy.FinalURL,
					Title:           policy.Title,
					Status:          policy.Status,
					Note:            policy.Note,
					ResolverOutcome: policy.Outcome,
				}
			},
		)
		if err != nil {
			return "", err
		}
		result := managed.Result
		recovery := managed.Recovery
		pageActionResult := agentxbrowserruntime.SharedSessionBrowserPageActionResultEventResult{
			TargetID: firstNonEmpty(clickReviewTargetID, strings.TrimSpace(target.TargetID)),
		}
		if browserResolverOutcomeAllowsTargetTracking(result.ResolverOutcome) {
			pageActionResult = agentxbrowserruntime.ApplySharedSessionBrowserPageActionResultWithContext(
				browserSharedMutationContext(ctx.watchManagerProvider, ctx.sessionRegistry),
				agentxbrowserruntime.SharedSessionBrowserPageActionResultEventRequest{
					SessionID:         ToolSessionIDFromContext(callCtx),
					Route:             browserSessionRoute(runtimeInfo, firstNonEmpty(strings.TrimSpace(result.BrowserApp), req.BrowserApp), strings.TrimSpace(result.Backend)),
					PreferredTargetID: firstNonEmpty(clickReviewTargetID, strings.TrimSpace(target.TargetID)),
					TabIndex:          req.TabIndex,
					URL:               strings.TrimSpace(result.FinalURL),
					Title:             strings.TrimSpace(result.Title),
					Source:            identity.EventSource,
					Actor:             identity.Actor,
					Force:             force,
					Review:            clickReview,
				},
			)
		}
		targetID := browserManagedResolverApplyTargetInvalidation(pageActionResult.TargetID, recovery)
		return marshalBrowserClickPayload(browserClickToolPayload{
			browserRegistrationRouteCapabilityPayload: browserRegistrationRouteCapabilityPayload{
				CapabilityMetadata: identity.CapabilityMetadata,
			},
			URL:                 req.URL,
			FinalURL:            strings.TrimSpace(result.FinalURL),
			Backend:             strings.TrimSpace(result.Backend),
			BrowserApp:          firstNonEmpty(strings.TrimSpace(result.BrowserApp), req.BrowserApp),
			Profile:             runtimeInfo.Profile,
			RuntimeTarget:       runtimeInfo.Target,
			BrowserSurface:      identity.BrowserSurface,
			BrowserOptInTargets: append([]string(nil), identity.BrowserOptInTargets...),
			Ref:                 req.ElementRef,
			Selector:            req.Selector,
			Title:               strings.TrimSpace(result.Title),
			Snapshot:            recovery.SnapshotText,
			SnapshotFormat:      strings.TrimSpace(recovery.Snapshot.Format),
			SnapshotMode:        strings.TrimSpace(recovery.Snapshot.Mode),
			SnapshotRefs: func() string {
				if recovery.SnapshotRecovered {
					return firstNonEmpty(strings.TrimSpace(recovery.Snapshot.Refs), "role")
				}
				return ""
			}(),
			SnapshotFrame:       strings.TrimSpace(recovery.Snapshot.Frame),
			Elements:            append([]BrowserSnapshotElement(nil), recovery.Snapshot.Elements...),
			Status:              firstNonEmpty(strings.TrimSpace(result.Status), "clicked"),
			RecoveryAction:      browserResolverRecoveryAction(result.ResolverOutcome),
			Force:               force,
			ReviewDecision:      pageActionResult.ReviewDecision,
			ReviewReady:         pageActionResult.ReviewReady,
			Target:              target.Value,
			TargetID:            targetID,
			TabIndex:            req.TabIndex,
			WaitMs:              req.WaitMs,
			PostWaitMs:          req.PostWaitMs,
			SnapshotInteractive: recovery.Snapshot.Interactive || recovery.SnapshotRecovered,
			SnapshotCompact:     recovery.Snapshot.Compact,
			SnapshotDepth:       recovery.Snapshot.Depth,
			Truncated:           recovery.SnapshotTruncated,
			ResolverOutcome:     result.ResolverOutcome,
			Actionability:       result.Actionability,
			FailureEvidence:     result.FailureEvidence,
			Note: func() string {
				if clickReview.Review != nil {
					return firstNonEmpty(strings.TrimSpace(result.Note), pageActionResult.Note)
				}
				return strings.TrimSpace(result.Note)
			}(),
		})
	})
}

func registerBrowserTypeTool(ctx browserRegistrationContext) {
	ctx.reg.Register(browserTypeDefinition(), func(callCtx context.Context, call types.FunctionCall) (string, error) {
		toolName, err := browserRegistrationCompatToolNameForActKind("type")
		if err != nil {
			return "", err
		}
		effectivePolicy, err := resolveToolRuntimeSharedFetchPolicy(callCtx, ctx.policy)
		if err != nil {
			return "", browserRegistrationCompatToolErrorf(toolName, "invalid runtime network policy: %w", err)
		}
		params, err := decodeArgs(call.Arguments)
		if err != nil {
			return "", err
		}
		lane, dispatch, err := resolveBrowserExecutionLaneForRegistrationPageAction(ctx, callCtx, params, browserRegistrationPageActionDispatchOptions{
			UseManagedRoute:               true,
			UseManagedDefaultDispatchBase: true,
		})
		if err != nil {
			return "", browserRegistrationCompatToolErrorf(toolName, "%w", err)
		}
		runtimeInfo := lane.Runtime
		hiddenImplicitHostDefaultBase := dispatch.HiddenImplicitHostDefaultBase
		explicitRuntimeTarget := dispatch.ExplicitRuntimeTarget
		target := dispatch.Target
		browserApp := dispatch.BrowserApp
		identity := browserRegistrationCompatRuntimeIdentity(ctx, lane.Capabilities, toolName, "")
		force := firstBool(params, "force")
		typeReview := browserPendingTargetReviewStateForRuntimeTarget(callCtx, ctx.sessionRegistry, runtimeInfo, browserApp, target)
		if strings.TrimSpace(firstString(params, "url")) == "" {
			typeReview = browserAutoFollowPendingTargetReviewStateForRuntimeTarget(callCtx, ctx.sessionRegistry, runtimeInfo, hiddenImplicitHostDefaultBase, browserApp, target)
		}
		typeReviewTargetID := strings.TrimSpace(target.TargetID)
		if typeReview.Review != nil {
			typeReviewTargetID = firstNonEmpty(typeReviewTargetID, strings.TrimSpace(typeReview.Review.ID))
		}
		elementTarget, err := resolveBrowserActionElementTarget(
			firstString(params, "selector"),
			firstString(params, "ref", "element_ref"),
			firstString(params, "element"),
		)
		if err != nil {
			return "", browserRegistrationCompatToolErrorf(toolName, "%w", err)
		}
		if !browserElementTargetHasActionableLocator(elementTarget) {
			repairable, safeAutorepair, repairs := browserLocatorRepairAdviceFromParams(params, "element", "label")
			return "", browserMissingLocatorError(
				toolName,
				fmt.Sprintf("%s: selector or ref is required", toolName),
				repairable,
				safeAutorepair,
				repairs,
			)
		}
		req := BrowserTypeRequest{
			URL:               strings.TrimSpace(firstString(params, "url")),
			BrowserApp:        browserApp,
			WaitMs:            firstInt(params, "wait_ms"),
			PostWaitMs:        firstInt(params, "post_wait_ms"),
			ElementRef:        elementTarget.Ref,
			ElementHint:       browserElementHintForTarget(elementTarget),
			Selector:          elementTarget.Selector,
			Text:              firstString(params, "text", "value"),
			Submit:            firstBool(params, "submit"),
			TabIndex:          target.TabIndex,
			PreferredTargetID: firstNonEmpty(typeReviewTargetID, strings.TrimSpace(target.TargetID)),
			Actor:             identity.Actor,
			Force:             force,
			Review:            typeReview,
		}
		if strings.TrimSpace(req.Text) == "" {
			repairable, safeAutorepair, repairs := browserTypeValueRepairAdviceFromParams(params)
			return "", browserMissingTypeTextErrorWithRepair(toolName, repairable, safeAutorepair, repairs)
		}
		if req.URL != "" {
			parsed, err := effectivePolicy.validateURL(callCtx, req.URL)
			if err != nil {
				return "", browserRegistrationCompatToolErrorf(toolName, "%w", err)
			}
			req.URL = parsed.String()
		}
		if err := identity.legacyHostFallbackError(hiddenImplicitHostDefaultBase, explicitRuntimeTarget, runtimeInfo, target, req.URL); err != nil {
			return "", err
		}
		if err := browserValidateElementTargetPageBinding(callCtx, ctx.sessionRegistry, runtimeInfo, hiddenImplicitHostDefaultBase, browserApp, target, req.URL, elementTarget); err != nil {
			return "", browserRegistrationCompatToolErrorf(toolName, "%w", err)
		}
		if req.WaitMs <= 0 {
			req.WaitMs = browserActInteractivePageActionWaitMs(req.URL, 0, ctx.openWaitMs, target)
		}
		if req.PostWaitMs <= 0 {
			req.PostWaitMs = 250
		}
		if typeReview.Review != nil && !force {
			return marshalBrowserPageActionReviewBlockedPayload(browserPageActionReviewBlockedPayload{
				browserRegistrationRouteCapabilityPayload: browserRegistrationRouteCapabilityPayload{
					CapabilityMetadata: identity.CapabilityMetadata,
				},
				Backend:             strings.TrimSpace(runtimeInfo.Backend),
				BrowserApp:          strings.TrimSpace(browserApp),
				Profile:             runtimeInfo.Profile,
				RuntimeTarget:       runtimeInfo.Target,
				BrowserSurface:      identity.BrowserSurface,
				BrowserOptInTargets: append([]string(nil), identity.BrowserOptInTargets...),
				Ref:                 req.ElementRef,
				Selector:            req.Selector,
				Value:               req.Text,
				Status:              "review_required",
				Force:               force,
				ReviewDecision:      browserPendingTargetReviewDecisionWithState(typeReview, force),
				ReviewReady:         false,
				Target:              target.Value,
				TargetID:            typeReviewTargetID,
				TabIndex:            req.TabIndex,
				WaitMs:              req.WaitMs,
				PostWaitMs:          req.PostWaitMs,
				Note:                identity.reviewReason(typeReview, force),
			})
		}
		managed, err := browserResolvedExecutionRouteExecuteManaged(
			callCtx,
			dispatch.Route,
			func(backend BrowserBackend) (BrowserTypeResult, error) {
				return backend.Type(callCtx, req)
			},
			browserManagedRouteExecutionArgs{
				URL:        req.URL,
				BrowserApp: browserApp,
				WaitMs:     req.WaitMs,
				TabIndex:   req.TabIndex,
				Force:      force,
				FinalURL:   dispatch.Route.managedFinalURL(callCtx, browserApp, target, req.URL),
			},
			func(policy browserManagedResolverFailurePolicyResult) BrowserTypeResult {
				return BrowserTypeResult{
					Backend:         policy.Backend,
					BrowserApp:      policy.BrowserApp,
					FinalURL:        policy.FinalURL,
					Title:           policy.Title,
					Value:           req.Text,
					Status:          policy.Status,
					ResolverOutcome: policy.Outcome,
					Note:            policy.Note,
				}
			},
		)
		if err != nil {
			return "", err
		}
		result := managed.Result
		recovery := managed.Recovery
		pageActionResult := agentxbrowserruntime.SharedSessionBrowserPageActionResultEventResult{
			TargetID: firstNonEmpty(typeReviewTargetID, strings.TrimSpace(target.TargetID)),
		}
		if browserResolverOutcomeAllowsTargetTracking(result.ResolverOutcome) {
			pageActionResult = agentxbrowserruntime.ApplySharedSessionBrowserPageActionResultWithContext(
				browserSharedMutationContext(ctx.watchManagerProvider, ctx.sessionRegistry),
				agentxbrowserruntime.SharedSessionBrowserPageActionResultEventRequest{
					SessionID:         ToolSessionIDFromContext(callCtx),
					Route:             browserSessionRoute(runtimeInfo, firstNonEmpty(strings.TrimSpace(result.BrowserApp), req.BrowserApp), strings.TrimSpace(result.Backend)),
					PreferredTargetID: firstNonEmpty(typeReviewTargetID, strings.TrimSpace(target.TargetID)),
					TabIndex:          req.TabIndex,
					URL:               strings.TrimSpace(result.FinalURL),
					Title:             strings.TrimSpace(result.Title),
					Source:            identity.EventSource,
					Actor:             identity.Actor,
					Force:             force,
					Review:            typeReview,
				},
			)
		}
		targetID := browserManagedResolverApplyTargetInvalidation(pageActionResult.TargetID, recovery)
		return marshalBrowserTypePayload(browserTypeToolPayload{
			browserRegistrationRouteCapabilityPayload: browserRegistrationRouteCapabilityPayload{
				CapabilityMetadata: identity.CapabilityMetadata,
			},
			URL:                 req.URL,
			FinalURL:            strings.TrimSpace(result.FinalURL),
			Backend:             strings.TrimSpace(result.Backend),
			BrowserApp:          firstNonEmpty(strings.TrimSpace(result.BrowserApp), req.BrowserApp),
			Profile:             runtimeInfo.Profile,
			RuntimeTarget:       runtimeInfo.Target,
			BrowserSurface:      identity.BrowserSurface,
			BrowserOptInTargets: append([]string(nil), identity.BrowserOptInTargets...),
			Ref:                 req.ElementRef,
			Selector:            req.Selector,
			Title:               strings.TrimSpace(result.Title),
			Value:               result.Value,
			Snapshot:            recovery.SnapshotText,
			SnapshotFormat:      strings.TrimSpace(recovery.Snapshot.Format),
			SnapshotMode:        strings.TrimSpace(recovery.Snapshot.Mode),
			SnapshotRefs: func() string {
				if recovery.SnapshotRecovered {
					return firstNonEmpty(strings.TrimSpace(recovery.Snapshot.Refs), "role")
				}
				return ""
			}(),
			SnapshotFrame:       strings.TrimSpace(recovery.Snapshot.Frame),
			Elements:            append([]BrowserSnapshotElement(nil), recovery.Snapshot.Elements...),
			Status:              firstNonEmpty(strings.TrimSpace(result.Status), "typed"),
			Submitted:           result.Submitted,
			ResolverOutcome:     result.ResolverOutcome,
			Actionability:       result.Actionability,
			FailureEvidence:     result.FailureEvidence,
			RecoveryAction:      browserResolverRecoveryAction(result.ResolverOutcome),
			Force:               force,
			ReviewDecision:      pageActionResult.ReviewDecision,
			ReviewReady:         pageActionResult.ReviewReady,
			Target:              target.Value,
			TargetID:            targetID,
			TabIndex:            req.TabIndex,
			WaitMs:              req.WaitMs,
			PostWaitMs:          req.PostWaitMs,
			SnapshotInteractive: recovery.Snapshot.Interactive || recovery.SnapshotRecovered,
			SnapshotCompact:     recovery.Snapshot.Compact,
			SnapshotDepth:       recovery.Snapshot.Depth,
			Truncated:           recovery.SnapshotTruncated,
			Note: func() string {
				if typeReview.Review != nil {
					return firstNonEmpty(strings.TrimSpace(result.Note), pageActionResult.Note)
				}
				return strings.TrimSpace(result.Note)
			}(),
		})
	})
}

func registerBrowserEvalTool(ctx browserRegistrationContext) {
	ctx.reg.Register(browserEvalDefinition(), func(callCtx context.Context, call types.FunctionCall) (string, error) {
		toolName, err := browserRegistrationCompatToolNameForActKind("evaluate")
		if err != nil {
			return "", err
		}
		effectivePolicy, err := resolveToolRuntimeSharedFetchPolicy(callCtx, ctx.policy)
		if err != nil {
			return "", browserRegistrationCompatToolErrorf(toolName, "invalid runtime network policy: %w", err)
		}
		params, err := decodeArgs(call.Arguments)
		if err != nil {
			return "", err
		}
		lane, dispatch, err := resolveBrowserExecutionLaneForRegistrationPageAction(ctx, callCtx, params, browserRegistrationPageActionDispatchOptions{
			UseManagedDefaultDispatchBase: true,
		})
		if err != nil {
			return "", browserRegistrationCompatToolErrorf(toolName, "%w", err)
		}
		routedBackend := lane.Backend
		runtimeInfo := lane.Runtime
		hiddenImplicitHostDefaultBase := dispatch.HiddenImplicitHostDefaultBase
		explicitRuntimeTarget := dispatch.ExplicitRuntimeTarget
		target := dispatch.Target
		browserApp := dispatch.BrowserApp
		identity := browserRegistrationCompatRuntimeIdentity(ctx, lane.Capabilities, toolName, "")
		force := firstBool(params, "force")
		evalReview := browserPendingTargetReviewStateForRuntimeTarget(callCtx, ctx.sessionRegistry, runtimeInfo, browserApp, target)
		if strings.TrimSpace(firstString(params, "url")) == "" {
			evalReview = browserAutoFollowPendingTargetReviewStateForRuntimeTarget(callCtx, ctx.sessionRegistry, runtimeInfo, hiddenImplicitHostDefaultBase, browserApp, target)
		}
		evalReviewTargetID := strings.TrimSpace(target.TargetID)
		if evalReview.Review != nil {
			evalReviewTargetID = firstNonEmpty(evalReviewTargetID, strings.TrimSpace(evalReview.Review.ID))
		}
		script := strings.TrimSpace(firstString(params, "script", "javascript", "js"))
		if script == "" {
			return "", browserRegistrationCompatToolErrorf(toolName, "script is required")
		}
		req := BrowserEvalRequest{
			URL:               strings.TrimSpace(firstString(params, "url")),
			BrowserApp:        browserApp,
			WaitMs:            firstInt(params, "wait_ms"),
			Script:            script,
			MaxChars:          firstInt(params, "max_chars"),
			TabIndex:          target.TabIndex,
			PreferredTargetID: evalReviewTargetID,
			Actor:             identity.Actor,
			Force:             force,
			Review:            evalReview,
		}
		if req.URL != "" {
			parsed, err := effectivePolicy.validateURL(callCtx, req.URL)
			if err != nil {
				return "", browserRegistrationCompatToolErrorf(toolName, "%w", err)
			}
			req.URL = parsed.String()
		}
		if err := identity.legacyHostFallbackError(hiddenImplicitHostDefaultBase, explicitRuntimeTarget, runtimeInfo, target, req.URL); err != nil {
			return "", err
		}
		if req.WaitMs <= 0 {
			if req.URL != "" {
				req.WaitMs = ctx.openWaitMs
			} else if target.Explicit {
				req.WaitMs = browserTabTargetWaitMs
			}
		}
		if req.MaxChars <= 0 || req.MaxChars > ctx.maxChars {
			req.MaxChars = ctx.maxChars
		}
		if evalReview.Review != nil && !force {
			return marshalBrowserPageActionReviewBlockedPayload(browserPageActionReviewBlockedPayload{
				browserRegistrationRouteCapabilityPayload: browserRegistrationRouteCapabilityPayload{
					CapabilityMetadata: identity.CapabilityMetadata,
				},
				Backend:             strings.TrimSpace(runtimeInfo.Backend),
				BrowserApp:          strings.TrimSpace(browserApp),
				Profile:             runtimeInfo.Profile,
				RuntimeTarget:       runtimeInfo.Target,
				BrowserSurface:      identity.BrowserSurface,
				BrowserOptInTargets: append([]string(nil), identity.BrowserOptInTargets...),
				Status:              "review_required",
				Force:               force,
				ReviewDecision:      browserPendingTargetReviewDecisionWithState(evalReview, force),
				ReviewReady:         false,
				Target:              target.Value,
				TargetID:            evalReviewTargetID,
				TabIndex:            req.TabIndex,
				WaitMs:              req.WaitMs,
				Note:                identity.reviewReason(evalReview, force),
			})
		}
		safetyDecision, err := browserCheckCDPEscapeHatchPolicy(ctx.opts)
		if err != nil {
			return "", browserRegistrationCompatToolErrorf(toolName, "%w", err)
		}
		result, err := routedBackend.Eval(callCtx, req)
		if err != nil {
			return "", err
		}
		rendered := strings.TrimSpace(result.Result)
		truncated := false
		if trimmed, changed := trimToMaxChars(rendered, req.MaxChars); changed {
			rendered = trimmed
			truncated = true
		}
		pageActionResult := agentxbrowserruntime.ApplySharedSessionBrowserPageActionResultWithContext(
			browserSharedMutationContext(ctx.watchManagerProvider, ctx.sessionRegistry),
			agentxbrowserruntime.SharedSessionBrowserPageActionResultEventRequest{
				SessionID:         ToolSessionIDFromContext(callCtx),
				Route:             browserSessionRoute(runtimeInfo, firstNonEmpty(strings.TrimSpace(result.BrowserApp), req.BrowserApp), strings.TrimSpace(result.Backend)),
				PreferredTargetID: evalReviewTargetID,
				TabIndex:          req.TabIndex,
				URL:               strings.TrimSpace(result.FinalURL),
				Title:             strings.TrimSpace(result.Title),
				Source:            identity.EventSource,
				Actor:             identity.Actor,
				Force:             force,
				Review:            evalReview,
			},
		)
		return marshalBrowserEvalPayload(browserEvalToolPayload{
			browserRegistrationRouteCapabilityPayload: browserRegistrationRouteCapabilityPayload{
				CapabilityMetadata: identity.CapabilityMetadata,
			},
			URL:                 req.URL,
			FinalURL:            strings.TrimSpace(result.FinalURL),
			Backend:             strings.TrimSpace(result.Backend),
			BrowserApp:          firstNonEmpty(strings.TrimSpace(result.BrowserApp), req.BrowserApp),
			Profile:             runtimeInfo.Profile,
			RuntimeTarget:       runtimeInfo.Target,
			BrowserSurface:      identity.BrowserSurface,
			BrowserOptInTargets: append([]string(nil), identity.BrowserOptInTargets...),
			Title:               strings.TrimSpace(result.Title),
			Result:              rendered,
			Status:              firstNonEmpty(strings.TrimSpace(result.Status), "evaluated"),
			Force:               force,
			ReviewDecision:      pageActionResult.ReviewDecision,
			ReviewReady:         pageActionResult.ReviewReady,
			SafetyEvent:         browserResultSafetyEventForCDPEscapeHatch(toolName, result.Backend, runtimeInfo.Target, safetyDecision),
			Target:              target.Value,
			TargetID:            pageActionResult.TargetID,
			TabIndex:            req.TabIndex,
			WaitMs:              req.WaitMs,
			Truncated:           truncated,
			Note: func() string {
				if evalReview.Review != nil {
					return firstNonEmpty(strings.TrimSpace(result.Note), pageActionResult.Note)
				}
				return strings.TrimSpace(result.Note)
			}(),
		})
	})
}

func registerBrowserActTool(ctx browserRegistrationContext) {
	ctx.reg.Register(browserActDefinition(browserRegistrationActKinds(ctx)), func(callCtx context.Context, call types.FunctionCall) (string, error) {
		params, err := decodeArgs(call.Arguments)
		if err != nil {
			return "", err
		}
		executionPreview := browserRegistrationExecutionPreviewForContext(ctx)
		kind := browserNormalizeToolToken(firstString(params, "kind"))
		if kind == "" {
			return "", browserMissingActKindError("browser_act", "")
		}
		if current := firstString(params, "kind"); current != kind {
			params = browserUnifiedCloneParams(params)
			params["kind"] = kind
		}
		watchManagerProvider := agentxbrowserruntime.SharedSessionBrowserObserverManagerFor(
			ctx.sessionRegistry,
			ctx.opts.SessionRunRegistry,
			ctx.sessionStateRegistry,
			browserRuntimeReconnectWatchdogWindow,
		)
		lane, dispatch, err := resolveBrowserExecutionLaneForActDispatch(
			callCtx,
			executionPreview.EffectiveBackend,
			ctx.sessionRegistry,
			ctx.sessionStateRegistry,
			watchManagerProvider,
			executionPreview.DispatchBase,
			executionPreview.HiddenImplicitHostDefaultBase,
			ctx.opts,
			ctx.maxChars,
			params,
			kind,
		)
		if err != nil {
			return "", err
		}
		result, err := executeBrowserActResolved(
			callCtx,
			ctx.sessionRegistry,
			ctx.sessionStateRegistry,
			ctx.policy,
			ctx.opts,
			ctx.openWaitMs,
			ctx.maxChars,
			params,
			lane,
			dispatch,
		)
		if err != nil {
			return "", err
		}
		if firstBool(params, "remember_target", "remember") {
			if !browserRememberDecisionRequiresPopupReview(result.RememberDecision, result.RememberReady) {
				rememberTargetID, rememberTabIndex := browserActRememberTargetCandidate(result)
				rememberResult := applyBrowserRememberTargetSelection(callCtx, ctx, browserRememberTargetApplyOptions{
					Route: browserSessionRoute(
						BrowserRuntimeInfo{
							Backend: result.Backend,
							Profile: result.Profile,
							Target:  result.RuntimeTarget,
						},
						result.BrowserApp,
						result.Backend,
					),
					TargetID:                 rememberTargetID,
					TabIndex:                 rememberTabIndex,
					ExistingDecision:         result.RememberDecision,
					ExistingReady:            result.RememberReady,
					PreserveExistingOnSelect: false,
				})
				result.SessionTargetSelection = rememberResult.Selection
				if strings.TrimSpace(result.RememberDecision) == "" || rememberResult.Selection == nil {
					result.RememberDecision = rememberResult.Decision
					result.RememberReady = rememberResult.Ready
				}
			}
		}
		var sessionProfileSelection *browserRuntimeSessionProfileSelection
		if result.SessionTargetSelection != nil {
			sessionProfileSelection = browserRuntimePromoteProfileFromTargetSelection(callCtx, ctx.sessionStateRegistry, browserRuntimeSessionTargetSelectionFromShared(result.SessionTargetSelection))
		}
		browserSurface, browserOptInTargets := browserRegistrationActPayloadRouteSurface(ctx, result.Kind)
		capabilityMetadata := browserRegistrationPayloadCapabilityMetadata(ctx, lane.Capabilities, browserSurface, browserOptInTargets)
		payloadStatus := firstNonEmpty(strings.TrimSpace(result.Status), "ok")
		note := strings.TrimSpace(result.Note)
		return marshalBrowserActPayload(browserActToolPayload{
			browserRegistrationRouteCapabilityPayload: browserRegistrationRouteCapabilityPayload{
				CapabilityMetadata: capabilityMetadata,
			},
			Kind:                    result.Kind,
			Action:                  strings.TrimSpace(result.Action),
			URL:                     strings.TrimSpace(firstString(params, "url")),
			Paths:                   append([]string(nil), result.Paths...),
			FilesTouched:            append([]string(nil), result.FilesTouched...),
			FinalURL:                strings.TrimSpace(result.FinalURL),
			Backend:                 strings.TrimSpace(result.Backend),
			BrowserApp:              strings.TrimSpace(result.BrowserApp),
			Profile:                 strings.TrimSpace(result.Profile),
			RuntimeTarget:           strings.TrimSpace(result.RuntimeTarget),
			BrowserSurface:          browserSurface,
			BrowserOptInTargets:     append([]string(nil), browserOptInTargets...),
			Title:                   strings.TrimSpace(result.Title),
			Content:                 result.Content,
			ContentType:             strings.TrimSpace(result.ContentType),
			Snapshot:                result.Snapshot,
			SnapshotFormat:          strings.TrimSpace(result.SnapshotFormat),
			SnapshotMode:            strings.TrimSpace(result.SnapshotMode),
			SnapshotRefs:            strings.TrimSpace(result.SnapshotRefs),
			SnapshotInteractive:     result.SnapshotInteractive,
			SnapshotCompact:         result.SnapshotCompact,
			SnapshotDepth:           result.SnapshotDepth,
			SnapshotFrame:           strings.TrimSpace(result.SnapshotFrame),
			Elements:                append([]BrowserSnapshotElement(nil), result.Elements...),
			Messages:                append([]BrowserConsoleMessage(nil), result.Messages...),
			Requests:                append([]BrowserRequestEntry(nil), result.Requests...),
			RequestURL:              strings.TrimSpace(result.RequestURL),
			RequestMethod:           strings.TrimSpace(result.RequestMethod),
			ResponseStatusCode:      result.ResponseStatusCode,
			Errors:                  append([]BrowserErrorEntry(nil), result.Errors...),
			Cookies:                 append([]BrowserCookieEntry(nil), result.Cookies...),
			StorageKind:             strings.TrimSpace(result.StorageKind),
			Storage:                 append([]BrowserStorageEntry(nil), result.Storage...),
			HeaderNames:             append([]string(nil), result.HeaderNames...),
			HeaderCount:             result.HeaderCount,
			Path:                    result.Path,
			Bytes:                   result.Bytes,
			Media:                   browserActResultMediaDescriptor(result),
			Artifacts:               browserActArtifacts(result),
			CaptureScope:            strings.TrimSpace(result.CaptureScope),
			CaptureWidth:            result.CaptureWidth,
			CaptureHeight:           result.CaptureHeight,
			Key:                     strings.TrimSpace(result.Key),
			Result:                  result.Result,
			Value:                   result.Value,
			Values:                  append([]string(nil), result.Values...),
			FieldCount:              result.FieldCount,
			Width:                   result.Width,
			Height:                  result.Height,
			Status:                  payloadStatus,
			Force:                   result.Force,
			ReviewDecision:          strings.TrimSpace(result.ReviewDecision),
			ReviewReady:             result.ReviewReady,
			PostNavigationSnapshot:  browserPostNavigationSnapshotRecommendationForAct(result.Kind, payloadStatus, result.FinalURL),
			BotDetectionWarning:     browserBotDetectionWarningFromNavigation(result.FinalURL, result.Title, note),
			SafetyEvent:             browserActResultSafetyEvent(ctx.opts, result),
			Target:                  strings.TrimSpace(result.Target),
			TargetID:                strings.TrimSpace(result.TargetID),
			RememberDecision:        strings.TrimSpace(result.RememberDecision),
			RememberReady:           result.RememberReady,
			SessionProfileSelection: sessionProfileSelection,
			SessionTargetSelection:  result.SessionTargetSelection,
			ResolverOutcome:         result.ResolverOutcome,
			Actionability:           result.Actionability,
			FailureEvidence:         result.FailureEvidence,
			ResolvedViaFallback:     result.ResolvedViaFallback,
			ResolverFallbackKind:    strings.TrimSpace(result.ResolverFallbackKind),
			ResolverFallbackIndex:   browserCloneOptionalInt(result.ResolverFallbackIndex),
			ResolverFallbackStrength: strings.TrimSpace(
				result.ResolverFallbackCandidateStrength,
			),
			ResolverFallbackBlockedBy:      strings.TrimSpace(result.ResolverFallbackBlockedBy),
			ResolverFallbackAmbiguityClass: strings.TrimSpace(result.ResolverFallbackAmbiguityClass),
			ResolverFallbackManualRetryHint: strings.TrimSpace(
				result.ResolverFallbackManualRetryHint,
			),
			ResolverFallbackSpecificityFields: append(
				[]string(nil),
				result.ResolverFallbackSpecificityFields...,
			),
			ResolverBlockedBy:      strings.TrimSpace(result.ResolverBlockedBy),
			ResolverAmbiguityClass: strings.TrimSpace(result.ResolverAmbiguityClass),
			ResolverCandidateKind:  strings.TrimSpace(result.ResolverCandidateKind),
			ResolverCandidateStrength: strings.TrimSpace(
				result.ResolverCandidateStrength,
			),
			ResolverRetryDisposition: strings.TrimSpace(result.ResolverRetryDisposition),
			ResolverManualRetryHint:  strings.TrimSpace(result.ResolverManualRetryHint),
			ResolverNextStepAlias:    strings.TrimSpace(result.ResolverNextStepAlias),
			BrowserLocalPlanner: browserLocalPlannerSummaryForActResult(
				ctx.opts,
				result.Kind,
				params,
				result,
			),
			RecoveryAction: strings.TrimSpace(result.RecoveryAction),
			Submitted:      result.Submitted,
			Truncated:      result.Truncated,
			TabIndex:       result.TabIndex,
			ActiveIndex:    result.ActiveIndex,
			Tabs:           append([]BrowserTab(nil), result.Tabs...),
			WaitMs:         firstInt(params, "wait_ms"),
			PostWaitMs:     firstInt(params, "post_wait_ms"),
			Ref:            strings.TrimSpace(result.Ref),
			Selector:       strings.TrimSpace(result.Selector),
			FullPage:       firstBool(params, "full_page"),
			Note:           note,
		})
	})
}

func browserRegistrationActKinds(ctx browserRegistrationContext) []string {
	kinds := append([]string(nil), ctx.capabilities.SupportedActKinds()...)
	if len(kinds) != 0 {
		return kinds
	}
	return browserRegistrationExplicitManagedActKinds(ctx)
}
