## ADDED Requirements

### Requirement: Automatic planned-task execution
When pending planned tasks exist, the orchestrator SHALL automatically execute all remaining tasks without asking a cadence or between-section continuation question. It MUST continue through clean section-bounded batches until no tasks remain or a mandatory stop occurs.

#### Scenario: Pending work spans multiple sections
- **WHEN** fresh task state shows pending tasks in more than one section
- **THEN** the orchestrator automatically dispatches each next incomplete section in sequence without asking for cadence or continuation between clean results

#### Scenario: All planned tasks are already complete
- **WHEN** the initial fresh task state contains no pending tasks
- **THEN** the orchestrator skips implementation dispatch and proceeds automatically to final verification

## MODIFIED Requirements

### Requirement: Current OpenSpec task state drives planned-task implementation
At every planned-task implementation decision point, the orchestrator MUST query the real OpenSpec status for the active change, read the resolved current `tasks.md`, and recompute the task tree and next batch. It MUST apply this invariant before the initial task-tree display, before every planned-task worker dispatch, and after every clean worker result. It MUST NOT select a planned-task batch from conversational history or a cached task list.

#### Scenario: Task state changed since the previous batch
- **WHEN** a prior batch or external actor changes task completion state
- **THEN** the orchestrator selects the next batch from a fresh OpenSpec status query and `tasks.md` read

#### Scenario: Current tasks cannot be resolved
- **WHEN** OpenSpec status does not report a complete tasks artifact or the resolved `tasks.md` cannot be read
- **THEN** the orchestrator stops implementation and reports the flow as blocked

### Requirement: Complete compact task tree
Before planned-task implementation begins, the orchestrator SHALL display the complete task hierarchy in a compact tree. At every mandatory implementation stop, it MUST refresh and display the complete tree when current task state can be resolved. The tree MUST show completed and total counts at the root and for each section, and every task MUST show its identifier, short text, and either `✓` when complete or `☐` when pending. Clean transitions between automatic section batches MUST NOT introduce a task-tree pause.

#### Scenario: Mixed task completion before implementation
- **WHEN** `tasks.md` contains completed and pending tasks across multiple sections before the first dispatch
- **THEN** the rendered tree includes every section and task with accurate section progress and completion markers

#### Scenario: Mandatory stop with resolvable task state
- **WHEN** a mandatory stop occurs and fresh OpenSpec task state can be resolved
- **THEN** the orchestrator displays the complete refreshed compact tree with the stop evidence before waiting for user direction

#### Scenario: Mandatory stop with unresolvable task state
- **WHEN** a mandatory stop occurs because current OpenSpec task state cannot be resolved
- **THEN** the orchestrator reports the blocking evidence and that the complete tree is unavailable without using cached task state

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
The orchestrator MUST stop automatic progression after an implementation result of `blocked` or `partial`, any command failure not corrected by a later equivalent-or-broader relevant successful command, a reported deviation or out-of-scope change, or a refreshed task state that conflicts with the requested batch. It MUST surface the evidence and MUST NOT improvise replacement work. When task state remains resolvable, it MUST also display the refreshed complete task tree at the stop.

#### Scenario: Worker reports a non-clean status
- **WHEN** an implementation worker reports `blocked` or `partial`
- **THEN** the orchestrator stops before dispatching another implementation batch and presents the refreshed complete tree when available

#### Scenario: A command remains failed
- **WHEN** an invoked command reports a non-zero exit code without a later equivalent-or-broader relevant successful rerun
- **THEN** the orchestrator stops before dispatching another implementation batch and reports the failed command

#### Scenario: Plan deviation is detected
- **WHEN** the worker reports a deviation or out-of-scope work, or refreshed task state conflicts with the requested batch
- **THEN** the orchestrator stops, presents the refreshed complete tree when available, and surfaces the deviation without selecting substitute work

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
- **THEN** the orchestrator stops before the review gate and reports the verification result

#### Scenario: Selected review findings are fixed
- **WHEN** the user selects post-verification review findings for a bounded fix batch
- **THEN** the orchestrator retains the existing finding-ID routing without requiring `tasks.md` task or section identifiers or automatically repeating verification

### Requirement: Manual validation without new tests
Each planned-task implementation batch MUST use only focused textual checks relevant to its instruction changes and MUST defer mandatory repository tests and build to final OpenSpec verification. The change MUST NOT add automated tests. Any focused check that is run remains subject to the mandatory implementation stop conditions.

#### Scenario: Section implementation is ready to report
- **WHEN** an implementer completes its bounded instruction changes
- **THEN** it runs focused textual checks for the section and does not run the repository's mandatory tests or build

#### Scenario: Final verification begins
- **WHEN** fresh task state shows all planned tasks complete
- **THEN** the verifier runs the mandatory repository tests and build and reports their exit codes

## REMOVED Requirements

### Requirement: One-time cadence selection
**Reason**: Planned-task execution no longer offers selectable pause modes; automatic execution is the only supported flow.

**Migration**: Remove the cadence question, retained cadence state, and task- or section-boundary continuation pauses. Use automatic section-bounded execution as defined above.
