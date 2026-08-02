package browserruntime

import "strings"

// SharedSessionBrowserSummary captures the stable summary/explanation contract
// that tool payloads derive from resolver/review/coordination guidance.
type SharedSessionBrowserSummary struct {
	Category             string
	State                string
	SummaryCode          string
	NextStepAlias        string
	ManualRetryHint      string
	ResolvedViaFallback  bool
	PrimaryBrowserAction string
	PrimaryNodeAction    string
	NextStep             string
}

// SharedSessionBrowserDisplay captures the stable ready/sections display shell
// that tool payloads derive from a shared guidance summary.
type SharedSessionBrowserDisplay struct {
	Ready    bool
	Sections []string
	SharedSessionBrowserSummary
}

// SharedSessionBrowserGuidanceProjection captures the shared resolver/review
// guidance summaries that tool payloads adapt into local explanation,
// diagnostics, summary, and display aliases.
type SharedSessionBrowserGuidanceProjection struct {
	ResolverExplanation    *SharedSessionBrowserSummary
	DiagnosticsExplanation *SharedSessionBrowserSummary
	WorkbenchExplanation   *SharedSessionBrowserSummary
	WorkbenchDiagnostics   *SharedSessionBrowserSummary
	WorkbenchSummary       *SharedSessionBrowserSummary
	WorkbenchDisplay       *SharedSessionBrowserDisplay
	Explanation            *SharedSessionBrowserSummary
	Diagnostics            *SharedSessionBrowserSummary
	Summary                *SharedSessionBrowserSummary
	Display                *SharedSessionBrowserDisplay
}

// SharedSessionBrowserGuidanceProjectionRequest carries the shared inputs
// needed to derive explicit review, resolver, and workbench guidance
// summaries.
type SharedSessionBrowserGuidanceProjectionRequest struct {
	IncludeWorkbenchSurface       bool
	ActionKind                    string
	ActionStatus                  string
	ActionabilityStatus           string
	ActionabilityFailedCheck      string
	ActionabilityFailureReason    string
	ActionabilityRetryDisposition string
	ActionabilityManualRetryHint  string
	ActionabilityRecoveryAction   string
	FailureReasonCode             string
	ReviewStatus                  string
	ReviewDecision                string
	ResolverBlockedBy             string
	ResolverAmbiguityClass        string
	ResolverCandidateKind         string
	ResolverRetryDisposition      string
	ResolverManualRetryHint       string
	ResolverNextStepAlias         string
	Routes                        []SharedSessionBrowserRouteCoordinationInput
	WorkbenchReady                bool
	WorkbenchSections             []string
	WorkbenchPrimaryBrowserAction string
	WorkbenchPrimaryNodeAction    string
	WorkbenchNextStep             string
}

// SharedSessionBrowserExplanationAliasRequest carries the shared inputs needed
// to build the top-level explanation alias from workbench/diagnostics/resolver
// guidance summaries.
type SharedSessionBrowserExplanationAliasRequest struct {
	Workbench   *SharedSessionBrowserSummary
	Diagnostics *SharedSessionBrowserSummary
	Resolver    *SharedSessionBrowserSummary
}

// SharedSessionBrowserDiagnosticsAliasRequest carries the shared inputs needed
// to build the top-level diagnostics alias from workbench/diagnostics/summary
// guidance summaries.
type SharedSessionBrowserDiagnosticsAliasRequest struct {
	Workbench   *SharedSessionBrowserSummary
	Diagnostics *SharedSessionBrowserSummary
	Summary     *SharedSessionBrowserSummary
}

type sharedSessionBrowserActionSuccessContract struct {
	ActionKind   string
	ActionStatus string
	Category     string
	State        string
	SummaryCode  string
}

var sharedSessionBrowserRuntimeSessionActionSuccessContracts = []sharedSessionBrowserActionSuccessContract{
	{ActionKind: "repair", ActionStatus: "ok", Category: "coordination", State: "completed", SummaryCode: "repair_completed"},
	{ActionKind: "prepare", ActionStatus: "ok", Category: "coordination", State: "completed", SummaryCode: "prepare_completed"},
	{ActionKind: "profiles", ActionStatus: "ok", Category: "inspection", State: "completed", SummaryCode: "profiles_completed"},
	{ActionKind: "sessions", ActionStatus: "ok", Category: "inspection", State: "completed", SummaryCode: "sessions_completed"},
	{ActionKind: "start", ActionStatus: "ok", Category: "coordination", State: "completed", SummaryCode: "start_completed"},
	{ActionKind: "restart", ActionStatus: "ok", Category: "coordination", State: "completed", SummaryCode: "restart_completed"},
	{ActionKind: "refresh", ActionStatus: "ok", Category: "coordination", State: "completed", SummaryCode: "refresh_completed"},
	{ActionKind: "stop", ActionStatus: "ok", Category: "coordination", State: "completed", SummaryCode: "stop_completed"},
	{ActionKind: "create_profile", ActionStatus: "ok", Category: "coordination", State: "completed", SummaryCode: "create_profile_completed"},
	{ActionKind: "delete_profile", ActionStatus: "ok", Category: "coordination", State: "completed", SummaryCode: "delete_profile_completed"},
	{ActionKind: "coordinate", ActionStatus: "ok", Category: "coordination", State: "completed", SummaryCode: "coordinate_completed"},
	{ActionKind: "select_profile", ActionStatus: "ok", Category: "coordination", State: "completed", SummaryCode: "select_profile_completed"},
	{ActionKind: "select_target", ActionStatus: "ok", Category: "coordination", State: "completed", SummaryCode: "select_target_completed"},
	{ActionKind: "sync_session", ActionStatus: "ok", Category: "coordination", State: "completed", SummaryCode: "sync_session_completed"},
	{ActionKind: "clear_profile", ActionStatus: "ok", Category: "coordination", State: "completed", SummaryCode: "clear_profile_completed"},
	{ActionKind: "clear_session", ActionStatus: "ok", Category: "coordination", State: "completed", SummaryCode: "clear_session_completed"},
	{ActionKind: "clear_target", ActionStatus: "ok", Category: "coordination", State: "completed", SummaryCode: "clear_target_completed"},
}

