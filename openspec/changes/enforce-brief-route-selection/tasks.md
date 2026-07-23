## 1. Orchestrator route-selection contract

- [x] 1.1 Update `assets/agents/angel-orchestrator.md` to require automatic emission of exactly one single-select route-selection `question` immediately after presenting a completed new-work Brief, with the orchestrator owning its payload and order.
- [x] 1.2 Preserve implicit Brief confirmation for valid offered routes and require Modify Brief or an incompatible custom route response to leave the Brief unconfirmed and reissue the route selection as specified.
- [x] 1.3 Make the route payload validation-aware: omit Direct Fast when the Brief requires tests, build, lint, or a reproduction; reject a custom Direct Fast response and reopen the same route-selection question.
- [x] 1.4 Update the agent-system-prompt shared implementation-result policy to classify a corrected command-syntax or invocation error as non-blocking only for the same worker's successful equivalent-or-broader relevant rerun, final `done`, and no deviation or out-of-scope work; require retained failed-command and exit-code evidence and preserve mandatory stops for real verification or implementation failures.

## 2. Contractual verification

- [x] 2.1 Update `internal/install/agent_assets_test.go` assertions to cover automatic post-Brief single-question emission, its ordered payload, and implicit confirmation without a separate confirmation or mode question.
- [x] 2.2 Add contractual assertions that validation-required Briefs omit Direct Fast and that a custom Direct Fast request is rejected and re-prompts the same route selection.
- [x] 2.3 Add contractual assertions for the corrected command-syntax/invocation-error exception, its complete evidence and same-worker conditions, and the distinction from real verification or implementation failures.
- [x] 2.4 Run the focused Go agent-asset contract tests and record their result.
