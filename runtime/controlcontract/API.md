# runtime/controlcontract API

导入路径：

```go
import controlcontract "github.com/wsnacj/agentx-go/runtime/controlcontract"
```

成熟度：**Experimental extension / private validation**。

`runtime/controlcontract` 是 AgentX 公共执行控制语义、Objective定义/策略/图、验证/恢复/
重规划、auto-delegation、runtime step、Host executor request/result/readback与确定性 reducer
的 portable source authority。它只处理数据合同、复制/规范化、display-safe校验、严格
解码和纯判定，不执行Runner、模型、工具、调度、存储或Host副作用。

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

## Objective定义、capability与策略准备

Objective定义合同：

```go
type ObjectiveSpec struct { ... }
type ObjectiveSuccessCriterion struct { ... }
type ObjectiveConstraint struct { ... }
type ObjectiveSpecBudget struct { ... }
type ObjectiveSpecJSONDecodeInput struct { ... }
type ObjectiveSpecJSONDecodeReport struct { ... }
type ObjectiveSpecBuilder interface { ... }

func BuildObjectiveSpecFrameProjection(ObjectiveSpecFrameProjectionInput) ObjectiveSpecFrameProjection
func BuildObjectiveSpecFromJSON(ObjectiveSpecJSONDecodeInput) ObjectiveSpecJSONDecodeReport
func BuildObjectiveSpecWithBuilder(context.Context, ObjectiveSpecBuildInput) ObjectiveSpecBuildReport
```

JSON入口使用严格解码：未知字段、尾随内容、空输入和unsafe raw output都会fail closed。
`ObjectiveSpecBuilder`只负责把Host提供的结构化builder接到portable decode/validation
流程；本包不拼prompt、不选择模型，也不拥有业务goal解析策略。

策略准备与强度判定：

```go
type AdapterMetadataRegistrySnapshot struct { ... }
type ObjectiveCapabilityDescriptor struct { ... }
type StrategyCatalogSnapshot struct { ... }
type ExecutionIntensityPolicy struct { ... }
type IntensityGateResult struct { ... }
type ObjectiveControllerDecision struct { ... }
type StrategyPlannerResult struct { ... }

func BuildAdapterMetadataRegistrySnapshot(AdapterMetadataRegistrySnapshotInput) AdapterMetadataRegistrySnapshot
func BuildObjectiveCapabilityDescriptorProjection(ObjectiveCapabilityDescriptorProjectionInput) ObjectiveCapabilityDescriptorProjection
func BuildExecutionIntensityPreGate(IntensityGateInput) IntensityGateResult
func BuildExecutionIntensityFinalGate(IntensityGateInput) IntensityGateResult
func BuildObjectiveRun(ObjectiveRunInput) ObjectiveRun
func BuildObjectiveControllerDecision(ObjectiveControllerInput) ObjectiveControllerDecision
func BuildStrategyPlanner(StrategyPlannerInput) StrategyPlannerResult
```

这些API只基于Host显式传入的metadata、policy、approval、budget和evidence做投影与排序。
它们不会探测真实capability、执行strategy、批准side effect或绕过Host授权。

## Objective Graph kernel

```go
type ObjectiveGraph struct { ... }
type ObjectiveNode struct { ... }
type ObjectiveGraphValidationReport struct { ... }
type ObjectiveGraphPlanner interface { ... }

func BuildObjectiveGraphValidation(ObjectiveGraphValidationInput) ObjectiveGraphValidationReport
func BuildObjectiveGraphFromJSON(ObjectiveGraphJSONDecodeInput) ObjectiveGraphJSONDecodeReport
func BuildObjectiveGraphWithPlanner(context.Context, ObjectiveGraphBuildInput) ObjectiveGraphBuildReport
```

Graph kernel拥有节点/依赖/attempt policy、严格JSON decode、cycle、预算、evidence、
capability/strategy匹配、side-effect与ready-node判定。它只返回可审计validation结果；
`ReadyForRuntimeLoop`不等于已执行。节点executor、runtime loop dispatch、scheduler、
RunStore和durable write仍由Host拥有。

## Objective evidence 与 verification kernel

Host先把具体adapter、tool或workflow输出翻译为display-safe的
`ObservationNormalizationResult`；本包不拥有具体输入翻译器：

