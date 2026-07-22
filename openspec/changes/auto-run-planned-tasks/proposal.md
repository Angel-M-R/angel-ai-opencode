## Why

The planned-task cadence selector offers pause modes that are no longer useful and can leave implementation waiting at artificial boundaries. Planned OpenSpec work should proceed automatically while retaining bounded section batches, fresh state, and all safety stops.

## What Changes

- **BREAKING**: Remove the planned-task cadence question and the “after each task” and “after each section” modes and instructions.
- Automatically execute all remaining planned tasks, one incomplete section per bounded implementer batch.
- Refresh OpenSpec status and the resolved task file between sections, preserving stale-state handling and every existing mandatory stop condition.
- Preserve the complete task-tree display before implementation and add it to every mandatory stop, without cadence-boundary pauses.
- Limit implementation batches to focused textual checks; reserve mandatory tests and builds for final OpenSpec verification.
- Update the canonical repository orchestrator asset and replace the installed global orchestrator from it so both copies are identical.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `implementation-cadence-selector`: Replace selectable cadence behavior with automatic section-bounded execution and revise task-tree, validation, and synchronization requirements accordingly.

## Impact

- Affected instructions: `assets/agents/angel-orchestrator.md` and `$HOME/.config/opencode/agents/angel-orchestrator.md`.
- Affected behavior: planned OpenSpec task batching, refresh points, stop presentation, and verification timing.
- No product runtime APIs or dependencies change; no automated tests are added.
