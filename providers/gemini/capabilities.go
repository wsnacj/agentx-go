package gemini

import llm "github.com/wsnacj/agentx-go/components/llm"

// ModelCapabilities projects Gemini Provider features into the neutral
// contract. ToolCalling remains false because the high-level Provider path
// currently rejects tools even though the lower-level Gemini API types expose
// them.
func (config ModelConfig) ModelCapabilities() llm.ModelCapabilities {
	return llm.ModelCapabilities{
		TextGeneration:  true,
		VisionInput:     config.Capability.Vision,
		Streaming:       config.Capability.Streaming,
		LocalMediaInput: config.Capability.LocalFiles,
	}
}
