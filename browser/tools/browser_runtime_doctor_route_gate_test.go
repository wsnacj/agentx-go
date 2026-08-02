package tools

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
	types "github.com/wsnacj/agentx-go/components/llm"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

const (
	runtimeOnlyDoctorRouteNextStepAlias = "prepare"
	runtimeOnlyDoctorRouteNextStep      = "browser_runtime action=prepare"
)

func TestRegisterBrowserTools_RuntimeStatusUsesDoctorRouteSummaryWhenManagedRouteStaysHidden(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-status-doctor-route-gate")
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-status-doctor-route-gate", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Status:        "running",
		Running:       true,
		Connected:     true,
		Note:          "cached implicit host default profile",
	})
	sessionStateRegistry.SelectBrowserProfile("browser-runtime-status-doctor-route-gate", agentxbrowserruntime.SharedSessionBrowserProfileSelection{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Source:        "remember_profile",
	})
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root: t.TempDir(),
		NodeBackend: browserHiddenManagedRouteDoctorGateRuntimeControlNodeBackend(BrowserCapabilities{
			Open:           true,
			RuntimeStatus:  true,
			RuntimePrepare: true,
		}),
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime", "browser_open"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"status"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime status doctor route gate: %v", err)
	}
	var payload struct {
		Action                     string                          `json:"action"`
		Status                     string                          `json:"status"`
		Note                       string                          `json:"note"`
		SubstratePosture           string                          `json:"substrate_posture"`
		SubstrateStatus            string                          `json:"substrate_status"`
		SubstrateReason            string                          `json:"substrate_reason"`
		SubstrateSelectionStrategy string                          `json:"substrate_selection_strategy"`
		SubstrateSelectionReason   string                          `json:"substrate_selection_reason"`
		DefaultRoute               browserRuntimeRouteDescriptor   `json:"default_route"`
		DefaultCandidateRoute      browserRuntimeRouteDescriptor   `json:"default_candidate_route"`
		SelectedRoute              *browserRuntimeRouteDescriptor  `json:"selected_route"`
		RouteResolution            any                             `json:"route_resolution"`
		Doctor                     *BrowserDoctorSummary           `json:"doctor"`
		Explanation                *browserTopLevelSummary         `json:"explanation"`
		Diagnostics                *browserTopLevelSummary         `json:"diagnostics"`
		Summary                    *browserTopLevelSummary         `json:"summary"`
		Display                    *browserTopLevelDisplaySummary  `json:"display"`
		Surface                    *browserTopLevelSurfaceSummary  `json:"surface"`
		View                       *browserTopLevelViewSummary     `json:"view"`
		BrowserSurface             string                          `json:"browser_surface"`
		BrowserOptInTargets        []string                        `json:"browser_opt_in_targets"`
		SubstrateMatrix            []browserRuntimeSubstrateStatus `json:"substrate_matrix"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_runtime status doctor route gate output: %v", err)
	}

	if payload.Action != "status" || payload.Status != "ok" {
		t.Fatalf("unexpected browser_runtime status payload: %#v", payload)
	}
	if payload.DefaultRoute.Backend != "proxy" || payload.DefaultRoute.Profile != "isolated" || payload.DefaultRoute.RuntimeTarget != "node" {
		t.Fatalf("expected doctor-route gate to promote managed candidate default_route, got %#v", payload.DefaultRoute)
	}
	if payload.DefaultCandidateRoute.Backend != "proxy" || payload.DefaultCandidateRoute.Profile != "isolated" || payload.DefaultCandidateRoute.RuntimeTarget != "node" {
		t.Fatalf("expected doctor-route gate to expose hidden managed default_candidate_route, got %#v", payload.DefaultCandidateRoute)
	}
	if payload.SelectedRoute != nil || payload.RouteResolution != nil {
		t.Fatalf("expected hidden legacy-host default route to stay out of top-level route selection, got selected=%#v resolution=%#v", payload.SelectedRoute, payload.RouteResolution)
	}
	if payload.Doctor == nil || payload.Doctor.Route == nil {
		t.Fatalf("expected doctor route summary, got %#v", payload.Doctor)
	}
	if payload.Doctor.Route.Code != "managed_route_hidden_by_legacy_host_default" || payload.Doctor.Route.Status != "warn" {
		t.Fatalf("unexpected doctor route summary: %#v", payload.Doctor.Route)
	}
	if payload.Doctor.Route.SelectionStrategy != BrowserSubstrateSelectionPreferNodeOverLegacy {
		t.Fatalf("expected doctor route selection_strategy to align with managed candidate default, got route=%#v", payload.Doctor.Route)
	}
	if payload.SubstrateSelectionStrategy != BrowserSubstrateSelectionPreferNodeOverLegacy {
		t.Fatalf("expected status substrate_selection_strategy to align with managed candidate default, got strategy=%q payload=%#v", payload.SubstrateSelectionStrategy, payload)
	}
	if payload.SubstrateSelectionReason != strings.TrimSpace(payload.Doctor.Route.Summary) {
		t.Fatalf("expected status substrate_selection_reason to align with doctor route summary, got reason=%q route=%#v", payload.SubstrateSelectionReason, payload.Doctor.Route)
	}
	if payload.SubstratePosture != BrowserSubstrateNodeRuntime ||
		payload.SubstrateStatus != "ok" ||
		payload.SubstrateReason != strings.TrimSpace(payload.Doctor.Route.Summary) {
		t.Fatalf("expected status top-level substrate summary to align with hidden managed candidate, got posture=%q status=%q reason=%q payload=%#v", payload.SubstratePosture, payload.SubstrateStatus, payload.SubstrateReason, payload)
	}
	if !strings.Contains(payload.Note, strings.TrimSpace(payload.Doctor.Route.Summary)) {
		t.Fatalf("expected note to surface hidden managed route guidance, got %#v", payload.Note)
	}
	assertManagedRouteDoctorGateExplanationSummary(t, payload.Explanation)
	assertManagedRouteDoctorGateTopLevelSummary(t, payload.Diagnostics)
	assertManagedRouteDoctorGateTopLevelSummary(t, payload.Summary)
	if payload.Summary == nil || payload.Summary.DefaultCandidateRoute != payload.DefaultCandidateRoute {
		t.Fatalf("expected status summary to expose hidden managed default_candidate_route, got %#v", payload.Summary)
	}
	assertManagedRouteDoctorGateDisplaySummary(t, payload.Display, false)
	assertManagedRouteDoctorGateSurfaceSummary(t, payload.Surface)
	assertManagedRouteDoctorGateViewSummary(t, payload.View)
	assertManagedRouteDoctorGateRootSurface(t, payload.BrowserSurface, payload.BrowserOptInTargets)
	assertManagedRouteDoctorGateSubstrateMatrix(t, payload.SubstrateMatrix, strings.TrimSpace(payload.Doctor.Route.Summary))
}

func TestRegisterBrowserTools_RuntimeDoctorUsesDoctorRouteSummaryWhenManagedRouteStaysHidden(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-doctor-route-gate")
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-doctor-route-gate", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Status:        "running",
		Running:       true,
		Connected:     true,
		Note:          "cached implicit host default profile",
	})
	sessionStateRegistry.SelectBrowserProfile("browser-runtime-doctor-route-gate", agentxbrowserruntime.SharedSessionBrowserProfileSelection{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Source:        "remember_profile",
	})
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root: t.TempDir(),
		NodeBackend: browserHiddenManagedRouteDoctorGateRuntimeControlNodeBackend(BrowserCapabilities{
			RuntimeStatus:  true,
			RuntimePrepare: true,
			Open:           true,
		}),
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime", "browser_open"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"doctor"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime doctor doctor route gate: %v", err)
	}
	var payload struct {
		Action                     string                          `json:"action"`
		Status                     string                          `json:"status"`
		Note                       string                          `json:"note"`
		DefaultProfile             string                          `json:"default_profile"`
		ProfileStatus              any                             `json:"profile_status"`
		SubstratePosture           string                          `json:"substrate_posture"`
		SubstrateStatus            string                          `json:"substrate_status"`
		SubstrateReason            string                          `json:"substrate_reason"`
		SubstrateSelectionStrategy string                          `json:"substrate_selection_strategy"`
		SubstrateSelectionReason   string                          `json:"substrate_selection_reason"`
		DefaultRoute               browserRuntimeRouteDescriptor   `json:"default_route"`
		DefaultCandidateRoute      browserRuntimeRouteDescriptor   `json:"default_candidate_route"`
		SelectedRoute              *browserRuntimeRouteDescriptor  `json:"selected_route"`
		RouteResolution            any                             `json:"route_resolution"`
		Doctor                     *BrowserDoctorSummary           `json:"doctor"`
		Explanation                *browserTopLevelSummary         `json:"explanation"`
		Diagnostics                *browserTopLevelSummary         `json:"diagnostics"`
		Summary                    *browserTopLevelSummary         `json:"summary"`
		Display                    *browserTopLevelDisplaySummary  `json:"display"`
		Surface                    *browserTopLevelSurfaceSummary  `json:"surface"`
		View                       *browserTopLevelViewSummary     `json:"view"`
		BrowserSurface             string                          `json:"browser_surface"`
		BrowserOptInTargets        []string                        `json:"browser_opt_in_targets"`
		SubstrateMatrix            []browserRuntimeSubstrateStatus `json:"substrate_matrix"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_runtime doctor doctor route gate output: %v", err)
	}

	if payload.Action != "status" || payload.Status != "ok" {
		t.Fatalf("unexpected browser_runtime doctor payload: %#v", payload)
	}
	if payload.DefaultRoute.Backend != "proxy" || payload.DefaultRoute.Profile != "isolated" || payload.DefaultRoute.RuntimeTarget != "node" {
		t.Fatalf("expected doctor-route gate to promote managed candidate default_route, got %#v", payload.DefaultRoute)
	}
	if payload.DefaultCandidateRoute.Backend != "proxy" || payload.DefaultCandidateRoute.Profile != "isolated" || payload.DefaultCandidateRoute.RuntimeTarget != "node" {
		t.Fatalf("expected doctor-route gate to expose hidden managed default_candidate_route, got %#v", payload.DefaultCandidateRoute)
	}
	if payload.SelectedRoute != nil || payload.RouteResolution != nil {
		t.Fatalf("expected hidden legacy-host default route to stay out of top-level route selection, got selected=%#v resolution=%#v", payload.SelectedRoute, payload.RouteResolution)
	}
	if payload.DefaultProfile != "" || payload.ProfileStatus != nil {
		t.Fatalf("expected hidden legacy-host profile summary to stay out of runtime doctor payload, got default_profile=%q profile_status=%#v", payload.DefaultProfile, payload.ProfileStatus)
	}
	if payload.Doctor == nil || payload.Doctor.Route == nil {
		t.Fatalf("expected doctor route summary, got %#v", payload.Doctor)
	}
	if payload.Doctor.Route.Code != "managed_route_hidden_by_legacy_host_default" || payload.Doctor.Route.Status != "warn" {
		t.Fatalf("unexpected doctor route summary: %#v", payload.Doctor.Route)
	}
	if payload.Doctor.Route.SelectionStrategy != BrowserSubstrateSelectionPreferNodeOverLegacy {
		t.Fatalf("expected doctor route selection_strategy to align with managed candidate default, got route=%#v", payload.Doctor.Route)
	}
	if payload.SubstrateSelectionStrategy != BrowserSubstrateSelectionPreferNodeOverLegacy {
		t.Fatalf("expected doctor substrate_selection_strategy to align with managed candidate default, got strategy=%q payload=%#v", payload.SubstrateSelectionStrategy, payload)
	}
	if payload.SubstrateSelectionReason != strings.TrimSpace(payload.Doctor.Route.Summary) {
		t.Fatalf("expected doctor substrate_selection_reason to align with doctor route summary, got reason=%q route=%#v", payload.SubstrateSelectionReason, payload.Doctor.Route)
	}
	if payload.SubstratePosture != BrowserSubstrateNodeRuntime ||
		payload.SubstrateStatus != "ok" ||
		payload.SubstrateReason != strings.TrimSpace(payload.Doctor.Route.Summary) {
		t.Fatalf("expected doctor top-level substrate summary to align with hidden managed candidate, got posture=%q status=%q reason=%q payload=%#v", payload.SubstratePosture, payload.SubstrateStatus, payload.SubstrateReason, payload)
	}
	if !strings.Contains(payload.Note, strings.TrimSpace(payload.Doctor.Route.Summary)) {
		t.Fatalf("expected note to surface hidden managed route guidance, got %#v", payload.Note)
	}
	assertManagedRouteDoctorGateExplanationSummary(t, payload.Explanation)
	assertManagedRouteDoctorGateTopLevelSummary(t, payload.Diagnostics)
	assertManagedRouteDoctorGateTopLevelSummary(t, payload.Summary)
	if payload.Summary == nil || payload.Summary.DefaultCandidateRoute != payload.DefaultCandidateRoute {
		t.Fatalf("expected doctor summary to expose hidden managed default_candidate_route, got %#v", payload.Summary)
	}
	assertManagedRouteDoctorGateDisplaySummary(t, payload.Display, false)
	assertManagedRouteDoctorGateSurfaceSummary(t, payload.Surface)
	assertManagedRouteDoctorGateViewSummary(t, payload.View)
	assertManagedRouteDoctorGateRootSurface(t, payload.BrowserSurface, payload.BrowserOptInTargets)
	assertManagedRouteDoctorGateSubstrateMatrix(t, payload.SubstrateMatrix, strings.TrimSpace(payload.Doctor.Route.Summary))
}

