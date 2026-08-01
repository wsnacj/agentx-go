# `agentx` Go API Reference

本页是 M3D 根 contract Developer Preview candidate 的中文 Reference。正文覆盖和
签名漂移已经进入 focused gate，但 API仍处于 private validation，不构成 Public、
Beta、Stable或 semver兼容性承诺。

## 创建与 Client

<!-- api:Config -->
### `Config`

```go
type Config struct {
    Adapter ExecutionAdapter
    Profile ExecutionProfile
}
```

`Adapter` 必填。`Profile` 为零值时解析为当前根合同唯一支持的默认画像。`New` 成功后，
调用方不得继续直接调用同一个 adapter。

<!-- api:New -->
### `New`

```go
func New(Config) (*Client, error)
```

配置无效时返回 `CodeInvalidArgument`；画像不受支持时返回
`CodeUnsupportedProfile`。

<!-- api:Client -->
### `Client`

`Client` 是底层 adapter 之上的薄合同。方法可以并发调用；同一 Client 的 Run
会串行进入 adapter，Shutdown 可以与活动 Run 并发。

<!-- api:Client.Run -->
### `Client.Run`

```go
func (*Client) Run(context.Context, RunRequest) (RunResult, error)
```

context 不得为 nil。等待并发 Run gate 时同样尊重 cancellation/deadline。

<!-- api:Client.Shutdown -->
### `Client.Shutdown`

```go
func (*Client) Shutdown(context.Context) error
```

开始或继续有界关闭。adapter 的 Shutdown 必须幂等，并允许调用方在前一次 context
到期后用新 context 继续等待。Shutdown 开始后，新 Run 返回
`CodeClientClosed`。

## Run 输入与结果

<!-- api:RunRequest -->
### `RunRequest`

```go
type RunRequest struct {
    Input     string
    SessionID string
}
```

`Input` 去除首尾空白后不得为空。`SessionID` 可为空；identity 生成策略由 adapter
或后续 Runtime owner 持有。

<!-- api:RunResult -->
### `RunResult`

```go
type RunResult struct {
    RunID      string
    SessionID  string
    Status     string
    Reply      string
    Evidence   []string
    Blockers   []string
    NextAction string
    Profile    ExecutionProfile
}
```

`Status` 只使用 `completed`、`blocked`、`canceled`、`failed`。`Evidence` 当前
只投影 run/session identity，不保存 provider 原始响应。

## Adapter 扩展接缝

<!-- api:ExecutionAdapter -->
### `ExecutionAdapter`

```go
type ExecutionAdapter interface {
    Run(context.Context, AdapterRunRequest) (*AdapterRunResult, error)
    Shutdown(context.Context) error
    ClassifyError(error) ErrorCode
}
```

adapter 负责底层 Runtime 构造、执行、状态提取、关闭和错误分类。Client 保证同一
adapter 不会收到重叠 Run，但 Shutdown 可以与 Run 并发。

<!-- api:AdapterRunRequest -->
### `AdapterRunRequest`

```go
type AdapterRunRequest struct {
    Input     string
    SessionID string
}
```

这是传给 adapter 的最小输入，不包含 provider、credential 或 Scene 配置。

<!-- api:AdapterRunResult -->
### `AdapterRunResult`

```go
type AdapterRunResult struct {
    RunID     string
    SessionID string
    Status    string
    Reply     string
}
```

adapter 可以返回 owner 状态别名；Client 会将其收敛为四种公共状态。未知状态按
`failed` 处理。

## 执行画像

<!-- api:ExecutionProfile -->
### `ExecutionProfile`

```go
type ExecutionProfile struct {
    Activation         string
    ControlMode        string
    ExecutionIntensity string
    Driver             string
    ResultPolicy       string
    Lifecycle          string
}
```

当前根合同唯一支持：

```text
off / tool / l2_bounded_tool_loop /
open_tool_loop / runner_final_reply / synchronous_run
```

其他组合返回 `CodeUnsupportedProfile`，不代表相应能力已实现。

## Typed error

<!-- api:ErrorCode -->
### `ErrorCode`

`ErrorCode` 是稳定机器分类。调用方应使用 `errors.Is/As` 或检查 `Error.Code`，
不应匹配 `Error()` 文本。

<!-- api:CodeInvalidArgument -->
### `CodeInvalidArgument`

请求、context 或构造参数无效。

<!-- api:CodeCanceled -->
### `CodeCanceled`

执行被调用方取消；原始 cause 保留 `context.Canceled`。

<!-- api:CodeDeadlineExceeded -->
### `CodeDeadlineExceeded`

Run 超过调用方 deadline；原始 cause 保留 `context.DeadlineExceeded`。

<!-- api:CodeClientClosed -->
### `CodeClientClosed`

Client 已开始关闭、已经关闭，或底层 owner 报告对应状态。

<!-- api:CodeUnsupportedProfile -->
### `CodeUnsupportedProfile`

请求的六维画像不属于当前根合同支持范围。

<!-- api:CodeExecutionFailed -->
### `CodeExecutionFailed`

adapter 未分类错误或普通执行失败。

<!-- api:CodeShutdownFailed -->
### `CodeShutdownFailed`

本次 Shutdown 未能在调用 context 内完成；可以使用新 context 继续调用。

<!-- api:Error -->
### `Error`

```go
type Error struct {
    Code      ErrorCode
    Retryable bool
    Message   string
}
```

Client 返回的 `Message` 是 display-safe 文本，不包含底层原始错误。

<!-- api:Error.Error -->
### `Error.Error`

返回 display-safe 文本。

<!-- api:Error.Unwrap -->
### `Error.Unwrap`

返回底层 cause，使 `errors.Is(err, context.Canceled)` 等检查继续有效。

<!-- api:Error.Is -->
### `Error.Is`

按非空 `ErrorCode` 比较 AgentX typed error。
