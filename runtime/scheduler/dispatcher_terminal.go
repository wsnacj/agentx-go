package scheduler

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

const defaultTerminalPersistenceTimeout = 5 * time.Second

var errTerminalTransitionSuperseded = errors.New("scheduler terminal transition superseded")

func (d *Dispatcher) failJobTerminal(parent context.Context, lane Lane, operation string, result Result) {
	if d == nil || d.queue == nil {
		return
	}
	ctx, cancel := d.terminalContext(parent)
	defer cancel()
	attempts := 1
	if resultOutcome(result, ResultOutcomeFailed) == ResultOutcomeCanceled {
		attempts = 2
	}
	var transitionErr error
	for attempt := 0; attempt < attempts; attempt++ {
		transitionErr = d.queue.Fail(ctx, result)
		if transitionErr == nil {
			d.recordPostFailOutcome(ctx, lane, result.JobID)
			return
		}
		state := d.reconcileTerminalState(ctx, result.JobID, resultOutcome(result, ResultOutcomeFailed))
		if state == terminalStateDesired || state == terminalStateSuperseded {
			d.metrics.RecordFailFailed(lane)
			d.recordPostFailOutcome(ctx, lane, result.JobID)
			return
		}
		if ctx.Err() != nil {
			break
		}
	}
	d.metrics.RecordFailFailed(lane)
	d.recordTerminalPersistenceError(operation, lane, result, transitionErr)
}

func (d *Dispatcher) ackJobTerminal(parent context.Context, lane Lane, result Result) error {
	if d == nil || d.queue == nil {
		return ErrQueueUnavailable
	}
	ctx, cancel := d.terminalContext(parent)
	defer cancel()
	var transitionErr error
	for attempt := 0; attempt < 2; attempt++ {
		transitionErr = d.queue.Ack(ctx, result)
		if transitionErr == nil {
			return nil
		}
		switch d.reconcileTerminalState(ctx, result.JobID, ResultOutcomeCompleted) {
		case terminalStateDesired:
			return errTerminalTransitionSuperseded
		case terminalStateSuperseded:
			return errTerminalTransitionSuperseded
		}
		if ctx.Err() != nil {
			break
		}
	}
	return transitionErr
}

func (d *Dispatcher) terminalContext(parent context.Context) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	timeout := d.terminal
	if timeout <= 0 {
		timeout = defaultTerminalPersistenceTimeout
	}
	return context.WithTimeout(context.WithoutCancel(parent), timeout)
}

type terminalState uint8

const (
	terminalStateUnresolved terminalState = iota
	terminalStateDesired
	terminalStateSuperseded
)

func (d *Dispatcher) reconcileTerminalState(ctx context.Context, jobID string, desired ResultOutcome) terminalState {
	if d == nil || d.queue == nil || strings.TrimSpace(jobID) == "" {
		return terminalStateUnresolved
	}
	if result, ok, err := d.queue.Result(ctx, jobID); err == nil && ok {
		if resultOutcome(result, "") == desired {
			return terminalStateDesired
		}
		return terminalStateSuperseded
	}
	return terminalStateUnresolved
}

func (d *Dispatcher) recordTerminalPersistenceError(operation string, lane Lane, result Result, err error) {
	if d == nil {
		return
	}
	if err == nil {
		err = ErrQueueUnavailable
	}
	operation = strings.TrimSpace(operation)
	if operation == "" {
		operation = "terminal_transition"
	}
	safeErr := fmt.Errorf(
		"scheduler terminal persistence failed: operation=%s lane=%s job_id=%q attempt=%d detail=%s",
		operation,
		lane,
		strings.TrimSpace(result.JobID),
		result.Attempt,
		schedulerFailureSummary(err, "terminal_persistence_failed"),
	)
	d.errMu.Lock()
	defer d.errMu.Unlock()
	d.errCount++
	if d.waitErr == nil {
		d.waitErr = safeErr
	}
}

func (d *Dispatcher) waitError() error {
	if d == nil {
		return nil
	}
	d.errMu.Lock()
	defer d.errMu.Unlock()
	if d.waitErr == nil {
		return nil
	}
	if d.errCount <= 1 {
		return d.waitErr
	}
	return fmt.Errorf("%w (terminal_persistence_failures=%d)", d.waitErr, d.errCount)
}
