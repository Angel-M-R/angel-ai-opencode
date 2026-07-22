## ADDED Requirements

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
