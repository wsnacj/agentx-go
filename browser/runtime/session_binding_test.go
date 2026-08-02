package browserruntime

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"
)

type testSharedSessionRunRegistry struct {
	items map[string][]SharedSessionRunInfo
}

func (r testSharedSessionRunRegistry) SnapshotSessionRuns(sessionID string) []SharedSessionRunInfo {
	if r.items == nil {
		return nil
	}
	return append([]SharedSessionRunInfo(nil), r.items[sessionID]...)
}

func TestSummarizeSharedSessionBrowserBindingAggregatesRouteRunAndProfileState(t *testing.T) {
	summary := SummarizeSharedSessionBrowserBinding(
		[]SharedSessionBrowserRouteSnapshot{
			{
				CurrentTargetID:          "tab-1",
				PendingTargetReview:      &BrowserSessionTargetReview{ID: "tab-2"},
				PendingTargetReviewCount: 2,
				FollowPolicyState:        "popup_storm_review_required",
				PopupPolicyState:         "popup_storm_review_required",
				Targets: []SharedSessionBrowserRouteTarget{
					{ID: "tab-1", Current: true},
					{ID: "tab-2"},
				},
			},
		},
		[]SharedSessionRunInfo{
			{RunID: "run-1", Status: "running"},
			{RunID: "run-2", Status: "failed"},
		},
		[]SharedSessionBrowserProfileState{
			{Profile: "workbench", Status: "running", Running: true, Connected: true},
			{Profile: "alt", Status: "stopped"},
		},
	)
	if summary.CurrentTargetID != "tab-1" || summary.RouteTargetCount != 2 || summary.PendingTargetReviewCount != 1 || summary.BlockedAutoFollowRouteCount != 1 || summary.PopupStormRouteCount != 1 {
		t.Fatalf("unexpected route summary: %#v", summary)
	}
	if summary.NodeRunCount != 2 || summary.ActiveNodeRunID != "run-1" || summary.NodeRunStatusCounts["running"] != 1 || summary.NodeRunStatusCounts["failed"] != 1 {
		t.Fatalf("unexpected run summary: %#v", summary)
	}
	if summary.BrowserProfileCount != 2 || summary.ActiveBrowserProfile != "workbench" || summary.BrowserProfileStatusCounts["running"] != 1 || summary.BrowserProfileStatusCounts["stopped"] != 1 {
		t.Fatalf("unexpected profile summary: %#v", summary)
	}
}

func TestSummarizeSharedSessionRunsFallsBackToFirstRun(t *testing.T) {
	active, counts := SummarizeSharedSessionRuns([]SharedSessionRunInfo{
		{RunID: "run-9", Status: "failed"},
		{RunID: "run-10", Status: "completed"},
	})
	if active != "run-9" {
		t.Fatalf("expected first run fallback, got %q", active)
	}
	if counts["failed"] != 1 || counts["completed"] != 1 {
		t.Fatalf("unexpected counts: %#v", counts)
	}
}

func TestSummarizeSharedSessionBrowserProfilesFallsBackToFirstProfile(t *testing.T) {
	active, counts := SummarizeSharedSessionBrowserProfiles([]SharedSessionBrowserProfileState{
		{Profile: "workbench", Status: "stopped"},
		{Profile: "alt", Status: "crashed"},
	})
	if active != "workbench" {
		t.Fatalf("expected first profile fallback, got %q", active)
	}
	if counts["stopped"] != 1 || counts["crashed"] != 1 {
		t.Fatalf("unexpected counts: %#v", counts)
	}
}

