package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	toolcontract "github.com/wsnacj/agentx-go/components/tool"
	agentxtoolerrors "github.com/wsnacj/agentx-go/runtime/toolerrors"
)

// Request is a normalized model-tool request. Arguments is a defensive copy.
// Concrete identity, persistence, authorization and scheduling remain Host-owned.
type Request struct {
	Name      string
	Action    string
	Arguments map[string]any
}

// Backend executes Task, Subagent and bounded Agent Step requests against a
// Host-owned lifecycle implementation.
type Backend interface {
	ExecuteTask(context.Context, Request) (string, error)
	ExecuteSubagent(context.Context, Request) (string, error)
	ExecuteAgentStep(context.Context, Request) (string, error)
}

// BackendFuncs adapts private Host functions without exporting Host types.
type BackendFuncs struct {
	TaskFunc      func(context.Context, Request) (string, error)
	SubagentFunc  func(context.Context, Request) (string, error)
	AgentStepFunc func(context.Context, Request) (string, error)
}

func (b BackendFuncs) ExecuteTask(ctx context.Context, request Request) (string, error) {
	return callBackend(ctx, request, b.TaskFunc)
}

func (b BackendFuncs) ExecuteSubagent(ctx context.Context, request Request) (string, error) {
	return callBackend(ctx, request, b.SubagentFunc)
}

func (b BackendFuncs) ExecuteAgentStep(ctx context.Context, request Request) (string, error) {
	return callBackend(ctx, request, b.AgentStepFunc)
}

func callBackend(ctx context.Context, request Request, fn func(context.Context, Request) (string, error)) (string, error) {
	if fn == nil {
		return "", fmt.Errorf("%s: host backend is unavailable", request.Name)
	}
	return fn(ctx, request)
}

// Options controls which model-facing tools are registered. An empty Enabled
// list registers the entire cohort, matching the existing AgentX Host behavior.
type Options struct {
	Enabled []string
	Backend Backend
}

// Register installs the portable Task/Subagent tool cohort over a Host Backend.
func Register(reg toolcontract.Registrar, opts Options) {
	if reg == nil || opts.Backend == nil {
		return
	}
	allowed := enabledSet(opts.Enabled)
	for _, item := range []struct {
		name string
		def  toolcontract.Definition
	}{
		{TasksSpawnName, TasksSpawnDefinition()},
		{TasksWaitName, TasksWaitDefinition()},
		{TasksRunName, TasksRunDefinition()},
		{TasksCancelName, TasksCancelDefinition()},
		{TasksReplayName, TasksReplayDefinition()},
		{TasksCollectName, TasksCollectDefinition()},
		{TasksDeadletterListName, TasksDeadletterListDefinition()},
	} {
		if enabled(allowed, item.name) {
			reg.Register(item.def, NewTaskHandler(item.name, opts.Backend))
		}
	}
	if enabled(allowed, SubagentsName) {
		reg.Register(SubagentsDefinition(), NewSubagentsHandler(opts.Backend))
	}
	if enabled(allowed, AgentStepName) {
		reg.Register(AgentStepDefinition(), NewAgentStepHandler(opts.Backend))
	}
}

// NewTaskHandler constructs deterministic argument parsing and Host dispatch
// for one direct task lifecycle tool.
func NewTaskHandler(name string, backend Backend) toolcontract.Handler {
	return func(ctx context.Context, call toolcontract.Call) (toolcontract.Result, error) {
		if backend == nil {
			return "", fmt.Errorf("%s: host backend is unavailable", name)
		}
		params, err := decodeArguments(name, call.Arguments)
		if err != nil {
			return "", err
		}
		return backend.ExecuteTask(ctx, Request{Name: name, Arguments: cloneArguments(params)})
	}
}

