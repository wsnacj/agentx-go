package browserruntime

import "testing"

func TestBuildSharedSessionBrowserCoordinationSummaryUsesHealthOverride(t *testing.T) {
	summary, ok := BuildSharedSessionBrowserCoordinationSummary(
		SharedSessionBrowserCoordinationSummaryInput{
			Action:           "refresh",
			CoordinationGoal: "prepare",
			Evaluation: &SharedSessionBrowserBindingEvaluation{
				Health: SharedSessionBrowserHealthEvaluation{
					Summary: &SharedSessionBrowserHealthSummary{
						State: "restart_failed_permanent",
					},
				},
				Coordination: SharedSessionBrowserCoordinationEvaluation{
					Plan: SharedSessionBrowserCoordinationPlan{
						State: "browser_ready",
					},
				},
			},
			ProfileStatus: &BrowserProfileStatusResult{
				Backend:   "proxy",
				Profile:   "workbench",
				Status:    "running",
				Running:   true,
				Connected: true,
			},
			RestartDecision: "session_restarted",
		},
	)

	if !ok {
		t.Fatalf("expected refresh coordination summary to be handled")
	}
	if summary.State != "browser_ready" || summary.Decision != "restart_failed_permanent" || summary.Ready {
		t.Fatalf("expected health override to preserve state and force not-ready, got %#v", summary)
	}
}

func TestBuildSharedSessionBrowserCoordinationSummaryUsesCoordinateGoal(t *testing.T) {
	summary, ok := BuildSharedSessionBrowserCoordinationSummary(
		SharedSessionBrowserCoordinationSummaryInput{
			Action:            "coordinate",
			CoordinationGoal:  "sync",
			CoordinationState: "browser_ready",
			PrepareDecision:   "started",
			SyncSessionReady:  true,
			ProfileStatus: &BrowserProfileStatusResult{
				Backend:   "proxy",
				Profile:   "workbench",
				Status:    "running",
				Running:   true,
				Connected: true,
			},
		},
	)

	if !ok {
		t.Fatalf("expected coordinate coordination summary to be handled")
	}
	if summary.State != "browser_ready" || summary.Decision != "sync_ready" || !summary.Ready {
		t.Fatalf("expected coordinate summary to use sync goal semantics, got %#v", summary)
	}
}

func TestBuildSharedSessionBrowserFinalizedActionSurfaceUsesWorkbenchProjection(t *testing.T) {
	surface := BuildSharedSessionBrowserFinalizedActionSurface(
		SharedSessionBrowserFinalizedActionSurfaceRequest{
			Action:    "workbench",
			SessionID: "sess-finalized-action-surface",
			Route: BrowserSessionRoute{
				Backend: "proxy",
				Profile: "workbench",
				Target:  "node",
			},
			BindingEvaluation: &SharedSessionBrowserBindingEvaluation{
				Routes: []SharedSessionBrowserRouteSnapshot{{
					Backend:             "proxy",
					Profile:             "workbench",
					RuntimeTarget:       "node",
					CurrentTargetID:     "tab-1",
					CurrentTargetSource: "tracked_active_tab",
				}},
				Snapshot: SharedSessionBrowserBindingSnapshot{
					SelectedProfileSelection: &SharedSessionBrowserProfileSelection{
						Backend:       "proxy",
						Profile:       "workbench",
						RuntimeTarget: "node",
						Source:        "select_profile",
					},
					SelectedTargetSelection: &BrowserSessionTargetSelection{
						ID:            "tab-1",
						Backend:       "proxy",
						Profile:       "workbench",
						RuntimeTarget: "node",
						Source:        "tracked_active_tab",
					},
					Profiles: []SharedSessionBrowserProfileState{{
						Backend:       "proxy",
						Profile:       "workbench",
						RuntimeTarget: "node",
						Status:        "running",
						Running:       true,
						Connected:     true,
					}},
					Summary: SharedSessionBrowserBindingSummary{
						CurrentTargetID:      "tab-1",
						RouteTargetCount:     1,
						BrowserProfileCount:  1,
						ActiveBrowserProfile: "workbench",
					},
				},
				Coordination: SharedSessionBrowserCoordinationEvaluation{
					Plan: SharedSessionBrowserCoordinationPlan{
						State: "browser_ready",
					},
				},
			},
		},
	)

	if surface.BindingProjection == nil || surface.BindingProjection.ProfileSelection == nil || surface.BindingProjection.ProfileSelection.Profile != "workbench" {
		t.Fatalf("expected finalized action surface to keep binding projection, got %#v", surface.BindingProjection)
	}
	if surface.BindingProjection.TargetSelection == nil || surface.BindingProjection.TargetSelection.ID != "tab-1" {
		t.Fatalf("expected finalized action surface to keep target selection, got %#v", surface.BindingProjection)
	}
	if surface.SessionProjection == nil ||
		surface.SessionProjection.Projection == nil ||
		len(surface.SessionProjection.Projection.Routes) != 1 ||
		surface.SessionProjection.Projection.Routes[0].Profile != "workbench" ||
		surface.SessionProjection.Projection.TargetCount != 1 {
		t.Fatalf("expected finalized action surface to keep session projection for workbench, got %#v", surface.SessionProjection)
	}
	if surface.WorkbenchSurface == nil || surface.WorkbenchSurface.SessionProjection == nil || surface.WorkbenchSurface.SessionProjection.Projection == nil {
		t.Fatalf("expected finalized action surface to carry shared workbench surface, got %#v", surface.WorkbenchSurface)
	}
	if !surface.UseWorkbenchSurface || !surface.SyncCoordinationSurface {
		t.Fatalf("expected finalized action surface to mark workbench sync behavior, got %#v", surface)
	}
	if surface.CoordinationSummary != nil {
		t.Fatalf("expected workbench finalized action surface not to synthesize coordination summary, got %#v", surface.CoordinationSummary)
	}
}