func TestSnapshotSharedSessionBrowserBindingIncludesSelectionsAndScopeSummary(t *testing.T) {
	sessionID := "sess-1"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "work", Target: "node"}
	registry := NewBrowserSessionRegistry()
	tracked := registry.TrackTabs(sessionID, []BrowserSessionTarget{
		{ID: "tab-1", TabIndex: 1, URL: "https://example.com", Backend: "proxy", Profile: "work", Target: "node"},
	}, 1)

	stateRegistry := NewBrowserSessionStateRegistry()
	stateRegistry.SelectBrowserProfile(sessionID, SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "work",
		RuntimeTarget: "node",
		BrowserApp:    "chrome",
		Source:        "select_profile",
	})
	stateRegistry.RecordBrowserProfileState(sessionID, SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "work",
		RuntimeTarget: "node",
		BrowserApp:    "chrome",
		Status:        "running",
		Running:       true,
		Connected:     true,
	})

	snapshot := SnapshotSharedSessionBrowserBinding(
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "work", Target: "node"},
		route,
		nil,
		registry,
		testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			sessionID: {{RunID: "run-1", Status: "running"}},
		}},
		stateRegistry,
	)

	if snapshot.SelectedProfileSelection == nil || snapshot.SelectedProfileSelection.Profile != "work" {
		t.Fatalf("expected selected profile selection, got %#v", snapshot.SelectedProfileSelection)
	}
	if snapshot.SelectedTargetSelection == nil || snapshot.SelectedTargetSelection.ID != tracked[0].ID {
		t.Fatalf("expected selected target selection, got %#v", snapshot.SelectedTargetSelection)
	}
	if snapshot.CurrentTargetID != tracked[0].ID {
		t.Fatalf("expected current target fallback, got %q", snapshot.CurrentTargetID)
	}
	if len(snapshot.Runs) != 1 || snapshot.Summary.ActiveNodeRunID != "run-1" {
		t.Fatalf("unexpected run snapshot: %#v / %#v", snapshot.Runs, snapshot.Summary)
	}
	if len(snapshot.Profiles) != 1 || snapshot.Summary.ActiveBrowserProfile != "work" {
		t.Fatalf("unexpected profile snapshot: %#v / %#v", snapshot.Profiles, snapshot.Summary)
	}
	if snapshot.Summary.RouteTargetCount != 1 {
		t.Fatalf("expected route target count, got %#v", snapshot.Summary)
	}
}

func TestBuildSharedSessionBrowserHealthAndCoordinationInputs(t *testing.T) {
	profiles := []SharedSessionBrowserProfileState{
		{Backend: "proxy", Profile: "work", RuntimeTarget: "node", Status: "running", Running: true, Connected: true},
	}
	health := BuildSharedSessionBrowserHealthInput("run-1", 2, "healthy", "ok", "none", profiles)
	if health.ActiveNodeRunID != "run-1" || health.RouteTargetCount != 2 || health.StoredState != "healthy" || len(health.Profiles) != 1 {
		t.Fatalf("unexpected health input: %#v", health)
	}

	coordination := BuildSharedSessionBrowserCoordinationInput(
		"run-1",
		2,
		&SharedSessionBrowserProfileSelection{Profile: "work"},
		&BrowserSessionTargetSelection{ID: "tab-1"},
		profiles,
	)
	if coordination.ActiveNodeRunID != "run-1" || coordination.RouteTargetCount != 2 || coordination.SelectedBrowserProfile != "work" || coordination.SelectedBrowserTargetID != "tab-1" || len(coordination.Profiles) != 1 {
		t.Fatalf("unexpected coordination input: %#v", coordination)
	}
}

