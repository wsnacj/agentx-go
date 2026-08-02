package browserruntime

import (
	"strings"
	"testing"
	"time"
)

func TestEvaluateSharedSessionBrowserHealthReconnectTimeout(t *testing.T) {
	staleAt := time.Now().Add(-2 * time.Minute)
	evaluation := EvaluateSharedSessionBrowserHealth(SharedSessionBrowserHealthInput{
		Profiles: []SharedSessionBrowserProfileState{{
			Backend:       "proxy",
			Profile:       "isolated",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Status:        "reconnecting",
			Running:       true,
			Connected:     false,
			ObservedAt:    staleAt,
			StatusSince:   staleAt,
			Note:          "cdp reconnect in progress",
		}},
	}, 54*time.Second)
	if evaluation.Summary == nil || evaluation.Summary.State != "profile_reconnecting" {
		t.Fatalf("expected reconnecting health evaluation, got %#v", evaluation)
	}
	if !evaluation.ReconnectTimedOut || evaluation.Summary.RecoveryAction != "browser action=refresh" {
		t.Fatalf("expected reconnect timeout to escalate to refresh, got %#v", evaluation)
	}
}

func TestEvaluateSharedSessionBrowserHealthReconnectTimeoutUsesReferenceTime(t *testing.T) {
	base := time.Date(2026, time.March, 29, 12, 0, 0, 0, time.UTC)
	staleAt := base.Add(-2 * time.Minute)
	evaluation := EvaluateSharedSessionBrowserHealth(SharedSessionBrowserHealthInput{
		ReferenceTime: staleAt.Add(12 * time.Second),
		Profiles: []SharedSessionBrowserProfileState{{
			Backend:       "proxy",
			Profile:       "isolated",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Status:        "reconnecting",
			Running:       true,
			Connected:     false,
			ObservedAt:    staleAt,
			StatusSince:   staleAt,
			Note:          "cdp reconnect in progress",
		}},
	}, 54*time.Second)
	if evaluation.Summary == nil || evaluation.Summary.State != "profile_reconnecting" {
		t.Fatalf("expected reconnecting health evaluation, got %#v", evaluation)
	}
	if evaluation.ReconnectTimedOut || evaluation.Summary.RecoveryAction != "" {
		t.Fatalf("expected source-time reference to keep reconnect inside watchdog window, got %#v", evaluation)
	}
	if !strings.Contains(evaluation.Summary.Reason, "12s elapsed") {
		t.Fatalf("expected source-time reconnect reason, got %#v", evaluation.Summary)
	}
}

func TestEvaluateSharedSessionBrowserHealthPrefersStoredSummary(t *testing.T) {
	reconnectingAt := time.Now().Add(-12 * time.Second)
	evaluation := EvaluateSharedSessionBrowserHealth(SharedSessionBrowserHealthInput{
		StoredState:          "profile_reconnecting",
		StoredReason:         "stored reconnecting posture",
		StoredRecoveryAction: "",
		Profiles: []SharedSessionBrowserProfileState{{
			Backend:       "proxy",
			Profile:       "isolated",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Status:        "reconnecting",
			Running:       true,
			Connected:     false,
			ObservedAt:    reconnectingAt,
			StatusSince:   reconnectingAt,
			Note:          "cdp reconnect in progress",
		}},
	}, 54*time.Second)
	if evaluation.Summary == nil || evaluation.Summary.State != "profile_reconnecting" || evaluation.Summary.Reason != "stored reconnecting posture" {
		t.Fatalf("expected stored summary to win when scoped profile matches, got %#v", evaluation)
	}
	if !evaluation.HasProfile || evaluation.Profile.Profile != "isolated" || evaluation.Profile.Status != "reconnecting" {
		t.Fatalf("expected reconnecting profile to remain attached to stored summary, got %#v", evaluation)
	}
}

func TestEvaluateSharedSessionBrowserHealthPreservesStoredReliabilityMetadata(t *testing.T) {
	evaluation := EvaluateSharedSessionBrowserHealth(SharedSessionBrowserHealthInput{
		StoredState:                       "cooldown_active",
		StoredReason:                      "browser restart cooldown active for 900ms after 2 disconnects",
		StoredRecoveryAction:              "browser action=wait",
		StoredReconnectHint:               "wait_for_restart",
		StoredDisconnectCount:             2,
		StoredDisconnectBurstCount:        1,
		StoredDisconnectBurstWindowMs:     30000,
		StoredCooldownRemainingMs:         900,
		StoredRetryBackoffRemainingMs:     450,
		StoredRestartAttemptCount:         4,
		StoredRestartFailureCount:         1,
		StoredLastDisconnectUnixMilli:     111,
		StoredLastReconnectUnixMilli:      222,
		StoredLastRestartAttemptUnixMilli: 333,
		StoredLastRestartResult:           "restarted",
		StoredLastRestartError:            "transport closed",
		StoredRecommendedBackoffMs:        1200,
	}, 54*time.Second)
	if evaluation.Summary == nil ||
		evaluation.Summary.State != "cooldown_active" ||
		evaluation.Summary.ReconnectHint != "wait_for_restart" ||
		evaluation.Summary.DisconnectCount != 2 ||
		evaluation.Summary.DisconnectBurstCount != 1 ||
		evaluation.Summary.DisconnectBurstWindowMs != 30000 ||
		evaluation.Summary.CooldownRemainingMs != 900 ||
		evaluation.Summary.RetryBackoffRemainingMs != 450 ||
		evaluation.Summary.RestartAttemptCount != 4 ||
		evaluation.Summary.RestartFailureCount != 1 ||
		evaluation.Summary.LastDisconnectUnixMilli != 111 ||
		evaluation.Summary.LastReconnectUnixMilli != 222 ||
		evaluation.Summary.LastRestartAttemptUnixMilli != 333 ||
		evaluation.Summary.LastRestartResult != "restarted" ||
		evaluation.Summary.LastRestartError != "transport closed" ||
		evaluation.Summary.RecommendedBackoffMs != 1200 {
		t.Fatalf("expected stored reliability metadata to survive shared health evaluation, got %#v", evaluation.Summary)
	}
}

func TestEvaluateSharedSessionBrowserHealthForScopePrefersRegistrySnapshot(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	reconnectingAt := time.Now().Add(-12 * time.Second)
	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "reconnecting",
		Running:       true,
		Connected:     false,
		ObservedAt:    reconnectingAt,
		StatusSince:   reconnectingAt,
		Note:          "cdp reconnect in progress",
	})
	evaluation := EvaluateSharedSessionBrowserHealthForScope(registry, "s1", BrowserRuntimeInfo{
		Backend: "proxy",
		Profile: "isolated",
		Target:  "node",
	}, "isolated", SharedSessionBrowserHealthInput{
		StoredState:          "profile_stopped",
		StoredReason:         "stale stopped posture",
		StoredRecoveryAction: "browser action=ensure",
		Profiles: []SharedSessionBrowserProfileState{{
			Backend:       "proxy",
			Profile:       "isolated",
			RuntimeTarget: "node",
			Status:        "stopped",
		}},
	}, 54*time.Second)
	if evaluation.Summary == nil || evaluation.Summary.State != "profile_reconnecting" {
		t.Fatalf("expected registry-scoped snapshot to override stale stored summary, got %#v", evaluation)
	}
	if !evaluation.HasProfile || evaluation.Profile.Profile != "isolated" || evaluation.Profile.Status != "reconnecting" {
		t.Fatalf("expected reconnecting scoped profile from registry snapshot, got %#v", evaluation)
	}
}

