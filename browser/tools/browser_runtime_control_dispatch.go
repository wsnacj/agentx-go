package tools

import (
	"context"
	"strings"
)

type browserRuntimeActionDispatchControlPlane struct {
	DefaultCandidateRoute      BrowserRuntimeInfo
	DefaultCandidateDescriptor browserRuntimeRouteDescriptor
	ConfiguredInfo             BrowserRuntimeInfo
	SelectedRoute              browserResolvedExecutionRoute
	SelectedInfo               BrowserRuntimeInfo
	SelectedBackend            BrowserBackend
	Capabilities               BrowserCapabilities
	RouteErr                   error
	Handled                    bool
}

func browserRuntimeHiddenImplicitHostDiagnosticsDegradeNote(preview browserRuntimeDiagnosticsPreview) string {
	route := browserRuntimeDoctorRouteSummary(&browserRuntimePayload{
		DefaultRoute:               preview.DefaultRouteDescriptor,
		ConfiguredTargets:          append([]string(nil), preview.ConfiguredTargets...),
		SubstrateSelectionStrategy: strings.TrimSpace(preview.Registration.SubstrateSummary.SelectionStrategy),
		SubstrateSelectionReason:   strings.TrimSpace(preview.Registration.SubstrateSummary.SelectionReason),
	}, preview)
	if route == nil {
		return ""
	}
	if _, ok := browserRuntimeDoctorRouteInspectionSummaryBase(route, "ready"); ok {
		return strings.TrimSpace(browserRuntimeDoctorRouteInspectionCanonicalSummary(route))
	}
	return ""
}

func browserRuntimePrepareActionDispatchControlPlane(
	ctx browserRegistrationContext,
	callCtx context.Context,
	payload *browserRuntimePayload,
	action string,
	defaultRoute BrowserRuntimeInfo,
	substrateAssessment browserDefaultSubstrateAssessment,
	selection browserRuntimeActionDispatchSelection,
) browserRuntimeActionDispatchControlPlane {
	return browserRuntimePrepareActionDispatchControlPlaneWithPreview(
		ctx,
		callCtx,
		payload,
		action,
		defaultRoute,
		substrateAssessment,
		browserRuntimeDiagnosticsPreviewBaseForRegistration(ctx),
		selection,
	)
}

func browserRuntimePrepareActionDispatchControlPlaneWithPreview(
	ctx browserRegistrationContext,
	callCtx context.Context,
	payload *browserRuntimePayload,
	action string,
	defaultRoute BrowserRuntimeInfo,
	substrateAssessment browserDefaultSubstrateAssessment,
	diagnosticsPreview browserRuntimeDiagnosticsPreview,
	selection browserRuntimeActionDispatchSelection,
) browserRuntimeActionDispatchControlPlane {
	prepared := browserRuntimeActionDispatchControlPlane{
		DefaultCandidateRoute:      normalizeBrowserRuntimeInfo(selection.DefaultCandidateRoute),
		DefaultCandidateDescriptor: selection.DefaultCandidateDescriptor,
		ConfiguredInfo:             defaultRoute,
		RouteErr:                   selection.RouteErr,
	}
	if payload == nil {
		return prepared
	}
	browserRuntimeRefreshSubstrateContextWithPreview(ctx, payload, diagnosticsPreview)
	browserRuntimeApplySelectionDefaultCandidateRoute(payload, diagnosticsPreview, prepared.DefaultCandidateRoute, prepared.DefaultCandidateDescriptor)
	if selection.UseHiddenImplicitHostDiagnosticsDegrade {
		if note := browserRuntimeHiddenImplicitHostDiagnosticsDegradeNote(diagnosticsPreview); note != "" {
			browserRuntimeApplyActionTerminalStatus(payload, browserRuntimeActionTerminalStatus{
				Note:                 note,
				PreserveExistingNote: true,
			})
		}
	}
	if selection.UseHiddenImplicitHostDiagnosticsDegrade || selection.RouteErr != nil {
		if selection.RouteErr != nil {
			browserRuntimeApplyActionTerminalStatus(payload, browserRuntimeActionTerminalStatus{
				Note: selection.RouteErr.Error(),
			})
		}
		if !browserRuntimeRouteErrIsManagedLaunchFailure(selection.RouteErr) &&
			browserRuntimeCanDegradeDefaultRouteFailure(
				callCtx,
				action,
				ctx.sessionRegistry,
				ctx.sessionStateRegistry,
				defaultRoute,
				substrateAssessment,
				payload.RequestedProfile,
				payload.RequestedRuntimeTarget,
			) {
			browserRuntimeApplyDegradedActionDispatchPayloadWithPreview(
				ctx,
				callCtx,
				payload,
				action,
				defaultRoute,
				payload.RequestedProfile,
				diagnosticsPreview,
			)
			prepared.Handled = true
			return prepared
		}
		if selection.RouteErr != nil {
			browserRuntimeApplyUnsupportedRouteActionOutcome(payload, action, selection.RouteErr)
		} else {
			browserRuntimeApplyActionTerminalStatus(payload, browserRuntimeActionTerminalStatus{
				Status:               "unsupported",
				PreserveExistingNote: true,
			})
		}
		prepared.Handled = true
		return prepared
	}

	prepared.SelectedRoute = selection.SelectedRoute
	prepared.SelectedInfo = selection.SelectedRoute.RuntimeInfo
	prepared.SelectedBackend = selection.SelectedRoute.Backend
	prepared.ConfiguredInfo = prepared.SelectedInfo
	projection, capabilities := browserRuntimeSelectedRouteProjection(ctx, diagnosticsPreview, prepared.SelectedRoute)
	browserRuntimeApplyTopLevelRouteProjection(payload, projection)
	browserRuntimeApplySelectionDefaultCandidateRoute(payload, diagnosticsPreview, prepared.DefaultCandidateRoute, prepared.DefaultCandidateDescriptor)
	prepared.Capabilities = capabilities
	return prepared
}