func TestBuildSharedSessionBrowserRequestsFromBindingEvaluation(t *testing.T) {
	referenceTime := time.Date(2026, time.March, 29, 12, 0, 0, 0, time.UTC)
	evaluation := SharedSessionBrowserBindingEvaluation{
		Snapshot: SharedSessionBrowserBindingSnapshot{
			SelectedProfileSelection: &SharedSessionBrowserProfileSelection{
				Backend:       "proxy",
				Profile:       "work",
				RuntimeTarget: "node",
				BrowserApp:    "Chromium",
				Source:        "select_profile",
			},
			SelectedTargetSelection: &BrowserSessionTargetSelection{
				ID:            "tab-1",
				Backend:       "proxy",
				Profile:       "work",
				RuntimeTarget: "node",
				BrowserApp:    "Chromium",
				Source:        "tracked_active_tab",
			},
			Profiles: []SharedSessionBrowserProfileState{{
				Backend:       "proxy",
				Profile:       "work",
				RuntimeTarget: "node",
				BrowserApp:    "Chromium",
				Status:        "running",
				Running:       true,
				Connected:     true,
				ObservedAt:    referenceTime,
			}},
			Summary: SharedSessionBrowserBindingSummary{
				RouteTargetCount:     2,
				ActiveNodeRunID:      "run-1",
				ActiveBrowserProfile: "work",
			},
		},
		ReferenceTime: referenceTime,
		Health: SharedSessionBrowserHealthEvaluation{
			Summary: &SharedSessionBrowserHealthSummary{
				State:             "healthy",
				Reason:            "ok",
				RecoveryAction:    "none",
				ResolverBlockedBy: "multiple_candidates_filtered",
				AmbiguityClass:    "filtered_residual",
				CandidateKind:     "label",
				CandidateStrength: "medium",
				RetryDisposition:  "manual_only",
				ManualRetryHint:   "add_ordinal",
				NextStepAlias:     "snapshot",
				SpecificityFields: []string{"tag", "type"},
			},
		},
	}

	health := BuildSharedSessionBrowserHealthInputFromBindingEvaluation(evaluation)
	if health.ActiveNodeRunID != "run-1" || health.RouteTargetCount != 2 || health.StoredState != "healthy" || !health.ReferenceTime.Equal(referenceTime) {
		t.Fatalf("unexpected binding-evaluation health input: %#v", health)
	}
	if health.StoredResolverBlockedBy != "multiple_candidates_filtered" ||
		health.StoredAmbiguityClass != "filtered_residual" ||
		health.StoredCandidateKind != "label" ||
		health.StoredCandidateStrength != "medium" ||
		health.StoredRetryDisposition != "manual_only" ||
		health.StoredManualRetryHint != "add_ordinal" ||
		health.StoredNextStepAlias != "snapshot" ||
		!reflect.DeepEqual(health.StoredSpecificityFields, []string{"tag", "type"}) {
		t.Fatalf("expected binding-evaluation health input to preserve resolver guidance, got %#v", health)
	}

	coordination := BuildSharedSessionBrowserCoordinationInputFromBindingEvaluation(evaluation)
	if coordination.ActiveNodeRunID != "run-1" || coordination.RouteTargetCount != 2 || coordination.SelectedBrowserProfile != "work" || coordination.SelectedBrowserTargetID != "tab-1" {
		t.Fatalf("unexpected binding-evaluation coordination input: %#v", coordination)
	}

	execution := BuildSharedSessionBrowserExecutionRequestFromBindingEvaluation(
		nil,
		"sess-1",
		"work",
		BrowserRuntimeInfo{Backend: "proxy", Profile: "work", Target: "node"},
		true,
		evaluation,
		time.Minute,
	)
	if execution.ActiveNodeRunID != "run-1" || execution.ActiveBrowserProfile != "work" || !execution.Force || !execution.HealthInput.ReferenceTime.Equal(referenceTime) {
		t.Fatalf("unexpected binding-evaluation execution request: %#v", execution)
	}

	clearReq := BuildSharedSessionBrowserClearRequestFromBindingEvaluation(
		nil,
		nil,
		"sess-1",
		BrowserRuntimeInfo{Backend: "proxy", Profile: "work", Target: "node"},
		BrowserSessionRoute{Backend: "proxy", Profile: "work", Target: "node"},
		true,
		evaluation,
		time.Minute,
	)
	if clearReq.ActiveNodeRunID != "run-1" || !clearReq.Force || !clearReq.HealthInput.ReferenceTime.Equal(referenceTime) {
		t.Fatalf("unexpected binding-evaluation clear request: %#v", clearReq)
	}
}