func TestRegisterBrowserTools_RuntimePrepareUsesDoctorRouteSummaryWhenManagedRouteStaysHidden(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-prepare-route-gate")
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-prepare-route-gate", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Status:        "running",
		Running:       true,
		Connected:     true,
		Note:          "cached implicit host default profile",
	})
	sessionStateRegistry.SelectBrowserProfile("browser-runtime-prepare-route-gate", agentxbrowserruntime.SharedSessionBrowserProfileSelection{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Source:        "remember_profile",
	})
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root: t.TempDir(),
		NodeBackend: browserHiddenManagedRouteDoctorGateResolvableRuntimeControlNodeBackend(BrowserCapabilities{
			RuntimeStatus:  true,
			RuntimePrepare: true,
			Open:           true,
		}),
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime", "browser_open"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"prepare"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime prepare doctor route gate: %v", err)
	}
	var payload struct {
		Action                     string                          `json:"action"`
		Status                     string                          `json:"status"`
		Note                       string                          `json:"note"`
		DefaultProfile             string                          `json:"default_profile"`
		ProfileStatus              any                             `json:"profile_status"`
		SubstratePosture           string                          `json:"substrate_posture"`
		SubstrateStatus            string                          `json:"substrate_status"`
		SubstrateReason            string                          `json:"substrate_reason"`
		SubstrateSelectionStrategy string                          `json:"substrate_selection_strategy"`
		SubstrateSelectionReason   string                          `json:"substrate_selection_reason"`
		DefaultRoute               browserRuntimeRouteDescriptor   `json:"default_route"`
		DefaultCandidateRoute      browserRuntimeRouteDescriptor   `json:"default_candidate_route"`
		SelectedRoute              *browserRuntimeRouteDescriptor  `json:"selected_route"`
		RouteResolution            any                             `json:"route_resolution"`
		Doctor                     *BrowserDoctorSummary           `json:"doctor"`
		Explanation                *browserTopLevelSummary         `json:"explanation"`
		Diagnostics                *browserTopLevelSummary         `json:"diagnostics"`
		Summary                    *browserTopLevelSummary         `json:"summary"`
		Display                    *browserTopLevelDisplaySummary  `json:"display"`
		Surface                    *browserTopLevelSurfaceSummary  `json:"surface"`
		View                       *browserTopLevelViewSummary     `json:"view"`
		BrowserSurface             string                          `json:"browser_surface"`
		BrowserOptInTargets        []string                        `json:"browser_opt_in_targets"`
		SubstrateMatrix            []browserRuntimeSubstrateStatus `json:"substrate_matrix"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_runtime prepare doctor route gate output: %v", err)
	}

	if payload.Action != "prepare" || payload.Status != "unsupported" {
		t.Fatalf("unexpected browser_runtime prepare payload: %#v", payload)
	}
	if payload.DefaultRoute.Backend != "proxy" || payload.DefaultRoute.Profile != "isolated" || payload.DefaultRoute.RuntimeTarget != "node" {
		t.Fatalf("expected doctor-route gate to promote managed candidate default_route, got %#v", payload.DefaultRoute)
	}
	if payload.DefaultCandidateRoute.Backend != "proxy" || payload.DefaultCandidateRoute.Profile != "isolated" || payload.DefaultCandidateRoute.RuntimeTarget != "node" {
		t.Fatalf("expected doctor-route gate to expose hidden managed default_candidate_route, got %#v", payload.DefaultCandidateRoute)
	}
	if payload.SelectedRoute != nil || payload.RouteResolution != nil {
		t.Fatalf("expected hidden legacy-host default route to stay out of top-level route selection, got selected=%#v resolution=%#v", payload.SelectedRoute, payload.RouteResolution)
	}
	if payload.DefaultProfile != "" || payload.ProfileStatus != nil {
		t.Fatalf("expected hidden legacy-host profile summary to stay out of runtime prepare payload, got default_profile=%q profile_status=%#v", payload.DefaultProfile, payload.ProfileStatus)
	}
	if payload.Doctor == nil || payload.Doctor.Route == nil {
		t.Fatalf("expected doctor route summary, got %#v", payload.Doctor)
	}
	if payload.Doctor.Route.Code != "managed_route_hidden_by_legacy_host_default" || payload.Doctor.Route.Status != "warn" {
		t.Fatalf("unexpected doctor route summary: %#v", payload.Doctor.Route)
	}
	if payload.SubstrateSelectionStrategy != BrowserSubstrateSelectionPreferNodeOverLegacy {
		t.Fatalf("expected prepare substrate_selection_strategy to align with managed candidate default, got strategy=%q payload=%#v", payload.SubstrateSelectionStrategy, payload)
	}
	if payload.SubstrateSelectionReason != strings.TrimSpace(payload.Doctor.Route.Summary) {
		t.Fatalf("expected prepare substrate_selection_reason to align with doctor route summary, got reason=%q route=%#v", payload.SubstrateSelectionReason, payload.Doctor.Route)
	}
	if payload.SubstratePosture != BrowserSubstrateNodeRuntime ||
		payload.SubstrateStatus != "ok" ||
		payload.SubstrateReason != strings.TrimSpace(payload.Doctor.Route.Summary) {
		t.Fatalf("expected prepare top-level substrate summary to align with hidden managed candidate, got posture=%q status=%q reason=%q payload=%#v", payload.SubstratePosture, payload.SubstrateStatus, payload.SubstrateReason, payload)
	}
	assertManagedRouteDoctorGateExplanationSummary(t, payload.Explanation)
	assertManagedRouteDoctorGateTopLevelSummary(t, payload.Diagnostics)
	assertManagedRouteDoctorGateTopLevelSummary(t, payload.Summary)
	assertManagedRouteDoctorGateDisplaySummary(t, payload.Display, false)
	assertManagedRouteDoctorGateSurfaceSummary(t, payload.Surface)
	assertManagedRouteDoctorGateViewSummary(t, payload.View)
	assertManagedRouteDoctorGateRootSurface(t, payload.BrowserSurface, payload.BrowserOptInTargets)
	assertManagedRouteDoctorGateSubstrateMatrix(t, payload.SubstrateMatrix, strings.TrimSpace(payload.Doctor.Route.Summary))
}

func TestRegisterBrowserTools_RuntimeCoordinateUsesDoctorRouteSummaryWhenManagedRouteStaysHidden(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-coordinate-route-gate")
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-coordinate-route-gate", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Status:        "running",
		Running:       true,
		Connected:     true,
		Note:          "cached implicit host default profile",
	})
	sessionStateRegistry.SelectBrowserProfile("browser-runtime-coordinate-route-gate", agentxbrowserruntime.SharedSessionBrowserProfileSelection{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Source:        "remember_profile",
	})
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root: t.TempDir(),
		NodeBackend: browserHiddenManagedRouteDoctorGateResolvableRuntimeControlNodeBackend(BrowserCapabilities{
			RuntimeCoordinate: true,
			Open:              true,
		}),
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime", "browser_open"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"coordinate","coordination_goal":"ensure"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime coordinate doctor route gate: %v", err)
	}
	var payload struct {
		Action                     string                          `json:"action"`
		Status                     string                          `json:"status"`
		Note                       string                          `json:"note"`
		DefaultRoute               browserRuntimeRouteDescriptor   `json:"default_route"`
		DefaultCandidateRoute      browserRuntimeRouteDescriptor   `json:"default_candidate_route"`
		SelectedRoute              *browserRuntimeRouteDescriptor  `json:"selected_route"`
		RouteResolution            any                             `json:"route_resolution"`
		Doctor                     *BrowserDoctorSummary           `json:"doctor"`
		Explanation                *browserTopLevelSummary         `json:"explanation"`
		Diagnostics                *browserTopLevelSummary         `json:"diagnostics"`
		Summary                    *browserTopLevelSummary         `json:"summary"`
		Display                    *browserTopLevelDisplaySummary  `json:"display"`
		Surface                    *browserTopLevelSurfaceSummary  `json:"surface"`
		View                       *browserTopLevelViewSummary     `json:"view"`
		BrowserSurface             string                          `json:"browser_surface"`
		BrowserOptInTargets        []string                        `json:"browser_opt_in_targets"`
		SubstrateSelectionStrategy string                          `json:"substrate_selection_strategy"`
		SubstrateSelectionReason   string                          `json:"substrate_selection_reason"`
		SubstratePosture           string                          `json:"substrate_posture"`
		SubstrateStatus            string                          `json:"substrate_status"`
		SubstrateReason            string                          `json:"substrate_reason"`
		SubstrateMatrix            []browserRuntimeSubstrateStatus `json:"substrate_matrix"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_runtime coordinate doctor route gate output: %v", err)
	}
	if payload.Action != "coordinate" || payload.Status != "unsupported" {
		t.Fatalf("unexpected browser_runtime coordinate payload: %#v", payload)
	}
	if payload.DefaultRoute.Backend != "proxy" || payload.DefaultRoute.Profile != "isolated" || payload.DefaultRoute.RuntimeTarget != "node" {
		t.Fatalf("expected coordinate doctor-route gate to promote managed candidate default_route, got %#v", payload.DefaultRoute)
	}
	if payload.DefaultCandidateRoute.Backend != "proxy" || payload.DefaultCandidateRoute.Profile != "isolated" || payload.DefaultCandidateRoute.RuntimeTarget != "node" {
		t.Fatalf("expected coordinate doctor-route gate to expose hidden managed default_candidate_route, got %#v", payload.DefaultCandidateRoute)
	}
	if payload.SelectedRoute != nil || payload.RouteResolution != nil {
		t.Fatalf("expected coordinate doctor-route gate to keep hidden legacy-host route state scrubbed, got selected=%#v resolution=%#v", payload.SelectedRoute, payload.RouteResolution)
	}
	if payload.Doctor == nil || payload.Doctor.Route == nil || payload.Doctor.Route.Code != "managed_route_hidden_by_legacy_host_default" {
		t.Fatalf("expected coordinate doctor route summary, got %#v", payload.Doctor)
	}
	if payload.SubstrateSelectionStrategy != BrowserSubstrateSelectionPreferNodeOverLegacy ||
		payload.SubstrateSelectionReason != strings.TrimSpace(payload.Doctor.Route.Summary) {
		t.Fatalf("expected coordinate substrate selection to align with doctor route summary, got payload=%#v", payload)
	}
	if payload.SubstratePosture != BrowserSubstrateNodeRuntime ||
		payload.SubstrateStatus != "ok" ||
		payload.SubstrateReason != strings.TrimSpace(payload.Doctor.Route.Summary) {
		t.Fatalf("expected coordinate top-level substrate summary to align with hidden managed candidate, got posture=%q status=%q reason=%q payload=%#v", payload.SubstratePosture, payload.SubstrateStatus, payload.SubstrateReason, payload)
	}
	assertManagedRouteDoctorGateExplanationSummary(t, payload.Explanation)
	assertManagedRouteDoctorGateTopLevelSummary(t, payload.Diagnostics)
	assertManagedRouteDoctorGateTopLevelSummary(t, payload.Summary)
	assertManagedRouteDoctorGateDisplaySummary(t, payload.Display, false)
	assertManagedRouteDoctorGateSurfaceSummary(t, payload.Surface)
	assertManagedRouteDoctorGateViewSummary(t, payload.View)
	assertManagedRouteDoctorGateRootSurface(t, payload.BrowserSurface, payload.BrowserOptInTargets)
	assertManagedRouteDoctorGateSubstrateMatrix(t, payload.SubstrateMatrix, strings.TrimSpace(payload.Doctor.Route.Summary))
}

