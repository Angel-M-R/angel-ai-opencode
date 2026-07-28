package verifiertasks

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const validTasks = `## 1. Implementation

- [x] 1.1 Implement feature.

## 2. Final Verification

<!-- owner: openspec-verifier -->
- [ ] 2.1 Run focused tests.
- [ ] 2.2 Run the build.
`

type recordedRunner struct {
	t           *testing.T
	status      func() []byte
	wantStore   string
	calls       [][]string
	workingDirs []string
}

func (runner *recordedRunner) Run(_ context.Context, directory, name string, args ...string) ([]byte, int, error) {
	runner.t.Helper()
	command := append([]string{name}, args...)
	runner.calls = append(runner.calls, command)
	runner.workingDirs = append(runner.workingDirs, directory)
	want := []string{"openspec", "status", "--change", "demo", "--json"}
	if runner.wantStore != "" {
		want = append(want, "--store", runner.wantStore)
	}
	if !reflect.DeepEqual(command, want) {
		runner.t.Fatalf("command = %v, want %v", command, want)
	}
	return runner.status(), 0, nil
}

type fixture struct {
	root       string
	changeRoot string
	tasksPath  string
	runner     *recordedRunner
	service    *Service
	resolve    ResolveRequest
}

func newFixture(t *testing.T, store string) fixture {
	t.Helper()
	root := t.TempDir()
	changeRoot := filepath.Join(root, "openspec", "changes", "demo")
	if err := os.MkdirAll(changeRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	tasksPath := filepath.Join(changeRoot, "tasks.md")
	if err := os.WriteFile(tasksPath, []byte(validTasks), 0o640); err != nil {
		t.Fatal(err)
	}
	runner := &recordedRunner{t: t, wantStore: store}
	runner.status = func() []byte { return statusJSON(t, root, changeRoot, tasksPath, "done", nil) }
	return fixture{
		root:       root,
		changeRoot: changeRoot,
		tasksPath:  tasksPath,
		runner:     runner,
		service:    &Service{runner: runner},
		resolve:    ResolveRequest{Change: "demo", Store: store, WorkingDirectory: root},
	}
}

func statusJSON(t *testing.T, root, changeRoot, tasksPath, artifactStatus string, existing []string) []byte {
	t.Helper()
	if existing == nil {
		existing = []string{tasksPath}
	}
	status := map[string]any{
		"changeName":   "demo",
		"planningHome": map[string]any{"root": root},
		"changeRoot":   changeRoot,
		"artifactPaths": map[string]any{"tasks": map[string]any{
			"resolvedOutputPath": tasksPath, "existingOutputPaths": existing,
		}},
		"artifacts": []map[string]any{{"id": "tasks", "status": artifactStatus}},
	}
	output, err := json.Marshal(status)
	if err != nil {
		t.Fatal(err)
	}
	return output
}

func captureSnapshot(t *testing.T, fixture fixture) Snapshot {
	t.Helper()
	result, err := fixture.service.Capture(context.Background(), fixture.resolve)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "done" || result.Snapshot == nil {
		t.Fatalf("capture result = %+v", result)
	}
	return *result.Snapshot
}

func passingRequest(snapshot Snapshot) CompleteRequest {
	evidence := make([]TaskEvidence, 0, len(snapshot.PendingTasks))
	for _, task := range snapshot.PendingTasks {
		evidence = append(evidence, TaskEvidence{
			Task: task, Status: "pass",
			Commands: []CommandRecord{{Command: []string{"go", "test", "./focused"}, ExitCode: 0}},
		})
	}
	return CompleteRequest{Snapshot: snapshot, Verdict: "pass", Tasks: snapshot.PendingTasks, Evidence: evidence}
}

func TestParseStrictVerifierMarker(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantPresent bool
		wantError   string
	}{
		{name: "valid", content: validTasks, wantPresent: true},
		{name: "prose does not infer ownership", content: "## Verification\n\n- [ ] 1.1 Ask the openspec-verifier to verify.\n"},
		{name: "duplicate", content: validTasks + "\n<!-- owner: openspec-verifier -->\n", wantError: "duplicate"},
		{name: "malformed", content: strings.Replace(validTasks, OwnerMarker, "<!-- owner:openspec-verifier -->", 1), wantError: "malformed"},
		{name: "misplaced", content: OwnerMarker + "\n\n## Verification\n- [ ] 1.1 Test.\n", wantError: "misplaced"},
		{name: "not first nonblank", content: strings.Replace(validTasks, OwnerMarker, "intro\n"+OwnerMarker, 1), wantError: "first nonblank"},
		{name: "nested", content: strings.Replace(validTasks, OwnerMarker, "### Commands\n"+OwnerMarker, 1), wantError: "nested"},
		{name: "non terminal", content: validTasks + "\n## 3. Notes\n\nNothing.\n", wantError: "final task section"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parsed, err := parseDocument([]byte(test.content))
			if test.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantError) {
					t.Fatalf("error = %v, want containing %q", err, test.wantError)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if parsed.Marker.Present != test.wantPresent {
				t.Fatalf("marker present = %v, want %v", parsed.Marker.Present, test.wantPresent)
			}
			if test.wantPresent && len(parsed.Pending) != 2 {
				t.Fatalf("pending tasks = %v", parsed.Pending)
			}
		})
	}
}

