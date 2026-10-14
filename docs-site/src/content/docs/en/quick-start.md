---
title: Quick Start
description: Five minutes from sign-up to the first request — Base URL, an sk- key, and a curl example.
---

waner~ API站 is an AI API gateway. Create an `sk-` token in the [console](https://api.aiwanai.cc), point your client at this site, and the gateway relays the request upstream.

## Connection details

| Address | Notes |
| --- | --- |
| `https://api.aiwanai.cc` | The single site-wide endpoint, usable globally — no need to pick a line |

Authentication header:

```http
Authorization: Bearer <token>
```

Create an API token in the console. Never hand your console session cookie to third-party AI tools.

## Five steps to the first call

### Step 1: Open the console and sign in

Open the [console](https://api.aiwanai.cc) to register or sign in. If the site is invite-only, get an invitation code first. The Console button in the header (or the mobile menu) opens the same URL.

### Step 2: Create an API key

In the console, open **API Keys** and click Create. Recommendations:

- Use one key per app or tool so you can disable and audit them independently
- The default group is fine for new users; "unlimited quota" keys draw from your wallet balance
- Keys are sensitive credentials — never share them publicly; delete and recreate if leaked

### Step 3: Check your balance

Open Wallet to see your balance. Top up or redeem a code if needed; every call is itemized in Usage Logs.

New accounts get a **¥1 bonus** (500,000 quota) on signup. Daily check-in is worth about **¥0.9**, and both sides get **¥1** on invitations; bonus quota is limited to `[free]`-prefixed models and `gpt-5.6-sol`. Top-ups via Alipay (Yi Pay) at 10 / 20 / 50 / 100 / 200 / 500 / 1000. For live model and pricing details, see Pricing in the console.

### Step 4: Send your first request

```bash
curl https://api.aiwanai.cc/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-your-api-key" \
  -d '{
    "model": "deepseek-v4-pro",
    "messages": [{"role": "user", "content": "Say hello in one sentence"}]
  }'
```

Replace the model with any name listed on the Pricing page.

### Step 5: Verify and continue

The fastest check is chatting inside the site: open Playground, pick a model, and send one message. A reply means the gateway, group, and key are all working.

Next, pick the guide for your tool:

- [Coding tools](/en/claude-code/): Claude Code, Codex CLI, DeepSeek with OpenCode / ZCode, or one-click import via CC Switch
- [Desktop clients](/en/cherry-studio/): Cherry Studio
- [Browser translation](/en/immersive-translate/): Immersive Translate
- [Writing code](/en/openai-sdk/): OpenAI SDK, [Anthropic SDK](/en/anthropic-sdk/)
- [Hitting errors](/en/troubleshooting/): Error codes & troubleshooting
