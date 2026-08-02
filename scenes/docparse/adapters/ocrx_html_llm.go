package adapters

import (
	"context"
	"fmt"
	"strings"

	"github.com/wsnacj/agentx-go/scenes/docparse/planner"
	"github.com/wsnacj/agentx-go/scenes/docparse/representation"
)

const (
	DefaultOCRXHTMLLLMMaxPromptChars      = representation.DefaultMaxPromptChars
	DefaultOCRXHTMLLLMMaxHTMLPreviewChars = 4000
)

// OCRXHTMLLLMParser is the host-owned execution seam for OCRX HTML/table
// representation plus LLM-based structured extraction.
type OCRXHTMLLLMParser func(context.Context, OCRXHTMLLLMRequest) (OCRXHTMLLLMResult, error)

// OCRXHTMLLLMRequest is the normalized request passed to the host parser.
type OCRXHTMLLLMRequest struct {
	Document       representation.Document
	DocumentPath   string
	ProfileID      string
	Route          planner.Route
	Params         map[string]any
	TextPages      []string
	PromptText     string
	MaxPromptChars int
}

// OCRXHTMLLLMResult is the source-neutral result returned by the host parser.
type OCRXHTMLLLMResult struct {
	Status         string
	Fields         []map[string]any
	Tables         []map[string]any
	Payload        map[string]any
	Warnings       []string
	Diagnostics    map[string]any
	ReviewRequired bool
	HTMLPages      []string
	CombinedHTML   string
	Model          string
	PromptID       string
}

// OCRXHTMLLLMAdapterOptions configures the opt-in OCRX HTML/LLM adapter.
type OCRXHTMLLLMAdapterOptions struct {
	Parser              OCRXHTMLLLMParser
	MaxPromptChars      int
	IncludeHTMLPreview  bool
	MaxHTMLPreviewChars int
}

// NewOCRXHTMLLLMExtractionAdapter creates an opt-in adapter for OCRX HTML/table
// plus LLM extraction.
//
// The adapter deliberately does not create OCRX clients or invoke LLMs itself.
// Hosts must provide Parser so provider credentials, prompt policy, caching,
// retry, and cost controls stay in the host layer.
func NewOCRXHTMLLLMExtractionAdapter(opts OCRXHTMLLLMAdapterOptions) RouteAdapter {
	return NewRouteAdapter(OCRXHTMLLLMAdapterID, planner.RouteOCRXHTMLLLM, "core/ocrx+llm", func(ctx context.Context, input Input) (Output, error) {
		return executeOCRXHTMLLLM(ctx, input, opts)
	})
}

func executeOCRXHTMLLLM(ctx context.Context, input Input, opts OCRXHTMLLLMAdapterOptions) (Output, error) {
	if opts.Parser == nil {
		return Output{}, fmt.Errorf("ocrx html llm adapter requires host parser because OCRX providers, LLM prompts, credentials, cache, retry, and cost policy are host-owned")
	}
	documentPath := strings.TrimSpace(input.Document.DocumentPath)
	if documentPath == "" {
		documentPath = readString(input.Params, "document_path")
	}
	if documentPath == "" {
		return Output{}, fmt.Errorf("document_path is required for ocrx html llm adapter")
	}
	maxPromptChars := opts.MaxPromptChars
	if maxPromptChars <= 0 {
		maxPromptChars = DefaultOCRXHTMLLLMMaxPromptChars
	}
	if value := readPositiveInt(input.Params, "max_prompt_chars"); value > 0 {
		maxPromptChars = value
	}
	req := OCRXHTMLLLMRequest{
		Document:       input.Document,
		DocumentPath:   documentPath,
		ProfileID:      strings.TrimSpace(input.Route.ProfileID),
		Route:          input.Route,
		Params:         cloneParams(input.Params),
		TextPages:      input.Document.TextPages(),
		PromptText:     input.Document.PromptText(maxPromptChars),
		MaxPromptChars: maxPromptChars,
	}
	result, err := opts.Parser(ctx, req)
	if err != nil {
		return Output{}, err
	}
	return ocrxHTMLLLMOutput(input, req, result, opts), nil
}

