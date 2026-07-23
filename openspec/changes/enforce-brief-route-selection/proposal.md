## Why

The orchestrator's route gate must consistently turn a completed Brief into one auditable user decision before any route-specific action begins. Direct Fast also cannot safely satisfy Briefs that require executable validation because its mode contract forbids tests, builds, linting, and reproductions.

## What Changes

- Require the orchestrator to automatically issue exactly one single-select `question` immediately after presenting a new-work Brief; choosing an execution route implicitly confirms that Brief.
- Define the required route-selection payload, ordering, and re-prompt behavior as part of the orchestrator contract.
- Exclude Direct Fast when the Brief requires executable validation, and reject a custom Direct Fast response by reopening the same route-selection question.
- Align the execution-route-selection specification with the combined single-question and implicit-confirmation semantics.
- Refine the agent-system-prompt result policy so a corrected command-syntax or invocation error does not force a mandatory stop when the same worker later validates equivalent-or-broader relevant scope successfully, finishes `done`, and reports no deviation or out-of-scope work; retain the failed-command evidence and distinguish this tooling correction from a real verification or implementation failure.
- Extend orchestrator asset contract tests for automatic route-question emission, Direct Fast incompatibility, and the corrected-tooling-error boundary while retaining existing route contracts.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `execution-route-selection`: Require the automatic, single-question route gate after a Brief; make Direct Fast unavailable when executable validation is required; and clarify the mandatory-stop exception for corrected command-syntax or invocation errors.

## Impact

- Affects `assets/agents/angel-orchestrator.md` route-selection instructions and its `question` payload contract.
- Affects `assets/agents/angel-orchestrator.md` shared implementation-result policy in the agent-system-prompt contract.
- Affects `internal/install/agent_assets_test.go` contractual assertions for the orchestrator asset, including corrected-tooling-error evidence and failure classification.
- Does not change application behavior outside the orchestrator contract and its contractual tests.
