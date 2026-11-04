# Native model validation, 2026-11-04

This is historical alpha 1 evidence. The current allowance is 40 registered
rounds and is exhausted. Updated controller/PR review evidence and unexercised
recovery/cancellation gates are in [alpha 2 acceptance](alpha2-acceptance.md).

The isolated lab used stable Codex 0.162.1 via the existing relay and DSH
0.2.1-rc.5 account model `deepseek-account/deepseek-flash`. Exactly 20 rounds
were pre-registered, including failed launches and preflight attempts. No further
model calls are authorized by the current budget.

- Internal controller: `ultimate-m3-1`, ownership revision 3. Two DSH ACP
  workers ran concurrently in separate worktrees with per-command approvals.
- Clamp worker `ultimate-m3-4`: independent review found the intentionally
  omitted reversed-bound check. One registered rework produced
  `8492e401a22cb96cbebcb9233ee654c9af7a48ba`. Actual tests passed (2), fresh
  independent review approved, internal Codex accepted both assignments.
- Sum worker `ultimate-m3-5`: `413069fbf9499e22e76251fb952eda6e4701e74e`,
  actual test exit 0 and independent scoped approval.
- External stable Codex CLI read the inbox and existing receipts, took control
  exactly once (revision 3 to 4), and assigned integration. A stale revision-3
  worker dispatch received HTTP 409; existing results remained readable.
- Integration worker `ultimate-m3-11` cherry-picked both accepted histories to
  `ca44fe0afffd0f6b75996b2633800cddef49ef72`, with a clean tree. The daemon
  executed `node --test`: 3 passed, exit 0. Independent reviewer
  `ultimate-m3-12` approved the fixed committed tree.
- External Codex verified the evidence and decided acceptance. It posted the
  verdict itself; a PowerShell variable collision in the validation helper
  prevented its final acceptance POST. The operator corrected the helper and
  recorded only the missing `external-combined-accept` receipt from that exact
  Codex decision, without another model call or duplicate dispatch. This is
  operator-assisted validation, not proof of fully unattended operation.
- The stable process-only native sandbox probe actually read files and denied
  a native write with UnauthorizedAccessException. Loopback access succeeded
  and is a documented trusted-local limitation, not network isolation.
- Daemon restarts preserved the lab sessions and receipts. Restart/cancel at
  every requested fault point and installer recovery remain pending.

The private synthetic evidence PR is
https://github.com/qianrrf/aidsu-validation-20261104/pull/1. No content was sent
to the upstream repository. Native PR Reviewer execution was not tested because
the model budget was consumed by the higher-priority controller flows; no model
call was made for this PR. The PR is not a release or an acceptance substitute.

The lab ledger, approval decisions, session conversations and test receipts are
retained outside Git under the product cache's `validation/m3-lab` directory.
Do not publish raw lab dumps: unrelated local sessions must remain private.
