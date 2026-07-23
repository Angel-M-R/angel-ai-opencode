## ADDED Requirements

### Requirement: Planned batch outcomes are classified for deferral
Only a result from a bounded planned OpenSpec task batch selected from the active change's fresh `tasks.md` MAY be deferred. The orchestrator SHALL classify local `partial`, local `blocked`, or red focused-test outcomes as deferrable only when the affected incomplete tasks remain unchecked and no hard blocker exists. It MUST retain the batch identity, task state, status, commands and exit codes, focused-validation evidence, blocker or incomplete-work reason, files touched, and deviations.

#### Scenario: Partial batch leaves unfinished tasks pending
- **WHEN** a planned batch returns `partial`, its unfinished tasks remain unchecked, and no hard blocker exists
- **THEN** the orchestrator records the batch as deferred instead of marking the tasks complete or immediately asking the user a question

#### Scenario: Local blocked result is deferrable
- **WHEN** a planned batch reports a local `blocked` condition, its tasks remain unchecked, and no hard blocker exists
- **THEN** the orchestrator retains the blocker evidence and may defer the batch under the independence rule

#### Scenario: Focused test remains red
- **WHEN** a focused test of code modified by the planned batch remains red and the relevant tasks remain unchecked
- **THEN** the orchestrator retains the red command evidence and may defer the batch without treating those tasks as complete

### Requirement: Later work requires explicit independence evidence
After deferring a planned batch, the orchestrator MUST dispatch a later section only when current planning artifacts, bounded task scopes, or retained worker diagnostics provide explicit evidence that the later work does not consume, validate, or depend on the pending batch. Missing, conflicting, or ambiguous evidence MUST be treated as dependency.

#### Scenario: Later section is explicitly independent
- **WHEN** fresh state shows a later pending section and explicit evidence establishes that it does not depend on a deferred batch
- **THEN** the orchestrator dispatches that later section as the next bounded batch without an intermediate user question

#### Scenario: Independence is uncertain
- **WHEN** the orchestrator cannot establish independence from explicit current evidence
- **THEN** it treats the later section as dependent and does not dispatch it

#### Scenario: Section order is the only evidence
- **WHEN** a pending section merely appears after the deferred section in `tasks.md`
- **THEN** the orchestrator does not treat ordering alone as evidence of independence

### Requirement: Deferred work receives one final retry round
After the orchestrator has considered all work that is explicitly independent of deferred batches, it SHALL refresh state, present one combined report of all deferred incidents and benign deviations, and run exactly one final retry round without asking an intermediate question. Each still-pending deferred batch MUST be retried at most once from fresh state and MUST NOT be re-queued for another round.

#### Scenario: Multiple incidents were deferred
- **WHEN** the independent forward pass ends with more than one deferred batch
- **THEN** the orchestrator reports the incidents and benign deviations together before starting the retry round

#### Scenario: Deferred task changed before retry
- **WHEN** fresh state before retry shows that some deferred tasks are already complete or otherwise changed
- **THEN** the orchestrator recomputes the bounded retry from current unchecked tasks and does not dispatch stale work

#### Scenario: A retried batch remains unresolved
- **WHEN** a deferred batch has received its one retry and remains pending, locally blocked, or red
- **THEN** the orchestrator does not queue another retry and retains the final evidence for mandatory-stop handling after the round

#### Scenario: Retry round exhausts runnable work
- **WHEN** the final retry round leaves pending tasks with no runnable batch
- **THEN** the orchestrator reports the refreshed state and retained evidence, then applies the mandatory-stop interaction

### Requirement: Planned batch validation preserves verifier ownership
The planned-task implementer SHALL run validation relevant to its bounded changes and MAY run focused tests that exercise modified code. It MUST NOT run the full repository test suite or builds, which remain reserved for the final OpenSpec verifier. Additional reads and successful focused tests SHALL be retained as benign, continuable deviations when they serve only the bounded batch.

