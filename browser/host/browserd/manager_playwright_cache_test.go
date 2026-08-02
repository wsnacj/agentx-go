package browserd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestMaterializeBundledBrowserdFilesBootstrapsNodeModulesWhenSourceModulesUnavailable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script npm stub is not portable to windows")
	}
	stateRoot := t.TempDir()
	homeRoot := filepath.Join(t.TempDir(), "home")
	t.Setenv("HOME", homeRoot)
	binDir := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir npm stub dir: %v", err)
	}
	expectedBrowsersPath := bundledPlaywrightBrowsersPath(stateRoot)
	argsPath := filepath.Join(stateRoot, "npm-args.txt")
	skipPath := filepath.Join(stateRoot, "playwright-skip.txt")
	cachePath := filepath.Join(stateRoot, "npm-cache.txt")
	browserArgsPath := filepath.Join(stateRoot, "node-browser-install-args.txt")
	browserCachePath := filepath.Join(stateRoot, "node-browser-cache.txt")
	browserTimeoutPath := filepath.Join(stateRoot, "node-browser-timeout.txt")
	browserExecutablePath := filepath.Join(expectedBrowsersPath, "chromium-1187", "chrome")
	npmStub := filepath.Join(binDir, "npm")
	nodeStub := filepath.Join(binDir, "node")
	script := "#!/bin/sh\n" +
		"set -eu\n" +
		"printf '%s\\n' \"$@\" > " + shellQuote(argsPath) + "\n" +
		"printf '%s\\n' \"${PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD:-}\" > " + shellQuote(skipPath) + "\n" +
		"printf '%s\\n' \"${npm_config_cache:-}\" > " + shellQuote(cachePath) + "\n" +
		"mkdir -p \"$PWD/node_modules/playwright\"\n" +
		"printf '%s\\n' '{\"name\":\"playwright\"}' > \"$PWD/node_modules/playwright/package.json\"\n" +
		"printf '%s\\n' '#!/usr/bin/env node' > \"$PWD/node_modules/playwright/cli.js\"\n"
	if err := os.WriteFile(npmStub, []byte(script), 0o755); err != nil {
		t.Fatalf("write npm stub: %v", err)
	}
	nodeScript := "#!/bin/sh\n" +
		"set -eu\n" +
		"if [ \"${1:-}\" = \"-e\" ]; then\n" +
		"  printf '%s' " + shellQuote(browserExecutablePath) + "\n" +
		"  exit 0\n" +
		"fi\n" +
		"printf '%s\\n' \"$@\" > " + shellQuote(browserArgsPath) + "\n" +
		"printf '%s\\n' \"${PLAYWRIGHT_BROWSERS_PATH:-}\" > " + shellQuote(browserCachePath) + "\n" +
		"printf '%s\\n' \"${PLAYWRIGHT_DOWNLOAD_CONNECTION_TIMEOUT:-}\" > " + shellQuote(browserTimeoutPath) + "\n" +
		"mkdir -p " + shellQuote(filepath.Dir(browserExecutablePath)) + "\n" +
		"printf '%s\\n' '#!/bin/sh' > " + shellQuote(browserExecutablePath) + "\n" +
		"chmod +x " + shellQuote(browserExecutablePath) + "\n"
	if err := os.WriteFile(nodeStub, []byte(nodeScript), 0o755); err != nil {
		t.Fatalf("write node stub: %v", err)
	}

	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("AGENTX_BROWSERD_TEST_SKIP_SOURCE_NODE_MODULES", "1")
	t.Setenv("AGENTX_BROWSERD_SKIP_PLAYWRIGHT_BROWSER_BOOTSTRAP", "0")

	entryPath, err := materializeBundledBrowserdFiles(stateRoot)
	if err != nil {
		t.Fatalf("materialize bundled browserd files with bootstrap fallback: %v", err)
	}
	if _, err := os.Stat(entryPath); err != nil {
		t.Fatalf("stat materialized bundled browserd entry: %v", err)
	}
	targetRoot := filepath.Dir(entryPath)
	modulesPath := filepath.Join(targetRoot, "node_modules", "playwright", "package.json")
	if _, err := os.Stat(modulesPath); err != nil {
		t.Fatalf("expected bootstrap-installed playwright package.json: %v", err)
	}
	if _, ok, err := readExistingSymlink(filepath.Join(targetRoot, "node_modules")); err != nil {
		t.Fatalf("inspect materialized node_modules path: %v", err)
	} else if ok {
		t.Fatalf("expected bootstrap-installed node_modules directory, got symlink")
	}
	if raw, err := os.ReadFile(argsPath); err != nil {
		t.Fatalf("read npm args log: %v", err)
	} else if got := strings.Fields(string(raw)); !reflect.DeepEqual(got, []string{"ci", "--no-audit", "--no-fund"}) {
		t.Fatalf("unexpected npm args: %#v", got)
	}
	if raw, err := os.ReadFile(skipPath); err != nil {
		t.Fatalf("read playwright skip log: %v", err)
	} else if strings.TrimSpace(string(raw)) != "1" {
		t.Fatalf("expected PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1, got %q", string(raw))
	}
	if raw, err := os.ReadFile(cachePath); err != nil {
		t.Fatalf("read npm cache log: %v", err)
	} else if strings.TrimSpace(string(raw)) != filepath.Join(stateRoot, "bundled", ".npm-cache") {
		t.Fatalf("unexpected npm cache root: %q", string(raw))
	}
	if raw, err := os.ReadFile(browserArgsPath); err != nil {
		t.Fatalf("read playwright browser install args: %v", err)
	} else if got := strings.Fields(string(raw)); len(got) != 3 || !strings.HasSuffix(got[0], filepath.ToSlash("node_modules/playwright/cli.js")) || got[1] != "install" || got[2] != "chromium" {
		t.Fatalf("unexpected playwright browser install args: %#v", got)
	}
	if raw, err := os.ReadFile(browserCachePath); err != nil {
		t.Fatalf("read playwright browser cache env: %v", err)
	} else if strings.TrimSpace(string(raw)) != expectedBrowsersPath {
		t.Fatalf("unexpected playwright browser cache root: %q", string(raw))
	}
	if raw, err := os.ReadFile(browserTimeoutPath); err != nil {
		t.Fatalf("read playwright browser timeout env: %v", err)
	} else if strings.TrimSpace(string(raw)) != bundledPlaywrightDownloadConnectionTimeoutDefault {
		t.Fatalf("unexpected playwright browser download timeout: %q", string(raw))
	}
	if _, err := os.Stat(browserExecutablePath); err != nil {
		t.Fatalf("expected bootstrap-installed fake browser executable: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(expectedBrowsersPath, playwrightCachePinFilename))
	if err != nil {
		t.Fatalf("read playwright cache pin metadata: %v", err)
	}
	var payload struct {
		BundleGeneration            string   `json:"bundle_generation"`
		DependencyGeneration        string   `json:"dependency_generation"`
		BrowserRevision             string   `json:"browser_revision"`
		DeliveryGeneration          string   `json:"delivery_generation"`
		TargetDeliveryGeneration    string   `json:"target_delivery_generation"`
		LastReadyDeliveryGeneration string   `json:"last_ready_delivery_generation"`
		RetainedDeliveries          []string `json:"retained_delivery_generations"`
		RetainedDeliveryRevision    string   `json:"retained_delivery_browser_revision"`
		RetainedDeliveryReady       bool     `json:"retained_delivery_cache_ready"`
		RetainedFallbackDelivery    string   `json:"retained_fallback_delivery_generation"`
		RetainedFallbackPayload     bool     `json:"retained_fallback_payload_ready"`
		RetainedFallbackPayloadBR   string   `json:"retained_fallback_payload_block_reason"`
		RetainedFallbackPayloadSrc  string   `json:"retained_fallback_payload_source"`
		RetainedFallbackPayloadDirs []string `json:"retained_fallback_payload_dirs"`
		RetainedFallbackLaunch      bool     `json:"retained_fallback_launch_ready"`
		RetainedFallbackBlock       string   `json:"retained_fallback_launch_block_reason"`
		SelectedLaunchDelivery      string   `json:"selected_launch_delivery_generation"`
		SelectedLaunchSource        string   `json:"selected_launch_source"`
		SelectedLaunchReady         bool     `json:"selected_launch_ready"`
		SelectedLaunchBlockReason   string   `json:"selected_launch_block_reason"`
		SelectedLaunchRevision      string   `json:"selected_launch_browser_revision"`
		SelectedLaunchPayloadSrc    string   `json:"selected_launch_payload_source"`
		SelectedLaunchPayloadDirs   []string `json:"selected_launch_payload_dirs"`
		SelectedLaunchPayloadReady  bool     `json:"selected_launch_payload_ready"`
		SelectedLaunchPayloadBR     string   `json:"selected_launch_payload_block_reason"`
		SelectedLaunchExecutable    string   `json:"selected_launch_executable_path"`
		SelectedLaunchExecutableOK  bool     `json:"selected_launch_executable_ready"`
		SelectedLaunchExecutableBR  string   `json:"selected_launch_executable_block_reason"`
		LastEvictedDelivery         string   `json:"last_evicted_delivery_generation"`
		LaunchReady                 bool     `json:"launch_ready"`
		LaunchBlockReason           string   `json:"launch_block_reason"`
		BundleReady                 bool     `json:"bundle_ready"`
		DeliveryReady               bool     `json:"delivery_ready"`
		BootstrapState              string   `json:"bootstrap_state"`
		BootstrapErrorCode          string   `json:"bootstrap_error_code"`
		NodeModulesReady            bool     `json:"node_modules_ready"`
		BrowserReady                bool     `json:"browser_ready"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("decode playwright cache pin metadata: %v", err)
	}
	if payload.BundleGeneration != bundledBrowserdBundleGeneration() || payload.DependencyGeneration != bundledBrowserdDependencyGenerationID() {
		t.Fatalf("unexpected bootstrap generation metadata: %#v", payload)
	}
	if payload.BrowserRevision != "1187" || payload.DeliveryGeneration != bundledBrowserdValueGenerationHash(payload.BundleGeneration, payload.DependencyGeneration, "1187") {
		t.Fatalf("unexpected bootstrap delivery generation metadata: %#v", payload)
	}
	if payload.TargetDeliveryGeneration != payload.DeliveryGeneration || payload.LastReadyDeliveryGeneration != payload.DeliveryGeneration {
		t.Fatalf("unexpected bootstrap target/last-ready delivery metadata: %#v", payload)
	}
	if payload.RetainedDeliveryRevision != "1187" || !payload.RetainedDeliveryReady {
		t.Fatalf("unexpected bootstrap retained delivery availability metadata: %#v", payload)
	}
	if payload.RetainedFallbackDelivery != "" || payload.RetainedFallbackPayload || payload.RetainedFallbackPayloadBR != "" || payload.RetainedFallbackPayloadSrc != "" || len(payload.RetainedFallbackPayloadDirs) != 0 || payload.RetainedFallbackLaunch || payload.RetainedFallbackBlock != "" {
		t.Fatalf("unexpected bootstrap retained fallback metadata: %#v", payload)
	}
	if payload.SelectedLaunchDelivery != payload.DeliveryGeneration || payload.SelectedLaunchSource != "current_delivery" || !payload.SelectedLaunchReady || payload.SelectedLaunchBlockReason != "" {
		t.Fatalf("unexpected bootstrap selected launch metadata: %#v", payload)
	}
	if payload.SelectedLaunchRevision != "1187" || payload.SelectedLaunchPayloadSrc != "retained_dirs" || !reflect.DeepEqual(payload.SelectedLaunchPayloadDirs, []string{"chromium-1187"}) {
		t.Fatalf("unexpected bootstrap selected launch provenance metadata: %#v", payload)
	}
	if !payload.SelectedLaunchPayloadReady || payload.SelectedLaunchPayloadBR != "" {
		t.Fatalf("unexpected bootstrap selected launch payload readiness metadata: %#v", payload)
	}
	if payload.SelectedLaunchExecutable != browserExecutablePath || !payload.SelectedLaunchExecutableOK || payload.SelectedLaunchExecutableBR != "" {
		t.Fatalf("unexpected bootstrap selected launch executable metadata: %#v", payload)
	}
	if !reflect.DeepEqual(payload.RetainedDeliveries, []string{payload.DeliveryGeneration}) || payload.LastEvictedDelivery != "" {
		t.Fatalf("unexpected bootstrap delivery retention metadata: %#v", payload)
	}
	if !payload.LaunchReady || payload.LaunchBlockReason != "" {
		t.Fatalf("unexpected bootstrap launch gate metadata: %#v", payload)
	}
	if !payload.BundleReady || !payload.DeliveryReady {
		t.Fatalf("unexpected bootstrap readiness metadata: %#v", payload)
	}
	if payload.BootstrapState != playwrightBootstrapStateReady || payload.BootstrapErrorCode != "" || !payload.NodeModulesReady || !payload.BrowserReady {
		t.Fatalf("unexpected bootstrap success metadata: %#v", payload)
	}
}

func TestMaterializeBundledBrowserdFilesWritesBootstrapFailureWhenNPMMissing(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("PATH isolation is not portable to windows")
	}
	stateRoot := t.TempDir()
	homeRoot := filepath.Join(t.TempDir(), "home")
	emptyBin := filepath.Join(t.TempDir(), "empty-bin")
	t.Setenv("HOME", homeRoot)
	if err := os.MkdirAll(emptyBin, 0o755); err != nil {
		t.Fatalf("mkdir empty bin dir: %v", err)
	}
	t.Setenv("PATH", emptyBin)
	t.Setenv("AGENTX_BROWSERD_TEST_SKIP_SOURCE_NODE_MODULES", "1")
	t.Setenv("AGENTX_BROWSERD_SKIP_PLAYWRIGHT_BROWSER_BOOTSTRAP", "0")

	_, err := materializeBundledBrowserdFiles(stateRoot)
	if err == nil {
		t.Fatal("expected bundled browserd materialize to fail when npm is missing")
	}
	cacheRoot := bundledPlaywrightBrowsersPath(stateRoot)
	raw, readErr := os.ReadFile(filepath.Join(cacheRoot, playwrightCachePinFilename))
	if readErr != nil {
		t.Fatalf("read playwright cache pin metadata: %v", readErr)
	}
	var payload struct {
		BundleGeneration            string   `json:"bundle_generation"`
		DependencyGeneration        string   `json:"dependency_generation"`
		BrowserRevision             string   `json:"browser_revision"`
		DeliveryGeneration          string   `json:"delivery_generation"`
		TargetDeliveryGeneration    string   `json:"target_delivery_generation"`
		LastReadyDeliveryGeneration string   `json:"last_ready_delivery_generation"`
		RetainedDeliveries          []string `json:"retained_delivery_generations"`
		RetainedDeliveryRevision    string   `json:"retained_delivery_browser_revision"`
		RetainedDeliveryReady       bool     `json:"retained_delivery_cache_ready"`
		RetainedFallbackDelivery    string   `json:"retained_fallback_delivery_generation"`
		RetainedFallbackPayload     bool     `json:"retained_fallback_payload_ready"`
		RetainedFallbackPayloadBR   string   `json:"retained_fallback_payload_block_reason"`
		RetainedFallbackPayloadSrc  string   `json:"retained_fallback_payload_source"`
		RetainedFallbackPayloadDirs []string `json:"retained_fallback_payload_dirs"`
		RetainedFallbackLaunch      bool     `json:"retained_fallback_launch_ready"`
		RetainedFallbackBlock       string   `json:"retained_fallback_launch_block_reason"`
		SelectedLaunchDelivery      string   `json:"selected_launch_delivery_generation"`
		SelectedLaunchSource        string   `json:"selected_launch_source"`
		SelectedLaunchReady         bool     `json:"selected_launch_ready"`
		SelectedLaunchBlockReason   string   `json:"selected_launch_block_reason"`
		SelectedLaunchRevision      string   `json:"selected_launch_browser_revision"`
		SelectedLaunchPayloadSrc    string   `json:"selected_launch_payload_source"`
		SelectedLaunchPayloadDirs   []string `json:"selected_launch_payload_dirs"`
		SelectedLaunchPayloadReady  bool     `json:"selected_launch_payload_ready"`
		SelectedLaunchPayloadBR     string   `json:"selected_launch_payload_block_reason"`
		SelectedLaunchExecutable    string   `json:"selected_launch_executable_path"`
		SelectedLaunchExecutableOK  bool     `json:"selected_launch_executable_ready"`
		SelectedLaunchExecutableBR  string   `json:"selected_launch_executable_block_reason"`
		LastEvictedDelivery         string   `json:"last_evicted_delivery_generation"`
		LaunchReady                 bool     `json:"launch_ready"`
		LaunchBlockReason           string   `json:"launch_block_reason"`
		BundleReady                 bool     `json:"bundle_ready"`
		DeliveryReady               bool     `json:"delivery_ready"`
		BootstrapState              string   `json:"bootstrap_state"`
		BootstrapErrorCode          string   `json:"bootstrap_error_code"`
		NodeModulesReady            bool     `json:"node_modules_ready"`
		BrowserReady                bool     `json:"browser_ready"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("decode playwright cache pin metadata: %v", err)
	}
	if payload.BundleGeneration != bundledBrowserdBundleGeneration() || payload.DependencyGeneration != bundledBrowserdDependencyGenerationID() {
		t.Fatalf("unexpected bootstrap generation metadata: %#v", payload)
	}
	if payload.BrowserRevision != "" || payload.DeliveryGeneration != "" {
		t.Fatalf("unexpected bootstrap delivery generation metadata: %#v", payload)
	}
	if payload.TargetDeliveryGeneration != "" || payload.LastReadyDeliveryGeneration != "" {
		t.Fatalf("unexpected bootstrap target/last-ready delivery metadata: %#v", payload)
	}
	if payload.RetainedDeliveryRevision != "" || payload.RetainedDeliveryReady {
		t.Fatalf("unexpected bootstrap retained delivery availability metadata: %#v", payload)
	}
	if payload.RetainedFallbackDelivery != "" || payload.RetainedFallbackPayload || payload.RetainedFallbackPayloadBR != "" || payload.RetainedFallbackPayloadSrc != "" || len(payload.RetainedFallbackPayloadDirs) != 0 || payload.RetainedFallbackLaunch || payload.RetainedFallbackBlock != "" {
		t.Fatalf("unexpected bootstrap retained fallback metadata: %#v", payload)
	}
	if payload.SelectedLaunchDelivery != "" || payload.SelectedLaunchSource != "" || payload.SelectedLaunchReady || payload.SelectedLaunchBlockReason != playwrightBootstrapErrorNPMMissing {
		t.Fatalf("unexpected bootstrap selected launch failure metadata: %#v", payload)
	}
	if payload.SelectedLaunchRevision != "" || payload.SelectedLaunchPayloadSrc != "" || len(payload.SelectedLaunchPayloadDirs) != 0 {
		t.Fatalf("unexpected bootstrap selected launch failure provenance metadata: %#v", payload)
	}
	if payload.SelectedLaunchPayloadReady || payload.SelectedLaunchPayloadBR != "" {
		t.Fatalf("unexpected bootstrap selected launch failure payload readiness metadata: %#v", payload)
	}
	if payload.SelectedLaunchExecutable != "" || payload.SelectedLaunchExecutableOK || payload.SelectedLaunchExecutableBR != "" {
		t.Fatalf("unexpected bootstrap selected launch failure executable metadata: %#v", payload)
	}
	if len(payload.RetainedDeliveries) != 0 || payload.LastEvictedDelivery != "" {
		t.Fatalf("unexpected bootstrap delivery retention metadata: %#v", payload)
	}
	if payload.LaunchReady || payload.LaunchBlockReason != playwrightBootstrapErrorNPMMissing {
		t.Fatalf("unexpected bootstrap launch gate metadata: %#v", payload)
	}
	if !payload.BundleReady || payload.DeliveryReady {
		t.Fatalf("unexpected bootstrap readiness metadata: %#v", payload)
	}
	if payload.BootstrapState != playwrightBootstrapStateFailed || payload.BootstrapErrorCode != playwrightBootstrapErrorNPMMissing || payload.NodeModulesReady || payload.BrowserReady {
		t.Fatalf("unexpected bootstrap failure metadata: %#v", payload)
	}
}

func TestMaterializeBundledBrowserdFilesRebuildsStaleBundledNodeModulesGeneration(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script npm stub is not portable to windows")
	}
	stateRoot := t.TempDir()
	homeRoot := filepath.Join(t.TempDir(), "home")
	t.Setenv("HOME", homeRoot)
	binDir := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir stub bin dir: %v", err)
	}
	targetRoot := filepath.Join(stateRoot, "bundled", "node")
	if err := os.MkdirAll(filepath.Join(targetRoot, "node_modules", "playwright"), 0o755); err != nil {
		t.Fatalf("mkdir stale node_modules: %v", err)
	}
	if err := os.WriteFile(filepath.Join(targetRoot, "node_modules", "playwright", "package.json"), []byte(`{"name":"playwright"}`), 0o644); err != nil {
		t.Fatalf("write stale playwright package.json: %v", err)
	}
	staleMetadata := bundledBrowserdInstallMetadata{
		BundleGeneration:     "stale-bundle",
		DependencyGeneration: "stale-deps",
		UpdatedAtUnixMilli:   time.Now().Add(-time.Hour).UnixMilli(),
	}
	raw, err := json.MarshalIndent(staleMetadata, "", "  ")
	if err != nil {
		t.Fatalf("marshal stale install metadata: %v", err)
	}
	if err := os.WriteFile(filepath.Join(targetRoot, bundledBrowserdInstallMetadataFilename), raw, 0o644); err != nil {
		t.Fatalf("write stale install metadata: %v", err)
	}

	expectedBrowsersPath := bundledPlaywrightBrowsersPath(stateRoot)
	npmArgsPath := filepath.Join(stateRoot, "npm-args.txt")
	nodeArgsPath := filepath.Join(stateRoot, "node-args.txt")
	nodeCachePath := filepath.Join(stateRoot, "node-cache.txt")
	browserExecutablePath := filepath.Join(expectedBrowsersPath, "chromium-1187", "chrome")
	npmStub := filepath.Join(binDir, "npm")
	nodeStub := filepath.Join(binDir, "node")
	npmScript := "#!/bin/sh\n" +
		"set -eu\n" +
		"printf '%s\\n' \"$@\" > " + shellQuote(npmArgsPath) + "\n" +
		"mkdir -p \"$PWD/node_modules/playwright\"\n" +
		"printf '%s\\n' '{\"name\":\"playwright\"}' > \"$PWD/node_modules/playwright/package.json\"\n" +
		"printf '%s\\n' '#!/usr/bin/env node' > \"$PWD/node_modules/playwright/cli.js\"\n"
	if err := os.WriteFile(npmStub, []byte(npmScript), 0o755); err != nil {
		t.Fatalf("write npm stub: %v", err)
	}
	nodeScript := "#!/bin/sh\n" +
		"set -eu\n" +
		"if [ \"${1:-}\" = \"-e\" ]; then\n" +
		"  printf '%s' " + shellQuote(browserExecutablePath) + "\n" +
		"  exit 0\n" +
		"fi\n" +
		"printf '%s\\n' \"$@\" > " + shellQuote(nodeArgsPath) + "\n" +
		"printf '%s\\n' \"${PLAYWRIGHT_BROWSERS_PATH:-}\" > " + shellQuote(nodeCachePath) + "\n" +
		"mkdir -p " + shellQuote(filepath.Dir(browserExecutablePath)) + "\n" +
		"printf '%s\\n' '#!/bin/sh' > " + shellQuote(browserExecutablePath) + "\n" +
		"chmod +x " + shellQuote(browserExecutablePath) + "\n"
	if err := os.WriteFile(nodeStub, []byte(nodeScript), 0o755); err != nil {
		t.Fatalf("write node stub: %v", err)
	}

	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("AGENTX_BROWSERD_TEST_SKIP_SOURCE_NODE_MODULES", "1")
	t.Setenv("AGENTX_BROWSERD_SKIP_PLAYWRIGHT_BROWSER_BOOTSTRAP", "0")

	if _, err := materializeBundledBrowserdFiles(stateRoot); err != nil {
		t.Fatalf("materialize bundled browserd files: %v", err)
	}
	if raw, err := os.ReadFile(npmArgsPath); err != nil {
		t.Fatalf("read npm args log: %v", err)
	} else if got := strings.Fields(string(raw)); !reflect.DeepEqual(got, []string{"ci", "--no-audit", "--no-fund"}) {
		t.Fatalf("unexpected npm args after stale rebuild: %#v", got)
	}
	install := readBundledBrowserdInstallMetadata(targetRoot)
	if install.BundleGeneration != bundledBrowserdBundleGeneration() || install.DependencyGeneration != bundledBrowserdDependencyGenerationID() {
		t.Fatalf("unexpected rebuilt install metadata: %#v", install)
	}
	if revision, generation, ready := bundledPlaywrightDeliveryInfo(targetRoot, stateRoot); !ready || revision != "1187" || generation == "" {
		t.Fatalf("unexpected bundled browserd delivery info after stale rebuild: revision=%q generation=%q ready=%v", revision, generation, ready)
	}
	if !bundledBrowserdBundleReady(targetRoot) || !bundledBrowserdDeliveryReady(targetRoot, stateRoot) {
		t.Fatalf("expected bundled browserd readiness after stale rebuild")
	}
}

