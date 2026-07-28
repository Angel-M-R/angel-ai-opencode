package verifiertasks

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestVerifierTaskWorkflowCompletesOnlyMarkedTasksInLocalAndStoreContexts(t *testing.T) {
	for _, store := range []string{"", "shared-specs"} {
		name := "local"
		if store != "" {
			name = "explicit store"
		}
		t.Run(name, func(t *testing.T) {
			fixture := newFixture(t, store)
			protected := seedWorkflowProtectedFiles(t, fixture)
			beforeInfo, err := os.Stat(fixture.tasksPath)
			if err != nil {
				t.Fatal(err)
			}
			snapshot := captureSnapshot(t, fixture)
			realRoot, err := filepath.EvalSymlinks(fixture.root)
			if err != nil {
				t.Fatal(err)
			}
			realTasks, err := filepath.EvalSymlinks(fixture.tasksPath)
			if err != nil {
				t.Fatal(err)
			}
			wantContext := "repo:" + realRoot
			if store != "" {
				wantContext = "store:" + store
			}
			if snapshot.ContextIdentity != wantContext || snapshot.TasksPath != realTasks || snapshot.ContentDigest == "" || snapshot.ArtifactDigest == "" || len(snapshot.PendingTasks) != 2 {
				t.Fatalf("snapshot = %+v", snapshot)
			}

			result, err := fixture.service.Complete(context.Background(), fixture.resolve, passingRequest(snapshot))
			if err != nil {
				t.Fatal(err)
			}
			if result.Status != "done" || result.Verdict != "pass" || result.Completion != "completed" || len(result.Evidence) != 2 || len(result.Conflicts) != 0 {
				t.Fatalf("completion result = %+v", result)
			}

			want := strings.Replace(validTasks, "- [ ] 2.1", "- [x] 2.1", 1)
			want = strings.Replace(want, "- [ ] 2.2", "- [x] 2.2", 1)
			if got := string(readWorkflowFile(t, fixture.tasksPath)); got != want {
				t.Fatalf("tasks.md changed outside marked checkbox tokens:\n%s", got)
			}
			assertNoPartialMarkedState(t, fixture.tasksPath)
			assertWorkflowFilesEqual(t, protected)
			afterInfo, err := os.Stat(fixture.tasksPath)
			if err != nil {
				t.Fatal(err)
			}
			if afterInfo.Mode().Perm() != beforeInfo.Mode().Perm() {
				t.Fatalf("mode = %v, want %v", afterInfo.Mode().Perm(), beforeInfo.Mode().Perm())
			}
			matches, err := filepath.Glob(filepath.Join(fixture.changeRoot, ".tasks.md.tmp-*"))
			if err != nil {
				t.Fatal(err)
			}
			if len(matches) != 0 {
				t.Fatalf("temporary files retained: %v", matches)
			}
			if len(fixture.runner.calls) != 2 {
				t.Fatalf("status command calls = %d, want snapshot and pre-write resolution", len(fixture.runner.calls))
			}
		})
	}
}

func seedWorkflowProtectedFiles(t *testing.T, fixture fixture) map[string][]byte {
	t.Helper()
	files := map[string][]byte{
		filepath.Join(fixture.root, "internal", "feature.go"):                                                []byte("package internal\n\nconst feature = true\n"),
		filepath.Join(fixture.root, "openspec", "changes", "other-change", "tasks.md"):                       []byte("## 1. Other change\n\n- [ ] 1.1 Leave this task alone.\n"),
		filepath.Join(fixture.root, "registered-stores", "other", "openspec", "changes", "demo", "tasks.md"): []byte(validTasks),
	}
	for path, content := range files {
		writeWorkflowFile(t, path, content)
	}
	return files
}

func writeWorkflowFile(t *testing.T, path string, content []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, content, 0o640); err != nil {
		t.Fatal(err)
	}
}

func readWorkflowFile(t *testing.T, path string) []byte {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return content
}

func assertWorkflowFilesEqual(t *testing.T, want map[string][]byte) {
	t.Helper()
	for path, content := range want {
		if got := readWorkflowFile(t, path); !reflect.DeepEqual(got, content) {
			t.Errorf("protected file changed: %s", path)
		}
	}
}

func assertNoPartialMarkedState(t *testing.T, path string) {
	t.Helper()
	parsed, err := parseDocument(readWorkflowFile(t, path))
	if err != nil {
		t.Fatal(err)
	}
	pending := 0
	for _, task := range parsed.Marker.Tasks {
		if task.Pending {
			pending++
		}
	}
	if pending != 0 && pending != len(parsed.Marker.Tasks) {
		t.Fatalf("marked section in %s is partially complete: %d/%d tasks remain", path, pending, len(parsed.Marker.Tasks))
	}
}
