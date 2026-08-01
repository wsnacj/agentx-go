package productshell

// HostProcessProgressObservationInput is typed host progress input.
type HostProcessProgressObservationInput struct {
	Source                            string   `json:"source,omitempty"`
	Available                         bool     `json:"available"`
	Enabled                           bool     `json:"enabled"`
	Status                            string   `json:"status,omitempty"`
	DisplayKind                       string   `json:"display_kind,omitempty"`
	SummaryCode                       string   `json:"summary_code,omitempty"`
	DisplayLine                       string   `json:"display_line,omitempty"`
	SessionKey                        string   `json:"session_key,omitempty"`
	ProductShellRef                   string   `json:"product_shell_ref,omitempty"`
	WorkspaceRef                      string   `json:"workspace_ref,omitempty"`
	ProcessRef                        string   `json:"process_ref,omitempty"`
	RunID                             string   `json:"run_id,omitempty"`
	BranchID                          string   `json:"branch_id,omitempty"`
	NodeExecID                        string   `json:"node_exec_id,omitempty"`
	RuntimeDecisionRef                string   `json:"runtime_decision_ref,omitempty"`
	RequestRef                        string   `json:"request_ref,omitempty"`
	ResultRef                         string   `json:"result_ref,omitempty"`
	ArtifactBundleRef                 string   `json:"artifact_bundle_ref,omitempty"`
	StdoutRef                         string   `json:"stdout_ref,omitempty"`
	StderrRef                         string   `json:"stderr_ref,omitempty"`
	ProcessStatus                     string   `json:"process_status,omitempty"`
	LastKind                          string   `json:"last_kind,omitempty"`
	ExitCode                          int      `json:"exit_code,omitempty"`
	ExitCodeKnown                     bool     `json:"exit_code_known"`
	ReadyForReadback                  bool     `json:"ready_for_readback"`
	Started                           bool     `json:"started"`
	Terminal                          bool     `json:"terminal"`
	Failed                            bool     `json:"failed"`
	Cancelled                         bool     `json:"cancelled"`
	TimedOut                          bool     `json:"timed_out"`
	ProcessCount                      int      `json:"process_count"`
	ActiveCount                       int      `json:"active_count,omitempty"`
	TerminalCount                     int      `json:"terminal_count,omitempty"`
	HostProcessEventCount             int      `json:"host_process_event_count,omitempty"`
	ViewReady                         bool     `json:"view_ready"`
	ProgressReady                     bool     `json:"progress_ready"`
	ConsumesHostProcessSessionView    bool     `json:"consumes_host_process_session_view"`
	ReadsToolOutput                   bool     `json:"reads_tool_output"`
	BuildsProcessMapFromToolOutput    bool     `json:"builds_process_map_from_tool_output"`
	RunstoreProtocolAuthorizesProcess bool     `json:"runstore_protocol_authorizes_process"`
	ProcessLifecycleControlsExecution bool     `json:"process_lifecycle_controls_execution"`
	MissingInputs                     []string `json:"missing_inputs,omitempty"`
	BlockedReasons                    []string `json:"blocked_reasons,omitempty"`
	Boundaries                        []string `json:"boundaries,omitempty"`
	NextHostAction                    string   `json:"next_host_action,omitempty"`
	RawOutputLoaded                   bool     `json:"raw_output_loaded"`
}

