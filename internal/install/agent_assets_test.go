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

func requireTextAbsent(t *testing.T, value string, snippets ...string) {
	t.Helper()
	normalizedValue := normalizedText(value)
	for _, snippet := range snippets {
		normalizedSnippet := normalizedText(snippet)
		if strings.Contains(normalizedValue, normalizedSnippet) {
			t.Fatalf("unexpected contract text %q", normalizedSnippet)
		}
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

type strictDefaultConsumerContract struct {
	name         string
	startHeading string
	endHeading   string
	continuation []string
}

var strictDefaultConsumerContracts = []strictDefaultConsumerContract{
	{
		name:         "existing target resolution",
		startHeading: "## Execution route selection",
		endHeading:   "### Shared implementation-result policy",
		continuation: []string{
			"clean under the shared implementation-result policy",
			"continue through the status-driven OpenSpec workflow",
		},
	},
	{
		name:         "initial Direct Safe result",
		startHeading: "### Safe direct execution",
		endHeading:   "### Fast direct execution",
		continuation: []string{
			"clean under the shared implementation-result policy",
			"proceed to the direct Safe review gate",
		},
	},
	{
		name:         "Direct Fast result",
		startHeading: "### Fast direct execution",
		endHeading:   "### Direct review gate",
		continuation: []string{
			"clean under the shared implementation-result policy",
			"Report the result explicitly as implemented but not verified and do not open the direct review gate.",
		},
	},
	{
		name:         "bounded Direct Safe review fix",
		startHeading: "### Direct review gate",
		endHeading:   "## Delegation rules",
		continuation: []string{
			"clean under the shared implementation-result policy",
			"After a clean fix",
			"Finish review (Recommended)",
		},
	},
	{
		name:         "OpenSpec bootstrap",
		startHeading: "### Bootstrap gate before OpenSpec workers",
		endHeading:   "### Workers and their official skills",
		continuation: []string{
			"dispatch the requested OpenSpec worker only when every blocking OpenSpec JSON readiness step succeeds",
			"the result is otherwise clean under the shared implementation-result policy",
		},
	},
	{
		name:         "post-verification finding-ID fix",
		startHeading: "## Review gate (after verification, before archive)",
		endHeading:   "## Language",
		continuation: []string{
			"clean under the shared implementation-result policy",
			"After a clean finding-ID fix result",
			"Archive without re-review (Recommended)",
		},
	},
	{
		name:         "final OpenSpec verification",
		startHeading: "**Completion rule:**",
		endHeading:   "### Between phases",
		continuation: []string{
			"clean under the shared implementation-result policy",
			"proceed directly to the existing Review gate",
		},
	},
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
			"the orchestrator inserts the complete authoritative Shared corrected-failure result fields defined below",
			"Direct-specific details: the selected mode, compliance with its mode obligations, and any deviation from the confirmed Brief",
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

	t.Run("uses one evidence rule for tooling and real failures", func(t *testing.T) {
		requireTextInOrder(t, section,
			"This shared strict policy is the default for every implementation result and every OpenSpec control point that invokes it.",
			"The sole route-specific classification exception is inside the automatic planned-task loop",
			"Any result that is not explicitly eligible for that exception remains subject to this strict default.",
			"Shared corrected-failure result fields",
			"single authoritative field set for every implementation or verification result",
			"insert this exact list into the worker prompt",
			"status (`done`, `partial`, or `blocked`)",
			"files touched",
			"every command in execution order with its exit code",
			"for each non-zero command, the failed command and exit code, diagnosed cause and bounded correction",
			"later equivalent-or-broader relevant validation command and exit code",
			"final relevant validation state",
			"deviations from the assigned Brief, change, task, or scope, including scope expansion and out-of-scope work",
			"Classify an intermediate non-zero command, whether caused by a syntax or invocation mistake or by a real implementation or verification failure, as corrected only when all of these are true:",
			"the same worker diagnoses the cause and repairs it within the same bounded invocation and authorized scope",
			"the worker retains the original failed command and non-zero exit code, the diagnosed cause and bounded correction, and the later successful validation command and exit code in execution order",
			"the later validation is equivalent to or broader than the failed command's relevant scope and exits zero",
			"the final status is `done` and the final relevant validation state is green",
			"the worker reports no deviation, scope expansion, or out-of-scope work",
			"The successful validation MUST cover the failed command's relevant scope or a superset of it.",
			"An eligible corrected failure is clean under this shared policy.",
			"Surface the complete incident evidence and follow the strict-default control point's existing clean-result route without an authorization question, mandatory stop, or archive delay solely because of that incident.",
		)

		if count := strings.Count(normalizedText(section), "Classify an intermediate non-zero command"); count != 1 {
			t.Fatalf("corrected-failure rule occurs %d times, want 1", count)
		}
		requireTextAbsent(t, section,
			"corrected tooling error",
			"A real verification or implementation failure remains a mandatory stop",
			"the failed command was not caused by syntax or invocation",
		)
	})

	t.Run("every strict consumer follows its existing clean route", func(t *testing.T) {
		for _, consumer := range strictDefaultConsumerContracts {
			t.Run(consumer.name, func(t *testing.T) {
				requireTextInOrder(t,
					orchestratorSection(t, consumer.startHeading, consumer.endHeading),
					consumer.continuation...,
				)
			})
		}
	})

	t.Run("route prompts reference the authoritative result fields", func(t *testing.T) {
		contracts := []struct {
			name         string
			startHeading string
			endHeading   string
			extra        string
		}{
			{name: "Direct", startHeading: "## Execution route selection", endHeading: "### Shared implementation-result policy", extra: "Direct-specific details"},
			{name: "bootstrap", startHeading: "### Bootstrap gate before OpenSpec workers", endHeading: "### Workers and their official skills", extra: "resolved context key, advisory warnings, and the blocking reason"},
			{name: "OpenSpec task template", startHeading: "### Task prompt template", endHeading: "### Inline single-change archive", extra: "route-specific next recommended action"},
			{name: "planned batch", startHeading: "### Automatic planned-task loop and bounded batches", endHeading: "### Implementation stops and completion routing", extra: "repair-progress evidence and every directly-necessary supporting adjustment"},
		}
		for _, contract := range contracts {
			t.Run(contract.name, func(t *testing.T) {
				requireTextInOrder(t,
					orchestratorSection(t, contract.startHeading, contract.endHeading),
					"authoritative Shared corrected-failure result fields",
					contract.extra,
				)
			})
		}
	})

	t.Run("clean verification reaches the one normal archive path", func(t *testing.T) {
		requireTextInOrder(t,
			orchestratorSection(t, "## Review gate (after verification, before archive)", "## Language"),
			"None, archive now",
			"proceed to archive automatically",
			"An empty selection closes the review without fixes and proceeds to archive.",
		)
		requireTextInOrder(t,
			orchestratorSection(t, "### Inline single-change archive", "### Planned-task implementation state"),
			"Whenever this workflow reaches \"proceed to archive\"",
			"the primary orchestrator MUST load and invoke `openspec-archive-change` itself",
		)
		requireTextInOrder(t, section,
			"without an authorization question, mandatory stop, or archive delay solely because of that incident",
		)
	})
}

func TestOrchestratorRecoveredMissingOpenSpecProbeContract(t *testing.T) {
	section := orchestratorSection(t, "### Shared implementation-result policy", "### Safe direct execution")

	// This is the exact `ls openspec`-before-creation incident contract: the
	// probe observes absent state, then the same worker creates and validates it.
	requireTextInOrder(t, section,
		"Classify a failed command separately as a **recovered non-destructive probe**, rather than as a corrected implementation or validation failure",
		"the command is inspection-only, performs no mutation or destructive action, and fails solely because expected pre-operation state is absent or not yet established",
		"the same worker subsequently completes the authorized operation successfully within the same bounded invocation and authorized scope",
		"the worker retains, in execution order, the failed probe and non-zero exit code, the absent-state cause, the successful authorized operation, the authoritative final validation command and exit code, and why that validation proves the requested final state",
		"the authoritative final validation exits zero and proves the requested final state",
		"the final status is `done` and the final relevant validation state is green",
		"the worker reports no deviation from the assigned Brief, scope expansion, or out-of-scope work",
		"A recovered probe requires no invented repair or equivalent rerun of the inspection command.",
		"Surface its complete ordered incident evidence and follow the strict-default control point's existing clean-result route without an authorization, continuation, or follow-up question, mandatory stop, or archive delay solely because of that incident.",
		"Never hide or relabel the failed probe or its exit code.",
	)
}

func TestOrchestratorMandatoryImplementationStopsContract(t *testing.T) {
	section := orchestratorSection(t, "### Shared implementation-result policy", "### Safe direct execution")
	conditions := []string{
		"a result containing an intermediate non-zero command fails any item in the authoritative corrected-failure eligibility checklist above",
		"a claimed recovered probe fails any item in the authoritative recovered-probe eligibility checklist above, including when the command is destructive or mutating rather than inspection-only",
		"the worker performs any destructive action before or after a failed command",
		"corrected-failure or recovered-probe evidence is incomplete, comes from different workers, or relies on a successful command that is unrelated, narrower than required, or does not authoritatively prove the requested final state",
		"the final relevant verification state is red",
		"status is `partial` or `blocked`",
		"the worker reports a deviation from the assigned Brief, change, task, or scope",
		"the worker reports scope expansion",
		"the worker reports out-of-scope work",
		"a TDD or expected failure remains red at batch end",
	}

	for _, condition := range conditions {
		t.Run(condition, func(t *testing.T) {
			requireTextInOrder(t, section, "A mandatory stop applies when any of these is true:", condition)
		})
	}
	requireTextInOrder(t, section,
		"For the strict default routes above, every listed condition is a mandatory stop.",
		"Only an eligible section-bounded planned-task batch",
		"local `partial`, local `blocked`, or red focused test as deferrable",
		"the affected incomplete tasks remain unchecked",
		"Classify an additional read or a successful focused test of modified code as a benign, continuable deviation",
		"These classifications never apply to Direct work, review-fix batches, bootstrap, target resolution, post-verification finding-ID fixes, or final verification",
		"On every mandatory stop, apply this shared mandatory-stop policy in two ordered, separate steps:",
		"First report the blocking status and all retained evidence needed to choose an action",
		"Then ask exactly one blocker-specific next-action `question`.",
		"Until the user selects an action, do not retry, continue, broaden scope, select substitute work, advance to the route's next phase, or dispatch any worker.",
	)
	requireTextInOrder(t, section,
		"The recovered-probe classification removes no route-local prerequisite or planned-task-specific safeguard and changes no unrelated routing or planned-task behavior.",
		"Any unresolved failure or ambiguous probe classification remains subject to the mandatory-stop policy.",
	)
}

func TestOrchestratorGeneratedValidationArtifactsResultContract(t *testing.T) {
	orchestrator := readRepositoryAsset(t, "agents", "angel-orchestrator.md")
	section := orchestratorSection(t, "### Shared implementation-result policy", "### Safe direct execution")
	const heading = "**Generated-validation-artifacts result category:**"

	if count := strings.Count(normalizedText(orchestrator), normalizedText(heading)); count != 1 {
		t.Fatalf("authoritative generated-validation-artifacts category occurs %d times, want 1", count)
	}
	requireTextInOrder(t, section,
		heading,
		"This is the single authoritative generated-validation-artifacts category for every implementation or verification result.",
		"Every result MUST include this category, using `none` when there are no eligible outputs.",
		"generated paths",
		"the producing authorized validation command and its zero exit code",
		"the command-specific before/after workspace evidence or equivalent attributable diff",
		"confirmation that the outputs remain retained in the workspace",
	)
	requireTextInOrder(t, section,
		"Classify an output in this category only when all of these are true:",
		"an authorized validation command exits zero",
		"command-specific evidence attributes every created or modified path to that command",
		"the output is regenerable",
		"no intervening manual mutation occurred",
		"the attributable diff contains no manual source-code edit",
	)
}

func TestOrchestratorGeneratedValidationArtifactsPositiveCases(t *testing.T) {
	section := orchestratorSection(t, "### Shared implementation-result policy", "### Safe direct execution")
	cases := []struct {
		name     string
		path     string
		gitState string
	}{
		{name: "Next output may be ignored", path: "`.next/`", gitState: "ignored"},
		{name: "CodeGraph output may be visible", path: "`.codegraph/`", gitState: "visible to Git"},
		{name: "TypeScript metadata may be tracked", path: "`*.tsbuildinfo`", gitState: "tracked"},
	}

	requireTextInOrder(t, section,
		"Eligibility does not depend on whether Git ignores or tracks an artifact and does not use a filename allowlist.",
		"Eligible outputs remain in the workspace: do not clean, revert, delete, stage, or commit them automatically.",
		"Do not report an eligible output as a deviation, scope expansion, or out-of-scope work, and do not request authorization solely because it exists.",
	)
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			requireTextInOrder(t, section, testCase.path, testCase.gitState)
		})
	}
}

