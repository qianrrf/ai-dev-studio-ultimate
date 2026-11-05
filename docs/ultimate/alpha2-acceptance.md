# Alpha 2 acceptance record

Updated 2026-11-05 (Asia/Shanghai). **Not accepted for delivery.** An installer
build, a successful installation, and complete business acceptance are separate
gates. This record supersedes the alpha 1 status in the older reports.

## Candidate and fixed inputs

- Product source: `01203b16b599ec34541470efd410a1f61a2818e0`.
- Upstream baseline: `b325d24b741297bf89296e76501d7e09018c4a04`.
- Version: `0.1.0-alpha.2`, Windows x64, unsigned.
- Local installer: `dist/0.1.0-alpha.2/AI Dev Studio Ultimate Setup 0.1.0-alpha.2.exe`.
- SHA-256: `7531dcf503525dba40348be25beb6c74b9cf33e1524446a22d320604f486f646`.
- Size: 191345935 bytes. `SOURCE.json` and `SHA256SUMS.txt` accompany the file.
- Build inputs: Node 24.20.0, npm 11.19.0, Go 1.27.1, Electron 44.7.0,
  Python 3.12.14, better-sqlite3 12.11.1, `CGO_ENABLED=0`. Native dependencies were rebuilt for
  Electron; no ABI substitution or fabricated build marker was used.
- `go version -m` on the bundled daemon confirms this revision and
  `vcs.modified=false`.

Artifacts and local execution data are ignored by Git. Older candidates remain
under the distribution cache; their results do not prove this candidate passed.

## Startup, installation and data checks

The alpha 1 startup gap was a Git installation/PATH dependency hidden by the
coding environment. The current-user official Git installation has a verified
signature and release checksum. The application refreshes persisted Windows
PATH and checks standard/registered Git locations without loading shell profiles.
An existing working Git retains priority. Missing Git still blocks readiness
with an official install link and restart guidance.

With only persisted machine/user PATH, the following actual programs execute
their version commands: Git 2.56.0.windows.4, Codex 0.162.1, DSH 0.2.1-rc.5 and
GitHub CLI 2.103.3. No model request is involved. The isolated desktop check also
starts with PATH reduced to Windows directories; fixture Git commands use the
absolute executable discovered by the real daemon, rather than assuming the
test runner inherited it.

Previous installation checks, last repeated at `2f84a5e29`, in the dedicated
Chinese/spaced installation directory:

| Check | Observed result and scope |
|---|---|
| Fresh alpha 2 build and silent installation | Exit 0; expected installed executable and version exist |
| Real Electron and bundled daemon | Native renderer, ready daemon, product title, zero renderer page errors in the completed core smoke |
| Disabled hosted services | Update enable attempt stays disabled; cloud account remains null |
| Bundled CLI and synthetic Chinese project | Project registration, stale-revision conflict, intake deduplication and release work |
| Daemon restart | New PID; control revision and intake survive |
| Start menu at the previous candidate | Real installed shortcut opens the product with a reduced PATH; old control/intake and dirty file remain intact |
| Alpha 1 to alpha 2 upgrade | Old project, revision 2, one intake and zero model sessions remain readable |
| Stopped-data backup/restore | Durable data restored to a second scratch root; same project/control/intake readable |
| Dirty files | A real Git-registered worktree's uncommitted README hash remains unchanged through upgrade and uninstall |
| Uninstall/reinstall | Exit 0; executable, product shortcut and product protocol removed; synthetic database and dirty file retained; reinstall succeeds |
| Long workspace root | **Fails.** Physical root exceeds MAX_PATH; Git/CreateProcess rejects it |

The previous `2f84a5e29` package was built, installed over the dedicated test
installation, and rechecked with PATH limited to Windows directories. Git was
resolved by the bundled daemon, controller/intake operations and restart passed,
and renderer errors were zero. Its full desktop smoke remains `passed=false`
because the long-root operation failed. The real Start-menu shortcut also
opened this revision against the old synthetic data: revision 2, one intake,
zero model sessions and the dirty file checksum were preserved. The current
package's uninstall returned 0 and removed its executable, protocol and Start
menu shortcut while retaining the database and dirty worktree checksums.
Reinstallation reopened the old records. A new stopped durable backup had an
identical database/file hash and reopened successfully from a second data root.
These checks still created zero model sessions.

