## Why

OpenSpec changes can reserve terminal tasks for final verification, but the current workflow cannot recognize, dispatch, and close that work without either treating it as ordinary implementation or requiring a second verification pass. A narrow, evidence-gated verifier-owned task protocol is needed to preserve the verifier's code read-only role while completing only its resolved task checkboxes safely.

## What Changes

- Define one optional terminal `tasks.md` section marked exactly `<!-- owner: openspec-verifier -->`; unmarked or legacy text is never inferred as verifier-owned.
- Require planners to emit the marker only for a single terminal verification section and keep ordinary implementation tasks executable by the currently installed workflow.
- Detect when the marked section is the only pending section and automatically dispatch the verifier in both local and explicit-store contexts.
- Narrow verifier mutation authority to the active change's freshly resolved `tasks.md`, while keeping all product code and every other file read-only.
- Revalidate resolved change and task state immediately before writing, aborting on stale, changed, missing, or conflicting state.
- Atomically check the entire verifier-owned section only when every task has task-specific executed evidence and the verifier's global verdict is `pass`; otherwise change no checkbox.
- Reuse the successful verification result, confirm fresh overall task completion, and continue directly to the existing review gate without a second verifier dispatch.
- Add contract and behavioral tests for planning, marker detection, local/store routing, permissions, evidence gating, atomicity, conflicts, and review continuity.
- Keep the separate `accept-corrected-worker-failures` change independent and preserve its result-classification scope.

## Capabilities

### New Capabilities

- `verifier-owned-task-completion`: Defines the marked terminal-section protocol, narrow verifier mutation authority, fresh-state and concurrency safeguards, evidence requirements, and atomic completion behavior.

### Modified Capabilities

- `execution-route-selection`: Route a change whose only pending work is the marked verifier-owned section through one final verification pass, then reuse a clean `pass` to enter the existing review gate after fresh completion confirmation.

## Impact

- OpenSpec planner, orchestrator, implementer, and verifier agent contracts and permissions.
- OpenSpec final-verification dispatch and local/explicit-store context propagation.
- Resolved `tasks.md` state handling, guarded atomic checkbox updates, and concurrent-change failure behavior.
- Installer/asset contract tests and focused workflow tests; final repository tests and builds remain verifier-only.
- No public API or dependency change, and no change to corrected-worker-failure classification.
