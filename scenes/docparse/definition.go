package docparse

import (
	agentxpack "github.com/wsnacj/agentx-go/extensions/pack"
	agentxexecution "github.com/wsnacj/agentx-go/runtime/executionpolicy"
	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"
	docparsehostkit "github.com/wsnacj/agentx-go/scenes/docparse/hostkit"
)

const (
	PackID = "docparse-document-pack"

	CaseTypeProfileProbe  = "document.profile_probe"
	CaseTypeExtractFields = "document.extract_fields"
	CaseTypeExtractTable  = "document.extract_table"
	CaseTypeVerifyFields  = "document.verify_fields"
	CaseTypeEvidenceTrace = "document.evidence_trace"
	CaseTypeGuard         = "document.guard"

	ProfileProbeWorkflow  = "docparse_profile_probe_v1"
	ExtractFieldsWorkflow = "docparse_extract_fields_v1"
	ExtractTableWorkflow  = "docparse_extract_table_v1"
	VerifyFieldsWorkflow  = "docparse_verify_fields_v1"
	EvidenceTraceWorkflow = "docparse_evidence_trace_v1"
	GuardWorkflow         = "docparse_guard_v1"

	DefaultPolicy       = "docparse_workspace_document_review"
	DefaultEvaluator    = "docparse_evidence_guard"
	DefaultSuccessSuite = "docparse_document_success_suite"

	ArtifactTypeEvidenceBundle = "docparse.evidence_bundle"
)

func Definition() agentxpack.Definition {
	return agentxpack.Definition{
		Manifest: agentxpack.Manifest{
			ID:      PackID,
			Version: "0.1.0",
			Domain:  "document_operations",
			RouteHints: []string{
				"document parse",
				"document extraction",
				"evidence trace",
				"字段抽取",
				"表格抽取",
				"证据回溯",
				"文档解析",
			},
			SupportedCaseTypes: supportedCaseTypes(),
			DefaultWorkflow:    ExtractFieldsWorkflow,
			PolicyProfiles:     []string{DefaultPolicy},
			Evaluators:         []string{DefaultEvaluator},
			EvalSuites:         []string{DefaultSuccessSuite},
		},
		CaseSchemas: caseSchemas(),
		Workflows: []agentxworkflow.Spec{
			toolWorkflow(ProfileProbeWorkflow, "Document Profile Probe", CaseTypeProfileProbe, docparsehostkit.ToolDocparseProfileProbe),
			toolWorkflow(ExtractFieldsWorkflow, "Document Field Extraction", CaseTypeExtractFields, docparsehostkit.ToolDocparseExtractFields),
			toolWorkflow(ExtractTableWorkflow, "Document Table Extraction", CaseTypeExtractTable, docparsehostkit.ToolDocparseExtractTable),
			toolWorkflow(VerifyFieldsWorkflow, "Document Field Verification", CaseTypeVerifyFields, docparsehostkit.ToolDocparseValidate),
			toolWorkflow(EvidenceTraceWorkflow, "Document Evidence Trace", CaseTypeEvidenceTrace, docparsehostkit.ToolDocparseTraceEvidence),
			toolWorkflow(GuardWorkflow, "Document Evidence Guard", CaseTypeGuard, docparsehostkit.ToolDocparseGuard),
		},
		Tools:          semanticTools(),
		Evaluators:     evaluators(),
		EvalSuites:     evalSuites(),
		PolicyProfiles: []agentxpack.PolicyProfile{policyProfile()},
		MemorySchemas:  []agentxpack.MemorySchema{memorySchema()},
		MemoryRecallPolicy: &agentxpack.MemoryRecallPolicy{
			QueryHints: []string{"文档解析", "字段抽取", "表格抽取", "evidence trace", "document extraction"},
			Limit:      4,
			MaxChars:   1600,
			ScopedOnly: true,
		},
	}
}

func supportedCaseTypes() []string {
	return []string{
		CaseTypeProfileProbe,
		CaseTypeExtractFields,
		CaseTypeExtractTable,
		CaseTypeVerifyFields,
		CaseTypeEvidenceTrace,
		CaseTypeGuard,
	}
}

