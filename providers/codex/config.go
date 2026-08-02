package codex

import (
	"context"
	"net/http"
	"time"

	"github.com/wsnacj/agentx-go/providers/transport"
)

// Authorizer injects host-resolved credentials and account identity headers.
// The canonical client never reads token stores, environment variables or user files.
type Authorizer func(context.Context, http.Header) error

// HTTPDoer is the minimal outbound HTTP client seam.
type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// Config contains explicit Codex Responses protocol settings.
type Config struct {
	Name       string
	BaseURL    string
	Timeout    time.Duration
	UserAgent  string
	Originator string
	Transport  transport.Config
	Authorize  Authorizer
	HTTPClient HTTPDoer
}

// ModelConfig contains host-selected defaults for one model.
type ModelConfig struct {
	Name             string
	Model            string
	MaxCompletion    int
	Temperature      float32
	Thinking         bool
	ReasoningEffort  *bool
	ReasoningDefault string
}
