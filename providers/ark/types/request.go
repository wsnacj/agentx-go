package types

import (
	"encoding/json"
	"fmt"
)

// ResponseRequest mirrors Ark Responses API request fields.
type ResponseRequest struct {
	Model              string             `json:"model"`
	Input              InputUnion         `json:"input"`
	Instructions       *string            `json:"instructions,omitempty"`
	PreviousResponseID *string            `json:"previous_response_id,omitempty"`
	ExpireAt           *int64             `json:"expire_at,omitempty"`
	MaxOutputTokens    *int               `json:"max_output_tokens,omitempty"`
	Thinking           *ThinkingConfig    `json:"thinking,omitempty"`
	Reasoning          *ReasoningConfig   `json:"reasoning,omitempty"`
	Caching            *CachingConfig     `json:"caching,omitempty"`
	Store              *bool              `json:"store,omitempty"`
	Stream             *bool              `json:"stream,omitempty"`
	Temperature        *float64           `json:"temperature,omitempty"`
	TopP               *float64           `json:"top_p,omitempty"`
	Text               *TextFormat        `json:"text,omitempty"`
	Tools              []Tool             `json:"tools,omitempty"`
	ToolChoice         any                `json:"tool_choice,omitempty"`
	ParallelToolCalls  *bool              `json:"parallel_tool_calls,omitempty"`
	ContextManagement  *ContextManagement `json:"context_management,omitempty"`
	Extra              map[string]any     `json:"-"`
	ExtraHeaders       map[string]string  `json:"-"`
}

// InputUnion supports string input or a list of input items.
type InputUnion struct {
	Text  *string
	Items []InputItem
}

func NewInputText(text string) InputUnion {
	return InputUnion{Text: &text}
}

func NewInputItems(items ...InputItem) InputUnion {
	return InputUnion{Items: items}
}

func (u InputUnion) MarshalJSON() ([]byte, error) {
	if u.Text != nil {
		return json.Marshal(*u.Text)
	}
	if u.Items != nil {
		return json.Marshal(u.Items)
	}
	return []byte("null"), nil
}

func (u *InputUnion) UnmarshalJSON(data []byte) error {
	if u == nil {
		return fmt.Errorf("ark types: input union is nil")
	}
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		u.Text = &text
		u.Items = nil
		return nil
	}
	var items []InputItem
	if err := json.Unmarshal(data, &items); err == nil {
		u.Items = items
		u.Text = nil
		return nil
	}
	return fmt.Errorf("ark types: unsupported input union")
}

// InputItem represents a single input element in Responses API.
type InputItem struct {
	Type      string        `json:"type,omitempty"`
	Role      string        `json:"role,omitempty"`
	Content   InputContent  `json:"content,omitempty"`
	ID        string        `json:"id,omitempty"`
	Summary   []SummaryText `json:"summary,omitempty"`
	Name      string        `json:"name,omitempty"`
	Arguments string        `json:"arguments,omitempty"`
	CallID    string        `json:"call_id,omitempty"`
	Output    string        `json:"output,omitempty"`
	Status    string        `json:"status,omitempty"`
}

func (i InputItem) MarshalJSON() ([]byte, error) {
	if i.Type == "function_call_output" {
		payload := map[string]any{
			"type": i.Type,
		}
		if i.CallID != "" {
			payload["call_id"] = i.CallID
		}
		if i.Output != "" {
			payload["output"] = i.Output
		}
		return json.Marshal(payload)
	}

	payload := map[string]any{}
	if i.Type != "" {
		payload["type"] = i.Type
	}
	if i.Role != "" {
		payload["role"] = i.Role
	}
	if i.Content.Text != nil || len(i.Content.Items) > 0 {
		payload["content"] = i.Content
	}
	if i.ID != "" {
		payload["id"] = i.ID
	}
	if len(i.Summary) > 0 {
		payload["summary"] = i.Summary
	}
	if i.Name != "" {
		payload["name"] = i.Name
	}
	if i.Arguments != "" {
		payload["arguments"] = i.Arguments
	}
	if i.CallID != "" {
		payload["call_id"] = i.CallID
	}
	if i.Output != "" {
		payload["output"] = i.Output
	}
	if i.Status != "" {
		payload["status"] = i.Status
	}
	return json.Marshal(payload)
}

// InputContent supports string content or a list of content items.
type InputContent struct {
	Text  *string
	Items []ContentItem
}

func NewInputContentText(text string) InputContent {
	return InputContent{Text: &text}
}

func NewInputContentItems(items ...ContentItem) InputContent {
	return InputContent{Items: items}
}

func (c InputContent) MarshalJSON() ([]byte, error) {
	if c.Text != nil {
		return json.Marshal(*c.Text)
	}
	if c.Items != nil {
		return json.Marshal(c.Items)
	}
	return []byte("null"), nil
}

func (c *InputContent) UnmarshalJSON(data []byte) error {
	if c == nil {
		return fmt.Errorf("ark types: input content is nil")
	}
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		c.Text = &text
		c.Items = nil
		return nil
	}
	var items []ContentItem
	if err := json.Unmarshal(data, &items); err == nil {
		c.Items = items
		c.Text = nil
		return nil
	}
	return fmt.Errorf("ark types: unsupported input content")
}

