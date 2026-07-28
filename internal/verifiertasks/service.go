package verifiertasks

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

const snapshotVersion = 1

type ResolveRequest struct {
	Change           string `json:"change"`
	Store            string `json:"store,omitempty"`
	WorkingDirectory string `json:"-"`
}

type FileIdentity struct {
	Mode             uint32 `json:"mode"`
	Size             int64  `json:"size"`
	ModifiedUnixNano int64  `json:"modifiedUnixNano"`
}

type Snapshot struct {
	Version         int             `json:"version"`
	Change          string          `json:"change"`
	Store           string          `json:"store,omitempty"`
	ContextIdentity string          `json:"contextIdentity"`
	PlanningRoot    string          `json:"planningRoot"`
	ChangeRoot      string          `json:"changeRoot"`
	TasksPath       string          `json:"tasksPath"`
	ArtifactDigest  string          `json:"artifactDigest"`
	ContentDigest   string          `json:"contentDigest"`
	File            FileIdentity    `json:"file"`
	Marker          MarkerStructure `json:"marker"`
	PendingTasks    []TaskIdentity  `json:"pendingTasks"`
}

type CommandRecord struct {
	Command  []string `json:"command"`
	ExitCode int      `json:"exitCode"`
}

type TaskEvidence struct {
	Task     TaskIdentity    `json:"task"`
	Status   string          `json:"status"`
	Commands []CommandRecord `json:"commands"`
}

type CompleteRequest struct {
	Snapshot Snapshot       `json:"snapshot"`
	Verdict  string         `json:"verdict"`
	Tasks    []TaskIdentity `json:"tasks"`
	Evidence []TaskEvidence `json:"evidence"`
}

type Conflict struct {
	Code   string `json:"code"`
	Detail string `json:"detail"`
}

type Result struct {
	Status     string          `json:"status"`
	Verdict    string          `json:"verdict"`
	Evidence   []TaskEvidence  `json:"evidence"`
	Completion string          `json:"completion"`
	Commands   []CommandRecord `json:"commands"`
	Conflicts  []Conflict      `json:"conflicts"`
	Snapshot   *Snapshot       `json:"snapshot,omitempty"`
}

type CommandRunner interface {
	Run(context.Context, string, string, ...string) ([]byte, int, error)
}

type execRunner struct{}

func (execRunner) Run(ctx context.Context, directory, name string, args ...string) ([]byte, int, error) {
	command := exec.CommandContext(ctx, name, args...)
	command.Dir = directory
	output, err := command.CombinedOutput()
	if err == nil {
		return output, 0, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return output, exitErr.ExitCode(), err
	}
	return output, -1, err
}

type Service struct {
	runner       CommandRunner
	beforeRename func(string) error
}

type capturedState struct {
	snapshot Snapshot
	content  []byte
	info     os.FileInfo
	parsed   parsedDocument
}

type stateConflictError struct {
	err error
}

func (err stateConflictError) Error() string {
	return err.err.Error()
}

func newStateConflictError(format string, arguments ...any) error {
	return stateConflictError{err: fmt.Errorf(format, arguments...)}
}

var errConcurrentCommit = errors.New("tasks.md changed during commit preparation")

func NewService() *Service {
	return &Service{runner: execRunner{}}
}

func (service *Service) Capture(ctx context.Context, request ResolveRequest) (Result, error) {
	state, command, err := service.capture(ctx, request)
	result := Result{
		Status:     "done",
		Verdict:    "not-verified",
		Completion: "not-attempted",
		Commands:   []CommandRecord{command},
		Conflicts:  []Conflict{},
		Evidence:   []TaskEvidence{},
	}
	if err != nil {
		result.Status = "blocked"
		result.Conflicts = []Conflict{{Code: "resolution_failed", Detail: err.Error()}}
		return result, err
	}
	result.Snapshot = &state.snapshot
	return result, nil
}

