## 1. Shared Global CLI Foundation

- [x] 1.1 Add an internal global-CLI descriptor and injectable executable/command seams for package-manager discovery, version checks, package execution, and post-install executable verification.
- [x] 1.2 Implement npm-first manager selection with validated `pnpm bin -g` fallback, no ownership detection, and no fallback after a selected npm command fails.
- [x] 1.3 Implement OpenSpec's pre-install Node.js semantic-version check for `>=20.19.0`, including missing and malformed version diagnostics.

## 2. Installer Planning and Application

- [x] 2.1 Register the CLI-only `OpenSpec` extra for `@fission-ai/openspec@latest`, keep it default-selected and included by `--all`, and retain all 12 vendored OpenSpec skills for the later bundle-layout migration.
- [x] 2.2 Route CodeGraph and OpenSpec through the shared descriptor list while keeping CodeGraph MCP and `AGENTS.md` reconciliation in the CodeGraph-specific preparation path.
- [x] 2.3 Update `PlanInstallation` to run non-mutating manager and Node.js preflight and report an `@latest` action for every selected CLI regardless of executable presence.
- [x] 2.4 Update `ApplyInstallation` to run selected CLI commands sequentially before file writes, retain partial success reporting on failure, verify installed executables, and reprepare exactly once after all CLI actions succeed.

## 3. Installer Behavior Tests

- [x] 3.1 Add `PlanInstallation` tests for npm preference, valid and invalid pnpm fallback, supported and unsupported Node.js versions, dry-run non-mutation, and selected CLIs that are already present.
- [x] 3.2 Add `ApplyInstallation` tests for deterministic `@latest` command order, existing-CLI updates, post-install executable checks, partial CLI failure with zero configuration writes, and exactly one post-success repreparation.
- [x] 3.3 Extend extra and CodeGraph regression tests to prove OpenSpec selection changes no OpenCode configuration, CodeGraph retains its MCP/`AGENTS.md` behavior, and the vendored OpenSpec skill set remains intact.

## 4. OpenSpec Worker Bootstrap

- [x] 4.1 Update the orchestrator asset to run a bounded `general` bootstrap before the first OpenSpec worker per project/store context in a session and to repeat it only after that context changes.
- [x] 4.2 Define the bootstrap prompt around `openspec list --json`, store-aware flags, and `openspec init --tools none` only for an unresolved local root, with a follow-up JSON readiness check.
- [x] 4.3 Add missing-CLI blocking guidance, advisory comparison of `openspec --version` with global skill `metadata.generatedBy`, silent local-duplicate handling, and explicit prohibitions on `openspec update` and profile/workflow/delivery changes.
- [x] 4.4 Add agent-asset contract tests for bootstrap ordering, per-context reuse, store handling, initialization limits, version mismatch continuation, and preserved vendored skills.

## 5. OpenSpec Skill Bundle Migration

- [x] 5.1 Move all 12 OpenSpec skill directories to `assets/skills/openspec/<skill-name>/SKILL.md` without changing their individual definitions.
- [x] 5.2 Ensure the catalog exposes one `openspec` `CopyDir` bundle and recursive installation preserves every nested child path under `<config>/skills/openspec/`.
- [x] 5.3 Add focused tests that assert grouped catalog discovery and recursive copying of nested OpenSpec child skills.
- [x] 5.4 Record the previously completed legacy skill-registry regeneration as implementation history only; the registry is no longer part of the target architecture or acceptance criteria.
- [x] 5.5 Update the completed worker bootstrap's version lookup and contract coverage to use child `metadata.generatedBy` values under the global `openspec` bundle.
- [x] 5.6 Complete the prior nested-layout migration by installing and verifying all 12 child skills under `~/.config/opencode/skills/openspec/`, then deleting the old `~/.config/opencode/skills/openspec-*` directories without installer cleanup logic.
- [x] 5.7 Complete the prior OpenCode restart so it loads the nested skill layout.

## 6. Validation

- [x] 6.1 Run formatting and focused catalog, recursive installer, and agent-asset tests, then run the complete repository test suite and resolve any regressions without generating local OpenSpec skills or performing additional real global package installs.

## 7. Evidence-Based Gentle AI Cleanup

- [x] 7.1 Inventory the confirmed Gentle AI-owned global footprint against `/Users/angelmr/Desktop/Github_Repos/gentle-ai` read-only, then create one timestamped backup containing every file or directory that will be removed or edited.
- [x] 7.2 Remove exactly the 21 obsolete skill directories and the confirmed Gentle AI `gentle-orchestrator`/`sdd-*` agent entries, SDD and registry commands, `prompts/sdd`, legacy plugins, global `.atl`, and ownership marker without modifying protected or unattributed files.
- [x] 7.3 Validate the cleaned `opencode.json` and prove `cognitive-doc-design`, `technical-grilling`, `investigate`, `product-grilling`, and all 12 nested OpenSpec skills remain byte-identical to their pre-cleanup copies.
- [x] 7.4 Re-run OpenSpec change verification after the obsolete registry requirement is removed, then restart OpenCode so the deleted skills and agents are unloaded.
