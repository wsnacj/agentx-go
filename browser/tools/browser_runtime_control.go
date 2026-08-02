package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
	types "github.com/wsnacj/agentx-go/components/llm"
)

type browserRuntimeRouteDescriptor struct {
	Backend       string `json:"backend,omitempty"`
	Profile       string `json:"profile,omitempty"`
	RuntimeTarget string `json:"runtime_target,omitempty"`
	Source        string `json:"source,omitempty"`
	Endpoint      string `json:"endpoint,omitempty"`
}

type browserRuntimeRouteStatus struct {
	Profile               string                        `json:"profile,omitempty"`
	RuntimeTarget         string                        `json:"runtime_target,omitempty"`
	Backend               string                        `json:"backend,omitempty"`
	Source                string                        `json:"source,omitempty"`
	Endpoint              string                        `json:"endpoint,omitempty"`
	DefaultCandidateRoute browserRuntimeRouteDescriptor `json:"default_candidate_route,omitempty"`
	Status                string                        `json:"status"`
	RuntimeActions        []string                      `json:"runtime_actions,omitempty"`
	BrowserTools          []string                      `json:"browser_tools,omitempty"`
	ArtifactTools         []string                      `json:"artifact_tools,omitempty"`
	ArtifactKinds         []string                      `json:"artifact_kinds,omitempty"`
	ArtifactContract      string                        `json:"artifact_contract,omitempty"`
	BrowserActKinds       []string                      `json:"browser_act_kinds,omitempty"`
	BrowserSurface        string                        `json:"browser_surface,omitempty"`
	BrowserOptInTargets   []string                      `json:"browser_opt_in_targets,omitempty"`
	Capabilities          map[string]bool               `json:"capabilities,omitempty"`
	Note                  string                        `json:"note,omitempty"`
}

type browserRuntimeSubstrateStatus struct {
	Role                  string                        `json:"role,omitempty"`
	SelectionState        string                        `json:"selection_state,omitempty"`
	SelectionReason       string                        `json:"selection_reason,omitempty"`
	Profile               string                        `json:"profile,omitempty"`
	RuntimeTarget         string                        `json:"runtime_target,omitempty"`
	Backend               string                        `json:"backend,omitempty"`
	Source                string                        `json:"source,omitempty"`
	Endpoint              string                        `json:"endpoint,omitempty"`
	DefaultCandidateRoute browserRuntimeRouteDescriptor `json:"default_candidate_route,omitempty"`
	Status                string                        `json:"status"`
	SubstratePosture      string                        `json:"substrate_posture,omitempty"`
	SubstrateStatus       string                        `json:"substrate_status,omitempty"`
	SubstrateReason       string                        `json:"substrate_reason,omitempty"`
	RuntimeActions        []string                      `json:"runtime_actions,omitempty"`
	BrowserTools          []string                      `json:"browser_tools,omitempty"`
	ArtifactTools         []string                      `json:"artifact_tools,omitempty"`
	ArtifactKinds         []string                      `json:"artifact_kinds,omitempty"`
	ArtifactContract      string                        `json:"artifact_contract,omitempty"`
	BrowserActKinds       []string                      `json:"browser_act_kinds,omitempty"`
	BrowserSurface        string                        `json:"browser_surface,omitempty"`
	BrowserOptInTargets   []string                      `json:"browser_opt_in_targets,omitempty"`
	Capabilities          map[string]bool               `json:"capabilities,omitempty"`
	Note                  string                        `json:"note,omitempty"`
}

type browserRuntimeProfileState struct {
	Backend       string    `json:"backend,omitempty"`
	Profile       string    `json:"profile,omitempty"`
	RuntimeTarget string    `json:"runtime_target,omitempty"`
	BrowserApp    string    `json:"browser_app,omitempty"`
	Status        string    `json:"status,omitempty"`
	Running       bool      `json:"running,omitempty"`
	Connected     bool      `json:"connected,omitempty"`
	Selected      bool      `json:"selected,omitempty"`
	Note          string    `json:"note,omitempty"`
	ObservedAt    time.Time `json:"-"`
	StatusSince   time.Time `json:"-"`
}

const browserRuntimeReconnectWatchdogWindow = 54 * time.Second

type browserRuntimeSessionProfileSelection struct {
	Backend       string `json:"backend,omitempty"`
	Profile       string `json:"profile,omitempty"`
	RuntimeTarget string `json:"runtime_target,omitempty"`
	BrowserApp    string `json:"browser_app,omitempty"`
	Source        string `json:"source,omitempty"`
}

type browserRuntimeSessionTargetSelection struct {
	ID            string `json:"id,omitempty"`
	TabIndex      int    `json:"tab_index,omitempty"`
	URL           string `json:"url,omitempty"`
	Title         string `json:"title,omitempty"`
	Backend       string `json:"backend,omitempty"`
	Profile       string `json:"profile,omitempty"`
	RuntimeTarget string `json:"runtime_target,omitempty"`
	BrowserApp    string `json:"browser_app,omitempty"`
	Source        string `json:"source,omitempty"`
}

type browserRuntimeSessionTarget struct {
	ID            string `json:"id,omitempty"`
	TabIndex      int    `json:"tab_index,omitempty"`
	URL           string `json:"url,omitempty"`
	Title         string `json:"title,omitempty"`
	BrowserApp    string `json:"browser_app,omitempty"`
	Backend       string `json:"backend,omitempty"`
	Profile       string `json:"profile,omitempty"`
	RuntimeTarget string `json:"runtime_target,omitempty"`
	Current       bool   `json:"current,omitempty"`
}

type browserRuntimeSessionTargetReview struct {
	ID            string `json:"id,omitempty"`
	TabIndex      int    `json:"tab_index,omitempty"`
	URL           string `json:"url,omitempty"`
	Title         string `json:"title,omitempty"`
	BrowserApp    string `json:"browser_app,omitempty"`
	Backend       string `json:"backend,omitempty"`
	Profile       string `json:"profile,omitempty"`
	RuntimeTarget string `json:"runtime_target,omitempty"`
	Decision      string `json:"decision,omitempty"`
	Reason        string `json:"reason,omitempty"`
}

type browserRuntimeSessionRoute struct {
	Backend                  string                             `json:"backend,omitempty"`
	Profile                  string                             `json:"profile,omitempty"`
	RuntimeTarget            string                             `json:"runtime_target,omitempty"`
	BrowserApp               string                             `json:"browser_app,omitempty"`
	CurrentTargetID          string                             `json:"current_target_id,omitempty"`
	CurrentTargetSource      string                             `json:"current_target_source,omitempty"`
	PendingTargetReview      *browserRuntimeSessionTargetReview `json:"pending_target_review,omitempty"`
	PendingTargetReviewCount int                                `json:"pending_target_review_count,omitempty"`
	FollowPolicyState        string                             `json:"follow_policy_state,omitempty"`
	FollowPolicyReason       string                             `json:"follow_policy_reason,omitempty"`
	PopupPolicyState         string                             `json:"popup_policy_state,omitempty"`
	PopupPolicyReason        string                             `json:"popup_policy_reason,omitempty"`
	Targets                  []browserRuntimeSessionTarget      `json:"targets,omitempty"`
}

type browserRuntimeSessionRun struct {
	RunID    string `json:"run_id,omitempty"`
	NodeID   string `json:"node_id,omitempty"`
	Status   string `json:"status,omitempty"`
	Provider string `json:"provider,omitempty"`
	Action   string `json:"action,omitempty"`
}

type browserRuntimeSessionHandoffSummary = agentxbrowserruntime.SharedSessionBrowserSessionHandoffSummary

type browserRuntimeSessionBinding struct {
	SessionKey                               string                                                     `json:"session_key,omitempty"`
	CurrentTargetID                          string                                                     `json:"current_target_id,omitempty"`
	PendingTargetReviewCount                 int                                                        `json:"pending_target_review_count,omitempty"`
	BlockedAutoFollowRouteCount              int                                                        `json:"blocked_auto_follow_route_count,omitempty"`
	PopupStormRouteCount                     int                                                        `json:"popup_storm_route_count,omitempty"`
	SelectedBrowserBackend                   string                                                     `json:"selected_browser_backend,omitempty"`
	SelectedBrowserApp                       string                                                     `json:"selected_browser_app,omitempty"`
	SelectedBrowserTargetID                  string                                                     `json:"selected_browser_target_id,omitempty"`
	SelectedBrowserTabIndex                  int                                                        `json:"selected_browser_tab_index,omitempty"`
	RouteTargetCount                         int                                                        `json:"route_target_count,omitempty"`
	SelectedBrowserProfile                   string                                                     `json:"selected_browser_profile,omitempty"`
	SelectedBrowserProfileSource             string                                                     `json:"selected_browser_profile_source,omitempty"`
	SelectedBrowserTarget                    string                                                     `json:"selected_browser_target,omitempty"`
	SelectedBrowserTargetSource              string                                                     `json:"selected_browser_target_source,omitempty"`
	NodeRunCount                             int                                                        `json:"node_run_count,omitempty"`
	ActiveNodeRunID                          string                                                     `json:"active_node_run_id,omitempty"`
	NodeRunStatusCounts                      map[string]int                                             `json:"node_run_status_counts,omitempty"`
	BrowserProfileCount                      int                                                        `json:"browser_profile_count,omitempty"`
	ActiveBrowserProfile                     string                                                     `json:"active_browser_profile,omitempty"`
	BrowserProfileStatusCounts               map[string]int                                             `json:"browser_profile_status_counts,omitempty"`
	SessionHealthState                       string                                                     `json:"session_health_state,omitempty"`
	SessionHealthReason                      string                                                     `json:"session_health_reason,omitempty"`
	SessionHealthRecoveryAction              string                                                     `json:"session_health_recovery_action,omitempty"`
	SessionHealthReconnectHint               string                                                     `json:"session_health_reconnect_hint,omitempty"`
	SessionHealthDisconnectCount             int                                                        `json:"session_health_disconnect_count,omitempty"`
	SessionHealthDisconnectBurstCount        int                                                        `json:"session_health_disconnect_burst_count,omitempty"`
	SessionHealthDisconnectBurstWindowMs     int                                                        `json:"session_health_disconnect_burst_window_ms,omitempty"`
	SessionHealthCooldownRemainingMs         int                                                        `json:"session_health_cooldown_remaining_ms,omitempty"`
	SessionHealthRetryBackoffRemainingMs     int                                                        `json:"session_health_retry_backoff_remaining_ms,omitempty"`
	SessionHealthRestartAttemptCount         int                                                        `json:"session_health_restart_attempt_count,omitempty"`
	SessionHealthRestartFailureCount         int                                                        `json:"session_health_restart_failure_count,omitempty"`
	SessionHealthLastDisconnectUnixMilli     int64                                                      `json:"session_health_last_disconnect_unix_milli,omitempty"`
	SessionHealthLastReconnectUnixMilli      int64                                                      `json:"session_health_last_reconnect_unix_milli,omitempty"`
	SessionHealthLastRestartAttemptUnixMilli int64                                                      `json:"session_health_last_restart_attempt_unix_milli,omitempty"`
	SessionHealthLastRestartResult           string                                                     `json:"session_health_last_restart_result,omitempty"`
	SessionHealthLastRestartError            string                                                     `json:"session_health_last_restart_error,omitempty"`
	SessionHealthRecommendedBackoffMs        int                                                        `json:"session_health_recommended_backoff_ms,omitempty"`
	SessionHealthResolverBlockedBy           string                                                     `json:"session_health_resolver_blocked_by,omitempty"`
	SessionHealthResolverAmbiguityClass      string                                                     `json:"session_health_resolver_ambiguity_class,omitempty"`
	SessionHealthResolverCandidateKind       string                                                     `json:"session_health_resolver_candidate_kind,omitempty"`
	SessionHealthResolverStrength            string                                                     `json:"session_health_resolver_candidate_strength,omitempty"`
	SessionHealthResolverRetryDisposition    string                                                     `json:"session_health_resolver_retry_disposition,omitempty"`
	SessionHealthResolverManualRetryHint     string                                                     `json:"session_health_resolver_manual_retry_hint,omitempty"`
	SessionHealthResolverNextStepAlias       string                                                     `json:"session_health_resolver_next_step_alias,omitempty"`
	SessionHealthResolverSpecificityFields   []string                                                   `json:"session_health_resolver_specificity_fields,omitempty"`
	PropagatedToProxy                        bool                                                       `json:"propagated_to_proxy,omitempty"`
	NodeRuns                                 []browserRuntimeSessionRun                                 `json:"node_runs,omitempty"`
	BrowserProfiles                          []browserRuntimeProfileState                               `json:"browser_profiles,omitempty"`
	SessionHandoff                           *browserRuntimeSessionHandoffSummary                       `json:"session_handoff,omitempty"`
	Coordination                             *browserRuntimeCoordination                                `json:"coordination,omitempty"`
	ReferenceTime                            time.Time                                                  `json:"-"`
	SharedEvaluation                         agentxbrowserruntime.SharedSessionBrowserBindingEvaluation `json:"-"`
	HasSharedEvaluation                      bool                                                       `json:"-"`
}

type browserRuntimeCoordination struct {
	State                     string   `json:"state,omitempty"`
	BrowserOnNode             bool     `json:"browser_on_node,omitempty"`
	HasActiveNodeRun          bool     `json:"has_active_node_run,omitempty"`
	HasRunningBrowserProfile  bool     `json:"has_running_browser_profile,omitempty"`
	SyncBrowserAction         string   `json:"sync_browser_action,omitempty"`
	PrepareBrowserAction      string   `json:"prepare_browser_action,omitempty"`
	RestartBrowserAction      string   `json:"restart_browser_action,omitempty"`
	TeardownBrowserAction     string   `json:"teardown_browser_action,omitempty"`
	PrimaryBrowserAction      string   `json:"primary_browser_action,omitempty"`
	PrimaryNodeAction         string   `json:"primary_node_action,omitempty"`
	NextStep                  string   `json:"next_step,omitempty"`
	RecommendedBrowserActions []string `json:"recommended_browser_actions,omitempty"`
	RecommendedNodeActions    []string `json:"recommended_node_actions,omitempty"`
}

type browserRuntimeSessionHealthSummary struct {
	State                       string
	Reason                      string
	RecoveryAction              string
	ReconnectHint               string
	DisconnectCount             int
	DisconnectBurstCount        int
	DisconnectBurstWindowMs     int
	CooldownRemainingMs         int
	RetryBackoffRemainingMs     int
	RestartAttemptCount         int
	RestartFailureCount         int
	LastDisconnectUnixMilli     int64
	LastReconnectUnixMilli      int64
	LastRestartAttemptUnixMilli int64
	LastRestartResult           string
	LastRestartError            string
	RecommendedBackoffMs        int
	ResolverBlockedBy           string
	AmbiguityClass              string
	CandidateKind               string
	CandidateStrength           string
	RetryDisposition            string
	ManualRetryHint             string
	NextStepAlias               string
	SpecificityFields           []string
}

type browserRuntimeResolverExplanationSummary struct {
	State           string `json:"state,omitempty"`
	SummaryCode     string `json:"summary_code,omitempty"`
	NextStepAlias   string `json:"next_step_alias,omitempty"`
	ManualRetryHint string `json:"manual_retry_hint,omitempty"`
}

type browserRuntimeDiagnosticsExplanationSummary struct {
	Category        string `json:"category,omitempty"`
	State           string `json:"state,omitempty"`
	SummaryCode     string `json:"summary_code,omitempty"`
	NextStepAlias   string `json:"next_step_alias,omitempty"`
	ManualRetryHint string `json:"manual_retry_hint,omitempty"`
}

type browserRuntimeWorkbenchDiagnosticsSummary struct {
	Category             string `json:"category,omitempty"`
	State                string `json:"state,omitempty"`
	SummaryCode          string `json:"summary_code,omitempty"`
	RepairCommand        string `json:"repair_command,omitempty"`
	NextStepAlias        string `json:"next_step_alias,omitempty"`
	ManualRetryHint      string `json:"manual_retry_hint,omitempty"`
	ResolvedViaFallback  bool   `json:"resolved_via_fallback,omitempty"`
	PrimaryBrowserAction string `json:"primary_browser_action,omitempty"`
	PrimaryNodeAction    string `json:"primary_node_action,omitempty"`
	NextStep             string `json:"next_step,omitempty"`
}

type browserRuntimeWorkbenchSurfaceSummary struct {
	Ready                     bool                          `json:"ready,omitempty"`
	Sections                  []string                      `json:"sections,omitempty"`
	RepairCommand             string                        `json:"repair_command,omitempty"`
	DefaultCandidateRoute     browserRuntimeRouteDescriptor `json:"default_candidate_route,omitempty"`
	Review                    *browserReviewSurfaceSummary  `json:"review,omitempty"`
	BrowserTools              []string                      `json:"browser_tools,omitempty"`
	ArtifactTools             []string                      `json:"artifact_tools,omitempty"`
	ArtifactKinds             []string                      `json:"artifact_kinds,omitempty"`
	ArtifactContract          string                        `json:"artifact_contract,omitempty"`
	BrowserActKinds           []string                      `json:"browser_act_kinds,omitempty"`
	BrowserSurface            string                        `json:"browser_surface,omitempty"`
	BrowserOptInTargets       []string                      `json:"browser_opt_in_targets,omitempty"`
	Explanation               *browserTopLevelSummary       `json:"explanation,omitempty"`
	Diagnostics               *browserTopLevelSummary       `json:"diagnostics,omitempty"`
	Summary                   *browserTopLevelSummary       `json:"summary,omitempty"`
	PrimaryBrowserAction      string                        `json:"primary_browser_action,omitempty"`
	PrimaryNodeAction         string                        `json:"primary_node_action,omitempty"`
	NextStep                  string                        `json:"next_step,omitempty"`
	RecommendedBrowserActions []string                      `json:"recommended_browser_actions,omitempty"`
	RecommendedNodeActions    []string                      `json:"recommended_node_actions,omitempty"`
}

type browserRuntimeWorkbenchDisplaySummary struct {
	Ready                 bool                          `json:"ready,omitempty"`
	Sections              []string                      `json:"sections,omitempty"`
	Category              string                        `json:"category,omitempty"`
	State                 string                        `json:"state,omitempty"`
	SummaryCode           string                        `json:"summary_code,omitempty"`
	RepairCommand         string                        `json:"repair_command,omitempty"`
	DefaultCandidateRoute browserRuntimeRouteDescriptor `json:"default_candidate_route,omitempty"`
	NextStepAlias         string                        `json:"next_step_alias,omitempty"`
	ManualRetryHint       string                        `json:"manual_retry_hint,omitempty"`
	ResolvedViaFallback   bool                          `json:"resolved_via_fallback,omitempty"`
	PrimaryBrowserAction  string                        `json:"primary_browser_action,omitempty"`
	PrimaryNodeAction     string                        `json:"primary_node_action,omitempty"`
	NextStep              string                        `json:"next_step,omitempty"`
}

type browserRuntimeRouteResolution struct {
	ProfileSource       string `json:"profile_source,omitempty"`
	RuntimeTargetSource string `json:"runtime_target_source,omitempty"`
	TargetSource        string `json:"target_source,omitempty"`
}