func TestEvaluateSharedSessionBrowserHealthForScopePreservesStoredLifecycleBlocker(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "disconnected",
		Running:       false,
		Connected:     false,
		Note:          "cdp transport closed",
	})
	evaluation := EvaluateSharedSessionBrowserHealthForScope(registry, "s1", BrowserRuntimeInfo{
		Backend: "proxy",
		Profile: "isolated",
		Target:  "node",
	}, "isolated", SharedSessionBrowserHealthInput{
		StoredState:               "cooldown_active",
		StoredReason:              "browser restart cooldown active for 900ms after 2 disconnects",
		StoredRecoveryAction:      "browser action=wait",
		StoredReconnectHint:       "retry_after_cooldown",
		StoredCooldownRemainingMs: 900,
	}, 54*time.Second)
	if evaluation.Summary == nil || evaluation.Summary.State != "cooldown_active" || evaluation.Summary.CooldownRemainingMs != 900 {
		t.Fatalf("expected registry-scoped health to preserve lifecycle blocker summary, got %#v", evaluation)
	}
	if evaluation.HasProfile {
		t.Fatalf("expected lifecycle blocker summary to defer without binding a synthetic profile, got %#v", evaluation)
	}
}

func TestMergeSharedSessionBrowserBindingEvaluationHealthSummaryPreservesLifecycleBlockerForScope(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "disconnected",
		Running:       false,
		Connected:     false,
		Note:          "cdp transport closed",
	})
	evaluation := SharedSessionBrowserBindingEvaluation{
		Snapshot: SharedSessionBrowserBindingSnapshot{
			Profiles: []SharedSessionBrowserProfileState{{
				Backend:       "proxy",
				Profile:       "isolated",
				RuntimeTarget: "node",
				BrowserApp:    "Chromium",
				Status:        "disconnected",
				Running:       false,
				Connected:     false,
				Note:          "cdp transport closed",
			}},
			Summary: SharedSessionBrowserBindingSummary{
				BrowserProfileCount:        1,
				ActiveBrowserProfile:       "isolated",
				BrowserProfileStatusCounts: map[string]int{"disconnected": 1},
			},
		},
	}

	evaluation = MergeSharedSessionBrowserBindingEvaluationHealthSummary(
		registry,
		"s1",
		BrowserRuntimeInfo{
			Backend: "proxy",
			Profile: "isolated",
			Target:  "node",
		},
		"isolated",
		evaluation,
		&SharedSessionBrowserHealthSummary{
			State:               "cooldown_active",
			Reason:              "browser restart cooldown active for 900ms after 2 disconnects",
			RecoveryAction:      "browser action=wait",
			ReconnectHint:       "retry_after_cooldown",
			CooldownRemainingMs: 900,
		},
		54*time.Second,
	)
	if evaluation.Health.Summary == nil ||
		evaluation.Health.Summary.State != "cooldown_active" ||
		evaluation.Health.Summary.ReconnectHint != "retry_after_cooldown" ||
		evaluation.Health.Summary.CooldownRemainingMs != 900 {
		t.Fatalf("expected merged health blocker summary to survive scoped binding evaluation refresh, got %#v", evaluation.Health.Summary)
	}
	if decision, blocked := SharedSessionBrowserExecutionBlockedDecision(evaluation.Health); !blocked || decision != "cooldown_active" {
		t.Fatalf("expected merged health blocker to remain lifecycle-blocking, got decision=%q blocked=%v summary=%#v", decision, blocked, evaluation.Health.Summary)
	}
}

func TestEvaluateSharedSessionBrowserHealthForScopeFiltersRequestedProfile(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	reconnectingAt := time.Now().Add(-12 * time.Second)
	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "reconnecting",
		Running:       true,
		Connected:     false,
		ObservedAt:    reconnectingAt,
		StatusSince:   reconnectingAt,
		Note:          "cdp reconnect in progress",
	})
	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "work",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "connected",
		Running:       true,
		Connected:     true,
	})
	evaluation := EvaluateSharedSessionBrowserHealthForScope(registry, "s1", BrowserRuntimeInfo{
		Backend: "proxy",
		Profile: "work",
		Target:  "node",
	}, "work", SharedSessionBrowserHealthInput{}, 54*time.Second)
	if evaluation.Summary == nil || evaluation.Summary.State != "healthy" {
		t.Fatalf("expected requested profile scope to ignore unrelated reconnecting profile, got %#v", evaluation)
	}
	if !evaluation.HasProfile || evaluation.Profile.Profile != "work" || evaluation.Profile.Status != "connected" {
		t.Fatalf("expected scoped healthy profile from requested profile, got %#v", evaluation)
	}
}

func TestEvaluateSharedSessionBrowserHealthForInputScopeFiltersRequestedProfile(t *testing.T) {
	reconnectingAt := time.Now().Add(-12 * time.Second)
	evaluation := EvaluateSharedSessionBrowserHealthForInputScope(SharedSessionBrowserHealthInput{
		Profiles: []SharedSessionBrowserProfileState{
			{
				Backend:       "proxy",
				Profile:       "isolated",
				RuntimeTarget: "node",
				BrowserApp:    "Chromium",
				Status:        "reconnecting",
				Running:       true,
				Connected:     false,
				ObservedAt:    reconnectingAt,
				StatusSince:   reconnectingAt,
				Note:          "cdp reconnect in progress",
			},
			{
				Backend:       "proxy",
				Profile:       "work",
				RuntimeTarget: "node",
				BrowserApp:    "Chromium",
				Status:        "connected",
				Running:       true,
				Connected:     true,
			},
		},
	}, BrowserRuntimeInfo{
		Backend: "proxy",
		Target:  "node",
	}, "work", 54*time.Second)
	if evaluation.Summary == nil || evaluation.Summary.State != "healthy" {
		t.Fatalf("expected input-scoped health evaluation to ignore unrelated reconnecting profile, got %#v", evaluation)
	}
	if !evaluation.HasProfile || evaluation.Profile.Profile != "work" || evaluation.Profile.Status != "connected" {
		t.Fatalf("expected connected requested profile after input scoping, got %#v", evaluation)
	}
}

func TestEvaluateSharedSessionBrowserCoordinationEvaluationForScopePrefersRegistrySnapshot(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	reconnectingAt := time.Now().Add(-12 * time.Second)
	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "reconnecting",
		Running:       true,
		Connected:     false,
		ObservedAt:    reconnectingAt,
		StatusSince:   reconnectingAt,
		Note:          "cdp reconnect in progress",
	})
	evaluation := EvaluateSharedSessionBrowserCoordinationEvaluationForScope(
		registry,
		"s1",
		BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		"isolated",
		SharedSessionBrowserHealthInput{
			StoredState:          "profile_stopped",
			StoredReason:         "stale stopped posture",
			StoredRecoveryAction: "browser action=ensure",
			Profiles: []SharedSessionBrowserProfileState{{
				Backend:       "proxy",
				Profile:       "isolated",
				RuntimeTarget: "node",
				Status:        "stopped",
			}},
		},
		SharedSessionBrowserCoordinationInput{
			Profiles: []SharedSessionBrowserProfileState{{
				Backend:       "proxy",
				Profile:       "isolated",
				RuntimeTarget: "node",
				Status:        "stopped",
			}},
		},
		nil,
		54*time.Second,
		false,
	)
	if evaluation.Plan.State != "browser_ready" {
		t.Fatalf("expected registry-scoped coordination to see running reconnecting profile, got %#v", evaluation)
	}
	if evaluation.RestartAction != "" {
		t.Fatalf("expected reconnect-in-progress guardrail to suppress refresh hint, got %#v", evaluation)
	}
	if sharedSessionBrowserActionSliceContains(evaluation.Guidance.RecommendedActions, "browser action=refresh") {
		t.Fatalf("expected scoped coordination to suppress refresh recommendation while reconnect is in progress, got %#v", evaluation)
	}
}