// HostProcessProgressObservation is a normalized, display-only host progress observation.
type HostProcessProgressObservation struct {
	Source                            string   `json:"source,omitempty"`
	Available                         bool     `json:"available"`
	Enabled                           bool     `json:"enabled"`
	Status                            string   `json:"status,omitempty"`
	DisplayKind                       string   `json:"display_kind,omitempty"`
	SummaryCode                       string   `json:"summary_code,omitempty"`
	DisplayLine                       string   `json:"display_line,omitempty"`
	SessionKey                        string   `json:"session_key,omitempty"`
	ProductShellRef                   string   `json:"product_shell_ref,omitempty"`
	WorkspaceRef                      string   `json:"workspace_ref,omitempty"`
	ProcessRef                        string   `json:"process_ref,omitempty"`
	RunID                             string   `json:"run_id,omitempty"`
	BranchID                          string   `json:"branch_id,omitempty"`
	NodeExecID                        string   `json:"node_exec_id,omitempty"`
	RuntimeDecisionRef                string   `json:"runtime_decision_ref,omitempty"`
	RequestRef                        string   `json:"request_ref,omitempty"`
	ResultRef                         string   `json:"result_ref,omitempty"`
	ArtifactBundleRef                 string   `json:"artifact_bundle_ref,omitempty"`
	StdoutRef                         string   `json:"stdout_ref,omitempty"`
	StderrRef                         string   `json:"stderr_ref,omitempty"`
	ProcessStatus                     string   `json:"process_status,omitempty"`
	LastKind                          string   `json:"last_kind,omitempty"`
	ExitCode                          int      `json:"exit_code,omitempty"`
	ExitCodeKnown                     bool     `json:"exit_code_known"`
	ReadyForReadback                  bool     `json:"ready_for_readback"`
	Started                           bool     `json:"started"`
	Terminal                          bool     `json:"terminal"`
	Failed                            bool     `json:"failed"`
	Cancelled                         bool     `json:"cancelled"`
	TimedOut                          bool     `json:"timed_out"`
	ProcessCount                      int      `json:"process_count"`
	ActiveCount                       int      `json:"active_count,omitempty"`
	TerminalCount                     int      `json:"terminal_count,omitempty"`
	HostProcessEventCount             int      `json:"host_process_event_count,omitempty"`
	ViewReady                         bool     `json:"view_ready"`
	ProgressReady                     bool     `json:"progress_ready"`
	ConsumesHostProcessSessionView    bool     `json:"consumes_host_process_session_view"`
	ReadsToolOutput                   bool     `json:"reads_tool_output"`
	BuildsProcessMapFromToolOutput    bool     `json:"builds_process_map_from_tool_output"`
	RunstoreProtocolAuthorizesProcess bool     `json:"runstore_protocol_authorizes_process"`
	ProcessLifecycleControlsExecution bool     `json:"process_lifecycle_controls_execution"`
	MissingInputs                     []string `json:"missing_inputs,omitempty"`
	BlockedReasons                    []string `json:"blocked_reasons,omitempty"`
	Boundaries                        []string `json:"boundaries,omitempty"`
	NextHostAction                    string   `json:"next_host_action,omitempty"`
	RawOutputLoaded                   bool     `json:"raw_output_loaded"`
}

// HostDiagnosticOperatorLineObservationInput is one typed operator line input.
type HostDiagnosticOperatorLineObservationInput struct {
	Source              string   `json:"source,omitempty"`
	Key                 string   `json:"key,omitempty"`
	Available           bool     `json:"available"`
	Status              string   `json:"status,omitempty"`
	OperatorDisplayLine string   `json:"operator_display_line,omitempty"`
	MissingInputs       []string `json:"missing_inputs,omitempty"`
	BlockedReasons      []string `json:"blocked_reasons,omitempty"`
	Boundaries          []string `json:"boundaries,omitempty"`
	NextHostAction      string   `json:"next_host_action,omitempty"`
}

// HostDiagnosticOperatorLineObservation is one normalized operator line.
type HostDiagnosticOperatorLineObservation struct {
	Source              string   `json:"source,omitempty"`
	Key                 string   `json:"key,omitempty"`
	Available           bool     `json:"available"`
	Status              string   `json:"status,omitempty"`
	OperatorDisplayLine string   `json:"operator_display_line,omitempty"`
	MissingInputs       []string `json:"missing_inputs,omitempty"`
	BlockedReasons      []string `json:"blocked_reasons,omitempty"`
	Boundaries          []string `json:"boundaries,omitempty"`
	NextHostAction      string   `json:"next_host_action,omitempty"`
}

// SessionObservationInput is typed session evidence supplied by a host.
type SessionObservationInput struct {
	SessionID        string
	CreatedUnixMs    int64
	Events           []SessionEventObservationInput
	Branches         []SessionBranchObservationInput
	Compaction       SessionCompactionObservationInput
	LatestSummary    string
	SummaryVersion   int
	SummaryUpdatedAt int64
}

