package browserops_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/wsnacj/agentx-go/components/llm"
	"github.com/wsnacj/agentx-go/extensions/pack"
	"github.com/wsnacj/agentx-go/runtime/workflow"
	browserops "github.com/wsnacj/agentx-go/scenes/browserops"
	"github.com/wsnacj/agentx-go/scenes/browserops/hostkit"
	"github.com/wsnacj/agentx-go/tools"
)

type externalValidator struct{}

func (externalValidator) ValidateSpec(workflow.Spec) error { return nil }

type externalLowerer struct{}

func (externalLowerer) LowerToolArguments(node workflow.NodeSpec) (string, error) {
	arguments, _ := node.Config["args"].(map[string]any)
	encoded, err := json.Marshal(arguments)
	return string(encoded), err
}

type recordingExecutor struct {
	call llm.FunctionCall
}

func (e *recordingExecutor) Execute(_ context.Context, call llm.FunctionCall) (string, error) {
	e.call = call
	return `{"status":"opened","final_url":"https://example.com/form"}`, nil
}

func TestPortableBrowserOpsContract(t *testing.T) {
	coordinator, err := pack.NewCoordinator(externalValidator{}, externalLowerer{})
	if err != nil {
		t.Fatalf("NewCoordinator(): %v", err)
	}
	registry, err := pack.NewMemoryRegistry(coordinator)
	if err != nil {
		t.Fatalf("NewMemoryRegistry(): %v", err)
	}
	if err := browserops.RegisterInto(registry); err != nil {
		t.Fatalf("RegisterInto(): %v", err)
	}
	spec, err := browserops.MaterializedDefaultWorkflow(coordinator)
	if err != nil || spec.ID != browserops.DefaultWorkflow || spec.Pack != browserops.PackID {
		t.Fatalf("MaterializedDefaultWorkflow() = %#v, %v", spec, err)
	}

	exec := &recordingExecutor{}
	toolRegistry := tools.NewRegistry()
	decoderCalled := false
	hostkit.RegisterToolsWithDecoder(
		toolRegistry,
		hostkit.BuildStandardToolHandlers(hostkit.Config{Executor: exec}),
		func(raw string) (map[string]any, error) {
			decoderCalled = true
			return map[string]any{"target_url": "https://example.com/form"}, nil
		},
	)
	result, err := toolRegistry.Execute(context.Background(), llm.FunctionCall{
		Name: hostkit.ToolBrowserOpenTarget, Arguments: "host-compatible-payload",
	})
	if err != nil || result == "" || !decoderCalled {
		t.Fatalf("browser open result=%q decoderCalled=%v err=%v", result, decoderCalled, err)
	}
	if exec.call.Name != hostkit.RuntimeToolBrowserAct {
		t.Fatalf("runtime call = %#v", exec.call)
	}

	evaluation := browserops.EvaluateBrowserVisualEvidenceGate(browserops.BrowserVisualEvidenceEvaluationInput{
		SnapshotText: "heading Submitted status Success", ScreenshotPath: "artifacts/final.png",
		FinalURL: "https://example.com/form/submitted", TargetURL: "https://example.com/form",
		RequireSnapshot: true, RequireScreenshot: true, RequireFinalURL: true,
	})
	if !evaluation.Passed || !evaluation.VisualEvidenceReady {
		t.Fatalf("evaluation = %#v", evaluation)
	}
}
