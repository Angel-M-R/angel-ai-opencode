## Context

The orchestrator previously rendered the OpenSpec task tree and then asked the user to choose task, section, or run-all cadence. Automatic section-bounded execution has now been deployed, and the repository and installed global copies are synchronized. Final verification exposed two stale cadence assertions in `internal/install/agent_assets_test.go`. It also exposed a route-wide interaction gap: stop clauses report evidence and prohibit automatic recovery, but they do not consistently require a contextual next-action question after that report.

## Goals / Non-Goals

**Goals:**

- Make automatic execution the only planned-task flow without a cadence or continuation question between clean section batches.
- Keep every dispatch bounded to the pending tasks in exactly one incomplete section.
- Re-resolve OpenSpec status and `tasks.md` before each dispatch and after each clean result.
- Preserve the complete task tree before implementation and show refreshed state at every mandatory stop.
- Keep implementation validation focused while reserving the full mandatory test suite and build for final verification.
- Make every mandatory stop across OpenSpec and Direct routes report blocking evidence before asking one contextual next-action question.
- Keep the orchestrator idle after a mandatory stop until the user selects an action, with a safe stop option always available.
- Add focused contract coverage for stop ordering, user gating, and the absence of automatic recovery, while correcting the two stale cadence assertions.
- Deploy one canonical instruction body to the repository and installed global locations.
- Repeat full tests, build, synchronization checks, and OpenSpec verification after the revised implementation.

**Non-Goals:**

- Changing clean-path Direct execution, post-verification review selection, finding-ID fix routing, archive behavior, or successful bootstrap behavior.
- Changing OpenSpec task syntax or adding cadence configuration.
- Weakening blocked, partial, command-failure, deviation, out-of-scope, or state-conflict stops.

## Decisions

### Replace cadence state with one automatic loop

Delete the cadence question, mode labels, retained-selection state, and cadence-boundary return behavior. The planned-task cycle renders the initial complete tree, selects the next incomplete section from fresh state, dispatches that exact section, classifies the result, refreshes state, and repeats automatically until completion or a mandatory stop.

This is preferred over retaining a hidden `run-all` mode because a single unconditional loop removes dead branches and prevents obsolete cadence terminology from continuing to shape the instructions.

### Preserve section-bounded dispatches

Each implementer prompt names one section and enumerates only its pending task identifiers and summaries. A fresh state read occurs before dispatch and after every clean result; a stale section completed before dispatch is skipped, while an unexpected conflict during or after dispatch remains a mandatory stop.

An unbounded “finish all tasks” prompt was rejected because it would weaken task-state reconciliation and increase the impact of an unsafe worker result.

### Render the tree at entry and stops, not between clean sections

The complete compact tree remains mandatory immediately before implementation begins. Clean automatic transitions do not render the tree or return control. On every mandatory stop, the orchestrator refreshes state when it can do so safely, renders the complete tree, and surfaces the stop evidence; if state cannot be resolved, it reports that tree rendering is unavailable alongside the blocking evidence.

This preserves user visibility where action is needed without recreating cadence pauses.

### Use one report-first, question-second stop protocol across routes

Every mandatory stop uses the same ordered protocol regardless of whether it originates in OpenSpec target resolution, bootstrap, planned-task implementation, final verification, Direct Safe or Fast execution, or a Direct review-fix batch. The orchestrator first reports the blocking status and retained evidence, including command and exit-code details when applicable. Only after that report does it ask exactly one contextual next-action question with the `question` tool.

The choices are derived from the blocker rather than imposed as one universal menu, but every question includes a safe stop option and continues to permit the tool's custom response. Reporting and questioning remain separate ordered steps so the user can make a decision from complete evidence.

This is preferred over a generic “continue?” prompt because retrying a command, adjusting scope, fixing state, or abandoning the route are not interchangeable actions. It is preferred over merely ending with “stop for user direction” because that leaves the interaction contract implicit and inconsistent across routes.

### Gate every recovery action on explicit user selection

After asking the stop question, the orchestrator does not retry, continue, broaden scope, reinterpret the plan, or dispatch any worker until the user selects an action. A selected action may authorize a new bounded step, but the orchestrator must not infer that authorization from the blocker or from a custom response it cannot map safely.

This preserves user control without weakening automatic progression on clean paths. Clean planned-task sections still chain automatically, and successful verification still reaches the existing review gate.

### Separate focused contract checks from final verification

Planned-task implementer prompts require focused textual checks relevant to the edited instructions, such as checking required and forbidden phrases and comparing synchronized files. The test batch updates `internal/install/agent_assets_test.go`, corrects the two stale cadence assertions, and adds explicit ordered-text and negative contract coverage for the shared stop protocol. Focused contract tests may run during implementation, but the repository's full mandatory test suite and build remain final-verification obligations. Any command that is run during a batch still falls under the existing non-zero-command stop policy.

Running mandatory tests and build after every section was rejected because it duplicates final verification and slows automatic progression without changing the stop semantics.

### Keep contract coverage in the existing asset test

Extend `internal/install/agent_assets_test.go` rather than introducing a new test file. Its existing section helpers already verify ordered instruction contracts and can prove that evidence precedes the question, that route-specific stop clauses reference the shared protocol, and that forbidden automatic-recovery language is absent. The two assertions that still require selected cadence wording are replaced with automatic bounded-implementation wording.

This keeps the asset contract concentrated in one place and avoids adding a second parser or test abstraction for the same Markdown source.

### Deploy from the canonical repository asset

Implementation edits `assets/agents/angel-orchestrator.md`, validates it textually, then overwrites `$HOME/.config/opencode/agents/angel-orchestrator.md` from that source without a backup and verifies byte equality. Rollback uses version control for the repository asset and repeats the same replacement operation.

## Risks / Trade-offs

- [Automatic execution reduces opportunities for discretionary user interruption] → Keep sections bounded and preserve all mandatory stops so control returns on any unsafe result.
- [Removing cadence text may leave contradictory references elsewhere in the long instruction file] → Use focused positive and negative textual checks before deployment.
- [A mandatory stop may coincide with unreadable OpenSpec state] → Surface the state-resolution failure and stop evidence rather than fabricating a task tree.
- [A generic recovery menu may suggest an unsafe action for a specific blocker] → Derive options from the reported blocker and always include a safe stop option.
- [Question wording could imply authorization before the user answers] → Contract-test report-first/question-second ordering and explicit prohibition of recovery before selection.
- [Route-specific stop clauses may drift from the shared protocol] → Require every stop route to reference the shared protocol and cover those references in the existing asset test.
- [The installed file may drift again after deployment] → Treat the repository asset as canonical and verify byte identity during implementation and final verification.

## Migration Plan

1. Rewrite the canonical shared mandatory-stop protocol and make every OpenSpec and Direct stop route reference it.
2. Correct the two stale cadence assertions and add focused route-wide ordering and no-automatic-recovery contract coverage.
3. Run focused textual and contract checks against the canonical asset.
4. Replace the installed global agent from the canonical asset and confirm byte equality.
5. Repeat focused checks, the mandatory repository tests and build, synchronization checks, and OpenSpec verification.

Rollback restores the repository asset from version control, replaces the installed global file from it, and confirms equality.

## Open Questions

None.
