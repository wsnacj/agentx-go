package hostkit

import (
	"context"
	"strings"
	"unicode"

	research "github.com/wsnacj/agentx-go/scenes/companyresearch"
)

type ToolPayloadHandler func(context.Context, map[string]any) (any, error)

type CompanyResearchHandlers struct {
	Finance     ToolPayloadHandler
	AStock      ToolPayloadHandler
	GlobalStock ToolPayloadHandler
	News        ToolPayloadHandler
	Guard       ToolPayloadHandler
}

type CompanyResearchConfig struct {
	Source              string
	SourcePolicyDefault string
	SubjectResolver     SubjectResolver
	TaskExecutor        TaskExecutor
	Handlers            CompanyResearchHandlers
}

func BuildCompanyResearchLookupHandler(cfg CompanyResearchConfig) research.ToolPayloadHandler {
	return func(ctx context.Context, params map[string]any) (any, error) {
		return BuildCompanyResearchLookupPayload(ctx, cfg, params)
	}
}

func BuildCompanyCompareLookupHandler(cfg CompanyResearchConfig) research.ToolPayloadHandler {
	return func(ctx context.Context, params map[string]any) (any, error) {
		return BuildCompanyComparePayload(ctx, cfg, params)
	}
}

func BuildCompanyResearchGuardHandler(cfg CompanyResearchConfig) research.ToolPayloadHandler {
	return func(ctx context.Context, params map[string]any) (any, error) {
		return BuildCompanyResearchGuardPayload(ctx, cfg, params)
	}
}

