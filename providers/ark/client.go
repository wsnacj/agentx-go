package ark

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/wsnacj/agentx-go/providers/ark/types"
	"github.com/wsnacj/agentx-go/providers/transport"
)

const (
	defaultTimeout       = 60 * time.Second
	defaultStreamTimeout = 5 * time.Minute
)

// Client handles raw HTTP requests to Ark endpoints.
type Client struct {
	transport     transport.Config
	authorize     Authorizer
	baseURL       string
	httpClient    HTTPDoer
	timeout       time.Duration
	streamTimeout time.Duration
}

// New builds a client from explicit host-supplied configuration.
func New(cfg Config) *Client {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://ark.cn-beijing.volces.com/api/v3"
	}
	timeout := defaultTimeout
	streamTimeout := defaultStreamTimeout
	if cfg.Timeout > 0 {
		timeout = cfg.Timeout
	}
	if cfg.StreamTimeout > 0 {
		streamTimeout = cfg.StreamTimeout
	} else if cfg.Timeout > 0 {
		streamTimeout = cfg.Timeout
	}
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{}
	}
	return &Client{
		transport:     transport.Config{Mode: cfg.Transport.Mode, Headers: cloneHeaders(cfg.Transport.Headers)},
		authorize:     cfg.Authorize,
		baseURL:       strings.TrimRight(baseURL, "/"),
		httpClient:    client,
		timeout:       timeout,
		streamTimeout: streamTimeout,
	}
}

// APIError represents non-2xx HTTP errors.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("ark api error: status %d: %s", e.StatusCode, e.Body)
}

// HTTPStatusCode returns the transport status code for shared fault classification.
func (e *APIError) HTTPStatusCode() int {
	if e == nil {
		return 0
	}
	return e.StatusCode
}

// HTTPResponseBody returns the raw response body for shared fault classification.
func (e *APIError) HTTPResponseBody() string {
	if e == nil {
		return ""
	}
	return e.Body
}

