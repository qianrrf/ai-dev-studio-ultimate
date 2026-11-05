# Windows 个人版 alpha 3 验收记录

检查日期：2026-11-05。**当前结论：候选包，未通过最终验收。**
本记录替代 alpha 2 的跨平台发布门槛；Linux、公开 Release、商业签名和自动更新不在本轮范围内。历史通过不能代替本候选的复测。

## 源码与安装包

- 版本：`0.1.0-alpha.3`，Windows x64，未签名本地测试安装器。
- 安装包源码：`a04bbdd40d27e237e73320bcc6ef134ff8bb1555`。
- 固定上游：`b325d24b741297bf89296e76501d7e09018c4a04`，完整历史保留。
- 安装器：`dist/0.1.0-alpha.3/AI Dev Studio Ultimate Setup 0.1.0-alpha.3.exe`。
- SHA-256：`8de5aca93c2903eb01ccc00fbccaf0f99b36bbc02329769c44c83b47691009ff`，191,364,331 字节；构建时间 2026-11-05 08:40:17 UTC。
- 同目录 `SOURCE.json`、`SHA256SUMS.txt` 保存构建收据；本报告的文档提交不是安装包源码。Authenticode 实查为 `NotSigned`。
- 构建使用独立、干净的 `validation/product-alpha` Worktree 及其原有真实 npm-ci 依赖，未共享 node_modules、改变 Electron ABI 或替换原生模块。
- 原候选 `a743949cb`、`2c1ccd6a9` 的安装包、校验值及安装验证独立归档，不能算作新包验证。

本机原始证据统一位于：

`%USERPROFILE%\.cache\ai-dev-studio-ultimate\validation\alpha3`

原始日志、数据库、测试 Worktree、构建产物及账户资料不提交 Git。

## 本轮实现

| 提交 | 实现及验证目的 |
|---|---|
| `39a000039` | 所有后台入口统一拒绝物理 Git 根目录过长；检查仓库、Worktree、Reviewer 和 Git 管理目录；删除无效的通用进程长路径改写 |
| `ea726312b` | alpha 3 版本、Windows 专用验证工作流、安装版 HTTP/CLI 无副作用拒绝检查 |
| `a0f3d0368` | 关闭设置时取消延迟账户刷新，修复真实计时器泄漏；48 项设置测试通过 |
| `a743949cb` | Electron 缓存使用锁定 npm 包自带官方校验值，不再额外下载校验文件；32 项 Forge 测试通过 |
| `912f64915` | Codex 中断后通过原生线程接口清理后台终端；清理完成前禁止下一轮派发；失败返回错误，不能仅把记录改成取消 |
| `6ab4d1a3e` | Codex Chat/TUI 默认使用人工审批和 workspace-write；只有明确 bypass 才授予全权限；记忆权限和 UI 同步修正 |
| `97244d291` | 用虚拟时钟验证 100ms 请求预算，避免系统负载导致时限测试误报；原断言保留，连续 20 次通过 |
| `2c1ccd6a9` | 取消回归测试按完整静态检查要求使用 errors.Is |
| `a04bbdd40` | 安装版 UI 检查发现新建草稿仍显示自动审批；修正本机 Codex 默认显示，保留明确选择；324 项相关测试通过，安装版菜单确认人工审批默认选中且没有创建模型会话 |

没有新增 Bridge、Adapter、调度服务、任务状态机、路由或数据库迁移。

## 路径范围

Windows 首版根目录上限是 **200 个 UTF-16 单位**，不是字符数或字节数。检查使用实际物理路径：Windows Junction 也经原生句柄解析，不能用链接绕过。`.git` 指针、`commondir` 及未来 Worktree 管理路径一起检查。

导入、克隆预检查与暂存目标、Worktree 创建/恢复、多仓库工作区边界、Chat 启动/恢复和会话启动/重启均在产生 Git/Agent 副作用前检查。错误沿用 API 包装，代码为 `WINDOWS_GIT_PATH_TOO_LONG`，HTTP 400；CLI 显示同一错误。

建议仓库放在 `D:\Projects\项目名` 等短位置；需要时对测试进程显式设置较短的 `AO_DATA_DIR`。已有超长项目保留记录、日志和未提交文件，不能自动迁移或删除。查询、取消、拒绝审批不受启动路径门禁影响。

