---
title: DeepSeek + OpenCode / ZCode
---

要用 DeepSeek 的中文模型（DS）做 Agent 编程，推荐 OpenCode 或 ZCode。

OpenCode 适合终端优先、希望配置透明的用户；ZCode 是桌面应用，提供更直观的项目与模型管理。

两款工具都必须通过 OpenAI 兼容的自定义供应商、使用 Responses API 接入本站。不要选择内置的 DeepSeek 供应商：它通常指向 DeepSeek 官方 API，而不是 waner~ API站。

下载：OpenCode 官方下载 · ZCode 官方下载（支持 Windows、macOS 与 Linux）

### 连接信息

| 设置项 | 取值 |
| --- | --- |
| 供应商类型 | Custom / OpenAI Compatible |
| API Base URL | `https://api.aiwanai.cc/v1` |
| API key | 在本站 控制台 → API Keys 创建的密钥 |
| 推荐端点 | `POST https://api.aiwanai.cc/v1/responses` |
| 推荐模型 | `deepseek-v4-pro` |
| 更低成本的模型 | `deepseek-v4-flash` |

模型 ID 区分大小写。当前 ID 是 `deepseek-v4-pro`，不是 `DeepSeek-V4-Pro`。

模型与分组可能调整，配置客户端前请先查看线上的 Pricing 页面。

### 为什么推荐 `/v1/responses`

我们验证了网关路由及其 DeepSeek 适配器：本站接受 POST `/v1/responses`，并通过对应的上游 DeepSeek 端点转发 Responses 请求。

OpenCode 经由 `@ai-sdk/openai` 使用它，ZCode 的自定义供应商设置则明确提供 Responses (`/responses`) API 格式。下面两种配置方式都使用 Responses API。

Responses API 更适合多轮 Agent 上下文与工具调用。本网关还能接收并透传缓存相关字段（如 `prompt_cache_key`），让长会话更有可能一致地复用缓存上下文，避免重复计算相同前缀。

但这并不保证每次请求都能命中缓存：请保持共享的系统指令与工具定义稳定，并使用一致的模型与路由。

重要提示：在任一客户端中，Base URL 只填 `https://api.aiwanai.cc/v1`。不要填 `https://api.aiwanai.cc/v1/responses`——客户端会自行拼 接 端 点 路 径，否则可能拼出无效 URL。

### 开始之前

登录本站，确认钱包余额充足。

在 控制台 → API Keys 下创建一个专用密钥。

选择支持 Pricing 页面所示 DS 模型的密钥分组。deepseek-v4-pro 目前出现在 domestic-model 及部分阿里云 domestic-model 分组中。

如果密钥配置了模型白名单，请把 deepseek-v4-pro 以及你打算使用的其他 DS 模型都加进去。

每个工具使用单独的密钥，这样用量与吊销互相独立。绝不要把密钥提交到 Git。

### 方式一：OpenCode（终端用户推荐）

OpenCode 是一款开源编程 Agent，提供终端、桌面应用与 IDE 扩展三种形态。

### 从零安装

在官方下载页面选择一种安装方式。如果已安装 Node.js，可以运行：

```bash
npm install -g opencode-ai
```

官方文档在 Windows 上推荐使用 WSL；npm、Scoop 与 Chocolatey 同样支持。

用 `opencode --version` 验证安装是否成功。

### 升级已有安装

升级到当前版本：

```bash
opencode upgrade
```

不必依赖某个硬性要求的最低版本。如果你的安装渠道不支持 upgrade，就通过该渠道更新。

把下面的供应商配置合并进现有配置，不要覆盖无关设置。

### 1. 配置供应商

若要让所有项目都能使用，编辑全局配置：

Windows：%USERPROFILE%\.config\opencode\`opencode.json`

macOS / Linux / WSL：~/.config/opencode/`opencode.json`

如果只想在单个项目中使用，就把 opencode.json 放在项目根目录下。

```json
{
  "$schema": "https://opencode.ai/config.json",
  "model": "aiwanai/deepseek-v4-pro",
  "provider": {
    "aiwanai": {
      "npm": "@ai-sdk/openai",
      "name": "waner~ API站 · DeepSeek",
      "options": {
        "baseURL": "https://api.aiwanai.cc/v1"
      },
      "models": {
        "deepseek-v4-pro": { "name": "DeepSeek V4 Pro" },
        "deepseek-v4-flash": { "name": "DeepSeek V4 Flash" }
      }
    }
  }
}
```

这里请使用 `@ai-sdk/openai`。OpenCode 会调用 `https://api.aiwanai.cc/v1` 下的 `/responses`。

