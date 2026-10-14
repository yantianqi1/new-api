---
title: Cherry Studio
---

Cherry Studio is a popular open-source desktop AI client (Windows / macOS / Linux) with multi-provider support, knowledge bases, and image generation. Add waner~ API站 as an OpenAI-type provider to use every model here.

### Step 1: Install

Download from the official site: https://www.cherry-ai.com

### Step 2: Add a provider

open Settings → Model Servicesand click "Add Provider"

Any name works (e.g. waner~ API站); choose the OpenAI type

Fill in:

API Key: sk-your-api-key

API Host: https://api.aiwanai.cc

The site root is enough — Cherry Studio appends /v1 automatically. If you enter a full path, end it with#to force the exact URL.

### Step 3: Add models

Click "Manage" in the provider's model section; Cherry Studio fetches this site's model list automatically. Tick the models you need (see the Pricing page).

### Step 4: Verify

Back in a chat, switch the model picker to one of the new models and send a message. A reply means success, and the call shows up in Usage Logs.

### Troubleshooting

Model list won't load: check that the host is https://api.aiwanai.cc and the key is valid.

404 during chat: the model isn't in your key's group — pick another model or adjust the key.

Prefer one-click setup: if the admin configured chat presets, the console sidebar's chat entries can deep-link Cherry Studio with your key pre-filled.
