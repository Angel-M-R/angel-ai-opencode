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
The installer SHALL plan and execute an update to `@latest` for every selected global CLI, including CodeGraph and OpenSpec, regardless of whether its executable is already present. The installer MUST NOT detect or preserve prior package-manager ownership.

#### Scenario: Selected CLI is absent
- **WHEN** a selected global CLI executable is not on `PATH`
- **THEN** the plan reports its `@latest` package action and application invokes the selected package manager for that package

#### Scenario: Selected CLI is already present
- **WHEN** a selected CodeGraph or OpenSpec executable is already on `PATH`
- **THEN** application still invokes the selected package manager with that CLI's `@latest` package

### Requirement: Package manager selection is deterministic
The installer SHALL use npm when npm is available. It SHALL use pnpm only when npm is unavailable and both pnpm is available and `pnpm bin -g` succeeds with a non-empty result. It SHALL fail preflight when neither manager is usable, and SHALL NOT retry a failed npm installation with pnpm.

#### Scenario: npm and pnpm are available
- **WHEN** preflight finds both npm and pnpm
- **THEN** it selects npm without using pnpm for installation

#### Scenario: npm is unavailable and pnpm global bin is valid
- **WHEN** npm is unavailable, pnpm is available, and `pnpm bin -g` returns a non-empty global bin directory
- **THEN** preflight selects pnpm for every selected global CLI

#### Scenario: pnpm global bin is invalid
- **WHEN** npm is unavailable and `pnpm bin -g` fails or returns an empty value
- **THEN** planning and application fail before any package installation or configuration write

#### Scenario: npm installation fails while pnpm exists
- **WHEN** npm was selected and a global package command fails
- **THEN** application reports that failure without retrying through pnpm

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
`PlanInstallation` SHALL perform the same package-manager and OpenSpec runtime preflight needed by application and SHALL describe every selected `@latest` CLI action together with file actions. It MUST NOT invoke package installation commands or write configuration, including during `--dry-run`.

#### Scenario: Dry run with selected CLIs
- **WHEN** a valid request is planned or executed through `--all --dry-run`
- **THEN** output includes every selected CLI action and no package installation or file reconciliation occurs

#### Scenario: Dry run detects invalid environment
- **WHEN** a selected CLI cannot pass package-manager or Node.js preflight
- **THEN** planning returns the preflight error without mutating the machine

### Requirement: Application defers writes until all CLI actions succeed
`ApplyInstallation` SHALL prepare and validate the full request, install selected CLIs sequentially in deterministic descriptor order, and write no configuration until every CLI action succeeds and each resulting executable is available. If a later CLI fails, successful earlier CLI actions MAY remain as external side effects, but their results SHALL be reported and no prepared file SHALL be reconciled.

#### Scenario: All selected CLIs succeed
- **WHEN** every sequential global CLI action succeeds and exposes its executable
- **THEN** application proceeds to a single repreparation followed by configuration reconciliation

#### Scenario: Later CLI fails
- **WHEN** an earlier selected CLI succeeds and a later selected CLI fails
- **THEN** application returns the earlier success and later error while leaving every destination configuration file unwritten

#### Scenario: Package command succeeds but executable is unavailable
- **WHEN** a selected CLI package command exits successfully but its executable cannot be resolved afterward
- **THEN** application fails before configuration writes

### Requirement: Application reprepares exactly once after CLI processing
When at least one global CLI is selected and every CLI action succeeds, `ApplyInstallation` SHALL rerun full installation preparation exactly once before file reconciliation. It SHALL NOT reprepare after each individual CLI.

#### Scenario: Two selected CLIs succeed
- **WHEN** CodeGraph and OpenSpec updates both complete successfully
- **THEN** application performs one post-CLI repreparation before any file write

#### Scenario: A selected CLI fails
- **WHEN** any global CLI update fails
- **THEN** application does not perform the post-success repreparation or reconcile files

### Requirement: CodeGraph configuration remains separate
Selecting CodeGraph SHALL continue to reconcile its MCP registration and managed `AGENTS.md` guidance, while its global package action uses the shared CLI mechanism. Selecting OpenSpec alone SHALL not trigger CodeGraph-specific configuration.

#### Scenario: CodeGraph is selected
- **WHEN** CodeGraph is selected in an installation request
- **THEN** its `@latest` CLI action and its existing MCP and `AGENTS.md` changes are all represented

#### Scenario: Only OpenSpec is selected
- **WHEN** OpenSpec is selected and CodeGraph is explicitly unselected
- **THEN** no CodeGraph MCP entry or managed `AGENTS.md` block is added

