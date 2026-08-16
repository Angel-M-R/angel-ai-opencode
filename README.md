
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

## What it edits

All content lives in `assets/` and is edited by hand—the Go code does not need
to be changed to update content. The following is written to the target machine,
grouped by destination:

| Installed in | What it is |
|---|---|
| `~/.config/opencode/agents/` | Agents (one file per agent: YAML frontmatter + system prompt) |
| `~/.config/opencode/commands/` | Slash commands (supported destination; none are currently shipped) |
| `~/.config/opencode/skills/` | Skills (preserves extra files at the destination) |
| `~/.config/opencode/plugins/` | JS/TS plugins (currently only cmux plugins, if the integration is selected) |
| `~/.config/opencode/themes/` | Themes |
| `~/.config/opencode/tui-plugins/` | TUI plugins (logo, etc.), enabled through the three toggles in the wizard's final step |
| `~/.config/opencode/AGENTS.md` | Global behavior rules (+ CodeGraph rules if selected) |
| `~/.config/opencode/opencode.json` | MCPs, permissions, and settings merged into the existing file (+ CodeGraph configuration if selected) |

## Usage from the repository

```sh
go run .                  # opens the wizard
go run . --all            # installs everything without the TUI
go run . --all --dry-run  # shows the plan without changing anything
go run . --target /path   # installs in another directory (for testing)
```
