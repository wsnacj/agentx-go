package browserruntime

import (
	"reflect"
	"testing"
)

func TestBuildSharedSessionBrowserWorkbenchProjectionUsesEvaluationCoordinationPlan(t *testing.T) {
	projection := BuildSharedSessionBrowserWorkbenchProjection(
		SharedSessionBrowserWorkbenchProjectionRequest{
			HasProfileStatus:        true,
			HasProfiles:             true,
			HasSessionProjection:    true,
			ExtraSections:           []string{"coordination"},
			SyncCoordinationSurface: true,
			Evaluation: &SharedSessionBrowserBindingEvaluation{
				Coordination: SharedSessionBrowserCoordinationEvaluation{
					Plan: SharedSessionBrowserCoordinationPlan{
						State:                     "browser_ready",
						PrimaryBrowserAction:      "browser action=refresh",
						PrimaryNodeAction:         "nodes action=run",
						NextStep:                  "browser action=refresh",
						RecommendedBrowserActions: []string{"browser action=refresh"},
						RecommendedNodeActions:    []string{"nodes action=run"},
					},
				},
			},
		},
	)

	if !projection.Ready || !reflect.DeepEqual(projection.Sections, []string{"route", "status", "profiles", "sessions", "coordination"}) {
		t.Fatalf("expected workbench projection to compose shared sections, got %#v", projection)
	}
	if projection.CoordinationSummary == nil || projection.CoordinationSummary.State != "browser_ready" || projection.CoordinationSummary.Decision != "" || projection.CoordinationSummary.Ready {
		t.Fatalf("expected workbench projection to surface coordination state without lifecycle decision, got %#v", projection.CoordinationSummary)
	}
	if projection.ActionPlan == nil ||
		projection.ActionPlan.PrimaryBrowserAction != "browser action=refresh" ||
		projection.ActionPlan.PrimaryNodeAction != "nodes action=run" ||
		projection.ActionPlan.NextStep != "browser action=refresh" ||
		!reflect.DeepEqual(projection.ActionPlan.RecommendedBrowserActions, []string{"browser action=refresh"}) ||
		!reflect.DeepEqual(projection.ActionPlan.RecommendedNodeActions, []string{"nodes action=run"}) {
		t.Fatalf("expected workbench projection to surface shared action plan, got %#v", projection.ActionPlan)
	}
}

func TestBuildSharedSessionBrowserWorkbenchProjectionUsesFallbackCoordinationPlan(t *testing.T) {
	projection := BuildSharedSessionBrowserWorkbenchProjection(
		SharedSessionBrowserWorkbenchProjectionRequest{
			HasSessionProjection:    true,
			SyncCoordinationSurface: true,
			CoordinationPlan: &SharedSessionBrowserCoordinationPlan{
				State:                  "idle",
				PrimaryNodeAction:      "nodes action=run_status",
				NextStep:               "nodes action=run_status",
				RecommendedNodeActions: []string{"nodes action=run_status"},
			},
		},
	)

	if !projection.Ready || !reflect.DeepEqual(projection.Sections, []string{"route", "sessions", "coordination"}) {
		t.Fatalf("expected fallback coordination plan to drive workbench sections, got %#v", projection)
	}
	if projection.CoordinationSummary == nil || projection.CoordinationSummary.State != "idle" {
		t.Fatalf("expected fallback coordination plan to surface state, got %#v", projection.CoordinationSummary)
	}
	if projection.ActionPlan == nil || projection.ActionPlan.PrimaryNodeAction != "nodes action=run_status" || projection.ActionPlan.NextStep != "nodes action=run_status" {
		t.Fatalf("expected fallback coordination plan to surface action plan, got %#v", projection.ActionPlan)
	}
}

func TestBuildSharedSessionBrowserWorkbenchProjectionClearsCoordinationSurfaceWithoutSync(t *testing.T) {
	projection := BuildSharedSessionBrowserWorkbenchProjection(
		SharedSessionBrowserWorkbenchProjectionRequest{
			HasSessionProjection: true,
			ExtraSections:        []string{"coordination"},
			ClearActionPlan:      true,
		},
	)

	if !projection.Ready || !reflect.DeepEqual(projection.Sections, []string{"route", "sessions", "coordination"}) {
		t.Fatalf("expected workbench projection to preserve explicit sections without sync, got %#v", projection)
	}
	if !projection.ClearActionPlan {
		t.Fatalf("expected workbench projection to preserve clear-action-plan request, got %#v", projection)
	}
	if !projection.ClearCoordinationSummary || projection.CoordinationSummary != nil || projection.ActionPlan != nil {
		t.Fatalf("expected workbench projection without sync to clear stale coordination surface, got %#v", projection)
	}
}