原先要求超长根目录成功的用例，按用户确认的范围改为验证提前拒绝且无副作用。中文、空格、深层文件的真实 Git 创建、提交、读取、Worktree 和清理用例继续保留；没有以跳过代替通过。

## 真实模型证据与额度

累计保守登记 **52/60 回合**，剩余最多 8 回合；失败和未发起的预留 Reviewer 45/46 仍计入，不删除历史或释放已计额度。账本位于 `validation/m3-lab/model-rounds.jsonl`。未充值、未切换账户、未切换官方 OAuth。

| 项目 | 观察结果 | 证据 |
|---|---|---|
| 原生只读恢复 | 同一 `a3-recovery-1` 会话重启前后均实际读取成功；恢复后 Set-Content 实际返回 Access denied，原文件未变 | `readonly-start.json`、`readonly-after.json`、`a3-recovery-1-conversation.json` |
| 原取消缺陷 | `a3-recovery-2` 的 native turn 已中断，但 Node 子进程仍运行并最终写出完成文件；明确失败，原证据保留 | `cancel-active-process.json`、`cancel-injection.jsonl`、`cancel-result.json` |
| 首次取消重测 | `a3-recovery-3` 在取消到达前已自然结束；不算通过，也不修改其完成文件 | `cancel-fixed-result.json` |
| 修复后的取消 | 自动脚本先观察运行中工具及真实 Node PID，再立即重启 Daemon；`a3-recovery-4` 取消返回成功，PID 已消失，超过原等待时间仍无完成文件 | `cancel-atomic-before.json`、`cancel-atomic-recovered.json`、`cancel-atomic-result.json` |
| 默认审批拒绝 | `a3-recovery-5` 实际产生 pending command approval；通过正式审批接口选择提供的 cancel，turn 中断，目标文件未生成 | `approval-deny-pending.json`、`approval-deny-response.json`、`approval-deny-after.json` |
| 等待审批时重启 | `a3-recovery-6` 的同一审批 ID 在重启后仍 pending；正式接口拒绝后正确结算，没有写入 | `approval-restart-before.json`、`approval-restart-recovered.json`、`approval-restart-result.json` |
| 双 DSH Worker | Codex 主控自行派发 `a3-loop-2`/`a3-loop-3`；两者 ACP 连接后均返回 Insufficient Balance，没有代码结果；主控停止，未重试或伪造验收 | `autonomous-state.json`、`a3-loop-1-conversation.json`、各 Worker conversation |

取消修复使用本机 Codex 0.162.1 的 `generate-json-schema --experimental` 确认的 `thread/backgroundTerminals/clean`，按 threadId 清理。工具事件的 processId 不是系统 PID，没有据此批量结束进程。OpenAI 官方文档页获取返回 403，因此协议依据明确限定为已安装版本和真实执行证据。旧 Codex 不支持清理接口时返回失败，不能报告执行进程已结束。

上述真实联调使用隔离 Daemon 和工作区构建；主要运行代码对应 `6ab4d1a3e`，其后 `97244d291`、`2c1ccd6a9` 只改变测试，`a04bbdd40` 仅改变前端草稿默认显示及对应测试。它们不能代替最终安装版的完整双入口业务闭环。alpha 2 的双入口验收及 PR Reviewer 是历史证据，未冒充 alpha 3 当前复测。

## 最终源码的无模型与安装验证

