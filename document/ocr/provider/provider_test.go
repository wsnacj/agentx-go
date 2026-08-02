package provider

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/wsnacj/agentx-go/document/ocr/config"
)

func TestDefaultFactoriesExposeBuiltins(t *testing.T) {
	factories := DefaultFactories()
	for _, key := range []string{"textin", "volcengine", "baidu"} {
		factory, ok := factories[key]
		if !ok {
			t.Fatalf("missing factory %q", key)
		}
		if factory == nil {
			t.Fatalf("factory %q is nil", key)
		}
	}
}

func TestDefaultConfigValidatorsTextInTrimsCredentials(t *testing.T) {
	cfg := config.ProviderConfig{
		Auth: map[string]string{
			"app_id":      " demo-app ",
			"secret_code": " secret ",
		},
	}

	if err := DefaultConfigValidators()["textin"](&cfg); err != nil {
		t.Fatalf("validate textin: %v", err)
	}
	if cfg.Auth["app_id"] != "demo-app" || cfg.Auth["secret_code"] != "secret" {
		t.Fatalf("expected trimmed credentials, got %+v", cfg.Auth)
	}
}

func TestDefaultConfigValidatorsVolcengineSetsDefaults(t *testing.T) {
	cfg := config.ProviderConfig{
		Auth: map[string]string{
			"access_key_id":     " ak ",
			"secret_access_key": " sk ",
			"security_token":    " token ",
		},
	}

	if err := DefaultConfigValidators()["volcengine"](&cfg); err != nil {
		t.Fatalf("validate volcengine: %v", err)
	}
	if cfg.Auth["access_key_id"] != "ak" || cfg.Auth["secret_access_key"] != "sk" || cfg.Auth["security_token"] != "token" {
		t.Fatalf("expected trimmed auth, got %+v", cfg.Auth)
	}
	if cfg.Additional["action"] != defaultVolcAction || cfg.Additional["version"] != defaultVolcVersion {
		t.Fatalf("expected default additional values, got %+v", cfg.Additional)
	}
}

func TestDefaultConfigValidatorsBaiduSetsDefaults(t *testing.T) {
	cfg := config.ProviderConfig{
		Auth: map[string]string{
			"api_key":    " ak ",
			"secret_key": " sk ",
		},
	}

	if err := DefaultConfigValidators()["baidu"](&cfg); err != nil {
		t.Fatalf("validate baidu: %v", err)
	}
	if cfg.BaseURL != baiduDefaultEndpoint {
		t.Fatalf("unexpected default base url: %s", cfg.BaseURL)
	}
	if cfg.Additional["token_url"] != baiduDefaultTokenURL {
		t.Fatalf("unexpected default token url: %+v", cfg.Additional)
	}
	if cfg.Auth["api_key"] != "ak" || cfg.Auth["secret_key"] != "sk" {
		t.Fatalf("expected trimmed credentials, got %+v", cfg.Auth)
	}
}

func TestTextInProviderBuildURLAndBody(t *testing.T) {
	dir := t.TempDir()
	filePath := dir + "/sample.txt"
	if err := osWriteFile(filePath, []byte("demo")); err != nil {
		t.Fatalf("write file: %v", err)
	}

	providerAny, err := newTextIn(config.ProviderConfig{
		BaseURL: "https://example.com/ocr?existing=1",
		Auth: map[string]string{
			"app_id":      "app",
			"secret_code": "secret",
		},
	})
	if err != nil {
		t.Fatalf("new textin provider: %v", err)
	}

	p := providerAny.(*textInProvider)
	urlWithChar, err := p.buildURL(Request{NeedCharacter: true})
	if err != nil {
		t.Fatalf("build url: %v", err)
	}
	if !strings.Contains(urlWithChar, "existing=1") || !strings.Contains(urlWithChar, "character=1") {
		t.Fatalf("unexpected url: %s", urlWithChar)
	}

	body, contentType, err := p.buildBody(Request{FilePath: filePath})
	if err != nil {
		t.Fatalf("build local body: %v", err)
	}
	if string(body) != "demo" || contentType != "application/octet-stream" {
		t.Fatalf("unexpected local body: %q %s", string(body), contentType)
	}

	body, contentType, err = p.buildBody(Request{FilePath: "https://example.com/file.png", IsRemote: true})
	if err != nil {
		t.Fatalf("build remote body: %v", err)
	}
	if string(body) != "https://example.com/file.png" || contentType != "text/plain" {
		t.Fatalf("unexpected remote body: %q %s", string(body), contentType)
	}
}

