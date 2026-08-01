# 生命周期与错误处理

## Context

`Run` 和 `Shutdown` 都要求非 nil context：

- Run 在等待并发 gate 和执行 adapter 时尊重 cancellation/deadline；
- Shutdown 使用调用方 context 约束本次等待；
- context 到期不会被转换为纯文本错误，`errors.Is` 仍能识别原始 cause。

## 并发

同一 Client 支持多个 goroutine 调用，但有两个明确边界：

1. 重叠 Run 会按进入 gate 的顺序串行到达 adapter；
2. Shutdown 不等待 Run gate，可以并发通知 adapter 取消活动工作。

等待 Run gate 的请求若 context 到期，不会进入 adapter。

## Shutdown

Shutdown 一旦开始，Client 进入 closing 状态，新的 Run 返回
`CodeClientClosed`。adapter 的 Shutdown 必须是幂等的：

- 首次调用可以启动取消和 drain；
- 如果首次 context 太短，返回 `CodeShutdownFailed` 并保留 context cause；
- 调用方可以换用更长 context 再次调用；
- 已完成关闭后的调用应快速返回 nil。

关闭后的 Client 不支持重新打开；应创建新的 Client 和 adapter。

## 推荐错误判断

```go
result, err := client.Run(ctx, request)
if err != nil {
	var typed *agentx.Error
	if errors.As(err, &typed) {
		switch typed.Code {
		case agentx.CodeCanceled:
			// 由调用方决定是否重试。
		case agentx.CodeDeadlineExceeded:
			// 调整 deadline 或按业务策略重试。
		case agentx.CodeClientClosed:
			// 创建新 Client。
		}
	}
}
```

不要匹配英文 Message，也不要把 `Unwrap()` 得到的 backend 原始错误直接展示给
终端用户。`Retryable` 当前固定为 false；M3E 不在缺少真实 backend证据时承诺通用
自动重试策略。
