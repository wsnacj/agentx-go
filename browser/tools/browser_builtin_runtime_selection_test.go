package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
	types "github.com/wsnacj/agentx-go/components/llm"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

func TestRegisterBrowserTools_RuntimeSelectProfile(t *testing.T) {
	reg := llmxtools.NewRegistry()
	hostBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "system",
				BrowserApp: "Safari",
				Profile:    "default",
				Status:     "running",
				Running:    true,
				Connected:  true,
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}}
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "workbench",
				Status:     "running",
				Running:    true,
				Connected:  true,
			},
			runtimeProfilesResult: BrowserProfilesResult{
				Backend: "proxy",
				Profiles: []BrowserProfileInfo{
					{Profile: "isolated", BrowserApp: "Chromium", Status: "stopped"},
					{Profile: "workbench", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
				},
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              hostBackend,
		NodeBackend:          nodeBackend,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-select-session")
	selectOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"select_profile","runtime_target":"node","profile":"workbench"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime select_profile: %v", err)
	}
	var selectPayload struct {
		Action                  string `json:"action"`
		Status                  string `json:"status"`
		SelectDecision          string `json:"select_decision"`
		SelectReady             bool   `json:"select_ready"`
		SessionProfileSelection struct {
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
			Source        string `json:"source"`
		} `json:"session_profile_selection"`
	}
	if err := json.Unmarshal([]byte(selectOut), &selectPayload); err != nil {
		t.Fatalf("decode select output: %v", err)
	}
	if selectPayload.Action != "select_profile" || selectPayload.Status != "ok" || selectPayload.SelectDecision != "session_profile_selected" || !selectPayload.SelectReady {
		t.Fatalf("unexpected select_profile payload: %#v", selectPayload)
	}
	if selectPayload.SessionProfileSelection.Profile != "workbench" || selectPayload.SessionProfileSelection.RuntimeTarget != "node" {
		t.Fatalf("unexpected session profile selection payload: %#v", selectPayload.SessionProfileSelection)
	}
	if selectPayload.SessionProfileSelection.Source != "select_profile" {
		t.Fatalf("unexpected session profile selection source: %#v", selectPayload.SessionProfileSelection)
	}

	statusOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"status"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime status after select_profile: %v", err)
	}
	if len(hostBackend.runtimeStatusReqs) != 0 {
		t.Fatalf("expected host backend status to stay unused after session profile selection, got %#v", hostBackend.runtimeStatusReqs)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || nodeBackend.runtimeStatusReqs[0].Profile != "workbench" {
		t.Fatalf("expected node backend status to use selected profile, got %#v", nodeBackend.runtimeStatusReqs)
	}
	var statusPayload struct {
		SelectedRoute struct {
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		RouteResolution struct {
			ProfileSource       string `json:"profile_source"`
			RuntimeTargetSource string `json:"runtime_target_source"`
		} `json:"route_resolution"`
		SessionBinding struct {
			SelectedBrowserProfile       string `json:"selected_browser_profile"`
			SelectedBrowserProfileSource string `json:"selected_browser_profile_source"`
			SelectedBrowserTarget        string `json:"selected_browser_target"`
		} `json:"session_binding"`
		SessionProfileSelection struct {
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
			Source        string `json:"source"`
		} `json:"session_profile_selection"`
	}
	if err := json.Unmarshal([]byte(statusOut), &statusPayload); err != nil {
		t.Fatalf("decode status output: %v", err)
	}
	if statusPayload.SelectedRoute.Profile != "workbench" || statusPayload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected selected route to reuse session profile selection, got %#v", statusPayload.SelectedRoute)
	}
	if statusPayload.SessionBinding.SelectedBrowserProfile != "workbench" || statusPayload.SessionBinding.SelectedBrowserTarget != "node" {
		t.Fatalf("unexpected session binding selected profile: %#v", statusPayload.SessionBinding)
	}
	if statusPayload.SessionBinding.SelectedBrowserProfileSource != "select_profile" {
		t.Fatalf("unexpected session binding profile source: %#v", statusPayload.SessionBinding)
	}
	if statusPayload.RouteResolution.ProfileSource != "select_profile" || statusPayload.RouteResolution.RuntimeTargetSource != "select_profile" {
		t.Fatalf("unexpected route resolution: %#v", statusPayload.RouteResolution)
	}
	if statusPayload.SessionProfileSelection.Profile != "workbench" || statusPayload.SessionProfileSelection.RuntimeTarget != "node" {
		t.Fatalf("unexpected session profile selection in status payload: %#v", statusPayload.SessionProfileSelection)
	}
	if statusPayload.SessionProfileSelection.Source != "select_profile" {
		t.Fatalf("unexpected session profile selection source in status payload: %#v", statusPayload.SessionProfileSelection)
	}
}