| 检查 | 结果 | 本机证据文件 |
|---|---|---|
| 完整 Windows 前端 | 10,935 通过、18 失败、15 未执行；18 项全部为 symlink EPERM。未删除、跳过或放宽安全断言 | `frontend-a04bbdd40.json`、`frontend-a04bbdd40-summary.json` |
| 完整 Windows 后端 | 16,330 条测试/子测试通过、25 失败、362 未执行，15 个失败包；包含父测试计数，不能当作独立用例数。失败为符号链接权限或其导致的凭据链接断言失败 | `backend-a04bbdd40.jsonl`、`backend-summary-a04bbdd40.json` |
| 完整 Windows race | GCC 16.2.0、CGO_ENABLED=1 实际执行完成；16,330 通过、25 失败、362 未执行，15 个失败包，0 数据竞争告警。失败用例与普通套件完全一致，exit 1，不能称为 race 全绿 | `backend-race-a04bbdd40.jsonl`、`backend-race-summary-a04bbdd40.json`、`race-compiler-a04bbdd40.txt` |
| Renderer 交互 | 148 通过、0 失败、0 未执行、0 flaky | `renderer-a04bbdd40.json`、`renderer-a04bbdd40.log` |
| TypeScript / e2e 类型 | 两项均 exit 0 | `typecheck-a04bbdd40.log`、`e2e-types-a04bbdd40.log` |
| Go build / vet | 两项均 exit 0 | `backend-build-a04bbdd40.log`、`backend-vet-a04bbdd40.log` |
| 全库静态分析 | 最终源码完整 golangci-lint 2.13.2 为 0 issues，exit 0；不是远端 CI 结论 | `lint-a04bbdd40.log` |
| OpenAPI、前端类型、sqlc | 最终源码重新生成，三个产物目录 git diff 均为零，exit 0 | `generated-a04bbdd40.log` |
| 原生 Electron API 边界 | 三项 Chromium 边界测试通过；对应代码自 `2c1ccd6a9` 未变 | `electron-boundaries-2c1ccd6a9.log` |
| 原生浏览器兼容 | 显式提供锁定原生程序后 34 通过、1 POSIX-only 未执行；代码自 `2c1ccd6a9` 未变 | `native-browser-2c1ccd6a9.log` |
| 共享包 | cloud-client 31 项、product-ui 162 项测试及生成/类型/构建通过；这些包自该次执行后未变 | `shared.log` |
| 构建 | 新源码独立 Worktree `npm run make` 成功，未更换 Electron 44.7.0 / ABI 149。首次重建误设 Python 路径失败，改为已安装 Python 的正确绝对路径后成功 | `build-a04bbdd40.log`、`build-a04bbdd40-python-fixed.log` |
| 安装生命周期 | 真实开始菜单、精简 PATH、覆盖升级、退出重开、停机备份哈希、恢复、完整卸载及重装通过；原合成项目、控制权、需求及已注册 Worktree 的未提交字节保留，0 模型调用 | `lifecycle-a04bbdd40/result.json` |
| 普通桌面及后台重启 | 安装版精简 PATH 找到官方 Git；中文/空格仓库、HTTP/CLI 超长根目录无副作用拒绝、控制权 CAS、需求去重、重启持久化、更新/云关闭及数据隔离通过；0 渲染错误 | `installed-clean-a04bbdd40/result.json` |
| 安装版人工审批菜单 | 实际打开新建 Codex 草稿，人工审批为选中项；退出后会话数仍为 0，截图人工查看通过 | `approval-ui-menu-a04bbdd40/result.json`、`manual-approval-menu.png` |
| 启动故障 | 端口占用和数据库独占锁均显示明确失败，未误连其他 Daemon | `startup-errors-a04bbdd40/result.json` |
| 二进制归属 | 安装目录 `ao.exe` 的 vcs.revision 为上述完整源码 SHA、vcs.modified=false，CLI 版本为 alpha.3 | `installed-binary-a04bbdd40.txt` |

测试工具为 Node 24.20.0、npm 11.19.0、Go 1.27.1。真实 Agent 固定 Codex 0.162.1、DSH 0.2.1-rc.5；Git 2.56.0.windows.4。未点击安装版显示的 Codex 更新通知。包装使用 MSVC 14.44.35207 和 Windows SDK；MSVC 不替代 Go race 的 GCC。

最终前端完整检查采用 `npm test -- --maxWorkers=4 --reporter=json`，未修改超时或断言。此前重测误用了已经重建为 Electron ABI 的打包工作区，产生 39 项额外 ABI 错误；该失败日志 `frontend-final-bounded-2c1ccd6a9.json` 独立保留，未混入最终 Node 测试结论。切回依赖匹配的测试工作区后，`2c1ccd6a9` 完整结果为 10,933 通过、18 权限失败、15 未执行；最终源码增加两项回归后为上表计数。

