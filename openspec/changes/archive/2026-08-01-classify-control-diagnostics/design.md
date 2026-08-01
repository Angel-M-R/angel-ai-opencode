## Context

The orchestrator has one shared implementation-result policy and field set that is reused by Direct and OpenSpec workers at every control point. Today, internal read-only bookkeeping can be placed in `deviations`; because every reported deviation is a mandatory-stop condition, harmless diagnostics can interrupt an otherwise green flow. The policy already has separate classifications for generated validation artifacts, corrected intermediate failures, recovered non-destructive probes, and real deviations, and those boundaries must remain intact.

The authoritative repository prompt is `assets/agents/angel-orchestrator.md`. Although an installed prompt may be distributed from that asset, this change intentionally updates only the repository asset and does not synchronize the local installed copy.

## Goals / Non-Goals

**Goals:**

- Add `control diagnostics` to the shared result contract for every control point.
- Keep diagnostic entries visible, structured, ordered, and explicitly `none` when absent.
- Permit automatic continuation for eligible bookkeeping-only diagnostics, including entries labeled `incomplete` when the final relevant state is green.
- Keep `control diagnostics` and `deviations` mutually exclusive and preserve all real mandatory stops.
- Limit implementation to the authoritative repository orchestrator asset and validate it textually.

**Non-Goals:**

- Change workflow prerequisites, route selection, planned-task deferral, corrected-failure eligibility, recovered-probe eligibility, or generated-validation-artifact classification.
- Reclassify functional work, scope expansion, out-of-scope work, destructive actions, mutations, touched files, failed commands, incomplete statuses, or red validation as diagnostics.
- Update tests, application code, installer behavior, generated copies, or the local installed orchestrator prompt.

## Decisions

### Add one shared result category rather than weaken deviations

The shared field set will include `control diagnostics` alongside, but separate from, `deviations`. The category is always present and uses `none` when empty. Entries remain in observation order and identify the applicable control point, the bookkeeping observation, completeness as `complete` or `incomplete`, and the green final-state evidence supporting continuation.

The two categories are result-level mutually exclusive: a result with any real deviation reports `control diagnostics: none`, and a result with diagnostics reports `deviations: none`. The rejected alternative is to introduce “benign deviations,” because existing policy deliberately treats any deviation as a stop and weakening that term would make safety routing ambiguous.

### Use a conjunctive, route-neutral eligibility gate

A record qualifies only as internal read-only bookkeeping at a control point and only when the bounded result has no failed command, no mutation, no touched file, status `done`, and green final relevant validation. An eligible diagnostic is surfaced but does not change the control point's existing clean-result route.

If descriptive bookkeeping details are incomplete while the result still establishes the eligibility facts and green final state, the entry is labeled `incomplete` and routing continues. Missing evidence for a safety fact is not converted into a diagnostic. The rejected alternative is severity-based judgment, which would be less auditable and could conceal a real failure or scope change.

### Preserve every existing stop and exceptional evidence path

Functional deviations, functional or scope expansion, out-of-scope actions, destructive actions, failed commands, mutations, touched files, `partial` or `blocked` status, and red final validation remain governed by the existing mandatory-stop policy. Corrected failures and recovered probes continue to use their current evidence rules and are not control diagnostics. No route-local prerequisite is removed.

### Keep implementation source-only and validation textual

Implementation will edit only `assets/agents/angel-orchestrator.md`, updating the shared fields, classification rule, and all-control-point usage coherently in that source. It will not synchronize the installed prompt despite the broader distribution relationship; synchronization is deferred by explicit scope decision.

Validation consists of reviewing the final textual diff for category separation and retained stop clauses, then running `git diff --check`. No contract test or generated validation artifact is required for this prompt-only policy change.

## Risks / Trade-offs

- [A real deviation is mislabeled as bookkeeping] → Require every eligibility condition conjunctively and retain the existing stop whenever a failed command, mutation, touched file, scope issue, destructive action, incomplete status, or red final state exists.
- [An `incomplete` label hides missing safety evidence] → Permit incompleteness only in descriptive diagnostic details after the result still establishes all classification facts and green final state.
- [Consumers conflate diagnostics with deviations] → Require both fields explicitly and make them mutually exclusive, with `none` for the unused category.
- [The local installed prompt remains stale] → State that repository-asset-only scope is intentional and defer synchronization rather than implying it occurred.
- [Textual validation misses behavior drift] → Keep the edit narrowly confined to one policy asset and require an explicit review of both the non-blocking category and unchanged stop clauses.

## Migration Plan

1. Update the shared result-field and classification text in `assets/agents/angel-orchestrator.md` without editing any other file.
2. Review the diff to confirm every control point receives the category, `control diagnostics` remains separate from `deviations`, and existing mandatory stops and evidence paths remain unchanged.
3. Run `git diff --check` and retain its exit-code evidence.
4. Do not synchronize the local installed prompt as part of this change; rollback is a single-file policy revert.

## Open Questions

None.
