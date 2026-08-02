package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	types "github.com/wsnacj/agentx-go/components/llm"
	toolcontract "github.com/wsnacj/agentx-go/components/tool"
	"github.com/wsnacj/agentx-go/document/pipeline"
	"github.com/wsnacj/agentx-go/document/pipeline/configs"
	pipelinetypes "github.com/wsnacj/agentx-go/document/pipeline/types"
	agentxtools "github.com/wsnacj/agentx-go/tools"
)

const (
	documentParseToolName                 = "document_parse"
	documentSpecRecommendToolName         = "document_spec_recommend"
	defaultDocumentParseTimeoutMs         = 300_000
	defaultDocumentParseMaxResponseBytes  = 120_000
	defaultDocumentSpecRecommendPageLimit = 3
)

type DocumentParseToolOptions struct {
	Root                  string
	DefaultModel          string
	DefaultArtifactPolicy pipeline.ArtifactPolicy
	DefaultExtractionMode pipeline.DocumentExtractionMode
	DefaultTimeoutMs      int
	MaxResponseBytes      int
	Host                  DocumentHost
	// EnabledTools is intentionally required for registration. document_parse
	// is a specialist surface and must not become visible by helper defaults.
	EnabledTools []string
}

type documentParseFieldSummary struct {
	Chapter         string      `json:"chapter,omitempty"`
	Key             string      `json:"key"`
	Value           interface{} `json:"value,omitempty"`
	RawValue        interface{} `json:"raw_value,omitempty"`
	NormalizedValue interface{} `json:"normalized_value,omitempty"`
	Source          string      `json:"source,omitempty"`
	Confidence      float64     `json:"confidence,omitempty"`
	Evidence        string      `json:"evidence,omitempty"`
	Unit            string      `json:"unit,omitempty"`
	Currency        string      `json:"currency,omitempty"`
	Period          string      `json:"period,omitempty"`
	PageRefs        []int       `json:"page_refs,omitempty"`
	ReviewRequired  bool        `json:"review_required,omitempty"`
	Warnings        []string    `json:"warnings,omitempty"`
	CandidateCount  int         `json:"candidate_count,omitempty"`
	SelectionReason string      `json:"selection_reason,omitempty"`
}

type documentParseChapterSummary struct {
	Key        string `json:"key"`
	TextSize   int    `json:"text_size,omitempty"`
	FieldCount int    `json:"field_count,omitempty"`
}

type documentParseResponse struct {
	Tool           string                             `json:"tool"`
	Status         string                             `json:"status"`
	DocumentPath   string                             `json:"document_path,omitempty"`
	SpecPath       string                             `json:"spec_path,omitempty"`
	OutputDir      string                             `json:"output_dir,omitempty"`
	ArtifactPolicy string                             `json:"artifact_policy,omitempty"`
	ExtractionMode string                             `json:"extraction_mode,omitempty"`
	FilesTouched   []string                           `json:"files_touched,omitempty"`
	PageCount      int                                `json:"page_count,omitempty"`
	TextQuality    string                             `json:"text_quality,omitempty"`
	TextSource     string                             `json:"text_source,omitempty"`
	ChapterCount   int                                `json:"chapter_count,omitempty"`
	FieldCount     int                                `json:"field_count,omitempty"`
	ReviewRequired bool                               `json:"review_required,omitempty"`
	Chapters       []documentParseChapterSummary      `json:"chapters,omitempty"`
	Fields         []documentParseFieldSummary        `json:"fields,omitempty"`
	Diagnostics    *pipelinetypes.DocumentDiagnostics `json:"diagnostics,omitempty"`
	Warnings       []string                           `json:"warnings,omitempty"`
	ErrorClass     string                             `json:"error_class,omitempty"`
	Error          string                             `json:"error,omitempty"`
	Result         *pipelinetypes.DocumentResult      `json:"result,omitempty"`
}

