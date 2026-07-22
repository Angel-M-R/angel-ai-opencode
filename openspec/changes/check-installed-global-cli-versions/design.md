## Context

CodeGraph and OpenSpec use the same descriptor-driven global CLI path. Today `PlanInstallation` validates the package manager and OpenSpec's Node.js floor, while `ApplyInstallation` unconditionally runs each descriptor's `@latest` package command before configuration reconciliation. The shared command seam already injects executable lookup and command execution, and the installer already has semantic-version parsing suitable for comparing local and registry versions.

The new policy must avoid mutating a working CLI, expose its version state in both plan and result output, and detect broken selected-manager installations before any package changes. This machine's OpenSpec state—registered by npm but missing from `PATH`—is an example of a blocking inconsistency that must receive guided recovery rather than automatic repair.

## Goals / Non-Goals

**Goals:**

- Build one shared inspection policy for selected CodeGraph and OpenSpec descriptors.
- Prevalidate all selected CLIs before any installation and carry the resulting status into planning and application output.
- Install `@latest` only for a genuinely absent CLI, while preserving healthy current, outdated, ahead-of-registry, and registry-unverified executables.
- Keep manager selection, Node preflight, descriptor ordering, post-CLI repreparation, and configuration-write boundaries intact.
- Make every new branch testable through injected command behavior.

**Non-Goals:**

- Updating an installed but outdated CLI.
- Repairing or removing global packages, links, or executables automatically.
- Searching a non-selected manager or inferring cross-manager ownership.
- Changing OpenSpec skill bundling, CodeGraph MCP/`AGENTS.md` reconciliation, or the OpenSpec worker bootstrap.
- Running real npm/pnpm operations or the full repository test suite as regression coverage.

## Decisions

### Represent preflight as an inspection snapshot

Preflight will return the selected manager plus one ordered inspection record per selected descriptor. Each record will contain package registration, executable discovery, parsed installed and registry versions when available, a disposition (`install`, `current`, `outdated`, `ahead`, or `registry-unverified`), and any warning. Planning and application will render this shared model instead of independently deriving status strings.

This keeps the plan and final result vocabulary coherent and prevents application from repeating discovery in a way that could select different work. A set of booleans passed between existing functions was rejected because it would scatter invalid state combinations and output decisions.

### Inspect only through the selected package manager

The existing npm-first/pnpm-fallback selection remains the first package-manager decision. The manager model will gain injected, manager-specific package-registration and latest-version probes. Package identifiers will be stored separately from their `@latest` installation specs so scoped package names remain unambiguous.

No npm query will run after pnpm is selected, and no pnpm query will run after npm is selected. A working executable that is not registered by the selected manager will be preserved without claiming ownership; searching the other manager was rejected because it would violate deterministic selection and create implicit migration behavior.

### Classify local health before deciding to install

After manager and existing Node.js preflight, every descriptor will be inspected before any installation:

1. Query package registration through the selected manager and resolve the executable from `PATH`.
2. If the selected manager registers the package but the executable is unavailable, block with manager-specific recovery instructions and perform no cleanup.
3. If the executable exists, run its descriptor-defined version command and require one parseable semantic version. Failure, empty output, or malformed output blocks with recovery guidance.
4. Query and parse the registry's latest version. Compare it with a healthy installed version to classify it as current, outdated, or ahead.
5. If registry lookup is unavailable or unusable, preserve a healthy executable as `registry-unverified` with a warning. If the CLI is absent, block because the complete prevalidation cannot establish a safe install action.
6. Only an executable-absent, selected-manager-unregistered CLI with a usable latest lookup receives the `install` disposition.

Treating every missing executable as installable was rejected because it would silently repair the known broken OpenSpec registration. Treating an outdated executable as an update candidate was rejected because the requested policy is install-if-missing, not converge-to-latest.

### Apply only prevalidated install dispositions

`ApplyInstallation` will complete preparation and all CLI inspections before invoking any package installation. It will then process records in descriptor order, run `@latest` only for `install` records, and verify the resulting executable and parseable version. Existing healthy records perform no package mutation and emit their preflight classification in the final results.

Any inspection or installation failure continues to prevent configuration writes. After all selected CLI records complete, application retains the existing single repreparation before file reconciliation. This preserves operation ordering and configuration boundaries while removing unnecessary package commands.

### Use focused injected behavior tests

Tests will extend the existing command seam rather than invoke the host package managers. Focused unit cases will cover both descriptors and both selected managers, all comparison states, registry failure behavior, broken registration, version-command failures, all-CLI prevalidation before installation, plan/result reporting, and unchanged Node/configuration ordering. Verification will run only targeted installer tests.

## Risks / Trade-offs

- [A working executable may belong to a different manager than the selected one] → Preserve it without asserting ownership and never query the other manager.
- [Package-manager list output differs between npm and pnpm] → Encapsulate commands and parsing per selected manager behind the injected seam, with fixtures for present and absent packages.
- [Registry outages make freshness unknowable] → Continue only for a healthy local executable and clearly report `registry-unverified`; block missing CLI installation before side effects.
- [Preflight adds network latency] → Query only selected CLIs and favor correctness and complete prevalidation over partial installation.
- [Version output may include tool-specific decoration] → Keep descriptor-owned extraction narrowly defined and require a single interpretable semantic version rather than guessing.

## Migration Plan

1. Extend descriptor, manager, and inspection models without changing selected extras or configuration preparation.
2. Replace unconditional plan/install decisions with shared inspection records.
3. Add focused injected unit coverage before relying on the new branches.
4. For currently inconsistent global installations, return recovery instructions; operators repair them outside the installer and rerun it.

Rollback restores unconditional `@latest` processing and the previous tests; no persistent data migration or automatic package cleanup is introduced.

## Open Questions

None. The confirmed brief resolves policy, failure handling, ordering, manager scope, and test boundaries.