```go
type ObservationNormalizationResult struct { ... }
type ObjectiveRequiredEvidenceContract struct { ... }
type ObjectiveVerificationGateResult struct { ... }

func CloneObservationNormalizationResult(ObservationNormalizationResult) ObservationNormalizationResult
func BuildObjectiveRequiredEvidenceContract(ObjectiveRequiredEvidenceContractInput) ObjectiveRequiredEvidenceContract
func BuildObjectiveVerificationGate(ObjectiveVerificationGateInput) ObjectiveVerificationGateResult
```

`ObservationNormalizationResult.Normalize`只规范化已经结构化的observation、evidence、
status和display-safe ref。它不会读取transcript、raw tool output或调用adapter。
verification gate逐条匹配显式required evidence；没有required-evidence合同、证据强度不足、
proposal-only observation或unsafe payload都会fail closed，不会仅凭success-criteria文本
推断Objective已完成。

可选semantic verifier通过窄Host port接入：

```go
type ObjectiveSemanticVerifier interface {
    VerifyObjectiveSemantics(context.Context, ObjectiveSemanticVerifierRequest) (ObjectiveSemanticVerifierResponse, error)
}

func BuildObjectiveSemanticVerificationFromJSON(ObjectiveSemanticVerificationJSONDecodeInput) ObjectiveSemanticVerificationJSONDecodeReport
func BuildObjectiveSemanticVerification(context.Context, ObjectiveSemanticVerificationInput) ObjectiveSemanticVerificationReport
```

semantic结果是advisory；它不能单独把Objective标记为satisfied，也不能绕过结构化evidence
gate。取消与deadline原样传递给Host verifier。

## Recovery 与 replanning kernel

```go
type ObjectiveRecoveryContract struct { ... }
type ObjectiveReplannerDecision struct { ... }
type ObjectiveReplanProposal struct { ... }
type ObjectiveReplanGraphPatch struct { ... }

func BuildObjectiveRecoveryContract(ObjectiveRecoveryContractInput) ObjectiveRecoveryContract
func BuildObjectiveRecoveryContractFromJSON(ObjectiveRecoveryContractJSONDecodeInput) ObjectiveRecoveryContractJSONDecodeReport
func BuildObjectiveSafeDefaultProposal(ObjectiveSafeDefaultProposalInput) ObjectiveSafeDefaultProposal
func BuildObjectiveSideEffectSplitProposal(ObjectiveSideEffectSplitProposalInput) ObjectiveSideEffectSplitProposal
func BuildObjectiveNoProgressSwitchGate(ObjectiveNoProgressSwitchGateInput) ObjectiveNoProgressSwitchGate
func BuildObjectiveReplannerDecision(ObjectiveReplannerDecisionInput) ObjectiveReplannerDecision
func BuildObjectiveReplanProposal(ObjectiveReplanProposalInput) ObjectiveReplanProposal
func BuildObjectiveReplanGraphPatch(ObjectiveReplanGraphPatchInput) ObjectiveReplanGraphPatch
```

这些API只形成display-safe、可审阅的恢复与重规划建议。它们不会修改现有graph、切换
strategy、安装capability、请求credential、执行side effect或写入RunStore；Host必须在
独立的authorization、budget、scheduler和durable-write边界内审阅并应用proposal。

## Host副作用合同、准入与调用门控

统一 independent-effect gate 用独立、显式的 policy、approval、budget、idempotency、
readback、eval 与 failure/compensation review ref 描述 Host-owned 副作用：

```go
type ProductionAdapterEffectGateKind string
type ProductionAdapterIndependentEffectGateSpec struct { ... }
type ProductionAdapterIndependentEffectGatePlan struct { ... }
type ProductionAdapterIndependentEffectGate struct { ... }

func KnownProductionAdapterEffectGateKinds() []ProductionAdapterEffectGateKind
func NormalizeProductionAdapterEffectGateKind(string) ProductionAdapterEffectGateKind
func BuildProductionAdapterIndependentEffectGatePlan(ProductionAdapterIndependentEffectGatePlanInput) ProductionAdapterIndependentEffectGatePlan
func BuildProductionAdapterIndependentEffectGate(ProductionAdapterIndependentEffectGateSpec) ProductionAdapterIndependentEffectGate
```

完整 plan 要求 scheduler、installer、workflow retry、runtime executor、delegation worker、
memory apply 与 compensation executor 各自拥有 gate；请求单一 aggregate auto-executor 会
fail closed。`ReadyForIndependentGatePlan`只表示合同输入完整，不授权或执行任何副作用。

本包同时拥有 workflow executor 的 Host 准入，以及 capability/scheduler 的完整
request-result-readback 和 adapter readiness/invocation 投影：