func TestMaterializeBundledBrowserdFilesWritesBundleMetadata(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script npm stub is not portable to windows")
	}
	stateRoot := t.TempDir()
	binDir := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir npm stub dir: %v", err)
	}
	npmStub := filepath.Join(binDir, "npm")
	npmScript := "#!/bin/sh\n" +
		"set -eu\n" +
		"mkdir -p \"$PWD/node_modules/playwright\"\n" +
		"printf '%s\\n' '{\"name\":\"playwright\"}' > \"$PWD/node_modules/playwright/package.json\"\n"
	if err := os.WriteFile(npmStub, []byte(npmScript), 0o755); err != nil {
		t.Fatalf("write npm stub: %v", err)
	}

	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("AGENTX_BROWSERD_TEST_SKIP_SOURCE_NODE_MODULES", "1")
	t.Setenv("AGENTX_BROWSERD_TEST_MODE", "mock")

	targetPath, err := materializeBundledBrowserdFiles(stateRoot)
	if err != nil {
		t.Fatalf("materialize bundled browserd files: %v", err)
	}
	targetRoot := filepath.Dir(targetPath)
	metadata := readBundledBrowserdBundleMetadata(targetRoot)
	if metadata.BundleGeneration != bundledBrowserdBundleGeneration() {
		t.Fatalf("unexpected bundled browserd bundle metadata: %#v", metadata)
	}
	if !bundledBrowserdBundleReady(targetRoot) {
		t.Fatalf("expected bundled browserd bundle readiness")
	}
}

