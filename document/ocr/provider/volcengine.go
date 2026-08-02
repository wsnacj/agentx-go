package provider

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/wsnacj/agentx-go/document/internal/credentialhttp"
	"github.com/wsnacj/agentx-go/document/internal/logging"
	"github.com/wsnacj/agentx-go/document/ocr/config"

	"go.uber.org/zap"
)

const (
	defaultVolcAction  = "OCRNormal"
	defaultVolcVersion = "2020-08-26"
	defaultVolcRegion  = "cn-north-1"
	defaultVolcService = "cv"
)

type volcEngineProvider struct {
	cfg           config.ProviderConfig
	client        *http.Client
	accessKey     string
	secretKey     string
	securityToken string
	region        string
	service       string
	action        string
	version       string
}

// NewVolcEngineProvider 构造火山引擎 OCR Provider。
func NewVolcEngineProvider(cfg config.ProviderConfig) (Provider, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("volcengine provider requires base_url")
	}
	if cfg.Auth == nil {
		return nil, fmt.Errorf("volcengine provider requires auth")
	}
	accessKey := strings.TrimSpace(cfg.Auth["access_key_id"])
	secretKey := strings.TrimSpace(cfg.Auth["secret_access_key"])
	if accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("volcengine provider requires auth.access_key_id and auth.secret_access_key")
	}
	securityToken := strings.TrimSpace(cfg.Auth["security_token"])

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 15 * time.Second
	}

	action := readAdditionalString(cfg.Additional, "action", defaultVolcAction)
	version := readAdditionalString(cfg.Additional, "version", defaultVolcVersion)
	region := readAdditionalString(cfg.Additional, "region", defaultVolcRegion)
	service := readAdditionalString(cfg.Additional, "service", defaultVolcService)

	return &volcEngineProvider{
		cfg:           cfg,
		client:        credentialhttp.NewClient(timeout),
		accessKey:     accessKey,
		secretKey:     secretKey,
		securityToken: securityToken,
		action:        action,
		version:       version,
		region:        region,
		service:       service,
	}, nil
}

func (p *volcEngineProvider) Call(ctx context.Context, req Request) (Response, error) {
	if req.FilePath == "" {
		return Response{}, fmt.Errorf("volcengine provider: file path empty")
	}

	endpoint, err := url.Parse(p.cfg.BaseURL)
	if err != nil {
		return Response{}, fmt.Errorf("volcengine provider: parse base_url: %w", err)
	}
	query := endpoint.Query()
	query.Set("Action", p.action)
	query.Set("Version", p.version)
	endpoint.RawQuery = query.Encode()

	form := url.Values{}
	if req.IsRemote {
		form.Set("image_url", req.FilePath)
	} else {
		data, err := os.ReadFile(req.FilePath)
		if err != nil {
			return Response{}, fmt.Errorf("volcengine provider: read file: %w", err)
		}
		form.Set("image_base64", base64.StdEncoding.EncodeToString(data))
	}

	applyOptionalParam := func(key string) {
		if req.Options == nil {
			return
		}
		if val, ok := req.Options[key]; ok {
			if str := optionToString(val); str != "" {
				form.Set(key, str)
			}
		}
	}
	applyOptionalParam("approximate_pixel")
	applyOptionalParam("mode")
	applyOptionalParam("filter_thresh")
	applyOptionalParam("half_to_full")

	body := form.Encode()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), strings.NewReader(body))
	if err != nil {
		return Response{}, fmt.Errorf("volcengine provider: new request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Host", endpoint.Host)
	httpReq.Host = endpoint.Host
	xDate := time.Now().UTC().Format("20060102T150405Z")
	shortDate := xDate[:8]
	httpReq.Header.Set("X-Date", xDate)
	if p.securityToken != "" {
		httpReq.Header.Set("X-Security-Token", p.securityToken)
	}
	for k, v := range p.cfg.Headers {
		httpReq.Header.Set(k, v)
	}

	payloadHash := sha256Hex([]byte(body))
	httpReq.Header.Set("X-Content-Sha256", payloadHash)
	auth, err := p.buildAuthorization(httpReq, payloadHash, shortDate, xDate)
	if err != nil {
		return Response{}, err
	}
	httpReq.Header.Set("Authorization", auth)

	start := time.Now()
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return Response{}, fmt.Errorf("volcengine provider: call: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, fmt.Errorf("volcengine provider: read response: %w", err)
	}

	logger.DebugWithFields("volcengine provider call",
		zap.String("url", endpoint.String()),
		zap.Int("status", resp.StatusCode),
		zap.Duration("duration", time.Since(start)),
	)

	if resp.StatusCode != http.StatusOK {
		return Response{}, &VolcError{
			Status:  resp.StatusCode,
			Message: strings.TrimSpace(string(data)),
			Raw:     string(data),
		}
	}

	var parsed struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return Response{}, fmt.Errorf("volcengine provider: parse response: %w", err)
	}
	if parsed.Code != 10000 {
		return Response{}, &VolcError{
			Status:  resp.StatusCode,
			Code:    parsed.Code,
			Message: parsed.Message,
			Raw:     string(data),
		}
	}

	return Response{
		Raw:       data,
		MediaType: resp.Header.Get("Content-Type"),
	}, nil
}

