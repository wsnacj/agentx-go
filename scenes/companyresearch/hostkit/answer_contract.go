package hostkit

import (
	"strings"

	research "github.com/wsnacj/agentx-go/scenes/companyresearch"
)

const (
	defaultCompanyResearchMaxSubjects        = 12
	defaultCompanyResearchMaxRunesPerSubject = 480
)

func CompanyResearchAnswerContract(payload research.CompanyResearchPayload) *research.CompanyResearchAnswerContract {
	readiness := payload.AnswerReadiness
	if readiness.AnswerReady {
		return readyCompanyResearchAnswerContract(payload)
	}
	if !readiness.Degraded {
		return nil
	}
	missing := cleanStrings(readiness.MissingDimensions)
	ready := cleanStrings(readiness.ReadyDimensions)
	entity := companyResearchContractEntity(payload)
	missingLabel := "缺失或未通过 guard 的维度："
	if len(payload.Subjects) > 0 {
		missingLabel = "未满足 required dimensions 的主体："
	}
	missingText := strings.Join(missing, "、")
	if missingText == "" {
		missingText = "部分必要证据"
	}
	readyText := strings.Join(ready, "、")
	if readyText == "" {
		if len(payload.Subjects) > 0 {
			readyText = "没有所有主体都完整 ready 的维度"
		} else {
			readyText = "暂无完整维度"
		}
	}
	lines := []string{
		"当前不能把这次公司研究包装成完整结论。",
		"主体：" + entity + "。",
		"已通过或可有限参考的维度：" + readyText + "。",
	}
	if evidenceLines := readyEvidenceContractLines(payload); len(evidenceLines) > 0 {
		lines = append(lines, evidenceLines...)
	}
	if failureLines := failureEvidenceContractLines(payload); len(failureLines) > 0 {
		lines = append(lines, failureLines...)
	}
	if taskLines := taskSummaryContractLines(payload); len(taskLines) > 0 {
		lines = append(lines, taskLines...)
	}
	possibleImpact := "由于 " + missingText + " 未通过 guard，不能判断缺失证据是否会改变已通过维度的解读。"
	riskBoundary := "只能基于已通过维度做有限摘要；不能把未 ready 的财报、行情、新闻或风险证据写成已核实事实，也不能给出无边界投资建议。"
	lines = append(lines,
		missingLabel+missingText+"。",
		"可能影响："+possibleImpact,
		"风险边界："+riskBoundary,
	)
	if len(payload.Subjects) > 0 {
		lines = compareContractLines(lines, payload.Subjects)
	}
	recovery := companyResearchRecoveryRecommendation(payload)
	finalAnswerRecommended := !recovery.Recommended
	if recovery.Recommended {
		lines = append(lines[:len(lines)-1], recovery.lines()...)
		riskBoundary = "当前不能把缺失维度写成已核实事实；应先按恢复建议补证据，若仍失败再给出 partial 或 blocked closeout。"
		lines = append(lines, "风险边界："+riskBoundary)
	}
	contract := &research.CompanyResearchAnswerContract{
		FinalAnswerRecommended: finalAnswerRecommended,
		Reason:                 firstNonEmpty(readiness.FailureCode, readiness.DegradeReason, "company_research_degraded"),
		AllowedSummaryScope:    firstNonEmpty(readiness.AllowedScope, "partial_company_research"),
		DoNotRetryTools:        companyResearchDoNotRetryTools(recovery.Recommended),
		RecoveryRecommended:    recovery.Recommended,
		RecoveryReason:         recovery.Reason,
		SuggestedRecoveryTools: recovery.Tools,
		RecoveryTargets:        recovery.Targets,
		PossibleImpact:         possibleImpact,
		RiskBoundary:           riskBoundary,
		FinalAnswerDraft:       strings.Join(lines, "\n"),
	}
	return addCompanyResearchSubjectSummaries(contract, payload)
}