func TestBuildSharedSessionBrowserRequestsForOptionalBindingEvaluation(t *testing.T) {
	referenceTime := time.Now()
	evaluation := &SharedSessionBrowserBindingEvaluation{
		Snapshot: SharedSessionBrowserBindingSnapshot{
			Summary: SharedSessionBrowserBindingSummary{
				ActiveNodeRunID:      "run-optional",
				ActiveBrowserProfile: "workbench",
				RouteTargetCount:     3,
			},
		},
		Health: SharedSessionBrowserHealthEvaluation{
			Summary: &SharedSessionBrowserHealthSummary{
				State:          "healthy",
				Reason:         "shared",
				RecoveryAction: "none",
			},
		},
		ReferenceTime: referenceTime,
	}

	execution := BuildSharedSessionBrowserExecutionRequestForBindingEvaluation(
		nil,
		"sess-optional",
		"workbench",
		BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		true,
		evaluation,
		time.Minute,
	)
	if execution.ActiveNodeRunID != "run-optional" || execution.ActiveBrowserProfile != "workbench" || !execution.Force {
		t.Fatalf("expected optional execution request to preserve shared evaluation summary, got %#v", execution)
	}
	if execution.HealthInput.StoredState != "healthy" || !execution.HealthInput.ReferenceTime.Equal(referenceTime) {
		t.Fatalf("expected optional execution request to preserve shared health input, got %#v", execution.HealthInput)
	}

	clearReq := BuildSharedSessionBrowserClearRequestForBindingEvaluation(
		nil,
		nil,
		"sess-optional",
		BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node"},
		true,
		evaluation,
		time.Minute,
	)
	if clearReq.ActiveNodeRunID != "run-optional" || !clearReq.Force {
		t.Fatalf("expected optional clear request to preserve shared evaluation summary, got %#v", clearReq)
	}
	if clearReq.HealthInput.StoredState != "healthy" || !clearReq.HealthInput.ReferenceTime.Equal(referenceTime) {
		t.Fatalf("expected optional clear request to preserve shared health input, got %#v", clearReq.HealthInput)
	}

	emptyExecution := BuildSharedSessionBrowserExecutionRequestForBindingEvaluation(
		nil,
		"sess-empty",
		"",
		BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
		false,
		nil,
		time.Minute,
	)
	if emptyExecution.ActiveNodeRunID != "" || emptyExecution.ActiveBrowserProfile != "" || emptyExecution.HealthInput.StoredState != "" {
		t.Fatalf("expected nil optional execution evaluation to fall back to empty shared contract, got %#v", emptyExecution)
	}
}

func TestSharedSessionBrowserRouteCoordinationInputs(t *testing.T) {
	inputs := SharedSessionBrowserRouteCoordinationInputs([]SharedSessionBrowserRouteSnapshot{
		{RuntimeTarget: "node", FollowPolicyState: "popup_review_required"},
		{RuntimeTarget: "host", FollowPolicyState: "auto_follow_allowed"},
	})
	if len(inputs) != 2 {
		t.Fatalf("expected route coordination inputs, got %#v", inputs)
	}
	if !inputs[0].ManagedRuntime || inputs[0].FollowPolicyState != "popup_review_required" {
		t.Fatalf("unexpected managed route input: %#v", inputs[0])
	}
	if inputs[1].ManagedRuntime || inputs[1].FollowPolicyState != "auto_follow_allowed" {
		t.Fatalf("unexpected host route input: %#v", inputs[1])
	}
}

func TestEvaluateSharedSessionBrowserBindingForScopeProjectsHealthAndCoordination(t *testing.T) {
	sessionID := "sess-binding-evaluation"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "work", Target: "node"}
	registry := NewBrowserSessionRegistry()
	tracked := registry.TrackTabs(sessionID, []BrowserSessionTarget{
		{ID: "tab-1", TabIndex: 1, URL: "https://example.com", Backend: "proxy", Profile: "work", Target: "node"},
	}, 1)

	stateRegistry := NewBrowserSessionStateRegistry()
	stateRegistry.SelectBrowserProfile(sessionID, SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "work",
		RuntimeTarget: "node",
		BrowserApp:    "chrome",
		Source:        "select_profile",
	})
	stateRegistry.RecordBrowserProfileState(sessionID, SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "work",
		RuntimeTarget: "node",
		BrowserApp:    "chrome",
		Status:        "running",
		Running:       true,
		Connected:     true,
	})

	evaluation := EvaluateSharedSessionBrowserBindingForScope(
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "work", Target: "node"},
		route,
		nil,
		registry,
		testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			sessionID: {{RunID: "run-1", Status: "running"}},
		}},
		stateRegistry,
		time.Minute,
	)

	if evaluation.Snapshot.SelectedProfileSelection == nil || evaluation.Snapshot.SelectedProfileSelection.Profile != "work" {
		t.Fatalf("expected selected profile in binding evaluation, got %#v", evaluation.Snapshot.SelectedProfileSelection)
	}
	if evaluation.Snapshot.SelectedTargetSelection == nil || evaluation.Snapshot.SelectedTargetSelection.ID != tracked[0].ID {
		t.Fatalf("expected selected target in binding evaluation, got %#v", evaluation.Snapshot.SelectedTargetSelection)
	}
	if evaluation.Health.Summary == nil || evaluation.Health.Summary.State != "healthy" {
		t.Fatalf("expected healthy binding health evaluation, got %#v", evaluation.Health)
	}
	if evaluation.Coordination.Plan.State != "coordinated" || !evaluation.Coordination.Plan.BrowserOnNode || !evaluation.Coordination.Plan.HasActiveNodeRun || !evaluation.Coordination.Plan.HasRunningBrowserProfile {
		t.Fatalf("unexpected binding coordination evaluation: %#v", evaluation.Coordination)
	}
}

