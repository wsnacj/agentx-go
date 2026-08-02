package tools

import (
	"fmt"
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

type browserDoctorRouteMetadata struct {
	Source   string
	Endpoint string
}

type browserDoctorRouteMetadataProvider interface {
	browserDoctorRouteMetadata() browserDoctorRouteMetadata
}

type browserRuntimeDoctorRouteProjection struct {
	Info     BrowserRuntimeInfo
	Metadata browserDoctorRouteMetadata
}

func browserRuntimeApplyDoctorSummary(
	ctx browserRegistrationContext,
	payload *browserRuntimePayload,
	preview browserRuntimeDiagnosticsPreview,
) {
	if payload == nil {
		return
	}
	doctor := browserRuntimeDoctorSummary(ctx, payload, preview)
	if doctor == nil {
		return
	}
	payload.Doctor = doctor
	browserRuntimeApplyRepairCommandToPayloadShells(payload)
}

func browserRuntimeDoctorSummary(
	ctx browserRegistrationContext,
	payload *browserRuntimePayload,
	preview browserRuntimeDiagnosticsPreview,
) *BrowserDoctorSummary {
	route := browserRuntimeDoctorRouteSummary(payload, preview)
	launch := browserRuntimeDoctorLaunchSummary(ctx, payload.LaunchDiagnostics)
	if route == nil && launch == nil {
		return nil
	}
	configuredTargets := append([]string(nil), payload.ConfiguredTargets...)
	if len(configuredTargets) == 0 {
		configuredTargets = append([]string(nil), preview.ConfiguredTargets...)
	}
	doctor := &BrowserDoctorSummary{
		Route:             route,
		Launch:            launch,
		ConfiguredTargets: configuredTargets,
		RepairCommand:     browserHostScriptCommand(ctx.opts.RepairScript),
		AcceptanceCommand: browserHostScriptCommand(ctx.opts.AcceptanceScript),
	}
	doctor.Status = agentxbrowserruntime.BrowserDoctorAggregateStatus(browserRuntimeDoctorRouteStatus(route), browserRuntimeDoctorLaunchStatus(launch))
	doctor.Ready = browserRuntimeDoctorRouteStatus(route) == agentxbrowserruntime.BrowserDoctorStatusOK &&
		browserRuntimeDoctorLaunchStatus(launch) == agentxbrowserruntime.BrowserDoctorStatusOK
	doctor.Bringup = browserRuntimeDoctorBringupSummary(ctx, doctor)
	doctor.Suggestions = browserRuntimeDoctorSuggestions(ctx, doctor)
	return doctor
}

func browserRuntimeDoctorRouteSummary(
	payload *browserRuntimePayload,
	preview browserRuntimeDiagnosticsPreview,
) *BrowserDoctorRouteSummary {
	if payload == nil {
		return nil
	}
	configuredTargets := append([]string(nil), payload.ConfiguredTargets...)
	if len(configuredTargets) == 0 {
		configuredTargets = append([]string(nil), preview.ConfiguredTargets...)
	}
	defaultRoute := payload.DefaultRoute
	if strings.TrimSpace(defaultRoute.RuntimeTarget) == "" {
		defaultRoute = preview.DefaultRouteDescriptor
	}
	selectionStrategy := strings.TrimSpace(payload.SubstrateSelectionStrategy)
	if selectionStrategy == "" {
		selectionStrategy = strings.TrimSpace(preview.Registration.SubstrateSummary.SelectionStrategy)
	}
	selectionReason := strings.TrimSpace(payload.SubstrateSelectionReason)
	if selectionReason == "" {
		selectionReason = strings.TrimSpace(preview.Registration.SubstrateSummary.SelectionReason)
	}
	routeSelectionStrategy := selectionStrategy
	actualInfo := normalizeBrowserRuntimeInfo(BrowserRuntimeInfo{
		Backend: defaultRoute.Backend,
		Profile: defaultRoute.Profile,
		Target:  defaultRoute.RuntimeTarget,
	})
	if actualInfo == (BrowserRuntimeInfo{}) &&
		!browserRuntimeUsesImplicitLegacyHostDefaultFallback(preview.DefaultRoute, preview.Registration.SubstrateAssessment) {
		actualInfo = normalizeBrowserRuntimeInfo(preview.DefaultRoute)
	}
	route := &BrowserDoctorRouteSummary{
		SelectionStrategy: selectionStrategy,
	}
	managedConfigured := containsString(configuredTargets, "node") || containsString(configuredTargets, "sandbox")
	projection := browserRuntimeDoctorVisibleDefaultRouteProjection(preview, actualInfo)
	if managedConfigured && (strings.TrimSpace(actualInfo.Target) == "" || strings.EqualFold(strings.TrimSpace(actualInfo.Target), "host")) {
		if managedProjection, ok := browserRuntimeDoctorManagedCandidateRouteProjection(preview); ok {
			projection = managedProjection
			routeSelectionStrategy = firstNonEmpty(
				strings.TrimSpace(browserRuntimeDoctorRouteInspectionSelectionStrategy(preview, managedProjection.Info)),
				routeSelectionStrategy,
			)
		}
	}
	projection.Info = firstBrowserRuntimeInfo(projection.Info, actualInfo)
	route.Backend = strings.TrimSpace(projection.Info.Backend)
	route.Profile = strings.TrimSpace(projection.Info.Profile)
	route.RuntimeTarget = strings.TrimSpace(projection.Info.Target)
	route.SelectionStrategy = strings.TrimSpace(routeSelectionStrategy)
	route.Source = firstNonEmpty(
		strings.TrimSpace(projection.Metadata.Source),
		browserRuntimeDoctorFallbackRouteSource(actualInfo, projection.Info, managedConfigured, selectionStrategy),
		"runtime_selection",
	)
	route.Endpoint = strings.TrimSpace(projection.Metadata.Endpoint)
	switch strings.TrimSpace(actualInfo.Target) {
	case "node", "sandbox":
		route.Status = agentxbrowserruntime.BrowserDoctorStatusOK
		route.Code = "managed_default_route"
		route.Summary = browserRuntimeDoctorManagedDefaultSummary(strings.TrimSpace(actualInfo.Target), route.Backend, selectionReason)
	case "host":
		route.Status = agentxbrowserruntime.BrowserDoctorStatusWarn
		if managedConfigured {
			route.Code = "managed_route_not_default"
			route.Summary = firstNonEmpty(selectionReason, "Managed browser route is configured, but browser workbench still resolves onto the legacy host path.")
		} else {
			route.Code = "host_only_default_route"
			route.Summary = "Browser node route is not configured; browser workbench stays on the legacy host path until an external browser proxy or managed agentx-browserd route is available."
		}
	default:
		if managedConfigured {
			route.Status = agentxbrowserruntime.BrowserDoctorStatusWarn
			route.Code = "managed_route_hidden_by_legacy_host_default"
			route.Summary = firstNonEmpty(selectionReason, "Managed browser route is configured, but the default route is still hidden behind the implicit legacy host fallback.")
		} else {
			route.Status = agentxbrowserruntime.BrowserDoctorStatusWarn
			route.Code = "default_route_not_visible"
			route.Summary = "Browser workbench is still using the implicit legacy host fallback; configure a managed route before expecting browserd-first startup."
		}
	}
	route.Note = selectionReason
	return route
}

func browserDoctorRouteMetadataForBackend(backend BrowserBackend) browserDoctorRouteMetadata {
	if provider, ok := backend.(BrowserDoctorRouteMetadataProvider); ok {
		metadata := provider.BrowserDoctorRouteMetadata()
		return browserDoctorRouteMetadata{
			Source:   strings.TrimSpace(metadata.Source),
			Endpoint: strings.TrimSpace(metadata.Endpoint),
		}
	}
	provider, ok := backend.(browserDoctorRouteMetadataProvider)
	if !ok {
		return browserDoctorRouteMetadata{}
	}
	return provider.browserDoctorRouteMetadata()
}

func browserRuntimeDoctorRouteMetadataForAssessment(
	assessment browserConcreteRouteAssessment,
	backend BrowserBackend,
) browserDoctorRouteMetadata {
	if assessment.RouteAvailable {
		if metadata := browserDoctorRouteMetadataForBackend(assessment.Route.Backend); metadata != (browserDoctorRouteMetadata{}) {
			return metadata
		}
	}
	return browserDoctorRouteMetadataForBackend(backend)
}

func browserRuntimeDoctorVisibleDefaultRouteProjection(
	preview browserRuntimeDiagnosticsPreview,
	actualInfo BrowserRuntimeInfo,
) browserRuntimeDoctorRouteProjection {
	actualInfo = normalizeBrowserRuntimeInfo(actualInfo)
	assessment := browserRuntimeDefaultSubstrateRouteAssessment(preview.DefaultRoute, preview.Registration.SubstrateAssessment)
	switch strings.TrimSpace(actualInfo.Target) {
	case "host":
		assessment = preview.Registration.SubstrateAssessment.HostRoute
	case "node":
		assessment = browserConcreteRouteAssessmentForDefaultPromotion(preview.Registration.SubstrateAssessment.NodeRoute)
	case "sandbox":
		assessment = preview.Registration.SubstrateAssessment.SandboxConcreteRoute
	}
	assessment = browserRuntimeSubstrateRouteAssessmentForBackend(preview.Registration.EffectiveBackend, actualInfo, assessment)
	metadata := browserRuntimeDoctorRouteMetadataForAssessment(assessment, nil)
	if metadata == (browserDoctorRouteMetadata{}) {
		switch strings.TrimSpace(actualInfo.Target) {
		case "node", "sandbox":
			metadata = browserDoctorRouteMetadataForBackend(
				browserRuntimePreviewManagedTargetBackend(preview, actualInfo.Target),
			)
		}
	}
	if metadata == (browserDoctorRouteMetadata{}) &&
		BrowserSubstratePosture(actualInfo.Backend, actualInfo.Target) != BrowserSubstrateLegacySystemHost {
		metadata = browserDoctorRouteMetadataForBackend(preview.Registration.EffectiveBackend)
	}
	return browserRuntimeDoctorRouteProjection{
		Info:     actualInfo,
		Metadata: metadata,
	}
}

func browserRuntimeDoctorManagedCandidateRouteProjection(
	preview browserRuntimeDiagnosticsPreview,
) (browserRuntimeDoctorRouteProjection, bool) {
	for _, target := range []string{"node", "sandbox"} {
		projection, ok := browserRuntimeDoctorManagedTargetRouteProjection(preview, target)
		if ok {
			return projection, true
		}
	}
	return browserRuntimeDoctorRouteProjection{}, false
}

func browserRuntimeDoctorManagedTargetRouteProjection(
	preview browserRuntimeDiagnosticsPreview,
	target string,
) (browserRuntimeDoctorRouteProjection, bool) {
	target = strings.ToLower(strings.TrimSpace(target))
	if !browserRuntimePreviewConfiguresManagedTarget(preview, target) {
		return browserRuntimeDoctorRouteProjection{}, false
	}
	var (
		fallback   BrowserRuntimeInfo
		assessment browserConcreteRouteAssessment
	)
	switch target {
	case "node":
		fallback = defaultBrowserNodeRuntimeInfo()
		assessment = browserConcreteRouteAssessmentForDefaultPromotion(preview.Registration.SubstrateAssessment.NodeRoute)
	case "sandbox":
		fallback = defaultBrowserSandboxRuntimeInfo()
		assessment = preview.Registration.SubstrateAssessment.SandboxConcreteRoute
	default:
		return browserRuntimeDoctorRouteProjection{}, false
	}
	backend := browserRuntimePreviewManagedTargetBackend(preview, target)
	return browserRuntimeDoctorRouteProjection{
		Info: firstBrowserRuntimeInfo(
			browserRuntimePreviewFallbackInfoForManagedTarget(preview, target, fallback),
			normalizeBrowserRuntimeInfo(assessment.Route.RuntimeInfo),
			browserRuntimeInfoForConcreteBackend(backend, fallback),
			fallback,
		),
		Metadata: browserRuntimeDoctorRouteMetadataForAssessment(assessment, backend),
	}, true
}

func browserRuntimeDoctorFallbackRouteSource(
	actualInfo BrowserRuntimeInfo,
	displayInfo BrowserRuntimeInfo,
	managedConfigured bool,
	selectionStrategy string,
) string {
	actualInfo = normalizeBrowserRuntimeInfo(actualInfo)
	displayInfo = normalizeBrowserRuntimeInfo(displayInfo)
	switch {
	case strings.EqualFold(strings.TrimSpace(actualInfo.Target), "host") && !managedConfigured:
		return "legacy_host"
	case strings.EqualFold(strings.TrimSpace(displayInfo.Target), "host"):
		return "legacy_host"
	default:
		return strings.TrimSpace(selectionStrategy)
	}
}

func browserRuntimeDoctorLaunchSummary(ctx browserRegistrationContext, summary *browserRuntimeLaunchDiagnosticsSummary) *BrowserDoctorLaunchSummary {
	hints := browserRuntimeActionHintsForRegistration(ctx)
	if summary == nil {
		actionSummary := browserRuntimeActionHintsSummary(hints.InspectCommand, hints.DoctorCommand, hints.ReadyCommand)
		if strings.TrimSpace(actionSummary) == "" {
			actionSummary = "browser action=inspect, browser action=doctor, or browser action=ready"
		}
		return &BrowserDoctorLaunchSummary{
			BrowserDoctorCheckSummary: agentxbrowserruntime.BrowserDoctorCheckSummary{
				Status:  agentxbrowserruntime.BrowserDoctorStatusPending,
				Code:    "launch_unconfirmed",
				Summary: "Launch readiness is not confirmed yet; run " + actionSummary + " to materialize launch diagnostics.",
			},
		}
	}
	launch := &BrowserDoctorLaunchSummary{
		Source:                              strings.TrimSpace(summary.Source),
		Backend:                             strings.TrimSpace(summary.Backend),
		Profile:                             strings.TrimSpace(summary.Profile),
		RuntimeTarget:                       strings.TrimSpace(summary.RuntimeTarget),
		PlaywrightCachePath:                 strings.TrimSpace(summary.PlaywrightCachePath),
		PlaywrightCacheSource:               strings.TrimSpace(summary.PlaywrightCacheSource),
		NodeVersion:                         strings.TrimSpace(summary.NodeVersion),
		PlaywrightPackage:                   strings.TrimSpace(summary.PlaywrightPackage),
		PlaywrightPackageVersion:            strings.TrimSpace(summary.PlaywrightPackageVersion),
		RuntimeBaselineReady:                browserBoolPtrValue(summary.RuntimeBaselineReady),
		RuntimeBaselineBlockReason:          strings.TrimSpace(summary.RuntimeBaselineBlockReason),
		SelectedLaunchSource:                strings.TrimSpace(summary.SelectedLaunchSource),
		SelectedLaunchDeliveryGeneration:    strings.TrimSpace(summary.SelectedLaunchDeliveryGeneration),
		SelectedLaunchPayloadSource:         strings.TrimSpace(summary.SelectedLaunchPayloadSource),
		SelectedLaunchPayloadReady:          browserBoolPtrValue(summary.SelectedLaunchPayloadReady),
		SelectedLaunchPayloadBlockReason:    strings.TrimSpace(summary.SelectedLaunchPayloadBlockReason),
		SelectedLaunchReady:                 browserBoolPtrValue(summary.SelectedLaunchReady),
		SelectedLaunchBlockReason:           strings.TrimSpace(summary.SelectedLaunchBlockReason),
		SelectedLaunchExecutableReady:       browserBoolPtrValue(summary.SelectedLaunchExecutableReady),
		SelectedLaunchExecutableBlockReason: strings.TrimSpace(summary.SelectedLaunchExecutableBlockReason),
		DeliveryTransitionPending:           browserBoolPtrValue(summary.DeliveryTransitionPending),
		DeliveryTransitionStage:             strings.TrimSpace(summary.DeliveryTransitionStage),
		BundleReady:                         browserBoolPtrValue(summary.BundleReady),
		DeliveryReady:                       browserBoolPtrValue(summary.DeliveryReady),
		NodeModulesReady:                    browserBoolPtrValue(summary.NodeModulesReady),
		BrowserReady:                        browserBoolPtrValue(summary.BrowserReady),
		BootstrapState:                      strings.TrimSpace(summary.BootstrapState),
		BootstrapErrorCode:                  strings.TrimSpace(summary.BootstrapErrorCode),
		LaunchBlockReason:                   strings.TrimSpace(summary.LaunchBlockReason),
	}
	switch {
	case summary.LaunchReady != nil && *summary.LaunchReady:
		launch.Status = agentxbrowserruntime.BrowserDoctorStatusOK
		launch.Code = "launch_ready"
		launch.Summary = browserRuntimeDoctorLaunchReadySummary(summary)
	case summary.LaunchReady != nil && !*summary.LaunchReady:
		launch.Status = agentxbrowserruntime.BrowserDoctorStatusError
		launch.Code = firstNonEmpty(strings.TrimSpace(summary.BootstrapErrorCode), strings.TrimSpace(summary.LaunchBlockReason), "launch_blocked")
		launch.Summary = browserRuntimeDoctorLaunchBlockedSummary(summary)
	default:
		launch.Status = agentxbrowserruntime.BrowserDoctorStatusPending
		launch.Code = "launch_pending"
		launch.Summary = "Launch diagnostics are present but readiness has not been confirmed yet."
	}
	launch.Note = strings.TrimSpace(summary.Note)
	return launch
}

func browserRuntimeDoctorManagedDefaultSummary(target string, backend string, selectionReason string) string {
	target = strings.TrimSpace(target)
	backend = strings.TrimSpace(backend)
	if backend != "" && target != "" {
		if selectionReason != "" {
			return fmt.Sprintf("Managed %s route (%s) is the active default. %s", target, backend, selectionReason)
		}
		return fmt.Sprintf("Managed %s route (%s) is the active default.", target, backend)
	}
	if selectionReason != "" {
		return selectionReason
	}
	return "Managed browser route is the active default."
}

func browserRuntimeDoctorLaunchReadySummary(summary *browserRuntimeLaunchDiagnosticsSummary) string {
	if summary == nil {
		return "Launch is ready."
	}
	if note := strings.TrimSpace(summary.Note); note != "" {
		return note
	}
	source := strings.TrimSpace(summary.SelectedLaunchSource)
	if source == "" {
		source = strings.TrimSpace(summary.Source)
	}
	if source != "" {
		return fmt.Sprintf("Launch is ready via %s.", source)
	}
	return "Launch is ready."
}

func browserBoolPtrValue(v *bool) *bool {
	if v == nil {
		return nil
	}
	out := *v
	return &out
}

func browserRuntimeDoctorLaunchBlockedSummary(summary *browserRuntimeLaunchDiagnosticsSummary) string {
	if summary == nil {
		return "Launch is blocked."
	}
	if bootstrap := browserRuntimeDoctorLaunchBootstrapBlockedSummary(summary); bootstrap != "" {
		return bootstrap
	}
	if baseline := browserRuntimeDoctorLaunchBaselineBlockedSummary(summary); baseline != "" {
		return baseline
	}
	if selected := browserRuntimeDoctorSelectedLaunchBlockedSummary(summary); selected != "" {
		return selected
	}
	if note := strings.TrimSpace(summary.Note); note != "" {
		return note
	}
	if reason := strings.TrimSpace(summary.LaunchBlockReason); reason != "" {
		return "Launch is blocked: " + reason
	}
	if code := strings.TrimSpace(summary.BootstrapErrorCode); code != "" {
		return "Bootstrap is blocked: " + code
	}
	if state := strings.TrimSpace(summary.BootstrapState); state != "" {
		return "Bootstrap state is " + state
	}
	return "Launch is blocked."
}

func browserRuntimeDoctorLaunchBootstrapBlockedSummary(summary *browserRuntimeLaunchDiagnosticsSummary) string {
	if summary == nil {
		return ""
	}
	code := strings.TrimSpace(summary.BootstrapErrorCode)
	if code == "" {
		return ""
	}
	context := browserRuntimeDoctorLaunchBootstrapContextSummary(summary)
	switch code {
	case "node_missing":
		return "Bundled browserd bootstrap is blocked because `node` is not available in PATH."
	case "npm_missing":
		return "Bundled browserd bootstrap is blocked because `npm` is not available in PATH."
	case "npm_ci_failed":
		return browserRuntimeDoctorLaunchAppendContext("Bundled browserd bootstrap is blocked because `npm ci` did not complete successfully.", context)
	case "playwright_dependency_missing":
		return browserRuntimeDoctorLaunchAppendContext("Bundled browserd bootstrap is blocked because Playwright dependencies are incomplete.", context)
	case "playwright_cli_missing":
		return browserRuntimeDoctorLaunchAppendContext("Bundled browserd bootstrap is blocked because the Playwright CLI is not available.", context)
	case "browser_install_timeout":
		return browserRuntimeDoctorLaunchAppendContext("Bundled browserd bootstrap is blocked because the Playwright browser download timed out.", context)
	case "browser_install_network_failed":
		return browserRuntimeDoctorLaunchAppendContext("Bundled browserd bootstrap is blocked because the Playwright browser download failed due to a network or TLS problem.", context)
	case "browser_install_failed":
		return browserRuntimeDoctorLaunchAppendContext("Bundled browserd bootstrap is blocked because Playwright browser installation failed.", context)
	case "browser_executable_missing":
		return browserRuntimeDoctorLaunchAppendContext("Bundled browserd bootstrap is blocked because no ready browser executable is available yet.", context)
	default:
		return ""
	}
}

func browserRuntimeDoctorLaunchAppendContext(base string, context string) string {
	base = strings.TrimSpace(base)
	context = strings.TrimSpace(context)
	if base == "" {
		return context
	}
	if context == "" {
		return base
	}
	return base + " " + context
}

func browserRuntimeDoctorLaunchBootstrapContextSummary(summary *browserRuntimeLaunchDiagnosticsSummary) string {
	if summary == nil {
		return ""
	}
	parts := make([]string, 0, 2)
	if summary.SelectedLaunchPayloadReady != nil && !*summary.SelectedLaunchPayloadReady {
		reason := strings.TrimSpace(summary.SelectedLaunchPayloadBlockReason)
		source := strings.TrimSpace(summary.SelectedLaunchPayloadSource)
		switch {
		case reason != "" && source != "":
			parts = append(parts, fmt.Sprintf("Selected Playwright payload (%s) is not ready: %s.", source, reason))
		case reason != "":
			parts = append(parts, "Selected Playwright payload is not ready: "+reason+".")
		case source != "":
			parts = append(parts, fmt.Sprintf("Selected Playwright payload (%s) is not ready.", source))
		}
	}
	if summary.DeliveryTransitionPending != nil && *summary.DeliveryTransitionPending {
		stage := strings.TrimSpace(summary.DeliveryTransitionStage)
		if stage != "" {
			parts = append(parts, "Playwright delivery transition is pending at "+stage+".")
		} else {
			parts = append(parts, "Playwright delivery transition is still pending.")
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " ")
}

func browserRuntimeDoctorLaunchBaselineBlockedSummary(summary *browserRuntimeLaunchDiagnosticsSummary) string {
	if summary == nil || summary.RuntimeBaselineReady == nil || *summary.RuntimeBaselineReady {
		return ""
	}
	if reason := strings.TrimSpace(summary.RuntimeBaselineBlockReason); reason != "" {
		return "Runtime baseline is not ready: " + reason
	}
	return "Runtime baseline is not ready."
}

func browserRuntimeDoctorSelectedLaunchBlockedSummary(summary *browserRuntimeLaunchDiagnosticsSummary) string {
	if summary == nil {
		return ""
	}
	if summary.SelectedLaunchExecutableReady != nil && !*summary.SelectedLaunchExecutableReady {
		if reason := strings.TrimSpace(summary.SelectedLaunchExecutableBlockReason); reason != "" {
			return "Selected launch executable is not ready: " + reason
		}
		return "Selected launch executable is not ready."
	}
	if summary.SelectedLaunchReady != nil && !*summary.SelectedLaunchReady {
		if reason := strings.TrimSpace(summary.SelectedLaunchBlockReason); reason != "" {
			return "Selected launch target is not ready: " + reason
		}
		return "Selected launch target is not ready."
	}
	return ""
}

func browserRuntimeDoctorSuggestions(ctx browserRegistrationContext, doctor *BrowserDoctorSummary) []string {
	if doctor == nil {
		return nil
	}
	hints := browserRuntimeActionHintsForRegistration(ctx)
	suggestions := make([]string, 0, 4)
	if doctor.Route != nil {
		switch strings.TrimSpace(doctor.Route.Code) {
		case "host_only_default_route", "default_route_not_visible":
			suggestions = append(suggestions, "configure a managed browser route (browserDaemonCommand + browserDaemonPort, or browserProxyEndpoint + browserProxyToken)")
		case "managed_route_not_default", "managed_route_hidden_by_legacy_host_default":
			actionSummary := browserRuntimeActionHintsSummary("", hints.DoctorCommand, hints.ReadyCommand)
			if strings.TrimSpace(actionSummary) == "" {
				actionSummary = "browser action=doctor or browser action=ready"
			}
			suggestions = append(suggestions, "enable the managed-first selector `group:browser-default-selection:prefer_node_over_legacy_host` (or an equivalent route policy), then use "+actionSummary+"; use runtime_target=node explicitly until the default route is promoted")
		}
	}
	if doctor.Launch != nil {
		switch strings.TrimSpace(doctor.Launch.Status) {
		case agentxbrowserruntime.BrowserDoctorStatusPending:
			actionSummary := browserRuntimeActionHintsSummary("", hints.DoctorCommand, hints.ReadyCommand)
			if strings.TrimSpace(actionSummary) == "" {
				actionSummary = "browser action=doctor or browser action=ready"
			}
			suggestions = append(suggestions, "run "+actionSummary+" to confirm launch readiness")
		case agentxbrowserruntime.BrowserDoctorStatusError:
			if cmd := strings.TrimSpace(hints.RepairCommand); cmd != "" && browserRuntimeDoctorLaunchShouldSuggestRepairCommand(doctor.Launch) {
				suggestions = append(suggestions, "run "+cmd)
			}
			if cmd := strings.TrimSpace(doctor.RepairCommand); cmd != "" && browserRuntimeDoctorLaunchShouldSuggestRepairCommand(doctor.Launch) {
				suggestions = append(suggestions, "run "+cmd)
			}
			if cmd := strings.TrimSpace(doctor.AcceptanceCommand); cmd != "" {
				suggestions = append(suggestions, "run "+cmd)
			}
			suggestions = append(suggestions, browserRuntimeDoctorLaunchSuggestions(doctor.Launch)...)
		}
	}
	return browserRuntimeDoctorUniqueSuggestions(suggestions)
}

func browserRuntimeDoctorBringupSummary(
	ctx browserRegistrationContext,
	doctor *BrowserDoctorSummary,
) *BrowserDoctorBringupSummary {
	if doctor == nil {
		return nil
	}
	hints := browserRuntimeActionHintsForRegistration(ctx)
	preferredRuntime := browserRuntimeDoctorPreferredBringupRuntime(doctor)
	repairAction := strings.TrimSpace(hints.RepairCommand)
	routeSource := "runtime_selection"
	routeEndpoint := ""
	if doctor.Route != nil {
		routeSource = firstNonEmpty(strings.TrimSpace(doctor.Route.Source), routeSource)
		routeEndpoint = strings.TrimSpace(doctor.Route.Endpoint)
	}
	bringup := &BrowserDoctorBringupSummary{
		SubstrateSource:         routeSource,
		SubstrateEndpoint:       routeEndpoint,
		PreferredRuntimeBackend: strings.TrimSpace(preferredRuntime.Backend),
		PreferredRuntimeProfile: strings.TrimSpace(preferredRuntime.Profile),
		PreferredRuntimeTarget:  strings.TrimSpace(preferredRuntime.Target),
		PrimaryStep:             browserRuntimeDoctorPrimaryBringupStep(doctor, hints),
		DoctorAction:            strings.TrimSpace(hints.DoctorCommand),
		RepairAction:            repairAction,
		ReadyAction:             strings.TrimSpace(hints.ReadyCommand),
		AcceptanceCommand:       strings.TrimSpace(doctor.AcceptanceCommand),
		Steps: agentxbrowserruntime.BrowserDoctorBringupSteps(
			hints.DoctorCommand,
			repairAction,
			hints.ReadyCommand,
			doctor.AcceptanceCommand,
		),
	}
	if strings.TrimSpace(bringup.PrimaryStep) == "" &&
		strings.TrimSpace(bringup.DoctorAction) == "" &&
		strings.TrimSpace(bringup.RepairAction) == "" &&
		strings.TrimSpace(bringup.ReadyAction) == "" &&
		strings.TrimSpace(bringup.AcceptanceCommand) == "" &&
		len(bringup.Steps) == 0 {
		return nil
	}
	return bringup
}

func browserRuntimeDoctorPreferredBringupRuntime(doctor *BrowserDoctorSummary) BrowserRuntimeInfo {
	if doctor == nil {
		return BrowserRuntimeInfo{}
	}
	if doctor.Route != nil {
		info := BrowserRuntimeInfo{
			Backend: strings.TrimSpace(doctor.Route.Backend),
			Profile: strings.TrimSpace(doctor.Route.Profile),
			Target:  strings.TrimSpace(doctor.Route.RuntimeTarget),
		}
		if strings.TrimSpace(info.Backend) != "" ||
			strings.TrimSpace(info.Profile) != "" ||
			strings.TrimSpace(info.Target) != "" {
			return info
		}
	}
	for _, target := range doctor.ConfiguredTargets {
		target = strings.TrimSpace(target)
		if target != "" && target != "host" {
			return BrowserRuntimeInfo{Target: target}
		}
	}
	if len(doctor.ConfiguredTargets) > 0 {
		return BrowserRuntimeInfo{Target: strings.TrimSpace(doctor.ConfiguredTargets[0])}
	}
	return BrowserRuntimeInfo{}
}

func browserRuntimeDoctorPrimaryBringupStep(
	doctor *BrowserDoctorSummary,
	hints browserRuntimeActionHints,
) string {
	if doctor == nil {
		return ""
	}
	doctorAction := strings.TrimSpace(hints.DoctorCommand)
	readyAction := strings.TrimSpace(hints.ReadyCommand)
	repairAction := strings.TrimSpace(hints.RepairCommand)
	acceptanceCommand := strings.TrimSpace(doctor.AcceptanceCommand)
	if doctor.Ready {
		return firstNonEmpty(acceptanceCommand, readyAction, doctorAction)
	}
	if doctor.Launch != nil {
		switch strings.TrimSpace(doctor.Launch.Status) {
		case agentxbrowserruntime.BrowserDoctorStatusError:
			if browserRuntimeDoctorLaunchShouldSuggestRepairCommand(doctor.Launch) {
				return firstNonEmpty(repairAction, readyAction, doctorAction, acceptanceCommand)
			}
			return firstNonEmpty(readyAction, doctorAction, acceptanceCommand)
		case agentxbrowserruntime.BrowserDoctorStatusPending:
			return firstNonEmpty(readyAction, doctorAction, acceptanceCommand)
		}
	}
	if doctor.Route != nil {
		switch strings.TrimSpace(doctor.Route.Code) {
		case "managed_route_not_default", "managed_route_hidden_by_legacy_host_default":
			return firstNonEmpty(readyAction, doctorAction, acceptanceCommand)
		}
	}
	return firstNonEmpty(doctorAction, readyAction, acceptanceCommand)
}

func browserRuntimeDoctorLaunchShouldSuggestRepairCommand(launch *BrowserDoctorLaunchSummary) bool {
	if launch == nil {
		return false
	}
	return browserRuntimeBootstrapErrorCodeSupportsRepair(launch.BootstrapErrorCode)
}

func browserRuntimeDoctorLaunchSuggestions(launch *BrowserDoctorLaunchSummary) []string {
	if launch == nil {
		return nil
	}
	suggestions := []string{}
	switch strings.TrimSpace(launch.BootstrapErrorCode) {
	case "node_missing":
		suggestions = append(suggestions, "install Node.js and ensure `node` is available in PATH for bundled agentx-browserd")
	case "npm_missing":
		suggestions = append(suggestions, "install npm and ensure `npm` is available in PATH for bundled agentx-browserd")
	case "npm_ci_failed":
		suggestions = append(suggestions, "repair bundled browserd dependencies so `npm ci` succeeds before retrying browser bring-up")
	case "playwright_dependency_missing", "playwright_cli_missing":
		suggestions = append(suggestions, "repair the host-provided Playwright installation before retrying browser bring-up")
	case "browser_install_timeout":
		suggestions = append(suggestions, "retry browser repair after checking reachability to the Playwright download hosts; slow or unstable networks can stall Chromium bootstrap")
	case "browser_install_network_failed":
		suggestions = append(suggestions, "check proxy, TLS, and outbound network access to the Playwright download hosts, then rerun browser repair")
	case "browser_install_failed":
		suggestions = append(suggestions, "reinstall Playwright browser binaries before retrying browser bring-up")
	case "browser_executable_missing":
		suggestions = append(suggestions, "materialize a ready Playwright browser executable before retrying browser bring-up")
	}
	suggestions = append(suggestions, browserRuntimeDoctorLaunchContextSuggestions(launch)...)
	if launch.RuntimeBaselineReady != nil && !*launch.RuntimeBaselineReady && strings.TrimSpace(launch.RuntimeBaselineBlockReason) != "" {
		suggestions = append(suggestions, "fix the runtime baseline blocker: "+strings.TrimSpace(launch.RuntimeBaselineBlockReason))
	}
	if launch.SelectedLaunchExecutableReady != nil && !*launch.SelectedLaunchExecutableReady && strings.TrimSpace(launch.SelectedLaunchExecutableBlockReason) != "" {
		suggestions = append(suggestions, "fix the selected launch executable blocker: "+strings.TrimSpace(launch.SelectedLaunchExecutableBlockReason))
	}
	if launch.SelectedLaunchReady != nil && !*launch.SelectedLaunchReady && strings.TrimSpace(launch.SelectedLaunchBlockReason) != "" {
		suggestions = append(suggestions, "fix the selected launch blocker: "+strings.TrimSpace(launch.SelectedLaunchBlockReason))
	}
	return suggestions
}

func browserRuntimeDoctorLaunchContextSuggestions(launch *BrowserDoctorLaunchSummary) []string {
	if launch == nil {
		return nil
	}
	suggestions := []string{}
	if launch.SelectedLaunchPayloadReady != nil && !*launch.SelectedLaunchPayloadReady {
		if reason := strings.TrimSpace(launch.SelectedLaunchPayloadBlockReason); reason != "" {
			suggestions = append(suggestions, "fix the selected launch payload blocker: "+reason)
		} else {
			suggestions = append(suggestions, "fix the selected launch payload blocker before retrying browser bring-up")
		}
	}
	if launch.DeliveryTransitionPending != nil && *launch.DeliveryTransitionPending {
		switch stage := strings.TrimSpace(launch.DeliveryTransitionStage); stage {
		case "dependencies_not_ready":
			suggestions = append(suggestions, "repair bundled browserd dependencies so the Playwright delivery transition can complete")
		case "browser_not_ready":
			suggestions = append(suggestions, "repair or reinstall Playwright browser binaries so the delivery transition can complete")
		case "delivery_not_ready":
			suggestions = append(suggestions, "wait for or repair the Playwright delivery transition before retrying browser bring-up")
		case "":
			suggestions = append(suggestions, "wait for or repair the Playwright delivery transition before retrying browser bring-up")
		default:
			suggestions = append(suggestions, "wait for or repair the Playwright delivery transition stage: "+stage)
		}
	}
	if cacheHint := browserRuntimeDoctorLaunchCacheSuggestion(launch); cacheHint != "" {
		switch strings.TrimSpace(launch.BootstrapErrorCode) {
		case "browser_install_failed", "browser_install_timeout", "browser_install_network_failed", "browser_executable_missing", "playwright_dependency_missing":
			suggestions = append(suggestions, cacheHint)
		}
	}
	return suggestions
}

func browserRuntimeDoctorLaunchCacheSuggestion(launch *BrowserDoctorLaunchSummary) string {
	if launch == nil {
		return ""
	}
	cachePath := strings.TrimSpace(launch.PlaywrightCachePath)
	cacheSource := strings.TrimSpace(launch.PlaywrightCacheSource)
	switch {
	case cachePath != "" && cacheSource != "":
		return fmt.Sprintf("inspect the Playwright cache at %s (source=%s) and rerun browser repair if Chromium payloads are incomplete", cachePath, cacheSource)
	case cachePath != "":
		return fmt.Sprintf("inspect the Playwright cache at %s and rerun browser repair if Chromium payloads are incomplete", cachePath)
	case cacheSource != "":
		return fmt.Sprintf("inspect the Playwright cache (source=%s) and rerun browser repair if Chromium payloads are incomplete", cacheSource)
	default:
		return ""
	}
}

func browserRuntimeDoctorRouteStatus(route *BrowserDoctorRouteSummary) string {
	if route == nil {
		return ""
	}
	return strings.TrimSpace(route.Status)
}

func browserRuntimeDoctorLaunchStatus(launch *BrowserDoctorLaunchSummary) string {
	if launch == nil {
		return ""
	}
	return strings.TrimSpace(launch.Status)
}

func browserRuntimeDoctorUniqueSuggestions(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		out = append(out, item)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
