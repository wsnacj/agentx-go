package browserops

import (
	"fmt"
	"sort"
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

// BrowserActionFailurePayloadEvaluationInput is the pack-level evidence shape
// for browser actionability regression gates. It keeps browser failure payload
// quality in the browserops pack instead of hard-coding case-specific checks in
// generic runtime or eval code.
type BrowserActionFailurePayloadEvaluationInput struct {
	Payloads                []BrowserActionFailurePayloadEvidence
	RequiredFailedChecks    []string
	RequireTraceArtifact    bool
	RequireSnapshotEvidence bool
	MinDistinctFailedChecks int
}

// BrowserActionFailurePayloadEvidence is the subset of browser_act payloads
// needed by the pack-level failure-payload gate.
type BrowserActionFailurePayloadEvidence struct {
	Action          string
	Status          string
	Actionability   *agentxbrowserruntime.BrowserActionabilityReport
	FailureEvidence *agentxbrowserruntime.BrowserActionFailureEvidence
}

// BrowserActionFailurePayloadEvaluation is the stable result contract emitted
// by the browserops failed-action payload evaluator.
type BrowserActionFailurePayloadEvaluation struct {
	Passed                   bool     `json:"passed"`
	Summary                  string   `json:"summary"`
	PayloadCount             int      `json:"payload_count"`
	ValidPayloadCount        int      `json:"valid_payload_count"`
	DistinctFailedChecks     []string `json:"distinct_failed_checks"`
	DistinctFailedCheckCount int      `json:"distinct_failed_check_count"`
	MissingFailedChecks      []string `json:"missing_failed_checks"`
	TraceArtifactReady       bool     `json:"trace_artifact_ready"`
	SnapshotEvidenceReady    bool     `json:"snapshot_evidence_ready"`
	FailureReasons           []string `json:"failure_reasons"`
}

func EvaluateBrowserActionFailurePayloadEvidence(input BrowserActionFailurePayloadEvaluationInput) BrowserActionFailurePayloadEvaluation {
	out := BrowserActionFailurePayloadEvaluation{
		PayloadCount:          len(input.Payloads),
		TraceArtifactReady:    true,
		SnapshotEvidenceReady: true,
	}
	required := normalizeBrowserActionFailureCheckList(input.RequiredFailedChecks)
	seenRequired := map[string]bool{}
	distinct := map[string]bool{}
	reasons := []string{}
	if len(input.Payloads) == 0 {
		reasons = append(reasons, "payloads_missing")
		out.TraceArtifactReady = !input.RequireTraceArtifact
		out.SnapshotEvidenceReady = !input.RequireSnapshotEvidence
	}
	for idx, payload := range input.Payloads {
		result := evaluateSingleBrowserActionFailurePayload(idx, payload, input)
		if result.FailedCheck != "" {
			distinct[result.FailedCheck] = true
			if containsBrowserActionFailureCheck(required, result.FailedCheck) {
				seenRequired[result.FailedCheck] = true
			}
		}
		if result.Valid {
			out.ValidPayloadCount++
		}
		if !result.TraceArtifactReady {
			out.TraceArtifactReady = false
		}
		if !result.SnapshotEvidenceReady {
			out.SnapshotEvidenceReady = false
		}
		reasons = append(reasons, result.FailureReasons...)
	}
	out.DistinctFailedChecks = sortedBrowserActionFailureChecks(distinct)
	out.DistinctFailedCheckCount = len(out.DistinctFailedChecks)
	for _, check := range required {
		if !seenRequired[check] {
			out.MissingFailedChecks = append(out.MissingFailedChecks, check)
		}
	}
	if len(out.MissingFailedChecks) > 0 {
		reasons = append(reasons, "required_failed_checks_missing")
	}
	if input.MinDistinctFailedChecks > 0 && out.DistinctFailedCheckCount < input.MinDistinctFailedChecks {
		reasons = append(reasons, "distinct_failed_checks_below_minimum")
	}
	out.FailureReasons = uniqueBrowserActionFailureReasons(reasons)
	out.Passed = len(out.FailureReasons) == 0
	out.Summary = browserActionFailurePayloadEvaluationSummary(out)
	return out
}

type browserActionFailurePayloadItemEvaluation struct {
	Valid                 bool
	FailedCheck           string
	TraceArtifactReady    bool
	SnapshotEvidenceReady bool
	FailureReasons        []string
}

func evaluateSingleBrowserActionFailurePayload(idx int, payload BrowserActionFailurePayloadEvidence, input BrowserActionFailurePayloadEvaluationInput) browserActionFailurePayloadItemEvaluation {
	result := browserActionFailurePayloadItemEvaluation{
		TraceArtifactReady:    true,
		SnapshotEvidenceReady: true,
	}
	prefix := fmt.Sprintf("payload_%d", idx)
	action := strings.TrimSpace(payload.Action)
	if action == "" && payload.Actionability != nil {
		action = strings.TrimSpace(payload.Actionability.Action)
	}
	if action == "" && payload.FailureEvidence != nil {
		action = strings.TrimSpace(payload.FailureEvidence.Action)
	}
	if strings.TrimSpace(payload.Status) != "action_failed" {
		result.FailureReasons = append(result.FailureReasons, prefix+":status_not_action_failed")
	}
	actionability := payload.Actionability
	if actionability == nil {
		result.FailureReasons = append(result.FailureReasons, prefix+":actionability_missing")
	} else {
		result.FailedCheck = strings.TrimSpace(actionability.FailedCheck)
		if strings.TrimSpace(actionability.Status) != agentxbrowserruntime.BrowserActionabilityStatusFailed {
			result.FailureReasons = append(result.FailureReasons, prefix+":actionability_not_failed")
		}
		if result.FailedCheck == "" {
			result.FailureReasons = append(result.FailureReasons, prefix+":failed_check_missing")
		}
		if actionability.FailureReason == "" {
			result.FailureReasons = append(result.FailureReasons, prefix+":failure_reason_missing")
		}
		if actionability.ManualRetryHint == "" {
			result.FailureReasons = append(result.FailureReasons, prefix+":manual_retry_hint_missing")
		}
		if actionability.RecoveryAction == "" {
			result.FailureReasons = append(result.FailureReasons, prefix+":recovery_action_missing")
		}
	}
	evidence := payload.FailureEvidence
	if evidence == nil {
		result.FailureReasons = append(result.FailureReasons, prefix+":failure_evidence_missing")
		result.TraceArtifactReady = !input.RequireTraceArtifact
		result.SnapshotEvidenceReady = !input.RequireSnapshotEvidence
	} else {
		result = evaluateBrowserActionFailureEvidencePayload(prefix, action, result, evidence, input)
	}
	result.Valid = len(result.FailureReasons) == 0
	return result
}

func evaluateBrowserActionFailureEvidencePayload(prefix string, action string, result browserActionFailurePayloadItemEvaluation, evidence *agentxbrowserruntime.BrowserActionFailureEvidence, input BrowserActionFailurePayloadEvaluationInput) browserActionFailurePayloadItemEvaluation {
	if strings.TrimSpace(evidence.Status) != "action_failed" {
		result.FailureReasons = append(result.FailureReasons, prefix+":failure_evidence_status_not_action_failed")
	}
	if action != "" && strings.TrimSpace(evidence.Action) != "" && strings.TrimSpace(evidence.Action) != action {
		result.FailureReasons = append(result.FailureReasons, prefix+":failure_evidence_action_mismatch")
	}
	expectedReason := browserActionFailureExpectedReasonCode(result.FailedCheck, evidence.Actionability)
	if expectedReason != "" && strings.TrimSpace(evidence.ReasonCode) != expectedReason {
		result.FailureReasons = append(result.FailureReasons, prefix+":reason_code_mismatch")
	}
	if strings.TrimSpace(evidence.RecoveryAction) == "" {
		result.FailureReasons = append(result.FailureReasons, prefix+":failure_evidence_recovery_action_missing")
	}
	if nested := evidence.Actionability; nested == nil {
		result.FailureReasons = append(result.FailureReasons, prefix+":failure_evidence_actionability_missing")
	} else if strings.TrimSpace(nested.FailedCheck) != result.FailedCheck || strings.TrimSpace(nested.Status) != agentxbrowserruntime.BrowserActionabilityStatusFailed {
		result.FailureReasons = append(result.FailureReasons, prefix+":failure_evidence_actionability_mismatch")
	}
	if input.RequireSnapshotEvidence && !browserActionFailureEvidenceHasSnapshot(evidence) {
		result.SnapshotEvidenceReady = false
		result.FailureReasons = append(result.FailureReasons, prefix+":snapshot_evidence_missing")
	}
	if input.RequireTraceArtifact {
		if !browserActionFailureEvidenceHasTraceArtifact(evidence, result.FailedCheck, expectedReason) {
			result.TraceArtifactReady = false
			result.FailureReasons = append(result.FailureReasons, prefix+":trace_artifact_missing")
		}
	}
	return result
}

func browserActionFailureExpectedReasonCode(failedCheck string, nested *agentxbrowserruntime.BrowserActionabilityReport) string {
	if nested != nil {
		if reason := strings.TrimSpace(nested.FailureReason); reason != "" {
			return reason
		}
	}
	failedCheck = strings.TrimSpace(failedCheck)
	if failedCheck == "" {
		return ""
	}
	return "actionability_" + failedCheck + "_failed"
}

func browserActionFailureEvidenceHasSnapshot(evidence *agentxbrowserruntime.BrowserActionFailureEvidence) bool {
	if evidence == nil {
		return false
	}
	if evidence.SnapshotAvailable || evidence.SnapshotElementCount > 0 || strings.TrimSpace(evidence.SnapshotRefs) != "" || strings.TrimSpace(evidence.SnapshotFormat) != "" {
		return true
	}
	artifact := evidence.Artifact
	return artifact != nil && (artifact.SnapshotAvailable || artifact.SnapshotElementCount > 0 || strings.TrimSpace(artifact.SnapshotRefs) != "" || strings.TrimSpace(artifact.SnapshotFormat) != "")
}

func browserActionFailureEvidenceHasTraceArtifact(evidence *agentxbrowserruntime.BrowserActionFailureEvidence, failedCheck string, expectedReason string) bool {
	if evidence == nil || evidence.Artifact == nil {
		return false
	}
	artifact := evidence.Artifact
	if strings.TrimSpace(artifact.Kind) != "trace_like" || strings.TrimSpace(artifact.Status) != "action_failed" {
		return false
	}
	if failedCheck != "" && strings.TrimSpace(artifact.FailedCheck) != failedCheck {
		return false
	}
	if expectedReason != "" && strings.TrimSpace(artifact.ReasonCode) != expectedReason {
		return false
	}
	return true
}

func normalizeBrowserActionFailureCheckList(values []string) []string {
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

func containsBrowserActionFailureCheck(values []string, target string) bool {
	target = strings.TrimSpace(target)
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func sortedBrowserActionFailureChecks(values map[string]bool) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			out = append(out, value)
		}
	}
	sort.Strings(out)
	return out
}

func uniqueBrowserActionFailureReasons(values []string) []string {
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

func browserActionFailurePayloadEvaluationSummary(eval BrowserActionFailurePayloadEvaluation) string {
	if eval.Passed {
		return fmt.Sprintf("validated %d failed browser action payloads across %d failed checks", eval.ValidPayloadCount, eval.DistinctFailedCheckCount)
	}
	return fmt.Sprintf("validated %d/%d failed browser action payloads; missing_checks=%s failures=%s",
		eval.ValidPayloadCount,
		eval.PayloadCount,
		strings.Join(eval.MissingFailedChecks, ","),
		strings.Join(eval.FailureReasons, ","),
	)
}
