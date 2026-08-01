## ADDED Requirements

### Requirement: Published TUI plugins are independent default extras
The installer SHALL expose `opencode-open-in-app` and `opencode-openspec-task-tui` as two independent extras immediately after Subagent statusline, and SHALL select both extras by default.

#### Scenario: Extras screen opens with defaults
- **WHEN** the installer initializes the extras screen
- **THEN** `opencode-open-in-app` and `opencode-openspec-task-tui` appear immediately after Subagent statusline and are both selected

#### Scenario: User changes only one published plugin selection
- **WHEN** the user deselects one of the two published plugin extras and leaves the other selected
- **THEN** the installer retains the two independent selection states

### Requirement: Selected extras reconcile unversioned npm entries
For each selected published plugin extra, the installer SHALL reconcile the corresponding unversioned package name into the `tui.json` plugin array. It MUST preserve unrelated plugin entries and their relative order.

#### Scenario: Selected package is not already configured
- **WHEN** a published plugin extra is selected and neither its npm entry nor a recognized local equivalent exists in `tui.json`
- **THEN** the installer appends the unversioned package name without removing or reordering unrelated entries

#### Scenario: Versioned npm entry already exists
- **WHEN** a selected package already appears in `tui.json` with a version suffix
- **THEN** the installer reconciles it to one unversioned package entry at the existing matching position

### Requirement: Equivalent local bundles are recognized without execution
When reconciling `tui.json`, the installer SHALL inspect absolute, relative, and `file:` local plugin entries by reading the referenced bundle without executing it. It SHALL recognize only a literal plugin `id` property equal to `opencode-open-in-app` or `openspec-task-progress`, mapping those IDs to `opencode-open-in-app` and `opencode-openspec-task-tui` respectively.

#### Scenario: Absolute local bundle has a recognized literal ID
- **WHEN** an absolute local plugin entry references a readable bundle whose literal plugin `id` is `opencode-open-in-app`
- **THEN** the installer treats that entry as equivalent to the `opencode-open-in-app` npm package without executing the bundle

#### Scenario: Relative or file-prefixed bundle has a recognized literal ID
- **WHEN** a relative or `file:` local plugin entry references a readable bundle whose literal plugin `id` is `openspec-task-progress`
- **THEN** the installer treats that entry as equivalent to the `opencode-openspec-task-tui` npm package without executing the bundle

#### Scenario: Local bundle cannot be recognized safely
- **WHEN** a local plugin entry is unreadable or does not contain either recognized literal plugin ID property
- **THEN** the installer preserves the entry and does not infer equivalence from its path or file name

### Requirement: Recognized duplicates migrate at the first position
For a selected published plugin, the installer SHALL replace all existing npm entries and recognized local equivalents for that plugin with exactly one unversioned npm entry at the first matching array position. It MUST preserve unrelated and unrecognized entries.

#### Scenario: Configuration contains duplicate local and npm equivalents
- **WHEN** `tui.json` contains multiple recognized local bundles, versioned or unversioned npm entries, or a mixture of those equivalents for a selected extra
- **THEN** the resulting plugin array contains one unversioned npm entry at the first matching position and no later recognized duplicates

#### Scenario: Configuration contains unrelated entries around duplicates
- **WHEN** recognized duplicates are interleaved with unrelated or unrecognized plugin entries
- **THEN** duplicate migration leaves those unrelated or unrecognized entries present and in their previous relative order

### Requirement: Deselection does not uninstall existing plugins
The installer MUST NOT remove an existing npm entry or local bundle for either published plugin solely because its extra is deselected in a later run.

#### Scenario: Previously installed npm package is deselected
- **WHEN** `tui.json` already contains a published npm entry and the corresponding extra is deselected
- **THEN** the installer leaves that npm entry unchanged

#### Scenario: Recognized local bundle is deselected
- **WHEN** `tui.json` contains a recognized local equivalent and the corresponding extra is deselected
- **THEN** the installer leaves that local entry unchanged rather than migrating or removing it

### Requirement: Extras screen explains non-destructive deselection
The installer TUI SHALL render the global notice `Si ya está instalado, desmarcarlo no lo desinstalará` on the extras screen.

#### Scenario: Extras screen is rendered
- **WHEN** the installer displays Integrations and extras
- **THEN** the rendered view contains the notice exactly once
