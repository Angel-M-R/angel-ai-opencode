//go:build darwin || linux

package commandrunner

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
)

func configureProcessTermination(command *exec.Cmd) {
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		return terminateProcessTree(command.Process)
	}
}

func terminateProcessTree(process *os.Process) error {
	if process == nil {
		return os.ErrProcessDone
	}
	err := syscall.Kill(-process.Pid, syscall.SIGKILL)
	if errors.Is(err, syscall.ESRCH) {
		return os.ErrProcessDone
	}
	return err
}
