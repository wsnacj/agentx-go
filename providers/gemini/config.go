package gemini

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

// ResolvedMedia is a host-approved media reference. Exactly one of URI or
// Base64Data should normally be set.
type ResolvedMedia struct {
	URI        string
	MIMEType   string
	Base64Data string
}

// MediaResolver converts a local or host-specific reference into portable
// Gemini media. A nil resolver prevents local filesystem access.
type MediaResolver func(context.Context, string) (ResolvedMedia, error)

// Config contains explicit Gemini native API construction settings.
type Config struct {
	Name          string
	BaseURL       string
	UploadBaseURL string
	Timeout       time.Duration
	Transport     transport.Config
	Authorize     Authorizer
	ResolveMedia  MediaResolver
	HTTPClient    HTTPDoer
}

// Capability controls optional request fields supported by one model.
type Capability struct {
	Vision     bool
	LocalFiles bool
	Streaming  bool
}

// ModelConfig contains one chat/vision model's host-selected defaults.
type ModelConfig struct {
	Name          string
	Model         string
	MaxCompletion int
	Temperature   float32
	Capability    Capability
}

// EmbeddingConfig contains one embedding model's host-selected defaults.
type EmbeddingConfig struct {
	Name       string
	Model      string
	Dimensions int
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
