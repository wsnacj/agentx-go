// chat 展示最小 Model Conversation 接入：调用方只注入一次模型请求函数。
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	agentx "github.com/wsnacj/agentx-go"
	llm "github.com/wsnacj/agentx-go/components/llm"
	"github.com/wsnacj/agentx-go/runtime/hostkit"
)

func run(ctx context.Context, input string) (agentx.RunResult, error) {
	client, err := hostkit.NewChatClient(hostkit.ChatClientConfig{
		Model: "fixture",
		RequestModel: func(_ context.Context, request llm.ChatRequest) (llm.ChatResponse, error) {
			message := request.Messages[len(request.Messages)-1].Content
			return llm.ChatResponse{Content: "fixture: " + message}, nil
		},
	})
	if err != nil {
		return agentx.RunResult{}, err
	}
	result, runErr := client.Run(ctx, agentx.RunRequest{Input: input, SessionID: "chat-example"})
	shutdownErr := client.Shutdown(context.Background())
	if runErr != nil {
		return result, runErr
	}
	return result, shutdownErr
}

func main() {
	result, err := run(context.Background(), "hello AgentX")
	if err != nil || result.Status != "completed" || !strings.HasPrefix(result.Reply, "fixture:") {
		fmt.Fprintln(os.Stderr, "chat example failed", result, err)
		os.Exit(1)
	}
	fmt.Println(result.Reply)
}