func TestWriteBundledPlaywrightBrowserCachePinTracksDeliveryGenerationSwitch(t *testing.T) {
	location := bundledPlaywrightBrowserCacheLocation{
		Path:   filepath.Join(t.TempDir(), "ms-playwright"),
		Source: playwrightCacheSourceAgentxOwned,
		Pinned: true,
	}
	firstGeneration := "delivery-a"
	if err := writeBundledPlaywrightBrowserCachePin(location, bundledPlaywrightBrowserCachePolicy{
		BundleGeneration:     "bundle-a",
		DependencyGeneration: "deps-a",
		BrowserRevision:      "1187",
		DeliveryGeneration:   firstGeneration,
		RetainedDeliveries:   []string{firstGeneration},
		BundleReady:          true,
		DeliveryReady:        true,
	}); err != nil {
		t.Fatalf("write first playwright cache pin: %v", err)
	}
	secondGeneration := "delivery-b"
	if err := writeBundledPlaywrightBrowserCachePin(location, bundledPlaywrightBrowserCachePolicy{
		BundleGeneration:     "bundle-b",
		DependencyGeneration: "deps-b",
		BrowserRevision:      "1187",
		DeliveryGeneration:   secondGeneration,
		RetainedDeliveries:   []string{secondGeneration},
		BundleReady:          true,
		DeliveryReady:        true,
	}); err != nil {
		t.Fatalf("write second playwright cache pin: %v", err)
	}
	payload := readBundledPlaywrightBrowserCachePinForLocation(location)
	if payload.DeliveryGeneration != secondGeneration || !reflect.DeepEqual(payload.RetainedDeliveries, []string{secondGeneration, firstGeneration}) {
		t.Fatalf("unexpected current delivery retention payload: %#v", payload)
	}
	if payload.LastEvictedDelivery != "" || payload.LastDeliverySwitchUnix == 0 {
		t.Fatalf("expected previous delivery generation to remain retained after same-revision switch: %#v", payload)
	}
	if payload.TargetDeliveryGeneration != secondGeneration || payload.LastReadyDelivery != secondGeneration {
		t.Fatalf("unexpected target/last-ready delivery payload after switch: %#v", payload)
	}
	if payload.DeliveryTransitionPending || payload.DeliveryTransitionStage != "" {
		t.Fatalf("unexpected delivery transition state for ready delivery switch: %#v", payload)
	}
}