type documentSpecRecommendResponse struct {
	Tool             string                       `json:"tool"`
	Status           string                       `json:"status"`
	DocumentPath     string                       `json:"document_path,omitempty"`
	SpecPaths        []string                     `json:"spec_paths,omitempty"`
	PageCount        int                          `json:"page_count,omitempty"`
	TextSize         int                          `json:"text_size,omitempty"`
	ExtractionMode   string                       `json:"extraction_mode,omitempty"`
	RecommendationBy string                       `json:"recommendation_by,omitempty"`
	Recommendations  []configs.SpecRecommendation `json:"recommendations,omitempty"`
	Warnings         []string                     `json:"warnings,omitempty"`
	ErrorClass       string                       `json:"error_class,omitempty"`
	Error            string                       `json:"error,omitempty"`
}

func RegisterDocumentParseTools(reg toolcontract.Registrar, opts DocumentParseToolOptions) error {
	if reg == nil {
		return fmt.Errorf("document tool registrar is required")
	}
	if err := opts.Host.validate(); err != nil {
		return err
	}
	enabled := documentParseEnabledToolSet(opts.EnabledTools)
	if enabled[documentParseToolName] {
		reg.Register(documentParseDefinition(), func(ctx context.Context, call types.FunctionCall) (string, error) {
			params, err := decodeArgs(call.Arguments)
			if err != nil {
				return "", err
			}
			payload, err := runDocumentParseTool(ctx, params, opts)
			if err != nil {
				return "", fmt.Errorf("%s: %w", documentParseToolName, err)
			}
			return marshalDocumentParseResponse(payload, documentParseMaxResponseBytes(params, opts))
		})
	}
	if enabled[documentSpecRecommendToolName] {
		reg.Register(documentSpecRecommendDefinition(), func(ctx context.Context, call types.FunctionCall) (string, error) {
			params, err := decodeArgs(call.Arguments)
			if err != nil {
				return "", err
			}
			payload, err := runDocumentSpecRecommendTool(ctx, params, opts)
			if err != nil {
				return "", fmt.Errorf("%s: %w", documentSpecRecommendToolName, err)
			}
			return marshalDocumentSpecRecommendResponse(payload, documentParseMaxResponseBytes(params, opts))
		})
	}
	return nil
}

func documentParseEnabledToolSet(enabledTools []string) map[string]bool {
	enabled := map[string]bool{}
	for _, raw := range enabledTools {
		name := agentxtools.NormalizeToolName(raw)
		switch name {
		case documentParseToolName, documentSpecRecommendToolName:
			enabled[name] = true
		}
	}
	return enabled
}

func documentParseDefinition() types.Tool {
	return types.Tool{
		Type: "function",
		Function: types.Function{
			Name:        documentParseToolName,
			Description: "Parse a local workspace document with a docparse spec and return structured fields, provenance, diagnostics, and artifact paths. This specialist tool does not fetch remote URLs; save remote documents locally first.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"document_path": map[string]any{"type": "string", "description": "Workspace-local document path."},
					"doc_path":      map[string]any{"type": "string", "description": "Alias for document_path kept for callers that use doc_path."},
					"path":          map[string]any{"type": "string", "description": "Alias for document_path kept for generic file-path callers."},
					"file_path":     map[string]any{"type": "string", "description": "Alias for document_path kept for file upload or artifact callers."},
					"spec_path":     map[string]any{"type": "string", "description": "Workspace-local docparse spec directory or main.yaml path."},
					"model":         map[string]any{"type": "string", "description": "Optional model name used by specs that require LLM extraction."},
					"output_dir":    map[string]any{"type": "string", "description": "Optional workspace-local artifact output directory."},
					"page_limit":    map[string]any{"type": "integer", "minimum": 1, "description": "Maximum number of pages to parse or inspect."},
					"max_chunk_chars": map[string]any{
						"type":        "integer",
						"minimum":     200,
						"description": "Optional per-chapter text budget override.",
					},
					"pdf_parse_mode": map[string]any{
						"type":        "string",
						"description": "default|simple|fast|normal|ocr|force_ocr",
					},
					"extraction_mode": map[string]any{
						"type":        "string",
						"description": "default|legacy|table_first|text_layer_first|ocr_first|auto. Controls document text representation source order; explicit caller value overrides host defaults.",
					},
					"artifact_policy": map[string]any{
						"type":        "string",
						"description": "summary|full|none",
					},
					"timeout_ms": map[string]any{
						"type":        "integer",
						"minimum":     1,
						"description": "Total parse timeout in milliseconds.",
					},
					"include_full_result": map[string]any{"type": "boolean", "description": "Include the full raw docparse result in the response when supported; otherwise the tool returns a compact summary."},
					"max_response_bytes":  map[string]any{"type": "integer", "minimum": 1024, "description": "Maximum response size before compacting or truncating large parse payloads."},
				},
				"required": []string{"document_path", "spec_path"},
			},
			OutputSchema: documentParseOutputSchema(),
		},
	}
}

