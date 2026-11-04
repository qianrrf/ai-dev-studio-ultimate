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
| M1 build | Original baseline worktree retained; Microsoft installer restarted with explicit VC tools and SDK selections; still downloading | Complete compiler installation, original installer and build log |
| M2 control | SQLite CAS ownership/release, intake with attachment preservation, transactional seed/dispatch reservation, HTTP and CLI source/tests added | UI/automation intake, report routing, complete concurrency/recovery tests and actual handovers |
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
