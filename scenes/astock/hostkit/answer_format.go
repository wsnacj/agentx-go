package hostkit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	astockcontracts "github.com/wsnacj/agentx-go/scenes/astock/contracts"
)

// BuildAStockAnswerFormatHandler returns a deterministic final-answer formatter
// for pack workflows. It formats only already-verified payload evidence.
func BuildAStockAnswerFormatHandler() ToolPayloadHandler {
	return func(_ context.Context, params map[string]any) (any, error) {
		return FormatAStockAnswer(params)
	}
}

func FormatAStockAnswer(params map[string]any) (string, error) {
	payload := mapArg(params["payload"])
	if len(payload) == 0 {
		return formatAStockAnswerFormatSkipped(params), nil
	}
	if reason := nonStandardPayloadReason(payload); reason != "" {
		params = cloneStringAnyMap(params)
		params["_argument_decode_error"] = reason
		return formatAStockAnswerFormatSkipped(params), nil
	}
	kind := strings.TrimSpace(StringArg(params["answer_kind"]))
	if kind == "" {
		kind = inferAnswerKind(payload)
	}
	switch kind {
	case "valuation":
		var value astockcontracts.QuotePayload
		if err := decodePayload(payload, &value); err != nil {
			return formatAStockAnswerPayloadDecodeSkipped(params, "invalid_valuation_payload"), nil
		}
		if reason := payloadReadinessContractReason(payload); reason != "" {
			return formatAStockAnswerPayloadDecodeSkipped(params, reason), nil
		}
		return formatQuoteAnswer(value), nil
	case "research":
		var value researchFormatPayload
		if err := decodePayload(payload, &value); err != nil {
			return formatAStockAnswerPayloadDecodeSkipped(params, "invalid_research_payload"), nil
		}
		if reason := payloadReadinessContractReason(payload); reason != "" {
			return formatAStockAnswerPayloadDecodeSkipped(params, reason), nil
		}
		return formatResearchAnswer(value), nil
	case "signal":
		var value astockcontracts.SignalPayload
		if err := decodePayload(payload, &value); err != nil {
			return formatAStockAnswerPayloadDecodeSkipped(params, "invalid_signal_payload"), nil
		}
		if reason := payloadReadinessContractReason(payload); reason != "" {
			return formatAStockAnswerPayloadDecodeSkipped(params, reason), nil
		}
		return formatSignalAnswer(value), nil
	default:
		return formatAStockAnswerPayloadDecodeSkipped(params, "unsupported_answer_kind"), nil
	}
}

func formatAStockAnswerPayloadDecodeSkipped(params map[string]any, reason string) string {
	params = cloneStringAnyMap(params)
	params["_argument_decode_error"] = reason
	return formatAStockAnswerFormatSkipped(params)
}

func formatAStockAnswerFormatSkipped(params map[string]any) string {
	reason := strings.TrimSpace(StringArg(params["_argument_decode_error"]))
	if reason == "" {
		reason = "payload is required"
	}
	return "a_stock_answer_format 未执行格式化：缺少可验证的标准 A 股工具 payload（" + reason + "）。请基于前面已经返回的 finance_report_lookup / a_stock_* 工具证据直接给出有边界的自然语言回答；不要把模型自行拼装的混合财报/行情 payload 当成新的事实来源。"
}

func nonStandardPayloadReason(payload map[string]any) string {
	tool := strings.TrimSpace(StringArg(payload["tool"]))
	if tool != "" {
		switch tool {
		case astockcontracts.ToolAStockQuoteLookup, astockcontracts.ToolAStockResearchLookup, astockcontracts.ToolAStockSignalLookup:
			return ""
		default:
			return "unsupported_tool_identity"
		}
	}
	markers := 0
	for _, key := range []string{"finance_report", "latest_report", "quote_valuation", "signals", "boundary", "boundary_note"} {
		if _, ok := payload[key]; ok {
			markers++
		}
	}
	if markers >= 2 {
		return "non_standard_composite_payload"
	}
	return "missing_standard_tool_identity"
}

func payloadReadinessContractReason(payload map[string]any) string {
	readiness := mapArg(payload["readiness"])
	if len(readiness) == 0 {
		return "missing_readiness_contract"
	}
	if _, ok := readiness["answer_ready"]; !ok {
		return "missing_readiness_answer_ready"
	}
	return ""
}

func cloneStringAnyMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in)+1)
	for key, value := range in {
		out[key] = value
	}
	return out
}