func TestProviderConstructorsApplyDefaultTimeouts(t *testing.T) {
	textInAny, err := newTextIn(config.ProviderConfig{
		BaseURL: "https://example.com/ocr",
		Auth: map[string]string{
			"app_id":      "app",
			"secret_code": "secret",
		},
	})
	if err != nil {
		t.Fatalf("new textin provider: %v", err)
	}
	if timeout := textInAny.(*textInProvider).client.Timeout; timeout != 30*time.Second {
		t.Fatalf("unexpected textin timeout: %s", timeout)
	}

	baiduAny, err := NewBaiduProvider(config.ProviderConfig{
		Kind: "baidu",
		Auth: map[string]string{"access_token": "token"},
	})
	if err != nil {
		t.Fatalf("new baidu provider: %v", err)
	}
	if timeout := baiduAny.(*baiduProvider).client.Timeout; timeout != 20*time.Second {
		t.Fatalf("unexpected baidu timeout: %s", timeout)
	}

	volcAny, err := NewVolcEngineProvider(config.ProviderConfig{
		BaseURL: "https://visual.volcengineapi.com",
		Auth: map[string]string{
			"access_key_id":     "ak",
			"secret_access_key": "sk",
		},
	})
	if err != nil {
		t.Fatalf("new volc provider: %v", err)
	}
	volc := volcAny.(*volcEngineProvider)
	if volc.client.Timeout != 15*time.Second {
		t.Fatalf("unexpected volc timeout: %s", volc.client.Timeout)
	}
	if volc.action != defaultVolcAction || volc.version != defaultVolcVersion || volc.region != defaultVolcRegion || volc.service != defaultVolcService {
		t.Fatalf("unexpected volc defaults: %+v", volc)
	}
}

func TestProviderConstructorsRejectMissingCredentials(t *testing.T) {
	if _, err := newTextIn(config.ProviderConfig{BaseURL: "https://example.com"}); err == nil {
		t.Fatal("expected textin credentials error")
	}
	if _, err := NewBaiduProvider(config.ProviderConfig{}); err == nil {
		t.Fatal("expected baidu credentials error")
	}
	if _, err := NewVolcEngineProvider(config.ProviderConfig{BaseURL: "https://example.com"}); err == nil {
		t.Fatal("expected volc credentials error")
	}
}

func TestTextInProviderRejectsReservedAuthenticationHeaders(t *testing.T) {
	for _, header := range []string{
		"x-ti-app-id",
		"X-Ti-App-Id",
		" x-ti-secret-code ",
	} {
		t.Run(header, func(t *testing.T) {
			cfg := config.ProviderConfig{
				BaseURL: "https://example.com/ocr",
				Auth: map[string]string{
					"app_id":      "app",
					"secret_code": "secret",
				},
				Headers: map[string]string{header: "must-not-override-auth"},
			}
			if _, err := newTextIn(cfg); err == nil {
				t.Fatalf("expected reserved header %q to be rejected", header)
			}
			if err := DefaultConfigValidators()["textin"](&cfg); err == nil {
				t.Fatalf("expected validator to reject reserved header %q", header)
			}
		})
	}
}

func TestNewTextInProviderFactoryReturnsProvider(t *testing.T) {
	p, err := NewTextInProvider(config.ProviderConfig{
		BaseURL: "https://example.com/ocr",
		Auth: map[string]string{
			"app_id":      "app",
			"secret_code": "secret",
		},
	})
	if err != nil {
		t.Fatalf("new textin provider: %v", err)
	}
	if _, ok := p.(*textInProvider); !ok {
		t.Fatalf("unexpected provider type %T", p)
	}
}

func TestBaiduProviderUsesStaticTokenWithoutRefresh(t *testing.T) {
	pAny, err := NewBaiduProvider(config.ProviderConfig{
		Kind: "baidu",
		Auth: map[string]string{
			"access_token": "static-token",
		},
	})
	if err != nil {
		t.Fatalf("new baidu provider: %v", err)
	}

	token, err := pAny.(*baiduProvider).ensureToken(context.Background())
	if err != nil {
		t.Fatalf("ensure token: %v", err)
	}
	if token != "static-token" {
		t.Fatalf("unexpected token: %s", token)
	}
}