func TestRegisterBrowserTools_RuntimeStartUsesDoctorRouteSummaryWhenManagedRouteStaysHidden(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-start-route-gate")
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-start-route-gate", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Status:        "running",
		Running:       true,
		Connected:     true,
		Note:          "cached implicit host default profile",
	})
	sessionStateRegistry.SelectBrowserProfile("browser-runtime-start-route-gate", agentxbrowserruntime.SharedSessionBrowserProfileSelection{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Source:        "remember_profile",
	})
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root: t.TempDir(),
		NodeBackend: browserHiddenManagedRouteDoctorGateResolvableRuntimeControlNodeBackend(BrowserCapabilities{
			RuntimeStart: true,
			Open:         true,
		}),
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime", "browser_open"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"start"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime start doctor route gate: %v", err)
	}
	var payload struct {
		Action                     string                          `json:"action"`
		Status                     string                          `json:"status"`
		Note                       string                          `json:"note"`
		DefaultRoute               browserRuntimeRouteDescriptor   `json:"default_route"`
		DefaultCandidateRoute      browserRuntimeRouteDescriptor   `json:"default_candidate_route"`
		SelectedRoute              *browserRuntimeRouteDescriptor  `json:"selected_route"`
		RouteResolution            any                             `json:"route_resolution"`
		Doctor                     *BrowserDoctorSummary           `json:"doctor"`
		Explanation                *browserTopLevelSummary         `json:"explanation"`
		Diagnostics                *browserTopLevelSummary         `json:"diagnostics"`
		Summary                    *browserTopLevelSummary         `json:"summary"`
		Display                    *browserTopLevelDisplaySummary  `json:"display"`
		Surface                    *browserTopLevelSurfaceSummary  `json:"surface"`
		View                       *browserTopLevelViewSummary     `json:"view"`
		BrowserSurface             string                          `json:"browser_surface"`
		BrowserOptInTargets        []string                        `json:"browser_opt_in_targets"`
		SubstrateSelectionStrategy string                          `json:"substrate_selection_strategy"`
		SubstrateSelectionReason   string                          `json:"substrate_selection_reason"`
		SubstratePosture           string                          `json:"substrate_posture"`
		SubstrateStatus            string                          `json:"substrate_status"`
		SubstrateReason            string                          `json:"substrate_reason"`
		SubstrateMatrix            []browserRuntimeSubstrateStatus `json:"substrate_matrix"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_runtime start doctor route gate output: %v", err)
	}
	if payload.Action != "start" || payload.Status != "unsupported" {
		t.Fatalf("unexpected browser_runtime start payload: %#v", payload)
	}
	if payload.DefaultRoute.Backend != "proxy" || payload.DefaultRoute.Profile != "isolated" || payload.DefaultRoute.RuntimeTarget != "node" {
		t.Fatalf("expected start doctor-route gate to promote managed candidate default_route, got %#v", payload.DefaultRoute)
	}
	if payload.DefaultCandidateRoute.Backend != "proxy" || payload.DefaultCandidateRoute.Profile != "isolated" || payload.DefaultCandidateRoute.RuntimeTarget != "node" {
		t.Fatalf("expected start doctor-route gate to expose hidden managed default_candidate_route, got %#v", payload.DefaultCandidateRoute)
	}
	if payload.SelectedRoute != nil || payload.RouteResolution != nil {
		t.Fatalf("expected start doctor-route gate to keep hidden legacy-host route state scrubbed, got selected=%#v resolution=%#v", payload.SelectedRoute, payload.RouteResolution)
	}
	if payload.Doctor == nil || payload.Doctor.Route == nil || payload.Doctor.Route.Code != "managed_route_hidden_by_legacy_host_default" {
		t.Fatalf("expected start doctor route summary, got %#v", payload.Doctor)
	}
	if payload.SubstrateSelectionStrategy != BrowserSubstrateSelectionPreferNodeOverLegacy ||
		payload.SubstrateSelectionReason != strings.TrimSpace(payload.Doctor.Route.Summary) {
		t.Fatalf("expected start substrate selection to align with doctor route summary, got payload=%#v", payload)
	}
	if payload.SubstratePosture != BrowserSubstrateNodeRuntime ||
		payload.SubstrateStatus != "ok" ||
		payload.SubstrateReason != strings.TrimSpace(payload.Doctor.Route.Summary) {
		t.Fatalf("expected start top-level substrate summary to align with hidden managed candidate, got posture=%q status=%q reason=%q payload=%#v", payload.SubstratePosture, payload.SubstrateStatus, payload.SubstrateReason, payload)
	}
	assertManagedRouteDoctorGateExplanationSummary(t, payload.Explanation)
	assertManagedRouteDoctorGateTopLevelSummary(t, payload.Diagnostics)
	assertManagedRouteDoctorGateTopLevelSummary(t, payload.Summary)
	assertManagedRouteDoctorGateDisplaySummary(t, payload.Display, false)
	assertManagedRouteDoctorGateSurfaceSummary(t, payload.Surface)
	assertManagedRouteDoctorGateViewSummary(t, payload.View)
	assertManagedRouteDoctorGateRootSurface(t, payload.BrowserSurface, payload.BrowserOptInTargets)
	assertManagedRouteDoctorGateSubstrateMatrix(t, payload.SubstrateMatrix, strings.TrimSpace(payload.Doctor.Route.Summary))
}

func TestRegisterBrowserTools_RuntimeWorkbenchUsesDoctorRouteSummaryWhenManagedRouteStaysHidden(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-workbench-doctor-route-gate")
	tracked := sessionRegistry.TrackCurrentTarget("browser-runtime-workbench-doctor-route-gate", BrowserSessionTarget{
		ID:         "host-current",
		TabIndex:   1,
		URL:        "https://host.example/current",
		Title:      "Host Current",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
	}, "tracked_active_tab")
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-workbench-doctor-route-gate", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Status:        "running",
		Running:       true,
		Connected:     true,
		Note:          "cached implicit host default profile",
	})
	sessionStateRegistry.SelectBrowserProfile("browser-runtime-workbench-doctor-route-gate", agentxbrowserruntime.SharedSessionBrowserProfileSelection{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Source:        "remember_profile",
	})
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		NodeBackend:          browserHiddenManagedRouteDoctorGateNodeBackend(),
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime", "browser_open"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"workbench"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime workbench doctor route gate: %v", err)
	}
	var payload struct {
		Action                             string                                       `json:"action"`
		Status                             string                                       `json:"status"`
		Note                               string                                       `json:"note"`
		SubstratePosture                   string                                       `json:"substrate_posture"`
		SubstrateStatus                    string                                       `json:"substrate_status"`
		SubstrateReason                    string                                       `json:"substrate_reason"`
		SubstrateSelectionStrategy         string                                       `json:"substrate_selection_strategy"`
		SubstrateSelectionReason           string                                       `json:"substrate_selection_reason"`
		DefaultRoute                       browserRuntimeRouteDescriptor                `json:"default_route"`
		DefaultCandidateRoute              browserRuntimeRouteDescriptor                `json:"default_candidate_route"`
		Doctor                             *BrowserDoctorSummary                        `json:"doctor"`
		WorkbenchReady                     bool                                         `json:"workbench_ready"`
		WorkbenchSections                  []string                                     `json:"workbench_sections"`
		WorkbenchPrimaryBrowserAction      string                                       `json:"workbench_primary_browser_action"`
		WorkbenchPrimaryNodeAction         string                                       `json:"workbench_primary_node_action"`
		WorkbenchNextStep                  string                                       `json:"workbench_next_step"`
		WorkbenchRecommendedBrowserActions []string                                     `json:"workbench_recommended_browser_actions"`
		WorkbenchExplanation               *browserRuntimeDiagnosticsExplanationSummary `json:"workbench_explanation"`
		WorkbenchDiagnostics               *browserRuntimeWorkbenchDiagnosticsSummary   `json:"workbench_diagnostics"`
		WorkbenchSummary                   *browserTopLevelSummary                      `json:"workbench_summary"`
		WorkbenchDisplay                   *browserRuntimeWorkbenchDisplaySummary       `json:"workbench_display"`
		Workbench                          *browserRuntimeWorkbenchSurfaceSummary       `json:"workbench"`
		Diagnostics                        *browserTopLevelSummary                      `json:"diagnostics"`
		Summary                            *browserTopLevelSummary                      `json:"summary"`
		Display                            *browserTopLevelDisplaySummary               `json:"display"`
		Surface                            *browserTopLevelSurfaceSummary               `json:"surface"`
		View                               *browserTopLevelViewSummary                  `json:"view"`
		BrowserSurface                     string                                       `json:"browser_surface"`
		BrowserOptInTargets                []string                                     `json:"browser_opt_in_targets"`
		SessionRoutes                      []browserRuntimeSessionRoute                 `json:"session_routes"`
		SubstrateMatrix                    []browserRuntimeSubstrateStatus              `json:"substrate_matrix"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_runtime workbench doctor route gate output: %v", err)
	}
	if payload.Action != "workbench" || payload.Status != "ok" {
		t.Fatalf("unexpected browser_runtime workbench payload: %#v", payload)
	}
	if payload.DefaultRoute.Backend != "proxy" || payload.DefaultRoute.Profile != "isolated" || payload.DefaultRoute.RuntimeTarget != "node" {
		t.Fatalf("expected workbench doctor-route gate to promote managed candidate default_route, got %#v", payload.DefaultRoute)
	}
	if payload.Doctor == nil || payload.Doctor.Route == nil {
		t.Fatalf("expected workbench doctor route summary, got %#v", payload.Doctor)
	}
	if payload.Doctor.Route.Code != "managed_route_hidden_by_legacy_host_default" || payload.Doctor.Route.Status != "warn" {
		t.Fatalf("unexpected workbench doctor route summary: %#v", payload.Doctor.Route)
	}
	if payload.Doctor.Route.SelectionStrategy != BrowserSubstrateSelectionPreferNodeOverLegacy {
		t.Fatalf("expected workbench doctor route selection_strategy to align with managed candidate default, got route=%#v", payload.Doctor.Route)
	}
	if payload.SubstrateSelectionStrategy != BrowserSubstrateSelectionPreferNodeOverLegacy {
		t.Fatalf("expected workbench substrate_selection_strategy to align with managed candidate default, got strategy=%q payload=%#v", payload.SubstrateSelectionStrategy, payload)
	}
	if payload.SubstrateSelectionReason != strings.TrimSpace(payload.Doctor.Route.Summary) {
		t.Fatalf("expected workbench substrate_selection_reason to align with doctor route summary, got reason=%q route=%#v", payload.SubstrateSelectionReason, payload.Doctor.Route)
	}
	if payload.SubstratePosture != BrowserSubstrateNodeRuntime ||
		payload.SubstrateStatus != "ok" ||
		payload.SubstrateReason != strings.TrimSpace(payload.Doctor.Route.Summary) {
		t.Fatalf("expected workbench top-level substrate summary to align with hidden managed candidate, got posture=%q status=%q reason=%q payload=%#v", payload.SubstratePosture, payload.SubstrateStatus, payload.SubstrateReason, payload)
	}
	if !strings.Contains(payload.Note, strings.TrimSpace(payload.Doctor.Route.Summary)) {
		t.Fatalf("expected workbench note to surface hidden managed route guidance, got %#v", payload.Note)
	}
	if !payload.WorkbenchReady || !browserStringSliceContains(payload.WorkbenchSections, "route") {
		t.Fatalf("expected workbench route section to stay visible, got ready=%v sections=%#v", payload.WorkbenchReady, payload.WorkbenchSections)
	}
	if payload.WorkbenchPrimaryBrowserAction != runtimeOnlyDoctorRouteNextStep ||
		payload.WorkbenchPrimaryNodeAction != "" ||
		payload.WorkbenchNextStep != runtimeOnlyDoctorRouteNextStep {
		t.Fatalf("expected workbench to promote runtime prepare guidance, got %#v", payload)
	}
	if !browserStringSliceContains(payload.WorkbenchRecommendedBrowserActions, runtimeOnlyDoctorRouteNextStep) {
		t.Fatalf("expected workbench to recommend runtime prepare guidance, got %#v", payload.WorkbenchRecommendedBrowserActions)
	}
	assertManagedRouteDoctorGateWorkbenchExplanationSummary(t, payload.WorkbenchExplanation)
	assertManagedRouteDoctorGateWorkbenchDiagnosticsSummary(t, payload.WorkbenchDiagnostics)
	assertManagedRouteDoctorGateTopLevelSummary(t, payload.WorkbenchSummary)
	if payload.WorkbenchSummary == nil || payload.WorkbenchSummary.DefaultCandidateRoute != payload.DefaultCandidateRoute {
		t.Fatalf("expected workbench summary to expose hidden managed default_candidate_route, got %#v", payload.WorkbenchSummary)
	}
	assertManagedRouteDoctorGateWorkbenchDisplaySummary(t, payload.WorkbenchDisplay)
	if payload.Workbench == nil ||
		!payload.Workbench.Ready ||
		!browserStringSliceContains(payload.Workbench.Sections, "route") ||
		payload.Workbench.Diagnostics == nil ||
		payload.Workbench.Summary == nil ||
		payload.Workbench.DefaultCandidateRoute != (browserRuntimeRouteDescriptor{Backend: "proxy", Profile: "isolated", RuntimeTarget: "node"}) ||
		payload.Workbench.BrowserSurface != "explicit_managed_opt_in" ||
		!reflect.DeepEqual(payload.Workbench.BrowserOptInTargets, []string{"node"}) ||
		payload.Workbench.PrimaryBrowserAction != runtimeOnlyDoctorRouteNextStep ||
		payload.Workbench.NextStep != runtimeOnlyDoctorRouteNextStep {
		t.Fatalf("unexpected workbench surface summary: %#v", payload.Workbench)
	}
	assertManagedRouteDoctorGateWorkbenchNestedSummary(t, payload.Workbench.Diagnostics)
	assertManagedRouteDoctorGateWorkbenchNestedSummary(t, payload.Workbench.Summary)
	assertManagedRouteDoctorGateTopLevelSummary(t, payload.Diagnostics)
	assertManagedRouteDoctorGateTopLevelSummary(t, payload.Summary)
	assertManagedRouteDoctorGateDisplaySummary(t, payload.Display, true)
	assertManagedRouteDoctorGateSurfaceSummary(t, payload.Surface)
	assertManagedRouteDoctorGateViewSummary(t, payload.View)
	assertManagedRouteDoctorGateRootSurface(t, payload.BrowserSurface, payload.BrowserOptInTargets)
	assertManagedRouteDoctorGateSubstrateMatrix(t, payload.SubstrateMatrix, strings.TrimSpace(payload.Doctor.Route.Summary))
	if len(payload.SessionRoutes) != 1 ||
		payload.SessionRoutes[0].Backend != "system" ||
		payload.SessionRoutes[0].Profile != "default" ||
		payload.SessionRoutes[0].RuntimeTarget != "host" ||
		payload.SessionRoutes[0].CurrentTargetID != tracked.ID {
		t.Fatalf("expected explicit host session route snapshot to remain visible, got %#v", payload.SessionRoutes)
	}
}

