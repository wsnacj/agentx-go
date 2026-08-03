//go:build !windows

package process

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
)

func configureCommandProcessControl(cmd *exec.Cmd) processControl {
	control := processControl{
		ProcessGroup: true,
		CancelSignal: "KILL",
		WaitDelayMs:  int(defaultWaitDelay.Milliseconds()),
	}
	if cmd == nil {
		return control
	}
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setpgid = true
	cmd.WaitDelay = defaultWaitDelay
	cmd.Cancel = func() error {
		if cmd.Process == nil || cmd.Process.Pid <= 0 {
			return os.ErrProcessDone
		}
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if err == nil {
			return nil
		}
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	return control
}

func commandExitTermination(err error, control processControl) (*Termination, bool) {
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr == nil {
		return nil, false
	}
	status, ok := exitErr.Sys().(syscall.WaitStatus)
	if !ok || !status.Signaled() {
		return nil, false
	}
	return exitTermination("interrupted", signalName(status.Signal()), control), true
}

func signalName(signal syscall.Signal) string {
	switch signal {
	case syscall.SIGINT:
		return "INT"
	case syscall.SIGTERM:
		return "TERM"
	case syscall.SIGKILL:
		return "KILL"
	default:
		return signal.String()
	}
}
