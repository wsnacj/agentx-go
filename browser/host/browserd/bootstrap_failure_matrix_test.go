package browserd

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestResolveManagerLaunchClassifiesMissingNodeWithoutMaterializingState(t *testing.T) {
	emptyBin := filepath.Join(t.TempDir(), "empty-bin")
	if err := os.MkdirAll(emptyBin, 0o755); err != nil {
		t.Fatalf("mkdir empty bin: %v", err)
	}
	t.Setenv("PATH", emptyBin)
	stateRoot := filepath.Join(t.TempDir(), "state")
	_, _, err := resolveManagerLaunchContext(context.Background(), Plan{
		Command:   bundledBrowserdCommandAuto,
		StateRoot: stateRoot,
	}, time.Second)
	assertBundledBootstrapFailureCode(t, err, playwrightBootstrapErrorNodeMissing)
	if _, statErr := os.Stat(stateRoot); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("missing node materialized state root, stat error = %v", statErr)
	}
}

func TestBootstrapBundledPlaywrightBrowserClassifiesNetworkFailureWithoutLeakingOutput(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script node stub is not portable to windows")
	}
	binDir := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir node stub directory: %v", err)
	}
	nodePath := filepath.Join(binDir, "node")
	const networkMarker = "getaddrinfo ENOTFOUND playwright.azureedge.net"
	stub := "#!/bin/sh\n" +
		"printf '%s\\n' " + shellQuote(networkMarker) + " >&2\n" +
		"exit 19\n"
	if err := os.WriteFile(nodePath, []byte(stub), 0o755); err != nil {
		t.Fatalf("write node stub: %v", err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	targetRoot := filepath.Join(t.TempDir(), "node")
	if err := os.MkdirAll(filepath.Join(targetRoot, "node_modules", "playwright"), 0o755); err != nil {
		t.Fatalf("mkdir playwright fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(targetRoot, "node_modules", "playwright", "cli.js"), []byte("// fixture\n"), 0o644); err != nil {
		t.Fatalf("write playwright cli fixture: %v", err)
	}

	err := bootstrapBundledPlaywrightBrowserContext(context.Background(), targetRoot, t.TempDir())
	assertBundledBootstrapFailureCode(t, err, playwrightBootstrapErrorBrowserInstallNetwork)
	if strings.Contains(err.Error(), networkMarker) {
		t.Fatalf("network bootstrap error leaked command output: %q", err)
	}
	if !strings.Contains(err.Error(), "command_failed exit_code=19") {
		t.Fatalf("network bootstrap error lacks safe command summary: %q", err)
	}
}

func TestEnsureBundledPlaywrightBrowserRecordsMissingExecutable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script node stub is not portable to windows")
	}
	stateRoot := t.TempDir()
	cacheRoot := filepath.Join(t.TempDir(), "cache")
	t.Setenv("AGENTX_BROWSERD_PLAYWRIGHT_BROWSERS_PATH", cacheRoot)
	binDir := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir node stub directory: %v", err)
	}
	nodePath := filepath.Join(binDir, "node")
	stub := "#!/bin/sh\nexit 0\n"
	if err := os.WriteFile(nodePath, []byte(stub), 0o755); err != nil {
		t.Fatalf("write node stub: %v", err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	targetRoot := filepath.Join(t.TempDir(), "node")
	if err := os.MkdirAll(filepath.Join(targetRoot, "node_modules", "playwright"), 0o755); err != nil {
		t.Fatalf("mkdir playwright fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(targetRoot, "node_modules", "playwright", "cli.js"), []byte("// fixture\n"), 0o644); err != nil {
		t.Fatalf("write playwright cli fixture: %v", err)
	}

	err := ensureBundledPlaywrightBrowserAvailableContext(context.Background(), targetRoot, stateRoot)
	assertBundledBootstrapFailureCode(t, err, playwrightBootstrapErrorBrowserExecutableMissing)
	if !strings.Contains(err.Error(), "still unavailable") {
		t.Fatalf("missing browser error = %v, want fail-closed unavailable error", err)
	}
	if location := bundledPlaywrightBrowsersLocation(stateRoot); location.Pinned {
		payload := readBundledPlaywrightBrowserCachePinForLocation(location)
		if payload.BootstrapState != playwrightBootstrapStateFailed || payload.BootstrapErrorCode != playwrightBootstrapErrorBrowserExecutableMissing || payload.LaunchReady {
			t.Fatalf("missing browser bootstrap metadata = %#v", payload)
		}
	}
}

func TestMaterializeBundledBrowserdFilesClassifiesUnwritableStateRoot(t *testing.T) {
	stateRoot := filepath.Join(t.TempDir(), "state-root-file")
	if err := os.WriteFile(stateRoot, []byte("not a directory"), 0o600); err != nil {
		t.Fatalf("write invalid state root: %v", err)
	}
	_, err := materializeBundledBrowserdFiles(stateRoot)
	assertBundledBootstrapFailureCode(t, err, playwrightBootstrapErrorStateRootUnwritable)
	if strings.Contains(err.Error(), "not a directory") {
		t.Fatalf("state root error must expose stable code rather than platform detail: %q", err)
	}
}

func TestBundledPlaywrightBrowsersLocationPrefersExistingDefaultCache(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("user cache layout fixture is only covered on Unix-like hosts")
	}
	homeRoot := filepath.Join(t.TempDir(), "home")
	t.Setenv("HOME", homeRoot)
	t.Setenv("XDG_CACHE_HOME", filepath.Join(homeRoot, "cache"))
	t.Setenv("AGENTX_BROWSERD_PLAYWRIGHT_BROWSERS_PATH", "")
	cacheRoot, err := os.UserCacheDir()
	if err != nil {
		t.Fatalf("resolve user cache root: %v", err)
	}
	defaultRoot := filepath.Join(cacheRoot, "ms-playwright")
	if err := os.MkdirAll(defaultRoot, 0o755); err != nil {
		t.Fatalf("mkdir existing Playwright default cache: %v", err)
	}

	location := bundledPlaywrightBrowsersLocation(t.TempDir())
	if location.Path != defaultRoot || location.Source != playwrightCacheSourceDefault || location.Pinned {
		t.Fatalf("existing default cache location = %#v", location)
	}
}

func assertBundledBootstrapFailureCode(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("bootstrap error = nil, want code %q", want)
	}
	var bootstrapErr *bundledBootstrapError
	if !errors.As(err, &bootstrapErr) {
		t.Fatalf("bootstrap error type = %T (%v), want bundledBootstrapError", err, err)
	}
	if bootstrapErr.code != want {
		t.Fatalf("bootstrap error code = %q, want %q; error=%v", bootstrapErr.code, want, err)
	}
	if !strings.Contains(err.Error(), "code="+want) {
		t.Fatalf("bootstrap error does not expose stable code %q: %v", want, err)
	}
}
