package tools

import "testing"

func TestBuildBrowserDoctorBringupReportPrefersHiddenManagedDefaultCandidate(t *testing.T) {
	report := BuildBrowserDoctorBringupReport(&BrowserDoctorBringupSummary{
		SubstrateSource:         "managed_browserd",
		SubstrateEndpoint:       "http://127.0.0.1:42123",
		PreferredRuntimeBackend: "system",
		PreferredRuntimeProfile: "default",
		PreferredRuntimeTarget:  "host",
		PrimaryStep:             "browser action=doctor",
		DoctorAction:            "browser action=doctor",
		RepairAction:            "browser action=repair",
		ReadyAction:             "browser action=ready",
		Steps: []BrowserDoctorBringupStep{
			{Run: "browser action=doctor"},
			{Run: "browser action=repair", When: "if_suggested"},
			{Run: "browser action=ready"},
		},
	}, BrowserToolMetadataRouteHints{
		RuntimeInfo: BrowserRuntimeInfo{
			Backend: "system",
			Profile: "default",
			Target:  "host",
		},
		DefaultCandidateSource:   "managed_browserd",
		DefaultCandidateEndpoint: "http://127.0.0.1:43123",
		DefaultCandidateRoute: BrowserRuntimeInfo{
			Backend: "proxy",
			Profile: "isolated",
			Target:  "node",
		},
		HiddenDefaultCandidate: true,
		SubstratePosture:       BrowserSubstrateLegacySystemHost,
		SubstrateStatus:        BrowserSubstrateStatus("system", "host"),
		SelectionStrategy:      BrowserSubstrateSelectionPreferNodeOverLegacy,
		SubstrateReason:        "Bundled browserd bootstrap is blocked because Chromium executable is not installed.",
		RepairAvailable:        true,
	})
	if report == nil {
		t.Fatal("expected browser doctor bring-up report")
	}
	if report.PreferredRuntimeBackend != "proxy" ||
		report.PreferredRuntimeProfile != "isolated" ||
		report.PreferredRuntimeTarget != "node" {
		t.Fatalf("expected preferred runtime to follow hidden managed default candidate, got %#v", report)
	}
	if report.DefaultCandidate != "hidden_managed" ||
		report.DefaultCandidateSource != "managed_browserd" ||
		report.DefaultCandidateEndpoint != "http://127.0.0.1:43123" ||
		report.DefaultCandidateBackend != "proxy" ||
		report.DefaultCandidateProfile != "isolated" ||
		report.DefaultCandidateTarget != "node" {
		t.Fatalf("expected explicit hidden managed candidate in bring-up report, got %#v", report)
	}
	if report.SubstrateBackend != "system" ||
		report.SubstrateProfile != "default" ||
		report.SubstrateTarget != "host" ||
		report.SubstratePosture != BrowserSubstrateLegacySystemHost ||
		report.SubstrateStatus != BrowserSubstrateStatus("system", "host") {
		t.Fatalf("expected current substrate route in bring-up report, got %#v", report)
	}
	if report.SubstrateSource != "managed_browserd" || report.SubstrateEndpoint != "http://127.0.0.1:42123" {
		t.Fatalf("expected substrate provenance in bring-up report, got %#v", report)
	}
	if report.SelectionStrategy != BrowserSubstrateSelectionPreferNodeOverLegacy ||
		report.SubstrateReason == "" ||
		!report.RepairAvailable {
		t.Fatalf("expected substrate hints in bring-up report, got %#v", report)
	}
}

