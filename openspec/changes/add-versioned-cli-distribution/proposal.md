## Why

Angel AI currently assumes users have Go and a local repository checkout, which prevents a usable first public release for end users. A versioned, self-contained macOS Apple Silicon distribution is needed so users can install and keep `angel-ai` current without those development prerequisites.

## What Changes

- Publish the first stable public release as `v0.1.0`, with a self-contained `angel-ai` binary for macOS Apple Silicon, its SHA-256 checksum, a latest-release manifest, and generated release notes.
- Embed the repository assets into the binary so installation does not require a repository checkout or a sibling `assets/` directory.
- Add an official curl-based one-line installer that places `angel-ai` in `~/.local/bin` and reports the exact `PATH` instruction when needed without editing shell configuration.
- Define `angel-ai` with no arguments as the TUI entry point while preserving the existing flags, and add `angel-ai update` and `angel-ai version` using the standard `flag` package.
- Add automatic stable-version checks on normal TUI launches, forced checks through `update`, verified downloads, atomic replacement, guarded relaunch, and limited immediate rollback.
- Inject stable versions from manual SemVer tags while keeping local builds at `dev`; `dev` builds neither check for nor apply updates.
- Gate release publication on unit, updater/installer integration, and released-artifact smoke coverage.
- Keep Intel macOS, Linux, Windows, Homebrew, beta channels, and Apple signing/notarization outside this initial scope.

## Capabilities

### New Capabilities
- `versioned-cli-interface`: Defines the installed command name, embedded runtime assets, version reporting, preserved flags, and command dispatch behavior.
- `macos-cli-installation`: Defines one-line installation of the macOS Apple Silicon binary into `~/.local/bin` and non-mutating `PATH` guidance.
- `cli-self-update`: Defines automatic and forced manifest-based updates, timeout and warning behavior, checksum verification, atomic replacement, relaunch protection, and limited rollback.
- `github-release-distribution`: Defines stable-tag release automation, version injection, published files and metadata, initial version, and release quality gates.

### Modified Capabilities

None.

## Impact

- The Go entry point and asset-loading boundary will change to support command dispatch, injected versions, and embedded assets while retaining existing installer flags and TUI behavior.
- A new internal standard-library updater module and corresponding unit/integration seams will be required.
- Repository release automation, an official installer script, and release-focused tests will be added.
- GitHub Releases becomes the public binary and manifest host; the updater will consume the manifest directly rather than the GitHub API.
- Existing installer behavior under `internal/catalog`, `internal/install`, and `internal/tui` remains the product payload and must operate from the embedded filesystem.
