package install

import (
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

func TestReviewerAssetsShareReadOnlyValidationContract(t *testing.T) {
	names := []string{
		"review-correctness",
		"review-security-risk",
		"review-simplicity",
	}
	const permissionContract = `permission:
  bash:
    "*": "allow"
    "git add*": "deny"
    "git commit*": "deny"
    "git push*": "deny"
  edit: "deny"
  write: "deny"
  read:
    "*": "allow"
    "*.env": "deny"
    "*.env.*": "deny"
    "*.key": "deny"
    "*.pem": "deny"
    ".aws/credentials": "deny"
    ".config/gh/hosts.yml": "deny"
    ".credentials/**": "deny"
    ".ssh/**": "deny"
    "Library/Keychains/**": "deny"
    "credentials.json": "deny"
    "secrets/**": "deny"
    "**/*.key": "deny"
    "**/*.pem": "deny"
    "**/.aws/credentials": "deny"
    "**/.config/gh/hosts.yml": "deny"
    "**/.credentials/**": "deny"
    "**/.env": "deny"
    "**/.env.*": "deny"
    "**/.ssh/**": "deny"
    "**/Library/Keychains/**": "deny"
    "**/credentials.json": "deny"
    "**/secrets/**": "deny"
    ".env.example": "allow"
    "**/.env.example": "allow"
    ".env.template": "allow"
    "**/.env.template": "allow"`
	const behaviorContract = `You may use Bash to inspect Git state, read or search non-secret repository
files, and run tests or linters. Those validation commands may use the network,
local services, or local artifacts. Remain read-only: never alter tracked files,
stage, commit, push, or read secrets. Do not use Bash indirection or wrappers to
bypass these limits; native permissions are not a complete sandbox.`

	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			asset := readRepositoryAsset(t, "agents", name+".md")
			frontmatter, body, found := strings.Cut(strings.TrimPrefix(asset, "---\n"), "---\n")
			if !found {
				t.Fatal("agent frontmatter is missing")
			}
			for _, tool := range []string{
				"bash: true",
				"edit: false",
				"read: true",
				"write: false",
				"task: false",
			} {
				if !strings.Contains(frontmatter, tool) {
					t.Errorf("missing tool contract %q", tool)
				}
			}
			permissionStart := strings.Index(frontmatter, "permission:\n")
			if permissionStart < 0 {
				t.Fatal("permission contract is missing")
			}
			if got := strings.TrimSpace(frontmatter[permissionStart:]); got != permissionContract {
				t.Errorf("permission contract differs:\n%s", got)
			}
			if !strings.Contains(body, behaviorContract) {
				t.Error("allowed and denied reviewer behavior contract is missing")
			}
			requireTextInOrder(t, body,
				"Include a **Validation evidence** section",
				"every validation command you actually ran",
				"its exit code",
				"with findings or `No findings.`",
				"Report non-zero exits without modifying files or attempting a fix.",
			)
		})
	}
}

func TestSelectedReviewerAssetsAreCatalogedAndInstalledUnchanged(t *testing.T) {
	assets := filepath.Join("..", "..", "assets")
	categories, err := catalog.Load(assets)
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]string{
		"review-correctness":   "review-correctness.md",
		"review-security-risk": "review-security-risk.md",
		"review-simplicity":    "review-simplicity.md",
	}
	var selected []catalog.Item
	for _, category := range categories {
		if category.Name != "agents" {
			continue
		}
		for _, item := range category.Items {
			fileName, ok := want[item.Name]
			if !ok {
				continue
			}
			if item.Kind != catalog.CopyFile || item.Source != filepath.Join(assets, "agents", fileName) || item.Dest != filepath.Join("agents", fileName) {
				t.Fatalf("catalog item %q = %#v", item.Name, item)
			}
			selected = append(selected, item)
		}
	}
	if len(selected) != len(want) {
		t.Fatalf("selected reviewer assets = %v, want %d", selected, len(want))
	}

	configDir := t.TempDir()
	if _, err := ApplyInstallation(InstallationRequest{
		Items: selected, AssetsDir: assets, ConfigDir: configDir,
	}); err != nil {
		t.Fatal(err)
	}
	for _, item := range selected {
		want, err := os.ReadFile(item.Source)
		if err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(filepath.Join(configDir, item.Dest))
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("installed %s differs from selected asset", item.Name)
		}
	}
}
