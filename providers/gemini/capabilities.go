package gemini

import llm "github.com/wsnacj/agentx-go/components/llm"

// ModelCapabilities projects explicitly enabled Gemini Provider features into
// the neutral contract. Each flag is independent; Streaming and ToolCalling do
// not by themselves claim that visual tool calling is available.
func (config ModelConfig) ModelCapabilities() llm.ModelCapabilities {
	return llm.ModelCapabilities{
		TextGeneration:  true,
		ToolCalling:     config.Capability.ToolCalling,
		VisionInput:     config.Capability.Vision,
		Streaming:       config.Capability.Streaming,
		LocalMediaInput: config.Capability.LocalFiles,
	}
}
