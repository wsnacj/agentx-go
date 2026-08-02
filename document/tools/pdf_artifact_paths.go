package tools

import (
	"strings"

	agentxmedia "github.com/wsnacj/agentx-go/runtime/mediaartifact"
)

func pdfRenderedPageTouchedPaths(rendered []agentxmedia.Descriptor) []string {
	if len(rendered) == 0 {
		return nil
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(rendered))
	for _, item := range rendered {
		path := strings.TrimSpace(item.Path)
		if path == "" || seen[path] {
			continue
		}
		seen[path] = true
		out = append(out, path)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func pdfVisualAnalysisTouchedPaths(analysis *pdfVisualAnalysis) []string {
	if analysis == nil {
		return nil
	}
	return pdfRenderedPageTouchedPaths(analysis.RenderedPages)
}

func pdfTouchedPathsFromVisualAnalyses(analyses ...*pdfVisualAnalysis) []string {
	if len(analyses) == 0 {
		return nil
	}
	seen := map[string]bool{}
	out := make([]string, 0)
	for _, analysis := range analyses {
		for _, path := range pdfVisualAnalysisTouchedPaths(analysis) {
			if seen[path] {
				continue
			}
			seen[path] = true
			out = append(out, path)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
