package process

import (
	"context"
	"errors"
	"strings"
)

type processControl struct {
	ProcessGroup bool
	CancelSignal string
	WaitDelayMs  int
}

func contextTermination(err error, control processControl) *Termination {
	if err == nil {
		return nil
	}
	reason := "cancelled"
	if errors.Is(err, context.DeadlineExceeded) {
		reason = "timed_out"
	}
	return &Termination{
		Reason: reason, Signal: control.CancelSignal,
		ProcessGroup: control.ProcessGroup, WaitDelayMs: control.WaitDelayMs,
	}
}

func exitTermination(reason string, signal string, control processControl) *Termination {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "interrupted"
	}
	signal = strings.ToUpper(strings.TrimSpace(signal))
	if signal == "" {
		signal = control.CancelSignal
	}
	return &Termination{
		Reason: reason, Signal: signal,
		ProcessGroup: control.ProcessGroup, WaitDelayMs: control.WaitDelayMs,
	}
}
