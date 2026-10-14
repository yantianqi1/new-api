---
title: "Error Codes & Troubleshooting"
---

When something fails, first find the request's full error in Usage Logs, then locate it in the table below. Most issues are self-solvable in two minutes.

### Three-step self-check

Verify inside the site: open Playgroundand message the same model. If it works there, the gateway is fine and the issue is in your client config.

Check the three essentials: Base URL (extra or missing /v1?), key (complete with the sk-prefix?), model name (exactly as on Pricing).

Read Usage Logs: every failed call's status code and error detail is recorded there.

### Error code reference

| Status | Typical error | Cause & fix |
| --- | --- | --- |
| 401 | `invalid_api_key` | Wrong/deleted key or missing `sk-` prefix; or the Base URL points at another site. Recreate the key and confirm `https://api.aiwanai.cc` |
| 403 | `Forbidden` | The key has model/IP restrictions, or its group can't access that model. Check key settings or create an unrestricted key |
| 404 | `model_not_found` | Misspelled model, model not in your group, or missing `/v1` in the path. Copy the exact name from Pricing |
| 408 | Request timeout | Raise the client timeout; use streaming; or pick a faster model |
| 413 | `context_length_exceeded` | Input or history exceeds the model's context. Trim history, lower `max_tokens`, or use a long-context model |
| 429 | `rate_limit_exceeded` | Concurrency/frequency limit hit. Reduce parallelism and add exponential-backoff retries |
| 429 | `insufficient_quota` | Balance or key quota exhausted. Top up in Wallet or raise the key quota |
| 500/502 | Upstream error | Temporary site/upstream fault — retry shortly; report if persistent |
| 503 | `no available channel` | No channel currently serves that model. Retry later, switch model or group; report the model + time if it repeats |

### Frequent cases

Claude Code can't connect

ANTHROPIC_BASE_URL must be https://api.aiwanai.cc — no trailing /v1

Restart the terminal after editing `settings.json`

See the Claude Code guide

Codex returns 401 or 404

base_url must be https://api.aiwanai.cc/v1 (with/v1) and wire_api = "responses"

The key lives in auth.json, not config.toml

See the Codex CLI guide

### Streams cut off midway

Raise client timeouts; add-Ntocurlto disable buffering

Self-hosted Nginx reverse proxies needproxy_buffering off;

### Content refused

Your input tripped content moderation. Rephrase it; jailbreak-style usage violates site rules and repeated triggers may restrict your account

### Still stuck?

Re-check your request parameters and the error-code explanations — confirm the Base URL, your key, and the group are correct

The AI assistant at the bottom-right of docs pages accepts pasted errors directly (if enabled on this site)
