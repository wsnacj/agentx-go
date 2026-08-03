package ark

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/wsnacj/agentx-go/providers/ark/types"
)

// StreamEventHandler handles a stream event with its optional decoded payload.
type StreamEventHandler func(event types.Event, payload any) error

// StreamToolRunOptions controls auto tool execution after a stream completes.
type StreamToolRunOptions struct {
	OnEvent  StreamEventHandler
	Followup ToolFollowupOptions
}

// StreamToolRunResult captures tool execution results after streaming.
type StreamToolRunResult struct {
	Calls    []FunctionCall
	Results  []FunctionResult
	Followup *types.Response
}

type streamCallState struct {
	call FunctionCall
	args strings.Builder
}

// StreamFunctionCalls captures aggregated function calls and response metadata from a stream.
type StreamFunctionCalls struct {
	ResponseID string
	Model      string
	Calls      []FunctionCall
}

// CollectFunctionCallsFromStream aggregates function call arguments from stream events.
func CollectFunctionCallsFromStream(ctx context.Context, stream *types.StreamResult, onEvent StreamEventHandler) (StreamFunctionCalls, error) {
	if stream == nil {
		return StreamFunctionCalls{}, fmt.Errorf("ark stream: stream is nil")
	}

	calls := map[string]*streamCallState{}
	order := make([]string, 0, 4)
	var responseID string
	var model string

	getState := func(itemID string) *streamCallState {
		if itemID == "" {
			return nil
		}
		if state, ok := calls[itemID]; ok {
			return state
		}
		state := &streamCallState{call: FunctionCall{ID: itemID}}
		calls[itemID] = state
		order = append(order, itemID)
		return state
	}

	streamCh := stream.Ch
	errCh := stream.Err
	for streamCh != nil || errCh != nil {
		select {
		case <-ctx.Done():
			return StreamFunctionCalls{}, ctx.Err()
		case ev, ok := <-streamCh:
			if !ok {
				streamCh = nil
				continue
			}
			var payload any
			switch ev.Type {
			case types.EventTypeResponseCreated, types.EventTypeResponseInProgress, types.EventTypeResponseCompleted, types.EventTypeResponseFailed, types.EventTypeResponseIncomplete:
				var out types.ResponseEvent
				if err := json.Unmarshal(ev.Raw, &out); err != nil {
					return StreamFunctionCalls{}, err
				}
				payload = out
				if out.Response != nil {
					if out.Response.ID != "" {
						responseID = out.Response.ID
					}
					if out.Response.Model != "" {
						model = out.Response.Model
					}
				}
			case types.EventTypeResponseOutputItemAdded, types.EventTypeResponseOutputItemDone:
				var out types.OutputItemEvent
				if err := json.Unmarshal(ev.Raw, &out); err != nil {
					return StreamFunctionCalls{}, err
				}
				payload = out
				if out.Item != nil && out.Item.Type == "function_call" {
					state := getState(out.Item.ID)
					if state != nil {
						state.call.Name = out.Item.Name
						state.call.CallID = out.Item.CallID
						if out.Item.Arguments != "" {
							state.call.Arguments = out.Item.Arguments
						}
					}
				}
			case types.EventTypeResponseFunctionCallArgsDelta:
				var out types.FunctionCallArgumentsDeltaEvent
				if err := json.Unmarshal(ev.Raw, &out); err != nil {
					return StreamFunctionCalls{}, err
				}
				payload = out
				if out.ItemID != "" && out.Delta != "" {
					if state := getState(out.ItemID); state != nil {
						state.args.WriteString(out.Delta)
					}
				}
			case types.EventTypeResponseFunctionCallArgsDone:
				var out types.FunctionCallArgumentsDoneEvent
				if err := json.Unmarshal(ev.Raw, &out); err != nil {
					return StreamFunctionCalls{}, err
				}
				payload = out
				if out.ItemID != "" {
					if state := getState(out.ItemID); state != nil {
						if out.Arguments != "" {
							state.call.Arguments = out.Arguments
						}
					}
				}
			case types.EventTypeError:
				var out types.ErrorEvent
				if err := json.Unmarshal(ev.Raw, &out); err != nil {
					return StreamFunctionCalls{}, err
				}
				payload = out
				if out.Message != "" {
					return StreamFunctionCalls{}, errors.New(out.Message)
				}
			default:
				payload = nil
			}

			if onEvent != nil {
				if err := onEvent(ev, payload); err != nil {
					return StreamFunctionCalls{}, err
				}
			}
		case err, ok := <-errCh:
			if !ok {
				errCh = nil
				continue
			}
			if err != nil {
				return StreamFunctionCalls{}, err
			}
		}
	}

	results := make([]FunctionCall, 0, len(order))
	for _, id := range order {
		state := calls[id]
		if state == nil {
			continue
		}
		if state.call.Arguments == "" && state.args.Len() > 0 {
			state.call.Arguments = state.args.String()
		}
		results = append(results, state.call)
	}
	return StreamFunctionCalls{ResponseID: responseID, Model: model, Calls: results}, nil
}

