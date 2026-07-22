## Why

Installer action lists are currently visually uniform, making it harder to distinguish harmless, mutating, replacement, and error outcomes at a glance. Risk-semantic color on action labels will improve scanability in both confirmation and completion views without changing installer behavior or output wording.

## What Changes

- Color only recognized action-label prefixes in the installation plan preview and final operation report.
- Use the existing ANSI-256 Lip Gloss palette semantics: gray for no change, green for create/install/completed, yellow for update, magenta for replace/backup, and red for errors.
- Keep paths, details, wording, capitalization, alignment, spacing, and unknown labels unchanged and in the normal terminal color.
- Preserve plans and reports as plain text, including their existing wrapping, scrolling, ordering, and width calculations; add ANSI styling only while rendering the TUI.
- Add coverage for every supported label mapping, the exact label-to-path boundary, and wrapping/scrolling behavior with styled output.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `installer-output-presentation`: Add risk-semantic action-label coloring to complete installation previews and results while preserving their plain-text content, dimensions, and navigation behavior.

## Impact

- Affects only the Bubble Tea installer presentation layer and its TUI tests, primarily `internal/tui/wizard.go` and `internal/tui/wizard_test.go`.
- Does not change installation planning, application, action text generation, operation ordering, APIs, dependencies, or the separate `add-direct-execution-route` change.