func BuildCompanyResearchLookupPayload(ctx context.Context, cfg CompanyResearchConfig, params map[string]any) (research.CompanyResearchPayload, error) {
	intent := research.IntentFromParams(params)
	if strings.TrimSpace(intent.SourcePolicy) == "" {
		intent.SourcePolicy = strings.TrimSpace(cfg.SourcePolicyDefault)
	}
	payload := research.CompanyResearchPayload{
		Tool:          research.ToolCompanyResearchLookup,
		Source:        firstNonEmpty(cfg.Source, "agentx_company_research_hostkit"),
		PackID:        research.PackID,
		CaseType:      "company_research.single_company",
		WorkflowID:    research.WorkflowID,
		AdapterID:     "company_research_hostkit",
		AdapterStatus: "ok",
		Intent:        intent,
	}
	subjectResolution, subjectWarnings := resolveSubject(ctx, cfg.SubjectResolver, intent)
	if subjectResolution != nil {
		payload.SubjectResolution = subjectResolution
	}
	payload.Warnings = append(payload.Warnings, subjectWarnings...)
	downstreamIntent := intentWithSubjectResolution(intent, subjectResolution)
	identityCheckIntent := subjectIdentityCheckIntent(intent, downstreamIntent, subjectResolution)
	taskPlan := research.BuildCompanyResearchTaskPlan(research.CompanyResearchTaskPlanInput{
		CaseType:          payload.CaseType,
		WorkflowID:        payload.WorkflowID,
		Intent:            intent,
		SubjectResolution: subjectResolution,
	})
	payload.TaskPlan = &taskPlan
	if result, ok := research.SubjectResolutionTaskResult(taskPlan); ok {
		appendTaskResult(&payload, result)
	}
	if needsDimension(intent, "financials") {
		params := downstreamParamsWithSubjectResolution(intent, "finance", subjectResolution)
		execution, ran := runTaskExecutor(ctx, cfg, payload, taskPlan, research.CompanyResearchRoleFinanceAnalyst, params)
		appendTaskExecutorWarnings(&payload, execution, ran)
		if ran && execution.Handled {
			payload.Evidence.Finance = normalizePayload(execution.Evidence)
			if task, ok := taskPlan.TaskByRole(research.CompanyResearchRoleFinanceAnalyst); ok {
				applyHandledTaskResult(&payload, task, execution)
			}
		} else {
			result, err := callHandler(ctx, cfg.Handlers.Finance, params)
			if err != nil {
				return research.CompanyResearchPayload{}, err
			}
			payload.Evidence.Finance = result
			if taskResult, ok := research.TaskResultFromEvidence(taskPlan, research.CompanyResearchRoleFinanceAnalyst, payload.Evidence.Finance); ok {
				applyFinanceEvidenceReadinessToTaskResult(&taskResult, payload.Evidence.Finance)
				mergeTaskExecutorDiagnostics(&taskResult, execution, ran)
				appendTaskResult(&payload, taskResult)
			}
		}
	}
	financeForHandoff := payload.Evidence.Finance
	if !financeSubjectMatchesIntent(identityCheckIntent, payload.Evidence.Finance) {
		payload.Warnings = append(payload.Warnings, "finance_subject_mismatch")
		financeForHandoff = nil
	}
	marketIntent := marketIntentWithFinanceIdentity(downstreamIntent, financeForHandoff)
	if needsMarketEvidence(intent) {
		params := downstreamParamsWithSubjectAndFinanceIdentity(marketIntent, "market", subjectResolution, financeForHandoff)
		execution, ran := runTaskExecutor(ctx, cfg, payload, taskPlan, research.CompanyResearchRoleMarketAnalyst, params)
		appendTaskExecutorWarnings(&payload, execution, ran)
		if ran && execution.Handled {
			assignMarketTaskEvidence(&payload, marketIntent, execution.Evidence)
			if task, ok := taskPlan.TaskByRole(research.CompanyResearchRoleMarketAnalyst); ok {
				applyHandledTaskResult(&payload, task, execution)
			}
		} else {
			if shouldCallAStock(marketIntent) {
				result, err := callHandler(ctx, cfg.Handlers.AStock, downstreamParamsWithSubjectAndFinanceIdentity(marketIntent, "a_stock", subjectResolution, financeForHandoff))
				if err != nil {
					return research.CompanyResearchPayload{}, err
				}
				payload.Evidence.AStock = result
			}
			if shouldCallGlobalStock(marketIntent) {
				result, err := callHandler(ctx, cfg.Handlers.GlobalStock, downstreamParamsWithSubjectAndFinanceIdentity(marketIntent, "global_stock", subjectResolution, financeForHandoff))
				if err != nil {
					return research.CompanyResearchPayload{}, err
				}
				payload.Evidence.GlobalStock = result
			}
			if taskResult, ok := research.TaskResultFromEvidence(taskPlan, research.CompanyResearchRoleMarketAnalyst, payload.Evidence.AStock, payload.Evidence.GlobalStock); ok {
				mergeTaskExecutorDiagnostics(&taskResult, execution, ran)
				appendTaskResult(&payload, taskResult)
			}
		}
	}
	if needsDimension(intent, "news", "risk") {
		params := downstreamNewsParamsWithSubjectAndFinanceIdentity(downstreamIntent, subjectResolution, financeForHandoff)
		execution, ran := runTaskExecutor(ctx, cfg, payload, taskPlan, research.CompanyResearchRoleNewsAnalyst, params)
		appendTaskExecutorWarnings(&payload, execution, ran)
		if ran && execution.Handled {
			payload.Evidence.News = normalizePayload(execution.Evidence)
			if task, ok := taskPlan.TaskByRole(research.CompanyResearchRoleNewsAnalyst); ok {
				applyHandledTaskResult(&payload, task, execution)
				if code := newsEvidenceDegradeCode(payload.Evidence.News); code != "" && code != "news_evidence_not_ready" {
					if taskResult, ok := research.TaskResultFromEvidence(research.CompanyResearchTaskPlan{Tasks: []research.CompanyResearchTaskSpec{task}}, task.Role, payload.Evidence.News); ok {
						applyNewsEvidenceQualityToTaskResult(&taskResult, payload.Evidence.News)
						mergeTaskExecutorDiagnostics(&taskResult, execution, ran)
						upsertTaskResult(&payload, taskResult)
					}
				}
			}
		} else {
			result, err := callHandler(ctx, cfg.Handlers.News, params)
			if err != nil {
				return research.CompanyResearchPayload{}, err
			}
			payload.Evidence.News = result
			if taskResult, ok := research.TaskResultFromEvidence(taskPlan, research.CompanyResearchRoleNewsAnalyst, payload.Evidence.News); ok {
				applyNewsEvidenceQualityToTaskResult(&taskResult, payload.Evidence.News)
				mergeTaskExecutorDiagnostics(&taskResult, execution, ran)
				appendTaskResult(&payload, taskResult)
			}
		}
	}
	applyDownstreamSubjectConfirmation(&payload, &taskPlan)
	readinessPayload := payload
	readinessPayload.Intent = identityCheckIntent
	payload.AnswerReadiness = AnswerReadiness(readinessPayload)
	payload.GuardStatus = guardStatus(payload.AnswerReadiness)
	if taskResult, ok := research.TaskResultFromReadiness(taskPlan, research.CompanyResearchRoleRiskReviewer, payload.AnswerReadiness); ok {
		appendTaskResult(&payload, taskResult)
	}
	if taskResult, ok := research.TaskResultFromReadiness(taskPlan, research.CompanyResearchRoleEvidenceGuard, payload.AnswerReadiness); ok {
		appendTaskResult(&payload, taskResult)
	}
	if cfg.Handlers.Guard != nil {
		guard, err := callHandler(ctx, cfg.Handlers.Guard, guardParams(payload))
		if err != nil {
			return research.CompanyResearchPayload{}, err
		}
		payload.Evidence.Guard = guard
	}
	if payload.AnswerReadiness.Degraded {
		payload.AdapterStatus = "degraded"
		payload.FailureCode = payload.AnswerReadiness.FailureCode
		payload.FailureClass = payload.AnswerReadiness.FailureClass
	}
	payload.TaskSummary = buildCompanyResearchTaskSummary(payload)
	payload.AnswerContract = CompanyResearchAnswerContract(payload)
	if taskResult, ok := research.TaskResultFromAnswerContract(taskPlan, payload.AnswerContract); ok {
		appendTaskResult(&payload, taskResult)
	}
	payload.TaskSummary = buildCompanyResearchTaskSummary(payload)
	return payload, nil
}

