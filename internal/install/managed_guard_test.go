package install

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	assetfs "angel-ai-opencode/internal/assets"
	"angel-ai-opencode/internal/catalog"
)

func TestApplyInstallationChecksManagedDigestAtWriteTime(t *testing.T) {
	assetRoot := t.TempDir()
	target := t.TempDir()
	sourcePath := filepath.Join(assetRoot, "agents", "worker.md")
	targetPath := filepath.Join(target, "agents", "worker.md")
	if err := os.MkdirAll(filepath.Dir(sourcePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sourcePath, []byte("bundle v2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	managed := []byte("managed v1\n")
	if err := os.WriteFile(targetPath, managed, 0o644); err != nil {
		t.Fatal(err)
	}
	digest := fmt.Sprintf("%x", sha256.Sum256(managed))

	previousPrepare := prepareInstallationForApply
	prepareInstallationForApply = func(request InstallationRequest) (preparedInstallation, error) {
		prepared, err := prepareInstallation(request)
		if err == nil {
			err = os.WriteFile(targetPath, []byte("edit during planning\n"), 0o644)
		}
		return prepared, err
	}
	t.Cleanup(func() { prepareInstallationForApply = previousPrepare })

	_, err := ApplyInstallation(InstallationRequest{
		Items: []catalog.Item{{
			Name: "worker", Source: "agents/worker.md", Dest: "agents/worker.md", Kind: catalog.CopyFile,
		}},
		Extras: map[string]bool{}, Assets: assetfs.Directory(assetRoot), ConfigDir: target,
		FileExpectations: map[string]FileExpectation{targetPath: {SHA256: digest}},
	})
	if err == nil || !strings.Contains(err.Error(), "changed after planning") {
		t.Fatalf("apply error = %v", err)
	}
	content, readErr := os.ReadFile(targetPath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(content) != "edit during planning\n" {
		t.Fatalf("guarded apply wrote %q", content)
	}
}

func TestApplyInstallationRevalidatesManagedDigestBeforeRename(t *testing.T) {
	assetRoot := t.TempDir()
	target := t.TempDir()
	sourcePath := filepath.Join(assetRoot, "agents", "worker.md")
	targetPath := filepath.Join(target, "agents", "worker.md")
	if err := os.MkdirAll(filepath.Dir(sourcePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sourcePath, []byte("bundle v2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	managed := []byte("managed v1\n")
	if err := os.WriteFile(targetPath, managed, 0o644); err != nil {
		t.Fatal(err)
	}
	digest := fmt.Sprintf("%x", sha256.Sum256(managed))

	previousHook := beforeFilePublish
	beforeFilePublish = func(path string) error {
		if path != targetPath {
			t.Fatalf("publish hook path = %q", path)
		}
		return os.WriteFile(path, []byte("edit before rename\n"), 0o644)
	}
	t.Cleanup(func() { beforeFilePublish = previousHook })

	_, err := ApplyInstallation(InstallationRequest{
		Items: []catalog.Item{{
			Name: "worker", Source: "agents/worker.md", Dest: "agents/worker.md", Kind: catalog.CopyFile,
		}},
		Extras: map[string]bool{}, Assets: assetfs.Directory(assetRoot), ConfigDir: target,
		FileExpectations: map[string]FileExpectation{targetPath: {SHA256: digest}},
	})
	if err == nil || !strings.Contains(err.Error(), "changed before publication") {
		t.Fatalf("apply error = %v", err)
	}
	content, readErr := os.ReadFile(targetPath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(content) != "edit before rename\n" {
		t.Fatalf("publish guard wrote %q", content)
	}
}

func TestApplyInstallationRevalidatesExpectedAbsenceBeforeRename(t *testing.T) {
	assetRoot := t.TempDir()
	target := t.TempDir()
	sourcePath := filepath.Join(assetRoot, "agents", "new-worker.md")
	targetPath := filepath.Join(target, "agents", "new-worker.md")
	if err := os.MkdirAll(filepath.Dir(sourcePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sourcePath, []byte("bundle file\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	previousHook := beforeFilePublish
	beforeFilePublish = func(path string) error {
		return os.WriteFile(path, []byte("created before rename\n"), 0o644)
	}
	t.Cleanup(func() { beforeFilePublish = previousHook })

	_, err := ApplyInstallation(InstallationRequest{
		Items: []catalog.Item{{
			Name: "new-worker", Source: "agents/new-worker.md", Dest: "agents/new-worker.md", Kind: catalog.CopyFile,
		}},
		Extras: map[string]bool{}, Assets: assetfs.Directory(assetRoot), ConfigDir: target,
		FileExpectations: map[string]FileExpectation{targetPath: {Absent: true}},
	})
	if err == nil || !strings.Contains(err.Error(), "expected to remain absent before publication") {
		t.Fatalf("apply error = %v", err)
	}
	content, readErr := os.ReadFile(targetPath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(content) != "created before rename\n" {
		t.Fatalf("publish guard wrote %q", content)
	}
}