func TestWriteBundledPlaywrightBrowserCachePinEvictsOldestRetainedDeliveryAtLimit(t *testing.T) {
	location := bundledPlaywrightBrowserCacheLocation{
		Path:   filepath.Join(t.TempDir(), "ms-playwright"),
		Source: playwrightCacheSourceAgentxOwned,
		Pinned: true,
	}
	firstGeneration := "delivery-a"
	secondGeneration := "delivery-b"
	thirdGeneration := "delivery-c"
	for _, tc := range []struct {
		generation string
		bundle     string
		deps       string
	}{
		{generation: firstGeneration, bundle: "bundle-a", deps: "deps-a"},
		{generation: secondGeneration, bundle: "bundle-b", deps: "deps-b"},
		{generation: thirdGeneration, bundle: "bundle-c", deps: "deps-c"},
	} {
		if err := writeBundledPlaywrightBrowserCachePin(location, bundledPlaywrightBrowserCachePolicy{
			BundleGeneration:     tc.bundle,
			DependencyGeneration: tc.deps,
			BrowserRevision:      "1187",
			DeliveryGeneration:   tc.generation,
			RetainedDeliveries:   []string{tc.generation},
			NodeModulesReady:     true,
			BrowserReady:         true,
			BundleReady:          true,
			DeliveryReady:        true,
		}); err != nil {
			t.Fatalf("write playwright cache pin for %s: %v", tc.generation, err)
		}
	}
	payload := readBundledPlaywrightBrowserCachePinForLocation(location)
	if payload.DeliveryGeneration != thirdGeneration || !reflect.DeepEqual(payload.RetainedDeliveries, []string{thirdGeneration, secondGeneration}) {
		t.Fatalf("unexpected current delivery retention payload after reaching limit: %#v", payload)
	}
	if payload.RetainedFallbackDelivery != secondGeneration || !payload.RetainedFallbackLaunch || payload.RetainedFallbackBlock != "" {
		t.Fatalf("expected most recent retained ready delivery to remain launchable fallback: %#v", payload)
	}
	if payload.LastEvictedDelivery != firstGeneration {
		t.Fatalf("expected oldest retained delivery generation to be evicted at limit: %#v", payload)
	}
	if payload.TargetDeliveryGeneration != thirdGeneration || payload.LastReadyDelivery != thirdGeneration {
		t.Fatalf("unexpected target/last-ready delivery payload after retention eviction: %#v", payload)
	}
}

