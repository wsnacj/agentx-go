package tools

import (
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

type browserResolverFallbackSummary struct {
	ResolvedViaFallback bool
	Kind                string
	Index               *int
	CandidateStrength   string
	BlockedBy           string
	AmbiguityClass      string
	ManualRetryHint     string
	SpecificityFields   []string
}

type browserResolverFallbackExplanationSummary struct {
	State           string `json:"state,omitempty"`
	SummaryCode     string `json:"summary_code,omitempty"`
	ManualRetryHint string `json:"manual_retry_hint,omitempty"`
}

type browserDiagnosticsExplanationSummary struct {
	Category        string `json:"category,omitempty"`
	State           string `json:"state,omitempty"`
	SummaryCode     string `json:"summary_code,omitempty"`
	NextStepAlias   string `json:"next_step_alias,omitempty"`
	ManualRetryHint string `json:"manual_retry_hint,omitempty"`
}

func browserApplyActionSuccessSummaryAliases(
	kind string,
	status string,
	diagnosticsExplanation **browserDiagnosticsExplanationSummary,
	explanation **browserTopLevelSummary,
	diagnostics **browserTopLevelSummary,
	summary **browserTopLevelSummary,
	display **browserTopLevelDisplaySummary,
) {
	if diagnosticsExplanation == nil || explanation == nil || diagnostics == nil || summary == nil || display == nil {
		return
	}
	projection := agentxbrowserruntime.BuildSharedSessionBrowserGuidanceProjection(
		agentxbrowserruntime.SharedSessionBrowserGuidanceProjectionRequest{
			ActionKind:   kind,
			ActionStatus: status,
		},
	)
	if projection.DiagnosticsExplanation == nil {
		return
	}
	successSummary := browserDiagnosticsExplanationSummaryFromShared(projection.DiagnosticsExplanation)
	if *diagnosticsExplanation == nil {
		*diagnosticsExplanation = successSummary
	}
	successExplanation := browserRuntimeTopLevelSummaryFromShared(projection.Explanation)
	successDiagnostics := browserRuntimeTopLevelSummaryFromShared(projection.Diagnostics)
	successTopLevel := browserRuntimeTopLevelSummaryFromShared(projection.Summary)
	successDisplay := browserRuntimeTopLevelDisplaySummaryFromShared(projection.Display)
	if *explanation == nil {
		*explanation = successExplanation
	}
	if *diagnostics == nil {
		*diagnostics = successDiagnostics
	}
	if *summary == nil {
		*summary = successTopLevel
	}
	if *display == nil {
		*display = successDisplay
	}
}

func browserApplyActionabilityFailureSummaryAliases(
	kind string,
	status string,
	actionability *agentxbrowserruntime.BrowserActionabilityReport,
	failureEvidence *agentxbrowserruntime.BrowserActionFailureEvidence,
	diagnosticsExplanation **browserDiagnosticsExplanationSummary,
	explanation **browserTopLevelSummary,
	diagnostics **browserTopLevelSummary,
	summary **browserTopLevelSummary,
	display **browserTopLevelDisplaySummary,
) {
	if diagnosticsExplanation == nil || explanation == nil || diagnostics == nil || summary == nil || display == nil {
		return
	}
	req := agentxbrowserruntime.SharedSessionBrowserGuidanceProjectionRequest{
		ActionKind:   kind,
		ActionStatus: status,
	}
	if !browserPopulateActionabilityGuidanceProjectionRequest(&req, actionability, failureEvidence) {
		return
	}
	projection := agentxbrowserruntime.BuildSharedSessionBrowserGuidanceProjection(req)
	if projection.DiagnosticsExplanation == nil || !strings.EqualFold(projection.DiagnosticsExplanation.Category, "actionability") {
		return
	}
	*diagnosticsExplanation = browserDiagnosticsExplanationSummaryFromShared(projection.DiagnosticsExplanation)
	*explanation = browserRuntimeTopLevelSummaryFromShared(projection.Explanation)
	*diagnostics = browserRuntimeTopLevelSummaryFromShared(projection.Diagnostics)
	*summary = browserRuntimeTopLevelSummaryFromShared(projection.Summary)
	*display = browserRuntimeTopLevelDisplaySummaryFromShared(projection.Display)
}

func browserPopulateActionabilityGuidanceProjectionRequest(
	req *agentxbrowserruntime.SharedSessionBrowserGuidanceProjectionRequest,
	actionability *agentxbrowserruntime.BrowserActionabilityReport,
	failureEvidence *agentxbrowserruntime.BrowserActionFailureEvidence,
) bool {
	if req == nil {
		return false
	}
	report := actionability
	if report == nil && failureEvidence != nil {
		report = failureEvidence.Actionability
	}
	status := ""
	failedCheck := ""
	failureReason := ""
	retryDisposition := ""
	manualRetryHint := ""
	recoveryAction := ""
	action := ""
	if report != nil {
		status = strings.TrimSpace(report.Status)
		failedCheck = strings.TrimSpace(report.FailedCheck)
		failureReason = strings.TrimSpace(report.FailureReason)
		retryDisposition = strings.TrimSpace(report.RetryDisposition)
		manualRetryHint = strings.TrimSpace(report.ManualRetryHint)
		recoveryAction = strings.TrimSpace(report.RecoveryAction)
		action = strings.TrimSpace(report.Action)
	}
	if failureEvidence != nil {
		if failedCheck == "" {
			failedCheck = browserActionabilityFailedCheckFromReasonCode(failureEvidence.ReasonCode)
		}
		if recoveryAction == "" {
			recoveryAction = strings.TrimSpace(failureEvidence.RecoveryAction)
		}
		if action == "" {
			action = strings.TrimSpace(failureEvidence.Action)
		}
		if req.FailureReasonCode == "" {
			req.FailureReasonCode = strings.TrimSpace(failureEvidence.ReasonCode)
		}
		if req.ActionStatus == "" {
			req.ActionStatus = strings.TrimSpace(failureEvidence.Status)
		}
	}
	if status == "" && failedCheck != "" {
		status = agentxbrowserruntime.BrowserActionabilityStatusFailed
	}
	if !strings.EqualFold(status, agentxbrowserruntime.BrowserActionabilityStatusFailed) || strings.TrimSpace(failedCheck) == "" {
		return false
	}
	if req.ActionKind == "" {
		req.ActionKind = action
	}
	req.ActionabilityStatus = status
	req.ActionabilityFailedCheck = failedCheck
	req.ActionabilityFailureReason = failureReason
	req.ActionabilityRetryDisposition = retryDisposition
	req.ActionabilityManualRetryHint = manualRetryHint
	req.ActionabilityRecoveryAction = recoveryAction
	return true
}

func browserActionabilityFailedCheckFromReasonCode(reasonCode string) string {
	code := strings.ToLower(strings.TrimSpace(reasonCode))
	if !strings.HasPrefix(code, "actionability_") || !strings.HasSuffix(code, "_failed") {
		return ""
	}
	return strings.TrimSuffix(strings.TrimPrefix(code, "actionability_"), "_failed")
}

func browserTabsSuccessKind(action string) string {
	switch strings.ToLower(strings.TrimSpace(action)) {
	case "list":
		return "list_tabs"
	case "focus":
		return "focus_tab"
	case "close":
		return "close_tab"
	default:
		return ""
	}
}

func browserApplyReviewSummaryAliases(
	status string,
	reviewDecision string,
	reviewReady bool,
	diagnosticsExplanation **browserDiagnosticsExplanationSummary,
	explanation **browserTopLevelSummary,
	diagnostics **browserTopLevelSummary,
	summary **browserTopLevelSummary,
	display **browserTopLevelDisplaySummary,
	review **browserReviewSurfaceSummary,
) {
	if diagnosticsExplanation == nil || explanation == nil || diagnostics == nil || summary == nil || display == nil || review == nil {
		return
	}
	projection := agentxbrowserruntime.BuildSharedSessionBrowserGuidanceProjection(
		agentxbrowserruntime.SharedSessionBrowserGuidanceProjectionRequest{
			ReviewStatus:   status,
			ReviewDecision: reviewDecision,
		},
	)
	if projection.DiagnosticsExplanation != nil {
		*diagnosticsExplanation = browserDiagnosticsExplanationSummaryFromShared(projection.DiagnosticsExplanation)
		*explanation = browserRuntimeTopLevelSummaryFromShared(projection.Explanation)
		*diagnostics = browserRuntimeTopLevelSummaryFromShared(projection.Diagnostics)
		*summary = browserRuntimeTopLevelSummaryFromShared(projection.Summary)
		*display = browserRuntimeTopLevelDisplaySummaryFromShared(projection.Display)
	}
	*review = browserRuntimeReviewSurfaceSummaryForTopLevel(
		reviewDecision,
		reviewReady,
		*explanation,
		*diagnostics,
		*summary,
		*display,
	)
}

func browserCloneDiagnosticsExplanationSummary(summary *browserDiagnosticsExplanationSummary) *browserDiagnosticsExplanationSummary {
	if summary == nil {
		return nil
	}
	cloned := *summary
	return &cloned
}

type browserResolverGuidanceSummary struct {
	BlockedBy         string
	AmbiguityClass    string
	CandidateKind     string
	CandidateStrength string
	RetryDisposition  string
	ManualRetryHint   string
	NextStepAlias     string
}

func browserResolverFallbackSummaryForOutcome(outcome *agentxbrowserruntime.BrowserElementResolverOutcome) browserResolverFallbackSummary {
	summary := browserResolverFallbackSummary{}
	normalized := outcome.Normalized()
	if normalized == nil || strings.TrimSpace(normalized.FallbackFromKind) == "" {
		return summary
	}
	summary.ResolvedViaFallback = true
	summary.Kind = strings.TrimSpace(normalized.FallbackFromKind)
	index := normalized.FallbackFromIndex
	summary.Index = &index
	summary.CandidateStrength = strings.TrimSpace(normalized.FallbackFromCandidateStrength)
	summary.BlockedBy = strings.TrimSpace(normalized.FallbackFromBlockedBy)
	summary.AmbiguityClass = strings.TrimSpace(normalized.FallbackFromAmbiguityClass)
	summary.ManualRetryHint = strings.TrimSpace(normalized.FallbackFromManualRetryHint)
	if len(normalized.FallbackFromSpecificityFields) > 0 {
		summary.SpecificityFields = append([]string(nil), normalized.FallbackFromSpecificityFields...)
	}
	return summary
}

func browserResolverFallbackSummaryFromFields(
	resolvedViaFallback bool,
	fallbackKind string,
	fallbackIndex *int,
	fallbackStrength string,
	fallbackBlockedBy string,
	fallbackAmbiguityClass string,
	fallbackManualRetryHint string,
	fallbackSpecificityFields []string,
) browserResolverFallbackSummary {
	summary := browserResolverFallbackSummary{
		ResolvedViaFallback: resolvedViaFallback,
		Kind:                strings.TrimSpace(fallbackKind),
		Index:               browserCloneOptionalInt(fallbackIndex),
		CandidateStrength:   strings.TrimSpace(fallbackStrength),
		BlockedBy:           strings.TrimSpace(fallbackBlockedBy),
		AmbiguityClass:      strings.TrimSpace(fallbackAmbiguityClass),
		ManualRetryHint:     strings.TrimSpace(fallbackManualRetryHint),
	}
	if len(fallbackSpecificityFields) > 0 {
		summary.SpecificityFields = append([]string(nil), fallbackSpecificityFields...)
	}
	if summary.ResolvedViaFallback ||
		summary.Kind != "" ||
		summary.Index != nil ||
		summary.CandidateStrength != "" ||
		summary.BlockedBy != "" ||
		summary.AmbiguityClass != "" ||
		summary.ManualRetryHint != "" ||
		len(summary.SpecificityFields) > 0 {
		summary.ResolvedViaFallback = true
	}
	return summary
}

func browserResolverGuidanceSummaryFromFields(
	blockedBy string,
	ambiguityClass string,
	candidateKind string,
	candidateStrength string,
	retryDisposition string,
	manualRetryHint string,
	nextStepAlias string,
) browserResolverGuidanceSummary {
	return browserResolverGuidanceSummary{
		BlockedBy:         strings.TrimSpace(blockedBy),
		AmbiguityClass:    strings.TrimSpace(ambiguityClass),
		CandidateKind:     strings.TrimSpace(candidateKind),
		CandidateStrength: strings.TrimSpace(candidateStrength),
		RetryDisposition:  strings.TrimSpace(retryDisposition),
		ManualRetryHint:   strings.TrimSpace(manualRetryHint),
		NextStepAlias:     strings.TrimSpace(nextStepAlias),
	}
}

func browserResolverFallbackExplanationSummaryForSummary(summary browserResolverFallbackSummary) *browserResolverFallbackExplanationSummary {
	state := browserResolverFallbackExplanationState(summary)
	summaryCode := browserResolverFallbackExplanationCode(summary)
	manualRetryHint := strings.TrimSpace(summary.ManualRetryHint)
	if state == "" && summaryCode == "" && manualRetryHint == "" {
		return nil
	}
	return &browserResolverFallbackExplanationSummary{
		State:           state,
		SummaryCode:     summaryCode,
		ManualRetryHint: manualRetryHint,
	}
}

func browserResolverExplanationSummaryForSummaries(
	fallback browserResolverFallbackSummary,
	guidance browserResolverGuidanceSummary,
) *browserRuntimeResolverExplanationSummary {
	if explanation := browserResolverFallbackExplanationSummaryForSummary(fallback); explanation != nil {
		return &browserRuntimeResolverExplanationSummary{
			State:           strings.TrimSpace(explanation.State),
			SummaryCode:     strings.TrimSpace(explanation.SummaryCode),
			ManualRetryHint: strings.TrimSpace(explanation.ManualRetryHint),
		}
	}
	state := browserResolverGuidanceExplanationState(guidance)
	summaryCode := browserResolverGuidanceExplanationCode(guidance)
	nextStepAlias := strings.TrimSpace(guidance.NextStepAlias)
	manualRetryHint := strings.TrimSpace(guidance.ManualRetryHint)
	if state == "" && summaryCode == "" && nextStepAlias == "" && manualRetryHint == "" {
		return nil
	}
	return &browserRuntimeResolverExplanationSummary{
		State:           state,
		SummaryCode:     summaryCode,
		NextStepAlias:   nextStepAlias,
		ManualRetryHint: manualRetryHint,
	}
}

func browserResolverFallbackExplanationState(summary browserResolverFallbackSummary) string {
	if summary.ResolvedViaFallback {
		return "resolved_via_fallback"
	}
	return ""
}

func browserResolverFallbackExplanationCode(summary browserResolverFallbackSummary) string {
	parts := make([]string, 0, 2)
	if kind := strings.TrimSpace(summary.Kind); kind != "" {
		parts = append(parts, kind)
	}
	switch {
	case strings.TrimSpace(summary.AmbiguityClass) != "":
		parts = append(parts, strings.TrimSpace(summary.AmbiguityClass))
	case strings.TrimSpace(summary.BlockedBy) != "":
		parts = append(parts, strings.TrimSpace(summary.BlockedBy))
	}
	return strings.Join(parts, "_")
}

func browserResolverGuidanceExplanationState(summary browserResolverGuidanceSummary) string {
	switch strings.TrimSpace(summary.RetryDisposition) {
	case "manual_only":
		return "manual_resolution_required"
	case "auto_retry_allowed":
		return "automatic_retry_available"
	}
	if strings.TrimSpace(summary.BlockedBy) != "" || strings.TrimSpace(summary.AmbiguityClass) != "" {
		return "resolver_attention_required"
	}
	if strings.TrimSpace(summary.NextStepAlias) != "" || strings.TrimSpace(summary.ManualRetryHint) != "" {
		return "guidance_available"
	}
	return ""
}

func browserResolverGuidanceExplanationCode(summary browserResolverGuidanceSummary) string {
	parts := make([]string, 0, 2)
	if candidateKind := strings.TrimSpace(summary.CandidateKind); candidateKind != "" {
		parts = append(parts, candidateKind)
	}
	switch {
	case strings.TrimSpace(summary.AmbiguityClass) != "":
		parts = append(parts, strings.TrimSpace(summary.AmbiguityClass))
	case strings.TrimSpace(summary.BlockedBy) != "":
		parts = append(parts, strings.TrimSpace(summary.BlockedBy))
	}
	return strings.Join(parts, "_")
}

func browserDiagnosticsExplanationSummaryForSummaries(
	fallback browserResolverFallbackSummary,
	guidance browserResolverGuidanceSummary,
) *browserDiagnosticsExplanationSummary {
	if explanation := browserResolverFallbackExplanationSummaryForSummary(fallback); explanation != nil {
		return &browserDiagnosticsExplanationSummary{
			Category:        "resolver_fallback",
			State:           strings.TrimSpace(explanation.State),
			SummaryCode:     strings.TrimSpace(explanation.SummaryCode),
			ManualRetryHint: strings.TrimSpace(explanation.ManualRetryHint),
		}
	}
	state := browserResolverGuidanceExplanationState(guidance)
	summaryCode := browserResolverGuidanceExplanationCode(guidance)
	nextStepAlias := strings.TrimSpace(guidance.NextStepAlias)
	manualRetryHint := strings.TrimSpace(guidance.ManualRetryHint)
	if state == "" && summaryCode == "" && nextStepAlias == "" && manualRetryHint == "" {
		return nil
	}
	return &browserDiagnosticsExplanationSummary{
		Category:        "resolver",
		State:           state,
		SummaryCode:     summaryCode,
		NextStepAlias:   nextStepAlias,
		ManualRetryHint: manualRetryHint,
	}
}

func browserTopLevelSummaryForSummaries(
	fallback browserResolverFallbackSummary,
	guidance browserResolverGuidanceSummary,
) *browserTopLevelSummary {
	if explanation := browserDiagnosticsExplanationSummaryForSummaries(fallback, guidance); explanation != nil {
		summary := &browserTopLevelSummary{
			Category:        strings.TrimSpace(explanation.Category),
			State:           strings.TrimSpace(explanation.State),
			SummaryCode:     strings.TrimSpace(explanation.SummaryCode),
			NextStepAlias:   strings.TrimSpace(explanation.NextStepAlias),
			ManualRetryHint: strings.TrimSpace(explanation.ManualRetryHint),
		}
		if strings.EqualFold(summary.State, "resolved_via_fallback") {
			summary.ResolvedViaFallback = true
		}
		if browserUnifiedSummaryEmpty(*summary) {
			return nil
		}
		return summary
	}
	if explanation := browserResolverExplanationSummaryForSummaries(fallback, guidance); explanation != nil {
		summary := &browserTopLevelSummary{
			State:           strings.TrimSpace(explanation.State),
			SummaryCode:     strings.TrimSpace(explanation.SummaryCode),
			NextStepAlias:   strings.TrimSpace(explanation.NextStepAlias),
			ManualRetryHint: strings.TrimSpace(explanation.ManualRetryHint),
		}
		if strings.EqualFold(summary.State, "resolved_via_fallback") {
			summary.Category = "resolver_fallback"
			summary.ResolvedViaFallback = true
		} else if summary.State != "" || summary.SummaryCode != "" || summary.NextStepAlias != "" || summary.ManualRetryHint != "" {
			summary.Category = "resolver"
		}
		if browserUnifiedSummaryEmpty(*summary) {
			return nil
		}
		return summary
	}
	return nil
}

func browserApplyResolverFallbackSummary(
	outcome *agentxbrowserruntime.BrowserElementResolverOutcome,
	resolvedViaFallback *bool,
	fallbackKind *string,
	fallbackIndex **int,
	fallbackStrength *string,
	fallbackBlockedBy *string,
	fallbackAmbiguityClass *string,
	fallbackManualRetryHint *string,
	fallbackSpecificityFields *[]string,
) {
	if resolvedViaFallback == nil || fallbackKind == nil || fallbackIndex == nil || fallbackStrength == nil || fallbackBlockedBy == nil || fallbackAmbiguityClass == nil || fallbackManualRetryHint == nil || fallbackSpecificityFields == nil {
		return
	}
	if browserNormalizeExistingResolverFallbackSummary(resolvedViaFallback, fallbackKind, fallbackIndex, fallbackStrength, fallbackBlockedBy, fallbackAmbiguityClass, fallbackManualRetryHint, fallbackSpecificityFields) {
		return
	}
	summary := browserResolverFallbackSummaryForOutcome(outcome)
	*resolvedViaFallback = summary.ResolvedViaFallback
	*fallbackKind = summary.Kind
	*fallbackIndex = browserCloneOptionalInt(summary.Index)
	*fallbackStrength = summary.CandidateStrength
	*fallbackBlockedBy = summary.BlockedBy
	*fallbackAmbiguityClass = summary.AmbiguityClass
	*fallbackManualRetryHint = summary.ManualRetryHint
	if len(summary.SpecificityFields) > 0 {
		*fallbackSpecificityFields = append([]string(nil), summary.SpecificityFields...)
	} else {
		*fallbackSpecificityFields = nil
	}
}

func browserNormalizeExistingResolverFallbackSummary(
	resolvedViaFallback *bool,
	fallbackKind *string,
	fallbackIndex **int,
	fallbackStrength *string,
	fallbackBlockedBy *string,
	fallbackAmbiguityClass *string,
	fallbackManualRetryHint *string,
	fallbackSpecificityFields *[]string,
) bool {
	if resolvedViaFallback == nil || fallbackKind == nil || fallbackIndex == nil || fallbackStrength == nil || fallbackBlockedBy == nil || fallbackAmbiguityClass == nil || fallbackManualRetryHint == nil || fallbackSpecificityFields == nil {
		return false
	}
	hasExisting := *resolvedViaFallback ||
		strings.TrimSpace(*fallbackKind) != "" ||
		*fallbackIndex != nil ||
		strings.TrimSpace(*fallbackStrength) != "" ||
		strings.TrimSpace(*fallbackBlockedBy) != "" ||
		strings.TrimSpace(*fallbackAmbiguityClass) != "" ||
		strings.TrimSpace(*fallbackManualRetryHint) != "" ||
		len(*fallbackSpecificityFields) > 0
	if !hasExisting {
		return false
	}
	*resolvedViaFallback = true
	*fallbackKind = strings.TrimSpace(*fallbackKind)
	*fallbackIndex = browserCloneOptionalInt(*fallbackIndex)
	*fallbackStrength = strings.TrimSpace(*fallbackStrength)
	*fallbackBlockedBy = strings.TrimSpace(*fallbackBlockedBy)
	*fallbackAmbiguityClass = strings.TrimSpace(*fallbackAmbiguityClass)
	*fallbackManualRetryHint = strings.TrimSpace(*fallbackManualRetryHint)
	if len(*fallbackSpecificityFields) > 0 {
		*fallbackSpecificityFields = append([]string(nil), (*fallbackSpecificityFields)...)
	} else {
		*fallbackSpecificityFields = nil
	}
	return true
}

func browserActResultWithResolverFallbackSummary(result BrowserActResult) BrowserActResult {
	summary := browserResolverFallbackSummaryForOutcome(result.ResolverOutcome)
	result.ResolvedViaFallback = summary.ResolvedViaFallback
	result.ResolverFallbackKind = summary.Kind
	result.ResolverFallbackIndex = browserCloneOptionalInt(summary.Index)
	result.ResolverFallbackCandidateStrength = summary.CandidateStrength
	result.ResolverFallbackBlockedBy = summary.BlockedBy
	result.ResolverFallbackAmbiguityClass = summary.AmbiguityClass
	result.ResolverFallbackManualRetryHint = summary.ManualRetryHint
	if len(summary.SpecificityFields) > 0 {
		result.ResolverFallbackSpecificityFields = append([]string(nil), summary.SpecificityFields...)
	} else {
		result.ResolverFallbackSpecificityFields = nil
	}
	guidance := browserResolverGuidanceSummaryForOutcome(result.ResolverOutcome)
	result.ResolverBlockedBy = guidance.BlockedBy
	result.ResolverAmbiguityClass = guidance.AmbiguityClass
	result.ResolverCandidateKind = guidance.CandidateKind
	result.ResolverCandidateStrength = guidance.CandidateStrength
	result.ResolverRetryDisposition = guidance.RetryDisposition
	result.ResolverManualRetryHint = guidance.ManualRetryHint
	result.ResolverNextStepAlias = guidance.NextStepAlias
	return browserActResultWithActionabilityEvidence(result)
}

func browserCloneOptionalInt(value *int) *int {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func browserResolverGuidanceSummaryForOutcome(outcome *agentxbrowserruntime.BrowserElementResolverOutcome) browserResolverGuidanceSummary {
	summary := browserResolverGuidanceSummary{}
	normalized := outcome.Normalized()
	if normalized == nil {
		return summary
	}
	summary.BlockedBy = strings.TrimSpace(normalized.BlockedBy)
	summary.AmbiguityClass = strings.TrimSpace(normalized.AmbiguityClass)
	summary.CandidateKind = strings.TrimSpace(normalized.CandidateKind)
	summary.CandidateStrength = strings.TrimSpace(normalized.CandidateStrength)
	summary.RetryDisposition = strings.TrimSpace(normalized.RetryDisposition)
	summary.ManualRetryHint = strings.TrimSpace(normalized.ManualRetryHint)
	summary.NextStepAlias = strings.TrimSpace(normalized.NextStepAlias)
	return summary
}

func browserApplyResolverGuidanceSummary(
	outcome *agentxbrowserruntime.BrowserElementResolverOutcome,
	blockedBy *string,
	ambiguityClass *string,
	candidateKind *string,
	candidateStrength *string,
	retryDisposition *string,
	manualRetryHint *string,
	nextStepAlias *string,
) {
	if blockedBy == nil || ambiguityClass == nil || candidateKind == nil || candidateStrength == nil || retryDisposition == nil || manualRetryHint == nil || nextStepAlias == nil {
		return
	}
	if browserNormalizeExistingResolverGuidanceSummary(blockedBy, ambiguityClass, candidateKind, candidateStrength, retryDisposition, manualRetryHint, nextStepAlias) {
		return
	}
	summary := browserResolverGuidanceSummaryForOutcome(outcome)
	*blockedBy = summary.BlockedBy
	*ambiguityClass = summary.AmbiguityClass
	*candidateKind = summary.CandidateKind
	*candidateStrength = summary.CandidateStrength
	*retryDisposition = summary.RetryDisposition
	*manualRetryHint = summary.ManualRetryHint
	*nextStepAlias = summary.NextStepAlias
}

func browserNormalizeExistingResolverGuidanceSummary(
	blockedBy *string,
	ambiguityClass *string,
	candidateKind *string,
	candidateStrength *string,
	retryDisposition *string,
	manualRetryHint *string,
	nextStepAlias *string,
) bool {
	if blockedBy == nil || ambiguityClass == nil || candidateKind == nil || candidateStrength == nil || retryDisposition == nil || manualRetryHint == nil || nextStepAlias == nil {
		return false
	}
	hasExisting := strings.TrimSpace(*blockedBy) != "" ||
		strings.TrimSpace(*ambiguityClass) != "" ||
		strings.TrimSpace(*candidateKind) != "" ||
		strings.TrimSpace(*candidateStrength) != "" ||
		strings.TrimSpace(*retryDisposition) != "" ||
		strings.TrimSpace(*manualRetryHint) != "" ||
		strings.TrimSpace(*nextStepAlias) != ""
	if !hasExisting {
		return false
	}
	*blockedBy = strings.TrimSpace(*blockedBy)
	*ambiguityClass = strings.TrimSpace(*ambiguityClass)
	*candidateKind = strings.TrimSpace(*candidateKind)
	*candidateStrength = strings.TrimSpace(*candidateStrength)
	*retryDisposition = strings.TrimSpace(*retryDisposition)
	*manualRetryHint = strings.TrimSpace(*manualRetryHint)
	*nextStepAlias = strings.TrimSpace(*nextStepAlias)
	return true
}
