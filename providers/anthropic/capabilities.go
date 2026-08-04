package anthropic

import llm "github.com/wsnacj/agentx-go/components/llm"

// ModelCapabilities reports capabilities implemented by the canonical
// Anthropic Messages adapter.
func (ModelConfig) ModelCapabilities() llm.ModelCapabilities {
	return llm.ModelCapabilities{TextGeneration: true, ToolCalling: true}
}
