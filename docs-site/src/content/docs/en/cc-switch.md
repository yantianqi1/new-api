---
title: CC Switch One-Click Import
---

CC Switch is an open-source, cross-platform desktop tool that manages provider configs for Claude Code, Codex, Gemini CLI, OpenCode, OpenClaw, Hermes, and Pi — and switches between sites with one click. waner~ API站 integrates its ccswitch:// import protocol, so no manual config editing is needed.

### Step 1: Download and install

Get the installer for your platform (Windows / macOS / Linux):

GitHub Releases: https://github.com/farion1231/cc-switch/releases/latest

Launch it once after installing so the ccswitch:// protocol registers with your system. Use v3.20.3 or newer. This site's exports were checked against that release and the same main commit on 2026-09-12.

### Step 2: One-click import

In **Console → API Keys**, open any key's action menu and pick **Import to CC Switch**. Your browser opens CC Switch with a confirmation dialog; after confirming, this site is written into the tool's config as a provider and activated.

One-click import supports Codex, Claude Code, Gemini CLI, OpenCode, OpenClaw, Hermes, and Grok Build (requires CC Switch v3.18+; older versions reject the app id). If this site publishes multiple API endpoints, you can choose to import backup endpoints too and switch between them in CC Switch.

"Include balance query" is checked by default: the imported provider carries a usage script so CC Switch shows both the available balance and cumulative usage. Limited keys show their own figures; unlimited keys show the owning account's figures. The script calls this site's/api/usage/token/with the key itself; no account-level credentials are included.

CC Switch itself also supports Claude Desktop, but its current official provider Deep Link protocol does not accept claude-desktop. Add a Claude Desktop provider manually inside CC Switch instead of using this site's one-click import.

Pi (Pi Coding Agent) configuration export: choose Pi in this site's import dialog, select the API format, endpoint, and model, then click Copy Pi configuration. In CC Switch, go to Pi → Add Provider → Custom Configuration, enter a name and a unique provider key, paste the copied object into Configuration JSON, and save. The export is a single provider object containing the endpoint, API key, and model; CC Switch writes it into Pi's `models.json`. The official v3.20.3 deep-link protocol still rejectsapp=pi, so this entry copies JSON.

For Claude models in Pi, choose Anthropic Messages; the export automatically uses the site root. For OpenAI formats, choose OpenAI Chat Completions or OpenAI Responses; the export adds /v1 automatically. Check model names against the Pricing page and your key group. Pi exports contain only the selected endpoint, without CC Switch backup endpoints or usage scripts. You can set context limits, output limits, and reasoning options in CC Switch according to the model's capabilities.

The importer handles URL suffixes for you: Claude Code and Gemini CLI use the site root; Codex, Grok Build, OpenCode, OpenClaw, and Hermes use the /v1 URL. Pi follows the selected API format. Do not add or remove /v1 manually.

### Pick models per target app when importing

Codex: prefer gpt-5.6; use this site's suffixed models only after checking the Pricing page and your key group. For web search, add `web_search` = "live"at the top level of the imported provider config — the import dialog offers a copy button for the line (the Deep Link protocol cannot carry this key). To apply it to every provider, put it into CC Switch's Codex common-config snippet and enable "Apply Common Config" on each provider; see the Codex CLI guide.

Claude Code: choose claude-fable-5 for strongest capability, claude-sonnet-5 for balanced daily use, and claude-haiku-4-5 for the Haiku slot.grok-4.5 is also available; verify actual cache usage in the logs before relying on it for long sessions.

Grok Build: choose grok-4.5.

Claude Code import writes the primary model plus the Opus / Sonnet / Haiku / Fable slots. Fable and Subagent Model travel in the config payload as ANTHROPIC_DEFAULT_FABLE_MODEL and CLAUDE_CODE_SUBAGENT_MODEL; ANTHROPIC_SMALL_FAST_MODEL is also set from the Haiku slot for older Claude Code versions. If the upstream rejects the Artifact tool schema, select Disable Artifact tool to export CLAUDE_CODE_DISABLE_ARTIFACT=1. This option is off by default and leaves other tools available.

Enable Claude local proxy/takeover in CC Switch only when you need to give Claude Code a model that exposes OpenAI-format endpoints only. This site's Anthropic Messages endpoint does not need it for the direct setup in this guide.

### Import from the keys page

Every key's action menu in Console → API Keys has the same "Import to CC Switch" entry.

### Troubleshooting

Nothing opens after clicking: make sure CC Switch is installed and has been launched at least once; allow the browser's "Open CC Switch?" prompt.

Import fails or says the app is invalid: update CC Switch and try again. Add Claude Desktop manually inside the app; for Pi, use Copy Pi configuration in this site's dialog (see above).

Tool still can't connect after import: open CC Switch and check the provider endpoint, key, and model name; then confirm the model is enabled for your key group on this site's Pricing page. Restarting the CLI usually applies changes for tools other than Claude Code.

Sharing configs manually: CC Switch also accepts hand-built ccswitch:// v1/import?...links — see its official docs.

Security: import links and Pi configuration JSON contain your API key. Do not screenshot or publicly share them; import only from this site or another source you trust.

### Trusted sources and self-checks

This site's live model list and groups: confirm the model name and key group before importing.

CC Switch official README: supported apps, installation packages, and switching behavior.

CC Switch Deep Link protocol: ccswitch:// parameters and manual import format.

CC Switch v3.20.3 release notes: current features, upgrade notes, and fixes.

Audited importer source: Pi deep-link restrictions, Claude environment payloads, and Codex configuration behavior.
