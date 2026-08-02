package openaicompat

import (
	"context"
	"net/http"
	"time"

	"github.com/wsnacj/agentx-go/providers/transport"
)

// Authorizer injects host-resolved credentials into request headers.
// The canonical client never reads environment variables or credential stores.
type Authorizer func(context.Context, http.Header) error

// MediaResolver converts a host-approved local media reference into a provider URL.
// A nil resolver keeps the supplied value unchanged and performs no filesystem access.
type MediaResolver func(string) (string, error)

// HTTPDoer is the minimal outbound HTTP client seam.
type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// Config contains explicit, provider-level construction settings.
type Config struct {
	Name          string
	BaseURL       string
	Timeout       time.Duration
	CompatProfile string
	Transport     transport.Config
	Authorize     Authorizer
	ResolveMedia  MediaResolver
	HTTPClient    HTTPDoer
}

// Capability controls optional request fields supported by one model.
type Capability struct {
	Vision                  bool
	LocalFiles              bool
	Thinking                bool
	Bots                    bool
	Streaming               bool
	MultimodalEmbed         bool
	ReasoningEffort         *bool
	ParallelToolCalls       *bool
	EmbeddingVideo          *bool
	EmbeddingDimensions     *bool
	EmbeddingInstructions   *bool
	SparseEmbedding         *bool
	EmbeddingEncodingBase64 *bool
}

// ModelConfig contains one chat/vision/bot model's host-selected defaults.
type ModelConfig struct {
	Name             string
	Model            string
	MaxCompletion    int
	Temperature      float32
	Capability       Capability
	PluginEndpoint   string
	ReasoningDefault string
	ThinkingDefault  string
	ResponseFormat   any
}

// EmbeddingConfig contains one embedding model's host-selected defaults.
type EmbeddingConfig struct {
	Name       string
	Model      string
	Dimensions int
	Path       string
	Encoding   string
	BatchSize  int
	MaxTokens  int
	Capability Capability
	Wrapper    string
}
