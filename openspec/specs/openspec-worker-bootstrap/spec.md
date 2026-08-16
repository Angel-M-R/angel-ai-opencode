# openspec-worker-bootstrap Specification

## Purpose
TBD - created by archiving change manage-openspec-cli-bootstrap. Update Purpose after archive.
## Requirements
### Requirement: OpenSpec workers are gated by a general bootstrap
Before launching an OpenSpec planner, implementer, or verifier for a project/store context and required official skill in an orchestration session, the orchestrator SHALL run one short bootstrap through the `general` agent and wait for it to succeed. It SHALL skip repeated bootstrap only for the same context-and-skill pair in that session and SHALL run it again when the project, store, or required skill changes.

#### Scenario: First OpenSpec worker in a context
- **WHEN** an orchestration session is about to launch its first OpenSpec worker for a project or store
- **THEN** it completes the `general` bootstrap before dispatching that worker

#### Scenario: Later OpenSpec worker in the same context
- **WHEN** bootstrap already succeeded for the current project/store context and exact required skill in the same session
- **THEN** the orchestrator launches the OpenSpec worker without repeating bootstrap

#### Scenario: Context or required skill changes
- **WHEN** the selected project root, store, or required skill differs from the bootstrapped context-and-skill pair
- **THEN** the orchestrator runs bootstrap for the new pair before launching an OpenSpec worker

### Requirement: OpenSpec list JSON determines readiness
The bootstrap SHALL use `openspec list --json` as the source of truth for local context readiness and SHALL pass `--store <id>` when an explicit registered store is selected. Conversational inference or filesystem presence alone MUST NOT mark a context ready.

#### Scenario: Local root resolves
- **WHEN** `openspec list --json` returns a resolvable local root
- **THEN** bootstrap accepts that root without running initialization

#### Scenario: Registered store is selected
- **WHEN** the OpenSpec action targets a registered store
- **THEN** bootstrap runs list with that store identifier and does not initialize an unrelated local root

### Requirement: Bootstrap initializes an unresolved local root for OpenCode
When no store is explicit and `openspec list --json` cannot resolve a root, the bootstrap SHALL run `openspec init --tools opencode` and recheck readiness with `openspec list --json`. The official CLI SHALL own the generated project-local skills and commands.

#### Scenario: Local root is unresolved
- **WHEN** the initial list result has no resolvable root and no store is selected
- **THEN** bootstrap runs `openspec init --tools opencode` and verifies the resulting root through a second JSON list

#### Scenario: Initialization does not resolve a root
- **WHEN** initialization fails or the follow-up JSON list still has no resolvable root
- **THEN** bootstrap blocks the OpenSpec worker launch and reports the failure

### Requirement: Missing OpenSpec CLI blocks worker launch
If the `openspec` CLI cannot be executed, bootstrap SHALL block the OpenSpec worker and instruct the user to install it through the installer's `OpenSpec` extra.

#### Scenario: CLI is absent
- **WHEN** bootstrap cannot execute the OpenSpec CLI
- **THEN** no OpenSpec worker is launched and the user receives installer-extra guidance

### Requirement: Bootstrap refreshes existing project integration
When the first JSON list resolves an existing local root, the bootstrap SHALL run `openspec update` without `--force` and recheck readiness with `openspec list --json`. If update reports that no tools are configured, bootstrap SHALL run `openspec init --tools opencode` once to migrate a project previously initialized without tool integration, then perform the JSON readiness recheck.

#### Scenario: Existing OpenCode integration is configured
- **WHEN** the first JSON list resolves a local root whose OpenCode integration is configured
- **THEN** bootstrap runs `openspec update` and verifies readiness through a second JSON list

#### Scenario: Existing project has no configured tools
- **WHEN** update reports that the resolved local root has no configured tools
- **THEN** bootstrap runs `openspec init --tools opencode` once and verifies readiness through a second JSON list

#### Scenario: Project integration refresh fails
- **WHEN** update, migration initialization, or the follow-up JSON list fails
- **THEN** bootstrap blocks the OpenSpec worker launch and reports the failure

### Requirement: Bootstrap preserves OpenSpec policy configuration
The bootstrap MUST use the user's current global OpenSpec profile, workflow selection, and delivery mode. It MUST NOT run `openspec config`, pass `--profile`, or otherwise alter those settings. It MUST NOT pass `--force` to update.

#### Scenario: Existing root is ready
- **WHEN** list JSON resolves the requested root or store
- **THEN** bootstrap preserves profile, workflow, and delivery settings while refreshing any existing local integration

#### Scenario: Initialization is required
- **WHEN** an unresolved local root requires initialization
- **THEN** bootstrap uses `--tools opencode` without a profile override and lets OpenSpec generate the project integration from the current global policy

### Requirement: Requested workflow skill must be available
Before dispatching an OpenSpec worker, bootstrap SHALL receive the exact official skill requested by the route and verify that it was generated for OpenCode. The successful bootstrap cache SHALL be keyed by both context and required skill. If the user's profile or delivery mode excludes that skill, bootstrap MUST block rather than altering the user's OpenSpec configuration.

#### Scenario: Requested skill is available
- **WHEN** the project-local OpenCode integration contains the exact requested official skill
- **THEN** bootstrap may dispatch the worker after all other readiness checks pass

#### Scenario: Requested skill is excluded by user policy
- **WHEN** the current profile or delivery mode does not generate the requested skill
- **THEN** bootstrap blocks and tells the user to adjust the profile through `openspec config profile` without changing it automatically
