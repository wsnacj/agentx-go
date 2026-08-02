package hostkit

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"

	llm "github.com/wsnacj/agentx-go/components/llm"
	"github.com/wsnacj/agentx-go/runtime/toolloop"
)

var errRoundAdapterPort = errors.New("round adapter port failed")

func TestModelToolRoundAdapterCompletesModelOnlyRound(t *testing.T) {
	var order []string
	adapter, err := NewModelToolRoundAdapter(ModelToolRoundConfig{
		RequestModel: func(_ context.Context, input toolloop.RoundExecutionInput) (ModelResult, error) {
			order = append(order, "request")
			input.State.Chunks[0] = "mutated"
			return ModelResult{Response: llm.ChatResponse{Content: "done"}, Model: "test-model"}, nil
		},
		ObserveResponse: func(_ context.Context, exchange ModelToolRoundExchange) (string, error) {
			order = append(order, "observe")
			if exchange.Model.Response.Content != "done" || exchange.Reply != "" {
				t.Fatalf("exchange = %#v", exchange)
			}
			return exchange.Model.Response.Content, nil
		},
		BeforeTools: func(context.Context, ModelToolRoundExchange) (bool, error) {
			t.Fatal("BeforeTools called without tool calls")
			return false, nil
		},
		ExecuteTools: func(context.Context, ModelToolRoundExchange) (ToolResult, error) {
			t.Fatal("ExecuteTools called without tool calls")
			return ToolResult{}, nil
		},
	})
	if err != nil {
		t.Fatalf("NewModelToolRoundAdapter() error = %v", err)
	}
	input := toolloop.RoundExecutionInput{Round: 1, MaxRounds: 2, State: toolloop.RoundState{Chunks: []string{"original"}}}
	result, err := adapter.ExecuteRound(context.Background(), input)
	if err != nil {
		t.Fatalf("ExecuteRound() error = %v", err)
	}
	if result.Kind != toolloop.OutcomeCompleted || result.Reply != "done" || result.Continuation != nil {
		t.Fatalf("result = %#v", result)
	}
	if !reflect.DeepEqual(order, []string{"request", "observe"}) || input.State.Chunks[0] != "original" {
		t.Fatalf("order=%v input=%#v", order, input)
	}
}

func TestModelToolRoundAdapterProjectsToolContinuation(t *testing.T) {
	var order []string
	adapter, err := NewModelToolRoundAdapter(ModelToolRoundConfig{
		RequestModel: func(context.Context, toolloop.RoundExecutionInput) (ModelResult, error) {
			order = append(order, "request")
			return ModelResult{Response: llm.ChatResponse{
				Content: "calling",
				Calls:   []llm.FunctionCall{{ID: "call-1", Name: "lookup", Arguments: `{"q":"agentx"}`}},
			}, Model: "test-model", Recovered: true}, nil
		},
		ObserveResponse: func(context.Context, ModelToolRoundExchange) (string, error) {
			order = append(order, "observe")
			return "observed", nil
		},
		BeforeTools: func(_ context.Context, exchange ModelToolRoundExchange) (bool, error) {
			order = append(order, "gate")
			if exchange.Reply != "observed" || len(exchange.Model.Response.Calls) != 1 {
				t.Fatalf("gate exchange = %#v", exchange)
			}
			return true, nil
		},
		ExecuteTools: func(_ context.Context, exchange ModelToolRoundExchange) (ToolResult, error) {
			order = append(order, "tools")
			return ToolResult{
				Runs:       []toolloop.RunObservation{{Name: "lookup", Output: "result"}},
				Failures:   []toolloop.FailureObservation{{Tool: "lookup"}},
				NextChunks: []string{"tool result"},
			}, nil
		},
	})
	if err != nil {
		t.Fatalf("NewModelToolRoundAdapter() error = %v", err)
	}
	result, err := adapter.ExecuteRound(context.Background(), toolloop.RoundExecutionInput{Round: 1, MaxRounds: 2})
	if err != nil {
		t.Fatalf("ExecuteRound() error = %v", err)
	}
	if result.Kind != toolloop.OutcomeContinue || result.Reply != "observed" || result.Continuation == nil {
		t.Fatalf("result = %#v", result)
	}
	if !reflect.DeepEqual(order, []string{"request", "observe", "gate", "tools"}) {
		t.Fatalf("order = %v", order)
	}
	if !reflect.DeepEqual(result.Continuation.Calls, []toolloop.Call{{Name: "lookup", Arguments: `{"q":"agentx"}`}}) ||
		!reflect.DeepEqual(result.Continuation.NextChunks, []string{"tool result"}) {
		t.Fatalf("continuation = %#v", result.Continuation)
	}
}

func TestModelToolRoundAdapterProjectsExplicitToolDirectAnswer(t *testing.T) {
	adapter, err := NewModelToolRoundAdapter(ModelToolRoundConfig{
		RequestModel: func(context.Context, toolloop.RoundExecutionInput) (ModelResult, error) {
			return ModelResult{Response: llm.ChatResponse{Calls: []llm.FunctionCall{{Name: "lookup"}}}}, nil
		},
		ExecuteTools: func(context.Context, ModelToolRoundExchange) (ToolResult, error) {
			return ToolResult{DirectAnswer: &ToolDirectAnswer{
				Reply:  "display-safe tool answer",
				Source: "lookup",
				Reason: "authoritative_result",
			}}, nil
		},
	})
	if err != nil {
		t.Fatalf("NewModelToolRoundAdapter() error = %v", err)
	}
	result, err := adapter.ExecuteRound(context.Background(), toolloop.RoundExecutionInput{Round: 1, MaxRounds: 2})
	if err != nil {
		t.Fatalf("ExecuteRound() error = %v", err)
	}
	if result.Kind != toolloop.OutcomeCompleted || result.Reply != "display-safe tool answer" || result.Continuation != nil {
		t.Fatalf("result = %#v", result)
	}
}

