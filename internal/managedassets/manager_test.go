package managedassets

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	assetfs "angel-ai-opencode/internal/assets"
	"angel-ai-opencode/internal/catalog"
	"angel-ai-opencode/internal/install"
)

func writeAsset(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func fixtureRequest(t *testing.T, assetRoot, target string) install.InstallationRequest {
	t.Helper()
	return install.InstallationRequest{
		Items: []catalog.Item{
			{Name: "worker", Source: "agents/worker.md", Dest: "agents/worker.md", Kind: catalog.CopyFile},
			{Name: "rules", Source: "agents-md/AGENTS.md", Dest: "AGENTS.md", Kind: catalog.CopyFile},
		},
		Extras: map[string]bool{
			"theme": false,
			"cmux":  false,
		},
		Assets:    assetfs.Directory(assetRoot),
		ConfigDir: target,
		AgentModels: install.AgentModelAssignments{
			"worker": {Model: "provider/model", Variant: "high"},
		},
	}
}

func TestStateRoundTripPreservesSelectionAndUsesPrivatePermissions(t *testing.T) {
	assetRoot := t.TempDir()
	target := t.TempDir()
	writeAsset(t, assetRoot, "agents/worker.md", "worker v1\n")
	writeAsset(t, assetRoot, "agents-md/AGENTS.md", "rules v1\n")
	request := fixtureRequest(t, assetRoot, target)

	if _, err := Apply(request); err != nil {
		t.Fatal(err)
	}
	state, err := load(target)
	if err != nil {
		t.Fatal(err)
	}
	if state.SchemaVersion != schemaVersion {
		t.Fatalf("schema version = %d", state.SchemaVersion)
	}
	wantCategories := []categorySelection{
		{Name: "agents", Sources: []string{"agents/worker.md"}},
		{Name: "agents-md", Sources: []string{"agents-md/AGENTS.md"}},
	}
	if !reflect.DeepEqual(state.Selection.Categories, wantCategories) {
		t.Fatalf("categories = %#v, want %#v", state.Selection.Categories, wantCategories)
	}
	if !reflect.DeepEqual(state.Selection.Extras, request.Extras) {
		t.Fatalf("extras = %#v, want %#v", state.Selection.Extras, request.Extras)
	}
	if !reflect.DeepEqual(state.Selection.AgentModels, request.AgentModels) {
		t.Fatalf("agent models = %#v, want %#v", state.Selection.AgentModels, request.AgentModels)
	}
	if len(state.Files) != 3 {
		t.Fatalf("managed files = %#v", state.Files)
	}

	info, err := os.Stat(statePath(target))
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("state permissions = %o", got)
	}
}

