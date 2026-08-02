package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	types "github.com/wsnacj/agentx-go/components/llm"
)

var browserUnifiedActionAliases = map[string]string{
	"tabs":  "list_tabs",
	"focus": "focus_tab",
	"close": "close_tab",
	"pdf":   "save_pdf",
}

type browserUnifiedRuntimeAlias struct {
	Action           string
	CoordinationGoal string
}

var browserUnifiedRuntimeActionAliases = map[string]browserUnifiedRuntimeAlias{
	"inspect":        {Action: "status"},
	"doctor":         {Action: "status"},
	"repair":         {Action: "repair"},
	"ensure":         {Action: "coordinate", CoordinationGoal: "ensure"},
	"refresh":        {Action: "refresh"},
	"sync":           {Action: "coordinate", CoordinationGoal: "sync"},
	"teardown":       {Action: "coordinate", CoordinationGoal: "teardown"},
	"ready":          {Action: "prepare"},
	"inventory":      {Action: "profiles"},
	"handles":        {Action: "sessions"},
	"launch":         {Action: "start"},
	"halt":           {Action: "stop"},
	"reset":          {Action: "clear_session"},
	"adopt":          {Action: "sync_session"},
	"new_profile":    {Action: "create_profile"},
	"remove_profile": {Action: "delete_profile"},
	"pin_profile":    {Action: "select_profile"},
	"unpin_profile":  {Action: "clear_profile"},
	"pin_target":     {Action: "select_target"},
	"unpin_target":   {Action: "clear_target"},
}

func browserDefinition(runtimeActions []string, actKinds []string) types.Tool {
	actDef := browserActDefinition(actKinds)
	runtimeDef := browserRuntimeDefinition(runtimeActions)
	availableActions := browserUnifiedAvailableActions(runtimeActions, actKinds)
	props := map[string]any{}
	if raw, ok := actDef.Function.Parameters["properties"].(map[string]any); ok {
		for key, value := range raw {
			props[key] = value
		}
	}
	if raw, ok := runtimeDef.Function.Parameters["properties"].(map[string]any); ok {
		for key, value := range raw {
			if strings.EqualFold(key, "action") {
				continue
			}
			props[key] = value
		}
	}
	delete(props, "action")
	props["action"] = map[string]any{
		"type":        "string",
		"enum":        append([]string(nil), availableActions...),
		"description": "Browser operation. Use extract for page title/text/summary, snapshot for element refs before follow-up actions, doctor/inspect/status for runtime diagnostics, and ready/prepare only for runtime bring-up. Use reset/clear_session only when explicitly clearing browser session state.",
	}
	props["operation"] = map[string]any{
		"type":        "string",
		"enum":        append([]string(nil), availableActions...),
		"description": "Compatibility alias for action. Prefer action for new calls.",
	}
	props["kind"] = map[string]any{
		"type":        "string",
		"enum":        append([]string(nil), actKinds...),
		"description": "Raw action kind when action=act. Prefer setting action directly to open, navigate, extract, snapshot, click, type, screenshot, or a visible action enum when possible.",
	}
	props["remember_target"] = map[string]any{
		"type":        "boolean",
		"description": "When true, remember the selected browser target for later actions in this session.",
	}
	props["remember"] = map[string]any{
		"type":        "boolean",
		"description": "Alias used by runtime profile/target selection actions to persist the selection in the session.",
	}
	return types.Tool{
		Type: "function",
		Function: types.Function{
			Name:         "browser",
			Description:  "Unified browser workbench entrypoint. Use one tool for managed browser runtime inspection/control plus page actions like open, navigate, tabs, extract, snapshot, screenshot, debug, form interaction, downloads, and state changes. For title/text/summary requests, prefer `action=extract`; use `action=snapshot` when you need element refs or structural grounding for follow-up page actions; if a click, modal, consent, or route change updates the page state, take a fresh snapshot before reusing targeted refs; for auth/registration/checkout flows, clear visible blockers first and verify each state-changing click before continuing. Runtime actions `doctor|inspect|status` are diagnostics, `ready|prepare` are bring-up, and `reset|clear_session` clears state; avoid them for ordinary page reads unless the user explicitly asked or a prior browser action failed. If diagnostics mention the legacy host path, pass explicit `runtime_target=host` for host-only runtime actions.",
			Parameters:   browserUnifiedParametersSchema(props),
			OutputSchema: browserUnifiedOutputSchema(),
		},
	}
}

