---
title: AQBot
---

AQBot 是一款开源的 AI 桌面客户端，支持多模型聊天、MCP 服务器与知识库。

本站支持 AQBot 的 `aqbot://providersimport` 协议，因此你可以一步将本站、端点和 API Key 发送给 AQBot。

### 官方链接

AQBot 官网： https://app.aqbot.top/

下载地址： https://app.aqbot.top/download

GitHub： https://github.com/AQBot-Desktop/AQBot

导入协议：Importing from a Website Link

### 一键导入

安装 AQBot 并至少启动一次，以便系统注册 `aqbot://` 协议。

在本站控制台侧边栏，打开 第三方 → AQBot。

选择要导入的 API Key。如果本站发布了多个端点，请同时选择目标端点。

浏览器会打开 AQBot。核对服务商名称、Base URL、类型和 API Key 前缀，然后确认导入。

你也可以在 控制台 → API Keys 下打开任意密钥的操作菜单，选择 Chat → AQBot 直接导出该密钥。

导入链接中包含完整的 API Key。请勿复制、截图或公开分享完整链接，并且只确认来自本站或其他你信任来源的导入。

### 手动配置

如果浏览器没有打开 AQBot，请在 AQBot 中前往 设置 → Providers，手动添加一个 OpenAI 类型的服务商：

名称：waner~ API站

Base URL： https://api.aiwanai.cc

API Key：sk-your-api-key

导入或保存后，点击 Fetch Models，或参照本站的 Pricing 页面手动添加模型 ID。

AQBot 在通过链接导入时不会验证 API Key，也不会自动同步模型。

### 常见问题

AQBot 未打开：确认已安装并启动过 AQBot，然后允许浏览器打开外部应用。

端点被拒绝：Base URL 必须使用 http 或 https，且不能包含查询字符串或 fragment。

导入后没有模型：在 AQBot 的服务商设置中点击 Fetch Models，或手动添加你的密钥分组可用的模型 ID。

请求返回 401/403：检查 API Key 是否正确，并确认其分组可以访问所选的模型。
