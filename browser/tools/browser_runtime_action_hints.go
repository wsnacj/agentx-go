package tools

import "strings"

type browserRuntimeActionHints struct {
	InspectCommand string
	DoctorCommand  string
	ReadyCommand   string
	ReadyAlias     string
	RepairCommand  string
	RepairAlias    string
	LaunchCommand  string
	LaunchAlias    string
}

func browserRuntimeSurfacePrefersUnifiedBrowser(ctx browserRegistrationContext) bool {
	return ctx.enabledTools["browser"]
}

func browserRuntimeActionHintsForRegistration(ctx browserRegistrationContext) browserRuntimeActionHints {
	repairCommand := ""
	repairAlias := ""
	if browserRuntimeSupportsRepairAction(ctx.opts) {
		repairAlias = "repair"
		if browserRuntimeSurfacePrefersUnifiedBrowser(ctx) {
			repairCommand = "browser action=repair"
		} else {
			repairCommand = "browser_runtime action=repair"
		}
	}
	if browserRuntimeSurfacePrefersUnifiedBrowser(ctx) {
		return browserRuntimeActionHints{
			InspectCommand: "browser action=inspect",
			DoctorCommand:  "browser action=doctor",
			ReadyCommand:   "browser action=ready",
			ReadyAlias:     "ready",
			RepairCommand:  repairCommand,
			RepairAlias:    repairAlias,
			LaunchCommand:  "browser action=launch",
			LaunchAlias:    "launch",
		}
	}
	return browserRuntimeActionHints{
		InspectCommand: "browser_runtime action=status",
		DoctorCommand:  "browser_runtime action=doctor",
		ReadyCommand:   "browser_runtime action=prepare",
		ReadyAlias:     "prepare",
		RepairCommand:  repairCommand,
		RepairAlias:    repairAlias,
		LaunchCommand:  "browser_runtime action=start",
		LaunchAlias:    "start",
	}
}

func browserRuntimeActionHintsSummary(
	inspectCommand string,
	doctorCommand string,
	readyCommand string,
) string {
	parts := make([]string, 0, 3)
	for _, candidate := range []string{
		strings.TrimSpace(inspectCommand),
		strings.TrimSpace(doctorCommand),
		strings.TrimSpace(readyCommand),
	} {
		if candidate == "" || containsString(parts, candidate) {
			continue
		}
		parts = append(parts, candidate)
	}
	switch len(parts) {
	case 0:
		return ""
	case 1:
		return parts[0]
	case 2:
		return parts[0] + " or " + parts[1]
	default:
		return parts[0] + ", " + parts[1] + ", or " + parts[2]
	}
}
