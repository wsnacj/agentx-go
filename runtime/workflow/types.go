package workflow

type PlanningMode string

const (
	PlanningOpen     PlanningMode = "open"
	PlanningBounded  PlanningMode = "bounded"
	PlanningPlanless PlanningMode = "planless"
)

type NodeKind string

const (
	NodeTool       NodeKind = "tool"
	NodeLLM        NodeKind = "llm"
	NodeAgent      NodeKind = "agent"
	NodeParallel   NodeKind = "parallel"
	NodeCollect    NodeKind = "collect"
	NodeWait       NodeKind = "wait"
	NodeEvaluate   NodeKind = "evaluate"
	NodeApprove    NodeKind = "approve"
	NodeSubflow    NodeKind = "subflow"
	NodeHumanInput NodeKind = "human_input"
)

type ExecutionMode string

const (
	ExecInline ExecutionMode = "inline"
	ExecTask   ExecutionMode = "task"
	ExecRemote ExecutionMode = "remote"
)

type Spec struct {
	ID              string            `json:"id,omitempty"`
	Title           string            `json:"title,omitempty"`
	Description     string            `json:"description,omitempty"`
	Version         string            `json:"version,omitempty"`
	Pack            string            `json:"pack,omitempty"`
	CaseTypes       []string          `json:"case_types,omitempty"`
	RouteHints      []string          `json:"route_hints,omitempty"`
	PlanningMode    PlanningMode      `json:"planning_mode,omitempty"`
	EntryNode       string            `json:"entry_node,omitempty"`
	Nodes           []NodeSpec        `json:"nodes,omitempty"`
	Edges           []EdgeSpec        `json:"edges,omitempty"`
	StateSchema     []StateSlotSpec   `json:"state_schema,omitempty"`
	ArtifactSchema  []ArtifactTypeRef `json:"artifact_schema,omitempty"`
	EvaluatorSchema []EvaluatorRef    `json:"evaluator_schema,omitempty"`
	DefaultContract string            `json:"default_contract,omitempty"`
}

type NodeSpec struct {
	ID            string         `json:"id,omitempty"`
	Kind          NodeKind       `json:"kind,omitempty"`
	Title         string         `json:"title,omitempty"`
	Description   string         `json:"description,omitempty"`
	ContractRef   string         `json:"contract_ref,omitempty"`
	Inputs        []BindingSpec  `json:"inputs,omitempty"`
	Outputs       []BindingSpec  `json:"outputs,omitempty"`
	Retry         RetryPolicy    `json:"retry,omitempty"`
	TimeoutMs     int64          `json:"timeout_ms,omitempty"`
	ExecutionMode ExecutionMode  `json:"execution_mode,omitempty"`
	Config        map[string]any `json:"config,omitempty"`
}

type EdgeSpec struct {
	From      string `json:"from,omitempty"`
	To        string `json:"to,omitempty"`
	On        string `json:"on,omitempty"`
	Condition string `json:"condition,omitempty"`
}

type StateSlotSpec struct {
	Name     string `json:"name,omitempty"`
	Type     string `json:"type,omitempty"`
	Required bool   `json:"required,omitempty"`
}

type BindingSpec struct {
	From     string `json:"from,omitempty"`
	To       string `json:"to,omitempty"`
	Optional bool   `json:"optional,omitempty"`
}

type RetryPolicy struct {
	MaxAttempts int   `json:"max_attempts,omitempty"`
	BackoffMs   []int `json:"backoff_ms,omitempty"`
}

type ArtifactTypeRef struct {
	Type string `json:"type,omitempty"`
}

type EvaluatorRef struct {
	Name string `json:"name,omitempty"`
}
