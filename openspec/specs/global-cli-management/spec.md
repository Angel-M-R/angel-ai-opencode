# global-cli-management Specification

## Purpose
TBD - created by archiving change manage-openspec-cli-bootstrap. Update Purpose after archive.
## Requirements
### Requirement: OpenSpec is a default installer extra
The installer SHALL expose an `OpenSpec` extra that is preselected in the interactive installer and selected by `--all`. Selecting it SHALL request installation or update of `@fission-ai/openspec@latest` without adding OpenSpec-specific MCP, `AGENTS.md`, profile, workflow, or delivery configuration.

#### Scenario: Interactive defaults include OpenSpec
- **WHEN** the interactive installer opens the extras step
- **THEN** the `OpenSpec` extra is selected by default

#### Scenario: Install all includes OpenSpec
- **WHEN** the installer runs with `--all`
- **THEN** its installation request includes the `OpenSpec` extra

### Requirement: OpenSpec skills install as one nested bundle
The repository SHALL retain all 12 vendored OpenSpec child skills under `assets/skills/openspec/<skill-name>/SKILL.md`. The installer catalog SHALL expose the top-level `openspec` directory as one selectable skill bundle, and installation SHALL recursively preserve every child path under `<config>/skills/openspec/` without removing, regenerating, or replacing individual skill definitions.

#### Scenario: Wizard loads the skill catalog
- **WHEN** the catalog scans the nested OpenSpec assets
- **THEN** it exposes one selectable `openspec` item rather than 12 top-level OpenSpec skill items

#### Scenario: OpenSpec bundle is selected
- **WHEN** the installer prepares the selected `openspec` skill bundle
- **THEN** it recursively copies all 12 child skills to `<config>/skills/openspec/<skill-name>/SKILL.md` without local skill generation

### Requirement: Gentle AI global cleanup is manual, backed up, and evidence-based
The installer MUST NOT detect, compare, back up, or remove legacy Gentle AI files. For this machine's one-time migration, cleanup SHALL use the Gentle AI source repository read-only as ownership evidence, create one timestamped backup of every item to be removed or edited before mutation, and limit deletion to the confirmed Gentle AI allowlist. The cleanup SHALL preserve the Angel-managed skill set and every file without ownership evidence, validate the resulting JSON configuration, prove retained skills remain byte-identical, and restart OpenCode only after validation succeeds.

#### Scenario: Installer encounters legacy Gentle AI files
- **WHEN** the installer writes the selected nested OpenSpec bundle to a destination that still contains legacy Gentle AI files
- **THEN** it leaves those legacy files untouched for the separate one-time cleanup

#### Scenario: Cleanup backup is prepared
- **WHEN** the confirmed Gentle AI-owned deletion and edit set has been inventoried
- **THEN** cleanup creates one timestamped backup containing that complete set before any mutation

#### Scenario: Confirmed Gentle AI footprint is cleaned
- **WHEN** the backup succeeds and ownership evidence matches the allowlist
- **THEN** cleanup removes exactly the 21 obsolete skill directories plus the confirmed legacy agent entries, commands, SDD prompts, plugins, global `.atl` directory, and ownership marker

#### Scenario: Ownership is absent or ambiguous
- **WHEN** a global file is not attributable to Gentle AI from the confirmed evidence
- **THEN** cleanup leaves it unchanged, including the Gentle AI source repository and executable, Engram, Context7, permissions, and Angel-managed agents

#### Scenario: Retained installation is validated
- **WHEN** cleanup has completed
- **THEN** `opencode.json` parses successfully and `cognitive-doc-design`, `technical-grilling`, `investigate`, `product-grilling`, and all 12 nested OpenSpec skills match their pre-cleanup bytes

#### Scenario: Cleanup validation succeeds
- **WHEN** JSON and byte-identity checks pass
- **THEN** OpenCode is restarted so the removed skills and agents are unloaded

### Requirement: Selected global CLIs target latest
For each selected CodeGraph or OpenSpec CLI, the installer SHALL use the selected package manager to inspect package registration, inspect any executable on `PATH`, and query the registry's `latest` version. It SHALL install the descriptor's `@latest` package only when the executable is absent, the selected manager does not register the package, and registry lookup is usable. A working executable SHALL NOT be installed, updated, downgraded, removed, or relinked, including when its installed version is older than or newer than the registry version.

#### Scenario: Selected CLI is absent and unregistered
- **WHEN** a selected CLI executable is absent, the selected manager does not register its package, and the registry reports a parseable latest version
- **THEN** the plan reports an installation action and application invokes the selected manager with that CLI's `@latest` package

