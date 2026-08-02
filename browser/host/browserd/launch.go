package browserd

import (
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	processcapture "github.com/wsnacj/agentx-go/browser/host/browserd/internal/processcapture"
)

const (
	bundledBrowserdCommandAlias                             = "agentx-browserd"
	bundledBrowserdCommandAuto                              = "bundled"
	bundledPlaywrightDownloadConnectionTimeoutDefault       = "120000"
	bundledBootstrapCommandStreamLimitBytes           int64 = 256 << 10

	playwrightCacheSourceOverride                    = "override"
	playwrightCacheSourceDefault                     = "default_cache"
	playwrightCacheSourceAgentxOwned                 = "agentx_owned"
	playwrightCacheSourceStateRoot                   = "state_root_fallback"
	playwrightCachePinFilename                       = ".agentx-browserd-cache.json"
	bundledBrowserdInstallMetadataFilename           = ".agentx-browserd-install.json"
	bundledBrowserdBundleMetadataFilename            = ".agentx-browserd-bundle.json"
	playwrightCachePolicyVersionV1                   = "v1"
	playwrightCacheRetentionChromium                 = "active_chromium_revision"
	playwrightCacheRetainedDeliveryLimit             = 2
	playwrightBootstrapStateReady                    = "ready"
	playwrightBootstrapStateFailed                   = "failed"
	playwrightBootstrapStateDepsReady                = "dependencies_ready"
	playwrightBootstrapErrorNPMMissing               = "npm_missing"
	playwrightBootstrapErrorNPMCIFailed              = "npm_ci_failed"
	playwrightBootstrapErrorPlaywrightDepMissing     = "playwright_dependency_missing"
	playwrightBootstrapErrorNodeMissing              = "node_missing"
	playwrightBootstrapErrorPlaywrightCLIMissing     = "playwright_cli_missing"
	playwrightBootstrapErrorBrowserInstallFailed     = "browser_install_failed"
	playwrightBootstrapErrorBrowserInstallTimeout    = "browser_install_timeout"
	playwrightBootstrapErrorBrowserInstallNetwork    = "browser_install_network_failed"
	playwrightBootstrapErrorBrowserExecutableMissing = "browser_executable_missing"
	playwrightBootstrapErrorStateRootUnwritable      = "state_root_unwritable"
)

type bundledBootstrapError struct {
	code      string
	operation string
	summary   string
	cause     error
}

func (e *bundledBootstrapError) Error() string {
	if e == nil {
		return "browserdaemon: bundled bootstrap failed"
	}
	message := "browserdaemon: " + strings.TrimSpace(e.operation)
	if code := strings.TrimSpace(e.code); code != "" {
		message += " code=" + code
	}
	if strings.TrimSpace(e.summary) != "" {
		message += ": " + strings.TrimSpace(e.summary)
	}
	return message
}

func (e *bundledBootstrapError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

//go:embed node/agentx-browserd.mjs node/package.json node/package-lock.json
var bundledBrowserdFiles embed.FS

var (
	bundledBrowserdGenerationOnce       sync.Once
	bundledBrowserdBundleGenerationID   string
	bundledBrowserdDependencyGeneration string
)

type bundledPlaywrightBrowserCacheLocation struct {
	Path   string
	Source string
	Pinned bool
}

type bundledPlaywrightBrowserCachePin struct {
	Owner                       string   `json:"owner,omitempty"`
	Path                        string   `json:"path,omitempty"`
	Source                      string   `json:"source,omitempty"`
	Pinned                      bool     `json:"pinned,omitempty"`
	BundleGeneration            string   `json:"bundle_generation,omitempty"`
	DependencyGeneration        string   `json:"dependency_generation,omitempty"`
	BrowserRevision             string   `json:"browser_revision,omitempty"`
	DeliveryGeneration          string   `json:"delivery_generation,omitempty"`
	TargetDeliveryGeneration    string   `json:"target_delivery_generation,omitempty"`
	LastReadyDelivery           string   `json:"last_ready_delivery_generation,omitempty"`
	RetainedDeliveries          []string `json:"retained_delivery_generations,omitempty"`
	LastEvictedDelivery         string   `json:"last_evicted_delivery_generation,omitempty"`
	LastDeliverySwitchUnix      int64    `json:"last_delivery_generation_switch_unix_milli,omitempty"`
	RetainedDeliveryRevision    string   `json:"retained_delivery_browser_revision,omitempty"`
	RetainedDeliveryReady       bool     `json:"retained_delivery_cache_ready,omitempty"`
	RetainedFallbackDelivery    string   `json:"retained_fallback_delivery_generation,omitempty"`
	RetainedFallbackPayload     bool     `json:"retained_fallback_payload_ready,omitempty"`
	RetainedFallbackPayloadBR   string   `json:"retained_fallback_payload_block_reason,omitempty"`
	RetainedFallbackPayloadSrc  string   `json:"retained_fallback_payload_source,omitempty"`
	RetainedFallbackPayloadDirs []string `json:"retained_fallback_payload_dirs,omitempty"`
	RetainedFallbackLaunch      bool     `json:"retained_fallback_launch_ready,omitempty"`
	RetainedFallbackBlock       string   `json:"retained_fallback_launch_block_reason,omitempty"`
	SelectedLaunchDelivery      string   `json:"selected_launch_delivery_generation,omitempty"`
	SelectedLaunchSource        string   `json:"selected_launch_source,omitempty"`
	SelectedLaunchReady         bool     `json:"selected_launch_ready,omitempty"`
	SelectedLaunchBlockReason   string   `json:"selected_launch_block_reason,omitempty"`
	SelectedLaunchRevision      string   `json:"selected_launch_browser_revision,omitempty"`
	SelectedLaunchPayloadSrc    string   `json:"selected_launch_payload_source,omitempty"`
	SelectedLaunchPayloadDirs   []string `json:"selected_launch_payload_dirs,omitempty"`
	SelectedLaunchPayloadReady  bool     `json:"selected_launch_payload_ready,omitempty"`
	SelectedLaunchPayloadBR     string   `json:"selected_launch_payload_block_reason,omitempty"`
	SelectedLaunchExecutable    string   `json:"selected_launch_executable_path,omitempty"`
	SelectedLaunchExecutableOK  bool     `json:"selected_launch_executable_ready,omitempty"`
	SelectedLaunchExecutableBR  string   `json:"selected_launch_executable_block_reason,omitempty"`
	DeliveryTransitionPending   bool     `json:"delivery_transition_pending,omitempty"`
	DeliveryTransitionStage     string   `json:"delivery_transition_stage,omitempty"`
	LaunchReady                 bool     `json:"launch_ready,omitempty"`
	LaunchBlockReason           string   `json:"launch_block_reason,omitempty"`
	BundleReady                 bool     `json:"bundle_ready,omitempty"`
	DeliveryReady               bool     `json:"delivery_ready,omitempty"`
	PolicyVersion               string   `json:"policy_version,omitempty"`
	RetentionMode               string   `json:"retention_mode,omitempty"`
	RetainedDirs                []string `json:"retained_dirs,omitempty"`
	LastGCPrunedDirCount        int      `json:"last_gc_pruned_dir_count,omitempty"`
	LastGCUnixMilli             int64    `json:"last_gc_unix_milli,omitempty"`
	BootstrapState              string   `json:"bootstrap_state,omitempty"`
	BootstrapErrorCode          string   `json:"bootstrap_error_code,omitempty"`
	NodeModulesReady            bool     `json:"node_modules_ready,omitempty"`
	BrowserReady                bool     `json:"browser_ready,omitempty"`
	LastBootstrapUnixMilli      int64    `json:"last_bootstrap_unix_milli,omitempty"`
	UpdatedAtUnixMilli          int64    `json:"updated_at_unix_milli,omitempty"`
}

type bundledPlaywrightBrowserCachePolicy struct {
	BundleGeneration          string
	DependencyGeneration      string
	BrowserRevision           string
	DeliveryGeneration        string
	TargetDeliveryGeneration  string
	LastReadyDelivery         string
	RetainedDeliveries        []string
	RetainedDeliveryRevision  string
	RetainedDeliveryReady     bool
	DeliveryTransitionPending bool
	DeliveryTransitionStage   string
	LaunchReady               bool
	LaunchBlockReason         string
	BundleReady               bool
	DeliveryReady             bool
	PolicyVersion             string
	RetentionMode             string
	RetainedDirs              []string
	LastGCPrunedDirCount      int
	LastGCUnixMilli           int64
	BootstrapState            string
	BootstrapErrorCode        string
	NodeModulesReady          bool
	BrowserReady              bool
	LastBootstrapUnixMilli    int64
}

type bundledBrowserdInstallMetadata struct {
	BundleGeneration     string `json:"bundle_generation,omitempty"`
	DependencyGeneration string `json:"dependency_generation,omitempty"`
	UpdatedAtUnixMilli   int64  `json:"updated_at_unix_milli,omitempty"`
}

type bundledBrowserdBundleMetadata struct {
	BundleGeneration   string `json:"bundle_generation,omitempty"`
	UpdatedAtUnixMilli int64  `json:"updated_at_unix_milli,omitempty"`
}

func resolveManagerLaunch(plan Plan) (string, []string, error) {
	return resolveManagerLaunchContext(context.Background(), plan, defaultBootstrapTimeout)
}

func resolveManagerLaunchContext(ctx context.Context, plan Plan, timeout time.Duration) (string, []string, error) {
	command := strings.TrimSpace(plan.Command)
	args := append([]string(nil), plan.Args...)
	switch {
	case command == "":
		return "", nil, fmt.Errorf("browserdaemon: managed browserd command is required")
	case isBundledBrowserdCommand(command):
		nodePath, err := exec.LookPath("node")
		if err != nil {
			return "", nil, &bundledBootstrapError{
				code:      playwrightBootstrapErrorNodeMissing,
				operation: "bundled browserd requires node in PATH",
				cause:     err,
			}
		}
		scriptPath, err := materializeBundledBrowserdFilesContext(ctx, plan.StateRoot, timeout)
		if err != nil {
			return "", nil, err
		}
		return nodePath, append([]string{scriptPath}, args...), nil
	default:
		return command, args, nil
	}
}

func isBundledBrowserdCommand(command string) bool {
	switch strings.ToLower(strings.TrimSpace(command)) {
	case bundledBrowserdCommandAlias, bundledBrowserdCommandAuto:
		return true
	default:
		return false
	}
}

func materializeBundledBrowserdFiles(stateRoot string) (string, error) {
	return materializeBundledBrowserdFilesContext(context.Background(), stateRoot, defaultBootstrapTimeout)
}

func materializeBundledBrowserdFilesContext(ctx context.Context, stateRoot string, timeout time.Duration) (string, error) {
	ctx, cancel := browserBootstrapContext(ctx, timeout)
	defer cancel()
	root := strings.TrimSpace(stateRoot)
	if root == "" {
		return "", fmt.Errorf("browserdaemon: state root is required for bundled browserd materialization")
	}
	targetRoot := filepath.Join(root, "bundled", "node")
	if err := os.MkdirAll(targetRoot, 0o755); err != nil {
		return "", &bundledBootstrapError{
			code:      playwrightBootstrapErrorStateRootUnwritable,
			operation: "create bundled browserd state root",
			cause:     fmt.Errorf("%s: %w", targetRoot, err),
		}
	}
	for _, name := range []string{"agentx-browserd.mjs", "package.json", "package-lock.json"} {
		blob, err := bundledBrowserdFiles.ReadFile(filepath.ToSlash(filepath.Join("node", name)))
		if err != nil {
			return "", fmt.Errorf("browserdaemon: read bundled file %s: %w", name, err)
		}
		targetPath := filepath.Join(targetRoot, name)
		if err := writeFileIfChanged(targetPath, blob, 0o644); err != nil {
			return "", fmt.Errorf("browserdaemon: write bundled file %s: %w", targetPath, err)
		}
	}
	if err := writeBundledBrowserdBundleMetadata(targetRoot); err != nil {
		return "", err
	}
	if err := ensureBundledNodeModulesLink(targetRoot); err != nil {
		return "", err
	}
	if err := ensureBundledNodeModulesAvailableContext(ctx, targetRoot, root); err != nil {
		return "", err
	}
	if err := pruneInactiveBundledPlaywrightFallbackCache(root); err != nil {
		return "", err
	}
	if err := ensureBundledPlaywrightBrowserCachePinContext(ctx, targetRoot, root); err != nil {
		return "", err
	}
	if err := ensureBundledPlaywrightBrowserAvailableContext(ctx, targetRoot, root); err != nil {
		return "", err
	}
	if err := pruneBundledPlaywrightPinnedChromiumCacheContext(ctx, targetRoot, root); err != nil {
		return "", err
	}
	return filepath.Join(targetRoot, bundledBrowserdEntry), nil
}

func browserBootstrapContext(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	if timeout <= 0 {
		timeout = defaultBootstrapTimeout
	}
	return context.WithTimeout(parent, timeout)
}

func writeFileIfChanged(path string, contents []byte, mode os.FileMode) error {
	if current, err := os.ReadFile(path); err == nil && string(current) == string(contents) {
		return nil
	}
	return os.WriteFile(path, contents, mode)
}

func ensureBundledNodeModulesLink(targetRoot string) error {
	if shouldSkipBundledSourceNodeModules() {
		return nil
	}
	sourceRoot, err := bundledBrowserdSourceRoot()
	if err != nil {
		return err
	}
	sourceModules := filepath.Join(sourceRoot, "node_modules")
	info, err := os.Stat(sourceModules)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("browserdaemon: stat bundled node_modules %s: %w", sourceModules, err)
	}
	if !info.IsDir() {
		return nil
	}
	linkPath := filepath.Join(targetRoot, "node_modules")
	if currentTarget, ok, err := readExistingSymlink(linkPath); err != nil {
		return fmt.Errorf("browserdaemon: inspect bundled node_modules link %s: %w", linkPath, err)
	} else if ok {
		if filepath.Clean(currentTarget) == filepath.Clean(sourceModules) {
			return nil
		}
		if err := os.Remove(linkPath); err != nil {
			return fmt.Errorf("browserdaemon: replace bundled node_modules link %s: %w", linkPath, err)
		}
	} else if _, err := os.Lstat(linkPath); err == nil {
		if err := os.RemoveAll(linkPath); err != nil {
			return fmt.Errorf("browserdaemon: remove bundled node_modules path %s: %w", linkPath, err)
		}
	}
	if err := os.Symlink(sourceModules, linkPath); err != nil {
		return fmt.Errorf("browserdaemon: create bundled node_modules link %s -> %s: %w", linkPath, sourceModules, err)
	}
	return nil
}