func TestRegisterBrowserTools_RuntimeSelectProfilePromotesCurrentTargetWhenAvailable(t *testing.T) {
	reg := llmxtools.NewRegistry()
	hostBackend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{
		extractResult: BrowserExtractResult{
			Backend:    "system-extract",
			BrowserApp: "Safari",
			Title:      "Host",
			Content:    "Host content",
			FinalURL:   "https://host.example/home",
		},
	}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}}
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			extractResult: BrowserExtractResult{
				Backend:    "proxy-extract",
				BrowserApp: "Chromium",
				Title:      "Workbench",
				Content:    "Node content",
				FinalURL:   "https://node.example/workbench",
			},
			runtimeProfilesResult: BrowserProfilesResult{
				Backend:        "proxy",
				DefaultProfile: "isolated",
				Profiles: []BrowserProfileInfo{
					{Profile: "isolated", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
					{Profile: "workbench", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
				},
			},
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "workbench",
				Status:     "running",
				Running:    true,
				Connected:  true,
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              hostBackend,
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime", "browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-select-profile-promotes-target")
	sessionRegistry.TrackTab("browser-runtime-select-profile-promotes-target", BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, true)

	selectOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"select_profile","runtime_target":"node","profile":"workbench"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime select_profile promote target: %v", err)
	}
	var selectPayload struct {
		Action                  string `json:"action"`
		Status                  string `json:"status"`
		SelectDecision          string `json:"select_decision"`
		SelectReady             bool   `json:"select_ready"`
		SessionProfileSelection struct {
			Profile string `json:"profile"`
			Source  string `json:"source"`
		} `json:"session_profile_selection"`
		SessionTargetSelection struct {
			TabIndex int    `json:"tab_index"`
			Source   string `json:"source"`
		} `json:"session_target_selection"`
		SessionBinding struct {
			SelectedBrowserProfile       string `json:"selected_browser_profile"`
			SelectedBrowserProfileSource string `json:"selected_browser_profile_source"`
			SelectedBrowserTargetID      string `json:"selected_browser_target_id"`
			SelectedBrowserTargetSource  string `json:"selected_browser_target_source"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(selectOut), &selectPayload); err != nil {
		t.Fatalf("decode select_profile promote target output: %v", err)
	}
	if selectPayload.Action != "select_profile" || selectPayload.Status != "ok" || selectPayload.SelectDecision != "session_profile_selected" || !selectPayload.SelectReady {
		t.Fatalf("unexpected select_profile promote target payload: %#v", selectPayload)
	}
	if selectPayload.SessionProfileSelection.Profile != "workbench" || selectPayload.SessionProfileSelection.Source != "select_profile" {
		t.Fatalf("unexpected session profile selection after select_profile promote target: %#v", selectPayload.SessionProfileSelection)
	}
	if selectPayload.SessionTargetSelection.TabIndex != 2 || selectPayload.SessionTargetSelection.Source != "select_profile" {
		t.Fatalf("expected select_profile to promote current target, got %#v", selectPayload.SessionTargetSelection)
	}
	if selectPayload.SessionBinding.SelectedBrowserProfile != "workbench" || selectPayload.SessionBinding.SelectedBrowserProfileSource != "select_profile" {
		t.Fatalf("unexpected session binding selected profile after select_profile promote target: %#v", selectPayload.SessionBinding)
	}
	if selectPayload.SessionBinding.SelectedBrowserTargetID == "" || selectPayload.SessionBinding.SelectedBrowserTargetSource != "select_profile" {
		t.Fatalf("expected session binding selected target to be promoted by select_profile, got %#v", selectPayload.SessionBinding)
	}

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"extract","max_chars":20}`,
	})
	if err != nil {
		t.Fatalf("browser_act extract after select_profile promote target: %v", err)
	}
	if len(hostBackend.extractReqs) != 0 {
		t.Fatalf("expected host backend extract to stay unused, got %#v", hostBackend.extractReqs)
	}
	if len(nodeBackend.extractReqs) != 1 || nodeBackend.extractReqs[0].TabIndex != 2 || nodeBackend.extractReqs[0].WaitMs != browserTabTargetWaitMs {
		t.Fatalf("expected node backend extract to inherit promoted target after select_profile, got %#v", nodeBackend.extractReqs)
	}
	var payload struct {
		Kind     string `json:"kind"`
		TargetID string `json:"target_id"`
		TabIndex int    `json:"tab_index"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_act extract after select_profile promote target: %v", err)
	}
	if payload.Kind != "extract" || payload.TabIndex != 2 || payload.TargetID == "" {
		t.Fatalf("unexpected browser_act payload after select_profile promote target: %#v", payload)
	}
}

func TestRegisterBrowserTools_RuntimeSelectProfileClearsMismatchedTargetSelection(t *testing.T) {
	reg := llmxtools.NewRegistry()
	hostBackend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{
		extractResult: BrowserExtractResult{
			Backend:    "system-extract",
			BrowserApp: "Safari",
			Title:      "Host",
			Content:    "Host content",
			FinalURL:   "https://host.example/home",
		},
	}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}}
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			extractResult: BrowserExtractResult{
				Backend:    "proxy-extract",
				BrowserApp: "Chromium",
				Title:      "Alternate",
				Content:    "Alternate content",
				FinalURL:   "https://node.example/alternate",
			},
			runtimeProfilesResult: BrowserProfilesResult{
				Backend:        "proxy",
				DefaultProfile: "isolated",
				Profiles: []BrowserProfileInfo{
					{Profile: "isolated", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
					{Profile: "workbench", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
					{Profile: "alternate", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
				},
			},
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "alternate",
				Status:     "running",
				Running:    true,
				Connected:  true,
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              hostBackend,
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime", "browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-select-profile-clears-mismatched-target")
	sessionRegistry.TrackTab("browser-runtime-select-profile-clears-mismatched-target", BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, true)

	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"select_target","runtime_target":"node","profile":"workbench","target":"tab:2"}`,
	}); err != nil {
		t.Fatalf("browser_runtime select_target before profile switch: %v", err)
	}

	selectOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"select_profile","runtime_target":"node","profile":"alternate"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime select_profile alternate: %v", err)
	}
	var selectPayload struct {
		Action                  string                                 `json:"action"`
		Status                  string                                 `json:"status"`
		SelectDecision          string                                 `json:"select_decision"`
		SelectReady             bool                                   `json:"select_ready"`
		SessionProfileSelection *browserRuntimeSessionProfileSelection `json:"session_profile_selection"`
		SessionTargetSelection  *browserRuntimeSessionTargetSelection  `json:"session_target_selection"`
		SessionBinding          struct {
			SelectedBrowserProfile      string `json:"selected_browser_profile"`
			SelectedBrowserTargetID     string `json:"selected_browser_target_id"`
			SelectedBrowserTargetSource string `json:"selected_browser_target_source"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(selectOut), &selectPayload); err != nil {
		t.Fatalf("decode select_profile alternate output: %v", err)
	}
	if selectPayload.Action != "select_profile" || selectPayload.Status != "ok" || selectPayload.SelectDecision != "session_profile_selected" || !selectPayload.SelectReady {
		t.Fatalf("unexpected select_profile alternate payload: %#v", selectPayload)
	}
	if selectPayload.SessionProfileSelection == nil || selectPayload.SessionProfileSelection.Profile != "alternate" || selectPayload.SessionProfileSelection.Source != "select_profile" {
		t.Fatalf("expected alternate profile selection after profile switch, got %#v", selectPayload.SessionProfileSelection)
	}
	if selectPayload.SessionTargetSelection != nil {
		t.Fatalf("expected mismatched remembered target to be cleared after profile switch, got %#v", selectPayload.SessionTargetSelection)
	}
	if selectPayload.SessionBinding.SelectedBrowserProfile != "alternate" {
		t.Fatalf("expected session binding to switch to alternate profile, got %#v", selectPayload.SessionBinding)
	}
	if selectPayload.SessionBinding.SelectedBrowserTargetID != "" || selectPayload.SessionBinding.SelectedBrowserTargetSource != "" {
		t.Fatalf("expected session binding target to clear after profile switch, got %#v", selectPayload.SessionBinding)
	}

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"extract","max_chars":20}`,
	})
	if err != nil {
		t.Fatalf("browser_act extract after select_profile alternate: %v", err)
	}
	if len(hostBackend.extractReqs) != 0 {
		t.Fatalf("expected host backend extract to stay unused, got %#v", hostBackend.extractReqs)
	}
	if len(nodeBackend.extractReqs) != 1 || nodeBackend.extractReqs[0].TabIndex != 0 || nodeBackend.extractReqs[0].WaitMs != browserTabTargetWaitMs {
		t.Fatalf("expected node backend extract to avoid old remembered target after profile switch, got %#v", nodeBackend.extractReqs)
	}
	var payload struct {
		Kind     string `json:"kind"`
		TargetID string `json:"target_id"`
		TabIndex int    `json:"tab_index"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_act extract after select_profile alternate: %v", err)
	}
	if payload.Kind != "extract" || payload.TabIndex != 0 || strings.TrimSpace(payload.TargetID) == "" {
		t.Fatalf("unexpected browser_act payload after select_profile alternate: %#v", payload)
	}
}

func TestRegisterBrowserTools_RuntimeSelectProfileAlreadySelected(t *testing.T) {
	reg := llmxtools.NewRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeProfilesResult: BrowserProfilesResult{
				Backend: "proxy",
				Profiles: []BrowserProfileInfo{
					{Profile: "workbench", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
				},
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          nodeBackend,
		SessionStateRegistry: NewBrowserSessionStateRegistry(),
		EnabledTools:         []string{"browser_runtime"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-select-already")
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"select_profile","runtime_target":"node","profile":"workbench"}`,
	}); err != nil {
		t.Fatalf("browser_runtime select_profile first: %v", err)
	}
	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"select_profile","runtime_target":"node","profile":"workbench"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime select_profile second: %v", err)
	}
	var payload struct {
		Action         string `json:"action"`
		Status         string `json:"status"`
		SelectDecision string `json:"select_decision"`
		SelectReady    bool   `json:"select_ready"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode select_profile already output: %v", err)
	}
	if payload.Action != "select_profile" || payload.Status != "ok" || payload.SelectDecision != "session_profile_already_selected" || !payload.SelectReady {
		t.Fatalf("unexpected select_profile already payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_RuntimeSelectProfileInvalidatesCachedStatusWatch(t *testing.T) {
	reg := llmxtools.NewRegistry()
	hostBackend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}}
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "workbench",
				Status:     "running",
				Running:    true,
				Connected:  true,
			},
			runtimeProfilesResult: BrowserProfilesResult{
				Backend:        "proxy",
				DefaultProfile: "isolated",
				Profiles: []BrowserProfileInfo{
					{Profile: "isolated", BrowserApp: "Chromium", Status: "stopped"},
					{Profile: "workbench", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
				},
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              hostBackend,
		NodeBackend:          nodeBackend,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-select-profile-invalidates-status-watch")
	statusArgs := `{"action":"status","runtime_target":"node","profile":"workbench"}`

	statusOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: statusArgs,
	})
	if err != nil {
		t.Fatalf("browser_runtime status before select_profile cache: %v", err)
	}
	var initial struct {
		SessionBinding struct {
			SelectedBrowserProfile string `json:"selected_browser_profile"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(statusOut), &initial); err != nil {
		t.Fatalf("decode initial status output: %v", err)
	}
	if initial.SessionBinding.SelectedBrowserProfile != "" {
		t.Fatalf("expected initial cached status to have no selected browser profile, got %#v", initial.SessionBinding)
	}

	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"select_profile","runtime_target":"node","profile":"workbench"}`,
	}); err != nil {
		t.Fatalf("browser_runtime select_profile after cached status: %v", err)
	}

	statusOut, err = reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: statusArgs,
	})
	if err != nil {
		t.Fatalf("browser_runtime status after select_profile cache invalidation: %v", err)
	}
	var updated struct {
		SessionBinding struct {
			SelectedBrowserProfile       string `json:"selected_browser_profile"`
			SelectedBrowserProfileSource string `json:"selected_browser_profile_source"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(statusOut), &updated); err != nil {
		t.Fatalf("decode updated status output: %v", err)
	}
	if updated.SessionBinding.SelectedBrowserProfile != "workbench" || updated.SessionBinding.SelectedBrowserProfileSource != "select_profile" {
		t.Fatalf("expected select_profile to invalidate cached status watch, got %#v", updated.SessionBinding)
	}
}

