## 1. Presentation Contracts

- [x] 1.1 Add table-driven TUI tests for the complete gray, green, yellow, magenta, and red mapping across confirmation actions, completion actions, and completion/error status labels.
- [x] 1.2 Add tests proving ANSI resets at the exact label boundary, paths/details/separator spaces retain normal color, unknown or near-match labels remain unstyled, and stripping ANSI preserves the original text exactly.
- [x] 1.3 Add confirmation and completion regressions proving styled long lines wrap on the same plain-text boundaries and preserve scrolling, resizing, visible order, and range feedback.

## 2. TUI-Only Action Styling

- [x] 2.1 Define the ANSI-256 Lip Gloss action styles and an exact recognized-prefix table in the TUI layer, without changing installer plan or report producers.
- [x] 2.2 Carry recognized plain-text label spans through visual-row wrapping and apply color only while rendering visible confirmation and completion rows.
- [x] 2.3 Render completion and error status labels with the semantic styles while leaving error details normal and preserving all existing wording, capitalization, alignment, and spacing.

## 3. Validation

- [x] 3.1 Run the focused `internal/tui` tests and confirm the action-color, boundary, wrapping, scrolling, and resize contracts pass.
- [x] 3.2 Run `go test ./...` and confirm the complete Go suite passes.
