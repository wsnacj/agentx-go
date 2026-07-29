package protocol

const (
	RuntimeSchemaV1           = "agentx.runtime.v1"
	RunEventSchemaV1          = "agentx.run_event.v1"
	TraceSpanSchemaV1         = "agentx.trace_span.v1"
	ToolExecutionPlanSchemaV1 = "agentx.tool_execution_plan.v1"
	HandoffSchemaV1           = "agentx.handoff.v1"
	SandboxManifestSchemaV1   = "agentx.sandbox_manifest.v1"
	ArtifactVersionSchemaV1   = "agentx.artifact_version.v1"
	ArtifactLinkSchemaV1      = "agentx.artifact_link.v1"

	KindToolPlanCreated      = "tool.plan.created"
	KindSandboxManifest      = "sandbox.manifest.resolved"
	KindHandoffRequested     = "handoff.requested"
	KindHandoffAccepted      = "handoff.accepted"
	KindHandoffRejected      = "handoff.rejected"
	KindHandoffStarted       = "handoff.started"
	KindHandoffCompleted     = "handoff.completed"
	KindHandoffFailed        = "handoff.failed"
	KindHandoffCollected     = "handoff.collected"
	KindHostProcessStarted   = "host.process.started"
	KindHostProcessCompleted = "host.process.completed"
	KindHostProcessFailed    = "host.process.failed"
	KindHostProcessCancelled = "host.process.cancelled"
	KindHostProcessTimedOut  = "host.process.timed_out"
	KindHostProcessReadback  = "host.process.readback"
	StatusPlanned            = "planned"
	StatusBlocked            = "blocked"
	StatusApprovalPending    = "approval_pending"
	StatusReady              = "ready"
	StatusRunning            = "running"
	StatusCompleted          = "completed"
	StatusFailed             = "failed"
	StatusSkipped            = "skipped"
	HandoffKindContextOnly   = "context_only"
	HandoffKindArtifact      = "artifact_transfer"
	HandoffKindTaskChild     = "task_child"
	HandoffKindAgentAsTool   = "agent_as_tool"
	HandoffKindWorkflowNode  = "workflow_node"
	HandoffKindExternalHost  = "external_host"
)

type Envelope struct {
	SchemaVersion      string `json:"schema_version"`
	Kind               string `json:"kind,omitempty"`
	TimestampUnixMilli int64  `json:"timestamp_unix_milli,omitempty"`
	RunID              string `json:"run_id,omitempty"`
	RootRunID          string `json:"root_run_id,omitempty"`
	ParentRunID        string `json:"parent_run_id,omitempty"`
	SessionID          string `json:"session_id,omitempty"`
	TurnID             string `json:"turn_id,omitempty"`
	BranchID           string `json:"branch_id,omitempty"`
	NodeExecID         string `json:"node_exec_id,omitempty"`
	TraceID            string `json:"trace_id,omitempty"`
	SpanID             string `json:"span_id,omitempty"`
	ParentSpanID       string `json:"parent_span_id,omitempty"`
}

type ErrorInfo struct {
	Class     string `json:"error_class,omitempty"`
	Code      string `json:"error_code,omitempty"`
	Reason    string `json:"reason,omitempty"`
	Message   string `json:"message,omitempty"`
	Retryable bool   `json:"retryable,omitempty"`
	Degraded  bool   `json:"degraded,omitempty"`
}

type RuntimeDecisionSnapshot struct {
	Action             string `json:"action,omitempty"`
	Reason             string `json:"reason,omitempty"`
	Detail             string `json:"detail,omitempty"`
	DecisionSubject    string `json:"decision_subject,omitempty"`
	TargetKind         string `json:"target_kind,omitempty"`
	Checked            bool   `json:"checked,omitempty"`
	Allowed            bool   `json:"allowed,omitempty"`
	Denied             bool   `json:"denied,omitempty"`
	RequiresConfirm    bool   `json:"requires_confirm,omitempty"`
	Degraded           bool   `json:"degraded,omitempty"`
	PolicySource       string `json:"policy_source,omitempty"`
	ControlSource      string `json:"control_source,omitempty"`
	EnforcementSurface string `json:"enforcement_surface,omitempty"`
}

