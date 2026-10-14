---
title: Image Generation (URL, not base64)
---

When you generate images on waner~ API站, you get back an image URL instead of a large base64 blob. That is not a missing feature — it is a deliberate design choice. This guide explains why, and what to keep in mind when integrating.

### Why a URL instead of base64?

base64 "transcribes" the image into plain text inside the JSON response. A single high-resolution image becomes 5–20MB of text, which causes three real problems:

Slow: nothing renders until the entire response finishes downloading — on an average connection that is many seconds of staring at a blank screen;

Laggy: many clients stutter (or truncate) when rendering huge base64 strings;

Fragile: oversized responses are far more likely to hit timeouts along the way.

With a URL, the API response is only a few hundred bytes and arrives instantly; the image itself lives in object storage and downloads through a CDN — fast and reliable. This is how mainstream image APIs work (including OpenAI's own DALL·E URL mode).

So on waner~ API站: image endpoints always return images in the `url` field. Even if you send `response_format`: "`b64_json`", you will receive a url — that is not a bug, it is platform policy, so that image generation stays fast for everyone.

### Quick start

```bash
curl https://api.aiwanai.cc/v1/images/generations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-your-api-key" \
  -d '{
    "model": "gpt-image-2",
    "prompt": "A shiba inu wearing an astronaut helmet, film grain",
    "size": "1024x1024",
    "response_format": "url"
  }'
```

A successful response looks like:

```json
{
  "created": 1710000000,
  "data": [
    {
      "url": "https://images.example.com/xxxx.png",
      "revised_prompt": "A shiba inu wearing an astronaut helmet..."
    }
  ]
}
```

Open the `url` in a browser, reference it from an<img>tag, or download it — your choice.

### Image links expire after about 1 hour

Links stay accessible for about 1 hour after they are issued; the image is then cleaned up from storage. This keeps storage costs sane and avoids retaining your generated content indefinitely.

If you need an image long-term, download it promptly. For example, in Python:

```python
import requests

url = resp.json()["data"][0]["url"]
with open("image.png", "wb") as f:
    f.write(requests.get(url, timeout=60).content)
```

On the console's Image Generation page, every unkept image shows its remaining time and offers a"Keep locally"button: the image bytes are then stored only in your browser's local storage (never uploaded to our server) so the image stays in your history permanently. Images you do not keep disappear from history automatically when the link expires.

### About streaming (stream)

Image endpoints do not need streaming. An image is produced in one shot — there is no "token by token" output:

Without stream: you get the plain JSON above. This is the recommended way;

With stream: true: the request still works — you receive one SSE completion event carrying the finalurl, just without intermediate preview frames.

In short: ignore streaming and write a plain JSON request.

### My client does not display the image

Most mainstream clients (and the official OpenAI SDKs) handle the `url` field correctly. If your client receives a response but shows no image, it most likely only reads the b64_json field:

Prefer switching to (or upgrading to) a client that supports the `url` return format;

If it is your own app, download the image first as shown above — a few lines of code makes it compatible.

### Content policy

Please do not use image generation for illegal or abusive content. Such requests are blocked and logged, and repeated violations may affect your account.

### Next steps

No API key yet? See Quick start

Getting errors? See Error codes & troubleshooting

Full parameter reference lives in the Images group of the API reference