func TestBuildSharedSessionBrowserFinalizedActionSurfaceAppliesHealthSummaryOverride(t *testing.T) {
	surface := BuildSharedSessionBrowserFinalizedActionSurface(
		SharedSessionBrowserFinalizedActionSurfaceRequest{
			Action:    "refresh",
			SessionID: "sess-finalized-action-surface-health-override",
			Route: BrowserSessionRoute{
				Backend: "proxy",
				Profile: "workbench",
				Target:  "node",
			},
			BindingEvaluation: &SharedSessionBrowserBindingEvaluation{
				Snapshot: SharedSessionBrowserBindingSnapshot{
					SelectedProfileSelection: &SharedSessionBrowserProfileSelection{
						Backend:       "proxy",
						Profile:       "workbench",
						RuntimeTarget: "node",
						Source:        "select_profile",
					},
					Profiles: []SharedSessionBrowserProfileState{{
						Backend:       "proxy",
						Profile:       "workbench",
						RuntimeTarget: "node",
						Status:        "running",
						Running:       true,
						Connected:     true,
					}},
				},
				Coordination: SharedSessionBrowserCoordinationEvaluation{
					Plan: SharedSessionBrowserCoordinationPlan{
						State: "browser_ready",
					},
				},
			},
			HealthSummary: &SharedSessionBrowserHealthSummary{
				State:          "restart_failed_permanent",
				Reason:         "browser relaunch failed twice",
				RecoveryAction: "browser action=start",
			},
			CoordinationSummary: SharedSessionBrowserCoordinationSummaryInput{
				ProfileStatus: &BrowserProfileStatusResult{
					Backend:   "proxy",
					Profile:   "workbench",
					Status:    "running",
					Running:   true,
					Connected: true,
				},
				RestartDecision: "session_restarted",
			},
		},
	)

	if surface.BindingProjection == nil || surface.BindingProjection.Evaluation.Health.Summary == nil {
		t.Fatalf("expected finalized action surface to preserve health summary override, got %#v", surface.BindingProjection)
	}
	if surface.BindingProjection.Evaluation.Health.Summary.State != "restart_failed_permanent" {
		t.Fatalf("expected finalized action surface to merge health summary override, got %#v", surface.BindingProjection.Evaluation.Health.Summary)
	}
	if surface.CoordinationSummary == nil {
		t.Fatalf("expected finalized action surface to rebuild coordination summary after health override")
	}
	if surface.CoordinationSummary.State != "browser_ready" || surface.CoordinationSummary.Decision != "restart_failed_permanent" || surface.CoordinationSummary.Ready {
		t.Fatalf("expected finalized action surface to apply health override to coordination summary, got %#v", surface.CoordinationSummary)
	}
}

