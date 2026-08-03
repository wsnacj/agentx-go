package finance

import (
	"path/filepath"
	"strings"

	financialreportmetrics "github.com/wsnacj/agentx-go/scenes/finance/metrics"
)

// DocumentMetricEvidenceProjectionInput carries host/project context for
// projecting parser-level report evidence into the standard report_metrics
// evidence shape. Parser execution and source-specific repair stay outside the
// reports kit.
type DocumentMetricEvidenceProjectionInput struct {
	Evidence                financialreportmetrics.ReportDocumentMetricEvidence
	EntityName              string
	StockCode               string
	Ticker                  string
	SourceURL               string
	Title                   string
	ReportPath              string
	SelectionReasonFallback string
	PageTitleFallback       string
}

// MetricsEvidenceFromDocumentMetricEvidence projects pack-level document parser
// evidence into the standard adapter-neutral metrics evidence DTO.
func MetricsEvidenceFromDocumentMetricEvidence(input DocumentMetricEvidenceProjectionInput) MetricsEvidence {
	evidence := input.Evidence
	return MetricsEvidence{
		CompanyName:       firstNonEmptyString(evidence.CompanyName, input.EntityName, "unknown"),
		StockCode:         firstNonEmptyString(evidence.StockCode, input.StockCode, input.Ticker, "unknown"),
		SelectionReason:   firstNonEmptyString(evidence.SelectionReason, input.SelectionReasonFallback, "selected_by_document_metric_evidence_adapter"),
		OfficialSource:    firstNonEmptyString(evidence.OfficialSource, input.SourceURL, "unknown"),
		ReportPeriod:      firstNonEmptyString(evidence.ReportPeriod, "unknown"),
		Revenue:           firstNonEmptyString(evidence.Revenue, "unknown"),
		RevenueGrowth:     firstNonEmptyString(evidence.RevenueGrowth, "unknown"),
		NetProfit:         firstNonEmptyString(evidence.NetProfit, "unknown"),
		NetProfitGrowth:   firstNonEmptyString(evidence.NetProfitGrowth, "unknown"),
		OperatingCashFlow: firstNonEmptyString(evidence.OperatingCashFlow, "unknown"),
		PageTitle:         documentMetricEvidencePageTitle(input),
		MetricEvidence:    evidence.FieldEvidence,
	}
}

func documentMetricEvidencePageTitle(input DocumentMetricEvidenceProjectionInput) string {
	if title := strings.TrimSpace(input.Title); title != "" {
		return title
	}
	path := firstNonEmptyString(input.Evidence.ArtifactPath, input.ReportPath)
	if path != "" {
		base := strings.TrimSpace(filepath.Base(path))
		if base != "" && base != "." && base != string(filepath.Separator) {
			return base
		}
	}
	return firstNonEmptyString(input.PageTitleFallback, "document_report_metrics")
}
