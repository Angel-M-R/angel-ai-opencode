## 1. Planned-Batch Policy and Contracts

- [x] 1.1 Revise `assets/agents/angel-orchestrator.md` so the shared strict policy remains the default while only section-bounded planned OpenSpec task batches can classify eligible `partial`, local `blocked`, red focused-test, additional-read, and successful-focused-test outcomes as deferrable or benign.
- [x] 1.2 Define the conservative independence gate, accumulated deferred-evidence record, one combined pre-retry report, exactly one fresh-state retry round, and final unresolved-work stop in the planned-task loop.
- [x] 1.3 Update the implementer batch contract to permit tests focused on modified code, continue prohibiting full suites and builds, preserve unchecked incomplete tasks, and stop on out-of-batch writes, functional expansion, destructive commands, unresolvable state, or checked-task/red-validation conflicts.
- [x] 1.4 Update `internal/install/agent_assets_test.go` contract coverage for deferral eligibility, explicit independence, grouped reporting, the single retry round, focused-test ownership, task-state integrity, and unchanged strict Direct, review-fix, bootstrap, target-resolution, and final-verification behavior.

## 2. Focused Validation

- [ ] 2.1 Run the focused orchestrator asset contract tests for `internal/install/agent_assets_test.go`, retain command and exit-code evidence, and leave any failing task unchecked.
- [x] 2.2 Inspect the bounded diff to confirm only the canonical orchestrator prompt and its contract tests changed, no full repository suite or build ran during implementation, and no unrelated OpenSpec change was touched.