func browserRuntimeApplySelectionDefaultCandidateRoute(
	payload *browserRuntimePayload,
	preview browserRuntimeDiagnosticsPreview,
	route BrowserRuntimeInfo,
	descriptor browserRuntimeRouteDescriptor,
) {
	if payload == nil {
		return
	}
	if payload.DefaultCandidateRoute != (browserRuntimeRouteDescriptor{}) {
		return
	}
	if descriptor != (browserRuntimeRouteDescriptor{}) {
		if route == (BrowserRuntimeInfo{}) ||
			normalizeBrowserRuntimeInfo(browserRuntimeInfoFromRouteDescriptor(&descriptor)) == normalizeBrowserRuntimeInfo(route) {
			payload.DefaultCandidateRoute = descriptor
			browserRuntimeApplyDefaultCandidateRouteToPayloadShells(payload)
			return
		}
	}
	descriptor = browserRuntimeSelectionDefaultCandidateRouteDescriptor(preview, route)
	if descriptor == (browserRuntimeRouteDescriptor{}) {
		return
	}
	payload.DefaultCandidateRoute = descriptor
	browserRuntimeApplyDefaultCandidateRouteToPayloadShells(payload)
}

func browserRuntimeSelectionDefaultCandidateRouteDescriptor(
	preview browserRuntimeDiagnosticsPreview,
	route BrowserRuntimeInfo,
) browserRuntimeRouteDescriptor {
	route = normalizeBrowserRuntimeInfo(route)
	if route == (BrowserRuntimeInfo{}) {
		return browserRuntimeRouteDescriptor{}
	}
	if candidate := browserRuntimeRouteDescriptorFromInfoWithProvenance(
		preview.Registration.SubstrateSummary.DefaultCandidateRoute,
		preview.Registration.SubstrateSummary.DefaultCandidateSource,
		preview.Registration.SubstrateSummary.DefaultCandidateEndpoint,
	); candidate != (browserRuntimeRouteDescriptor{}) &&
		normalizeBrowserRuntimeInfo(browserRuntimeInfoFromRouteDescriptor(&candidate)) == route {
		return candidate
	}
	if projection, ok := browserRuntimeDoctorManagedCandidateRouteProjection(preview); ok &&
		normalizeBrowserRuntimeInfo(projection.Info) == route {
		return browserRuntimeRouteDescriptorFromInfoWithProvenance(
			projection.Info,
			projection.Metadata.Source,
			projection.Metadata.Endpoint,
		)
	}
	return browserRuntimeRouteDescriptorFromInfo(route)
}

func browserRuntimeApplyDegradedActionDispatchPayload(
	ctx browserRegistrationContext,
	callCtx context.Context,
	payload *browserRuntimePayload,
	action string,
	defaultRoute BrowserRuntimeInfo,
	requestedProfile string,
) {
	browserRuntimeApplyDegradedActionDispatchPayloadWithPreview(
		ctx,
		callCtx,
		payload,
		action,
		defaultRoute,
		requestedProfile,
		browserRuntimeDiagnosticsPreviewBaseForRegistration(ctx),
	)
}

func browserRuntimeApplyDegradedActionDispatchPayloadWithPreview(
	ctx browserRegistrationContext,
	callCtx context.Context,
	payload *browserRuntimePayload,
	action string,
	defaultRoute BrowserRuntimeInfo,
	requestedProfile string,
	preview browserRuntimeDiagnosticsPreview,
) {
	if payload == nil {
		return
	}
	projection := browserRuntimeDegradedRouteProjectionFromSnapshot(
		callCtx,
		ctx.sessionRegistry,
		ctx.sessionRunRegistry,
		ctx.sessionStateRegistry,
		defaultRoute,
		requestedProfile,
	)
	if metadata, ok := browserRuntimeDegradedRouteSurfaceMetadataWithPreview(ctx, preview, projection.Route); ok {
		browserRuntimeApplyCapabilityMetadataToPayload(payload, metadata)
	}
	browserRuntimeApplySharedActionSurface(callCtx, payload, browserRuntimeDegradedActionSurface(action, projection))
}
