---
title: Anthropic SDK Integration
---

This site speaks the Anthropic Messages protocol: point the official SDK's `base_url` at https://api.aiwanai.cc and call Claude models directly.

### Python

```bash
pip install anthropic
```

```python
import anthropic
```

```text
client = anthropic. Anthropic(
    api_key="sk-your-api-key",
    base_url="https://api.aiwanai.cc",
)
```

```python
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

### Key points

base_url is the site root https://api.aiwanai.cc (the SDK appends `/v1/messages` itself) — no/v1 suffix

Both the SDK's default x-api-key header and Authorization: Bearer work

max_tokens is required by the Anthropic protocol

Claude models are also reachable via the OpenAI-compatible endpoint — pick either

Model names come from the Pricing page; for errors see Error codes & troubleshooting
