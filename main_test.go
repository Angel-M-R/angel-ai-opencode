package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"reflect"
	"strings"
	"testing"

	assetfs "angel-ai-opencode/internal/assets"
	"angel-ai-opencode/internal/install"
	"angel-ai-opencode/internal/openspecbootstrap"
	"angel-ai-opencode/internal/verifiertasks"
)

type recordedUpdatePolicy struct {
	calls  []updatePolicyCall
	events *[]string
	err    error
}

type updatePolicyCall struct {
	version string
	forced  bool
}

func (policy *recordedUpdatePolicy) Run(currentVersion string, forced bool) error {
	policy.calls = append(policy.calls, updatePolicyCall{version: currentVersion, forced: forced})
	if policy.events != nil {
		*policy.events = append(*policy.events, "update")
	}
	return policy.err
}

func useVersion(t *testing.T, value string) {
	t.Helper()
	previous := version
	version = value
	t.Cleanup(func() { version = previous })
}

func TestRunCLINoArgumentsChecksForUpdatesBeforeTUI(t *testing.T) {
	useVersion(t, "v0.1.0")
	var events []string
	policy := &recordedUpdatePolicy{events: &events}
	var gotOptions rootOptions

	err := runCLI(nil, cliDependencies{
		stdout: &bytes.Buffer{},
		runInstaller: func(options rootOptions) error {
			events = append(events, "installer")
			gotOptions = options
			return nil
		},
		newUpdatePolicy: func() updatePolicy {
			events = append(events, "construct-update-policy")
			return policy
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(events, []string{"construct-update-policy", "update", "installer"}) {
		t.Fatalf("dispatch events = %v", events)
	}
	if !reflect.DeepEqual(policy.calls, []updatePolicyCall{{version: "v0.1.0", forced: false}}) {
		t.Fatalf("update calls = %#v", policy.calls)
	}
	if gotOptions != (rootOptions{}) {
		t.Fatalf("root options = %#v", gotOptions)
	}
}

func TestRunCLIAutomaticUpdateFailureWarnsAndContinuesTUI(t *testing.T) {
	useVersion(t, "v0.1.0")
	var output bytes.Buffer
	installerCalls := 0
	policy := &recordedUpdatePolicy{err: errors.New("offline")}

	err := runCLI(nil, cliDependencies{
		stdout: &output,
		runInstaller: func(rootOptions) error {
			installerCalls++
			return nil
		},
		newUpdatePolicy: func() updatePolicy { return policy },
	})
	if err != nil {
		t.Fatal(err)
	}
	if installerCalls != 1 {
		t.Fatalf("installer calls = %d, want 1", installerCalls)
	}
	if output.String() != "warning: update failed: offline\n" {
		t.Fatalf("output = %q", output.String())
	}
}

func TestRunCLIPreservesRootFlags(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		want       rootOptions
		wantUpdate bool
	}{
		{
			name:       "assets override",
			args:       []string{"--assets", "/tmp/assets"},
			want:       rootOptions{assetsDir: "/tmp/assets"},
			wantUpdate: true,
		},
		{
			name:       "target directory",
			args:       []string{"--target", "/tmp/config"},
			want:       rootOptions{configDir: "/tmp/config"},
			wantUpdate: true,
		},
		{
			name: "all",
			args: []string{"--all"},
			want: rootOptions{all: true},
		},
		{
			name: "all dry run",
			args: []string{"--all", "--dry-run"},
			want: rootOptions{all: true, dryRun: true},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			useVersion(t, "v0.1.0")
			policy := &recordedUpdatePolicy{}
			constructed := 0
			installerCalls := 0
			var got rootOptions

			err := runCLI(test.args, cliDependencies{
				stdout: &bytes.Buffer{},
				runInstaller: func(options rootOptions) error {
					installerCalls++
					got = options
					return nil
				},
				newUpdatePolicy: func() updatePolicy {
					constructed++
					return policy
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			if installerCalls != 1 {
				t.Fatalf("installer calls = %d", installerCalls)
			}
			if got != test.want {
				t.Fatalf("root options = %#v, want %#v", got, test.want)
			}
			wantConstructed := 0
			var wantPolicyCalls []updatePolicyCall
			if test.wantUpdate {
				wantConstructed = 1
				wantPolicyCalls = []updatePolicyCall{{version: "v0.1.0", forced: false}}
			}
			if constructed != wantConstructed {
				t.Fatalf("update policy constructions = %d, want %d", constructed, wantConstructed)
			}
			if !reflect.DeepEqual(policy.calls, wantPolicyCalls) {
				t.Fatalf("update calls = %#v, want %#v", policy.calls, wantPolicyCalls)
			}
		})
	}
}

func TestRunCLIRejectsUnknownCommandsAndFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "command", args: []string{"missing"}, want: "unknown command"},
		{name: "root flag", args: []string{"--missing"}, want: "flag provided but not defined"},
		{name: "version flag", args: []string{"version", "--missing"}, want: "flag provided but not defined"},
		{name: "update flag", args: []string{"update", "--missing"}, want: "flag provided but not defined"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := runCLI(test.args, cliDependencies{})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want text %q", err, test.want)
			}
		})
	}
}

