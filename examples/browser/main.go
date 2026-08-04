// browser 展示显式注入 backend 的最小 Browser Tool 路径。
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	browsertools "github.com/wsnacj/agentx-go/browser/tools"
	toolcontract "github.com/wsnacj/agentx-go/components/tool"
	agentxtools "github.com/wsnacj/agentx-go/tools"
)

type browserResult struct {
	Status  string `json:"status"`
	Backend string `json:"backend"`
	Note    string `json:"note"`
}

func run(ctx context.Context) (browserResult, error) {
	registry := agentxtools.NewRegistry()
	browsertools.RegisterBrowserTools(registry, browsertools.BrowserToolOptions{
		Backend:      memoryBrowser{},
		EnabledTools: []string{"browser"},
	})
	raw, err := registry.Execute(ctx, toolcontract.Call{
		Name:      "browser",
		Arguments: `{"action":"open","url":"https://93.184.216.34/agentx"}`,
	})
	if err != nil {
		return browserResult{}, err
	}
	var result browserResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return browserResult{}, err
	}
	return result, nil
}

func main() {
	result, err := run(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	encoded, _ := json.Marshal(result)
	fmt.Println(string(encoded))
}

// memoryBrowser 是无网络 fixture。真实项目在同一 seam 注入 Playwright、
// browserd 或其它经过授权的 backend。
type memoryBrowser struct{}

func (memoryBrowser) BrowserCapabilities() browsertools.BrowserCapabilities {
	return browsertools.BrowserCapabilities{Open: true}
}

func (memoryBrowser) Open(ctx context.Context, request browsertools.BrowserOpenRequest) (browsertools.BrowserOpenResult, error) {
	if err := ctx.Err(); err != nil {
		return browsertools.BrowserOpenResult{}, err
	}
	return browsertools.BrowserOpenResult{
		Status: "opened", Backend: "memory", BrowserApp: "fixture", Note: request.URL,
	}, nil
}

func (memoryBrowser) Navigate(context.Context, browsertools.BrowserNavigateRequest) (browsertools.BrowserNavigateResult, error) {
	return browsertools.BrowserNavigateResult{}, nil
}
func (memoryBrowser) Tabs(context.Context, browsertools.BrowserTabsRequest) (browsertools.BrowserTabsResult, error) {
	return browsertools.BrowserTabsResult{}, nil
}
func (memoryBrowser) Extract(context.Context, browsertools.BrowserExtractRequest) (browsertools.BrowserExtractResult, error) {
	return browsertools.BrowserExtractResult{}, nil
}
func (memoryBrowser) Snapshot(context.Context, browsertools.BrowserSnapshotRequest) (browsertools.BrowserSnapshotResult, error) {
	return browsertools.BrowserSnapshotResult{}, nil
}
func (memoryBrowser) Screenshot(context.Context, browsertools.BrowserScreenshotRequest) (browsertools.BrowserScreenshotResult, error) {
	return browsertools.BrowserScreenshotResult{}, nil
}
func (memoryBrowser) Click(context.Context, browsertools.BrowserClickRequest) (browsertools.BrowserClickResult, error) {
	return browsertools.BrowserClickResult{}, nil
}
func (memoryBrowser) Type(context.Context, browsertools.BrowserTypeRequest) (browsertools.BrowserTypeResult, error) {
	return browsertools.BrowserTypeResult{}, nil
}
func (memoryBrowser) Eval(context.Context, browsertools.BrowserEvalRequest) (browsertools.BrowserEvalResult, error) {
	return browsertools.BrowserEvalResult{}, nil
}