func TestResolveSharedSessionBrowserProfileStatusForScopePrefersRegistrySnapshot(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	reconnectingAt := time.Now().Add(-12 * time.Second)
	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "reconnecting",
		Running:       true,
		Connected:     false,
		ObservedAt:    reconnectingAt,
		StatusSince:   reconnectingAt,
		Note:          "cdp reconnect in progress",
	})
	status := ResolveSharedSessionBrowserProfileStatusForScope(
		registry,
		"s1",
		BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		"isolated",
		SharedSessionBrowserHealthInput{
			StoredState:          "profile_stopped",
			StoredReason:         "stale stopped posture",
			StoredRecoveryAction: "browser action=ensure",
			Profiles: []SharedSessionBrowserProfileState{{
				Backend:       "proxy",
				Profile:       "isolated",
				RuntimeTarget: "node",
				Status:        "stopped",
			}},
		},
		BrowserProfileStatusResult{},
		54*time.Second,
	)
	if status.Profile != "isolated" || status.Status != "reconnecting" || !status.Running || status.Connected {
		t.Fatalf("expected registry-scoped lifecycle status to override stale stored state, got %#v", status)
	}
}

func TestResolveSharedSessionBrowserProfileStatusForScopePreservesExplicitFallback(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		Status:        "reconnecting",
		Running:       true,
		Connected:     false,
	})
	fallback := BrowserProfileStatusResult{
		Backend:   "proxy",
		Profile:   "isolated",
		Status:    "connected",
		Running:   true,
		Connected: true,
		Note:      "fresh cdp status",
	}
	status := ResolveSharedSessionBrowserProfileStatusForScope(
		registry,
		"s1",
		BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		"isolated",
		SharedSessionBrowserHealthInput{
			StoredState:          "profile_reconnecting",
			StoredReason:         "stale registry summary",
			StoredRecoveryAction: "browser action=wait",
		},
		fallback,
		54*time.Second,
	)
	if status != fallback {
		t.Fatalf("expected explicit lifecycle fallback to be preserved, got %#v", status)
	}
}

func TestAssessSharedSessionBrowserProfileRecoveryForScopePrefersRegistrySnapshot(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	reconnectingAt := time.Now().Add(-12 * time.Second)
	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "reconnecting",
		Running:       true,
		Connected:     false,
		ObservedAt:    reconnectingAt,
		StatusSince:   reconnectingAt,
		Note:          "cdp reconnect in progress",
	})
	evaluation, assessment := AssessSharedSessionBrowserProfileRecoveryForScope(
		registry,
		"s1",
		BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		"isolated",
		SharedSessionBrowserHealthInput{
			StoredState:          "profile_stopped",
			StoredReason:         "stale stopped posture",
			StoredRecoveryAction: "browser action=ensure",
			Profiles: []SharedSessionBrowserProfileState{{
				Backend:       "proxy",
				Profile:       "isolated",
				RuntimeTarget: "node",
				Status:        "stopped",
			}},
		},
		BrowserProfileStatusResult{},
		54*time.Second,
	)
	if evaluation.Summary == nil || evaluation.Summary.State != "profile_reconnecting" {
		t.Fatalf("expected registry-scoped recovery to reuse reconnecting snapshot, got %#v", evaluation)
	}
	if !assessment.ReconnectInProgress || !assessment.HasSyntheticStatus {
		t.Fatalf("expected reconnect-in-progress recovery assessment, got %#v", assessment)
	}
	if assessment.SyntheticStatus.Profile != "isolated" || assessment.SyntheticStatus.Status != "reconnecting" {
		t.Fatalf("expected reconnecting synthetic status from scoped snapshot, got %#v", assessment)
	}
}

func TestAssessSharedSessionBrowserProfileRecoveryForScopeUsesFallbackSessionHealthBlocker(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "disconnected",
		Running:       false,
		Connected:     false,
		Note:          "cdp transport closed",
	})
	evaluation, assessment := AssessSharedSessionBrowserProfileRecoveryForScope(
		registry,
		"s1",
		BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		"isolated",
		SharedSessionBrowserHealthInput{},
		BrowserProfileStatusResult{
			Backend:    "proxy",
			BrowserApp: "Chromium",
			Profile:    "isolated",
			Status:     "disconnected",
			Running:    false,
			Connected:  false,
			Note:       "cdp transport closed",
			SessionHealth: &BrowserSessionHealthSummary{
				State:               "cooldown_active",
				Reason:              "browser restart cooldown active for 900ms after 2 disconnects",
				RecoveryAction:      "browser action=wait",
				ReconnectHint:       "retry_after_cooldown",
				CooldownRemainingMs: 900,
			},
		},
		54*time.Second,
	)
	if evaluation.Summary == nil || evaluation.Summary.State != "cooldown_active" || evaluation.Summary.CooldownRemainingMs != 900 {
		t.Fatalf("expected fallback session health blocker to survive scoped recovery, got %#v", evaluation)
	}
	if assessment.NeedsRefreshRecovery || assessment.ShouldStopBeforeRecovery {
		t.Fatalf("expected fallback session health blocker to suppress lifecycle churn, got %#v", assessment)
	}
}

func TestRecommendSharedSessionBrowserHealthActionsReconnectSuppressesRefresh(t *testing.T) {
	actions := RecommendSharedSessionBrowserHealthActions(false, SharedSessionBrowserHealthEvaluation{
		Summary: &SharedSessionBrowserHealthSummary{
			State:          "profile_reconnecting",
			Reason:         "cdp reconnect in progress",
			RecoveryAction: "",
		},
		ReconnectTimedOut: false,
	})
	if actions.PrimaryAction != "" || len(actions.RecommendedActions) != 0 {
		t.Fatalf("expected reconnecting health to avoid proactive browser actions inside watchdog window, got %#v", actions)
	}
	if !actions.SuppressRefresh {
		t.Fatalf("expected reconnecting health to suppress refresh inside watchdog window, got %#v", actions)
	}
}

func TestRecommendSharedSessionBrowserHealthActionsDisconnectedHintsRefreshAndEnsure(t *testing.T) {
	actions := RecommendSharedSessionBrowserHealthActions(false, SharedSessionBrowserHealthEvaluation{
		Summary: &SharedSessionBrowserHealthSummary{
			State:          "profile_disconnected",
			Reason:         "cdp transport closed",
			RecoveryAction: "browser action=refresh",
		},
	})
	if actions.PrimaryAction != "browser action=refresh" {
		t.Fatalf("expected disconnected health to prefer refresh, got %#v", actions)
	}
	if len(actions.RecommendedActions) != 2 ||
		actions.RecommendedActions[0] != "browser action=refresh" ||
		actions.RecommendedActions[1] != "browser action=ensure" {
		t.Fatalf("expected disconnected health to recommend refresh+ensure, got %#v", actions)
	}
	if actions.SuppressRefresh {
		t.Fatalf("expected disconnected health not to suppress refresh, got %#v", actions)
	}
}

func TestRecommendSharedSessionBrowserHealthActionsStoppedHintsEnsureOnly(t *testing.T) {
	actions := RecommendSharedSessionBrowserHealthActions(false, SharedSessionBrowserHealthEvaluation{
		Summary: &SharedSessionBrowserHealthSummary{
			State:          "profile_stopped",
			Reason:         "managed browser profile is stopped",
			RecoveryAction: "browser action=ensure",
		},
	})
	if actions.PrimaryAction != "browser action=ensure" {
		t.Fatalf("expected stopped health to prefer ensure, got %#v", actions)
	}
	if len(actions.RecommendedActions) != 1 || actions.RecommendedActions[0] != "browser action=ensure" {
		t.Fatalf("expected stopped health to recommend ensure only, got %#v", actions)
	}
	if actions.SuppressRefresh {
		t.Fatalf("expected stopped health not to suppress refresh, got %#v", actions)
	}
}

