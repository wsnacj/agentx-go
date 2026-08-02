package tools

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
	types "github.com/wsnacj/agentx-go/components/llm"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

func browserRuntimeDefinitionActions(reg *llmxtools.Registry) []string {
	for _, def := range reg.Definitions() {
		if def.Function.Name != "browser_runtime" {
			continue
		}
		properties, _ := def.Function.Parameters["properties"].(map[string]any)
		actionDef, _ := properties["action"].(map[string]any)
		rawActions, _ := actionDef["enum"].([]string)
		if len(rawActions) > 0 {
			return append([]string(nil), rawActions...)
		}
		rawAny, _ := actionDef["enum"].([]any)
		actions := make([]string, 0, len(rawAny))
		for _, item := range rawAny {
			if text, ok := item.(string); ok {
				actions = append(actions, text)
			}
		}
		return actions
	}
	return nil
}

func writeBrowserDoctorRepoScriptForTest(t *testing.T, root string, name string, body string) string {
	t.Helper()
	path := filepath.Join(root, "core", "agentx", "browserdaemon", name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir script dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}
	return path
}

func TestRegisterBrowserTools_RuntimeRepairActionRunsBootstrapRepairScript(t *testing.T) {
	root := t.TempDir()
	marker := filepath.Join(root, "repair.marker")
	repairScript := writeBrowserDoctorRepoScriptForTest(t, root, "browserd-bootstrap-repair.sh", "#!/usr/bin/env bash\nset -euo pipefail\necho repair-completed\nprintf 'ok' > \""+marker+"\"\n")

	reg := llmxtools.NewRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         root,
		EnabledTools: []string{"browser_runtime"},
		RepairScript: repairScript,
		RunCommand: func(ctx context.Context, name string, args []string) ([]byte, error) {
			return []byte("repair-completed"), os.WriteFile(marker, []byte("ok"), 0o644)
		},
	})

	if !browserStringSliceContains(browserRuntimeDefinitionActions(reg), "repair") {
		t.Fatalf("expected browser_runtime definition to expose repair action, got %#v", browserRuntimeDefinitionActions(reg))
	}

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"repair","include_routes":true}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime repair: %v", err)
	}

	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("expected repair script to create marker file: %v", err)
	}

	var payload struct {
		Action         string                         `json:"action"`
		Status         string                         `json:"status"`
		RepairDecision string                         `json:"repair_decision"`
		RepairReady    bool                           `json:"repair_ready"`
		RepairOutput   string                         `json:"repair_output"`
		Note           string                         `json:"note"`
		RuntimeActions []string                       `json:"runtime_actions"`
		Doctor         *BrowserDoctorSummary          `json:"doctor"`
		Summary        *browserTopLevelSummary        `json:"summary"`
		Display        *browserTopLevelDisplaySummary `json:"display"`
		Surface        *browserTopLevelSurfaceSummary `json:"surface"`
		View           *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode runtime repair payload: %v", err)
	}
	if payload.Action != "repair" || payload.Status != "ok" || payload.RepairDecision != "repaired" || !payload.RepairReady {
		t.Fatalf("unexpected runtime repair payload: %#v", payload)
	}
	if !strings.Contains(payload.RepairOutput, "repair-completed") {
		t.Fatalf("expected repair output to include script output, got %#v", payload)
	}
	if payload.Note != "Bundled browserd bootstrap repair completed." {
		t.Fatalf("expected runtime repair note to summarize completion, got %#v", payload)
	}
	if !browserStringSliceContains(payload.RuntimeActions, "repair") {
		t.Fatalf("expected runtime action metadata to expose repair, got %#v", payload.RuntimeActions)
	}
	if payload.Doctor == nil || strings.TrimSpace(payload.Doctor.RepairCommand) == "" {
		t.Fatalf("expected runtime repair payload to retain doctor summary, got %#v", payload)
	}
	if payload.Summary == nil ||
		payload.Summary.Category != "coordination" ||
		payload.Summary.State != "completed" ||
		payload.Summary.SummaryCode != "repair_completed" {
		t.Fatalf("expected runtime repair payload to expose repair summary aliases, got %#v", payload.Summary)
	}
	if payload.Display == nil ||
		payload.Display.Category != "coordination" ||
		payload.Display.State != "completed" ||
		payload.Display.SummaryCode != "repair_completed" {
		t.Fatalf("expected runtime repair payload to expose repair display aliases, got %#v", payload.Display)
	}
	if payload.Surface == nil ||
		payload.Surface.Category != "coordination" ||
		payload.Surface.State != "completed" ||
		payload.Surface.SummaryCode != "repair_completed" {
		t.Fatalf("expected runtime repair payload to expose repair surface aliases, got %#v", payload.Surface)
	}
	if payload.View == nil ||
		payload.View.Kind != "result" ||
		payload.View.Category != "coordination" ||
		payload.View.State != "completed" ||
		payload.View.SummaryCode != "repair_completed" {
		t.Fatalf("expected runtime repair payload to expose repair view aliases, got %#v", payload.View)
	}
}

