## Context

The interactive installer receives complete ordered slices from `install.PlanInstallation` and `install.ApplyInstallation`, but `internal/tui/wizard.go` replaces entries after fixed limits with an omission summary. The model does not currently track terminal dimensions or scroll position. Confirmation also computes and displays planning errors inline, while completion displays an error followed by any operations completed before that error.

The change is presentation-only: the installer must keep using the same selection, planning, application, ordering, and error paths while making every plan and report entry accessible in terminals of varying height.

## Goals / Non-Goals

**Goals:**
- Present every planned change before installation can be confirmed.
- Present every reported operation after installation, including partial results returned with an error.
- Preserve the order supplied by the installation layer.
- Keep long lists usable through explicit terminal-aware navigation without omission summaries.
- Keep confirmation, back, exit, and error behavior available and understandable.

**Non-Goals:**
- Changing selected installer content or the operations produced by planning and application.
- Reordering, filtering, grouping, or otherwise transforming plan or report entries.
- Changing installation preflight, execution, rollback, or error semantics.
- Redesigning catalog and extras selection screens.

## Decisions

### Use a terminal-aware list window backed by the complete slice

The model will retain terminal dimensions from Bubble Tea window-size messages and maintain independent scroll offsets for confirmation and completion output. Rendering will reserve lines for existing headings, summaries, errors, and controls, then display the corresponding contiguous window of the complete plan or report. Navigation will make every entry reachable and a position indicator will communicate the visible range and total count without implying omitted data.

This keeps fixed controls usable when content exceeds terminal height. Rendering the entire list without scrolling was rejected because terminal overflow can push confirmation and navigation controls off-screen. Keeping larger fixed limits was rejected because any fixed limit still hides entries.

### Implement list navigation locally without a new runtime dependency

The wizard will use small local helpers to calculate the available list height, clamp offsets after content or terminal-size changes, and return the visible ordered range. Existing directional conventions will be reused, with page and boundary navigation where useful, and help text will describe the active controls.

Adding the Bubbles viewport component was considered, but a single-line-list viewport does not justify a new dependency and would add integration state beyond the required presentation change.

### Preserve installer-produced data and error paths

The confirmation view will continue to call `PlanInstallation` and display its result in the returned order. A planning error will continue to follow the existing inline-error path. The completion view will continue to display the existing error heading and then the entire partial or complete report returned by `ApplyInstallation`, in order. No sorting, deduplication, filtering, or changes to installer calls will be introduced.

### Replace implicit completion exit with explicit controls when navigation is needed

Scroll keys must remain available on long completion reports, so they cannot also trigger the current any-key exit behavior. The completion help will identify explicit exit keys while navigation keys move through the report. Existing immediate quit keys remain available globally.

## Risks / Trade-offs

- [A very short terminal leaves little room for list content] → Calculate height from fixed chrome, guarantee a usable list region when possible, and keep resize handling and offset clamping deterministic.
- [Users may not realize more entries are reachable] → Show a visible-range/total indicator and phase-specific navigation help whenever the list exceeds the available region.
- [New key handling could interfere with confirmation or exit] → Keep installation confirmation and back keys unchanged, reserve scrolling to non-conflicting keys, and cover phase-specific behavior with model tests.
- [Repeated planning during view rendering can change list length after a resize or input] → Clamp the confirmation offset against the current plan on every relevant render/update without changing planning behavior.

## Migration Plan

No data or configuration migration is required. Deploy the wizard presentation and tests together; rollback restores the previous renderer without affecting installed content.

## Open Questions

None.
