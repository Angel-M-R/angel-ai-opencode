package install

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type publicationResult struct {
	landed bool
	info   os.FileInfo
}

var readPublishedFile = os.ReadFile

// writeFileAtomically runs precondition after the temporary file is durable
// and immediately before rename. The callback narrows the race with arbitrary
// writers but does not turn rename into a portable atomic compare-and-swap.
func writeFileAtomically(
	path string,
	content []byte,
	perm os.FileMode,
	precondition func() error,
) (publicationResult, error) {
	dir := filepath.Dir(path)
	if err := ensureDirectoryDurable(dir, 0o755); err != nil {
		return publicationResult{}, err
	}
	temp, err := os.CreateTemp(dir, ".angel-ai-*.tmp")
	if err != nil {
		return publicationResult{}, err
	}
	tempPath := temp.Name()
	cleanup := func() {
		_ = temp.Close()
		_ = os.Remove(tempPath)
	}
	if err := temp.Chmod(perm.Perm()); err != nil {
		cleanup()
		return publicationResult{}, err
	}
	if _, err := temp.Write(content); err != nil {
		cleanup()
		return publicationResult{}, err
	}
	if err := temp.Sync(); err != nil {
		cleanup()
		return publicationResult{}, err
	}
	tempInfo, err := temp.Stat()
	if err != nil {
		cleanup()
		return publicationResult{}, err
	}
	if err := temp.Close(); err != nil {
		_ = os.Remove(tempPath)
		return publicationResult{}, err
	}
	if precondition != nil {
		if err := precondition(); err != nil {
			_ = os.Remove(tempPath)
			return publicationResult{}, err
		}
	}
	if err := os.Rename(tempPath, path); err != nil {
		_ = os.Remove(tempPath)
		return publicationResult{}, err
	}
	result := publicationResult{landed: true, info: tempInfo}
	if err := syncDirectory(dir); err != nil {
		return result, fmt.Errorf("syncing published file %q: %w", path, err)
	}
	currentInfo, err := os.Lstat(path)
	if err != nil {
		return result, fmt.Errorf("confirming published identity for %q: %w", path, err)
	}
	if !sameFileIdentity(tempInfo, currentInfo) {
		return result, fmt.Errorf("confirming published identity for %q: file was replaced", path)
	}
	published, err := readPublishedFile(path)
	if err != nil {
		return result, fmt.Errorf("confirming published content for %q: %w", path, err)
	}
	if !bytes.Equal(published, content) {
		return result, fmt.Errorf("confirming published content for %q: readback differs", path)
	}
	confirmedInfo, err := os.Lstat(path)
	if err != nil {
		return result, fmt.Errorf("confirming published identity for %q: %w", path, err)
	}
	if !sameFileIdentity(tempInfo, confirmedInfo) {
		return result, fmt.Errorf("confirming published identity for %q: file changed during readback", path)
	}
	return result, nil
}

func sameFileIdentity(expected, current os.FileInfo) bool {
	return expected != nil && current != nil &&
		expected.Mode().IsRegular() && current.Mode().IsRegular() &&
		expected.Mode().Perm() == current.Mode().Perm() &&
		expected.Size() == current.Size() &&
		expected.ModTime().Equal(current.ModTime()) &&
		os.SameFile(expected, current)
}

func ensureDirectoryDurable(path string, perm os.FileMode) error {
	current := filepath.Clean(path)
	var missing []string
	for {
		info, err := os.Stat(current)
		if err == nil {
			if !info.IsDir() {
				return fmt.Errorf("directory path %q is not a directory", current)
			}
			break
		}
		if !os.IsNotExist(err) {
			return err
		}
		parent := filepath.Dir(current)
		if parent == current {
			return fmt.Errorf("no existing ancestor for directory %q", path)
		}
		missing = append(missing, current)
		current = parent
	}
	for index := len(missing) - 1; index >= 0; index-- {
		directory := missing[index]
		if err := os.Mkdir(directory, perm); err != nil {
			if !os.IsExist(err) {
				return err
			}
			info, statErr := os.Stat(directory)
			if statErr != nil || !info.IsDir() {
				return fmt.Errorf("directory path %q was replaced during creation", directory)
			}
		}
		if err := syncDirectory(filepath.Dir(directory)); err != nil {
			return fmt.Errorf("syncing created directory %q: %w", directory, err)
		}
	}
	return nil
}

func syncDirectory(path string) error {
	directory, err := os.Open(path)
	if err != nil {
		return err
	}
	defer directory.Close()
	return directory.Sync()
}

func writeBackup(targetPath string, content []byte) (string, error) {
	pattern := "." + filepath.Base(targetPath) + ".bak-" + time.Now().Format("20060102-150405") + "-*"
	backup, err := os.CreateTemp(filepath.Dir(targetPath), pattern)
	if err != nil {
		return "", err
	}
	tempPath := backup.Name()
	backupPath := filepath.Join(filepath.Dir(targetPath), strings.TrimPrefix(filepath.Base(tempPath), "."))
	cleanup := func() {
		_ = backup.Close()
		_ = os.Remove(tempPath)
	}
	if err := backup.Chmod(0o600); err != nil {
		cleanup()
		return "", err
	}
	if _, err := backup.Write(content); err != nil {
		cleanup()
		return "", err
	}
	if err := backup.Sync(); err != nil {
		cleanup()
		return "", err
	}
	if err := backup.Close(); err != nil {
		_ = os.Remove(tempPath)
		return "", err
	}
	if err := os.Link(tempPath, backupPath); err != nil {
		_ = os.Remove(tempPath)
		return "", err
	}
	if err := os.Remove(tempPath); err != nil {
		return backupPath, err
	}
	if err := syncDirectory(filepath.Dir(targetPath)); err != nil {
		return backupPath, err
	}
	return backupPath, nil
}
