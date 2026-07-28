## Context

The shared orchestrator policy currently treats intermediate non-zero commands through a corrected-failure model that expects diagnosis, bounded repair, and equivalent-or-broader validation. A read-only probe such as `ls openspec` before the directory exists is different: the command accurately observes pre-operation state, requires no repair, and may be followed by successful creation plus authoritative final validation. Applying the corrected-failure checklist literally can therefore force a continuation question despite a clean final result.

The repository policy source is `assets/agents/angel-orchestrator.md`. The executable captures that asset tree through `go:embed`, and installation reconciles a selected agent asset into `<configDir>/agents/` unchanged. Repository contract tests read the source asset directly, while embedded/source parity and temporary-directory installation tests cover other parts of propagation. The observed installed orchestrator is older than the repository source because changing the repository does not mutate an already installed copy: the binary must contain the new asset and the installer must be rerun for that selected agent.

## Goals / Non-Goals

**Goals:**

- Define a narrow recovered-probe classification for failed non-destructive observations followed by successful in-scope work and authoritative green validation from the same worker.
- Preserve complete incident evidence while avoiding a blocker question solely because of an eligible probe failure.
- Keep every existing final-state, evidence, scope, and safety gate conjunctive, and make destructive actions an explicit mandatory stop.
- Make source, embedded, and installed-policy propagation testable and explain why an installed copy may lag.
- Add regression coverage for the exact missing-`openspec/` probe scenario and applicable negative cases.

**Non-Goals:**

- Relax the treatment of unresolved implementation, verification, syntax, or invocation failures.
- Permit cross-worker recovery evidence, unrelated success evidence, scope expansion, destructive commands, or missing final validation.
- Change OpenSpec routing, worker permissions, planned-task deferral, installation selection semantics, or unrelated agent behavior.
- Automatically mutate a user's installed agent configuration during tests or planning.

## Decisions

### Model recovered probes as a separate evidence path

The shared policy will define a recovered non-destructive probe independently from corrected implementation or verification failures. Eligibility requires that the failed command only inspected state, caused no mutation, and failed because expected pre-operation state was absent or not yet established. The same worker must then complete the authorized operation and run authoritative final validation that proves the requested final state.

This avoids inventing a “repair” for an accurate observation and avoids requiring the final validation to be an equivalent rerun of the probe. The rejected alternative is to broaden the existing corrected-failure definition, which would blur the distinction between observational probes and failures in work or validation.

### Use one conjunctive automatic-continuation gate

Automatic continuation is allowed only when all probe-specific evidence is present, the worker ends `done`, final authoritative validation is green, and the worker reports no Brief deviation, scope expansion, or out-of-scope work. The failed command, non-zero exit code, cause, successful operation, final validation command and exit code, and why that validation proves the final state remain visible in execution order.

Any failed condition uses the existing mandatory-stop path. Destructive actions are explicitly ineligible regardless of later success. This strict conjunction is preferred over heuristic severity because it is auditable and preserves the current safety boundary.

### Keep the repository asset authoritative and test each propagation boundary

`assets/agents/angel-orchestrator.md` remains the source of truth. Current compilation embeds that tree, and the installer copies selected agent assets unchanged into the configured `agents/` directory, backing up changed destinations. Implementation will update the source policy and focused contract assertions, retain embedded-versus-directory parity coverage, and extend temporary installation coverage to include the orchestrator policy and recovered-probe text.

No test will inspect or overwrite the developer's live `~/.config/opencode` destination. Release/update propagation remains: build or obtain a binary containing the updated embedded asset, then rerun installation with the orchestrator selected. The rejected alternative is a second generated policy file, which would create another source of drift.

### Test the exact incident and its safety boundaries

Focused regression assertions will encode the sequence: `ls openspec` fails because the directory does not yet exist, the same worker creates the requested OpenSpec state, authoritative status validation exits zero, the worker returns `done`, and no scope issue is reported. Assertions will require incident visibility and prohibit a continuation question solely for that probe. Negative cases will retain stops for unresolved or destructive incidents, red final validation, `partial`/`blocked`, insufficient evidence, different-worker evidence, and deviations or scope growth.

## Risks / Trade-offs

- [A worker labels a real failure as a probe] → Require a non-destructive inspection-only command, an absence/unestablished-state cause, and authoritative final-state evidence; ambiguous classification stops.
- [A green but unrelated command is accepted] → Require the reported final validation to be authoritative for the requested operation and to explain the satisfied final state.
- [Installed behavior still appears stale] → Test source/embed parity and temporary installation, and document that existing destinations update only after a new binary or source override is installed with the agent selected.
- [Text-contract tests become brittle] → Assert the normative eligibility and stop clauses in focused sections without refactoring unrelated policy text.
- [The exception weakens safety] → Keep `done`, green final validation, complete evidence, same-worker identity, non-destructive behavior, and no-scope-issue conditions conjunctive.

## Migration Plan

1. Update the authoritative orchestrator asset and focused contract tests.
2. Verify embedded asset parity and unchanged temporary installation of the selected orchestrator asset.
3. Run applicable Go tests and OpenSpec validation.
4. Deliver the policy through the normal rebuilt/released binary and rerun agent installation to reconcile existing configured copies; rollback uses the installer's destination backup or the prior binary asset.

## Open Questions

None.
