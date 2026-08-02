package browserruntime

import "strings"

type sharedSessionBrowserCapabilitySurface struct {
	BrowserTools     []string
	ArtifactTools    []string
	ArtifactKinds    []string
	ArtifactContract string
	BrowserActKinds  []string
}

type sharedSessionBrowserRouteSurface struct {
	BrowserSurface      string
	BrowserOptInTargets []string
}

// SharedSessionBrowserRouteSurface captures the selected-route hints projected
// onto top-level browser aliases and nested shells.
type SharedSessionBrowserRouteSurface = sharedSessionBrowserRouteSurface

// SharedSessionBrowserCapabilitySurface captures the selected-route capability
// hints projected onto top-level browser result aliases and nested shells.
type SharedSessionBrowserCapabilitySurface = sharedSessionBrowserCapabilitySurface

// SharedSessionBrowserReviewSurface captures the shared review-specific
// explanation/diagnostics/display surface derived from a review decision.
type SharedSessionBrowserReviewSurface struct {
	PolicyState string
	Decision    string
	Ready       bool
	Explanation *SharedSessionBrowserSummary
	Diagnostics *SharedSessionBrowserSummary
	Summary     *SharedSessionBrowserSummary
	Display     *SharedSessionBrowserDisplay
}

// SharedSessionBrowserSurface captures the top-level display surface plus
// review posture exposed by runtime payloads.
type SharedSessionBrowserSurface struct {
	Ready               bool
	Sections            []string
	ReviewPolicyState   string
	ReviewDecision      string
	ReviewReady         bool
	BrowserTools        []string
	ArtifactTools       []string
	ArtifactKinds       []string
	ArtifactContract    string
	BrowserActKinds     []string
	BrowserSurface      string
	BrowserOptInTargets []string
	SharedSessionBrowserSummary
}

// SharedSessionBrowserView captures the stable view alias derived from either
// workbench or result/review surfaces.
type SharedSessionBrowserView struct {
	Kind                string
	Ready               bool
	Sections            []string
	Review              *SharedSessionBrowserReviewSurface
	BrowserTools        []string
	ArtifactTools       []string
	ArtifactKinds       []string
	ArtifactContract    string
	BrowserActKinds     []string
	BrowserSurface      string
	BrowserOptInTargets []string
	SharedSessionBrowserSummary
}

// SharedSessionBrowserSurfaceRequest carries the shared inputs needed to build
// the stable top-level surface including optional route-surface diagnostics.
type SharedSessionBrowserSurfaceRequest struct {
	Display             *SharedSessionBrowserDisplay
	Review              *SharedSessionBrowserReviewSurface
	BrowserTools        []string
	ArtifactTools       []string
	ArtifactKinds       []string
	ArtifactContract    string
	BrowserActKinds     []string
	BrowserSurface      string
	BrowserOptInTargets []string
}

// SharedSessionBrowserDisplayRequest carries the shared inputs needed to build
// the stable top-level display alias from existing display/surface/view/review
// shells plus optional summary fallbacks.
type SharedSessionBrowserDisplayRequest struct {
	Display          *SharedSessionBrowserDisplay
	WorkbenchDisplay *SharedSessionBrowserDisplay
	Surface          *SharedSessionBrowserSurface
	View             *SharedSessionBrowserView
	Review           *SharedSessionBrowserReviewSurface
	Summary          *SharedSessionBrowserSummary
	Diagnostics      *SharedSessionBrowserSummary
	Explanation      *SharedSessionBrowserSummary
}

// SharedSessionBrowserSurfaceAliasRequest carries the shared inputs needed to
// build the stable top-level surface alias while honoring explicit surface
// shells and route-surface precedence.
type SharedSessionBrowserSurfaceAliasRequest struct {
	Surface             *SharedSessionBrowserSurface
	Display             *SharedSessionBrowserDisplay
	Review              *SharedSessionBrowserReviewSurface
	View                *SharedSessionBrowserView
	BrowserTools        []string
	ArtifactTools       []string
	ArtifactKinds       []string
	ArtifactContract    string
	BrowserActKinds     []string
	BrowserSurface      string
	BrowserOptInTargets []string
}

// SharedSessionBrowserReviewAliasRequest carries the shared inputs needed to
// build the stable top-level review alias from explicit review/view/workbench
// shells.
type SharedSessionBrowserReviewAliasRequest struct {
	Review    *SharedSessionBrowserReviewSurface
	View      *SharedSessionBrowserView
	Workbench *SharedSessionBrowserWorkbenchSurface
}

// SharedSessionBrowserViewRequest carries the shared inputs needed to build
// the final view alias including optional route-surface diagnostics.
type SharedSessionBrowserViewRequest struct {
	Workbench           *SharedSessionBrowserWorkbenchSurface
	WorkbenchDisplay    *SharedSessionBrowserDisplay
	Surface             *SharedSessionBrowserSurface
	Review              *SharedSessionBrowserReviewSurface
	BrowserTools        []string
	ArtifactTools       []string
	ArtifactKinds       []string
	ArtifactContract    string
	BrowserActKinds     []string
	BrowserSurface      string
	BrowserOptInTargets []string
}

// SharedSessionBrowserViewAliasRequest carries the shared inputs needed to
// build the final view alias while honoring explicit view shells and
// route-surface precedence.
type SharedSessionBrowserViewAliasRequest struct {
	View                *SharedSessionBrowserView
	Workbench           *SharedSessionBrowserWorkbenchSurface
	WorkbenchDisplay    *SharedSessionBrowserDisplay
	Surface             *SharedSessionBrowserSurface
	Review              *SharedSessionBrowserReviewSurface
	BrowserTools        []string
	ArtifactTools       []string
	ArtifactKinds       []string
	ArtifactContract    string
	BrowserActKinds     []string
	BrowserSurface      string
	BrowserOptInTargets []string
}

// SharedSessionBrowserSummaryAliasRequest carries the shared inputs needed to
// build the stable top-level summary alias from explicit summary-like shells
// plus review/display/surface/view/workbench fallbacks.
type SharedSessionBrowserSummaryAliasRequest struct {
	Summary          *SharedSessionBrowserSummary
	WorkbenchSummary *SharedSessionBrowserSummary
	Diagnostics      *SharedSessionBrowserSummary
	Explanation      *SharedSessionBrowserSummary
	Review           *SharedSessionBrowserReviewSurface
	Display          *SharedSessionBrowserDisplay
	Surface          *SharedSessionBrowserSurface
	View             *SharedSessionBrowserView
	WorkbenchDisplay *SharedSessionBrowserDisplay
	Workbench        *SharedSessionBrowserWorkbenchSurface
}

// SharedSessionBrowserTopLevelAliasProjectionRequest carries the shared
// requests needed to build the review->summary->display->surface->view alias
// chain in one owner.
type SharedSessionBrowserTopLevelAliasProjectionRequest struct {
	Review  SharedSessionBrowserReviewAliasRequest
	Summary SharedSessionBrowserSummaryAliasRequest
	Display SharedSessionBrowserDisplayRequest
	Surface SharedSessionBrowserSurfaceAliasRequest
	View    SharedSessionBrowserViewAliasRequest
}

// SharedSessionBrowserTopLevelAliasProjection captures the shared top-level
// alias chain after review, summary, display, surface, and view have been
// threaded through the same owner.
type SharedSessionBrowserTopLevelAliasProjection struct {
	Review  *SharedSessionBrowserReviewSurface
	Summary *SharedSessionBrowserSummary
	Display *SharedSessionBrowserDisplay
	Surface *SharedSessionBrowserSurface
	View    *SharedSessionBrowserView
}

// SharedSessionBrowserCapabilityAliasRequest carries the shared inputs needed
// to recover top-level capability hints from explicit root fields or nested
// surface/view/workbench shells.
type SharedSessionBrowserCapabilityAliasRequest struct {
	Surface          *SharedSessionBrowserSurface
	View             *SharedSessionBrowserView
	Workbench        *SharedSessionBrowserWorkbenchSurface
	BrowserTools     []string
	ArtifactTools    []string
	ArtifactKinds    []string
	ArtifactContract string
	BrowserActKinds  []string
}

