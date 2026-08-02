// Package openaicompat implements the OpenAI-compatible HTTP provider protocol.
package openaicompat

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	llm "github.com/wsnacj/agentx-go/components/llm"
	"github.com/wsnacj/agentx-go/providers"
	"github.com/wsnacj/agentx-go/providers/transport"
)

const (
	defaultChatTimeout      = 2 * time.Minute
	defaultEmbeddingTimeout = 30 * time.Second
	defaultStreamTimeout    = 5 * time.Minute
	streamScannerBuffer     = 64 * 1024
	streamScannerMaxBuffer  = 1024 * 1024
)

// Client is an immutable, concurrency-safe OpenAI-compatible client.
type Client struct {
	name         string
	baseURL      string
	profile      profile
	transport    transport.Config
	authorize    Authorizer
	resolveMedia MediaResolver
	httpClient   HTTPDoer
}

// New constructs a client without reading credentials or making network calls.
func New(cfg Config) (*Client, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		return nil, fmt.Errorf("openaicompat: base URL is required")
	}
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: cfg.Timeout}
	}
	return &Client{name: cfg.Name, baseURL: baseURL, profile: newProfile(cfg.CompatProfile), transport: transport.Config{Mode: cfg.Transport.Mode, Headers: cloneHeaders(cfg.Transport.Headers)}, authorize: cfg.Authorize, resolveMedia: cfg.ResolveMedia, httpClient: client}, nil
}

func (c *Client) Chat(ctx context.Context, cfg ModelConfig, req llm.ChatRequest) (*llm.ChatResponse, *llm.Usage, error) {
	req = resolveChatRequest(cfg, req)
	payload, settings := c.buildChatPayload(cfg, req, false)
	body, usage, err := c.post(ctx, c.chatPath(cfg), payload, defaultChatTimeout, settings)
	if err != nil {
		return nil, nil, err
	}
	response, err := parseChatResponse(body)
	return response, usage, err
}
func (c *Client) Vision(ctx context.Context, cfg ModelConfig, req llm.VisualRequest) (*llm.VisualResponse, *llm.Usage, error) {
	if !cfg.Capability.Vision {
		return nil, nil, fmt.Errorf("model %s does not support vision", cfg.Name)
	}
	req = resolveVisualRequest(cfg, req)
	payload, settings := c.buildVisionPayload(cfg, req, false)
	body, usage, err := c.post(ctx, c.chatPath(cfg), payload, defaultChatTimeout, settings)
	if err != nil {
		return nil, nil, err
	}
	response, err := parseChatResponse(body)
	if err != nil {
		return nil, nil, err
	}
	return &llm.VisualResponse{Content: response.Content, Raw: response.Raw}, usage, nil
}
func (c *Client) Embedding(ctx context.Context, cfg EmbeddingConfig, req llm.EmbeddingRequest) (*llm.EmbeddingResponse, *llm.Usage, error) {
	payload := map[string]any{"model": cfg.Model, "input": c.embeddingInput(cfg, req)}
	if cfg.Dimensions > 0 {
		payload["dimensions"] = cfg.Dimensions
	}
	if cfg.Encoding != "" {
		payload["encoding_format"] = cfg.Encoding
	}
	c.profile.mergeEmbeddingOptions(payload, cfg, req.Options)
	settings := transport.Resolve(c.transport, llm.RequestOptionsFromMap(req.Options))
	path := firstNonEmpty(cfg.Path, req.Path, "/embeddings")
	body, usage, err := c.post(ctx, path, payload, defaultEmbeddingTimeout, settings)
	if err != nil {
		return nil, nil, err
	}
	response, err := parseEmbeddingResponse(body)
	return response, usage, err
}
func (c *Client) Bot(ctx context.Context, cfg ModelConfig, req llm.ChatRequest) (*llm.BotResponse, *llm.Usage, error) {
	if !cfg.Capability.Bots {
		return nil, nil, fmt.Errorf("model %s does not enable bot capability", cfg.Name)
	}
	req = resolveChatRequest(cfg, req)
	payload, settings := c.buildChatPayload(cfg, req, false)
	body, usage, err := c.post(ctx, "/bots/chat/completions", payload, defaultChatTimeout, settings)
	if err != nil {
		return nil, nil, err
	}
	response, err := parseBotResponse(body)
	return response, usage, err
}
func (c *Client) StreamChat(ctx context.Context, cfg ModelConfig, req llm.ChatRequest) (*llm.StreamResult, error) {
	events, err := c.StreamChatEvents(ctx, cfg, req)
	if err != nil {
		return nil, err
	}
	return llm.BridgeEventStreamResult(events), nil
}
func (c *Client) StreamVision(ctx context.Context, cfg ModelConfig, req llm.VisualRequest) (*llm.StreamResult, error) {
	events, err := c.StreamVisionEvents(ctx, cfg, req)
	if err != nil {
		return nil, err
	}
	return llm.BridgeEventStreamResult(events), nil
}
func (c *Client) StreamChatEvents(ctx context.Context, cfg ModelConfig, req llm.ChatRequest) (*llm.EventStreamResult, error) {
	if !cfg.Capability.Streaming {
		return nil, providers.ErrUnsupported
	}
	req = resolveChatRequest(cfg, req)
	payload, settings := c.buildChatPayload(cfg, req, true)
	return c.stream(ctx, c.chatPath(cfg), payload, settings)
}
func (c *Client) StreamVisionEvents(ctx context.Context, cfg ModelConfig, req llm.VisualRequest) (*llm.EventStreamResult, error) {
	if !cfg.Capability.Vision {
		return nil, fmt.Errorf("model %s does not support vision", cfg.Name)
	}
	if !cfg.Capability.Streaming {
		return nil, providers.ErrUnsupported
	}
	req = resolveVisualRequest(cfg, req)
	payload, settings := c.buildVisionPayload(cfg, req, true)
	return c.stream(ctx, c.chatPath(cfg), payload, settings)
}