// ContentItem describes a content part for input messages.
type ContentItem struct {
	Type            string           `json:"type"`
	Text            string           `json:"text,omitempty"`
	ImageURL        *ImageURL        `json:"image_url,omitempty"`
	FileID          string           `json:"file_id,omitempty"`
	Detail          string           `json:"detail,omitempty"`
	ImagePixelLimit *ImagePixelLimit `json:"image_pixel_limit,omitempty"`
}

// ImageURL supports string or object payload.
type ImageURL struct {
	URL    string `json:"url,omitempty"`
	Detail string `json:"detail,omitempty"`
	Raw    string `json:"-"`
}

// ImagePixelLimit controls image size constraints.
type ImagePixelLimit struct {
	MaxPixels int `json:"max_pixels,omitempty"`
	MinPixels int `json:"min_pixels,omitempty"`
}

func (u ImageURL) MarshalJSON() ([]byte, error) {
	if u.Raw != "" && u.URL == "" {
		return json.Marshal(u.Raw)
	}
	payload := map[string]any{}
	if u.URL != "" {
		payload["url"] = u.URL
	}
	if u.Detail != "" {
		payload["detail"] = u.Detail
	}
	return json.Marshal(payload)
}

func (u *ImageURL) UnmarshalJSON(data []byte) error {
	if u == nil {
		return fmt.Errorf("ark types: image url is nil")
	}
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	var raw string
	if err := json.Unmarshal(data, &raw); err == nil {
		u.Raw = raw
		u.URL = raw
		return nil
	}
	var obj struct {
		URL    string `json:"url"`
		Detail string `json:"detail"`
	}
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	u.URL = obj.URL
	u.Detail = obj.Detail
	return nil
}

// ThinkingConfig toggles deep thinking mode.
type ThinkingConfig struct {
	Type string `json:"type"`
}

// ReasoningConfig controls reasoning effort.
type ReasoningConfig struct {
	Effort string `json:"effort,omitempty"`
}

// CachingConfig controls caching behavior.
type CachingConfig struct {
	Type string `json:"type"`
}

// TextFormat defines structured output format.
type TextFormat struct {
	Format any `json:"format"`
}

// ContextManagement controls context edits.
type ContextManagement struct {
	Edits []ContextEdit `json:"edits,omitempty"`
}

// ContextEdit describes an edit policy.
type ContextEdit struct {
	Type           string          `json:"type"`
	Keep           *ContextKeep    `json:"keep,omitempty"`
	Trigger        *ContextTrigger `json:"trigger,omitempty"`
	ExcludeTools   []string        `json:"exclude_tools,omitempty"`
	ClearToolInput *bool           `json:"clear_tool_input,omitempty"`
}

// ContextKeep defines retention policy.
type ContextKeep struct {
	Type  string `json:"type"`
	Value int    `json:"value,omitempty"`
}

// ContextTrigger defines trigger policy.
type ContextTrigger struct {
	Type  string `json:"type"`
	Value int    `json:"value,omitempty"`
}

// MergeExtra merges custom fields into the request payload.
func (r ResponseRequest) MarshalJSON() ([]byte, error) {
	payload := map[string]any{}
	data, err := json.Marshal(struct {
		Model              string             `json:"model"`
		Input              InputUnion         `json:"input"`
		Instructions       *string            `json:"instructions,omitempty"`
		PreviousResponseID *string            `json:"previous_response_id,omitempty"`
		ExpireAt           *int64             `json:"expire_at,omitempty"`
		MaxOutputTokens    *int               `json:"max_output_tokens,omitempty"`
		Thinking           *ThinkingConfig    `json:"thinking,omitempty"`
		Reasoning          *ReasoningConfig   `json:"reasoning,omitempty"`
		Caching            *CachingConfig     `json:"caching,omitempty"`
		Store              *bool              `json:"store,omitempty"`
		Stream             *bool              `json:"stream,omitempty"`
		Temperature        *float64           `json:"temperature,omitempty"`
		TopP               *float64           `json:"top_p,omitempty"`
		Text               *TextFormat        `json:"text,omitempty"`
		Tools              []Tool             `json:"tools,omitempty"`
		ToolChoice         any                `json:"tool_choice,omitempty"`
		ParallelToolCalls  *bool              `json:"parallel_tool_calls,omitempty"`
		ContextManagement  *ContextManagement `json:"context_management,omitempty"`
	}{
		Model:              r.Model,
		Input:              r.Input,
		Instructions:       r.Instructions,
		PreviousResponseID: r.PreviousResponseID,
		ExpireAt:           r.ExpireAt,
		MaxOutputTokens:    r.MaxOutputTokens,
		Thinking:           r.Thinking,
		Reasoning:          r.Reasoning,
		Caching:            r.Caching,
		Store:              r.Store,
		Stream:             r.Stream,
		Temperature:        r.Temperature,
		TopP:               r.TopP,
		Text:               r.Text,
		Tools:              r.Tools,
		ToolChoice:         r.ToolChoice,
		ParallelToolCalls:  r.ParallelToolCalls,
		ContextManagement:  r.ContextManagement,
	})
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	for k, v := range r.Extra {
		payload[k] = v
	}
	return json.Marshal(payload)
}