The previous `35cb6feae` package was also built from the clean, independent
product worktree and installed in that Chinese/spaced directory. A fresh smoke
with PATH reduced to Windows directories passed actual Electron/daemon startup,
Git discovery, update/cloud disablement, bundled CLI project registration,
stale-control rejection, intake deduplication, daemon replacement and persisted
control/intake checks. Renderer errors and model-backed sessions were zero.
Its overall result is still `passed=false`: the 314-character repository-root
probe cannot launch Git. The earlier Start-menu, uninstall and backup results
above were not silently reassigned to this new source revision.

At `5579b9190`, the independently built installer was exercised against the
previous candidate's synthetic state. Actual Start-menu launches used a PATH
containing Windows directories only. The old project, control revision 2, one
intake and zero model sessions remained readable before upgrade, after upgrade,
from a stopped durable backup and after uninstall/reinstallation. The backup
and restored database hashes matched before reopening. Complete uninstall
removed the executable, shortcut and product protocol while preserving the
original database hash and the exact uncommitted bytes of a Git-registered
worktree. That worktree stayed registered. The result is recorded in local
`alpha2/lifecycle-5579b9190/result.json`; it passed with zero model calls.
The current installed package also passed occupied-port and exclusive database
file-lock probes. Both surfaced a stopped daemon with a failure code, exit 1
and the concrete bind/SQLite error; neither became ready or attached to the
synthetic foreign service. The initial probe incorrectly expected only the
`error` status; its corrected assertion follows the existing documented
`stopped` plus failure-code representation and refuses a clean user stop.

The preceding `13c5cfbc5` candidate was rebuilt in that clean worktree after a
test-only fixture correction. Its new lifecycle run independently passed
Start-menu launch with the reduced PATH, upgrade from the preceding candidate,
stopped backup/restore with matching hashes, uninstall/reinstall, persisted
project/control/intake records and preservation of another registered dirty
worktree. Occupied-port and exclusively locked-database checks also passed.
Its actual desktop smoke again has zero renderer errors and passes startup,
Git discovery, CLI/control/intake and daemon-restart checks; the overall result
remains false because the 314-character repository root fails. Evidence is in
`alpha2/lifecycle-13c5cfbc5/result.json`, `startup-errors-13c5/result.json` and
`13c5-clean-path-smoke/result.json`. All these runs used zero model calls.

The preceding `c296d69e7` candidate was then built cleanly after correcting the
queued-batch test's synchronization. Its independent installation lifecycle
passed again, including preservation of the old synthetic records, matching
stopped-backup hashes and a newly registered dirty worktree through complete
uninstall/reinstallation. Evidence: `alpha2/lifecycle-c296d69e7/result.json`.
Its occupied-port and locked-database probes passed with zero model calls
(`startup-errors-c296/result.json`). The fresh reduced-PATH desktop smoke
again has zero renderer errors and passes Git discovery, project/control/intake
and daemon restart, but remains `passed=false` at the 314-character repository
root (`c296-clean-path-smoke/result.json`).

