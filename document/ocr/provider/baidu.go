package provider

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/wsnacj/agentx-go/document/internal/credentialhttp"
	"github.com/wsnacj/agentx-go/document/internal/logging"
	"github.com/wsnacj/agentx-go/document/ocr/config"

	"go.uber.org/zap"
)

const (
	baiduDefaultEndpoint = "https://aip.baidubce.com/rest/2.0/ocr/v1/accurate"
	baiduDefaultTokenURL = "https://aip.baidubce.com/oauth/2.0/token"
)

type baiduProvider struct {
	cfg       config.ProviderConfig
	client    *http.Client
	tokenURL  string
	apiKey    string
	secretKey string

	mu          sync.Mutex
	accessToken string
	expiresAt   time.Time
}

// NewBaiduProvider constructs a Baidu OCR provider.
func NewBaiduProvider(cfg config.ProviderConfig) (Provider, error) {
	endpoint := strings.TrimSpace(cfg.BaseURL)
	if endpoint == "" {
		endpoint = baiduDefaultEndpoint
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	auth := cfg.Auth
	accessToken := strings.TrimSpace(auth["access_token"])
	apiKey := strings.TrimSpace(auth["api_key"])
	secretKey := strings.TrimSpace(auth["secret_key"])
	tokenURL := baiduDefaultTokenURL
	if cfg.Additional != nil {
		if v, ok := cfg.Additional["token_url"].(string); ok && strings.TrimSpace(v) != "" {
			tokenURL = strings.TrimSpace(v)
		}
	}

	if accessToken == "" && (apiKey == "" || secretKey == "") {
		return nil, fmt.Errorf("baidu provider requires access_token or api_key/secret_key")
	}

	return &baiduProvider{
		cfg: config.ProviderConfig{
			Kind:    cfg.Kind,
			BaseURL: endpoint,
			Headers: cfg.Headers,
		},
		client:      credentialhttp.NewClient(timeout),
		tokenURL:    tokenURL,
		apiKey:      apiKey,
		secretKey:   secretKey,
		accessToken: accessToken,
	}, nil
}

func (p *baiduProvider) Call(ctx context.Context, req Request) (Response, error) {
	token, err := p.ensureToken(ctx)
	if err != nil {
		return Response{}, err
	}

	endpoint, err := url.Parse(p.cfg.BaseURL)
	if err != nil {
		return Response{}, fmt.Errorf("baidu provider: parse base_url: %w", err)
	}
	query := endpoint.Query()
	query.Set("access_token", token)
	endpoint.RawQuery = query.Encode()

	form := url.Values{}
	if req.IsRemote {
		form.Set("url", req.FilePath)
	} else {
		data, err := os.ReadFile(req.FilePath)
		if err != nil {
			return Response{}, fmt.Errorf("baidu provider: read file: %w", err)
		}
		form.Set("image", base64.StdEncoding.EncodeToString(data))
	}

	// Attach optional parameters if provided.
	for key, value := range sanitizeBaiduOptions(req.Options) {
		form.Set(key, value)
	}

	body := form.Encode()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), strings.NewReader(body))
	if err != nil {
		return Response{}, fmt.Errorf("baidu provider: new request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	start := time.Now()
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return Response{}, fmt.Errorf("baidu provider: call: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, fmt.Errorf("baidu provider: read response: %w", err)
	}

	logger.DebugWithFields("baidu provider call",
		zap.String("url", endpoint.String()),
		zap.Int("status", resp.StatusCode),
		zap.Duration("duration", time.Since(start)),
	)

	if resp.StatusCode != http.StatusOK {
		return Response{}, parseBaiduError(resp.StatusCode, data)
	}

	if err := detectBaiduAPIError(data); err != nil {
		return Response{}, err
	}

	return Response{Raw: data, MediaType: resp.Header.Get("Content-Type")}, nil
}

func (p *baiduProvider) ensureToken(ctx context.Context) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.accessToken != "" && time.Now().Before(p.expiresAt.Add(-1*time.Minute)) {
		return p.accessToken, nil
	}

	// If static token provided and no API key, just use it as-is.
	if p.accessToken != "" && (p.apiKey == "" || p.secretKey == "") {
		return p.accessToken, nil
	}

	if p.apiKey == "" || p.secretKey == "" {
		return "", fmt.Errorf("baidu provider: no credentials to fetch access token")
	}

	tokenURL, err := url.Parse(p.tokenURL)
	if err != nil {
		return "", fmt.Errorf("baidu provider: parse token url: %w", err)
	}
	query := tokenURL.Query()
	query.Set("grant_type", "client_credentials")
	query.Set("client_id", p.apiKey)
	query.Set("client_secret", p.secretKey)
	tokenURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, tokenURL.String(), nil)
	if err != nil {
		return "", fmt.Errorf("baidu provider: new token request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("baidu provider: token request: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("baidu provider: token read: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("baidu provider: token status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var parsed struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("baidu provider: token decode: %w", err)
	}
	if parsed.AccessToken == "" {
		return "", fmt.Errorf("baidu provider: token error %s: %s", parsed.Error, parsed.ErrorDesc)
	}
	p.accessToken = parsed.AccessToken
	if parsed.ExpiresIn > 0 {
		p.expiresAt = time.Now().Add(time.Duration(parsed.ExpiresIn) * time.Second)
	} else {
		p.expiresAt = time.Now().Add(30 * time.Minute)
	}
	return p.accessToken, nil
}

func sanitizeBaiduOptions(opts map[string]any) map[string]string {
	if len(opts) == 0 {
		return nil
	}
	allowed := map[string]struct{}{
		"language_type":              {},
		"eng_granularity":            {},
		"recognize_granularity":      {},
		"detect_direction":           {},
		"vertexes_location":          {},
		"paragraph":                  {},
		"probability":                {},
		"char_probability":           {},
		"multidirectional_recognize": {},
		"return_excel":               {},
		"cell_contents":              {},
		"pdf_file_num":               {},
		"ofd_file":                   {},
		"ofd_file_num":               {},
	}
	result := make(map[string]string)
	for k, v := range opts {
		if _, ok := allowed[k]; !ok {
			continue
		}
		result[k] = fmt.Sprint(v)
	}
	return result
}

func parseBaiduError(status int, body []byte) error {
	var apiErr struct {
		ErrorCode int    `json:"error_code"`
		ErrorMsg  string `json:"error_msg"`
	}
	if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.ErrorCode != 0 {
		return &BaiduError{Status: status, Code: apiErr.ErrorCode, Message: apiErr.ErrorMsg, Raw: strings.TrimSpace(string(body))}
	}
	return &BaiduError{Status: status, Message: strings.TrimSpace(string(body))}
}

func detectBaiduAPIError(body []byte) error {
	var apiErr struct {
		ErrorCode int    `json:"error_code"`
		ErrorMsg  string `json:"error_msg"`
	}
	if err := json.Unmarshal(body, &apiErr); err == nil {
		if apiErr.ErrorCode != 0 {
			return &BaiduError{Status: http.StatusBadRequest, Code: apiErr.ErrorCode, Message: apiErr.ErrorMsg, Raw: strings.TrimSpace(string(body))}
		}
	}
	return nil
}

// BaiduError represents errors returned by Baidu OCR API.
type BaiduError struct {
	Status  int
	Code    int
	Message string
	Raw     string
}

func (e *BaiduError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Code != 0 {
		return fmt.Sprintf("baidu error status=%d code=%d msg=%s", e.Status, e.Code, e.Message)
	}
	return fmt.Sprintf("baidu error status=%d msg=%s", e.Status, e.Message)
}

func (e *BaiduError) ErrorCategory() string {
	if e == nil {
		return ""
	}
	if e.Status >= 500 {
		return "server"
	}
	if e.Status == http.StatusUnauthorized || e.Status == http.StatusForbidden {
		return "auth"
	}
	if e.Status >= 400 {
		return "client"
	}
	return "baidu"
}