// DocumentParseDefinition returns the model-facing document_parse contract.
func DocumentParseDefinition() types.Tool { return documentParseDefinition() }

func documentParseOutputSchema() map[string]any {
	return closedOutputSchema(map[string]any{
		"tool":            stringSchema("Tool name that produced this response."),
		"status":          stringSchema("Execution status, usually success or failed."),
		"document_path":   stringSchema("Workspace display path for the parsed document."),
		"spec_path":       stringSchema("Workspace display path for the docparse spec used by the run."),
		"output_dir":      stringSchema("Workspace artifact directory used for parse outputs."),
		"artifact_policy": stringSchema("Artifact policy applied to the parse run."),
		"extraction_mode": stringSchema("Resolved document extraction mode used by the parse run."),
		"files_touched":   stringArraySchema("Workspace artifacts actually written by document_parse, such as result, diagnostics, and manifest files."),
		"page_count":      intSchema("Number of document pages represented by diagnostics.", 0),
		"text_quality":    stringSchema("Text quality signal reported by docparse diagnostics."),
		"text_source":     stringSchema("Text extraction backend reported by docparse diagnostics, such as text, pdf_simple, pdf_normal, pdf_ocrx, or image_ocrx."),
		"chapter_count":   intSchema("Number of chapters represented in the compact response.", 0),
		"field_count":     intSchema("Number of extracted fields represented in the compact response.", 0),
		"review_required": boolSchema("True when any extracted field or parse signal requires review."),
		"chapters":        looseObjectArraySchema("Compact chapter summaries returned by document_parse."),
		"fields":          looseObjectArraySchema("Compact field summaries returned by document_parse."),
		"diagnostics":     looseObjectSchema("Docparse diagnostics returned by the parse pipeline."),
		"warnings":        stringArraySchema("Warnings emitted by docparse diagnostics or response compaction."),
		"error_class":     stringSchema("Structured failure class when parsing fails."),
		"error":           stringSchema("Human-readable failure message when parsing fails."),
		"result":          looseObjectSchema("Full docparse result when include_full_result is requested."),
	}, []string{"tool", "status"})
}

func documentSpecRecommendDefinition() types.Tool {
	return types.Tool{
		Type: "function",
		Function: types.Function{
			Name:        documentSpecRecommendToolName,
			Description: "Recommend the best matching docparse spec from caller-provided workspace-local candidate specs for a local workspace document. This read-only specialist tool does not parse the document or auto-select a spec.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"document_path": map[string]any{"type": "string", "description": "Workspace-local document path to inspect."},
					"doc_path":      map[string]any{"type": "string", "description": "Alias for document_path kept for callers that use doc_path."},
					"path":          map[string]any{"type": "string", "description": "Alias for document_path kept for generic file-path callers."},
					"file_path":     map[string]any{"type": "string", "description": "Alias for document_path kept for file upload or artifact callers."},
					"spec_paths": map[string]any{
						"type":        "array",
						"items":       map[string]any{"type": "string"},
						"description": "Workspace-local docparse spec directories or main.yaml paths to rank.",
					},
					"spec_path":            map[string]any{"type": "string", "description": "Single workspace-local candidate spec path."},
					"candidate_spec_paths": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Alias for spec_paths kept for callers that name the candidate list explicitly."},
					"page_limit":           map[string]any{"type": "integer", "minimum": 1, "description": "Number of leading pages to inspect; defaults to 3."},
					"pdf_parse_mode": map[string]any{
						"type":        "string",
						"description": "default|simple|fast|normal|ocr|force_ocr",
					},
					"extraction_mode": map[string]any{
						"type":        "string",
						"description": "default|legacy|table_first|text_layer_first|ocr_first|auto. Controls document text representation source order for candidate scoring.",
					},
					"max_response_bytes": map[string]any{"type": "integer", "minimum": 1024, "description": "Maximum response size before compacting large recommendation payloads."},
				},
				"required": []string{"document_path"},
			},
		},
	}
}