func TestWriteBundledPlaywrightBrowserCachePinDoesNotRetainAcrossBrowserRevisionSwitch(t *testing.T) {
	location := bundledPlaywrightBrowserCacheLocation{
		Path:   filepath.Join(t.TempDir(), "ms-playwright"),
		Source: playwrightCacheSourceAgentxOwned,
		Pinned: true,
	}
	firstGeneration := "delivery-a"
	if err := writeBundledPlaywrightBrowserCachePin(location, bundledPlaywrightBrowserCachePolicy{
		BundleGeneration:     "bundle-a",
		DependencyGeneration: "deps-a",
		BrowserRevision:      "1187",
		DeliveryGeneration:   firstGeneration,
		RetainedDeliveries:   []string{firstGeneration},
		BundleReady:          true,
		DeliveryReady:        true,
	}); err != nil {
		t.Fatalf("write first playwright cache pin: %v", err)
	}
	secondGeneration := "delivery-b"
	if err := writeBundledPlaywrightBrowserCachePin(location, bundledPlaywrightBrowserCachePolicy{
		BundleGeneration:     "bundle-b",
		DependencyGeneration: "deps-b",
		BrowserRevision:      "2200",
		DeliveryGeneration:   secondGeneration,
		RetainedDeliveries:   []string{secondGeneration},
		BundleReady:          true,
		DeliveryReady:        true,
	}); err != nil {
		t.Fatalf("write second playwright cache pin with new revision: %v", err)
	}
	payload := readBundledPlaywrightBrowserCachePinForLocation(location)
	if payload.DeliveryGeneration != secondGeneration || !reflect.DeepEqual(payload.RetainedDeliveries, []string{secondGeneration}) {
		t.Fatalf("expected cross-revision switch to reset retained deliveries to current generation: %#v", payload)
	}
	if payload.LastEvictedDelivery != firstGeneration {
		t.Fatalf("expected previous delivery generation to be evicted on browser revision switch: %#v", payload)
	}
}

func TestWriteBundledPlaywrightBrowserCachePinPreservesRetainedDeliveryDuringPendingTransition(t *testing.T) {
	location := bundledPlaywrightBrowserCacheLocation{
		Path:   filepath.Join(t.TempDir(), "ms-playwright"),
		Source: playwrightCacheSourceAgentxOwned,
		Pinned: true,
	}
	firstGeneration := "delivery-a"
	if err := writeBundledPlaywrightBrowserCachePin(location, bundledPlaywrightBrowserCachePolicy{
		BundleGeneration:     "bundle-a",
		DependencyGeneration: "deps-a",
		BrowserRevision:      "1187",
		DeliveryGeneration:   firstGeneration,
		RetainedDeliveries:   []string{firstGeneration},
		RetainedDirs:         []string{"chromium-1187", "chromium_headless_shell-1187"},
		NodeModulesReady:     true,
		BrowserReady:         true,
		BundleReady:          true,
		DeliveryReady:        true,
	}); err != nil {
		t.Fatalf("write initial playwright cache pin: %v", err)
	}
	if err := writeBundledPlaywrightBrowserCachePin(location, bundledPlaywrightBrowserCachePolicy{
		BundleGeneration:          "bundle-b",
		DependencyGeneration:      "deps-b",
		BrowserRevision:           "",
		DeliveryGeneration:        "",
		RetainedDeliveries:        nil,
		DeliveryTransitionPending: true,
		DeliveryTransitionStage:   "dependencies_not_ready",
		BundleReady:               true,
		DeliveryReady:             false,
		BootstrapState:            playwrightBootstrapStateFailed,
		BootstrapErrorCode:        playwrightBootstrapErrorNPMCIFailed,
		NodeModulesReady:          false,
		BrowserReady:              false,
		LastBootstrapUnixMilli:    time.Now().UnixMilli(),
	}); err != nil {
		t.Fatalf("write pending playwright cache pin: %v", err)
	}
	payload := readBundledPlaywrightBrowserCachePinForLocation(location)
	if payload.DeliveryGeneration != "" {
		t.Fatalf("expected pending delivery generation to remain unset: %#v", payload)
	}
	if payload.TargetDeliveryGeneration != "" || payload.LastReadyDelivery != firstGeneration {
		t.Fatalf("expected target to remain empty and last-ready delivery to stay pinned during transition: %#v", payload)
	}
	if !reflect.DeepEqual(payload.RetainedDeliveries, []string{firstGeneration}) || payload.LastEvictedDelivery != "" {
		t.Fatalf("expected retained ready delivery to be preserved during transition: %#v", payload)
	}
	if payload.RetainedFallbackDelivery != firstGeneration || !payload.RetainedFallbackPayload || payload.RetainedFallbackPayloadBR != "" || payload.RetainedFallbackLaunch || payload.RetainedFallbackBlock != "dependencies_not_ready" {
		t.Fatalf("expected retained ready delivery to remain explicit fallback target during transition: %#v", payload)
	}
	if payload.RetainedFallbackPayloadSrc != "retained_dirs" || !reflect.DeepEqual(payload.RetainedFallbackPayloadDirs, []string{"chromium-1187", "chromium_headless_shell-1187"}) {
		t.Fatalf("expected retained fallback payload provenance to stay explicit during transition: %#v", payload)
	}
	if payload.SelectedLaunchDelivery != firstGeneration || payload.SelectedLaunchSource != "retained_fallback" || payload.SelectedLaunchReady || payload.SelectedLaunchBlockReason != "dependencies_not_ready" {
		t.Fatalf("expected retained fallback to become the selected launch target during transition: %#v", payload)
	}
	if payload.SelectedLaunchRevision != "1187" || payload.SelectedLaunchPayloadSrc != "retained_dirs" || !reflect.DeepEqual(payload.SelectedLaunchPayloadDirs, []string{"chromium-1187", "chromium_headless_shell-1187"}) {
		t.Fatalf("expected retained fallback selected launch provenance during transition: %#v", payload)
	}
	if !payload.SelectedLaunchPayloadReady || payload.SelectedLaunchPayloadBR != "" {
		t.Fatalf("expected retained fallback selected launch payload to stay ready during transition: %#v", payload)
	}
	if payload.SelectedLaunchExecutable != "" || payload.SelectedLaunchExecutableOK || payload.SelectedLaunchExecutableBR != "selected_launch_executable_not_resolved" {
		t.Fatalf("expected retained fallback selected launch executable to remain unresolved without a local browser binary: %#v", payload)
	}
	if !payload.DeliveryTransitionPending || payload.DeliveryTransitionStage != "dependencies_not_ready" {
		t.Fatalf("expected explicit pending delivery transition state: %#v", payload)
	}
}

