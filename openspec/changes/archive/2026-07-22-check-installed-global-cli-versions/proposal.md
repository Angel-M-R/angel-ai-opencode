## Why

The shared global CLI installer currently reinstalls every selected CodeGraph and OpenSpec CLI at `@latest`, even when a working executable is already available. This creates unnecessary global mutations and cannot distinguish a healthy current installation from an outdated, ahead-of-registry, or recoverably offline installation.

## What Changes

- Prevalidate every selected global CLI before installing any package, using only the selected npm-first/pnpm-fallback manager to inspect package registration and registry state.
- Install `@latest` only when the selected CLI is not installed; preserve a working installed executable and report whether its version is current, outdated, or ahead of the registry.
- Show the same CLI status in both planning output and final application results.
- Preserve a healthy installed CLI with a warning when the registry is unavailable, while blocking malformed or failing version output and manager-registered packages whose executable is unavailable.
- Provide recovery instructions for inconsistent global installations without automatic cleanup or cross-manager ownership inference.
- Retain the existing Node.js preflight, configuration-write boundaries, npm-first/pnpm-fallback selection, and operation ordering except where complete CLI prevalidation must precede all package installation.
- Add focused injected unit tests only; do not invoke real npm/pnpm installation or require the full repository test suite.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `global-cli-management`: Replace unconditional `@latest` updates with prevalidated install-if-missing behavior, installed-versus-registry version reporting, offline preservation, and blocking diagnostics for inconsistent selected-manager state.

## Impact

- Affects shared CodeGraph/OpenSpec global CLI discovery, preflight, planning, application, diagnostics, and result formatting in the installer.
- Extends injected command seams and focused unit coverage for package registration, executable versions, registry versions, and failure states.
- Supersedes the archived requirement that every selected global CLI is always updated to `@latest`; it does not alter OpenSpec skill bundling, CodeGraph-specific configuration, or global package cleanup behavior.
