package main

import (
	"context"
	"fmt"
	"time"

	agentx "github.com/wsnacj/agentx-go"
	llm "github.com/wsnacj/agentx-go/components/llm"
	toolcontract "github.com/wsnacj/agentx-go/components/tool"
	"github.com/wsnacj/agentx-go/runtime/execution"
	"github.com/wsnacj/agentx-go/runtime/hostkit"
	runstore "github.com/wsnacj/agentx-go/runtime/runstore"
	"github.com/wsnacj/agentx-go/runtime/toolloop"
	tools "github.com/wsnacj/agentx-go/tools"
	"github.com/wsnacj/agentx-go/tools/diffs"
)

type referenceHost struct {
	client *agentx.Client
	store  runstore.Store
	input  string
	runID  string
}

func newReferenceHost(value config) (*referenceHost, error) {
	value, err := normalizeConfig(value)
	if err != nil {
		return nil, err
	}
	store := runstore.NewMemoryStore()
	provider := fixtureProvider{}
	runID := "reference-host-run"
	resolveIdentity := func(request execution.Request) (string, string) {
		return runID, request.SessionID
	}

	var client *agentx.Client
	switch value.Mode {
	case modeChat:
		client, err = hostkit.NewChatClient(hostkit.ChatClientConfig{
			Model:           value.Provider,
			RequestModel:    provider.Request,
			ResolveIdentity: resolveIdentity,
		})
	case modeToolLoop:
		provider.toolName = value.Tools
		registry := tools.NewRegistry()
		diffs.Register(registry)
		client, err = hostkit.NewModelToolClient(hostkit.ModelToolClientConfig{
			MaxRounds:       2,
			ResolveIdentity: resolveIdentity,
			BuildRound: func(context.Context, execution.Request) (hostkit.ModelToolRoundConfig, error) {
				return hostkit.ModelToolRoundConfig{
					RequestModel: func(ctx context.Context, input toolloop.RoundExecutionInput) (hostkit.ModelResult, error) {
						response, requestErr := provider.Request(ctx, llm.ChatRequest{Model: value.Provider, Messages: llm.Conversation{{Role: "user", Content: input.State.Chunks[0]}}})
						return hostkit.ModelResult{Model: value.Provider, Response: response}, requestErr
					},
					ExecuteTools: func(ctx context.Context, exchange hostkit.ModelToolRoundExchange) (hostkit.ToolResult, error) {
						call := exchange.Model.Response.Calls[0]
						result, executeErr := registry.Execute(ctx, toolcontract.Call{Name: call.Name, Arguments: call.Arguments})
						if executeErr != nil {
							return hostkit.ToolResult{}, executeErr
						}
						return hostkit.ToolResult{DirectAnswer: &hostkit.ToolDirectAnswer{Reply: result, Source: call.Name, Reason: "offline fixture"}}, nil
					},
				}, nil
			},
		})
	}
	if err != nil {
		return nil, err
	}
	return &referenceHost{client: client, store: store, input: value.Input, runID: runID}, nil
}

func (host *referenceHost) Run(ctx context.Context) (agentx.RunResult, error) {
	if host == nil || host.client == nil || host.store == nil {
		return agentx.RunResult{}, fmt.Errorf("reference host: host is not configured")
	}
	now := time.Now().UnixMilli()
	if err := host.store.CreateRun(ctx, runstore.Run{RunID: host.runID, Status: "running", StartedAt: now}); err != nil {
		return agentx.RunResult{}, err
	}
	result, err := host.client.Run(ctx, agentx.RunRequest{Input: host.input, SessionID: "reference-host-session"})
	status := result.Status
	if status == "" {
		status = "failed"
	}
	if updateErr := host.store.UpdateRun(ctx, runstore.Run{
		RunID: host.runID, Status: status, StartedAt: now, FinishedAt: time.Now().UnixMilli(), Summary: result.Reply,
	}); err == nil && updateErr != nil {
		err = updateErr
	}
	return result, err
}

func (host *referenceHost) Shutdown(ctx context.Context) error {
	if host == nil || host.client == nil {
		return nil
	}
	return host.client.Shutdown(ctx)
}