func ensureBundledNodeModulesAvailable(targetRoot string, stateRoot string) error {
	return ensureBundledNodeModulesAvailableContext(context.Background(), targetRoot, stateRoot)
}

func ensureBundledNodeModulesAvailableContext(ctx context.Context, targetRoot string, stateRoot string) error {
	location := bundledPlaywrightBrowsersLocation(stateRoot)
	if bundledNodeModulesReady(targetRoot) {
		_ = writeBundledBrowserdInstallMetadata(targetRoot)
		_ = recordBundledPlaywrightBootstrapStateContext(ctx, targetRoot, stateRoot, location, playwrightBootstrapStateDepsReady, "", true, false)
		return nil
	}
	if err := removeBundledBrowserdInstallMetadata(targetRoot); err != nil {
		return err
	}
	if err := removeBundledNodeModulesForBootstrap(targetRoot); err != nil {
		return err
	}
	if err := bootstrapBundledNodeModulesContext(ctx, targetRoot, stateRoot); err != nil {
		_ = recordBundledPlaywrightBootstrapStateContext(ctx, targetRoot, stateRoot, location, playwrightBootstrapStateFailed, classifyBundledNodeModulesBootstrapError(err), false, false)
		return err
	}
	if bundledPlaywrightDependencyPresent(targetRoot) {
		if err := writeBundledBrowserdInstallMetadata(targetRoot); err != nil {
			return err
		}
	}
	if bundledNodeModulesReady(targetRoot) {
		_ = recordBundledPlaywrightBootstrapStateContext(ctx, targetRoot, stateRoot, location, playwrightBootstrapStateDepsReady, "", true, false)
		return nil
	}
	_ = recordBundledPlaywrightBootstrapStateContext(ctx, targetRoot, stateRoot, location, playwrightBootstrapStateFailed, playwrightBootstrapErrorPlaywrightDepMissing, false, false)
	return fmt.Errorf("browserdaemon: bundled node_modules missing playwright dependency after bootstrap")
}

func bundledNodeModulesReady(targetRoot string) bool {
	if !bundledPlaywrightDependencyPresent(targetRoot) {
		return false
	}
	if bundledNodeModulesUseSourceLink(targetRoot) {
		return true
	}
	install := readBundledBrowserdInstallMetadata(targetRoot)
	return strings.TrimSpace(install.DependencyGeneration) == bundledBrowserdDependencyGenerationID()
}

func bundledPlaywrightDependencyPresent(targetRoot string) bool {
	info, err := os.Stat(filepath.Join(targetRoot, "node_modules", "playwright", "package.json"))
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func bootstrapBundledNodeModules(targetRoot string, stateRoot string) error {
	ctx, cancel := browserBootstrapContext(context.Background(), defaultBootstrapTimeout)
	defer cancel()
	return bootstrapBundledNodeModulesContext(ctx, targetRoot, stateRoot)
}

func bootstrapBundledNodeModulesContext(ctx context.Context, targetRoot string, stateRoot string) error {
	npmPath, err := exec.LookPath("npm")
	if err != nil {
		return &bundledBootstrapError{
			code:      playwrightBootstrapErrorNPMMissing,
			operation: "bundled browserd requires npm in PATH when source node_modules are unavailable",
			cause:     err,
		}
	}
	cacheRoot := filepath.Join(stateRoot, "bundled", ".npm-cache")
	if err := os.MkdirAll(cacheRoot, 0o755); err != nil {
		return fmt.Errorf("browserdaemon: create npm cache directory %s: %w", cacheRoot, err)
	}
	cmd := exec.CommandContext(ctx, npmPath, "ci", "--no-audit", "--no-fund")
	cmd.Dir = targetRoot
	cmd.Env = append(os.Environ(),
		"PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1",
		"npm_config_cache="+cacheRoot,
	)
	result, err := processcapture.Run(cmd, processcapture.Limits{
		StdoutBytes: bundledBootstrapCommandStreamLimitBytes,
		StderrBytes: bundledBootstrapCommandStreamLimitBytes,
	})
	if err != nil {
		code := playwrightBootstrapErrorNPMCIFailed
		cause := err
		if ctx.Err() != nil {
			cause = errors.Join(err, ctx.Err())
		}
		return &bundledBootstrapError{
			code:      code,
			operation: "bootstrap bundled node_modules via npm ci",
			summary:   result.Summary(),
			cause:     cause,
		}
	}
	return nil
}

func shouldSkipBundledSourceNodeModules() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("AGENTX_BROWSERD_TEST_SKIP_SOURCE_NODE_MODULES")), "1")
}

func ensureBundledPlaywrightBrowserAvailable(targetRoot string, stateRoot string) error {
	return ensureBundledPlaywrightBrowserAvailableContext(context.Background(), targetRoot, stateRoot)
}

