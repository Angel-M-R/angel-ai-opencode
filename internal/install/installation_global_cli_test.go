package install

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	assetfs "angel-ai-opencode/internal/assets"
	"angel-ai-opencode/internal/catalog"
)

func useGlobalCLICommands(t *testing.T, commands globalCLICommands) {
	t.Helper()
	previous := systemGlobalCLICommands
	systemGlobalCLICommands = commands
	t.Cleanup(func() { systemGlobalCLICommands = previous })
}

func countApplyPreparations(t *testing.T) *int {
	t.Helper()
	previous := prepareInstallationForApply
	count := 0
	prepareInstallationForApply = func(request InstallationRequest) (preparedInstallation, error) {
		count++
		return prepareInstallation(request)
	}
	t.Cleanup(func() { prepareInstallationForApply = previous })
	return &count
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func testCodegraphAssets(t *testing.T) string {
	t.Helper()
	assets := t.TempDir()
	writeTestFile(t, filepath.Join(assets, "integrations", "codegraph", "mcp.json"), `{
  "mcp": {
    "codegraph": {
      "command": ["codegraph", "serve", "--mcp"],
      "enabled": true,
      "type": "local"
    }
  }
}`)
	writeTestFile(t, filepath.Join(assets, "integrations", "codegraph", "AGENTS.md"), `<!-- codegraph-guidance -->
## CodeGraph
<!-- /codegraph-guidance -->
`)
	return assets
}

func hasLineContaining(lines []string, value string) bool {
	for _, line := range lines {
		if strings.Contains(line, value) {
			return true
		}
	}
	return false
}

type injectedCLIEnvironment struct {
	t                *testing.T
	manager          string
	nodeVersion      string
	pnpmGlobalBin    []byte
	localVersions    map[string]string
	registrations    map[string]bool
	latestVersions   map[string]string
	registryOutput   map[string][]byte
	registryErrors   map[string]error
	versionOutput    map[string][]byte
	versionErrors    map[string]error
	installationErrs map[string]error
	events           []string
	installations    []string
}

func newInjectedCLIEnvironment(t *testing.T, manager string) *injectedCLIEnvironment {
	t.Helper()
	return &injectedCLIEnvironment{
		t:                t,
		manager:          manager,
		nodeVersion:      "v22.0.0\n",
		pnpmGlobalBin:    []byte("/global/pnpm/bin\n"),
		localVersions:    map[string]string{},
		registrations:    map[string]bool{},
		latestVersions:   map[string]string{},
		registryOutput:   map[string][]byte{},
		registryErrors:   map[string]error{},
		versionOutput:    map[string][]byte{},
		versionErrors:    map[string]error{},
		installationErrs: map[string]error{},
	}
}

func (environment *injectedCLIEnvironment) commands() globalCLICommands {
	return globalCLICommands{
		lookPath: func(name string) (string, error) {
			environment.events = append(environment.events, "look "+name)
			switch name {
			case "npm":
				if environment.manager == "npm" {
					return "/tools/npm", nil
				}
			case "pnpm":
				return "/tools/pnpm", nil
			case "node":
				return "/tools/node", nil
			case "codegraph", "openspec":
				if _, ok := environment.localVersions[name]; ok {
					return "/tools/" + name, nil
				}
			}
			return "", errors.New("not found")
		},
		run: func(path string, args ...string) ([]byte, error) {
			environment.events = append(environment.events, "run "+strings.Join(append([]string{path}, args...), " "))
			if path == "/tools/node" {
				return []byte(environment.nodeVersion), nil
			}
			if path == "/tools/pnpm" && reflect.DeepEqual(args, []string{"bin", "-g"}) {
				return environment.pnpmGlobalBin, nil
			}
			if path == "/tools/npm" || path == "/tools/pnpm" {
				return environment.runManagerCommand(path, args)
			}
			for _, descriptor := range globalCLIDescriptors {
				if path == "/tools/"+descriptor.executable && reflect.DeepEqual(args, descriptor.versionArgs) {
					if err := environment.versionErrors[descriptor.executable]; err != nil {
						return environment.versionOutput[descriptor.executable], err
					}
					if output, ok := environment.versionOutput[descriptor.executable]; ok {
						return output, nil
					}
					return []byte(environment.localVersions[descriptor.executable] + "\n"), nil
				}
			}
			environment.t.Fatalf("unexpected injected command %q %v", path, args)
			return nil, nil
		},
	}
}

func (environment *injectedCLIEnvironment) runManagerCommand(path string, args []string) ([]byte, error) {
	environment.t.Helper()
	if len(args) == 0 {
		environment.t.Fatalf("manager command has no arguments: %q", path)
	}
	switch args[0] {
	case "list":
		packageName := args[len(args)-1]
		return packageRegistrationJSON(environment.t, environment.manager, packageName, environment.registrations[packageName]), nil
	case "view":
		packageName := strings.TrimSuffix(args[1], "@latest")
		if err := environment.registryErrors[packageName]; err != nil {
			return environment.registryOutput[packageName], err
		}
		if output, ok := environment.registryOutput[packageName]; ok {
			return output, nil
		}
		output, err := json.Marshal(environment.latestVersions[packageName])
		if err != nil {
			environment.t.Fatal(err)
		}
		return output, nil
	case "install", "add":
		installSpec := args[len(args)-1]
		environment.installations = append(environment.installations, installSpec)
		if err := environment.installationErrs[installSpec]; err != nil {
			return []byte("injected install failure\n"), err
		}
		descriptor := descriptorForInstallSpec(environment.t, installSpec)
		environment.localVersions[descriptor.executable] = environment.latestVersions[descriptor.registryPackage]
		return nil, nil
	default:
		environment.t.Fatalf("unexpected injected manager command %q %v", path, args)
		return nil, nil
	}
}

func packageRegistrationJSON(t *testing.T, manager, packageName string, registered bool) []byte {
	t.Helper()
	dependencies := map[string]any{}
	if registered {
		dependencies[packageName] = map[string]any{"version": "1.0.0"}
	}
	var value any = map[string]any{"dependencies": dependencies}
	if manager == "pnpm" {
		value = []any{value}
	}
	output, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return output
}

func descriptorForInstallSpec(t *testing.T, installSpec string) globalCLIDescriptor {
	t.Helper()
	for _, descriptor := range globalCLIDescriptors {
		if descriptor.installSpec == installSpec {
			return descriptor
		}
	}
	t.Fatalf("unknown install spec %q", installSpec)
	return globalCLIDescriptor{}
}

func installationRequestForDescriptors(
	t *testing.T,
	target string,
	withManagedFile bool,
	descriptors ...globalCLIDescriptor,
) InstallationRequest {
	t.Helper()
	extras := make(map[string]bool, len(descriptors))
	for _, descriptor := range descriptors {
		extras[descriptor.optionKey] = true
	}
	assetsDir := testCodegraphAssets(t)
	request := InstallationRequest{
		Extras:    extras,
		Assets:    assetfs.Directory(assetsDir),
		ConfigDir: target,
	}
	if withManagedFile {
		source := filepath.Join(assetsDir, "managed.md")
		writeTestFile(t, source, "managed\n")
		request.Items = []catalog.Item{{
			Name: "managed", Source: "managed.md", Dest: "managed.md", Kind: catalog.CopyFile,
		}}
	}
	return request
}

func assertNoConfigurationWrites(t *testing.T, target string) {
	t.Helper()
	entries, err := os.ReadDir(target)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("configuration was written before CLI success: %v", entries)
	}
}

