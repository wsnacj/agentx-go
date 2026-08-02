package tools

import "strings"

func browserRuntimePayloadRepairCommand(payload *browserRuntimePayload) string {
	if payload == nil {
		return ""
	}
	if summary := payload.WorkbenchLaunchDiagnostics; summary != nil {
		if command := strings.TrimSpace(summary.RepairCommand); command != "" {
			return command
		}
	}
	if summary := payload.LaunchDiagnostics; summary != nil {
		if command := strings.TrimSpace(summary.RepairCommand); command != "" {
			return command
		}
	}
	if doctor := payload.Doctor; doctor != nil && doctor.Launch != nil &&
		browserRuntimeBootstrapErrorCodeSupportsRepair(strings.TrimSpace(doctor.Launch.BootstrapErrorCode)) {
		return strings.TrimSpace(doctor.RepairCommand)
	}
	return ""
}

func browserRuntimeApplyRepairCommandToPayloadShells(payload *browserRuntimePayload) {
	if payload == nil {
		return
	}
	command := browserRuntimePayloadRepairCommand(payload)
	if command == "" {
		return
	}
	browserRuntimeApplyRepairCommandToTopLevelSummary(payload.Explanation, command)
	browserRuntimeApplyRepairCommandToTopLevelSummary(payload.Diagnostics, command)
	browserRuntimeApplyRepairCommandToTopLevelSummary(payload.Summary, command)
	browserRuntimeApplyRepairCommandToTopLevelDisplay(payload.Display, command)
	browserRuntimeApplyRepairCommandToReviewSurface(payload.Review, command)
	browserRuntimeApplyRepairCommandToTopLevelSurface(payload.Surface, command)
	browserRuntimeApplyRepairCommandToTopLevelView(payload.View, command)
	browserRuntimeApplyRepairCommandToWorkbenchDiagnostics(payload.WorkbenchDiagnostics, command)
	browserRuntimeApplyRepairCommandToTopLevelSummary(payload.WorkbenchSummary, command)
	browserRuntimeApplyRepairCommandToWorkbenchDisplay(payload.WorkbenchDisplay, command)
	browserRuntimeApplyRepairCommandToWorkbenchSurface(payload.Workbench, command)
}

func browserRuntimeApplyRepairCommandToTopLevelSummary(summary *browserTopLevelSummary, command string) {
	if summary == nil || strings.TrimSpace(summary.RepairCommand) != "" {
		return
	}
	summary.RepairCommand = strings.TrimSpace(command)
}

func browserRuntimeApplyRepairCommandToTopLevelDisplay(summary *browserTopLevelDisplaySummary, command string) {
	if summary == nil || strings.TrimSpace(summary.RepairCommand) != "" {
		return
	}
	summary.RepairCommand = strings.TrimSpace(command)
}

func browserRuntimeApplyRepairCommandToReviewSurface(summary *browserReviewSurfaceSummary, command string) {
	if summary == nil {
		return
	}
	browserRuntimeApplyRepairCommandToTopLevelSummary(summary.Explanation, command)
	browserRuntimeApplyRepairCommandToTopLevelSummary(summary.Diagnostics, command)
	browserRuntimeApplyRepairCommandToTopLevelSummary(summary.Summary, command)
	browserRuntimeApplyRepairCommandToTopLevelDisplay(summary.Display, command)
}

func browserRuntimeApplyRepairCommandToTopLevelSurface(summary *browserTopLevelSurfaceSummary, command string) {
	if summary == nil {
		return
	}
	if strings.TrimSpace(summary.RepairCommand) == "" {
		summary.RepairCommand = strings.TrimSpace(command)
	}
}

func browserRuntimeApplyRepairCommandToTopLevelView(summary *browserTopLevelViewSummary, command string) {
	if summary == nil {
		return
	}
	if strings.TrimSpace(summary.RepairCommand) == "" {
		summary.RepairCommand = strings.TrimSpace(command)
	}
	browserRuntimeApplyRepairCommandToReviewSurface(summary.Review, command)
}

func browserRuntimeApplyRepairCommandToWorkbenchDiagnostics(summary *browserRuntimeWorkbenchDiagnosticsSummary, command string) {
	if summary == nil || strings.TrimSpace(summary.RepairCommand) != "" {
		return
	}
	summary.RepairCommand = strings.TrimSpace(command)
}

func browserRuntimeApplyRepairCommandToWorkbenchDisplay(summary *browserRuntimeWorkbenchDisplaySummary, command string) {
	if summary == nil || strings.TrimSpace(summary.RepairCommand) != "" {
		return
	}
	summary.RepairCommand = strings.TrimSpace(command)
}

func browserRuntimeApplyRepairCommandToWorkbenchSurface(summary *browserRuntimeWorkbenchSurfaceSummary, command string) {
	if summary == nil {
		return
	}
	if strings.TrimSpace(summary.RepairCommand) == "" {
		summary.RepairCommand = strings.TrimSpace(command)
	}
	browserRuntimeApplyRepairCommandToReviewSurface(summary.Review, command)
	browserRuntimeApplyRepairCommandToTopLevelSummary(summary.Explanation, command)
	browserRuntimeApplyRepairCommandToTopLevelSummary(summary.Diagnostics, command)
	browserRuntimeApplyRepairCommandToTopLevelSummary(summary.Summary, command)
}