func readyCompanyResearchAnswerContract(payload research.CompanyResearchPayload) *research.CompanyResearchAnswerContract {
	entity := companyResearchContractEntity(payload)
	ready := cleanStrings(payload.AnswerReadiness.ReadyDimensions)
	readyText := strings.Join(ready, "、")
	if readyText == "" {
		readyText = "已请求维度"
	}
	lines := []string{
		"公司研究已完成。",
		"主体：" + entity + "。",
		"已通过证据维度：" + readyText + "。",
		"缺失或未通过 guard 的维度：无。",
	}
	if evidenceLines := readyEvidenceContractLines(payload); len(evidenceLines) > 0 {
		lines = append(lines, evidenceLines...)
	}
	if taskLines := taskSummaryContractLines(payload); len(taskLines) > 0 {
		lines = append(lines, taskLines...)
	}
	possibleImpact := "已通过证据可用于描述当前经营、市场与公开新闻状态，但不能单独证明未来影响或投资结果。"
	if valuationJudgment, ok := companyResearchValuationNewsJudgment(payload); ok {
		lines = append(lines, "估值影响判断："+valuationJudgment)
		possibleImpact = "当前新闻事件和点时估值均有通过 guard 的公开证据，但缺少经验证的盈利预测、现金流修订或估值倍数传导证据；后续出现这些证据时需要重新评估。"
	}
	riskBoundary := "以上仅基于已通过的公开证据摘要，不构成投资建议；如需正式材料，应回到原始来源复核关键数字、日期和风险表述。"
	lines = append(lines,
		"可能影响："+possibleImpact,
		"风险边界："+riskBoundary,
	)
	contract := &research.CompanyResearchAnswerContract{
		FinalAnswerRecommended: true,
		Reason:                 "company_research_ready",
		AllowedSummaryScope:    firstNonEmpty(payload.AnswerReadiness.AllowedScope, "requested_scope"),
		DoNotRetryTools: []string{
			research.ToolCompanyResearchLookup,
			research.ToolCompanyCompareLookup,
			research.ToolCompanyResearchGuard,
		},
		PossibleImpact:   possibleImpact,
		RiskBoundary:     riskBoundary,
		FinalAnswerDraft: strings.Join(lines, "\n"),
	}
	return addCompanyResearchSubjectSummaries(contract, payload)
}

func addCompanyResearchSubjectSummaries(contract *research.CompanyResearchAnswerContract, payload research.CompanyResearchPayload) *research.CompanyResearchAnswerContract {
	if contract == nil || len(payload.Subjects) == 0 {
		return contract
	}
	limit := len(payload.Subjects)
	if limit > defaultCompanyResearchMaxSubjects {
		limit = defaultCompanyResearchMaxSubjects
	}
	contract.SubjectBudget = &research.CompanyResearchSubjectBudget{
		MaxSubjects:        defaultCompanyResearchMaxSubjects,
		MaxRunesPerSubject: defaultCompanyResearchMaxRunesPerSubject,
		TotalSubjects:      len(payload.Subjects),
		ReturnedSubjects:   limit,
	}
	contract.Truncated = len(payload.Subjects) > limit || contractSummaryMarkedTruncated(contract.FinalAnswerDraft)
	for _, subject := range payload.Subjects[:limit] {
		summary := ""
		childTruncated := false
		if subject.AnswerContract != nil {
			summary = subject.AnswerContract.FinalAnswerDraft
			childTruncated = subject.AnswerContract.Truncated
		}
		if strings.TrimSpace(summary) == "" {
			summary = strings.Join(readyEvidenceContractLines(subject), "\n")
		}
		if strings.TrimSpace(summary) == "" {
			summary = firstNonEmpty(subject.AnswerReadiness.DegradeReason, subject.AnswerReadiness.FailureCode, "该主体暂无可摘要证据。")
		}
		summary, budgetTruncated := boundedContractSummary(summary, defaultCompanyResearchMaxRunesPerSubject)
		truncated := childTruncated || budgetTruncated || contractSummaryMarkedTruncated(summary)
		contract.SubjectSummaries = append(contract.SubjectSummaries, research.CompanyResearchSubjectSummary{
			EntityName:        firstNonEmpty(subject.Intent.EntityName, "unknown_subject"),
			AnswerReady:       subject.AnswerReadiness.AnswerReady,
			ReadyDimensions:   cleanStrings(subject.AnswerReadiness.ReadyDimensions),
			MissingDimensions: cleanStrings(subject.AnswerReadiness.MissingDimensions),
			Summary:           summary,
			Truncated:         truncated,
		})
		contract.Truncated = contract.Truncated || truncated
	}
	return contract
}

