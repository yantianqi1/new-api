---
title: AQBot
---

AQBot is an open-source AI desktop client with multi-model chat, MCP servers, and knowledge bases. This site supports AQBot's aqbot:// providersimport protocol, so you can send the current site, endpoint, and API key to AQBot in one step.

### Official links

AQBot website: https://app.aqbot.top/

Downloads: https://app.aqbot.top/download

GitHub: https://github.com/AQBot-Desktop/AQBot

Import protocol: Importing from a Website Link

### One-click import

Install AQBot and launch it at least once so your system registers the aqbot:// protocol.

In this site's console sidebar, open Third-party → AQBot.

Select the API key to import. If this site publishes multiple endpoints, select the target endpoint too.

Your browser opens AQBot. Review the provider name, Base URL, type, and API key prefix, then confirm the import.

You can also open any key's action menu under Console → API Keys and choose Chat → AQBot to export that key directly.

The import link contains the complete API key. Do not copy, screenshot, or publicly share the full link, and approve imports only from this site or another source you trust.

### Manual setup

If the browser does not open AQBot, go to Settings → Providers in AQBot and add an OpenAI provider manually:

Name: waner~ API站

Base URL: https://api.aiwanai.cc

API Key: sk-your-api-key

After importing or saving, click Fetch Models, or add model IDs manually based on this site's Pricing page. AQBot does not validate the API key or sync models automatically during link import.

### Troubleshooting

AQBot does not open: make sure AQBot is installed and has been launched, then allow the browser to open the external application.

The endpoint is rejected: the Base URL must use http or https and cannot contain a query string or fragment.

No models appear after import: click Fetch Models in AQBot's provider settings, or add a model ID available to your key group.

Requests return 401/403: verify the API key and confirm that its group can access the selected model.
