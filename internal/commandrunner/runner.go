// Package commandrunner executes bounded child processes for Angel AI.
package commandrunner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Cause classifies why a command did not complete successfully.
type Cause string

const (
	CauseInvalidConfiguration Cause = "invalid-configuration"
	CauseWorkingDirectory     Cause = "working-directory"
	CauseTimeout              Cause = "timeout"
	CauseCanceled             Cause = "canceled"
	CauseOutputLimit          Cause = "output-limit"
	CauseStart                Cause = "start"
	CauseExit                 Cause = "exit"
	CauseWait                 Cause = "wait"
)

// Error is the typed failure returned by Runner.
type Error struct {
	Cause    Cause
	ExitCode int
	err      error
}

func (err *Error) Error() string {
	switch err.Cause {
	case CauseInvalidConfiguration:
		return "command runner configuration is invalid"
	case CauseWorkingDirectory:
		return "command working directory is unavailable"
	case CauseTimeout:
		return "command timed out"
	case CauseCanceled:
		return "command was canceled"
	case CauseOutputLimit:
		return "command output exceeded its limit"
	case CauseStart:
		return "command could not start"
	case CauseExit:
		return fmt.Sprintf("command exited with code %d", err.ExitCode)
	default:
		return "command did not finish cleanly"
	}
}

func (err *Error) Unwrap() error {
	return err.err
}

// Result contains the data a caller may safely retain about one command.
type Result struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
	Duration time.Duration
}

// DiagnosticOutput joins stdout and stderr without losing their separation in
// the Result. Callers use it only when a failed command needs one message.
func (result Result) DiagnosticOutput() []byte {
	if len(result.Stdout) == 0 {
		return append([]byte(nil), result.Stderr...)
	}
	output := append([]byte(nil), result.Stdout...)
	if len(result.Stderr) == 0 {
		return output
	}
	if output[len(output)-1] != '\n' {
		output = append(output, '\n')
	}
	return append(output, result.Stderr...)
}

// Runner applies one timeout and one per-stream byte limit to every call.
type Runner struct {
	timeout        time.Duration
	outputLimit    int
	getwd          func() (string, error)
	userHomeDir    func() (string, error)
	tempDir        func() string
	stat           func(string) (os.FileInfo, error)
	commandContext func(context.Context, string, ...string) *exec.Cmd
}

// New builds a runner with a deadline and a byte limit for each output stream.
func New(timeout time.Duration, outputLimit int) *Runner {
	return &Runner{
		timeout:        timeout,
		outputLimit:    outputLimit,
		getwd:          os.Getwd,
		userHomeDir:    os.UserHomeDir,
		tempDir:        os.TempDir,
		stat:           os.Stat,
		commandContext: exec.CommandContext,
	}
}

// RunProject runs a command in the supplied project directory. It never falls
// back to another directory because that would change which project a command
// reads or modifies.
func (runner *Runner) RunProject(
	ctx context.Context,
	directory string,
	path string,
	args ...string,
) (Result, error) {
	if err := runner.configurationError(); err != nil {
		return failedResult(), err
	}
	if !runner.directoryExists(directory) {
		return failedResult(), commandError(CauseWorkingDirectory, -1, os.ErrNotExist)
	}
	return runner.run(ctx, directory, path, args...)
}

// RunGlobal runs a machine-level command from an existing directory. If the
// process cwd was removed, it tries the user's home and then the system temp
// directory.
func (runner *Runner) RunGlobal(ctx context.Context, path string, args ...string) (Result, error) {
	if err := runner.configurationError(); err != nil {
		return failedResult(), err
	}
	directory, err := runner.globalDirectory()
	if err != nil {
		return failedResult(), commandError(CauseWorkingDirectory, -1, err)
	}
	return runner.run(ctx, directory, path, args...)
}