type RuntimeDecisionSummary struct {
	Checked  int `json:"checked,omitempty"`
	Allowed  int `json:"allowed,omitempty"`
	Denied   int `json:"denied,omitempty"`
	Degraded int `json:"degraded,omitempty"`
}

type Usage struct {
	ModelID          string  `json:"model_id,omitempty"`
	Provider         string  `json:"provider,omitempty"`
	InputTokens      int64   `json:"input_tokens,omitempty"`
	OutputTokens     int64   `json:"output_tokens,omitempty"`
	CacheReadTokens  int64   `json:"cache_read_tokens,omitempty"`
	CacheWriteTokens int64   `json:"cache_write_tokens,omitempty"`
	CostUSD          float64 `json:"cost_usd,omitempty"`
	PricingKnown     bool    `json:"pricing_known,omitempty"`
}

type RunEvent struct {
	Envelope
	SourceEvent           string                   `json:"source_event,omitempty"`
	SourceEventID         string                   `json:"source_event_id,omitempty"`
	Level                 string                   `json:"level,omitempty"`
	Status                string                   `json:"status,omitempty"`
	Reason                string                   `json:"reason,omitempty"`
	Stage                 string                   `json:"stage,omitempty"`
	ToolName              string                   `json:"tool_name,omitempty"`
	ToolCallID            string                   `json:"tool_call_id,omitempty"`
	ModelID               string                   `json:"model_id,omitempty"`
	Provider              string                   `json:"provider,omitempty"`
	DurationMs            int64                    `json:"duration_ms,omitempty"`
	Attempt               int                      `json:"attempt,omitempty"`
	Cached                bool                     `json:"cached,omitempty"`
	RetryCount            int                      `json:"retry_count,omitempty"`
	Error                 *ErrorInfo               `json:"error,omitempty"`
	Usage                 *Usage                   `json:"usage,omitempty"`
	RuntimeDecision       *RuntimeDecisionSnapshot `json:"runtime_decision,omitempty"`
	ExecutionContractID   string                   `json:"execution_contract_id,omitempty"`
	ExecutionContractDiff []string                 `json:"execution_contract_diff,omitempty"`
	Attrs                 map[string]any           `json:"attrs,omitempty"`
}

// TraceSpan is schema-ready but reserved in runtime protocol v1. The protocol
// package can normalize, validate, and marshal spans, but v1 does not define a
// producer, projection, or consumer contract for materializing span trees.
type TraceSpan struct {
	Envelope
	Type               string         `json:"type,omitempty"`
	Name               string         `json:"name,omitempty"`
	Status             string         `json:"status,omitempty"`
	StartedAtUnixMilli int64          `json:"started_at_unix_milli,omitempty"`
	EndedAtUnixMilli   int64          `json:"ended_at_unix_milli,omitempty"`
	DurationMs         int64          `json:"duration_ms,omitempty"`
	Usage              *Usage         `json:"usage,omitempty"`
	Error              *ErrorInfo     `json:"error,omitempty"`
	Attrs              map[string]any `json:"attrs,omitempty"`
}

type ToolExecutionPlan struct {
	Envelope
	PlanID                 string                 `json:"plan_id,omitempty"`
	MaxConcurrency         int                    `json:"max_concurrency,omitempty"`
	ExecutionContractID    string                 `json:"execution_contract_id,omitempty"`
	ExecutionContractDiff  []string               `json:"execution_contract_diff,omitempty"`
	Calls                  []ToolPlanCall         `json:"calls,omitempty"`
	Interruptions          []ToolPlanInterruption `json:"interruptions,omitempty"`
	BlockedCalls           []ToolPlanBlockedCall  `json:"blocked_calls,omitempty"`
	RuntimeDecisionSummary RuntimeDecisionSummary `json:"runtime_decision_summary,omitempty"`
}

