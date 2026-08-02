package tools

import "strings"

func browserRuntimeApplyTopLevelSubstrateSummary(payload *browserRuntimePayload) {
	if payload == nil {
		return
	}
	posture, status, reason := browserRuntimeTopLevelSubstrateSummary(*payload)
	payload.SubstratePosture = strings.TrimSpace(posture)
	payload.SubstrateStatus = strings.TrimSpace(status)
	payload.SubstrateReason = strings.TrimSpace(reason)
}

func browserRuntimeTopLevelSubstrateSummary(payload browserRuntimePayload) (string, string, string) {
	if route := payload.SelectedRoute; route != nil {
		if posture, status, reason, ok := browserRuntimeTopLevelSubstrateValues(
			route.Backend,
			route.RuntimeTarget,
			"",
			"",
			"",
			"",
			"",
		); ok {
			return posture, status, reason
		}
	}
	if browserTopLevelShouldPreferDefaultCandidateRoute(
		payload.DefaultRoute.Backend,
		payload.DefaultRoute.RuntimeTarget,
		payload.DefaultCandidateRoute.Backend,
		payload.DefaultCandidateRoute.RuntimeTarget,
		payload.SubstrateSelectionStrategy,
	) {
		if posture, status, reason, ok := browserRuntimeTopLevelSubstrateValues(
			payload.DefaultCandidateRoute.Backend,
			payload.DefaultCandidateRoute.RuntimeTarget,
			"",
			"",
			"",
			payload.SubstrateSelectionReason,
			"",
		); ok {
			return posture, status, reason
		}
	}
	if posture, status, reason, ok := browserRuntimeTopLevelSubstrateValues(
		payload.DefaultRoute.Backend,
		payload.DefaultRoute.RuntimeTarget,
		"",
		"",
		"",
		browserRuntimeTopLevelSubstrateSelectionReason(payload),
		"",
	); ok {
		return posture, status, reason
	}
	if posture, status, reason, ok := browserRuntimeTopLevelSubstrateValues(
		payload.DefaultCandidateRoute.Backend,
		payload.DefaultCandidateRoute.RuntimeTarget,
		"",
		"",
		"",
		payload.SubstrateSelectionReason,
		"",
	); ok {
		return posture, status, reason
	}
	row, ok := browserRuntimeTopLevelSubstrateRow(payload)
	if !ok {
		return "", "", ""
	}
	posture, status, reason, ok := browserRuntimeTopLevelSubstrateValues(
		row.Backend,
		row.RuntimeTarget,
		row.SubstratePosture,
		row.SubstrateStatus,
		row.SubstrateReason,
		row.SelectionReason,
		row.Note,
	)
	if !ok {
		return "", "", ""
	}
	return posture, status, reason
}

func browserTopLevelShouldPreferDefaultCandidateRoute(
	defaultBackend string,
	defaultTarget string,
	candidateBackend string,
	candidateTarget string,
	selectionStrategy string,
) bool {
	defaultPosture := BrowserSubstratePosture(defaultBackend, defaultTarget)
	candidatePosture := BrowserSubstratePosture(candidateBackend, candidateTarget)
	if defaultPosture != BrowserSubstrateLegacySystemHost || candidatePosture == "" || candidatePosture == defaultPosture {
		return false
	}
	for _, target := range browserRuntimeTopLevelSubstratePreferredTargets(selectionStrategy) {
		if strings.EqualFold(strings.TrimSpace(target), strings.TrimSpace(candidateTarget)) {
			return true
		}
	}
	return false
}

func browserRuntimeTopLevelSubstrateSelectionReason(payload browserRuntimePayload) string {
	if payload.SelectedRoute != nil {
		return ""
	}
	if payload.DefaultCandidateRoute == (browserRuntimeRouteDescriptor{}) {
		return ""
	}
	return strings.TrimSpace(payload.SubstrateSelectionReason)
}

func browserRuntimeTopLevelSubstrateValues(
	backend string,
	target string,
	posture string,
	status string,
	reason string,
	selectionReason string,
	note string,
) (string, string, string, bool) {
	posture = firstNonEmpty(strings.TrimSpace(posture), BrowserSubstratePosture(backend, target))
	if posture == "" {
		return "", "", "", false
	}
	status = firstNonEmpty(strings.TrimSpace(status), BrowserSubstrateStatus(backend, target))
	reason = firstNonEmpty(
		strings.TrimSpace(reason),
		strings.TrimSpace(selectionReason),
		strings.TrimSpace(note),
		BrowserSubstrateReason(backend, target),
	)
	return posture, status, reason, true
}

func browserRuntimeTopLevelSubstrateRow(payload browserRuntimePayload) (browserRuntimeSubstrateStatus, bool) {
	rows := payload.SubstrateMatrix
	if len(rows) == 0 {
		return browserRuntimeSubstrateStatus{}, false
	}
	preferredTargets := browserRuntimeTopLevelSubstratePreferredTargets(payload.SubstrateSelectionStrategy)
	for _, target := range preferredTargets {
		if row, ok := browserRuntimeFindTopLevelSubstrateRow(rows, target, "default"); ok {
			return row, true
		}
	}
	for _, target := range preferredTargets {
		if row, ok := browserRuntimeFindTopLevelSubstrateRow(rows, target); ok {
			return row, true
		}
	}
	if row, ok := browserRuntimeFindTopLevelSubstrateRow(rows, "", "default"); ok {
		return row, true
	}
	if row, ok := browserRuntimeFindTopLevelSubstrateRow(rows, "host"); ok {
		return row, true
	}
	if row, ok := browserRuntimeFindTopLevelSubstrateRow(rows, "node"); ok {
		return row, true
	}
	if row, ok := browserRuntimeFindTopLevelSubstrateRow(rows, "sandbox"); ok {
		return row, true
	}
	for _, row := range rows {
		if strings.TrimSpace(row.Backend) != "" || strings.TrimSpace(row.RuntimeTarget) != "" {
			return row, true
		}
	}
	return browserRuntimeSubstrateStatus{}, false
}

func browserRuntimeTopLevelSubstratePreferredTargets(selectionStrategy string) []string {
	switch strings.TrimSpace(selectionStrategy) {
	case BrowserSubstrateSelectionPreferNodeOverLegacy, BrowserSubstrateSelectionPreferNodeRuntime:
		return []string{"node"}
	case BrowserSubstrateSelectionPreferSandboxRuntime:
		return []string{"sandbox"}
	case BrowserSubstrateSelectionPreferHostRuntime, BrowserSubstrateSelectionPreferCustomBackend, BrowserSubstrateSelectionLegacyHostDefault:
		return []string{"host"}
	default:
		return nil
	}
}

func browserRuntimeFindTopLevelSubstrateRow(rows []browserRuntimeSubstrateStatus, runtimeTarget string, roles ...string) (browserRuntimeSubstrateStatus, bool) {
	roleSet := map[string]bool{}
	for _, role := range roles {
		if trimmed := strings.TrimSpace(role); trimmed != "" {
			roleSet[trimmed] = true
		}
	}
	for _, row := range rows {
		if len(roleSet) != 0 && !roleSet[strings.TrimSpace(row.Role)] {
			continue
		}
		if strings.TrimSpace(runtimeTarget) != "" && !strings.EqualFold(strings.TrimSpace(row.RuntimeTarget), strings.TrimSpace(runtimeTarget)) {
			continue
		}
		if strings.TrimSpace(row.Backend) == "" && strings.TrimSpace(row.RuntimeTarget) == "" {
			continue
		}
		return row, true
	}
	return browserRuntimeSubstrateStatus{}, false
}