// NewSubagentsHandler owns the model-facing action contract and routes it to a
// Host lifecycle backend. It deliberately does not own Store or Scheduler policy.
func NewSubagentsHandler(backend Backend) toolcontract.Handler {
	return func(ctx context.Context, call toolcontract.Call) (toolcontract.Result, error) {
		if backend == nil {
			return "", fmt.Errorf("%s: host backend is unavailable", SubagentsName)
		}
		params, err := decodeArguments(SubagentsName, call.Arguments)
		if err != nil {
			return "", err
		}
		action := strings.ToLower(strings.TrimSpace(readString(params, "action")))
		if action == "" {
			return "", agentxtoolerrors.NewMissingRequiredToolArgumentError(SubagentsName, []string{"action"}, "subagents: action is required")
		}
		switch action {
		case "list", "status", "cancel", "replay", "steer":
		case "run":
			normalizeRunInstruction(params)
			if err := validateRunInstruction(params, -1); err != nil {
				return "", err
			}
		case "fanout":
			if err := validateFanout(params); err != nil {
				return "", err
			}
		default:
			return "", agentxtoolerrors.NewInvalidToolArgumentError(SubagentsName, []string{"action"}, fmt.Sprintf("subagents: unsupported action %q", action))
		}
		return backend.ExecuteSubagent(ctx, Request{Name: SubagentsName, Action: action, Arguments: cloneArguments(params)})
	}
}

// NewAgentStepHandler parses one bounded child step and hands it to the Host.
func NewAgentStepHandler(backend Backend) toolcontract.Handler {
	return func(ctx context.Context, call toolcontract.Call) (toolcontract.Result, error) {
		if backend == nil {
			return "", fmt.Errorf("%s: host backend is unavailable", AgentStepName)
		}
		params, err := decodeArguments(AgentStepName, call.Arguments)
		if err != nil {
			return "", err
		}
		return backend.ExecuteAgentStep(ctx, Request{Name: AgentStepName, Arguments: cloneArguments(params)})
	}
}

func validateFanout(params map[string]any) error {
	raw, ok := params["items"]
	if !ok {
		return agentxtoolerrors.NewMissingRequiredToolArgumentError(SubagentsName, []string{"items"}, "subagents: items is required for action=fanout")
	}
	items, ok := raw.([]any)
	if !ok {
		return agentxtoolerrors.NewInvalidToolArgumentError(SubagentsName, []string{"items"}, "subagents: items must be an array")
	}
	if len(items) == 0 {
		return agentxtoolerrors.NewInvalidToolArgumentError(SubagentsName, []string{"items"}, "subagents: items must contain at least one entry")
	}
	if _, present := params["expected_count"]; present {
		expected := readInt(params, "expected_count")
		if expected <= 0 {
			return agentxtoolerrors.NewInvalidToolArgumentError(SubagentsName, []string{"expected_count"}, "subagents: expected_count must be a positive integer for action=fanout")
		}
		if len(items) != expected {
			return agentxtoolerrors.NewInvalidToolArgumentError(SubagentsName, []string{"items", "expected_count"}, fmt.Sprintf("subagents: expected_count=%d but items has %d entries", expected, len(items)))
		}
	}
	for idx, rawItem := range items {
		item, ok := rawItem.(map[string]any)
		if !ok {
			return fmt.Errorf("subagents: items[%d] must be an object", idx)
		}
		normalizeRunInstruction(item)
		if err := validateRunInstruction(item, idx); err != nil {
			return err
		}
	}
	return nil
}

func normalizeRunInstruction(params map[string]any) {
	if readString(params, "seed_message") != "" {
		return
	}
	if message := readString(params, "message"); message != "" {
		params["seed_message"] = message
		if readString(params, "seed_role") == "" {
			params["seed_role"] = "user"
		}
	}
}

func validateRunInstruction(params map[string]any, idx int) error {
	if readString(params, "seed_message") != "" || readString(params, "session_id") != "" {
		return nil
	}
	if idx >= 0 {
		return agentxtoolerrors.NewMissingRequiredToolArgumentError(SubagentsName,
			[]string{fmt.Sprintf("items[%d].message", idx), fmt.Sprintf("items[%d].seed_message", idx)},
			fmt.Sprintf("subagents: items[%d] requires message or seed_message for action=fanout", idx))
	}
	return agentxtoolerrors.NewMissingRequiredToolArgumentError(SubagentsName, []string{"message", "seed_message"}, "subagents: message or seed_message is required for action=run")
}

