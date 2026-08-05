package llm

import "time"

// Usage tracks token accounting data.
type Usage struct {
	PromptTokens      int `json:"prompt_tokens"`
	CompletionTokens  int `json:"completion_tokens"`
	TotalTokens       int `json:"total_tokens"`
	CachedInputTokens int `json:"cached_input_tokens,omitempty"`
	CacheWriteTokens  int `json:"cache_write_tokens,omitempty"`
	ReasoningTokens   int `json:"reasoning_tokens,omitempty"`
}

// UsageRecord tags usage with source metadata.
type UsageRecord struct {
	Timestamp time.Time   `json:"timestamp"`
	Feature   string      `json:"feature"`
	Metadata  interface{} `json:"metadata"`
	Usage     Usage       `json:"usage"`
}
