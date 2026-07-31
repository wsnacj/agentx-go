// Package validation 提供 portable Workflow structural validation
// orchestration、graph/binding kernel 和显式 host policy port。
package validation

import workflow "github.com/wsnacj/agentx-go/runtime/workflow"

// Policy 保留 host-owned product/runtime policy 的九个既有调用位置。
//
// 实现必须保持同步、无后台生命周期；返回的错误会按调用顺序原样传播或由
// structural context 使用 %w 包装。
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
