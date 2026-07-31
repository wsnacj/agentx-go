# runtime/workflow/composition API

导入路径：

```go
import composition "github.com/wsnacj/agentx-go/runtime/workflow/composition"
```

成熟度：**Experimental / private validation**。

该 package 是 canonical Workflow Runtime composition owner，负责：

- 在 construction 时验证 lowering/orchestration 必需依赖；
- 每次 Run 先执行 canonical lowering，再执行 canonical orchestration；
- exact-empty runtime Workflow ID回退到 Spec ID；
- 同时返回 lowering plan和 partial/full execution result；
- 保留 validator、mapper、executor、journal等 error的 `errors.Is/As`
  identity。

它不拥有 validation/mapping/default policy、具体 executor、RunStore backend、
UUID/clock实现、error display policy、provider、credential 或 Scene。

## Dependencies 与 New

```go
type Dependencies struct {
    Lowering      lowering.Dependencies
    Orchestration orchestration.Dependencies
}

func New(Dependencies) (*Runtime, error)
```

以下依赖必须非 nil，并按表中顺序 fail closed：

```text
workflow composition: lowering validator is required
workflow composition: lowering mapper is required
workflow composition: journal is required
workflow composition: node execution is required
workflow composition: node execution id generator is required
workflow composition: clock is required
```

`ProjectError` 继续是 orchestration 的可选 host policy；composition 不改变其
fallback。

## Run

```go
type Inputs struct {
    WorkflowID string
    Run        orchestration.Inputs
}

type Result struct {
    LoweringPlan lowering.Plan
    Execution    orchestration.Result
}

func (*Runtime) Run(
    context.Context,
    workflow.Spec,
    Inputs,
) (Result, error)
```

顺序固定为：

```text
lowering.LowerSpec
resolve WorkflowID
lowering.Plan.OrchestrationPlan
orchestration.Run
```

`Inputs.WorkflowID` exact-empty时使用 `Spec.ID`；其它非空值原样保留，不 trim、
normalize或改写。

lowering失败时返回零 `Result` 并原样保留 error identity。orchestration开始后
失败时，返回已经生成的 `LoweringPlan` 和 partial `Execution`，同时返回原 error。

## 并发与生命周期

`Runtime` 只保存 construction 时传入的依赖，不持有 run state、不创建 goroutine，
也不提供 Shutdown。不同 Run 的 state/result相互独立；并发调用是否安全取决于
host注入的 Validator、Mapper、Journal Port、NodeExecution和生成器是否并发安全。
package 不为这些依赖加锁或自动重试。

## 非目标

- 不提供根 `agentxruntime.New` 或完整 embedded Agent SDK；
- 不选择 tool/model/task、queue、evaluator、provider或 credential；
- 不实现具体 RunStore、filesystem、network或 Scene side effect；
- 不拥有 retry/resume/long-task/product lifecycle policy；
- 不构成 Public、Beta、Stable或 production-ready声明。
