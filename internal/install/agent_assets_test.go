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

func orchestratorSection(t *testing.T, startHeading, endHeading string) string {
	t.Helper()
	orchestrator := readRepositoryAsset(t, "agents", "angel-orchestrator.md")
	start := strings.Index(orchestrator, startHeading)
	if start < 0 {
		t.Fatalf("orchestrator section %q is missing", startHeading)
	}
	if endHeading == "" {
		return orchestrator[start:]
	}
	endOffset := strings.Index(orchestrator[start+len(startHeading):], endHeading)
	if endOffset < 0 {
		t.Fatalf("orchestrator section %q has no following %q section", startHeading, endHeading)
	}
	end := start + len(startHeading) + endOffset
	return orchestrator[start:end]
}

func orchestratorBootstrapSection(t *testing.T) string {
	t.Helper()
	return orchestratorSection(t,
		"### Bootstrap gate before OpenSpec workers",
		"### Workers and their official skills",
	)
}

func TestOrchestratorExecutionRouteOrderingContract(t *testing.T) {
	orchestrator := readRepositoryAsset(t, "agents", "angel-orchestrator.md")
	requireTextInOrder(t, orchestrator,
		"Route the confirmed Brief through OpenSpec or Direct execution as selected.",
		"Keep the confirmed Brief route-neutral.",
		"## Execution route selection",
		"Ask ONE single-select `question`: **OpenSpec** / **Direct**.",
		"**OpenSpec branch boundary:** Only after OpenSpec is selected",
		"**Direct branch boundary:** Only after Direct is selected",
		"## OpenSpec workflow",
	)

	section := orchestratorSection(t, "## Execution route selection", "## OpenSpec workflow")
	requireTextInOrder(t, section,
		"Do not run OpenSpec bootstrap, invoke the OpenSpec CLI, dispatch an OpenSpec worker, or create an OpenSpec change or artifact before this choice.",
		"Ask ONE single-select `question`: **OpenSpec** / **Direct**.",
	)
}

func TestOrchestratorDirectRoutingContract(t *testing.T) {
	section := orchestratorSection(t, "## Execution route selection", "## OpenSpec workflow")

	t.Run("recommendation is risk based and non-binding", func(t *testing.T) {
		requireTextInOrder(t, section,
			"For a clear, isolated, reversible change, recommend **Direct**.",
			"For architecture, security, data, migrations, cross-cutting scope, or material uncertainty, recommend **OpenSpec**.",
			"The recommendation is non-binding: accept either route, and treat the user's selection as authoritative.",
		)
	})

	t.Run("mode choice routes implementation only to general", func(t *testing.T) {
		requireTextInOrder(t, section,
			"**Direct branch boundary:** Only after Direct is selected",
			"ask ONE single-select `question`: **Safe** / **Fast**",
			"pass the confirmed Brief verbatim to the bounded `general` implementation worker",
			"Do not pass it to `openspec-planner`.",
			"Both modes dispatch exactly one `general` worker to implement the bounded work.",
			"Never implement Direct work inline or delegate it to `openspec-implementer` or any other OpenSpec worker.",
		)
	})

	t.Run("existing targets require successful fresh status", func(t *testing.T) {
		requireTextInOrder(t, section,
			"First determine whether the request targets an existing OpenSpec change.",
			"do not offer or use Direct execution",
			"run `openspec status --change <name> --json`",
			"only when that fresh command succeeds and resolves the referenced existing change",
			"If the target is missing, stale, or otherwise unresolvable, report that target and stop for user direction.",
			"Do not offer or infer Direct execution as a fallback.",
		)
	})

	t.Run("worker prompt is bounded and auditable", func(t *testing.T) {
		requireTextInOrder(t, section,
			"Pass the confirmed Brief verbatim",
			"the selected Safe or Fast mode",
			"explicit scope limits",
			"Require this return contract",
			"status (`done`, `partial`, or `blocked`)",
			"files touched",
			"commands run with exit codes",
			"deviations from the Brief or scope",
		)
		requireTextInOrder(t, section,
			"Mode obligations:",
			"Safe: implement the bounded Brief and run the repository's existing applicable tests and build commands.",
			"Fast: implement only the bounded Brief. Do not run tests or reviews.",
		)
	})

	t.Run("direct mode excludes openspec", func(t *testing.T) {
		requireTextInOrder(t, section,
			"Direct mode MUST NOT run OpenSpec bootstrap, invoke the OpenSpec CLI, or create or modify OpenSpec artifacts.",
			"Direct mode MUST NOT invoke OpenSpec verification or archive behavior.",
			"Do not delegate Direct implementation to the orchestrator, `openspec-implementer`, or any other OpenSpec worker; only `general` may implement it.",
		)
	})
}

func TestOrchestratorSafeDirectExecutionContract(t *testing.T) {
	section := orchestratorSection(t, "### Safe direct execution", "### Fast direct execution")

	t.Run("same worker implements and verifies", func(t *testing.T) {
		requireTextInOrder(t, section,
			"The same `general` worker MUST implement the bounded Brief and run the repository's existing applicable tests and build commands.",
			"Treat Safe as clean only when",
			"executable verification was available and run",
			"the worker reports the executable test/build commands and exit codes",
			"the result is clean under the shared implementation-result policy",
			"Only after a clean Safe result proceed to the direct Safe review gate.",
		)
	})

	t.Run("unavailable or omitted verification stops without fallback", func(t *testing.T) {
		requireTextInOrder(t, section,
			"If executable verification is unavailable or its command/exit-code evidence is omitted",
			"report the result as not verified and stop without retrying, dispatching a fallback worker, opening reviews, or continuing implementation.",
			"Unavailable verification must be reported as `partial` or `blocked`.",
			"Apply the shared mandatory-stop policy to every other unsafe result.",
		)
	})
}

