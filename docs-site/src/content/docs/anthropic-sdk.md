---
title: Anthropic SDK 接入
---

本站支持 Anthropic Messages 协议：将官方 SDK 的 `base_url` 指向 https://api.aiwanai.cc，即可直接调用 Claude 模型。

### Python

```bash
pip install anthropic
```

```python
import anthropic

client = anthropic.Anthropic(
    api_key="sk-your-api-key",
    base_url="https://api.aiwanai.cc",
)

message = client.messages.create(
    model="claude-sonnet-4-5-20250929",
    max_tokens=1024,
    messages=[{"role": "user", "content": "Say hello in one sentence"}],
)
print(message.content[0].text)
```

### JavaScript / TypeScript

```bash
npm install @anthropic-ai/sdk
```

```javascript
import Anthropic from '@anthropic-ai/sdk'

const client = new Anthropic({
  apiKey: 'sk-your-api-key',
  baseURL: 'https://api.aiwanai.cc',
})

const message = await client.messages.create({
  model: 'claude-sonnet-4-5-20250929',
  max_tokens: 1024,
  messages: [{ role: 'user', content: 'Say hello in one sentence' }],
})
console.log(message.content)
```

### 要点

`base_url` 为站点根地址 https://api.aiwanai.cc（SDK 会自动追加 `/v1/messages`）——不要带 `/v1` 后缀

SDK 默认的 x-api-key 请求头与 Authorization: Bearer 均可使用

Anthropic 协议要求必须指定 max_tokens

Claude 模型也可通过 OpenAI 兼容端点访问——两种方式任选其一

模型名来自 Pricing 页面；遇到错误请参阅错误码与故障排查