type browserRuntimePayload struct {
	Action                             string                                             `json:"action,omitempty"`
	Status                             string                                             `json:"status"`
	Force                              bool                                               `json:"force,omitempty"`
	RequestedProfile                   string                                             `json:"requested_profile,omitempty"`
	RequestedRuntimeTarget             string                                             `json:"requested_runtime_target,omitempty"`
	CoordinationGoal                   string                                             `json:"coordination_goal,omitempty"`
	RememberProfile                    bool                                               `json:"remember_profile,omitempty"`
	RequestedBrowserApp                string                                             `json:"requested_browser_app,omitempty"`
	RequestedColor                     string                                             `json:"requested_color,omitempty"`
	RequestedCopyFrom                  string                                             `json:"requested_copy_from,omitempty"`
	DefaultRoute                       browserRuntimeRouteDescriptor                      `json:"default_route"`
	DefaultCandidateRoute              browserRuntimeRouteDescriptor                      `json:"default_candidate_route,omitempty"`
	SelectedRoute                      *browserRuntimeRouteDescriptor                     `json:"selected_route,omitempty"`
	SessionProfileSelection            *browserRuntimeSessionProfileSelection             `json:"session_profile_selection,omitempty"`
	SessionTargetSelection             *browserRuntimeSessionTargetSelection              `json:"session_target_selection,omitempty"`
	RouteResolution                    *browserRuntimeRouteResolution                     `json:"route_resolution,omitempty"`
	DefaultProfile                     string                                             `json:"default_profile,omitempty"`
	ConfiguredProfiles                 []string                                           `json:"configured_profiles,omitempty"`
	RuntimeActions                     []string                                           `json:"runtime_actions,omitempty"`
	WorkbenchReady                     bool                                               `json:"workbench_ready,omitempty"`
	WorkbenchSections                  []string                                           `json:"workbench_sections,omitempty"`
	WorkbenchPrimaryBrowserAction      string                                             `json:"workbench_primary_browser_action,omitempty"`
	WorkbenchPrimaryNodeAction         string                                             `json:"workbench_primary_node_action,omitempty"`
	WorkbenchNextStep                  string                                             `json:"workbench_next_step,omitempty"`
	WorkbenchRecommendedBrowserActions []string                                           `json:"workbench_recommended_browser_actions,omitempty"`
	WorkbenchRecommendedNodeActions    []string                                           `json:"workbench_recommended_node_actions,omitempty"`
	WorkbenchExplanation               *browserRuntimeDiagnosticsExplanationSummary       `json:"workbench_explanation,omitempty"`
	WorkbenchDiagnostics               *browserRuntimeWorkbenchDiagnosticsSummary         `json:"workbench_diagnostics,omitempty"`
	WorkbenchSummary                   *browserTopLevelSummary                            `json:"workbench_summary,omitempty"`
	Workbench                          *browserRuntimeWorkbenchSurfaceSummary             `json:"workbench,omitempty"`
	WorkbenchDisplay                   *browserRuntimeWorkbenchDisplaySummary             `json:"workbench_display,omitempty"`
	Review                             *browserReviewSurfaceSummary                       `json:"review,omitempty"`
	Explanation                        *browserTopLevelSummary                            `json:"explanation,omitempty"`
	Diagnostics                        *browserTopLevelSummary                            `json:"diagnostics,omitempty"`
	Summary                            *browserTopLevelSummary                            `json:"summary,omitempty"`
	Display                            *browserTopLevelDisplaySummary                     `json:"display,omitempty"`
	Surface                            *browserTopLevelSurfaceSummary                     `json:"surface,omitempty"`
	View                               *browserTopLevelViewSummary                        `json:"view,omitempty"`
	BrowserTools                       []string                                           `json:"browser_tools,omitempty"`
	ArtifactTools                      []string                                           `json:"artifact_tools,omitempty"`
	ArtifactKinds                      []string                                           `json:"artifact_kinds,omitempty"`
	ArtifactContract                   string                                             `json:"artifact_contract,omitempty"`
	BrowserActKinds                    []string                                           `json:"browser_act_kinds,omitempty"`
	BrowserSurface                     string                                             `json:"browser_surface,omitempty"`
	BrowserOptInTargets                []string                                           `json:"browser_opt_in_targets,omitempty"`
	Capabilities                       map[string]bool                                    `json:"capabilities,omitempty"`
	ConfiguredTargets                  []string                                           `json:"configured_targets,omitempty"`
	SubstratePosture                   string                                             `json:"substrate_posture,omitempty"`
	SubstrateStatus                    string                                             `json:"substrate_status,omitempty"`
	SubstrateReason                    string                                             `json:"substrate_reason,omitempty"`
	SubstrateSelectionStrategy         string                                             `json:"substrate_selection_strategy,omitempty"`
	SubstrateSelectionReason           string                                             `json:"substrate_selection_reason,omitempty"`
	SubstrateMatrix                    []browserRuntimeSubstrateStatus                    `json:"substrate_matrix,omitempty"`
	LaunchDiagnostics                  *browserRuntimeLaunchDiagnosticsSummary            `json:"launch_diagnostics,omitempty"`
	WorkbenchLaunchDiagnostics         *browserRuntimeLaunchDiagnosticsSummary            `json:"workbench_launch_diagnostics,omitempty"`
	Doctor                             *BrowserDoctorSummary                              `json:"doctor,omitempty"`
	RepairDecision                     string                                             `json:"repair_decision,omitempty"`
	RepairReady                        bool                                               `json:"repair_ready,omitempty"`
	RepairOutput                       string                                             `json:"repair_output,omitempty"`
	ProfileStatus                      *browserRuntimeProfileState                        `json:"profile_status,omitempty"`
	PreparedProfile                    string                                             `json:"prepared_profile,omitempty"`
	PrepareDecision                    string                                             `json:"prepare_decision,omitempty"`
	PrepareReady                       bool                                               `json:"prepare_ready,omitempty"`
	StopDecision                       string                                             `json:"stop_decision,omitempty"`
	StopReady                          bool                                               `json:"stop_ready,omitempty"`
	RestartDecision                    string                                             `json:"restart_decision,omitempty"`
	RestartReady                       bool                                               `json:"restart_ready,omitempty"`
	CreateDecision                     string                                             `json:"create_decision,omitempty"`
	CreateReady                        bool                                               `json:"create_ready,omitempty"`
	DeleteDecision                     string                                             `json:"delete_decision,omitempty"`
	DeleteReady                        bool                                               `json:"delete_ready,omitempty"`
	SelectDecision                     string                                             `json:"select_decision,omitempty"`
	SelectReady                        bool                                               `json:"select_ready,omitempty"`
	ClearDecision                      string                                             `json:"clear_decision,omitempty"`
	ClearReady                         bool                                               `json:"clear_ready,omitempty"`
	ClearSessionDecision               string                                             `json:"clear_session_decision,omitempty"`
	ClearSessionReady                  bool                                               `json:"clear_session_ready,omitempty"`
	SyncSessionDecision                string                                             `json:"sync_session_decision,omitempty"`
	SyncSessionReady                   bool                                               `json:"sync_session_ready,omitempty"`
	SelectTargetDecision               string                                             `json:"select_target_decision,omitempty"`
	SelectTargetReady                  bool                                               `json:"select_target_ready,omitempty"`
	ClearTargetDecision                string                                             `json:"clear_target_decision,omitempty"`
	ClearTargetReady                   bool                                               `json:"clear_target_ready,omitempty"`
	ClearedSessionTargets              int                                                `json:"cleared_session_targets,omitempty"`
	ClearedSessionProfiles             int                                                `json:"cleared_session_profiles,omitempty"`
	RememberDecision                   string                                             `json:"remember_decision,omitempty"`
	RememberReady                      bool                                               `json:"remember_ready,omitempty"`
	CoordinationDecision               string                                             `json:"coordination_decision,omitempty"`
	CoordinationState                  string                                             `json:"coordination_state,omitempty"`
	CoordinationReady                  bool                                               `json:"coordination_ready,omitempty"`
	ResolverBlockedBy                  string                                             `json:"resolver_blocked_by,omitempty"`
	ResolverAmbiguityClass             string                                             `json:"resolver_ambiguity_class,omitempty"`
	ResolverCandidateKind              string                                             `json:"resolver_candidate_kind,omitempty"`
	ResolverCandidateStrength          string                                             `json:"resolver_candidate_strength,omitempty"`
	ResolverRetryDisposition           string                                             `json:"resolver_retry_disposition,omitempty"`
	ResolverManualRetryHint            string                                             `json:"resolver_manual_retry_hint,omitempty"`
	ResolverNextStepAlias              string                                             `json:"resolver_next_step_alias,omitempty"`
	ResolverSpecificityFields          []string                                           `json:"resolver_specificity_fields,omitempty"`
	ResolverExplanation                *browserRuntimeResolverExplanationSummary          `json:"resolver_explanation,omitempty"`
	DiagnosticsExplanation             *browserRuntimeDiagnosticsExplanationSummary       `json:"diagnostics_explanation,omitempty"`
	Actionability                      *agentxbrowserruntime.BrowserActionabilityReport   `json:"actionability,omitempty"`
	FailureEvidence                    *agentxbrowserruntime.BrowserActionFailureEvidence `json:"failure_evidence,omitempty"`
	Profiles                           []browserRuntimeProfileState                       `json:"profiles,omitempty"`
	SessionID                          string                                             `json:"session_id,omitempty"`
	SessionBinding                     *browserRuntimeSessionBinding                      `json:"session_binding,omitempty"`
	SessionTargetCount                 int                                                `json:"session_target_count,omitempty"`
	SessionRoutes                      []browserRuntimeSessionRoute                       `json:"session_routes,omitempty"`
	SessionRuns                        []browserRuntimeSessionRun                         `json:"session_runs,omitempty"`
	SessionProfiles                    []browserRuntimeProfileState                       `json:"session_profiles,omitempty"`
	SessionHandoff                     *browserRuntimeSessionHandoffSummary               `json:"session_handoff,omitempty"`
	Routes                             []browserRuntimeRouteStatus                        `json:"routes,omitempty"`
	Note                               string                                             `json:"note,omitempty"`
	discoveredProfiles                 []string
	finalizedSessionHealthSummary      *agentxbrowserruntime.SharedSessionBrowserHealthSummary
}

type browserRuntimeCapabilityMetadata struct {
	RuntimeActions      []string
	BrowserTools        []string
	ArtifactTools       []string
	ArtifactKinds       []string
	ArtifactContract    string
	BrowserActKinds     []string
	BrowserSurface      string
	BrowserOptInTargets []string
	Capabilities        map[string]bool
}

func browserRuntimeDefinition(actions []string) types.Tool {
	if len(actions) == 0 {
		actions = []string{"status"}
	}
	return types.Tool{
		Type: "function",
		Function: types.Function{
			Name:        "browser_runtime",
			Description: "Specialist companion to the unified `browser` workbench. Use this raw runtime/profile/session control surface when you explicitly want the browser_runtime tool name or need to isolate managed browser runtime inspection and lifecycle control from the broader unified browser entrypoint. Use action=status/doctor for diagnostics, action=prepare/ready for bring-up, and action=clear_session/reset only when explicitly clearing state. If diagnostics mention the legacy host path, pass explicit runtime_target=host for host-only runtime actions.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"action":            map[string]any{"type": "string", "enum": actions, "description": "Runtime operation. Use status/doctor for diagnostics, prepare for readiness, and clear_session only when clearing session state is intended."},
					"profile":           browserRuntimeProfileSchema(),
					"target":            map[string]any{"type": "string"},
					"tab_index":         map[string]any{"type": "integer", "minimum": 1},
					"browser_app":       map[string]any{"type": "string"},
					"color":             map[string]any{"type": "string"},
					"copy_from":         map[string]any{"type": "string"},
					"runtime_target":    browserRuntimeTargetSchema(),
					"coordination_goal": map[string]any{"type": "string", "enum": []string{"ensure", "sync", "restart", "teardown"}},
					"remember_profile":  map[string]any{"type": "boolean"},
					"force":             map[string]any{"type": "boolean"},
					"include_routes":    map[string]any{"type": "boolean"},
				},
			},
		},
	}
}

func registerBrowserRuntimeTool(ctx browserRegistrationContext) {
	availableActions := browserRuntimeAvailableActions(ctx)
	ctx.reg.Register(browserRuntimeDefinition(availableActions), func(callCtx context.Context, call types.FunctionCall) (string, error) {
		params, err := decodeArgs(call.Arguments)
		if err != nil {
			return "", err
		}
		action := browserRuntimeCanonicalAction(firstString(params, "action"))
		if action == "" {
			action = "status"
		}
		if !containsString(availableActions, action) {
			return "", fmt.Errorf("browser_runtime: action must be one of %s", strings.Join(availableActions, ", "))
		}

		requestedProfile := strings.ToLower(strings.TrimSpace(firstString(params, "profile", "browser_profile", "runtime_profile")))
		requestedTarget := browserRequestedRuntimeTarget(params)
		coordinationGoal := browserRuntimeCoordinationGoal(params)

		executionPreview := browserRegistrationExecutionPreviewForContext(ctx)
		diagnosticsPreview := browserRuntimeDiagnosticsPreviewForExecutionPreview(ctx, executionPreview)
		defaultRoute := executionPreview.DefaultRoute
		substrateAssessment := executionPreview.Registration.SubstrateAssessment
		dispatchBase := executionPreview.DispatchBase
		hiddenImplicitHostDefaultBase := executionPreview.HiddenImplicitHostDefaultBase
		sessionExecutionPreview := browserRuntimePreviewSessionSelectionsForExecution(
			callCtx,
			ctx.sessionStateRegistry,
			ctx.sessionRegistry,
			params,
			dispatchBase,
			hiddenImplicitHostDefaultBase,
			executionPreview.EffectiveBackend,
		)
		dispatchBase = sessionExecutionPreview.Base
		hiddenImplicitHostDefaultBase = sessionExecutionPreview.HiddenImplicitHostDefaultBase
		sessionSelectionPreview := sessionExecutionPreview.SessionSelectionPreview
		requestedBrowserApp := strings.TrimSpace(firstString(params, "browser_app", "browser"))
		if requestedBrowserApp == "" {
			requestedBrowserApp = sessionSelectionPreview.RequestedBrowserApp
		}
		payload := browserRuntimePayload{
			Action:                 action,
			Status:                 "ok",
			Force:                  firstBool(params, "force"),
			RequestedProfile:       requestedProfile,
			RequestedRuntimeTarget: requestedTarget,
			CoordinationGoal:       coordinationGoal,
			RememberProfile:        firstBool(params, "remember_profile", "remember"),
			RequestedBrowserApp:    requestedBrowserApp,
			RequestedColor:         strings.TrimSpace(firstString(params, "color")),
			RequestedCopyFrom:      strings.TrimSpace(firstString(params, "copy_from", "copy_from_profile")),
			SessionID:              ToolSessionIDFromContext(callCtx),
		}
		if action == "refresh" {
			payload.CoordinationGoal = "restart"
		}
		if action == "repair" {
			browserRuntimeApplyRepairActionOutcome(
				&payload,
				browserRuntimeDispatchRepairAction(ctx, callCtx),
			)
			executionPreview = browserRegistrationExecutionPreviewForContext(ctx)
			diagnosticsPreview = browserRuntimeDiagnosticsPreviewForExecutionPreview(ctx, executionPreview)
			browserRuntimeRefreshSubstrateContextWithPreview(ctx, &payload, diagnosticsPreview)
			browserRuntimeApplyCapabilityMetadataToPayload(&payload, browserRuntimeDiagnosticsMetadata(ctx))
			browserRuntimeFinalizeActionDispatchPayload(ctx, &payload, browserRuntimeActionDispatchResultPostProcess{
				ConfiguredInfo:                executionPreview.DefaultRoute,
				ResolutionDefaultRoute:        executionPreview.DefaultRoute,
				HiddenImplicitHostDefaultBase: executionPreview.HiddenImplicitHostDefaultBase,
				DiagnosticsPreview:            diagnosticsPreview,
				UseDiagnosticsPreview:         true,
				IncludeRoutes:                 firstBool(params, "include_routes"),
			})
			browserRuntimeApplyDoctorSummary(ctx, &payload, diagnosticsPreview)
			browserRuntimeSyncTopLevelSurfaceSummary(&payload)
			browserRuntimeApplyTopLevelSubstrateSummary(&payload)
			browserRuntimeApplyToolAwareActionCommands(ctx, &payload)
			out, err := json.Marshal(payload)
			if err != nil {
				return "", err
			}
			return string(out), nil
		}

		dispatchSelection, dispatchErr := browserResolveRuntimeActionDispatchSelection(
			callCtx,
			sessionSelectionPreview,
			params,
			action,
			requestedProfile,
			requestedTarget,
			dispatchBase,
			hiddenImplicitHostDefaultBase,
			executionPreview.EffectiveBackend,
		)
		if dispatchErr != nil {
			return "", fmt.Errorf("browser_runtime: %w", dispatchErr)
		}
		dispatchControl := browserRuntimePrepareActionDispatchControlPlaneWithPreview(
			ctx,
			callCtx,
			&payload,
			action,
			defaultRoute,
			substrateAssessment,
			diagnosticsPreview,
			dispatchSelection,
		)
		routeErr := dispatchControl.RouteErr
		configuredInfo := dispatchControl.ConfiguredInfo
		if !dispatchControl.Handled {
			selectedBackend := dispatchControl.SelectedBackend
			selectedInfo := dispatchControl.SelectedInfo
			capabilities := dispatchControl.Capabilities
			effectiveProfile := firstNonEmpty(strings.TrimSpace(payload.RequestedProfile), strings.TrimSpace(selectedInfo.Profile))
			control, _ := selectedBackend.(BrowserRuntimeControlBackend)
			watchManagerProvider := browserRuntimeSharedWatchManagerProvider(ctx)
			watchManager := watchManagerProvider.Bind(control)
			bindingProjection := browserRuntimeTopLevelBindingProjection(
				callCtx,
				ctx.sessionRegistry,
				ctx.sessionRunRegistry,
				ctx.sessionStateRegistry,
				payload.SelectedRoute,
				nil,
				nil,
				nil,
			)
			browserRuntimeApplyTopLevelBindingProjection(callCtx, &payload, bindingProjection)
			var bindingEvaluation *agentxbrowserruntime.SharedSessionBrowserBindingEvaluation
			switch action {
			case "status", "workbench", "profiles", "sessions":
				bindingEvaluation = browserRuntimeDispatchInspectionAction(
					ctx,
					callCtx,
					&payload,
					selectedBackend,
					watchManager,
					browserRuntimeInspectionActionOptions{
						Action:           action,
						SelectedInfo:     selectedInfo,
						EffectiveProfile: effectiveProfile,
						Capabilities:     capabilities,
					},
				)
			case "prepare", "start", "restart", "refresh", "stop", "create_profile", "delete_profile":
				if browserRuntimeDispatchLifecycleAction(
					callCtx,
					&payload,
					browserRuntimeLifecycleDispatchOptions{
						RepairScript:         ctx.opts.RepairScript,
						Action:               action,
						Capabilities:         capabilities,
						SelectedBackend:      selectedBackend,
						EffectiveProfile:     effectiveProfile,
						SelectedInfo:         selectedInfo,
						SelectedRoute:        payload.SelectedRoute,
						WatchManagerProvider: watchManagerProvider,
						SessionRegistry:      ctx.sessionRegistry,
						StateRegistry:        ctx.sessionStateRegistry,
					},
				) {
					break
				}
			case "coordinate":
				if browserRuntimeDispatchCoordinateAction(
					callCtx,
					&payload,
					browserRuntimeCoordinateDispatchOptions{
						RepairScript:         ctx.opts.RepairScript,
						Action:               action,
						CoordinationGoal:     coordinationGoal,
						Capabilities:         capabilities,
						SelectedBackend:      selectedBackend,
						EffectiveProfile:     effectiveProfile,
						SelectedInfo:         selectedInfo,
						SelectedRoute:        payload.SelectedRoute,
						WatchManagerProvider: watchManagerProvider,
						SessionRegistry:      ctx.sessionRegistry,
						StateRegistry:        ctx.sessionStateRegistry,
						RequestedBrowserApp:  payload.RequestedBrowserApp,
						Force:                payload.Force,
					},
				) {
					break
				}
			case "select_profile":
				if browserRuntimeDispatchSessionSelectionAction(
					callCtx,
					&payload,
					browserRuntimeSessionSelectionDispatchOptions{
						Action:               action,
						Capabilities:         capabilities,
						SelectedBackend:      selectedBackend,
						WatchManagerProvider: watchManagerProvider,
						StateRegistry:        ctx.sessionStateRegistry,
						SelectedInfo:         selectedInfo,
						SelectedRoute:        payload.SelectedRoute,
						RequestedBrowserApp:  payload.RequestedBrowserApp,
						Params:               params,
						Force:                payload.Force,
					},
				) {
					break
				}
			case "clear_profile", "clear_session", "clear_target":
				if browserRuntimeDispatchClearSessionAction(
					callCtx,
					&payload,
					browserRuntimeClearDispatchOptions{
						Action:               action,
						Capabilities:         capabilities,
						SessionRegistry:      ctx.sessionRegistry,
						WatchManagerProvider: watchManagerProvider,
						StateRegistry:        ctx.sessionStateRegistry,
						SelectedInfo:         selectedInfo,
						SelectedRoute:        payload.SelectedRoute,
					},
				) {
					break
				}
			case "sync_session":
				if browserRuntimeDispatchSessionSyncAction(
					callCtx,
					&payload,
					browserRuntimeSessionSyncDispatchOptions{
						Action:               action,
						Capabilities:         capabilities,
						SelectedBackend:      selectedBackend,
						WatchManagerProvider: watchManagerProvider,
						SelectedInfo:         selectedInfo,
						SelectedRoute:        payload.SelectedRoute,
						RequestedBrowserApp:  payload.RequestedBrowserApp,
					},
				) {
					break
				}
			case "select_target":
				if browserRuntimeDispatchSessionSelectionAction(
					callCtx,
					&payload,
					browserRuntimeSessionSelectionDispatchOptions{
						Action:               action,
						Capabilities:         capabilities,
						SelectedBackend:      selectedBackend,
						WatchManagerProvider: watchManagerProvider,
						StateRegistry:        ctx.sessionStateRegistry,
						SelectedInfo:         selectedInfo,
						SelectedRoute:        payload.SelectedRoute,
						RequestedBrowserApp:  payload.RequestedBrowserApp,
						Params:               params,
						Force:                payload.Force,
					},
				) {
					break
				}
			}
			browserRuntimeFinalizeActionSessionPayload(ctx, callCtx, &payload, browserRuntimeActionSessionResultPostProcess{
				Action:            action,
				CoordinationGoal:  coordinationGoal,
				BindingEvaluation: bindingEvaluation,
			})
		}

		browserRuntimeFinalizeActionDispatchPayload(ctx, &payload, browserRuntimeActionDispatchResultPostProcess{
			ConfiguredInfo:                configuredInfo,
			ResolutionDefaultRoute:        defaultRoute,
			HiddenImplicitHostDefaultBase: hiddenImplicitHostDefaultBase,
			DiagnosticsPreview:            diagnosticsPreview,
			UseDiagnosticsPreview:         true,
			RouteErr:                      routeErr,
			IncludeRoutes:                 firstBool(params, "include_routes"),
		})
		browserRuntimeMaybeApplyManagedLaunchFailureSurface(&payload, action, requestedTarget, diagnosticsPreview)
		browserRuntimeMaybeApplyManagedLaunchFailureInspectionSummary(ctx, &payload, action, requestedTarget, diagnosticsPreview)
		browserRuntimeApplyDoctorSummary(ctx, &payload, diagnosticsPreview)
		browserRuntimeMaybeApplyDoctorRouteInspectionSummary(ctx, &payload, action, requestedTarget, diagnosticsPreview)
		browserRuntimeApplyTopLevelSubstrateSummary(&payload)
		browserRuntimeApplyToolAwareActionCommands(ctx, &payload)

		out, err := json.Marshal(payload)
		if err != nil {
			return "", err
		}
		return string(out), nil
	})
}

func browserRuntimeHideImplicitLegacyHostFallbackSelections(payload *browserRuntimePayload, hiddenImplicitHostDefaultBase bool) {
	if payload == nil {
		return
	}
	if strings.TrimSpace(payload.RequestedRuntimeTarget) != "" {
		return
	}
	if !hiddenImplicitHostDefaultBase {
		return
	}
	action := browserRuntimeCanonicalAction(payload.Action)
	switch action {
	case "doctor", "status", "profiles", "sessions", "workbench", "prepare", "coordinate", "start":
	default:
		return
	}
	if payload.SelectedRoute != nil && BrowserSubstratePosture(payload.SelectedRoute.Backend, payload.SelectedRoute.RuntimeTarget) != BrowserSubstrateLegacySystemHost {
		return
	}
	browserRuntimeHideImplicitLegacyHostSessionSurface(payload)
	browserRuntimeHideInspectionProjection(payload, action)
}

func browserRuntimeCanDegradeDefaultRouteFailure(
	ctx context.Context,
	action string,
	sessionRegistry *BrowserSessionRegistry,
	registry agentxbrowserruntime.SharedSessionBrowserStateRegistry,
	defaultRoute BrowserRuntimeInfo,
	substrate browserDefaultSubstrateAssessment,
	requestedProfile string,
	requestedTarget string,
) bool {
	if strings.TrimSpace(requestedTarget) != "" {
		return false
	}
	action = browserRuntimeCanonicalAction(action)
	if !browserImplicitLegacyHostRuntimeActionUsesDiagnosticsPath(action) {
		return false
	}
	requestedProfile = strings.TrimSpace(requestedProfile)
	if browserRuntimeUsesImplicitLegacyHostDefaultFallback(defaultRoute, substrate) {
		defaultProfile := strings.TrimSpace(firstNonEmpty(defaultRoute.Profile, defaultBrowserRuntimeInfo().Profile))
		normalizedProfile, ok := browserImplicitLegacyHostRuntimeCanUseCachedDiagnosticsSnapshot(action, requestedProfile, defaultProfile)
		if !ok {
			return false
		}
		requestedProfile = normalizedProfile
	}
	if requestedProfile == "" {
		return browserRuntimeHasDefaultRouteCachedDiagnosticsSnapshot(ctx, action, sessionRegistry, registry, defaultRoute)
	}
	defaultRoute = normalizeBrowserRuntimeInfo(defaultRoute)
	if browserRuntimeHasDefaultRouteStateProfileSnapshot(ctx, registry, defaultRoute, requestedProfile) {
		return true
	}
	switch action {
	case "status", "profiles", "sessions", "workbench":
		return browserRuntimeHasDefaultRouteSessionSnapshot(ctx, sessionRegistry, defaultRoute, requestedProfile)
	default:
		return false
	}
}

