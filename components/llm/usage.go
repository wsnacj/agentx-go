package llm

import "time"

// Usage tracks token accounting data.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// UsageRecord tags usage with source metadata.
type UsageRecord struct {
	Timestamp time.Time   `json:"timestamp"`
	Feature   string      `json:"feature"`
	Metadata  interface{} `json:"metadata"`
	Usage     Usage       `json:"usage"`
}
