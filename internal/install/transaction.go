package install

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type transactionEntry struct {
	file        preparedFile
	path        string
	allowlist   managedPathAllowlist
	before      fileSnapshot
	after       fileSnapshot
	published   bool
	backupPath  string
	backup      fileSnapshot
	expectation FileExpectation
	guarded     bool
}

type fileSnapshot struct {
	exists  bool
	content []byte
	info    os.FileInfo
}

// installationTransaction provides conditional rollback for paths protected
// by the cooperative installation lock. Before each replace or removal it
// rechecks the canonical path, inode, mode, and bytes. Portable filesystem APIs
// do not provide an atomic content compare-and-swap against arbitrary writers,
// so a non-cooperating process can still race the final rename or removal.
type installationTransaction struct {
	entries []*transactionEntry
}

var applyTransactionFile = applyTransactionEntry

func validatePreparedFiles(allowlist managedPathAllowlist, files []preparedFile) error {
	_, err := validatedPreparedPaths(allowlist, files)
	return err
}

func validatedPreparedPaths(allowlist managedPathAllowlist, files []preparedFile) ([]string, error) {
	seen := make(map[string]struct{}, len(files))
	paths := make([]string, 0, len(files))
	for _, file := range files {
		path, err := allowlist.validatePrepared(file.path)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[path]; ok {
			return nil, fmt.Errorf("managed path %q appears more than once in one installation", file.path)
		}
		seen[path] = struct{}{}
		paths = append(paths, path)
	}
	return paths, nil
}

func newInstallationTransaction(
	allowlist managedPathAllowlist,
	files []preparedFile,
	expectations map[string]FileExpectation,
) (*installationTransaction, error) {
	paths, err := validatedPreparedPaths(allowlist, files)
	if err != nil {
		return nil, err
	}
	transaction := &installationTransaction{entries: make([]*transactionEntry, 0, len(files))}
	for index, file := range files {
		entry := &transactionEntry{file: file, path: paths[index], allowlist: allowlist}
		entry.expectation, entry.guarded = expectations[file.path]
		before, err := captureFileSnapshot(entry.path, entry.allowlist)
		if err != nil {
			return nil, fmt.Errorf("capturing before-image %q: %w", entry.path, err)
		}
		entry.before = before
		if err := entry.verifyExpectation(before, "after planning"); err != nil {
			return nil, err
		}
		transaction.entries = append(transaction.entries, entry)
	}
	return transaction, nil
}

// verifyExpectation checks a guarded managed destination against the digest or
// absence the caller planned from, translating a snapshot into the read result
// shape verifyFileExpectation validates.
func (entry *transactionEntry) verifyExpectation(snapshot fileSnapshot, phase string) error {
	if !entry.guarded {
		return nil
	}
	content := snapshot.content
	var readErr error
	if !snapshot.exists {
		content = nil
		readErr = os.ErrNotExist
	}
	return verifyFileExpectation(entry.file.path, content, readErr, entry.expectation, true, phase)
}

// publishedDigests reports the bytes each managed destination holds as this
// transaction completed: the readback-verified published content for written
// entries, the before-image for entries left unchanged. Computed from the
// in-memory transaction, it cannot observe writes by any later process.
func (transaction *installationTransaction) publishedDigests() map[string]string {
	digests := make(map[string]string, len(transaction.entries))
	for _, entry := range transaction.entries {
		content := entry.before.content
		if entry.published {
			content = entry.file.content
		}
		digests[entry.file.path] = fmt.Sprintf("%x", sha256.Sum256(content))
	}
	return digests
}

func (transaction *installationTransaction) apply() ([]fileWriteResult, error) {
	results := make([]fileWriteResult, 0, len(transaction.entries))
	for _, entry := range transaction.entries {
		result, err := applyTransactionFile(entry)
		if err == nil {
			results = append(results, result)
			continue
		}
		rollbackErr := transaction.rollback()
		if rollbackErr != nil {
			return nil, fmt.Errorf(
				"applying managed file %q: %w; rollback was incomplete: %v",
				entry.file.path,
				err,
				rollbackErr,
			)
		}
		return nil, fmt.Errorf(
			"applying managed file %q: %w; managed file changes were rolled back",
			entry.file.path,
			err,
		)
	}
	return results, nil
}

