package hostkit

import (
	"strconv"
	"strings"

	research "github.com/wsnacj/agentx-go/scenes/companyresearch"
)

func readyEvidenceContractLines(payload research.CompanyResearchPayload) []string {
	if len(payload.Subjects) > 0 {
		return compareReadyEvidenceContractLines(payload.Subjects)
	}
	lines := []string{}
	if readyDimension(payload, "financials") {
		if summary := financeEvidenceContractSummary(payload.Evidence.Finance); summary != "" {
			lines = append(lines, "财务："+summary)
		}
	}
	if readyDimension(payload, "market_data", "valuation") {
		if summary := marketEvidenceContractSummary(firstReadyEvidence(payload.Evidence.GlobalStock, payload.Evidence.AStock)); summary != "" {
			lines = append(lines, "行情/估值："+summary)
		}
	}
	if readyDimension(payload, "news", "risk") {
		if summary := newsEvidenceContractSummary(payload.Evidence.News); summary != "" {
			lines = append(lines, "新闻/风险："+summary)
		}
		lines = append(lines, newsEvidenceAssessmentContractLines(payload.Evidence.News)...)
	}
	if len(lines) == 0 {
		return nil
	}
	out := []string{"已通过证据摘要："}
	for _, line := range lines {
		out = append(out, "- "+line)
	}
	return out
}

func compareReadyEvidenceContractLines(subjects []research.CompanyResearchPayload) []string {
	lines := []string{}
	for _, subject := range subjects {
		parts := []string{}
		if readyDimension(subject, "financials") {
			if summary := financeEvidenceContractSummary(subject.Evidence.Finance); summary != "" {
				parts = append(parts, "财务："+summary)
			}
		}
		if readyDimension(subject, "market_data", "valuation") {
			if summary := marketEvidenceContractSummary(firstReadyEvidence(subject.Evidence.GlobalStock, subject.Evidence.AStock)); summary != "" {
				parts = append(parts, "行情/估值："+summary)
			}
		}
		if readyDimension(subject, "news", "risk") {
			if summary := newsEvidenceContractSummary(subject.Evidence.News); summary != "" {
				parts = append(parts, "新闻/风险："+summary)
			}
			parts = append(parts, newsEvidenceAssessmentContractLines(subject.Evidence.News)...)
		}
		if len(parts) == 0 {
			continue
		}
		name := firstNonEmpty(subject.Intent.EntityName, "unknown_subject")
		lines = append(lines, "- "+name+"："+strings.Join(parts, "；"))
	}
	if len(lines) == 0 {
		return nil
	}
	return append([]string{"已通过证据摘要："}, lines...)
}

func failureEvidenceContractLines(payload research.CompanyResearchPayload) []string {
	if len(payload.Subjects) > 0 {
		return compareFailureEvidenceContractLines(payload.Subjects)
	}
	lines := singleFailureEvidenceContractLines(payload)
	if len(lines) == 0 {
		return nil
	}
	out := []string{"未通过原因摘要："}
	for _, line := range lines {
		out = append(out, "- "+line)
	}
	return out
}

func compareFailureEvidenceContractLines(subjects []research.CompanyResearchPayload) []string {
	lines := []string{}
	for _, subject := range subjects {
		parts := singleFailureEvidenceContractLines(subject)
		if len(parts) == 0 {
			continue
		}
		name := firstNonEmpty(subject.Intent.EntityName, "unknown_subject")
		lines = append(lines, "- "+name+"："+strings.Join(parts, "；"))
	}
	if len(lines) == 0 {
		return nil
	}
	return append([]string{"未通过原因摘要："}, lines...)
}

func singleFailureEvidenceContractLines(payload research.CompanyResearchPayload) []string {
	lines := []string{}
	if !readyDimension(payload, "financials") {
		if summary := financeFailureContractSummary(payload.Evidence.Finance); summary != "" {
			lines = append(lines, "财务："+summary)
		}
	}
	if !readyDimension(payload, "market_data", "valuation") {
		if summary := marketFailureContractSummary(payload.Evidence.AStock, payload.Evidence.GlobalStock); summary != "" {
			lines = append(lines, "行情/估值："+summary)
		}
	}
	if !readyDimension(payload, "news", "risk") {
		if summary := newsFailureContractSummary(payload.Evidence.News); summary != "" {
			lines = append(lines, "新闻/风险："+summary)
		}
	}
	return lines
}

func readyDimension(payload research.CompanyResearchPayload, dims ...string) bool {
	ready := cleanStrings(payload.AnswerReadiness.ReadyDimensions)
	for _, dim := range dims {
		dim = strings.TrimSpace(dim)
		for _, value := range ready {
			if value == dim {
				return true
			}
		}
	}
	return false
}

