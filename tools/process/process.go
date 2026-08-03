// Package process provides an explicit, bounded local process adapter.
//
// The package owns portable command/result, cancellation, timeout and output
// capture semantics. It does not authorize commands, provide a sandbox, choose
// an approval policy or decide which signals may be sent. Hosts must perform
// those checks before calling the local adapter.
package process

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	stdruntime "runtime"
	"strings"
	"sync"
	"time"
)

const (
	defaultMaxOutputBytes     = 120_000
	defaultProcessOutputBytes = 4 << 20
	defaultWaitDelay          = 2 * time.Second
)

// ErrorCode is a stable, programmatic failure category.
type ErrorCode string

const (
	// ErrorCodeInvalidCommand marks an empty command.
	ErrorCodeInvalidCommand ErrorCode = "invalid_command"
	// ErrorCodeWorkdirResolution marks a workdir outside the configured root.
	ErrorCodeWorkdirResolution ErrorCode = "workdir_resolution_failed"
	// ErrorCodeCommandFailed marks a command that could not be started or waited.
	ErrorCodeCommandFailed ErrorCode = "command_failed"
	// ErrorCodeContextCanceled marks caller cancellation.
	ErrorCodeContextCanceled ErrorCode = "context_canceled"
	// ErrorCodeContextDeadlineExceeded marks a caller-owned deadline.
	ErrorCodeContextDeadlineExceeded ErrorCode = "context_deadline_exceeded"
	// ErrorCodeProcessListFailed marks a local process-list failure.
	ErrorCodeProcessListFailed ErrorCode = "process_list_failed"
	// ErrorCodeOutputLimitExceeded marks a process-list capture beyond its bound.
	ErrorCodeOutputLimitExceeded ErrorCode = "output_limit_exceeded"
)

// Error is a typed local adapter error. Cause remains available through
// errors.Is/errors.As; Message is safe for developer-facing diagnostics but
// must not be treated as an authorization decision.
type Error struct {
	Code    ErrorCode
	Op      string
	Message string
	Cause   error
}

// Error implements error.
func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	message := strings.TrimSpace(e.Message)
	if message == "" && e.Cause != nil {
		message = e.Cause.Error()
	}
	if message == "" {
		message = string(e.Code)
	}
	if op := strings.TrimSpace(e.Op); op != "" {
		return op + ": " + message
	}
	return message
}

// Unwrap exposes the underlying context, OS or command error.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// AsError extracts a typed process error.
func AsError(err error) (*Error, bool) {
	var target *Error
	if !errors.As(err, &target) || target == nil {
		return nil, false
	}
	return target, true
}

// LocalOptions configures the opt-in local adapter. Root is a containment
// boundary, not an OS sandbox. MaxOutputBytes and ProcessOutputBytes are hard
// adapter ceilings; individual calls may request a lower command-output bound.
type LocalOptions struct {
	Root               string
	DefaultTimeout     time.Duration
	MaxOutputBytes     int
	ProcessOutputBytes int
}

// LocalAdapter performs explicit local side effects. Constructing it does not
// execute a command and does not read credentials or provider configuration.
type LocalAdapter struct {
	root               string
	defaultTimeout     time.Duration
	maxOutputBytes     int
	processOutputBytes int
}

// NewLocalAdapter constructs an opt-in local adapter. Root defaults to the
// current directory and is resolved again for every call so symlink escapes are
// rejected at the side-effect boundary.
func NewLocalAdapter(options LocalOptions) *LocalAdapter {
	root := strings.TrimSpace(options.Root)
	if root == "" {
		root = "."
	}
	maxOutput := options.MaxOutputBytes
	if maxOutput <= 0 {
		maxOutput = defaultMaxOutputBytes
	}
	processOutput := options.ProcessOutputBytes
	if processOutput <= 0 {
		processOutput = defaultProcessOutputBytes
	}
	return &LocalAdapter{
		root:               root,
		defaultTimeout:     options.DefaultTimeout,
		maxOutputBytes:     maxOutput,
		processOutputBytes: processOutput,
	}
}

// Command describes one non-interactive shell command. Workdir may be relative
// to LocalOptions.Root or an absolute path within that root. Env entries extend
// the current process environment for this command only.
type Command struct {
	Command        string
	Workdir        string
	Env            map[string]string
	Timeout        time.Duration
	MaxOutputBytes int
}

