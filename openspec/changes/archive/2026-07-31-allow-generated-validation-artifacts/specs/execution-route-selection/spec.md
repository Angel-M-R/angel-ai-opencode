## ADDED Requirements

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