// SharedSessionBrowserRouteAliasRequest carries the shared inputs needed to
// recover top-level route hints from explicit root fields or nested
// surface/view/workbench shells.
type SharedSessionBrowserRouteAliasRequest struct {
	Surface             *SharedSessionBrowserSurface
	View                *SharedSessionBrowserView
	Workbench           *SharedSessionBrowserWorkbenchSurface
	BrowserSurface      string
	BrowserOptInTargets []string
}

// SharedSessionBrowserWorkbenchSurface captures the shared workbench surface
// derived from workbench summaries, action-plan fields, and review posture.
type SharedSessionBrowserWorkbenchSurface struct {
	Ready                     bool
	Sections                  []string
	Review                    *SharedSessionBrowserReviewSurface
	BrowserTools              []string
	ArtifactTools             []string
	ArtifactKinds             []string
	ArtifactContract          string
	BrowserActKinds           []string
	BrowserSurface            string
	BrowserOptInTargets       []string
	Explanation               *SharedSessionBrowserSummary
	Diagnostics               *SharedSessionBrowserSummary
	Summary                   *SharedSessionBrowserSummary
	PrimaryBrowserAction      string
	PrimaryNodeAction         string
	NextStep                  string
	RecommendedBrowserActions []string
	RecommendedNodeActions    []string
}

// SharedSessionBrowserReviewSurfaceRequest carries the shared inputs needed to
// build review-specific explanation/diagnostics/display aliases.
type SharedSessionBrowserReviewSurfaceRequest struct {
	ReviewDecision string
	ReviewReady    bool
	Explanation    *SharedSessionBrowserSummary
	Diagnostics    *SharedSessionBrowserSummary
	Summary        *SharedSessionBrowserSummary
	Display        *SharedSessionBrowserDisplay
}

// SharedSessionBrowserWorkbenchSurfaceRequest carries the shared inputs needed
// to build the workbench surface and nested review projection.
type SharedSessionBrowserWorkbenchSurfaceRequest struct {
	Ready                     bool
	Sections                  []string
	BrowserTools              []string
	ArtifactTools             []string
	ArtifactKinds             []string
	ArtifactContract          string
	BrowserActKinds           []string
	BrowserSurface            string
	BrowserOptInTargets       []string
	Explanation               *SharedSessionBrowserSummary
	Diagnostics               *SharedSessionBrowserSummary
	Summary                   *SharedSessionBrowserSummary
	PrimaryBrowserAction      string
	PrimaryNodeAction         string
	NextStep                  string
	RecommendedBrowserActions []string
	RecommendedNodeActions    []string
	ReviewDecision            string
	ReviewReady               bool
}

// BuildSharedSessionBrowserReviewSurface projects the shared review aliases
// derived from review decision/ready posture plus summary/display fields.
func BuildSharedSessionBrowserReviewSurface(
	req SharedSessionBrowserReviewSurfaceRequest,
) *SharedSessionBrowserReviewSurface {
	req.ReviewDecision = strings.TrimSpace(req.ReviewDecision)

	var topLevel *SharedSessionBrowserSummary
	switch {
	case sharedSessionBrowserSummaryIsReview(req.Summary):
		topLevel = req.Summary
	case sharedSessionBrowserSummaryIsReview(req.Diagnostics):
		topLevel = req.Diagnostics
	case sharedSessionBrowserSummaryIsReview(req.Explanation):
		topLevel = req.Explanation
	}
	policyState := strings.TrimSpace(sharedSessionBrowserReviewSummaryCode(req.ReviewDecision))
	if policyState == "" && topLevel != nil {
		policyState = strings.TrimSpace(topLevel.SummaryCode)
	}
	if policyState == "" &&
		req.ReviewDecision == "" &&
		!req.ReviewReady &&
		topLevel == nil &&
		(req.Display == nil || !strings.EqualFold(strings.TrimSpace(req.Display.Category), "review")) {
		return nil
	}

	out := &SharedSessionBrowserReviewSurface{
		PolicyState: policyState,
		Decision:    req.ReviewDecision,
		Ready:       req.ReviewReady,
	}
	if topLevel != nil || req.Explanation != nil || req.Diagnostics != nil || req.Summary != nil || req.Display != nil {
		out.Explanation = cloneSharedSessionBrowserSummary(req.Explanation)
		out.Diagnostics = cloneSharedSessionBrowserSummary(req.Diagnostics)
		out.Summary = cloneSharedSessionBrowserSummary(req.Summary)
		if req.Display != nil {
			out.Display = cloneSharedSessionBrowserDisplay(req.Display)
		} else {
			out.Display = sharedSessionBrowserDisplayFromSummary(out.Summary)
			if out.Display == nil {
				if out.Diagnostics != nil {
					out.Display = sharedSessionBrowserDisplayFromSummary(out.Diagnostics)
				} else if out.Explanation != nil {
					out.Display = sharedSessionBrowserDisplayFromSummary(out.Explanation)
				}
			}
		}
	}
	if sharedSessionBrowserReviewSurfaceEmpty(*out) {
		return nil
	}
	return out
}

// BuildSharedSessionBrowserWorkbenchSurface projects the stable workbench
// surface including its nested review surface.
func BuildSharedSessionBrowserWorkbenchSurface(
	req SharedSessionBrowserWorkbenchSurfaceRequest,
) *SharedSessionBrowserWorkbenchSurface {
	capability := normalizeSharedSessionBrowserCapabilitySurface(
		req.BrowserTools,
		req.ArtifactTools,
		req.ArtifactKinds,
		req.ArtifactContract,
		req.BrowserActKinds,
	)
	surface := &SharedSessionBrowserWorkbenchSurface{
		Ready:                     req.Ready,
		Sections:                  mergeSharedSessionBrowserGuidanceSections(nil, req.Sections),
		BrowserTools:              capability.BrowserTools,
		ArtifactTools:             capability.ArtifactTools,
		ArtifactKinds:             capability.ArtifactKinds,
		ArtifactContract:          capability.ArtifactContract,
		BrowserActKinds:           capability.BrowserActKinds,
		Explanation:               cloneSharedSessionBrowserSummary(req.Explanation),
		Diagnostics:               cloneSharedSessionBrowserSummary(req.Diagnostics),
		Summary:                   cloneSharedSessionBrowserSummary(req.Summary),
		PrimaryBrowserAction:      strings.TrimSpace(req.PrimaryBrowserAction),
		PrimaryNodeAction:         strings.TrimSpace(req.PrimaryNodeAction),
		NextStep:                  strings.TrimSpace(req.NextStep),
		RecommendedBrowserActions: cloneSharedSessionBrowserWorkbenchActions(req.RecommendedBrowserActions),
		RecommendedNodeActions:    cloneSharedSessionBrowserWorkbenchActions(req.RecommendedNodeActions),
	}
	surface.BrowserSurface, surface.BrowserOptInTargets = normalizeSharedSessionBrowserRouteSurface(
		req.BrowserSurface,
		req.BrowserOptInTargets,
	)
	if surface.Summary == nil {
		if surface.Diagnostics != nil {
			surface.Summary = cloneSharedSessionBrowserSummary(surface.Diagnostics)
		} else if surface.Explanation != nil {
			surface.Summary = cloneSharedSessionBrowserSummary(surface.Explanation)
		}
	}
	surface.Review = BuildSharedSessionBrowserReviewSurface(
		SharedSessionBrowserReviewSurfaceRequest{
			ReviewDecision: strings.TrimSpace(req.ReviewDecision),
			ReviewReady:    req.ReviewReady,
			Explanation:    surface.Explanation,
			Diagnostics:    surface.Diagnostics,
			Summary:        surface.Summary,
			Display:        sharedSessionBrowserDisplayFromSummary(surface.Summary),
		},
	)
	if sharedSessionBrowserWorkbenchSurfaceEmpty(*surface) {
		return nil
	}
	return surface
}