// Termination records how local process cleanup was attempted.
type Termination struct {
	Reason       string `json:"reason,omitempty"`
	Signal       string `json:"signal,omitempty"`
	ProcessGroup bool   `json:"process_group,omitempty"`
	WaitDelayMs  int    `json:"wait_delay_ms,omitempty"`
}

// CommandResult is the bounded, structured result of Run. A non-zero process
// exit code is a result, not an adapter error. An adapter-owned timeout is also
// returned as a timed_out result so callers can present partial output.
type CommandResult struct {
	Command         string
	Workdir         string
	ExitCode        int
	Stdout          string
	Stderr          string
	Duration        time.Duration
	StdoutTruncated bool
	StderrTruncated bool
	Status          string
	TimedOut        bool
	Cancelled       bool
	Interrupted     bool
	Termination     *Termination
}

// Run executes one bounded foreground command. It never starts a detached or
// background process. Callers must authorize the command before invoking Run.
func (a *LocalAdapter) Run(ctx context.Context, request Command) (CommandResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	command := strings.TrimSpace(request.Command)
	if command == "" {
		return CommandResult{}, newError(ErrorCodeInvalidCommand, "exec", "command is required", nil)
	}
	workdir, err := resolveDirWithinRoot(a.rootValue(), request.Workdir)
	if err != nil {
		return CommandResult{}, newError(ErrorCodeWorkdirResolution, "exec", err.Error(), err)
	}
	maxOutput := request.MaxOutputBytes
	if maxOutput <= 0 || maxOutput > a.maxOutputValue() {
		maxOutput = a.maxOutputValue()
	}
	timeout := request.Timeout
	if timeout <= 0 {
		timeout = a.defaultTimeoutValue()
	}
	runCtx := ctx
	cancel := func() {}
	if timeout > 0 {
		runCtx, cancel = context.WithTimeout(ctx, timeout)
	}
	defer cancel()

	result := CommandResult{Command: request.Command, Workdir: workdir}
	started := time.Now()
	shell := "sh"
	args := []string{"-lc", request.Command}
	if stdruntime.GOOS == "windows" {
		shell = "cmd"
		args = []string{"/C", request.Command}
	}
	cmd := exec.CommandContext(runCtx, shell, args...)
	control := configureCommandProcessControl(cmd)
	cmd.Dir = workdir
	if len(request.Env) > 0 {
		cmd.Env = append(os.Environ(), encodeEnv(request.Env)...)
	}
	stdout := &limitedWriter{max: maxOutput}
	stderr := &limitedWriter{max: maxOutput}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	runErr := cmd.Run()
	result.Duration = time.Since(started)
	result.Stdout = stdout.String()
	result.Stderr = stderr.String()
	result.StdoutTruncated = stdout.Truncated()
	result.StderrTruncated = stderr.Truncated()
	if runErr == nil {
		return result, nil
	}
	if runCtx.Err() != nil {
		result.ExitCode = -1
		result.Termination = contextTermination(runCtx.Err(), control)
		if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
			result.Status = "timed_out"
			result.TimedOut = true
		} else {
			result.Status = "cancelled"
			result.Cancelled = true
		}
		if ctx.Err() == nil && result.TimedOut {
			return result, nil
		}
		return result, contextError("exec", runCtx.Err())
	}
	var exitErr *exec.ExitError
	if errors.As(runErr, &exitErr) {
		result.ExitCode = exitErr.ExitCode()
		if termination, ok := commandExitTermination(runErr, control); ok {
			result.Status = "interrupted"
			result.Interrupted = true
			result.Termination = termination
		}
		return result, nil
	}
	return result, newError(ErrorCodeCommandFailed, "exec", runErr.Error(), runErr)
}

// ListRequest configures read-only local process inspection.
type ListRequest struct {
	Limit int
}

// ListResult contains normalized non-empty process listing lines.
type ListResult struct {
	Lines     []string
	Truncated bool
}

