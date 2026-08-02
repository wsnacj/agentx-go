package tools

import (
	"encoding/json"
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
	agentxmedia "github.com/wsnacj/agentx-go/runtime/mediaartifact"
)

type browserRegistrationRouteCapabilityPayload struct {
	CapabilityMetadata    browserRuntimeCapabilityMetadata `json:"-"`
	DefaultCandidateRoute browserRuntimeRouteDescriptor    `json:"default_candidate_route,omitempty"`
	BrowserTools          []string                         `json:"browser_tools,omitempty"`
	ArtifactTools         []string                         `json:"artifact_tools,omitempty"`
	ArtifactKinds         []string                         `json:"artifact_kinds,omitempty"`
	ArtifactContract      string                           `json:"artifact_contract,omitempty"`
	BrowserActKinds       []string                         `json:"browser_act_kinds,omitempty"`
}

func browserRegistrationPayloadCapabilityMetadata(
	ctx browserRegistrationContext,
	capabilities BrowserCapabilities,
	browserSurface string,
	browserOptInTargets []string,
) browserRuntimeCapabilityMetadata {
	metadata := browserRuntimeCapabilityMetadataForCapabilities(ctx, capabilities)
	metadata.BrowserSurface = strings.TrimSpace(browserSurface)
	metadata.BrowserOptInTargets = append([]string(nil), browserOptInTargets...)
	return metadata
}

func browserApplyRegistrationRouteCapabilityMetadata(
	payload *browserRegistrationRouteCapabilityPayload,
	browserSurface *string,
	browserOptInTargets *[]string,
) {
	if payload == nil {
		return
	}
	metadata := payload.CapabilityMetadata
	payload.BrowserTools = mergeToolMetadataStrings(nil, metadata.BrowserTools)
	payload.ArtifactTools = mergeToolMetadataStrings(nil, metadata.ArtifactTools)
	payload.ArtifactKinds = mergeToolMetadataStrings(nil, metadata.ArtifactKinds)
	payload.ArtifactContract = firstNonEmpty(strings.TrimSpace(metadata.ArtifactContract), strings.TrimSpace(payload.ArtifactContract))
	payload.BrowserActKinds = mergeToolMetadataStrings(nil, metadata.BrowserActKinds)
	if browserSurface != nil {
		*browserSurface = firstNonEmpty(strings.TrimSpace(metadata.BrowserSurface), strings.TrimSpace(*browserSurface))
	}
	if browserOptInTargets != nil {
		switch {
		case len(metadata.BrowserOptInTargets) > 0:
			*browserOptInTargets = append([]string(nil), metadata.BrowserOptInTargets...)
		case len(*browserOptInTargets) > 0:
			*browserOptInTargets = append([]string(nil), (*browserOptInTargets)...)
		default:
			*browserOptInTargets = nil
		}
	}
}

func browserApplyDefaultCandidateRouteToRegistrationPayloadShells(
	explanation **browserTopLevelSummary,
	diagnostics **browserTopLevelSummary,
	summary **browserTopLevelSummary,
	display **browserTopLevelDisplaySummary,
	review **browserReviewSurfaceSummary,
	surface **browserTopLevelSurfaceSummary,
	view **browserTopLevelViewSummary,
) browserRuntimeRouteDescriptor {
	route := firstBrowserRuntimeRouteDescriptor(
		browserRuntimeRouteDescriptorFromTopLevelSummary(derefTopLevelSummaryPtr(explanation)),
		browserRuntimeRouteDescriptorFromTopLevelSummary(derefTopLevelSummaryPtr(diagnostics)),
		browserRuntimeRouteDescriptorFromTopLevelSummary(derefTopLevelSummaryPtr(summary)),
		browserRuntimeRouteDescriptorFromTopLevelDisplay(derefTopLevelDisplaySummaryPtr(display)),
		browserRuntimeRouteDescriptorFromReviewSurface(derefReviewSurfaceSummaryPtr(review)),
		browserRuntimeRouteDescriptorFromTopLevelSurface(derefTopLevelSurfaceSummaryPtr(surface)),
		browserRuntimeRouteDescriptorFromTopLevelView(derefTopLevelViewSummaryPtr(view)),
	)
	if route == (browserRuntimeRouteDescriptor{}) {
		return browserRuntimeRouteDescriptor{}
	}
	if explanation != nil {
		*explanation = browserTopLevelSummaryWithDefaultCandidateRoute(*explanation, route)
	}
	if diagnostics != nil {
		*diagnostics = browserTopLevelSummaryWithDefaultCandidateRoute(*diagnostics, route)
	}
	if summary != nil {
		*summary = browserTopLevelSummaryWithDefaultCandidateRoute(*summary, route)
	}
	if display != nil {
		*display = browserTopLevelDisplayWithDefaultCandidateRoute(*display, route)
	}
	if review != nil {
		*review = browserReviewSurfaceSummaryWithDefaultCandidateRoute(*review, route)
	}
	if surface != nil {
		*surface = browserTopLevelSurfaceWithDefaultCandidateRoute(*surface, route)
	}
	if view != nil {
		*view = browserTopLevelViewWithDefaultCandidateRoute(*view, route)
	}
	return route
}

func derefTopLevelSummaryPtr(summary **browserTopLevelSummary) *browserTopLevelSummary {
	if summary == nil {
		return nil
	}
	return *summary
}

func derefTopLevelDisplaySummaryPtr(display **browserTopLevelDisplaySummary) *browserTopLevelDisplaySummary {
	if display == nil {
		return nil
	}
	return *display
}

func derefReviewSurfaceSummaryPtr(review **browserReviewSurfaceSummary) *browserReviewSurfaceSummary {
	if review == nil {
		return nil
	}
	return *review
}

func derefTopLevelSurfaceSummaryPtr(surface **browserTopLevelSurfaceSummary) *browserTopLevelSurfaceSummary {
	if surface == nil {
		return nil
	}
	return *surface
}

func derefTopLevelViewSummaryPtr(view **browserTopLevelViewSummary) *browserTopLevelViewSummary {
	if view == nil {
		return nil
	}
	return *view
}

type browserTabsToolPayload struct {
	browserRegistrationRouteCapabilityPayload
	Backend                 string                                              `json:"backend"`
	BrowserApp              string                                              `json:"browser_app,omitempty"`
	Profile                 string                                              `json:"profile,omitempty"`
	RuntimeTarget           string                                              `json:"runtime_target,omitempty"`
	Action                  string                                              `json:"action"`
	Status                  string                                              `json:"status"`
	Force                   bool                                                `json:"force,omitempty"`
	ReviewDecision          string                                              `json:"review_decision,omitempty"`
	ReviewReady             bool                                                `json:"review_ready,omitempty"`
	Target                  string                                              `json:"target,omitempty"`
	TargetID                string                                              `json:"target_id,omitempty"`
	TabIndex                int                                                 `json:"tab_index,omitempty"`
	ActiveIndex             int                                                 `json:"active_index,omitempty"`
	Tabs                    []BrowserTab                                        `json:"tabs,omitempty"`
	RememberDecision        string                                              `json:"remember_target_decision,omitempty"`
	RememberReady           bool                                                `json:"remember_target_ready,omitempty"`
	SessionProfileSelection *browserRuntimeSessionProfileSelection              `json:"session_profile_selection,omitempty"`
	SessionTargetSelection  *agentxbrowserruntime.BrowserSessionTargetSelection `json:"session_target_selection,omitempty"`
	Explanation             *browserTopLevelSummary                             `json:"explanation,omitempty"`
	DiagnosticsExplanation  *browserDiagnosticsExplanationSummary               `json:"diagnostics_explanation,omitempty"`
	Diagnostics             *browserTopLevelSummary                             `json:"diagnostics,omitempty"`
	Summary                 *browserTopLevelSummary                             `json:"summary,omitempty"`
	Display                 *browserTopLevelDisplaySummary                      `json:"display,omitempty"`
	Review                  *browserReviewSurfaceSummary                        `json:"review,omitempty"`
	Surface                 *browserTopLevelSurfaceSummary                      `json:"surface,omitempty"`
	View                    *browserTopLevelViewSummary                         `json:"view,omitempty"`
	BrowserSurface          string                                              `json:"browser_surface,omitempty"`
	BrowserOptInTargets     []string                                            `json:"browser_opt_in_targets,omitempty"`
	WaitMs                  int                                                 `json:"wait_ms,omitempty"`
	Note                    string                                              `json:"note,omitempty"`
}

