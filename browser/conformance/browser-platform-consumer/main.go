package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	browserd "github.com/wsnacj/agentx-go/browser/host/browserd"
	browserruntime "github.com/wsnacj/agentx-go/browser/runtime"
	browsertools "github.com/wsnacj/agentx-go/browser/tools"
	toolcontract "github.com/wsnacj/agentx-go/components/tool"
	"github.com/wsnacj/agentx-go/tools"
)

type fakeBackend struct{}

func (*fakeBackend) BrowserCapabilities() browsertools.BrowserCapabilities {
	return browsertools.BrowserCapabilities{Open: true}
}

func (*fakeBackend) Open(ctx context.Context, req browsertools.BrowserOpenRequest) (browsertools.BrowserOpenResult, error) {
	if err := ctx.Err(); err != nil {
		return browsertools.BrowserOpenResult{}, err
	}
	return browsertools.BrowserOpenResult{Backend: "memory", BrowserApp: "fixture", Status: "opened", Note: req.URL}, nil
}
func (*fakeBackend) Navigate(context.Context, browsertools.BrowserNavigateRequest) (browsertools.BrowserNavigateResult, error) {
	return browsertools.BrowserNavigateResult{}, nil
}
func (*fakeBackend) Tabs(context.Context, browsertools.BrowserTabsRequest) (browsertools.BrowserTabsResult, error) {
	return browsertools.BrowserTabsResult{}, nil
}
func (*fakeBackend) Extract(context.Context, browsertools.BrowserExtractRequest) (browsertools.BrowserExtractResult, error) {
	return browsertools.BrowserExtractResult{}, nil
}
func (*fakeBackend) Snapshot(context.Context, browsertools.BrowserSnapshotRequest) (browsertools.BrowserSnapshotResult, error) {
	return browsertools.BrowserSnapshotResult{}, nil
}
func (*fakeBackend) Screenshot(context.Context, browsertools.BrowserScreenshotRequest) (browsertools.BrowserScreenshotResult, error) {
	return browsertools.BrowserScreenshotResult{}, nil
}
func (*fakeBackend) Click(context.Context, browsertools.BrowserClickRequest) (browsertools.BrowserClickResult, error) {
	return browsertools.BrowserClickResult{}, nil
}
func (*fakeBackend) Type(context.Context, browsertools.BrowserTypeRequest) (browsertools.BrowserTypeResult, error) {
	return browsertools.BrowserTypeResult{}, nil
}
func (*fakeBackend) Eval(context.Context, browsertools.BrowserEvalRequest) (browsertools.BrowserEvalResult, error) {
	return browsertools.BrowserEvalResult{}, nil
}

type result struct {
	ToolStatus       string `json:"tool_status"`
	ToolBackend      string `json:"tool_backend"`
	RepairApplied    bool   `json:"repair_applied"`
	BrowserdStatus   string `json:"browserd_status"`
	NoManagedProcess bool   `json:"no_managed_process"`
}

func run(ctx context.Context) (result, error) {
	registry := tools.NewRegistry()
	browsertools.RegisterBrowserTools(registry, browsertools.BrowserToolOptions{
		Backend:      &fakeBackend{},
		EnabledTools: []string{"browser"},
	})
	raw, err := registry.Execute(ctx, toolcontract.Call{
		Name: "browser", Arguments: `{"action":"open","url":"https://93.184.216.34"}`,
	})
	if err != nil {
		return result{}, err
	}
	var payload struct {
		Status  string `json:"status"`
		Backend string `json:"backend"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return result{}, err
	}
	_, _, repaired, err := browsertools.AttemptConstrainedBrowserArgumentRepair(
		browsertools.BrowserCompatToolNameForActKind("click"), `{"element":"button.buy"}`,
		&browsertools.ToolArgumentError{
			Code: "missing_locator", Repairable: true, SafeAutorepair: true,
			AllowedRepairs: []browsertools.ToolArgumentRepair{{Kind: "use_alias_field", From: "element", To: "selector_or_ref"}},
		},
	)
	if err != nil {
		return result{}, err
	}
	manager, err := browserd.NewManager(browserd.ManagerOptions{
		Plan: browserd.Plan{Enabled: true, Command: "fixture-browserd", Endpoint: "memory://browserd"},
		Probe: func(ctx context.Context, _ browserd.Plan, _ int) (browserruntime.BrowserProfileStatusResult, error) {
			if err := ctx.Err(); err != nil {
				return browserruntime.BrowserProfileStatusResult{}, err
			}
			return browserruntime.BrowserProfileStatusResult{Status: "ready", Backend: "memory"}, nil
		},
	})
	if err != nil {
		return result{}, err
	}
	defer manager.Close()
	status, err := manager.Probe(ctx)
	if err != nil {
		return result{}, err
	}
	return result{
		ToolStatus: payload.Status, ToolBackend: payload.Backend, RepairApplied: repaired,
		BrowserdStatus: status.Status, NoManagedProcess: manager.ManagedProcessID() == 0,
	}, nil
}

func main() {
	value, err := run(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := json.NewEncoder(os.Stdout).Encode(value); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
