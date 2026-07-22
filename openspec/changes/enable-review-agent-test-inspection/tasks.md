## 1. Reviewer permission and instruction updates

- [x] 1.1 Update only `assets/agents/review-correctness.md`, `assets/agents/review-security-risk.md`, and `assets/agents/review-simplicity.md` to enable Bash and add the same explicit native reviewer permission matrix.
- [x] 1.2 Configure that matrix to allow Git inspection, non-secret file reading/searching, and tests/linters that require network access, local services, or local artifacts.
- [x] 1.3 Configure that matrix and the three reviewer instructions to deny tracked-file changes, staging, commits, pushes, and secret reads while retaining the established environment-template read exceptions.
- [x] 1.4 Update each reviewer output contract to list every executed validation command and exit code with either findings or `No findings.`, and to report failures without attempting fixes.

## 2. Distribution-contract validation

- [x] 2.1 Add or update narrowly scoped agent-asset validation to assert the identical allowed/denied reviewer contract and command/exit-code reporting requirement across exactly the three named reviewer assets.
- [x] 2.2 Verify the existing catalog/installer copies each selected reviewer asset unchanged; change installer code only if that verification identifies a delivery gap.

## 3. Verification and rollout

- [x] 3.1 Run the relevant Go tests plus repository formatting/lint checks, and record each command with its exit code.
  - `go test ./...` — exit 0
  - `gofmt -d $(git ls-files '*.go') && test -z "$(gofmt -d $(git ls-files '*.go'))"` — exit 0
  - `go vet ./...` — exit 0
- [x] 3.2 Install or synchronize only the selected reviewer assets through the existing installer, reload OpenCode if needed, and manually probe allowed Git/file/test operations and denied tracked-file, Git-mutation, and secret-read operations.
  - Interactive `go run . --target <temporary OpenCode config>` selected only the three reviewer agents — exit 0.
  - Fresh `OPENCODE_CONFIG_DIR=<temporary OpenCode config> opencode agent list` reloaded the installed definitions — exit 0.
  - Installed assets matched their sources with `cmp -s` for all three reviewers — exit 0 each.
  - The OpenCode policy probe allowed `git status --short`, `README.md` read, and the targeted Go test (exit 0); it denied `git add --dry-run`, `git commit --dry-run`, `git push --dry-run`, and `.env` read before execution.
- [x] 3.3 Confirm the three reviewer reports preserve their discipline-specific findings format and include the executed-command/exit-code evidence section.
  - `go test ./internal/install -run '^(TestReviewerAssetsShareReadOnlyValidationContract|TestSelectedReviewerAssetsAreCatalogedAndInstalledUnchanged)$' -count=1` — exit 0.
  - `rg` checks confirmed the correctness, security-risk, and simplicity finding-format markers plus the required Validation evidence section — exit 0.
