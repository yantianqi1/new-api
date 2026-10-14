---
title: Cherry Studio
---

Cherry Studio 是一款流行的开源桌面端 AI 客户端（Windows / macOS / Linux），支持多服务商接入、知识库与图片生成。

将 waner~ API站 作为 OpenAI 类型的服务商添加，即可使用本站所有模型。

### 第 1 步：安装

前往官网下载： https://www.cherry-ai.com

### 第 2 步：添加服务商

打开 设置 → 模型服务，点击「添加服务商」（Add Provider）。

名称随意（例如 waner~ API站），类型选择 OpenAI。

### 填写

API Key：sk-your-api-key

API Host： https://api.aiwanai.cc

填站点根地址即可——Cherry Studio 会自动补上 /v1。如果你填写的是完整路径，请在末尾加上 #，以强制使用该确切地址。

### 第 3 步：添加模型

在服务商的模型区域点击「管理」（Manage），Cherry Studio 会自动拉取本站的模型列表。勾选你需要的模型（详见 Pricing 页面）。

### 第 4 步：验证

回到聊天界面，在模型选择器中切换到新添加的模型并发送一条消息。能收到回复即表示成功，该次调用会显示在 Usage Logs 中。

### 常见问题

模型列表无法加载：检查地址是否为 https://api.aiwanai.cc，以及密钥是否有效。

聊天时出现 404：该模型不在你密钥所属的分组中——换一个模型，或调整密钥分组。

想要一键配置：如果管理员配置了聊天预设，控制台侧边栏的聊天入口可以深链拉起 Cherry Studio，并自动带入你的密钥。
