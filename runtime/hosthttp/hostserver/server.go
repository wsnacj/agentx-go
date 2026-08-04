// Package hostserver provides a bounded local-first HTTP transport for AgentX
// host-deployed Scene adapters.
package hostserver

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var (
	ErrExposedTokenRequired      = errors.New("business host exposure requires an access token")
	ErrTrustedProxyCIDRsRequired = errors.New("business host exposure requires trusted proxy cidrs")
	ErrInvalidTrustedProxyCIDR   = errors.New("business host trusted proxy cidr is invalid")
)

const (
	DefaultMaxBodyBytes   int64 = 2 << 20
	DefaultMaxHeaderBytes       = 64 << 10
	RequestIDHeader             = "X-Request-ID"
	MaxRequestIDLength          = 64
)

type requestIDContextKey struct{}

// Config defines the shared transport and exposure boundary for scene hosts.
type Config struct {
	Addr              string
	AccessToken       string
	TrustedProxyCIDRs string
	MaxBodyBytes      int64
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	MaxHeaderBytes    int
}

// ServeOptions supplies the Scene handler and an optional post-listen notification.
type ServeOptions struct {
	Handler     http.Handler
	OnListening func(string)
}

// DefaultConfig returns the bounded local-first host transport profile.
//
// It is a pure default and never reads credentials or deployment policy from
// the process environment. Hosts that intentionally use the conventional
// environment variables must call ConfigFromEnv explicitly.
func DefaultConfig(addr string) Config {
	return Config{
		Addr:              strings.TrimSpace(addr),
		MaxBodyBytes:      DefaultMaxBodyBytes,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      120 * time.Second,
		IdleTimeout:       60 * time.Second,
		ShutdownTimeout:   10 * time.Second,
		MaxHeaderBytes:    DefaultMaxHeaderBytes,
	}
}

// ConfigFromEnv returns DefaultConfig enriched with the conventional Host
// access token and trusted proxy CIDRs from the process environment.
//
// This function is an explicit deployment adapter. It does not mutate the
// environment, log credentials, or read any other provider or Scene setting.
func ConfigFromEnv(addr string) Config {
	config := DefaultConfig(addr)
	config.AccessToken = strings.TrimSpace(os.Getenv("AGENTX_BUSINESS_HOST_TOKEN"))
	config.TrustedProxyCIDRs = strings.TrimSpace(os.Getenv("AGENTX_BUSINESS_HOST_TRUSTED_PROXY_CIDRS"))
	return config
}

// BindFlags registers the common host transport flags.
func (c *Config) BindFlags(flags *flag.FlagSet) {
	if c == nil || flags == nil {
		return
	}
	defaults := normalizeConfig(*c)
	*c = defaults
	flags.StringVar(&c.Addr, "addr", defaults.Addr, "HTTP listen address")
	flags.Var(secretFlagValue{target: &c.AccessToken}, "access-token", "Bearer token required outside unauthenticated health checks")
	flags.StringVar(&c.TrustedProxyCIDRs, "trusted-proxy-cidrs", defaults.TrustedProxyCIDRs, "comma-separated immediate proxy CIDRs required for non-loopback exposure")
	flags.Int64Var(&c.MaxBodyBytes, "max-body-bytes", defaults.MaxBodyBytes, "maximum HTTP request body bytes")
	flags.DurationVar(&c.ShutdownTimeout, "shutdown-timeout", defaults.ShutdownTimeout, "graceful HTTP shutdown timeout")
}

type secretFlagValue struct {
	target *string
}

func (v secretFlagValue) String() string { return "" }

func (v secretFlagValue) Set(raw string) error {
	if v.target != nil {
		*v.target = strings.TrimSpace(raw)
	}
	return nil
}

// Validate rejects malformed or incompletely protected exposure configuration.
func (c Config) Validate() error {
	c = normalizeConfig(c)
	_, err := parseTrustedProxyCIDRs(c.TrustedProxyCIDRs)
	if err != nil {
		return err
	}
	if isLoopbackListenAddress(c.Addr) {
		return nil
	}
	if strings.TrimSpace(c.AccessToken) == "" {
		return ErrExposedTokenRequired
	}
	if strings.TrimSpace(c.TrustedProxyCIDRs) == "" {
		return ErrTrustedProxyCIDRsRequired
	}
	return nil
}

