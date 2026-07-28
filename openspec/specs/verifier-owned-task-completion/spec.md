# Verifier-Owned Task Completion Specification

## Purpose
TBD

## Requirements

### Requirement: Planner marks at most one terminal verifier-owned section
When a change includes tasks that can be completed only by final verification, the planner SHALL place the exact marker `<!-- owner: openspec-verifier -->` as the first nonblank line of one named top-level task section, and that section MUST be the final task section. The marked section SHALL contain only final-verification tasks with independently reportable executed evidence. The planner MUST NOT emit more than one marker or place implementation work in the marked section.

#### Scenario: Planner creates a verifier-owned terminal section
- **WHEN** a plan includes final repository test, build, or scenario-evidence obligations that belong to the verifier
- **THEN** the planner places those obligations in one final section with the exact owner marker and keeps implementation tasks in earlier unmarked sections

#### Scenario: Plan has no verifier-owned tasks
- **WHEN** a plan does not represent final-verification obligations as tasks
- **THEN** the planner omits the owner marker and leaves the existing task workflow unchanged

#### Scenario: Change is planned before marker support is installed
- **WHEN** the change implementing this capability is itself planned under the current workflow
- **THEN** its own `tasks.md` remains unmarked and executable without the new marker behavior

### Requirement: Ownership detection is exact and structural
The workflow SHALL recognize a verifier-owned section only from the exact `<!-- owner: openspec-verifier -->` marker in the valid terminal-section position. It MUST NOT infer ownership from section titles, task wording, comments with different text, or legacy verification prose. Duplicate, misplaced, malformed, nested, or non-terminal owner markers MUST be treated as an invalid task-state conflict.

#### Scenario: Legacy section mentions verification
- **WHEN** an unmarked section title or task text refers to verification or the verifier
- **THEN** the workflow does not classify that section as verifier-owned

#### Scenario: Marker is valid
- **WHEN** exactly one marker appears as the first nonblank line in the final named top-level task section
- **THEN** the workflow classifies only that section as verifier-owned

#### Scenario: Marker layout is invalid
- **WHEN** markers are duplicated, malformed, misplaced, nested, or attached to a non-terminal section
- **THEN** the workflow reports a task-state conflict and changes no checkbox

### Requirement: Verifier-owned work is excluded from implementation batches
At every fresh planned-task decision, the orchestrator SHALL resolve and reread the active `tasks.md`, identify a valid marked section, and exclude that section from implementer batches. Earlier ordinary sections SHALL continue under the existing bounded implementation loop. When the marked section is the only section with pending tasks, the orchestrator SHALL automatically dispatch exactly one verifier rather than an implementer.

#### Scenario: Ordinary work remains before the marked section
- **WHEN** fresh task state contains pending ordinary tasks before a valid marked terminal section
- **THEN** the orchestrator dispatches only the next ordinary section to the implementer and does not include marked tasks

#### Scenario: Only verifier-owned work remains
- **WHEN** fresh task state shows every ordinary task complete and only tasks in the valid marked section pending
- **THEN** the orchestrator automatically dispatches one verifier for the active change

#### Scenario: Fresh state cannot be resolved
- **WHEN** status, the tasks artifact, or the valid marked-section state cannot be resolved at the dispatch decision
- **THEN** the orchestrator stops without dispatching an implementer or verifier and changes no checkbox

### Requirement: Context resolution supports local changes and explicit stores
Verifier-owned detection, dispatch, and completion SHALL use the active OpenSpec context. A local change SHALL be resolved by running status in its resolved project context, and an explicit store SHALL retain `--store <id>` on every applicable status operation and use `store:<id>` as its context identity. The workflow MUST NOT infer a local path for an explicit store or switch contexts during verification.

#### Scenario: Local change reaches verification
- **WHEN** only verifier-owned tasks remain in a repo-local change
- **THEN** the orchestrator and verifier resolve status and `tasks.md` from that local project context

#### Scenario: Explicit-store change reaches verification
- **WHEN** only verifier-owned tasks remain in a change selected from an explicit store
- **THEN** every applicable status operation retains the same store id and completion targets the tasks path returned for that store

#### Scenario: Context changes before completion
- **WHEN** fresh pre-write resolution returns a different context identity or resolved tasks path
- **THEN** completion reports a conflict and writes nothing

### Requirement: Verifier mutation authority is confined to resolved tasks
The verifier MUST remain unable to use general edit or write tools and MUST NOT modify product code. Its only permitted mutation SHALL be a guarded completion operation targeting the freshly resolved `tasks.md` of the active change and context. That operation MUST reject path overrides, symlink or traversal escapes, other changes or stores, content edits, checkbox changes outside the valid marked section, and any transition other than pending-to-complete for the entire target section.