func caseSchemas() []agentxpack.CaseSchema {
	return []agentxpack.CaseSchema{
		{
			CaseType:    CaseTypeProfileProbe,
			Description: "对未知或待确认文档执行轻量类型判断，输出 review-required profile proposal、候选类型、建议字段和 route hints；不抽取最终字段或完整表格 rows。",
			RouteHints:  []string{"文档类型判断", "profile probe", "document classification"},
			Schema:      documentCaseSchema(),
		},
		{
			CaseType:    CaseTypeExtractFields,
			Description: "从 host-approved local document、result fixture 或 caller-provided spec 中抽取字段，并输出 evidence bundle、diagnostics、readiness 与 review flags。",
			RouteHints:  []string{"字段抽取", "extract fields", "document fields"},
			Schema:      documentCaseSchema(),
		},
		{
			CaseType:    CaseTypeExtractTable,
			Description: "从 host-approved local document 或 result fixture 中抽取表格结构，并保留 row/column/cell evidence。",
			RouteHints:  []string{"表格抽取", "extract table", "table evidence"},
			Schema:      documentCaseSchema(),
		},
		{
			CaseType:    CaseTypeVerifyFields,
			Description: "对已抽取字段做 expected value、period、unit、bbox/cell requirement 和 review-required 验证。",
			RouteHints:  []string{"字段校验", "verify fields", "validate extraction"},
			Schema:      documentCaseSchema(),
		},
		{
			CaseType:    CaseTypeEvidenceTrace,
			Description: "输出字段和表格到 page refs、bbox、table cells、source snippet 的可审计 trace。",
			RouteHints:  []string{"证据回溯", "evidence trace", "source grounding"},
			Schema:      documentCaseSchema(),
		},
		{
			CaseType:    CaseTypeGuard,
			Description: "将文档解析证据转换为 answer-ready、review-required 或 failed 的可交付边界。",
			RouteHints:  []string{"文档 guard", "readiness", "review required"},
			Schema:      documentCaseSchema(),
		},
	}
}

func documentCaseSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"user_message": map[string]any{"type": "string"},
			"document_path": map[string]any{
				"type":        "string",
				"description": "Workspace-local or host-approved document artifact reference.",
			},
			"result_path":                      map[string]any{"type": "string"},
			"parse_result":                     map[string]any{"type": "object", "additionalProperties": true},
			"document_type":                    map[string]any{"type": "string"},
			"expected_document_type":           map[string]any{"type": "string"},
			"profile_id":                       map[string]any{"type": "string"},
			"spec_path":                        map[string]any{"type": "string"},
			"spec_paths":                       stringArraySchema("Caller-provided candidate spec paths."),
			"page_range":                       map[string]any{"type": "string", "description": "Optional 1-based contiguous page range such as 2 or 2-4 for host-approved bundled documents."},
			"requested_fields":                 stringArraySchema("Requested field keys or output columns."),
			"required_fields":                  stringArraySchema("Required field keys."),
			"expected_fields":                  objectArraySchema("Expected field assertions."),
			"expected_tables":                  objectArraySchema("Expected table assertions."),
			"required_evidence":                map[string]any{"type": "object", "additionalProperties": true},
			"require_page_refs":                map[string]any{"type": "boolean"},
			"require_bounding_boxes":           map[string]any{"type": "boolean"},
			"require_bbox":                     map[string]any{"type": "boolean"},
			"require_table_cells":              map[string]any{"type": "boolean"},
			"require_complete_table_rows":      map[string]any{"type": "boolean"},
			"allow_review_required":            map[string]any{"type": "boolean"},
			"output_schema":                    map[string]any{"type": "object", "additionalProperties": true},
			"artifact_policy":                  map[string]any{"type": "string"},
			"profile_probe_only":               map[string]any{"type": "boolean"},
			"classify_only":                    map[string]any{"type": "boolean"},
			"profile_discovery_search":         map[string]any{"type": "boolean"},
			"enable_profile_discovery_search":  map[string]any{"type": "boolean"},
			"disable_profile_discovery_search": map[string]any{"type": "boolean"},
		},
		"required": []string{"user_message"},
	}
}

