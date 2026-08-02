package tools

import (
	"context"
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

func browserRuntimeApplyRepairActionOutcome(
	payload *browserRuntimePayload,
	outcome agentxbrowserruntime.SharedSessionBrowserRepairActionOutcome,
) {
	if payload == nil {
		return
	}
	payload.Status = strings.TrimSpace(outcome.Status)
	payload.Note = strings.TrimSpace(outcome.Note)
	payload.RepairDecision = strings.TrimSpace(outcome.RepairDecision)
	payload.RepairReady = outcome.RepairReady
	payload.RepairOutput = strings.TrimSpace(outcome.RepairOutput)
	browserRuntimeApplyGuidanceShellProjection(
		payload,
		browserRuntimeGuidanceShellProjectionFromShared(outcome.GuidanceProjection),
	)
}

func browserRuntimeSupportsRepairAction(opts BrowserToolOptions) bool {
	return strings.TrimSpace(opts.RepairScript) != "" && opts.RunCommand != nil
}

func browserRuntimeAugmentActionsWithRepair(opts BrowserToolOptions, actions []string) []string {
	actions = mergeToolMetadataStrings(nil, actions)
	if !browserRuntimeSupportsRepairAction(opts) {
		return actions
	}
	return mergeToolMetadataStrings(actions, []string{"repair"})
}

func browserHostScriptCommand(script string) string {
	script = strings.TrimSpace(script)
	if script == "" {
		return ""
	}
	return "bash " + script
}

func browserRuntimeDispatchRepairAction(
	ctx browserRegistrationContext,
	callCtx context.Context,
) agentxbrowserruntime.SharedSessionBrowserRepairActionOutcome {
	hints := browserRuntimeActionHintsForRegistration(ctx)
	req := agentxbrowserruntime.SharedSessionBrowserRepairActionOutcomeRequest{
		RepairAction: strings.TrimSpace(hints.RepairCommand),
		DoctorAction: strings.TrimSpace(hints.DoctorCommand),
	}
	script := strings.TrimSpace(ctx.opts.RepairScript)
	if script == "" {
		req.Status = "unsupported"
		req.RepairDecision = "repair_command_unavailable"
		req.RepairReady = false
		req.Note = "Bundled browserd bootstrap repair script is not available from the current workspace root."
		return agentxbrowserruntime.BuildSharedSessionBrowserRepairActionOutcome(req)
	}
	out, err := runBrowserCommand(callCtx, ctx.opts.RunCommand, "bash", []string{script})
	req.RepairOutput = truncateToolText(strings.TrimSpace(string(out)), 2_000)
	if err != nil {
		req.Status = "error"
		req.RepairDecision = "repair_failed"
		req.RepairReady = false
		req.Note = firstNonEmpty(strings.TrimSpace(req.RepairOutput), strings.TrimSpace(err.Error()), "Bundled browserd bootstrap repair failed.")
		return agentxbrowserruntime.BuildSharedSessionBrowserRepairActionOutcome(req)
	}
	req.Status = "ok"
	req.RepairDecision = "repaired"
	req.RepairReady = true
	req.Note = "Bundled browserd bootstrap repair completed."
	return agentxbrowserruntime.BuildSharedSessionBrowserRepairActionOutcome(req)
}
