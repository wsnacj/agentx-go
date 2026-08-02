package tools

import "strings"

type browserRuntimeLaunchDiagnosticsSummary struct {
	Source                              string `json:"source,omitempty"`
	Backend                             string `json:"backend,omitempty"`
	Profile                             string `json:"profile,omitempty"`
	RuntimeTarget                       string `json:"runtime_target,omitempty"`
	BrowserApp                          string `json:"browser_app,omitempty"`
	Status                              string `json:"status,omitempty"`
	Running                             *bool  `json:"running,omitempty"`
	Connected                           *bool  `json:"connected,omitempty"`
	Note                                string `json:"note,omitempty"`
	RepairCommand                       string `json:"repair_command,omitempty"`
	HostOS                              string `json:"host_os,omitempty"`
	HostArch                            string `json:"host_arch,omitempty"`
	PlaywrightCachePath                 string `json:"playwright_cache_path,omitempty"`
	PlaywrightCacheSource               string `json:"playwright_cache_source,omitempty"`
	NodeVersion                         string `json:"node_version,omitempty"`
	PlaywrightPackage                   string `json:"playwright_package,omitempty"`
	PlaywrightPackageVersion            string `json:"playwright_package_version,omitempty"`
	RuntimeSummaryGeneration            string `json:"runtime_summary_generation,omitempty"`
	RuntimeBaselineReady                *bool  `json:"runtime_baseline_ready,omitempty"`
	RuntimeBaselineBlockReason          string `json:"runtime_baseline_block_reason,omitempty"`
	SelectedLaunchSource                string `json:"selected_launch_source,omitempty"`
	SelectedLaunchDeliveryGeneration    string `json:"selected_launch_delivery_generation,omitempty"`
	SelectedLaunchPayloadSource         string `json:"selected_launch_payload_source,omitempty"`
	SelectedLaunchPayloadReady          *bool  `json:"selected_launch_payload_ready,omitempty"`
	SelectedLaunchPayloadBlockReason    string `json:"selected_launch_payload_block_reason,omitempty"`
	SelectedLaunchReady                 *bool  `json:"selected_launch_ready,omitempty"`
	SelectedLaunchBlockReason           string `json:"selected_launch_block_reason,omitempty"`
	SelectedLaunchExecutableReady       *bool  `json:"selected_launch_executable_ready,omitempty"`
	SelectedLaunchExecutableBlockReason string `json:"selected_launch_executable_block_reason,omitempty"`
	DeliveryTransitionPending           *bool  `json:"delivery_transition_pending,omitempty"`
	DeliveryTransitionStage             string `json:"delivery_transition_stage,omitempty"`
	LaunchReady                         *bool  `json:"launch_ready,omitempty"`
	LaunchBlockReason                   string `json:"launch_block_reason,omitempty"`
	BundleReady                         *bool  `json:"bundle_ready,omitempty"`
	DeliveryReady                       *bool  `json:"delivery_ready,omitempty"`
	NodeModulesReady                    *bool  `json:"node_modules_ready,omitempty"`
	BrowserReady                        *bool  `json:"browser_ready,omitempty"`
	BootstrapState                      string `json:"bootstrap_state,omitempty"`
	BootstrapErrorCode                  string `json:"bootstrap_error_code,omitempty"`
}

func browserBoolPtr(v bool) *bool {
	return &v
}

func browserRuntimeCloneLaunchDiagnosticsSummary(
	summary *browserRuntimeLaunchDiagnosticsSummary,
) *browserRuntimeLaunchDiagnosticsSummary {
	if summary == nil {
		return nil
	}
	cloned := *summary
	return &cloned
}