func TestPreExistingUnchangedTopLevelDotpathFilter(t *testing.T) {
	orchestrator := readRepositoryAsset(t, "agents", "angel-orchestrator.md")
	section := orchestratorSection(t, "### Shared implementation-result policy", "### Safe direct execution")
	const heading = "**Pre-existing unchanged top-level dotpath filter:**"
	const propagation = "apply the authoritative pre-existing unchanged top-level dotpath filter"

	if count := strings.Count(normalizedText(orchestrator), normalizedText(heading)); count != 1 {
		t.Fatalf("authoritative pre-existing unchanged top-level dotpath filter occurs %d times, want 1", count)
	}
	requireTextInOrder(t, section,
		heading,
		"Before doing work, every worker that may audit the workspace MUST internally capture reliable worker-start evidence",
		"For each candidate repository-relative path, inspect only its first component.",
		"Silently omit the path before any result classification or reporting only when all of these are true:",
		"the first component begins with `.`",
		"the reliable worker-start evidence proves that the path was already modified",
		"worker-end evidence proves that the path's complete state is identical to that baseline",
		"For example, already-modified `.vscode/settings.json` that is identical at worker completion qualifies and is silently omitted.",
		"A qualifying path MUST NOT appear in files touched, generated-validation-artifacts, deviations, scope expansion, out-of-scope work, or any other result category.",
		"Establish and compare its state internally; do not expose its contents or diff.",
		"The filter grants no authority to create or modify a dotpath.",
		"When the first component does not begin with `.`, worker-start evidence is missing, ambiguous, or unreliable, or the path changes during the worker invocation, keep the path under normal generated-output, corrected-failure, deviation, scope, destructive-action, ambiguity, red-state, and mandatory-stop handling.",
	)

	routes := []struct {
		name         string
		startHeading string
		endHeading   string
	}{
		{name: "Direct Safe", startHeading: "### Safe direct execution", endHeading: "### Fast direct execution"},
		{name: "planned OpenSpec implementation", startHeading: "### Automatic planned-task loop and bounded batches", endHeading: "### Between phases"},
		{name: "bounded Direct fix", startHeading: "### Direct review gate", endHeading: "## Delegation rules"},
		{name: "bounded OpenSpec fix", startHeading: "## Review gate (after verification, before archive)", endHeading: "## Language"},
		{name: "final OpenSpec verification", startHeading: "**Completion rule:**", endHeading: "### Between phases"},
	}
	for _, route := range routes {
		t.Run(route.name, func(t *testing.T) {
			requireTextInOrder(t, orchestratorSection(t, route.startHeading, route.endHeading), propagation)
		})
	}

	implementer := readRepositoryAsset(t, "agents", "openspec-implementer.md")
	requireTextInOrder(t, implementer,
		"Before any work, internally capture reliable worker-start evidence for paths that are already modified.",
		"Silently omit a path from every result category only when its first repository-relative component begins with `.`, that baseline proves it was already modified, and its complete worker-end state is identical.",
		"Do not expose the path's contents or diff.",
		"This filter grants no dotpath write authority; absent, ambiguous, or unreliable baseline evidence and every worker-time change keep normal handling.",
	)

	verifier := readRepositoryAsset(t, "agents", "openspec-verifier.md")
	requireTextInOrder(t, verifier,
		"At worker start, before any command that may change the workspace, internally capture reliable evidence for paths that are already modified.",
		"Silently omit a path from every result category only when its first repository-relative component begins with `.`, that baseline proves it was already modified, and its complete worker-end state is identical.",
		"Do not expose the path's contents or diff.",
		"This filter grants no dotpath write authority; absent, ambiguous, or unreliable baseline evidence and every worker-time change keep normal handling.",
	)
}

