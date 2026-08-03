package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	toolcontract "github.com/wsnacj/agentx-go/components/tool"
	agentxtools "github.com/wsnacj/agentx-go/tools"
	agent "github.com/wsnacj/agentx-go/tools/agent"
)

type memoryBackend struct {
	requests []agent.Request
}

func (b *memoryBackend) ExecuteTask(_ context.Context, request agent.Request) (string, error) {
	b.requests = append(b.requests, request)
	return `{"task_id":"task-demo","session_id":"session-demo","status":"queued"}`, nil
}

func (b *memoryBackend) ExecuteSubagent(_ context.Context, request agent.Request) (string, error) {
	b.requests = append(b.requests, request)
	return `{"action":"run","task_id":"task-demo","session_id":"session-demo"}`, nil
}

func (b *memoryBackend) ExecuteAgentStep(_ context.Context, request agent.Request) (string, error) {
	b.requests = append(b.requests, request)
	return `{"result":{"summary":"done"}}`, nil
}

func run(ctx context.Context) error {
	registry := agentxtools.NewRegistry()
	backend := &memoryBackend{}
	agent.Register(registry, agent.Options{Backend: backend})

	if _, err := registry.Execute(ctx, toolcontract.Call{
		Name: agent.TasksRunName, Arguments: `{"seed_message":"inspect portable runtime"}`,
	}); err != nil {
		return err
	}
	result, err := registry.Execute(ctx, toolcontract.Call{
		Name: agent.SubagentsName, Arguments: `{"action":"run","message":"verify the result"}`,
	})
	if err != nil {
		return err
	}
	if !json.Valid([]byte(result)) || len(backend.requests) != 2 {
		return fmt.Errorf("unexpected result or request count")
	}
	request := backend.requests[1]
	if request.Action != "run" || strings.TrimSpace(request.Arguments["seed_message"].(string)) == "" {
		return fmt.Errorf("subagent request was not normalized")
	}
	return nil
}

func main() {
	if err := run(context.Background()); err != nil {
		panic(err)
	}
	fmt.Println("verified=true capability=agent_task_subagent")
}
