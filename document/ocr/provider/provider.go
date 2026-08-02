package provider

import (
	"context"
	"fmt"

	"github.com/wsnacj/agentx-go/document/ocr/config"
)

// Request captures the payload sent to an OCR provider.
type Request struct {
	FilePath      string
	IsRemote      bool
	NeedCharacter bool
	Options       map[string]any
}

// Response is the generic provider response container.
type Response struct {
	Raw       []byte
	MediaType string
	Stats     map[string]any
}

// Provider defines the contract for upstream OCR services.
type Provider interface {
	Call(ctx context.Context, req Request) (Response, error)
}

// Factory constructs providers from configuration.
type Factory func(config.ProviderConfig) (Provider, error)

// Registry allows registering providers by kind.
type Registry map[string]Factory

// Lookup resolves provider factory by kind.
func (r Registry) Lookup(kind string) (Factory, error) {
	factory, ok := r[kind]
	if !ok {
		return nil, fmt.Errorf("provider %s not registered", kind)
	}
	return factory, nil
}
