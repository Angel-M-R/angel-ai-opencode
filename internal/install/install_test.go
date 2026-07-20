package install_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

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
	if _, err := install.Apply(items, target); err != nil {
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
}
