package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	agentx "github.com/wsnacj/agentx-go"
	llm "github.com/wsnacj/agentx-go/components/llm"
	"github.com/wsnacj/agentx-go/providers/openaicompat"
	"github.com/wsnacj/agentx-go/runtime/execution"
	"github.com/wsnacj/agentx-go/runtime/hostkit"
	"github.com/wsnacj/agentx-go/runtime/toolloop"
)

const (
	envEnable  = "AGENTX_LIVE_PROVIDER_SMOKE_ENABLE"
	envBaseURL = "AGENTX_LIVE_PROVIDER_SMOKE_BASE_URL"
	envAPIKey  = "AGENTX_LIVE_PROVIDER_SMOKE_API_KEY"
	envModel   = "AGENTX_LIVE_PROVIDER_SMOKE_MODEL"
	envPrompt  = "AGENTX_LIVE_PROVIDER_SMOKE_PROMPT"
	envExpect  = "AGENTX_LIVE_PROVIDER_SMOKE_EXPECT"
	envTimeout = "AGENTX_LIVE_PROVIDER_SMOKE_TIMEOUT"
	envMode    = "AGENTX_LIVE_PROVIDER_SMOKE_MODE"

	defaultMarker  = "AGENTX_LIVE_PROVIDER_SMOKE_OK"
	modeChat       = "chat"
	modeToolLoop   = "tool-loop"
	markerToolName = "agentx_smoke_marker"
)

var errLiveDisabled = errors.New("live provider smoke is disabled")

type config struct {
	BaseURL string
	APIKey  string
	Model   string
	Prompt  string
	Expect  string
	Mode    string
	Timeout time.Duration
}