func TestRecommendSharedSessionBrowserHealthActionsCooldownActivePrefersWaitAndSuppressesRestart(t *testing.T) {
	actions := RecommendSharedSessionBrowserHealthActions(false, SharedSessionBrowserHealthEvaluation{
		Summary: &SharedSessionBrowserHealthSummary{
			State:          "cooldown_active",
			Reason:         "browser restart cooldown active for 900ms after 2 disconnects",
			RecoveryAction: "browser action=wait",
		},
	})
	if actions.PrimaryAction != "browser action=wait" || len(actions.RecommendedActions) != 1 || actions.RecommendedActions[0] != "browser action=wait" {
		t.Fatalf("expected cooldown_active to prefer wait only, got %#v", actions)
	}
	if !actions.SuppressRefresh || !actions.ClearRestartAction || actions.RestartAction != "" {
		t.Fatalf("expected cooldown_active to suppress restart/refresh, got %#v", actions)
	}
}

func TestRecommendSharedSessionBrowserHealthActionsCooldownActiveFallsBackToWaitFromHint(t *testing.T) {
	actions := RecommendSharedSessionBrowserHealthActions(false, SharedSessionBrowserHealthEvaluation{
		Summary: &SharedSessionBrowserHealthSummary{
			State:               "cooldown_active",
			Reason:              "browser restart cooldown active for 900ms after 2 disconnects",
			ReconnectHint:       "retry_after_cooldown",
			CooldownRemainingMs: 900,
		},
	})
	if actions.PrimaryAction != "browser action=wait" || len(actions.RecommendedActions) != 1 || actions.RecommendedActions[0] != "browser action=wait" {
		t.Fatalf("expected cooldown_active hint fallback to prefer wait only, got %#v", actions)
	}
	if !actions.SuppressRefresh || !actions.ClearRestartAction || actions.RestartAction != "" {
		t.Fatalf("expected cooldown_active hint fallback to suppress restart/refresh, got %#v", actions)
	}
}

func TestRecommendSharedSessionBrowserHealthActionsRestartPendingFallsBackToWait(t *testing.T) {
	actions := RecommendSharedSessionBrowserHealthActions(false, SharedSessionBrowserHealthEvaluation{
		Summary: &SharedSessionBrowserHealthSummary{
			State:                   "restart_pending",
			Reason:                  "browser relaunch backoff active for 450ms after restart failure",
			ReconnectHint:           "retry_after_backoff",
			RetryBackoffRemainingMs: 450,
			RecommendedBackoffMs:    1200,
		},
	})
	if actions.PrimaryAction != "browser action=wait" || len(actions.RecommendedActions) != 1 || actions.RecommendedActions[0] != "browser action=wait" {
		t.Fatalf("expected restart_pending to prefer wait only, got %#v", actions)
	}
	if !actions.SuppressRefresh || !actions.ClearRestartAction || actions.RestartAction != "" {
		t.Fatalf("expected restart_pending to suppress restart/refresh, got %#v", actions)
	}
}

func TestRecommendSharedSessionBrowserHealthActionsRestartFailedPermanentRequiresStart(t *testing.T) {
	actions := RecommendSharedSessionBrowserHealthActions(false, SharedSessionBrowserHealthEvaluation{
		Summary: &SharedSessionBrowserHealthSummary{
			State:          "restart_failed_permanent",
			Reason:         "browser restart failed 2 times; explicit browser.start required",
			RecoveryAction: "browser action=start",
		},
	})
	if actions.PrimaryAction != "browser action=start" || actions.RestartAction != "browser action=start" {
		t.Fatalf("expected restart_failed_permanent to prefer browser.start, got %#v", actions)
	}
	if len(actions.RecommendedActions) != 2 ||
		actions.RecommendedActions[0] != "browser action=start" ||
		actions.RecommendedActions[1] != "browser action=ensure" {
		t.Fatalf("expected restart_failed_permanent to recommend start+ensure, got %#v", actions)
	}
	if actions.SuppressRefresh || actions.ClearRestartAction {
		t.Fatalf("expected restart_failed_permanent not to use refresh suppression path, got %#v", actions)
	}
}

func TestRecommendSharedSessionBrowserHealthActionsRestartFailedPermanentFallsBackToStartFromHint(t *testing.T) {
	actions := RecommendSharedSessionBrowserHealthActions(false, SharedSessionBrowserHealthEvaluation{
		Summary: &SharedSessionBrowserHealthSummary{
			State:         "restart_failed_permanent",
			Reason:        "browser restart failed 2 times; explicit browser.start required",
			ReconnectHint: "manual_restart_required",
		},
	})
	if actions.PrimaryAction != "browser action=start" || actions.RestartAction != "browser action=start" {
		t.Fatalf("expected restart_failed_permanent hint fallback to prefer browser.start, got %#v", actions)
	}
	if len(actions.RecommendedActions) != 2 ||
		actions.RecommendedActions[0] != "browser action=start" ||
		actions.RecommendedActions[1] != "browser action=ensure" {
		t.Fatalf("expected restart_failed_permanent hint fallback to recommend start+ensure, got %#v", actions)
	}
}

func TestApplySharedSessionBrowserHealthActionsReconnectSuppressesRefreshGuidance(t *testing.T) {
	guidance := ApplySharedSessionBrowserHealthActions(SharedSessionBrowserHealthGuidance{
		RestartAction:      "browser action=refresh",
		PrimaryAction:      "browser action=refresh",
		PrimaryNodeAction:  "nodes action=run",
		NextStep:           "browser action=refresh",
		RecommendedActions: []string{"browser action=sync", "browser action=refresh", "browser"},
	}, SharedSessionBrowserHealthActions{
		SuppressRefresh: true,
	})
	if guidance.RestartAction != "" {
		t.Fatalf("expected reconnect suppression to clear restart action, got %#v", guidance)
	}
	if guidance.PrimaryAction != "" {
		t.Fatalf("expected reconnect suppression to clear refresh primary action, got %#v", guidance)
	}
	if guidance.NextStep != "nodes action=run" {
		t.Fatalf("expected reconnect suppression to fall back to node next step, got %#v", guidance)
	}
	if len(guidance.RecommendedActions) != 2 ||
		guidance.RecommendedActions[0] != "browser action=sync" ||
		guidance.RecommendedActions[1] != "browser" {
		t.Fatalf("expected reconnect suppression to drop refresh from recommendations, got %#v", guidance)
	}
}

