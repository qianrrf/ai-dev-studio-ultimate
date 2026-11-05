# AI Dev Studio Ultimate 0.1.0-alpha.3 使用说明

这是供个人使用的未签名 Windows x64 候选测试版，尚未通过完整验收。Linux、公开发布和自动更新不在本轮范围内。具体状态见 [交付记录](delivery.zh-CN.md) 和 [alpha 3 验收记录](alpha3-acceptance.md)。

## 安装与启动

1. 对照随包 `SHA256SUMS.txt` 校验安装器，再运行 `AI Dev Studio Ultimate Setup 0.1.0-alpha.3.exe`。未签名是本轮已约定的交付方式；不要把构建日志中的 signing 字样当作数字签名通过。
2. 通过开始菜单或安装目录的 `ai-dev-studio-ultimate.exe` 启动。默认后台端口为 31338；占用时会显示启动失败及绑定错误，不自动换端口。
3. 打开一个可信本机 Git 仓库，例如 `D:\Projects\项目名`。首版每项目一个仓库；仓库、Worker、Reviewer 和 Git 管理目录的实际物理路径均须不超过 200 个 UTF-16 单位，目录链接不能绕过检查。中文、空格及仓库内深层文件继续支持。不要导入产品数据目录或原版 `.ao` 目录。
4. 保持 Codex 原有中转站配置和 DSH 原有账户。产品不会改用官方 OAuth，也不保存或复制这些账户凭据到交付目录。
5. 初次真实使用会消耗模型账户额度。本轮开发联调累计登记 52/60 回合；DSH 当前返回余额不足，因此双入口完整闭环仍未验收。不会自动充值、切换账户或超额测试。

默认数据根目录是 `%USERPROFILE%\.ai-dev-studio-ultimate`，与原版 `%USERPROFILE%\.ao` 分开。里面包括后台数据库、运行发现文件、Worktree、日志和 Electron 配置。不要删除有未保存修改的 Worktree。

## Codex 与 DSH

产品沿用原生 Codex app-server 和 DSH Chat/ACP。已测版本为 Codex 0.162.1、DSH 0.2.1-rc.5；不代表所有后续版本自动兼容。Git 必须可用；需要 PR 功能时还需要 GitHub CLI 登录。