func assertNoCleanupOrInstall(t *testing.T, environment *injectedCLIEnvironment) {
	t.Helper()
	if len(environment.installations) != 0 {
		t.Fatalf("unexpected package installations: %v", environment.installations)
	}
	for _, event := range environment.events {
		for _, cleanup := range []string{" uninstall ", " remove ", " unlink "} {
			if strings.Contains(" "+event+" ", cleanup) {
				t.Fatalf("unexpected cleanup command: %s", event)
			}
		}
	}
}

func indexOfEventContaining(events []string, value string) int {
	for index, event := range events {
		if strings.Contains(event, value) {
			return index
		}
	}
	return -1
}

func TestInjectedGlobalCLIManagerSelectionAndProbes(t *testing.T) {
	t.Run("npm preferred", func(t *testing.T) {
		environment := newInjectedCLIEnvironment(t, "npm")
		environment.latestVersions[codegraphRegistryPackage] = "2.0.0"
		useGlobalCLICommands(t, environment.commands())

		plan, err := PlanInstallation(installationRequestForDescriptors(
			t, t.TempDir(), false, globalCLIDescriptors[0],
		))
		if err != nil {
			t.Fatal(err)
		}
		if !hasLineContaining(plan, codegraphPackage) {
			t.Fatalf("CodeGraph install action missing from plan: %v", plan)
		}
		if !reflect.DeepEqual(environment.events[:1], []string{"look npm"}) {
			t.Fatalf("manager selection events = %v", environment.events)
		}
		for _, command := range []string{
			"run /tools/npm list --global --depth=0 --json " + codegraphRegistryPackage,
			"run /tools/npm view " + codegraphPackage + " version --json",
		} {
			if indexOfEventContaining(environment.events, command) < 0 {
				t.Errorf("selected npm probe missing: %s; events=%v", command, environment.events)
			}
		}
		if indexOfEventContaining(environment.events, "/tools/pnpm") >= 0 {
			t.Fatalf("unselected pnpm was queried: %v", environment.events)
		}
	})

	t.Run("validated pnpm fallback", func(t *testing.T) {
		environment := newInjectedCLIEnvironment(t, "pnpm")
		environment.latestVersions[codegraphRegistryPackage] = "2.0.0"
		useGlobalCLICommands(t, environment.commands())

		if _, err := PlanInstallation(installationRequestForDescriptors(
			t, t.TempDir(), false, globalCLIDescriptors[0],
		)); err != nil {
			t.Fatal(err)
		}
		binIndex := indexOfEventContaining(environment.events, "run /tools/pnpm bin -g")
		listIndex := indexOfEventContaining(environment.events, "run /tools/pnpm list --global --depth=0 --json "+codegraphRegistryPackage)
		viewIndex := indexOfEventContaining(environment.events, "run /tools/pnpm view "+codegraphPackage+" version --json")
		if binIndex < 0 || listIndex <= binIndex || viewIndex <= listIndex {
			t.Fatalf("pnpm validation/probe order = %v", environment.events)
		}
		if indexOfEventContaining(environment.events, "run /tools/npm") >= 0 {
			t.Fatalf("unselected npm was queried after fallback: %v", environment.events)
		}
	})

	for _, test := range []struct {
		name   string
		output []byte
		want   string
	}{
		{name: "empty pnpm global bin", output: []byte(" \n"), want: "empty path"},
		{name: "failed pnpm global bin", output: []byte("PNPM_HOME is unset\n"), want: "PNPM_HOME is unset"},
	} {
		t.Run(test.name, func(t *testing.T) {
			environment := newInjectedCLIEnvironment(t, "pnpm")
			environment.pnpmGlobalBin = test.output
			commands := environment.commands()
			if strings.HasPrefix(test.name, "failed") {
				baseRun := commands.run
				commands.run = func(path string, args ...string) ([]byte, error) {
					output, err := baseRun(path, args...)
					if path == "/tools/pnpm" && reflect.DeepEqual(args, []string{"bin", "-g"}) {
						return output, errors.New("exit 1")
					}
					return output, err
				}
			}
			useGlobalCLICommands(t, commands)
			_, err := PlanInstallation(installationRequestForDescriptors(
				t, t.TempDir(), false, globalCLIDescriptors[0],
			))
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("expected pnpm validation error containing %q, got %v", test.want, err)
			}
			if indexOfEventContaining(environment.events, " list ") >= 0 ||
				indexOfEventContaining(environment.events, " view ") >= 0 {
				t.Fatalf("package probes ran before pnpm validation: %v", environment.events)
			}
		})
	}
}

