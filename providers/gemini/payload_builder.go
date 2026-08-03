package gemini

import (
	"context"

	types "github.com/wsnacj/agentx-go/components/llm"
)

func chatRequestFromVisualRequest(req types.VisualRequest) types.ChatRequest {
	return types.ChatRequest{
		Model:       req.Model,
		System:      req.System,
		Messages:    req.Messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		Options:     req.Options,
	}
}

func buildGeneratePayload(ctx context.Context, resolver MediaResolver, cfg ModelConfig, req types.ChatRequest, visuals []types.VisualContent) (map[string]any, error) {
	sanitizedOptions := types.SanitizeProviderOptionMap(req.Options)
	payload := map[string]any{}
	contents, err := buildContents(ctx, resolver, cfg, req.Messages, visuals)
	if err != nil {
		return nil, err
	}
	payload["contents"] = contents

	systemInstruction := buildSystemInstruction(req.System)
	if systemInstruction != nil {
		payload["systemInstruction"] = systemInstruction
	}

	generationConfig := buildGenerationConfig(req.MaxTokens, req.Temperature, sanitizedOptions)
	if len(generationConfig) > 0 {
		payload["generationConfig"] = generationConfig
	}
	return payload, nil
}
