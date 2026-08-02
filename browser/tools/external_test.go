package tools_test

import (
	"context"
	"errors"
	"testing"

	browserruntime "github.com/wsnacj/agentx-go/browser/runtime"
	browsertools "github.com/wsnacj/agentx-go/browser/tools"
	llm "github.com/wsnacj/agentx-go/components/llm"
	agentxtools "github.com/wsnacj/agentx-go/tools"
)

type externalBackend struct{}

func (externalBackend) Open(ctx context.Context, req browserruntime.BrowserOpenRequest) (browserruntime.BrowserOpenResult, error) {
	if err := ctx.Err(); err != nil {
		return browserruntime.BrowserOpenResult{}, err
	}
	return browserruntime.BrowserOpenResult{Backend: "fixture", BrowserApp: req.BrowserApp, Status: "opened"}, nil
}

func (externalBackend) Navigate(context.Context, browserruntime.BrowserNavigateRequest) (browserruntime.BrowserNavigateResult, error) {
	return browserruntime.BrowserNavigateResult{}, nil
}

func (externalBackend) Tabs(context.Context, browserruntime.BrowserTabsRequest) (browserruntime.BrowserTabsResult, error) {
	return browserruntime.BrowserTabsResult{}, nil
}

func (externalBackend) Extract(context.Context, browserruntime.BrowserExtractRequest) (browserruntime.BrowserExtractResult, error) {
	return browserruntime.BrowserExtractResult{}, nil
}

func (externalBackend) Snapshot(context.Context, browserruntime.BrowserSnapshotRequest) (browserruntime.BrowserSnapshotResult, error) {
	return browserruntime.BrowserSnapshotResult{}, nil
}

func (externalBackend) Screenshot(context.Context, browserruntime.BrowserScreenshotRequest) (browserruntime.BrowserScreenshotResult, error) {
	return browserruntime.BrowserScreenshotResult{}, nil
}

func (externalBackend) Click(context.Context, browserruntime.BrowserClickRequest) (browserruntime.BrowserClickResult, error) {
	return browserruntime.BrowserClickResult{}, nil
}

func (externalBackend) Type(context.Context, browserruntime.BrowserTypeRequest) (browserruntime.BrowserTypeResult, error) {
	return browserruntime.BrowserTypeResult{}, nil
}

func (externalBackend) Eval(context.Context, browserruntime.BrowserEvalRequest) (browserruntime.BrowserEvalResult, error) {
	return browserruntime.BrowserEvalResult{}, nil
}

func TestExternalConsumerRegistersAndExecutesUnifiedBrowser(t *testing.T) {
	registry := agentxtools.NewRegistry()
	browsertools.RegisterBrowserTools(registry, browsertools.BrowserToolOptions{
		Backend:      externalBackend{},
		EnabledTools: []string{"browser"},
	})

	out, err := registry.Execute(context.Background(), llm.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"open","url":"https://93.184.216.34/","browser_app":"Chromium"}`,
	})
	if err != nil {
		t.Fatalf("execute browser open: %v", err)
	}
	if out == "" {
		t.Fatal("expected browser result")
	}
}

func TestExternalConsumerCancellationAndTypedArgumentError(t *testing.T) {
	registry := agentxtools.NewRegistry()
	browsertools.RegisterBrowserTools(registry, browsertools.BrowserToolOptions{
		Backend:      externalBackend{},
		EnabledTools: []string{"browser"},
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := registry.Execute(ctx, llm.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"open","url":"https://93.184.216.34/"}`,
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation error = %v", err)
	}

	_, err = registry.Execute(context.Background(), llm.FunctionCall{Name: "browser", Arguments: "{"})
	var argumentError *browsertools.ToolArgumentError
	if !errors.As(err, &argumentError) {
		t.Fatalf("argument error = %T %v", err, err)
	}
}
