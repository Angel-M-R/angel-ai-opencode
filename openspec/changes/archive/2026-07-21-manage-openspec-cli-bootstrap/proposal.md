## Why

OpenSpec workers currently assume that a usable project root and CLI already exist, while the installer manages CodeGraph through one-off logic and cannot provision OpenSpec itself. The 12 OpenSpec skills also appear as separate top-level wizard entries and global directories, so a shared CLI path, bounded worker bootstrap, and one recursively installed OpenSpec skill bundle are needed.

## What Changes

- Add a preselected `OpenSpec` installer extra, included by `--all`, that installs or updates `@fission-ai/openspec@latest` globally.
- Replace the CodeGraph-specific global install path with a shared internal global-CLI module used by both CodeGraph and OpenSpec, while retaining CodeGraph-only MCP and `AGENTS.md` reconciliation outside that module.
- Update every selected CLI sequentially to its `latest` package version before writing configuration, with npm preferred and pnpm used only as a validated fallback.
- Validate Node.js `>=20.19.0` before any installation when OpenSpec is selected, preserve dry-run behavior, and reprepare the installation once after all CLI operations succeed.
- Add a short `general`-agent bootstrap before the first OpenSpec worker launch for each project/store context, using `openspec list --json` and, only when no root resolves, `openspec init --tools none`.
- Move all 12 vendored OpenSpec skills to `assets/skills/openspec/<skill-name>/SKILL.md`, expose `openspec` as one wizard bundle, and recursively install that tree at `~/.config/opencode/skills/openspec/` without changing the individual skill contents.
- Treat cleanup of the previous Gentle AI-managed global OpenCode installation as a one-time migration on this machine, not installer behavior: create a timestamped backup, remove only source-attributed legacy skills and configuration, and preserve the current Angel-managed OpenSpec bundle and unrelated configuration.
- Add focused coverage for grouped catalog discovery and recursive skill copying, validate the cleaned global JSON configuration and byte-identical preservation of retained skills, and restart OpenCode after the global migration.
- Block worker launch with installer-extra guidance when the CLI is absent, warn but continue on CLI/skill version mismatch, and silently tolerate duplicate local skills.

## Capabilities

### New Capabilities
- `global-cli-management`: Declarative planning and transactional installation or update of selected global CLIs, plus grouped discovery and recursive installation of the nested OpenSpec skill bundle.
- `openspec-worker-bootstrap`: Context-aware OpenSpec readiness checks and minimal project initialization before OpenSpec workers run.

### Modified Capabilities

None.

## Impact

- Installer extras, `--all` selection, catalog grouping, recursive directory installation, and the current CodeGraph installer path under `internal/catalog/`, `internal/install/`, and their CLI/TUI callers.
- OpenSpec orchestration and worker-launch guidance in managed agent assets.
- OpenSpec skill assets, this machine's global OpenCode skill and configuration layout, timestamped migration backup, and OpenCode process state after migration.
- Installer and orchestration tests covering package managers, Node.js compatibility, dry runs, existing CLIs, partial failures, repreparation, grouped catalog discovery, recursive copying, and per-context bootstrap behavior.
- Global npm or pnpm package installation for CodeGraph and OpenSpec; Gentle AI remains only a verified behavioral reference and is not introduced as a dependency.
