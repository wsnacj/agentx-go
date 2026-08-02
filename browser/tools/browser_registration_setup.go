package tools

import (
	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

type browserToolRegistration struct {
	name     string
	register func(browserRegistrationContext)
}

var browserToolRegistrations = browserToolRegistrationsList()

func browserToolRegistrationsList() []browserToolRegistration {
	out := []browserToolRegistration{
		{name: "browser_runtime", register: registerBrowserRuntimeTool},
	}
	for _, descriptor := range browserCompatDescriptors {
		var ok bool
		descriptor, ok = browserCompatDescriptorWithResolvedName(descriptor)
		if !ok {
			continue
		}
		register := browserCompatRegistrationHandler(descriptor.Name)
		if register == nil {
			continue
		}
		out = append(out, browserToolRegistration{
			name:     descriptor.Name,
			register: register,
		})
	}
	out = append(out,
		browserToolRegistration{name: "browser_act", register: registerBrowserActTool},
		browserToolRegistration{name: "browser", register: registerBrowserUnifiedTool},
	)
	return out
}

func browserCompatRegistrationHandler(name string) func(browserRegistrationContext) {
	switch browserCompatManagedOptInActKind(name) {
	case "open":
		return registerBrowserOpenTool
	case "navigate":
		return registerBrowserNavigateTool
	case "list_tabs":
		return registerBrowserTabsTool
	case "extract":
		return registerBrowserExtractTool
	case "screenshot":
		return registerBrowserScreenshotTool
	case "click":
		return registerBrowserClickTool
	case "type":
		return registerBrowserTypeTool
	case "evaluate":
		return registerBrowserEvalTool
	default:
		return nil
	}
}

func newBrowserRegistrationContext(reg *llmxtools.Registry, opts BrowserToolOptions) (browserRegistrationContext, map[string]bool, bool) {
	if reg == nil {
		return browserRegistrationContext{}, nil, false
	}
	enabled := buildEnabledToolSet(opts.EnabledTools)
	timeoutMs := opts.TimeoutMs
	if timeoutMs <= 0 {
		timeoutMs = 20_000
	}
	maxChars := opts.MaxChars
	if maxChars <= 0 {
		maxChars = 40_000
	}
	openWaitMs := opts.OpenWaitMs
	if openWaitMs < 0 {
		openWaitMs = 0
	}
	if openWaitMs == 0 {
		openWaitMs = 4_000
	}
	screenshotWaitMs := opts.ScreenshotWaitMs
	if screenshotWaitMs < 0 {
		screenshotWaitMs = 0
	}
	if screenshotWaitMs == 0 {
		screenshotWaitMs = 2_500
	}
	policy, err := newOutboundNetworkPolicy(outboundNetworkOptions{
		AllowPrivateHosts: opts.AllowPrivateHosts,
		AllowCIDRs:        append([]string(nil), opts.AllowCIDRs...),
		DenyCIDRs:         append([]string(nil), opts.DenyCIDRs...),
		AllowPorts:        append([]int(nil), opts.AllowPorts...),
		DenyPorts:         append([]int(nil), opts.DenyPorts...),
		DefaultDenyCIDRs:  defaultOutboundDeniedCIDRs,
		DefaultDenyPorts:  defaultOutboundDeniedPorts,
	})
	if err != nil {
		return browserRegistrationContext{}, nil, false
	}
	preview := browserDefaultRuntimePreviewForDispatchOptions(opts, policy, timeoutMs)
	sessionRegistry := browserSessionRegistryForOptions(opts)
	ctx := browserRegistrationContext{
		reg:                  reg,
		opts:                 opts,
		enabledTools:         enabled,
		capabilities:         browserCapabilitiesForRegistrationWithBackend(preview.EffectiveBackend, preview.SubstrateAssessment),
		substrateAssessment:  preview.SubstrateAssessment,
		substrateSummary:     preview.SubstrateSummary,
		sessionRegistry:      sessionRegistry,
		sessionRunRegistry:   opts.SessionRunRegistry,
		sessionStateRegistry: opts.SessionStateRegistry,
		watchManagerProvider: agentxbrowserruntime.SharedSessionBrowserObserverManagerFor(
			sessionRegistry,
			opts.SessionRunRegistry,
			opts.SessionStateRegistry,
			browserRuntimeReconnectWatchdogWindow,
		),
		policy:           policy,
		backend:          preview.EffectiveBackend,
		timeoutMs:        timeoutMs,
		openWaitMs:       openWaitMs,
		screenshotWaitMs: screenshotWaitMs,
		maxChars:         maxChars,
		derivedCache:     newBrowserRegistrationDerivedCache(),
	}
	ctx.runtimeActions = browserRuntimeComputeAvailableActions(ctx)
	ctx.managedOptInProjection = browserManagedOptInProjectionForCapabilities(
		ctx.enabledTools,
		ctx.capabilities,
		ctx.opts.NodeBackend,
		ctx.opts.SandboxBackend,
	)
	ctx.managedOptInProjectionReady = true
	return ctx,
		enabled,
		true
}

func registerEnabledBrowserTools(ctx browserRegistrationContext, enabled map[string]bool) {
	browserEnabled := toolEnabled(enabled, "browser")
	for _, item := range browserToolRegistrations {
		if !toolEnabled(enabled, item.name) &&
			!(browserEnabled && (item.name == "browser_runtime" || item.name == "browser_act")) {
			continue
		}
		if !ctx.capabilities.SupportsTool(item.name) &&
			!browserRegistrationSupportsExplicitManagedCompatTool(ctx, enabled, item.name) &&
			!(item.name == "browser_act" && browserRegistrationSupportsExplicitManagedActTool(ctx, enabled)) {
			continue
		}
		item.register(ctx)
	}
}
