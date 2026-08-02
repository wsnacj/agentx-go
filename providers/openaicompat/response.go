package openaicompat

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"

	llm "github.com/wsnacj/agentx-go/components/llm"
)

type toolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}
type toolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function toolCallFunction `json:"function"`
}
type chatChoice struct {
	Message struct {
		Role      string     `json:"role"`
		Content   string     `json:"content"`
		ToolCalls []toolCall `json:"tool_calls"`
	} `json:"message"`
}
type chatResponse struct {
	Choices []chatChoice `json:"choices"`
	Usage   llm.Usage    `json:"usage"`
}
type embeddingVector struct {
	Embedding       any               `json:"embedding"`
	SparseEmbedding []llm.SparseEntry `json:"sparse_embedding"`
}
type embeddingResponse struct {
	Data  []embeddingVector `json:"data"`
	Usage llm.Usage         `json:"usage"`
}
type botReference struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	URL         string            `json:"url"`
	Summary     string            `json:"summary"`
	SiteName    string            `json:"site_name"`
	PublishTime string            `json:"publish_time"`
	Extra       map[string]string `json:"extra"`
}
type botUsage struct {
	ActionUsage []llm.BotUsageAction `json:"action_usage"`
	ModelUsage  []llm.BotUsageModel  `json:"model_usage"`
}
type botAPIResponse struct {
	chatResponse
	References []botReference `json:"references"`
	BotUsage   botUsage       `json:"bot_usage"`
	ID         string         `json:"id"`
}

func parseChatResponse(body []byte) (*llm.ChatResponse, error) {
	var response chatResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("decode chat response: %w", err)
	}
	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("empty chat choices")
	}
	calls := make([]llm.FunctionCall, 0, len(response.Choices[0].Message.ToolCalls))
	for _, call := range response.Choices[0].Message.ToolCalls {
		if call.Function.Name != "" || call.Function.Arguments != "" {
			calls = append(calls, llm.FunctionCall{ID: call.ID, Type: call.Type, Name: call.Function.Name, Arguments: call.Function.Arguments})
		}
	}
	return &llm.ChatResponse{Content: response.Choices[0].Message.Content, Raw: body, Calls: calls}, nil
}

func parseEmbeddingResponse(body []byte) (*llm.EmbeddingResponse, error) {
	var response embeddingResponse
	if err := json.Unmarshal(body, &response); err == nil && len(response.Data) > 0 {
		vectors := make([][]float32, 0, len(response.Data))
		sparse := make([][]llm.SparseEntry, 0, len(response.Data))
		hasSparse := false
		for _, item := range response.Data {
			vector, err := decodeEmbeddingValue(item.Embedding)
			if err != nil {
				return nil, fmt.Errorf("decode embedding response: %w", err)
			}
			vectors = append(vectors, vector)
			sparse = append(sparse, item.SparseEmbedding)
			hasSparse = hasSparse || len(item.SparseEmbedding) > 0
		}
		out := &llm.EmbeddingResponse{Vectors: vectors, Raw: body}
		if hasSparse {
			out.SparseVectors = sparse
		}
		return out, nil
	}
	var single struct {
		Data struct {
			Embedding       any               `json:"embedding"`
			SparseEmbedding []llm.SparseEntry `json:"sparse_embedding"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &single); err == nil && single.Data.Embedding != nil {
		vector, err := decodeEmbeddingValue(single.Data.Embedding)
		if err != nil {
			return nil, fmt.Errorf("decode embedding response: %w", err)
		}
		out := &llm.EmbeddingResponse{Vectors: [][]float32{vector}, Raw: body}
		if len(single.Data.SparseEmbedding) > 0 {
			out.SparseVectors = [][]llm.SparseEntry{single.Data.SparseEmbedding}
		}
		return out, nil
	}
	return nil, fmt.Errorf("decode embedding response: unsupported format")
}

func decodeEmbeddingValue(value any) ([]float32, error) {
	switch value := value.(type) {
	case []float32:
		return value, nil
	case []float64:
		out := make([]float32, len(value))
		for i, item := range value {
			out[i] = float32(item)
		}
		return out, nil
	case []any:
		out := make([]float32, len(value))
		for i, item := range value {
			number, ok := item.(float64)
			if !ok {
				return nil, fmt.Errorf("invalid embedding element type %T", item)
			}
			out[i] = float32(number)
		}
		return out, nil
	case string:
		data, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return nil, fmt.Errorf("decode embedding base64: %w", err)
		}
		if len(data)%4 != 0 {
			return nil, fmt.Errorf("invalid embedding byte length %d", len(data))
		}
		out := make([]float32, len(data)/4)
		for i := range out {
			out[i] = math.Float32frombits(binary.LittleEndian.Uint32(data[i*4 : (i+1)*4]))
		}
		return out, nil
	case nil:
		return nil, fmt.Errorf("empty embedding")
	default:
		return nil, fmt.Errorf("unsupported embedding type %T", value)
	}
}

func parseBotResponse(body []byte) (*llm.BotResponse, error) {
	var response botAPIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("decode bot response: %w", err)
	}
	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("empty bot choices")
	}
	out := &llm.BotResponse{Content: response.Choices[0].Message.Content, References: make([]llm.BotReference, len(response.References)), Usage: llm.BotUsage{Actions: response.BotUsage.ActionUsage, Models: response.BotUsage.ModelUsage}, RequestID: response.ID, Raw: body}
	for i, ref := range response.References {
		out.References[i] = llm.BotReference{ID: ref.ID, Title: ref.Title, URL: ref.URL, Summary: ref.Summary, SiteName: ref.SiteName, PublishTime: ref.PublishTime, Extra: ref.Extra}
	}
	return out, nil
}

func extractUsage(body []byte) (*llm.Usage, error) {
	var response struct {
		Usage llm.Usage `json:"usage"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	return &response.Usage, nil
}
