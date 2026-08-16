# openspec-worker-bootstrap Specification

## Purpose
TBD - created by archiving change manage-openspec-cli-bootstrap. Update Purpose after archive.
## Requirements
### Requirement: OpenSpec workers are gated by a general bootstrap
Before launching an OpenSpec planner, implementer, or verifier for a project/store context in an orchestration session, the orchestrator SHALL run one short bootstrap through the `general` agent and wait for it to succeed. It SHALL skip repeated bootstrap for the same local root or explicit store/tool-host pair in that session and SHALL run it again when the local project, store, or store tool host changes.

#### Scenario: First OpenSpec worker in a context
- **WHEN** an orchestration session is about to launch its first OpenSpec worker for a project or store
- **THEN** it completes the `general` bootstrap before dispatching that worker

#### Scenario: Later OpenSpec worker in the same context
- **WHEN** bootstrap already succeeded for the current project/store context in the same session
- **THEN** the orchestrator launches the OpenSpec worker without repeating bootstrap

#### Scenario: Context changes
- **WHEN** the selected project root or store differs from the bootstrapped context
- **THEN** the orchestrator runs bootstrap for the new context before launching an OpenSpec worker

#### Scenario: Same store is used from a different tool host
- **WHEN** an explicit store was bootstrapped from one working project and a worker later targets it from another working project
- **THEN** the orchestrator bootstraps the second tool host before launching that worker

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

### Requirement: Store planning and OpenCode integration use separate roots
For an explicit store, bootstrap SHALL treat the registered store as the planning context and the working project as the OpenCode tool host. It SHALL initialize or update the official OpenCode integration in the tool host, SHALL verify skills there, and SHALL retain the planning paths resolved by store-scoped JSON commands. It MUST NOT initialize or update the store root merely to provision the working project's agent skills.

Bootstrap SHALL determine tool-host initialization by running `openspec context --json` without `--store` in the working directory. A local root whose source is `nearest` SHALL be treated as initialized. The exact `no_openspec_root` or `no_root_with_registered_stores` diagnostic, or a context whose source is a declared or global store, SHALL be treated as an uninitialized local tool host; any other failure SHALL block.

#### Scenario: Clean tool host targets an explicit store
- **WHEN** a registered store resolves successfully and the working project has no local OpenSpec initialization
- **THEN** bootstrap runs `openspec init --tools opencode` in the tool host and rechecks the same store-scoped JSON readiness

#### Scenario: Initialized tool host targets an explicit store
- **WHEN** a registered store resolves successfully and the working project already has OpenSpec initialized
- **THEN** bootstrap runs `openspec update` in the tool host and rechecks the same store-scoped JSON readiness

#### Scenario: Store worker receives planning paths
- **WHEN** bootstrap succeeds for an explicit store
- **THEN** the worker receives the exact store id and resolved planning paths and writes only within those paths

### Requirement: Bootstrap enforces the official core profile
Before local initialization or update, the bootstrap SHALL run `openspec config profile core` followed by `openspec config set delivery both`. Failure of either command MUST block. Bootstrap MUST NOT select, preserve, or generate a custom workflow profile and MUST NOT pass `--force` to update.

#### Scenario: Custom profile existed previously
- **WHEN** global OpenSpec configuration names a custom profile or workflow list
- **THEN** bootstrap replaces it with the official `core` profile and `both` delivery before generating project integration

#### Scenario: Initialization is required
- **WHEN** an unresolved local root requires initialization
- **THEN** bootstrap configures `core` with `both`, uses `--tools opencode`, and lets OpenSpec generate the official core project integration

#### Scenario: Existing project requires update
- **WHEN** list JSON resolves an initialized local project
- **THEN** bootstrap configures `core` with `both` before running `openspec update` without `--force`

### Requirement: Complete core workflow integration must be available
Before dispatching any OpenSpec worker, bootstrap SHALL verify that all six official core skills were generated for OpenCode: `openspec-propose`, `openspec-explore`, `openspec-apply-change`, `openspec-update-change`, `openspec-sync-specs`, and `openspec-archive-change`. A missing core skill MUST block as an incomplete or corrupt OpenCode integration; bootstrap MUST NOT select a custom workflow to obtain it.

#### Scenario: Complete core integration is available
- **WHEN** the project-local OpenCode integration contains all six official core skills
- **THEN** bootstrap may dispatch the worker after all other readiness checks pass

#### Scenario: A core skill is missing
- **WHEN** the project-local OpenCode integration lacks any official core skill after initialization or update
- **THEN** bootstrap blocks and reports an incomplete or corrupt integration without switching to a custom profile

### Requirement: Angel verification is core-compatible
The `openspec-verifier` SHALL NOT load or require a non-core OpenSpec verification skill. It SHALL use `openspec status --change <name> --json` and `openspec instructions apply --change <name> --json` in the bootstrapped context to resolve and read the verification contract, then apply Angel AI's executed-evidence verification policy.

#### Scenario: Core workflow reaches final verification
- **WHEN** all planned implementation tasks are complete under the official core workflow
- **THEN** the Angel verifier resolves the change through the official CLI and evaluates completeness, correctness, coherence, and scenario evidence without loading a custom workflow skill

### Requirement: Partial changes continue through official artifact instructions
When an existing change is missing an artifact required for apply, the planner SHALL use `openspec status --change <name> --json` and `openspec instructions <artifact-id> --change <name> --json` with the active optional store id. It SHALL calculate the required transitive artifact set from the status graph, write only each instruction's `resolvedOutputPath`, and re-run status after each artifact. For named core skills, the planner SHALL instead permit the exact writes authorized by that skill within the planning home resolved by the official CLI, including new-change creation by `openspec-propose` and main-spec updates by `openspec-sync-specs`. It MUST NOT require or emulate the non-core continue workflow.

#### Scenario: Existing change is partially planned
- **WHEN** status reports that an apply-required artifact or one of its transitive dependencies is ready or missing
- **THEN** the planner follows official artifact instructions in dependency order until the required set is complete or no artifact can advance

#### Scenario: A conditional dependency is deliberately omitted
- **WHEN** a required artifact is blocked only by a dependency whose official instruction explicitly permits it to be omitted
- **THEN** the planner follows the required artifact's own instructions despite its blocked status and does not treat the conditional dependency as an absolute gate

#### Scenario: Partial change lives in a registered store
- **WHEN** the active change is in an explicit store
- **THEN** every applicable status and instructions command retains the store id and every write targets a path resolved from that store