func TestInjectedGlobalCLIClassificationsInPlanAndApply(t *testing.T) {
	classifications := []struct {
		name       string
		installed  string
		latest     string
		registered bool
		status     string
	}{
		{name: "current", installed: "2.0.0", latest: "2.0.0", registered: true, status: "current"},
		{name: "outdated", installed: "1.0.0", latest: "2.0.0", registered: true, status: "outdated"},
		{name: "ahead", installed: "3.0.0", latest: "2.0.0", registered: true, status: "ahead"},
		{name: "absent-installable", latest: "2.0.0", status: "install"},
		{name: "working-unregistered", installed: "2.0.0", latest: "2.0.0", status: "current"},
	}

	for _, descriptor := range globalCLIDescriptors {
		for _, classification := range classifications {
			t.Run(descriptor.displayName+"/"+classification.name, func(t *testing.T) {
				setup := func(t *testing.T) *injectedCLIEnvironment {
					environment := newInjectedCLIEnvironment(t, "npm")
					environment.registrations[descriptor.registryPackage] = classification.registered
					environment.latestVersions[descriptor.registryPackage] = classification.latest
					if classification.installed != "" {
						environment.localVersions[descriptor.executable] = classification.installed
					}
					return environment
				}

				t.Run("plan", func(t *testing.T) {
					environment := setup(t)
					target := t.TempDir()
					useGlobalCLICommands(t, environment.commands())
					plan, err := PlanInstallation(installationRequestForDescriptors(t, target, false, descriptor))
					if err != nil {
						t.Fatal(err)
					}
					want := descriptor.displayName + ": " + classification.status
					if classification.status == "install" {
						want = "INSTALAR   " + descriptor.installSpec
					}
					if !hasLineContaining(plan, want) {
						t.Fatalf("classification %q missing from plan: %v", want, plan)
					}
					assertNoCleanupOrInstall(t, environment)
					assertNoConfigurationWrites(t, target)
				})

				t.Run("apply", func(t *testing.T) {
					environment := setup(t)
					target := t.TempDir()
					useGlobalCLICommands(t, environment.commands())
					report, err := ApplyInstallation(installationRequestForDescriptors(t, target, false, descriptor))
					if err != nil {
						t.Fatal(err)
					}
					want := descriptor.displayName + ": " + classification.status
					if classification.status == "install" {
						want = "instalado  " + descriptor.installSpec + " (version " + classification.latest + ")"
						if !reflect.DeepEqual(environment.installations, []string{descriptor.installSpec}) {
							t.Fatalf("installations = %v", environment.installations)
						}
					} else {
						assertNoCleanupOrInstall(t, environment)
					}
					if !hasLineContaining(report, want) {
						t.Fatalf("classification %q missing from final report: %v", want, report)
					}
				})
			})
		}
	}
}