func TestApplySharedSessionBrowserHealthActionsCooldownClearsRestartGuidance(t *testing.T) {
	guidance := ApplySharedSessionBrowserHealthActions(SharedSessionBrowserHealthGuidance{
		RestartAction:      "browser action=refresh",
		PrimaryAction:      "browser action=sync",
		PrimaryNodeAction:  "nodes action=run",
		NextStep:           "browser action=sync",
		RecommendedActions: []string{"browser action=sync", "browser action=refresh", "browser"},
	}, RecommendSharedSessionBrowserHealthActions(false, SharedSessionBrowserHealthEvaluation{
		Summary: &SharedSessionBrowserHealthSummary{
			State:          "cooldown_active",
			Reason:         "browser restart cooldown active for 900ms after 2 disconnects",
			RecoveryAction: "browser action=wait",
		},
	}))
	if guidance.RestartAction != "" {
		t.Fatalf("expected cooldown_active to clear restart action, got %#v", guidance)
	}
	if guidance.PrimaryAction != "browser action=wait" || guidance.NextStep != "browser action=wait" {
		t.Fatalf("expected cooldown_active to force wait next step, got %#v", guidance)
	}
	if len(guidance.RecommendedActions) != 3 ||
		guidance.RecommendedActions[0] != "browser action=wait" ||
		guidance.RecommendedActions[1] != "browser action=sync" ||
		guidance.RecommendedActions[2] != "browser" {
		t.Fatalf("expected cooldown_active to prepend wait and drop refresh, got %#v", guidance)
	}
}

func TestApplySharedSessionBrowserHealthActionsDisconnectedOverridesGuidance(t *testing.T) {
	guidance := ApplySharedSessionBrowserHealthActions(SharedSessionBrowserHealthGuidance{
		RestartAction:      "browser action=refresh",
		PrimaryAction:      "browser",
		PrimaryNodeAction:  "nodes action=run",
		NextStep:           "browser",
		RecommendedActions: []string{"browser action=sync", "browser"},
	}, SharedSessionBrowserHealthActions{
		PrimaryAction:      "browser action=refresh",
		RecommendedActions: []string{"browser action=refresh", "browser action=ensure"},
	})
	if guidance.PrimaryAction != "browser action=refresh" || guidance.NextStep != "browser action=refresh" {
		t.Fatalf("expected disconnected health to override primary browser action, got %#v", guidance)
	}
	if len(guidance.RecommendedActions) != 4 ||
		guidance.RecommendedActions[0] != "browser action=refresh" ||
		guidance.RecommendedActions[1] != "browser action=ensure" ||
		guidance.RecommendedActions[2] != "browser action=sync" ||
		guidance.RecommendedActions[3] != "browser" {
		t.Fatalf("expected disconnected health to prepend recommended actions, got %#v", guidance)
	}
}

func TestEvaluateSharedSessionBrowserCoordinationEvaluationPermanentRestartFailureOverridesPlan(t *testing.T) {
	evaluation := EvaluateSharedSessionBrowserCoordinationEvaluation(SharedSessionBrowserCoordinationEvaluationInput{
		Coordination: SharedSessionBrowserCoordinationInput{
			RouteTargetCount:        1,
			SelectedBrowserProfile:  "",
			SelectedBrowserTargetID: "",
			Profiles: []SharedSessionBrowserProfileState{{
				Backend:       "proxy",
				Profile:       "isolated",
				RuntimeTarget: "node",
				Status:        "running",
				Running:       true,
				Connected:     true,
			}},
		},
		HealthEvaluation: SharedSessionBrowserHealthEvaluation{
			Summary: &SharedSessionBrowserHealthSummary{
				State:          "restart_failed_permanent",
				Reason:         "browser restart failed 2 times; explicit browser.start required",
				RecoveryAction: "browser action=start",
			},
		},
	})
	if evaluation.RestartAction != "browser action=start" {
		t.Fatalf("expected permanent restart failure to override restart action with browser.start, got %#v", evaluation)
	}
	if evaluation.Guidance.PrimaryAction != "browser action=start" || evaluation.Guidance.NextStep != "browser action=start" {
		t.Fatalf("expected permanent restart failure to override primary action with browser.start, got %#v", evaluation.Guidance)
	}
	if len(evaluation.Guidance.RecommendedActions) < 2 ||
		evaluation.Guidance.RecommendedActions[0] != "browser action=start" ||
		evaluation.Guidance.RecommendedActions[1] != "browser action=ensure" {
		t.Fatalf("expected permanent restart failure to prepend start/ensure guidance, got %#v", evaluation.Guidance)
	}
}

func TestEvaluateSharedSessionBrowserCoordinationEvaluationRestartPendingFallsBackToWaitGuidance(t *testing.T) {
	evaluation := EvaluateSharedSessionBrowserCoordinationEvaluation(SharedSessionBrowserCoordinationEvaluationInput{
		Coordination: SharedSessionBrowserCoordinationInput{
			RouteTargetCount:        1,
			SelectedBrowserProfile:  "isolated",
			SelectedBrowserTargetID: "target-1",
			Profiles: []SharedSessionBrowserProfileState{{
				Backend:       "proxy",
				Profile:       "isolated",
				RuntimeTarget: "node",
				Status:        "running",
				Running:       true,
				Connected:     true,
			}},
		},
		HealthEvaluation: SharedSessionBrowserHealthEvaluation{
			Summary: &SharedSessionBrowserHealthSummary{
				State:                   "restart_pending",
				Reason:                  "browser relaunch backoff active for 450ms after restart failure",
				ReconnectHint:           "retry_after_backoff",
				RetryBackoffRemainingMs: 450,
				RecommendedBackoffMs:    1200,
			},
		},
	})
	if evaluation.RestartAction != "" {
		t.Fatalf("expected restart_pending to clear restart action, got %#v", evaluation)
	}
	if evaluation.Guidance.PrimaryAction != "browser action=wait" || evaluation.Guidance.NextStep != "browser action=wait" {
		t.Fatalf("expected restart_pending to fall back to wait guidance, got %#v", evaluation.Guidance)
	}
	if len(evaluation.Guidance.RecommendedActions) == 0 || evaluation.Guidance.RecommendedActions[0] != "browser action=wait" {
		t.Fatalf("expected restart_pending to prepend wait guidance, got %#v", evaluation.Guidance)
	}
}

func TestEvaluateSharedSessionBrowserCoordinationBrowserReadyNeedsSync(t *testing.T) {
	plan := EvaluateSharedSessionBrowserCoordination(SharedSessionBrowserCoordinationInput{
		RouteTargetCount:        1,
		SelectedBrowserProfile:  "",
		SelectedBrowserTargetID: "",
		Profiles: []SharedSessionBrowserProfileState{{
			Backend:       "proxy",
			Profile:       "isolated",
			RuntimeTarget: "node",
			Status:        "running",
			Running:       true,
			Connected:     true,
		}},
	})
	if plan.State != "browser_ready" || !plan.BrowserOnNode || !plan.HasRunningBrowserProfile || plan.HasActiveNodeRun {
		t.Fatalf("expected browser_ready node-backed plan, got %#v", plan)
	}
	if !plan.NeedsSessionSync || plan.SyncAction != "browser action=sync" {
		t.Fatalf("expected browser_ready plan to request session sync, got %#v", plan)
	}
	if plan.RestartAction != "browser action=refresh" || plan.TeardownAction != "browser action=teardown" {
		t.Fatalf("expected browser_ready plan to expose refresh+teardown, got %#v", plan)
	}
	if len(plan.RecommendedBrowserActions) != 3 ||
		plan.RecommendedBrowserActions[0] != "browser action=sync" ||
		plan.RecommendedBrowserActions[1] != "browser" ||
		plan.RecommendedBrowserActions[2] != "browser action=refresh" {
		t.Fatalf("expected browser_ready browser action plan, got %#v", plan)
	}
	if plan.PrimaryBrowserAction != "browser action=sync" || plan.PrimaryNodeAction != "nodes action=run" || plan.NextStep != "browser action=sync" {
		t.Fatalf("expected browser_ready primary actions to follow sync-first plan, got %#v", plan)
	}
}

