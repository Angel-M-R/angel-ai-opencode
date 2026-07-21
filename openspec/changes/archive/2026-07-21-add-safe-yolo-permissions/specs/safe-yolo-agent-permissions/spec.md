## ADDED Requirements

### Requirement: Explicit near-YOLO agent coverage
The system SHALL apply the near-YOLO permission policy explicitly to `angel-orchestrator`, `openspec-planner`, `openspec-implementer`, `openspec-verifier`, and the built-in `general` agent. A future Bash-capable agent MUST be added explicitly before it is considered covered by this policy.

#### Scenario: Covered agents perform routine work without approval
- **WHEN** any covered agent requests an ordinary operation that does not match a denial
- **THEN** OpenCode allows the operation without asking for approval

#### Scenario: Future agent is not implicitly claimed as covered
- **WHEN** a new Bash-capable agent definition is introduced without the near-YOLO rules
- **THEN** the policy does not represent that agent as covered and requires an explicit configuration update

### Requirement: Direct destructive Git operations are denied
The system SHALL deny direct variants of `git reset` using `--hard` and `git push` using `--force`, `--force-with-lease`, or the standalone `-f` flag for every covered agent. Other Git operations SHALL remain allowed without approval.

#### Scenario: Hard reset without revision is denied
- **WHEN** a covered agent requests `git reset --hard`
- **THEN** OpenCode denies the command

#### Scenario: Hard reset with revision is denied
- **WHEN** a covered agent requests a direct `git reset --hard <revision>` variant
- **THEN** OpenCode denies the command

#### Scenario: Long-form force push is denied
- **WHEN** a covered agent requests a direct `git push` variant containing `--force` or `--force-with-lease`
- **THEN** OpenCode denies the command

#### Scenario: Short-form force push is denied
- **WHEN** a covered agent requests a direct `git push` variant containing the standalone `-f` flag
- **THEN** OpenCode denies the command

#### Scenario: Normal Git remains allowed
- **WHEN** a covered agent requests a Git operation without a denied reset or push flag
- **THEN** OpenCode allows the command without asking for approval

### Requirement: Direct destructive deletion targets are denied
The system SHALL deny direct `rm` commands targeting `/`, `~`, `$HOME`, or named critical system roots and their descendants. The policy SHALL allow deletion within the active project and under `~/tmp` when no protected target is also present.

#### Scenario: Root deletion is denied
- **WHEN** a covered agent requests a direct `rm` variant whose operand is `/`
- **THEN** OpenCode denies the command

#### Scenario: Home deletion is denied
- **WHEN** a covered agent requests a direct `rm` variant whose operand is `~` or `$HOME`
- **THEN** OpenCode denies the command

#### Scenario: Critical system deletion is denied
- **WHEN** a covered agent requests a direct `rm` variant targeting a configured critical system root or a path beneath it
- **THEN** OpenCode denies the command

#### Scenario: Project deletion remains allowed
- **WHEN** a covered agent requests deletion of a path within the active project and no protected target is included
- **THEN** OpenCode allows the command without asking for approval

#### Scenario: Temporary home deletion remains allowed
- **WHEN** a covered agent requests deletion of a path beneath `~/tmp` and no protected target is included
- **THEN** OpenCode allows the command without asking for approval

### Requirement: Secret reads are denied with template exceptions
The system SHALL deny the read tool access to `.env` files, credential files and stores, SSH material, keychains, secret directories, and `.pem` or `.key` files for every covered agent. The system SHALL allow read-tool access to `.env.example` and `.env.template`. This requirement SHALL NOT deny editing or deletion of those files.

#### Scenario: Environment secret read is denied
- **WHEN** a covered agent uses the read tool on `.env` or a non-template `.env.*` file at any depth
- **THEN** OpenCode denies the read

#### Scenario: Credential material read is denied
- **WHEN** a covered agent uses the read tool on configured credential, SSH, keychain, secret, `.pem`, or `.key` paths
- **THEN** OpenCode denies the read

#### Scenario: Environment examples remain readable
- **WHEN** a covered agent uses the read tool on `.env.example` or `.env.template`
- **THEN** OpenCode allows the read without asking for approval

#### Scenario: Secret modification is outside the read restriction
- **WHEN** a covered agent edits or deletes a secret path without first using the read tool
- **THEN** this policy does not deny the operation on the basis of the secret-read rules

### Requirement: Native-permission scope is explicit
The system SHALL implement these controls only with native OpenCode permissions and SHALL document that Bash, wrappers, aliases, scripts, alternate tools, and indirection can bypass command- or read-oriented protections.

#### Scenario: Indirect bypass is not represented as prevented
- **WHEN** an operation reaches protected content or behavior through an unrecognized wrapper or Bash indirection
- **THEN** the documented policy identifies that behavior as outside its enforcement guarantee

#### Scenario: Explicit orchestrator permissions are not overridden by deprecated tools
- **WHEN** the `angel-orchestrator` agent declares the canonical permission matrix
- **THEN** its repository and live definitions omit the deprecated `tools:` block so OpenCode does not append a catch-all allow after specific denials

### Requirement: Installer and live configuration are updated in place
The system SHALL distribute the policy through the installer assets and SHALL apply the same policy to the current global OpenCode configuration. The repository and live `angel-orchestrator` definitions SHALL omit the deprecated `tools:` block, preserve the complete permission matrix, and be byte-identical after synchronization. The one-time update SHALL NOT create a pre-change backup of the current global configuration.

#### Scenario: Fresh installer selection carries the policy
- **WHEN** the relevant agent and configuration assets are installed
- **THEN** the installed covered agents contain the ordered near-YOLO permission rules

#### Scenario: Current global configuration is updated without backup
- **WHEN** this change is applied to the current `~/.config/opencode/` installation
- **THEN** the covered live agents receive the policy and no one-off pre-change backup is created

#### Scenario: Orchestrator repair is synchronized before validation
- **WHEN** the deprecated `tools:` block is removed from the repository orchestrator asset
- **THEN** the live global orchestrator copy is synchronized byte-for-byte before remaining permission probes are completed

### Requirement: Validation is manual only
The change SHALL be validated with manual configuration inspection and representative permission checks, and implementation SHALL NOT add or modify automated tests for this behavior.

#### Scenario: Implementation validation completes
- **WHEN** the orchestrator repair has been synchronized, OpenCode has been restarted or reloaded, and the policy has been applied to repository assets and the current global configuration
- **THEN** the implementer reruns and records manual evidence for all representative allowed and denied cases without adding or modifying automated tests
