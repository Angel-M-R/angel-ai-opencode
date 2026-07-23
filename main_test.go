package main

import (
	"bytes"
	"errors"
	"io/fs"
	"reflect"
	"strings"
	"testing"

	assetfs "angel-ai-opencode/internal/assets"
	"angel-ai-opencode/internal/install"
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
	if _, ok := embeddedFiles["skills/openspec/openspec-apply-change/SKILL.md"]; !ok {
		t.Fatal("embedded assets are missing nested OpenSpec skill content")
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
