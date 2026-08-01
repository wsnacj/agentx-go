package hostserver

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"flag"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestConfigValidateExposurePolicy(t *testing.T) {
	if err := DefaultConfig("127.0.0.1:8792").Validate(); err != nil {
		t.Fatalf("loopback validation: %v", err)
	}
	exposed := DefaultConfig("0.0.0.0:8792")
	if err := exposed.Validate(); !errors.Is(err, ErrExposedTokenRequired) {
		t.Fatalf("exposed without token = %v, want token required", err)
	}
	exposed.AccessToken = "secret"
	if err := exposed.Validate(); !errors.Is(err, ErrTrustedProxyCIDRsRequired) {
		t.Fatalf("exposed without proxy cidrs = %v, want proxy required", err)
	}
	exposed.TrustedProxyCIDRs = "10.0.0.0/8"
	if err := exposed.Validate(); err != nil {
		t.Fatalf("trusted exposed config: %v", err)
	}
	exposed.TrustedProxyCIDRs = "not-a-cidr"
	if err := exposed.Validate(); !errors.Is(err, ErrInvalidTrustedProxyCIDR) {
		t.Fatalf("invalid trusted proxy = %v, want invalid cidr", err)
	}
}

func TestServeUntilSignalStableMethodContract(t *testing.T) {
	var serve func(Config, ServeOptions) error = Config.ServeUntilSignal
	if serve == nil {
		t.Fatal("ServeUntilSignal method contract is unavailable")
	}
}

func TestConfigBindFlagsDoesNotExposeEnvironmentToken(t *testing.T) {
	t.Setenv("AGENTX_BUSINESS_HOST_TOKEN", "environment-secret-token")
	config := DefaultConfig("127.0.0.1:8792")
	flags := flag.NewFlagSet("hostserver", flag.ContinueOnError)
	var output bytes.Buffer
	flags.SetOutput(&output)
	config.BindFlags(flags)
	if err := flags.Parse([]string{"-h"}); !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("parse help = %v, want flag.ErrHelp", err)
	}
	if strings.Contains(output.String(), "environment-secret-token") {
		t.Fatalf("flag help leaked access token: %s", output.String())
	}
}

func TestRequestIdentityAcceptsGeneratesAndRejects(t *testing.T) {
	var called atomic.Int32
	handler := RequestIdentityHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called.Add(1)
		_, _ = io.WriteString(w, RequestIDFromContext(r.Context()))
	}))

	accepted := httptest.NewRequest(http.MethodGet, "/", nil)
	accepted.Header.Set(RequestIDHeader, "caller-123")
	acceptedRecorder := httptest.NewRecorder()
	handler.ServeHTTP(acceptedRecorder, accepted)
	if got := acceptedRecorder.Header().Get(RequestIDHeader); got != "caller-123" || acceptedRecorder.Body.String() != "caller-123" {
		t.Fatalf("accepted request identity header=%q body=%q", got, acceptedRecorder.Body.String())
	}

	generated := httptest.NewRequest(http.MethodGet, "/", nil)
	generatedRecorder := httptest.NewRecorder()
	handler.ServeHTTP(generatedRecorder, generated)
	generatedID := generatedRecorder.Header().Get(RequestIDHeader)
	if generatedID == "" || generatedRecorder.Body.String() != generatedID || !validRequestID(generatedID) {
		t.Fatalf("generated request identity header=%q body=%q", generatedID, generatedRecorder.Body.String())
	}

	invalid := httptest.NewRequest(http.MethodGet, "/", nil)
	invalid.Header.Set(RequestIDHeader, "contains space")
	invalidRecorder := httptest.NewRecorder()
	handler.ServeHTTP(invalidRecorder, invalid)
	if invalidRecorder.Code != http.StatusBadRequest || !strings.Contains(invalidRecorder.Body.String(), "invalid_request_id") {
		t.Fatalf("invalid request identity response=%d %s", invalidRecorder.Code, invalidRecorder.Body.String())
	}
	if invalidRecorder.Header().Get(RequestIDHeader) == "" || called.Load() != 2 {
		t.Fatalf("invalid request reached handler or lacked replacement identity: calls=%d header=%q", called.Load(), invalidRecorder.Header().Get(RequestIDHeader))
	}
}