type browserActToolPayload struct {
	browserRegistrationRouteCapabilityPayload
	Kind                              string                                                 `json:"kind"`
	Action                            string                                                 `json:"action,omitempty"`
	URL                               string                                                 `json:"url,omitempty"`
	Paths                             []string                                               `json:"paths,omitempty"`
	FilesTouched                      []string                                               `json:"files_touched,omitempty"`
	FinalURL                          string                                                 `json:"final_url,omitempty"`
	Backend                           string                                                 `json:"backend,omitempty"`
	BrowserApp                        string                                                 `json:"browser_app,omitempty"`
	Profile                           string                                                 `json:"profile,omitempty"`
	RuntimeTarget                     string                                                 `json:"runtime_target,omitempty"`
	Title                             string                                                 `json:"title,omitempty"`
	Content                           string                                                 `json:"content,omitempty"`
	ContentType                       string                                                 `json:"content_type,omitempty"`
	Snapshot                          string                                                 `json:"snapshot,omitempty"`
	SnapshotFormat                    string                                                 `json:"snapshot_format,omitempty"`
	SnapshotMode                      string                                                 `json:"snapshot_mode,omitempty"`
	SnapshotRefs                      string                                                 `json:"snapshot_refs,omitempty"`
	SnapshotInteractive               bool                                                   `json:"snapshot_interactive,omitempty"`
	SnapshotCompact                   bool                                                   `json:"snapshot_compact,omitempty"`
	SnapshotDepth                     int                                                    `json:"snapshot_depth,omitempty"`
	SnapshotFrame                     string                                                 `json:"snapshot_frame,omitempty"`
	Elements                          []BrowserSnapshotElement                               `json:"elements,omitempty"`
	Messages                          []BrowserConsoleMessage                                `json:"messages,omitempty"`
	Requests                          []BrowserRequestEntry                                  `json:"requests,omitempty"`
	RequestURL                        string                                                 `json:"request_url,omitempty"`
	RequestMethod                     string                                                 `json:"request_method,omitempty"`
	ResponseStatusCode                int                                                    `json:"response_status_code,omitempty"`
	Errors                            []BrowserErrorEntry                                    `json:"errors,omitempty"`
	Cookies                           []BrowserCookieEntry                                   `json:"cookies,omitempty"`
	StorageKind                       string                                                 `json:"storage_kind,omitempty"`
	Storage                           []BrowserStorageEntry                                  `json:"storage,omitempty"`
	HeaderNames                       []string                                               `json:"header_names,omitempty"`
	HeaderCount                       int                                                    `json:"header_count,omitempty"`
	Path                              string                                                 `json:"path,omitempty"`
	Bytes                             int64                                                  `json:"bytes,omitempty"`
	Media                             *agentxmedia.Descriptor                                `json:"media,omitempty"`
	Artifacts                         []browserArtifactPayload                               `json:"artifacts,omitempty"`
	CaptureScope                      string                                                 `json:"capture_scope,omitempty"`
	CaptureWidth                      int                                                    `json:"capture_width,omitempty"`
	CaptureHeight                     int                                                    `json:"capture_height,omitempty"`
	Key                               string                                                 `json:"key,omitempty"`
	Result                            string                                                 `json:"result,omitempty"`
	Value                             string                                                 `json:"value,omitempty"`
	Values                            []string                                               `json:"values,omitempty"`
	FieldCount                        int                                                    `json:"field_count,omitempty"`
	Width                             int                                                    `json:"width,omitempty"`
	Height                            int                                                    `json:"height,omitempty"`
	Status                            string                                                 `json:"status"`
	Force                             bool                                                   `json:"force,omitempty"`
	ReviewDecision                    string                                                 `json:"review_decision,omitempty"`
	ReviewReady                       bool                                                   `json:"review_ready,omitempty"`
	Target                            string                                                 `json:"target,omitempty"`
	TargetID                          string                                                 `json:"target_id,omitempty"`
	RememberDecision                  string                                                 `json:"remember_target_decision,omitempty"`
	RememberReady                     bool                                                   `json:"remember_target_ready,omitempty"`
	SessionProfileSelection           *browserRuntimeSessionProfileSelection                 `json:"session_profile_selection,omitempty"`
	SessionTargetSelection            *agentxbrowserruntime.BrowserSessionTargetSelection    `json:"session_target_selection,omitempty"`
	ResolverOutcome                   *agentxbrowserruntime.BrowserElementResolverOutcome    `json:"resolver_outcome,omitempty"`
	Actionability                     *agentxbrowserruntime.BrowserActionabilityReport       `json:"actionability,omitempty"`
	FailureEvidence                   *agentxbrowserruntime.BrowserActionFailureEvidence     `json:"failure_evidence,omitempty"`
	ResolvedViaFallback               bool                                                   `json:"resolved_via_fallback,omitempty"`
	ResolverFallbackKind              string                                                 `json:"resolver_fallback_kind,omitempty"`
	ResolverFallbackIndex             *int                                                   `json:"resolver_fallback_index,omitempty"`
	ResolverFallbackStrength          string                                                 `json:"resolver_fallback_candidate_strength,omitempty"`
	ResolverFallbackBlockedBy         string                                                 `json:"resolver_fallback_blocked_by,omitempty"`
	ResolverFallbackAmbiguityClass    string                                                 `json:"resolver_fallback_ambiguity_class,omitempty"`
	ResolverFallbackManualRetryHint   string                                                 `json:"resolver_fallback_manual_retry_hint,omitempty"`
	ResolverFallbackSpecificityFields []string                                               `json:"resolver_fallback_specificity_fields,omitempty"`
	ResolverFallbackExplanation       *browserResolverFallbackExplanationSummary             `json:"resolver_fallback_explanation,omitempty"`
	ResolverExplanation               *browserRuntimeResolverExplanationSummary              `json:"resolver_explanation,omitempty"`
	ResolverBlockedBy                 string                                                 `json:"resolver_blocked_by,omitempty"`
	ResolverAmbiguityClass            string                                                 `json:"resolver_ambiguity_class,omitempty"`
	ResolverCandidateKind             string                                                 `json:"resolver_candidate_kind,omitempty"`
	ResolverCandidateStrength         string                                                 `json:"resolver_candidate_strength,omitempty"`
	ResolverRetryDisposition          string                                                 `json:"resolver_retry_disposition,omitempty"`
	ResolverManualRetryHint           string                                                 `json:"resolver_manual_retry_hint,omitempty"`
	ResolverNextStepAlias             string                                                 `json:"resolver_next_step_alias,omitempty"`
	BrowserLocalPlanner               *agentxbrowserruntime.BrowserLocalPlannerResultSummary `json:"browser_local_planner,omitempty"`
	DiagnosticsExplanation            *browserDiagnosticsExplanationSummary                  `json:"diagnostics_explanation,omitempty"`
	Explanation                       *browserTopLevelSummary                                `json:"explanation,omitempty"`
	Diagnostics                       *browserTopLevelSummary                                `json:"diagnostics,omitempty"`
	Summary                           *browserTopLevelSummary                                `json:"summary,omitempty"`
	Display                           *browserTopLevelDisplaySummary                         `json:"display,omitempty"`
	Surface                           *browserTopLevelSurfaceSummary                         `json:"surface,omitempty"`
	Review                            *browserReviewSurfaceSummary                           `json:"review,omitempty"`
	View                              *browserTopLevelViewSummary                            `json:"view,omitempty"`
	PostNavigationSnapshot            *browserPostNavigationSnapshotRecommendation           `json:"post_navigation_snapshot,omitempty"`
	BotDetectionWarning               *browserBotDetectionWarning                            `json:"bot_detection_warning,omitempty"`
	SafetyEvent                       *browserResultSafetyEvent                              `json:"safety_event,omitempty"`
	BrowserSurface                    string                                                 `json:"browser_surface,omitempty"`
	BrowserOptInTargets               []string                                               `json:"browser_opt_in_targets,omitempty"`
	RecoveryAction                    string                                                 `json:"recovery_action,omitempty"`
	Submitted                         bool                                                   `json:"submitted,omitempty"`
	Truncated                         bool                                                   `json:"truncated,omitempty"`
	TabIndex                          int                                                    `json:"tab_index,omitempty"`
	ActiveIndex                       int                                                    `json:"active_index,omitempty"`
	Tabs                              []BrowserTab                                           `json:"tabs,omitempty"`
	WaitMs                            int                                                    `json:"wait_ms,omitempty"`
	PostWaitMs                        int                                                    `json:"post_wait_ms,omitempty"`
	Ref                               string                                                 `json:"ref,omitempty"`
	Selector                          string                                                 `json:"selector,omitempty"`
	FullPage                          bool                                                   `json:"full_page,omitempty"`
	Note                              string                                                 `json:"note,omitempty"`
}