func ensureBundledPlaywrightBrowserAvailableContext(ctx context.Context, targetRoot string, stateRoot string) error {
	location := bundledPlaywrightBrowsersLocation(stateRoot)
	if shouldSkipBundledPlaywrightBrowserBootstrap() {
		return nil
	}
	if bundledPlaywrightBrowserReadyContext(ctx, targetRoot, stateRoot) {
		_ = recordBundledPlaywrightBootstrapStateContext(ctx, targetRoot, stateRoot, location, playwrightBootstrapStateReady, "", bundledNodeModulesReady(targetRoot), true)
		return nil
	}
	if err := bootstrapBundledPlaywrightBrowserContext(ctx, targetRoot, stateRoot); err != nil {
		_ = recordBundledPlaywrightBootstrapStateContext(ctx, targetRoot, stateRoot, location, playwrightBootstrapStateFailed, classifyBundledPlaywrightBrowserBootstrapError(err), bundledNodeModulesReady(targetRoot), false)
		return err
	}
	if bundledPlaywrightBrowserReadyContext(ctx, targetRoot, stateRoot) {
		_ = recordBundledPlaywrightBootstrapStateContext(ctx, targetRoot, stateRoot, location, playwrightBootstrapStateReady, "", bundledNodeModulesReady(targetRoot), true)
		return nil
	}
	_ = recordBundledPlaywrightBootstrapStateContext(ctx, targetRoot, stateRoot, location, playwrightBootstrapStateFailed, playwrightBootstrapErrorBrowserExecutableMissing, bundledNodeModulesReady(targetRoot), false)
	return &bundledBootstrapError{
		code:      playwrightBootstrapErrorBrowserExecutableMissing,
		operation: "bundled playwright browser is still unavailable after bootstrap",
	}
}

func bundledPlaywrightBrowserReady(targetRoot string, stateRoot string) bool {
	ctx, cancel := browserBootstrapContext(context.Background(), defaultBootstrapTimeout)
	defer cancel()
	return bundledPlaywrightBrowserReadyContext(ctx, targetRoot, stateRoot)
}

func bundledPlaywrightBrowserReadyContext(ctx context.Context, targetRoot string, stateRoot string) bool {
	executablePath, err := bundledPlaywrightBrowserExecutablePathContext(ctx, targetRoot, stateRoot)
	if err != nil {
		return false
	}
	if executablePath == "" {
		return false
	}
	info, err := os.Stat(executablePath)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func bootstrapBundledPlaywrightBrowser(targetRoot string, stateRoot string) error {
	ctx, cancel := browserBootstrapContext(context.Background(), defaultBootstrapTimeout)
	defer cancel()
	return bootstrapBundledPlaywrightBrowserContext(ctx, targetRoot, stateRoot)
}

func bootstrapBundledPlaywrightBrowserContext(ctx context.Context, targetRoot string, stateRoot string) error {
	nodePath, err := exec.LookPath("node")
	if err != nil {
		return &bundledBootstrapError{
			code:      playwrightBootstrapErrorNodeMissing,
			operation: "bundled browser bootstrap requires node in PATH",
			cause:     err,
		}
	}
	cliPath := filepath.Join(targetRoot, "node_modules", "playwright", "cli.js")
	if _, err := os.Stat(cliPath); err != nil {
		return &bundledBootstrapError{
			code:      playwrightBootstrapErrorPlaywrightCLIMissing,
			operation: "bundled playwright cli is unavailable",
			cause:     err,
		}
	}
	cmd := exec.CommandContext(ctx, nodePath, cliPath, "install", "chromium")
	cmd.Dir = targetRoot
	cmd.Env = append(os.Environ(),
		"PLAYWRIGHT_BROWSERS_PATH="+bundledPlaywrightBrowsersPath(stateRoot),
		"PLAYWRIGHT_DOWNLOAD_CONNECTION_TIMEOUT="+bundledPlaywrightDownloadConnectionTimeout(),
	)
	result, err := processcapture.Run(cmd, processcapture.Limits{
		StdoutBytes: bundledBootstrapCommandStreamLimitBytes,
		StderrBytes: bundledBootstrapCommandStreamLimitBytes,
	})
	if err != nil {
		code := playwrightBootstrapErrorBrowserInstallFailed
		cause := err
		if ctx.Err() != nil {
			code = playwrightBootstrapErrorBrowserInstallTimeout
			cause = errors.Join(err, ctx.Err())
		} else if bundledPlaywrightBrowserInstallNetworkFailure(result) {
			code = playwrightBootstrapErrorBrowserInstallNetwork
		}
		return &bundledBootstrapError{
			code:      code,
			operation: "bootstrap playwright browser via install chromium",
			summary:   result.Summary(),
			cause:     cause,
		}
	}
	return nil
}

// bundledPlaywrightBrowserInstallNetworkFailure classifies only well-known
// transport failures from the private, bounded Playwright command capture.
// The captured output is never put into a public error or metadata surface;
// callers receive the stable error code plus the bounded command summary.
func bundledPlaywrightBrowserInstallNetworkFailure(result processcapture.Result) bool {
	if result.StdoutLimitExceeded || result.StderrLimitExceeded {
		return false
	}
	output := strings.ToLower(string(result.Stdout) + "\n" + string(result.Stderr))
	for _, marker := range []string{
		"econnrefused",
		"econnreset",
		"enotfound",
		"eai_again",
		"getaddrinfo",
		"network is unreachable",
		"connection timed out",
		"connect etimedout",
		"socket hang up",
	} {
		if strings.Contains(output, marker) {
			return true
		}
	}
	return false
}

func bundledPlaywrightDownloadConnectionTimeout() string {
	if value := strings.TrimSpace(os.Getenv("AGENTX_BROWSERD_PLAYWRIGHT_DOWNLOAD_CONNECTION_TIMEOUT")); value != "" {
		return value
	}
	if value := strings.TrimSpace(os.Getenv("PLAYWRIGHT_DOWNLOAD_CONNECTION_TIMEOUT")); value != "" {
		return value
	}
	return bundledPlaywrightDownloadConnectionTimeoutDefault
}

func bundledPlaywrightBrowserExecutablePath(targetRoot string, stateRoot string) (string, error) {
	ctx, cancel := browserBootstrapContext(context.Background(), defaultBootstrapTimeout)
	defer cancel()
	return bundledPlaywrightBrowserExecutablePathContext(ctx, targetRoot, stateRoot)
}

func bundledPlaywrightBrowserExecutablePathContext(ctx context.Context, targetRoot string, stateRoot string) (string, error) {
	return bundledPlaywrightBrowserExecutablePathForBrowsersPathContext(ctx, targetRoot, bundledPlaywrightBrowsersPath(stateRoot))
}

func bundledPlaywrightBrowserExecutablePathForBrowsersPath(targetRoot string, browsersPath string) (string, error) {
	ctx, cancel := browserBootstrapContext(context.Background(), defaultBootstrapTimeout)
	defer cancel()
	return bundledPlaywrightBrowserExecutablePathForBrowsersPathContext(ctx, targetRoot, browsersPath)
}

func bundledPlaywrightBrowserExecutablePathForBrowsersPathContext(ctx context.Context, targetRoot string, browsersPath string) (string, error) {
	nodePath, err := exec.LookPath("node")
	if err != nil {
		return "", err
	}
	cmd := exec.CommandContext(ctx, nodePath, "-e", `
import('playwright').then((p) => {
  const executable = p.chromium.executablePath();
  process.stdout.write(executable || '');
}).catch(() => {
  process.exit(1);
});
`)
	cmd.Dir = targetRoot
	cmd.Env = append(os.Environ(),
		"PLAYWRIGHT_BROWSERS_PATH="+strings.TrimSpace(browsersPath),
	)
	result, err := processcapture.Run(cmd, processcapture.Limits{
		StdoutBytes: 16 << 10,
		StderrBytes: 16 << 10,
	})
	if err != nil {
		cause := err
		if ctx.Err() != nil {
			cause = errors.Join(err, ctx.Err())
		}
		return "", &bundledBootstrapError{
			code:      playwrightBootstrapErrorBrowserExecutableMissing,
			operation: "resolve bundled playwright browser executable",
			summary:   result.Summary(),
			cause:     cause,
		}
	}
	return strings.TrimSpace(string(result.Stdout)), nil
}

func shouldSkipBundledPlaywrightBrowserBootstrap() bool {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("AGENTX_BROWSERD_SKIP_PLAYWRIGHT_BROWSER_BOOTSTRAP")), "1") {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(os.Getenv("AGENTX_BROWSERD_TEST_MODE")), "mock")
}

func ensureBundledPlaywrightBrowserCachePin(targetRoot string, stateRoot string) error {
	ctx, cancel := browserBootstrapContext(context.Background(), defaultBootstrapTimeout)
	defer cancel()
	return ensureBundledPlaywrightBrowserCachePinContext(ctx, targetRoot, stateRoot)
}

func ensureBundledPlaywrightBrowserCachePinContext(ctx context.Context, targetRoot string, stateRoot string) error {
	location := bundledPlaywrightBrowsersLocation(stateRoot)
	targetBrowserRevision := bundledPlaywrightBrowserRevisionContext(ctx, targetRoot, stateRoot)
	browserRevision, deliveryGeneration, deliveryReady := bundledPlaywrightDeliveryInfoContext(ctx, targetRoot, stateRoot)
	bundleReady := bundledBrowserdBundleReady(targetRoot)
	nodeModulesReady := bundledNodeModulesReady(targetRoot)
	browserReady := bundledPlaywrightBrowserReadyContext(ctx, targetRoot, stateRoot)
	targetDeliveryGeneration := bundledTargetDeliveryGeneration(targetBrowserRevision)
	launchReady, launchBlockReason := bundledPlaywrightLaunchGate(bundleReady, nodeModulesReady, browserReady, deliveryReady, "", "")
	retainedDeliveries := bundledRetainedDeliveries(deliveryGeneration)
	retainedDeliveryReady, retainedDeliveryRevision := bundledRetainedDeliveryCacheState(
		retainedDeliveries,
		browserRevision,
		nil,
		browserReady,
	)
	return writeBundledPlaywrightBrowserCachePin(location, bundledPlaywrightBrowserCachePolicy{
		BundleGeneration:          bundledBrowserdBundleGeneration(),
		DependencyGeneration:      bundledBrowserdDependencyGenerationID(),
		BrowserRevision:           browserRevision,
		DeliveryGeneration:        deliveryGeneration,
		TargetDeliveryGeneration:  targetDeliveryGeneration,
		LastReadyDelivery:         bundledLastReadyDeliveryGeneration(deliveryGeneration, nil),
		RetainedDeliveries:        retainedDeliveries,
		RetainedDeliveryRevision:  retainedDeliveryRevision,
		RetainedDeliveryReady:     retainedDeliveryReady,
		DeliveryTransitionPending: false,
		DeliveryTransitionStage:   "",
		LaunchReady:               launchReady,
		LaunchBlockReason:         launchBlockReason,
		BundleReady:               bundleReady,
		DeliveryReady:             deliveryReady,
		NodeModulesReady:          nodeModulesReady,
		BrowserReady:              browserReady,
	})
}

