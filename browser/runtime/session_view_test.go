package browserruntime

import (
	"context"
	"testing"
	"time"
)

func TestSnapshotSharedSessionBrowserSessionViewProjectsRoutesRunsAndProfiles(t *testing.T) {
	sessionID := "sess-view"
	registry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	runRegistry := testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
		sessionID: {{RunID: "run-1", Status: "running"}},
	}}

	registry.TrackTabs(sessionID, []BrowserSessionTarget{
		{ID: "tab-1", TabIndex: 1, URL: "https://example.com/1", Backend: "proxy", Profile: "work", Target: "node", BrowserApp: "Chromium"},
		{ID: "tab-2", TabIndex: 2, URL: "https://example.com/2", Backend: "proxy", Profile: "work", Target: "node", BrowserApp: "Chromium"},
	}, 2)
	stateRegistry.SelectBrowserProfile(sessionID, SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "work",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Source:        "select_profile",
	})
	stateRegistry.RecordBrowserProfileState(sessionID, SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "work",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "running",
		Running:       true,
		Connected:     true,
	})

	view := SnapshotSharedSessionBrowserSessionView(
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "work", Target: "node"},
		"work",
		BrowserSessionRoute{Backend: "proxy", Profile: "work", Target: "node"},
		registry,
		runRegistry,
		stateRegistry,
	)

	if len(view.Routes) != 1 || len(view.Routes[0].Targets) != 2 {
		t.Fatalf("expected route-scoped targets in session view, got %#v", view.Routes)
	}
	if view.TargetCount != 2 {
		t.Fatalf("expected target count from session view, got %#v", view)
	}
	if len(view.Runs) != 1 || view.Runs[0].RunID != "run-1" {
		t.Fatalf("expected run snapshot from session view, got %#v", view.Runs)
	}
	if len(view.Profiles) != 1 || view.Profiles[0].State.Profile != "work" || !view.Profiles[0].Selected {
		t.Fatalf("expected projected profile snapshot from session view, got %#v", view.Profiles)
	}
	if view.Handoff == nil ||
		view.Handoff.State != SharedSessionBrowserSessionHandoffStateReady ||
		view.Handoff.TargetCount != 2 ||
		view.Handoff.ActiveRunID != "run-1" ||
		view.Handoff.SelectedProfile != "work" {
		t.Fatalf("expected compact handoff summary from session view, got %#v", view.Handoff)
	}
}

func TestObserveSharedSessionBrowserViewForScopeReusesBindingObservationAndSessionView(t *testing.T) {
	sessionID := "sess-view-watch"
	registry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	runRegistry := testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
		sessionID: {{RunID: "run-1", Status: "running"}},
	}}
	registry.TrackTabs(sessionID, []BrowserSessionTarget{
		{ID: "tab-1", TabIndex: 1, URL: "https://example.com/1", Backend: "proxy", Profile: "work", Target: "node", BrowserApp: "Chromium"},
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
				Profile:    "work",
				BrowserApp: "Chromium",
				Status:     "running",
				Running:    true,
				Connected:  true,
			}},
		},
	}

	view := ObserveSharedSessionBrowserViewForScope(
		context.Background(),
		backend,
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "work", Target: "node"},
		BrowserSessionRoute{Backend: "proxy", Profile: "work", Target: "node"},
		"work",
		true,
		true,
		true,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "work", Target: "node"},
		BrowserSessionRoute{Backend: "proxy", Profile: "work", Target: "node"},
		"work",
		registry,
		runRegistry,
		stateRegistry,
		time.Minute,
	)

	if view.Observation.Status == nil || view.Observation.Profiles == nil {
		t.Fatalf("expected binding watch cycle to preserve status/profiles observation, got %#v", view.Observation)
	}
	if len(view.Binding.Snapshot.Runs) != 1 || view.Binding.Snapshot.Summary.ActiveNodeRunID != "run-1" {
		t.Fatalf("expected binding evaluation to include session runs, got %#v", view.Binding)
	}
	if len(view.Session.Routes) != 1 || view.Session.TargetCount != 1 {
		t.Fatalf("expected session view snapshot, got %#v", view.Session)
	}
	if len(view.Session.Profiles) != 1 || view.Session.Profiles[0].State.Profile != "work" {
		t.Fatalf("expected projected session profiles from shared view, got %#v", view.Session.Profiles)
	}
	if view.Session.Handoff == nil || view.Session.Handoff.State != SharedSessionBrowserSessionHandoffStateReady || view.Session.Handoff.ActiveRunID != "run-1" {
		t.Fatalf("expected reused binding session view to keep handoff summary, got %#v", view.Session.Handoff)
	}
}

