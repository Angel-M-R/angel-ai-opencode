# implementation-cadence-selector Specification

## Purpose
TBD - created by archiving change add-openspec-implementation-cadence. Update Purpose after archive.
## Requirements
### Requirement: Current OpenSpec task state drives planned-task implementation
At every planned-task implementation decision point, the orchestrator MUST query the real OpenSpec status for the active change, read the resolved current `tasks.md`, and recompute the task tree and next batch. It MUST apply this invariant before the initial cadence question, before every planned-task worker dispatch, and after every clean worker result. It MUST NOT select a planned-task batch from conversational history or a cached task list.

#### Scenario: Task state changed since the previous batch
- **WHEN** a prior batch or external actor changes task completion state
- **THEN** the orchestrator selects the next batch from a fresh OpenSpec status query and `tasks.md` read

#### Scenario: Current tasks cannot be resolved
- **WHEN** OpenSpec status does not report a complete tasks artifact or the resolved `tasks.md` cannot be read
- **THEN** the orchestrator stops implementation and reports the flow as blocked

### Requirement: Complete compact task tree
Before asking for implementation cadence, the orchestrator SHALL display the complete task hierarchy in a compact tree. The tree MUST show completed and total counts at the root and for each section, and every task MUST show its identifier, short text, and either `✓` when complete or `☐` when pending.

#### Scenario: Mixed task completion
- **WHEN** `tasks.md` contains completed and pending tasks across multiple sections
- **THEN** the rendered tree includes every section and task with accurate section progress and completion markers

#### Scenario: Control returns at a cadence boundary
- **WHEN** task or section cadence pauses after a clean batch
- **THEN** the orchestrator refreshes and displays the complete compact tree before waiting for continuation

### Requirement: One-time cadence selection
The orchestrator SHALL ask exactly once per planned-task implementation cycle whether to pause after each task, pause after each section, or execute all remaining tasks. It MUST retain that selection for the cycle without asking the cadence question again.

#### Scenario: Pause after each task
- **WHEN** the user selects task cadence
- **THEN** the orchestrator dispatches exactly the next pending task and returns control after a clean result

#### Scenario: Pause after each section
- **WHEN** the user selects section cadence
- **THEN** the orchestrator dispatches the pending tasks in exactly the next incomplete section and returns control after a clean result

#### Scenario: Execute all remaining tasks
- **WHEN** the user selects run-all cadence
- **THEN** the orchestrator automatically chains bounded batches without further continuation or cadence questions while all results remain clean

### Requirement: Bounded implementation batches
Every planned-task implementation worker prompt MUST identify an exact bounded task batch. Task cadence batches MUST contain one pending task. Section cadence and run-all batches MUST contain only the pending tasks from one named section, and run-all MUST chain those section-bounded batches rather than issue an unbounded completion request.

#### Scenario: Run-all spans multiple sections
- **WHEN** pending tasks exist in more than one section and run-all cadence is active
- **THEN** the orchestrator dispatches one explicitly identified section batch at a time and refreshes task state before selecting each subsequent batch

#### Scenario: Intended batch is already complete
- **WHEN** the refreshed `tasks.md` shows that the intended task or section no longer has pending work
- **THEN** the orchestrator skips the stale batch and recomputes the next bounded batch

### Requirement: Mandatory implementation stop conditions
The orchestrator MUST stop cadence progression and automatic chaining after an implementation result of `blocked` or `partial`, a failed test or build command, a reported deviation from the plan, or a refreshed task state that conflicts with the requested batch. It MUST surface the evidence and MUST NOT improvise replacement work.

#### Scenario: Worker reports a non-clean status
- **WHEN** an implementation worker reports `blocked` or `partial`
- **THEN** the orchestrator stops before dispatching another implementation batch

#### Scenario: Test or build fails
- **WHEN** an invoked test or build command reports a non-zero exit code
- **THEN** the orchestrator stops before dispatching another implementation batch and reports the failed command

#### Scenario: Plan deviation is detected
- **WHEN** the worker reports a deviation from the planned batch or refreshed task state conflicts with that batch
- **THEN** the orchestrator stops and surfaces the deviation without selecting substitute work

### Requirement: Automatic verification and retained review gate
After a fresh task-state read shows every task complete, the orchestrator MUST dispatch OpenSpec verification automatically. It MUST stop on failed or incomplete verification, and after successful verification it MUST run the existing post-verification review gate unchanged.

#### Scenario: Final task completes cleanly
- **WHEN** the refreshed task tree contains no pending tasks after a clean implementation batch
- **THEN** the orchestrator dispatches the OpenSpec verifier without asking another continuation question

#### Scenario: Verification succeeds
- **WHEN** the verifier reports successful executed evidence
- **THEN** the orchestrator presents the existing review selection gate before archive

#### Scenario: Verification is not successful
- **WHEN** verification fails, blocks, or is incomplete
- **THEN** the orchestrator stops before the review gate and reports the verification result

#### Scenario: Selected review findings are fixed
- **WHEN** the user selects post-verification review findings for a bounded fix batch
- **THEN** the orchestrator retains the existing finding-ID routing without requiring `tasks.md` task or section identifiers, reopening cadence, or automatically repeating verification

### Requirement: Canonical asset and global agent synchronization
Implementation MUST update the repository orchestrator asset as the canonical instruction source and then replace the live global orchestrator file from it without creating a backup. The two files MUST be byte-for-byte identical after the update.

#### Scenario: Instructions are deployed
- **WHEN** the repository asset has passed manual validation
- **THEN** implementation overwrites `$HOME/.config/opencode/agents/angel-orchestrator.md` from `assets/agents/angel-orchestrator.md` without creating a backup and confirms equality

### Requirement: Manual validation without new tests
The change MUST be validated through manual instruction and flow walkthroughs and MUST NOT add automated tests.

#### Scenario: Change validation
- **WHEN** the cadence instructions are ready for deployment
- **THEN** the implementer manually checks all cadence choices, mandatory stop conditions, automatic verification, retained review routing, and file synchronization without adding test files