func TestRegisterBrowserTools_RuntimeProfilesUsesDoctorRouteSummaryWhenManagedRouteStaysHidden(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-profiles-doctor-route-gate")
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-profiles-doctor-route-gate", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Status:        "running",
		Running:       true,
		Connected:     true,
		Note:          "cached implicit host default profile",
	})
	sessionStateRegistry.SelectBrowserProfile("browser-runtime-profiles-doctor-route-gate", agentxbrowserruntime.SharedSessionBrowserProfileSelection{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Source:        "remember_profile",
	})
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		NodeBackend:          browserHiddenManagedRouteDoctorGateRuntimeControlNodeBackend(BrowserCapabilities{RuntimeList: true, Open: true}),
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime", "browser_open"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"profiles"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime profiles doctor route gate: %v", err)
	}
	var payload struct {
		Action                     string                          `json:"action"`
		Status                     string                          `json:"status"`
		Note                       string                          `json:"note"`
		SubstrateSelectionStrategy string                          `json:"substrate_selection_strategy"`
		SubstrateSelectionReason   string                          `json:"substrate_selection_reason"`
		DefaultRoute               browserRuntimeRouteDescriptor   `json:"default_route"`
		SelectedRoute              *browserRuntimeRouteDescriptor  `json:"selected_route"`
		RouteResolution            any                             `json:"route_resolution"`
		Doctor                     *BrowserDoctorSummary           `json:"doctor"`
		DefaultProfile             string                          `json:"default_profile"`
		Profiles                   []browserRuntimeProfileState    `json:"profiles"`
		Explanation                *browserTopLevelSummary         `json:"explanation"`
		Diagnostics                *browserTopLevelSummary         `json:"diagnostics"`
		Summary                    *browserTopLevelSummary         `json:"summary"`
		Display                    *browserTopLevelDisplaySummary  `json:"display"`
		Surface                    *browserTopLevelSurfaceSummary  `json:"surface"`
		View                       *browserTopLevelViewSummary     `json:"view"`
		BrowserSurface             string                          `json:"browser_surface"`
		BrowserOptInTargets        []string                        `json:"browser_opt_in_targets"`
		SubstrateMatrix            []browserRuntimeSubstrateStatus `json:"substrate_matrix"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_runtime profiles doctor route gate output: %v", err)
	}

	if payload.Action != "profiles" || payload.Status != "unsupported" {
		t.Fatalf("unexpected browser_runtime profiles payload: %#v", payload)
	}
	if payload.DefaultRoute.Backend != "proxy" || payload.DefaultRoute.Profile != "isolated" || payload.DefaultRoute.RuntimeTarget != "node" {
		t.Fatalf("expected profiles doctor-route gate to promote managed candidate default_route, got %#v", payload.DefaultRoute)
	}
	if payload.SelectedRoute != nil || payload.RouteResolution != nil {
		t.Fatalf("expected hidden legacy-host default route to stay out of top-level route selection, got selected=%#v resolution=%#v", payload.SelectedRoute, payload.RouteResolution)
	}
	if payload.Doctor == nil || payload.Doctor.Route == nil {
		t.Fatalf("expected profiles doctor route summary, got %#v", payload.Doctor)
	}
	if payload.Doctor.Route.Code != "managed_route_hidden_by_legacy_host_default" || payload.Doctor.Route.Status != "warn" {
		t.Fatalf("unexpected profiles doctor route summary: %#v", payload.Doctor.Route)
	}
	if payload.SubstrateSelectionStrategy != BrowserSubstrateSelectionPreferNodeOverLegacy {
		t.Fatalf("expected profiles substrate_selection_strategy to align with managed candidate default, got strategy=%q payload=%#v", payload.SubstrateSelectionStrategy, payload)
	}
	if payload.SubstrateSelectionReason != strings.TrimSpace(payload.Doctor.Route.Summary) {
		t.Fatalf("expected profiles substrate_selection_reason to align with doctor route summary, got reason=%q route=%#v", payload.SubstrateSelectionReason, payload.Doctor.Route)
	}
	if !strings.Contains(payload.Note, strings.TrimSpace(payload.Doctor.Route.Summary)) {
		t.Fatalf("expected profiles note to surface hidden managed route guidance, got %#v", payload.Note)
	}
	if payload.DefaultProfile != "" || len(payload.Profiles) != 0 {
		t.Fatalf("expected profiles top-level inventory to stay hidden behind doctor guidance, got %#v", payload)
	}
	assertManagedRouteDoctorGateExplanationSummary(t, payload.Explanation)
	assertManagedRouteDoctorGateTopLevelSummary(t, payload.Diagnostics)
	assertManagedRouteDoctorGateTopLevelSummary(t, payload.Summary)
	assertManagedRouteDoctorGateDisplaySummary(t, payload.Display, false)
	assertManagedRouteDoctorGateSurfaceSummary(t, payload.Surface)
	assertManagedRouteDoctorGateViewSummary(t, payload.View)
	assertManagedRouteDoctorGateRootSurface(t, payload.BrowserSurface, payload.BrowserOptInTargets)
	assertManagedRouteDoctorGateSubstrateMatrix(t, payload.SubstrateMatrix, strings.TrimSpace(payload.Doctor.Route.Summary))
}

func TestRegisterBrowserTools_RuntimeSessionsUsesDoctorRouteSummaryWhenManagedRouteStaysHidden(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-sessions-doctor-route-gate")
	tracked := sessionRegistry.TrackCurrentTarget("browser-runtime-sessions-doctor-route-gate", BrowserSessionTarget{
		ID:         "host-current",
		TabIndex:   1,
		URL:        "https://host.example/current",
		Title:      "Host Current",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
	}, "tracked_active_tab")
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-sessions-doctor-route-gate", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Status:        "running",
		Running:       true,
		Connected:     true,
		Note:          "cached implicit host default profile",
	})
	sessionStateRegistry.SelectBrowserProfile("browser-runtime-sessions-doctor-route-gate", agentxbrowserruntime.SharedSessionBrowserProfileSelection{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Source:        "remember_profile",
	})
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		NodeBackend:          browserHiddenManagedRouteDoctorGateRuntimeControlNodeBackend(BrowserCapabilities{RuntimeSessions: true, Open: true}),
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime", "browser_open"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"sessions"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime sessions doctor route gate: %v", err)
	}
	var payload struct {
		Action                     string                          `json:"action"`
		Status                     string                          `json:"status"`
		Note                       string                          `json:"note"`
		SubstrateSelectionStrategy string                          `json:"substrate_selection_strategy"`
		SubstrateSelectionReason   string                          `json:"substrate_selection_reason"`
		DefaultRoute               browserRuntimeRouteDescriptor   `json:"default_route"`
		SelectedRoute              *browserRuntimeRouteDescriptor  `json:"selected_route"`
		RouteResolution            any                             `json:"route_resolution"`
		Doctor                     *BrowserDoctorSummary           `json:"doctor"`
		Explanation                *browserTopLevelSummary         `json:"explanation"`
		Diagnostics                *browserTopLevelSummary         `json:"diagnostics"`
		Summary                    *browserTopLevelSummary         `json:"summary"`
		Display                    *browserTopLevelDisplaySummary  `json:"display"`
		Surface                    *browserTopLevelSurfaceSummary  `json:"surface"`
		View                       *browserTopLevelViewSummary     `json:"view"`
		BrowserSurface             string                          `json:"browser_surface"`
		BrowserOptInTargets        []string                        `json:"browser_opt_in_targets"`
		SessionRoutes              []browserRuntimeSessionRoute    `json:"session_routes"`
		SubstrateMatrix            []browserRuntimeSubstrateStatus `json:"substrate_matrix"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_runtime sessions doctor route gate output: %v", err)
	}

	if payload.Action != "sessions" || payload.Status != "ok" {
		t.Fatalf("unexpected browser_runtime sessions payload: %#v", payload)
	}
	if payload.DefaultRoute.Backend != "proxy" || payload.DefaultRoute.Profile != "isolated" || payload.DefaultRoute.RuntimeTarget != "node" {
		t.Fatalf("expected sessions doctor-route gate to promote managed candidate default_route, got %#v", payload.DefaultRoute)
	}
	if payload.SelectedRoute != nil || payload.RouteResolution != nil {
		t.Fatalf("expected hidden legacy-host default route to stay out of top-level route selection, got selected=%#v resolution=%#v", payload.SelectedRoute, payload.RouteResolution)
	}
	if payload.Doctor == nil || payload.Doctor.Route == nil {
		t.Fatalf("expected sessions doctor route summary, got %#v", payload.Doctor)
	}
	if payload.Doctor.Route.Code != "managed_route_hidden_by_legacy_host_default" || payload.Doctor.Route.Status != "warn" {
		t.Fatalf("unexpected sessions doctor route summary: %#v", payload.Doctor.Route)
	}
	if payload.SubstrateSelectionStrategy != BrowserSubstrateSelectionPreferNodeOverLegacy {
		t.Fatalf("expected sessions substrate_selection_strategy to align with managed candidate default, got strategy=%q payload=%#v", payload.SubstrateSelectionStrategy, payload)
	}
	if payload.SubstrateSelectionReason != strings.TrimSpace(payload.Doctor.Route.Summary) {
		t.Fatalf("expected sessions substrate_selection_reason to align with doctor route summary, got reason=%q route=%#v", payload.SubstrateSelectionReason, payload.Doctor.Route)
	}
	if !strings.Contains(payload.Note, strings.TrimSpace(payload.Doctor.Route.Summary)) {
		t.Fatalf("expected sessions note to surface hidden managed route guidance, got %#v", payload.Note)
	}
	assertManagedRouteDoctorGateExplanationSummary(t, payload.Explanation)
	assertManagedRouteDoctorGateTopLevelSummary(t, payload.Diagnostics)
	assertManagedRouteDoctorGateTopLevelSummary(t, payload.Summary)
	assertManagedRouteDoctorGateDisplaySummary(t, payload.Display, false)
	assertManagedRouteDoctorGateSurfaceSummary(t, payload.Surface)
	assertManagedRouteDoctorGateViewSummary(t, payload.View)
	assertManagedRouteDoctorGateRootSurface(t, payload.BrowserSurface, payload.BrowserOptInTargets)
	assertManagedRouteDoctorGateSubstrateMatrix(t, payload.SubstrateMatrix, strings.TrimSpace(payload.Doctor.Route.Summary))
	if len(payload.SessionRoutes) != 1 ||
		payload.SessionRoutes[0].Backend != "system" ||
		payload.SessionRoutes[0].Profile != "default" ||
		payload.SessionRoutes[0].RuntimeTarget != "host" ||
		payload.SessionRoutes[0].CurrentTargetID != tracked.ID {
		t.Fatalf("expected explicit host session route snapshot to remain visible, got %#v", payload.SessionRoutes)
	}
}