#### Scenario: Implementer runs a focused test
- **WHEN** a planned-task implementer runs a test targeted to code modified in its bounded batch and the test succeeds
- **THEN** the orchestrator records the command as a benign deviation and may continue

#### Scenario: Implementer performs an additional relevant read
- **WHEN** the implementer reads additional repository context needed for the bounded batch without writing outside it
- **THEN** the orchestrator records the read deviation and may continue

#### Scenario: Implementer attempts full verification
- **WHEN** a planned-task implementer runs a full repository suite or build
- **THEN** the orchestrator treats the command as a blocking policy violation rather than final verification

### Requirement: Hard blockers cannot be deferred
The planned-task loop MUST stop immediately when a worker writes outside the bounded batch, expands functional scope, runs a destructive command, OpenSpec state cannot be resolved safely, a checked task has relevant red validation, or another result leaves unsafe final evidence that is not eligible for local deferral. It MUST NOT ignore red tasks or alter checkbox state to manufacture completion.

#### Scenario: Worker writes outside the batch
- **WHEN** worker evidence shows a write outside the exact planned batch
- **THEN** the orchestrator stops and reports the scope violation without deferring it

#### Scenario: Worker expands functionality
- **WHEN** a planned-task worker implements behavior not required by its bounded tasks
- **THEN** the orchestrator stops and reports the functional expansion

#### Scenario: Destructive command is reported
- **WHEN** a planned-task result contains a destructive command
- **THEN** the orchestrator stops immediately and preserves the command evidence

#### Scenario: Checked task remains red
- **WHEN** fresh `tasks.md` marks a task complete while relevant retained validation remains red
- **THEN** the orchestrator reports a real `tasks.md` state conflict and stops without unchecking, ignoring, or relabeling the task

#### Scenario: OpenSpec state is unresolvable
- **WHEN** fresh status or the resolved `tasks.md` cannot establish current task state
- **THEN** the orchestrator stops the planned-task cycle and does not use cached state for deferral or retry

### Requirement: Orchestrator policy and contract tests change together
Implementation MUST update `assets/agents/angel-orchestrator.md` and the corresponding focused contracts in `internal/install/agent_assets_test.go` together. The tests MUST distinguish deferrable planned-batch outcomes from unchanged strict routes and MUST cover the independence gate, combined report, single retry round, focused-test allowance, and checkbox/red-evidence conflict.

#### Scenario: Planned-batch policy is implemented
- **WHEN** the orchestrator prompt gains deferral behavior
- **THEN** focused contract tests assert that behavior and continue to assert strict Direct, bootstrap, target-resolution, review-fix, and final-verification handling

## MODIFIED Requirements

### Requirement: Automatic planned-task execution
When pending planned tasks exist, the orchestrator SHALL automatically execute all safe runnable work without asking a cadence or between-section continuation question. It MUST continue through clean section-bounded batches and, after a deferrable outcome, through later sections only when explicit evidence establishes independence. After the independent forward pass, it MUST perform the single final deferred-batch retry round. It SHALL stop when no tasks remain, a hard blocker occurs, or the retry round ends with pending work or final red evidence.

#### Scenario: Pending work spans multiple clean sections
- **WHEN** fresh task state shows pending tasks in more than one section and each batch completes cleanly
- **THEN** the orchestrator automatically dispatches each next incomplete section in sequence without asking for cadence or continuation

#### Scenario: A batch is deferred before independent work
- **WHEN** a batch has a deferrable outcome and explicit evidence shows later pending sections are independent
- **THEN** the orchestrator records the incident and continues through those independent sections without an intermediate question

#### Scenario: All planned tasks are already complete
- **WHEN** the initial fresh task state contains no pending tasks
- **THEN** the orchestrator skips implementation dispatch and proceeds automatically to final verification

#### Scenario: Final retry leaves incomplete work
- **WHEN** the single retry round ends with unchecked tasks, no runnable work, or final red evidence
- **THEN** the orchestrator applies the mandatory-stop interaction and does not begin verification

