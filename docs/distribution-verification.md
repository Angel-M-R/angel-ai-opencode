# Distribution Verification Handoff

This is an execution handoff, not a validation result. Every gate below is
`NOT RUN` until the final `openspec-verifier` executes it after all OpenSpec
tasks for `add-versioned-cli-distribution` are complete. Preparing or reviewing
this document does not pass a gate.

## Fixed inputs and environment

The verifier must use one clean, committed source tree that is the exact commit
proposed for `v0.1.0`. Run every command from that tree on macOS 14 Apple
Silicon (`Darwin/arm64`) with the Go version declared by `go.mod` (`go1.25.3`).
Required system tools are `/bin/bash`, `/bin/sh`, `cat`, `chmod`, `cmp`, `cp`,
`curl`, `file`, `git`, `grep`, `mkdir`, `mktemp`, `plutil`, `printf`, `script`,
and `shasum`.

Initialize the verifier shell exactly once:

```bash
export REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"
export RC_VERSION="v0.1.0"
export ARTIFACT_NAME="angel-ai-darwin-arm64"
export EVIDENCE_ROOT="$REPO_ROOT/.artifacts/distribution-verification/$RC_VERSION"
test ! -e "$EVIDENCE_ROOT"
mkdir -p "$EVIDENCE_ROOT"/{metadata,unit,updater-integration,installer-integration,build,smoke,release-candidate}
git rev-parse HEAD >"$EVIDENCE_ROOT/metadata/source-commit.txt"
git status --porcelain=v1 >"$EVIDENCE_ROOT/metadata/worktree-status.txt"
test ! -s "$EVIDENCE_ROOT/metadata/worktree-status.txt"
uname -s >"$EVIDENCE_ROOT/metadata/uname-s.txt"
uname -m >"$EVIDENCE_ROOT/metadata/uname-m.txt"
go env GOVERSION >"$EVIDENCE_ROOT/metadata/go-version.txt"
test "$(cat "$EVIDENCE_ROOT/metadata/uname-s.txt")" = "Darwin"
test "$(cat "$EVIDENCE_ROOT/metadata/uname-m.txt")" = "arm64"
test "$(cat "$EVIDENCE_ROOT/metadata/go-version.txt")" = "go1.25.3"
```

The source commit, clean-tree record, host values, Go version, command logs, and
numeric status files under `EVIDENCE_ROOT` are the mandatory evidence. A gate
passes only when its status file contains `0` and the verifier has reviewed its
log. Run each gate block in the same Bash shell without `set -e`, so a failing
command can still have its status recorded.

## Mandatory full suites

### Full unit suite

```bash
go test ./... >"$EVIDENCE_ROOT/unit/output.txt" 2>&1
unit_status=$?
printf '%s\n' "$unit_status" >"$EVIDENCE_ROOT/unit/status.txt"
test "$unit_status" -eq 0
```

### Updater integration suite

```bash
go test ./internal/updater -count=1 \
  -run '^(TestUpdaterRejectsChecksumMismatchAndCleansTemporaryState|TestUpdaterAtomicallyReplacesAndRelaunchesWithOriginalProcessState|TestUpdaterRelaunchMarkerCompletesAutomaticTUIFlowWithoutLooping|TestUpdaterReplacementFailureRestoresCurrentExecutable|TestUpdaterRelaunchFailureRollsBackImmediately)$' \
  >"$EVIDENCE_ROOT/updater-integration/output.txt" 2>&1
updater_status=$?
printf '%s\n' "$updater_status" >"$EVIDENCE_ROOT/updater-integration/status.txt"
test "$updater_status" -eq 0
```

### Installer integration suite

```bash
/bin/sh tests/install_integration_test.sh \
  >"$EVIDENCE_ROOT/installer-integration/output.txt" 2>&1
installer_status=$?
printf '%s\n' "$installer_status" >"$EVIDENCE_ROOT/installer-integration/status.txt"
test "$installer_status" -eq 0
```

## Mandatory release build

The build output below is the only artifact accepted by the packaged-artifact
and release-candidate procedures. Do not rebuild between gates.

```bash
export ARTIFACT="$EVIDENCE_ROOT/build/$ARTIFACT_NAME"
(
  CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build \
    -trimpath \
    -ldflags="-s -w -X main.version=$RC_VERSION" \
    -o "$ARTIFACT" \
    . &&
  chmod 0755 "$ARTIFACT" &&
  reported_version=$(env \
    HTTP_PROXY=http://127.0.0.1:9 \
    HTTPS_PROXY=http://127.0.0.1:9 \
    ALL_PROXY=http://127.0.0.1:9 \
    NO_PROXY= \
    "$ARTIFACT" version) &&
  test "$reported_version" = "$RC_VERSION"
) >"$EVIDENCE_ROOT/build/output.txt" 2>&1
build_status=$?
printf '%s\n' "$build_status" >"$EVIDENCE_ROOT/build/status.txt"
test "$build_status" -eq 0
```