func browserRuntimeLaunchDiagnosticsSummaryEmpty(summary browserRuntimeLaunchDiagnosticsSummary) bool {
	return strings.TrimSpace(summary.Source) == "" &&
		strings.TrimSpace(summary.Backend) == "" &&
		strings.TrimSpace(summary.Profile) == "" &&
		strings.TrimSpace(summary.RuntimeTarget) == "" &&
		strings.TrimSpace(summary.BrowserApp) == "" &&
		strings.TrimSpace(summary.Status) == "" &&
		summary.Running == nil &&
		summary.Connected == nil &&
		strings.TrimSpace(summary.Note) == "" &&
		strings.TrimSpace(summary.RepairCommand) == "" &&
		strings.TrimSpace(summary.HostOS) == "" &&
		strings.TrimSpace(summary.HostArch) == "" &&
		strings.TrimSpace(summary.PlaywrightCachePath) == "" &&
		strings.TrimSpace(summary.PlaywrightCacheSource) == "" &&
		strings.TrimSpace(summary.NodeVersion) == "" &&
		strings.TrimSpace(summary.PlaywrightPackage) == "" &&
		strings.TrimSpace(summary.PlaywrightPackageVersion) == "" &&
		strings.TrimSpace(summary.RuntimeSummaryGeneration) == "" &&
		summary.RuntimeBaselineReady == nil &&
		strings.TrimSpace(summary.RuntimeBaselineBlockReason) == "" &&
		strings.TrimSpace(summary.SelectedLaunchSource) == "" &&
		strings.TrimSpace(summary.SelectedLaunchDeliveryGeneration) == "" &&
		strings.TrimSpace(summary.SelectedLaunchPayloadSource) == "" &&
		summary.SelectedLaunchPayloadReady == nil &&
		strings.TrimSpace(summary.SelectedLaunchPayloadBlockReason) == "" &&
		summary.SelectedLaunchReady == nil &&
		strings.TrimSpace(summary.SelectedLaunchBlockReason) == "" &&
		summary.SelectedLaunchExecutableReady == nil &&
		strings.TrimSpace(summary.SelectedLaunchExecutableBlockReason) == "" &&
		summary.DeliveryTransitionPending == nil &&
		strings.TrimSpace(summary.DeliveryTransitionStage) == "" &&
		summary.LaunchReady == nil &&
		strings.TrimSpace(summary.LaunchBlockReason) == "" &&
		summary.BundleReady == nil &&
		summary.DeliveryReady == nil &&
		summary.NodeModulesReady == nil &&
		summary.BrowserReady == nil &&
		strings.TrimSpace(summary.BootstrapState) == "" &&
		strings.TrimSpace(summary.BootstrapErrorCode) == ""
}

func browserRuntimeLaunchDiagnosticsSummaryFromStatusResult(
	repairScript string,
	result BrowserProfileStatusResult,
	selectedInfo BrowserRuntimeInfo,
) *browserRuntimeLaunchDiagnosticsSummary {
	cache := result.PlaywrightCache
	if cache == nil {
		return nil
	}
	bootstrapCode := firstNonEmpty(
		strings.TrimSpace(cache.BootstrapErrorCode),
		browserRuntimeBootstrapErrorCodeFromFailureText(
			result.Note,
			cache.LaunchBlockReason,
			cache.SelectedLaunchBlockReason,
			cache.SelectedLaunchExecutableBR,
			cache.RuntimeBaselineBlockReason,
		),
	)
	selectedInfo = normalizeBrowserRuntimeInfo(selectedInfo)
	summary := browserRuntimeLaunchDiagnosticsSummary{
		Source:                              "runtime_status",
		Backend:                             firstNonEmpty(strings.TrimSpace(result.Backend), strings.TrimSpace(selectedInfo.Backend)),
		Profile:                             firstNonEmpty(strings.TrimSpace(result.Profile), strings.TrimSpace(selectedInfo.Profile)),
		RuntimeTarget:                       strings.TrimSpace(selectedInfo.Target),
		BrowserApp:                          strings.TrimSpace(result.BrowserApp),
		Status:                              strings.TrimSpace(result.Status),
		Running:                             browserBoolPtr(result.Running),
		Connected:                           browserBoolPtr(result.Connected),
		Note:                                strings.TrimSpace(result.Note),
		RepairCommand:                       browserRuntimeBootstrapRepairCommand(repairScript, bootstrapCode),
		HostOS:                              strings.TrimSpace(cache.HostOS),
		HostArch:                            strings.TrimSpace(cache.HostArch),
		PlaywrightCachePath:                 strings.TrimSpace(cache.Path),
		PlaywrightCacheSource:               strings.TrimSpace(cache.Source),
		NodeVersion:                         strings.TrimSpace(cache.NodeVersion),
		PlaywrightPackage:                   strings.TrimSpace(cache.PlaywrightPackage),
		PlaywrightPackageVersion:            strings.TrimSpace(cache.PlaywrightPackageVersion),
		RuntimeSummaryGeneration:            strings.TrimSpace(cache.RuntimeSummaryGeneration),
		RuntimeBaselineReady:                browserBoolPtr(cache.RuntimeBaselineReady),
		RuntimeBaselineBlockReason:          strings.TrimSpace(cache.RuntimeBaselineBlockReason),
		SelectedLaunchSource:                strings.TrimSpace(cache.SelectedLaunchSource),
		SelectedLaunchDeliveryGeneration:    strings.TrimSpace(cache.SelectedLaunchDelivery),
		SelectedLaunchPayloadSource:         strings.TrimSpace(cache.SelectedLaunchPayloadSrc),
		SelectedLaunchPayloadReady:          browserBoolPtr(cache.SelectedLaunchPayloadReady),
		SelectedLaunchPayloadBlockReason:    strings.TrimSpace(cache.SelectedLaunchPayloadBR),
		SelectedLaunchReady:                 browserBoolPtr(cache.SelectedLaunchReady),
		SelectedLaunchBlockReason:           strings.TrimSpace(cache.SelectedLaunchBlockReason),
		SelectedLaunchExecutableReady:       browserBoolPtr(cache.SelectedLaunchExecutableOK),
		SelectedLaunchExecutableBlockReason: strings.TrimSpace(cache.SelectedLaunchExecutableBR),
		DeliveryTransitionPending:           browserBoolPtr(cache.DeliveryTransitionPending),
		DeliveryTransitionStage:             strings.TrimSpace(cache.DeliveryTransitionStage),
		LaunchReady:                         browserBoolPtr(cache.LaunchReady),
		LaunchBlockReason:                   strings.TrimSpace(cache.LaunchBlockReason),
		BundleReady:                         browserBoolPtr(cache.BundleReady),
		DeliveryReady:                       browserBoolPtr(cache.DeliveryReady),
		NodeModulesReady:                    browserBoolPtr(cache.NodeModulesReady),
		BrowserReady:                        browserBoolPtr(cache.BrowserReady),
		BootstrapState:                      strings.TrimSpace(cache.BootstrapState),
		BootstrapErrorCode:                  bootstrapCode,
	}
	if browserRuntimeLaunchDiagnosticsSummaryEmpty(summary) {
		return nil
	}
	return &summary
}

