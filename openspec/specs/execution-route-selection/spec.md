## Purpose
Define how the orchestrator selects Direct or OpenSpec execution and safely
routes implementation, verification, review, and archive work.

## Requirements

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

### Requirement: Control diagnostics are a distinct shared result category
Every worker result and every control point that uses the shared implementation-result policy SHALL include a `control diagnostics` field in addition to `deviations`. The field SHALL contain structured records in observation order and SHALL be `none` when no record exists. Each record MUST identify the applicable control point, the read-only bookkeeping observation, whether the record is `complete` or `incomplete`, and the green final-state evidence supporting continuation. `control diagnostics` and `deviations` MUST be mutually exclusive at the result level, and a control diagnostic MUST NOT be reported, interpreted, or routed as a deviation.

#### Scenario: Complete bookkeeping diagnostic is visible
- **WHEN** a control point records an eligible read-only bookkeeping observation with complete descriptive details
- **THEN** the result reports the structured record as `complete` in observation order, reports `deviations: none`, and follows the control point's existing clean-result route

#### Scenario: No diagnostic exists
- **WHEN** a result contains no control diagnostic
- **THEN** the result reports `control diagnostics: none`

#### Scenario: Real deviation excludes diagnostics
- **WHEN** a result contains a functional deviation, scope expansion, or out-of-scope work
- **THEN** the result reports the issue under `deviations`, reports `control diagnostics: none`, and applies the existing mandatory-stop policy

### Requirement: Only benign read-only bookkeeping is non-blocking
The orchestrator SHALL classify a record as a non-blocking control diagnostic only when it is internal read-only bookkeeping, no command failed, no mutation or destructive action occurred, no file was touched, the result status is `done`, and final relevant validation is green. An eligible diagnostic SHALL remain visible but MUST NOT cause an authorization question, mandatory stop, archive delay, or deviation classification by itself. When descriptive diagnostic details are incomplete but the result still establishes every eligibility fact and a green final relevant state, the record SHALL be labeled `incomplete` and the orchestrator SHALL continue through the existing clean-result route. This classification MUST NOT alter any workflow prerequisite or the existing corrected-failure, recovered-probe, generated-validation-artifact, planned-task, or route-local policy.

#### Scenario: Read-only bookkeeping is non-blocking at any control point
- **WHEN** any control point reports only internal read-only bookkeeping, no failed command, mutation, destructive action, or touched file, final status `done`, and green final relevant validation
- **THEN** the orchestrator surfaces the diagnostic and continues through that control point's existing clean-result route

#### Scenario: Incomplete descriptive record remains continuable
- **WHEN** an otherwise eligible control diagnostic has incomplete descriptive bookkeeping details while every classification fact is established and final relevant validation is green
- **THEN** the orchestrator labels the record `incomplete`, surfaces it, and continues without treating it as a deviation

#### Scenario: Failed command cannot be a control diagnostic
- **WHEN** any command exits non-zero
- **THEN** the orchestrator does not use the control-diagnostic classification and applies the existing corrected-failure, recovered-probe, or mandatory-stop policy as applicable

#### Scenario: Mutation or touched file cannot be a control diagnostic
- **WHEN** bookkeeping mutates state, performs a destructive action, or touches a file
- **THEN** the orchestrator does not use the control-diagnostic classification and retains the existing scope, deviation, destructive-action, and mandatory-stop handling

#### Scenario: Unsafe final result still stops
- **WHEN** a result is `partial` or `blocked`, final relevant validation is red, or functional deviation, functional or scope expansion, out-of-scope work, or another existing mandatory-stop condition is present
- **THEN** the orchestrator reports the blocking evidence and applies the existing mandatory-stop route without suppression by any bookkeeping record

#### Scenario: Corrected failure and recovered probe rules are unchanged
- **WHEN** an incident is evaluated as a corrected intermediate failure or recovered non-destructive probe
- **THEN** the orchestrator applies the existing evidence and routing rules for that incident rather than reclassifying it as a control diagnostic

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

### Requirement: Final verification dispatch accounts for marked terminal tasks
The orchestrator SHALL treat a valid terminal section marked `<!-- owner: openspec-verifier -->` as final-verification work rather than planned implementation. When fresh state shows that section is the only pending work, it SHALL dispatch the verifier automatically in the active local or explicit-store context. It MUST NOT dispatch the verifier from unmarked legacy text or send marked tasks to an implementer.

#### Scenario: Marked terminal section is the only pending work
- **WHEN** fresh active-change state shows all ordinary tasks checked and only a valid marked terminal section pending
- **THEN** the orchestrator dispatches the verifier automatically with the same change and context identity

#### Scenario: Unmarked verification prose remains pending
- **WHEN** only an unmarked section whose title or tasks mention verification remains pending
- **THEN** the orchestrator follows the ordinary planned-task workflow and does not infer verifier ownership

#### Scenario: Explicit store is active
- **WHEN** a marked terminal section is the only pending work for a change selected with an explicit store id
- **THEN** the verifier dispatch and all applicable state checks retain that store id

### Requirement: One successful verification continues directly to review
After the verifier returns a clean `status: done`, global `verdict: pass`, task-specific evidence for every marked task, and successful atomic completion, the orchestrator SHALL refresh the same active context and confirm that the tasks artifact is complete with no pending checkbox. It SHALL then reuse that verifier pass as final-verification evidence and proceed to the existing review gate without dispatching a second verification. Any non-pass, incomplete, conflicting, non-clean, or still-pending state MUST stop before review.