func TestMergeSharedSessionBrowserBindingEvaluationProfileSnapshotRecomputesHealthAndCoordination(t *testing.T) {
	evaluation := SharedSessionBrowserBindingEvaluation{
		Routes: []SharedSessionBrowserRouteSnapshot{{
			Backend:       "proxy",
			Profile:       "work",
			RuntimeTarget: "node",
			Targets: []SharedSessionBrowserRouteTarget{{
				ID:            "tab-1",
				Backend:       "proxy",
				Profile:       "work",
				RuntimeTarget: "node",
				Current:       true,
			}},
		}},
		Snapshot: SharedSessionBrowserBindingSnapshot{
			Runs: []SharedSessionRunInfo{{RunID: "run-1", Status: "running"}},
			Summary: SharedSessionBrowserBindingSummary{
				RouteTargetCount: 1,
				NodeRunCount:     1,
				ActiveNodeRunID:  "run-1",
			},
		},
	}

	merged := MergeSharedSessionBrowserBindingEvaluationProfileSnapshot(
		nil,
		"",
		BrowserRuntimeInfo{Backend: "proxy", Profile: "work", Target: "node"},
		"work",
		evaluation,
		[]SharedSessionBrowserProfileState{{
			Backend:       "proxy",
			Profile:       "work",
			RuntimeTarget: "node",
			Status:        "running",
			Running:       true,
			Connected:     true,
		}},
		time.Minute,
	)

	if len(merged.Snapshot.Profiles) != 1 || merged.Snapshot.Summary.ActiveBrowserProfile != "work" {
		t.Fatalf("expected merged profile snapshot, got %#v", merged.Snapshot)
	}
	if merged.Health.Summary == nil || merged.Health.Summary.State != "healthy" {
		t.Fatalf("expected healthy merged binding evaluation, got %#v", merged.Health)
	}
	if merged.Coordination.Plan.State != "coordinated" || !merged.Coordination.Plan.HasRunningBrowserProfile {
		t.Fatalf("expected coordinated merged binding evaluation, got %#v", merged.Coordination)
	}
}

func TestObserveSharedSessionBrowserBindingForScopeProjectsObservationSnapshot(t *testing.T) {
	sessionID := "sess-binding-watch"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "work", Target: "node"}
	registry := NewBrowserSessionRegistry()
	registry.TrackTabs(sessionID, []BrowserSessionTarget{
		{ID: "tab-1", TabIndex: 1, URL: "https://example.com", Backend: "proxy", Profile: "work", Target: "node"},
	}, 1)
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:   "proxy",
			Profile:   "work",
			Status:    "running",
			Running:   true,
			Connected: true,
		},
		profilesResp: BrowserProfilesResult{
			Backend: "proxy",
			Profiles: []BrowserProfileInfo{{
				Profile:   "work",
				Status:    "running",
				Running:   true,
				Connected: true,
			}},
		},
	}

	observation := ObserveSharedSessionBrowserBindingForScope(
		context.Background(),
		backend,
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "work", Target: "node"},
		route,
		"work",
		true,
		true,
		registry,
		testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			sessionID: {{RunID: "run-1", Status: "running"}},
		}},
		nil,
		time.Minute,
	)

	if observation.Observation.Status == nil || observation.Observation.Profiles == nil {
		t.Fatalf("expected status and profiles observation, got %#v", observation.Observation)
	}
	if len(observation.Evaluation.Snapshot.Profiles) != 1 || observation.Evaluation.Snapshot.Summary.ActiveBrowserProfile != "work" {
		t.Fatalf("expected observation snapshot to project binding profiles, got %#v", observation.Evaluation.Snapshot)
	}
	if observation.Evaluation.Health.Summary == nil || observation.Evaluation.Health.Summary.State != "healthy" {
		t.Fatalf("expected healthy binding observation, got %#v", observation.Evaluation.Health)
	}
	if !observation.Evaluation.ReferenceTime.Equal(observation.Observation.ProfilesObservedAt) {
		t.Fatalf("expected binding observation to carry profiles observed_at as reference time, got %v want %v", observation.Evaluation.ReferenceTime, observation.Observation.ProfilesObservedAt)
	}
	if observation.Evaluation.Coordination.Plan.State != "coordinated" || !observation.Evaluation.Coordination.Plan.HasRunningBrowserProfile {
		t.Fatalf("expected coordinated binding observation, got %#v", observation.Evaluation.Coordination)
	}
}

