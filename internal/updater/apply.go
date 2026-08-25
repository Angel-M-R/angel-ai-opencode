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

	candidatePath, candidateIdentity, err := updater.downloadCandidate(manifest, executable)
	if err != nil {
		return err
	}
	defer updater.fileSystem.Remove(candidatePath)

	backup, err := updater.prepareBackup(executable)
	if err != nil {
		return err
	}

	if err := updater.publishCandidate(candidatePath, executable, candidateIdentity); err != nil {
		if restoreErr := updater.restoreBackup(backup, executable); restoreErr != nil {
			return fmt.Errorf("replacing executable: %w; restoring previous executable: %v", err, restoreErr)
		}
		return fmt.Errorf("replacing executable: %w", err)
	}

	arguments := updater.process.Args()
	environment := withRelaunchMarker(updater.process.Environ())
	if err := updater.process.Exec(executable, arguments, environment); err != nil {
		if restoreErr := updater.restoreBackup(backup, executable); restoreErr != nil {
			return fmt.Errorf("relaunching replacement: %w; restoring previous executable: %v", err, restoreErr)
		}
		return fmt.Errorf("relaunching replacement: %w", err)
	}

	// A successful syscall.Exec does not return. A nil result from an injected
	// process means the replacement has started, so rollback is no longer ours.
	return nil
}

type fileIdentity struct {
	size   int64
	digest string
}

type persistedFile struct {
	path     string
	identity fileIdentity
}

