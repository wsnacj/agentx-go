package browserruntime

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestObserveSharedSessionBrowserWatchForScopeProjectsSyncedProfilesAndSessionView(t *testing.T) {
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Profile:   "isolated",
			Status:    "running",
			Running:   true,
			Connected: true,
		},
		profilesResp: BrowserProfilesResult{
			Backend:        "proxy",
			DefaultProfile: "isolated",
			Note:           "profiles ok",
			Profiles: []BrowserProfileInfo{
				{Profile: "isolated", BrowserApp: "Chromium", Status: "connected", Running: true, Connected: true},
			},
		},
	}
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionID := "sess-inspection"
	runRegistry := testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
		sessionID: {{RunID: "run-1", Status: "running"}},
	}}

	sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://example.com",
		Title:      "Example",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
		BrowserApp: "Chromium",
	}, true)
	stateRegistry.SelectBrowserProfile(sessionID, SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Source:        "remember_profile",
	})

	watch := ObserveSharedSessionBrowserWatchForScope(
		context.Background(),
		backend,
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"},
		"isolated",
		true,
		true,
		true,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"},
		"isolated",
		sessionRegistry,
		runRegistry,
		stateRegistry,
		time.Minute,
	)

	if watch.View.Observation.Status == nil || watch.View.Observation.Profiles == nil {
		t.Fatalf("expected status and profiles observation, got %#v", watch.View.Observation)
	}
	if len(watch.Profiles) != 1 || watch.Profiles[0].State.Profile != "isolated" || !watch.Profiles[0].Selected {
		t.Fatalf("expected projected synced scoped profiles, got %#v", watch.Profiles)
	}
	if len(watch.DiscoveredProfiles) != 1 || watch.DiscoveredProfiles[0] != "isolated" {
		t.Fatalf("expected discovered profiles, got %#v", watch.DiscoveredProfiles)
	}
	if watch.DefaultProfile != "isolated" || watch.Note != "profiles ok" {
		t.Fatalf("expected default profile and note projection, got %#v", watch)
	}
	if len(watch.View.Session.Routes) != 1 || len(watch.View.Session.Runs) != 1 || len(watch.View.Session.Profiles) != 1 {
		t.Fatalf("expected watch to include session view snapshot, got %#v", watch.View.Session)
	}
	if !watch.ReferenceTime.Equal(watch.View.Binding.ReferenceTime) {
		t.Fatalf("expected watch reference time to follow binding evaluation, got %v want %v", watch.ReferenceTime, watch.View.Binding.ReferenceTime)
	}
}

func TestObserveSharedSessionBrowserWatchForScopeFallsBackToObservedProfilesWithoutRegistry(t *testing.T) {
	backend := &statusProfilesObservationTestBackend{
		profilesResp: BrowserProfilesResult{
			Backend:        "proxy",
			DefaultProfile: "isolated",
			Profiles: []BrowserProfileInfo{
				{Profile: "isolated", BrowserApp: "Chromium", Status: "running", Running: true},
			},
		},
	}

	watch := ObserveSharedSessionBrowserWatchForScope(
		context.Background(),
		backend,
		"",
		BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"},
		"isolated",
		false,
		true,
		false,
		BrowserRuntimeInfo{},
		BrowserSessionRoute{},
		"",
		nil,
		nil,
		nil,
		time.Minute,
	)

	if watch.View.Observation.Profiles == nil {
		t.Fatalf("expected profiles observation")
	}
	if len(watch.Profiles) != 1 || watch.Profiles[0].State.Profile != "isolated" || watch.Profiles[0].Selected {
		t.Fatalf("expected observed profiles fallback without registry selection, got %#v", watch.Profiles)
	}
	if watch.DefaultProfile != "isolated" {
		t.Fatalf("expected default profile fallback, got %#v", watch)
	}
}

