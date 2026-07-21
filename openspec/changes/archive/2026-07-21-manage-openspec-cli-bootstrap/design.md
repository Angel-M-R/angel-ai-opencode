## Context

The installer currently treats CodeGraph as a special case: it checks whether the executable is already present, installs only through npm when absent, then prepares the destination again before reconciling files. OpenSpec has no installer extra, although the repository vendors and globally installs 12 OpenSpec skills whose `metadata.generatedBy` value identifies the OpenSpec version that generated them. Those skills currently occupy separate top-level directories, and the catalog therefore presents them as separate wizard choices even though directory installation is already recursive. The orchestrator launches OpenSpec workers without a dedicated readiness step and assumes the CLI and a resolvable project or store already exist.

This change crosses installer planning/application, catalog and TUI selection, vendored skill layout, one-time global migration, CodeGraph-specific configuration, and orchestrator worker dispatch. It must preserve the 12 skill definitions and must not make Gentle AI a runtime or build dependency. The remaining global migration must also retire only files attributable to the previous Gentle AI installation while preserving Angel-managed and unrelated configuration.

## Goals / Non-Goals

**Goals:**

- Give CodeGraph and OpenSpec one testable mechanism for planning and executing global CLI updates.
- Make OpenSpec a preselected installer extra and include it in `--all`.
- Preserve all-or-nothing configuration writes even when more than one CLI is selected.
- Establish OpenSpec readiness once per project/store context in an orchestration session before launching an OpenSpec worker.
- Keep bootstrap behavior minimal, structured around OpenSpec CLI JSON output, and safe for registered stores.
- Present and install all OpenSpec skills as one recursively copied `openspec` bundle while retaining each child skill.
- Complete this machine's migration by backing up and removing the source-attributed Gentle AI global footprint, validating the retained Angel installation, and reloading OpenCode.

**Non-Goals:**

- Removing, regenerating, or replacing any of the 12 individual OpenSpec skill definitions.
- Detecting which package manager originally owns an installed CLI.
- Adding Gentle AI as a dependency or copying unrelated Gentle AI behavior.
- Running `openspec update` or changing OpenSpec profiles, workflows, delivery configuration, or local skill precedence.
- Making global package updates rollback-capable.
- Teaching the installer to detect, compare, back up, or remove legacy Gentle AI files.
- Modifying or deleting the Gentle AI source repository, the `gentle-ai` executable, Engram, Context7, permissions, Angel-managed agents, or files without ownership evidence.

## Decisions

### Represent selected global CLIs as data

Introduce an internal global-CLI descriptor used by installation planning and application. Each descriptor carries its display name, executable, and package pinned to `@latest`; OpenSpec additionally declares Node.js `20.19.0` as a minimum. A fixed descriptor order, aligned with the extras order, determines plan and execution order.

`preparedInstallation` will carry the selected descriptors rather than a CodeGraph-only installation flag. CodeGraph's MCP object and managed `AGENTS.md` block remain in the existing configuration preparation path and are not moved into the shared CLI module. The OpenSpec descriptor has no configuration-file side effects.

Alternative considered: keep separate `ensureCodegraphInstalled` and `ensureOpenSpecInstalled` functions. This duplicates manager selection, command execution, diagnostics, and test seams, and would allow their behavior to diverge again.

### Always request the latest selected CLI version

Planning will report an install/update action for every selected global CLI, even when its executable is already on `PATH`. Application will run the package manager for every selected descriptor and verify that its executable is available afterward. Existing executable presence affects diagnostics only; it never suppresses the `@latest` update. No ownership lookup is performed.

Alternative considered: skip a CLI already on `PATH` or compare registry and local versions. Presence does not prove freshness, while registry comparison adds network work and a second source of truth before the package manager performs the same resolution.

### Resolve one package manager with npm-first fallback

Preflight resolves npm first. If npm is unavailable, it resolves pnpm and accepts it only when `pnpm bin -g` succeeds and returns a non-empty global binary directory. If neither path is usable, planning and application fail before installation or file writes. Once npm is selected, an npm installation failure is reported directly rather than retried through pnpm. Commands are `npm install --global <package>@latest` or `pnpm add --global <package>@latest`.

This matches the verified npm-first/pnpm-fallback behavior from Gentle AI without introducing a dependency or ownership detection.

