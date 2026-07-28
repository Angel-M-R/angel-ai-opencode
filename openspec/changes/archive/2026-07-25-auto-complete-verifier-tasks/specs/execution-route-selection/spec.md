## ADDED Requirements

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
