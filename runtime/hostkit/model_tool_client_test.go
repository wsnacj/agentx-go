package hostkit

import (
	"context"
	"errors"
	"testing"

	agentx "github.com/wsnacj/agentx-go"
	llm "github.com/wsnacj/agentx-go/components/llm"
	"github.com/wsnacj/agentx-go/runtime/execution"
	"github.com/wsnacj/agentx-go/runtime/toolloop"
)

func TestNewModelToolClientRunsWithoutCustomFactoryOrRoundExecutor(t *testing.T) {
	shutdownCalls := 0
	client, err := NewModelToolClient(ModelToolClientConfig{
		MaxRounds: 2,
		ResolveIdentity: func(request execution.Request) (string, string) {
			return "run-simple", request.SessionID
		},
		BuildRound: func(context.Context, execution.Request) (ModelToolRoundConfig, error) {
			return ModelToolRoundConfig{
				RequestModel: func(_ context.Context, input toolloop.RoundExecutionInput) (ModelResult, error) {
					if input.Round == 1 {
						return ModelResult{Response: llm.ChatResponse{Calls: []llm.FunctionCall{{Name: "lookup"}}}}, nil
					}
					if len(input.State.Chunks) != 1 || input.State.Chunks[0] != "tool result" {
						t.Fatalf("continuation state = %#v", input.State)
					}
					return ModelResult{Response: llm.ChatResponse{Content: "done"}}, nil
				},
				ExecuteTools: func(context.Context, ModelToolRoundExchange) (ToolResult, error) {
					return ToolResult{NextChunks: []string{"tool result"}}, nil
				},
			}, nil
		},
		Shutdown: func(context.Context) error {
			shutdownCalls++
			return nil
		},
	})
	if err != nil {
		t.Fatalf("NewModelToolClient() error = %v", err)
	}
	result, err := client.Run(context.Background(), agentx.RunRequest{Input: "inspect", SessionID: "session-simple"})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.RunID != "run-simple" || result.SessionID != "session-simple" || result.Status != "completed" || result.Reply != "done" {
		t.Fatalf("result = %#v", result)
	}
	if err := client.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
	if shutdownCalls != 1 {
		t.Fatalf("shutdown calls = %d", shutdownCalls)
	}
}

func TestNewModelToolClientPreservesBuildErrorIdentityAndResolvedIdentity(t *testing.T) {
	client, err := NewModelToolClient(ModelToolClientConfig{
		MaxRounds: 1,
		ResolveIdentity: func(request execution.Request) (string, string) {
			return "run-build-error", request.SessionID
		},
		BuildRound: func(context.Context, execution.Request) (ModelToolRoundConfig, error) {
			return ModelToolRoundConfig{}, errRoundAdapterPort
		},
		ClassifyError: func(err error) agentx.ErrorCode {
			if errors.Is(err, errRoundAdapterPort) {
				return agentx.CodeClientClosed
			}
			return agentx.CodeExecutionFailed
		},
	})
	if err != nil {
		t.Fatalf("NewModelToolClient() error = %v", err)
	}
	result, err := client.Run(context.Background(), agentx.RunRequest{Input: "fail", SessionID: "session-build-error"})
	if !errors.Is(err, errRoundAdapterPort) {
		t.Fatalf("Run() error = %v", err)
	}
	var typed *agentx.Error
	if !errors.As(err, &typed) || typed.Code != agentx.CodeClientClosed {
		t.Fatalf("typed error = %#v", typed)
	}
	if result.RunID != "run-build-error" || result.SessionID != "session-build-error" || result.Status != "failed" {
		t.Fatalf("result = %#v", result)
	}
}

func TestNewModelToolClientValidatesMinimalConfig(t *testing.T) {
	if _, err := NewModelToolClient(ModelToolClientConfig{}); err == nil || err.Error() != "agentx host kit: max rounds must be positive" {
		t.Fatalf("max-rounds error = %v", err)
	}
	if _, err := NewModelToolClient(ModelToolClientConfig{MaxRounds: 1}); err == nil || err.Error() != "agentx host kit: round builder is required" {
		t.Fatalf("round-builder error = %v", err)
	}
}