type researchFormatPayload struct {
	Tool          string                         `json:"tool,omitempty"`
	Source        string                         `json:"source,omitempty"`
	AdapterID     string                         `json:"adapter_id,omitempty"`
	AdapterStatus astockcontracts.AdapterStatus  `json:"adapter_status,omitempty"`
	FailureCode   astockcontracts.FailureCode    `json:"failure_code,omitempty"`
	Subject       astockcontracts.Subject        `json:"subject,omitempty"`
	Freshness     astockcontracts.Freshness      `json:"freshness,omitempty"`
	Evidence      astockcontracts.SourceEvidence `json:"evidence,omitempty"`
	Readiness     astockcontracts.Readiness      `json:"readiness,omitempty"`
	Reports       []researchFormatReport         `json:"reports,omitempty"`
	Warnings      []string                       `json:"warnings,omitempty"`
}

type researchFormatReport struct {
	Title       string                         `json:"title,omitempty"`
	Institution string                         `json:"institution,omitempty"`
	Analyst     string                         `json:"analyst,omitempty"`
	PublishedAt string                         `json:"published_at,omitempty"`
	Rating      string                         `json:"rating,omitempty"`
	Summary     string                         `json:"summary,omitempty"`
	PDFURL      string                         `json:"pdf_url,omitempty"`
	SourceURL   string                         `json:"source_url,omitempty"`
	Evidence    astockcontracts.SourceEvidence `json:"evidence,omitempty"`
}

func inferAnswerKind(payload map[string]any) string {
	tool := strings.TrimSpace(StringArg(payload["tool"]))
	switch tool {
	case astockcontracts.ToolAStockQuoteLookup:
		return "valuation"
	case astockcontracts.ToolAStockResearchLookup:
		return "research"
	case astockcontracts.ToolAStockSignalLookup:
		return "signal"
	default:
		return ""
	}
}

func decodePayload(in map[string]any, out any) error {
	blob, err := json.Marshal(in)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(blob, out); err != nil {
		return err
	}
	return nil
}

func formatQuoteAnswer(payload astockcontracts.QuotePayload) string {
	if !payload.Readiness.AnswerReady {
		return formatBlockedAnswer(payload.Subject, payload.AdapterStatus, payload.FailureCode, payload.Readiness.DegradeReason)
	}
	subject := formatSubject(payload.Subject)
	lines := []string{
		fmt.Sprintf("%s行情和估值快照如下（%s）：", subject, formatAsOf(payload.Freshness.AsOf, payload.Evidence.AsOf)),
	}
	if item := formatMetric("价格", payload.Quote.Price); item != "" {
		lines = append(lines, "- "+item)
	}
	if item := formatMetric("涨跌幅", payload.Quote.ChangePercent); item != "" {
		lines = append(lines, "- "+item)
	}
	if item := formatMetric("换手率", payload.Quote.TurnoverPercent); item != "" {
		lines = append(lines, "- "+item)
	}
	if item := formatMetric("PE(TTM)", payload.Quote.PETTM); item != "" {
		lines = append(lines, "- "+item)
	}
	if item := formatMetric("静态 PE", payload.Quote.PEStatic); item != "" {
		lines = append(lines, "- "+item)
	}
	if item := formatMetric("PB", payload.Quote.PB); item != "" {
		lines = append(lines, "- "+item)
	}
	if item := formatMetric("总市值", payload.Quote.MarketCap); item != "" {
		lines = append(lines, "- "+item)
	}
	if item := formatMetric("流通市值", payload.Quote.FloatMarketCap); item != "" {
		lines = append(lines, "- "+item)
	}
	if source := firstNonEmpty(payload.Evidence.SourceURL, payload.Subject.Evidence.SourceURL); source != "" {
		lines = append(lines, "来源："+source)
	}
	lines = append(lines, "说明：以上仅为点时行情/估值证据，不代表历史估值分位，不构成任何投资建议。")
	return strings.Join(lines, "\n")
}

func formatResearchAnswer(payload researchFormatPayload) string {
	if !payload.Readiness.AnswerReady {
		return formatBlockedAnswer(payload.Subject, payload.AdapterStatus, payload.FailureCode, payload.Readiness.DegradeReason)
	}
	subject := formatSubject(payload.Subject)
	lines := []string{
		fmt.Sprintf("%s最近研报和机构评级如下（%s）：", subject, formatAsOf(payload.Freshness.AsOf, payload.Evidence.AsOf)),
	}
	limit := minInt(len(payload.Reports), 5)
	for idx := 0; idx < limit; idx++ {
		report := payload.Reports[idx]
		line := fmt.Sprintf("%d. %s", idx+1, firstNonEmpty(report.Title, "未命名研报"))
		meta := compactStrings([]string{report.Institution, report.PublishedAt, report.Rating})
		if len(meta) > 0 {
			line += "（" + strings.Join(meta, "，") + "）"
		}
		if report.PDFURL != "" {
			line += " " + report.PDFURL
		}
		lines = append(lines, line)
	}
	if len(payload.Reports) == 0 {
		lines = append(lines, "- 未返回可用研报记录。")
	}
	if source := payload.Evidence.SourceURL; source != "" {
		lines = append(lines, "来源："+source)
	}
	lines = append(lines, "说明：研报评级和预测仅代表机构观点，不等同于确定性业绩，不构成任何投资建议。")
	return strings.Join(lines, "\n")
}

