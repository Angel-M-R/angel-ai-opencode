package install_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	assetfs "angel-ai-opencode/internal/assets"
	"angel-ai-opencode/internal/install"
)

func applyAgentModels(t *testing.T, target string, assignments install.AgentModelAssignments) map[string]any {
	t.Helper()
	if _, err := install.ApplyInstallation(install.InstallationRequest{
		Assets:      assetfs.Directory(t.TempDir()),
		ConfigDir:   target,
		AgentModels: assignments,
	}); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(target, "opencode.json"))
	if err != nil {
		t.Fatal(err)
	}
	var config map[string]any
	if err := json.Unmarshal(raw, &config); err != nil {
		t.Fatal(err)
	}
	return config
}

func TestAgentModelsWriteBothKeysAndPreserveEverythingElse(t *testing.T) {
	target := t.TempDir()
	write(t, filepath.Join(target, "opencode.json"), `{
  "$schema": "https://opencode.ai/config.json",
  "theme": "one-dark-pro",
  "agent": {
    "my-own-agent": {"model": "openai/gpt-5"},
    "openspec-verifier": {"variant": "xhigh"}
  }
}
`)

	config := applyAgentModels(t, target, install.AgentModelAssignments{
		"review-simplicity": {Model: "anthropic/claude-sonnet-4-6", Variant: "high"},
		"openspec-verifier": {Model: "openai/gpt-5-mini", Variant: ""},
	})

	if config["theme"] != "one-dark-pro" {
		t.Fatalf("unrelated key lost: %v", config["theme"])
	}
	agents, ok := config["agent"].(map[string]any)
	if !ok {
		t.Fatalf("agent object missing: %v", config["agent"])
	}
	want := map[string]any{
		"my-own-agent":      map[string]any{"model": "openai/gpt-5"},
		"openspec-verifier": map[string]any{"model": "openai/gpt-5-mini", "variant": ""},
		"review-simplicity": map[string]any{"model": "anthropic/claude-sonnet-4-6", "variant": "high"},
	}
	if !reflect.DeepEqual(agents, want) {
		t.Fatalf("agent entries = %v, want %v", agents, want)
	}
}

func TestEmptyAgentModelsLeaveConfigByteIdentical(t *testing.T) {
	target := t.TempDir()
	original := `{
  "$schema": "https://opencode.ai/config.json",
  "agent": {
    "openspec-planner": {"model": "anthropic/claude-opus-4-1", "variant": "high"}
  }
}
`
	write(t, filepath.Join(target, "opencode.json"), original)

	for _, assignments := range []install.AgentModelAssignments{nil, {}} {
		if _, err := install.ApplyInstallation(install.InstallationRequest{
			Assets:      assetfs.Directory(t.TempDir()),
			ConfigDir:   target,
			AgentModels: assignments,
		}); err != nil {
			t.Fatal(err)
		}
		raw, err := os.ReadFile(filepath.Join(target, "opencode.json"))
		if err != nil {
			t.Fatal(err)
		}
		if string(raw) != original {
			t.Fatalf("config rewritten for %v assignments:\n%s", assignments, raw)
		}
	}
}

func TestLoadAgentModelSelectionsRoundTrips(t *testing.T) {
	target := t.TempDir()
	write(t, filepath.Join(target, "opencode.json"), `{
  "agent": {
    "angel-orchestrator": {"model": "anthropic/claude-sonnet-4-6", "variant": "high"},
    "openspec-planner": {"model": "openrouter/anthropic/claude-opus-4-1"},
    "openspec-verifier": {"model": "malformed"},
    "my-own-agent": {"model": "openai/gpt-5", "variant": "low"}
  }
}
`)

	got := install.LoadAgentModelSelections(target)
	want := map[string]install.AgentModelSelection{
		"angel-orchestrator": {Provider: "anthropic", Model: "claude-sonnet-4-6", Effort: "high"},
		"openspec-planner":   {Provider: "openrouter", Model: "anthropic/claude-opus-4-1"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("selections = %v, want %v", got, want)
	}
	if assignment := got["angel-orchestrator"].Assignment(); assignment != (install.AgentModelAssignment{
		Model:   "anthropic/claude-sonnet-4-6",
		Variant: "high",
	}) {
		t.Fatalf("round trip lost the assignment: %v", assignment)
	}
}

func TestHalfBuiltAgentModelsAreNeverWritten(t *testing.T) {
	target := t.TempDir()
	original := `{"$schema": "https://opencode.ai/config.json"}
`
	write(t, filepath.Join(target, "opencode.json"), original)

	if _, err := install.ApplyInstallation(install.InstallationRequest{
		Assets:    assetfs.Directory(t.TempDir()),
		ConfigDir: target,
		AgentModels: install.AgentModelAssignments{
			"review-simplicity":  {Model: "anthropic/"},
			"review-correctness": {Model: "/gpt-5"},
			"openspec-planner":   {Model: "gpt-5"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(target, "opencode.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != original {
		t.Fatalf("malformed assignments reached opencode.json:\n%s", raw)
	}
}

func TestLoadAgentModelSelectionsToleratesMissingAndMalformedConfig(t *testing.T) {
	missing := t.TempDir()
	if got := install.LoadAgentModelSelections(missing); got != nil {
		t.Fatalf("missing config yielded %v", got)
	}

	broken := t.TempDir()
	write(t, filepath.Join(broken, "opencode.json"), "{not json")
	if got := install.LoadAgentModelSelections(broken); got != nil {
		t.Fatalf("malformed config yielded %v", got)
	}
}
