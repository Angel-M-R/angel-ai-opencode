## ADDED Requirements

### Requirement: Model catalog is read from the OpenCode cache
The installer SHALL build its list of selectable models by reading OpenCode's own model cache at `~/.cache/opencode/models.json`. The file is a JSON object keyed by provider id; each provider object contains a `models` object keyed by model id. A model's available reasoning effort levels SHALL be taken from the entry in its `reasoning_options` array whose `type` is `"effort"`, using that entry's `values` array in the order given. A model with no such entry SHALL be treated as having no effort levels.

#### Scenario: Catalog parsed from cache
- **WHEN** `models.json` exists and parses as a provider-keyed object
- **THEN** the installer exposes each provider with its models, and each model carries the effort values from its `reasoning_options` entry of `type: "effort"`

#### Scenario: Model without effort reasoning options
- **WHEN** a model has no `reasoning_options` entry of `type: "effort"`
- **THEN** that model is offered with an empty list of effort levels and the effort step is not presented for it

#### Scenario: Catalog missing or unparseable
- **WHEN** `~/.cache/opencode/models.json` does not exist, cannot be read, or is not valid JSON
- **THEN** the installer SHALL treat the catalog as unavailable, SHALL NOT fail the installation, and SHALL continue with every other installer step unchanged

### Requirement: Catalog is filtered by authentication and tool-calling support
The installer SHALL offer only models that the user can realistically run as an agent. Providers SHALL be restricted to those whose id appears as a top-level key of `~/.local/share/opencode/auth.json`. Models SHALL be restricted to those with `tool_call: true`. The installer SHALL read only the top-level keys of `auth.json` and SHALL NOT read, retain, display, or log any credential value it contains.

#### Scenario: Only authenticated providers are offered
- **WHEN** `auth.json` contains provider keys `anthropic` and `openai`, and the catalog contains many more providers
- **THEN** only `anthropic` and `openai` are offered in the provider list

#### Scenario: Non-tool-calling models are excluded
- **WHEN** an authenticated provider contains a model with `tool_call: false`
- **THEN** that model is not offered for assignment

#### Scenario: Auth file missing or empty
- **WHEN** `~/.local/share/opencode/auth.json` does not exist, is unreadable, or contains no provider keys
- **THEN** the installer SHALL fall back to offering every provider in the catalog, still filtered to `tool_call: true` models

#### Scenario: Provider with no eligible models
- **WHEN** filtering leaves a provider with zero models
- **THEN** that provider is not offered in the provider list

### Requirement: Per-agent model and effort assignment in the wizard
The installer wizard SHALL present a step, positioned after the extras step and before the confirmation step, in which the user assigns a model and a reasoning effort to each Angel AI agent individually. The configurable agents SHALL be exactly the seven shipped under `assets/agents`: `angel-orchestrator`, `openspec-planner`, `openspec-implementer`, `openspec-verifier`, `review-correctness`, `review-security-risk`, `review-simplicity`. The step SHALL flow agent list -> provider -> model -> effort, SHALL offer a skip action on each level, and SHALL NOT offer an "apply to all agents" action.

#### Scenario: Assigning a model and effort to one agent
- **WHEN** the user picks `openspec-implementer`, then a provider, then a model that has effort levels, then an effort level
- **THEN** that agent shows the chosen `provider/model` and effort in the agent list, and the other six agents remain unassigned

#### Scenario: Model without effort levels skips the effort level
- **WHEN** the user picks a model whose effort level list is empty
- **THEN** the effort step is not presented and the agent is recorded as assigned with an empty effort

#### Scenario: Skipping the step
- **WHEN** the user chooses to skip the step without assigning any agent
- **THEN** no agent is recorded as assigned and the wizard advances to confirmation

#### Scenario: Step hidden when unavailable
- **WHEN** the `agents` asset category is not selected, or the model catalog is unavailable
- **THEN** the step SHALL NOT be shown and the wizard SHALL keep the pre-existing step sequence and step numbering

### Requirement: Existing assignments are preloaded on reinstall
When the wizard opens the assignment step, it SHALL read the existing `agent.<name>.model` and `agent.<name>.variant` values from `~/.config/opencode/opencode.json` and use them as the initial selection for the corresponding agents.

#### Scenario: Reinstall preloads prior choices
- **WHEN** `opencode.json` already contains `agent.angel-orchestrator.model` set to `anthropic/claude-sonnet-4-6` and `variant` set to `high`
- **THEN** the assignment step opens with `angel-orchestrator` shown as assigned to that model and effort

#### Scenario: Fresh install has no preloaded assignments
- **WHEN** `opencode.json` does not exist or contains no `agent` entries for the seven agents
- **THEN** all seven agents open as unassigned

### Requirement: Assignments are written to opencode.json, and only when chosen
For each agent the user assigned, the installer SHALL write both `agent.<name>.model` as the string `"<provider>/<model>"` and `agent.<name>.variant`, using an empty string for `variant` when the chosen model has no effort levels. Both keys SHALL always be written together so a stale effort cannot survive a model change. The write SHALL go through the installer's existing `opencode.json` JSON deep-merge, leaving unrelated configuration untouched. For any agent the user did not assign, the installer SHALL write neither key, leaving the agent to inherit the user's global OpenCode default.

#### Scenario: Assigned agent gets model and variant
- **WHEN** the user assigns `review-simplicity` to `anthropic/claude-sonnet-4-6` with effort `high`
- **THEN** `opencode.json` contains `agent."review-simplicity".model` = `"anthropic/claude-sonnet-4-6"` and `agent."review-simplicity".variant` = `"high"`

#### Scenario: Assigned model without effort clears the variant
- **WHEN** an agent previously had `variant` set to `high` and the user reassigns it to a model with no effort levels
- **THEN** `agent.<name>.variant` is written as `""` so the previous effort does not survive

#### Scenario: Unassigned agents are not written
- **WHEN** the user assigns only `openspec-planner` and leaves the other six unassigned
- **THEN** `opencode.json` gains `model` and `variant` only under `agent."openspec-planner"`, and no key is added or removed for the other six

#### Scenario: Non-interactive install writes nothing
- **WHEN** the installer runs with `--all`
- **THEN** no `agent.<name>.model` or `agent.<name>.variant` key is written by this feature, and the resulting configuration is byte-identical to the pre-change `--all` result apart from unrelated changes

#### Scenario: Unrelated configuration is preserved
- **WHEN** `opencode.json` already contains other top-level keys and other `agent` entries
- **THEN** the deep-merge preserves them and only adds or replaces `model` and `variant` under the assigned Angel AI agents

### Requirement: Agent markdown no longer declares a variant
The seven agent markdown files under `assets/agents` SHALL NOT declare a `variant` key in their frontmatter, so that `opencode.json` is the single source of truth for model and reasoning effort.

#### Scenario: Installed agent files carry no variant
- **WHEN** the agent assets are installed
- **THEN** none of the seven installed agent markdown files contains a `variant` frontmatter key

#### Scenario: Unassigned agent falls back to the global default
- **WHEN** an agent has no `variant` in its markdown and no `agent.<name>` entry in `opencode.json`
- **THEN** OpenCode runs it with the user's global default model and reasoning effort
