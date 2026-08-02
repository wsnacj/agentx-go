package browserd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

const (
	defaultManagerHealthTimeout = 10 * time.Second
	defaultManagerProbeInterval = 150 * time.Millisecond
	defaultBootstrapTimeout     = 10 * time.Minute
)

type managedBrowserdOwnershipMismatchError struct {
	Endpoint              string
	ExpectedStateRoot     string
	ExpectedProfilesRoot  string
	ExpectedArtifactsRoot string
	ExpectedLogsRoot      string
	ActualStateRoot       string
	ActualProfilesRoot    string
	ActualArtifactsRoot   string
	ActualLogsRoot        string
}

func (e *managedBrowserdOwnershipMismatchError) Error() string {
	if e == nil {
		return "browserdaemon: ownership mismatch"
	}
	return fmt.Sprintf(
		"browserdaemon: endpoint %s is owned by a different state root (expected=%s actual=%s)",
		strings.TrimSpace(e.Endpoint),
		strings.TrimSpace(e.ExpectedStateRoot),
		strings.TrimSpace(e.ActualStateRoot),
	)
}

type ManagerOptions struct {
	WorkspaceRoot    string
	Plan             Plan
	Probe            StatusProbe
	TransportTimeout int
	HealthTimeout    time.Duration
	ProbeInterval    time.Duration
	BootstrapTimeout time.Duration
}

type Manager struct {
	workspaceRoot    string
	plan             Plan
	probe            StatusProbe
	transportTimeout int
	healthTimeout    time.Duration
	probeInterval    time.Duration
	bootstrapTimeout time.Duration
	launchCommand    string
	launchArgs       []string
	launchResolved   bool
	lifecycleCtx     context.Context
	lifecycleCancel  context.CancelFunc

	startMu  sync.Mutex
	mu       sync.Mutex
	cmd      *exec.Cmd
	logFile  *os.File
	waitDone chan struct{}
	waitErr  error
	closed   bool
}

func NewManager(opts ManagerOptions) (*Manager, error) {
	plan := opts.Plan
	if !plan.Enabled {
		return nil, errors.New("browserdaemon: managed browserd plan is disabled")
	}
	if strings.TrimSpace(plan.Endpoint) == "" {
		return nil, errors.New("browserdaemon: managed browserd endpoint is required")
	}
	if strings.TrimSpace(plan.Command) == "" {
		return nil, errors.New("browserdaemon: managed browserd command is required")
	}
	if isBundledBrowserdCommand(plan.Command) && strings.TrimSpace(plan.StateRoot) == "" {
		return nil, errors.New("browserdaemon: state root is required for bundled browserd materialization")
	}
	healthTimeout := opts.HealthTimeout
	if healthTimeout <= 0 {
		healthTimeout = defaultManagerHealthTimeout
	}
	probeInterval := opts.ProbeInterval
	if probeInterval <= 0 {
		probeInterval = defaultManagerProbeInterval
	}
	bootstrapTimeout := opts.BootstrapTimeout
	if bootstrapTimeout <= 0 {
		bootstrapTimeout = defaultBootstrapTimeout
	}
	lifecycleCtx, lifecycleCancel := context.WithCancel(context.Background())
	return &Manager{
		workspaceRoot:    strings.TrimSpace(opts.WorkspaceRoot),
		plan:             plan,
		probe:            opts.Probe,
		transportTimeout: opts.TransportTimeout,
		healthTimeout:    healthTimeout,
		probeInterval:    probeInterval,
		bootstrapTimeout: bootstrapTimeout,
		lifecycleCtx:     lifecycleCtx,
		lifecycleCancel:  lifecycleCancel,
	}, nil
}

