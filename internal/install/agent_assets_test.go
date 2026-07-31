package install

import (
	assetfs "angel-ai-opencode/internal/assets"
	"angel-ai-opencode/internal/catalog"
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

// Vendored OpenSpec skills remain complete and pinned to their generated
// version.
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
