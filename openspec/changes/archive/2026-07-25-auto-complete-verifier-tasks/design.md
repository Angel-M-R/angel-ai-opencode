## Context

The automatic planned-task loop currently treats every pending named section as implementation work and enters final verification only after all task checkboxes are complete. The verifier is configured without general edit or write tools and is contractually code read-only. This leaves no safe representation for terminal tasks that can be completed only by final verification, and attempting to send those tasks through the implementer either violates ownership or creates a second verification pass.

The workflow already resolves active state through `openspec status --change <name> --json`, carries `--store <id>` for explicit stores, and reserves full repository suites and builds for the verifier. This design adds a narrowly guarded `tasks.md` commit after verification while leaving `accept-corrected-worker-failures` and the shared result-classification policy unchanged.

## Goals / Non-Goals

**Goals:**

- Give planners an explicit, machine-detectable contract for one terminal verifier-owned section.
- Dispatch the verifier automatically when that marked section is the only pending work.
- Preserve code read-only verification while permitting one constrained mutation of the freshly resolved `tasks.md`.
- Make completion evidence-gated, fresh-state checked, conflict-aware, and atomic across all tasks in the section.
- Support both repo-local changes and explicitly selected stores without path inference.
- Reuse one successful verifier result and continue to the existing review gate after fresh overall-completion confirmation.
- Keep planned implementer batches focused; full suites and builds remain final-verifier work.

**Non-Goals:**

- Inferring ownership from section titles, task wording, or legacy conventions without the exact marker.
- Allowing the verifier to edit product code, planning prose, any file other than the resolved active `tasks.md`, or any checkbox outside the marked section.
- Supporting multiple, nested, or non-terminal verifier-owned sections.
- Changing shared corrected-worker-failure classification, review choices, archive behavior, or the separate `accept-corrected-worker-failures` change.
- Making this change's own completion depend on marker-aware behavior before that behavior is installed.

## Decisions

### 1. Use one exact structural marker on the terminal task section

The planner contract will permit exactly one `<!-- owner: openspec-verifier -->` marker as the first nonblank line under a named top-level task-section heading, and that section must be the final task section. Detection is structural and exact: a duplicate, misplaced, non-terminal, or malformed marker is invalid, while verifier-like prose without the marker remains ordinary implementation work.

The planner will use the marker only when the section consists exclusively of final-verification obligations that the verifier can prove through executed evidence. It must keep setup, implementation, focused tests, and contract updates in ordinary sections. This change's own `tasks.md` will deliberately contain no owner marker so the currently installed workflow can execute it.

Alternative considered: infer ownership from headings such as “Verification.” Rejected because wording is ambiguous, breaks legacy compatibility, and contradicts the explicit-marker requirement.

### 2. Treat the marked section as a terminal scheduler boundary

On each fresh planned-task decision, the orchestrator will resolve the current status and resolved `tasks.md` in the active local or explicit-store context, parse the marker contract, and exclude the marked section from implementer batches. Earlier pending ordinary sections continue through the existing bounded loop. When every ordinary task is checked and the marked section alone remains pending, the orchestrator automatically dispatches exactly one verifier with the change name and the same context/store identity.

An invalid marker layout, a checked-task/red-evidence conflict, or an inability to resolve fresh state is a stop rather than a fallback to the implementer. If no exact marker exists, the current workflow remains unchanged and no section is inferred as verifier-owned.

Alternative considered: send the verifier-owned section through the implementer and verify later. Rejected because the tasks are not implementation work and this would force duplicate final verification.

### 3. Keep general verifier writes disabled and expose one guarded completion operation

The verifier will retain no general edit or write capability. Its only mutation route will be a purpose-built completion operation that accepts the change name, optional explicit store id, the exact marked section and task identities, and the verifier's structured task-evidence map and global verdict. The operation independently resolves OpenSpec state and rejects any target other than the returned `artifactPaths.tasks.resolvedOutputPath` for that same change and context.

