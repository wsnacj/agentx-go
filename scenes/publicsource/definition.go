package publicsource

import (
	agentxpack "github.com/wsnacj/agentx-go/extensions/pack"
	agentxexecution "github.com/wsnacj/agentx-go/runtime/executionpolicy"
	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

const (
	PackID          = "public-source-readonly-pack"
	CaseTypeAcquire = "public_source.acquire"
	DefaultWorkflow = "public_source_acquire_v1"
	AcquireTool     = "public_source_acquire"
)

func Definition() agentxpack.Definition {
	return agentxpack.Definition{
		Manifest:       agentxpack.Manifest{ID: PackID, Version: "0.1.0", Domain: "public_source_readonly", RouteHints: []string{"公开来源", "读取网页", "public source", "read source"}, SupportedCaseTypes: []string{CaseTypeAcquire}, DefaultWorkflow: DefaultWorkflow, PolicyProfiles: []string{"public_source_readonly"}, Evaluators: []string{"public_source_evidence_guard"}, EvalSuites: []string{"public_source_readonly_success"}},
		CaseSchemas:    []agentxpack.CaseSchema{{CaseType: CaseTypeAcquire, Description: "读取宿主明确批准的公开来源并验证 display-safe evidence。", RouteHints: []string{"公开网页", "来源验证", "source evidence"}, Schema: map[string]any{"type": "object", "properties": map[string]any{"query_ref": map[string]any{"type": "string"}, "source_ref": map[string]any{"type": "string"}, "require_attestation": map[string]any{"type": "boolean"}}, "required": []string{"query_ref"}}}},
		Workflows:      []agentxworkflow.Spec{{ID: DefaultWorkflow, Title: "Public Source Read-only Acquisition", Description: "通过宿主注入的 Collector 获取公开来源，并以确定性 evidence guard 收口。", Version: "v1", Pack: PackID, CaseTypes: []string{CaseTypeAcquire}, RouteHints: []string{"公开来源读取", "来源证据"}, PlanningMode: agentxworkflow.PlanningBounded, EntryNode: "acquire_source", DefaultContract: "public_source_readonly", StateSchema: []agentxworkflow.StateSlotSpec{{Name: "source.evidence_observed", Type: "boolean", Required: true}, {Name: "source.passed", Type: "boolean", Required: true}, {Name: "source.failure_reasons", Type: "array"}, {Name: "source.summary", Type: "string"}}, EvaluatorSchema: []agentxworkflow.EvaluatorRef{{Name: "public_source_evidence_guard"}}, Nodes: []agentxworkflow.NodeSpec{{ID: "acquire_source", Kind: agentxworkflow.NodeTool, Title: "Acquire public source", Description: "只调用宿主明确注入的 source Collector；网络、provider、凭据与安全策略由宿主负责。", Inputs: []agentxworkflow.BindingSpec{{From: "case.input.query_ref", To: "args.query_ref"}, {From: "case.input.source_ref", To: "args.source_ref"}, {From: "case.input.require_attestation", To: "args.require_attestation"}}, Outputs: []agentxworkflow.BindingSpec{{From: "result.evidence_observed", To: "state.source.evidence_observed"}, {From: "result.passed", To: "state.source.passed"}, {From: "result.failure_reasons", To: "state.source.failure_reasons"}, {From: "result.summary", To: "state.source.summary"}}, Config: map[string]any{"tool_name": AcquireTool, "args": map[string]any{"mode": "readonly"}}}}}},
		Tools:          []agentxpack.SemanticTool{{Name: AcquireTool, Description: "公开来源只读采集入口。宿主负责网络、provider、allowlist、凭据和内容授权。", RuntimeTool: AcquireTool, Tags: []string{"public-source", "read-only", "host-adapter"}}},
		Evaluators:     []agentxpack.Evaluator{{Name: "public_source_evidence_guard", Description: "验证公开来源 evidence、display-safe summary、attestation 与无 raw output。", OutputSchema: map[string]any{"type": "object", "properties": map[string]any{"passed": map[string]any{"type": "boolean"}, "evidence_observed": map[string]any{"type": "boolean"}, "attested_summary_observed": map[string]any{"type": "boolean"}, "raw_output_absent": map[string]any{"type": "boolean"}, "failure_reasons": map[string]any{"type": "array", "items": map[string]any{"type": "string"}}}, "required": []string{"passed", "evidence_observed", "raw_output_absent"}}}},
		EvalSuites:     []agentxpack.EvalSuite{{Name: "public_source_readonly_success", Description: "要求可验证来源 evidence 且无 raw output 或写副作用。", Mode: agentxpack.EvalSuiteModeGate, WorkflowIDs: []string{DefaultWorkflow}, RequiredState: []string{"source.evidence_observed", "source.passed"}, PassPath: "source.passed", SummaryPath: "source.summary", Default: true}},
		PolicyProfiles: []agentxpack.PolicyProfile{{Name: "public_source_readonly", Contract: agentxexecution.Contract{ID: "public-source-readonly", Strict: true, Version: 1, Visibility: agentxexecution.VisibilityPolicy{AllowTools: []string{AcquireTool}, DeclaredTools: []string{AcquireTool}, RequireDeclared: true, MaxRisk: "medium"}, Budget: agentxexecution.BudgetPolicy{MaxToolCalls: 1}, Loop: agentxexecution.LoopPolicy{MaxRounds: 1, LoopDetectionEnabled: true, ToolFailureFuseEnabled: true, ToolFailureFuseThreshold: 1}, SideEffects: agentxexecution.SideEffectPolicy{MaxClass: agentxexecution.SideEffectReadOnly, StrictRecovery: true}, Audit: agentxexecution.AuditPolicy{PersistSnapshot: true}}, Default: true}},
	}
}

func PackDefinition() agentxpack.Definition { return Definition() }
func MaterializedDefaultWorkflow(coordinator *agentxpack.Coordinator) (agentxworkflow.Spec, error) {
	return coordinator.MaterializeWorkflow(Definition(), DefaultWorkflow)
}
func RegisterInto(reg agentxpack.Registry) error {
	if reg == nil {
		return nil
	}
	return reg.Register(Definition())
}