### Validate OpenSpec's Node requirement as a preflight

When OpenSpec is selected, preflight resolves Node.js, parses `node --version`, and requires a semantic version greater than or equal to `20.19.0`. This check runs before any selected CLI installation, including CodeGraph, so an invalid OpenSpec runtime cannot leave a partial global update. Planning performs the same non-mutating validation so `--dry-run` and the TUI confirmation surface the blocker early.

Alternative considered: rely on npm package engine enforcement. That would occur during installation, too late to guarantee that no earlier selected CLI had already been updated.

### Separate preparation, CLI side effects, and configuration writes

`ApplyInstallation` follows one ordered transaction boundary:

1. Prepare the full desired configuration and selected descriptor list without writing.
2. Complete package-manager and Node.js preflight.
3. Install/update selected CLIs sequentially, preserving successful result lines if a later CLI fails.
4. If every CLI succeeds, run `prepareInstallation` exactly once more.
5. Reconcile prepared files only after that repreparation succeeds.

If any CLI fails, no configuration file is written. A previously updated global CLI is an acknowledged external partial side effect and is reported; it is not rolled back. `PlanInstallation` performs preparation and preflight but runs no package installation or file reconciliation.

Alternative considered: reprepare after each CLI, following the current single-CodeGraph flow. That adds redundant reads and makes the number of preparations depend on selected extras.

### Add OpenSpec as a default-on, CLI-only extra

Add an `OpenSpec` entry to the fixed extras list. The TUI's existing default-on initialization will preselect it, and the existing `--all` map construction will include it. Selecting or deselecting OpenSpec affects only the global CLI descriptor; it does not add or remove local skills or OpenCode configuration.

### Make OpenSpec one recursively installed skill bundle

Move every existing OpenSpec skill directory beneath `assets/skills/openspec/`, producing `assets/skills/openspec/<skill-name>/SKILL.md`. The catalog intentionally scans only immediate children of `assets/skills`, so the new top-level `openspec` directory becomes one selectable `CopyDir` item in the wizard. The existing recursive directory preparation must preserve every nested relative path when copying the selected bundle to `<config>/skills/openspec/`.

The move changes grouping and destination paths, not the set or contents of the 12 child skills. A focused catalog test will assert that one `openspec` item replaces the 12 top-level entries, and a focused installation test will assert that recursive preparation copies representative nested child files to their corresponding destinations.

Alternative considered: add explicit grouping metadata while retaining top-level skill directories. That would add catalog concepts and special cases while leaving the global layout cluttered; the existing top-level catalog boundary and recursive `CopyDir` behavior already provide the desired bundle semantics.

### Keep evidence-based global cleanup outside the installer

The repository change installs only the nested destination and contains no migration code for prior global state. The one-time cleanup on this machine uses `/Users/angelmr/Desktop/Github_Repos/gentle-ai` read-only as ownership evidence. Its deletion allowlist contains exactly 21 obsolete Gentle AI skill directories: the 10 `sdd-*` skills, `skill-registry`, `_shared`, and `branch-pr`, `chained-pr`, `comment-writer`, `go-testing`, `issue-creation`, `judgment-day`, `skill-creator`, `skill-improver`, and `work-unit-commits`. It also covers the source-attributed `gentle-orchestrator` and `sdd-*` agent entries in `opencode.json`, SDD and registry commands, `prompts/sdd`, the identified legacy plugins, the global `.atl` directory, and the Gentle AI ownership marker.

Before deleting or editing any allowlisted item, copy the complete affected set to one timestamped backup. Then remove the allowlisted footprint, validate the resulting JSON, and compare the retained `cognitive-doc-design`, `technical-grilling`, `investigate`, `product-grilling`, and 12 nested OpenSpec skills byte-for-byte with their pre-cleanup copies. Do not modify or delete the source repository, the `gentle-ai` executable, Engram, Context7, permission configuration, Angel-managed agents, or any file without ownership evidence. Restart OpenCode only after validation succeeds so the retired skills and agents are unloaded.

Alternative considered: remove every global file not present in the current Angel payload. Absence is not ownership evidence and could destroy user-managed configuration, so cleanup is constrained to the confirmed source-attributed allowlist.

### Bootstrap through a bounded general worker before OpenSpec dispatch

