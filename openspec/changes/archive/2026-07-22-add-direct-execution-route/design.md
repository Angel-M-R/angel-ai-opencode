## Context

The repository-managed orchestrator is a thin coordinator whose current non-trivial change flow proceeds from a confirmed interview Brief into OpenSpec planning, bootstrap, bounded implementation, verification, review, and archive. The new route must introduce direct implementation without weakening or accidentally entering that OpenSpec state machine. The behavior is encoded in the orchestrator Markdown asset and guarded by textual contract tests.

## Goals / Non-Goals

**Goals:**

- Place one explicit execution-route decision after Brief confirmation and before any new OpenSpec change creation or OpenSpec bootstrap.
- Recommend a route from change characteristics while leaving the user's choice authoritative.
- Define bounded Safe and Fast direct modes delegated only to `general`.
- Make Safe-mode verification, stopping, reviews, and selected-finding fixes deterministic.
- Apply one evidence-based corrected-intermediate-failure policy to planned OpenSpec implementation batches and Direct Safe workers.
- Keep existing changes and OpenSpec selections on the complete current OpenSpec workflow.
- Add contract tests that detect omissions, wrong ordering, and accidental cross-routing.

**Non-Goals:**

- Replacing OpenSpec planning, implementation cadence, verification, review, or archive behavior.
- Running OpenSpec bootstrap, CLI commands, or artifact generation in direct mode.
- Allowing the orchestrator or `openspec-implementer` to implement direct work.
- Adding orchestrator retries, fallback workers, persisted route preferences, new reviewers, or a direct-mode archive concept.
- Changing the interview modes or the requirement to confirm a Brief.

## Decisions

### Route before any new-change side effect

For non-trivial work with a confirmed Brief, the orchestrator will first determine whether the request targets an already existing OpenSpec change. Existing changes bypass the new route choice and continue through status-driven OpenSpec handling. Otherwise, it will recommend Direct for clear, isolated, reversible work and OpenSpec for architecture, security, data, migrations, cross-cutting scope, or uncertainty, then ask a single OpenSpec-versus-Direct question. The recommendation is explanatory and never changes or blocks the selected route.

This ordering prevents bootstrap or artifact creation before the user has selected OpenSpec. Placing the gate earlier than Brief confirmation was rejected because the recommendation would lack confirmed scope; placing it after `openspec new change` was rejected because direct mode must have no OpenSpec side effects.

### Direct mode has a second bounded mode choice

After Direct is selected, the orchestrator asks for Safe or Fast. Both modes dispatch exactly one `general` worker to implement the confirmed Brief within an explicitly bounded scope. The prompt carries the Brief verbatim, selected mode, scope limits, and a return contract requiring `done|partial|blocked`, touched files, commands in order with exit codes, deviations, and any corrected-failure relationship between a failed command and its later equivalent-or-broader successful rerun.

Using the orchestrator inline was rejected because it violates its coordinator role. Using `openspec-implementer` was rejected because direct work has no OpenSpec tasks or artifacts and must remain outside the OpenSpec workflow.

### Safe mode uses one implementation-and-verification worker

The Safe prompt requires the same `general` worker to implement and run the repository's existing applicable tests/build. The orchestrator treats a Safe result as clean only when status is `done`, executable verification was run, the final relevant verification state is green, and no Brief deviation or out-of-scope work is reported. A non-zero command may still be clean only under the corrected-intermediate-failure policy below. Missing executable verification, an uncorrected or finally red result, `partial` or `blocked`, or any deviation causes an immediate report-and-stop with no orchestrator retry, fallback, review, or additional implementation dispatch.

A separate verifier was rejected because the confirmed design requires the implementing worker to execute verification and direct mode must not use OpenSpec workers. Orchestrator-issued retries were rejected because they could conceal scope drift or mutate the repository after a failed safety boundary; a same-worker corrective rerun reported in one result is evidence from the bounded batch, not an orchestrator retry.

### Corrected intermediate failures preserve evidence without forcing a stop

Apply the same classification to every planned OpenSpec implementation batch and Direct Safe worker result, including a bounded Direct Safe review-fix result. An intermediate non-zero command is corrected only when the same worker identifies the failure, later reports an equivalent or broader relevant command with exit code zero, returns final status `done`, and reports no deviation or out-of-scope work. Equivalent or broader means the successful rerun validates the failed command's relevant scope or a superset of it; an unrelated green command cannot correct the failure.

