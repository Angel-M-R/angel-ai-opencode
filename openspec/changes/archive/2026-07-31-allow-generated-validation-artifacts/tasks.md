## 1. Generated Artifact Contracts and Policy Assets

- [x] 1.1 Extend `internal/install/agent_assets_test.go` with a focused contract for the single authoritative generated-validation-artifacts result category, including command, zero exit code, attributable diff, retained paths, and exclusion of manual source edits.
- [x] 1.2 Add positive contract cases for `.next/`, `.codegraph/`, and `*.tsbuildinfo` proving that Git ignore or tracking state does not control eligibility and that eligible outputs are retained without authorization or deviation reporting.
- [x] 1.3 Add negative contract cases for failed commands, red final validation, destructive cleanup, missing or ambiguous causality, intervening mutation, manual source edits, functional expansion, and out-of-scope work.
- [x] 1.4 Add route-propagation assertions covering Direct Safe, planned OpenSpec implementation, bounded Direct and OpenSpec fixes, and final verification while retaining each route's existing prerequisites and stop behavior.
- [x] 1.5 Update `assets/agents/angel-orchestrator.md` so the authoritative shared result fields include a separate generated-validation-artifacts category with paths, producing command and exit code, causal-diff evidence, and retention status.
- [x] 1.6 Define eligibility from an authorized successful validation command and attributable regenerable diff with no intervening manual mutation or manual source-code edits, explicitly independent of Git ignore status and filename allowlists.
- [x] 1.7 Apply the shared classification to Direct Safe, planned OpenSpec batches, bounded review/post-verification fixes, and final OpenSpec verification without duplicating divergent route-local rules.
- [x] 1.8 Preserve existing corrected-failure and recovered-probe handling and mandatory stops for non-zero commands without complete recovery evidence, red final state, destructive actions, ambiguous provenance, incomplete status, deviations, scope expansion, and out-of-scope work.
- [x] 1.9 Require eligible outputs to remain in the workspace and prohibit automatic cleanup, reversion, deletion, staging, or committing while making clear that the classification does not expand a batch's functional scope.
- [x] 1.10 Update `assets/agents/openspec-implementer.md` to collect command-specific attribution, report eligible focused-validation outputs separately, retain them, and continue enforcing exact task scope and focused-only validation.
- [x] 1.11 Update `assets/agents/openspec-verifier.md` to permit only causally proven side effects of executed validation commands, including tracked generated artifacts, while keeping manual edit/write operations, generated-source commands, destructive cleanup, and all existing failure paths prohibited.
- [x] 1.12 Keep the repository agent assets canonical and update applicable embedded/source and temporary-installation contract coverage so the revised policies propagate unchanged through the normal build and install path.
- [x] 1.13 Add focused implementer and verifier asset assertions that generated command effects do not grant manual edit authority, widen a task batch, permit full-suite execution by an implementer, or allow automatic cleanup.
- [x] 1.14 Review every changed worker return contract to ensure it preserves all authoritative Shared corrected-failure fields in addition to the new generated-artifact category and route-specific evidence.

## 2. Focused Implementation Validation

- [x] 2.1 Format changed Go contract tests and run the focused agent-asset test set covering generated classification, route propagation, corrected failures, planned-batch scope, and verifier restrictions; record every command and exit code.
- [x] 2.2 Run applicable focused embedded/source parity and temporary-installation tests, retaining causal evidence for any generated outputs instead of cleaning them.
- [x] 2.3 Validate the change's OpenSpec artifacts and inspect the final focused diff for manual source edits or scope expansion that cannot be attributed to the planned implementation tasks.

## 3. Cross-Route Pre-existing Dotpath Filter

- [x] 3.1 Add the single positive contract case for an already-modified `.vscode/settings.json` that is identical at worker completion, asserting silent omission across every workspace-auditing route while preserving existing `.next/`, `.codegraph/`, `*.tsbuildinfo`, and safety-negative coverage.
- [x] 3.2 Update the canonical orchestrator and OpenSpec worker policy assets so every workspace audit filters a path only when its first relative component begins with `.`, a reliable worker-start baseline shows it was already modified, and its end state is identical; do not grant dotpath write authority or weaken normal handling when evidence is absent or the path changes.
- [x] 3.3 Run the focused agent-asset and propagation tests for the shared dotpath filter and inspect the focused diff to confirm no VS Code configuration objective, functional batch expansion, or generated-output safety regression was introduced.

## 4. Final Verification

<!-- owner: openspec-verifier -->
- [x] 4.1 Execute the focused agent-asset and propagation tests again and report each command, exit code, and any retained generated-validation-artifact evidence independently.
- [x] 4.2 Execute the repository's mandatory full test suite and build and report each command and exit code, preserving generated outputs and applying the causal classification without weakening any red-result stop.
- [x] 4.3 Execute authoritative OpenSpec validation for `allow-generated-validation-artifacts` and report the command, exit code, and scenario coverage for every positive and negative contract requirement, including the single pre-existing unchanged `.vscode/settings.json` case.
