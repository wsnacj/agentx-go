package tools

import "strings"

var browserRuntimeCompatTabsActions = map[string]string{
	"tabs":  "list",
	"focus": "focus",
	"close": "close",
}

func browserRuntimeTranslatedCommandForUnifiedAction(ctx browserRegistrationContext, action string) string {
	action = strings.ToLower(strings.TrimSpace(action))
	if action == "" {
		return ""
	}
	if action == "browser" && browserRuntimeRegisteredOrEnabledTool(ctx, "browser_runtime") {
		return "browser_runtime action=workbench"
	}
	if alias, ok := browserUnifiedRuntimeActionAliases[action]; ok && browserRuntimeRegisteredOrEnabledTool(ctx, "browser_runtime") {
		command := "browser_runtime action=" + strings.TrimSpace(alias.Action)
		if goal := strings.TrimSpace(alias.CoordinationGoal); goal != "" {
			command += " coordination_goal=" + goal
		}
		return command
	}
	if tabsAction, ok := browserRuntimeCompatTabsActions[action]; ok {
		if tabsTool := browserCompatRegisteredOrEnabledToolForActKind(ctx, "list_tabs"); tabsTool != "" {
			return tabsTool + " action=" + tabsAction
		}
		if browserRuntimeRegisteredOrEnabledTool(ctx, "browser_runtime") {
			return "browser_runtime action=sessions"
		}
	}
	return ""
}

func browserRuntimeToolAwareActionCommand(ctx browserRegistrationContext, command string) string {
	command = strings.TrimSpace(command)
	if command == "" || browserRuntimeSurfacePrefersUnifiedBrowser(ctx) {
		return command
	}
	lower := strings.ToLower(command)
	if translated := browserRuntimeTranslatedCommandForUnifiedAction(ctx, lower); translated != "" && lower == strings.TrimSpace(lower) {
		return translated
	}
	const prefix = "browser action="
	if !strings.HasPrefix(lower, prefix) {
		return command
	}
	if translated := browserRuntimeTranslatedCommandForUnifiedAction(ctx, command[len(prefix):]); translated != "" {
		return translated
	}
	return command
}

func browserRuntimeToolAwareActionCommands(ctx browserRegistrationContext, commands []string) []string {
	if len(commands) == 0 {
		return nil
	}
	out := make([]string, 0, len(commands))
	for _, command := range commands {
		translated := browserRuntimeToolAwareActionCommand(ctx, command)
		if translated == "" {
			continue
		}
		out = append(out, translated)
	}
	return mergeToolMetadataStrings(nil, out)
}

func browserRuntimeToolAwareNextStepAlias(ctx browserRegistrationContext, alias string) string {
	alias = strings.ToLower(strings.TrimSpace(alias))
	if alias == "" || browserRuntimeSurfacePrefersUnifiedBrowser(ctx) {
		return alias
	}
	if runtimeAlias, ok := browserUnifiedRuntimeActionAliases[alias]; ok {
		return strings.TrimSpace(runtimeAlias.Action)
	}
	switch alias {
	case "browser":
		if browserRuntimeRegisteredOrEnabledTool(ctx, "browser_runtime") {
			return "workbench"
		}
	}
	if tabsAction, ok := browserRuntimeCompatTabsActions[alias]; ok {
		if tabsTool := browserCompatRegisteredOrEnabledToolForActKind(ctx, "list_tabs"); tabsTool != "" {
			return tabsAction
		}
		if browserRuntimeRegisteredOrEnabledTool(ctx, "browser_runtime") {
			return "sessions"
		}
	}
	return alias
}

func browserRuntimeApplyToolAwareSummaryCommands(ctx browserRegistrationContext, summary *browserTopLevelSummary) {
	if summary == nil {
		return
	}
	summary.NextStepAlias = browserRuntimeToolAwareNextStepAlias(ctx, summary.NextStepAlias)
	summary.PrimaryBrowserAction = browserRuntimeToolAwareActionCommand(ctx, summary.PrimaryBrowserAction)
	summary.NextStep = browserRuntimeToolAwareActionCommand(ctx, summary.NextStep)
}