func TestObserveSharedSessionBrowserBindingForScopeInvalidatesManagedCurrentTargetBeforeSnapshot(t *testing.T) {
	sessionID := "sess-binding-watch-invalidate"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "work", Target: "node"}
	registry := NewBrowserSessionRegistry()
	registry.TrackTabs(sessionID, []BrowserSessionTarget{
		{ID: "tab-1", TabIndex: 1, URL: "https://example.com", Backend: "proxy", Profile: "work", Target: "node"},
	}, 1)
	stateRegistry := NewBrowserSessionStateRegistry()
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:    "proxy",
			Profile:    "work",
			BrowserApp: "Chromium",
			Status:     "disconnected",
			Running:    true,
			Connected:  false,
		},
	}

	observation := ObserveSharedSessionBrowserBindingForScope(
		context.Background(),
		backend,
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "work", Target: "node"},
		route,
		"work",
		true,
		false,
		registry,
		nil,
		stateRegistry,
		time.Minute,
	)

	if !observation.Observation.HasSyncedState || observation.Observation.SyncedState.Status != "disconnected" {
		t.Fatalf("expected synced disconnected state, got %#v", observation.Observation)
	}
	if observation.Evaluation.Snapshot.CurrentTargetID != "" || observation.Evaluation.Snapshot.Summary.CurrentTargetID != "" {
		t.Fatalf("expected binding snapshot to reflect invalidated current target, got %#v", observation.Evaluation.Snapshot)
	}
	if len(observation.Evaluation.Routes) != 1 || observation.Evaluation.Routes[0].CurrentTargetID != "" {
		t.Fatalf("expected route snapshot to clear current target after invalidation, got %#v", observation.Evaluation.Routes)
	}
	if _, ok := registry.CurrentTargetForRoute(sessionID, route); ok {
		t.Fatalf("expected managed current target selection to be invalidated")
	}
	if target, ok := registry.ResolveTabForRoute(sessionID, route, 1); !ok || strings.TrimSpace(target.ID) == "" {
		t.Fatalf("expected tracked tab to remain after invalidation, got %#v ok=%v", target, ok)
	}
}