func browserRuntimeHasDefaultRouteCachedDiagnosticsSnapshot(
	ctx context.Context,
	action string,
	sessionRegistry *BrowserSessionRegistry,
	registry agentxbrowserruntime.SharedSessionBrowserStateRegistry,
	defaultRoute BrowserRuntimeInfo,
) bool {
	defaultRoute = normalizeBrowserRuntimeInfo(defaultRoute)
	defaultProfile := strings.TrimSpace(defaultRoute.Profile)
	switch browserRuntimeCanonicalAction(action) {
	case "status":
		if defaultProfile == "" {
			return false
		}
		if browserRuntimeHasDefaultRouteStateProfileSnapshot(ctx, registry, defaultRoute, defaultProfile) {
			return true
		}
		return browserRuntimeHasDefaultRouteSessionSnapshot(ctx, sessionRegistry, defaultRoute, defaultProfile)
	case "profiles":
		if browserRuntimeHasDefaultRouteStateProfileSnapshot(ctx, registry, defaultRoute, "") {
			return true
		}
		return browserRuntimeHasDefaultRouteSessionSnapshot(ctx, sessionRegistry, defaultRoute, "")
	case "sessions", "workbench":
		return true
	default:
		return false
	}
}

func browserRuntimeUsesManagedLaunchFailureSurface(action string) bool {
	switch browserRuntimeCanonicalAction(action) {
	case "prepare", "coordinate", "start":
		return true
	default:
		return false
	}
}

func browserRuntimeManagedLaunchFailureNoteForPreview(preview browserRuntimeDiagnosticsPreview) string {
	candidates := []string{
		preview.Registration.SubstrateSummary.NodePromotionFailureCause,
		preview.Registration.SubstrateAssessment.NodeRoute.FailureNote,
		preview.Registration.SubstrateAssessment.NodeRoute.FailureReason,
	}
	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		if note := browserRuntimeBootstrapBlockedSurfaceNoteForFailureText(candidate); note != "" {
			return note
		}
		if strings.Contains(strings.ToLower(candidate), "managed_route_unavailable") {
			return candidate
		}
	}
	return ""
}

func browserRuntimeMaybeApplyManagedLaunchFailureSurface(
	payload *browserRuntimePayload,
	action string,
	requestedTarget string,
	preview browserRuntimeDiagnosticsPreview,
) {
	if payload == nil || strings.TrimSpace(requestedTarget) != "" || !browserRuntimeUsesManagedLaunchFailureSurface(action) {
		return
	}
	if payload.Status != "unsupported" || !strings.Contains(strings.TrimSpace(payload.Note), "selected route does not support action") {
		return
	}
	if note := browserRuntimeManagedLaunchFailureNoteForPreview(preview); note != "" {
		payload.Note = browserRuntimeManagedLaunchFailureSurfaceNote(payload, note)
	}
}

func browserRuntimeManagedLaunchFailureSurfaceNote(payload *browserRuntimePayload, fallback string) string {
	return browserRuntimeManagedLaunchFailureSurfaceNoteWithCandidate(payload, nil, fallback)
}

func browserRuntimeManagedLaunchFailureSurfaceNoteWithCandidate(
	payload *browserRuntimePayload,
	candidate *browserRuntimeLaunchDiagnosticsSummary,
	fallback string,
) string {
	if note := browserRuntimeLaunchDiagnosticsSurfaceNote(candidate); note != "" {
		return note
	}
	if payload != nil {
		if note := browserRuntimeLaunchDiagnosticsSurfaceNote(payload.LaunchDiagnostics); note != "" {
			return note
		}
		if note := browserRuntimeLaunchDiagnosticsSurfaceNote(payload.WorkbenchLaunchDiagnostics); note != "" {
			return note
		}
	}
	return strings.TrimSpace(fallback)
}

func browserRuntimeLaunchDiagnosticsSurfaceNote(summary *browserRuntimeLaunchDiagnosticsSummary) string {
	if summary == nil {
		return ""
	}
	return firstNonEmpty(
		browserRuntimeDoctorLaunchBootstrapBlockedSummary(summary),
		browserRuntimeDoctorLaunchBaselineBlockedSummary(summary),
		browserRuntimeDoctorSelectedLaunchBlockedSummary(summary),
	)
}

func browserRuntimeHasDefaultRouteStateProfileSnapshot(
	ctx context.Context,
	registry agentxbrowserruntime.SharedSessionBrowserStateRegistry,
	defaultRoute BrowserRuntimeInfo,
	requestedProfile string,
) bool {
	sessionID, ok := browserRuntimeSessionStateRegistrySessionID(ctx, registry)
	if !ok {
		return false
	}
	return len(agentxbrowserruntime.SnapshotSharedSessionBrowserProjectedProfilesForScope(
		registry,
		sessionID,
		BrowserRuntimeInfo{
			Backend: defaultRoute.Backend,
			Target:  defaultRoute.Target,
		},
		requestedProfile,
	)) > 0
}

func browserRuntimeHasDefaultRouteSessionSnapshot(
	ctx context.Context,
	registry *BrowserSessionRegistry,
	defaultRoute BrowserRuntimeInfo,
	requestedProfile string,
) bool {
	if registry == nil {
		return false
	}
	sessionID := ToolSessionIDFromContext(ctx)
	if sessionID == "" {
		return false
	}
	return len(agentxbrowserruntime.SnapshotSharedSessionBrowserRoutes(
		registry,
		sessionID,
		BrowserSessionRoute{
			Backend: defaultRoute.Backend,
			Profile: strings.TrimSpace(requestedProfile),
			Target:  defaultRoute.Target,
		},
	)) > 0
}

func browserRuntimeSharedWatchManagerProvider(ctx browserRegistrationContext) agentxbrowserruntime.SharedSessionBrowserObserverManager {
	provider := ctx.watchManagerProvider
	if provider.SessionRegistry == nil && provider.RunRegistry == nil && provider.StateRegistry == nil && provider.ReconnectWindow == 0 {
		provider = agentxbrowserruntime.SharedSessionBrowserObserverManagerFor(
			ctx.sessionRegistry,
			ctx.sessionRunRegistry,
			ctx.sessionStateRegistry,
			browserRuntimeReconnectWatchdogWindow,
		)
	}
	return provider
}

func browserRuntimeSharedWatchManager(ctx browserRegistrationContext, control BrowserRuntimeControlBackend) agentxbrowserruntime.SharedSessionBrowserWatchManager {
	return browserRuntimeSharedWatchManagerProvider(ctx).Bind(control)
}

func browserRuntimeRouteDescriptorFromInfo(info BrowserRuntimeInfo) browserRuntimeRouteDescriptor {
	return browserRuntimeRouteDescriptorFromInfoWithProvenance(info, "", "")
}

func browserRuntimeRouteDescriptorFromInfoWithProvenance(
	info BrowserRuntimeInfo,
	source string,
	endpoint string,
) browserRuntimeRouteDescriptor {
	info = normalizeBrowserRuntimeInfo(info)
	return browserRuntimeRouteDescriptor{
		Backend:       info.Backend,
		Profile:       info.Profile,
		RuntimeTarget: info.Target,
		Source:        strings.TrimSpace(source),
		Endpoint:      strings.TrimSpace(endpoint),
	}
}

func browserRuntimeInfoFromRouteDescriptor(route *browserRuntimeRouteDescriptor) BrowserRuntimeInfo {
	if route == nil {
		return BrowserRuntimeInfo{}
	}
	return normalizeBrowserRuntimeInfo(BrowserRuntimeInfo{
		Backend: route.Backend,
		Profile: route.Profile,
		Target:  route.RuntimeTarget,
	})
}

func browserRuntimeRouteDescriptorPtr(info BrowserRuntimeInfo) *browserRuntimeRouteDescriptor {
	route := browserRuntimeRouteDescriptorFromInfo(info)
	if route == (browserRuntimeRouteDescriptor{}) {
		return nil
	}
	return &route
}

func browserRuntimeRouteDescriptorPtrWithProvenance(
	info BrowserRuntimeInfo,
	source string,
	endpoint string,
) *browserRuntimeRouteDescriptor {
	route := browserRuntimeRouteDescriptorFromInfoWithProvenance(info, source, endpoint)
	if route == (browserRuntimeRouteDescriptor{}) {
		return nil
	}
	return &route
}

func browserRuntimeRouteDescriptorPtrFromResolvedRoute(route browserResolvedExecutionRoute) *browserRuntimeRouteDescriptor {
	return browserRuntimeRouteDescriptorPtrWithProvenance(route.RuntimeInfo, route.Source, route.Endpoint)
}

func browserRuntimePayloadDefaultRouteDescriptor(ctx browserRegistrationContext) browserRuntimeRouteDescriptor {
	return browserRuntimeDiagnosticsPreviewBaseForRegistration(ctx).DefaultRouteDescriptor
}

func browserRegistrationSubstrateSummary(ctx browserRegistrationContext) BrowserWorkbenchSubstrateSummary {
	return browserRegistrationDefaultRuntimePreview(ctx).SubstrateSummary
}

func browserRegistrationSubstrateAssessment(ctx browserRegistrationContext) browserDefaultSubstrateAssessment {
	return browserRegistrationDefaultRuntimePreview(ctx).SubstrateAssessment
}

func browserRegistrationHasStoredSubstrateAssessment(assessment browserDefaultSubstrateAssessment) bool {
	return assessment.DefaultRuntime != (BrowserRuntimeInfo{}) ||
		assessment.HostRuntime != (BrowserRuntimeInfo{}) ||
		assessment.HostRoute.Configured ||
		assessment.NodeRoute.Configured ||
		assessment.SandboxRoute.Configured ||
		assessment.HostRoute.RouteAvailable ||
		assessment.NodeRoute.RouteAvailable ||
		assessment.SandboxRoute.RouteAvailable ||
		assessment.SandboxConcreteRoute.Configured ||
		assessment.SandboxConcreteRoute.RouteAvailable ||
		assessment.DefaultConcreteRoute.Configured ||
		assessment.DefaultConcreteRoute.RouteAvailable ||
		strings.TrimSpace(assessment.HostRoute.FailureReason) != "" ||
		strings.TrimSpace(assessment.NodeRoute.FailureReason) != "" ||
		strings.TrimSpace(assessment.SandboxRoute.FailureReason) != "" ||
		strings.TrimSpace(assessment.DefaultConcreteRoute.FailureReason) != "" ||
		strings.TrimSpace(assessment.DefaultConcreteRoute.FailureNote) != ""
}

func browserRegistrationHasStoredSubstrateSummary(summary BrowserWorkbenchSubstrateSummary) bool {
	return summary.SelectionStrategy != "" ||
		strings.TrimSpace(summary.SubstratePosture) != "" ||
		strings.TrimSpace(summary.SubstrateStatus) != "" ||
		strings.TrimSpace(summary.SubstrateReason) != "" ||
		len(summary.ConfiguredTargets) > 0 ||
		summary.DefaultRoute != (BrowserRuntimeInfo{}) ||
		summary.DefaultCandidateRoute != (BrowserRuntimeInfo{}) ||
		summary.HostRoute != (BrowserRuntimeInfo{}) ||
		summary.HostRouteAvailable ||
		strings.TrimSpace(summary.HostFailureCause) != "" ||
		strings.TrimSpace(summary.RepairCommand) != "" ||
		summary.NodeConfigured ||
		summary.SandboxConfigured ||
		summary.NodePromotionReady ||
		summary.SandboxPromotionReady ||
		strings.TrimSpace(summary.NodePromotionFailureCause) != "" ||
		strings.TrimSpace(summary.SandboxPromotionFailureCause) != "" ||
		strings.TrimSpace(summary.SandboxFailureCause) != ""
}

func browserConfiguredRuntimeTargets(ctx browserRegistrationContext) []string {
	return browserRuntimeDiagnosticsPreviewBaseForRegistration(ctx).ConfiguredTargets
}

func browserRuntimeAvailableActions(ctx browserRegistrationContext) []string {
	if len(ctx.runtimeActions) > 0 {
		return append([]string(nil), ctx.runtimeActions...)
	}
	return browserRuntimeComputeAvailableActions(ctx)
}

func browserRuntimeComputeAvailableActions(ctx browserRegistrationContext) []string {
	actions := browserRuntimeDiagnosticsPreviewForRegistration(ctx).AvailableActions
	if containsString(actions, "status") {
		actions = append(actions, "doctor")
	}
	return browserRuntimeAugmentActionsWithRepair(ctx.opts, actions)
}

func browserRuntimeDiagnosticsCapabilities(ctx browserRegistrationContext) BrowserCapabilities {
	capabilities := BrowserCapabilities{RuntimeStatus: true}
	if browserRuntimeHasSharedSessionStateRegistry(ctx.sessionStateRegistry) {
		capabilities.RuntimeWorkbench = true
		capabilities.RuntimeList = true
	}
	if ctx.sessionRegistry != nil {
		capabilities.RuntimeWorkbench = true
		capabilities.RuntimeSessions = true
	}
	return capabilities
}

func browserRuntimeCapabilityMetadataForCapabilities(ctx browserRegistrationContext, capabilities BrowserCapabilities) browserRuntimeCapabilityMetadata {
	if ctx.derivedCache != nil {
		return ctx.derivedCache.capabilityMetadataForCapabilities(ctx, capabilities)
	}
	return browserRuntimeCapabilityMetadataForCapabilitiesUncached(ctx, capabilities)
}

func browserRuntimeCapabilityMetadataForCapabilitiesUncached(ctx browserRegistrationContext, capabilities BrowserCapabilities) browserRuntimeCapabilityMetadata {
	surface := browserRuntimeRegisteredSurfaceForCapabilities(ctx, capabilities)
	return browserRuntimeCapabilityMetadata{
		RuntimeActions:   browserRuntimeAugmentActionsWithRepair(ctx.opts, capabilities.SupportedRuntimeActions()),
		BrowserTools:     surface.BrowserTools,
		ArtifactTools:    surface.ArtifactTools,
		ArtifactKinds:    browserRuntimeArtifactKinds(capabilities),
		ArtifactContract: browserRuntimeArtifactContract(capabilities),
		BrowserActKinds:  surface.BrowserActKinds,
		Capabilities:     browserRuntimeCapabilitiesMap(capabilities),
	}
}

func browserRuntimeApplyCapabilityMetadataToPayload(payload *browserRuntimePayload, metadata browserRuntimeCapabilityMetadata) {
	payload.RuntimeActions = metadata.RuntimeActions
	payload.BrowserTools = metadata.BrowserTools
	payload.ArtifactTools = metadata.ArtifactTools
	payload.ArtifactKinds = metadata.ArtifactKinds
	payload.ArtifactContract = metadata.ArtifactContract
	payload.BrowserActKinds = metadata.BrowserActKinds
	payload.BrowserSurface = strings.TrimSpace(metadata.BrowserSurface)
	payload.BrowserOptInTargets = append([]string(nil), metadata.BrowserOptInTargets...)
	payload.Capabilities = metadata.Capabilities
}

func browserRuntimeDiagnosticsMetadata(ctx browserRegistrationContext) browserRuntimeCapabilityMetadata {
	if ctx.derivedCache != nil {
		return ctx.derivedCache.diagnosticsCapabilityMetadata(ctx)
	}
	return browserRuntimeDiagnosticsMetadataUncached(ctx)
}

func browserRuntimeDiagnosticsMetadataUncached(ctx browserRegistrationContext) browserRuntimeCapabilityMetadata {
	baseCapabilities := browserRuntimeDiagnosticsCapabilities(ctx)
	optInSurface := browserRuntimeManagedOptInCapabilities(ctx)
	return browserRuntimeCapabilityMetadataForDiagnosticsSurface(ctx, baseCapabilities, optInSurface)
}

func browserRuntimeCapabilityMetadataForDiagnosticsSurface(
	ctx browserRegistrationContext,
	baseCapabilities BrowserCapabilities,
	optInSurface browserRuntimeManagedOptInDiagnosticsSurface,
) browserRuntimeCapabilityMetadata {
	if ctx.derivedCache != nil {
		return ctx.derivedCache.diagnosticsSurfaceCapabilityMetadata(ctx, baseCapabilities, optInSurface)
	}
	return browserRuntimeCapabilityMetadataForDiagnosticsSurfaceUncached(ctx, baseCapabilities, optInSurface)
}

func browserRuntimeCapabilityMetadataForDiagnosticsSurfaceUncached(
	ctx browserRegistrationContext,
	baseCapabilities BrowserCapabilities,
	optInSurface browserRuntimeManagedOptInDiagnosticsSurface,
) browserRuntimeCapabilityMetadata {
	mergedCapabilities := mergeBrowserCapabilities(baseCapabilities, optInSurface.Capabilities)
	surface := browserRuntimeMergeRegisteredSurfaces(
		browserRuntimeRegisteredSurfaceForCapabilities(ctx, baseCapabilities),
		browserRuntimeRegisteredSurfaceForCapabilities(ctx, optInSurface.Capabilities),
	)
	return browserRuntimeCapabilityMetadata{
		RuntimeActions:      browserRuntimeAugmentActionsWithRepair(ctx.opts, baseCapabilities.SupportedRuntimeActions()),
		BrowserTools:        surface.BrowserTools,
		ArtifactTools:       surface.ArtifactTools,
		ArtifactKinds:       mergeToolMetadataStrings(browserRuntimeArtifactKinds(baseCapabilities), browserRuntimeArtifactKinds(optInSurface.Capabilities)),
		ArtifactContract:    browserRuntimeArtifactContract(mergedCapabilities),
		BrowserActKinds:     surface.BrowserActKinds,
		BrowserSurface:      browserRuntimeManagedOptInSurfaceLabel(optInSurface),
		BrowserOptInTargets: append([]string(nil), optInSurface.Targets...),
		Capabilities:        browserRuntimeCapabilitiesMap(mergedCapabilities),
	}
}

func browserRuntimeHasSharedSessionStateRegistry(registry agentxbrowserruntime.SharedSessionBrowserStateRegistry) bool {
	if registry == nil {
		return false
	}
	if concrete, ok := registry.(*BrowserSessionStateRegistry); ok {
		return concrete != nil
	}
	return true
}

func browserRuntimeAvailableActionRouteAssessments(ctx browserRegistrationContext, substrate browserDefaultSubstrateAssessment, summary BrowserWorkbenchSubstrateSummary) []browserConcreteRouteAssessment {
	return browserRuntimeAvailableActionRouteAssessmentsForBackend(
		ctx,
		browserRegistrationDefaultRuntimePreview(ctx).EffectiveBackend,
		substrate,
		summary,
	)
}

func browserRuntimeAvailableActionRouteAssessmentsForBackend(ctx browserRegistrationContext, backend BrowserBackend, substrate browserDefaultSubstrateAssessment, summary BrowserWorkbenchSubstrateSummary) []browserConcreteRouteAssessment {
	assessments := []browserConcreteRouteAssessment{
		browserRuntimeSubstrateRouteAssessmentForBackend(
			backend,
			BrowserRuntimeInfo{Target: "host"},
			substrate.HostRoute,
		),
	}
	for _, managed := range browserRuntimeManagedRouteAssessments(ctx.opts, substrate, summary) {
		if !managed.Configured {
			continue
		}
		assessments = append(assessments, browserRuntimeSubstrateRouteAssessmentForBackend(
			backend,
			BrowserRuntimeInfo{Target: managed.Role},
			managed.Assessment,
		))
	}
	return assessments
}

func browserRuntimeManagedLaunchFailureActionCapabilities(ctx browserRegistrationContext) BrowserCapabilities {
	capabilities := browserRuntimeDiagnosticsCapabilities(ctx)
	capabilities.RuntimePrepare = true
	capabilities.RuntimeCoordinate = true
	capabilities.RuntimeStart = true
	capabilities.RuntimeList = true
	return capabilities
}

func browserRuntimeShouldKeepManagedLaunchFailureActionSurface(assessment browserConcreteRouteAssessment) bool {
	if !assessment.Configured || assessment.RouteAvailable {
		return false
	}
	failure := strings.ToLower(strings.TrimSpace(firstNonEmpty(
		assessment.FailureNote,
		assessment.FailureReason,
	)))
	return strings.Contains(failure, "managed_route_unavailable")
}

func browserRuntimeActionCapabilitiesForAssessment(ctx browserRegistrationContext, assessment browserConcreteRouteAssessment) (BrowserCapabilities, bool) {
	if capabilities, ok := browserRuntimeActionCapabilitiesForResolvedRoute(ctx, assessment.Route, assessment.RouteAvailable); ok {
		return capabilities, true
	}
	if !browserRuntimeShouldKeepManagedLaunchFailureActionSurface(assessment) {
		return BrowserCapabilities{}, false
	}
	return browserRuntimeManagedLaunchFailureActionCapabilities(ctx), true
}

func browserRuntimeActionCapabilitiesForConcreteRoute(ctx browserRegistrationContext, backend BrowserBackend, fallback BrowserRuntimeInfo) (BrowserCapabilities, bool) {
	assessment := browserConcreteRouteAssessmentForConcreteBackend(backend, fallback, normalizeBrowserRuntimeInfo(fallback).Target)
	return browserRuntimeActionCapabilitiesForAssessment(ctx, assessment)
}

func browserRuntimeCapabilitiesForResolvedRoute(ctx browserRegistrationContext, route browserResolvedExecutionRoute) BrowserCapabilities {
	capabilities := route.Capabilities
	if capabilities == (BrowserCapabilities{}) {
		capabilities = browserCapabilitiesForConcreteBackend(route.Backend)
	}
	if capabilities == (BrowserCapabilities{}) {
		capabilities = defaultBrowserCapabilities()
	}
	backend := route.Backend
	if _, ok := backend.(BrowserRuntimeControlBackend); ok {
		capabilities.RuntimeStatus = true
		capabilities.RuntimeWorkbench = true
		capabilities.RuntimePrepare = true
		capabilities.RuntimeCoordinate = true
		capabilities.RuntimeStart = true
		capabilities.RuntimeRestart = true
		capabilities.RuntimeStop = true
		capabilities.RuntimeList = true
	}
	if _, ok := backend.(BrowserRuntimeProfileManagementBackend); ok {
		capabilities.RuntimeCreate = true
		capabilities.RuntimeDelete = true
	}
	if ctx.sessionStateRegistry != nil {
		capabilities.RuntimeWorkbench = true
		capabilities.RuntimeSelect = true
		capabilities.RuntimeClear = true
	}
	if ctx.sessionStateRegistry != nil || ctx.sessionRegistry != nil {
		capabilities.RuntimeWorkbench = true
		capabilities.RuntimeClearSession = true
		capabilities.RuntimeSyncSession = true
	}
	if ctx.sessionRegistry != nil {
		capabilities.RuntimeWorkbench = true
		capabilities.RuntimeSelectTarget = true
		capabilities.RuntimeClearTarget = true
		capabilities.RuntimeSessions = true
	}
	return capabilities
}

func browserRuntimeActionCapabilitiesForResolvedRoute(ctx browserRegistrationContext, route browserResolvedExecutionRoute, available bool) (BrowserCapabilities, bool) {
	if !available {
		return BrowserCapabilities{}, false
	}
	return browserRuntimeCapabilitiesForResolvedRoute(ctx, route), true
}