func TestRegisterBrowserTools_RuntimeSelectProfileMissingProfile(t *testing.T) {
	reg := llmxtools.NewRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeProfilesResult: BrowserProfilesResult{
				Backend: "proxy",
				Profiles: []BrowserProfileInfo{
					{Profile: "isolated", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
				},
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          nodeBackend,
		SessionStateRegistry: NewBrowserSessionStateRegistry(),
		EnabledTools:         []string{"browser_runtime"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-select-missing")
	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"select_profile","runtime_target":"node","profile":"workbench"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime select_profile missing: %v", err)
	}
	var payload struct {
		Action                  string `json:"action"`
		Status                  string `json:"status"`
		SelectDecision          string `json:"select_decision"`
		Note                    string `json:"note"`
		SessionProfileSelection any    `json:"session_profile_selection"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode select_profile missing output: %v", err)
	}
	if payload.Action != "select_profile" || payload.Status != "error" || payload.SelectDecision != "session_profile_missing" || !strings.Contains(payload.Note, `profile "workbench" is not available`) {
		t.Fatalf("unexpected select_profile missing payload: %#v", payload)
	}
	if payload.SessionProfileSelection != nil {
		t.Fatalf("expected missing select_profile not to create session selection, got %#v", payload.SessionProfileSelection)
	}
}

func TestRegisterBrowserTools_RuntimeClearProfile(t *testing.T) {
	reg := llmxtools.NewRegistry()
	hostBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "system",
				BrowserApp: "Safari",
				Profile:    "default",
				Status:     "running",
				Running:    true,
				Connected:  true,
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}}
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "workbench",
				Status:     "running",
				Running:    true,
				Connected:  true,
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              hostBackend,
		NodeBackend:          nodeBackend,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-clear-session")
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"select_profile","runtime_target":"node","profile":"workbench"}`,
	}); err != nil {
		t.Fatalf("browser_runtime select_profile: %v", err)
	}
	clearOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"clear_profile"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime clear_profile: %v", err)
	}
	var clearPayload struct {
		Action        string `json:"action"`
		Status        string `json:"status"`
		ClearDecision string `json:"clear_decision"`
		ClearReady    bool   `json:"clear_ready"`
		SelectedRoute struct {
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		SessionBinding struct {
			SelectedBrowserProfile string `json:"selected_browser_profile"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(clearOut), &clearPayload); err != nil {
		t.Fatalf("decode clear output: %v", err)
	}
	if clearPayload.Action != "clear_profile" || clearPayload.Status != "ok" || clearPayload.ClearDecision != "session_profile_cleared" || !clearPayload.ClearReady {
		t.Fatalf("unexpected clear_profile payload: %#v", clearPayload)
	}
	if clearPayload.SessionBinding.SelectedBrowserProfile != "" {
		t.Fatalf("expected session binding to clear selected profile, got %#v", clearPayload.SessionBinding)
	}

	statusOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"status"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime status after clear_profile: %v", err)
	}
	if len(hostBackend.runtimeStatusReqs) != 0 {
		t.Fatalf("expected host backend status to stay unused after clearing session profile, got %#v", hostBackend.runtimeStatusReqs)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || nodeBackend.runtimeStatusReqs[0].Profile != "isolated" {
		t.Fatalf("expected default node route status after clearing session profile, got %#v", nodeBackend.runtimeStatusReqs)
	}
	var statusPayload struct {
		SelectedRoute struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		SessionProfileSelection any `json:"session_profile_selection"`
	}
	if err := json.Unmarshal([]byte(statusOut), &statusPayload); err != nil {
		t.Fatalf("decode status output: %v", err)
	}
	if statusPayload.SelectedRoute.Backend != "proxy" || statusPayload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected status to fall back to promoted node route after clear_profile, got %#v", statusPayload.SelectedRoute)
	}
	if statusPayload.SessionProfileSelection != nil {
		t.Fatalf("expected session profile selection to be cleared, got %#v", statusPayload.SessionProfileSelection)
	}
}

func TestRegisterBrowserTools_RuntimeClearProfileBlockedByActiveNodeRun(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRunRegistry := newTestSessionRunRegistry()
	sessionRunRegistry.Record("browser-runtime-clear-profile-blocked", agentxbrowserruntime.SharedSessionRunInfo{
		RunID:    "run-81",
		NodeID:   "node-alpha",
		Status:   "running",
		Provider: "gateway",
		Action:   "run_status",
	})
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"}},
		SessionRunRegistry:   sessionRunRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-clear-profile-blocked")
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"select_profile","runtime_target":"node","profile":"workbench"}`,
	}); err != nil {
		t.Fatalf("browser_runtime select_profile: %v", err)
	}
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-clear-profile-blocked", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "workbench",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "running",
		Running:       true,
		Connected:     true,
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"clear_profile","runtime_target":"node"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime clear_profile blocked: %v", err)
	}
	var payload struct {
		Action                  string `json:"action"`
		Status                  string `json:"status"`
		ClearDecision           string `json:"clear_decision"`
		ClearReady              bool   `json:"clear_ready"`
		Force                   bool   `json:"force"`
		SessionProfileSelection struct {
			Profile string `json:"profile"`
			Source  string `json:"source"`
		} `json:"session_profile_selection"`
		SessionBinding struct {
			SelectedBrowserProfile       string `json:"selected_browser_profile"`
			SelectedBrowserProfileSource string `json:"selected_browser_profile_source"`
			ActiveNodeRunID              string `json:"active_node_run_id"`
		} `json:"session_binding"`
		ProfileStatus struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
			Status        string `json:"status"`
			Running       bool   `json:"running"`
			Connected     bool   `json:"connected"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode clear_profile blocked output: %v", err)
	}
	if payload.Action != "clear_profile" || payload.Status != "ok" || payload.ClearDecision != "clear_profile_blocked_active_node_run" || payload.ClearReady || payload.Force {
		t.Fatalf("unexpected blocked clear_profile payload: %#v", payload)
	}
	if payload.SessionProfileSelection.Profile != "workbench" || payload.SessionProfileSelection.Source != "select_profile" {
		t.Fatalf("expected blocked clear_profile to preserve session selection, got %#v", payload.SessionProfileSelection)
	}
	if payload.SessionBinding.SelectedBrowserProfile != "workbench" || payload.SessionBinding.SelectedBrowserProfileSource != "select_profile" || payload.SessionBinding.ActiveNodeRunID != "run-81" {
		t.Fatalf("expected blocked clear_profile to preserve session binding, got %#v", payload.SessionBinding)
	}
	if payload.ProfileStatus.Backend != "proxy" || payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.RuntimeTarget != "node" || payload.ProfileStatus.Status != "running" || !payload.ProfileStatus.Running || !payload.ProfileStatus.Connected {
		t.Fatalf("expected blocked clear_profile to preserve effective lifecycle state, got %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_RuntimeClearProfileForceBypassesActiveNodeRun(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRunRegistry := newTestSessionRunRegistry()
	sessionRunRegistry.Record("browser-runtime-clear-profile-force", agentxbrowserruntime.SharedSessionRunInfo{
		RunID:    "run-82",
		NodeID:   "node-alpha",
		Status:   "running",
		Provider: "gateway",
		Action:   "run_status",
	})
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"}},
		SessionRunRegistry:   sessionRunRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-clear-profile-force")
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"select_profile","runtime_target":"node","profile":"workbench"}`,
	}); err != nil {
		t.Fatalf("browser_runtime select_profile: %v", err)
	}

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"clear_profile","runtime_target":"node","force":true}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime clear_profile force: %v", err)
	}
	var payload struct {
		Action                  string `json:"action"`
		Status                  string `json:"status"`
		ClearDecision           string `json:"clear_decision"`
		ClearReady              bool   `json:"clear_ready"`
		Force                   bool   `json:"force"`
		SessionProfileSelection any    `json:"session_profile_selection"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode clear_profile force output: %v", err)
	}
	if payload.Action != "clear_profile" || payload.Status != "ok" || payload.ClearDecision != "session_profile_cleared" || !payload.ClearReady || !payload.Force {
		t.Fatalf("unexpected forced clear_profile payload: %#v", payload)
	}
	if payload.SessionProfileSelection != nil {
		t.Fatalf("expected forced clear_profile to clear session selection, got %#v", payload.SessionProfileSelection)
	}
}

func TestRegisterBrowserTools_RuntimeClearProfileClearsAutoPromotedTarget(t *testing.T) {
	reg := llmxtools.NewRegistry()
	hostBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "system",
				BrowserApp: "Safari",
				Profile:    "default",
				Status:     "running",
				Running:    true,
				Connected:  true,
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}}
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "workbench",
				Status:     "running",
				Running:    true,
				Connected:  true,
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              hostBackend,
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-clear-profile-clears-target")
	sessionRegistry.TrackTab("browser-runtime-clear-profile-clears-target", BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://node.example/two",
		Title:      "Two",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, true)
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"select_profile","runtime_target":"node","profile":"workbench"}`,
	}); err != nil {
		t.Fatalf("browser_runtime select_profile: %v", err)
	}

	clearOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"clear_profile"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime clear_profile: %v", err)
	}
	var clearPayload struct {
		SessionProfileSelection any `json:"session_profile_selection"`
		SessionTargetSelection  any `json:"session_target_selection"`
		SessionBinding          struct {
			SelectedBrowserProfile  string `json:"selected_browser_profile"`
			SelectedBrowserTargetID string `json:"selected_browser_target_id"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(clearOut), &clearPayload); err != nil {
		t.Fatalf("decode clear_profile output: %v", err)
	}
	if clearPayload.SessionProfileSelection != nil {
		t.Fatalf("expected clear_profile to clear session profile selection, got %#v", clearPayload.SessionProfileSelection)
	}
	if clearPayload.SessionTargetSelection != nil {
		t.Fatalf("expected clear_profile to clear auto-promoted target selection, got %#v", clearPayload.SessionTargetSelection)
	}
	if clearPayload.SessionBinding.SelectedBrowserProfile != "" || clearPayload.SessionBinding.SelectedBrowserTargetID != "" {
		t.Fatalf("expected clear_profile to clear auto-promoted browser selections from binding, got %#v", clearPayload.SessionBinding)
	}
}

func TestRegisterBrowserTools_RuntimeSelectTargetPropagatesToBrowserAct(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	hostBackend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}}
	nodeBackend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{
		extractResult: BrowserExtractResult{
			Backend:    "proxy-extract",
			BrowserApp: "Chromium",
			Title:      "Workbench",
			Content:    "Extracted content",
			FinalURL:   "https://node.example/workbench",
		},
	}, runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              hostBackend,
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime", "browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-select-target")
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"select_profile","runtime_target":"node","profile":"workbench"}`,
	}); err != nil {
		t.Fatalf("browser_runtime select_profile: %v", err)
	}
	sessionRegistry.TrackTab("browser-runtime-select-target", BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://node.example/one",
		Title:      "One",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, false)
	second := sessionRegistry.TrackTab("browser-runtime-select-target", BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://node.example/two",
		Title:      "Two",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, false)

	selectOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: fmt.Sprintf(`{"action":"select_target","target":"target:%s"}`, second.ID),
	})
	if err != nil {
		t.Fatalf("browser_runtime select_target: %v", err)
	}
	var selectPayload struct {
		Action                 string `json:"action"`
		Status                 string `json:"status"`
		SelectTargetDecision   string `json:"select_target_decision"`
		SelectTargetReady      bool   `json:"select_target_ready"`
		SessionTargetSelection struct {
			ID       string `json:"id"`
			TabIndex int    `json:"tab_index"`
			Source   string `json:"source"`
		} `json:"session_target_selection"`
		SessionBinding struct {
			SelectedBrowserTargetID     string `json:"selected_browser_target_id"`
			SelectedBrowserTabIndex     int    `json:"selected_browser_tab_index"`
			SelectedBrowserTargetSource string `json:"selected_browser_target_source"`
			CurrentTargetID             string `json:"current_target_id"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(selectOut), &selectPayload); err != nil {
		t.Fatalf("decode select_target output: %v", err)
	}
	if selectPayload.Action != "select_target" || selectPayload.Status != "ok" || selectPayload.SelectTargetDecision != "session_target_selected" || !selectPayload.SelectTargetReady {
		t.Fatalf("unexpected select_target payload: %#v", selectPayload)
	}
	if selectPayload.SessionTargetSelection.ID != second.ID || selectPayload.SessionTargetSelection.TabIndex != 2 {
		t.Fatalf("unexpected session target selection: %#v", selectPayload.SessionTargetSelection)
	}
	if selectPayload.SessionTargetSelection.Source != "select_target" {
		t.Fatalf("unexpected session target selection source: %#v", selectPayload.SessionTargetSelection)
	}
	if selectPayload.SessionBinding.SelectedBrowserTargetID != second.ID || selectPayload.SessionBinding.SelectedBrowserTabIndex != 2 || selectPayload.SessionBinding.CurrentTargetID != second.ID {
		t.Fatalf("unexpected session binding selected target: %#v", selectPayload.SessionBinding)
	}
	if selectPayload.SessionBinding.SelectedBrowserTargetSource != "select_target" {
		t.Fatalf("unexpected session binding selected target source: %#v", selectPayload.SessionBinding)
	}

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"extract","max_chars":20}`,
	})
	if err != nil {
		t.Fatalf("browser_act extract after select_target: %v", err)
	}
	if len(hostBackend.extractReqs) != 0 {
		t.Fatalf("expected host backend extract to stay unused, got %#v", hostBackend.extractReqs)
	}
	if len(nodeBackend.extractReqs) != 1 || nodeBackend.extractReqs[0].TabIndex != 2 || nodeBackend.extractReqs[0].WaitMs != browserTabTargetWaitMs {
		t.Fatalf("expected node backend extract to inherit selected target, got %#v", nodeBackend.extractReqs)
	}
	var payload struct {
		Kind     string `json:"kind"`
		Target   string `json:"target"`
		TargetID string `json:"target_id"`
		TabIndex int    `json:"tab_index"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_act output: %v", err)
	}
	if payload.Kind != "extract" || payload.Target != "target:"+second.ID || payload.TargetID != second.ID || payload.TabIndex != 2 {
		t.Fatalf("unexpected browser_act payload after select_target: %#v", payload)
	}
}

func TestRegisterBrowserTools_RuntimeSelectTargetInvalidatesCachedStatusWatch(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	hostBackend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}}
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "workbench",
				Status:     "running",
				Running:    true,
				Connected:  true,
			},
			runtimeProfilesResult: BrowserProfilesResult{
				Backend:        "proxy",
				DefaultProfile: "isolated",
				Profiles: []BrowserProfileInfo{
					{Profile: "workbench", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
				},
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              hostBackend,
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-select-target-invalidates-status-watch")
	second := sessionRegistry.TrackTab("browser-runtime-select-target-invalidates-status-watch", BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://node.example/two",
		Title:      "Two",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, false)
	statusArgs := `{"action":"status","runtime_target":"node","profile":"workbench"}`

	statusOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: statusArgs,
	})
	if err != nil {
		t.Fatalf("browser_runtime status before select_target cache: %v", err)
	}
	var initial struct {
		SessionBinding struct {
			SelectedBrowserTargetID string `json:"selected_browser_target_id"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(statusOut), &initial); err != nil {
		t.Fatalf("decode initial status before select_target: %v", err)
	}
	if initial.SessionBinding.SelectedBrowserTargetID != "" {
		t.Fatalf("expected initial cached status to have no selected target, got %#v", initial.SessionBinding)
	}

	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: fmt.Sprintf(`{"action":"select_target","runtime_target":"node","profile":"workbench","target":"target:%s"}`, second.ID),
	}); err != nil {
		t.Fatalf("browser_runtime select_target after cached status: %v", err)
	}

	statusOut, err = reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: statusArgs,
	})
	if err != nil {
		t.Fatalf("browser_runtime status after select_target cache invalidation: %v", err)
	}
	var updated struct {
		SessionBinding struct {
			SelectedBrowserTargetID     string `json:"selected_browser_target_id"`
			SelectedBrowserTargetSource string `json:"selected_browser_target_source"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(statusOut), &updated); err != nil {
		t.Fatalf("decode updated status after select_target: %v", err)
	}
	if updated.SessionBinding.SelectedBrowserTargetID != second.ID || updated.SessionBinding.SelectedBrowserTargetSource != "select_target" {
		t.Fatalf("expected select_target to invalidate cached status watch, got %#v", updated.SessionBinding)
	}
}

func TestRegisterBrowserTools_RuntimeSelectTargetRequiresPendingPopupReview(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	hostBackend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}}
	nodeBackend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{
		extractResult: BrowserExtractResult{
			Backend:    "proxy-extract",
			BrowserApp: "Chromium",
			Title:      "One",
			Content:    "Extracted content",
			FinalURL:   "https://node.example/one",
		},
	}, runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              hostBackend,
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime", "browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-select-target-popup-review")
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"select_profile","runtime_target":"node","profile":"workbench"}`,
	}); err != nil {
		t.Fatalf("browser_runtime select_profile: %v", err)
	}
	current := sessionRegistry.TrackTab("browser-runtime-select-target-popup-review", BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://node.example/one",
		Title:      "One",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, true)
	popup := sessionRegistry.TrackTab("browser-runtime-select-target-popup-review", BrowserSessionTarget{
		TabIndex:   3,
		URL:        "https://popup.example/offer",
		Title:      "Offer",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, false)
	sessionRegistry.RecordPendingTargetReviewForRoute("browser-runtime-select-target-popup-review", BrowserSessionRoute{
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
		BrowserApp: "Chromium",
	}, agentxbrowserruntime.BrowserSessionTargetReview{
		ID:         popup.ID,
		TabIndex:   popup.TabIndex,
		URL:        popup.URL,
		Title:      popup.Title,
		BrowserApp: popup.BrowserApp,
		Backend:    popup.Backend,
		Profile:    popup.Profile,
		Target:     popup.Target,
		Decision:   "session_target_popup_review_required",
		Reason:     "pending popup review",
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: fmt.Sprintf(`{"action":"select_target","runtime_target":"node","profile":"workbench","target":"target:%s"}`, popup.ID),
	})
	if err != nil {
		t.Fatalf("browser_runtime select_target popup review: %v", err)
	}
	var payload struct {
		Action                 string `json:"action"`
		Status                 string `json:"status"`
		Force                  bool   `json:"force"`
		SelectTargetDecision   string `json:"select_target_decision"`
		SelectTargetReady      bool   `json:"select_target_ready"`
		Note                   string `json:"note"`
		SessionTargetSelection struct {
			ID       string `json:"id"`
			TabIndex int    `json:"tab_index"`
			Source   string `json:"source"`
		} `json:"session_target_selection"`
		SessionBinding struct {
			CurrentTargetID          string `json:"current_target_id"`
			SelectedBrowserTargetID  string `json:"selected_browser_target_id"`
			PendingTargetReviewCount int    `json:"pending_target_review_count"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode select_target popup review output: %v", err)
	}
	if payload.Action != "select_target" || payload.Status != "review_required" || payload.Force || payload.SelectTargetDecision != "session_target_popup_review_required" || payload.SelectTargetReady {
		t.Fatalf("unexpected select_target popup review payload: %#v", payload)
	}
	if payload.SessionTargetSelection.ID != current.ID || payload.SessionTargetSelection.TabIndex != 1 || payload.SessionTargetSelection.Source != "tracked_active_tab" {
		t.Fatalf("expected popup review to preserve prior session target selection, got %#v", payload.SessionTargetSelection)
	}
	if payload.SessionBinding.CurrentTargetID != current.ID || payload.SessionBinding.SelectedBrowserTargetID == popup.ID || payload.SessionBinding.PendingTargetReviewCount != 1 || !strings.Contains(payload.Note, "pending popup target") {
		t.Fatalf("expected popup review to preserve current target and pending review, got %#v", payload)
	}

	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"extract","runtime_target":"node","profile":"workbench","max_chars":20}`,
	}); err != nil {
		t.Fatalf("browser_act extract after blocked select_target: %v", err)
	}
	if len(hostBackend.extractReqs) != 0 {
		t.Fatalf("expected host backend extract to stay unused, got %#v", hostBackend.extractReqs)
	}
	if len(nodeBackend.extractReqs) != 1 || nodeBackend.extractReqs[0].TabIndex != 1 {
		t.Fatalf("expected node extract to stay on prior current target after blocked select_target, got %#v", nodeBackend.extractReqs)
	}
}

