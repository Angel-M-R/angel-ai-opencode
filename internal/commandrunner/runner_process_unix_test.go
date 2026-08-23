//go:build darwin || linux

package commandrunner

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRunnerKillsDescendantsOnTimeout(t *testing.T) {
	assertGrandchildDoesNotSurvive(t, New(500*time.Millisecond, 1024), "spawn-grandchild", CauseTimeout)
}

func TestRunnerKillsDescendantsOnOutputLimit(t *testing.T) {
	assertGrandchildDoesNotSurvive(t, New(5*time.Second, 64), "spawn-grandchild-overflow", CauseOutputLimit)
}

func assertGrandchildDoesNotSurvive(t *testing.T, runner *Runner, mode string, wantCause Cause) {
	t.Helper()
	directory := t.TempDir()
	readyPath := filepath.Join(directory, "grandchild-ready")
	markerPath := filepath.Join(directory, "grandchild-survived")
	path, args := helperCommand(mode, readyPath, markerPath)

	_, err := runner.RunProject(context.Background(), directory, path, args...)
	requireCause(t, err, wantCause)
	if _, err := os.Stat(readyPath); err != nil {
		t.Fatalf("grandchild did not start before cancellation: %v", err)
	}

	time.Sleep(1200 * time.Millisecond)
	if _, err := os.Stat(markerPath); err == nil {
		t.Fatal("grandchild survived command cancellation and wrote its marker")
	} else if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("checking grandchild marker: %v", err)
	}
}

func TestTerminateProcessTreeTreatsMissingProcessGroupAsDone(t *testing.T) {
	process, err := os.FindProcess(1 << 30)
	if err != nil {
		t.Fatal(err)
	}
	if err := terminateProcessTree(process); !errors.Is(err, os.ErrProcessDone) {
		t.Fatalf("terminate missing process group error = %v, want os.ErrProcessDone", err)
	}
	if err := terminateProcessTree(nil); !errors.Is(err, os.ErrProcessDone) {
		t.Fatalf("terminate nil process error = %v, want os.ErrProcessDone", err)
	}
}
