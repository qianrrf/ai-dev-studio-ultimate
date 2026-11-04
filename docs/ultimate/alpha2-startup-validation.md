# Alpha 2 startup repair

The alpha 1 desktop checks inherited Codex's private Git directory. A normal
Explorer launch had neither Git in its persistent PATH nor a standard Git
installation. The previous installer smoke result did not establish independent
desktop startup.

## Environment repair

- Installed the official Git for Windows 2.56.0.windows.4 for the current user
  under `%LOCALAPPDATA%\Programs\Git`.
- Source: `git-for-windows/git`, release `v2.56.0.windows.4`, asset
  `Git-2.56.0.4-64-bit.exe`.
- Verified SHA-256:
  `4961a4a8cc778d8b6e00a5819e5fe4d078b754555aff7aedc2d77a5d74c2de80`.
- Authenticode: Valid; publisher Johannes Schindelin.
- Silent current-user installer exit: 0; `git --version` succeeded.
- Local install log: the validation cache's `alpha2/git-install.log` (not committed).

## Product changes

The desktop refreshes persisted machine/user PATH without loading a PowerShell
profile, adds existing standard and registered Git installations, and sends a
single Windows PATH spelling to its daemon. The daemon also prepares its process
environment before starting services. An existing valid PATH entry retains
priority. Missing Git still blocks readiness and offers the official installer
website; no install is silently attempted by the product.

## Checks so far

- Frontend shell-environment and dependency-dialog suites: 48 tests passed.
- The native Windows probe executed Git with inherited PATH reduced to System32;
  the recovered PATH contained no Codex runtime directory.
- Frontend full typecheck passed.
- Complete systemexec package tests passed with official Git Bash available.
- Regression fixture covers standard/current-user/registered roots, missing
  installations, Chinese/spaced paths, and existing PATH priority.

These checks are not final installer or business acceptance. Alpha 2 still
requires the complete matrix and a freshly built installer launched independently
of the coding environment.