var sharedSessionBrowserRuntimeObservationOnlyActionKinds = []string{
	"status",
	"workbench",
	"doctor",
}

// BuildSharedSessionBrowserGuidanceProjection lowers resolver/review guidance
// plus optional workbench action-plan state into the stable shared summary and
// display contracts consumed by tool payloads.
func BuildSharedSessionBrowserGuidanceProjection(
	req SharedSessionBrowserGuidanceProjectionRequest,
) SharedSessionBrowserGuidanceProjection {
	req = normalizeSharedSessionBrowserGuidanceProjectionRequest(req)

	resolver := buildSharedSessionBrowserResolverExplanation(req)
	diagnosticsExplanation := buildSharedSessionBrowserDiagnosticsExplanation(req, resolver)

	projection := SharedSessionBrowserGuidanceProjection{
		ResolverExplanation:    cloneSharedSessionBrowserSummary(resolver),
		DiagnosticsExplanation: cloneSharedSessionBrowserSummary(diagnosticsExplanation),
	}
	if req.IncludeWorkbenchSurface {
		projection.WorkbenchExplanation = cloneSharedSessionBrowserSummary(diagnosticsExplanation)
		projection.WorkbenchDiagnostics = buildSharedSessionBrowserWorkbenchDiagnostics(
			projection.WorkbenchExplanation,
			req.WorkbenchPrimaryBrowserAction,
			req.WorkbenchPrimaryNodeAction,
			req.WorkbenchNextStep,
		)
		projection.WorkbenchSummary = buildSharedSessionBrowserWorkbenchSummary(
			projection.WorkbenchDiagnostics,
			projection.WorkbenchExplanation,
		)
		projection.WorkbenchDisplay = buildSharedSessionBrowserWorkbenchDisplay(
			req.WorkbenchReady,
			req.WorkbenchSections,
			projection.WorkbenchSummary,
			projection.WorkbenchDiagnostics,
			projection.WorkbenchExplanation,
			req.WorkbenchPrimaryBrowserAction,
			req.WorkbenchPrimaryNodeAction,
			req.WorkbenchNextStep,
		)
	}

	projection.Explanation = BuildSharedSessionBrowserExplanationAliasFromRequest(
		SharedSessionBrowserExplanationAliasRequest{
			Workbench:   projection.WorkbenchExplanation,
			Diagnostics: projection.DiagnosticsExplanation,
			Resolver:    projection.ResolverExplanation,
		},
	)
	projection.Diagnostics = BuildSharedSessionBrowserDiagnosticsAliasFromRequest(
		SharedSessionBrowserDiagnosticsAliasRequest{
			Workbench:   projection.WorkbenchSummary,
			Diagnostics: projection.DiagnosticsExplanation,
		},
	)
	projection.Summary = buildSharedSessionBrowserTopLevelSummary(
		projection.WorkbenchSummary,
		projection.DiagnosticsExplanation,
		projection.ResolverExplanation,
	)
	projection.Display = buildSharedSessionBrowserDisplayAlias(
		projection.WorkbenchDisplay,
		projection.Summary,
		projection.Diagnostics,
		projection.Explanation,
	)

	return projection
}

// BuildSharedSessionBrowserExplanationAliasFromRequest projects the stable
// top-level explanation alias from workbench, diagnostics, and resolver
// guidance summaries in descending priority order.
func BuildSharedSessionBrowserExplanationAliasFromRequest(
	req SharedSessionBrowserExplanationAliasRequest,
) *SharedSessionBrowserSummary {
	return buildSharedSessionBrowserExplanationAlias(
		req.Workbench,
		req.Diagnostics,
		req.Resolver,
	)
}

// BuildSharedSessionBrowserDiagnosticsAliasFromRequest projects the stable
// top-level diagnostics alias from workbench, diagnostics, and summary
// guidance summaries in descending priority order.
func BuildSharedSessionBrowserDiagnosticsAliasFromRequest(
	req SharedSessionBrowserDiagnosticsAliasRequest,
) *SharedSessionBrowserSummary {
	return buildSharedSessionBrowserDiagnosticsAlias(
		req.Workbench,
		req.Diagnostics,
		req.Summary,
	)
}