func (updater *Updater) downloadCandidate(manifest Manifest, executable string) (candidatePath string, identity fileIdentity, resultErr error) {
	directory := filepath.Dir(executable)
	base := filepath.Base(executable)
	candidate, err := updater.fileSystem.CreateTemp(directory, "."+base+".update-*")
	if err != nil {
		return "", fileIdentity{}, fmt.Errorf("creating update temporary file: %w", err)
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

	ctx, cancel := context.WithTimeout(context.Background(), updater.artifactTimeout)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, manifest.ArtifactURL, nil)
	if err != nil {
		return candidatePath, fileIdentity{}, fmt.Errorf("creating artifact request: %w", err)
	}
	response, err := updater.http.Do(request)
	if err != nil {
		return candidatePath, fileIdentity{}, fmt.Errorf("downloading update artifact: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return candidatePath, fileIdentity{}, fmt.Errorf("downloading update artifact: unexpected HTTP status %s", response.Status)
	}
	if response.ContentLength > updater.maxArtifactBytes {
		return candidatePath, fileIdentity{}, fmt.Errorf("downloading update artifact: response exceeds %d bytes", updater.maxArtifactBytes)
	}

	hash := sha256.New()
	written, err := copyAtMost(io.MultiWriter(candidate, hash), response.Body, updater.maxArtifactBytes)
	if err != nil {
		return candidatePath, fileIdentity{}, fmt.Errorf("downloading update artifact: %w", err)
	}
	if err := updater.fileSystem.Chmod(candidatePath, 0o755); err != nil {
		return candidatePath, fileIdentity{}, fmt.Errorf("making update artifact executable: %w", err)
	}
	if err := candidate.Sync(); err != nil {
		return candidatePath, fileIdentity{}, fmt.Errorf("synchronizing update temporary file: %w", err)
	}
	if err := candidate.Close(); err != nil {
		return candidatePath, fileIdentity{}, fmt.Errorf("closing update temporary file: %w", err)
	}
	closed = true

	actualDigest := hex.EncodeToString(hash.Sum(nil))
	if actualDigest != manifest.SHA256 {
		return candidatePath, fileIdentity{}, fmt.Errorf("verifying update artifact: SHA-256 mismatch (got %s)", actualDigest)
	}
	identity, err = updater.readFileIdentity(candidatePath)
	if err != nil {
		return candidatePath, fileIdentity{}, fmt.Errorf("reading update artifact back from disk: %w", err)
	}
	if identity.size != written || identity.digest != manifest.SHA256 {
		return candidatePath, fileIdentity{}, fmt.Errorf("verifying update artifact readback: got %d bytes with SHA-256 %s", identity.size, identity.digest)
	}
	return candidatePath, identity, nil
}

func (updater *Updater) prepareBackup(executable string) (backup persistedFile, resultErr error) {
	info, err := updater.fileSystem.Stat(executable)
	if err != nil {
		return persistedFile{}, fmt.Errorf("reading executable permissions: %w", err)
	}
	source, err := updater.fileSystem.Open(executable)
	if err != nil {
		return persistedFile{}, fmt.Errorf("opening current executable: %w", err)
	}
	defer source.Close()

	directory := filepath.Dir(executable)
	base := filepath.Base(executable)
	temporary, err := updater.fileSystem.CreateTemp(directory, "."+base+".backup-*")
	if err != nil {
		return persistedFile{}, fmt.Errorf("creating executable backup: %w", err)
	}
	temporaryPath := temporary.Name()
	closed := false
	backupPath := updater.backupPath(executable)
	landed := false
	defer func() {
		if !closed {
			_ = temporary.Close()
		}
		_ = updater.fileSystem.Remove(temporaryPath)
		if resultErr != nil && landed {
			_ = updater.fileSystem.Remove(backupPath)
		}
	}()

	hash := sha256.New()
	size, err := io.Copy(io.MultiWriter(temporary, hash), source)
	if err != nil {
		return persistedFile{}, fmt.Errorf("copying executable backup: %w", err)
	}
	if err := updater.fileSystem.Chmod(temporaryPath, info.Mode().Perm()); err != nil {
		return persistedFile{}, fmt.Errorf("preserving executable backup permissions: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		return persistedFile{}, fmt.Errorf("synchronizing executable backup: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return persistedFile{}, fmt.Errorf("closing executable backup: %w", err)
	}
	closed = true
	expected := fileIdentity{size: size, digest: hex.EncodeToString(hash.Sum(nil))}
	readback, err := updater.readFileIdentity(temporaryPath)
	if err != nil {
		return persistedFile{}, fmt.Errorf("reading executable backup back from disk: %w", err)
	}
	if readback != expected {
		return persistedFile{}, fmt.Errorf("executable backup readback mismatch")
	}

	if err := updater.fileSystem.Rename(temporaryPath, backupPath); err != nil {
		return persistedFile{}, fmt.Errorf("preparing executable backup: %w", err)
	}
	landed = true
	if err := updater.fileSystem.SyncDir(directory); err != nil {
		return persistedFile{}, fmt.Errorf("synchronizing executable backup directory: %w", err)
	}
	readback, err = updater.readFileIdentity(backupPath)
	if err != nil {
		return persistedFile{}, fmt.Errorf("confirming executable backup: %w", err)
	}
	if readback != expected {
		return persistedFile{}, fmt.Errorf("executable backup published with unexpected bytes")
	}
	return persistedFile{path: backupPath, identity: expected}, nil
}

func (updater *Updater) publishCandidate(candidatePath, executable string, expected fileIdentity) error {
	if err := updater.fileSystem.Rename(candidatePath, executable); err != nil {
		return err
	}
	if err := updater.fileSystem.SyncDir(filepath.Dir(executable)); err != nil {
		return fmt.Errorf("synchronizing executable directory: %w", err)
	}
	actual, err := updater.readFileIdentity(executable)
	if err != nil {
		return fmt.Errorf("reading replacement back from disk: %w", err)
	}
	if actual != expected {
		return fmt.Errorf("replacement readback mismatch: got %d bytes with SHA-256 %s", actual.size, actual.digest)
	}
	return nil
}

func (updater *Updater) restoreBackup(backup persistedFile, executable string) error {
	if err := updater.fileSystem.Rename(backup.path, executable); err != nil {
		return err
	}
	if err := updater.fileSystem.SyncDir(filepath.Dir(executable)); err != nil {
		return fmt.Errorf("synchronizing restored executable directory: %w", err)
	}
	actual, err := updater.readFileIdentity(executable)
	if err != nil {
		return fmt.Errorf("reading restored executable back from disk: %w", err)
	}
	if actual != backup.identity {
		return fmt.Errorf("restored executable readback mismatch")
	}
	return nil
}

func (updater *Updater) readFileIdentity(path string) (fileIdentity, error) {
	file, err := updater.fileSystem.Open(path)
	if err != nil {
		return fileIdentity{}, err
	}
	hash := sha256.New()
	size, copyErr := io.Copy(hash, file)
	closeErr := file.Close()
	if copyErr != nil {
		return fileIdentity{}, copyErr
	}
	if closeErr != nil {
		return fileIdentity{}, closeErr
	}
	return fileIdentity{size: size, digest: hex.EncodeToString(hash.Sum(nil))}, nil
}

func copyAtMost(destination io.Writer, source io.Reader, maxBytes int64) (int64, error) {
	limited := &io.LimitedReader{R: source, N: maxBytes + 1}
	written, err := io.Copy(destination, limited)
	if err != nil {
		return written, err
	}
	if written > maxBytes {
		return written, fmt.Errorf("response exceeds %d bytes", maxBytes)
	}
	return written, nil
}

func (updater *Updater) completeRelaunch(currentVersion string, forced bool) error {
	executable, err := updater.resolvedExecutable()
	if err != nil {
		return updater.warn("update cleanup failed: %v", err)
	}
	if err := updater.fileSystem.Remove(updater.backupPath(executable)); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return updater.warn("update cleanup failed: %v", err)
		}
	} else if err := updater.fileSystem.SyncDir(filepath.Dir(executable)); err != nil {
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