The result must preserve command order and identify the failed command, its exit code, and the successful equivalent-or-broader rerun so the orchestrator can retain and surface the complete corrected-failure evidence. For a planned OpenSpec batch, the orchestrator may then apply the fresh-state invariant and continue or pause under the already selected cadence. For Direct Safe work, it may proceed to the applicable review step. The orchestrator does not erase, relabel, or hide the intermediate failure.

Stopping remains mandatory when a non-zero command has no clean equivalent-or-broader rerun, the final relevant state is red, status is `partial` or `blocked`, the worker reports deviation or out-of-scope work, or a TDD/expected failure remains red at batch end. Treating every non-zero command as a stop was rejected because it turns a failure corrected within the same bounded worker run into a false terminal state. Ignoring intermediate failures was rejected because it removes audit evidence and could conceal an unrelated or insufficient rerun.

### Fast mode ends explicitly unverified

The Fast prompt prohibits tests and reviews. After the bounded `general` worker returns, the orchestrator reports the work as implemented but not verified and does not enter the review gate. Fast mode retains the result contract so blocked, partial, or deviating work is still reported accurately.

Implicitly treating a successful worker status as verification was rejected because no executed evidence exists.

### Safe reviews reuse choices but not OpenSpec fix routing

After a clean Safe result, the orchestrator presents the existing reviewer choices—Security risk, Simplicity, Correctness, or none—and runs selected reviewers against the bounded direct diff and confirmed Brief. It deduplicates findings and asks which findings to fix. Only selected findings are sent together as one bounded batch to a new `general` worker; `openspec-implementer` is prohibited. The fix worker receives the selected finding IDs and text, the direct Brief and scope, and the same structured result contract, including corrected-failure evidence when applicable. A non-clean fix result stops without retry. After a clean fix, only reviewers whose selected findings were addressed may be offered for rerun; no OpenSpec verification or archive action occurs.

Reusing `openspec-implementer` was rejected because no active change owns the findings. Rerunning every reviewer was rejected as unnecessary work and contrary to affected-reviewer routing.

### Contract tests assert behavior and ordering at the asset boundary

Extend the repository asset tests with section extraction and ordered-text assertions for the route gate, recommendation criteria, direct prompt contract, Safe and Fast semantics, review fixes, existing-change behavior, and the shared corrected-intermediate-failure policy. Assert both allowed continuation cases and every mandatory-stop case for Direct Safe workers and planned OpenSpec implementation batches. Preserve the existing bootstrap contract tests and assert that direct mode excludes bootstrap, OpenSpec CLI/artifacts, and OpenSpec workers while the OpenSpec branch still reaches the existing workflow sections.

Broad end-to-end agent simulation was rejected because the behavior is a managed instruction asset; focused contract tests provide deterministic regression coverage consistent with existing tests.

## Risks / Trade-offs

- **Textual tests can pass while agent interpretation remains imperfect** → Assert explicit normative wording, forbidden routing, and section order at each decision boundary.
- **A route recommendation may be subjective** → Encode conservative criteria and keep the user's explicit choice authoritative.
- **A worker could claim an unrelated green command corrected a failure** → Require an explicit equivalent-or-broader relationship, chronological command evidence, final `done`, and no deviation or out-of-scope work.
- **A genuinely red batch could be allowed to continue** → Stop on any uncorrected failure, final red state, incomplete status, deviation, out-of-scope work, or TDD/expected failure still red at batch end.
- **Fast mode can ship defects** → Prohibit verification claims and report prominently that implementation is unverified.
- **Direct review scope may drift** → Bind reviewers and the fix worker to the confirmed Brief and direct worker's bounded diff/result.
- **New instructions could disturb OpenSpec behavior** → Keep direct sections before OpenSpec entry, route existing changes directly to OpenSpec, and retain focused preservation assertions for the current workflow.

## Migration Plan

1. Add failing contract tests for the route, shared corrected-failure policy, mandatory-stop conditions, and preservation boundaries.
2. Update the orchestrator asset to satisfy those contracts without restructuring the existing OpenSpec sections or changing the selected cadence model.
3. Run the focused asset tests and the repository's existing full test/build checks.
4. Roll back by reverting the orchestrator and its new tests together; no data or generated artifacts require migration.

## Open Questions

None.