func writeBundledPlaywrightBrowserCachePin(location bundledPlaywrightBrowserCacheLocation, policy bundledPlaywrightBrowserCachePolicy) error {
	if !location.Pinned || strings.TrimSpace(location.Path) == "" {
		return nil
	}
	if err := os.MkdirAll(location.Path, 0o755); err != nil {
		return fmt.Errorf("browserdaemon: create playwright cache directory %s: %w", location.Path, err)
	}
	current := readBundledPlaywrightBrowserCachePinForLocation(location)
	current.Owner = "agentx-browserd"
	current.Path = location.Path
	current.Source = location.Source
	current.Pinned = location.Pinned
	if strings.TrimSpace(policy.BundleGeneration) != "" {
		current.BundleGeneration = strings.TrimSpace(policy.BundleGeneration)
	}
	if strings.TrimSpace(policy.DependencyGeneration) != "" {
		current.DependencyGeneration = strings.TrimSpace(policy.DependencyGeneration)
	}
	previousDelivery := strings.TrimSpace(current.DeliveryGeneration)
	previousBrowserRevision := strings.TrimSpace(current.BrowserRevision)
	current.DeliveryGeneration = strings.TrimSpace(policy.DeliveryGeneration)
	current.TargetDeliveryGeneration = strings.TrimSpace(policy.TargetDeliveryGeneration)
	if current.TargetDeliveryGeneration == "" && current.DeliveryGeneration != "" {
		current.TargetDeliveryGeneration = current.DeliveryGeneration
	}
	current.RetainedDeliveries, current.LastEvictedDelivery = bundledMergeRetainedDeliveries(
		policy.RetainedDeliveries,
		current.DeliveryGeneration,
		strings.TrimSpace(policy.BrowserRevision),
		previousDelivery,
		previousBrowserRevision,
		current.RetainedDeliveries,
	)
	if previousDelivery != "" && current.DeliveryGeneration != "" && previousDelivery != current.DeliveryGeneration {
		current.LastDeliverySwitchUnix = time.Now().UnixMilli()
	}
	switch {
	case strings.TrimSpace(policy.LastReadyDelivery) != "":
		current.LastReadyDelivery = strings.TrimSpace(policy.LastReadyDelivery)
	case current.DeliveryGeneration != "":
		current.LastReadyDelivery = current.DeliveryGeneration
	case len(current.RetainedDeliveries) > 0:
		current.LastReadyDelivery = strings.TrimSpace(current.RetainedDeliveries[0])
	default:
		current.LastReadyDelivery = ""
	}
	current.BrowserRevision = strings.TrimSpace(policy.BrowserRevision)
	switch {
	case strings.TrimSpace(policy.RetainedDeliveryRevision) != "":
		current.RetainedDeliveryRevision = strings.TrimSpace(policy.RetainedDeliveryRevision)
	case len(current.RetainedDeliveries) > 0 && current.BrowserRevision != "":
		current.RetainedDeliveryRevision = current.BrowserRevision
	case len(current.RetainedDeliveries) > 0 && strings.TrimSpace(current.RetainedDeliveryRevision) != "":
		current.RetainedDeliveryRevision = strings.TrimSpace(current.RetainedDeliveryRevision)
	default:
		current.RetainedDeliveryRevision = ""
	}
	current.DeliveryTransitionPending = policy.DeliveryTransitionPending
	current.DeliveryTransitionStage = strings.TrimSpace(policy.DeliveryTransitionStage)
	current.RetainedDeliveryReady = policy.RetainedDeliveryReady
	current.LaunchReady = policy.LaunchReady
	current.LaunchBlockReason = strings.TrimSpace(policy.LaunchBlockReason)
	current.BundleReady = policy.BundleReady
	current.DeliveryReady = policy.DeliveryReady
	if strings.TrimSpace(policy.PolicyVersion) != "" {
		current.PolicyVersion = strings.TrimSpace(policy.PolicyVersion)
	}
	if strings.TrimSpace(policy.RetentionMode) != "" {
		current.RetentionMode = strings.TrimSpace(policy.RetentionMode)
	}
	if policy.RetainedDirs != nil {
		current.RetainedDirs = append([]string(nil), policy.RetainedDirs...)
	}
	if policy.LastGCPrunedDirCount != 0 || policy.LastGCUnixMilli != 0 {
		current.LastGCPrunedDirCount = policy.LastGCPrunedDirCount
		current.LastGCUnixMilli = policy.LastGCUnixMilli
	}
	if strings.TrimSpace(policy.BootstrapState) != "" {
		current.BootstrapState = strings.TrimSpace(policy.BootstrapState)
	}
	if policy.BootstrapErrorCode != "" || policy.BootstrapState == playwrightBootstrapStateReady || policy.BootstrapState == playwrightBootstrapStateDepsReady {
		current.BootstrapErrorCode = strings.TrimSpace(policy.BootstrapErrorCode)
	}
	if policy.BootstrapState != "" ||
		policy.LastBootstrapUnixMilli != 0 ||
		strings.TrimSpace(policy.BundleGeneration) != "" ||
		strings.TrimSpace(policy.DependencyGeneration) != "" ||
		strings.TrimSpace(policy.BrowserRevision) != "" ||
		policy.BundleReady ||
		policy.DeliveryReady ||
		policy.NodeModulesReady ||
		policy.BrowserReady {
		current.NodeModulesReady = policy.NodeModulesReady
		current.BrowserReady = policy.BrowserReady
		if policy.LastBootstrapUnixMilli != 0 {
			current.LastBootstrapUnixMilli = policy.LastBootstrapUnixMilli
		}
	}
	current.RetainedDeliveryReady, current.RetainedDeliveryRevision = bundledRetainedDeliveryCacheState(
		current.RetainedDeliveries,
		current.RetainedDeliveryRevision,
		current.RetainedDirs,
		current.BrowserReady,
	)
	current.RetainedFallbackDelivery = bundledRetainedFallbackDeliveryGeneration(
		current.RetainedDeliveries,
		current.DeliveryGeneration,
		current.TargetDeliveryGeneration,
	)
	current.RetainedFallbackPayloadSrc, current.RetainedFallbackPayloadDirs = bundledRetainedFallbackPayloadProvenance(
		current.RetainedFallbackDelivery,
		current.RetainedDeliveryRevision,
		current.RetainedDirs,
		current.BrowserReady,
	)
	current.RetainedFallbackPayload, current.RetainedFallbackPayloadBR = bundledRetainedFallbackPayloadGate(
		current.RetainedFallbackDelivery,
		current.RetainedFallbackPayloadSrc,
	)
	current.RetainedFallbackLaunch, current.RetainedFallbackBlock = bundledRetainedFallbackLaunchGate(
		current.RetainedFallbackDelivery,
		current.RetainedFallbackPayload,
		current.BundleReady,
		current.NodeModulesReady,
	)
	current.SelectedLaunchDelivery, current.SelectedLaunchSource, current.SelectedLaunchReady, current.SelectedLaunchBlockReason = bundledSelectedPlaywrightLaunchTarget(
		current.DeliveryGeneration,
		current.TargetDeliveryGeneration,
		current.LaunchReady,
		current.LaunchBlockReason,
		current.RetainedFallbackDelivery,
		current.RetainedFallbackLaunch,
		current.RetainedFallbackBlock,
	)
	current.SelectedLaunchRevision, current.SelectedLaunchPayloadSrc, current.SelectedLaunchPayloadDirs = bundledSelectedPlaywrightLaunchProvenance(
		current.SelectedLaunchSource,
		current.BrowserRevision,
		current.RetainedDeliveryRevision,
		current.RetainedDirs,
		current.BrowserReady,
		current.RetainedFallbackPayloadSrc,
		current.RetainedFallbackPayloadDirs,
	)
	current.SelectedLaunchPayloadReady, current.SelectedLaunchPayloadBR = bundledSelectedPlaywrightLaunchPayloadGate(
		current.SelectedLaunchSource,
		current.SelectedLaunchPayloadSrc,
		current.RetainedFallbackPayload,
		current.RetainedFallbackPayloadBR,
	)
	current.SelectedLaunchExecutable, current.SelectedLaunchExecutableOK, current.SelectedLaunchExecutableBR = bundledSelectedPlaywrightLaunchExecutableState(
		location.Path,
		current.SelectedLaunchRevision,
		current.SelectedLaunchPayloadSrc,
		current.SelectedLaunchPayloadDirs,
		current.SelectedLaunchPayloadReady,
		current.SelectedLaunchPayloadBR,
	)
	current.UpdatedAtUnixMilli = time.Now().UnixMilli()
	payload, err := json.MarshalIndent(current, "", "  ")
	if err != nil {
		return fmt.Errorf("browserdaemon: encode playwright cache pin metadata: %w", err)
	}
	pinPath := filepath.Join(location.Path, playwrightCachePinFilename)
	if err := os.WriteFile(pinPath, payload, 0o644); err != nil {
		return fmt.Errorf("browserdaemon: write playwright cache pin metadata %s: %w", pinPath, err)
	}
	return nil
}

func recordBundledPlaywrightBootstrapState(targetRoot string, stateRoot string, location bundledPlaywrightBrowserCacheLocation, state string, code string, nodeModulesReady bool, browserReady bool) error {
	ctx, cancel := browserBootstrapContext(context.Background(), defaultBootstrapTimeout)
	defer cancel()
	return recordBundledPlaywrightBootstrapStateContext(ctx, targetRoot, stateRoot, location, state, code, nodeModulesReady, browserReady)
}

