package tools

import (
	"context"
	"fmt"
	"strings"

	llm "github.com/wsnacj/agentx-go/components/llm"
	"github.com/wsnacj/agentx-go/document/pipeline"
	"github.com/wsnacj/agentx-go/document/pipeline/configs"
	"github.com/wsnacj/agentx-go/document/pipeline/types"
	mediaartifact "github.com/wsnacj/agentx-go/runtime/mediaartifact"
)

// PathRequest describes one Host-authorized workspace path resolution.
// Canonical tools never infer an allowlist from the process working directory.
type PathRequest struct {
	Root      string
	Value     string
	Field     string
	MustExist bool
	FileOnly  bool
}

// ResolvedPath separates a backend path from the display-safe path returned to
// the model.
type ResolvedPath struct {
	Path    string
	Display string
}

// PathResolver is the authorization seam for document and artifact paths.
type PathResolver interface {
	ResolvePath(context.Context, PathRequest) (ResolvedPath, error)
}

// PathResolverFunc adapts a function to PathResolver.
type PathResolverFunc func(context.Context, PathRequest) (ResolvedPath, error)

func (f PathResolverFunc) ResolvePath(ctx context.Context, request PathRequest) (ResolvedPath, error) {
	return f(ctx, request)
}

// DocumentTextRequest asks the Host for a bounded page representation.
type DocumentTextRequest struct {
	Path           string
	PageLimit      int
	PDFParseMode   pipeline.PDFParseMode
	ExtractionMode pipeline.DocumentExtractionMode
}

// DocumentTextLoader owns concrete file decoding, OCR routing and filesystem
// access used by document_spec_recommend.
type DocumentTextLoader interface {
	LoadDocumentText(context.Context, DocumentTextRequest) ([]string, error)
}

// DocumentTextLoaderFunc adapts a function to DocumentTextLoader.
type DocumentTextLoaderFunc func(context.Context, DocumentTextRequest) ([]string, error)

func (f DocumentTextLoaderFunc) LoadDocumentText(ctx context.Context, request DocumentTextRequest) ([]string, error) {
	return f(ctx, request)
}

// ArtifactLister returns display-safe files written by an already authorized
// output directory.
type ArtifactLister interface {
	ListArtifacts(context.Context, string) ([]string, error)
}

// ArtifactListerFunc adapts a function to ArtifactLister.
type ArtifactListerFunc func(context.Context, string) ([]string, error)

func (f ArtifactListerFunc) ListArtifacts(ctx context.Context, outputDir string) ([]string, error) {
	return f(ctx, outputDir)
}

// ErrorProjector maps Host errors to stable, display-safe tool output.
type ErrorProjector interface {
	Classify(error) string
	Display(error, string, string) string
}

// ErrorProjectorFuncs is the function adapter for ErrorProjector.
type ErrorProjectorFuncs struct {
	ClassifyFunc func(error) string
	DisplayFunc  func(error, string, string) string
}

func (p ErrorProjectorFuncs) Classify(err error) string {
	if p.ClassifyFunc == nil {
		return ""
	}
	return strings.TrimSpace(p.ClassifyFunc(err))
}

func (p ErrorProjectorFuncs) Display(err error, toolName, class string) string {
	if p.DisplayFunc != nil {
		return p.DisplayFunc(err, toolName, class)
	}
	if err == nil {
		return ""
	}
	return err.Error()
}

// DocumentHost is the explicit construction seam for document_parse and
// document_spec_recommend. Runtime owns the portable parsing implementation;
// the remaining ports own authorization and concrete file access.
type DocumentHost struct {
	Runtime       DocumentParser
	Paths         PathResolver
	Text          DocumentTextLoader
	Artifacts     ArtifactLister
	Errors        ErrorProjector
	LoadSpec      func(string) (*configs.DocSpec, error)
	RecommendSpec func(string, []*configs.DocSpec) []configs.SpecRecommendation
}

// DocumentParser is the narrow execution seam consumed by document_parse.
// *pipeline.Runtime satisfies it directly; compatibility Hosts can adapt an
// existing entry point without exposing their construction graph.
type DocumentParser interface {
	Run(context.Context, pipeline.ParseRequest) (*types.DocumentResult, error)
}

// DocumentParserFunc adapts a function to DocumentParser.
type DocumentParserFunc func(context.Context, pipeline.ParseRequest) (*types.DocumentResult, error)

func (f DocumentParserFunc) Run(ctx context.Context, request pipeline.ParseRequest) (*types.DocumentResult, error) {
	return f(ctx, request)
}

func (h DocumentHost) validate() error {
	if h.Runtime == nil {
		return fmt.Errorf("document runtime is required")
	}
	if h.Paths == nil {
		return fmt.Errorf("document path resolver is required")
	}
	return nil
}

