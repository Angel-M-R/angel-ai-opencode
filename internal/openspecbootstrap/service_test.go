package openspecbootstrap

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type scriptedStep struct {
	dir     string // "" skips the directory assertion
	command string // space-joined executable and arguments
	output  string
	exit    int
}

type scriptedRunner struct {
	t     *testing.T
	steps []scriptedStep
	index int
}

func (r *scriptedRunner) Run(_ context.Context, dir, name string, args ...string) ([]byte, int, error) {
	r.t.Helper()
	command := strings.Join(append([]string{name}, args...), " ")
	if r.index >= len(r.steps) {
		r.t.Fatalf("unexpected command %q after the scripted sequence ended", command)
	}
	step := r.steps[r.index]
	r.index++
	if command != step.command {
		r.t.Fatalf("command %d = %q, want %q", r.index, command, step.command)
	}
	if step.dir != "" && dir != step.dir {
		r.t.Fatalf("command %q ran in %q, want %q", command, dir, step.dir)
	}
	var err error
	if step.exit != 0 {
		err = fmt.Errorf("exit status %d", step.exit)
	}
	return []byte(step.output), step.exit, err
}

func (r *scriptedRunner) assertDrained() {
	r.t.Helper()
	if r.index != len(r.steps) {
		r.t.Fatalf("only %d of %d scripted commands ran", r.index, len(r.steps))
	}
}

func lookPathAll(string) (string, error) { return "/bin/fake", nil }

func lookPathMissing(missing string) func(string) (string, error) {
	return func(name string) (string, error) {
		if name == missing {
			return "", fmt.Errorf("%s: not found", name)
		}
		return "/bin/fake", nil
	}
}