```go
func BuildHostOwnedWorkflowRuntimeExecutorReadiness(HostOwnedWorkflowRuntimeExecutorReadinessInput) HostOwnedWorkflowRuntimeExecutorReadiness
func BuildHostOwnedWorkflowRuntimeExecutorInvocation(HostOwnedWorkflowRuntimeExecutorInvocationInput) HostOwnedWorkflowRuntimeExecutorInvocation

type CapabilityApplyAction string
func BuildHostOwnedCapabilityApplyDescriptor(HostOwnedCapabilityApplyDescriptor) HostOwnedCapabilityApplyDescriptor
func BuildHostOwnedCapabilityApplyRequest(HostOwnedCapabilityApplyRequestInput) HostOwnedCapabilityApplyRequest
func BuildHostOwnedCapabilityApplyResult(HostOwnedCapabilityApplyResultInput) HostOwnedCapabilityApplyResult
func BuildHostOwnedCapabilityApplyReadback(HostOwnedCapabilityApplyReadbackInput) HostOwnedCapabilityApplyReadback
func BuildHostOwnedCapabilityApplyAdapterReadiness(HostOwnedCapabilityApplyAdapterReadinessInput) HostOwnedCapabilityApplyAdapterReadiness
func BuildHostOwnedCapabilityApplyAdapterInvocation(HostOwnedCapabilityApplyAdapterInvocationInput) HostOwnedCapabilityApplyAdapterInvocation

type SchedulerApplyAction string
func BuildHostOwnedSchedulerApplyDescriptor(HostOwnedSchedulerApplyDescriptor) HostOwnedSchedulerApplyDescriptor
func BuildHostOwnedSchedulerApplyRequest(HostOwnedSchedulerApplyRequestInput) HostOwnedSchedulerApplyRequest
func BuildHostOwnedSchedulerApplyResult(HostOwnedSchedulerApplyResultInput) HostOwnedSchedulerApplyResult
func BuildHostOwnedSchedulerApplyReadback(HostOwnedSchedulerApplyReadbackInput) HostOwnedSchedulerApplyReadback
func BuildHostOwnedSchedulerApplyAdapterReadiness(HostOwnedSchedulerApplyAdapterReadinessInput) HostOwnedSchedulerApplyAdapterReadiness
func BuildHostOwnedSchedulerApplyAdapterInvocation(HostOwnedSchedulerApplyAdapterInvocationInput) HostOwnedSchedulerApplyAdapterInvocation
```

Capability与scheduler API只核对显式descriptor、action、final intensity gate、approval、
dry-run、幂等、结果与readback ref；所有`Core*Executed`、dispatch和mutation flag保持
false。具体安装器、package manager、scheduler backend、取消/删除实现与真实readback仍
由Host拥有。

Repeated-success memory proposal只从结构化、已成功的attempt和strategy catalog生成
proposal；review/apply gate继续要求显式Host审核和独立memory gate：

```go
type MemoryProposalKind string
type RepeatedSuccessMemoryProposalSet struct { ... }
type MemoryProposalReviewPacket struct { ... }

func BuildRepeatedSuccessMemoryProposal(RepeatedSuccessMemoryProposalInput) RepeatedSuccessMemoryProposalSet
func BuildMemoryProposalReviewPacket(MemoryProposalReviewPacketInput) MemoryProposalReviewPacket
func BuildHostOwnedMemoryProposalApplyReadiness(HostOwnedMemoryProposalApplyReadinessInput) HostOwnedMemoryProposalApplyReadiness
func BuildHostOwnedMemoryProposalApplyInvocation(HostOwnedMemoryProposalApplyInvocationInput) HostOwnedMemoryProposalApplyInvocation
```

它不会写skill、workflow或template，不执行install/reload，也不拥有memory backend。
## Objective Runtime step、delegation 与 Host executor合同

Objective runtime step把已经准备好的Run、verification、ledger patch与可选auto-delegation
合同归并为下一步纯状态投影：

