package publictransport

import (
	agentxpack "github.com/wsnacj/agentx-go/extensions/pack"
	agentxexecution "github.com/wsnacj/agentx-go/runtime/executionpolicy"
	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

const (
	PackID               = "public-transport-readonly-pack"
	CaseTypeTicketLookup = "public_transport.ticket_lookup"
	DefaultWorkflow      = "public_transport_ticket_lookup_v1"
	LookupTool           = "public_transport_ticket_lookup"
)

// Definition returns the portable read-only public-transport Pack.
func Definition() agentxpack.Definition {
	return agentxpack.Definition{
		Manifest: agentxpack.Manifest{
			ID: PackID, Version: "0.1.0", Domain: "public_transport_readonly",
			RouteHints:         []string{"高铁余票", "火车票查询", "train availability", "rail fare"},
			SupportedCaseTypes: []string{CaseTypeTicketLookup}, DefaultWorkflow: DefaultWorkflow,
			PolicyProfiles: []string{"public_transport_readonly"}, Evaluators: []string{"public_transport_inventory_guard"},
			EvalSuites: []string{"public_transport_readonly_success"},
		},
		CaseSchemas: []agentxpack.CaseSchema{{
			CaseType:    CaseTypeTicketLookup,
			Description: "查询宿主批准的公共交通只读库存，不执行订票、购票或支付。",
			RouteHints:  []string{"余票", "车次", "票价", "train ticket"},
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"from": map[string]any{"type": "string"}, "to": map[string]any{"type": "string"},
					"travel_date":             map[string]any{"type": "string"},
					"required_train_prefixes": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
					"required_seat_tokens":    map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
					"require_fare_evidence":   map[string]any{"type": "boolean"},
				},
				"required": []string{"from", "to", "travel_date"},
			},
		}},
		Workflows: []agentxworkflow.Spec{{
			ID: DefaultWorkflow, Title: "Public Transport Read-only Ticket Lookup",
			Description: "调用宿主注入的公共交通只读 Collector，并以确定性 evidence guard 验证结果。",
			Version:     "v1", Pack: PackID, CaseTypes: []string{CaseTypeTicketLookup},
			RouteHints: []string{"余票查询", "车次查询", "票价查询"}, PlanningMode: agentxworkflow.PlanningBounded,
			EntryNode: "lookup_inventory", DefaultContract: "public_transport_readonly",
			StateSchema: []agentxworkflow.StateSlotSpec{
				{Name: "transport.inventory_observed", Type: "boolean", Required: true},
				{Name: "transport.passed", Type: "boolean", Required: true},
				{Name: "transport.no_booking_or_purchase", Type: "boolean", Required: true},
				{Name: "transport.rows", Type: "array"},
				{Name: "transport.failure_reasons", Type: "array"},
				{Name: "transport.summary", Type: "string"},
			},
			EvaluatorSchema: []agentxworkflow.EvaluatorRef{{Name: "public_transport_inventory_guard"}},
			Nodes: []agentxworkflow.NodeSpec{{
				ID: "lookup_inventory", Kind: agentxworkflow.NodeTool, Title: "Lookup public transport inventory",
				Description: "通过宿主明确注入的只读适配器查询库存；provider、凭据、网络和限流均由宿主负责。",
				Inputs: []agentxworkflow.BindingSpec{
					{From: "case.input.from", To: "args.from"}, {From: "case.input.to", To: "args.to"},
					{From: "case.input.travel_date", To: "args.travel_date"},
					{From: "case.input.required_train_prefixes", To: "args.required_train_prefixes"},
					{From: "case.input.required_seat_tokens", To: "args.required_seat_tokens"},
					{From: "case.input.require_fare_evidence", To: "args.require_fare_evidence"},
				},
				Outputs: []agentxworkflow.BindingSpec{
					{From: "result.inventory_observed", To: "state.transport.inventory_observed"},
					{From: "result.passed", To: "state.transport.passed"},
					{From: "result.no_booking_or_purchase", To: "state.transport.no_booking_or_purchase"},
					{From: "result.matching_rows", To: "state.transport.rows"},
					{From: "result.failure_reasons", To: "state.transport.failure_reasons"},
					{From: "result.summary", To: "state.transport.summary"},
				},
				Config: map[string]any{"tool_name": LookupTool, "args": map[string]any{"mode": "readonly", "booking": false, "purchase": false}},
			}},
		}},
		Tools: []agentxpack.SemanticTool{{
			Name:        LookupTool,
			Description: "公共交通只读查询入口。宿主负责 provider、站码、日期策略、凭据、网络、限流、条款和免责声明。",
			RuntimeTool: LookupTool, Tags: []string{"public-transport", "read-only", "host-adapter"},
		}},
		Evaluators: []agentxpack.Evaluator{{
			Name:        "public_transport_inventory_guard",
			Description: "验证库存 evidence、车次/席别/票价约束以及未发生订票或购票。",
			OutputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"passed":                 map[string]any{"type": "boolean"},
					"inventory_observed":     map[string]any{"type": "boolean"},
					"no_booking_or_purchase": map[string]any{"type": "boolean"},
					"matching_rows":          map[string]any{"type": "array"},
					"failure_reasons":        map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
				},
				"required": []string{"passed", "inventory_observed", "no_booking_or_purchase"},
			},
		}},
		EvalSuites: []agentxpack.EvalSuite{{
			Name: "public_transport_readonly_success", Description: "要求可验证库存且无订票、购票或支付副作用。",
			Mode: agentxpack.EvalSuiteModeGate, WorkflowIDs: []string{DefaultWorkflow},
			RequiredState: []string{"transport.inventory_observed", "transport.passed", "transport.no_booking_or_purchase"},
			PassPath:      "transport.passed", SummaryPath: "transport.summary", Default: true,
		}},
		PolicyProfiles: []agentxpack.PolicyProfile{{
			Name: "public_transport_readonly",
			Contract: agentxexecution.Contract{
				ID: "public-transport-readonly", Strict: true, Version: 1,
				Visibility:  agentxexecution.VisibilityPolicy{AllowTools: []string{LookupTool}, DeclaredTools: []string{LookupTool}, RequireDeclared: true, MaxRisk: "medium"},
				Budget:      agentxexecution.BudgetPolicy{MaxToolCalls: 1},
				Loop:        agentxexecution.LoopPolicy{MaxRounds: 1, LoopDetectionEnabled: true, ToolFailureFuseEnabled: true, ToolFailureFuseThreshold: 1},
				SideEffects: agentxexecution.SideEffectPolicy{MaxClass: agentxexecution.SideEffectReadOnly, StrictRecovery: true},
				Audit:       agentxexecution.AuditPolicy{PersistSnapshot: true},
			},
			Default: true,
		}},
	}
}

// PackDefinition is the compatibility spelling used by HS registries.
func PackDefinition() agentxpack.Definition { return Definition() }

// MaterializedDefaultWorkflow validates and lowers the default Workflow through
// an explicitly supplied pack Coordinator.
func MaterializedDefaultWorkflow(coordinator *agentxpack.Coordinator) (agentxworkflow.Spec, error) {
	return coordinator.MaterializeWorkflow(Definition(), DefaultWorkflow)
}

// RegisterInto registers the Definition in a caller-owned registry.
func RegisterInto(reg agentxpack.Registry) error {
	if reg == nil {
		return nil
	}
	return reg.Register(Definition())
}