type ToolPlanCall struct {
	ToolCallID         string                   `json:"tool_call_id,omitempty"`
	ToolName           string                   `json:"tool_name,omitempty"`
	ArgumentsHash      string                   `json:"arguments_hash,omitempty"`
	ArgumentsSummary   string                   `json:"arguments_summary,omitempty"`
	Origin             string                   `json:"origin,omitempty"`
	Category           string                   `json:"category,omitempty"`
	ExecutionMode      string                   `json:"execution_mode,omitempty"`
	RequiresApproval   bool                     `json:"requires_approval,omitempty"`
	RequiresSandbox    bool                     `json:"requires_sandbox,omitempty"`
	IdempotencyKey     string                   `json:"idempotency_key,omitempty"`
	ExpectedSideEffect string                   `json:"expected_side_effect,omitempty"`
	Status             string                   `json:"status,omitempty"`
	Reason             string                   `json:"reason,omitempty"`
	ErrorCode          string                   `json:"error_code,omitempty"`
	RuntimeDecision    *RuntimeDecisionSnapshot `json:"runtime_decision,omitempty"`
}

type ToolPlanInterruption struct {
	InterruptionID string `json:"interruption_id,omitempty"`
	ToolCallID     string `json:"tool_call_id,omitempty"`
	Type           string `json:"type,omitempty"`
	ApprovalID     string `json:"approval_id,omitempty"`
	Status         string `json:"status,omitempty"`
	Reason         string `json:"reason,omitempty"`
}

type ToolPlanBlockedCall struct {
	ToolCallID string `json:"tool_call_id,omitempty"`
	ToolName   string `json:"tool_name,omitempty"`
	Reason     string `json:"reason,omitempty"`
	ErrorCode  string `json:"error_code,omitempty"`
}

type HandoffRecord struct {
	Envelope
	HandoffID   string             `json:"handoff_id,omitempty"`
	HandoffKind string             `json:"handoff_kind,omitempty"`
	Source      HandoffEndpoint    `json:"source,omitempty"`
	Target      HandoffEndpoint    `json:"target,omitempty"`
	InputFilter HandoffInputFilter `json:"input_filter,omitempty"`
	Isolation   HandoffIsolation   `json:"isolation,omitempty"`
	Status      string             `json:"status,omitempty"`
	Reason      string             `json:"reason,omitempty"`
	Error       *ErrorInfo         `json:"error,omitempty"`
	Attrs       map[string]any     `json:"attrs,omitempty"`
}

type HandoffEndpoint struct {
	AgentID       string `json:"agent_id,omitempty"`
	PackID        string `json:"pack_id,omitempty"`
	WorkflowID    string `json:"workflow_id,omitempty"`
	NodeID        string `json:"node_id,omitempty"`
	RunID         string `json:"run_id,omitempty"`
	ExpectedRunID string `json:"expected_run_id,omitempty"`
	SessionID     string `json:"session_id,omitempty"`
	TaskID        string `json:"task_id,omitempty"`
	ToolCallID    string `json:"tool_call_id,omitempty"`
}

type HandoffInputFilter struct {
	Mode                string   `json:"mode,omitempty"`
	IncludedArtifactIDs []string `json:"included_artifact_ids,omitempty"`
	IncludedMessageIDs  []string `json:"included_message_ids,omitempty"`
	Redacted            bool     `json:"redacted,omitempty"`
	Reason              string   `json:"reason,omitempty"`
}

type HandoffIsolation struct {
	Scope              string `json:"scope,omitempty"`
	Branch             string `json:"branch,omitempty"`
	PeerHistoryVisible bool   `json:"peer_history_visible"`
}