func toolWorkflow(id string, title string, caseType string, toolName string) agentxworkflow.Spec {
	return agentxworkflow.Spec{
		ID:              id,
		Title:           title,
		Description:     "执行 docparse module semantic tool；host 继续拥有 parser executor、private schema、OCR/model credential 和 artifact/export policy。",
		Version:         "v1",
		Pack:            PackID,
		CaseTypes:       []string{caseType},
		RouteHints:      []string{caseType, title},
		PlanningMode:    agentxworkflow.PlanningBounded,
		EntryNode:       "run_docparse",
		DefaultContract: DefaultPolicy,
		StateSchema:     workflowStateSchema(),
		EvaluatorSchema: []agentxworkflow.EvaluatorRef{{Name: DefaultEvaluator}},
		Nodes: []agentxworkflow.NodeSpec{
			{
				ID:          "run_docparse",
				Kind:        agentxworkflow.NodeTool,
				Title:       title,
				Description: "调用 docparse semantic tool；parser executor 仍由 host 注入，本包只处理通用 evidence projection、validation 和 guard。",
				Inputs: []agentxworkflow.BindingSpec{
					{From: "case.input.user_message", To: "args.user_message"},
					optionalInput("case.input.document_path", "args.document_path"),
					optionalInput("case.input.result_path", "args.result_path"),
					optionalInput("case.input.parse_result", "args.parse_result"),
					optionalInput("case.input.document_type", "args.document_type"),
					optionalInput("case.input.expected_document_type", "args.expected_document_type"),
					optionalInput("case.input.profile_id", "args.profile_id"),
					optionalInput("case.input.spec_path", "args.spec_path"),
					optionalInput("case.input.spec_paths", "args.spec_paths"),
					optionalInput("case.input.page_range", "args.page_range"),
					optionalInput("case.input.requested_fields", "args.requested_fields"),
					optionalInput("case.input.required_fields", "args.required_fields"),
					optionalInput("case.input.expected_fields", "args.expected_fields"),
					optionalInput("case.input.expected_tables", "args.expected_tables"),
					optionalInput("case.input.required_evidence", "args.required_evidence"),
					optionalInput("case.input.require_page_refs", "args.require_page_refs"),
					optionalInput("case.input.require_bounding_boxes", "args.require_bounding_boxes"),
					optionalInput("case.input.require_bbox", "args.require_bbox"),
					optionalInput("case.input.require_table_cells", "args.require_table_cells"),
					optionalInput("case.input.require_complete_table_rows", "args.require_complete_table_rows"),
					optionalInput("case.input.allow_review_required", "args.allow_review_required"),
					optionalInput("case.input.output_schema", "args.output_schema"),
					optionalInput("case.input.artifact_policy", "args.artifact_policy"),
					optionalInput("case.input.profile_probe_only", "args.profile_probe_only"),
					optionalInput("case.input.classify_only", "args.classify_only"),
					optionalInput("case.input.profile_discovery_search", "args.profile_discovery_search"),
					optionalInput("case.input.enable_profile_discovery_search", "args.enable_profile_discovery_search"),
					optionalInput("case.input.disable_profile_discovery_search", "args.disable_profile_discovery_search"),
				},
				Outputs: workflowOutputs(),
				Config: map[string]any{
					"tool_name": toolName,
					"args": map[string]any{
						"task_kind":      caseType,
						"source_policy":  "host_approved_workspace_documents",
						"stop_condition": "evidence_ready_or_review_required",
					},
				},
			},
		},
	}
}

func optionalInput(from string, to string) agentxworkflow.BindingSpec {
	return agentxworkflow.BindingSpec{From: from, To: to, Optional: true}
}

func workflowStateSchema() []agentxworkflow.StateSlotSpec {
	return []agentxworkflow.StateSlotSpec{
		{Name: "docparse.status", Type: "string", Required: true},
		{Name: "docparse.adapter_status", Type: "string"},
		{Name: "docparse.failure_code", Type: "string"},
		{Name: "docparse.failure_class", Type: "string"},
		{Name: "docparse.field_count", Type: "integer"},
		{Name: "docparse.table_count", Type: "integer"},
		{Name: "docparse.evidence_complete", Type: "boolean"},
		{Name: "docparse.review_required", Type: "boolean"},
		{Name: "docparse.passed", Type: "boolean", Required: true},
		{Name: "docparse.summary", Type: "string"},
	}
}

