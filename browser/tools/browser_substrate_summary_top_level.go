package tools

import "strings"

func browserWorkbenchApplyTopLevelSubstrateSummary(summary BrowserWorkbenchSubstrateSummary) BrowserWorkbenchSubstrateSummary {
	posture, status, reason := browserWorkbenchTopLevelSubstrateSummary(summary)
	summary.SubstratePosture = strings.TrimSpace(posture)
	summary.SubstrateStatus = strings.TrimSpace(status)
	summary.SubstrateReason = strings.TrimSpace(reason)
	return summary
}

func browserWorkbenchTopLevelSubstrateSummary(summary BrowserWorkbenchSubstrateSummary) (string, string, string) {
	if browserTopLevelShouldPreferDefaultCandidateRoute(
		summary.DefaultRoute.Backend,
		summary.DefaultRoute.Target,
		summary.DefaultCandidateRoute.Backend,
		summary.DefaultCandidateRoute.Target,
		summary.SelectionStrategy,
	) {
		candidate := normalizeBrowserRuntimeInfo(summary.DefaultCandidateRoute)
		if posture, status, reason, ok := browserRuntimeTopLevelSubstrateValues(
			candidate.Backend,
			candidate.Target,
			"",
			"",
			"",
			summary.SelectionReason,
			"",
		); ok {
			return posture, status, reason
		}
	}
	if posture, status, reason, ok := browserRuntimeTopLevelSubstrateValues(
		summary.DefaultRoute.Backend,
		summary.DefaultRoute.Target,
		"",
		"",
		"",
		"",
		"",
	); ok {
		return posture, status, reason
	}
	if candidate := normalizeBrowserRuntimeInfo(summary.DefaultCandidateRoute); candidate != (BrowserRuntimeInfo{}) {
		if posture, status, reason, ok := browserRuntimeTopLevelSubstrateValues(
			candidate.Backend,
			candidate.Target,
			"",
			"",
			"",
			summary.SelectionReason,
			"",
		); ok {
			return posture, status, reason
		}
	}
	for _, target := range browserRuntimeTopLevelSubstratePreferredTargets(summary.SelectionStrategy) {
		if posture, status, reason, ok := browserWorkbenchTopLevelSubstrateTargetSummary(summary, target); ok {
			return posture, status, reason
		}
	}
	for _, target := range []string{"host", "node", "sandbox"} {
		if posture, status, reason, ok := browserWorkbenchTopLevelSubstrateTargetSummary(summary, target); ok {
			return posture, status, reason
		}
	}
	return "", "", ""
}

func browserWorkbenchTopLevelSubstrateTargetSummary(summary BrowserWorkbenchSubstrateSummary, runtimeTarget string) (string, string, string, bool) {
	switch strings.ToLower(strings.TrimSpace(runtimeTarget)) {
	case "host":
		info := normalizeBrowserRuntimeInfo(summary.HostRoute)
		if info == (BrowserRuntimeInfo{}) {
			info = defaultBrowserRuntimeInfo()
		}
		return browserRuntimeTopLevelSubstrateValues(
			info.Backend,
			info.Target,
			"",
			"",
			"",
			summary.SelectionReason,
			summary.HostFailureCause,
		)
	case "node":
		if normalizeBrowserRuntimeInfo(summary.DefaultRoute).Target != "node" &&
			!summary.NodeConfigured &&
			!summary.NodeRouteAvailable &&
			!summary.NodePromotionReady &&
			strings.TrimSpace(summary.NodePromotionFailureCause) == "" {
			return "", "", "", false
		}
		return browserRuntimeTopLevelSubstrateValues(
			"",
			"node",
			"",
			"",
			"",
			summary.SelectionReason,
			summary.NodePromotionFailureCause,
		)
	case "sandbox":
		if normalizeBrowserRuntimeInfo(summary.DefaultRoute).Target != "sandbox" &&
			!summary.SandboxConfigured &&
			!summary.SandboxRouteAvailable &&
			!summary.SandboxPromotionReady &&
			strings.TrimSpace(summary.SandboxPromotionFailureCause) == "" &&
			strings.TrimSpace(summary.SandboxFailureCause) == "" {
			return "", "", "", false
		}
		return browserRuntimeTopLevelSubstrateValues(
			"",
			"sandbox",
			"",
			"",
			"",
			summary.SelectionReason,
			firstNonEmpty(summary.SandboxFailureCause, summary.SandboxPromotionFailureCause),
		)
	default:
		return "", "", "", false
	}
}