func TestRegisterBrowserTools_RuntimeSelectTargetRequiresRedirectReview(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}},
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-select-target-redirect-review")
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"select_profile","runtime_target":"node","profile":"workbench"}`,
	}); err != nil {
		t.Fatalf("browser_runtime select_profile: %v", err)
	}
	current := sessionRegistry.TrackTab("browser-runtime-select-target-redirect-review", BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://93.184.216.35/landing",
		Title:      "Redirected",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, true)
	sessionRegistry.RecordPendingTargetReviewForRoute("browser-runtime-select-target-redirect-review", BrowserSessionRoute{
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
		BrowserApp: "Chromium",
	}, agentxbrowserruntime.BrowserSessionTargetReview{
		ID:         current.ID,
		TabIndex:   current.TabIndex,
		URL:        current.URL,
		Title:      current.Title,
		BrowserApp: current.BrowserApp,
		Backend:    current.Backend,
		Target:     current.Target,
		Decision:   "session_target_redirect_review_required",
		Reason:     "browser navigation redirected across origin from \"https://93.184.216.34\" to \"https://93.184.216.35/landing\"; rerun with force=true after review",
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: fmt.Sprintf(`{"action":"select_target","runtime_target":"node","profile":"workbench","target":"target:%s"}`, current.ID),
	})
	if err != nil {
		t.Fatalf("browser_runtime select_target redirect review: %v", err)
	}
	var payload struct {
		Action                 string `json:"action"`
		Status                 string `json:"status"`
		SelectTargetDecision   string `json:"select_target_decision"`
		SelectTargetReady      bool   `json:"select_target_ready"`
		Note                   string `json:"note"`
		SessionTargetSelection struct {
			ID       string `json:"id"`
			TabIndex int    `json:"tab_index"`
			Source   string `json:"source"`
		} `json:"session_target_selection"`
		SessionBinding struct {
			CurrentTargetID          string `json:"current_target_id"`
			SelectedBrowserTargetID  string `json:"selected_browser_target_id"`
			PendingTargetReviewCount int    `json:"pending_target_review_count"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode select_target redirect review output: %v", err)
	}
	if payload.Action != "select_target" || payload.Status != "review_required" || payload.SelectTargetDecision != "session_target_redirect_review_required" || payload.SelectTargetReady {
		t.Fatalf("unexpected select_target redirect review payload: %#v", payload)
	}
	if payload.SessionBinding.CurrentTargetID != current.ID || payload.SessionBinding.SelectedBrowserTargetID != current.ID || payload.SessionBinding.PendingTargetReviewCount != 1 || !strings.Contains(payload.Note, "redirected target") {
		t.Fatalf("expected redirect review to preserve current target and note, got %#v", payload)
	}
}

