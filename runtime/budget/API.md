# runtime/budget API

导入路径：

```go
import budget "github.com/wsnacj/agentx-go/runtime/budget"
```

成熟度：**Experimental / private validation**。

该 package 对调用方提供的预算上限和当前快照执行无副作用判定，返回是否允许
继续、当前阶段、停止原因和近限额警告。它不拥有预算值来源、execution
snapshot、计量采集、模型价格、停止动作、恢复策略或授权决策。

## 阶段与原因

```go
const (
    StageOK       = "ok"
    StageWarn     = "warn"
    StageSoftStop = "soft_stop"
    StageHardStop = "hard_stop"

    ReasonMaxToolCalls    = "max_tool_calls"
    ReasonMaxDurationMs   = "max_duration_ms"
    ReasonMaxInputTokens  = "max_input_tokens"
    ReasonMaxOutputTokens = "max_output_tokens"
    ReasonMaxCostMicros   = "max_cost_micros_usd"
)
```

工具调用次数超限返回 `soft_stop`；时长、输入 token、输出 token 或成本超限
返回 `hard_stop`。字符串值属于兼容合同，调用方不应依赖未声明的新值。

## 数据类型

```go
type Limit struct {
    MaxToolCalls     int
    MaxDurationMs    int64
    MaxInputTokens   int64
    MaxOutputTokens  int64
    MaxCostMicrosUSD int64
}

type Snapshot struct {
    ToolCalls     int
    DurationMs    int64
    InputTokens   int64
    OutputTokens  int64
    CostMicrosUSD int64
}

type Verdict struct {
    Allowed  bool
    Stage    string
    Reason   string
    Warnings []string
}
```

任一 limit 字段小于或等于零表示该维度不限制。本 package 不校验负数
snapshot，也不执行 normalization。

## Controller

```go
type Controller struct {
    // contains filtered or unexported fields
}

func NewController() Controller
func (Controller) Check(Limit, Snapshot) Verdict
```

`NewController` 与 `Controller{}` 均使用 0.8 的近限额阈值。判定规则：

- 超限使用严格 `>`；等于上限不会停止；
- 停止原因按工具调用、时长、输入 token、输出 token、成本的顺序选择；
- 任一停止判定优先于 warning；
- warning 只覆盖工具调用和时长，使用 `>= 80%` 阈值；
- warning 顺序固定为工具调用、时长；
- 正常结果为 `Allowed=true, Stage="ok"`；
- warning 结果为 `Allowed=true, Stage="warn"`；
- 停止结果不携带 warning。

warning 文本格式：

```text
budget near limit (max_tool_calls): <current>/<limit>
budget near limit (max_duration_ms): <current>/<limit>ms
```

## 请求前 Token Window Plan

```go
type TokenPlanRequest struct {
    Limits                llm.ModelLimits
    Input                 llm.TokenCount
    RequestedOutputTokens int64
    ReservedOutputTokens  int64
}

type TokenPlan struct {
    Allowed             bool
    Reason              string
    Input               llm.TokenCount
    InputLimitTokens     int64
    OutputLimitTokens    int64
    PlannedOutputTokens  int64
    ReservedOutputTokens int64
    ContextWindowTokens  int64
}

func PlanTokens(TokenPlanRequest) TokenPlan
```

`PlanTokens` 同时考虑 context window、模型输入上限、模型输出上限和显式输出预留。它不会
调用 tokenizer、provider 或价格服务，也不会把 `Input.Exact=false` 的估算改写成实际 usage。
新增稳定 reason 为 `context_window_tokens` 和 `invalid_token_plan`；输入或输出方向的既有限制
继续使用 `max_input_tokens` 与 `max_output_tokens`。

## 非目标

- 不定义 execution contract 或 snapshot 的唯一来源；
- 不采集 usage、计算 provider 价格或持久化预算；
- 不选择 tokenizer、模型 profile 或 provider；
- 不执行停止、取消、重试、恢复或 approval；
- 不构成 Public、Beta、Stable 或 production-ready 声明。