func (p *volcEngineProvider) buildAuthorization(req *http.Request, payloadHash, shortDate, longDate string) (string, error) {
	canonicalURI := canonicalURIPath(req.URL)
	canonicalQuery := canonicalQueryString(req.URL)

	signedHeaders := collectSignedHeaders(req.Header)

	canonicalHeaders := buildCanonicalHeaders(req.Header, signedHeaders)
	canonicalRequest := strings.Join([]string{
		req.Method,
		canonicalURI,
		canonicalQuery,
		canonicalHeaders,
		"",
		strings.Join(signedHeaders, ";"),
		payloadHash,
	}, "\n")

	hashCanonical := sha256Hex([]byte(canonicalRequest))
	credentialScope := fmt.Sprintf("%s/%s/%s/request", shortDate, p.region, p.service)
	stringToSign := strings.Join([]string{
		"HMAC-SHA256",
		longDate,
		credentialScope,
		hashCanonical,
	}, "\n")

	signingKey := deriveVolcSigningKey(p.secretKey, shortDate, p.region, p.service)
	signature := hex.EncodeToString(hmacSHA256(signingKey, stringToSign))

	logger.DebugWithFields("volcengine signer debug",
		zap.String("canonical_request", canonicalRequest),
		zap.String("string_to_sign", stringToSign),
		zap.String("signed_headers", strings.Join(signedHeaders, ";")),
		zap.String("payload_hash", payloadHash),
		zap.String("signature", signature),
	)

	return fmt.Sprintf("HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		p.accessKey,
		credentialScope,
		strings.Join(signedHeaders, ";"),
		signature,
	), nil
}

func buildCanonicalHeaders(h http.Header, keys []string) string {
	var lines []string
	for _, key := range keys {
		canonicalKey := http.CanonicalHeaderKey(key)
		values := h.Values(canonicalKey)
		if len(values) == 0 {
			continue
		}
		joined := strings.Join(values, ",")
		joined = normalizeSpaces(joined)
		value := strings.TrimSpace(joined)
		if strings.EqualFold(canonicalKey, "host") {
			if idx := strings.LastIndex(value, ":"); idx > -1 {
				port := value[idx+1:]
				if port == "80" || port == "443" {
					value = value[:idx]
				}
			}
		}
		lines = append(lines, fmt.Sprintf("%s:%s", strings.ToLower(canonicalKey), value))
	}
	sort.Strings(lines)
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

func canonicalURIPath(u *url.URL) string {
	if u.Path == "" {
		return "/"
	}
	return encodePath(u.EscapedPath())
}

func canonicalQueryString(u *url.URL) string {
	if u.RawQuery == "" {
		return ""
	}
	values, _ := url.ParseQuery(u.RawQuery)
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		vals := values[k]
		sort.Strings(vals)
		encodedKey := percentEncode(k)
		for _, v := range vals {
			parts = append(parts, fmt.Sprintf("%s=%s", encodedKey, percentEncode(v)))
		}
	}
	return strings.Join(parts, "&")
}