func BuildCompanyComparePayload(ctx context.Context, cfg CompanyResearchConfig, params map[string]any) (research.CompanyResearchPayload, error) {
	intent := research.IntentFromParams(params)
	payload := research.CompanyResearchPayload{
		Tool:          research.ToolCompanyCompareLookup,
		Source:        firstNonEmpty(cfg.Source, "agentx_company_research_hostkit"),
		PackID:        research.PackID,
		CaseType:      "company_research.multi_company_compare",
		WorkflowID:    research.WorkflowID,
		AdapterID:     "company_research_hostkit_compare",
		AdapterStatus: "ok",
		Intent:        intent,
	}
	taskPlan := research.BuildCompanyResearchTaskPlan(research.CompanyResearchTaskPlanInput{
		CaseType:   payload.CaseType,
		WorkflowID: payload.WorkflowID,
		Intent:     intent,
	})
	payload.TaskPlan = &taskPlan
	for _, subject := range intent.ComparisonSubjects {
		child, err := buildBestComparisonSubjectPayload(ctx, cfg, comparisonSubjectIntent(intent, subject))
		if err != nil {
			return research.CompanyResearchPayload{}, err
		}
		payload.Subjects = append(payload.Subjects, child)
	}
	payload.AnswerReadiness = CompareReadiness(payload)
	payload.GuardStatus = guardStatus(payload.AnswerReadiness)
	if payload.AnswerReadiness.Degraded {
		payload.AdapterStatus = "degraded"
		payload.FailureCode = payload.AnswerReadiness.FailureCode
		payload.FailureClass = payload.AnswerReadiness.FailureClass
	}
	if taskResult, ok := research.TaskResultFromReadiness(taskPlan, research.CompanyResearchRoleEvidenceGuard, payload.AnswerReadiness); ok {
		appendTaskResult(&payload, taskResult)
	}
	payload.TaskSummary = buildCompanyResearchTaskSummary(payload)
	payload.AnswerContract = CompanyResearchAnswerContract(payload)
	if taskResult, ok := research.TaskResultFromAnswerContract(taskPlan, payload.AnswerContract); ok {
		appendTaskResult(&payload, taskResult)
	}
	payload.TaskSummary = buildCompanyResearchTaskSummary(payload)
	return payload, nil
}

func comparisonSubjectIntent(parent research.CompanyResearchIntent, subject research.CompanySubject) research.CompanyResearchIntent {
	childIntent := parent
	childIntent.EntityName = subject.EntityName
	childIntent.EntityMentions = subject.EntityMentions
	childIntent.MarketHint = subject.MarketHint
	childIntent.ComparisonSubjects = nil
	return childIntent
}

func buildBestComparisonSubjectPayload(ctx context.Context, cfg CompanyResearchConfig, intent research.CompanyResearchIntent) (research.CompanyResearchPayload, error) {
	best, err := BuildCompanyResearchLookupPayload(ctx, cfg, research.ParamsFromIntent(intent))
	if err != nil {
		return research.CompanyResearchPayload{}, err
	}
	if best.AnswerReadiness.AnswerReady {
		return best, nil
	}
	bestScore := comparisonSubjectPayloadScore(best)
	for _, marketHint := range comparisonFallbackMarketHints(intent.MarketHint) {
		fallbackIntent := intent
		fallbackIntent.MarketHint = marketHint
		candidate, err := BuildCompanyResearchLookupPayload(ctx, cfg, research.ParamsFromIntent(fallbackIntent))
		if err != nil {
			return research.CompanyResearchPayload{}, err
		}
		score := comparisonSubjectPayloadScore(candidate)
		if score > bestScore {
			best = candidate
			bestScore = score
		}
		if candidate.AnswerReadiness.AnswerReady {
			return best, nil
		}
	}
	return best, nil
}

func comparisonFallbackMarketHints(marketHint string) []string {
	categories := marketHintCategories(marketHint)
	if len(categories) <= 1 {
		return nil
	}
	order := []string{"hk", "us", "A-share"}
	out := []string{}
	for _, category := range order {
		if categories[strings.ToLower(category)] || categories[category] {
			out = append(out, category)
		}
	}
	return cleanStrings(out)
}

