# Windows baseline build — 2026-11-04

The unmodified audit checkout at `b325d24b741297bf89296e76501d7e09018c4a04`
produced a Windows x64 NSIS installer. This verifies build generation only;
the upstream installer has not been installed in the current user's profile.

- Output: `Agent Orchestrator Setup 0.14.8.exe`
- Size: 192453434 bytes
- SHA-256: `769c21adb4e8027c4474e4eaf0276bd5b146da5175abd5bdb157829f71997287`
- Authenticode inspection: `NotSigned` (builder's signing log lines do not
  establish a signed artifact).
- Electron 44.7.0 / native ABI 149; matching native dependency preparation
  completed. No Electron ABI substitution or copied native module was used.
- Go 1.27.1 (`GOTOOLCHAIN=local`, `CGO_ENABLED=0`), Node 24.20.0,
  npm 11.19.0, Python from the bundled Codex runtime, MSVC 14.44.35207.

## Reproduction

Use a separate worktree and its own `npm ci`. Set process-only Go/Python paths
and invoke the actual installed `Common7\Tools\VsDevCmd.bat -arch=x64
-host_arch=x64` before `npm run build` in `frontend`. After packaging succeeds,
run `npx electron-forge make --skip-package` in the same checkout. Do not publish.

The task cache holds logs `validation/baseline-msvc-build.log` and
`validation/baseline-installer-make.log`, the original worktree
`validation/baseline-repro`, and its `frontend/out/make` artifacts. They are
intentionally excluded from Git. The process-only download hook uses the
checksum-verified official Node 22.23.2 archive cached for the bundled ACP
runtime; it does not substitute arbitrary downloads or bypass validation.

## Toolchain completion

Microsoft setup completed at 21:52 with exit code 0. A fresh `vswhere` query
confirmed `isComplete=true`, `isLaunchable=true`, and no reboot required.
The earlier optional CMake download stalled but eventually completed through
the native installer. Compiler installation and native packaging do not
establish Go race support: a compatible GCC-style C compiler or CI is still
required. Catalog/payload mismatches observed earlier remain failed checks;
no mismatched payload was injected into setup.

The upstream build still includes upstream services. Product identity, service
disablement, install coexistence and business-loop verification are subsequent
gates; this installer is not the requested product deliverable.
