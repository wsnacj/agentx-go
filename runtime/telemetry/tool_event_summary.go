package telemetry

import "strings"

const (
	ToolEventSurfaceRetrieval = "retrieval"
	ToolEventSurfaceBrowser   = "browser"
	ToolEventSurfacePDF       = "pdf"
	ToolEventSurfaceExec      = "exec"
	ToolEventSurfaceOther     = "other"
)

type ToolEventSummary struct {
	TotalEvents                    int            `json:"total_events"`
	ByKind                         map[string]int `json:"by_kind,omitempty"`
	ByTool                         map[string]int `json:"by_tool,omitempty"`
	BySurface                      map[string]int `json:"by_surface,omitempty"`
	ByRuntimeDecisionReason        map[string]int `json:"by_runtime_decision_reason,omitempty"`
	ByRuntimeDecisionSubject       map[string]int `json:"by_runtime_decision_subject,omitempty"`
	ByRuntimeDecisionPolicySource  map[string]int `json:"by_runtime_decision_policy_source,omitempty"`
	BySoftRejectionSource          map[string]int `json:"by_soft_rejection_source,omitempty"`
	BySoftRejectionReason          map[string]int `json:"by_soft_rejection_reason,omitempty"`
	Failed                         int            `json:"failed,omitempty"`
	Retried                        int            `json:"retried,omitempty"`
	RuntimeDecisions               int            `json:"runtime_decisions,omitempty"`
	RuntimeDecisionDenied          int            `json:"runtime_decision_denied,omitempty"`
	RuntimeDecisionDegraded        int            `json:"runtime_decision_degraded,omitempty"`
	RuntimeDecisionRequiresConfirm int            `json:"runtime_decision_requires_confirm,omitempty"`
	SoftRejections                 int            `json:"soft_rejections,omitempty"`
	SoftRejectContent              int            `json:"soft_reject_content,omitempty"`
	SoftHalt                       int            `json:"soft_halt,omitempty"`
	ProviderFallbacks              int            `json:"provider_fallbacks,omitempty"`
	ResultMiddlewareObserved       int            `json:"result_middleware_observed,omitempty"`
	ResultMiddlewareApplied        int            `json:"result_middleware_applied,omitempty"`
	OutputSchemaObserved           int            `json:"output_schema_observed,omitempty"`
	OutputSchemaDrift              int            `json:"output_schema_drift,omitempty"`
}