// DocumentSpecRecommendDefinition returns the read-only recommendation contract.
func DocumentSpecRecommendDefinition() types.Tool { return documentSpecRecommendDefinition() }

func runDocumentParseTool(ctx context.Context, params map[string]any, opts DocumentParseToolOptions) (documentParseResponse, error) {
	documentPath := firstString(params, "document_path", "doc_path", "path", "file_path")
	specPath := firstString(params, "spec_path")
	if looksLikeRemoteDocumentPath(documentPath) {
		return documentParseResponse{}, fmt.Errorf("remote document URLs are not supported; fetch or save the document to a workspace artifact first")
	}
	if looksLikeRemoteDocumentPath(specPath) {
		return documentParseResponse{}, fmt.Errorf("remote spec URLs are not supported")
	}
	resolvedDocument, err := opts.Host.Paths.ResolvePath(ctx, PathRequest{
		Root: opts.Root, Value: documentPath, Field: "document_path", MustExist: true, FileOnly: true,
	})
	if err != nil {
		return documentParseResponse{}, err
	}
	resolvedSpec, err := opts.Host.Paths.ResolvePath(ctx, PathRequest{
		Root: opts.Root, Value: specPath, Field: "spec_path", MustExist: true,
	})
	if err != nil {
		return documentParseResponse{}, err
	}
	outputDir, err := resolveDocumentParseOutputDir(ctx, opts.Host.Paths, opts.Root, resolvedDocument.Path, firstString(params, "output_dir"))
	if err != nil {
		return documentParseResponse{}, err
	}
	artifactPolicy, err := documentParseArtifactPolicy(firstString(params, "artifact_policy"), opts.DefaultArtifactPolicy)
	if err != nil {
		return documentParseResponse{}, err
	}
	pdfParseMode, err := documentParsePDFParseMode(firstString(params, "pdf_parse_mode"))
	if err != nil {
		return documentParseResponse{}, err
	}
	extractionMode, err := documentParseExtractionMode(firstString(params, "extraction_mode"), opts.DefaultExtractionMode)
	if err != nil {
		return documentParseResponse{}, err
	}
	timeoutMs := firstInt(params, "timeout_ms")
	if timeoutMs < 0 {
		return documentParseResponse{}, fmt.Errorf("timeout_ms must be non-negative")
	}
	if timeoutMs == 0 {
		timeoutMs = opts.DefaultTimeoutMs
	}
	if timeoutMs <= 0 {
		timeoutMs = defaultDocumentParseTimeoutMs
	}
	pageLimit := firstInt(params, "page_limit")
	if pageLimit < 0 {
		return documentParseResponse{}, fmt.Errorf("page_limit must be non-negative")
	}
	maxChunkChars := firstInt(params, "max_chunk_chars")
	if maxChunkChars < 0 {
		return documentParseResponse{}, fmt.Errorf("max_chunk_chars must be non-negative")
	}
	parseCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()
	modelName := strings.TrimSpace(firstString(params, "model", "model_name", "modelName"))
	if modelName == "" {
		modelName = strings.TrimSpace(opts.DefaultModel)
	}

	result, parseErr := opts.Host.Runtime.Run(parseCtx, pipeline.ParseRequest{
		DocPath:        resolvedDocument.Path,
		SpecPath:       resolvedSpec.Path,
		ModelName:      modelName,
		OutputDir:      outputDir.Path,
		MaxChunkChars:  maxChunkChars,
		PageLimit:      pageLimit,
		PDFParseMode:   pdfParseMode,
		ExtractionMode: extractionMode,
		ArtifactPolicy: artifactPolicy,
	})
	if parseErr != nil {
		errorClass := opts.Host.errorClass(parseErr)
		return documentParseResponse{
			Tool:           documentParseToolName,
			Status:         "failed",
			DocumentPath:   resolvedDocument.Display,
			SpecPath:       resolvedSpec.Display,
			OutputDir:      outputDir.Display,
			ArtifactPolicy: string(artifactPolicy),
			ExtractionMode: documentParseExtractionModeName(extractionMode),
			ErrorClass:     errorClass,
			Error:          opts.Host.displayError(parseErr, documentParseToolName, errorClass),
		}, nil
	}
	payload := buildDocumentParseResponse(result, resolvedDocument.Display, resolvedSpec.Display, outputDir.Display, artifactPolicy)
	if opts.Host.Artifacts != nil {
		payload.FilesTouched, _ = opts.Host.Artifacts.ListArtifacts(ctx, outputDir.Path)
		sort.Strings(payload.FilesTouched)
	}
	if readBool(params, "include_full_result") {
		payload.Result = result
	}
	return payload, nil
}