func TestProjectSharedSessionBrowserTopLevelSessionViewFallsBackToProjectedProfiles(t *testing.T) {
	projection := ProjectSharedSessionBrowserTopLevelSessionView(
		SharedSessionBrowserSessionViewSnapshot{
			Routes: []SharedSessionBrowserRouteSnapshot{{
				Backend:       "proxy",
				Profile:       "workbench",
				RuntimeTarget: "node",
				Targets:       []SharedSessionBrowserRouteTarget{{ID: "tab-1"}},
			}},
			Runs: []SharedSessionRunInfo{{RunID: "run-1", Status: "running"}},
		},
		[]SharedSessionBrowserProjectedProfileState{{
			State: SharedSessionBrowserProfileState{
				Backend:       "proxy",
				Profile:       "workbench",
				RuntimeTarget: "node",
				BrowserApp:    "Chromium",
				Status:        "running",
				Running:       true,
				Connected:     true,
			},
			Selected: true,
		}},
	)

	if projection.TargetCount != 1 {
		t.Fatalf("expected top-level session projection to derive target count from routes, got %#v", projection)
	}
	if len(projection.Runs) != 1 || projection.Runs[0].RunID != "run-1" {
		t.Fatalf("expected top-level session projection to keep runs, got %#v", projection.Runs)
	}
	if len(projection.Profiles) != 1 || projection.Profiles[0].State.Profile != "workbench" || !projection.Profiles[0].Selected {
		t.Fatalf("expected top-level session projection to fall back to supplied profiles, got %#v", projection.Profiles)
	}
	if projection.Handoff == nil || projection.Handoff.TargetCount != 1 || projection.Handoff.SelectedProfile != "workbench" {
		t.Fatalf("expected top-level session projection to derive handoff from fallback profiles, got %#v", projection.Handoff)
	}
}

func TestProjectSharedSessionBrowserTopLevelSessionFromBindingEvaluationSynthesizesSelectionRoute(t *testing.T) {
	projection := ProjectSharedSessionBrowserTopLevelSessionFromBindingEvaluation(
		SharedSessionBrowserBindingEvaluation{
			Snapshot: SharedSessionBrowserBindingSnapshot{
				CurrentTargetID: "tab-2",
				SelectedProfileSelection: &SharedSessionBrowserProfileSelection{
					Backend:       "proxy",
					Profile:       "workbench",
					RuntimeTarget: "node",
					BrowserApp:    "Chromium",
					Source:        "remember_profile",
				},
				SelectedTargetSelection: &BrowserSessionTargetSelection{
					ID:            "tab-2",
					Backend:       "proxy",
					Profile:       "workbench",
					RuntimeTarget: "node",
					BrowserApp:    "Chromium",
					Source:        "tracked_active_tab",
				},
				Runs: []SharedSessionRunInfo{{RunID: "run-1", Status: "running"}},
				Profiles: []SharedSessionBrowserProfileState{{
					Backend:       "proxy",
					Profile:       "workbench",
					RuntimeTarget: "node",
					BrowserApp:    "Chromium",
					Status:        "running",
					Running:       true,
					Connected:     true,
				}},
			},
		},
	)

	if projection.TargetCount != 1 {
		t.Fatalf("expected synthesized session projection to infer target count, got %#v", projection)
	}
	if len(projection.Routes) != 1 ||
		projection.Routes[0].Backend != "proxy" ||
		projection.Routes[0].Profile != "workbench" ||
		projection.Routes[0].RuntimeTarget != "node" ||
		projection.Routes[0].BrowserApp != "Chromium" ||
		projection.Routes[0].CurrentTargetID != "tab-2" ||
		projection.Routes[0].CurrentTargetSource != "tracked_active_tab" {
		t.Fatalf("expected binding-evaluation session projection to synthesize minimal route snapshot, got %#v", projection.Routes)
	}
	if len(projection.Runs) != 1 || projection.Runs[0].RunID != "run-1" {
		t.Fatalf("expected binding-evaluation session projection to keep runs, got %#v", projection.Runs)
	}
	if len(projection.Profiles) != 1 || projection.Profiles[0].State.Profile != "workbench" || !projection.Profiles[0].Selected {
		t.Fatalf("expected binding-evaluation session projection to keep projected profiles, got %#v", projection.Profiles)
	}
	if projection.Handoff == nil ||
		projection.Handoff.State != SharedSessionBrowserSessionHandoffStateReady ||
		projection.Handoff.CurrentTarget == nil ||
		projection.Handoff.CurrentTarget.ID != "tab-2" {
		t.Fatalf("expected binding-evaluation session projection to include handoff, got %#v", projection.Handoff)
	}
}