func SummarizeToolEvents(events []ToolEvent) ToolEventSummary {
	summary := ToolEventSummary{
		ByKind:                        map[string]int{},
		ByTool:                        map[string]int{},
		BySurface:                     map[string]int{},
		ByRuntimeDecisionReason:       map[string]int{},
		ByRuntimeDecisionSubject:      map[string]int{},
		ByRuntimeDecisionPolicySource: map[string]int{},
		BySoftRejectionSource:         map[string]int{},
		BySoftRejectionReason:         map[string]int{},
	}
	for _, event := range events {
		event = normalizeToolEvent(event)
		if event.Kind == "" {
			continue
		}
		summary.TotalEvents++
		summary.ByKind[event.Kind]++
		if event.Tool != "" {
			summary.ByTool[event.Tool]++
			summary.BySurface[ToolEventSurfaceForTool(event.Tool)]++
		}
		switch event.Kind {
		case ToolEventKindFailed:
			summary.Failed++
		case ToolEventKindRetried:
			summary.Retried++
		case ToolEventKindProviderFallback:
			summary.ProviderFallbacks++
		case ToolEventKindResultMiddlewareObserved:
			summary.ResultMiddlewareObserved++
		case ToolEventKindResultMiddlewareApplied:
			summary.ResultMiddlewareApplied++
		}
		if event.ResultMiddleware != nil && event.ResultMiddleware.OutputSchema != nil {
			summary.OutputSchemaObserved++
			if event.ResultMiddleware.OutputSchema.Drift {
				summary.OutputSchemaDrift++
			}
		}
		if event.SoftRejection != nil {
			count := event.SoftRejection.Count
			if count <= 0 {
				count = 1
			}
			summary.SoftRejections += count
			for _, action := range primaryOrListedToolEventValues(event.SoftRejection.Action, event.SoftRejection.Actions) {
				switch strings.ToLower(strings.TrimSpace(action)) {
				case "reject_content":
					summary.SoftRejectContent += softRejectionBucketIncrement(count, event.SoftRejection.Actions)
				case "halt":
					summary.SoftHalt += softRejectionBucketIncrement(count, event.SoftRejection.Actions)
				}
			}
			for _, source := range primaryOrListedToolEventValues(event.SoftRejection.Source, event.SoftRejection.Sources) {
				summary.BySoftRejectionSource[source] += softRejectionBucketIncrement(count, event.SoftRejection.Sources)
			}
			for _, reason := range primaryOrListedToolEventValues(event.SoftRejection.Reason, event.SoftRejection.Reasons) {
				summary.BySoftRejectionReason[reason] += softRejectionBucketIncrement(count, event.SoftRejection.Reasons)
			}
		}
		if event.RuntimeDecision != nil && event.RuntimeDecision.Checked {
			summary.RuntimeDecisions++
			if event.RuntimeDecision.Denied {
				summary.RuntimeDecisionDenied++
			}
			if event.RuntimeDecision.Degraded {
				summary.RuntimeDecisionDegraded++
			}
			if event.RuntimeDecision.RequiresConfirm {
				summary.RuntimeDecisionRequiresConfirm++
			}
			if reason := strings.TrimSpace(firstNonEmptyString(event.RuntimeDecision.Reason, event.Reason)); reason != "" {
				summary.ByRuntimeDecisionReason[reason]++
			}
			if subject := strings.TrimSpace(event.RuntimeDecision.DecisionSubject); subject != "" {
				summary.ByRuntimeDecisionSubject[subject]++
			}
			if policySource := strings.TrimSpace(event.RuntimeDecision.PolicySource); policySource != "" {
				summary.ByRuntimeDecisionPolicySource[policySource]++
			}
		}
	}
	if len(summary.ByKind) == 0 {
		summary.ByKind = nil
	}
	if len(summary.ByTool) == 0 {
		summary.ByTool = nil
	}
	if len(summary.BySurface) == 0 {
		summary.BySurface = nil
	}
	if len(summary.ByRuntimeDecisionReason) == 0 {
		summary.ByRuntimeDecisionReason = nil
	}
	if len(summary.ByRuntimeDecisionSubject) == 0 {
		summary.ByRuntimeDecisionSubject = nil
	}
	if len(summary.ByRuntimeDecisionPolicySource) == 0 {
		summary.ByRuntimeDecisionPolicySource = nil
	}
	if len(summary.BySoftRejectionSource) == 0 {
		summary.BySoftRejectionSource = nil
	}
	if len(summary.BySoftRejectionReason) == 0 {
		summary.BySoftRejectionReason = nil
	}
	return summary
}

func primaryOrListedToolEventValues(primary string, listed []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, value := range append([]string{primary}, listed...) {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" || seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		out = append(out, trimmed)
	}
	return out
}

func softRejectionBucketIncrement(count int, listed []string) int {
	if count <= 0 {
		return 1
	}
	if len(listed) <= 1 {
		return count
	}
	return 1
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func ToolEventSurfaceForTool(tool string) string {
	normalized := strings.ToLower(strings.TrimSpace(tool))
	if normalized == "" {
		return ToolEventSurfaceOther
	}
	switch normalized {
	case "search", "web_search", "open_page", "find_in_page", "web_fetch", "http_request":
		return ToolEventSurfaceRetrieval
	case "browser", "browserruntime", "browser_runtime":
		return ToolEventSurfaceBrowser
	case "pdf", "pdf_extract", "pdf_read_pages", "pdf_extract_structured":
		return ToolEventSurfacePDF
	case "exec", "process", "shell":
		return ToolEventSurfaceExec
	default:
		switch {
		case strings.HasPrefix(normalized, "browser_"):
			return ToolEventSurfaceBrowser
		case strings.HasPrefix(normalized, "pdf_"):
			return ToolEventSurfacePDF
		default:
			return ToolEventSurfaceOther
		}
	}
}