func TestOrchestratorCorrectedIntermediateFailureContract(t *testing.T) {
	section := orchestratorSection(t, "### Shared implementation-result policy", "### Safe direct execution")
	requireTextInOrder(t, section,
		"Apply this policy to every planned OpenSpec implementation batch, initial Direct Safe implementation result, and bounded Direct Safe review-fix result.",
		"An intermediate non-zero command is corrected only when",
		"the same worker identifies the failure",
		"later runs an equivalent or broader relevant command with exit code zero",
		"final status `done`",
		"no deviation or out-of-scope work",
		"The successful command MUST validate the failed command's relevant scope or a superset of it.",
		"Retain and surface the failed command, its exit code, and the successful rerun",
	)

	t.Run("route-specific sections reference the shared policy", func(t *testing.T) {
		requireTextInOrder(t,
			orchestratorSection(t, "### Safe direct execution", "### Fast direct execution"),
			"clean under the shared implementation-result policy",
			"Apply the shared mandatory-stop policy",
		)
		requireTextInOrder(t,
			orchestratorSection(t, "### Direct review gate", "## Delegation rules"),
			"clean under the shared implementation-result policy",
			"Apply the shared mandatory-stop policy",
		)
		requireTextInOrder(t,
			orchestratorSection(t, "### Implementation stops and completion routing", "### Between phases"),
			"apply the shared implementation-result policy",
			"On a shared mandatory stop",
			"continue or pause under the already selected cadence",
		)
	})
}

func TestOrchestratorMandatoryImplementationStopsContract(t *testing.T) {
	section := orchestratorSection(t, "### Shared implementation-result policy", "### Safe direct execution")
	conditions := []string{
		"a non-zero command has no later equivalent-or-broader relevant command exiting zero",
		"the final relevant verification state is red",
		"status is `partial` or `blocked`",
		"the worker reports a deviation",
		"the worker reports out-of-scope work",
		"a later successful command is unrelated to or narrower than the failed command's relevant scope",
		"a TDD or expected failure remains red at batch end",
	}

	for _, condition := range conditions {
		t.Run(condition, func(t *testing.T) {
			requireTextInOrder(t, section, "A mandatory stop applies when any of these is true:", condition)
		})
	}
	requireTextInOrder(t, section,
		"On a mandatory stop, surface the evidence",
		"do not retry, dispatch a fallback worker, continue implementation, or advance to the route's next phase.",
	)
}

func TestOrchestratorFastDirectExecutionContract(t *testing.T) {
	section := orchestratorSection(t, "### Fast direct execution", "### Direct review gate")
	requireTextInOrder(t, section,
		"The `general` worker implements only the bounded Brief.",
		"It MUST NOT run tests or reviews.",
		"Report the result explicitly as implemented but not verified",
		"do not open the direct review gate",
	)
}

func TestOrchestratorDirectReviewContract(t *testing.T) {
	section := orchestratorSection(t, "### Direct review gate", "## OpenSpec workflow")

	t.Run("runs only selected bounded reviews", func(t *testing.T) {
		requireTextInOrder(t, section,
			"Only after a clean Safe result",
			"**Security risk** / **Simplicity** / **Correctness** / **None**",
			"**None** is mutually exclusive",
			"If a response mixes **None** with any reviewer, reject it and re-prompt the same review question.",
			"Run only the selected reviewers against the bounded direct diff and confirmed Brief.",
			"Deduplicate their findings and ask the user which findings to fix.",
		)
	})

	t.Run("fixes only selected findings through general", func(t *testing.T) {
		requireTextInOrder(t, section,
			"Only user-selected findings become work.",
			"Send exactly those findings together as one bounded fix batch to `general`",
			"the same structured result contract",
			"MUST NOT use `openspec-implementer`",
			"must run the existing applicable tests and build commands and return their executable command/exit-code evidence",
			"Unavailable or omitted verification means the fix is not verified",
			"report and stop without retrying or rerunning a reviewer",
		)
	})

	t.Run("reruns only affected reviewers", func(t *testing.T) {
		requireTextInOrder(t, section,
			"After a clean fix",
			"rerun only reviewers responsible for the addressed selected findings",
			"Do not invoke OpenSpec verification or archive behavior.",
		)
	})
}

func TestOrchestratorOpenSpecBranchReachesWorkflowBoundary(t *testing.T) {
	routeSection := orchestratorSection(t, "## Execution route selection", "## OpenSpec workflow")
	requireTextInOrder(t, routeSection,
		"**OpenSpec branch boundary:** Only after OpenSpec is selected, enter `## OpenSpec workflow`.",
		"Pass the confirmed Brief verbatim to `openspec-planner` only when dispatching that worker after the required OpenSpec bootstrap succeeds.",
		"Do not pass it to a Direct `general` implementation worker.",
	)

	workflow := orchestratorSection(t, "## OpenSpec workflow", "## Language")
	requireTextInOrder(t, workflow,
		"Enter this workflow boundary only after the user selects OpenSpec for new work, or after fresh successful status resolution of a referenced existing change.",
		"The OpenSpec branch preserves the bootstrap gate, official planner and artifact lifecycle, bounded implementation cadence, verification policy, review gate and review-fix routing, and archive path below.",
	)

	t.Run("retains openspec review-fix routing", func(t *testing.T) {
		requireTextInOrder(t, workflow,
			"**Review-fix routing:** Only findings the user selects become a task",
			"delegate them to `openspec-implementer` as one bounded batch",
			"This finding-ID batch is outside the planned-task cadence",
			"Never delegate a fix for an unselected or SUGGESTION-only finding",
			"re-run only the reviewers whose findings were addressed",
			"otherwise proceed to archive",
		)
	})
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