func (c *Client) chatPath(cfg ModelConfig) string {
	if cfg.Capability.Bots {
		return "/bots/chat/completions"
	}
	return "/chat/completions"
}
func (c *Client) buildChatPayload(cfg ModelConfig, req llm.ChatRequest, stream bool) (map[string]any, transport.Settings) {
	options := llm.SanitizeProviderOptionMap(req.Options)
	payload := map[string]any{"model": req.Model, "messages": marshalMessages(c.profile, req.System, req.Messages), "temperature": req.Temperature}
	if stream {
		payload["stream"] = true
	}
	c.profile.applyTooling(payload, req.Tools, req.ToolChoice)
	c.profile.mergeOptions(payload, cfg.Capability, options)
	c.profile.applyMaxTokens(payload, req.MaxTokens, options)
	c.profile.applyStreamingDefaults(payload)
	settings := transport.Resolve(c.transport, llm.RequestOptionsFromMap(req.Options))
	transport.ApplyPayload(payload, settings)
	if stream {
		payload["stream"] = true
	}
	return payload, settings
}
func (c *Client) buildVisionPayload(cfg ModelConfig, req llm.VisualRequest, stream bool) (map[string]any, transport.Settings) {
	options := llm.SanitizeProviderOptionMap(req.Options)
	payload := map[string]any{"model": req.Model, "messages": c.marshalVisualMessages(cfg, req), "temperature": req.Temperature}
	if stream {
		payload["stream"] = true
	}
	c.profile.applyTooling(payload, req.Tools, req.ToolChoice)
	c.profile.mergeOptions(payload, cfg.Capability, options)
	c.profile.applyMaxTokens(payload, req.MaxTokens, options)
	c.profile.applyStreamingDefaults(payload)
	settings := transport.Resolve(c.transport, llm.RequestOptionsFromMap(req.Options))
	transport.ApplyPayload(payload, settings)
	if stream {
		payload["stream"] = true
	}
	return payload, settings
}

func (c *Client) post(ctx context.Context, path string, payload map[string]any, timeout time.Duration, settings transport.Settings) ([]byte, *llm.Usage, error) {
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	response, err := c.do(ctx, path, payload, settings, false)
	if err != nil {
		return nil, nil, err
	}
	defer response.Body.Close()
	raw, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read response: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, nil, &providers.APIError{StatusCode: response.StatusCode, Body: string(raw)}
	}
	usage, _ := extractUsage(raw)
	return raw, usage, nil
}

func (c *Client) do(ctx context.Context, path string, payload map[string]any, settings transport.Settings, stream bool) (*http.Response, error) {
	payloadAny, err := transport.ApplyPayloadHook(ctx, settings, payload)
	if err != nil {
		return nil, fmt.Errorf("apply payload hook: %w", err)
	}
	data, err := json.Marshal(payloadAny)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}
	target := c.baseURL + path
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	if stream {
		request.Header.Set("Accept", "text/event-stream")
	}
	transport.ApplyHeaders(request.Header, settings)
	if c.authorize != nil {
		if err := c.authorize(ctx, request.Header); err != nil {
			return nil, err
		}
	}
	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	if err := transport.ApplyResponseHook(ctx, settings, transport.ResponseMetadataFromHTTP(http.MethodPost, target, response)); err != nil {
		response.Body.Close()
		return nil, fmt.Errorf("apply response hook: %w", err)
	}
	return response, nil
}