func browserRuntimeLaunchDiagnosticsSummaryFromManagedLaunchFailure(
	ctx browserRegistrationContext,
	preview browserRuntimeDiagnosticsPreview,
	note string,
) *browserRuntimeLaunchDiagnosticsSummary {
	info := normalizeBrowserRuntimeInfo(browserRuntimeManagedLaunchFailurePrimaryInfoForPreview(ctx, preview))
	target := strings.TrimSpace(info.Target)
	if target == "" {
		target = browserRuntimeManagedLaunchFailurePrimaryTarget(preview)
	}
	blockReason := "managed_route_unavailable"
	if assessment, ok := browserRuntimePreviewManagedRouteAssessment(preview, target); ok {
		blockReason = firstNonEmpty(strings.TrimSpace(assessment.FailureReason), blockReason)
		note = firstNonEmpty(strings.TrimSpace(note), strings.TrimSpace(assessment.FailureNote))
	}
	bootstrapCode := browserRuntimeBootstrapErrorCodeFromFailureText(note, blockReason)
	summary := browserRuntimeLaunchDiagnosticsSummary{
		Source:             "route_assessment",
		Backend:            strings.TrimSpace(info.Backend),
		Profile:            strings.TrimSpace(info.Profile),
		RuntimeTarget:      target,
		Status:             "unsupported",
		Note:               strings.TrimSpace(note),
		RepairCommand:      browserRuntimeBootstrapRepairCommand(ctx.opts.RepairScript, bootstrapCode),
		LaunchReady:        browserBoolPtr(false),
		LaunchBlockReason:  strings.TrimSpace(blockReason),
		BootstrapState:     browserRuntimeLaunchBootstrapStateForErrorCode(bootstrapCode),
		BootstrapErrorCode: bootstrapCode,
	}
	if browserRuntimeLaunchDiagnosticsSummaryEmpty(summary) {
		return nil
	}
	return &summary
}

func browserRuntimeLaunchBootstrapStateForErrorCode(code string) string {
	code = strings.TrimSpace(code)
	if code == "" {
		return ""
	}
	return "failed"
}

