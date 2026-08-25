//go:build darwin || linux

package install

import (
	"errors"
	"fmt"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

func openInstallationLockFile(root, name string) (*os.File, error) {
	rootInfo, err := os.Lstat(root)
	if err != nil {
		return nil, fmt.Errorf("inspecting installation lock root %q: %w", root, err)
	}
	if rootInfo.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("installation lock root %q is a symbolic link", root)
	}
	rootFD, err := unix.Open(root, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		if errors.Is(err, unix.ELOOP) {
			return nil, fmt.Errorf("installation lock root %q is a symbolic link", root)
		}
		return nil, fmt.Errorf("opening installation lock root %q: %w", root, err)
	}
	defer unix.Close(rootFD)

	var rootStat unix.Stat_t
	if err := unix.Fstat(rootFD, &rootStat); err != nil {
		return nil, fmt.Errorf("inspecting installation lock root %q: %w", root, err)
	}
	if err := validateInstallationLockRoot(root, rootStat); err != nil {
		return nil, err
	}

	fileFD, err := unix.Openat(
		rootFD,
		name,
		unix.O_CREAT|unix.O_RDWR|unix.O_CLOEXEC|unix.O_NOFOLLOW,
		0o600,
	)
	if err != nil {
		if errors.Is(err, unix.ELOOP) {
			return nil, fmt.Errorf("installation lock file %q is a symbolic link", name)
		}
		return nil, fmt.Errorf("opening installation lock file %q: %w", name, err)
	}
	valid := false
	defer func() {
		if !valid {
			_ = unix.Close(fileFD)
		}
	}()

	var fileStat unix.Stat_t
	if err := unix.Fstat(fileFD, &fileStat); err != nil {
		return nil, fmt.Errorf("inspecting installation lock file %q: %w", name, err)
	}
	if err := validateInstallationLockFile(name, fileStat); err != nil {
		return nil, err
	}
	valid = true
	return os.NewFile(uintptr(fileFD), name), nil
}

func validateInstallationLockRoot(path string, stat unix.Stat_t) error {
	if stat.Mode&unix.S_IFMT != unix.S_IFDIR {
		return fmt.Errorf("installation lock root %q is not a directory", path)
	}
	if stat.Uid != uint32(os.Geteuid()) {
		return fmt.Errorf("installation lock root %q is not owned by the current user", path)
	}
	if permissions := stat.Mode & 0o7777; permissions != 0o700 {
		return fmt.Errorf("installation lock root %q has insecure permissions %#o; want 0700", path, permissions)
	}
	return nil
}

func validateInstallationLockFile(name string, stat unix.Stat_t) error {
	if stat.Mode&unix.S_IFMT != unix.S_IFREG {
		return fmt.Errorf("installation lock file %q is not a regular file", name)
	}
	if stat.Uid != uint32(os.Geteuid()) {
		return fmt.Errorf("installation lock file %q is not owned by the current user", name)
	}
	if permissions := stat.Mode & 0o7777; permissions != 0o600 {
		return fmt.Errorf("installation lock file %q has insecure permissions %#o; want 0600", name, permissions)
	}
	if stat.Nlink != 1 {
		return fmt.Errorf("installation lock file %q has unexpected link count %d; want 1", name, stat.Nlink)
	}
	return nil
}

func lockInstallationFile(file *os.File) error {
	return syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
}

func unlockInstallationFile(file *os.File) error {
	return syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
}

func isInstallationLockBusy(err error) bool {
	return errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN)
}
