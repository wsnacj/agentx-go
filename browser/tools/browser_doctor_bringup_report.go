package tools

import (
	"strconv"
	"strings"
)

type BrowserDoctorBringupReport struct {
	SubstrateSource          string                     `json:"substrate_source,omitempty"`
	SubstrateEndpoint        string                     `json:"substrate_endpoint,omitempty"`
	SubstrateBackend         string                     `json:"substrate_backend,omitempty"`
	SubstrateProfile         string                     `json:"substrate_profile,omitempty"`
	SubstrateTarget          string                     `json:"substrate_target,omitempty"`
	SubstratePosture         string                     `json:"substrate_posture,omitempty"`
	SubstrateStatus          string                     `json:"substrate_status,omitempty"`
	PreferredRuntimeBackend  string                     `json:"preferred_runtime_backend,omitempty"`
	PreferredRuntimeProfile  string                     `json:"preferred_runtime_profile,omitempty"`
	PreferredRuntimeTarget   string                     `json:"preferred_runtime_target,omitempty"`
	DefaultCandidate         string                     `json:"default_candidate,omitempty"`
	DefaultCandidateSource   string                     `json:"default_candidate_source,omitempty"`
	DefaultCandidateEndpoint string                     `json:"default_candidate_endpoint,omitempty"`
	DefaultCandidateBackend  string                     `json:"default_candidate_backend,omitempty"`
	DefaultCandidateProfile  string                     `json:"default_candidate_profile,omitempty"`
	DefaultCandidateTarget   string                     `json:"default_candidate_target,omitempty"`
	SelectionStrategy        string                     `json:"selection_strategy,omitempty"`
	SubstrateReason          string                     `json:"substrate_reason,omitempty"`
	RepairAvailable          bool                       `json:"repair_available,omitempty"`
	PrimaryStep              string                     `json:"primary_step,omitempty"`
	DoctorAction             string                     `json:"doctor_action,omitempty"`
	RepairAction             string                     `json:"repair_action,omitempty"`
	ReadyAction              string                     `json:"ready_action,omitempty"`
	AcceptanceCommand        string                     `json:"acceptance_command,omitempty"`
	Steps                    []BrowserDoctorBringupStep `json:"steps,omitempty"`
	Summary                  string                     `json:"summary,omitempty"`
}

