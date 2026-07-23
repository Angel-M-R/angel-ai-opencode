## ADDED Requirements

### Requirement: Installed command opens the TUI by default
The distributed executable SHALL be named `angel-ai`. Invoking it without a subcommand SHALL open the existing interactive TUI, including when supported root flags configure that TUI invocation.

#### Scenario: User invokes the installed command without arguments
- **WHEN** the user runs `angel-ai`
- **THEN** the interactive installer TUI opens

#### Scenario: User configures a TUI invocation
- **WHEN** the user runs `angel-ai --target <directory>`
- **THEN** the TUI opens using the requested target directory

### Requirement: Existing root flags remain compatible
The CLI SHALL preserve the `--assets`, `--target`, `--all`, and `--dry-run` flags and their existing behavior. `--dry-run` SHALL remain meaningful with `--all`, and an explicit `--assets` value SHALL select that filesystem source instead of embedded assets.

#### Scenario: User runs non-interactive installation
- **WHEN** the user runs `angel-ai --all`
- **THEN** the existing non-interactive installation path runs with all selectable content

#### Scenario: User requests a dry run
- **WHEN** the user runs `angel-ai --all --dry-run`
- **THEN** the CLI reports the installation plan without applying it

#### Scenario: Developer supplies an asset directory
- **WHEN** the user runs the CLI with `--assets <directory>`
- **THEN** catalog and installation reads use that directory rather than the embedded source

### Requirement: Production assets are self-contained
The release binary SHALL embed the complete runtime `assets/` tree and SHALL use those embedded assets when `--assets` is not specified. Normal installed operation MUST NOT require Go, a repository checkout, the current working directory, or a sibling assets directory.

#### Scenario: Installed binary runs outside the repository
- **WHEN** `angel-ai` is launched from an arbitrary working directory without `--assets`
- **THEN** the TUI loads its catalog and installation content from the embedded assets

#### Scenario: Embedded nested content is selected
- **WHEN** a user selects content located in a nested assets directory
- **THEN** installation preserves the same relative files and bytes as the repository asset tree

### Requirement: Version and update subcommands are available
The CLI SHALL expose `angel-ai version` and `angel-ai update` in addition to the root installer flags. Command parsing SHALL use Go's standard `flag` package and SHALL NOT introduce a CLI framework dependency.

#### Scenario: User requests the local version
- **WHEN** the user runs `angel-ai version`
- **THEN** the CLI prints the version embedded in that executable and exits without opening the TUI

#### Scenario: User requests an update
- **WHEN** the user runs `angel-ai update`
- **THEN** the CLI runs a forced update check without opening the TUI unless a successful replacement relaunch requires command completion

### Requirement: Version reporting is offline
The `version` command SHALL NOT perform a network request or trigger update application. A local build without an injected release version SHALL report `dev`.

#### Scenario: Stable build reports its tag version
- **WHEN** `angel-ai version` runs from a release build produced from tag `v0.1.0`
- **THEN** it reports `v0.1.0` without consulting the network

#### Scenario: Local build reports development version
- **WHEN** `angel-ai version` runs from a build without release `ldflags`
- **THEN** it reports `dev` without consulting the network
