package main

import (
	"context"
	"fmt"

	agentx "github.com/wsnacj/agentx-go"
	llm "github.com/wsnacj/agentx-go/components/llm"
	astock "github.com/wsnacj/agentx-go/extensions/astock"
	astockcontracts "github.com/wsnacj/agentx-go/extensions/astock/contracts"
	"github.com/wsnacj/agentx-go/runtime/execution"
	"github.com/wsnacj/agentx-go/runtime/hostkit"
	"github.com/wsnacj/agentx-go/runtime/toolloop"
	workflow "github.com/wsnacj/agentx-go/runtime/workflow"
	workflowhostkit "github.com/wsnacj/agentx-go/runtime/workflow/hostkit"
)

func buildModelToolRound(context.Context, execution.Request) (hostkit.ModelToolRoundConfig, error) {
	return hostkit.ModelToolRoundConfig{
		RequestModel: func(_ context.Context, input toolloop.RoundExecutionInput) (hostkit.ModelResult, error) {
			if input.Round == 1 {
				return hostkit.ModelResult{Response: llm.ChatResponse{
					Content: "tool requested",
					Calls:   []llm.FunctionCall{{Name: "lookup", Arguments: `{"topic":"agentx"}`}},
				}}, nil
			}
			return hostkit.ModelResult{Response: llm.ChatResponse{Content: "model-tool-ok"}}, nil
		},
		ExecuteTools: func(context.Context, hostkit.ModelToolRoundExchange) (hostkit.ToolResult, error) {
			return hostkit.ToolResult{NextChunks: []string{"portable tool result"}}, nil
		},
	}, nil
}

func runModelToolProbe(ctx context.Context) (string, error) {
	client, err := hostkit.NewModelToolClient(hostkit.ModelToolClientConfig{
		MaxRounds: 3,
		ResolveIdentity: func(request execution.Request) (string, string) {
			return "run-model-tool", request.SessionID
		},
		BuildRound: buildModelToolRound,
	})
	if err != nil {
		return "", err
	}
	result, runErr := client.Run(ctx, agentx.RunRequest{
		Input:     "exercise model/tool host kit",
		SessionID: "model-tool-session",
	})
	shutdownErr := client.Shutdown(context.Background())
	if runErr != nil {
		return "", runErr
	}
	if shutdownErr != nil {
		return "", shutdownErr
	}
	return fmt.Sprintf("%s:%s:%s", result.RunID, result.Status, result.Reply), nil
}

type workflowValidator struct{}

func (workflowValidator) ValidateSpec(workflow.Spec) error     { return nil }
func (workflowValidator) ValidateNode(workflow.NodeSpec) error { return nil }

type workflowMapper struct{}

func (workflowMapper) MapNode(workflow.NodeSpec, workflow.ExecutionMode) (workflowhostkit.MappedCall, error) {
	return workflowhostkit.MappedCall{
		Name:      "echo",
		Arguments: map[string]any{"source": "core-developer-preview-consumer"},
	}, nil
}

type workflowExecutor struct{}

func (workflowExecutor) Execute(context.Context, workflowhostkit.Call) (string, error) {
	return "workflow-ok", nil
}

func runWorkflowProbe(ctx context.Context) (string, error) {
	runtime, err := workflowhostkit.New(workflowhostkit.Config{
		Validator:          workflowValidator{},
		Mapper:             workflowMapper{},
		BasicExecutor:      workflowExecutor{},
		NewRunID:           func() string { return "run-workflow" },
		NewNodeExecutionID: func() string { return "nodeexec-workflow" },
		NowUnixMilli:       func() int64 { return 1 },
	})
	if err != nil {
		return "", err
	}
	result, err := runtime.Run(ctx, workflow.Spec{
		ID:        "workflow-consumer",
		Version:   "1",
		EntryNode: "echo",
		Nodes: []workflow.NodeSpec{{
			ID:   "echo",
			Kind: workflow.NodeTool,
		}},
	}, workflowhostkit.Inputs{})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%s:%s", result.Execution.RunID, result.Execution.FinalStatus, result.Execution.NodeOutput["echo"]), nil
}

func runAllProbes(ctx context.Context) (string, error) {
	if _, err := runProbe(); err != nil {
		return "", err
	}
	modelTool, err := runModelToolProbe(ctx)
	if err != nil {
		return "", err
	}
	workflowResult, err := runWorkflowProbe(ctx)
	if err != nil {
		return "", err
	}
	extensionResult, err := runExtensionProbe()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("agentx-core-developer-preview-ok:model-tool=%s:workflow=%s:extension=%s", modelTool, workflowResult, extensionResult), nil
}

func runExtensionProbe() (string, error) {
	manifest := astock.Manifest()
	if manifest.ID != astock.ModuleID || len(astock.ToolDefinitions()) != 7 {
		return "", fmt.Errorf("unexpected A-stock surface: manifest=%#v tools=%d", manifest, len(astock.ToolDefinitions()))
	}
	evaluation := astock.EvaluateValuationEvidence(astock.ValuationEvaluationInput{
		ExpectedEntityName: "平安银行",
		ExpectedStockCode:  "000001",
		EvidenceEntityName: "平安银行",
		EvidenceStockCode:  "sz000001",
		AdapterStatus:      astockcontracts.AdapterStatusOK,
		FailureCode:        astockcontracts.FailureCodeNone,
		AnswerReady:        true,
		RequestedFields:    []string{"price"},
		FieldValues:        map[string]string{"price": "11.38"},
		AsOf:               "2026-08-02T15:00:00+08:00",
		SourceURL:          "https://example.invalid/quote/sz000001",
	})
	if !evaluation.Passed {
		return "", fmt.Errorf("A-stock fixture evaluation failed: %#v", evaluation)
	}
	return fmt.Sprintf("%s:%d:%t", manifest.ID, len(astock.ToolDefinitions()), evaluation.Passed), nil
}
