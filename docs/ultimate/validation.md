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
| Original Electron desktop package | Blocked rebuilding native `better-sqlite3`: Visual Studio C++ installation unavailable |
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

### Original desktop build reproduction

A separate, untouched worktree at the fixed baseline ran real frontend and
product-ui `npm ci` installs. The product-ui build passed. `npm run build` in
that worktree reached Electron Forge's native-dependency rebuild after daemon,
browser extension, Windows runtime preparation and ACP runtime steps. Python
was first unavailable to node-gyp; using the desktop's bundled Python resolved
that prerequisite. The next failure was `Could not find any Visual Studio
installation to use` while rebuilding `better-sqlite3`. No installer was built,
and no real Electron UI launch is claimed.

Large upstream downloads stalled. Official Agent Browser 0.38.1 and Node
22.23.2 Windows x64 assets were downloaded separately in ranges and verified:

| Asset | SHA-256 |
|---|---|
| Agent Browser | `70bd758b2a5a72b18055f60ff877d2297d036c4104e3bc94af6f7c1ceb941408` |
| Node runtime ZIP | `1177b4137ba5adaa56354ae40f1080c7450e8ae09cecb47da459d1c52ac99f97` |

Only that build's Node process used an external fetch hook mapping the exact
official Node archive URL to the verified local asset. The original runtime
script still performed its checksum/extraction steps; no source or build marker
was fabricated. This is a cached-asset reproduction, not a clean online build.
ACP runtime `npm ci` reported nine audit findings (four moderate, four high,
one critical); exploitability was not investigated and no automatic dependency
rewrite was applied. Retain this as an unresolved release dependency audit.

Evidence: `baseline-desktop-build-python.log`, earlier build/download logs and
the verified assets under the external tool cache.

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

## Approval, cancellation and recovery

DSH turn `72ab6eeb-1eb7-4883-a58d-e84215ef00e8` requested escalation
for a harmless print command under the default permission policy. Its tool kind
was absent from the permission request. AO displayed an approval instead of
silently approving it. The pending request ID survived a graceful daemon restart.
Answering `reject-once` produced the native rejection error and
`APPROVAL-REJECTED`; the print command did not run.

DSH turn `b3fa2346-9c5b-4dca-bb16-cb57ff50d18a` was interrupted while its
shell tool was running. AO recorded `interrupted`, the provider recorded
`tool call aborted`, and the success marker was absent. An earlier tool attempt
in that turn failed because the requested sandbox option lacked a justification;
the model retried it. This is protocol cancellation evidence, not proof of every
descendant process being killed. No process-tree cancellation claim is made.

The original exit-agent/resume-agent endpoints then stopped and restarted this
DSH controller. Its four durable turns, native session and interrupted outcome
remained; no extra turn was dispatched. This checks deliberate provider exit and
native resume, not forced process termination at every lifecycle boundary.

Evidence: `dsh-approval-probe.json`, `dsh-approval-after-restart.json`,
`dsh-approval-final.json`, `dsh-cancel-before.json`, `dsh-cancel-after.json`,
`dsh-controlled-exit.json`, `dsh-native-resume.json`,
`dsh-after-native-resume.json` in the external validation directory.

## Persistent read-only foundation — release gate remains blocked

Added the public `readOnly` session property and `ao spawn --read-only`.
A new SQLite migration persists the immutable constraint; ordinary metadata
updates cannot remove it. Only Codex Chat can own these sessions. Recovery,
history branches, side conversations, provider changes, interface changes and
permission changes preserve the constraint or refuse incompatible operations.

Native filesystem sandboxing alone does not cover MCP tools. A live test of the
original reviewer policy exposed an inherited Node MCP tool outside the native
sandbox. Public read-only sessions therefore disable inherited and supplied MCP
servers, plugins, apps and native subagents. Before each turn, the adapter checks
the actual thread MCP inventory and refuses exposed external tools or an
unavailable/incomplete inventory. Reload is refused for these sessions. Existing
PR-review policies are not globally replaced by the new public-session policy.

Actual Windows results with Codex 0.162.1:

| Native sandbox | Observed result | Product qualification |
|---|---|---|
| User's configured `elevated` | Both shell read and write fail during sandbox setup; sharing violation on a loaded Codex runtime executable | Blocked; missing file is not proof of a working sandbox |
| Explicit `unelevated` lab override | Reads succeed; direct writes denied, including after daemon restart; predefined defect detected | Insufficient: loopback HTTP still succeeds despite `networkAccess=false` |
| Explicit `mxc` probe | Native executor reports MXC unavailable | Blocked on this machine |

