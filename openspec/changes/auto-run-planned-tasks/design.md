## Context

The orchestrator currently renders the OpenSpec task tree and then asks the user to choose task, section, or run-all cadence. Two choices deliberately pause after clean work, while run-all already uses the desired safe shape: one incomplete section per worker, followed by a fresh OpenSpec state read. The repository asset is canonical, but the installed global copy is currently not identical and must be replaced during implementation.

## Goals / Non-Goals

**Goals:**

- Make automatic execution the only planned-task flow without a cadence or continuation question between clean section batches.
- Keep every dispatch bounded to the pending tasks in exactly one incomplete section.
- Re-resolve OpenSpec status and `tasks.md` before each dispatch and after each clean result.
- Preserve the complete task tree before implementation and show refreshed state at every mandatory stop.
- Keep implementation validation lightweight and textual while requiring tests and build in final verification.
- Deploy one canonical instruction body to the repository and installed global locations.

**Non-Goals:**

- Changing Direct execution, post-verification review selection, finding-ID fix routing, archive behavior, or bootstrap behavior.
- Changing OpenSpec task syntax, adding cadence configuration, or adding automated tests.
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

### Separate batch checks from final verification

Planned-task implementer prompts require only focused textual checks relevant to the edited instructions, such as checking required and forbidden phrases and comparing synchronized files. They explicitly defer the repository's mandatory test and build commands to `openspec-verifier`. Any command that is run during a batch still falls under the existing non-zero-command stop policy.

Running mandatory tests and build after every section was rejected because it duplicates final verification and slows automatic progression without changing the stop semantics.

### Deploy from the canonical repository asset

Implementation edits `assets/agents/angel-orchestrator.md`, validates it textually, then overwrites `$HOME/.config/opencode/agents/angel-orchestrator.md` from that source without a backup and verifies byte equality. Rollback uses version control for the repository asset and repeats the same replacement operation.

## Risks / Trade-offs

- [Automatic execution reduces opportunities for discretionary user interruption] → Keep sections bounded and preserve all mandatory stops so control returns on any unsafe result.
- [Removing cadence text may leave contradictory references elsewhere in the long instruction file] → Use focused positive and negative textual checks before deployment.
- [A mandatory stop may coincide with unreadable OpenSpec state] → Surface the state-resolution failure and stop evidence rather than fabricating a task tree.
- [The installed file may drift again after deployment] → Treat the repository asset as canonical and verify byte identity during implementation and final verification.

## Migration Plan

1. Rewrite the canonical planned-task instructions and related prompt/result-policy wording.
2. Run focused textual checks for required automatic behavior and removed cadence behavior.
3. Replace the installed global agent from the canonical asset and confirm byte equality.
4. Run mandatory repository tests and build only through final OpenSpec verification.

Rollback restores the repository asset from version control, replaces the installed global file from it, and confirms equality.

## Open Questions

None.
