package openaicompat

import llm "github.com/wsnacj/agentx-go/components/llm"

// ModelCapabilities projects the explicitly configured OpenAI-compatible
// model features into the provider-neutral contract.
func (config ModelConfig) ModelCapabilities() llm.ModelCapabilities {
	return llm.ModelCapabilities{
		TextGeneration:   true,
		ToolCalling:      true,
		VisionInput:      config.Capability.Vision,
		Streaming:        config.Capability.Streaming,
		LocalMediaInput:  config.Capability.LocalFiles,
		ReasoningControl: config.Capability.Thinking || enabled(config.Capability.ReasoningEffort),
		ParallelTools:    enabled(config.Capability.ParallelToolCalls),
		BotCompletion:    config.Capability.Bots,
	}
}

func enabled(value *bool) bool { return value != nil && *value }
