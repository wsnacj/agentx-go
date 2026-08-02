package tools

import (
	"context"
	"fmt"
	"strings"
)

const (
	defaultPDFRemoteTimeoutMs = 30_000
	hardPDFRemoteTimeoutMs    = 120_000
	defaultPDFRemoteMaxBytes  = 10 * 1024 * 1024
	hardPDFRemoteMaxBytes     = 50 * 1024 * 1024
	hardPDFToolInputCount     = 8
)

type pdfToolResolvedInput = ResolvedPDFInput

func resolvePDFToolInput(ctx context.Context, root string, params map[string]any, defaultTimeoutMs int, defaultMaxBytes int) (pdfToolResolvedInput, error) {
	raw := strings.TrimSpace(firstString(params, "pdf", "path", "file_path", "url", "source_url"))
	if raw == "" {
		return pdfToolResolvedInput{}, newMissingRequiredToolArgumentError("pdf", []string{"pdf", "path", "file_path", "url", "source_url"}, "pdf: pdf/path/file_path/url/source_url is required")
	}
	if strings.HasPrefix(strings.ToLower(raw), "data:") {
		return pdfToolResolvedInput{}, fmt.Errorf("inline pdf data urls are not supported: write the pdf to a workspace file and pass its path")
	}
	host := pdfHostFromContext(ctx)
	if host.Inputs == nil {
		return pdfToolResolvedInput{}, fmt.Errorf("pdf input resolver is required")
	}
	resolved, err := host.Inputs.ResolvePDFInput(ctx, PDFInputRequest{
		Root: root, Reference: raw,
		TimeoutMS:  pdfToolRequestTimeoutMs(params, defaultTimeoutMs),
		MaxBytes:   pdfToolRequestMaxResponseBytes(params, defaultMaxBytes),
		Parameters: cloneMap(params),
	})
	if err != nil {
		return pdfToolResolvedInput{}, err
	}
	if strings.TrimSpace(resolved.Path) == "" {
		return pdfToolResolvedInput{}, fmt.Errorf("pdf input resolver returned an empty path")
	}
	if resolved.Cleanup == nil {
		resolved.Cleanup = func() {}
	}
	return resolved, nil
}

func resolvePDFToolInputs(ctx context.Context, root string, params map[string]any, defaultTimeoutMs int, defaultMaxBytes int, maxInputs int) ([]pdfToolResolvedInput, error) {
	rawInputs, err := readPDFToolInputs(params, maxInputs)
	if err != nil {
		return nil, err
	}
	if len(rawInputs) == 0 {
		input, err := resolvePDFToolInput(ctx, root, params, defaultTimeoutMs, defaultMaxBytes)
		if err != nil {
			return nil, err
		}
		return []pdfToolResolvedInput{input}, nil
	}
	out := make([]pdfToolResolvedInput, 0, len(rawInputs))
	for _, raw := range rawInputs {
		inputParams := cloneMap(params)
		delete(inputParams, "pdfs")
		inputParams["pdf"] = raw
		input, err := resolvePDFToolInput(ctx, root, inputParams, defaultTimeoutMs, defaultMaxBytes)
		if err != nil {
			cleanupPDFToolInputs(out)
			return nil, err
		}
		out = append(out, input)
	}
	return out, nil
}

func cleanupPDFToolInputs(inputs []pdfToolResolvedInput) {
	for _, input := range inputs {
		if input.Cleanup != nil {
			input.Cleanup()
		}
	}
}

func resolvedPDFToolInputPaths(inputs []pdfToolResolvedInput) []string {
	out := make([]string, 0, len(inputs))
	for _, input := range inputs {
		if path := strings.TrimSpace(input.Path); path != "" {
			out = append(out, path)
		}
	}
	return out
}

func readPDFToolInputs(params map[string]any, maxInputs int) ([]string, error) {
	if maxInputs <= 0 {
		maxInputs = hardPDFToolInputCount
	}
	rawItems := append(normalizePDFToolInputList(params["pdf"]), normalizePDFToolInputList(params["pdfs"])...)
	seen := map[string]bool{}
	out := make([]string, 0, len(rawItems))
	for _, item := range rawItems {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" || seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		out = append(out, trimmed)
	}
	if len(rawItems) > 0 && len(out) == 0 {
		return nil, fmt.Errorf("pdfs must include at least one path")
	}
	if len(out) > maxInputs {
		return nil, fmt.Errorf("too many pdfs requested: %d > %d", len(out), maxInputs)
	}
	return out, nil
}

func normalizePDFToolInputList(raw any) []string {
	switch typed := raw.(type) {
	case string:
		if strings.TrimSpace(typed) != "" {
			return []string{typed}
		}
	case []string:
		return append([]string(nil), typed...)
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			if value, ok := item.(string); ok {
				out = append(out, value)
			}
		}
		return out
	}
	return nil
}

func normalizePDFToolTimeoutMs(value int) int {
	if value <= 0 {
		value = defaultPDFRemoteTimeoutMs
	}
	if value > hardPDFRemoteTimeoutMs {
		value = hardPDFRemoteTimeoutMs
	}
	return value
}

func normalizePDFToolMaxResponseBytes(value int) int {
	if value <= 0 {
		value = defaultPDFRemoteMaxBytes
	}
	if value > hardPDFRemoteMaxBytes {
		value = hardPDFRemoteMaxBytes
	}
	return value
}

func pdfToolRequestTimeoutMs(params map[string]any, configured int) int {
	effective := normalizePDFToolTimeoutMs(configured)
	if requested := firstInt(params, "timeout_ms"); requested > 0 && requested < effective {
		effective = requested
	}
	return effective
}

func pdfToolRequestMaxResponseBytes(params map[string]any, configured int) int {
	effective := normalizePDFToolMaxResponseBytes(configured)
	if requested := firstInt(params, "max_response_bytes"); requested > 0 && requested < effective {
		effective = requested
	}
	for _, key := range []string{"max_bytes_mb", "maxBytesMb"} {
		value, ok := readFloat(params, key)
		if ok && value > 0 {
			requested := int(value * 1024 * 1024)
			if requested > 0 && requested < effective {
				effective = requested
			}
		}
	}
	return effective
}
