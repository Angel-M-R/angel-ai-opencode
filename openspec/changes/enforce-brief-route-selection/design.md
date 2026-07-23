## Context

The orchestrator currently describes a combined Brief, execution-route, and Direct-mode selection plus a shared implementation-result policy in `assets/agents/angel-orchestrator.md`. Its contract tests in `internal/install/agent_assets_test.go` assert text ordering, route boundaries, and result-policy clauses. The confirmed Brief requires the agent-system-prompt contract to explicitly guarantee automatic single-question emission after the Brief, prevent Direct Fast from being selected when the Brief requires executable validation, and classify a corrected command-syntax or invocation error separately from a real verification or implementation failure.

## Goals / Non-Goals

**Goals:**
- Make the post-Brief route selection an explicit, single-select `question` contract owned by the orchestrator.
- Preserve implicit Brief confirmation for valid route selections.
- Prevent Direct Fast from bypassing required tests, build, lint, or reproduction validation.
- Allow a corrected command-syntax or invocation error to continue without a mandatory stop only under the confirmed same-worker, successful-equivalent-or-broader-rerun, `done`, and no-deviation conditions.
- Preserve failed-command and exit-code evidence while ensuring that the correction cannot conceal a real verification or implementation failure.
- Verify the new rules through focused agent-asset contract tests.

**Non-Goals:**
- Implement or alter application behavior outside the orchestrator instruction asset and its contract tests.
- Change Direct Safe, OpenSpec lifecycle, worker selection, or validation policies beyond the Direct Fast eligibility rule.
- Add a separate Brief-confirmation or Direct-mode question.
- Treat a red final verification result, an implementation failure, an unresolved command failure, or a corrected command outside the stated evidence conditions as continuable.

## Decisions

### Define one automatic route-selection payload in the orchestrator
Immediately after presenting a new-work Brief, the orchestrator will call exactly one single-select `question` tool invocation. The contract will state its sequence, that it retains custom responses, the ordered routes it may contain, and that a valid selection confirms the Brief implicitly.

This makes the source of the payload and the order of interaction unambiguous. A separate confirmation question was rejected because it contradicts the confirmed one-question interaction model.

### Derive Direct Fast eligibility from Brief validation obligations
Before constructing route options, the orchestrator will determine whether the Brief requires executable validation: tests, a build, lint, or a reproduction. When it does, the route payload will omit Direct Fast. If the user supplies Direct Fast through the question's custom response, the orchestrator will reject it without confirming the Brief and reissue the same route-selection question.

Omitting only the displayed option is insufficient because custom responses remain available. Allowing Fast with a later warning was rejected because Fast explicitly forbids executable validation and would violate the confirmed Brief.

### Keep contracts executable through asset tests
Focused tests will assert the automatic post-Brief question, the single-select payload/order and implicit confirmation, Fast omission for validation-required Briefs, rejection/re-prompt behavior for a custom Fast request, and the corrected-tooling-error exception. The latter assertions will require the agent-system prompt to retain the original failed command and exit code, require a later equivalent-or-broader relevant successful command by the same worker, require final `done` and no deviation or out-of-scope work, and state that a real verification or implementation failure remains a mandatory-stop condition. Existing assertions for risk ordering, route boundaries, and Direct isolation remain and will be adjusted only where their expected route lists change.

### Classify corrected tooling errors without weakening verification
A non-zero command caused by command syntax or invocation is a corrected tooling error only when the same worker identifies it, later executes an equivalent-or-broader command covering the failed command's relevant scope with exit code zero, returns `done`, and reports neither deviation nor out-of-scope work. The prompt contract will require reporting both the original failed command and its exit code alongside the successful rerun.

This exception is limited to correcting the tooling error; it does not convert a real verification or implementation failure into success. A red final relevant verification state, incomplete or failed implementation, absent or narrower/unrelated successful rerun, another worker's rerun, non-`done` status, deviation, or out-of-scope work continues to require the existing mandatory stop.

## Risks / Trade-offs

- [Textual asset tests can pass while instructions are semantically ambiguous] → Assert ordered, normative phrases for all new gate conditions and preserve the end-to-end ordering test.
- [Custom answers could silently circumvent the displayed route options] → Require explicit rejection and re-prompting of Direct Fast whenever executable validation is required.
- [Route choice lists become conditional] → Specify the exact alternatives and ordering for validation-required and validation-free Briefs.
- [A corrected invocation can mask a substantive failure] → Require retained failed-command/exit-code evidence and explicit contract-test assertions that real verification and implementation failures still stop.
