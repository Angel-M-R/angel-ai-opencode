## Context

The orchestration policy currently treats files touched outside an assigned implementation batch, and tracked files touched by final verification, as potential deviations or scope violations. Validation tools routinely create or refresh regenerable state such as Next.js output, CodeGraph indexes, and TypeScript build metadata. Git ignore status is not a reliable safety boundary: a valid generated artifact may be tracked or newly visible, while an ignored path can still conceal an unsafe manual edit. Separately, a worker can inherit an already-dirty top-level dotpath such as `.vscode/settings.json`; repeatedly reporting it despite proving that the worker left it byte-for-byte unchanged adds noise without describing worker activity.

The canonical shared classification and result fields live in `assets/agents/angel-orchestrator.md`. Direct Safe work and bounded fixes receive that contract through orchestrator prompts; planned OpenSpec work and final verification additionally depend on `assets/agents/openspec-implementer.md` and `assets/agents/openspec-verifier.md`. Focused prompt-contract tests in `internal/install/agent_assets_test.go` protect these route boundaries.

## Goals / Non-Goals

**Goals:**

- Recognize any regenerable artifact produced by an authorized successful validation command, independent of Git ignore status.
- Require auditable causal evidence tying each generated diff to its command and excluding manual source edits.
- Use one separate generated-artifact report category across Direct Safe, OpenSpec planned batches, bounded fixes, and final verification.
- Silently exclude only reliably baselined, pre-existing, unchanged paths whose first workspace-relative component begins with `.`, consistently across every workspace-auditing route.
- Retain generated outputs while preserving every existing failure, destructive-action, ambiguity, scope, and final-red safeguard.
- Preserve positive `.next/`, `.codegraph/`, and `*.tsbuildinfo` cases and representative safety negatives, and add exactly one positive `.vscode/settings.json` contract case for the dotpath filter.

**Non-Goals:**

- Expand the functional scope, writable source paths, or task boundaries of any batch.
- Treat failed validation, red final evidence, destructive actions, ambiguous provenance, or manual source edits as safe.
- Add a generated-path allowlist, make Git ignore rules authoritative, or infer provenance from a filename alone.
- Authorize creating or modifying dotpaths, suppress paths below a non-dot first component, or suppress any path without a reliable unchanged baseline.
- Introduce a standalone VS Code configuration objective or modify `.vscode/settings.json`.
- Clean, revert, delete, stage, or commit generated artifacts automatically.
- Change route selection, review selection, archive prerequisites, or corrected-failure and recovered-probe eligibility.

## Decisions

### Define eligibility by causality and regenerability, not path or Git state

An output is eligible only when an authorized validation command exits zero and the worker can attribute the resulting created or modified paths to that command. The causal record will identify the command and exit code, the immediately relevant before/after workspace evidence or equivalent attributable diff, the generated paths, and why those paths are reproducible outputs. No intervening manual mutation may be folded into the evidence, and the attributable diff must contain no manual source-code edit.

This applies to `.next/`, `.codegraph/`, `*.tsbuildinfo`, and other outputs that satisfy the same evidence rule. A fixed allowlist was rejected because it would make the policy incomplete and would incorrectly treat a familiar name as proof. Git ignore status was rejected because it expresses repository tracking preference, not causal provenance or safety.

### Add a separate generated-validation-artifacts result category

The authoritative shared result field set will gain one category containing generated paths, the producing command and zero exit code, causal-diff evidence, and confirmation that outputs were retained. Eligible entries are neither deviations nor out-of-scope writes and do not trigger an authorization question. The ordinary touched-files and command history remain complete; the separate category explains classification rather than hiding workspace changes.

This category will be referenced by every route contract instead of duplicated route-local definitions. The alternative of recording outputs as benign deviations was rejected because the Brief explicitly distinguishes generated output from deviation and because final verification does not use the planned-batch benign-deviation exception.

### Filter only baseline-stable top-level dotpaths from workspace audits

Every route that audits the workspace will use the same worker-boundary filter before reporting or classifying paths. For each candidate path relative to the workspace root, the worker will inspect the first path component. It will silently omit the path only when that component begins with `.`, reliable start-of-worker evidence shows that the path was already modified, and end-of-worker evidence proves that the path remains identical to that baseline. A nested dot component below a non-dot first component does not qualify. Qualifying paths are absent from the result rather than being relabeled as generated output or a benign deviation.

