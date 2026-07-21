package install

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"angel-ai-opencode/internal/catalog"
)

func useCMUXAvailable(t *testing.T) {
	t.Helper()
	useGlobalCLICommands(t, globalCLICommands{
		lookPath: func(name string) (string, error) {
			if name != "cmux" {
				t.Fatalf("unexpected executable lookup %q", name)
			}
			return "/tools/cmux", nil
		},
		run: func(path string, args ...string) ([]byte, error) {
			t.Fatalf("unexpected command %q %v", path, args)
			return nil, nil
		},
	})
}

func TestCMUXPlanAndApplyVendoredPluginsWithoutConfigMutation(t *testing.T) {
	useCMUXAvailable(t)
	assets := filepath.Join("..", "..", "assets")
	target := t.TempDir()
	configPath := filepath.Join(target, "opencode.json")
	originalConfig := []byte("{\"share\":\"disabled\"}\n")
	writeTestFile(t, configPath, string(originalConfig))
	request := InstallationRequest{
		Extras:    map[string]bool{cmuxOptionKey: true},
		AssetsDir: assets,
		ConfigDir: target,
	}

	plan, err := PlanInstallation(request)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range cmuxPluginFiles {
		pluginPath := filepath.Join(target, "plugins", name)
		if !hasLineContaining(plan, pluginPath) {
			t.Errorf("cmux plugin target missing from plan: %s; plan=%v", pluginPath, plan)
		}
	}
	if got, err := os.ReadFile(configPath); err != nil || !bytes.Equal(got, originalConfig) {
		t.Fatalf("planning changed opencode.json: content=%q err=%v", got, err)
	}

	report, err := ApplyInstallation(request)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range cmuxPluginFiles {
		sourcePath := filepath.Join(assets, "integrations", "cmux", name)
		pluginPath := filepath.Join(target, "plugins", name)
		want, err := os.ReadFile(sourcePath)
		if err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(pluginPath)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("installed %s does not match vendored snapshot", name)
		}
		if !hasLineContaining(report, pluginPath) {
			t.Errorf("installed cmux plugin missing from report: %s; report=%v", pluginPath, report)
		}
	}
	if got, err := os.ReadFile(configPath); err != nil || !bytes.Equal(got, originalConfig) {
		t.Fatalf("cmux application changed opencode.json: content=%q err=%v", got, err)
	}
}

func TestCMUXSelectedMissingPreflightStopsBeforeMutations(t *testing.T) {
	assets := filepath.Join("..", "..", "assets")
	source := filepath.Join(t.TempDir(), "managed.md")
	writeTestFile(t, source, "managed\n")
	tests := []struct {
		name string
		run  func(InstallationRequest) error
	}{
		{
			name: "plan",
			run: func(request InstallationRequest) error {
				_, err := PlanInstallation(request)
				return err
			},
		},
		{
			name: "apply",
			run: func(request InstallationRequest) error {
				_, err := ApplyInstallation(request)
				return err
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			target := t.TempDir()
			useGlobalCLICommands(t, globalCLICommands{
				lookPath: func(name string) (string, error) {
					if name != "cmux" {
						t.Fatalf("unexpected executable lookup %q", name)
					}
					return "", errors.New("not found")
				},
				run: func(path string, args ...string) ([]byte, error) {
					t.Fatalf("mutation command attempted: %q %v", path, args)
					return nil, nil
				},
			})
			request := InstallationRequest{
				Items: []catalog.Item{{
					Name: "managed", Source: source, Dest: "managed.md", Kind: catalog.CopyFile,
				}},
				Extras:    map[string]bool{cmuxOptionKey: true},
				AssetsDir: assets,
				ConfigDir: target,
			}

			err := test.run(request)
			if err == nil || !strings.Contains(err.Error(), "cmux extra requires cmux") || !strings.Contains(err.Error(), "PATH") {
				t.Fatalf("expected missing cmux preflight error, got %v", err)
			}
			for _, path := range []string{
				filepath.Join(target, "managed.md"),
				filepath.Join(target, "plugins", "cmux-session.js"),
				filepath.Join(target, "plugins", "cmux-feed.js"),
			} {
				if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
					t.Errorf("path mutated before cmux preflight failure: %s (err=%v)", path, statErr)
				}
			}
		})
	}
}

func TestCMUXUnselectedBypassesExecutableLookup(t *testing.T) {
	useGlobalCLICommands(t, globalCLICommands{
		lookPath: func(name string) (string, error) {
			t.Fatalf("unexpected executable lookup %q", name)
			return "", nil
		},
		run: func(path string, args ...string) ([]byte, error) {
			t.Fatalf("unexpected command %q %v", path, args)
			return nil, nil
		},
	})
	request := InstallationRequest{
		Extras:    map[string]bool{cmuxOptionKey: false},
		AssetsDir: filepath.Join("..", "..", "assets"),
		ConfigDir: t.TempDir(),
	}
	if _, err := PlanInstallation(request); err != nil {
		t.Fatalf("unselected cmux planning failed: %v", err)
	}
	if _, err := ApplyInstallation(request); err != nil {
		t.Fatalf("unselected cmux application failed: %v", err)
	}
}

