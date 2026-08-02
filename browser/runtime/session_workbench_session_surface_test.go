package browserruntime

import "testing"

func TestBuildSharedSessionBrowserWorkbenchSessionSurfaceProjectionBuildsFromBindingEvaluation(t *testing.T) {
	projection := BuildSharedSessionBrowserWorkbenchSessionSurfaceProjection(
		SharedSessionBrowserWorkbenchSessionSurfaceRequest{
			SelectedInfo:              BrowserRuntimeInfo{Backend: "proxy", Profile: "alpha", Target: "node"},
			RequestedDefaultProfile:   "alpha",
			NeedProfileStatus:         true,
			NeedProfileInventory:      true,
			ApplyConfiguredProfiles:   true,
			ApplyMissingSessionIDNote: true,
			ApplyBindingEvaluation:    true,
			BindingEvaluation: &SharedSessionBrowserBindingEvaluation{
				Routes: []SharedSessionBrowserRouteSnapshot{{
					Backend:       "proxy",
					Profile:       "alpha",
					RuntimeTarget: "node",
				}},
				Snapshot: SharedSessionBrowserBindingSnapshot{
					SelectedProfileSelection: &SharedSessionBrowserProfileSelection{
						Backend:       "proxy",
						Profile:       "alpha",
						RuntimeTarget: "node",
					},
					SelectedTargetSelection: &BrowserSessionTargetSelection{
						Backend:       "proxy",
						Profile:       "alpha",
						RuntimeTarget: "node",
						BrowserApp:    "chrome",
					},
					Profiles: []SharedSessionBrowserProfileState{{
						Backend:       "proxy",
						Profile:       "alpha",
						RuntimeTarget: "node",
						Status:        "running",
					}},
				},
			},
		},
	)

	if projection.BindingProjection == nil || projection.BindingProjection.TargetSelection == nil || projection.BindingProjection.TargetSelection.BrowserApp != "chrome" {
		t.Fatalf("expected workbench session surface to keep shared binding projection, got %#v", projection.BindingProjection)
	}
	if projection.SessionProjection == nil || projection.SessionProjection.Projection == nil || len(projection.SessionProjection.Projection.Routes) != 1 {
		t.Fatalf("expected workbench session surface to derive shared session projection, got %#v", projection.SessionProjection)
	}
	if projection.SessionProjection.MissingSessionIDNote != "browser_runtime: no tool session context is available" {
		t.Fatalf("expected shared workbench session projection to carry canonical missing-session note, got %#v", projection.SessionProjection)
	}
	if projection.ProfileInventory == nil || !projection.ProfileInventory.HasProfileStatus || projection.ProfileInventory.ProfileStatus.Profile != "alpha" {
		t.Fatalf("expected workbench session surface to derive shared profile inventory, got %#v", projection.ProfileInventory)
	}
}

func TestBuildSharedSessionBrowserWorkbenchSessionSurfaceProjectionPrefersExplicitSessionProjection(t *testing.T) {
	explicitSession := &SharedSessionBrowserTopLevelSessionProjection{
		Routes: []SharedSessionBrowserRouteSnapshot{{
			Backend:       "proxy",
			Profile:       "explicit",
			RuntimeTarget: "sandbox",
		}},
		TargetCount: 1,
	}
	explicitBinding := &SharedSessionBrowserTopLevelBindingProjection{
		ProfileSelection: &SharedSessionBrowserProfileSelection{
			Backend:       "proxy",
			Profile:       "explicit",
			RuntimeTarget: "sandbox",
		},
	}

	projection := BuildSharedSessionBrowserWorkbenchSessionSurfaceProjection(
		SharedSessionBrowserWorkbenchSessionSurfaceRequest{
			SelectedInfo:              BrowserRuntimeInfo{Backend: "proxy", Profile: "explicit", Target: "sandbox"},
			BindingProjection:         explicitBinding,
			SessionProjection:         explicitSession,
			ApplyConfiguredProfiles:   false,
			ApplyMissingSessionIDNote: false,
			NeedProfileStatus:         false,
			NeedProfileInventory:      false,
		},
	)

	if projection.BindingProjection == nil || projection.BindingProjection.ProfileSelection == nil || projection.BindingProjection.ProfileSelection.Profile != "explicit" {
		t.Fatalf("expected explicit binding projection to be preserved, got %#v", projection.BindingProjection)
	}
	if projection.SessionProjection == nil || projection.SessionProjection.Projection == nil || len(projection.SessionProjection.Projection.Routes) != 1 || projection.SessionProjection.Projection.Routes[0].Profile != "explicit" {
		t.Fatalf("expected explicit session projection to be preserved, got %#v", projection.SessionProjection)
	}
	if projection.SessionProjection.MissingSessionIDNote != "" {
		t.Fatalf("expected explicit session projection without missing-session note, got %#v", projection.SessionProjection)
	}
	if projection.ProfileInventory != nil {
		t.Fatalf("expected no fallback inventory when not requested, got %#v", projection.ProfileInventory)
	}
}