func normalizeSharedSessionBrowserGuidanceProjectionRequest(
	req SharedSessionBrowserGuidanceProjectionRequest,
) SharedSessionBrowserGuidanceProjectionRequest {
	req.ActionKind = strings.TrimSpace(req.ActionKind)
	req.ActionStatus = strings.TrimSpace(req.ActionStatus)
	req.ActionabilityStatus = strings.TrimSpace(req.ActionabilityStatus)
	req.ActionabilityFailedCheck = sharedSessionBrowserActionabilityToken(req.ActionabilityFailedCheck)
	req.ActionabilityFailureReason = strings.TrimSpace(req.ActionabilityFailureReason)
	req.ActionabilityRetryDisposition = strings.TrimSpace(req.ActionabilityRetryDisposition)
	req.ActionabilityManualRetryHint = strings.TrimSpace(req.ActionabilityManualRetryHint)
	req.ActionabilityRecoveryAction = strings.TrimSpace(req.ActionabilityRecoveryAction)
	req.FailureReasonCode = strings.TrimSpace(req.FailureReasonCode)
	req.ReviewStatus = strings.TrimSpace(req.ReviewStatus)
	req.ReviewDecision = strings.TrimSpace(req.ReviewDecision)
	req.ResolverBlockedBy = strings.TrimSpace(req.ResolverBlockedBy)
	req.ResolverAmbiguityClass = strings.TrimSpace(req.ResolverAmbiguityClass)
	req.ResolverCandidateKind = strings.TrimSpace(req.ResolverCandidateKind)
	req.ResolverRetryDisposition = strings.TrimSpace(req.ResolverRetryDisposition)
	req.ResolverManualRetryHint = strings.TrimSpace(req.ResolverManualRetryHint)
	req.ResolverNextStepAlias = strings.TrimSpace(req.ResolverNextStepAlias)
	req.WorkbenchPrimaryBrowserAction = strings.TrimSpace(req.WorkbenchPrimaryBrowserAction)
	req.WorkbenchPrimaryNodeAction = strings.TrimSpace(req.WorkbenchPrimaryNodeAction)
	req.WorkbenchNextStep = strings.TrimSpace(req.WorkbenchNextStep)
	if len(req.WorkbenchSections) > 0 {
		req.WorkbenchSections = mergeSharedSessionBrowserGuidanceSections(nil, req.WorkbenchSections)
	}
	return req
}

func buildSharedSessionBrowserResolverExplanation(
	req SharedSessionBrowserGuidanceProjectionRequest,
) *SharedSessionBrowserSummary {
	state := sharedSessionBrowserResolverExplanationState(req)
	summaryCode := sharedSessionBrowserResolverExplanationCode(req)
	if state == "" && summaryCode == "" && req.ResolverNextStepAlias == "" && req.ResolverManualRetryHint == "" {
		return nil
	}
	summary := &SharedSessionBrowserSummary{
		State:           state,
		SummaryCode:     summaryCode,
		NextStepAlias:   req.ResolverNextStepAlias,
		ManualRetryHint: req.ResolverManualRetryHint,
	}
	if strings.EqualFold(summary.State, "resolved_via_fallback") {
		summary.Category = "resolver_fallback"
		summary.ResolvedViaFallback = true
	} else if summary.State != "" || summary.SummaryCode != "" || summary.NextStepAlias != "" || summary.ManualRetryHint != "" {
		summary.Category = "resolver"
	}
	if sharedSessionBrowserSummaryEmpty(*summary) {
		return nil
	}
	return summary
}

func buildSharedSessionBrowserDiagnosticsExplanation(
	req SharedSessionBrowserGuidanceProjectionRequest,
	resolver *SharedSessionBrowserSummary,
) *SharedSessionBrowserSummary {
	if review := buildSharedSessionBrowserExplicitReviewExplanation(req); review != nil {
		return review
	}
	if review := buildSharedSessionBrowserReviewExplanation(req.Routes); review != nil {
		return review
	}
	if actionability := buildSharedSessionBrowserActionabilityExplanation(req); actionability != nil {
		return actionability
	}
	if resolver == nil {
		return buildSharedSessionBrowserActionSuccessExplanation(req)
	}
	summary := cloneSharedSessionBrowserSummary(resolver)
	summary.Category = "resolver"
	summary.PrimaryBrowserAction = ""
	summary.PrimaryNodeAction = ""
	summary.NextStep = ""
	if sharedSessionBrowserSummaryEmpty(*summary) {
		return nil
	}
	return summary
}

