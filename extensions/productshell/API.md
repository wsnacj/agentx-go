# `extensions/productshell` 中文 API Reference

导入路径：

```go
import productshell "github.com/wsnacj/agentx-go/extensions/productshell"
```

成熟度：**Experimental extension / private validation**。

本包拥有 ProductShell 在进入执行前的 portable 输入投影、Case/Shell Binding helper、
Workflow 绑定检查和确定性准备顺序；同时提供 typed session/host-process observation 的
归一化，以及从 display-safe operator line 到 Host UI handoff envelope 的确定性投影。
产品路由、Pack registry、Case 草拟、Workflow物化、持久化、raw诊断解析、真实观测聚合
和交付副作用必须由Host拥有。

完整接入说明见 [ProductShell 两阶段准备指南](../../docs/guides/product-shell-preparation.md)、
[Observation与Host Handoff指南](../../docs/guides/product-shell-observation-handoff.md)、
[preparation fixed-version consumer](../conformance/productshell-consumer)和
[observation fixed-version consumer](../conformance/productshell-observation-consumer)。

## 安装

当前 private-preview 固定版本：

```bash
go get github.com/wsnacj/agentx-go/extensions@v0.0.0-20260801133815-af05058a8a7f
```

该 pseudo-version 只用于当前验证，可以在后续 checkpoint 更新；它不是正式 semver，
也不构成 Public、Beta 或 Stable 兼容承诺。

## 两阶段 preparation

### 第一阶段：确定性输入投影

调用方先把外部 option 投影为 typed `Input`：

```go
shell, requestedBinding, sessionInput, workflowState, residual :=
    productshell.ProjectInputOptions(rawOptions)

input := productshell.Input{
    UserMessage:           userMessage,
    ProductShell:          productShell,
    RequestedShellBinding: requestedBinding,
    SessionInput:          sessionInput,
    WorkflowState:         workflowState,
    ShellOptions:          shell,
    Options:               residual,
}
```

这一阶段只做解码、规范化、显式 skill/path 解析、map/slice 复制和 typed field 优先级
处理，不查询 provider、registry 或 backend。主要 helper 包括：

- `ProjectInputOptions`、`ComposeInputOptions`、`MergeInputOptions`；
- `ParseInputMapOption`、`ParseRequestedShellBindingOption`；
- `DecodeShellBindingOption`、`EncodeRequestedShellBindingOption`；
- `DecodeWorkflowSpec`、`ResolveWorkflowBinding`、`ExplicitRawWorkflowOptIn`；
- `ParseRequestedSkills`、`ParseSkillActivationPaths`；
- `CloneShellBindingMap`、`MergeShellBindingMaps`、`NormalizeShellBinding`。

`ParseRequestedSkills`只接受 typed option、受支持的 option key和显式
`[skill:name]` directive，不会根据 `ProductShell`或自然语言自行推断技能。coding或其它
产品特定推断继续由 Host拥有。

### 第二阶段：Host 端口驱动的固定顺序准备

```go
runtime := productshell.PreparationRuntimeFuncs{
    ResolveShellBindingFn: resolveShellBinding,
    ResolveWorkflowFn:     resolveWorkflow,
    ResolveEffectiveCaseFn: resolveEffectiveCase,
    ValidateEffectiveCaseFn: validateEffectiveCase,
}

pipeline := productshell.NewPreparationPipeline(runtime)
result, err := pipeline.PrepareWithInput(ctx, productshell.PrepareInput{
    SessionID: sessionID,
    Input:     input,
})
```

`PreparationPipeline`固定以下高层顺序：

1. 合并显式 Case 输入；
2. 解析并应用 session Shell Binding；
3. 解析并应用 command dispatch；
4. 解析 requested skills和 activation paths；
5. 在 Host明确允许时解析并应用 Pack选择；
6. 在 Host明确允许时解析并应用 Case Binding草稿；
7. 解析 Workflow；
8. 解析、校验并应用 effective Case；
9. 汇总 command、shell、pack和case metrics。

任一阶段返回错误后，pipeline立即停止，不运行后续阶段；错误 identity原样返回。
传入的 `context.Context`会转交给需要 I/O或策略决策的 Host hook，取消和 deadline由
Host hook遵守。pipeline自身不启动 goroutine，也不持有需要关闭的资源。

## 主要数据合同