func (h DocumentHost) loadSpec(path string) (*configs.DocSpec, error) {
	if h.LoadSpec != nil {
		return h.LoadSpec(path)
	}
	return configs.LoadSpec(path)
}

func (h DocumentHost) recommend(text string, specs []*configs.DocSpec) []configs.SpecRecommendation {
	if h.RecommendSpec != nil {
		return h.RecommendSpec(text, specs)
	}
	return configs.RecommendSpecsForText(text, specs)
}

func (h DocumentHost) errorClass(err error) string {
	if h.Errors == nil {
		return ""
	}
	return h.Errors.Classify(err)
}

func (h DocumentHost) displayError(err error, toolName, class string) string {
	if h.Errors == nil {
		if err == nil {
			return ""
		}
		return err.Error()
	}
	return h.Errors.Display(err, toolName, class)
}

// PDFModelCandidate is one explicitly supplied model route. Model discovery,
// provider credentials and product routing remain Host-owned.
type PDFModelCandidate struct {
	Name           string `json:"name"`
	Client         string `json:"client,omitempty"`
	Model          string `json:"model,omitempty"`
	NativePDF      bool   `json:"native_pdf,omitempty"`
	SupportsVision bool   `json:"-"`
	ConfigKey      string `json:"-"`
	Vendor         string `json:"-"`
}

type pdfVisionModelCandidate = PDFModelCandidate

// PDFInputRequest asks the Host to authorize and materialize one local,
// file-URL or remote reference. Remote fetch policy belongs entirely to Host.
type PDFInputRequest struct {
	Root       string
	Reference  string
	TimeoutMS  int
	MaxBytes   int
	Parameters map[string]any
}

// ResolvedPDFInput is the bounded local materialization consumed by portable
// PDF coordination.
type ResolvedPDFInput struct {
	Path          string
	Display       string
	CacheIdentity string
	Remote        bool
	Cleanup       func()
}

// PDFInputResolver is the only local/remote input access used by PDF tools.
type PDFInputResolver interface {
	ResolvePDFInput(context.Context, PDFInputRequest) (ResolvedPDFInput, error)
}

// PDFInputResolverFunc adapts a function to PDFInputResolver.
type PDFInputResolverFunc func(context.Context, PDFInputRequest) (ResolvedPDFInput, error)

func (f PDFInputResolverFunc) ResolvePDFInput(ctx context.Context, request PDFInputRequest) (ResolvedPDFInput, error) {
	return f(ctx, request)
}

// PDFRenderedPage is an explicitly rendered page owned by a Host adapter.
type PDFRenderedPage struct {
	Page     int
	Path     string
	Data     []byte
	MIMEType string
}

type pdfRenderedPage = PDFRenderedPage

// PDFNativeRequest contains no credential or provider client. The Host chooses
// and executes the concrete native-PDF provider.
type PDFNativeRequest struct {
	Candidate PDFModelCandidate
	Prompt    string
	Paths     []string
}

// PDFOCRRequest delegates optional OCR enrichment to the Host.
type PDFOCRRequest struct {
	Profile string
	Path    string
}

// PDFHost contains every side-effecting/model capability needed by the
// portable PDF implementation. Nil optional capabilities fail closed or follow
// the existing deterministic fallback path.
type PDFHost struct {
	Inputs          PDFInputResolver
	LayoutName      string
	Layout          func(context.Context, string, []int) (PDFTextResult, error)
	Chat            func(context.Context, llm.ChatInput) (*llm.ChatResponse, error)
	Vision          func(context.Context, llm.VisionInput) (*llm.VisualResponse, error)
	Native          func(context.Context, PDFNativeRequest) (string, error)
	OCR             func(context.Context, PDFOCRRequest) ([]PDFPageText, error)
	Render          func(context.Context, string, []int, int) ([]PDFRenderedPage, func() error, error)
	PublishRendered func(context.Context, string, string, []PDFRenderedPage) ([]mediaartifact.Descriptor, error)
}

func (h PDFHost) validate() error {
	if h.Inputs == nil {
		return fmt.Errorf("pdf input resolver is required")
	}
	return nil
}

type pdfHostContextKey struct{}

func contextWithPDFHost(ctx context.Context, host PDFHost) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, pdfHostContextKey{}, host)
}

func pdfHostFromContext(ctx context.Context) PDFHost {
	if ctx == nil {
		return PDFHost{}
	}
	host, _ := ctx.Value(pdfHostContextKey{}).(PDFHost)
	return host
}

// Compile-time references keep the tool response types attached to the
// canonical pipeline graph and prevent an accidental duplicate DTO family.
var _ *types.DocumentResult