// NewServer constructs an HTTP server with request identity and boundary middleware.
func (c Config) NewServer(handler http.Handler) (*http.Server, error) {
	c = normalizeConfig(c)
	if err := c.Validate(); err != nil {
		return nil, err
	}
	if handler == nil {
		handler = http.NotFoundHandler()
	}
	networks, err := parseTrustedProxyCIDRs(c.TrustedProxyCIDRs)
	if err != nil {
		return nil, err
	}
	exposed := !isLoopbackListenAddress(c.Addr)
	bounded := boundaryHandler(handler, boundaryOptions{
		token:          c.AccessToken,
		maxBodyBytes:   c.MaxBodyBytes,
		exposed:        exposed,
		trustedProxies: networks,
	})
	return &http.Server{
		Addr:              c.Addr,
		Handler:           RequestIdentityHandler(bounded),
		ReadTimeout:       c.ReadTimeout,
		ReadHeaderTimeout: c.ReadHeaderTimeout,
		WriteTimeout:      c.WriteTimeout,
		IdleTimeout:       c.IdleTimeout,
		MaxHeaderBytes:    c.MaxHeaderBytes,
	}, nil
}

// Serve listens until ctx ends, then performs a bounded graceful shutdown.
func (c Config) Serve(ctx context.Context, options ServeOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	server, err := c.NewServer(options.Handler)
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return fmt.Errorf("business host listen: %w", err)
	}
	defer listener.Close()
	serveErr := make(chan error, 1)
	go func() {
		serveErr <- server.Serve(listener)
	}()
	if options.OnListening != nil {
		options.OnListening(listener.Addr().String())
	}
	select {
	case err := <-serveErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownTimeout := normalizeConfig(c).ShutdownTimeout
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			_ = server.Close()
			return fmt.Errorf("business host shutdown: %w", err)
		}
		err := <-serveErr
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}
}

// ServeUntilSignal serves until SIGINT or SIGTERM requests shutdown.
func (c Config) ServeUntilSignal(options ServeOptions) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return c.Serve(ctx, options)
}

// RequestIDFromContext returns the validated request identity installed by
// RequestIdentityHandler. It returns an empty string outside that boundary.
func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	value, _ := ctx.Value(requestIDContextKey{}).(string)
	return value
}