```go
type AutoDelegationPolicy struct { ... }
type AutoDelegationPlan struct { ... }
type AutoDelegationParentMerge struct { ... }
type ObjectiveRuntimeLoopInput struct { ... }
type ObjectiveRuntimeLoopStep struct { ... }

func BuildAutoDelegationPolicyReview(AutoDelegationPolicy) AutoDelegationPolicyReview
func BuildAutoDelegationPlanReview(AutoDelegationPolicyReview, AutoDelegationPlan) AutoDelegationPlanReview
func BuildAutoDelegationHostBridge(AutoDelegationHostBridgeInput) AutoDelegationHostBridge
func BuildAutoDelegationParentMerge(AutoDelegationParentMergeInput) AutoDelegationParentMerge
func BuildObjectiveRuntimeLoopStep(ObjectiveRuntimeLoopInput) ObjectiveRuntimeLoopStep
```

planner candidate JSON使用严格解码和行为固定的fenced-JSON抽取；未知字段、尾随内容、unsafe
raw output和不完整plan都会fail closed。本包不生成planner prompt、不调用模型，也不创建
真实child session。

Host-owned executor合同只描述请求、结果、readback与adapter调用报告：

```go
type HostOwnedObjectiveExecutorStepRequest struct { ... }
type HostOwnedObjectiveExecutorStepResult struct { ... }
type HostOwnedObjectiveExecutorStepReadback struct { ... }
type ObjectiveRuntimeProductizationReport struct { ... }

func BuildHostOwnedObjectiveExecutorStepRequest(HostOwnedObjectiveExecutorStepRequestInput) HostOwnedObjectiveExecutorStepRequest
func BuildHostOwnedObjectiveExecutorStepResult(HostOwnedObjectiveExecutorStepResultInput) HostOwnedObjectiveExecutorStepResult
func BuildHostOwnedObjectiveExecutorStepReadback(HostOwnedObjectiveExecutorStepReadbackInput) HostOwnedObjectiveExecutorStepReadback
func BuildHostOwnedObjectiveExecutorAdapterReadiness(HostOwnedObjectiveExecutorAdapterReadinessInput) HostOwnedObjectiveExecutorAdapterReadiness
func BuildHostOwnedObjectiveExecutorAdapterInvocation(HostOwnedObjectiveExecutorAdapterInvocationInput) HostOwnedObjectiveExecutorAdapterInvocation
func BuildObjectiveRuntimeProductization(ObjectiveRuntimeProductizationInput) ObjectiveRuntimeProductizationReport
```

所有`Core*Executed`、dispatch、scheduler、store mutation和compensation flag保持false；
`ReadyForHostExecution`或`ReadyForHostProductization`只表示显式Host输入满足合同，不执行、
不授权、不持久化。concrete executor/adapter、authorization、RunStore和durable write仍由Host
拥有。

已结构化Observation可直接进入portable normalizer：

```go
func BuildStructuredObservationNormalization(StructuredObservationNormalizationInput) ObservationNormalizationResult
```

它不接受具体`RuntimeAdapterExecutionResult`。provider/backend输出到结构化Observation的
翻译仍是Host职责。

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

本包不保存可变全局状态，纯函数可由多个 goroutine 并发调用。只有调用Host port的
`BuildObjectiveSpecWithBuilder`、`BuildObjectiveGraphWithPlanner`与
`BuildObjectiveSemanticVerification`接收
`context.Context`，并把取消/deadline传给Host实现；其余计算是内存中的有界同步判定。
Shutdown属于上层 Runtime/Host 合同。

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
`v0.0.0-20260801180223-57ea36658ea2`，不使用 `replace`，也不 import HS、Runner、
Scene、provider 或 backend。它组合 managed-objective projection、retry budget、
lifecycle transition、unsafe-ref fail-closed、单节点Objective Graph validation，以及
Objective evidence verification/recovery proposal、Host effect gate，以及Objective runtime/
executor/productization fail-closed路径：

```bash
cd runtime/conformance/controlcontract-consumer
GOWORK=off go test ./...
GOWORK=off go run .
```

预期输出：

```text
agentx-controlcontract-ok:ready_for_host_action:1:applied:evidence_weak:objective_graph_ready:objective_verification_recovery_ready:host_effect_gate_ready:objective_runtime_contract_ready
```

## 非目标

- 不提供具体Objective executor、runtime loop dispatch、scheduler backend/execution或RunStore；
- 不提供具体runtime/production adapter或adapter-result normalization input翻译；
- 不创建真实delegation worker/session，不执行child dispatch或parent durable merge；
- 不探测真实capability，不选择provider，不执行strategy或graph node；
- 不执行 approval、authorization、sandbox、model/tool 或 backend；
- 不包含 ProductShellRuntime、Scene、provider、credential 或真实网络；
- 不构成 Public、Beta、Stable、production-ready 或正式发行声明。
