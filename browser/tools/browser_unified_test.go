package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	types "github.com/wsnacj/agentx-go/components/llm"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

func TestBrowserUnifiedAvailableActionsIncludesAliasesAndDeduplicates(t *testing.T) {
	got := browserUnifiedAvailableActions(
		[]string{"status", "workbench", "prepare", "profiles", "sessions", "coordinate", "start", "stop", "sync_session", "create_profile", "delete_profile", "select_profile", "clear_profile", "select_target", "clear_target", "status"},
		[]string{"open", "list_tabs", "focus_tab", "close_tab", "save_pdf", "open"},
	)
	want := []string{"inspect", "doctor", "status", "workbench", "ready", "prepare", "inventory", "profiles", "handles", "sessions", "ensure", "refresh", "sync", "teardown", "reset", "coordinate", "launch", "start", "halt", "stop", "adopt", "sync_session", "new_profile", "create_profile", "remove_profile", "delete_profile", "pin_profile", "select_profile", "unpin_profile", "clear_profile", "pin_target", "select_target", "unpin_target", "clear_target", "open", "list_tabs", "focus_tab", "close_tab", "save_pdf", "tabs", "focus", "close", "pdf", "act"}
	if len(got) != len(want) {
		t.Fatalf("expected %d actions, got %#v", len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected actions %#v, got %#v", want, got)
		}
	}
}

func TestBrowserUnifiedActKindResolvesAliasesAndExplicitKind(t *testing.T) {
	if got := browserUnifiedActKind(map[string]any{}, "tabs"); got != "list_tabs" {
		t.Fatalf("expected tabs alias to resolve to list_tabs, got %q", got)
	}
	if got := browserUnifiedActKind(map[string]any{}, "pdf"); got != "save_pdf" {
		t.Fatalf("expected pdf alias to resolve to save_pdf, got %q", got)
	}
	if got := browserUnifiedActKind(map[string]any{"kind": "  CLICK  "}, "act"); got != "click" {
		t.Fatalf("expected explicit act kind to be normalized, got %q", got)
	}
	if got := browserUnifiedActKind(map[string]any{"kind": "\n\t\" SNAPSHOT \"\u007f"}, " \n`act`\t "); got != "snapshot" {
		t.Fatalf("expected quoted act kind to be sanitized, got %q", got)
	}
	if got := browserUnifiedActKind(map[string]any{}, "hover"); got != "hover" {
		t.Fatalf("expected passthrough act kind, got %q", got)
	}
}

func TestBrowserNormalizeToolTokenStripsWrappingQuotesAndControls(t *testing.T) {
	cases := map[string]string{
		"\n \"OPEN\"\t":     "open",
		"`snapshot`":        "snapshot",
		"\u007f' CLICK '\n": "click",
	}
	for raw, want := range cases {
		if got := browserNormalizeToolToken(raw); got != want {
			t.Fatalf("normalize %q: want %q got %q", raw, want, got)
		}
	}
}

func TestBrowserCanonicalSnapshotFormatNormalizesAliases(t *testing.T) {
	cases := map[string]string{
		"":             "",
		"ai":           "ai",
		"ARIA":         "aria",
		" text ":       "ai",
		"`accessible`": "aria",
		"unknown":      "ai",
	}
	for raw, want := range cases {
		if got := browserCanonicalSnapshotFormat(raw); got != want {
			t.Fatalf("snapshot format %q: want %q got %q", raw, want, got)
		}
	}
}

func TestCanonicalizeToolArguments_BrowserRecoversNoisySnapshotActionTokens(t *testing.T) {
	raw := "{\"action\":\"\\n\\n   \\\"snapshot\\\"\",\"format\":\"\\n `AI` \",\"mode\":\"\\n\\t\\\"EFFICIENT\\\"\\t\",\"refs\":\"\\n \\\"ROLE\\\" \",\"selector\":\"\\n `text=2025 年度报告` \"}"
	canonical, changed, err := CanonicalizeToolArguments("browser", raw)
	if err != nil {
		t.Fatalf("CanonicalizeToolArguments: %v", err)
	}
	if !changed {
		t.Fatalf("expected canonicalization change, got changed=%v canonical=%q", changed, canonical)
	}
	params, err := DecodeToolArguments(canonical)
	if err != nil {
		t.Fatalf("DecodeToolArguments canonical: %v", err)
	}
	if got := firstString(params, "action"); got != "snapshot" {
		t.Fatalf("expected action=snapshot, got %#v", params)
	}
	if got := firstString(params, "format"); got != "ai" {
		t.Fatalf("expected format=ai, got %#v", params)
	}
	if got := firstString(params, "mode"); got != "efficient" {
		t.Fatalf("expected mode=efficient, got %#v", params)
	}
	if got := firstString(params, "refs"); got != "role" {
		t.Fatalf("expected refs=role, got %#v", params)
	}
	if got := firstString(params, "selector"); got != "text=2025 年度报告" {
		t.Fatalf("expected selector to be sanitized, got %#v", params)
	}
}

func TestCanonicalizeToolArguments_BrowserCompatExtractRecoversNoisySelectorAndURL(t *testing.T) {
	raw := "{\"selector\":\"\\n `text=Annual report` \",\"url\":\"\\n \\\" https://93.184.216.34/report.pdf \\\" \"}"
	canonical, changed, err := CanonicalizeToolArguments("browser_extract", raw)
	if err != nil {
		t.Fatalf("CanonicalizeToolArguments: %v", err)
	}
	if !changed {
		t.Fatalf("expected canonicalization change, got changed=%v canonical=%q", changed, canonical)
	}
	params, err := DecodeToolArguments(canonical)
	if err != nil {
		t.Fatalf("DecodeToolArguments canonical: %v", err)
	}
	if got := firstString(params, "selector"); got != "text=Annual report" {
		t.Fatalf("expected selector to be sanitized, got %#v", params)
	}
	if got := firstString(params, "url"); got != "https://93.184.216.34/report.pdf" {
		t.Fatalf("expected url to be sanitized, got %#v", params)
	}
}

func TestBrowserUnifiedActArgsDropsActionAliasesAndSetsKind(t *testing.T) {
	input := map[string]any{
		"action":    "tabs",
		"operation": "ignored",
		"mode":      "ignored",
		"profile":   "isolated",
	}
	got := browserUnifiedActArgs(input, "tabs")
	if got["kind"] != "list_tabs" {
		t.Fatalf("expected list_tabs kind, got %#v", got)
	}
	if _, ok := got["action"]; ok {
		t.Fatalf("expected action to be removed, got %#v", got)
	}
	if _, ok := got["operation"]; ok {
		t.Fatalf("expected operation to be removed, got %#v", got)
	}
	if _, ok := got["mode"]; ok {
		t.Fatalf("expected mode to be removed, got %#v", got)
	}
	if got["profile"] != "isolated" {
		t.Fatalf("expected unrelated params to be preserved, got %#v", got)
	}
	if _, ok := input["action"]; !ok {
		t.Fatalf("expected source params to remain unchanged, got %#v", input)
	}
}

