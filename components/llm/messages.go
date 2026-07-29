package llm

// Message represents a single role based exchange.
type Message struct {
	Role       string         `json:"role"`
	Content    string         `json:"content"`
	ToolName   string         `json:"tool_name,omitempty"`
	ToolCallID string         `json:"tool_call_id,omitempty"`
	ToolCalls  []FunctionCall `json:"tool_calls,omitempty"`
}

// Conversation maintains ordered message history.
type Conversation []Message

// VisualContent describes the payload for multi-modal requests.
type VisualContent struct {
	Type     string   `json:"type"`
	Text     string   `json:"text,omitempty"`
	ImageURL string   `json:"image_url,omitempty"`
	VideoURL string   `json:"video_url,omitempty"`
	Detail   string   `json:"detail,omitempty"`
	DataURI  string   `json:"data_uri,omitempty"`
	Labels   []string `json:"labels,omitempty"`
	FPS      *float32 `json:"fps,omitempty"`
}

// ChatRequest consolidates system prompt, conversation and user input.
type ChatRequest struct {
	Model       string
	System      string
	Messages    Conversation
	Temperature float32
	MaxTokens   int
	Options     map[string]any
	Tools       []Tool
	ToolChoice  *ToolChoice
}

// ChatResponse captures primary assistant output with optional raw payload.
type ChatResponse struct {
	Content string
	Raw     []byte
	Calls   []FunctionCall
	Usage   *Usage
}

// VisualRequest wraps a visual conversation.
type VisualRequest struct {
	Model       string
	System      string
	Messages    Conversation
	Visual      []VisualContent
	Temperature float32
	MaxTokens   int
	Options     map[string]any
	Tools       []Tool
	ToolChoice  *ToolChoice
}

// VisualResponse mirrors ChatResponse for ergonomic parity.
type VisualResponse struct {
	Content string
	Raw     []byte
	Usage   *Usage
}