func TestProjectSharedSessionBrowserTopLevelSessionFromBindingEvaluationFallsBackToRouteScopedProfiles(t *testing.T) {
	projection := ProjectSharedSessionBrowserTopLevelSessionFromBindingEvaluation(
		SharedSessionBrowserBindingEvaluation{
			Routes: []SharedSessionBrowserRouteSnapshot{{
				Backend:         "proxy",
				Profile:         "workbench",
				RuntimeTarget:   "node",
				BrowserApp:      "Chromium",
				CurrentTargetID: "tab-2",
				Targets: []SharedSessionBrowserRouteTarget{{
					ID:         "tab-2",
					BrowserApp: "Chromium",
				}},
			}},
			Snapshot: SharedSessionBrowserBindingSnapshot{
				CurrentTargetID: "tab-2",
				SelectedProfileSelection: &SharedSessionBrowserProfileSelection{
					Backend:       "proxy",
					Profile:       "workbench",
					RuntimeTarget: "node",
					Source:        "remember_profile",
				},
				SelectedTargetSelection: &BrowserSessionTargetSelection{
					ID:            "tab-2",
					Backend:       "proxy",
					Profile:       "workbench",
					RuntimeTarget: "node",
					Source:        "tracked_active_tab",
				},
			},
		},
	)

	if len(projection.Profiles) != 1 || projection.Profiles[0].State.Profile != "workbench" || !projection.Profiles[0].Selected {
		t.Fatalf("expected binding-evaluation session projection to synthesize fallback profiles from route snapshots, got %#v", projection.Profiles)
	}
	if projection.Profiles[0].State.Note != "cached route-scoped session snapshot" {
		t.Fatalf("expected binding-evaluation fallback profile note, got %#v", projection.Profiles[0])
	}
}

func TestSharedSessionBrowserSessionViewSnapshotFromBindingSynthesizesSelectionRoute(t *testing.T) {
	view := sharedSessionBrowserSessionViewSnapshotFromBinding(
		SharedSessionBrowserBindingEvaluation{
			Snapshot: SharedSessionBrowserBindingSnapshot{
				CurrentTargetID: "tab-4",
				SelectedProfileSelection: &SharedSessionBrowserProfileSelection{
					Backend:       "proxy",
					Profile:       "workbench",
					RuntimeTarget: "node",
					BrowserApp:    "Chromium",
					Source:        "remember_profile",
				},
				SelectedTargetSelection: &BrowserSessionTargetSelection{
					ID:            "tab-4",
					Backend:       "proxy",
					Profile:       "workbench",
					RuntimeTarget: "node",
					BrowserApp:    "Chromium",
					Source:        "tracked_active_tab",
				},
				Runs: []SharedSessionRunInfo{{RunID: "run-1", Status: "running"}},
				Profiles: []SharedSessionBrowserProfileState{{
					Backend:       "proxy",
					Profile:       "workbench",
					RuntimeTarget: "node",
					BrowserApp:    "Chromium",
					Status:        "running",
					Running:       true,
					Connected:     true,
				}},
			},
		},
	)

	if view.TargetCount != 1 {
		t.Fatalf("expected binding-backed session view to infer target count, got %#v", view)
	}
	if len(view.Routes) != 1 ||
		view.Routes[0].Backend != "proxy" ||
		view.Routes[0].Profile != "workbench" ||
		view.Routes[0].RuntimeTarget != "node" ||
		view.Routes[0].BrowserApp != "Chromium" ||
		view.Routes[0].CurrentTargetID != "tab-4" ||
		view.Routes[0].CurrentTargetSource != "tracked_active_tab" {
		t.Fatalf("expected binding-backed session view to synthesize minimal route snapshot, got %#v", view.Routes)
	}
	if len(view.Runs) != 1 || view.Runs[0].RunID != "run-1" {
		t.Fatalf("expected binding-backed session view to keep runs, got %#v", view.Runs)
	}
	if len(view.Profiles) != 1 || view.Profiles[0].State.Profile != "workbench" || !view.Profiles[0].Selected {
		t.Fatalf("expected binding-backed session view to keep projected profiles, got %#v", view.Profiles)
	}
	if view.Handoff == nil ||
		view.Handoff.State != SharedSessionBrowserSessionHandoffStateReady ||
		view.Handoff.CurrentTarget == nil ||
		view.Handoff.CurrentTarget.ID != "tab-4" {
		t.Fatalf("expected binding-backed session view to include handoff, got %#v", view.Handoff)
	}
}

