---
title: OpenAI SDK Integration
---

This site is fully OpenAI-compatible: point the official SDK's `base_url` at https://api.aiwanai.cc/v1 and everything else works exactly like the official API.

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

### Streaming

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

### Notes

Keys belong in env vars: use OPENAI_API_KEY/OPENAI_BASE_URL in production instead of hardcoding

Model names come from the Pricing page; cross-vendor models (Claude, Gemini, …) are all callable through this OpenAI-compatible endpoint

LangChain / LlamaIndex and friends: any OpenAI integration with a configurable base_url works as-is

For errors, see Error codes & troubleshooting
