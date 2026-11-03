# AI Dev Studio Ultimate — implementation record

This directory records the fork's audit, reproducible validation, and release
gates. A documented capability is not a passing live test.

## Source and scope

- Upstream: https://github.com/OrchestratorInc/agent-orchestrator
- Fixed baseline: `b325d24b741297bf89296e76501d7e09018c4a04`
- Baseline tree: `9fdd0fe2fd1b260d49941fb6ce64720a0699e5c0`
- Local branch: `aidsu/windows-mvp`; complete baseline ancestry retained.
- Hosted fork: pending GitHub authentication; no `origin` has been fabricated.
- Target: native Windows, one trusted local Git repository per project, one
  effective Codex controller, two DSH Chat workers, separate Codex review.
- Preserve Apache-2.0 and the original component/third-party notices.

See [source audit](source-audit.md) and [validation](validation.md).
中文交付进展、阻塞与继续步骤见 [交付记录](delivery.zh-CN.md)。

## Delivery gates

1. A: reproduce baseline build/tests, isolate all test data, commit evidence.
2. B: exercise the original DSH adapter and Codex drivers; distinguish protocol
   startup from billable turns, and ACP from headless execution. Record failures.
3. C: implement controller fencing and task routing, durable read-only local
   reviews, evidence and a maximum of two business rework rounds. Reuse existing
   session/queue/workspace/recovery machinery. Verify before branding.
4. D: product identity, state isolation, disabled upstream updates/telemetry,
   Windows installation and coexistence. No release until critical gates pass.

Do not publish packages, PRs or GitHub review comments as an incidental test.
Do not label unexecuted checks as passed. Keep secrets and local run state out
of commits.
