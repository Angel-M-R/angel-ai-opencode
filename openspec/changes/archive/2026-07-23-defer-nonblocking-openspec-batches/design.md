## Context

The orchestrator currently applies one shared strict result policy to planned OpenSpec task batches, Direct Safe work, Direct review fixes, and several OpenSpec control points. Any `partial`, `blocked`, deviation, or uncorrected command failure therefore interrupts the planned-task loop even when later plan sections are demonstrably independent. The change is limited to the section-bounded loop driven by the active change's resolved `tasks.md`; all other routes and phases retain their current strict stops.

The canonical policy lives in `assets/agents/angel-orchestrator.md`, with textual contract coverage in `internal/install/agent_assets_test.go`. OpenSpec fresh state and unchecked task boxes remain the source of truth for incomplete work.

## Goals / Non-Goals

**Goals:**

- Continue safe, independent planned work after a locally deferrable batch outcome.
- Preserve complete evidence for every deferred incident and benign deviation.
- Retry deferred batches exactly once after the independent forward pass.
- Keep incomplete or red work visibly unchecked and prevent final verification while tasks remain pending.
- Allow focused tests of modified code during implementation while reserving full suites and builds for the verifier.
- Keep the prompt and its contract tests synchronized.

**Non-Goals:**

- Relax Direct Safe, Direct Fast, review-fix, bootstrap, target-resolution, or final-verification behavior.
- Infer independence merely from section order or continue when dependency evidence is ambiguous.
- Ignore red validation, convert pending tasks to complete, add automatic task repair, or broaden a batch's functional scope.
- Change product code, dependencies, release behavior, or OpenSpec's artifact lifecycle.

## Decisions

### Split planned-batch handling from the shared strict policy

The orchestrator will retain the shared strict policy as the default and introduce an explicit planned-task-loop exception. Only results from section-bounded tasks selected from fresh `tasks.md` may enter deferral classification. Direct work, finding-ID review fixes, bootstrap, active-change resolution, and final verification bypass the deferral path.

Alternative considered: weaken the shared policy for every route. Rejected because it would silently relax verification and recovery guarantees outside the requested loop.

### Use a narrow deferred-batch record

A planned batch may be queued only when its outcome is local `partial`, local `blocked`, or a red focused test, its unfinished tasks remain unchecked, and no hard blocker is present. The record will retain the section and task identifiers, fresh checkbox state, worker status, every command and exit code, focused-validation state, blocker or incomplete-work reason, files touched, and deviations.

Additional reads and successful focused tests are benign, continuable deviations when they stay within the batch's read and validation purpose. Writes outside the batch, functional expansion, destructive commands, unresolvable OpenSpec state, or any checked task whose relevant validation remains red are hard blockers and never enter the queue.

Alternative considered: queue any non-clean result. Rejected because that would conceal scope violations and task-state conflicts.

### Require affirmative independence evidence

Before skipping a deferred section for a later one, the orchestrator will refresh OpenSpec state and identify explicit evidence that the later pending section does not consume, validate, or depend on the deferred work. Evidence must come from current planning artifacts, bounded task scopes, or retained worker diagnostics; section ordering or silence is insufficient. If the evidence is missing, conflicting, or ambiguous, the later section is treated as dependent and is not dispatched.

Alternative considered: assume separate named sections are independent. Rejected because section boundaries express batching, not necessarily dependency boundaries.

### Run one forward pass and one final retry round

During the forward pass, the orchestrator records deferrable incidents and continues only into explicitly independent later sections. After all currently independent work has been considered, it refreshes state, renders one combined deferred-incident and benign-deviation report, and starts one final retry round without asking an intermediate question. Each still-pending deferred batch receives at most one retry, selected from fresh state and bounded to its unchecked tasks.

The retry round may continue among mutually independent deferred batches, but it never creates a second queue or retry cycle. Hard blockers stop immediately. At round end, any pending batch with no runnable work, unresolved local block, or final red evidence triggers the existing report-first mandatory-stop interaction. If all tasks are complete and evidence is green, normal final verification begins.

Alternative considered: retry each batch immediately or indefinitely. Rejected because immediate retries prevent independent progress and repeated retries can loop without new evidence.

### Separate focused implementation validation from final verification

Implementer prompts will require validation relevant to the bounded batch and may run focused tests that exercise modified code. They will continue to prohibit full repository suites and builds. The verifier remains solely responsible for mandatory full suites, builds, and final evidence after fresh state shows every planned task complete.

### Enforce checkbox and red-evidence consistency

A `partial`, `blocked`, or red focused-test result is deferrable only while corresponding incomplete tasks remain unchecked. If fresh `tasks.md` marks a task complete while its relevant validation is still red, the orchestrator treats this as a real state conflict and stops rather than editing or reinterpreting the checkbox.

### Lock behavior with focused contract tests

Contract tests will assert the route boundary, deferral eligibility, conservative independence gate, single combined report, one retry round, focused-test allowance, checkbox integrity, and unchanged strict paths. Existing assertions that encode an immediate stop for every planned-batch non-clean result will be revised without weakening Direct or final-verification assertions.

## Risks / Trade-offs

- **[Independence can be difficult to prove]** → Default to dependent and stop advancing rather than infer safety.
- **[Deferred evidence can become stale]** → Apply the fresh-state invariant before every later dispatch, before the combined report, and before each retry.
- **[Benign deviations could mask scope expansion]** → Limit continuable deviations to additional reads and successful focused tests serving the bounded batch; treat writes and functional expansion as blockers.
- **[A retry round could become an implicit loop]** → Track retry eligibility per deferred batch and prohibit re-queuing or a second retry.
- **[Checkboxes can disagree with validation]** → Give fresh checkbox state and retained red evidence equal conflict-detection roles; stop rather than normalize either one.

## Migration Plan

1. Revise the canonical orchestrator policy and result-classification language.
2. Update focused asset contract tests in the same implementation batch structure.
3. Run focused validation during implementation as permitted, leaving full suites and builds to final verification.
4. Roll back by reverting the prompt and matching contract-test changes together, restoring immediate planned-batch stops.

## Open Questions

None. The confirmed Brief defines the deferral boundary, retry count, validation ownership, and strict exclusions.
