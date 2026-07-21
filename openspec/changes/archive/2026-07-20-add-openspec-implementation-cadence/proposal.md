## Why

OpenSpec implementation currently advances through bounded worker batches without first giving the user a clear view of the real task state or a single, explicit cadence choice. The orchestrator should make implementation progress predictable while preserving existing stop conditions, verification, and review safeguards.

## What Changes

- Before each planned-task implementation run, read the change's current `tasks.md` through the real OpenSpec-resolved change state rather than relying on conversational history.
- Present a complete, compact task tree with section progress, task identifiers and summaries, and `✓`/`☐` completion markers.
- Ask once whether to pause after each task, after each section, or run all remaining tasks.
- For run-all cadence, automatically chain bounded implementation batches while stopping on `blocked`, `partial`, failed tests/build, or plan deviation.
- Automatically invoke verification after every planned task is complete, then retain the existing post-verification review gate.
- Keep post-verification review-fix batches on their existing finding-ID routing, outside the planned-task cadence.
- Keep the repository orchestrator asset and the live global orchestrator agent synchronized during implementation, replacing the global file without creating a backup.
- Validate the instruction-only change manually; do not add automated tests.

## Capabilities

### New Capabilities

- `implementation-cadence-selector`: Defines task-state presentation, one-time cadence selection, bounded batch chaining, mandatory stop conditions, automatic verification, and preservation of the review gate.

### Modified Capabilities

None.

## Impact

- Repository orchestrator instruction asset: `assets/agents/angel-orchestrator.md`.
- Live global orchestrator instruction file: `$HOME/.config/opencode/agents/angel-orchestrator.md`, updated only during implementation and without a backup.
- OpenSpec implementation routing behavior and user interaction; no product runtime APIs, dependencies, or application code are affected.
- Validation is manual and adds no tests.