func workflowOutputs() []agentxworkflow.BindingSpec {
	return []agentxworkflow.BindingSpec{
		{From: "result.status", To: "state.docparse.status"},
		optionalOutput("result.adapter_status", "state.docparse.adapter_status"),
		optionalOutput("result.failure_code", "state.docparse.failure_code"),
		optionalOutput("result.failure_class", "state.docparse.failure_class"),
		optionalOutput("result.field_count", "state.docparse.field_count"),
		optionalOutput("result.table_count", "state.docparse.table_count"),
		{From: "result.evidence_complete", To: "state.docparse.evidence_complete"},
		optionalOutput("result.review_required", "state.docparse.review_required"),
		{From: "result.passed", To: "state.docparse.passed"},
		{From: "result.summary", To: "state.docparse.summary"},
	}
}

func optionalOutput(from string, to string) agentxworkflow.BindingSpec {
	return agentxworkflow.BindingSpec{From: from, To: to, Optional: true}
}

func semanticTools() []agentxpack.SemanticTool {
	return []agentxpack.SemanticTool{
		{
			Name:        docparsehostkit.ToolDocparseSpecSelect,
			Description: "对 caller-provided spec candidates 做只读排序或显式选择；host/livekit 可代理到 document_spec_recommend。",
			RuntimeTool: docparsehostkit.ToolDocparseSpecSelect,
			Tags:        []string{"docparse", "spec", "readonly"},
		},
		{
			Name:        docparsehostkit.ToolDocparseProfileProbe,
			Description: "对未知或待确认文档做 classify-only profile probe，输出 review-required proposal，不输出最终字段或完整表格明细。",
			RuntimeTool: docparsehostkit.ToolDocparseProfileProbe,
			Tags:        []string{"docparse", "profile_probe", "classification", "review_required"},
		},
		{
			Name:        docparsehostkit.ToolDocparseExtractFields,
			Description: "调用 host-injected parser executor 或 opt-in document_parse，输出 field evidence payload。",
			RuntimeTool: docparsehostkit.ToolDocparseExtractFields,
			Tags:        []string{"docparse", "fields", "adapter", "evidence_payload"},
		},
		{
			Name:        docparsehostkit.ToolDocparseExtractTable,
			Description: "调用 host-injected parser executor 或 opt-in document_parse，输出 table evidence payload。",
			RuntimeTool: docparsehostkit.ToolDocparseExtractTable,
			Tags:        []string{"docparse", "tables", "adapter", "evidence_payload"},
		},
		{
			Name:        docparsehostkit.ToolDocparseTraceEvidence,
			Description: "从 parse result/result fixture 投影字段和表格 evidence trace payload。",
			RuntimeTool: docparsehostkit.ToolDocparseTraceEvidence,
			Tags:        []string{"docparse", "evidence", "evaluator", "evidence_payload"},
		},
		{
			Name:        docparsehostkit.ToolDocparseValidate,
			Description: "对字段、表格、bbox/cell 和 required evidence 做验证。",
			RuntimeTool: docparsehostkit.ToolDocparseValidate,
			Tags:        []string{"docparse", "validation", "evaluator", "evidence_payload"},
		},
		{
			Name:        docparsehostkit.ToolDocparseGuard,
			Description: "统一 readiness、review_required、failure_class 和 answer boundary。",
			RuntimeTool: docparsehostkit.ToolDocparseGuard,
			Tags:        []string{"docparse", "guard", "evaluator", "evidence_payload"},
		},
	}
}

func evaluators() []agentxpack.Evaluator {
	return []agentxpack.Evaluator{
		{
			Name:         DefaultEvaluator,
			Description:  "检查文档解析输出是否 evidence complete、failure_class 可解释、review_required 明确，并避免把缺证据结果当成可交付答案。",
			OutputSchema: evaluatorOutputSchema(),
		},
	}
}

func evaluatorOutputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"passed":            map[string]any{"type": "boolean"},
			"status":            map[string]any{"type": "string"},
			"failure_class":     map[string]any{"type": "string"},
			"failure_code":      map[string]any{"type": "string"},
			"review_required":   map[string]any{"type": "boolean"},
			"evidence_complete": map[string]any{"type": "boolean"},
			"field_count":       map[string]any{"type": "integer"},
			"table_count":       map[string]any{"type": "integer"},
			"summary":           map[string]any{"type": "string"},
		},
		"required":             []string{"passed", "status"},
		"additionalProperties": true,
	}
}

func evalSuites() []agentxpack.EvalSuite {
	return []agentxpack.EvalSuite{
		{
			Name:        DefaultSuccessSuite,
			Description: "要求 docparse workflow 输出可解释状态、evidence_complete、failure_class/review_required，并保留 readiness summary；真实 artifact/export policy 仍由 host 持有。",
			Mode:        agentxpack.EvalSuiteModeGate,
			WorkflowIDs: []string{
				ProfileProbeWorkflow,
				ExtractFieldsWorkflow,
				ExtractTableWorkflow,
				VerifyFieldsWorkflow,
				EvidenceTraceWorkflow,
				GuardWorkflow,
			},
			RequiredState: []string{
				"docparse.status",
				"docparse.passed",
				"docparse.evidence_complete",
				"docparse.summary",
			},
			PassPath:    "docparse.passed",
			SummaryPath: "docparse.summary",
			Default:     true,
		},
	}
}

func policyProfile() agentxpack.PolicyProfile {
	tools := append([]string{}, docparsehostkit.ToolNames()...)
	tools = append(tools, docparsehostkit.RuntimeToolDocumentParse, docparsehostkit.RuntimeToolDocumentSpecRecommend)
	return agentxpack.PolicyProfile{
		Name: DefaultPolicy,
		Contract: agentxexecution.Contract{
			ID:      "docparse-workspace-document-review",
			Strict:  true,
			Version: 1,
			Visibility: agentxexecution.VisibilityPolicy{
				AllowTools:      tools,
				DeclaredTools:   tools,
				RequireDeclared: true,
				MaxRisk:         "medium",
			},
			Budget: agentxexecution.BudgetPolicy{MaxToolCalls: 8, MaxToolResultChars: 120000},
			Loop: agentxexecution.LoopPolicy{
				MaxRounds:                8,
				LoopDetectionEnabled:     true,
				ToolFailureFuseEnabled:   true,
				ToolFailureFuseThreshold: 3,
			},
			SideEffects: agentxexecution.SideEffectPolicy{
				MaxClass:       agentxexecution.SideEffectWorkspaceWrite,
				StrictRecovery: true,
			},
			Audit: agentxexecution.AuditPolicy{PersistSnapshot: true},
		},
		Default: true,
	}
}

func memorySchema() agentxpack.MemorySchema {
	return agentxpack.MemorySchema{
		Name:        "docparse_document_memory",
		Description: "沉淀文档解析 case、workflow、readiness、failure_class、review_required 和 evidence summary。",
		Default:     true,
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"pack_id":         map[string]any{"type": "string"},
				"case_type":       map[string]any{"type": "string"},
				"workflow_id":     map[string]any{"type": "string"},
				"status":          map[string]any{"type": "string"},
				"failure_class":   map[string]any{"type": "string"},
				"review_required": map[string]any{"type": "boolean"},
				"field_count":     map[string]any{"type": "integer"},
				"table_count":     map[string]any{"type": "integer"},
				"summary":         map[string]any{"type": "string"},
			},
			"required": []string{"pack_id", "case_type", "workflow_id", "status", "summary"},
		},
	}
}

func stringArraySchema(description string) map[string]any {
	return map[string]any{
		"type":        "array",
		"items":       map[string]any{"type": "string"},
		"description": description,
	}
}

func objectArraySchema(description string) map[string]any {
	return map[string]any{
		"type":        "array",
		"items":       map[string]any{"type": "object", "additionalProperties": true},
		"description": description,
	}
}

func RegisterInto(reg agentxpack.Registry) error {
	if reg == nil {
		return nil
	}
	return reg.Register(Definition())
}
