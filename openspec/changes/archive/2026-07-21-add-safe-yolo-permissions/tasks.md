## 1. Installer-Distributed Policy

- [x] 1.1 Replace the approval-oriented permission fragment with the canonical ordered Bash/read matrix and an explicit `agent.general.permission` override, including the environment-template allow exceptions.
- [x] 1.2 Add the identical ordered policy to the frontmatter of `angel-orchestrator`, `openspec-planner`, `openspec-implementer`, and `openspec-verifier`; leave agents without Bash unchanged.
- [x] 1.3 Document in the installed global guidance that the policy covers only direct native permission matches, that Bash/wrappers can bypass it, and that each future Bash-capable agent must be added explicitly.

## 2. Current Global OpenCode Configuration

- [x] 2.1 Update the installed `~/.config/opencode/agents/` definitions for the four managed Bash-capable agents to match the repository assets exactly, without creating a pre-change backup.
- [x] 2.2 Update `~/.config/opencode/opencode.json` in place with the explicit `general` override and near-YOLO rule ordering, preserving unrelated user configuration and creating no pre-change backup.
- [x] 2.3 Update the installed global `AGENTS.md` policy note to match the distributed guidance, without creating a pre-change backup.
- [x] 2.4 Remove the deprecated `tools:` block from `assets/agents/angel-orchestrator.md` while preserving its complete explicit permission matrix; do not change any other agent behavior as part of this repair.
- [x] 2.5 Synchronize the repaired repository orchestrator asset to `~/.config/opencode/agents/angel-orchestrator.md` byte-for-byte, then restart or reload OpenCode before further permission probes.

## 3. Manual Validation Only

- [x] 3.1 Manually inspect the repository and live configurations to confirm all five covered agents contain equivalent ordered rules, template exceptions come after secret denials, and no Bash-capable managed agent was omitted.
- [x] 3.2 After tasks 2.4-2.5, rerun all representative safe cases and direct denied cases through the restarted or reloaded OpenCode runtime using disposable paths/repositories where needed; confirm normal Git, project deletion, `~/tmp` deletion, and template reads are allowed while hard reset, force push, protected-target deletion, and secret reads are denied.
- [x] 3.3 After tasks 2.4-2.5 and the rerun in 3.2, confirm the implementation adds or modifies no automated tests, test fixtures, or test infrastructure; record only the manual commands/checks and their outcomes.
