# AI Dev Studio Ultimate — implementation record

This directory records the fork's audit, reproducible validation, and release
gates. A documented capability is not a passing live test.

Current Windows personal-use alpha 3 candidate, observed results and outstanding
acceptance gates: [alpha 3 acceptance](alpha3-acceptance.md).
[Alpha 2](alpha2-acceptance.md) and alpha 1 results are historical.

## Source and scope

- Upstream: https://github.com/OrchestratorInc/agent-orchestrator
- Fixed baseline: `b325d24b741297bf89296e76501d7e09018c4a04`
- Baseline tree: `9fdd0fe2fd1b260d49941fb6ce64720a0699e5c0`
- Local branch: `aidsu/windows-mvp`; complete baseline ancestry retained.
- Hosted fork: https://github.com/qianrrf/ai-dev-studio-ultimate
- `origin` tracks that fork. Earlier commits were pushed; alpha 3 commits remain
  local while GitHub connections reset. Do not infer current remote CI success.
- Target: native Windows, one trusted local Git repository per project, one
  effective Codex controller, two DSH Chat workers, separate Codex review.
- Preserve Apache-2.0 and the original component/third-party notices.

See [source audit](source-audit.md) and [validation](validation.md).
中文交付进展、阻塞与继续步骤见 [交付记录](delivery.zh-CN.md)。
Remaining implementation starts from `faf97410a`; see
[milestones and current evidence](remaining-implementation.md).

## Delivery gates

1. A: reproduce baseline build/tests, isolate all test data, commit evidence.
2. B: exercise the original DSH adapter and Codex drivers; distinguish protocol
   startup from billable turns, and ACP from headless execution. Record failures.
3. C: implement controller fencing and task routing, durable read-only local
   reviews, evidence and a maximum of two business rework rounds. Reuse existing
   session/queue/workspace/recovery machinery. Verify before branding.
4. D: product identity, state isolation, disabled upstream updates/telemetry,
   Windows installation and coexistence. No release until critical gates pass.

Only the authorized private synthetic test repository may receive test PRs and
review comments. Public releases and upstream messages remain outside scope.
Do not label unexecuted checks as passed. Keep secrets and local run state out
of commits.
