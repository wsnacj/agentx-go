package httprequest_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	toolcontract "github.com/wsnacj/agentx-go/components/tool"
	agentxtoolerrors "github.com/wsnacj/agentx-go/runtime/toolerrors"
	"github.com/wsnacj/agentx-go/tools/httprequest"
)

type doerFunc func(*http.Request) (*http.Response, error)

func (fn doerFunc) Do(request *http.Request) (*http.Response, error) { return fn(request) }

func TestRunUsesExplicitPreparedClientAndProducesStablePayload(t *testing.T) {
	parsed, _ := url.Parse("https://example.test/base?keep=1")
	result, err := httprequest.Run(context.Background(), httprequest.Request{
		Method: "post", URL: parsed.String(), Query: map[string]string{"city": "beijing"},
		Headers: map[string]string{"X-Trace": "abc"}, Body: "payload", MaxChars: 64,
	}, httprequest.Options{Prepare: func(_ context.Context, input httprequest.PrepareInput) (httprequest.PreparedRequest, error) {
		if input.RawURL != parsed.String() || !input.FollowRedirects || !input.CredentialSensitive {
			t.Fatalf("prepare input = %#v", input)
		}
		return httprequest.PreparedRequest{URL: parsed, Doer: doerFunc(func(request *http.Request) (*http.Response, error) {
			if request.Method != http.MethodPost || request.URL.Query().Get("city") != "beijing" || request.Header.Get("X-Trace") != "abc" {
				t.Fatalf("request = %#v", request)
			}
			return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"application/json"}},
				Body: io.NopCloser(strings.NewReader(`{"ok":true}`)), Request: request}, nil
		})}, nil
	}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(result), &payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if payload["method"] != "POST" || payload["status"] != float64(200) || payload["body"] != `{"ok":true}` {
		t.Fatalf("payload = %#v", payload)
	}
	external := payload["external_content"].(map[string]any)
	if external["untrusted"] != true || external["source"] != "http_request" {
		t.Fatalf("external_content = %#v", external)
	}
	diagnostics := payload["provider_diagnostics"].(map[string]any)
	if diagnostics["effective_provider"] != "direct_http" {
		t.Fatalf("provider_diagnostics = %#v", diagnostics)
	}
}

func TestHandlerReturnsTypedMissingURL(t *testing.T) {
	handler := httprequest.NewHandler(httprequest.Options{Prepare: func(context.Context, httprequest.PrepareInput) (httprequest.PreparedRequest, error) {
		return httprequest.PreparedRequest{}, errors.New("must not run")
	}})
	_, err := handler(context.Background(), toolcontract.Call{Name: httprequest.Name, Arguments: `{}`})
	var argumentError *agentxtoolerrors.ToolArgumentError
	if !errors.As(err, &argumentError) || argumentError.Code != agentxtoolerrors.ToolArgumentErrorCodeMissingRequiredArgument {
		t.Fatalf("error = %#v", err)
	}
}

func TestRunTruncatesWithoutNetworkOrCredentialDiscovery(t *testing.T) {
	parsed, _ := url.Parse("https://example.test/")
	result, err := httprequest.Run(context.Background(), httprequest.Request{URL: parsed.String(), MaxChars: 4}, httprequest.Options{
		MaxChars: 8,
		Prepare: func(context.Context, httprequest.PrepareInput) (httprequest.PreparedRequest, error) {
			return httprequest.PreparedRequest{URL: parsed, Doer: doerFunc(func(request *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader("abcdefgh")), Request: request}, nil
			})}, nil
		},
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(result, `"body":"abcd"`) || !strings.Contains(result, `"truncated":true`) {
		t.Fatalf("result = %s", result)
	}
}

func TestDefinitionAndRegistration(t *testing.T) {
	definition := httprequest.Definition()
	if definition.Function.Name != httprequest.Name {
		t.Fatalf("definition = %#v", definition)
	}
}
