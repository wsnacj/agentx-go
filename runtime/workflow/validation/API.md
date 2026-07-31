# runtime/workflow/validation API

导入路径：

```go
import validation "github.com/wsnacj/agentx-go/runtime/workflow/validation"
```

成熟度：**Experimental / private validation**。

该 package 是 portable Workflow structural validation 的 canonical
implementation owner，负责 Spec/Node orchestration、graph reachability、
cycle detection、binding path/reference 和 field shape mechanism。

它不拥有 pack、node config、runtime capability、binding root、provider 或
backend policy。调用方必须显式提供 `Policy`。

## Policy

```go
type Policy interface {
    ValidatePackScopedContractUsage(workflow.Spec) error
    ValidatePackScopedWorkflowMetadataUsage(workflow.Spec) error
    ValidateNodeRuntimeCapabilities(workflow.NodeSpec) error
    ValidateNodeConfig(workflow.NodeSpec) error
    ValidateEdgeRuntimeCapabilities(workflow.EdgeSpec, string, string) error
    ValidateLinearRuntimeEdgeDeterminism([]workflow.EdgeSpec) error
    ValidateReachableCycleRuntimeCapability(string) error
    ValidateBindingTargetShape(string, string) error
    ValidateBindingSourceShape(string, string) error
}
```

九个方法精确对应 structural mechanism 的既有 policy 调用位置，以保持
first-error precedence。它不是通用 hook registry，也不允许 policy 取得
graph、dominator 或其它内部状态。

package 不提供默认或 permissive policy。host 必须明确实现全部方法；把方法
简单返回 `nil` 是调用方自己的产品决定，不代表 AgentX 支持对应 runtime
能力。

## ValidateSpec

```go
func ValidateSpec(spec workflow.Spec, policy Policy) error
```

验证顺序保持为：

```text
spec fields
pack/metadata policy
planning mode / entry / state slots
nodes and node policy
edges and edge policy
reachability / cycle policy
binding source references
```

policy error 在顶层调用点原样返回；node/binding context 使用 `%w` 包装，因此
`errors.Is/As` identity 保持。nil policy 返回：

```text
workflow validation: policy is required
```

## ValidateNodeSpec

```go
func ValidateNodeSpec(node workflow.NodeSpec, policy Policy) error
```

该函数用于 lowering 等需要单节点 admission 的 host path。它验证 field、
known node kind、input/output binding structural shape，并按既有顺序调用
node runtime、binding target 和 node config policy。

## Field helpers

```go
func ValidateTrimmedField(value, label string) error
func ValidateOptionalField(value, label string) error
```

这两个 helper 使 host policy 与 canonical structural orchestration 共享同一
whitespace/error-text contract，避免 HS 维护第二份实现。

## 并发与生命周期

package 不保存全局状态、不创建 goroutine，也不提供 Shutdown。每次调用使用
独立 graph。是否可并发取决于调用方提供的 Policy；package 不为非并发安全
policy 增加锁。

## 非目标

- 不提供默认 HS/Scene/product policy；
- 不验证 JSON Schema definition；该能力属于 `workflow/schema`；
- 不执行 lowering、node execution、retry、resume 或 durable lifecycle；
- 不提供 provider、credential、Scene、具体 backend 或完整 Runtime；
- 不构成 Public、Beta、Stable 或 production-ready 声明。