func TestWriteBundledPlaywrightBrowserCachePinMarksFallbackPayloadNotReady(t *testing.T) {
	location := bundledPlaywrightBrowserCacheLocation{
		Path:   filepath.Join(t.TempDir(), "ms-playwright"),
		Source: playwrightCacheSourceAgentxOwned,
		Pinned: true,
	}
	firstGeneration := "delivery-a"
	if err := writeBundledPlaywrightBrowserCachePin(location, bundledPlaywrightBrowserCachePolicy{
		BundleGeneration:     "bundle-a",
		DependencyGeneration: "deps-a",
		BrowserRevision:      "1187",
		DeliveryGeneration:   firstGeneration,
		RetainedDeliveries:   []string{firstGeneration},
		NodeModulesReady:     true,
		BrowserReady:         true,
		BundleReady:          true,
		DeliveryReady:        true,
	}); err != nil {
		t.Fatalf("write initial playwright cache pin: %v", err)
	}
	if err := writeBundledPlaywrightBrowserCachePin(location, bundledPlaywrightBrowserCachePolicy{
		BundleGeneration:          "bundle-b",
		DependencyGeneration:      "deps-b",
		BrowserRevision:           "1187",
		DeliveryGeneration:        "",
		RetainedDeliveries:        nil,
		RetainedDeliveryRevision:  "1187",
		RetainedDeliveryReady:     false,
		DeliveryTransitionPending: true,
		DeliveryTransitionStage:   "browser_not_ready",
		NodeModulesReady:          true,
		BrowserReady:              false,
		BundleReady:               true,
		DeliveryReady:             false,
	}); err != nil {
		t.Fatalf("write pending playwright cache pin without fallback payload: %v", err)
	}
	payload := readBundledPlaywrightBrowserCachePinForLocation(location)
	if payload.RetainedFallbackDelivery != firstGeneration {
		t.Fatalf("expected retained fallback delivery to remain explicit: %#v", payload)
	}
	if payload.RetainedFallbackPayload || payload.RetainedFallbackPayloadBR != "retained_delivery_cache_not_ready" {
		t.Fatalf("expected retained fallback payload to remain explicitly not ready: %#v", payload)
	}
	if payload.RetainedFallbackPayloadSrc != "" || len(payload.RetainedFallbackPayloadDirs) != 0 {
		t.Fatalf("expected retained fallback payload provenance to remain empty when payload is unavailable: %#v", payload)
	}
	if payload.RetainedFallbackLaunch || payload.RetainedFallbackBlock != "retained_fallback_payload_not_ready" {
		t.Fatalf("expected retained fallback launch gate to depend on payload readiness: %#v", payload)
	}
	if payload.SelectedLaunchDelivery != firstGeneration || payload.SelectedLaunchSource != "retained_fallback" || payload.SelectedLaunchReady || payload.SelectedLaunchBlockReason != "retained_fallback_payload_not_ready" {
		t.Fatalf("expected retained fallback payload failure to become selected launch policy: %#v", payload)
	}
	if payload.SelectedLaunchRevision != "1187" || payload.SelectedLaunchPayloadSrc != "" || len(payload.SelectedLaunchPayloadDirs) != 0 {
		t.Fatalf("expected selected launch provenance to remain incomplete when fallback payload is unavailable: %#v", payload)
	}
	if payload.SelectedLaunchPayloadReady || payload.SelectedLaunchPayloadBR != "retained_delivery_cache_not_ready" {
		t.Fatalf("expected selected launch payload readiness to reflect fallback payload failure: %#v", payload)
	}
	if payload.SelectedLaunchExecutable != "" || payload.SelectedLaunchExecutableOK || payload.SelectedLaunchExecutableBR != "retained_delivery_cache_not_ready" {
		t.Fatalf("expected selected launch executable readiness to mirror fallback payload failure: %#v", payload)
	}
}

func TestWriteBundledPlaywrightBrowserCachePinSelectsRetainedFallbackWhenCurrentDeliveryNotReady(t *testing.T) {
	location := bundledPlaywrightBrowserCacheLocation{
		Path:   filepath.Join(t.TempDir(), "ms-playwright"),
		Source: playwrightCacheSourceAgentxOwned,
		Pinned: true,
	}
	firstGeneration := "delivery-a"
	browserExecutablePath := writeFakeBundledPlaywrightChromiumExecutable(t, location.Path, "1187")
	if err := writeBundledPlaywrightBrowserCachePin(location, bundledPlaywrightBrowserCachePolicy{
		BundleGeneration:     "bundle-a",
		DependencyGeneration: "deps-a",
		BrowserRevision:      "1187",
		DeliveryGeneration:   firstGeneration,
		RetainedDeliveries:   []string{firstGeneration},
		RetainedDirs:         []string{"chromium-1187", "chromium_headless_shell-1187"},
		NodeModulesReady:     true,
		BrowserReady:         true,
		BundleReady:          true,
		DeliveryReady:        true,
		LaunchReady:          true,
	}); err != nil {
		t.Fatalf("write initial playwright cache pin: %v", err)
	}
	if err := writeBundledPlaywrightBrowserCachePin(location, bundledPlaywrightBrowserCachePolicy{
		BundleGeneration:          "bundle-b",
		DependencyGeneration:      "deps-b",
		BrowserRevision:           "",
		DeliveryGeneration:        "",
		RetainedDeliveries:        nil,
		DeliveryTransitionPending: true,
		DeliveryTransitionStage:   "browser_not_ready",
		BundleReady:               true,
		NodeModulesReady:          true,
		BrowserReady:              false,
		DeliveryReady:             false,
		LaunchReady:               false,
		LaunchBlockReason:         "browser_not_ready",
	}); err != nil {
		t.Fatalf("write pending playwright cache pin with fallback-eligible launch: %v", err)
	}
	payload := readBundledPlaywrightBrowserCachePinForLocation(location)
	if payload.RetainedFallbackDelivery != firstGeneration || !payload.RetainedFallbackLaunch || payload.RetainedFallbackBlock != "" {
		t.Fatalf("expected retained fallback to remain launch-ready: %#v", payload)
	}
	if payload.SelectedLaunchDelivery != firstGeneration || payload.SelectedLaunchSource != "retained_fallback" || !payload.SelectedLaunchReady || payload.SelectedLaunchBlockReason != "" {
		t.Fatalf("expected retained fallback to become the selected ready launch target: %#v", payload)
	}
	if payload.SelectedLaunchRevision != "1187" || payload.SelectedLaunchPayloadSrc != "retained_dirs" || !reflect.DeepEqual(payload.SelectedLaunchPayloadDirs, []string{"chromium-1187", "chromium_headless_shell-1187"}) {
		t.Fatalf("expected selected retained fallback launch provenance: %#v", payload)
	}
	if !payload.SelectedLaunchPayloadReady || payload.SelectedLaunchPayloadBR != "" {
		t.Fatalf("expected selected retained fallback launch payload readiness: %#v", payload)
	}
	if payload.SelectedLaunchExecutable != browserExecutablePath || !payload.SelectedLaunchExecutableOK || payload.SelectedLaunchExecutableBR != "" {
		t.Fatalf("expected selected retained fallback executable readiness: %#v", payload)
	}
}