不要换成 `@ai-sdk/openai-compatible`，它面向的是 `/v1/chat/completions`；只有当旧版 OpenCode 无法使用 Responses 时，才把它当作兼容降级方案。

### 2. 保存 API 密钥

在项目中运行 opencode，进入 /connect，然后：

选择 Other，不要选内置的 DeepSeek 供应商。

输入 aiwanai 作为供应商 ID，必须与配置完全一致。

粘贴在本站创建的 API 密钥。

OpenCode 会把凭证保存在它自己的认证数据中，因此密钥不必出现在 `opencode.json` 里。

### 3. 选择并验证模型

输入 `/models`，选择 `aiwanai/deepseek-v4-pro`。让它检查当前项目并总结技术栈。

文件与工具访问成功，且在本站的 Usage Logs 中能看到对应记录，即说明连接正常。

### 方式二：ZCode（桌面用户推荐）

ZCode 是一款桌面编程工具，支持自定义的 OpenAI 兼容与 Anthropic 兼容供应商。

### 1. 安装或升级

从 ZCode 官方下载页面下载当前版本。它提供 Windows x64 / ARM64、macOS Apple Silicon / Intel 以及 Linux x64 安装包。

老用户可以直接升级，不要求特定历史版本。

### 2. 打开模型设置

首次启动且未配置模型时，选择 Use API Key。

在工作区中打开模型选择器，点击 Manage Models，然后打开 Settings → Model Settings。

### 3. 添加自定义供应商

在供应商列表底部点击 Add Provider，填写：

| 字段 | 取值 |
| --- | --- |
| Name | waner~ API站 - DeepSeek |
| Base URL | `https://api.aiwanai.cc/v1` |
| API Key | 你在本站的密钥 |
| API format | Responses (/responses) |

保存并启用该供应商。不要连接 Z.ai / 大模型编码计划，也不要使用指向 `api.deepseek.com` 的内置配置：它们都不经过本网关。

在 API 格式菜单中选择 Responses (`/responses`)。ZCode 会在 `https://api.aiwanai.cc/v1` 后追加 `/responses`，正好是本站的 POST `/v1/responses` 端点。不要在 Base URL 里重复写 `/responses`。

ZCode 通常会自动加载模型列表。如果看不到目标模型，点击 Add Model，输入确切的 ID `deepseek-v4-pro`。

按需添加 `deepseek-v4-flash` 或 Pricing 页面上的其他 DS ID。单模型最大输出 tokens 初始留空即可。

### 4. 选择并验证模型

在刚添加的供应商下选择 deepseek-v4-pro，然后让 Agent 列出项目文件并总结仓库。到本站的 Usage Logs 核对模型、分组与扣费。

### 该选哪一个？

| 偏好 | 推荐 |
| --- | --- |
| 终端优先、配置可审计、配置与项目放在一起 | OpenCode |
| 可视化管理项目、模型与会话 | ZCode |
| 一套配置在多个仓库复用 | OpenCode 全局配置 |
| 第一次使用编程 Agent | ZCode 更容易上手 |

两款工具使用的是同一套站点密钥、模型与计费，不需要单独订阅模型。

### 故障排查

401 / Unauthorized：请使用本站的密钥，而不是 DeepSeek、Z.ai 或其他平台的密钥。

404 / Not Found：请使用 `https://api.aiwanai.cc/v1`；不要使用完整的 `/v1/responses` 路径，不要省略 `/v1`，也不要写成 `/v1/v1`。

模型未找到：请使用小写 ID `deepseek-v4-pro`，然后检查 Pricing 页面、密钥分组与模型白名单。

OpenCode 供应商缺失：通过 `/connect` 输入的供应商 ID 必须是 `aiwanai`，且配置必须是合法的 JSON。

ZCode 未加载出模型：确认已选择 Responses (`/responses`)、已保存并启用供应商，然后手动添加确切的模型 ID；若仍不行，复查密钥与 URL。

聊天正常但编程工具表现异常：开启新会话，确认所选模型属于该自定义供应商，并把复杂任务拆成可验证的小步。

### 可核对的来源

本站实时模型列表与分组：模型 ID、协议与分组的权威依据。

OpenCode 安装与自定义供应商：安装、/connect、@ai-sdk/openai、Responses API 与 `baseURL` 的官方说明。

OpenCode 模型：/models 与供应商/模型格式的官方说明。

ZCode 下载页与模型连接：平台安装包、自定义 OpenAI 兼容供应商、模型管理与验证的官方说明；在当前客户端的 API 格式下选择 Responses (/responses)。

OpenAI 模型与 Responses 指南：多轮、工具调用与缓存工作流的协议背景；DeepSeek 的支持情况仍以本网关为准。
