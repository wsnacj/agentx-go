# runtime/controlcontract API

导入路径：

```go
import controlcontract "github.com/wsnacj/agentx-go/runtime/controlcontract"
```

成熟度：**Experimental extension / private validation**。

`runtime/controlcontract` 是 AgentX 公共执行控制语义与确定性 reducer 的 portable
source authority。它只处理数据合同、复制/规范化、display-safe 校验和纯判定，不执行
Runner、模型、工具、调度、存储或 Host 副作用。

## 核心状态合同

本包提供下列 string-backed 枚举及稳定的 normalize/known-set helper：

```go
type Activation string
type ControlMode string
type ExecutionIntensity string
type EvidenceStrength string
type VerificationStatus string
type HostActionStatus string
type FailureClass string
type ReplannerSourceKind string

func KnownActivations() []Activation
func KnownControlModes() []ControlMode
func KnownExecutionIntensities() []ExecutionIntensity
func KnownFailureClasses() []FailureClass
func KnownReplannerSourceKinds() []ReplannerSourceKind

func NormalizeActivation(string) Activation
func NormalizeControlMode(string) ControlMode
func NormalizeExecutionIntensity(string) ExecutionIntensity
func NormalizeEvidenceStrength(string) EvidenceStrength
func NormalizeVerificationStatus(string) VerificationStatus
func NormalizeHostActionStatus(string) HostActionStatus
func NormalizeFailureClass(string) FailureClass
func NormalizeReplannerSourceKind(string) ReplannerSourceKind
```

未知值采用各类型既有的 fail-closed 结果；例如未知 activation 归一为 `off`，未知
verification status 归一为 `review_required`，未知 failure class 归一为
`internal_error`。调用方不应绕过 normalize helper 自行推断兼容别名。

## Identity、证据与执行投影

基础 identity 与边界类型：

```go
type Boundary string
type MissingInput string
type NextHostAction string
type DisplaySafeRef string
type AttemptRef string
```

主要数据合同：

```go
type RedactionPolicy struct { ... }
type EvidenceRef struct { ... }
type ObjectiveFrame struct { ... }
type StrategyCandidate struct { ... }
type Observation struct { ... }
type AttemptSummary struct { ... }
type Attempt struct { ... }
type AttemptLedgerPatch struct { ... }
type ManagedObjectiveProjection struct { ... }
type ManagedObjectiveReplannerProjection struct { ... }
type ReplannerSourceProjection struct { ... }
type VerificationResult struct { ... }
type HostActionProposal struct { ... }
```

这些 struct 保留固定 JSON tag、`Clone` 和 `Normalize` 语义。Clone 会复制 slice，
Normalize 会清理枚举、display-safe ref、boundary、missing input 与 evidence；不会读取
文件、URL、日志、prompt 或 secret，也不会触发执行。

## Display-safe 与合并 helper

```go
func NormalizeDisplaySafeRef(string) (DisplaySafeRef, bool)
func NormalizeAttemptRef(string) (AttemptRef, bool)
func DisplaySafeRefs([]string) []DisplaySafeRef
func ContainsUnsafeRawOutput(...string) bool
func VerifyDisplaySafeOnly(bool, []string) VerificationResult
func NormalizeNextHostAction(string) NextHostAction

func MergeBoundaries(...[]Boundary) []Boundary
func AppendBoundaries([]Boundary, ...Boundary) []Boundary
func MergeMissingInputs(...[]MissingInput) []MissingInput
func AppendMissingInputs([]MissingInput, ...MissingInput) []MissingInput
func MergeEvidenceRefs(...[]EvidenceRef) []EvidenceRef
```

`DisplaySafeRef` 只接受字母开头、最多 128 个字符的受限 token。URL、本地绝对路径、
疑似 credential/secret 和 PEM 私钥文本会被拒绝。该机制是 display-safe 投影保护，
不是完整的内容安全、授权或脱敏系统。

## Managed Objective 纯 reducer

```go
func BuildManagedObjectiveProjection(ManagedObjectiveProjectionInput) ManagedObjectiveProjection
func BuildManagedObjectiveReplannerProjection(ManagedObjectiveReplannerInput) ManagedObjectiveReplannerProjection
func BuildReplannerSourceProjection(ReplannerSourceInput) ReplannerSourceProjection
```

三者只根据显式输入构造稳定投影，始终保持 `RunnerEffect == "none"` 与
`PromptEffect == "none"`。它们不会选择真实 provider、执行策略、提交 proposal、调度
Objective 或写入 store。Host 仍需单独完成授权、执行、持久化和验证。

## Approval、budget、幂等与生命周期

```go
func EvaluateHostApprovalGate(ApprovalGateInput) HostActionProposal
func EvaluateRetryBudgetGate(BudgetGateInput) BudgetGateResult
func CheckIdempotency(IdempotencyCheckInput) IdempotencyCheckResult
func SelectLatestEvent([]EventRef) (EventRef, bool)
func EventIsNewer(EventRef, EventRef) bool
func CheckLatestSourceEvent(EventRef, EventRef) LatestSourceCheckResult
func RequireVerificationReview(VerificationResult, ReviewRequiredInput) VerificationResult
func RequireHostActionReview(HostActionProposal, ReviewRequiredInput) HostActionProposal
func NormalizeLifecycleStage(string) LifecycleStage
func CheckLifecycleTransition(LifecycleStage, LifecycleStage) LifecycleTransitionResult
```

这些函数都是同步、无 I/O、无副作用的确定性判定。它们不自动执行获准操作，不拥有
真实 retry timer、事务、锁、scheduler 或 durable lifecycle。

## 并发、取消与错误

本包不保存可变全局状态，函数可由多个 goroutine 并发调用。API 不接收
`context.Context`，因为它们只进行内存中的有界同步计算；取消、deadline 与 Shutdown
属于上层 Runtime/Host 合同。

本包不返回 Go `error`，而用 typed status、`FailureClass`、`MissingInput`、
`Boundary` 与 `NextHostAction` 表示可审计结果。调用方不得把 display-safe projection
误当成已执行、已持久化或已验证的事实。

## 最小示例

```go
result := controlcontract.EvaluateRetryBudgetGate(controlcontract.BudgetGateInput{
    Limit:     3,
    Used:      1,
    Increment: 1,
    Scope:     "objective:demo",
})
if !result.Allowed {
    // 使用 result.FailureClass、MissingInputs 与 NextHostAction 交给 Host 处理。
}
```

## Fixed-version external consumer

[`runtime/conformance/controlcontract-consumer`](../conformance/controlcontract-consumer)
是独立 nested module，固定依赖
`v0.0.0-20260801153451-11cc3fc9419e`，不使用 `replace`，也不 import HS、Runner、
Scene、provider 或 backend。它组合 managed-objective projection、retry budget、
lifecycle transition 与 unsafe-ref fail-closed 路径：

```bash
cd runtime/conformance/controlcontract-consumer
GOWORK=off go test ./...
GOWORK=off go run .
```

预期输出：

```text
agentx-controlcontract-ok:ready_for_host_action:1:applied:evidence_weak
```

## 非目标

- 不提供 Objective graph、executor、runtime loop、scheduler 或 RunStore；
- 不执行 approval、authorization、sandbox、model/tool 或 backend；
- 不包含 ProductShellRuntime、Scene、provider、credential 或真实网络；
- 不构成 Public、Beta、Stable、production-ready 或正式发行声明。