func TestInjectedGlobalCLIRegistryFailures(t *testing.T) {
	healthyCases := []struct {
		name       string
		descriptor globalCLIDescriptor
		configure  func(*injectedCLIEnvironment)
		wantDetail string
	}{
		{
			name: "registry command failure", descriptor: globalCLIDescriptors[0], wantDetail: "registry offline",
			configure: func(environment *injectedCLIEnvironment) {
				environment.registryOutput[codegraphRegistryPackage] = []byte("registry offline\n")
				environment.registryErrors[codegraphRegistryPackage] = errors.New("exit 1")
			},
		},
		{
			name: "malformed registry output", descriptor: openSpecGlobalCLI, wantDetail: "invalid registry latest JSON",
			configure: func(environment *injectedCLIEnvironment) {
				environment.registryOutput[openSpecRegistryPackage] = []byte("not-json")
			},
		},
	}
	for _, test := range healthyCases {
		t.Run("healthy local/"+test.name, func(t *testing.T) {
			for _, action := range []string{"plan", "apply"} {
				t.Run(action, func(t *testing.T) {
					environment := newInjectedCLIEnvironment(t, "npm")
					environment.localVersions[test.descriptor.executable] = "1.5.0"
					environment.registrations[test.descriptor.registryPackage] = true
					test.configure(environment)
					target := t.TempDir()
					useGlobalCLICommands(t, environment.commands())
					request := installationRequestForDescriptors(t, target, false, test.descriptor)
					var output []string
					var err error
					if action == "plan" {
						output, err = PlanInstallation(request)
					} else {
						output, err = ApplyInstallation(request)
					}
					if err != nil {
						t.Fatal(err)
					}
					for _, want := range []string{"registry-unverified", "WARNING:", test.wantDetail} {
						if !hasLineContaining(output, want) {
							t.Errorf("%q missing from %s output: %v", want, action, output)
						}
					}
					assertNoCleanupOrInstall(t, environment)
				})
			}
		})
	}

	for _, test := range []struct {
		name      string
		configure func(*injectedCLIEnvironment)
		wantState string
	}{
		{
			name: "registry command failure",
			configure: func(environment *injectedCLIEnvironment) {
				environment.registryErrors[openSpecRegistryPackage] = errors.New("registry offline")
			},
			wantState: "unavailable",
		},
		{
			name: "malformed registry output",
			configure: func(environment *injectedCLIEnvironment) {
				environment.registryOutput[openSpecRegistryPackage] = []byte(`{"version": "2.0.0"}`)
			},
			wantState: "malformed",
		},
	} {
		t.Run("absent CLI/"+test.name, func(t *testing.T) {
			for _, action := range []string{"plan", "apply"} {
				t.Run(action, func(t *testing.T) {
					environment := newInjectedCLIEnvironment(t, "npm")
					environment.latestVersions[codegraphRegistryPackage] = "2.0.0"
					test.configure(environment)
					target := t.TempDir()
					useGlobalCLICommands(t, environment.commands())
					request := installationRequestForDescriptors(
						t, target, true, globalCLIDescriptors[0], openSpecGlobalCLI,
					)
					var err error
					if action == "plan" {
						_, err = PlanInstallation(request)
					} else {
						_, err = ApplyInstallation(request)
					}
					if err == nil || !strings.Contains(err.Error(), "cannot install OpenSpec with npm") ||
						!strings.Contains(err.Error(), test.wantState) ||
						!strings.Contains(err.Error(), "no package changes were performed") {
						t.Fatalf("expected guided absent-CLI registry error, got %v", err)
					}
					assertNoCleanupOrInstall(t, environment)
					assertNoConfigurationWrites(t, target)
				})
			}
		})
	}
}

