package llmtask_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	llm "github.com/wsnacj/agentx-go/components/llm"
	toolcontract "github.com/wsnacj/agentx-go/components/tool"
	agentxtoolerrors "github.com/wsnacj/agentx-go/runtime/toolerrors"
	"github.com/wsnacj/agentx-go/tools"
	"github.com/wsnacj/agentx-go/tools/llmtask"
)

func TestExternalHostCanRegisterAndRunLLMTask(t *testing.T) {
	registry := tools.NewRegistry()
	llmtask.Register(registry, llmtask.Options{
		ModelConfig: "fixture-model",
		ChatWithInput: func(ctx context.Context, input llm.ChatInput) (*llm.ChatResponse, error) {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			if input.ConfigName != "fixture-model" || input.ToolChoice == nil || input.ToolChoice.Type != "none" {
				t.Fatalf("unexpected input: %#v", input)
			}
			return &llm.ChatResponse{Content: `{"answer":"ready"}`}, nil
		},
	})
	value, err := registry.Execute(context.Background(), toolcontract.Call{
		Name:      llmtask.Name,
		Arguments: `{"instruction":"return readiness","schema":{"type":"object","required":["answer"],"properties":{"answer":{"type":"string"}}}}`,
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	var payload struct {
		Tool          string         `json:"tool"`
		Model         string         `json:"model"`
		SchemaApplied bool           `json:"schema_applied"`
		Result        map[string]any `json:"result"`
	}
	if err := json.Unmarshal([]byte(value), &payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if payload.Tool != llmtask.Name || payload.Model != "fixture-model" || !payload.SchemaApplied || payload.Result["answer"] != "ready" {
		t.Fatalf("payload: %#v", payload)
	}
}

func TestExternalContractPreservesTypedArgumentsAndCancellation(t *testing.T) {
	handler := llmtask.NewHandler(llmtask.Options{
		ModelConfig:      "fixture-model",
		DefaultTimeoutMs: 10,
		ChatWithInput: func(ctx context.Context, _ llm.ChatInput) (*llm.ChatResponse, error) {
			<-ctx.Done()
			return nil, ctx.Err()
		},
	})
	_, err := handler(context.Background(), toolcontract.Call{Name: llmtask.Name, Arguments: `{}`})
	var argumentError *agentxtoolerrors.ToolArgumentError
	if !errors.As(err, &argumentError) || argumentError.Code != agentxtoolerrors.ToolArgumentErrorCodeMissingRequiredArgument {
		t.Fatalf("typed argument error = %T %v", err, err)
	}
	started := time.Now()
	_, err = handler(context.Background(), toolcontract.Call{Name: llmtask.Name, Arguments: `{"instruction":"wait"}`})
	if !errors.Is(err, context.DeadlineExceeded) || time.Since(started) > time.Second {
		t.Fatalf("deadline error=%v duration=%s", err, time.Since(started))
	}
}

func TestRegisterWithoutModelAdapterIsFailClosed(t *testing.T) {
	registry := tools.NewRegistry()
	llmtask.Register(registry, llmtask.Options{ModelConfig: "fixture-model"})
	if len(registry.Definitions()) != 0 {
		t.Fatalf("unexpected definitions: %#v", registry.Definitions())
	}
}
