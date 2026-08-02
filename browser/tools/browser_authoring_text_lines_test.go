package tools

import (
	"strings"
	"testing"
)

func TestBuildBrowserAuthoringTextSummariesBuildsCanonicalSummaries(t *testing.T) {
	summaries := BuildBrowserAuthoringTextSummaries(
		[]string{
			"group:browser-default-selection:prefer_node_over_legacy_host",
			"group:browser-bootstrap-repairable",
		},
		&BrowserToolMetadataRouteHints{
			RuntimeInfo: BrowserRuntimeInfo{
				Backend: "system",
				Profile: "default",
				Target:  "host",
			},
			DefaultCandidateRoute: BrowserRuntimeInfo{
				Backend: "proxy",
				Profile: "isolated",
				Target:  "node",
			},
			HiddenDefaultCandidate: true,
			SelectionStrategy:      BrowserSubstrateSelectionPreferNodeOverLegacy,
			SubstrateReason:        "Bundled browserd bootstrap is blocked because Chromium executable is not installed.",
		},
		&BrowserDoctorBringupReport{
			Summary: "browser action=doctor",
		},
	)
	if summaries.PolicySelectors != "group:browser-default-selection:prefer_node_over_legacy_host, group:browser-bootstrap-repairable" {
		t.Fatalf("unexpected policy selector summary: %q", summaries.PolicySelectors)
	}
	if !strings.Contains(summaries.RouteHints, "Browser route note: current tool metadata indicates a hidden managed browserd default candidate on `proxy/isolated/node`") {
		t.Fatalf("unexpected route hints summary: %q", summaries.RouteHints)
	}
	if summaries.Bringup != "browser action=doctor" {
		t.Fatalf("unexpected bring-up summary: %q", summaries.Bringup)
	}
}

func TestBuildBrowserAuthoringTextLinesBuildsCanonicalLines(t *testing.T) {
	summaries := BrowserAuthoringTextSummaries{
		PolicySelectors: "group:browser-default-selection:prefer_node_over_legacy_host, group:browser-bootstrap-repairable",
		RouteHints:      "Browser route note: current tool metadata indicates a hidden managed browserd default candidate on `proxy/isolated/node`.",
		Bringup:         "browser action=doctor",
	}
	lines := summaries.Lines(BrowserAuthoringTextLineLabels{
		PolicySelectors: "browser policy selectors: ",
		RouteHints:      "browser route hints: ",
		Bringup:         "browser bring-up: ",
	})
	if len(lines) != 3 {
		t.Fatalf("expected three authoring lines, got %#v", lines)
	}
	if lines[0] != "browser policy selectors: group:browser-default-selection:prefer_node_over_legacy_host, group:browser-bootstrap-repairable" {
		t.Fatalf("unexpected policy selector line: %q", lines[0])
	}
	if !strings.Contains(lines[1], "browser route hints: Browser route note: current tool metadata indicates a hidden managed browserd default candidate on `proxy/isolated/node`") {
		t.Fatalf("unexpected route hints line: %q", lines[1])
	}
	if lines[2] != "browser bring-up: browser action=doctor" {
		t.Fatalf("unexpected bring-up line: %q", lines[2])
	}
}

func TestBuildBrowserAuthoringTextLinesSkipsEmptySections(t *testing.T) {
	lines := BrowserAuthoringTextSummaries{}.Lines(BrowserAuthoringTextLineLabels{
		PolicySelectors: "browser policy selectors: ",
		RouteHints:      "browser route hints: ",
		Bringup:         "browser bring-up: ",
	})
	if len(lines) != 0 {
		t.Fatalf("expected no lines for empty inputs, got %#v", lines)
	}
}
