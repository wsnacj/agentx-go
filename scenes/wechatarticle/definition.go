package wechatarticle

import (
	agentxpack "github.com/wsnacj/agentx-go/extensions/pack"
	agentxexecution "github.com/wsnacj/agentx-go/runtime/executionpolicy"
	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

const (
	PackID          = "wechat-article-readonly-pack"
	CaseTypeAcquire = "wechat_article.acquire"
	DefaultWorkflow = "wechat_article_search_list_download_v1"
	AcquireTool     = "wechat_article_acquire"
)

func Definition() agentxpack.Definition {
	return agentxpack.Definition{
		Manifest:       agentxpack.Manifest{ID: PackID, Version: "0.1.0", Domain: "wechat_article_readonly", RouteHints: []string{"微信公众号文章", "公众号文章同步", "WeChat article"}, SupportedCaseTypes: []string{CaseTypeAcquire}, DefaultWorkflow: DefaultWorkflow, PolicyProfiles: []string{"wechat_article_readonly"}, Evaluators: []string{"wechat_article_evidence_guard"}, EvalSuites: []string{"wechat_article_readonly_success"}},
		CaseSchemas:    []agentxpack.CaseSchema{{CaseType: CaseTypeAcquire, Description: "通过宿主提供的 exporter 只读搜索、列出并可选下载公众号文章。", RouteHints: []string{"公众号文章", "文章列表", "文章下载"}, Schema: map[string]any{"type": "object", "properties": map[string]any{"account_keyword": map[string]any{"type": "string"}, "fakeid": map[string]any{"type": "string"}, "article_keyword": map[string]any{"type": "string"}, "download_first": map[string]any{"type": "boolean"}}}}},
		Workflows:      []agentxworkflow.Spec{{ID: DefaultWorkflow, Title: "WeChat Article Read-only Acquisition", Description: "通过显式 Host Client 获取公众号文章 metadata 和 display-safe digest evidence。", Version: "v1", Pack: PackID, CaseTypes: []string{CaseTypeAcquire}, RouteHints: []string{"公众号文章搜索", "公众号文章同步"}, PlanningMode: agentxworkflow.PlanningBounded, EntryNode: "acquire_articles", DefaultContract: "wechat_article_readonly", StateSchema: []agentxworkflow.StateSlotSpec{{Name: "wechat.login_valid", Type: "boolean", Required: true}, {Name: "wechat.articles_observed", Type: "boolean", Required: true}, {Name: "wechat.passed", Type: "boolean", Required: true}, {Name: "wechat.failure_reasons", Type: "array"}, {Name: "wechat.summary", Type: "string"}}, EvaluatorSchema: []agentxworkflow.EvaluatorRef{{Name: "wechat_article_evidence_guard"}}, Nodes: []agentxworkflow.NodeSpec{{ID: "acquire_articles", Kind: agentxworkflow.NodeTool, Title: "Acquire WeChat articles", Description: "调用宿主注入的 exporter Client；登录、credential、cookie、网络和artifact由宿主负责。", Inputs: []agentxworkflow.BindingSpec{{From: "case.input.account_keyword", To: "args.account_keyword"}, {From: "case.input.fakeid", To: "args.fakeid"}, {From: "case.input.article_keyword", To: "args.article_keyword"}, {From: "case.input.download_first", To: "args.download_first"}}, Outputs: []agentxworkflow.BindingSpec{{From: "result.login_valid", To: "state.wechat.login_valid"}, {From: "result.articles_observed", To: "state.wechat.articles_observed"}, {From: "result.passed", To: "state.wechat.passed"}, {From: "result.failure_reasons", To: "state.wechat.failure_reasons"}, {From: "result.summary", To: "state.wechat.summary"}}, Config: map[string]any{"tool_name": AcquireTool, "args": map[string]any{"mode": "readonly"}}}}}},
		Tools:          []agentxpack.SemanticTool{{Name: AcquireTool, Description: "公众号文章只读采集入口。宿主负责 exporter、登录、credential、cookie、网络和artifact。", RuntimeTool: AcquireTool, Tags: []string{"wechat-article", "read-only", "host-adapter"}}},
		Evaluators:     []agentxpack.Evaluator{{Name: "wechat_article_evidence_guard", Description: "验证登录、文章列表、可选下载 digest 与 evidence。", OutputSchema: map[string]any{"type": "object", "properties": map[string]any{"passed": map[string]any{"type": "boolean"}, "login_valid": map[string]any{"type": "boolean"}, "articles_observed": map[string]any{"type": "boolean"}, "download_observed": map[string]any{"type": "boolean"}, "evidence_observed": map[string]any{"type": "boolean"}, "failure_reasons": map[string]any{"type": "array", "items": map[string]any{"type": "string"}}}, "required": []string{"passed", "login_valid", "articles_observed", "evidence_observed"}}}},
		EvalSuites:     []agentxpack.EvalSuite{{Name: "wechat_article_readonly_success", Description: "要求已登录、文章 evidence 可验证且只读。", Mode: agentxpack.EvalSuiteModeGate, WorkflowIDs: []string{DefaultWorkflow}, RequiredState: []string{"wechat.login_valid", "wechat.articles_observed", "wechat.passed"}, PassPath: "wechat.passed", SummaryPath: "wechat.summary", Default: true}},
		PolicyProfiles: []agentxpack.PolicyProfile{{Name: "wechat_article_readonly", Contract: agentxexecution.Contract{ID: "wechat-article-readonly", Strict: true, Version: 1, Visibility: agentxexecution.VisibilityPolicy{AllowTools: []string{AcquireTool}, DeclaredTools: []string{AcquireTool}, RequireDeclared: true, MaxRisk: "medium"}, Budget: agentxexecution.BudgetPolicy{MaxToolCalls: 1}, Loop: agentxexecution.LoopPolicy{MaxRounds: 1, LoopDetectionEnabled: true, ToolFailureFuseEnabled: true, ToolFailureFuseThreshold: 1}, SideEffects: agentxexecution.SideEffectPolicy{MaxClass: agentxexecution.SideEffectReadOnly, StrictRecovery: true}, Audit: agentxexecution.AuditPolicy{PersistSnapshot: true}}, Default: true}},
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
