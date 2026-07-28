## 1. Shared result policy

- [x] 1.1 Update the authoritative orchestrator asset to classify an inspection-only, non-destructive probe that fails solely on absent pre-operation state as a recovered probe rather than as a corrected implementation or validation failure.
- [x] 1.2 Require same-worker successful completion, authoritative green final validation, final `done`, complete ordered incident evidence, and no deviation, scope expansion, or out-of-scope work before automatic continuation without a follow-up question.
- [x] 1.3 Preserve and make explicit mandatory stops for unresolved failures, insufficient or unrelated evidence, red final validation, `partial`/`blocked`, destructive actions, and any scope or Brief deviation without changing unrelated routing or planned-task behavior.

## 2. Regression and propagation coverage

- [x] 2.1 Add a focused orchestrator contract test for `ls openspec` failing before the directory exists, followed by same-worker creation, authoritative successful status validation, visible incident evidence, and no continuation question solely for that probe.
- [x] 2.2 Add or retain negative contract assertions for destructive actions, unresolved or red outcomes, incomplete statuses, insufficient evidence, different-worker evidence, and deviations or scope growth.
- [x] 2.3 Extend temporary-directory agent installation coverage to select `angel-orchestrator`, verify it is cataloged and copied unchanged, and assert that the installed fixture contains the recovered-probe policy.
- [x] 2.4 Retain embedded-versus-directory asset parity coverage and document in the tested contract that repository edits reach an existing configured copy only through a rebuilt or explicit asset source followed by selected-agent reconciliation.

## 3. Validation

- [x] 3.1 As one bounded planned-task batch, run only the focused orchestrator contract and temporary installer tests covering the recovered-probe and propagation boundaries, and record each command with its exit code; do not run any build or the full repository test suite.
- [x] 3.2 Run authoritative OpenSpec validation for this change and confirm all scenarios remain coherent with the main `execution-route-selection` capability.
- [x] 3.3 Inspect the final diff and confirm it contains no unrelated refactor, behavior change, live installed-agent mutation, or work outside the change scope.
- [x] 3.4 Hand the focused-test, OpenSpec-validation, and scope evidence to the final `openspec-verifier`; only that verifier, after every planned task is checked, runs any mandatory full repository suites or builds.
