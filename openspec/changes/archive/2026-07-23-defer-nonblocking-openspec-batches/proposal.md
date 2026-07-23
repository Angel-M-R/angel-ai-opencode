## Why

The planned OpenSpec task loop currently stops at every incomplete, blocked, deviating, or red focused-check batch even when later sections are demonstrably independent. This prevents useful bounded work from progressing and interrupts the user before the orchestrator has exhausted safe, plan-defined work.

## What Changes

- Classify non-clean outcomes only inside planned OpenSpec task batches as either deferrable or immediately blocking.
- Allow `partial`, local `blocked`, and red focused-test outcomes to be deferred when incomplete tasks remain unchecked and later sections are explicitly evidenced as independent; uncertainty means dependency.
- Accumulate deferrable incidents and benign deviations without intermediate questions, continue through independent sections, then present one combined report before one final retry round.
- Permit implementers to run focused tests for modified code while reserving full suites and builds for final verification.
- Stop immediately for scope violations, functional expansion, destructive commands, unresolvable OpenSpec state, checked tasks with relevant red validation, and other non-deferrable conflicts.
- Preserve strict behavior for Direct Safe, Direct Fast, review-fix batches, bootstrap, change resolution, and final verification.
- Update the canonical orchestrator prompt and its focused contract tests together; do not ignore red tasks or mark incomplete tasks complete.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `implementation-cadence-selector`: Allow evidence-gated continuation past deferrable planned-task batch outcomes, aggregate incidents, and perform one final retry round while preserving task-state integrity and final verification.
- `execution-route-selection`: Narrow the shared mandatory-stop policy so only planned OpenSpec task batches receive deferral behavior while all Direct and non-batch OpenSpec routes remain strict.

## Impact

- Canonical orchestrator policy: `assets/agents/angel-orchestrator.md`.
- Contract coverage: `internal/install/agent_assets_test.go`.
- OpenSpec requirements for planned implementation cadence and route-specific result handling.
- No product behavior, dependency, release, bootstrap, review-fix, target-resolution, or final-verification changes.