func (service *Service) Complete(ctx context.Context, resolve ResolveRequest, request CompleteRequest) (Result, error) {
	result := Result{
		Status:     "blocked",
		Verdict:    request.Verdict,
		Evidence:   request.Evidence,
		Completion: "rejected",
		Commands:   evidenceCommands(request.Evidence),
		Conflicts:  []Conflict{},
	}
	if conflicts := validateCompletionRequest(resolve, request); len(conflicts) > 0 {
		result.Conflicts = conflicts
		return result, nil
	}

	freshState, command, err := service.capture(ctx, resolve)
	result.Commands = append(result.Commands, command)
	if err != nil {
		result.Completion = "conflict"
		result.Conflicts = []Conflict{{Code: "fresh_resolution_failed", Detail: err.Error()}}
		var conflict stateConflictError
		if errors.As(err, &conflict) {
			return result, nil
		}
		return result, err
	}
	fresh := freshState.snapshot
	if conflicts := compareSnapshots(request.Snapshot, fresh); len(conflicts) > 0 {
		result.Completion = "conflict"
		result.Conflicts = conflicts
		return result, nil
	}

	content := freshState.content
	info := freshState.info
	parsed := freshState.parsed
	replacement := append([]byte(nil), content...)
	for _, task := range request.Tasks {
		offset, exists := parsed.pendingOffsets[task]
		if !exists || offset < 0 || offset >= len(replacement) || replacement[offset] != ' ' {
			result.Completion = "conflict"
			result.Conflicts = []Conflict{{Code: "task_transition_invalid", Detail: fmt.Sprintf("task %s is no longer pending", task.ID)}}
			return result, nil
		}
		replacement[offset] = 'x'
	}
	if err := atomicReplace(fresh.TasksPath, content, replacement, info, service.beforeRename); err != nil {
		result.Completion = "conflict"
		result.Conflicts = []Conflict{{Code: "atomic_commit_failed", Detail: err.Error()}}
		if errors.Is(err, errConcurrentCommit) {
			return result, nil
		}
		return result, err
	}
	result.Status = "done"
	result.Completion = "completed"
	return result, nil
}

func validateCompletionRequest(resolve ResolveRequest, request CompleteRequest) []Conflict {
	var conflicts []Conflict
	if request.Verdict != "pass" {
		conflicts = append(conflicts, Conflict{Code: "verdict_not_pass", Detail: "global verdict must be exactly pass"})
	}
	if request.Snapshot.Version != snapshotVersion {
		conflicts = append(conflicts, Conflict{Code: "snapshot_version", Detail: "unsupported or missing snapshot version"})
	}
	if resolve.Change == "" || request.Snapshot.Change != resolve.Change || request.Snapshot.Store != resolve.Store {
		conflicts = append(conflicts, Conflict{Code: "context_mismatch", Detail: "completion context differs from snapshot context"})
	}
	if !request.Snapshot.Marker.Present {
		conflicts = append(conflicts, Conflict{Code: "marker_required", Detail: "completion requires a valid verifier owner marker"})
	}
	if len(request.Snapshot.PendingTasks) == 0 {
		conflicts = append(conflicts, Conflict{Code: "no_pending_tasks", Detail: "completion requires a non-empty pending marked task set"})
	} else if !reflect.DeepEqual(request.Tasks, request.Snapshot.PendingTasks) {
		conflicts = append(conflicts, Conflict{Code: "partial_task_set", Detail: "requested tasks must exactly equal the complete pending marked task set"})
	}
	if evidenceConflicts := validateEvidence(request.Tasks, request.Evidence); len(evidenceConflicts) > 0 {
		conflicts = append(conflicts, evidenceConflicts...)
	}
	return conflicts
}

func validateEvidence(tasks []TaskIdentity, evidence []TaskEvidence) []Conflict {
	if len(tasks) != len(evidence) {
		return []Conflict{{Code: "incomplete_evidence", Detail: "evidence must contain exactly one entry for every requested task"}}
	}
	wanted := make(map[TaskIdentity]struct{}, len(tasks))
	for _, task := range tasks {
		wanted[task] = struct{}{}
	}
	seen := make(map[TaskIdentity]struct{}, len(evidence))
	for _, item := range evidence {
		if _, exists := wanted[item.Task]; !exists {
			return []Conflict{{Code: "unmatched_evidence", Detail: fmt.Sprintf("evidence does not match marked task %q", item.Task.ID)}}
		}
		if _, exists := seen[item.Task]; exists {
			return []Conflict{{Code: "ambiguous_evidence", Detail: fmt.Sprintf("duplicate evidence for task %q", item.Task.ID)}}
		}
		seen[item.Task] = struct{}{}
		if item.Status != "pass" || len(item.Commands) == 0 {
			return []Conflict{{Code: "unsuccessful_evidence", Detail: fmt.Sprintf("task %q lacks successful executed evidence", item.Task.ID)}}
		}
		for _, command := range item.Commands {
			if len(command.Command) == 0 || command.ExitCode != 0 {
				return []Conflict{{Code: "unsuccessful_evidence", Detail: fmt.Sprintf("task %q has missing or non-zero command evidence", item.Task.ID)}}
			}
		}
	}
	return nil
}

