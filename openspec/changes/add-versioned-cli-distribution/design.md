## Context

The current Go entry point parses four root flags, discovers `assets/` beside the executable or in the working directory, and passes filesystem paths through catalog, planning, installation, and TUI layers. That model works for `go run .` and repository builds but cannot produce a standalone end-user artifact. There is no product version, release workflow, installation script, command dispatch, or updater today.

This change crosses the CLI entry point, asset source abstraction, installer/TUI startup, a new internal updater, repository automation, and distribution tests. The first audience is users on macOS Apple Silicon who do not have Go and do not clone the repository.

## Goals / Non-Goals

**Goals:**

- Ship a single self-contained `angel-ai` binary for macOS Apple Silicon, beginning with `v0.1.0`.
- Preserve existing root flags while adding no-argument TUI startup plus `update` and `version` commands.
- Install through an official curl one-liner into `~/.local/bin` without modifying shell startup files.
- Keep stable installations current through a direct-manifest, checksum-verified, atomic self-update path.
- Make releases reproducible from stable SemVer tags and block publication until distribution-focused validation passes.

**Non-Goals:**

- Intel macOS, Linux, Windows, Homebrew, beta or prerelease channels.
- Apple code signing or notarization.
- Recovery from failures that occur after the replacement binary has successfully started.
- A general CLI framework, GitHub API client, external updater dependency, or background update service.

## Decisions

### 1. Treat embedded assets as the production default

Use `go:embed` to package the complete `assets/` tree and expose it through an `fs.FS`-based source boundary shared by catalog and installation preparation. The default CLI source is the embedded filesystem; the existing `--assets` flag remains as an explicit filesystem override for development and tests. Destination inspection and writes remain OS-filesystem operations.

This avoids extracting a permanent support directory and keeps the release to one artifact. Retaining `--assets` preserves existing behavior where callers intentionally supply editable assets. The rejected alternative is embedding followed by eager extraction, which creates lifecycle, cleanup, and partial-extraction failure modes without a runtime need for mutable source assets.

### 2. Keep command dispatch in the standard `flag` package

Add a small root dispatcher around dedicated `flag.FlagSet` values. No arguments, or existing root flags such as `--target`, launch the existing installer path; `version` prints the local version without constructing a network client; `update` invokes a forced update check. Existing `--assets`, `--target`, `--all`, and `--dry-run` semantics remain intact.

This keeps compatibility and avoids adding a framework for two subcommands. The rejected alternative is a third-party CLI framework, which would expand dependency and migration surface without providing needed behavior.

### 3. Use one injected version value with `dev` as the safe default

Define one package-level version value initialized to `dev`. Release builds set it from the triggering stable tag with `-ldflags`; the release pipeline validates that it matches `vMAJOR.MINOR.PATCH`. Only injected stable versions participate in update checks. `version` always reports the value as built and never performs network work.

This makes local builds predictable and prevents development binaries from replacing themselves. Generating a version source file was rejected because it would introduce generated-tree state and make local provenance less clear.

### 4. Use a direct latest-release manifest and a standard-library updater

Each stable GitHub Release carries a manifest at a stable `releases/latest/download` URL. The manifest contains exactly the stable version, artifact URL, and lowercase SHA-256 digest needed by both the updater and installer. A new internal updater uses `net/http`, `encoding/json`, `crypto/sha256`, filesystem primitives, and SemVer parsing owned by the module; it does not call the GitHub API.

The manifest request has a two-second timeout. Only a strictly newer stable SemVer is eligible; malformed, prerelease, equal, or older versions do not replace the current binary. Automatic checks run before each TUI start, while `update` forces the check even when no TUI is launched. Any check or application failure emits a warning and preserves the current version; TUI startup continues when applicable.

The direct manifest is simpler and less rate-limit-prone than GitHub API discovery. A third-party updater library was rejected to keep the update trust and replacement surface explicit and standard-library-only.

### 5. Verify before atomically replacing, then guard the relaunch

Download the candidate into the installed executable's directory, hash it before execution, apply executable permissions, and retain a same-directory backup of the current binary. Replace the executable with a same-filesystem atomic rename only after verification. On successful replacement, replace the running process with the new binary using the original argument vector and an internal environment marker.

If replacement or immediate process relaunch fails, restore the retained binary atomically and warn. The marker tells the relaunched process to skip another update check, clean up immediate-update state, and continue the original command, preventing loops. A later application crash is intentionally outside rollback scope.