func TestInjectedGlobalCLIBrokenInstallations(t *testing.T) {
	tests := []struct {
		name       string
		descriptor globalCLIDescriptor
		configure  func(*injectedCLIEnvironment)
		want       []string
	}{
		{
			name: "registered package missing executable", descriptor: globalCLIDescriptors[0],
			configure: func(environment *injectedCLIEnvironment) {
				environment.registrations[codegraphRegistryPackage] = true
			},
			want: []string{"npm registers " + codegraphRegistryPackage, "CodeGraph", "repair the npm global registration", "no package cleanup was performed"},
		},
		{
			name: "version command failure", descriptor: openSpecGlobalCLI,
			configure: func(environment *injectedCLIEnvironment) {
				environment.localVersions["openspec"] = "1.0.0"
				environment.registrations[openSpecRegistryPackage] = true
				environment.versionOutput["openspec"] = []byte("broken executable\n")
				environment.versionErrors["openspec"] = errors.New("exit 1")
			},
			want: []string{"OpenSpec executable openspec", "with npm", "version command failed", "repair the OpenSpec executable", "no package cleanup was performed"},
		},
		{
			name: "empty version output", descriptor: globalCLIDescriptors[0],
			configure: func(environment *injectedCLIEnvironment) {
				environment.localVersions["codegraph"] = "1.0.0"
				environment.versionOutput["codegraph"] = []byte(" \n")
			},
			want: []string{"CodeGraph executable codegraph", "version command returned no version", "repair the CodeGraph executable", "no package cleanup was performed"},
		},
		{
			name: "malformed version output", descriptor: openSpecGlobalCLI,
			configure: func(environment *injectedCLIEnvironment) {
				environment.localVersions["openspec"] = "1.0.0"
				environment.versionOutput["openspec"] = []byte("OpenSpec development build\n")
			},
			want: []string{"OpenSpec executable openspec", "uninterpretable output", "repair the OpenSpec executable", "no package cleanup was performed"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, action := range []string{"plan", "apply"} {
				t.Run(action, func(t *testing.T) {
					environment := newInjectedCLIEnvironment(t, "npm")
					test.configure(environment)
					target := t.TempDir()
					useGlobalCLICommands(t, environment.commands())
					request := installationRequestForDescriptors(t, target, true, test.descriptor)
					var err error
					if action == "plan" {
						_, err = PlanInstallation(request)
					} else {
						_, err = ApplyInstallation(request)
					}
					if err == nil {
						t.Fatal("expected guided CLI recovery error")
					}
					for _, want := range test.want {
						if !strings.Contains(err.Error(), want) {
							t.Errorf("%q missing from error: %v", want, err)
						}
					}
					assertNoCleanupOrInstall(t, environment)
					assertNoConfigurationWrites(t, target)
				})
			}
		})
	}
}

