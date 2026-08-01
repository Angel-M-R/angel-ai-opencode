## Why

The installer currently leaves users to configure local development bundles for the Open-in-App and OpenSpec task sidebar integrations. Adding their published npm packages as installer extras provides a stable setup path while safely migrating recognized local equivalents without disturbing unrelated TUI plugins.

## What Changes

- Add `opencode-open-in-app` and `opencode-openspec-task-tui` as independent, default-selected installer extras immediately after Subagent statusline.
- When either extra is selected, add its unversioned npm package to `tui.json` and replace recognized equivalent local bundles while preserving unrelated, unreadable, and unrecognized plugin entries.
- Recognize absolute, relative, and `file:` local bundle entries by reading, but not executing, the bundle and matching its literal plugin `id` property.
- Collapse duplicate recognized local or npm entries to one npm entry at the first matching position.
- Preserve existing npm and local entries when an extra is deselected; deselection is not an uninstall operation.
- Display the global extras notice `Si ya está instalado, desmarcarlo no lo desinstalará` in the installer TUI.
- Add installer migration coverage and TUI rendering coverage for the new behavior.

## Capabilities

### New Capabilities
- `published-tui-plugin-extras`: Defines selection, safe local-bundle recognition and migration, duplicate handling, non-destructive deselection, and installer notice behavior for the two published TUI plugins.

### Modified Capabilities

None.

## Impact

- Installer extra definitions and TUI configuration preparation in `internal/install/extras.go` and `internal/install/installation.go`.
- Plugin-array merge behavior in `internal/install/install.go`, retaining pure identity matching for `opencode.json` while permitting TUI-specific local bundle inspection.
- Installer integration tests in `internal/install` and extras-screen rendering tests in `internal/tui`.
- No published plugin package, npm release, installed OpenCode configuration outside the requested installer target, or unrelated installer behavior is changed.