func browserRuntimeApplyToolAwareDisplayCommands(ctx browserRegistrationContext, summary *browserTopLevelDisplaySummary) {
	if summary == nil {
		return
	}
	summary.NextStepAlias = browserRuntimeToolAwareNextStepAlias(ctx, summary.NextStepAlias)
	summary.PrimaryBrowserAction = browserRuntimeToolAwareActionCommand(ctx, summary.PrimaryBrowserAction)
	summary.NextStep = browserRuntimeToolAwareActionCommand(ctx, summary.NextStep)
}

func browserRuntimeApplyToolAwareSurfaceCommands(ctx browserRegistrationContext, summary *browserTopLevelSurfaceSummary) {
	if summary == nil {
		return
	}
	summary.NextStepAlias = browserRuntimeToolAwareNextStepAlias(ctx, summary.NextStepAlias)
	summary.PrimaryBrowserAction = browserRuntimeToolAwareActionCommand(ctx, summary.PrimaryBrowserAction)
	summary.NextStep = browserRuntimeToolAwareActionCommand(ctx, summary.NextStep)
}

func browserRuntimeApplyToolAwareReviewCommands(ctx browserRegistrationContext, review *browserReviewSurfaceSummary) {
	if review == nil {
		return
	}
	browserRuntimeApplyToolAwareSummaryCommands(ctx, review.Explanation)
	browserRuntimeApplyToolAwareSummaryCommands(ctx, review.Diagnostics)
	browserRuntimeApplyToolAwareSummaryCommands(ctx, review.Summary)
	browserRuntimeApplyToolAwareDisplayCommands(ctx, review.Display)
}

func browserRuntimeApplyToolAwareViewCommands(ctx browserRegistrationContext, summary *browserTopLevelViewSummary) {
	if summary == nil {
		return
	}
	summary.NextStepAlias = browserRuntimeToolAwareNextStepAlias(ctx, summary.NextStepAlias)
	summary.PrimaryBrowserAction = browserRuntimeToolAwareActionCommand(ctx, summary.PrimaryBrowserAction)
	summary.NextStep = browserRuntimeToolAwareActionCommand(ctx, summary.NextStep)
	browserRuntimeApplyToolAwareReviewCommands(ctx, summary.Review)
}

func browserRuntimeApplyToolAwareWorkbenchDiagnosticsCommands(ctx browserRegistrationContext, summary *browserRuntimeWorkbenchDiagnosticsSummary) {
	if summary == nil {
		return
	}
	summary.NextStepAlias = browserRuntimeToolAwareNextStepAlias(ctx, summary.NextStepAlias)
	summary.PrimaryBrowserAction = browserRuntimeToolAwareActionCommand(ctx, summary.PrimaryBrowserAction)
	summary.NextStep = browserRuntimeToolAwareActionCommand(ctx, summary.NextStep)
}

func browserRuntimeApplyToolAwareWorkbenchDisplayCommands(ctx browserRegistrationContext, summary *browserRuntimeWorkbenchDisplaySummary) {
	if summary == nil {
		return
	}
	summary.NextStepAlias = browserRuntimeToolAwareNextStepAlias(ctx, summary.NextStepAlias)
	summary.PrimaryBrowserAction = browserRuntimeToolAwareActionCommand(ctx, summary.PrimaryBrowserAction)
	summary.NextStep = browserRuntimeToolAwareActionCommand(ctx, summary.NextStep)
}

func browserRuntimeApplyToolAwareWorkbenchSurfaceCommands(ctx browserRegistrationContext, summary *browserRuntimeWorkbenchSurfaceSummary) {
	if summary == nil {
		return
	}
	summary.Explanation = browserRuntimeApplyToolAwareSummaryCommandsRet(ctx, summary.Explanation)
	summary.Diagnostics = browserRuntimeApplyToolAwareSummaryCommandsRet(ctx, summary.Diagnostics)
	summary.Summary = browserRuntimeApplyToolAwareSummaryCommandsRet(ctx, summary.Summary)
	summary.PrimaryBrowserAction = browserRuntimeToolAwareActionCommand(ctx, summary.PrimaryBrowserAction)
	summary.NextStep = browserRuntimeToolAwareActionCommand(ctx, summary.NextStep)
	summary.RecommendedBrowserActions = browserRuntimeToolAwareActionCommands(ctx, summary.RecommendedBrowserActions)
	browserRuntimeApplyToolAwareReviewCommands(ctx, summary.Review)
}