func TestBuildSharedSessionBrowserFinalizedActionSurfaceUsesProvidedEvaluationWithoutRoute(t *testing.T) {
	surface := BuildSharedSessionBrowserFinalizedActionSurface(
		SharedSessionBrowserFinalizedActionSurfaceRequest{
			Action:    "status",
			SessionID: "sess-finalized-action-surface-route-less",
			BindingEvaluation: &SharedSessionBrowserBindingEvaluation{
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
				},
			},
			CurrentProfile: &SharedSessionBrowserProfileState{
				Backend:       "proxy",
				Profile:       "workbench",
				RuntimeTarget: "node",
				BrowserApp:    "Chromium",
				Status:        "running",
				Running:       true,
				Connected:     true,
				Note:          "fresh current status",
			},
		},
	)

	if surface.BindingProjection == nil || surface.BindingProjection.ProfileSelection == nil || surface.BindingProjection.ProfileSelection.Profile != "workbench" || surface.BindingProjection.ProfileSelection.Source != "remember_profile" {
		t.Fatalf("expected route-less finalized action surface to keep provided profile selection, got %#v", surface.BindingProjection)
	}
	if surface.BindingProjection.TargetSelection == nil || surface.BindingProjection.TargetSelection.ID != "tab-2" || surface.BindingProjection.TargetSelection.Source != "tracked_active_tab" {
		t.Fatalf("expected route-less finalized action surface to keep provided target selection, got %#v", surface.BindingProjection)
	}
	if len(surface.BindingProjection.Evaluation.Snapshot.Profiles) != 1 {
		t.Fatalf("expected route-less finalized action surface to merge current profile, got %#v", surface.BindingProjection.Evaluation.Snapshot.Profiles)
	}
	got := surface.BindingProjection.Evaluation.Snapshot.Profiles[0]
	if got.Profile != "workbench" || got.Status != "running" || got.Note != "fresh current status" {
		t.Fatalf("expected route-less finalized action surface to retain merged current profile state, got %#v", got)
	}
}

func TestBuildSharedSessionBrowserFinalizedActionSurfaceUsesProfileStatusFallbackAsCurrentProfile(t *testing.T) {
	surface := BuildSharedSessionBrowserFinalizedActionSurface(
		SharedSessionBrowserFinalizedActionSurfaceRequest{
			Action:    "status",
			SessionID: "sess-finalized-action-surface-profile-status-fallback",
			Route: BrowserSessionRoute{
				Backend: "proxy",
				Profile: "workbench",
				Target:  "node",
			},
			ProfileStatus: &BrowserProfileStatusResult{
				Backend:   "proxy",
				Profile:   "workbench",
				Status:    "running",
				Running:   true,
				Connected: true,
				Note:      "fresh current status",
			},
		},
	)

	if surface.BindingProjection == nil {
		t.Fatalf("expected finalized action surface to build binding projection")
	}
	if len(surface.BindingProjection.Evaluation.Snapshot.Profiles) != 1 {
		t.Fatalf("expected finalized action surface to merge profile status fallback into projection, got %#v", surface.BindingProjection.Evaluation.Snapshot.Profiles)
	}
	got := surface.BindingProjection.Evaluation.Snapshot.Profiles[0]
	if got.Profile != "workbench" || got.RuntimeTarget != "node" || got.Status != "running" || got.Note != "fresh current status" {
		t.Fatalf("expected finalized action surface to derive current profile from profile status fallback, got %#v", got)
	}
}

func TestBuildSharedSessionBrowserFinalizedActionSurfaceDefaultsCoordinationProfileStatus(t *testing.T) {
	surface := BuildSharedSessionBrowserFinalizedActionSurface(
		SharedSessionBrowserFinalizedActionSurfaceRequest{
			Action:    "refresh",
			SessionID: "sess-finalized-action-surface-coordination-profile-status",
			Route: BrowserSessionRoute{
				Backend: "proxy",
				Profile: "workbench",
				Target:  "node",
			},
			BindingEvaluation: &SharedSessionBrowserBindingEvaluation{
				Coordination: SharedSessionBrowserCoordinationEvaluation{
					Plan: SharedSessionBrowserCoordinationPlan{
						State: "browser_ready",
					},
				},
			},
			ProfileStatus: &BrowserProfileStatusResult{
				Backend:   "proxy",
				Profile:   "workbench",
				Status:    "running",
				Running:   true,
				Connected: true,
			},
			CoordinationSummary: SharedSessionBrowserCoordinationSummaryInput{
				RestartDecision: "session_restarted",
			},
		},
	)

	if surface.CoordinationSummary == nil {
		t.Fatalf("expected finalized action surface to synthesize coordination summary from profile-status fallback")
	}
	if surface.CoordinationSummary.State != "browser_ready" || surface.CoordinationSummary.Decision != "restart_ready" || !surface.CoordinationSummary.Ready {
		t.Fatalf("expected finalized action surface to default coordination profile status from request, got %#v", surface.CoordinationSummary)
	}
}

