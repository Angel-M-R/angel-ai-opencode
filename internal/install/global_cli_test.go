package install

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestSelectGlobalPackageManagerPrefersNPM(t *testing.T) {
	var lookups []string
	commands := globalCLICommands{
		lookPath: func(name string) (string, error) {
			lookups = append(lookups, name)
			return "/tools/" + name, nil
		},
		run: func(string, ...string) ([]byte, error) {
			t.Fatal("npm selection must not invoke pnpm")
			return nil, nil
		},
	}

	manager, err := selectGlobalPackageManager(commands)
	if err != nil {
		t.Fatal(err)
	}
	if manager.name != "npm" || manager.executable != "/tools/npm" {
		t.Fatalf("selected manager = %+v", manager)
	}
	if !reflect.DeepEqual(lookups, []string{"npm"}) {
		t.Fatalf("executable lookups = %v, want only npm", lookups)
	}
}

func TestSelectGlobalPackageManagerValidatesPNPMGlobalBin(t *testing.T) {
	var gotPath string
	var gotArgs []string
	commands := globalCLICommands{
		lookPath: func(name string) (string, error) {
			if name == "pnpm" {
				return "/tools/pnpm", nil
			}
			return "", errors.New("not found")
		},
		run: func(path string, args ...string) ([]byte, error) {
			gotPath = path
			gotArgs = append([]string{}, args...)
			return []byte(" /global/pnpm/bin\n"), nil
		},
	}

	manager, err := selectGlobalPackageManager(commands)
	if err != nil {
		t.Fatal(err)
	}
	if manager.name != "pnpm" || manager.globalBin != "/global/pnpm/bin" {
		t.Fatalf("selected manager = %+v", manager)
	}
	if gotPath != "/tools/pnpm" || !reflect.DeepEqual(gotArgs, []string{"bin", "-g"}) {
		t.Fatalf("pnpm validation command = %q %v", gotPath, gotArgs)
	}
}

func TestSelectGlobalPackageManagerRejectsInvalidPNPMGlobalBin(t *testing.T) {
	tests := []struct {
		name   string
		output []byte
		err    error
		want   string
	}{
		{name: "empty", output: []byte(" \n"), want: "returned an empty path"},
		{name: "command failure", output: []byte("PNPM_HOME is unset\n"), err: errors.New("exit 1"), want: "PNPM_HOME is unset"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			commands := globalCLICommands{
				lookPath: func(name string) (string, error) {
					if name == "pnpm" {
						return "/tools/pnpm", nil
					}
					return "", errors.New("not found")
				},
				run: func(string, ...string) ([]byte, error) { return test.output, test.err },
			}

			_, err := selectGlobalPackageManager(commands)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("expected error containing %q, got %v", test.want, err)
			}
		})
	}
}

func TestInstallGlobalCLIDoesNotFallbackAfterNPMFailure(t *testing.T) {
	var lookups []string
	var runs int
	commands := globalCLICommands{
		lookPath: func(name string) (string, error) {
			lookups = append(lookups, name)
			if name == "npm" || name == "pnpm" {
				return "/tools/" + name, nil
			}
			return "", errors.New("not found")
		},
		run: func(path string, args ...string) ([]byte, error) {
			runs++
			return []byte("npm failed"), errors.New("exit 1")
		},
	}

	manager, err := selectGlobalPackageManager(commands)
	if err != nil {
		t.Fatal(err)
	}
	_, err = installGlobalCLI(globalCLIDescriptor{
		displayName: "Example CLI", executable: "example", installSpec: "@example/cli@latest",
	}, manager, commands)
	if err == nil || !strings.Contains(err.Error(), "installing Example CLI with npm") {
		t.Fatalf("expected npm install error, got %v", err)
	}
	if !reflect.DeepEqual(lookups, []string{"npm"}) {
		t.Fatalf("executable lookups = %v; pnpm fallback must not occur", lookups)
	}
	if runs != 1 {
		t.Fatalf("package command count = %d, want 1", runs)
	}
}