func comparisonSubjectPayloadScore(payload research.CompanyResearchPayload) int {
	score := len(cleanStrings(payload.AnswerReadiness.ReadyDimensions)) * 20
	score -= len(cleanStrings(payload.AnswerReadiness.MissingDimensions)) * 30
	if payload.AnswerReadiness.AnswerReady {
		score += 1000
	}
	switch payload.AnswerReadiness.FailureClass {
	case "entity_ambiguous":
		score -= 20
	case "config_invalid", "auth_missing":
		score -= 10
	}
	if payload.TaskSummary != nil {
		score += len(payload.TaskSummary.ReadyRoles) * 5
		score -= len(payload.TaskSummary.DegradedRoles) * 5
		score -= len(payload.TaskSummary.FailedRoles) * 10
		score -= len(payload.TaskSummary.Conflicts) * 15
	}
	return score
}

func BuildCompanyResearchGuardPayload(ctx context.Context, cfg CompanyResearchConfig, params map[string]any) (research.CompanyResearchPayload, error) {
	_ = ctx
	_ = cfg
	intent := research.IntentFromParams(params)
	payload := research.CompanyResearchPayload{
		Tool:          research.ToolCompanyResearchGuard,
		Source:        firstNonEmpty(cfg.Source, "agentx_company_research_hostkit"),
		PackID:        research.PackID,
		CaseType:      "company_research.guard",
		WorkflowID:    research.WorkflowID,
		AdapterID:     "company_research_guard",
		AdapterStatus: "ok",
		Intent:        intent,
		GuardStatus:   "needs_review",
	}
	payload.AnswerReadiness = research.CompanyResearchAnswerReadiness{
		AnswerReady:   false,
		SafeToAnswer:  false,
		Degraded:      true,
		DegradeReason: "standalone_guard_requires_lookup_payload",
		AllowedScope:  "missing_lookup_evidence",
		FailureCode:   "standalone_guard_requires_lookup_payload",
	}
	return payload, nil
}

func AnswerReadiness(payload research.CompanyResearchPayload) research.CompanyResearchAnswerReadiness {
	intent := payload.Intent
	missing := []string{}
	ready := []string{}
	financeSubjectMatch := financeSubjectMatchesIntent(intent, payload.Evidence.Finance)
	if needsDimension(intent, "financials") {
		if !financeSubjectMatch || !financeEvidenceReady(payload.Evidence.Finance) {
			missing = append(missing, "financials")
		} else {
			ready = append(ready, "financials")
		}
	}
	if needsMarketEvidence(intent) {
		if !evidenceReady(payload.Evidence.AStock) && !evidenceReady(payload.Evidence.GlobalStock) {
			missing = append(missing, "market_data")
		} else {
			ready = append(ready, "market_data")
		}
	}
	if needsDimension(intent, "news", "risk") {
		if !newsEvidenceReady(payload.Evidence.News) {
			missing = append(missing, "news")
		} else {
			ready = append(ready, "news")
		}
	}
	if len(missing) > 0 {
		return research.CompanyResearchAnswerReadiness{
			AnswerReady:       false,
			SafeToAnswer:      true,
			Degraded:          true,
			DegradeReason:     "missing_required_dimensions",
			AllowedScope:      "partial_company_research",
			MissingDimensions: missing,
			ReadyDimensions:   ready,
			FailureCode:       "company_research_missing_required_dimensions",
			FailureClass:      companyResearchFailureClassForMissingDimensions(payload, missing),
		}
	}
	if len(taskConflicts(payload)) > 0 {
		return research.CompanyResearchAnswerReadiness{
			AnswerReady:   false,
			SafeToAnswer:  true,
			Degraded:      true,
			DegradeReason: "task_conflicts_detected",
			AllowedScope:  "conflict_review_required",
			FailureCode:   "company_research_task_conflicts",
			FailureClass:  "entity_ambiguous",
		}
	}
	return research.CompanyResearchAnswerReadiness{
		AnswerReady:     true,
		SafeToAnswer:    true,
		AllowedScope:    "requested_scope",
		ReadyDimensions: ready,
	}
}

