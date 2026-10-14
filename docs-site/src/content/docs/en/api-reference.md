---
title: API Reference
---

The waner~ API站 API is compatible with the official OpenAI / Anthropic / Gemini protocols and uses Bearer tokens — one `sk-` key works across every endpoint.

### Main endpoints

- **`POST /v1/chat/completions`** — generates a model reply from a conversation history; OpenAI Chat Completions compatible, with streaming when the model supports it. This is the primary endpoint for OpenAI-compatible clients; streamed responses use SSE, so the client must handle incremental chunks and the final usage; when forwarding, the backend preserves explicit zero and `false` values the client sends.
- **`POST /v1/responses`** — the OpenAI Responses protocol route; `/v1/responses/compact` is its slimmed-down variant.
- **`POST /v1/messages`** — the Anthropic-compatible message route.
- **`POST /v1 beta/models/{model}: generateContent`** — the Gemini-compatible route.
- **`POST /v1/images/generations`** — image generation; always returns an image URL rather than base64, see [Image generation](/en/image-generation/).
- **`POST /v1/embeddings`** / **`POST /v1/rerank`** — embeddings and reranking.
- **`GET /v1/models`** — the list of currently available models.

**Full endpoint list:**

| Type | Method | Path |
| --- | --- | --- |
| models | GET | `/v1/models` |
| anthropic | POST | `/v1/messages` |
| embeddings | POST | `/v1/embeddings` |
| gemini | POST | `/v1 beta/models/{model}: generateContent` |
| image-generation | POST | `/v1/images/generations` |
| jina-rerank | POST | `/v1/rerank` |
| openai | POST | `/v1/chat/completions` |
| openai-response | POST | `/v1/responses` |
| openai-response-compact | POST | `/v1/responses/compact` |

**Compatibility notes:**
- `/v1/chat/completions` follows OpenAI-compatible chat-completion semantics.
- `/v1/responses` follows OpenAI-compatible responses semantics where configured.
- `/v1/messages` provides the Anthropic-compatible message route.
- `/v1 beta/models/{model}: generate Content` provides the Gemini-compatible route.
- The models actually available to you depend on the enabled channels, your group, token restrictions and billing configuration.