func browserHiddenManagedRouteDoctorGateNodeBackend() *runtimeInfoCapabilityRouteResolverBrowserBackend {
	return &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
			capabilities: BrowserCapabilities{
				RuntimeStatus: true,
				Open:          true,
			},
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			return BrowserRuntimeInfo{}, context.DeadlineExceeded
		},
	}
}

func browserHiddenManagedRouteDoctorGateRuntimeControlNodeBackend(capabilities BrowserCapabilities) *routeResolverCapabilityRuntimeControlBrowserBackend {
	return &routeResolverCapabilityRuntimeControlBrowserBackend{
		capabilityRuntimeControlBrowserBackend: &capabilityRuntimeControlBrowserBackend{
			runtimeControlBrowserBackend: &runtimeControlBrowserBackend{
				runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
					fakeBrowserBackend: &fakeBrowserBackend{},
					runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
				},
			},
			capabilities: capabilities,
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			return BrowserRuntimeInfo{}, context.DeadlineExceeded
		},
	}
}

func browserHiddenManagedRouteDoctorGateResolvableRuntimeControlNodeBackend(capabilities BrowserCapabilities) *routeResolverCapabilityRuntimeControlBrowserBackend {
	return &routeResolverCapabilityRuntimeControlBrowserBackend{
		capabilityRuntimeControlBrowserBackend: &capabilityRuntimeControlBrowserBackend{
			runtimeControlBrowserBackend: &runtimeControlBrowserBackend{
				runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
					fakeBrowserBackend: &fakeBrowserBackend{},
					runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
				},
			},
			capabilities: capabilities,
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			return BrowserRuntimeInfo{}, errors.New("managed_route_unavailable: hidden managed route remains explicit opt-in")
		},
	}
}