func ocrxHTMLLLMOutput(input Input, req OCRXHTMLLLMRequest, result OCRXHTMLLLMResult, opts OCRXHTMLLLMAdapterOptions) Output {
	status := strings.TrimSpace(result.Status)
	if status == "" {
		status = "success"
	}
	fields := cloneObjectMaps(result.Fields)
	tables := cloneObjectMaps(result.Tables)
	payload := cloneParams(result.Payload)
	payload["text_page_count"] = len(req.TextPages)
	if len(result.HTMLPages) > 0 {
		payload["html_page_count"] = len(result.HTMLPages)
	}
	if strings.TrimSpace(result.CombinedHTML) != "" {
		payload["combined_html_chars"] = runeLen(result.CombinedHTML)
		if opts.IncludeHTMLPreview {
			limit := opts.MaxHTMLPreviewChars
			if limit <= 0 {
				limit = DefaultOCRXHTMLLLMMaxHTMLPreviewChars
			}
			payload["html_preview"] = truncateRunes(result.CombinedHTML, limit)
		}
	}
	warnings := uniqueAppend(nil, result.Warnings...)
	reviewRequired := result.ReviewRequired || status != "success"
	evidenceComplete := ocrxHTMLLLMEvidenceComplete(fields, tables)
	reviewReason := ""
	if len(fields) == 0 && len(tables) == 0 {
		reviewRequired = true
		reviewReason = "ocrx_html_llm_no_structured_output"
		warnings = uniqueAppend(warnings, reviewReason)
	} else if !evidenceComplete {
		reviewRequired = true
		reviewReason = "ocrx_html_llm_evidence_incomplete"
		warnings = uniqueAppend(warnings, reviewReason)
	}
	diagnostics := cloneParams(result.Diagnostics)
	diagnostics["backend"] = "host_ocrx_html_llm_parser"
	diagnostics["parser_backend"] = "host_ocrx_html_llm_parser"
	diagnostics["document_path"] = strings.TrimSpace(req.DocumentPath)
	diagnostics["profile_id"] = strings.TrimSpace(input.Route.ProfileID)
	diagnostics["field_count"] = len(fields)
	diagnostics["table_count"] = len(tables)
	diagnostics["text_page_count"] = len(req.TextPages)
	diagnostics["max_prompt_chars"] = req.MaxPromptChars
	diagnostics["html_page_count"] = len(result.HTMLPages)
	diagnostics["combined_html_chars"] = runeLen(result.CombinedHTML)
	diagnostics["evidence_complete"] = evidenceComplete
	if model := strings.TrimSpace(result.Model); model != "" {
		diagnostics["model"] = model
	}
	if promptID := strings.TrimSpace(result.PromptID); promptID != "" {
		diagnostics["prompt_id"] = promptID
	}
	if reviewReason != "" {
		diagnostics["review_reason"] = reviewReason
	}
	return Output{
		AdapterID:      OCRXHTMLLLMAdapterID,
		RouteKind:      planner.RouteOCRXHTMLLLM,
		Status:         status,
		Payload:        payload,
		Fields:         fields,
		Tables:         tables,
		ReviewRequired: reviewRequired,
		Warnings:       warnings,
		Diagnostics:    diagnostics,
	}
}

func cloneObjectMaps(items []map[string]any) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, cloneParams(item))
	}
	return out
}

func ocrxHTMLLLMEvidenceComplete(fields []map[string]any, tables []map[string]any) bool {
	if len(fields) == 0 && len(tables) == 0 {
		return false
	}
	for _, field := range fields {
		if !mapHasValue(field, "normalized_value", "value", "raw_value") || !mapHasEvidence(field, "evidence", "page_refs", "page_ref", "bounding_boxes", "bbox", "table_cells") {
			return false
		}
	}
	for _, table := range tables {
		if !mapHasEvidence(table, "evidence", "page_refs", "page_ref", "cell_refs") {
			return false
		}
	}
	return true
}

func mapHasValue(item map[string]any, keys ...string) bool {
	for _, key := range keys {
		value, ok := item[key]
		if !ok || value == nil {
			continue
		}
		text, isString := value.(string)
		if !isString {
			return true
		}
		if strings.TrimSpace(text) != "" {
			return true
		}
	}
	return false
}

func mapHasEvidence(item map[string]any, keys ...string) bool {
	for _, key := range keys {
		value, ok := item[key]
		if !ok || value == nil {
			continue
		}
		switch typed := value.(type) {
		case string:
			if strings.TrimSpace(typed) != "" {
				return true
			}
		case []any:
			if len(typed) > 0 {
				return true
			}
		case []int:
			if len(typed) > 0 {
				return true
			}
		case []string:
			if len(typed) > 0 {
				return true
			}
		case []map[string]any:
			if len(typed) > 0 {
				return true
			}
		case int:
			if typed > 0 {
				return true
			}
		case int64:
			if typed > 0 {
				return true
			}
		case float64:
			if typed > 0 {
				return true
			}
		case map[string]any:
			if len(typed) > 0 {
				return true
			}
		case bool:
			if typed {
				return true
			}
		default:
			return true
		}
	}
	return false
}

func readPositiveInt(payload map[string]any, key string) int {
	value, ok := payload[key]
	if !ok || value == nil {
		return 0
	}
	switch typed := value.(type) {
	case int:
		if typed > 0 {
			return typed
		}
	case int64:
		if typed > 0 {
			return int(typed)
		}
	case float64:
		if typed > 0 {
			return int(typed)
		}
	case string:
		var out int
		if _, err := fmt.Sscanf(strings.TrimSpace(typed), "%d", &out); err == nil && out > 0 {
			return out
		}
	}
	return 0
}

func truncateRunes(value string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}

func runeLen(value string) int {
	return len([]rune(value))
}
