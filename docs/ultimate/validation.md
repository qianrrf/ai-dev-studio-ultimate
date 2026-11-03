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
| Go build | Passed (`go build ./...` and native daemon executable) |
| Go vet | Passed (`go vet ./...`) |
| Go baseline tests | 156 packages passed, 56 failed; 288 failing test names recorded |
| Native Go DSH ACP model-catalog live test | Passed with installed `.cmd` |
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

Backend baseline failures include unsupported Unix assumptions and Windows
symlink/ancestor-ACL requirements. A native C compiler was not found and
`CGO_ENABLED=0`, so race checks are not marked passed.

## Daemon validation and first compatibility fix

Isolated daemon: loopback port 31337, state and discovery under the lab root.
Disposable source repo has a Chinese/spaced path and explicit `main` base.
Two DSH Chat workers (`ultimate-validation-1`, `ultimate-validation-2`) started
concurrently on separate branches `validation-1` and `validation-2`.

The first two prompt attempts both failed durably with `provider_key_missing`:
the default `deepseek-official` route has no API key. Inspection of the native
catalog and credential *names* showed an existing DeepSeek account route.
Selecting the advertised `["deepseek-account","deepseek-flash"]` value through
the original config-option API fixed configuration without an adapter change.
Both workers then returned their expected distinct markers. Recorded turns:

| Worker | Turn | Start UTC | Complete UTC | Result |
|---|---|---|---|---|
| 1 | `4d484edf-bca4-45e5-bff7-d7cdaa1633e6` | 17:14:43 | 17:14:48 | ACP-ACCOUNT-1-OK |
| 2 | `8d0ec455-9458-4335-be89-d54a14dda94e` | 17:14:43 | 17:14:46 | ACP-ACCOUNT-2-OK |

The intervals overlap. These are real native ACP turns, not headless tests.

Codex uses the user's custom Responses provider with `requires_openai_auth=false`
and an inline bearer credential. No credential value was printed or copied.
Native `account/read` returns `requiresOpenaiAuth=false` with no account.
The original account catalog nevertheless classified the absence of auth.json
as signed out and blocked spawn (`CODEX_ACCOUNT_AUTH_UNVERIFIED`).

Minimal fix: for a reconciled device with no saved credential, ask the native
account client about the provider's actual authentication requirement, without
refreshing credentials. Keep signed-out and inconclusive results distinct;
reuse the existing process cap and shared operation gate. New regression tests
failed before the fix and pass after it (custom provider / signed out /
unavailable native read). The full affected package still has 87 baseline
failures, with no newly failing test names compared with the baseline.

After this fix, native Codex Chat orchestrator `ultimate-validation-3` completed
turn `63f2aa20-01ae-49b7-977f-c351f3a887d5` through the existing custom provider,
returning `CODEX-CONTROLLER-OK` (17:13:09–17:13:18 UTC). This proves a real turn;
it does not yet prove controller fencing or independent local review.

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