// SessionEventObservationInput is one typed session event summary input.
type SessionEventObservationInput struct {
	Role          string
	Content       string
	ToolCallID    string
	ToolCallCount int
}

// SessionBranchObservationInput is one typed branch transition input.
type SessionBranchObservationInput struct {
	BranchID   string
	NodeExecID string
	NodeID     string
	Status     string
	StartedAt  int64
	FinishedAt int64
}

// SessionCompactionObservationInput contains transcript compaction counters.
type SessionCompactionObservationInput struct {
	Passes                    int
	StrictProvider            bool
	SynthesizedToolCallIDs    int
	RecoveredToolResults      int
	DowngradedToolResults     int
	StrippedReasoningMsgs     int
	MergedMessages            int
	CompactedToolOutputs      int
	CompactedHistoryBodies    int
	ProtocolAwareHistoryDrops int
}

// SessionObservation is a normalized session observation.
type SessionObservation struct {
	Source                 string                        `json:"source,omitempty"`
	SessionID              string                        `json:"session_id,omitempty"`
	CreatedUnixMs          int64                         `json:"created_unix_ms,omitempty"`
	Labels                 []string                      `json:"labels,omitempty"`
	EventCount             int                           `json:"event_count,omitempty"`
	UserMessageCount       int                           `json:"user_message_count,omitempty"`
	AssistantMessageCount  int                           `json:"assistant_message_count,omitempty"`
	ToolResultCount        int                           `json:"tool_result_count,omitempty"`
	ToolCallMessageCount   int                           `json:"tool_call_message_count,omitempty"`
	LatestUserPreview      string                        `json:"latest_user_preview,omitempty"`
	LatestAssistantPreview string                        `json:"latest_assistant_preview,omitempty"`
	LatestToolPreview      string                        `json:"latest_tool_preview,omitempty"`
	LatestSummary          string                        `json:"latest_summary,omitempty"`
	SummaryVersion         int                           `json:"summary_version,omitempty"`
	SummaryUpdatedAt       int64                         `json:"summary_updated_at,omitempty"`
	BranchCount            int                           `json:"branch_count,omitempty"`
	Branches               []SessionBranchObservation    `json:"branches,omitempty"`
	Compaction             *SessionCompactionObservation `json:"compaction,omitempty"`
}

// SessionBranchObservation is a normalized branch summary.
type SessionBranchObservation struct {
	BranchID       string `json:"branch_id,omitempty"`
	NodeCount      int    `json:"node_count,omitempty"`
	LastNodeExecID string `json:"last_node_exec_id,omitempty"`
	LastNodeID     string `json:"last_node_id,omitempty"`
	LastStatus     string `json:"last_status,omitempty"`
	StartedAt      int64  `json:"started_at,omitempty"`
	FinishedAt     int64  `json:"finished_at,omitempty"`
	Terminal       bool   `json:"terminal,omitempty"`
}

// SessionCompactionObservation is a normalized compaction summary.
type SessionCompactionObservation struct {
	Source                    string `json:"source,omitempty"`
	Sanitized                 bool   `json:"sanitized,omitempty"`
	Applied                   bool   `json:"applied,omitempty"`
	Passes                    int    `json:"passes,omitempty"`
	StrictProvider            bool   `json:"strict_provider,omitempty"`
	SynthesizedToolCallIDs    int    `json:"synthesized_tool_call_ids,omitempty"`
	RecoveredToolResults      int    `json:"recovered_tool_results,omitempty"`
	DowngradedToolResults     int    `json:"downgraded_tool_results,omitempty"`
	StrippedReasoningMsgs     int    `json:"stripped_reasoning_msgs,omitempty"`
	MergedMessages            int    `json:"merged_messages,omitempty"`
	CompactedToolOutputs      int    `json:"compacted_tool_outputs,omitempty"`
	CompactedHistoryBodies    int    `json:"compacted_history_bodies,omitempty"`
	ProtocolAwareHistoryDrops int    `json:"protocol_aware_history_drops,omitempty"`
}
