package install_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"angel-ai-opencode/internal/catalog"
	"angel-ai-opencode/internal/install"
)

const codegraphGuidance = `<!-- codegraph-guidance -->
## CodeGraph

Use CodeGraph before broad filesystem searches.
<!-- /codegraph-guidance -->`

func codegraphAssets(t *testing.T) string {
	t.Helper()
	assets := t.TempDir()
	write(t, filepath.Join(assets, "integrations", "codegraph", "mcp.json"), `{
  "mcp": {
    "codegraph": {
      "command": ["codegraph", "serve", "--mcp"],
      "enabled": true,
      "type": "local"
    }
  }
}`)
	write(t, filepath.Join(assets, "integrations", "codegraph", "AGENTS.md"), codegraphGuidance)
	return assets
}

func writeExecutable(t *testing.T, path, content string) {
	t.Helper()
	write(t, path, content)
	if err := os.Chmod(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestInstallationRemovesManagedCodegraphWhenNotSelected(t *testing.T) {
	assets := codegraphAssets(t)
	target := t.TempDir()
	opencodePath := filepath.Join(target, "opencode.json")
	agentsPath := filepath.Join(target, "AGENTS.md")
	write(t, opencodePath, `{"mcp":{"context7":{"type":"remote"},"codegraph":{"command":["custom"],"owner":"user"}}}`)
	write(t, agentsPath, "# Existing rules\n\n"+codegraphGuidance+"\n")

	if _, err := install.ApplyInstallation(install.InstallationRequest{
		Extras: map[string]bool{"codegraph": false}, AssetsDir: assets, ConfigDir: target,
	}); err != nil {
		t.Fatal(err)
	}

	var config map[string]any
	if err := json.Unmarshal(readFile(t, opencodePath), &config); err != nil {
		t.Fatal(err)
	}
	mcp := config["mcp"].(map[string]any)
	if _, ok := mcp["context7"]; !ok {
		t.Error("deselecting CodeGraph removed the unrelated Context7 MCP")
	}
	if _, ok := mcp["codegraph"]; ok {
		t.Error("deselecting CodeGraph left its managed MCP configured")
	}
	agents := string(readFile(t, agentsPath))
	if !strings.Contains(agents, "# Existing rules") {
		t.Error("deselecting CodeGraph dropped unrelated AGENTS.md content")
	}
	if strings.Contains(agents, "codegraph-guidance") {
		t.Error("deselecting CodeGraph left its managed guidance in AGENTS.md")
	}
}

func TestInstallationConfiguresSelectedCodegraphIdempotently(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test uses a POSIX executable stub")
	}
	assets := codegraphAssets(t)
	target := t.TempDir()
	binDir := t.TempDir()
	writeExecutable(t, filepath.Join(binDir, "codegraph"), "#!/bin/sh\nexit 0\n")
	t.Setenv("PATH", binDir)

	opencodePath := filepath.Join(target, "opencode.json")
	agentsPath := filepath.Join(target, "AGENTS.md")
	write(t, opencodePath, `{"mcp":{"context7":{"type":"remote"}}}`)
	write(t, agentsPath, `# Existing rules

<!-- codegraph-guidance -->
## Old CodeGraph rules
<!-- /codegraph-guidance -->
`)

	request := install.InstallationRequest{
		Extras: map[string]bool{"codegraph": true}, AssetsDir: assets, ConfigDir: target,
	}
	if _, err := install.ApplyInstallation(request); err != nil {
		t.Fatal(err)
	}
	if _, err := install.ApplyInstallation(request); err != nil {
		t.Fatal(err)
	}

	var config map[string]any
	if err := json.Unmarshal(readFile(t, opencodePath), &config); err != nil {
		t.Fatal(err)
	}
	mcp := config["mcp"].(map[string]any)
	if _, ok := mcp["context7"]; !ok {
		t.Error("CodeGraph integration dropped the existing Context7 MCP")
	}
	if _, ok := mcp["codegraph"]; !ok {
		t.Error("CodeGraph integration did not register the local MCP")
	}

	agents := string(readFile(t, agentsPath))
	if strings.Count(agents, "<!-- codegraph-guidance -->") != 1 {
		t.Errorf("expected one CodeGraph guidance block, got:\n%s", agents)
	}
	if !strings.Contains(agents, "# Existing rules") {
		t.Error("CodeGraph guidance update dropped existing AGENTS.md content")
	}
	backups, err := filepath.Glob(opencodePath + ".bak-*")
	if err != nil {
		t.Fatal(err)
	}
	if len(backups) != 1 {
		t.Errorf("identical reinstall created another opencode.json backup, got %d", len(backups))
	}
}

func TestInstallationInstallsCodegraphWhenMissing(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test uses a POSIX executable stub")
	}
	assets := codegraphAssets(t)
	target := t.TempDir()
	binDir := t.TempDir()
	commandLog := filepath.Join(t.TempDir(), "npm-args")
	writeExecutable(t, filepath.Join(binDir, "npm"), `#!/bin/sh
printf '%s\n' "$@" > "$COMMAND_LOG"
printf '# Concurrent rules\n' > "$RACE_TARGET"
printf '{"unrelated":true}\n' > "$RACE_CONFIG_TARGET"
printf '#!/bin/sh\nexit 0\n' > "$FAKE_BIN/codegraph"
/bin/chmod +x "$FAKE_BIN/codegraph"
`)
	t.Setenv("PATH", binDir)
	t.Setenv("COMMAND_LOG", commandLog)
	t.Setenv("FAKE_BIN", binDir)
	agentsPath := filepath.Join(target, "AGENTS.md")
	t.Setenv("RACE_TARGET", agentsPath)
	opencodePath := filepath.Join(target, "opencode.json")
	t.Setenv("RACE_CONFIG_TARGET", opencodePath)

	if _, err := install.ApplyInstallation(install.InstallationRequest{
		Extras: map[string]bool{"codegraph": true}, AssetsDir: assets, ConfigDir: target,
	}); err != nil {
		t.Fatal(err)
	}
	raw := readFile(t, commandLog)
	args := strings.Fields(string(raw))
	want := []string{"install", "--global", "@colbymchenry/codegraph@latest"}
	if strings.Join(args, " ") != strings.Join(want, " ") {
		t.Errorf("unexpected npm arguments: got %v, want %v", args, want)
	}
	backups, err := filepath.Glob(agentsPath + ".bak-*")
	if err != nil {
		t.Fatal(err)
	}
	if len(backups) != 1 || string(readFile(t, backups[0])) != "# Concurrent rules\n" {
		t.Fatalf("concurrent AGENTS.md was not backed up: %v", backups)
	}
	backupInfo, err := os.Stat(backups[0])
	if err != nil {
		t.Fatal(err)
	}
	if backupInfo.Mode().Perm() != 0o600 {
		t.Fatalf("backup permissions = %o, want 600", backupInfo.Mode().Perm())
	}
	if agents := string(readFile(t, agentsPath)); !strings.Contains(agents, "# Concurrent rules") || !strings.Contains(agents, codegraphGuidance) {
		t.Fatalf("concurrent AGENTS.md rules were not preserved:\n%s", agents)
	}
	var config map[string]any
	if err := json.Unmarshal(readFile(t, opencodePath), &config); err != nil {
		t.Fatal(err)
	}
	if config["unrelated"] != true {
		t.Fatalf("concurrent opencode.json key was not preserved: %v", config)
	}
	mcp, _ := config["mcp"].(map[string]any)
	if _, ok := mcp["codegraph"]; !ok {
		t.Fatalf("CodeGraph MCP missing after concurrent merge: %v", config)
	}
}