The orchestrator will maintain an in-session set of bootstrapped OpenSpec contexts. Before the first planner, implementer, or verifier launch for a context, it dispatches one short `general` worker and waits for success. The context key is the selected store identifier when a store is explicit; otherwise it is the project root returned by OpenSpec. A changed project or store gets a new bootstrap, while subsequent OpenSpec workers in the same context and session skip it.

The bootstrap invokes `openspec list --json` as the source of truth and passes `--store <id>` when operating on a registered store. For a local context whose JSON result has no resolvable root, it runs `openspec init --tools none` and then rechecks with `openspec list --json`. It does not initialize a local project as a substitute for an explicit store. Missing CLI execution, failed initialization, or a still-unresolved root blocks worker launch.

The bootstrap also compares `openspec --version` with the `metadata.generatedBy` value in the child skills under the globally installed `openspec` bundle. A mismatch produces a warning but does not block. Duplicate project-local OpenSpec skills are neither diagnosed nor warned about because OpenCode's precedence is not guaranteed.

Alternative considered: put bootstrap logic in each OpenSpec worker. Central orchestration avoids repeated initialization, gives all worker types the same gate, and allows the user-facing session to retain the context cache.

## Risks / Trade-offs

- **A later CLI update can fail after an earlier global package was updated** → Preserve sequential result reporting, perform no configuration writes, and make rerunning the installer safe because every descriptor targets `@latest`.
- **Package-manager discovery can differ from shell ownership expectations** → Use deterministic npm-first resolution and explicitly avoid ownership inference.
- **`pnpm` can exist without a usable global bin setup** → Require a successful, non-empty `pnpm bin -g` before selecting it.
- **Node version output can be malformed or unsupported** → Treat missing or unparseable output as a preflight failure before side effects.
- **Prompt-managed bootstrap state exists only for the orchestration session** → Intentionally rerun in a new session; within a session, key the cache by resolved project/store context.
- **Global skills and CLI can drift** → Warn using child-skill `metadata.generatedBy` values while continuing.
- **Local duplicate skills may take precedence unexpectedly** → Accept the current OpenCode behavior silently rather than claiming a precedence guarantee.
- **A shallow copy could install only the bundle directory and lose child skills** → Preserve recursive relative paths and cover catalog grouping plus recursive installation with focused tests.
- **Evidence-based cleanup can still remove a needed file by mistake** → Back up the complete deletion set first, use an explicit source-attributed allowlist, and stop on any ownership ambiguity.
- **Cleanup can corrupt `opencode.json` or alter retained skills** → Validate JSON and require byte-identical comparisons for every protected skill before completion.
- **OpenCode can retain the previous discovery state after migration** → Restart OpenCode after the cleanup and validation succeed.

## Migration Plan

1. Add shared global-CLI planning, preflight, execution, and injectable command seams while preserving CodeGraph configuration behavior.
2. Register OpenSpec in extras and route both selected CLIs through the shared path.
3. Move the 12 OpenSpec skill directories beneath `assets/skills/openspec/` and verify one grouped catalog item plus recursive installation behavior.
4. Update orchestrator guidance with the bounded bootstrap gate and context cache, including nested global-skill version lookup.
5. Install the nested bundle to `~/.config/opencode/skills/openspec/` on this machine and verify all 12 child skills.
6. Inventory the confirmed Gentle AI-owned global footprint against the read-only source repository and create one timestamped backup of every item that will be removed or edited.
7. Remove the 21 obsolete skill directories and the allowlisted agent, command, prompt, plugin, `.atl`, and ownership-marker references without touching protected or unattributed files.
8. Validate `opencode.json`, prove the four retained general skills and 12 nested OpenSpec skills remain byte-identical, then restart OpenCode.
9. Validate with focused installer/catalog tests, agent-asset contract tests, and the repository's full test suite.

Rollback restores the previous CodeGraph-specific installer path and removes the OpenSpec extra and bootstrap instructions. Repository skill paths can be moved back if needed; the one-time global cleanup is restored from its timestamped backup. Global package versions already updated by a run are not downgraded automatically.

## Open Questions

None. The confirmed brief resolves the package-manager order, runtime floor, transaction boundary, bootstrap commands, version-drift behavior, nested bundle layout, evidence-based cleanup boundary, backup and validation sequence, and OpenCode restart.
