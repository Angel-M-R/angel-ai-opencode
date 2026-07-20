package install_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

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

func TestApplyExtrasSkipsCodegraphWhenNotSelected(t *testing.T) {
	assets := codegraphAssets(t)
	target := t.TempDir()
	opencodePath := filepath.Join(target, "opencode.json")
	agentsPath := filepath.Join(target, "AGENTS.md")
	write(t, opencodePath, `{"mcp":{"context7":{"type":"remote"}}}`)
	write(t, agentsPath, "# Existing rules\n")

	beforeConfig, _ := os.ReadFile(opencodePath)
	beforeAgents, _ := os.ReadFile(agentsPath)
	if _, err := install.ApplyExtras(map[string]bool{"codegraph": false}, assets, target); err != nil {
		t.Fatal(err)
	}
	afterConfig, _ := os.ReadFile(opencodePath)
	afterAgents, _ := os.ReadFile(agentsPath)

	if string(afterConfig) != string(beforeConfig) {
		t.Error("unselected CodeGraph changed opencode.json")
	}
	if string(afterAgents) != string(beforeAgents) {
		t.Error("unselected CodeGraph changed AGENTS.md")
	}
}

func TestApplyExtrasConfiguresSelectedCodegraphIdempotently(t *testing.T) {
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

<!-- gentle-ai:codegraph-guidance -->
## Old CodeGraph rules
<!-- /gentle-ai:codegraph-guidance -->
`)

	for range 2 {
		if _, err := install.ApplyExtras(map[string]bool{"codegraph": true}, assets, target); err != nil {
			t.Fatal(err)
		}
	}

	rawConfig, err := os.ReadFile(opencodePath)
	if err != nil {
		t.Fatal(err)
	}
	var config map[string]any
	if err := json.Unmarshal(rawConfig, &config); err != nil {
		t.Fatal(err)
	}
	mcp := config["mcp"].(map[string]any)
	if _, ok := mcp["context7"]; !ok {
		t.Error("CodeGraph integration dropped the existing Context7 MCP")
	}
	if _, ok := mcp["codegraph"]; !ok {
		t.Error("CodeGraph integration did not register the local MCP")
	}

	rawAgents, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatal(err)
	}
	agents := string(rawAgents)
	if strings.Count(agents, "<!-- codegraph-guidance -->") != 1 {
		t.Errorf("expected one CodeGraph guidance block, got:\n%s", agents)
	}
	if strings.Contains(agents, "gentle-ai:codegraph-guidance") {
		t.Error("legacy gentle-ai CodeGraph guidance was not replaced")
	}
	if !strings.Contains(agents, "# Existing rules") {
		t.Error("CodeGraph guidance update dropped existing AGENTS.md content")
	}
}

func TestApplyExtrasInstallsCodegraphWhenMissing(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test uses a POSIX executable stub")
	}
	assets := codegraphAssets(t)
	target := t.TempDir()
	binDir := t.TempDir()
	commandLog := filepath.Join(t.TempDir(), "npm-args")
	writeExecutable(t, filepath.Join(binDir, "npm"), `#!/bin/sh
printf '%s\n' "$@" > "$COMMAND_LOG"
printf '#!/bin/sh\nexit 0\n' > "$FAKE_BIN/codegraph"
/bin/chmod +x "$FAKE_BIN/codegraph"
`)
	t.Setenv("PATH", binDir)
	t.Setenv("COMMAND_LOG", commandLog)
	t.Setenv("FAKE_BIN", binDir)

	if _, err := install.ApplyExtras(map[string]bool{"codegraph": true}, assets, target); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(commandLog)
	if err != nil {
		t.Fatal(err)
	}
	args := strings.Fields(string(raw))
	want := []string{"install", "--global", "@colbymchenry/codegraph@latest"}
	if strings.Join(args, " ") != strings.Join(want, " ") {
		t.Errorf("unexpected npm arguments: got %v, want %v", args, want)
	}
}
