package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"sort"
	"strings"

	toolcontract "github.com/wsnacj/agentx-go/components/tool"
)

// ToolSet is one stable MCP tools/list snapshot. It implements existing Tool
// contracts instead of creating a second execution surface.
type ToolSet struct {
	client      *Client
	definitions []toolcontract.Definition
	names       map[string]bool
}

func newToolSet(client *Client, tools []Tool) *ToolSet {
	definitions := make([]toolcontract.Definition, 0, len(tools))
	names := make(map[string]bool, len(tools))
	for _, item := range tools {
		definitions = append(definitions, toolcontract.Definition{
			Type: "function",
			Function: toolcontract.Function{
				Name:         item.Name,
				Description:  item.Description,
				Parameters:   cloneAnyMap(item.InputSchema),
				OutputSchema: cloneAnyMap(item.OutputSchema),
			},
		})
		names[item.Name] = true
	}
	sort.Slice(definitions, func(i, j int) bool {
		return definitions[i].Function.Name < definitions[j].Function.Name
	})
	return &ToolSet{client: client, definitions: definitions, names: names}
}

func (s *ToolSet) Definitions() []toolcontract.Definition {
	if s == nil {
		return nil
	}
	out := make([]toolcontract.Definition, 0, len(s.definitions))
	for _, definition := range s.definitions {
		definition.Function.Parameters = cloneAnyMap(definition.Function.Parameters)
		definition.Function.OutputSchema = cloneAnyMap(definition.Function.OutputSchema)
		out = append(out, definition)
	}
	return out
}

// Execute validates a discovered Tool name and object arguments, invokes MCP,
// and returns the complete CallToolResult as JSON. isError remains a normal
// Tool result so the model can observe and self-correct as required by MCP.
func (s *ToolSet) Execute(ctx context.Context, call toolcontract.Call) (toolcontract.Result, error) {
	if s == nil || s.client == nil || !s.names[call.Name] {
		return "", &Error{Code: ErrorCodeToolUnavailable}
	}
	arguments := map[string]any{}
	if strings.TrimSpace(call.Arguments) != "" {
		decoder := json.NewDecoder(strings.NewReader(call.Arguments))
		decoder.UseNumber()
		if err := decoder.Decode(&arguments); err != nil {
			return "", &Error{Code: ErrorCodeInvalidTool, Cause: err}
		}
		var trailing any
		if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
			return "", &Error{Code: ErrorCodeInvalidTool, Cause: errors.New("multiple argument values")}
		}
	}
	result, err := s.client.CallTool(ctx, call.Name, arguments)
	if err != nil {
		return "", err
	}
	content, err := json.Marshal(result)
	if err != nil {
		return "", &Error{Code: ErrorCodeProtocol, Cause: err}
	}
	return toolcontract.Result(content), nil
}

func normalizeTool(raw Tool) (Tool, error) {
	raw.Name = strings.TrimSpace(raw.Name)
	raw.Title = strings.TrimSpace(raw.Title)
	raw.Description = strings.TrimSpace(raw.Description)
	if !validToolName(raw.Name) || raw.InputSchema == nil {
		return Tool{}, &Error{Code: ErrorCodeInvalidTool}
	}
	if kind, ok := raw.InputSchema["type"].(string); ok && kind != "object" {
		return Tool{}, &Error{Code: ErrorCodeInvalidTool, Cause: errors.New("input schema must describe an object")}
	}
	raw.InputSchema = cloneAnyMap(raw.InputSchema)
	raw.OutputSchema = cloneAnyMap(raw.OutputSchema)
	raw.Annotations = cloneAnyMap(raw.Annotations)
	raw.Execution = cloneAnyMap(raw.Execution)
	return raw, nil
}

func validToolName(raw string) bool {
	if raw == "" || len(raw) > 128 {
		return false
	}
	for _, char := range raw {
		switch {
		case char >= 'a' && char <= 'z':
		case char >= 'A' && char <= 'Z':
		case char >= '0' && char <= '9':
		case char == '-', char == '_', char == '.':
		default:
			return false
		}
	}
	return true
}

func cloneCallToolResult(result CallToolResult) CallToolResult {
	result.Meta = cloneAnyMap(result.Meta)
	result.StructuredContent = cloneAnyMap(result.StructuredContent)
	if result.Content != nil {
		content := make([]ContentBlock, len(result.Content))
		for index, block := range result.Content {
			content[index] = ContentBlock(cloneAnyMap(map[string]any(block)))
		}
		result.Content = content
	}
	return result
}

func cloneAnyMap(input map[string]any) map[string]any {
	if input == nil {
		return nil
	}
	out := make(map[string]any, len(input))
	for key, value := range input {
		out[key] = cloneAnyValue(value)
	}
	return out
}

func cloneAnyValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneAnyMap(typed)
	case []any:
		items := make([]any, len(typed))
		for index, item := range typed {
			items[index] = cloneAnyValue(item)
		}
		return items
	default:
		return value
	}
}

var _ toolcontract.DefinitionProvider = (*ToolSet)(nil)
var _ toolcontract.Executor = (*ToolSet)(nil)
