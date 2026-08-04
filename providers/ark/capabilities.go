package ark

import llm "github.com/wsnacj/agentx-go/components/llm"

// ResponseModelCapabilities reports the portable model features implemented
// by the canonical Ark Responses path. Files and image generation are separate
// provider services and intentionally do not appear here.
func ResponseModelCapabilities() llm.ModelCapabilities {
	return llm.ModelCapabilities{
		TextGeneration:   true,
		ToolCalling:      true,
		Streaming:        true,
		ReasoningControl: true,
	}
}
