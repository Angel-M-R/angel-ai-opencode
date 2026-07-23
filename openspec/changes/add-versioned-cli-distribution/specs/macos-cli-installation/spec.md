## ADDED Requirements

### Requirement: Official one-line installation supports macOS Apple Silicon only
The project SHALL provide an official curl-based one-line installer for end users without Go or a repository checkout. The installer SHALL proceed only when the host reports macOS on Apple Silicon and SHALL reject Intel macOS, Linux, Windows environments, and unsupported architectures before installing a binary.

#### Scenario: Supported host runs the one-liner
- **WHEN** the official installer runs on `Darwin` with `arm64` architecture
- **THEN** it proceeds to obtain the latest stable release metadata

#### Scenario: Unsupported host runs the one-liner
- **WHEN** the official installer runs on any platform or architecture other than `Darwin/arm64`
- **THEN** it exits with an actionable unsupported-platform message without installing `angel-ai`

### Requirement: Installer obtains and verifies the latest stable artifact
The installer SHALL fetch the published latest-release manifest over HTTPS, download the artifact URL declared by that manifest, and verify the downloaded bytes against the manifest SHA-256 before installation. It MUST NOT install a file when metadata retrieval, download, or checksum verification fails.

#### Scenario: Download matches the manifest
- **WHEN** the downloaded macOS Apple Silicon artifact has the declared SHA-256 digest
- **THEN** the installer makes it executable and proceeds with installation

#### Scenario: Download does not match the manifest
- **WHEN** the downloaded artifact digest differs from the manifest digest
- **THEN** the installer reports verification failure and leaves any existing installation unchanged

### Requirement: Installer targets the user-local binary directory safely
The installer SHALL create `~/.local/bin` when absent and install the verified artifact as `~/.local/bin/angel-ai` through a temporary file and atomic replacement. A failed installation MUST leave an existing `angel-ai` executable usable.

#### Scenario: First installation succeeds
- **WHEN** verification succeeds and `~/.local/bin/angel-ai` does not exist
- **THEN** an executable `angel-ai` is atomically installed at that path

#### Scenario: Existing installation is replaced
- **WHEN** verification succeeds and `~/.local/bin/angel-ai` already exists
- **THEN** the verified artifact atomically replaces it without exposing a partial binary

#### Scenario: Installation fails before replacement
- **WHEN** directory creation, download, verification, permission setting, or temporary-file preparation fails
- **THEN** any existing `~/.local/bin/angel-ai` remains unchanged

### Requirement: PATH guidance is explicit and non-mutating
After installation, the installer SHALL detect whether `~/.local/bin` is represented in the effective `PATH`. When it is absent, the installer SHALL print the exact command `export PATH="$HOME/.local/bin:$PATH"` with guidance for applying it manually and MUST NOT edit `.zshrc` or any other shell profile.

#### Scenario: Install directory is already on PATH
- **WHEN** installation succeeds and `~/.local/bin` is already available through `PATH`
- **THEN** the installer reports that `angel-ai` is ready without requesting a profile change

#### Scenario: Install directory is missing from PATH
- **WHEN** installation succeeds and `~/.local/bin` is not available through `PATH`
- **THEN** the installer prints `export PATH="$HOME/.local/bin:$PATH"` and leaves all shell configuration files unchanged
