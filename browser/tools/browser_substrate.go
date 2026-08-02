package tools

import "fmt"

const (
	BrowserSubstrateLegacySystemHost = "legacy_system_host"
	BrowserSubstrateHostRuntime      = "host_runtime"
	BrowserSubstrateNodeRuntime      = "node_runtime"
	BrowserSubstrateSandboxRuntime   = "sandbox_runtime"
	BrowserSubstrateCustomBackend    = "custom_backend"

	BrowserSubstrateSelectionLegacyHostDefault    = "legacy_host_default"
	BrowserSubstrateSelectionPreferNodeOverLegacy = "prefer_node_over_legacy_host"
	BrowserSubstrateSelectionPreferNodeRuntime    = "prefer_node_runtime"
	BrowserSubstrateSelectionPreferHostRuntime    = "prefer_host_runtime"
	BrowserSubstrateSelectionPreferSandboxRuntime = "prefer_sandbox_runtime"
	BrowserSubstrateSelectionPreferCustomBackend  = "prefer_custom_backend"
)

func BrowserSubstratePosture(backend string, target string) string {
	info := normalizeBrowserRuntimeInfo(BrowserRuntimeInfo{Backend: backend, Target: target})
	info.Backend = browserRuntimeCanonicalBackend(info.Backend)
	switch {
	case info.Backend == "" && info.Target == "":
		return ""
	case info.Backend == "custom":
		return BrowserSubstrateCustomBackend
	case info.Target == "node":
		return BrowserSubstrateNodeRuntime
	case info.Target == "sandbox":
		return BrowserSubstrateSandboxRuntime
	case info.Target == "host" && info.Backend == "system":
		return BrowserSubstrateLegacySystemHost
	case info.Target == "host":
		return BrowserSubstrateHostRuntime
	default:
		return ""
	}
}

func BrowserSubstrateStatus(backend string, target string) string {
	switch BrowserSubstratePosture(backend, target) {
	case BrowserSubstrateLegacySystemHost, BrowserSubstrateCustomBackend:
		return "warn"
	case BrowserSubstrateHostRuntime, BrowserSubstrateNodeRuntime, BrowserSubstrateSandboxRuntime:
		return "ok"
	default:
		return ""
	}
}

func BrowserSubstrateReason(backend string, target string) string {
	info := normalizeBrowserRuntimeInfo(BrowserRuntimeInfo{Backend: backend, Target: target})
	switch BrowserSubstratePosture(info.Backend, info.Target) {
	case BrowserSubstrateLegacySystemHost:
		return "default browser route still reflects the legacy system host backend (`system` + `host`), so interactive capabilities depend on the local Safari/system path until an explicit runtime_target or promoted managed route is used"
	case BrowserSubstrateHostRuntime:
		return fmt.Sprintf("default browser execution resolves to host runtime backend `%s`", info.Backend)
	case BrowserSubstrateNodeRuntime:
		return fmt.Sprintf("default browser execution resolves to node runtime backend `%s`", info.Backend)
	case BrowserSubstrateSandboxRuntime:
		return fmt.Sprintf("default browser execution resolves to sandbox runtime backend `%s`", info.Backend)
	case BrowserSubstrateCustomBackend:
		return "default browser execution resolves to a custom backend; verify its route and capability contract separately"
	default:
		return ""
	}
}

func BrowserSubstrateSelectionStrategy(defaultInfo BrowserRuntimeInfo, hostInfo BrowserRuntimeInfo) string {
	defaultInfo = normalizeBrowserRuntimeInfo(defaultInfo)
	hostInfo = normalizeBrowserRuntimeInfo(hostInfo)
	defaultPosture := BrowserSubstratePosture(defaultInfo.Backend, defaultInfo.Target)
	hostPosture := BrowserSubstratePosture(hostInfo.Backend, hostInfo.Target)
	switch {
	case defaultPosture == BrowserSubstrateNodeRuntime && hostPosture == BrowserSubstrateLegacySystemHost:
		return BrowserSubstrateSelectionPreferNodeOverLegacy
	case defaultPosture == BrowserSubstrateNodeRuntime:
		return BrowserSubstrateSelectionPreferNodeRuntime
	case defaultPosture == BrowserSubstrateHostRuntime:
		return BrowserSubstrateSelectionPreferHostRuntime
	case defaultPosture == BrowserSubstrateLegacySystemHost:
		return BrowserSubstrateSelectionLegacyHostDefault
	case defaultPosture == BrowserSubstrateSandboxRuntime:
		return BrowserSubstrateSelectionPreferSandboxRuntime
	case defaultPosture == BrowserSubstrateCustomBackend:
		return BrowserSubstrateSelectionPreferCustomBackend
	default:
		return ""
	}
}