// DoJSON performs one authenticated JSON request against a relative Ark API path.
func (c *Client) DoJSON(ctx context.Context, method, path string, payload any, out any) error {
	reqCtx := ctx
	cancel := func() {}
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		timeout := c.timeout
		if timeout == 0 {
			timeout = defaultTimeout
		}
		reqCtx, cancel = context.WithTimeout(ctx, timeout)
	}
	defer cancel()
	settings := transport.ResolveFromContext(reqCtx, c.transport)

	var body io.Reader
	if payload != nil {
		payloadAny, err := transport.ApplyPayloadHook(reqCtx, settings, payload)
		if err != nil {
			return fmt.Errorf("apply payload hook: %w", err)
		}
		data, err := json.Marshal(payloadAny)
		if err != nil {
			return fmt.Errorf("marshal payload: %w", err)
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(reqCtx, method, c.baseURL+path, body)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	if err := c.applyJSONHeaders(req, payload, settings); err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()
	if err := transport.ApplyResponseHook(reqCtx, settings, transport.ResponseMetadataFromHTTP(method, c.baseURL+path, resp)); err != nil {
		return fmt.Errorf("apply response hook: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return &APIError{StatusCode: resp.StatusCode, Body: string(raw)}
	}

	if out == nil {
		return nil
	}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

// DoStream opens one authenticated Ark SSE response body.
func (c *Client) DoStream(ctx context.Context, path string, payload any) (io.ReadCloser, func(), error) {
	settings := transport.ResolveFromContext(ctx, c.transport)
	payloadAny, err := transport.ApplyPayloadHook(ctx, settings, payload)
	if err != nil {
		return nil, nil, fmt.Errorf("apply payload hook: %w", err)
	}
	data, err := json.Marshal(payloadAny)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal payload: %w", err)
	}

	streamCtx := ctx
	cancel := func() {}
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		timeout := c.streamTimeout
		if timeout == 0 {
			timeout = defaultStreamTimeout
		}
		streamCtx, cancel = context.WithTimeout(ctx, timeout)
	} else {
		streamCtx, cancel = context.WithCancel(ctx)
	}

	req, err := http.NewRequestWithContext(streamCtx, http.MethodPost, c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		cancel()
		return nil, nil, fmt.Errorf("build request: %w", err)
	}
	if err := c.applyStreamHeaders(req, payload, settings); err != nil {
		cancel()
		return nil, nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		cancel()
		return nil, nil, fmt.Errorf("send request: %w", err)
	}
	if err := transport.ApplyResponseHook(streamCtx, settings, transport.ResponseMetadataFromHTTP(http.MethodPost, c.baseURL+path, resp)); err != nil {
		resp.Body.Close()
		cancel()
		return nil, nil, fmt.Errorf("apply response hook: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		cancel()
		return nil, nil, &APIError{StatusCode: resp.StatusCode, Body: string(raw)}
	}
	return resp.Body, cancel, nil
}

// UploadFile uploads one file through the Ark Files API.
func (c *Client) UploadFile(ctx context.Context, req *types.UploadFileRequest) (*types.FileObject, error) {
	if req == nil || req.File == nil {
		return nil, fmt.Errorf("ark upload: file is required")
	}

	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)
	filename := req.Filename
	if filename == "" {
		filename = "upload"
	}
	part, err := writer.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(part, req.File); err != nil {
		return nil, fmt.Errorf("copy file: %w", err)
	}

	if req.Purpose != "" {
		if err := writer.WriteField("purpose", req.Purpose); err != nil {
			return nil, fmt.Errorf("write purpose: %w", err)
		}
	}
	if req.ExpireAt != nil {
		if err := writer.WriteField("expire_at", fmt.Sprintf("%d", *req.ExpireAt)); err != nil {
			return nil, fmt.Errorf("write expire_at: %w", err)
		}
	}
	if req.PreprocessConfigs != nil {
		payload, err := json.Marshal(req.PreprocessConfigs)
		if err != nil {
			return nil, fmt.Errorf("marshal preprocess_configs: %w", err)
		}
		if err := writer.WriteField("preprocess_configs", string(payload)); err != nil {
			return nil, fmt.Errorf("write preprocess_configs: %w", err)
		}
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close multipart: %w", err)
	}
	settings := transport.ResolveFromContext(ctx, c.transport)

	reqHTTP, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/files", buf)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	reqHTTP.Header.Set("Content-Type", writer.FormDataContentType())
	if err := c.applyResolvedHeaders(reqHTTP, settings); err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(reqHTTP)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()
	if err := transport.ApplyResponseHook(ctx, settings, transport.ResponseMetadataFromHTTP(http.MethodPost, c.baseURL+"/files", resp)); err != nil {
		return nil, fmt.Errorf("apply response hook: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return nil, &APIError{StatusCode: resp.StatusCode, Body: string(raw)}
	}

	var out types.FileObject
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &out, nil
}

func extractExtraHeaders(payload any) map[string]string {
	switch value := payload.(type) {
	case types.ResponseRequest:
		return value.ExtraHeaders
	case *types.ResponseRequest:
		if value == nil {
			return nil
		}
		return value.ExtraHeaders
	case types.ImageGenerationRequest:
		return value.ExtraHeaders
	case *types.ImageGenerationRequest:
		if value == nil {
			return nil
		}
		return value.ExtraHeaders
	default:
		return nil
	}
}

func (c *Client) applyJSONHeaders(req *http.Request, payload any, settings transport.Settings) error {
	req.Header.Set("Content-Type", "application/json")
	if err := c.applyResolvedHeaders(req, settings); err != nil {
		return err
	}
	applyExtraHeaders(req.Header, extractExtraHeaders(payload))
	return nil
}

func (c *Client) applyStreamHeaders(req *http.Request, payload any, settings transport.Settings) error {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	if err := c.applyResolvedHeaders(req, settings); err != nil {
		return err
	}
	applyExtraHeaders(req.Header, extractExtraHeaders(payload))
	return nil
}

func (c *Client) applyResolvedHeaders(req *http.Request, settings transport.Settings) error {
	transport.ApplyHeaders(req.Header, settings)
	if c.authorize != nil {
		if err := c.authorize(req.Context(), req.Header); err != nil {
			return err
		}
	}
	return nil
}

func applyExtraHeaders(headers http.Header, extra map[string]string) {
	for key, value := range extra {
		if key == "" || value == "" {
			continue
		}
		headers.Set(key, value)
	}
}

// ListFiles lists Ark file objects.
func (c *Client) ListFiles(ctx context.Context, opts types.ListFileOptions) (*types.FileList, error) {
	query := url.Values{}
	if opts.After != "" {
		query.Set("after", opts.After)
	}
	if opts.Limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", opts.Limit))
	}
	if opts.Purpose != "" {
		query.Set("purpose", opts.Purpose)
	}
	if opts.Order != "" {
		query.Set("order", opts.Order)
	}
	path := "/files"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var out types.FileList
	if err := c.DoJSON(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListInputItems lists input items for one Ark response.
func (c *Client) ListInputItems(ctx context.Context, responseID string, opts types.ListInputOptions) (*types.InputItemList, error) {
	query := url.Values{}
	if opts.After != "" {
		query.Set("after", opts.After)
	}
	if opts.Before != "" {
		query.Set("before", opts.Before)
	}
	if opts.Limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", opts.Limit))
	}
	if opts.Order != "" {
		query.Set("order", opts.Order)
	}
	for _, inc := range opts.Include {
		query.Add("include[]", inc)
	}
	path := fmt.Sprintf("/responses/%s/input_items", responseID)
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var out types.InputItemList
	if err := c.DoJSON(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