func TestEvaluateSharedSessionBrowserCoordinationCoordinatedPrefersNodeRunStatus(t *testing.T) {
	plan := EvaluateSharedSessionBrowserCoordination(SharedSessionBrowserCoordinationInput{
		ActiveNodeRunID:         "run-55",
		SelectedBrowserProfile:  "isolated",
		SelectedBrowserTargetID: "tab-1",
		Profiles: []SharedSessionBrowserProfileState{{
			Backend:       "proxy",
			Profile:       "isolated",
			RuntimeTarget: "node",
			Status:        "running",
			Running:       true,
			Connected:     true,
		}},
	})
	if plan.State != "coordinated" || !plan.BrowserOnNode || !plan.HasRunningBrowserProfile || !plan.HasActiveNodeRun {
		t.Fatalf("expected coordinated node-backed plan, got %#v", plan)
	}
	if plan.NeedsSessionSync || plan.SyncAction != "" {
		t.Fatalf("expected coordinated plan not to require session sync when selection is complete, got %#v", plan)
	}
	if len(plan.RecommendedBrowserActions) != 1 || plan.RecommendedBrowserActions[0] != "browser" {
		t.Fatalf("expected coordinated browser actions to collapse to browser, got %#v", plan)
	}
	if len(plan.RecommendedNodeActions) != 3 || plan.RecommendedNodeActions[0] != "nodes action=run_status" {
		t.Fatalf("expected coordinated node actions to prefer run_status, got %#v", plan)
	}
	if plan.PrimaryBrowserAction != "browser" || plan.PrimaryNodeAction != "nodes action=run_status" || plan.NextStep != "nodes action=run_status" {
		t.Fatalf("expected coordinated next step to prefer node run status, got %#v", plan)
	}
}

func TestRecommendSharedSessionBrowserHealthActionsStaleRouteTargetsHintsResetAndEnsure(t *testing.T) {
	actions := RecommendSharedSessionBrowserHealthActions(false, SharedSessionBrowserHealthEvaluation{
		Summary: &SharedSessionBrowserHealthSummary{
			State:          "stale_route_targets",
			Reason:         "session still tracks browser targets after the managed browser profile stopped",
			RecoveryAction: "browser action=reset",
		},
	})
	if actions.PrimaryAction != "browser action=reset" {
		t.Fatalf("expected stale route targets to prefer reset, got %#v", actions)
	}
	if len(actions.RecommendedActions) != 2 ||
		actions.RecommendedActions[0] != "browser action=reset" ||
		actions.RecommendedActions[1] != "browser action=ensure" {
		t.Fatalf("expected stale route targets to recommend reset+ensure, got %#v", actions)
	}
	if actions.SuppressRefresh {
		t.Fatalf("expected stale route targets not to suppress refresh, got %#v", actions)
	}
}

func TestEvaluateSharedSessionBrowserFollowPolicyActionsPopupStorm(t *testing.T) {
	actions := EvaluateSharedSessionBrowserFollowPolicyActions([]SharedSessionBrowserRouteCoordinationInput{
		{FollowPolicyState: "popup_review_required", ManagedRuntime: true},
		{FollowPolicyState: "popup_storm_review_required", ManagedRuntime: true},
	})
	if actions.PrimaryAction != "browser action=close" {
		t.Fatalf("expected popup storm to prefer close, got %#v", actions)
	}
	if len(actions.RecommendedActions) != 3 ||
		actions.RecommendedActions[0] != "browser action=close" ||
		actions.RecommendedActions[1] != "browser action=tabs" ||
		actions.RecommendedActions[2] != "browser action=pin_target" {
		t.Fatalf("expected popup storm guidance, got %#v", actions)
	}
}

func TestSharedSessionBrowserSelectedFollowPolicyState(t *testing.T) {
	state := SharedSessionBrowserSelectedFollowPolicyState([]SharedSessionBrowserRouteCoordinationInput{
		{FollowPolicyState: "popup_review_required", ManagedRuntime: true},
		{FollowPolicyState: "redirect_review_required", ManagedRuntime: true},
		{FollowPolicyState: "popup_storm_review_required", ManagedRuntime: true},
	})
	if state != "popup_storm_review_required" {
		t.Fatalf("expected highest-priority follow policy state, got %q", state)
	}
}

func TestEvaluateSharedSessionBrowserManagedRouteRecoveryActionsStaleManagedRoute(t *testing.T) {
	actions := EvaluateSharedSessionBrowserManagedRouteRecoveryActions(SharedSessionBrowserManagedRouteRecoveryInput{
		Profiles: []SharedSessionBrowserProfileState{{
			Backend:       "proxy",
			Profile:       "isolated",
			RuntimeTarget: "node",
			Status:        "stopped",
			Running:       false,
			Connected:     false,
		}},
		Routes: []SharedSessionBrowserRouteCoordinationInput{{
			FollowPolicyState: "redirect_review_required",
			ManagedRuntime:    true,
		}},
	})
	if actions.PrimaryAction != "browser action=reset" {
		t.Fatalf("expected stale managed route to prefer reset, got %#v", actions)
	}
	if len(actions.RecommendedActions) != 2 ||
		actions.RecommendedActions[0] != "browser action=reset" ||
		actions.RecommendedActions[1] != "browser action=ensure" {
		t.Fatalf("expected stale managed route reset+ensure guidance, got %#v", actions)
	}
}

func TestApplySharedSessionBrowserCoordinationActionsOverridesPrimaryAction(t *testing.T) {
	guidance := ApplySharedSessionBrowserCoordinationActions(SharedSessionBrowserCoordinationGuidance{
		PrimaryAction:      "browser action=sync",
		NextStep:           "browser action=sync",
		RecommendedActions: []string{"browser action=sync", "browser"},
	}, SharedSessionBrowserCoordinationActions{
		PrimaryAction:      "browser action=close",
		RecommendedActions: []string{"browser action=close", "browser action=tabs", "browser action=pin_target"},
	})
	if guidance.PrimaryAction != "browser action=close" || guidance.NextStep != "browser action=close" {
		t.Fatalf("expected overlay to override primary browser action, got %#v", guidance)
	}
	if len(guidance.RecommendedActions) != 5 ||
		guidance.RecommendedActions[0] != "browser action=close" ||
		guidance.RecommendedActions[1] != "browser action=tabs" ||
		guidance.RecommendedActions[2] != "browser action=pin_target" ||
		guidance.RecommendedActions[3] != "browser action=sync" ||
		guidance.RecommendedActions[4] != "browser" {
		t.Fatalf("expected overlay to prepend recommended actions, got %#v", guidance)
	}
}

func TestEvaluateSharedSessionBrowserCoordinationEvaluationPopupStormOverridesPlan(t *testing.T) {
	evaluation := EvaluateSharedSessionBrowserCoordinationEvaluation(SharedSessionBrowserCoordinationEvaluationInput{
		Coordination: SharedSessionBrowserCoordinationInput{
			RouteTargetCount:        1,
			SelectedBrowserProfile:  "",
			SelectedBrowserTargetID: "",
			Profiles: []SharedSessionBrowserProfileState{{
				Backend:       "proxy",
				Profile:       "isolated",
				RuntimeTarget: "node",
				Status:        "running",
				Running:       true,
				Connected:     true,
			}},
		},
		Routes: []SharedSessionBrowserRouteCoordinationInput{
			{FollowPolicyState: "popup_storm_review_required", ManagedRuntime: true},
		},
	})
	if evaluation.Plan.State != "browser_ready" || evaluation.RestartAction != "browser action=refresh" {
		t.Fatalf("expected browser_ready base plan to stay intact, got %#v", evaluation)
	}
	if evaluation.Guidance.PrimaryAction != "browser action=close" || evaluation.Guidance.NextStep != "browser action=close" {
		t.Fatalf("expected popup storm overlay to override primary browser action, got %#v", evaluation)
	}
	if len(evaluation.Guidance.RecommendedActions) < 4 ||
		evaluation.Guidance.RecommendedActions[0] != "browser action=close" ||
		evaluation.Guidance.RecommendedActions[1] != "browser action=tabs" ||
		evaluation.Guidance.RecommendedActions[2] != "browser action=pin_target" ||
		evaluation.Guidance.RecommendedActions[3] != "browser action=sync" {
		t.Fatalf("expected popup storm overlay to prepend cleanup actions ahead of base plan, got %#v", evaluation.Guidance)
	}
}