func TestInstallationRejectsUnterminatedCodegraphGuidance(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test uses a POSIX executable stub")
	}
	assets := codegraphAssets(t)
	target := t.TempDir()
	binDir := t.TempDir()
	writeExecutable(t, filepath.Join(binDir, "codegraph"), "#!/bin/sh\nexit 0\n")
	t.Setenv("PATH", binDir)
	write(t, filepath.Join(target, "AGENTS.md"), "# Rules\n\n<!-- codegraph-guidance -->\nunterminated\n")

	_, err := install.ApplyInstallation(install.InstallationRequest{
		Extras: map[string]bool{"codegraph": true}, AssetsDir: assets, ConfigDir: target,
	})
	if err == nil || !strings.Contains(err.Error(), "unterminated") {
		t.Fatalf("expected an unterminated guidance error, got %v", err)
	}
}

func TestInstallationComposesGlobalAgentsAndCodegraphIdempotently(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test uses a POSIX executable stub")
	}
	assets := codegraphAssets(t)
	write(t, filepath.Join(assets, "agents-md", "AGENTS.md"), "# Angel rules\n")
	write(t, filepath.Join(assets, "fragments", "mcp.json"), `{"mcp":{"context7":{"type":"remote"}}}`)
	categories, err := catalog.Load(assets)
	if err != nil {
		t.Fatal(err)
	}
	var items []catalog.Item
	for _, category := range categories {
		items = append(items, category.Items...)
	}

	target := t.TempDir()
	agentsPath := filepath.Join(target, "AGENTS.md")
	opencodePath := filepath.Join(target, "opencode.json")
	write(t, agentsPath, "# User rules that will be replaced\n")
	write(t, opencodePath, `{"mcp":{"context7":{"type":"remote"},"codegraph":{"command":["custom"],"enabled":false}}}`)
	binDir := t.TempDir()
	writeExecutable(t, filepath.Join(binDir, "codegraph"), "#!/bin/sh\nexit 0\n")
	t.Setenv("PATH", binDir)

	request := install.InstallationRequest{
		Items: items, Extras: map[string]bool{"codegraph": true}, AssetsDir: assets, ConfigDir: target,
	}
	plan, err := install.PlanInstallation(request)
	if err != nil {
		t.Fatal(err)
	}
	if !containsLineWith(plan, "REEMPLAZAR  "+agentsPath) {
		t.Fatalf("global AGENTS.md replacement path missing from plan: %v", plan)
	}
	if _, err := install.ApplyInstallation(request); err != nil {
		t.Fatal(err)
	}
	agentBackups, err := filepath.Glob(agentsPath + ".bak-*")
	if err != nil {
		t.Fatal(err)
	}
	if len(agentBackups) != 1 || string(readFile(t, agentBackups[0])) != "# User rules that will be replaced\n" {
		t.Fatalf("global AGENTS.md previous content was not backed up: %v", agentBackups)
	}

	agents := string(readFile(t, agentsPath))
	if strings.Contains(agents, "User rules") || !strings.Contains(agents, "# Angel rules") {
		t.Fatalf("global AGENTS.md was not replaced as planned:\n%s", agents)
	}
	if strings.Count(agents, "<!-- codegraph-guidance -->") != 1 {
		t.Fatalf("CodeGraph guidance count = %d", strings.Count(agents, "<!-- codegraph-guidance -->"))
	}
	var config map[string]any
	if err := json.Unmarshal(readFile(t, opencodePath), &config); err != nil {
		t.Fatal(err)
	}
	command := config["mcp"].(map[string]any)["codegraph"].(map[string]any)["command"].([]any)
	if strings.Join(anyStrings(command), " ") != "codegraph serve --mcp" {
		t.Fatalf("CodeGraph command = %v", command)
	}

	fixed := time.Unix(1_700_000_000, 0)
	if err := os.Chtimes(agentsPath, fixed, fixed); err != nil {
		t.Fatal(err)
	}
	beforeBackups, _ := filepath.Glob(filepath.Join(target, "*.bak-*"))
	report, err := install.ApplyInstallation(request)
	if err != nil {
		t.Fatal(err)
	}
	afterBackups, _ := filepath.Glob(filepath.Join(target, "*.bak-*"))
	if len(afterBackups) != len(beforeBackups) {
		t.Fatalf("identical reinstall created backups: before=%d after=%d", len(beforeBackups), len(afterBackups))
	}
	info, err := os.Stat(agentsPath)
	if err != nil {
		t.Fatal(err)
	}
	if !info.ModTime().Equal(fixed) {
		t.Fatalf("identical AGENTS.md was rewritten: %s", info.ModTime())
	}
	if !containsLineWith(report, "sin cambios "+agentsPath) {
		t.Fatalf("unchanged AGENTS.md not reported: %v", report)
	}
}

func anyStrings(values []any) []string {
	result := make([]string, len(values))
	for index, value := range values {
		result[index], _ = value.(string)
	}
	return result
}