type browserPageActionReviewBlockedPayload struct {
	browserRegistrationRouteCapabilityPayload
	Path                   string                                `json:"path,omitempty"`
	Backend                string                                `json:"backend"`
	BrowserApp             string                                `json:"browser_app,omitempty"`
	Profile                string                                `json:"profile,omitempty"`
	RuntimeTarget          string                                `json:"runtime_target,omitempty"`
	Ref                    string                                `json:"ref,omitempty"`
	Selector               string                                `json:"selector,omitempty"`
	Value                  string                                `json:"value,omitempty"`
	Status                 string                                `json:"status"`
	Force                  bool                                  `json:"force,omitempty"`
	ReviewDecision         string                                `json:"review_decision,omitempty"`
	ReviewReady            bool                                  `json:"review_ready,omitempty"`
	Explanation            *browserTopLevelSummary               `json:"explanation,omitempty"`
	DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation,omitempty"`
	Diagnostics            *browserTopLevelSummary               `json:"diagnostics,omitempty"`
	Summary                *browserTopLevelSummary               `json:"summary,omitempty"`
	Display                *browserTopLevelDisplaySummary        `json:"display,omitempty"`
	Surface                *browserTopLevelSurfaceSummary        `json:"surface,omitempty"`
	Review                 *browserReviewSurfaceSummary          `json:"review,omitempty"`
	View                   *browserTopLevelViewSummary           `json:"view,omitempty"`
	BrowserSurface         string                                `json:"browser_surface,omitempty"`
	BrowserOptInTargets    []string                              `json:"browser_opt_in_targets,omitempty"`
	Target                 string                                `json:"target,omitempty"`
	TargetID               string                                `json:"target_id,omitempty"`
	TabIndex               int                                   `json:"tab_index,omitempty"`
	WaitMs                 int                                   `json:"wait_ms,omitempty"`
	PostWaitMs             int                                   `json:"post_wait_ms,omitempty"`
	Note                   string                                `json:"note,omitempty"`
}

type browserOpenToolPayload struct {
	browserRegistrationRouteCapabilityPayload
	URL                    string                                `json:"url"`
	Backend                string                                `json:"backend"`
	BrowserApp             string                                `json:"browser_app,omitempty"`
	Profile                string                                `json:"profile,omitempty"`
	RuntimeTarget          string                                `json:"runtime_target,omitempty"`
	Target                 string                                `json:"target,omitempty"`
	TargetID               string                                `json:"target_id,omitempty"`
	Status                 string                                `json:"status"`
	Explanation            *browserTopLevelSummary               `json:"explanation,omitempty"`
	DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation,omitempty"`
	Diagnostics            *browserTopLevelSummary               `json:"diagnostics,omitempty"`
	Summary                *browserTopLevelSummary               `json:"summary,omitempty"`
	Display                *browserTopLevelDisplaySummary        `json:"display,omitempty"`
	Review                 *browserReviewSurfaceSummary          `json:"review,omitempty"`
	Surface                *browserTopLevelSurfaceSummary        `json:"surface,omitempty"`
	View                   *browserTopLevelViewSummary           `json:"view,omitempty"`
	BrowserSurface         string                                `json:"browser_surface,omitempty"`
	BrowserOptInTargets    []string                              `json:"browser_opt_in_targets,omitempty"`
	WaitMs                 int                                   `json:"wait_ms"`
	Note                   string                                `json:"note,omitempty"`
}

type browserNavigateToolPayload struct {
	browserRegistrationRouteCapabilityPayload
	URL                    string                                       `json:"url"`
	FinalURL               string                                       `json:"final_url,omitempty"`
	Backend                string                                       `json:"backend"`
	BrowserApp             string                                       `json:"browser_app,omitempty"`
	Profile                string                                       `json:"profile,omitempty"`
	RuntimeTarget          string                                       `json:"runtime_target,omitempty"`
	Title                  string                                       `json:"title,omitempty"`
	Target                 string                                       `json:"target,omitempty"`
	TargetID               string                                       `json:"target_id,omitempty"`
	Status                 string                                       `json:"status"`
	Force                  bool                                         `json:"force,omitempty"`
	ReviewDecision         string                                       `json:"review_decision,omitempty"`
	ReviewReady            bool                                         `json:"review_ready,omitempty"`
	Explanation            *browserTopLevelSummary                      `json:"explanation,omitempty"`
	DiagnosticsExplanation *browserDiagnosticsExplanationSummary        `json:"diagnostics_explanation,omitempty"`
	Diagnostics            *browserTopLevelSummary                      `json:"diagnostics,omitempty"`
	Summary                *browserTopLevelSummary                      `json:"summary,omitempty"`
	Display                *browserTopLevelDisplaySummary               `json:"display,omitempty"`
	Surface                *browserTopLevelSurfaceSummary               `json:"surface,omitempty"`
	Review                 *browserReviewSurfaceSummary                 `json:"review,omitempty"`
	View                   *browserTopLevelViewSummary                  `json:"view,omitempty"`
	PostNavigationSnapshot *browserPostNavigationSnapshotRecommendation `json:"post_navigation_snapshot,omitempty"`
	BotDetectionWarning    *browserBotDetectionWarning                  `json:"bot_detection_warning,omitempty"`
	BrowserSurface         string                                       `json:"browser_surface,omitempty"`
	BrowserOptInTargets    []string                                     `json:"browser_opt_in_targets,omitempty"`
	TabIndex               int                                          `json:"tab_index,omitempty"`
	WaitMs                 int                                          `json:"wait_ms"`
	Note                   string                                       `json:"note,omitempty"`
}

type browserExtractToolPayload struct {
	browserRegistrationRouteCapabilityPayload
	URL                    string                                `json:"url,omitempty"`
	FinalURL               string                                `json:"final_url,omitempty"`
	Backend                string                                `json:"backend"`
	BrowserApp             string                                `json:"browser_app,omitempty"`
	Profile                string                                `json:"profile,omitempty"`
	RuntimeTarget          string                                `json:"runtime_target,omitempty"`
	Title                  string                                `json:"title,omitempty"`
	Content                string                                `json:"content"`
	ContentType            string                                `json:"content_type,omitempty"`
	Status                 string                                `json:"status"`
	Force                  bool                                  `json:"force,omitempty"`
	ReviewDecision         string                                `json:"review_decision,omitempty"`
	ReviewReady            bool                                  `json:"review_ready,omitempty"`
	Explanation            *browserTopLevelSummary               `json:"explanation,omitempty"`
	DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation,omitempty"`
	Diagnostics            *browserTopLevelSummary               `json:"diagnostics,omitempty"`
	Summary                *browserTopLevelSummary               `json:"summary,omitempty"`
	Display                *browserTopLevelDisplaySummary        `json:"display,omitempty"`
	Surface                *browserTopLevelSurfaceSummary        `json:"surface,omitempty"`
	Review                 *browserReviewSurfaceSummary          `json:"review,omitempty"`
	View                   *browserTopLevelViewSummary           `json:"view,omitempty"`
	BrowserSurface         string                                `json:"browser_surface,omitempty"`
	BrowserOptInTargets    []string                              `json:"browser_opt_in_targets,omitempty"`
	Truncated              bool                                  `json:"truncated"`
	Target                 string                                `json:"target,omitempty"`
	TargetID               string                                `json:"target_id,omitempty"`
	TabIndex               int                                   `json:"tab_index,omitempty"`
	WaitMs                 int                                   `json:"wait_ms"`
	Note                   string                                `json:"note,omitempty"`
}

