# Alpha 2 acceptance record

Updated 2026-11-05 (Asia/Shanghai). **Not accepted for delivery.** An installer
build, a successful installation, and complete business acceptance are separate
gates. This record supersedes the alpha 1 status in the older reports.

## Candidate and fixed inputs

- Product source: `90ff261abf03a6cbb559873dd542af78309fcffe`.
- Upstream baseline: `b325d24b741297bf89296e76501d7e09018c4a04`.
- Version: `0.1.0-alpha.2`, Windows x64, unsigned.
- Local installer: `dist/0.1.0-alpha.2/AI Dev Studio Ultimate Setup 0.1.0-alpha.2.exe`.
- SHA-256: `bed88fdd4d2f02099d638c1bafddf8d2db8b70c4fa2e5227ce66a5dfa2d42cdb`.
- Size: 191339008 bytes. `SOURCE.json` and `SHA256SUMS.txt` accompany the file.
- Build inputs: Node 24.20.0, npm 11.19.0, Go 1.27.1, Electron 44.7.0,
  better-sqlite3 12.11.1, `CGO_ENABLED=0`. Native dependencies were rebuilt for
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

Observed in the dedicated Chinese/spaced installation directory:

| Check | Observed result and scope |
|---|---|
| Fresh alpha 2 build and silent installation | Exit 0; expected installed executable and version exist |
| Real Electron and bundled daemon | Native renderer, ready daemon, product title, zero renderer page errors in the completed core smoke |
| Disabled hosted services | Update enable attempt stays disabled; cloud account remains null |
| Bundled CLI and synthetic Chinese project | Project registration, stale-revision conflict, intake deduplication and release work |
| Daemon restart | New PID; control revision and intake survive |
| Start menu at the current candidate | Real installed shortcut opens the product with a reduced PATH; old control/intake and dirty file remain intact |
| Alpha 1 to alpha 2 upgrade | Old project, revision 2, one intake and zero model sessions remain readable |
| Stopped-data backup/restore | Durable data restored to a second scratch root; same project/control/intake readable |
| Dirty files | A real Git-registered worktree's uncommitted README hash remains unchanged through upgrade and uninstall |
| Uninstall/reinstall | Exit 0; executable, product shortcut and product protocol removed; synthetic database and dirty file retained; reinstall succeeds |
| Long workspace root | **Fails.** Physical root exceeds MAX_PATH; Git/CreateProcess rejects it |

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

| Scope | Actual evidence |
|---|---|
| Local renderer interaction suite | 148 passed, no skips |
| Local complete Go static analysis at `8b90666f7` | 0 issues with golangci-lint 2.13.2 |
| Windows HTTP/process/persistent-host complete packages | Passed with a private test TEMP ACL |
| Sharing-lock and IPv6 shutdown regressions | 20 consecutive runs passed |
| Linux frontend at `8b90666f7` | 11004 passed, 0 failed, 13 pending; CI job succeeded |
| Linux renderer at `8b90666f7` | 148 expected, 0 unexpected, 0 flaky, 0 skipped |
| New symlink checks locally | Fixture creation fails with EPERM on this non-elevated host; not labelled passed |
| Native browser focused suite locally | Pinned Windows binary enabled the three native compatibility cases; package passed, with its separate Unix socket case inapplicable on Windows |
| Final-source complete Windows/Linux CI | Pending; no all-green conclusion |
| Windows long workspace operations in CI | Fails even after runner enables Developer Mode and LongPathsEnabled |

The CI workflow now runs complete build/vet/race/lint, API/sqlc drift, frontend
types/full tests, renderer interactions and native browser compatibility.
Windows race uses MSYS2 UCRT GCC. Each generator/type/lint check is independent
of previous test failures. The 40-minute race package timeout permits the full
SQLite migration suites; it is not a passing result by itself.

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

## Coexistence and remaining hard gates

The current user has not installed the original upstream application. Its
portable coexistence evidence cannot substitute for two installed products.
The new CI installer check builds the fixed original baseline and installs both
NSIS packages on a disposable Windows runner. It checks simultaneous startup,
default ports 31337/31338, distinct profiles/protocols/shortcuts, and uninstalling
each while the other remains available. That new check is **not yet verified**.

Delivery remains blocked on the long-root Git path, final complete CI and
symlink results, two-installer coexistence, and the unexercised live recovery /
cancellation checks. Do not publish or label the candidate fully accepted.
The complete zero-failure/no-unexecuted-required-case gate has not been met.

Local evidence is under the validation cache's `alpha2` directory; model ledger
and live-session state are under `m3-lab`. These contain run state and must not
be committed. Public release, signing, account reconfiguration, recharge and
automatic updating remain outside the task.
