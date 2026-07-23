## MODIFIED Requirements

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