type SandboxManifest struct {
	Envelope
	ManifestID    string               `json:"manifest_id,omitempty"`
	Root          string               `json:"root,omitempty"`
	Backend       string               `json:"backend,omitempty"`
	Platform      SandboxPlatform      `json:"platform,omitempty"`
	Entries       []SandboxEntry       `json:"entries,omitempty"`
	Environment   []SandboxEnvVar      `json:"environment,omitempty"`
	PathGrants    []SandboxPathGrant   `json:"path_grants,omitempty"`
	Network       SandboxNetworkPolicy `json:"network,omitempty"`
	CommandPolicy SandboxCommandPolicy `json:"command_policy,omitempty"`
	Degraded      bool                 `json:"degraded,omitempty"`
	Reason        string               `json:"reason,omitempty"`
}

type SandboxPlatform struct {
	GOOS string `json:"goos,omitempty"`
	Arch string `json:"arch,omitempty"`
}

type SandboxEntry struct {
	Path      string `json:"path,omitempty"`
	Kind      string `json:"kind,omitempty"`
	Source    string `json:"source,omitempty"`
	Mode      string `json:"mode,omitempty"`
	Ephemeral bool   `json:"ephemeral,omitempty"`
}

type SandboxEnvVar struct {
	Name        string `json:"name,omitempty"`
	Source      string `json:"source,omitempty"`
	Description string `json:"description,omitempty"`
	Ephemeral   bool   `json:"ephemeral,omitempty"`
	Present     bool   `json:"present,omitempty"`
}

type SandboxPathGrant struct {
	Path   string `json:"path,omitempty"`
	Mode   string `json:"mode,omitempty"`
	Reason string `json:"reason,omitempty"`
}

type SandboxNetworkPolicy struct {
	Mode  string   `json:"mode,omitempty"`
	Allow []string `json:"allow,omitempty"`
	Deny  []string `json:"deny,omitempty"`
}

type SandboxCommandPolicy struct {
	Allow []string `json:"allow,omitempty"`
	Deny  []string `json:"deny,omitempty"`
}

type ArtifactVersion struct {
	Envelope
	ArtifactID         string            `json:"artifact_id,omitempty"`
	Version            int               `json:"version,omitempty"`
	CanonicalURI       string            `json:"canonical_uri,omitempty"`
	Scope              string            `json:"scope,omitempty"`
	MIMEType           string            `json:"mime_type,omitempty"`
	CreatedAtUnixMilli int64             `json:"created_at_unix_milli,omitempty"`
	CreatedBy          ArtifactCreatedBy `json:"created_by,omitempty"`
	Metadata           map[string]any    `json:"metadata,omitempty"`
	Payload            ArtifactPayload   `json:"payload,omitempty"`
}

type ArtifactCreatedBy struct {
	ToolName   string `json:"tool_name,omitempty"`
	ToolCallID string `json:"tool_call_id,omitempty"`
	AgentID    string `json:"agent_id,omitempty"`
	Producer   string `json:"producer,omitempty"`
}

type ArtifactPayload struct {
	Storage    string `json:"storage,omitempty"`
	BlobRef    string `json:"blob_ref,omitempty"`
	Path       string `json:"path,omitempty"`
	URL        string `json:"url,omitempty"`
	Digest     string `json:"digest,omitempty"`
	Bytes      int64  `json:"bytes,omitempty"`
	ExternalID string `json:"external_id,omitempty"`
}

type ArtifactLink struct {
	SchemaVersion      string  `json:"schema_version"`
	SourceArtifactID   string  `json:"source_artifact_id,omitempty"`
	SourceVersion      int     `json:"source_version,omitempty"`
	TargetArtifactID   string  `json:"target_artifact_id,omitempty"`
	TargetVersion      int     `json:"target_version,omitempty"`
	Relation           string  `json:"relation,omitempty"`
	Confidence         float64 `json:"confidence,omitempty"`
	CreatedAtUnixMilli int64   `json:"created_at_unix_milli,omitempty"`
}
