package gemini

import (
	"encoding/json"
	"fmt"

	types "github.com/wsnacj/agentx-go/components/llm"
)

type contentEmbedding struct {
	Values []float32 `json:"values"`
}

type embedContentResponse struct {
	Embedding contentEmbedding `json:"embedding"`
}

type batchEmbedContentResponse struct {
	Embeddings []contentEmbedding `json:"embeddings"`
}

func decodeGenerateContent(raw []byte) (*GenerateContentResponse, error) {
	var resp GenerateContentResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &resp, nil
}

func extractText(resp *GenerateContentResponse) string {
	if resp == nil || len(resp.Candidates) == 0 {
		return ""
	}
	parts := resp.Candidates[0].Content.Parts
	if len(parts) == 0 {
		return ""
	}
	out := ""
	for _, p := range parts {
		if p.Text == "" {
			continue
		}
		out += p.Text
	}
	return out
}

func extractUsage(resp *GenerateContentResponse) *types.Usage {
	if resp == nil || resp.UsageMetadata == nil {
		return nil
	}
	usage := resp.UsageMetadata
	return &types.Usage{
		PromptTokens:     usage.PromptTokenCount,
		CompletionTokens: usage.CandidatesTokenCount,
		TotalTokens:      usage.TotalTokenCount,
		ReasoningTokens:  usage.ThoughtsTokenCount,
	}
}

func buildEmbeddingPayload(cfg EmbeddingConfig, req types.EmbeddingRequest) (any, error) {
	inputs := req.Inputs
	if len(inputs) == 0 {
		return nil, fmt.Errorf("empty embedding input")
	}

	if len(inputs) == 1 {
		payload := map[string]any{
			"content": map[string]any{
				"parts": []map[string]any{{"text": inputs[0]}},
			},
		}
		if cfg.Dimensions > 0 {
			payload["outputDimensionality"] = cfg.Dimensions
		}
		return payload, nil
	}

	requests := make([]map[string]any, 0, len(inputs))
	for _, input := range inputs {
		request := map[string]any{
			"content": map[string]any{
				"parts": []map[string]any{{"text": input}},
			},
		}
		if cfg.Dimensions > 0 {
			request["outputDimensionality"] = cfg.Dimensions
		}
		requests = append(requests, request)
	}

	payload := map[string]any{
		"requests": requests,
	}
	return payload, nil
}

func parseEmbeddingResponse(raw []byte) ([][]float32, error) {
	var single embedContentResponse
	if err := json.Unmarshal(raw, &single); err == nil && len(single.Embedding.Values) > 0 {
		return [][]float32{single.Embedding.Values}, nil
	}

	var batch batchEmbedContentResponse
	if err := json.Unmarshal(raw, &batch); err != nil {
		return nil, fmt.Errorf("decode embedding response: %w", err)
	}
	if len(batch.Embeddings) == 0 {
		return nil, fmt.Errorf("empty embedding response")
	}

	vectors := make([][]float32, 0, len(batch.Embeddings))
	for _, emb := range batch.Embeddings {
		vectors = append(vectors, emb.Values)
	}
	return vectors, nil
}
