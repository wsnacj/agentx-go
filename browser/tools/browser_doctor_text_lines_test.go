package tools

import (
	"strings"
	"testing"
)

func boolPtr(v bool) *bool {
	return &v
}

func TestBuildBrowserDoctorDisplayLinesBuildsCanonicalLines(t *testing.T) {
	lines := BuildBrowserDoctorDisplayLines(&BrowserDoctorSummary{
		Status: "pending",
		Ready:  false,
		Route: &BrowserDoctorRouteSummary{
			BrowserDoctorCheckSummary: BrowserDoctorCheckSummary{
				Status:  "ok",
				Code:    "managed_route_configured",
				Summary: "managed browser route is configured",
			},
		},
		Launch: &BrowserDoctorLaunchSummary{
			BrowserDoctorCheckSummary: BrowserDoctorCheckSummary{
				Status:  "pending",
				Code:    "bootstrap_pending",
				Summary: "browser daemon launch still pending",
			},
			BootstrapState:                      "failed",
			BootstrapErrorCode:                  "npm_missing",
			NodeVersion:                         "v22.0.0",
			PlaywrightPackage:                   "playwright",
			PlaywrightPackageVersion:            "1.52.0",
			RuntimeBaselineReady:                boolPtr(false),
			RuntimeBaselineBlockReason:          "runtime_observed_launch_missing",
			SelectedLaunchExecutableReady:       boolPtr(false),
			SelectedLaunchExecutableBlockReason: "selected_launch_executable_not_ready",
		},
		RepairCommand:     "bash /tmp/browserd-bootstrap-repair.sh",
		AcceptanceCommand: "bash /tmp/browserd-platform-acceptance-check.sh",
		Suggestions:       []string{"repair browserd bootstrap"},
	}, &BrowserDoctorBringupReport{
		Summary: "browser action=doctor -> browser action=repair -> browser action=ready",
	})
	if len(lines) != 8 {
		t.Fatalf("expected eight browser doctor lines, got %#v", lines)
	}
	if lines[0] != "browser doctor: status=pending ready=false route=ok launch=pending" {
		t.Fatalf("unexpected status line: %q", lines[0])
	}
	if lines[1] != "- browser route [ok/managed_route_configured]: managed browser route is configured" {
		t.Fatalf("unexpected route line: %q", lines[1])
	}
	if lines[2] != "- browser launch [pending/bootstrap_pending]: browser daemon launch still pending" {
		t.Fatalf("unexpected launch line: %q", lines[2])
	}
	if lines[3] != "- browser bring-up: browser action=doctor -> browser action=repair -> browser action=ready" {
		t.Fatalf("unexpected bring-up line: %q", lines[3])
	}
	if !strings.Contains(lines[4], "bootstrap=failed code=npm_missing node=v22.0.0 playwright=1.52.0 baseline=runtime_observed_launch_missing executable=selected_launch_executable_not_ready") {
		t.Fatalf("unexpected launch detail line: %q", lines[4])
	}
	if lines[5] != "- browser repair: bash /tmp/browserd-bootstrap-repair.sh" {
		t.Fatalf("unexpected repair line: %q", lines[5])
	}
	if lines[6] != "- browser acceptance: bash /tmp/browserd-platform-acceptance-check.sh" {
		t.Fatalf("unexpected acceptance line ordering: %#v", lines)
	}
	if lines[7] != "- browser suggestion: repair browserd bootstrap" {
		t.Fatalf("unexpected suggestion line: %#v", lines)
	}
}

func TestBuildBrowserDoctorDisplayLinesIncludesSuggestions(t *testing.T) {
	lines := BuildBrowserDoctorDisplayLines(&BrowserDoctorSummary{
		Suggestions: []string{"repair browserd bootstrap", "run browser action=ready"},
	}, nil)
	if len(lines) != 3 {
		t.Fatalf("expected status plus suggestions, got %#v", lines)
	}
	if lines[1] != "- browser suggestion: repair browserd bootstrap" || lines[2] != "- browser suggestion: run browser action=ready" {
		t.Fatalf("unexpected suggestion lines: %#v", lines)
	}
}
