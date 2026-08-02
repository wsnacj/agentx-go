package tools

import (
	"encoding/json"
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

// browserRuntimeTopLevelAliasInputs keeps the decoded local alias inputs close
// to the surface projection bridge so unified browser only orchestrates payload
// mutation order instead of owning decode/bridge details.
type browserRuntimeTopLevelAliasInputs struct {
	workbenchExplanation   *browserTopLevelSummary
	diagnosticsExplanation *browserTopLevelSummary
	resolverExplanation    *browserTopLevelSummary
	workbenchDiagnostics   *browserTopLevelSummary
	explanation            *browserTopLevelSummary
	diagnostics            *browserTopLevelSummary
	summary                *browserTopLevelSummary
	workbenchSummary       *browserTopLevelSummary
	review                 *browserReviewSurfaceSummary
	display                *browserTopLevelDisplaySummary
	workbenchDisplay       *browserRuntimeWorkbenchDisplaySummary
	surface                *browserTopLevelSurfaceSummary
	view                   *browserTopLevelViewSummary
	workbench              *browserRuntimeWorkbenchSurfaceSummary
	defaultCandidateRoute  browserRuntimeRouteDescriptor
	browserTools           []string
	artifactTools          []string
	artifactKinds          []string
	artifactContract       string
	browserActKinds        []string
	browserSurface         string
	browserOptInTargets    []string
}

type browserRuntimeTopLevelRouteCapabilityAliasProjection struct {
	Route               browserTopLevelRouteSurface
	Capability          browserTopLevelCapabilitySurface
	BrowserSurface      json.RawMessage
	BrowserOptInTargets json.RawMessage
	BrowserTools        json.RawMessage
	ArtifactTools       json.RawMessage
	ArtifactKinds       json.RawMessage
	ArtifactContract    json.RawMessage
	BrowserActKinds     json.RawMessage
}

type browserRuntimeTopLevelDefaultCandidateAliasProjection struct {
	Route            browserRuntimeRouteDescriptor
	DefaultCandidate json.RawMessage
	Explanation      *browserTopLevelSummary
	Diagnostics      *browserTopLevelSummary
	Summary          *browserTopLevelSummary
	Display          *browserTopLevelDisplaySummary
	WorkbenchSummary *browserTopLevelSummary
	Review           *browserReviewSurfaceSummary
	WorkbenchDisplay *browserRuntimeWorkbenchDisplaySummary
	Workbench        *browserRuntimeWorkbenchSurfaceSummary
}

type browserRuntimeTopLevelAliasProjection struct {
	Explanation      json.RawMessage
	ExplanationValue *browserTopLevelSummary
	Diagnostics      json.RawMessage
	DiagnosticsValue *browserTopLevelSummary
	Summary          json.RawMessage
	SummaryValue     *browserTopLevelSummary
	Display          json.RawMessage
	DisplayValue     *browserTopLevelDisplaySummary
	Surface          json.RawMessage
	SurfaceValue     *browserTopLevelSurfaceSummary
	ApplySurface     bool
	View             json.RawMessage
	ViewValue        *browserTopLevelViewSummary
	ApplyView        bool
}

func browserRuntimeTopLevelAliasInputsFromPayload(payload map[string]json.RawMessage) browserRuntimeTopLevelAliasInputs {
	decoder := newBrowserRuntimeTopLevelAliasPayloadDecoder(payload)
	inputs := browserRuntimeTopLevelAliasInputs{
		workbenchExplanation:   decoder.runtimeDiagnosticsExplanationSummary("workbench_explanation"),
		diagnosticsExplanation: decoder.diagnosticsExplanationSummary("diagnostics_explanation"),
		resolverExplanation:    decoder.resolverExplanationSummary("resolver_explanation"),
		workbenchDiagnostics:   decoder.workbenchDiagnosticsSummary("workbench_diagnostics"),
		browserTools:           decoder.stringSlice("browser_tools"),
		artifactTools:          decoder.stringSlice("artifact_tools"),
		artifactKinds:          decoder.stringSlice("artifact_kinds"),
		artifactContract:       decoder.stringValue("artifact_contract"),
		browserActKinds:        decoder.stringSlice("browser_act_kinds"),
		defaultCandidateRoute:  decoder.routeDescriptor("default_candidate_route"),
		browserSurface:         decoder.stringValue("browser_surface"),
		browserOptInTargets:    decoder.stringSlice("browser_opt_in_targets"),
	}
	inputs.explanation = decoder.summary("explanation")
	inputs.diagnostics = decoder.summary("diagnostics")
	inputs.summary = decoder.summary("summary")
	inputs.workbenchSummary = decoder.summary("workbench_summary")
	inputs.review = decoder.review("review")
	inputs.display = decoder.display("display")
	inputs.workbenchDisplay = decoder.workbenchDisplay("workbench_display")
	inputs.surface = decoder.surface("surface")
	inputs.view = decoder.view("view")
	if workbench := decoder.workbench("workbench"); workbench != nil {
		inputs.workbench = workbench
		if strings.TrimSpace(inputs.browserSurface) == "" && len(inputs.browserOptInTargets) == 0 {
			inputs.browserSurface = strings.TrimSpace(workbench.BrowserSurface)
			inputs.browserOptInTargets = append([]string(nil), workbench.BrowserOptInTargets...)
		}
	}
	if inputs.defaultCandidateRoute == (browserRuntimeRouteDescriptor{}) {
		inputs.defaultCandidateRoute = browserRuntimeTopLevelAliasDefaultCandidateRoute(inputs)
	}
	return inputs
}

func browserRuntimeTopLevelAliasDefaultCandidateRoute(inputs browserRuntimeTopLevelAliasInputs) browserRuntimeRouteDescriptor {
	for _, route := range []browserRuntimeRouteDescriptor{
		browserRuntimeRouteDescriptorFromTopLevelSummary(inputs.explanation),
		browserRuntimeRouteDescriptorFromTopLevelSummary(inputs.diagnostics),
		browserRuntimeRouteDescriptorFromTopLevelSummary(inputs.summary),
		browserRuntimeRouteDescriptorFromTopLevelSummary(inputs.workbenchSummary),
		browserRuntimeRouteDescriptorFromReviewSurface(inputs.review),
		browserRuntimeRouteDescriptorFromTopLevelDisplay(inputs.display),
		browserRuntimeRouteDescriptorFromWorkbenchDisplay(inputs.workbenchDisplay),
		browserRuntimeRouteDescriptorFromTopLevelSurface(inputs.surface),
		browserRuntimeRouteDescriptorFromTopLevelView(inputs.view),
		browserRuntimeRouteDescriptorFromWorkbenchSurface(inputs.workbench),
	} {
		if route != (browserRuntimeRouteDescriptor{}) {
			return route
		}
	}
	return browserRuntimeRouteDescriptor{}
}

func browserRuntimeRouteDescriptorFromTopLevelSummary(summary *browserTopLevelSummary) browserRuntimeRouteDescriptor {
	if summary == nil {
		return browserRuntimeRouteDescriptor{}
	}
	return summary.DefaultCandidateRoute
}

func browserRuntimeRouteDescriptorFromTopLevelDisplay(summary *browserTopLevelDisplaySummary) browserRuntimeRouteDescriptor {
	if summary == nil {
		return browserRuntimeRouteDescriptor{}
	}
	return summary.DefaultCandidateRoute
}

func browserRuntimeRouteDescriptorFromTopLevelSurface(summary *browserTopLevelSurfaceSummary) browserRuntimeRouteDescriptor {
	if summary == nil {
		return browserRuntimeRouteDescriptor{}
	}
	return summary.DefaultCandidateRoute
}

func browserRuntimeRouteDescriptorFromTopLevelView(summary *browserTopLevelViewSummary) browserRuntimeRouteDescriptor {
	if summary == nil {
		return browserRuntimeRouteDescriptor{}
	}
	if summary.DefaultCandidateRoute != (browserRuntimeRouteDescriptor{}) {
		return summary.DefaultCandidateRoute
	}
	return browserRuntimeRouteDescriptorFromReviewSurface(summary.Review)
}

func browserRuntimeRouteDescriptorFromReviewSurface(summary *browserReviewSurfaceSummary) browserRuntimeRouteDescriptor {
	if summary == nil {
		return browserRuntimeRouteDescriptor{}
	}
	for _, route := range []browserRuntimeRouteDescriptor{
		browserRuntimeRouteDescriptorFromTopLevelSummary(summary.Explanation),
		browserRuntimeRouteDescriptorFromTopLevelSummary(summary.Diagnostics),
		browserRuntimeRouteDescriptorFromTopLevelSummary(summary.Summary),
		browserRuntimeRouteDescriptorFromTopLevelDisplay(summary.Display),
	} {
		if route != (browserRuntimeRouteDescriptor{}) {
			return route
		}
	}
	return browserRuntimeRouteDescriptor{}
}

func browserRuntimeRouteDescriptorFromWorkbenchDisplay(summary *browserRuntimeWorkbenchDisplaySummary) browserRuntimeRouteDescriptor {
	if summary == nil {
		return browserRuntimeRouteDescriptor{}
	}
	return summary.DefaultCandidateRoute
}

func browserRuntimeRouteDescriptorFromWorkbenchSurface(summary *browserRuntimeWorkbenchSurfaceSummary) browserRuntimeRouteDescriptor {
	if summary == nil {
		return browserRuntimeRouteDescriptor{}
	}
	for _, route := range []browserRuntimeRouteDescriptor{
		summary.DefaultCandidateRoute,
		browserRuntimeRouteDescriptorFromTopLevelSummary(summary.Explanation),
		browserRuntimeRouteDescriptorFromTopLevelSummary(summary.Diagnostics),
		browserRuntimeRouteDescriptorFromTopLevelSummary(summary.Summary),
		browserRuntimeRouteDescriptorFromReviewSurface(summary.Review),
	} {
		if route != (browserRuntimeRouteDescriptor{}) {
			return route
		}
	}
	return browserRuntimeRouteDescriptor{}
}

func (inputs *browserRuntimeTopLevelAliasInputs) setExplanation(summary *browserTopLevelSummary) {
	inputs.explanation = browserCloneTopLevelSummary(summary)
}

func (inputs *browserRuntimeTopLevelAliasInputs) setDiagnostics(summary *browserTopLevelSummary) {
	inputs.diagnostics = browserCloneTopLevelSummary(summary)
}

func (inputs *browserRuntimeTopLevelAliasInputs) setSummary(summary *browserTopLevelSummary) {
	inputs.summary = browserCloneTopLevelSummary(summary)
}

func (inputs *browserRuntimeTopLevelAliasInputs) setDisplay(display *browserTopLevelDisplaySummary) {
	inputs.display = browserCloneTopLevelDisplaySummary(display)
}

func (inputs *browserRuntimeTopLevelAliasInputs) setSurface(surface *browserTopLevelSurfaceSummary) {
	inputs.surface = browserCloneTopLevelSurfaceSummary(surface)
}

func (inputs *browserRuntimeTopLevelAliasInputs) setView(view *browserTopLevelViewSummary) {
	inputs.view = browserCloneTopLevelViewSummary(view)
}

func (inputs *browserRuntimeTopLevelAliasInputs) setRouteSurface(route browserTopLevelRouteSurface) {
	inputs.browserSurface = strings.TrimSpace(route.BrowserSurface)
	inputs.browserOptInTargets = append([]string(nil), route.BrowserOptInTargets...)
}

func (inputs *browserRuntimeTopLevelAliasInputs) setCapabilitySurface(capability browserTopLevelCapabilitySurface) {
	inputs.browserTools = append([]string(nil), capability.BrowserTools...)
	inputs.artifactTools = append([]string(nil), capability.ArtifactTools...)
	inputs.artifactKinds = append([]string(nil), capability.ArtifactKinds...)
	inputs.artifactContract = strings.TrimSpace(capability.ArtifactContract)
	inputs.browserActKinds = append([]string(nil), capability.BrowserActKinds...)
}

func (inputs browserRuntimeTopLevelAliasInputs) sharedReviewAliasRequest() agentxbrowserruntime.SharedSessionBrowserReviewAliasRequest {
	return agentxbrowserruntime.SharedSessionBrowserReviewAliasRequest{
		Review:    browserRuntimeSharedReviewSurfacePtr(inputs.review),
		View:      browserRuntimeSharedViewPtr(inputs.view),
		Workbench: browserRuntimeSharedWorkbenchSurfacePtr(inputs.workbench),
	}
}

func (inputs browserRuntimeTopLevelAliasInputs) sharedSummaryAliasRequest(
	review *agentxbrowserruntime.SharedSessionBrowserReviewSurface,
) agentxbrowserruntime.SharedSessionBrowserSummaryAliasRequest {
	return agentxbrowserruntime.SharedSessionBrowserSummaryAliasRequest{
		Summary:          browserRuntimeSharedSummaryPtr(inputs.summary),
		WorkbenchSummary: browserRuntimeSharedSummaryPtr(inputs.workbenchSummary),
		Diagnostics:      browserRuntimeSharedSummaryPtr(inputs.diagnostics),
		Explanation:      browserRuntimeSharedSummaryPtr(inputs.explanation),
		Review:           review,
		Display:          browserRuntimeSharedDisplayPtr(inputs.display),
		Surface:          browserRuntimeSharedSurfacePtr(inputs.surface),
		View:             browserRuntimeSharedViewPtr(inputs.view),
		WorkbenchDisplay: browserRuntimeSharedWorkbenchDisplayPtr(inputs.workbenchDisplay),
		Workbench:        browserRuntimeSharedWorkbenchSurfacePtr(inputs.workbench),
	}
}

func (inputs browserRuntimeTopLevelAliasInputs) sharedExplanationAliasRequest() agentxbrowserruntime.SharedSessionBrowserExplanationAliasRequest {
	return agentxbrowserruntime.SharedSessionBrowserExplanationAliasRequest{
		Workbench:   browserRuntimeSharedSummaryPtr(inputs.workbenchExplanation),
		Diagnostics: browserRuntimeSharedSummaryPtr(inputs.diagnosticsExplanation),
		Resolver:    browserRuntimeSharedSummaryPtr(inputs.resolverExplanation),
	}
}

func (inputs browserRuntimeTopLevelAliasInputs) sharedDiagnosticsAliasRequest() agentxbrowserruntime.SharedSessionBrowserDiagnosticsAliasRequest {
	return agentxbrowserruntime.SharedSessionBrowserDiagnosticsAliasRequest{
		Workbench:   browserRuntimeSharedSummaryPtr(inputs.workbenchDiagnostics),
		Diagnostics: browserRuntimeSharedSummaryPtr(inputs.diagnosticsExplanation),
		Summary:     browserRuntimeSharedSummaryPtr(inputs.summary),
	}
}

func (inputs browserRuntimeTopLevelAliasInputs) sharedTopLevelAliasProjectionRequest() agentxbrowserruntime.SharedSessionBrowserTopLevelAliasProjectionRequest {
	return agentxbrowserruntime.SharedSessionBrowserTopLevelAliasProjectionRequest{
		Review:  inputs.sharedReviewAliasRequest(),
		Summary: inputs.sharedSummaryAliasRequest(nil),
		Display: inputs.sharedDisplayRequest(nil, nil),
		Surface: inputs.sharedSurfaceAliasRequest(nil, nil),
		View:    inputs.sharedViewAliasRequest(nil, nil),
	}
}

func (inputs browserRuntimeTopLevelAliasInputs) sharedDisplayRequest(
	review *agentxbrowserruntime.SharedSessionBrowserReviewSurface,
	summary *agentxbrowserruntime.SharedSessionBrowserSummary,
) agentxbrowserruntime.SharedSessionBrowserDisplayRequest {
	return agentxbrowserruntime.SharedSessionBrowserDisplayRequest{
		Display:          browserRuntimeSharedDisplayPtr(inputs.display),
		WorkbenchDisplay: browserRuntimeSharedWorkbenchDisplayPtr(inputs.workbenchDisplay),
		Surface:          browserRuntimeSharedSurfacePtr(inputs.surface),
		View:             browserRuntimeSharedViewPtr(inputs.view),
		Review:           review,
		Summary:          summary,
	}
}

func (inputs browserRuntimeTopLevelAliasInputs) sharedSurfaceAliasRequest(
	display *agentxbrowserruntime.SharedSessionBrowserDisplay,
	review *agentxbrowserruntime.SharedSessionBrowserReviewSurface,
) agentxbrowserruntime.SharedSessionBrowserSurfaceAliasRequest {
	return agentxbrowserruntime.SharedSessionBrowserSurfaceAliasRequest{
		Surface:             browserRuntimeSharedSurfacePtr(inputs.surface),
		Display:             display,
		Review:              review,
		View:                browserRuntimeSharedViewPtr(inputs.view),
		BrowserTools:        append([]string(nil), inputs.browserTools...),
		ArtifactTools:       append([]string(nil), inputs.artifactTools...),
		ArtifactKinds:       append([]string(nil), inputs.artifactKinds...),
		ArtifactContract:    strings.TrimSpace(inputs.artifactContract),
		BrowserActKinds:     append([]string(nil), inputs.browserActKinds...),
		BrowserSurface:      strings.TrimSpace(inputs.browserSurface),
		BrowserOptInTargets: append([]string(nil), inputs.browserOptInTargets...),
	}
}

func (inputs browserRuntimeTopLevelAliasInputs) sharedViewAliasRequest(
	surface *agentxbrowserruntime.SharedSessionBrowserSurface,
	review *agentxbrowserruntime.SharedSessionBrowserReviewSurface,
) agentxbrowserruntime.SharedSessionBrowserViewAliasRequest {
	return agentxbrowserruntime.SharedSessionBrowserViewAliasRequest{
		View:                browserRuntimeSharedViewPtr(inputs.view),
		Workbench:           browserRuntimeSharedWorkbenchSurfacePtr(inputs.workbench),
		WorkbenchDisplay:    browserRuntimeSharedWorkbenchDisplayPtr(inputs.workbenchDisplay),
		Surface:             surface,
		Review:              review,
		BrowserTools:        append([]string(nil), inputs.browserTools...),
		ArtifactTools:       append([]string(nil), inputs.artifactTools...),
		ArtifactKinds:       append([]string(nil), inputs.artifactKinds...),
		ArtifactContract:    strings.TrimSpace(inputs.artifactContract),
		BrowserActKinds:     append([]string(nil), inputs.browserActKinds...),
		BrowserSurface:      strings.TrimSpace(inputs.browserSurface),
		BrowserOptInTargets: append([]string(nil), inputs.browserOptInTargets...),
	}
}

func (inputs browserRuntimeTopLevelAliasInputs) sharedCapabilityAliasRequest() agentxbrowserruntime.SharedSessionBrowserCapabilityAliasRequest {
	return agentxbrowserruntime.SharedSessionBrowserCapabilityAliasRequest{
		Surface:          browserRuntimeSharedSurfacePtr(inputs.surface),
		View:             browserRuntimeSharedViewPtr(inputs.view),
		Workbench:        browserRuntimeSharedWorkbenchSurfacePtr(inputs.workbench),
		BrowserTools:     append([]string(nil), inputs.browserTools...),
		ArtifactTools:    append([]string(nil), inputs.artifactTools...),
		ArtifactKinds:    append([]string(nil), inputs.artifactKinds...),
		ArtifactContract: strings.TrimSpace(inputs.artifactContract),
		BrowserActKinds:  append([]string(nil), inputs.browserActKinds...),
	}
}

func (inputs browserRuntimeTopLevelAliasInputs) sharedRouteAliasRequest() agentxbrowserruntime.SharedSessionBrowserRouteAliasRequest {
	return agentxbrowserruntime.SharedSessionBrowserRouteAliasRequest{
		Surface:             browserRuntimeSharedSurfacePtr(inputs.surface),
		View:                browserRuntimeSharedViewPtr(inputs.view),
		Workbench:           browserRuntimeSharedWorkbenchSurfacePtr(inputs.workbench),
		BrowserSurface:      strings.TrimSpace(inputs.browserSurface),
		BrowserOptInTargets: append([]string(nil), inputs.browserOptInTargets...),
	}
}

func browserRuntimeBuildTopLevelSummaryAlias(inputs browserRuntimeTopLevelAliasInputs) (json.RawMessage, *browserTopLevelSummary, bool) {
	summary, ok := browserRuntimeTopLevelSummaryFromInputs(inputs)
	if !ok {
		return nil, nil, false
	}
	blob, err := json.Marshal(summary)
	if err != nil {
		return nil, nil, false
	}
	return blob, &summary, true
}

func browserRuntimeBuildTopLevelDisplayAlias(inputs browserRuntimeTopLevelAliasInputs) (json.RawMessage, *browserTopLevelDisplaySummary, bool) {
	display, ok := browserRuntimeTopLevelDisplayFromInputs(inputs)
	if !ok {
		return nil, nil, false
	}
	blob, err := json.Marshal(display)
	if err != nil {
		return nil, nil, false
	}
	return blob, display, true
}

func browserRuntimeBuildTopLevelSurfaceAlias(inputs browserRuntimeTopLevelAliasInputs) (json.RawMessage, *browserTopLevelSurfaceSummary, bool) {
	surface, ok := browserRuntimeTopLevelSurfaceFromInputs(inputs)
	if !ok {
		return nil, nil, false
	}
	blob, err := json.Marshal(surface)
	if err != nil {
		return nil, nil, false
	}
	return blob, surface, true
}

func browserRuntimeBuildTopLevelViewAlias(inputs browserRuntimeTopLevelAliasInputs) (json.RawMessage, *browserTopLevelViewSummary, bool) {
	view, ok := browserRuntimeTopLevelViewFromInputs(inputs)
	if !ok {
		return nil, nil, false
	}
	blob, err := json.Marshal(view)
	if err != nil {
		return nil, nil, false
	}
	return blob, view, true
}

func browserRuntimeBuildTopLevelExplanationAlias(inputs browserRuntimeTopLevelAliasInputs) (json.RawMessage, *browserTopLevelSummary, bool) {
	summary, ok := browserRuntimeTopLevelExplanationAliasFromInputs(inputs)
	if !ok {
		return nil, nil, false
	}
	blob, ok := browserRuntimeMarshalTopLevelSummaryAlias(summary)
	if !ok {
		return nil, nil, false
	}
	return blob, summary, true
}

func browserRuntimeBuildTopLevelDiagnosticsAlias(inputs browserRuntimeTopLevelAliasInputs) (json.RawMessage, *browserTopLevelSummary, bool) {
	summary, ok := browserRuntimeTopLevelDiagnosticsAliasFromInputs(inputs)
	if !ok {
		return nil, nil, false
	}
	blob, ok := browserRuntimeMarshalTopLevelSummaryAlias(summary)
	if !ok {
		return nil, nil, false
	}
	return blob, summary, true
}

func browserRuntimeTopLevelExplanationAliasFromInputs(inputs browserRuntimeTopLevelAliasInputs) (*browserTopLevelSummary, bool) {
	summary := browserRuntimeTopLevelSummaryFromShared(
		agentxbrowserruntime.BuildSharedSessionBrowserExplanationAliasFromRequest(
			inputs.sharedExplanationAliasRequest(),
		),
	)
	summary = browserTopLevelSummaryWithDefaultCandidateRoute(summary, inputs.defaultCandidateRoute)
	return summary, summary != nil
}

func browserRuntimeTopLevelDiagnosticsAliasFromInputs(inputs browserRuntimeTopLevelAliasInputs) (*browserTopLevelSummary, bool) {
	summary := browserRuntimeTopLevelSummaryFromShared(
		agentxbrowserruntime.BuildSharedSessionBrowserDiagnosticsAliasFromRequest(
			inputs.sharedDiagnosticsAliasRequest(),
		),
	)
	summary = browserTopLevelSummaryWithDefaultCandidateRoute(summary, inputs.defaultCandidateRoute)
	return summary, summary != nil
}

func browserRuntimeTopLevelSummaryFromInputs(inputs browserRuntimeTopLevelAliasInputs) (browserTopLevelSummary, bool) {
	summary, ok := browserRuntimeTopLevelSummarySourceFromInputs(inputs)
	if !ok {
		return browserTopLevelSummary{}, false
	}
	if out := browserTopLevelSummaryWithDefaultCandidateRoute(&summary, inputs.defaultCandidateRoute); out != nil {
		summary = *out
	}
	if command := browserRuntimeTopLevelAliasRepairCommand(inputs); command != "" && strings.TrimSpace(summary.RepairCommand) == "" {
		summary.RepairCommand = command
	}
	if browserUnifiedSummaryEmpty(summary) {
		return browserTopLevelSummary{}, false
	}
	return summary, true
}

func browserRuntimeTopLevelDisplayFromInputs(inputs browserRuntimeTopLevelAliasInputs) (*browserTopLevelDisplaySummary, bool) {
	display := browserRuntimeTopLevelDisplaySummaryFromShared(
		browserRuntimeSharedTopLevelAliasProjectionFromInputs(inputs).Display,
	)
	display = browserTopLevelDisplayWithDefaultCandidateRoute(display, inputs.defaultCandidateRoute)
	if command := browserRuntimeTopLevelAliasRepairCommand(inputs); command != "" {
		browserRuntimeApplyRepairCommandToTopLevelDisplay(display, command)
	}
	return display, display != nil
}

func browserRuntimeTopLevelSurfaceFromInputs(inputs browserRuntimeTopLevelAliasInputs) (*browserTopLevelSurfaceSummary, bool) {
	surface := browserRuntimeTopLevelSurfaceSummaryFromShared(
		browserRuntimeSharedTopLevelAliasProjectionFromInputs(inputs).Surface,
	)
	surface = browserTopLevelSurfaceWithDefaultCandidateRoute(surface, inputs.defaultCandidateRoute)
	if command := browserRuntimeTopLevelAliasRepairCommand(inputs); command != "" {
		browserRuntimeApplyRepairCommandToTopLevelSurface(surface, command)
	}
	return surface, surface != nil
}

func browserRuntimeTopLevelViewFromInputs(inputs browserRuntimeTopLevelAliasInputs) (*browserTopLevelViewSummary, bool) {
	view := browserRuntimeTopLevelViewSummaryFromShared(
		browserRuntimeSharedTopLevelAliasProjectionFromInputs(inputs).View,
	)
	view = browserTopLevelViewWithDefaultCandidateRoute(view, inputs.defaultCandidateRoute)
	if command := browserRuntimeTopLevelAliasRepairCommand(inputs); command != "" {
		browserRuntimeApplyRepairCommandToTopLevelView(view, command)
	}
	return view, view != nil
}

func browserRuntimeTopLevelSummarySourceFromInputs(inputs browserRuntimeTopLevelAliasInputs) (browserTopLevelSummary, bool) {
	summary := browserRuntimeTopLevelSummaryFromShared(
		browserRuntimeSharedTopLevelAliasProjectionFromInputs(inputs).Summary,
	)
	if summary != nil {
		return *summary, true
	}
	return browserTopLevelSummary{}, false
}

func browserRuntimeTopLevelReviewFromInputs(inputs browserRuntimeTopLevelAliasInputs) (browserReviewSurfaceSummary, bool) {
	review := browserRuntimeReviewSurfaceSummaryFromShared(
		browserRuntimeSharedTopLevelAliasProjectionFromInputs(inputs).Review,
	)
	review = browserReviewSurfaceSummaryWithDefaultCandidateRoute(review, inputs.defaultCandidateRoute)
	if command := browserRuntimeTopLevelAliasRepairCommand(inputs); command != "" {
		browserRuntimeApplyRepairCommandToReviewSurface(review, command)
	}
	if review != nil {
		return *review, true
	}
	return browserReviewSurfaceSummary{}, false
}

func browserRuntimeTopLevelAliasRepairCommand(inputs browserRuntimeTopLevelAliasInputs) string {
	for _, summary := range []*browserTopLevelSummary{
		inputs.summary,
		inputs.diagnostics,
		inputs.explanation,
		inputs.workbenchSummary,
		inputs.workbenchDiagnostics,
		inputs.workbenchExplanation,
		inputs.diagnosticsExplanation,
		inputs.resolverExplanation,
	} {
		if summary != nil {
			if command := strings.TrimSpace(summary.RepairCommand); command != "" {
				return command
			}
		}
	}
	if inputs.display != nil {
		if command := strings.TrimSpace(inputs.display.RepairCommand); command != "" {
			return command
		}
	}
	if inputs.workbenchDisplay != nil {
		if command := strings.TrimSpace(inputs.workbenchDisplay.RepairCommand); command != "" {
			return command
		}
	}
	if inputs.surface != nil {
		if command := strings.TrimSpace(inputs.surface.RepairCommand); command != "" {
			return command
		}
	}
	if inputs.view != nil {
		if command := strings.TrimSpace(inputs.view.RepairCommand); command != "" {
			return command
		}
	}
	if inputs.review != nil {
		for _, summary := range []*browserTopLevelSummary{
			inputs.review.Explanation,
			inputs.review.Diagnostics,
			inputs.review.Summary,
		} {
			if summary != nil {
				if command := strings.TrimSpace(summary.RepairCommand); command != "" {
					return command
				}
			}
		}
		if inputs.review.Display != nil {
			if command := strings.TrimSpace(inputs.review.Display.RepairCommand); command != "" {
				return command
			}
		}
	}
	if inputs.workbench != nil {
		if command := strings.TrimSpace(inputs.workbench.RepairCommand); command != "" {
			return command
		}
		for _, summary := range []*browserTopLevelSummary{
			inputs.workbench.Explanation,
			inputs.workbench.Diagnostics,
			inputs.workbench.Summary,
		} {
			if summary != nil {
				if command := strings.TrimSpace(summary.RepairCommand); command != "" {
					return command
				}
			}
		}
	}
	return ""
}

func browserRuntimeSharedTopLevelAliasProjectionFromInputs(
	inputs browserRuntimeTopLevelAliasInputs,
) agentxbrowserruntime.SharedSessionBrowserTopLevelAliasProjection {
	return agentxbrowserruntime.BuildSharedSessionBrowserTopLevelAliasProjection(
		inputs.sharedTopLevelAliasProjectionRequest(),
	)
}

func browserRuntimeTopLevelPayloadHasSurfaceMetadata(inputs browserRuntimeTopLevelAliasInputs) bool {
	return !browserTopLevelRouteSurfaceEmpty(browserRuntimeTopLevelRouteSurfaceFromInputs(inputs)) ||
		inputs.defaultCandidateRoute != (browserRuntimeRouteDescriptor{}) ||
		!browserTopLevelCapabilitySurfaceEmpty(browserRuntimeTopLevelCapabilitySurfaceFromInputs(inputs))
}

func browserRuntimeTopLevelRouteSurfaceFromInputs(inputs browserRuntimeTopLevelAliasInputs) browserTopLevelRouteSurface {
	return browserTopLevelRouteSurfaceFromShared(
		agentxbrowserruntime.BuildSharedSessionBrowserRouteAliasFromRequest(
			inputs.sharedRouteAliasRequest(),
		),
	)
}

func browserRuntimeTopLevelCapabilitySurfaceFromInputs(inputs browserRuntimeTopLevelAliasInputs) browserTopLevelCapabilitySurface {
	return browserTopLevelCapabilitySurfaceFromShared(
		agentxbrowserruntime.BuildSharedSessionBrowserCapabilityAliasFromRequest(
			inputs.sharedCapabilityAliasRequest(),
		),
	)
}

func browserRuntimeBuildTopLevelRouteCapabilityAliasProjection(
	inputs browserRuntimeTopLevelAliasInputs,
) (browserRuntimeTopLevelRouteCapabilityAliasProjection, bool) {
	projection := browserRuntimeTopLevelRouteCapabilityAliasProjection{}
	mutated := false
	if route := browserRuntimeTopLevelRouteSurfaceFromInputs(inputs); !browserTopLevelRouteSurfaceEmpty(route) {
		routeBlob, err := json.Marshal(route.BrowserSurface)
		if err != nil {
			return browserRuntimeTopLevelRouteCapabilityAliasProjection{}, false
		}
		targetsBlob, err := json.Marshal(route.BrowserOptInTargets)
		if err != nil {
			return browserRuntimeTopLevelRouteCapabilityAliasProjection{}, false
		}
		projection.Route = route
		projection.BrowserSurface = routeBlob
		projection.BrowserOptInTargets = targetsBlob
		mutated = true
	}
	if capability := browserRuntimeTopLevelCapabilitySurfaceFromInputs(inputs); !browserTopLevelCapabilitySurfaceEmpty(capability) {
		var err error
		projection.BrowserTools, err = json.Marshal(capability.BrowserTools)
		if err != nil {
			return browserRuntimeTopLevelRouteCapabilityAliasProjection{}, false
		}
		projection.ArtifactTools, err = json.Marshal(capability.ArtifactTools)
		if err != nil {
			return browserRuntimeTopLevelRouteCapabilityAliasProjection{}, false
		}
		projection.ArtifactKinds, err = json.Marshal(capability.ArtifactKinds)
		if err != nil {
			return browserRuntimeTopLevelRouteCapabilityAliasProjection{}, false
		}
		projection.ArtifactContract, err = json.Marshal(capability.ArtifactContract)
		if err != nil {
			return browserRuntimeTopLevelRouteCapabilityAliasProjection{}, false
		}
		projection.BrowserActKinds, err = json.Marshal(capability.BrowserActKinds)
		if err != nil {
			return browserRuntimeTopLevelRouteCapabilityAliasProjection{}, false
		}
		projection.Capability = capability
		mutated = true
	}
	return projection, mutated
}

func browserRuntimeApplyTopLevelRouteCapabilityAliasProjection(
	payload map[string]json.RawMessage,
	inputs *browserRuntimeTopLevelAliasInputs,
) (bool, error) {
	if payload == nil || inputs == nil {
		return false, nil
	}
	projection, ok := browserRuntimeBuildTopLevelRouteCapabilityAliasProjection(*inputs)
	if !ok {
		return false, nil
	}
	mutated := false
	if !browserTopLevelRouteSurfaceEmpty(projection.Route) {
		if !browserUnifiedHasNonNullJSONField(payload, "browser_surface") {
			payload["browser_surface"] = projection.BrowserSurface
			mutated = true
		}
		if !browserUnifiedHasNonNullJSONField(payload, "browser_opt_in_targets") {
			payload["browser_opt_in_targets"] = projection.BrowserOptInTargets
			mutated = true
		}
		inputs.setRouteSurface(projection.Route)
	}
	if !browserTopLevelCapabilitySurfaceEmpty(projection.Capability) {
		if !browserUnifiedHasNonNullJSONField(payload, "browser_tools") {
			payload["browser_tools"] = projection.BrowserTools
			mutated = true
		}
		if !browserUnifiedHasNonNullJSONField(payload, "artifact_tools") {
			payload["artifact_tools"] = projection.ArtifactTools
			mutated = true
		}
		if !browserUnifiedHasNonNullJSONField(payload, "artifact_kinds") {
			payload["artifact_kinds"] = projection.ArtifactKinds
			mutated = true
		}
		if !browserUnifiedHasNonNullJSONField(payload, "artifact_contract") {
			payload["artifact_contract"] = projection.ArtifactContract
			mutated = true
		}
		if !browserUnifiedHasNonNullJSONField(payload, "browser_act_kinds") {
			payload["browser_act_kinds"] = projection.BrowserActKinds
			mutated = true
		}
		inputs.setCapabilitySurface(projection.Capability)
	}
	return mutated, nil
}

func browserRuntimeBuildTopLevelDefaultCandidateAliasProjection(
	inputs browserRuntimeTopLevelAliasInputs,
) (browserRuntimeTopLevelDefaultCandidateAliasProjection, bool) {
	route := inputs.defaultCandidateRoute
	if route == (browserRuntimeRouteDescriptor{}) {
		return browserRuntimeTopLevelDefaultCandidateAliasProjection{}, false
	}
	routeBlob, err := json.Marshal(route)
	if err != nil {
		return browserRuntimeTopLevelDefaultCandidateAliasProjection{}, false
	}
	return browserRuntimeTopLevelDefaultCandidateAliasProjection{
		Route:            route,
		DefaultCandidate: routeBlob,
		Explanation:      browserTopLevelSummaryWithDefaultCandidateRoute(inputs.explanation, route),
		Diagnostics:      browserTopLevelSummaryWithDefaultCandidateRoute(inputs.diagnostics, route),
		Summary:          browserTopLevelSummaryWithDefaultCandidateRoute(inputs.summary, route),
		Display:          browserTopLevelDisplayWithDefaultCandidateRoute(inputs.display, route),
		WorkbenchSummary: browserTopLevelSummaryWithDefaultCandidateRoute(inputs.workbenchSummary, route),
		Review:           browserReviewSurfaceSummaryWithDefaultCandidateRoute(inputs.review, route),
		WorkbenchDisplay: browserWorkbenchDisplayWithDefaultCandidateRoute(inputs.workbenchDisplay, route),
		Workbench:        browserWorkbenchWithDefaultCandidateRoute(inputs.workbench, route),
	}, true
}

func browserRuntimeApplyTopLevelDefaultCandidateAliasProjection(
	payload map[string]json.RawMessage,
	inputs *browserRuntimeTopLevelAliasInputs,
) (bool, error) {
	if payload == nil || inputs == nil {
		return false, nil
	}
	projection, ok := browserRuntimeBuildTopLevelDefaultCandidateAliasProjection(*inputs)
	if !ok {
		return false, nil
	}
	mutated := false
	if !browserUnifiedHasNonNullJSONField(payload, "default_candidate_route") {
		payload["default_candidate_route"] = projection.DefaultCandidate
		mutated = true
	}
	writeSummary := func(key string, summary *browserTopLevelSummary, assign func(*browserTopLevelSummary), force bool) error {
		if summary == nil || (!force && browserUnifiedHasNonNullJSONField(payload, key)) {
			return nil
		}
		blob, err := json.Marshal(summary)
		if err != nil {
			return err
		}
		payload[key] = blob
		assign(summary)
		mutated = true
		return nil
	}
	if err := writeSummary("explanation", projection.Explanation, inputs.setExplanation, true); err != nil {
		return false, err
	}
	if err := writeSummary("diagnostics", projection.Diagnostics, inputs.setDiagnostics, true); err != nil {
		return false, err
	}
	if projection.Summary != nil &&
		(inputs.summary == nil || inputs.summary.DefaultCandidateRoute != projection.Route) {
		if err := writeSummary("summary", projection.Summary, inputs.setSummary, true); err != nil {
			return false, err
		}
	}
	if projection.Display != nil &&
		(inputs.display == nil || inputs.display.DefaultCandidateRoute != projection.Route) {
		blob, err := json.Marshal(projection.Display)
		if err != nil {
			return false, err
		}
		payload["display"] = blob
		inputs.setDisplay(projection.Display)
		mutated = true
	}
	if projection.WorkbenchSummary != nil &&
		(inputs.workbenchSummary == nil || inputs.workbenchSummary.DefaultCandidateRoute != projection.Route) {
		blob, err := json.Marshal(projection.WorkbenchSummary)
		if err != nil {
			return false, err
		}
		payload["workbench_summary"] = blob
		inputs.workbenchSummary = projection.WorkbenchSummary
		mutated = true
	}
	if projection.Review != nil {
		blob, err := json.Marshal(projection.Review)
		if err != nil {
			return false, err
		}
		payload["review"] = blob
		inputs.review = projection.Review
		mutated = true
	}
	if projection.WorkbenchDisplay != nil &&
		(inputs.workbenchDisplay == nil || inputs.workbenchDisplay.DefaultCandidateRoute != projection.Route) {
		blob, err := json.Marshal(projection.WorkbenchDisplay)
		if err != nil {
			return false, err
		}
		payload["workbench_display"] = blob
		inputs.workbenchDisplay = projection.WorkbenchDisplay
		mutated = true
	}
	if projection.Workbench != nil {
		blob, err := json.Marshal(projection.Workbench)
		if err != nil {
			return false, err
		}
		payload["workbench"] = blob
		inputs.workbench = projection.Workbench
		mutated = true
	}
	return mutated, nil
}

func browserRuntimeBuildTopLevelAliasProjection(
	inputs browserRuntimeTopLevelAliasInputs,
) (browserRuntimeTopLevelAliasProjection, bool, error) {
	working := inputs
	projection := browserRuntimeTopLevelAliasProjection{}
	mutated := false
	if blob, summary, ok := browserRuntimeBuildTopLevelExplanationAlias(working); ok {
		projection.Explanation = blob
		projection.ExplanationValue = summary
		working.setExplanation(summary)
		mutated = true
	}
	if blob, summary, ok := browserRuntimeBuildTopLevelDiagnosticsAlias(working); ok {
		projection.Diagnostics = blob
		projection.DiagnosticsValue = summary
		working.setDiagnostics(summary)
		mutated = true
	}
	if blob, summary, ok := browserRuntimeBuildTopLevelSummaryAlias(working); ok {
		projection.Summary = blob
		projection.SummaryValue = summary
		working.setSummary(summary)
		mutated = true
	}
	if blob, display, ok := browserRuntimeBuildTopLevelDisplayAlias(working); ok {
		projection.Display = blob
		projection.DisplayValue = display
		working.setDisplay(display)
		mutated = true
	}
	if blob, surface, ok := browserRuntimeBuildTopLevelSurfaceAlias(working); ok {
		projection.Surface = blob
		projection.SurfaceValue = surface
		projection.ApplySurface = true
		working.setSurface(surface)
		mutated = true
	}
	if blob, view, ok := browserRuntimeBuildTopLevelViewAlias(working); ok {
		projection.View = blob
		projection.ViewValue = view
		projection.ApplyView = true
		working.setView(view)
		mutated = true
	}
	return projection, mutated, nil
}

func browserRuntimeApplyTopLevelAliasProjection(
	payload map[string]json.RawMessage,
	inputs *browserRuntimeTopLevelAliasInputs,
) (bool, error) {
	if payload == nil || inputs == nil {
		return false, nil
	}
	projection, ok, err := browserRuntimeBuildTopLevelAliasProjection(*inputs)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	mutated := false
	if !browserUnifiedHasNonNullJSONField(payload, "explanation") && projection.ExplanationValue != nil {
		payload["explanation"] = projection.Explanation
		inputs.setExplanation(projection.ExplanationValue)
		mutated = true
	}
	if !browserUnifiedHasNonNullJSONField(payload, "diagnostics") && projection.DiagnosticsValue != nil {
		payload["diagnostics"] = projection.Diagnostics
		inputs.setDiagnostics(projection.DiagnosticsValue)
		mutated = true
	}
	if !browserUnifiedHasNonNullJSONField(payload, "summary") && projection.SummaryValue != nil {
		payload["summary"] = projection.Summary
		inputs.setSummary(projection.SummaryValue)
		mutated = true
	}
	if !browserUnifiedHasNonNullJSONField(payload, "display") && projection.DisplayValue != nil {
		payload["display"] = projection.Display
		inputs.setDisplay(projection.DisplayValue)
		mutated = true
	}
	if projection.ApplySurface && projection.SurfaceValue != nil &&
		(!browserUnifiedHasNonNullJSONField(payload, "surface") || browserRuntimeTopLevelPayloadHasSurfaceMetadata(*inputs)) {
		payload["surface"] = projection.Surface
		inputs.setSurface(projection.SurfaceValue)
		mutated = true
	}
	if projection.ApplyView && projection.ViewValue != nil &&
		(!browserUnifiedHasNonNullJSONField(payload, "view") || browserRuntimeTopLevelPayloadHasSurfaceMetadata(*inputs)) {
		payload["view"] = projection.View
		inputs.setView(projection.ViewValue)
		mutated = true
	}
	return mutated, nil
}