func browserRuntimeActionSupported(capabilities BrowserCapabilities, action string) bool {
	action = browserRuntimeCanonicalAction(action)
	switch action {
	case "", "status":
		return true
	case "workbench":
		return capabilities.RuntimeWorkbench
	case "prepare":
		return capabilities.RuntimePrepare
	case "coordinate":
		return capabilities.RuntimeCoordinate
	case "start":
		return capabilities.RuntimeStart
	case "restart":
		return capabilities.RuntimeRestart
	case "refresh":
		return capabilities.RuntimeRestart
	case "stop":
		return capabilities.RuntimeStop
	case "create_profile":
		return capabilities.RuntimeCreate
	case "delete_profile":
		return capabilities.RuntimeDelete
	case "select_profile":
		return capabilities.RuntimeSelect
	case "clear_profile":
		return capabilities.RuntimeClear
	case "clear_session":
		return capabilities.RuntimeClearSession
	case "sync_session":
		return capabilities.RuntimeSyncSession
	case "select_target":
		return capabilities.RuntimeSelectTarget
	case "clear_target":
		return capabilities.RuntimeClearTarget
	case "profiles":
		return capabilities.RuntimeList
	case "sessions":
		return capabilities.RuntimeSessions
	default:
		return false
	}
}

func browserRuntimePopulatePayload(ctx browserRegistrationContext, payload *browserRuntimePayload, route browserResolvedExecutionRoute) BrowserCapabilities {
	projection, capabilities := browserRuntimeSelectedRouteProjection(
		ctx,
		browserRuntimeDiagnosticsPreviewBaseForRegistration(ctx),
		route,
	)
	browserRuntimeApplyTopLevelRouteProjection(payload, projection)
	return capabilities
}

func browserRuntimeProfileStatePtr(result BrowserProfileStatusResult) *browserRuntimeProfileState {
	state := browserRuntimeProfileStateFromSharedSessionState(agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:    strings.TrimSpace(result.Backend),
		Profile:    strings.TrimSpace(result.Profile),
		BrowserApp: strings.TrimSpace(result.BrowserApp),
		Status:     strings.TrimSpace(result.Status),
		Running:    result.Running,
		Connected:  result.Connected,
		Note:       strings.TrimSpace(result.Note),
	})
	state.RuntimeTarget = ""
	if state.Profile == "" && state.BrowserApp == "" && state.Status == "" && !state.Running && !state.Connected && state.Note == "" {
		return nil
	}
	return &state
}

func browserRuntimeProfileStatePtrFromSharedSessionState(state agentxbrowserruntime.SharedSessionBrowserProfileState) *browserRuntimeProfileState {
	profileState := browserRuntimeProfileStateFromSharedSessionState(state)
	if profileState.Profile == "" && profileState.BrowserApp == "" && profileState.Status == "" && !profileState.Running && !profileState.Connected && profileState.Note == "" {
		return nil
	}
	return &profileState
}

func browserRuntimeProfileStatusResultFromSharedSessionState(state agentxbrowserruntime.SharedSessionBrowserProfileState) BrowserProfileStatusResult {
	state = browserRuntimeNormalizeSessionProfileState(state)
	return BrowserProfileStatusResult{
		Backend:    strings.TrimSpace(state.Backend),
		BrowserApp: strings.TrimSpace(state.BrowserApp),
		Profile:    strings.TrimSpace(state.Profile),
		Status:     strings.TrimSpace(state.Status),
		Running:    state.Running,
		Connected:  state.Connected,
		Note:       strings.TrimSpace(state.Note),
	}
}

func browserRuntimeProfileStateFromSharedSessionState(state agentxbrowserruntime.SharedSessionBrowserProfileState) browserRuntimeProfileState {
	state = browserRuntimeNormalizeSessionProfileState(state)
	return browserRuntimeProfileState{
		Backend:       strings.TrimSpace(state.Backend),
		Profile:       strings.TrimSpace(state.Profile),
		RuntimeTarget: strings.TrimSpace(state.RuntimeTarget),
		BrowserApp:    strings.TrimSpace(state.BrowserApp),
		Status:        strings.TrimSpace(state.Status),
		Running:       state.Running,
		Connected:     state.Connected,
		Note:          strings.TrimSpace(state.Note),
		ObservedAt:    state.ObservedAt,
		StatusSince:   state.StatusSince,
	}
}

func browserRuntimeSharedSessionProfileState(state browserRuntimeProfileState) agentxbrowserruntime.SharedSessionBrowserProfileState {
	return agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       strings.TrimSpace(state.Backend),
		Profile:       strings.TrimSpace(state.Profile),
		RuntimeTarget: strings.TrimSpace(state.RuntimeTarget),
		BrowserApp:    strings.TrimSpace(state.BrowserApp),
		Status:        strings.TrimSpace(state.Status),
		Running:       state.Running,
		Connected:     state.Connected,
		Note:          strings.TrimSpace(state.Note),
		ObservedAt:    state.ObservedAt,
		StatusSince:   state.StatusSince,
	}
}

func browserRuntimeSharedSessionProfileStates(items []browserRuntimeProfileState) []agentxbrowserruntime.SharedSessionBrowserProfileState {
	if len(items) == 0 {
		return nil
	}
	out := make([]agentxbrowserruntime.SharedSessionBrowserProfileState, 0, len(items))
	for _, item := range items {
		out = append(out, browserRuntimeSharedSessionProfileState(item))
	}
	return out
}

func browserRuntimeProfileStatePtrFromResultOrState(result BrowserProfileStatusResult, state agentxbrowserruntime.SharedSessionBrowserProfileState, ok bool) *browserRuntimeProfileState {
	if ok {
		if profileState := browserRuntimeProfileStatePtrFromSharedSessionState(state); profileState != nil {
			return profileState
		}
	}
	return browserRuntimeProfileStatePtr(result)
}

func browserRuntimeEffectiveProfileStatePtrFromBinding(ctx context.Context, registry agentxbrowserruntime.SharedSessionBrowserStateRegistry, binding *browserRuntimeSessionBinding, payload *browserRuntimePayload, selectedInfo BrowserRuntimeInfo) *browserRuntimeProfileState {
	if payload == nil {
		return nil
	}
	profileSelection := payload.SessionProfileSelection
	targetSelection := payload.SessionTargetSelection
	selectedRoute := payload.SelectedRoute
	profile := strings.TrimSpace(firstNonEmpty(
		payload.RequestedProfile,
		func() string {
			if profileSelection != nil {
				return strings.TrimSpace(profileSelection.Profile)
			}
			return ""
		}(),
		func() string {
			if targetSelection != nil {
				return strings.TrimSpace(targetSelection.Profile)
			}
			return ""
		}(),
		func() string {
			if binding != nil {
				return strings.TrimSpace(binding.SelectedBrowserProfile)
			}
			return ""
		}(),
		func() string {
			if binding != nil {
				return strings.TrimSpace(binding.ActiveBrowserProfile)
			}
			return ""
		}(),
		func() string {
			if selectedRoute != nil {
				return strings.TrimSpace(selectedRoute.Profile)
			}
			return ""
		}(),
		strings.TrimSpace(selectedInfo.Profile),
	))
	status := browserRuntimeEffectiveRouteProfileStatus(ctx, registry, binding, profile, selectedInfo, BrowserProfileStatusResult{
		Backend: firstNonEmpty(
			func() string {
				if targetSelection != nil {
					return strings.TrimSpace(targetSelection.Backend)
				}
				return ""
			}(),
			strings.TrimSpace(selectedInfo.Backend),
		),
		BrowserApp: firstNonEmpty(
			func() string {
				if targetSelection != nil {
					return strings.TrimSpace(targetSelection.BrowserApp)
				}
				return ""
			}(),
			func() string {
				if profileSelection != nil {
					return strings.TrimSpace(profileSelection.BrowserApp)
				}
				return ""
			}(),
		),
		Profile: profile,
	})
	return browserRuntimeProfileStatePtrFromSharedSessionState(
		agentxbrowserruntime.SharedSessionBrowserProfileStateFromStatus(selectedInfo, status),
	)
}

func browserRuntimeEffectiveRouteProfileStatus(ctx context.Context, registry agentxbrowserruntime.SharedSessionBrowserStateRegistry, binding *browserRuntimeSessionBinding, profile string, selectedInfo BrowserRuntimeInfo, fallback BrowserProfileStatusResult) BrowserProfileStatusResult {
	input := agentxbrowserruntime.SharedSessionBrowserHealthInput{}
	if binding != nil {
		input = browserRuntimeSessionHealthInputFromBinding(*binding)
	}
	if sessionID, ok := browserRuntimeSessionStateRegistrySessionID(ctx, registry); ok {
		return agentxbrowserruntime.ResolveSharedSessionBrowserProfileStatusForScope(
			registry,
			sessionID,
			selectedInfo,
			profile,
			input,
			fallback,
			browserRuntimeReconnectWatchdogWindow,
		)
	}
	return agentxbrowserruntime.ResolveSharedSessionBrowserProfileStatus(
		input,
		selectedInfo,
		profile,
		fallback,
		browserRuntimeReconnectWatchdogWindow,
	)
}

func browserRuntimeConfiguredProfiles(payload browserRuntimePayload, selectedInfo BrowserRuntimeInfo) []string {
	return agentxbrowserruntime.ProjectSharedSessionBrowserConfiguredProfiles(
		browserRuntimeSharedConfiguredProfilesProjectionRequest(payload, selectedInfo, false),
	)
}

func firstValueOrZero(result *BrowserProfileStatusResult) BrowserProfileStatusResult {
	if result == nil {
		return BrowserProfileStatusResult{}
	}
	return *result
}

func browserRuntimeSelectSessionProfile(ctx context.Context, watchManagerProvider agentxbrowserruntime.SharedSessionBrowserObserverManager, selectedInfo BrowserRuntimeInfo, browserApp string, control BrowserRuntimeControlBackend, capabilities BrowserCapabilities, source string) (*browserRuntimeSessionProfileSelection, string, error) {
	sessionID := ToolSessionIDFromContext(ctx)
	selection, decision, ok, err := agentxbrowserruntime.SelectSharedSessionBrowserProfileEvent(
		ctx,
		watchManagerProvider.SessionRegistry,
		watchManagerProvider.RunRegistry,
		watchManagerProvider.StateRegistry,
		sessionID,
		selectedInfo,
		strings.TrimSpace(browserApp),
		control,
		control != nil && browserRuntimeActionSupported(capabilities, "profiles"),
		strings.TrimSpace(source),
		browserRuntimeReconnectWatchdogWindow,
	)
	if err != nil || !ok {
		return nil, decision, err
	}
	return browserRuntimeSelectionPtr(selection), decision, nil
}

func browserRuntimeSessionProfileSelectionPtr(ctx context.Context, registry agentxbrowserruntime.SharedSessionBrowserStateRegistry, selectedRoute *browserRuntimeRouteDescriptor) *browserRuntimeSessionProfileSelection {
	if registry == nil {
		return nil
	}
	sessionID := ToolSessionIDFromContext(ctx)
	if sessionID == "" {
		return nil
	}
	runtimeTarget := ""
	if selectedRoute != nil {
		runtimeTarget = strings.TrimSpace(selectedRoute.RuntimeTarget)
	}
	selection, ok := registry.SelectedBrowserProfile(sessionID, runtimeTarget)
	if !ok {
		return nil
	}
	return browserRuntimeSelectionPtr(selection)
}

func browserRuntimeSelectionPtr(selection agentxbrowserruntime.SharedSessionBrowserProfileSelection) *browserRuntimeSessionProfileSelection {
	profile := strings.TrimSpace(selection.Profile)
	runtimeTarget := strings.TrimSpace(selection.RuntimeTarget)
	if profile == "" && runtimeTarget == "" && strings.TrimSpace(selection.Backend) == "" && strings.TrimSpace(selection.BrowserApp) == "" && strings.TrimSpace(selection.Source) == "" {
		return nil
	}
	return &browserRuntimeSessionProfileSelection{
		Backend:       strings.TrimSpace(selection.Backend),
		Profile:       profile,
		RuntimeTarget: runtimeTarget,
		BrowserApp:    strings.TrimSpace(selection.BrowserApp),
		Source:        strings.TrimSpace(selection.Source),
	}
}

func browserRuntimeSelectionPtrValue(selection *agentxbrowserruntime.SharedSessionBrowserProfileSelection) *browserRuntimeSessionProfileSelection {
	if selection == nil {
		return nil
	}
	return browserRuntimeSelectionPtr(*selection)
}

func browserRuntimeSharedProfileSelection(selection *browserRuntimeSessionProfileSelection) *agentxbrowserruntime.SharedSessionBrowserProfileSelection {
	if selection == nil {
		return nil
	}
	if strings.TrimSpace(selection.Profile) == "" &&
		strings.TrimSpace(selection.RuntimeTarget) == "" &&
		strings.TrimSpace(selection.Backend) == "" &&
		strings.TrimSpace(selection.BrowserApp) == "" &&
		strings.TrimSpace(selection.Source) == "" {
		return nil
	}
	return &agentxbrowserruntime.SharedSessionBrowserProfileSelection{
		Backend:       strings.TrimSpace(selection.Backend),
		Profile:       strings.TrimSpace(selection.Profile),
		RuntimeTarget: strings.TrimSpace(selection.RuntimeTarget),
		BrowserApp:    strings.TrimSpace(selection.BrowserApp),
		Source:        strings.TrimSpace(selection.Source),
	}
}

func browserRuntimeSharedProfileSelectionFromBinding(binding browserRuntimeSessionBinding) *agentxbrowserruntime.SharedSessionBrowserProfileSelection {
	return agentxbrowserruntime.ProjectSharedSessionBrowserProfileSelectionFromBindingSnapshot(
		browserRuntimeSharedProfileSelection(&browserRuntimeSessionProfileSelection{
			Backend:       strings.TrimSpace(binding.SelectedBrowserBackend),
			Profile:       strings.TrimSpace(binding.SelectedBrowserProfile),
			RuntimeTarget: strings.TrimSpace(binding.SelectedBrowserTarget),
			BrowserApp:    strings.TrimSpace(binding.SelectedBrowserApp),
			Source:        strings.TrimSpace(binding.SelectedBrowserProfileSource),
		}),
		browserRuntimeSharedTargetSelection(&browserRuntimeSessionTargetSelection{
			ID:            strings.TrimSpace(binding.SelectedBrowserTargetID),
			TabIndex:      binding.SelectedBrowserTabIndex,
			Backend:       strings.TrimSpace(binding.SelectedBrowserBackend),
			Profile:       strings.TrimSpace(binding.SelectedBrowserProfile),
			RuntimeTarget: strings.TrimSpace(binding.SelectedBrowserTarget),
			BrowserApp:    strings.TrimSpace(binding.SelectedBrowserApp),
			Source:        strings.TrimSpace(binding.SelectedBrowserTargetSource),
		}),
		browserRuntimeSharedSessionProfileStates(binding.BrowserProfiles),
	)
}

func browserRuntimeSharedTargetSelectionFromBinding(binding browserRuntimeSessionBinding) *agentxbrowserruntime.BrowserSessionTargetSelection {
	return agentxbrowserruntime.ProjectSharedSessionBrowserTargetSelectionFromBindingSnapshot(
		browserRuntimeSharedTargetSelection(&browserRuntimeSessionTargetSelection{
			ID:            strings.TrimSpace(binding.SelectedBrowserTargetID),
			TabIndex:      binding.SelectedBrowserTabIndex,
			Backend:       strings.TrimSpace(binding.SelectedBrowserBackend),
			Profile:       strings.TrimSpace(binding.SelectedBrowserProfile),
			RuntimeTarget: strings.TrimSpace(binding.SelectedBrowserTarget),
			BrowserApp:    strings.TrimSpace(binding.SelectedBrowserApp),
			Source:        strings.TrimSpace(binding.SelectedBrowserTargetSource),
		}),
		browserRuntimeSharedProfileSelection(&browserRuntimeSessionProfileSelection{
			Backend:       strings.TrimSpace(binding.SelectedBrowserBackend),
			Profile:       strings.TrimSpace(binding.SelectedBrowserProfile),
			RuntimeTarget: strings.TrimSpace(binding.SelectedBrowserTarget),
			BrowserApp:    strings.TrimSpace(binding.SelectedBrowserApp),
			Source:        strings.TrimSpace(binding.SelectedBrowserProfileSource),
		}),
		browserRuntimeSharedSessionProfileStates(binding.BrowserProfiles),
	)
}

func browserRuntimeSharedTargetSelection(selection *browserRuntimeSessionTargetSelection) *agentxbrowserruntime.BrowserSessionTargetSelection {
	if selection == nil {
		return nil
	}
	if strings.TrimSpace(selection.ID) == "" &&
		selection.TabIndex <= 0 &&
		strings.TrimSpace(selection.Source) == "" &&
		strings.TrimSpace(selection.URL) == "" &&
		strings.TrimSpace(selection.Title) == "" &&
		strings.TrimSpace(selection.Backend) == "" &&
		strings.TrimSpace(selection.Profile) == "" &&
		strings.TrimSpace(selection.RuntimeTarget) == "" &&
		strings.TrimSpace(selection.BrowserApp) == "" {
		return nil
	}
	return &agentxbrowserruntime.BrowserSessionTargetSelection{
		ID:            strings.TrimSpace(selection.ID),
		TabIndex:      selection.TabIndex,
		URL:           strings.TrimSpace(selection.URL),
		Title:         strings.TrimSpace(selection.Title),
		Backend:       strings.TrimSpace(selection.Backend),
		Profile:       strings.TrimSpace(selection.Profile),
		RuntimeTarget: strings.TrimSpace(selection.RuntimeTarget),
		BrowserApp:    strings.TrimSpace(selection.BrowserApp),
		Source:        strings.TrimSpace(selection.Source),
	}
}

func browserRuntimeSessionTargetSelectionPtrFromShared(selection *agentxbrowserruntime.BrowserSessionTargetSelection) *browserRuntimeSessionTargetSelection {
	if selection == nil {
		return nil
	}
	return &browserRuntimeSessionTargetSelection{
		ID:            strings.TrimSpace(selection.ID),
		TabIndex:      selection.TabIndex,
		URL:           strings.TrimSpace(selection.URL),
		Title:         strings.TrimSpace(selection.Title),
		Backend:       strings.TrimSpace(selection.Backend),
		Profile:       strings.TrimSpace(selection.Profile),
		RuntimeTarget: strings.TrimSpace(selection.RuntimeTarget),
		BrowserApp:    strings.TrimSpace(selection.BrowserApp),
		Source:        strings.TrimSpace(selection.Source),
	}
}

func browserRuntimeSyncSession(ctx context.Context, watchManagerProvider agentxbrowserruntime.SharedSessionBrowserObserverManager, selectedInfo BrowserRuntimeInfo, selectedRoute *browserRuntimeRouteDescriptor, browserApp string, control BrowserRuntimeControlBackend, capabilities BrowserCapabilities) (*browserRuntimeSessionProfileSelection, *browserRuntimeSessionTargetSelection, string, bool, error) {
	if selectedRoute == nil {
		return nil, nil, "", false, nil
	}
	result, err := agentxbrowserruntime.SyncSharedSessionBrowserRouteSelectionEvent(
		ctx,
		watchManagerProvider.SessionRegistry,
		watchManagerProvider.RunRegistry,
		watchManagerProvider.StateRegistry,
		ToolSessionIDFromContext(ctx),
		selectedInfo,
		browserRuntimeSessionRouteFilter(selectedRoute),
		browserApp,
		control,
		control != nil && browserRuntimeActionSupported(capabilities, "profiles"),
		"sync_session",
		browserRuntimeReconnectWatchdogWindow,
	)
	return browserRuntimeSelectionPtrValue(result.ProfileSelection), browserRuntimeSessionTargetSelectionPtrFromShared(result.TargetSelection), result.Decision, result.Ready, err
}

func browserRuntimeSyncOrClearCurrentTargetForProfileSelection(ctx context.Context, watchManagerProvider agentxbrowserruntime.SharedSessionBrowserObserverManager, selectedRoute *browserRuntimeRouteDescriptor, profileSelection *browserRuntimeSessionProfileSelection, source string) (*browserRuntimeSessionTargetSelection, string, error) {
	if selectedRoute == nil {
		return nil, "", nil
	}
	sessionID := ToolSessionIDFromContext(ctx)
	targetSelection, decision, err := agentxbrowserruntime.SyncOrClearSharedSessionBrowserCurrentTargetForProfileSelectionEvent(
		watchManagerProvider.SessionRegistry,
		watchManagerProvider.RunRegistry,
		watchManagerProvider.StateRegistry,
		sessionID,
		browserRuntimeSessionRouteFilter(selectedRoute),
		browserRuntimeSharedProfileSelection(profileSelection),
		source,
		browserRuntimeReconnectWatchdogWindow,
	)
	return browserRuntimeSessionTargetSelectionPtrFromShared(targetSelection), decision, err
}

func browserRuntimeShouldClearSessionTargetOnProfileSelect(ctx context.Context, registry *BrowserSessionRegistry, selectedRoute *browserRuntimeRouteDescriptor, profileSelection *browserRuntimeSessionProfileSelection) bool {
	if selectedRoute == nil || profileSelection == nil {
		return false
	}
	return agentxbrowserruntime.ShouldClearSharedSessionBrowserTargetOnProfileSelect(
		registry,
		ToolSessionIDFromContext(ctx),
		browserRuntimeSessionRouteFilter(selectedRoute),
		browserRuntimeSharedProfileSelection(profileSelection),
	)
}

func browserRuntimeTargetSelectionMatchesRoute(selection *browserRuntimeSessionTargetSelection, selectedRoute *browserRuntimeRouteDescriptor) bool {
	if selection == nil || selectedRoute == nil {
		return false
	}
	if selected := strings.TrimSpace(selection.Backend); selected != "" && strings.TrimSpace(selectedRoute.Backend) != "" && !browserRuntimeBackendMatches(selected, strings.TrimSpace(selectedRoute.Backend)) {
		return false
	}
	if selected := strings.TrimSpace(selection.RuntimeTarget); selected != "" && strings.TrimSpace(selectedRoute.RuntimeTarget) != "" && !strings.EqualFold(selected, strings.TrimSpace(selectedRoute.RuntimeTarget)) {
		return false
	}
	if selected := strings.TrimSpace(selection.Profile); selected != "" && strings.TrimSpace(selectedRoute.Profile) != "" && !strings.EqualFold(selected, strings.TrimSpace(selectedRoute.Profile)) {
		return false
	}
	return true
}

func browserRuntimeSessionTargetSelectionPtr(ctx context.Context, registry *BrowserSessionRegistry, selectedRoute *browserRuntimeRouteDescriptor) *browserRuntimeSessionTargetSelection {
	if selectedRoute == nil {
		return nil
	}
	return browserRuntimeSessionTargetSelectionPtrFromShared(
		agentxbrowserruntime.CurrentSharedSessionBrowserTargetSelection(
			registry,
			ToolSessionIDFromContext(ctx),
			browserRuntimeSessionRouteFilter(selectedRoute),
		),
	)
}

func browserRuntimePromoteProfileFromTargetSelection(ctx context.Context, stateRegistry agentxbrowserruntime.SharedSessionBrowserStateRegistry, selection *browserRuntimeSessionTargetSelection) *browserRuntimeSessionProfileSelection {
	sharedSelection, ok := agentxbrowserruntime.PromoteSharedSessionBrowserProfileFromTargetSelection(
		stateRegistry,
		ToolSessionIDFromContext(ctx),
		&agentxbrowserruntime.BrowserSessionTargetSelection{
			ID:            strings.TrimSpace(selection.ID),
			TabIndex:      selection.TabIndex,
			URL:           strings.TrimSpace(selection.URL),
			Title:         strings.TrimSpace(selection.Title),
			Backend:       strings.TrimSpace(selection.Backend),
			Profile:       strings.TrimSpace(selection.Profile),
			RuntimeTarget: strings.TrimSpace(selection.RuntimeTarget),
			BrowserApp:    strings.TrimSpace(selection.BrowserApp),
			Source:        strings.TrimSpace(selection.Source),
		},
	)
	if !ok {
		return nil
	}
	return browserRuntimeSelectionPtr(sharedSelection)
}

