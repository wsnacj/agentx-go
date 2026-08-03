package gemini

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/wsnacj/agentx-go/providers"
	"github.com/wsnacj/agentx-go/providers/transport"
)

const (
	defaultGeminiBaseURL       = "https://generativelanguage.googleapis.com/v1beta"
	defaultGeminiUploadBaseURL = "https://generativelanguage.googleapis.com/upload/v1beta"
	streamScannerCap           = 64 * 1024
	streamScannerMax           = 1024 * 1024
)

// Client performs Gemini native API calls.
type Client struct {
	baseURL       string
	uploadBaseURL string
	transport     transport.Config
	authorize     Authorizer
	httpClient    HTTPDoer
}

// New constructs a Gemini client without reading credentials or making network calls.
func New(cfg Config) *Client {
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: cfg.Timeout}
	}
	return &Client{
		baseURL:       strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/"),
		uploadBaseURL: strings.TrimRight(strings.TrimSpace(cfg.UploadBaseURL), "/"),
		transport:     transport.Config{Mode: cfg.Transport.Mode, Headers: cloneHeaders(cfg.Transport.Headers)},
		authorize:     cfg.Authorize,
		httpClient:    client,
	}
}

// BaseURL returns the API base URL.
func (c *Client) BaseURL() string {
	base := c.baseURL
	if base == "" {
		base = defaultGeminiBaseURL
	}
	return base
}

// GenerateContent calls models.generateContent.
func (c *Client) GenerateContent(ctx context.Context, model string, req *GenerateContentRequest) (*GenerateContentResponse, error) {
	payload := req
	if payload == nil {
		payload = &GenerateContentRequest{}
	}
	endpoint := c.buildModelEndpoint(model, "generateContent")
	settings := transport.ResolveFromContext(ctx, c.transport)
	body, err := c.postJSON(ctx, endpoint, payload, settings)
	if err != nil {
		return nil, err
	}
	var resp GenerateContentResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode generate response: %w", err)
	}
	return &resp, nil
}

// StreamGenerateContent streams models.streamGenerateContent.
func (c *Client) StreamGenerateContent(ctx context.Context, model string, req *GenerateContentRequest) (<-chan StreamChunk, func(), error) {
	payload := req
	if payload == nil {
		payload = &GenerateContentRequest{}
	}
	settings := transport.ResolveFromContext(ctx, c.transport)
	endpoint := addAltSSE(c.buildModelEndpoint(model, "streamGenerateContent"))
	payloadAny, err := transport.ApplyPayloadHook(ctx, settings, payload)
	if err != nil {
		return nil, nil, fmt.Errorf("apply payload hook: %w", err)
	}
	data, err := json.Marshal(payloadAny)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal payload: %w", err)
	}

	streamCtx := ctx
	var cancel context.CancelFunc
	if _, hasDeadline := ctx.Deadline(); hasDeadline {
		streamCtx, cancel = context.WithCancel(ctx)
	} else {
		streamCtx, cancel = context.WithTimeout(ctx, 5*time.Minute)
	}

	reqHTTP, err := http.NewRequestWithContext(streamCtx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		cancel()
		return nil, nil, fmt.Errorf("build request: %w", err)
	}

	if err := c.applyHeaders(reqHTTP, settings); err != nil {
		cancel()
		return nil, nil, err
	}

	resp, err := c.httpClient.Do(reqHTTP)
	if err != nil {
		cancel()
		return nil, nil, fmt.Errorf("send request: %w", err)
	}
	if err := transport.ApplyResponseHook(streamCtx, settings, transport.ResponseMetadataFromHTTP(http.MethodPost, endpoint, resp)); err != nil {
		resp.Body.Close()
		cancel()
		return nil, nil, fmt.Errorf("apply response hook: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		cancel()
		return nil, nil, &providers.APIError{StatusCode: resp.StatusCode, Body: string(raw)}
	}

	ch := make(chan StreamChunk, 16)
	go func() {
		defer close(ch)
		defer resp.Body.Close()
		defer cancel()

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, streamScannerCap), streamScannerMax)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			if !strings.HasPrefix(line, "data:") {
				continue
			}
			payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if payload == "" || payload == "[DONE]" {
				return
			}
			var resp GenerateContentResponse
			if err := json.Unmarshal([]byte(payload), &resp); err != nil {
				ch <- StreamChunk{Err: fmt.Errorf("decode stream payload: %w", err)}
				return
			}
			ch <- StreamChunk{Response: &resp, Raw: []byte(payload)}
		}
		if err := scanner.Err(); err != nil {
			ch <- StreamChunk{Err: fmt.Errorf("stream read error: %w", err)}
		}
	}()

	return ch, func() {
		cancel()
		resp.Body.Close()
	}, nil
}

