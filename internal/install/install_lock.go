package install

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var errInstallationLockBusy = errors.New("installation lock is busy")

type installationLock struct {
	file   *os.File
	target string
	once   sync.Once
	err    error
}

func acquireInstallationLock(configDir string) (*installationLock, error) {
	target, err := resolveExistingPath(configDir)
	if err != nil {
		return nil, fmt.Errorf("resolving installation config directory: %w", err)
	}
	root, err := resolveInstallationLockRoot()
	if err != nil {
		return nil, fmt.Errorf("resolving installation lock directory: %w", err)
	}
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, fmt.Errorf("creating installation lock directory: %w", err)
	}
	name := installationLockFileName(target)
	// The file may outlive the process. Ownership comes from the kernel lock,
	// not from creating or deleting this path.
	file, err := openInstallationLockFile(root, name)
	if err != nil {
		return nil, fmt.Errorf("opening secure installation lock: %w", err)
	}
	if err := lockInstallationFile(file); err != nil {
		_ = file.Close()
		if isInstallationLockBusy(err) {
			return nil, fmt.Errorf("%w for %s", errInstallationLockBusy, target)
		}
		return nil, fmt.Errorf("acquiring installation lock for %s: %w", target, err)
	}
	return &installationLock{file: file, target: target}, nil
}

var resolveInstallationLockRoot = defaultInstallationLockRoot

func defaultInstallationLockRoot() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, "angel-ai-opencode-install-locks"), nil
}

func installationLockFileName(target string) string {
	key := sha256.Sum256([]byte(target))
	return hex.EncodeToString(key[:]) + ".lock"
}

func (lock *installationLock) release() error {
	if lock == nil {
		return nil
	}
	lock.once.Do(func() {
		lock.err = errors.Join(unlockInstallationFile(lock.file), lock.file.Close())
		lock.file = nil
	})
	return lock.err
}
