// Package openspecbootstrap prepares a project (and optionally an explicit
// planning store) for the Angel AI OpenSpec workflow. It replaces the prompt
// procedure the orchestrator previously delegated to an LLM worker with the
// same deterministic sequence: pin the official core profile, resolve
// planning readiness and the tool host through the OpenSpec CLI, initialize
// or update the OpenCode integration, verify the generated core skills, and
// advisorily initialize CodeGraph.
package openspecbootstrap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// coreSkills are the six official core workflow skills a healthy OpenCode
// integration exposes under <tool-host>/.opencode/skills/ once the core
// profile is active.
var coreSkills = []string{
	"openspec-propose",
	"openspec-explore",
	"openspec-apply-change",
	"openspec-update-change",
	"openspec-sync-specs",
	"openspec-archive-change",
}

// recoverableContextCodes are the `openspec context --json` diagnostics that
// mean the working directory is simply an uninitialized tool host, fixable by
// running init there.
var recoverableContextCodes = map[string]bool{
	"no_openspec_root":               true,
	"no_root_with_registered_stores": true,
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

type Request struct {
	WorkingDirectory string
	Store            string
}

type CommandRecord struct {
	Command  []string `json:"command"`
	ExitCode int      `json:"exitCode"`
}

type Warning struct {
	Command []string `json:"command,omitempty"`
	Note    string   `json:"note"`
}

type Result struct {
	Status         string          `json:"status"` // "ready" or "blocked"
	IntegrationKey string          `json:"integrationKey,omitempty"`
	ToolHost       string          `json:"toolHost,omitempty"`
	PlanningRoot   string          `json:"planningRoot,omitempty"`
	Commands       []CommandRecord `json:"commands"`
	Warnings       []Warning       `json:"warnings"`
	BlockingReason string          `json:"blockingReason,omitempty"`
}

type Service struct {
	runner   CommandRunner
	lookPath func(string) (string, error)
}

func NewService() *Service {
	return &Service{runner: execRunner{}, lookPath: exec.LookPath}
}

// Run executes the bootstrap sequence. A logical block is a complete
// structured result with a nil error; non-nil errors are reserved for
// malformed input.
func (s *Service) Run(ctx context.Context, request Request) (Result, error) {
	result := Result{
		Status:   "blocked",
		Commands: []CommandRecord{},
		Warnings: []Warning{},
	}
	directory := request.WorkingDirectory
	if strings.TrimSpace(directory) == "" {
		return result, errors.New("working directory is required")
	}

	record := func(dir, name string, args ...string) ([]byte, int) {
		output, code, _ := s.runner.Run(ctx, dir, name, args...)
		result.Commands = append(result.Commands, CommandRecord{
			Command:  append([]string{name}, args...),
			ExitCode: code,
		})
		return output, code
	}
	blocked := func(format string, args ...any) (Result, error) {
		result.BlockingReason = fmt.Sprintf(format, args...)
		return result, nil
	}

	if _, err := s.lookPath("openspec"); err != nil {
		return blocked("openspec CLI not found; install it with this repository installer's OpenSpec extra")
	}

	// The official core profile and delivery mode are Angel AI's OpenSpec
	// contract; a custom profile would generate an incomplete skill set.
	for _, args := range [][]string{
		{"config", "profile", "core"},
		{"config", "set", "delivery", "both"},
	} {
		if output, code := record(directory, "openspec", args...); code != 0 {
			return blocked("openspec %s failed: %s", strings.Join(args, " "), summarize(output))
		}
	}

	// Planning readiness comes only from CLI JSON, never from filesystem
	// presence.
	listArgs := []string{"list", "--json"}
	if request.Store != "" {
		listArgs = append(listArgs, "--store", request.Store)
	}
	output, code := record(directory, "openspec", listArgs...)
	planningRoot, rootErr := parseRootPath(output)
	if code != 0 || rootErr != nil {
		return blocked("openspec %s did not resolve a planning root: %s", strings.Join(listArgs, " "), summarize(output))
	}
	result.PlanningRoot = planningRoot

	// The tool host is resolved separately from the selected planning store:
	// it is the project whose OpenCode process must load the generated skills.
	output, _ = record(directory, "openspec", "context", "--json")
	contextState := parseContext(output)
	var toolHost string
	var initialized bool
	switch {
	case contextState.rootPath != "" && contextState.source == "nearest":
		toolHost = contextState.rootPath
		initialized = true
	case contextState.recoverable || (contextState.rootPath != "" && contextState.source != "nearest"):
		toolHost = directory
	default:
		return blocked("openspec context --json reported an unhealthy tool host: %s", summarize(output))
	}
	result.ToolHost = toolHost

	// Prepare the OpenCode integration in the tool host (never in the
	// selected store root). Initialize at most once.
	if initialized {
		if output, code = record(toolHost, "openspec", "update"); code != 0 {
			return blocked("openspec update failed in tool host: %s", summarize(output))
		}
		if len(missingCoreSkills(toolHost)) > 0 {
			if output, code = record(toolHost, "openspec", "init", "--tools", "opencode"); code != 0 {
				return blocked("openspec init --tools opencode failed in tool host: %s", summarize(output))
			}
		}
	} else {
		if output, code = record(toolHost, "openspec", "init", "--tools", "opencode"); code != 0 {
			return blocked("openspec init --tools opencode failed in tool host: %s", summarize(output))
		}
	}

	// Integration setup must not have changed which context planning
	// resolves to.
	output, code = record(directory, "openspec", listArgs...)
	recheckedRoot, recheckErr := parseRootPath(output)
	if code != 0 || recheckErr != nil {
		return blocked("openspec %s failed after integration setup: %s", strings.Join(listArgs, " "), summarize(output))
	}
	if recheckedRoot != planningRoot {
		return blocked("planning root changed after integration setup: %s became %s", planningRoot, recheckedRoot)
	}

	// The generated core skills are the integration health check; OpenSpec
	// owns them and a missing one means a corrupt integration, never a reason
	// to switch profiles.
	if missing := missingCoreSkills(toolHost); len(missing) > 0 {
		return blocked("incomplete OpenCode integration: missing core skills %s", strings.Join(missing, ", "))
	}

	// CodeGraph preparation is advisory and never blocks green readiness.
	if _, err := os.Stat(filepath.Join(toolHost, ".codegraph")); err != nil {
		if _, err := s.lookPath("codegraph"); err != nil {
			result.Warnings = append(result.Warnings, Warning{
				Note: "codegraph CLI not found; workers continue with filesystem tools",
			})
		} else if output, code = record(toolHost, "codegraph", "init", toolHost); code != 0 {
			result.Warnings = append(result.Warnings, Warning{
				Command: []string{"codegraph", "init", toolHost},
				Note:    fmt.Sprintf("codegraph init exited %d; workers continue with filesystem tools: %s", code, summarize(output)),
			})
		}
	}

	result.Status = "ready"
	if request.Store != "" {
		result.IntegrationKey = "store:" + request.Store + "@" + toolHost
	} else {
		result.IntegrationKey = planningRoot
	}
	return result, nil
}

// missingCoreSkills reports which official core skills are absent under the
// tool host's .opencode/skills/ directory, accepting either the generated
// directory-with-SKILL.md layout or a flat <name>.md file.
func missingCoreSkills(toolHost string) []string {
	var missing []string
	skillsDir := filepath.Join(toolHost, ".opencode", "skills")
	for _, name := range coreSkills {
		if _, err := os.Stat(filepath.Join(skillsDir, name, "SKILL.md")); err == nil {
			continue
		}
		if _, err := os.Stat(filepath.Join(skillsDir, name+".md")); err == nil {
			continue
		}
		missing = append(missing, name)
	}
	return missing
}

func parseRootPath(output []byte) (string, error) {
	var payload struct {
		Root *struct {
			Path string `json:"path"`
		} `json:"root"`
	}
	if err := json.Unmarshal(output, &payload); err != nil {
		return "", err
	}
	if payload.Root == nil || payload.Root.Path == "" {
		return "", errors.New("no resolved root in output")
	}
	return payload.Root.Path, nil
}

type contextState struct {
	rootPath    string
	source      string
	recoverable bool
}

func parseContext(output []byte) contextState {
	var payload struct {
		Root *struct {
			Path   string `json:"path"`
			Source string `json:"source"`
		} `json:"root"`
		Status []struct {
			Code string `json:"code"`
		} `json:"status"`
	}
	state := contextState{}
	if err := json.Unmarshal(output, &payload); err != nil {
		return state
	}
	if payload.Root != nil {
		state.rootPath = payload.Root.Path
		state.source = payload.Root.Source
	}
	for _, status := range payload.Status {
		if recoverableContextCodes[status.Code] {
			state.recoverable = true
		}
	}
	return state
}

// summarize compresses command output into a single diagnostic line.
func summarize(output []byte) string {
	text := strings.Join(strings.Fields(string(output)), " ")
	if text == "" {
		return "(no output)"
	}
	if len(text) > 240 {
		text = text[:240] + "…"
	}
	return text
}
