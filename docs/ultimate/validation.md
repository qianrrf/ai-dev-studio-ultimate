# Validation ledger

Validation date: 2026-11-04, native Windows. Unlisted checks are not passed.
All application validation uses scratch state outside the user's original AO
data and projects.

## Tool versions

| Tool | Observed version |
|---|---|
| Git | 2.53.0.windows.3 |
| Node | 24.20.0 |
| npm | 11.19.0 |
| Go | 1.27.1 (isolated tool cache; no global PATH changes) |
| Codex CLI | 0.162.1 |
| DSH CLI | 0.2.1-rc.5 |

Go Windows amd64 archive was checked against official SHA-256
`a3911b5e0e1b1053f25ed0675f4c1c6aad1e2bfcf253df2b9be4caabd2edd95d`.
Source snapshot was checked against every baseline Git blob before checkout
was accepted (5,680 files); Git history is not shallow.

## Baseline checks

| Check | Result |
|---|---|
| Root, frontend, product-ui locked dependency installs | Passed |
| product-ui build | Passed |
| Frontend typecheck | Passed |
| Frontend tests (unmodified baseline) | 603 files passed, 34 failed; 10,793 tests passed, 194 failed, 17 skipped; 272 seconds |
| Go build and tests | In progress |
| Hosted GitHub fork | Blocked: no authenticated GitHub client |

## Native DSH protocol probe (no model prompts)

Installed executable: `D:\AI\deepseekh\resources\runtime\cli\bin\dsh.cmd`.
Started through Windows command processor, `--profile acp`, disposable workspace.
ACP initialize v1, session/new and session/close passed. Agent identifies as
`deepseek-harness-acp` protocol version `0.0.1`; two config options returned.
No non-protocol stdout lines. Zero model prompts; no claim of task completion,
concurrency, approvals or daemon-driver success from this check alone.

Machine-local evidence lives under
`%USERPROFILE%\.cache\ai-dev-studio-ultimate\validation`; raw logs and local
session databases are deliberately excluded from Git.

The baseline frontend suite includes macOS-only updater fixtures, Unix symlink
and login-PATH expectations, date/locale assertions and some renderer failures.
These are preexisting Windows baseline failures, not a green suite. Individual
failures still need classification before release; no unrelated fixes are
included merely to make a broad test count green.

## Reproduction

On Windows with Git, Node/npm and the Go version in `backend/go.mod`:

```powershell
. ./scripts/ultimate/enter-dev.ps1 -GoBin '<Go bin directory>' -DshBin '<DSH bin directory>'
npm.cmd ci
npm.cmd --prefix frontend ci
npm.cmd --prefix packages/product-ui ci
npm.cmd --prefix packages/product-ui run build
npm.cmd run frontend:typecheck
Push-Location backend
go build ./...
go test -timeout 8m ./...
Pop-Location
npm.cmd --prefix frontend test -- --reporter=dot
$env:DSH_BINARY = '<full path to dsh.cmd>'
node scripts/ultimate/probe-dsh-acp.mjs
```

The ACP probe sends no model prompts and rejects provider requests for client
tools. It is a validation client, not a product adapter. Dependency download
mirrors may be configured per process; Go's module checksum validation remains
enabled. No global PATH or agent configuration changes are required.
