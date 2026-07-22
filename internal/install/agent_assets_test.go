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
		"Present the Brief with the combined new-work choice and route it through the selected execution path.",
		"The interview ends with a draft Brief",
		"do not ask a separate confirmation question.",
		"Keep the Brief route-neutral.",
		"## Execution route selection",
		"Present the Brief and ask the combined question below",
		"Selecting **Direct Safe**, **Direct Fast**, or **OpenSpec** implicitly confirms the presented Brief",
		"**OpenSpec branch boundary:** Only after OpenSpec is selected",
		"**Direct branch boundary:** Only after **Direct Safe** or **Direct Fast** is selected",
		"## OpenSpec workflow",
	)

	section := orchestratorSection(t, "## Execution route selection", "## OpenSpec workflow")
	requireTextInOrder(t, section,
		"do not ask a separate Brief confirmation, route, or Direct mode question.",
		"Do not run OpenSpec bootstrap, invoke the OpenSpec CLI, dispatch an OpenSpec worker, or create an OpenSpec change or artifact before this choice.",
		"For new work, give a risk-based recommendation from the Brief and order the single-select `question` choices accordingly:",
	)
}

func TestOrchestratorDirectRoutingContract(t *testing.T) {
	section := orchestratorSection(t, "## Execution route selection", "## OpenSpec workflow")

	t.Run("recommendation is risk based and non-binding", func(t *testing.T) {
		requireTextInOrder(t, section,
			"For a clear, isolated, reversible change, order the choices **Direct Safe (Recommended)** / **Direct Fast** / **OpenSpec** / **Modify Brief**.",
			"For architecture, security, data, migrations, cross-cutting scope, or material uncertainty, order the choices **OpenSpec (Recommended)** / **Direct Safe** / **Direct Fast** / **Modify Brief**.",
			"The recommendation is non-binding: accept any of the three execution routes, and treat the user's selection as authoritative.",
			"Never recommend **Direct Fast** by default.",
			"Keep the `question` tool's custom response available.",
		)
	})

	t.Run("combined choice confirms or reopens the brief", func(t *testing.T) {
		requireTextInOrder(t, section,
			"Selecting **Direct Safe**, **Direct Fast**, or **OpenSpec** implicitly confirms the presented Brief; do not ask for separate confirmation.",
			"Selecting **Modify Brief** does not confirm it",
			"reopen the interview, update the Brief from the user's answers, reassess risk, and present this same combined choice again.",
		)
	})

	t.Run("direct selection routes implementation only to general", func(t *testing.T) {
		requireTextInOrder(t, section,
			"**Direct branch boundary:** Only after **Direct Safe** or **Direct Fast** is selected",
			"use its Safe or Fast mode",
			"pass the confirmed Brief verbatim to the bounded `general` implementation worker",
			"Do not ask another route or mode question.",
			"Do not pass it to `openspec-planner`.",
			"Both modes dispatch exactly one `general` worker to implement the bounded work.",
			"Never implement Direct work inline or delegate it to `openspec-implementer` or any other OpenSpec worker.",
		)
	})

	if strings.Contains(section, "Ask ONE single-select `question`: **OpenSpec** / **Direct**.") {
		t.Fatal("new work must not use a separate route question")
	}
	if strings.Contains(section, "ask ONE single-select `question`: **Safe** / **Fast**") {
		t.Fatal("Direct work must not use a separate mode question")
	}

	t.Run("existing targets require successful fresh status", func(t *testing.T) {
		requireTextInOrder(t, section,
			"First determine whether the request targets an existing OpenSpec change.",
			"do not offer or use Direct execution",
			"run `openspec status --change <name> --json`",
			"only when that fresh command succeeds and resolves the referenced existing change",
			"If the target is missing, stale, or otherwise unresolvable",
			"retain and report the target-resolution command, exit code, and diagnostic",
			"apply the shared mandatory-stop policy",
			"Do not offer or infer Direct execution as a fallback or select substitute work before the user chooses an action.",
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
			"Only after a clean Safe result proceed to the direct Safe review gate",
		)
	})

	t.Run("unavailable or omitted verification stops without fallback", func(t *testing.T) {
		requireTextInOrder(t, section,
			"If executable verification is unavailable or its command/exit-code evidence is omitted",
			"retain the result and report it as not verified with status `partial` or `blocked`",
			"apply the shared mandatory-stop policy",
			"before the user selects an action at a stop, do not retry, dispatch a fallback worker, open reviews, or continue implementation",
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
			"apply the shared mandatory-stop policy",
		)
		requireTextInOrder(t,
			orchestratorSection(t, "### Direct review gate", "## Delegation rules"),
			"clean under the shared implementation-result policy",
			"apply the shared mandatory-stop policy",
		)
		requireTextInOrder(t,
			orchestratorSection(t, "### Implementation stops and completion routing", "### Between phases"),
			"apply the shared implementation-result policy",
			"On a shared mandatory stop",
			"Only after a clean result",
			"apply the fresh-state invariant and continue automatically",
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
		"On every mandatory stop, apply this shared mandatory-stop policy in two ordered, separate steps:",
		"First report the blocking status and all retained evidence needed to choose an action",
		"Then ask exactly one blocker-specific next-action `question`.",
		"Until the user selects an action, do not retry, continue, broaden scope, select substitute work, advance to the route's next phase, or dispatch any worker.",
	)
}

func TestOrchestratorMandatoryStopInteractionContract(t *testing.T) {
	sharedPolicy := orchestratorSection(t, "### Shared implementation-result policy", "### Safe direct execution")

	t.Run("reports evidence before asking one contextual question", func(t *testing.T) {
		requireTextInOrder(t, sharedPolicy,
			"On every mandatory stop, apply this shared mandatory-stop policy in two ordered, separate steps:",
			"First report the blocking status and all retained evidence needed to choose an action",
			"Do not ask the stop question before this report.",
			"Then ask exactly one blocker-specific next-action `question`.",
		)
	})

	t.Run("offers a safe stop and custom response", func(t *testing.T) {
		requireTextInOrder(t, sharedPolicy,
			"Derive its choices from the reported blocker",
			"always include a safe stop option",
			"keep the question tool's custom response available",
		)
	})

	t.Run("forbids recovery and worker dispatch until selection", func(t *testing.T) {
		requireTextInOrder(t, sharedPolicy,
			"Until the user selects an action",
			"do not retry, continue, broaden scope, select substitute work, advance to the route's next phase, or dispatch any worker",
			"Do not infer authorization from the blocker itself.",
		)
	})

	stopRoutes := []struct {
		name         string
		startHeading string
		endHeading   string
		evidence     string
	}{
		{
			name:         "existing target resolution",
			startHeading: "## Execution route selection",
			endHeading:   "### Shared implementation-result policy",
			evidence:     "retain and report the target-resolution command, exit code, and diagnostic",
		},
		{
			name:         "OpenSpec bootstrap",
			startHeading: "### Bootstrap gate before OpenSpec workers",
			endHeading:   "### Workers and their official skills",
			evidence:     "retain and report its status, diagnostic, commands, and exit codes",
		},
		{
			name:         "planned-task implementation",
			startHeading: "### Implementation stops and completion routing",
			endHeading:   "### Between phases",
			evidence:     "reporting the worker and command evidence or state conflict before asking its one next-action question",
		},
		{
			name:         "final verification",
			startHeading: "### Implementation stops and completion routing",
			endHeading:   "### Between phases",
			evidence:     "retain its status, commands, exit codes, and diagnostic",
		},
		{
			name:         "Direct Safe",
			startHeading: "### Safe direct execution",
			endHeading:   "### Fast direct execution",
			evidence:     "retain the result and report it as not verified with status `partial` or `blocked`",
		},
		{
			name:         "Direct Fast",
			startHeading: "### Fast direct execution",
			endHeading:   "### Direct review gate",
			evidence:     "report the retained result and command evidence",
		},
		{
			name:         "Direct review fix",
			startHeading: "### Direct review gate",
			endHeading:   "## Delegation rules",
			evidence:     "retain the fix result and command evidence, report it as `partial` or `blocked`",
		},
	}

	for _, route := range stopRoutes {
		t.Run(route.name+" references shared policy", func(t *testing.T) {
			requireTextInOrder(t,
				orchestratorSection(t, route.startHeading, route.endHeading),
				route.evidence,
				"apply the shared mandatory-stop policy",
			)
		})
	}
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
			"retain the fix result and command evidence, report it as `partial` or `blocked`",
			"apply the shared mandatory-stop policy",
			"Do not retry, broaden the selected finding set, rerun a reviewer, or dispatch another worker before the user selects an action.",
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
		"The OpenSpec branch preserves the bootstrap gate, official planner and artifact lifecycle, bounded automatic implementation, verification policy, review gate and review-fix routing, and archive path below.",
	)

	t.Run("retains openspec review-fix routing", func(t *testing.T) {
		requireTextInOrder(t, workflow,
			"**Review-fix routing:** Only findings the user selects become a task",
			"delegate them to `openspec-implementer` as one bounded batch",
			"This finding-ID batch is outside the automatic planned-task loop",
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
			"retain and report its status, diagnostic, commands, and exit codes",
			"apply the shared mandatory-stop policy",
			"Only after a successful bootstrap may the requested OpenSpec worker be dispatched",
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
