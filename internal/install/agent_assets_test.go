package install

import (
	assetfs "angel-ai-opencode/internal/assets"
	"angel-ai-opencode/internal/catalog"
	"os"
	"path/filepath"
	"reflect"
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

// Every Angel AI agent asset is cataloged as a plain file copy and installed
// byte-identical to the repository source.
func TestAgentAssetsAreCatalogedAndInstalledUnchanged(t *testing.T) {
	assetsRoot := filepath.Join("..", "..", "assets")
	assetSource := assetfs.Directory(assetsRoot)
	categories, err := catalog.Load(assetSource)
	if err != nil {
		t.Fatal(err)
	}

	items := map[string]catalog.Item{}
	for _, category := range categories {
		if category.Name != "agents" {
			continue
		}
		for _, item := range category.Items {
			items[item.Name] = item
		}
	}

	var selected []catalog.Item
	for _, name := range ConfigurableAgents() {
		item, ok := items[name]
		if !ok {
			t.Fatalf("agent %q is not cataloged", name)
		}
		fileName := name + ".md"
		if item.Kind != catalog.CopyFile || item.Source != "agents/"+fileName || item.Dest != filepath.Join("agents", fileName) {
			t.Fatalf("catalog item %q = %#v", name, item)
		}
		selected = append(selected, item)
	}

	configDir := t.TempDir()
	if _, err := ApplyInstallation(InstallationRequest{
		Items: selected, Assets: assetSource, ConfigDir: configDir,
	}); err != nil {
		t.Fatal(err)
	}
	for _, item := range selected {
		want, err := assetSource.ReadFile(item.Source)
		if err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(filepath.Join(configDir, item.Dest))
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("installed %s differs from the repository asset", item.Name)
		}
	}
}

// Agent frontmatter stays structurally safe without pinning prompt prose: it
// parses, declares no deprecated variant, keeps a permission block, read-only
// agents keep edit and write disabled, and the verifier keeps default-deny
// bash.
func TestAgentFrontmatterRemainsStructurallySafe(t *testing.T) {
	readOnly := map[string]bool{
		"openspec-verifier":    true,
		"review-correctness":   true,
		"review-security-risk": true,
		"review-simplicity":    true,
	}
	for _, name := range ConfigurableAgents() {
		t.Run(name, func(t *testing.T) {
			asset := readRepositoryAsset(t, "agents", name+".md")
			frontmatter, _, found := strings.Cut(strings.TrimPrefix(asset, "---\n"), "---\n")
			if !found {
				t.Fatal("agent frontmatter is missing")
			}
			for _, line := range strings.Split(frontmatter, "\n") {
				if strings.HasPrefix(strings.TrimSpace(line), "variant:") {
					t.Errorf("frontmatter still declares a variant: %q", strings.TrimSpace(line))
				}
			}
			if !strings.Contains(frontmatter, "permission:") {
				t.Error("frontmatter lost its permission block")
			}
			if readOnly[name] {
				for _, contract := range []string{"edit: false", "write: false"} {
					if !strings.Contains(frontmatter, contract) {
						t.Errorf("read-only agent lost %q", contract)
					}
				}
			}
			if name == "openspec-verifier" && !strings.Contains(frontmatter, `"*": "deny"`) {
				t.Error("verifier lost its default-deny bash permission")
			}
		})
	}
}

func TestOpenSpecPlannerRequiresEvidenceCompleteResults(t *testing.T) {
	planner := strings.ToLower(strings.Join(strings.Fields(readRepositoryAsset(t, "agents", "openspec-planner.md")), " "))
	for _, required := range []string{
		"audit the complete tool transcript",
		"files touched",
		"every command executed in exact order",
		"failed command and exit code",
		"diagnosed cause",
		"bounded correction",
		"equivalent-or-broader",
		"evidence that the successful validation covers",
		"same worker in the same bounded invocation",
		"final relevant validation state",
		"evidence gap",
		"do not report `done` or recommend implementation",
	} {
		if !strings.Contains(planner, required) {
			t.Errorf("planner result contract is missing %q", required)
		}
	}
	if strings.Contains(planner, "return a compact result: status (done|blocked|partial)") {
		t.Error("planner still has the incomplete compact-only result contract")
	}
}