func TestCaptureRepresentsOrdinaryAndCompletedMarkedTaskStates(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		markerPresent bool
	}{
		{
			name:          "unmarked ordinary tasks",
			content:       "## 1. Verification\n\n- [ ] 1.1 Run tests.\n",
			markerPresent: false,
		},
		{
			name:          "marked section has no pending tasks",
			content:       strings.ReplaceAll(validTasks, "- [ ] 2.", "- [x] 2."),
			markerPresent: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newFixture(t, "")
			if err := os.WriteFile(fixture.tasksPath, []byte(test.content), 0o640); err != nil {
				t.Fatal(err)
			}
			result, err := fixture.service.Capture(context.Background(), fixture.resolve)
			if err != nil {
				t.Fatal(err)
			}
			if result.Status != "done" || result.Snapshot == nil {
				t.Fatalf("capture result = %+v", result)
			}
			if result.Snapshot.Marker.Present != test.markerPresent || len(result.Snapshot.PendingTasks) != 0 {
				t.Fatalf("snapshot = %+v", result.Snapshot)
			}
		})
	}
}

func TestCompleteRequiresMarkedNonEmptyPendingSet(t *testing.T) {
	tests := []struct {
		name    string
		content string
		code    string
	}{
		{name: "unmarked", content: "## 1. Verification\n\n- [ ] 1.1 Run tests.\n", code: "marker_required"},
		{name: "already complete", content: strings.ReplaceAll(validTasks, "- [ ] 2.", "- [x] 2."), code: "no_pending_tasks"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newFixture(t, "")
			if err := os.WriteFile(fixture.tasksPath, []byte(test.content), 0o640); err != nil {
				t.Fatal(err)
			}
			snapshot := captureSnapshot(t, fixture)
			before, err := os.ReadFile(fixture.tasksPath)
			if err != nil {
				t.Fatal(err)
			}
			result, err := fixture.service.Complete(context.Background(), fixture.resolve, passingRequest(snapshot))
			if err != nil {
				t.Fatal(err)
			}
			if result.Status != "blocked" || result.Completion != "rejected" || !hasConflict(result, test.code) {
				t.Fatalf("completion result = %+v", result)
			}
			after, err := os.ReadFile(fixture.tasksPath)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(after, before) {
				t.Fatal("rejected completion changed tasks.md")
			}
		})
	}
}