func (c *Client) stream(ctx context.Context, path string, payload map[string]any, settings transport.Settings) (*llm.EventStreamResult, error) {
	streamCtx := ctx
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); ok {
		streamCtx, cancel = context.WithCancel(ctx)
	} else {
		streamCtx, cancel = context.WithTimeout(ctx, defaultStreamTimeout)
	}
	response, err := c.do(streamCtx, path, payload, settings, true)
	if err != nil {
		cancel()
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(response.Body)
		response.Body.Close()
		cancel()
		return nil, &providers.APIError{StatusCode: response.StatusCode, Body: string(raw)}
	}
	events := make(chan llm.StreamEvent)
	go func() {
		defer close(events)
		defer response.Body.Close()
		defer cancel()
		events <- llm.StreamEvent{Type: llm.StreamEventStart}
		scanner := bufio.NewScanner(response.Body)
		scanner.Buffer(make([]byte, 0, streamScannerBuffer), streamScannerMaxBuffer)
		parser := streamParser{}
		for scanner.Scan() {
			select {
			case <-streamCtx.Done():
				events <- llm.StreamEvent{Type: llm.StreamEventError, Err: streamCtx.Err()}
				events <- llm.StreamEvent{Type: llm.StreamEventDone}
				return
			default:
			}
			line := strings.TrimSpace(scanner.Text())
			if !strings.HasPrefix(line, "data:") {
				continue
			}
			parsed, done := parser.parse(strings.TrimSpace(strings.TrimPrefix(line, "data:")))
			for _, event := range parsed {
				events <- event
			}
			if done {
				return
			}
		}
		if err := scanner.Err(); err != nil {
			events <- llm.StreamEvent{Type: llm.StreamEventError, Err: fmt.Errorf("stream read error: %w", err)}
		}
	}()
	return &llm.EventStreamResult{Ch: events, Cancel: func() { cancel(); response.Body.Close() }}, nil
}

func (c *Client) embeddingInput(cfg EmbeddingConfig, req llm.EmbeddingRequest) any {
	if value, ok := req.Options["embedding_input"]; ok {
		return value
	}
	if value, ok := req.Options["multimodal_input"]; ok {
		return value
	}
	videos := stringSlice(req.Options["video_urls"])
	if len(videos) == 0 {
		videos = stringSlice(req.Options["video_url"])
	}
	if optionDisabled(cfg.Capability.EmbeddingVideo) {
		videos = nil
	}
	if len(req.Images) == 0 && len(videos) == 0 {
		return req.Inputs
	}
	wrapper := cfg.Wrapper
	if wrapper == "" {
		wrapper = "typed_array"
	}
	if len(req.Images) == 0 {
		items := make([]map[string]any, 0, len(req.Inputs)+len(videos))
		for _, text := range req.Inputs {
			if text != "" {
				items = append(items, map[string]any{"type": "text", "text": text})
			}
		}
		for _, video := range videos {
			items = append(items, map[string]any{"type": "video_url", "video_url": map[string]string{"url": video}})
		}
		return items
	}
	if wrapper == "image_array" && len(videos) == 0 {
		items := make([]map[string]string, 0, len(req.Images)*2)
		for i, image := range req.Images {
			if i < len(req.Inputs) && req.Inputs[i] != "" {
				items = append(items, map[string]string{"text": req.Inputs[i]})
			}
			image = c.resolveLocalMedia(cfg.Capability.LocalFiles, image)
			items = append(items, map[string]string{"image": image})
		}
		return items
	}
	items := make([]map[string]any, 0, len(req.Images)*2+len(videos))
	for i, image := range req.Images {
		if i < len(req.Inputs) && req.Inputs[i] != "" {
			items = append(items, map[string]any{"type": "text", "text": req.Inputs[i]})
		}
		image = c.resolveLocalMedia(cfg.Capability.LocalFiles, image)
		items = append(items, map[string]any{"type": "image_url", "image_url": map[string]string{"url": image}})
	}
	for _, video := range videos {
		items = append(items, map[string]any{"type": "video_url", "video_url": map[string]string{"url": video}})
	}
	return items
}

func (c *Client) resolveLocalMedia(enabled bool, value string) string {
	if enabled && c.resolveMedia != nil {
		if resolved, err := c.resolveMedia(value); err == nil {
			return resolved
		}
	}
	return value
}

func stringSlice(value any) []string {
	switch value := value.(type) {
	case string:
		if value != "" {
			return []string{value}
		}
	case []string:
		return value
	case []any:
		out := make([]string, 0, len(value))
		for _, item := range value {
			if text, ok := item.(string); ok && text != "" {
				out = append(out, text)
			}
		}
		return out
	}
	return nil
}

func resolveChatRequest(cfg ModelConfig, req llm.ChatRequest) llm.ChatRequest {
	if req.Model == "" {
		req.Model = cfg.Model
	}
	if req.MaxTokens <= 0 {
		req.MaxTokens = cfg.MaxCompletion
	}
	if req.Temperature == 0 {
		if value := llm.RequestOptionsFromMap(req.Options).Temperature; value != nil {
			req.Temperature = float32(*value)
		} else {
			req.Temperature = cfg.Temperature
		}
	}
	return req
}
func resolveVisualRequest(cfg ModelConfig, req llm.VisualRequest) llm.VisualRequest {
	if req.Model == "" {
		req.Model = cfg.Model
	}
	if req.MaxTokens <= 0 {
		req.MaxTokens = cfg.MaxCompletion
	}
	if req.Temperature == 0 {
		if value := llm.RequestOptionsFromMap(req.Options).Temperature; value != nil {
			req.Temperature = float32(*value)
		} else {
			req.Temperature = cfg.Temperature
		}
	}
	return req
}
func cloneHeaders(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}
