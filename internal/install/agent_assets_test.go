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
		"Present the Brief, then immediately invoke the one route-selection question",
		"The interview ends with a draft Brief",
		"present the completed Brief, then immediately invoke exactly one single-select route-selection `question`",
		"do not ask a separate confirmation question.",
		"Keep the Brief route-neutral.",
		"## Execution route selection",
		"Immediately after presenting it, invoke exactly one single-select route-selection `question`.",
		"The orchestrator owns that question's payload and option order; do not delegate its construction.",
		"Do not ask a separate Brief confirmation, route, or Direct mode question.",
		"Selecting a valid offered **Direct Safe**, **Direct Fast**, or **OpenSpec** route implicitly confirms the presented Brief",
		"**OpenSpec branch boundary:** Only after OpenSpec is selected",
		"**Direct branch boundary:** Only after **Direct Safe** or **Direct Fast** is selected",
		"## OpenSpec workflow",
	)

	section := orchestratorSection(t, "## Execution route selection", "## OpenSpec workflow")
	requireTextInOrder(t, section,
		"Do not ask a separate Brief confirmation, route, or Direct mode question.",
		"Do not run OpenSpec bootstrap, invoke the OpenSpec CLI, dispatch an OpenSpec worker, or create an OpenSpec change or artifact before this choice.",
		"For new work, determine first whether the Brief requires executable validation:",
		"construct the orchestrator-owned single-select `question` payload in this order, keeping its custom response available:",
	)
}

