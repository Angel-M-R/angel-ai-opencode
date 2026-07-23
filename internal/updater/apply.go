package updater

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
)

const relaunchMarker = "ANGEL_AI_UPDATE_RELAUNCHED"

func (updater *Updater) apply(manifest Manifest) error {
	executable, err := updater.resolvedExecutable()
	if err != nil {
		return err
	}

	candidatePath, err := updater.downloadCandidate(manifest, executable)
	if err != nil {
		return err
	}
	defer updater.fileSystem.Remove(candidatePath)

	backupPath, err := updater.prepareBackup(executable)
	if err != nil {
		return err
	}

	if err := updater.fileSystem.Rename(candidatePath, executable); err != nil {
		if restoreErr := updater.fileSystem.Rename(backupPath, executable); restoreErr != nil {
			return fmt.Errorf("replacing executable: %w; restoring previous executable: %v", err, restoreErr)
		}
		return fmt.Errorf("replacing executable: %w", err)
	}

	arguments := updater.process.Args()
	environment := withRelaunchMarker(updater.process.Environ())
	if err := updater.process.Exec(executable, arguments, environment); err != nil {
		if restoreErr := updater.fileSystem.Rename(backupPath, executable); restoreErr != nil {
			return fmt.Errorf("relaunching replacement: %w; restoring previous executable: %v", err, restoreErr)
		}
		return fmt.Errorf("relaunching replacement: %w", err)
	}

	// A successful syscall.Exec does not return. A nil result from an injected
	// process means the replacement has started, so rollback is no longer ours.
	return nil
}

func (updater *Updater) downloadCandidate(manifest Manifest, executable string) (candidatePath string, resultErr error) {
	directory := filepath.Dir(executable)
	base := filepath.Base(executable)
	candidate, err := updater.fileSystem.CreateTemp(directory, "."+base+".update-*")
	if err != nil {
		return "", fmt.Errorf("creating update temporary file: %w", err)
	}
	candidatePath = candidate.Name()
	closed := false
	defer func() {
		if !closed {
			_ = candidate.Close()
		}
		if resultErr != nil {
			_ = updater.fileSystem.Remove(candidatePath)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), updater.timeout)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, manifest.ArtifactURL, nil)
	if err != nil {
		return candidatePath, fmt.Errorf("creating artifact request: %w", err)
	}
	response, err := updater.http.Do(request)
	if err != nil {
		return candidatePath, fmt.Errorf("downloading update artifact: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return candidatePath, fmt.Errorf("downloading update artifact: unexpected HTTP status %s", response.Status)
	}

	hash := sha256.New()
	if _, err := io.Copy(io.MultiWriter(candidate, hash), response.Body); err != nil {
		return candidatePath, fmt.Errorf("downloading update artifact: %w", err)
	}
	if err := candidate.Close(); err != nil {
		return candidatePath, fmt.Errorf("closing update temporary file: %w", err)
	}
	closed = true

	actualDigest := hex.EncodeToString(hash.Sum(nil))
	if actualDigest != manifest.SHA256 {
		return candidatePath, fmt.Errorf("verifying update artifact: SHA-256 mismatch (got %s)", actualDigest)
	}
	if err := updater.fileSystem.Chmod(candidatePath, 0o755); err != nil {
		return candidatePath, fmt.Errorf("making update artifact executable: %w", err)
	}
	return candidatePath, nil
}

func (updater *Updater) prepareBackup(executable string) (backupPath string, resultErr error) {
	info, err := updater.fileSystem.Stat(executable)
	if err != nil {
		return "", fmt.Errorf("reading executable permissions: %w", err)
	}
	source, err := updater.fileSystem.Open(executable)
	if err != nil {
		return "", fmt.Errorf("opening current executable: %w", err)
	}
	defer source.Close()

	directory := filepath.Dir(executable)
	base := filepath.Base(executable)
	backup, err := updater.fileSystem.CreateTemp(directory, "."+base+".backup-*")
	if err != nil {
		return "", fmt.Errorf("creating executable backup: %w", err)
	}
	temporaryPath := backup.Name()
	closed := false
	defer func() {
		if !closed {
			_ = backup.Close()
		}
		_ = updater.fileSystem.Remove(temporaryPath)
	}()

	if _, err := io.Copy(backup, source); err != nil {
		return "", fmt.Errorf("copying executable backup: %w", err)
	}
	if err := backup.Close(); err != nil {
		return "", fmt.Errorf("closing executable backup: %w", err)
	}
	closed = true
	if err := updater.fileSystem.Chmod(temporaryPath, info.Mode().Perm()); err != nil {
		return "", fmt.Errorf("preserving executable backup permissions: %w", err)
	}

	backupPath = updater.backupPath(executable)
	if err := updater.fileSystem.Rename(temporaryPath, backupPath); err != nil {
		return "", fmt.Errorf("preparing executable backup: %w", err)
	}
	return backupPath, nil
}

func (updater *Updater) completeRelaunch(currentVersion string, forced bool) error {
	executable, err := updater.resolvedExecutable()
	if err != nil {
		return updater.warn("update cleanup failed: %v", err)
	}
	if err := updater.fileSystem.Remove(updater.backupPath(executable)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return updater.warn("update cleanup failed: %v", err)
	}
	if forced {
		_, err := fmt.Fprintf(updater.output, "angel-ai updated to %s\n", currentVersion)
		return err
	}
	return nil
}

func (updater *Updater) resolvedExecutable() (string, error) {
	executable, err := updater.fileSystem.Executable()
	if err != nil {
		return "", fmt.Errorf("locating current executable: %w", err)
	}
	resolved, err := updater.fileSystem.EvalSymlinks(executable)
	if err != nil {
		return "", fmt.Errorf("resolving current executable: %w", err)
	}
	return resolved, nil
}

func (updater *Updater) backupPath(executable string) string {
	return filepath.Join(filepath.Dir(executable), "."+filepath.Base(executable)+".update-backup")
}

func hasRelaunchMarker(environment []string) bool {
	prefix := relaunchMarker + "="
	for _, value := range environment {
		if value == prefix+"1" {
			return true
		}
	}
	return false
}

func withRelaunchMarker(environment []string) []string {
	prefix := relaunchMarker + "="
	result := make([]string, 0, len(environment)+1)
	for _, value := range environment {
		if !strings.HasPrefix(value, prefix) {
			result = append(result, value)
		}
	}
	return append(result, prefix+"1")
}
