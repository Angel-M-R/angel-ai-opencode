# openspec-worker-bootstrap Specification

## Purpose
TBD - created by archiving change manage-openspec-cli-bootstrap. Update Purpose after archive.
## Requirements
### Requirement: OpenSpec workers are gated by a general bootstrap
Before launching the first OpenSpec planner, implementer, or verifier for a project/store context in an orchestration session, the orchestrator SHALL run one short bootstrap through the `general` agent and wait for it to succeed. It SHALL skip repeated bootstrap for the same context in that session and SHALL run it again when the project or store context changes.

#### Scenario: First OpenSpec worker in a context
- **WHEN** an orchestration session is about to launch its first OpenSpec worker for a project or store
- **THEN** it completes the `general` bootstrap before dispatching that worker

#### Scenario: Later OpenSpec worker in the same context
- **WHEN** bootstrap already succeeded for the current project/store context in the same session
- **THEN** the orchestrator launches the OpenSpec worker without repeating bootstrap

#### Scenario: Context changes
- **WHEN** the selected project root or store differs from the bootstrapped context
- **THEN** the orchestrator runs bootstrap for the new context before launching an OpenSpec worker

### Requirement: OpenSpec list JSON determines readiness
The bootstrap SHALL use `openspec list --json` as the source of truth for local context readiness and SHALL pass `--store <id>` when an explicit registered store is selected. Conversational inference or filesystem presence alone MUST NOT mark a context ready.

#### Scenario: Local root resolves
- **WHEN** `openspec list --json` returns a resolvable local root
- **THEN** bootstrap accepts that root without running initialization

#### Scenario: Registered store is selected
- **WHEN** the OpenSpec action targets a registered store
- **THEN** bootstrap runs list with that store identifier and does not initialize an unrelated local root

### Requirement: Bootstrap initializes only an unresolved local root
When no store is explicit and `openspec list --json` cannot resolve a root, the bootstrap SHALL run `openspec init --tools none` and recheck readiness with `openspec list --json`. Initialization MUST NOT generate local skills.

#### Scenario: Local root is unresolved
- **WHEN** the initial list result has no resolvable root and no store is selected
- **THEN** bootstrap runs `openspec init --tools none` and verifies the resulting root through a second JSON list

#### Scenario: Initialization does not resolve a root
- **WHEN** initialization fails or the follow-up JSON list still has no resolvable root
- **THEN** bootstrap blocks the OpenSpec worker launch and reports the failure

### Requirement: Missing OpenSpec CLI blocks worker launch
If the `openspec` CLI cannot be executed, bootstrap SHALL block the OpenSpec worker and instruct the user to install it through the installer's `OpenSpec` extra.

#### Scenario: CLI is absent
- **WHEN** bootstrap cannot execute the OpenSpec CLI
- **THEN** no OpenSpec worker is launched and the user receives installer-extra guidance

### Requirement: Version drift is advisory
The bootstrap SHALL compare the OpenSpec CLI version with the `metadata.generatedBy` version of the child skills under the globally installed `openspec` bundle. If they differ, it SHALL emit a warning and continue when all other readiness checks pass.

#### Scenario: CLI and global skills match
- **WHEN** the CLI version equals the global skills' generated version
- **THEN** bootstrap continues without a drift warning

#### Scenario: CLI and global skills differ
- **WHEN** the CLI version differs from the global skills' generated version
- **THEN** bootstrap warns about the mismatch and still permits the OpenSpec worker launch

### Requirement: Duplicate local skills are tolerated silently
The bootstrap SHALL neither block nor warn when project-local OpenSpec skills duplicate globally installed skills, and SHALL NOT claim which copy OpenCode will select.

#### Scenario: Duplicate local OpenSpec skills exist
- **WHEN** readiness checks otherwise pass and duplicate project-local skills are present
- **THEN** bootstrap continues without a duplicate-skill message or precedence assertion

### Requirement: Bootstrap does not alter OpenSpec policy configuration
The bootstrap MUST NOT run `openspec update` and MUST NOT modify OpenSpec profile, workflow, or delivery configuration. Its only permitted initialization mutation is `openspec init --tools none` for an unresolved local root.

#### Scenario: Existing root is ready
- **WHEN** list JSON resolves the requested root or store
- **THEN** bootstrap performs no OpenSpec configuration mutation

#### Scenario: Initialization is required
- **WHEN** an unresolved local root requires initialization
- **THEN** bootstrap uses `--tools none` and performs no update, profile, workflow, delivery, or local-skill generation action

