## 1. Embedded Asset Runtime

- [x] 1.1 Introduce an `fs.FS`-based asset source abstraction for catalog loading and installation preparation while keeping destination reads and writes on the OS filesystem.
- [x] 1.2 Embed the complete `assets/` tree and make it the default source when `--assets` is omitted, while retaining a filesystem-backed `--assets` override.
- [x] 1.3 Adapt catalog, installer, and TUI call boundaries to consume the source abstraction without changing existing selection, planning, reconciliation, or backup behavior.
- [x] 1.4 Add parity tests proving embedded and directory-backed sources expose identical nested files and that an installed-style invocation has no repository or working-directory dependency.

## 2. Versioned CLI Interface

- [x] 2.1 Refactor the entry point into standard-library `flag.FlagSet` dispatch for the root installer path, `version`, and `update`, preserving `--assets`, `--target`, `--all`, and `--dry-run` behavior.
- [x] 2.2 Add one linker-injectable version value with a `dev` default and make `angel-ai version` print it without constructing or invoking update networking.
- [x] 2.3 Wire no-argument and root-flag interactive invocations to perform the update policy before opening the existing TUI, while non-interactive root invocations preserve current behavior.
- [x] 2.4 Add CLI tests covering no-argument TUI dispatch, every preserved root flag path, unknown command/flag errors, offline `version`, forced `update`, and `dev` update suppression.

## 3. Manifest and Update Policy

- [x] 3.1 Create an internal standard-library updater module with strict stable SemVer parsing/comparison and validation for manifest version, HTTPS artifact URL, and lowercase SHA-256 fields.
- [x] 3.2 Implement direct latest-release manifest retrieval with a two-second request timeout and injectable HTTP/filesystem/process seams for deterministic tests.
- [x] 3.3 Implement automatic stable-build checks before every TUI launch, forced checks for `update`, equal/older release handling, and complete network bypass for `dev` builds.
- [x] 3.4 Add unit tests for valid, malformed, prerelease, equal, older, and newer manifests plus timeout, transport, status, decoding, and `dev` cases.

## 4. Verified Atomic Self-Update

- [x] 4.1 Download eligible artifacts to same-directory temporary files, compute and compare SHA-256 before replacement, apply executable permissions, and clean temporary state on every failure path.
- [x] 4.2 Implement same-directory backup and atomic executable replacement so replacement failures retain or restore the current binary.
- [x] 4.3 Relaunch the replacement with the original argument vector and an internal environment marker that prevents repeated checks and completes both TUI and explicit-update flows.
- [x] 4.4 Implement immediate rollback when replacement or relaunch fails, without claiming rollback after the new application process has started.
- [x] 4.5 Emit visible non-fatal warnings for discovery and application failures and continue the requested TUI with the running version whenever safe.
- [x] 4.6 Add updater integration tests with temporary executables and local HTTP fixtures covering checksum mismatch, successful atomic replacement, argument preservation, loop prevention, replacement failure, relaunch failure, rollback, and warning-based continuation.

## 5. Official macOS Apple Silicon Installer

- [x] 5.1 Add the official curl-pipe-compatible shell installer with an early `Darwin/arm64` gate and actionable rejection for every unsupported host.
- [x] 5.2 Fetch and validate the shared latest-release manifest, download and SHA-256-verify its artifact, and atomically install it as executable `~/.local/bin/angel-ai` without requiring Go or a repository clone.
- [x] 5.3 Detect whether `~/.local/bin` is on `PATH`; when absent, print exactly `export PATH="$HOME/.local/bin:$PATH"` with manual guidance and never edit a shell profile.
- [x] 5.4 Add isolated installer integration tests for supported and unsupported platforms, first install, replacement, download/manifest/checksum failures, existing-binary preservation, PATH-present output, PATH-missing output, and absence of profile writes.

## 6. Stable Release Pipeline

- [x] 6.1 Add a GitHub Actions workflow triggered only by manually pushed stable `vMAJOR.MINOR.PATCH` tags and validate the tag before any publication step.
- [x] 6.2 Gate release work on the full unit suite plus updater and installer integration suites, with publication impossible when a required job fails or is skipped.
- [x] 6.3 Build only the self-contained macOS Apple Silicon artifact with tag-derived `ldflags`, then verify its reported version matches the triggering tag.
- [x] 6.4 Generate the artifact SHA-256 file and latest manifest with matching version, direct artifact URL, and digest, then configure GitHub Releases to publish them with automatically generated notes.
- [x] 6.5 Add smoke CI against the exact packaged artifact for executable format/architecture, offline `version`, embedded-asset operation outside the repository, and no-argument TUI startup.
- [x] 6.6 Ensure workflow naming and release text do not claim Apple signing/notarization and do not create Intel, Linux, Windows, Homebrew, or beta outputs.

## 7. Distribution Validation and Handoff

- [x] 7.1 Update end-user installation documentation with the official one-liner, `~/.local/bin` destination, command examples, automatic-update behavior, supported platform, and explicit initial non-goals.
- [x] 7.2 Prepare a deterministic final-verifier handoff that identifies the exact commands, inputs, environment, and evidence locations for the mandatory full unit, updater integration, and installer integration suites and release build, without executing them or marking their gates as passed.
- [x] 7.3 Document the deterministic exact packaged-artifact smoke and `v0.1.0` release-candidate procedure, including manifest, checksum, isolated one-line install, offline `angel-ai version`, embedded-assets operation, and no-argument TUI checks, for execution only by the final `openspec-verifier` after all planned tasks are complete.
- [x] 7.4 Prepare maintainer release instructions and handoff artifacts that require recorded final-verifier success for every full-suite, build, packaged-artifact smoke, and release-candidate gate before manually creating and pushing the immutable `v0.1.0` tag.
