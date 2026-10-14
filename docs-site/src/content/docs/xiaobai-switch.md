---
title: XiaoBaiSwitch 一键导入
---

XiaoBaiSwitch 是一款开源桌面工具，以“上游站点”为核心管理 Claude Code、Codex 和 Pi 的配置：一个站点承载多条线路，支持测速、一键切换，并在应用前自动备份。

waner~ API站 集成了它的 `xiaobaiswitch://import` 协议——站点、所有端点以及 Codex 能力预设都通过一条链接传递，无需手动编辑配置。

### 第 1 步：下载并安装

获取对应平台的安装包（Windows x64 / ARM64，macOS Apple Silicon / Intel）：

GitHub Releases： https://github.com/Licoy/xiaobai-switch/releases/latest

安装后请启动一次，以便系统注册 `xiaobaiswitch://` 协议。

请使用 v0.0.6 或更高版本（链接导入自 v0.0.3 起可用，Pi Coding Agent 支持于 v0.0.6 引入）。

macOS 安装包为临时签名（ad-hoc），系统可能提示应用“已损坏”。请将应用拖入 Applications，在终端中执行 `xattr -cr /Applications/XiaoBaiSwitch.app`，然后右键点击 → 打开。

Windows 上若首次运行时系统提示，请安装 Web View2。

### 第 2 步：一键导入

在 **控制台 → API Keys** 中，打开任意密钥的操作菜单，选择 **Import to XiaoBaiSwitch**。浏览器会打开应用并弹出确认对话框；核对站点名称、线路和密钥前缀后确认，本站点即被写入为一个上游站点。

与 CC Switch 按工具逐个导入不同，XiaoBaiSwitch 导入的是整个站点：本站点的所有 API 端点一次性写入（第一个为默认线路，可在应用内测速和切换）。模型列表不随链接传递——应用会从本站点拉取实时列表，因此始终保持最新。

导入不会自动应用到 Claude Code / Codex / Pi。导入后，请打开应用中心，为本站点和目标工具选择模型并应用；原配置会自动备份，可随时恢复。

导入对话框中的“Codex 能力预设”（网页搜索 / 图像理解 / 图像生成 / 远程压缩）存储在站点上，并在应用到 Codex 时作为默认值提供。

本站点支持网页搜索和图像理解，默认勾选；图像生成需要包含图像模型的密钥分组，远程压缩取决于上游支持，因此这两项默认不勾选——可按需开启。

### 应用时按目标工具选择模型

Codex：优先选择 gpt-5.6；使用本站点的带后缀模型前，请先对照 Pricing 页面和你的密钥分组确认。

Claude Code：追求最强能力选 claude-fable-5，日常均衡使用选 claude-sonnet-5。

Pi（需要 XiaoBaiSwitch v0.0.6+）：本站点以 OpenAI 兼容协议导入，因此应用到 Pi 时会写入 `openai-completions` 条目（`/v1` URL 自动生成），并可选择把本站点的完整模型列表复制到 Pi 的目录中。

`gpt-5.6` 等 OpenAI 系模型可直接使用。如果你主要在 Pi 中运行 `claude-fable-5` / `claude-sonnet-5`，请先在站点编辑器中把本站点的协议切换为 Anthropic——随后 Pi 会被写入为 `anthropic-messages` + 站点根地址，以保持 Claude 的 prompt caching 生效（该协议切换不影响 Claude Code / Codex 的应用方式）。

### 从密钥页面导入

控制台 → API Keys 中每个密钥的操作菜单都有相同的“XiaoBaiSwitch”条目。侧边栏的 Third-party → Xiao Bai Switch 条目可随时打开同样的导入。

### 问题排查

点击后没有任何反应：确保已安装 XiaoBaiSwitch 且至少启动过一次，并允许浏览器的“Open XiaoBaiSwitch?”提示。

重复导入会怎样：协议和线路集合相同（与顺序无关）的站点会被视为同一站点——密钥相同则复用，密钥不同则更新；增删线路则会创建新站点，而不会合并。

导入后工具仍无法连接：先在应用中心确认本站点已应用到该工具；然后在本站点的 Pricing 页面上，对照你的密钥分组核对模型名称。对于 Claude Code 之外的工具，重启 CLI 通常即可应用更改。

手动共享配置：手工构造的 `xiaobaiswitch://sites?...` 链接同样可用，参数说明参见 README 的链接导入部分。

安全：导入链接中包含你的 API 密钥。请勿复制、截图或公开分享完整链接；仅从本站点或其他你信任的来源导入。

### 可信来源与自查

本站点的实时模型列表与分组：应用前确认模型名称和密钥分组。

XiaoBaiSwitch 官方 README：功能、安装包和 `xiaobaiswitch://` 协议参数。

XiaoBaiSwitch Releases：各版本的安装包和发布说明。