func TestOrchestratorGeneratedValidationArtifactsSafetyNegatives(t *testing.T) {
	section := orchestratorSection(t, "### Shared implementation-result policy", "### Safe direct execution")
	conditions := []string{
		"the producing command exits non-zero",
		"the final relevant validation state is red",
		"the worker performs destructive cleanup or any other destructive action before or after producing an artifact",
		"command-specific causal evidence is missing or ambiguous",
		"an intervening manual mutation occurred",
		"the attributable diff contains a manual source-code edit",
		"the worker expands functional behavior beyond the assigned scope",
		"the worker reports out-of-scope work",
	}

	for _, condition := range conditions {
		t.Run(condition, func(t *testing.T) {
			requireTextInOrder(t, section,
				"Generated-validation-artifact eligibility does not apply when any of these is true:",
				condition,
				"apply the existing corrected-failure, recovered-probe, deviation, scope, and mandatory-stop rules",
			)
		})
	}
}

func TestOrchestratorGeneratedValidationArtifactsRoutePropagation(t *testing.T) {
	const classification = "apply the authoritative generated-validation-artifacts category"
	routes := []struct {
		name         string
		startHeading string
		endHeading   string
		before       string
		after        string
	}{
		{
			name:         "Direct Safe",
			startHeading: "### Safe direct execution",
			endHeading:   "### Fast direct execution",
			before:       "executable verification was available and run",
			after:        "If executable verification is unavailable or its command/exit-code evidence is omitted",
		},
		{
			name:         "planned OpenSpec implementation",
			startHeading: "### Automatic planned-task loop and bounded batches",
			endHeading:   "### Between phases",
			before:       "Every planned-task implementer prompt MUST name the section, list the exact task identifiers and short summaries in the batch",
			after:        "A planned-loop hard blocker exists",
		},
		{
			name:         "bounded Direct fix",
			startHeading: "### Direct review gate",
			endHeading:   "## Delegation rules",
			before:       "Only user-selected findings become work.",
			after:        "Apply that policy to every other unsafe fix result.",
		},
		{
			name:         "bounded OpenSpec fix",
			startHeading: "## Review gate (after verification, before archive)",
			endHeading:   "## Language",
			before:       "Only findings the user selects become a task",
			after:        "Treat the finding-ID fix as clean only when its result is clean under the shared implementation-result policy.",
		},
		{
			name:         "final OpenSpec verification",
			startHeading: "**Completion rule:**",
			endHeading:   "### Between phases",
			before:       "successful executed verification evidence",
			after:        "a failed, blocked, or incomplete verification retains its status, commands, exit codes, and diagnostic",
		},
	}

	for _, route := range routes {
		t.Run(route.name, func(t *testing.T) {
			section := orchestratorSection(t, route.startHeading, route.endHeading)
			requireTextInOrder(t, section, route.before, route.after)
			requireTextInOrder(t, section, classification)
		})
	}
}