// EmbedContent calls models.embedContent.
func (c *Client) EmbedContent(ctx context.Context, model string, req *EmbedContentRequest) (*EmbedContentResponse, error) {
	payload := req
	if payload == nil {
		payload = &EmbedContentRequest{}
	}
	endpoint := c.buildModelEndpoint(model, "embedContent")
	settings := transport.ResolveFromContext(ctx, c.transport)
	body, err := c.postJSON(ctx, endpoint, payload, settings)
	if err != nil {
		return nil, err
	}
	var resp EmbedContentResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode embed response: %w", err)
	}
	return &resp, nil
}

// BatchEmbedContents calls models.batchEmbedContents.
func (c *Client) BatchEmbedContents(ctx context.Context, model string, req *BatchEmbedContentsRequest) (*BatchEmbedContentsResponse, error) {
	payload := req
	if payload == nil {
		payload = &BatchEmbedContentsRequest{}
	}
	endpoint := c.buildModelEndpoint(model, "batchEmbedContents")
	settings := transport.ResolveFromContext(ctx, c.transport)
	body, err := c.postJSON(ctx, endpoint, payload, settings)
	if err != nil {
		return nil, err
	}
	var resp BatchEmbedContentsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode batch embed response: %w", err)
	}
	return &resp, nil
}

// UploadFile uploads a file using the resumable protocol.
func (c *Client) UploadFile(ctx context.Context, displayName string, mimeType string, size int64, data io.Reader) (*File, error) {
	if size <= 0 {
		return nil, fmt.Errorf("invalid file size")
	}
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	startURL := c.buildUploadEndpoint()
	startReq := map[string]any{"file": map[string]any{"displayName": displayName}}
	settings := transport.ResolveFromContext(ctx, c.transport)
	startPayload, err := transport.ApplyPayloadHook(ctx, settings, startReq)
	if err != nil {
		return nil, fmt.Errorf("apply payload hook: %w", err)
	}
	startBody, err := json.Marshal(startPayload)
	if err != nil {
		return nil, fmt.Errorf("marshal upload start: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, startURL, bytes.NewReader(startBody))
	if err != nil {
		return nil, fmt.Errorf("build upload request: %w", err)
	}

	if err := c.applyHeaders(req, settings); err != nil {
		return nil, err
	}
	req.Header.Set("X-Goog-Upload-Protocol", "resumable")
	req.Header.Set("X-Goog-Upload-Command", "start")
	req.Header.Set("X-Goog-Upload-Header-Content-Length", fmt.Sprintf("%d", size))
	req.Header.Set("X-Goog-Upload-Header-Content-Type", mimeType)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("start upload: %w", err)
	}
	defer resp.Body.Close()
	if err := transport.ApplyResponseHook(ctx, settings, transport.ResponseMetadataFromHTTP(http.MethodPost, startURL, resp)); err != nil {
		return nil, fmt.Errorf("apply response hook: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return nil, &providers.APIError{StatusCode: resp.StatusCode, Body: string(raw)}
	}

	uploadURL := resp.Header.Get("X-Goog-Upload-URL")
	if uploadURL == "" {
		return nil, fmt.Errorf("missing upload url")
	}

	uploadReq, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, data)
	if err != nil {
		return nil, fmt.Errorf("build upload finalize: %w", err)
	}
	uploadReq.Header.Set("X-Goog-Upload-Offset", "0")
	uploadReq.Header.Set("X-Goog-Upload-Command", "upload, finalize")
	uploadReq.Header.Set("Content-Length", fmt.Sprintf("%d", size))

	uploadResp, err := c.httpClient.Do(uploadReq)
	if err != nil {
		return nil, fmt.Errorf("finalize upload: %w", err)
	}
	defer uploadResp.Body.Close()
	if err := transport.ApplyResponseHook(ctx, settings, transport.ResponseMetadataFromHTTP(http.MethodPost, uploadURL, uploadResp)); err != nil {
		return nil, fmt.Errorf("apply response hook: %w", err)
	}

	raw, err := io.ReadAll(uploadResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read upload response: %w", err)
	}
	if uploadResp.StatusCode != http.StatusOK {
		return nil, &providers.APIError{StatusCode: uploadResp.StatusCode, Body: string(raw)}
	}

	var out FileResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decode upload response: %w", err)
	}
	return out.File, nil
}

