## 1. Repository Orchestrator Cadence

- [x] 1.1 Update `assets/agents/angel-orchestrator.md` so every planned-task implementation decision applies the centralized fresh-state invariant and renders the complete compact progress tree with section counts, task identifiers, summaries, and `✓`/`☐` markers.
- [x] 1.2 Add the one-time task/section/run-all cadence question and session-only selection behavior, including task-bounded and section-bounded implementer prompts and automatic section-batch chaining for run-all.
- [x] 1.3 Add mandatory stops for `blocked`, `partial`, failed test/build commands, plan deviations, and refreshed-state conflicts; route a fully completed task tree directly to automatic verification while preserving the existing post-verification review gate.

## 2. Manual Validation

- [x] 2.1 Manually inspect and walk through mixed task-state rendering plus task, section, and multi-section run-all cadence flows without adding automated tests.
- [x] 2.2 Manually walk through every mandatory stop condition and the completed-tasks → verification → existing review-gate transition, confirming the instructions do not permit improvised continuation.

## 3. Global Agent Synchronization

- [x] 3.1 After manual validation, overwrite `$HOME/.config/opencode/agents/angel-orchestrator.md` from the repository asset without creating a backup or sidecar file.
- [x] 3.2 Confirm the repository asset and live global agent are byte-for-byte identical and report the manual validation and synchronization evidence.
