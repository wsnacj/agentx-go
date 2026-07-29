# 实现自定义 ExecutionAdapter

`ExecutionAdapter` 是 W1 唯一的 extension-authoring seam。它用于把已有 Runtime
或确定性执行器接到公共合同，不用于把 provider、credential 或业务策略塞进根包。

## 责任边界

adapter 必须负责：

- 将 `AdapterRunRequest` 转换为 owner 的原生输入；
- 返回 run/session identity、owner status 和 reply；
- 将底层错误归类为允许的 `ErrorCode`；
- 在 Shutdown 时取消、收敛活动执行；
- 让 Shutdown 可重复调用，并尊重每次调用的 context。

Client 负责：

- 校验公共输入；
- 串行化同一 Client 的 Run；
- 映射 cancellation/deadline 和 adapter 错误；
- 收敛公共 status、evidence、blocker 和 next action；
- Shutdown 开始后拒绝新 Run。

## 错误分类

`ClassifyError` 只应返回：

- `CodeCanceled`
- `CodeDeadlineExceeded`
- `CodeClientClosed`
- `CodeExecutionFailed`

其他值会被规范化为 `CodeExecutionFailed`。构造参数、画像和 Shutdown 的公共错误
由 Client 自己生成，adapter 不负责伪造。

## 所有权

一个 adapter 在传给 `New` 后由该 Client 独占。不要把同一 adapter 交给多个
Client，也不要在外部并发调用其 Run/Shutdown。需要共享底层资源时，应由 Runtime
owner 提供明确的线程安全资源层，而不是绕过合同。

W1-B 将提供可运行的 `custom-adapter` example；当前页面只固定实现责任，不把示例
当作完整 Runtime 能力证明。
