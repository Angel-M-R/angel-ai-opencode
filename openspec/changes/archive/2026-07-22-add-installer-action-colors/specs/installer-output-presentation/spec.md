## ADDED Requirements

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