func browserRuntimeApplySessionSelectionsToRoute(route *browserRuntimeRouteDescriptor, profileSelection *browserRuntimeSessionProfileSelection, targetSelection *browserRuntimeSessionTargetSelection) {
	if route == nil {
		return
	}
	if profileSelection != nil {
		if backend := strings.TrimSpace(profileSelection.Backend); backend != "" {
			route.Backend = backend
		}
		if profile := strings.TrimSpace(profileSelection.Profile); profile != "" {
			route.Profile = profile
		}
		if runtimeTarget := strings.TrimSpace(profileSelection.RuntimeTarget); runtimeTarget != "" {
			route.RuntimeTarget = runtimeTarget
		}
	}
	if targetSelection != nil {
		if backend := strings.TrimSpace(targetSelection.Backend); backend != "" {
			route.Backend = backend
		}
		if profile := strings.TrimSpace(targetSelection.Profile); profile != "" {
			route.Profile = profile
		}
		if runtimeTarget := strings.TrimSpace(targetSelection.RuntimeTarget); runtimeTarget != "" {
			route.RuntimeTarget = runtimeTarget
		}
	}
}

func browserRuntimeSessionRouteFilter(selectedRoute *browserRuntimeRouteDescriptor) BrowserSessionRoute {
	if selectedRoute == nil {
		return BrowserSessionRoute{}
	}
	return BrowserSessionRoute{
		Backend:    strings.TrimSpace(selectedRoute.Backend),
		Profile:    strings.TrimSpace(selectedRoute.Profile),
		Target:     strings.TrimSpace(selectedRoute.RuntimeTarget),
		BrowserApp: "",
	}
}

func browserRuntimeSessionRoutes(ctx context.Context, registry *BrowserSessionRegistry, selectedRoute *browserRuntimeRouteDescriptor) []browserRuntimeSessionRoute {
	snapshot := agentxbrowserruntime.SnapshotSharedSessionBrowserRoutes(
		registry,
		ToolSessionIDFromContext(ctx),
		browserRuntimeSessionRouteFilter(selectedRoute),
	)
	return browserRuntimeSessionRoutesFromShared(snapshot)
}

func browserRuntimeSessionRoutesFromShared(snapshot []agentxbrowserruntime.SharedSessionBrowserRouteSnapshot) []browserRuntimeSessionRoute {
	if len(snapshot) == 0 {
		return nil
	}
	out := make([]browserRuntimeSessionRoute, 0, len(snapshot))
	for _, routeState := range snapshot {
		targets := make([]browserRuntimeSessionTarget, 0, len(routeState.Targets))
		for _, target := range routeState.Targets {
			targets = append(targets, browserRuntimeSessionTarget{
				ID:            strings.TrimSpace(target.ID),
				TabIndex:      target.TabIndex,
				URL:           strings.TrimSpace(target.URL),
				Title:         strings.TrimSpace(target.Title),
				BrowserApp:    strings.TrimSpace(target.BrowserApp),
				Backend:       strings.TrimSpace(target.Backend),
				Profile:       strings.TrimSpace(target.Profile),
				RuntimeTarget: strings.TrimSpace(target.RuntimeTarget),
				Current:       target.Current,
			})
		}
		out = append(out, browserRuntimeSessionRoute{
			Backend:                  strings.TrimSpace(routeState.Backend),
			Profile:                  strings.TrimSpace(routeState.Profile),
			RuntimeTarget:            strings.TrimSpace(routeState.RuntimeTarget),
			BrowserApp:               strings.TrimSpace(routeState.BrowserApp),
			CurrentTargetID:          strings.TrimSpace(routeState.CurrentTargetID),
			CurrentTargetSource:      strings.TrimSpace(routeState.CurrentTargetSource),
			PendingTargetReview:      browserRuntimeSessionTargetReviewPtr(routeState.PendingTargetReview),
			PendingTargetReviewCount: routeState.PendingTargetReviewCount,
			FollowPolicyState:        strings.TrimSpace(routeState.FollowPolicyState),
			FollowPolicyReason:       strings.TrimSpace(routeState.FollowPolicyReason),
			PopupPolicyState:         strings.TrimSpace(routeState.PopupPolicyState),
			PopupPolicyReason:        strings.TrimSpace(routeState.PopupPolicyReason),
			Targets:                  targets,
		})
	}
	return out
}

func browserRuntimeSessionTargetReviewPtr(review *agentxbrowserruntime.BrowserSessionTargetReview) *browserRuntimeSessionTargetReview {
	if review == nil {
		return nil
	}
	return &browserRuntimeSessionTargetReview{
		ID:            strings.TrimSpace(review.ID),
		TabIndex:      review.TabIndex,
		URL:           strings.TrimSpace(review.URL),
		Title:         strings.TrimSpace(review.Title),
		BrowserApp:    strings.TrimSpace(review.BrowserApp),
		Backend:       strings.TrimSpace(review.Backend),
		Profile:       strings.TrimSpace(review.Profile),
		RuntimeTarget: strings.TrimSpace(review.Target),
		Decision:      strings.TrimSpace(review.Decision),
		Reason:        strings.TrimSpace(review.Reason),
	}
}

func browserRuntimeBuildSessionBinding(ctx context.Context, registry *BrowserSessionRegistry, sessionRunRegistry BrowserSessionRunRegistry, sessionStateRegistry agentxbrowserruntime.SharedSessionBrowserStateRegistry, selectedRoute *browserRuntimeRouteDescriptor, routes []browserRuntimeSessionRoute, evaluation *agentxbrowserruntime.SharedSessionBrowserBindingEvaluation) *browserRuntimeSessionBinding {
	sessionKey := ToolSessionIDFromContext(ctx)
	if sessionKey == "" {
		return nil
	}
	if selectedRoute == nil && evaluation == nil {
		return nil
	}
	binding := &browserRuntimeSessionBinding{
		SessionKey: strings.TrimSpace(sessionKey),
	}
	selectedInfo := browserRuntimeSelectedInfoForBinding(selectedRoute, evaluation)
	if strings.EqualFold(strings.TrimSpace(selectedInfo.Target), "node") || strings.EqualFold(strings.TrimSpace(selectedInfo.Backend), "proxy") {
		binding.PropagatedToProxy = true
	}
	if evaluation == nil {
		var sharedRoutes []agentxbrowserruntime.SharedSessionBrowserRouteSnapshot
		if routes != nil {
			sharedRoutes = browserRuntimeSharedSessionRouteSnapshots(routes)
		}
		sharedEvaluation := agentxbrowserruntime.EvaluateSharedSessionBrowserBindingForScope(
			sessionKey,
			selectedInfo,
			browserRuntimeSessionRouteFilter(selectedRoute),
			sharedRoutes,
			registry,
			sessionRunRegistry,
			sessionStateRegistry,
			browserRuntimeReconnectWatchdogWindow,
		)
		evaluation = &sharedEvaluation
	}
	browserRuntimeApplySharedBindingEvaluation(binding, *evaluation)
	return binding
}

func browserRuntimeSelectedInfoForBinding(
	selectedRoute *browserRuntimeRouteDescriptor,
	evaluation *agentxbrowserruntime.SharedSessionBrowserBindingEvaluation,
) agentxbrowserruntime.BrowserRuntimeInfo {
	if selectedRoute != nil {
		return agentxbrowserruntime.BrowserRuntimeInfo{
			Backend: strings.TrimSpace(selectedRoute.Backend),
			Profile: strings.TrimSpace(selectedRoute.Profile),
			Target:  strings.TrimSpace(selectedRoute.RuntimeTarget),
		}
	}
	if evaluation == nil {
		return agentxbrowserruntime.BrowserRuntimeInfo{}
	}
	if selection := evaluation.Snapshot.SelectedProfileSelection; selection != nil {
		info := agentxbrowserruntime.BrowserRuntimeInfo{
			Backend: strings.TrimSpace(selection.Backend),
			Profile: strings.TrimSpace(selection.Profile),
			Target:  strings.TrimSpace(selection.RuntimeTarget),
		}
		if info != (agentxbrowserruntime.BrowserRuntimeInfo{}) {
			return info
		}
	}
	if selection := evaluation.Snapshot.SelectedTargetSelection; selection != nil {
		info := agentxbrowserruntime.BrowserRuntimeInfo{
			Backend: strings.TrimSpace(selection.Backend),
			Profile: strings.TrimSpace(selection.Profile),
			Target:  strings.TrimSpace(selection.RuntimeTarget),
		}
		if info != (agentxbrowserruntime.BrowserRuntimeInfo{}) {
			return info
		}
	}
	for _, state := range evaluation.Snapshot.Profiles {
		info := agentxbrowserruntime.BrowserRuntimeInfo{
			Backend: strings.TrimSpace(state.Backend),
			Profile: strings.TrimSpace(state.Profile),
			Target:  strings.TrimSpace(state.RuntimeTarget),
		}
		if info != (agentxbrowserruntime.BrowserRuntimeInfo{}) {
			return info
		}
	}
	return agentxbrowserruntime.BrowserRuntimeInfo{}
}

func browserRuntimeRouteResolutionDefaultInfo(payload browserRuntimePayload, fallback BrowserRuntimeInfo) agentxbrowserruntime.BrowserRuntimeInfo {
	fallback = normalizeBrowserRuntimeInfo(fallback)
	if fallback != (BrowserRuntimeInfo{}) {
		return agentxbrowserruntime.BrowserRuntimeInfo{
			Backend: fallback.Backend,
			Profile: fallback.Profile,
			Target:  fallback.Target,
		}
	}
	return agentxbrowserruntime.BrowserRuntimeInfo{
		Backend: strings.TrimSpace(payload.DefaultRoute.Backend),
		Profile: strings.TrimSpace(payload.DefaultRoute.Profile),
		Target:  strings.TrimSpace(payload.DefaultRoute.RuntimeTarget),
	}
}

func browserRuntimeRouteResolutionPtr(
	payload browserRuntimePayload,
	defaultRoute BrowserRuntimeInfo,
	hiddenImplicitHostDefaultBase bool,
) *browserRuntimeRouteResolution {
	if payload.SelectedRoute == nil {
		return nil
	}
	if browserRuntimeShouldHideImplicitLegacyHostRouteResolution(payload, hiddenImplicitHostDefaultBase) {
		return nil
	}
	resolution, ok := agentxbrowserruntime.ResolveSharedSessionBrowserRouteResolution(
		strings.TrimSpace(payload.RequestedProfile),
		strings.TrimSpace(payload.RequestedRuntimeTarget),
		browserRuntimeRouteResolutionDefaultInfo(payload, defaultRoute),
		agentxbrowserruntime.BrowserRuntimeInfo{
			Backend: strings.TrimSpace(payload.SelectedRoute.Backend),
			Profile: strings.TrimSpace(payload.SelectedRoute.Profile),
			Target:  strings.TrimSpace(payload.SelectedRoute.RuntimeTarget),
		},
		browserRuntimeSharedProfileSelection(payload.SessionProfileSelection),
		browserRuntimeSharedTargetSelection(payload.SessionTargetSelection),
	)
	if !ok {
		return nil
	}
	return &browserRuntimeRouteResolution{
		ProfileSource:       resolution.ProfileSource,
		RuntimeTargetSource: resolution.RuntimeTargetSource,
		TargetSource:        resolution.TargetSource,
	}
}

func browserRuntimeShouldHideImplicitLegacyHostRouteResolution(payload browserRuntimePayload, hiddenImplicitHostDefaultBase bool) bool {
	if strings.TrimSpace(payload.RequestedProfile) != "" || strings.TrimSpace(payload.RequestedRuntimeTarget) != "" {
		return false
	}
	if payload.SessionProfileSelection != nil || payload.SessionTargetSelection != nil {
		return false
	}
	if !hiddenImplicitHostDefaultBase {
		return false
	}
	if payload.SelectedRoute == nil {
		return false
	}
	return BrowserSubstratePosture(payload.SelectedRoute.Backend, payload.SelectedRoute.RuntimeTarget) == BrowserSubstrateLegacySystemHost
}

func browserRuntimeSharedSessionRuns(ctx context.Context, registry BrowserSessionRunRegistry) []agentxbrowserruntime.SharedSessionRunInfo {
	sessionKey := ToolSessionIDFromContext(ctx)
	if sessionKey == "" || registry == nil {
		return nil
	}
	snapshot := registry.SnapshotSessionRuns(sessionKey)
	if len(snapshot) == 0 {
		return nil
	}
	out := make([]agentxbrowserruntime.SharedSessionRunInfo, 0, len(snapshot))
	for _, item := range snapshot {
		out = append(out, agentxbrowserruntime.SharedSessionRunInfo{
			RunID:    strings.TrimSpace(item.RunID),
			NodeID:   strings.TrimSpace(item.NodeID),
			Status:   strings.TrimSpace(item.Status),
			Provider: strings.TrimSpace(item.Provider),
			Action:   strings.TrimSpace(item.Action),
		})
	}
	return out
}

func browserRuntimeSessionRuns(ctx context.Context, registry BrowserSessionRunRegistry) []browserRuntimeSessionRun {
	return browserRuntimeSessionRunsFromShared(browserRuntimeSharedSessionRuns(ctx, registry))
}

func browserRuntimeSessionRunsFromShared(items []agentxbrowserruntime.SharedSessionRunInfo) []browserRuntimeSessionRun {
	if len(items) == 0 {
		return nil
	}
	out := make([]browserRuntimeSessionRun, 0, len(items))
	for _, item := range items {
		out = append(out, browserRuntimeSessionRun{
			RunID:    strings.TrimSpace(item.RunID),
			NodeID:   strings.TrimSpace(item.NodeID),
			Status:   strings.TrimSpace(item.Status),
			Provider: strings.TrimSpace(item.Provider),
			Action:   strings.TrimSpace(item.Action),
		})
	}
	return out
}

func browserRuntimeSharedSessionRunsFromBinding(items []browserRuntimeSessionRun) []agentxbrowserruntime.SharedSessionRunInfo {
	if len(items) == 0 {
		return nil
	}
	out := make([]agentxbrowserruntime.SharedSessionRunInfo, 0, len(items))
	for _, item := range items {
		out = append(out, agentxbrowserruntime.SharedSessionRunInfo{
			RunID:    strings.TrimSpace(item.RunID),
			NodeID:   strings.TrimSpace(item.NodeID),
			Status:   strings.TrimSpace(item.Status),
			Provider: strings.TrimSpace(item.Provider),
			Action:   strings.TrimSpace(item.Action),
		})
	}
	return out
}

func browserRuntimeSessionProfiles(ctx context.Context, registry agentxbrowserruntime.SharedSessionBrowserStateRegistry, selectedRoute *browserRuntimeRouteDescriptor) []browserRuntimeProfileState {
	sessionKey := ToolSessionIDFromContext(ctx)
	if sessionKey == "" || registry == nil {
		return nil
	}
	selectedInfo := BrowserRuntimeInfo{}
	requestedProfile := ""
	if selectedRoute != nil {
		selectedInfo = BrowserRuntimeInfo{
			Backend: strings.TrimSpace(selectedRoute.Backend),
			Profile: strings.TrimSpace(selectedRoute.Profile),
			Target:  strings.TrimSpace(selectedRoute.RuntimeTarget),
		}
		requestedProfile = strings.TrimSpace(selectedRoute.Profile)
	}
	return browserRuntimeProfileStatesFromProjected(
		agentxbrowserruntime.SnapshotSharedSessionBrowserProjectedProfilesForScope(
			registry,
			sessionKey,
			selectedInfo,
			requestedProfile,
		),
	)
}

func browserRuntimeDegradedProfilesFromRouteSnapshot(ctx context.Context, registry agentxbrowserruntime.SharedSessionBrowserStateRegistry, route BrowserRuntimeInfo, requestedProfile string) []browserRuntimeProfileState {
	route = normalizeBrowserRuntimeInfo(route)
	return browserRuntimeSessionProfiles(ctx, registry, browserRuntimeRouteDescriptorPtr(BrowserRuntimeInfo{
		Backend: route.Backend,
		Profile: strings.TrimSpace(requestedProfile),
		Target:  route.Target,
	}))
}

func browserRuntimeDegradedProfileStatusFromRouteSnapshot(ctx context.Context, registry agentxbrowserruntime.SharedSessionBrowserStateRegistry, route BrowserRuntimeInfo, requestedProfile string) *browserRuntimeProfileState {
	route = normalizeBrowserRuntimeInfo(route)
	profile := strings.TrimSpace(firstNonEmpty(requestedProfile, route.Profile))
	if profile == "" {
		return nil
	}
	for _, item := range browserRuntimeDegradedProfilesFromRouteSnapshot(ctx, registry, route, profile) {
		if strings.EqualFold(strings.TrimSpace(item.Profile), profile) {
			state := item
			return &state
		}
	}
	return nil
}

func browserRuntimeDegradedProfilesFromSessionRouteSnapshot(
	ctx context.Context,
	registry *BrowserSessionRegistry,
	route BrowserRuntimeInfo,
	requestedProfile string,
) []browserRuntimeProfileState {
	return browserRuntimeProfileStatesFromProjected(
		browserRuntimeDegradedProjectedProfilesFromSessionSnapshot(
			ctx,
			registry,
			route,
			requestedProfile,
			nil,
		),
	)
}

func browserRuntimeDegradedProfileStatusFromSessionRouteSnapshot(
	ctx context.Context,
	registry *BrowserSessionRegistry,
	route BrowserRuntimeInfo,
	requestedProfile string,
) *browserRuntimeProfileState {
	return browserRuntimeProfileStatusFromStates(
		browserRuntimeDegradedProfilesFromSessionRouteSnapshot(ctx, registry, route, requestedProfile),
		strings.TrimSpace(firstNonEmpty(requestedProfile, route.Profile)),
	)
}

func browserRuntimeDegradedSessionViewFromRouteSnapshot(
	ctx context.Context,
	registry *BrowserSessionRegistry,
	runRegistry agentxbrowserruntime.SharedSessionRunRegistry,
	stateRegistry agentxbrowserruntime.SharedSessionBrowserStateRegistry,
	route BrowserRuntimeInfo,
	requestedProfile string,
) agentxbrowserruntime.SharedSessionBrowserSessionViewSnapshot {
	sessionID := ToolSessionIDFromContext(ctx)
	if sessionID == "" {
		return agentxbrowserruntime.SharedSessionBrowserSessionViewSnapshot{}
	}
	route = normalizeBrowserRuntimeInfo(route)
	selectedInfo := BrowserRuntimeInfo{
		Backend: route.Backend,
		Target:  route.Target,
	}
	routeFilter := BrowserSessionRoute{
		Backend: route.Backend,
		Target:  route.Target,
	}
	viewRequestedProfile := ""
	if strings.TrimSpace(requestedProfile) != "" {
		selectedInfo.Profile = strings.TrimSpace(requestedProfile)
		routeFilter.Profile = strings.TrimSpace(requestedProfile)
		viewRequestedProfile = strings.TrimSpace(requestedProfile)
	}
	return agentxbrowserruntime.SnapshotSharedSessionBrowserSessionView(
		sessionID,
		selectedInfo,
		viewRequestedProfile,
		routeFilter,
		registry,
		runRegistry,
		stateRegistry,
	)
}

func browserRuntimeDegradedProjectedProfilesFromSessionSnapshot(
	ctx context.Context,
	registry *BrowserSessionRegistry,
	route BrowserRuntimeInfo,
	requestedProfile string,
	selection *browserRuntimeSessionProfileSelection,
) []agentxbrowserruntime.SharedSessionBrowserProjectedProfileState {
	return browserRuntimeDegradedProjectedProfilesFromSessionView(
		route,
		requestedProfile,
		browserRuntimeDegradedSessionViewFromRouteSnapshot(
			ctx,
			registry,
			nil,
			nil,
			route,
			requestedProfile,
		),
		selection,
	)
}