缺少 Git 时，从 [Git for Windows 官方网站](https://gitforwindows.org/) 安装后退出并重开产品。桌面会验证 PATH 和标准安装位置中的 Git 程序，不依赖 Codex 私有运行目录。缺少依赖时按界面提示修复，不要关闭依赖检查。

本机 DSH 启动程序位于 `D:\AI\deepseekh\resources\runtime\cli\bin\dsh.cmd`。如果桌面找不到它，可在 PowerShell 中仅为本次启动设置路径：

```powershell
$env:PATH = 'D:\AI\deepseekh\resources\runtime\cli\bin;' + $env:PATH
& '你的安装目录\ai-dev-studio-ultimate.exe'
```

不要把真实账户密钥写入项目、脚本或报告。中转站配置沿用现有 Codex 配置，不需要填写新的官方登录信息。Windows 原生只读审查使用已验证的 `unelevated` 配置；它可以拒绝原生文件写入，但不阻断全部本机 loopback 访问。

## 产品内主控

在项目设置的 Codex 主控区域选择产品内 Codex。该动作创建 Codex Chat 主控会话，会涉及真实账户调用；只打开设置不会启动模型。新需求进入项目收件记录，由当前主控读取、拆分和派发。默认容量为两个 DSH Worker，审查使用独立 Codex 只读会话。

首次使用应确认 Worker 为 `deepseek-harness`、模式为 Chat，Orchestrator 和 Reviewer 均为 Codex，保留工具审批。不要将全自动批准作为解决权限兼容问题的方法。

Codex 默认权限显示为 **Manual approvals**，原生策略为 `untrusted` 与 `workspace-write`。只有明确选择 **Bypass permissions** 才使用全权限。升级保留已有配置，不会替你撤销曾保存的 bypass；使用前检查项目和会话记忆的权限。只读审查不能改成可写会话。

## 外部 Codex、接管与释放

使用安装目录内 `resources\daemon\ao.exe`，不用覆盖全局 `ao`。先启动桌面应用。默认运行发现路径与产品数据目录一致；隔离启动时，CLI 必须使用相同的 `AO_DATA_DIR` 和 `AO_RUN_FILE`。

下面示例中的仓库路径、项目名和主控标识应替换为实际值：

```powershell
$studioCli = '你的安装目录\resources\daemon\ao.exe'
& $studioCli project add --path 'D:\Projects\demo' --id demo --name demo --orchestrator-agent codex --worker-agent deepseek-harness
& $studioCli project set-config demo --reviewer codex
& $studioCli project control demo
```

读取当前修订号后，再接管。例如查询得到修订号 0 时：

```powershell
& $studioCli project control demo --owner external-codex --entry external --revision 0
& $studioCli project request demo '实现需求并附测试证据' --id requirement-001
& $studioCli project inbox demo
```

接管成功会返回新的 `revision`。每次派工、返工和验收都需要当前主控标识、当前修订号和固定请求 ID。旧修订返回冲突时先查询，不要改用新请求 ID 反复派工。主控切换不会结束已有 Worker；新主控应先读收件、会话与交付证据。

释放时再次查询最新修订号，然后传入查询结果：

```powershell
& $studioCli project control demo
& $studioCli project control demo --release --revision 1
```

这里的 `1` 只是示例，不能假设它始终有效。外部 Codex 离线不会自动启动第二个主控。外部主控仍需明确按本产品 CLI/HTTP 协议操作；alpha 2 的合成闭环是历史证据，不能代替当前版本复测。当前候选的未验收项见 [alpha 3 验收记录](alpha3-acceptance.md)。

## 测试、独立审查与验收

Worker 必须提交待审代码。主控使用 `session delivery` 保存真正执行的测试命令及退出码，再创建固定提交的独立只读审查，提取结论，必要时登记返工，最后验收。完整命令与各字段含义见 [交付证据接口](delivery-evidence.md)。

审查后的提交、目录或未提交修改发生变化时，旧结论不能验收。最多两轮业务返工，仍不合格时交给用户。执行结果不明时先查询持久记录和日志；不要盲目重新运行。取消与审批拒绝始终可用。

## 故障排查与回退

- 后台启动失败：查看界面详细错误和产品目录下日志。端口占用时退出占用程序或使用显式隔离端口；不要将产品连接到原版后台。
- DSH 不可用：先检查上述 `.cmd` 的版本及进程 PATH；`Insufficient Balance` 是账户服务返回的余额不足，反复派单不能修复。不要安装第二套 Adapter 或 Bridge。
- 中转站不被识别：保持原配置，核对已测 Codex 版本；不要为了消除提示覆盖现有认证文件。
- 数据库被占用：先退出使用同一数据目录的产品实例，保留原数据库；不要删除数据库或强行清空锁文件。
- 长路径：遇到 `WINDOWS_GIT_PATH_TOO_LONG` 时，将项目放在较短位置，并为隔离测试使用较短的 `AO_DATA_DIR`。产品会在 Git 或 Agent 副作用前拒绝超长根目录；已有项目的记录、取消、审批拒绝和未提交文件保留，不会自动搬迁或删除。UTF-16 单位不是字节数，部分字符占两个单位。
- 回退：退出产品，备份整个产品数据根目录和外部项目。只恢复与对应程序版本兼容的数据库备份，不直接让旧程序打开新迁移后的数据库。原版数据未迁移，不需要回滚 `.ao`。

自动更新、云执行、上游遥测出口和错误报告出口关闭。本轮没有发布公开 Release、上传账户凭据或配置商业签名。卸载后应保留项目与数据备份，不手工清除用户 Worktree。