func (m *Manager) EnsureStarted(ctx context.Context) error {
	if m == nil {
		return errors.New("browserdaemon: manager is nil")
	}
	m.startMu.Lock()
	defer m.startMu.Unlock()

	operationCtx, cancel := m.operationContext(ctx)
	defer cancel()
	if err := m.ensureOpen(); err != nil {
		return err
	}
	status, err := m.Probe(operationCtx)
	if err == nil {
		return nil
	}
	var ownershipErr *managedBrowserdOwnershipMismatchError
	if errors.As(err, &ownershipErr) {
		if shutdownErr := shutdownManagedBrowserd(operationCtx, m.plan.Endpoint, m.plan.Token, m.transportTimeout); shutdownErr != nil {
			return fmt.Errorf("%w: shutdown foreign daemon: %v", ownershipErr, shutdownErr)
		}
		if waitErr := m.waitManagedBrowserdDown(operationCtx); waitErr != nil {
			return fmt.Errorf("%w: wait for foreign daemon shutdown: %v", ownershipErr, waitErr)
		}
	} else {
		_ = status
	}

	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return errors.New("browserdaemon: manager is closed")
	}
	restartCmd, restartDone := m.detachProcessLocked()
	m.mu.Unlock()
	if restartCmd != nil {
		stopDetachedProcess(restartCmd, restartDone)
	}
	if err := m.ensureLaunchResolved(operationCtx); err != nil {
		return err
	}

	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return errors.New("browserdaemon: manager is closed")
	}
	logFile, err := m.prepareLaunchLocked()
	if err != nil {
		m.mu.Unlock()
		return err
	}
	cmd := exec.Command(m.launchCommand, m.launchArgs...)
	if m.workspaceRoot != "" {
		cmd.Dir = m.workspaceRoot
	}
	cmd.Env = append(os.Environ(), m.launchEnv()...)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	if err := cmd.Start(); err != nil {
		_ = logFile.Close()
		m.mu.Unlock()
		return fmt.Errorf("browserdaemon: start managed browserd command=%s: %w", m.launchCommand, err)
	}
	waitDone := make(chan struct{})
	m.cmd = cmd
	m.logFile = logFile
	m.waitDone = waitDone
	m.waitErr = nil
	go m.wait(cmd, logFile, waitDone)
	m.mu.Unlock()

	probeCtx := operationCtx
	if _, ok := probeCtx.Deadline(); !ok && m.healthTimeout > 0 {
		var cancel context.CancelFunc
		probeCtx, cancel = context.WithTimeout(probeCtx, m.healthTimeout)
		defer cancel()
	}
	var lastErr error
	for {
		if _, err := m.Probe(probeCtx); err == nil {
			return nil
		} else {
			lastErr = err
		}
		if probeCtx.Err() != nil {
			break
		}
		if err := sleepWithContext(probeCtx, m.probeInterval); err != nil {
			break
		}
	}
	_ = m.Close()
	if lastErr == nil {
		lastErr = probeCtx.Err()
	}
	return lastErr
}

// StatusProbe reads the daemon's status using a host-supplied transport.
//
// The browserd host deliberately does not discover credentials or choose an HTTP
// backend. Callers must bind those policies explicitly.
type StatusProbe func(context.Context, Plan, int) (agentxbrowserruntime.BrowserProfileStatusResult, error)

func (m *Manager) Probe(ctx context.Context) (agentxbrowserruntime.BrowserProfileStatusResult, error) {
	if m == nil {
		return agentxbrowserruntime.BrowserProfileStatusResult{}, errors.New("browserdaemon: manager is nil")
	}
	if m.probe == nil {
		return agentxbrowserruntime.BrowserProfileStatusResult{}, errors.New("browserdaemon: status probe is required")
	}
	result, err := m.probe(ctx, m.plan, m.transportTimeout)
	if err != nil {
		return agentxbrowserruntime.BrowserProfileStatusResult{}, err
	}
	if ownershipErr := validateManagedBrowserdOwnership(m.plan, result); ownershipErr != nil {
		return result, ownershipErr
	}
	return result, nil
}

