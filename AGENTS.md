# Angel AI OpenCode — project guidelines

## Testing agent prompt assets

The Markdown files under `assets/agents/` and `assets/agents-md/` are prompt
prose. Never write tests that pin their wording: no `strings.Contains` on
sentences or phrases, no required-phrase lists, no section-order index checks.
Such tests freeze the phrasing instead of the behavior, so every prompt edit
becomes a test-fixing chore without catching real regressions.

Test functional contracts instead:

- **Install behavior** — assets are cataloged and installed byte-identical
  (`TestAgentAssetsAreCatalogedAndInstalledUnchanged`,
  `TestExistingOrchestratorCopyRequiresUpdatedSourceAndSelectedReconciliation`).
- **Structured data the code consumes** — frontmatter fields the harness reads
  (`TestAgentFrontmatterRemainsStructurallySafe`), embedded examples decoded
  strictly against the real Go types
  (`TestOpenSpecVerifierCompletionExampleMatchesServiceSchema`).
- **References validated against a registry** — every `openspec-*` name in an
  agent asset must be an official core skill or an Angel worker agent
  (`TestAgentAssetsReferenceOnlyOfficialOpenSpecNames`).

If a prompt rule feels worth guarding, extract the machine-checkable part — a
name, a schema, a config value, an installer behavior — and test that. Never
the sentence.