func buildSharedSessionBrowserActionSuccessExplanation(
	req SharedSessionBrowserGuidanceProjectionRequest,
) *SharedSessionBrowserSummary {
	if summary := buildSharedSessionBrowserRuntimeSessionActionSuccessExplanation(req); summary != nil {
		return finalizeSharedSessionBrowserActionSuccessSummary(req.ActionKind, summary)
	}
	var summary *SharedSessionBrowserSummary
	switch {
	case strings.EqualFold(req.ActionKind, "open") && strings.EqualFold(req.ActionStatus, "opened"):
		summary = &SharedSessionBrowserSummary{Category: "navigation", State: "completed", SummaryCode: "open_completed"}
	case strings.EqualFold(req.ActionKind, "navigate") && strings.EqualFold(req.ActionStatus, "navigated"):
		summary = &SharedSessionBrowserSummary{Category: "navigation", State: "completed", SummaryCode: "navigate_completed"}
	case strings.EqualFold(req.ActionKind, "extract") && strings.EqualFold(req.ActionStatus, "extracted"):
		summary = &SharedSessionBrowserSummary{Category: "content", State: "completed", SummaryCode: "extract_completed"}
	case strings.EqualFold(req.ActionKind, "snapshot") && strings.EqualFold(req.ActionStatus, "snapshotted"):
		summary = &SharedSessionBrowserSummary{Category: "content", State: "completed", SummaryCode: "snapshot_completed"}
	case strings.EqualFold(req.ActionKind, "evaluate") && strings.EqualFold(req.ActionStatus, "evaluated"):
		summary = &SharedSessionBrowserSummary{Category: "script", State: "completed", SummaryCode: "evaluate_completed"}
	case strings.EqualFold(req.ActionKind, "trace_start") && strings.EqualFold(req.ActionStatus, "started"):
		summary = &SharedSessionBrowserSummary{Category: "trace", State: "started", SummaryCode: "trace_start_started"}
	case strings.EqualFold(req.ActionKind, "wait") && strings.EqualFold(req.ActionStatus, "waited"):
		summary = &SharedSessionBrowserSummary{Category: "timing", State: "completed", SummaryCode: "wait_completed"}
	case strings.EqualFold(req.ActionKind, "click") && strings.EqualFold(req.ActionStatus, "clicked"):
		summary = &SharedSessionBrowserSummary{Category: "interaction", State: "completed", SummaryCode: "click_completed"}
	case strings.EqualFold(req.ActionKind, "dialog") && strings.EqualFold(req.ActionStatus, "armed"):
		summary = &SharedSessionBrowserSummary{Category: "interaction", State: "started", SummaryCode: "dialog_armed"}
	case strings.EqualFold(req.ActionKind, "console") && strings.EqualFold(req.ActionStatus, "ok"):
		summary = &SharedSessionBrowserSummary{Category: "observability", State: "completed", SummaryCode: "console_collected"}
	case strings.EqualFold(req.ActionKind, "requests") && strings.EqualFold(req.ActionStatus, "ok"):
		summary = &SharedSessionBrowserSummary{Category: "observability", State: "completed", SummaryCode: "requests_collected"}
	case strings.EqualFold(req.ActionKind, "requests") && strings.EqualFold(req.ActionStatus, "cleared"):
		summary = &SharedSessionBrowserSummary{Category: "observability", State: "completed", SummaryCode: "requests_cleared"}
	case strings.EqualFold(req.ActionKind, "response_body") && strings.EqualFold(req.ActionStatus, "ok"):
		summary = &SharedSessionBrowserSummary{Category: "content", State: "completed", SummaryCode: "response_body_collected"}
	case strings.EqualFold(req.ActionKind, "errors") && strings.EqualFold(req.ActionStatus, "ok"):
		summary = &SharedSessionBrowserSummary{Category: "observability", State: "completed", SummaryCode: "errors_collected"}
	case strings.EqualFold(req.ActionKind, "errors") && strings.EqualFold(req.ActionStatus, "cleared"):
		summary = &SharedSessionBrowserSummary{Category: "observability", State: "completed", SummaryCode: "errors_cleared"}
	case strings.EqualFold(req.ActionKind, "type") && strings.EqualFold(req.ActionStatus, "typed"):
		summary = &SharedSessionBrowserSummary{Category: "form", State: "completed", SummaryCode: "type_completed"}
	case strings.EqualFold(req.ActionKind, "press") && strings.EqualFold(req.ActionStatus, "pressed"):
		summary = &SharedSessionBrowserSummary{Category: "interaction", State: "completed", SummaryCode: "press_completed"}
	case strings.EqualFold(req.ActionKind, "highlight") && strings.EqualFold(req.ActionStatus, "highlighted"):
		summary = &SharedSessionBrowserSummary{Category: "interaction", State: "completed", SummaryCode: "highlight_completed"}
	case strings.EqualFold(req.ActionKind, "hover") && strings.EqualFold(req.ActionStatus, "hovered"):
		summary = &SharedSessionBrowserSummary{Category: "interaction", State: "completed", SummaryCode: "hover_completed"}
	case strings.EqualFold(req.ActionKind, "drag") && strings.EqualFold(req.ActionStatus, "dragged"):
		summary = &SharedSessionBrowserSummary{Category: "interaction", State: "completed", SummaryCode: "drag_completed"}
	case strings.EqualFold(req.ActionKind, "upload") && strings.EqualFold(req.ActionStatus, "uploaded"):
		summary = &SharedSessionBrowserSummary{Category: "form", State: "completed", SummaryCode: "upload_completed"}
	case strings.EqualFold(req.ActionKind, "screenshot") && strings.EqualFold(req.ActionStatus, "captured"):
		summary = &SharedSessionBrowserSummary{Category: "capture", State: "completed", SummaryCode: "screenshot_completed"}
	case strings.EqualFold(req.ActionKind, "download") && strings.EqualFold(req.ActionStatus, "downloaded"):
		summary = &SharedSessionBrowserSummary{Category: "artifact", State: "completed", SummaryCode: "download_completed"}
	case strings.EqualFold(req.ActionKind, "wait_download") && strings.EqualFold(req.ActionStatus, "downloaded"):
		summary = &SharedSessionBrowserSummary{Category: "artifact", State: "completed", SummaryCode: "wait_download_completed"}
	case strings.EqualFold(req.ActionKind, "save_pdf") && strings.EqualFold(req.ActionStatus, "saved"):
		summary = &SharedSessionBrowserSummary{Category: "artifact", State: "completed", SummaryCode: "save_pdf_completed"}
	case strings.EqualFold(req.ActionKind, "save_html") && strings.EqualFold(req.ActionStatus, "saved"):
		summary = &SharedSessionBrowserSummary{Category: "artifact", State: "completed", SummaryCode: "save_html_completed"}
	case strings.EqualFold(req.ActionKind, "trace_stop") && strings.EqualFold(req.ActionStatus, "saved"):
		summary = &SharedSessionBrowserSummary{Category: "artifact", State: "completed", SummaryCode: "trace_stop_completed"}
	case strings.EqualFold(req.ActionKind, "select") && strings.EqualFold(req.ActionStatus, "selected"):
		summary = &SharedSessionBrowserSummary{Category: "form", State: "completed", SummaryCode: "select_completed"}
	case strings.EqualFold(req.ActionKind, "fill") && strings.EqualFold(req.ActionStatus, "filled"):
		summary = &SharedSessionBrowserSummary{Category: "form", State: "completed", SummaryCode: "fill_completed"}
	case strings.EqualFold(req.ActionKind, "storage_set") && strings.EqualFold(req.ActionStatus, "updated"):
		summary = &SharedSessionBrowserSummary{Category: "storage", State: "completed", SummaryCode: "storage_set_completed"}
	case strings.EqualFold(req.ActionKind, "storage_clear") && strings.EqualFold(req.ActionStatus, "cleared"):
		summary = &SharedSessionBrowserSummary{Category: "storage", State: "completed", SummaryCode: "storage_clear_completed"}
	case strings.EqualFold(req.ActionKind, "cookies_set") && strings.EqualFold(req.ActionStatus, "updated"):
		summary = &SharedSessionBrowserSummary{Category: "storage", State: "completed", SummaryCode: "cookies_set_completed"}
	case strings.EqualFold(req.ActionKind, "cookies_clear") && strings.EqualFold(req.ActionStatus, "cleared"):
		summary = &SharedSessionBrowserSummary{Category: "storage", State: "completed", SummaryCode: "cookies_clear_completed"}
	case strings.EqualFold(req.ActionKind, "cookies") && strings.EqualFold(req.ActionStatus, "ok"):
		summary = &SharedSessionBrowserSummary{Category: "storage", State: "completed", SummaryCode: "cookies_collected"}
	case strings.EqualFold(req.ActionKind, "storage") && strings.EqualFold(req.ActionStatus, "ok"):
		summary = &SharedSessionBrowserSummary{Category: "storage", State: "completed", SummaryCode: "storage_collected"}
	case strings.EqualFold(req.ActionKind, "headers") && strings.EqualFold(req.ActionStatus, "updated"):
		summary = &SharedSessionBrowserSummary{Category: "network", State: "completed", SummaryCode: "headers_updated"}
	case strings.EqualFold(req.ActionKind, "headers") && strings.EqualFold(req.ActionStatus, "cleared"):
		summary = &SharedSessionBrowserSummary{Category: "network", State: "completed", SummaryCode: "headers_cleared"}
	case strings.EqualFold(req.ActionKind, "offline") && strings.EqualFold(req.ActionStatus, "updated"):
		summary = &SharedSessionBrowserSummary{Category: "network", State: "completed", SummaryCode: "offline_updated"}
	case strings.EqualFold(req.ActionKind, "offline") && strings.EqualFold(req.ActionStatus, "cleared"):
		summary = &SharedSessionBrowserSummary{Category: "network", State: "completed", SummaryCode: "offline_cleared"}
	case strings.EqualFold(req.ActionKind, "credentials") && strings.EqualFold(req.ActionStatus, "updated"):
		summary = &SharedSessionBrowserSummary{Category: "auth", State: "completed", SummaryCode: "credentials_updated"}
	case strings.EqualFold(req.ActionKind, "credentials") && strings.EqualFold(req.ActionStatus, "cleared"):
		summary = &SharedSessionBrowserSummary{Category: "auth", State: "completed", SummaryCode: "credentials_cleared"}
	case strings.EqualFold(req.ActionKind, "geolocation") && strings.EqualFold(req.ActionStatus, "updated"):
		summary = &SharedSessionBrowserSummary{Category: "settings", State: "completed", SummaryCode: "geolocation_updated"}
	case strings.EqualFold(req.ActionKind, "geolocation") && strings.EqualFold(req.ActionStatus, "cleared"):
		summary = &SharedSessionBrowserSummary{Category: "settings", State: "completed", SummaryCode: "geolocation_cleared"}
	case strings.EqualFold(req.ActionKind, "media") && strings.EqualFold(req.ActionStatus, "updated"):
		summary = &SharedSessionBrowserSummary{Category: "settings", State: "completed", SummaryCode: "media_updated"}
	case strings.EqualFold(req.ActionKind, "media") && strings.EqualFold(req.ActionStatus, "cleared"):
		summary = &SharedSessionBrowserSummary{Category: "settings", State: "completed", SummaryCode: "media_cleared"}
	case strings.EqualFold(req.ActionKind, "timezone") && strings.EqualFold(req.ActionStatus, "updated"):
		summary = &SharedSessionBrowserSummary{Category: "settings", State: "completed", SummaryCode: "timezone_updated"}
	case strings.EqualFold(req.ActionKind, "timezone") && strings.EqualFold(req.ActionStatus, "cleared"):
		summary = &SharedSessionBrowserSummary{Category: "settings", State: "completed", SummaryCode: "timezone_cleared"}
	case strings.EqualFold(req.ActionKind, "locale") && strings.EqualFold(req.ActionStatus, "updated"):
		summary = &SharedSessionBrowserSummary{Category: "settings", State: "completed", SummaryCode: "locale_updated"}
	case strings.EqualFold(req.ActionKind, "locale") && strings.EqualFold(req.ActionStatus, "cleared"):
		summary = &SharedSessionBrowserSummary{Category: "settings", State: "completed", SummaryCode: "locale_cleared"}
	case strings.EqualFold(req.ActionKind, "device") && strings.EqualFold(req.ActionStatus, "updated"):
		summary = &SharedSessionBrowserSummary{Category: "settings", State: "completed", SummaryCode: "device_updated"}
	case strings.EqualFold(req.ActionKind, "device") && strings.EqualFold(req.ActionStatus, "cleared"):
		summary = &SharedSessionBrowserSummary{Category: "settings", State: "completed", SummaryCode: "device_cleared"}
	case strings.EqualFold(req.ActionKind, "resize") && strings.EqualFold(req.ActionStatus, "resized"):
		summary = &SharedSessionBrowserSummary{Category: "viewport", State: "completed", SummaryCode: "resize_completed"}
	case strings.EqualFold(req.ActionKind, "list_tabs") && (strings.EqualFold(req.ActionStatus, "listed") || strings.EqualFold(req.ActionStatus, "ok")):
		summary = &SharedSessionBrowserSummary{Category: "tabs", State: "completed", SummaryCode: "list_tabs_completed"}
	case strings.EqualFold(req.ActionKind, "focus_tab") && strings.EqualFold(req.ActionStatus, "focused"):
		summary = &SharedSessionBrowserSummary{Category: "tabs", State: "completed", SummaryCode: "focus_tab_completed"}
	case strings.EqualFold(req.ActionKind, "close_tab") && strings.EqualFold(req.ActionStatus, "closed"):
		summary = &SharedSessionBrowserSummary{Category: "tabs", State: "completed", SummaryCode: "close_tab_completed"}
	default:
		return nil
	}
	return finalizeSharedSessionBrowserActionSuccessSummary(req.ActionKind, summary)
}