func (m *Manager) Close() error {
	if m == nil {
		return nil
	}
	if m.lifecycleCancel != nil {
		m.lifecycleCancel()
	}
	m.mu.Lock()
	m.closed = true
	cmd, done := m.detachProcessLocked()
	m.mu.Unlock()
	stopDetachedProcess(cmd, done)
	m.mu.Lock()
	waitErr := m.waitErr
	m.waitErr = nil
	m.mu.Unlock()
	if waitErr != nil && !isExpectedProcessStopErr(waitErr) {
		return waitErr
	}
	return nil
}

func (m *Manager) ensureOpen() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return errors.New("browserdaemon: manager is closed")
	}
	return nil
}

func (m *Manager) operationContext(parent context.Context) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithCancel(parent)
	if m.lifecycleCtx == nil {
		return ctx, cancel
	}
	stop := context.AfterFunc(m.lifecycleCtx, cancel)
	return ctx, func() {
		stop()
		cancel()
	}
}

func (m *Manager) ensureLaunchResolved(ctx context.Context) error {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return errors.New("browserdaemon: manager is closed")
	}
	if m.launchResolved {
		m.mu.Unlock()
		return nil
	}
	plan := m.plan
	timeout := m.bootstrapTimeout
	m.mu.Unlock()

	launchCommand, launchArgs, err := resolveManagerLaunchContext(ctx, plan, timeout)
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return errors.New("browserdaemon: manager is closed")
	}
	m.launchCommand = launchCommand
	m.launchArgs = launchArgs
	m.launchResolved = true
	return nil
}

func (m *Manager) wait(cmd *exec.Cmd, logFile *os.File, done chan struct{}) {
	err := cmd.Wait()
	if logFile != nil {
		_ = logFile.Close()
	}
	m.mu.Lock()
	m.waitErr = err
	if m.cmd == cmd {
		m.cmd = nil
		m.logFile = nil
		m.waitDone = nil
	}
	m.mu.Unlock()
	close(done)
}

func (m *Manager) detachProcessLocked() (*exec.Cmd, chan struct{}) {
	cmd := m.cmd
	done := m.waitDone
	m.cmd = nil
	m.logFile = nil
	m.waitDone = nil
	return cmd, done
}

func (m *Manager) prepareLaunchLocked() (*os.File, error) {
	for _, dir := range []string{m.plan.StateRoot, m.plan.ProfilesRoot, m.plan.ArtifactsRoot, m.plan.LogsRoot} {
		if strings.TrimSpace(dir) == "" {
			continue
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("browserdaemon: create directory %s: %w", dir, err)
		}
	}
	logPath := filepath.Join(m.plan.LogsRoot, "browserd.log")
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("browserdaemon: open log file %s: %w", logPath, err)
	}
	return file, nil
}