func CompareReadiness(payload research.CompanyResearchPayload) research.CompanyResearchAnswerReadiness {
	if len(payload.Subjects) == 0 {
		return research.CompanyResearchAnswerReadiness{
			AnswerReady:   false,
			SafeToAnswer:  true,
			Degraded:      true,
			DegradeReason: "missing_comparison_subjects",
			AllowedScope:  "missing_subjects",
			FailureCode:   "company_compare_missing_subjects",
		}
	}
	commonReady := commonCompanyResearchDimensions(payload.Subjects)
	missing := []string{}
	failureClasses := []string{}
	for _, subject := range payload.Subjects {
		if !subject.AnswerReadiness.AnswerReady {
			missing = append(missing, firstNonEmpty(subject.Intent.EntityName, "unknown_subject"))
			failureClasses = append(failureClasses, subject.AnswerReadiness.FailureClass, subject.FailureClass)
		}
	}
	if len(missing) > 0 {
		return research.CompanyResearchAnswerReadiness{
			AnswerReady:       false,
			SafeToAnswer:      true,
			Degraded:          true,
			DegradeReason:     "some_subjects_missing_required_dimensions",
			AllowedScope:      "partial_company_comparison",
			MissingDimensions: missing,
			ReadyDimensions:   commonReady,
			FailureCode:       "company_compare_partial",
			FailureClass:      strings.Join(cleanStrings(failureClasses), ","),
		}
	}
	return research.CompanyResearchAnswerReadiness{
		AnswerReady:     true,
		SafeToAnswer:    true,
		AllowedScope:    "requested_scope",
		ReadyDimensions: commonReady,
	}
}

func commonCompanyResearchDimensions(subjects []research.CompanyResearchPayload) []string {
	if len(subjects) == 0 {
		return nil
	}
	common := cleanStrings(subjects[0].AnswerReadiness.ReadyDimensions)
	for _, subject := range subjects[1:] {
		ready := map[string]bool{}
		for _, dimension := range cleanStrings(subject.AnswerReadiness.ReadyDimensions) {
			ready[dimension] = true
		}
		filtered := common[:0]
		for _, dimension := range common {
			if ready[dimension] {
				filtered = append(filtered, dimension)
			}
		}
		common = filtered
	}
	return common
}

func callHandler(ctx context.Context, handler ToolPayloadHandler, params map[string]any) (map[string]any, error) {
	if handler == nil {
		return nil, nil
	}
	payload, err := handler(ctx, params)
	if err != nil {
		return nil, err
	}
	return normalizePayload(payload), nil
}

func appendTaskResult(payload *research.CompanyResearchPayload, result research.CompanyResearchTaskResult) {
	if payload == nil {
		return
	}
	payload.TaskResults = append(payload.TaskResults, result)
}

func upsertTaskResult(payload *research.CompanyResearchPayload, result research.CompanyResearchTaskResult) {
	if payload == nil {
		return
	}
	for i := range payload.TaskResults {
		if payload.TaskResults[i].Role == result.Role && payload.TaskResults[i].TaskID == result.TaskID {
			payload.TaskResults[i] = result
			return
		}
	}
	appendTaskResult(payload, result)
}

func appendTaskExecutorWarnings(payload *research.CompanyResearchPayload, execution CompanyResearchTaskExecutionResult, ran bool) {
	if payload == nil || !ran || len(execution.Warnings) == 0 {
		return
	}
	payload.Warnings = append(payload.Warnings, execution.Warnings...)
}

func mergeTaskExecutorDiagnostics(result *research.CompanyResearchTaskResult, execution CompanyResearchTaskExecutionResult, ran bool) {
	if result == nil || !ran || execution.TaskResult == nil {
		return
	}
	if len(execution.TaskResult.Warnings) > 0 {
		result.Warnings = append(result.Warnings, execution.TaskResult.Warnings...)
	}
	if strings.TrimSpace(execution.TaskResult.FailureCode) == "" && strings.TrimSpace(execution.TaskResult.AdapterStatus) == "" {
		return
	}
	if result.Diagnostics == nil {
		result.Diagnostics = map[string]string{}
	}
	if execution.TaskResult.FailureCode != "" {
		result.Diagnostics["task_executor_failure_code"] = execution.TaskResult.FailureCode
	}
	if execution.TaskResult.FailureClass != "" {
		result.FailureClass = firstNonEmpty(result.FailureClass, execution.TaskResult.FailureClass)
		result.Diagnostics["task_executor_failure_class"] = execution.TaskResult.FailureClass
	}
	if execution.TaskResult.AdapterStatus != "" {
		result.Diagnostics["task_executor_adapter_status"] = execution.TaskResult.AdapterStatus
	}
}

func assignMarketTaskEvidence(payload *research.CompanyResearchPayload, intent research.CompanyResearchIntent, evidence map[string]any) {
	if payload == nil {
		return
	}
	normalized := normalizePayload(evidence)
	switch strings.TrimSpace(research.StringArg(normalized["tool"])) {
	case "a_stock_investigation":
		payload.Evidence.AStock = normalized
	case "global_stock_investigation":
		payload.Evidence.GlobalStock = normalized
	default:
		if shouldCallAStock(intent) && !shouldCallGlobalStock(intent) {
			payload.Evidence.AStock = normalized
		} else {
			payload.Evidence.GlobalStock = normalized
		}
	}
}

func downstreamParams(intent research.CompanyResearchIntent, target string) map[string]any {
	out := research.ParamsFromIntent(intent)
	out["company_research_target"] = target
	return out
}