func runDocumentSpecRecommendTool(ctx context.Context, params map[string]any, opts DocumentParseToolOptions) (documentSpecRecommendResponse, error) {
	documentPath := firstString(params, "document_path", "doc_path", "path", "file_path")
	if looksLikeRemoteDocumentPath(documentPath) {
		return documentSpecRecommendResponse{}, fmt.Errorf("remote document URLs are not supported; fetch or save the document to a workspace artifact first")
	}
	resolvedDocument, err := opts.Host.Paths.ResolvePath(ctx, PathRequest{
		Root: opts.Root, Value: documentPath, Field: "document_path", MustExist: true, FileOnly: true,
	})
	if err != nil {
		return documentSpecRecommendResponse{}, err
	}
	specPaths := documentSpecRecommendPaths(params)
	if len(specPaths) == 0 {
		return documentSpecRecommendResponse{}, fmt.Errorf("spec_paths is required")
	}
	resolvedSpecPaths := make([]string, 0, len(specPaths))
	displaySpecPaths := make([]string, 0, len(specPaths))
	for _, specPath := range specPaths {
		if looksLikeRemoteDocumentPath(specPath) {
			return documentSpecRecommendResponse{}, fmt.Errorf("remote spec URLs are not supported")
		}
		resolvedSpec, err := opts.Host.Paths.ResolvePath(ctx, PathRequest{
			Root: opts.Root, Value: specPath, Field: "spec_paths", MustExist: true,
		})
		if err != nil {
			return documentSpecRecommendResponse{}, err
		}
		resolvedSpecPaths = append(resolvedSpecPaths, resolvedSpec.Path)
		displaySpecPaths = append(displaySpecPaths, resolvedSpec.Display)
	}
	pdfParseMode, err := documentParsePDFParseMode(firstString(params, "pdf_parse_mode"))
	if err != nil {
		return documentSpecRecommendResponse{}, err
	}
	extractionMode, err := documentParseExtractionMode(firstString(params, "extraction_mode"), opts.DefaultExtractionMode)
	if err != nil {
		return documentSpecRecommendResponse{}, err
	}
	pageLimit := firstInt(params, "page_limit")
	if pageLimit < 0 {
		return documentSpecRecommendResponse{}, fmt.Errorf("page_limit must be non-negative")
	}
	if pageLimit == 0 {
		pageLimit = defaultDocumentSpecRecommendPageLimit
	}
	if opts.Host.Text == nil {
		return documentSpecRecommendResponse{}, fmt.Errorf("document text loader is required")
	}
	parseMode := pipeline.PDFParseSimple
	if extractionMode != nil {
		parseMode = pipeline.PDFParseNormal
	}
	if pdfParseMode != nil {
		parseMode = *pdfParseMode
	}
	resolvedExtractionMode := pipeline.DocumentExtractionModeLegacy
	if extractionMode != nil {
		resolvedExtractionMode = *extractionMode
	}
	pages, err := opts.Host.Text.LoadDocumentText(ctx, DocumentTextRequest{
		Path: resolvedDocument.Path, PageLimit: pageLimit, PDFParseMode: parseMode, ExtractionMode: resolvedExtractionMode,
	})
	if err != nil {
		errorClass := opts.Host.errorClass(err)
		return documentSpecRecommendResponse{
			Tool:           documentSpecRecommendToolName,
			Status:         "failed",
			DocumentPath:   resolvedDocument.Display,
			SpecPaths:      displaySpecPaths,
			ExtractionMode: documentParseExtractionModeName(extractionMode),
			ErrorClass:     errorClass,
			Error:          opts.Host.displayError(err, documentSpecRecommendToolName, errorClass),
		}, nil
	}
	specs := make([]*configs.DocSpec, 0, len(resolvedSpecPaths))
	displayByConfigDir := make(map[string]string, len(resolvedSpecPaths))
	for index, specPath := range resolvedSpecPaths {
		spec, err := opts.Host.loadSpec(specPath)
		if err != nil {
			return documentSpecRecommendResponse{}, fmt.Errorf("load spec %s: %w", displaySpecPaths[index], err)
		}
		displayByConfigDir[filepath.Clean(spec.ConfigDir)] = displaySpecPaths[index]
		specs = append(specs, spec)
	}
	text := strings.Join(pages, "\n\n")
	recommendations := opts.Host.recommend(text, specs)
	for i := range recommendations {
		if display := displayByConfigDir[filepath.Clean(recommendations[i].ConfigDir)]; display != "" {
			recommendations[i].ConfigDir = display
		}
	}
	payload := documentSpecRecommendResponse{
		Tool:             documentSpecRecommendToolName,
		Status:           "success",
		DocumentPath:     resolvedDocument.Display,
		SpecPaths:        displaySpecPaths,
		PageCount:        len(pages),
		TextSize:         len(text),
		ExtractionMode:   documentParseExtractionModeName(extractionMode),
		RecommendationBy: "source_neutral_text_spec_scoring",
		Recommendations:  recommendations,
	}
	if len(recommendations) == 0 {
		payload.Warnings = append(payload.Warnings, "no_matching_specs")
	}
	return payload, nil
}

