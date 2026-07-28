## Why

A harmless non-destructive probe can fail because the state it is checking does not exist yet, even though the same worker then completes the requested operation and proves the final result. Treating that recovered probe as a mandatory stop creates an unnecessary continuation question, while source and installed orchestrator assets can disagree about the applicable result policy.

## What Changes

- Distinguish a recovered non-destructive probe incident from unresolved implementation or validation failures.
- Continue automatically after such an incident only when the same worker finishes `done`, authoritative final validation is green, and there is no deviation, scope expansion, or out-of-scope work.
- Preserve and surface the failed probe, exit code, cause, and satisfactory final evidence instead of hiding the incident.
- Preserve mandatory stops for unresolved failures, red final validation, `partial` or `blocked` results, destructive actions, insufficient evidence, and any scope or Brief deviation.
- Align the source policy, generated or installed copies, and their update path so deployed behavior cannot silently retain an older rule.
- Add focused regression coverage for a failed `ls openspec` probe before directory creation followed by successful creation and authoritative validation, while retaining existing negative safeguards.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `execution-route-selection`: Refine the shared worker-result policy to permit evidence-complete recovery from a non-destructive probe failure and require consistency across source and installed orchestrator assets.

## Impact

- Orchestrator policy source under `assets/agents/` and its contract tests.
- Agent asset embedding, installation, and update behavior used to propagate the policy to configured OpenCode agents.
- Existing worker result reporting and mandatory-stop behavior; no unrelated route, scope, or implementation behavior changes.