// BuildSharedSessionBrowserSurface projects the stable top-level surface from
// a display alias and optional review surface.
func BuildSharedSessionBrowserSurface(
	display *SharedSessionBrowserDisplay,
	review *SharedSessionBrowserReviewSurface,
) *SharedSessionBrowserSurface {
	return BuildSharedSessionBrowserSurfaceFromRequest(
		SharedSessionBrowserSurfaceRequest{
			Display: display,
			Review:  review,
		},
	)
}

// BuildSharedSessionBrowserDisplayFromRequest projects the stable top-level
// display alias from explicit display/workbench display first, then existing
// view/surface/review shells, and finally summary-style fallbacks.
func BuildSharedSessionBrowserDisplayFromRequest(
	req SharedSessionBrowserDisplayRequest,
) *SharedSessionBrowserDisplay {
	if display := cloneSharedSessionBrowserDisplay(req.Display); display != nil {
		return display
	}
	if display := cloneSharedSessionBrowserDisplay(req.WorkbenchDisplay); display != nil {
		return display
	}
	if display := sharedSessionBrowserDisplayFromView(req.View); display != nil {
		return display
	}
	if display := sharedSessionBrowserDisplayFromSurface(req.Surface); display != nil {
		return display
	}
	if req.Review != nil {
		if display := cloneSharedSessionBrowserDisplay(req.Review.Display); display != nil {
			return display
		}
		if display := sharedSessionBrowserDisplayFromSummary(req.Review.Summary); display != nil {
			return display
		}
		if display := sharedSessionBrowserDisplayFromSummary(req.Review.Diagnostics); display != nil {
			return display
		}
		if display := sharedSessionBrowserDisplayFromSummary(req.Review.Explanation); display != nil {
			return display
		}
	}
	if display := sharedSessionBrowserDisplayFromSummary(req.Summary); display != nil {
		return display
	}
	if display := sharedSessionBrowserDisplayFromSummary(req.Diagnostics); display != nil {
		return display
	}
	return sharedSessionBrowserDisplayFromSummary(req.Explanation)
}

// BuildSharedSessionBrowserReviewAliasFromRequest projects the stable
// top-level review alias from either an explicit review shell or nested
// view/workbench review posture.
func BuildSharedSessionBrowserReviewAliasFromRequest(
	req SharedSessionBrowserReviewAliasRequest,
) *SharedSessionBrowserReviewSurface {
	if review := cloneSharedSessionBrowserReviewSurface(req.Review); review != nil {
		return review
	}
	if req.View != nil {
		if review := cloneSharedSessionBrowserReviewSurface(req.View.Review); review != nil {
			return review
		}
	}
	if req.Workbench != nil {
		if review := cloneSharedSessionBrowserReviewSurface(req.Workbench.Review); review != nil {
			return review
		}
	}
	return nil
}

// BuildSharedSessionBrowserSummaryAliasFromRequest projects the stable
// top-level summary alias from explicit summary-like shells first, then
// review/display/surface/view/workbench fallbacks.
func BuildSharedSessionBrowserSummaryAliasFromRequest(
	req SharedSessionBrowserSummaryAliasRequest,
) *SharedSessionBrowserSummary {
	if summary := cloneSharedSessionBrowserSummary(req.Summary); summary != nil {
		return normalizeSharedSessionBrowserTopLevelSummary(summary)
	}
	if summary := cloneSharedSessionBrowserSummary(req.WorkbenchSummary); summary != nil {
		return normalizeSharedSessionBrowserTopLevelSummary(summary)
	}
	if summary := cloneSharedSessionBrowserSummary(req.Diagnostics); summary != nil {
		return normalizeSharedSessionBrowserTopLevelSummary(summary)
	}
	if summary := cloneSharedSessionBrowserSummary(req.Explanation); summary != nil {
		return normalizeSharedSessionBrowserTopLevelSummary(summary)
	}
	if summary := sharedSessionBrowserSummaryFromReview(req.Review); summary != nil {
		return normalizeSharedSessionBrowserTopLevelSummary(summary)
	}
	if summary := sharedSessionBrowserSummaryFromDisplay(req.Display); summary != nil {
		return normalizeSharedSessionBrowserTopLevelSummary(summary)
	}
	if summary := sharedSessionBrowserSummaryFromSurface(req.Surface); summary != nil {
		return normalizeSharedSessionBrowserTopLevelSummary(summary)
	}
	if summary := sharedSessionBrowserSummaryFromView(req.View); summary != nil {
		return normalizeSharedSessionBrowserTopLevelSummary(summary)
	}
	if summary := sharedSessionBrowserSummaryFromDisplay(req.WorkbenchDisplay); summary != nil {
		return normalizeSharedSessionBrowserTopLevelSummary(summary)
	}
	return normalizeSharedSessionBrowserTopLevelSummary(sharedSessionBrowserSummaryFromWorkbench(req.Workbench))
}

// BuildSharedSessionBrowserTopLevelAliasProjection projects the shared
// review->summary->display->surface->view chain while keeping the sequencing
// in browserruntime instead of tool-local orchestration.
func BuildSharedSessionBrowserTopLevelAliasProjection(
	req SharedSessionBrowserTopLevelAliasProjectionRequest,
) SharedSessionBrowserTopLevelAliasProjection {
	reviewReq := req.Review
	summaryReq := req.Summary
	displayReq := req.Display
	surfaceReq := req.Surface
	viewReq := req.View

	review := BuildSharedSessionBrowserReviewAliasFromRequest(reviewReq)
	if summaryReq.Review == nil {
		summaryReq.Review = review
	}
	summary := BuildSharedSessionBrowserSummaryAliasFromRequest(summaryReq)

	if displayReq.Review == nil {
		displayReq.Review = review
	}
	if displayReq.Summary == nil {
		displayReq.Summary = summary
	}
	display := BuildSharedSessionBrowserDisplayFromRequest(displayReq)

	if surfaceReq.Review == nil {
		surfaceReq.Review = review
	}
	if surfaceReq.Display == nil {
		surfaceReq.Display = display
	}
	surface := BuildSharedSessionBrowserSurfaceAliasFromRequest(surfaceReq)

	if viewReq.Review == nil {
		viewReq.Review = review
	}
	if viewReq.Surface == nil {
		viewReq.Surface = surface
	}
	view := BuildSharedSessionBrowserViewAliasFromRequest(viewReq)

	return SharedSessionBrowserTopLevelAliasProjection{
		Review:  review,
		Summary: summary,
		Display: display,
		Surface: surface,
		View:    view,
	}
}

// BuildSharedSessionBrowserRouteAliasFromRequest projects the stable
// top-level route-hint alias from explicit root metadata first, then nested
// surface/view/workbench shells.
func BuildSharedSessionBrowserRouteAliasFromRequest(
	req SharedSessionBrowserRouteAliasRequest,
) *SharedSessionBrowserRouteSurface {
	route := sharedSessionBrowserRouteSurfaceFromAliasRequest(req)
	if sharedSessionBrowserRouteSurfaceEmpty(route) {
		return nil
	}
	return &route
}

// BuildSharedSessionBrowserCapabilityAliasFromRequest projects the stable
// top-level capability-hint alias from explicit root metadata first, then
// nested surface/view/workbench shells for any remaining fields.
func BuildSharedSessionBrowserCapabilityAliasFromRequest(
	req SharedSessionBrowserCapabilityAliasRequest,
) *SharedSessionBrowserCapabilitySurface {
	capability := sharedSessionBrowserCapabilitySurfaceFromAliasRequest(req)
	if sharedSessionBrowserCapabilitySurfaceEmpty(capability) {
		return nil
	}
	return &capability
}