前端 15 项未执行逐项列于 `frontend-a04bbdd40-summary.json`：3 项条件原生浏览器测试已通过专门执行补测；其余为 1 项 POSIX socket、2 项 POSIX 进程时间、1 项 macOS quarantine、2 项 POSIX shell escaping，以及 6 项旧 Connect Mobile 功能。均未算作 Windows 个人版通过项。后端 362 项未执行的名称和原始原因在 `backend-skipped-a04bbdd40.json`：至少 37 项为缺少符号链接权限，必须保留为未完成；5 项为显式开关控制的真实 DSH 集成测试，当前余额错误也不能算补测成功。其余包括 POSIX fixture、Unix launcher、Homebrew、tmux 及可选运行时前提；3 项 Electron 边界已专门执行，不把其余缺失项笼统计作通过。

安装身份自 alpha 2 共存验证后未变，沿用该次两个安装器共存证据，明确属于历史验证。当前候选仍未完成 PR Reviewer、双入口真实完整业务及最终远端检查。

DSH 返回的原生请求 ID：`1d601ea4-8fa7-4338-bb16-10dcf4a2f7be`、`cd258c16-4947-4fc3-b475-d8078c33d012`。没有通过其他账户或额外模型绕过失败。

## 当前硬阻塞

1. **DSH 账户返回余额不足。** 当前版本双 Worker、后续独立审查/返工/正式验收、外部主控完整接管闭环不能验收；停止无效重试。
2. **GitHub 连接重置/超时。** 本轮新提交无法推送，远端 Windows 验证及既有专用私有仓库 `qianrrf/aidsu-validation-20261104` 的 PR/原生 Reviewer 复测未完成。未修改现有凭据或向上游发送内容。
3. **本机无符号链接创建权限。** 当前普通进程无 SeCreateSymbolicLinkPrivilege，Windows Developer Mode 也未启用；安全边界用例真实失败，不能删断言或改为 skip。恢复测试前提后需重跑完整套件。

此前缺失的 GCC 已补齐：官方 MSYS2 20260927 基础包的签名经官方文档指纹 `0EBF782C5D53F7E5FB02A66746BD761F7A49B0EC` 校验通过，随后 pacman 按正常签名校验安装 UCRT64 GCC 16.2.0-6。工具链仅在本轮测试进程 PATH/CC 中启用，没有改全局 PATH 或产品依赖；下载、签名及安装日志保留。原先 `gcc not found` 的记录是修复前结果，不能继续作为当前阻塞。

本机官方工具链位置为 `%USERPROFILE%\.cache\ai-dev-studio-ultimate\toolchains\msys2-20260927\msys64\ucrt64\bin`。复测时在当前进程 PATH 中加入该目录，将 `CC` 指向其中的 `gcc.exe`，设置 `CGO_ENABLED=1`、`GOTOOLCHAIN=local`，加载 `scripts/ultimate/prepare-windows-test-process.ps1` 后在 `backend` 执行 `go test -race -json -count=1 -timeout=40m ./...`。本轮完整命令已结束，未因权限失败提前停止其他包。

GitHub 最后复查中，`github.com` HTTPS 连接失败、`api.github.com` 12 秒超时；没有修改代理、退出账户或替换凭据。专用真实联调 Daemon 已通过自身 CLI 正常停机，合成项目、失败证据和工作区保留；见 `live-stop.log`。

## 继续执行顺序

1. 恢复本机符号链接测试前提后，重跑完整前后端及 race；保留全部 skipped 清单并逐项判断范围。当前进程未提升，不能直接代替系统安全桌面的人工确认；未尝试绕过该限制。
2. GitHub 恢复后推送阶段提交，只运行 Windows 验证；下载最终报告，不能借用旧提交的成功结果。
3. DSH 可执行后，先核对账本剩余额度；在剩余 8 回合内优先跑双入口、双 Worker 和原生 PR Reviewer。不足时明确保留未完成，不能超额调用。
4. 若修补运行源码，重建新候选，重新关联源码/hash及受影响安装和业务检查。
5. 只有全部个人版硬门槛满足才更新本报告为验收通过。安装成功、单项通过、旧版闭环均不能替代这一结论。

本报告与随包收据均保持“未验收”状态；修复安装成功不代表业务和完整测试已通过。
