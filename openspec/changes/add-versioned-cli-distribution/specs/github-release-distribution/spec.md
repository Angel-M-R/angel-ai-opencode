## ADDED Requirements

### Requirement: Stable SemVer tags drive releases
A manually created and pushed tag matching `vMAJOR.MINOR.PATCH` SHALL trigger the release workflow. Prerelease tags, beta channels, branches, and non-SemVer tags MUST NOT publish a production release through this workflow.

#### Scenario: Stable tag is pushed
- **WHEN** a maintainer pushes a tag such as `v0.1.0`
- **THEN** GitHub Actions starts the stable release pipeline for that exact tag

#### Scenario: Non-stable ref is pushed
- **WHEN** a branch, malformed tag, or prerelease tag is pushed
- **THEN** the stable release workflow does not publish a production release

### Requirement: Release version is injected from the tag
The release build SHALL receive its version from the triggering tag through Go linker flags, and the workflow SHALL validate that the binary reports that exact stable tag version. Builds without this injection SHALL retain the version `dev`.

#### Scenario: Tagged binary is built
- **WHEN** the release workflow builds tag `v0.1.0`
- **THEN** the resulting artifact reports `v0.1.0` through `angel-ai version`

#### Scenario: Local binary is built
- **WHEN** a developer builds without release linker flags
- **THEN** the resulting executable reports `dev`

### Requirement: Initial distribution targets macOS Apple Silicon only
The release workflow SHALL build one self-contained artifact with `GOOS=darwin` and `GOARCH=arm64`. It SHALL NOT publish Intel macOS, Linux, or Windows binaries in the initial distribution.

#### Scenario: Release build matrix is evaluated
- **WHEN** the stable workflow prepares platform artifacts
- **THEN** it produces only the macOS Apple Silicon `angel-ai` artifact

### Requirement: Every release publishes complete update metadata
Each stable GitHub Release SHALL publish the macOS Apple Silicon binary, its SHA-256 checksum file, and a latest-release manifest containing the release version, direct artifact URL, and digest. The release SHALL include automatically generated release notes, and the manifest SHALL be reachable through a stable latest-release download URL without the GitHub API.

#### Scenario: Stable release is published
- **WHEN** all release gates succeed for a stable tag
- **THEN** the GitHub Release contains the binary, checksum, manifest, and automatically generated notes for the same version

#### Scenario: Updater resolves latest metadata
- **WHEN** an installed client requests the stable latest-release manifest URL
- **THEN** GitHub Releases serves or redirects to the manifest for the latest stable release without an API request

### Requirement: First public release is v0.1.0
The first production artifact published under this distribution contract SHALL use version `v0.1.0`.

#### Scenario: Initial public release is created
- **WHEN** maintainers publish the first stable release for end users
- **THEN** its tag, injected binary version, manifest version, and release title identify `v0.1.0`

### Requirement: Validation blocks publication
The release workflow SHALL require successful unit tests, updater and installer integration tests, and smoke validation of the exact built artifact before creating or updating a GitHub Release. A failing or skipped required gate MUST prevent publication. Planned-task implementers MAY run focused tests for code they modify but MUST NOT execute the full repository suite or any build as task completion work. After all planned tasks are complete, the final `openspec-verifier` SHALL execute the mandatory full suites, release builds, exact packaged-artifact smoke validation, and `v0.1.0` release-candidate procedure before release approval; handoff preparation MUST NOT be represented as successful execution of any gate.

#### Scenario: All required gates pass
- **WHEN** unit, updater integration, installer integration, and exact-artifact smoke validation succeed
- **THEN** the workflow may publish the release assets and notes

#### Scenario: A required gate fails
- **WHEN** any required validation fails or the artifact reports a version different from the tag
- **THEN** the workflow stops without publishing the release

#### Scenario: Planned tasks prepare final verification
- **WHEN** planned-task implementation and handoff preparation are in progress
- **THEN** implementers may run focused tests of modified code but leave all mandatory full-suite, build, packaged-artifact smoke, and release-candidate execution to the final `openspec-verifier` without marking those gates as passed

#### Scenario: Final verifier evaluates release readiness
- **WHEN** all planned tasks are complete
- **THEN** the final `openspec-verifier` executes every mandatory full suite, release build, exact packaged-artifact smoke check, and `v0.1.0` release-candidate check before release approval

### Requirement: Initial artifacts are unsigned
The initial release process SHALL NOT claim Apple signing or notarization and SHALL NOT make either a publication requirement for `v0.1.0`.

#### Scenario: Initial release completes
- **WHEN** `v0.1.0` passes all in-scope release gates
- **THEN** it may publish without Apple signing or notarization and without representing the artifact as signed or notarized
