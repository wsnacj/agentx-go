package hostkit

import (
	"context"
	"errors"
	"testing"

	agentx "github.com/wsnacj/agentx-go"
	llm "github.com/wsnacj/agentx-go/components/llm"
	"github.com/wsnacj/agentx-go/runtime/execution"
)

func TestNewChatClientUsesDefaultSingleTurnRequest(t *testing.T) {
	client, err := NewChatClient(ChatClientConfig{
		Model:  "test-model",
		System: "be concise",
		ResolveIdentity: func(request execution.Request) (string, string) {
			return "chat-run", request.SessionID
		},
		RequestModel: func(_ context.Context, request llm.ChatRequest) (llm.ChatResponse, error) {
			if request.Model != "test-model" || request.System != "be concise" || len(request.Messages) != 1 ||
				request.Messages[0].Role != "user" || request.Messages[0].Content != "hello" {
				t.Fatalf("request = %#v", request)
			}
			return llm.ChatResponse{Content: "hello from model"}, nil
		},
	})
	if err != nil {
		t.Fatalf("NewChatClient() error = %v", err)
	}
	result, err := client.Run(context.Background(), agentx.RunRequest{Input: "hello", SessionID: "chat-session"})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.RunID != "chat-run" || result.SessionID != "chat-session" || result.Status != "completed" || result.Reply != "hello from model" {
		t.Fatalf("result = %#v", result)
	}
}

func TestNewChatClientPreservesCancellationAndRejectsToolCalls(t *testing.T) {
	client, err := NewChatClient(ChatClientConfig{
		RequestModel: func(ctx context.Context, _ llm.ChatRequest) (llm.ChatResponse, error) {
			if err := ctx.Err(); err != nil {
				return llm.ChatResponse{}, err
			}
			return llm.ChatResponse{Calls: []llm.FunctionCall{{Name: "unexpected"}}}, nil
		},
	})
	if err != nil {
		t.Fatalf("NewChatClient() error = %v", err)
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := client.Run(cancelled, agentx.RunRequest{Input: "cancel"}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled Run() error = %v", err)
	}
	if _, err := client.Run(context.Background(), agentx.RunRequest{Input: "tool"}); err == nil {
		t.Fatal("expected tool-call error")
	} else {
		var typed *agentx.Error
		if !errors.As(err, &typed) || typed.Code != agentx.CodeExecutionFailed || err.Error() != "execution failed" {
			t.Fatalf("tool-call Run() error = %v, typed = %#v", err, typed)
		}
	}
}

func TestNewChatClientValidatesRequester(t *testing.T) {
	if _, err := NewChatClient(ChatClientConfig{}); err == nil || err.Error() != "agentx host kit: chat model requester is required" {
		t.Fatalf("NewChatClient() error = %v", err)
	}
}