func TestSharedSessionBrowserWatchManagerObserveBindingSeedsSiblingProviderWatchLoopAfterLifecycleInvalidation(t *testing.T) {
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:    "proxy",
			Profile:    "isolated",
			BrowserApp: "Chromium",
			Status:     "running",
			Running:    true,
			Connected:  true,
		},
	}
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionID := "sess-binding-sibling-provider-watchloop-cache"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
	sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://example.com/one",
		Title:      "One",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, true)
	stateRegistry.SelectBrowserProfile(sessionID, SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Source:        "remember_profile",
	})

	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			sessionID: {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			sessionID: {{RunID: "run-b", Status: "running"}},
		}},
	}
	managerA := SharedSessionBrowserObserverManagerFor(sessionRegistry, runRegistryA, stateRegistry, time.Minute)
	managerB := SharedSessionBrowserObserverManagerFor(sessionRegistry, runRegistryB, stateRegistry, time.Minute)
	boundA := managerA.Bind(backend)
	boundB := managerB.Bind(backend)
	watchReq := SharedSessionBrowserObserverRequest{
		SessionID:                   sessionID,
		SelectedInfo:                BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		BindingRoute:                route,
		RequestedProfile:            "isolated",
		IncludeStatus:               true,
		IncludeProfiles:             false,
		IncludeSessionView:          true,
		SessionViewInfo:             BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		SessionViewRouteFilter:      route,
		SessionViewRequestedProfile: "isolated",
	}

	initialA := boundA.ObserveWatchLoop(context.Background(), watchReq)
	initialB := boundB.ObserveWatchLoop(context.Background(), watchReq)
	if len(initialA.View.Session.Routes) != 1 || initialA.View.Session.Routes[0].CurrentTargetID == "" {
		t.Fatalf("expected first provider watch loop to retain current target, got %#v", initialA.View.Session.Routes)
	}
	if len(initialB.View.Session.Routes) != 1 || initialB.View.Session.Routes[0].CurrentTargetID == "" {
		t.Fatalf("expected second provider watch loop to retain current target, got %#v", initialB.View.Session.Routes)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 0 {
		t.Fatalf("expected sibling watch loops to poll RuntimeStatus once each, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistryA.callCount() != 1 || runRegistryB.callCount() != 1 {
		t.Fatalf("expected initial sibling watch loops to snapshot runs once each, got runA=%d runB=%d", runRegistryA.callCount(), runRegistryB.callCount())
	}

	backend.statusResp = BrowserProfileStatusResult{
		Backend:    "proxy",
		Profile:    "isolated",
		BrowserApp: "Chromium",
		Status:     "disconnected",
		Running:    true,
		Connected:  false,
	}
	managerA.Invalidate()
	bindingReq := SharedSessionBrowserObserverRequest{
		SessionID:        sessionID,
		SelectedInfo:     BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		BindingRoute:     route,
		RequestedProfile: "isolated",
		IncludeStatus:    true,
		IncludeProfiles:  false,
	}
	observation := boundA.ObserveBinding(context.Background(), bindingReq)
	if !observation.Observation.HasSyncedState || observation.Observation.SyncedState.Status != "disconnected" {
		t.Fatalf("expected binding observation to sync disconnected lifecycle state, got %#v", observation.Observation)
	}
	if observation.Evaluation.Snapshot.CurrentTargetID != "" || observation.Evaluation.Snapshot.Summary.CurrentTargetID != "" {
		t.Fatalf("expected binding observation to clear current target after invalidation, got %#v", observation.Evaluation.Snapshot)
	}
	if runRegistryA.callCount() != 2 || runRegistryB.callCount() != 2 {
		t.Fatalf("expected binding invalidation to refresh both sibling provider projections once, got runA=%d runB=%d", runRegistryA.callCount(), runRegistryB.callCount())
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), watchReq)
	if len(seededB.View.Session.Routes) != 1 || seededB.View.Session.Routes[0].CurrentTargetID != "" {
		t.Fatalf("expected sibling watch loop to reuse invalidated projection cache, got %#v", seededB.View.Session.Routes)
	}
	if len(backend.statusReqs) != 3 || len(backend.profilesReqs) != 0 {
		t.Fatalf("expected sibling watch loop to reuse binding-seeded source without extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistryA.callCount() != 2 || runRegistryB.callCount() != 2 {
		t.Fatalf("expected sibling watch loop to reuse binding-seeded projection cache without extra rebuilds, got runA=%d runB=%d", runRegistryA.callCount(), runRegistryB.callCount())
	}
}

func TestMergeSharedSessionBrowserBindingEvaluationProfileSnapshotUsesObservedSnapshotAsHealthReferenceTime(t *testing.T) {
	evaluation := SharedSessionBrowserBindingEvaluation{
		Snapshot: SharedSessionBrowserBindingSnapshot{
			Summary: SharedSessionBrowserBindingSummary{
				RouteTargetCount: 1,
			},
		},
	}
	staleAt := time.Date(2026, time.March, 29, 12, 0, 0, 0, time.UTC).Add(-2 * time.Minute)
	merged := MergeSharedSessionBrowserBindingEvaluationProfileSnapshot(
		nil,
		"",
		BrowserRuntimeInfo{Backend: "proxy", Profile: "work", Target: "node"},
		"work",
		evaluation,
		[]SharedSessionBrowserProfileState{{
			Backend:       "proxy",
			Profile:       "work",
			RuntimeTarget: "node",
			BrowserApp:    "chrome",
			Status:        "reconnecting",
			Running:       true,
			Connected:     false,
			ObservedAt:    staleAt,
			StatusSince:   staleAt,
			Note:          "cdp reconnect in progress",
		}},
		time.Minute,
	)

	if merged.Health.Summary == nil || merged.Health.Summary.State != "profile_reconnecting" {
		t.Fatalf("expected reconnecting binding health evaluation, got %#v", merged.Health)
	}
	if !merged.ReferenceTime.Equal(staleAt) {
		t.Fatalf("expected merged binding evaluation to keep observed snapshot time as reference, got %v want %v", merged.ReferenceTime, staleAt)
	}
	if merged.Health.ReconnectTimedOut || merged.Health.Summary.RecoveryAction != "" {
		t.Fatalf("expected merged binding snapshot to use observed snapshot time instead of wall clock, got %#v", merged.Health)
	}
}

func TestEvaluateSharedSessionBrowserBindingForScopeDoesNotInventReferenceTime(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	registry.RecordBrowserProfileState("sess-static-binding", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "work",
		RuntimeTarget: "node",
		BrowserApp:    "chrome",
		Status:        "reconnecting",
		Running:       true,
		Connected:     false,
		ObservedAt:    time.Date(2026, time.March, 29, 12, 0, 0, 0, time.UTC).Add(-2 * time.Minute),
		StatusSince:   time.Date(2026, time.March, 29, 12, 0, 0, 0, time.UTC).Add(-2 * time.Minute),
	})

	evaluation := EvaluateSharedSessionBrowserBindingForScope(
		"sess-static-binding",
		BrowserRuntimeInfo{Backend: "proxy", Profile: "work", Target: "node"},
		BrowserSessionRoute{Backend: "proxy", Profile: "work", Target: "node"},
		nil,
		nil,
		nil,
		registry,
		time.Minute,
	)

	if !evaluation.ReferenceTime.IsZero() {
		t.Fatalf("expected static binding evaluation without fresh watch cycle to keep zero reference time, got %v", evaluation.ReferenceTime)
	}
}

