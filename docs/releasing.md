# Stable Release Handoff

These instructions are for the maintainer who manually creates the first stable
tag. They do not authorize a release until the final `openspec-verifier` has
executed `docs/distribution-verification.md` and recorded every mandatory gate
as successful. Section 7 implementation prepared this handoff but did not run
or pass any gate.

## Required handoff

For `v0.1.0`, obtain the complete
`.artifacts/distribution-verification/v0.1.0` evidence directory from the final
verifier. It must come from a clean macOS Apple Silicon checkout of the exact
commit to tag and contain:

- source commit, clean worktree, host, and Go-version metadata;
- logs and zero-valued status files for the full unit, updater integration,
  installer integration, release build, packaged-artifact smoke, and release
  candidate gates;
- the exact binary, SHA-256 file, and manifest used by both artifact gates;
- embedded-assets output and TUI transcripts from the packaged artifact and
  isolated one-line installation;
- `final-verifier-summary.md` with the source commit followed by exactly these
  six success records:

```text
PASS: full unit suite
PASS: updater integration suite
PASS: installer integration suite
PASS: release build
PASS: packaged-artifact smoke
PASS: v0.1.0 release candidate
```

Any missing evidence, absent/skipped gate, non-zero status, `FAIL`, `NOT RUN`,
different source commit, or unexplained rerun blocks tagging. Handoff
preparation and focused documentation checks are not substitutes.

## Release authorization check

From the clean committed source tree, point `EVIDENCE_ROOT` at the verifier's
unchanged evidence directory and run this read-only authorization check before
creating a tag:

```bash
export RC_VERSION="v0.1.0"
export EVIDENCE_ROOT="$(git rev-parse --show-toplevel)/.artifacts/distribution-verification/$RC_VERSION"
export SOURCE_COMMIT="$(cat "$EVIDENCE_ROOT/metadata/source-commit.txt")"

test -z "$(git status --porcelain=v1)"
test "$(git rev-parse HEAD)" = "$SOURCE_COMMIT"
test ! -s "$EVIDENCE_ROOT/metadata/worktree-status.txt"
for status_file in \
  "$EVIDENCE_ROOT/unit/status.txt" \
  "$EVIDENCE_ROOT/updater-integration/status.txt" \
  "$EVIDENCE_ROOT/installer-integration/status.txt" \
  "$EVIDENCE_ROOT/build/status.txt" \
  "$EVIDENCE_ROOT/smoke/status.txt" \
  "$EVIDENCE_ROOT/release-candidate/status.txt"
do
  test "$(cat "$status_file")" = "0"
done
for required_result in \
  'PASS: full unit suite' \
  'PASS: updater integration suite' \
  'PASS: installer integration suite' \
  'PASS: release build' \
  'PASS: packaged-artifact smoke' \
  'PASS: v0.1.0 release candidate'
do
  grep -Fxq "$required_result" "$EVIDENCE_ROOT/final-verifier-summary.md"
done
```

The maintainer must also review all logs, manifest fields, checksum evidence,
and transcripts rather than relying only on the numeric checks.

## Create the immutable tag manually

Confirm the verified commit is the intended `main` commit and that `v0.1.0`
does not exist locally or remotely. A network error while checking the remote is
not proof that the tag is absent.

```bash
git fetch origin main --tags
test "$(git rev-parse origin/main)" = "$SOURCE_COMMIT"
test "$(git rev-parse HEAD)" = "$SOURCE_COMMIT"
test -z "$(git status --porcelain=v1)"
! git rev-parse --verify --quiet "refs/tags/$RC_VERSION"
remote_tag_status=0
git ls-remote --exit-code --tags origin "refs/tags/$RC_VERSION" >/dev/null 2>&1 || remote_tag_status=$?
test "$remote_tag_status" -eq 2
git tag -a "$RC_VERSION" "$SOURCE_COMMIT" -m "$RC_VERSION"
git push origin "refs/tags/$RC_VERSION"
```

The pushed tag is immutable: never force-push, delete, move, or reuse it. The
tag-triggered `Stable macOS Apple Silicon Release` workflow independently reruns
the full unit, updater integration, installer integration, build, and exact
artifact smoke gates. Publication must remain blocked unless every workflow job
succeeds. Verify that the resulting GitHub Release contains only
`angel-ai-darwin-arm64`, its `.sha256` file, `manifest.json`, and generated
notes, and that the release text says the macOS Apple Silicon artifact is not
Apple signed or notarized.

If pre-tag authorization fails, do not tag; fix the source and have the final
verifier rerun every invalidated gate against the new commit. If a failure is
found after pushing `v0.1.0`, preserve that tag, do not publish or repair it by
replacement, and prepare a newly verified stable patch tag such as `v0.1.1`.