func TestCaptureRejectsPathEscapeAndSymlink(t *testing.T) {
	t.Run("outside change root", func(t *testing.T) {
		fixture := newFixture(t, "")
		outside := filepath.Join(fixture.root, "tasks.md")
		if err := os.WriteFile(outside, []byte(validTasks), 0o644); err != nil {
			t.Fatal(err)
		}
		fixture.runner.status = func() []byte {
			return statusJSON(t, fixture.root, fixture.changeRoot, outside, "done", nil)
		}
		_, err := fixture.service.Capture(context.Background(), fixture.resolve)
		if err == nil || !strings.Contains(err.Error(), "escapes") {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("symlink", func(t *testing.T) {
		fixture := newFixture(t, "")
		real := filepath.Join(fixture.changeRoot, "real.md")
		if err := os.Rename(fixture.tasksPath, real); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(real, fixture.tasksPath); err != nil {
			t.Fatal(err)
		}
		_, err := fixture.service.Capture(context.Background(), fixture.resolve)
		if err == nil || !strings.Contains(err.Error(), "symlink") {
			t.Fatalf("error = %v", err)
		}
	})
}

func TestCompleteRequiresExactTaskSetAndSuccessfulEvidence(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*CompleteRequest)
		code   string
	}{
		{name: "partial task set", code: "partial_task_set", mutate: func(request *CompleteRequest) {
			request.Tasks = request.Tasks[:1]
			request.Evidence = request.Evidence[:1]
		}},
		{name: "missing evidence", code: "incomplete_evidence", mutate: func(request *CompleteRequest) {
			request.Evidence = request.Evidence[:1]
		}},
		{name: "failed evidence", code: "unsuccessful_evidence", mutate: func(request *CompleteRequest) {
			request.Evidence[1].Commands[0].ExitCode = 1
		}},
		{name: "global fail", code: "verdict_not_pass", mutate: func(request *CompleteRequest) {
			request.Verdict = "fail"
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newFixture(t, "")
			snapshot := captureSnapshot(t, fixture)
			before, err := os.ReadFile(fixture.tasksPath)
			if err != nil {
				t.Fatal(err)
			}
			request := passingRequest(snapshot)
			test.mutate(&request)
			result, err := fixture.service.Complete(context.Background(), fixture.resolve, request)
			if err != nil {
				t.Fatal(err)
			}
			if result.Status != "blocked" || !hasConflict(result, test.code) {
				t.Fatalf("result = %+v", result)
			}
			after, err := os.ReadFile(fixture.tasksPath)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(after, before) {
				t.Fatal("rejected completion changed tasks.md")
			}
		})
	}
}

func TestCompleteRejectsStaleContentAndArtifactState(t *testing.T) {
	t.Run("content", func(t *testing.T) {
		fixture := newFixture(t, "")
		snapshot := captureSnapshot(t, fixture)
		changed := strings.Replace(validTasks, "Implement feature.", "Implement corrected feature.", 1)
		if err := os.WriteFile(fixture.tasksPath, []byte(changed), 0o640); err != nil {
			t.Fatal(err)
		}
		result, err := fixture.service.Complete(context.Background(), fixture.resolve, passingRequest(snapshot))
		if err != nil {
			t.Fatal(err)
		}
		if result.Completion != "conflict" || !hasConflict(result, "content_changed") {
			t.Fatalf("result = %+v", result)
		}
		assertAllPending(t, fixture.tasksPath)
	})
	t.Run("artifact", func(t *testing.T) {
		fixture := newFixture(t, "")
		snapshot := captureSnapshot(t, fixture)
		fixture.runner.status = func() []byte {
			return statusJSON(t, fixture.root, fixture.changeRoot, fixture.tasksPath, "changed", nil)
		}
		result, err := fixture.service.Complete(context.Background(), fixture.resolve, passingRequest(snapshot))
		if err != nil {
			t.Fatal(err)
		}
		if result.Completion != "conflict" || !hasConflict(result, "artifact_changed") {
			t.Fatalf("result = %+v", result)
		}
		assertAllPending(t, fixture.tasksPath)
	})
	t.Run("invalid fresh marker state is a structured conflict", func(t *testing.T) {
		fixture := newFixture(t, "")
		snapshot := captureSnapshot(t, fixture)
		changed := strings.Replace(validTasks, OwnerMarker, "<!-- owner:openspec-verifier -->", 1)
		if err := os.WriteFile(fixture.tasksPath, []byte(changed), 0o640); err != nil {
			t.Fatal(err)
		}
		result, err := fixture.service.Complete(context.Background(), fixture.resolve, passingRequest(snapshot))
		if err != nil {
			t.Fatal(err)
		}
		if result.Status != "blocked" || result.Completion != "conflict" || !hasConflict(result, "fresh_resolution_failed") {
			t.Fatalf("result = %+v", result)
		}
		assertAllPending(t, fixture.tasksPath)
	})
}

func TestCompleteDetectsConcurrentChangeBeforeRename(t *testing.T) {
	fixture := newFixture(t, "")
	snapshot := captureSnapshot(t, fixture)
	fixture.service.beforeRename = func(path string) error {
		changed := strings.Replace(validTasks, "Implement feature.", "Concurrent edit.", 1)
		return os.WriteFile(path, []byte(changed), 0o640)
	}
	result, err := fixture.service.Complete(context.Background(), fixture.resolve, passingRequest(snapshot))
	if err != nil || result.Status != "blocked" || result.Completion != "conflict" || !hasConflict(result, "atomic_commit_failed") {
		t.Fatalf("result = %+v, error = %v", result, err)
	}
	assertAllPending(t, fixture.tasksPath)
}

func TestCompletePrecommitFailureLeavesOriginalUntouched(t *testing.T) {
	fixture := newFixture(t, "")
	snapshot := captureSnapshot(t, fixture)
	before, err := os.ReadFile(fixture.tasksPath)
	if err != nil {
		t.Fatal(err)
	}
	fixture.service.beforeRename = func(string) error { return errors.New("injected preparation failure") }
	result, err := fixture.service.Complete(context.Background(), fixture.resolve, passingRequest(snapshot))
	if err == nil || result.Completion != "conflict" {
		t.Fatalf("result = %+v, error = %v", result, err)
	}
	after, err := os.ReadFile(fixture.tasksPath)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(after, before) {
		t.Fatal("precommit failure changed original tasks.md")
	}
	matches, err := filepath.Glob(filepath.Join(fixture.changeRoot, ".tasks.md.tmp-*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 0 {
		t.Fatalf("temporary files retained: %v", matches)
	}
}

func hasConflict(result Result, code string) bool {
	for _, conflict := range result.Conflicts {
		if conflict.Code == code {
			return true
		}
	}
	return false
}

func assertAllPending(t *testing.T, path string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "- [ ] 2.1") || !strings.Contains(string(content), "- [ ] 2.2") {
		t.Fatalf("marked tasks were partially changed: %s", content)
	}
}
