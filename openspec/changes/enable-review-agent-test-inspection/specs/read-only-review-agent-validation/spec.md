## ADDED Requirements

### Requirement: Named reviewers can gather validation evidence
The system SHALL grant `review-correctness`, `review-security-risk`, and `review-simplicity` explicit native permissions and instructions to inspect Git state, read and search non-secret repository files, and run tests and linters. The permissions SHALL allow such test and linter commands when they require network access, local services, or local artifacts.

#### Scenario: Reviewer inspects a change and repository context
- **WHEN** one of the three named reviewers needs Git metadata, a diff, or non-secret repository files to assess a change
- **THEN** the reviewer can execute the required inspection command and read or search the requested non-secret files

#### Scenario: Reviewer runs an environment-dependent validation command
- **WHEN** one of the three named reviewers runs a test or linter that requires network access, a local service, or local artifacts
- **THEN** the reviewer can execute that command and use its result as review evidence

### Requirement: Named reviewers remain read-only with Git mutation and secret protections
The system SHALL prohibit `review-correctness`, `review-security-risk`, and `review-simplicity` from changing tracked files, staging changes, committing, pushing, or accessing secrets. Their instructions SHALL require report-only behavior and SHALL NOT authorize fixes or configuration changes.

#### Scenario: Reviewer attempts a prohibited repository mutation
- **WHEN** one of the three named reviewers attempts to modify a tracked file, stage a change, create a commit, or push a branch
- **THEN** the system denies the action and the reviewer remains limited to reporting evidence

#### Scenario: Reviewer attempts to read secret material
- **WHEN** one of the three named reviewers requests a protected environment, credential, key, SSH, keychain, or secret path
- **THEN** the system denies the read while retaining the established `.env.example` and `.env.template` exceptions

### Requirement: Reviewer reports include executed-command results
The instructions for `review-correctness`, `review-security-risk`, and `review-simplicity` SHALL require each review report to list every validation command the reviewer actually ran with its exit code. This evidence SHALL accompany either findings or a clean result and SHALL NOT cause the reviewer to apply a fix after a failed command.

#### Scenario: Reviewer completes with findings
- **WHEN** a named reviewer runs one or more commands and reports findings
- **THEN** the report includes every executed command and its exit code in addition to the required finding details

#### Scenario: Reviewer completes with no findings
- **WHEN** a named reviewer runs one or more commands and finds no applicable issue
- **THEN** the report states `No findings.` and includes every executed command and its exit code

#### Scenario: Reviewer validation command fails
- **WHEN** a named reviewer runs a validation command that exits non-zero
- **THEN** the reviewer reports the command and its non-zero exit code without modifying files or attempting a fix