// BuildSharedSessionBrowserSurfaceAliasFromRequest projects the stable
// top-level surface alias from either an explicit surface shell or the shared
// display/review chain while applying route-surface precedence.
func BuildSharedSessionBrowserSurfaceAliasFromRequest(
	req SharedSessionBrowserSurfaceAliasRequest,
) *SharedSessionBrowserSurface {
	browserSurface, browserOptInTargets := sharedSessionBrowserSurfaceAliasRouteSurface(req)
	capability := sharedSessionBrowserSurfaceAliasCapabilitySurface(req)
	if req.Surface != nil {
		if surface := sharedSessionBrowserSurfaceWithRouteSurface(req.Surface, browserSurface, browserOptInTargets); surface != nil {
			surface = sharedSessionBrowserSurfaceWithCapabilitySurface(surface, capability)
			return surface
		}
	}
	if req.Display == nil && req.Review == nil {
		return nil
	}
	return BuildSharedSessionBrowserSurfaceFromRequest(
		SharedSessionBrowserSurfaceRequest{
			Display:             req.Display,
			Review:              req.Review,
			BrowserTools:        capability.BrowserTools,
			ArtifactTools:       capability.ArtifactTools,
			ArtifactKinds:       capability.ArtifactKinds,
			ArtifactContract:    capability.ArtifactContract,
			BrowserActKinds:     capability.BrowserActKinds,
			BrowserSurface:      browserSurface,
			BrowserOptInTargets: browserOptInTargets,
		},
	)
}

// BuildSharedSessionBrowserSurfaceFromRequest projects the stable top-level
// surface from display/review aliases plus optional route-surface diagnostics.
func BuildSharedSessionBrowserSurfaceFromRequest(
	req SharedSessionBrowserSurfaceRequest,
) *SharedSessionBrowserSurface {
	var out SharedSessionBrowserSurface
	capability := normalizeSharedSessionBrowserCapabilitySurface(
		req.BrowserTools,
		req.ArtifactTools,
		req.ArtifactKinds,
		req.ArtifactContract,
		req.BrowserActKinds,
	)
	if req.Display != nil {
		out.Ready = req.Display.Ready
		out.Sections = mergeSharedSessionBrowserGuidanceSections(nil, req.Display.Sections)
		out.SharedSessionBrowserSummary = req.Display.SharedSessionBrowserSummary
	}
	if req.Review != nil {
		out.ReviewPolicyState = strings.TrimSpace(req.Review.PolicyState)
		out.ReviewDecision = strings.TrimSpace(req.Review.Decision)
		out.ReviewReady = req.Review.Ready
		if strings.TrimSpace(out.Category) == "" {
			if req.Review.Display != nil {
				out.Ready = out.Ready || req.Review.Display.Ready
				if len(out.Sections) == 0 && len(req.Review.Display.Sections) > 0 {
					out.Sections = mergeSharedSessionBrowserGuidanceSections(nil, req.Review.Display.Sections)
				}
				out.SharedSessionBrowserSummary = req.Review.Display.SharedSessionBrowserSummary
			} else if req.Review.Summary != nil {
				if summaryDisplay := sharedSessionBrowserDisplayFromSummary(req.Review.Summary); summaryDisplay != nil {
					out.SharedSessionBrowserSummary = summaryDisplay.SharedSessionBrowserSummary
				}
			}
		}
	}
	out.BrowserTools = capability.BrowserTools
	out.ArtifactTools = capability.ArtifactTools
	out.ArtifactKinds = capability.ArtifactKinds
	out.ArtifactContract = capability.ArtifactContract
	out.BrowserActKinds = capability.BrowserActKinds
	out.BrowserSurface, out.BrowserOptInTargets = normalizeSharedSessionBrowserRouteSurface(req.BrowserSurface, req.BrowserOptInTargets)
	if sharedSessionBrowserSurfaceEmpty(out) {
		return nil
	}
	return &out
}

// BuildSharedSessionBrowserView projects the final view alias from either the
// workbench surface chain or the top-level result/review surface chain.
func BuildSharedSessionBrowserView(
	workbench *SharedSessionBrowserWorkbenchSurface,
	workbenchDisplay *SharedSessionBrowserDisplay,
	surface *SharedSessionBrowserSurface,
	review *SharedSessionBrowserReviewSurface,
) *SharedSessionBrowserView {
	return BuildSharedSessionBrowserViewFromRequest(
		SharedSessionBrowserViewRequest{
			Workbench:        workbench,
			WorkbenchDisplay: workbenchDisplay,
			Surface:          surface,
			Review:           review,
		},
	)
}

// BuildSharedSessionBrowserViewFromRequest projects the final view alias from
// either the workbench or result/review surface chain plus optional route
// diagnostics.
func BuildSharedSessionBrowserViewFromRequest(
	req SharedSessionBrowserViewRequest,
) *SharedSessionBrowserView {
	browserSurface, browserOptInTargets := sharedSessionBrowserViewRouteSurface(req)
	capability := sharedSessionBrowserViewCapabilitySurface(req)

	if req.Workbench != nil || req.WorkbenchDisplay != nil {
		view := &SharedSessionBrowserView{
			Kind:                "workbench",
			Review:              cloneSharedSessionBrowserReviewSurface(req.Review),
			BrowserTools:        cloneSharedSessionBrowserCapabilityItems(capability.BrowserTools),
			ArtifactTools:       cloneSharedSessionBrowserCapabilityItems(capability.ArtifactTools),
			ArtifactKinds:       cloneSharedSessionBrowserCapabilityItems(capability.ArtifactKinds),
			ArtifactContract:    strings.TrimSpace(capability.ArtifactContract),
			BrowserActKinds:     cloneSharedSessionBrowserCapabilityItems(capability.BrowserActKinds),
			BrowserSurface:      browserSurface,
			BrowserOptInTargets: cloneSharedSessionBrowserRouteTargets(browserOptInTargets),
		}
		if req.Workbench != nil {
			view.Ready = req.Workbench.Ready
			view.Sections = mergeSharedSessionBrowserGuidanceSections(nil, req.Workbench.Sections)
			if req.Workbench.Review != nil {
				view.Review = cloneSharedSessionBrowserReviewSurface(req.Workbench.Review)
			}
			view.PrimaryBrowserAction = strings.TrimSpace(req.Workbench.PrimaryBrowserAction)
			view.PrimaryNodeAction = strings.TrimSpace(req.Workbench.PrimaryNodeAction)
			view.NextStep = strings.TrimSpace(req.Workbench.NextStep)

			source := cloneSharedSessionBrowserSummary(req.Workbench.Summary)
			if source == nil {
				source = cloneSharedSessionBrowserSummary(req.Workbench.Diagnostics)
			}
			if source == nil {
				source = cloneSharedSessionBrowserSummary(req.Workbench.Explanation)
			}
			if source != nil {
				view.SharedSessionBrowserSummary = *source
				if view.PrimaryBrowserAction == "" {
					view.PrimaryBrowserAction = strings.TrimSpace(source.PrimaryBrowserAction)
				}
				if view.PrimaryNodeAction == "" {
					view.PrimaryNodeAction = strings.TrimSpace(source.PrimaryNodeAction)
				}
				if view.NextStep == "" {
					view.NextStep = strings.TrimSpace(source.NextStep)
				}
			}
		}
		if req.WorkbenchDisplay != nil {
			view.Ready = view.Ready || req.WorkbenchDisplay.Ready
			if len(view.Sections) == 0 && len(req.WorkbenchDisplay.Sections) > 0 {
				view.Sections = mergeSharedSessionBrowserGuidanceSections(nil, req.WorkbenchDisplay.Sections)
			}
			if strings.TrimSpace(view.Category) == "" {
				view.Category = strings.TrimSpace(req.WorkbenchDisplay.Category)
			}
			if strings.TrimSpace(view.State) == "" {
				view.State = strings.TrimSpace(req.WorkbenchDisplay.State)
			}
			if strings.TrimSpace(view.SummaryCode) == "" {
				view.SummaryCode = strings.TrimSpace(req.WorkbenchDisplay.SummaryCode)
			}
			if strings.TrimSpace(view.NextStepAlias) == "" {
				view.NextStepAlias = strings.TrimSpace(req.WorkbenchDisplay.NextStepAlias)
			}
			if strings.TrimSpace(view.ManualRetryHint) == "" {
				view.ManualRetryHint = strings.TrimSpace(req.WorkbenchDisplay.ManualRetryHint)
			}
			if !view.ResolvedViaFallback {
				view.ResolvedViaFallback = req.WorkbenchDisplay.ResolvedViaFallback
			}
			if view.PrimaryBrowserAction == "" {
				view.PrimaryBrowserAction = strings.TrimSpace(req.WorkbenchDisplay.PrimaryBrowserAction)
			}
			if view.PrimaryNodeAction == "" {
				view.PrimaryNodeAction = strings.TrimSpace(req.WorkbenchDisplay.PrimaryNodeAction)
			}
			if view.NextStep == "" {
				view.NextStep = strings.TrimSpace(req.WorkbenchDisplay.NextStep)
			}
		}
		if view.Review != nil {
			reviewDisplay := cloneSharedSessionBrowserDisplay(view.Review.Display)
			if reviewDisplay == nil && view.Review.Summary != nil {
				reviewDisplay = sharedSessionBrowserDisplayFromSummary(view.Review.Summary)
			}
			if reviewDisplay != nil {
				if view.Category == "" {
					view.Category = strings.TrimSpace(reviewDisplay.Category)
				}
				if view.State == "" {
					view.State = strings.TrimSpace(reviewDisplay.State)
				}
				if view.SummaryCode == "" {
					view.SummaryCode = strings.TrimSpace(reviewDisplay.SummaryCode)
				}
				if view.NextStepAlias == "" {
					view.NextStepAlias = strings.TrimSpace(reviewDisplay.NextStepAlias)
				}
				if view.ManualRetryHint == "" {
					view.ManualRetryHint = strings.TrimSpace(reviewDisplay.ManualRetryHint)
				}
				if !view.ResolvedViaFallback {
					view.ResolvedViaFallback = reviewDisplay.ResolvedViaFallback
				}
				if view.PrimaryBrowserAction == "" {
					view.PrimaryBrowserAction = strings.TrimSpace(reviewDisplay.PrimaryBrowserAction)
				}
				if view.PrimaryNodeAction == "" {
					view.PrimaryNodeAction = strings.TrimSpace(reviewDisplay.PrimaryNodeAction)
				}
				if view.NextStep == "" {
					view.NextStep = strings.TrimSpace(reviewDisplay.NextStep)
				}
			}
		}
		if sharedSessionBrowserViewEmpty(*view) {
			return nil
		}
		return view
	}

	if req.Surface == nil && req.Review == nil && browserSurface == "" && len(browserOptInTargets) == 0 {
		return nil
	}
	view := &SharedSessionBrowserView{
		Review:              cloneSharedSessionBrowserReviewSurface(req.Review),
		BrowserTools:        cloneSharedSessionBrowserCapabilityItems(capability.BrowserTools),
		ArtifactTools:       cloneSharedSessionBrowserCapabilityItems(capability.ArtifactTools),
		ArtifactKinds:       cloneSharedSessionBrowserCapabilityItems(capability.ArtifactKinds),
		ArtifactContract:    strings.TrimSpace(capability.ArtifactContract),
		BrowserActKinds:     cloneSharedSessionBrowserCapabilityItems(capability.BrowserActKinds),
		BrowserSurface:      browserSurface,
		BrowserOptInTargets: cloneSharedSessionBrowserRouteTargets(browserOptInTargets),
	}
	if req.Surface != nil {
		view.Kind = "result"
		view.Ready = req.Surface.Ready
		view.Sections = mergeSharedSessionBrowserGuidanceSections(nil, req.Surface.Sections)
		view.SharedSessionBrowserSummary = req.Surface.SharedSessionBrowserSummary
	}
	if strings.EqualFold(strings.TrimSpace(view.Category), "review") || req.Review != nil {
		view.Kind = "review"
	}
	if sharedSessionBrowserViewEmpty(*view) {
		return nil
	}
	return view
}

