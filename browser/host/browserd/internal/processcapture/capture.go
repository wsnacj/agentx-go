// Package processcapture provides bounded process output capture for browserd
// command adapters.
package processcapture

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

var ErrLimitExceeded = errors.New("read limit exceeded")

const DefaultStreamLimitBytes int64 = 4 << 20

type Limits struct {
	StdoutBytes int64
	StderrBytes int64
}

type Result struct {
	Stdout              []byte
	Stderr              []byte
	ExitCode            int
	StdoutObservedBytes int64
	StderrObservedBytes int64
	StdoutLimitExceeded bool
	StderrLimitExceeded bool
}

func Run(command *exec.Cmd, limits Limits) (Result, error) {
	if command == nil {
		return Result{ExitCode: -1}, fmt.Errorf("command capture requires a command")
	}
	if command.Stdout != nil || command.Stderr != nil {
		return Result{ExitCode: -1}, fmt.Errorf("command capture requires unconfigured stdout and stderr")
	}
	stdoutLimit, err := normalizeLimit(limits.StdoutBytes)
	if err != nil {
		return Result{ExitCode: -1}, fmt.Errorf("stdout limit: %w", err)
	}
	stderrLimit, err := normalizeLimit(limits.StderrBytes)
	if err != nil {
		return Result{ExitCode: -1}, fmt.Errorf("stderr limit: %w", err)
	}
	stdout := newCaptureWriter(stdoutLimit)
	stderr := newCaptureWriter(stderrLimit)
	command.Stdout = stdout
	command.Stderr = stderr
	runErr := command.Run()
	result := Result{
		Stdout:              stdout.bytes(),
		Stderr:              stderr.bytes(),
		ExitCode:            commandExitCode(runErr),
		StdoutObservedBytes: stdout.observedBytes(),
		StderrObservedBytes: stderr.observedBytes(),
		StdoutLimitExceeded: stdout.limitExceeded(),
		StderrLimitExceeded: stderr.limitExceeded(),
	}
	var errs []error
	if runErr != nil {
		errs = append(errs, runErr)
	}
	if result.StdoutLimitExceeded {
		errs = append(errs, &streamLimitError{stream: "stdout", limitBytes: stdoutLimit, observedBytes: result.StdoutObservedBytes})
	}
	if result.StderrLimitExceeded {
		errs = append(errs, &streamLimitError{stream: "stderr", limitBytes: stderrLimit, observedBytes: result.StderrObservedBytes})
	}
	return result, errors.Join(errs...)
}

func (r Result) Summary() string {
	status := "completed"
	if r.StdoutLimitExceeded || r.StderrLimitExceeded {
		status = "output_limit_exceeded"
	} else if r.ExitCode != 0 {
		status = "failed"
	}
	return fmt.Sprintf(
		"command_%s exit_code=%d stdout_bytes=%d stderr_bytes=%d stdout_limit_exceeded=%t stderr_limit_exceeded=%t",
		status,
		r.ExitCode,
		r.StdoutObservedBytes,
		r.StderrObservedBytes,
		r.StdoutLimitExceeded,
		r.StderrLimitExceeded,
	)
}

type streamLimitError struct {
	stream        string
	limitBytes    int64
	observedBytes int64
}

func (e *streamLimitError) Error() string {
	if e == nil {
		return ErrLimitExceeded.Error()
	}
	return fmt.Sprintf(
		"command %s: %v: observed %d bytes (limit %d)",
		strings.TrimSpace(e.stream),
		ErrLimitExceeded,
		e.observedBytes,
		e.limitBytes,
	)
}

func (e *streamLimitError) Unwrap() error { return ErrLimitExceeded }

type captureWriter struct {
	mu       sync.Mutex
	maxBytes int64
	observed int64
	buffer   bytes.Buffer
}

func newCaptureWriter(maxBytes int64) *captureWriter {
	writer := &captureWriter{maxBytes: maxBytes}
	if maxBytes < 64<<10 {
		writer.buffer.Grow(int(maxBytes))
	}
	return writer
}

func (w *captureWriter) Write(payload []byte) (int, error) {
	if len(payload) == 0 {
		return 0, nil
	}
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

func (w *captureWriter) observedBytes() int64 {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.observed
}

func (w *captureWriter) limitExceeded() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.observed > w.maxBytes
}

func normalizeLimit(value int64) (int64, error) {
	if value == 0 {
		return DefaultStreamLimitBytes, nil
	}
	if value < 0 {
		return 0, fmt.Errorf("limit must not be negative")
	}
	return value, nil
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
