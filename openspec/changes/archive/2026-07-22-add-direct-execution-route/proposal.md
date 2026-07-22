## Why

The orchestrator currently sends non-trivial implementation work through OpenSpec even when a confirmed Brief describes a clear, isolated, reversible change. Adding an explicit route choice preserves the full planning workflow while allowing users to choose bounded direct execution with transparent verification tradeoffs.

## What Changes

- After a Brief is confirmed and before a new OpenSpec change is created, ask the user to choose OpenSpec or direct execution and provide a non-binding recommendation based on risk, scope, and uncertainty.
- Add Safe and Fast direct modes, both delegated to a bounded `general` worker with the confirmed Brief and a strict result contract.
- In Safe mode, require the implementation worker to run existing tests/builds and stop on unavailable or finally red verification, `partial` or `blocked` status, Brief deviation, or out-of-scope work. An intermediate non-zero command is corrected rather than stopping when the same worker identifies it, later runs an equivalent or broader relevant command successfully, returns `done`, and reports no deviation or out-of-scope work; retain and surface both command results before offering the existing review choices.
- Route selected direct-mode review fixes through one bounded `general` batch and permit rerunning only affected reviewers.
- In Fast mode, implement without tests or reviews and explicitly report the result as unverified.
- Keep existing OpenSpec changes on the OpenSpec workflow and preserve the current bootstrap, planning, implementation cadence, verification, review, and archive behavior, except that planned implementation batches use the same corrected-intermediate-failure exception and may continue under the selected cadence after surfacing the retained evidence.
- Keep mandatory stopping for uncorrected failures, final red state, `partial` or `blocked` results, Brief or plan deviation, out-of-scope work, and TDD or expected failures that remain red at batch end.
- Add contract tests for the orchestrator asset covering route-gate ordering, direct worker and prompt contracts, shared corrected-failure and mandatory-stop semantics, review routing, existing-change routing, and preservation of all other OpenSpec behavior.

## Capabilities

### New Capabilities

- `execution-route-selection`: Defines route selection after Brief confirmation, bounded Safe and Fast direct execution, shared corrected-intermediate-failure handling for Direct Safe workers and planned OpenSpec implementation batches, stop and review behavior, existing-change routing, and compatibility with the complete OpenSpec workflow.

### Modified Capabilities

None.

## Impact

- Affects the repository-managed orchestrator agent asset and its textual contract-test coverage.
- Adds no product API, persisted data, migration, dependency, bootstrap, OpenSpec CLI, or OpenSpec artifact-generation behavior to the direct route.
- Leaves the OpenSpec worker assets and all unrelated workflow contracts intact while refining orchestrator handling of implementation-batch command evidence.