func guardParams(payload research.CompanyResearchPayload) map[string]any {
	return map[string]any{
		"user_message":       payload.Intent.UserMessage,
		"intent":             research.ParamsFromIntent(payload.Intent),
		"subject_resolution": payload.SubjectResolution,
		"evidence":           normalizePayload(payload.Evidence),
		"source_policy":      payload.Intent.SourcePolicy,
		"warnings":           payload.Warnings,
	}
}

func needsDimension(intent research.CompanyResearchIntent, dimensions ...string) bool {
	return research.ContainsToken(intent.RequestedDimensions, dimensions...)
}

func needsMarketEvidence(intent research.CompanyResearchIntent) bool {
	return needsDimension(intent, "market_data", "valuation", "announcements", "research")
}

func shouldCallAStock(intent research.CompanyResearchIntent) bool {
	market := strings.TrimSpace(intent.MarketHint)
	if market == "" {
		return true
	}
	return marketHintCategories(market)["a-share"]
}

func shouldCallGlobalStock(intent research.CompanyResearchIntent) bool {
	market := strings.TrimSpace(intent.MarketHint)
	if market == "" {
		return true
	}
	categories := marketHintCategories(market)
	return categories["hk"] || categories["us"] || categories["global"]
}

func marketHintHasSpecificExchange(market string) bool {
	categories := marketHintCategories(market)
	return categories["a-share"] || categories["hk"] || categories["us"]
}

func marketHintCategories(market string) map[string]bool {
	categories := map[string]bool{}
	add := func(value string) {
		switch normalizeFinanceMarketHint(value) {
		case "A-share":
			categories["a-share"] = true
		case "hk":
			categories["hk"] = true
		case "us":
			categories["us"] = true
		}
	}
	add(market)
	for token := range marketHintTokens(market) {
		add(token)
	}
	compact := strings.NewReplacer(
		" ", "",
		"-", "",
		"_", "",
		"/", "",
		"\\", "",
		",", "",
		"，", "",
		"、", "",
	).Replace(strings.ToLower(strings.TrimSpace(market)))
	switch compact {
	case "港美股", "港股美股", "香港美国", "香港美股", "hkus", "hkusa", "hongkongus", "hongkongusa":
		categories["hk"] = true
		categories["us"] = true
	case "沪深", "沪深a股", "沪市深市", "shsz", "shszse":
		categories["a-share"] = true
	case "global", "globalstock", "globalstocks", "international", "overseas", "intl", "境外", "海外":
		categories["global"] = true
	}
	if strings.Contains(compact, "港股") || strings.Contains(compact, "香港") || strings.Contains(compact, "hshare") {
		categories["hk"] = true
	}
	if strings.Contains(compact, "美股") || strings.Contains(compact, "美国") || strings.Contains(compact, "nasdaq") || strings.Contains(compact, "nyse") {
		categories["us"] = true
	}
	if strings.Contains(compact, "a股") || strings.Contains(compact, "ashare") || strings.Contains(compact, "沪深") {
		categories["a-share"] = true
	}
	return categories
}

func marketHintTokens(market string) map[string]bool {
	out := map[string]bool{}
	for _, token := range strings.FieldsFunc(strings.ToLower(strings.TrimSpace(market)), func(r rune) bool {
		switch r {
		case '/', '\\', ',', ';', '|', '+', '&', '，', '、', '；', '｜', '／', '＋', '＆':
			return true
		default:
			return unicode.IsSpace(r)
		}
	}) {
		token = strings.Trim(token, " ._-")
		if token == "" {
			continue
		}
		out[token] = true
		switch token {
		case "a":
			out["a-share"] = true
		case "h":
			out["h-share"] = true
		}
	}
	compact := strings.NewReplacer(" ", "", "-", "", "_", "").Replace(strings.ToLower(strings.TrimSpace(market)))
	if compact != "" {
		out[compact] = true
	}
	return out
}

func guardStatus(readiness research.CompanyResearchAnswerReadiness) string {
	if readiness.AnswerReady {
		return "passed"
	}
	return "needs_review"
}

func evidenceReady(evidence map[string]any) bool {
	return research.EvidencePayloadReady(evidence)
}

func financeEvidenceReady(evidence map[string]any) bool {
	if evidenceReady(evidence) {
		return true
	}
	return financeStructuredMetricsReady(evidence)
}

