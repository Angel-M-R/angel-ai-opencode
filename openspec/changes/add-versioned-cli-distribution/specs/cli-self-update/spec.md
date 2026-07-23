## ADDED Requirements

### Requirement: Stable TUI launches check for updates automatically
A stable release build SHALL check for a newer stable release before every TUI launch and SHALL apply an available update without requesting confirmation. A build whose version is `dev` MUST NOT check for or apply updates.

#### Scenario: Stable TUI launch has a newer release
- **WHEN** a stable `angel-ai` invocation is about to open the TUI and the manifest declares a valid newer stable version
- **THEN** the updater attempts to verify, replace, and relaunch before the TUI opens

#### Scenario: Stable TUI launch is current
- **WHEN** a stable `angel-ai` invocation is about to open the TUI and the manifest version is equal to or older than the running version
- **THEN** the current binary opens the TUI without replacement

#### Scenario: Development build launches the TUI
- **WHEN** a `dev` build is about to open the TUI
- **THEN** it opens the TUI without making an update network request

### Requirement: Explicit update forces a check
The `angel-ai update` command SHALL perform an update check even though no TUI launch is requested. It SHALL apply only a strictly newer stable version and SHALL report when the installed version is already current.

#### Scenario: Forced check finds a newer version
- **WHEN** the user runs `angel-ai update` and the manifest declares a valid newer stable version
- **THEN** the updater attempts the same verified replacement and guarded relaunch used by automatic updates

#### Scenario: Forced check finds no newer version
- **WHEN** the user runs `angel-ai update` and the manifest version is equal to or older than the installed stable version
- **THEN** the command reports that no update is required and does not replace the binary

#### Scenario: Development build requests an update
- **WHEN** the user runs `angel-ai update` from a `dev` build
- **THEN** the command reports that self-update is disabled and performs no update network request

### Requirement: Update discovery uses a bounded direct manifest request
The updater SHALL retrieve a latest-release manifest directly from GitHub Releases without using the GitHub API. The manifest SHALL provide a stable SemVer version, artifact URL, and lowercase SHA-256 digest. The manifest request SHALL time out after two seconds.

#### Scenario: Valid latest manifest is returned
- **WHEN** the direct manifest request completes within two seconds with valid required fields
- **THEN** the updater evaluates the declared stable version against the running stable version

#### Scenario: Manifest request exceeds the deadline
- **WHEN** the direct manifest request does not complete within two seconds
- **THEN** the updater cancels the request, warns the user, and continues with the running version

#### Scenario: Manifest data is invalid
- **WHEN** the manifest is malformed, omits a required field, declares a prerelease, or contains an invalid digest
- **THEN** the updater warns the user and continues with the running version without downloading an artifact

### Requirement: Downloaded updates are verified before replacement
The updater SHALL download an eligible artifact to a temporary file, compute its SHA-256 digest, and compare it to the manifest before making the file executable or replacing the installed binary. A mismatch MUST leave the installed executable unchanged.

#### Scenario: Candidate checksum matches
- **WHEN** an eligible downloaded artifact matches the manifest SHA-256
- **THEN** the updater may prepare it for atomic replacement

#### Scenario: Candidate checksum mismatches
- **WHEN** an eligible downloaded artifact does not match the manifest SHA-256
- **THEN** the updater removes temporary state, warns the user, and continues with the installed binary

### Requirement: Replacement is atomic and preserves immediate rollback
The updater SHALL prepare update and backup state in the installed executable's directory and SHALL replace the executable through a same-filesystem atomic operation. If replacement fails, the current executable SHALL remain usable or be restored immediately.

#### Scenario: Atomic replacement succeeds
- **WHEN** a verified executable has been fully prepared in the installation directory
- **THEN** it atomically replaces the installed path without exposing partial bytes

#### Scenario: Replacement fails
- **WHEN** the atomic replacement operation cannot complete
- **THEN** the updater restores or retains the previous executable, warns the user, and continues the current process

### Requirement: Successful updates relaunch once with identical arguments
After successful replacement, the updater SHALL relaunch the new executable automatically with the same argument vector. It SHALL carry an internal marker that prevents another update check in the relaunched process and preserves completion behavior for both TUI and explicit update invocations.

#### Scenario: Automatic update relaunches the TUI invocation
- **WHEN** an automatic update replaces the binary successfully
- **THEN** the new binary starts with the original TUI arguments, skips a repeated update check, and continues to the TUI

#### Scenario: Explicit update relaunches command completion
- **WHEN** `angel-ai update` replaces the binary successfully
- **THEN** the new binary starts with the original arguments, skips another update attempt, and reports completion without looping

#### Scenario: Immediate relaunch fails
- **WHEN** the running process cannot start the replacement executable
- **THEN** the updater atomically restores the prior binary and warns the user

### Requirement: Update failures fail open with visible warnings
Any discovery, download, verification, preparation, replacement, or immediate relaunch failure SHALL keep or restore the current version and display a warning. For an automatic check, the original TUI invocation SHALL continue whenever the current executable remains runnable. The updater SHALL NOT claim rollback coverage for failures after the replacement process has successfully started.

#### Scenario: Automatic update encounters a recoverable failure
- **WHEN** an update step fails before the replacement process successfully starts and the current binary remains available
- **THEN** the CLI displays a warning and opens the requested TUI with the current process

#### Scenario: Replacement process later fails inside the application
- **WHEN** the new executable has started successfully and later encounters an application failure
- **THEN** the updater performs no automatic rollback based on that later failure
