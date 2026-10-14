---
title: DeepSeek + OpenCode / ZCode
---

For agentic coding with the Chinese DeepSeek models (DS), we recommend OpenCode or ZCode. OpenCode suits terminal-first users who want transparent configuration; ZCode is a desktop app with more visual project and model management.

Both tools must connect to this site through an OpenAI-compatible custom provider using the Responses API. Do not select a built-in DeepSeek provider: it normally points to DeepSeek's official API instead of waner~ API站.

Downloads: OpenCode official download·ZCode official download for Windows, macOS, and Linux

### Connection details

| Setting | Value |
| --- | --- |
| Provider type | Custom / OpenAI Compatible |
| API Base URL | `https://api.aiwanai.cc/v1` |
| API key | A key created under Console → API Keys on this site |
| Recommended endpoint | `POST https://api.aiwanai.cc/v1/responses` |
| Recommended model | `deepseek-v4-pro` |
| Lower-cost model | `deepseek-v4-flash` |

Model IDs are case-sensitive. The current ID is deepseek-v4-pro, not DeepSeek-V4-Pro. Models and groups may change; check the live Pricing page before configuring a client.

Why `/v1/responses` is recommended

We verified the gateway route and its DeepSeek adapter: this site accepts POST `/v1/responses` and relays Responses requests through the corresponding upstream DeepSeek endpoint. OpenCode uses it through @ai-sdk/openai, and ZCode's custom-provider settings offer an explicit Responses (/responses) API format. Both configurations below use the Responses API.

The Responses API is a better fit for multi-turn agent context and tool calls. This gateway can also accept and pass cache-related fields such as prompt_cache_key, making long sessions more likely to reuse cached context consistently and avoid recomputing repeated prefixes. It does not guarantee a cache hit on every request: keep shared system instructions and tool definitions stable, and use a consistent model and route.

Important: enter only https://api.aiwanai.cc/v1 as the Base URL in either client. Do not enter https://api.aiwanai.cc/v1/responses; clients append their own endpoint path and may otherwise construct an invalid URL.

Before you begin

Sign in and make sure your wallet has sufficient balance.

Create a dedicated key under Console → API Keys.

Choose a key group that supports the DS model shown on the Pricing page. deepseek-v4-pro currently appears in the domestic-model and selected Alibaba Cloud domestic-model groups.

If the key has a model allowlist, add deepseek-v4-pro and every other DS model you plan to use.

Use a separate key for each tool so usage and revocation remain independent. Never commit a key to Git.

### Option 1: OpenCode (recommended for terminal users)

OpenCode is an open-source coding agent available in the terminal, desktop app, and IDE extensions.

### Install from scratch

Choose an installation method from the official download page. If Node.js is already installed, run:

```bash
npm install -g opencode-ai
```

The official documentation recommends WSL on Windows; npm, Scoop, and Chocolatey are also supported. Verify the installation with opencode --version.

### Migrate an existing installation

Update to the current release:

```bash
opencode upgrade
```

There is no need to rely on a hard-coded minimum version. If your installation channel does not support upgrade, update through that channel. Merge the provider below into an existing config rather than overwriting unrelated settings.

### 1. Configure the provider

For access from every project, edit the global config:

Windows:%USERPROFILE%\.config\opencode\`opencode.json`

macOS / Linux / WSL: ~/.config/opencode/`opencode.json`

For project-only access, place opencode. JSON in the project root instead.

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

Use @ai-sdk/openai here. OpenCode will call /responses under https://api.aiwanai.cc/v1. Do not replace it with @ai-sdk/openai-compatible, which targets `/v1/chat/completions`; use that package only as a compatibility fallback if an old OpenCode release cannot use Responses.

### 2. Store the API key

Run opencodein a project, enter /connect, then:

Select Other, not the built-in DeepSeek provider.

Enter aiwanai as the provider ID; it must match the config exactly.

Paste the API key created on this site.

OpenCode stores the credential in its authentication data, so the key does not need to appear in `opencode.json`.

### 3. Select and verify the model

Enter /models and selectaiwanai/deepseek-v4-pro. Ask it to inspect the current project and summarize its stack. Successful file/tool access and a matching entry under this site's Usage Logs confirm the connection.

### Option 2: ZCode (recommended for desktop users)

ZCode is a desktop coding tool that supports custom OpenAI- and Anthropic-compatible providers.

### 1. Install or upgrade

Download the current build from the official ZCode download page. It provides packages for Windows x64 / ARM64, macOS Apple Silicon / Intel, and Linux x64. Existing users can upgrade directly; no specific historical version is required.

### 2. Open model settings

On first launch with no model configured, choose Use API Key.

From a workspace, open the model picker, click Manage Models, then open Settings → Model Settings.

### 3. Add a custom provider

Click Add Providerat the bottom of the provider list and enter:

Field Value

| Field | Value |
| --- | --- |
| Name | waner~ API站 - DeepSeek |
| Base URL | `https://api.aiwanai.cc/v1` |
| API Key | Your key from this site |
| API format | Responses (/responses) |

Save and enable the provider. Do not connect a Z.ai / Big Model Coding Plan or use a built-in configuration pointing to api.deepseek.com; neither uses this gateway.

Select Responses (/responses) from the API format menu. ZCode appends/responses to https://api.aiwanai.cc/v1, producing this site's POST `/v1/responses` endpoint. Do not repeat/responses in the Base URL.

ZCode normally loads the model list automatically. If the target is absent, click Add Model and enter the exact ID deepseek-v4-pro. Add deepseek-v4-flash or other DS IDs from the Pricing page as needed. Leave the per-model maximum output tokens empty initially.

### 4. Select and verify the model

Select deepseek-v4-pro under the provider you just added, then ask the agent to list the project files and summarize the repository. Confirm the model, group, and charge under this site's Usage Logs.

### Which one should you use?

| Preference | Recommendation |
| --- | --- |
| Terminal-first, auditable config, config alongside projects | OpenCode |
| Visual project, model, and conversation management | ZCode |
| Reuse one setup across many repositories | OpenCode global config |
| First coding-agent experience | ZCode is more approachable |

Both tools use the same site keys, models, and billing; no separate model subscription is required.

### Troubleshooting

401 / Unauthorized: use a key from this site, not a key from DeepSeek, Z.ai, or another platform.

404 / Not Found: use https://api.aiwanai.cc/v1; do not use the full `/v1/responses` path, omit /v1, or duplicate it as /v1/v1.

Model not found: use the lowercase ID deepseek-v4-pro, then check the Pricing page, key group, and model allowlist.

OpenCode provider missing: the provider ID entered through/connectmust be aiwanai, and the config must be valid JSON.

ZCode did not load models: confirm Responses (/responses) is selected, save and enable the provider, then add the exact model ID manually; recheck the key and URL if that fails.

Chat works but coding tools behave poorly: start a fresh session, confirm the selected model belongs to the custom provider, and split complex work into verifiable steps.

### Sources to verify

Live model list and groups: source of truth for model IDs, protocols, and groups.

OpenCode installation and custom providers: installation, /connect, @ai-sdk/openai, the Responses API, and `baseURL`.

OpenCode models:/models and the provider/model format.

ZCode download page and model connections: platform packages, custom OpenAI-compatible providers, model management, and verification; select Responses (/responses) under API format in the current client.

OpenAI model and Responses guidance: protocol background for multi-turn, tool-calling, and caching workflows; this gateway remains the source of truth for DeepSeek support.
