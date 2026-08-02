// Package scheduler provides portable scheduled-command tool coordination.
package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	toolcontract "github.com/wsnacj/agentx-go/components/tool"
	agentxtoolerrors "github.com/wsnacj/agentx-go/runtime/toolerrors"
)

// Name is the catalog name of the scheduled-command tool.
const Name = "cron"

const (
	ActionAdd    = "add"
	ActionList   = "list"
	ActionStatus = "status"
	ActionRun    = "run"
	ActionRemove = "remove"
)

// Request is a normalized scheduled command. Arguments is a defensive copy of
// the accepted JSON object so a Host can apply its product-specific policy.
type Request struct {
	Action    string
	Arguments map[string]any
}

// Backend owns concrete scheduler, task/session store, authorization,
// visibility and durable-write behavior.
type Backend interface {
	Add(context.Context, Request) (string, error)
	List(context.Context, Request) (string, error)
	Status(context.Context, Request) (string, error)
	Run(context.Context, Request) (string, error)
	Remove(context.Context, Request) (string, error)
}

// BackendFuncs adapts private Host functions to Backend.
type BackendFuncs struct {
	AddFunc    func(context.Context, Request) (string, error)
	ListFunc   func(context.Context, Request) (string, error)
	StatusFunc func(context.Context, Request) (string, error)
	RunFunc    func(context.Context, Request) (string, error)
	RemoveFunc func(context.Context, Request) (string, error)
}

func (b BackendFuncs) Add(ctx context.Context, request Request) (string, error) {
	return callBackend(ctx, request, ActionAdd, b.AddFunc)
}
func (b BackendFuncs) List(ctx context.Context, request Request) (string, error) {
	return callBackend(ctx, request, ActionList, b.ListFunc)
}
func (b BackendFuncs) Status(ctx context.Context, request Request) (string, error) {
	return callBackend(ctx, request, ActionStatus, b.StatusFunc)
}
func (b BackendFuncs) Run(ctx context.Context, request Request) (string, error) {
	return callBackend(ctx, request, ActionRun, b.RunFunc)
}
func (b BackendFuncs) Remove(ctx context.Context, request Request) (string, error) {
	return callBackend(ctx, request, ActionRemove, b.RemoveFunc)
}

func callBackend(ctx context.Context, request Request, action string, fn func(context.Context, Request) (string, error)) (string, error) {
	if fn == nil {
		return "", fmt.Errorf("%s: action=%s backend is unavailable", Name, action)
	}
	return fn(ctx, request)
}

// Register adds the scheduled-command tool when a Backend is available.
func Register(reg toolcontract.Registrar, backend Backend) {
	if reg == nil || backend == nil {
		return
	}
	reg.Register(Definition(), NewHandler(backend))
}

// NewHandler constructs deterministic action routing over the Host backend.
func NewHandler(backend Backend) toolcontract.Handler {
	return func(ctx context.Context, call toolcontract.Call) (toolcontract.Result, error) {
		params, err := decodeArguments(call.Arguments)
		if err != nil {
			return "", err
		}
		action, _ := params["action"].(string)
		action = strings.ToLower(strings.TrimSpace(action))
		if action == "" {
			return "", agentxtoolerrors.NewMissingRequiredToolArgumentError(Name, []string{"action"}, Name+": action is required")
		}
		request := Request{Action: action, Arguments: cloneArguments(params)}
		switch action {
		case ActionAdd:
			return backend.Add(ctx, request)
		case ActionList:
			return backend.List(ctx, request)
		case ActionStatus:
			return backend.Status(ctx, request)
		case ActionRun:
			return backend.Run(ctx, request)
		case ActionRemove:
			return backend.Remove(ctx, request)
		default:
			return "", agentxtoolerrors.NewInvalidToolArgumentError(Name, []string{"action"}, fmt.Sprintf("%s: unsupported action %q", Name, action))
		}
	}
}

func decodeArguments(raw string) (map[string]any, error) {
	if strings.TrimSpace(raw) == "" {
		return map[string]any{}, nil
	}
	var params map[string]any
	if err := json.Unmarshal([]byte(raw), &params); err != nil {
		return nil, agentxtoolerrors.NewInvalidJSONToolArgumentError(Name, fmt.Errorf("decode tool args: %w", err))
	}
	if params == nil {
		return nil, agentxtoolerrors.NewInvalidJSONToolArgumentError(Name, fmt.Errorf("decode tool args: top-level JSON object is required"))
	}
	return params, nil
}

func cloneArguments(params map[string]any) map[string]any {
	if len(params) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(params))
	for key, value := range params {
		out[key] = value
	}
	return out
}