func BrowserSubstrateSelectionReason(defaultInfo BrowserRuntimeInfo, hostInfo BrowserRuntimeInfo) string {
	defaultInfo = normalizeBrowserRuntimeInfo(defaultInfo)
	hostInfo = normalizeBrowserRuntimeInfo(hostInfo)
	switch BrowserSubstrateSelectionStrategy(defaultInfo, hostInfo) {
	case BrowserSubstrateSelectionPreferNodeOverLegacy:
		return "default browser execution promotes to node runtime because the host substrate is still the legacy system host path"
	case BrowserSubstrateSelectionPreferNodeRuntime:
		return fmt.Sprintf("default browser execution prefers node runtime backend `%s`", defaultInfo.Backend)
	case BrowserSubstrateSelectionPreferHostRuntime:
		return fmt.Sprintf("default browser execution stays on host runtime backend `%s`", defaultInfo.Backend)
	case BrowserSubstrateSelectionLegacyHostDefault:
		return "default browser route still reflects the legacy system host path because no promoted runtime route is configured, so targetless execution requires explicit runtime_target"
	case BrowserSubstrateSelectionPreferSandboxRuntime:
		return fmt.Sprintf("default browser execution prefers sandbox runtime backend `%s`", defaultInfo.Backend)
	case BrowserSubstrateSelectionPreferCustomBackend:
		return "default browser execution prefers a custom backend; verify its route and capability contract separately"
	default:
		return ""
	}
}

func BrowserSubstrateSelectionReasonWithPromotionState(defaultInfo BrowserRuntimeInfo, hostInfo BrowserRuntimeInfo, nodeConfigured bool, nodePromotionReady bool, sandboxConfigured bool, sandboxPromotionReady bool) string {
	defaultInfo = normalizeBrowserRuntimeInfo(defaultInfo)
	hostInfo = normalizeBrowserRuntimeInfo(hostInfo)
	if BrowserSubstrateSelectionStrategy(defaultInfo, hostInfo) == BrowserSubstrateSelectionLegacyHostDefault &&
		nodeConfigured && !nodePromotionReady &&
		BrowserSubstratePosture(hostInfo.Backend, hostInfo.Target) == BrowserSubstrateLegacySystemHost {
		return "default browser route keeps the legacy system host path because the configured node runtime does not yet advertise the required default browser capabilities, so targetless execution still requires explicit runtime_target"
	}
	if BrowserSubstrateSelectionStrategy(defaultInfo, hostInfo) == BrowserSubstrateSelectionLegacyHostDefault &&
		sandboxConfigured && !sandboxPromotionReady &&
		BrowserSubstratePosture(hostInfo.Backend, hostInfo.Target) == BrowserSubstrateLegacySystemHost {
		return "default browser route keeps the legacy system host path because the configured sandbox runtime does not yet advertise the required default browser capabilities, so targetless execution still requires explicit runtime_target"
	}
	if BrowserSubstrateSelectionStrategy(defaultInfo, hostInfo) == BrowserSubstrateSelectionLegacyHostDefault &&
		!nodeConfigured &&
		sandboxConfigured && sandboxPromotionReady &&
		BrowserSubstratePosture(hostInfo.Backend, hostInfo.Target) == BrowserSubstrateLegacySystemHost {
		return "default browser route keeps the legacy system host path because sandbox remains an explicit managed lane until the browserd/node default strategy expands, so targetless execution still requires explicit runtime_target"
	}
	return BrowserSubstrateSelectionReason(defaultInfo, hostInfo)
}
