package install

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

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

func TestPlanInstallationPrefersNPM(t *testing.T) {
	var lookups []string
	useGlobalCLICommands(t, globalCLICommands{
		lookPath: func(name string) (string, error) {
			lookups = append(lookups, name)
			if name == "npm" {
				return "/tools/npm", nil
			}
			return "", errors.New("not found")
		},
		run: func(path string, args ...string) ([]byte, error) {
			t.Fatalf("planning with npm must not run %q %v", path, args)
			return nil, nil
		},
	})

	plan, err := PlanInstallation(InstallationRequest{
		Extras:    map[string]bool{codegraphOptionKey: true},
		AssetsDir: testCodegraphAssets(t),
		ConfigDir: t.TempDir(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(lookups, []string{"npm"}) {
		t.Fatalf("executable lookups = %v, want npm only", lookups)
	}
	if !hasLineContaining(plan, codegraphPackage) {
		t.Fatalf("CodeGraph latest action missing from plan: %v", plan)
	}
}

func TestPlanInstallationUsesValidatedPNPMFallback(t *testing.T) {
	var commandsRun []string
	useGlobalCLICommands(t, globalCLICommands{
		lookPath: func(name string) (string, error) {
			if name == "pnpm" {
				return "/tools/pnpm", nil
			}
			return "", errors.New("not found")
		},
		run: func(path string, args ...string) ([]byte, error) {
			commandsRun = append(commandsRun, strings.Join(append([]string{path}, args...), " "))
			return []byte("/global/pnpm/bin\n"), nil
		},
	})

	plan, err := PlanInstallation(InstallationRequest{
		Extras:    map[string]bool{codegraphOptionKey: true},
		AssetsDir: testCodegraphAssets(t),
		ConfigDir: t.TempDir(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(commandsRun, []string{"/tools/pnpm bin -g"}) {
		t.Fatalf("planning commands = %v", commandsRun)
	}
	if !hasLineContaining(plan, codegraphPackage) {
		t.Fatalf("CodeGraph latest action missing from plan: %v", plan)
	}
}

func TestPlanInstallationRejectsInvalidPNPMFallback(t *testing.T) {
	tests := []struct {
		name   string
		output []byte
		err    error
		want   string
	}{
		{name: "empty global bin", output: []byte(" \n"), want: "empty path"},
		{name: "global bin failure", output: []byte("PNPM_HOME is unset\n"), err: errors.New("exit 1"), want: "PNPM_HOME is unset"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			useGlobalCLICommands(t, globalCLICommands{
				lookPath: func(name string) (string, error) {
					if name == "pnpm" {
						return "/tools/pnpm", nil
					}
					return "", errors.New("not found")
				},
				run: func(string, ...string) ([]byte, error) { return test.output, test.err },
			})

			_, err := PlanInstallation(InstallationRequest{
				Extras:    map[string]bool{codegraphOptionKey: true},
				AssetsDir: testCodegraphAssets(t),
				ConfigDir: t.TempDir(),
			})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("expected error containing %q, got %v", test.want, err)
			}
		})
	}
}

func TestPlanInstallationChecksOpenSpecNodeVersion(t *testing.T) {
	tests := []struct {
		name      string
		version   string
		wantError string
	}{
		{name: "supported", version: "v20.19.0\n"},
		{name: "unsupported", version: "v20.18.1\n", wantError: "requires Node.js >=20.19.0"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			useGlobalCLICommands(t, globalCLICommands{
				lookPath: func(name string) (string, error) { return "/tools/" + name, nil },
				run: func(path string, args ...string) ([]byte, error) {
					if path != "/tools/node" || !reflect.DeepEqual(args, []string{"--version"}) {
						t.Fatalf("unexpected planning command %q %v", path, args)
					}
					return []byte(test.version), nil
				},
			})

			plan, err := PlanInstallation(InstallationRequest{
				Extras:    map[string]bool{openSpecOptionKey: true},
				ConfigDir: t.TempDir(),
			})
			if test.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantError) {
					t.Fatalf("expected error containing %q, got %v", test.wantError, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !hasLineContaining(plan, openSpecPackage) {
				t.Fatalf("OpenSpec latest action missing from plan: %v", plan)
			}
		})
	}
}

func TestPlanInstallationIsNonMutatingForExistingSelectedCLIs(t *testing.T) {
	assets := testCodegraphAssets(t)
	source := filepath.Join(assets, "managed.md")
	writeTestFile(t, source, "desired\n")
	target := t.TempDir()
	destination := filepath.Join(target, "managed.md")
	writeTestFile(t, destination, "original\n")

	var commandsRun []string
	useGlobalCLICommands(t, globalCLICommands{
		lookPath: func(name string) (string, error) {
			return "/tools/" + name, nil
		},
		run: func(path string, args ...string) ([]byte, error) {
			commandsRun = append(commandsRun, strings.Join(append([]string{path}, args...), " "))
			if path == "/tools/node" {
				return []byte("v22.0.0\n"), nil
			}
			return nil, errors.New("package installation attempted during planning")
		},
	})

	plan, err := PlanInstallation(InstallationRequest{
		Items: []catalog.Item{{
			Name: "managed", Source: source, Dest: "managed.md", Kind: catalog.CopyFile,
		}},
		Extras: map[string]bool{
			codegraphOptionKey: true,
			openSpecOptionKey:  true,
		},
		AssetsDir: assets,
		ConfigDir: target,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(destination); err != nil || string(got) != "original\n" {
		t.Fatalf("planning mutated destination: content=%q err=%v", got, err)
	}
	if !reflect.DeepEqual(commandsRun, []string{"/tools/node --version"}) {
		t.Fatalf("planning commands = %v, want only Node.js preflight", commandsRun)
	}
	for _, packageSpec := range []string{codegraphPackage, openSpecPackage} {
		if !hasLineContaining(plan, packageSpec) {
			t.Errorf("existing selected CLI %s missing from plan: %v", packageSpec, plan)
		}
	}
}

func TestApplyInstallationUpdatesSelectedCLIsInOrderAndRepreparesOnce(t *testing.T) {
	preparations := countApplyPreparations(t)
	target := t.TempDir()
	var packageCommands []string
	var executableChecks []string
	useGlobalCLICommands(t, globalCLICommands{
		lookPath: func(name string) (string, error) {
			switch name {
			case "codegraph", "openspec":
				executableChecks = append(executableChecks, name)
			}
			return "/tools/" + name, nil
		},
		run: func(path string, args ...string) ([]byte, error) {
			if path == "/tools/node" {
				return []byte("v22.0.0\n"), nil
			}
			packageCommands = append(packageCommands, strings.Join(args, " "))
			return nil, nil
		},
	})

	report, err := ApplyInstallation(InstallationRequest{
		Extras: map[string]bool{
			codegraphOptionKey: true,
			openSpecOptionKey:  true,
		},
		AssetsDir: testCodegraphAssets(t),
		ConfigDir: target,
	})
	if err != nil {
		t.Fatal(err)
	}
	wantCommands := []string{
		"install --global " + codegraphPackage,
		"install --global " + openSpecPackage,
	}
	if !reflect.DeepEqual(packageCommands, wantCommands) {
		t.Fatalf("package commands = %v, want %v", packageCommands, wantCommands)
	}
	if !reflect.DeepEqual(executableChecks, []string{"codegraph", "openspec"}) {
		t.Fatalf("post-install executable checks = %v", executableChecks)
	}
	if *preparations != 2 {
		t.Fatalf("preparation count = %d, want initial preparation plus one repreparation", *preparations)
	}
	if len(report) < 2 || report[0] != "instalado  "+codegraphPackage || report[1] != "instalado  "+openSpecPackage {
		t.Fatalf("CLI results are not in descriptor order: %v", report)
	}
	var config map[string]any
	rawConfig, err := os.ReadFile(filepath.Join(target, "opencode.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(rawConfig, &config); err != nil {
		t.Fatal(err)
	}
	mcp, _ := config["mcp"].(map[string]any)
	if _, ok := mcp["codegraph"]; !ok {
		t.Fatalf("CodeGraph MCP missing after shared CLI updates: %v", config)
	}
	agents, err := os.ReadFile(filepath.Join(target, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(agents), "<!-- codegraph-guidance -->") {
		t.Fatalf("CodeGraph AGENTS.md guidance missing after shared CLI updates: %s", agents)
	}
}

func TestApplyInstallationReturnsPartialCLISuccessWithoutConfigurationWrites(t *testing.T) {
	preparations := countApplyPreparations(t)
	assets := testCodegraphAssets(t)
	source := filepath.Join(assets, "managed.md")
	writeTestFile(t, source, "managed\n")
	target := t.TempDir()
	var packageCommands []string
	useGlobalCLICommands(t, globalCLICommands{
		lookPath: func(name string) (string, error) { return "/tools/" + name, nil },
		run: func(path string, args ...string) ([]byte, error) {
			if path == "/tools/node" {
				return []byte("v22.0.0\n"), nil
			}
			packageCommands = append(packageCommands, strings.Join(args, " "))
			if args[len(args)-1] == openSpecPackage {
				return []byte("registry unavailable\n"), errors.New("exit 1")
			}
			return nil, nil
		},
	})

	report, err := ApplyInstallation(InstallationRequest{
		Items: []catalog.Item{{
			Name: "managed", Source: source, Dest: "managed.md", Kind: catalog.CopyFile,
		}},
		Extras: map[string]bool{
			codegraphOptionKey: true,
			openSpecOptionKey:  true,
		},
		AssetsDir: assets,
		ConfigDir: target,
	})
	if err == nil || !strings.Contains(err.Error(), "installing OpenSpec with npm") {
		t.Fatalf("expected second CLI failure, got %v", err)
	}
	if !reflect.DeepEqual(report, []string{"instalado  " + codegraphPackage}) {
		t.Fatalf("partial CLI report = %v", report)
	}
	wantCommands := []string{
		"install --global " + codegraphPackage,
		"install --global " + openSpecPackage,
	}
	if !reflect.DeepEqual(packageCommands, wantCommands) {
		t.Fatalf("package commands = %v, want %v", packageCommands, wantCommands)
	}
	if *preparations != 1 {
		t.Fatalf("preparation count after CLI failure = %d, want 1", *preparations)
	}
	for _, path := range []string{
		filepath.Join(target, "managed.md"),
		filepath.Join(target, "opencode.json"),
		filepath.Join(target, "AGENTS.md"),
	} {
		if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
			t.Errorf("configuration path was written after CLI failure: %s (err=%v)", path, statErr)
		}
	}
}

func TestApplyInstallationRequiresPostInstallExecutable(t *testing.T) {
	target := t.TempDir()
	source := filepath.Join(t.TempDir(), "managed.md")
	writeTestFile(t, source, "managed\n")
	useGlobalCLICommands(t, globalCLICommands{
		lookPath: func(name string) (string, error) {
			if name == "openspec" {
				return "", errors.New("not found")
			}
			return "/tools/" + name, nil
		},
		run: func(path string, args ...string) ([]byte, error) {
			if path == "/tools/node" {
				return []byte("v22.0.0\n"), nil
			}
			return nil, nil
		},
	})

	_, err := ApplyInstallation(InstallationRequest{
		Items: []catalog.Item{{
			Name: "managed", Source: source, Dest: "managed.md", Kind: catalog.CopyFile,
		}},
		Extras:    map[string]bool{openSpecOptionKey: true},
		ConfigDir: target,
	})
	if err == nil || !strings.Contains(err.Error(), "openspec is not available on PATH") {
		t.Fatalf("expected post-install executable error, got %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(target, "managed.md")); !os.IsNotExist(statErr) {
		t.Fatalf("configuration was written after executable verification failure: %v", statErr)
	}
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
	categories, err := catalog.Load(assets)
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
			entries, err := os.ReadDir(item.Source)
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
	var packageCommands []string
	useGlobalCLICommands(t, globalCLICommands{
		lookPath: func(name string) (string, error) { return "/tools/" + name, nil },
		run: func(path string, args ...string) ([]byte, error) {
			if path == "/tools/node" {
				return []byte("v22.0.0\n"), nil
			}
			packageCommands = append(packageCommands, strings.Join(args, " "))
			return nil, nil
		},
	})

	if _, err := ApplyInstallation(InstallationRequest{
		Extras:    map[string]bool{openSpecOptionKey: true},
		AssetsDir: assets,
		ConfigDir: target,
	}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(packageCommands, []string{"install --global " + openSpecPackage}) {
		t.Fatalf("OpenSpec package commands = %v", packageCommands)
	}
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