func TestRunCLIVerifierTasksSnapshotPreservesExplicitStore(t *testing.T) {
	var output bytes.Buffer
	wantSnapshot := verifiertasks.Snapshot{Version: 1, Change: "demo", Store: "shared", ContextIdentity: "store:shared"}
	err := runCLI([]string{"verifier-tasks", "snapshot", "--change", "demo", "--store", "shared"}, cliDependencies{
		stdout:           &output,
		workingDirectory: func() (string, error) { return "/repo", nil },
		captureVerifierTasks: func(_ context.Context, request verifiertasks.ResolveRequest) (verifiertasks.Result, error) {
			if request.Change != "demo" || request.Store != "shared" || request.WorkingDirectory != "/repo" {
				t.Fatalf("resolve request = %+v", request)
			}
			return verifiertasks.Result{
				Status: "done", Verdict: "not-verified", Completion: "not-attempted",
				Evidence: []verifiertasks.TaskEvidence{}, Commands: []verifiertasks.CommandRecord{},
				Conflicts: []verifiertasks.Conflict{}, Snapshot: &wantSnapshot,
			}, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	var result verifiertasks.Result
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Snapshot == nil || !reflect.DeepEqual(*result.Snapshot, wantSnapshot) {
		t.Fatalf("result = %+v", result)
	}
}

func TestRunCLIOpenSpecBootstrapEmitsStructuredJSON(t *testing.T) {
	var output bytes.Buffer
	err := runCLI([]string{"openspec-bootstrap", "--store", "shared"}, cliDependencies{
		stdout:           &output,
		workingDirectory: func() (string, error) { return "/repo", nil },
		runOpenSpecBootstrap: func(_ context.Context, request openspecbootstrap.Request) (openspecbootstrap.Result, error) {
			if request.WorkingDirectory != "/repo" || request.Store != "shared" {
				t.Fatalf("bootstrap request = %+v", request)
			}
			return openspecbootstrap.Result{
				Status:         "ready",
				IntegrationKey: "store:shared@/repo",
				ToolHost:       "/repo",
				Commands:       []openspecbootstrap.CommandRecord{},
				Warnings:       []openspecbootstrap.Warning{},
			}, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	var result openspecbootstrap.Result
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Status != "ready" || result.IntegrationKey != "store:shared@/repo" {
		t.Fatalf("result = %+v", result)
	}
}

func TestRunCLIVerifierTasksCompleteReadsStructuredJSONFromStdin(t *testing.T) {
	task := verifiertasks.TaskIdentity{ID: "2.1", Text: "2.1 Run focused tests."}
	request := verifiertasks.CompleteRequest{
		Snapshot: verifiertasks.Snapshot{Version: 1, Change: "demo", PendingTasks: []verifiertasks.TaskIdentity{task}},
		Verdict:  "pass",
		Tasks:    []verifiertasks.TaskIdentity{task},
		Evidence: []verifiertasks.TaskEvidence{{
			Task: task, Status: "pass",
			Commands: []verifiertasks.CommandRecord{{Command: []string{"go", "test", "./focused"}, ExitCode: 0}},
		}},
	}
	input, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	err = runCLI([]string{"verifier-tasks", "complete", "--change", "demo"}, cliDependencies{
		stdin:            bytes.NewReader(input),
		stdout:           &output,
		workingDirectory: func() (string, error) { return "/repo", nil },
		completeVerifierTasks: func(_ context.Context, resolve verifiertasks.ResolveRequest, got verifiertasks.CompleteRequest) (verifiertasks.Result, error) {
			if resolve.Change != "demo" || resolve.Store != "" || resolve.WorkingDirectory != "/repo" {
				t.Fatalf("resolve request = %+v", resolve)
			}
			if !reflect.DeepEqual(got, request) {
				t.Fatalf("completion request = %+v, want %+v", got, request)
			}
			return verifiertasks.Result{
				Status: "done", Verdict: "pass", Evidence: got.Evidence, Completion: "completed",
				Commands: got.Evidence[0].Commands, Conflicts: []verifiertasks.Conflict{},
			}, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	var result verifiertasks.Result
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Status != "done" || result.Verdict != "pass" || result.Completion != "completed" || len(result.Evidence) != 1 || len(result.Commands) != 1 {
		t.Fatalf("result = %+v", result)
	}
}

func TestRunCLIVerifierTasksCompletionExitContract(t *testing.T) {
	request := verifiertasks.CompleteRequest{
		Snapshot: verifiertasks.Snapshot{Version: 1, Change: "demo"},
		Verdict:  "pass",
	}
	input, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("logical conflict is structured success", func(t *testing.T) {
		var output bytes.Buffer
		err := runCLI([]string{"verifier-tasks", "complete", "--change", "demo"}, cliDependencies{
			stdin:            bytes.NewReader(input),
			stdout:           &output,
			workingDirectory: func() (string, error) { return "/repo", nil },
			completeVerifierTasks: func(context.Context, verifiertasks.ResolveRequest, verifiertasks.CompleteRequest) (verifiertasks.Result, error) {
				return verifiertasks.Result{
					Status: "blocked", Verdict: "pass", Completion: "conflict",
					Evidence: []verifiertasks.TaskEvidence{}, Commands: []verifiertasks.CommandRecord{},
					Conflicts: []verifiertasks.Conflict{{Code: "content_changed", Detail: "tasks.md changed"}},
				}, nil
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		var result verifiertasks.Result
		if err := json.Unmarshal(output.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if result.Status != "blocked" || result.Completion != "conflict" || len(result.Conflicts) != 1 {
			t.Fatalf("result = %+v", result)
		}
	})

	t.Run("infrastructure failure returns an error after structured output", func(t *testing.T) {
		var output bytes.Buffer
		infrastructureErr := errors.New("atomic rename failed")
		err := runCLI([]string{"verifier-tasks", "complete", "--change", "demo"}, cliDependencies{
			stdin:            bytes.NewReader(input),
			stdout:           &output,
			workingDirectory: func() (string, error) { return "/repo", nil },
			completeVerifierTasks: func(context.Context, verifiertasks.ResolveRequest, verifiertasks.CompleteRequest) (verifiertasks.Result, error) {
				return verifiertasks.Result{
					Status: "blocked", Verdict: "pass", Completion: "conflict",
					Evidence: []verifiertasks.TaskEvidence{}, Commands: []verifiertasks.CommandRecord{},
					Conflicts: []verifiertasks.Conflict{{Code: "atomic_commit_failed", Detail: infrastructureErr.Error()}},
				}, infrastructureErr
			},
		})
		if !errors.Is(err, infrastructureErr) {
			t.Fatalf("error = %v, want infrastructure failure", err)
		}
		if output.Len() == 0 {
			t.Fatal("infrastructure failure omitted structured result")
		}
	})
}

func TestRunCLIVerifierTasksRejectsUnstructuredInputs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		body string
		want string
	}{
		{name: "unknown phase", args: []string{"verifier-tasks", "edit", "--change", "demo"}, want: "unknown phase"},
		{name: "missing change", args: []string{"verifier-tasks", "snapshot"}, want: "--change is required"},
		{name: "unknown JSON field", args: []string{"verifier-tasks", "complete", "--change", "demo"}, body: `{"edit":"anything"}`, want: "unknown field"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := runCLI(test.args, cliDependencies{
				stdin:            strings.NewReader(test.body),
				stdout:           &bytes.Buffer{},
				workingDirectory: func() (string, error) { return "/repo", nil },
				completeVerifierTasks: func(context.Context, verifiertasks.ResolveRequest, verifiertasks.CompleteRequest) (verifiertasks.Result, error) {
					return verifiertasks.Result{}, nil
				},
			})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want containing %q", err, test.want)
			}
		})
	}
}

func TestRunCLIVersionIsOffline(t *testing.T) {
	useVersion(t, "v0.1.0")
	var stdout bytes.Buffer

	err := runCLI([]string{"version"}, cliDependencies{
		stdout: &stdout,
		runInstaller: func(rootOptions) error {
			t.Fatal("version invoked the installer")
			return nil
		},
		newUpdatePolicy: func() updatePolicy {
			t.Fatal("version constructed update networking")
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if stdout.String() != "v0.1.0\n" {
		t.Fatalf("version output = %q", stdout.String())
	}
}

func TestRunCLIUpdateForcesPolicy(t *testing.T) {
	useVersion(t, "v0.1.0")
	policy := &recordedUpdatePolicy{}

	err := runCLI([]string{"update"}, cliDependencies{
		stdout: &bytes.Buffer{},
		runInstaller: func(rootOptions) error {
			t.Fatal("update invoked the installer")
			return nil
		},
		newUpdatePolicy: func() updatePolicy { return policy },
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(policy.calls, []updatePolicyCall{{version: "v0.1.0", forced: true}}) {
		t.Fatalf("update calls = %#v", policy.calls)
	}
}

func TestRunCLIDevSuppressesAutomaticAndForcedUpdates(t *testing.T) {
	useVersion(t, "dev")
	tests := []struct {
		name       string
		args       []string
		wantOutput string
		wantTUI    bool
	}{
		{name: "automatic", wantTUI: true},
		{name: "forced", args: []string{"update"}, wantOutput: "self-update is disabled for dev builds\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout bytes.Buffer
			installerCalls := 0
			err := runCLI(test.args, cliDependencies{
				stdout: &stdout,
				runInstaller: func(rootOptions) error {
					installerCalls++
					return nil
				},
				newUpdatePolicy: func() updatePolicy {
					t.Fatal("dev build constructed update networking")
					return &recordedUpdatePolicy{err: errors.New("unreachable")}
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			wantInstallerCalls := 0
			if test.wantTUI {
				wantInstallerCalls = 1
			}
			if installerCalls != wantInstallerCalls {
				t.Fatalf("installer calls = %d, want %d", installerCalls, wantInstallerCalls)
			}
			if stdout.String() != test.wantOutput {
				t.Fatalf("output = %q, want %q", stdout.String(), test.wantOutput)
			}
		})
	}
}

func TestEmbeddedAndDirectoryAssetSourcesHaveParity(t *testing.T) {
	embedded, err := sourceForAssets("")
	if err != nil {
		t.Fatal(err)
	}
	directory := assetfs.Directory("assets")

	embeddedFiles := assetFiles(t, embedded)
	directoryFiles := assetFiles(t, directory)
	if !reflect.DeepEqual(embeddedFiles, directoryFiles) {
		t.Fatal("embedded assets differ from the repository assets directory")
	}
	for name := range embeddedFiles {
		if strings.HasPrefix(name, "skills/openspec/") || strings.HasPrefix(name, "skills/openspec-") {
			t.Fatalf("embedded assets include vendored OpenSpec skill %q", name)
		}
	}
}

func TestDefaultInvocationUsesEmbeddedAssetsOutsideRepository(t *testing.T) {
	t.Chdir(t.TempDir())

	previousExtras := install.ExtraOptions
	install.ExtraOptions = nil
	t.Cleanup(func() { install.ExtraOptions = previousExtras })

	if err := run("", t.TempDir(), true, true); err != nil {
		t.Fatalf("installed-style invocation failed outside the repository: %v", err)
	}
}

func assetFiles(t *testing.T, source assetfs.Source) map[string]string {
	t.Helper()
	files := make(map[string]string)
	err := fs.WalkDir(source.FS(), ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		content, err := source.ReadFile(name)
		if err != nil {
			return err
		}
		files[name] = string(content)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}
