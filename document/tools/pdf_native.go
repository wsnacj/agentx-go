package tools

import (
	"context"
	"fmt"
	"strings"
)

func pdfVisionCandidateConfigName(candidate pdfVisionModelCandidate) string {
	if value := strings.TrimSpace(candidate.ConfigKey); value != "" {
		return value
	}
	return strings.TrimSpace(candidate.Name)
}

func buildPDFNativePrompt(query string, mediaProfile pdfMediaProfile, plan pdfAnalysisPlan) string {
	mode := selectPDFVisualPromptMode(mediaProfile, plan)
	lines := []string{
		"Analyze the provided PDF document directly.",
		fmt.Sprintf("Planned mode: %s.", plan.Mode),
		fmt.Sprintf("Specialized native PDF prompt mode: %s.", mode),
		"Focus on layout, charts, tables, diagrams, OCR-visible text, and any high-signal visual structure that plain text extraction may miss.",
		"Return a concise, high-signal summary grounded in the visible document content.",
	}
	if trimmed := strings.TrimSpace(query); trimmed != "" {
		lines = append(lines, fmt.Sprintf("Focus query: %s.", trimmed))
	}
	lines = append(lines, pdfVisualPromptInstructions(mode)...)
	return strings.Join(lines, "\n")
}

func runPDFNativeAnalysis(ctx context.Context, candidate pdfVisionModelCandidate, prompt string, paths []string) (string, error) {
	host := pdfHostFromContext(ctx)
	if host.Native == nil {
		return "", fmt.Errorf("native pdf analyzer is not configured")
	}
	return host.Native(ctx, PDFNativeRequest{
		Candidate: candidate,
		Prompt:    prompt,
		Paths:     append([]string(nil), paths...),
	})
}
