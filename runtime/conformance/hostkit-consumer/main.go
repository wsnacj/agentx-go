package main

import (
	"context"
	"fmt"

	agentx "github.com/wsnacj/agentx-go"
	llm "github.com/wsnacj/agentx-go/components/llm"
	"github.com/wsnacj/agentx-go/runtime/execution"
	"github.com/wsnacj/agentx-go/runtime/hostkit"
	"github.com/wsnacj/agentx-go/runtime/toolloop"
)

func buildRound(context.Context, execution.Request) (hostkit.ModelToolRoundConfig, error) {
	return hostkit.ModelToolRoundConfig{
		RequestModel: func(_ context.Context, input toolloop.RoundExecutionInput) (hostkit.ModelResult, error) {
			if input.Round == 1 {
				return hostkit.ModelResult{Response: llm.ChatResponse{
					Content: "tool requested",
					Calls:   []llm.FunctionCall{{Name: "lookup", Arguments: `{"topic":"agentx"}`}},
				}}, nil
			}
			return hostkit.ModelResult{Response: llm.ChatResponse{Content: "hostkit-conformance:2"}}, nil
		},
		ExecuteTools: func(context.Context, hostkit.ModelToolRoundExchange) (hostkit.ToolResult, error) {
			return hostkit.ToolResult{NextChunks: []string{"portable result"}}, nil
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
	return fmt.Sprintf("agentx-hostkit-ok:%s:%s", result.Status, result.Reply), nil
}

func main() {
	output, err := run(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println(output)
}