func (m *Manager) launchEnv() []string {
	playwrightCachePath := bundledPlaywrightBrowsersPath(m.plan.StateRoot)
	playwrightCacheSource := bundledPlaywrightBrowsersSource(m.plan.StateRoot)
	playwrightCachePinned := "0"
	if bundledPlaywrightBrowsersPinned(m.plan.StateRoot) {
		playwrightCachePinned = "1"
	}
	playwrightCachePin := readBundledPlaywrightBrowserCachePin(m.plan.StateRoot)
	bundleGeneration := strings.TrimSpace(playwrightCachePin.BundleGeneration)
	if bundleGeneration == "" {
		bundleGeneration = bundledBrowserdBundleGeneration()
	}
	dependencyGeneration := strings.TrimSpace(playwrightCachePin.DependencyGeneration)
	if dependencyGeneration == "" {
		dependencyGeneration = bundledBrowserdDependencyGenerationID()
	}
	browserRevision := strings.TrimSpace(playwrightCachePin.BrowserRevision)
	deliveryGeneration := strings.TrimSpace(playwrightCachePin.DeliveryGeneration)
	targetDeliveryGeneration := strings.TrimSpace(playwrightCachePin.TargetDeliveryGeneration)
	lastReadyDelivery := strings.TrimSpace(playwrightCachePin.LastReadyDelivery)
	retainedDeliveries := strings.Join(playwrightCachePin.RetainedDeliveries, ",")
	lastEvictedDelivery := strings.TrimSpace(playwrightCachePin.LastEvictedDelivery)
	lastDeliverySwitchUnix := fmt.Sprintf("%d", playwrightCachePin.LastDeliverySwitchUnix)
	retainedDeliveryRevision := strings.TrimSpace(playwrightCachePin.RetainedDeliveryRevision)
	retainedDeliveryReady := boolEnvValue(playwrightCachePin.RetainedDeliveryReady)
	retainedFallbackDelivery := strings.TrimSpace(playwrightCachePin.RetainedFallbackDelivery)
	retainedFallbackPayloadReady := boolEnvValue(playwrightCachePin.RetainedFallbackPayload)
	retainedFallbackPayloadBlock := strings.TrimSpace(playwrightCachePin.RetainedFallbackPayloadBR)
	retainedFallbackLaunchReady := boolEnvValue(playwrightCachePin.RetainedFallbackLaunch)
	retainedFallbackLaunchBlock := strings.TrimSpace(playwrightCachePin.RetainedFallbackBlock)
	selectedLaunchDelivery := strings.TrimSpace(playwrightCachePin.SelectedLaunchDelivery)
	selectedLaunchSource := strings.TrimSpace(playwrightCachePin.SelectedLaunchSource)
	selectedLaunchReady := boolEnvValue(playwrightCachePin.SelectedLaunchReady)
	selectedLaunchBlockReason := strings.TrimSpace(playwrightCachePin.SelectedLaunchBlockReason)
	selectedLaunchRevision := strings.TrimSpace(playwrightCachePin.SelectedLaunchRevision)
	selectedLaunchPayloadSource := strings.TrimSpace(playwrightCachePin.SelectedLaunchPayloadSrc)
	selectedLaunchPayloadReady := boolEnvValue(playwrightCachePin.SelectedLaunchPayloadReady)
	selectedLaunchPayloadBlock := strings.TrimSpace(playwrightCachePin.SelectedLaunchPayloadBR)
	selectedLaunchExecutable := strings.TrimSpace(playwrightCachePin.SelectedLaunchExecutable)
	selectedLaunchExecutableReady := boolEnvValue(playwrightCachePin.SelectedLaunchExecutableOK)
	selectedLaunchExecutableBlock := strings.TrimSpace(playwrightCachePin.SelectedLaunchExecutableBR)
	deliveryTransitionPending := boolEnvValue(playwrightCachePin.DeliveryTransitionPending)
	deliveryTransitionStage := strings.TrimSpace(playwrightCachePin.DeliveryTransitionStage)
	launchReady := boolEnvValue(playwrightCachePin.LaunchReady)
	launchBlockReason := strings.TrimSpace(playwrightCachePin.LaunchBlockReason)
	bundleReady := boolEnvValue(playwrightCachePin.BundleReady)
	deliveryReady := boolEnvValue(playwrightCachePin.DeliveryReady)
	return []string{
		"AGENTX_BROWSERD_HOST=" + strings.TrimSpace(m.plan.Host),
		"AGENTX_BROWSERD_PORT=" + strings.TrimSpace(fmt.Sprintf("%d", m.plan.Port)),
		"AGENTX_BROWSERD_ENDPOINT=" + strings.TrimSpace(m.plan.Endpoint),
		"AGENTX_BROWSERD_TOKEN=" + strings.TrimSpace(m.plan.Token),
		"AGENTX_BROWSERD_STATE_ROOT=" + strings.TrimSpace(m.plan.StateRoot),
		"AGENTX_BROWSERD_PROFILES_ROOT=" + strings.TrimSpace(m.plan.ProfilesRoot),
		"AGENTX_BROWSERD_ARTIFACTS_ROOT=" + strings.TrimSpace(m.plan.ArtifactsRoot),
		"AGENTX_BROWSERD_LOGS_ROOT=" + strings.TrimSpace(m.plan.LogsRoot),
		"AGENTX_BROWSERD_ATTACH_CDP_ENDPOINT=" + strings.TrimSpace(m.plan.AttachCDPEndpoint),
		"PLAYWRIGHT_BROWSERS_PATH=" + playwrightCachePath,
		"AGENTX_BROWSERD_PLAYWRIGHT_CACHE_SOURCE=" + playwrightCacheSource,
		"AGENTX_BROWSERD_PLAYWRIGHT_CACHE_PINNED=" + playwrightCachePinned,
		"AGENTX_BROWSERD_PLAYWRIGHT_BUNDLE_GENERATION=" + bundleGeneration,
		"AGENTX_BROWSERD_PLAYWRIGHT_DEPENDENCY_GENERATION=" + dependencyGeneration,
		"AGENTX_BROWSERD_PLAYWRIGHT_BROWSER_REVISION=" + browserRevision,
		"AGENTX_BROWSERD_PLAYWRIGHT_DELIVERY_GENERATION=" + deliveryGeneration,
		"AGENTX_BROWSERD_PLAYWRIGHT_TARGET_DELIVERY_GENERATION=" + targetDeliveryGeneration,
		"AGENTX_BROWSERD_PLAYWRIGHT_LAST_READY_DELIVERY_GENERATION=" + lastReadyDelivery,
		"AGENTX_BROWSERD_PLAYWRIGHT_RETAINED_DELIVERIES=" + retainedDeliveries,
		"AGENTX_BROWSERD_PLAYWRIGHT_LAST_EVICTED_DELIVERY_GENERATION=" + lastEvictedDelivery,
		"AGENTX_BROWSERD_PLAYWRIGHT_LAST_DELIVERY_SWITCH_UNIX_MILLI=" + lastDeliverySwitchUnix,
		"AGENTX_BROWSERD_PLAYWRIGHT_RETAINED_DELIVERY_BROWSER_REVISION=" + retainedDeliveryRevision,
		"AGENTX_BROWSERD_PLAYWRIGHT_RETAINED_DELIVERY_CACHE_READY=" + retainedDeliveryReady,
		"AGENTX_BROWSERD_PLAYWRIGHT_RETAINED_FALLBACK_DELIVERY_GENERATION=" + retainedFallbackDelivery,
		"AGENTX_BROWSERD_PLAYWRIGHT_RETAINED_FALLBACK_PAYLOAD_READY=" + retainedFallbackPayloadReady,
		"AGENTX_BROWSERD_PLAYWRIGHT_RETAINED_FALLBACK_PAYLOAD_BLOCK_REASON=" + retainedFallbackPayloadBlock,
		"AGENTX_BROWSERD_PLAYWRIGHT_RETAINED_FALLBACK_PAYLOAD_SOURCE=" + strings.TrimSpace(playwrightCachePin.RetainedFallbackPayloadSrc),
		"AGENTX_BROWSERD_PLAYWRIGHT_RETAINED_FALLBACK_PAYLOAD_DIRS=" + strings.Join(playwrightCachePin.RetainedFallbackPayloadDirs, ","),
		"AGENTX_BROWSERD_PLAYWRIGHT_RETAINED_FALLBACK_LAUNCH_READY=" + retainedFallbackLaunchReady,
		"AGENTX_BROWSERD_PLAYWRIGHT_RETAINED_FALLBACK_LAUNCH_BLOCK_REASON=" + retainedFallbackLaunchBlock,
		"AGENTX_BROWSERD_PLAYWRIGHT_SELECTED_LAUNCH_DELIVERY_GENERATION=" + selectedLaunchDelivery,
		"AGENTX_BROWSERD_PLAYWRIGHT_SELECTED_LAUNCH_SOURCE=" + selectedLaunchSource,
		"AGENTX_BROWSERD_PLAYWRIGHT_SELECTED_LAUNCH_READY=" + selectedLaunchReady,
		"AGENTX_BROWSERD_PLAYWRIGHT_SELECTED_LAUNCH_BLOCK_REASON=" + selectedLaunchBlockReason,
		"AGENTX_BROWSERD_PLAYWRIGHT_SELECTED_LAUNCH_BROWSER_REVISION=" + selectedLaunchRevision,
		"AGENTX_BROWSERD_PLAYWRIGHT_SELECTED_LAUNCH_PAYLOAD_SOURCE=" + selectedLaunchPayloadSource,
		"AGENTX_BROWSERD_PLAYWRIGHT_SELECTED_LAUNCH_PAYLOAD_DIRS=" + strings.Join(playwrightCachePin.SelectedLaunchPayloadDirs, ","),
		"AGENTX_BROWSERD_PLAYWRIGHT_SELECTED_LAUNCH_PAYLOAD_READY=" + selectedLaunchPayloadReady,
		"AGENTX_BROWSERD_PLAYWRIGHT_SELECTED_LAUNCH_PAYLOAD_BLOCK_REASON=" + selectedLaunchPayloadBlock,
		"AGENTX_BROWSERD_PLAYWRIGHT_SELECTED_LAUNCH_EXECUTABLE_PATH=" + selectedLaunchExecutable,
		"AGENTX_BROWSERD_PLAYWRIGHT_SELECTED_LAUNCH_EXECUTABLE_READY=" + selectedLaunchExecutableReady,
		"AGENTX_BROWSERD_PLAYWRIGHT_SELECTED_LAUNCH_EXECUTABLE_BLOCK_REASON=" + selectedLaunchExecutableBlock,
		"AGENTX_BROWSERD_PLAYWRIGHT_DELIVERY_TRANSITION_PENDING=" + deliveryTransitionPending,
		"AGENTX_BROWSERD_PLAYWRIGHT_DELIVERY_TRANSITION_STAGE=" + deliveryTransitionStage,
		"AGENTX_BROWSERD_PLAYWRIGHT_LAUNCH_READY=" + launchReady,
		"AGENTX_BROWSERD_PLAYWRIGHT_LAUNCH_BLOCK_REASON=" + launchBlockReason,
		"AGENTX_BROWSERD_PLAYWRIGHT_BUNDLE_READY=" + bundleReady,
		"AGENTX_BROWSERD_PLAYWRIGHT_DELIVERY_READY=" + deliveryReady,
		"AGENTX_BROWSERD_PLAYWRIGHT_CACHE_POLICY_VERSION=" + strings.TrimSpace(playwrightCachePin.PolicyVersion),
		"AGENTX_BROWSERD_PLAYWRIGHT_CACHE_RETENTION_MODE=" + strings.TrimSpace(playwrightCachePin.RetentionMode),
		"AGENTX_BROWSERD_PLAYWRIGHT_CACHE_RETAINED_DIRS=" + strings.Join(playwrightCachePin.RetainedDirs, ","),
		"AGENTX_BROWSERD_PLAYWRIGHT_CACHE_LAST_GC_PRUNED_DIR_COUNT=" + strings.TrimSpace(fmt.Sprintf("%d", playwrightCachePin.LastGCPrunedDirCount)),
		"AGENTX_BROWSERD_PLAYWRIGHT_BOOTSTRAP_STATE=" + strings.TrimSpace(playwrightCachePin.BootstrapState),
		"AGENTX_BROWSERD_PLAYWRIGHT_BOOTSTRAP_ERROR_CODE=" + strings.TrimSpace(playwrightCachePin.BootstrapErrorCode),
		"AGENTX_BROWSERD_PLAYWRIGHT_NODE_MODULES_READY=" + boolEnvValue(playwrightCachePin.NodeModulesReady),
		"AGENTX_BROWSERD_PLAYWRIGHT_BROWSER_READY=" + boolEnvValue(playwrightCachePin.BrowserReady),
	}
}