func browserRuntimeBootstrapErrorCodeFromFailureText(values ...string) string {
	for _, value := range values {
		lower := strings.ToLower(strings.TrimSpace(value))
		browserInstallContext := strings.Contains(lower, "bootstrap playwright browser") ||
			strings.Contains(lower, "install chromium") ||
			strings.Contains(lower, "playwright download")
		switch {
		case lower == "":
			continue
		case strings.Contains(lower, "npm_missing"),
			strings.Contains(lower, "requires npm in path"):
			return "npm_missing"
		case strings.Contains(lower, "npm_ci_failed"),
			strings.Contains(lower, "npm ci"):
			return "npm_ci_failed"
		case strings.Contains(lower, "playwright_dependency_missing"),
			strings.Contains(lower, "playwright dependency"):
			return "playwright_dependency_missing"
		case strings.Contains(lower, "node_missing"),
			strings.Contains(lower, "requires node in path"):
			return "node_missing"
		case strings.Contains(lower, "playwright_cli_missing"),
			strings.Contains(lower, "playwright cli missing"):
			return "playwright_cli_missing"
		case strings.Contains(lower, "browser_install_timeout"),
			(browserInstallContext &&
				(strings.Contains(lower, "timed out after") ||
					strings.Contains(lower, "context deadline exceeded"))):
			return "browser_install_timeout"
		case strings.Contains(lower, "browser_install_network_failed"),
			(browserInstallContext &&
				(strings.Contains(lower, "econnreset") ||
					strings.Contains(lower, "enotfound") ||
					strings.Contains(lower, "econnrefused") ||
					strings.Contains(lower, "eai_again") ||
					strings.Contains(lower, "network is unreachable") ||
					strings.Contains(lower, "self signed certificate") ||
					strings.Contains(lower, "unable to get local issuer certificate") ||
					strings.Contains(lower, "certificate has expired") ||
					strings.Contains(lower, "download failed") ||
					strings.Contains(lower, "fetch failed"))):
			return "browser_install_network_failed"
		case strings.Contains(lower, "browser_install_failed"),
			strings.Contains(lower, "install chromium"):
			return "browser_install_failed"
		case strings.Contains(lower, "selected_launch_payload_not_ready"),
			strings.Contains(lower, "retained_fallback_payload_not_ready"),
			strings.Contains(lower, "retained_delivery_cache_not_ready"),
			strings.Contains(lower, "current_delivery_payload_not_ready"),
			strings.Contains(lower, "target_delivery_payload_not_ready"):
			return "browser_executable_missing"
		case strings.Contains(lower, "browser_executable_missing"),
			strings.Contains(lower, "browser executable missing"),
			strings.Contains(lower, "browser is still unavailable after bootstrap"),
			strings.Contains(lower, "selected_launch_executable_not_ready"),
			strings.Contains(lower, "selected_launch_executable_not_resolved"):
			return "browser_executable_missing"
		}
	}
	return ""
}

func browserRuntimeBootstrapBlockedSurfaceNoteForFailureText(values ...string) string {
	code := browserRuntimeBootstrapErrorCodeFromFailureText(values...)
	if strings.TrimSpace(code) == "" {
		return ""
	}
	return browserRuntimeDoctorLaunchBootstrapBlockedSummary(&browserRuntimeLaunchDiagnosticsSummary{
		BootstrapErrorCode: code,
	})
}

func browserRuntimeBootstrapErrorCodeSupportsRepair(code string) bool {
	switch strings.TrimSpace(code) {
	case "npm_ci_failed", "playwright_dependency_missing", "playwright_cli_missing", "browser_install_failed", "browser_install_timeout", "browser_install_network_failed", "browser_executable_missing":
		return true
	default:
		return false
	}
}

func browserRuntimeBootstrapRepairCommand(repairScript string, bootstrapCode string) string {
	if !browserRuntimeBootstrapErrorCodeSupportsRepair(bootstrapCode) {
		return ""
	}
	return browserHostScriptCommand(repairScript)
}

func browserRuntimeApplyLaunchDiagnostics(
	payload *browserRuntimePayload,
	action string,
	summary *browserRuntimeLaunchDiagnosticsSummary,
) {
	if payload == nil || summary == nil {
		return
	}
	payload.LaunchDiagnostics = browserRuntimeCloneLaunchDiagnosticsSummary(summary)
	if browserRuntimeCanonicalAction(action) == "workbench" {
		payload.WorkbenchLaunchDiagnostics = browserRuntimeCloneLaunchDiagnosticsSummary(summary)
	}
	browserRuntimeApplyRepairCommandToPayloadShells(payload)
}

func browserRuntimeUsesLifecycleLaunchDiagnostics(action string) bool {
	switch browserRuntimeCanonicalAction(action) {
	case "prepare", "coordinate", "start":
		return true
	default:
		return false
	}
}

func browserRuntimeMaybeApplyLifecycleLaunchDiagnostics(
	repairScript string,
	payload *browserRuntimePayload,
	action string,
	result browserRuntimePrepareResult,
	selectedInfo BrowserRuntimeInfo,
) {
	if payload == nil || !browserRuntimeUsesLifecycleLaunchDiagnostics(action) {
		return
	}
	browserRuntimeApplyLaunchDiagnostics(
		payload,
		action,
		browserRuntimeLaunchDiagnosticsSummaryFromStatusResult(repairScript, result.ProfileStatus, selectedInfo),
	)
}
