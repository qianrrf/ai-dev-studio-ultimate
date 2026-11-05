# Local Windows alpha build

The product identity is AI Dev Studio Ultimate, version `0.1.0-alpha.3`.
The executable is `ai-dev-studio-ultimate.exe`, application ID is
`dev.ai-dev-studio-ultimate.desktop`, and protocol is
`ai-dev-studio-ultimate`. The daemon remains loopback-only and defaults to
port 31338 (desktop development uses 31340). A daemon from another installation
must fail its existing executable-identity check, rather than be adopted.

## State and coexistence

Default state is `%USERPROFILE%\.ai-dev-studio-ultimate`: daemon data,
run-file, worktrees, logs, desktop preferences and Electron profile. Explicit
`AO_DATA_DIR` / `AO_RUN_FILE` overrides remain available for isolated tests.
Use the bundled `resources\daemon\ao.exe` CLI by absolute path. The product
does not install a global `ao` command on Windows. Neither product nor upstream
state directories may be imported as a project.

The alpha has no hosted cloud offering, upstream update feed, telemetry exporter
or error-reporting transport. Main-process cloud and update entry points use
local-only implementations without loading hosted SDKs. Renderer analytics
initialization is disabled and its SDK has no runtime import. Update preference
writes persist `enabled=false`; install-update requests are rejected. No
publisher is configured. Keep the original Apache-2.0 license, copyright and
third-party license texts in the packaged resources.

## Build

Use an independent worktree at the recorded product commit. Run real `npm ci`
in `packages/product-ui` and `frontend`; do not share another checkout's
`node_modules`. Use the exact Windows toolchain documented in
[baseline reproduction](windows-build-baseline.md). In the Visual Studio x64
developer environment, run `npm run build` in `frontend`, followed by
`npx electron-forge make --skip-package` against that package. The equivalent
clean CI entry is `npm run make`. No publishing is part of either command.

`prepare-notices.mjs` collects license texts from installed, locked dependencies
before packaging. Build outputs and generated license collections stay out of
Git. The final artifact must be accompanied by a SHA-256 and its source commit.
Go sources use `*.go text eol=lf` so Windows checkout does not make goimports
report every CRLF source file as a formatting failure. Keep this rule even
when a developer's Git configuration already has `core.autocrlf=false`.

## Validation status

Product configuration tests, native Codex read-only argument/reconnect tests,
project path protection, Go build/vet, frontend typecheck and local-only service
boundary tests passed before packaging. Startup/discovery/launch/cloud-session
focused tests passed. A shared telemetry test requires a Windows symlink
privilege unavailable on this host; it fails before the behavior under test.
Full-suite and installer results are recorded separately after execution.

The current model budget is 52 registered rounds out of an authorized maximum
of 60. DSH returned Insufficient Balance; retry only after account availability
changes. Do not recharge or switch accounts. Current artifact, fixed
versions, actual recovery evidence and remaining business gates are recorded in
[alpha 3 acceptance](alpha3-acceptance.md).
A generated installer is a candidate, not evidence that every release gate
has passed. Public releases, signing and automatic update publication are out
of scope.

The latest local candidate, its exact source/hash, installed Electron smoke,
installed lifecycle checks and successful two-installer coexistence results are recorded in
[alpha 3 acceptance](alpha3-acceptance.md). Alpha 2 coexistence evidence remains
historical and applicable only to unchanged installer identities. The older
installer-validation file describes alpha 1 only. See the
[Chinese quickstart](quickstart.zh-CN.md) for installation and controller handover.
