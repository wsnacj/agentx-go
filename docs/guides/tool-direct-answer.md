# Tool Direct Answer 接入

Tool Direct Answer 是一种显式结果策略：工具执行后，Host已经确认某个结果可安全、完整地
直接展示，因此不再发起下一轮模型合成。它不是独立Runner mode，也不会自动解析任意工具
JSON。

## 合同

```go
type ToolDirectAnswer struct {
    Reply  string
    Source string
    Reason string
}

type ToolResult struct {
    // Runs、Failures、NextChunks、ForceNoToolCalls省略
    DirectAnswer *ToolDirectAnswer
}
```

`ExecuteTools`返回非空`DirectAnswer`后，canonical Host Kit验证`Reply`非空，并把本轮投影
为`OutcomeCompleted`。`Source`和`Reason`是不透明诊断事实，不影响portable状态机。

## Host责任

设置`DirectAnswer`前，Host必须完成：

1. 工具调用授权、sandbox与副作用控制；
2. 结果来源和完整度判断；
3. credential、内部错误和敏感字段脱敏；
4. display-safe reply生成；
5. 产品需要的persistence、telemetry和审计投影。

Core只拥有“显式直接答案 -> 完成Run”的通用机制，不拥有具体工具名、JSON schema、
provider、审批策略或业务可信度规则。

## 与 HS 的映射

HS仍解析自己的`answer_contract`并处理recovery、budget、session persistence和event；当
业务规则确认结果可直接回答时，HS构造`ToolDirectAnswer`，再调用 canonical
`ModelToolRoundResult.ExecutionResult()`完成portable outcome。这样既保持原有错误、事件和
持久化顺序，也避免HS重新实现通用结果策略。

## 错误与边界

- `Reply`为空返回执行错误，不会产生空的completed结果；
- 未设置`DirectAnswer`时，工具批次仍按原合同进入下一轮；
- Host gate停止仍是`OutcomeTerminated`；
- Direct Answer不等于模型合成，不应把原始内部JSON直接作为`Reply`。

可运行的无HS固定版本consumer位于
[`runtime/conformance/hostkit-consumer`](../../runtime/conformance/hostkit-consumer)。