func TestBundledPlaywrightBrowsersPathPrefersOverrideThenExistingDefaultThenAgentxCache(t *testing.T) {
	stateRoot := t.TempDir()
	overrideRoot := filepath.Join(stateRoot, "override-cache")
	t.Setenv("AGENTX_BROWSERD_PLAYWRIGHT_BROWSERS_PATH", overrideRoot)
	if got := bundledPlaywrightBrowsersPath(stateRoot); got != overrideRoot {
		t.Fatalf("expected override browsers path %q, got %q", overrideRoot, got)
	}
	if got := bundledPlaywrightBrowsersSource(stateRoot); got != playwrightCacheSourceOverride {
		t.Fatalf("expected override browsers source %q, got %q", playwrightCacheSourceOverride, got)
	}
	if bundledPlaywrightBrowsersPinned(stateRoot) {
		t.Fatalf("expected override browsers path to remain unpinned")
	}

	t.Setenv("AGENTX_BROWSERD_PLAYWRIGHT_BROWSERS_PATH", "")
	cacheRoot := configureIsolatedBrowserdUserCache(t)
	defaultCacheRoot := filepath.Join(cacheRoot, "ms-playwright")
	if err := os.MkdirAll(defaultCacheRoot, 0o755); err != nil {
		t.Fatalf("mkdir default playwright cache root: %v", err)
	}
	if got := bundledPlaywrightBrowsersPath(stateRoot); got != defaultCacheRoot {
		t.Fatalf("expected existing default playwright cache %q, got %q", defaultCacheRoot, got)
	}
	if got := bundledPlaywrightBrowsersSource(stateRoot); got != playwrightCacheSourceDefault {
		t.Fatalf("expected default-cache browsers source %q, got %q", playwrightCacheSourceDefault, got)
	}
	if bundledPlaywrightBrowsersPinned(stateRoot) {
		t.Fatalf("expected existing default playwright cache to remain unpinned")
	}

	if err := os.RemoveAll(defaultCacheRoot); err != nil {
		t.Fatalf("remove default playwright cache root: %v", err)
	}
	expectedAgentxCache := filepath.Join(cacheRoot, "agentx-browserd", "ms-playwright")
	if got := bundledPlaywrightBrowsersPath(stateRoot); got != expectedAgentxCache {
		t.Fatalf("expected agentx-owned playwright cache %q, got %q", expectedAgentxCache, got)
	}
	if got := bundledPlaywrightBrowsersSource(stateRoot); got != playwrightCacheSourceAgentxOwned {
		t.Fatalf("expected agentx-owned browsers source %q, got %q", playwrightCacheSourceAgentxOwned, got)
	}
	if !bundledPlaywrightBrowsersPinned(stateRoot) {
		t.Fatalf("expected agentx-owned playwright cache to be pinned")
	}
}

func TestEnsureBundledPlaywrightBrowserCachePinWritesAgentxOwnedMetadata(t *testing.T) {
	stateRoot := t.TempDir()
	homeRoot := filepath.Join(t.TempDir(), "home")
	targetRoot := filepath.Join(stateRoot, "bundled", "node")
	t.Setenv("HOME", homeRoot)
	if err := os.MkdirAll(targetRoot, 0o755); err != nil {
		t.Fatalf("mkdir bundled node root: %v", err)
	}

	if err := ensureBundledPlaywrightBrowserCachePin(targetRoot, stateRoot); err != nil {
		t.Fatalf("ensure playwright cache pin: %v", err)
	}

	cacheRoot := bundledPlaywrightBrowsersPath(stateRoot)
	raw, err := os.ReadFile(filepath.Join(cacheRoot, playwrightCachePinFilename))
	if err != nil {
		t.Fatalf("read playwright cache pin metadata: %v", err)
	}
	var payload struct {
		Owner         string   `json:"owner"`
		Path          string   `json:"path"`
		Source        string   `json:"source"`
		Pinned        bool     `json:"pinned"`
		PolicyVersion string   `json:"policy_version"`
		RetentionMode string   `json:"retention_mode"`
		RetainedDirs  []string `json:"retained_dirs"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("decode playwright cache pin metadata: %v", err)
	}
	if payload.Owner != "agentx-browserd" || payload.Path != cacheRoot || payload.Source != playwrightCacheSourceAgentxOwned || !payload.Pinned {
		t.Fatalf("unexpected playwright cache pin metadata: %#v", payload)
	}
	if payload.PolicyVersion != "" || payload.RetentionMode != "" || len(payload.RetainedDirs) != 0 {
		t.Fatalf("expected base playwright cache pin metadata without GC policy, got %#v", payload)
	}
}

func TestPruneInactiveBundledPlaywrightFallbackCacheRemovesLegacyStateRootCache(t *testing.T) {
	stateRoot := t.TempDir()
	homeRoot := filepath.Join(t.TempDir(), "home")
	t.Setenv("HOME", homeRoot)

	fallbackRoot := filepath.Join(stateRoot, "bundled", "ms-playwright")
	staleArtifact := filepath.Join(fallbackRoot, "chromium-legacy", "chrome")
	if err := os.MkdirAll(filepath.Dir(staleArtifact), 0o755); err != nil {
		t.Fatalf("mkdir legacy fallback playwright cache: %v", err)
	}
	if err := os.WriteFile(staleArtifact, []byte("legacy"), 0o644); err != nil {
		t.Fatalf("write legacy fallback playwright cache artifact: %v", err)
	}

	if err := pruneInactiveBundledPlaywrightFallbackCache(stateRoot); err != nil {
		t.Fatalf("prune inactive fallback playwright cache: %v", err)
	}
	if _, err := os.Stat(fallbackRoot); !os.IsNotExist(err) {
		t.Fatalf("expected legacy fallback playwright cache to be pruned, got err=%v", err)
	}
}

func TestPruneInactiveBundledPlaywrightFallbackCachePreservesActiveFallback(t *testing.T) {
	stateRoot := t.TempDir()
	fallbackRoot := filepath.Join(stateRoot, "bundled", "ms-playwright")
	t.Setenv("AGENTX_BROWSERD_PLAYWRIGHT_BROWSERS_PATH", fallbackRoot)

	activeArtifact := filepath.Join(fallbackRoot, "chromium-active", "chrome")
	if err := os.MkdirAll(filepath.Dir(activeArtifact), 0o755); err != nil {
		t.Fatalf("mkdir active fallback playwright cache: %v", err)
	}
	if err := os.WriteFile(activeArtifact, []byte("active"), 0o644); err != nil {
		t.Fatalf("write active fallback playwright cache artifact: %v", err)
	}

	if err := pruneInactiveBundledPlaywrightFallbackCache(stateRoot); err != nil {
		t.Fatalf("prune active fallback playwright cache: %v", err)
	}
	if _, err := os.Stat(activeArtifact); err != nil {
		t.Fatalf("expected active fallback playwright cache to remain, got %v", err)
	}
}

func TestPruneBundledPlaywrightPinnedChromiumCacheRemovesStaleRevisions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script node stub is not portable to windows")
	}
	stateRoot := t.TempDir()
	homeRoot := filepath.Join(t.TempDir(), "home")
	binDir := filepath.Join(t.TempDir(), "bin")
	t.Setenv("HOME", homeRoot)
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin dir: %v", err)
	}

	cacheRoot := bundledPlaywrightBrowsersPath(stateRoot)
	activeExecutable := filepath.Join(cacheRoot, "chromium-1187", "chrome")
	targetRoot := filepath.Join(stateRoot, "bundled", "node")
	nodeStub := filepath.Join(binDir, "node")
	nodeScript := "#!/bin/sh\n" +
		"set -eu\n" +
		"if [ \"${1:-}\" = \"-e\" ]; then\n" +
		"  printf '%s' " + shellQuote(activeExecutable) + "\n" +
		"  exit 0\n" +
		"fi\n" +
		"exit 1\n"
	if err := os.WriteFile(nodeStub, []byte(nodeScript), 0o755); err != nil {
		t.Fatalf("write node stub: %v", err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	if err := os.MkdirAll(targetRoot, 0o755); err != nil {
		t.Fatalf("mkdir target root: %v", err)
	}

	keepChrome := filepath.Join(cacheRoot, "chromium-1187", "chrome")
	keepHeadless := filepath.Join(cacheRoot, "chromium_headless_shell-1187", "headless-shell")
	staleChrome := filepath.Join(cacheRoot, "chromium-1111", "chrome")
	staleHeadless := filepath.Join(cacheRoot, "chromium_headless_shell-1111", "headless-shell")
	unrelated := filepath.Join(cacheRoot, "ffmpeg-100", "ffmpeg")
	for _, path := range []string{keepChrome, keepHeadless, staleChrome, staleHeadless, unrelated} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", path, err)
		}
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}

	if err := pruneBundledPlaywrightPinnedChromiumCache(targetRoot, stateRoot); err != nil {
		t.Fatalf("prune pinned chromium cache: %v", err)
	}
	for _, path := range []string{keepChrome, keepHeadless, unrelated} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected %s to remain, got %v", path, err)
		}
	}
	for _, path := range []string{
		filepath.Dir(staleChrome),
		filepath.Dir(staleHeadless),
	} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("expected stale cache %s to be pruned, got err=%v", path, err)
		}
	}
	raw, err := os.ReadFile(filepath.Join(cacheRoot, playwrightCachePinFilename))
	if err != nil {
		t.Fatalf("read playwright cache pin metadata: %v", err)
	}
	var payload struct {
		PolicyVersion            string   `json:"policy_version"`
		RetentionMode            string   `json:"retention_mode"`
		RetainedDeliveries       []string `json:"retained_delivery_generations"`
		RetainedDeliveryRevision string   `json:"retained_delivery_browser_revision"`
		RetainedDeliveryReady    bool     `json:"retained_delivery_cache_ready"`
		RetainedDirs             []string `json:"retained_dirs"`
		LastGCPrunedDirCount     int      `json:"last_gc_pruned_dir_count"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("decode playwright cache pin metadata: %v", err)
	}
	if payload.PolicyVersion != playwrightCachePolicyVersionV1 || payload.RetentionMode != playwrightCacheRetentionChromium {
		t.Fatalf("unexpected playwright cache GC policy metadata: %#v", payload)
	}
	if !reflect.DeepEqual(payload.RetainedDirs, []string{"chromium-1187", "chromium_headless_shell-1187"}) {
		t.Fatalf("unexpected retained dirs in playwright cache GC policy: %#v", payload)
	}
	if payload.RetainedDeliveryRevision != "" || payload.RetainedDeliveryReady || len(payload.RetainedDeliveries) != 0 {
		t.Fatalf("unexpected retained delivery availability in playwright cache GC policy without retained deliveries: %#v", payload)
	}
	if payload.LastGCPrunedDirCount != 2 {
		t.Fatalf("unexpected pruned dir count in playwright cache GC policy: %#v", payload)
	}
}

