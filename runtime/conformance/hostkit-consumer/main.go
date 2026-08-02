package main

import (
	"context"
	"errors"
	"fmt"

	agentx "github.com/wsnacj/agentx-go"
	llm "github.com/wsnacj/agentx-go/components/llm"
	"github.com/wsnacj/agentx-go/runtime/execution"
	"github.com/wsnacj/agentx-go/runtime/hostkit"
	"github.com/wsnacj/agentx-go/runtime/toolloop"
)

func buildRound(context.Context, execution.Request) (hostkit.ModelToolRoundConfig, error) {
	return hostkit.ModelToolRoundConfig{
		RequestModel: func(_ context.Context, _ toolloop.RoundExecutionInput) (hostkit.ModelResult, error) {
			return hostkit.ModelResult{Response: llm.ChatResponse{
				Content: "tool requested",
				Calls:   []llm.FunctionCall{{Name: "lookup", Arguments: `{"topic":"agentx"}`}},
			}}, nil
		},
		ExecuteTools: func(context.Context, hostkit.ModelToolRoundExchange) (hostkit.ToolResult, error) {
			return hostkit.ToolResult{DirectAnswer: &hostkit.ToolDirectAnswer{
				Reply:  "hostkit-conformance:direct",
				Source: "lookup",
				Reason: "conformance",
			}}, nil
		},
	}, nil
}

func run(ctx context.Context) (string, error) {
	client, err := hostkit.NewModelToolClient(hostkit.ModelToolClientConfig{
		MaxRounds: 3,
		ResolveIdentity: func(request execution.Request) (string, string) {
			return "hostkit-conformance", request.SessionID
		},
		BuildRound: buildRound,
	})
	if err != nil {
		return "", err
	}
	result, runErr := client.Run(ctx, agentx.RunRequest{
		Input:     "exercise portable host kit",
		SessionID: "hostkit-session",
	})
	shutdownErr := client.Shutdown(context.Background())
	if runErr != nil {
		return "", runErr
	}
	if shutdownErr != nil {
		return "", shutdownErr
	}
	chat, err := hostkit.NewChatClient(hostkit.ChatClientConfig{
		RequestModel: func(context.Context, llm.ChatRequest) (llm.ChatResponse, error) {
			return llm.ChatResponse{Content: "hostkit-conformance:chat"}, nil
		},
	})
	if err != nil {
		return "", err
	}
	chatResult, chatRunErr := chat.Run(ctx, agentx.RunRequest{Input: "chat"})
	chatShutdownErr := chat.Shutdown(context.Background())
	if chatRunErr != nil {
		return "", chatRunErr
	}
	if chatShutdownErr != nil {
		return "", chatShutdownErr
	}
	if _, closedErr := chat.Run(ctx, agentx.RunRequest{Input: "closed"}); !errors.Is(closedErr, &agentx.Error{Code: agentx.CodeClientClosed}) {
		return "", fmt.Errorf("chat client closed error = %v", closedErr)
	}
	return fmt.Sprintf("agentx-hostkit-ok:%s:%s:%s", result.Status, result.Reply, chatResult.Reply), nil
}

func main() {
	output, err := run(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println(output)
}