// List returns a bounded local process listing. It does not start, stop or
// signal a process.
func (a *LocalAdapter) List(ctx context.Context, request ListRequest) (ListResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	limit := request.Limit
	if limit <= 0 {
		limit = 20
	}
	var cmd *exec.Cmd
	if stdruntime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "tasklist")
	} else {
		cmd = exec.CommandContext(ctx, "ps", "-eo", "pid,ppid,state,etime,command")
	}
	stdout := &limitedWriter{max: a.processOutputValue()}
	stderr := &limitedWriter{max: 256 << 10}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	if ctx.Err() != nil {
		return ListResult{}, contextError("process_list", ctx.Err())
	}
	if stdout.Truncated() || stderr.Truncated() {
		cause := fmt.Errorf("process listing exceeded bounded capture")
		return ListResult{}, newError(ErrorCodeOutputLimitExceeded, "process_list", cause.Error(), cause)
	}
	if err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		return ListResult{}, newError(ErrorCodeProcessListFailed, "process_list", message, err)
	}
	linesRaw := strings.Split(strings.ReplaceAll(stdout.String(), "\r\n", "\n"), "\n")
	lines := make([]string, 0, len(linesRaw))
	for _, line := range linesRaw {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			lines = append(lines, trimmed)
		}
	}
	if len(lines) <= limit {
		return ListResult{Lines: lines}, nil
	}
	return ListResult{Lines: lines[:limit], Truncated: true}, nil
}

func (a *LocalAdapter) rootValue() string {
	if a == nil || strings.TrimSpace(a.root) == "" {
		return "."
	}
	return a.root
}

func (a *LocalAdapter) maxOutputValue() int {
	if a == nil || a.maxOutputBytes <= 0 {
		return defaultMaxOutputBytes
	}
	return a.maxOutputBytes
}

func (a *LocalAdapter) processOutputValue() int {
	if a == nil || a.processOutputBytes <= 0 {
		return defaultProcessOutputBytes
	}
	return a.processOutputBytes
}

func (a *LocalAdapter) defaultTimeoutValue() time.Duration {
	if a == nil {
		return 0
	}
	return a.defaultTimeout
}

func newError(code ErrorCode, op string, message string, cause error) error {
	return &Error{Code: code, Op: op, Message: message, Cause: cause}
}

func contextError(op string, err error) error {
	code := ErrorCodeContextCanceled
	if errors.Is(err, context.DeadlineExceeded) {
		code = ErrorCodeContextDeadlineExceeded
	}
	return newError(code, op, err.Error(), err)
}

func encodeEnv(env map[string]string) []string {
	if len(env) == 0 {
		return nil
	}
	out := make([]string, 0, len(env))
	for key, value := range env {
		if name := strings.TrimSpace(key); name != "" {
			out = append(out, name+"="+value)
		}
	}
	return out
}

type limitedWriter struct {
	mu        sync.Mutex
	max       int
	buf       bytes.Buffer
	truncated bool
}

func (w *limitedWriter) Write(payload []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	written := len(payload)
	remaining := w.max - w.buf.Len()
	if remaining <= 0 {
		if written > 0 {
			w.truncated = true
		}
		return written, nil
	}
	if len(payload) > remaining {
		w.truncated = true
		payload = payload[:remaining]
	}
	_, _ = w.buf.Write(payload)
	return written, nil
}

func (w *limitedWriter) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.String()
}

func (w *limitedWriter) Truncated() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.truncated
}

func resolveDirWithinRoot(root string, target string) (string, error) {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve root: %w", err)
	}
	rootAbs = filepath.Clean(rootAbs)
	rootReal, err := filepath.EvalSymlinks(rootAbs)
	if err != nil {
		return "", fmt.Errorf("resolve root symlinks: %w", err)
	}
	candidate := strings.TrimSpace(target)
	if candidate == "" {
		candidate = rootAbs
	} else if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(rootAbs, candidate)
	}
	candidate, err = filepath.Abs(candidate)
	if err != nil {
		return "", fmt.Errorf("resolve workdir: %w", err)
	}
	candidate = filepath.Clean(candidate)
	realCandidate, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		return "", fmt.Errorf("resolve workdir symlinks: %w", err)
	}
	rel, err := filepath.Rel(filepath.Clean(rootReal), filepath.Clean(realCandidate))
	if err != nil {
		return "", fmt.Errorf("resolve workdir relative path: %w", err)
	}
	rel = filepath.Clean(rel)
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("workdir escapes root: %s", target)
	}
	info, err := os.Stat(candidate)
	if err != nil {
		return "", fmt.Errorf("stat workdir: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("workdir is not a directory: %s", target)
	}
	return candidate, nil
}
