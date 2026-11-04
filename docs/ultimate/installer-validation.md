# Windows 候选安装包验收 — 2026-11-05

This is the historical alpha 1 report. Current results and outstanding gates
are in [alpha 2 acceptance](alpha2-acceptance.md). The old counts, model allowance
and core-only smoke interpretation below do not describe alpha 2.

结论：已生成、测试安装并卸载 `0.1.0-alpha.1` Windows x64 候选包。业务联调仍有人工辅助，完整测试与发布门槛未全部通过。本报告不会把安装成功等同于产品完整验收。

## 产物身份

| 字段 | 实际值 |
|---|---|
| 文件 | `AI Dev Studio Ultimate Setup 0.1.0-alpha.1.exe` |
| 大小 | 191,348,267 字节 |
| SHA-256 | `65ec49f009a093634db8fbb5bc4eafebbbb1f79cbefb5c9dab07056734a980ef` |
| 构建源码 | `f81e71519bc49af7af1832ae82add0e254f3426b` |
| 随附 CLI 构建时间 | `2026-11-04T17:35:32Z` |
| 数字签名 | `NotSigned`，符合本次未签名测试包范围 |
| 应用 ID | `dev.ai-dev-studio-ultimate.desktop` |
| 协议 / 可执行文件 | `ai-dev-studio-ultimate` / `ai-dev-studio-ultimate.exe` |
| 默认状态根目录 | `%USERPROFILE%\.ai-dev-studio-ultimate` |
| 默认主监听 | `127.0.0.1:31338` |

保留原项目的可选 IPv6 loopback 同端口监听；实测 `ipv6Loopback=true`，所以不能声称只有 IPv4。没有新增公网监听。可信本机模式不承诺阻断恶意本机进程或所有 loopback。

最终包位于仓库忽略的 `dist/0.1.0-alpha.1/`，随附 `SOURCE.json`、`SHA256SUMS.txt`、使用说明、截图及结构化测试记录。安装器没有上传公开 Release。

## 构建与包检查

- 固定独立产品 Worktree；其依赖为真实 `npm ci`，不共享其他 checkout 的依赖目录。
- Go 1.27.1，Node 24.20.0，npm 11.19.0，MSVC 14.44.35207，Electron 44.7.0 / ABI 149；没有更换 ABI 或注入不匹配的原生模块。
- 最终构建日志：本地验证目录的 `m5-lint-qualified-build.log`。随附和安装后 CLI 均报告上述源码提交。
- 实际 ASAR 检查确认产品名称与版本，Windows UTF-8/UTF-16 环境读取修补在最终 main bundle 中。
- 实际资源中没有 `app-update.yml`、cloudflared 或 PostHog 文件；主进程云/更新/遥测使用禁用边界。保留的 SDK 算法测试不代表产品重新开启出口。
- 根许可证、原作者 NOTICE、第三方许可文本均存在，第三方许可集合约 2.42 MB。用户凭据与原始模型对话未随包分发。

## 安装版无模型实测

下列项目使用真实安装后的加固 Electron、原生 preload 和随附 Daemon，非模拟网页。通过 Chromium 调试连接测试，生产包没有打开 Node inspector。

| 检查 | 结果与边界 |
|---|---|
| 中文和空格安装路径 | 安装退出 0；最终版本和文件存在 |
| 覆盖安装 | 多次候选版本覆盖，最后安装为精确构建提交；应用重开后旧测试控制权仍可读取 |
| 原生界面 | `app://renderer/index.html`；标题、侧栏品牌正确；最终 smoke 的 renderer pageerror 为 0 |
| 独立配置 | Electron Cache 位于指定 scratch 数据目录下，没有使用原版用户数据 |
| 更新与云 | 尝试开启更新后仍为禁用；cloud session 为 null |
| CLI 与中文项目 | 随附 CLI 正确发现本产品后台并注册中文/空格 Git 路径 |
| 控制权 | CAS 接管成功、旧修订 409、重复需求仅一条、释放递增修订 |
| 后台重启 | PID 改变、界面重连、sessions API 200、控制权和需求保留 |
| 零模型 | 此阶段未创建任何主控、Worker 或 Reviewer 会话，sessions 数量为 0 |
| 默认端口 | 产品使用 31338；不依赖开发端口 |
| 端口冲突 | 后台退出 1，详细错误包含绑定失败；测试占用服务仅收到 health probes，没有收到 sessions/projects 请求；无随机端口回退 |
| 文件锁 | 独占锁定 scratch SQLite 后启动退出 1、无运行发现文件，数据库前后 SHA 相同 |
| DSH `.cmd` | 原生版本命令返回 `0.2.1-rc.5`；ACP 真实执行证据在先前原生联调报告中 |
| 长路径 | 约 300 字符 synthetic 路径被本机 Git 拒绝；`LongPathsEnabled=0`。这是未通过项，不计入核心 smoke 的成功断言，完整长路径链路未验收 |
| 共存 | 原版 portable 与产品同时运行，不同数据、后台 PID 和端口；退出产品后原版健康检查仍为 200 |
| 重复启动 | 同一产品配置的第二实例退出 0，未替换已有后台 |
| 卸载 | 退出 0；测试安装的 exe、卸载项、链接协议及开始菜单快捷方式移除；测试数据库和 synthetic 项目保留；原版卸载项集合未变 |

