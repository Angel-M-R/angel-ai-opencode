## Why

The three review agents are currently unable to use Bash, so they cannot inspect Git state, search the repository through command-line tooling, or independently run the checks needed to substantiate review findings. Their installed permissions and instructions must allow evidence-gathering while retaining their strictly read-only role and existing secret protections.

## What Changes

- Grant `review-correctness`, `review-security-risk`, and `review-simplicity` the ability to inspect Git state and read/search repository files.
- Permit those three agents to run tests and linters, including commands that need network access, local services, or local artifacts.
- Keep tracked-file changes, staging, commits, pushes, secret access, and other write-oriented behavior prohibited for the three review agents.
- Require review reports to include each executed validation command and its exit code alongside findings or a clean result.
- Update the distributable agent configuration and reviewer instructions only; do not change other agents.

## Capabilities

### New Capabilities
- `read-only-review-agent-validation`: Defines evidence-gathering permissions, immutable review boundaries, and command/exit-code reporting for the three named review agents.

### Modified Capabilities
- `safe-yolo-agent-permissions`: Clarifies the explicit permission coverage and secret-read protections for Bash-capable review agents without expanding the near-YOLO policy to other agents.

## Impact

- `assets/agents/review-correctness.md`
- `assets/agents/review-security-risk.md`
- `assets/agents/review-simplicity.md`
- Installer or permission configuration that distributes and enforces those reviewer permissions, if required by the existing configuration model.
- Relevant installer/configuration validation and documentation, limited to the three named review agents.