func TestOpenSpecVerifierDocumentsCompletionEvidenceSchema(t *testing.T) {
	verifier := readRepositoryAsset(t, "agents", "openspec-verifier.md")
	for _, required := range []string{
		`"task": {"id": "4.1", "text": "4.1 Run focused verification"}`,
		`"status": "pass"`,
		`"commands": [`,
		`"command": ["go", "test", "./internal/install", "-run", "TestOpenSpecVerifierDocumentsCompletionEvidenceSchema"]`,
		`"exitCode": 0`,
	} {
		if !strings.Contains(verifier, required) {
			t.Errorf("verifier completion evidence schema is missing %q", required)
		}
	}
	if strings.Contains(verifier, `"argv"`) {
		t.Error("verifier completion command record defines an argv field")
	}
}

func TestOrchestratorGatesImplementationOnPlannerEvidence(t *testing.T) {
	orchestrator := strings.ToLower(strings.Join(strings.Fields(readRepositoryAsset(t, "agents", "angel-orchestrator.md")), " "))
	for _, required := range []string{
		"### Planner result gate",
		"the planner owns the detailed evidence report; the orchestrator owns the gate",
		"artifacts exist is not sufficient",
		"clean under the shared corrected-failure result fields",
		"treat the result as an evidence gap",
		"never dispatch an implementer from an incomplete planning report",
	} {
		if !strings.Contains(orchestrator, strings.ToLower(required)) {
			t.Errorf("orchestrator planner gate is missing %q", required)
		}
	}
}

func TestOrchestratorRequiresPreBriefSolutionComparison(t *testing.T) {
	orchestrator := strings.ToLower(strings.Join(strings.Fields(readRepositoryAsset(t, "agents", "angel-orchestrator.md")), " "))
	for _, required := range []string{
		"### solution comparison gate",
		"after the selected product/technical interview work",
		"before the brief is complete",
		"briefly inspect the relevant repository",
		"compare 2-3 viable alternatives when that many exist",
		"never invent alternatives",
		"if only one option is viable",
		"complexity",
		"risk",
		"guarantee",
		"operational impact",
		"reversibility",
		"scope change",
		"state one recommendation",
		"explicit selection",
		"separate solution-choice `question`",
		"separate from the existing route-selection question",
		"materially changes scope",
		"before that choice",
		"preserve the repository evidence, full matrix, recommendation",
		"pass them verbatim in the brief to `openspec-planner`",
	} {
		if !strings.Contains(orchestrator, required) {
			t.Errorf("orchestrator solution-comparison gate is missing %q", required)
		}
	}
	if strings.Contains(orchestrator, "/grill-me") {
		t.Error("orchestrator introduced the out-of-scope /grill-me path")
	}

	gate := strings.Index(orchestrator, "### solution comparison gate")
	brief := strings.Index(orchestrator, "the interview ends with a completed brief")
	route := strings.Index(orchestrator, "## execution route selection")
	if gate < 0 || brief < 0 || route < 0 || !(brief < gate && gate < route) {
		t.Fatalf("solution-comparison gate ordering is invalid: brief=%d gate=%d route=%d", brief, gate, route)
	}
}

