package commandrunner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunnerKeepsOutputStreamsSeparate(t *testing.T) {
	runner := New(2*time.Second, 64)
	path, args := helperCommand("streams")

	result, err := runner.RunProject(context.Background(), t.TempDir(), path, args...)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(result.Stdout), "stdout"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
	if got, want := string(result.Stderr), "stderr"; got != want {
		t.Fatalf("stderr = %q, want %q", got, want)
	}
	if result.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0", result.ExitCode)
	}
	if result.Duration <= 0 {
		t.Fatalf("duration = %s, want a measured duration", result.Duration)
	}
}

func TestRunnerReturnsTypedExitFailure(t *testing.T) {
	runner := New(2*time.Second, 64)
	path, args := helperCommand("exit")

	result, err := runner.RunProject(context.Background(), t.TempDir(), path, args...)
	requireCause(t, err, CauseExit)
	if result.ExitCode != 7 {
		t.Fatalf("exit code = %d, want 7", result.ExitCode)
	}
	if got, want := string(result.DiagnosticOutput()), "before exit\nfailed\n"; got != want {
		t.Fatalf("diagnostic output = %q, want %q", got, want)
	}
}

func TestRunnerStopsAtItsDeadline(t *testing.T) {
	runner := New(40*time.Millisecond, 64)
	path, args := helperCommand("sleep")
	started := time.Now()

	result, err := runner.RunProject(context.Background(), t.TempDir(), path, args...)
	requireCause(t, err, CauseTimeout)
	if result.ExitCode != -1 {
		t.Fatalf("exit code = %d, want -1", result.ExitCode)
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("timed command returned after %s", elapsed)
	}
}

func TestRunnerCapsOutputAndStopsTheProcess(t *testing.T) {
	const limit = 16
	for _, test := range []struct {
		name       string
		mode       string
		streamSize func(Result) int
	}{
		{name: "stdout", mode: "overflow-stdout", streamSize: func(result Result) int { return len(result.Stdout) }},
		{name: "stderr", mode: "overflow-stderr", streamSize: func(result Result) int { return len(result.Stderr) }},
	} {
		t.Run(test.name, func(t *testing.T) {
			runner := New(2*time.Second, limit)
			path, args := helperCommand(test.mode)

			result, err := runner.RunProject(context.Background(), t.TempDir(), path, args...)
			requireCause(t, err, CauseOutputLimit)
			if len(result.Stdout) > limit || len(result.Stderr) > limit {
				t.Fatalf("captured stdout/stderr lengths = %d/%d, limit %d", len(result.Stdout), len(result.Stderr), limit)
			}
			if got := test.streamSize(result); got != limit {
				t.Fatalf("limited stream length = %d, want %d", got, limit)
			}
		})
	}
}

func TestZeroRunnerReturnsTypedConfigurationFailure(t *testing.T) {
	var runner Runner
	path, args := helperCommand("pwd")

	_, err := runner.RunGlobal(context.Background(), path, args...)
	requireCause(t, err, CauseInvalidConfiguration)
}

func TestProjectCommandRejectsMissingWorkingDirectory(t *testing.T) {
	runner := New(2*time.Second, 64)
	path, args := helperCommand("pwd")
	missing := filepath.Join(t.TempDir(), "removed")

	result, err := runner.RunProject(context.Background(), missing, path, args...)
	requireCause(t, err, CauseWorkingDirectory)
	if result.ExitCode != -1 {
		t.Fatalf("exit code = %d, want -1", result.ExitCode)
	}
}

func TestGlobalCommandFallsBackWhenCurrentDirectoryWasDeleted(t *testing.T) {
	home := t.TempDir()
	runner := New(2*time.Second, 128)
	runner.getwd = func() (string, error) { return "", errors.New("cwd was deleted") }
	runner.userHomeDir = func() (string, error) { return home, nil }
	runner.tempDir = func() string { return t.TempDir() }
	path, args := helperCommand("pwd")

	result, err := runner.RunGlobal(context.Background(), path, args...)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := strings.TrimSpace(string(result.Stdout)), home; got != want {
		t.Fatalf("global command directory = %q, want %q", got, want)
	}
}