func boolEnvValue(v bool) string {
	if v {
		return "1"
	}
	return "0"
}

func validateManagedBrowserdOwnership(plan Plan, status agentxbrowserruntime.BrowserProfileStatusResult) error {
	expectedStateRoot := strings.TrimSpace(plan.StateRoot)
	expectedProfilesRoot := strings.TrimSpace(plan.ProfilesRoot)
	expectedArtifactsRoot := strings.TrimSpace(plan.ArtifactsRoot)
	expectedLogsRoot := strings.TrimSpace(plan.LogsRoot)
	actualStateRoot := strings.TrimSpace(status.StateRoot)
	actualProfilesRoot := strings.TrimSpace(status.ProfilesRoot)
	actualArtifactsRoot := strings.TrimSpace(status.ArtifactsRoot)
	actualLogsRoot := strings.TrimSpace(status.LogsRoot)
	if managedRootMatches(expectedStateRoot, actualStateRoot) &&
		managedRootMatches(expectedProfilesRoot, actualProfilesRoot) &&
		managedRootMatches(expectedArtifactsRoot, actualArtifactsRoot) &&
		managedRootMatches(expectedLogsRoot, actualLogsRoot) {
		return nil
	}
	return &managedBrowserdOwnershipMismatchError{
		Endpoint:              plan.Endpoint,
		ExpectedStateRoot:     expectedStateRoot,
		ExpectedProfilesRoot:  expectedProfilesRoot,
		ExpectedArtifactsRoot: expectedArtifactsRoot,
		ExpectedLogsRoot:      expectedLogsRoot,
		ActualStateRoot:       actualStateRoot,
		ActualProfilesRoot:    actualProfilesRoot,
		ActualArtifactsRoot:   actualArtifactsRoot,
		ActualLogsRoot:        actualLogsRoot,
	}
}