func TestRegisterBrowserTools_RuntimeRepairActionReportsBootstrapRepairFailure(t *testing.T) {
	root := t.TempDir()
	repairScript := writeBrowserDoctorRepoScriptForTest(t, root, "browserd-bootstrap-repair.sh", "#!/usr/bin/env bash\nset -euo pipefail\necho repair-failed >&2\nexit 17\n")

	reg := llmxtools.NewRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         root,
		EnabledTools: []string{"browser_runtime"},
		RepairScript: repairScript,
		RunCommand: func(context.Context, string, []string) ([]byte, error) {
			return nil, errors.New("command_failed: exit_code=17 stderr_bytes=14")
		},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"repair","include_routes":true}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime repair failure payload: %v", err)
	}

	var payload struct {
		Action         string                         `json:"action"`
		Status         string                         `json:"status"`
		RepairDecision string                         `json:"repair_decision"`
		RepairReady    bool                           `json:"repair_ready"`
		RepairOutput   string                         `json:"repair_output"`
		Note           string                         `json:"note"`
		Summary        *browserTopLevelSummary        `json:"summary"`
		Display        *browserTopLevelDisplaySummary `json:"display"`
		Surface        *browserTopLevelSurfaceSummary `json:"surface"`
		View           *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode runtime repair failure payload: %v", err)
	}
	if payload.Action != "repair" || payload.Status != "error" || payload.RepairDecision != "repair_failed" || payload.RepairReady {
		t.Fatalf("unexpected runtime repair failure payload: %#v", payload)
	}
	if strings.Contains(payload.RepairOutput+payload.Note, "repair-failed") ||
		!strings.Contains(payload.Note, "command_failed") ||
		!strings.Contains(payload.Note, "stderr_bytes=14") {
		t.Fatalf("expected runtime repair failure to expose only bounded command summary, got %#v", payload)
	}
	if payload.Summary == nil ||
		payload.Summary.Category != "coordination" ||
		payload.Summary.State != "failed" ||
		payload.Summary.SummaryCode != "repair_failed" ||
		payload.Summary.NextStepAlias != "repair" ||
		payload.Summary.ManualRetryHint != "repair_bootstrap" ||
		payload.Summary.PrimaryBrowserAction != "browser_runtime action=repair" ||
		payload.Summary.NextStep != "browser_runtime action=repair" {
		t.Fatalf("expected runtime repair failure summary aliases, got %#v", payload.Summary)
	}
	if payload.Display == nil ||
		payload.Display.Category != "coordination" ||
		payload.Display.State != "failed" ||
		payload.Display.SummaryCode != "repair_failed" {
		t.Fatalf("expected runtime repair failure display aliases, got %#v", payload.Display)
	}
	if payload.Surface == nil ||
		payload.Surface.Category != "coordination" ||
		payload.Surface.State != "failed" ||
		payload.Surface.SummaryCode != "repair_failed" {
		t.Fatalf("expected runtime repair failure surface aliases, got %#v", payload.Surface)
	}
	if payload.View == nil ||
		payload.View.Kind != "result" ||
		payload.View.Category != "coordination" ||
		payload.View.State != "failed" ||
		payload.View.SummaryCode != "repair_failed" {
		t.Fatalf("expected runtime repair failure view aliases, got %#v", payload.View)
	}
}