The loopback result matters: a shell could invoke local daemon actions indirectly.
Do not qualify the legacy fallback as an isolated read-only reviewer. The final
implementation **refuses** public read-only Windows launches/reconnects that
request network isolation without an explicit `elevated` or `mxc` configuration.
The unsafe fallback was used only to diagnose the boundary; it is not a product
default. Neither the user's global Codex config nor credential values were changed.

The optional lab sandbox override is applied to native process arguments,
not merely to a future thread request. On reconnect the actual surviving native
process configuration is checked; a replacement daemon's new preference cannot
qualify an old unsafe host. Regression coverage includes that recovery boundary.
Final live recovery with an explicit `unelevated` diagnostic host returned
HTTP 409 / `SESSION_MODE_UNSUPPORTED` before a new turn was dispatched. The two
stored turns remained unchanged (`readonly-unsafe-resume-refusal.json`).

The lab session `ultimate-validation-6` detected the seeded `a - b` defect with
no external MCP tools exposed. Its first direct write returned exit 1 and read
returned 0 (turn `24d3e77f-3196-438a-bb6d-8a3e2b406f96`). After daemon
restart the constraint still denied writes (turn
`631f0ce1-bdb8-4e0f-899d-95ff5c714062`), but the loopback read returned HTTP
200. These are diagnostic results, not end-to-end review acceptance.

Relevant evidence: `readonly-public-probe-result.json`,
`readonly-recovery-final.json`, `codex-isolated-review-unelevated.log`,
`codex-isolated-review-mxc.log`. A separate process-start override probe did not
complete its handshake within its deadline; it provides no passing evidence.

Focused new tests pass for persistence, supported harness/mode, permission
escalation refusal, CLI serialization, external-tool exclusion on
start/resume/fork, managed-tool inventory and unsafe Windows sandbox refusal.
The broader affected suites still have preexisting Windows failures. Two
history-import test names newly observed during repeated runs were reproduced
on the untouched baseline worktree: duplicate IDs from timestamp-based fixtures.
Do not count those suites as passed. See `readonly-final-tests.log`,
`baseline-history-id-repro.log` and `baseline-history-recovery-repro.log`.

sqlc v1.31.1 and the normal OpenAPI/TypeScript generators regenerated all changed
artifacts. Storage/API-spec suites, backend build, relevant Go vet and frontend
typecheck passed. Race testing remains unavailable without a native C toolchain.

Latest focused checks: `go test -count=1 -timeout 3m` on the Codex driver, Chat
service, session manager and CLI with `-run TestReadOnly` passed. Final native
daemon build, affected-package vet and frontend typecheck also passed. Broader
reconnect tests using `/tmp/ws` fail Windows absolute-path validation, consistent
with the upstream fixtures; do not present the broader suites as green.

### Sandbox check without model use

```powershell
$env:AIDSU_VALIDATION_DIR = '<disposable validation directory>'
$env:CODEX_BINARY = '<installed codex.cmd path>'
node scripts/ultimate/probe-codex-sandbox.mjs
# Optional diagnostic override for only this probe process:
$env:AIDSU_CODEX_WINDOWS_SANDBOX = 'unelevated'
node scripts/ultimate/probe-codex-sandbox.mjs
Remove-Item Env:AIDSU_CODEX_WINDOWS_SANDBOX
```

The script uses native `command/exec`, creates no model thread/turn, and never
refreshes credentials. It checks a seeded read, a forbidden write and access to
its own temporary loopback server. Setup/RPC errors cannot count as isolation
success. The configured elevated mode fails setup; unelevated allows loopback;
both reports have `ok=false` and zero model prompts. An invalid executable also
fails cleanly. Evidence: `native-sandbox-no-model.log`,
`native-sandbox-unelevated-no-model.log`, `native-sandbox-invalid-binary.log`.

## Stage status and retained state

| Stage | Current state |
|---|---|
| A | Full local history, fixed branch, isolated build/test environment and evidence committed; hosted fork and desktop package blocked |
| B | Real Codex relay turn, two parallel DSH ACP workers, approval rejection, interruption, deliberate controller exit/resume and daemon restart evidence obtained; PR reviewer and strict Windows local review not qualified |
| C | Durable public read-only foundation implemented and tested; unified intake, control ownership/revision fencing, local base/head review binding, acceptance evidence and two-round rework policy remain unimplemented |
| D | Branding, installer identity, product data root, update/telemetry replacement and coexistence not implemented; held behind critical validation gates |

No GitHub PR/comment was published. No hosted `origin` was invented without
authentication. The user relay provider remains in place. Scratch controller
processes were stopped with their own exit-agent APIs; the isolated daemon was
shut down with its own endpoint. Scratch sessions, logs and Git worktrees remain,
including Worker 1's uncommitted calculator fix. No user Codex desktop process
was killed and no dirty worktree was deleted. See `scratch-host-exit.json`.

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