func TestCMUXReconciliationBackupsAndIdempotency(t *testing.T) {
	useCMUXAvailable(t)
	assets := filepath.Join("..", "..", "assets")
	target := t.TempDir()
	request := InstallationRequest{
		Extras:    map[string]bool{cmuxOptionKey: true},
		AssetsDir: assets,
		ConfigDir: target,
	}
	previous := make(map[string][]byte, len(cmuxPluginFiles))
	for _, name := range cmuxPluginFiles {
		previous[name] = []byte("divergent " + name + "\n")
		writeTestFile(t, filepath.Join(target, "plugins", name), string(previous[name]))
	}

	if _, err := ApplyInstallation(request); err != nil {
		t.Fatal(err)
	}
	fixed := time.Unix(1_700_000_000, 0)
	for _, name := range cmuxPluginFiles {
		pluginPath := filepath.Join(target, "plugins", name)
		want, err := os.ReadFile(filepath.Join(assets, "integrations", "cmux", name))
		if err != nil {
			t.Fatal(err)
		}
		if got, err := os.ReadFile(pluginPath); err != nil || !bytes.Equal(got, want) {
			t.Errorf("reconciled %s content mismatch: err=%v", name, err)
		}
		backups, err := filepath.Glob(pluginPath + ".bak-*")
		if err != nil {
			t.Fatal(err)
		}
		if len(backups) != 1 {
			t.Fatalf("%s backups = %d, want 1", name, len(backups))
		}
		if got, err := os.ReadFile(backups[0]); err != nil || !bytes.Equal(got, previous[name]) {
			t.Errorf("%s backup did not preserve divergent content: content=%q err=%v", name, got, err)
		}
		if err := os.Chtimes(pluginPath, fixed, fixed); err != nil {
			t.Fatal(err)
		}
	}

	report, err := ApplyInstallation(request)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range cmuxPluginFiles {
		pluginPath := filepath.Join(target, "plugins", name)
		backups, err := filepath.Glob(pluginPath + ".bak-*")
		if err != nil {
			t.Fatal(err)
		}
		if len(backups) != 1 {
			t.Errorf("idempotent reapplication created extra %s backups: %v", name, backups)
		}
		info, err := os.Stat(pluginPath)
		if err != nil {
			t.Fatal(err)
		}
		if !info.ModTime().Equal(fixed) {
			t.Errorf("idempotent reapplication rewrote %s: modtime=%s", name, info.ModTime())
		}
		if !hasLineContaining(report, "sin cambios "+pluginPath) {
			t.Errorf("unchanged %s missing from reapplication report: %v", name, report)
		}
	}
}

func TestCMUXPartialRetryConvergesIndependently(t *testing.T) {
	useCMUXAvailable(t)
	assets := filepath.Join("..", "..", "assets")
	target := t.TempDir()
	matchingName := cmuxPluginFiles[0]
	missingName := cmuxPluginFiles[1]
	matchingPath := filepath.Join(target, "plugins", matchingName)
	matchingContent, err := os.ReadFile(filepath.Join(assets, "integrations", "cmux", matchingName))
	if err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, matchingPath, string(matchingContent))
	fixed := time.Unix(1_700_000_000, 0)
	if err := os.Chtimes(matchingPath, fixed, fixed); err != nil {
		t.Fatal(err)
	}

	report, err := ApplyInstallation(InstallationRequest{
		Extras:    map[string]bool{cmuxOptionKey: true},
		AssetsDir: assets,
		ConfigDir: target,
	})
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(matchingPath)
	if err != nil {
		t.Fatal(err)
	}
	if !info.ModTime().Equal(fixed) {
		t.Fatalf("retry rewrote completed plugin %s", matchingName)
	}
	missingPath := filepath.Join(target, "plugins", missingName)
	wantMissing, err := os.ReadFile(filepath.Join(assets, "integrations", "cmux", missingName))
	if err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(missingPath); err != nil || !bytes.Equal(got, wantMissing) {
		t.Fatalf("retry did not install incomplete plugin %s: err=%v", missingName, err)
	}
	if !hasLineContaining(report, "sin cambios "+matchingPath) || !hasLineContaining(report, "creado    "+missingPath) {
		t.Fatalf("partial retry report does not show independent convergence: %v", report)
	}
}

func TestCMUXDeselectionPreservesInstalledPlugins(t *testing.T) {
	useCMUXAvailable(t)
	assets := filepath.Join("..", "..", "assets")
	target := t.TempDir()
	selected := InstallationRequest{
		Extras:    map[string]bool{cmuxOptionKey: true},
		AssetsDir: assets,
		ConfigDir: target,
	}
	if _, err := ApplyInstallation(selected); err != nil {
		t.Fatal(err)
	}
	fixed := time.Unix(1_700_000_000, 0)
	preserved := make(map[string][]byte, len(cmuxPluginFiles))
	for _, name := range cmuxPluginFiles {
		pluginPath := filepath.Join(target, "plugins", name)
		preserved[name] = []byte("preserved after deselection: " + name + "\n")
		if err := os.WriteFile(pluginPath, preserved[name], 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(pluginPath, fixed, fixed); err != nil {
			t.Fatal(err)
		}
	}
	useGlobalCLICommands(t, globalCLICommands{
		lookPath: func(name string) (string, error) {
			t.Fatalf("deselected cmux looked up executable %q", name)
			return "", nil
		},
		run: func(path string, args ...string) ([]byte, error) {
			t.Fatalf("deselection ran command %q %v", path, args)
			return nil, nil
		},
	})

	if _, err := ApplyInstallation(InstallationRequest{
		Extras:    map[string]bool{cmuxOptionKey: false},
		AssetsDir: assets,
		ConfigDir: target,
	}); err != nil {
		t.Fatal(err)
	}
	for _, name := range cmuxPluginFiles {
		pluginPath := filepath.Join(target, "plugins", name)
		if got, err := os.ReadFile(pluginPath); err != nil || !bytes.Equal(got, preserved[name]) {
			t.Errorf("deselection changed %s: content=%q err=%v", name, got, err)
		}
		info, err := os.Stat(pluginPath)
		if err != nil {
			t.Fatal(err)
		}
		if !info.ModTime().Equal(fixed) {
			t.Errorf("deselection rewrote %s: modtime=%s", name, info.ModTime())
		}
	}
}