func assertManagedRouteDoctorGateExplanationSummary(t *testing.T, summary *browserTopLevelSummary) {
	t.Helper()
	if summary == nil ||
		summary.Category != "coordination" ||
		summary.State != "managed_route_pending_default" ||
		summary.SummaryCode != "managed_route_hidden_by_legacy_host_default" ||
		summary.NextStepAlias != runtimeOnlyDoctorRouteNextStepAlias ||
		summary.ManualRetryHint != "promote_managed_default" {
		t.Fatalf("unexpected doctor-route explanation summary: %#v", summary)
	}
}

func assertManagedRouteDoctorGateTopLevelSummary(t *testing.T, summary *browserTopLevelSummary) {
	t.Helper()
	if summary == nil ||
		summary.Category != "coordination" ||
		summary.State != "managed_route_pending_default" ||
		summary.SummaryCode != "managed_route_hidden_by_legacy_host_default" ||
		summary.NextStepAlias != runtimeOnlyDoctorRouteNextStepAlias ||
		summary.ManualRetryHint != "promote_managed_default" ||
		!summary.ResolvedViaFallback ||
		summary.PrimaryBrowserAction != runtimeOnlyDoctorRouteNextStep ||
		summary.PrimaryNodeAction != "" ||
		summary.NextStep != runtimeOnlyDoctorRouteNextStep {
		t.Fatalf("unexpected doctor-route top-level summary: %#v", summary)
	}
}