#### Scenario: Selected CLI is current
- **WHEN** a selected CLI executable reports the same parseable version as the registry's latest version
- **THEN** planning and application report it as current without invoking a package installation command

#### Scenario: Selected CLI is outdated
- **WHEN** a selected CLI executable reports a parseable version older than the registry's latest version
- **THEN** planning and application report it as outdated and preserve it without invoking a package installation command

#### Scenario: Selected CLI is ahead of the registry
- **WHEN** a selected CLI executable reports a parseable version newer than the registry's latest version
- **THEN** planning and application report it as ahead of the registry and preserve it without invoking a package installation command

### Requirement: Package manager selection is deterministic
The installer SHALL use npm when npm is available. It SHALL use pnpm only when npm is unavailable and both pnpm is available and `pnpm bin -g` succeeds with a non-empty result. It SHALL fail preflight when neither manager is usable, SHALL NOT retry a failed npm installation with pnpm, and SHALL use only the selected manager for package-registration and registry-version inspection without querying or inferring ownership from the other manager.

#### Scenario: npm and pnpm are available
- **WHEN** preflight finds both npm and pnpm
- **THEN** it selects npm and performs no pnpm package inspection or installation

#### Scenario: npm is unavailable and pnpm global bin is valid
- **WHEN** npm is unavailable, pnpm is available, and `pnpm bin -g` returns a non-empty global bin directory
- **THEN** preflight selects pnpm for inspection and installation of every selected global CLI

#### Scenario: pnpm global bin is invalid
- **WHEN** npm is unavailable and `pnpm bin -g` fails or returns an empty value
- **THEN** planning and application fail before any package inspection that depends on pnpm, package installation, or configuration write

#### Scenario: npm installation fails while pnpm exists
- **WHEN** npm was selected and a global package command fails
- **THEN** application reports that failure without retrying or inspecting ownership through pnpm

#### Scenario: Working executable is not registered by the selected manager
- **WHEN** a selected CLI executable reports a parseable version but its package is not registered by the selected manager
- **THEN** the installer preserves the executable and does not query the other manager or claim cross-manager ownership

### Requirement: OpenSpec enforces its Node.js floor before installation
When OpenSpec is selected, planning and application SHALL require a parseable Node.js version greater than or equal to `20.19.0` before any selected CLI installation begins. Missing, malformed, or older Node.js versions SHALL fail preflight without package or configuration side effects.

#### Scenario: Supported Node.js version
- **WHEN** OpenSpec is selected and `node --version` reports `20.19.0` or newer
- **THEN** global CLI processing may proceed

#### Scenario: Unsupported Node.js version with multiple CLIs selected
- **WHEN** OpenSpec and another CLI are selected and Node.js is older than `20.19.0`
- **THEN** application fails before installing either CLI and before writing configuration

#### Scenario: OpenSpec is not selected
- **WHEN** only CodeGraph is selected
- **THEN** the OpenSpec-specific minimum Node.js check is not required

### Requirement: Planning is non-mutating and complete
`PlanInstallation` SHALL perform the same package-manager, OpenSpec runtime, package-registration, executable-version, and registry-version preflight needed by application for every selected CLI. It SHALL describe each CLI as pending installation, current, outdated, ahead of the registry, or registry-unverified together with file actions. It MUST NOT invoke package installation commands, repair packages, remove packages, or write configuration, including during `--dry-run`.

#### Scenario: Dry run with selected CLIs
- **WHEN** a valid request is planned or executed through `--all --dry-run`
- **THEN** output includes the status or installation action for every selected CLI and no package installation, cleanup, relinking, or file reconciliation occurs

#### Scenario: Dry run detects invalid environment
- **WHEN** any selected CLI cannot pass package-manager, Node.js, package-registration, executable-version, or required registry preflight
- **THEN** planning returns the preflight error without mutating the machine

#### Scenario: Multiple CLIs require inspection
- **WHEN** more than one global CLI is selected
- **THEN** planning inspects every selected CLI before reporting any action that application could install

### Requirement: Application defers writes until all CLI actions succeed
`ApplyInstallation` SHALL prepare and validate the full request, complete prevalidation for every selected CLI before installing any package, and process prevalidated records sequentially in deterministic descriptor order. It SHALL install only records classified as absent and installable, preserve healthy installed records, and write no configuration until every required installation succeeds and each installed executable reports a parseable version. If a later installation fails, successful earlier installations MAY remain as external side effects, but their results SHALL be reported and no prepared file SHALL be reconciled.

