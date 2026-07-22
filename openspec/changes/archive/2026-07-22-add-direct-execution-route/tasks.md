## 1. Orchestrator Contract Tests

- [x] 1.1 Extend `internal/install/agent_assets_test.go` with section helpers and ordered assertions proving that the new-work route gate occurs after Brief confirmation and before OpenSpec bootstrap, CLI use, worker dispatch, or change creation, while existing OpenSpec changes bypass Direct and continue by status.
- [x] 1.2 Add contract cases for the non-binding Direct/OpenSpec recommendation criteria, the Safe/Fast choice, exclusive `general` implementation routing, the bounded prompt fields and result contract, and direct-mode exclusion of bootstrap, OpenSpec CLI/artifacts, the orchestrator, and OpenSpec workers.
- [x] 1.3 Add baseline contract cases for Safe same-worker verification and stop behavior, Fast's no-test/no-review unverified result, selected-review routing, selected-finding-only `general` fixes, affected-reviewer-only reruns, and preservation of the existing OpenSpec workflow and review-fix route.
- [x] 1.4 Run the focused orchestrator asset contract tests and confirm the new assertions fail for the expected missing direct-route instructions before updating the asset.
- [x] 1.5 Add contract cases for planned OpenSpec implementation batches and Direct Safe workers proving that a reported intermediate non-zero command permits progression only after the same worker reports a successful equivalent-or-broader relevant rerun, final `done`, and no deviation or out-of-scope work, while retaining the complete corrected-failure evidence.
- [x] 1.6 Add contract cases proving mandatory stops for a missing equivalent-or-broader rerun, final red state, `partial`/`blocked`, deviation, out-of-scope work, an unrelated or narrower green command, and a TDD/expected failure still red at batch end.

## 2. Execution Route Gate

- [x] 2.1 Update `assets/agents/angel-orchestrator.md` so a confirmed Brief for new non-trivial work receives a risk-based, non-binding recommendation and an OpenSpec-versus-Direct question before any new-change OpenSpec side effect.
- [x] 2.2 Encode that requests targeting an existing OpenSpec change skip Direct selection, resolve fresh OpenSpec status, and continue through the current OpenSpec workflow.
- [x] 2.3 Connect the OpenSpec selection to the existing bootstrap, planner, artifact, bounded implementation, verification, review, and archive flow.
- [x] 2.4 Update the orchestrator asset's planned-task result contract and stop/completion routing to retain and surface corrected-failure evidence, permit progression under the selected cadence only after a clean equivalent-or-broader rerun, and preserve every mandatory-stop condition.

## 3. Direct Execution Modes

- [x] 3.1 Add the Safe/Fast selector and bounded direct task template carrying the confirmed Brief verbatim, selected mode, scope, and required status, touched-files, command/exit-code, and deviation fields to `general` only.
- [x] 3.2 Define Safe mode so the same `general` worker implements and runs existing applicable tests/builds, with immediate report-and-stop behavior without retries for unavailable verification, `partial`/`blocked`, Brief deviation, or out-of-scope work.
- [x] 3.3 Define Fast mode so `general` only implements, no tests or reviews run, and the orchestrator explicitly reports the result as implemented but unverified.
- [x] 3.4 Update the orchestrator asset's Direct Safe worker contract and clean-result classification to accept only fully evidenced corrected intermediate failures, surface both failed and successful rerun commands, and stop on every unresolved or finally red result.

## 4. Direct Review Routing

- [x] 4.1 After a clean Safe result, reuse the Security risk, Simplicity, Correctness, or no-review choices and scope selected reviewers to the bounded direct diff and confirmed Brief.
- [x] 4.2 Route only user-selected findings as one bounded fix batch to `general`, apply the structured result, shared corrected-failure policy, and mandatory-stop contract, and allow only reviewers responsible for addressed findings to be rerun after a clean fix.
- [x] 4.3 Ensure the direct review path never invokes `openspec-implementer`, OpenSpec verification, or OpenSpec archive behavior.

## 5. Verification

- [x] 5.1 Run the focused orchestrator asset contract tests and confirm route ordering, shared corrected-failure and mandatory-stop semantics, selected-cadence continuation, direct review routing, existing-change behavior, and OpenSpec preservation all pass.
- [x] 5.2 Run the repository's existing full test suite and build checks, recording every command and exit code.
- [x] 5.3 Review the final diff to confirm only the orchestrator asset, its contract tests, and this change's task checkboxes were modified, preserving all unrelated uncommitted work.
