## 1. Output Windowing Tests

- [x] 1.1 Add focused wizard tests for list-window sizing, offset clamping, resize behavior, and first/page/last navigation over ordered synthetic entries.
- [x] 1.2 Add confirmation-view tests proving short and overflowing plans expose every entry in order, never render an omission summary, and retain install, back, and quit controls.
- [x] 1.3 Add completion-view tests proving successful and partial-error reports expose every returned operation in order, preserve the error heading, and use explicit navigation and exit behavior.

## 2. Terminal-Aware Presentation

- [x] 2.1 Track terminal dimensions and independent confirmation/result scroll offsets in the wizard model, handling window-size messages and clamping offsets as content or height changes.
- [x] 2.2 Replace the fixed-limit truncation helper with local ordered list-window helpers that reserve fixed UI chrome, return a contiguous visible range, and provide range/total feedback without hiding data or adding a runtime dependency.
- [x] 2.3 Render the complete installation plan through the confirmation list window and add non-conflicting line, page, and boundary navigation while preserving existing planning errors and confirm/back/quit actions.
- [x] 2.4 Render the complete operation report through the completion list window, including all partial results before an error, and update help/key handling so scrolling and explicit exit remain available.

## 3. Verification

- [x] 3.1 Run the focused TUI tests and verify both fitting and overflowing terminal-height cases contain no `… and N more` output.
- [x] 3.2 Run the full Go test suite and confirm installer planning, application, operation order, and error behavior remain unchanged.
