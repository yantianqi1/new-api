---
title: 快速开始
description: 五分钟从注册到发出第一请求：Base URL、sk- 密钥与 curl 示例。
---

waner~ API站 是一个 AI API 网关。你在[控制台](https://api.aiwanai.cc)创建 `sk-` 密钥，把请求指到本站，站点会中继到上游模型。

## 接入信息

| 地址 | 说明 |
| --- | --- |
| `https://api.aiwanai.cc` | 全站统一接入地址，全球可用，无需选择线路 |

认证头：

```http
Authorization: Bearer <token>
```

在控制台创建 API token。切勿把控制台登录态（session cookie）交给第三方 AI 工具。

## 五步发出第一请求

### 第 1 步：打开控制台并登录

打开 [控制台](https://api.aiwanai.cc) 注册或登录。若站点处于邀请制，请先获取邀请码。文档站顶栏或菜单里的「控制台」也会打开同一地址。

### 第 2 步：创建 API key

在控制台打开 **API Keys**，点击 Create。建议：

- 每个应用或工具使用独立的 key，便于单独禁用与审计
- 新用户用默认分组即可；「无限额度」的 key 会消耗钱包余额
- key 是敏感凭证，切勿公开泄露；一旦泄露请立即删除并重新创建

### 第 3 步：查看余额

打开钱包查看余额。按需充值或兑换码；每次调用都会在 Usage Logs 中逐条记录。

新账号注册即获 **1 元额度**（50 万配额）。每日签到约值 **0.9 元**，邀请双方各得 **1 元**；赠送额度仅限 `[free]` 前缀模型与 `gpt-5.6-sol`。可通过支付宝（易支付）充值，档位为 10 / 20 / 50 / 100 / 200 / 500 / 1000。实时模型与价格见控制台 Pricing 页。

### 第 4 步：发送第一个请求

```bash
curl https://api.aiwanai.cc/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-your-api-key" \
  -d '{
    "model": "deepseek-v4-pro",
    "messages": [{"role": "user", "content": "Say hello in one sentence"}]
  }'
```

模型名换成控制台 Pricing 页里的任意可用模型即可。

### 第 5 步：验证与继续

最快的验证方式是在站内聊天：打开 Playground，选一个模型，发一条消息。能收到回复，就说明网关、分组和 key 都正常。

接下来按工具选指南：

- [编码工具](/claude-code/)：Claude Code、Codex CLI、DeepSeek + OpenCode / ZCode，或 CC Switch 一键导入
- [桌面客户端](/cherry-studio/)：Cherry Studio
- [浏览器翻译](/immersive-translate/)：沉浸式翻译
- [写代码](/openai-sdk/)：OpenAI SDK、[Anthropic SDK](/anthropic-sdk/)
- [遇到报错](/troubleshooting/)：错误码与排障
