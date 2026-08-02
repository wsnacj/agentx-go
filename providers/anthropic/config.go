package anthropic

import (
	"context"
	"net/http"
	"time"

	"github.com/wsnacj/agentx-go/providers/transport"
)

// Authorizer injects host-resolved credentials into request headers.
// The canonical client never reads environment variables or credential stores.
type Authorizer func(context.Context, http.Header) error

// HTTPDoer is the minimal outbound HTTP client seam.
type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// Config contains explicit Anthropic Messages client settings.
type Config struct {
	Name       string
	BaseURL    string
	Version    string
	Timeout    time.Duration
	Transport  transport.Config
	Authorize  Authorizer
	HTTPClient HTTPDoer
}

// ModelConfig contains host-selected defaults for one model.
type ModelConfig struct {
	Name          string
	Model         string
	MaxCompletion int
	Temperature   float32
}
