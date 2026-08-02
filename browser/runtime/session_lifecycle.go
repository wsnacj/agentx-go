package browserruntime

import "strings"

func SharedSessionBrowserLifecycleDecisionReady(selectedInfo BrowserRuntimeInfo, profile string, result BrowserProfileStatusResult, decision string) bool {
	result = SharedSessionBrowserLifecycleDecisionStatus(selectedInfo, profile, result, decision)
	switch strings.ToLower(strings.TrimSpace(decision)) {
	case "stop_already_stopped", "stopped", "teardown_stopped", "teardown_already_stopped":
		return true
	}
	switch strings.ToLower(strings.TrimSpace(result.Status)) {
	case "starting", "reconnecting":
		return false
	}
	return SharedSessionBrowserProfileReady(result)
}

func SharedSessionBrowserLifecycleDecisionStatus(selectedInfo BrowserRuntimeInfo, profile string, result BrowserProfileStatusResult, decision string) BrowserProfileStatusResult {
	if state, ok := SharedSessionBrowserProfileStateFromLifecycle(selectedInfo, profile, result, decision); ok {
		return SharedSessionBrowserProfileStatusResultFromState(state, selectedInfo, profile)
	}
	return result
}

func SharedSessionBrowserEnsurePreparedInProgressStatus(selectedInfo BrowserRuntimeInfo, profile string, result BrowserProfileStatusResult) (BrowserProfileStatusResult, string, bool) {
	switch strings.ToLower(strings.TrimSpace(result.Status)) {
	case "starting":
		return SharedSessionBrowserLifecycleDecisionStatus(selectedInfo, profile, result, "started"), "started", true
	case "started":
		if !result.Connected {
			return SharedSessionBrowserLifecycleDecisionStatus(selectedInfo, profile, result, "started"), "started", true
		}
	case "reconnecting":
		return SharedSessionBrowserLifecycleDecisionStatus(selectedInfo, profile, result, "restart_reconnect_in_progress"), "restart_reconnect_in_progress", true
	}
	return BrowserProfileStatusResult{}, "", false
}

func SharedSessionBrowserProfileStateReady(state SharedSessionBrowserProfileState) bool {
	status := strings.ToLower(strings.TrimSpace(state.Status))
	if state.Running || state.Connected {
		return true
	}
	return status == "running" || status == "started" || status == "ready" || status == "connected"
}

func SharedSessionBrowserProfileReady(result BrowserProfileStatusResult) bool {
	status := strings.ToLower(strings.TrimSpace(result.Status))
	if result.Running || result.Connected {
		return true
	}
	return status == "running" || status == "started" || status == "ready" || status == "connected"
}