func encodePath(path string) string {
	if path == "" {
		return "/"
	}
	var builder strings.Builder
	for i := 0; i < len(path); i++ {
		ch := path[i]
		if isUnreserved(ch) || ch == '/' {
			builder.WriteByte(ch)
			continue
		}
		if ch == '%' && i+2 < len(path) {
			builder.WriteByte(ch)
			builder.WriteByte(path[i+1])
			builder.WriteByte(path[i+2])
			i += 2
			continue
		}
		builder.WriteString(fmt.Sprintf("%%%02X", ch))
	}
	encoded := builder.String()
	if encoded == "" {
		return "/"
	}
	if encoded[0] != '/' {
		return "/" + encoded
	}
	return encoded
}

func normalizeSpaces(s string) string {
	if s == "" {
		return ""
	}
	return strings.Join(strings.Fields(s), " ")
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func hmacSHA256(key []byte, data string) []byte {
	m := hmac.New(sha256.New, key)
	if _, err := m.Write([]byte(data)); err != nil {
		return nil
	}
	return m.Sum(nil)
}

func deriveVolcSigningKey(secret, shortDate, region, service string) []byte {
	kDate := hmacSHA256([]byte(secret), shortDate)
	kRegion := hmacSHA256(kDate, region)
	kService := hmacSHA256(kRegion, service)
	return hmacSHA256(kService, "request")
}

func optionToString(val any) string {
	switch v := val.(type) {
	case string:
		return strings.TrimSpace(v)
	case fmt.Stringer:
		return strings.TrimSpace(v.String())
	case bool:
		if v {
			return "true"
		}
		return "false"
	case int:
		return fmt.Sprintf("%d", v)
	case int32:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case float32:
		return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%f", v), "0"), ".")
	case float64:
		return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%f", v), "0"), ".")
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}

func readAdditionalString(additional map[string]any, key, fallback string) string {
	if additional == nil {
		return fallback
	}
	raw, ok := additional[key]
	if !ok {
		return fallback
	}
	switch v := raw.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return fallback
		}
		return strings.TrimSpace(v)
	case time.Time:
		return v.Format("2006-01-02")
	case *time.Time:
		if v == nil {
			return fallback
		}
		return v.Format("2006-01-02")
	case fmt.Stringer:
		return strings.TrimSpace(v.String())
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}

func percentEncode(s string) string {
	if s == "" {
		return ""
	}
	var builder strings.Builder
	builder.Grow(len(s) * 3)
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if isUnreserved(ch) {
			builder.WriteByte(ch)
		} else {
			builder.WriteString(fmt.Sprintf("%%%02X", ch))
		}
	}
	return builder.String()
}

func isUnreserved(b byte) bool {
	return b >= 'A' && b <= 'Z' ||
		b >= 'a' && b <= 'z' ||
		b >= '0' && b <= '9' ||
		b == '-' || b == '_' || b == '.' || b == '~'
}

// VolcError 表示 Volcengine OCR 返回的错误。

func collectSignedHeaders(header http.Header) []string {
	var result []string
	for name := range header {
		lower := strings.ToLower(name)
		if lower == "content-type" || lower == "content-md5" || lower == "host" || strings.HasPrefix(lower, "x-") {
			result = append(result, lower)
		}
	}
	sort.Strings(result)
	return result
}

type VolcError struct {
	Status  int
	Code    int
	Message string
	Raw     string
}

func (e *VolcError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Code != 0 {
		return fmt.Sprintf("volcengine error status=%d code=%d msg=%s", e.Status, e.Code, e.Message)
	}
	return fmt.Sprintf("volcengine error status=%d msg=%s", e.Status, e.Message)
}

// ErrorCategory 实现 ErrorCategorizer 接口。
func (e *VolcError) ErrorCategory() string {
	if e == nil {
		return ""
	}
	if e.Status >= 500 {
		return "server"
	}
	if e.Status >= 400 {
		return "client"
	}
	return "volcengine"
}
