# implementation-cadence-selector Specification

## Purpose
TBD - created by archiving change add-openspec-implementation-cadence. Update Purpose after archive.

## Requirements

### Requirement: Automatic planned-task execution
When pending planned tasks exist, the orchestrator SHALL automatically execute all remaining tasks without asking a cadence or between-section continuation question. It MUST continue through clean section-bounded batches until no tasks remain or a mandatory stop occurs.

#### Scenario: Pending work spans multiple sections
- **WHEN** fresh task state shows pending tasks in more than one section
- **THEN** the orchestrator automatically dispatches each next incomplete section in sequence without asking for cadence or continuation between clean results

#### Scenario: All planned tasks are already complete
- **WHEN** the initial fresh task state contains no pending tasks
- **THEN** the orchestrator skips implementation dispatch and proceeds automatically to final verification

### Requirement: Focused contract validation and repeated final verification
The change MUST correct the two stale cadence assertions in `internal/install/agent_assets_test.go` and add focused contract coverage for route-wide report-first/question-second stop behavior and the absence of automatic recovery. Implementation MAY run focused textual and contract checks, but final verification MUST repeat the focused checks, mandatory repository tests and build, orchestrator synchronization check, and OpenSpec verification.

#### Scenario: Stop contract tests are updated
- **WHEN** the orchestrator stop protocol is revised
- **THEN** the existing asset test corrects both stale cadence assertions and verifies evidence-before-question ordering, a safe stop option, route coverage, and no recovery action before user selection

#### Scenario: Revised implementation reaches final verification
- **WHEN** all newly added tasks are complete
- **THEN** final verification reruns focused checks, the mandatory repository tests and build, the repository-to-global synchronization check, and OpenSpec verification with command and exit-code evidence

### Requirement: Current OpenSpec task state drives planned-task implementation
At every planned-task implementation decision point, the orchestrator MUST query the real OpenSpec status for the active change, read the resolved current `tasks.md`, and recompute the task tree and next batch. It MUST apply this invariant before the initial task-tree display, before every planned-task worker dispatch, and after every clean worker result. It MUST NOT select a planned-task batch from conversational history or a cached task list.

#### Scenario: Task state changed since the previous batch
- **WHEN** a prior batch or external actor changes task completion state
- **THEN** the orchestrator selects the next batch from a fresh OpenSpec status query and `tasks.md` read

#### Scenario: Current tasks cannot be resolved
- **WHEN** OpenSpec status does not report a complete tasks artifact or the resolved `tasks.md` cannot be read
- **THEN** the orchestrator reports the blocking evidence and unavailable task tree, then asks one contextual next-action question before taking any recovery action

### Requirement: Complete compact task tree
Before planned-task implementation begins, the orchestrator SHALL display the complete task hierarchy in a compact tree. At every mandatory implementation stop, it MUST refresh and display the complete tree when current task state can be resolved. The tree MUST show completed and total counts at the root and for each section, and every task MUST show its identifier, short text, and either `✓` when complete or `☐` when pending. Clean transitions between automatic section batches MUST NOT introduce a task-tree pause.

#### Scenario: Mixed task completion before implementation
- **WHEN** `tasks.md` contains completed and pending tasks across multiple sections before the first dispatch
- **THEN** the rendered tree includes every section and task with accurate section progress and completion markers

#### Scenario: Mandatory stop with resolvable task state
- **WHEN** a mandatory stop occurs and fresh OpenSpec task state can be resolved
- **THEN** the orchestrator displays the complete refreshed compact tree, reports the stop evidence, and then asks one contextual next-action question before taking any recovery action

#### Scenario: Mandatory stop with unresolvable task state
- **WHEN** a mandatory stop occurs because current OpenSpec task state cannot be resolved
- **THEN** the orchestrator reports the blocking evidence and that the complete tree is unavailable without using cached task state, then asks one contextual next-action question

#### Scenario: Clean section completes
- **WHEN** an implementation section completes cleanly and pending tasks remain
- **THEN** the orchestrator refreshes state and dispatches the next bounded section without a task-tree or cadence-boundary pause

### Requirement: Bounded implementation batches
Every planned-task implementation worker prompt MUST identify an exact bounded task batch containing only the pending tasks from one named incomplete section. The orchestrator MUST chain these section-bounded batches automatically rather than issue an unbounded completion request, and it MUST refresh task state before selecting each subsequent batch.

#### Scenario: Automatic execution spans multiple sections
- **WHEN** pending tasks exist in more than one section
- **THEN** the orchestrator dispatches one explicitly identified incomplete section at a time and refreshes task state before selecting each subsequent section