func TestInjectedGlobalCLIMultiCLIPrevalidationAndWriteBoundaries(t *testing.T) {
	setup := func(t *testing.T) *injectedCLIEnvironment {
		environment := newInjectedCLIEnvironment(t, "npm")
		environment.latestVersions[codegraphRegistryPackage] = "2.0.0"
		environment.latestVersions[openSpecRegistryPackage] = "3.0.0"
		return environment
	}

	t.Run("success", func(t *testing.T) {
		environment := setup(t)
		preparations := countApplyPreparations(t)
		target := t.TempDir()
		useGlobalCLICommands(t, environment.commands())
		report, err := ApplyInstallation(installationRequestForDescriptors(
			t, target, true, globalCLIDescriptors[0], openSpecGlobalCLI,
		))
		if err != nil {
			t.Fatal(err)
		}
		firstInstall := indexOfEventContaining(environment.events, "run /tools/npm install --global "+codegraphPackage)
		openSpecPrevalidation := indexOfEventContaining(environment.events, "run /tools/npm view "+openSpecPackage+" version --json")
		nodePreflight := indexOfEventContaining(environment.events, "run /tools/node --version")
		firstInspection := indexOfEventContaining(environment.events, "run /tools/npm list --global --depth=0 --json "+codegraphRegistryPackage)
		if nodePreflight < 0 || firstInspection <= nodePreflight || openSpecPrevalidation <= firstInspection || firstInstall <= openSpecPrevalidation {
			t.Fatalf("Node/complete-prevalidation/install order = %v", environment.events)
		}
		if !reflect.DeepEqual(environment.installations, []string{codegraphPackage, openSpecPackage}) {
			t.Fatalf("deterministic installation order = %v", environment.installations)
		}
		if *preparations != 2 {
			t.Fatalf("preparation count = %d, want initial preparation plus one repreparation", *preparations)
		}
		if len(report) < 2 || !strings.Contains(report[0], codegraphPackage) || !strings.Contains(report[1], openSpecPackage) {
			t.Fatalf("CLI report order = %v", report)
		}
		for _, path := range []string{"managed.md", "opencode.json", "AGENTS.md"} {
			if _, err := os.Stat(filepath.Join(target, path)); err != nil {
				t.Fatalf("configuration %s was not written after all CLI actions succeeded: %v", path, err)
			}
		}
	})

	t.Run("later installation failure", func(t *testing.T) {
		environment := setup(t)
		environment.installationErrs[openSpecPackage] = errors.New("exit 1")
		preparations := countApplyPreparations(t)
		target := t.TempDir()
		useGlobalCLICommands(t, environment.commands())
		report, err := ApplyInstallation(installationRequestForDescriptors(
			t, target, true, globalCLIDescriptors[0], openSpecGlobalCLI,
		))
		if err == nil || !strings.Contains(err.Error(), "installing OpenSpec with npm") {
			t.Fatalf("expected later installation failure, got %v", err)
		}
		firstInstall := indexOfEventContaining(environment.events, "run /tools/npm install --global "+codegraphPackage)
		openSpecPrevalidation := indexOfEventContaining(environment.events, "run /tools/npm view "+openSpecPackage+" version --json")
		if firstInstall <= openSpecPrevalidation {
			t.Fatalf("installation began before complete prevalidation: %v", environment.events)
		}
		if !reflect.DeepEqual(environment.installations, []string{codegraphPackage, openSpecPackage}) {
			t.Fatalf("installation attempt order = %v", environment.installations)
		}
		if len(report) != 1 || !strings.Contains(report[0], codegraphPackage) {
			t.Fatalf("partial CLI report = %v", report)
		}
		if *preparations != 1 {
			t.Fatalf("preparation count after CLI failure = %d, want 1", *preparations)
		}
		assertNoConfigurationWrites(t, target)
	})
}

