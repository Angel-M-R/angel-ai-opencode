package install_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"angel-ai-opencode/internal/catalog"
	"angel-ai-opencode/internal/install"
)

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestApplyReplacesOnlyMatchingFilesAndSkipsIdenticalContent(t *testing.T) {
	assets := t.TempDir()
	write(t, filepath.Join(assets, "skills", "my-skill", "SKILL.md"), "# managed\n")
	categories, err := catalog.Load(assets)
	if err != nil {
		t.Fatal(err)
	}

	target := t.TempDir()
	managedPath := filepath.Join(target, "skills", "my-skill", "SKILL.md")
	userPath := filepath.Join(target, "skills", "my-skill", "notes.md")
	write(t, managedPath, "# previous\n")
	write(t, userPath, "keep me\n")

	request := install.InstallationRequest{Items: categories[0].Items, AssetsDir: assets, ConfigDir: target}
	report, err := install.ApplyInstallation(request)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(readFile(t, managedPath)); got != "# managed\n" {
		t.Fatalf("managed file = %q", got)
	}
	if got := string(readFile(t, userPath)); got != "keep me\n" {
		t.Fatalf("unrelated file was changed: %q", got)
	}
	if !containsLineWith(report, "actualizado "+managedPath) {
		t.Fatalf("update not reported: %v", report)
	}
	backups, _ := filepath.Glob(managedPath + ".bak-*")
	if len(backups) != 1 {
		t.Fatalf("expected one managed-file backup, got %d", len(backups))
	}

	fixed := time.Unix(1_700_000_000, 0)
	if err := os.Chtimes(managedPath, fixed, fixed); err != nil {
		t.Fatal(err)
	}
	report, err = install.ApplyInstallation(request)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(managedPath)
	if err != nil {
		t.Fatal(err)
	}
	if !info.ModTime().Equal(fixed) {
		t.Fatalf("identical file was rewritten: modtime = %s", info.ModTime())
	}
	if !containsLineWith(report, "sin cambios "+managedPath) {
		t.Fatalf("unchanged file not reported: %v", report)
	}
}

func TestApplyUsesKeySpecificArrayMergeRules(t *testing.T) {
	assets := t.TempDir()
	write(t, filepath.Join(assets, "fragments", "settings.json"), `{
  "plugin": ["managed-plugin@latest"],
  "mcp": {
    "codegraph": {
      "command": ["codegraph", "serve", "--mcp"],
      "enabled": true,
      "type": "local"
    }
  }
}`)
	categories, err := catalog.Load(assets)
	if err != nil {
		t.Fatal(err)
	}

	target := t.TempDir()
	configPath := filepath.Join(target, "opencode.json")
	write(t, configPath, `{
  "plugin": ["foreign-plugin", "managed-plugin@1.0.0"],
  "mcp": {
    "codegraph": {
      "command": ["old-codegraph", "serve"],
      "enabled": false,
      "type": "local"
    }
  }
}`)

	if _, err := install.ApplyInstallation(install.InstallationRequest{
		Items: categories[0].Items, AssetsDir: assets, ConfigDir: target,
	}); err != nil {
		t.Fatal(err)
	}
	var config map[string]any
	if err := json.Unmarshal(readFile(t, configPath), &config); err != nil {
		t.Fatal(err)
	}
	plugins := config["plugin"].([]any)
	wantPlugins := []string{"foreign-plugin", "managed-plugin@latest"}
	if len(plugins) != len(wantPlugins) {
		t.Fatalf("plugins = %v, want %v", plugins, wantPlugins)
	}
	for index, want := range wantPlugins {
		if plugins[index] != want {
			t.Fatalf("plugins = %v, want %v", plugins, wantPlugins)
		}
	}
	command := config["mcp"].(map[string]any)["codegraph"].(map[string]any)["command"].([]any)
	wantCommand := []string{"codegraph", "serve", "--mcp"}
	if len(command) != len(wantCommand) {
		t.Fatalf("command = %v, want %v", command, wantCommand)
	}
	for index, want := range wantCommand {
		if command[index] != want {
			t.Fatalf("command = %v, want %v", command, wantCommand)
		}
	}
}

func TestApplyDoesNotReorderMatchingPlugins(t *testing.T) {
	assets := t.TempDir()
	write(t, filepath.Join(assets, "fragments", "settings.json"), `{
  "$schema":"https://opencode.ai/config.json",
  "plugin":["managed-plugin@latest"]
}`)
	categories, err := catalog.Load(assets)
	if err != nil {
		t.Fatal(err)
	}
	target := t.TempDir()
	configPath := filepath.Join(target, "opencode.json")
	original := `{"$schema":"https://opencode.ai/config.json","plugin":["managed-plugin@latest","foreign-plugin"]}`
	write(t, configPath, original)

	if _, err := install.ApplyInstallation(install.InstallationRequest{
		Items: categories[0].Items, AssetsDir: assets, ConfigDir: target,
	}); err != nil {
		t.Fatal(err)
	}
	if got := string(readFile(t, configPath)); got != original {
		t.Fatalf("matching plugin was reordered: got %s", got)
	}
	backups, _ := filepath.Glob(configPath + ".bak-*")
	if len(backups) != 0 {
		t.Fatalf("unchanged plugin config created backups: %v", backups)
	}
}