type browserScreenshotToolPayload struct {
	browserRegistrationRouteCapabilityPayload
	URL                               string                                              `json:"url,omitempty"`
	FinalURL                          string                                              `json:"final_url,omitempty"`
	Title                             string                                              `json:"title,omitempty"`
	Path                              string                                              `json:"path"`
	FilesTouched                      []string                                            `json:"files_touched,omitempty"`
	Bytes                             int64                                               `json:"bytes"`
	Media                             *agentxmedia.Descriptor                             `json:"media,omitempty"`
	Artifacts                         []browserArtifactPayload                            `json:"artifacts,omitempty"`
	Backend                           string                                              `json:"backend"`
	BrowserApp                        string                                              `json:"browser_app,omitempty"`
	Profile                           string                                              `json:"profile,omitempty"`
	RuntimeTarget                     string                                              `json:"runtime_target,omitempty"`
	CaptureScope                      string                                              `json:"capture_scope,omitempty"`
	CaptureWidth                      int                                                 `json:"capture_width,omitempty"`
	CaptureHeight                     int                                                 `json:"capture_height,omitempty"`
	Status                            string                                              `json:"status"`
	ResolverOutcome                   *agentxbrowserruntime.BrowserElementResolverOutcome `json:"resolver_outcome,omitempty"`
	Actionability                     *agentxbrowserruntime.BrowserActionabilityReport    `json:"actionability,omitempty"`
	FailureEvidence                   *agentxbrowserruntime.BrowserActionFailureEvidence  `json:"failure_evidence,omitempty"`
	ResolvedViaFallback               bool                                                `json:"resolved_via_fallback,omitempty"`
	ResolverFallbackKind              string                                              `json:"resolver_fallback_kind,omitempty"`
	ResolverFallbackIndex             *int                                                `json:"resolver_fallback_index,omitempty"`
	ResolverFallbackStrength          string                                              `json:"resolver_fallback_candidate_strength,omitempty"`
	ResolverFallbackBlockedBy         string                                              `json:"resolver_fallback_blocked_by,omitempty"`
	ResolverFallbackAmbiguityClass    string                                              `json:"resolver_fallback_ambiguity_class,omitempty"`
	ResolverFallbackManualRetryHint   string                                              `json:"resolver_fallback_manual_retry_hint,omitempty"`
	ResolverFallbackSpecificityFields []string                                            `json:"resolver_fallback_specificity_fields,omitempty"`
	ResolverFallbackExplanation       *browserResolverFallbackExplanationSummary          `json:"resolver_fallback_explanation,omitempty"`
	ResolverExplanation               *browserRuntimeResolverExplanationSummary           `json:"resolver_explanation,omitempty"`
	DiagnosticsExplanation            *browserDiagnosticsExplanationSummary               `json:"diagnostics_explanation,omitempty"`
	Explanation                       *browserTopLevelSummary                             `json:"explanation,omitempty"`
	Diagnostics                       *browserTopLevelSummary                             `json:"diagnostics,omitempty"`
	Summary                           *browserTopLevelSummary                             `json:"summary,omitempty"`
	Display                           *browserTopLevelDisplaySummary                      `json:"display,omitempty"`
	Surface                           *browserTopLevelSurfaceSummary                      `json:"surface,omitempty"`
	Review                            *browserReviewSurfaceSummary                        `json:"review,omitempty"`
	View                              *browserTopLevelViewSummary                         `json:"view,omitempty"`
	BrowserSurface                    string                                              `json:"browser_surface,omitempty"`
	BrowserOptInTargets               []string                                            `json:"browser_opt_in_targets,omitempty"`
	RecoveryAction                    string                                              `json:"recovery_action,omitempty"`
	Force                             bool                                                `json:"force,omitempty"`
	ReviewDecision                    string                                              `json:"review_decision,omitempty"`
	ReviewReady                       bool                                                `json:"review_ready,omitempty"`
	Target                            string                                              `json:"target,omitempty"`
	TargetID                          string                                              `json:"target_id,omitempty"`
	TabIndex                          int                                                 `json:"tab_index,omitempty"`
	Ref                               string                                              `json:"ref,omitempty"`
	Selector                          string                                              `json:"selector,omitempty"`
	FullPage                          bool                                                `json:"full_page,omitempty"`
	WaitMs                            int                                                 `json:"wait_ms"`
	Note                              string                                              `json:"note,omitempty"`
}

type browserClickToolPayload struct {
	browserRegistrationRouteCapabilityPayload
	URL                               string                                              `json:"url,omitempty"`
	FinalURL                          string                                              `json:"final_url,omitempty"`
	Backend                           string                                              `json:"backend"`
	BrowserApp                        string                                              `json:"browser_app,omitempty"`
	Profile                           string                                              `json:"profile,omitempty"`
	RuntimeTarget                     string                                              `json:"runtime_target,omitempty"`
	Ref                               string                                              `json:"ref,omitempty"`
	Selector                          string                                              `json:"selector"`
	Title                             string                                              `json:"title,omitempty"`
	Snapshot                          string                                              `json:"snapshot,omitempty"`
	SnapshotFormat                    string                                              `json:"snapshot_format,omitempty"`
	SnapshotMode                      string                                              `json:"snapshot_mode,omitempty"`
	SnapshotRefs                      string                                              `json:"snapshot_refs,omitempty"`
	SnapshotFrame                     string                                              `json:"snapshot_frame,omitempty"`
	Elements                          []BrowserSnapshotElement                            `json:"elements,omitempty"`
	Status                            string                                              `json:"status"`
	RecoveryAction                    string                                              `json:"recovery_action,omitempty"`
	Force                             bool                                                `json:"force,omitempty"`
	ReviewDecision                    string                                              `json:"review_decision,omitempty"`
	ReviewReady                       bool                                                `json:"review_ready,omitempty"`
	Target                            string                                              `json:"target,omitempty"`
	TargetID                          string                                              `json:"target_id,omitempty"`
	TabIndex                          int                                                 `json:"tab_index,omitempty"`
	WaitMs                            int                                                 `json:"wait_ms"`
	PostWaitMs                        int                                                 `json:"post_wait_ms"`
	SnapshotInteractive               bool                                                `json:"snapshot_interactive,omitempty"`
	SnapshotCompact                   bool                                                `json:"snapshot_compact,omitempty"`
	SnapshotDepth                     int                                                 `json:"snapshot_depth,omitempty"`
	Truncated                         bool                                                `json:"truncated,omitempty"`
	ResolverOutcome                   *agentxbrowserruntime.BrowserElementResolverOutcome `json:"resolver_outcome,omitempty"`
	Actionability                     *agentxbrowserruntime.BrowserActionabilityReport    `json:"actionability,omitempty"`
	FailureEvidence                   *agentxbrowserruntime.BrowserActionFailureEvidence  `json:"failure_evidence,omitempty"`
	ResolvedViaFallback               bool                                                `json:"resolved_via_fallback,omitempty"`
	ResolverFallbackKind              string                                              `json:"resolver_fallback_kind,omitempty"`
	ResolverFallbackIndex             *int                                                `json:"resolver_fallback_index,omitempty"`
	ResolverFallbackStrength          string                                              `json:"resolver_fallback_candidate_strength,omitempty"`
	ResolverFallbackBlockedBy         string                                              `json:"resolver_fallback_blocked_by,omitempty"`
	ResolverFallbackAmbiguityClass    string                                              `json:"resolver_fallback_ambiguity_class,omitempty"`
	ResolverFallbackManualRetryHint   string                                              `json:"resolver_fallback_manual_retry_hint,omitempty"`
	ResolverFallbackSpecificityFields []string                                            `json:"resolver_fallback_specificity_fields,omitempty"`
	ResolverFallbackExplanation       *browserResolverFallbackExplanationSummary          `json:"resolver_fallback_explanation,omitempty"`
	ResolverExplanation               *browserRuntimeResolverExplanationSummary           `json:"resolver_explanation,omitempty"`
	DiagnosticsExplanation            *browserDiagnosticsExplanationSummary               `json:"diagnostics_explanation,omitempty"`
	Explanation                       *browserTopLevelSummary                             `json:"explanation,omitempty"`
	Diagnostics                       *browserTopLevelSummary                             `json:"diagnostics,omitempty"`
	Summary                           *browserTopLevelSummary                             `json:"summary,omitempty"`
	Display                           *browserTopLevelDisplaySummary                      `json:"display,omitempty"`
	Surface                           *browserTopLevelSurfaceSummary                      `json:"surface,omitempty"`
	Review                            *browserReviewSurfaceSummary                        `json:"review,omitempty"`
	View                              *browserTopLevelViewSummary                         `json:"view,omitempty"`
	BrowserSurface                    string                                              `json:"browser_surface,omitempty"`
	BrowserOptInTargets               []string                                            `json:"browser_opt_in_targets,omitempty"`
	Note                              string                                              `json:"note,omitempty"`
}