#### Scenario: Pass closes the marked section
- **WHEN** guarded completion succeeds after a clean verifier pass and fresh follow-up state confirms all tasks complete
- **THEN** the orchestrator enters the existing review gate using that pass without a second verifier dispatch

#### Scenario: Pass encounters a completion conflict
- **WHEN** verification evidence passes but atomic completion reports stale or concurrent state
- **THEN** the orchestrator retains the evidence, stops before review, and does not dispatch another worker automatically

#### Scenario: Follow-up state remains incomplete
- **WHEN** the verifier reports successful completion but fresh follow-up status or `tasks.md` still contains pending work
- **THEN** the orchestrator treats the state as conflicting and stops before review without reusing the pass

#### Scenario: Verification fails or is incomplete
- **WHEN** the verifier returns `fail`, `not-verified`, non-clean status, failed task evidence, or incomplete task coverage
- **THEN** the orchestrator changes no checkbox through fallback behavior and applies the existing mandatory-stop route before review

### Requirement: Generated validation artifacts use causal classification
Every Direct Safe result, section-bounded OpenSpec implementation result, bounded Direct or OpenSpec fix result, and final OpenSpec verification result SHALL classify a created or modified regenerable artifact as a generated validation artifact when an authorized validation command exits zero and command-specific evidence attributes the artifact diff to that command without intervening manual mutation or manual source-code edits. Eligibility SHALL NOT depend on whether Git ignores or tracks the artifact. The result MUST retain the artifact, list it in a separate generated-validation-artifacts category with the producing command, exit code, and causal-diff evidence, and MUST NOT report the eligible artifact as a deviation, scope expansion, or out-of-scope work or request authorization solely because it exists. This classification MUST NOT expand the functional scope of the assigned Brief, change, task, or fix batch and MUST NOT alter corrected-failure, recovered-probe, final-validation, destructive-action, or mandatory-stop rules.

#### Scenario: Next.js output is causally generated
- **WHEN** an authorized successful validation command creates or updates `.next/`, and its attributable diff contains only regenerable output with no manual source-code edit
- **THEN** the worker retains `.next/` and reports it in the separate generated-validation-artifacts category without treating it as a deviation or requesting authorization

#### Scenario: CodeGraph output is visible to Git
- **WHEN** an authorized successful validation command creates or updates `.codegraph/` with complete causal evidence and Git does not ignore some or all of that output
- **THEN** Git visibility does not change eligibility, and the worker retains and reports the output as generated validation artifacts

#### Scenario: TypeScript build metadata is tracked
- **WHEN** an authorized successful validation command updates one or more `*.tsbuildinfo` files and the command-specific diff proves that no manual source-code edit is included
- **THEN** the worker retains and separately reports those files as generated validation artifacts regardless of tracking status

#### Scenario: Classification is route-neutral
- **WHEN** identical evidence-complete generated output occurs in Direct Safe work, a planned OpenSpec batch, a bounded fix batch, or final OpenSpec verification
- **THEN** the applicable worker and orchestrator apply the same generated-validation-artifact classification and preserve every route-local prerequisite

#### Scenario: Generated output is retained
- **WHEN** an artifact qualifies as generated validation output
- **THEN** no worker or orchestrator automatically cleans, reverts, deletes, stages, or commits it

#### Scenario: Producing command fails
- **WHEN** the command associated with an output exits non-zero or final relevant validation remains red
- **THEN** the output does not make the result clean, and the command and final state remain governed by the existing corrected-failure, recovered-probe, and mandatory-stop policies

#### Scenario: Attribution is missing or ambiguous
- **WHEN** a worker cannot show an authorized zero-exit validation command and an attributable artifact diff free of intervening manual mutation
- **THEN** the worker MUST NOT use the generated-validation-artifact classification and the orchestrator applies the existing strict scope and mandatory-stop policy

#### Scenario: Attributable diff includes a manual source edit
- **WHEN** the claimed generated-output diff contains a manual source-code edit, functional expansion, or other out-of-scope work
- **THEN** the generated-output classification does not authorize or conceal that change, and the existing deviation and scope-violation policy applies

#### Scenario: Destructive action targets generated output
- **WHEN** a worker runs a destructive cleanup or other destructive action before or after producing a generated artifact
- **THEN** later green validation or otherwise valid artifact causality does not suppress the existing mandatory stop

### Requirement: Workspace audits omit only unchanged pre-existing top-level dotpaths
Every route that audits the workspace SHALL apply one shared worker-boundary filter before classifying or reporting paths. A path MUST be silently omitted only when the first component of its workspace-relative path begins with `.`, reliable worker-start evidence proves that the path was already modified, and worker-end evidence proves that its state is identical to that baseline. An omitted path MUST NOT appear in the result, including files touched, generated-validation-artifacts, deviations, scope expansion, or out-of-scope work. The filter MUST NOT authorize creating or modifying any dotpath, and a path whose first component does not begin with `.`, whose baseline is absent or unreliable, or whose state changes during the worker MUST remain subject to all normal generated-output, corrected-failure, deviation, scope, and mandatory-stop rules.

#### Scenario: Pre-existing VS Code settings remain unchanged
- **WHEN** `.vscode/settings.json` was already modified in a reliable worker-start baseline and remains identical at worker completion
- **THEN** every applicable workspace-auditing route silently omits it from the result without treating the omission as permission to create or modify VS Code configuration
