package tools

import "strings"

type BrowserAuthoringTextLineLabels struct {
	PolicySelectors string
	RouteHints      string
	Bringup         string
}

type BrowserAuthoringTextSummaries struct {
	PolicySelectors string
	RouteHints      string
	Bringup         string
}

func BuildBrowserAuthoringTextSummaries(
	policySelectors []string,
	routeHints *BrowserToolMetadataRouteHints,
	bringup *BrowserDoctorBringupReport,
) BrowserAuthoringTextSummaries {
	summaries := BrowserAuthoringTextSummaries{}
	if len(policySelectors) > 0 {
		summaries.PolicySelectors = strings.Join(policySelectors, ", ")
	}
	if routeHints != nil {
		summaries.RouteHints = strings.TrimSpace(BrowserToolMetadataRouteHintsDisplayText(*routeHints))
	}
	if bringup != nil {
		summaries.Bringup = strings.TrimSpace(BrowserDoctorBringupDisplayText(bringup, nil))
	}
	return summaries
}

func (s BrowserAuthoringTextSummaries) Lines(labels BrowserAuthoringTextLineLabels) []string {
	lines := make([]string, 0, 3)
	if prefix := labels.PolicySelectors; prefix != "" && strings.TrimSpace(s.PolicySelectors) != "" {
		lines = append(lines, prefix+s.PolicySelectors)
	}
	if prefix := labels.RouteHints; prefix != "" && strings.TrimSpace(s.RouteHints) != "" {
		lines = append(lines, prefix+s.RouteHints)
	}
	if prefix := labels.Bringup; prefix != "" && strings.TrimSpace(s.Bringup) != "" {
		lines = append(lines, prefix+s.Bringup)
	}
	return lines
}

func BuildBrowserAuthoringTextLines(
	labels BrowserAuthoringTextLineLabels,
	policySelectors []string,
	routeHints *BrowserToolMetadataRouteHints,
	bringup *BrowserDoctorBringupReport,
) []string {
	return BuildBrowserAuthoringTextSummaries(policySelectors, routeHints, bringup).Lines(labels)
}
