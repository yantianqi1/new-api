---
title: Image API Skill Page (for AI assistants)
---

This page is the complete calling manual (Skill) for the image capabilities of waner~ API站, written for AI assistants and automation scripts: hand this whole page plus an API key from the image-generation group to your AI (Claude, ChatGPT, Cursor, …) and it can perform every text-to-image, image-to-image and async-task call correctly — no other docs needed.

One-line instruction for your AI: "Read the API manual below and generate/edit images for me with this key: sk-xxx"

### 0. Basics

| Item | Value |
| --- | --- |
| Base URL | `https://api.aiwanai.cc` |
| Auth header | `Authorization: Bearer sk-your-api-key` |
| Key source | Console → API Keys, choose the image-generation group |
| Response format | Images are always returned as a `url` field (even if you ask for `b64_json`); links stay valid for about 1 hour — download promptly to keep them |
| Billing | Per-call flat pricing; cost and image URL s are recorded in the usage log; failed requests are not billed

### 1. Endpoint decision table (start here)

| What you want | Endpoint |
| --- | --- |
| Text-to-image, normal size, done within ~30s | `POST /v1/images/generations` (sync) |
| Text-to-image at 2k/4k, high quality, or n>1 | `POST /v1/images/generations/async` + polling (recommended default) |
| Image-to-image / edits / style transfer (with reference images) | `POST /v1/images/edits` (multipart, sync) or `POST /v1/images/edits/async` |
| Check an async task | `GET /v1/images/tasks/{task_id}` |

### Hard rules

Never call gpt-image-* models through chat endpoints (`/v1/chat/completions`, `/v1/responses`) — you will get a 400 redirecting you to the endpoints on this page.

Sync calls go through a CDN with a ~100-second connection cap; any generation that may exceed 60 seconds (4k, high quality, n>1) should use the async endpoints, otherwise you may hit 524 timeouts.

Async submission returns immediately and holds no connection; each account may queue at most 20 tasks at once.

### 2. Model capability table

| Model | Notes | size options | quality | edits support |
| --- | --- | --- | --- | --- |
| `gpt-image-2` | Recommended default, 1024 tier | auto, 1024x1024, 1536x864, 864x1536 | auto/low/medium/high | ✅ (up to 4 reference images) |
| `gpt-image-2-2k` | 2k tier, prefer async | auto, 2048x2048, 2560x1440, 1440x2560 | same | ✅ |
| `gpt-image-2-4k` | 4k/UHD tier, async only | auto, 3840x2160, 2160x3840, 2880x2880 | same | ✅ |

Extras: the gpt-image-2 family also accepts free-form sizes (multiples of 16, total pixels ≤ 3840×2160, aspect ratio ≤ 3:1).nis capped at 4 per request; for the gpt-image-2 family n>1 fans out into parallel single-image requests server-side.

### 3. Sync text-to-image

```bash
curl https://api.aiwanai.cc/v1/images/generations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-your-api-key" \
  -d '{
    "model": "gpt-image-2",
    "prompt": "a shiba inu wearing an astronaut helmet, film grain",
    "size": "1024x1024",
    "quality": "medium",
    "n": 1,
    "response_format": "url"
  }'
```

### Success response

```json
{
  "created": 1710000000,
  "data": [
    { "url": "https://images.example.com/xxxx.png", "revised_prompt": "..." }
  ]
}
```

### 4. Async text-to-image (recommended)

### Step 1 — submit (body identical to the sync endpoint)

```bash
curl https://api.aiwanai.cc/v1/images/generations/async \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-your-api-key" \
  -d '{"model": "gpt-image-2-4k", "prompt": "cyberpunk city at night, neon rain", "size": "3840x2160", "quality": "high"}'
```

### Immediate response

```json
{ "task_id": "3f2b9c1e-....", "status": "queued" }
```

### Step 2 — poll (every 3–5 seconds, up to 10 minutes)

```bash
curl https://api.aiwanai.cc/v1/images/tasks/3f2b9c1e-.... \
  -H "Authorization: Bearer sk-your-api-key"
```

State machine: queued→running→succeeded or failed; there are no other values.

succeeded: the result field is the full sync-endpoint response (grabresult.data [0].url);

failed: error.message explains why; failed tasks are not billed (unless the message explicitly says the generation was billed — then the image URLs can be recovered from the usage log detail).

```json
{
  "task_id": "3f2b9c1e-....",
  "status": "succeeded",
  "progress": "100%",
  "model": "gpt-image-2-4k",
  "created_at": 1710000000,
  "finished_at": 1710000123,
  "result": { "created": 1710000123, "data": [{ "url": "https://..." }] }
}
```

### Python polling template (AIs: use as-is)

```python
import time, requests

BASE, KEY = "https://api.aiwanai.cc", "sk-your-api-key"
H = {"Authorization": f"Bearer {KEY}"}

task = requests.post(f"{BASE}/v1/images/generations/async", headers=H, json={
    "model": "gpt-image-2-4k", "prompt": "cyberpunk city at night", "quality": "high",
}, timeout=30).json()

deadline = time.time() + 600
while time.time() < deadline:
    r = requests.get(f"{BASE}/v1/images/tasks/{task['task_id']}", headers=H, timeout=30).json()
    if r["status"] == "succeeded":
        print(r["result"]["data"][0]["url"]); break
    if r["status"] == "failed":
        raise RuntimeError(r["error"]["message"])
    time.sleep(3)
```

### 5. Image-to-image / edits

Multipart form; pass up to 4 reference images as files via the image []field name:

```bash
curl https://api.aiwanai.cc/v1/images/edits \
  -H "Authorization: Bearer sk-your-api-key" \
  -F model="gpt-image-2" \
  -F prompt="repaint in watercolor style, keep the composition" \
  -F "image[]=@reference1.png" \
  -F "image[]=@reference2.png" \
  -F size="1024x1024" \
  -F response_format="url"
```

The response format is identical to text-to-image (data [].url).

For large references or high-resolution output use POST /v1/images/edits/async (same multipart body, returnstask_id, poll as above); async submissions are capped at 15MB total.

### 6. Error table

| Status / symptom | Meaning | What the AI should do |
| --- | --- | --- |
| 400 "image-generation-only model" | Image model sent to a chat endpoint | Switch to the images endpoints on this page |
| 400 missing model | No `model` in body | Add it and retry |
| 401 | Invalid/disabled key | Ask the user to check the key |
| 403 | Group has no access to this model | Use a key from the image-generation group, or switch models |
| 413 | Async body over 15MB | Compress references, or use the sync endpoint |
| 429 (chat/sync) | Rate limited | Wait 10–30s and retry |
| 429 "queued image tasks reached the limit of 20" | Async queue full | Wait for existing tasks to finish |
| 524 / timeout | Sync generation exceeded the CDN limit | Use the async endpoints (that is what they are for) |
| Task `failed` | See `error.message` | Handle per message; unbilled failures can simply be resubmitted |
| 5xx | Upstream turbulence | Back off 10–30s and retry once |

### 7. Checklist for AIs

You have a key from the image-generation group; Base URL is https://api.aiwanai.cc;

Pick the model: everyday →gpt-image-2; 2k/4k → the matching tier via async; edits → a model with edits support;

Never send gpt-image-* to chat/responses endpoints;

Async flow: submit → poll every 3s → on succeeded read result.data [*].url, on failed read error.message;

Image links expire in ~1 hour: download immediately when the user wants to keep them;

Failed requests are not billed; costs and past image URLs live in the usage log.
