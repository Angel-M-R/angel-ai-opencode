## 1. Shared CLI Inspection Model

- [x] 1.1 Separate each descriptor's registry package name, `@latest` install spec, executable, and version invocation while preserving CodeGraph-before-OpenSpec descriptor order and the OpenSpec Node.js floor.
- [x] 1.2 Extend npm-first/pnpm-fallback manager data with injected package-registration and latest-version probes whose parsing distinguishes registered, unregistered, unavailable, and malformed responses.
- [x] 1.3 Add an ordered inspection snapshot that classifies each selected CLI as install, current, outdated, ahead, or registry-unverified using the existing semantic-version comparison.
- [x] 1.4 Return manager- and CLI-specific guided recovery errors for registered-without-executable and failing, empty, or uninterpretable executable versions without cleanup or cross-manager inspection.

## 2. Planning and Application Flow

- [x] 2.1 Update shared preflight to preserve manager selection and Node validation order, then inspect every selected CLI before exposing any installable record.
- [x] 2.2 Render each inspection classification or pending `@latest` installation consistently in `PlanInstallation` without package or configuration mutation.
- [x] 2.3 Update `ApplyInstallation` to consume the complete preflight snapshot, install only absent unregistered CLIs, and verify each new executable and parseable version.
- [x] 2.4 Preserve descriptor processing order, no-write-on-CLI-failure boundaries, and exactly one post-CLI repreparation before configuration reconciliation when global CLIs are selected.
- [x] 2.5 Preserve healthy installed CLIs with a registry-unverified warning when latest lookup fails, while blocking an absent CLI's unusable registry lookup before any selected CLI installation.

## 3. Focused Injected Unit Regression

- [x] 3.1 Add injected tests for npm preference, validated pnpm fallback, manager-specific registration/latest probes, and proof that the unselected manager is never queried.
- [x] 3.2 Add injected CodeGraph and OpenSpec tests for current, outdated, ahead, absent-installable, and working-but-unregistered classifications in both plan and final output.
- [x] 3.3 Add injected tests for registry failure and malformed registry output, covering healthy-local continuation with warnings and absent-CLI blocking before installations.
- [x] 3.4 Add injected tests for registered packages missing executables and failing, empty, or malformed `--version` output, asserting guided errors, no cleanup, no installs, and no configuration writes.
- [x] 3.5 Add an injected multi-CLI regression proving all CLIs are prevalidated before the first install and preserving Node preflight, deterministic install order, one repreparation, and configuration-write boundaries.
- [x] 3.6 Run only the focused injected installer unit-test selection for these cases; do not execute real npm/pnpm commands or the full repository test suite.

## 4. Pending Final-Verification Corrections

- [x] 4.1 Update the two applicable fake npm fixtures in `internal/install/extras_test.go` to answer the new package-registration and registry-latest probes while preserving their existing installation and idempotency assertions. Prove the correction with `go test ./internal/install -run '^(TestInstallationConfiguresSelectedCodegraphIdempotently|TestInstallationInstallsCodegraphWhenMissing|TestInstallationComposesGlobalAgentsAndCodegraphIdempotently)$'`; the rerun MUST use only the tests' temporary fake executables and MUST NOT invoke real npm/pnpm or mutate global packages.
- [x] 4.2 Add a full injected `ApplyInstallation` regression in `internal/install/installation_global_cli_test.go` where an absent CLI passes prevalidation and package installation, then returns an uninterpretable post-install version; assert the application fails and writes no prepared configuration files. Prove the correction with `go test ./internal/install -run '^TestApplyInstallationRejectsInvalidPostInstallVersionWithoutConfigurationWrites$'`; do not run the full suite or perform global mutations.
