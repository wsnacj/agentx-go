package tools

import "testing"

func TestBrowserSubstratePosture(t *testing.T) {
	tests := []struct {
		name       string
		backend    string
		target     string
		wantPost   string
		wantStatus string
		wantReason string
	}{
		{
			name:       "legacy system host",
			backend:    "system",
			target:     "host",
			wantPost:   BrowserSubstrateLegacySystemHost,
			wantStatus: "warn",
			wantReason: "default browser route still reflects the legacy system host backend (`system` + `host`), so interactive capabilities depend on the local Safari/system path until an explicit runtime_target or promoted managed route is used",
		},
		{
			name:       "proxy node",
			backend:    "proxy",
			target:     "node",
			wantPost:   BrowserSubstrateNodeRuntime,
			wantStatus: "ok",
			wantReason: "default browser execution resolves to node runtime backend `proxy`",
		},
		{
			name:       "sandbox runtime",
			backend:    "proxy",
			target:     "sandbox",
			wantPost:   BrowserSubstrateSandboxRuntime,
			wantStatus: "ok",
			wantReason: "default browser execution resolves to sandbox runtime backend `proxy`",
		},
		{
			name:       "future host runtime",
			backend:    "cdp",
			target:     "host",
			wantPost:   BrowserSubstrateHostRuntime,
			wantStatus: "ok",
			wantReason: "default browser execution resolves to host runtime backend `cdp`",
		},
		{
			name:       "custom backend",
			backend:    "custom",
			target:     "host",
			wantPost:   BrowserSubstrateCustomBackend,
			wantStatus: "warn",
			wantReason: "default browser execution resolves to a custom backend; verify its route and capability contract separately",
		},
		{
			name:       "custom backend variant",
			backend:    "custom-playwright",
			target:     "host",
			wantPost:   BrowserSubstrateCustomBackend,
			wantStatus: "warn",
			wantReason: "default browser execution resolves to a custom backend; verify its route and capability contract separately",
		},
		{
			name:       "empty",
			wantPost:   "",
			wantStatus: "",
			wantReason: "",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := BrowserSubstratePosture(tc.backend, tc.target); got != tc.wantPost {
				t.Fatalf("BrowserSubstratePosture() = %q, want %q", got, tc.wantPost)
			}
			if got := BrowserSubstrateStatus(tc.backend, tc.target); got != tc.wantStatus {
				t.Fatalf("BrowserSubstrateStatus() = %q, want %q", got, tc.wantStatus)
			}
			if got := BrowserSubstrateReason(tc.backend, tc.target); got != tc.wantReason {
				t.Fatalf("BrowserSubstrateReason() = %q, want %q", got, tc.wantReason)
			}
		})
	}
}

func TestBrowserSubstrateSelectionStrategy(t *testing.T) {
	t.Run("promoted node over legacy host", func(t *testing.T) {
		got := BrowserSubstrateSelectionStrategy(
			BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
			BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		)
		if got != BrowserSubstrateSelectionPreferNodeOverLegacy {
			t.Fatalf("BrowserSubstrateSelectionStrategy() = %q, want %q", got, BrowserSubstrateSelectionPreferNodeOverLegacy)
		}
		reason := BrowserSubstrateSelectionReason(
			BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
			BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		)
		if reason == "" {
			t.Fatalf("expected non-empty selection reason")
		}
	})

	t.Run("legacy host remains default", func(t *testing.T) {
		got := BrowserSubstrateSelectionStrategy(
			BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
			BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		)
		if got != BrowserSubstrateSelectionLegacyHostDefault {
			t.Fatalf("BrowserSubstrateSelectionStrategy() = %q, want %q", got, BrowserSubstrateSelectionLegacyHostDefault)
		}
		reason := BrowserSubstrateSelectionReason(
			BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
			BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		)
		if reason == "" || reason != "default browser route still reflects the legacy system host path because no promoted runtime route is configured, so targetless execution requires explicit runtime_target" {
			t.Fatalf("unexpected legacy-host default selection reason: %q", reason)
		}
	})

	t.Run("modern host runtime stays default", func(t *testing.T) {
		got := BrowserSubstrateSelectionStrategy(
			BrowserRuntimeInfo{Backend: "cdp", Profile: "default", Target: "host"},
			BrowserRuntimeInfo{Backend: "cdp", Profile: "default", Target: "host"},
		)
		if got != BrowserSubstrateSelectionPreferHostRuntime {
			t.Fatalf("BrowserSubstrateSelectionStrategy() = %q, want %q", got, BrowserSubstrateSelectionPreferHostRuntime)
		}
	})

	t.Run("sandbox runtime becomes default", func(t *testing.T) {
		got := BrowserSubstrateSelectionStrategy(
			BrowserRuntimeInfo{Backend: "sandbox", Profile: "default", Target: "sandbox"},
			BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		)
		if got != BrowserSubstrateSelectionPreferSandboxRuntime {
			t.Fatalf("BrowserSubstrateSelectionStrategy() = %q, want %q", got, BrowserSubstrateSelectionPreferSandboxRuntime)
		}
	})

	t.Run("sandbox ready stays explicit while legacy host remains default", func(t *testing.T) {
		reason := BrowserSubstrateSelectionReasonWithPromotionState(
			BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
			BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
			false,
			false,
			true,
			true,
		)
		if reason != "default browser route keeps the legacy system host path because sandbox remains an explicit managed lane until the browserd/node default strategy expands, so targetless execution still requires explicit runtime_target" {
			t.Fatalf("unexpected sandbox-explicit selection reason: %q", reason)
		}
	})

	t.Run("explicit custom host stays default", func(t *testing.T) {
		got := BrowserSubstrateSelectionStrategy(
			BrowserRuntimeInfo{Backend: "custom-playwright", Profile: "default", Target: "host"},
			BrowserRuntimeInfo{Backend: "custom-playwright", Profile: "default", Target: "host"},
		)
		if got != BrowserSubstrateSelectionPreferCustomBackend {
			t.Fatalf("BrowserSubstrateSelectionStrategy() = %q, want %q", got, BrowserSubstrateSelectionPreferCustomBackend)
		}
	})
}
