## Why

Angel's current OpenCode permissions still prompt for routine Git operations and do not consistently protect every Bash-capable agent from the agreed destructive commands and secret reads. A near-YOLO policy is needed so normal work proceeds autonomously while a small, explicit set of high-impact operations remains blocked.

## What Changes

- Replace approval-oriented defaults with native OpenCode permission rules that allow ordinary operations without prompting.
- Apply the policy explicitly to every Bash-capable managed agent and to the built-in `general` agent; future Bash-capable agents must opt in explicitly.
- Deny direct hard resets and force pushes, destructive deletion of root, home, and critical system targets, and reads of environment files, credentials, SSH material, keychains, secrets, and private key files.
- Preserve normal Git usage, project-local deletion, deletion under `~/tmp`, and reads of `.env.example` and `.env.template`.
- Update the installer-distributed assets and the current global OpenCode configuration in place without creating a one-off pre-change backup.
- Document the intentional limitation that native command/read permissions can be bypassed through Bash indirection or wrappers.
- Validate the change manually only; do not add automated tests.

## Capabilities

### New Capabilities
- `safe-yolo-agent-permissions`: Defines the managed near-YOLO policy, its protected operations and paths, per-agent coverage, installer behavior, and manual validation expectations.

### Modified Capabilities

None.

## Impact

- Managed OpenCode permission fragments under `assets/fragments/`.
- Markdown definitions for the orchestrator, OpenSpec workers, and any other managed Bash-capable agents under `assets/agents/`.
- Installer-distributed permission and agent assets, plus any installer plumbing needed to deploy them.
- The installed global OpenCode configuration under `~/.config/opencode/`, including the built-in `general` agent override.
- No new dependencies and no automated test additions.