func financeStructuredMetricsReady(evidence map[string]any) bool {
	if len(evidence) == 0 {
		return false
	}
	status := strings.ToLower(strings.TrimSpace(deepString(evidence, "adapter_status")))
	if status != "" && status != "ok" && status != "needs_review" && status != "degraded" {
		return false
	}
	if financeTerminalFailureClass(evidence) {
		return false
	}
	company := firstNonEmpty(
		deepString(evidence, "metrics", "evidence", "company_name"),
		deepString(evidence, "candidates", "resolved_company"),
		deepString(evidence, "brief", "evidence", "company_name"),
	)
	period := firstNonEmpty(
		deepString(evidence, "metrics", "evidence", "report_period"),
		deepString(evidence, "brief", "evidence", "report_period"),
	)
	source := firstNonEmpty(
		deepString(evidence, "metrics", "evidence", "official_source"),
		deepString(evidence, "metrics", "evidence", "source_url"),
		deepString(evidence, "brief", "evidence", "source_url"),
	)
	if company == "" || period == "" || source == "" {
		return false
	}
	readyMetrics := 0
	metricEvidence := contractMapAt(evidence, "metrics", "evidence", "metric_evidence")
	for _, key := range []string{"revenue", "net_profit", "operating_cash_flow"} {
		value := firstNonEmpty(
			deepString(evidence, "metrics", "evidence", key),
			deepString(metricEvidence, key, "value"),
		)
		metricSource := firstNonEmpty(
			deepString(metricEvidence, key, "source"),
			source,
		)
		if value != "" && metricSource != "" {
			readyMetrics++
		}
	}
	return readyMetrics >= 2
}

func financeTerminalFailureClass(evidence map[string]any) bool {
	failureClass := strings.ToLower(strings.TrimSpace(firstNonEmpty(
		deepString(evidence, "failure_class"),
		deepString(evidence, "answer_readiness", "failure_class"),
		deepString(evidence, "evaluator_report", "failure_class"),
	)))
	switch failureClass {
	case "config_invalid", "auth_missing", "provider_unavailable", "unsupported":
		return true
	default:
		return false
	}
}

func newsEvidenceReady(evidence map[string]any) bool {
	return newsEvidenceDegradeCode(evidence) == ""
}

func newsEvidenceDegradeCode(evidence map[string]any) string {
	if !evidenceReady(evidence) {
		return "news_evidence_not_ready"
	}
	for _, path := range [][]string{{"guard_status"}, {"guard", "guard_status"}} {
		status := strings.ToLower(strings.TrimSpace(deepString(evidence, path...)))
		if status != "" && status != "passed" {
			return "news_guard_not_passed"
		}
	}
	for _, path := range [][]string{{"passed"}, {"guard", "passed"}} {
		if passed, ok := deepBool(evidence, path...); ok && !passed {
			return "news_guard_not_passed"
		}
	}
	reason := strings.ToLower(strings.TrimSpace(deepString(evidence, "answer_contract", "reason")))
	if reason == "source_quality_needs_review" {
		return "news_source_quality_needs_review"
	}
	if evidenceSourceRejected(evidence,
		[]string{"source_accepted"},
		[]string{"evaluator_report", "source_accepted"},
		[]string{"guard", "source_accepted"},
	) {
		return "news_source_not_accepted"
	}
	return ""
}

func companyResearchFailureClassForMissingDimensions(payload research.CompanyResearchPayload, missing []string) string {
	classes := []string{}
	for _, dim := range missing {
		switch strings.TrimSpace(dim) {
		case "financials":
			if stringSliceContains(payload.Warnings, "finance_subject_mismatch") {
				classes = append(classes, "entity_ambiguous")
				continue
			}
			classes = append(classes, evidenceFailureClass(payload.Evidence.Finance))
		case "market_data", "valuation":
			classes = append(classes, evidenceFailureClass(payload.Evidence.AStock))
			classes = append(classes, evidenceFailureClass(payload.Evidence.GlobalStock))
		case "news", "risk":
			classes = append(classes, evidenceFailureClass(payload.Evidence.News))
		}
	}
	classes = cleanStrings(classes)
	if len(classes) > 0 {
		return strings.Join(classes, ",")
	}
	if len(missing) > 0 {
		return "evidence_missing"
	}
	return ""
}

func evidenceFailureClass(evidence map[string]any) string {
	if len(evidence) == 0 {
		return ""
	}
	for _, path := range [][]string{
		{"failure_class"},
		{"sources", "failure_class"},
		{"answer_readiness", "failure_class"},
		{"evaluator_report", "failure_class"},
	} {
		if value := deepString(evidence, path...); value != "" {
			return normalizeCompanyFailureClass(value)
		}
	}
	return companyFailureClassFromCode(firstNonEmpty(
		deepString(evidence, "failure_code"),
		deepString(evidence, "sources", "failure_code"),
		deepString(evidence, "adapter_status"),
		deepString(evidence, "sources", "adapter_status"),
		deepString(evidence, "provider_status"),
		deepString(evidence, "sources", "provider_status"),
	))
}

func stringSliceContains(values []string, want string) bool {
	want = strings.TrimSpace(want)
	for _, value := range values {
		if strings.TrimSpace(value) == want {
			return true
		}
	}
	return false
}