func formatSignalAnswer(payload astockcontracts.SignalPayload) string {
	if !payload.Readiness.AnswerReady {
		return formatBlockedAnswer(payload.Subject, payload.AdapterStatus, payload.FailureCode, payload.Readiness.DegradeReason)
	}
	subject := formatSubject(payload.Subject)
	if subject == "" {
		subject = "A 股市场"
	}
	lines := []string{
		fmt.Sprintf("%s相关市场信号如下（%s）：", subject, formatAsOf(payload.Freshness.TradeDate, payload.Freshness.AsOf, payload.Evidence.AsOf)),
	}
	limit := minInt(len(payload.Signals), 8)
	for idx := 0; idx < limit; idx++ {
		event := payload.Signals[idx]
		lines = append(lines, fmt.Sprintf("%d. %s", idx+1, formatSignalEvent(event)))
	}
	if len(payload.Signals) == 0 {
		lines = append(lines, "- 未返回可用市场信号。")
	}
	if source := payload.Evidence.SourceURL; source != "" {
		lines = append(lines, "来源："+source)
	}
	if payload.Readiness.Degraded || len(payload.Warnings) > 0 {
		lines = append(lines, "说明：部分信号存在降级或缺字段，请以来源页面为准。")
	} else {
		lines = append(lines, "说明：市场信号只用于解释当期公开数据，不构成投资建议。")
	}
	return strings.Join(lines, "\n")
}

func formatSignalEvent(event astockcontracts.SignalEvent) string {
	parts := compactStrings([]string{
		firstNonEmpty(event.Title, string(event.Type)),
		event.TradeDate,
		event.Reason,
	})
	if event.Concept != "" {
		parts = append(parts, "概念："+event.Concept)
	}
	if event.Industry != "" {
		parts = append(parts, "行业："+event.Industry)
	}
	if item := formatMetric("净买额", event.NetBuy); item != "" {
		parts = append(parts, item)
	}
	if item := formatMetric("金额", event.Amount); item != "" {
		parts = append(parts, item)
	}
	if item := formatMetric("比例", event.Ratio); item != "" {
		parts = append(parts, item)
	}
	if event.SourceURL != "" {
		parts = append(parts, event.SourceURL)
	}
	return strings.Join(parts, "；")
}

func formatBlockedAnswer(subject astockcontracts.Subject, status astockcontracts.AdapterStatus, failure astockcontracts.FailureCode, reason string) string {
	target := formatSubject(subject)
	if target == "" {
		target = "该 A 股任务"
	}
	detail := firstNonEmpty(string(failure), reason, string(status), "not_answer_ready")
	return fmt.Sprintf("%s当前不能形成可靠回答：%s。", target, detail)
}

func formatSubject(subject astockcontracts.Subject) string {
	name := strings.TrimSpace(subject.EntityName)
	code := strings.TrimSpace(subject.StockCode)
	market := strings.ToUpper(strings.TrimSpace(string(subject.Market)))
	if name == "" && code == "" {
		return ""
	}
	if code == "" {
		return name
	}
	if market != "" {
		code += "." + market
	}
	if name == "" {
		return code
	}
	return name + "（" + code + "）"
}

func formatMetric(label string, value astockcontracts.MetricValue) string {
	raw := strings.TrimSpace(value.Value)
	if raw == "" {
		return ""
	}
	unit := strings.TrimSpace(value.Unit)
	if unit != "" {
		raw += unit
	}
	return label + "：" + raw
}

func formatAsOf(values ...string) string {
	if value := firstNonEmpty(values...); value != "" {
		return "截至 " + value
	}
	return "时点未标明"
}

func mapArg(value any) map[string]any {
	switch typed := value.(type) {
	case map[string]any:
		return typed
	case map[string]string:
		out := make(map[string]any, len(typed))
		for key, item := range typed {
			out[key] = item
		}
		return out
	case string:
		out := map[string]any{}
		if strings.TrimSpace(typed) != "" {
			_ = json.Unmarshal([]byte(typed), &out)
		}
		return out
	default:
		blob, err := json.Marshal(typed)
		if err != nil {
			return nil
		}
		out := map[string]any{}
		if err := json.Unmarshal(blob, &out); err != nil {
			return nil
		}
		return out
	}
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