func recordBundledPlaywrightBootstrapStateContext(ctx context.Context, targetRoot string, stateRoot string, location bundledPlaywrightBrowserCacheLocation, state string, code string, nodeModulesReady bool, browserReady bool) error {
	targetBrowserRevision := bundledPlaywrightBrowserRevisionContext(ctx, targetRoot, stateRoot)
	browserRevision, deliveryGeneration, deliveryReady := bundledPlaywrightDeliveryInfoContext(ctx, targetRoot, stateRoot)
	currentPin := readBundledPlaywrightBrowserCachePinForLocation(location)
	bundleReady := bundledBrowserdBundleReady(targetRoot)
	transitionPending, transitionStage := bundledPlaywrightDeliveryTransitionState(
		bundleReady,
		nodeModulesReady,
		browserReady,
		deliveryReady,
		currentPin.RetainedDeliveries,
	)
	launchReady, launchBlockReason := bundledPlaywrightLaunchGate(bundleReady, nodeModulesReady, browserReady, deliveryReady, state, code)
	retainedDeliveries := bundledRetainedDeliveries(deliveryGeneration)
	targetDeliveryGeneration := bundledTargetDeliveryGeneration(targetBrowserRevision)
	retainedDeliveryReady, retainedDeliveryRevision := bundledRetainedDeliveryCacheState(
		retainedDeliveries,
		browserRevision,
		currentPin.RetainedDirs,
		browserReady,
	)
	return writeBundledPlaywrightBrowserCachePin(location, bundledPlaywrightBrowserCachePolicy{
		BundleGeneration:          bundledBrowserdBundleGeneration(),
		DependencyGeneration:      bundledBrowserdDependencyGenerationID(),
		BrowserRevision:           browserRevision,
		DeliveryGeneration:        deliveryGeneration,
		TargetDeliveryGeneration:  targetDeliveryGeneration,
		LastReadyDelivery:         bundledLastReadyDeliveryGeneration(deliveryGeneration, currentPin.RetainedDeliveries),
		RetainedDeliveries:        retainedDeliveries,
		RetainedDeliveryRevision:  retainedDeliveryRevision,
		RetainedDeliveryReady:     retainedDeliveryReady,
		DeliveryTransitionPending: transitionPending,
		DeliveryTransitionStage:   transitionStage,
		LaunchReady:               launchReady,
		LaunchBlockReason:         launchBlockReason,
		BundleReady:               bundleReady,
		DeliveryReady:             deliveryReady,
		BootstrapState:            state,
		BootstrapErrorCode:        code,
		NodeModulesReady:          nodeModulesReady,
		BrowserReady:              browserReady,
		LastBootstrapUnixMilli:    time.Now().UnixMilli(),
	})
}

func pruneInactiveBundledPlaywrightFallbackCache(stateRoot string) error {
	root := strings.TrimSpace(stateRoot)
	if root == "" {
		return nil
	}
	fallbackRoot := filepath.Join(root, "bundled", "ms-playwright")
	activeLocation := bundledPlaywrightBrowsersLocation(root)
	if pathsEqual(filepath.Clean(activeLocation.Path), filepath.Clean(fallbackRoot)) {
		return nil
	}
	if _, err := os.Stat(fallbackRoot); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("browserdaemon: stat fallback playwright cache %s: %w", fallbackRoot, err)
	}
	if err := os.RemoveAll(fallbackRoot); err != nil {
		return fmt.Errorf("browserdaemon: prune inactive fallback playwright cache %s: %w", fallbackRoot, err)
	}
	return nil
}

func pruneBundledPlaywrightPinnedChromiumCache(targetRoot string, stateRoot string) error {
	ctx, cancel := browserBootstrapContext(context.Background(), defaultBootstrapTimeout)
	defer cancel()
	return pruneBundledPlaywrightPinnedChromiumCacheContext(ctx, targetRoot, stateRoot)
}

func pruneBundledPlaywrightPinnedChromiumCacheContext(ctx context.Context, targetRoot string, stateRoot string) error {
	location := bundledPlaywrightBrowsersLocation(stateRoot)
	return pruneBundledPlaywrightChromiumCacheForLocationContext(ctx, targetRoot, location)
}

func pruneBundledPlaywrightChromiumCacheForLocation(targetRoot string, location bundledPlaywrightBrowserCacheLocation) error {
	ctx, cancel := browserBootstrapContext(context.Background(), defaultBootstrapTimeout)
	defer cancel()
	return pruneBundledPlaywrightChromiumCacheForLocationContext(ctx, targetRoot, location)
}

func pruneBundledPlaywrightChromiumCacheForLocationContext(ctx context.Context, targetRoot string, location bundledPlaywrightBrowserCacheLocation) error {
	if !location.Pinned || strings.TrimSpace(location.Path) == "" {
		return nil
	}
	executablePath, err := bundledPlaywrightBrowserExecutablePathForBrowsersPathContext(ctx, targetRoot, location.Path)
	if err != nil || strings.TrimSpace(executablePath) == "" {
		return nil
	}
	activeRoot, keepNames, ok := bundledPlaywrightChromiumRevisionKeepSet(location.Path, executablePath)
	if !ok {
		return nil
	}
	revision, _ := bundledPlaywrightChromiumRevisionSuffix(filepath.Base(activeRoot))
	deliveryGeneration := ""
	bundleReady := bundledBrowserdBundleReady(targetRoot)
	nodeModulesReady := bundledNodeModulesReady(targetRoot)
	if revision != "" && bundleReady && nodeModulesReady {
		deliveryGeneration = bundledBrowserdValueGenerationHash(
			bundledBrowserdBundleGeneration(),
			bundledBrowserdDependencyGenerationID(),
			revision,
		)
	}
	entries, err := os.ReadDir(location.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("browserdaemon: read playwright cache root %s: %w", location.Path, err)
	}
	prunedCount := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := strings.TrimSpace(entry.Name())
		if !isBundledPlaywrightChromiumRevisionDir(name) {
			continue
		}
		if keepNames[name] {
			continue
		}
		targetPath := filepath.Join(location.Path, name)
		if pathsEqual(targetPath, activeRoot) {
			continue
		}
		if err := os.RemoveAll(targetPath); err != nil {
			return fmt.Errorf("browserdaemon: prune stale playwright chromium cache %s: %w", targetPath, err)
		}
		prunedCount++
	}
	launchReady, launchBlockReason := bundledPlaywrightLaunchGate(bundleReady, nodeModulesReady, revision != "", deliveryGeneration != "", "", "")
	retainedDeliveries := bundledRetainedDeliveries(deliveryGeneration)
	retainedDirs := bundledPlaywrightChromiumRetainedDirs(location.Path, keepNames)
	retainedDeliveryReady, retainedDeliveryRevision := bundledRetainedDeliveryCacheState(
		retainedDeliveries,
		revision,
		retainedDirs,
		revision != "",
	)
	return writeBundledPlaywrightBrowserCachePin(location, bundledPlaywrightBrowserCachePolicy{
		BundleGeneration:          bundledBrowserdBundleGeneration(),
		DependencyGeneration:      bundledBrowserdDependencyGenerationID(),
		BrowserRevision:           revision,
		DeliveryGeneration:        deliveryGeneration,
		TargetDeliveryGeneration:  bundledTargetDeliveryGeneration(revision),
		LastReadyDelivery:         bundledLastReadyDeliveryGeneration(deliveryGeneration, nil),
		RetainedDeliveries:        retainedDeliveries,
		RetainedDeliveryRevision:  retainedDeliveryRevision,
		RetainedDeliveryReady:     retainedDeliveryReady,
		DeliveryTransitionPending: false,
		DeliveryTransitionStage:   "",
		LaunchReady:               launchReady,
		LaunchBlockReason:         launchBlockReason,
		BundleReady:               bundleReady,
		DeliveryReady:             deliveryGeneration != "",
		NodeModulesReady:          nodeModulesReady,
		BrowserReady:              revision != "",
		PolicyVersion:             playwrightCachePolicyVersionV1,
		RetentionMode:             playwrightCacheRetentionChromium,
		RetainedDirs:              retainedDirs,
		LastGCPrunedDirCount:      prunedCount,
		LastGCUnixMilli:           time.Now().UnixMilli(),
	})
}

func bundledPlaywrightBrowsersPath(stateRoot string) string {
	return bundledPlaywrightBrowsersLocation(stateRoot).Path
}

func bundledPlaywrightBrowsersSource(stateRoot string) string {
	return bundledPlaywrightBrowsersLocation(stateRoot).Source
}

func bundledPlaywrightBrowsersPinned(stateRoot string) bool {
	return bundledPlaywrightBrowsersLocation(stateRoot).Pinned
}

func bundledPlaywrightBrowsersLocation(stateRoot string) bundledPlaywrightBrowserCacheLocation {
	if override := strings.TrimSpace(os.Getenv("AGENTX_BROWSERD_PLAYWRIGHT_BROWSERS_PATH")); override != "" {
		return bundledPlaywrightBrowserCacheLocation{
			Path:   override,
			Source: playwrightCacheSourceOverride,
			Pinned: false,
		}
	}
	if existingDefault := existingDefaultPlaywrightBrowsersPath(); existingDefault != "" {
		return bundledPlaywrightBrowserCacheLocation{
			Path:   existingDefault,
			Source: playwrightCacheSourceDefault,
			Pinned: false,
		}
	}
	if cacheRoot, err := os.UserCacheDir(); err == nil && strings.TrimSpace(cacheRoot) != "" {
		return bundledPlaywrightBrowserCacheLocation{
			Path:   filepath.Join(cacheRoot, "agentx-browserd", "ms-playwright"),
			Source: playwrightCacheSourceAgentxOwned,
			Pinned: true,
		}
	}
	return bundledPlaywrightBrowserCacheLocation{
		Path:   filepath.Join(strings.TrimSpace(stateRoot), "bundled", "ms-playwright"),
		Source: playwrightCacheSourceStateRoot,
		Pinned: true,
	}
}