func TestApplyMergesTUIPluginsIdempotently(t *testing.T) {
	assets := t.TempDir()
	write(t, filepath.Join(assets, "tui-plugins", "angel-logo.tsx"), "export default {}\n")
	write(t, filepath.Join(assets, "tui-plugins", "mcp-footer-state.ts"), "export default {}\n")
	target := t.TempDir()
	tuiPath := filepath.Join(target, "tui.json")
	write(t, tuiPath, `{
  "plugin": ["foreign-plugin", "opencode-subagent-statusline@1.0.0"],
  "theme": "previous-theme"
}`)

	request := install.InstallationRequest{
		Extras: map[string]bool{
			"angel-logo": true, "theme": true, "subagent-statusline": true,
		},
		AssetsDir: assets,
		ConfigDir: target,
	}
	if _, err := install.ApplyInstallation(request); err != nil {
		t.Fatal(err)
	}
	var config map[string]any
	if err := json.Unmarshal(readFile(t, tuiPath), &config); err != nil {
		t.Fatal(err)
	}
	plugins := config["plugin"].([]any)
	want := []string{
		"foreign-plugin",
		"opencode-subagent-statusline",
		filepath.Join(target, "tui-plugins", "angel-logo.tsx"),
	}
	if len(plugins) != len(want) {
		t.Fatalf("TUI plugins = %v, want %v", plugins, want)
	}
	for index, expected := range want {
		if plugins[index] != expected {
			t.Fatalf("TUI plugins = %v, want %v", plugins, want)
		}
	}
	if config["theme"] != "one-dark-pro" {
		t.Fatalf("TUI theme = %v", config["theme"])
	}

	backups, _ := filepath.Glob(tuiPath + ".bak-*")
	if len(backups) != 1 {
		t.Fatalf("first TUI merge backups = %d, want 1", len(backups))
	}
	if _, err := install.ApplyInstallation(request); err != nil {
		t.Fatal(err)
	}
	backups, _ = filepath.Glob(tuiPath + ".bak-*")
	if len(backups) != 1 {
		t.Fatalf("identical TUI reinstall created another backup: %d", len(backups))
	}
	var reinstalled map[string]any
	if err := json.Unmarshal(readFile(t, tuiPath), &reinstalled); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(config, reinstalled) {
		t.Fatalf("TUI reinstall changed config: before=%v after=%v", config, reinstalled)
	}
}

func containsLineWith(lines []string, expected string) bool {
	for _, line := range lines {
		if strings.Contains(line, expected) {
			return true
		}
	}
	return false
}

func TestLoadAndApply(t *testing.T) {
	assets := t.TempDir()
	write(t, filepath.Join(assets, "agents", "angel-orchestrator.md"), "---\ndescription: x\n---\nprompt")
	write(t, filepath.Join(assets, "skills", "my-skill", "SKILL.md"), "# skill")
	write(t, filepath.Join(assets, "fragments", "mcp.json"), `{"mcp":{"context7":{"type":"remote"}},"plugin":["a"]}`)

	categories, err := catalog.Load(assets)
	if err != nil {
		t.Fatal(err)
	}
	if len(categories) != 3 {
		t.Fatalf("expected 3 categories, got %d", len(categories))
	}

	target := t.TempDir()
	// Existing config: merge must keep unknown keys and union arrays.
	write(t, filepath.Join(target, "opencode.json"), `{"share":"disabled","plugin":["a","b"],"mcp":{"engram":{"type":"local"}}}`)

	var items []catalog.Item
	for _, category := range categories {
		items = append(items, category.Items...)
	}
	request := install.InstallationRequest{Items: items, AssetsDir: assets, ConfigDir: target}
	if _, err := install.ApplyInstallation(request); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		filepath.Join(target, "agents", "angel-orchestrator.md"),
		filepath.Join(target, "skills", "my-skill", "SKILL.md"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("missing installed file: %s", path)
		}
	}

	raw, err := os.ReadFile(filepath.Join(target, "opencode.json"))
	if err != nil {
		t.Fatal(err)
	}
	var config map[string]any
	if err := json.Unmarshal(raw, &config); err != nil {
		t.Fatal(err)
	}
	if config["share"] != "disabled" {
		t.Error("merge dropped existing share key")
	}
	mcp := config["mcp"].(map[string]any)
	if _, ok := mcp["engram"]; !ok {
		t.Error("merge dropped existing mcp.engram")
	}
	if _, ok := mcp["context7"]; !ok {
		t.Error("merge did not add mcp.context7")
	}
	if plugins := config["plugin"].([]any); len(plugins) != 2 {
		t.Errorf("plugin array union failed: %v", plugins)
	}

	backups, _ := filepath.Glob(filepath.Join(target, "opencode.json.bak-*"))
	if len(backups) != 1 {
		t.Errorf("expected 1 backup, got %d", len(backups))
	}

	if _, err := install.ApplyInstallation(request); err != nil {
		t.Fatal(err)
	}
	backups, _ = filepath.Glob(filepath.Join(target, "opencode.json.bak-*"))
	if len(backups) != 1 {
		t.Errorf("identical reinstall created another backup, got %d", len(backups))
	}
}