func TestApplyInstallationRejectsInvalidPostInstallVersionWithoutConfigurationWrites(t *testing.T) {
	environment := newInjectedCLIEnvironment(t, "npm")
	environment.latestVersions[codegraphRegistryPackage] = "2.0.0"
	environment.versionOutput["codegraph"] = []byte("CodeGraph development build\n")
	target := t.TempDir()
	useGlobalCLICommands(t, environment.commands())

	report, err := ApplyInstallation(installationRequestForDescriptors(
		t, target, true, globalCLIDescriptors[0],
	))
	if err == nil || !strings.Contains(err.Error(), "uninterpretable output") {
		t.Fatalf("expected invalid post-install version error, got report=%v err=%v", report, err)
	}
	if !reflect.DeepEqual(environment.installations, []string{codegraphPackage}) {
		t.Fatalf("package installation did not complete before version verification: %v", environment.installations)
	}
	assertNoConfigurationWrites(t, target)
}

func TestOpenSpecSelectionIsCLIOnlyAndPreservesVendoredSkills(t *testing.T) {
	foundExtra := false
	for _, extra := range ExtraOptions {
		if extra.Key == openSpecOptionKey && extra.Label == "OpenSpec" {
			foundExtra = true
			break
		}
	}
	if !foundExtra {
		t.Fatal("OpenSpec installer extra is missing")
	}

	assets := filepath.Join("..", "..", "assets")
	assetSource := assetfs.Directory(assets)
	categories, err := catalog.Load(assetSource)
	if err != nil {
		t.Fatal(err)
	}
	var openSpecSkills []string
	for _, category := range categories {
		if category.Name != "skills" {
			continue
		}
		for _, item := range category.Items {
			if item.Name != "openspec" {
				continue
			}
			entries, err := assetSource.ReadDir(item.Source)
			if err != nil {
				t.Fatal(err)
			}
			for _, entry := range entries {
				if entry.IsDir() && strings.HasPrefix(entry.Name(), "openspec-") {
					openSpecSkills = append(openSpecSkills, entry.Name())
				}
			}
		}
	}
	wantSkills := []string{
		"openspec-apply-change",
		"openspec-archive-change",
		"openspec-bulk-archive-change",
		"openspec-continue-change",
		"openspec-explore",
		"openspec-ff-change",
		"openspec-new-change",
		"openspec-onboard",
		"openspec-propose",
		"openspec-sync-specs",
		"openspec-update-change",
		"openspec-verify-change",
	}
	if !reflect.DeepEqual(openSpecSkills, wantSkills) {
		t.Fatalf("vendored OpenSpec skills = %v, want %v", openSpecSkills, wantSkills)
	}

	target := t.TempDir()
	opencodePath := filepath.Join(target, "opencode.json")
	agentsPath := filepath.Join(target, "AGENTS.md")
	wantConfig := "{\"mcp\":{\"context7\":{\"type\":\"remote\"}}}\n"
	wantAgents := "# Existing rules\n"
	writeTestFile(t, opencodePath, wantConfig)
	writeTestFile(t, agentsPath, wantAgents)
	environment := newInjectedCLIEnvironment(t, "npm")
	environment.localVersions["openspec"] = "1.0.0"
	environment.registrations[openSpecRegistryPackage] = true
	environment.latestVersions[openSpecRegistryPackage] = "1.0.0"
	useGlobalCLICommands(t, environment.commands())

	if _, err := ApplyInstallation(InstallationRequest{
		Extras:    map[string]bool{openSpecOptionKey: true},
		Assets:    assetSource,
		ConfigDir: target,
	}); err != nil {
		t.Fatal(err)
	}
	assertNoCleanupOrInstall(t, environment)
	if got, err := os.ReadFile(opencodePath); err != nil || string(got) != wantConfig {
		t.Fatalf("OpenSpec selection changed opencode.json: content=%q err=%v", got, err)
	}
	if got, err := os.ReadFile(agentsPath); err != nil || string(got) != wantAgents {
		t.Fatalf("OpenSpec selection changed AGENTS.md: content=%q err=%v", got, err)
	}
	entries, err := os.ReadDir(target)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("OpenSpec CLI-only selection created configuration artifacts: %v", entries)
	}
}