// BuildSharedSessionBrowserViewAliasFromRequest projects the final view alias
// from either an explicit view shell or the shared workbench/result chain while
// applying route-surface precedence.
func BuildSharedSessionBrowserViewAliasFromRequest(
	req SharedSessionBrowserViewAliasRequest,
) *SharedSessionBrowserView {
	browserSurface, browserOptInTargets := sharedSessionBrowserViewAliasRouteSurface(req)
	capability := sharedSessionBrowserViewAliasCapabilitySurface(req)
	if req.View != nil {
		if view := sharedSessionBrowserViewWithRouteSurface(req.View, browserSurface, browserOptInTargets); view != nil {
			view = sharedSessionBrowserViewWithCapabilitySurface(view, capability)
			return view
		}
	}
	if req.Workbench == nil && req.WorkbenchDisplay == nil && req.Surface == nil && req.Review == nil {
		return nil
	}
	return BuildSharedSessionBrowserViewFromRequest(
		SharedSessionBrowserViewRequest{
			Workbench:           req.Workbench,
			WorkbenchDisplay:    req.WorkbenchDisplay,
			Surface:             req.Surface,
			Review:              req.Review,
			BrowserTools:        capability.BrowserTools,
			ArtifactTools:       capability.ArtifactTools,
			ArtifactKinds:       capability.ArtifactKinds,
			ArtifactContract:    capability.ArtifactContract,
			BrowserActKinds:     capability.BrowserActKinds,
			BrowserSurface:      browserSurface,
			BrowserOptInTargets: browserOptInTargets,
		},
	)
}

func cloneSharedSessionBrowserSurface(
	surface *SharedSessionBrowserSurface,
) *SharedSessionBrowserSurface {
	if surface == nil {
		return nil
	}
	cloned := *surface
	cloned.Sections = mergeSharedSessionBrowserGuidanceSections(nil, surface.Sections)
	cloned.BrowserTools = cloneSharedSessionBrowserCapabilityItems(surface.BrowserTools)
	cloned.ArtifactTools = cloneSharedSessionBrowserCapabilityItems(surface.ArtifactTools)
	cloned.ArtifactKinds = cloneSharedSessionBrowserCapabilityItems(surface.ArtifactKinds)
	cloned.ArtifactContract = strings.TrimSpace(surface.ArtifactContract)
	cloned.BrowserActKinds = cloneSharedSessionBrowserCapabilityItems(surface.BrowserActKinds)
	cloned.BrowserOptInTargets = cloneSharedSessionBrowserRouteTargets(surface.BrowserOptInTargets)
	if sharedSessionBrowserSurfaceEmpty(cloned) {
		return nil
	}
	return &cloned
}

func normalizeSharedSessionBrowserTopLevelSummary(
	summary *SharedSessionBrowserSummary,
) *SharedSessionBrowserSummary {
	out := cloneSharedSessionBrowserSummary(summary)
	if out == nil {
		return nil
	}
	if !out.ResolvedViaFallback && strings.EqualFold(strings.TrimSpace(out.State), "resolved_via_fallback") {
		out.ResolvedViaFallback = true
	}
	if strings.TrimSpace(out.Category) == "" {
		if out.ResolvedViaFallback {
			out.Category = "resolver_fallback"
		} else if strings.TrimSpace(out.State) != "" ||
			strings.TrimSpace(out.SummaryCode) != "" ||
			strings.TrimSpace(out.NextStepAlias) != "" ||
			strings.TrimSpace(out.ManualRetryHint) != "" {
			out.Category = "resolver"
		}
	}
	if sharedSessionBrowserSummaryEmpty(*out) {
		return nil
	}
	return out
}

func cloneSharedSessionBrowserView(
	view *SharedSessionBrowserView,
) *SharedSessionBrowserView {
	if view == nil {
		return nil
	}
	cloned := *view
	cloned.Sections = mergeSharedSessionBrowserGuidanceSections(nil, view.Sections)
	cloned.BrowserTools = cloneSharedSessionBrowserCapabilityItems(view.BrowserTools)
	cloned.ArtifactTools = cloneSharedSessionBrowserCapabilityItems(view.ArtifactTools)
	cloned.ArtifactKinds = cloneSharedSessionBrowserCapabilityItems(view.ArtifactKinds)
	cloned.ArtifactContract = strings.TrimSpace(view.ArtifactContract)
	cloned.BrowserActKinds = cloneSharedSessionBrowserCapabilityItems(view.BrowserActKinds)
	cloned.BrowserOptInTargets = cloneSharedSessionBrowserRouteTargets(view.BrowserOptInTargets)
	cloned.Review = cloneSharedSessionBrowserReviewSurface(view.Review)
	if sharedSessionBrowserViewEmpty(cloned) {
		return nil
	}
	return &cloned
}