func TestOrchestratorDirectRoutingContract(t *testing.T) {
	section := orchestratorSection(t, "## Execution route selection", "## OpenSpec workflow")

	t.Run("recommendation is risk based and non-binding", func(t *testing.T) {
		requireTextInOrder(t, section,
			"For a clear, isolated, reversible change, order the choices **Direct Safe (Recommended)** / **Direct Fast** / **OpenSpec** / **Modify Brief**.",
			"For architecture, security, data, migrations, cross-cutting scope, or material uncertainty, order the choices **OpenSpec (Recommended)** / **Direct Safe** / **Direct Fast** / **Modify Brief**.",
			"The recommendation is non-binding: accept any valid offered execution route and treat the user's selection as authoritative.",
			"Never recommend **Direct Fast** by default.",
		)
	})

	t.Run("combined choice confirms or reopens the brief", func(t *testing.T) {
		requireTextInOrder(t, section,
			"Selecting a valid offered **Direct Safe**, **Direct Fast**, or **OpenSpec** route implicitly confirms the presented Brief; do not ask for separate confirmation.",
			"Selecting **Modify Brief** does not confirm it",
			"reopen the interview, update the Brief from the user's answers, reassess risk and executable-validation requirements, present the updated Brief, and reissue the route-selection question.",
		)
	})

	t.Run("validation required excludes and rejects Direct Fast", func(t *testing.T) {
		requireTextInOrder(t, section,
			"When the Brief requires executable validation, **Direct Fast** is incompatible.",
			"Omit it from the payload while preserving the applicable risk-based ordering among **Direct Safe**, **OpenSpec**, and **Modify Brief**.",
			"If the user requests **Direct Fast** through a custom response in this state, reject it without confirming the Brief and reissue the same route-selection `question` with **Direct Fast** omitted.",
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
		"This shared strict policy is the default for every implementation result and every OpenSpec control point that invokes it.",
		"The sole route-specific classification exception is inside the automatic planned-task loop",
		"Any result that is not explicitly eligible for that exception remains subject to this strict default.",
		"An intermediate non-zero command caused by command syntax or invocation is a corrected tooling error, rather than a mandatory stop, only when",
		"the same worker identifies that cause",
		"later runs an equivalent-or-broader relevant command with exit code zero",
		"final status `done`",
		"no deviation or out-of-scope work",
		"The successful command MUST cover the failed command's relevant scope or a superset of it.",
		"Retain and surface the original failed command and its exit code together with the later successful command and its exit code",
		"A real verification or implementation failure remains a mandatory stop and is not a corrected tooling error.",
		"another worker performs the rerun",
		"the final status is not `done`",
		"the rerun is unrelated or narrower",
		"the final relevant verification state is red.",
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
			"classify the result under the planned-task exception to the shared implementation-result policy",
			"Every other non-clean result applies the shared strict default.",
			"Stop immediately and dispatch no further batch.",
			"apply the shared mandatory-stop policy",
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
		"For the strict default routes above, every listed condition is a mandatory stop.",
		"Only for an eligible section-bounded planned-task batch",
		"local `partial`, local `blocked`, or red focused test as deferrable",
		"the affected incomplete tasks remain unchecked",
		"Classify an additional read or a successful focused test of modified code as a benign, continuable deviation",
		"These classifications never apply to Direct work, review-fix batches, bootstrap, target resolution, or final verification",
		"On every mandatory stop, apply this shared mandatory-stop policy in two ordered, separate steps:",
		"First report the blocking status and all retained evidence needed to choose an action",
		"Then ask exactly one blocker-specific next-action `question`.",
		"Until the user selects an action, do not retry, continue, broaden scope, select substitute work, advance to the route's next phase, or dispatch any worker.",
	)
}

func TestOrchestratorPlannedBatchDeferralContract(t *testing.T) {
	sharedPolicy := orchestratorSection(t, "### Shared implementation-result policy", "### Safe direct execution")
	plannedLoop := orchestratorSection(t, "### Planned-task implementation state", "### Between phases")

	t.Run("limits eligibility to fresh section bounded planned batches", func(t *testing.T) {
		requireTextInOrder(t, sharedPolicy,
			"only a section-bounded planned OpenSpec task batch selected from the active change's fresh `tasks.md`",
			"Any result that is not explicitly eligible for that exception remains subject to this strict default.",
			"require the same planned-task implementer to diagnose and repair real failures attributable to its bounded changes within the same invocation",
			"A returned attributable failure is not deferrable while another safe bounded repair cycle can make demonstrable progress.",
			"local `partial`, local `blocked`, or red focused test as deferrable",
			"only after required bounded self-repair is exhausted or a pre-existing or unrelated blocker is identified",
			"the affected incomplete tasks remain unchecked and no planned-loop hard blocker exists",
		)
	})

	t.Run("accumulates incidents and benign deviations", func(t *testing.T) {
		requireTextInOrder(t, plannedLoop,
			"**Deferred-evidence record:**",
			"Accumulate one record for every deferrable planned-task incident and every benign continuable deviation.",
			"section and task identifiers",
			"fresh checkbox state",
			"worker status",
			"every command and exit code in execution order",
			"focused-validation state",
			"blocker or incomplete-work reason",
			"files touched, and deviations",
			"Keep the corresponding incomplete tasks unchecked.",
		)
	})

	t.Run("requires affirmative independence", func(t *testing.T) {
		requireTextInOrder(t, plannedLoop,
			"**Conservative independence gate:**",
			"current planning artifacts, the bounded task scopes, or retained worker diagnostics explicitly establish",
			"does not consume, validate, or depend on any deferred work",
			"Section order, different section names, silence, and assumptions are not independence evidence.",
			"Missing, conflicting, or ambiguous evidence means dependency",
			"against every currently deferred batch before each later dispatch",
		)
	})

	t.Run("reports once before one fresh retry round", func(t *testing.T) {
		requireTextInOrder(t, plannedLoop,
			"**Single retry round:**",
			"apply the fresh-state invariant and present one combined report containing all accumulated deferred incidents and benign deviations",
			"Do not ask an intermediate question.",
			"Then run exactly one final retry round.",
			"Before each retry, refresh state and recompute the bounded batch from only its current unchecked tasks",
			"Retry each still-pending deferred batch at most once",
			"Never re-queue a retried batch, create a second deferred queue, or start another retry round.",
		)
		if count := strings.Count(plannedLoop, "Then run exactly one final retry round."); count != 1 {
			t.Fatalf("final retry-round contract occurs %d times, want 1", count)
		}
	})

	t.Run("stops unresolved work before verification", func(t *testing.T) {
		requireTextInOrder(t, plannedLoop,
			"At the end of that one retry round, apply the fresh-state invariant.",
			"any planned task remains unchecked",
			"no retry batch is runnable",
			"a local block is unresolved",
			"focused-validation evidence remains red",
			"stop before final verification",
			"apply the shared mandatory-stop interaction exactly once",
			"Only fresh state with every task complete and no relevant red evidence may enter final verification.",
		)
	})
}

func TestOrchestratorPlannedBatchImplementerContract(t *testing.T) {
	plannedLoop := orchestratorSection(t, "### Automatic planned-task loop and bounded batches", "### Between phases")

	t.Run("keeps focused validation with implementer", func(t *testing.T) {
		requireTextInOrder(t, plannedLoop,
			"MUST require validation relevant to the bounded changes",
			"MAY run focused lint, focused typecheck, and the minimum tests relevant to behavior modified by that batch",
			"When an applicable lint or typecheck tool has no supported filtering mechanism, the implementer MAY run its global non-destructive check",
			"MUST NOT run the full repository test suite or any build",
			"mandatory full suites and builds are reserved for final OpenSpec verification",
		)
	})

	t.Run("self repairs attributable failures while progress continues", func(t *testing.T) {
		requireTextInOrder(t, plannedLoop,
			"within the same worker invocation",
			"Treat a failure as attributable only when it was caused by files or behavior changed for the assigned batch.",
			"continue bounded repair and rerun relevant validation while each cycle makes demonstrable progress",
			"changed diagnostic evidence, a narrower attributable cause, a completed necessary bounded correction, or improved relevant validation",
			"materially the same failure repeats without new progress",
			"report a real blocker with all retained command evidence",
		)
	})

	t.Run("bounds repair writes and supporting adjustments", func(t *testing.T) {
		requireTextInOrder(t, plannedLoop,
			"limit writes to files assigned to the batch",
			"a minimal adjustment elsewhere only when it is directly necessary for those bounded changes to validate",
			"report each supporting adjustment, its path, and its direct necessity",
			"MUST NOT repair pre-existing failures, unrelated failures, adjacent functionality, speculative cleanup, or broad refactors",
			"stop before making a correction that would expand functional scope or before performing a destructive operation",
		)
	})

	t.Run("gates task checkboxes on green validation", func(t *testing.T) {
		requireTextInOrder(t, plannedLoop,
			"leave every affected task unchecked throughout diagnosis and repair",
			"mark only the assigned completed tasks and only after their relevant validation is green",
			"A failure, unavailable relevant validation, or real blocker leaves those tasks unchecked.",
		)
	})

	t.Run("preserves task state and hard stops", func(t *testing.T) {
		requireTextInOrder(t, plannedLoop,
			"leave every incomplete or red task unchecked",
			"never mark a task complete merely because the batch ended",
			"out-of-batch writes, functional expansion, destructive commands, unresolvable OpenSpec state, or a checked-task/red-validation conflict",
			"A planned-loop hard blocker exists",
			"performs writes outside the assigned batch that are not reported minimal directly necessary adjustments",
			"expands functional behavior beyond its tasks",
			"runs a destructive command",
			"runs a full repository suite or build",
			"fresh OpenSpec state cannot be resolved safely",
			"a checked task has relevant red validation",
			"Stop immediately and dispatch no further batch.",
			"Never ignore red evidence, uncheck or check tasks to remove a conflict, or relabel incomplete work as complete.",
		)
	})
}

func TestOrchestratorStrictRoutesExcludePlannedDeferral(t *testing.T) {
	sharedPolicy := orchestratorSection(t, "### Shared implementation-result policy", "### Safe direct execution")
	requireTextInOrder(t, sharedPolicy,
		"Apply it without exception to an initial Direct Safe result, a bounded Direct Safe review-fix result, Direct Fast",
		"OpenSpec bootstrap and target resolution",
		"post-verification finding-ID fixes, and final OpenSpec verification",
		"Planned-batch self-repair and deferral never apply to Direct work, review-fix batches, bootstrap, target resolution, post-verification finding-ID fixes, or final verification.",
		"These classifications never apply to Direct work, review-fix batches, bootstrap, target resolution, or final verification",
	)

	t.Run("existing target resolution stays strict", func(t *testing.T) {
		requireTextInOrder(t,
			orchestratorSection(t, "## Execution route selection", "### Shared implementation-result policy"),
			"retain and report the target-resolution command, exit code, and diagnostic",
			"apply the shared mandatory-stop policy",
		)
	})

	t.Run("bootstrap stays strict", func(t *testing.T) {
		requireTextInOrder(t, orchestratorBootstrapSection(t),
			"If bootstrap blocks or fails, do not launch the OpenSpec worker",
			"apply the shared mandatory-stop policy",
		)
	})

	t.Run("direct and review fixes stay strict", func(t *testing.T) {
		requireTextInOrder(t,
			orchestratorSection(t, "### Safe direct execution", "## Delegation rules"),
			"clean under the shared implementation-result policy",
			"apply the shared mandatory-stop policy",
			"same structured result contract used for the initial Direct task",
			"Apply that policy to every other unsafe fix result.",
		)
	})

	t.Run("final verification stays strict", func(t *testing.T) {
		requireTextInOrder(t,
			orchestratorSection(t, "### Implementation stops and completion routing", "### Between phases"),
			"Automatically dispatch `openspec-verifier`",
			"If verification fails, blocks, or is incomplete",
			"apply the shared mandatory-stop policy",
		)
	})
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
			"Run only the selected reviewers. Give each the confirmed Brief as intent context",
			"do not inject a complete patch.",
			"independently use Git/Bash to inspect the current staged,",
			"unstaged, and untracked non-ignored local changes",
			"excluding ignored",
			"files and secrets.",
			"The Brief informs intended behavior",
			"not a boundary on",
			"supported findings from those local changes.",
			"Reviewers remain report-only.",
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

	t.Run("orders advisory CodeGraph preparation after blocking JSON readiness", func(t *testing.T) {
		requireTextInOrder(t, section,
			"Treat OpenSpec JSON output as the only readiness source",
			"If initialization fails or the follow-up JSON still has no resolvable root, block",
			"Complete every blocking OpenSpec JSON readiness step above before CodeGraph preparation",
			"For a local project root, inspect `<project-root>/.codegraph/`",
			"before dispatching any OpenSpec worker",
		)
	})

	t.Run("initializes a missing local CodeGraph index at most once", func(t *testing.T) {
		requireTextInOrder(t, section,
			"If `<project-root>/.codegraph/` exists, skip CodeGraph initialization",
			"When it is absent, run exactly `codegraph init <project-root>` once",
			"Do not make a second CodeGraph initialization attempt in this bootstrap",
		)
		if count := strings.Count(normalizedText(section), "run exactly `codegraph init <project-root>` once"); count != 1 {
			t.Fatalf("CodeGraph init command contract occurs %d times, want 1", count)
		}
	})

	t.Run("warns and falls back to filesystem tools", func(t *testing.T) {
		requireTextInOrder(t, section,
			"If `codegraph init <project-root>` is unavailable or exits non-zero",
			"retain the exact command, exit code, and advisory warning",
			"continue the OpenSpec workflow with filesystem tools",
			"CodeGraph preparation is advisory and MUST NOT weaken or replace blocking OpenSpec JSON readiness",
		)
	})

	t.Run("skips stores without local roots and prohibits worker reinitialization", func(t *testing.T) {
		requireTextInOrder(t, section,
			"An explicit store without a local project root skips CodeGraph preparation",
		)

		workflow := orchestratorSection(t, "## OpenSpec workflow", "## Language")
		requireTextInOrder(t, workflow,
			"Every OpenSpec worker prompt MUST state that CodeGraph initialization is owned by bootstrap",
			"MUST NOT run `codegraph init` or rerun CodeGraph initialization",
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
	assetSource := assetfs.Directory(assets)
	categories, err := catalog.Load(assetSource)
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
			if item.Kind != catalog.CopyFile || item.Source != "agents/"+fileName || item.Dest != filepath.Join("agents", fileName) {
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
			t.Errorf("installed %s differs from selected asset", item.Name)
		}
	}
}
