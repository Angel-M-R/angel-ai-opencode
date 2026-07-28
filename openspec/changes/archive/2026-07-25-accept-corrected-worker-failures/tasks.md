## 1. Shared Policy Contract Tests

- [x] 1.1 Update the orchestrator asset contract tests to require one evidence-based corrected-failure rule for both tooling mistakes and real intermediate failures, including same-worker bounded correction, ordered failed/correction/success evidence, equivalent-or-broader `exit 0` validation, final `done`/green state, and no deviation or scope violation.
- [x] 1.2 Add or update negative contract cases for cross-worker reruns, missing or ambiguous correction evidence, narrower or unrelated validation, final red evidence, `partial`/`blocked`, deviations, scope expansion, and out-of-scope work.
- [x] 1.3 Add contract coverage that inventories every current strict-default consumer of the shared policy and requires each to continue through its existing clean-result route after an eligible corrected real failure, including normal final-verification-to-archive progression without an incident-only authorization stop.
- [x] 1.4 Retain focused assertions that planned-task-only bounded repair, checkbox, deferral, independence, cadence, retry, and hard-stop safeguards do not become available to strict-default consumers.

## 2. Shared Orchestrator Policy

- [x] 2.1 Revise the shared implementation-result policy to remove the tooling-only restriction and define the evidence-complete corrected-real-failure classification without weakening any mandatory-stop condition.
- [x] 2.2 Update applicable worker return contracts to preserve the failed command and exit code, diagnosed cause and bounded correction, successful equivalent-or-broader validation and exit code, final validation state, status, files touched, deviations, and scope evidence.
- [x] 2.3 Audit and reconcile every strict-default control point that invokes the shared policy so an eligible corrected incident is surfaced and follows that point's existing clean next step, while non-clean results still use the shared mandatory-stop interaction.
- [x] 2.4 Reconcile planned-task completion routing so clean corrected incidents use the common classification and every existing planned-task-specific repair, deferral, dependency, cadence, retry, and conflict rule remains unchanged for non-clean results.
- [x] 2.5 Ensure final verification proceeds to the existing archive path when all prerequisites are green and the only incident was an eligible corrected failure, without changing any other verification, review, or archive rule.

## 3. Validation

- [x] 3.1 Run the focused orchestrator asset contract tests and confirm all corrected-failure continuation, retained-evidence, mandatory-stop, planned-safeguard, and archive-routing cases pass.
- [x] 3.2 Run the repository's broader applicable Go test suite and confirm the orchestrator asset and installer contracts remain green.
- [x] 3.3 Review the final diff for changes outside the shared result policy, its consumer/result contracts, and applicable tests; remove any unrelated implementation, review, cadence, or archive changes.
