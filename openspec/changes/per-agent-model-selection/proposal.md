## Why

The seven Angel AI agents ship with a hardcoded `variant` in their markdown frontmatter and no model at all, so every user gets the same reasoning effort and whatever model their global OpenCode default happens to be. There is no way to say "run the orchestrator on a cheap model and the implementer on a strong one" without hand-editing installed files, which the next reinstall overwrites verbatim.

## What Changes

- Add a new installer wizard step, after the existing extras step, that lets the user assign a **model** and a **reasoning effort** to each of the seven Angel AI agents individually (`angel-orchestrator`, `openspec-planner`, `openspec-implementer`, `openspec-verifier`, `review-correctness`, `review-security-risk`, `review-simplicity`).
- Read the available model catalog from OpenCode's own cache at `~/.cache/opencode/models.json`; derive effort levels from each model's `reasoning_options` entry of `type: "effort"`.
- Restrict the offered providers to those the user is actually authenticated against (top-level keys of `~/.local/share/opencode/auth.json`) and to models with `tool_call: true`, falling back to all providers when the auth file is missing or empty.
- Write assignments into `~/.config/opencode/opencode.json` as `agent.<name>.model` and `agent.<name>.variant`, via the installer's existing JSON deep-merge.
- **BREAKING (behavioral, opt-in)**: remove the hardcoded `variant` from the seven agent markdown files under `assets/agents/`, making `opencode.json` the single source of truth for model and effort.
- Preserve today's behavior when the user makes no choice: skipping the step, or running `--all`, writes neither `model` nor `variant`, so agents keep inheriting the user's global OpenCode default.
- On reinstall, preload the current assignments from the existing `opencode.json` as the wizard's initial selection.

Non-goals: role presets, a global default with per-agent override, and configuring agents not owned by Angel AI.

## Capabilities

### New Capabilities
- `agent-model-assignment`: how the installer discovers the model catalog, filters it by authentication and tool-calling support, lets the user assign a model plus reasoning effort per Angel AI agent, and persists (or deliberately omits) that assignment in `opencode.json`.

### Modified Capabilities
<!-- None: no existing spec's requirements change. -->

## Impact

- `internal/tui/wizard.go` — new phase between `extrasPhase` and `analyzing`/`confirming`, plus step counting and confirmation summary.
- `internal/install/` — new catalog-reading and assignment-writing data layer; `InstallationRequest` gains the agent assignments; `prepareInstallation` feeds them into the existing `opencode.json` fragment merge.
- `assets/agents/*.md` — the `variant:` frontmatter line is removed from all seven files.
- `main.go` — `--all` passes an empty assignment set (no writes).
- Read-only external inputs: `~/.cache/opencode/models.json`, `~/.local/share/opencode/auth.json` (keys only; secret values are never read or logged).
- No new Go dependencies; tests stay at `go test ./...` with JSON fixtures.