原版安装器没有在当前用户下安装，故共存结论限定为产品安装版与原版 portable 的同时运行，不宣称完成两个安装器的相互覆盖矩阵。卸载实测使用专用测试目录；用户默认生产数据目录保留行为未单独验收。未模拟旧版生产数据库升级。

协议由安装器 lifecycle hooks 注册与清理；实测首次启动前已注册，最终卸载后消失。最后候选代码还修正了导入格式、API 注释、包裹错误匹配与只读 MCP 配置类型检查，并删除禁用下载后不可达的旧安装函数；API 重新生成后无差异。

修复过程发现并解决两个实际 Windows 问题：终端环境输出按错误编码解析导致中文运行发现路径乱码；上游端口冲突自动改用随机端口。前者加入实际 PowerShell/cmd 子进程的中文路径回归检查，后者加入真实端口占用回归检查。

可重复运行安装版 smoke（需要仓库开发依赖；只使用新的 scratch 目录，不会调用模型）：

```powershell
node scripts/ultimate/verify-installed-desktop.cjs '安装目录\ai-dev-studio-ultimate.exe' '新的测试目录' 31341 31342
```

脚本的 `passed` 仅表示所列核心无模型检查通过。长路径结果单独记录在 `longPath.supported`，必须一起读取；它不是完整发布门槛的通过标记。

## 完整测试与 CI

| 验证 | 实际结果 |
|---|---|
| Go 最终 build / vet | 通过 |
| 前端最终类型检查 | 通过；本次最后的 Go 修补未改变前端或生成 API 类型 |
| 固定版 golangci-lint 2.13.2 | 使用官方预编译包并校验官方 SHA-256；初查 100 项，修正本次新增的 39 项后仍有 61 项。相对固定上游基线的变更行检查为 0 issues；这不等于全库 lint 通过 |
| 最后 Go 修补后相关完整包 | SQLite store、automation、HTTP 核心及 spec 包通过；session、CLI、codexappserver 和 controllers 仍失败。Codex 驱动套件的 Unix 绝对路径 fixture 在 Windows 失败，并出现测试结束后 goroutine 报错；未宣称全套通过 |
| 实际 Windows shell / 本地化回归 | 36 个通过，包含两种原生 shell 中文环境探针 |
| BrowserView 及本地化 focused | 64 个通过；完整套件中出现过的普通弹窗时序失败未在 focused 复现 |
| 完整前端（修复环境编码后） | 608 文件通过 / 32 失败；10,804 测试通过 / 192 失败 / 17 跳过 |
| 与 M3 失败名比较 | 上述 192 个失败名均存在于 M3 Windows 结果；M3 另一个 HarnessUpdateNotice 失败在本次未出现。不是全绿 |
| 完整 Go 基线与产品套件 | 均已执行；Windows 路径、符号链接、CRLF、shell 和 secure-pairing fixture 等仍失败；不泛化未核查的失败原因 |
| 最终 HTTP 全套 | HTTP 核心、API spec、specgen 通过；controllers 有四类 Windows fixture 失败：路径、secure-pairing rename、符号链接、file URL |
| 后端品牌 fixture | CI 发现的六个 PTY 注册表路径、安装 exe 路径、默认 Codex 及 dev namespace 预期已经修正；focused 通过。完整受影响包仍有其他失败 |
| CLI 最后全套 | 一个已有 Windows symlink 权限失败；没有标成通过 |
| Linux 完整 race | 已运行过；先前运行报告失败，不是 race 全套通过证明。未把 MSVC 安装等同于 race 通过 |
| Windows CI 构建 | `1b97fbbeb` 和 `c6809aa17` 的打包 job 成功；只是构建验收，非业务验收 |
| 候选代码的完整 CI | 尚无全部通过结论；当前状态见随包 evidence/ci-status.json 与工作流链接 |

已修复的本次新增失败包括中转站授权门槛、品牌预期、本地化产品名称清单、托盘文案、重启 bridge 测试 fixture、遥测禁用后的测试 transport。没有通过改 Electron ABI、跳过文件只读验证或全部自动批准来掩盖问题。

CI 入口：<https://github.com/qianrrf/ai-dev-studio-ultimate/actions/workflows/ultimate-validation.yml>。`c6809aa17` 的运行号为 `41872076413`；较早 `7f036da6d` 的运行号为 `41876222955`，它不是最终候选代码。2026-11-04 17:40 UTC 查询时，`fda9250d3` 的运行 `41873742462` 仍在执行：Windows 打包 job 成功，前端的完整测试步骤报告 failure，backend race 步骤仍执行中。最终代码为 `f81e71519`，随包 CI 快照记录推送后的实际状态。后续只修改验收文档时，新的 CI 不改变安装包 provenance。不能把其他提交的成功 job 当作最终提交的完整通过。

## 业务与发布缺口

- 20 次模型额度已全部使用，未额外发起模型调用；产品内与外部主控均有真实执行，但外部验收有操作员辅助。
- 原生 PR Reviewer 未测；全部故障点与无人值守闭环未完全验收。
- Windows 长路径、生产升级/双安装器矩阵和完整 CI 全绿仍未通过。
- 默认数据为可信本机单用户模式。只读审查实际拒绝文件写入，但允许某些本机网络访问。
- 本轮没有公开发布、商业签名或启用更新源。候选包可供本机检查，不能据此宣布原计划全面完成。

完整模型证据见 [原生验证](native-validation.md)，用户操作见 [使用说明](quickstart.zh-CN.md)。原始日志保留在本地缓存，分享前必须筛查隐私。