func TestMergeSharedSessionBrowserBindingEvaluationProfileStateRecomputesSnapshotAndHealth(t *testing.T) {
	evaluation := SharedSessionBrowserBindingEvaluation{
		Routes: []SharedSessionBrowserRouteSnapshot{{
			Backend:       "proxy",
			Profile:       "work",
			RuntimeTarget: "node",
			Targets: []SharedSessionBrowserRouteTarget{{
				ID:            "tab-1",
				Backend:       "proxy",
				Profile:       "work",
				RuntimeTarget: "node",
				Current:       true,
			}},
		}},
		Snapshot: SharedSessionBrowserBindingSnapshot{
			Runs: []SharedSessionRunInfo{{RunID: "run-1", Status: "running"}},
			Summary: SharedSessionBrowserBindingSummary{
				RouteTargetCount: 1,
				NodeRunCount:     1,
				ActiveNodeRunID:  "run-1",
			},
		},
	}

	merged := MergeSharedSessionBrowserBindingEvaluationProfileState(
		nil,
		"",
		BrowserRuntimeInfo{Backend: "proxy", Profile: "work", Target: "node"},
		"work",
		evaluation,
		SharedSessionBrowserProfileState{
			Status:     "running",
			Running:    true,
			Connected:  true,
			BrowserApp: "Arc",
		},
		time.Minute,
	)

	if len(merged.Snapshot.Profiles) != 1 {
		t.Fatalf("expected merged snapshot profile, got %#v", merged.Snapshot.Profiles)
	}
	if merged.Snapshot.Profiles[0].Profile != "work" || merged.Snapshot.Profiles[0].RuntimeTarget != "node" {
		t.Fatalf("expected selected route identity fallback, got %#v", merged.Snapshot.Profiles[0])
	}
	if merged.Snapshot.Summary.BrowserProfileCount != 1 || merged.Snapshot.Summary.ActiveBrowserProfile != "work" {
		t.Fatalf("expected recomputed browser profile summary, got %#v", merged.Snapshot.Summary)
	}
	if merged.Health.Summary == nil || merged.Health.Summary.State != "healthy" {
		t.Fatalf("expected healthy merged evaluation, got %#v", merged.Health)
	}
	if !merged.Coordination.Plan.HasRunningBrowserProfile {
		t.Fatalf("expected coordination to reflect merged running profile, got %#v", merged.Coordination.Plan)
	}
}