func existingDefaultPlaywrightBrowsersPath() string {
	cacheRoot, err := os.UserCacheDir()
	if err != nil || strings.TrimSpace(cacheRoot) == "" {
		return ""
	}
	defaultRoot := filepath.Join(cacheRoot, "ms-playwright")
	info, err := os.Stat(defaultRoot)
	if err != nil || !info.IsDir() {
		return ""
	}
	return defaultRoot
}

func bundledPlaywrightChromiumRevisionKeepSet(cacheRoot string, executablePath string) (string, map[string]bool, bool) {
	activeRoot, ok := bundledPlaywrightChromiumRevisionRoot(cacheRoot, executablePath)
	if !ok {
		return "", nil, false
	}
	activeName := strings.TrimSpace(filepath.Base(activeRoot))
	keep := map[string]bool{activeName: true}
	if suffix, ok := bundledPlaywrightChromiumRevisionSuffix(activeName); ok {
		for _, candidate := range []string{
			"chromium-" + suffix,
			"chromium_headless_shell-" + suffix,
		} {
			keep[candidate] = true
		}
	}
	return activeRoot, keep, true
}

func bundledPlaywrightChromiumRetainedDirs(cacheRoot string, keepNames map[string]bool) []string {
	if len(keepNames) == 0 {
		return nil
	}
	names := make([]string, 0, len(keepNames))
	for name := range keepNames {
		if _, err := os.Stat(filepath.Join(cacheRoot, name)); err == nil {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return nil
	}
	slices.Sort(names)
	return names
}

func readBundledPlaywrightBrowserCachePin(stateRoot string) bundledPlaywrightBrowserCachePin {
	location := bundledPlaywrightBrowsersLocation(stateRoot)
	return readBundledPlaywrightBrowserCachePinForLocation(location)
}

func readBundledPlaywrightBrowserCachePinForLocation(location bundledPlaywrightBrowserCacheLocation) bundledPlaywrightBrowserCachePin {
	if !location.Pinned || strings.TrimSpace(location.Path) == "" {
		return bundledPlaywrightBrowserCachePin{}
	}
	raw, err := os.ReadFile(filepath.Join(location.Path, playwrightCachePinFilename))
	if err != nil {
		return bundledPlaywrightBrowserCachePin{}
	}
	var payload bundledPlaywrightBrowserCachePin
	if err := json.Unmarshal(raw, &payload); err != nil {
		return bundledPlaywrightBrowserCachePin{}
	}
	return payload
}

func readBundledBrowserdInstallMetadata(targetRoot string) bundledBrowserdInstallMetadata {
	raw, err := os.ReadFile(filepath.Join(targetRoot, bundledBrowserdInstallMetadataFilename))
	if err != nil {
		return bundledBrowserdInstallMetadata{}
	}
	var payload bundledBrowserdInstallMetadata
	if err := json.Unmarshal(raw, &payload); err != nil {
		return bundledBrowserdInstallMetadata{}
	}
	return payload
}

func readBundledBrowserdBundleMetadata(targetRoot string) bundledBrowserdBundleMetadata {
	raw, err := os.ReadFile(filepath.Join(targetRoot, bundledBrowserdBundleMetadataFilename))
	if err != nil {
		return bundledBrowserdBundleMetadata{}
	}
	var payload bundledBrowserdBundleMetadata
	if err := json.Unmarshal(raw, &payload); err != nil {
		return bundledBrowserdBundleMetadata{}
	}
	return payload
}

func writeBundledBrowserdBundleMetadata(targetRoot string) error {
	payload := bundledBrowserdBundleMetadata{
		BundleGeneration:   bundledBrowserdBundleGeneration(),
		UpdatedAtUnixMilli: time.Now().UnixMilli(),
	}
	raw, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("browserdaemon: encode bundled browserd bundle metadata: %w", err)
	}
	if err := os.WriteFile(filepath.Join(targetRoot, bundledBrowserdBundleMetadataFilename), raw, 0o644); err != nil {
		return fmt.Errorf("browserdaemon: write bundled browserd bundle metadata: %w", err)
	}
	return nil
}

func writeBundledBrowserdInstallMetadata(targetRoot string) error {
	payload := bundledBrowserdInstallMetadata{
		BundleGeneration:     bundledBrowserdBundleGeneration(),
		DependencyGeneration: bundledBrowserdDependencyGenerationID(),
		UpdatedAtUnixMilli:   time.Now().UnixMilli(),
	}
	raw, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("browserdaemon: encode bundled browserd install metadata: %w", err)
	}
	if err := os.WriteFile(filepath.Join(targetRoot, bundledBrowserdInstallMetadataFilename), raw, 0o644); err != nil {
		return fmt.Errorf("browserdaemon: write bundled browserd install metadata: %w", err)
	}
	return nil
}

func removeBundledBrowserdInstallMetadata(targetRoot string) error {
	err := os.Remove(filepath.Join(targetRoot, bundledBrowserdInstallMetadataFilename))
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("browserdaemon: remove bundled browserd install metadata: %w", err)
	}
	return nil
}

func removeBundledNodeModulesForBootstrap(targetRoot string) error {
	nodeModulesPath := filepath.Join(targetRoot, "node_modules")
	if _, err := os.Lstat(nodeModulesPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("browserdaemon: inspect bundled node_modules before bootstrap: %w", err)
	}
	if err := os.RemoveAll(nodeModulesPath); err != nil {
		return fmt.Errorf("browserdaemon: remove bundled node_modules before bootstrap: %w", err)
	}
	return nil
}

func bundledNodeModulesUseSourceLink(targetRoot string) bool {
	if shouldSkipBundledSourceNodeModules() {
		return false
	}
	currentTarget, ok, err := readExistingSymlink(filepath.Join(targetRoot, "node_modules"))
	if err != nil || !ok {
		return false
	}
	sourceRoot, err := bundledBrowserdSourceRoot()
	if err != nil {
		return false
	}
	return pathsEqual(currentTarget, filepath.Join(sourceRoot, "node_modules"))
}

func bundledBrowserdBundleReady(targetRoot string) bool {
	for _, name := range []string{"agentx-browserd.mjs", "package.json", "package-lock.json"} {
		info, err := os.Stat(filepath.Join(targetRoot, name))
		if err != nil || info.IsDir() {
			return false
		}
	}
	metadata := readBundledBrowserdBundleMetadata(targetRoot)
	return strings.TrimSpace(metadata.BundleGeneration) == bundledBrowserdBundleGeneration()
}

func bundledBrowserdDeliveryReady(targetRoot string, stateRoot string) bool {
	ctx, cancel := browserBootstrapContext(context.Background(), defaultBootstrapTimeout)
	defer cancel()
	_, _, ready := bundledPlaywrightDeliveryInfoContext(ctx, targetRoot, stateRoot)
	return ready
}

func bundledPlaywrightDeliveryInfo(targetRoot string, stateRoot string) (string, string, bool) {
	ctx, cancel := browserBootstrapContext(context.Background(), defaultBootstrapTimeout)
	defer cancel()
	return bundledPlaywrightDeliveryInfoContext(ctx, targetRoot, stateRoot)
}

func bundledPlaywrightDeliveryInfoContext(ctx context.Context, targetRoot string, stateRoot string) (string, string, bool) {
	if !bundledBrowserdBundleReady(targetRoot) || !bundledNodeModulesReady(targetRoot) {
		return "", "", false
	}
	browserRevision := bundledPlaywrightBrowserRevisionContext(ctx, targetRoot, stateRoot)
	if browserRevision == "" {
		return "", "", false
	}
	return browserRevision, bundledBrowserdValueGenerationHash(
		bundledBrowserdBundleGeneration(),
		bundledBrowserdDependencyGenerationID(),
		browserRevision,
	), true
}

func bundledPlaywrightBrowserRevision(targetRoot string, stateRoot string) string {
	ctx, cancel := browserBootstrapContext(context.Background(), defaultBootstrapTimeout)
	defer cancel()
	return bundledPlaywrightBrowserRevisionContext(ctx, targetRoot, stateRoot)
}

func bundledPlaywrightBrowserRevisionContext(ctx context.Context, targetRoot string, stateRoot string) string {
	executablePath, err := bundledPlaywrightBrowserExecutablePathContext(ctx, targetRoot, stateRoot)
	if err != nil || strings.TrimSpace(executablePath) == "" {
		return ""
	}
	activeRoot, ok := bundledPlaywrightChromiumRevisionRoot(bundledPlaywrightBrowsersPath(stateRoot), executablePath)
	if !ok {
		return ""
	}
	suffix, ok := bundledPlaywrightChromiumRevisionSuffix(filepath.Base(activeRoot))
	if !ok {
		return ""
	}
	return suffix
}

func bundledBrowserdBundleGeneration() string {
	bundledBrowserdGenerationOnce.Do(computeBundledBrowserdGenerations)
	return bundledBrowserdBundleGenerationID
}

func bundledBrowserdDependencyGenerationID() string {
	bundledBrowserdGenerationOnce.Do(computeBundledBrowserdGenerations)
	return bundledBrowserdDependencyGeneration
}

func computeBundledBrowserdGenerations() {
	bundledBrowserdBundleGenerationID = bundledBrowserdGenerationHash([]string{"agentx-browserd.mjs", "package.json", "package-lock.json"})
	bundledBrowserdDependencyGeneration = bundledBrowserdGenerationHash([]string{"package.json", "package-lock.json"})
}

func bundledBrowserdGenerationHash(names []string) string {
	hash := sha256.New()
	for _, name := range names {
		path := filepath.ToSlash(filepath.Join("node", name))
		blob, err := bundledBrowserdFiles.ReadFile(path)
		if err != nil {
			continue
		}
		_, _ = hash.Write([]byte(name))
		_, _ = hash.Write([]byte{0})
		_, _ = hash.Write(blob)
		_, _ = hash.Write([]byte{0})
	}
	sum := hash.Sum(nil)
	return hex.EncodeToString(sum[:8])
}

