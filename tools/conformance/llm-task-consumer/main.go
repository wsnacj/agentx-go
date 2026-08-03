package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	llm "github.com/wsnacj/agentx-go/components/llm"
	toolcontract "github.com/wsnacj/agentx-go/components/tool"
	"github.com/wsnacj/agentx-go/tools"
	"github.com/wsnacj/agentx-go/tools/llmtask"
)

type result struct {
	Registered []string `json:"registered"`
	Model      string   `json:"model"`
	Answer     string   `json:"answer"`
	Verified   bool     `json:"verified"`
}

func run(ctx context.Context) (result, error) {
	registry := tools.NewRegistry()
	llmtask.Register(registry, llmtask.Options{
		ModelConfig: "fixture-model",
		ChatWithInput: func(ctx context.Context, input llm.ChatInput) (*llm.ChatResponse, error) {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			if input.ConfigName != "fixture-model" || input.ToolChoice == nil || input.ToolChoice.Type != "none" {
				return nil, fmt.Errorf("unexpected model input")
			}
			return &llm.ChatResponse{Content: "```json\n{\n// external fixture\n\"answer\":\"ready\"\n}\n```"}, nil
		},
	})
	definitions := registry.Definitions()
	registered := make([]string, 0, len(definitions))
	for _, definition := range definitions {
		registered = append(registered, definition.Function.Name)
	}
	raw, err := registry.Execute(ctx, toolcontract.Call{
		Name:      llmtask.Name,
		Arguments: `{"instruction":"return readiness","schema":{"type":"object","additionalProperties":false,"required":["answer"],"properties":{"answer":{"type":"string","const":"ready"}}}}`,
	})
	if err != nil {
		return result{}, err
	}
	var payload struct {
		Model  string         `json:"model"`
		Result map[string]any `json:"result"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return result{}, err
	}
	answer, _ := payload.Result["answer"].(string)
	verified := len(registered) == 1 && registered[0] == llmtask.Name &&
		payload.Model == "fixture-model" && strings.TrimSpace(answer) == "ready"
	return result{Registered: registered, Model: payload.Model, Answer: answer, Verified: verified}, nil
}

func main() {
	value, err := run(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := json.NewEncoder(os.Stdout).Encode(value); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