func applyTransactionEntry(entry *transactionEntry) (fileWriteResult, error) {
	if err := entry.verifySnapshot(entry.before, "managed file conflict"); err != nil {
		return fileWriteResult{}, err
	}
	if entry.before.exists && entry.file.contentMatches(entry.before.content) {
		return fileWriteResult{}, nil
	}

	result := fileWriteResult{changed: true, created: !entry.before.exists}
	if entry.before.exists {
		backupPath, err := writeBackup(entry.path, entry.before.content)
		if backupPath != "" {
			entry.backupPath = backupPath
			result.backupPath = backupPath
			backup, snapshotErr := captureFileSnapshot(backupPath, entry.allowlist)
			if snapshotErr != nil {
				return result, errors.Join(
					backupWriteError(err),
					fmt.Errorf("capturing backup identity %q: %w", backupPath, snapshotErr),
				)
			}
			if !backup.exists {
				return result, errors.Join(
					backupWriteError(err),
					fmt.Errorf("capturing backup identity %q: backup is absent", backupPath),
				)
			}
			entry.backup = backup
		}
		if err != nil {
			return result, fmt.Errorf("writing backup: %w", err)
		}
	}

	publication, err := writeFileAtomically(
		entry.path,
		entry.file.content,
		entry.file.perm,
		func() error {
			if err := beforeFilePublish(entry.file.path); err != nil {
				return err
			}
			if entry.guarded {
				current, readErr := os.ReadFile(entry.file.path)
				if err := verifyFileExpectation(
					entry.file.path, current, readErr, entry.expectation, true, "before publication",
				); err != nil {
					return err
				}
			}
			return entry.verifySnapshot(entry.before, "managed file conflict")
		},
	)
	if publication.landed {
		entry.after = fileSnapshot{
			exists:  true,
			content: append([]byte(nil), entry.file.content...),
			info:    publication.info,
		}
		entry.published = true
	}
	if err != nil {
		return result, err
	}
	return result, nil
}

func (transaction *installationTransaction) rollback() error {
	var failures []error
	for index := len(transaction.entries) - 1; index >= 0; index-- {
		entry := transaction.entries[index]
		if entry.published {
			if err := entry.verifySnapshot(entry.after, "rollback conflict"); err != nil {
				if entry.backupPath != "" {
					err = fmt.Errorf("%w; before-image backup preserved at %q", err, entry.backupPath)
				}
				failures = append(failures, err)
				continue
			}
			if entry.before.exists {
				_, err := writeFileAtomically(
					entry.path,
					entry.before.content,
					entry.before.info.Mode().Perm(),
					func() error { return entry.verifySnapshot(entry.after, "rollback conflict") },
				)
				if err != nil {
					failures = append(failures, fmt.Errorf("restoring %q: %w", entry.path, err))
					continue
				}
			} else {
				if err := entry.verifySnapshot(entry.after, "rollback conflict"); err != nil {
					failures = append(failures, err)
					continue
				}
				if err := os.Remove(entry.path); err != nil {
					failures = append(failures, fmt.Errorf("removing created file %q: %w", entry.path, err))
					continue
				}
				if err := syncDirectory(filepath.Dir(entry.path)); err != nil {
					failures = append(failures, fmt.Errorf("syncing removal of %q: %w", entry.path, err))
					continue
				}
			}
			entry.published = false
		}
		if entry.backupPath != "" {
			if err := removeBackupAfterRollback(entry.backupPath, entry.backup, entry.allowlist); err != nil {
				failures = append(failures, err)
			}
		}
	}
	return errors.Join(failures...)
}

func backupWriteError(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("writing backup: %w", err)
}

// removeBackupAfterRollback conditionally removes the exact backup inode that
// this transaction published. As with managed-file rollback, the cooperative
// lock excludes other installers but portable filesystem APIs cannot make the
// final identity check and removal atomic against writers that ignore the lock.
func removeBackupAfterRollback(path string, expected fileSnapshot, allowlist managedPathAllowlist) error {
	if !expected.exists {
		return fmt.Errorf("rollback conflict for backup %q: backup identity was not captured", path)
	}
	current, err := captureFileSnapshot(path, allowlist)
	if err != nil {
		return fmt.Errorf("rollback conflict for backup %q: %w", path, err)
	}
	if !current.exists {
		return nil
	}
	if !sameFileIdentity(expected.info, current.info) || !bytes.Equal(current.content, expected.content) {
		return fmt.Errorf("rollback conflict for backup %q: backup identity, mode, or content changed", path)
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("removing rollback backup %q: %w", path, err)
	}
	if err := syncDirectory(filepath.Dir(path)); err != nil {
		return fmt.Errorf("syncing rollback backup removal %q: %w", path, err)
	}
	return nil
}