func assertManagedRouteDoctorGateDisplaySummary(t *testing.T, summary *browserTopLevelDisplaySummary, ready bool) {
	t.Helper()
	if summary == nil ||
		summary.Ready != ready ||
		summary.Category != "coordination" ||
		summary.State != "managed_route_pending_default" ||
		summary.SummaryCode != "managed_route_hidden_by_legacy_host_default" ||
		summary.DefaultCandidateRoute != (browserRuntimeRouteDescriptor{Backend: "proxy", Profile: "isolated", RuntimeTarget: "node"}) ||
		summary.NextStepAlias != runtimeOnlyDoctorRouteNextStepAlias ||
		summary.ManualRetryHint != "promote_managed_default" ||
		!summary.ResolvedViaFallback ||
		summary.PrimaryBrowserAction != runtimeOnlyDoctorRouteNextStep ||
		summary.PrimaryNodeAction != "" ||
		summary.NextStep != runtimeOnlyDoctorRouteNextStep {
		t.Fatalf("unexpected doctor-route display summary: %#v", summary)
	}
}

func assertManagedRouteDoctorGateSurfaceSummary(t *testing.T, summary *browserTopLevelSurfaceSummary) {
	t.Helper()
	if summary == nil ||
		summary.Category != "coordination" ||
		summary.State != "managed_route_pending_default" ||
		summary.SummaryCode != "managed_route_hidden_by_legacy_host_default" ||
		summary.DefaultCandidateRoute != (browserRuntimeRouteDescriptor{Backend: "proxy", Profile: "isolated", RuntimeTarget: "node"}) ||
		summary.NextStepAlias != runtimeOnlyDoctorRouteNextStepAlias ||
		summary.ManualRetryHint != "promote_managed_default" ||
		!summary.ResolvedViaFallback ||
		summary.PrimaryBrowserAction != runtimeOnlyDoctorRouteNextStep ||
		summary.PrimaryNodeAction != "" ||
		summary.NextStep != runtimeOnlyDoctorRouteNextStep ||
		summary.BrowserSurface != "explicit_managed_opt_in" ||
		!reflect.DeepEqual(summary.BrowserOptInTargets, []string{"node"}) {
		t.Fatalf("unexpected doctor-route surface summary: %#v", summary)
	}
}

