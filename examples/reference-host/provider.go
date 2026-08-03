package main

import (
	"context"
	"encoding/json"

	llm "github.com/wsnacj/agentx-go/components/llm"
)

type modelProvider interface {
	Request(context.Context, llm.ChatRequest) (llm.ChatResponse, error)
}

type fixtureProvider struct {
	toolName string
}

func (provider fixtureProvider) Request(ctx context.Context, request llm.ChatRequest) (llm.ChatResponse, error) {
	if err := ctx.Err(); err != nil {
		return llm.ChatResponse{}, err
	}
	message := request.Messages[len(request.Messages)-1].Content
	if provider.toolName == "" {
		return llm.ChatResponse{Content: "fixture: " + message}, nil
	}
	arguments, _ := json.Marshal(map[string]any{
		"before": "old\n",
		"after":  "new\n",
		"path":   "reference.txt",
	})
	return llm.ChatResponse{
		Content: "fixture tool request",
		Calls: []llm.FunctionCall{{
			Name:      provider.toolName,
			Arguments: string(arguments),
		}},
	}, nil
}