func cloneSharedSessionBrowserReviewSurface(
	review *SharedSessionBrowserReviewSurface,
) *SharedSessionBrowserReviewSurface {
	if review == nil {
		return nil
	}
	cloned := *review
	cloned.Explanation = cloneSharedSessionBrowserSummary(review.Explanation)
	cloned.Diagnostics = cloneSharedSessionBrowserSummary(review.Diagnostics)
	cloned.Summary = cloneSharedSessionBrowserSummary(review.Summary)
	cloned.Display = cloneSharedSessionBrowserDisplay(review.Display)
	if sharedSessionBrowserReviewSurfaceEmpty(cloned) {
		return nil
	}
	return &cloned
}

func sharedSessionBrowserReviewSurfaceEmpty(review SharedSessionBrowserReviewSurface) bool {
	return strings.TrimSpace(review.PolicyState) == "" &&
		strings.TrimSpace(review.Decision) == "" &&
		!review.Ready &&
		review.Explanation == nil &&
		review.Diagnostics == nil &&
		review.Summary == nil &&
		review.Display == nil
}

func sharedSessionBrowserWorkbenchSurfaceEmpty(surface SharedSessionBrowserWorkbenchSurface) bool {
	return !surface.Ready &&
		len(surface.Sections) == 0 &&
		surface.Review == nil &&
		len(surface.BrowserTools) == 0 &&
		len(surface.ArtifactTools) == 0 &&
		len(surface.ArtifactKinds) == 0 &&
		strings.TrimSpace(surface.ArtifactContract) == "" &&
		len(surface.BrowserActKinds) == 0 &&
		strings.TrimSpace(surface.BrowserSurface) == "" &&
		len(surface.BrowserOptInTargets) == 0 &&
		surface.Explanation == nil &&
		surface.Diagnostics == nil &&
		surface.Summary == nil &&
		strings.TrimSpace(surface.PrimaryBrowserAction) == "" &&
		strings.TrimSpace(surface.PrimaryNodeAction) == "" &&
		strings.TrimSpace(surface.NextStep) == "" &&
		len(surface.RecommendedBrowserActions) == 0 &&
		len(surface.RecommendedNodeActions) == 0
}

func sharedSessionBrowserSurfaceEmpty(surface SharedSessionBrowserSurface) bool {
	return !surface.Ready &&
		len(surface.Sections) == 0 &&
		sharedSessionBrowserSummaryEmpty(surface.SharedSessionBrowserSummary) &&
		strings.TrimSpace(surface.ReviewPolicyState) == "" &&
		strings.TrimSpace(surface.ReviewDecision) == "" &&
		!surface.ReviewReady &&
		len(surface.BrowserTools) == 0 &&
		len(surface.ArtifactTools) == 0 &&
		len(surface.ArtifactKinds) == 0 &&
		strings.TrimSpace(surface.ArtifactContract) == "" &&
		len(surface.BrowserActKinds) == 0 &&
		strings.TrimSpace(surface.BrowserSurface) == "" &&
		len(surface.BrowserOptInTargets) == 0
}

func sharedSessionBrowserViewEmpty(view SharedSessionBrowserView) bool {
	return strings.TrimSpace(view.Kind) == "" &&
		!view.Ready &&
		len(view.Sections) == 0 &&
		sharedSessionBrowserSummaryEmpty(view.SharedSessionBrowserSummary) &&
		view.Review == nil &&
		len(view.BrowserTools) == 0 &&
		len(view.ArtifactTools) == 0 &&
		len(view.ArtifactKinds) == 0 &&
		strings.TrimSpace(view.ArtifactContract) == "" &&
		len(view.BrowserActKinds) == 0 &&
		strings.TrimSpace(view.BrowserSurface) == "" &&
		len(view.BrowserOptInTargets) == 0
}

func sharedSessionBrowserViewRouteSurface(req SharedSessionBrowserViewRequest) (string, []string) {
	route := sharedSessionBrowserRouteSurfaceFromAliasRequest(
		SharedSessionBrowserRouteAliasRequest{
			Surface:             req.Surface,
			Workbench:           req.Workbench,
			BrowserSurface:      req.BrowserSurface,
			BrowserOptInTargets: req.BrowserOptInTargets,
		},
	)
	return route.BrowserSurface, route.BrowserOptInTargets
}

func sharedSessionBrowserViewCapabilitySurface(
	req SharedSessionBrowserViewRequest,
) sharedSessionBrowserCapabilitySurface {
	if capability := normalizeSharedSessionBrowserCapabilitySurface(
		req.BrowserTools,
		req.ArtifactTools,
		req.ArtifactKinds,
		req.ArtifactContract,
		req.BrowserActKinds,
	); !sharedSessionBrowserCapabilitySurfaceEmpty(capability) {
		return capability
	}
	if req.Surface != nil {
		if capability := sharedSessionBrowserCapabilitySurfaceFromSurface(req.Surface); !sharedSessionBrowserCapabilitySurfaceEmpty(capability) {
			return capability
		}
	}
	if req.Workbench != nil {
		return sharedSessionBrowserCapabilitySurfaceFromWorkbench(req.Workbench)
	}
	return sharedSessionBrowserCapabilitySurface{}
}

func sharedSessionBrowserSurfaceAliasRouteSurface(req SharedSessionBrowserSurfaceAliasRequest) (string, []string) {
	route := sharedSessionBrowserRouteSurfaceFromAliasRequest(
		SharedSessionBrowserRouteAliasRequest{
			Surface:             req.Surface,
			View:                req.View,
			BrowserSurface:      req.BrowserSurface,
			BrowserOptInTargets: req.BrowserOptInTargets,
		},
	)
	return route.BrowserSurface, route.BrowserOptInTargets
}

func sharedSessionBrowserSurfaceAliasCapabilitySurface(
	req SharedSessionBrowserSurfaceAliasRequest,
) sharedSessionBrowserCapabilitySurface {
	if capability := normalizeSharedSessionBrowserCapabilitySurface(
		req.BrowserTools,
		req.ArtifactTools,
		req.ArtifactKinds,
		req.ArtifactContract,
		req.BrowserActKinds,
	); !sharedSessionBrowserCapabilitySurfaceEmpty(capability) {
		return capability
	}
	if req.Surface != nil {
		if capability := sharedSessionBrowserCapabilitySurfaceFromSurface(req.Surface); !sharedSessionBrowserCapabilitySurfaceEmpty(capability) {
			return capability
		}
	}
	if req.View != nil {
		return sharedSessionBrowserCapabilitySurfaceFromView(req.View)
	}
	return sharedSessionBrowserCapabilitySurface{}
}

func sharedSessionBrowserViewAliasRouteSurface(req SharedSessionBrowserViewAliasRequest) (string, []string) {
	route := sharedSessionBrowserRouteSurfaceFromAliasRequest(
		SharedSessionBrowserRouteAliasRequest{
			Surface:             req.Surface,
			View:                req.View,
			BrowserSurface:      req.BrowserSurface,
			BrowserOptInTargets: req.BrowserOptInTargets,
		},
	)
	return route.BrowserSurface, route.BrowserOptInTargets
}

