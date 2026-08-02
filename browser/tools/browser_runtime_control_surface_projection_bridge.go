package tools

import (
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

type browserRuntimeTopLevelSurfaceProjection struct {
	Review  *browserReviewSurfaceSummary
	Surface *browserTopLevelSurfaceSummary
	View    *browserTopLevelViewSummary
}

func browserRuntimeSharedSummaryPtr(
	summary *browserTopLevelSummary,
) *agentxbrowserruntime.SharedSessionBrowserSummary {
	if summary == nil {
		return nil
	}
	out := &agentxbrowserruntime.SharedSessionBrowserSummary{
		Category:             strings.TrimSpace(summary.Category),
		State:                strings.TrimSpace(summary.State),
		SummaryCode:          strings.TrimSpace(summary.SummaryCode),
		NextStepAlias:        strings.TrimSpace(summary.NextStepAlias),
		ManualRetryHint:      strings.TrimSpace(summary.ManualRetryHint),
		ResolvedViaFallback:  summary.ResolvedViaFallback,
		PrimaryBrowserAction: strings.TrimSpace(summary.PrimaryBrowserAction),
		PrimaryNodeAction:    strings.TrimSpace(summary.PrimaryNodeAction),
		NextStep:             strings.TrimSpace(summary.NextStep),
	}
	if browserUnifiedSummaryEmpty(*summary) {
		return nil
	}
	return out
}

func browserRuntimeSharedDisplayPtr(
	display *browserTopLevelDisplaySummary,
) *agentxbrowserruntime.SharedSessionBrowserDisplay {
	if display == nil {
		return nil
	}
	out := &agentxbrowserruntime.SharedSessionBrowserDisplay{
		Ready:    display.Ready,
		Sections: append([]string(nil), display.Sections...),
		SharedSessionBrowserSummary: agentxbrowserruntime.SharedSessionBrowserSummary{
			Category:             strings.TrimSpace(display.Category),
			State:                strings.TrimSpace(display.State),
			SummaryCode:          strings.TrimSpace(display.SummaryCode),
			NextStepAlias:        strings.TrimSpace(display.NextStepAlias),
			ManualRetryHint:      strings.TrimSpace(display.ManualRetryHint),
			ResolvedViaFallback:  display.ResolvedViaFallback,
			PrimaryBrowserAction: strings.TrimSpace(display.PrimaryBrowserAction),
			PrimaryNodeAction:    strings.TrimSpace(display.PrimaryNodeAction),
			NextStep:             strings.TrimSpace(display.NextStep),
		},
	}
	if browserTopLevelDisplayEmpty(*display) {
		return nil
	}
	return out
}

func browserRuntimeSharedWorkbenchDisplayPtr(
	display *browserRuntimeWorkbenchDisplaySummary,
) *agentxbrowserruntime.SharedSessionBrowserDisplay {
	if display == nil {
		return nil
	}
	out := &agentxbrowserruntime.SharedSessionBrowserDisplay{
		Ready:    display.Ready,
		Sections: append([]string(nil), display.Sections...),
		SharedSessionBrowserSummary: agentxbrowserruntime.SharedSessionBrowserSummary{
			Category:             strings.TrimSpace(display.Category),
			State:                strings.TrimSpace(display.State),
			SummaryCode:          strings.TrimSpace(display.SummaryCode),
			NextStepAlias:        strings.TrimSpace(display.NextStepAlias),
			ManualRetryHint:      strings.TrimSpace(display.ManualRetryHint),
			ResolvedViaFallback:  display.ResolvedViaFallback,
			PrimaryBrowserAction: strings.TrimSpace(display.PrimaryBrowserAction),
			PrimaryNodeAction:    strings.TrimSpace(display.PrimaryNodeAction),
			NextStep:             strings.TrimSpace(display.NextStep),
		},
	}
	if browserUnifiedWorkbenchDisplayEmpty(*display) {
		return nil
	}
	return out
}

func browserRuntimeSharedReviewSurfacePtr(
	review *browserReviewSurfaceSummary,
) *agentxbrowserruntime.SharedSessionBrowserReviewSurface {
	if review == nil || browserReviewSurfaceSummaryEmpty(*review) {
		return nil
	}
	return &agentxbrowserruntime.SharedSessionBrowserReviewSurface{
		PolicyState: strings.TrimSpace(review.PolicyState),
		Decision:    strings.TrimSpace(review.Decision),
		Ready:       review.Ready,
		Explanation: browserRuntimeSharedSummaryPtr(review.Explanation),
		Diagnostics: browserRuntimeSharedSummaryPtr(review.Diagnostics),
		Summary:     browserRuntimeSharedSummaryPtr(review.Summary),
		Display:     browserRuntimeSharedDisplayPtr(review.Display),
	}
}

func browserRuntimeSharedViewPtr(
	view *browserTopLevelViewSummary,
) *agentxbrowserruntime.SharedSessionBrowserView {
	if view == nil || browserTopLevelViewEmpty(*view) {
		return nil
	}
	return &agentxbrowserruntime.SharedSessionBrowserView{
		Kind:                strings.TrimSpace(view.Kind),
		Ready:               view.Ready,
		Sections:            append([]string(nil), view.Sections...),
		Review:              browserRuntimeSharedReviewSurfacePtr(view.Review),
		BrowserTools:        append([]string(nil), view.BrowserTools...),
		ArtifactTools:       append([]string(nil), view.ArtifactTools...),
		ArtifactKinds:       append([]string(nil), view.ArtifactKinds...),
		ArtifactContract:    strings.TrimSpace(view.ArtifactContract),
		BrowserActKinds:     append([]string(nil), view.BrowserActKinds...),
		BrowserSurface:      strings.TrimSpace(view.BrowserSurface),
		BrowserOptInTargets: append([]string(nil), view.BrowserOptInTargets...),
		SharedSessionBrowserSummary: agentxbrowserruntime.SharedSessionBrowserSummary{
			Category:             strings.TrimSpace(view.Category),
			State:                strings.TrimSpace(view.State),
			SummaryCode:          strings.TrimSpace(view.SummaryCode),
			NextStepAlias:        strings.TrimSpace(view.NextStepAlias),
			ManualRetryHint:      strings.TrimSpace(view.ManualRetryHint),
			ResolvedViaFallback:  view.ResolvedViaFallback,
			PrimaryBrowserAction: strings.TrimSpace(view.PrimaryBrowserAction),
			PrimaryNodeAction:    strings.TrimSpace(view.PrimaryNodeAction),
			NextStep:             strings.TrimSpace(view.NextStep),
		},
	}
}

func browserRuntimeSharedSurfacePtr(
	surface *browserTopLevelSurfaceSummary,
) *agentxbrowserruntime.SharedSessionBrowserSurface {
	if surface == nil || browserTopLevelSurfaceEmpty(*surface) {
		return nil
	}
	return &agentxbrowserruntime.SharedSessionBrowserSurface{
		Ready:               surface.Ready,
		Sections:            append([]string(nil), surface.Sections...),
		ReviewPolicyState:   strings.TrimSpace(surface.ReviewPolicyState),
		ReviewDecision:      strings.TrimSpace(surface.ReviewDecision),
		ReviewReady:         surface.ReviewReady,
		BrowserTools:        append([]string(nil), surface.BrowserTools...),
		ArtifactTools:       append([]string(nil), surface.ArtifactTools...),
		ArtifactKinds:       append([]string(nil), surface.ArtifactKinds...),
		ArtifactContract:    strings.TrimSpace(surface.ArtifactContract),
		BrowserActKinds:     append([]string(nil), surface.BrowserActKinds...),
		BrowserSurface:      strings.TrimSpace(surface.BrowserSurface),
		BrowserOptInTargets: append([]string(nil), surface.BrowserOptInTargets...),
		SharedSessionBrowserSummary: agentxbrowserruntime.SharedSessionBrowserSummary{
			Category:             strings.TrimSpace(surface.Category),
			State:                strings.TrimSpace(surface.State),
			SummaryCode:          strings.TrimSpace(surface.SummaryCode),
			NextStepAlias:        strings.TrimSpace(surface.NextStepAlias),
			ManualRetryHint:      strings.TrimSpace(surface.ManualRetryHint),
			ResolvedViaFallback:  surface.ResolvedViaFallback,
			PrimaryBrowserAction: strings.TrimSpace(surface.PrimaryBrowserAction),
			PrimaryNodeAction:    strings.TrimSpace(surface.PrimaryNodeAction),
			NextStep:             strings.TrimSpace(surface.NextStep),
		},
	}
}

func browserRuntimeSharedWorkbenchSurfacePtr(
	surface *browserRuntimeWorkbenchSurfaceSummary,
) *agentxbrowserruntime.SharedSessionBrowserWorkbenchSurface {
	if surface == nil || browserUnifiedWorkbenchEmpty(*surface) {
		return nil
	}
	return &agentxbrowserruntime.SharedSessionBrowserWorkbenchSurface{
		Ready:                     surface.Ready,
		Sections:                  append([]string(nil), surface.Sections...),
		Review:                    browserRuntimeSharedReviewSurfacePtr(surface.Review),
		BrowserTools:              append([]string(nil), surface.BrowserTools...),
		ArtifactTools:             append([]string(nil), surface.ArtifactTools...),
		ArtifactKinds:             append([]string(nil), surface.ArtifactKinds...),
		ArtifactContract:          strings.TrimSpace(surface.ArtifactContract),
		BrowserActKinds:           append([]string(nil), surface.BrowserActKinds...),
		BrowserSurface:            strings.TrimSpace(surface.BrowserSurface),
		BrowserOptInTargets:       append([]string(nil), surface.BrowserOptInTargets...),
		Explanation:               browserRuntimeSharedSummaryPtr(surface.Explanation),
		Diagnostics:               browserRuntimeSharedSummaryPtr(surface.Diagnostics),
		Summary:                   browserRuntimeSharedSummaryPtr(surface.Summary),
		PrimaryBrowserAction:      strings.TrimSpace(surface.PrimaryBrowserAction),
		PrimaryNodeAction:         strings.TrimSpace(surface.PrimaryNodeAction),
		NextStep:                  strings.TrimSpace(surface.NextStep),
		RecommendedBrowserActions: append([]string(nil), surface.RecommendedBrowserActions...),
		RecommendedNodeActions:    append([]string(nil), surface.RecommendedNodeActions...),
	}
}

func browserRuntimeSharedReviewState(
	payload browserRuntimePayload,
) agentxbrowserruntime.SharedSessionBrowserReviewState {
	return agentxbrowserruntime.SelectSharedSessionBrowserReviewState(
		agentxbrowserruntime.SharedSessionBrowserReviewStateRequest{
			Candidates: []agentxbrowserruntime.SharedSessionBrowserReviewDecisionCandidate{
				{Decision: payload.SelectTargetDecision, Ready: payload.SelectTargetReady},
				{Decision: payload.PrepareDecision, Ready: payload.PrepareReady},
				{Decision: payload.StopDecision, Ready: payload.StopReady},
				{Decision: payload.RestartDecision, Ready: payload.RestartReady},
				{Decision: payload.CreateDecision, Ready: payload.CreateReady},
				{Decision: payload.DeleteDecision, Ready: payload.DeleteReady},
				{Decision: payload.SelectDecision, Ready: payload.SelectReady},
				{Decision: payload.ClearDecision, Ready: payload.ClearReady},
				{Decision: payload.ClearSessionDecision, Ready: payload.ClearSessionReady},
				{Decision: payload.SyncSessionDecision, Ready: payload.SyncSessionReady},
				{Decision: payload.CoordinationDecision, Ready: payload.CoordinationReady},
			},
			Routes: browserRuntimeRouteCoordinationInputs(payload.SessionRoutes),
		},
	)
}

func browserRuntimeTopLevelSurfaceAndViewFromDisplayAndReview(
	display *browserTopLevelDisplaySummary,
	review *browserReviewSurfaceSummary,
) (*browserTopLevelSurfaceSummary, *browserTopLevelViewSummary) {
	return browserRuntimeTopLevelSurfaceAndViewFromDisplayAndReviewWithMetadata(
		display,
		review,
		browserTopLevelCapabilitySurface{},
		"",
		nil,
	)
}

func browserRuntimeTopLevelSurfaceAndViewFromDisplayAndReviewWithRouteSurface(
	display *browserTopLevelDisplaySummary,
	review *browserReviewSurfaceSummary,
	browserSurface string,
	browserOptInTargets []string,
) (*browserTopLevelSurfaceSummary, *browserTopLevelViewSummary) {
	return browserRuntimeTopLevelSurfaceAndViewFromDisplayAndReviewWithMetadata(
		display,
		review,
		browserTopLevelCapabilitySurface{},
		browserSurface,
		browserOptInTargets,
	)
}

func browserRuntimeTopLevelSurfaceAndViewFromDisplayAndReviewWithMetadata(
	display *browserTopLevelDisplaySummary,
	review *browserReviewSurfaceSummary,
	capability browserTopLevelCapabilitySurface,
	browserSurface string,
	browserOptInTargets []string,
) (*browserTopLevelSurfaceSummary, *browserTopLevelViewSummary) {
	sharedReview := browserRuntimeSharedReviewSurfacePtr(review)
	sharedSurface := agentxbrowserruntime.BuildSharedSessionBrowserSurfaceFromRequest(
		agentxbrowserruntime.SharedSessionBrowserSurfaceRequest{
			Display:             browserRuntimeSharedDisplayPtr(display),
			Review:              sharedReview,
			BrowserTools:        append([]string(nil), capability.BrowserTools...),
			ArtifactTools:       append([]string(nil), capability.ArtifactTools...),
			ArtifactKinds:       append([]string(nil), capability.ArtifactKinds...),
			ArtifactContract:    strings.TrimSpace(capability.ArtifactContract),
			BrowserActKinds:     append([]string(nil), capability.BrowserActKinds...),
			BrowserSurface:      strings.TrimSpace(browserSurface),
			BrowserOptInTargets: append([]string(nil), browserOptInTargets...),
		},
	)
	defaultCandidateRoute := firstBrowserRuntimeRouteDescriptor(
		browserRuntimeRouteDescriptorFromTopLevelDisplay(display),
		browserRuntimeRouteDescriptorFromReviewSurface(review),
	)
	surfaceSummary := browserRuntimeTopLevelSurfaceSummaryFromShared(sharedSurface)
	viewSummary := browserRuntimeTopLevelViewSummaryFromShared(
		agentxbrowserruntime.BuildSharedSessionBrowserViewFromRequest(
			agentxbrowserruntime.SharedSessionBrowserViewRequest{
				Surface:             sharedSurface,
				Review:              sharedReview,
				BrowserTools:        append([]string(nil), capability.BrowserTools...),
				ArtifactTools:       append([]string(nil), capability.ArtifactTools...),
				ArtifactKinds:       append([]string(nil), capability.ArtifactKinds...),
				ArtifactContract:    strings.TrimSpace(capability.ArtifactContract),
				BrowserActKinds:     append([]string(nil), capability.BrowserActKinds...),
				BrowserSurface:      strings.TrimSpace(browserSurface),
				BrowserOptInTargets: append([]string(nil), browserOptInTargets...),
			},
		),
	)
	if defaultCandidateRoute != (browserRuntimeRouteDescriptor{}) {
		surfaceSummary = browserTopLevelSurfaceWithDefaultCandidateRoute(surfaceSummary, defaultCandidateRoute)
		viewSummary = browserTopLevelViewWithDefaultCandidateRoute(viewSummary, defaultCandidateRoute)
	}
	return surfaceSummary, viewSummary
}

func browserRuntimeProjectTopLevelSurface(payload browserRuntimePayload) browserRuntimeTopLevelSurfaceProjection {
	review := agentxbrowserruntime.BuildSharedSessionBrowserReviewSurface(
		browserRuntimeSharedReviewSurfaceRequest(&payload),
	)
	surface := agentxbrowserruntime.BuildSharedSessionBrowserSurfaceFromRequest(
		agentxbrowserruntime.SharedSessionBrowserSurfaceRequest{
			Display:             browserRuntimeSharedDisplayPtr(payload.Display),
			Review:              review,
			BrowserTools:        append([]string(nil), payload.BrowserTools...),
			ArtifactTools:       append([]string(nil), payload.ArtifactTools...),
			ArtifactKinds:       append([]string(nil), payload.ArtifactKinds...),
			ArtifactContract:    strings.TrimSpace(payload.ArtifactContract),
			BrowserActKinds:     append([]string(nil), payload.BrowserActKinds...),
			BrowserSurface:      payload.BrowserSurface,
			BrowserOptInTargets: append([]string(nil), payload.BrowserOptInTargets...),
		},
	)
	var workbench *agentxbrowserruntime.SharedSessionBrowserWorkbenchSurface
	if browserRuntimeHasWorkbenchSurface(payload) {
		workbench = agentxbrowserruntime.BuildSharedSessionBrowserWorkbenchSurface(
			browserRuntimeSharedWorkbenchSurfaceRequest(&payload),
		)
	}
	view := agentxbrowserruntime.BuildSharedSessionBrowserViewFromRequest(
		agentxbrowserruntime.SharedSessionBrowserViewRequest{
			Workbench:           workbench,
			WorkbenchDisplay:    browserRuntimeSharedWorkbenchDisplayPtr(payload.WorkbenchDisplay),
			Surface:             surface,
			Review:              review,
			BrowserTools:        append([]string(nil), payload.BrowserTools...),
			ArtifactTools:       append([]string(nil), payload.ArtifactTools...),
			ArtifactKinds:       append([]string(nil), payload.ArtifactKinds...),
			ArtifactContract:    strings.TrimSpace(payload.ArtifactContract),
			BrowserActKinds:     append([]string(nil), payload.BrowserActKinds...),
			BrowserSurface:      payload.BrowserSurface,
			BrowserOptInTargets: append([]string(nil), payload.BrowserOptInTargets...),
		},
	)
	return browserRuntimeTopLevelSurfaceProjection{
		Review:  browserRuntimeReviewSurfaceSummaryFromShared(review),
		Surface: browserRuntimeTopLevelSurfaceSummaryFromShared(surface),
		View:    browserRuntimeTopLevelViewSummaryFromShared(view),
	}
}

func browserRuntimeApplyTopLevelSurfaceProjection(
	payload *browserRuntimePayload,
	projection browserRuntimeTopLevelSurfaceProjection,
) {
	if payload == nil {
		return
	}
	payload.Review = projection.Review
	payload.Surface = projection.Surface
	payload.View = projection.View
}

func browserRuntimeReviewSurfaceSummaryForTopLevel(
	reviewDecision string,
	reviewReady bool,
	explanation *browserTopLevelSummary,
	diagnostics *browserTopLevelSummary,
	summary *browserTopLevelSummary,
	display *browserTopLevelDisplaySummary,
) *browserReviewSurfaceSummary {
	review := browserRuntimeReviewSurfaceSummaryFromShared(
		agentxbrowserruntime.BuildSharedSessionBrowserReviewSurface(
			agentxbrowserruntime.SharedSessionBrowserReviewSurfaceRequest{
				ReviewDecision: strings.TrimSpace(reviewDecision),
				ReviewReady:    reviewReady,
				Explanation:    browserRuntimeSharedSummaryPtr(explanation),
				Diagnostics:    browserRuntimeSharedSummaryPtr(diagnostics),
				Summary:        browserRuntimeSharedSummaryPtr(summary),
				Display:        browserRuntimeSharedDisplayPtr(display),
			},
		),
	)
	route := firstBrowserRuntimeRouteDescriptor(
		browserRuntimeRouteDescriptorFromTopLevelSummary(explanation),
		browserRuntimeRouteDescriptorFromTopLevelSummary(diagnostics),
		browserRuntimeRouteDescriptorFromTopLevelSummary(summary),
		browserRuntimeRouteDescriptorFromTopLevelDisplay(display),
	)
	return browserReviewSurfaceSummaryWithDefaultCandidateRoute(review, route)
}

func browserRuntimeSharedReviewSurfaceRequest(
	payload *browserRuntimePayload,
) agentxbrowserruntime.SharedSessionBrowserReviewSurfaceRequest {
	if payload == nil {
		return agentxbrowserruntime.SharedSessionBrowserReviewSurfaceRequest{}
	}
	reviewState := browserRuntimeSharedReviewState(*payload)
	return agentxbrowserruntime.SharedSessionBrowserReviewSurfaceRequest{
		ReviewDecision: reviewState.Decision,
		ReviewReady:    reviewState.Ready,
		Explanation:    browserRuntimeSharedSummaryPtr(payload.Explanation),
		Diagnostics:    browserRuntimeSharedSummaryPtr(payload.Diagnostics),
		Summary:        browserRuntimeSharedSummaryPtr(payload.Summary),
		Display:        browserRuntimeSharedDisplayPtr(payload.Display),
	}
}

func browserRuntimeSharedWorkbenchSurfaceRequest(
	payload *browserRuntimePayload,
) agentxbrowserruntime.SharedSessionBrowserWorkbenchSurfaceRequest {
	if payload == nil {
		return agentxbrowserruntime.SharedSessionBrowserWorkbenchSurfaceRequest{}
	}
	reviewState := browserRuntimeSharedReviewState(*payload)
	return agentxbrowserruntime.SharedSessionBrowserWorkbenchSurfaceRequest{
		Ready:                     payload.WorkbenchReady,
		Sections:                  append([]string(nil), payload.WorkbenchSections...),
		BrowserTools:              append([]string(nil), payload.BrowserTools...),
		ArtifactTools:             append([]string(nil), payload.ArtifactTools...),
		ArtifactKinds:             append([]string(nil), payload.ArtifactKinds...),
		ArtifactContract:          strings.TrimSpace(payload.ArtifactContract),
		BrowserActKinds:           append([]string(nil), payload.BrowserActKinds...),
		BrowserSurface:            strings.TrimSpace(payload.BrowserSurface),
		BrowserOptInTargets:       append([]string(nil), payload.BrowserOptInTargets...),
		Explanation:               browserRuntimeSharedSummaryPtr(browserTopLevelSummaryFromRuntimeDiagnosticsExplanation(payload.WorkbenchExplanation)),
		Diagnostics:               browserRuntimeSharedSummaryPtr(browserTopLevelSummaryFromWorkbenchDiagnostics(payload.WorkbenchDiagnostics)),
		Summary:                   browserRuntimeSharedSummaryPtr(payload.WorkbenchSummary),
		PrimaryBrowserAction:      payload.WorkbenchPrimaryBrowserAction,
		PrimaryNodeAction:         payload.WorkbenchPrimaryNodeAction,
		NextStep:                  payload.WorkbenchNextStep,
		RecommendedBrowserActions: append([]string(nil), payload.WorkbenchRecommendedBrowserActions...),
		RecommendedNodeActions:    append([]string(nil), payload.WorkbenchRecommendedNodeActions...),
		ReviewDecision:            reviewState.Decision,
		ReviewReady:               reviewState.Ready,
	}
}

func browserRuntimeReviewSurfaceSummaryFromShared(
	review *agentxbrowserruntime.SharedSessionBrowserReviewSurface,
) *browserReviewSurfaceSummary {
	if review == nil {
		return nil
	}
	out := &browserReviewSurfaceSummary{
		PolicyState: strings.TrimSpace(review.PolicyState),
		Decision:    strings.TrimSpace(review.Decision),
		Ready:       review.Ready,
		Explanation: browserRuntimeTopLevelSummaryFromShared(review.Explanation),
		Diagnostics: browserRuntimeTopLevelSummaryFromShared(review.Diagnostics),
		Summary:     browserRuntimeTopLevelSummaryFromShared(review.Summary),
		Display:     browserRuntimeTopLevelDisplaySummaryFromShared(review.Display),
	}
	if browserReviewSurfaceSummaryEmpty(*out) {
		return nil
	}
	return out
}

func browserRuntimeWorkbenchSurfaceSummaryFromShared(
	surface *agentxbrowserruntime.SharedSessionBrowserWorkbenchSurface,
) *browserRuntimeWorkbenchSurfaceSummary {
	if surface == nil {
		return nil
	}
	out := &browserRuntimeWorkbenchSurfaceSummary{
		Ready:                     surface.Ready,
		Sections:                  append([]string(nil), surface.Sections...),
		Review:                    browserRuntimeReviewSurfaceSummaryFromShared(surface.Review),
		BrowserTools:              append([]string(nil), surface.BrowserTools...),
		ArtifactTools:             append([]string(nil), surface.ArtifactTools...),
		ArtifactKinds:             append([]string(nil), surface.ArtifactKinds...),
		ArtifactContract:          strings.TrimSpace(surface.ArtifactContract),
		BrowserActKinds:           append([]string(nil), surface.BrowserActKinds...),
		BrowserSurface:            strings.TrimSpace(surface.BrowserSurface),
		BrowserOptInTargets:       append([]string(nil), surface.BrowserOptInTargets...),
		Explanation:               browserRuntimeTopLevelSummaryFromShared(surface.Explanation),
		Diagnostics:               browserRuntimeTopLevelSummaryFromShared(surface.Diagnostics),
		Summary:                   browserRuntimeTopLevelSummaryFromShared(surface.Summary),
		PrimaryBrowserAction:      strings.TrimSpace(surface.PrimaryBrowserAction),
		PrimaryNodeAction:         strings.TrimSpace(surface.PrimaryNodeAction),
		NextStep:                  strings.TrimSpace(surface.NextStep),
		RecommendedBrowserActions: append([]string(nil), surface.RecommendedBrowserActions...),
		RecommendedNodeActions:    append([]string(nil), surface.RecommendedNodeActions...),
	}
	if browserUnifiedWorkbenchEmpty(*out) {
		return nil
	}
	return out
}

func browserRuntimeTopLevelSurfaceSummaryFromShared(
	surface *agentxbrowserruntime.SharedSessionBrowserSurface,
) *browserTopLevelSurfaceSummary {
	if surface == nil {
		return nil
	}
	out := &browserTopLevelSurfaceSummary{
		Ready:               surface.Ready,
		Sections:            append([]string(nil), surface.Sections...),
		Category:            strings.TrimSpace(surface.Category),
		State:               strings.TrimSpace(surface.State),
		SummaryCode:         strings.TrimSpace(surface.SummaryCode),
		NextStepAlias:       strings.TrimSpace(surface.NextStepAlias),
		ManualRetryHint:     strings.TrimSpace(surface.ManualRetryHint),
		ResolvedViaFallback: surface.ResolvedViaFallback,
		PrimaryBrowserAction: strings.TrimSpace(
			surface.PrimaryBrowserAction,
		),
		PrimaryNodeAction:   strings.TrimSpace(surface.PrimaryNodeAction),
		NextStep:            strings.TrimSpace(surface.NextStep),
		ReviewPolicyState:   strings.TrimSpace(surface.ReviewPolicyState),
		ReviewDecision:      strings.TrimSpace(surface.ReviewDecision),
		ReviewReady:         surface.ReviewReady,
		BrowserTools:        append([]string(nil), surface.BrowserTools...),
		ArtifactTools:       append([]string(nil), surface.ArtifactTools...),
		ArtifactKinds:       append([]string(nil), surface.ArtifactKinds...),
		ArtifactContract:    strings.TrimSpace(surface.ArtifactContract),
		BrowserActKinds:     append([]string(nil), surface.BrowserActKinds...),
		BrowserSurface:      strings.TrimSpace(surface.BrowserSurface),
		BrowserOptInTargets: append([]string(nil), surface.BrowserOptInTargets...),
	}
	if browserTopLevelSurfaceEmpty(*out) {
		return nil
	}
	return out
}

func browserRuntimeTopLevelViewSummaryFromShared(
	view *agentxbrowserruntime.SharedSessionBrowserView,
) *browserTopLevelViewSummary {
	if view == nil {
		return nil
	}
	out := &browserTopLevelViewSummary{
		Kind:                strings.TrimSpace(view.Kind),
		Ready:               view.Ready,
		Sections:            append([]string(nil), view.Sections...),
		Category:            strings.TrimSpace(view.Category),
		State:               strings.TrimSpace(view.State),
		SummaryCode:         strings.TrimSpace(view.SummaryCode),
		NextStepAlias:       strings.TrimSpace(view.NextStepAlias),
		ManualRetryHint:     strings.TrimSpace(view.ManualRetryHint),
		ResolvedViaFallback: view.ResolvedViaFallback,
		PrimaryBrowserAction: strings.TrimSpace(
			view.PrimaryBrowserAction,
		),
		PrimaryNodeAction:   strings.TrimSpace(view.PrimaryNodeAction),
		NextStep:            strings.TrimSpace(view.NextStep),
		BrowserTools:        append([]string(nil), view.BrowserTools...),
		ArtifactTools:       append([]string(nil), view.ArtifactTools...),
		ArtifactKinds:       append([]string(nil), view.ArtifactKinds...),
		ArtifactContract:    strings.TrimSpace(view.ArtifactContract),
		BrowserActKinds:     append([]string(nil), view.BrowserActKinds...),
		BrowserSurface:      strings.TrimSpace(view.BrowserSurface),
		BrowserOptInTargets: append([]string(nil), view.BrowserOptInTargets...),
		Review:              browserRuntimeReviewSurfaceSummaryFromShared(view.Review),
	}
	if browserTopLevelViewEmpty(*out) {
		return nil
	}
	return out
}
