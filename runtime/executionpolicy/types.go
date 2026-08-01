// Package executionpolicy定义 substrate-neutral 的执行身份、可见性、预算、循环、
// approval、replay、sandbox和 evidence policy DTO，以及 Host编译 port。
//
// 本 package只拥有合同和 JSON shape，不执行授权、审批、sandbox或真实副作用。
package executionpolicy

import "context"

type SideEffectClass string

const (
	SideEffectReadOnly       SideEffectClass = "read_only"
	SideEffectWorkspaceWrite SideEffectClass = "workspace_write"
	SideEffectLocalProcess   SideEffectClass = "local_process"
	SideEffectBrowserMutate  SideEffectClass = "browser_mutate"
	SideEffectNetworkWrite   SideEffectClass = "network_write"
	SideEffectExternalCommit SideEffectClass = "external_commit"
)

type Identity struct {
	Profile      string `json:"profile,omitempty"`
	ProductShell string `json:"product_shell,omitempty"`
	ModelConfig  string `json:"model_config,omitempty"`
	Environment  string `json:"environment,omitempty"`
	Pack         string `json:"pack,omitempty"`
	WorkflowID   string `json:"workflow_id,omitempty"`
	WorkflowVer  string `json:"workflow_ver,omitempty"`
	CaseType     string `json:"case_type,omitempty"`
	CaseID       string `json:"case_id,omitempty"`
	RunID        string `json:"run_id,omitempty"`
	NodeID       string `json:"node_id,omitempty"`
	ParentRunID  string `json:"parent_run_id,omitempty"`
	ParentNodeID string `json:"parent_node_id,omitempty"`
	Tenant       string `json:"tenant,omitempty"`
	Trigger      string `json:"trigger,omitempty"`
}

type VisibilityPolicy struct {
	AllowTools      []string `json:"allow_tools,omitempty"`
	DenyTools       []string `json:"deny_tools,omitempty"`
	DeclaredTools   []string `json:"declared_tools,omitempty"`
	RequireDeclared bool     `json:"require_declared,omitempty"`
	MaxRisk         string   `json:"max_risk,omitempty"`
}

type ToolCallDedupeMode string

const (
	ToolCallDedupeUnspecified ToolCallDedupeMode = ""
	ToolCallDedupeDisabled    ToolCallDedupeMode = "disabled"
	ToolCallDedupeEnabled     ToolCallDedupeMode = "enabled"
)

type BudgetPolicy struct {
	Enabled            bool               `json:"enabled,omitempty"`
	MaxToolCalls       int                `json:"max_tool_calls,omitempty"`
	MaxDurationMs      int64              `json:"max_duration_ms,omitempty"`
	MaxInputTokens     int64              `json:"max_input_tokens,omitempty"`
	MaxOutputTokens    int64              `json:"max_output_tokens,omitempty"`
	MaxCostMicrosUSD   int64              `json:"max_cost_micros_usd,omitempty"`
	MaxToolTimeoutMs   int                `json:"max_tool_timeout_ms,omitempty"`
	LLMTaskTimeoutMs   int                `json:"llm_task_timeout_ms,omitempty"`
	MaxToolResultChars int                `json:"max_tool_result_chars,omitempty"`
	ToolCallDedupe     ToolCallDedupeMode `json:"tool_call_dedupe,omitempty"`
	MaxParallelism     int                `json:"max_parallelism,omitempty"`
	MaxChildRuns       int                `json:"max_child_runs,omitempty"`
	MaxRetries         int                `json:"max_retries,omitempty"`
	MaxRunRetries      int                `json:"max_run_retries,omitempty"`
}

type LoopPolicy struct {
	MaxRounds                int  `json:"max_rounds,omitempty"`
	ContinueOnError          bool `json:"continue_on_error,omitempty"`
	LoopDetectionEnabled     bool `json:"loop_detection_enabled,omitempty"`
	LoopRepeatThreshold      int  `json:"loop_repeat_threshold,omitempty"`
	LoopPingPongThreshold    int  `json:"loop_ping_pong_threshold,omitempty"`
	LoopNoProgressThreshold  int  `json:"loop_no_progress_threshold,omitempty"`
	ToolFailureFuseEnabled   bool `json:"tool_failure_fuse_enabled,omitempty"`
	ToolFailureFuseThreshold int  `json:"tool_failure_fuse_threshold,omitempty"`
}

type ApprovalPolicy struct {
	Mode       string   `json:"mode,omitempty"`
	AllowTools []string `json:"allow_tools,omitempty"`
	DenyTools  []string `json:"deny_tools,omitempty"`
	MaxRisk    string   `json:"max_risk,omitempty"`
}

