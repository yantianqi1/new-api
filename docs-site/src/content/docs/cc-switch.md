---
title: CC Switch 一键导入
---

CC Switch 是一款开源、跨平台的桌面工具，用于管理 Claude Code、Codex、Gemini CLI、OpenCode、OpenClaw、Hermes、Pi 的供应商配置，并支持一键在不同站点之间切换。

waner~ API站 集成了它的 `ccswitch://import` 协议，无需手动编辑配置。

### 第 1 步：下载并安装

获取对应平台的安装包（Windows / macOS / Linux）：

GitHub Releases： https://github.com/farion1231/cc-switch/releases/latest

安装后请启动一次，以便系统注册 `ccswitch://` 协议。

请使用 v3.20.3 或更高版本。本站点的导出配置已于 2026-09-12 对照该版本及相同主提交验证过。

### 第 2 步：一键导入

在 **控制台 → API Keys** 中，打开任意密钥的操作菜单，选择 **Import to CC Switch**。

浏览器会打开 CC Switch 并弹出确认对话框；确认后，本站点会作为供应商写入工具配置并自动激活。

一键导入支持 Codex、Claude Code、Gemini CLI、OpenCode、OpenClaw、Hermes 和 Grok Build（需要 CC Switch v3.18+；旧版本会拒绝该 app id）。

如果本站点发布了多个 API 端点，你也可以选择同时导入备用端点，并在 CC Switch 中切换。

“Include balance query”默认勾选：导入的供应商会携带用量查询脚本，CC Switch 会同时显示可用余额和累计用量。有限额度的密钥显示其自身的数值；无限额度的密钥显示所属账户的数值。

该脚本使用密钥本身调用本站点的 `/api/usage/token/`，不包含任何账户级凭据。

CC Switch 本身也支持 Claude Desktop，但其当前官方的供应商 Deep Link 协议不接受 claude-desktop。请在 CC Switch 中手动添加 Claude Desktop 供应商，而不是使用本站点的一键导入。

Pi（Pi Coding Agent）配置导出：在本站点的导入对话框中选择 Pi，选定 API 格式、端点和模型，然后点击 Copy Pi configuration。

在 CC Switch 中进入 Pi → Add Provider → Custom Configuration，输入名称和唯一的供应商 key，把复制的对象粘贴到 Configuration JSON 中并保存。

导出内容是一个包含端点、API 密钥和模型的单一供应商对象，CC Switch 会把它写入 Pi 的 `models.json`。

官方 v3.20.3 的 deep-link 协议仍拒绝 `app=pi`，因此该入口采用复制 JSON 的方式。

对于 Pi 中的 Claude 模型，选择 Anthropic Messages，导出会自动使用站点根地址。

对于 OpenAI 格式，选择 OpenAI Chat Completions 或 OpenAI Responses，导出会自动补上 `/v1`。请对照 Pricing 页面和你的密钥分组核对模型名称。

Pi 导出仅包含所选端点，不包含 CC Switch 的备用端点或用量脚本。你可以在 CC Switch 中根据模型能力设置上下文上限、输出上限和推理选项。

导入器会自动处理 URL 后缀：Claude Code 和 Gemini CLI 使用站点根地址；Codex、Grok Build、OpenCode、OpenClaw、Hermes 使用 `/v1` URL；Pi 遵循所选的 API 格式。

请勿手动添加或移除 `/v1`。

### 导入时按目标应用选择模型

Codex：优先选择 `gpt-5.6`；使用本站点的带后缀模型前，请先对照 Pricing 页面和你的密钥分组确认。

如需网页搜索，请在导入的供应商配置顶层添加 `web_search` = `"live"` —— 导入对话框提供了该行的复制按钮（Deep Link 协议无法携带此键）。

若要将其应用到所有供应商，可放入 CC Switch 的 Codex 通用配置片段，并在每个供应商上启用“Apply Common Config”，参见 Codex CLI 指南。

Claude Code：追求最强能力选 `claude-fable-5`，日常均衡使用选 `claude-sonnet-5`，Haiku 槽位选 `claude-haiku-4-5`。

`grok-4.5` 也可用；在依赖它进行长会话之前，请先在日志中核实实际的缓存用量。

Grok Build：选择 grok-4.5。

Claude Code 导入会写入主模型以及 Opus / Sonnet / Haiku / Fable 槽位。Fable 和 Subagent Model 通过配置负载以 `ANTHROPIC_DEFAULT_FABLE_MODEL` 和 `CLAUDE_CODE_SUBAGENT_MODEL` 传递；对于旧版 Claude Code，还会根据 Haiku 槽位设置 `ANTHROPIC_SMALL_FAST_MODEL`。

如果上游拒绝 Artifact 工具的 schema，选择 Disable Artifact tool 以导出 `CLAUDE_CODE_DISABLE_ARTIFACT=1`。该选项默认关闭，且不影响其他工具的可用性。

仅当你需要给 Claude Code 分配一个只暴露 OpenAI 格式端点的模型时，才在 CC Switch 中启用 Claude 本地代理/接管。对于本指南中的直连方式，本站点的 Anthropic Messages 端点不需要它。

### 从密钥页面导入

控制台 → API Keys 中每个密钥的操作菜单都有相同的“Import to CC Switch”条目。

### 问题排查

点击后没有任何反应：确保已安装 CC Switch 且至少启动过一次，并允许浏览器的“Open CC Switch?”提示。

导入失败或提示应用无效：更新 CC Switch 后重试。Claude Desktop 请在应用内手动添加；Pi 请使用本站点对话框中的 Copy Pi configuration（见上文）。

导入后工具仍无法连接：打开 CC Switch 检查供应商的端点、密钥和模型名称，然后在本站点的 Pricing 页面确认该模型已对你的密钥分组启用。

对于 Claude Code 之外的工具，重启 CLI 通常即可应用更改。

手动共享配置：CC Switch 也接受手工构造的 `ccswitch://v1/import?...` 链接，参见其官方文档。

安全：导入链接和 Pi 配置 JSON 中包含你的 API 密钥。请勿截图或公开分享，仅从本站点或其他你信任的来源导入。

### 可信来源与自查

本站点的实时模型列表与分组：导入前确认模型名称和密钥分组。

CC Switch 官方 README：支持的应用、安装包和切换行为。

CC Switch Deep Link 协议：`ccswitch://` 参数与手动导入格式。

CC Switch v3.20.3 发布说明：当前功能、升级说明和修复。

已审计的导入器源码：Pi deep-link 限制、Claude 环境负载和 Codex 配置行为。
