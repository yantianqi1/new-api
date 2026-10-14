---
title: XiaoBaiSwitch One-Click Import
---

XiaoBaiSwitch is an open-source desktop tool that manages Claude Code, Codex, and Pi configs around "upstream sites": one site carries multiple routes with speed tests, one-click switching, and automatic backups before applying. waner~ API站 integrates its xiaobaiswitch:// import protocol — the site, every endpoint, and the Codex capability presets travel in a single link, no manual config editing needed.

### Step 1: Download and install

Get the installer for your platform (Windows x64 / ARM64, macOS Apple Silicon / Intel):

GitHub Releases: https://github.com/Licoy/xiaobai-switch/releases/latest

Launch it once after installing so the xiaobaiswitch:// protocol registers with your system. Use v0.0.6 or newer (link import has been available since v0.0.3; Pi Coding Agent support arrived in v0.0.6).

The macOS package is ad-hoc signed, so the system may claim it is "damaged". Drag the app into Applications, run `xattr` -cr /Applications/XiaoBaiSwitch.appin a terminal, then right-click → Open. On Windows, install Web View2 if the system prompts for it on first run.

### Step 2: One-click import

In **Console → API Keys**, open any key's action menu and pick **Import to XiaoBaiSwitch**. Your browser opens the app with a confirmation dialog; check the site name, routes, and key prefix, then confirm to write this site in as an upstream site.

Unlike CC Switch's per-tool imports, XiaoBaiSwitch imports the whole site: every API endpoint of this site is written at once (the first one is the default route; speed-test and switch inside the app). Models are not carried in the link — the app fetches the live list from this site, so it is always current.

Importing does not auto-apply to Claude Code / Codex / Pi. After importing, open the apply center, pick this site and a model for the target tool, and apply; the original config is backed up automatically and can be restored anytime.

The "Codex capability presets" in the import dialog (web search / image understanding / image generation / remote compaction) are stored on the site and offered as defaults when applying to Codex. Web search and image understanding are supported here and checked by default; image generation needs a key group with image models and remote compaction depends on upstream support, so both start unchecked — enable them as needed.

### Pick models per target tool when applying

Codex: prefer gpt-5.6; use this site's suffixed models only after checking the Pricing page and your key group.

Claude Code: choose claude-fable-5 for strongest capability and claude-sonnet-5 for balanced daily use.

Pi (requires XiaoBaiSwitch v0.0.6+): this site imports with the OpenAI-compatible protocol, so applying it to Pi writes an openai-completions entry (the /v1 URL is derived automatically), optionally copying this site's full model list into Pi's catalog. OpenAI-family models such as gpt-5.6 work as-is. If you mainly run claude-fable-5/claude-sonnet-5 in Pi, first switch the site's protocol to Anthropic in the site editor — Pi is then written as Anthropic-messages+ the site root, which keeps Claude prompt caching effective (the protocol change does not affect how Claude Code / Codex are applied).

### Import from the keys page

Every key's action menu in Console → API Keys has the same "XiaoBaiSwitch" entry. The sidebar Third-party → Xiao Bai Switch entry opens the same import at any time.

### Troubleshooting

Nothing opens after clicking: make sure XiaoBaiSwitch is installed and has been launched at least once; allow the browser's "Open XiaoBaiSwitch?" prompt.

What happens on re-import: a site with the same protocol and route set (order-insensitive) is treated as the same site — the same key is reused and a different key updates it; adding or removing routes creates a new site instead of merging.

Tool still can't connect after import: first confirm the site has been applied to the tool in the apply center; then check the model name against your key group on this site's Pricing page. Restarting the CLI usually applies changes for tools other than Claude Code.

Sharing configs manually: hand-built xiaobaiswitch:// sites?...links also work — see the README link-import section for parameters.

Security: import links contain your API key. Do not copy, screenshot, or publicly share a complete link; import only from this site or another source you trust.

### Trusted sources and self-checks

This site's live model list and groups: confirm the model name and key group before applying the site.

XiaoBaiSwitch official README: features, installers, and xiaobaiswitch:// protocol parameters.

XiaoBaiSwitch Releases: installers and release notes per version.