func finalizeSharedSessionBrowserActionSuccessSummary(
	actionKind string,
	summary *SharedSessionBrowserSummary,
) *SharedSessionBrowserSummary {
	if summary == nil {
		return nil
	}
	if primaryAction := sharedSessionBrowserPrimaryActionForActionKind(actionKind); primaryAction != "" {
		summary.PrimaryBrowserAction = primaryAction
		summary.NextStep = primaryAction
	}
	if sharedSessionBrowserSummaryEmpty(*summary) {
		return nil
	}
	return summary
}

func buildSharedSessionBrowserRuntimeSessionActionSuccessExplanation(
	req SharedSessionBrowserGuidanceProjectionRequest,
) *SharedSessionBrowserSummary {
	for _, contract := range sharedSessionBrowserRuntimeSessionActionSuccessContracts {
		if !strings.EqualFold(req.ActionKind, contract.ActionKind) ||
			!strings.EqualFold(req.ActionStatus, contract.ActionStatus) {
			continue
		}
		return &SharedSessionBrowserSummary{
			Category:    contract.Category,
			State:       contract.State,
			SummaryCode: contract.SummaryCode,
		}
	}
	return nil
}

func buildSharedSessionBrowserActionabilityExplanation(
	req SharedSessionBrowserGuidanceProjectionRequest,
) *SharedSessionBrowserSummary {
	if !strings.EqualFold(req.ActionabilityStatus, BrowserActionabilityStatusFailed) {
		return nil
	}
	failedCheck := sharedSessionBrowserActionabilityToken(req.ActionabilityFailedCheck)
	if failedCheck == "" || failedCheck == "resolve_target" {
		return nil
	}
	nextStepAlias := sharedSessionBrowserActionabilityNextStepAlias(req)
	summary := &SharedSessionBrowserSummary{
		Category:        "actionability",
		State:           sharedSessionBrowserActionabilityState(failedCheck),
		SummaryCode:     sharedSessionBrowserActionabilitySummaryCode(req, failedCheck),
		NextStepAlias:   nextStepAlias,
		ManualRetryHint: req.ActionabilityManualRetryHint,
	}
	browserAction, nodeAction, nextStep := sharedSessionBrowserActionabilityPrimaryActions(req.ActionabilityRecoveryAction, nextStepAlias)
	summary.PrimaryBrowserAction = browserAction
	summary.PrimaryNodeAction = nodeAction
	summary.NextStep = nextStep
	if sharedSessionBrowserSummaryEmpty(*summary) {
		return nil
	}
	return summary
}

