## 1. Extra Definitions and TUI Patches

- [x] 1.1 Add independent default-selected `opencode-open-in-app` and `opencode-openspec-task-tui` options immediately after Subagent statusline in `internal/install/extras.go`, using the unversioned package names.
- [x] 1.2 Extend TUI extra preparation in `internal/install/installation.go` so each selected option contributes only its own npm plugin patch and a deselected option contributes no removal behavior.

## 2. TUI Plugin Identity and Migration

- [x] 2.1 Refactor plugin-array reconciliation in `internal/install/install.go` to accept an injectable identity resolver while retaining the existing pure resolver as the unchanged default for `opencode.json`.
- [x] 2.2 Implement the TUI-scoped resolver for absolute, relative, and `file:` entries: read bundles without executing them, map only literal IDs `opencode-open-in-app` and `openspec-task-progress` to their canonical npm identities, and fall back to pure identity for unreadable or unrecognized entries.
- [x] 2.3 Wire the TUI merge to the scoped resolver so selected plugins replace recognized equivalents at the first matching position, collapse recognized duplicates to one unversioned npm entry, preserve unrelated entries, and remain idempotent.

## 3. Installer TUI Notice

- [x] 3.1 Render `Si ya está instalado, desmarcarlo no lo desinstalará` exactly once as a global notice on the Integrations and extras screen in `internal/tui/wizard.go`.

## 4. Automated Coverage

- [x] 4.1 Extend extras selection coverage to verify both new options are adjacent after Subagent statusline, selected by default, and independently selectable.
- [x] 4.2 Add `TestApplyMigratesPublishedTUIPlugins` as an installer integration test covering absolute, relative, and `file:` recognized bundles, first-position replacement, duplicate collapse, unversioned npm output, preservation of unrelated or unrecognized entries, and non-destructive deselection.
- [x] 4.3 Add `TestExtrasViewShowsNonUninstallNotice` as a TUI render test that verifies the exact global notice appears once.
- [x] 4.4 Add `TestApplyAppendsSelectedPublishedTUIPluginWhenAbsent` as a dedicated installer integration test proving that selecting a published plugin with no equivalent local bundle appends its canonical unversioned npm entry while preserving all unrelated plugin entries.
- [x] 4.5 Add `TestTUIPluginIdentityResolverDoesNotExecuteBundle` as a focused side-effect sentinel test proving that the TUI identity resolver recognizes a local bundle by reading its source while never executing the bundle.

## 5. Final Verification
<!-- owner: openspec-verifier -->

- [x] 5.1 Run `go test ./internal/install ./internal/tui -count=1` and report exit code 0.
- [x] 5.2 Run `go test ./internal/install -run '^TestApplyMigratesPublishedTUIPlugins$' -count=1 -v` and report passing executed evidence that the resulting `tui.json` replaces recognized bundles at the first position, preserves unrecognized entries, and collapses duplicates.
- [x] 5.3 Run `go test ./internal/tui -run '^TestExtrasViewShowsNonUninstallNotice$' -count=1 -v` and report passing executed evidence that the rendered extras view contains `Si ya está instalado, desmarcarlo no lo desinstalará` exactly once.
- [x] 5.4 Run `go test ./internal/install -run '^TestApplyAppendsSelectedPublishedTUIPluginWhenAbsent$' -count=1 -v` and report passing executed evidence that the canonical unversioned npm entry is appended when no equivalent local bundle exists and unrelated entries remain intact.
- [x] 5.5 Run `go test ./internal/install -run '^TestTUIPluginIdentityResolverDoesNotExecuteBundle$' -count=1 -v` and report passing executed evidence from the side-effect sentinel that recognized local bundle source is read without being executed.
