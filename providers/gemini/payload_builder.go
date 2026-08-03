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

// ChatRequestFromVisualRequest projects shared text fields from a visual request.
func ChatRequestFromVisualRequest(request types.VisualRequest) types.ChatRequest {
	return chatRequestFromVisualRequest(request)
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

// BuildGeneratePayload builds a Gemini native payload without performing I/O.
// Local media references require an explicit resolver.
func BuildGeneratePayload(ctx context.Context, resolver MediaResolver, config ModelConfig, request types.ChatRequest, visuals []types.VisualContent) (map[string]any, error) {
	return buildGeneratePayload(ctx, resolver, config, request, visuals)
}