func normalizeCompanyFailureClass(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	switch strings.ToLower(value) {
	case "config_invalid", "auth_missing", "quota_limited", "rate_limited", "transient_network", "temporary_provider_error", "provider_unavailable", "entity_ambiguous", "unsupported", "evidence_missing", "evidence_weak":
		return strings.ToLower(value)
	default:
		return value
	}
}

func companyFailureClassFromCode(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch {
	case normalized == "":
		return ""
	case strings.Contains(normalized, "subscription_token_invalid") ||
		strings.Contains(normalized, "credential_invalid") ||
		strings.Contains(normalized, "credential invalid") ||
		strings.Contains(normalized, "invalid_api_key") ||
		strings.Contains(normalized, "invalid api key") ||
		strings.Contains(normalized, "invalid_token") ||
		strings.Contains(normalized, "unauthorized") ||
		strings.Contains(normalized, "forbidden"):
		return "config_invalid"
	case strings.Contains(normalized, "missing_credentials") ||
		strings.Contains(normalized, "not_configured") ||
		strings.Contains(normalized, "missing credential"):
		return "auth_missing"
	case strings.Contains(normalized, "quota_limited") || strings.Contains(normalized, "quota exhausted"):
		return "quota_limited"
	case strings.Contains(normalized, "rate_limited") || strings.Contains(normalized, "status_429"):
		return "rate_limited"
	case strings.Contains(normalized, "timeout") ||
		strings.Contains(normalized, "status_502") ||
		strings.Contains(normalized, "status_503") ||
		strings.Contains(normalized, "status_504") ||
		strings.Contains(normalized, "connection reset"):
		return "transient_network"
	case strings.Contains(normalized, "identity_not_found") ||
		strings.Contains(normalized, "subject_mismatch") ||
		strings.Contains(normalized, "entity_mismatch"):
		return "entity_ambiguous"
	case strings.Contains(normalized, "search_provider_failure") ||
		strings.Contains(normalized, "provider_unavailable"):
		return "provider_unavailable"
	case strings.Contains(normalized, "unsupported"):
		return "unsupported"
	case strings.Contains(normalized, "missing") || strings.Contains(normalized, "incomplete"):
		return "evidence_missing"
	default:
		return ""
	}
}

func applyNewsEvidenceQualityToTaskResult(result *research.CompanyResearchTaskResult, evidence map[string]any) {
	if result == nil {
		return
	}
	code := newsEvidenceDegradeCode(evidence)
	if code == "" || code == "news_evidence_not_ready" {
		return
	}
	result.Status = research.CompanyResearchTaskStatusDegraded
	result.EvidenceReady = false
	result.AdapterStatus = "degraded"
	result.FailureCode = firstNonEmpty(result.FailureCode, code)
	result.FailureClass = firstNonEmpty(result.FailureClass, evidenceFailureClass(evidence), companyFailureClassFromCode(code), "evidence_weak")
	result.Summary = "news evidence did not pass latest-news source-quality guard"
	if result.Diagnostics == nil {
		result.Diagnostics = map[string]string{}
	}
	result.Diagnostics["news_quality_code"] = code
	if result.FailureClass != "" {
		result.Diagnostics["news_failure_class"] = result.FailureClass
	}
}

func applyFinanceEvidenceReadinessToTaskResult(result *research.CompanyResearchTaskResult, evidence map[string]any) {
	if result == nil || !financeEvidenceReady(evidence) || evidenceReady(evidence) {
		return
	}
	originalFailureCode := strings.TrimSpace(result.FailureCode)
	originalFailureClass := strings.TrimSpace(result.FailureClass)
	result.Status = research.CompanyResearchTaskStatusReady
	result.EvidenceReady = true
	result.AdapterStatus = firstNonEmpty(result.AdapterStatus, "ok")
	result.FailureCode = ""
	result.FailureClass = ""
	result.Summary = "structured financial metrics evidence is ready"
	if result.Diagnostics == nil {
		result.Diagnostics = map[string]string{}
	}
	if originalFailureCode != "" {
		result.Diagnostics["finance_brief_failure_code"] = originalFailureCode
	}
	if originalFailureClass != "" {
		result.Diagnostics["finance_brief_failure_class"] = originalFailureClass
	}
}

func normalizePayload(payload any) map[string]any {
	switch typed := payload.(type) {
	case nil:
		return nil
	case map[string]any:
		return typed
	case research.CompanyResearchEvidence:
		out := map[string]any{}
		if len(typed.Finance) > 0 {
			out["finance"] = typed.Finance
		}
		if len(typed.AStock) > 0 {
			out["a_stock"] = typed.AStock
		}
		if len(typed.GlobalStock) > 0 {
			out["global_stock"] = typed.GlobalStock
		}
		if len(typed.News) > 0 {
			out["news"] = typed.News
		}
		return out
	default:
		return map[string]any{"payload": typed}
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
