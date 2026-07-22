## Why

The planned-task cadence selector offers pause modes that are no longer useful and can leave implementation waiting at artificial boundaries. Planned OpenSpec work should proceed automatically while retaining bounded section batches, fresh state, and all safety stops. Final verification also exposed two stale cadence assertions and a broader safety gap: mandatory stops surface evidence but do not consistently ask the user what to do next before any recovery action.

## What Changes

- **BREAKING**: Remove the planned-task cadence question and the “after each task” and “after each section” modes and instructions.
- Automatically execute all remaining planned tasks, one incomplete section per bounded implementer batch.
- Refresh OpenSpec status and the resolved task file between sections, preserving stale-state handling and every existing mandatory stop condition.
- Preserve the complete task-tree display before implementation and add it to every mandatory stop, without cadence-boundary pauses.
- Limit implementation batches to focused textual checks; reserve mandatory tests and builds for final OpenSpec verification.
- Across every orchestrator route, require each mandatory stop to report its blocking evidence first and then ask one contextual next-action question that includes a safe stop option and permits a custom response.
- Prohibit retrying, continuing, broadening scope, or dispatching another worker after a mandatory stop until the user selects an action.
- Correct the two stale cadence assertions and add focused contract coverage for report-first/question-second ordering and the absence of automatic recovery.
- Update the canonical repository orchestrator asset and replace the installed global orchestrator from it so both copies are identical.
- Repeat focused checks, mandatory tests and build, synchronization checks, and OpenSpec verification after implementation.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `implementation-cadence-selector`: Replace selectable cadence behavior with automatic section-bounded execution and revise task-tree, route-wide mandatory-stop, validation, and synchronization requirements accordingly.

## Impact

- Affected instructions: `assets/agents/angel-orchestrator.md` and `$HOME/.config/opencode/agents/angel-orchestrator.md`.
- Affected tests: two stale assertions and focused route-wide stop-contract coverage in `internal/install/agent_assets_test.go`.
- Affected behavior: planned OpenSpec task batching, refresh points, route-wide stop presentation and next-action gating, and verification timing.
- No product runtime APIs or dependencies change.
