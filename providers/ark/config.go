package ark

import (
	"context"
	"net/http"
	"time"

	"github.com/wsnacj/agentx-go/providers/transport"
)

// Authorizer injects host-resolved credentials into request headers.
// The client never reads environment variables or credential stores.
type Authorizer func(context.Context, http.Header) error

// HTTPDoer is the minimal outbound HTTP client seam.
type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// Config contains explicit Ark API construction settings.
type Config struct {
	BaseURL       string
	Timeout       time.Duration
	StreamTimeout time.Duration
	Transport     transport.Config
	Authorize     Authorizer
	HTTPClient    HTTPDoer
}

func cloneHeaders(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}