- `Input`：准备阶段的 typed输入，包含消息、Case、Workflow Spec、Pack/Case/
  Workflow identity、session input、workflow state和 options；
- `InputShellOptions`：自动绑定/规划开关、显式 requested skills及 activation paths；
- `PrepareInput`：为一次准备补充 Session ID和可选 Host LLM任务超时时间；超时值只
  传给 Host的 Case草拟 hook，本包不会发起 LLM请求；
- `PrepareResult`：返回最终 `Input`、解析后的消息/技能、Workflow/Case/Binding结果
  以及各阶段 metrics；它不是执行结果，也不会启动 Workflow；
- `ShellBinding`、`RequestedShellBinding`、`PreparedShellBinding`：Pack、Case、
  Workflow、session input和state的显式绑定合同；
- `PreparedPackSelection`、`PreparedCaseBinding`、`ResolvedWorkflow`：Host解析结果；
- `CommandDispatchMetrics`、`ShellBindingMetrics`、`PackSelectionMetrics`、
  `CaseBindingMetrics`：准备阶段的可观察证据，不是授权或业务审计结论。

`runtime/cases.Case`、`runtime/workflow.Spec`、`extensions/pack.Binding`和
`extensions/skills.RequestedSkillSemantic`保持各自 package 的 source authority；
本包只组合它们，不复制定义。

## Host port

```go
type PreparationRuntime interface {
    // 输入、Shell Binding与command dispatch
    // requested skills、Pack/Case Binding
    // Workflow、effective Case与metrics
}
```

完整签名以 `go doc`为准。端口按责任可分为：

- **纯投影 hook**：`ApplyInputCase`、`ApplyShellBinding`、`ApplyCommandDispatch`、
  `ApplyPackSelection`、`ApplyCaseBinding`、`ApplyEffectiveCase`；
- **Host查询/策略 hook**：`ResolveShellBinding`、`ResolveCommandDispatch`、
  `ShouldAttemptPackSelection`、`ResolvePackSelection`、`ShouldAttemptCaseBinding`、
  `ResolveCandidateCaseBinding`、`ResolveCaseBindingDraft`；
- **Workflow/Case hook**：`ResolveWorkflow`、`ResolveEffectiveCase`、
  `ValidateEffectiveCase`；
- **观察投影 hook**：`MergeCaseBindingMetrics`、`FinalizeShellBindingMetrics`、
  `PackSelectionMetricsFromPrepared`。

`PreparationRuntimeFuncs`把函数适配为该接口。未提供策略/查询 hook时默认不尝试相关
功能；可由 canonical helper完成的 apply/parse/metrics hook使用无副作用默认实现。
`NewPreparationPipeline(nil)`和 nil pipeline均返回 pass-through结果，不能被解释为
Pack、Workflow或Case能力已经配置。

## Workflow 绑定

`WorkflowResolutionRuntime`提供独立的 portable绑定检查：

- typed `WorkflowSpec`优先于 option中的 `workflow_spec`/`workflowSpec`/`workflow`；
- 不带 Pack的显式 raw Workflow必须通过 `RawWorkflowOptIn`或对应 option显式启用；
- 带 Pack的 Workflow必须能由 Host注册表解析为同一 execution semantics；
- Pack注册检查、Binding解析和 Workflow物化仍由 Host函数提供；
- 本包不校验完整 Workflow结构，也不执行 Workflow。结构校验和执行由
  `runtime/workflow`及其 Host Kit负责。

## Shell Binding、JSON与复制语义

`DecodeShellBindingOption`接受 typed值、JSON string/bytes及可JSON编码对象，返回
`(binding, hasValues, persist, error)`。`NormalizeShellBinding` trim identity并递归
复制 `map[string]any`/`[]any`；merge采用 overlay优先，嵌套map递归合并。

`LoadSessionShellBindingMetaJSON`和 `MergeSessionShellBindingMetaJSON`只读写
`agentx_shell_binding`字段，并保留meta JSON中的其它字段；它们不负责 session存储。
是否读取、持久化、加密或过期仍由 Host决定。

`ProjectInputOptions`仅保留已知 ProductShell输入 key。无法解析的 typed option不会
触发外部副作用；需要向用户返回严格输入错误时，Host应在进入 pipeline前调用对应的
`Parse*`/`Decode*`函数并处理错误。

## Typed Observation