func TestBuildSharedSessionBrowserFinalizedActionSurfaceProjectsWorkbenchSessionViewFromEvaluation(t *testing.T) {
	surface := BuildSharedSessionBrowserFinalizedActionSurface(
		SharedSessionBrowserFinalizedActionSurfaceRequest{
			Action:    "workbench",
			SessionID: "sess-finalized-action-surface-workbench-route-less",
			BindingEvaluation: &SharedSessionBrowserBindingEvaluation{
				Routes: []SharedSessionBrowserRouteSnapshot{{
					Backend:             "proxy",
					Profile:             "workbench",
					RuntimeTarget:       "node",
					CurrentTargetID:     "tab-2",
					CurrentTargetSource: "tracked_active_tab",
				}},
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
					Runs: []SharedSessionRunInfo{{
						RunID:    "run-1",
						NodeID:   "node-1",
						Status:   "running",
						Action:   "nodes action=run",
						Provider: "browser",
					}},
					Profiles: []SharedSessionBrowserProfileState{{
						Backend:       "proxy",
						Profile:       "workbench",
						RuntimeTarget: "node",
						BrowserApp:    "Chromium",
						Status:        "running",
						Running:       true,
						Connected:     true,
					}},
					Summary: SharedSessionBrowserBindingSummary{
						CurrentTargetID:     "tab-2",
						RouteTargetCount:    1,
						NodeRunCount:        1,
						BrowserProfileCount: 1,
					},
				},
			},
		},
	)

	if surface.SessionProjection == nil {
		t.Fatalf("expected workbench finalized action surface to project session view from binding evaluation")
	}
	if surface.WorkbenchSurface == nil || surface.WorkbenchSurface.SessionProjection == nil {
		t.Fatalf("expected workbench finalized action surface to project shared workbench session surface, got %#v", surface.WorkbenchSurface)
	}
	if surface.SessionProjection.Projection == nil ||
		len(surface.SessionProjection.Projection.Routes) != 1 ||
		surface.SessionProjection.Projection.Routes[0].CurrentTargetID != "tab-2" ||
		surface.SessionProjection.Projection.TargetCount != 1 {
		t.Fatalf("expected workbench finalized action surface to keep route-scoped session projection, got %#v", surface.SessionProjection)
	}
	if len(surface.SessionProjection.Projection.Runs) != 1 || surface.SessionProjection.Projection.Runs[0].RunID != "run-1" {
		t.Fatalf("expected workbench finalized action surface to keep shared run projection, got %#v", surface.SessionProjection.Projection.Runs)
	}
	if len(surface.SessionProjection.Projection.Profiles) != 1 || surface.SessionProjection.Projection.Profiles[0].State.Profile != "workbench" || !surface.SessionProjection.Projection.Profiles[0].Selected {
		t.Fatalf("expected workbench finalized action surface to keep projected session profiles, got %#v", surface.SessionProjection.Projection.Profiles)
	}
}

func TestBuildSharedSessionBrowserFinalizedActionSurfaceSynthesizesWorkbenchSelectionRouteWithoutRouteSnapshot(t *testing.T) {
	surface := BuildSharedSessionBrowserFinalizedActionSurface(
		SharedSessionBrowserFinalizedActionSurfaceRequest{
			Action:    "workbench",
			SessionID: "sess-finalized-action-surface-workbench-selection-only",
			BindingEvaluation: &SharedSessionBrowserBindingEvaluation{
				Snapshot: SharedSessionBrowserBindingSnapshot{
					CurrentTargetID: "tab-3",
					SelectedProfileSelection: &SharedSessionBrowserProfileSelection{
						Backend:       "proxy",
						Profile:       "workbench",
						RuntimeTarget: "node",
						BrowserApp:    "Chromium",
						Source:        "remember_profile",
					},
					SelectedTargetSelection: &BrowserSessionTargetSelection{
						ID:            "tab-3",
						Backend:       "proxy",
						Profile:       "workbench",
						RuntimeTarget: "node",
						BrowserApp:    "Chromium",
						Source:        "tracked_active_tab",
					},
					Summary: SharedSessionBrowserBindingSummary{
						CurrentTargetID: "tab-3",
					},
				},
			},
		},
	)

	if surface.SessionProjection == nil {
		t.Fatalf("expected workbench finalized action surface to synthesize session view from binding selections")
	}
	if surface.WorkbenchSurface == nil || surface.WorkbenchSurface.SessionProjection == nil {
		t.Fatalf("expected workbench finalized action surface to synthesize shared workbench surface, got %#v", surface.WorkbenchSurface)
	}
	if surface.SessionProjection.Projection == nil ||
		len(surface.SessionProjection.Projection.Routes) != 1 ||
		surface.SessionProjection.Projection.Routes[0].Backend != "proxy" ||
		surface.SessionProjection.Projection.Routes[0].Profile != "workbench" ||
		surface.SessionProjection.Projection.Routes[0].RuntimeTarget != "node" ||
		surface.SessionProjection.Projection.Routes[0].BrowserApp != "Chromium" ||
		surface.SessionProjection.Projection.Routes[0].CurrentTargetID != "tab-3" ||
		surface.SessionProjection.Projection.Routes[0].CurrentTargetSource != "tracked_active_tab" ||
		surface.SessionProjection.Projection.TargetCount != 1 {
		t.Fatalf("expected workbench finalized action surface to synthesize minimal session route, got %#v", surface.SessionProjection)
	}
}