func browserRuntimeProfileStatesFromProjected(snapshot []agentxbrowserruntime.SharedSessionBrowserProjectedProfileState) []browserRuntimeProfileState {
	if len(snapshot) == 0 {
		return nil
	}
	out := make([]browserRuntimeProfileState, 0, len(snapshot))
	for _, item := range snapshot {
		state := item.State
		out = append(out, browserRuntimeProfileState{
			Backend:       strings.TrimSpace(state.Backend),
			Profile:       strings.TrimSpace(state.Profile),
			RuntimeTarget: strings.TrimSpace(state.RuntimeTarget),
			BrowserApp:    strings.TrimSpace(state.BrowserApp),
			Status:        strings.TrimSpace(state.Status),
			Running:       state.Running,
			Connected:     state.Connected,
			Selected:      item.Selected,
			Note:          strings.TrimSpace(state.Note),
			ObservedAt:    state.ObservedAt,
			StatusSince:   state.StatusSince,
		})
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func browserRuntimeDegradedProjectedProfilesFromSessionView(
	route BrowserRuntimeInfo,
	requestedProfile string,
	sessionView agentxbrowserruntime.SharedSessionBrowserSessionViewSnapshot,
	selection *browserRuntimeSessionProfileSelection,
) []agentxbrowserruntime.SharedSessionBrowserProjectedProfileState {
	return agentxbrowserruntime.ProjectSharedSessionBrowserFallbackProfilesFromRouteSnapshots(
		route,
		requestedProfile,
		sessionView.Routes,
		browserRuntimeSharedProfileSelection(selection),
	)
}

func browserRuntimeProfileStatusFromStates(states []browserRuntimeProfileState, requestedProfile string) *browserRuntimeProfileState {
	if len(states) == 0 {
		return nil
	}
	requestedProfile = strings.TrimSpace(requestedProfile)
	if requestedProfile != "" {
		for i := range states {
			if strings.EqualFold(strings.TrimSpace(states[i].Profile), requestedProfile) {
				state := states[i]
				return &state
			}
		}
	}
	if len(states) == 1 {
		state := states[0]
		return &state
	}
	return nil
}

func browserRuntimeCanonicalBackend(backend string) string {
	backend = strings.ToLower(strings.TrimSpace(backend))
	switch {
	case strings.HasPrefix(backend, "proxy-"):
		return "proxy"
	case strings.HasPrefix(backend, "system-"):
		return "system"
	case strings.HasPrefix(backend, "sandbox-"):
		return "sandbox"
	case strings.HasPrefix(backend, "custom-"):
		return "custom"
	default:
		return backend
	}
}

func browserRuntimeBackendMatches(left string, right string) bool {
	left = browserRuntimeCanonicalBackend(left)
	right = browserRuntimeCanonicalBackend(right)
	if left == "" || right == "" {
		return left == right
	}
	return strings.EqualFold(left, right)
}

func browserRuntimeMarkCurrentProfileSelected(payload *browserRuntimePayload) {
	if payload == nil || payload.ProfileStatus == nil {
		return
	}
	runtimeTarget := strings.TrimSpace(payload.RequestedRuntimeTarget)
	if runtimeTarget == "" && payload.SelectedRoute != nil {
		runtimeTarget = strings.TrimSpace(payload.SelectedRoute.RuntimeTarget)
	}
	payload.ProfileStatus.Selected = agentxbrowserruntime.SharedSessionBrowserProfileStateSelected(
		browserRuntimeSharedProfileSelection(payload.SessionProfileSelection),
		runtimeTarget,
		agentxbrowserruntime.SharedSessionBrowserProfileState{
			Backend:       strings.TrimSpace(payload.ProfileStatus.Backend),
			Profile:       strings.TrimSpace(payload.ProfileStatus.Profile),
			RuntimeTarget: strings.TrimSpace(payload.ProfileStatus.RuntimeTarget),
		},
	)
}

func browserRuntimeSharedSessionRouteSnapshots(routes []browserRuntimeSessionRoute) []agentxbrowserruntime.SharedSessionBrowserRouteSnapshot {
	if len(routes) == 0 {
		return nil
	}
	out := make([]agentxbrowserruntime.SharedSessionBrowserRouteSnapshot, 0, len(routes))
	for _, route := range routes {
		targets := make([]agentxbrowserruntime.SharedSessionBrowserRouteTarget, 0, len(route.Targets))
		for _, target := range route.Targets {
			targets = append(targets, agentxbrowserruntime.SharedSessionBrowserRouteTarget{
				ID:            strings.TrimSpace(target.ID),
				TabIndex:      target.TabIndex,
				URL:           strings.TrimSpace(target.URL),
				Title:         strings.TrimSpace(target.Title),
				BrowserApp:    strings.TrimSpace(target.BrowserApp),
				Backend:       strings.TrimSpace(target.Backend),
				Profile:       strings.TrimSpace(target.Profile),
				RuntimeTarget: strings.TrimSpace(target.RuntimeTarget),
				Current:       target.Current,
			})
		}
		var pending *agentxbrowserruntime.BrowserSessionTargetReview
		if route.PendingTargetReview != nil {
			pending = &agentxbrowserruntime.BrowserSessionTargetReview{
				ID:         strings.TrimSpace(route.PendingTargetReview.ID),
				TabIndex:   route.PendingTargetReview.TabIndex,
				URL:        strings.TrimSpace(route.PendingTargetReview.URL),
				Title:      strings.TrimSpace(route.PendingTargetReview.Title),
				BrowserApp: strings.TrimSpace(route.PendingTargetReview.BrowserApp),
				Backend:    strings.TrimSpace(route.PendingTargetReview.Backend),
				Profile:    strings.TrimSpace(route.PendingTargetReview.Profile),
				Target:     strings.TrimSpace(route.PendingTargetReview.RuntimeTarget),
				Decision:   strings.TrimSpace(route.PendingTargetReview.Decision),
				Reason:     strings.TrimSpace(route.PendingTargetReview.Reason),
			}
		}
		out = append(out, agentxbrowserruntime.SharedSessionBrowserRouteSnapshot{
			Backend:                  strings.TrimSpace(route.Backend),
			Profile:                  strings.TrimSpace(route.Profile),
			RuntimeTarget:            strings.TrimSpace(route.RuntimeTarget),
			BrowserApp:               strings.TrimSpace(route.BrowserApp),
			CurrentTargetID:          strings.TrimSpace(route.CurrentTargetID),
			CurrentTargetSource:      strings.TrimSpace(route.CurrentTargetSource),
			PendingTargetReview:      pending,
			PendingTargetReviewCount: route.PendingTargetReviewCount,
			FollowPolicyState:        strings.TrimSpace(route.FollowPolicyState),
			FollowPolicyReason:       strings.TrimSpace(route.FollowPolicyReason),
			PopupPolicyState:         strings.TrimSpace(route.PopupPolicyState),
			PopupPolicyReason:        strings.TrimSpace(route.PopupPolicyReason),
			Targets:                  targets,
		})
	}
	return out
}

func browserRuntimeSessionCoordination(ctx context.Context, registry agentxbrowserruntime.SharedSessionBrowserStateRegistry, binding browserRuntimeSessionBinding, selectedRoute *browserRuntimeRouteDescriptor, routes []browserRuntimeSessionRoute) *browserRuntimeCoordination {
	selectedInfo := BrowserRuntimeInfo{}
	requestedProfile := ""
	if selectedRoute != nil {
		selectedInfo.Backend = strings.TrimSpace(selectedRoute.Backend)
		selectedInfo.Target = strings.TrimSpace(selectedRoute.RuntimeTarget)
		requestedProfile = strings.TrimSpace(firstNonEmpty(selectedRoute.Profile, inputProfileFromBinding(&binding)))
	}
	sessionID, _ := browserRuntimeSessionStateRegistrySessionID(ctx, registry)
	evaluation := agentxbrowserruntime.EvaluateSharedSessionBrowserCoordinationEvaluationForScope(
		registry,
		sessionID,
		selectedInfo,
		requestedProfile,
		browserRuntimeSessionHealthInputFromBinding(binding),
		browserRuntimeSessionCoordinationInputFromBinding(binding),
		browserRuntimeRouteCoordinationInputs(routes),
		browserRuntimeReconnectWatchdogWindow,
		binding.BlockedAutoFollowRouteCount > 0,
	)
	plan := evaluation.Plan
	if len(plan.RecommendedBrowserActions) == 0 && len(plan.RecommendedNodeActions) == 0 && plan.State == "idle" {
		return &browserRuntimeCoordination{State: plan.State}
	}
	return browserRuntimeCoordinationFromSharedEvaluation(evaluation)
}

func browserRuntimeCoordinationFromSharedEvaluation(evaluation agentxbrowserruntime.SharedSessionBrowserCoordinationEvaluation) *browserRuntimeCoordination {
	plan := evaluation.Plan
	if len(plan.RecommendedBrowserActions) == 0 && len(plan.RecommendedNodeActions) == 0 && plan.State == "idle" {
		return &browserRuntimeCoordination{State: plan.State}
	}
	return &browserRuntimeCoordination{
		State:                     plan.State,
		BrowserOnNode:             plan.BrowserOnNode,
		HasActiveNodeRun:          plan.HasActiveNodeRun,
		HasRunningBrowserProfile:  plan.HasRunningBrowserProfile,
		SyncBrowserAction:         strings.TrimSpace(plan.SyncAction),
		PrepareBrowserAction:      strings.TrimSpace(plan.PrepareAction),
		RestartBrowserAction:      strings.TrimSpace(evaluation.RestartAction),
		TeardownBrowserAction:     strings.TrimSpace(plan.TeardownAction),
		PrimaryBrowserAction:      strings.TrimSpace(evaluation.Guidance.PrimaryAction),
		PrimaryNodeAction:         strings.TrimSpace(plan.PrimaryNodeAction),
		NextStep:                  strings.TrimSpace(evaluation.Guidance.NextStep),
		RecommendedBrowserActions: evaluation.Guidance.RecommendedActions,
		RecommendedNodeActions:    plan.RecommendedNodeActions,
	}
}

func browserRuntimeSharedBindingEvaluation(binding browserRuntimeSessionBinding, routes []browserRuntimeSessionRoute) agentxbrowserruntime.SharedSessionBrowserBindingEvaluation {
	if binding.HasSharedEvaluation {
		evaluation := binding.SharedEvaluation
		if routes != nil {
			evaluation.Routes = browserRuntimeSharedSessionRouteSnapshots(routes)
			evaluation.Handoff = nil
		}
		return browserRuntimeSharedBindingEvaluationWithHandoff(evaluation)
	}
	return browserRuntimeSharedBindingEvaluationWithHandoff(agentxbrowserruntime.SharedSessionBrowserBindingEvaluation{
		Routes: browserRuntimeSharedSessionRouteSnapshots(routes),
		Snapshot: agentxbrowserruntime.SharedSessionBrowserBindingSnapshot{
			CurrentTargetID:          strings.TrimSpace(binding.CurrentTargetID),
			SelectedProfileSelection: browserRuntimeSharedProfileSelectionFromBinding(binding),
			SelectedTargetSelection:  browserRuntimeSharedTargetSelectionFromBinding(binding),
			Runs:                     browserRuntimeSharedSessionRunsFromBinding(binding.NodeRuns),
			Profiles:                 browserRuntimeSharedSessionProfileStates(binding.BrowserProfiles),
			Summary: agentxbrowserruntime.SharedSessionBrowserBindingSummary{
				CurrentTargetID:             strings.TrimSpace(binding.CurrentTargetID),
				RouteTargetCount:            binding.RouteTargetCount,
				PendingTargetReviewCount:    binding.PendingTargetReviewCount,
				BlockedAutoFollowRouteCount: binding.BlockedAutoFollowRouteCount,
				PopupStormRouteCount:        binding.PopupStormRouteCount,
				NodeRunCount:                binding.NodeRunCount,
				ActiveNodeRunID:             strings.TrimSpace(binding.ActiveNodeRunID),
				NodeRunStatusCounts:         cloneStringIntMap(binding.NodeRunStatusCounts),
				BrowserProfileCount:         binding.BrowserProfileCount,
				ActiveBrowserProfile:        strings.TrimSpace(binding.ActiveBrowserProfile),
				BrowserProfileStatusCounts:  cloneStringIntMap(binding.BrowserProfileStatusCounts),
			},
		},
		Health: agentxbrowserruntime.SharedSessionBrowserHealthEvaluation{
			Summary: browserRuntimeSharedSessionHealthSummary(binding),
		},
		Handoff:       agentxbrowserruntime.CloneSharedSessionBrowserSessionHandoffSummary(binding.SessionHandoff),
		ReferenceTime: binding.ReferenceTime,
	})
}

func browserRuntimeSharedBindingEvaluationWithHandoff(
	evaluation agentxbrowserruntime.SharedSessionBrowserBindingEvaluation,
) agentxbrowserruntime.SharedSessionBrowserBindingEvaluation {
	if evaluation.Handoff != nil {
		return evaluation
	}
	targetCount := evaluation.Snapshot.Summary.RouteTargetCount
	if len(evaluation.Routes) > 0 {
		targetCount = agentxbrowserruntime.SharedSessionBrowserRouteTargetCount(evaluation.Routes)
	}
	evaluation.Handoff = agentxbrowserruntime.BuildSharedSessionBrowserSessionHandoffSummary(
		agentxbrowserruntime.SharedSessionBrowserSessionHandoffRequest{
			Routes:          evaluation.Routes,
			Runs:            evaluation.Snapshot.Runs,
			Profiles:        agentxbrowserruntime.ProjectSharedSessionBrowserProfileSnapshot(evaluation.Snapshot.Profiles, evaluation.Snapshot.SelectedProfileSelection),
			TargetCount:     targetCount,
			SelectedProfile: evaluation.Snapshot.SelectedProfileSelection,
			SelectedTarget:  evaluation.Snapshot.SelectedTargetSelection,
			Health:          evaluation.Health.Summary,
		},
	)
	return evaluation
}

func browserRuntimeApplySharedBindingEvaluation(binding *browserRuntimeSessionBinding, evaluation agentxbrowserruntime.SharedSessionBrowserBindingEvaluation) {
	if binding == nil {
		return
	}
	binding.SharedEvaluation = evaluation
	binding.HasSharedEvaluation = true
	sharedSnapshot := evaluation.Snapshot
	if selection := browserRuntimeSelectionPtrValue(sharedSnapshot.SelectedProfileSelection); selection != nil {
		binding.SelectedBrowserBackend = strings.TrimSpace(selection.Backend)
		binding.SelectedBrowserApp = strings.TrimSpace(selection.BrowserApp)
		binding.SelectedBrowserProfile = strings.TrimSpace(selection.Profile)
		binding.SelectedBrowserTarget = strings.TrimSpace(selection.RuntimeTarget)
		binding.SelectedBrowserProfileSource = strings.TrimSpace(selection.Source)
	} else {
		binding.SelectedBrowserBackend = ""
		binding.SelectedBrowserApp = ""
		binding.SelectedBrowserProfile = ""
		binding.SelectedBrowserTarget = ""
		binding.SelectedBrowserProfileSource = ""
	}
	if selection := browserRuntimeSessionTargetSelectionPtrFromShared(sharedSnapshot.SelectedTargetSelection); selection != nil {
		if binding.SelectedBrowserBackend == "" {
			binding.SelectedBrowserBackend = strings.TrimSpace(selection.Backend)
		}
		if binding.SelectedBrowserApp == "" {
			binding.SelectedBrowserApp = strings.TrimSpace(selection.BrowserApp)
		}
		binding.SelectedBrowserTargetID = strings.TrimSpace(selection.ID)
		binding.SelectedBrowserTabIndex = selection.TabIndex
		binding.SelectedBrowserTargetSource = strings.TrimSpace(selection.Source)
	} else {
		binding.SelectedBrowserTargetID = ""
		binding.SelectedBrowserTabIndex = 0
		binding.SelectedBrowserTargetSource = ""
	}
	binding.NodeRuns = browserRuntimeSessionRunsFromShared(sharedSnapshot.Runs)
	binding.BrowserProfiles = browserRuntimeProfileStatesFromProjected(
		agentxbrowserruntime.ProjectSharedSessionBrowserProfileSnapshot(
			sharedSnapshot.Profiles,
			sharedSnapshot.SelectedProfileSelection,
		),
	)
	sharedSummary := sharedSnapshot.Summary
	binding.RouteTargetCount = sharedSummary.RouteTargetCount
	binding.PendingTargetReviewCount = sharedSummary.PendingTargetReviewCount
	binding.BlockedAutoFollowRouteCount = sharedSummary.BlockedAutoFollowRouteCount
	binding.PopupStormRouteCount = sharedSummary.PopupStormRouteCount
	binding.NodeRunCount = sharedSummary.NodeRunCount
	binding.ActiveNodeRunID = strings.TrimSpace(sharedSummary.ActiveNodeRunID)
	binding.NodeRunStatusCounts = cloneStringIntMap(sharedSummary.NodeRunStatusCounts)
	binding.BrowserProfileCount = sharedSummary.BrowserProfileCount
	binding.ActiveBrowserProfile = strings.TrimSpace(sharedSummary.ActiveBrowserProfile)
	binding.BrowserProfileStatusCounts = cloneStringIntMap(sharedSummary.BrowserProfileStatusCounts)
	binding.CurrentTargetID = strings.TrimSpace(sharedSnapshot.CurrentTargetID)
	binding.ReferenceTime = evaluation.ReferenceTime
	browserRuntimeApplySessionHealthEvaluation(binding, browserRuntimeSessionHealthEvaluationFromShared(evaluation.Health))
	if evaluation.Handoff != nil {
		binding.SessionHandoff = agentxbrowserruntime.CloneSharedSessionBrowserSessionHandoffSummary(evaluation.Handoff)
	} else {
		binding.SessionHandoff = agentxbrowserruntime.BuildSharedSessionBrowserSessionHandoffSummary(
			agentxbrowserruntime.SharedSessionBrowserSessionHandoffRequest{
				Routes:          evaluation.Routes,
				Runs:            sharedSnapshot.Runs,
				Profiles:        agentxbrowserruntime.ProjectSharedSessionBrowserProfileSnapshot(sharedSnapshot.Profiles, sharedSnapshot.SelectedProfileSelection),
				TargetCount:     sharedSummary.RouteTargetCount,
				SelectedProfile: sharedSnapshot.SelectedProfileSelection,
				SelectedTarget:  sharedSnapshot.SelectedTargetSelection,
				Health:          evaluation.Health.Summary,
			},
		)
	}
	binding.Coordination = browserRuntimeCoordinationFromSharedEvaluation(evaluation.Coordination)
}

func browserRuntimeSharedSessionHealthSummary(binding browserRuntimeSessionBinding) *agentxbrowserruntime.SharedSessionBrowserHealthSummary {
	if strings.TrimSpace(binding.SessionHealthState) == "" &&
		strings.TrimSpace(binding.SessionHealthReason) == "" &&
		strings.TrimSpace(binding.SessionHealthRecoveryAction) == "" &&
		strings.TrimSpace(binding.SessionHealthReconnectHint) == "" &&
		binding.SessionHealthDisconnectCount == 0 &&
		binding.SessionHealthDisconnectBurstCount == 0 &&
		binding.SessionHealthDisconnectBurstWindowMs == 0 &&
		binding.SessionHealthCooldownRemainingMs == 0 &&
		binding.SessionHealthRetryBackoffRemainingMs == 0 &&
		binding.SessionHealthRestartAttemptCount == 0 &&
		binding.SessionHealthRestartFailureCount == 0 &&
		binding.SessionHealthLastDisconnectUnixMilli == 0 &&
		binding.SessionHealthLastReconnectUnixMilli == 0 &&
		binding.SessionHealthLastRestartAttemptUnixMilli == 0 &&
		strings.TrimSpace(binding.SessionHealthLastRestartResult) == "" &&
		strings.TrimSpace(binding.SessionHealthLastRestartError) == "" &&
		binding.SessionHealthRecommendedBackoffMs == 0 &&
		strings.TrimSpace(binding.SessionHealthResolverBlockedBy) == "" &&
		strings.TrimSpace(binding.SessionHealthResolverAmbiguityClass) == "" &&
		strings.TrimSpace(binding.SessionHealthResolverCandidateKind) == "" &&
		strings.TrimSpace(binding.SessionHealthResolverStrength) == "" &&
		strings.TrimSpace(binding.SessionHealthResolverRetryDisposition) == "" &&
		strings.TrimSpace(binding.SessionHealthResolverManualRetryHint) == "" &&
		strings.TrimSpace(binding.SessionHealthResolverNextStepAlias) == "" &&
		len(binding.SessionHealthResolverSpecificityFields) == 0 {
		return nil
	}
	return &agentxbrowserruntime.SharedSessionBrowserHealthSummary{
		State:                       strings.TrimSpace(binding.SessionHealthState),
		Reason:                      strings.TrimSpace(binding.SessionHealthReason),
		RecoveryAction:              strings.TrimSpace(binding.SessionHealthRecoveryAction),
		ReconnectHint:               strings.TrimSpace(binding.SessionHealthReconnectHint),
		DisconnectCount:             binding.SessionHealthDisconnectCount,
		DisconnectBurstCount:        binding.SessionHealthDisconnectBurstCount,
		DisconnectBurstWindowMs:     binding.SessionHealthDisconnectBurstWindowMs,
		CooldownRemainingMs:         binding.SessionHealthCooldownRemainingMs,
		RetryBackoffRemainingMs:     binding.SessionHealthRetryBackoffRemainingMs,
		RestartAttemptCount:         binding.SessionHealthRestartAttemptCount,
		RestartFailureCount:         binding.SessionHealthRestartFailureCount,
		LastDisconnectUnixMilli:     binding.SessionHealthLastDisconnectUnixMilli,
		LastReconnectUnixMilli:      binding.SessionHealthLastReconnectUnixMilli,
		LastRestartAttemptUnixMilli: binding.SessionHealthLastRestartAttemptUnixMilli,
		LastRestartResult:           strings.TrimSpace(binding.SessionHealthLastRestartResult),
		LastRestartError:            strings.TrimSpace(binding.SessionHealthLastRestartError),
		RecommendedBackoffMs:        binding.SessionHealthRecommendedBackoffMs,
		ResolverBlockedBy:           strings.TrimSpace(binding.SessionHealthResolverBlockedBy),
		AmbiguityClass:              strings.TrimSpace(binding.SessionHealthResolverAmbiguityClass),
		CandidateKind:               strings.TrimSpace(binding.SessionHealthResolverCandidateKind),
		CandidateStrength:           strings.TrimSpace(binding.SessionHealthResolverStrength),
		RetryDisposition:            strings.TrimSpace(binding.SessionHealthResolverRetryDisposition),
		ManualRetryHint:             strings.TrimSpace(binding.SessionHealthResolverManualRetryHint),
		NextStepAlias:               strings.TrimSpace(binding.SessionHealthResolverNextStepAlias),
		SpecificityFields:           append([]string(nil), binding.SessionHealthResolverSpecificityFields...),
	}
}

func cloneStringIntMap(in map[string]int) map[string]int {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]int, len(in))
	for key, value := range in {
		out[strings.TrimSpace(key)] = value
	}
	return out
}

func browserRuntimeSessionCoordinationInputFromBinding(binding browserRuntimeSessionBinding) agentxbrowserruntime.SharedSessionBrowserCoordinationInput {
	if binding.HasSharedEvaluation {
		return agentxbrowserruntime.BuildSharedSessionBrowserCoordinationInputFromBindingEvaluation(binding.SharedEvaluation)
	}
	return agentxbrowserruntime.BuildSharedSessionBrowserCoordinationInput(
		binding.ActiveNodeRunID,
		binding.RouteTargetCount,
		browserRuntimeSharedProfileSelectionFromBinding(binding),
		browserRuntimeSharedTargetSelectionFromBinding(binding),
		browserRuntimeSharedSessionProfileStates(binding.BrowserProfiles),
	)
}

func browserRuntimeRouteCoordinationInputs(routes []browserRuntimeSessionRoute) []agentxbrowserruntime.SharedSessionBrowserRouteCoordinationInput {
	return agentxbrowserruntime.SharedSessionBrowserRouteCoordinationInputs(browserRuntimeSharedSessionRouteSnapshots(routes))
}

func browserRuntimeSupportedToolNames(ctx browserRegistrationContext, capabilities BrowserCapabilities) []string {
	tools := capabilities.SupportedToolNames()
	out := make([]string, 0, len(tools))
	for _, name := range tools {
		if browserRuntimeRegisteredOrEnabledTool(ctx, name) {
			out = append(out, name)
		}
	}
	return out
}

func browserRuntimeSupportedActKinds(ctx browserRegistrationContext, capabilities BrowserCapabilities) []string {
	if !browserRuntimeRegisteredOrEnabledTool(ctx, "browser_act") {
		return nil
	}
	return capabilities.SupportedActKinds()
}

func browserRuntimeArtifactTools(ctx browserRegistrationContext, capabilities BrowserCapabilities) []string {
	out := make([]string, 0, 2)
	if screenshotTool := browserCompatRegisteredOrEnabledToolForActKind(ctx, "screenshot"); capabilities.Screenshot && screenshotTool != "" {
		out = append(out, screenshotTool)
	}
	if (capabilities.Screenshot || capabilities.Download || capabilities.WaitDownload || capabilities.SavePDF || capabilities.SaveHTML || capabilities.TraceStop) && browserRuntimeRegisteredOrEnabledTool(ctx, "browser_act") {
		out = append(out, "browser_act")
	}
	return out
}

func browserRuntimeArtifactKinds(capabilities BrowserCapabilities) []string {
	out := make([]string, 0, 4)
	if capabilities.Screenshot {
		out = append(out, "screenshot")
	}
	if capabilities.Download || capabilities.WaitDownload {
		out = append(out, "download")
	}
	if capabilities.SavePDF {
		out = append(out, "pdf")
	}
	if capabilities.SaveHTML {
		out = append(out, "html")
	}
	if capabilities.TraceStop {
		out = append(out, "trace")
	}
	return out
}

func browserRuntimeArtifactContract(capabilities BrowserCapabilities) string {
	if len(browserRuntimeArtifactKinds(capabilities)) == 0 {
		return ""
	}
	return browserArtifactContract
}

