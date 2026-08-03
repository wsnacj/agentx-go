package videoframes

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"sync"
)

type commandLimits struct{ stdoutBytes, stderrBytes int64 }

type commandResult struct {
	stdout, stderr                 []byte
	exitCode                       int
	stdoutObserved, stderrObserved int64
	stdoutExceeded, stderrExceeded bool
}

func runBoundedCommand(command *exec.Cmd, limits commandLimits) (commandResult, error) {
	if command == nil {
		return commandResult{exitCode: -1}, fmt.Errorf("command capture requires a command")
	}
	if command.Stdout != nil || command.Stderr != nil {
		return commandResult{exitCode: -1}, fmt.Errorf("command capture requires unconfigured stdout and stderr")
	}
	stdout := newCaptureWriter(limits.stdoutBytes)
	stderr := newCaptureWriter(limits.stderrBytes)
	command.Stdout, command.Stderr = stdout, stderr
	runErr := command.Run()
	result := commandResult{stdout: stdout.bytes(), stderr: stderr.bytes(), exitCode: commandExitCode(runErr), stdoutObserved: stdout.observedBytes(), stderrObserved: stderr.observedBytes(), stdoutExceeded: stdout.limitExceeded(), stderrExceeded: stderr.limitExceeded()}
	var errs []error
	if runErr != nil {
		errs = append(errs, runErr)
	}
	if result.stdoutExceeded {
		errs = append(errs, fmt.Errorf("command stdout output limit exceeded: observed %d bytes (limit %d)", result.stdoutObserved, limits.stdoutBytes))
	}
	if result.stderrExceeded {
		errs = append(errs, fmt.Errorf("command stderr output limit exceeded: observed %d bytes (limit %d)", result.stderrObserved, limits.stderrBytes))
	}
	return result, errors.Join(errs...)
}

func (r commandResult) summary() string {
	status := "completed"
	if r.stdoutExceeded || r.stderrExceeded {
		status = "output_limit_exceeded"
	} else if r.exitCode != 0 {
		status = "failed"
	}
	return fmt.Sprintf("command_%s exit_code=%d stdout_bytes=%d stderr_bytes=%d stdout_limit_exceeded=%t stderr_limit_exceeded=%t", status, r.exitCode, r.stdoutObserved, r.stderrObserved, r.stdoutExceeded, r.stderrExceeded)
}

type captureWriter struct {
	mu                 sync.Mutex
	maxBytes, observed int64
	buffer             bytes.Buffer
}

func newCaptureWriter(maxBytes int64) *captureWriter { return &captureWriter{maxBytes: maxBytes} }

func (w *captureWriter) Write(payload []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.observed += int64(len(payload))
	remaining := w.maxBytes - int64(w.buffer.Len())
	if remaining > 0 {
		captureBytes := int64(len(payload))
		if captureBytes > remaining {
			captureBytes = remaining
		}
		_, _ = w.buffer.Write(payload[:int(captureBytes)])
	}
	return len(payload), nil
}

func (w *captureWriter) bytes() []byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	return append([]byte(nil), w.buffer.Bytes()...)
}
func (w *captureWriter) observedBytes() int64 { w.mu.Lock(); defer w.mu.Unlock(); return w.observed }
func (w *captureWriter) limitExceeded() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.observed > w.maxBytes
}

func commandExitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return -1
}