// hostWithCoreSkills builds a tool host containing every official core skill
// in the generated directory-with-SKILL.md layout.
func hostWithCoreSkills(t *testing.T) string {
	t.Helper()
	host := t.TempDir()
	for _, name := range coreSkills {
		dir := filepath.Join(host, ".opencode", "skills", name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("skill\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return host
}

func contextNearest(path string) string {
	return fmt.Sprintf(`{"root":{"path":%q,"source":"nearest","role":"openspec_root"},"members":[],"status":[]}`, path)
}

func listRoot(path string) string {
	return fmt.Sprintf(`{"changes":[],"root":{"path":%q,"source":"nearest"}}`, path)
}

func run(t *testing.T, service *Service, request Request) Result {
	t.Helper()
	result, err := service.Run(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestRunBlocksWhenOpenSpecBinaryIsMissing(t *testing.T) {
	runner := &scriptedRunner{t: t}
	service := &Service{runner: runner, lookPath: lookPathMissing("openspec")}
	result := run(t, service, Request{WorkingDirectory: t.TempDir()})
	if result.Status != "blocked" || !strings.Contains(result.BlockingReason, "OpenSpec extra") {
		t.Fatalf("result = %+v", result)
	}
	if len(result.Commands) != 0 {
		t.Fatalf("commands = %+v, want none", result.Commands)
	}
}

func TestRunBlocksWhenCoreProfileCannotBePinned(t *testing.T) {
	directory := t.TempDir()
	runner := &scriptedRunner{t: t, steps: []scriptedStep{
		{dir: directory, command: "openspec config profile core", output: "unknown profile", exit: 1},
	}}
	service := &Service{runner: runner, lookPath: lookPathAll}
	result := run(t, service, Request{WorkingDirectory: directory})
	runner.assertDrained()
	if result.Status != "blocked" || !strings.Contains(result.BlockingReason, "config profile core") {
		t.Fatalf("result = %+v", result)
	}
	if len(result.Commands) != 1 || result.Commands[0].ExitCode != 1 {
		t.Fatalf("commands = %+v", result.Commands)
	}
}

func TestRunBlocksWhenPlanningRootCannotBeResolved(t *testing.T) {
	directory := t.TempDir()
	runner := &scriptedRunner{t: t, steps: []scriptedStep{
		{command: "openspec config profile core"},
		{command: "openspec config set delivery both"},
		{command: "openspec list --json", output: `{"changes":[],"root":null}`},
	}}
	service := &Service{runner: runner, lookPath: lookPathAll}
	result := run(t, service, Request{WorkingDirectory: directory})
	runner.assertDrained()
	if result.Status != "blocked" || !strings.Contains(result.BlockingReason, "planning root") {
		t.Fatalf("result = %+v", result)
	}
}

func TestRunReadyOnInitializedHostWithExistingCodeGraph(t *testing.T) {
	host := hostWithCoreSkills(t)
	if err := os.Mkdir(filepath.Join(host, ".codegraph"), 0o755); err != nil {
		t.Fatal(err)
	}
	runner := &scriptedRunner{t: t, steps: []scriptedStep{
		{dir: host, command: "openspec config profile core"},
		{dir: host, command: "openspec config set delivery both"},
		{dir: host, command: "openspec list --json", output: listRoot(host)},
		{dir: host, command: "openspec context --json", output: contextNearest(host)},
		{dir: host, command: "openspec update", output: "updated"},
		{dir: host, command: "openspec list --json", output: listRoot(host)},
	}}
	service := &Service{runner: runner, lookPath: lookPathAll}
	result := run(t, service, Request{WorkingDirectory: host})
	runner.assertDrained()
	if result.Status != "ready" || result.IntegrationKey != host || result.ToolHost != host {
		t.Fatalf("result = %+v", result)
	}
	if len(result.Warnings) != 0 {
		t.Fatalf("warnings = %+v", result.Warnings)
	}
}

func TestRunInitializesAnUninitializedHostAndWarnsWithoutCodeGraph(t *testing.T) {
	host := hostWithCoreSkills(t)
	uninitializedContext := `{"root":null,"members":[],"status":[{"severity":"error","code":"no_openspec_root"}]}`
	runner := &scriptedRunner{t: t, steps: []scriptedStep{
		{command: "openspec config profile core"},
		{command: "openspec config set delivery both"},
		{command: "openspec list --json", output: listRoot(host)},
		{command: "openspec context --json", output: uninitializedContext, exit: 1},
		{dir: host, command: "openspec init --tools opencode"},
		{command: "openspec list --json", output: listRoot(host)},
	}}
	service := &Service{runner: runner, lookPath: lookPathMissing("codegraph")}
	result := run(t, service, Request{WorkingDirectory: host})
	runner.assertDrained()
	if result.Status != "ready" {
		t.Fatalf("result = %+v", result)
	}
	if len(result.Warnings) != 1 || !strings.Contains(result.Warnings[0].Note, "codegraph") {
		t.Fatalf("warnings = %+v", result.Warnings)
	}
}

func TestRunTreatsStoreSourcedContextRootAsUninitializedHost(t *testing.T) {
	host := hostWithCoreSkills(t)
	if err := os.Mkdir(filepath.Join(host, ".codegraph"), 0o755); err != nil {
		t.Fatal(err)
	}
	storeContext := `{"root":{"path":"/elsewhere/store","source":"global"},"members":[],"status":[]}`
	runner := &scriptedRunner{t: t, steps: []scriptedStep{
		{command: "openspec config profile core"},
		{command: "openspec config set delivery both"},
		{command: "openspec list --json --store docs", output: listRoot("/elsewhere/store")},
		{command: "openspec context --json", output: storeContext},
		{dir: host, command: "openspec init --tools opencode"},
		{command: "openspec list --json --store docs", output: listRoot("/elsewhere/store")},
	}}
	service := &Service{runner: runner, lookPath: lookPathAll}
	result := run(t, service, Request{WorkingDirectory: host, Store: "docs"})
	runner.assertDrained()
	if result.Status != "ready" {
		t.Fatalf("result = %+v", result)
	}
	if want := "store:docs@" + host; result.IntegrationKey != want {
		t.Fatalf("integration key = %q, want %q", result.IntegrationKey, want)
	}
	if result.PlanningRoot != "/elsewhere/store" {
		t.Fatalf("planning root = %q", result.PlanningRoot)
	}
}

func TestRunRepairsInitializedHostMissingSkillsWithSingleInit(t *testing.T) {
	host := t.TempDir() // no skills: update alone leaves the integration bare
	runner := &scriptedRunner{t: t, steps: []scriptedStep{
		{command: "openspec config profile core"},
		{command: "openspec config set delivery both"},
		{command: "openspec list --json", output: listRoot(host)},
		{command: "openspec context --json", output: contextNearest(host)},
		{dir: host, command: "openspec update", output: "No tools are configured."},
		{dir: host, command: "openspec init --tools opencode"},
		{command: "openspec list --json", output: listRoot(host)},
	}}
	service := &Service{runner: runner, lookPath: lookPathAll}
	result := run(t, service, Request{WorkingDirectory: host})
	runner.assertDrained()
	if result.Status != "blocked" || !strings.Contains(result.BlockingReason, "missing core skills") {
		t.Fatalf("result = %+v", result)
	}
}

func TestRunBlocksWhenPlanningRootChangesAfterSetup(t *testing.T) {
	host := hostWithCoreSkills(t)
	runner := &scriptedRunner{t: t, steps: []scriptedStep{
		{command: "openspec config profile core"},
		{command: "openspec config set delivery both"},
		{command: "openspec list --json", output: listRoot(host)},
		{command: "openspec context --json", output: contextNearest(host)},
		{dir: host, command: "openspec update"},
		{command: "openspec list --json", output: listRoot("/somewhere/else")},
	}}
	service := &Service{runner: runner, lookPath: lookPathAll}
	result := run(t, service, Request{WorkingDirectory: host})
	runner.assertDrained()
	if result.Status != "blocked" || !strings.Contains(result.BlockingReason, "changed after integration setup") {
		t.Fatalf("result = %+v", result)
	}
}

func TestRunKeepsReadyWhenCodeGraphInitFails(t *testing.T) {
	host := hostWithCoreSkills(t)
	runner := &scriptedRunner{t: t, steps: []scriptedStep{
		{command: "openspec config profile core"},
		{command: "openspec config set delivery both"},
		{command: "openspec list --json", output: listRoot(host)},
		{command: "openspec context --json", output: contextNearest(host)},
		{dir: host, command: "openspec update"},
		{command: "openspec list --json", output: listRoot(host)},
		{dir: host, command: "codegraph init " + host, output: "boom", exit: 2},
	}}
	service := &Service{runner: runner, lookPath: lookPathAll}
	result := run(t, service, Request{WorkingDirectory: host})
	runner.assertDrained()
	if result.Status != "ready" {
		t.Fatalf("result = %+v", result)
	}
	if len(result.Warnings) != 1 || !strings.Contains(result.Warnings[0].Note, "exited 2") {
		t.Fatalf("warnings = %+v", result.Warnings)
	}
	last := result.Commands[len(result.Commands)-1]
	if last.ExitCode != 2 || last.Command[0] != "codegraph" {
		t.Fatalf("last command = %+v", last)
	}
}

func TestRunRequiresWorkingDirectory(t *testing.T) {
	service := &Service{runner: &scriptedRunner{t: t}, lookPath: lookPathAll}
	if _, err := service.Run(context.Background(), Request{}); err == nil {
		t.Fatal("expected an error for a missing working directory")
	}
}