func browserCapabilitiesForRuntimeInspection(ctx browserRegistrationContext, route browserResolvedExecutionRoute) BrowserCapabilities {
	return browserRuntimeCapabilitiesForResolvedRoute(ctx, route)
}

func browserRuntimeCapabilitiesMap(capabilities BrowserCapabilities) map[string]bool {
	return map[string]bool{
		"runtime_status":        capabilities.RuntimeStatus,
		"runtime_workbench":     capabilities.RuntimeWorkbench,
		"runtime_prepare":       capabilities.RuntimePrepare,
		"runtime_coordinate":    capabilities.RuntimeCoordinate,
		"runtime_start":         capabilities.RuntimeStart,
		"runtime_restart":       capabilities.RuntimeRestart,
		"runtime_stop":          capabilities.RuntimeStop,
		"runtime_create":        capabilities.RuntimeCreate,
		"runtime_delete":        capabilities.RuntimeDelete,
		"runtime_select":        capabilities.RuntimeSelect,
		"runtime_clear":         capabilities.RuntimeClear,
		"runtime_clear_session": capabilities.RuntimeClearSession,
		"runtime_sync_session":  capabilities.RuntimeSyncSession,
		"runtime_select_target": capabilities.RuntimeSelectTarget,
		"runtime_clear_target":  capabilities.RuntimeClearTarget,
		"runtime_list":          capabilities.RuntimeList,
		"runtime_sessions":      capabilities.RuntimeSessions,
		"open":                  capabilities.Open,
		"navigate":              capabilities.Navigate,
		"tabs":                  capabilities.Tabs,
		"extract":               capabilities.Extract,
		"snapshot":              capabilities.Snapshot,
		"screenshot":            capabilities.Screenshot,
		"console":               capabilities.Console,
		"requests":              capabilities.Requests,
		"response_body":         capabilities.ResponseBody,
		"errors":                capabilities.Errors,
		"cookies":               capabilities.Cookies,
		"cookies_set":           capabilities.CookiesSet,
		"cookies_clear":         capabilities.CookiesClear,
		"storage":               capabilities.Storage,
		"storage_set":           capabilities.StorageSet,
		"storage_clear":         capabilities.StorageClear,
		"offline":               capabilities.Offline,
		"headers":               capabilities.Headers,
		"credentials":           capabilities.Credentials,
		"geolocation":           capabilities.Geolocation,
		"media":                 capabilities.Media,
		"timezone":              capabilities.Timezone,
		"locale":                capabilities.Locale,
		"device":                capabilities.Device,
		"highlight":             capabilities.Highlight,
		"trace_start":           capabilities.TraceStart,
		"trace_stop":            capabilities.TraceStop,
		"download":              capabilities.Download,
		"wait_download":         capabilities.WaitDownload,
		"save_pdf":              capabilities.SavePDF,
		"save_html":             capabilities.SaveHTML,
		"dialog":                capabilities.Dialog,
		"upload":                capabilities.Upload,
		"click":                 capabilities.Click,
		"type":                  capabilities.TypeText,
		"evaluate":              capabilities.Evaluate,
		"wait":                  capabilities.Wait,
	}
}

func browserRuntimeCoordinationGoal(params map[string]any) string {
	goal := strings.ToLower(strings.TrimSpace(firstString(params, "coordination_goal", "goal", "mode")))
	switch goal {
	case "sync":
		return "sync"
	case "restart":
		return "restart"
	case "teardown":
		return "teardown"
	default:
		return "ensure"
	}
}

func browserRuntimeAdvanceSessionProfileState(ctx context.Context, registry agentxbrowserruntime.SharedSessionBrowserStateRegistry, sessionRegistry *BrowserSessionRegistry, selectedInfo BrowserRuntimeInfo, result browserRuntimePrepareResult) (agentxbrowserruntime.SharedSessionBrowserProfileState, bool) {
	sessionID, _ := browserRuntimeSessionStateRegistrySessionID(ctx, registry)
	return agentxbrowserruntime.SyncSharedSessionBrowserProfileLifecycleEvent(
		registry,
		sessionRegistry,
		nil,
		sessionID,
		selectedInfo,
		firstNonEmpty(strings.TrimSpace(result.ProfileStatus.Profile), strings.TrimSpace(result.Profile), strings.TrimSpace(selectedInfo.Profile)),
		result.ProfileStatus,
		result.Decision,
		time.Time{},
		browserRuntimeReconnectWatchdogWindow,
	)
}

func browserRuntimeSessionStateRegistrySessionID(ctx context.Context, registry agentxbrowserruntime.SharedSessionBrowserStateRegistry) (string, bool) {
	if registry == nil {
		return "", false
	}
	sessionID := ToolSessionIDFromContext(ctx)
	if sessionID == "" {
		return "", false
	}
	return sessionID, true
}

func browserRuntimeNormalizeSessionProfileState(state agentxbrowserruntime.SharedSessionBrowserProfileState) agentxbrowserruntime.SharedSessionBrowserProfileState {
	state.Backend = strings.TrimSpace(state.Backend)
	state.Profile = strings.TrimSpace(state.Profile)
	state.RuntimeTarget = strings.TrimSpace(state.RuntimeTarget)
	state.BrowserApp = strings.TrimSpace(state.BrowserApp)
	state.Status = strings.TrimSpace(state.Status)
	state.Note = strings.TrimSpace(state.Note)
	return state
}

func browserRuntimeCoordinationStatus(state string, goal string, profile *browserRuntimeProfileState, syncReady bool, prepareDecision string) agentxbrowserruntime.SharedSessionBrowserCoordinationStatus {
	var profileStatus *BrowserProfileStatusResult
	if profile != nil {
		status := agentxbrowserruntime.SharedSessionBrowserProfileStatusResultFromState(
			browserRuntimeSharedSessionProfileState(*profile),
			BrowserRuntimeInfo{
				Backend: profile.Backend,
				Profile: profile.Profile,
				Target:  profile.RuntimeTarget,
			},
			profile.Profile,
		)
		profileStatus = &status
	}
	return agentxbrowserruntime.EvaluateSharedSessionBrowserCoordinationStatus(state, goal, profileStatus, syncReady, prepareDecision)
}

func browserRuntimeLifecycleDecisionReady(selectedInfo BrowserRuntimeInfo, profile string, result BrowserProfileStatusResult, decision string) bool {
	return agentxbrowserruntime.SharedSessionBrowserLifecycleDecisionReady(selectedInfo, profile, result, decision)
}

func browserRuntimeLifecycleDecisionStatus(selectedInfo BrowserRuntimeInfo, profile string, result BrowserProfileStatusResult, decision string) BrowserProfileStatusResult {
	return agentxbrowserruntime.SharedSessionBrowserLifecycleDecisionStatus(selectedInfo, profile, result, decision)
}

func browserRuntimeProfileStateReady(state browserRuntimeProfileState) bool {
	return agentxbrowserruntime.SharedSessionBrowserProfileStateReady(agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       strings.TrimSpace(state.Backend),
		Profile:       strings.TrimSpace(state.Profile),
		RuntimeTarget: strings.TrimSpace(state.RuntimeTarget),
		BrowserApp:    strings.TrimSpace(state.BrowserApp),
		Status:        strings.TrimSpace(state.Status),
		Running:       state.Running,
		Connected:     state.Connected,
		Note:          strings.TrimSpace(state.Note),
		ObservedAt:    state.ObservedAt,
		StatusSince:   state.StatusSince,
	})
}

func browserRuntimeProfileReady(result BrowserProfileStatusResult) bool {
	return agentxbrowserruntime.SharedSessionBrowserProfileReady(result)
}

func browserRuntimeHostRuntimeInfo(ctx browserRegistrationContext) BrowserRuntimeInfo {
	return browserRuntimeHostRuntimeInfoForOptions(ctx.opts)
}

func browserRuntimeNodePromotionReady(ctx browserRegistrationContext) bool {
	return browserRuntimeDiagnosticsPreviewBaseForRegistration(ctx).Registration.SubstrateSummary.NodePromotionReady
}

func browserRuntimeRefreshSubstrateContext(ctx browserRegistrationContext, payload *browserRuntimePayload) {
	if payload == nil {
		return
	}
	preview := browserRuntimeDiagnosticsPreviewBaseForRegistration(ctx)
	browserRuntimeRefreshSubstrateContextWithPreview(ctx, payload, preview)
}

func browserRuntimeRefreshSubstrateContextWithPreview(ctx browserRegistrationContext, payload *browserRuntimePayload, preview browserRuntimeDiagnosticsPreview) {
	browserRuntimeApplyTopLevelRouteProjection(payload, browserRuntimeDiagnosticsRouteProjection(ctx, preview))
	browserRuntimePopulateSubstrateContextWithPreview(ctx, payload, preview)
}

func browserRuntimePopulateSubstrateContext(ctx browserRegistrationContext, payload *browserRuntimePayload) {
	if payload == nil {
		return
	}
	preview := browserRuntimeDiagnosticsPreviewBaseForRegistration(ctx)
	browserRuntimePopulateSubstrateContextWithPreview(ctx, payload, preview)
}

func browserRuntimeSubstrateMatrix(ctx browserRegistrationContext, substrate BrowserWorkbenchSubstrateSummary) []browserRuntimeSubstrateStatus {
	defaultRoute := browserRegistrationDefaultRuntimeInfo(ctx)
	substrateAssessment := browserRegistrationSubstrateAssessment(ctx)
	backend := browserRegistrationDefaultRuntimePreview(ctx).EffectiveBackend
	defaultAssessment := browserRuntimeSubstrateRouteAssessmentForBackend(
		backend,
		BrowserRuntimeInfo{},
		browserRuntimeDefaultSubstrateRouteAssessment(defaultRoute, substrateAssessment),
	)
	matrix := []browserRuntimeSubstrateStatus{
		browserRuntimeSubstrateDefaultRouteStatus(
			ctx,
			"default",
			substrate.SelectionReason,
			defaultRoute,
			defaultAssessment,
		),
	}
	if defaultRoute.Target != "host" || browserRuntimeUsesImplicitLegacyHostDefaultFallback(defaultRoute, substrateAssessment) {
		hostSelectionState := "explicit_fallback"
		hostReason := browserRuntimeSubstrateSelectionReason("host", substrate.HostRoute, defaultRoute)
		hostAssessment := browserRuntimeSubstrateRouteAssessmentForBackend(
			backend,
			BrowserRuntimeInfo{Target: "host"},
			substrateAssessment.HostRoute,
		)
		if !substrate.HostRouteAvailable {
			hostSelectionState = "unsupported"
			hostReason = firstNonEmpty(strings.TrimSpace(substrate.HostFailureCause), hostReason)
		}
		matrix = append(matrix, browserRuntimeSubstrateRouteStatus(
			ctx,
			"host",
			hostSelectionState,
			hostReason,
			substrate.HostRoute,
			hostAssessment,
		))
	}
	for _, managed := range browserRuntimeManagedRouteAssessments(ctx.opts, substrateAssessment, substrate) {
		if !managed.Configured || defaultRoute.Target == managed.Role {
			continue
		}
		managed.Assessment = browserRuntimeSubstrateRouteAssessmentForBackend(
			backend,
			BrowserRuntimeInfo{Target: managed.Role},
			managed.Assessment,
		)
		matrix = append(matrix, browserRuntimeSubstrateRouteStatus(
			ctx,
			managed.Role,
			managed.SelectionState(),
			managed.SelectionReason(defaultRoute),
			managed.RuntimeInfo,
			managed.Assessment,
		))
	}
	return matrix
}

func browserRuntimeSubstrateRouteAssessmentForBackend(backend BrowserBackend, requested BrowserRuntimeInfo, fallback browserConcreteRouteAssessment) browserConcreteRouteAssessment {
	requested = normalizeBrowserRuntimeInfo(requested)
	if assessment, ok := browserRuntimeRouterCachedRouteAssessment(backend, requested); ok {
		return assessment
	}
	return fallback
}

type browserRuntimeManagedRouteAssessment struct {
	Role                  string
	RuntimeInfo           BrowserRuntimeInfo
	Configured            bool
	PromotionReady        bool
	PromotionFailureCause string
	RouteFailureCause     string
	Assessment            browserConcreteRouteAssessment
}

func browserRuntimeManagedRouteAssessments(opts BrowserToolOptions, substrate browserDefaultSubstrateAssessment, summary BrowserWorkbenchSubstrateSummary) []browserRuntimeManagedRouteAssessment {
	return []browserRuntimeManagedRouteAssessment{
		{
			Role:                  "node",
			RuntimeInfo:           browserRuntimeInfoForConcreteBackend(opts.NodeBackend, defaultBrowserNodeRuntimeInfo()),
			Configured:            summary.NodeConfigured,
			PromotionReady:        summary.NodePromotionReady,
			PromotionFailureCause: strings.TrimSpace(summary.NodePromotionFailureCause),
			RouteFailureCause:     strings.TrimSpace(summary.NodePromotionFailureCause),
			Assessment:            browserConcreteRouteAssessmentForDefaultPromotion(substrate.NodeRoute),
		},
		{
			Role:                  "sandbox",
			RuntimeInfo:           browserRuntimeInfoForConcreteBackend(opts.SandboxBackend, defaultBrowserSandboxRuntimeInfo()),
			Configured:            summary.SandboxConfigured,
			PromotionReady:        summary.SandboxPromotionReady,
			PromotionFailureCause: strings.TrimSpace(summary.SandboxPromotionFailureCause),
			RouteFailureCause:     strings.TrimSpace(summary.SandboxFailureCause),
			Assessment:            substrate.SandboxConcreteRoute,
		},
	}
}

func browserRuntimeManagedRouteAssessmentsForPreview(
	ctx browserRegistrationContext,
	preview browserRuntimeDiagnosticsPreview,
) []browserRuntimeManagedRouteAssessment {
	summary := preview.Registration.SubstrateSummary
	substrate := preview.Registration.SubstrateAssessment
	return []browserRuntimeManagedRouteAssessment{
		{
			Role: "node",
			RuntimeInfo: firstBrowserRuntimeInfo(
				browserRuntimePreviewFallbackInfoForManagedTarget(
					preview,
					"node",
					defaultBrowserNodeRuntimeInfo(),
				),
				browserRuntimeInfoForConcreteBackend(ctx.opts.NodeBackend, defaultBrowserNodeRuntimeInfo()),
			),
			Configured:            summary.NodeConfigured,
			PromotionReady:        summary.NodePromotionReady,
			PromotionFailureCause: strings.TrimSpace(summary.NodePromotionFailureCause),
			RouteFailureCause:     strings.TrimSpace(summary.NodePromotionFailureCause),
			Assessment:            browserConcreteRouteAssessmentForDefaultPromotion(substrate.NodeRoute),
		},
		{
			Role: "sandbox",
			RuntimeInfo: firstBrowserRuntimeInfo(
				browserRuntimePreviewFallbackInfoForManagedTarget(
					preview,
					"sandbox",
					defaultBrowserSandboxRuntimeInfo(),
				),
				browserRuntimeInfoForConcreteBackend(ctx.opts.SandboxBackend, defaultBrowserSandboxRuntimeInfo()),
			),
			Configured:            summary.SandboxConfigured,
			PromotionReady:        summary.SandboxPromotionReady,
			PromotionFailureCause: strings.TrimSpace(summary.SandboxPromotionFailureCause),
			RouteFailureCause:     strings.TrimSpace(summary.SandboxFailureCause),
			Assessment:            substrate.SandboxConcreteRoute,
		},
	}
}

func firstBrowserRuntimeInfo(candidates ...BrowserRuntimeInfo) BrowserRuntimeInfo {
	for _, candidate := range candidates {
		if info := normalizeBrowserRuntimeInfo(candidate); info != (BrowserRuntimeInfo{}) {
			return info
		}
	}
	return BrowserRuntimeInfo{}
}

func (assessment browserRuntimeManagedRouteAssessment) SelectionState() string {
	if !assessment.Assessment.RouteAvailable {
		return "unsupported"
	}
	return "available"
}

func (assessment browserRuntimeManagedRouteAssessment) SelectionReason(defaultRoute BrowserRuntimeInfo) string {
	if !assessment.Assessment.RouteAvailable {
		return firstNonEmpty(
			strings.TrimSpace(assessment.RouteFailureCause),
			strings.TrimSpace(assessment.Assessment.FailureReason),
			strings.TrimSpace(assessment.PromotionFailureCause),
		)
	}
	if !assessment.PromotionReady && strings.TrimSpace(assessment.PromotionFailureCause) != "" {
		return strings.TrimSpace(assessment.PromotionFailureCause)
	}
	return browserRuntimeSubstrateSelectionReason(assessment.Role, assessment.Assessment.Route.RuntimeInfo, defaultRoute)
}

func browserRuntimeDefaultSubstrateRouteAssessment(defaultRoute BrowserRuntimeInfo, substrate browserDefaultSubstrateAssessment) browserConcreteRouteAssessment {
	if normalizeBrowserRuntimeInfo(defaultRoute) == browserDefaultRouteRuntimeInfoForAssessment(substrate) &&
		(substrate.DefaultConcreteRoute.Configured ||
			substrate.DefaultConcreteRoute.RouteAvailable ||
			strings.TrimSpace(substrate.DefaultConcreteRoute.FailureReason) != "" ||
			strings.TrimSpace(substrate.DefaultConcreteRoute.FailureNote) != "") {
		return substrate.DefaultConcreteRoute
	}
	if browserRuntimeUsesImplicitLegacyHostDefaultFallback(defaultRoute, substrate) {
		reason := strings.TrimSpace(browserWorkbenchSubstrateSelectionReason(
			defaultRoute,
			substrate.HostRuntime,
			substrate.HostRoute,
			substrate.NodeRoute,
			substrate.SandboxRoute,
		))
		return browserConcreteRouteAssessment{
			FailureReason: reason,
			FailureNote:   reason,
		}
	}
	return browserDefaultConcreteRouteAssessment(defaultRoute, substrate)
}

func browserConcreteRouteAssessmentForDefaultPromotion(assessment browserDefaultPromotionRouteAssessment) browserConcreteRouteAssessment {
	return browserConcreteRouteAssessment{
		Configured:     assessment.Configured,
		RouteAvailable: assessment.RouteAvailable,
		Route:          assessment.Route,
		FailureReason:  assessment.FailureReason,
		FailureNote:    assessment.FailureNote,
	}
}

func browserRuntimeSubstrateRouteStatus(ctx browserRegistrationContext, role string, selectionState string, selectionReason string, info BrowserRuntimeInfo, assessment browserConcreteRouteAssessment) browserRuntimeSubstrateStatus {
	return browserRuntimeSubstrateRouteStatusForAssessment(ctx, role, selectionState, selectionReason, info, assessment)
}

func browserRuntimeSubstrateRouteStatusForAssessment(ctx browserRegistrationContext, role string, selectionState string, selectionReason string, info BrowserRuntimeInfo, assessment browserConcreteRouteAssessment) browserRuntimeSubstrateStatus {
	defaultCandidateRoute := browserRuntimeHiddenDefaultCandidateRouteDescriptorForStatus(ctx, role, info, assessment)
	projection := browserRuntimeSubstrateAssessmentSurfaceProjection(ctx, role, info, assessment)
	info = projection.Info
	routeMetadata := browserRuntimeStatusRouteMetadataForAssessment(ctx, info, assessment)
	if !assessment.RouteAvailable {
		failureSurfaceNote := browserRuntimeAssessmentFailureSurfaceNote(assessment, selectionReason)
		status := browserRuntimeSubstrateStatus{
			Role:                  role,
			SelectionState:        firstNonEmpty(strings.TrimSpace(selectionState), "unsupported"),
			SelectionReason:       firstNonEmpty(failureSurfaceNote, strings.TrimSpace(selectionReason), strings.TrimSpace(assessment.FailureReason)),
			Profile:               strings.TrimSpace(info.Profile),
			RuntimeTarget:         strings.TrimSpace(info.Target),
			Backend:               strings.TrimSpace(info.Backend),
			Source:                strings.TrimSpace(routeMetadata.Source),
			Endpoint:              strings.TrimSpace(routeMetadata.Endpoint),
			DefaultCandidateRoute: defaultCandidateRoute,
			Status:                "unsupported",
			Note:                  firstNonEmpty(failureSurfaceNote, strings.TrimSpace(assessment.FailureNote), strings.TrimSpace(assessment.FailureReason)),
		}
		if projection.HasMetadata {
			browserRuntimeApplyCapabilityMetadataToSubstrateStatus(&status, projection.Metadata)
		}
		return status
	}
	route := assessment.Route
	status := "available"
	if strings.EqualFold(strings.TrimSpace(role), "default") {
		status = "default"
	}
	out := browserRuntimeSubstrateStatus{
		Role:                  role,
		SelectionState:        firstNonEmpty(strings.TrimSpace(selectionState), "available"),
		SelectionReason:       strings.TrimSpace(selectionReason),
		Profile:               route.RuntimeInfo.Profile,
		RuntimeTarget:         route.RuntimeInfo.Target,
		Backend:               route.RuntimeInfo.Backend,
		Source:                strings.TrimSpace(routeMetadata.Source),
		Endpoint:              strings.TrimSpace(routeMetadata.Endpoint),
		DefaultCandidateRoute: defaultCandidateRoute,
		Status:                status,
		SubstratePosture:      BrowserSubstratePosture(route.RuntimeInfo.Backend, route.RuntimeInfo.Target),
		SubstrateStatus:       BrowserSubstrateStatus(route.RuntimeInfo.Backend, route.RuntimeInfo.Target),
		SubstrateReason:       BrowserSubstrateReason(route.RuntimeInfo.Backend, route.RuntimeInfo.Target),
	}
	if projection.HasMetadata {
		browserRuntimeApplyCapabilityMetadataToSubstrateStatus(&out, projection.Metadata)
	}
	return out
}