type report struct {
	Status    string `json:"status"`
	RunID     string `json:"run_id,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	Matched   bool   `json:"matched"`
	Mode      string `json:"mode"`
	ToolCalls int64  `json:"tool_calls,omitempty"`
	Reply     string `json:"reply,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

func loadConfig(lookup func(string) string) (config, error) {
	if !truthy(lookup(envEnable)) {
		return config{}, errLiveDisabled
	}
	cfg := config{
		BaseURL: strings.TrimSpace(lookup(envBaseURL)),
		APIKey:  strings.TrimSpace(lookup(envAPIKey)),
		Model:   strings.TrimSpace(lookup(envModel)),
		Prompt:  strings.TrimSpace(lookup(envPrompt)),
		Expect:  strings.TrimSpace(lookup(envExpect)),
		Mode:    strings.ToLower(strings.TrimSpace(lookup(envMode))),
		Timeout: 45 * time.Second,
	}
	if cfg.Mode == "" {
		cfg.Mode = modeChat
	}
	if cfg.Mode != modeChat && cfg.Mode != modeToolLoop {
		return config{}, fmt.Errorf("%s must be %q or %q", envMode, modeChat, modeToolLoop)
	}
	if cfg.Prompt == "" {
		if cfg.Mode == modeToolLoop {
			cfg.Prompt = "这是 AgentX 真实 provider tool-loop smoke。必须调用 " + markerToolName + " 工具，参数 value 必须为 " + defaultMarker + "，不要直接回答。"
		} else {
			cfg.Prompt = "这是 AgentX 真实 provider smoke。请只输出 " + defaultMarker
		}
	}
	if cfg.Expect == "" {
		cfg.Expect = defaultMarker
	}
	if raw := strings.TrimSpace(lookup(envTimeout)); raw != "" {
		value, err := time.ParseDuration(raw)
		if err != nil || value <= 0 {
			return config{}, fmt.Errorf("%s must be a positive duration", envTimeout)
		}
		cfg.Timeout = value
	}
	missing := make([]string, 0, 3)
	if cfg.BaseURL == "" {
		missing = append(missing, envBaseURL)
	}
	if cfg.APIKey == "" {
		missing = append(missing, envAPIKey)
	}
	if cfg.Model == "" {
		missing = append(missing, envModel)
	}
	if len(missing) != 0 {
		return config{}, fmt.Errorf("live provider smoke missing required environment: %s", strings.Join(missing, ", "))
	}
	return cfg, nil
}

func run(parent context.Context, cfg config) (report, error) {
	provider, err := openaicompat.New(openaicompat.Config{
		Name:    "agentx-live-provider-smoke",
		BaseURL: cfg.BaseURL,
		Timeout: cfg.Timeout,
		Authorize: func(_ context.Context, header http.Header) error {
			header.Set("Authorization", "Bearer "+cfg.APIKey)
			return nil
		},
	})
	if err != nil {
		return report{}, err
	}
	var toolCalls atomic.Int64
	var client *agentx.Client
	if cfg.Mode == modeToolLoop {
		client, err = newToolLoopClient(provider, cfg, &toolCalls)
	} else {
		client, err = newChatClient(provider, cfg)
	}
	if err != nil {
		return report{}, err
	}

	ctx, cancel := context.WithTimeout(parent, cfg.Timeout)
	defer cancel()
	result, runErr := client.Run(ctx, agentx.RunRequest{
		Input:     cfg.Prompt,
		SessionID: "live-provider-smoke-" + cfg.Mode,
	})
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	shutdownErr := client.Shutdown(shutdownCtx)
	if runErr != nil {
		return report{}, runErr
	}
	if shutdownErr != nil {
		return report{}, shutdownErr
	}
	matched := strings.Contains(strings.ToUpper(result.Reply), strings.ToUpper(cfg.Expect))
	value := report{
		Status:    result.Status,
		RunID:     result.RunID,
		SessionID: result.SessionID,
		Matched:   matched,
		Mode:      cfg.Mode,
		ToolCalls: toolCalls.Load(),
		Reply:     result.Reply,
	}
	if result.Status != "completed" {
		return value, fmt.Errorf("live provider smoke status = %q", result.Status)
	}
	if cfg.Mode == modeToolLoop && value.ToolCalls != 1 {
		return value, fmt.Errorf("live provider tool-loop expected exactly one tool call, got %d", value.ToolCalls)
	}
	if !matched {
		return value, fmt.Errorf("live provider reply did not contain expected marker")
	}
	return value, nil
}

func newChatClient(provider *openaicompat.Client, cfg config) (*agentx.Client, error) {
	return hostkit.NewChatClient(hostkit.ChatClientConfig{
		Model: cfg.Model,
		RequestModel: func(ctx context.Context, request llm.ChatRequest) (llm.ChatResponse, error) {
			response, _, requestErr := provider.Chat(ctx, openaicompat.ModelConfig{
				Name:  cfg.Model,
				Model: cfg.Model,
			}, request)
			if requestErr != nil {
				return llm.ChatResponse{}, requestErr
			}
			if response == nil {
				return llm.ChatResponse{}, errors.New("live provider returned nil response")
			}
			return *response, nil
		},
	})
}

func newToolLoopClient(provider *openaicompat.Client, cfg config, toolCalls *atomic.Int64) (*agentx.Client, error) {
	return hostkit.NewModelToolClient(hostkit.ModelToolClientConfig{
		MaxRounds: 1,
		BuildRound: func(_ context.Context, request execution.Request) (hostkit.ModelToolRoundConfig, error) {
			return hostkit.ModelToolRoundConfig{
				RequestModel: func(ctx context.Context, _ toolloop.RoundExecutionInput) (hostkit.ModelResult, error) {
					response, _, err := provider.Chat(ctx, openaicompat.ModelConfig{Name: cfg.Model, Model: cfg.Model}, llm.ChatRequest{
						Model:    cfg.Model,
						Messages: llm.Conversation{{Role: "user", Content: request.Input}},
						Tools: []llm.Tool{{Type: "function", Function: llm.Function{
							Name: markerToolName, Description: "Return the exact smoke marker provided by the caller.",
							Parameters: map[string]any{
								"type": "object", "properties": map[string]any{"value": map[string]any{"type": "string"}}, "required": []string{"value"},
							},
						}}},
						ToolChoice: &llm.ToolChoice{Type: "function", Function: &llm.ToolChoiceFunction{Name: markerToolName}},
					})
					if err != nil {
						return hostkit.ModelResult{}, err
					}
					if response == nil {
						return hostkit.ModelResult{}, errors.New("live provider returned nil response")
					}
					return hostkit.ModelResult{Response: *response, Model: cfg.Model}, nil
				},
				ExecuteTools: func(_ context.Context, exchange hostkit.ModelToolRoundExchange) (hostkit.ToolResult, error) {
					if len(exchange.Model.Response.Calls) != 1 || exchange.Model.Response.Calls[0].Name != markerToolName {
						return hostkit.ToolResult{}, fmt.Errorf("live provider did not call %s exactly once", markerToolName)
					}
					var arguments struct {
						Value string `json:"value"`
					}
					if err := json.Unmarshal([]byte(exchange.Model.Response.Calls[0].Arguments), &arguments); err != nil {
						return hostkit.ToolResult{}, fmt.Errorf("decode live tool arguments: %w", err)
					}
					if arguments.Value != cfg.Expect {
						return hostkit.ToolResult{}, fmt.Errorf("live provider tool marker mismatch")
					}
					toolCalls.Add(1)
					return hostkit.ToolResult{DirectAnswer: &hostkit.ToolDirectAnswer{
						Reply: cfg.Expect, Source: markerToolName, Reason: "host verified pure smoke tool arguments",
					}}, nil
				},
			}, nil
		},
	})
}

func truthy(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func writeJSON(value report) {
	encoded, err := json.Marshal(value)
	if err != nil {
		fmt.Fprintln(os.Stderr, "encode live provider report:", err)
		return
	}
	fmt.Println(string(encoded))
}

func main() {
	cfg, err := loadConfig(os.Getenv)
	if errors.Is(err, errLiveDisabled) {
		writeJSON(report{Status: "skipped", Reason: "explicit_opt_in_required"})
		return
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	value, err := run(context.Background(), cfg)
	if err != nil {
		if value.Status == "" {
			value.Status = "failed"
		}
		writeJSON(value)
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	writeJSON(value)
}