func TestGlobalCommandUsesTempWhenCurrentAndHomeDirectoriesAreUnavailable(t *testing.T) {
	temporary := t.TempDir()
	runner := New(2*time.Second, 128)
	runner.getwd = func() (string, error) { return "", errors.New("cwd was deleted") }
	runner.userHomeDir = func() (string, error) { return filepath.Join(t.TempDir(), "missing"), nil }
	runner.tempDir = func() string { return temporary }
	path, args := helperCommand("pwd")

	result, err := runner.RunGlobal(context.Background(), path, args...)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := strings.TrimSpace(string(result.Stdout)), temporary; got != want {
		t.Fatalf("global command directory = %q, want %q", got, want)
	}
}

func requireCause(t *testing.T, err error, want Cause) {
	t.Helper()
	if err == nil {
		t.Fatalf("error = nil, want cause %q", want)
	}
	var commandErr *Error
	if !errors.As(err, &commandErr) {
		t.Fatalf("error type = %T, want *commandrunner.Error", err)
	}
	if commandErr.Cause != want {
		t.Fatalf("error cause = %q, want %q", commandErr.Cause, want)
	}
}

func helperCommand(mode string, extra ...string) (string, []string) {
	args := []string{
		"-test.run=^TestCommandRunnerHelper$",
		"--",
		"commandrunner-helper",
		mode,
	}
	return os.Args[0], append(args, extra...)
}

func TestCommandRunnerHelper(t *testing.T) {
	var helperArgs []string
	for index := range os.Args {
		if os.Args[index] == "commandrunner-helper" && index+1 < len(os.Args) {
			helperArgs = os.Args[index+1:]
			break
		}
	}
	if len(helperArgs) == 0 {
		return
	}

	switch helperArgs[0] {
	case "streams":
		_, _ = os.Stdout.WriteString("stdout")
		_, _ = os.Stderr.WriteString("stderr")
		os.Exit(0)
	case "exit":
		_, _ = os.Stdout.WriteString("before exit")
		_, _ = os.Stderr.WriteString("failed\n")
		os.Exit(7)
	case "sleep":
		time.Sleep(5 * time.Second)
		os.Exit(0)
	case "overflow-stdout":
		_, _ = os.Stdout.WriteString(strings.Repeat("o", 256))
		os.Exit(0)
	case "overflow-stderr":
		_, _ = os.Stderr.WriteString(strings.Repeat("e", 256))
		os.Exit(0)
	case "pwd":
		directory, err := os.Getwd()
		if err != nil {
			_, _ = fmt.Fprint(os.Stderr, err)
			os.Exit(2)
		}
		_, _ = fmt.Fprint(os.Stdout, directory)
		os.Exit(0)
	case "spawn-grandchild", "spawn-grandchild-overflow":
		if len(helperArgs) != 3 {
			os.Exit(2)
		}
		path, args := helperCommand("delayed-marker", helperArgs[2])
		child := exec.Command(path, args...)
		child.Stdout = os.Stdout
		child.Stderr = os.Stderr
		if err := child.Start(); err != nil {
			_, _ = fmt.Fprint(os.Stderr, err)
			os.Exit(2)
		}
		if err := os.WriteFile(helperArgs[1], []byte("ready"), 0o600); err != nil {
			_, _ = fmt.Fprint(os.Stderr, err)
			os.Exit(2)
		}
		if helperArgs[0] == "spawn-grandchild-overflow" {
			_, _ = os.Stdout.WriteString(strings.Repeat("overflow", 1024))
		}
		time.Sleep(5 * time.Second)
		os.Exit(0)
	case "delayed-marker":
		if len(helperArgs) != 2 {
			os.Exit(2)
		}
		time.Sleep(time.Second)
		if err := os.WriteFile(helperArgs[1], []byte("survived"), 0o600); err != nil {
			_, _ = fmt.Fprint(os.Stderr, err)
			os.Exit(2)
		}
		os.Exit(0)
	default:
		_, _ = fmt.Fprintf(os.Stderr, "unknown helper mode %q", helperArgs[0])
		os.Exit(2)
	}
}