type browserTypeToolPayload struct {
	browserRegistrationRouteCapabilityPayload
	URL                               string                                              `json:"url,omitempty"`
	FinalURL                          string                                              `json:"final_url,omitempty"`
	Backend                           string                                              `json:"backend"`
	BrowserApp                        string                                              `json:"browser_app,omitempty"`
	Profile                           string                                              `json:"profile,omitempty"`
	RuntimeTarget                     string                                              `json:"runtime_target,omitempty"`
	Ref                               string                                              `json:"ref,omitempty"`
	Selector                          string                                              `json:"selector"`
	Title                             string                                              `json:"title,omitempty"`
	Value                             string                                              `json:"value"`
	Snapshot                          string                                              `json:"snapshot,omitempty"`
	SnapshotFormat                    string                                              `json:"snapshot_format,omitempty"`
	SnapshotMode                      string                                              `json:"snapshot_mode,omitempty"`
	SnapshotRefs                      string                                              `json:"snapshot_refs,omitempty"`
	SnapshotFrame                     string                                              `json:"snapshot_frame,omitempty"`
	Elements                          []BrowserSnapshotElement                            `json:"elements,omitempty"`
	Status                            string                                              `json:"status"`
	Submitted                         bool                                                `json:"submitted"`
	ResolverOutcome                   *agentxbrowserruntime.BrowserElementResolverOutcome `json:"resolver_outcome,omitempty"`
	Actionability                     *agentxbrowserruntime.BrowserActionabilityReport    `json:"actionability,omitempty"`
	FailureEvidence                   *agentxbrowserruntime.BrowserActionFailureEvidence  `json:"failure_evidence,omitempty"`
	ResolvedViaFallback               bool                                                `json:"resolved_via_fallback,omitempty"`
	ResolverFallbackKind              string                                              `json:"resolver_fallback_kind,omitempty"`
	ResolverFallbackIndex             *int                                                `json:"resolver_fallback_index,omitempty"`
	ResolverFallbackStrength          string                                              `json:"resolver_fallback_candidate_strength,omitempty"`
	ResolverFallbackBlockedBy         string                                              `json:"resolver_fallback_blocked_by,omitempty"`
	ResolverFallbackAmbiguityClass    string                                              `json:"resolver_fallback_ambiguity_class,omitempty"`
	ResolverFallbackManualRetryHint   string                                              `json:"resolver_fallback_manual_retry_hint,omitempty"`
	ResolverFallbackSpecificityFields []string                                            `json:"resolver_fallback_specificity_fields,omitempty"`
	ResolverFallbackExplanation       *browserResolverFallbackExplanationSummary          `json:"resolver_fallback_explanation,omitempty"`
	ResolverExplanation               *browserRuntimeResolverExplanationSummary           `json:"resolver_explanation,omitempty"`
	DiagnosticsExplanation            *browserDiagnosticsExplanationSummary               `json:"diagnostics_explanation,omitempty"`
	Explanation                       *browserTopLevelSummary                             `json:"explanation,omitempty"`
	Diagnostics                       *browserTopLevelSummary                             `json:"diagnostics,omitempty"`
	Summary                           *browserTopLevelSummary                             `json:"summary,omitempty"`
	Display                           *browserTopLevelDisplaySummary                      `json:"display,omitempty"`
	Surface                           *browserTopLevelSurfaceSummary                      `json:"surface,omitempty"`
	Review                            *browserReviewSurfaceSummary                        `json:"review,omitempty"`
	View                              *browserTopLevelViewSummary                         `json:"view,omitempty"`
	BrowserSurface                    string                                              `json:"browser_surface,omitempty"`
	BrowserOptInTargets               []string                                            `json:"browser_opt_in_targets,omitempty"`
	RecoveryAction                    string                                              `json:"recovery_action,omitempty"`
	Force                             bool                                                `json:"force,omitempty"`
	ReviewDecision                    string                                              `json:"review_decision,omitempty"`
	ReviewReady                       bool                                                `json:"review_ready,omitempty"`
	Target                            string                                              `json:"target,omitempty"`
	TargetID                          string                                              `json:"target_id,omitempty"`
	TabIndex                          int                                                 `json:"tab_index,omitempty"`
	WaitMs                            int                                                 `json:"wait_ms"`
	PostWaitMs                        int                                                 `json:"post_wait_ms"`
	SnapshotInteractive               bool                                                `json:"snapshot_interactive,omitempty"`
	SnapshotCompact                   bool                                                `json:"snapshot_compact,omitempty"`
	SnapshotDepth                     int                                                 `json:"snapshot_depth,omitempty"`
	Truncated                         bool                                                `json:"truncated,omitempty"`
	Note                              string                                              `json:"note,omitempty"`
}

type browserEvalToolPayload struct {
	browserRegistrationRouteCapabilityPayload
	URL                    string                                `json:"url,omitempty"`
	FinalURL               string                                `json:"final_url,omitempty"`
	Backend                string                                `json:"backend"`
	BrowserApp             string                                `json:"browser_app,omitempty"`
	Profile                string                                `json:"profile,omitempty"`
	RuntimeTarget          string                                `json:"runtime_target,omitempty"`
	Title                  string                                `json:"title,omitempty"`
	Result                 string                                `json:"result"`
	Status                 string                                `json:"status"`
	Force                  bool                                  `json:"force,omitempty"`
	ReviewDecision         string                                `json:"review_decision,omitempty"`
	ReviewReady            bool                                  `json:"review_ready,omitempty"`
	Explanation            *browserTopLevelSummary               `json:"explanation,omitempty"`
	DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation,omitempty"`
	Diagnostics            *browserTopLevelSummary               `json:"diagnostics,omitempty"`
	Summary                *browserTopLevelSummary               `json:"summary,omitempty"`
	Display                *browserTopLevelDisplaySummary        `json:"display,omitempty"`
	Surface                *browserTopLevelSurfaceSummary        `json:"surface,omitempty"`
	Review                 *browserReviewSurfaceSummary          `json:"review,omitempty"`
	View                   *browserTopLevelViewSummary           `json:"view,omitempty"`
	SafetyEvent            *browserResultSafetyEvent             `json:"safety_event,omitempty"`
	BrowserSurface         string                                `json:"browser_surface,omitempty"`
	BrowserOptInTargets    []string                              `json:"browser_opt_in_targets,omitempty"`
	Target                 string                                `json:"target,omitempty"`
	TargetID               string                                `json:"target_id,omitempty"`
	TabIndex               int                                   `json:"tab_index,omitempty"`
	WaitMs                 int                                   `json:"wait_ms"`
	Truncated              bool                                  `json:"truncated"`
	Note                   string                                `json:"note,omitempty"`
}

func marshalBrowserTabsReviewBlockedPayload(runtimeInfo BrowserRuntimeInfo, browserApp string, action string, force bool, target browserToolTarget, waitMs int, review browserPendingTargetReviewState, note string) (string, error) {
	return marshalBrowserTabsReviewBlockedPayloadWithRouteSurface(runtimeInfo, browserApp, action, force, target, waitMs, review, browserRuntimeCapabilityMetadata{}, "", nil, note)
}

func marshalBrowserTabsReviewBlockedPayloadWithRouteSurface(runtimeInfo BrowserRuntimeInfo, browserApp string, action string, force bool, target browserToolTarget, waitMs int, review browserPendingTargetReviewState, capabilityMetadata browserRuntimeCapabilityMetadata, browserSurface string, browserOptInTargets []string, note string) (string, error) {
	return marshalBrowserTabsPayload(browserTabsToolPayload{
		browserRegistrationRouteCapabilityPayload: browserRegistrationRouteCapabilityPayload{
			CapabilityMetadata: capabilityMetadata,
		},
		Backend:             runtimeInfo.Backend,
		BrowserApp:          browserApp,
		Profile:             runtimeInfo.Profile,
		RuntimeTarget:       runtimeInfo.Target,
		Action:              action,
		Status:              "review_required",
		Force:               force,
		ReviewDecision:      browserPendingTargetReviewDecisionWithState(review, force),
		ReviewReady:         false,
		BrowserSurface:      browserSurface,
		BrowserOptInTargets: append([]string(nil), browserOptInTargets...),
		Target:              target.Value,
		TargetID:            strings.TrimSpace(target.TargetID),
		TabIndex:            target.TabIndex,
		WaitMs:              waitMs,
		Note:                strings.TrimSpace(note),
	})
}