func TestOpenSpecWorkersGeneratedValidationArtifactsBoundaries(t *testing.T) {
	implementer := readRepositoryAsset(t, "agents", "openspec-implementer.md")
	verifier := readRepositoryAsset(t, "agents", "openspec-verifier.md")

	t.Run("implementer retains exact task and validation boundaries", func(t *testing.T) {
		requireTextInOrder(t, implementer,
			"Implement ONLY the task batch assigned in your task prompt.",
			"Generated command effects do not grant authority for a manual source edit or widen the assigned task batch.",
			"The implementer MUST NOT run the full repository test suite or any build.",
			"Retain eligible generated outputs and report their command-specific attribution in the generated-validation-artifacts category; never clean, revert, or delete them automatically.",
		)
		requireTextInOrder(t, implementer,
			"Return every authoritative Shared corrected-failure result field supplied in the orchestrator prompt",
			"complete corrected-failure or recovered-probe evidence",
			"final relevant validation state",
			"deviations including scope expansion and out-of-scope work",
			"generated-validation-artifacts category",
			"repair-progress evidence",
			"every directly-necessary supporting adjustment",
			"route-specific next recommended action",
		)
	})

	t.Run("verifier permits command effects without edit authority", func(t *testing.T) {
			requireTextInOrder(t, verifier,
				"Validation commands may leave only their normal causally proven generated outputs, including tracked generated artifacts.",
				"Those command effects do not grant manual edit or write authority.",
				"never edit, fix, reformat, or write product code or any tracked/project file",
				"Do not stage, commit, install dependencies, generate sources",
				"Retain eligible generated outputs and report them through the generated-validation-artifacts category; never clean, revert, or delete them automatically.",
			)
		requireTextInOrder(t, verifier,
			"Return every authoritative Shared corrected-failure result field supplied in the orchestrator prompt",
			"ordered failed/correction/success evidence",
			"equivalent-or-broader scope coverage",
			"final relevant validation state",
			"files touched",
			"deviations, scope expansion, or out-of-scope evidence",
			"generated-validation-artifacts category",
			"`verdict` (`pass`, `fail`, or `not-verified`)",
		)
	})
}

