package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	types "github.com/wsnacj/agentx-go/components/llm"
	"github.com/wsnacj/agentx-go/providers"
	"github.com/wsnacj/agentx-go/providers/transport"
)

const (
	defaultBaseURL         = "https://generativelanguage.googleapis.com/v1beta"
	defaultChatTimeout     = 2 * time.Minute
	defaultEmbTimeout      = 30 * time.Second
	defaultStreamTimeout   = 5 * time.Minute
	streamScannerBuffer    = 64 * 1024
	streamScannerMaxBuffer = 1024 * 1024
)

// Provider implements the Gemini native API.
type Provider struct {
	cfg          Config
	httpClient   HTTPDoer
	resolveMedia MediaResolver
}

// NewProvider creates a Gemini provider for native API calls.
func NewProvider(cfg Config) *Provider {
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: cfg.Timeout}
	}
	cfg.Transport = transport.Config{Mode: cfg.Transport.Mode, Headers: cloneHeaders(cfg.Transport.Headers)}
	return &Provider{
		cfg:          cfg,
		httpClient:   client,
		resolveMedia: cfg.ResolveMedia,
	}
}

func (p *Provider) Chat(ctx context.Context, cfg ModelConfig, req types.ChatRequest) (*types.ChatResponse, *types.Usage, error) {
	req = resolveChatRequest(cfg, req)
	payload, usage, err := p.generate(ctx, cfg, req, nil, defaultChatTimeout)
	if err != nil {
		return nil, nil, err
	}
	return &types.ChatResponse{Content: payload.Text, Calls: payload.Calls, Raw: payload.Raw, Usage: usage}, usage, nil
}

func (p *Provider) Vision(ctx context.Context, cfg ModelConfig, req types.VisualRequest) (*types.VisualResponse, *types.Usage, error) {
	if !cfg.Capability.Vision {
		return nil, nil, fmt.Errorf("model %s does not support vision", cfg.Name)
	}
	req = resolveVisualRequest(cfg, req)
	if len(req.Tools) > 0 || req.ToolChoice != nil {
		return nil, nil, providers.ErrUnsupported
	}
	payload, usage, err := p.generate(ctx, cfg, chatRequestFromVisualRequest(req), req.Visual, defaultChatTimeout)
	if err != nil {
		return nil, nil, err
	}
	return &types.VisualResponse{Content: payload.Text, Raw: payload.Raw}, usage, nil
}

func (p *Provider) Embedding(ctx context.Context, cfg EmbeddingConfig, req types.EmbeddingRequest) (*types.EmbeddingResponse, *types.Usage, error) {
	if len(req.Images) > 0 {
		return nil, nil, fmt.Errorf("gemini embedding does not support images")
	}
	vectors, usage, err := p.embed(ctx, cfg, req, defaultEmbTimeout)
	if err != nil {
		return nil, nil, err
	}
	return &types.EmbeddingResponse{Vectors: vectors}, usage, nil
}

func (p *Provider) Bot(ctx context.Context, cfg ModelConfig, req types.ChatRequest) (*types.BotResponse, *types.Usage, error) {
	return nil, nil, providers.ErrUnsupported
}

func (p *Provider) StreamChat(ctx context.Context, cfg ModelConfig, req types.ChatRequest) (*types.StreamResult, error) {
	events, err := p.StreamChatEvents(ctx, cfg, req)
	if err != nil {
		return nil, err
	}
	return types.BridgeEventStreamResult(events), nil
}

func (p *Provider) StreamVision(ctx context.Context, cfg ModelConfig, req types.VisualRequest) (*types.StreamResult, error) {
	events, err := p.StreamVisionEvents(ctx, cfg, req)
	if err != nil {
		return nil, err
	}
	return types.BridgeEventStreamResult(events), nil
}

func (p *Provider) generate(ctx context.Context, cfg ModelConfig, req types.ChatRequest, visuals []types.VisualContent, timeout time.Duration) (generatedContent, *types.Usage, error) {
	payload, err := buildGeneratePayload(ctx, p.resolveMedia, cfg, req, visuals)
	if err != nil {
		return generatedContent{}, nil, err
	}
	settings := transport.Resolve(p.cfg.Transport, types.RequestOptionsFromMap(req.Options))

	endpoint := p.buildEndpoint(req.Model, "generateContent")
	raw, err := p.post(ctx, endpoint, payload, timeout, settings)
	if err != nil {
		return generatedContent{}, nil, err
	}

	resp, err := decodeGenerateContent(raw)
	if err != nil {
		return generatedContent{}, nil, err
	}
	return generatedContent{Text: extractText(resp), Calls: extractFunctionCalls(resp), Raw: raw}, extractUsage(resp), nil
}