func sharedSessionBrowserViewAliasCapabilitySurface(
	req SharedSessionBrowserViewAliasRequest,
) sharedSessionBrowserCapabilitySurface {
	if capability := normalizeSharedSessionBrowserCapabilitySurface(
		req.BrowserTools,
		req.ArtifactTools,
		req.ArtifactKinds,
		req.ArtifactContract,
		req.BrowserActKinds,
	); !sharedSessionBrowserCapabilitySurfaceEmpty(capability) {
		return capability
	}
	if req.Surface != nil {
		if capability := sharedSessionBrowserCapabilitySurfaceFromSurface(req.Surface); !sharedSessionBrowserCapabilitySurfaceEmpty(capability) {
			return capability
		}
	}
	if req.View != nil {
		return sharedSessionBrowserCapabilitySurfaceFromView(req.View)
	}
	return sharedSessionBrowserCapabilitySurface{}
}

func sharedSessionBrowserDisplayFromSurface(surface *SharedSessionBrowserSurface) *SharedSessionBrowserDisplay {
	if surface == nil {
		return nil
	}
	display := &SharedSessionBrowserDisplay{
		Ready:                       surface.Ready,
		Sections:                    mergeSharedSessionBrowserGuidanceSections(nil, surface.Sections),
		SharedSessionBrowserSummary: surface.SharedSessionBrowserSummary,
	}
	if sharedSessionBrowserDisplayEmpty(*display) {
		return nil
	}
	return display
}

func sharedSessionBrowserSummaryFromDisplay(display *SharedSessionBrowserDisplay) *SharedSessionBrowserSummary {
	if display == nil {
		return nil
	}
	return cloneSharedSessionBrowserSummary(&display.SharedSessionBrowserSummary)
}

func sharedSessionBrowserSummaryFromSurface(surface *SharedSessionBrowserSurface) *SharedSessionBrowserSummary {
	if surface == nil {
		return nil
	}
	return cloneSharedSessionBrowserSummary(&surface.SharedSessionBrowserSummary)
}

func sharedSessionBrowserSummaryFromView(view *SharedSessionBrowserView) *SharedSessionBrowserSummary {
	if view == nil {
		return nil
	}
	return cloneSharedSessionBrowserSummary(&view.SharedSessionBrowserSummary)
}

func sharedSessionBrowserSummaryFromReview(review *SharedSessionBrowserReviewSurface) *SharedSessionBrowserSummary {
	if review == nil {
		return nil
	}
	if summary := cloneSharedSessionBrowserSummary(review.Summary); summary != nil {
		return summary
	}
	if summary := cloneSharedSessionBrowserSummary(review.Diagnostics); summary != nil {
		return summary
	}
	if summary := cloneSharedSessionBrowserSummary(review.Explanation); summary != nil {
		return summary
	}
	if display := cloneSharedSessionBrowserDisplay(review.Display); display != nil {
		return sharedSessionBrowserSummaryFromDisplay(display)
	}
	return nil
}

func sharedSessionBrowserSummaryFromWorkbench(workbench *SharedSessionBrowserWorkbenchSurface) *SharedSessionBrowserSummary {
	if workbench == nil {
		return nil
	}
	if summary := cloneSharedSessionBrowserSummary(workbench.Summary); summary != nil {
		return summary
	}
	if summary := cloneSharedSessionBrowserSummary(workbench.Diagnostics); summary != nil {
		return summary
	}
	return cloneSharedSessionBrowserSummary(workbench.Explanation)
}

func sharedSessionBrowserCapabilitySurfaceFromAliasRequest(
	req SharedSessionBrowserCapabilityAliasRequest,
) sharedSessionBrowserCapabilitySurface {
	capability := normalizeSharedSessionBrowserCapabilitySurface(
		req.BrowserTools,
		req.ArtifactTools,
		req.ArtifactKinds,
		req.ArtifactContract,
		req.BrowserActKinds,
	)
	if req.Surface != nil {
		capability = sharedSessionBrowserCapabilitySurfaceMerged(
			capability,
			sharedSessionBrowserCapabilitySurfaceFromSurface(req.Surface),
		)
	}
	if req.View != nil {
		capability = sharedSessionBrowserCapabilitySurfaceMerged(
			capability,
			sharedSessionBrowserCapabilitySurfaceFromView(req.View),
		)
	}
	if req.Workbench != nil {
		capability = sharedSessionBrowserCapabilitySurfaceMerged(
			capability,
			sharedSessionBrowserCapabilitySurfaceFromWorkbench(req.Workbench),
		)
	}
	return capability
}

func sharedSessionBrowserRouteSurfaceFromAliasRequest(
	req SharedSessionBrowserRouteAliasRequest,
) sharedSessionBrowserRouteSurface {
	if route := sharedSessionBrowserRouteSurfaceFromFields(
		req.BrowserSurface,
		req.BrowserOptInTargets,
	); !sharedSessionBrowserRouteSurfaceEmpty(route) {
		return route
	}
	if route := sharedSessionBrowserRouteSurfaceFromSurface(req.Surface); !sharedSessionBrowserRouteSurfaceEmpty(route) {
		return route
	}
	if route := sharedSessionBrowserRouteSurfaceFromView(req.View); !sharedSessionBrowserRouteSurfaceEmpty(route) {
		return route
	}
	return sharedSessionBrowserRouteSurfaceFromWorkbench(req.Workbench)
}

func sharedSessionBrowserRouteSurfaceFromFields(
	browserSurface string,
	browserOptInTargets []string,
) sharedSessionBrowserRouteSurface {
	label, targets := normalizeSharedSessionBrowserRouteSurface(browserSurface, browserOptInTargets)
	return sharedSessionBrowserRouteSurface{
		BrowserSurface:      label,
		BrowserOptInTargets: targets,
	}
}

func sharedSessionBrowserRouteSurfaceFromSurface(
	surface *SharedSessionBrowserSurface,
) sharedSessionBrowserRouteSurface {
	if surface == nil {
		return sharedSessionBrowserRouteSurface{}
	}
	return sharedSessionBrowserRouteSurfaceFromFields(surface.BrowserSurface, surface.BrowserOptInTargets)
}

func sharedSessionBrowserRouteSurfaceFromView(
	view *SharedSessionBrowserView,
) sharedSessionBrowserRouteSurface {
	if view == nil {
		return sharedSessionBrowserRouteSurface{}
	}
	return sharedSessionBrowserRouteSurfaceFromFields(view.BrowserSurface, view.BrowserOptInTargets)
}

func sharedSessionBrowserRouteSurfaceFromWorkbench(
	workbench *SharedSessionBrowserWorkbenchSurface,
) sharedSessionBrowserRouteSurface {
	if workbench == nil {
		return sharedSessionBrowserRouteSurface{}
	}
	return sharedSessionBrowserRouteSurfaceFromFields(workbench.BrowserSurface, workbench.BrowserOptInTargets)
}

func sharedSessionBrowserRouteSurfaceEmpty(route sharedSessionBrowserRouteSurface) bool {
	return strings.TrimSpace(route.BrowserSurface) == "" && len(route.BrowserOptInTargets) == 0
}

func sharedSessionBrowserCapabilitySurfaceFromSurface(
	surface *SharedSessionBrowserSurface,
) sharedSessionBrowserCapabilitySurface {
	if surface == nil {
		return sharedSessionBrowserCapabilitySurface{}
	}
	return normalizeSharedSessionBrowserCapabilitySurface(
		surface.BrowserTools,
		surface.ArtifactTools,
		surface.ArtifactKinds,
		surface.ArtifactContract,
		surface.BrowserActKinds,
	)
}

func sharedSessionBrowserCapabilitySurfaceFromView(
	view *SharedSessionBrowserView,
) sharedSessionBrowserCapabilitySurface {
	if view == nil {
		return sharedSessionBrowserCapabilitySurface{}
	}
	return normalizeSharedSessionBrowserCapabilitySurface(
		view.BrowserTools,
		view.ArtifactTools,
		view.ArtifactKinds,
		view.ArtifactContract,
		view.BrowserActKinds,
	)
}

func sharedSessionBrowserCapabilitySurfaceFromWorkbench(
	workbench *SharedSessionBrowserWorkbenchSurface,
) sharedSessionBrowserCapabilitySurface {
	if workbench == nil {
		return sharedSessionBrowserCapabilitySurface{}
	}
	return normalizeSharedSessionBrowserCapabilitySurface(
		workbench.BrowserTools,
		workbench.ArtifactTools,
		workbench.ArtifactKinds,
		workbench.ArtifactContract,
		workbench.BrowserActKinds,
	)
}

