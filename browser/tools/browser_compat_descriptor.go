package tools

import (
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

type browserCompatDescriptor struct {
	Name                             string
	ActKind                          string
	ArtifactSource                   string
	ArtifactKind                     string
	ReadOnly                         bool
	LikelySideEffect                 bool
	BuiltinRisk                      RiskLevel
	MetadataCapabilities             []string
	MetadataAuditTags                []string
	ForceConfirmationGuardianReview  bool
	ForceConfirmationGuardianActions []string
	ForceConfirmationRememberActions []string
}

var browserCompatDescriptors = []browserCompatDescriptor{
	{
		ActKind:              "open",
		LikelySideEffect:     true,
		BuiltinRisk:          RiskMedium,
		MetadataCapabilities: []string{"browser", "network", "browser_kind:open"},
		MetadataAuditTags:    []string{"browser", "external_content"},
	},
	{
		ActKind:                         "navigate",
		LikelySideEffect:                true,
		BuiltinRisk:                     RiskMedium,
		MetadataCapabilities:            []string{"browser", "network", "tabs", "browser_kind:navigate"},
		MetadataAuditTags:               []string{"browser", "interactive_browser"},
		ForceConfirmationGuardianReview: true,
	},
	{
		ActKind:                          "list_tabs",
		LikelySideEffect:                 true,
		BuiltinRisk:                      RiskMedium,
		MetadataCapabilities:             []string{"browser", "read", "tabs", "browser_kind:list_tabs", "browser_kind:focus_tab", "browser_kind:close_tab"},
		MetadataAuditTags:                []string{"browser", "interactive_browser"},
		ForceConfirmationGuardianActions: []string{"focus"},
		ForceConfirmationRememberActions: []string{"list"},
	},
	{
		ActKind:                         "extract",
		ArtifactSource:                  "browser",
		ReadOnly:                        true,
		BuiltinRisk:                     RiskMedium,
		MetadataCapabilities:            []string{"browser", "network", "read", "browser_kind:extract"},
		MetadataAuditTags:               []string{"browser", "external_content"},
		ForceConfirmationGuardianReview: true,
	},
	{
		ActKind:                         "screenshot",
		ArtifactSource:                  "browser",
		ArtifactKind:                    "screenshot",
		LikelySideEffect:                true,
		BuiltinRisk:                     RiskHigh,
		MetadataCapabilities:            []string{"browser", "read", "screenshot", "artifact_output", "artifact_contract:" + strings.ReplaceAll(browserArtifactContract, "+", "_"), "artifact_kind:screenshot", "browser_kind:screenshot"},
		MetadataAuditTags:               []string{"browser", "external_content"},
		ForceConfirmationGuardianReview: true,
	},
	{
		ActKind:                         "click",
		LikelySideEffect:                true,
		BuiltinRisk:                     RiskMedium,
		MetadataCapabilities:            []string{"browser", "write", "dom", "browser_kind:click"},
		MetadataAuditTags:               []string{"browser", "interactive_browser", "side_effect"},
		ForceConfirmationGuardianReview: true,
	},
	{
		ActKind:                         "type",
		LikelySideEffect:                true,
		BuiltinRisk:                     RiskMedium,
		MetadataCapabilities:            []string{"browser", "write", "dom", "browser_kind:type"},
		MetadataAuditTags:               []string{"browser", "interactive_browser", "side_effect"},
		ForceConfirmationGuardianReview: true,
	},
	{
		ActKind:                         "evaluate",
		LikelySideEffect:                true,
		BuiltinRisk:                     RiskHigh,
		MetadataCapabilities:            []string{"browser", "exec", "dom", "browser_kind:evaluate"},
		MetadataAuditTags:               []string{"browser", "interactive_browser", "side_effect"},
		ForceConfirmationGuardianReview: true,
	},
}

var browserCompatDescriptorsByName = func() map[string]browserCompatDescriptor {
	out := make(map[string]browserCompatDescriptor, len(browserCompatDescriptors))
	for _, item := range browserCompatDescriptors {
		descriptor, ok := browserCompatDescriptorWithResolvedName(item)
		if !ok {
			continue
		}
		out[descriptor.Name] = descriptor
	}
	return out
}()

func browserCompatDescriptorWithResolvedName(item browserCompatDescriptor) (browserCompatDescriptor, bool) {
	item.ActKind = browserNormalizeToolToken(item.ActKind)
	name := browserCompatToolForManagedOptInActKind(item.ActKind)
	if name == "" {
		return browserCompatDescriptor{}, false
	}
	item.Name = name
	return item, true
}

func browserCompatDescriptorForTool(name string) (browserCompatDescriptor, bool) {
	descriptor, ok := browserCompatDescriptorsByName[NormalizeToolName(name)]
	return descriptor, ok
}

func browserCompatMetadataForTool(name string) (ToolMetadata, bool) {
	descriptor, ok := browserCompatDescriptorForTool(name)
	if !ok {
		return ToolMetadata{}, false
	}
	meta := ToolMetadata{
		Groups:       []string{BrowserSurfaceCompat},
		Capabilities: append([]string{"browser_deprecated_compat_wrapper", "browser_migration_fallback"}, descriptor.MetadataCapabilities...),
		AuditTags:    append([]string{"deprecated_compatibility"}, descriptor.MetadataAuditTags...),
	}
	if descriptor.ReadOnly {
		meta.ReadOnly = builtinToolMetadataBoolPtr(true)
	} else if descriptor.LikelySideEffect {
		meta.ReadOnly = builtinToolMetadataBoolPtr(false)
		meta.ConcurrencySafe = builtinToolMetadataBoolPtr(false)
	}
	return meta, true
}

func browserCompatManagedOptInActKind(name string) string {
	return agentxbrowserruntime.BrowserCompatActKindForToolName(name)
}

func browserCompatToolForManagedOptInActKind(kind string) string {
	return agentxbrowserruntime.BrowserCompatToolNameForActKind(kind)
}

func browserCompatIsReadOnly(name string) bool {
	descriptor, ok := browserCompatDescriptorForTool(name)
	return ok && descriptor.ReadOnly
}

func browserCompatIsLikelySideEffect(name string) bool {
	descriptor, ok := browserCompatDescriptorForTool(name)
	return ok && descriptor.LikelySideEffect
}

func browserCompatBuiltinRiskLevel(name string) (RiskLevel, bool) {
	descriptor, ok := browserCompatDescriptorForTool(name)
	if !ok {
		return RiskUnknown, false
	}
	if descriptor.BuiltinRisk == RiskUnknown {
		return RiskUnknown, false
	}
	return descriptor.BuiltinRisk, true
}

func browserCompatArtifactSource(name string) string {
	descriptor, ok := browserCompatDescriptorForTool(name)
	if !ok {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(descriptor.ArtifactSource))
}

func browserCompatArtifactKind(name string) string {
	descriptor, ok := browserCompatDescriptorForTool(name)
	if !ok {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(descriptor.ArtifactKind))
}

func browserCompatSkipsInteractiveApprovalPrompt(name string, params map[string]any) bool {
	if _, ok := browserCompatDescriptorForTool(name); !ok {
		return false
	}
	actKind := browserCompatManagedOptInActKind(name)
	switch actKind {
	case "extract", "screenshot":
		return true
	case "list_tabs":
		action := strings.ToLower(strings.TrimSpace(firstString(params, "action")))
		return action == "" || action == "list"
	default:
		return false
	}
}

func browserCompatToolName(name string) string {
	descriptor, ok := browserCompatDescriptorForTool(name)
	if !ok {
		return ""
	}
	return descriptor.Name
}

func browserCompatEventSource(name string) string {
	return browserCompatToolName(name)
}

func browserCompatActor(name string, action string) string {
	base := browserCompatToolName(name)
	if base == "" {
		return ""
	}
	if browserCompatManagedOptInActKind(base) != "list_tabs" {
		return base
	}
	action = browserNormalizeToolToken(action)
	if action == "" {
		return base
	}
	return base + " " + action
}

func browserArtifactSourceForTool(name string) string {
	switch NormalizeToolName(name) {
	case "browser_act", "browser_runtime":
		return "browser"
	default:
		return browserCompatArtifactSource(name)
	}
}

func browserCompatForceConfirmationNeedsGuardianReview(name string, params map[string]any) bool {
	if !firstBool(params, "force") {
		return false
	}
	descriptor, ok := browserCompatDescriptorForTool(name)
	if !ok {
		return false
	}
	if descriptor.ForceConfirmationGuardianReview {
		return true
	}
	action := browserNormalizeToolToken(firstString(params, "action", "operation", "mode"))
	if browserCompatActionListed(action, descriptor.ForceConfirmationGuardianActions) {
		return true
	}
	if browserCompatActionListed(action, descriptor.ForceConfirmationRememberActions) {
		return firstBool(params, "remember_target", "remember")
	}
	return false
}

func browserCompatPendingTargetReviewReason(name string, action string, state browserPendingTargetReviewState, force bool) string {
	actor := browserCompatActor(name, action)
	if actor == "" {
		actor = NormalizeToolName(name)
	}
	return browserPendingTargetReviewReasonWithState(actor, state, force)
}

func browserCompatImplicitLegacyHostFallbackError(
	name string,
	hiddenImplicitHostDefaultBase bool,
	explicitRuntimeTarget bool,
	runtimeInfo BrowserRuntimeInfo,
	target browserToolTarget,
	action string,
	requestURL string,
) error {
	base := browserCompatToolName(name)
	if base == "" {
		base = NormalizeToolName(name)
	}
	switch browserCompatManagedOptInActKind(base) {
	case "open", "navigate":
		return browserImplicitLegacyHostURLNavigationFallbackError(base, hiddenImplicitHostDefaultBase, explicitRuntimeTarget, runtimeInfo, target, requestURL)
	case "list_tabs":
		return browserImplicitLegacyHostTabsActionFallbackError(base, hiddenImplicitHostDefaultBase, runtimeInfo, action, target)
	default:
		return browserImplicitLegacyHostPageExecutionFallbackError(base, hiddenImplicitHostDefaultBase, explicitRuntimeTarget, runtimeInfo, target, requestURL)
	}
}

func browserCompatActionListed(action string, allow []string) bool {
	if action == "" {
		return false
	}
	for _, item := range allow {
		if browserNormalizeToolToken(item) == action {
			return true
		}
	}
	return false
}
