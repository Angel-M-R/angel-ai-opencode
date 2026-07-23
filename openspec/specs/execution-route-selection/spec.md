## ADDED Requirements

### Requirement: Route selection follows Brief confirmation
For non-trivial work that does not target an existing OpenSpec change, the orchestrator SHALL ask the user to select OpenSpec or Direct execution after the Brief is confirmed and before it performs OpenSpec bootstrap, invokes an OpenSpec worker, runs an OpenSpec CLI command for the new work, or creates an OpenSpec change or artifact.

#### Scenario: New work reaches the route gate in order
- **WHEN** the user confirms a Brief for non-trivial new work
- **THEN** the orchestrator asks for OpenSpec or Direct execution before causing any OpenSpec side effect for that work

#### Scenario: Unconfirmed work does not reach route selection
- **WHEN** a non-trivial request has not yet produced a user-confirmed Brief
- **THEN** the orchestrator completes the existing interview gate before asking for an execution route

### Requirement: Recommendation is risk-based and non-blocking
The orchestrator SHALL recommend Direct for clear, isolated, reversible changes and SHALL recommend OpenSpec for architecture, security, data, migrations, cross-cutting scope, or material uncertainty, while allowing the user to select either route regardless of the recommendation.

#### Scenario: Low-risk bounded change
- **WHEN** the confirmed Brief describes a clear, isolated, reversible change
- **THEN** the orchestrator recommends Direct and accepts either route selection

#### Scenario: High-risk or uncertain change
- **WHEN** the confirmed Brief affects architecture, security, data, migrations, cross-cutting scope, or has material uncertainty
- **THEN** the orchestrator recommends OpenSpec and accepts either route selection

### Requirement: Existing OpenSpec changes remain in OpenSpec
The orchestrator MUST continue work that targets an existing OpenSpec change through the status-driven OpenSpec workflow and MUST NOT offer or use Direct execution for that change.

#### Scenario: Request names an existing change
- **WHEN** the requested work targets an OpenSpec change that already exists
- **THEN** the orchestrator resolves its current status and routes the next action through the existing OpenSpec workflow

### Requirement: OpenSpec selection preserves the complete workflow
When OpenSpec is selected for new work, the orchestrator SHALL preserve the existing bootstrap gate, official planning-worker routing, artifact lifecycle, bounded implementation cadence, verification policy, review gate, review-fix routing, and archive path. Only section-bounded planned-task implementation MAY use the evidence-gated deferral and single-retry policy defined by `implementation-cadence-selector`; Direct execution and every OpenSpec action outside that loop MUST retain strict result handling.

#### Scenario: User chooses OpenSpec
- **WHEN** the user selects OpenSpec at the execution-route gate
- **THEN** the orchestrator enters the current OpenSpec workflow without substituting direct workers or omitting an existing phase

#### Scenario: Planned batch returns a deferrable result
- **WHEN** a section-bounded planned-task batch meets every deferral eligibility condition
- **THEN** the orchestrator may continue only through explicitly independent planned sections under the implementation cadence policy

#### Scenario: Non-batch OpenSpec action is non-clean
- **WHEN** bootstrap, target resolution, final verification, or a review-fix batch reports a non-clean result
- **THEN** the orchestrator applies the existing strict mandatory-stop policy without planned-batch deferral

### Requirement: Corrected intermediate failures use a shared evidence policy
For every planned OpenSpec implementation batch and every Direct Safe worker result, including a bounded Direct Safe review-fix result, the orchestrator SHALL classify an intermediate non-zero command as corrected only when the same worker identifies the failure, later runs an equivalent or broader relevant command with exit code zero, returns final status `done`, and reports no deviation or out-of-scope work. The successful command MUST validate the failed command's relevant scope or a superset of it. The orchestrator SHALL retain and surface the failed command, its exit code, and the successful rerun instead of hiding the intermediate failure.

