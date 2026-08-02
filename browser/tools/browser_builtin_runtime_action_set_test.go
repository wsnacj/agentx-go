package tools

import (
	"context"
	"strings"
	"testing"

	types "github.com/wsnacj/agentx-go/components/llm"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

func TestRegisterBrowserTools_RuntimeRejectsUnsupportedActionWithoutReprobingAvailableActions(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &countingCapabilityRuntimeControlBrowserBackend{
		runtimeControlBrowserBackend: &runtimeControlBrowserBackend{
			runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
				fakeBrowserBackend: &fakeBrowserBackend{},
				runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
			},
		},
		capabilities: BrowserCapabilities{},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_runtime"},
	})

	capabilityCallsAfterRegister := backend.capabilityCalls
	if capabilityCallsAfterRegister == 0 {
		t.Fatalf("expected browser_runtime registration to inspect available actions at least once")
	}

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"hover"}`,
	})
	if err == nil || !strings.Contains(err.Error(), "action must be one of") {
		t.Fatalf("expected browser_runtime to reject unsupported hover action early, got %v", err)
	}
	if backend.capabilityCalls != capabilityCallsAfterRegister {
		t.Fatalf("expected unsupported runtime action validation to reuse registration-time available actions, before=%d after=%d", capabilityCallsAfterRegister, backend.capabilityCalls)
	}
}

func TestRegisterBrowserTools_BrowserUnifiedReusesRuntimeActionSetAtRegistration(t *testing.T) {
	newBackend := func() *countingCapabilityRuntimeControlBrowserBackend {
		return &countingCapabilityRuntimeControlBrowserBackend{
			runtimeControlBrowserBackend: &runtimeControlBrowserBackend{
				runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
					fakeBrowserBackend: &fakeBrowserBackend{},
					runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
				},
			},
			capabilities: BrowserCapabilities{},
		}
	}

	runtimeBackend := newBackend()
	RegisterBrowserTools(llmxtools.NewRegistry(), BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      runtimeBackend,
		EnabledTools: []string{"browser_runtime"},
	})
	if runtimeBackend.capabilityCalls == 0 {
		t.Fatalf("expected browser_runtime registration to inspect runtime action capabilities")
	}

	unifiedBackend := newBackend()
	RegisterBrowserTools(llmxtools.NewRegistry(), BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      unifiedBackend,
		EnabledTools: []string{"browser"},
	})
	if unifiedBackend.capabilityCalls == 0 {
		t.Fatalf("expected unified browser registration to inspect runtime action capabilities")
	}
	if unifiedBackend.capabilityCalls != runtimeBackend.capabilityCalls {
		t.Fatalf("expected unified browser registration to reuse cached runtime action set, runtime-only=%d unified=%d", runtimeBackend.capabilityCalls, unifiedBackend.capabilityCalls)
	}
}
