---
title: Immersive Translate
---

Immersive Translate is a widely used bilingual web-translation browser extension. Backed by waner~ API站, it produces translations far beyond ordinary machine translation.

### Step 1: Install the extension

Install from the official site: https://Immersive Translate.com

### Step 2: Configure a custom API

Click the extension icon →Settings→Translation Services

Open OpenAI (or "Custom AI translation service")

Fill in:

APIKEY: sk-your-api-key

Model: choose "Custom" and enter a model name (a light, fast model such as gpt-4o-miniis ideal)

Custom API endpoint: https://api.aiwanai.cc/v1/chat/completions

### Step 3: Tune concurrency and cost

Whole-page translation fires many small requests, so:

Max requests per second: lower it to 5–10 to avoid rate limits (429)

Max paragraphs per request: raising it reduces request count and cost

Pick a cheap model — translation doesn't need a frontier model (prices on the Pricing page)

### Step 4: Verify

Open any foreign-language page and enable translation from the extension icon. Translated text appearing means success; details are in Usage Logs.

### Troubleshooting

401: is the APIKEY complete (starts with `sk-`)?

Frequent 429: lower max requests per second, or see Error codes & troubleshooting.

Empty translations: try another model; a few models handle translation prompts poorly.
