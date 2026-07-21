package install

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func readRepositoryAsset(t *testing.T, elements ...string) string {
	t.Helper()
	path := filepath.Join(append([]string{"..", "..", "assets"}, elements...)...)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func normalizedText(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func requireTextInOrder(t *testing.T, value string, snippets ...string) {
	t.Helper()
	remaining := normalizedText(value)
	for _, snippet := range snippets {
		normalizedSnippet := normalizedText(snippet)
		position := strings.Index(remaining, normalizedSnippet)
		if position < 0 {
			t.Fatalf("missing or out-of-order contract text %q", normalizedSnippet)
		}
		remaining = remaining[position+len(normalizedSnippet):]
	}
}

func orchestratorBootstrapSection(t *testing.T) string {
	t.Helper()
	orchestrator := readRepositoryAsset(t, "agents", "angel-orchestrator.md")
	start := strings.Index(orchestrator, "### Bootstrap gate before OpenSpec workers")
	end := strings.Index(orchestrator, "### Workers and their official skills")
	if start < 0 || end <= start {
		t.Fatal("orchestrator OpenSpec bootstrap section is missing or misplaced")
	}
	return orchestrator[start:end]
}

func TestOrchestratorOpenSpecBootstrapContract(t *testing.T) {
	section := orchestratorBootstrapSection(t)

	t.Run("gates workers before dispatch", func(t *testing.T) {
		requireTextInOrder(t, section,
			"Before dispatching `openspec-planner`, `openspec-implementer`, or `openspec-verifier`",
			"dispatch one short `general` task",
			"wait for it to succeed",
			"Add the returned context key to the set only after success",
			"If bootstrap blocks or fails, do not launch the OpenSpec worker",
			"Only then dispatch the requested OpenSpec worker",
		)
	})

	t.Run("reuses only a successful matching context", func(t *testing.T) {
		requireTextInOrder(t, section,
			"session-only set of successfully bootstrapped OpenSpec context keys",
			"Never persist this cache",
			"If the exact context key is already in the successful set, skip bootstrap",
			"A different project root or store is a different context and MUST be bootstrapped",
		)
	})

	t.Run("uses store-aware list without local initialization", func(t *testing.T) {
		requireTextInOrder(t, section,
			"An explicit registered store uses `store:<id>` as its context key",
			"For an explicit registered store <id>, run `openspec list --json --store <id>`",
			"Never initialize for an explicit store",
		)
		if strings.Contains(section, "openspec init --store") {
			t.Fatal("bootstrap must not initialize a store context")
		}
	})

	t.Run("initializes an unresolved local root once and rechecks JSON", func(t *testing.T) {
		requireTextInOrder(t, section,
			"Otherwise run `openspec list --json` in the requested working directory",
			"For a local context only, when the first list JSON has no resolvable root",
			"`openspec init --tools none`, then run `openspec list --json` once more",
			"Run initialization at most once",
			"This is the only permitted mutation",
		)
		if count := strings.Count(section, "`openspec init --tools none`"); count != 1 {
			t.Fatalf("init command contract occurs %d times, want 1", count)
		}
	})

	t.Run("blocks missing CLI but continues after version drift", func(t *testing.T) {
		requireTextInOrder(t, section,
			"If `openspec` cannot be executed, block",
			"install it with this repository installer's `OpenSpec` extra",
			"Run `openspec --version` and compare it with the child `metadata.generatedBy` values",
			"`~/.config/opencode/skills/openspec/<skill-name>/SKILL.md`",
			"If they differ, report an advisory warning but continue",
			"If local OpenSpec skills duplicate global skills, stay silent",
			"Never run `openspec update`",
			"Do not generate local skills or change OpenSpec profile, workflow, or delivery configuration",
		)
	})
}

func TestVendoredOpenSpecAgentAssetsRemainPreserved(t *testing.T) {
	skillsRoot := filepath.Join("..", "..", "assets", "skills", "openspec")
	entries, err := os.ReadDir(skillsRoot)
	if err != nil {
		t.Fatal(err)
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "openspec-") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	want := []string{
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
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("vendored OpenSpec skills = %v, want %v", names, want)
	}
	for _, name := range names {
		content, err := os.ReadFile(filepath.Join(skillsRoot, name, "SKILL.md"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(content), "generatedBy: \"1.6.0\"") {
			t.Errorf("%s lost its vendored generatedBy contract", name)
		}
	}
}