func TestBuildSharedSessionBrowserInspectionObserverRequestClearsTargetlessWorkbenchSessionScope(t *testing.T) {
	req := BuildSharedSessionBrowserInspectionObserverRequest(SharedSessionBrowserInspectionActionRequest{
		Action:             "workbench",
		SessionID:          "sess-inspect-workbench",
		SelectedInfo:       BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		BindingRoute:       BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node"},
		RequestedProfile:   "workbench",
		IncludeStatus:      true,
		IncludeProfiles:    true,
		IncludeSessionView: true,
	})
	if !req.IncludeStatus || !req.IncludeProfiles || !req.IncludeSessionView {
		t.Fatalf("expected workbench inspection request to include status/profiles/session view, got %#v", req)
	}
	if req.SessionViewInfo != (BrowserRuntimeInfo{}) || req.SessionViewRouteFilter != (BrowserSessionRoute{}) || req.SessionViewRequestedProfile != "" {
		t.Fatalf("expected targetless workbench inspection request to clear session-view scope, got %#v", req)
	}
	if req.BindingRoute != (BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node"}) {
		t.Fatalf("expected workbench inspection request to preserve binding route scope, got %#v", req.BindingRoute)
	}
}

func TestBuildSharedSessionBrowserInspectionObserverRequestClearsImplicitProfileScopeForProfiles(t *testing.T) {
	req := BuildSharedSessionBrowserInspectionObserverRequest(SharedSessionBrowserInspectionActionRequest{
		Action:                   "profiles",
		SessionID:                "sess-inspect-profiles",
		SelectedInfo:             BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		BindingRoute:             BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node"},
		RequestedProfile:         "workbench",
		IncludeProfiles:          true,
		ExplicitRequestedProfile: false,
	})
	if !req.IncludeProfiles {
		t.Fatalf("expected profiles inspection request to include profiles, got %#v", req)
	}
	if req.BindingRoute.Profile != "" {
		t.Fatalf("expected profiles inspection request without explicit requested profile to clear binding profile scope, got %#v", req.BindingRoute)
	}
	if req.RequestedProfile != "workbench" {
		t.Fatalf("expected profiles inspection request to fall back to selected profile, got %#v", req)
	}
}

func TestBuildSharedSessionBrowserInspectionActionRequestNormalizesInput(t *testing.T) {
	req := BuildSharedSessionBrowserInspectionActionRequest(SharedSessionBrowserInspectionActionInput{
		Action:                   " Workbench ",
		SessionID:                " sess-inspect-build ",
		SelectedInfo:             BrowserRuntimeInfo{Backend: " proxy ", Profile: " workbench ", Target: " node "},
		BindingRoute:             BrowserSessionRoute{Backend: " proxy ", Profile: " workbench ", Target: " node "},
		RequestedProfile:         " workbench ",
		ExplicitRequestedProfile: true,
		ExplicitSessionScope:     true,
		IncludeStatus:            true,
		IncludeProfiles:          true,
		IncludeSessionView:       true,
	})

	if req.Action != "workbench" || req.SessionID != "sess-inspect-build" {
		t.Fatalf("expected normalized action/session id, got %#v", req)
	}
	if req.SelectedInfo != (BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}) {
		t.Fatalf("expected normalized runtime info, got %#v", req.SelectedInfo)
	}
	if req.BindingRoute != (BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node"}) {
		t.Fatalf("expected normalized binding route, got %#v", req.BindingRoute)
	}
	if req.RequestedProfile != "workbench" || !req.ExplicitRequestedProfile || !req.ExplicitSessionScope {
		t.Fatalf("expected normalized requested profile and flags, got %#v", req)
	}
	if !req.IncludeStatus || !req.IncludeProfiles || !req.IncludeSessionView {
		t.Fatalf("expected include flags preserved, got %#v", req)
	}
}

