## 1. Canonical Planned-Task Flow

- [x] 1.1 Revise `assets/agents/angel-orchestrator.md` so fresh OpenSpec state drives the initial complete task tree, every section dispatch, every clean-result refresh, and the complete tree shown at mandatory stops.
- [x] 1.2 Remove the cadence question, task and section pause modes, retained cadence state, and cadence-boundary pauses; automatically chain one exact incomplete-section batch at a time until completion or a mandatory stop.
- [x] 1.3 Update planned-task prompts and result handling to require focused textual checks during implementation, preserve every existing stop condition for commands and worker results, and reserve mandatory repository tests and build for final OpenSpec verification.

## 2. Deployment and Focused Validation

- [x] 2.1 Run focused positive and negative textual checks against the canonical asset for automatic execution, section bounds, state refreshes, tree-at-stop behavior, final-only mandatory verification, and removal of obsolete cadence instructions.
- [x] 2.2 Replace `$HOME/.config/opencode/agents/angel-orchestrator.md` from the validated canonical asset without creating a backup.
- [x] 2.3 Confirm the repository and installed orchestrator files are byte-for-byte identical and repeat focused textual checks on the deployed content; leave mandatory tests and build to final OpenSpec verification.

## 3. Route-Wide Mandatory-Stop Interaction

- [x] 3.1 Revise the shared mandatory-stop policy in `assets/agents/angel-orchestrator.md` so every stop reports retained blocking evidence first, then asks exactly one blocker-specific `question` with a safe stop option and custom responses still available, and performs no retry, continuation, scope broadening, substitute selection, phase advance, or worker dispatch before the user selects an action.
- [x] 3.2 Route existing-target resolution, OpenSpec bootstrap, planned-task implementation, and final-verification stops through the shared report-first/question-second policy while preserving automatic chaining between clean planned-task sections.
- [x] 3.3 Route Direct Safe, Direct Fast, and Direct review-fix stops through the shared policy while preserving clean-path review behavior and existing route boundaries.

## 4. Focused and Contract Validation

- [x] 4.1 Correct the two stale cadence assertions in `internal/install/agent_assets_test.go` so they require bounded automatic implementation and automatic clean-section progression rather than retained cadence wording.
- [x] 4.2 Extend the existing orchestrator asset tests with explicit ordered coverage for blocking evidence before the contextual question, references from every stop route, a safe stop option with custom-response support, and negative coverage proving that no recovery or worker dispatch occurs before user selection.
- [x] 4.3 Run focused textual checks and focused orchestrator contract tests for the revised canonical asset, preserving every command and exit code and stopping on any uncorrected failure.

## 5. Synchronization and Final Pre-Verification

- [x] 5.1 Replace `$HOME/.config/opencode/agents/angel-orchestrator.md` from the validated canonical repository asset without creating a backup, confirm byte-for-byte identity, and repeat the focused stop-contract checks against the installed copy.
- [x] 5.2 Before handoff to final OpenSpec verification, repeat the focused checks, `go test ./...`, `go build ./...`, synchronization checks, and strict OpenSpec validation with command and exit-code evidence; leave the completed change ready for the required final `openspec-verifier` pass.
