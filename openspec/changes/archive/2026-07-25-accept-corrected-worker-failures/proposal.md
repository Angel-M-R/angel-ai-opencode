## Why

The shared orchestrator policy currently recognizes only syntax or invocation mistakes as corrected intermediate failures, so a worker that fixes a real bounded defect and proves the fix can still trigger an unnecessary stop. The policy should judge the final, fully evidenced result without concealing the failed command or weakening unresolved-failure safeguards.

## What Changes

- Allow a real intermediate implementation or verification failure to be classified as corrected when the same worker fixes its in-scope cause, runs equivalent-or-broader validation successfully, returns `done`, and reports no deviation or out-of-scope work.
- Preserve and report the failed command, correction, and successful validation as ordered incident evidence.
- Continue the applicable workflow, including archive progression, without an authorization prompt solely because of an eligible corrected failure.
- Preserve mandatory stops for insufficient evidence, unresolved or finally red validation, `partial`/`blocked` results, deviations, scope expansion, and out-of-scope work.
- Apply the shared classification consistently at every strict-default control point while retaining the planned-task loop's separate safeguards and deferral rules.
- Update orchestrator contracts and tests for corrected real failures and retained stop conditions without changing unrelated implementation, review, or archive rules.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `execution-route-selection`: Broaden the shared corrected-intermediate-failure policy from tooling-only mistakes to fully repaired, same-worker, in-scope failures across all strict-default control points while preserving planned-task-specific safeguards.

## Impact

- Shared orchestration policy and each OpenSpec or Direct control point that consumes its implementation-result classification.
- Worker result/evidence contracts for failed commands, bounded corrections, and successful equivalent-or-broader validation.
- Orchestrator asset contract tests covering strict-default continuation, archive progression, and mandatory stops.
- No dependency, public API, or unrelated workflow-rule changes.