func buildDocumentParseResponse(result *pipelinetypes.DocumentResult, documentPath string, specPath string, outputDir string, artifactPolicy pipeline.ArtifactPolicy) documentParseResponse {
	payload := documentParseResponse{
		Tool:           documentParseToolName,
		Status:         "success",
		DocumentPath:   documentPath,
		SpecPath:       specPath,
		OutputDir:      outputDir,
		ArtifactPolicy: string(artifactPolicy),
	}
	if result == nil {
		payload.Status = "failed"
		payload.Error = "docparse returned nil result"
		return payload
	}
	payload.Diagnostics = result.Diagnostics
	if result.Fingerprint != nil {
		payload.ExtractionMode = result.Fingerprint.ExtractionMode
	}
	if result.Diagnostics != nil {
		payload.PageCount = result.Diagnostics.PageCount
		payload.TextQuality = result.Diagnostics.TextQuality
		payload.TextSource = result.Diagnostics.TextSource
		payload.Warnings = append(payload.Warnings, result.Diagnostics.Warnings...)
	}
	keys := documentParseChapterKeys(result)
	payload.ChapterCount = len(keys)
	for _, key := range keys {
		chapter := result.Chapters[key]
		if chapter == nil {
			continue
		}
		payload.Chapters = append(payload.Chapters, documentParseChapterSummary{
			Key:        key,
			TextSize:   chapter.TextSize,
			FieldCount: len(chapter.Fields),
		})
		payload.Fields = append(payload.Fields, documentParseFieldSummaries(key, chapter.Fields)...)
	}
	payload.FieldCount = len(payload.Fields)
	for _, field := range payload.Fields {
		if field.ReviewRequired {
			payload.ReviewRequired = true
			break
		}
	}
	return payload
}