func firstReadyEvidence(values ...map[string]any) map[string]any {
	for _, value := range values {
		if evidenceReady(value) {
			return value
		}
	}
	return nil
}

func financeEvidenceContractSummary(evidence map[string]any) string {
	if !financeEvidenceReady(evidence) {
		return ""
	}
	if summary := deepString(evidence, "brief", "evidence", "brief"); summary != "" {
		return truncateContractSummary(summary, 220)
	}
	parts := []string{}
	if company := firstNonEmpty(
		deepString(evidence, "candidates", "resolved_company"),
		deepString(evidence, "metrics", "evidence", "company_name"),
		deepString(evidence, "intent", "entity_name"),
	); company != "" {
		parts = append(parts, company)
	}
	if period := firstNonEmpty(
		deepString(evidence, "metrics", "evidence", "report_period"),
		deepString(evidence, "brief", "evidence", "report_period"),
	); period != "" {
		parts = append(parts, period)
	}
	metricEvidence := contractMapAt(evidence, "metrics", "evidence", "metric_evidence")
	for _, item := range []struct {
		key   string
		label string
	}{
		{key: "revenue", label: "收入"},
		{key: "revenue_growth", label: "收入增速"},
		{key: "net_profit", label: "净利润"},
		{key: "net_profit_growth", label: "净利润增速"},
		{key: "operating_cash_flow", label: "经营现金流"},
	} {
		value := firstNonEmpty(
			deepString(evidence, "metrics", "evidence", item.key),
			deepString(metricEvidence, item.key, "value"),
		)
		if value != "" {
			parts = append(parts, item.label+" "+value)
		}
	}
	if len(parts) > 0 {
		return truncateContractSummary(strings.Join(parts, "，"), 220)
	}
	for _, fact := range contractObjectListAt(evidence, "assessment_projection", "verified_facts") {
		field := strings.TrimSpace(toContractString(fact["field"]))
		value := strings.TrimSpace(toContractString(fact["value"]))
		if field != "" && value != "" {
			parts = append(parts, field+"="+value)
		}
		if len(parts) >= 4 {
			break
		}
	}
	if len(parts) > 0 {
		return truncateContractSummary(strings.Join(parts, "，"), 220)
	}
	return ""
}

func marketEvidenceContractSummary(evidence map[string]any) string {
	if !evidenceReady(evidence) {
		return ""
	}
	quoteEvidence := primaryMarketQuoteEvidence(evidence)
	parts := []string{}
	if subject := firstNonEmpty(
		deepString(evidence, "quote", "subject", "entity_name"),
		deepString(evidence, "quote", "subject", "display_name"),
		deepString(evidence, "subject", "entity_name"),
		deepString(evidence, "subject", "display_name"),
		deepString(quoteEvidence, "subject", "entity_name"),
		deepString(quoteEvidence, "subject", "display_name"),
	); subject != "" {
		code := firstNonEmpty(
			deepString(evidence, "quote", "subject", "stock_code"),
			deepString(evidence, "quote", "subject", "ticker"),
			deepString(evidence, "subject", "stock_code"),
			deepString(evidence, "subject", "ticker"),
			deepString(quoteEvidence, "subject", "stock_code"),
			deepString(quoteEvidence, "subject", "ticker"),
		)
		market := firstNonEmpty(
			deepString(evidence, "quote", "subject", "market"),
			deepString(evidence, "subject", "market"),
			deepString(quoteEvidence, "subject", "market"),
		)
		if displayCode := contractSecurityCode(code, market); displayCode != "" && !strings.Contains(strings.ToUpper(subject), displayCode) {
			subject += " (" + displayCode + ")"
		}
		parts = append(parts, subject)
	}
	for _, item := range []struct {
		key   string
		label string
	}{
		{key: "price", label: "价格"},
		{key: "change_percent", label: "涨跌幅"},
		{key: "pe_ttm", label: "PE(TTM)"},
		{key: "pb", label: "PB"},
		{key: "market_cap", label: "市值"},
	} {
		if value := marketMetricSummary(evidence, item.key); value != "" {
			parts = append(parts, item.label+" "+value)
		}
	}
	if asOf := firstNonEmpty(
		deepString(evidence, "quote", "freshness", "as_of"),
		deepString(evidence, "quote", "evidence", "as_of"),
		deepString(evidence, "freshness", "as_of"),
		deepString(quoteEvidence, "freshness", "as_of"),
		deepString(quoteEvidence, "evidence", "as_of"),
	); asOf != "" {
		parts = append(parts, "截至 "+asOf)
	}
	if len(parts) > 0 {
		return truncateContractSummary(strings.Join(parts, "，"), 220)
	}
	return truncateContractSummary(firstNonEmpty(
		deepString(evidence, "summary"),
		deepString(evidence, "brief"),
	), 220)
}

