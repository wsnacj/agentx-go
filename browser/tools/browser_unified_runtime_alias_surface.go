package tools

import (
	"encoding/json"
	"strings"
)

func browserUnifiedUsesDoctorRouteInspectionSummary(action string) bool {
	switch strings.ToLower(strings.TrimSpace(action)) {
	case "inspect", "doctor", "status", "ready", "prepare", "inventory", "profiles", "handles", "sessions", "ensure", "coordinate", "launch", "start":
		return true
	default:
		return false
	}
}

func browserUnifiedDoctorRouteInspectionRuntimeAction(action string) string {
	action = strings.ToLower(strings.TrimSpace(action))
	if alias, ok := browserUnifiedRuntimeActionAliases[action]; ok {
		return strings.TrimSpace(alias.Action)
	}
	return action
}

func browserUnifiedShouldApplyDoctorRouteInspectionSummary(
	payload *browserRuntimePayload,
	action string,
) (*BrowserDoctorRouteSummary, bool) {
	if payload == nil ||
		payload.Doctor == nil ||
		payload.Doctor.Route == nil ||
		strings.TrimSpace(payload.RequestedRuntimeTarget) != "" ||
		strings.TrimSpace(payload.RequestedProfile) != "" ||
		!browserUnifiedUsesDoctorRouteInspectionSummary(action) {
		return nil, false
	}
	switch strings.TrimSpace(payload.Status) {
	case "ok", "unsupported":
	default:
		return nil, false
	}
	if browserRuntimeHasDoctorRouteInspectionSummary(payload.Diagnostics) ||
		browserRuntimeHasDoctorRouteInspectionSummary(payload.Summary) ||
		browserRuntimeHasDoctorRouteInspectionDisplay(payload.Display) {
		return nil, false
	}
	route := payload.Doctor.Route
	_, ok := browserRuntimeDoctorRouteInspectionSummaryBase(route, "ready")
	if !ok {
		return nil, false
	}
	return route, true
}

func browserUnifiedMaybeApplyDoctorRouteInspectionSummary(
	ctx browserRegistrationContext,
	payload *browserRuntimePayload,
	action string,
) bool {
	route, ok := browserUnifiedShouldApplyDoctorRouteInspectionSummary(payload, action)
	if !ok {
		return false
	}
	runtimeAction := browserUnifiedDoctorRouteInspectionRuntimeAction(action)
	preview := browserRuntimeDiagnosticsPreviewBaseForRegistration(ctx)
	if !browserRuntimeShouldPromoteDoctorRouteInspectionSummary(
		ctx,
		payload,
		runtimeAction,
		preview,
		route,
	) {
		return false
	}
	browserRuntimeApplyDoctorRouteInspectionSummaryPayload(
		ctx,
		payload,
		runtimeAction,
		preview,
		route,
		browserRuntimeActionHintsForRegistration(ctx),
		false,
	)
	return true
}

func browserUnifiedApplyRuntimeAliasSurface(
	ctx browserRegistrationContext,
	action string,
	raw string,
) (string, error) {
	if !browserUnifiedUsesDoctorRouteInspectionSummary(action) {
		return raw, nil
	}
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || !strings.HasPrefix(trimmed, "{") {
		return raw, nil
	}
	var payload browserRuntimePayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return raw, nil
	}
	if !browserUnifiedMaybeApplyDoctorRouteInspectionSummary(ctx, &payload, action) {
		return raw, nil
	}
	browserRuntimeApplyToolAwareActionCommands(ctx, &payload)
	blob, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(blob), nil
}