func bundledBrowserdValueGenerationHash(values ...string) string {
	hash := sha256.New()
	for _, value := range values {
		_, _ = hash.Write([]byte(strings.TrimSpace(value)))
		_, _ = hash.Write([]byte{0})
	}
	sum := hash.Sum(nil)
	return hex.EncodeToString(sum[:8])
}

func bundledRetainedDeliveries(deliveryGeneration string) []string {
	if strings.TrimSpace(deliveryGeneration) == "" {
		return nil
	}
	return []string{strings.TrimSpace(deliveryGeneration)}
}

func bundledMergeRetainedDeliveries(explicit []string, currentDelivery string, currentBrowserRevision string, previousDelivery string, previousBrowserRevision string, previousRetained []string) ([]string, string) {
	currentDelivery = strings.TrimSpace(currentDelivery)
	currentBrowserRevision = strings.TrimSpace(currentBrowserRevision)
	previousDelivery = strings.TrimSpace(previousDelivery)
	previousBrowserRevision = strings.TrimSpace(previousBrowserRevision)
	previousRetained = bundledUniqueTrimmedDeliveries(previousRetained)

	var retained []string
	switch {
	case explicit != nil:
		retained = bundledUniqueTrimmedDeliveries(explicit)
	case currentDelivery != "":
		retained = []string{currentDelivery}
	default:
		retained = append([]string(nil), previousRetained...)
	}

	if currentDelivery != "" && currentBrowserRevision != "" && currentBrowserRevision == previousBrowserRevision {
		merged := []string{currentDelivery}
		for _, delivery := range previousRetained {
			if delivery == currentDelivery {
				continue
			}
			merged = append(merged, delivery)
		}
		if len(previousRetained) == 0 && previousDelivery != "" && previousDelivery != currentDelivery {
			merged = append(merged, previousDelivery)
		}
		retained = bundledUniqueTrimmedDeliveries(merged)
	}

	evicted := ""
	if len(retained) > playwrightCacheRetainedDeliveryLimit {
		evicted = strings.TrimSpace(retained[playwrightCacheRetainedDeliveryLimit])
		retained = retained[:playwrightCacheRetainedDeliveryLimit]
	}
	if evicted == "" {
		for _, delivery := range previousRetained {
			if delivery == "" || slices.Contains(retained, delivery) {
				continue
			}
			evicted = delivery
			break
		}
	}
	if evicted == "" && previousDelivery != "" && !slices.Contains(retained, previousDelivery) && currentDelivery != "" {
		evicted = previousDelivery
	}
	return retained, evicted
}

func bundledUniqueTrimmedDeliveries(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func bundledRetainedDeliveryCacheState(retainedDeliveries []string, browserRevision string, retainedDirs []string, browserReady bool) (bool, string) {
	retainedDeliveries = bundledUniqueTrimmedDeliveries(retainedDeliveries)
	browserRevision = strings.TrimSpace(browserRevision)
	if len(retainedDeliveries) == 0 || browserRevision == "" {
		return false, ""
	}
	if len(bundledRetainedDeliveryRevisionDirs(browserRevision, retainedDirs)) > 0 {
		return true, browserRevision
	}
	if browserReady {
		return true, browserRevision
	}
	return false, browserRevision
}

func bundledRetainedFallbackDeliveryGeneration(retainedDeliveries []string, currentDelivery string, targetDelivery string) string {
	retainedDeliveries = bundledUniqueTrimmedDeliveries(retainedDeliveries)
	currentDelivery = strings.TrimSpace(currentDelivery)
	targetDelivery = strings.TrimSpace(targetDelivery)
	for _, delivery := range retainedDeliveries {
		if delivery == "" {
			continue
		}
		if currentDelivery != "" && delivery == currentDelivery {
			continue
		}
		if targetDelivery != "" && delivery == targetDelivery {
			continue
		}
		return delivery
	}
	if currentDelivery == "" && targetDelivery == "" && len(retainedDeliveries) > 0 {
		return retainedDeliveries[0]
	}
	return ""
}

func bundledRetainedFallbackPayloadProvenance(retainedFallbackDelivery string, retainedDeliveryRevision string, retainedDirs []string, browserReady bool) (string, []string) {
	if strings.TrimSpace(retainedFallbackDelivery) == "" {
		return "", nil
	}
	matchedDirs := bundledRetainedDeliveryRevisionDirs(retainedDeliveryRevision, retainedDirs)
	if len(matchedDirs) > 0 {
		return "retained_dirs", matchedDirs
	}
	if browserReady && strings.TrimSpace(retainedDeliveryRevision) != "" {
		return "active_browser_revision", nil
	}
	return "", nil
}

func bundledRetainedFallbackPayloadGate(retainedFallbackDelivery string, retainedFallbackPayloadSource string) (bool, string) {
	if strings.TrimSpace(retainedFallbackDelivery) == "" {
		return false, ""
	}
	if strings.TrimSpace(retainedFallbackPayloadSource) == "" {
		return false, "retained_delivery_cache_not_ready"
	}
	return true, ""
}

func bundledRetainedFallbackLaunchGate(retainedFallbackDelivery string, retainedFallbackPayloadReady bool, bundleReady bool, nodeModulesReady bool) (bool, string) {
	if strings.TrimSpace(retainedFallbackDelivery) == "" {
		return false, ""
	}
	switch {
	case !retainedFallbackPayloadReady:
		return false, "retained_fallback_payload_not_ready"
	case !bundleReady:
		return false, "bundle_not_ready"
	case !nodeModulesReady:
		return false, "dependencies_not_ready"
	default:
		return true, ""
	}
}

func bundledSelectedPlaywrightLaunchTarget(currentDelivery string, targetDelivery string, currentLaunchReady bool, currentLaunchBlockReason string, retainedFallbackDelivery string, retainedFallbackLaunchReady bool, retainedFallbackLaunchBlockReason string) (string, string, bool, string) {
	currentDelivery = strings.TrimSpace(currentDelivery)
	targetDelivery = strings.TrimSpace(targetDelivery)
	currentLaunchBlockReason = strings.TrimSpace(currentLaunchBlockReason)
	retainedFallbackDelivery = strings.TrimSpace(retainedFallbackDelivery)
	retainedFallbackLaunchBlockReason = strings.TrimSpace(retainedFallbackLaunchBlockReason)
	switch {
	case currentLaunchReady && currentDelivery != "":
		return currentDelivery, "current_delivery", true, ""
	case retainedFallbackLaunchReady && retainedFallbackDelivery != "":
		return retainedFallbackDelivery, "retained_fallback", true, ""
	case retainedFallbackDelivery != "":
		return retainedFallbackDelivery, "retained_fallback", false, retainedFallbackLaunchBlockReason
	case currentDelivery != "":
		return currentDelivery, "current_delivery", false, currentLaunchBlockReason
	case targetDelivery != "":
		return targetDelivery, "target_delivery", false, currentLaunchBlockReason
	default:
		return "", "", false, currentLaunchBlockReason
	}
}

func bundledSelectedPlaywrightLaunchProvenance(selectedLaunchSource string, currentBrowserRevision string, retainedDeliveryRevision string, retainedDirs []string, browserReady bool, retainedFallbackPayloadSource string, retainedFallbackPayloadDirs []string) (string, string, []string) {
	selectedLaunchSource = strings.TrimSpace(selectedLaunchSource)
	currentBrowserRevision = strings.TrimSpace(currentBrowserRevision)
	retainedDeliveryRevision = strings.TrimSpace(retainedDeliveryRevision)
	retainedFallbackPayloadSource = strings.TrimSpace(retainedFallbackPayloadSource)
	switch selectedLaunchSource {
	case "retained_fallback":
		return retainedDeliveryRevision, retainedFallbackPayloadSource, append([]string(nil), retainedFallbackPayloadDirs...)
	case "current_delivery", "target_delivery":
		matchedDirs := bundledRetainedDeliveryRevisionDirs(currentBrowserRevision, retainedDirs)
		if len(matchedDirs) > 0 {
			return currentBrowserRevision, "retained_dirs", matchedDirs
		}
		if browserReady && currentBrowserRevision != "" {
			return currentBrowserRevision, "active_browser_revision", nil
		}
		return currentBrowserRevision, "", nil
	default:
		return "", "", nil
	}
}

func bundledSelectedPlaywrightLaunchPayloadGate(selectedLaunchSource string, selectedLaunchPayloadSource string, retainedFallbackPayloadReady bool, retainedFallbackPayloadBlockReason string) (bool, string) {
	selectedLaunchSource = strings.TrimSpace(selectedLaunchSource)
	selectedLaunchPayloadSource = strings.TrimSpace(selectedLaunchPayloadSource)
	retainedFallbackPayloadBlockReason = strings.TrimSpace(retainedFallbackPayloadBlockReason)
	switch selectedLaunchSource {
	case "current_delivery", "target_delivery":
		if selectedLaunchPayloadSource != "" {
			return true, ""
		}
		return false, selectedLaunchSource + "_payload_not_ready"
	case "retained_fallback":
		if retainedFallbackPayloadReady {
			return true, ""
		}
		if retainedFallbackPayloadBlockReason != "" {
			return false, retainedFallbackPayloadBlockReason
		}
		return false, "retained_fallback_payload_not_ready"
	default:
		return false, ""
	}
}

func bundledSelectedPlaywrightLaunchExecutableState(cacheRoot string, selectedLaunchRevision string, selectedLaunchPayloadSource string, selectedLaunchPayloadDirs []string, selectedLaunchPayloadReady bool, selectedLaunchPayloadBlockReason string) (string, bool, string) {
	selectedLaunchRevision = strings.TrimSpace(selectedLaunchRevision)
	selectedLaunchPayloadSource = strings.TrimSpace(selectedLaunchPayloadSource)
	selectedLaunchPayloadBlockReason = strings.TrimSpace(selectedLaunchPayloadBlockReason)
	if !selectedLaunchPayloadReady {
		if selectedLaunchPayloadBlockReason != "" {
			return "", false, selectedLaunchPayloadBlockReason
		}
		if selectedLaunchRevision != "" || selectedLaunchPayloadSource != "" || len(selectedLaunchPayloadDirs) > 0 {
			return "", false, "selected_launch_payload_not_ready"
		}
		return "", false, ""
	}
	executablePath, ok := bundledSelectedPlaywrightLaunchExecutablePath(cacheRoot, selectedLaunchRevision, selectedLaunchPayloadSource, selectedLaunchPayloadDirs)
	if !ok {
		return "", false, "selected_launch_executable_not_resolved"
	}
	info, err := os.Stat(executablePath)
	if err != nil || info.IsDir() {
		return "", false, "selected_launch_executable_not_ready"
	}
	return executablePath, true, ""
}

func bundledSelectedPlaywrightLaunchExecutablePath(cacheRoot string, selectedLaunchRevision string, selectedLaunchPayloadSource string, selectedLaunchPayloadDirs []string) (string, bool) {
	cacheRoot = strings.TrimSpace(cacheRoot)
	selectedLaunchRevision = strings.TrimSpace(selectedLaunchRevision)
	selectedLaunchPayloadSource = strings.TrimSpace(selectedLaunchPayloadSource)
	if cacheRoot == "" {
		return "", false
	}
	if selectedLaunchPayloadSource == "retained_dirs" {
		if executablePath, ok := bundledSelectedPlaywrightLaunchExecutablePathForDirs(cacheRoot, selectedLaunchRevision, selectedLaunchPayloadDirs); ok {
			return executablePath, true
		}
	}
	if selectedLaunchRevision == "" {
		return "", false
	}
	return bundledPlaywrightChromiumExecutablePathForDir(cacheRoot, "chromium-"+selectedLaunchRevision)
}

func bundledSelectedPlaywrightLaunchExecutablePathForDirs(cacheRoot string, selectedLaunchRevision string, selectedLaunchPayloadDirs []string) (string, bool) {
	selectedLaunchRevision = strings.TrimSpace(selectedLaunchRevision)
	if len(selectedLaunchPayloadDirs) == 0 {
		return "", false
	}
	for _, candidate := range bundledRetainedDeliveryRevisionDirs(selectedLaunchRevision, selectedLaunchPayloadDirs) {
		if executablePath, ok := bundledPlaywrightChromiumExecutablePathForDir(cacheRoot, candidate); ok {
			return executablePath, true
		}
	}
	for _, candidate := range selectedLaunchPayloadDirs {
		trimmed := strings.TrimSpace(candidate)
		if !strings.HasPrefix(trimmed, "chromium-") {
			continue
		}
		if executablePath, ok := bundledPlaywrightChromiumExecutablePathForDir(cacheRoot, trimmed); ok {
			return executablePath, true
		}
	}
	return "", false
}

func bundledPlaywrightChromiumExecutablePathForDir(cacheRoot string, revisionDir string) (string, bool) {
	cacheRoot = strings.TrimSpace(cacheRoot)
	revisionDir = strings.TrimSpace(revisionDir)
	if cacheRoot == "" || revisionDir == "" {
		return "", false
	}
	for _, candidate := range bundledPlaywrightChromiumExecutableCandidates(cacheRoot, revisionDir) {
		info, err := os.Stat(candidate)
		if err != nil || info.IsDir() {
			continue
		}
		return candidate, true
	}
	return "", false
}

func bundledPlaywrightChromiumExecutableCandidates(cacheRoot string, revisionDir string) []string {
	cacheRoot = strings.TrimSpace(cacheRoot)
	revisionDir = strings.TrimSpace(revisionDir)
	if cacheRoot == "" || revisionDir == "" {
		return nil
	}
	fileName := "chrome"
	if runtime.GOOS == "windows" {
		fileName = "chrome.exe"
	}
	candidates := []string{
		filepath.Join(cacheRoot, revisionDir, fileName),
	}
	if relativePath, ok := bundledPlaywrightChromiumExecutableRelativePath(); ok {
		candidates = append(candidates, filepath.Join(cacheRoot, revisionDir, relativePath))
	}
	return candidates
}

func bundledPlaywrightChromiumExecutableRelativePath() (string, bool) {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join("chrome-mac", "Chromium.app", "Contents", "MacOS", "Chromium"), true
	case "linux":
		return filepath.Join("chrome-linux", "chrome"), true
	case "windows":
		return filepath.Join("chrome-win", "chrome.exe"), true
	default:
		return "", false
	}
}