func TestRegisterBrowserTools_RuntimeSelectTargetRequiresPopupStormReview(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	hostBackend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}}
	nodeBackend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{
		extractResult: BrowserExtractResult{
			Backend:    "proxy-extract",
			BrowserApp: "Chromium",
			Title:      "One",
			Content:    "Extracted content",
			FinalURL:   "https://node.example/one",
		},
	}, runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              hostBackend,
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime", "browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-select-target-popup-storm")
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"select_profile","runtime_target":"node","profile":"workbench"}`,
	}); err != nil {
		t.Fatalf("browser_runtime select_profile: %v", err)
	}
	current := sessionRegistry.TrackTab("browser-runtime-select-target-popup-storm", BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://node.example/one",
		Title:      "One",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, true)
	first := sessionRegistry.TrackTab("browser-runtime-select-target-popup-storm", BrowserSessionTarget{
		TabIndex:   3,
		URL:        "https://popup.example/offer",
		Title:      "Offer",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, false)
	sessionRegistry.RecordPendingTargetReviewForRoute("browser-runtime-select-target-popup-storm", BrowserSessionRoute{
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
		BrowserApp: "Chromium",
	}, agentxbrowserruntime.BrowserSessionTargetReview{
		ID:         first.ID,
		TabIndex:   first.TabIndex,
		URL:        first.URL,
		Title:      first.Title,
		BrowserApp: first.BrowserApp,
		Backend:    first.Backend,
		Profile:    first.Profile,
		Target:     first.Target,
		Decision:   "session_target_popup_review_required",
		Reason:     "pending popup review",
	})
	second := sessionRegistry.TrackTab("browser-runtime-select-target-popup-storm", BrowserSessionTarget{
		TabIndex:   4,
		URL:        "https://popup.example/bonus",
		Title:      "Bonus",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, false)
	sessionRegistry.RecordPendingTargetReviewForRoute("browser-runtime-select-target-popup-storm", BrowserSessionRoute{
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
		BrowserApp: "Chromium",
	}, agentxbrowserruntime.BrowserSessionTargetReview{
		ID:         second.ID,
		TabIndex:   second.TabIndex,
		URL:        second.URL,
		Title:      second.Title,
		BrowserApp: second.BrowserApp,
		Backend:    second.Backend,
		Profile:    second.Profile,
		Target:     second.Target,
		Decision:   "session_target_popup_review_required",
		Reason:     "pending popup review",
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: fmt.Sprintf(`{"action":"select_target","runtime_target":"node","profile":"workbench","target":"target:%s"}`, second.ID),
	})
	if err != nil {
		t.Fatalf("browser_runtime select_target popup storm review: %v", err)
	}
	var payload struct {
		Action                 string `json:"action"`
		Status                 string `json:"status"`
		SelectTargetDecision   string `json:"select_target_decision"`
		SelectTargetReady      bool   `json:"select_target_ready"`
		Note                   string `json:"note"`
		SessionTargetSelection struct {
			ID       string `json:"id"`
			TabIndex int    `json:"tab_index"`
			Source   string `json:"source"`
		} `json:"session_target_selection"`
		SessionBinding struct {
			CurrentTargetID          string `json:"current_target_id"`
			SelectedBrowserTargetID  string `json:"selected_browser_target_id"`
			PendingTargetReviewCount int    `json:"pending_target_review_count"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode select_target popup storm output: %v", err)
	}
	if payload.Action != "select_target" || payload.Status != "review_required" || payload.SelectTargetDecision != "session_target_popup_review_required" || payload.SelectTargetReady {
		t.Fatalf("unexpected select_target popup storm payload: %#v", payload)
	}
	if payload.SessionTargetSelection.ID != current.ID || payload.SessionTargetSelection.TabIndex != 1 || payload.SessionTargetSelection.Source != "tracked_active_tab" {
		t.Fatalf("expected popup storm review to preserve prior selection, got %#v", payload.SessionTargetSelection)
	}
	if payload.SessionBinding.CurrentTargetID != current.ID || payload.SessionBinding.SelectedBrowserTargetID == second.ID || payload.SessionBinding.PendingTargetReviewCount != 1 || !strings.Contains(payload.Note, "accumulated 2 pending popup targets") {
		t.Fatalf("expected popup storm review guidance, got %#v", payload)
	}
}

func TestRegisterBrowserTools_RuntimeSelectTargetConfirmsPendingPopupWithForce(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}},
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-select-target-popup-confirmed")
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"select_profile","runtime_target":"node","profile":"workbench"}`,
	}); err != nil {
		t.Fatalf("browser_runtime select_profile: %v", err)
	}
	sessionRegistry.TrackTab("browser-runtime-select-target-popup-confirmed", BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://node.example/one",
		Title:      "One",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, true)
	popup := sessionRegistry.TrackTab("browser-runtime-select-target-popup-confirmed", BrowserSessionTarget{
		TabIndex:   3,
		URL:        "https://popup.example/offer",
		Title:      "Offer",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, false)
	sessionRegistry.RecordPendingTargetReviewForRoute("browser-runtime-select-target-popup-confirmed", BrowserSessionRoute{
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
		BrowserApp: "Chromium",
	}, agentxbrowserruntime.BrowserSessionTargetReview{
		ID:         popup.ID,
		TabIndex:   popup.TabIndex,
		URL:        popup.URL,
		Title:      popup.Title,
		BrowserApp: popup.BrowserApp,
		Backend:    popup.Backend,
		Profile:    popup.Profile,
		Target:     popup.Target,
		Decision:   "session_target_popup_review_required",
		Reason:     "pending popup review",
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: fmt.Sprintf(`{"action":"select_target","runtime_target":"node","profile":"workbench","target":"target:%s","force":true}`, popup.ID),
	})
	if err != nil {
		t.Fatalf("browser_runtime select_target popup force: %v", err)
	}
	var payload struct {
		Action                 string `json:"action"`
		Status                 string `json:"status"`
		Force                  bool   `json:"force"`
		SelectTargetDecision   string `json:"select_target_decision"`
		SelectTargetReady      bool   `json:"select_target_ready"`
		Note                   string `json:"note"`
		SessionTargetSelection struct {
			ID       string `json:"id"`
			TabIndex int    `json:"tab_index"`
			Source   string `json:"source"`
		} `json:"session_target_selection"`
		SessionBinding struct {
			CurrentTargetID          string `json:"current_target_id"`
			SelectedBrowserTargetID  string `json:"selected_browser_target_id"`
			PendingTargetReviewCount int    `json:"pending_target_review_count"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode select_target popup force output: %v", err)
	}
	if payload.Action != "select_target" || payload.Status != "ok" || !payload.Force || payload.SelectTargetDecision != "session_target_popup_review_confirmed" || !payload.SelectTargetReady {
		t.Fatalf("unexpected forced select_target popup payload: %#v", payload)
	}
	if payload.SessionTargetSelection.ID != popup.ID || payload.SessionTargetSelection.TabIndex != 3 || payload.SessionTargetSelection.Source != "select_target" {
		t.Fatalf("expected forced popup selection to adopt target, got %#v", payload.SessionTargetSelection)
	}
	if payload.SessionBinding.CurrentTargetID != popup.ID || payload.SessionBinding.SelectedBrowserTargetID != popup.ID || payload.SessionBinding.PendingTargetReviewCount != 0 || !strings.Contains(payload.Note, "force=true") {
		t.Fatalf("expected forced popup selection to clear pending review and promote target, got %#v", payload)
	}
}