func TestEvaluateSharedSessionBrowserCoordinationEvaluationStaleManagedRoutePrefersReset(t *testing.T) {
	evaluation := EvaluateSharedSessionBrowserCoordinationEvaluation(SharedSessionBrowserCoordinationEvaluationInput{
		Coordination: SharedSessionBrowserCoordinationInput{
			Profiles: []SharedSessionBrowserProfileState{{
				Backend:       "proxy",
				Profile:       "isolated",
				RuntimeTarget: "node",
				Status:        "stopped",
				Running:       false,
				Connected:     false,
			}},
		},
		Routes: []SharedSessionBrowserRouteCoordinationInput{
			{FollowPolicyState: "redirect_review_required", ManagedRuntime: true},
		},
	})
	if evaluation.Plan.State != "browser_attached" {
		t.Fatalf("expected browser_attached base plan, got %#v", evaluation)
	}
	if evaluation.Guidance.PrimaryAction != "browser action=reset" || evaluation.Guidance.NextStep != "browser action=reset" {
		t.Fatalf("expected stale managed route recovery to prefer reset, got %#v", evaluation)
	}
	if len(evaluation.Guidance.RecommendedActions) < 2 ||
		evaluation.Guidance.RecommendedActions[0] != "browser action=reset" ||
		evaluation.Guidance.RecommendedActions[1] != "browser action=ensure" {
		t.Fatalf("expected stale managed route recovery to prepend reset/ensure, got %#v", evaluation.Guidance)
	}
	for _, action := range evaluation.Guidance.RecommendedActions {
		if action == "browser action=pin_target" {
			t.Fatalf("expected stale managed route recovery to replace follow-policy cleanup actions, got %#v", evaluation.Guidance)
		}
	}
}

func TestDecideSharedSessionBrowserCoordinationRestartReadyAndStartedAliases(t *testing.T) {
	if decision := DecideSharedSessionBrowserCoordination("browser_ready", "", "restart"); decision != "restart_ready" {
		t.Fatalf("expected browser_ready restart to become restart_ready, got %q", decision)
	}
	if decision := DecideSharedSessionBrowserCoordination("coordinated", "started", ""); decision != "started_for_active_node_run" {
		t.Fatalf("expected coordinated started decision alias, got %q", decision)
	}
	if decision := DecideSharedSessionBrowserCoordination("browser_ready", "started", ""); decision != "started_browser_profile" {
		t.Fatalf("expected browser_ready started decision alias, got %q", decision)
	}
}

func TestSharedSessionBrowserCoordinationReadyUsesProfileAndGoal(t *testing.T) {
	reconnecting := &BrowserProfileStatusResult{
		Status:    "reconnecting",
		Running:   true,
		Connected: false,
	}
	if SharedSessionBrowserCoordinationReady("browser_ready", "restart", reconnecting, false) {
		t.Fatalf("expected reconnecting restart coordination to remain not ready")
	}
	connected := &BrowserProfileStatusResult{
		Status:    "connected",
		Running:   true,
		Connected: true,
	}
	if !SharedSessionBrowserCoordinationReady("browser_ready", "restart", connected, false) {
		t.Fatalf("expected connected restart coordination to be ready")
	}
	if SharedSessionBrowserCoordinationReady("browser_ready", "teardown", connected, false) {
		t.Fatalf("expected teardown coordination with connected profile to remain not ready")
	}
	if !SharedSessionBrowserCoordinationReady("", "sync", nil, true) {
		t.Fatalf("expected sync coordination to follow syncReady")
	}
}

func TestEvaluateSharedSessionBrowserCoordinationStatusRestartReadyAlias(t *testing.T) {
	status := EvaluateSharedSessionBrowserCoordinationStatus("browser_ready", "restart", &BrowserProfileStatusResult{
		Status:    "connected",
		Running:   true,
		Connected: true,
	}, false, "restarted")
	if status.Decision != "restart_ready" || !status.Ready {
		t.Fatalf("expected ready restart coordination to alias to restart_ready, got %#v", status)
	}

	status = EvaluateSharedSessionBrowserCoordinationStatus("browser_ready", "restart", &BrowserProfileStatusResult{
		Status:    "reconnecting",
		Running:   true,
		Connected: false,
	}, false, "restart_reconnect_in_progress")
	if status.Decision != "restart_reconnect_in_progress" || status.Ready {
		t.Fatalf("expected reconnect-in-progress restart coordination to stay in-progress, got %#v", status)
	}
}

func TestAssessSharedSessionBrowserProfileRecoveryReconnectInProgress(t *testing.T) {
	reconnectingAt := time.Now().Add(-12 * time.Second)
	assessment := AssessSharedSessionBrowserProfileRecovery(SharedSessionBrowserHealthEvaluation{
		Summary: &SharedSessionBrowserHealthSummary{
			State:  "profile_reconnecting",
			Reason: "cdp reconnect in progress",
		},
		Profile: SharedSessionBrowserProfileState{
			Backend:       "proxy",
			Profile:       "isolated",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Status:        "reconnecting",
			Running:       true,
			Connected:     false,
			ObservedAt:    reconnectingAt,
			StatusSince:   reconnectingAt,
			Note:          "cdp reconnect in progress",
		},
		HasProfile:        true,
		ReconnectTimedOut: false,
	}, BrowserRuntimeInfo{
		Backend: "proxy",
		Profile: "isolated",
		Target:  "node",
	}, "isolated", BrowserProfileStatusResult{})
	if !assessment.ReconnectInProgress || !assessment.HasSyntheticStatus {
		t.Fatalf("expected reconnecting health to keep reconnect-in-progress synthetic status, got %#v", assessment)
	}
	if assessment.NeedsRefreshRecovery || !assessment.ShouldStopBeforeRecovery {
		t.Fatalf("expected reconnecting health inside watchdog window to avoid refresh but keep stop-before-recovery posture, got %#v", assessment)
	}
	if assessment.SyntheticStatus.Profile != "isolated" || assessment.SyntheticStatus.Status != "reconnecting" || !assessment.SyntheticStatus.Running || assessment.SyntheticStatus.Connected {
		t.Fatalf("expected reconnecting synthetic status, got %#v", assessment.SyntheticStatus)
	}
}

func TestAssessSharedSessionBrowserProfileRecoveryExplicitHealthyRawStatusDoesNotSuppressStoredRecovery(t *testing.T) {
	assessment := AssessSharedSessionBrowserProfileRecovery(SharedSessionBrowserHealthEvaluation{
		Summary: &SharedSessionBrowserHealthSummary{
			State:          "profile_disconnected",
			Reason:         "stored disconnected posture",
			RecoveryAction: "browser action=refresh",
		},
		Profile: SharedSessionBrowserProfileState{
			Backend:       "proxy",
			Profile:       "isolated",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Status:        "disconnected",
			Running:       true,
			Connected:     false,
			Note:          "cdp transport closed",
		},
		HasProfile: true,
	}, BrowserRuntimeInfo{
		Backend: "proxy",
		Profile: "isolated",
		Target:  "node",
	}, "isolated", BrowserProfileStatusResult{
		Backend:    "proxy",
		BrowserApp: "Chromium",
		Profile:    "isolated",
		Status:     "connected",
		Running:    true,
		Connected:  true,
	})
	if !assessment.NeedsRefreshRecovery || !assessment.ShouldStopBeforeRecovery {
		t.Fatalf("expected stored disconnected posture to preserve refresh+stop recovery, got %#v", assessment)
	}
	if assessment.ReconnectInProgress {
		t.Fatalf("expected disconnected posture not to become reconnect-in-progress, got %#v", assessment)
	}
	if assessment.EffectiveStatus.Status != "disconnected" || assessment.EffectiveStatus.Profile != "isolated" {
		t.Fatalf("expected effective status to keep stored disconnected profile, got %#v", assessment.EffectiveStatus)
	}
}

