## Why

The installer currently hides planned and completed operations behind fixed list limits, so users cannot review the full impact before confirming or verify every operation afterward. The output should remain complete and usable even when it exceeds the terminal height.

## What Changes

- Remove summary truncation such as `… and N more` from the installer confirmation preview and completion result.
- Show every planned installation change before confirmation, preserving the existing order.
- Show every reported operation after installation, including operations reported before an error, preserving the existing order and error behavior.
- Make long preview and result lists navigable within the terminal without hiding entries.
- Keep installation selection, planning, execution, and error semantics unchanged; this change affects presentation only.

## Capabilities

### New Capabilities
- `installer-output-presentation`: Defines complete, ordered, and navigable presentation of installer plan previews and completion results.

### Modified Capabilities

None.

## Impact

- Affects the interactive installer UI that renders the confirmation plan and final operation report.
- Requires UI-focused tests for complete rendering, preserved ordering, long-list navigation, and unchanged error presentation.
- Does not change installer planning or application APIs, selected content, operation ordering, dependencies, or filesystem behavior.
