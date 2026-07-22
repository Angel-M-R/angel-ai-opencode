## Context

`review-correctness`, `review-security-risk`, and `review-simplicity` are installed as individual assets under `assets/agents/`. Each currently exposes the read tool but disables Bash and contains an instruction-only read-only contract. The installer copies selected agent assets unchanged into the OpenCode configuration directory, while `assets/fragments/permissions.json` supplies global permission defaults for separately selected configuration.

The existing near-YOLO specification deliberately excludes read-only reviewers because they did not expose Bash. This change makes only the three named reviewers Bash-capable, so their permission matrix and instructions must explicitly preserve the user-confirmed read-only boundary. The policy must continue denying secret reads and must not claim that native permissions are a complete sandbox against Bash indirection.

## Goals / Non-Goals

**Goals:**
- Give only the three named review agents explicit, usable permissions for Git inspection, repository file reading/searching, and test/linter execution.
- Permit validation commands even when they use network access, local services, or local artifacts.
- Deny tracked-file modification, Git staging, commits, pushes, and secret reads for those agents.
- Make command and exit-code evidence a mandatory part of reviewer reports.
- Preserve the existing installation model so selected reviewer assets deliver the same behavior to installed configurations.

**Non-Goals:**
- Changing permissions, tools, instructions, or scope for any agent other than the three named reviewers.
- Adding an implementation sandbox, shell parser, wrapper detector, secret scanner, or credential injection system.
- Allowing reviewers to write source, configuration, test, or generated tracked files; stage changes; commit; or push.
- Changing the global near-YOLO defaults or relaxing the existing secret-read policy for other agents.

## Decisions

### Give each reviewer an explicit ordered permission matrix

Each of the three reviewer assets will enable Bash and declare the same explicit native permission matrix rather than relying on global defaults. The matrix will allow read/search and general test/linter commands, including network- and service-dependent validation, then deny tracked-file write operations and the direct Git mutations: staging, commits, and pushes. It will repeat the established secret-read denials with the existing `.env.example` and `.env.template` exceptions.

This makes reviewer coverage auditable and keeps their restrictions effective when the agent asset is installed without the optional global permission fragment. It also scopes the policy to exactly the named agents.

Alternative considered: put reviewer permissions only in `assets/fragments/permissions.json`. Rejected because fragments are separately selectable, and a reviewer asset must enforce its own immutable role regardless of whether a global fragment is installed.

### Preserve a report-only reviewer contract while allowing validation side effects outside tracked files

Reviewer instructions will explicitly authorize inspection commands and test/linter execution, including operations that use the network, local services, or untracked local artifacts. They will prohibit using Bash or any other tool to alter tracked files, stage, commit, push, or access secrets, and will state that reviewers report evidence only.

This distinguishes permitted validation environment activity from repository mutation, as required by the brief.

Alternative considered: forbid any command that could write temporary or untracked outputs. Rejected because common test and linter workflows require caches, artifacts, local services, or network access, and the confirmed requirement permits them.

### Standardize command and exit-code evidence in each reviewer output contract

Every reviewer instruction will require a concise validation-evidence section listing each command actually run and its exit code. Commands not run will not be invented, and failed commands will be reported with their non-zero code without triggering fixes.

This lets the orchestrator and user distinguish review conclusions supported by executed validation from static inspection while maintaining the report-only role.

Alternative considered: document evidence reporting only in a shared global guide. Rejected because these reviewers can be installed independently and their agent-local output contracts are the authoritative reviewer behavior.

### Keep the installer distribution mechanism unchanged

The existing catalog and installer already copy each selected agent asset to the matching OpenCode agent location unchanged. Implementation will update only the three selected assets and add narrowly scoped validation of their frontmatter/instruction contracts if the repository’s existing tests cover agent assets; no installer behavior change is needed unless validation reveals a delivery gap.

## Risks / Trade-offs

- [Native permissions cannot fully prevent Bash indirection or wrappers from bypassing command/read rules] → Retain and reference the established native-permission limitation; instructions prohibit prohibited actions but do not misrepresent enforcement as a sandbox.
- [Broad test commands can create local untracked artifacts or contact external systems] → Permit this explicitly as required, while prohibiting tracked-file modifications and Git mutations and requiring command/exit-code reporting.
- [Permission pattern ordering could accidentally override a denial] → Use explicit ordered matrices per reviewer and validate representative allowed inspection/test commands and denied write/Git/secret operations.
- [Duplicated matrices across three agent assets can drift] → Use a single documented canonical reviewer matrix in the change design/tasks and add contract tests or equivalent assertions limited to those three assets.

## Migration Plan

1. Update the three repository reviewer assets with their explicit Bash, read, edit/write, and Git permission boundaries plus revised output contracts.
2. Add or update narrow validation that confirms all three assets expose the same allowed/denied reviewer contract and command/exit-code reporting requirement.
3. Run the relevant repository tests and lint/format checks; record command results.
4. Install or synchronize the selected reviewer assets through the existing installer, reload OpenCode if required, and manually probe representative allowed and denied operations.
5. Roll back by reverting only the three reviewer asset changes and their scoped validation changes, then reinstall the prior assets.

## Open Questions

None. The confirmed brief fixes the agent scope, allowed validation capabilities, prohibited mutations and secret access, and reporting requirement.