func decodeArguments(name, raw string) (map[string]any, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return map[string]any{}, nil
	}
	candidates := []string{trimmed}
	if stripped := stripCodeFence(trimmed); stripped != "" && stripped != trimmed {
		candidates = append(candidates, stripped)
	}
	var firstErr error
	for _, candidate := range candidates {
		out, err := decodeArgumentCandidate(candidate, 0)
		if err == nil {
			return out, nil
		}
		if firstErr == nil {
			firstErr = err
		}
	}
	return nil, agentxtoolerrors.NewInvalidJSONToolArgumentError(name, fmt.Errorf("decode tool args: %w", firstErr))
}

func decodeArgumentCandidate(candidate string, depth int) (map[string]any, error) {
	candidate = strings.TrimSpace(candidate)
	var out map[string]any
	if err := json.Unmarshal([]byte(candidate), &out); err == nil {
		if out == nil {
			return nil, fmt.Errorf("decode tool args: top-level JSON object is required")
		}
		return out, nil
	}
	if repaired := collapseJSONStringConcats(candidate); repaired != candidate {
		if err := json.Unmarshal([]byte(repaired), &out); err == nil && out != nil {
			return out, nil
		}
	}
	decoder := json.NewDecoder(strings.NewReader(candidate))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	switch typed := value.(type) {
	case map[string]any:
		return typed, nil
	case []any:
		if len(typed) == 1 {
			if first, ok := typed[0].(map[string]any); ok && first != nil {
				return first, nil
			}
		}
	case string:
		if depth < 1 {
			return decodeArgumentCandidate(stripCodeFence(typed), depth+1)
		}
	}
	return nil, fmt.Errorf("decode tool args: top-level JSON object is required")
}

var jsonStringConcatPattern = regexp.MustCompile(`"((?:\\.|[^"\\])*)"\s*\+\s*"((?:\\.|[^"\\])*)"`)

func collapseJSONStringConcats(raw string) string {
	for i := 0; i < 16; i++ {
		next := jsonStringConcatPattern.ReplaceAllString(raw, `"$1$2"`)
		if next == raw {
			break
		}
		raw = next
	}
	return raw
}

func stripCodeFence(raw string) string {
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, "```") || !strings.HasSuffix(raw, "```") {
		return raw
	}
	inner := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(raw, "```"), "```"))
	if newline := strings.IndexByte(inner, '\n'); newline >= 0 {
		first := strings.TrimSpace(inner[:newline])
		if first == "json" || first == "javascript" || first == "js" {
			inner = inner[newline+1:]
		}
	}
	return strings.TrimSpace(inner)
}

func cloneArguments(params map[string]any) map[string]any {
	out := make(map[string]any, len(params))
	for key, value := range params {
		out[key] = value
	}
	return out
}

func readString(params map[string]any, key string) string {
	value, _ := params[key].(string)
	return strings.TrimSpace(value)
}

func readInt(params map[string]any, key string) int {
	switch value := params[key].(type) {
	case float64:
		return int(value)
	case json.Number:
		parsed, _ := strconv.Atoi(value.String())
		return parsed
	case int:
		return value
	case int64:
		return int(value)
	case string:
		parsed, _ := strconv.Atoi(strings.TrimSpace(value))
		return parsed
	default:
		return 0
	}
}

func enabledSet(items []string) map[string]bool {
	if len(items) == 0 {
		return nil
	}
	out := make(map[string]bool, len(items))
	for _, item := range items {
		if name := strings.ToLower(strings.TrimSpace(item)); name != "" {
			out[name] = true
		}
	}
	return out
}

func enabled(allowed map[string]bool, name string) bool {
	return len(allowed) == 0 || allowed[strings.ToLower(strings.TrimSpace(name))]
}