func companyResearchValuationNewsJudgment(payload research.CompanyResearchPayload) (string, bool) {
	if len(payload.Subjects) > 0 ||
		!companyResearchDimensionReady(payload.AnswerReadiness.ReadyDimensions, "market_data", "valuation") ||
		!companyResearchDimensionReady(payload.AnswerReadiness.ReadyDimensions, "news") {
		return "", false
	}
	request := strings.ToLower(strings.Join(cleanStrings([]string{
		payload.Intent.UserMessage,
		payload.Intent.OriginalIntent,
	}), " "))
	if !containsAnyMarker(request, "估值", "市盈率", "市净率", "valuation", "multiple", "p/e", "p/b") ||
		!containsAnyMarker(request, "新闻", "消息", "事件", "公告", "news", "announcement", "event") ||
		!companyResearchAsksNewsValuationRelationship(request) {
		return "", false
	}
	return "现有证据不足以认定该新闻已经改变前述估值判断。", true
}

func companyResearchAsksNewsValuationRelationship(request string) bool {
	return containsAnyMarker(request, "改变", "重估", "change", "affect", "impact valuation", "valuation impact", "re-rate", "rerate") ||
		(strings.Contains(request, "影响") && strings.Contains(request, "估值判断"))
}

func companyResearchDimensionReady(ready []string, dimensions ...string) bool {
	for _, value := range ready {
		for _, dimension := range dimensions {
			if strings.EqualFold(strings.TrimSpace(value), dimension) {
				return true
			}
		}
	}
	return false
}

func containsAnyMarker(value string, markers ...string) bool {
	for _, marker := range markers {
		if strings.Contains(value, marker) {
			return true
		}
	}
	return false
}

type companyResearchRecoveryRecommendationResult struct {
	Recommended bool
	Reason      string
	Tools       []string
	Targets     []research.CompanyResearchRecoveryTarget
}

func (r companyResearchRecoveryRecommendationResult) lines() []string {
	if !r.Recommended {
		return nil
	}
	lines := []string{
		"恢复建议：当前缺口属于可尝试补证据的状态，应优先补齐缺失维度后再生成最终结论。",
	}
	if len(r.Tools) > 0 {
		lines = append(lines, "建议调用能力："+strings.Join(r.Tools, "、")+"。")
	}
	for _, target := range r.Targets {
		entity := firstNonEmpty(target.EntityName, "未命名主体")
		dimension := firstNonEmpty(target.MissingDimension, "missing_evidence")
		tools := strings.Join(cleanStrings(target.SuggestedTools), "、")
		if tools == "" {
			tools = "按该维度对应的公开证据能力补证据"
		}
		lines = append(lines, "补证据目标："+entity+" / "+dimension+" / "+tools+"。")
	}
	return lines
}