func TestBrowserRuntimeApplyRepairActionOutcomeSurfacesUnsupportedSummary(t *testing.T) {
	payload := browserRuntimePayload{
		Action: "repair",
		Note:   "Bundled browserd bootstrap repair script is not available from the current workspace root.",
	}
	browserRuntimeApplyRepairActionOutcome(
		&payload,
		agentxbrowserruntime.BuildSharedSessionBrowserRepairActionOutcome(
			agentxbrowserruntime.SharedSessionBrowserRepairActionOutcomeRequest{
				Status:         "unsupported",
				Note:           payload.Note,
				RepairDecision: "repair_command_unavailable",
				DoctorAction:   "browser_runtime action=doctor",
			},
		),
	)
	browserRuntimeSyncTopLevelSurfaceSummary(&payload)
	if payload.Summary == nil ||
		payload.Summary.Category != "coordination" ||
		payload.Summary.State != "unsupported" ||
		payload.Summary.SummaryCode != "repair_command_unavailable" ||
		payload.Summary.PrimaryBrowserAction != "browser_runtime action=doctor" ||
		payload.Summary.NextStep != "browser_runtime action=doctor" {
		t.Fatalf("expected runtime repair unsupported summary aliases, got %#v", payload.Summary)
	}
	if payload.Display == nil ||
		payload.Display.Category != "coordination" ||
		payload.Display.State != "unsupported" ||
		payload.Display.SummaryCode != "repair_command_unavailable" {
		t.Fatalf("expected runtime repair unsupported display aliases, got %#v", payload.Display)
	}
	if payload.Surface == nil ||
		payload.Surface.Category != "coordination" ||
		payload.Surface.State != "unsupported" ||
		payload.Surface.SummaryCode != "repair_command_unavailable" {
		t.Fatalf("expected runtime repair unsupported surface aliases, got %#v", payload.Surface)
	}
	if payload.View == nil ||
		payload.View.Kind != "result" ||
		payload.View.Category != "coordination" ||
		payload.View.State != "unsupported" ||
		payload.View.SummaryCode != "repair_command_unavailable" {
		t.Fatalf("expected runtime repair unsupported view aliases, got %#v", payload.View)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserRepairAliasDelegatesRuntimeRepair(t *testing.T) {
	root := t.TempDir()
	repairScript := writeBrowserDoctorRepoScriptForTest(t, root, "browserd-bootstrap-repair.sh", "#!/usr/bin/env bash\nset -euo pipefail\necho unified-repair-completed\n")

	reg := llmxtools.NewRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         root,
		EnabledTools: []string{"browser"},
		RepairScript: repairScript,
		RunCommand: func(context.Context, string, []string) ([]byte, error) {
			return []byte("unified-repair-completed"), nil
		},
	})

	if !browserStringSliceContains(browserUnifiedDefinitionActions(reg), "repair") {
		t.Fatalf("expected unified browser definition to expose repair action, got %#v", browserUnifiedDefinitionActions(reg))
	}

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"repair"}`,
	})
	if err != nil {
		t.Fatalf("browser unified repair: %v", err)
	}

	var payload struct {
		Action         string                         `json:"action"`
		Status         string                         `json:"status"`
		RepairDecision string                         `json:"repair_decision"`
		RepairReady    bool                           `json:"repair_ready"`
		RepairOutput   string                         `json:"repair_output"`
		Summary        *browserTopLevelSummary        `json:"summary"`
		Display        *browserTopLevelDisplaySummary `json:"display"`
		Surface        *browserTopLevelSurfaceSummary `json:"surface"`
		View           *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode unified repair payload: %v", err)
	}
	if payload.Action != "repair" || payload.Status != "ok" || payload.RepairDecision != "repaired" || !payload.RepairReady {
		t.Fatalf("unexpected unified repair payload: %#v", payload)
	}
	if !strings.Contains(payload.RepairOutput, "unified-repair-completed") {
		t.Fatalf("expected unified repair to preserve script output, got %#v", payload)
	}
	if payload.Summary == nil ||
		payload.Summary.Category != "coordination" ||
		payload.Summary.State != "completed" ||
		payload.Summary.SummaryCode != "repair_completed" {
		t.Fatalf("expected unified repair payload to expose repair summary aliases, got %#v", payload.Summary)
	}
	if payload.Display == nil ||
		payload.Display.Category != "coordination" ||
		payload.Display.State != "completed" ||
		payload.Display.SummaryCode != "repair_completed" {
		t.Fatalf("expected unified repair payload to expose repair display aliases, got %#v", payload.Display)
	}
	if payload.Surface == nil ||
		payload.Surface.Category != "coordination" ||
		payload.Surface.State != "completed" ||
		payload.Surface.SummaryCode != "repair_completed" {
		t.Fatalf("expected unified repair payload to expose repair surface aliases, got %#v", payload.Surface)
	}
	if payload.View == nil ||
		payload.View.Kind != "result" ||
		payload.View.Category != "coordination" ||
		payload.View.State != "completed" ||
		payload.View.SummaryCode != "repair_completed" {
		t.Fatalf("expected unified repair payload to expose repair view aliases, got %#v", payload.View)
	}
}

func TestBrowserRuntimeDoctorSuggestionsPreferRepairActionWhenAvailable(t *testing.T) {
	root := t.TempDir()
	repairScript := writeBrowserDoctorRepoScriptForTest(t, root, "browserd-bootstrap-repair.sh", "#!/usr/bin/env bash\nset -euo pipefail\necho doctor-repair\n")

	doctor := &BrowserDoctorSummary{
		Launch: &BrowserDoctorLaunchSummary{
			BrowserDoctorCheckSummary: BrowserDoctorCheckSummary{
				Status: "error",
				Code:   "npm_ci_failed",
			},
			BootstrapErrorCode: "npm_ci_failed",
		},
		RepairCommand:     "bash /tmp/browserd-bootstrap-repair.sh",
		AcceptanceCommand: "bash /tmp/browserd-platform-acceptance-check.sh",
	}
	suggestions := browserRuntimeDoctorSuggestions(browserRegistrationContext{
		opts: BrowserToolOptions{
			Root:         root,
			RepairScript: repairScript,
			RunCommand: func(context.Context, string, []string) ([]byte, error) {
				return nil, nil
			},
		},
		enabledTools: map[string]bool{"browser": true, "browser_runtime": true},
	}, doctor)

	for _, expected := range []string{
		"run browser action=repair",
		"run bash /tmp/browserd-bootstrap-repair.sh",
		"run bash /tmp/browserd-platform-acceptance-check.sh",
	} {
		if !browserStringSliceContains(suggestions, expected) {
			t.Fatalf("missing repair suggestion %q from %#v", expected, suggestions)
		}
	}
}