func TestOrchestratorPlannedBatchDeferralContract(t *testing.T) {
	sharedPolicy := orchestratorSection(t, "### Shared implementation-result policy", "### Safe direct execution")
	plannedLoop := orchestratorSection(t, "### Planned-task implementation state", "### Between phases")

	t.Run("limits eligibility to fresh section bounded planned batches", func(t *testing.T) {
		requireTextInOrder(t, sharedPolicy,
			"only a section-bounded planned OpenSpec task batch selected from the active change's fresh `tasks.md`",
			"Any result that is not explicitly eligible for that exception remains subject to this strict default.",
			"the same planned-task implementer must diagnose and repair real failures attributable to its bounded changes within the same invocation",
			"continue bounded repair and rerun relevant validation only while each cycle makes demonstrable progress",
			"changed diagnostic evidence, a narrower attributable cause, a completed necessary bounded correction, or improved relevant validation",
			"for at most three repair/rerun cycles",
			"A returned attributable failure is not deferrable while a further safe bounded repair cycle can still make demonstrable progress within that cap.",
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
			"the planned-task self-repair rule defined in the shared implementation-result policy",
			"treating a failure as attributable only when it was caused by files or behavior changed for the assigned batch",
		)
	})

	t.Run("bounds repair writes and supporting adjustments", func(t *testing.T) {
		requireTextInOrder(t, plannedLoop,
			"MUST limit writes to files assigned to the batch",
			"A **directly-necessary supporting adjustment** is a minimal write outside that batch made only when it is directly necessary for those bounded changes to validate",
			"MUST NOT silently self-authorize such an adjustment",
			"report the adjustment, its path, and its direct necessity, and surface it through the shared mandatory-stop policy for user authorization",
			"MUST NOT repair pre-existing failures, unrelated failures, adjacent functionality, speculative cleanup, or broad refactors",
			"stop before making a correction that would expand functional scope or before performing a destructive operation",
		)
	})

	t.Run("gates task checkboxes on green validation", func(t *testing.T) {
		requireTextInOrder(t, plannedLoop,
			"mark only the assigned completed tasks after their relevant validation is green",
			"leave every other task unchecked",
			"any task under diagnosis or repair, any incomplete or red task, and any task affected by a failure, unavailable relevant validation, or real blocker",
			"the end of a batch never by itself completes a task",
		)
	})

	t.Run("preserves task state and hard stops", func(t *testing.T) {
		requireTextInOrder(t, plannedLoop,
			"the end of a batch never by itself completes a task",
			"out-of-batch writes, functional expansion, destructive commands, unresolvable OpenSpec state, or a checked-task/red-validation conflict",
			"A planned-loop hard blocker exists",
			"performs any write outside the assigned batch",
			"including a claimed directly-necessary supporting adjustment, which the orchestrator surfaces through the shared mandatory-stop policy for user authorization rather than treating as self-approved",
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

func TestOrchestratorVerifierOwnedTaskRoutingContract(t *testing.T) {
	freshState := orchestratorSection(t, "### Planned-task implementation state", "### Automatic planned-task loop and bounded batches")
	plannedLoop := orchestratorSection(t, "### Automatic planned-task loop and bounded batches", "### Implementation stops and completion routing")
	completion := orchestratorSection(t, "**Completion rule:**", "### Between phases")

	t.Run("schedules ordinary sections before marked work", func(t *testing.T) {
		requireTextInOrder(t, plannedLoop,
			"With a valid marked terminal section, exclude every task in that section from every implementer batch",
			"select the next incomplete ordinary section before it",
			"When fresh state shows every ordinary task complete and only pending tasks from the valid marked terminal section remain",
			"do not dispatch an implementer",
			"Automatically dispatch exactly one `openspec-verifier`",
		)
	})

	t.Run("recognizes only exact structurally valid markers", func(t *testing.T) {
		requireTextInOrder(t, freshState,
			"structurally validate owner-marker state before scheduling",
			"ownership exists only for one exact `<!-- owner: openspec-verifier -->` marker",
			"the first nonblank line of the final named top-level task section",
			"A section title, task wording, legacy verification prose, or any other comment does not establish ownership",
			"Duplicate, malformed, misplaced, nested, or non-terminal owner markers are an invalid task-state conflict",
			"stop without dispatching an implementer or verifier and change no checkbox",
		)
		requireTextInOrder(t, plannedLoop,
			"With no valid exact owner marker, preserve the existing behavior",
			"verification-like titles or prose are ordinary",
		)
	})

	t.Run("propagates exact local and store contexts", func(t *testing.T) {
		requireTextInOrder(t, freshState,
			"For a local change, use the exact repo-local root returned by successful bootstrap as the working directory and context identity",
			"run `openspec status --change <name> --json` there",
			"For an explicit store, retain the exact store id and append `--store <id>` to every applicable status or guarded verifier-task command",
			"use `store:<id>` as the context identity and never infer, substitute, or switch to a local project path for that store",
			"Propagate that exact local root or store id in every worker dispatch and all later refreshes.",
		)
		requireTextInOrder(t, completion,
			"propagate the exact repo-local root as the working directory or the exact explicit store id without local-path inference",
			"`angel-ai verifier-tasks snapshot --change <name>`",
			"the same `--store <id>` when applicable",
			"guarded completion, and result",
			"make one completion confirmation by applying the fresh-state invariant in that same context",
		)
	})

	t.Run("stops on invalid stale or incomplete state", func(t *testing.T) {
		requireTextInOrder(t, freshState,
			"its resolved path or active context changes, or marker state is invalid",
			"stop the planned-task cycle as `blocked`",
		)
		requireTextInOrder(t, completion,
			"Every failure, `not-verified`, partial or blocked status, evidence gap, non-successful completion, stale snapshot, changed context or resolved path, or conflict is a mandatory stop",
			"before retry, review, archive, fallback checkbox changes, or any worker dispatch",
			"make one completion confirmation by applying the fresh-state invariant in that same context",
			"any remaining task, changed resolution, unreadable state, or checked-task/red-evidence disagreement is a mandatory-stop conflict",
		)
	})

	t.Run("ordinary unmarked pass proceeds without guarded completion", func(t *testing.T) {
		requireTextInOrder(t, completion,
			"For the ordinary unmarked entry",
			"clean exact `status: done`",
			"global `verdict: pass`",
			"successful executed verification evidence",
			"`completion: not-attempted` (or the equivalent ordinary no-completion state)",
			"proceed directly to the existing Review gate",
			"guarded completion is neither required nor permitted for this entry",
		)
	})

	t.Run("accepts atomic completion and reuses one pass", func(t *testing.T) {
		requireTextInOrder(t, completion,
			"Accept only the complete tuple: a result clean under the shared implementation-result policy",
			"exact `status: done`",
			"global `verdict: pass`",
			"successful executed task-specific evidence for every captured marked task",
			"exact `completion: completed`",
			"no conflict",
			"After accepting the marked-only tuple",
			"make one completion confirmation by applying the fresh-state invariant in that same context",
			"complete artifact status, and no pending checkbox",
			"retain and reuse the returned pass and command evidence as final verification",
		)
		if count := strings.Count(normalizedText(completion), normalizedText("Automatically dispatch `openspec-verifier`")); count != 1 {
			t.Fatalf("ordinary completion dispatch count = %d, want 1", count)
		}
		requireTextAbsent(t, completion,
			"dispatch a second verifier",
			"rerun verification after checkbox completion",
		)
	})

	t.Run("permits verification while only marked terminal tasks remain", func(t *testing.T) {
		verificationPolicy := orchestratorSection(t, "## Verification policy", "## Review gate (after verification, before archive)")
		requireTextInOrder(t, verificationPolicy,
			"the verifier runs the mandatory repository tests and build only after fresh task state shows either all planned tasks complete or only the valid marked terminal section pending",
			"planned-task implementers run only their permitted bounded lint, typecheck, and minimum relevant test checks",
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
		"These classifications never apply to Direct work, review-fix batches, bootstrap, target resolution, post-verification finding-ID fixes, or final verification",
	)

	plannedLoop := orchestratorSection(t, "### Automatic planned-task loop and bounded batches", "### Between phases")
	t.Run("retains every planned-only safeguard", func(t *testing.T) {
		requireTextInOrder(t, plannedLoop,
			"Automatic execution rule",
			"After every clean result",
			"automatically repeat for the next incomplete section",
			"planned-task self-repair rule",
			"leave every other task unchecked",
			"Conservative independence gate",
			"Single retry round",
			"A planned-loop hard blocker exists",
		)
	})

	plannedOnlyContracts := []string{
		"planned-task self-repair rule",
		"leave every other task unchecked",
		"deferrable result",
		"Conservative independence gate",
		"Automatic execution rule",
		"Single retry round",
		"A planned-loop hard blocker exists",
	}
	for _, consumer := range strictDefaultConsumerContracts {
		t.Run(consumer.name+" cannot use planned-only safeguards", func(t *testing.T) {
			requireTextAbsent(t,
				orchestratorSection(t, consumer.startHeading, consumer.endHeading),
				plannedOnlyContracts...,
			)
		})
	}

	t.Run("existing target resolution stays strict", func(t *testing.T) {
		requireTextInOrder(t,
			orchestratorSection(t, "## Execution route selection", "### Shared implementation-result policy"),
			"retain and report the target-resolution command, exit code, and diagnostic",
			"apply the shared mandatory-stop policy",
		)
	})

	t.Run("bootstrap stays strict", func(t *testing.T) {
		requireTextInOrder(t, orchestratorBootstrapSection(t),
			"every blocking OpenSpec JSON readiness step succeeds",
			"Any real readiness failure or other non-clean result remains a mandatory stop",
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
			"a failed, blocked, or incomplete verification retains its status, commands, exit codes, and diagnostic",
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
			evidence:     "retains its status, commands, exit codes, and diagnostic",
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
			"the primary orchestrator, never a report-only reviewer, MUST invoke ONE multi-select `question` asking which reviews to run",
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
		)
	})

	t.Run("closes automatically when reviewers have no findings", func(t *testing.T) {
		requireTextInOrder(t, section,
			"If every selected reviewer reports `No findings.`, end the Direct review automatically",
			"without invoking an empty findings-selection question.",
		)
	})

	t.Run("uses an empty-default finding question owned by the orchestrator", func(t *testing.T) {
		requireTextInOrder(t, section,
			"deduplicate their findings, present one numbered list",
			"have the primary orchestrator invoke ONE multi-select `question` asking which findings to fix",
			"with no option preselected",
			"Reviewers MUST NOT invoke this question.",
			"An empty selection ends the Direct review without fixes.",
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

	t.Run("recommends finishing and reruns only affected reviewers", func(t *testing.T) {
		requireTextInOrder(t, section,
			"After a clean fix",
			"the primary orchestrator MUST invoke ONE single-select `question`",
			"**Finish review (Recommended)** / **Re-run responsible reviewers**",
			"Recommend finishing without re-review.",
			"If the user requests confirmation",
			"rerun only reviewers responsible for the addressed selected findings",
			"If every re-run reviewer reports `No findings.`, end the Direct review automatically.",
			"If a re-review reports new or pending findings",
			"return to the same findings-selection multi-select `question`, again with no option preselected.",
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
			"**Archive without re-review (Recommended)** / **Re-run responsible reviewers**",
			"re-run only the reviewers whose findings were addressed",
		)
	})
}

func TestOrchestratorOpenSpecReviewContract(t *testing.T) {
	section := orchestratorSection(t, "## Review gate (after verification, before archive)", "## Language")

	t.Run("orchestrator owns reviewer selection", func(t *testing.T) {
		requireTextInOrder(t, section,
			"Once `openspec-verifier` reports the change verified",
			"the primary orchestrator, never a report-only reviewer, MUST invoke ONE multi-select `question` asking which reviews to run",
			"**Security risk** / **Simplicity** / **Correctness** / **None, archive now**",
			"Reviewers remain report-only.",
		)
	})

	t.Run("closes automatically when reviewers have no findings", func(t *testing.T) {
		requireTextInOrder(t, section,
			"If every selected reviewer reports `No findings.`, proceed to archive automatically",
			"without invoking an empty findings-selection question.",
		)
	})

	t.Run("uses an empty-default finding question owned by the orchestrator", func(t *testing.T) {
		requireTextInOrder(t, section,
			"have the primary orchestrator invoke ONE multi-select `question` asking which findings to fix",
			"with no option preselected",
			"Reviewers MUST NOT invoke this question.",
			"An empty selection closes the review without fixes and proceeds to archive.",
		)
	})

	t.Run("recommends archive and reruns only affected reviewers", func(t *testing.T) {
		requireTextInOrder(t, section,
			"After fixes land",
			"the primary orchestrator MUST invoke ONE single-select `question`",
			"**Archive without re-review (Recommended)** / **Re-run responsible reviewers**",
			"Recommend archive without re-review.",
			"If the user requests confirmation",
			"re-run only the reviewers whose findings were addressed",
			"If every re-run reviewer reports `No findings.`, proceed to archive automatically.",
			"If a re-review reports new or pending findings",
			"return to the same findings-selection multi-select `question`, again with no option preselected.",
		)
	})
}

func TestOrchestratorArchivesSingleChangesInline(t *testing.T) {
	orchestrator := readRepositoryAsset(t, "agents", "angel-orchestrator.md")
	requireTextInOrder(t, orchestrator,
		"Archive one named OpenSpec change after authorization",
		"Yes",
		"primary orchestrator via `openspec-archive-change`",
		"Bulk archive OpenSpec changes",
		"No",
		"`openspec-planner`",
	)

	section := orchestratorSection(t, "### Inline single-change archive", "### Planned-task implementation state")
	requireTextInOrder(t, section,
		"Archiving one named change is a bounded lifecycle control action, not planned implementation.",
		"the primary orchestrator MUST load and invoke `openspec-archive-change` itself",
		"Do not dispatch `openspec-planner` or `general` solely to perform that archive.",
		"The primary orchestrator owns every question required by the archive skill.",
		"If the user chooses to sync delta specs, delegate only that sync to `openspec-planner` with `openspec-sync-specs`",
		"resume the archive inline after a clean sync result.",
	)

	planner := readRepositoryAsset(t, "agents", "openspec-planner.md")
	if strings.Contains(planner, "`openspec-archive-change`") {
		t.Fatal("single-change archive must stay in the primary orchestrator")
	}
}

func TestOrchestratorOpenSpecBootstrapContract(t *testing.T) {
	section := orchestratorBootstrapSection(t)

	t.Run("gates workers on strict readiness with one dispatch rule", func(t *testing.T) {
		requireTextInOrder(t, section,
			"Before dispatching `openspec-planner`, `openspec-implementer`, or `openspec-verifier`",
			"dispatch one short `general` task",
			"Add the returned context key and dispatch the requested OpenSpec worker only when every blocking OpenSpec JSON readiness step succeeds",
			"the result is otherwise clean under the shared implementation-result policy",
			"Any real readiness failure or other non-clean result remains a mandatory stop",
			"retain and report its status, diagnostic, commands, and exit codes",
			"apply the shared mandatory-stop policy",
		)
		for _, redundant := range []string{
			"Only after a clean bootstrap result may the requested OpenSpec worker be dispatched.",
			"Only after a successful bootstrap may the requested OpenSpec worker be dispatched",
			"success here requires that clean classification",
		} {
			if strings.Contains(section, redundant) {
				t.Fatalf("redundant bootstrap gate retained: %q", redundant)
			}
		}
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

	t.Run("warns and falls back without making bootstrap non-clean", func(t *testing.T) {
		requireTextInOrder(t, section,
			"an unavailable or non-zero `codegraph init` is a retained advisory warning",
			"is excluded from that clean classification",
			"never blocks otherwise-green OpenSpec readiness",
			"If `codegraph init <project-root>` is unavailable or exits non-zero",
			"retain the exact command, exit code, and advisory warning",
			"return status `done` when OpenSpec readiness is otherwise green",
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
			"Every OpenSpec worker prompt MUST reference the bootstrap CodeGraph-ownership rule above",
			"the worker MUST NOT run `codegraph init`",
			"a bootstrap warning instead requires it to use filesystem tools for codebase discovery",
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

func TestConfigurableAgentAssetsDeclareNoVariant(t *testing.T) {
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
		})
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

func TestSelectedOrchestratorAssetIsCatalogedAndInstalledUnchanged(t *testing.T) {
	assetsRoot := filepath.Join("..", "..", "assets")
	assetSource := assetfs.Directory(assetsRoot)
	categories, err := catalog.Load(assetSource)
	if err != nil {
		t.Fatal(err)
	}

	var selected catalog.Item
	for _, category := range categories {
		if category.Name != "agents" {
			continue
		}
		for _, item := range category.Items {
			if item.Name == "angel-orchestrator" {
				selected = item
			}
		}
	}
	if selected != (catalog.Item{
		Name: "angel-orchestrator", Source: "agents/angel-orchestrator.md",
		Dest: filepath.Join("agents", "angel-orchestrator.md"), Kind: catalog.CopyFile,
	}) {
		t.Fatalf("orchestrator catalog item = %#v", selected)
	}

	configDir := t.TempDir()
	if _, err := ApplyInstallation(InstallationRequest{
		Items: []catalog.Item{selected}, Assets: assetSource, ConfigDir: configDir,
	}); err != nil {
		t.Fatal(err)
	}
	want, err := assetSource.ReadFile(selected.Source)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(configDir, selected.Dest))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatal("installed orchestrator differs from the selected asset")
	}
	for _, contract := range []string{
		"recovered non-destructive probe",
		"the same worker subsequently completes the authorized operation successfully",
		"without an authorization, continuation, or follow-up question",
		"Generated-validation-artifacts result category",
		"the producing authorized validation command and its zero exit code",
		"confirmation that the outputs remain retained in the workspace",
	} {
		if !strings.Contains(normalizedText(string(got)), normalizedText(contract)) {
			t.Errorf("installed orchestrator is missing recovered-probe contract %q", contract)
		}
	}
}

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

func TestOpenSpecPlannerVerifierOwnedTaskContract(t *testing.T) {
	planner := readRepositoryAsset(t, "agents", "openspec-planner.md")
	requireTextInOrder(t, planner,
		"When creating or updating `tasks.md`, you may emit at most one exact `<!-- owner: openspec-verifier -->` marker.",
		"place it as the first nonblank line immediately below the heading of one named top-level task section",
		"make that section the final task section",
		"only terminal final-verification obligations with independently reportable executed evidence",
		"Keep setup, implementation, test implementation, contract updates, and focused validation in earlier ordinary unmarked sections.",
		"If the plan has no terminal final-verification task, omit the marker",
		"never infer or retrofit ownership from an existing section title, task wording, comment, or legacy verification prose",
		"never emit a duplicate, malformed, nested, misplaced, or non-terminal owner marker",
	)
}

func TestOpenSpecVerifierGuardedTaskCompletionContract(t *testing.T) {
	verifier := readRepositoryAsset(t, "agents", "openspec-verifier.md")
	frontmatter, body, found := strings.Cut(strings.TrimPrefix(verifier, "---\n"), "---\n")
	if !found {
		t.Fatal("verifier frontmatter is missing")
	}

	for _, contract := range []string{
		"edit: false",
		"write: false",
		`"*": "deny"`,
		`"pnpm validate:snapshots": "allow"`,
		`"pnpm test": "allow"`,
		`"pnpm typecheck": "allow"`,
		`"pnpm build": "allow"`,
		`"pnpm --filter web exec vitest run src/App.integration.test.tsx": "allow"`,
		`"mktemp *": "allow"`,
		`"shasum *": "allow"`,
		`"sort *": "allow"`,
		`"cmp *": "allow"`,
		`"xargs *": "allow"`,
		`"test *": "allow"`,
		`"git ls-tree *": "allow"`,
		`"git archive *": "allow"`,
		`"tar *": "allow"`,
		`"diff *": "allow"`,
		`"git status*": "allow"`,
		`"git diff*": "allow"`,
		`"rg *": "allow"`,
		`"go test *": "allow"`,
		`"go build *": "allow"`,
		`"openspec status *": "allow"`,
		`"openspec validate *": "allow"`,
		`"angel-ai verifier-tasks snapshot --change *": "allow"`,
		`"angel-ai verifier-tasks complete --change *": "allow"`,
	} {
		if !strings.Contains(frontmatter, contract) {
			t.Errorf("missing verifier permission contract %q", contract)
		}
	}
	for _, forbidden := range []string{
		"bash:\n    \"*\": \"allow\"",
		`"pnpm *": "allow"`,
		`"sh *": "allow"`,
		`"bash *": "allow"`,
		`"zsh *": "allow"`,
	} {
		if strings.Contains(frontmatter, forbidden) {
			t.Errorf("verifier permission contract must not contain broad authorization %q", forbidden)
		}
	}

	requireTextInOrder(t, body,
		"Before verification, capture fresh resolved task state with `angel-ai verifier-tasks snapshot --change <name>`",
		"for an explicit store, append `--store <id>`",
		"Snapshot capture represents valid unmarked documents and valid marked sections with no pending tasks",
		"continue ordinary verification in either state and do not invoke completion",
		"For every pending task returned in the marked section snapshot, record one task-specific evidence entry using the exact task identity.",
		"Evidence must cite applicable commands you actually executed, their exit codes, and successful results",
		"Only when the snapshot has a valid marker and a non-empty exact pending task set",
		"after a global `pass` with successful executed evidence covering every task in that set",
		"may you invoke `angel-ai verifier-tasks complete --change <name>`",
		"the exact `snapshot`, `verdict`, `tasks`, and `evidence` JSON on stdin",
		"This guarded operation is the only permitted mutation",
		"never edit, fix, reformat, or write product code or any tracked/project file",
		"Generic edit and write tools remain disabled.",
		"Shell redirection and pipelines may write only verifier-assigned baseline,",
		"hash, or archive-comparison data under an external temporary directory created",
		"with `mktemp`; all other mutating shell commands, command wrappers, and",
		"arbitrary shell writes remain forbidden.",
		"On `fail`, `not-verified`, incomplete task evidence, or red final evidence, do not invoke completion",
		"report `completion: not-attempted` with no checkbox changes",
		"If verification passes but guarded completion reports a conflict",
		"report `status: blocked` and `completion: conflict`",
		"do not claim or attempt to mark any task",
		"Return every authoritative Shared corrected-failure result field supplied in the orchestrator prompt",
		"ordered failed/correction/success evidence",
		"equivalent-or-broader scope coverage",
		"final relevant validation state",
		"files touched",
		"deviations, scope expansion, or out-of-scope evidence",
		"`verdict` (`pass`, `fail`, or `not-verified`)",
		"per-task `evidence`",
		"`completion`, `conflicts`",
		"findings ordered by severity with file:line references",
		"scenario→evidence coverage summary",
	)
}

func TestOpenSpecWorkerAssetsAreCatalogedAndInstalledUnchanged(t *testing.T) {
	assetsRoot := filepath.Join("..", "..", "assets")
	assetSource := assetfs.Directory(assetsRoot)
	categories, err := catalog.Load(assetSource)
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]string{
		"openspec-implementer": "openspec-implementer.md",
		"openspec-planner":     "openspec-planner.md",
		"openspec-verifier":    "openspec-verifier.md",
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
		t.Fatalf("selected OpenSpec worker assets = %v, want %d", selected, len(want))
	}

	configDir := t.TempDir()
	if _, err := ApplyInstallation(InstallationRequest{
		Items: selected, Assets: assetSource, ConfigDir: configDir,
	}); err != nil {
		t.Fatal(err)
	}
	for _, item := range selected {
		wantContent, err := assetSource.ReadFile(item.Source)
		if err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(filepath.Join(configDir, item.Dest))
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, wantContent) {
			t.Errorf("installed %s differs from selected asset", item.Name)
		}
	}
}