func TestOrchestratorSupportsManualReviewWithSharedReviewerSelection(t *testing.T) {
	orchestrator := strings.ToLower(strings.Join(strings.Fields(readRepositoryAsset(t, "agents", "angel-orchestrator.md")), " "))
	for _, required := range []string{
		"manual review request",
		"explicit user request to review the current state",
		"planned tasks remain pending",
		"before `openspec-verifier`",
		"same one multi-select `question`",
		"security risk",
		"simplicity",
		"correctness",
		"none",
		"mutually exclusive",
		"no reviewer option is preselected",
		"do not infer the reviewer selection",
		"reviewed, not verified",
		"must not mark or unmark openspec tasks",
		"satisfy the verifier gate",
		"archive a change",
		"do not trigger verification or archive automatically",
	} {
		if !strings.Contains(orchestrator, required) {
			t.Errorf("orchestrator manual-review contract is missing %q", required)
		}
	}

	manual := strings.Index(orchestrator, "### manual review request")
	automatic := strings.Index(orchestrator, "### automatic review gate")
	if manual < 0 || automatic < 0 || manual >= automatic {
		t.Fatalf("manual review entry point must precede the automatic gate: manual=%d automatic=%d", manual, automatic)
	}
	if !strings.Contains(orchestrator[automatic:], "once `openspec-verifier` reports the change verified") {
		t.Error("automatic OpenSpec review gate no longer remains verifier-gated")
	}
}

func TestOpenSpecBootstrapDelegatesProjectIntegrationToOpenSpec(t *testing.T) {
	orchestrator := readRepositoryAsset(t, "agents", "angel-orchestrator.md")
	normalized := strings.Join(strings.Fields(orchestrator), " ")
	for _, required := range []string{
		"openspec init --tools opencode",
		"openspec update",
		"current global OpenSpec profile, workflow, and delivery configuration",
		"context-and-skill keys",
		"<required-skill>/SKILL.md",
	} {
		if !strings.Contains(normalized, required) {
			t.Errorf("OpenSpec bootstrap contract is missing %q", required)
		}
	}
	for _, forbidden := range []string{
		"openspec init --tools none",
		"metadata.generatedBy",
		"~/.config/opencode/skills/openspec/",
		"Never run `openspec update`",
	} {
		if strings.Contains(orchestrator, forbidden) {
			t.Errorf("OpenSpec bootstrap retains obsolete contract %q", forbidden)
		}
	}
}

// Installer reconciliation propagates an updated repository asset only when the
// agent is selected again, and never mutates an existing configured copy on its
// own.
func TestExistingOrchestratorCopyRequiresUpdatedSourceAndSelectedReconciliation(t *testing.T) {
	assetRoot := t.TempDir()
	agentsDir := filepath.Join(assetRoot, "agents")
	if err := os.Mkdir(agentsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	assetPath := filepath.Join(agentsDir, "angel-orchestrator.md")
	original := []byte("original orchestrator policy\n")
	updated := []byte("updated recovered-probe policy\n")
	if err := os.WriteFile(assetPath, original, 0o644); err != nil {
		t.Fatal(err)
	}

	assetSource := assetfs.Directory(assetRoot)
	categories, err := catalog.Load(assetSource)
	if err != nil {
		t.Fatal(err)
	}
	var selected catalog.Item
	for _, category := range categories {
		if category.Name == "agents" && len(category.Items) == 1 {
			selected = category.Items[0]
		}
	}
	if selected.Name != "angel-orchestrator" {
		t.Fatalf("selected agent = %#v", selected)
	}

	configDir := t.TempDir()
	reconcile := func() {
		t.Helper()
		if _, err := ApplyInstallation(InstallationRequest{
			Items: []catalog.Item{selected}, Assets: assetSource, ConfigDir: configDir,
		}); err != nil {
			t.Fatal(err)
		}
	}
	reconcile()
	installedPath := filepath.Join(configDir, selected.Dest)

	// Repository edits similarly do not mutate an existing configured copy. A
	// rebuild or explicit updated asset source must still be followed by
	// reconciliation with angel-orchestrator selected.
	if err := os.WriteFile(assetPath, updated, 0o644); err != nil {
		t.Fatal(err)
	}
	installed, err := os.ReadFile(installedPath)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(installed, original) {
		t.Fatal("configured copy changed before selected-agent reconciliation")
	}

	reconcile()
	installed, err = os.ReadFile(installedPath)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(installed, updated) {
		t.Fatal("selected-agent reconciliation did not propagate the updated source")
	}
}
