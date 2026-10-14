---
title: Claude Code
---

Claude Code 是 Anthropic 官方的终端编程助手。只需两个环境变量，就能让它走 waner~ API站 的通道。

💡推荐：用 CC Switch 一键导入跳过手动配置——它会自动帮你写好 Claude Code 的配置，并支持在不同站点间一键切换。

### 前提条件

已安装 Node.js 18+

在本站创建好 API 密钥（控制台 → API Keys）

### 第 1 步：安装 Claude Code

根据你的系统，在官方 Claude Code 安装指南中选择合适的安装方式（macOS / Linux 推荐使用官方安装器；Windows 推荐 WinGet 或 PowerShell 安装器）。

`npm install -g @anthropic-ai/claude-code` 仍然可用，但 Anthropic 已将 npm 安装方式标记为弃用。仅在你确实需要 npm 时才使用它。

### 第 2 步：写入配置

编辑（或创建）`~/.claude/settings.json`：

Windows：按 Win + R，输入 `%userprofile%\.claude`，在该目录下创建 settings.json

macOS / Linux：执行 `mkdir -p ~/.claude` 并打开该目录

```json
{
  "env": {
    "ANTHROPIC_BASE_URL": "https://api.aiwanai.cc",
    "ANTHROPIC_AUTH_TOKEN": "sk-your-api-key"
  }
}
```

注意：ANTHROPIC_BASE_URL 是站点根地址——不要在末尾追加 /v1。

如需固定模型（可选），扩展 env 配置即可。Claude Code 通过模型别名和 Default 选项来选择不同层级，下列变量决定每个别名实际解析到哪个模型：

```json
{
  "env": {
    "ANTHROPIC_BASE_URL": "https://api.aiwanai.cc",
    "ANTHROPIC_AUTH_TOKEN": "sk-your-api-key",
    "ANTHROPIC_MODEL": "claude-fable-5",
    "ANTHROPIC_DEFAULT_OPUS_MODEL": "claude-opus-4-8",
    "ANTHROPIC_DEFAULT_SONNET_MODEL": "claude-sonnet-5",
    "ANTHROPIC_DEFAULT_HAIKU_MODEL": "claude-haiku-4-5"
  }
}
```

ANTHROPIC_MODEL：主对话模型。claude-fable-5 是当前最强的旗舰模型；日常使用想更便宜，可换成 claude-sonnet-5。

ANTHROPIC_DEFAULT_OPUS_MODEL / ANTHROPIC_DEFAULT_SONNET_MODEL / ANTHROPIC_DEFAULT_HAIKU_MODEL：控制 opus、sonnet、haiku 别名——以及 Default 选项——分别解析到哪个模型。

`ANTHROPIC_DEFAULT_HAIKU_MODEL` 还承担会话标题生成等后台任务。请将其设为一个可用的 Haiku 模型。

`ANTHROPIC_SMALL_FAST_MODEL` 已弃用，不应再使用。

上述模型名仅为示例；请严格使用本站 Pricing 页面上列出的模型名称。

### 进阶：在 Claude Code 中运行 grok-4.5

本站会自动把 Claude Code 的请求转换为 grok-4.5 能识别的格式，因此你只需把主模型指向它即可：

```json
{
  "env": {
    "ANTHROPIC_BASE_URL": "https://api.aiwanai.cc",
    "ANTHROPIC_AUTH_TOKEN": "sk-your-api-key",
    "ANTHROPIC_MODEL": "grok-4.5"
  }
}
```

`grok-4.5` 是 xAI 的官方模型，已包含在本站当前模型列表中。xAI 支持 prompt caching，但实际的缓存命中与成本取决于请求前缀、路由以及本网关的实现。

在把它当作日常模型之前，请先在本站的日志中查看缓存使用情况。使用 CC Switch 一键导入时选择 `grok-4.5` 作为主模型，应用的是同样的配置。

### 第 3 步：验证

在任意项目中执行：

```bash
claude
```

随便问它点什么——能收到回复就说明接入成功。

在 Claude Code 中运行 `/status` 可查看 Anthropic base URL 和当前使用的凭证，并到本站的 Usage Logs 核对调用记录。如果你用的是 shell 环境变量而非 `settings.json`，请先打开一个新的终端窗口再测试。

### 故障排查

连接错误 / fetch failed：`ANTHROPIC_BASE_URL` 是否确为 `https://api.aiwanai.cc`（ 末 尾 没 有 `/v1`）？用 `/status` 确认当前会话已加载该地址。

401：密钥是否完整（包含 `sk-` 前缀）且仍然有效？

模型不可用：你的密钥所在分组是否包含该模型？参见错误码与故障排查。

### 可信来源与自查

本站的实时模型列表与分组：可用模型名称和密钥分组以此为准。

Claude Code 网关连接：关于路径、ANTHROPIC_BASE_URL、凭证与 /status 的官方说明。

Claude Code 模型配置：关于模型别名、ANTHROPIC_DEFAULT_*_MODEL 以及已弃用变量的官方说明。

xAI：grok-4.5 与 prompt caching。
