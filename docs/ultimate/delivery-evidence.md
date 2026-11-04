# Controller delivery evidence

This implementation checkpoint is not an installer or a completed M3 acceptance.
The daemon retains session lifecycle ownership. The new delivery table contains
immutable operation requests and completion receipts, not another task state machine.

## CLI and HTTP

`ao session delivery <worker>` reads receipts, including commit freshness derived
from the live worker and reviewer worktrees. POST and GET share
`/api/v1/sessions/{sessionId}/delivery`.

Every mutation supplies a stable `--request-id`, `--controller`,
`--control-revision`, and full `--base` / `--head` commit SHAs:

1. `--kind test -- node --test <file>` executes the actual command. The daemon
   stores its exit code, bounded output, timeout/cancellation and errors.
2. `--kind review --test-id <test>` requires a successful matching test receipt,
   then starts a fresh Codex Chat read-only session in a distinct fixed-commit
   worktree. Include the worker's assigned scope in `--message`; unrelated base
   defects belong to their own assignments.
3. `--kind verdict --review-id <review>` reads the completed reviewer's durable
   assistant output. Its first final line must be `APPROVED` or
   `CHANGES_REQUESTED`; the controller cannot supply that verdict itself.
4. `--kind rework --review-id <verdict> --message <instructions>` continues the
   existing worker only after a changes-requested verdict. One verdict permits
   one continuation; at most two business rework rounds are registered per worker.
5. `--kind accept --review-id <verdict> --test-id <test> --message <conclusion>`
   requires an approving verdict and the exact successful test it reviewed.

Ownership checks and receipt registration share one SQLite writer transaction.
Previously registered work may finish after takeover. Stale new requests conflict.
Incomplete receipts remain visible and are not blindly rerun. A daemon crash
between execution and persistence requires reconciliation from the session,
worktree and logs; the API does not infer success from a missing result.

The worker must have committed every change. Changed HEADs, dirty worktrees,
changed reviewer paths or lost read-only constraints refuse acceptance. Direct
worker messages, edits, steering and side conversations cannot bypass registered
rework. A readonly reviewer cannot create independent side agents. Cancellation
and approval rejection retain their existing paths.

## Windows implementation checks, 2026-11-04

- The new focused storage, CLI, service, manager and HTTP/spec tests pass.
- Real-Git tests cover actual command exit codes, evidence mismatch, stale HEAD,
  dirty reviewer invalidation and fixed-commit worktree creation. A real ACP
  attempt exposed the old fresh-branch resolver overwriting the requested SHA;
  this now has a regression test in the workspace adapter.
- Frontend type checking and 90 control/settings/localization tests pass.
- The repeated complete frontend suite reports 605 files passing, 33 failing;
  10,797 tests pass, 193 fail, 17 skip. Every failure name exists in the original
  Windows baseline. The new localization and settings-fixture failures found on
  the first run were fixed before the repeat.
- Repeated complete connected backend suites still report original Windows
  symlink, path, CRLF, shell-fixture and secure-pairing failures; no new failure
  names appeared. These suites are not labelled passed.
- `Ultimate validation` runs the complete backend race suite, frontend suite,
  types, API drift and linter in CI. Windows packaging is enabled only once the
  package version is the product's `0.1.0-alpha.1`; it never publishes releases.

Real model evidence, approval decisions and the pre-registered 20-round budget
ledger are retained under the isolated validation directory, outside Git. Product
acceptance and installer qualification must be recorded separately after they run.