func companyResearchRecoveryRecommendation(payload research.CompanyResearchPayload) companyResearchRecoveryRecommendationResult {
	if payload.AnswerReadiness.AnswerReady || !payload.AnswerReadiness.Degraded {
		return companyResearchRecoveryRecommendationResult{}
	}
	if len(payload.Subjects) == 0 {
		return companyResearchRecoveryRecommendationResult{}
	}
	if !companyResearchFailureClassRecoverable(payload.AnswerReadiness.FailureClass) {
		return companyResearchRecoveryRecommendationResult{}
	}
	result := companyResearchRecoveryRecommendationResult{
		Reason: "recoverable_missing_evidence",
	}
	for _, subject := range payload.Subjects {
		if subject.AnswerReadiness.AnswerReady {
			continue
		}
		if !companyResearchFailureClassRecoverable(firstNonEmpty(subject.AnswerReadiness.FailureClass, subject.FailureClass)) {
			continue
		}
		for _, dimension := range cleanStrings(subject.AnswerReadiness.MissingDimensions) {
			tools := suggestedRecoveryToolsForCompanyDimension(subject, dimension)
			tools = filterTerminalDownstreamRecoveryTools(subject, dimension, tools)
			if len(tools) == 0 {
				continue
			}
			result.Targets = append(result.Targets, research.CompanyResearchRecoveryTarget{
				EntityName:       firstNonEmpty(subject.Intent.EntityName, subjectIdentityDisplayName(subject), "unknown_subject"),
				MissingDimension: dimension,
				FailureClass:     firstNonEmpty(subject.AnswerReadiness.FailureClass, subject.FailureClass),
				FailureCode:      firstNonEmpty(subject.AnswerReadiness.FailureCode, subject.FailureCode),
				SuggestedTools:   tools,
			})
			result.Tools = append(result.Tools, tools...)
		}
	}
	result.Tools = cleanStrings(result.Tools)
	result.Recommended = len(result.Targets) > 0
	return result
}

func filterTerminalDownstreamRecoveryTools(subject research.CompanyResearchPayload, dimension string, tools []string) []string {
	evidence := companyResearchDimensionEvidence(subject, dimension)
	finalRecommended, ok := deepBool(evidence, "answer_contract", "final_answer_recommended")
	if !ok || !finalRecommended {
		return tools
	}
	contract := contractMapAt(evidence, "answer_contract")
	blocked := research.StringListArg(contract["do_not_retry_tools"])
	if len(blocked) == 0 {
		return tools
	}
	out := make([]string, 0, len(tools))
	for _, tool := range tools {
		if !containsStringFold(blocked, tool) {
			out = append(out, tool)
		}
	}
	return out
}

func companyResearchDimensionEvidence(subject research.CompanyResearchPayload, dimension string) map[string]any {
	switch strings.ToLower(strings.TrimSpace(dimension)) {
	case "financials":
		return subject.Evidence.Finance
	case "market_data", "valuation":
		if len(subject.Evidence.AStock) > 0 {
			return subject.Evidence.AStock
		}
		return subject.Evidence.GlobalStock
	case "news", "risk":
		return subject.Evidence.News
	default:
		return nil
	}
}

func containsStringFold(values []string, want string) bool {
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(want)) {
			return true
		}
	}
	return false
}

func companyResearchFailureClassRecoverable(value string) bool {
	recoverable := false
	for _, token := range strings.Split(strings.ToLower(strings.TrimSpace(value)), ",") {
		token = strings.TrimSpace(token)
		switch token {
		case "evidence_missing", "evidence_weak", "transient_network", "temporary_provider_error", "rate_limited", "":
			recoverable = true
		case "config_invalid", "auth_missing", "quota_limited", "unsupported", "entity_ambiguous", "provider_unavailable":
			return false
		}
	}
	return recoverable
}

func suggestedRecoveryToolsForCompanyDimension(subject research.CompanyResearchPayload, dimension string) []string {
	switch strings.ToLower(strings.TrimSpace(dimension)) {
	case "financials":
		return []string{"finance_report_lookup"}
	case "market_data", "valuation":
		if shouldCallAStock(subject.Intent) && !shouldCallGlobalStock(subject.Intent) {
			return []string{"a_stock_investigation"}
		}
		if shouldCallGlobalStock(subject.Intent) && !shouldCallAStock(subject.Intent) {
			return []string{"global_stock_investigation"}
		}
		return []string{"a_stock_investigation", "global_stock_investigation"}
	case "news", "risk":
		return []string{"latest_news_lookup"}
	default:
		return nil
	}
}

func companyResearchDoNotRetryTools(recoveryRecommended bool) []string {
	if recoveryRecommended {
		return []string{
			research.ToolCompanyCompareLookup,
			research.ToolCompanyResearchGuard,
		}
	}
	return []string{
		research.ToolCompanyResearchLookup,
		research.ToolCompanyCompareLookup,
		research.ToolCompanyResearchGuard,
	}
}