The operation will permit only pending-to-checked checkbox transitions for every task in the one valid marked section. It will reject content edits, checkbox changes outside the section, partial task sets, path overrides, symlink/path escape, a different change/store, and attempts to target code. Verifier policy will prohibit every other mutating shell command, while existing general `edit` and `write` tools remain disabled.

Alternative considered: enable a general file-edit tool and rely only on prompt wording. Rejected because it unnecessarily weakens the verifier's code read-only guarantee and cannot enforce the resolved-file boundary.

### 4. Bind atomic completion to evidence and a fresh-state compare

The verifier first captures a resolved snapshot containing context identity, absolute normalized `tasks.md` path, full file digest/content, marker structure, and exact pending task identities. It runs verification and records successful executed evidence separately for each marked task. The guarded operation is callable only with global verdict `pass` and a complete evidence map; `fail`, `not-verified`, missing task evidence, non-zero final evidence, or any incomplete task mapping produces no write.

Immediately before mutation, the operation reruns status with the same local context or `--store <id>`, resolves and rereads `tasks.md`, and compares the fresh context, path, full content digest, marker structure, and target task set with the captured snapshot. Any difference is a stale/concurrent-state conflict and leaves the file unchanged. The operation then builds one full replacement that changes only the target checkbox tokens, writes a sibling temporary file, preserves required file metadata, and atomically replaces the resolved file. Cleanup on any pre-rename failure leaves the original intact; no task is marked unless all target tasks are committed together.

The verifier result distinguishes verification verdict from completion state, so a technically passing verification plus a write conflict returns `status: blocked`, `verdict: pass`, and `completion: conflict` rather than claiming workflow completion.

Alternative considered: update each checkbox as its evidence becomes available. Rejected because a later failure or conflict would leave a partially closed section.

### 5. Reuse the pass only after fresh overall completion

After a verifier reports `status: done`, `verdict: pass`, and atomic completion success, the orchestrator reruns fresh status in the same context and confirms that the tasks artifact is complete and no checkbox remains pending. It then treats the already returned pass and command evidence as final verification and enters the existing review gate without dispatching the verifier again.

Any failed, incomplete, conflicting, non-clean, or still-pending result follows the existing stop policy and cannot reach review or archive. This continuation rule does not alter corrected-failure classification; it only avoids redundant verification after the successful verifier-owned commit.

Alternative considered: dispatch a second verifier after checkbox closure. Rejected because the first pass already produced final evidence and the second run adds cost and a new race without improving the completion proof.

### 6. Test contracts and behavior at focused layers

Asset contract tests will cover planner marker rules, orchestrator scheduling/context propagation/pass reuse, and verifier permissions/result shape. Focused operation tests will use temporary local and explicit-store fixtures to cover exact marker parsing, path confinement, evidence mapping, fresh-state changes, task-set changes, atomic all-or-nothing writes, and success continuation. Implementer batches will run only affected focused tests; full repository suites and builds remain tasks for final verification.

Alternative considered: rely only on text-contract assertions. Rejected because concurrency, path resolution, and atomic mutation require executable behavioral coverage.

## Risks / Trade-offs

- [A marker is hand-authored incorrectly] → Fail closed on any duplicate, malformed, misplaced, or non-terminal marker and report the structural conflict.
- [A store resolves a path outside the caller's working directory] → Trust only the successful status result for the explicit store and bind the operation to its normalized resolved path and `store:<id>` identity.
- [State changes during verification] → Compare the entire fresh resolved snapshot immediately before the atomic replacement and abort without checkbox changes on any mismatch.
- [A writer ignores the guarded operation during the final commit window] → Keep the commit window minimal, use one atomic replacement, and require workflow-owned task writers to honor the same guarded state protocol; report detected divergence as conflict.
- [A passing command is mapped too broadly to several tasks] → Require explicit task-to-evidence entries and task-specific applicability, with missing or ambiguous coverage treated as incomplete.
- [Verifier mutation authority expands over time] → Keep generic write tools disabled and test path, transition, and content-diff rejection as security boundaries.
- [This proposal deadlocks under the old scheduler] → Leave this change's own tasks unmarked and make marker-dependent behavior an implementation outcome, not a prerequisite for executing this plan.
