## Context

The orchestrator has one shared strict-default implementation-result policy, route-local consumers for Direct Safe and OpenSpec phases, and a separate planned-task exception with bounded repair, deferral, independence, and retry safeguards. Today the shared clean-result classification is limited to corrected syntax or invocation mistakes, even when a worker repairs a real in-scope cause and subsequently produces stronger green evidence. This creates unnecessary authorization stops and can block normal archive progression despite a clean final result.

The change is policy- and contract-focused: the installed orchestrator asset is the behavior source, and installer tests assert its required language and routing. The confirmed Brief requires every strict-default consumer to use the broadened shared classification without changing any unrelated implementation, review, cadence, verification, or archive rule.

## Goals / Non-Goals

**Goals:**

- Classify both tooling mistakes and real intermediate failures by the same evidence-complete correction criteria.
- Preserve an ordered audit trail of the failed command, diagnosed cause, bounded correction, and successful equivalent-or-broader validation.
- Let every strict-default consumer continue its existing next step when that classification is clean, including normal archive progression after final verification.
- Keep unresolved outcomes and every existing scope, status, deviation, evidence, and final-red guard as mandatory stops.
- Preserve all planned-task-only repair, deferral, dependency, checkbox, cadence, and retry safeguards.

**Non-Goals:**

- Changing which workers may edit files, widening any worker's authorized scope, or allowing cross-worker correction evidence.
- Adding orchestrator retries to strict-default routes or treating a narrower or unrelated green command as correction.
- Changing review selection, implementation cadence, final-verification coverage, archive eligibility, or any rule unrelated to result classification.
- Implementing a new runtime policy engine or changing public APIs or dependencies.

## Decisions

### 1. Make evidence, not failure category, determine correction eligibility

The shared policy will remove the tooling-only gate and use one conjunctive eligibility rule. A corrected incident requires the same worker, in the same bounded invocation, to identify the cause, repair it entirely within authorized scope, run relevant validation that covers the failed command's scope or a superset with exit code zero, finish with status `done` and a green final relevant state, and report no deviation, scope expansion, or out-of-scope work.

The retained result contract will include the original command and non-zero exit code, the diagnosed cause and correction, and the later successful command and exit code in execution order. Missing or ambiguous scope-equivalence, correction, command, or exit-code evidence remains non-clean.

Alternative considered: enumerate acceptable real failure types. Rejected because type-based allowlists are incomplete and weaker than proving bounded repair plus final relevant validation.

### 2. Keep one shared clean-result classification for every strict-default consumer

All control points that invoke the shared implementation-result policy will consume the same broadened classification. An eligible corrected incident is clean for routing purposes: the control point surfaces the incident and takes its already-defined next action rather than stopping or asking authorization solely because the intermediate command failed. For final verification, that means following the existing archive path when all other archive prerequisites hold.

Route-local requirements remain additive. For example, Direct Safe still requires executable verification, review-fix workers remain bounded to selected findings, final verification still requires its full coverage, and archive still depends on the existing completed state. The shared policy changes only how an intermediate failure is classified.

Alternative considered: duplicate corrected-real-failure language in each route. Rejected because duplicated criteria can drift and would undermine the shared-policy intent; route tests should instead prove that each consumer references and honors the common classification.

### 3. Preserve the planned-task exception as a non-clean-result overlay

The planned-task loop will first recognize evidence-complete corrected incidents as clean under the shared rule. Its existing special logic continues to govern results that are not clean: bounded self-repair cycles, incomplete-task checkbox handling, eligible local deferral, conservative independence, the selected cadence, the single retry round, and hard-stop conflicts all remain unchanged.

Alternative considered: fold planned-task deferral into the shared strict policy. Rejected because that would weaken safeguards at Direct, bootstrap, target-resolution, review-fix, post-verification-fix, final-verification, or archive-related control points.

### 4. Test the common criteria and every consumer boundary

Contract tests will assert the complete corrected-real-failure evidence rule once in the shared section, retain negative assertions for every mandatory-stop condition, and verify each strict-default consumer uses the common clean/non-clean outcome. Focused cases will cover normal continuation, final-verification-to-archive progression, same-worker identity, equivalent-or-broader coverage, final status/state, deviations, scope expansion, and insufficient evidence. Planned-loop tests will continue to assert its narrower exception and safeguards.

Alternative considered: test only the shared paragraph. Rejected because a correct definition can still be bypassed or contradicted at a route-local control point.

## Risks / Trade-offs

- [A worker overstates that a rerun covers the failed scope] → Require explicit ordered commands, exit codes, cause, correction, and scope-equivalence evidence; ambiguity remains a mandatory stop.
- [Broader clean classification accidentally permits scope growth] → Keep authorized-scope repair, no-deviation, no-scope-expansion, and no-out-of-scope-work conditions conjunctive and test each negative case.
- [Route-local wording preserves the old tooling-only restriction] → Audit every shared-policy consumer and add contract coverage for consistent classification and continuation.
- [Planned-task deferral safeguards are weakened during simplification] → Treat the planned exception as an unchanged overlay for non-clean results and retain its existing focused tests.
- [Archive appears automatic after any corrected failure] → State that only the incident-based stop is removed; every existing verification, review, completion, and archive prerequisite still applies.