func (runner *Runner) run(
	ctx context.Context,
	directory string,
	path string,
	args ...string,
) (Result, error) {
	startedAt := time.Now()
	if strings.TrimSpace(path) == "" {
		return resultAt(startedAt, nil, nil, -1), commandError(CauseStart, -1, exec.ErrNotFound)
	}

	runCtx, cancel := context.WithTimeout(ctx, runner.timeout)
	defer cancel()
	stdout := newLimitedBuffer(runner.outputLimit, cancel)
	stderr := newLimitedBuffer(runner.outputLimit, cancel)
	command := runner.commandContext(runCtx, path, args...)
	command.Dir = directory
	command.Stdout = stdout
	command.Stderr = stderr
	command.WaitDelay = time.Second
	configureProcessTermination(command)

	if err := command.Start(); err != nil {
		result := resultAt(startedAt, stdout.Bytes(), stderr.Bytes(), -1)
		return result, runner.classifyError(ctx, runCtx, CauseStart, -1, err)
	}
	err := command.Wait()
	exitCode := 0
	if err != nil {
		exitCode = -1
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		}
	}
	result := resultAt(startedAt, stdout.Bytes(), stderr.Bytes(), exitCode)
	if stdout.Exceeded() || stderr.Exceeded() {
		return result, commandError(CauseOutputLimit, -1, err)
	}
	if err == nil {
		return result, nil
	}
	return result, runner.classifyError(ctx, runCtx, CauseWait, exitCode, err)
}

func (runner *Runner) classifyError(
	parent context.Context,
	runCtx context.Context,
	fallback Cause,
	exitCode int,
	err error,
) error {
	if parentErr := parent.Err(); parentErr != nil {
		if errors.Is(parentErr, context.DeadlineExceeded) {
			return commandError(CauseTimeout, -1, parentErr)
		}
		return commandError(CauseCanceled, -1, parentErr)
	}
	if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
		return commandError(CauseTimeout, -1, context.DeadlineExceeded)
	}
	if errors.Is(runCtx.Err(), context.Canceled) {
		return commandError(CauseCanceled, -1, context.Canceled)
	}
	if fallback == CauseStart && isChdirError(err) {
		return commandError(CauseWorkingDirectory, -1, err)
	}
	if exitCode >= 0 {
		return commandError(CauseExit, exitCode, err)
	}
	return commandError(fallback, -1, err)
}

func (runner *Runner) configurationError() error {
	if runner == nil || runner.timeout <= 0 || runner.outputLimit <= 0 ||
		runner.getwd == nil || runner.userHomeDir == nil || runner.tempDir == nil ||
		runner.stat == nil || runner.commandContext == nil {
		return commandError(CauseInvalidConfiguration, -1, nil)
	}
	return nil
}

func (runner *Runner) globalDirectory() (string, error) {
	if runner == nil {
		return "", errors.New("runner is nil")
	}
	if directory, err := runner.getwd(); err == nil && runner.directoryExists(directory) {
		return directory, nil
	}
	if directory, err := runner.userHomeDir(); err == nil && runner.directoryExists(directory) {
		return directory, nil
	}
	if directory := runner.tempDir(); runner.directoryExists(directory) {
		return directory, nil
	}
	return "", errors.New("no existing directory is available for a global command")
}

func (runner *Runner) directoryExists(path string) bool {
	if runner == nil || runner.stat == nil || path == "" {
		return false
	}
	info, err := runner.stat(path)
	return err == nil && info.IsDir()
}

func isChdirError(err error) bool {
	var pathErr *os.PathError
	return errors.As(err, &pathErr) && pathErr.Op == "chdir"
}

func failedResult() Result {
	return Result{ExitCode: -1}
}

func resultAt(startedAt time.Time, stdout, stderr []byte, exitCode int) Result {
	return Result{
		Stdout:   append([]byte(nil), stdout...),
		Stderr:   append([]byte(nil), stderr...),
		ExitCode: exitCode,
		Duration: time.Since(startedAt),
	}
}

func commandError(cause Cause, exitCode int, err error) *Error {
	return &Error{Cause: cause, ExitCode: exitCode, err: err}
}

type limitedBuffer struct {
	buffer   bytes.Buffer
	limit    int
	cancel   context.CancelFunc
	exceeded bool
}

func newLimitedBuffer(limit int, cancel context.CancelFunc) *limitedBuffer {
	return &limitedBuffer{limit: limit, cancel: cancel}
}

func (buffer *limitedBuffer) Write(content []byte) (int, error) {
	remaining := buffer.limit - buffer.buffer.Len()
	if len(content) <= remaining {
		return buffer.buffer.Write(content)
	}
	if remaining > 0 {
		_, _ = buffer.buffer.Write(content[:remaining])
	}
	buffer.exceeded = true
	buffer.cancel()
	return len(content), nil
}

func (buffer *limitedBuffer) Bytes() []byte {
	return buffer.buffer.Bytes()
}

func (buffer *limitedBuffer) Exceeded() bool {
	return buffer.exceeded
}