type ReplayPolicy struct {
	AllowTools             []string            `json:"allow_tools,omitempty"`
	DenyTools              []string            `json:"deny_tools,omitempty"`
	MaxRisk                string              `json:"max_risk,omitempty"`
	AllowIdempotencyLevels []string            `json:"allow_idempotency_levels,omitempty"`
	AllowIdempotencyByEnv  map[string][]string `json:"allow_idempotency_by_env,omitempty"`
	ApprovalMode           string              `json:"approval_mode,omitempty"`
	ApprovalAllowTools     []string            `json:"approval_allow_tools,omitempty"`
	ApprovalDenyTools      []string            `json:"approval_deny_tools,omitempty"`
	ApprovalMaxRisk        string              `json:"approval_max_risk,omitempty"`
	AutoAllowTools         []string            `json:"auto_allow_tools,omitempty"`
	AutoAllowByEnv         map[string][]string `json:"auto_allow_by_env,omitempty"`
	SkipExecuted           bool                `json:"skip_executed,omitempty"`
}

type RuntimeControlMode string

const (
	RuntimeControlModeOff     RuntimeControlMode = "off"
	RuntimeControlModeEnabled RuntimeControlMode = "enabled"

	ControlPlaneRetryRuntimeGateRef = "contract:control_plane.retry_runtime"
)

type RuntimeControlPolicy struct {
	ControlPlane ControlPlaneRuntimeControlPolicy `json:"control_plane,omitempty"`
}

type ControlPlaneRuntimeControlPolicy struct {
	RetryRuntime RuntimeControlMode `json:"retry_runtime,omitempty"`
}

type SideEffectPolicy struct {
	MaxClass                   SideEffectClass `json:"max_class,omitempty"`
	StrictRecovery             bool            `json:"strict_recovery,omitempty"`
	CrossSystemConfirm         bool            `json:"cross_system_confirm,omitempty"`
	CrossSystemConfirmTools    []string        `json:"cross_system_confirm_tools,omitempty"`
	CrossSystemConfirmTokenKey string          `json:"cross_system_confirm_token_key,omitempty"`
}

type SandboxPolicy struct {
	ExecAllowlist                 []string `json:"exec_allowlist,omitempty"`
	ExecDenyPatterns              []string `json:"exec_deny_patterns,omitempty"`
	ProcessSignals                []string `json:"process_signals,omitempty"`
	WebSearchAllowPrivateHosts    bool     `json:"web_search_allow_private_hosts,omitempty"`
	WebSearchTrustedEnvProxy      bool     `json:"web_search_trusted_env_proxy,omitempty"`
	WebSearchAllowCIDRs           []string `json:"web_search_allow_cidrs,omitempty"`
	WebSearchDenyCIDRs            []string `json:"web_search_deny_cidrs,omitempty"`
	WebSearchAllowPorts           []int    `json:"web_search_allow_ports,omitempty"`
	WebSearchDenyPorts            []int    `json:"web_search_deny_ports,omitempty"`
	WebFetchAllowPrivateHosts     bool     `json:"web_fetch_allow_private_hosts,omitempty"`
	WebFetchTrustedEnvProxy       bool     `json:"web_fetch_trusted_env_proxy,omitempty"`
	WebFetchAllowCIDRs            []string `json:"web_fetch_allow_cidrs,omitempty"`
	WebFetchDenyCIDRs             []string `json:"web_fetch_deny_cidrs,omitempty"`
	WebFetchAllowPorts            []int    `json:"web_fetch_allow_ports,omitempty"`
	WebFetchDenyPorts             []int    `json:"web_fetch_deny_ports,omitempty"`
	HTTPRequestAllowPrivateHosts  bool     `json:"http_request_allow_private_hosts,omitempty"`
	HTTPRequestTrustedEnvProxy    bool     `json:"http_request_trusted_env_proxy,omitempty"`
	HTTPRequestAllowCIDRs         []string `json:"http_request_allow_cidrs,omitempty"`
	HTTPRequestDenyCIDRs          []string `json:"http_request_deny_cidrs,omitempty"`
	HTTPRequestAllowPorts         []int    `json:"http_request_allow_ports,omitempty"`
	HTTPRequestDenyPorts          []int    `json:"http_request_deny_ports,omitempty"`
	BrowserProxyAllowPrivateHosts bool     `json:"browser_proxy_allow_private_hosts,omitempty"`
	BrowserProxyTrustedEnvProxy   bool     `json:"browser_proxy_trusted_env_proxy,omitempty"`
	BrowserProxyAllowCIDRs        []string `json:"browser_proxy_allow_cidrs,omitempty"`
	BrowserProxyDenyCIDRs         []string `json:"browser_proxy_deny_cidrs,omitempty"`
	BrowserProxyAllowPorts        []int    `json:"browser_proxy_allow_ports,omitempty"`
	BrowserProxyDenyPorts         []int    `json:"browser_proxy_deny_ports,omitempty"`
	BrowserProxyActKinds          []string `json:"browser_proxy_act_kinds,omitempty"`
	NodesGatewayAllowPrivateHosts bool     `json:"nodes_gateway_allow_private_hosts,omitempty"`
	NodesGatewayTrustedEnvProxy   bool     `json:"nodes_gateway_trusted_env_proxy,omitempty"`
	NodesGatewayAllowCIDRs        []string `json:"nodes_gateway_allow_cidrs,omitempty"`
	NodesGatewayDenyCIDRs         []string `json:"nodes_gateway_deny_cidrs,omitempty"`
	NodesGatewayAllowPorts        []int    `json:"nodes_gateway_allow_ports,omitempty"`
	NodesGatewayDenyPorts         []int    `json:"nodes_gateway_deny_ports,omitempty"`
	SessionVisibilityScope        string   `json:"session_visibility_scope,omitempty"`
	SessionTreeMaxDepth           int      `json:"session_tree_max_depth,omitempty"`
	SessionLeafLimit              int      `json:"session_leaf_limit,omitempty"`
}