func TestTextInProviderCallSendsHeadersAndReturnsRaw(t *testing.T) {
	dir := t.TempDir()
	filePath := dir + "/sample.bin"
	if err := osWriteFile(filePath, []byte("payload")); err != nil {
		t.Fatalf("write file: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("x-ti-app-id"); got != "app" {
			t.Fatalf("unexpected app id header: %s", got)
		}
		if got := r.Header.Get("x-ti-secret-code"); got != "secret" {
			t.Fatalf("unexpected secret header: %s", got)
		}
		if got := r.Header.Get("X-Custom"); got != "custom" {
			t.Fatalf("unexpected custom header: %s", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/custom" {
			t.Fatalf("unexpected custom content type: %s", got)
		}
		if got := r.URL.Query().Get("character"); got != "1" {
			t.Fatalf("expected character query, got %q", got)
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != "payload" {
			t.Fatalf("unexpected request body: %q", string(body))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":200,"message":"OK"}`))
	}))
	defer server.Close()

	pAny, err := newTextIn(config.ProviderConfig{
		BaseURL: server.URL,
		Auth: map[string]string{
			"app_id":      "app",
			"secret_code": "secret",
		},
		Headers: map[string]string{
			"X-Custom":     "custom",
			"Content-Type": "application/custom",
		},
	})
	if err != nil {
		t.Fatalf("new textin provider: %v", err)
	}

	resp, err := pAny.Call(context.Background(), Request{FilePath: filePath, NeedCharacter: true})
	if err != nil {
		t.Fatalf("textin call: %v", err)
	}
	if string(resp.Raw) != `{"code":200,"message":"OK"}` {
		t.Fatalf("unexpected response raw: %s", string(resp.Raw))
	}
	if resp.MediaType != "application/json" {
		t.Fatalf("unexpected media type: %s", resp.MediaType)
	}
}

func TestTextInProviderCallReturnsCategorizedError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"code":9001,"message":"upstream failed"}`))
	}))
	defer server.Close()

	pAny, err := newTextIn(config.ProviderConfig{
		BaseURL: server.URL,
		Auth: map[string]string{
			"app_id":      "app",
			"secret_code": "secret",
		},
	})
	if err != nil {
		t.Fatalf("new textin provider: %v", err)
	}

	_, err = pAny.Call(context.Background(), Request{FilePath: server.URL, IsRemote: true})
	if err == nil {
		t.Fatal("expected textin error")
	}
	apiErr, ok := err.(*TextInError)
	if !ok {
		t.Fatalf("unexpected error type %T", err)
	}
	if apiErr.Code != 9001 || apiErr.ErrorCategory() != "server" {
		t.Fatalf("unexpected textin error: %+v", apiErr)
	}
}

func TestBaiduProviderCallUsesStaticTokenAndAllowedOptions(t *testing.T) {
	dir := t.TempDir()
	filePath := dir + "/sample.bin"
	if err := osWriteFile(filePath, []byte("hello")); err != nil {
		t.Fatalf("write file: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("access_token"); got != "static-token" {
			t.Fatalf("unexpected access token: %s", got)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		rawImage := r.PostForm.Get("image")
		data, err := base64.StdEncoding.DecodeString(rawImage)
		if err != nil {
			t.Fatalf("decode image: %v", err)
		}
		if string(data) != "hello" {
			t.Fatalf("unexpected image payload: %q", string(data))
		}
		if got := r.PostForm.Get("language_type"); got != "CHN_ENG" {
			t.Fatalf("unexpected allowed option: %s", got)
		}
		if got := r.PostForm.Get("ignored_option"); got != "" {
			t.Fatalf("unexpected ignored option propagated: %s", got)
		}
		_, _ = w.Write([]byte(`{"words_result":[{"words":"ok"}]}`))
	}))
	defer server.Close()

	pAny, err := NewBaiduProvider(config.ProviderConfig{
		Kind:    "baidu",
		BaseURL: server.URL,
		Auth: map[string]string{
			"access_token": "static-token",
		},
	})
	if err != nil {
		t.Fatalf("new baidu provider: %v", err)
	}

	resp, err := pAny.Call(context.Background(), Request{
		FilePath: filePath,
		Options: map[string]any{
			"language_type":  "CHN_ENG",
			"ignored_option": "skip",
		},
	})
	if err != nil {
		t.Fatalf("baidu call: %v", err)
	}
	if !strings.Contains(string(resp.Raw), `"words_result"`) {
		t.Fatalf("unexpected baidu response: %s", string(resp.Raw))
	}
}

func TestBaiduProviderCallReturnsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"error_code":17,"error_msg":"Open api daily request limit reached"}`))
	}))
	defer server.Close()

	pAny, err := NewBaiduProvider(config.ProviderConfig{
		Kind:    "baidu",
		BaseURL: server.URL,
		Auth: map[string]string{
			"access_token": "static-token",
		},
	})
	if err != nil {
		t.Fatalf("new baidu provider: %v", err)
	}

	_, err = pAny.Call(context.Background(), Request{FilePath: server.URL, IsRemote: true})
	if err == nil {
		t.Fatal("expected baidu api error")
	}
	apiErr, ok := err.(*BaiduError)
	if !ok {
		t.Fatalf("unexpected error type %T", err)
	}
	if apiErr.Code != 17 || apiErr.ErrorCategory() != "client" {
		t.Fatalf("unexpected baidu error: %+v", apiErr)
	}
}

