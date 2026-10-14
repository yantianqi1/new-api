---
title: "Endpoints & Local Speed Test"
---

waner~ API站 usually offers several endpoints (different API domains), and the same key works on all of them. The "API Info" card on Console → Overview lists every endpoint, and the local speed test helps you find the fastest, most reliable one from your own network.

### Why the test runs on your machine

Endpoint performance depends on where you are — region, ISP, proxy setup — so a test run on the site's servers would be meaningless. Testing directly from the browser is unreliable too: extensions and proxy rules routinely interfere with the numbers.

So the test runs on your computer: the page generates a PowerShell script, you run it locally, then import the result back into the page. Results stay in your browser only — nothing is uploaded to the server.

### How to use it

Open Console → Overview and click Local speed test in the corner of the "API Info" card.

In the dialog, copy the script (or download it as a .ps1 file). You can expand and audit it: it only sends GET requests to /api/ping on the configured endpoints, downloads and executes nothing, and refuses to probe private/internal addresses.

Open PowerShell, paste the script, and hit Enter. The PowerShell built into Windows (5.1+) is enough; on macOS / Linux install PowerShell 7 first (if the result can't be written to the clipboard automatically, copy it from the terminal output).

When it finishes, the result is copied to your clipboard. Back on the page, click Import result from clipboard, or paste it manually.

### Reading the results

Each endpoint is probed 3 times with a 6-second timeout; latency is the median of the successful runs. After importing, the list re-sorts by reliability → success rate → latency, and a banner at the top names the recommended endpoint.

| Badge | Meaning |
| --- | --- |
| Green 3/3 · 128ms | All 3 probes passed — stable |
| Yellow 1/3 · … | Partially passed — intermittent |
| Blue "reachable, unverified" | Connects, but the response is not this site's gateway (possibly intercepted by a middlebox) — use with caution |
| Red "connection failed" | Unreachable |

Hover the result badge for the latency range and the concrete failure reason (timeout, DNS failure, TLS handshake failure, and so on).

### Good to know

Results are valid for 10 minutes, then cleared automatically; they also expire if the site changes its endpoint list — just re-run the script.

The test measures gateway connectivity and latency only, not model requests — generation speed also depends on the upstream provider.

The "external speed test" button on each row opens a third-party multi-region tester as a complementary reference.

Once you've picked an endpoint, hitcopyand replace the API base URL in your client — the key stays the same.

### Related reading

First-time setup: Quick start

Connection errors: Error codes & troubleshooting