func TestRegisterBrowserTools_RuntimeClearProfilePreservesExplicitTargetSelection(t *testing.T) {
	reg := llmxtools.NewRegistry()
	hostBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "system",
				BrowserApp: "Safari",
				Profile:    "default",
				Status:     "running",
				Running:    true,
				Connected:  true,
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}}
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "workbench",
				Status:     "running",
				Running:    true,
				Connected:  true,
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              hostBackend,
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-clear-profile-preserves-target")
	first := sessionRegistry.TrackTab("browser-runtime-clear-profile-preserves-target", BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://node.example/one",
		Title:      "One",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, false)
	second := sessionRegistry.TrackTab("browser-runtime-clear-profile-preserves-target", BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://node.example/two",
		Title:      "Two",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, false)
	if strings.TrimSpace(first.ID) == "" || strings.TrimSpace(second.ID) == "" {
		t.Fatalf("expected tracked targets to have IDs, got %#v %#v", first, second)
	}

	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: fmt.Sprintf(`{"action":"select_target","runtime_target":"node","target":"target:%s"}`, second.ID),
	}); err != nil {
		t.Fatalf("browser_runtime select_target: %v", err)
	}

	clearOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"clear_profile"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime clear_profile: %v", err)
	}
	var clearPayload struct {
		SessionProfileSelection any `json:"session_profile_selection"`
		SessionTargetSelection  struct {
			ID       string `json:"id"`
			TabIndex int    `json:"tab_index"`
			Source   string `json:"source"`
		} `json:"session_target_selection"`
		SessionBinding struct {
			SelectedBrowserProfile      string `json:"selected_browser_profile"`
			SelectedBrowserTargetID     string `json:"selected_browser_target_id"`
			SelectedBrowserTarget       string `json:"selected_browser_target"`
			SelectedBrowserTargetSource string `json:"selected_browser_target_source"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(clearOut), &clearPayload); err != nil {
		t.Fatalf("decode clear_profile output: %v", err)
	}
	if clearPayload.SessionProfileSelection != nil {
		t.Fatalf("expected clear_profile to clear promoted session profile selection, got %#v", clearPayload.SessionProfileSelection)
	}
	if clearPayload.SessionTargetSelection.ID != second.ID || clearPayload.SessionTargetSelection.TabIndex != 2 || clearPayload.SessionTargetSelection.Source != "select_target" {
		t.Fatalf("expected explicit select_target selection to survive clear_profile, got %#v", clearPayload.SessionTargetSelection)
	}
	if clearPayload.SessionBinding.SelectedBrowserProfile != "" {
		t.Fatalf("expected clear_profile to clear selected browser profile from binding, got %#v", clearPayload.SessionBinding)
	}
	if clearPayload.SessionBinding.SelectedBrowserTargetID != second.ID || clearPayload.SessionBinding.SelectedBrowserTargetSource != "select_target" {
		t.Fatalf("expected clear_profile to preserve explicit selected browser target in binding, got %#v", clearPayload.SessionBinding)
	}
}

func TestRegisterBrowserTools_RuntimeSelectTargetPromotesSessionProfileWhenMissing(t *testing.T) {
	reg := llmxtools.NewRegistry()
	hostBackend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{
		extractResult: BrowserExtractResult{
			Backend:    "system-extract",
			BrowserApp: "Safari",
			Title:      "Host",
			Content:    "Host content",
			FinalURL:   "https://host.example/home",
		},
	}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}}
	nodeBackend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{
		extractResult: BrowserExtractResult{
			Backend:    "proxy-extract",
			BrowserApp: "Chromium",
			Title:      "Node",
			Content:    "Node content",
			FinalURL:   "https://node.example/two",
		},
	}, runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}}
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              hostBackend,
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime", "browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-select-target-promotes-profile")
	first := sessionRegistry.TrackTab("browser-runtime-select-target-promotes-profile", BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://node.example/one",
		Title:      "One",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, false)
	second := sessionRegistry.TrackTab("browser-runtime-select-target-promotes-profile", BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://node.example/two",
		Title:      "Two",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, false)
	if got, ok := sessionRegistry.ResolveTarget("browser-runtime-select-target-promotes-profile", first.ID); !ok || got.TabIndex != 1 {
		t.Fatalf("expected first tracked target to exist, got %#v ok=%v", got, ok)
	}

	selectOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: fmt.Sprintf(`{"action":"select_target","runtime_target":"node","target":"target:%s"}`, second.ID),
	})
	if err != nil {
		t.Fatalf("browser_runtime select_target promote profile: %v", err)
	}
	var selectPayload struct {
		Action                 string `json:"action"`
		Status                 string `json:"status"`
		SelectTargetDecision   string `json:"select_target_decision"`
		SelectTargetReady      bool   `json:"select_target_ready"`
		SessionTargetSelection struct {
			ID       string `json:"id"`
			TabIndex int    `json:"tab_index"`
			Source   string `json:"source"`
		} `json:"session_target_selection"`
		SessionProfileSelection struct {
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
			Source        string `json:"source"`
		} `json:"session_profile_selection"`
		SessionBinding struct {
			SelectedBrowserProfile       string `json:"selected_browser_profile"`
			SelectedBrowserProfileSource string `json:"selected_browser_profile_source"`
			SelectedBrowserTargetID      string `json:"selected_browser_target_id"`
			SelectedBrowserTargetSource  string `json:"selected_browser_target_source"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(selectOut), &selectPayload); err != nil {
		t.Fatalf("decode select_target promote profile output: %v", err)
	}
	if selectPayload.Action != "select_target" || selectPayload.Status != "ok" || selectPayload.SelectTargetDecision != "session_target_selected" || !selectPayload.SelectTargetReady {
		t.Fatalf("unexpected select_target promote profile payload: %#v", selectPayload)
	}
	if selectPayload.SessionTargetSelection.ID != second.ID || selectPayload.SessionTargetSelection.TabIndex != 2 || selectPayload.SessionTargetSelection.Source != "select_target" {
		t.Fatalf("unexpected promoted session target selection: %#v", selectPayload.SessionTargetSelection)
	}
	if selectPayload.SessionProfileSelection.Profile != "workbench" || selectPayload.SessionProfileSelection.RuntimeTarget != "node" || selectPayload.SessionProfileSelection.Source != "select_target" {
		t.Fatalf("expected select_target to promote matching session profile, got %#v", selectPayload.SessionProfileSelection)
	}
	if selectPayload.SessionBinding.SelectedBrowserProfile != "workbench" || selectPayload.SessionBinding.SelectedBrowserProfileSource != "select_target" {
		t.Fatalf("expected session binding selected profile to follow select_target, got %#v", selectPayload.SessionBinding)
	}
	if selectPayload.SessionBinding.SelectedBrowserTargetID != second.ID || selectPayload.SessionBinding.SelectedBrowserTargetSource != "select_target" {
		t.Fatalf("expected session binding selected target to follow select_target, got %#v", selectPayload.SessionBinding)
	}

	statusOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"status"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime status after select_target promote profile: %v", err)
	}
	var statusPayload struct {
		SelectedRoute struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		SessionProfileSelection struct {
			Profile string `json:"profile"`
			Source  string `json:"source"`
		} `json:"session_profile_selection"`
		RouteResolution struct {
			ProfileSource       string `json:"profile_source"`
			RuntimeTargetSource string `json:"runtime_target_source"`
			TargetSource        string `json:"target_source"`
		} `json:"route_resolution"`
	}
	if err := json.Unmarshal([]byte(statusOut), &statusPayload); err != nil {
		t.Fatalf("decode status output after select_target promote profile: %v", err)
	}
	if statusPayload.SelectedRoute.Backend != "proxy" || statusPayload.SelectedRoute.Profile != "workbench" || statusPayload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected status after select_target promote profile to resolve node workbench route, got %#v", statusPayload.SelectedRoute)
	}
	if statusPayload.SessionProfileSelection.Profile != "workbench" || statusPayload.SessionProfileSelection.Source != "select_target" {
		t.Fatalf("expected session profile selection to persist after select_target, got %#v", statusPayload.SessionProfileSelection)
	}
	if statusPayload.RouteResolution.ProfileSource != "select_target" || statusPayload.RouteResolution.RuntimeTargetSource != "select_target" || statusPayload.RouteResolution.TargetSource != "select_target" {
		t.Fatalf("unexpected route resolution after select_target promote profile: %#v", statusPayload.RouteResolution)
	}

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"extract","max_chars":20}`,
	})
	if err != nil {
		t.Fatalf("browser_act extract after select_target promote profile: %v", err)
	}
	if len(hostBackend.extractReqs) != 0 {
		t.Fatalf("expected host backend extract to stay unused, got %#v", hostBackend.extractReqs)
	}
	if len(nodeBackend.extractReqs) != 1 || nodeBackend.extractReqs[0].TabIndex != 2 || nodeBackend.extractReqs[0].WaitMs != browserTabTargetWaitMs {
		t.Fatalf("expected node backend extract to inherit promoted profile/target, got %#v", nodeBackend.extractReqs)
	}
	var payload struct {
		Kind     string `json:"kind"`
		Target   string `json:"target"`
		TargetID string `json:"target_id"`
		TabIndex int    `json:"tab_index"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_act extract after select_target promote profile: %v", err)
	}
	if payload.Kind != "extract" || payload.TargetID != second.ID || payload.TabIndex != 2 {
		t.Fatalf("unexpected browser_act payload after select_target promote profile: %#v", payload)
	}
}