The latest `01203b16b` candidate is a clean build after correcting checkout
line endings. Its bundled daemon has `vcs.modified=false` and the exact source
revision above. A separate fresh worktree checked out with `core.autocrlf=true`
contains 2387 Go files and none with CRLF. Complete Windows static analysis
there reports zero issues (`alpha2/lint-fresh-autocrlf-01203.log`). The same
source copied as CRLF reproduced the whole-file goimports diff; converting
that probe to LF produced no diff. The new installation lifecycle passed with
zero model calls: actual Start-menu launch, upgrade, stopped backup/restore,
uninstall/reinstall and exact preservation of a registered dirty worktree
(`alpha2/lifecycle-01203b16b/result.json`). Port-occupation and exclusive SQLite
lock probes passed (`startup-errors-01203/result.json`). Its fresh reduced-PATH
desktop smoke passes Git discovery, project/control/intake and daemon restart
with zero renderer errors. Its overall result remains false at a 315-character
physical repository root (`01203-clean-path-smoke/result.json`). The installed
daemon hash matches the clean build (`installed-source-01203.json`).
Complete remote validation run
[41948560176](https://github.com/qianrrf/ai-dev-studio-ultimate/actions/runs/41948560176)
has confirmed Linux backend, both frontends and installer/coexistence success.
The Windows backend race step failed; final failure details and its final
static check remain unverified because GitHub connectivity failed before the
watcher could collect completion. This is not an in-progress or passing claim.
The preceding candidate is retained under
`dist/0.1.0-alpha.2/previous-c296d69e7`; its evidence is not reassigned.

The dirty fixture is registered with Git, not a model-backed product session.
No model-backed session was created during installation checks. The stopped
backup excludes transient `run` sockets and `daemon.lock`; two earlier copy
attempts involving an AF_UNIX socket were incomplete and are not restore
evidence. The successful durable backup has matching database SHA-256.

The upgrade used the alpha 1 candidate's database, not an arbitrary old
production database. This is not a claim of compatibility with every historical
upstream schema. Earlier tests of port conflicts, database locks and portable
upstream coexistence remain historical evidence in [installer validation](installer-validation.md).

## Test status

Current installer source `01203b16b`:

| Required check | Current-source result |
|---|---|
| Linux frontend | CI job succeeded: 11005 passed, zero failed, 13 separately classified pending cases |
| Linux renderer | 148 passed, zero unexpected, zero flaky, zero skipped; native browser and real Electron boundary steps passed |
| Windows frontend | CI job succeeded: 10950 passed, zero failed, 15 separately classified pending cases |
| Windows renderer | 148 passed, zero unexpected, zero flaky, zero skipped; native browser and real Electron boundary steps passed |
| Linux complete backend | Complete CI job succeeded in 39m1s; final JSONL counts could not be downloaded after connectivity failed |
| Windows complete backend / static | Dedicated long-root and complete race steps failed; last watcher snapshot showed SQL generation running. Final failure details and static status could not be retrieved |
| Installed lifecycle and startup errors | Passed with zero model calls; fresh full desktop check still fails the 315-character root |
| Actual two-installer coexistence | Passed on Windows CI; upstream's own stale protocol registration remains documented |
| Final-source live business / recovery | Not qualified; 40/40 authorized model rounds already registered |

Earlier evidence, with its own source boundaries:

| Scope | Actual evidence |
|---|---|
| Local renderer interaction suite | 148 passed, no skips |
| Previous source `c296d69e7`, Linux frontend | 11005 passed, 0 failed, 13 pending; complete CI job succeeded |
| Previous source `c296d69e7`, Linux renderer | 148 expected, 0 unexpected, 0 flaky, 0 skipped; native browser, actual Electron boundaries and guarded navigation passed |
| Previous source `c296d69e7`, Windows frontend | 10950 passed, 0 failed, 15 pending; complete CI job succeeded |
| Previous source `c296d69e7`, Windows renderer | 148 expected, 0 unexpected, 0 flaky, 0 skipped; native browser, actual Electron boundaries and guarded navigation passed |
| Previous source `c296d69e7`, Windows complete backend race | 16400 passing, 1 failing and 319 skipped test/subtest records; 211 passing, 1 failing and 18 no-test packages; zero data-race warnings. The sole failing test is the retained long physical root case |
| Previous source `c296d69e7`, Linux complete backend | CI job succeeded: 16988 passing and 109 skipped test/subtest records, zero failures; 215 passing and 16 no-test packages, zero data-race warnings. Build, vet, API/sqlc drift and static checks passed |
| Previous source `c296d69e7`, Windows remote static check | Failed on source files checked out as CRLF; a CRLF copy reproduces a whole-file goimports diff, while the same source using LF produces no diff. The following checkout-policy fix is recorded separately |
| Local complete Go static analysis | 0 issues for both Windows and Linux target selection with golangci-lint 2.13.2; Linux target analysis is not native Linux execution |
| Build/vet, types and generation | Windows build/vet, Linux cross-build/vet, frontend and browser-test types, API/sqlc/type generation drift checks passed |
| Real Electron daemon boundaries | All three actual Chromium integration cases passed after removing the Node-mode environment variable rather than setting it empty |
| Guarded browser navigation integration | Passed with the checksum-verified pinned native browser; includes canceled clicks, popup navigation and daemon restart |
| Windows HTTP/process/persistent-host complete packages | Passed with a private test TEMP ACL |
| Sharing-lock and IPv6 shutdown regressions | 20 consecutive runs passed |
| Linux frontend at `8b90666f7` | 11004 passed, 0 failed, 13 pending; CI job succeeded |
| Linux renderer at `8b90666f7` | 148 expected, 0 unexpected, 0 flaky, 0 skipped |
| Linux frontend at `5579b9190` | 11005 passed, 0 failed, 13 pending; complete CI job succeeded |
| Linux renderer at `5579b9190` | 148 expected, 0 unexpected, 0 flaky, 0 skipped; native browser, actual Electron boundaries and guarded navigation also passed |
| Windows frontend at `5579b9190` | 10950 passed, 0 failed, 15 pending; complete CI job succeeded |
| Windows renderer at `5579b9190` | 148 expected, 0 unexpected, 0 flaky, 0 skipped; native browser, actual Electron boundaries and guarded navigation also passed |
| Linux full backend race suite at `5579b9190` | 16987 passing test/subtest records, 1 failing record, 109 skipped records; 214 packages passed, 1 failed, 16 had no tests; the only failure was a long-file fixture measuring 252 UTF-16 units; zero data-race warnings |
| Windows full backend race suite at `5579b9190` | 16398 passing test/subtest records, 3 failing records (including a parent), 319 skipped records; 210 packages passed, 2 failed, 18 had no tests; failures were the long-root case and an early queued-batch assertion; zero data-race warnings |
| Deep-file fixture correction at `13c5cfbc5` | Keep the assertion above MAX_PATH on shorter Linux temporary roots by increasing the fixture's depth; Windows actual Git/worktree/hook/write/commit/read/cleanup passed 3 times at 363–364 UTF-16 units |
| Queued-batch test synchronization at `c296d69e7` | Wait for the stored follower receipt as well as the first running turn; original state/batch-content/send-count assertions remain; both affected top-level cases passed 20 repetitions |
| Local complete chat package after that correction | Executed; one artifact-symlink fixture fails because this account lacks symlink privilege; not green and not skipped |
| Windows frontend CI at `c73d2408a` | 10950 passed, 0 failed, 15 pending; renderer 148 passed with no skips/flakes; native browser step passed |
| Local complete frontend at `35cb6feae` | 10932 passed, 18 failed, 15 skipped; all 18 failures are fixture symlink EPERM on this non-elevated host |
| Local complete backend after test-environment/ID corrections | 16315 passing test/subtest records, 26 failing records and 362 skipped records; remaining failures are unavailable symlink privileges and the long-root case, including parent failure records; not green |
| Chat history ID collision | A timestamp-based injected test ID collided in the complete suite; UUID fixtures replace that assumption; the reproducing case passed 20 consecutive runs |
| Revealed-line timer regression | Windows CI exposed one asynchronous clock failure; explicit frame/time advancement retains the before-expiry and expiry assertions; the complete 12-test file passed 20 consecutive runs |
| ACP concurrent completion | Linux race found a real shared-map read/write; settlement now snapshots message maps and tool values under the existing lock; the regression fails against old source and passes 20 times after the fix; the complete ACP package passes locally |
| Source `c296d69e7`, native Linux ACP race package | Passed in 52.699 seconds; its complete backend result is recorded separately above |
| Windows shell fixtures | Git's own POSIX tools are added only to the test process; fake agent, process environment, tmux, system execution and session-manager complete packages passed |
| Enabled Windows symlink cases in CI | Passed at `c73d2408a`, including runtime-key and diagnostics redirect checks |
| PDF teardown regression | Correct PDF.js TextLayer render/cancel mock; both complete affected test files passed (54 tests) and TypeScript check passed |
| Windows descriptor/project packages after fixture corrections | Both complete packages passed; sharing-violation regression passed 20 consecutive runs |
| Local complete workspace package after fixture corrections | Still fails the long-root case and symlink fixture privilege; no green claim |
| Native browser focused suite locally | Pinned Windows binary enabled the three native compatibility cases; package passed, with its separate Unix socket case inapplicable on Windows |
| Previous `c296d69e7` complete Windows/Linux CI | Run completed with failure: all jobs succeeded except Windows backend; no all-green conclusion |
| Current `01203b16b` complete Windows/Linux CI | Linux backend, both frontends and installer jobs succeeded. Windows backend has failed test steps; final metadata unavailable after connectivity failed. No all-green claim |
| Windows long workspace operations in CI | Fails even after runner enables Developer Mode and LongPathsEnabled |
| Deep Chinese file paths with an ordinary workspace root | Actual create, hook, write, commit, read and cleanup passed for a 289-UTF-16-unit path; the separate failing long-root case is retained |

The CI workflow now runs complete build/vet/race/lint, API/sqlc drift, frontend
types/full tests, renderer interactions and native browser compatibility.
Windows race uses MSYS2 UCRT GCC. Each generator/type/lint check is independent
of previous test failures. The 40-minute race package timeout permits the full
SQLite migration suites; it is not a passing result by itself.

The `c73d2408a` Windows race run also exposed a lock-fixture precondition failure
and two incidental Git path failures in ordinary fixtures. `2f84a5e29` makes
the lock use the production extended-path reader, keeps an isolated private
temporary parent short, and sets `core.longpaths` in the 101-character-name
repository. No assertion, long-root fixture or required case was removed.
That run was superseded by
[41928895662](https://github.com/qianrrf/ai-dev-studio-ultimate/actions/runs/41928895662)
at `35cb6feae`. Its Linux frontend job succeeded. Windows completed 10949 tests
with one timer failure and 15 pending tests; the 148 renderer cases, native
Electron boundaries and guarded navigation passed. Its Linux race log exposed
the ACP snapshot defect described above. Obsolete runs were cancelled when
their replacements were pushed; their interrupted race suites are not passes.
The `5579b9190` source run is
[41934511657](https://github.com/qianrrf/ai-dev-studio-ultimate/actions/runs/41934511657)
at `5579b9190`. Its Linux backend job failed only the fixture-length assertion;
its ACP package passed with race detection and its final API/sqlc/static checks
passed. Its Windows backend completed with the long-root failure and an early
queued-batch assertion: the first turn becomes running before the follower's
completion receipt is persisted. The correction now waits for that receipt in
storage; the presentation snapshot deliberately omits batch followers.
An initial attempted snapshot-only wait could never observe that hidden row;
it failed locally and was corrected before committing. It is not passing evidence.
The `13c5cfbc5` run
[41939153415](https://github.com/qianrrf/ai-dev-studio-ultimate/actions/runs/41939153415)
was cancelled while pending so both test corrections could be validated together.
The preceding source `c296d69e7` ran the complete replacement matrix in
[41941537845](https://github.com/qianrrf/ai-dev-studio-ultimate/actions/runs/41941537845).
The run completed with failure. Linux backend, both frontends and the
installer/coexistence jobs succeeded. Windows completed its race suite with
only the long-root test failing; the queued-batch tests now pass in the native
complete race run. Its build, vet and API/sqlc drift checks passed, but its
static check failed because Windows checkout converted Go sources to CRLF.
Commit `01203b16b` adds `*.go text eol=lf` to Git attributes; it neither rewrites
source files nor disables goimports. The installer above remains built from
`01203b16b`; the previous `c296d69e7` installer is retained separately.
The preceding Linux backend race suite passed, but its final lint step found
an import-order defect in a Unix-only test; that defect is fixed. Local
focused/full-package passes do not establish the final race outcome.

Security symlink tests are no longer skipped on Windows. Existing key redirects
are refused before reads/appends, and diagnostic log redirects are refused
where Node lacks O_NOFOLLOW. These checks assume trusted local state; they do
not claim protection against a malicious process racing filesystem operations.

Remaining old pending tests must be classified explicitly. They include absent
mobile connection-picker UI, native-browser tests when their binary is not
provided, Unix-only socket/ps/shell cases, and a macOS quarantine test. Native
browser cases have a separate enabled CI step; Linux covers its native OS
cases. Disabled/mobile/macOS cases are not silently counted as passing Windows
tests or as completed required product acceptance.

The 15 pending cases in the Windows frontend report break down as follows:
three pinned-native-browser cases are actually executed by the separate native
compatibility step; one Unix socket-alias case, two POSIX process-start cases
and two POSIX shell-quoting cases are executed on Linux; one macOS quarantine
case is outside the Windows/Linux product matrix; six old mobile connection
picker cases describe an unimplemented mobile UI outside this single-machine
alpha. Those six remain pending rather than being presented as delivered
features. No required product test was removed or newly skipped.

The `5579b9190` Linux frontend report has 13 pending cases: the three native-browser
cases run in its separate enabled step, three Windows environment cases belong
to the Windows job, one quarantine case requires macOS, and six mobile-picker
cases are outside this alpha. Pending cases are not included in its 11005 passes.

The `5579b9190` Linux backend's 109 skipped test/subtest records include
platform-specific cases, opt-in provider probes, and 50 real-model end-to-end
cases gated by `AO_CHAT_E2E`. The complete automated suite runs without model
credentials; its result does not certify those live cases. The separately
budgeted product loops are observed evidence for their exact scenarios, not a
claim that every upstream provider's opt-in integration suite ran. The three
Electron boundary cases have an enabled, separate CI step.

## Real business evidence and the model limit

The ledger contains **40 registered rounds, the full authorized allowance**.
No further model requests are permitted without a new allowance. Failed and
blocked attempts are counted. External Codex resume used the account's default
model on one invocation; it must not be described as exclusively Sol execution.

The product controller independently dispatched two concurrent DSH workers,
registered actual test results, created separate read-only Codex reviews and
recorded acceptance through the delivery API. Tests passed 5/5 for each worker.
External Codex took control at revision 2, completed two more DSH assignments
(3/3 each), read independent reviews and recorded acceptance itself. An operator
did not supplement these four final acceptance receipts.

The private synthetic [PR 2](https://github.com/qianrrf/aidsu-validation-20261104/pull/2)
received a native Codex review that found the planted addition/subtraction bug,
requested changes and associated head
`8a5025fad9bac0a2544697c760daa15c9d4d5846` (review ID `5774560558`).

Native Codex 0.162.1 successfully read a file and received an actual permission
denial when writing under the verified Windows unelevated configuration.
Loopback remained reachable; this is trusted-local operation, not network
isolation. Read-only metadata survived daemon restart.

Two required live checks remain unaccepted:

1. Round 39's resumed read-only prompt received HTTP 409 because continuations
   require a registered delivery operation. No model recovery was exercised.
2. Round 40's sleeping worker completed before the operator injected a daemon
   failure or interrupt. Its sentinel exists. The recorded result is
   `not_exercised`, not a cancellation defect and not a cancellation pass.

The loop evidence was collected on an earlier daemon revision. It does not
satisfy the requirement that all final-source business/fault checks pass on the
same commit as the final installer.

Further live qualification must first receive additional model allowance.
Its recovery probe must interrupt a newly registered read-only review while an
actual provider operation is observed active, then reconnect after daemon
restart and test the restored file boundary. Sending an ordinary continuation
to a completed controlled reviewer is not that recovery path. A worker
cancellation probe must wait for the durable running tool event of a bounded,
delayed synthetic command, inject cancellation immediately, and verify both
process termination and the absence of its completion sentinel. Elapsed time
or a model's claim that it started sleeping is insufficient. Register each
attempt before dispatch; do not automatically retry an unexercised injection.
Repeat both controller loops and PR review on the final runtime, retaining the
official operation receipts rather than operator-supplied acceptance records.

## Coexistence and remaining hard gates

At `01203b16b`, the complete two-installer coexistence check passed on a
disposable Windows CI runner (job `126986730178` in run `41948560176`). It built
the fixed original baseline and installed both actual NSIS packages. Both
applications ran concurrently with distinct default ports 3001/31338, daemon
PIDs, profiles, protocols and Start-menu shortcuts. Uninstalling either left
the other application's executable, protocol, shortcut and running daemon
available. Both final uninstalls retained their synthetic databases and
preservation files. The artifact `coexistence-result.json` reports `passed=true`
and zero model calls. The upstream application leaves its own protocol
registration (`upstreamProtocolRemoved=false`); do not describe upstream's own
cleanup as complete. Product protocol removal was verified.
That CI product installer was a separate build of `01203b16b`; the current
SHA-256 above identifies the local build of that same source used for its own
installation tests. The CI coexistence artifact reports completion at
2026-11-05T01:39:11Z. The preceding candidate and its evidence remain archived
under `previous-c296d69e7`.

Earlier coexistence failures were test issues: the fixed baseline's real
default port is 3001, and NSIS's relaunched uninstaller must finish removing
the executable, shortcut and product protocol before cleanup is asserted.
The corrected full check has now actually run successfully; the older partial
runs are not the basis for that conclusion. The original upstream installer was
not installed into the current user's environment.

Delivery remains blocked on the long-root Git path, the checkout correction's
complete validation, and the unexercised live recovery /
cancellation checks. Do not publish or label the candidate fully accepted.
The complete zero-failure/no-unexecuted-required-case gate has not been met.

At 2026-11-05 11:40–11:45 Asia/Shanghai, GitHub API status/artifact retrieval
failed with connection timeouts/resets. The watcher had already recorded the
Linux backend job succeeding in 39m1s and Windows race failing, but did not
record the latter job's final static result. Network diagnostics retained TLS
verification and made no system DNS/hosts/proxy changes. Download attempts
for both final backend JSONL artifacts failed; previous-source counts above
must not be substituted. Restoring connectivity and collecting run
`41948560176` is the next evidence step; a new complete CI run is unnecessary
just to retrieve an existing result.

The long-root probe is distinct from the successful deep-file probe. The
installed candidate's complete desktop smoke still fails at a physical
repository root exceeding MAX_PATH. Stock Git for Windows 2.56.0.windows.4
defaults long-path expansion to false until repository configuration is
initialized (`compat/mingw.c`, `are_long_paths_enabled`); initial directory
changes occur before that. Its `mingw_getcwd` also uses MAX_PATH-sized buffers
for both the inherited current directory and its final physical path; a short
junction therefore does not establish support for the long physical root.
Direct native Git, explicit Git/worktree paths and a scratch directory junction
also reproduced the failure. An isolated read-only check with the existing
2.53.0.windows.3 binary failed too; that private runtime is not a product
dependency or a proposed fix. A final isolated Unicode Windows drive-alias
probe also failed in three modes (inherited directory, `-C`, and explicit Git
directory/worktree); Git resolved the same long physical path. The exact owned
temporary mapping was removed in `finally` after verifying its target. The
initial `subst.exe` attempt failed before creating a mapping and is not a Git
test result. Evidence: `alpha2/long-subst-probe.json`. No persistent drive
mapping, patched Git, weakened root assertion or skipped case was introduced.

The diagnosis is based on the official
[Git for Windows 2.56.0 source](https://github.com/git-for-windows/git/blob/v2.56.0.windows.4/compat/mingw.c)
and actual process failures. Changing repository configuration or the Windows
policy alone has not solved this startup boundary. Supporting a shorter root
would change the agreed acceptance scope and has not been accepted as a fix.

Local evidence is under the validation cache's `alpha2` directory; model ledger
and live-session state are under `m3-lab`. These contain run state and must not
be committed. Public release, signing, account reconfiguration, recharge and
automatic updating remain outside the task.
