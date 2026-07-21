## Context

The current `assets/fragments/permissions.json` policy allows most Bash commands but still asks for routine Git operations, while only the orchestrator has an agent-local permission block. The OpenSpec planner, implementer, and verifier declare Bash and read tools without the agreed restrictions, and the built-in `general` agent is not overridden by the installer. OpenCode evaluates granular rules in order with the last matching rule winning, and agent-local rules take precedence over global rules.

OpenCode 1.18.4 runtime validation subsequently showed that `git reset --hard` remained allowed after a fresh subprocess, refuting the hypothesis that the process merely needed a restart. The repository `angel-orchestrator` asset combines the deprecated `tools:` frontmatter with `permission:`. OpenCode seeds permissions from `tools`, merges the explicit permissions with `Object.assign`, and appends a catch-all allow after the specific denials; last-match evaluation therefore reopens Bash, read, and task operations that the matrix intends to deny.

This is a security-sensitive, cross-cutting configuration change. It must use only OpenCode's native permissions, update both distributed assets and the currently installed global configuration, create no one-off backup before those edits, and add no automated tests.

## Goals / Non-Goals

**Goals:**

- Let routine Bash and read operations execute without approval prompts for the explicitly covered agents.
- Deny the confirmed direct destructive Git commands, direct destructive deletion targets, and read-tool access to secret material.
- Keep ordinary Git, project deletion, `~/tmp` deletion, and template environment-file reads available.
- Make coverage explicit in each currently managed Bash-capable agent and in the built-in `general` override.
- Apply the same policy to the repository's installable assets and the current global OpenCode installation.

**Non-Goals:**

- Building a shell parser, custom permission plugin, wrapper detector, or sandbox.
- Preventing secret access through Bash, scripts, symlinks, custom tools, or indirect command execution.
- Blocking edits or deletion of secret files.
- Automatically enrolling future agents in the policy.
- Adding automated tests or changing the generic installer backup mechanism.

## Decisions

### Use an ordered native permission policy

Each covered agent will start from allow-by-default behavior, then apply more specific Bash and read denials, followed by the narrow read exceptions for `.env.example` and `.env.template`. This order is required because OpenCode uses the last matching rule. Ordinary Git commands are not enumerated because the catch-all already permits them.

The Bash denial matrix will cover direct command-string variants for:

- `git reset` containing `--hard`, with or without a revision and with direct Git executable/options forms represented by the native wildcard patterns.
- `git push` containing `--force`, `--force-with-lease`, or the standalone short flag `-f`, including arguments before or after the dangerous flag.
- `rm` forms whose explicit operand is `/`, `~`, `$HOME`, or a named critical system root and its descendants. The critical roots will include the conventional macOS/Linux system locations `/System`, `/Library`, `/Applications`, `/bin`, `/sbin`, `/usr`, `/etc`, `/var`, `/private`, `/opt`, `/dev`, `/proc`, `/sys`, and `/boot`.

The matrix will not deny arbitrary absolute paths, `~/*`, or `$HOME/*`; doing so would also block project deletion and `~/tmp`. Alternative approaches such as globally asking for Bash approval or denying all `rm`/Git push commands were rejected because they contradict near-YOLO operation.

### Repeat the policy at each explicit agent boundary

The policy will be present in the Markdown frontmatter for `angel-orchestrator`, `openspec-planner`, `openspec-implementer`, and `openspec-verifier`, and in an `agent.general.permission` override distributed through the OpenCode JSON fragment. This makes agent coverage auditable and prevents a future agent from silently inheriting a safety claim it was never reviewed against. Read-only reviewers remain unchanged because they do not expose Bash.

The repetition is intentional. A shared global policy alone was rejected because the confirmed rule requires future Bash-capable agents to be added explicitly.

### Remove the deprecated orchestrator tool declaration before revalidation

The repository `assets/agents/angel-orchestrator.md` definition will remove its deprecated `tools:` block and retain the complete explicit `permission:` matrix. Its installed `~/.config/opencode/agents/angel-orchestrator.md` counterpart will then be synchronized to the repository asset byte-for-byte. No other repair is accepted for this runtime failure because removing or weakening matrix entries, adding a second workaround layer, or relying on restart alone would leave the permission source ambiguous or contradict the confirmed policy.

After both copies are repaired and synchronized, OpenCode must be restarted or reloaded before all representative allowed and denied probes are rerun. The repair is a prerequisite for completing the remaining manual-validation tasks, not a replacement for them.

### Restrict secret protection to the read tool

Read rules will deny `.env` variants, credential stores/files, SSH content, keychains, secret directories, and `.pem`/`.key` files across root and nested paths. Later allow rules will restore `.env.example` and `.env.template`. Edit remains allowed, and Bash remains capable of indirect reads, as explicitly accepted.

### Deploy assets and the live configuration without a one-off backup

Implementation will update the repository assets first, then mirror the resulting agent and JSON permission configuration into `~/.config/opencode/`. It will not create a pre-change backup of the live configuration. The installer's existing general reconciliation backup behavior is outside this change; changing it would broaden the risk beyond the confirmed request.

### Validate manually instead of adding tests

Validation will inspect the final configuration and exercise representative allowed and denied operations through OpenCode for each covered agent. No Go tests, fixtures, or other automated test assets will be added or modified for this change.

## Risks / Trade-offs

- **Native command patterns can be bypassed by wrappers, aliases, shell indirection, alternate tools, or transformed command strings** → State this limitation in the policy documentation and validate only the direct forms promised by the spec.
- **Repeated agent-local matrices can drift** → Use one canonical matrix while implementing, copy it exactly to each covered agent, and compare all copies during manual validation.
- **Over-broad `rm` patterns could block legitimate project work** → Deny only explicit protected operands and named critical roots; manually verify project paths and `~/tmp` remain allowed.
- **Rule ordering or deprecated `tools:` expansion could reopen a denial or block template reads** → Do not combine the orchestrator's explicit matrix with deprecated `tools:` frontmatter; keep permission catch-all allows first, denials next, and the two environment-template exceptions last; inspect resolved ordering in every target.
- **Editing the live global configuration without backup reduces rollback safety** → Keep the edit narrowly scoped and make rollback a direct reapplication of the previous policy from version control or a manual reversal, without creating a backup artifact.

## Migration Plan

1. Define the canonical ordered Bash/read matrix in the installable permission configuration.
2. Add an identical agent-local policy to every managed Bash-capable Markdown agent and an explicit override for `general`.
3. Update the currently installed global agent files and `opencode.json` in place, without creating a backup.
4. Remove the deprecated `tools:` block from the repository orchestrator asset while preserving its complete permission matrix, then synchronize the live orchestrator copy byte-for-byte.
5. Restart or reload OpenCode, manually inspect the effective files, and rerun all representative allowed and denied scenarios.
6. Roll back, if necessary, by reverting the repository changes and manually restoring the prior global permission entries; do not rely on a generated backup.

## Open Questions

None. The accepted bypass limitations, explicit agent set, protected operations, deployment scope, backup choice, and manual-only validation are confirmed.
