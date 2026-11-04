# Remaining implementation from faf97410a

Target: unsigned Windows x64 `0.1.0-alpha.1`. Installer generation, installation
success, and the business loop are separate acceptance gates. No completion
percentage is inferred from source changes.

## Fixed decisions

- One trusted local repository per project, one effective Codex controller,
  internal/external handover, at most two executing DSH Chat workers.
- Preserve the existing relay and DSH account. No forced OAuth migration.
- Native filesystem read-only review; hostile local processes and complete
  loopback isolation are outside the alpha boundary. External tools, plugins,
  applications, child agents and elevation remain disabled for review sessions.
- Maximum 20 new real model turns, including failures, retries and cosmetic
  calls. At this checkpoint zero new model turns have been sent. Every future
  call must first be entered in the external validation ledger.
- Private synthetic test PRs/comments are authorized. Public publishing,
  recharge and account credentials in CI are not authorized.

## Milestones

| Gate | Current observed evidence | Remaining acceptance |
| --- | --- | --- |
| M1 build | Microsoft setup completed (exit 0; vswhere complete/launchable); unmodified baseline Windows x64 package and unsigned NSIS installer generated, SHA-256 recorded | Original installer isolation/run check |
| M2 control | SQLite CAS ownership/release, attachment-preserving intake, transactional dispatch, two-worker concurrency tests, HTTP/CLI and settings controls, automation intake and report routing | Fence existing-session continuations with delivery receipts; real handovers |
| M3 loop | Persisted readOnly foundation; stable native filesystem probe passed | Immutable base/head review, actual test evidence, bounded rework and both controller loops |
| M4 product | Constants and isolated product identity fixed in the accepted plan | Branding, data/port identities, disabled upstream services, candidate package |
| M5 install | Not executed | Installation, recovery, upgrade, uninstall, coexistence, SHA-256 and source tag |

## Checks at this checkpoint (2026-11-04)

- Storage suites (`go test ./internal/storage/sqlite/...`) passed after the
  transaction change. Focused control tests also passed after attachment changes.
- New control HTTP and CLI tests passed: initial ownership, stale revisions,
  monotonic release, invalid internal owner, old-worker entry rejection,
  durable intake/attachments, and preserved HTTP error envelope/request ID.
- Full HTTP package command executed. Router/spec tests passed. Controllers
  failed in path handling, Windows secure-pairing rename, symlink privilege,
  and clone URL tests. These need comparison with retained baseline failures;
  the full command is not marked passed.
- ReadOnly Codex driver tests passed. The zero-model native probe at 20:11
  used stable CLI 0.162.1 with process-only `unelevated`. Read succeeded,
  `Set-Content` raised `UnauthorizedAccessException`, and the target did not
  exist. The checking shell caught that exact exception and exited normally.
  Other RPC/setup errors still fail the probe. Loopback succeeded and is
  reported independently; strict network isolation is not claimed.
- MSVC does not establish Go race-test support. No installer, signing,
  installation or full business-loop pass is claimed.

Raw logs and scratch workspaces remain outside Git under the task cache.
The previous uncommitted Worker calculator fix is preserved.

## Control checkpoint (20:56)

- Settings now shows controller ownership/revision, retained requirements and
  explicit internal/external takeover/release actions. Merely opening the page
  does not start a model. The daemon pins Codex/DSH/Codex roles on activation.
- Concurrent dispatch testing admits exactly two of three simultaneous worker
  seeds; cross-project internal controllers and stale/released claims fail.
- Focused storage/controller/service tests and frontend typecheck passed.
  Control/settings/API-client frontend suites: 99 tests passed. Earlier intake
  and New Chat suites: 197 tests passed.
- Complete CLI suite rerun after command classification repair: only the retained
  baseline Windows symlink test failed. All four full-HTTP failure names also
  occur in the retained baseline log; neither full suite is labeled passed.
- Complete storage suite, Go build and vet previously passed. Subsequent delivery
  changes still require their own complete affected suites before handover.
- Microsoft setup remains incomplete (`vswhere isComplete=false`), stalled on
  CMake package delivery after compiler/MSBuild files appeared. This is not a
  verified compiler installation or an original-installer pass.