func buildSharedSessionBrowserExplicitReviewExplanation(
	req SharedSessionBrowserGuidanceProjectionRequest,
) *SharedSessionBrowserSummary {
	if !strings.EqualFold(req.ReviewStatus, "review_required") {
		return nil
	}
	summaryCode := strings.TrimSpace(sharedSessionBrowserReviewSummaryCode(req.ReviewDecision))
	if summaryCode == "" {
		return nil
	}
	summary := &SharedSessionBrowserSummary{
		Category:        "review",
		State:           "manual_confirmation_required",
		SummaryCode:     summaryCode,
		ManualRetryHint: "rerun_with_force",
	}
	if sharedSessionBrowserSummaryEmpty(*summary) {
		return nil
	}
	return summary
}

func buildSharedSessionBrowserReviewExplanation(
	routes []SharedSessionBrowserRouteCoordinationInput,
) *SharedSessionBrowserSummary {
	selectedState := strings.TrimSpace(SharedSessionBrowserSelectedFollowPolicyState(routes))
	if selectedState == "" || strings.EqualFold(selectedState, "auto_follow_allowed") {
		return nil
	}
	actions := EvaluateSharedSessionBrowserFollowPolicyActions(routes)
	summary := &SharedSessionBrowserSummary{
		Category:        "review",
		State:           "manual_confirmation_required",
		SummaryCode:     selectedState,
		NextStepAlias:   sharedSessionBrowserNextStepAliasFromAction(actions.PrimaryAction),
		ManualRetryHint: "rerun_with_force",
	}
	if sharedSessionBrowserSummaryEmpty(*summary) {
		return nil
	}
	return summary
}