func (p *Provider) embed(ctx context.Context, cfg EmbeddingConfig, req types.EmbeddingRequest, timeout time.Duration) ([][]float32, *types.Usage, error) {
	inputs := req.Inputs
	if len(inputs) == 0 {
		return nil, nil, fmt.Errorf("empty embedding input")
	}

	payload, err := buildEmbeddingPayload(cfg, req)
	if err != nil {
		return nil, nil, err
	}

	endpoint := p.buildEndpoint(cfg.Model, "embedContent")
	if len(inputs) > 1 {
		endpoint = p.buildEndpoint(cfg.Model, "batchEmbedContents")
	}

	raw, err := p.post(ctx, endpoint, payload, timeout, transport.Resolve(p.cfg.Transport, types.RequestOptionsFromMap(req.Options)))
	if err != nil {
		return nil, nil, err
	}

	vectors, err := parseEmbeddingResponse(raw)
	if err != nil {
		return nil, nil, err
	}
	return vectors, nil, nil
}

func (p *Provider) buildEndpoint(model string, action string) string {
	base := strings.TrimRight(p.cfg.BaseURL, "/")
	if base == "" {
		base = defaultBaseURL
	}

	resource := strings.TrimSpace(model)
	if resource == "" {
		resource = "gemini-3-flash-preview"
	}
	if !strings.HasPrefix(resource, "models/") && !strings.HasPrefix(resource, "tunedModels/") {
		resource = "models/" + resource
	}

	return strings.TrimRight(base, "/") + "/" + resource + ":" + action
}

