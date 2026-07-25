## 1. Model catalog data layer

- [x] 1.1 Add a models source in `internal/install/` that resolves `~/.cache/opencode/models.json` (with an injectable path for tests) and parses it into provider -> models, keeping per model: id, display name, `tool_call`, and the effort values from the `reasoning_options` entry whose `type` is `"effort"`.
- [x] 1.2 Make the parser tolerant: unknown fields ignored; a model with no effort entry yields an empty effort list; a missing, unreadable, or invalid-JSON file returns "catalog unavailable" without an error that could fail the install.
- [x] 1.3 Add auth-based filtering that reads only the top-level keys of `~/.local/share/opencode/auth.json` (injectable path), discards all values, keeps only providers present in those keys, and keeps only models with `tool_call: true`.
- [x] 1.4 Implement the fallback: when `auth.json` is missing, unreadable, or has no keys, keep all providers (still `tool_call`-filtered). Drop providers left with zero models. Sort providers and models deterministically for stable picker order.
- [x] 1.5 Add unit tests with JSON fixtures covering 1.1–1.4: effort extraction, `tool_call` exclusion, auth filtering, both fallbacks, empty-provider pruning, and corrupt-catalog degradation.

## 2. Assignment write path

- [x] 2.1 Define the assignment type (agent name -> `{model string; variant string}`) and the fixed list of the seven configurable agents: `angel-orchestrator`, `openspec-planner`, `openspec-implementer`, `openspec-verifier`, `review-correctness`, `review-security-risk`, `review-simplicity`.
- [x] 2.2 Add an `AgentModels` field to `install.InstallationRequest` so plan and apply consume the identical assignment set.
- [x] 2.3 Build the assignment into a `map[string]any` patch of shape `{"agent": {"<name>": {"model": "<provider>/<model>", "variant": "<effort|empty string>"}}}`, emitting an entry only for assigned agents and always emitting both keys together.
- [x] 2.4 Append that patch to the existing `fragments` slice in `prepareInstallation` so it flows through the current `opencode.json` deep-merge, backup, and plan/apply reporting. Nil/empty assignments must produce no patch at all.
- [x] 2.5 Add a reader that loads existing `agent.<name>.model` / `agent.<name>.variant` from `opencode.json` for the seven agents, splitting `model` back into provider and model id, and tolerating a missing or malformed file by returning no assignments.
- [x] 2.6 Add unit tests for 2.3–2.5: assigned agent writes both keys; model without effort writes `variant: ""` and overwrites a stale effort; unassigned agents add and remove nothing; unrelated `opencode.json` keys and non-Angel `agent` entries survive the merge; empty assignments leave the config byte-identical to the pre-change result; preload round-trips an existing config.

## 3. Wizard step

- [x] 3.1 Add a new phase constant after `extrasPhase` in `internal/tui/wizard.go` and wire it into `Update`'s phase switch and the `extrasPhase` -> confirmation transition.
- [x] 3.2 On entering the step, load the filtered catalog and preload existing assignments from `opencode.json` as the initial selection; if the catalog is unavailable or the `agents` category is not selected, skip the phase entirely and keep the existing sequence and `totalSteps()` numbering unchanged.
- [x] 3.3 Implement the drill-down state: agent list -> provider list -> model list -> effort list, with `s` to skip at each level, `left/h/b` to go back consistently with existing phases, and the effort level bypassed for models with no effort values.
- [x] 3.4 Render the step: agent rows showing the currently assigned `provider/model` and effort (or "sin asignar"), plus a help footer matching the existing style and language of the other phases.
- [x] 3.5 Include the assignments in `installationPlan()` and `installCmd(...)` via `InstallationRequest`, and add an assignment count line to the confirmation summary next to the extras count.

## 4. Assets and non-interactive path

- [x] 4.1 Remove the `variant:` frontmatter line from all seven files in `assets/agents/`.
- [x] 4.2 Confirm `main.go`'s `--all` path passes no assignments, so a non-interactive install writes neither `model` nor `variant`.
- [x] 4.3 Update `internal/install/agent_assets_test.go` (and any other test asserting on agent frontmatter) so it asserts the absence of `variant` rather than its value.

## 5. Verification

- [x] 5.1 Run `go test ./...` and confirm it passes.
- [x] 5.2 Run `go run . --all --dry-run --target <temp dir>` and confirm the plan contains no `agent.*.model` / `agent.*.variant` changes.
- [ ] 5.3 Manually walk the wizard against a temp `--target`: assign two agents, skip the rest, install, and confirm `opencode.json` contains exactly those two agents' `model` + `variant` and nothing else new; then rerun and confirm the step preloads those assignments.