#### Scenario: Verifier completes the authorized section
- **WHEN** the guarded operation receives the active context, freshly resolved tasks path, exact marked task set, complete evidence, and a global `pass`
- **THEN** it may change only those marked tasks from unchecked to checked

#### Scenario: Mutation targets product code
- **WHEN** a verifier action attempts to target code or any path other than the freshly resolved active `tasks.md`
- **THEN** the operation rejects the mutation and leaves every file unchanged

#### Scenario: Mutation includes an unrelated tasks edit
- **WHEN** the proposed replacement changes prose, ordering, whitespace outside required checkbox tokens, or a checkbox outside the marked section
- **THEN** the operation rejects the whole mutation and leaves `tasks.md` unchanged

### Requirement: Completion requires per-task evidence and a global pass
The verifier SHALL record concrete successful executed evidence for every task in the marked section and SHALL produce a global verdict. The guarded operation MUST complete the section only when the global verdict is `pass` and every marked task has task-specific evidence. A `fail` or `not-verified` verdict, failed evidence, missing or ambiguous task coverage, or an incomplete task set MUST result in no checkbox change.

#### Scenario: All verification evidence passes
- **WHEN** every marked task maps to applicable successful executed evidence and the global verdict is `pass`
- **THEN** the section is eligible for guarded atomic completion

#### Scenario: One task lacks evidence
- **WHEN** the global verdict is `pass` but at least one marked task lacks concrete successful evidence
- **THEN** no marked task is checked

#### Scenario: Verification does not pass
- **WHEN** the verifier returns `fail` or `not-verified`, or final relevant evidence is red or incomplete
- **THEN** no marked task is checked

### Requirement: Completion revalidates fresh state and fails on conflicts
The guarded operation SHALL capture the initial context identity, normalized resolved path, full `tasks.md` content digest, marker structure, and exact pending task identities. Immediately before writing, it MUST rerun status in the same context, reread the resolved file, and require every captured value and target task to remain unchanged. Any stale, missing, already changed, newly pending, differently resolved, or otherwise conflicting state MUST abort completion without modifying a checkbox.

#### Scenario: Tasks change during verification
- **WHEN** any content or checkbox in `tasks.md` differs between the captured snapshot and the fresh pre-write state
- **THEN** completion reports a conflict and writes nothing

#### Scenario: Resolved target changes during verification
- **WHEN** the active change, context identity, tasks artifact status, or resolved tasks path differs at pre-write resolution
- **THEN** completion reports a conflict and writes nothing

#### Scenario: Snapshot remains current
- **WHEN** fresh pre-write resolution exactly matches the captured context, path, content, marker, and target tasks
- **THEN** the operation may proceed to the evidence and atomic-commit checks

### Requirement: Marked-section completion is atomic
The guarded operation SHALL construct one replacement that changes every target checkbox and no other content, persist it through a sibling temporary file, and atomically replace the resolved `tasks.md`. If validation, temporary-file creation, persistence, replacement, or cleanup fails before commit, the original file MUST remain with all target tasks unchecked. The operation MUST NOT expose or retain a partially checked marked section.

#### Scenario: Atomic replacement succeeds
- **WHEN** every guard passes and the replacement commits successfully
- **THEN** all tasks in the marked section become checked together and no unrelated content changes

#### Scenario: Commit preparation fails
- **WHEN** any error occurs before the atomic replacement commits
- **THEN** the original `tasks.md` remains unchanged with no target task checked

#### Scenario: Partial task set is requested
- **WHEN** a completion request omits any pending task from the marked section
- **THEN** the operation rejects the request and checks none of the tasks

### Requirement: Verifier result reports completion independently from verdict
The verifier result SHALL report worker status, global verdict, per-task evidence, commands with exit codes, and completion state. A successful verification whose task commit conflicts MUST report `status: blocked`, `verdict: pass`, and a conflict completion state rather than claiming the change complete. Failed or incomplete verification SHALL report that no completion was attempted or performed.

#### Scenario: Verification passes and completion commits
- **WHEN** the global verdict is `pass` and guarded atomic completion succeeds
- **THEN** the verifier reports done status, pass verdict, task evidence, and successful completion

#### Scenario: Verification passes but state conflicts
- **WHEN** verification passes but fresh-state validation rejects the completion
- **THEN** the verifier reports blocked status, retains the pass evidence, identifies the conflict, and reports no checkbox changes