func TestBoundaryHandlerAuthHealthProxyAndBodyLimit(t *testing.T) {
	config := DefaultConfig("0.0.0.0:8792")
	config.AccessToken = "secret-token"
	config.TrustedProxyCIDRs = "192.0.2.0/24"
	config.MaxBodyBytes = 8
	server, err := config.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" || r.URL.Path == "/readyz" {
			w.WriteHeader(http.StatusOK)
			return
		}
		raw, _ := io.ReadAll(r.Body)
		_, _ = w.Write(raw)
	}))
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	for _, path := range []string{"/healthz", "/readyz"} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		request.RemoteAddr = "192.0.2.10:1234"
		recorder := httptest.NewRecorder()
		server.Handler.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK || recorder.Header().Get(RequestIDHeader) == "" {
			t.Fatalf("probe %s status=%d request-id=%q", path, recorder.Code, recorder.Header().Get(RequestIDHeader))
		}
	}

	unauthorized := httptest.NewRequest(http.MethodPost, "/v1/query", strings.NewReader("small"))
	unauthorized.RemoteAddr = "192.0.2.10:1234"
	unauthorizedRecorder := httptest.NewRecorder()
	server.Handler.ServeHTTP(unauthorizedRecorder, unauthorized)
	if unauthorizedRecorder.Code != http.StatusUnauthorized || strings.Contains(unauthorizedRecorder.Body.String(), "secret-token") || unauthorizedRecorder.Header().Get(RequestIDHeader) == "" {
		t.Fatalf("unauthorized response = %d %s", unauthorizedRecorder.Code, unauthorizedRecorder.Body.String())
	}

	untrusted := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	untrusted.RemoteAddr = "198.51.100.1:1234"
	untrustedRecorder := httptest.NewRecorder()
	server.Handler.ServeHTTP(untrustedRecorder, untrusted)
	if untrustedRecorder.Code != http.StatusForbidden {
		t.Fatalf("untrusted proxy status = %d, want 403", untrustedRecorder.Code)
	}

	for _, input := range []struct {
		name    string
		body    string
		chunked bool
	}{
		{name: "content-length", body: "123456789"},
		{name: "chunked", body: "123456789", chunked: true},
	} {
		t.Run(input.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/v1/query", bytes.NewBufferString(input.body))
			request.RemoteAddr = "192.0.2.10:1234"
			request.Header.Set("Authorization", "Bearer secret-token")
			if input.chunked {
				request.ContentLength = -1
				request.TransferEncoding = []string{"chunked"}
			}
			recorder := httptest.NewRecorder()
			server.Handler.ServeHTTP(recorder, request)
			if recorder.Code != http.StatusRequestEntityTooLarge || !strings.Contains(recorder.Body.String(), "request_body_too_large") {
				t.Fatalf("oversized response = %d %s", recorder.Code, recorder.Body.String())
			}
		})
	}

	allowed := httptest.NewRequest(http.MethodPost, "/v1/query", bytes.NewBufferString("12345678"))
	allowed.RemoteAddr = "192.0.2.10:1234"
	allowed.Header.Set("Authorization", "Bearer secret-token")
	allowedRecorder := httptest.NewRecorder()
	server.Handler.ServeHTTP(allowedRecorder, allowed)
	if allowedRecorder.Code != http.StatusOK || allowedRecorder.Body.String() != "12345678" {
		t.Fatalf("allowed response = %d %q", allowedRecorder.Code, allowedRecorder.Body.String())
	}
}

func TestConfigServeGracefulShutdown(t *testing.T) {
	config := DefaultConfig("127.0.0.1:0")
	config.ShutdownTimeout = time.Second
	ctx, cancel := context.WithCancel(context.Background())
	listening := make(chan string, 1)
	done := make(chan error, 1)
	go func() {
		done <- config.Serve(ctx, ServeOptions{
			Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}),
			OnListening: func(addr string) { listening <- addr },
		})
	}()
	var addr string
	select {
	case addr = <-listening:
	case <-time.After(time.Second):
		t.Fatal("server did not start")
	}
	response, err := http.Get("http://" + addr + "/healthz")
	if err != nil {
		t.Fatalf("health request: %v", err)
	}
	_ = response.Body.Close()
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("graceful shutdown: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not shut down")
	}
}

func TestConfigServeBoundsSlowRequestBody(t *testing.T) {
	config := DefaultConfig("127.0.0.1:0")
	config.ReadTimeout = 75 * time.Millisecond
	config.ShutdownTimeout = time.Second
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	listening := make(chan string, 1)
	done := make(chan error, 1)
	var handlerCalled atomic.Bool
	go func() {
		done <- config.Serve(ctx, ServeOptions{
			Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				handlerCalled.Store(true)
				w.WriteHeader(http.StatusNoContent)
			}),
			OnListening: func(addr string) { listening <- addr },
		})
	}()
	var addr string
	select {
	case addr = <-listening:
	case <-time.After(time.Second):
		t.Fatal("server did not start")
	}
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		t.Fatalf("dial server: %v", err)
	}
	defer conn.Close()
	if _, err := io.WriteString(conn, "POST /v1/query HTTP/1.1\r\nHost: agentx.test\r\nContent-Length: 5\r\n\r\n1"); err != nil {
		t.Fatalf("write slow request prefix: %v", err)
	}
	time.Sleep(2 * config.ReadTimeout)
	_ = conn.SetReadDeadline(time.Now().Add(time.Second))
	statusLine, readErr := bufio.NewReader(conn).ReadString('\n')
	if readErr == nil && !strings.Contains(statusLine, " 400 ") {
		t.Fatalf("slow body status = %q, want 400 or closed connection", statusLine)
	}
	if handlerCalled.Load() {
		t.Fatal("slow request body reached application handler")
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("shutdown after slow body: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not shut down after slow body")
	}
}
