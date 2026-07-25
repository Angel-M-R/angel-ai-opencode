## Context

`angel-ai-opencode` is a Go 1.25.3 CLI (bubbletea v1.3.10 + lipgloss as the only direct deps) that installs assets into `~/.config/opencode`. Assets are copied verbatim — there is no templating — and JSON fragments from `assets/fragments/*.json` are deep-merged into `opencode.json` by `prepareJSONObject`/`merge` in `internal/install/`.

Today the seven Angel AI agents hardcode `variant: "high"`/`"xhigh"` in their markdown frontmatter and never set a model. Because assets are copied verbatim, the only way to change either is to edit installed files, which the next install overwrites. The wizard (`internal/tui/wizard.go`) currently runs `selecting -> extrasPhase -> analyzing -> confirming -> installing -> finished`.

Verified facts this design rests on:
- `~/.cache/opencode/models.json` exists and is a provider-keyed object (172 providers, 5823 models). Each model carries `tool_call`, `reasoning`, and `reasoning_options` (an array where the effort entry has `type: "effort"` and a `values` array such as `["low","medium","high","max"]`). No external variants plugin is needed.
- `~/.local/share/opencode/auth.json` is an object whose top-level keys are provider ids.

## Goals / Non-Goals

**Goals:**
- Let the user choose a model + reasoning effort per Angel AI agent during the install wizard.
- Make `opencode.json` the single source of truth for that choice.
- Keep today's behavior byte-for-byte when the user makes no choice (skip, or `--all`).
- Add no new Go dependencies and no new install-time network calls.

**Non-Goals:**
- Role presets (e.g. "cheap", "max power").
- A global default with per-agent override, or an "apply to all agents" row.
- Configuring agents that Angel AI does not own.
- Managing authentication or provider setup.

## Decisions

### 1. Read the catalog from OpenCode's own cache instead of shipping one

`~/.cache/opencode/models.json` is already maintained by OpenCode and stays current without any work from this installer. Alternatives rejected: (a) vendoring a static model list — goes stale immediately and bloats the binary; (b) fetching models.dev over HTTP at install time — adds a network dependency and a failure mode to a tool that is otherwise offline.

Consequence: the catalog is an *optional* input. If the file is missing or unparseable, the feature degrades to "not offered" rather than failing the install. This is the same posture the installer already takes toward missing asset subdirectories in `catalog.Load`.

### 2. Filter by `auth.json` keys and `tool_call: true`

172 providers / 5823 models is unusable in a terminal picker. Authentication is the sharpest available signal for "the user can actually run this", and `tool_call: true` is a hard requirement for an OpenCode agent. Filtering happens in the data layer, before the TUI sees anything.

Only the *keys* of `auth.json` are consumed. The parsed values are discarded immediately; nothing from that file other than a provider id may reach a struct field, the TUI, or an error message. Fallback: if the file is absent or has no keys, offer all providers (still `tool_call`-filtered) — a user with `OPENAI_API_KEY` in their environment and no `auth.json` should not see an empty picker.

### 3. Always write both `model` and `variant` for assigned agents; write nothing for unassigned ones

The installer's `merge` is a deep-merge, so an absent key preserves whatever is already in `opencode.json`. If only `model` were written, a stale `variant: "xhigh"` from a previous install could survive onto a model that has no such effort level. Writing both together makes each assignment self-consistent. When the chosen model has no effort levels, `variant` is written as `""` — an explicit "no effort" that overwrites any previous value.

Symmetrically, an unassigned agent must produce *no keys at all*, not `""` or `null`, so that "user made no choice" is indistinguishable from the pre-change state and the agent inherits the global default.

The assignments are expressed as one more `map[string]any` patch appended to the existing `fragments` slice in `prepareInstallation`, so they flow through the same merge, the same backup, and the same plan/apply reporting as every other fragment. No new write path.

Rejected alternative: templating the `variant` into the agent markdown at copy time. It would break the "assets are copied verbatim" invariant the whole installer is built on, and would leave two sources of truth.

### 4. Assignment lives in `InstallationRequest`

`InstallationRequest` is already the single object both `PlanInstallation` and `ApplyInstallation` consume, which is what stops plan and apply from disagreeing. Adding an `AgentModels` field (agent name -> {model, variant}) preserves that property: the confirmation plan reflects the assignments the user just made. An empty/nil map means "no assignments", which is what `--all` passes.

### 5. A separate wizard phase, shown conditionally

A new phase sits between `extrasPhase` and the confirmation transition. It is shown only when the `agents` category is selected *and* the catalog is available; otherwise the wizard keeps its current sequence and step numbering exactly. The phase is a small drill-down: agent list -> provider -> model -> effort, with `s` to skip at each level and back-navigation consistent with the existing `left/h/b` convention. The effort level is skipped entirely for models with no effort values.

The picker is a nested cursor rather than a filterable list because bubbletea's list/textinput components are not current direct dependencies and adding them for one step is not worth the dependency surface. Filtering by auth already reduces the provider list to a navigable size.

### 6. Data-layer tests only

Unit tests with JSON fixtures cover: parsing `models.json` (including the `reasoning_options` effort extraction), filtering by `auth.json` keys + `tool_call`, the fallbacks (missing/empty auth, missing/corrupt catalog), and the assignment patch as it lands in `opencode.json` through the deep-merge — including the "unassigned agents write nothing" and "empty variant overwrites stale effort" cases. Picker navigation reducers are deliberately not covered; they carry no invariants that a test would protect. Test command stays `go test ./...`.

## Risks / Trade-offs

- **`models.json` schema drifts** (OpenCode renames `reasoning_options` or `tool_call`) → parsing is tolerant: unknown fields are ignored and a model that fails to yield effort values is simply offered without effort. A total parse failure degrades to "step not offered", never to a failed install.
- **Cache is stale or absent on a fresh machine** (user has never launched OpenCode) → step is silently skipped; the user gets today's behavior and can rerun the installer after OpenCode populates the cache.
- **User assigns a model they are not actually entitled to** (auth.json key present but the specific model unavailable on their plan) → the installer cannot verify entitlement; the failure surfaces at agent run time in OpenCode, not at install time. Accepted: the same is true of any hand-written `opencode.json`.
- **Removing `variant` from the markdown changes behavior for existing users who never open the new step** → they fall back to their global OpenCode default instead of the previously hardcoded `high`/`xhigh`. This is the intended semantics ("no choice = inherit"), and the wizard step is where they express a preference. Worth calling out in the release notes.
- **Reading `auth.json` at all** → mitigated by consuming keys only and never letting a value escape the parsing function; a bug here would be a credential leak into the TUI or a plan line.
- **Extra wizard step lengthens the happy path** → the step is skippable with a single key and is hidden entirely when agents are not being installed.

## Migration Plan

No data migration. Existing `opencode.json` files are only added to, never rewritten, and the installer already writes a backup before any update. Rollback is reverting the release: the agent markdown regains its `variant`, and any `agent.<name>.model`/`variant` keys already written to `opencode.json` remain valid OpenCode configuration (they continue to work, and take precedence in the same way).

## Open Questions

None — the product and technical decisions are fixed by the confirmed brief.
