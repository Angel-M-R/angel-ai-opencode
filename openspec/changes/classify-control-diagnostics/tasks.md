## 1. Shared control-diagnostic policy

- [x] 1.1 Update only `assets/agents/angel-orchestrator.md` so the shared result contract used by every control point includes ordered, structured `control diagnostics` records with `complete` or `incomplete` state and explicit `none` when empty.
- [x] 1.2 Define the non-blocking classification as read-only internal bookkeeping with no failed command, mutation, destructive action, or touched file, final status `done`, and green final relevant validation; allow incomplete descriptive records to continue only when every classification fact and the green final state remain established.
- [x] 1.3 Make `control diagnostics` and `deviations` mutually exclusive while preserving mandatory stops for real deviations, functional or scope expansion, out-of-scope work, destructive actions, failed commands, mutations, touched files, `partial` or `blocked` status, and red validation.
- [x] 1.4 Preserve every workflow prerequisite and the existing corrected-failure, recovered-probe, generated-validation-artifact, planned-task, and route-local rules, and do not modify or synchronize the local installed prompt.

## 2. Final verification
<!-- owner: openspec-verifier -->

- [x] 2.1 Review the final implementation diff and report evidence that `control diagnostics` is visible and separate from `deviations`, applies at every shared-policy control point, leaves real mandatory stops intact, and changes no implementation target other than `assets/agents/angel-orchestrator.md`.
- [x] 2.2 Run `git diff --check` and report the command and exit code, confirming that it creates no generated validation artifacts.