func subjectIdentityDisplayName(payload research.CompanyResearchPayload) string {
	if payload.SubjectResolution != nil && payload.SubjectResolution.SelectedCandidate != nil {
		return firstNonEmpty(payload.SubjectResolution.SelectedCandidate.DisplayName, payload.SubjectResolution.SelectedCandidate.EntityName)
	}
	return ""
}

func companyResearchContractEntity(payload research.CompanyResearchPayload) string {
	if len(payload.Subjects) > 0 {
		return "本次对比对象"
	}
	if hint, _, ok := downstreamSubjectConfirmation(payload.Intent, payload.Evidence); ok {
		return companyResearchContractEntityFromParts(
			firstNonEmpty(hint.CompanyName, payload.Intent.EntityName, "该公司"),
			firstNonEmpty(hint.Ticker, hint.StockCode),
			hint.MarketHint,
		)
	}
	if payload.SubjectResolution != nil && payload.SubjectResolution.SelectedCandidate != nil {
		candidate := payload.SubjectResolution.SelectedCandidate
		if !candidate.Verified {
			return firstNonEmpty(payload.Intent.EntityName, "该公司")
		}
		name := firstNonEmpty(candidate.EntityName, candidate.DisplayName)
		return companyResearchContractEntityFromParts(
			firstNonEmpty(name, payload.Intent.EntityName, "该公司"),
			firstNonEmpty(candidate.Ticker, candidate.StockCode),
			candidate.Market,
		)
	}
	return firstNonEmpty(payload.Intent.EntityName, "该公司")
}

func companyResearchContractEntityFromParts(name string, code string, market string) string {
	name = firstNonEmpty(name, "该公司")
	code = strings.TrimSpace(code)
	market = strings.TrimSpace(market)
	if code != "" && market != "" {
		displayCode := contractSecurityCode(code, market)
		if displayCode != "" && (!strings.EqualFold(displayCode, code) || strings.Contains(code, ".")) {
			return name + "（" + displayCode + "）"
		}
		return name + "（" + code + "/" + market + "）"
	}
	if code != "" {
		return name + "（" + code + "）"
	}
	return name
}

func compareContractLines(lines []string, subjects []research.CompanyResearchPayload) []string {
	readySubjects := []string{}
	partialReadySubjects := []string{}
	missingSubjects := []string{}
	for _, subject := range subjects {
		name := firstNonEmpty(subject.Intent.EntityName, "unknown_subject")
		if subject.AnswerReadiness.AnswerReady {
			readySubjects = append(readySubjects, name)
			continue
		}
		readyDims := cleanStrings(subject.AnswerReadiness.ReadyDimensions)
		if len(readyDims) > 0 {
			partialReadySubjects = append(partialReadySubjects, name+"("+strings.Join(readyDims, "、")+")")
		}
		dims := cleanStrings(subject.AnswerReadiness.MissingDimensions)
		if len(dims) > 0 {
			missingSubjects = append(missingSubjects, name+"("+strings.Join(dims, "、")+")")
		} else {
			missingSubjects = append(missingSubjects, name)
		}
	}
	extra := []string{}
	if len(readySubjects) > 0 {
		extra = append(extra, "已 ready 的主体："+strings.Join(readySubjects, "、")+"。")
	}
	if len(partialReadySubjects) > 0 {
		extra = append(extra, "部分 ready 的主体维度："+strings.Join(partialReadySubjects, "、")+"。")
	}
	if len(missingSubjects) > 0 {
		extra = append(extra, "未 ready 的主体及缺失维度："+strings.Join(missingSubjects, "、")+"。")
	}
	if len(extra) == 0 {
		return lines
	}
	out := append([]string{}, lines[:len(lines)-1]...)
	out = append(out, extra...)
	out = append(out, lines[len(lines)-1])
	return out
}

func cleanStrings(values []string) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}