func TestVolcEngineProviderCallBuildsSignedRequest(t *testing.T) {
	dir := t.TempDir()
	filePath := dir + "/sample.bin"
	if err := osWriteFile(filePath, []byte("hello")); err != nil {
		t.Fatalf("write file: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("Action"); got != defaultVolcAction {
			t.Fatalf("unexpected action: %s", got)
		}
		if got := r.URL.Query().Get("Version"); got != defaultVolcVersion {
			t.Fatalf("unexpected version: %s", got)
		}
		if !strings.HasPrefix(r.Header.Get("Authorization"), "HMAC-SHA256 Credential=ak/") {
			t.Fatalf("unexpected authorization header: %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("X-Date") == "" || r.Header.Get("X-Content-Sha256") == "" {
			t.Fatalf("expected signing headers, got %+v", r.Header)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if got := r.PostForm.Get("mode"); got != "fast" {
			t.Fatalf("unexpected mode option: %s", got)
		}
		if got := r.PostForm.Get("image_base64"); got == "" {
			t.Fatal("expected image_base64 payload")
		}
		_, _ = w.Write([]byte(`{"code":10000,"message":"OK"}`))
	}))
	defer server.Close()

	pAny, err := NewVolcEngineProvider(config.ProviderConfig{
		BaseURL: server.URL,
		Auth: map[string]string{
			"access_key_id":     "ak",
			"secret_access_key": "sk",
		},
	})
	if err != nil {
		t.Fatalf("new volc provider: %v", err)
	}

	resp, err := pAny.Call(context.Background(), Request{
		FilePath: filePath,
		Options:  map[string]any{"mode": "fast"},
	})
	if err != nil {
		t.Fatalf("volc call: %v", err)
	}
	if !strings.Contains(string(resp.Raw), `"code":10000`) {
		t.Fatalf("unexpected volc response: %s", string(resp.Raw))
	}
}

func TestVolcEngineProviderCallReturnsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":12345,"message":"bad request"}`))
	}))
	defer server.Close()

	pAny, err := NewVolcEngineProvider(config.ProviderConfig{
		BaseURL: server.URL,
		Auth: map[string]string{
			"access_key_id":     "ak",
			"secret_access_key": "sk",
		},
	})
	if err != nil {
		t.Fatalf("new volc provider: %v", err)
	}

	_, err = pAny.Call(context.Background(), Request{FilePath: server.URL, IsRemote: true})
	if err == nil {
		t.Fatal("expected volc api error")
	}
	apiErr, ok := err.(*VolcError)
	if !ok {
		t.Fatalf("unexpected error type %T", err)
	}
	if apiErr.Code != 12345 || apiErr.ErrorCategory() != "volcengine" {
		t.Fatalf("unexpected volc error: %+v", apiErr)
	}
}

func osWriteFile(path string, data []byte) error { return os.WriteFile(path, data, 0o644) }
