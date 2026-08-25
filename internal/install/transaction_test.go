package install

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	assetfs "angel-ai-opencode/internal/assets"
	"angel-ai-opencode/internal/catalog"
)

func useInstallationLockRoot(t *testing.T, root string) {
	t.Helper()
	previous := resolveInstallationLockRoot
	resolveInstallationLockRoot = func() (string, error) { return root, nil }
	t.Cleanup(func() { resolveInstallationLockRoot = previous })
}

func TestInstallationLockRejectsSymlinkRoot(t *testing.T) {
	parent := t.TempDir()
	privateRoot := filepath.Join(parent, "private")
	if err := os.Mkdir(privateRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	lockRoot := filepath.Join(parent, "locks")
	if err := os.Symlink(privateRoot, lockRoot); err != nil {
		t.Skipf("symlinks are unavailable: %v", err)
	}
	useInstallationLockRoot(t, lockRoot)

	lease, err := acquireInstallationLock(filepath.Join(t.TempDir(), "opencode"))
	if lease != nil || err == nil || !strings.Contains(err.Error(), "symbolic link") {
		t.Fatalf("lock with symlink root = %v, %v; want symbolic-link refusal", lease, err)
	}
}

func TestInstallationLockRejectsInsecureRootPermissions(t *testing.T) {
	lockRoot := filepath.Join(t.TempDir(), "locks")
	if err := os.Mkdir(lockRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(lockRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	useInstallationLockRoot(t, lockRoot)

	lease, err := acquireInstallationLock(filepath.Join(t.TempDir(), "opencode"))
	if lease != nil || err == nil || !strings.Contains(err.Error(), "permissions") {
		t.Fatalf("lock with public root = %v, %v; want permissions refusal", lease, err)
	}
}

func TestInstallationLockRejectsSymlinkFile(t *testing.T) {
	lockRoot := filepath.Join(t.TempDir(), "locks")
	if err := os.Mkdir(lockRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	useInstallationLockRoot(t, lockRoot)

	target := filepath.Join(t.TempDir(), "opencode")
	canonicalTarget, err := resolveExistingPath(target)
	if err != nil {
		t.Fatal(err)
	}
	victim := filepath.Join(t.TempDir(), "victim")
	writeTestFile(t, victim, "keep\n")
	lockPath := filepath.Join(lockRoot, installationLockFileName(canonicalTarget))
	if err := os.Symlink(victim, lockPath); err != nil {
		t.Skipf("symlinks are unavailable: %v", err)
	}

	lease, err := acquireInstallationLock(target)
	if lease != nil || err == nil || !strings.Contains(err.Error(), "symbolic link") {
		t.Fatalf("lock with symlink file = %v, %v; want symbolic-link refusal", lease, err)
	}
	if got := string(readTestFile(t, victim)); got != "keep\n" {
		t.Fatalf("symlink target changed: %q", got)
	}
}

func TestInstallationLockRejectsInsecureFilePermissions(t *testing.T) {
	lockRoot := filepath.Join(t.TempDir(), "locks")
	if err := os.Mkdir(lockRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	useInstallationLockRoot(t, lockRoot)

	target := filepath.Join(t.TempDir(), "opencode")
	canonicalTarget, err := resolveExistingPath(target)
	if err != nil {
		t.Fatal(err)
	}
	lockPath := filepath.Join(lockRoot, installationLockFileName(canonicalTarget))
	if err := os.WriteFile(lockPath, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(lockPath, 0o644); err != nil {
		t.Fatal(err)
	}

	lease, err := acquireInstallationLock(target)
	if lease != nil || err == nil || !strings.Contains(err.Error(), "permissions") {
		t.Fatalf("lock with public file = %v, %v; want permissions refusal", lease, err)
	}
}

func TestInstallationLockCoordinatesIndependentHandles(t *testing.T) {
	useInstallationLockRoot(t, filepath.Join(t.TempDir(), "locks"))
	target := filepath.Join(t.TempDir(), "opencode")

	first, err := acquireInstallationLock(target)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = first.release() }()

	second, err := acquireInstallationLock(target)
	if second != nil || !errors.Is(err, errInstallationLockBusy) {
		t.Fatalf("second lock = %v, %v; want busy", second, err)
	}

	if err := first.release(); err != nil {
		t.Fatal(err)
	}
	reacquired, err := acquireInstallationLock(target)
	if err != nil {
		t.Fatalf("lock after release: %v", err)
	}
	if err := reacquired.release(); err != nil {
		t.Fatal(err)
	}
}

func TestInstallationLockCoordinatesSymlinkAliases(t *testing.T) {
	parent := t.TempDir()
	target := filepath.Join(parent, "opencode")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(parent, "opencode-alias")
	if err := os.Symlink(target, alias); err != nil {
		t.Skipf("symlinks are unavailable: %v", err)
	}

	first, err := acquireInstallationLock(target)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = first.release() }()

	second, err := acquireInstallationLock(alias)
	if second != nil || !errors.Is(err, errInstallationLockBusy) {
		t.Fatalf("lock through symlink alias = %v, %v; want busy", second, err)
	}
}

func TestApplyInstallationHoldsLockWhilePreparing(t *testing.T) {
	previous := prepareInstallationForApply
	preparations := 0
	prepareInstallationForApply = func(request InstallationRequest) (preparedInstallation, error) {
		preparations++
		lease, err := acquireInstallationLock(request.ConfigDir)
		if lease != nil || !errors.Is(err, errInstallationLockBusy) {
			t.Fatalf("lock during preparation = %v, %v; want busy", lease, err)
		}
		return prepareInstallation(request)
	}
	t.Cleanup(func() { prepareInstallationForApply = previous })

	target := t.TempDir()
	if _, err := ApplyInstallation(InstallationRequest{ConfigDir: target}); err != nil {
		t.Fatal(err)
	}
	if preparations != 1 {
		t.Fatalf("preparations = %d, want 1", preparations)
	}
}

func TestApplyInstallationRollsBackUpdatedAndCreatedFiles(t *testing.T) {
	assetsDir := t.TempDir()
	writeTestFile(t, filepath.Join(assetsDir, "a-existing.md"), "new existing\n")
	writeTestFile(t, filepath.Join(assetsDir, "b-created.md"), "new created\n")
	writeTestFile(t, filepath.Join(assetsDir, "c-fails.md"), "never published\n")

	target := t.TempDir()
	existingPath := filepath.Join(target, "a-existing.md")
	createdPath := filepath.Join(target, "b-created.md")
	writeTestFile(t, existingPath, "old existing\n")

	useTransactionFileApply(t, func(entry *transactionEntry) (fileWriteResult, error) {
		if filepath.Base(entry.file.path) == "c-fails.md" {
			return fileWriteResult{}, errors.New("injected third-file failure")
		}
		return applyTransactionEntry(entry)
	})

	report, err := ApplyInstallation(InstallationRequest{
		Items: []catalog.Item{
			{Name: "existing", Source: "a-existing.md", Dest: "a-existing.md", Kind: catalog.CopyFile},
			{Name: "created", Source: "b-created.md", Dest: "b-created.md", Kind: catalog.CopyFile},
			{Name: "failure", Source: "c-fails.md", Dest: "c-fails.md", Kind: catalog.CopyFile},
		},
		Assets:    assetfs.Directory(assetsDir),
		ConfigDir: target,
	})
	if err == nil || !strings.Contains(err.Error(), "managed file changes were rolled back") {
		t.Fatalf("ApplyInstallation error = %v, want successful rollback", err)
	}
	if len(report) != 0 {
		t.Fatalf("rolled-back files were reported as complete: %v", report)
	}
	if got := string(readTestFile(t, existingPath)); got != "old existing\n" {
		t.Fatalf("existing file after rollback = %q", got)
	}
	if _, err := os.Stat(createdPath); !os.IsNotExist(err) {
		t.Fatalf("created file survived rollback: %v", err)
	}
	backups, err := filepath.Glob(filepath.Join(target, "*.bak-*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(backups) != 0 {
		t.Fatalf("aborted transaction left backups: %v", backups)
	}
}

func TestApplyInstallationRollbackPreservesDetectedExternalEdit(t *testing.T) {
	assetsDir := t.TempDir()
	writeTestFile(t, filepath.Join(assetsDir, "a-first.md"), "managed first\n")
	writeTestFile(t, filepath.Join(assetsDir, "b-fails.md"), "managed second\n")

	target := t.TempDir()
	firstPath := filepath.Join(target, "a-first.md")
	writeTestFile(t, firstPath, "original first\n")

	useTransactionFileApply(t, func(entry *transactionEntry) (fileWriteResult, error) {
		if filepath.Base(entry.file.path) == "b-fails.md" {
			if err := os.WriteFile(firstPath, []byte("external edit\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			return fileWriteResult{}, errors.New("injected second-file failure")
		}
		return applyTransactionEntry(entry)
	})

	_, err := ApplyInstallation(InstallationRequest{
		Items: []catalog.Item{
			{Name: "first", Source: "a-first.md", Dest: "a-first.md", Kind: catalog.CopyFile},
			{Name: "failure", Source: "b-fails.md", Dest: "b-fails.md", Kind: catalog.CopyFile},
		},
		Assets:    assetfs.Directory(assetsDir),
		ConfigDir: target,
	})
	if err == nil || !strings.Contains(err.Error(), "rollback conflict") {
		t.Fatalf("ApplyInstallation error = %v, want rollback conflict", err)
	}
	if got := string(readTestFile(t, firstPath)); got != "external edit\n" {
		t.Fatalf("rollback overwrote external edit: %q", got)
	}
	backups, globErr := filepath.Glob(firstPath + ".bak-*")
	if globErr != nil {
		t.Fatal(globErr)
	}
	if len(backups) != 1 || string(readTestFile(t, backups[0])) != "original first\n" {
		t.Fatalf("conflicted rollback backup = %v", backups)
	}
	if !strings.Contains(err.Error(), backups[0]) {
		t.Fatalf("rollback error does not identify preserved backup %q: %v", backups[0], err)
	}
}

func TestApplyInstallationRollbackDetectsReplacementWithSameBytes(t *testing.T) {
	assetsDir := t.TempDir()
	writeTestFile(t, filepath.Join(assetsDir, "a-first.md"), "managed first\n")
	writeTestFile(t, filepath.Join(assetsDir, "b-fails.md"), "managed second\n")

	target := t.TempDir()
	firstPath := filepath.Join(target, "a-first.md")
	writeTestFile(t, firstPath, "original first\n")
	var replacement os.FileInfo

	useTransactionFileApply(t, func(entry *transactionEntry) (fileWriteResult, error) {
		if filepath.Base(entry.file.path) == "b-fails.md" {
			tempPath := filepath.Join(target, "replacement.tmp")
			writeTestFile(t, tempPath, "managed first\n")
			info, err := os.Stat(tempPath)
			if err != nil {
				t.Fatal(err)
			}
			replacement = info
			if err := os.Rename(tempPath, firstPath); err != nil {
				t.Fatal(err)
			}
			return fileWriteResult{}, errors.New("injected second-file failure")
		}
		return applyTransactionEntry(entry)
	})

	_, err := ApplyInstallation(InstallationRequest{
		Items: []catalog.Item{
			{Name: "first", Source: "a-first.md", Dest: "a-first.md", Kind: catalog.CopyFile},
			{Name: "failure", Source: "b-fails.md", Dest: "b-fails.md", Kind: catalog.CopyFile},
		},
		Assets:    assetfs.Directory(assetsDir),
		ConfigDir: target,
	})
	if err == nil || !strings.Contains(err.Error(), "rollback conflict") {
		t.Fatalf("ApplyInstallation error = %v, want rollback conflict", err)
	}
	current, statErr := os.Stat(firstPath)
	if statErr != nil {
		t.Fatal(statErr)
	}
	if !os.SameFile(current, replacement) {
		t.Fatal("rollback replaced an external inode with identical bytes")
	}
	if got := string(readTestFile(t, firstPath)); got != "managed first\n" {
		t.Fatalf("external replacement content = %q", got)
	}
}

func TestApplyInstallationRollbackDetectsSameInodeRewriteWithSameBytes(t *testing.T) {
	assetsDir := t.TempDir()
	writeTestFile(t, filepath.Join(assetsDir, "a-first.md"), "managed first\n")
	writeTestFile(t, filepath.Join(assetsDir, "b-fails.md"), "managed second\n")

	target := t.TempDir()
	firstPath := filepath.Join(target, "a-first.md")
	writeTestFile(t, firstPath, "original first\n")
	changedAt := time.Unix(1_600_000_000, 0)

	useTransactionFileApply(t, func(entry *transactionEntry) (fileWriteResult, error) {
		if filepath.Base(entry.file.path) == "b-fails.md" {
			if err := os.WriteFile(firstPath, []byte("managed first\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.Chtimes(firstPath, changedAt, changedAt); err != nil {
				t.Fatal(err)
			}
			return fileWriteResult{}, errors.New("injected second-file failure")
		}
		return applyTransactionEntry(entry)
	})

	_, err := ApplyInstallation(InstallationRequest{
		Items: []catalog.Item{
			{Name: "first", Source: "a-first.md", Dest: "a-first.md", Kind: catalog.CopyFile},
			{Name: "failure", Source: "b-fails.md", Dest: "b-fails.md", Kind: catalog.CopyFile},
		},
		Assets:    assetfs.Directory(assetsDir),
		ConfigDir: target,
	})
	if err == nil || !strings.Contains(err.Error(), "rollback conflict") {
		t.Fatalf("ApplyInstallation error = %v, want rollback conflict", err)
	}
	info, statErr := os.Stat(firstPath)
	if statErr != nil {
		t.Fatal(statErr)
	}
	if !info.ModTime().Equal(changedAt) {
		t.Fatalf("rollback replaced same-inode external rewrite: modtime = %s", info.ModTime())
	}
	if got := string(readTestFile(t, firstPath)); got != "managed first\n" {
		t.Fatalf("external rewrite content = %q", got)
	}
}

func TestApplyInstallationRejectsManagedPathOutsideConfigDir(t *testing.T) {
	assetsDir := t.TempDir()
	writeTestFile(t, filepath.Join(assetsDir, "managed.md"), "managed\n")

	parent := t.TempDir()
	target := filepath.Join(parent, "opencode")
	outside := filepath.Join(parent, "outside.md")
	writeTestFile(t, outside, "keep\n")

	_, err := ApplyInstallation(InstallationRequest{
		Items: []catalog.Item{{
			Name: "escape", Source: "managed.md", Dest: "../outside.md", Kind: catalog.CopyFile,
		}},
		Assets:    assetfs.Directory(assetsDir),
		ConfigDir: target,
	})
	if err == nil || !strings.Contains(err.Error(), "outside installation config directory") {
		t.Fatalf("ApplyInstallation error = %v, want allow-list refusal", err)
	}
	if got := string(readTestFile(t, outside)); got != "keep\n" {
		t.Fatalf("outside file changed: %q", got)
	}
}

func TestApplyInstallationKeepsCanonicalTargetAfterConfigAliasChanges(t *testing.T) {
	assetsDir := t.TempDir()
	writeTestFile(t, filepath.Join(assetsDir, "managed.md"), "managed\n")

	parent := t.TempDir()
	realTarget := filepath.Join(parent, "opencode-real")
	outsideTarget := filepath.Join(parent, "outside")
	if err := os.Mkdir(realTarget, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(outsideTarget, 0o755); err != nil {
		t.Fatal(err)
	}
	realFile := filepath.Join(realTarget, "managed.md")
	outsideFile := filepath.Join(outsideTarget, "managed.md")
	writeTestFile(t, realFile, "original\n")
	writeTestFile(t, outsideFile, "outside\n")

	alias := filepath.Join(parent, "opencode")
	if err := os.Symlink(realTarget, alias); err != nil {
		t.Skipf("symlinks are unavailable: %v", err)
	}
	useTransactionFileApply(t, func(entry *transactionEntry) (fileWriteResult, error) {
		if err := os.Remove(alias); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(outsideTarget, alias); err != nil {
			t.Fatal(err)
		}
		return applyTransactionEntry(entry)
	})

	_, err := ApplyInstallation(InstallationRequest{
		Items: []catalog.Item{{
			Name: "managed", Source: "managed.md", Dest: "managed.md", Kind: catalog.CopyFile,
		}},
		Assets:    assetfs.Directory(assetsDir),
		ConfigDir: alias,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := string(readTestFile(t, realFile)); got != "managed\n" {
		t.Fatalf("canonical managed file = %q", got)
	}
	if got := string(readTestFile(t, outsideFile)); got != "outside\n" {
		t.Fatalf("retargeted alias changed outside file: %q", got)
	}
}

func TestApplyInstallationMapsPreparedPathsToLockedCanonicalConfigRoot(t *testing.T) {
	assetsDir := t.TempDir()
	writeTestFile(t, filepath.Join(assetsDir, "managed.md"), "managed\n")

	parent := t.TempDir()
	realTarget := filepath.Join(parent, "opencode-real")
	nestedTarget := filepath.Join(realTarget, "nested")
	if err := os.Mkdir(realTarget, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(nestedTarget, 0o755); err != nil {
		t.Fatal(err)
	}
	realFile := filepath.Join(realTarget, "managed.md")
	nestedFile := filepath.Join(nestedTarget, "managed.md")
	writeTestFile(t, realFile, "original\n")
	writeTestFile(t, nestedFile, "nested\n")

	alias := filepath.Join(parent, "opencode")
	if err := os.Symlink(realTarget, alias); err != nil {
		t.Skipf("symlinks are unavailable: %v", err)
	}
	previous := prepareInstallationForApply
	prepareInstallationForApply = func(request InstallationRequest) (preparedInstallation, error) {
		prepared, err := prepareInstallation(request)
		if err != nil {
			return preparedInstallation{}, err
		}
		if err := os.Remove(alias); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(nestedTarget, alias); err != nil {
			t.Fatal(err)
		}
		return prepared, nil
	}
	t.Cleanup(func() { prepareInstallationForApply = previous })

	_, err := ApplyInstallation(InstallationRequest{
		Items: []catalog.Item{{
			Name: "managed", Source: "managed.md", Dest: "managed.md", Kind: catalog.CopyFile,
		}},
		Assets:    assetfs.Directory(assetsDir),
		ConfigDir: alias,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := string(readTestFile(t, realFile)); got != "managed\n" {
		t.Fatalf("locked canonical managed file = %q", got)
	}
	if got := string(readTestFile(t, nestedFile)); got != "nested\n" {
		t.Fatalf("retargeted nested file changed: %q", got)
	}
}

func TestApplyInstallationRollbackPreservesReplacedBackupWithSameBytes(t *testing.T) {
	assetsDir := t.TempDir()
	writeTestFile(t, filepath.Join(assetsDir, "a-first.md"), "managed first\n")
	writeTestFile(t, filepath.Join(assetsDir, "b-fails.md"), "managed second\n")

	target := t.TempDir()
	firstPath := filepath.Join(target, "a-first.md")
	writeTestFile(t, firstPath, "original first\n")
	var replacement os.FileInfo
	var backupPath string

	useTransactionFileApply(t, func(entry *transactionEntry) (fileWriteResult, error) {
		if filepath.Base(entry.file.path) == "b-fails.md" {
			backups, err := filepath.Glob(firstPath + ".bak-*")
			if err != nil || len(backups) != 1 {
				t.Fatalf("backup before injected failure = %v, %v", backups, err)
			}
			backupPath = backups[0]
			tempPath := filepath.Join(target, "replacement-backup.tmp")
			if err := os.WriteFile(tempPath, []byte("original first\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			info, err := os.Stat(tempPath)
			if err != nil {
				t.Fatal(err)
			}
			replacement = info
			if err := os.Rename(tempPath, backupPath); err != nil {
				t.Fatal(err)
			}
			return fileWriteResult{}, errors.New("injected second-file failure")
		}
		return applyTransactionEntry(entry)
	})

	_, err := ApplyInstallation(InstallationRequest{
		Items: []catalog.Item{
			{Name: "first", Source: "a-first.md", Dest: "a-first.md", Kind: catalog.CopyFile},
			{Name: "failure", Source: "b-fails.md", Dest: "b-fails.md", Kind: catalog.CopyFile},
		},
		Assets:    assetfs.Directory(assetsDir),
		ConfigDir: target,
	})
	if err == nil || !strings.Contains(err.Error(), "rollback conflict") {
		t.Fatalf("ApplyInstallation error = %v, want backup rollback conflict", err)
	}
	current, statErr := os.Stat(backupPath)
	if statErr != nil {
		t.Fatal(statErr)
	}
	if !os.SameFile(current, replacement) {
		t.Fatal("rollback removed a replacement backup with identical bytes")
	}
}

func TestApplyInstallationRollsBackWhenPublishedContentCannotBeConfirmed(t *testing.T) {
	assetsDir := t.TempDir()
	writeTestFile(t, filepath.Join(assetsDir, "managed.md"), "managed\n")

	target := t.TempDir()
	managedPath := filepath.Join(target, "managed.md")
	writeTestFile(t, managedPath, "original\n")

	previous := readPublishedFile
	reads := 0
	readPublishedFile = func(path string) ([]byte, error) {
		reads++
		if reads == 1 {
			return []byte("unexpected bytes\n"), nil
		}
		return os.ReadFile(path)
	}
	t.Cleanup(func() { readPublishedFile = previous })

	_, err := ApplyInstallation(InstallationRequest{
		Items: []catalog.Item{{
			Name: "managed", Source: "managed.md", Dest: "managed.md", Kind: catalog.CopyFile,
		}},
		Assets:    assetfs.Directory(assetsDir),
		ConfigDir: target,
	})
	if err == nil || !strings.Contains(err.Error(), "confirming published content") {
		t.Fatalf("ApplyInstallation error = %v, want readback failure", err)
	}
	if got := string(readTestFile(t, managedPath)); got != "original\n" {
		t.Fatalf("unconfirmed publish was not rolled back: %q", got)
	}
}

func TestApplyInstallationReportsPackageChangesAsExternalToFileRollback(t *testing.T) {
	environment := newInjectedCLIEnvironment(t, "npm")
	environment.latestVersions[codegraphRegistryPackage] = "2.0.0"
	useGlobalCLICommands(t, environment.commands())

	target := t.TempDir()
	useTransactionFileApply(t, func(*transactionEntry) (fileWriteResult, error) {
		return fileWriteResult{}, errors.New("injected file failure")
	})

	report, err := ApplyInstallation(installationRequestForDescriptors(
		t, target, true, globalCLIDescriptors[0],
	))
	if err == nil || !strings.Contains(err.Error(), "external package-manager effects were not reverted") {
		t.Fatalf("ApplyInstallation error = %v, want external-effect notice", err)
	}
	if len(report) != 1 || !strings.Contains(report[0], "efecto externo") {
		t.Fatalf("external package change report = %v", report)
	}
	if len(environment.installations) != 1 {
		t.Fatalf("package installation count = %d, want 1", len(environment.installations))
	}
}

func TestApplyInstallationTreatsFailedPackageCommandAsExternalAttempt(t *testing.T) {
	environment := newInjectedCLIEnvironment(t, "npm")
	environment.latestVersions[codegraphRegistryPackage] = "2.0.0"
	environment.installationErrs[codegraphPackage] = errors.New("exit 1")
	useGlobalCLICommands(t, environment.commands())

	_, err := ApplyInstallation(installationRequestForDescriptors(
		t, t.TempDir(), false, globalCLIDescriptors[0],
	))
	if err == nil || !strings.Contains(err.Error(), "external package-manager effects were not reverted") {
		t.Fatalf("ApplyInstallation error = %v, want external-attempt notice", err)
	}
}

func useTransactionFileApply(
	t *testing.T,
	apply func(*transactionEntry) (fileWriteResult, error),
) {
	t.Helper()
	previous := applyTransactionFile
	applyTransactionFile = apply
	t.Cleanup(func() { applyTransactionFile = previous })
}