func browserRuntimeSubstrateRouteStatusForAssessmentWithPreview(ctx browserRegistrationContext, preview browserRuntimeDiagnosticsPreview, role string, selectionState string, selectionReason string, info BrowserRuntimeInfo, assessment browserConcreteRouteAssessment) browserRuntimeSubstrateStatus {
	projection := browserRuntimeSubstrateAssessmentSurfaceProjectionForPreview(
		ctx,
		role,
		preview,
		info,
		assessment,
	)
	info = projection.Info
	routeMetadata := browserRuntimeStatusRouteMetadataForAssessmentWithPreview(ctx, preview, info, assessment)
	if !assessment.RouteAvailable {
		failureSurfaceNote := browserRuntimeAssessmentFailureSurfaceNote(assessment, selectionReason)
		status := browserRuntimeSubstrateStatus{
			Role:            role,
			SelectionState:  firstNonEmpty(strings.TrimSpace(selectionState), "unsupported"),
			SelectionReason: firstNonEmpty(failureSurfaceNote, strings.TrimSpace(selectionReason), strings.TrimSpace(assessment.FailureReason)),
			Profile:         strings.TrimSpace(info.Profile),
			RuntimeTarget:   strings.TrimSpace(info.Target),
			Backend:         strings.TrimSpace(info.Backend),
			Source:          strings.TrimSpace(routeMetadata.Source),
			Endpoint:        strings.TrimSpace(routeMetadata.Endpoint),
			Status:          "unsupported",
			Note:            firstNonEmpty(failureSurfaceNote, strings.TrimSpace(assessment.FailureNote), strings.TrimSpace(assessment.FailureReason)),
		}
		if projection.HasMetadata {
			browserRuntimeApplyCapabilityMetadataToSubstrateStatus(&status, projection.Metadata)
		}
		return status
	}
	route := assessment.Route
	status := "available"
	if strings.EqualFold(strings.TrimSpace(role), "default") {
		status = "default"
	}
	out := browserRuntimeSubstrateStatus{
		Role:             role,
		SelectionState:   firstNonEmpty(strings.TrimSpace(selectionState), "available"),
		SelectionReason:  strings.TrimSpace(selectionReason),
		Profile:          route.RuntimeInfo.Profile,
		RuntimeTarget:    route.RuntimeInfo.Target,
		Backend:          route.RuntimeInfo.Backend,
		Source:           strings.TrimSpace(routeMetadata.Source),
		Endpoint:         strings.TrimSpace(routeMetadata.Endpoint),
		Status:           status,
		SubstratePosture: BrowserSubstratePosture(route.RuntimeInfo.Backend, route.RuntimeInfo.Target),
		SubstrateStatus:  BrowserSubstrateStatus(route.RuntimeInfo.Backend, route.RuntimeInfo.Target),
		SubstrateReason:  BrowserSubstrateReason(route.RuntimeInfo.Backend, route.RuntimeInfo.Target),
	}
	if projection.HasMetadata {
		browserRuntimeApplyCapabilityMetadataToSubstrateStatus(&out, projection.Metadata)
	}
	return out
}

func browserRuntimeSubstrateSelectionReason(role string, info BrowserRuntimeInfo, defaultRoute BrowserRuntimeInfo) string {
	info = normalizeBrowserRuntimeInfo(info)
	defaultRoute = normalizeBrowserRuntimeInfo(defaultRoute)
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "host":
		if BrowserSubstratePosture(info.Backend, info.Target) == BrowserSubstrateLegacySystemHost && defaultRoute.Target != "host" {
			return "legacy host route remains available only via explicit `runtime_target=host`"
		}
		return fmt.Sprintf("host runtime backend `%s` remains available via explicit `runtime_target=host`", info.Backend)
	case "node":
		return fmt.Sprintf("node runtime backend `%s` is available via `runtime_target=node`", info.Backend)
	case "sandbox":
		return fmt.Sprintf("sandbox runtime backend `%s` is available via `runtime_target=sandbox`", info.Backend)
	default:
		return ""
	}
}

func browserRuntimeRouteMatrix(ctx browserRegistrationContext, profiles []string) []browserRuntimeRouteStatus {
	defaultRoute := browserRegistrationDefaultRuntimeInfo(ctx)
	routes := []browserRuntimeRouteStatus{}
	substrateAssessment := browserRegistrationSubstrateAssessment(ctx)
	backend := browserRegistrationDefaultRuntimePreview(ctx).EffectiveBackend
	defaultCandidateRoute := browserRuntimeDefaultCandidateRouteDescriptor(browserRuntimeDiagnosticsPreviewBaseForRegistration(ctx))
	implicitLegacyHostFallback := browserRuntimeUsesImplicitLegacyHostDefaultFallback(defaultRoute, substrateAssessment)
	defaultAssessment, ok := browserRuntimeRouterCachedRouteAssessment(backend, BrowserRuntimeInfo{})
	if !ok {
		defaultAssessment = browserRuntimeDefaultSubstrateRouteAssessment(defaultRoute, substrateAssessment)
	}
	if defaultAssessment.RouteAvailable {
		status := browserRuntimeRouteStatusForResolvedRoute(ctx, "default", defaultAssessment.Route)
		status.DefaultCandidateRoute = defaultCandidateRoute
		routes = append(routes, status)
	} else if browserConcreteRouteAssessmentHasResult(defaultAssessment) {
		status := browserRuntimeRouteStatusForAssessment(ctx, "unsupported", defaultRoute, defaultAssessment)
		status.DefaultCandidateRoute = defaultCandidateRoute
		routes = append(routes, status)
	}

	if len(profiles) == 0 {
		profiles = []string{"default", "isolated", "relay"}
	}
	for _, target := range browserRegistrationSubstrateSummary(ctx).ConfiguredTargets {
		for _, profile := range profiles {
			if profile == defaultRoute.Profile && target == defaultRoute.Target && !implicitLegacyHostFallback {
				continue
			}
			if cachedAssessment, ok := browserRuntimeRouterCachedRouteAssessment(backend, BrowserRuntimeInfo{
				Profile: profile,
				Target:  target,
			}); ok {
				if cachedAssessment.RouteAvailable {
					routes = append(routes, browserRuntimeRouteStatusForResolvedRoute(ctx, "available", cachedAssessment.Route))
				} else {
					routes = append(routes, browserRuntimeRouteStatusForAssessment(
						ctx,
						"unsupported",
						browserRuntimeRouteFallbackInfoForTarget(ctx, profile, target),
						cachedAssessment,
					))
				}
				continue
			}
			params := map[string]any{
				"profile":        profile,
				"runtime_target": target,
			}
			route, routeErr := resolveBrowserExecutionRoute(params, defaultRoute, backend)
			if routeErr != nil {
				routes = append(routes, browserRuntimeRouteStatusForAssessment(
					ctx,
					"unsupported",
					browserRuntimeRouteFallbackInfoForTarget(ctx, profile, target),
					browserConcreteRouteAssessment{
						Configured:    true,
						FailureReason: routeErr.Error(),
						FailureNote:   routeErr.Error(),
					},
				))
				continue
			}
			routes = append(routes, browserRuntimeRouteStatusForResolvedRoute(ctx, "available", route))
		}
	}
	return routes
}

func browserRuntimeRouteMatrixWithPreview(ctx browserRegistrationContext, preview browserRuntimeDiagnosticsPreview, profiles []string) []browserRuntimeRouteStatus {
	defaultRoute := preview.DefaultRoute
	routes := []browserRuntimeRouteStatus{}
	substrateAssessment := preview.Registration.SubstrateAssessment
	backend := preview.Registration.EffectiveBackend
	implicitLegacyHostFallback := browserRuntimeUsesImplicitLegacyHostDefaultFallback(defaultRoute, substrateAssessment)
	defaultAssessment, ok := browserRuntimeRouterCachedRouteAssessment(backend, BrowserRuntimeInfo{})
	if !ok {
		defaultAssessment = browserRuntimeDefaultSubstrateRouteAssessment(defaultRoute, substrateAssessment)
	}
	if defaultAssessment.RouteAvailable {
		status := browserRuntimeRouteStatusForResolvedRouteWithPreview(ctx, preview, "default", defaultAssessment.Route)
		status.DefaultCandidateRoute = browserRuntimeDefaultCandidateRouteDescriptor(preview)
		routes = append(routes, status)
	} else if browserConcreteRouteAssessmentHasResult(defaultAssessment) {
		status := browserRuntimeRouteStatusForAssessmentWithPreview(ctx, preview, "unsupported", defaultRoute, defaultAssessment)
		status.DefaultCandidateRoute = browserRuntimeDefaultCandidateRouteDescriptor(preview)
		routes = append(routes, status)
	}

	if len(profiles) == 0 {
		profiles = []string{"default", "isolated", "relay"}
	}
	for _, target := range preview.ConfiguredTargets {
		for _, profile := range profiles {
			if profile == defaultRoute.Profile && target == defaultRoute.Target && !implicitLegacyHostFallback {
				continue
			}
			if cachedAssessment, ok := browserRuntimeRouterCachedRouteAssessment(backend, BrowserRuntimeInfo{
				Profile: profile,
				Target:  target,
			}); ok {
				if cachedAssessment.RouteAvailable {
					routes = append(routes, browserRuntimeRouteStatusForResolvedRouteWithPreview(ctx, preview, "available", cachedAssessment.Route))
				} else {
					routes = append(routes, browserRuntimeRouteStatusForAssessmentWithPreview(
						ctx,
						preview,
						"unsupported",
						browserRuntimeRouteFallbackInfoForPreviewTarget(ctx, preview, profile, target),
						cachedAssessment,
					))
				}
				continue
			}
			params := map[string]any{
				"profile":        profile,
				"runtime_target": target,
			}
			route, routeErr := resolveBrowserExecutionRoute(params, defaultRoute, backend)
			if routeErr != nil {
				routes = append(routes, browserRuntimeRouteStatusForAssessmentWithPreview(
					ctx,
					preview,
					"unsupported",
					browserRuntimeRouteFallbackInfoForPreviewTarget(ctx, preview, profile, target),
					browserConcreteRouteAssessment{
						Configured:    true,
						FailureReason: routeErr.Error(),
						FailureNote:   routeErr.Error(),
					},
				))
				continue
			}
			routes = append(routes, browserRuntimeRouteStatusForResolvedRouteWithPreview(ctx, preview, "available", route))
		}
	}
	return routes
}

func browserRuntimeRouteFallbackInfoForTarget(ctx browserRegistrationContext, profile string, target string) BrowserRuntimeInfo {
	if ctx.derivedCache != nil {
		return ctx.derivedCache.routeFallbackInfo(ctx, profile, target)
	}
	return browserRuntimeRouteFallbackInfoForPreviewTarget(
		ctx,
		browserRuntimeDiagnosticsPreviewBaseForRegistration(ctx),
		profile,
		target,
	)
}

func browserRuntimeRouteFallbackInfoForPreviewTarget(
	ctx browserRegistrationContext,
	preview browserRuntimeDiagnosticsPreview,
	profile string,
	target string,
) BrowserRuntimeInfo {
	profile = strings.ToLower(strings.TrimSpace(profile))
	target = strings.ToLower(strings.TrimSpace(target))
	substrateAssessment := preview.Registration.SubstrateAssessment
	switch target {
	case "host":
		info := normalizeBrowserRuntimeInfo(substrateAssessment.HostRuntime)
		info.Profile = firstNonEmpty(profile, info.Profile)
		info.Target = "host"
		return info
	case "node":
		info := browserRuntimePreviewFallbackInfoForManagedTarget(
			preview,
			"node",
			defaultBrowserNodeRuntimeInfo(),
		)
		if info == (BrowserRuntimeInfo{}) {
			info = browserRuntimeInfoForConcreteBackend(ctx.opts.NodeBackend, defaultBrowserNodeRuntimeInfo())
		}
		info = normalizeBrowserRuntimeInfo(info)
		info.Profile = firstNonEmpty(profile, info.Profile)
		info.Target = "node"
		return info
	case "sandbox":
		info := browserRuntimePreviewFallbackInfoForManagedTarget(
			preview,
			"sandbox",
			defaultBrowserSandboxRuntimeInfo(),
		)
		if info == (BrowserRuntimeInfo{}) {
			info = browserRuntimeInfoForConcreteBackend(ctx.opts.SandboxBackend, defaultBrowserSandboxRuntimeInfo())
		}
		info = normalizeBrowserRuntimeInfo(info)
		info.Profile = firstNonEmpty(profile, info.Profile)
		info.Target = "sandbox"
		return info
	default:
		info := normalizeBrowserRuntimeInfo(preview.DefaultRoute)
		info.Profile = firstNonEmpty(profile, info.Profile)
		info.Target = firstNonEmpty(target, info.Target)
		return info
	}
}

func browserRuntimePreviewFallbackInfoForManagedTarget(
	preview browserRuntimeDiagnosticsPreview,
	target string,
	fallback BrowserRuntimeInfo,
) BrowserRuntimeInfo {
	if assessment, ok := browserRuntimePreviewManagedRouteAssessment(preview, target); ok && assessment.RouteAvailable {
		if info := normalizeBrowserRuntimeInfo(assessment.Route.RuntimeInfo); info != (BrowserRuntimeInfo{}) {
			return info
		}
	}
	backend := browserRuntimePreviewManagedTargetBackend(preview, target)
	if backend == nil {
		return BrowserRuntimeInfo{}
	}
	return normalizeBrowserRuntimeInfo(browserRuntimeInfoForConcreteBackend(backend, fallback))
}

func browserRuntimePreviewManagedTargetBackend(preview browserRuntimeDiagnosticsPreview, target string) BrowserBackend {
	switch backend := preview.Registration.EffectiveBackend.(type) {
	case browserRuntimeRouterBackend:
		return browserRuntimeRouterManagedTargetBackend(backend, target)
	case *browserRuntimeRouterBackend:
		if backend == nil {
			return nil
		}
		return browserRuntimeRouterManagedTargetBackend(*backend, target)
	default:
		return nil
	}
}

func browserRuntimeRouterManagedTargetBackend(router browserRuntimeRouterBackend, target string) BrowserBackend {
	switch strings.ToLower(strings.TrimSpace(target)) {
	case "node":
		return router.nodeBackend
	case "sandbox":
		return router.sandboxBackend
	default:
		return nil
	}
}

func browserCachedRouteAssessmentProfile(assessment browserConcreteRouteAssessment, fallbackProfile string) string {
	profile := strings.ToLower(strings.TrimSpace(fallbackProfile))
	if assessment.RouteAvailable {
		profile = firstNonEmpty(strings.ToLower(strings.TrimSpace(assessment.Route.RuntimeInfo.Profile)), profile)
	}
	return profile
}

func browserCachedRouteAssessmentForProfile(profile string, cachedProfile string, assessment browserConcreteRouteAssessment) (browserConcreteRouteAssessment, bool) {
	if profile == "" {
		return browserConcreteRouteAssessment{}, false
	}
	if strings.ToLower(strings.TrimSpace(cachedProfile)) != profile {
		return browserConcreteRouteAssessment{}, false
	}
	if assessment.RouteAvailable || strings.TrimSpace(assessment.FailureReason) != "" || strings.TrimSpace(assessment.FailureNote) != "" {
		return assessment, true
	}
	return browserConcreteRouteAssessment{}, false
}

func browserCachedRouteAssessmentForDiagnosticsProfile(profile string, cachedProfile string, assessment browserConcreteRouteAssessment) (browserConcreteRouteAssessment, bool) {
	if cached, ok := browserCachedRouteAssessmentForProfile(profile, cachedProfile, assessment); ok {
		return cached, true
	}
	if assessment.RouteAvailable {
		return browserConcreteRouteAssessment{}, false
	}
	if strings.TrimSpace(assessment.FailureReason) == "" && strings.TrimSpace(assessment.FailureNote) == "" {
		return browserConcreteRouteAssessment{}, false
	}
	if strings.TrimSpace(profile) == "" {
		return browserConcreteRouteAssessment{}, false
	}
	// Diagnostics matrix rows can safely reuse a lane-level cached failure even
	// when the requested profile differs, because the row still renders the
	// caller's requested profile through browserRuntimeRouteFallbackInfoForTarget.
	return assessment, true
}

func browserConcreteRouteAssessmentHasResult(assessment browserConcreteRouteAssessment) bool {
	return assessment.Configured ||
		assessment.RouteAvailable ||
		strings.TrimSpace(assessment.FailureReason) != "" ||
		strings.TrimSpace(assessment.FailureNote) != ""
}

func browserRuntimeShouldHideImplicitLegacyHostDefaultInfo(ctx browserRegistrationContext, role string, info BrowserRuntimeInfo, assessment browserConcreteRouteAssessment) bool {
	return browserRuntimeShouldHideImplicitLegacyHostDefaultInfoForDefaultRoute(
		role,
		info,
		assessment,
		browserRegistrationDefaultRuntimeInfo(ctx),
		browserRegistrationSubstrateAssessment(ctx),
	)
}

func browserRuntimeShouldHideImplicitLegacyHostDefaultInfoForDefaultRoute(role string, info BrowserRuntimeInfo, assessment browserConcreteRouteAssessment, defaultRoute BrowserRuntimeInfo, substrateAssessment browserDefaultSubstrateAssessment) bool {
	if !strings.EqualFold(strings.TrimSpace(role), "default") || assessment.RouteAvailable {
		return false
	}
	defaultRoute = normalizeBrowserRuntimeInfo(defaultRoute)
	if normalizeBrowserRuntimeInfo(info) != defaultRoute {
		return false
	}
	return browserRuntimeUsesImplicitLegacyHostDefaultFallback(defaultRoute, substrateAssessment)
}

func browserRuntimeRouteStatusForAssessment(ctx browserRegistrationContext, status string, info BrowserRuntimeInfo, assessment browserConcreteRouteAssessment) browserRuntimeRouteStatus {
	defaultCandidateRoute := browserRuntimeHiddenDefaultCandidateRouteDescriptorForStatus(ctx, "default", info, assessment)
	projection := browserRuntimeRouteAssessmentSurfaceProjection(ctx, info, assessment)
	info = projection.Info
	failureSurfaceNote := browserRuntimeAssessmentFailureSurfaceNote(assessment)
	routeMetadata := browserRuntimeStatusRouteMetadataForAssessment(ctx, info, assessment)
	routeStatus := browserRuntimeRouteStatus{
		Profile:               strings.TrimSpace(info.Profile),
		RuntimeTarget:         strings.TrimSpace(info.Target),
		Backend:               strings.TrimSpace(info.Backend),
		Source:                strings.TrimSpace(routeMetadata.Source),
		Endpoint:              strings.TrimSpace(routeMetadata.Endpoint),
		DefaultCandidateRoute: defaultCandidateRoute,
		Status:                firstNonEmpty(strings.TrimSpace(status), "unsupported"),
		Note:                  firstNonEmpty(failureSurfaceNote, strings.TrimSpace(assessment.FailureNote), strings.TrimSpace(assessment.FailureReason)),
	}
	if projection.HasMetadata {
		browserRuntimeApplyCapabilityMetadataToRouteStatus(&routeStatus, projection.Metadata)
	}
	return routeStatus
}

func browserRuntimeRouteStatusForAssessmentWithPreview(ctx browserRegistrationContext, preview browserRuntimeDiagnosticsPreview, status string, info BrowserRuntimeInfo, assessment browserConcreteRouteAssessment) browserRuntimeRouteStatus {
	projection := browserRuntimeRouteAssessmentSurfaceProjectionForPreview(
		ctx,
		preview,
		info,
		assessment,
	)
	info = projection.Info
	failureSurfaceNote := browserRuntimeAssessmentFailureSurfaceNote(assessment)
	defaultCandidateRoute := browserRuntimeDefaultCandidateRouteDescriptor(preview)
	routeMetadata := browserRuntimeStatusRouteMetadataForAssessmentWithPreview(ctx, preview, info, assessment)
	routeStatus := browserRuntimeRouteStatus{
		Profile:               strings.TrimSpace(info.Profile),
		RuntimeTarget:         strings.TrimSpace(info.Target),
		Backend:               strings.TrimSpace(info.Backend),
		Source:                strings.TrimSpace(routeMetadata.Source),
		Endpoint:              strings.TrimSpace(routeMetadata.Endpoint),
		DefaultCandidateRoute: defaultCandidateRoute,
		Status:                firstNonEmpty(strings.TrimSpace(status), "unsupported"),
		Note:                  firstNonEmpty(failureSurfaceNote, strings.TrimSpace(assessment.FailureNote), strings.TrimSpace(assessment.FailureReason)),
	}
	if projection.HasMetadata {
		browserRuntimeApplyCapabilityMetadataToRouteStatus(&routeStatus, projection.Metadata)
	}
	return routeStatus
}

func browserRuntimeHiddenDefaultCandidateRouteDescriptorForStatus(
	ctx browserRegistrationContext,
	role string,
	info BrowserRuntimeInfo,
	assessment browserConcreteRouteAssessment,
) browserRuntimeRouteDescriptor {
	if !strings.EqualFold(strings.TrimSpace(role), "default") || assessment.RouteAvailable {
		return browserRuntimeRouteDescriptor{}
	}
	defaultRoute := normalizeBrowserRuntimeInfo(browserRegistrationDefaultRuntimeInfo(ctx))
	if normalizeBrowserRuntimeInfo(info) != defaultRoute {
		return browserRuntimeRouteDescriptor{}
	}
	if !browserRuntimeUsesImplicitLegacyHostDefaultFallback(defaultRoute, browserRegistrationSubstrateAssessment(ctx)) {
		return browserRuntimeRouteDescriptor{}
	}
	return browserRuntimeDefaultCandidateRouteDescriptor(browserRuntimeDiagnosticsPreviewBaseForRegistration(ctx))
}

func browserRuntimeAssessmentFailureSurfaceNote(assessment browserConcreteRouteAssessment, values ...string) string {
	args := make([]string, 0, len(values)+2)
	args = append(args, values...)
	args = append(args, strings.TrimSpace(assessment.FailureNote), strings.TrimSpace(assessment.FailureReason))
	return browserRuntimeBootstrapBlockedSurfaceNoteForFailureText(args...)
}

func browserRuntimeRouteStatusForResolvedRoute(ctx browserRegistrationContext, status string, route browserResolvedExecutionRoute) browserRuntimeRouteStatus {
	routeMetadata := browserRuntimeStatusRouteMetadataForResolvedRoute(route)
	if routeMetadata == (browserDoctorRouteMetadata{}) {
		routeMetadata = browserRuntimeStatusRouteMetadataForAssessment(
			ctx,
			route.RuntimeInfo,
			browserConcreteRouteAssessment{
				Configured:     true,
				RouteAvailable: true,
				Route:          route,
			},
		)
	}
	routeStatus := browserRuntimeRouteStatus{
		Profile:       route.RuntimeInfo.Profile,
		RuntimeTarget: route.RuntimeInfo.Target,
		Backend:       route.RuntimeInfo.Backend,
		Source:        strings.TrimSpace(routeMetadata.Source),
		Endpoint:      strings.TrimSpace(routeMetadata.Endpoint),
		Status:        status,
	}
	metadata, _ := browserRuntimeResolvedRouteSurfaceMetadata(ctx, route)
	browserRuntimeApplyCapabilityMetadataToRouteStatus(&routeStatus, metadata)
	return routeStatus
}

func browserRuntimeRouteStatusForResolvedRouteWithPreview(ctx browserRegistrationContext, preview browserRuntimeDiagnosticsPreview, status string, route browserResolvedExecutionRoute) browserRuntimeRouteStatus {
	routeMetadata := browserRuntimeStatusRouteMetadataForResolvedRoute(route)
	if routeMetadata == (browserDoctorRouteMetadata{}) {
		routeMetadata = browserRuntimeStatusRouteMetadataForAssessmentWithPreview(
			ctx,
			preview,
			route.RuntimeInfo,
			browserConcreteRouteAssessment{
				Configured:     true,
				RouteAvailable: true,
				Route:          route,
			},
		)
	}
	routeStatus := browserRuntimeRouteStatus{
		Profile:       route.RuntimeInfo.Profile,
		RuntimeTarget: route.RuntimeInfo.Target,
		Backend:       route.RuntimeInfo.Backend,
		Source:        strings.TrimSpace(routeMetadata.Source),
		Endpoint:      strings.TrimSpace(routeMetadata.Endpoint),
		Status:        status,
	}
	metadata, _ := browserRuntimeResolvedRouteSurfaceMetadataWithPreview(ctx, preview, route)
	browserRuntimeApplyCapabilityMetadataToRouteStatus(&routeStatus, metadata)
	return routeStatus
}