func BuildBrowserDoctorBringupReport(bringup *BrowserDoctorBringupSummary, hints BrowserToolMetadataRouteHints) *BrowserDoctorBringupReport {
	if bringup == nil && hints.Empty() {
		return nil
	}
	currentRuntime := normalizeBrowserRuntimeInfo(hints.RuntimeInfo)
	currentPosture := strings.TrimSpace(hints.SubstratePosture)
	currentStatus := strings.TrimSpace(hints.SubstrateStatus)
	currentReason := strings.TrimSpace(hints.SubstrateReason)
	hints = hints.Canonicalized()
	preferredRuntime := hints.PreferredRuntimeInfo()
	if currentRuntime == (BrowserRuntimeInfo{}) {
		currentRuntime = normalizeBrowserRuntimeInfo(hints.RuntimeInfo)
	}
	if currentPosture == "" {
		currentPosture = strings.TrimSpace(hints.SubstratePosture)
	}
	if currentStatus == "" {
		currentStatus = strings.TrimSpace(hints.SubstrateStatus)
	}
	if currentReason == "" {
		currentReason = strings.TrimSpace(hints.SubstrateReason)
	}
	report := &BrowserDoctorBringupReport{
		SubstrateSource:          "",
		SubstrateEndpoint:        "",
		SubstrateBackend:         strings.TrimSpace(currentRuntime.Backend),
		SubstrateProfile:         strings.TrimSpace(currentRuntime.Profile),
		SubstrateTarget:          strings.TrimSpace(currentRuntime.Target),
		SubstratePosture:         strings.TrimSpace(currentPosture),
		SubstrateStatus:          strings.TrimSpace(currentStatus),
		DefaultCandidate:         strings.TrimSpace(hints.DefaultCandidateLabel()),
		DefaultCandidateSource:   strings.TrimSpace(hints.DefaultCandidateSource),
		DefaultCandidateEndpoint: strings.TrimSpace(hints.DefaultCandidateEndpoint),
		DefaultCandidateBackend:  strings.TrimSpace(hints.DefaultCandidateRoute.Backend),
		DefaultCandidateProfile:  strings.TrimSpace(hints.DefaultCandidateRoute.Profile),
		DefaultCandidateTarget:   strings.TrimSpace(hints.DefaultCandidateRoute.Target),
		SelectionStrategy:        strings.TrimSpace(hints.SelectionStrategy),
		SubstrateReason:          strings.TrimSpace(currentReason),
		RepairAvailable:          hints.RepairAvailable,
	}
	if currentRuntime != (BrowserRuntimeInfo{}) {
		report.SubstratePosture = BrowserSubstratePosture(report.SubstrateBackend, report.SubstrateTarget)
		report.SubstrateStatus = BrowserSubstrateStatus(report.SubstrateBackend, report.SubstrateTarget)
	} else if report.SubstratePosture == "" {
		report.SubstratePosture = BrowserSubstratePosture(report.SubstrateBackend, report.SubstrateTarget)
	}
	if currentRuntime == (BrowserRuntimeInfo{}) && report.SubstrateStatus == "" {
		report.SubstrateStatus = BrowserSubstrateStatus(report.SubstrateBackend, report.SubstrateTarget)
	}
	if bringup != nil {
		report.SubstrateSource = strings.TrimSpace(bringup.SubstrateSource)
		report.SubstrateEndpoint = strings.TrimSpace(bringup.SubstrateEndpoint)
		report.PreferredRuntimeBackend = strings.TrimSpace(bringup.PreferredRuntimeBackend)
		report.PreferredRuntimeProfile = strings.TrimSpace(bringup.PreferredRuntimeProfile)
		report.PreferredRuntimeTarget = strings.TrimSpace(bringup.PreferredRuntimeTarget)
		report.PrimaryStep = strings.TrimSpace(bringup.PrimaryStep)
		report.DoctorAction = strings.TrimSpace(bringup.DoctorAction)
		report.RepairAction = strings.TrimSpace(bringup.RepairAction)
		if !report.RepairAvailable && strings.TrimSpace(bringup.RepairAction) != "" {
			report.RepairAvailable = true
		}
		report.ReadyAction = strings.TrimSpace(bringup.ReadyAction)
		report.AcceptanceCommand = strings.TrimSpace(bringup.AcceptanceCommand)
		report.Steps = append([]BrowserDoctorBringupStep(nil), bringup.Steps...)
		report.Summary = BrowserDoctorBringupSummaryText(bringup)
	}
	if value := strings.TrimSpace(preferredRuntime.Backend); value != "" {
		report.PreferredRuntimeBackend = value
	}
	if value := strings.TrimSpace(preferredRuntime.Profile); value != "" {
		report.PreferredRuntimeProfile = value
	}
	if value := strings.TrimSpace(preferredRuntime.Target); value != "" {
		report.PreferredRuntimeTarget = value
	}
	if browserDoctorBringupReportEmpty(report) {
		return nil
	}
	return report
}

