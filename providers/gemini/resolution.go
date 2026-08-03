package gemini

import llm "github.com/wsnacj/agentx-go/components/llm"

func resolveChatRequest(config ModelConfig, request llm.ChatRequest) llm.ChatRequest {
	out := request
	if out.Model == "" {
		out.Model = config.Model
	}
	if out.MaxTokens <= 0 {
		out.MaxTokens = config.MaxCompletion
	}
	out.Temperature = resolveTemperature(config.Temperature, out.Temperature, out.Options)
	return out
}

func resolveVisualRequest(config ModelConfig, request llm.VisualRequest) llm.VisualRequest {
	out := request
	if out.Model == "" {
		out.Model = config.Model
	}
	if out.MaxTokens <= 0 {
		out.MaxTokens = config.MaxCompletion
	}
	out.Temperature = resolveTemperature(config.Temperature, out.Temperature, out.Options)
	return out
}

func resolveTemperature(defaultValue, requestValue float32, options map[string]any) float32 {
	if requestValue != 0 {
		return requestValue
	}
	typed := llm.RequestOptionsFromMap(options)
	if typed.Temperature != nil {
		return float32(*typed.Temperature)
	}
	return defaultValue
}