#### Scenario: Intended section is already complete
- **WHEN** the refreshed `tasks.md` shows that the intended section no longer has pending work
- **THEN** the orchestrator skips the stale batch and recomputes the next bounded section batch

### Requirement: Mandatory implementation stop conditions
At every mandatory stop across OpenSpec target resolution, bootstrap, planned-task implementation, final verification, Direct Safe or Fast execution, and Direct review-fix routing, the orchestrator MUST preserve and report the blocking evidence before asking exactly one contextual next-action question with the `question` tool. The question MUST derive its choices from the blocker, include a safe stop option, and leave the tool's custom response available. Until the user selects an action, the orchestrator MUST NOT retry, continue, broaden scope, select substitute work, dispatch another worker, or advance to the route's next phase. Planned-task stops MUST also display the refreshed complete task tree when state remains resolvable.

#### Scenario: Worker reports a non-clean status
- **WHEN** an implementation worker reports `blocked` or `partial`
- **THEN** the orchestrator stops, presents the refreshed complete tree when available, reports the worker status and evidence, and then asks one contextual next-action question

#### Scenario: A command remains failed
- **WHEN** an invoked command reports a non-zero exit code without a later equivalent-or-broader relevant successful rerun
- **THEN** the orchestrator reports the failed command and exit code, then asks one contextual next-action question before any rerun or worker dispatch

#### Scenario: Plan deviation is detected
- **WHEN** the worker reports a deviation or out-of-scope work, or refreshed task state conflicts with the requested batch
- **THEN** the orchestrator stops, presents the refreshed complete tree when available, reports the deviation, and then asks one contextual next-action question without selecting substitute work

#### Scenario: Bootstrap or target resolution blocks
- **WHEN** an existing OpenSpec target cannot be resolved or an OpenSpec readiness bootstrap blocks or fails
- **THEN** the orchestrator reports the target or bootstrap diagnostic, then asks one contextual next-action question before retrying, falling back, or dispatching an OpenSpec worker

#### Scenario: Direct route stops
- **WHEN** a Direct Safe or Fast implementation or a Direct review-fix result meets a mandatory-stop condition
- **THEN** the orchestrator reports the retained result and verification evidence, then asks one contextual next-action question before retrying, broadening scope, opening reviews, or dispatching another worker

#### Scenario: Stop question is presented
- **WHEN** blocking evidence has been reported for any mandatory stop
- **THEN** the orchestrator asks exactly one blocker-specific question whose choices include a safe stop option and whose question tool permits a custom response

#### Scenario: User has not selected an action
- **WHEN** the mandatory-stop question is unanswered
- **THEN** the orchestrator performs no retry, continuation, scope expansion, substitute selection, phase advance, or worker dispatch

### Requirement: Automatic verification and retained review gate
After a fresh task-state read shows every task complete, the orchestrator MUST dispatch OpenSpec verification automatically. The verifier MUST run the repository's mandatory tests and build and report their commands with exit codes. The orchestrator MUST stop on failed or incomplete verification, and after successful verification it MUST run the existing post-verification review gate unchanged.

#### Scenario: Final task completes cleanly
- **WHEN** the refreshed task tree contains no pending tasks after a clean implementation batch
- **THEN** the orchestrator dispatches the OpenSpec verifier without asking a continuation question

#### Scenario: Verification succeeds with executable evidence
- **WHEN** the verifier reports successful mandatory tests and build with command and exit-code evidence
- **THEN** the orchestrator presents the existing review selection gate before archive

#### Scenario: Verification is not successful
- **WHEN** mandatory tests or build fail, verification blocks, or executable evidence is incomplete
- **THEN** the orchestrator stops before the review gate, reports the verification result, and then asks one contextual next-action question before any retry or other recovery action

#### Scenario: Selected review findings are fixed
- **WHEN** the user selects post-verification review findings for a bounded fix batch
- **THEN** the orchestrator retains the existing finding-ID routing without requiring `tasks.md` task or section identifiers or automatically repeating verification

### Requirement: Canonical asset and global agent synchronization
Implementation MUST update the repository orchestrator asset as the canonical instruction source and then replace the live global orchestrator file from it without creating a backup. The two files MUST be byte-for-byte identical after the update, and final verification MUST repeat the equality check.

#### Scenario: Instructions are deployed
- **WHEN** the repository asset has passed focused validation
- **THEN** implementation overwrites `$HOME/.config/opencode/agents/angel-orchestrator.md` from `assets/agents/angel-orchestrator.md` without creating a backup and confirms equality

#### Scenario: Synchronization is rechecked before verification completes
- **WHEN** final verification runs after the revised implementation
- **THEN** it confirms that the repository and installed global orchestrator files remain byte-for-byte identical
