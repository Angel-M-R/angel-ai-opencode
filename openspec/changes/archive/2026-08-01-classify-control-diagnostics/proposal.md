## Why

Read-only internal bookkeeping can currently be reported as a deviation and trigger a mandatory stop even when no command failed, no mutation occurred, no file was touched, and final validation is green. The shared result policy needs an explicit visible classification for these benign control-point diagnostics without weakening any existing safety stop.

## What Changes

- Add a structured, ordered `control diagnostics` result category to every control point, reported as `none` when empty and kept mutually exclusive with `deviations`.
- Allow eligible read-only bookkeeping diagnostics to remain visible without blocking clean routing; label incomplete diagnostic records as `incomplete` and continue only when the final relevant state is green.
- Preserve mandatory stops for real deviations, functional or scope expansion, out-of-scope work, destructive actions, failed commands, mutations, touched files, `partial` or `blocked` status, and red final validation.
- Preserve all workflow prerequisites and the existing corrected-failure and recovered-probe rules.
- Update only the repository source asset `assets/agents/angel-orchestrator.md`; do not synchronize or mutate the local installed prompt in this change.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `execution-route-selection`: Extend the shared result contract and clean-result classification with a non-blocking `control diagnostics` category at every control point while preserving existing mandatory-stop behavior.

## Impact

- Shared orchestrator result contracts, reporting, and control-point classification policy.
- Repository policy source at `assets/agents/angel-orchestrator.md` only during implementation.
- Textual diff review and `git diff --check` validation; no prompt installation, synchronization, workflow-prerequisite change, corrected-failure change, or recovered-probe change.
