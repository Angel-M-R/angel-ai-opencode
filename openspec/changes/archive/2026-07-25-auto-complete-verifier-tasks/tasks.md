## 1. Guarded Verifier Task Completion

- [x] 1.1 Add a strict parser and state model for one exact `<!-- owner: openspec-verifier -->` marker on the final named task section, rejecting duplicate, malformed, misplaced, nested, or non-terminal markers without inferring ownership from prose.
- [x] 1.2 Add active-change resolution and snapshot capture for repo-local and explicit-store contexts, retaining `--store <id>` and recording context identity, normalized resolved `tasks.md` path, full content digest, marker structure, and exact pending task identities.
- [x] 1.3 Implement the guarded completion operation so it accepts only a global `pass` with successful task-specific evidence for the complete marked task set, re-resolves fresh state immediately before writing, and returns a conflict without mutation on any snapshot, context, path, artifact, content, or task-set difference.
- [x] 1.4 Implement sibling-temporary-file persistence and atomic replacement that changes only all marked pending checkbox tokens together, preserves required metadata, rejects path escape or unrelated edits, and leaves the original unchanged on every pre-commit failure.
- [x] 1.5 Expose the snapshot and completion phases through a structured verifier-only command interface that supports stdin/JSON evidence without granting generic file editing, and return machine-readable status, verdict, per-task evidence, completion state, commands, and conflict diagnostics.
- [x] 1.6 Add focused unit tests for strict marker parsing, evidence completeness, local and explicit-store resolution, path confinement, stale and concurrent state, partial-set rejection, atomic success, and all-or-nothing failure behavior.

## 2. Planner and Verifier Contracts

- [x] 2.1 Update the planner contract to emit at most one exact marker as the first nonblank line of a terminal verification-only section, keep implementation and focused validation in ordinary sections, and never infer or retrofit ownership from legacy text.
- [x] 2.2 Update the verifier contract to capture fresh resolved state before verification, map successful executed evidence to each marked task, invoke only the guarded completion operation after a global `pass`, and report pass-with-conflict as blocked without marking tasks.
- [x] 2.3 Preserve the verifier's disabled generic edit/write tools and code read-only policy while narrowly allowing only the guarded resolved-`tasks.md` operation; explicitly prohibit other mutating shell commands and every product-code write.
- [x] 2.4 Extend focused agent-asset contract tests for planner marker placement, verifier permissions, evidence/result fields, no-write failure behavior, and installation of the updated assets unchanged.

## 3. Orchestrator Routing and Continuity

- [x] 3.1 Update fresh planned-task detection to exclude a valid marked terminal section from implementer batches, continue earlier ordinary sections normally, stop on invalid marker state, and automatically dispatch one verifier when only marked tasks remain.
- [x] 3.2 Propagate the exact repo-local root or explicit store id through verifier dispatch, snapshot, pre-write revalidation, and post-result status checks without local-path inference for stores.
- [x] 3.3 Update completion routing to accept only a clean done/pass/successful-completion result, refresh and confirm all tasks complete, reuse that pass at the existing review gate, and avoid a second verifier dispatch; retain mandatory stops for failure, incomplete evidence, conflict, or remaining tasks.
- [x] 3.4 Add focused orchestrator contract tests covering ordinary-before-marked scheduling, exact-marker-only detection, local/store command propagation, invalid and stale state stops, atomic-completion result handling, pass reuse, and no-verifier-redispatch continuity.

## 4. Focused Workflow Coverage

- [x] 4.1 Add temporary-fixture workflow tests that exercise successful local and explicit-store snapshot-to-completion round trips and prove that code, unmarked sections, unrelated task content, and other changes or stores remain untouched.
- [x] 4.2 Add failure-path workflow tests for failed or incomplete verification, missing per-task evidence, concurrent edits, changed resolution, commit preparation failure, and post-completion fresh-state disagreement, asserting zero partial checkbox updates and no review transition.
- [x] 4.3 Run only the focused tests applicable to the guarded operation and changed agent/orchestrator contracts during implementation, retaining command and exit-code evidence; do not run the full repository suite or build in an implementer batch because final verification owns them.