func TestProjectSharedSessionBrowserInspectionActionProfilesUsesScopedFallbackWithoutExplicitRequest(t *testing.T) {
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionID := "sess-inspect-profiles-fallback"
	stateRegistry.SelectBrowserProfile(sessionID, SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "workbench",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Source:        "select_profile",
	})
	stateRegistry.RecordBrowserProfileState(sessionID, SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "workbench",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "running",
		Running:       true,
		Connected:     true,
	})
	stateRegistry.RecordBrowserProfileState(sessionID, SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "stopped",
	})

	projection := ProjectSharedSessionBrowserInspectionAction(
		SharedSessionBrowserInspectionActionRequest{
			Action:                   "profiles",
			SessionID:                sessionID,
			SelectedInfo:             BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
			BindingRoute:             BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node"},
			RequestedProfile:         "workbench",
			IncludeProfiles:          true,
			ExplicitRequestedProfile: false,
		},
		SharedSessionBrowserWatchObservation{
			View: SharedSessionBrowserViewObservation{
				Observation: SharedSessionBrowserStatusAndProfilesObservation{
					Profiles: &BrowserProfilesResult{
						Backend:        "proxy",
						DefaultProfile: "workbench",
						Profiles: []BrowserProfileInfo{
							{Profile: "workbench", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
						},
					},
				},
			},
			Profiles: []SharedSessionBrowserProjectedProfileState{{
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
			DiscoveredProfiles: []string{"workbench"},
			DefaultProfile:     "workbench",
			Note:               "profiles ok",
		},
		stateRegistry,
	)

	if !projection.HasProfiles {
		t.Fatalf("expected profiles projection, got %#v", projection)
	}
	if len(projection.Profiles) != 2 {
		t.Fatalf("expected scoped fallback to include full route-scoped registry snapshot, got %#v", projection.Profiles)
	}
	if !sharedSessionBrowserProjectedProfileSelected(projection.Profiles, "workbench") {
		t.Fatalf("expected scoped fallback to keep remembered profile selection, got %#v", projection.Profiles)
	}
	foundIsolated := false
	for _, item := range projection.Profiles {
		if item.State.Profile == "isolated" {
			foundIsolated = true
		}
	}
	if !foundIsolated {
		t.Fatalf("expected scoped fallback to include sibling managed profile, got %#v", projection.Profiles)
	}
	if projection.Note != "profiles ok" || projection.DefaultProfile != "workbench" {
		t.Fatalf("expected profiles projection to keep shared note/default profile, got %#v", projection)
	}
}

func TestProjectSharedSessionBrowserInspectionActionWorkbenchPrefersSyncedStatusAndCarriesSessionView(t *testing.T) {
	projection := ProjectSharedSessionBrowserInspectionAction(
		SharedSessionBrowserInspectionActionRequest{
			Action:               "workbench",
			SessionID:            "sess-inspect-workbench-projection",
			SelectedInfo:         BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
			BindingRoute:         BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node"},
			RequestedProfile:     "workbench",
			ExplicitSessionScope: true,
			IncludeStatus:        true,
			IncludeProfiles:      true,
			IncludeSessionView:   true,
		},
		SharedSessionBrowserWatchObservation{
			View: SharedSessionBrowserViewObservation{
				Observation: SharedSessionBrowserStatusAndProfilesObservation{
					StatusErr:      errors.New("status unavailable"),
					Profiles:       &BrowserProfilesResult{DefaultProfile: "workbench"},
					HasSyncedState: true,
					SyncedState: SharedSessionBrowserProfileState{
						Backend:       "proxy",
						Profile:       "workbench",
						RuntimeTarget: "node",
						BrowserApp:    "Chromium",
						Status:        "running",
						Running:       true,
						Connected:     true,
					},
				},
				Session: SharedSessionBrowserSessionViewSnapshot{
					TargetCount: 1,
					Routes: []SharedSessionBrowserRouteSnapshot{{
						Backend:       "proxy",
						Profile:       "workbench",
						RuntimeTarget: "node",
						BrowserApp:    "Chromium",
					}},
				},
			},
			Profiles: []SharedSessionBrowserProjectedProfileState{{
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
			DefaultProfile: "workbench",
			Note:           "profiles ok",
		},
		nil,
	)

	if !projection.HasProfileStatus || projection.ProfileStatus.Profile != "workbench" || projection.ProfileStatus.Status != "running" || !projection.ProfileStatus.Connected {
		t.Fatalf("expected workbench projection to prefer synced profile status, got %#v", projection)
	}
	if !projection.HasProfiles || len(projection.Profiles) != 1 {
		t.Fatalf("expected workbench projection to retain projected profiles, got %#v", projection)
	}
	if !projection.HasSessionView || projection.SessionProjection.TargetCount != 1 || len(projection.SessionProjection.Routes) != 1 {
		t.Fatalf("expected workbench projection to carry shared session projection, got %#v", projection.SessionProjection)
	}
	if projection.Note != "status unavailable" {
		t.Fatalf("expected workbench projection to prefer status error over shared note, got %#v", projection)
	}
}

func TestObserveProjectedSharedSessionBrowserInspectionActionProjectsWorkbench(t *testing.T) {
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:    "proxy",
			BrowserApp: "Chromium",
			Profile:    "workbench",
			Status:     "running",
			Running:    true,
			Connected:  true,
			PlaywrightCache: &BrowserPlaywrightCacheSummary{
				HostOS:                     "darwin",
				NodeVersion:                "24.2.0",
				PlaywrightPackageVersion:   "1.55.0",
				RuntimeSummaryGeneration:   "runtime-123",
				RuntimeBaselineReady:       true,
				SelectedLaunchSource:       "runtime_observed",
				SelectedLaunchDelivery:     "delivery-123",
				SelectedLaunchReady:        true,
				SelectedLaunchExecutableOK: true,
				LaunchReady:                true,
			},
		},
		profilesResp: BrowserProfilesResult{
			Backend:        "proxy",
			DefaultProfile: "workbench",
			Note:           "profiles ok",
			Profiles: []BrowserProfileInfo{
				{Profile: "workbench", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
			},
		},
	}
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionID := "sess-inspect-observe-projected"

	sessionRegistry.TrackCurrentTarget(sessionID, BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	observation := ObserveProjectedSharedSessionBrowserInspectionAction(
		context.Background(),
		SharedSessionBrowserObserverManagerFor(sessionRegistry, nil, stateRegistry, time.Minute).Bind(backend),
		BuildSharedSessionBrowserInspectionActionRequest(SharedSessionBrowserInspectionActionInput{
			Action:               "workbench",
			SessionID:            sessionID,
			SelectedInfo:         BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
			BindingRoute:         BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node"},
			RequestedProfile:     "workbench",
			ExplicitSessionScope: true,
			IncludeStatus:        true,
			IncludeProfiles:      true,
			IncludeSessionView:   true,
		}),
		stateRegistry,
	)

	if observation.Request.Action != "workbench" {
		t.Fatalf("expected normalized inspection request, got %#v", observation.Request)
	}
	if observation.Watch.View.Observation.Status == nil || observation.Watch.View.Observation.Profiles == nil {
		t.Fatalf("expected shared watch observation, got %#v", observation.Watch.View.Observation)
	}
	if !observation.Projection.HasProfileStatus || observation.Projection.ProfileStatus.Profile != "workbench" {
		t.Fatalf("expected projected profile status, got %#v", observation.Projection)
	}
	if observation.Projection.RuntimeStatus == nil ||
		observation.Projection.RuntimeStatus.PlaywrightCache == nil ||
		observation.Projection.RuntimeStatus.PlaywrightCache.RuntimeSummaryGeneration != "runtime-123" ||
		!observation.Projection.RuntimeStatus.PlaywrightCache.LaunchReady {
		t.Fatalf("expected projected runtime status to preserve playwright cache, got %#v", observation.Projection.RuntimeStatus)
	}
	if !observation.Projection.HasProfiles || len(observation.Projection.Profiles) != 1 {
		t.Fatalf("expected projected profiles, got %#v", observation.Projection.Profiles)
	}
	if !observation.Projection.HasSessionView || observation.Projection.SessionProjection.TargetCount != 1 {
		t.Fatalf("expected projected session view, got %#v", observation.Projection.SessionProjection)
	}
}

func TestProjectSharedSessionBrowserInspectionActionProjectsSelectionOnlyWorkbenchSessionView(t *testing.T) {
	projection := ProjectSharedSessionBrowserInspectionAction(
		BuildSharedSessionBrowserInspectionActionRequest(SharedSessionBrowserInspectionActionInput{
			Action:               "workbench",
			SessionID:            "sess-inspect-workbench-selection-only",
			SelectedInfo:         BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
			BindingRoute:         BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node"},
			RequestedProfile:     "workbench",
			ExplicitSessionScope: true,
			IncludeStatus:        true,
			IncludeProfiles:      true,
			IncludeSessionView:   true,
		}),
		SharedSessionBrowserWatchObservation{
			View: SharedSessionBrowserViewObservation{
				Session: sharedSessionBrowserSessionViewSnapshotFromBinding(
					SharedSessionBrowserBindingEvaluation{
						Snapshot: SharedSessionBrowserBindingSnapshot{
							CurrentTargetID: "tab-5",
							SelectedProfileSelection: &SharedSessionBrowserProfileSelection{
								Backend:       "proxy",
								Profile:       "workbench",
								RuntimeTarget: "node",
								BrowserApp:    "Chromium",
								Source:        "remember_profile",
							},
							SelectedTargetSelection: &BrowserSessionTargetSelection{
								ID:            "tab-5",
								Backend:       "proxy",
								Profile:       "workbench",
								RuntimeTarget: "node",
								BrowserApp:    "Chromium",
								Source:        "tracked_active_tab",
							},
						},
					},
				),
			},
		},
		nil,
	)

	if !projection.HasSessionView || projection.SessionProjection.TargetCount != 1 {
		t.Fatalf("expected workbench inspection projection to keep synthesized session view, got %#v", projection.SessionProjection)
	}
	if len(projection.SessionProjection.Routes) != 1 ||
		projection.SessionProjection.Routes[0].Backend != "proxy" ||
		projection.SessionProjection.Routes[0].Profile != "workbench" ||
		projection.SessionProjection.Routes[0].RuntimeTarget != "node" ||
		projection.SessionProjection.Routes[0].BrowserApp != "Chromium" ||
		projection.SessionProjection.Routes[0].CurrentTargetID != "tab-5" ||
		projection.SessionProjection.Routes[0].CurrentTargetSource != "tracked_active_tab" {
		t.Fatalf("expected workbench inspection projection to preserve synthesized route snapshot, got %#v", projection.SessionProjection.Routes)
	}
}

func TestProjectSharedSessionBrowserInspectionActionFallsBackToRouteScopedSessionProfiles(t *testing.T) {
	projection := ProjectSharedSessionBrowserInspectionAction(
		BuildSharedSessionBrowserInspectionActionRequest(SharedSessionBrowserInspectionActionInput{
			Action:               "workbench",
			SessionID:            "sess-inspect-workbench-route-profiles",
			SelectedInfo:         BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
			BindingRoute:         BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node"},
			RequestedProfile:     "workbench",
			ExplicitSessionScope: true,
			IncludeStatus:        true,
			IncludeProfiles:      true,
			IncludeSessionView:   true,
		}),
		SharedSessionBrowserWatchObservation{
			View: SharedSessionBrowserViewObservation{
				Binding: SharedSessionBrowserBindingEvaluation{
					Snapshot: SharedSessionBrowserBindingSnapshot{
						SelectedProfileSelection: &SharedSessionBrowserProfileSelection{
							Backend:       "proxy",
							Profile:       "workbench",
							RuntimeTarget: "node",
							Source:        "remember_profile",
						},
					},
				},
				Session: SharedSessionBrowserSessionViewSnapshot{
					Routes: []SharedSessionBrowserRouteSnapshot{{
						Backend:       "proxy",
						Profile:       "workbench",
						RuntimeTarget: "node",
						BrowserApp:    "Chromium",
						Targets: []SharedSessionBrowserRouteTarget{{
							ID:         "tab-7",
							BrowserApp: "Chromium",
						}},
					}},
					TargetCount: 1,
				},
			},
		},
		nil,
	)

	if !projection.HasSessionView || len(projection.SessionProjection.Profiles) != 1 {
		t.Fatalf("expected workbench inspection projection to synthesize route-scoped session profiles, got %#v", projection.SessionProjection)
	}
	if projection.SessionProjection.Profiles[0].State.Profile != "workbench" ||
		projection.SessionProjection.Profiles[0].State.RuntimeTarget != "node" ||
		projection.SessionProjection.Profiles[0].State.Note != "cached route-scoped session snapshot" ||
		!projection.SessionProjection.Profiles[0].Selected {
		t.Fatalf("expected workbench inspection projection to preserve fallback route-scoped profile selection, got %#v", projection.SessionProjection.Profiles)
	}
}
