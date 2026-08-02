package hostkit

import (
	"context"
	"fmt"

	agentx "github.com/wsnacj/agentx-go"
	llm "github.com/wsnacj/agentx-go/components/llm"
	"github.com/wsnacj/agentx-go/runtime/execution"
	"github.com/wsnacj/agentx-go/runtime/toolloop"
)

// ChatClientConfig is the low-boilerplate Model Conversation path. It executes
// exactly one model request per Run and deliberately rejects tool calls.
// Conversation persistence and provider selection remain host-owned.
type ChatClientConfig struct {
	Profile agentx.ExecutionProfile
	Model   string
	System  string

	BuildRequest func(context.Context, execution.Request) (llm.ChatRequest, error)
	RequestModel func(context.Context, llm.ChatRequest) (llm.ChatResponse, error)

	ResolveIdentity func(execution.Request) (runID string, sessionID string)
	Shutdown        func(context.Context) error
	ClassifyError   func(error) agentx.ErrorCode
}

// NewChatClient constructs a single-round Model Conversation Client. When
// BuildRequest is nil, the Run input becomes one user message. A custom
// BuildRequest may load prior conversation by SessionID without exposing a
// storage backend to the canonical API.
func NewChatClient(config ChatClientConfig) (*agentx.Client, error) {
	if config.RequestModel == nil {
		return nil, fmt.Errorf("agentx host kit: chat model requester is required")
	}
	return NewModelToolClient(ModelToolClientConfig{
		Profile:         config.Profile,
		MaxRounds:       1,
		ResolveIdentity: config.ResolveIdentity,
		Shutdown:        config.Shutdown,
		ClassifyError:   config.ClassifyError,
		BuildRound: func(ctx context.Context, request execution.Request) (ModelToolRoundConfig, error) {
			chatRequest := llm.ChatRequest{
				Model:  config.Model,
				System: config.System,
				Messages: llm.Conversation{{
					Role:    "user",
					Content: request.Input,
				}},
			}
			if config.BuildRequest != nil {
				var err error
				chatRequest, err = config.BuildRequest(ctx, request)
				if err != nil {
					return ModelToolRoundConfig{}, err
				}
			}
			return ModelToolRoundConfig{
				RequestModel: func(ctx context.Context, _ toolloop.RoundExecutionInput) (ModelResult, error) {
					response, err := config.RequestModel(ctx, chatRequest)
					if err != nil {
						return ModelResult{}, err
					}
					if len(response.Calls) != 0 {
						return ModelResult{}, fmt.Errorf("agentx host kit: chat model returned tool calls")
					}
					return ModelResult{Response: response, Model: chatRequest.Model}, nil
				},
			}, nil
		},
	})
}