func contractSecurityCode(code string, market string) string {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return ""
	}
	for _, suffix := range []string{".HK", ".SH", ".SZ", ".SS", ".US", ".O", ".N"} {
		if strings.HasSuffix(code, suffix) {
			return code
		}
	}
	switch strings.ToLower(strings.TrimSpace(market)) {
	case "hk", "hkg", "hkex", "hongkong", "hong kong":
		return code + ".HK"
	case "sh", "sha", "sse", "shanghai":
		return code + ".SH"
	case "sz", "sza", "szse", "shenzhen":
		return code + ".SZ"
	default:
		return code
	}
}

func marketMetricSummary(evidence map[string]any, key string) string {
	quoteEvidence := primaryMarketQuoteEvidence(evidence)
	value := firstNonEmpty(
		deepString(evidence, "quote", "quote", key, "value"),
		deepString(evidence, "quote", key, "value"),
		deepString(evidence, key, "value"),
		deepString(quoteEvidence, "quote", key, "value"),
		deepString(quoteEvidence, key, "value"),
	)
	if value == "" {
		return ""
	}
	unit := firstNonEmpty(
		deepString(evidence, "quote", "quote", key, "unit"),
		deepString(evidence, "quote", key, "unit"),
		deepString(evidence, key, "unit"),
		deepString(quoteEvidence, "quote", key, "unit"),
		deepString(quoteEvidence, key, "unit"),
	)
	currency := firstNonEmpty(
		deepString(evidence, "quote", "quote", key, "currency"),
		deepString(evidence, "quote", key, "currency"),
		deepString(evidence, key, "currency"),
		deepString(quoteEvidence, "quote", key, "currency"),
		deepString(quoteEvidence, key, "currency"),
	)
	if unit == "100_million" {
		unit = "亿"
	}
	suffix := unit
	if currency != "" && !strings.EqualFold(currency, unit) {
		if suffix != "" {
			suffix += " "
		}
		suffix += currency
	}
	formatted := value
	if suffix != "" && !strings.Contains(value, suffix) {
		formatted += suffix
	}
	if key == "pe_ttm" && contractMetricIsNegative(value) {
		formatted += "（TTM 盈利为负，PE 不具常规可比意义）"
	}
	return formatted
}

func contractMetricIsNegative(value string) bool {
	number, err := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(value), ",", ""), 64)
	return err == nil && number < 0
}

func primaryMarketQuoteEvidence(evidence map[string]any) map[string]any {
	if len(evidence) == 0 {
		return nil
	}
	if quote := contractMapAt(evidence, "quote"); len(quote) > 0 {
		return quote
	}
	for _, quote := range contractObjectListAt(evidence, "quotes") {
		if evidenceReady(quote) {
			return quote
		}
	}
	quotes := contractObjectListAt(evidence, "quotes")
	if len(quotes) > 0 {
		return quotes[0]
	}
	return nil
}

func newsEvidenceContractSummary(evidence map[string]any) string {
	if !evidenceReady(evidence) {
		return ""
	}
	parts := []string{}
	if headline := firstNonEmpty(
		deepString(evidence, "headline"),
		deepString(evidence, "primary_source", "headline"),
		deepString(evidence, "sources", "primary_source", "headline"),
	); headline != "" {
		parts = append(parts, headline)
	}
	if update := firstNonEmpty(
		deepString(evidence, "key_update"),
		deepString(evidence, "primary_source", "key_update"),
		deepString(evidence, "sources", "primary_source", "key_update"),
		deepString(evidence, "summary"),
	); update != "" {
		parts = append(parts, update)
	}
	if publishedAt := firstNonEmpty(
		deepString(evidence, "published_at"),
		deepString(evidence, "primary_source", "published_at"),
		deepString(evidence, "sources", "primary_source", "published_at"),
	); publishedAt != "" {
		parts = append(parts, "发布时间 "+publishedAt)
	}
	if len(parts) > 0 {
		return truncateContractSummary(strings.Join(parts, "，"), 220)
	}
	return ""
}

func newsEvidenceAssessmentContractLines(evidence map[string]any) []string {
	if !evidenceReady(evidence) {
		return nil
	}
	lines := []string{}
	if impact := deepString(evidence, "answer_contract", "possible_impact"); impact != "" {
		lines = append(lines, "新闻可能影响："+truncateContractSummary(impact, 220))
	}
	if boundary := deepString(evidence, "answer_contract", "risk_boundary"); boundary != "" {
		lines = append(lines, "新闻风险边界："+truncateContractSummary(boundary, 220))
	}
	return lines
}