func TestBrowserUnifiedRuntimeArgsPreservesParamsAndOverridesAction(t *testing.T) {
	input := map[string]any{
		"action":         "open",
		"profile":        "relay",
		"runtime_target": "node",
	}
	got := browserUnifiedRuntimeArgs(input, "workbench")
	if got["action"] != "workbench" {
		t.Fatalf("expected runtime action override, got %#v", got)
	}
	if got["profile"] != "relay" || got["runtime_target"] != "node" {
		t.Fatalf("expected runtime params to be preserved, got %#v", got)
	}
	if input["action"] != "open" {
		t.Fatalf("expected source params to remain unchanged, got %#v", input)
	}
}

func TestBrowserUnifiedRuntimeArgsMapsCoordinationAliases(t *testing.T) {
	input := map[string]any{
		"profile":        "relay",
		"runtime_target": "node",
	}
	got := browserUnifiedRuntimeArgs(input, "ready")
	if got["action"] != "prepare" {
		t.Fatalf("expected ready alias to map to prepare, got %#v", got)
	}
	got = browserUnifiedRuntimeArgs(input, "inspect")
	if got["action"] != "status" {
		t.Fatalf("expected inspect alias to map to status, got %#v", got)
	}
	got = browserUnifiedRuntimeArgs(input, "inventory")
	if got["action"] != "profiles" {
		t.Fatalf("expected inventory alias to map to profiles, got %#v", got)
	}
	got = browserUnifiedRuntimeArgs(input, "handles")
	if got["action"] != "sessions" {
		t.Fatalf("expected handles alias to map to sessions, got %#v", got)
	}
	got = browserUnifiedRuntimeArgs(input, "launch")
	if got["action"] != "start" {
		t.Fatalf("expected launch alias to map to start, got %#v", got)
	}
	got = browserUnifiedRuntimeArgs(input, "halt")
	if got["action"] != "stop" {
		t.Fatalf("expected halt alias to map to stop, got %#v", got)
	}
	got = browserUnifiedRuntimeArgs(input, "sync")
	if got["action"] != "coordinate" || got["coordination_goal"] != "sync" {
		t.Fatalf("expected sync alias to map to coordinate/sync, got %#v", got)
	}
	got = browserUnifiedRuntimeArgs(input, "ensure")
	if got["action"] != "coordinate" || got["coordination_goal"] != "ensure" {
		t.Fatalf("expected ensure alias to map to coordinate/ensure, got %#v", got)
	}
	got = browserUnifiedRuntimeArgs(input, "refresh")
	if got["action"] != "refresh" {
		t.Fatalf("expected refresh alias to map to explicit refresh action, got %#v", got)
	}
	if _, ok := got["coordination_goal"]; ok {
		t.Fatalf("expected explicit refresh action to avoid synthetic coordination_goal, got %#v", got)
	}
	got = browserUnifiedRuntimeArgs(input, "teardown")
	if got["action"] != "coordinate" || got["coordination_goal"] != "teardown" {
		t.Fatalf("expected teardown alias to map to coordinate/teardown, got %#v", got)
	}
	got = browserUnifiedRuntimeArgs(input, "reset")
	if got["action"] != "clear_session" {
		t.Fatalf("expected reset alias to map to clear_session, got %#v", got)
	}
	got = browserUnifiedRuntimeArgs(input, "adopt")
	if got["action"] != "sync_session" {
		t.Fatalf("expected adopt alias to map to sync_session, got %#v", got)
	}
	got = browserUnifiedRuntimeArgs(input, "new_profile")
	if got["action"] != "create_profile" {
		t.Fatalf("expected new_profile alias to map to create_profile, got %#v", got)
	}
	got = browserUnifiedRuntimeArgs(input, "remove_profile")
	if got["action"] != "delete_profile" {
		t.Fatalf("expected remove_profile alias to map to delete_profile, got %#v", got)
	}
	got = browserUnifiedRuntimeArgs(input, "pin_profile")
	if got["action"] != "select_profile" {
		t.Fatalf("expected pin_profile alias to map to select_profile, got %#v", got)
	}
	got = browserUnifiedRuntimeArgs(input, "unpin_profile")
	if got["action"] != "clear_profile" {
		t.Fatalf("expected unpin_profile alias to map to clear_profile, got %#v", got)
	}
	got = browserUnifiedRuntimeArgs(input, "pin_target")
	if got["action"] != "select_target" {
		t.Fatalf("expected pin_target alias to map to select_target, got %#v", got)
	}
	got = browserUnifiedRuntimeArgs(input, "unpin_target")
	if got["action"] != "clear_target" {
		t.Fatalf("expected unpin_target alias to map to clear_target, got %#v", got)
	}
}

func TestBrowserUnifiedDefaultsToWorkbenchWhenAvailable(t *testing.T) {
	reg := llmxtools.NewRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "workbench",
				Status:     "running",
				Running:    true,
			},
			runtimeProfilesResult: BrowserProfilesResult{
				Backend: "proxy",
				Profiles: []BrowserProfileInfo{
					{Profile: "workbench", Status: "running", Running: true},
				},
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          nodeBackend,
		SessionRegistry:      NewBrowserSessionRegistry(),
		SessionStateRegistry: NewBrowserSessionStateRegistry(),
		EnabledTools:         []string{"browser"},
	})

	out, err := reg.Execute(WithToolSessionID(context.Background(), "browser-unified-default"), types.FunctionCall{
		Name:      "browser",
		Arguments: `{"runtime_target":"node","profile":"workbench"}`,
	})
	if err != nil {
		t.Fatalf("browser default workbench: %v", err)
	}
	var payload struct {
		Action string `json:"action"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Action != "workbench" || payload.Status != "ok" {
		t.Fatalf("expected unified browser to default to workbench, got %#v", payload)
	}
}

func TestBrowserUnifiedApplyExplanationAliasPrefersDiagnosticsExplanation(t *testing.T) {
	out, err := browserUnifiedApplyExplanationAlias(`{"status":"ok","diagnostics_explanation":{"category":"resolver","state":"manual_resolution_required","summary_code":"label_filtered_residual","next_step_alias":"snapshot","manual_retry_hint":"add_ordinal"},"resolver_explanation":{"state":"manual_resolution_required","summary_code":"label_filtered_residual","next_step_alias":"snapshot","manual_retry_hint":"add_ordinal"}}`)
	if err != nil {
		t.Fatalf("browserUnifiedApplyExplanationAlias: %v", err)
	}
	var payload struct {
		Explanation            *browserTopLevelSummary               `json:"explanation"`
		DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode aliased output: %v", err)
	}
	if payload.Explanation == nil ||
		payload.Explanation.Category != "resolver" ||
		payload.Explanation.State != "manual_resolution_required" ||
		payload.Explanation.SummaryCode != "label_filtered_residual" ||
		payload.Explanation.NextStepAlias != "snapshot" ||
		payload.Explanation.ManualRetryHint != "add_ordinal" ||
		payload.Explanation.ResolvedViaFallback {
		t.Fatalf("unexpected explanation alias: %#v", payload.Explanation)
	}
	if payload.DiagnosticsExplanation == nil || payload.DiagnosticsExplanation.Category != "resolver" {
		t.Fatalf("expected diagnostics_explanation to remain intact, got %#v", payload.DiagnosticsExplanation)
	}
}

func TestBrowserUnifiedApplyExplanationAliasBuildsSummaryAndDisplayFromSynthesizedDiagnostics(t *testing.T) {
	out, err := browserUnifiedApplyExplanationAlias(`{"status":"ok","diagnostics_explanation":{"category":"resolver","state":"manual_resolution_required","summary_code":"label_filtered_residual","next_step_alias":"snapshot","manual_retry_hint":"add_ordinal"}}`)
	if err != nil {
		t.Fatalf("browserUnifiedApplyExplanationAlias: %v", err)
	}
	var payload struct {
		Explanation *browserTopLevelSummary        `json:"explanation"`
		Diagnostics *browserTopLevelSummary        `json:"diagnostics"`
		Summary     *browserTopLevelSummary        `json:"summary"`
		Display     *browserTopLevelDisplaySummary `json:"display"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode aliased synthesized diagnostics output: %v", err)
	}
	if payload.Explanation == nil ||
		payload.Explanation.Category != "resolver" ||
		payload.Explanation.State != "manual_resolution_required" ||
		payload.Explanation.SummaryCode != "label_filtered_residual" ||
		payload.Explanation.NextStepAlias != "snapshot" ||
		payload.Explanation.ManualRetryHint != "add_ordinal" {
		t.Fatalf("unexpected synthesized explanation alias: %#v", payload.Explanation)
	}
	if payload.Diagnostics == nil ||
		payload.Diagnostics.Category != "resolver" ||
		payload.Diagnostics.State != "manual_resolution_required" ||
		payload.Diagnostics.SummaryCode != "label_filtered_residual" ||
		payload.Diagnostics.NextStepAlias != "snapshot" ||
		payload.Diagnostics.ManualRetryHint != "add_ordinal" {
		t.Fatalf("unexpected synthesized diagnostics alias: %#v", payload.Diagnostics)
	}
	if payload.Summary == nil ||
		payload.Summary.Category != "resolver" ||
		payload.Summary.State != "manual_resolution_required" ||
		payload.Summary.SummaryCode != "label_filtered_residual" ||
		payload.Summary.NextStepAlias != "snapshot" ||
		payload.Summary.ManualRetryHint != "add_ordinal" {
		t.Fatalf("unexpected synthesized summary alias: %#v", payload.Summary)
	}
	if payload.Display == nil ||
		payload.Display.Category != "resolver" ||
		payload.Display.State != "manual_resolution_required" ||
		payload.Display.SummaryCode != "label_filtered_residual" ||
		payload.Display.NextStepAlias != "snapshot" ||
		payload.Display.ManualRetryHint != "add_ordinal" {
		t.Fatalf("unexpected synthesized display alias: %#v", payload.Display)
	}
}

