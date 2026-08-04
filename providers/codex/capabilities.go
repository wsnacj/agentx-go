package codex

import (
	"strings"

	llm "github.com/wsnacj/agentx-go/components/llm"
)

// ModelCapabilities reports capabilities enabled by one canonical Codex
// Responses adapter configuration.
func (config ModelConfig) ModelCapabilities() llm.ModelCapabilities {
	reasoning := config.Thinking || strings.TrimSpace(config.ReasoningDefault) != ""
	if config.ReasoningEffort != nil {
		reasoning = reasoning || *config.ReasoningEffort
	}
	return llm.ModelCapabilities{
		TextGeneration:   true,
		ToolCalling:      true,
		ReasoningControl: reasoning,
	}
}