func sharedSessionBrowserCapabilitySurfaceMerged(
	primary sharedSessionBrowserCapabilitySurface,
	fallback sharedSessionBrowserCapabilitySurface,
) sharedSessionBrowserCapabilitySurface {
	out := normalizeSharedSessionBrowserCapabilitySurface(
		primary.BrowserTools,
		primary.ArtifactTools,
		primary.ArtifactKinds,
		primary.ArtifactContract,
		primary.BrowserActKinds,
	)
	if len(out.BrowserTools) == 0 {
		out.BrowserTools = cloneSharedSessionBrowserCapabilityItems(fallback.BrowserTools)
	}
	if len(out.ArtifactTools) == 0 {
		out.ArtifactTools = cloneSharedSessionBrowserCapabilityItems(fallback.ArtifactTools)
	}
	if len(out.ArtifactKinds) == 0 {
		out.ArtifactKinds = cloneSharedSessionBrowserCapabilityItems(fallback.ArtifactKinds)
	}
	if out.ArtifactContract == "" {
		out.ArtifactContract = strings.TrimSpace(fallback.ArtifactContract)
	}
	if len(out.BrowserActKinds) == 0 {
		out.BrowserActKinds = cloneSharedSessionBrowserCapabilityItems(fallback.BrowserActKinds)
	}
	return out
}

func sharedSessionBrowserSurfaceWithRouteSurface(
	surface *SharedSessionBrowserSurface,
	browserSurface string,
	browserOptInTargets []string,
) *SharedSessionBrowserSurface {
	label, targets := normalizeSharedSessionBrowserRouteSurface(browserSurface, browserOptInTargets)
	if surface == nil && label == "" && len(targets) == 0 {
		return nil
	}
	out := cloneSharedSessionBrowserSurface(surface)
	if out == nil {
		out = &SharedSessionBrowserSurface{}
	}
	out.BrowserSurface = label
	out.BrowserOptInTargets = cloneSharedSessionBrowserRouteTargets(targets)
	if sharedSessionBrowserSurfaceEmpty(*out) {
		return nil
	}
	return out
}

func sharedSessionBrowserSurfaceWithCapabilitySurface(
	surface *SharedSessionBrowserSurface,
	capability sharedSessionBrowserCapabilitySurface,
) *SharedSessionBrowserSurface {
	if surface == nil && sharedSessionBrowserCapabilitySurfaceEmpty(capability) {
		return nil
	}
	out := cloneSharedSessionBrowserSurface(surface)
	if out == nil {
		out = &SharedSessionBrowserSurface{}
	}
	out.BrowserTools = cloneSharedSessionBrowserCapabilityItems(capability.BrowserTools)
	out.ArtifactTools = cloneSharedSessionBrowserCapabilityItems(capability.ArtifactTools)
	out.ArtifactKinds = cloneSharedSessionBrowserCapabilityItems(capability.ArtifactKinds)
	out.ArtifactContract = strings.TrimSpace(capability.ArtifactContract)
	out.BrowserActKinds = cloneSharedSessionBrowserCapabilityItems(capability.BrowserActKinds)
	if sharedSessionBrowserSurfaceEmpty(*out) {
		return nil
	}
	return out
}

func sharedSessionBrowserViewWithRouteSurface(
	view *SharedSessionBrowserView,
	browserSurface string,
	browserOptInTargets []string,
) *SharedSessionBrowserView {
	label, targets := normalizeSharedSessionBrowserRouteSurface(browserSurface, browserOptInTargets)
	if view == nil && label == "" && len(targets) == 0 {
		return nil
	}
	out := cloneSharedSessionBrowserView(view)
	if out == nil {
		out = &SharedSessionBrowserView{}
	}
	out.BrowserSurface = label
	out.BrowserOptInTargets = cloneSharedSessionBrowserRouteTargets(targets)
	if sharedSessionBrowserViewEmpty(*out) {
		return nil
	}
	return out
}

func sharedSessionBrowserViewWithCapabilitySurface(
	view *SharedSessionBrowserView,
	capability sharedSessionBrowserCapabilitySurface,
) *SharedSessionBrowserView {
	if view == nil && sharedSessionBrowserCapabilitySurfaceEmpty(capability) {
		return nil
	}
	out := cloneSharedSessionBrowserView(view)
	if out == nil {
		out = &SharedSessionBrowserView{}
	}
	out.BrowserTools = cloneSharedSessionBrowserCapabilityItems(capability.BrowserTools)
	out.ArtifactTools = cloneSharedSessionBrowserCapabilityItems(capability.ArtifactTools)
	out.ArtifactKinds = cloneSharedSessionBrowserCapabilityItems(capability.ArtifactKinds)
	out.ArtifactContract = strings.TrimSpace(capability.ArtifactContract)
	out.BrowserActKinds = cloneSharedSessionBrowserCapabilityItems(capability.BrowserActKinds)
	if sharedSessionBrowserViewEmpty(*out) {
		return nil
	}
	return out
}

func sharedSessionBrowserDisplayFromView(view *SharedSessionBrowserView) *SharedSessionBrowserDisplay {
	if view == nil {
		return nil
	}
	display := &SharedSessionBrowserDisplay{
		Ready:                       view.Ready,
		Sections:                    mergeSharedSessionBrowserGuidanceSections(nil, view.Sections),
		SharedSessionBrowserSummary: view.SharedSessionBrowserSummary,
	}
	if sharedSessionBrowserDisplayEmpty(*display) {
		return nil
	}
	return display
}

func normalizeSharedSessionBrowserRouteSurface(browserSurface string, browserOptInTargets []string) (string, []string) {
	return strings.TrimSpace(browserSurface), cloneSharedSessionBrowserRouteTargets(browserOptInTargets)
}

func normalizeSharedSessionBrowserCapabilitySurface(
	browserTools []string,
	artifactTools []string,
	artifactKinds []string,
	artifactContract string,
	browserActKinds []string,
) sharedSessionBrowserCapabilitySurface {
	return sharedSessionBrowserCapabilitySurface{
		BrowserTools:     cloneSharedSessionBrowserCapabilityItems(browserTools),
		ArtifactTools:    cloneSharedSessionBrowserCapabilityItems(artifactTools),
		ArtifactKinds:    cloneSharedSessionBrowserCapabilityItems(artifactKinds),
		ArtifactContract: strings.TrimSpace(artifactContract),
		BrowserActKinds:  cloneSharedSessionBrowserCapabilityItems(browserActKinds),
	}
}

func sharedSessionBrowserCapabilitySurfaceEmpty(capability sharedSessionBrowserCapabilitySurface) bool {
	return len(capability.BrowserTools) == 0 &&
		len(capability.ArtifactTools) == 0 &&
		len(capability.ArtifactKinds) == 0 &&
		strings.TrimSpace(capability.ArtifactContract) == "" &&
		len(capability.BrowserActKinds) == 0
}

func cloneSharedSessionBrowserCapabilityItems(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(items))
	for _, raw := range items {
		value := strings.TrimSpace(raw)
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

func cloneSharedSessionBrowserRouteTargets(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(items))
	for _, raw := range items {
		value := strings.TrimSpace(raw)
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

func sharedSessionBrowserSummaryIsReview(summary *SharedSessionBrowserSummary) bool {
	return summary != nil && strings.EqualFold(strings.TrimSpace(summary.Category), "review")
}

func sharedSessionBrowserReviewSummaryCode(reviewDecision string) string {
	decision := strings.TrimSpace(reviewDecision)
	if decision == "" {
		return ""
	}
	if !strings.HasSuffix(decision, "_review_required") {
		return decision
	}
	base := strings.TrimSuffix(decision, "_review_required")
	base = strings.TrimPrefix(base, "session_target_")
	base = strings.TrimPrefix(base, "navigate_")
	if base == "" {
		return decision
	}
	return base + "_review_required"
}