func bundledRetainedDeliveryRevisionAvailable(browserRevision string, retainedDirs []string) bool {
	return len(bundledRetainedDeliveryRevisionDirs(browserRevision, retainedDirs)) > 0
}

func bundledRetainedDeliveryRevisionDirs(browserRevision string, retainedDirs []string) []string {
	browserRevision = strings.TrimSpace(browserRevision)
	if browserRevision == "" || len(retainedDirs) == 0 {
		return nil
	}
	matched := make([]string, 0, 2)
	for _, candidate := range retainedDirs {
		trimmed := strings.TrimSpace(candidate)
		if trimmed == "" {
			continue
		}
		if trimmed == "chromium-"+browserRevision || trimmed == "chromium_headless_shell-"+browserRevision {
			matched = append(matched, trimmed)
		}
	}
	return matched
}

func bundledTargetDeliveryGeneration(browserRevision string) string {
	if strings.TrimSpace(browserRevision) == "" {
		return ""
	}
	return bundledBrowserdValueGenerationHash(
		bundledBrowserdBundleGeneration(),
		bundledBrowserdDependencyGenerationID(),
		browserRevision,
	)
}

func bundledLastReadyDeliveryGeneration(deliveryGeneration string, retainedDeliveries []string) string {
	if strings.TrimSpace(deliveryGeneration) != "" {
		return strings.TrimSpace(deliveryGeneration)
	}
	for _, delivery := range retainedDeliveries {
		if strings.TrimSpace(delivery) != "" {
			return strings.TrimSpace(delivery)
		}
	}
	return ""
}

func bundledPlaywrightDeliveryTransitionState(bundleReady bool, nodeModulesReady bool, browserReady bool, deliveryReady bool, retainedDeliveries []string) (bool, string) {
	if deliveryReady || len(retainedDeliveries) == 0 {
		return false, ""
	}
	switch {
	case !bundleReady:
		return true, "bundle_not_ready"
	case !nodeModulesReady:
		return true, "dependencies_not_ready"
	case !browserReady:
		return true, "browser_not_ready"
	default:
		return true, "delivery_not_ready"
	}
}

func bundledPlaywrightLaunchGate(bundleReady bool, nodeModulesReady bool, browserReady bool, deliveryReady bool, bootstrapState string, bootstrapErrorCode string) (bool, string) {
	if deliveryReady {
		return true, ""
	}
	if strings.EqualFold(strings.TrimSpace(bootstrapState), playwrightBootstrapStateFailed) && strings.TrimSpace(bootstrapErrorCode) != "" {
		return false, strings.TrimSpace(bootstrapErrorCode)
	}
	switch {
	case !bundleReady:
		return false, "bundle_not_ready"
	case !nodeModulesReady:
		return false, "dependencies_not_ready"
	case !browserReady:
		return false, "browser_not_ready"
	default:
		return false, "delivery_not_ready"
	}
}

func classifyBundledNodeModulesBootstrapError(err error) string {
	var bootstrapErr *bundledBootstrapError
	if errors.As(err, &bootstrapErr) && strings.TrimSpace(bootstrapErr.code) != "" {
		return strings.TrimSpace(bootstrapErr.code)
	}
	return "node_modules_bootstrap_failed"
}

func classifyBundledPlaywrightBrowserBootstrapError(err error) string {
	var bootstrapErr *bundledBootstrapError
	if errors.As(err, &bootstrapErr) && strings.TrimSpace(bootstrapErr.code) != "" {
		return strings.TrimSpace(bootstrapErr.code)
	}
	return "browser_bootstrap_failed"
}

func bundledPlaywrightChromiumRevisionRoot(cacheRoot string, executablePath string) (string, bool) {
	root := filepath.Clean(strings.TrimSpace(cacheRoot))
	executable := filepath.Clean(strings.TrimSpace(executablePath))
	if root == "" || executable == "" {
		return "", false
	}
	rel, err := filepath.Rel(root, executable)
	if err != nil {
		return "", false
	}
	if rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", false
	}
	parts := strings.Split(rel, string(os.PathSeparator))
	if len(parts) == 0 || !isBundledPlaywrightChromiumRevisionDir(parts[0]) {
		return "", false
	}
	return filepath.Join(root, parts[0]), true
}

func bundledPlaywrightChromiumRevisionSuffix(name string) (string, bool) {
	trimmed := strings.TrimSpace(name)
	switch {
	case strings.HasPrefix(trimmed, "chromium_headless_shell-"):
		suffix := strings.TrimPrefix(trimmed, "chromium_headless_shell-")
		return suffix, suffix != ""
	case strings.HasPrefix(trimmed, "chromium-"):
		suffix := strings.TrimPrefix(trimmed, "chromium-")
		return suffix, suffix != ""
	default:
		return "", false
	}
}

func isBundledPlaywrightChromiumRevisionDir(name string) bool {
	_, ok := bundledPlaywrightChromiumRevisionSuffix(name)
	return ok
}

func pathsEqual(a string, b string) bool {
	return filepath.Clean(strings.TrimSpace(a)) == filepath.Clean(strings.TrimSpace(b))
}

func readExistingSymlink(path string) (string, bool, error) {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, err
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return "", false, nil
	}
	target, err := os.Readlink(path)
	if err != nil {
		return "", false, err
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(path), target)
	}
	return target, true, nil
}

func bundledBrowserdSourceRoot() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("browserdaemon: resolve bundled browserd source root")
	}
	return filepath.Join(filepath.Dir(file), "node"), nil
}
