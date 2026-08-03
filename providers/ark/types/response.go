package types

import "encoding/json"

// Response represents a Responses API response object.
type Response struct {
	CreatedAt         int64              `json:"created_at,omitempty"`
	ID                string             `json:"id,omitempty"`
	MaxOutput         *int               `json:"max_output_tokens,omitempty"`
	Model             string             `json:"model,omitempty"`
	RequestModel      string             `json:"-"`
	Object            string             `json:"object,omitempty"`
	Output            []OutputItem       `json:"output,omitempty"`
	ServiceTier       string             `json:"service_tier,omitempty"`
	Status            string             `json:"status,omitempty"`
	Usage             *Usage             `json:"usage,omitempty"`
	Error             *APIError          `json:"error,omitempty"`
	Caching           *CachingConfig     `json:"caching,omitempty"`
	Store             *bool              `json:"store,omitempty"`
	ExpireAt          *int64             `json:"expire_at,omitempty"`
	Thinking          *ThinkingConfig    `json:"thinking,omitempty"`
	Text              *TextFormat        `json:"text,omitempty"`
	Tools             []Tool             `json:"tools,omitempty"`
	ContextManagement *ContextManagement `json:"context_management,omitempty"`
}

// OutputItem represents an output item in a response.
type OutputItem struct {
	ID        string              `json:"id,omitempty"`
	Type      string              `json:"type,omitempty"`
	Role      string              `json:"role,omitempty"`
	Content   []OutputContentItem `json:"content,omitempty"`
	Summary   []SummaryText       `json:"summary,omitempty"`
	Name      string              `json:"name,omitempty"`
	Arguments string              `json:"arguments,omitempty"`
	CallID    string              `json:"call_id,omitempty"`
	Status    string              `json:"status,omitempty"`
	Extra     map[string]any      `json:"-"`
}

func (o *OutputItem) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	var extra map[string]any
	setExtra := func(key string, value json.RawMessage) {
		if extra == nil {
			extra = make(map[string]any)
		}
		var decoded any
		if err := json.Unmarshal(value, &decoded); err == nil {
			extra[key] = decoded
		}
	}

	for key, value := range raw {
		switch key {
		case "id":
			_ = json.Unmarshal(value, &o.ID)
		case "type":
			_ = json.Unmarshal(value, &o.Type)
		case "role":
			_ = json.Unmarshal(value, &o.Role)
		case "content":
			var parsed []OutputContentItem
			if err := json.Unmarshal(value, &parsed); err == nil {
				o.Content = parsed
			} else {
				setExtra(key, value)
			}
		case "summary":
			_ = json.Unmarshal(value, &o.Summary)
		case "name":
			_ = json.Unmarshal(value, &o.Name)
		case "arguments":
			_ = json.Unmarshal(value, &o.Arguments)
		case "call_id":
			_ = json.Unmarshal(value, &o.CallID)
		case "status":
			_ = json.Unmarshal(value, &o.Status)
		default:
			setExtra(key, value)
		}
	}

	if len(extra) > 0 {
		o.Extra = extra
	}
	return nil
}

// OutputContentItem represents a text output piece.
type OutputContentItem struct {
	Type        string             `json:"type,omitempty"`
	Text        string             `json:"text,omitempty"`
	Annotations []OutputAnnotation `json:"annotations,omitempty"`
}

// OutputAnnotation captures provider annotations attached to assistant output.
// URL citations are emitted by the Ark Responses API web_search tool.
type OutputAnnotation struct {
	Type        string `json:"type,omitempty"`
	Title       string `json:"title,omitempty"`
	URL         string `json:"url,omitempty"`
	LogoURL     string `json:"logo_url,omitempty"`
	SiteName    string `json:"site_name,omitempty"`
	PublishTime string `json:"publish_time,omitempty"`
	Summary     string `json:"summary,omitempty"`
	StartIndex  *int   `json:"start_index,omitempty"`
	EndIndex    *int   `json:"end_index,omitempty"`
}

// SummaryText represents reasoning summary content.
type SummaryText struct {
	Type string `json:"type,omitempty"`
	Text string `json:"text,omitempty"`
}

// Usage captures token usage.
type Usage struct {
	InputTokens         int            `json:"input_tokens,omitempty"`
	OutputTokens        int            `json:"output_tokens,omitempty"`
	TotalTokens         int            `json:"total_tokens,omitempty"`
	InputTokensDetails  map[string]any `json:"input_tokens_details,omitempty"`
	OutputTokensDetails map[string]any `json:"output_tokens_details,omitempty"`
	ToolUsage           map[string]any `json:"tool_usage,omitempty"`
}

// APIError represents an error payload from Ark Responses API.
type APIError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Param   string `json:"param,omitempty"`
}