func TestDoctorFindsMissingStateBundleDriftAndFileDrift(t *testing.T) {
	assetRoot := t.TempDir()
	target := t.TempDir()
	writeAsset(t, assetRoot, "agents/worker.md", "worker v1\n")
	writeAsset(t, assetRoot, "agents-md/AGENTS.md", "rules v1\n")
	source := assetfs.Directory(assetRoot)

	report, err := Doctor(source, target)
	if err != nil {
		t.Fatal(err)
	}
	if report.Healthy || !hasFinding(report, FindingStateMissing, "") {
		t.Fatalf("missing-state report = %+v", report)
	}

	request := fixtureRequest(t, assetRoot, target)
	if _, err := Apply(request); err != nil {
		t.Fatal(err)
	}
	writeAsset(t, assetRoot, "agents/worker.md", "worker v2\n")
	if err := os.Remove(filepath.Join(target, "AGENTS.md")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "agents", "worker.md"), []byte("user edit\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	report, err = Doctor(source, target)
	if err != nil {
		t.Fatal(err)
	}
	if report.Healthy {
		t.Fatalf("doctor reported healthy: %+v", report)
	}
	if !hasFinding(report, FindingBundleChanged, "") {
		t.Fatalf("bundle drift not found: %+v", report)
	}
	if !hasFinding(report, FindingFileMissing, "AGENTS.md") {
		t.Fatalf("missing file not found: %+v", report)
	}
	if !hasFinding(report, FindingFileModified, "agents/worker.md") {
		t.Fatalf("modified file not found: %+v", report)
	}
}

func TestSyncDryRunReusesSavedSelectionWithoutWriting(t *testing.T) {
	assetRoot := t.TempDir()
	target := t.TempDir()
	writeAsset(t, assetRoot, "agents/worker.md", "worker v1\n")
	writeAsset(t, assetRoot, "agents/ignored.md", "ignored v1\n")
	writeAsset(t, assetRoot, "agents-md/AGENTS.md", "rules v1\n")
	request := fixtureRequest(t, assetRoot, target)
	if _, err := Apply(request); err != nil {
		t.Fatal(err)
	}
	beforeState, err := os.ReadFile(statePath(target))
	if err != nil {
		t.Fatal(err)
	}

	writeAsset(t, assetRoot, "agents/worker.md", "worker v2\n")
	plan, err := Sync(assetfs.Directory(assetRoot), target, true)
	if err != nil {
		t.Fatal(err)
	}
	if !linesContain(plan, "ACTUALIZAR", filepath.Join(target, "agents", "worker.md")) {
		t.Fatalf("sync plan = %v", plan)
	}
	if linesContain(plan, "ignored.md") {
		t.Fatalf("sync selected a newly discovered asset: %v", plan)
	}
	content, err := os.ReadFile(filepath.Join(target, "agents", "worker.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "worker v1\n" {
		t.Fatalf("dry-run wrote target: %q", content)
	}
	afterState, err := os.ReadFile(statePath(target))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(afterState, beforeState) {
		t.Fatal("dry-run rewrote managed state")
	}
}

func TestSyncUpdatesManagedAssetsAndState(t *testing.T) {
	assetRoot := t.TempDir()
	target := t.TempDir()
	writeAsset(t, assetRoot, "agents/worker.md", "worker v1\n")
	writeAsset(t, assetRoot, "agents-md/AGENTS.md", "rules v1\n")
	source := assetfs.Directory(assetRoot)
	if _, err := Apply(fixtureRequest(t, assetRoot, target)); err != nil {
		t.Fatal(err)
	}
	oldState, err := load(target)
	if err != nil {
		t.Fatal(err)
	}

	writeAsset(t, assetRoot, "agents/worker.md", "worker v2\n")
	if _, err := Sync(source, target, false); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(target, "agents", "worker.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "worker v2\n" {
		t.Fatalf("synced content = %q", content)
	}
	newState, err := load(target)
	if err != nil {
		t.Fatal(err)
	}
	if newState.BundleDigest == oldState.BundleDigest {
		t.Fatal("sync did not advance bundle digest")
	}
	report, err := Doctor(source, target)
	if err != nil {
		t.Fatal(err)
	}
	if !report.Healthy {
		t.Fatalf("doctor after sync = %+v", report)
	}
}

func TestSyncRefusesToOverwriteUserEdit(t *testing.T) {
	assetRoot := t.TempDir()
	target := t.TempDir()
	writeAsset(t, assetRoot, "agents/worker.md", "worker v1\n")
	writeAsset(t, assetRoot, "agents-md/AGENTS.md", "rules v1\n")
	source := assetfs.Directory(assetRoot)
	if _, err := Apply(fixtureRequest(t, assetRoot, target)); err != nil {
		t.Fatal(err)
	}

	targetPath := filepath.Join(target, "agents", "worker.md")
	if err := os.WriteFile(targetPath, []byte("my local edit\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeAsset(t, assetRoot, "agents/worker.md", "worker v2\n")
	_, err := Sync(source, target, false)
	var drift *DriftError
	if !errors.As(err, &drift) {
		t.Fatalf("sync error = %v, want DriftError", err)
	}
	content, readErr := os.ReadFile(targetPath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(content) != "my local edit\n" {
		t.Fatalf("sync overwrote user edit: %q", content)
	}
}

func TestSyncDryRunStillPlansWhenUserEditBlocksApply(t *testing.T) {
	assetRoot := t.TempDir()
	target := t.TempDir()
	writeAsset(t, assetRoot, "agents/worker.md", "worker v1\n")
	writeAsset(t, assetRoot, "agents-md/AGENTS.md", "rules v1\n")
	source := assetfs.Directory(assetRoot)
	if _, err := Apply(fixtureRequest(t, assetRoot, target)); err != nil {
		t.Fatal(err)
	}

	targetPath := filepath.Join(target, "agents", "worker.md")
	if err := os.WriteFile(targetPath, []byte("my local edit\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeAsset(t, assetRoot, "agents/worker.md", "worker v2\n")
	plan, err := Sync(source, target, true)
	var drift *DriftError
	if !errors.As(err, &drift) {
		t.Fatalf("sync error = %v, want DriftError", err)
	}
	if !linesContain(plan, "ACTUALIZAR", targetPath) {
		t.Fatalf("dry-run did not return planner output: %v", plan)
	}
	content, readErr := os.ReadFile(targetPath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(content) != "my local edit\n" {
		t.Fatalf("dry-run overwrote user edit: %q", content)
	}
}

func TestSaveRejectsStateLargerThanLoadLimit(t *testing.T) {
	state := state{
		SchemaVersion: schemaVersion,
		BundleDigest:  strings.Repeat("0", 64),
		Selection: selection{Categories: []categorySelection{{
			Name: "agents", Sources: []string{"agents/" + strings.Repeat("x", maxStateBytes)},
		}}},
		Files: []managedFile{},
	}
	target := t.TempDir()
	err := save(target, state)
	if err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("save error = %v", err)
	}
	if _, statErr := os.Stat(statePath(target)); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("state file exists after rejected save: %v", statErr)
	}
}

func TestSyncExpectsNewCopyDirChildToRemainAbsent(t *testing.T) {
	assetRoot := t.TempDir()
	target := t.TempDir()
	writeAsset(t, assetRoot, "skills/demo/original.txt", "managed original\n")
	request := install.InstallationRequest{
		Items: []catalog.Item{{
			Name: "demo", Source: "skills/demo", Dest: "skills/demo", Kind: catalog.CopyDir,
		}},
		Extras: map[string]bool{}, Assets: assetfs.Directory(assetRoot), ConfigDir: target,
	}
	if _, err := Apply(request); err != nil {
		t.Fatal(err)
	}
	beforeState, err := os.ReadFile(statePath(target))
	if err != nil {
		t.Fatal(err)
	}
	newTarget := filepath.Join(target, "skills", "demo", "new.txt")
	if err := os.WriteFile(newTarget, []byte("user file\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeAsset(t, assetRoot, "skills/demo/new.txt", "bundle file\n")

	_, err = Sync(assetfs.Directory(assetRoot), target, false)
	if err == nil || !strings.Contains(err.Error(), "expected to remain absent") {
		t.Fatalf("sync error = %v", err)
	}
	content, readErr := os.ReadFile(newTarget)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(content) != "user file\n" {
		t.Fatalf("sync overwrote new child: %q", content)
	}
	afterState, readErr := os.ReadFile(statePath(target))
	if readErr != nil {
		t.Fatal(readErr)
	}
	if !reflect.DeepEqual(afterState, beforeState) {
		t.Fatal("failed sync advanced managed state")
	}
}

func TestDoctorAndSyncBlockRetiredCopyDirChild(t *testing.T) {
	assetRoot := t.TempDir()
	target := t.TempDir()
	writeAsset(t, assetRoot, "skills/demo/kept.txt", "kept\n")
	writeAsset(t, assetRoot, "skills/demo/retired.txt", "retired\n")
	request := install.InstallationRequest{
		Items: []catalog.Item{{
			Name: "demo", Source: "skills/demo", Dest: "skills/demo", Kind: catalog.CopyDir,
		}},
		Extras: map[string]bool{}, Assets: assetfs.Directory(assetRoot), ConfigDir: target,
	}
	if _, err := Apply(request); err != nil {
		t.Fatal(err)
	}
	beforeState, err := os.ReadFile(statePath(target))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(assetRoot, "skills", "demo", "retired.txt")); err != nil {
		t.Fatal(err)
	}
	source := assetfs.Directory(assetRoot)

	report, err := Doctor(source, target)
	if err != nil {
		t.Fatal(err)
	}
	if !hasFinding(report, FindingRetiredFile, "skills/demo/retired.txt") {
		t.Fatalf("retired file not found: %+v", report)
	}
	_, err = Sync(source, target, false)
	var drift *DriftError
	if !errors.As(err, &drift) || !findingSliceContains(drift.Findings, FindingRetiredFile, "skills/demo/retired.txt") {
		t.Fatalf("sync error = %v", err)
	}
	retiredContent, readErr := os.ReadFile(filepath.Join(target, "skills", "demo", "retired.txt"))
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(retiredContent) != "retired\n" {
		t.Fatalf("retired file changed: %q", retiredContent)
	}
	afterState, readErr := os.ReadFile(statePath(target))
	if readErr != nil {
		t.Fatal(readErr)
	}
	if !reflect.DeepEqual(afterState, beforeState) {
		t.Fatal("blocked retirement advanced managed state")
	}

	if err := os.Remove(filepath.Join(target, "skills", "demo", "retired.txt")); err != nil {
		t.Fatal(err)
	}
	if _, err := Sync(source, target, false); err != nil {
		t.Fatalf("sync after removing retired file: %v", err)
	}
	updated, err := load(target)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range updated.Files {
		if file.Path == "skills/demo/retired.txt" {
			t.Fatalf("retired file remained in state: %+v", updated.Files)
		}
	}
	report, err = Doctor(source, target)
	if err != nil {
		t.Fatal(err)
	}
	if !report.Healthy {
		t.Fatalf("doctor after resolving retired file = %+v", report)
	}
}

func TestApplyExplainsStateSaveFailureAfterInstallingFiles(t *testing.T) {
	assetRoot := t.TempDir()
	target := t.TempDir()
	writeAsset(t, assetRoot, "agents/worker.md", "worker\n")
	if err := os.Mkdir(statePath(target), 0o755); err != nil {
		t.Fatal(err)
	}
	request := install.InstallationRequest{
		Items: []catalog.Item{{
			Name: "worker", Source: "agents/worker.md", Dest: "agents/worker.md", Kind: catalog.CopyFile,
		}},
		Extras: map[string]bool{}, Assets: assetfs.Directory(assetRoot), ConfigDir: target,
	}

	_, err := Apply(request)
	if err == nil || !strings.Contains(err.Error(), "assets were installed but managed state could not be saved") {
		t.Fatalf("apply error = %v", err)
	}
	content, readErr := os.ReadFile(filepath.Join(target, "agents", "worker.md"))
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(content) != "worker\n" {
		t.Fatalf("installed content = %q", content)
	}
}

func hasFinding(report DoctorReport, code FindingCode, path string) bool {
	return findingSliceContains(report.Findings, code, path)
}

func findingSliceContains(findings []Finding, code FindingCode, path string) bool {
	for _, finding := range findings {
		if finding.Code == code && (path == "" || finding.Path == path) {
			return true
		}
	}
	return false
}

func linesContain(lines []string, values ...string) bool {
	for _, line := range lines {
		matches := true
		for _, value := range values {
			if !strings.Contains(line, value) {
				matches = false
				break
			}
		}
		if matches {
			return true
		}
	}
	return false
}

func TestDoctorAndSyncSurviveWholeItemRetirement(t *testing.T) {
	assetRoot := t.TempDir()
	target := t.TempDir()
	writeAsset(t, assetRoot, "agents/kept.md", "kept\n")
	writeAsset(t, assetRoot, "agents/retired.md", "retired\n")
	request := install.InstallationRequest{
		Items: []catalog.Item{
			{Name: "kept", Source: "agents/kept.md", Dest: "agents/kept.md", Kind: catalog.CopyFile},
			{Name: "retired", Source: "agents/retired.md", Dest: "agents/retired.md", Kind: catalog.CopyFile},
		},
		Extras: map[string]bool{}, Assets: assetfs.Directory(assetRoot), ConfigDir: target,
	}
	if _, err := Apply(request); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(assetRoot, "agents", "retired.md")); err != nil {
		t.Fatal(err)
	}
	source := assetfs.Directory(assetRoot)

	report, err := Doctor(source, target)
	if err != nil {
		t.Fatalf("doctor after whole-item retirement: %v", err)
	}
	if !hasFinding(report, FindingRetiredFile, "agents/retired.md") {
		t.Fatalf("retired item file not identified: %+v", report)
	}
	_, err = Sync(source, target, false)
	var drift *DriftError
	if !errors.As(err, &drift) || !findingSliceContains(drift.Findings, FindingRetiredFile, "agents/retired.md") {
		t.Fatalf("sync error = %v", err)
	}
	retiredContent, readErr := os.ReadFile(filepath.Join(target, "agents", "retired.md"))
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(retiredContent) != "retired\n" {
		t.Fatalf("retired item file changed: %q", retiredContent)
	}

	if err := os.Remove(filepath.Join(target, "agents", "retired.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := Sync(source, target, false); err != nil {
		t.Fatalf("sync after removing retired item file: %v", err)
	}
	updated, err := load(target)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range updated.Files {
		if file.Path == "agents/retired.md" {
			t.Fatalf("retired item file remained in state: %+v", updated.Files)
		}
	}
	for _, category := range updated.Selection.Categories {
		for _, source := range category.Sources {
			if source == "agents/retired.md" {
				t.Fatalf("retired item remained selected: %+v", updated.Selection)
			}
		}
	}
	report, err = Doctor(source, target)
	if err != nil {
		t.Fatal(err)
	}
	if !report.Healthy {
		t.Fatalf("doctor after resolving retired item = %+v", report)
	}
}
