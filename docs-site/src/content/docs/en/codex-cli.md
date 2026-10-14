---
title: Codex CLI
---

Codex CLI is OpenAI's official terminal coding assistant. One config.toml routes it through waner~ API站.

💡Recommended: skip the manual setup with CC Switch one-click import — it writes the Codex config for you and lets you switch between sites in one click.

### Prerequisites

Node.js 18+ installed

An API key created on this site

### Step 1: Install Codex CLI

```bash
npm install -g @openai/codex
```

### Step 2: Write the configuration

Edit (or create)~/.codex/`config.toml`:

Windows: press Win + R, enter%userprofile%\.codex

macOS / Linux: mkdir -p ~/.codex

```toml
#:schema https://developers.openai.com/codex/config-schema.json
model_provider = "custom"
model = "gpt-5.6"
model_reasoning_effort = "high"

[model_providers.custom]
name = "waner~ API站"
base_url = "https://api.aiwanai.cc/v1"
wire_api = "responses"
requires_openai_auth = true
```

Unlike Claude Code, Codex's base_url must include the /v1 suffix, and wire_api uses"responses"(the current and only supported value). The first schema line lets compatible editors flag invalid TOML keys.

gpt-5.6 is listed in both OpenAI's official Codex documentation and this site's current model list. This site may also offer suffixed routing models; those are site-specific names, not official OpenAI aliases. Before using one, confirm that it appears on the Pricing page and is enabled for your key group.

Then store the key in~/.codex/`auth.json`:

```json
{
  "OPENAI_API_KEY": "sk-your-api-key"
}
```

### Step 3: Verify

Restart the terminal and run inside any project:

```bash
codex
```

A working conversation means you're connected; calls appear in this site's Usage Logs.

### Optional: enable web search

Recent Codex versions gate web search behind the top-level `web_search` key in `config.toml` (not inside [model_providers.custom]):

```toml
web_search = "live"
```

When enabled, Codex performs searches through this site's POST /v1/alpha/search endpoint. Notes:

Searches are served by the site channels that support this endpoint; if your key's group has none yet, searches fail while normal chat keeps working.

Searches are billed per call (the `web_search` tool price, $0.01/call by default) and appear as separate entries in Usage Logs; failed searches are not charged.

Stability: when one channel fails a search (overload, timeout, ...), the site automatically retries on another channel — no action needed. A real live search can take tens of seconds; that is normal.

Remove the line, or set `web_search` = "disabled", when you don't need it.

If you manage Codex through CC Switch, the ccswitch:// import protocol cannot carry this key: this site's import dialog offers a copy button for the line — paste it at the top level of the imported provider config. To apply it globally, put it into CC Switch's Codex common-config snippet and enable "Apply Common Config" on each provider.

### Troubleshooting

401 / Unauthorized: is the key inauth. JSON complete? (It does not belong in `config.toml`.)

404 / model not found: does themodelname exist on the Pricing page and in your key's group?

Stream errors / timeouts: system proxies often interfere — try disabling them for the terminal.

### Trusted sources and self-checks

This site's live model list and groups: source of truth for site model names and key groups.

Codex config basics: official~/.codex/config. TOML location and baseline configuration.

Codex advanced configuration: custom-provider, `base_url`, and wire_api examples.

Codex complete configuration reference: every currently supported TOML key and value. Check this and the schema after upgrading Codex.
