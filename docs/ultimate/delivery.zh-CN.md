# AI Dev Studio Ultimate 开发交付记录

核查日期：2026-11-04。当前是已审计并完成部分原生验证的开发分支，尚未达到可安装产品的验收门槛。

## 已完成的工作

1. 保留上游完整 Git 历史，以固定提交 `b325d24b741297bf89296e76501d7e09018c4a04` 为基线，建立 `aidsu/windows-mvp` 分支和 `upstream`。已建立独立线上 [Fork](https://github.com/qianrrf/ai-dev-studio-ultimate)，配置 `origin`，并推送开发分支。
2. 核查实际源码、数据库迁移、原生 Agent 驱动、Windows 运行与构建流程，以及当前官方文档。架构图、模块归属、复用项、必需扩展点与风险见 [源码审计](source-audit.md)。
3. 建立隔离的开发数据、运行发现文件、工具缓存和一次性测试仓库；没有连接原版应用数据。
4. 完成后端构建、vet、前端类型检查与原版全量测试执行，记录实际失败；没有把存在失败的测试集标成通过。
5. 修复自定义 Codex 中转站认证误判。通过原生 `account/read` 判断是否要求 OpenAI 认证，保留你现有的中转站和凭据；真实 Codex 主控调用已成功。
6. 使用已有 DSH ACP 适配器，在两个独立 Git Worktree 中完成真实并行任务。选用本机已有 DeepSeek 账户的原生模型配置，没有新建 Bridge 或 Adapter。
7. 验证 DSH 工具审批拒绝、执行中取消、等待审批时 Daemon 重启，以及受控退出后恢复。各项证据、会话和 turn ID 见 [验证记录](validation.md)。
8. 新增持久化 `readOnly` 会话和 CLI 入口。限制 Codex Chat，阻止设置/模式/引擎切换绕过，恢复后保留约束；关闭只读会话的外部工具并在派发前核查真实工具列表。
9. 新增不消耗模型额度的 Windows 沙箱检查脚本，便于后续复测。

## 当前真实能力

| 能力 | 结论 |
|---|---|
| Codex 使用现有中转站启动主控 | 已真实验证 |
| DSH 原生 ACP 启动、输出与结果读取 | 已真实验证 |
| 两个 DSH Worker 同时运行、独立 Worktree | 已真实验证 |
| 未知工具审批与拒绝 | 已真实验证，未静默批准 |
| 取消与受控恢复 | 已验证部分场景；未覆盖全部强杀/后代进程场景 |
| Codex 无 PR 只读审查 | 原生文件只读探针通过；完整独立审查闭环仍待验证 |
| GitHub 原生 PR Reviewer | 未验证；GitHub 登录及 Fork 已完成，尚未发布测试 PR 或评论 |
| 内外主控切换、旧主控派单拦截 | 持久控制权、事务内派单拦截与入口测试正在实施；尚未完成全流程验收 |
| 完整派工—审查—返工—验收闭环 | 尚未完成 |
| Windows 安装包、品牌与共存 | 尚未完成 |

## 阻塞与处理顺序

### 1. Windows 原生只读审查环境

本机 Codex 0.162.1 的 `elevated` 沙箱初始化失败，日志指向正在占用的 Codex 运行时文件及权限设置。2026-11-04 14:02 重启后复测，问题仍然存在；移除测试子进程继承的 `CODEX_MCP_NODE_PATH` 也没有解决，不能继续把重启当作已证实的修复方法。

最新确认的首版边界为可信本机项目：要求原生文件只读，但不承诺隔离恶意进程或阻断全部 loopback。产品现在允许明确指定的原生 `unelevated` 配置。2026-11-04 20:11，稳定 Codex 0.162.1 的无模型探针读取成功，实际写文件触发 `UnauthorizedAccessException`，文件不存在；loopback 请求成功，单独记录为已知限制。探针捕获明确的文件权限异常后正常返回，不把 RPC 或初始化失败当作拒写证据。

另外诊断了桌面附带的 Codex `0.166.0-alpha.5.1`，没有替换产品默认版本。它能读取文件，写入探针返回错误且没有生成文件，但本机 HTTP 探针仍收到请求，因此整体结果仍为失败。RPC 错误本身不作为成功拒写的证据。

本机 Windows 11 23H2 build 22631 不满足当前官方文档列出的 MXC 主机版本要求；官方还说明 MXC 允许主机 loopback，不能假设升级 Windows 就能满足本项目的网络边界。

没有结束你正在使用的 Codex 桌面进程，也没有修改全局 Codex 配置。用测试进程的环境变量明确选择原生配置后运行：

```powershell
$env:AIDSU_CODEX_WINDOWS_SANDBOX = 'unelevated'
node scripts/ultimate/probe-codex-sandbox.mjs
```

读取成功且实际写入被权限拒绝，是继续真实审查验收的前提；网络观察独立报告。检查脚本不会调用模型。完整审查、重启恢复及固定提交审查仍须通过真实联调。

### 2. Windows 原版桌面构建工具

原版桌面构建已走到 Electron Forge 的 `better-sqlite3` 原生模块编译。使用本机附带 Python 后，确认缺少 Visual Studio C++ 构建工具。微软官方 Build Tools 17.14.44 已启动安装，目标目录为 `C:\BuildTools\AI-Dev-Studio`，没有自动重启。当前仍在下载组件，尚未获得完整工具链安装成功的证据；不能将启动安装当作构建缺口已解决。

安装完成后，在原版复现 Worktree 重跑正常构建。已核对 `better-sqlite3` 官方发布资产，没有当前 Electron ABI 149 的 Windows x64 预编译包，不能用其他 ABI 的二进制替代编译。未通过此步骤前，没有可交付的 Windows 安装包。安装器与组件校验细节见 [验证记录](validation.md#windows-environment-follow-up)。

同时保留上游依赖安装的审计告警和原版 Windows 测试失败记录，后续逐项归因，不使用自动升级掩盖风险。

### 3. GitHub Fork 与 PR 验证

GitHub 登录已完成，独立 Fork 为 <https://github.com/qianrrf/ai-dev-studio-ultimate>，`origin` 指向该 Fork，`upstream` 保持原项目。开发分支的前四个开发提交已经推送并核对远端 SHA。

线上 Fork 不等于原生 PR Reviewer 验收通过；尚未发布测试 PR、评论或 Release，也没有远端 CI 通过的证据。真实 PR 审查应在一次性测试仓库中执行，并单独记录其发布行为与审查结果。

## 后续开发顺序

1. 解决以上环境缺口，补齐阶段 A/B 验收；记录版本与实测结果。
2. 在 Daemon 实现统一任务入口和项目控制权修订号检查，覆盖 UI、CLI 和自动入口；切换主控不终止 Worker，旧主控请求必须被拒绝。
3. 接入独立本地审查，绑定 Worker 和 base/head 提交，提交变化后使旧结论失效；保留测试、日志、审查和主控验收证据，最多两轮业务返工。
4. 两个 Codex 入口分别完成闭环和故障测试后，再修改品牌、安装身份、独立状态目录、更新源和遥测配置。
5. 验证安装、卸载、与原版共存，再交付安装包。原版的名称、更新源和数据目录仍存在于当前代码，不能把此分支作为已经独立品牌的发布版使用。

## 保留物与回退

- 源码：当前仓库，分支 `aidsu/windows-mvp`；通过 `git log --oneline b325d24b741297bf89296e76501d7e09018c4a04..HEAD` 查看本轮提交。
- 机器本地证据：`%USERPROFILE%\.cache\ai-dev-studio-ultimate\validation`。
- 隔离测试数据：`%USERPROFILE%\.cache\ai-dev-studio-ultimate\lab`。
- 原版复现 Worktree：证据目录下 `baseline-repro`，具有独立的真实依赖安装。
- 一次性项目和 Worker Worktree 已保留，包括尚未提交的测试修复；本任务的 Agent 宿主和隔离 Daemon 已退出。

回退时优先建立固定基线的新 Worktree 对照，不对现有工作区执行破坏性重置。只读会话迁移已经用于隔离数据库；旧程序使用前应备份该隔离数据并检查迁移兼容性。原版用户数据没有迁移或回滚。

Apache-2.0 许可证和原项目版权、第三方声明均保留。

可回退提交：`20437010f` 为环境与基线记录，`d244fa197` 为中转站认证修复，`46db31e8f` 为持久只读约束基础。后续文档提交记录最终验证结果。
