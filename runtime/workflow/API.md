# runtime/workflow API

导入路径：

```go
import workflow "github.com/wsnacj/agentx-go/runtime/workflow"
```

成熟度：**Experimental / private validation**。

该 package 只定义 AgentX Workflow Spec 的数据合同：planning mode、node kind、
execution mode、nodes、edges、state/artifact/evaluator schema。它不提供
validation、lowering、executor、replanning、RunStore 或 provider。

## Workflow Spec

```go
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
```

`Spec` 是值数据，不读取文件、环境变量或 registry。`Version` 是调用方的
Workflow 版本字符串，不等于 AgentX module semver，也不表示 Stable v1。

## Planning mode

```go
type PlanningMode string

const (
    PlanningOpen     PlanningMode = "open"
    PlanningBounded  PlanningMode = "bounded"
    PlanningPlanless PlanningMode = "planless"
)
```

本 package 只保存 mode，不解释 planning policy。未知值不会自动修正或报错。

## Node

```go
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
```

`Config` 是 opaque node-local 数据。本 package 不解析其中的 provider、
credential、tool 参数或业务规则；host/Scene 必须在自己的边界验证。

## Execution mode

```go
type ExecutionMode string

const (
    ExecInline ExecutionMode = "inline"
    ExecTask   ExecutionMode = "task"
    ExecRemote ExecutionMode = "remote"
)
```

mode 只描述声明值，不启动 task、remote backend 或 scheduler。

## Edge、Binding 与 Schema

```go
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
```

所有类型都保留 Go zero value、nil/empty slice/map 与现有 `omitempty` 行为。
本 package 不补默认值、不排序、不去重、不检查 edge/node 引用，也不执行 retry。

## 使用示例

```go
spec := workflow.Spec{
    ID:           "research",
    PlanningMode: workflow.PlanningBounded,
    EntryNode:    "collect",
    Nodes: []workflow.NodeSpec{{
        ID:            "collect",
        Kind:          workflow.NodeTool,
        ExecutionMode: workflow.ExecInline,
    }},
}
```

构造成功不表示 Spec 已通过验证或可执行。调用方必须交给获准的 validator/
executor adapter，并处理其 error。

## 与首版 Facade 的边界

首版 AgentX Facade 仍只提供 `Run`，Workflow 是后续能力。该 package 当前不提供：

- `ValidateSpec`、`LowerSpec`、`ExecuteInline`；
- Node executor、LLM/tool/provider wiring；
- RunStore、artifact persistence、resume 或 durable lifecycle；
- built-in delegation policy、temporary planner 或 Scene workflow；
- Public、Beta、Stable 或 production-ready 声明。