func captureFileSnapshot(path string, allowlist managedPathAllowlist) (fileSnapshot, error) {
	resolved, err := allowlist.validate(path)
	if err != nil {
		return fileSnapshot{}, err
	}
	if resolved != path {
		return fileSnapshot{}, fmt.Errorf("managed path %q changed after validation", path)
	}
	before, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return fileSnapshot{}, nil
	}
	if err != nil {
		return fileSnapshot{}, err
	}
	if !before.Mode().IsRegular() {
		return fileSnapshot{}, fmt.Errorf("managed path %q is not a regular file", path)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return fileSnapshot{}, err
	}
	after, err := os.Lstat(path)
	if err != nil {
		return fileSnapshot{}, err
	}
	if !sameFileIdentity(before, after) {
		return fileSnapshot{}, fmt.Errorf("managed path %q changed while it was read", path)
	}
	return fileSnapshot{exists: true, content: content, info: after}, nil
}

func (entry *transactionEntry) verifySnapshot(expected fileSnapshot, conflict string) error {
	current, err := captureFileSnapshot(entry.path, entry.allowlist)
	if err != nil {
		return fmt.Errorf("%s for %q: %w", conflict, entry.path, err)
	}
	if current.exists != expected.exists {
		if expected.exists {
			return fmt.Errorf("%s for %q: expected file to exist", conflict, entry.path)
		}
		return fmt.Errorf("%s for %q: expected path to remain absent", conflict, entry.path)
	}
	if !expected.exists {
		return nil
	}
	if !sameFileIdentity(expected.info, current.info) {
		return fmt.Errorf("%s for %q: file identity or mode changed", conflict, entry.path)
	}
	if !bytes.Equal(current.content, expected.content) {
		return fmt.Errorf("%s for %q: content changed", conflict, entry.path)
	}
	return nil
}

type managedPathAllowlist struct {
	requestedRoot string
	root          string
}

func newLockedManagedPathAllowlist(configDir, lockedRoot string) (managedPathAllowlist, error) {
	requestedRoot, err := filepath.Abs(filepath.Clean(configDir))
	if err != nil {
		return managedPathAllowlist{}, fmt.Errorf("resolving requested installation config directory: %w", err)
	}
	return managedPathAllowlist{requestedRoot: requestedRoot, root: lockedRoot}, nil

}

// validatePrepared maps the lexical path produced from the requested config
// directory onto the canonical root whose lock is held. It intentionally does
// not follow the requested root again: a symlink retarget after acquisition
// must not change the destination, even when the new target remains below the
// locked root.
func (allowlist managedPathAllowlist) validatePrepared(path string) (string, error) {
	absolute, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", fmt.Errorf("resolving prepared managed path %q: %w", path, err)
	}
	relative, err := filepath.Rel(allowlist.requestedRoot, absolute)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("managed path %q is outside installation config directory %q", path, allowlist.requestedRoot)
	}
	return allowlist.validate(filepath.Join(allowlist.root, relative))
}

func (allowlist managedPathAllowlist) validate(path string) (string, error) {
	resolved, err := resolveExistingPath(path)
	if err != nil {
		return "", fmt.Errorf("resolving managed path %q: %w", path, err)
	}
	relative, err := filepath.Rel(allowlist.root, resolved)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("managed path %q is outside installation config directory %q", path, allowlist.root)
	}
	return resolved, nil
}

func resolveExistingPath(path string) (string, error) {
	absolute, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", err
	}
	probe := absolute
	var missing []string
	for {
		if _, err := os.Lstat(probe); err == nil {
			resolved, err := filepath.EvalSymlinks(probe)
			if err != nil {
				return "", err
			}
			for index := len(missing) - 1; index >= 0; index-- {
				resolved = filepath.Join(resolved, missing[index])
			}
			return filepath.Clean(resolved), nil
		} else if !os.IsNotExist(err) {
			return "", err
		}
		parent := filepath.Dir(probe)
		if parent == probe {
			return absolute, nil
		}
		missing = append(missing, filepath.Base(probe))
		probe = parent
	}
}