Do not approve release readiness from these four gates alone. Exact-artifact
smoke and release-candidate gates below are independently mandatory.

## Exact packaged-artifact smoke

Run this only after all preceding status files contain `0`. It packages metadata
for the already-built artifact, then checks those exact bytes. The manifest URL
is the immutable `v0.1.0` release URL; no latest-release or GitHub API lookup is
an input to this gate.

```bash
export REPOSITORY="Angel-M-R/angel-ai-opencode"
export ARTIFACT="$EVIDENCE_ROOT/build/$ARTIFACT_NAME"
export CHECKSUM="$EVIDENCE_ROOT/build/$ARTIFACT_NAME.sha256"
export MANIFEST="$EVIDENCE_ROOT/build/manifest.json"
export ARTIFACT_URL="https://github.com/$REPOSITORY/releases/download/$RC_VERSION/$ARTIFACT_NAME"
digest=$(shasum -a 256 "$ARTIFACT")
digest=${digest%% *}
printf '%s  %s\n' "$digest" "$ARTIFACT_NAME" >"$CHECKSUM"
printf '{"version":"%s","artifact_url":"%s","sha256":"%s"}\n' \
  "$RC_VERSION" "$ARTIFACT_URL" "$digest" >"$MANIFEST"

(
  set -euo pipefail
  cd "$EVIDENCE_ROOT/build"
  test "$(uname -s)" = "Darwin"
  test "$(uname -m)" = "arm64"
  file_output=$(file "$ARTIFACT_NAME")
  [[ "$file_output" == *"Mach-O 64-bit executable arm64"* ]]
  shasum -a 256 -c "$ARTIFACT_NAME.sha256"
  test "$(plutil -extract version raw -o - manifest.json)" = "$RC_VERSION"
  test "$(plutil -extract artifact_url raw -o - manifest.json)" = "$ARTIFACT_URL"
  test "$(plutil -extract sha256 raw -o - manifest.json)" = "$digest"
  test "$(plutil -convert xml1 -o - manifest.json | grep -c '<key>')" -eq 3

  version_output=$(env \
    HTTP_PROXY=http://127.0.0.1:9 \
    HTTPS_PROXY=http://127.0.0.1:9 \
    ALL_PROXY=http://127.0.0.1:9 \
    NO_PROXY= \
    "$ARTIFACT" version)
  test "$version_output" = "$RC_VERSION"

  embedded_cwd=$(mktemp -d "$EVIDENCE_ROOT/smoke/embedded-cwd.XXXXXX")
  embedded_target=$(mktemp -d "$EVIDENCE_ROOT/smoke/embedded-target.XXXXXX")
  (
    cd "$embedded_cwd"
    "$ARTIFACT" --all --dry-run --target "$embedded_target" \
      >"$EVIDENCE_ROOT/smoke/embedded-assets.txt"
  )
  test -s "$EVIDENCE_ROOT/smoke/embedded-assets.txt"

  tui_cwd=$(mktemp -d "$EVIDENCE_ROOT/smoke/tui-cwd.XXXXXX")
  (
    cd "$tui_cwd"
    (sleep 2; printf 'q') | env \
      ANGEL_AI_UPDATE_RELAUNCHED=1 \
      TERM=xterm-256color \
      script -q "$EVIDENCE_ROOT/smoke/tui-transcript.txt" "$ARTIFACT"
  )
  grep -q 'Angel AI' "$EVIDENCE_ROOT/smoke/tui-transcript.txt"
) >"$EVIDENCE_ROOT/smoke/output.txt" 2>&1
smoke_status=$?
printf '%s\n' "$smoke_status" >"$EVIDENCE_ROOT/smoke/status.txt"
test "$smoke_status" -eq 0
```

## `v0.1.0` release candidate

This gate runs the documented one-line command in an isolated home while a
PATH-local `curl` fixture serves `install.sh`, the generated manifest, and the
same artifact bytes already checked above. It makes no external request and
does not modify a real home directory. The fixture must reject every unexpected
URL so the command's complete input set remains deterministic.

