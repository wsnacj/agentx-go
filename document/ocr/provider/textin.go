package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/wsnacj/agentx-go/document/internal/credentialhttp"
	"github.com/wsnacj/agentx-go/document/internal/logging"
	"github.com/wsnacj/agentx-go/document/ocr/config"
	"go.uber.org/zap"
)

type textInProvider struct {
	cfg    config.ProviderConfig
	client *http.Client
}

func newTextIn(cfg config.ProviderConfig) (Provider, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("textin provider requires base_url")
	}
	appID := cfg.Auth["app_id"]
	secret := cfg.Auth["secret_code"]
	if appID == "" || secret == "" {
		return nil, fmt.Errorf("textin provider requires auth app_id and secret_code")
	}
	if err := validateTextInHeaders(cfg.Headers); err != nil {
		return nil, err
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	return &textInProvider{
		cfg:    cfg,
		client: credentialhttp.NewClient(timeout),
	}, nil
}

func (p *textInProvider) Call(ctx context.Context, req Request) (Response, error) {
	endpoint, err := p.buildURL(req)
	if err != nil {
		return Response{}, err
	}

	body, contentType, err := p.buildBody(req)
	if err != nil {
		return Response{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return Response{}, fmt.Errorf("textin request: %w", err)
	}
	httpReq.Header.Set("Content-Type", contentType)
	for k, v := range p.cfg.Headers {
		httpReq.Header.Set(k, v)
	}
	httpReq.Header.Set("x-ti-app-id", p.cfg.Auth["app_id"])
	httpReq.Header.Set("x-ti-secret-code", p.cfg.Auth["secret_code"])

	start := time.Now()
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return Response{}, fmt.Errorf("textin call: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, fmt.Errorf("textin read response: %w", err)
	}

	logger.DebugWithFields("textin provider call",
		zap.String("url", endpoint),
		zap.Int("status", resp.StatusCode),
		zap.Duration("duration", time.Since(start)),
	)

	if resp.StatusCode != http.StatusOK {
		var apiErr struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		msg := strings.TrimSpace(string(data))
		errCategory := classifyStatus(resp.StatusCode)
		if err := json.Unmarshal(data, &apiErr); err == nil && apiErr.Code != 0 {
			return Response{}, &TextInError{Status: resp.StatusCode, Code: apiErr.Code, Message: apiErr.Message, Raw: msg, Category: errCategory}
		}
		return Response{}, &TextInError{Status: resp.StatusCode, Message: msg, Raw: msg, Category: errCategory}
	}

	return Response{Raw: data, MediaType: resp.Header.Get("Content-Type")}, nil
}

func validateTextInHeaders(headers map[string]string) error {
	for name := range headers {
		switch strings.ToLower(strings.TrimSpace(name)) {
		case "x-ti-app-id", "x-ti-secret-code":
			return fmt.Errorf("textin provider header %q is reserved for configured authentication", name)
		}
	}
	return nil
}

func classifyStatus(status int) string {
	if status >= 500 {
		return "server"
	}
	if status >= 400 {
		return "client"
	}
	return "http"
}

func (p *textInProvider) buildURL(req Request) (string, error) {
	parsed, err := url.Parse(p.cfg.BaseURL)
	if err != nil {
		return "", fmt.Errorf("textin parse base_url: %w", err)
	}
	if req.NeedCharacter {
		q := parsed.Query()
		q.Set("character", "1")
		parsed.RawQuery = q.Encode()
	}
	return parsed.String(), nil
}

func (p *textInProvider) buildBody(req Request) ([]byte, string, error) {
	if req.IsRemote {
		return []byte(req.FilePath), "text/plain", nil
	}
	data, err := os.ReadFile(req.FilePath)
	if err != nil {
		return nil, "", fmt.Errorf("textin read file: %w", err)
	}
	return data, "application/octet-stream", nil
}