type EvidencePolicy struct {
	RequiredArtifacts []string `json:"required_artifacts,omitempty"`
}

type InheritancePolicy struct {
	ChildCannotEscalate bool `json:"child_cannot_escalate,omitempty"`
	BudgetMustClamp     bool `json:"budget_must_clamp,omitempty"`
}

type AuditPolicy struct {
	PersistSnapshot bool `json:"persist_snapshot,omitempty"`
}

type Contract struct {
	ID              string               `json:"id,omitempty"`
	Version         int                  `json:"version,omitempty"`
	Strict          bool                 `json:"strict,omitempty"`
	Identity        Identity             `json:"identity,omitempty"`
	Visibility      VisibilityPolicy     `json:"visibility"`
	Budget          BudgetPolicy         `json:"budget,omitempty"`
	Loop            LoopPolicy           `json:"loop,omitempty"`
	Approval        ApprovalPolicy       `json:"approval,omitempty"`
	Replay          ReplayPolicy         `json:"replay,omitempty"`
	RuntimeControls RuntimeControlPolicy `json:"runtime_controls,omitempty"`
	SideEffects     SideEffectPolicy     `json:"side_effects,omitempty"`
	Sandbox         SandboxPolicy        `json:"sandbox,omitempty"`
	Evidence        EvidencePolicy       `json:"evidence,omitempty"`
	Inherit         InheritancePolicy    `json:"inherit,omitempty"`
	Audit           AuditPolicy          `json:"audit,omitempty"`
}

type Snapshot struct {
	Contract     Contract `json:"contract"`
	SourceLayers []string `json:"source_layers,omitempty"`
	CreatedAt    int64    `json:"created_at,omitempty"`
}

type Diff struct {
	ChangedFields []string `json:"changed_fields,omitempty"`
}

type CompileInput struct {
	Identity        Identity             `json:"identity,omitempty"`
	ProfileAllow    []string             `json:"profile_allow,omitempty"`
	ProfileDeny     []string             `json:"profile_deny,omitempty"`
	ConfigAllow     []string             `json:"config_allow,omitempty"`
	ConfigDeny      []string             `json:"config_deny,omitempty"`
	DeclaredTools   []string             `json:"declared_tools,omitempty"`
	MaxRisk         string               `json:"max_risk,omitempty"`
	Strict          bool                 `json:"strict,omitempty"`
	RequireDeclared bool                 `json:"require_declared,omitempty"`
	Budget          BudgetPolicy         `json:"budget,omitempty"`
	Loop            LoopPolicy           `json:"loop,omitempty"`
	Approval        ApprovalPolicy       `json:"approval,omitempty"`
	Replay          ReplayPolicy         `json:"replay,omitempty"`
	RuntimeControls RuntimeControlPolicy `json:"runtime_controls,omitempty"`
	SideEffects     SideEffectPolicy     `json:"side_effects,omitempty"`
	Sandbox         SandboxPolicy        `json:"sandbox,omitempty"`
	Evidence        EvidencePolicy       `json:"evidence,omitempty"`
	Inherit         InheritancePolicy    `json:"inherit,omitempty"`
	Audit           AuditPolicy          `json:"audit,omitempty"`
	SourceLayers    []string             `json:"source_layers,omitempty"`
}

type Compiler interface {
	Compile(ctx context.Context, input CompileInput) (Snapshot, error)
}
