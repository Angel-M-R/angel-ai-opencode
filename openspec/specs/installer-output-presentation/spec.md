# installer-output-presentation Specification

## Purpose
TBD - created by archiving change remove-installer-output-truncation. Update Purpose after archive.
## Requirements
### Requirement: Complete installation plan preview
The installer SHALL make every line returned by installation planning available in the confirmation preview before installation can be confirmed. It SHALL preserve the returned order and SHALL NOT replace any plan line with an omission summary such as `… and N more`.

#### Scenario: Plan fits within the terminal
- **WHEN** installation planning returns a list that fits in the available confirmation output region
- **THEN** the confirmation preview displays every plan line in the returned order before the confirmation controls

#### Scenario: Plan exceeds the terminal height
- **WHEN** installation planning returns more lines than fit in the available confirmation output region
- **THEN** the confirmation preview provides navigation that can reach every plan line in the returned order without replacing any line with an omission summary

#### Scenario: Planning returns an error
- **WHEN** installation planning returns an error
- **THEN** the confirmation preview preserves the existing inline planning-error presentation and confirmation behavior

### Requirement: Complete installation result output
The installer SHALL make every operation line returned by installation application available in the completion result. It SHALL preserve the returned order and SHALL NOT replace any operation line with an omission summary such as `… and N more`.

#### Scenario: Successful installation returns many operations
- **WHEN** installation completes successfully with more operation lines than fit in the available completion output region
- **THEN** the completion result provides navigation that can reach every operation line in the returned order

#### Scenario: Installation fails after completed operations
- **WHEN** installation application returns an error together with operation lines completed before the error
- **THEN** the completion result preserves the existing error presentation and makes every returned operation line reachable in the returned order

#### Scenario: Completion result fits within the terminal
- **WHEN** installation application returns a result that fits in the available completion output region
- **THEN** the completion result displays every returned operation line in the returned order

### Requirement: Usable long-list navigation
The installer SHALL size confirmation and completion list regions according to the current terminal height, SHALL keep their surrounding status and controls usable, and SHALL expose navigation and position feedback whenever not all entries fit at once.

#### Scenario: User traverses an overflowing list
- **WHEN** the current plan or result contains entries above and below the visible list region
- **THEN** the user can navigate to the first, previous, next, and last portions of the same complete ordered list

#### Scenario: Terminal height changes
- **WHEN** the terminal is resized while a confirmation plan or completion result is displayed
- **THEN** the visible range and scroll position are recalculated and clamped without dropping or reordering entries

#### Scenario: User acts from a long confirmation preview
- **WHEN** the confirmation plan exceeds the terminal height
- **THEN** the existing install, back, and quit actions remain available alongside list navigation

### Requirement: Presentation-only scope
The output presentation change MUST NOT alter installation selection, planned operations, applied operations, operation order, or error propagation.

#### Scenario: Same installation request before and after the presentation change
- **WHEN** the installer processes the same selections, extras, assets directory, and configuration directory
- **THEN** it passes the same request to planning and application and changes only how their returned lines are presented

### Requirement: Risk-semantic installer action labels
The installer TUI SHALL color recognized action labels in both the confirmation preview and the completion result according to their risk semantics. It SHALL render `SIN CAMBIOS` and `sin cambios` in gray; `CREAR`, `INSTALAR`, `creado`, `instalado`, and the completion label `Instalación completada` in green; `ACTUALIZAR` and `actualizado` in yellow; `REEMPLAZAR` and `backup` in magenta; and `ERROR` and `Error:` in red, using Lip Gloss ANSI-256 styles consistent with the existing installer palette.

#### Scenario: Confirmation preview contains every planned action type
- **WHEN** the confirmation plan contains no-change, create, install, update, replace, and planning-error labels
- **THEN** the TUI renders each recognized label in its specified semantic color without changing the plan lines

#### Scenario: Completion result contains every operation type
- **WHEN** the completion result contains backup, created, installed, updated, unchanged, completion, or error labels
- **THEN** the TUI renders each recognized label in its specified semantic color without changing the report or status text

### Requirement: Label-only styling boundary
The installer TUI SHALL apply action color only to the exact recognized label prefix. It MUST leave existing separator spacing, paths, package specifications, backup or replacement details, and error details in the normal terminal color, and MUST preserve all existing text, capitalization, alignment, and spacing. Labels that are unknown or do not match an exact recognized prefix boundary SHALL remain unstyled.

#### Scenario: Recognized label precedes a path or detail
- **WHEN** a recognized action label is followed by its existing alignment spaces and a path, package specification, backup detail, replacement detail, or error detail
- **THEN** the ANSI style ends at the exact label boundary and every following character is unchanged and rendered in the normal color

#### Scenario: Action label is unknown
- **WHEN** a preview or result line begins with an unknown label or a near-match of a recognized label
- **THEN** the complete line remains unchanged and unstyled

### Requirement: ANSI-independent list geometry
The installer TUI SHALL retain plans and reports as plain text and SHALL calculate wrapping, visible list windows, scroll offsets, resize clamping, and position feedback before applying ANSI styles to recognized label spans. Styled presentation MUST expose the same ordered content and visual-row geometry as the corresponding plain-text presentation.

#### Scenario: Styled action line wraps
- **WHEN** a recognized action line exceeds the available terminal width
- **THEN** it wraps at the same plain-text boundaries as before and only characters belonging to the recognized label receive color

#### Scenario: User scrolls styled overflowing content
- **WHEN** styled preview or result entries exceed the available list height and the user navigates or resizes the terminal
- **THEN** the same entries and visual rows remain reachable in the same order with correctly clamped offsets and unchanged position feedback

#### Scenario: ANSI is removed from rendered output
- **WHEN** ANSI escape sequences are stripped from a styled confirmation or completion view
- **THEN** its action and status text, capitalization, alignment, spacing, and wrapping are identical to the unstyled output