func marshalBrowserTabsPayload(payload browserTabsToolPayload) (string, error) {
	browserApplyRegistrationRouteCapabilityMetadata(&payload.browserRegistrationRouteCapabilityPayload, &payload.BrowserSurface, &payload.BrowserOptInTargets)
	reviewDecision, reviewReady := browserTabsTopLevelReviewState(payload)
	reviewStatus := strings.TrimSpace(payload.Status)
	if strings.TrimSpace(reviewDecision) != "" && !reviewReady {
		reviewStatus = "review_required"
	}
	browserApplyActionSuccessSummaryAliases(browserTabsSuccessKind(payload.Action), reviewStatus, &payload.DiagnosticsExplanation, &payload.Explanation, &payload.Diagnostics, &payload.Summary, &payload.Display)
	browserApplyReviewSummaryAliases(reviewStatus, reviewDecision, reviewReady, &payload.DiagnosticsExplanation, &payload.Explanation, &payload.Diagnostics, &payload.Summary, &payload.Display, &payload.Review)
	payload.Surface, payload.View = browserRuntimeTopLevelSurfaceAndViewFromDisplayAndReviewWithMetadata(payload.Display, payload.Review, browserTopLevelCapabilitySurfaceFromMetadata(payload.CapabilityMetadata), payload.BrowserSurface, payload.BrowserOptInTargets)
	payload.DefaultCandidateRoute = browserApplyDefaultCandidateRouteToRegistrationPayloadShells(&payload.Explanation, &payload.Diagnostics, &payload.Summary, &payload.Display, &payload.Review, &payload.Surface, &payload.View)
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func browserTabsTopLevelReviewState(payload browserTabsToolPayload) (string, bool) {
	rememberDecision := ""
	rememberReady := false
	if browserRememberDecisionSurfacesTopLevelReview(payload.RememberDecision, payload.RememberReady) {
		rememberDecision = strings.TrimSpace(payload.RememberDecision)
		rememberReady = payload.RememberReady
	}
	state := agentxbrowserruntime.SelectSharedSessionBrowserReviewState(
		agentxbrowserruntime.SharedSessionBrowserReviewStateRequest{
			Candidates: []agentxbrowserruntime.SharedSessionBrowserReviewDecisionCandidate{
				{Decision: payload.ReviewDecision, Ready: payload.ReviewReady},
				{Decision: rememberDecision, Ready: rememberReady},
			},
		},
	)
	return state.Decision, state.Ready
}

func browserActTopLevelReviewState(payload browserActToolPayload) (string, bool) {
	rememberDecision := ""
	rememberReady := false
	switch strings.TrimSpace(payload.Kind) {
	case "list_tabs", "focus_tab", "close_tab":
		if browserRememberDecisionSurfacesTopLevelReview(payload.RememberDecision, payload.RememberReady) {
			rememberDecision = strings.TrimSpace(payload.RememberDecision)
			rememberReady = payload.RememberReady
		}
	}
	state := agentxbrowserruntime.SelectSharedSessionBrowserReviewState(
		agentxbrowserruntime.SharedSessionBrowserReviewStateRequest{
			Candidates: []agentxbrowserruntime.SharedSessionBrowserReviewDecisionCandidate{
				{Decision: payload.ReviewDecision, Ready: payload.ReviewReady},
				{Decision: rememberDecision, Ready: rememberReady},
			},
		},
	)
	return state.Decision, state.Ready
}

func marshalBrowserActPayload(payload browserActToolPayload) (string, error) {
	browserApplyRegistrationRouteCapabilityMetadata(&payload.browserRegistrationRouteCapabilityPayload, &payload.BrowserSurface, &payload.BrowserOptInTargets)
	browserApplyActionabilityEvidence(
		payload.Kind,
		payload.Status,
		payload.Note,
		payload.RecoveryAction,
		payload.Ref,
		payload.Selector,
		payload.Snapshot,
		payload.SnapshotFormat,
		payload.SnapshotRefs,
		payload.SnapshotFrame,
		payload.Elements,
		payload.Truncated,
		payload.Path,
		payload.FinalURL,
		payload.Title,
		payload.Messages,
		payload.Errors,
		payload.ResolverOutcome,
		&payload.Actionability,
		&payload.FailureEvidence,
	)
	browserApplyResolverFallbackSummary(
		payload.ResolverOutcome,
		&payload.ResolvedViaFallback,
		&payload.ResolverFallbackKind,
		&payload.ResolverFallbackIndex,
		&payload.ResolverFallbackStrength,
		&payload.ResolverFallbackBlockedBy,
		&payload.ResolverFallbackAmbiguityClass,
		&payload.ResolverFallbackManualRetryHint,
		&payload.ResolverFallbackSpecificityFields,
	)
	browserApplyResolverGuidanceSummary(
		payload.ResolverOutcome,
		&payload.ResolverBlockedBy,
		&payload.ResolverAmbiguityClass,
		&payload.ResolverCandidateKind,
		&payload.ResolverCandidateStrength,
		&payload.ResolverRetryDisposition,
		&payload.ResolverManualRetryHint,
		&payload.ResolverNextStepAlias,
	)
	fallbackSummary := browserResolverFallbackSummaryFromFields(
		payload.ResolvedViaFallback,
		payload.ResolverFallbackKind,
		payload.ResolverFallbackIndex,
		payload.ResolverFallbackStrength,
		payload.ResolverFallbackBlockedBy,
		payload.ResolverFallbackAmbiguityClass,
		payload.ResolverFallbackManualRetryHint,
		payload.ResolverFallbackSpecificityFields,
	)
	guidanceSummary := browserResolverGuidanceSummaryFromFields(
		payload.ResolverBlockedBy,
		payload.ResolverAmbiguityClass,
		payload.ResolverCandidateKind,
		payload.ResolverCandidateStrength,
		payload.ResolverRetryDisposition,
		payload.ResolverManualRetryHint,
		payload.ResolverNextStepAlias,
	)
	payload.ResolverFallbackExplanation = browserResolverFallbackExplanationSummaryForSummary(fallbackSummary)
	payload.ResolverExplanation = browserResolverExplanationSummaryForSummaries(fallbackSummary, guidanceSummary)
	payload.DiagnosticsExplanation = browserDiagnosticsExplanationSummaryForSummaries(fallbackSummary, guidanceSummary)
	payload.Summary = browserTopLevelSummaryForSummaries(fallbackSummary, guidanceSummary)
	payload.Explanation = browserCloneTopLevelSummary(payload.Summary)
	payload.Diagnostics = browserCloneTopLevelSummary(payload.Summary)
	payload.Display = browserTopLevelDisplayFromSummary(payload.Summary)
	browserApplyActionabilityFailureSummaryAliases(payload.Kind, payload.Status, payload.Actionability, payload.FailureEvidence, &payload.DiagnosticsExplanation, &payload.Explanation, &payload.Diagnostics, &payload.Summary, &payload.Display)
	browserApplyActionSuccessSummaryAliases(payload.Kind, payload.Status, &payload.DiagnosticsExplanation, &payload.Explanation, &payload.Diagnostics, &payload.Summary, &payload.Display)
	reviewDecision, reviewReady := browserActTopLevelReviewState(payload)
	reviewStatus := strings.TrimSpace(payload.Status)
	if strings.TrimSpace(reviewDecision) != "" && !reviewReady {
		reviewStatus = "review_required"
	}
	browserApplyReviewSummaryAliases(reviewStatus, reviewDecision, reviewReady, &payload.DiagnosticsExplanation, &payload.Explanation, &payload.Diagnostics, &payload.Summary, &payload.Display, &payload.Review)
	payload.Surface, payload.View = browserRuntimeTopLevelSurfaceAndViewFromDisplayAndReviewWithMetadata(payload.Display, payload.Review, browserTopLevelCapabilitySurfaceFromMetadata(payload.CapabilityMetadata), payload.BrowserSurface, payload.BrowserOptInTargets)
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func marshalBrowserPageActionReviewBlockedPayload(payload browserPageActionReviewBlockedPayload) (string, error) {
	browserApplyRegistrationRouteCapabilityMetadata(&payload.browserRegistrationRouteCapabilityPayload, &payload.BrowserSurface, &payload.BrowserOptInTargets)
	browserApplyReviewSummaryAliases(payload.Status, payload.ReviewDecision, payload.ReviewReady, &payload.DiagnosticsExplanation, &payload.Explanation, &payload.Diagnostics, &payload.Summary, &payload.Display, &payload.Review)
	payload.Surface, payload.View = browserRuntimeTopLevelSurfaceAndViewFromDisplayAndReviewWithMetadata(payload.Display, payload.Review, browserTopLevelCapabilitySurfaceFromMetadata(payload.CapabilityMetadata), payload.BrowserSurface, payload.BrowserOptInTargets)
	payload.DefaultCandidateRoute = browserApplyDefaultCandidateRouteToRegistrationPayloadShells(&payload.Explanation, &payload.Diagnostics, &payload.Summary, &payload.Display, &payload.Review, &payload.Surface, &payload.View)
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func marshalBrowserOpenPayload(payload browserOpenToolPayload) (string, error) {
	browserApplyRegistrationRouteCapabilityMetadata(&payload.browserRegistrationRouteCapabilityPayload, &payload.BrowserSurface, &payload.BrowserOptInTargets)
	browserApplyActionSuccessSummaryAliases("open", payload.Status, &payload.DiagnosticsExplanation, &payload.Explanation, &payload.Diagnostics, &payload.Summary, &payload.Display)
	payload.Surface, payload.View = browserRuntimeTopLevelSurfaceAndViewFromDisplayAndReviewWithMetadata(payload.Display, payload.Review, browserTopLevelCapabilitySurfaceFromMetadata(payload.CapabilityMetadata), payload.BrowserSurface, payload.BrowserOptInTargets)
	payload.DefaultCandidateRoute = browserApplyDefaultCandidateRouteToRegistrationPayloadShells(&payload.Explanation, &payload.Diagnostics, &payload.Summary, &payload.Display, &payload.Review, &payload.Surface, &payload.View)
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func marshalBrowserNavigatePayload(payload browserNavigateToolPayload) (string, error) {
	browserApplyRegistrationRouteCapabilityMetadata(&payload.browserRegistrationRouteCapabilityPayload, &payload.BrowserSurface, &payload.BrowserOptInTargets)
	browserApplyActionSuccessSummaryAliases("navigate", payload.Status, &payload.DiagnosticsExplanation, &payload.Explanation, &payload.Diagnostics, &payload.Summary, &payload.Display)
	browserApplyReviewSummaryAliases(payload.Status, payload.ReviewDecision, payload.ReviewReady, &payload.DiagnosticsExplanation, &payload.Explanation, &payload.Diagnostics, &payload.Summary, &payload.Display, &payload.Review)
	payload.Surface, payload.View = browserRuntimeTopLevelSurfaceAndViewFromDisplayAndReviewWithMetadata(payload.Display, payload.Review, browserTopLevelCapabilitySurfaceFromMetadata(payload.CapabilityMetadata), payload.BrowserSurface, payload.BrowserOptInTargets)
	payload.DefaultCandidateRoute = browserApplyDefaultCandidateRouteToRegistrationPayloadShells(&payload.Explanation, &payload.Diagnostics, &payload.Summary, &payload.Display, &payload.Review, &payload.Surface, &payload.View)
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func browserOpenCurrentTargetValue(targetID string) string {
	if strings.TrimSpace(targetID) == "" {
		return ""
	}
	return "current"
}

func browserNavigateStatus(result BrowserNavigateResult, navigationResult agentxbrowserruntime.SharedSessionBrowserNavigationResultEventResult) string {
	if navigationResult.ReviewRequired {
		return "review_required"
	}
	return firstNonEmpty(strings.TrimSpace(result.Status), "navigated")
}

func marshalBrowserExtractPayload(payload browserExtractToolPayload) (string, error) {
	browserApplyRegistrationRouteCapabilityMetadata(&payload.browserRegistrationRouteCapabilityPayload, &payload.BrowserSurface, &payload.BrowserOptInTargets)
	browserApplyActionSuccessSummaryAliases("extract", payload.Status, &payload.DiagnosticsExplanation, &payload.Explanation, &payload.Diagnostics, &payload.Summary, &payload.Display)
	browserApplyReviewSummaryAliases(payload.Status, payload.ReviewDecision, payload.ReviewReady, &payload.DiagnosticsExplanation, &payload.Explanation, &payload.Diagnostics, &payload.Summary, &payload.Display, &payload.Review)
	payload.Surface, payload.View = browserRuntimeTopLevelSurfaceAndViewFromDisplayAndReviewWithMetadata(payload.Display, payload.Review, browserTopLevelCapabilitySurfaceFromMetadata(payload.CapabilityMetadata), payload.BrowserSurface, payload.BrowserOptInTargets)
	payload.DefaultCandidateRoute = browserApplyDefaultCandidateRouteToRegistrationPayloadShells(&payload.Explanation, &payload.Diagnostics, &payload.Summary, &payload.Display, &payload.Review, &payload.Surface, &payload.View)
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func marshalBrowserScreenshotPayload(payload browserScreenshotToolPayload) (string, error) {
	browserApplyRegistrationRouteCapabilityMetadata(&payload.browserRegistrationRouteCapabilityPayload, &payload.BrowserSurface, &payload.BrowserOptInTargets)
	browserApplyActionabilityEvidence(
		"screenshot",
		payload.Status,
		payload.Note,
		payload.RecoveryAction,
		payload.Ref,
		payload.Selector,
		"",
		"",
		"",
		"",
		nil,
		false,
		payload.Path,
		payload.FinalURL,
		payload.Title,
		nil,
		nil,
		payload.ResolverOutcome,
		&payload.Actionability,
		&payload.FailureEvidence,
	)
	browserApplyResolverFallbackSummary(
		payload.ResolverOutcome,
		&payload.ResolvedViaFallback,
		&payload.ResolverFallbackKind,
		&payload.ResolverFallbackIndex,
		&payload.ResolverFallbackStrength,
		&payload.ResolverFallbackBlockedBy,
		&payload.ResolverFallbackAmbiguityClass,
		&payload.ResolverFallbackManualRetryHint,
		&payload.ResolverFallbackSpecificityFields,
	)
	fallbackSummary := browserResolverFallbackSummaryFromFields(
		payload.ResolvedViaFallback,
		payload.ResolverFallbackKind,
		payload.ResolverFallbackIndex,
		payload.ResolverFallbackStrength,
		payload.ResolverFallbackBlockedBy,
		payload.ResolverFallbackAmbiguityClass,
		payload.ResolverFallbackManualRetryHint,
		payload.ResolverFallbackSpecificityFields,
	)
	payload.ResolverFallbackExplanation = browserResolverFallbackExplanationSummaryForSummary(fallbackSummary)
	payload.ResolverExplanation = browserResolverExplanationSummaryForSummaries(fallbackSummary, browserResolverGuidanceSummary{})
	payload.DiagnosticsExplanation = browserDiagnosticsExplanationSummaryForSummaries(fallbackSummary, browserResolverGuidanceSummary{})
	payload.Summary = browserTopLevelSummaryForSummaries(fallbackSummary, browserResolverGuidanceSummary{})
	payload.Explanation = browserCloneTopLevelSummary(payload.Summary)
	payload.Diagnostics = browserCloneTopLevelSummary(payload.Summary)
	payload.Display = browserTopLevelDisplayFromSummary(payload.Summary)
	browserApplyActionSuccessSummaryAliases("screenshot", payload.Status, &payload.DiagnosticsExplanation, &payload.Explanation, &payload.Diagnostics, &payload.Summary, &payload.Display)
	browserApplyReviewSummaryAliases(payload.Status, payload.ReviewDecision, payload.ReviewReady, &payload.DiagnosticsExplanation, &payload.Explanation, &payload.Diagnostics, &payload.Summary, &payload.Display, &payload.Review)
	payload.Surface, payload.View = browserRuntimeTopLevelSurfaceAndViewFromDisplayAndReviewWithMetadata(payload.Display, payload.Review, browserTopLevelCapabilitySurfaceFromMetadata(payload.CapabilityMetadata), payload.BrowserSurface, payload.BrowserOptInTargets)
	payload.DefaultCandidateRoute = browserApplyDefaultCandidateRouteToRegistrationPayloadShells(&payload.Explanation, &payload.Diagnostics, &payload.Summary, &payload.Display, &payload.Review, &payload.Surface, &payload.View)
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func marshalBrowserClickPayload(payload browserClickToolPayload) (string, error) {
	browserApplyRegistrationRouteCapabilityMetadata(&payload.browserRegistrationRouteCapabilityPayload, &payload.BrowserSurface, &payload.BrowserOptInTargets)
	browserApplyActionabilityEvidence(
		"click",
		payload.Status,
		payload.Note,
		payload.RecoveryAction,
		payload.Ref,
		payload.Selector,
		payload.Snapshot,
		payload.SnapshotFormat,
		payload.SnapshotRefs,
		payload.SnapshotFrame,
		payload.Elements,
		payload.Truncated,
		"",
		payload.FinalURL,
		payload.Title,
		nil,
		nil,
		payload.ResolverOutcome,
		&payload.Actionability,
		&payload.FailureEvidence,
	)
	browserApplyResolverFallbackSummary(
		payload.ResolverOutcome,
		&payload.ResolvedViaFallback,
		&payload.ResolverFallbackKind,
		&payload.ResolverFallbackIndex,
		&payload.ResolverFallbackStrength,
		&payload.ResolverFallbackBlockedBy,
		&payload.ResolverFallbackAmbiguityClass,
		&payload.ResolverFallbackManualRetryHint,
		&payload.ResolverFallbackSpecificityFields,
	)
	fallbackSummary := browserResolverFallbackSummaryFromFields(
		payload.ResolvedViaFallback,
		payload.ResolverFallbackKind,
		payload.ResolverFallbackIndex,
		payload.ResolverFallbackStrength,
		payload.ResolverFallbackBlockedBy,
		payload.ResolverFallbackAmbiguityClass,
		payload.ResolverFallbackManualRetryHint,
		payload.ResolverFallbackSpecificityFields,
	)
	payload.ResolverFallbackExplanation = browserResolverFallbackExplanationSummaryForSummary(fallbackSummary)
	payload.ResolverExplanation = browserResolverExplanationSummaryForSummaries(fallbackSummary, browserResolverGuidanceSummary{})
	payload.DiagnosticsExplanation = browserDiagnosticsExplanationSummaryForSummaries(fallbackSummary, browserResolverGuidanceSummary{})
	payload.Summary = browserTopLevelSummaryForSummaries(fallbackSummary, browserResolverGuidanceSummary{})
	payload.Explanation = browserCloneTopLevelSummary(payload.Summary)
	payload.Diagnostics = browserCloneTopLevelSummary(payload.Summary)
	payload.Display = browserTopLevelDisplayFromSummary(payload.Summary)
	browserApplyActionSuccessSummaryAliases("click", payload.Status, &payload.DiagnosticsExplanation, &payload.Explanation, &payload.Diagnostics, &payload.Summary, &payload.Display)
	browserApplyReviewSummaryAliases(payload.Status, payload.ReviewDecision, payload.ReviewReady, &payload.DiagnosticsExplanation, &payload.Explanation, &payload.Diagnostics, &payload.Summary, &payload.Display, &payload.Review)
	payload.Surface, payload.View = browserRuntimeTopLevelSurfaceAndViewFromDisplayAndReviewWithMetadata(payload.Display, payload.Review, browserTopLevelCapabilitySurfaceFromMetadata(payload.CapabilityMetadata), payload.BrowserSurface, payload.BrowserOptInTargets)
	payload.DefaultCandidateRoute = browserApplyDefaultCandidateRouteToRegistrationPayloadShells(&payload.Explanation, &payload.Diagnostics, &payload.Summary, &payload.Display, &payload.Review, &payload.Surface, &payload.View)
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func marshalBrowserTypePayload(payload browserTypeToolPayload) (string, error) {
	browserApplyRegistrationRouteCapabilityMetadata(&payload.browserRegistrationRouteCapabilityPayload, &payload.BrowserSurface, &payload.BrowserOptInTargets)
	browserApplyActionabilityEvidence(
		"type",
		payload.Status,
		payload.Note,
		payload.RecoveryAction,
		payload.Ref,
		payload.Selector,
		payload.Snapshot,
		payload.SnapshotFormat,
		payload.SnapshotRefs,
		payload.SnapshotFrame,
		payload.Elements,
		payload.Truncated,
		"",
		payload.FinalURL,
		payload.Title,
		nil,
		nil,
		payload.ResolverOutcome,
		&payload.Actionability,
		&payload.FailureEvidence,
	)
	browserApplyResolverFallbackSummary(
		payload.ResolverOutcome,
		&payload.ResolvedViaFallback,
		&payload.ResolverFallbackKind,
		&payload.ResolverFallbackIndex,
		&payload.ResolverFallbackStrength,
		&payload.ResolverFallbackBlockedBy,
		&payload.ResolverFallbackAmbiguityClass,
		&payload.ResolverFallbackManualRetryHint,
		&payload.ResolverFallbackSpecificityFields,
	)
	fallbackSummary := browserResolverFallbackSummaryFromFields(
		payload.ResolvedViaFallback,
		payload.ResolverFallbackKind,
		payload.ResolverFallbackIndex,
		payload.ResolverFallbackStrength,
		payload.ResolverFallbackBlockedBy,
		payload.ResolverFallbackAmbiguityClass,
		payload.ResolverFallbackManualRetryHint,
		payload.ResolverFallbackSpecificityFields,
	)
	payload.ResolverFallbackExplanation = browserResolverFallbackExplanationSummaryForSummary(fallbackSummary)
	payload.ResolverExplanation = browserResolverExplanationSummaryForSummaries(fallbackSummary, browserResolverGuidanceSummary{})
	payload.DiagnosticsExplanation = browserDiagnosticsExplanationSummaryForSummaries(fallbackSummary, browserResolverGuidanceSummary{})
	payload.Summary = browserTopLevelSummaryForSummaries(fallbackSummary, browserResolverGuidanceSummary{})
	payload.Explanation = browserCloneTopLevelSummary(payload.Summary)
	payload.Diagnostics = browserCloneTopLevelSummary(payload.Summary)
	payload.Display = browserTopLevelDisplayFromSummary(payload.Summary)
	browserApplyActionSuccessSummaryAliases("type", payload.Status, &payload.DiagnosticsExplanation, &payload.Explanation, &payload.Diagnostics, &payload.Summary, &payload.Display)
	browserApplyReviewSummaryAliases(payload.Status, payload.ReviewDecision, payload.ReviewReady, &payload.DiagnosticsExplanation, &payload.Explanation, &payload.Diagnostics, &payload.Summary, &payload.Display, &payload.Review)
	payload.Surface, payload.View = browserRuntimeTopLevelSurfaceAndViewFromDisplayAndReviewWithMetadata(payload.Display, payload.Review, browserTopLevelCapabilitySurfaceFromMetadata(payload.CapabilityMetadata), payload.BrowserSurface, payload.BrowserOptInTargets)
	payload.DefaultCandidateRoute = browserApplyDefaultCandidateRouteToRegistrationPayloadShells(&payload.Explanation, &payload.Diagnostics, &payload.Summary, &payload.Display, &payload.Review, &payload.Surface, &payload.View)
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func marshalBrowserEvalPayload(payload browserEvalToolPayload) (string, error) {
	browserApplyRegistrationRouteCapabilityMetadata(&payload.browserRegistrationRouteCapabilityPayload, &payload.BrowserSurface, &payload.BrowserOptInTargets)
	browserApplyActionSuccessSummaryAliases("evaluate", payload.Status, &payload.DiagnosticsExplanation, &payload.Explanation, &payload.Diagnostics, &payload.Summary, &payload.Display)
	browserApplyReviewSummaryAliases(payload.Status, payload.ReviewDecision, payload.ReviewReady, &payload.DiagnosticsExplanation, &payload.Explanation, &payload.Diagnostics, &payload.Summary, &payload.Display, &payload.Review)
	payload.Surface, payload.View = browserRuntimeTopLevelSurfaceAndViewFromDisplayAndReviewWithMetadata(payload.Display, payload.Review, browserTopLevelCapabilitySurfaceFromMetadata(payload.CapabilityMetadata), payload.BrowserSurface, payload.BrowserOptInTargets)
	payload.DefaultCandidateRoute = browserApplyDefaultCandidateRouteToRegistrationPayloadShells(&payload.Explanation, &payload.Diagnostics, &payload.Summary, &payload.Display, &payload.Review, &payload.Surface, &payload.View)
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
