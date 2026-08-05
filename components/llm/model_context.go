package llm

import "context"

// ModelLimits describes provider-neutral token boundaries for one configured
// model. Zero means the Host has not declared that limit.
type ModelLimits struct {
	ContextWindowTokens int64 `json:"context_window_tokens,omitempty"`
	MaxInputTokens      int64 `json:"max_input_tokens,omitempty"`
	MaxOutputTokens     int64 `json:"max_output_tokens,omitempty"`
}

// Normalize returns non-negative limits and clamps per-direction limits to a
// known context window. It does not invent an unknown limit.
func (l ModelLimits) Normalize() ModelLimits {
	if l.ContextWindowTokens < 0 {
		l.ContextWindowTokens = 0
	}
	if l.MaxInputTokens < 0 {
		l.MaxInputTokens = 0
	}
	if l.MaxOutputTokens < 0 {
		l.MaxOutputTokens = 0
	}
	if l.ContextWindowTokens > 0 {
		if l.MaxInputTokens > l.ContextWindowTokens {
			l.MaxInputTokens = l.ContextWindowTokens
		}
		if l.MaxOutputTokens > l.ContextWindowTokens {
			l.MaxOutputTokens = l.ContextWindowTokens
		}
	}
	return l
}

// TokenCount is a request-time input count. Exact distinguishes a model
// tokenizer result from a conservative estimate; provider-reported response
// usage continues to use Usage.
type TokenCount struct {
	Tokens int64  `json:"tokens"`
	Exact  bool   `json:"exact"`
	Source string `json:"source,omitempty"`
}

// TokenCountRequest contains the complete provider-bound input surface that
// can affect input token usage.
type TokenCountRequest struct {
	Model    string
	System   string
	Messages Conversation
	Tools    []Tool
}

// TokenCounter counts or conservatively estimates provider-bound input. It
// must honor ctx and must not mutate the request.
type TokenCounter interface {
	CountInput(context.Context, TokenCountRequest) (TokenCount, error)
}

// TokenCounterFunc adapts a Host function to TokenCounter.
type TokenCounterFunc func(context.Context, TokenCountRequest) (TokenCount, error)

// CountInput delegates to f.
func (f TokenCounterFunc) CountInput(ctx context.Context, request TokenCountRequest) (TokenCount, error) {
	return f(ctx, request)
}