func documentParseChapterKeys(result *pipelinetypes.DocumentResult) []string {
	if result == nil || len(result.Chapters) == 0 {
		return nil
	}
	if len(result.ChapterOrder) > 0 {
		out := make([]string, 0, len(result.ChapterOrder))
		seen := map[string]bool{}
		for _, key := range result.ChapterOrder {
			if key == "" || seen[key] {
				continue
			}
			if _, ok := result.Chapters[key]; !ok {
				continue
			}
			seen[key] = true
			out = append(out, key)
		}
		if len(out) > 0 {
			return out
		}
	}
	out := make([]string, 0, len(result.Chapters))
	for key := range result.Chapters {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}

func documentParseFieldSummaries(chapterKey string, fields map[string]pipelinetypes.FieldResult) []documentParseFieldSummary {
	if len(fields) == 0 {
		return nil
	}
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]documentParseFieldSummary, 0, len(keys))
	for _, key := range keys {
		field := fields[key]
		chapter := strings.TrimSpace(field.Chapter)
		if chapter == "" {
			chapter = chapterKey
		}
		out = append(out, documentParseFieldSummary{
			Chapter:         chapter,
			Key:             documentParseFirstNonEmptyString(field.Key, key),
			Value:           field.Value,
			RawValue:        field.RawValue,
			NormalizedValue: field.NormalizedValue,
			Source:          field.Source,
			Confidence:      field.Confidence,
			Evidence:        field.Evidence,
			Unit:            field.Unit,
			Currency:        field.Currency,
			Period:          field.Period,
			PageRefs:        append([]int(nil), field.PageRefs...),
			ReviewRequired:  field.ReviewRequired,
			Warnings:        append([]string(nil), field.Warnings...),
			CandidateCount:  len(field.Candidates),
			SelectionReason: field.SelectionReason,
		})
	}
	return out
}

func resolveDocumentParseOutputDir(ctx context.Context, resolver PathResolver, root string, documentPath string, outputDir string) (ResolvedPath, error) {
	if strings.TrimSpace(outputDir) != "" {
		return resolver.ResolvePath(ctx, PathRequest{Root: root, Value: outputDir, Field: "output_dir"})
	}
	stem := documentParseSafeStem(filepath.Base(documentPath))
	if stem == "" {
		stem = "document"
	}
	rel := filepath.Join(".agentx", "docparse", fmt.Sprintf("%s-%s", stem, time.Now().UTC().Format("20060102T150405.000000000Z")))
	return resolver.ResolvePath(ctx, PathRequest{Root: root, Value: rel, Field: "output_dir"})
}

func documentParseArtifactPolicy(raw string, fallback pipeline.ArtifactPolicy) (pipeline.ArtifactPolicy, error) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" {
		normalized = strings.ToLower(strings.TrimSpace(string(fallback)))
	}
	switch normalized {
	case "", string(pipeline.ArtifactPolicySummary):
		return pipeline.ArtifactPolicySummary, nil
	case string(pipeline.ArtifactPolicyFull):
		return pipeline.ArtifactPolicyFull, nil
	case string(pipeline.ArtifactPolicyNone):
		return pipeline.ArtifactPolicyNone, nil
	default:
		return "", fmt.Errorf("unknown artifact_policy %q", raw)
	}
}

func documentParsePDFParseMode(raw string) (*pipeline.PDFParseMode, error) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" || normalized == "default" {
		return nil, nil
	}
	var mode pipeline.PDFParseMode
	switch normalized {
	case "simple", "fast":
		mode = pipeline.PDFParseSimple
	case "normal":
		mode = pipeline.PDFParseNormal
	case "ocr", "force_ocr", "force-ocr":
		mode = pipeline.PDFParseForceOCR
	default:
		return nil, fmt.Errorf("unknown pdf_parse_mode %q", raw)
	}
	return &mode, nil
}