// ExecuteFunctionCalls runs tool executor for aggregated calls.
func ExecuteFunctionCalls(ctx context.Context, calls []FunctionCall, exec ToolExecutor) ([]FunctionResult, error) {
	if exec == nil {
		return nil, nil
	}
	results := make([]FunctionResult, 0, len(calls))
	var errs []error
	for _, call := range calls {
		output, err := exec(ctx, call)
		if err != nil {
			errs = append(errs, err)
		}
		results = append(results, FunctionResult{Call: call, Output: output, Err: err})
	}
	return results, errors.Join(errs...)
}

// StreamResponseWithTools runs a streaming request and auto-executes tools after the stream ends.
func (c *Client) StreamResponseWithTools(ctx context.Context, req types.ResponseRequest, exec ToolExecutor, opts StreamToolRunOptions) (*StreamToolRunResult, error) {
	stream, err := c.StreamResponse(ctx, req)
	if err != nil {
		return nil, err
	}
	defer stream.Cancel()

	streamCalls, err := CollectFunctionCallsFromStream(ctx, stream, opts.OnEvent)
	if err != nil {
		return nil, err
	}
	if len(streamCalls.Calls) == 0 {
		return &StreamToolRunResult{Calls: nil, Results: nil, Followup: nil}, nil
	}
	if exec == nil {
		return &StreamToolRunResult{Calls: streamCalls.Calls, Results: nil, Followup: nil}, errors.New("ark stream: tool executor is required")
	}

	results, execErr := ExecuteFunctionCalls(ctx, streamCalls.Calls, exec)
	if execErr != nil {
		return &StreamToolRunResult{Calls: streamCalls.Calls, Results: results, Followup: nil}, execErr
	}

	followupOpts := opts.Followup
	if len(followupOpts.Tools) == 0 {
		followupOpts.Tools = req.Tools
	}
	if followupOpts.Thinking == nil {
		followupOpts.Thinking = req.Thinking
	}

	respModel := req.Model
	if respModel == "" {
		respModel = streamCalls.Model
	}
	respID := streamCalls.ResponseID
	if respID == "" && req.PreviousResponseID != nil {
		respID = *req.PreviousResponseID
	}
	if respID == "" {
		return &StreamToolRunResult{Calls: streamCalls.Calls, Results: results, Followup: nil}, errors.New("ark stream: missing response id for follow-up")
	}
	resp, err := c.ContinueWithToolOutputs(ctx, &types.Response{ID: respID, Model: respModel, RequestModel: req.Model}, results, followupOpts)
	if err != nil {
		return &StreamToolRunResult{Calls: streamCalls.Calls, Results: results, Followup: nil}, err
	}
	return &StreamToolRunResult{Calls: streamCalls.Calls, Results: results, Followup: resp}, nil
}