func assertManagedRouteDoctorGateViewSummary(t *testing.T, summary *browserTopLevelViewSummary) {
	t.Helper()
	if summary == nil ||
		summary.Category != "coordination" ||
		summary.State != "managed_route_pending_default" ||
		summary.SummaryCode != "managed_route_hidden_by_legacy_host_default" ||
		summary.DefaultCandidateRoute != (browserRuntimeRouteDescriptor{Backend: "proxy", Profile: "isolated", RuntimeTarget: "node"}) ||
		summary.NextStepAlias != runtimeOnlyDoctorRouteNextStepAlias ||
		summary.ManualRetryHint != "promote_managed_default" ||
		!summary.ResolvedViaFallback ||
		summary.PrimaryBrowserAction != runtimeOnlyDoctorRouteNextStep ||
		summary.PrimaryNodeAction != "" ||
		summary.NextStep != runtimeOnlyDoctorRouteNextStep ||
		summary.BrowserSurface != "explicit_managed_opt_in" ||
		!reflect.DeepEqual(summary.BrowserOptInTargets, []string{"node"}) {
		t.Fatalf("unexpected doctor-route view summary: %#v", summary)
	}
}

func assertManagedRouteDoctorGateWorkbenchExplanationSummary(t *testing.T, summary *browserRuntimeDiagnosticsExplanationSummary) {
	t.Helper()
	if summary == nil ||
		summary.Category != "coordination" ||
		summary.State != "managed_route_pending_default" ||
		summary.SummaryCode != "managed_route_hidden_by_legacy_host_default" ||
		summary.NextStepAlias != runtimeOnlyDoctorRouteNextStepAlias ||
		summary.ManualRetryHint != "promote_managed_default" {
		t.Fatalf("unexpected doctor-route workbench explanation summary: %#v", summary)
	}
}

func assertManagedRouteDoctorGateWorkbenchDiagnosticsSummary(t *testing.T, summary *browserRuntimeWorkbenchDiagnosticsSummary) {
	t.Helper()
	if summary == nil ||
		summary.Category != "coordination" ||
		summary.State != "managed_route_pending_default" ||
		summary.SummaryCode != "managed_route_hidden_by_legacy_host_default" ||
		summary.NextStepAlias != runtimeOnlyDoctorRouteNextStepAlias ||
		summary.ManualRetryHint != "promote_managed_default" ||
		summary.PrimaryBrowserAction != runtimeOnlyDoctorRouteNextStep ||
		summary.PrimaryNodeAction != "" ||
		summary.NextStep != runtimeOnlyDoctorRouteNextStep {
		t.Fatalf("unexpected doctor-route workbench diagnostics summary: %#v", summary)
	}
}

func assertManagedRouteDoctorGateWorkbenchNestedSummary(t *testing.T, summary *browserTopLevelSummary) {
	t.Helper()
	if summary == nil ||
		summary.Category != "coordination" ||
		summary.State != "managed_route_pending_default" ||
		summary.SummaryCode != "managed_route_hidden_by_legacy_host_default" ||
		summary.DefaultCandidateRoute != (browserRuntimeRouteDescriptor{Backend: "proxy", Profile: "isolated", RuntimeTarget: "node"}) ||
		summary.NextStepAlias != runtimeOnlyDoctorRouteNextStepAlias ||
		summary.ManualRetryHint != "promote_managed_default" ||
		summary.PrimaryBrowserAction != runtimeOnlyDoctorRouteNextStep ||
		summary.PrimaryNodeAction != "" ||
		summary.NextStep != runtimeOnlyDoctorRouteNextStep {
		t.Fatalf("unexpected doctor-route nested workbench summary: %#v", summary)
	}
}

func assertManagedRouteDoctorGateWorkbenchDisplaySummary(t *testing.T, summary *browserRuntimeWorkbenchDisplaySummary) {
	t.Helper()
	if summary == nil ||
		!summary.Ready ||
		!browserStringSliceContains(summary.Sections, "route") ||
		summary.Category != "coordination" ||
		summary.State != "managed_route_pending_default" ||
		summary.SummaryCode != "managed_route_hidden_by_legacy_host_default" ||
		summary.DefaultCandidateRoute != (browserRuntimeRouteDescriptor{Backend: "proxy", Profile: "isolated", RuntimeTarget: "node"}) ||
		summary.NextStepAlias != runtimeOnlyDoctorRouteNextStepAlias ||
		summary.ManualRetryHint != "promote_managed_default" ||
		!summary.ResolvedViaFallback ||
		summary.PrimaryBrowserAction != runtimeOnlyDoctorRouteNextStep ||
		summary.PrimaryNodeAction != "" ||
		summary.NextStep != runtimeOnlyDoctorRouteNextStep {
		t.Fatalf("unexpected doctor-route workbench display summary: %#v", summary)
	}
}

func assertManagedRouteDoctorGateRootSurface(t *testing.T, surface string, targets []string) {
	t.Helper()
	if surface != "explicit_managed_opt_in" || !reflect.DeepEqual(targets, []string{"node"}) {
		t.Fatalf("unexpected doctor-route root surface metadata: surface=%q targets=%#v", surface, targets)
	}
}

func assertManagedRouteDoctorGateSubstrateMatrix(t *testing.T, matrix []browserRuntimeSubstrateStatus, note string) {
	t.Helper()
	foundDefault := false
	foundHostFallback := false
	for _, row := range matrix {
		switch row.Role {
		case "default":
			if row.Status != "unsupported" ||
				row.SelectionState != "default" ||
				row.Backend != "proxy" ||
				row.Profile != "isolated" ||
				row.RuntimeTarget != "node" ||
				!strings.Contains(row.Note, note) ||
				row.BrowserSurface != "explicit_managed_opt_in" ||
				!reflect.DeepEqual(row.BrowserOptInTargets, []string{"node"}) {
				t.Fatalf("unexpected doctor-route default substrate row: %#v", row)
			}
			foundDefault = true
		case "host":
			if row.Status == "available" &&
				row.SelectionState == "explicit_fallback" &&
				row.Backend == "system" &&
				row.Profile == "default" &&
				row.RuntimeTarget == "host" {
				foundHostFallback = true
			}
		}
	}
	if !foundDefault {
		t.Fatalf("expected doctor-route substrate matrix to include managed candidate default row, got %#v", matrix)
	}
	if !foundHostFallback {
		t.Fatalf("expected doctor-route substrate matrix to keep explicit host fallback row, got %#v", matrix)
	}
}
