package tools

import (
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

func browserApplyActionabilityEvidence(
	action string,
	status string,
	note string,
	recoveryAction string,
	ref string,
	selector string,
	snapshot string,
	snapshotFormat string,
	snapshotRefs string,
	snapshotFrame string,
	elements []BrowserSnapshotElement,
	snapshotTruncated bool,
	artifactPath string,
	finalURL string,
	title string,
	messages []BrowserConsoleMessage,
	errors []BrowserErrorEntry,
	resolverOutcome *agentxbrowserruntime.BrowserElementResolverOutcome,
	actionability **agentxbrowserruntime.BrowserActionabilityReport,
	failureEvidence **agentxbrowserruntime.BrowserActionFailureEvidence,
) {
	if actionability == nil || failureEvidence == nil {
		return
	}
	if *actionability == nil {
		*actionability = agentxbrowserruntime.BuildBrowserActionabilityReport(
			agentxbrowserruntime.BrowserActionabilityReportRequest{
				Action:          action,
				ElementRef:      ref,
				Selector:        selector,
				ResolverOutcome: resolverOutcome,
			},
		)
	}
	if *failureEvidence == nil {
		*failureEvidence = agentxbrowserruntime.BuildBrowserActionFailureEvidence(
			agentxbrowserruntime.BrowserActionFailureEvidenceRequest{
				Action:            action,
				Status:            status,
				Note:              note,
				RecoveryAction:    recoveryAction,
				ElementRef:        ref,
				Selector:          selector,
				ResolverOutcome:   resolverOutcome,
				Actionability:     *actionability,
				SnapshotFormat:    snapshotFormat,
				SnapshotRefs:      snapshotRefs,
				SnapshotFrame:     snapshotFrame,
				ElementCount:      len(elements),
				SnapshotText:      snapshot,
				SnapshotTruncated: snapshotTruncated,
				ArtifactPath:      artifactPath,
				FinalURL:          finalURL,
				Title:             title,
				ConsoleMessages:   messages,
				Errors:            errors,
			},
		)
	}
}

func browserActResultWithActionabilityEvidence(result BrowserActResult) BrowserActResult {
	browserApplyActionabilityEvidence(
		result.Kind,
		result.Status,
		result.Note,
		result.RecoveryAction,
		result.Ref,
		result.Selector,
		result.Snapshot,
		result.SnapshotFormat,
		result.SnapshotRefs,
		result.SnapshotFrame,
		result.Elements,
		result.Truncated,
		firstNonEmpty(strings.TrimSpace(result.Path), firstStringFromSlice(result.Paths)),
		result.FinalURL,
		result.Title,
		result.Messages,
		result.Errors,
		result.ResolverOutcome,
		&result.Actionability,
		&result.FailureEvidence,
	)
	return result
}

func firstStringFromSlice(values []string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