func BrowserDoctorBringupReportDetailText(report *BrowserDoctorBringupReport) string {
	if browserDoctorBringupReportEmpty(report) {
		return ""
	}
	summary := strings.TrimSpace(report.Summary)
	if summary == "" {
		summary = BrowserDoctorBringupSummaryText(&BrowserDoctorBringupSummary{
			PreferredRuntimeBackend: strings.TrimSpace(report.PreferredRuntimeBackend),
			PreferredRuntimeProfile: strings.TrimSpace(report.PreferredRuntimeProfile),
			PreferredRuntimeTarget:  strings.TrimSpace(report.PreferredRuntimeTarget),
			PrimaryStep:             strings.TrimSpace(report.PrimaryStep),
			DoctorAction:            strings.TrimSpace(report.DoctorAction),
			RepairAction:            strings.TrimSpace(report.RepairAction),
			ReadyAction:             strings.TrimSpace(report.ReadyAction),
			AcceptanceCommand:       strings.TrimSpace(report.AcceptanceCommand),
			Steps:                   append([]BrowserDoctorBringupStep(nil), report.Steps...),
		})
	}
	if summary == "" {
		return ""
	}
	extras := make([]string, 0, 10)
	if value := strings.TrimSpace(report.PrimaryStep); value != "" {
		extras = append(extras, "primary_step="+value)
	}
	if value := strings.TrimSpace(report.SubstrateSource); value != "" {
		extras = append(extras, "substrate_source="+value)
	}
	if value := strings.TrimSpace(report.SubstrateEndpoint); value != "" {
		extras = append(extras, "substrate_endpoint="+value)
	}
	if value := strings.TrimSpace(report.SubstrateBackend); value != "" {
		extras = append(extras, "substrate_backend="+value)
	}
	if value := strings.TrimSpace(report.SubstrateProfile); value != "" {
		extras = append(extras, "substrate_profile="+value)
	}
	if value := strings.TrimSpace(report.SubstrateTarget); value != "" {
		extras = append(extras, "substrate_target="+value)
	}
	if value := strings.TrimSpace(report.SubstratePosture); value != "" {
		extras = append(extras, "substrate_posture="+value)
	}
	if value := strings.TrimSpace(report.SubstrateStatus); value != "" {
		extras = append(extras, "substrate_status="+value)
	}
	if value := strings.TrimSpace(report.PreferredRuntimeBackend); value != "" {
		extras = append(extras, "preferred_runtime_backend="+value)
	}
	if value := strings.TrimSpace(report.PreferredRuntimeProfile); value != "" {
		extras = append(extras, "preferred_runtime_profile="+value)
	}
	if value := strings.TrimSpace(report.PreferredRuntimeTarget); value != "" {
		extras = append(extras, "preferred_runtime_target="+value)
	}
	if value := strings.TrimSpace(report.DefaultCandidate); value != "" {
		extras = append(extras, "default_candidate="+value)
	}
	if value := strings.TrimSpace(report.DefaultCandidateSource); value != "" {
		extras = append(extras, "default_candidate_source="+value)
	}
	if value := strings.TrimSpace(report.DefaultCandidateEndpoint); value != "" {
		extras = append(extras, "default_candidate_endpoint="+value)
	}
	if value := strings.TrimSpace(report.DefaultCandidateBackend); value != "" {
		extras = append(extras, "default_candidate_backend="+value)
	}
	if value := strings.TrimSpace(report.DefaultCandidateProfile); value != "" {
		extras = append(extras, "default_candidate_profile="+value)
	}
	if value := strings.TrimSpace(report.DefaultCandidateTarget); value != "" {
		extras = append(extras, "default_candidate_target="+value)
	}
	if value := strings.TrimSpace(report.SelectionStrategy); value != "" {
		extras = append(extras, "selection_strategy="+value)
	}
	if value := strings.TrimSpace(report.SubstrateReason); value != "" {
		extras = append(extras, "substrate_reason="+strconv.Quote(value))
	}
	if report.RepairAvailable {
		extras = append(extras, "repair_available=true")
	}
	if len(extras) == 0 {
		return summary
	}
	return summary + " [" + strings.Join(extras, " ") + "]"
}

func BrowserDoctorBringupDisplayText(report *BrowserDoctorBringupReport, bringup *BrowserDoctorBringupSummary) string {
	if summary := strings.TrimSpace(BrowserDoctorBringupReportDetailText(report)); summary != "" {
		return summary
	}
	return strings.TrimSpace(BrowserDoctorBringupDetailText(bringup))
}

func browserDoctorBringupReportEmpty(report *BrowserDoctorBringupReport) bool {
	return report == nil ||
		(strings.TrimSpace(report.SubstrateSource) == "" &&
			strings.TrimSpace(report.SubstrateEndpoint) == "" &&
			strings.TrimSpace(report.SubstrateBackend) == "" &&
			strings.TrimSpace(report.SubstrateProfile) == "" &&
			strings.TrimSpace(report.SubstrateTarget) == "" &&
			strings.TrimSpace(report.SubstratePosture) == "" &&
			strings.TrimSpace(report.SubstrateStatus) == "" &&
			strings.TrimSpace(report.PreferredRuntimeBackend) == "" &&
			strings.TrimSpace(report.PreferredRuntimeProfile) == "" &&
			strings.TrimSpace(report.PreferredRuntimeTarget) == "" &&
			strings.TrimSpace(report.DefaultCandidate) == "" &&
			strings.TrimSpace(report.DefaultCandidateSource) == "" &&
			strings.TrimSpace(report.DefaultCandidateEndpoint) == "" &&
			strings.TrimSpace(report.DefaultCandidateBackend) == "" &&
			strings.TrimSpace(report.DefaultCandidateProfile) == "" &&
			strings.TrimSpace(report.DefaultCandidateTarget) == "" &&
			strings.TrimSpace(report.SelectionStrategy) == "" &&
			strings.TrimSpace(report.SubstrateReason) == "" &&
			!report.RepairAvailable &&
			strings.TrimSpace(report.PrimaryStep) == "" &&
			strings.TrimSpace(report.DoctorAction) == "" &&
			strings.TrimSpace(report.RepairAction) == "" &&
			strings.TrimSpace(report.ReadyAction) == "" &&
			strings.TrimSpace(report.AcceptanceCommand) == "" &&
			len(report.Steps) == 0 &&
			strings.TrimSpace(report.Summary) == "")
}