func financeFailureContractSummary(evidence map[string]any) string {
	if len(evidence) == 0 || evidenceReady(evidence) {
		return ""
	}
	reason := firstNonEmpty(
		deepString(evidence, "candidates", "failure_code"),
		deepString(evidence, "failure_code"),
		deepString(evidence, "answer_readiness", "failure_code"),
		deepString(evidence, "metrics", "failure_code"),
		deepString(evidence, "metrics", "evaluation", "failure_reason"),
	)
	if reason == "" {
		return ""
	}
	if isIdentityFailureReason(reason) {
		return "下游财报解析器未确认主体身份（" + reason + "），不能把输入名称直接等同为上市主体。"
	}
	return "财报证据未通过 guard（" + truncateContractSummary(reason, 160) + "）。"
}

func marketFailureContractSummary(aStock, globalStock map[string]any) string {
	reasons := []string{}
	if reason := marketFailureReason(aStock); reason != "" {
		reasons = append(reasons, "A股："+reason)
	}
	if reason := marketFailureReason(globalStock); reason != "" {
		reasons = append(reasons, "港美股："+reason)
	}
	if len(reasons) == 0 {
		return ""
	}
	allIdentity := true
	for _, reason := range reasons {
		if !isIdentityFailureReason(reason) {
			allIdentity = false
			break
		}
	}
	if allIdentity {
		if len(reasons) > 1 {
			return "A股和港美股下游解析器都未确认主体身份（" + strings.Join(reasons, "；") + "），不能输出股价或估值结论。"
		}
		return "下游行情解析器未确认主体身份（" + strings.Join(reasons, "；") + "），不能输出股价或估值结论。"
	}
	return "下游行情证据未通过（" + truncateContractSummary(strings.Join(reasons, "；"), 180) + "）。"
}

func marketFailureReason(evidence map[string]any) string {
	if len(evidence) == 0 || evidenceReady(evidence) {
		return ""
	}
	return firstNonEmpty(
		deepString(evidence, "failure_code"),
		deepString(evidence, "readiness", "failure_code"),
		deepString(evidence, "quote", "failure_code"),
		deepString(evidence, "quote", "readiness", "failure_code"),
	)
}

func newsFailureContractSummary(evidence map[string]any) string {
	if len(evidence) == 0 || evidenceReady(evidence) {
		return ""
	}
	reason := firstNonEmpty(
		deepString(evidence, "failure_code"),
		deepString(evidence, "sources", "failure_code"),
		deepString(evidence, "evaluator_report", "degrade_reason"),
	)
	if reason == "" {
		return ""
	}
	if strings.Contains(strings.ToLower(reason), "provider") {
		return "搜索 provider 不可用（" + reason + "），无法核验最新新闻来源。"
	}
	return "新闻证据未通过 guard（" + truncateContractSummary(reason, 160) + "）。"
}

func isIdentityFailureReason(reason string) bool {
	reason = strings.ToLower(strings.TrimSpace(reason))
	return strings.Contains(reason, "identity_not_found") ||
		strings.Contains(reason, "subject_mismatch") ||
		strings.Contains(reason, "identity not found")
}

func contractMapAt(object map[string]any, path ...string) map[string]any {
	var current any = object
	for _, key := range path {
		next, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		current = next[key]
	}
	next, _ := current.(map[string]any)
	return next
}

func contractObjectListAt(object map[string]any, path ...string) []map[string]any {
	var current any = object
	for _, key := range path {
		next, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		current = next[key]
	}
	switch typed := current.(type) {
	case []map[string]any:
		return typed
	case []any:
		out := make([]map[string]any, 0, len(typed))
		for _, item := range typed {
			if object, ok := item.(map[string]any); ok {
				out = append(out, object)
			}
		}
		return out
	default:
		return nil
	}
}

func truncateContractSummary(value string, limit int) string {
	value, _ = boundedContractSummary(value, limit)
	return value
}

const contractSummaryTruncationMarker = " [truncated]"

func boundedContractSummary(value string, limit int) (string, bool) {
	value = strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
	if limit <= 0 {
		return value, false
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value, false
	}
	return string(runes[:limit]) + contractSummaryTruncationMarker, true
}

func contractSummaryMarkedTruncated(value string) bool {
	return strings.Contains(value, contractSummaryTruncationMarker)
}

func toContractString(value any) string {
	text, _ := value.(string)
	return text
}