M5H新增三组纯构造函数，用于把Host已经读取并解释的typed事实归一化为可观察DTO：

- `BuildSessionObservation(SessionObservationInput)`：归一化session事件计数、最新内容
  preview、branch摘要和compaction计数；
- `BuildHostProcessProgressObservation(HostProcessProgressObservationInput)`：归一化Host
  process view、状态/identity、ready标志和明确边界；
- `BuildHostDiagnosticOperatorLineObservation(HostDiagnosticOperatorLineObservationInput)`：
  归一化一条准备展示的operator line。

空输入返回 `nil`；输入字符串会trim，slice会复制；HostProcess的 `ProcessCount`、
`ActiveCount`、`TerminalCount`和 `HostProcessEventCount`负值会归零。session preview
有界且压缩空白；session compaction计数保持调用方提供的原值。这些函数不读取engine
session、transcript、tool output、RunStore、process registry或Host diagnostics JSON，
也不推断执行权限、生命周期或readback结论。

`SessionObservation`、`HostProcessProgressObservation`和
`HostDiagnosticOperatorLineObservation`是不同粒度的typed数据合同，不组成由本包长期
持有的 `ObservationSnapshot`。聚合时刻、历史保留、增量订阅和一致性仍由Host决定。

## Display-safe Host UI Handoff

Host完成产品判断后，可把明确选定的operator line投影为display-safe envelope：

```go
line := productshell.BuildHostDiagnosticOperatorLineObservation(
    productshell.HostDiagnosticOperatorLineObservationInput{
        Source:              "host_adapter",
        Key:                 "progress.completed",
        Available:           true,
        Status:              "completed",
        OperatorDisplayLine: "kind=progress;status=completed",
        NextHostAction:      "render_log_fields",
    },
)

envelope := productshell.BuildHostUIHandoffEnvelopeFromOperatorLines(
    []productshell.HostDiagnosticOperatorLineObservation{*line},
    productshell.HostUIHandoffInput{
        Target: productshell.HostUIHandoffTargetLog,
        Source: "host_agent",
    },
)
```

`HostUIHandoffEnvelope`固定schema、surface、target/source、entries、latest entry和边界。
token/display line只允许有限字符；URL、换行或其它不安全值会被替换为 `redacted`，不会
原样流向展示层。`RenderHostUIHandoffLogFields`可产生单行display-safe字段，但真正写
日志、side panel、admin surface或run output仍由Host adapter负责。

可选验证函数：

- `BuildHostUIHandoffConsumerConformanceReport`：验证schema/surface/target/source、
  entry数量、latest entry、display-safe字段和delivery boundary；
- `BuildHostUIHandoffRuntimeUseReport`：记录具体Host adapter消费了多少entry，并拒绝
  raw diagnostics decode或把Runtime误声明为delivery source。

这两类report是接入验证证据，不是全仓inventory/evidence registry，也不授予生产准入。
它们不发送消息、不写日志、不调用HTTP/UI、不执行readback。

## 并发、生命周期和错误

- `PreparationPipeline`构造后只保存一个 `PreparationRuntime`引用；多个 goroutine并发
  使用时，Host实现及其依赖必须自行保证并发安全；
- observation与handoff builder不持有共享状态；返回值包含调用方slice的规范化副本，
  调用方仍应把每次构建结果视为独立快照值；
- 本包没有共享 registry、后台任务或 `Shutdown`；backend生命周期由 Host拥有；
- pipeline按固定顺序返回首个错误，不包装 Host错误，因此 `errors.Is/As` identity得以
  保留；
- codec/Workflow绑定错误包含当前输入上下文，但其文本仍处于 Experimental，不能作为
  正式稳定协议依赖。

## 明确 non-goal

- 自然语言规划、LLM调用、model/tool round或 provider选择；
- ProductShell路由策略、command/skill推断策略、authorization、approval或sandbox；
- Pack registry、Case Store、Workflow registry、RunStore或其它具体 backend；
- Workflow执行、Objective Runtime、Resume、scheduler或durable lifecycle；
- raw Host diagnostics/tool output parser、完整 `ObservationSnapshot`、process inventory、
  observation storage/stream、readback或真实delivery；
- credential、Scene、HTTP/CLI、真实网络或生产副作用；
- Public、Beta、Stable、正式tag、semver或兼容性承诺。