#### Scenario: Planned OpenSpec batch corrects an intermediate failure
- **WHEN** an OpenSpec implementation worker identifies a non-zero command, later runs an equivalent or broader relevant command successfully, returns `done`, and reports no deviation or out-of-scope work
- **THEN** the orchestrator surfaces the corrected-failure evidence and may continue or pause under the already selected cadence

#### Scenario: Direct Safe worker corrects an intermediate failure
- **WHEN** a Direct Safe worker identifies a non-zero command, later runs an equivalent or broader relevant command successfully, returns `done`, and reports no deviation or out-of-scope work
- **THEN** the orchestrator surfaces the corrected-failure evidence and may proceed to the applicable Direct Safe review step

#### Scenario: Green command does not cover the failed scope
- **WHEN** a worker reports a later successful command that is unrelated to or narrower than the relevant scope of an earlier failed command
- **THEN** the orchestrator treats the failure as uncorrected and stops

### Requirement: Unsafe final implementation results require a stop
Direct Safe worker results, bounded Direct Safe review-fix results, OpenSpec bootstrap and target-resolution failures, and final OpenSpec verification failures MUST retain strict mandatory-stop handling when a non-zero command lacks a clean equivalent-or-broader relevant rerun, final relevant evidence is red, status is `partial` or `blocked`, or a deviation or out-of-scope action is reported. A section-bounded planned OpenSpec task batch MAY instead defer only eligible local `partial`, local `blocked`, or red focused-test outcomes while its unfinished tasks remain unchecked; hard blockers and checked-task red-validation conflicts MUST stop immediately, and unresolved outcomes MUST stop after the single final retry round.

#### Scenario: Direct Safe failure has no clean rerun
- **WHEN** a Direct Safe worker reports a non-zero command without a later equivalent-or-broader relevant command exiting zero
- **THEN** the orchestrator surfaces the failed command and exit code and stops without deferral

#### Scenario: Direct or final-verification state remains red
- **WHEN** the final relevant state for Direct work or final OpenSpec verification is red
- **THEN** the orchestrator reports the final red evidence and stops without deferral

#### Scenario: Planned focused test remains red with unchecked tasks
- **WHEN** a section-bounded planned-task focused test remains red, its relevant tasks remain unchecked, and no hard blocker exists
- **THEN** the orchestrator may defer the batch and continue only through explicitly independent planned work

#### Scenario: Planned task is checked while validation is red
- **WHEN** fresh task state marks a planned task complete while its relevant validation remains red
- **THEN** the orchestrator reports a real task-state conflict and stops immediately

#### Scenario: Planned retry remains unresolved
- **WHEN** the one final retry round ends with pending work, no runnable batch, an unresolved local block, or final red evidence
- **THEN** the orchestrator reports the retained evidence and stops before final verification

#### Scenario: Worker exceeds scope
- **WHEN** any worker writes outside its bounded work or expands functional scope
- **THEN** the orchestrator reports the scope violation and stops without deferral

### Requirement: Direct execution is isolated from OpenSpec
The Direct route MUST NOT run OpenSpec bootstrap, invoke the OpenSpec CLI, generate or modify OpenSpec artifacts, or delegate implementation to the orchestrator, `openspec-implementer`, or any other OpenSpec worker. Direct implementation SHALL be delegated to `general`.

#### Scenario: User chooses Direct
- **WHEN** the user selects Direct execution
- **THEN** the orchestrator asks for Safe or Fast mode and delegates the resulting bounded implementation to `general` without any OpenSpec action

### Requirement: Direct prompts are bounded and auditable
Every direct implementation prompt SHALL include the confirmed Brief, the selected Safe or Fast mode, explicit scope limits, and a return contract requiring status as `done`, `partial`, or `blocked`; files touched; commands run in order with exit codes; any deviations from the Brief or scope; and, when applicable, identification of each failed command and its equivalent-or-broader successful rerun.