func (p *Provider) post(ctx context.Context, endpoint string, payload any, timeout time.Duration, settings transport.Settings) ([]byte, error) {
	payloadAny, err := transport.ApplyPayloadHook(ctx, settings, payload)
	if err != nil {
		return nil, fmt.Errorf("apply payload hook: %w", err)
	}
	data, err := json.Marshal(payloadAny)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	transport.ApplyHeaders(req.Header, settings)
	if p.cfg.Authorize != nil {
		if err := p.cfg.Authorize(ctx, req.Header); err != nil {
			return nil, err
		}
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()
	if err := transport.ApplyResponseHook(ctx, settings, transport.ResponseMetadataFromHTTP(http.MethodPost, endpoint, resp)); err != nil {
		return nil, fmt.Errorf("apply response hook: %w", err)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &providers.APIError{StatusCode: resp.StatusCode, Body: string(raw)}
	}
	return raw, nil
}

func buildContents(ctx context.Context, resolver MediaResolver, cfg ModelConfig, messages types.Conversation, visuals []types.VisualContent) ([]map[string]any, error) {
	if len(visuals) == 0 {
		return buildMessageContents(messages)
	}
	// Preserve the established visual contract: message text and media are one
	// user turn. Visual Tool calling remains unsupported by the neutral response
	// surface and is rejected before this builder is reached.
	parts := make([]map[string]any, 0, len(messages)+len(visuals))
	for _, message := range messages {
		if message.Content != "" {
			parts = append(parts, map[string]any{"text": message.Content})
		}
	}
	for _, visual := range visuals {
		part, err := buildVisualPart(ctx, resolver, cfg, visual)
		if err != nil {
			return nil, err
		}
		if part != nil {
			parts = append(parts, part)
		}
	}

	if len(parts) == 0 {
		return nil, fmt.Errorf("empty content")
	}
	return []map[string]any{{"role": "user", "parts": parts}}, nil
}

func buildSystemInstruction(systemPrompt string) map[string]any {
	if strings.TrimSpace(systemPrompt) == "" {
		return nil
	}
	return map[string]any{
		"role":  "user",
		"parts": []map[string]any{{"text": systemPrompt}},
	}
}

func buildGenerationConfig(maxTokens int, temperature float32, options map[string]any) map[string]any {
	config := map[string]any{}
	if maxTokens > 0 {
		config["maxOutputTokens"] = maxTokens
	}
	config["temperature"] = temperature

	if options == nil {
		return config
	}

	if raw, ok := options["generationConfig"]; ok {
		if incoming, ok := raw.(map[string]any); ok {
			for k, v := range incoming {
				config[k] = v
			}
		}
	}

	for _, key := range []string{"temperature", "maxOutputTokens", "topP", "topK", "candidateCount", "presencePenalty", "frequencyPenalty", "stopSequences"} {
		if value, ok := options[key]; ok {
			config[key] = value
		}
	}
	opts := types.RequestOptionsFromMap(options)
	if opts.Thinking != nil {
		thinkingConfig, _ := config["thinkingConfig"].(map[string]any)
		if thinkingConfig == nil {
			thinkingConfig = map[string]any{}
		}
		if _, exists := thinkingConfig["includeThoughts"]; !exists {
			thinkingConfig["includeThoughts"] = opts.Thinking.Enabled
		}
		if len(thinkingConfig) > 0 {
			config["thinkingConfig"] = thinkingConfig
		}
	}
	return config
}

func buildVisualPart(ctx context.Context, resolver MediaResolver, cfg ModelConfig, visual types.VisualContent) (map[string]any, error) {
	switch visual.Type {
	case "text":
		if visual.Text == "" {
			return nil, nil
		}
		return map[string]any{"text": visual.Text}, nil
	case "video_url":
		return buildFilePart(ctx, resolver, cfg, pickFirstNonEmpty(visual.VideoURL, visual.ImageURL, visual.DataURI, visual.Text), visual.FPS)
	default:
		return buildFilePart(ctx, resolver, cfg, pickFirstNonEmpty(visual.ImageURL, visual.DataURI, visual.Text), nil)
	}
}

func buildFilePart(ctx context.Context, resolver MediaResolver, cfg ModelConfig, source string, fps *float32) (map[string]any, error) {
	if source == "" {
		return nil, fmt.Errorf("empty media source")
	}

	if mimeType, data, ok := parseDataURI(source); ok {
		part := map[string]any{
			"inlineData": map[string]any{
				"mimeType": mimeType,
				"data":     data,
			},
		}
		if fps != nil {
			part["videoMetadata"] = map[string]any{"fps": *fps}
		}
		return part, nil
	}

	if !isRemote(source) {
		if !cfg.Capability.LocalFiles {
			return nil, fmt.Errorf("local files disabled")
		}
		if resolver == nil {
			return nil, fmt.Errorf("local media resolver is required")
		}
		media, err := resolver(ctx, source)
		if err != nil {
			return nil, err
		}
		if media.Base64Data == "" && media.URI != "" {
			return buildFilePart(ctx, nil, cfg, media.URI, fps)
		}
		part := map[string]any{
			"inlineData": map[string]any{
				"mimeType": media.MIMEType,
				"data":     media.Base64Data,
			},
		}
		if fps != nil {
			part["videoMetadata"] = map[string]any{"fps": *fps}
		}
		return part, nil
	}

	mimeType := guessMimeTypeFromURL(source)
	fileData := map[string]any{
		"fileUri": source,
	}
	if mimeType != "" {
		fileData["mimeType"] = mimeType
	}

	part := map[string]any{
		"fileData": fileData,
	}
	if fps != nil {
		part["videoMetadata"] = map[string]any{"fps": *fps}
	}
	return part, nil
}

func isRemote(raw string) bool {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	return err == nil && parsed.Scheme != "" && !strings.EqualFold(parsed.Scheme, "file")
}

func guessMimeTypeFromURL(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	ext := path.Ext(parsed.Path)
	if ext == "" {
		return ""
	}
	if mimeType := mime.TypeByExtension(ext); mimeType != "" {
		return mimeType
	}
	switch strings.ToLower(ext) {
	case ".mp4":
		return "video/mp4"
	case ".mov":
		return "video/quicktime"
	case ".mkv":
		return "video/x-matroska"
	case ".webm":
		return "video/webm"
	default:
		return ""
	}
}

func pickFirstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

type generatedContent struct {
	Text  string
	Calls []types.FunctionCall
	Raw   []byte
}

func applyResolvedHeaders(headers http.Header, resolved map[string]string) {
	for key, value := range resolved {
		if headers.Get(key) != "" {
			continue
		}
		headers.Set(key, value)
	}
}
