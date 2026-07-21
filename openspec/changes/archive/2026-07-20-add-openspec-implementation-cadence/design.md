## Context

The orchestrator currently delegates implementation as bounded batches and asks whether to continue between phases, but it does not define a task-aware cadence selector. Because OpenSpec task completion can change outside the conversation, the selector must derive every batch from the current change state rather than a cached summary. This is an instruction-only change affecting the repository orchestrator asset and its installed global copy.

## Goals / Non-Goals

**Goals:**

- Make the current task state visible before implementation begins.
- Let the user choose task, section, or run-all cadence once per implementation cycle.
- Preserve bounded worker prompts and stop immediately on unsafe or inconclusive outcomes.
- Move directly to verification when all tasks are complete, followed by the existing review gate.
- Keep the repository and live global orchestrator instructions identical after implementation.

**Non-Goals:**

- Changing OpenSpec schemas, task syntax, planner or worker responsibilities, review semantics, or product runtime behavior.
- Persisting cadence across sessions or changes.
- Adding automated tests or backups of the live global agent file.

## Decisions

### Define one fresh-state invariant for planned-task implementation

At every planned-task decision point—before cadence selection, before dispatch, and after a clean worker result—the orchestrator will run `openspec status --change <name> --json` in the active context, retaining `--store <id>` for an explicit store, use the returned artifact path to read `tasks.md`, and recompute the tree and next batch. All later refresh requirements reference this single invariant. It will never select tasks from a prior message or an in-memory copy. If status cannot resolve a complete tasks artifact or the file cannot be read, planned-task implementation stops as blocked.

This uses OpenSpec's existing source of truth instead of introducing a parallel tracker. The alternative—parsing a previously pasted task list—was rejected because external completion changes would make batching stale.

### Render one compact, complete tree before cadence selection

The orchestrator will map task sections and markdown checkboxes into a tree. The root and every section show completed/total counts; every task shows its identifier, a short version of its text, and either `✓` for complete or `☐` for pending. The complete tree is shown before the cadence question and again whenever control returns after a task- or section-level pause.

The alternative of listing only pending tasks was rejected because the user would lose section context and evidence of completed work.

### Ask for cadence once and retain it for the current implementation cycle

The orchestrator will ask one single-select question with exactly three semantic choices: pause after each task, pause after each section, or run all remaining tasks. The choice is session-only and applies until implementation completes or a mandatory stop occurs. Task and section modes return control after their selected boundary and wait for an explicit continuation without asking the cadence question again. Run-all mode continues automatically.

The alternative of asking after every batch was rejected because it creates repetitive interaction and makes run-all ineffective.

### Keep every worker dispatch bounded and explicit

Task cadence sends exactly the next pending task. Section cadence sends the pending tasks in the next incomplete section as one explicit batch. Run-all uses the same section-bounded batches and chains them automatically; it never sends an unbounded "finish everything" prompt. Each dispatch uses the tree and exact pending identifiers produced by the fresh-state invariant.

Section boundaries provide a stable batching unit already present in `tasks.md`. If the fresh-state invariant shows that the intended batch is already complete, the orchestrator skips it and uses the recomputed next batch rather than dispatching stale work.

### Stop on any non-clean batch outcome

Automatic chaining and cadence progression stop immediately when an implementer reports `blocked` or `partial`, any invoked test/build command fails, the worker reports a deviation from the plan, or the fresh state produced by the invariant conflicts with the requested batch. The orchestrator surfaces the evidence and does not improvise a replacement batch. A clean `done` result alone does not prove overall completion; only the fresh-state invariant does.

### Verify automatically, then preserve the review gate

When the fresh-state invariant shows no pending tasks, the orchestrator dispatches `openspec-verifier` automatically without asking another continuation question. Failed or incomplete verification stops the flow. Successful verification proceeds to the existing post-verification review selection unchanged. A user-selected review-fix batch remains outside the planned-task cadence and keeps the existing finding-ID routing, bounded fix, and reviewer rerun behavior without reopening cadence or automatically repeating verification.

### Update the repository asset first, then replace the live global file

Implementation will edit `assets/agents/angel-orchestrator.md`, manually validate the revised instructions, and then overwrite `$HOME/.config/opencode/agents/angel-orchestrator.md` with the repository asset. No backup or sidecar copy will be created. A final byte-for-byte comparison will confirm synchronization.

The repository asset remains the canonical version. Editing only the global file was rejected because reinstalling would lose the behavior; creating a backup was rejected by the confirmed deployment decision.

## Risks / Trade-offs

- [A section can contain more work than an ideal batch] → Keep the prompt limited to one named section and its explicit pending task identifiers; stop rather than silently splitting if the worker cannot complete it cleanly.
- [External edits can change completion state during a run] → Apply the fresh-state invariant at every planned-task decision point.
- [Replacing the global file can discard local divergence] → Treat the repository asset as canonical and explicitly accept replacement without backup.
- [Instruction behavior has no automated regression coverage] → Perform manual scenario walkthroughs for all three cadences and every mandatory stop condition; add no tests.

## Migration Plan

1. Revise the repository orchestrator asset with the selector, batching, stop, verification, and review-gate rules.
2. Manually inspect and walk through task, section, run-all, blocked/partial, failed-check, deviation, completion, verification, and review transitions.
3. Replace the live global orchestrator file from the validated repository asset without creating a backup.
4. Confirm the two files are byte-for-byte identical.

Rollback requires restoring an earlier repository revision and copying that canonical asset to the global path; no snapshot of pre-change global-only divergence will exist.

## Open Questions

None.