func managedRootMatches(expected string, actual string) bool {
	expected = strings.TrimSpace(expected)
	actual = strings.TrimSpace(actual)
	if expected == "" && actual == "" {
		return true
	}
	if expected == "" || actual == "" {
		return false
	}
	return filepath.Clean(expected) == filepath.Clean(actual)
}

func shutdownManagedBrowserd(ctx context.Context, endpoint string, token string, timeoutMs int) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if timeoutMs <= 0 {
		timeoutMs = 10000
	}
	runCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	body, err := json.Marshal(map[string]any{
		"method": "browser.shutdown",
		"params": map[string]any{},
	})
	if err != nil {
		return fmt.Errorf("encode shutdown request: %w", err)
	}
	req, err := http.NewRequestWithContext(runCtx, http.MethodPost, strings.TrimSpace(endpoint), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build shutdown request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if strings.TrimSpace(token) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(token))
	}
	req.Header.Set("X-AgentX-Browser-Method", "browser.shutdown")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		blob, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(blob)))
	}
	return nil
}

func (m *Manager) waitManagedBrowserdDown(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	waitCtx := ctx
	if _, ok := waitCtx.Deadline(); !ok && m.healthTimeout > 0 {
		var cancel context.CancelFunc
		waitCtx, cancel = context.WithTimeout(waitCtx, m.healthTimeout)
		defer cancel()
	}
	for {
		if _, err := m.Probe(waitCtx); err != nil {
			return nil
		}
		if waitCtx.Err() != nil {
			return waitCtx.Err()
		}
		if err := sleepWithContext(waitCtx, m.probeInterval); err != nil {
			return err
		}
	}
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func isExpectedProcessStopErr(err error) bool {
	if err == nil || errors.Is(err, exec.ErrWaitDelay) {
		return true
	}
	msg := strings.ToLower(strings.TrimSpace(err.Error()))
	return strings.Contains(msg, "signal: killed") || strings.Contains(msg, "signal: terminated") || strings.Contains(msg, "exit status 143")
}

func stopDetachedProcess(cmd *exec.Cmd, done chan struct{}) {
	if cmd == nil {
		return
	}
	if cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
	if done != nil {
		<-done
	}
}
