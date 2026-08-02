package browserd

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

func TestNewManagerValidatesExplicitPlan(t *testing.T) {
	tests := []struct {
		name string
		plan Plan
		want string
	}{
		{name: "disabled", plan: Plan{}, want: "managed browserd plan is disabled"},
		{name: "endpoint", plan: Plan{Enabled: true}, want: "managed browserd endpoint is required"},
		{name: "command", plan: Plan{Enabled: true, Endpoint: "http://127.0.0.1:43123"}, want: "managed browserd command is required"},
		{name: "bundled_state", plan: Plan{Enabled: true, Endpoint: "http://127.0.0.1:43123", Command: bundledBrowserdCommandAuto}, want: "state root is required"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewManager(ManagerOptions{Plan: tc.plan})
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("NewManager error = %v, want token %q", err, tc.want)
			}
		})
	}
}

func TestManagerProbeUsesExplicitPortAndChecksOwnership(t *testing.T) {
	root := t.TempDir()
	plan := Plan{
		Enabled:       true,
		Command:       os.Args[0],
		Endpoint:      "http://127.0.0.1:43123",
		StateRoot:     root,
		ProfilesRoot:  filepath.Join(root, "profiles"),
		ArtifactsRoot: filepath.Join(root, "artifacts"),
		LogsRoot:      filepath.Join(root, "logs"),
	}
	probeCalls := 0
	manager, err := NewManager(ManagerOptions{
		Plan: plan,
		Probe: func(_ context.Context, got Plan, timeout int) (agentxbrowserruntime.BrowserProfileStatusResult, error) {
			probeCalls++
			if got.Endpoint != plan.Endpoint || timeout != 2500 {
				t.Fatalf("probe input = %#v timeout=%d", got, timeout)
			}
			return agentxbrowserruntime.BrowserProfileStatusResult{
				Backend:       "managed-browserd",
				Running:       true,
				Connected:     true,
				StateRoot:     plan.StateRoot,
				ProfilesRoot:  plan.ProfilesRoot,
				ArtifactsRoot: plan.ArtifactsRoot,
				LogsRoot:      plan.LogsRoot,
			}, nil
		},
		TransportTimeout: 2500,
	})
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	status, err := manager.Probe(context.Background())
	if err != nil || !status.Running || probeCalls != 1 {
		t.Fatalf("Probe status=%#v calls=%d err=%v", status, probeCalls, err)
	}

	manager.probe = func(context.Context, Plan, int) (agentxbrowserruntime.BrowserProfileStatusResult, error) {
		return agentxbrowserruntime.BrowserProfileStatusResult{StateRoot: filepath.Join(root, "foreign")}, nil
	}
	if _, err := manager.Probe(context.Background()); err == nil || !strings.Contains(err.Error(), "different state root") {
		t.Fatalf("ownership mismatch error = %v", err)
	}
}

func TestManagerFailsClosedWithoutProbeAndAfterClose(t *testing.T) {
	manager, err := NewManager(ManagerOptions{Plan: Plan{
		Enabled:  true,
		Command:  os.Args[0],
		Endpoint: "http://127.0.0.1:43123",
	}})
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	if _, err := manager.Probe(context.Background()); err == nil || !strings.Contains(err.Error(), "status probe is required") {
		t.Fatalf("Probe error = %v", err)
	}
	if err := manager.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if err := manager.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
	if err := manager.EnsureStarted(context.Background()); err == nil || !strings.Contains(err.Error(), "manager is closed") {
		t.Fatalf("EnsureStarted after Close error = %v", err)
	}
}

func TestBundledAssetsStayEmbeddedWithoutConstructorSideEffects(t *testing.T) {
	root := filepath.Join(t.TempDir(), "state")
	manager, err := NewManager(ManagerOptions{Plan: Plan{
		Enabled:   true,
		Command:   bundledBrowserdCommandAuto,
		Endpoint:  "http://127.0.0.1:43123",
		StateRoot: root,
	}})
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	if _, err := os.Stat(root); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("constructor materialized state root: %v", err)
	}
	for _, name := range []string{bundledBrowserdEntry, "package.json", "package-lock.json"} {
		blob, err := bundledBrowserdFiles.ReadFile(filepath.ToSlash(filepath.Join("node", name)))
		if err != nil || len(blob) == 0 {
			t.Fatalf("read bundled asset %s: bytes=%d err=%v", name, len(blob), err)
		}
	}
	_ = manager.Close()
}

func TestCapabilitiesForBundledPlanStayNarrow(t *testing.T) {
	got := CapabilitiesForNodeBackendPlan(NodeBackendPlan{
		Source:  NodeBackendSourceManagedBrowser,
		Managed: Plan{Enabled: true, Command: bundledBrowserdCommandAuto},
	}, agentxbrowserruntime.BrowserCapabilities{})
	if !got.Open || !got.Navigate || !got.Click || !got.RuntimeStatus || !got.RuntimeStart {
		t.Fatalf("bundled capabilities = %#v", got)
	}
	if got.Console || got.Requests || got.Resize || got.Cookies || got.Credentials {
		t.Fatalf("bundled fallback widened privileged capabilities: %#v", got)
	}
}
