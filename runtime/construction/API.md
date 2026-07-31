# `runtime/construction` 中文 API Reference

状态：**Experimental / private validation**。

本 package拥有底层无关的 Runtime 构造生命周期：校验输入，依次取得 model
runtime、runner runtime、execution adapter和根 `agentx.Client`，并在任一阶段
失败时按 ownership逆序执行有界清理。它不提供模型、Runner、provider或 backend。

## `Config`

```go
type Config struct {
    WorkspaceRoot string
    ModelProfile  string
    Profile       agentx.ExecutionProfile
}
```

- `WorkspaceRoot` 必填，trim/clean 后必须是当前平台的绝对路径；package不会创建、
  修改或删除该目录。
- `ModelProfile` 必填，只接受 host catalog中的逻辑名称；路径、URL和配置片段
  形态会在取得资源前被拒绝。
- `Profile` 使用根 `agentx` 合同校验；当前只接受零值或唯一已支持画像。

## `ModelRuntime` 与 `RunnerRuntime`

两个接口都只暴露：

```go
Shutdown(context.Context) error
```

它们是 construction owner在阶段间传递的 opaque lifecycle resource。实现可以
是具体模型服务或 Runner，但这些类型不会进入 canonical API。

## `Host`

```go
type Host interface {
    ResolveModel(context.Context, Config) (ModelRuntime, error)
    NewRunner(context.Context, Config, ModelRuntime) (RunnerRuntime, error)
    NewAdapter(
        context.Context,
        Config,
        RunnerRuntime,
        ModelRuntime,
    ) (agentx.ExecutionAdapter, error)
    ClassifyError(error) agentx.ErrorCode
}
```

Host拥有具体 catalog、provider、Runner/config和 adapter wiring。返回非 nil
adapter即表示 adapter已经接管 runner/model；此后失败时 construction只关闭
adapter。`ClassifyError` 只分类稳定错误码，不得把 backend错误文本当作
display-safe message。

## `New`

```go
func New(
    ctx context.Context,
    config Config,
    host Host,
) (*agentx.Client, error)
```

构造顺序固定为：

```text
validate
  → ResolveModel
  → NewRunner
  → NewAdapter
  → agentx.New
```

每个阶段后都会重新检查 caller context。失败 cleanup使用独立五秒 deadline，
顺序为 adapter，或 runner → model。成功后 Client独占 adapter，package不再
保存或关闭资源。

返回错误同时满足：

- `errors.Is/As` 可检查 `*agentx.Error`、context cause和 host cause；
- `Error()` 只返回稳定 display-safe文本；
- cancellation/deadline优先于 Host分类；
- 未识别 Host错误归为 `agentx.CodeExecutionFailed`。

## 并发与非目标

`New` 不持有进程级可变状态，可并发调用；资源隔离由 Host实现保证。

本 package不提供：

- concrete model resolver、provider、credential或网络 client；
- concrete Runner、工具/技能注册、Scene或产品 defaults；
- Workflow/Objective/Resume/长任务入口；
- 最终无 Host 参数的根 `agentxruntime.New`；
- Public、Beta、Stable或发行承诺。