func browserRuntimeApplyToolAwareSummaryCommandsRet(ctx browserRegistrationContext, summary *browserTopLevelSummary) *browserTopLevelSummary {
	browserRuntimeApplyToolAwareSummaryCommands(ctx, summary)
	return summary
}

func browserRuntimeApplyToolAwareCoordinationCommands(ctx browserRegistrationContext, coordination *browserRuntimeCoordination) {
	if coordination == nil {
		return
	}
	coordination.SyncBrowserAction = browserRuntimeToolAwareActionCommand(ctx, coordination.SyncBrowserAction)
	coordination.PrepareBrowserAction = browserRuntimeToolAwareActionCommand(ctx, coordination.PrepareBrowserAction)
	coordination.RestartBrowserAction = browserRuntimeToolAwareActionCommand(ctx, coordination.RestartBrowserAction)
	coordination.TeardownBrowserAction = browserRuntimeToolAwareActionCommand(ctx, coordination.TeardownBrowserAction)
	coordination.PrimaryBrowserAction = browserRuntimeToolAwareActionCommand(ctx, coordination.PrimaryBrowserAction)
	coordination.NextStep = browserRuntimeToolAwareActionCommand(ctx, coordination.NextStep)
	coordination.RecommendedBrowserActions = browserRuntimeToolAwareActionCommands(ctx, coordination.RecommendedBrowserActions)
}

func browserRuntimeApplyToolAwareActionCommands(ctx browserRegistrationContext, payload *browserRuntimePayload) {
	if payload == nil || browserRuntimeSurfacePrefersUnifiedBrowser(ctx) {
		return
	}
	payload.WorkbenchPrimaryBrowserAction = browserRuntimeToolAwareActionCommand(ctx, payload.WorkbenchPrimaryBrowserAction)
	payload.WorkbenchNextStep = browserRuntimeToolAwareActionCommand(ctx, payload.WorkbenchNextStep)
	payload.WorkbenchRecommendedBrowserActions = browserRuntimeToolAwareActionCommands(ctx, payload.WorkbenchRecommendedBrowserActions)
	browserRuntimeApplyToolAwareWorkbenchDiagnosticsCommands(ctx, payload.WorkbenchDiagnostics)
	browserRuntimeApplyToolAwareSummaryCommands(ctx, payload.WorkbenchSummary)
	browserRuntimeApplyToolAwareWorkbenchDisplayCommands(ctx, payload.WorkbenchDisplay)
	browserRuntimeApplyToolAwareWorkbenchSurfaceCommands(ctx, payload.Workbench)
	browserRuntimeApplyToolAwareReviewCommands(ctx, payload.Review)
	browserRuntimeApplyToolAwareSummaryCommands(ctx, payload.Explanation)
	browserRuntimeApplyToolAwareSummaryCommands(ctx, payload.Diagnostics)
	browserRuntimeApplyToolAwareSummaryCommands(ctx, payload.Summary)
	browserRuntimeApplyToolAwareDisplayCommands(ctx, payload.Display)
	browserRuntimeApplyToolAwareSurfaceCommands(ctx, payload.Surface)
	browserRuntimeApplyToolAwareViewCommands(ctx, payload.View)
	if payload.SessionBinding != nil {
		payload.SessionBinding.SessionHealthRecoveryAction = browserRuntimeToolAwareActionCommand(ctx, payload.SessionBinding.SessionHealthRecoveryAction)
		browserRuntimeApplyToolAwareCoordinationCommands(ctx, payload.SessionBinding.Coordination)
	}
}
