## ADDED Requirements

### Requirement: Bash-capable reviewers have explicit limited coverage
The system SHALL apply an explicit native permission matrix to `review-correctness`, `review-security-risk`, and `review-simplicity` when their Bash tools are enabled. The matrix SHALL permit their defined read-only inspection and validation operations while denying tracked-file changes, staging, commits, pushes, and secret reads. This coverage SHALL apply only to these three reviewers and SHALL NOT expand the near-YOLO policy to other agents.

#### Scenario: Reviewer permissions are installed independently
- **WHEN** any of the three named reviewer assets is selected for installation without a global permission fragment
- **THEN** its installed definition retains the explicit reviewer permission matrix and read-only restrictions

#### Scenario: Other agents remain out of scope
- **WHEN** the reviewer permission matrix is added to the three named reviewer definitions
- **THEN** no other agent receives expanded reviewer capabilities or changed permission coverage from this change