func buildSharedSessionBrowserWorkbenchDiagnostics(
	explanation *SharedSessionBrowserSummary,
	primaryBrowserAction string,
	primaryNodeAction string,
	nextStep string,
) *SharedSessionBrowserSummary {
	if explanation == nil && primaryBrowserAction == "" && primaryNodeAction == "" && nextStep == "" {
		return nil
	}
	summary := &SharedSessionBrowserSummary{
		PrimaryBrowserAction: primaryBrowserAction,
		PrimaryNodeAction:    primaryNodeAction,
		NextStep:             nextStep,
	}
	if explanation != nil {
		summary.Category = strings.TrimSpace(explanation.Category)
		summary.State = strings.TrimSpace(explanation.State)
		summary.SummaryCode = strings.TrimSpace(explanation.SummaryCode)
		summary.NextStepAlias = strings.TrimSpace(explanation.NextStepAlias)
		summary.ManualRetryHint = strings.TrimSpace(explanation.ManualRetryHint)
		summary.ResolvedViaFallback = explanation.ResolvedViaFallback
		if strings.TrimSpace(summary.PrimaryBrowserAction) == "" {
			summary.PrimaryBrowserAction = strings.TrimSpace(explanation.PrimaryBrowserAction)
		}
		if strings.TrimSpace(summary.PrimaryNodeAction) == "" {
			summary.PrimaryNodeAction = strings.TrimSpace(explanation.PrimaryNodeAction)
		}
		if strings.TrimSpace(summary.NextStep) == "" {
			summary.NextStep = strings.TrimSpace(explanation.NextStep)
		}
	} else {
		summary.Category = "coordination"
		summary.State = "action_plan_available"
		summary.SummaryCode = "workbench_action_plan"
	}
	if sharedSessionBrowserSummaryEmpty(*summary) {
		return nil
	}
	return summary
}

func buildSharedSessionBrowserWorkbenchSummary(
	diagnostics *SharedSessionBrowserSummary,
	explanation *SharedSessionBrowserSummary,
) *SharedSessionBrowserSummary {
	if diagnostics != nil {
		return cloneSharedSessionBrowserSummary(diagnostics)
	}
	return cloneSharedSessionBrowserSummary(explanation)
}

func buildSharedSessionBrowserWorkbenchDisplay(
	ready bool,
	sections []string,
	summary *SharedSessionBrowserSummary,
	diagnostics *SharedSessionBrowserSummary,
	explanation *SharedSessionBrowserSummary,
	primaryBrowserAction string,
	primaryNodeAction string,
	nextStep string,
) *SharedSessionBrowserDisplay {
	source := cloneSharedSessionBrowserSummary(summary)
	if source == nil {
		source = cloneSharedSessionBrowserSummary(diagnostics)
	}
	if source == nil {
		source = cloneSharedSessionBrowserSummary(explanation)
	}
	display := &SharedSessionBrowserDisplay{
		Ready:    ready,
		Sections: mergeSharedSessionBrowserGuidanceSections(nil, sections),
	}
	if source != nil {
		display.SharedSessionBrowserSummary = *source
	}
	if strings.TrimSpace(display.PrimaryBrowserAction) == "" {
		display.PrimaryBrowserAction = primaryBrowserAction
	}
	if strings.TrimSpace(display.PrimaryNodeAction) == "" {
		display.PrimaryNodeAction = primaryNodeAction
	}
	if strings.TrimSpace(display.NextStep) == "" {
		display.NextStep = nextStep
	}
	if sharedSessionBrowserDisplayEmpty(*display) {
		return nil
	}
	return display
}

func buildSharedSessionBrowserExplanationAlias(
	workbench *SharedSessionBrowserSummary,
	diagnostics *SharedSessionBrowserSummary,
	resolver *SharedSessionBrowserSummary,
) *SharedSessionBrowserSummary {
	if workbench != nil {
		return cloneSharedSessionBrowserSummary(workbench)
	}
	if diagnostics != nil {
		return cloneSharedSessionBrowserSummary(diagnostics)
	}
	return cloneSharedSessionBrowserSummary(resolver)
}

func buildSharedSessionBrowserDiagnosticsAlias(
	workbench *SharedSessionBrowserSummary,
	diagnostics *SharedSessionBrowserSummary,
	summary *SharedSessionBrowserSummary,
) *SharedSessionBrowserSummary {
	if workbench != nil {
		return cloneSharedSessionBrowserSummary(workbench)
	}
	if diagnostics != nil {
		return cloneSharedSessionBrowserSummary(diagnostics)
	}
	return cloneSharedSessionBrowserSummary(summary)
}

func buildSharedSessionBrowserTopLevelSummary(
	workbench *SharedSessionBrowserSummary,
	diagnostics *SharedSessionBrowserSummary,
	resolver *SharedSessionBrowserSummary,
) *SharedSessionBrowserSummary {
	if workbench != nil {
		return cloneSharedSessionBrowserSummary(workbench)
	}
	if diagnostics != nil {
		return cloneSharedSessionBrowserSummary(diagnostics)
	}
	return cloneSharedSessionBrowserSummary(resolver)
}

func buildSharedSessionBrowserDisplayAlias(
	workbench *SharedSessionBrowserDisplay,
	summary *SharedSessionBrowserSummary,
	diagnostics *SharedSessionBrowserSummary,
	explanation *SharedSessionBrowserSummary,
) *SharedSessionBrowserDisplay {
	if workbench != nil {
		return cloneSharedSessionBrowserDisplay(workbench)
	}
	if summary != nil {
		return sharedSessionBrowserDisplayFromSummary(summary)
	}
	if diagnostics != nil {
		return sharedSessionBrowserDisplayFromSummary(diagnostics)
	}
	return sharedSessionBrowserDisplayFromSummary(explanation)
}

func cloneSharedSessionBrowserSummary(summary *SharedSessionBrowserSummary) *SharedSessionBrowserSummary {
	if summary == nil {
		return nil
	}
	cloned := *summary
	if sharedSessionBrowserSummaryEmpty(cloned) {
		return nil
	}
	return &cloned
}

func cloneSharedSessionBrowserDisplay(display *SharedSessionBrowserDisplay) *SharedSessionBrowserDisplay {
	if display == nil {
		return nil
	}
	cloned := *display
	cloned.Sections = mergeSharedSessionBrowserGuidanceSections(nil, display.Sections)
	if sharedSessionBrowserDisplayEmpty(cloned) {
		return nil
	}
	return &cloned
}

