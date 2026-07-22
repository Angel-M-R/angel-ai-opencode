## Context

The installer keeps preview and completion entries as plain strings produced by `internal/install`, while `internal/tui/wizard.go` adds indentation, wraps those strings into visual rows, calculates list windows, and renders them. Existing TUI colors use Lip Gloss ANSI-256 values, including gray `241`, green `42`, magenta `212`, and red `196`; this change adds a matching yellow action style. Styling must not alter the producer strings or the geometry used by long-list navigation.

## Goals / Non-Goals

**Goals:**

- Make action risk immediately scannable in both confirmation and completion views.
- Style only exact, recognized action or status labels while leaving separators, paths, details, and unknown text in the normal color.
- Preserve existing text, wrapping, visible ranges, resize behavior, scrolling, and feedback.
- Cover the complete label map and ANSI-sensitive layout behavior with focused tests and the full Go suite.

**Non-Goals:**

- Changing planner or installer output, action wording, capitalization, alignment, spacing, order, or error propagation.
- Adding colors outside the installer confirmation and completion presentation.
- Altering the existing `add-direct-execution-route` change or introducing a configurable theme.

## Decisions

### Keep action data plain and style only during TUI rendering

`confirmationPlan`, `report`, and installer package return values remain ANSI-free. The TUI recognizes labels on each logical plain-text line, performs wrapping and list-window calculations on plain rows, and injects ANSI only when emitting visible rows. Plain data remains directly comparable during confirmation revalidation, and escape sequences cannot affect width or scroll calculations.

Styling data earlier in `internal/install` was rejected because it would couple installation behavior to terminal presentation and make non-TUI consumers receive ANSI. Styling before wrapping was rejected because escape sequences could corrupt row boundaries and resize calculations.

### Use an exact label table and preserve the label span through wrapping

Recognition is anchored to the start of the logical action text after the existing two-space list indent and requires the exact current label boundary. The table covers:

| Semantic | ANSI-256 | Recognized labels |
| --- | --- | --- |
| No change | `241` | `SIN CAMBIOS`, `sin cambios` |
| Create/install/complete | `42` | `CREAR`, `INSTALAR`, `creado`, `instalado`, `Instalación completada` |
| Update | `220` | `ACTUALIZAR`, `actualizado` |
| Replace/backup | `212` | `REEMPLAZAR`, `backup` |
| Error | `196` | `ERROR`, `Error:` |

Recognition records a plain-text label span rather than rendering a styled string immediately. Wrapped rows retain any overlap with that span, so even narrow wrapping and scrolling operate on plain text before only the label characters are styled. Separating spaces and all following path or detail text are emitted after an ANSI reset. Unknown or near-match prefixes carry no style.

Loose case-insensitive or token-based matching was rejected because it could color paths or future free-form messages accidentally. Styling a fixed-width action column was rejected because current labels use different spacing and the product requirement protects that formatting.

### Treat completion and error headers as label-only status rendering

The existing completion label remains green. For an error result, only the existing `Error:` label is red and the appended error detail returns to the normal color. Planning errors use the same prefix renderer for the `ERROR` action label. No status text is rewritten.

### Test plain-text invariants as well as ANSI output

Table-driven tests assert every recognized label and color, unknown-label passthrough, and an ANSI reset at the exact boundary before spacing, path, or detail. View-level tests cover both confirmation and completion. Wrapping, navigation, resize, and range-feedback tests compare plain visual geometry and verify that stripping ANSI reproduces the unchanged original text. The complete Go test suite is the final validation gate.

## Risks / Trade-offs

- **[New labels remain uncolored]** → Exact matching deliberately falls back to unstyled text; extending the explicit table and its tests is required when producers add labels.
- **[ANSI leaks into geometry calculations]** → Keep recognition metadata separate from rendered text and add wrapping, resize, and scroll regressions that exercise styled lines.
- **[A style reset changes surrounding presentation]** → Render only the label span and test the first separator and following path/detail for normal-color output.
- **[Terminal color support varies]** → Use the same Lip Gloss ANSI-256 mechanism and established palette values already used by the installer.