### Requirement: Mandatory implementation stop conditions
Strict report-first mandatory stops SHALL remain unchanged for OpenSpec target resolution, bootstrap, final verification, Direct Safe or Fast execution, and Direct review-fix routing. Inside the planned-task loop, the orchestrator MUST stop immediately for hard blockers, but SHALL defer eligible local `partial`, local `blocked`, and red focused-test outcomes while explicitly independent work remains. After the single retry round, any remaining pending work, unavailable runnable batch, unresolved block, or final red evidence MUST stop. At a stop, the orchestrator MUST preserve and report the evidence before asking exactly one contextual next-action question, and MUST perform no recovery action until the user responds.

#### Scenario: Planned worker reports an eligible non-clean status
- **WHEN** a planned-task worker reports `partial` or local `blocked`, affected tasks remain unchecked, and no hard blocker exists
- **THEN** the orchestrator records the deferred incident and applies the independence gate instead of immediately asking a question

#### Scenario: Plan scope violation is detected
- **WHEN** the worker reports out-of-batch writes, functional expansion, a destructive command, or a checked-task red-validation conflict
- **THEN** the orchestrator stops, presents the refreshed tree when available, reports the evidence, and then asks one contextual next-action question

#### Scenario: Bootstrap or target resolution blocks
- **WHEN** an existing OpenSpec target cannot be resolved or an OpenSpec readiness bootstrap blocks or fails
- **THEN** the orchestrator reports the target or bootstrap diagnostic, then asks one contextual next-action question before retrying, falling back, or dispatching an OpenSpec worker

#### Scenario: Direct route stops
- **WHEN** a Direct Safe or Fast implementation or a Direct review-fix result meets a mandatory-stop condition
- **THEN** the orchestrator reports the retained result and verification evidence, then asks one contextual next-action question before retrying, broadening scope, opening reviews, or dispatching another worker

#### Scenario: Final verification is red
- **WHEN** final OpenSpec verification fails, blocks, or lacks executable evidence
- **THEN** the orchestrator reports the final evidence and asks one contextual next-action question without deferral

#### Scenario: Stop question is presented
- **WHEN** blocking evidence has been reported for any mandatory stop
- **THEN** the orchestrator asks exactly one blocker-specific question whose choices include a safe stop option and whose question tool permits a custom response

#### Scenario: User has not selected an action
- **WHEN** the mandatory-stop question is unanswered
- **THEN** the orchestrator performs no retry, continuation, scope expansion, substitute selection, phase advance, or worker dispatch

### Requirement: Automatic verification and retained review gate
After a fresh task-state read shows every task complete and no relevant validation remains red, the orchestrator MUST dispatch OpenSpec verification automatically. The verifier MUST run the repository's mandatory tests and build and report their commands with exit codes. The orchestrator MUST stop strictly on failed or incomplete verification, and after successful verification it MUST run the existing post-verification review gate unchanged.

#### Scenario: Final task completes after the retry round
- **WHEN** the refreshed task tree contains no pending tasks and retained relevant evidence is green
- **THEN** the orchestrator dispatches the OpenSpec verifier without asking a continuation question

#### Scenario: Pending tasks remain after retry
- **WHEN** the final retry round ends with one or more unchecked tasks
- **THEN** the orchestrator stops before final verification and reports the retained batch evidence

#### Scenario: Verification succeeds with executable evidence
- **WHEN** the verifier reports successful mandatory tests and build with command and exit-code evidence
- **THEN** the orchestrator presents the existing review selection gate before archive

#### Scenario: Verification is not successful
- **WHEN** mandatory tests or build fail, verification blocks, or executable evidence is incomplete
- **THEN** the orchestrator stops before the review gate, reports the verification result, and then asks one contextual next-action question before any retry or other recovery action

#### Scenario: Selected review findings are fixed
- **WHEN** the user selects post-verification review findings for a bounded fix batch
- **THEN** the orchestrator retains the existing finding-ID routing without requiring `tasks.md` task or section identifiers or automatically repeating verification
