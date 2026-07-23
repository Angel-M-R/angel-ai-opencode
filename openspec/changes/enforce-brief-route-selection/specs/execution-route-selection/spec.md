## MODIFIED Requirements

### Requirement: Route selection follows Brief confirmation
For non-trivial work that does not target an existing OpenSpec change, the orchestrator SHALL present the completed Brief and immediately invoke exactly one single-select `question` tool call that asks the user to select the execution route before it performs OpenSpec bootstrap, invokes an OpenSpec worker, runs an OpenSpec CLI command for the new work, or creates an OpenSpec change or artifact. The orchestrator SHALL own the route-selection payload and SHALL keep the `question` tool's custom response available. Selecting a valid offered execution route SHALL implicitly confirm the presented Brief; the orchestrator MUST NOT ask a separate Brief-confirmation, route, or Direct-mode question.

#### Scenario: New work automatically reaches the route gate in order
- **WHEN** a non-trivial new-work interview produces a completed Brief
- **THEN** the orchestrator presents the Brief and immediately emits its one single-select route-selection `question` before causing any OpenSpec side effect or delegating route-specific work

#### Scenario: Valid route selection implicitly confirms the Brief
- **WHEN** the user selects an offered valid route from the post-Brief route-selection question
- **THEN** the orchestrator treats the Brief as confirmed and enters only the selected route without asking an additional confirmation or Direct-mode question

#### Scenario: Unconfirmed work does not reach route selection
- **WHEN** a non-trivial request has not yet produced a completed Brief
- **THEN** the orchestrator completes the existing interview gate before presenting the Brief and emitting the route-selection question

### Requirement: Recommendation and route payload are risk- and validation-aware
The orchestrator SHALL derive a risk-based recommendation from the Brief and construct the single-select route-selection payload in the stated order. For a Brief that does not require executable validation, a clear, isolated, reversible change SHALL offer `Direct Safe (Recommended)`, `Direct Fast`, `OpenSpec`, and `Modify Brief`; a Brief affecting architecture, security, data, migrations, cross-cutting scope, or material uncertainty SHALL offer `OpenSpec (Recommended)`, `Direct Safe`, `Direct Fast`, and `Modify Brief`. The recommendation SHALL be non-binding, and the orchestrator MUST NOT recommend Direct Fast by default.

When the Brief requires executable validation, including tests, build, lint, or a reproduction, Direct Fast SHALL be incompatible. The orchestrator SHALL omit Direct Fast from the payload while preserving the applicable risk-based ordering among `Direct Safe`, `OpenSpec`, and `Modify Brief`. If the user requests Direct Fast through a custom response in this state, the orchestrator MUST reject that request without confirming the Brief and MUST reissue the same single-select route-selection question with Direct Fast omitted.

#### Scenario: Validation-free low-risk Brief offers Direct Fast
- **WHEN** a completed Brief is clear, isolated, reversible, and does not require executable validation
- **THEN** the orchestrator emits the route payload in this order: `Direct Safe (Recommended)`, `Direct Fast`, `OpenSpec`, `Modify Brief`

#### Scenario: Validation-free high-risk Brief offers Direct Fast
- **WHEN** a completed Brief affects architecture, security, data, migrations, cross-cutting scope, or has material uncertainty and does not require executable validation
- **THEN** the orchestrator emits the route payload in this order: `OpenSpec (Recommended)`, `Direct Safe`, `Direct Fast`, `Modify Brief`

#### Scenario: Validation-required Brief omits Direct Fast
- **WHEN** a completed Brief requires tests, a build, lint, or a reproduction as executable validation
- **THEN** the orchestrator emits the applicable risk-ordered payload containing only `Direct Safe`, `OpenSpec`, and `Modify Brief` and does not offer Direct Fast

#### Scenario: Custom Direct Fast request is incompatible with validation-required Brief
- **WHEN** a Brief requires executable validation and the user supplies Direct Fast through the route question's custom response
- **THEN** the orchestrator rejects Direct Fast, leaves the Brief unconfirmed, and reissues the same route-selection question without Direct Fast

#### Scenario: User modifies the Brief
- **WHEN** the user selects Modify Brief
- **THEN** the orchestrator leaves the Brief unconfirmed, reopens the interview, updates the Brief from the user's answers, reassesses risk and executable-validation requirements, and emits a new route-selection question

### Requirement: Corrected tooling errors preserve evidence without becoming verification success
The agent-system-prompt contract SHALL classify a non-zero command caused by command syntax or invocation as a corrected tooling error, rather than a mandatory-stop condition, only when the same worker identifies the error, later runs an equivalent-or-broader relevant command that successfully covers the failed command's relevant scope, returns final status `done`, and reports no deviation or out-of-scope work. The worker SHALL retain and report the original failed command and its exit code together with the later successful command and exit code.

The agent-system-prompt contract MUST distinguish a corrected tooling error from a real verification or implementation failure. A red final relevant verification state, failed or incomplete implementation, an unrelated or narrower rerun, a rerun by another worker, a non-`done` worker result, a deviation, or out-of-scope work SHALL remain subject to the mandatory-stop policy.

#### Scenario: Same worker corrects a command invocation error
- **WHEN** a worker identifies that a non-zero command failed because of command syntax or invocation and later runs an equivalent-or-broader relevant command successfully
- **THEN** the orchestrator permits the work to continue without a mandatory stop only if that worker ends `done`, reports no deviation or out-of-scope work, and retains both commands with their exit codes in the evidence

#### Scenario: A corrected command does not conceal a verification or implementation failure
- **WHEN** a worker has a real verification or implementation failure, or any condition for the corrected-tooling-error exception is not met
- **THEN** the orchestrator retains the failed-command evidence and applies the mandatory-stop policy
