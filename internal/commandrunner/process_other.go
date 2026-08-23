//go:build !darwin && !linux

package commandrunner

import (
	"errors"
	"os"
	"os/exec"
)

// Unsupported platforms retain the standard direct-child cancellation. They
// do not claim process-tree termination because their process model needs a
// platform-specific implementation.
func configureProcessTermination(command *exec.Cmd) {
	command.Cancel = func() error {
		if command.Process == nil {
			return os.ErrProcessDone
		}
		err := command.Process.Kill()
		if errors.Is(err, os.ErrProcessDone) {
			return os.ErrProcessDone
		}
		return err
	}
}