func evidenceCommands(evidence []TaskEvidence) []CommandRecord {
	var commands []CommandRecord
	for _, item := range evidence {
		commands = append(commands, item.Commands...)
	}
	return commands
}

type statusResponse struct {
	ChangeName   string `json:"changeName"`
	PlanningHome struct {
		Root string `json:"root"`
	} `json:"planningHome"`
	ChangeRoot    string `json:"changeRoot"`
	ArtifactPaths struct {
		Tasks struct {
			ResolvedOutputPath  string   `json:"resolvedOutputPath"`
			ExistingOutputPaths []string `json:"existingOutputPaths"`
		} `json:"tasks"`
	} `json:"artifactPaths"`
	Artifacts []struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	} `json:"artifacts"`
}

func (service *Service) capture(ctx context.Context, request ResolveRequest) (capturedState, CommandRecord, error) {
	if strings.TrimSpace(request.Change) == "" {
		return capturedState{}, CommandRecord{}, fmt.Errorf("change is required")
	}
	if strings.TrimSpace(request.WorkingDirectory) == "" {
		return capturedState{}, CommandRecord{}, fmt.Errorf("working directory is required")
	}
	args := []string{"status", "--change", request.Change, "--json"}
	if request.Store != "" {
		args = append(args, "--store", request.Store)
	}
	output, exitCode, runErr := service.runner.Run(ctx, request.WorkingDirectory, "openspec", args...)
	command := CommandRecord{Command: append([]string{"openspec"}, args...), ExitCode: exitCode}
	if runErr != nil {
		return capturedState{}, command, fmt.Errorf("resolving OpenSpec status: %w: %s", runErr, strings.TrimSpace(string(output)))
	}
	var status statusResponse
	decoder := json.NewDecoder(bytes.NewReader(output))
	if err := decoder.Decode(&status); err != nil {
		return capturedState{}, command, fmt.Errorf("decoding OpenSpec status: %w", err)
	}
	if status.ChangeName != request.Change {
		return capturedState{}, command, newStateConflictError("resolved change %q, expected %q", status.ChangeName, request.Change)
	}
	tasksStatus := ""
	for _, artifact := range status.Artifacts {
		if artifact.ID == "tasks" {
			tasksStatus = artifact.Status
			break
		}
	}
	if tasksStatus == "" || status.ArtifactPaths.Tasks.ResolvedOutputPath == "" {
		return capturedState{}, command, newStateConflictError("resolved status has no tasks artifact")
	}
	planningRoot, err := normalizeExistingDirectory(status.PlanningHome.Root)
	if err != nil {
		return capturedState{}, command, newStateConflictError("normalizing planning root: %v", err)
	}
	changeRoot, err := normalizeExistingDirectory(status.ChangeRoot)
	if err != nil {
		return capturedState{}, command, newStateConflictError("normalizing change root: %v", err)
	}
	content, info, tasksPath, err := readResolvedTasks(status.ArtifactPaths.Tasks.ResolvedOutputPath, changeRoot, planningRoot)
	if err != nil {
		return capturedState{}, command, stateConflictError{err: err}
	}
	parsed, err := parseDocument(content)
	if err != nil {
		return capturedState{}, command, newStateConflictError("parsing resolved tasks: %v", err)
	}
	contextIdentity := "repo:" + planningRoot
	if request.Store != "" {
		contextIdentity = "store:" + request.Store
	}
	artifactPaths := append([]string(nil), status.ArtifactPaths.Tasks.ExistingOutputPaths...)
	for index, path := range artifactPaths {
		artifactPaths[index] = filepath.Clean(path)
	}
	sort.Strings(artifactPaths)
	artifactDigest := digest([]byte(strings.Join(append([]string{tasksStatus, tasksPath}, artifactPaths...), "\x00")))
	snapshot := Snapshot{
		Version:         snapshotVersion,
		Change:          request.Change,
		Store:           request.Store,
		ContextIdentity: contextIdentity,
		PlanningRoot:    planningRoot,
		ChangeRoot:      changeRoot,
		TasksPath:       tasksPath,
		ArtifactDigest:  artifactDigest,
		ContentDigest:   digest(content),
		File:            fileIdentity(info),
		Marker:          parsed.Marker,
		PendingTasks:    parsed.Pending,
	}
	return capturedState{snapshot: snapshot, content: content, info: info, parsed: parsed}, command, nil
}