func TestRegisterBrowserTools_RuntimeSelectTargetAlreadySelected(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	nodeBackend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-select-target-already")
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"select_profile","runtime_target":"node","profile":"workbench"}`,
	}); err != nil {
		t.Fatalf("browser_runtime select_profile: %v", err)
	}
	tracked := sessionRegistry.TrackTab("browser-runtime-select-target-already", BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://node.example/two",
		Title:      "Two",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, true)
	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: fmt.Sprintf(`{"action":"select_target","target":"target:%s"}`, tracked.ID),
	})
	if err != nil {
		t.Fatalf("browser_runtime select_target already: %v", err)
	}
	var payload struct {
		Action               string `json:"action"`
		Status               string `json:"status"`
		SelectTargetDecision string `json:"select_target_decision"`
		SelectTargetReady    bool   `json:"select_target_ready"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode select_target already output: %v", err)
	}
	if payload.Action != "select_target" || payload.Status != "ok" || payload.SelectTargetDecision != "session_target_already_selected" || !payload.SelectTargetReady {
		t.Fatalf("unexpected select_target already payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_RuntimeSelectTargetMissingTarget(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	nodeBackend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-select-target-missing")
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"select_profile","runtime_target":"node","profile":"workbench"}`,
	}); err != nil {
		t.Fatalf("browser_runtime select_profile: %v", err)
	}
	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"select_target","target":"target:missing-target"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime select_target missing: %v", err)
	}
	var payload struct {
		Action               string `json:"action"`
		Status               string `json:"status"`
		SelectTargetDecision string `json:"select_target_decision"`
		Note                 string `json:"note"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode select_target missing output: %v", err)
	}
	if payload.Action != "select_target" || payload.Status != "error" || payload.SelectTargetDecision != "session_target_missing" || !strings.Contains(payload.Note, `target "missing-target" is not available`) {
		t.Fatalf("unexpected select_target missing payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_RuntimeClearTargetStopsImplicitTargetReuse(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	nodeBackend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{
		extractResult: BrowserExtractResult{
			Backend:    "proxy-extract",
			BrowserApp: "Chromium",
			Title:      "Workbench",
			Content:    "Extracted content",
			FinalURL:   "https://node.example/workbench",
		},
	}, runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime", "browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-clear-target")
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"select_profile","runtime_target":"node","profile":"workbench"}`,
	}); err != nil {
		t.Fatalf("browser_runtime select_profile: %v", err)
	}
	tracked := sessionRegistry.TrackTab("browser-runtime-clear-target", BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://node.example/two",
		Title:      "Two",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, false)
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: fmt.Sprintf(`{"action":"select_target","target":"target:%s"}`, tracked.ID),
	}); err != nil {
		t.Fatalf("browser_runtime select_target: %v", err)
	}

	clearOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"clear_target"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime clear_target: %v", err)
	}
	var clearPayload struct {
		Action                  string `json:"action"`
		Status                  string `json:"status"`
		ClearTargetDecision     string `json:"clear_target_decision"`
		ClearTargetReady        bool   `json:"clear_target_ready"`
		SessionTargetSelection  any    `json:"session_target_selection"`
		SessionProfileSelection struct {
			Profile string `json:"profile"`
			Source  string `json:"source"`
		} `json:"session_profile_selection"`
	}
	if err := json.Unmarshal([]byte(clearOut), &clearPayload); err != nil {
		t.Fatalf("decode clear_target output: %v", err)
	}
	if clearPayload.Action != "clear_target" || clearPayload.Status != "ok" || clearPayload.ClearTargetDecision != "session_target_cleared" || !clearPayload.ClearTargetReady {
		t.Fatalf("unexpected clear_target payload: %#v", clearPayload)
	}
	if clearPayload.SessionTargetSelection != nil {
		t.Fatalf("expected session target selection to clear, got %#v", clearPayload.SessionTargetSelection)
	}
	if clearPayload.SessionProfileSelection.Profile != "workbench" || clearPayload.SessionProfileSelection.Source != "select_profile" {
		t.Fatalf("expected explicit session profile selection to survive clear_target, got %#v", clearPayload.SessionProfileSelection)
	}

	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"extract","max_chars":20}`,
	}); err != nil {
		t.Fatalf("browser_act extract after clear_target: %v", err)
	}
	if len(nodeBackend.extractReqs) != 1 || nodeBackend.extractReqs[0].TabIndex != 0 {
		t.Fatalf("expected node backend extract to stop reusing selected target, got %#v", nodeBackend.extractReqs)
	}
}

func TestRegisterBrowserTools_RuntimeClearTargetInvalidatesCachedStatusWatch(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	hostBackend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}}
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "workbench",
				Status:     "running",
				Running:    true,
				Connected:  true,
			},
			runtimeProfilesResult: BrowserProfilesResult{
				Backend:        "proxy",
				DefaultProfile: "isolated",
				Profiles: []BrowserProfileInfo{
					{Profile: "workbench", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
				},
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              hostBackend,
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-clear-target-invalidates-status-watch")
	tracked := sessionRegistry.TrackTab("browser-runtime-clear-target-invalidates-status-watch", BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://node.example/two",
		Title:      "Two",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, false)
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: fmt.Sprintf(`{"action":"select_target","runtime_target":"node","profile":"workbench","target":"target:%s"}`, tracked.ID),
	}); err != nil {
		t.Fatalf("browser_runtime select_target before clear_target cache: %v", err)
	}
	statusArgs := `{"action":"status","runtime_target":"node","profile":"workbench"}`
	statusOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: statusArgs,
	})
	if err != nil {
		t.Fatalf("browser_runtime status before clear_target cache: %v", err)
	}
	var initial struct {
		SessionBinding struct {
			SelectedBrowserTargetID string `json:"selected_browser_target_id"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(statusOut), &initial); err != nil {
		t.Fatalf("decode initial status before clear_target: %v", err)
	}
	if initial.SessionBinding.SelectedBrowserTargetID != tracked.ID {
		t.Fatalf("expected initial cached status to include selected target, got %#v", initial.SessionBinding)
	}

	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"clear_target","runtime_target":"node","profile":"workbench"}`,
	}); err != nil {
		t.Fatalf("browser_runtime clear_target after cached status: %v", err)
	}

	statusOut, err = reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: statusArgs,
	})
	if err != nil {
		t.Fatalf("browser_runtime status after clear_target cache invalidation: %v", err)
	}
	var updated struct {
		SessionBinding struct {
			SelectedBrowserTargetID     string `json:"selected_browser_target_id"`
			SelectedBrowserTargetSource string `json:"selected_browser_target_source"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(statusOut), &updated); err != nil {
		t.Fatalf("decode updated status after clear_target: %v", err)
	}
	if updated.SessionBinding.SelectedBrowserTargetID != "" || updated.SessionBinding.SelectedBrowserTargetSource != "" {
		t.Fatalf("expected clear_target to invalidate cached status watch, got %#v", updated.SessionBinding)
	}
}

func TestRegisterBrowserTools_RuntimeClearTargetBlockedByActiveNodeRun(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionRunRegistry := newTestSessionRunRegistry()
	sessionRunRegistry.Record("browser-runtime-clear-target-blocked", agentxbrowserruntime.SharedSessionRunInfo{
		RunID:    "run-83",
		NodeID:   "node-alpha",
		Status:   "running",
		Provider: "gateway",
		Action:   "run_status",
	})
	nodeBackend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		SessionRunRegistry:   sessionRunRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-clear-target-blocked")
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"select_profile","runtime_target":"node","profile":"workbench"}`,
	}); err != nil {
		t.Fatalf("browser_runtime select_profile: %v", err)
	}
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-clear-target-blocked", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "workbench",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "running",
		Running:       true,
		Connected:     true,
	})
	tracked := sessionRegistry.TrackTab("browser-runtime-clear-target-blocked", BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://node.example/two",
		Title:      "Two",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, false)
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: fmt.Sprintf(`{"action":"select_target","runtime_target":"node","target":"target:%s"}`, tracked.ID),
	}); err != nil {
		t.Fatalf("browser_runtime select_target: %v", err)
	}

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"clear_target","runtime_target":"node"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime clear_target blocked: %v", err)
	}
	var payload struct {
		Action                 string `json:"action"`
		Status                 string `json:"status"`
		ClearTargetDecision    string `json:"clear_target_decision"`
		ClearTargetReady       bool   `json:"clear_target_ready"`
		Force                  bool   `json:"force"`
		SessionTargetSelection struct {
			ID     string `json:"id"`
			Source string `json:"source"`
		} `json:"session_target_selection"`
		SessionBinding struct {
			SelectedBrowserTargetID     string `json:"selected_browser_target_id"`
			SelectedBrowserTargetSource string `json:"selected_browser_target_source"`
			ActiveNodeRunID             string `json:"active_node_run_id"`
		} `json:"session_binding"`
		ProfileStatus struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
			Status        string `json:"status"`
			Running       bool   `json:"running"`
			Connected     bool   `json:"connected"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode clear_target blocked output: %v", err)
	}
	if payload.Action != "clear_target" || payload.Status != "ok" || payload.ClearTargetDecision != "clear_target_blocked_active_node_run" || payload.ClearTargetReady || payload.Force {
		t.Fatalf("unexpected blocked clear_target payload: %#v", payload)
	}
	if payload.SessionTargetSelection.ID != tracked.ID || payload.SessionTargetSelection.Source != "select_target" {
		t.Fatalf("expected blocked clear_target to preserve selected target, got %#v", payload.SessionTargetSelection)
	}
	if payload.SessionBinding.SelectedBrowserTargetID != tracked.ID || payload.SessionBinding.SelectedBrowserTargetSource != "select_target" || payload.SessionBinding.ActiveNodeRunID != "run-83" {
		t.Fatalf("expected blocked clear_target to preserve session binding, got %#v", payload.SessionBinding)
	}
	if payload.ProfileStatus.Backend != "proxy" || payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.RuntimeTarget != "node" || payload.ProfileStatus.Status != "running" || !payload.ProfileStatus.Running || !payload.ProfileStatus.Connected {
		t.Fatalf("expected blocked clear_target to preserve effective lifecycle state, got %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_RuntimeClearTargetForceBypassesActiveNodeRun(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionRunRegistry := newTestSessionRunRegistry()
	sessionRunRegistry.Record("browser-runtime-clear-target-force", agentxbrowserruntime.SharedSessionRunInfo{
		RunID:    "run-84",
		NodeID:   "node-alpha",
		Status:   "running",
		Provider: "gateway",
		Action:   "run_status",
	})
	nodeBackend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		SessionRunRegistry:   sessionRunRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-clear-target-force")
	tracked := sessionRegistry.TrackTab("browser-runtime-clear-target-force", BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://node.example/two",
		Title:      "Two",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, false)
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: fmt.Sprintf(`{"action":"select_target","runtime_target":"node","target":"target:%s"}`, tracked.ID),
	}); err != nil {
		t.Fatalf("browser_runtime select_target: %v", err)
	}

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"clear_target","runtime_target":"node","force":true}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime clear_target force: %v", err)
	}
	var payload struct {
		Action                 string `json:"action"`
		Status                 string `json:"status"`
		ClearTargetDecision    string `json:"clear_target_decision"`
		ClearTargetReady       bool   `json:"clear_target_ready"`
		Force                  bool   `json:"force"`
		SessionTargetSelection any    `json:"session_target_selection"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode clear_target force output: %v", err)
	}
	if payload.Action != "clear_target" || payload.Status != "ok" || payload.ClearTargetDecision != "session_target_cleared" || !payload.ClearTargetReady || !payload.Force {
		t.Fatalf("unexpected forced clear_target payload: %#v", payload)
	}
	if payload.SessionTargetSelection != nil {
		t.Fatalf("expected forced clear_target to clear selection, got %#v", payload.SessionTargetSelection)
	}
}

