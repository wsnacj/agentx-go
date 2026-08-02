package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/wsnacj/agentx-go/components/llm"
	browserops "github.com/wsnacj/agentx-go/scenes/browserops"
	"github.com/wsnacj/agentx-go/scenes/browserops/hostkit"
	"github.com/wsnacj/agentx-go/tools"
)

type fixtureExecutor struct {
	call llm.FunctionCall
}

func (e *fixtureExecutor) Execute(_ context.Context, call llm.FunctionCall) (string, error) {
	e.call = call
	return `{"status":"opened","final_url":"https://example.com/form"}`, nil
}

func run() (string, error) {
	executor := &fixtureExecutor{}
	registry := tools.NewRegistry()
	hostkit.RegisterTools(registry, hostkit.BuildStandardToolHandlers(hostkit.Config{Executor: executor}))
	out, err := registry.Execute(context.Background(), llm.FunctionCall{
		Name:      hostkit.ToolBrowserOpenTarget,
		Arguments: `{"target_url":"https://example.com/form","runtime_target":"fixture"}`,
	})
	if err != nil {
		return "", err
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		return "", err
	}
	definition := browserops.Definition()
	evaluation := browserops.EvaluateBrowserEvidenceBundle(
		browserops.BuildBrowserEvidenceBundleFromState(map[string]any{
			"case": map[string]any{"input": map[string]any{"target_url": "https://example.com/form"}},
			"review": map[string]any{
				"snapshot":      "heading Submitted status Success",
				"evidence_path": "artifacts/final.png",
				"final_url":     payload["final_url"],
			},
		}),
		browserops.BrowserEvidenceRequirements{
			RequireSnapshot: true, RequireScreenshot: true, RequireFinalURL: true,
		},
	)
	if executor.call.Name != hostkit.RuntimeToolBrowserAct || !evaluation.Passed {
		return "", fmt.Errorf("browserops fixture mismatch: call=%q evaluation=%#v", executor.call.Name, evaluation)
	}
	return fmt.Sprintf("agentx-browserops-ok:%s:%s:%s:%t",
		definition.Manifest.ID, definition.Manifest.DefaultWorkflow, executor.call.Name, evaluation.Passed), nil
}

func main() {
	result, err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(result)
}
