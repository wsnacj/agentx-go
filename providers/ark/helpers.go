package ark

import (
	"context"
	"errors"
	"strings"

	"github.com/wsnacj/agentx-go/providers/ark/types"
)

// FunctionCall represents a function call extracted from a response.
type FunctionCall struct {
	Name      string
	Arguments string
	CallID    string
	ID        string
}

// FunctionResult captures execution output.
type FunctionResult struct {
	Call   FunctionCall
	Output string
	Err    error
}

// ToolFollowupOptions controls the follow-up request after tool execution.
type ToolFollowupOptions struct {
	Model    string
	Tools    []types.Tool
	Thinking *types.ThinkingConfig
	Text     *types.TextFormat
	Caching  *types.CachingConfig
	Store    *bool
	Extra    map[string]any
}

// ToolExecutor executes a function call and returns output.
type ToolExecutor func(ctx context.Context, call FunctionCall) (string, error)

// ToolResult captures tool output items from a response.
type ToolResult struct {
	ID      string
	Type    string
	Name    string
	CallID  string
	Status  string
	Content []types.OutputContentItem
	Extra   map[string]any
}

// ExtractFunctionCalls returns function calls from a response.
func ExtractFunctionCalls(resp *types.Response) []FunctionCall {
	if resp == nil {
		return nil
	}
	calls := []FunctionCall{}
	for _, item := range resp.Output {
		if item.Type != "function_call" {
			continue
		}
		calls = append(calls, FunctionCall{
			Name:      item.Name,
			Arguments: item.Arguments,
			CallID:    item.CallID,
			ID:        item.ID,
		})
	}
	return calls
}

// ExtractToolResults returns non-message, non-reasoning output items.
func ExtractToolResults(resp *types.Response) []ToolResult {
	if resp == nil {
		return nil
	}
	results := []ToolResult{}
	for _, item := range resp.Output {
		if item.Type == "" || item.Type == "message" || item.Type == "reasoning" || item.Type == "function_call" {
			continue
		}
		results = append(results, ToolResult{
			ID:      item.ID,
			Type:    item.Type,
			Name:    item.Name,
			CallID:  item.CallID,
			Status:  item.Status,
			Content: item.Content,
			Extra:   item.Extra,
		})
	}
	return results
}

// ExtractToolResultsByType filters tool results by type.
func ExtractToolResultsByType(resp *types.Response, toolType string) []ToolResult {
	if toolType == "" {
		return ExtractToolResults(resp)
	}
	results := ExtractToolResults(resp)
	filtered := make([]ToolResult, 0, len(results))
	for _, result := range results {
		if result.Type == toolType {
			filtered = append(filtered, result)
		}
	}
	return filtered
}

// BuildToolOutputItem builds a tool output input item.
func BuildToolOutputItem(callID string, output string) types.InputItem {
	return types.InputItem{
		Type:   "function_call_output",
		CallID: callID,
		Output: output,
	}
}

// BuildToolOutputItems converts tool results into input items.
func BuildToolOutputItems(results []FunctionResult) []types.InputItem {
	items := make([]types.InputItem, 0, len(results))
	for _, result := range results {
		output := result.Output
		if output == "" && result.Err != nil {
			output = result.Err.Error()
		}
		items = append(items, BuildToolOutputItem(result.Call.CallID, output))
	}
	return items
}

// BuildToolOutputRequest builds a follow-up request using tool results.
func BuildToolOutputRequest(prev *types.Response, results []FunctionResult, opts ToolFollowupOptions) (types.ResponseRequest, error) {
	if prev == nil || prev.ID == "" {
		return types.ResponseRequest{}, errors.New("ark: previous response id is required")
	}
	model := opts.Model
	if model == "" {
		if prev.RequestModel != "" {
			model = prev.RequestModel
		} else {
			model = prev.Model
		}
	}
	items := BuildToolOutputItems(results)
	req := types.ResponseRequest{
		Model:              model,
		Input:              types.NewInputItems(items...),
		PreviousResponseID: &prev.ID,
		Thinking:           opts.Thinking,
		Text:               opts.Text,
		Caching:            opts.Caching,
		Store:              opts.Store,
		Tools:              opts.Tools,
		Extra:              opts.Extra,
	}
	return req, nil
}

// HandleFunctionCalls executes calls and returns results.
func HandleFunctionCalls(ctx context.Context, resp *types.Response, exec ToolExecutor) ([]FunctionResult, error) {
	if exec == nil {
		return nil, nil
	}
	calls := ExtractFunctionCalls(resp)
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

// ResponseText aggregates assistant output text from a response.
func ResponseText(resp *types.Response) string {
	if resp == nil {
		return ""
	}
	var builder strings.Builder
	for _, item := range resp.Output {
		if item.Type != "message" {
			continue
		}
		for _, content := range item.Content {
			if content.Text == "" {
				continue
			}
			builder.WriteString(content.Text)
		}
	}
	return builder.String()
}

// URLCitations returns de-duplicated URL citations attached to message output.
func URLCitations(resp *types.Response) []types.OutputAnnotation {
	if resp == nil {
		return nil
	}
	seen := map[string]bool{}
	out := []types.OutputAnnotation{}
	for _, item := range resp.Output {
		if item.Type != "message" {
			continue
		}
		for _, content := range item.Content {
			for _, annotation := range content.Annotations {
				if annotation.Type != "url_citation" {
					continue
				}
				url := strings.TrimSpace(annotation.URL)
				if url == "" || seen[url] {
					continue
				}
				seen[url] = true
				annotation.URL = url
				out = append(out, annotation)
			}
		}
	}
	return out
}

// ReasoningText aggregates reasoning summary text from a response.
func ReasoningText(resp *types.Response) string {
	if resp == nil {
		return ""
	}
	var builder strings.Builder
	for _, item := range resp.Output {
		if item.Type != "reasoning" {
			continue
		}
		for _, summary := range item.Summary {
			if summary.Text == "" {
				continue
			}
			builder.WriteString(summary.Text)
		}
	}
	return builder.String()
}