func TestPruneBundledPlaywrightPinnedChromiumCacheSkipsUnpinnedSource(t *testing.T) {
	stateRoot := t.TempDir()
	overrideRoot := filepath.Join(t.TempDir(), "override-ms-playwright")
	t.Setenv("AGENTX_BROWSERD_PLAYWRIGHT_BROWSERS_PATH", overrideRoot)

	staleChrome := filepath.Join(overrideRoot, "chromium-1111", "chrome")
	staleHeadless := filepath.Join(overrideRoot, "chromium_headless_shell-1111", "headless-shell")
	for _, path := range []string{staleChrome, staleHeadless} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", path, err)
		}
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}

	if err := pruneBundledPlaywrightPinnedChromiumCache(filepath.Join(stateRoot, "bundled", "node"), stateRoot); err != nil {
		t.Fatalf("prune unpinned chromium cache: %v", err)
	}
	for _, path := range []string{staleChrome, staleHeadless} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected %s to remain for unpinned cache, got %v", path, err)
		}
	}
}

func TestPruneBundledPlaywrightChromiumCacheForLocationPrunesPinnedStateRootFallback(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script node stub is not portable to windows")
	}
	targetRoot := filepath.Join(t.TempDir(), "bundled", "node")
	cacheRoot := filepath.Join(t.TempDir(), "bundled", "ms-playwright")
	binDir := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin dir: %v", err)
	}
	if err := os.MkdirAll(targetRoot, 0o755); err != nil {
		t.Fatalf("mkdir target root: %v", err)
	}

	activeExecutable := filepath.Join(cacheRoot, "chromium-2200", "chrome")
	nodeStub := filepath.Join(binDir, "node")
	nodeScript := "#!/bin/sh\n" +
		"set -eu\n" +
		"if [ \"${1:-}\" = \"-e\" ]; then\n" +
		"  printf '%s' " + shellQuote(activeExecutable) + "\n" +
		"  exit 0\n" +
		"fi\n" +
		"exit 1\n"
	if err := os.WriteFile(nodeStub, []byte(nodeScript), 0o755); err != nil {
		t.Fatalf("write node stub: %v", err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	keepChrome := filepath.Join(cacheRoot, "chromium-2200", "chrome")
	staleChrome := filepath.Join(cacheRoot, "chromium-1111", "chrome")
	staleHeadless := filepath.Join(cacheRoot, "chromium_headless_shell-1111", "headless-shell")
	for _, path := range []string{keepChrome, staleChrome, staleHeadless} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", path, err)
		}
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}

	location := bundledPlaywrightBrowserCacheLocation{
		Path:   cacheRoot,
		Source: playwrightCacheSourceStateRoot,
		Pinned: true,
	}
	if err := pruneBundledPlaywrightChromiumCacheForLocation(targetRoot, location); err != nil {
		t.Fatalf("prune pinned state-root chromium cache: %v", err)
	}
	if _, err := os.Stat(keepChrome); err != nil {
		t.Fatalf("expected active state-root cache to remain, got %v", err)
	}
	for _, path := range []string{filepath.Dir(staleChrome), filepath.Dir(staleHeadless)} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("expected stale state-root cache %s to be pruned, got err=%v", path, err)
		}
	}
	raw, err := os.ReadFile(filepath.Join(cacheRoot, playwrightCachePinFilename))
	if err != nil {
		t.Fatalf("read state-root playwright cache pin metadata: %v", err)
	}
	var payload struct {
		Source                   string   `json:"source"`
		PolicyVersion            string   `json:"policy_version"`
		RetentionMode            string   `json:"retention_mode"`
		RetainedDeliveries       []string `json:"retained_delivery_generations"`
		RetainedDeliveryRevision string   `json:"retained_delivery_browser_revision"`
		RetainedDeliveryReady    bool     `json:"retained_delivery_cache_ready"`
		RetainedDirs             []string `json:"retained_dirs"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("decode state-root playwright cache pin metadata: %v", err)
	}
	if payload.Source != playwrightCacheSourceStateRoot || payload.PolicyVersion != playwrightCachePolicyVersionV1 || payload.RetentionMode != playwrightCacheRetentionChromium {
		t.Fatalf("unexpected state-root playwright cache GC policy metadata: %#v", payload)
	}
	if !reflect.DeepEqual(payload.RetainedDirs, []string{"chromium-2200"}) {
		t.Fatalf("unexpected state-root retained dirs: %#v", payload)
	}
	if payload.RetainedDeliveryRevision != "" || payload.RetainedDeliveryReady || len(payload.RetainedDeliveries) != 0 {
		t.Fatalf("unexpected state-root retained delivery availability without retained deliveries: %#v", payload)
	}
}