func TestBrowserDoctorBringupReportDetailTextIncludesSubstrateHints(t *testing.T) {
	report := BuildBrowserDoctorBringupReport(&BrowserDoctorBringupSummary{
		SubstrateSource:         "managed_browserd",
		SubstrateEndpoint:       "http://127.0.0.1:42123",
		PreferredRuntimeBackend: "proxy",
		PreferredRuntimeProfile: "isolated",
		PreferredRuntimeTarget:  "node",
		PrimaryStep:             "browser action=doctor",
		Steps: []BrowserDoctorBringupStep{
			{Run: "browser action=doctor"},
			{Run: "browser action=repair", When: "if_suggested"},
			{Run: "browser action=ready"},
			{Run: "bash /tmp/browserd-platform-acceptance-check.sh", When: "optional_validation"},
		},
	}, BrowserToolMetadataRouteHints{
		RuntimeInfo: BrowserRuntimeInfo{
			Backend: "system",
			Profile: "default",
			Target:  "host",
		},
		DefaultCandidateSource:   "managed_browserd",
		DefaultCandidateEndpoint: "http://127.0.0.1:43123",
		DefaultCandidateRoute: BrowserRuntimeInfo{
			Backend: "proxy",
			Profile: "isolated",
			Target:  "node",
		},
		HiddenDefaultCandidate: true,
		SubstratePosture:       BrowserSubstrateLegacySystemHost,
		SubstrateStatus:        BrowserSubstrateStatus("system", "host"),
		SelectionStrategy:      BrowserSubstrateSelectionPreferNodeOverLegacy,
		SubstrateReason:        "Bundled browserd bootstrap is blocked because Chromium executable is not installed.",
		RepairAvailable:        true,
	})
	got := BrowserDoctorBringupReportDetailText(report)
	want := "browser action=doctor -> browser action=repair (if suggested) -> browser action=ready -> validate: bash /tmp/browserd-platform-acceptance-check.sh [primary_step=browser action=doctor substrate_source=managed_browserd substrate_endpoint=http://127.0.0.1:42123 substrate_backend=system substrate_profile=default substrate_target=host substrate_posture=legacy_system_host substrate_status=warn preferred_runtime_backend=proxy preferred_runtime_profile=isolated preferred_runtime_target=node default_candidate=hidden_managed default_candidate_source=managed_browserd default_candidate_endpoint=http://127.0.0.1:43123 default_candidate_backend=proxy default_candidate_profile=isolated default_candidate_target=node selection_strategy=prefer_node_over_legacy_host substrate_reason=\"Bundled browserd bootstrap is blocked because Chromium executable is not installed.\" repair_available=true]"
	if got != want {
		t.Fatalf("unexpected browser doctor bring-up report detail text: %q", got)
	}
}

func TestBuildBrowserDoctorBringupReportMarksRepairAvailableFromBringup(t *testing.T) {
	report := BuildBrowserDoctorBringupReport(&BrowserDoctorBringupSummary{
		PreferredRuntimeBackend: "proxy",
		PreferredRuntimeProfile: "isolated",
		PreferredRuntimeTarget:  "node",
		PrimaryStep:             "browser action=doctor",
		DoctorAction:            "browser action=doctor",
		RepairAction:            "browser action=repair",
		ReadyAction:             "browser action=ready",
	}, BrowserToolMetadataRouteHints{
		RuntimeInfo: BrowserRuntimeInfo{
			Backend: "proxy",
			Profile: "isolated",
			Target:  "node",
		},
		SelectionStrategy: BrowserSubstrateSelectionPreferNodeOverLegacy,
	})
	if report == nil || !report.RepairAvailable {
		t.Fatalf("expected repair_available from bringup repair action, got %#v", report)
	}
}

func TestBrowserDoctorBringupDisplayTextFallsBackToBringupSummary(t *testing.T) {
	got := BrowserDoctorBringupDisplayText(nil, &BrowserDoctorBringupSummary{
		PreferredRuntimeBackend: "proxy",
		PreferredRuntimeProfile: "isolated",
		PreferredRuntimeTarget:  "node",
		PrimaryStep:             "browser action=doctor",
		Steps: []BrowserDoctorBringupStep{
			{Run: "browser action=doctor"},
			{Run: "browser action=ready"},
		},
	})
	if got != "browser action=doctor -> browser action=ready [primary_step=browser action=doctor preferred_runtime_backend=proxy preferred_runtime_profile=isolated preferred_runtime_target=node]" {
		t.Fatalf("unexpected browser doctor bring-up display text: %q", got)
	}
}
