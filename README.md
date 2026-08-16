
![Angel AI OpenCode interface](docs/images/angel-ai-opencode-interface.png)

## Installation

The initial distribution supports **macOS on Apple Silicon** only
(`Darwin/arm64`). It does not require Go or cloning this repository. Install the
latest stable version with:

```sh
curl --proto '=https' --tlsv1.2 -fsSL https://raw.githubusercontent.com/Angel-M-R/angel-ai-opencode/main/install.sh | /bin/sh
```

The installer verifies the download and places the executable in
`~/.local/bin/angel-ai`.

```sh
angel-ai                       # opens the interactive wizard
angel-ai version               # shows the installed version without network access
angel-ai update                # forces an update check
```

## Design comparison

The comparison brings together tables and explanations covering agents,
planning, memory, specs, token savings, final review, and MCPs across eight
projects.

**[Read the full design comparison](docs/harness-comparison.md)**

## What it does

Angel AI creates or updates the selected files under `~/.config/opencode/`.
When a managed file already exists and its contents change, the installer first
creates a timestamped backup. Agent, skill, theme, and plugin assets are updated
file by file, so files not managed by Angel AI remain untouched. `AGENTS.md` is
the only full replacement; `opencode.json` and `tui.json` are merged with the
existing configuration.

| File modified | What it is |
|---|---|
| `~/.config/opencode/agents/*.md` | Selected [agent definitions](assets/agents/) are created or replaced. Each file contains YAML frontmatter and a system prompt. |
| `~/.config/opencode/skills/<skill>/**` | Selected [skills](assets/skills/) are updated recursively. Additional files already present in the destination are preserved. |
| `~/.config/opencode/plugins/cmux-*.js` | The [cmux session and feed plugins](assets/integrations/cmux/) are created or replaced when the cmux integration is selected. |
| `~/.config/opencode/themes/*.json` | Selected [themes](assets/themes/) are created or replaced. |
| `~/.config/opencode/tui-plugins/*` | The selected [Angel AI TUI plugins](assets/tui-plugins/) are created or replaced. |
| `~/.config/opencode/AGENTS.md` | The existing file is fully replaced with the [global Angel AI rules](assets/agents-md/AGENTS.md), plus the [CodeGraph guidance](assets/integrations/codegraph/AGENTS.md) when selected. |
| `~/.config/opencode/opencode.json` | The [MCP](assets/fragments/mcp.json), [permission](assets/fragments/permissions.json), and [settings](assets/fragments/settings.json) fragments are deep-merged into the existing configuration. Selected agent models, CodeGraph, and tsgo settings are also reconciled without removing unrelated keys. |
| `~/.config/opencode/tui.json` | Theme and plugin selections are merged into the existing TUI configuration, including the [one-dark-pro theme](assets/themes/one-dark-pro.json) and selected [TUI plugins](assets/tui-plugins/), without removing unrelated keys. |

The OpenSpec extra installs or updates only the official OpenSpec CLI. When an
OpenSpec workflow is first used in a project, Angel AI runs
`openspec init --tools opencode`; for an existing initialized project, it runs
`openspec update`. OpenSpec owns the resulting project-local `.opencode/` files
and generates them from the user's current global profile, workflow selection,
and delivery mode. Angel AI does not ship copies of those skills.

## Usage from the repository

```sh
go run .                  # opens the wizard
go run . --all            # installs everything without the TUI
go run . --all --dry-run  # shows the plan without changing anything
go run . --target /path   # installs in another directory (for testing)
```