func TestAssessSharedSessionBrowserProfileRecoveryTypedBlockersSuppressLifecycleChurn(t *testing.T) {
	tests := []struct {
		name           string
		summary        SharedSessionBrowserHealthSummary
		wantDecision   string
		wantDisconnect bool
	}{
		{
			name: "cooldown_active",
			summary: SharedSessionBrowserHealthSummary{
				State:               "cooldown_active",
				Reason:              "browser restart cooldown active for 900ms after 2 disconnects",
				RecoveryAction:      "browser action=wait",
				ReconnectHint:       "retry_after_cooldown",
				CooldownRemainingMs: 900,
			},
			wantDecision:   "cooldown_active",
			wantDisconnect: true,
		},
		{
			name: "restart_failed_permanent",
			summary: SharedSessionBrowserHealthSummary{
				State:          "restart_failed_permanent",
				Reason:         "browser restart failed 2 times; explicit browser.start required",
				RecoveryAction: "browser action=start",
				ReconnectHint:  "manual_restart_required",
			},
			wantDecision:   "restart_failed_permanent",
			wantDisconnect: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluation := SharedSessionBrowserHealthEvaluation{
				Summary: &tt.summary,
			}
			assessment := AssessSharedSessionBrowserProfileRecovery(evaluation, BrowserRuntimeInfo{
				Backend: "proxy",
				Profile: "isolated",
				Target:  "node",
			}, "isolated", BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "isolated",
				Status:     "disconnected",
				Running:    false,
				Connected:  false,
				Note:       "cdp transport closed",
			})
			if assessment.NeedsRefreshRecovery || assessment.ShouldStopBeforeRecovery || assessment.ReconnectInProgress {
				t.Fatalf("expected typed blocker %q to suppress lifecycle churn, got %#v", tt.wantDecision, assessment)
			}
			if assessment.EffectiveStatus.Status != "disconnected" || assessment.EffectiveStatus.Profile != "isolated" {
				t.Fatalf("expected typed blocker %q to preserve effective disconnected status, got %#v", tt.wantDecision, assessment.EffectiveStatus)
			}
			if decision, ok := SharedSessionBrowserExecutionBlockedDecision(evaluation); !ok || decision != tt.wantDecision {
				t.Fatalf("expected typed blocker helper to expose %q, got decision=%q ok=%v", tt.wantDecision, decision, ok)
			}
		})
	}
}

func TestAssessSharedSessionBrowserProfileRecoveryForInputScopeIgnoresUnrelatedProfile(t *testing.T) {
	reconnectingAt := time.Now().Add(-12 * time.Second)
	health, assessment := AssessSharedSessionBrowserProfileRecoveryForInputScope(SharedSessionBrowserHealthInput{
		Profiles: []SharedSessionBrowserProfileState{
			{
				Backend:       "proxy",
				Profile:       "isolated",
				RuntimeTarget: "node",
				BrowserApp:    "Chromium",
				Status:        "reconnecting",
				Running:       true,
				Connected:     false,
				ObservedAt:    reconnectingAt,
				StatusSince:   reconnectingAt,
				Note:          "cdp reconnect in progress",
			},
			{
				Backend:       "proxy",
				Profile:       "work",
				RuntimeTarget: "node",
				BrowserApp:    "Chromium",
				Status:        "connected",
				Running:       true,
				Connected:     true,
			},
		},
	}, BrowserRuntimeInfo{
		Backend: "proxy",
		Target:  "node",
	}, "work", BrowserProfileStatusResult{}, 54*time.Second)
	if health.Summary == nil || health.Summary.State != "healthy" {
		t.Fatalf("expected input-scoped recovery health to ignore unrelated reconnecting profile, got %#v", health)
	}
	if assessment.EffectiveStatus.Profile != "work" || assessment.EffectiveStatus.Status != "connected" || !assessment.EffectiveStatus.Connected {
		t.Fatalf("expected connected requested profile to determine scoped recovery status, got %#v", assessment)
	}
	if assessment.NeedsRefreshRecovery || assessment.ReconnectInProgress {
		t.Fatalf("expected connected requested profile to avoid reconnect recovery, got %#v", assessment)
	}
}

func sharedSessionBrowserActionSliceContains(items []string, needle string) bool {
	needle = strings.TrimSpace(needle)
	for _, item := range items {
		if strings.EqualFold(strings.TrimSpace(item), needle) {
			return true
		}
	}
	return false
}

func TestSharedSessionBrowserLifecycleDecisionStatusUsesLifecycleState(t *testing.T) {
	status := SharedSessionBrowserLifecycleDecisionStatus(BrowserRuntimeInfo{
		Backend: "proxy",
		Profile: "isolated",
		Target:  "node",
	}, "isolated", BrowserProfileStatusResult{
		Backend:    "proxy",
		Profile:    "isolated",
		BrowserApp: "Chromium",
		Status:     "started",
		Running:    true,
		Connected:  false,
	}, "restart_started")
	if status.Profile != "isolated" || status.Status != "reconnecting" || !status.Running || status.Connected || status.Note != "restart requested" {
		t.Fatalf("expected lifecycle decision to collapse weak restart status into reconnecting state, got %#v", status)
	}
}

func TestSharedSessionBrowserLifecycleDecisionReadyFollowsStartingAndConnectedStates(t *testing.T) {
	if SharedSessionBrowserLifecycleDecisionReady(BrowserRuntimeInfo{
		Backend: "proxy",
		Profile: "isolated",
		Target:  "node",
	}, "isolated", BrowserProfileStatusResult{
		Backend:    "proxy",
		Profile:    "isolated",
		BrowserApp: "Chromium",
		Status:     "started",
		Running:    true,
		Connected:  false,
	}, "started") {
		t.Fatalf("expected started decision to remain not-ready while lifecycle state is starting")
	}
	if !SharedSessionBrowserLifecycleDecisionReady(BrowserRuntimeInfo{
		Backend: "proxy",
		Profile: "isolated",
		Target:  "node",
	}, "isolated", BrowserProfileStatusResult{
		Backend:    "proxy",
		Profile:    "isolated",
		BrowserApp: "Chromium",
		Status:     "connected",
		Running:    true,
		Connected:  true,
	}, "already_ready") {
		t.Fatalf("expected connected lifecycle decision to stay ready")
	}
	if !SharedSessionBrowserLifecycleDecisionReady(BrowserRuntimeInfo{
		Backend: "proxy",
		Profile: "isolated",
		Target:  "node",
	}, "isolated", BrowserProfileStatusResult{
		Backend:    "proxy",
		Profile:    "isolated",
		BrowserApp: "Chromium",
		Status:     "stopped",
		Running:    false,
		Connected:  false,
	}, "stopped") {
		t.Fatalf("expected stopped lifecycle decision to stay ready once stop completed")
	}
}