func TestValidateGlobalCLIRuntimesChecksOpenSpecNodeFloor(t *testing.T) {
	descriptor := globalCLIDescriptor{
		displayName:        "OpenSpec",
		executable:         "openspec",
		installSpec:        "@fission-ai/openspec@latest",
		minimumNodeVersion: openSpecMinimumNodeVersion,
	}
	tests := []struct {
		name       string
		version    string
		lookupErr  error
		commandErr error
		wantError  string
	}{
		{name: "minimum", version: "v20.19.0\n"},
		{name: "newer", version: "v22.1.0\n"},
		{name: "newer prerelease", version: "v20.20.0-rc.1\n"},
		{name: "older", version: "v20.18.1\n", wantError: "requires Node.js >=20.19.0"},
		{name: "minimum prerelease", version: "v20.19.0-rc.1\n", wantError: "found Node.js 20.19.0-rc.1"},
		{name: "missing executable", lookupErr: errors.New("not found"), wantError: "node is not available on PATH"},
		{name: "missing output", version: " \n", wantError: "returned no version"},
		{name: "malformed", version: "v20.19\n", wantError: "returned malformed version"},
		{name: "command failure", version: "version check failed\n", commandErr: errors.New("exit 1"), wantError: "checking Node.js version"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			commands := globalCLICommands{
				lookPath: func(name string) (string, error) {
					if name != "node" {
						t.Fatalf("unexpected executable lookup %q", name)
					}
					return "/tools/node", test.lookupErr
				},
				run: func(path string, args ...string) ([]byte, error) {
					if path != "/tools/node" || !reflect.DeepEqual(args, []string{"--version"}) {
						t.Fatalf("version command = %q %v", path, args)
					}
					return []byte(test.version), test.commandErr
				},
			}

			err := validateGlobalCLIRuntimes([]globalCLIDescriptor{descriptor}, commands)
			if test.wantError == "" && err != nil {
				t.Fatal(err)
			}
			if test.wantError != "" && (err == nil || !strings.Contains(err.Error(), test.wantError)) {
				t.Fatalf("expected error containing %q, got %v", test.wantError, err)
			}
		})
	}
}

func TestValidateGlobalCLIRuntimesSkipsNodeWithoutRequirement(t *testing.T) {
	commands := globalCLICommands{
		lookPath: func(name string) (string, error) {
			t.Fatalf("unexpected executable lookup %q", name)
			return "", nil
		},
		run: func(path string, args ...string) ([]byte, error) {
			t.Fatalf("unexpected command %q %v", path, args)
			return nil, nil
		},
	}

	if err := validateGlobalCLIRuntimes([]globalCLIDescriptor{{
		displayName: "CodeGraph", executable: "codegraph", installSpec: "@colbymchenry/codegraph@latest",
	}}, commands); err != nil {
		t.Fatal(err)
	}
}

func TestInstallGlobalCLIUsesInjectedCommandsAndVerifiesExecutable(t *testing.T) {
	descriptor := globalCLIDescriptor{
		displayName: "Example CLI",
		executable:  "example",
		installSpec: "@example/cli@latest",
	}
	manager := globalPackageManager{
		name:       "npm",
		executable: "/tools/npm",
		install:    []string{"install", "--global"},
	}
	installed := false
	var gotPath string
	var gotArgs []string
	commands := globalCLICommands{
		lookPath: func(name string) (string, error) {
			if name == descriptor.executable && installed {
				return "/tools/example", nil
			}
			return "", errors.New("not found")
		},
		run: func(path string, args ...string) ([]byte, error) {
			gotPath = path
			gotArgs = append([]string{}, args...)
			installed = true
			return nil, nil
		},
	}

	line, err := installGlobalCLI(descriptor, manager, commands)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != manager.executable {
		t.Fatalf("command path = %q, want %q", gotPath, manager.executable)
	}
	wantArgs := []string{"install", "--global", descriptor.installSpec}
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("command arguments = %v, want %v", gotArgs, wantArgs)
	}
	if line != "instalado  "+descriptor.installSpec {
		t.Fatalf("result = %q", line)
	}
}

func TestInstallGlobalCLIFailsWhenExecutableIsUnavailableAfterCommand(t *testing.T) {
	descriptor := globalCLIDescriptor{
		displayName: "Example CLI",
		executable:  "example",
		installSpec: "@example/cli@latest",
	}
	commands := globalCLICommands{
		lookPath: func(string) (string, error) { return "", errors.New("not found") },
		run:      func(string, ...string) ([]byte, error) { return nil, nil },
	}

	_, err := installGlobalCLI(descriptor, globalPackageManager{
		name: "npm", executable: "/tools/npm", install: []string{"install", "--global"},
	}, commands)
	if err == nil || !strings.Contains(err.Error(), "example is not available on PATH") {
		t.Fatalf("expected post-install verification error, got %v", err)
	}
}