The filter grants no write authority. If the worker lacks a reliable baseline, creates the path, modifies it directly, or observes any state change during its invocation, normal generated-output, deviation, scope, corrected-failure, and mandatory-stop rules apply. This keeps `.next/`, `.codegraph/`, and other changed validation outputs visible for causal classification while allowing an inherited unchanged `.vscode/settings.json` to disappear from route results. Applying the rule at every workspace-audit boundary was chosen over a route-local exception because partial propagation would produce inconsistent safety decisions. Ignoring every dot-prefixed component was rejected because it could conceal nested or newly changed state.

### Keep safety gates conjunctive and independent

Generated-output eligibility does not make a command successful, repair a failure, turn red validation green, or authorize a destructive action. A failed command remains governed by the existing corrected-failure or recovered-probe rules. Red final evidence, `partial` or `blocked` status, ambiguous attribution, manual source edits, functional expansion, and out-of-scope work retain their existing mandatory-stop behavior.

An eligible generated output also does not broaden a planned task: it is accepted only as incidental output of validation already authorized for that bounded work. Any source change or unrelated generated activity remains subject to normal batch boundaries. This separation avoids converting a tracing exception into functional authorization.

### Permit verifier-created generated outputs without granting edit authority

The verifier remains unable to use edit/write tools or perform manual tracked-file mutation. Its executed validation commands may leave their normal causally proven generated outputs even when Git tracks them. The verifier will report and retain those outputs through the shared category, while any manual source change, ambiguous diff, or destructive cleanup remains blocking.

Granting general write capability was rejected because it would erase the distinction between command effects and manual edits. Automatic cleanup was rejected because it is destructive, loses evidence, and contradicts the retention requirement.

### Protect the policy with focused route-wide contracts

Contract tests will first assert the authoritative category and eligibility rule once, then assert propagation to Direct Safe, OpenSpec planned implementation, Direct/OpenSpec fix prompts, and final verification. Existing positive fixtures will continue to cover tracked or unignored `.next/`, `.codegraph/`, and `*.tsbuildinfo` outputs. Existing negative assertions will continue to cover non-zero commands, red validation, destructive actions, ambiguous or missing causality, intervening/manual source edits, scope expansion, and attempted cleanup. One additional positive fixture will prove that a reliably baselined `.vscode/settings.json` already modified before the worker and identical afterward is silently omitted across workspace-auditing routes; no broader VS Code behavior matrix is added.

Tests will continue to assert corrected-failure evidence, planned-batch limits, verifier read-only behavior, and strict stop routing. This focused approach was chosen over integration tests that manufacture each tool's real cache because the behavior being changed is the instruction contract and classification policy, not those tools.

## Risks / Trade-offs

- [Workers overclaim causality from a familiar generated path] → Require command-specific before/after or equivalent diff evidence and reject filename-only or ambiguous attribution.
- [A validation command also rewrites source] → Require the attributable diff to contain no manual source edit; classify the result through existing deviation and scope rules instead.
- [Tracked generated files make the verifier appear to edit] → Distinguish command side effects from manual edit capability and retain both command and diff evidence.
- [Route copies drift] → Keep one authoritative shared field/category definition and test that every route references it.
- [Generated output accumulates] → Accept retention as an explicit trade-off; cleanup requires separate authorization and is not part of this change.
- [The exception is mistaken for wider batch scope] → State and test that only incidental validation output is reclassified and all functional boundaries remain unchanged.
- [An inherited dotpath is mistaken for unchanged worker state] → Require a reliable worker-start baseline and exact end-state identity; otherwise apply normal rules.
- [The dotpath filter drifts between routes] → Define it once at the workspace-audit boundary and contract-test propagation to every auditing route.

## Migration Plan

1. Preserve the implemented generated-artifact contracts, route propagation, and safety negatives; add the single cross-route pre-existing unchanged `.vscode/settings.json` contract.
2. Update the canonical orchestrator and OpenSpec worker assets with the shared top-level dotpath filter without changing generated-artifact safety boundaries.
3. Run focused asset contracts and applicable propagation tests, then the repository's mandatory final verification and OpenSpec validation.
4. Deliver updated embedded assets through the normal build/install path; rollback uses the prior agent assets and does not delete generated workspace artifacts or alter inherited dotpaths.

## Open Questions

None.