func TestBuildSharedSessionBrowserFinalizedActionSurfaceKeepsWorkbenchSurfaceWithoutSessionView(t *testing.T) {
	surface := BuildSharedSessionBrowserFinalizedActionSurface(
		SharedSessionBrowserFinalizedActionSurfaceRequest{
			Action:    "workbench",
			SessionID: "sess-finalized-action-surface-binding-only-workbench",
			BindingEvaluation: &SharedSessionBrowserBindingEvaluation{
				Snapshot: SharedSessionBrowserBindingSnapshot{
					SelectedProfileSelection: &SharedSessionBrowserProfileSelection{
						Backend:       "proxy",
						Profile:       "workbench",
						RuntimeTarget: "node",
						BrowserApp:    "Chromium",
						Source:        "select_profile",
					},
				},
			},
			ProfileStatus: &BrowserProfileStatusResult{
				Backend:   "proxy",
				Profile:   "workbench",
				Status:    "running",
				Running:   true,
				Connected: true,
			},
		},
	)

	if !surface.UseWorkbenchSurface || !surface.SyncCoordinationSurface {
		t.Fatalf("expected binding-only workbench finalized action to preserve workbench sync semantics, got %#v", surface)
	}
	if surface.SessionProjection == nil || surface.SessionProjection.Projection == nil {
		t.Fatalf("expected binding-only workbench finalized action to synthesize minimal shared session projection, got %#v", surface.SessionProjection)
	}
	if len(surface.SessionProjection.Projection.Routes) != 1 ||
		surface.SessionProjection.Projection.Routes[0].Backend != "proxy" ||
		surface.SessionProjection.Projection.Routes[0].Profile != "workbench" ||
		surface.SessionProjection.Projection.Routes[0].RuntimeTarget != "node" ||
		surface.SessionProjection.Projection.Routes[0].BrowserApp != "Chromium" ||
		surface.SessionProjection.Projection.TargetCount != 1 {
		t.Fatalf("expected binding-only workbench finalized action to keep minimal session route, got %#v", surface.SessionProjection)
	}
	if surface.WorkbenchSurface == nil || surface.WorkbenchSurface.BindingProjection == nil {
		t.Fatalf("expected binding-only workbench finalized action to keep shared workbench surface, got %#v", surface.WorkbenchSurface)
	}
	if surface.WorkbenchSurface.SessionProjection == nil ||
		surface.WorkbenchSurface.SessionProjection.Projection == nil ||
		len(surface.WorkbenchSurface.SessionProjection.Projection.Routes) != 1 {
		t.Fatalf("expected binding-only workbench finalized action to keep synthesized session projection inside shared workbench surface, got %#v", surface.WorkbenchSurface)
	}
	if surface.WorkbenchSurface.ProfileInventory == nil ||
		!surface.WorkbenchSurface.ProfileInventory.ApplyProfileInventory ||
		surface.WorkbenchSurface.ProfileInventory.DefaultProfile != "workbench" ||
		len(surface.WorkbenchSurface.ProfileInventory.Profiles) != 1 ||
		surface.WorkbenchSurface.ProfileInventory.Profiles[0].State.Profile != "workbench" ||
		!surface.WorkbenchSurface.ProfileInventory.Profiles[0].Selected {
		t.Fatalf("expected binding-only workbench finalized action to keep profile inventory in shared workbench surface, got %#v", surface.WorkbenchSurface)
	}
}
