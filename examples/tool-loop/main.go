// tool-loop 展示 Open Tool Loop 与 Tool Direct Answer 的推荐组合。
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	agentx "github.com/wsnacj/agentx-go"
	llm "github.com/wsnacj/agentx-go/components/llm"
	toolcontract "github.com/wsnacj/agentx-go/components/tool"
	"github.com/wsnacj/agentx-go/runtime/execution"
	"github.com/wsnacj/agentx-go/runtime/hostkit"
	"github.com/wsnacj/agentx-go/runtime/toolloop"
	tools "github.com/wsnacj/agentx-go/tools"
	"github.com/wsnacj/agentx-go/tools/diffs"
)

func run(ctx context.Context) (agentx.RunResult, error) {
	registry := tools.NewRegistry()
	diffs.Register(registry) // Host显式选择纯文本、无文件副作用的tool。
	client, err := hostkit.NewModelToolClient(hostkit.ModelToolClientConfig{
		MaxRounds: 2,
		ResolveIdentity: func(request execution.Request) (string, string) {
			return "run-tool-loop-example", request.SessionID
		},
		BuildRound: func(context.Context, execution.Request) (hostkit.ModelToolRoundConfig, error) {
			return hostkit.ModelToolRoundConfig{
				RequestModel: func(context.Context, toolloop.RoundExecutionInput) (hostkit.ModelResult, error) {
					arguments, _ := json.Marshal(map[string]any{"before": "old\n", "after": "new\n", "path": "demo.txt"})
					return hostkit.ModelResult{Model: "fixture", Response: llm.ChatResponse{
						Content: "compare text", Calls: []llm.FunctionCall{{Name: "diffs", Arguments: string(arguments)}},
					}}, nil
				},
				ExecuteTools: func(ctx context.Context, exchange hostkit.ModelToolRoundExchange) (hostkit.ToolResult, error) {
					call := exchange.Model.Response.Calls[0]
					result, err := registry.Execute(ctx, toolcontract.Call{Name: call.Name, Arguments: call.Arguments})
					if err != nil {
						return hostkit.ToolResult{}, err
					}
					return hostkit.ToolResult{DirectAnswer: &hostkit.ToolDirectAnswer{
						Reply: "diff ready: " + result, Source: "diffs", Reason: "fixture result verified by host",
					}}, nil
				},
			}, nil
		},
	})
	if err != nil {
		return agentx.RunResult{}, err
	}
	result, runErr := client.Run(ctx, agentx.RunRequest{Input: "compare", SessionID: "tool-loop-example"})
	shutdownErr := client.Shutdown(context.Background())
	if runErr != nil {
		return result, runErr
	}
	return result, shutdownErr
}

func main() {
	result, err := run(context.Background())
	if err != nil || result.Status != "completed" || !strings.HasPrefix(result.Reply, "diff ready:") {
		fmt.Fprintln(os.Stderr, "tool-loop example failed", result, err)
		os.Exit(1)
	}
	fmt.Println(result.Reply)
}
