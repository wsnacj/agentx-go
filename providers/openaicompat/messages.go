package openaicompat

import (
	"strings"

	llm "github.com/wsnacj/agentx-go/components/llm"
)

func marshalMessages(profile profile, system string, messages llm.Conversation) []map[string]any {
	out := make([]map[string]any, 0, len(messages)+1)
	if system != "" {
		out = append(out, map[string]any{"role": "system", "content": system})
	}
	lastRole := ""
	for _, message := range messages {
		role := strings.ToLower(strings.TrimSpace(message.Role))
		if role == "" {
			role = "assistant"
		}
		if profile.requiresAssistantAfterToolResult && lastRole == "tool" && role == "user" {
			out = append(out, map[string]any{"role": "assistant", "content": "I have processed the tool results."})
			lastRole = "assistant"
		}
		callID := strings.TrimSpace(message.ToolCallID)
		if role == "tool" && profile.requiresToolCallIDForToolRole && callID == "" {
			role = "assistant"
		}
		row := map[string]any{"role": role, "content": message.Content}
		if callID != "" {
			row["tool_call_id"] = callID
		}
		if role == "tool" && profile.requiresToolResultName {
			if name := strings.TrimSpace(message.ToolName); name != "" {
				row["name"] = name
			}
		}
		if len(message.ToolCalls) > 0 {
			calls := make([]map[string]any, 0, len(message.ToolCalls))
			for _, call := range message.ToolCalls {
				name, arguments := strings.TrimSpace(call.Name), strings.TrimSpace(call.Arguments)
				if name == "" && arguments == "" {
					continue
				}
				kind := strings.TrimSpace(call.Type)
				if kind == "" {
					kind = "function"
				}
				item := map[string]any{"type": kind, "function": map[string]any{"name": name, "arguments": arguments}}
				if id := strings.TrimSpace(call.ID); id != "" {
					item["id"] = id
				}
				calls = append(calls, item)
			}
			if len(calls) > 0 {
				row["tool_calls"] = calls
			}
		}
		out = append(out, row)
		lastRole = role
	}
	return out
}

func (c *Client) marshalVisualMessages(cfg ModelConfig, req llm.VisualRequest) []map[string]any {
	detail := "auto"
	if value, ok := req.Options["detail"].(string); ok && value != "" {
		detail = value
	}
	system := []map[string]any{}
	if req.System != "" {
		system = append(system, map[string]any{"type": "text", "text": req.System})
	}
	content := make([]map[string]any, 0, len(req.Messages)+len(req.Visual)+1)
	for _, message := range req.Messages {
		content = append(content, map[string]any{"type": "text", "text": message.Content})
	}
	for _, visual := range req.Visual {
		switch visual.Type {
		case "text":
			content = append(content, map[string]any{"type": "text", "text": visual.Text})
		case "video_url":
			url := firstNonEmpty(visual.VideoURL, visual.ImageURL, visual.DataURI)
			if cfg.Capability.LocalFiles && url == "" {
				url = visual.Text
			}
			item := map[string]any{"type": "video_url", "video_url": map[string]any{"url": url}}
			if visual.FPS != nil {
				item["fps"] = *visual.FPS
			}
			content = append(content, item)
		default:
			url := firstNonEmpty(visual.ImageURL, visual.DataURI)
			if cfg.Capability.LocalFiles && url == "" {
				url = visual.Text
			}
			if cfg.Capability.LocalFiles && c.resolveMedia != nil {
				if resolved, err := c.resolveMedia(url); err == nil {
					url = resolved
				}
			}
			content = append(content, map[string]any{"type": "image_url", "image_url": map[string]any{"url": url, "detail": detail}})
		}
	}
	return []map[string]any{{"role": "system", "content": system}, {"role": "user", "content": content}}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