#### Scenario: All selected CLIs pass prevalidation
- **WHEN** every selected CLI is either healthy or absent and installable, and every required installation succeeds with a parseable executable version
- **THEN** application proceeds to a single repreparation followed by configuration reconciliation

#### Scenario: Later CLI is invalid during prevalidation
- **WHEN** an earlier selected CLI is absent and installable but a later selected CLI fails prevalidation
- **THEN** application installs neither CLI and writes no configuration

#### Scenario: Later installation fails
- **WHEN** all selected CLIs pass prevalidation, an earlier required installation succeeds, and a later required installation fails
- **THEN** application returns the earlier success and later error while leaving every destination configuration file unwritten

#### Scenario: Package command succeeds but executable is unavailable or invalid
- **WHEN** a required package command exits successfully but the resulting executable is unavailable or its version is not parseable
- **THEN** application fails before configuration writes and does not perform automatic cleanup

#### Scenario: Installed CLI requires no package action
- **WHEN** prevalidation classifies a selected CLI as current, outdated, ahead of the registry, or registry-unverified
- **THEN** application reports that classification without invoking an install command for that CLI

### Requirement: Application reprepares exactly once after CLI processing
When at least one global CLI is selected and every CLI action succeeds, `ApplyInstallation` SHALL rerun full installation preparation exactly once before file reconciliation. It SHALL NOT reprepare after each individual CLI.

#### Scenario: Two selected CLIs succeed
- **WHEN** CodeGraph and OpenSpec updates both complete successfully
- **THEN** application performs one post-CLI repreparation before any file write

#### Scenario: A selected CLI fails
- **WHEN** any global CLI update fails
- **THEN** application does not perform the post-success repreparation or reconcile files

### Requirement: CodeGraph configuration remains separate
Selecting CodeGraph SHALL continue to reconcile its MCP registration and managed `AGENTS.md` guidance, while its global executable inspection and conditional package installation use the shared CLI mechanism. Selecting OpenSpec alone SHALL not trigger CodeGraph-specific configuration.

#### Scenario: CodeGraph is selected
- **WHEN** CodeGraph is selected in an installation request
- **THEN** its CLI status or required `@latest` installation action and its existing MCP and `AGENTS.md` changes are all represented

#### Scenario: Only OpenSpec is selected
- **WHEN** OpenSpec is selected and CodeGraph is explicitly unselected
- **THEN** no CodeGraph MCP entry or managed `AGENTS.md` block is added

### Requirement: Broken selected-manager CLI state blocks with guided recovery
The installer SHALL block before any package installation or configuration write when the selected package manager registers a selected CLI package but its executable is unavailable, or when an available executable's version command fails, returns no version, or returns an uninterpretable version. The error SHALL identify the selected manager and CLI, provide recovery instructions, and MUST NOT remove, reinstall, relink, or otherwise clean up the package automatically.

#### Scenario: Registered package has no executable
- **WHEN** the selected manager registers a selected CLI package but its executable is not available on `PATH`
- **THEN** planning and application block with selected-manager recovery instructions and perform no installation or cleanup

#### Scenario: Executable version command fails
- **WHEN** a selected CLI executable is available but its version command exits unsuccessfully
- **THEN** planning and application block with recovery instructions before any package installation or configuration write

#### Scenario: Executable version is uninterpretable
- **WHEN** a selected CLI version command returns an empty or non-semantic version
- **THEN** planning and application block with recovery instructions rather than guessing or modifying the installation

### Requirement: Registry failures preserve healthy installed CLIs
When registry latest-version lookup fails or returns an uninterpretable version, the installer SHALL preserve a selected CLI whose executable reports a healthy parseable version, continue processing with a warning, and report its status as registry-unverified in both planning and application output. If that CLI is absent, the installer SHALL block during complete prevalidation before any selected CLI installation begins.

#### Scenario: Registry is unavailable for a healthy installed CLI
- **WHEN** a selected CLI executable reports a parseable installed version but registry latest-version lookup fails
- **THEN** planning and application preserve the executable, report a registry-unverified warning, and continue without an install command for that CLI

#### Scenario: Registry version is malformed for a healthy installed CLI
- **WHEN** a selected CLI executable reports a parseable installed version but the registry returns an uninterpretable latest version
- **THEN** planning and application preserve the executable, report a registry-unverified warning, and continue

#### Scenario: Registry is unavailable for an absent CLI
- **WHEN** a selected CLI is absent and its registry latest version cannot be obtained and parsed
- **THEN** application blocks before installing any selected CLI or writing configuration
