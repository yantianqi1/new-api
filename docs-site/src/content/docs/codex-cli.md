---
title: Codex CLI
---

Codex CLI 是 OpenAI 官方的终端编程助手。只需一个 config.toml，即可让它经 waner~ API站 中转。

💡推荐：使用 CC Switch 一键导入跳过手动配置——它会帮你写好 Codex 配置，还能一键在多个站点之间切换。

### 前提条件

已安装 Node.js 18+

已在本站创建一个 API key

### 第 1 步：安装 Codex CLI

```bash
npm install -g @openai/codex
```

### 第 2 步：写入配置

编辑（或创建）~/.codex/`config.toml`：

Windows：按 Win + R，输入 %userprofile%\.codex

macOS / Linux：mkdir -p ~/.codex

```toml
#:schema https://developers.openai.com/codex/config-schema.json
model_provider = "custom"
model = "gpt-5.6"
model_reasoning_effort = "high"

[model_providers.custom]
name = "waner~ API站"
base_url = "https://api.aiwanai.cc/v1"
wire_api = "responses"
requires_openai_auth = true
```

与 Claude Code 不同，Codex 的 `base_url` 必须包含 `/v1` 后缀，且 `wire_api` 使用 `"responses"`（目前唯一支持的取值）。第一行 schema 让支持的编辑器能识别无效的 TOML 键。

`gpt-5.6` 同时收录于 OpenAI 官方 Codex 文档与本站当前模型列表。本站可能还提供带后缀的路由模型；这些是站点专用的名称，不是 OpenAI 官方别名。

使用前，请先确认它出现在 Pricing 页面，且已在你的 key 分组中启用。

然后在 ~/.codex/`auth.json` 中保存密钥：

```json
{
  "OPENAI_API_KEY": "sk-your-api-key"
}
```

### 第 3 步：验证

重启终端，在任意项目中运行：

```bash
codex
```

能正常对话即表示连接成功；调用记录会出现在本站的 Usage Logs 中。

### 可选：启用网络搜索

近期的 Codex 版本将网络搜索开关放在 `config.toml` 的顶层 `web_search` 键上（不在 [model_providers.custom] 内）：

```toml
web_search = "live"
```

启用后，Codex 会通过本站的 POST /v1/alpha/search 端点执行搜索。注意事项：

搜索由支持该端点的站点渠道提供；如果你的 key 所在分组还没有这类渠道，搜索会失败，但普通对话不受影响。

搜索按次计费（`web_search` 工具价格，默认 $0.01/次），并在 Usage Logs 中显示为独立条目；失败的搜索不扣费。

稳定性：某个渠道搜索失败（过载、超时……）时，站点会自动改用其他渠道重试——无需人工干预。真实的联网搜索可能需要数十秒，这属于正常现象。

不需要时，删除这一行，或将 `web_search` 设为 "disabled"。

如果你通过 CC Switch 管理 Codex，`ccswitch://import` 协议无法携带这个键：本站导入对话框提供了该行的复制按钮——请把它粘贴到导入后的 provider 配置的顶层。

如需全局生效，把它放进 CC Switch 的 Codex 通用配置片段，并在每个 provider 上启用「Apply Common Config」。

### 故障排查

401 / Unauthorized：auth.json 里的 key 是否完整？（它不应放在 `config.toml` 里。）

404 / model not found：模型名是否存在于 Pricing 页面，且在你的 key 分组中？

流式报错 / 超时：系统代理常有干扰——试试为终端禁用代理。

### 可信来源与自查

本站实时模型列表与分组：站点模型名和 key 分组的权威来源。

Codex 配置基础：官方 ~/.codex/config.toml 的位置与基础配置。

Codex 高级配置：custom-provider、`base_url` 与 wire_api 的示例。

Codex 完整配置参考：当前支持的每个 TOML 键与取值。升级 Codex 后请对照它和 schema 检查。
