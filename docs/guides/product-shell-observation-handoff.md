# ProductShell Observation与Host Handoff

`extensions/productshell`提供一条Experimental、无副作用的观测交接路径：Host先把自己
拥有的session/process事实投影为typed observation，再选择display-safe operator line，
canonical helper只负责构建可验证的handoff envelope，真实交付继续由Host adapter完成。

```go
import productshell "github.com/wsnacj/agentx-go/extensions/productshell"
```

当前private-preview固定版本：

```bash
go get github.com/wsnacj/agentx-go/extensions@v0.0.0-20260801133815-af05058a8a7f
```

该pseudo-version只用于本轮可重复验证，不是正式semver，也不构成Public、Beta或Stable
兼容承诺。

## 责任链

```text
Host-owned session/process/backend facts
        -> typed observation input
        -> canonical normalization
        -> Host-owned operator-line selection
        -> canonical display-safe envelope + conformance
        -> Host-owned log/UI/HTTP delivery and readback
```

这个边界让外部项目复用数据合同、规范化、redaction和接入验证，又不会把HS的raw parser、
产品观测聚合或交付策略搬入通用extension。

## 1. 构造typed session observation

```go
session := productshell.BuildSessionObservation(productshell.SessionObservationInput{
    SessionID: "session-001",
    Events: []productshell.SessionEventObservationInput{
        {Role: "user", Content: "run tests"},
        {Role: "assistant", Content: "running", ToolCallCount: 1},
        {Role: "tool", Content: "ok", ToolCallID: "call-001"},
    },
    Branches: []productshell.SessionBranchObservationInput{{
        BranchID: "main", NodeExecID: "node-exec-001", Status: "completed",
    }},
})
```

调用方必须从自身session/transcript backend构造typed input。canonical builder不会连接
backend或解码transcript；它只计算计数、latest preview、branch摘要和compaction标签。

## 2. 构造typed host-process observation

```go
progress := productshell.BuildHostProcessProgressObservation(
    productshell.HostProcessProgressObservationInput{
        Source:                         "host_process_view",
        Available:                      true,
        Status:                         "completed",
        DisplayLine:                    "kind=progress;status=completed;process_count=1",
        SessionKey:                     session.SessionID,
        ProcessRef:                     "process-001",
        Terminal:                       true,
        ProcessCount:                   1,
        TerminalCount:                  1,
        ConsumesHostProcessSessionView: true,
    },
)
```

status、terminal、ready和readback字段必须来自Host已经授权的typed process view。本包不
从tool output猜测process map，不决定runstore protocol，也不控制process lifecycle。

## 3. 由Host选择operator line

```go
line := productshell.BuildHostDiagnosticOperatorLineObservation(
    productshell.HostDiagnosticOperatorLineObservationInput{
        Source:              "host_adapter",
        Key:                 "progress.completed",
        Available:           progress.Available,
        Status:              progress.Status,
        OperatorDisplayLine: progress.DisplayLine,
        NextHostAction:      "render_log_fields",
    },
)
```

从session/process observation到operator line的选择属于Host产品逻辑。canonical API不
自动拼出用户可见结论，也不拥有跨来源 `ObservationSnapshot`。

## 4. 构建display-safe envelope

```go
envelope := productshell.BuildHostUIHandoffEnvelopeFromOperatorLines(
    []productshell.HostDiagnosticOperatorLineObservation{*line},
    productshell.HostUIHandoffInput{
        Target: productshell.HostUIHandoffTargetLog,
        Source: "host_agent",
    },
)

fields, ok := productshell.RenderHostUIHandoffLogFields(envelope.Entries[0])
if !ok {
    return errors.New("handoff has no display-safe fields")
}
hostLogger.Info(fields) // 真正delivery由Host拥有
```

envelope只接受受限token和单行display text。URL、换行、tab或未允许字符会被替换为
`redacted`；它不是通用HTML sanitizer，Host针对具体UI仍需执行自身escape/CSP策略。

可用target常量包括 `log`、`side_panel`、`admin_surface`和 `run_output_json`。这些值只
标记交接意图，不会启动对应transport。

## 5. 检查conformance与runtime-use

```go
conformance := productshell.BuildHostUIHandoffConsumerConformanceReport(
    productshell.HostUIHandoffConsumerConformanceInput{
        Consumer:       "host_log_renderer",
        ExpectedTarget: productshell.HostUIHandoffTargetLog,
        ExpectedSource: "host_agent",
        Envelope:       envelope,
    },
)

runtimeUse := productshell.BuildHostUIHandoffRuntimeUseReport(
    productshell.HostUIHandoffRuntimeUseInput{
        Consumer:           "host_log_renderer",
        HostAdapter:        "host_agent",
        Target:             productshell.HostUIHandoffTargetLog,
        Source:             "host_agent",
        Envelope:           envelope,
        ConsumedEntryCount: len(envelope.Entries),
    },
)
```

conformance验证envelope形状和display-safe边界；runtime-use还要求消费全部entry，并确认
consumer没有解码raw diagnostics，也没有把Runtime伪装成delivery source。report只用于
focused接入测试，不替代产品审批、security审阅或发布证据。

## External consumer

完整可运行示例位于
[`extensions/conformance/productshell-observation-consumer`](../../extensions/conformance/productshell-observation-consumer)。
它固定实际pseudo-version，在无HS、Runner、Scene、长期`replace`、网络和凭据的条件下
验证typed Session/HostProcess/OperatorLine到handoff/conformance/runtime-use的纵向路径。

## 明确留在Host的能力

- raw diagnostics、tool output、session transcript与RunStore parser；
- `ObservationSnapshot`聚合、历史存储、inventory、订阅和readback；
- process生命周期、authorization、approval和产品状态解释；
- operator line选择、国际化、日志级别、UI布局和真实delivery；
- provider、credential、Scene、HTTP/CLI和其它生产副作用。

该路径仍为Experimental，不构成Public、Beta、Stable、semver或正式发行承诺。