func documentParseExtractionMode(raw string, fallback pipeline.DocumentExtractionMode) (*pipeline.DocumentExtractionMode, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		trimmed = string(fallback)
	}
	if strings.TrimSpace(trimmed) == "" || strings.EqualFold(strings.TrimSpace(trimmed), "default") {
		return nil, nil
	}
	mode, err := pipeline.NormalizeDocumentExtractionMode(trimmed)
	if err != nil {
		return nil, err
	}
	if mode == pipeline.DocumentExtractionModeLegacy {
		return nil, nil
	}
	return &mode, nil
}

func documentParseExtractionModeName(mode *pipeline.DocumentExtractionMode) string {
	if mode == nil {
		return ""
	}
	return string(*mode)
}

func documentParseMaxResponseBytes(params map[string]any, opts DocumentParseToolOptions) int {
	maxBytes := firstInt(params, "max_response_bytes")
	if maxBytes <= 0 {
		maxBytes = opts.MaxResponseBytes
	}
	if maxBytes <= 0 {
		maxBytes = defaultDocumentParseMaxResponseBytes
	}
	return maxBytes
}

func marshalDocumentParseResponse(payload documentParseResponse, maxBytes int) (string, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	if maxBytes <= 0 || len(raw) <= maxBytes {
		return string(raw), nil
	}
	if payload.Result != nil {
		payload.Result = nil
		payload.Warnings = append(payload.Warnings, "full_result_omitted_response_too_large")
	}
	raw, err = json.Marshal(payload)
	if err != nil {
		return "", err
	}
	if len(raw) <= maxBytes {
		return string(raw), nil
	}
	if payload.Diagnostics != nil {
		payload.Diagnostics = nil
		payload.Warnings = append(payload.Warnings, "diagnostics_omitted_response_too_large")
	}
	if len(payload.Fields) > 50 {
		payload.Fields = payload.Fields[:50]
		payload.Warnings = append(payload.Warnings, "fields_truncated_response_too_large")
	}
	if len(payload.Chapters) > 50 {
		payload.Chapters = payload.Chapters[:50]
		payload.Warnings = append(payload.Warnings, "chapters_truncated_response_too_large")
	}
	raw, err = json.Marshal(payload)
	if err != nil {
		return "", err
	}
	if maxBytes > 0 && len(raw) > maxBytes {
		return "", fmt.Errorf("response too large after compaction (%d > %d bytes)", len(raw), maxBytes)
	}
	return string(raw), nil
}

func marshalDocumentSpecRecommendResponse(payload documentSpecRecommendResponse, maxBytes int) (string, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	if maxBytes <= 0 || len(raw) <= maxBytes {
		return string(raw), nil
	}
	if len(payload.Recommendations) > 20 {
		payload.Recommendations = payload.Recommendations[:20]
		payload.Warnings = append(payload.Warnings, "recommendations_truncated_response_too_large")
	}
	raw, err = json.Marshal(payload)
	if err != nil {
		return "", err
	}
	if maxBytes > 0 && len(raw) > maxBytes {
		return "", fmt.Errorf("response too large after compaction (%d > %d bytes)", len(raw), maxBytes)
	}
	return string(raw), nil
}

func documentParseSafeStem(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	raw = strings.TrimSuffix(raw, filepath.Ext(raw))
	var b strings.Builder
	for _, r := range raw {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r + ('a' - 'A'))
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	stem := strings.Trim(b.String(), "-_")
	if len(stem) > 48 {
		stem = stem[:48]
	}
	return stem
}

func looksLikeRemoteDocumentPath(path string) bool {
	normalized := strings.ToLower(strings.TrimSpace(path))
	return strings.HasPrefix(normalized, "http://") ||
		strings.HasPrefix(normalized, "https://") ||
		strings.Contains(normalized, "://")
}

func documentSpecRecommendPaths(params map[string]any) []string {
	out := make([]string, 0)
	seen := map[string]bool{}
	appendPath := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			return
		}
		seen[value] = true
		out = append(out, value)
	}
	for _, key := range []string{"spec_paths", "candidate_spec_paths", "specs"} {
		for _, item := range readStringList(params, key) {
			appendPath(item)
		}
	}
	appendPath(firstString(params, "spec_path"))
	return out
}

func documentParseFirstNonEmptyString(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
