## ADDED Requirements

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
