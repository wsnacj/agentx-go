package pipeline

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/wsnacj/agentx-go/document/pdf"
	"github.com/wsnacj/agentx-go/document/pipeline/section"
)

// RetryOptions communicates the legacy-compatible retry intent to a host model
// adapter. The portable pipeline never discovers providers or credentials and
// does not own network retry policy.
type RetryOptions struct {
	MaxRetries        int
	AttemptTimeout    time.Duration
	TotalTimeout      time.Duration
	BackoffBase       time.Duration
	BackoffMultiplier float64
	BackoffJitter     float64
}

// ModelRequest is one text completion required by document coordination.
type ModelRequest struct {
	ModelName string
	Prompt    string
	Chunks    []string
	Retry     RetryOptions
}

// Model is the only model capability required by Runtime. Provider selection,
// credentials, authorization, networking and retry execution remain host-owned.
type Model interface {
	Complete(context.Context, ModelRequest) (string, error)
}

// ModelFunc adapts a function to Model.
type ModelFunc func(context.Context, ModelRequest) (string, error)

func (f ModelFunc) Complete(ctx context.Context, req ModelRequest) (string, error) {
	return f(ctx, req)
}

// ExtractRequest describes a host-owned document extraction operation.
type ExtractRequest struct {
	Path           string
	PageLimit      int
	PDFParseMode   PDFParseMode
	ExtractionMode DocumentExtractionMode
}

// ExtractedDocument is the substrate-neutral result consumed by Runtime.
type ExtractedDocument struct {
	Pages       []string
	TextSource  string
	PDFResponse *pdfparser.TableResponse
}

// DocumentLoader supplies file decoding, OCR selection and concrete PDF access.
type DocumentLoader interface {
	Extract(context.Context, ExtractRequest) (*ExtractedDocument, error)
}

// DocumentLoaderFunc adapts a function to DocumentLoader.
type DocumentLoaderFunc func(context.Context, ExtractRequest) (*ExtractedDocument, error)

func (f DocumentLoaderFunc) Extract(ctx context.Context, req ExtractRequest) (*ExtractedDocument, error) {
	return f(ctx, req)
}

// SectionRequest is the complete input to a host-provided section splitter.
type SectionRequest struct {
	ConfigPath string
	Pages      []string
}

// Sectioner owns the concrete section-rule backend while returning portable
// section trees to Runtime.
type Sectioner interface {
	Split(context.Context, SectionRequest) ([]*section.Node, error)
}

// SectionerFunc adapts a function to Sectioner.
type SectionerFunc func(context.Context, SectionRequest) ([]*section.Node, error)

func (f SectionerFunc) Split(ctx context.Context, req SectionRequest) ([]*section.Node, error) {
	return f(ctx, req)
}

// StageEvent is a display-safe observation emitted at a major pipeline boundary.
type StageEvent struct {
	Stage  string
	Status string
	Detail string
}

// Observer receives optional runtime progress without controlling behavior.
type Observer interface {
	Observe(context.Context, StageEvent)
}

// ObserverFunc adapts a function to Observer.
type ObserverFunc func(context.Context, StageEvent)

func (f ObserverFunc) Observe(ctx context.Context, event StageEvent) { f(ctx, event) }

// Dependencies are the explicit construction seam for the document Runtime.
type Dependencies struct {
	Loader    DocumentLoader
	Sectioner Sectioner
	Model     Model
	Observer  Observer
}

// Runtime coordinates the portable document pipeline over host-provided I/O.
// It is safe for concurrent Run calls when the injected dependencies are safe.
type Runtime struct {
	loader    DocumentLoader
	sectioner Sectioner
	model     Model
	observer  Observer
}

// New constructs a Runtime without discovering host configuration.
func New(deps Dependencies) (*Runtime, error) {
	if deps.Loader == nil {
		return nil, fmt.Errorf("document loader is required")
	}
	if deps.Sectioner == nil {
		return nil, fmt.Errorf("sectioner is required")
	}
	return &Runtime{loader: deps.Loader, sectioner: deps.Sectioner, model: deps.Model, observer: deps.Observer}, nil
}

func (r *Runtime) complete(ctx context.Context, modelName, prompt string, chunks []string, retry RetryOptions) (string, error) {
	if r == nil || r.model == nil {
		return "", fmt.Errorf("model adapter is required")
	}
	return r.model.Complete(ctx, ModelRequest{
		ModelName: strings.TrimSpace(modelName),
		Prompt:    prompt,
		Chunks:    append([]string{}, chunks...),
		Retry:     retry,
	})
}

func (r *Runtime) observe(ctx context.Context, stage, status, detail string) {
	if r != nil && r.observer != nil {
		r.observer.Observe(ctx, StageEvent{Stage: stage, Status: status, Detail: detail})
	}
}