// GetFile fetches a file by name (e.g. "files/xxx").
func (c *Client) GetFile(ctx context.Context, name string) (*File, error) {
	if name == "" {
		return nil, fmt.Errorf("empty file name")
	}
	endpoint := c.buildFilesEndpoint(name)
	resp, err := c.getJSON(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	var file File
	if err := json.Unmarshal(resp, &file); err != nil {
		return nil, fmt.Errorf("decode file response: %w", err)
	}
	return &file, nil
}

// DeleteFile deletes a file by name.
func (c *Client) DeleteFile(ctx context.Context, name string) error {
	if name == "" {
		return fmt.Errorf("empty file name")
	}
	endpoint := c.buildFilesEndpoint(name)
	settings := transport.ResolveFromContext(ctx, c.transport)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return fmt.Errorf("build delete request: %w", err)
	}
	if err := c.applyHeaders(req, settings); err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("delete request: %w", err)
	}
	defer resp.Body.Close()
	if err := transport.ApplyResponseHook(ctx, settings, transport.ResponseMetadataFromHTTP(http.MethodDelete, endpoint, resp)); err != nil {
		return fmt.Errorf("apply response hook: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return &providers.APIError{StatusCode: resp.StatusCode, Body: string(raw)}
	}
	return nil
}

func (c *Client) buildModelEndpoint(model string, action string) string {
	resource := normalizeModelName(model)
	base := strings.TrimRight(c.BaseURL(), "/")
	return base + "/" + resource + ":" + action
}

func (c *Client) buildUploadEndpoint() string {
	base := strings.TrimRight(c.BaseURL(), "/")
	if strings.Contains(base, "/openai") {
		base = strings.TrimSuffix(base, "/openai")
	}
	uploadBase := c.uploadBaseURL
	if uploadBase == "" {
		uploadBase = defaultGeminiUploadBaseURL
	}
	if strings.Contains(base, "/v1beta") {
		uploadBase = strings.TrimSuffix(base, "/v1beta") + "/upload/v1beta"
	}
	return strings.TrimRight(uploadBase, "/") + "/files"
}

func (c *Client) buildFilesEndpoint(name string) string {
	base := strings.TrimRight(c.BaseURL(), "/")
	return base + "/" + strings.TrimLeft(name, "/")
}

func normalizeModelName(model string) string {
	name := strings.TrimSpace(model)
	if name == "" {
		return "models/gemini-3-flash-preview"
	}
	if strings.HasPrefix(name, "models/") || strings.HasPrefix(name, "tunedModels/") {
		return name
	}
	return "models/" + name
}

func addAltSSE(endpoint string) string {
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return endpoint + "?alt=sse"
	}
	q := parsed.Query()
	q.Set("alt", "sse")
	parsed.RawQuery = q.Encode()
	return parsed.String()
}

func (c *Client) postJSON(ctx context.Context, endpoint string, payload any, settings transport.Settings) ([]byte, error) {
	payloadAny, err := transport.ApplyPayloadHook(ctx, settings, payload)
	if err != nil {
		return nil, fmt.Errorf("apply payload hook: %w", err)
	}
	data, err := json.Marshal(payloadAny)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	if err := c.applyHeaders(req, settings); err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
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

func (c *Client) getJSON(ctx context.Context, endpoint string) ([]byte, error) {
	settings := transport.ResolveFromContext(ctx, c.transport)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	if err := c.applyHeaders(req, settings); err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()
	if err := transport.ApplyResponseHook(ctx, settings, transport.ResponseMetadataFromHTTP(http.MethodGet, endpoint, resp)); err != nil {
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

func (c *Client) applyHeaders(req *http.Request, settings transport.Settings) error {
	req.Header.Set("Content-Type", "application/json")
	transport.ApplyHeaders(req.Header, settings)
	if c.authorize != nil {
		if err := c.authorize(req.Context(), req.Header); err != nil {
			return err
		}
	}
	return nil
}

// StreamChunk holds a streaming response payload.
type StreamChunk struct {
	Response *GenerateContentResponse
	Raw      []byte
	Err      error
}