#### Scenario: Direct worker is dispatched
- **WHEN** the orchestrator delegates a Safe or Fast direct implementation
- **THEN** the `general` worker receives the Brief, mode, bounded scope, complete result contract, and corrected-failure evidence requirement

### Requirement: Safe mode combines implementation and executable verification
In Safe mode, the same `general` worker SHALL implement the bounded Brief and run the repository's existing applicable tests and build commands. The orchestrator SHALL consider the result clean only when the worker reports `done`, executable verification was available and run, the final relevant verification state is green under the shared corrected-intermediate-failure policy, and no Brief deviation or out-of-scope work is reported.

#### Scenario: Safe execution is clean
- **WHEN** the Safe worker reports `done`, runs applicable existing tests/builds with zero exit codes, and reports no deviation or out-of-scope work
- **THEN** the orchestrator proceeds to the direct Safe review gate

#### Scenario: Safe execution corrects an intermediate failure
- **WHEN** the Safe worker satisfies every condition of the shared corrected-intermediate-failure policy and ends with a green relevant verification state
- **THEN** the orchestrator surfaces both the failure and successful rerun evidence and proceeds to the direct Safe review gate

#### Scenario: No executable verification exists
- **WHEN** the Safe worker cannot identify or run an applicable existing test or build command
- **THEN** the orchestrator reports the result as `partial` or `blocked` and stops

### Requirement: Direct mode stops without orchestrator retry on unsafe results
The orchestrator MUST stop the direct flow and report the evidence without retrying, dispatching a fallback worker, opening reviews, or continuing implementation when executable verification is unavailable or a Direct Safe result meets any shared mandatory-stop condition. A same-worker corrected intermediate failure is not an orchestrator retry and does not by itself require a stop.

#### Scenario: Test or build failure remains uncorrected
- **WHEN** a Safe worker test or build command has a non-zero exit code without a clean equivalent-or-broader relevant rerun
- **THEN** the orchestrator reports the failed command and exit code and performs no further direct action

#### Scenario: Worker deviates from the Brief
- **WHEN** the Safe worker reports work outside the confirmed Brief or bounded scope
- **THEN** the orchestrator reports the deviation and stops without retry

#### Scenario: Direct worker returns incomplete status
- **WHEN** the Safe worker returns `partial` or `blocked`
- **THEN** the orchestrator reports the blocker or partial result and stops without retry

### Requirement: Safe review routing uses bounded direct workers
After a clean Safe result, the orchestrator SHALL offer Security risk, Simplicity, Correctness, or no review; run only selected reviewers against the bounded direct diff and confirmed Brief; and present deduplicated findings for user selection. It MUST delegate only user-selected findings as one bounded fix batch to `general`, MUST NOT use `openspec-implementer`, and after a clean fix SHALL permit rerunning only reviewers whose selected findings were addressed.

#### Scenario: User selects direct reviews
- **WHEN** a Safe result is clean and the user selects one or more review types
- **THEN** the orchestrator runs only those reviewers against the direct diff and confirmed Brief

#### Scenario: User selects findings to fix
- **WHEN** the user selects findings from the direct review result
- **THEN** the orchestrator sends exactly those findings in one bounded `general` batch with the direct result contract

#### Scenario: Direct review fixes are incomplete
- **WHEN** the direct review-fix worker returns a non-clean result
- **THEN** the orchestrator reports and stops without retrying or rerunning a reviewer

#### Scenario: User requests review confirmation after fixes
- **WHEN** selected findings are fixed cleanly and the user wants confirmation
- **THEN** the orchestrator reruns only the reviewers responsible for the addressed findings

### Requirement: Fast mode is implemented but unverified
In Fast mode, the `general` worker SHALL implement only the bounded Brief and MUST NOT run tests or reviews. The orchestrator SHALL explicitly report the result as implemented but not verified and MUST NOT open the review gate.

#### Scenario: Fast worker completes
- **WHEN** the Fast worker completes the bounded implementation
- **THEN** the orchestrator reports it as implemented but unverified without running tests or reviews