```bash
export ONE_LINE_INSTALLER_URL="https://raw.githubusercontent.com/Angel-M-R/angel-ai-opencode/main/install.sh"
export LATEST_MANIFEST_URL="https://github.com/Angel-M-R/angel-ai-opencode/releases/latest/download/manifest.json"
export RC_ROOT=$(mktemp -d "$EVIDENCE_ROOT/release-candidate/work.XXXXXX")
export RC_HOME="$RC_ROOT/home"
export RC_SHIMS="$RC_ROOT/shims"
mkdir -p "$RC_HOME" "$RC_SHIMS"
printf 'zsh sentinel\n' >"$RC_HOME/.zshrc"
printf 'profile sentinel\n' >"$RC_HOME/.profile"

cat >"$RC_SHIMS/curl" <<'SH'
#!/bin/sh
set -eu
output=""
url=""
while [ "$#" -gt 0 ]; do
  case "$1" in
    --output) output=$2; shift 2 ;;
    --*) shift ;;
    *) url=$1; shift ;;
  esac
done
case "$url" in
  "$VERIFY_ONE_LINE_URL") source=$VERIFY_INSTALLER ;;
  "$VERIFY_MANIFEST_URL") source=$VERIFY_MANIFEST ;;
  "$VERIFY_ARTIFACT_URL") source=$VERIFY_ARTIFACT ;;
  *) printf 'unexpected curl URL: %s\n' "$url" >&2; exit 22 ;;
esac
if [ -n "$output" ]; then
  cp "$source" "$output"
else
  cat "$source"
fi
SH
chmod 0755 "$RC_SHIMS/curl"

(
  set -euo pipefail
  export HOME="$RC_HOME"
  export PATH="$RC_SHIMS:/usr/bin:/bin"
  export TMPDIR="$RC_ROOT"
  export VERIFY_ONE_LINE_URL="$ONE_LINE_INSTALLER_URL"
  export VERIFY_MANIFEST_URL="$LATEST_MANIFEST_URL"
  export VERIFY_ARTIFACT_URL="$ARTIFACT_URL"
  export VERIFY_INSTALLER="$REPO_ROOT/install.sh"
  export VERIFY_MANIFEST="$MANIFEST"
  export VERIFY_ARTIFACT="$ARTIFACT"

  curl --proto '=https' --tlsv1.2 -fsSL https://raw.githubusercontent.com/Angel-M-R/angel-ai-opencode/main/install.sh | /bin/sh

  installed="$RC_HOME/.local/bin/angel-ai"
  test -x "$installed"
  cmp "$installed" "$ARTIFACT"
  test "$(cat "$RC_HOME/.zshrc")" = "zsh sentinel"
  test "$(cat "$RC_HOME/.profile")" = "profile sentinel"

  installed_version=$(env \
    HTTP_PROXY=http://127.0.0.1:9 \
    HTTPS_PROXY=http://127.0.0.1:9 \
    ALL_PROXY=http://127.0.0.1:9 \
    NO_PROXY= \
    "$installed" version)
  test "$installed_version" = "$RC_VERSION"

  installed_cwd=$(mktemp -d "$RC_ROOT/installed-cwd.XXXXXX")
  installed_target=$(mktemp -d "$RC_ROOT/installed-target.XXXXXX")
  (
    cd "$installed_cwd"
    "$installed" --all --dry-run --target "$installed_target" \
      >"$EVIDENCE_ROOT/release-candidate/embedded-assets.txt"
  )
  test -s "$EVIDENCE_ROOT/release-candidate/embedded-assets.txt"

  tui_cwd=$(mktemp -d "$RC_ROOT/tui-cwd.XXXXXX")
  (
    cd "$tui_cwd"
    (sleep 2; printf 'q') | env \
      ANGEL_AI_UPDATE_RELAUNCHED=1 \
      TERM=xterm-256color \
      script -q "$EVIDENCE_ROOT/release-candidate/tui-transcript.txt" "$installed"
  )
  grep -q 'Angel AI' "$EVIDENCE_ROOT/release-candidate/tui-transcript.txt"
) >"$EVIDENCE_ROOT/release-candidate/output.txt" 2>&1
candidate_status=$?
printf '%s\n' "$candidate_status" >"$EVIDENCE_ROOT/release-candidate/status.txt"
test "$candidate_status" -eq 0
```

The final verifier must review the generated manifest and checksum, every log,
the embedded-assets outputs, both TUI transcripts, and all six numeric status
files before recording gate results. It must not create a release or tag.

## Gate record

The final verifier must create
`$EVIDENCE_ROOT/final-verifier-summary.md` containing the source commit and one
explicit `PASS` or `FAIL` row for each gate. Until execution, all rows are
`NOT RUN`: full unit suite, updater integration suite, installer integration
suite, release build, packaged-artifact smoke, and `v0.1.0` release candidate.