func TestBrowserUnifiedApplyTopLevelAliasProjectionBuildsAliasChain(t *testing.T) {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal([]byte(`{
		"status":"ok",
		"workbench_explanation":{"category":"resolver","state":"manual_resolution_required","summary_code":"label_filtered_residual","next_step_alias":"snapshot","manual_retry_hint":"add_ordinal"},
		"diagnostics_explanation":{"category":"resolver","state":"manual_resolution_required","summary_code":"label_filtered_residual","next_step_alias":"snapshot","manual_retry_hint":"add_ordinal"},
		"workbench":{"ready":true,"sections":["coordination"],"browser_surface":"explicit_managed_opt_in","browser_opt_in_targets":["node"],"summary":{"category":"coordination","state":"action_plan_available","summary_code":"workbench_action_plan"}},
		"browser_tools":["browser_runtime"],
		"artifact_tools":["browser.artifact.resolve"],
		"artifact_kinds":["download"],
		"artifact_contract":"artifacts+media",
		"browser_act_kinds":["open"]
	}`), &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	inputs := browserRuntimeTopLevelAliasInputsFromPayload(payload)

	mutated, err := browserUnifiedApplyTopLevelAliasProjection(payload, &inputs)
	if err != nil {
		t.Fatalf("browserUnifiedApplyTopLevelAliasProjection: %v", err)
	}
	if !mutated {
		t.Fatalf("expected top-level alias projection to mutate payload")
	}
	if !browserUnifiedHasNonNullJSONField(payload, "explanation") ||
		!browserUnifiedHasNonNullJSONField(payload, "diagnostics") ||
		!browserUnifiedHasNonNullJSONField(payload, "summary") ||
		!browserUnifiedHasNonNullJSONField(payload, "display") ||
		!browserUnifiedHasNonNullJSONField(payload, "surface") ||
		!browserUnifiedHasNonNullJSONField(payload, "view") {
		t.Fatalf("expected alias chain to populate explanation/diagnostics/summary/display/surface/view, got %#v", payload)
	}

	var aliased struct {
		Explanation *browserTopLevelSummary        `json:"explanation"`
		Diagnostics *browserTopLevelSummary        `json:"diagnostics"`
		Summary     *browserTopLevelSummary        `json:"summary"`
		Display     *browserTopLevelDisplaySummary `json:"display"`
		Surface     *browserTopLevelSurfaceSummary `json:"surface"`
		View        *browserTopLevelViewSummary    `json:"view"`
	}
	blob, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if err := json.Unmarshal(blob, &aliased); err != nil {
		t.Fatalf("decode aliased payload: %v", err)
	}
	if aliased.Explanation == nil || aliased.Explanation.Category != "resolver" {
		t.Fatalf("unexpected explanation alias: %#v", aliased.Explanation)
	}
	if aliased.Diagnostics == nil || aliased.Diagnostics.Category != "resolver" {
		t.Fatalf("unexpected diagnostics alias: %#v", aliased.Diagnostics)
	}
	if aliased.Summary == nil || aliased.Summary.Category != "resolver" {
		t.Fatalf("unexpected summary alias: %#v", aliased.Summary)
	}
	if aliased.Display == nil || aliased.Display.Category != "resolver" {
		t.Fatalf("unexpected display alias: %#v", aliased.Display)
	}
	if aliased.Surface == nil || aliased.Surface.BrowserSurface != "explicit_managed_opt_in" {
		t.Fatalf("unexpected surface alias: %#v", aliased.Surface)
	}
	if aliased.View == nil || aliased.View.BrowserSurface != "explicit_managed_opt_in" {
		t.Fatalf("unexpected view alias: %#v", aliased.View)
	}
}

func TestBrowserUnifiedApplyTopLevelAliasProjectionBuildsSummaryAndDisplayFromWorkbenchShells(t *testing.T) {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal([]byte(`{
		"status":"ok",
		"default_candidate_route":{"backend":"proxy","profile":"isolated","runtime_target":"node"},
		"workbench_summary":{"category":"coordination","state":"action_plan_available","summary_code":"workbench_action_plan","primary_browser_action":"browser action=refresh","primary_node_action":"nodes action=run","next_step":"browser action=refresh"},
		"workbench_display":{"ready":true,"sections":["route","coordination"],"category":"coordination","state":"action_plan_available","summary_code":"workbench_action_plan","primary_browser_action":"browser action=refresh","primary_node_action":"nodes action=run","next_step":"browser action=refresh"}
	}`), &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	inputs := browserRuntimeTopLevelAliasInputsFromPayload(payload)

	mutated, err := browserUnifiedApplyTopLevelAliasProjection(payload, &inputs)
	if err != nil {
		t.Fatalf("browserUnifiedApplyTopLevelAliasProjection: %v", err)
	}
	if !mutated {
		t.Fatalf("expected top-level alias projection to synthesize summary/display from workbench shells")
	}

	var aliased struct {
		Summary *browserTopLevelSummary        `json:"summary"`
		Display *browserTopLevelDisplaySummary `json:"display"`
	}
	blob, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if err := json.Unmarshal(blob, &aliased); err != nil {
		t.Fatalf("decode aliased payload: %v", err)
	}
	if aliased.Summary == nil ||
		aliased.Summary.Category != "coordination" ||
		aliased.Summary.SummaryCode != "workbench_action_plan" ||
		aliased.Summary.DefaultCandidateRoute != (browserRuntimeRouteDescriptor{Backend: "proxy", Profile: "isolated", RuntimeTarget: "node"}) {
		t.Fatalf("unexpected synthesized summary alias from workbench shells: %#v", aliased.Summary)
	}
	if aliased.Display == nil ||
		!aliased.Display.Ready ||
		!browserStringSliceContains(aliased.Display.Sections, "route") ||
		aliased.Display.Category != "coordination" ||
		aliased.Display.SummaryCode != "workbench_action_plan" ||
		aliased.Display.DefaultCandidateRoute != (browserRuntimeRouteDescriptor{Backend: "proxy", Profile: "isolated", RuntimeTarget: "node"}) {
		t.Fatalf("unexpected synthesized display alias from workbench shells: %#v", aliased.Display)
	}
}

func TestBrowserUnifiedApplyExplanationAliasPrefersWorkbenchExplanation(t *testing.T) {
	out, err := browserUnifiedApplyExplanationAlias(`{"status":"ok","workbench_explanation":{"category":"resolver","state":"manual_resolution_required","summary_code":"label_filtered_residual","next_step_alias":"snapshot","manual_retry_hint":"add_ordinal"},"diagnostics_explanation":{"category":"resolver","state":"resolver_attention_required","summary_code":"label_filtered_residual","next_step_alias":"snapshot","manual_retry_hint":"add_ordinal"},"resolver_explanation":{"state":"manual_resolution_required","summary_code":"label_filtered_residual","next_step_alias":"snapshot","manual_retry_hint":"add_ordinal"}}`)
	if err != nil {
		t.Fatalf("browserUnifiedApplyExplanationAlias: %v", err)
	}
	var payload struct {
		Explanation          *browserTopLevelSummary               `json:"explanation"`
		WorkbenchExplanation *browserDiagnosticsExplanationSummary `json:"workbench_explanation"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode aliased workbench output: %v", err)
	}
	if payload.Explanation == nil ||
		payload.Explanation.Category != "resolver" ||
		payload.Explanation.State != "manual_resolution_required" ||
		payload.Explanation.SummaryCode != "label_filtered_residual" ||
		payload.Explanation.NextStepAlias != "snapshot" ||
		payload.Explanation.ManualRetryHint != "add_ordinal" ||
		payload.Explanation.ResolvedViaFallback {
		t.Fatalf("unexpected workbench explanation alias: %#v", payload.Explanation)
	}
	if payload.WorkbenchExplanation == nil ||
		payload.WorkbenchExplanation.State != "manual_resolution_required" {
		t.Fatalf("expected workbench_explanation to remain intact, got %#v", payload.WorkbenchExplanation)
	}
}

func TestBrowserUnifiedApplyExplanationAliasFallsBackToResolverExplanation(t *testing.T) {
	out, err := browserUnifiedApplyExplanationAlias(`{"status":"clicked","resolver_explanation":{"state":"resolved_via_fallback","summary_code":"label_filtered_residual","manual_retry_hint":"add_ordinal"}}`)
	if err != nil {
		t.Fatalf("browserUnifiedApplyExplanationAlias: %v", err)
	}
	var payload struct {
		Explanation *browserTopLevelSummary        `json:"explanation"`
		Summary     *browserTopLevelSummary        `json:"summary"`
		Display     *browserTopLevelDisplaySummary `json:"display"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode aliased resolver output: %v", err)
	}
	if payload.Explanation == nil ||
		payload.Explanation.Category != "resolver_fallback" ||
		payload.Explanation.State != "resolved_via_fallback" ||
		payload.Explanation.SummaryCode != "label_filtered_residual" ||
		payload.Explanation.NextStepAlias != "" ||
		payload.Explanation.ManualRetryHint != "add_ordinal" ||
		!payload.Explanation.ResolvedViaFallback {
		t.Fatalf("unexpected resolver explanation alias: %#v", payload.Explanation)
	}
	if payload.Summary == nil ||
		payload.Summary.Category != "resolver_fallback" ||
		payload.Summary.State != "resolved_via_fallback" ||
		payload.Summary.SummaryCode != "label_filtered_residual" ||
		payload.Summary.ManualRetryHint != "add_ordinal" ||
		!payload.Summary.ResolvedViaFallback {
		t.Fatalf("unexpected unified summary alias: %#v", payload.Summary)
	}
	if payload.Display == nil ||
		payload.Display.Ready ||
		len(payload.Display.Sections) != 0 ||
		payload.Display.Category != "resolver_fallback" ||
		payload.Display.State != "resolved_via_fallback" ||
		payload.Display.SummaryCode != "label_filtered_residual" ||
		payload.Display.ManualRetryHint != "add_ordinal" ||
		!payload.Display.ResolvedViaFallback {
		t.Fatalf("unexpected unified display alias: %#v", payload.Display)
	}
}

func TestBrowserUnifiedApplyExplanationAliasPrefersWorkbenchDiagnostics(t *testing.T) {
	out, err := browserUnifiedApplyExplanationAlias(`{"status":"ok","workbench_diagnostics":{"category":"coordination","state":"action_plan_available","summary_code":"workbench_action_plan","primary_browser_action":"browser","primary_node_action":"nodes action=run_status","next_step":"nodes action=run_status"},"diagnostics_explanation":{"category":"resolver","state":"manual_resolution_required","summary_code":"label_filtered_residual","next_step_alias":"snapshot","manual_retry_hint":"add_ordinal"}}`)
	if err != nil {
		t.Fatalf("browserUnifiedApplyExplanationAlias: %v", err)
	}
	var payload struct {
		Diagnostics            *browserTopLevelSummary                    `json:"diagnostics"`
		WorkbenchDiagnostics   *browserRuntimeWorkbenchDiagnosticsSummary `json:"workbench_diagnostics"`
		DiagnosticsExplanation *browserDiagnosticsExplanationSummary      `json:"diagnostics_explanation"`
		Summary                *browserTopLevelSummary                    `json:"summary"`
		Display                *browserTopLevelDisplaySummary             `json:"display"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode aliased workbench diagnostics output: %v", err)
	}
	if payload.Diagnostics == nil ||
		payload.Diagnostics.Category != "coordination" ||
		payload.Diagnostics.State != "action_plan_available" ||
		payload.Diagnostics.SummaryCode != "workbench_action_plan" ||
		payload.Diagnostics.PrimaryBrowserAction != "browser" ||
		payload.Diagnostics.PrimaryNodeAction != "nodes action=run_status" ||
		payload.Diagnostics.NextStep != "nodes action=run_status" ||
		payload.Diagnostics.ResolvedViaFallback {
		t.Fatalf("unexpected diagnostics alias: %#v", payload.Diagnostics)
	}
	if payload.WorkbenchDiagnostics == nil || payload.WorkbenchDiagnostics.SummaryCode != "workbench_action_plan" {
		t.Fatalf("expected workbench_diagnostics to remain intact, got %#v", payload.WorkbenchDiagnostics)
	}
	if payload.DiagnosticsExplanation == nil || payload.DiagnosticsExplanation.SummaryCode != "label_filtered_residual" {
		t.Fatalf("expected diagnostics_explanation to remain intact, got %#v", payload.DiagnosticsExplanation)
	}
	if payload.Summary == nil ||
		payload.Summary.Category != "coordination" ||
		payload.Summary.State != "action_plan_available" ||
		payload.Summary.SummaryCode != "workbench_action_plan" ||
		payload.Summary.PrimaryBrowserAction != "browser" ||
		payload.Summary.PrimaryNodeAction != "nodes action=run_status" ||
		payload.Summary.NextStep != "nodes action=run_status" ||
		payload.Summary.ResolvedViaFallback {
		t.Fatalf("unexpected unified summary alias from workbench diagnostics: %#v", payload.Summary)
	}
	if payload.Display == nil ||
		payload.Display.Ready ||
		len(payload.Display.Sections) != 0 ||
		payload.Display.Category != "coordination" ||
		payload.Display.State != "action_plan_available" ||
		payload.Display.SummaryCode != "workbench_action_plan" ||
		payload.Display.PrimaryBrowserAction != "browser" ||
		payload.Display.PrimaryNodeAction != "nodes action=run_status" ||
		payload.Display.NextStep != "nodes action=run_status" ||
		payload.Display.ResolvedViaFallback {
		t.Fatalf("unexpected unified display alias from workbench diagnostics: %#v", payload.Display)
	}
}

func TestBrowserUnifiedApplyExplanationAliasPrefersWorkbenchSummary(t *testing.T) {
	out, err := browserUnifiedApplyExplanationAlias(`{"status":"ok","workbench_summary":{"category":"coordination","state":"action_plan_available","summary_code":"workbench_action_plan","primary_browser_action":"browser","primary_node_action":"nodes action=run_status","next_step":"nodes action=run_status"},"diagnostics_explanation":{"category":"resolver","state":"manual_resolution_required","summary_code":"label_filtered_residual","next_step_alias":"snapshot","manual_retry_hint":"add_ordinal"}}`)
	if err != nil {
		t.Fatalf("browserUnifiedApplyExplanationAlias: %v", err)
	}
	var payload struct {
		DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation"`
		WorkbenchSummary       *browserTopLevelSummary               `json:"workbench_summary"`
		Summary                *browserTopLevelSummary               `json:"summary"`
		Display                *browserTopLevelDisplaySummary        `json:"display"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode aliased workbench summary output: %v", err)
	}
	if payload.WorkbenchSummary == nil ||
		payload.WorkbenchSummary.Category != "coordination" ||
		payload.WorkbenchSummary.State != "action_plan_available" ||
		payload.WorkbenchSummary.SummaryCode != "workbench_action_plan" ||
		payload.WorkbenchSummary.PrimaryBrowserAction != "browser" ||
		payload.WorkbenchSummary.PrimaryNodeAction != "nodes action=run_status" ||
		payload.WorkbenchSummary.NextStep != "nodes action=run_status" {
		t.Fatalf("expected workbench_summary to remain intact, got %#v", payload.WorkbenchSummary)
	}
	if payload.DiagnosticsExplanation == nil || payload.DiagnosticsExplanation.SummaryCode != "label_filtered_residual" {
		t.Fatalf("expected diagnostics_explanation to remain intact, got %#v", payload.DiagnosticsExplanation)
	}
	if payload.Summary == nil ||
		payload.Summary.Category != "coordination" ||
		payload.Summary.State != "action_plan_available" ||
		payload.Summary.SummaryCode != "workbench_action_plan" ||
		payload.Summary.PrimaryBrowserAction != "browser" ||
		payload.Summary.PrimaryNodeAction != "nodes action=run_status" ||
		payload.Summary.NextStep != "nodes action=run_status" ||
		payload.Summary.ResolvedViaFallback {
		t.Fatalf("unexpected unified summary alias from workbench summary: %#v", payload.Summary)
	}
	if payload.Display == nil ||
		payload.Display.Ready ||
		len(payload.Display.Sections) != 0 ||
		payload.Display.Category != "coordination" ||
		payload.Display.State != "action_plan_available" ||
		payload.Display.SummaryCode != "workbench_action_plan" ||
		payload.Display.PrimaryBrowserAction != "browser" ||
		payload.Display.PrimaryNodeAction != "nodes action=run_status" ||
		payload.Display.NextStep != "nodes action=run_status" ||
		payload.Display.ResolvedViaFallback {
		t.Fatalf("unexpected unified display alias from workbench summary: %#v", payload.Display)
	}
}

func TestBrowserUnifiedApplyExplanationAliasBuildsDisplayFromRootSummary(t *testing.T) {
	out, err := browserUnifiedApplyExplanationAlias(`{"status":"ok","summary":{"category":"coordination","state":"action_plan_available","summary_code":"workbench_action_plan","primary_browser_action":"browser action=refresh","primary_node_action":"nodes action=run","next_step":"browser action=refresh"}}`)
	if err != nil {
		t.Fatalf("browserUnifiedApplyExplanationAlias: %v", err)
	}
	var payload struct {
		Summary *browserTopLevelSummary        `json:"summary"`
		Display *browserTopLevelDisplaySummary `json:"display"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode aliased root summary output: %v", err)
	}
	if payload.Summary == nil || payload.Summary.SummaryCode != "workbench_action_plan" {
		t.Fatalf("expected root summary to remain intact, got %#v", payload.Summary)
	}
	if payload.Display == nil ||
		payload.Display.Category != "coordination" ||
		payload.Display.State != "action_plan_available" ||
		payload.Display.SummaryCode != "workbench_action_plan" ||
		payload.Display.PrimaryBrowserAction != "browser action=refresh" ||
		payload.Display.PrimaryNodeAction != "nodes action=run" ||
		payload.Display.NextStep != "browser action=refresh" {
		t.Fatalf("expected display alias to recover from root summary, got %#v", payload.Display)
	}
}

func TestBrowserUnifiedApplyExplanationAliasBuildsSurfaceFromDisplayAndReview(t *testing.T) {
	out, err := browserUnifiedApplyExplanationAlias(`{"status":"review_required","display":{"category":"review","state":"manual_confirmation_required","summary_code":"popup_review_required","next_step_alias":"tabs","manual_retry_hint":"rerun_with_force","primary_browser_action":"browser action=tabs","next_step":"browser action=tabs"},"review":{"policy_state":"popup_review_required","decision":"session_target_popup_review_required","summary":{"category":"review","state":"manual_confirmation_required","summary_code":"popup_review_required","next_step_alias":"tabs","manual_retry_hint":"rerun_with_force"},"display":{"category":"review","state":"manual_confirmation_required","summary_code":"popup_review_required","next_step_alias":"tabs","manual_retry_hint":"rerun_with_force","primary_browser_action":"browser action=tabs","next_step":"browser action=tabs"}}}`)
	if err != nil {
		t.Fatalf("browserUnifiedApplyExplanationAlias: %v", err)
	}
	var payload struct {
		Display *browserTopLevelDisplaySummary `json:"display"`
		Review  *browserReviewSurfaceSummary   `json:"review"`
		Surface *browserTopLevelSurfaceSummary `json:"surface"`
		View    *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode aliased surface output: %v", err)
	}
	if payload.Display == nil || payload.Display.SummaryCode != "popup_review_required" {
		t.Fatalf("expected display to remain intact, got %#v", payload.Display)
	}
	if payload.Review == nil || payload.Review.PolicyState != "popup_review_required" {
		t.Fatalf("expected review to remain intact, got %#v", payload.Review)
	}
	if payload.Surface == nil ||
		payload.Surface.Category != "review" ||
		payload.Surface.State != "manual_confirmation_required" ||
		payload.Surface.SummaryCode != "popup_review_required" ||
		payload.Surface.NextStepAlias != "tabs" ||
		payload.Surface.ManualRetryHint != "rerun_with_force" ||
		payload.Surface.PrimaryBrowserAction != "browser action=tabs" ||
		payload.Surface.NextStep != "browser action=tabs" ||
		payload.Surface.ReviewPolicyState != "popup_review_required" ||
		payload.Surface.ReviewDecision != "session_target_popup_review_required" ||
		payload.Surface.ReviewReady {
		t.Fatalf("unexpected unified surface alias: %#v", payload.Surface)
	}
	if payload.View == nil ||
		payload.View.Kind != "review" ||
		payload.View.Category != "review" ||
		payload.View.State != "manual_confirmation_required" ||
		payload.View.SummaryCode != "popup_review_required" ||
		payload.View.NextStepAlias != "tabs" ||
		payload.View.ManualRetryHint != "rerun_with_force" ||
		payload.View.PrimaryBrowserAction != "browser action=tabs" ||
		payload.View.NextStep != "browser action=tabs" ||
		payload.View.Review == nil ||
		payload.View.Review.PolicyState != "popup_review_required" ||
		payload.View.Review.Decision != "session_target_popup_review_required" ||
		payload.View.Review.Ready {
		t.Fatalf("unexpected unified view alias: %#v", payload.View)
	}
}

func TestBrowserUnifiedApplyExplanationAliasBuildsViewFromWorkbenchDisplay(t *testing.T) {
	out, err := browserUnifiedApplyExplanationAlias(`{"status":"ok","workbench_display":{"ready":true,"sections":["route","coordination"],"category":"coordination","state":"action_plan_available","summary_code":"workbench_action_plan","primary_browser_action":"browser action=refresh","primary_node_action":"nodes action=run","next_step":"browser action=refresh"}}`)
	if err != nil {
		t.Fatalf("browserUnifiedApplyExplanationAlias: %v", err)
	}
	var payload struct {
		Summary          *browserTopLevelSummary                `json:"summary"`
		WorkbenchDisplay *browserRuntimeWorkbenchDisplaySummary `json:"workbench_display"`
		View             *browserTopLevelViewSummary            `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode aliased workbench view output: %v", err)
	}
	if payload.WorkbenchDisplay == nil || payload.WorkbenchDisplay.SummaryCode != "workbench_action_plan" {
		t.Fatalf("expected workbench_display to remain intact, got %#v", payload.WorkbenchDisplay)
	}
	if payload.Summary == nil ||
		payload.Summary.Category != "coordination" ||
		payload.Summary.State != "action_plan_available" ||
		payload.Summary.SummaryCode != "workbench_action_plan" ||
		payload.Summary.PrimaryBrowserAction != "browser action=refresh" ||
		payload.Summary.PrimaryNodeAction != "nodes action=run" ||
		payload.Summary.NextStep != "browser action=refresh" {
		t.Fatalf("expected summary alias to recover from workbench_display, got %#v", payload.Summary)
	}
	if payload.View == nil ||
		payload.View.Kind != "workbench" ||
		!payload.View.Ready ||
		!browserStringSliceContains(payload.View.Sections, "route") ||
		!browserStringSliceContains(payload.View.Sections, "coordination") ||
		payload.View.Category != "coordination" ||
		payload.View.State != "action_plan_available" ||
		payload.View.SummaryCode != "workbench_action_plan" ||
		payload.View.PrimaryBrowserAction != "browser action=refresh" ||
		payload.View.PrimaryNodeAction != "nodes action=run" ||
		payload.View.NextStep != "browser action=refresh" ||
		payload.View.Review != nil {
		t.Fatalf("unexpected unified workbench view alias: %#v", payload.View)
	}
}

func TestBrowserUnifiedApplyExplanationAliasPrefersWorkbenchDisplayOverViewFallback(t *testing.T) {
	out, err := browserUnifiedApplyExplanationAlias(`{"status":"review_required","workbench_display":{"ready":true,"sections":["route","coordination"],"category":"coordination","state":"action_plan_available","summary_code":"workbench_action_plan","primary_browser_action":"browser action=refresh","next_step":"browser action=refresh"},"view":{"kind":"review","category":"review","state":"manual_confirmation_required","summary_code":"popup_review_required","next_step_alias":"tabs","manual_retry_hint":"rerun_with_force","primary_browser_action":"browser action=tabs","next_step":"browser action=tabs","review":{"policy_state":"popup_review_required","decision":"session_target_popup_review_required","summary":{"category":"review","state":"manual_confirmation_required","summary_code":"popup_review_required","next_step_alias":"tabs","manual_retry_hint":"rerun_with_force"}}}}`)
	if err != nil {
		t.Fatalf("browserUnifiedApplyExplanationAlias: %v", err)
	}
	var payload struct {
		Display *browserTopLevelDisplaySummary `json:"display"`
		View    *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode aliased workbench display precedence output: %v", err)
	}
	if payload.Display == nil ||
		payload.Display.Category != "coordination" ||
		payload.Display.SummaryCode != "workbench_action_plan" ||
		payload.Display.PrimaryBrowserAction != "browser action=refresh" {
		t.Fatalf("expected display alias to preserve workbench_display precedence, got %#v", payload.Display)
	}
	if payload.View == nil || payload.View.Review == nil || payload.View.Review.PolicyState != "popup_review_required" {
		t.Fatalf("expected view review payload to remain intact, got %#v", payload.View)
	}
}

func TestBrowserUnifiedApplyExplanationAliasPrefersExplicitReviewOverViewAndWorkbench(t *testing.T) {
	out, err := browserUnifiedApplyExplanationAlias(`{"status":"review_required","review":{"policy_state":"redirect_review_required","decision":"session_target_redirect_review_required","summary":{"category":"review","state":"manual_confirmation_required","summary_code":"redirect_review_required"}},"view":{"kind":"review","review":{"policy_state":"popup_review_required","decision":"session_target_popup_review_required","summary":{"category":"review","state":"manual_confirmation_required","summary_code":"popup_review_required"}}},"workbench":{"review":{"policy_state":"download_review_required","decision":"download_review_required","summary":{"category":"review","state":"manual_confirmation_required","summary_code":"download_review_required"}}}}`)
	if err != nil {
		t.Fatalf("browserUnifiedApplyExplanationAlias: %v", err)
	}
	var payload struct {
		Review *browserReviewSurfaceSummary `json:"review"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode aliased explicit review precedence output: %v", err)
	}
	if payload.Review == nil ||
		payload.Review.PolicyState != "redirect_review_required" ||
		payload.Review.Decision != "session_target_redirect_review_required" ||
		payload.Review.Summary == nil ||
		payload.Review.Summary.SummaryCode != "redirect_review_required" {
		t.Fatalf("expected explicit review to remain primary, got %#v", payload.Review)
	}
}

func TestBrowserUnifiedApplyExplanationAliasPrefersWorkbenchReviewInView(t *testing.T) {
	out, err := browserUnifiedApplyExplanationAlias(`{"status":"ok","review":{"policy_state":"redirect_review_required","decision":"session_target_redirect_review_required","summary":{"category":"review","state":"manual_confirmation_required","summary_code":"redirect_review_required"}},"workbench":{"ready":true,"sections":["coordination"],"summary":{"category":"coordination","state":"action_plan_available","summary_code":"workbench_action_plan","primary_browser_action":"browser action=refresh","next_step":"browser action=refresh"},"review":{"policy_state":"popup_review_required","decision":"session_target_popup_review_required","summary":{"category":"review","state":"manual_confirmation_required","summary_code":"popup_review_required","next_step_alias":"tabs","manual_retry_hint":"rerun_with_force"},"display":{"category":"review","state":"manual_confirmation_required","summary_code":"popup_review_required","next_step_alias":"tabs","manual_retry_hint":"rerun_with_force","primary_browser_action":"browser action=tabs","next_step":"browser action=tabs"}}}}`)
	if err != nil {
		t.Fatalf("browserUnifiedApplyExplanationAlias: %v", err)
	}
	var payload struct {
		View *browserTopLevelViewSummary `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode aliased workbench review output: %v", err)
	}
	if payload.View == nil ||
		payload.View.Kind != "workbench" ||
		!payload.View.Ready ||
		payload.View.Category != "coordination" ||
		payload.View.State != "action_plan_available" ||
		payload.View.SummaryCode != "workbench_action_plan" ||
		payload.View.PrimaryBrowserAction != "browser action=refresh" ||
		payload.View.NextStep != "browser action=refresh" ||
		payload.View.Review == nil ||
		payload.View.Review.PolicyState != "popup_review_required" ||
		payload.View.Review.Decision != "session_target_popup_review_required" ||
		payload.View.Review.Ready ||
		payload.View.Review.Display == nil ||
		payload.View.Review.Display.SummaryCode != "popup_review_required" {
		t.Fatalf("unexpected unified workbench review precedence: %#v", payload.View)
	}
}

func TestBrowserUnifiedApplyExplanationAliasBuildsDisplayAndSurfaceFromViewReview(t *testing.T) {
	out, err := browserUnifiedApplyExplanationAlias(`{"status":"review_required","view":{"kind":"review","category":"review","state":"manual_confirmation_required","summary_code":"popup_review_required","next_step_alias":"tabs","manual_retry_hint":"rerun_with_force","primary_browser_action":"browser action=tabs","next_step":"browser action=tabs","review":{"policy_state":"popup_review_required","decision":"session_target_popup_review_required","summary":{"category":"review","state":"manual_confirmation_required","summary_code":"popup_review_required","next_step_alias":"tabs","manual_retry_hint":"rerun_with_force"},"display":{"category":"review","state":"manual_confirmation_required","summary_code":"popup_review_required","next_step_alias":"tabs","manual_retry_hint":"rerun_with_force","primary_browser_action":"browser action=tabs","next_step":"browser action=tabs"}}}}`)
	if err != nil {
		t.Fatalf("browserUnifiedApplyExplanationAlias: %v", err)
	}
	var payload struct {
		Summary *browserTopLevelSummary        `json:"summary"`
		Display *browserTopLevelDisplaySummary `json:"display"`
		Surface *browserTopLevelSurfaceSummary `json:"surface"`
		View    *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode aliased view review output: %v", err)
	}
	if payload.Summary == nil ||
		payload.Summary.Category != "review" ||
		payload.Summary.State != "manual_confirmation_required" ||
		payload.Summary.SummaryCode != "popup_review_required" ||
		payload.Summary.NextStepAlias != "tabs" {
		t.Fatalf("expected summary alias to recover from view review, got %#v", payload.Summary)
	}
	if payload.Display == nil ||
		payload.Display.Category != "review" ||
		payload.Display.State != "manual_confirmation_required" ||
		payload.Display.SummaryCode != "popup_review_required" ||
		payload.Display.NextStepAlias != "tabs" {
		t.Fatalf("expected display alias to recover from view, got %#v", payload.Display)
	}
	if payload.Surface == nil ||
		payload.Surface.Category != "review" ||
		payload.Surface.State != "manual_confirmation_required" ||
		payload.Surface.SummaryCode != "popup_review_required" ||
		payload.Surface.ReviewPolicyState != "popup_review_required" ||
		payload.Surface.ReviewDecision != "session_target_popup_review_required" {
		t.Fatalf("expected surface alias to recover review/display from view, got %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Review == nil || payload.View.Review.Display == nil {
		t.Fatalf("expected view review payload to remain intact, got %#v", payload.View)
	}
}

func TestBrowserUnifiedApplyExplanationAliasBuildsDisplayAndSurfaceFromWorkbenchReview(t *testing.T) {
	out, err := browserUnifiedApplyExplanationAlias(`{"status":"review_required","workbench":{"ready":true,"sections":["coordination"],"summary":{"category":"coordination","state":"action_plan_available","summary_code":"workbench_action_plan","primary_browser_action":"browser action=refresh","next_step":"browser action=refresh"},"review":{"policy_state":"popup_review_required","decision":"session_target_popup_review_required","summary":{"category":"review","state":"manual_confirmation_required","summary_code":"popup_review_required","next_step_alias":"tabs","manual_retry_hint":"rerun_with_force"},"display":{"category":"review","state":"manual_confirmation_required","summary_code":"popup_review_required","next_step_alias":"tabs","manual_retry_hint":"rerun_with_force","primary_browser_action":"browser action=tabs","next_step":"browser action=tabs"}}}}`)
	if err != nil {
		t.Fatalf("browserUnifiedApplyExplanationAlias: %v", err)
	}
	var payload struct {
		Summary *browserTopLevelSummary        `json:"summary"`
		Display *browserTopLevelDisplaySummary `json:"display"`
		Surface *browserTopLevelSurfaceSummary `json:"surface"`
		View    *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode aliased workbench review fallback output: %v", err)
	}
	if payload.Summary == nil ||
		payload.Summary.Category != "review" ||
		payload.Summary.State != "manual_confirmation_required" ||
		payload.Summary.SummaryCode != "popup_review_required" ||
		payload.Summary.NextStepAlias != "tabs" {
		t.Fatalf("expected summary alias to prefer review fallback over workbench summary, got %#v", payload.Summary)
	}
	if payload.Display == nil ||
		payload.Display.Category != "review" ||
		payload.Display.State != "manual_confirmation_required" ||
		payload.Display.SummaryCode != "popup_review_required" {
		t.Fatalf("expected display alias to recover from workbench review, got %#v", payload.Display)
	}
	if payload.Surface == nil ||
		payload.Surface.ReviewPolicyState != "popup_review_required" ||
		payload.Surface.ReviewDecision != "session_target_popup_review_required" ||
		payload.Surface.SummaryCode != "popup_review_required" {
		t.Fatalf("expected surface alias to recover review/display from workbench review, got %#v", payload.Surface)
	}
	if payload.View == nil ||
		payload.View.Review == nil ||
		payload.View.Review.PolicyState != "popup_review_required" ||
		payload.View.Review.Display == nil ||
		payload.View.Review.Display.SummaryCode != "popup_review_required" {
		t.Fatalf("expected workbench review to remain visible on view alias, got %#v", payload.View)
	}
}

func TestBrowserDefinitionIncludesUnifiedActionEnum(t *testing.T) {
	def := browserDefinition([]string{"status", "workbench", "prepare", "profiles", "sessions", "coordinate", "start", "stop", "sync_session", "create_profile", "delete_profile", "select_profile", "clear_profile", "select_target", "clear_target"}, []string{"open", "list_tabs", "save_pdf"})
	if def.Function.Name != "browser" {
		t.Fatalf("expected browser tool definition, got %#v", def.Function.Name)
	}
	props, ok := def.Function.Parameters["properties"].(map[string]any)
	if !ok {
		t.Fatalf("expected browser definition properties, got %#v", def.Function.Parameters)
	}
	actionDef, ok := props["action"].(map[string]any)
	if !ok {
		t.Fatalf("expected action schema, got %#v", props["action"])
	}
	if !strings.Contains(readStringFromMap(actionDef, "description"), "doctor") ||
		!strings.Contains(readStringFromMap(actionDef, "description"), "reset") {
		t.Fatalf("expected browser action description to clarify diagnostics and reset semantics, got %#v", actionDef)
	}
	enum, ok := actionDef["enum"].([]string)
	if !ok {
		t.Fatalf("expected string enum for action schema, got %#v", actionDef["enum"])
	}
	expected := []string{"inspect", "doctor", "status", "workbench", "ready", "prepare", "inventory", "profiles", "handles", "sessions", "ensure", "refresh", "sync", "teardown", "reset", "coordinate", "launch", "start", "halt", "stop", "adopt", "sync_session", "new_profile", "create_profile", "remove_profile", "delete_profile", "pin_profile", "select_profile", "unpin_profile", "clear_profile", "pin_target", "select_target", "unpin_target", "clear_target", "open", "list_tabs", "save_pdf", "tabs", "pdf", "act"}
	if len(enum) != len(expected) {
		t.Fatalf("expected %d browser actions, got %#v", len(expected), enum)
	}
	for i := range expected {
		if enum[i] != expected[i] {
			t.Fatalf("expected browser action enum %#v, got %#v", expected, enum)
		}
	}
	runtimeTargetDef, ok := props["runtime_target"].(map[string]any)
	if !ok {
		t.Fatalf("expected runtime_target schema, got %#v", props["runtime_target"])
	}
	if !strings.Contains(readStringFromMap(runtimeTargetDef, "description"), "runtime_target=host") {
		t.Fatalf("expected runtime_target description to clarify explicit host usage, got %#v", runtimeTargetDef)
	}
}

func TestBrowserDefinitionHasClosedSchemaContract(t *testing.T) {
	def := browserDefinition(browserUnifiedInventoryRuntimeActions(), browserUnifiedInventoryActKinds())
	if def.Function.Parameters["additionalProperties"] != false {
		t.Fatalf("expected closed input schema, got %#v", def.Function.Parameters)
	}
	if _, ok := def.Function.Parameters["required"]; ok {
		t.Fatalf("expected unified browser action alternatives to stay optional in schema, got required=%#v", def.Function.Parameters["required"])
	}
	assertSchemaProperties(t, def.Function.Parameters, []string{
		"action",
		"operation",
		"kind",
		"url",
		"target",
		"selector",
		"ref",
		"text",
		"max_chars",
		"profile",
		"runtime_target",
		"remember_target",
		"remember",
	})

	if def.Function.OutputSchema["additionalProperties"] != false {
		t.Fatalf("expected closed output schema, got %#v", def.Function.OutputSchema)
	}
	assertRequiredFields(t, def.Function.OutputSchema, []string{"status"})
	assertSchemaProperties(t, def.Function.OutputSchema, []string{
		"status",
		"kind",
		"action",
		"backend",
		"browser_app",
		"profile",
		"runtime_target",
		"url",
		"final_url",
		"title",
		"content",
		"content_type",
		"snapshot",
		"elements",
		"path",
		"artifacts",
		"tabs",
		"target",
		"target_id",
		"summary",
		"diagnostics",
		"explanation",
		"display",
		"surface",
		"view",
		"workbench",
		"runtime_actions",
		"browser_tools",
		"browser_act_kinds",
		"capabilities",
		"note",
	})
}

func assertBrowserUnifiedOutputKeysCoveredBySchema(t *testing.T, raw string) {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		t.Fatalf("decode unified browser payload: %v", err)
	}
	props, ok := browserUnifiedOutputSchema()["properties"].(map[string]any)
	if !ok {
		t.Fatalf("expected browser output properties map")
	}
	for key := range payload {
		if _, ok := props[key]; !ok {
			t.Fatalf("runtime output key %q missing from output schema; payload=%#v schema=%#v", key, payload, props)
		}
	}
}

func TestBrowserUnifiedCanDelegateActActionRestrictsToVisibleAndExplicitManagedKinds(t *testing.T) {
	ctx := browserRegistrationContext{
		capabilities: BrowserCapabilitiesForActKinds([]string{"open"}),
		enabledTools: buildEnabledToolSet([]string{"browser"}),
		opts: BrowserToolOptions{
			NodeBackend: &runtimeInfoCapabilityBrowserBackend{
				fakeBrowserBackend: &fakeBrowserBackend{},
				runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
				capabilities:       BrowserCapabilitiesForActKinds([]string{"click"}),
			},
		},
	}

	if !browserUnifiedCanDelegateActAction(ctx, nil, map[string]any{}, "open") {
		t.Fatalf("expected unified browser to delegate visible act kind")
	}
	if !browserUnifiedCanDelegateActAction(ctx, nil, map[string]any{}, "click") {
		t.Fatalf("expected unified browser to delegate explicit managed act kind")
	}
	if !browserUnifiedCanDelegateActAction(ctx, nil, map[string]any{"kind": "click"}, "act") {
		t.Fatalf("expected unified browser to delegate explicit managed act kind via action=act")
	}
	if browserUnifiedCanDelegateActAction(ctx, nil, map[string]any{}, "hover") {
		t.Fatalf("expected unified browser to reject unsupported hover act kind")
	}
	if browserUnifiedCanDelegateActAction(ctx, nil, map[string]any{"kind": "hover"}, "act") {
		t.Fatalf("expected unified browser to reject unsupported act kind via action=act")
	}
	if browserUnifiedCanDelegateActAction(ctx, []string{"status"}, map[string]any{}, "status") {
		t.Fatalf("expected runtime actions to stay out of act delegation fallback")
	}
}
