## MODIFIED Requirements

### Requirement: Corrected intermediate failures use a shared evidence policy
For every implementation result and every strict-default control point that invokes the shared implementation-result policy, the orchestrator SHALL classify an intermediate non-zero command, including a real implementation or verification failure, as corrected only when the same worker diagnoses and repairs its cause within the authorized scope, later runs an equivalent-or-broader relevant validation command with exit code zero, returns final status `done`, ends with a green final relevant validation state, and reports no deviation, scope expansion, or out-of-scope work. The successful validation MUST cover the failed command's relevant scope or a superset of it. Applicable worker return contracts and orchestrator reporting SHALL preserve, in execution order, the failed command and exit code, the diagnosed cause and bounded correction, and the successful validation command and exit code. An eligible corrected failure SHALL be treated as clean for the control point's existing routing and SHALL NOT by itself cause an authorization question, mandatory stop, or archive delay. These rules do not remove any route-local prerequisite or planned-task-specific safeguard.

#### Scenario: Worker corrects a real bounded failure
- **WHEN** the same worker diagnoses a real intermediate failure, repairs its cause within authorized scope, runs equivalent-or-broader relevant validation with exit code zero, returns `done` with a green final relevant state, and reports no deviation, scope expansion, or out-of-scope work
- **THEN** the orchestrator surfaces the failed command, correction, and successful validation and continues through the control point's existing clean-result route without requesting authorization solely for that incident

#### Scenario: Final verification corrects a real intermediate failure
- **WHEN** the final-verification worker satisfies every shared corrected-failure condition and all other verification and archive prerequisites are met
- **THEN** the orchestrator surfaces the complete incident evidence and continues to the existing archive path without stopping solely because of the corrected failure

#### Scenario: Tooling mistake remains eligible under the common rule
- **WHEN** a syntax or invocation mistake satisfies every shared corrected-failure condition
- **THEN** the orchestrator classifies it through the same evidence policy and follows the applicable clean-result route

#### Scenario: Another worker supplies the successful validation
- **WHEN** a failed command and its successful equivalent-or-broader validation come from different workers
- **THEN** the orchestrator treats the correction evidence as insufficient and applies the shared mandatory-stop policy

#### Scenario: Green command does not cover the failed scope
- **WHEN** a worker reports a later successful command that is unrelated to or narrower than the relevant scope of an earlier failed command
- **THEN** the orchestrator treats the failure as uncorrected and applies the shared mandatory-stop policy

#### Scenario: Correction evidence is incomplete
- **WHEN** the result omits the failed command or exit code, diagnosed cause, bounded correction, successful command or exit code, or evidence that the successful validation covers the failed scope
- **THEN** the orchestrator treats the result as non-clean and applies the shared mandatory-stop policy

#### Scenario: Corrected command accompanies an unsafe final result
- **WHEN** a worker reports a successful equivalent-or-broader rerun but ends `partial` or `blocked`, has a red final relevant state, or reports a deviation, scope expansion, or out-of-scope work
- **THEN** the orchestrator applies the shared mandatory-stop policy

### Requirement: Unsafe final implementation results require a stop
Every strict-default control point that invokes the shared implementation-result policy MUST apply mandatory-stop handling when a non-zero command lacks an evidence-complete same-worker correction and equivalent-or-broader relevant validation exiting zero, the final relevant evidence is red, status is `partial` or `blocked`, correction evidence is insufficient, or a deviation, scope expansion, or out-of-scope action is reported. An evidence-complete corrected intermediate failure with a green final result is not unsafe and MUST NOT cause a stop by itself. A section-bounded planned OpenSpec task batch MAY instead defer only eligible local `partial`, local `blocked`, or red focused-test outcomes while its unfinished tasks remain unchecked; hard blockers and checked-task red-validation conflicts MUST stop immediately, and unresolved outcomes MUST stop after the single final retry round.

#### Scenario: Strict-default failure has no clean rerun
- **WHEN** a strict-default worker reports a non-zero command without an evidence-complete same-worker correction and later equivalent-or-broader relevant command exiting zero
- **THEN** the orchestrator surfaces the failed command and exit code and stops without deferral

#### Scenario: Strict-default final state remains red
- **WHEN** the final relevant state at a strict-default control point is red
- **THEN** the orchestrator reports the final red evidence and stops without deferral

#### Scenario: Strict-default worker returns incomplete status
- **WHEN** a strict-default worker returns `partial` or `blocked`
- **THEN** the orchestrator reports the incomplete result and applies the shared mandatory-stop policy

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