// RequestIdentityHandler accepts a bounded caller identity or generates one,
// writes it to the response header, and makes it available through context.
// Invalid caller identities are rejected before the application handler runs.
func RequestIdentityHandler(next http.Handler) http.Handler {
	if next == nil {
		next = http.NotFoundHandler()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		incoming := strings.TrimSpace(r.Header.Get(RequestIDHeader))
		if incoming != "" && !validRequestID(incoming) {
			requestID := newRequestID()
			w.Header().Set(RequestIDHeader, requestID)
			writeBoundaryError(w, http.StatusBadRequest, "invalid_request_id", requestID)
			return
		}
		requestID := incoming
		if requestID == "" {
			requestID = newRequestID()
		}
		w.Header().Set(RequestIDHeader, requestID)
		ctx := context.WithValue(r.Context(), requestIDContextKey{}, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type boundaryOptions struct {
	token          string
	maxBodyBytes   int64
	exposed        bool
	trustedProxies []*net.IPNet
}

func boundaryHandler(next http.Handler, options boundaryOptions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := RequestIDFromContext(r.Context())
		if options.exposed && !remoteAllowed(r.RemoteAddr, options.trustedProxies) {
			writeBoundaryError(w, http.StatusForbidden, "untrusted_proxy", requestID)
			return
		}
		if !isUnauthenticatedProbe(r.URL.Path) && strings.TrimSpace(options.token) != "" && !validBearerToken(r.Header.Get("Authorization"), options.token) {
			w.Header().Set("WWW-Authenticate", `Bearer realm="agentx-business-host"`)
			writeBoundaryError(w, http.StatusUnauthorized, "unauthorized", requestID)
			return
		}
		if r.Body != nil && options.maxBodyBytes > 0 {
			if r.ContentLength > options.maxBodyBytes {
				writeBoundaryError(w, http.StatusRequestEntityTooLarge, "request_body_too_large", requestID)
				return
			}
			raw, err := io.ReadAll(io.LimitReader(r.Body, options.maxBodyBytes+1))
			_ = r.Body.Close()
			if err != nil {
				writeBoundaryError(w, http.StatusBadRequest, "request_body_read_failed", requestID)
				return
			}
			if int64(len(raw)) > options.maxBodyBytes {
				writeBoundaryError(w, http.StatusRequestEntityTooLarge, "request_body_too_large", requestID)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(raw))
			r.ContentLength = int64(len(raw))
		}
		next.ServeHTTP(w, r)
	})
}

func normalizeConfig(c Config) Config {
	c.Addr = strings.TrimSpace(c.Addr)
	if c.Addr == "" {
		c.Addr = "127.0.0.1:0"
	}
	c.AccessToken = strings.TrimSpace(c.AccessToken)
	c.TrustedProxyCIDRs = strings.TrimSpace(c.TrustedProxyCIDRs)
	if c.MaxBodyBytes <= 0 {
		c.MaxBodyBytes = DefaultMaxBodyBytes
	}
	if c.ReadTimeout <= 0 {
		c.ReadTimeout = 15 * time.Second
	}
	if c.ReadHeaderTimeout <= 0 {
		c.ReadHeaderTimeout = 5 * time.Second
	}
	if c.WriteTimeout <= 0 {
		c.WriteTimeout = 120 * time.Second
	}
	if c.IdleTimeout <= 0 {
		c.IdleTimeout = 60 * time.Second
	}
	if c.ShutdownTimeout <= 0 {
		c.ShutdownTimeout = 10 * time.Second
	}
	if c.MaxHeaderBytes <= 0 {
		c.MaxHeaderBytes = DefaultMaxHeaderBytes
	}
	return c
}

func isLoopbackListenAddress(addr string) bool {
	host, _, err := net.SplitHostPort(strings.TrimSpace(addr))
	if err != nil {
		return false
	}
	host = strings.Trim(strings.TrimSpace(host), "[]")
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func parseTrustedProxyCIDRs(raw string) ([]*net.IPNet, error) {
	parts := strings.Split(raw, ",")
	out := make([]*net.IPNet, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if ip := net.ParseIP(part); ip != nil {
			bits := 128
			if ip.To4() != nil {
				ip = ip.To4()
				bits = 32
			}
			out = append(out, &net.IPNet{IP: ip, Mask: net.CIDRMask(bits, bits)})
			continue
		}
		_, network, err := net.ParseCIDR(part)
		if err != nil {
			return nil, fmt.Errorf("%w", ErrInvalidTrustedProxyCIDR)
		}
		out = append(out, network)
	}
	return out, nil
}

func remoteAllowed(remoteAddr string, networks []*net.IPNet) bool {
	host, _, err := net.SplitHostPort(strings.TrimSpace(remoteAddr))
	if err != nil {
		return false
	}
	ip := net.ParseIP(strings.Trim(host, "[]"))
	if ip == nil {
		return false
	}
	for _, network := range networks {
		if network != nil && network.Contains(ip) {
			return true
		}
	}
	return false
}

func validBearerToken(header string, expected string) bool {
	parts := strings.Fields(strings.TrimSpace(header))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return false
	}
	provided := []byte(parts[1])
	want := []byte(strings.TrimSpace(expected))
	return len(provided) == len(want) && subtle.ConstantTimeCompare(provided, want) == 1
}

func isUnauthenticatedProbe(path string) bool {
	return path == "/healthz" || path == "/readyz"
}

func validRequestID(value string) bool {
	if value == "" || len(value) > MaxRequestIDLength {
		return false
	}
	for _, char := range []byte(value) {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '-' || char == '_' || char == '.' || char == ':' {
			continue
		}
		return false
	}
	return true
}

func newRequestID() string {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err == nil {
		return "req_" + hex.EncodeToString(raw)
	}
	return fmt.Sprintf("req_%x", time.Now().UTC().UnixNano())
}

func writeBoundaryError(w http.ResponseWriter, status int, code string, requestID string) {
	w.Header().Set("Content-Type", "application/json")
	if requestID != "" {
		w.Header().Set(RequestIDHeader, requestID)
	}
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":     "blocked",
		"ready":      false,
		"error":      code,
		"request_id": requestID,
	})
}
