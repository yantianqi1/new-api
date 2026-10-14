---
title: API 参考
---

waner~ API站 API 兼容官方 OpenAI / Anthropic / Gemini 协议，采用 Bearer Token 鉴权——一个 `sk-` 密钥即可访问所有端点。

### 主要端点

- **`POST /v1/chat/completions`** —— 根据对话历史生成模型回复；兼容 OpenAI Chat Completions，模型支持时可用流式输出。这是 OpenAI 兼容客户端的主端点；流式响应使用 SSE，客户端需自行处理增量分片与最终的用量统计；转发时后端会保留客户端显式传入的零值与 `false`。
- **`POST /v1/responses`** —— OpenAI Responses 协议路由；`/v1/responses/compact` 是其精简变体。
- **`POST /v1/messages`** —— Anthropic 兼容的消息路由。
- **`POST /v1beta/models/{model}: generateContent`** —— Gemini 兼容路由。
- **`POST /v1/images/generations`** —— 图像生成；始终返回图片 URL 而非 base64，详见[图像生成](/image-generation/)。
- **`POST /v1/embeddings`** / **`POST /v1/rerank`** —— 向量嵌入与重排序。
- **`GET /v1/models`** —— 当前可用模型列表。

**完整端点列表：**

| 类型 | 方法 | 路径 |
| --- | --- | --- |
| models | GET | `/v1/models` |
| anthropic | POST | `/v1/messages` |
| embeddings | POST | `/v1/embeddings` |
| gemini | POST | `/v1beta/models/{model}: generateContent` |
| image-generation | POST | `/v1/images/generations` |
| jina-rerank | POST | `/v1/rerank` |
| openai | POST | `/v1/chat/completions` |
| openai-response | POST | `/v1/responses` |
| openai-response-compact | POST | `/v1/responses/compact` |

**兼容性说明：**
- `/v1/chat/completions` 遵循 OpenAI 兼容的 chat completion 语义。
- `/v1/responses` 在已配置的情况下遵循 OpenAI 兼容的 responses 语义。
- `/v1/messages` 提供 Anthropic 兼容的消息路由。
- `/v1beta/models/{model}: generateContent` 提供 Gemini 兼容路由。
- 您实际可用的模型取决于已启用的渠道、您的分组、令牌限制以及计费配置。