func compareSnapshots(captured, fresh Snapshot) []Conflict {
	checks := []struct {
		code   string
		detail string
		equal  bool
	}{
		{"snapshot_identity", "change or store snapshot identity changed", captured.Version == fresh.Version && captured.Change == fresh.Change && captured.Store == fresh.Store},
		{"context_changed", "OpenSpec context identity changed", captured.ContextIdentity == fresh.ContextIdentity && captured.PlanningRoot == fresh.PlanningRoot && captured.ChangeRoot == fresh.ChangeRoot},
		{"path_changed", "resolved tasks.md path changed", captured.TasksPath == fresh.TasksPath},
		{"artifact_changed", "resolved tasks artifact changed", captured.ArtifactDigest == fresh.ArtifactDigest},
		{"content_changed", "tasks.md content changed", captured.ContentDigest == fresh.ContentDigest && captured.File == fresh.File},
		{"marker_changed", "verifier marker structure changed", reflect.DeepEqual(captured.Marker, fresh.Marker)},
		{"task_set_changed", "pending verifier task set changed", reflect.DeepEqual(captured.PendingTasks, fresh.PendingTasks)},
	}
	var conflicts []Conflict
	for _, check := range checks {
		if !check.equal {
			conflicts = append(conflicts, Conflict{Code: check.code, Detail: check.detail})
		}
	}
	return conflicts
}

func normalizeExistingDirectory(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path is empty")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	real, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(real)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", real)
	}
	return filepath.Clean(real), nil
}

func readResolvedTasks(path, changeRoot, planningRoot string) ([]byte, os.FileInfo, string, error) {
	if filepath.Base(filepath.Clean(path)) != "tasks.md" {
		return nil, nil, "", fmt.Errorf("resolved tasks path does not name tasks.md")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, nil, "", err
	}
	if info, err := os.Lstat(abs); err != nil {
		return nil, nil, "", err
	} else if info.Mode()&os.ModeSymlink != 0 {
		return nil, nil, "", fmt.Errorf("resolved tasks path must not be a symlink")
	}
	real, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return nil, nil, "", err
	}
	real = filepath.Clean(real)
	content, info, err := readConfinedRegularFile(real, changeRoot, planningRoot)
	return content, info, real, err
}

func readConfinedRegularFile(path, changeRoot, planningRoot string) ([]byte, os.FileInfo, error) {
	if !pathWithin(planningRoot, path) || !pathWithin(changeRoot, path) {
		return nil, nil, fmt.Errorf("tasks path escapes the active change")
	}
	info, err := os.Lstat(path)
	if err != nil {
		return nil, nil, err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return nil, nil, fmt.Errorf("tasks path is not a regular non-symlink file")
	}
	content, err := os.ReadFile(path)
	return content, info, err
}

func pathWithin(root, path string) bool {
	relative, err := filepath.Rel(root, path)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) && !filepath.IsAbs(relative)
}

func digest(content []byte) string {
	sum := sha256.Sum256(content)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func fileIdentity(info os.FileInfo) FileIdentity {
	return FileIdentity{Mode: uint32(info.Mode()), Size: info.Size(), ModifiedUnixNano: info.ModTime().UnixNano()}
}

func atomicReplace(path string, original, replacement []byte, originalInfo os.FileInfo, beforeRename func(string) error) (returnErr error) {
	directory := filepath.Dir(path)
	temporary, err := os.CreateTemp(directory, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	committed := false
	defer func() {
		if !committed {
			_ = temporary.Close()
			_ = os.Remove(temporaryPath)
		}
	}()
	if err := temporary.Chmod(originalInfo.Mode().Perm()); err != nil {
		return err
	}
	if _, err := temporary.Write(replacement); err != nil {
		return err
	}
	if err := temporary.Sync(); err != nil {
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if beforeRename != nil {
		if err := beforeRename(path); err != nil {
			return err
		}
	}
	current, currentInfo, err := readConfinedRegularFile(path, directory, directory)
	if err != nil {
		return err
	}
	if !bytes.Equal(current, original) || fileIdentity(currentInfo) != fileIdentity(originalInfo) {
		return errConcurrentCommit
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return err
	}
	committed = true
	return nil
}