func sharedSessionBrowserDisplayFromSummary(summary *SharedSessionBrowserSummary) *SharedSessionBrowserDisplay {
	if summary == nil {
		return nil
	}
	display := &SharedSessionBrowserDisplay{SharedSessionBrowserSummary: *summary}
	if sharedSessionBrowserDisplayEmpty(*display) {
		return nil
	}
	return display
}

func mergeSharedSessionBrowserGuidanceSections(current []string, next []string) []string {
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

func sharedSessionBrowserSummaryEmpty(summary SharedSessionBrowserSummary) bool {
	return strings.TrimSpace(summary.Category) == "" &&
		strings.TrimSpace(summary.State) == "" &&
		strings.TrimSpace(summary.SummaryCode) == "" &&
		strings.TrimSpace(summary.NextStepAlias) == "" &&
		strings.TrimSpace(summary.ManualRetryHint) == "" &&
		!summary.ResolvedViaFallback &&
		strings.TrimSpace(summary.PrimaryBrowserAction) == "" &&
		strings.TrimSpace(summary.PrimaryNodeAction) == "" &&
		strings.TrimSpace(summary.NextStep) == ""
}

func sharedSessionBrowserDisplayEmpty(display SharedSessionBrowserDisplay) bool {
	return !display.Ready &&
		len(display.Sections) == 0 &&
		sharedSessionBrowserSummaryEmpty(display.SharedSessionBrowserSummary)
}

func sharedSessionBrowserResolverExplanationState(req SharedSessionBrowserGuidanceProjectionRequest) string {
	switch req.ResolverRetryDisposition {
	case "manual_only":
		return "manual_resolution_required"
	case "auto_retry_allowed":
		return "automatic_retry_available"
	}
	if req.ResolverBlockedBy != "" || req.ResolverAmbiguityClass != "" {
		return "resolver_attention_required"
	}
	if req.ResolverNextStepAlias != "" || req.ResolverManualRetryHint != "" {
		return "guidance_available"
	}
	return ""
}

func sharedSessionBrowserResolverExplanationCode(req SharedSessionBrowserGuidanceProjectionRequest) string {
	parts := make([]string, 0, 2)
	if req.ResolverCandidateKind != "" {
		parts = append(parts, req.ResolverCandidateKind)
	}
	switch {
	case req.ResolverAmbiguityClass != "":
		parts = append(parts, req.ResolverAmbiguityClass)
	case req.ResolverBlockedBy != "":
		parts = append(parts, req.ResolverBlockedBy)
	}
	return strings.Join(parts, "_")
}

func sharedSessionBrowserNextStepAliasFromAction(action string) string {
	action = strings.TrimSpace(action)
	if action == "" {
		return ""
	}
	lower := strings.ToLower(action)
	switch {
	case strings.HasPrefix(lower, "browser action="):
		return strings.TrimSpace(action[len("browser action="):])
	case strings.HasPrefix(lower, "nodes action="):
		return strings.TrimSpace(action[len("nodes action="):])
	default:
		return action
	}
}

func sharedSessionBrowserActionabilityState(failedCheck string) string {
	switch sharedSessionBrowserActionabilityToken(failedCheck) {
	case "navigation_wait":
		return "post_action_wait_failed"
	default:
		return "actionability_failed"
	}
}

func sharedSessionBrowserActionabilitySummaryCode(req SharedSessionBrowserGuidanceProjectionRequest, failedCheck string) string {
	if code := strings.TrimSpace(req.FailureReasonCode); code != "" {
		return code
	}
	if reason := strings.TrimSpace(req.ActionabilityFailureReason); reason != "" {
		return reason
	}
	return "actionability_" + sharedSessionBrowserActionabilityToken(failedCheck) + "_failed"
}

func sharedSessionBrowserActionabilityNextStepAlias(req SharedSessionBrowserGuidanceProjectionRequest) string {
	if alias := sharedSessionBrowserNextStepAliasFromAction(req.ActionabilityRecoveryAction); alias != "" {
		return alias
	}
	switch sharedSessionBrowserActionabilityToken(req.ActionabilityFailedCheck) {
	case "stable", "navigation_wait":
		return "wait"
	case "attached", "visible", "receives_events", "frame_hit_target", "enabled", "editable":
		return "snapshot"
	default:
		return ""
	}
}

func sharedSessionBrowserActionabilityPrimaryActions(recoveryAction string, nextStepAlias string) (string, string, string) {
	action := strings.TrimSpace(recoveryAction)
	lower := strings.ToLower(action)
	switch {
	case strings.HasPrefix(lower, "browser action="):
		return action, "", action
	case strings.HasPrefix(lower, "nodes action="):
		return "", action, action
	}
	if nextStepAlias == "" {
		return "", "", ""
	}
	browserAction := sharedSessionBrowserPrimaryActionForActionKind(nextStepAlias)
	return browserAction, "", browserAction
}

func sharedSessionBrowserActionabilityToken(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}
	value = strings.NewReplacer(" ", "_", "-", "_").Replace(value)
	for strings.Contains(value, "__") {
		value = strings.ReplaceAll(value, "__", "_")
	}
	return strings.Trim(value, "_")
}

func sharedSessionBrowserPrimaryActionForActionKind(kind string) string {
	label := sharedSessionBrowserActionLabelForKind(kind)
	if label == "" {
		return ""
	}
	return "browser action=" + label
}

func sharedSessionBrowserActionLabelForKind(kind string) string {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "":
		return ""
	case "list_tabs":
		return "tabs"
	case "focus_tab":
		return "focus"
	case "close_tab":
		return "close"
	case "save_pdf":
		return "pdf"
	default:
		return strings.ToLower(strings.TrimSpace(kind))
	}
}
