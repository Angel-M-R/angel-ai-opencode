## Context

The installer already exposes hardcoded extras and merges selected TUI plugins into `tui.json`. Plugin arrays are currently reconciled by a pure string identity function that normalizes npm versions, which is appropriate for `opencode.json` but cannot identify local bundles that are equivalent to the two published plugins. The change spans extra definitions, TUI-specific merge preparation, merge identity resolution, and the extras-screen presentation.

The two published packages already provide the intended sidebar slot orders: `opencode-open-in-app` uses 89, the existing Angel logo uses 100, and `opencode-openspec-task-tui` uses 350. The installer therefore does not need to derive visual order from the `tui.json` plugin array.

## Goals / Non-Goals

**Goals:**
- Offer both published packages as independent default-selected extras in the confirmed list position.
- Reconcile selected packages into `tui.json` while migrating recognized local equivalents safely and idempotently.
- Preserve the current pure plugin identity behavior for `opencode.json` and allow only TUI merging to inspect local bundles.
- Make the non-uninstall behavior of deselection visible in the extras screen.
- Cover migration behavior with an installer integration test and the notice with a TUI render test.

**Non-Goals:**
- Modifying either published plugin, its slot order, or its npm release.
- Treating `tui.json` array position as the source of visual sidebar order.
- Removing npm or local plugin entries when an extra is deselected.
- Changing local installed OpenCode configuration except through a requested installer plan or apply operation.
- Refactoring unrelated installer extras or merge behavior.

## Decisions

### Add two explicit extra descriptors

Add independent descriptors for `opencode-open-in-app` and `opencode-openspec-task-tui` immediately after Subagent statusline, both selected by default. Their selected patches use the unversioned npm package names. This follows the existing hardcoded-extra model and keeps selection independent; combining them under one toggle would prevent users from choosing only one integration.

### Inject plugin identity resolution at the array-merge seam

Allow plugin-array reconciliation to receive an identity resolver. The existing default resolver remains the current pure npm/path identity function, and `opencode.json` continues to use it unchanged. TUI preparation supplies a resolver scoped to the destination configuration directory so that only `tui.json` merging performs filesystem inspection.

The alternative of embedding filesystem reads in the global identity function is rejected because it would make ordinary `opencode.json` merging impure and broaden migration behavior beyond this change.

### Canonicalize recognized local bundles to published package identities

For string entries that represent absolute paths, relative paths, or `file:` paths, the TUI resolver resolves the bundle path relative to the TUI configuration directory when necessary and reads the file as data. It does not load, import, or execute the bundle. A literal plugin `id` property of `opencode-open-in-app` maps to the `opencode-open-in-app` package identity; a literal `id` of `openspec-task-progress` maps to the `opencode-openspec-task-tui` package identity.

If path resolution, reading, or literal-id recognition fails, the resolver falls back to the existing pure identity. This preserves unreadable and unrecognized entries instead of guessing from directory or file names. Matching only the two exact literal IDs avoids replacing unrelated local code.

The alternative of matching checkout names or path substrings is rejected because users can rename directories and unrelated bundles can contain similar names.

### Reuse first-match replacement semantics for migration and deduplication

The selected npm package and every recognized equivalent resolve to one canonical identity. Existing merge behavior then emits the unversioned npm entry at the first matching existing position, suppresses later matches, and appends the npm entry only when no match exists. Unrelated entries retain their existing relative order. This provides deterministic, idempotent migration without relying on array order for visual placement.

### Keep deselection non-destructive and explain it globally

TUI plugin patches are created only for selected extras. A false selection does not create a removal instruction, so existing npm and local entries survive subsequent runs. The extras view renders `Si ya está instalado, desmarcarlo no lo desinstalará` once as a global notice rather than repeating it per option.

## Risks / Trade-offs

- [A bundle computes its ID dynamically or uses a non-literal form] → Treat it as unrecognized and preserve it; safety takes precedence over aggressive migration.
- [A local path is unreadable or resolves outside the configuration directory] → Read only the explicitly configured path, never execute it, and preserve the original entry on failure.
- [Static literal matching produces a false positive inside unrelated source text] → Limit recognition to the literal `id` property form and the two exact IDs, with focused resolver tests.
- [Filesystem-aware identities make merging harder to reason about] → Inject the resolver only for TUI preparation and retain the pure resolver as the default everywhere else.

## Migration Plan

1. Ship the new default-selected extras and TUI-specific resolver with installer and render coverage.
2. On the next selected installer run, back up and reconcile `tui.json` through the existing file-reconciliation path.
3. Replace recognized local equivalents and duplicates with one unversioned npm entry while preserving unrelated entries.
4. For rollback, restore the installer-created `tui.json` backup; deselecting an extra alone intentionally does not remove installed entries.

## Open Questions

None. The package names, legacy IDs, list placement, selection defaults, migration rules, notice text, and validation command are confirmed.