func browserUnifiedAvailableActions(runtimeActions []string, actKinds []string) []string {
	actions := make([]string, 0, len(runtimeActions)+len(actKinds)+5)
	for _, action := range runtimeActions {
		normalized := strings.ToLower(strings.TrimSpace(action))
		if normalized == "" {
			continue
		}
		switch normalized {
		case "status":
			actions = append(actions, "inspect", "doctor", "status")
			continue
		case "prepare":
			actions = append(actions, "ready", "prepare")
			continue
		case "profiles":
			actions = append(actions, "inventory", "profiles")
			continue
		case "sessions":
			actions = append(actions, "handles", "sessions")
			continue
		case "start":
			actions = append(actions, "launch", "start")
			continue
		case "stop":
			actions = append(actions, "halt", "stop")
			continue
		case "coordinate":
			actions = append(actions, "ensure", "refresh", "sync", "teardown", "reset", "coordinate")
			continue
		case "clear_session":
			actions = append(actions, "reset", "clear_session")
			continue
		case "sync_session":
			actions = append(actions, "adopt", "sync_session")
			continue
		case "create_profile":
			actions = append(actions, "new_profile", "create_profile")
			continue
		case "delete_profile":
			actions = append(actions, "remove_profile", "delete_profile")
			continue
		case "select_profile":
			actions = append(actions, "pin_profile", "select_profile")
			continue
		case "clear_profile":
			actions = append(actions, "unpin_profile", "clear_profile")
			continue
		case "select_target":
			actions = append(actions, "pin_target", "select_target")
			continue
		case "clear_target":
			actions = append(actions, "unpin_target", "clear_target")
			continue
		}
		actions = append(actions, normalized)
	}
	actions = append(actions, actKinds...)
	if containsString(actKinds, "list_tabs") {
		actions = append(actions, "tabs")
	}
	if containsString(actKinds, "focus_tab") {
		actions = append(actions, "focus")
	}
	if containsString(actKinds, "close_tab") {
		actions = append(actions, "close")
	}
	if containsString(actKinds, "save_pdf") {
		actions = append(actions, "pdf")
	}
	if len(actKinds) > 0 {
		actions = append(actions, "act")
	}
	return mergeToolMetadataStrings(nil, actions)
}

func registerBrowserUnifiedTool(ctx browserRegistrationContext) {
	runtimeActions := browserRuntimeAvailableActions(ctx)
	actKinds := ctx.capabilities.SupportedActKinds()
	ctx.reg.Register(browserDefinition(runtimeActions, actKinds), func(callCtx context.Context, call types.FunctionCall) (string, error) {
		params, err := decodeArgs(call.Arguments)
		if err != nil {
			return "", err
		}
		action := browserNormalizeToolToken(firstString(params, "action", "operation", "mode"))
		if action == "" {
			if containsString(runtimeActions, "workbench") {
				action = "workbench"
			} else {
				action = "status"
			}
		}
		availableActions := browserUnifiedAvailableActions(runtimeActions, actKinds)
		if !containsString(availableActions, action) &&
			!browserUnifiedCanDelegateActAction(ctx, runtimeActions, params, action) {
			return "", fmt.Errorf("browser: action must be one of %s", strings.Join(availableActions, ", "))
		}
		if _, ok := browserUnifiedRuntimeActionAliases[action]; ok || containsString(runtimeActions, action) {
			return browserUnifiedExecuteDelegated(callCtx, ctx, "browser_runtime", browserUnifiedRuntimeArgs(params, action), action)
		}
		if action == "act" && browserNormalizeToolToken(firstString(params, "kind")) == "" {
			return "", browserMissingActKindError("browser", "act")
		}
		return browserUnifiedExecuteDelegated(callCtx, ctx, "browser_act", browserUnifiedActArgs(params, action), action)
	})
}

func browserUnifiedExecuteDelegated(callCtx context.Context, ctx browserRegistrationContext, name string, args map[string]any, action string) (string, error) {
	blob, err := json.Marshal(args)
	if err != nil {
		return "", err
	}
	out, err := ctx.reg.Execute(callCtx, types.FunctionCall{
		Name:      name,
		Arguments: string(blob),
	})
	if err != nil {
		return "", err
	}
	if name == "browser_runtime" {
		out, err = browserUnifiedApplyRuntimeAliasSurface(ctx, action, out)
		if err != nil {
			return "", err
		}
	}
	return browserUnifiedApplyExplanationAlias(out)
}

func browserUnifiedApplyExplanationAlias(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || !strings.HasPrefix(trimmed, "{") {
		return raw, nil
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return raw, nil
	}
	inputs := browserRuntimeTopLevelAliasInputsFromPayload(payload)
	mutated := false
	if inputs.defaultCandidateRoute != (browserRuntimeRouteDescriptor{}) {
		if defaultCandidateMutated, err := browserRuntimeApplyTopLevelDefaultCandidateAliasProjection(payload, &inputs); err != nil {
			return "", err
		} else if defaultCandidateMutated {
			mutated = true
		}
	}
	if routeCapabilityMutated, err := browserRuntimeApplyTopLevelRouteCapabilityAliasProjection(payload, &inputs); err != nil {
		return "", err
	} else if routeCapabilityMutated {
		mutated = true
	}
	topLevelMutated, err := browserUnifiedApplyTopLevelAliasProjection(payload, &inputs)
	if err != nil {
		return "", err
	}
	if topLevelMutated {
		mutated = true
	}
	if !mutated {
		return raw, nil
	}
	blob, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(blob), nil
}

