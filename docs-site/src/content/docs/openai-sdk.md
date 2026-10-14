---
title: OpenAI SDK 接入
---

本站完全兼容 OpenAI 接口：将官方 SDK 的 `base_url` 指向 https://api.aiwanai.cc/v1，其余用法与官方 API 完全一致。

### Python

```bash
pip install openai
```

```python
from openai import OpenAI

client = OpenAI(
    api_key="sk-your-api-key",
    base_url="https://api.aiwanai.cc/v1",
)

response = client.chat.completions.create(
    model="gpt-4o-mini",
    messages=[{"role": "user", "content": "Say hello in one sentence"}],
)
print(response.choices[0].message.content)
```

### 流式输出

```python
stream = client.chat.completions.create(
    model="gpt-4o-mini",
    messages=[{"role": "user", "content": "Write a four-line poem"}],
    stream=True,
)
for chunk in stream:
    delta = chunk.choices[0].delta.content
    if delta:
        print(delta, end="", flush=True)
```

### JavaScript / TypeScript

```bash
npm install openai
```

```javascript
import OpenAI from 'openai'

const client = new OpenAI({
  apiKey: 'sk-your-api-key',
  baseURL: 'https://api.aiwanai.cc/v1',
})

const response = await client.chat.completions.create({
  model: 'gpt-4o-mini',
  messages: [{ role: 'user', content: 'Say hello in one sentence' }],
})
console.log(response.choices[0].message.content)
```

### 注意事项

密钥应放入环境变量：生产环境请使用 OPENAI_API_KEY/OPENAI_BASE_URL，不要硬编码在代码中

模型名来自 Pricing 页面；跨厂商模型（Claude、Gemini 等）均可通过此 OpenAI 兼容端点调用

LangChain / LlamaIndex 及类似框架：任何支持配置 base_url 的 OpenAI 集成都可直接使用

遇到错误请参阅错误码与故障排查
