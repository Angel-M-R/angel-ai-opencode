## Why

Authorized validation commands can create regenerable workspace artifacts such as `.next/`, `.codegraph/`, and `*.tsbuildinfo`. Treating those outputs as manual scope deviations merely because Git tracks or exposes them creates false blockers across otherwise safe Direct and OpenSpec flows.

## What Changes

- Introduce an evidence-based classification for regenerable artifacts produced by an authorized successful validation command, regardless of Git ignore status.
- Require causal traceability from the command to an attributable diff and require the diff to contain no manual source-code edits before the generated-output classification applies.
- Apply the classification consistently to Direct Safe, planned OpenSpec implementation, bounded fix batches, and final verification.
- Report retained generated outputs separately with their command and attribution evidence; do not classify them as deviations, request authorization for them, or clean them automatically.
- Apply one workspace-audit filter across every route: silently omit a path only when the first component of its workspace-relative path begins with `.`, the path was already modified in a reliable worker-start baseline, and its state is identical at worker completion.
- Keep dotpaths under normal classification when the baseline is unavailable or unreliable or when their state changes; the filter does not authorize creating or modifying them, and qualifying unchanged pre-existing dotpaths do not appear in results.
- Preserve mandatory stops for failed commands, red final validation, destructive actions, ambiguous provenance, manual out-of-scope edits, scope expansion, and all existing route-local safeguards.
- Keep positive generated-output contract cases for `.next/`, `.codegraph/`, and `*.tsbuildinfo` and the existing safety negatives, and add only one positive dotpath contract case for a pre-existing unchanged `.vscode/settings.json`.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `execution-route-selection`: Extend shared result classification and every workspace-auditing route to recognize causally proven generated validation artifacts and silently omit only reliably baselined, unchanged pre-existing top-level dotpaths without weakening strict failure or scope handling.
- `implementation-cadence-selector`: Ensure generated validation artifacts are not treated as planned-batch deviations or out-of-batch writes, and inherited unchanged qualifying dotpaths are not mistaken for worker writes, while preserving bounded task scope, verifier ownership, and hard blockers.

## Impact

- Canonical orchestration and worker policy assets, including Direct, OpenSpec implementation, fix, and final-verification result contracts.
- Focused agent-asset contract tests in `internal/install/agent_assets_test.go`, including cross-route workspace-audit propagation, and applicable installer/asset propagation coverage.
- OpenSpec delta requirements for shared route classification and planned implementation cadence.
- No functional expansion of implementation batches, no VS Code configuration objective, no new destructive permissions, and no automatic cleanup of generated artifacts.
