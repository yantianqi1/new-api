---
title: Claude Code
---

Claude Code is Anthropic's official terminal coding assistant. Two environment variables are all it takes to route it through waner~ API站.

💡Recommended: skip the manual setup with CC Switch one-click import — it writes the Claude Code config for you and lets you switch between sites in one click.

### Prerequisites

Node.js 18+ installed

An API key created on this site (Console → API Keys)

### Step 1: Install Claude Code

Choose an installation method for your system from the official Claude Code installation guide (the official installer is recommended for macOS / Linux; WinGet or the PowerShell installer for Windows).

npm install -g @anthropic-ai/claude-code still works, but Anthropic marks npm installation as deprecated. Use it only when you specifically need npm.

### Step 2: Write the configuration

Edit (or create)~/.claude/`settings.json`:

Windows: press Win + R, enter%userprofile%\.claude, create settings. JSON there

macOS / Linux: run mkdir -p ~/.claude and open the folder

```json
{
  "env": {
    "ANTHROPIC_BASE_URL": "https://api.aiwanai.cc",
    "ANTHROPIC_AUTH_TOKEN": "sk-your-api-key"
  }
}
```

Note: ANTHROPIC_BASE_URL is the site root — do not append /v1.

To pin models (optional), extendenv. Claude Code selects different tiers through model aliases and the Default option; the variables below control which model each alias resolves to:

```json
{
  "env": {
    "ANTHROPIC_BASE_URL": "https://api.aiwanai.cc",
    "ANTHROPIC_AUTH_TOKEN": "sk-your-api-key",
    "ANTHROPIC_MODEL": "claude-fable-5",
    "ANTHROPIC_DEFAULT_OPUS_MODEL": "claude-opus-4-8",
    "ANTHROPIC_DEFAULT_SONNET_MODEL": "claude-sonnet-5",
    "ANTHROPIC_DEFAULT_HAIKU_MODEL": "claude-haiku-4-5"
  }
}
```

ANTHROPIC_MODEL: the main conversation model.claude-fable-5 is the strongest current flagship; swap in claude-sonnet-5 for cheaper daily use.

ANTHROPIC_DEFAULT_OPUS_MODEL/ANTHROPIC_DEFAULT_SONNET_MODEL/ANTHROPIC_DEFAULT_HAIKU_MODEL: control what theopus, sonnet, and Haiku aliases — and the Default option — resolve to.

ANTHROPIC_DEFAULT_HAIKU_MODEL also serves background work such as session-title generation. Set it to an available Haiku model; ANTHROPIC_SMALL_FAST_MODEL is deprecated and should not be used.

The names above are examples; use model names exactly as shown on this site's Pricing page.

### Advanced: run grok-4.5 inside Claude Code

This site automatically converts Claude Code's requests into the format grok-4.5 understands, so you can simply point the main model at it:

```json
{
  "env": {
    "ANTHROPIC_BASE_URL": "https://api.aiwanai.cc",
    "ANTHROPIC_AUTH_TOKEN": "sk-your-api-key",
    "ANTHROPIC_MODEL": "grok-4.5"
  }
}
```

grok-4.5 is an official xAI model and appears in this site's current model list. xAI supports prompt caching, but real cache hits and cost depend on the request prefix, routing, and this gateway's implementation. Check cache usage in this site's logs before choosing it as your daily model. Selecting grok-4.5 as the primary model during CC Switch one-click import applies the same configuration.

### Step 3: Verify

Run this inside any project:

```bash
claude
```

Ask it anything — a reply means the integration works. Run /status in Claude Code to check the Anthropic base URL and active credential, and verify the call in this site's Usage Logs. If you set shell environment variables instead of using `settings.json`, open a new terminal first.

### Troubleshooting

Connection error / fetch failed: is ANTHROPIC_BASE_URL exactly https://api.aiwanai.cc (no trailing /v1)? Use /status to confirm the active session loaded it.

401: is the key complete (including the sk-prefix) and still active?

Model unavailable: does your key's group include that model? See Error codes & troubleshooting.

### Trusted sources and self-checks

This site's live model list and groups: use this as the source of truth for available model names and key groups.

Claude Code: gateway connection: official guidance for paths, ANTHROPIC_BASE_URL, credentials, and /status.

Claude Code: model configuration: model aliases, ANTHROPIC_DEFAULT_*_MODEL, and deprecated variables.

xAI:  grok-4.5 and prompt caching.