In-place writes were rejected because interruption could leave a truncated executable. Spawning an independent updater helper was rejected because it would violate the single-binary distribution goal and add coordination complexity.

### 6. Share release metadata between the installer and updater

The official shell installer supports only `Darwin/arm64`, fetches the same latest manifest over HTTPS, downloads and verifies the declared artifact, and installs it as `~/.local/bin/angel-ai` through a temporary file and atomic rename. It creates `~/.local/bin` when needed. If that directory is not on `PATH`, it prints the exact export command `export PATH="$HOME/.local/bin:$PATH"` and guidance to add it manually, but never edits `.zshrc` or another profile.

Using the same manifest avoids duplicating latest-version discovery rules. Package managers and repository cloning were rejected because neither matches the first-release audience or scope.

### 7. Make a stable tag the sole release trigger and gate publication

A manually pushed tag matching `vMAJOR.MINOR.PATCH` triggers GitHub Actions. The workflow runs unit coverage, updater and installer integration coverage, builds with `GOOS=darwin`, `GOARCH=arm64`, and the tag-derived `ldflags`, then smoke-tests the exact artifact before publishing it, its SHA-256 file, the latest manifest, and automatically generated release notes. Publication does not proceed when any gate fails.

The artifact is unsigned and unnotarized in this phase, which must remain visible in release expectations. A branch-driven release and mutable manually uploaded artifacts were rejected because they weaken the relationship between source tag, embedded version, and published bytes.

Planning-task implementers may run focused tests for code they modify, but they must not run the full repository suite or any build. After all planned tasks are complete, the final `openspec-verifier` owns execution of the mandatory full unit, updater integration, and installer integration suites; release builds; exact packaged-artifact smoke validation; and the `v0.1.0` release-candidate procedure. Section 7 therefore prepares deterministic verifier and maintainer handoff artifacts without executing or claiming those gates. This ownership boundary avoids making task completion depend on final verification while preserving every publication requirement; the tag-triggered workflow independently enforces the same gates before release publication.

## Risks / Trade-offs

- [The manifest and binary are hosted under the same GitHub trust boundary, and no signature is present] → Require HTTPS and SHA-256 for corruption/substitution detection within the published release, document that signing/notarization is deferred, and keep release permissions minimal.
- [A two-second manifest timeout can be too short on slow networks] → Treat timeout as non-fatal, warn, and continue immediately with the installed version.
- [Self-replacement can fail because of directory permissions, symlinks, or filesystem behavior] → Use the resolved executable location, same-directory temporary files, explicit cleanup, and rollback before leaving the current process.
- [Refactoring source assets from paths to `fs.FS` touches mature installer code] → Preserve a filesystem-backed adapter for existing tests and `--assets`, and add parity tests for embedded and directory sources.
- [Automatic updates can surprise users] → Limit them to stable non-`dev` TUI launches, verify bytes before replacement, preserve arguments, and fail open with a visible warning as explicitly required.
- [Unsigned macOS binaries can trigger platform warnings] → Keep signing/notarization out of scope for `v0.1.0` and avoid implying Gatekeeper-free installation.

## Migration Plan

1. Introduce the asset source abstraction and embedded default while preserving filesystem overrides and current installer behavior.
2. Add version-aware command dispatch and the internal updater behind the `dev` update-disable rule.
3. Add the installer and release workflow with isolated integration fixtures and artifact smoke coverage.
4. Complete the end-user documentation and deterministic verifier and maintainer handoff artifacts without running a full suite, build, packaged-artifact smoke check, or release-candidate procedure as a planned task.
5. After all planned tasks are complete, have the final `openspec-verifier` execute the mandatory full suites, release build, exact packaged-artifact smoke validation, and `v0.1.0` release-candidate procedure and record every blocking result without weakening or pre-marking a gate.
6. Create and push the first stable tag `v0.1.0` only after the verifier records all release gates and candidate checks as passing; verify the published manifest, checksum, installer path, `version`, and no-argument startup against the exact artifact.
7. If release validation fails, do not move or reuse the tag; correct the source and issue a new stable patch tag. Existing installed binaries remain usable because update failures are non-fatal.

## Open Questions

None. The confirmed brief fixes the initial platform, channel, installation path, update policy, rollback boundary, and release contents.