func TestBuildSharedSessionBrowserWorkbenchSessionSurfaceProjectionAddsFallbackBindingSessionProjection(t *testing.T) {
	projection := BuildSharedSessionBrowserWorkbenchSessionSurfaceProjection(
		SharedSessionBrowserWorkbenchSessionSurfaceRequest{
			SelectedInfo:              BrowserRuntimeInfo{Backend: "proxy", Profile: "alpha", Target: "node"},
			ConfiguredInfo:            BrowserRuntimeInfo{Backend: "proxy", Profile: "alpha", Target: "node"},
			RequestedDefaultProfile:   "alpha",
			NeedProfileStatus:         true,
			NeedProfileInventory:      true,
			ApplyConfiguredProfiles:   true,
			ApplyMissingSessionIDNote: true,
			ApplyBindingEvaluation:    true,
			SessionProjection: &SharedSessionBrowserTopLevelSessionProjection{
				Routes: []SharedSessionBrowserRouteSnapshot{{
					Backend:       "proxy",
					Profile:       "alpha",
					RuntimeTarget: "node",
				}},
			},
			BindingEvaluation: &SharedSessionBrowserBindingEvaluation{
				Routes: []SharedSessionBrowserRouteSnapshot{{
					Backend:       "proxy",
					Profile:       "alpha",
					RuntimeTarget: "node",
				}},
				Snapshot: SharedSessionBrowserBindingSnapshot{
					Runs: []SharedSessionRunInfo{{
						RunID: "run-1",
					}},
					Profiles: []SharedSessionBrowserProfileState{{
						Backend:       "proxy",
						Profile:       "alpha",
						RuntimeTarget: "node",
						Status:        "running",
					}},
					Summary: SharedSessionBrowserBindingSummary{
						RouteTargetCount: 1,
						ActiveNodeRunID:  "run-1",
					},
				},
			},
		},
	)

	if projection.SessionProjection == nil || projection.SessionProjection.Projection == nil || len(projection.SessionProjection.Projection.Routes) != 1 {
		t.Fatalf("expected explicit session projection to remain primary, got %#v", projection.SessionProjection)
	}
	if projection.FallbackSessionProjection == nil || projection.FallbackSessionProjection.Projection == nil {
		t.Fatalf("expected binding-backed fallback session projection, got %#v", projection.FallbackSessionProjection)
	}
	if projection.FallbackSessionProjection.ConfiguredInfo != (BrowserRuntimeInfo{Backend: "proxy", Profile: "alpha", Target: "node"}) {
		t.Fatalf("expected fallback session projection to carry configured info, got %#v", projection.FallbackSessionProjection.ConfiguredInfo)
	}
	if projection.FallbackSessionProjection.MissingSessionIDNote != "browser_runtime: no tool session context is available" {
		t.Fatalf("expected fallback session projection to carry canonical missing-session note, got %#v", projection.FallbackSessionProjection)
	}
	if len(projection.FallbackSessionProjection.Projection.Runs) != 1 || len(projection.FallbackSessionProjection.Projection.Profiles) != 1 {
		t.Fatalf("expected fallback session projection to restore runs/profiles, got %#v", projection.FallbackSessionProjection.Projection)
	}
	if projection.ProfileInventory == nil || !projection.ProfileInventory.HasProfileStatus || len(projection.ProfileInventory.Profiles) != 1 {
		t.Fatalf("expected profile inventory to fall back to binding-backed session projection, got %#v", projection.ProfileInventory)
	}
}
