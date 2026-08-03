//go:build windows

package process

import "os/exec"

func configureCommandProcessControl(cmd *exec.Cmd) processControl {
	control := processControl{WaitDelayMs: int(defaultWaitDelay.Milliseconds())}
	if cmd != nil {
		cmd.WaitDelay = defaultWaitDelay
	}
	return control
}

func commandExitTermination(error, processControl) (*Termination, bool) {
	return nil, false
}
