//go:build !darwin && !linux

package install

import (
	"errors"
	"os"
)

var errInstallationLockUnsupported = errors.New("installation lock is unsupported on this platform")

func openInstallationLockFile(string, string) (*os.File, error) {
	return nil, errInstallationLockUnsupported
}

func lockInstallationFile(*os.File) error {
	return errInstallationLockUnsupported
}

func unlockInstallationFile(*os.File) error {
	return nil
}

func isInstallationLockBusy(error) bool {
	return false
}