func browserUnifiedHasNonNullJSONField(payload map[string]json.RawMessage, key string) bool {
	value, ok := payload[key]
	if !ok {
		return false
	}
	trimmed := strings.TrimSpace(string(value))
	return trimmed != "" && trimmed != "null"
}

func browserUnifiedApplyTopLevelAliasProjection(
	payload map[string]json.RawMessage,
	inputs *browserRuntimeTopLevelAliasInputs,
) (bool, error) {
	return browserRuntimeApplyTopLevelAliasProjection(payload, inputs)
}

func browserUnifiedSummaryEmpty(summary browserTopLevelSummary) bool {
	return strings.TrimSpace(summary.Category) == "" &&
		strings.TrimSpace(summary.State) == "" &&
		strings.TrimSpace(summary.SummaryCode) == "" &&
		strings.TrimSpace(summary.RepairCommand) == "" &&
		summary.DefaultCandidateRoute == (browserRuntimeRouteDescriptor{}) &&
		strings.TrimSpace(summary.NextStepAlias) == "" &&
		strings.TrimSpace(summary.ManualRetryHint) == "" &&
		!summary.ResolvedViaFallback &&
		strings.TrimSpace(summary.PrimaryBrowserAction) == "" &&
		strings.TrimSpace(summary.PrimaryNodeAction) == "" &&
		strings.TrimSpace(summary.NextStep) == ""
}

func browserUnifiedWorkbenchEmpty(summary browserRuntimeWorkbenchSurfaceSummary) bool {
	return !summary.Ready &&
		len(summary.Sections) == 0 &&
		strings.TrimSpace(summary.RepairCommand) == "" &&
		summary.DefaultCandidateRoute == (browserRuntimeRouteDescriptor{}) &&
		summary.Review == nil &&
		len(summary.BrowserTools) == 0 &&
		len(summary.ArtifactTools) == 0 &&
		len(summary.ArtifactKinds) == 0 &&
		strings.TrimSpace(summary.ArtifactContract) == "" &&
		len(summary.BrowserActKinds) == 0 &&
		strings.TrimSpace(summary.BrowserSurface) == "" &&
		len(summary.BrowserOptInTargets) == 0 &&
		summary.Explanation == nil &&
		summary.Diagnostics == nil &&
		summary.Summary == nil &&
		strings.TrimSpace(summary.PrimaryBrowserAction) == "" &&
		strings.TrimSpace(summary.PrimaryNodeAction) == "" &&
		strings.TrimSpace(summary.NextStep) == "" &&
		len(summary.RecommendedBrowserActions) == 0 &&
		len(summary.RecommendedNodeActions) == 0
}

func browserUnifiedWorkbenchDisplayEmpty(summary browserRuntimeWorkbenchDisplaySummary) bool {
	return !summary.Ready &&
		len(summary.Sections) == 0 &&
		strings.TrimSpace(summary.Category) == "" &&
		strings.TrimSpace(summary.State) == "" &&
		strings.TrimSpace(summary.SummaryCode) == "" &&
		strings.TrimSpace(summary.RepairCommand) == "" &&
		summary.DefaultCandidateRoute == (browserRuntimeRouteDescriptor{}) &&
		strings.TrimSpace(summary.NextStepAlias) == "" &&
		strings.TrimSpace(summary.ManualRetryHint) == "" &&
		!summary.ResolvedViaFallback &&
		strings.TrimSpace(summary.PrimaryBrowserAction) == "" &&
		strings.TrimSpace(summary.PrimaryNodeAction) == "" &&
		strings.TrimSpace(summary.NextStep) == ""
}

func browserUnifiedRuntimeArgs(params map[string]any, action string) map[string]any {
	out := browserUnifiedCloneParams(params)
	if alias, ok := browserUnifiedRuntimeActionAliases[action]; ok {
		out["action"] = alias.Action
		if strings.TrimSpace(alias.CoordinationGoal) != "" {
			out["coordination_goal"] = alias.CoordinationGoal
		}
		return out
	}
	out["action"] = action
	return out
}

func browserUnifiedActArgs(params map[string]any, action string) map[string]any {
	out := browserUnifiedCloneParams(params)
	kind := browserUnifiedActKind(params, action)
	out["kind"] = kind
	delete(out, "action")
	delete(out, "operation")
	delete(out, "mode")
	return out
}

func browserUnifiedActKind(params map[string]any, action string) string {
	action = browserNormalizeToolToken(action)
	if alias := browserNormalizeToolToken(browserUnifiedActionAliases[action]); alias != "" {
		return alias
	}
	if action == "act" {
		return browserNormalizeToolToken(firstString(params, "kind"))
	}
	return action
}

func browserUnifiedCloneParams(params map[string]any) map[string]any {
	out := make(map[string]any, len(params))
	for key, value := range params {
		out[key] = value
	}
	return out
}