func TestRegisterBrowserTools_RuntimeClearTargetClearsAutoPromotedProfile(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	nodeBackend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-clear-target-clears-profile")
	tracked := sessionRegistry.TrackTab("browser-runtime-clear-target-clears-profile", BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://node.example/two",
		Title:      "Two",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, false)
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: fmt.Sprintf(`{"action":"select_target","runtime_target":"node","target":"target:%s"}`, tracked.ID),
	}); err != nil {
		t.Fatalf("browser_runtime select_target: %v", err)
	}

	clearOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"clear_target","runtime_target":"node"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime clear_target auto profile: %v", err)
	}
	var clearPayload struct {
		Action                  string `json:"action"`
		Status                  string `json:"status"`
		ClearTargetDecision     string `json:"clear_target_decision"`
		ClearTargetReady        bool   `json:"clear_target_ready"`
		SessionTargetSelection  any    `json:"session_target_selection"`
		SessionProfileSelection any    `json:"session_profile_selection"`
		SessionBinding          struct {
			SelectedBrowserProfile       string `json:"selected_browser_profile"`
			SelectedBrowserProfileSource string `json:"selected_browser_profile_source"`
			SelectedBrowserTargetID      string `json:"selected_browser_target_id"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(clearOut), &clearPayload); err != nil {
		t.Fatalf("decode clear_target auto profile output: %v", err)
	}
	if clearPayload.Action != "clear_target" || clearPayload.Status != "ok" || clearPayload.ClearTargetDecision != "session_target_cleared" || !clearPayload.ClearTargetReady {
		t.Fatalf("unexpected clear_target auto profile payload: %#v", clearPayload)
	}
	if clearPayload.SessionTargetSelection != nil {
		t.Fatalf("expected session target selection to clear, got %#v", clearPayload.SessionTargetSelection)
	}
	if clearPayload.SessionProfileSelection != nil {
		t.Fatalf("expected auto-promoted session profile selection to clear with target, got %#v", clearPayload.SessionProfileSelection)
	}
	if clearPayload.SessionBinding.SelectedBrowserProfile != "" || clearPayload.SessionBinding.SelectedBrowserProfileSource != "" || clearPayload.SessionBinding.SelectedBrowserTargetID != "" {
		t.Fatalf("expected session binding to clear auto-promoted browser defaults, got %#v", clearPayload.SessionBinding)
	}
}

func TestRegisterBrowserTools_RuntimeClearTargetClearsRememberPromotedProfile(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	hostBackend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}}
	nodeBackend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{
		tabsResult: BrowserTabsResult{
			Backend:     "proxy-tabs",
			BrowserApp:  "Chromium",
			Action:      "focus",
			Status:      "focused",
			ActiveIndex: 3,
			Tabs: []BrowserTab{
				{Index: 2, Title: "Docs", URL: "https://node.example/docs", Active: false},
				{Index: 3, Title: "Workbench", URL: "https://node.example/workbench", Active: true},
			},
		},
	}, runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              hostBackend,
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_act", "browser_runtime"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-clear-target-clears-remember-profile")
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"focus_tab","runtime_target":"node","profile":"workbench","tab_index":3,"remember_target":true}`,
	}); err != nil {
		t.Fatalf("browser_act focus_tab remember_target node: %v", err)
	}

	clearOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"clear_target","runtime_target":"node"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime clear_target remembered profile: %v", err)
	}
	var clearPayload struct {
		Action                  string `json:"action"`
		Status                  string `json:"status"`
		ClearTargetDecision     string `json:"clear_target_decision"`
		ClearTargetReady        bool   `json:"clear_target_ready"`
		SessionTargetSelection  any    `json:"session_target_selection"`
		SessionProfileSelection any    `json:"session_profile_selection"`
		SessionBinding          struct {
			SelectedBrowserProfile       string `json:"selected_browser_profile"`
			SelectedBrowserProfileSource string `json:"selected_browser_profile_source"`
			SelectedBrowserTargetID      string `json:"selected_browser_target_id"`
			SelectedBrowserTargetSource  string `json:"selected_browser_target_source"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(clearOut), &clearPayload); err != nil {
		t.Fatalf("decode clear_target remembered profile output: %v", err)
	}
	if clearPayload.Action != "clear_target" || clearPayload.Status != "ok" || clearPayload.ClearTargetDecision != "session_target_cleared" || !clearPayload.ClearTargetReady {
		t.Fatalf("unexpected clear_target remembered profile payload: %#v", clearPayload)
	}
	if clearPayload.SessionTargetSelection != nil {
		t.Fatalf("expected remembered session target selection to clear, got %#v", clearPayload.SessionTargetSelection)
	}
	if clearPayload.SessionProfileSelection != nil {
		t.Fatalf("expected remember_target-promoted session profile selection to clear with target, got %#v", clearPayload.SessionProfileSelection)
	}
	if clearPayload.SessionBinding.SelectedBrowserProfile != "" || clearPayload.SessionBinding.SelectedBrowserProfileSource != "" || clearPayload.SessionBinding.SelectedBrowserTargetID != "" || clearPayload.SessionBinding.SelectedBrowserTargetSource != "" {
		t.Fatalf("expected session binding to clear remember_target-promoted browser defaults, got %#v", clearPayload.SessionBinding)
	}
}

func TestRegisterBrowserTools_RuntimeClearTargetClearsRememberProfilePromotedProfile(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	hostBackend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}}
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "workbench",
				Status:     "stopped",
			},
			runtimeStartResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "workbench",
				Status:     "started",
				Running:    true,
				Connected:  true,
			},
			tabsResult: BrowserTabsResult{
				Backend:     "proxy-tabs",
				BrowserApp:  "Chromium",
				ActiveIndex: 1,
				Tabs: []BrowserTab{
					{Index: 1, Title: "Workbench", URL: "https://node.example/workbench", Active: true},
				},
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              hostBackend,
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime", "browser_tabs"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-clear-target-clears-remember-profile-source")
	sessionRegistry.TrackTab("browser-runtime-clear-target-clears-remember-profile-source", BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, true)
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"prepare","runtime_target":"node","profile":"workbench","remember_profile":true}`,
	}); err != nil {
		t.Fatalf("browser_runtime prepare remember_profile: %v", err)
	}

	clearOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"clear_target","runtime_target":"node"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime clear_target remember_profile source: %v", err)
	}
	var clearPayload struct {
		Action                  string `json:"action"`
		Status                  string `json:"status"`
		ClearTargetDecision     string `json:"clear_target_decision"`
		ClearTargetReady        bool   `json:"clear_target_ready"`
		SessionTargetSelection  any    `json:"session_target_selection"`
		SessionProfileSelection any    `json:"session_profile_selection"`
		SessionBinding          struct {
			SelectedBrowserProfile       string `json:"selected_browser_profile"`
			SelectedBrowserProfileSource string `json:"selected_browser_profile_source"`
			SelectedBrowserTargetID      string `json:"selected_browser_target_id"`
			SelectedBrowserTargetSource  string `json:"selected_browser_target_source"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(clearOut), &clearPayload); err != nil {
		t.Fatalf("decode clear_target remember_profile output: %v", err)
	}
	if clearPayload.Action != "clear_target" || clearPayload.Status != "ok" || clearPayload.ClearTargetDecision != "session_target_cleared" || !clearPayload.ClearTargetReady {
		t.Fatalf("unexpected clear_target remember_profile payload: %#v", clearPayload)
	}
	if clearPayload.SessionTargetSelection != nil {
		t.Fatalf("expected remembered target selection to clear, got %#v", clearPayload.SessionTargetSelection)
	}
	if clearPayload.SessionProfileSelection != nil {
		t.Fatalf("expected remember_profile-promoted session profile selection to clear with target, got %#v", clearPayload.SessionProfileSelection)
	}
	if clearPayload.SessionBinding.SelectedBrowserProfile != "" || clearPayload.SessionBinding.SelectedBrowserProfileSource != "" || clearPayload.SessionBinding.SelectedBrowserTargetID != "" || clearPayload.SessionBinding.SelectedBrowserTargetSource != "" {
		t.Fatalf("expected session binding to clear remember_profile-promoted defaults, got %#v", clearPayload.SessionBinding)
	}
}
