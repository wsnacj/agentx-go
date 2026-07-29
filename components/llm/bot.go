package llm

// BotReference points to external artefacts used in responses.
type BotReference struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	URL         string            `json:"url"`
	Summary     string            `json:"summary"`
	SiteName    string            `json:"site_name"`
	PublishTime string            `json:"publish_time"`
	Extra       map[string]string `json:"extra,omitempty"`
}

// BotUsageAction describes tool invocation usage.
type BotUsageAction struct {
	ActionName string `json:"action_name"`
	Count      int    `json:"count"`
}

// BotUsageModel exposes model level usage data.
type BotUsageModel struct {
	Name             string `json:"name"`
	PromptTokens     int    `json:"prompt_tokens"`
	CompletionTokens int    `json:"completion_tokens"`
	TotalTokens      int    `json:"total_tokens"`
}

// BotUsage aggregates action and model usage.
type BotUsage struct {
	Actions []BotUsageAction `json:"actions"`
	Models  []BotUsageModel  `json:"models"`
}

// BotResponse is a high level bot response schema.
type BotResponse struct {
	Content    string         `json:"content"`
	References []BotReference `json:"references"`
	Usage      BotUsage       `json:"usage"`
	RequestID  string         `json:"request_id"`
	Raw        []byte         `json:"-"`
}
