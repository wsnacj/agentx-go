# `runtime/executionpolicy` 中文 API Reference

成熟度：**Experimental extension / private validation**。

本包定义执行身份和 policy 的 portable 数据合同，并提供确定性的 Snapshot metadata、
bounded retry command、runtime decision packet、observation-only execution-loop report和
soft-rejection reducer。它只生成合同、命令或报告，不执行授权、审批、sandbox、provider
调用、工具调用、Host adapter或真实副作用。

## 主要类型

- `Contract`：一次执行使用的完整 policy快照输入。
- `Identity`：Profile、Pack、Workflow、Case、Run、Node和租户身份。
- `VisibilityPolicy`：工具 allow/deny、声明要求与风险上限。
- `BudgetPolicy`、`LoopPolicy`：工具调用、时限、token、成本、并发、重试和循环限制。
- `ApprovalPolicy`、`ReplayPolicy`、`RuntimeControlPolicy`：审批、replay和控制面声明。
- `SideEffectPolicy`、`SandboxPolicy`、`EvidencePolicy`：副作用、sandbox和证据要求。
- `Snapshot`、`Diff`、`CompileInput`：Host编译和持久化边界。
- `Compiler`：由 Host实现的 policy编译 port。
- `BudgetUsage`、`RetryRuntimeCommandInput`、`RetryRuntimeCommand`：同一逻辑 Run 的聚合
  usage和一次有界重试命令。
- `RuntimeDecision`、`DecisionStep`、`DecisionPacket`、`DecisionAuditRecord`：把调用方
  已完成的 policy/authorization事实规范化为确定性、可审计 packet。
- `ExecutionLoopReport`：把 decision packet投影为 Host下一步动作；始终保持
  `NoCoreExecution`、`NoToolInvocation`和`NoHostAdapterInvocation`。
- `SoftRejectionDecision`：规范化 allow/reject-content/halt，并选出最严格结果。

所有字段保持现有 JSON tag。空值是否生效、policy合并、授权决定和 backend行为由
Host拥有；本包不提供默认策略，也不暗示生产安全授权。

## Snapshot metadata

```go
func LoadSnapshotMetaJSON(string) (Snapshot, bool)
func MergeSnapshotMetaJSON(string, Snapshot) (string, error)
```

`MergeSnapshotMetaJSON`只更新Host metadata中的`agentx_execution_contract`字段，并保留
其它顶层字段。空值、非法JSON、缺失或类型错误的contract由`LoadSnapshotMetaJSON`
返回`false`；它不读写Store，也不迁移schema。

## Bounded retry command

```go
func CompileRetryRuntimeCommand(RetryRuntimeCommandInput) RetryRuntimeCommand
func RemainingBudgetAfterUsage(BudgetPolicy, BudgetUsage) BudgetPolicy
```

只有显式启用runtime retry、executor为tool loop、verification可重试、run retry预算未耗尽、
dedupe开启、聚合预算仍可用且所有Host确认齐全时，命令才进入`preflight_ready`。结果不调度
重试，也不复用tool-level retry budget；零预算沿用现有“unlimited”合同。

## Decision packet与execution-loop report

```go
func RuntimeDecisionSummaryFromDecision(RuntimeDecision) RuntimeDecisionSummary
func MergeRuntimeDecisionSummary(RuntimeDecisionSummary, RuntimeDecisionSummary) RuntimeDecisionSummary
func NewDecisionPacket(DecisionPacketInput) DecisionPacket
func BuildExecutionLoopReport(ExecutionLoopReportInput) ExecutionLoopReport
```

`NewDecisionPacket`按forbidden > prompt > allow的顺序收口最终动作，并生成稳定schema、summary
和display-safe audit。`BuildExecutionLoopReport`只根据packet和Host提供的opaque refs投影
blocked、ready-for-host-execution或ready-for-retry；检测到疑似raw output时fail closed。
即使结果ready，也只代表Host可以继续，不代表Core已经执行。

## Soft rejection

```go
func NewSoftRejectionDecision(string, string, string, string, string) SoftRejectionDecision
func NormalizeSoftRejectionDecision(SoftRejectionDecision) SoftRejectionDecision
func PrimarySoftRejectionDecision([]SoftRejectionDecision) (SoftRejectionDecision, bool)
```

优先级固定为`halt > reject_content > allow`。未知action规范化为空值，不会自动升级为允许。

## 并发与错误

本包没有可变全局状态或后台生命周期；reducer可由多个goroutine并发调用。
`Compiler`实现必须遵守调用方`context.Context`。除metadata JSON encode/decode外，新增
kernel不返回Go error；blocker、missing input、status和next Host action保留为typed字段。
`RuntimeEnforcementResult.Err`不会进入JSON，调用方不得把raw error写入display-safe refs。

## Fixed-version external consumer

[`runtime/conformance/execution-consumer`](../conformance/execution-consumer)固定依赖
`v0.0.0-20260801195859-a83a74b77995`，不使用`replace`，也不import HS、Runner、Scene、
provider或backend。除验证根execution Client的Run/Shutdown外，它还覆盖Snapshot metadata、
ready bounded retry command、decision packet和no-side-effect execution-loop report：

```bash
cd runtime/conformance/execution-consumer
GOWORK=off go test ./...
GOWORK=off go run .
```

预期输出：

```text
agentx-execution-ok:completed:execution-conformance:execution_policy_kernel_ready
```

## 非目标

- 具体 authorization/approval执行；
- 默认`Compiler`、Snapshot derive/tighten或具体tool-selector语义；
- retry scheduler、timer、dispatch、durable attempt ledger或RunStore；
- sandbox或网络隔离 backend；
- browser/network/gateway route policy和具体Host runtime；
- provider、credential、Scene业务策略；
- Public/Beta/Stable或 semver承诺。
