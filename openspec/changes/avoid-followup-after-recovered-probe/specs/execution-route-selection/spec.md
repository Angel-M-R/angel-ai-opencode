## MODIFIED Requirements

### Requirement: Corrected intermediate failures use a shared evidence policy
For every implementation result and every strict-default control point that invokes the shared implementation-result policy, the orchestrator SHALL evaluate intermediate non-zero commands through one of two explicit evidence paths. It SHALL classify an implementation, verification, syntax, or invocation failure as corrected only when the same worker diagnoses and repairs its cause within the authorized scope, later runs an equivalent-or-broader relevant validation command with exit code zero, returns final status `done`, ends with a green final relevant validation state, and reports no deviation, scope expansion, or out-of-scope work. It SHALL classify a failed non-destructive probe as recovered without requiring a repair or equivalent rerun only when the command inspected state without mutation, failed solely because expected pre-operation state was absent or not yet established, and the same worker subsequently completes the authorized operation and runs authoritative final validation with exit code zero that proves the requested final state. The recovered-probe path MUST also end `done` with green final relevant validation and no deviation, scope expansion, or out-of-scope work. Applicable worker return contracts and orchestrator reporting SHALL preserve, in execution order, the failed command and exit code, its cause, any bounded correction or successful operation, the successful validation command and exit code, and evidence that the validation satisfies the applicable coverage rule. An eligible corrected failure or recovered probe SHALL be treated as clean for the control point's existing routing and SHALL NOT by itself cause an authorization or continuation question, mandatory stop, or archive delay. The orchestrator MUST surface the incident evidence and MUST NOT hide or relabel the failed command. These rules do not remove any route-local prerequisite or planned-task-specific safeguard.

#### Scenario: Worker corrects a real bounded failure
- **WHEN** the same worker diagnoses a real intermediate failure, repairs its cause within authorized scope, runs equivalent-or-broader relevant validation with exit code zero, returns `done` with a green final relevant state, and reports no deviation, scope expansion, or out-of-scope work
- **THEN** the orchestrator surfaces the failed command, correction, and successful validation and continues through the control point's existing clean-result route without requesting authorization solely for that incident

#### Scenario: Final verification corrects a real intermediate failure
- **WHEN** the final-verification worker satisfies every shared corrected-failure condition and all other verification and archive prerequisites are met
- **THEN** the orchestrator surfaces the complete incident evidence and continues to the existing archive path without stopping solely because of the corrected failure

#### Scenario: Tooling mistake remains eligible under the corrected-failure rule
- **WHEN** a syntax or invocation mistake satisfies every corrected-failure condition
- **THEN** the orchestrator classifies it through the corrected-failure evidence path and follows the applicable clean-result route

#### Scenario: Missing directory probe is recovered by successful creation
- **WHEN** `ls openspec` exits non-zero solely because `openspec/` does not exist before the requested operation, and the same worker then creates the authorized OpenSpec state, runs authoritative final status validation with exit code zero, returns `done` with a green final relevant state, and reports no deviation, scope expansion, or out-of-scope work
- **THEN** the orchestrator surfaces the failed probe, exit code, absence cause, successful operation, and satisfactory final validation, then continues automatically without asking a continuation question solely for the probe incident

#### Scenario: Another worker supplies the successful validation
- **WHEN** a failed command and its required successful validation come from different workers
- **THEN** the orchestrator treats the recovery evidence as insufficient and applies the shared mandatory-stop policy

#### Scenario: Green command does not satisfy the applicable coverage rule
- **WHEN** a corrected failure has only a later unrelated or narrower command, or a recovered probe has only a green command that does not authoritatively prove the requested final state
- **THEN** the orchestrator treats the incident as unresolved and applies the shared mandatory-stop policy

#### Scenario: Recovery evidence is incomplete
- **WHEN** the result omits the failed command or exit code, cause, required correction or successful operation, successful validation command or exit code, or evidence that the validation satisfies the applicable coverage rule
- **THEN** the orchestrator treats the result as non-clean and applies the shared mandatory-stop policy

#### Scenario: Recovered command accompanies an unsafe final result
- **WHEN** a worker reports later successful evidence but ends `partial` or `blocked`, has a red final relevant state, or reports a deviation, scope expansion, or out-of-scope work
- **THEN** the orchestrator applies the shared mandatory-stop policy

### Requirement: Unsafe final implementation results require a stop
Every strict-default control point that invokes the shared implementation-result policy MUST apply mandatory-stop handling when a non-zero command lacks an evidence-complete same-worker correction with equivalent-or-broader relevant validation exiting zero or an evidence-complete same-worker recovered non-destructive probe with authoritative final-state validation exiting zero, the final relevant evidence is red, status is `partial` or `blocked`, recovery evidence is insufficient, a destructive action occurs, or a deviation, scope expansion, or out-of-scope action is reported. An evidence-complete corrected intermediate failure or recovered non-destructive probe with a green final result is not unsafe and MUST NOT cause a stop by itself. A section-bounded planned OpenSpec task batch MAY instead defer only eligible local `partial`, local `blocked`, or red focused-test outcomes while its unfinished tasks remain unchecked; hard blockers and checked-task red-validation conflicts MUST stop immediately, and unresolved outcomes MUST stop after the single final retry round.

#### Scenario: Strict-default failure has no clean recovery evidence
- **WHEN** a strict-default worker reports a non-zero command without either an evidence-complete corrected-failure path or an evidence-complete recovered-probe path
- **THEN** the orchestrator surfaces the failed command and exit code and stops without deferral

#### Scenario: Strict-default final state remains red
- **WHEN** the final relevant state at a strict-default control point is red
- **THEN** the orchestrator reports the final red evidence and stops without deferral

#### Scenario: Strict-default worker returns incomplete status
- **WHEN** a strict-default worker returns `partial` or `blocked`
- **THEN** the orchestrator reports the incomplete result and applies the shared mandatory-stop policy

#### Scenario: Destructive action is not a recovered probe
- **WHEN** a worker executes a destructive action before or after a failed command
- **THEN** the orchestrator reports the action and applies the shared mandatory-stop policy regardless of later green evidence

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

## ADDED Requirements

### Requirement: Orchestrator policy propagation remains traceable
The repository orchestrator asset SHALL remain the authoritative policy source. The embedded asset used by a built executable MUST match the repository asset at build and test time, and installation MUST copy the selected orchestrator asset unchanged into the configured agent destination. Updating repository source alone MUST NOT be represented as updating an already installed copy; the documented update path SHALL require a binary or explicit asset source containing the new policy followed by reconciliation of the selected orchestrator asset.

#### Scenario: Embedded policy matches repository source
- **WHEN** the applicable embedded-versus-directory parity validation runs
- **THEN** the embedded orchestrator policy is byte-equivalent to the repository source

#### Scenario: Selected orchestrator is installed unchanged
- **WHEN** installation selects the orchestrator asset and targets a temporary configuration directory
- **THEN** the installed orchestrator policy is byte-equivalent to the selected source and contains the recovered-probe contract

#### Scenario: Existing installed copy predates source policy
- **WHEN** inspection finds that a configured orchestrator copy predates the repository policy
- **THEN** the discrepancy is explained by the build-and-reinstall propagation boundary rather than by treating the configured copy as an independent source of truth
