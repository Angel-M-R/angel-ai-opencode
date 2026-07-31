## MODIFIED Requirements

### Requirement: Planned batch validation preserves verifier ownership
The planned-task implementer SHALL run validation relevant to its bounded changes and MAY run focused tests that exercise modified code. It MUST NOT run the full repository test suite or builds, which remain reserved for the final OpenSpec verifier. Additional reads and successful focused tests SHALL be retained as benign, continuable deviations when they serve only the bounded batch. Regenerable artifacts produced by an authorized successful focused validation command SHALL instead be retained and reported in the separate generated-validation-artifacts category when command-specific evidence attributes their diff to that command without manual source-code edits; eligible generated artifacts MUST NOT be classified as benign deviations or as functional expansion. Before classifying workspace paths, the implementer MUST silently omit only a path whose first workspace-relative component begins with `.`, that reliable worker-start evidence shows was already modified, and whose worker-end state is identical to that baseline. The omission MUST NOT authorize a dotpath write or hide a path with an absent or unreliable baseline or any worker-time change.

#### Scenario: Implementer runs a focused test
- **WHEN** a planned-task implementer runs a test targeted to code modified in its bounded batch and the test succeeds
- **THEN** the orchestrator records the command as a benign deviation and may continue

#### Scenario: Implementer performs an additional relevant read
- **WHEN** the implementer reads additional repository context needed for the bounded batch without writing outside it
- **THEN** the orchestrator records the read deviation and may continue

#### Scenario: Focused validation creates an eligible artifact
- **WHEN** an authorized successful focused validation command produces a causally attributable regenerable artifact without manual source-code edits
- **THEN** the implementer retains and separately reports the artifact without treating it as a deviation or widening the planned batch

#### Scenario: Implementer attempts full verification
- **WHEN** a planned-task implementer runs a full repository suite or build
- **THEN** the orchestrator treats the command as a blocking policy violation rather than final verification, regardless of whether it produces regenerable artifacts

### Requirement: Hard blockers cannot be deferred
The planned-task loop MUST stop immediately when a worker writes outside the bounded batch, expands functional scope, runs a destructive command, OpenSpec state cannot be resolved safely, a checked task has relevant red validation, or another result leaves unsafe final evidence that is not eligible for local deferral. A causally proven regenerable artifact produced by an authorized successful focused validation command is not an out-of-batch write or functional expansion, but this classification MUST NOT authorize any manual source edit, unrelated output, failed command, red final evidence, or destructive cleanup. A pre-existing path omitted by the shared top-level-dotpath filter is not a worker write and MUST NOT enter the result; without reliable proof that it remained identical, the normal hard-blocker classification applies. The loop MUST NOT ignore red tasks or alter checkbox state to manufacture completion.

#### Scenario: Worker writes outside the batch
- **WHEN** worker evidence shows a manual write outside the exact planned batch that is not an eligible generated validation artifact
- **THEN** the orchestrator stops and reports the scope violation without deferring it

#### Scenario: Worker expands functionality
- **WHEN** a planned-task worker implements behavior not required by its bounded tasks
- **THEN** the orchestrator stops and reports the functional expansion

#### Scenario: Eligible artifact remains within classification scope
- **WHEN** focused validation for the bounded batch succeeds and produces an evidence-complete regenerable artifact without manual source edits
- **THEN** the orchestrator retains and separately reports the artifact and does not treat its path alone as an out-of-batch write

#### Scenario: Destructive command is reported
- **WHEN** a planned-task result contains a destructive command, including cleanup of generated artifacts
- **THEN** the orchestrator stops immediately and preserves the command evidence

#### Scenario: Checked task remains red
- **WHEN** fresh `tasks.md` marks a task complete while relevant retained validation remains red
- **THEN** the orchestrator reports a real `tasks.md` state conflict and stops without unchecking, ignoring, or relabeling the task

#### Scenario: OpenSpec state is unresolvable
- **WHEN** fresh status or the resolved `tasks.md` cannot establish current task state
- **THEN** the orchestrator stops the planned-task cycle and does not use cached state for deferral or retry

#### Scenario: Claimed generated output has ambiguous provenance
- **WHEN** an artifact lacks command-specific zero-exit causality, includes an intervening manual mutation, or its attributable diff contains a manual source-code edit
- **THEN** the planned-task loop does not grant the generated-output exception and applies the existing hard-blocker classification