func TestSharedSessionBrowserSessionViewSnapshotFromBindingFallsBackToRouteScopedProfiles(t *testing.T) {
	view := sharedSessionBrowserSessionViewSnapshotFromBinding(
		SharedSessionBrowserBindingEvaluation{
			Routes: []SharedSessionBrowserRouteSnapshot{{
				Backend:         "proxy",
				Profile:         "workbench",
				RuntimeTarget:   "node",
				BrowserApp:      "Chromium",
				CurrentTargetID: "tab-4",
				Targets: []SharedSessionBrowserRouteTarget{{
					ID:         "tab-4",
					BrowserApp: "Chromium",
				}},
			}},
			Snapshot: SharedSessionBrowserBindingSnapshot{
				CurrentTargetID: "tab-4",
				SelectedProfileSelection: &SharedSessionBrowserProfileSelection{
					Backend:       "proxy",
					Profile:       "workbench",
					RuntimeTarget: "node",
					Source:        "remember_profile",
				},
				SelectedTargetSelection: &BrowserSessionTargetSelection{
					ID:            "tab-4",
					Backend:       "proxy",
					Profile:       "workbench",
					RuntimeTarget: "node",
					Source:        "tracked_active_tab",
				},
			},
		},
	)

	if len(view.Profiles) != 1 || view.Profiles[0].State.Profile != "workbench" || !view.Profiles[0].Selected {
		t.Fatalf("expected binding-backed session view to synthesize fallback profiles from route snapshots, got %#v", view.Profiles)
	}
	if view.Profiles[0].State.Note != "cached route-scoped session snapshot" {
		t.Fatalf("expected binding-backed session view fallback profile note, got %#v", view.Profiles[0])
	}
}

func TestProjectSharedSessionBrowserFallbackProfilesFromRouteSnapshotsSynthesizesRouteScopedProfiles(t *testing.T) {
	profiles := ProjectSharedSessionBrowserFallbackProfilesFromRouteSnapshots(
		BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
		"",
		[]SharedSessionBrowserRouteSnapshot{
			{
				Backend:       "proxy",
				Profile:       "workbench",
				RuntimeTarget: "node",
				Targets:       []SharedSessionBrowserRouteTarget{{ID: "tab-1", BrowserApp: "Chromium"}},
			},
			{
				Backend:       "proxy",
				Profile:       "isolated",
				RuntimeTarget: "node",
				Targets:       []SharedSessionBrowserRouteTarget{{ID: "tab-2", BrowserApp: "Chromium"}},
			},
		},
		&SharedSessionBrowserProfileSelection{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
			Source:        "select_profile",
		},
	)

	if len(profiles) != 2 {
		t.Fatalf("expected fallback route-snapshot projection to retain both managed profiles, got %#v", profiles)
	}
	if !sharedSessionBrowserProjectedProfileSelected(profiles, "workbench") {
		t.Fatalf("expected fallback route-snapshot projection to keep remembered profile selection, got %#v", profiles)
	}
	foundIsolated := false
	for _, item := range profiles {
		if item.State.Profile == "isolated" {
			foundIsolated = true
			if item.State.Note != "cached route-scoped session snapshot" {
				t.Fatalf("expected synthesized fallback profile note, got %#v", item)
			}
		}
	}
	if !foundIsolated {
		t.Fatalf("expected fallback route-snapshot projection to include sibling profile, got %#v", profiles)
	}
}