func TestModelToolRoundResultRejectsEmptyToolDirectAnswer(t *testing.T) {
	_, err := (ModelToolRoundResult{
		Phase: toolloop.RoundPhaseResult{Kind: toolloop.RoundPhaseActionCompleted},
		Tools: ToolResult{DirectAnswer: &ToolDirectAnswer{Reply: "  "}},
	}).ExecutionResult()
	if err == nil || err.Error() != "agentx host kit: tool direct answer reply is required" {
		t.Fatalf("ExecutionResult() error = %v", err)
	}
}

func TestModelToolRoundAdapterHostStopSkipsTools(t *testing.T) {
	adapter, err := NewModelToolRoundAdapter(ModelToolRoundConfig{
		RequestModel: func(context.Context, toolloop.RoundExecutionInput) (ModelResult, error) {
			return ModelResult{Response: llm.ChatResponse{Content: "calling", Calls: []llm.FunctionCall{{Name: "lookup"}}}}, nil
		},
		BeforeTools: func(context.Context, ModelToolRoundExchange) (bool, error) { return false, nil },
		ExecuteTools: func(context.Context, ModelToolRoundExchange) (ToolResult, error) {
			t.Fatal("ExecuteTools called after host stop")
			return ToolResult{}, nil
		},
	})
	if err != nil {
		t.Fatalf("NewModelToolRoundAdapter() error = %v", err)
	}
	result, err := adapter.ExecuteRound(context.Background(), toolloop.RoundExecutionInput{Round: 1, MaxRounds: 1})
	if err != nil {
		t.Fatalf("ExecuteRound() error = %v", err)
	}
	if result.Kind != toolloop.OutcomeTerminated || result.Reply != "calling" {
		t.Fatalf("result = %#v", result)
	}
}

func TestModelToolRoundAdapterPreservesPhaseErrorIdentity(t *testing.T) {
	for _, test := range []struct {
		name   string
		config ModelToolRoundConfig
		phase  toolloop.RoundPhase
	}{
		{
			name: "request",
			config: ModelToolRoundConfig{RequestModel: func(context.Context, toolloop.RoundExecutionInput) (ModelResult, error) {
				return ModelResult{}, errRoundAdapterPort
			}},
			phase: toolloop.RoundPhaseRequest,
		},
		{
			name: "observe",
			config: ModelToolRoundConfig{
				RequestModel:    func(context.Context, toolloop.RoundExecutionInput) (ModelResult, error) { return ModelResult{}, nil },
				ObserveResponse: func(context.Context, ModelToolRoundExchange) (string, error) { return "", errRoundAdapterPort },
			},
			phase: toolloop.RoundPhaseObserve,
		},
		{
			name: "gate",
			config: ModelToolRoundConfig{
				RequestModel: func(context.Context, toolloop.RoundExecutionInput) (ModelResult, error) {
					return ModelResult{Response: llm.ChatResponse{Calls: []llm.FunctionCall{{Name: "tool"}}}}, nil
				},
				BeforeTools: func(context.Context, ModelToolRoundExchange) (bool, error) { return false, errRoundAdapterPort },
			},
			phase: toolloop.RoundPhaseBeforeAction,
		},
		{
			name: "tools",
			config: ModelToolRoundConfig{RequestModel: func(context.Context, toolloop.RoundExecutionInput) (ModelResult, error) {
				return ModelResult{Response: llm.ChatResponse{Calls: []llm.FunctionCall{{Name: "tool"}}}}, nil
			}},
			phase: toolloop.RoundPhaseAct,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			adapter, err := NewModelToolRoundAdapter(test.config)
			if err != nil {
				t.Fatalf("NewModelToolRoundAdapter() error = %v", err)
			}
			result, err := adapter.Execute(context.Background(), toolloop.RoundExecutionInput{Round: 1, MaxRounds: 1})
			if !errors.Is(err, errRoundAdapterPort) && !(test.name == "tools" && err != nil) {
				t.Fatalf("Execute() error = %v", err)
			}
			if result.Phase.LastPhase != test.phase {
				t.Fatalf("phase = %q, want %q", result.Phase.LastPhase, test.phase)
			}
		})
	}
}

func TestModelToolRoundAdapterCanRunConcurrentlyWithIndependentPorts(t *testing.T) {
	var mu sync.Mutex
	requests := 0
	adapter, err := NewModelToolRoundAdapter(ModelToolRoundConfig{RequestModel: func(context.Context, toolloop.RoundExecutionInput) (ModelResult, error) {
		mu.Lock()
		requests++
		mu.Unlock()
		return ModelResult{Response: llm.ChatResponse{Content: "done"}}, nil
	}})
	if err != nil {
		t.Fatalf("NewModelToolRoundAdapter() error = %v", err)
	}
	var wg sync.WaitGroup
	for index := 0; index < 8; index++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, runErr := adapter.ExecuteRound(context.Background(), toolloop.RoundExecutionInput{Round: 1, MaxRounds: 1})
			if runErr != nil || result.Kind != toolloop.OutcomeCompleted {
				t.Errorf("ExecuteRound() = %#v, %v", result, runErr)
			}
		}()
	}
	wg.Wait()
	if requests != 8 {
		t.Fatalf("requests = %d", requests)
	}
}
