---
title: "Platform Tour: Wallet / Logs / Security"
---

Beyond relaying API calls, the waner~ API站 console ships a full toolbox covering integration, spending, observability, and security. This is a scenario-by-scenario tour — some features are enabled per site, so if an entry is missing, the site's actual configuration prevails.

### Integrate & debug

Playground (sidebar Chat → Text Chat): pick a model and group and chat right on the site — verify your token and connectivity without installing anything; only the parameters you explicitly enable are sent with requests.

Image Generation (Chat → Image Generation): type a prompt and get images with a masonry gallery, lightbox preview, and one-click download; history is stored locally in your browser. API usage: image generation guide.

Docs center (top bar Docs): built-in bilingual integration guides; once signed in, pick one of your keys and every code sample is rewritten with your real base URL and key — copy and run; you can also ask AI questions about the current page.

CC Switch one-click import (API Keys → row actions): push the site address, key, and model mapping to your local CC Switch in one click, covering Claude Code, Codex CLI, and more — see the CC Switch guide.

Third-party apps (Chat → Third-party / key row actions): open site-curated third-party clients (web chat UIs, Immersive Translate, …) with your key attached — no manual copy-pasting of configs.

Advanced key options (API Keys): batch creation, expiry presets, quota caps, a model allowlist, and an IP allowlist (CIDR supported) to scope each key to the minimum it needs.

### Balance & perks

Wallet (Personal → Wallet): balance, online top-up (multiple payment methods), redemption codes, and billing history (searchable by order number) on one page.

Subscription plans (in Wallet, when enabled): periodic model quota bundles; pick your own deduction preference among subscription-first / wallet-first / subscription-only / wallet-only.

Gift credit batches (in Wallet): permanent balance and expiring credit are tracked separately, with each batch labeled by source (top-up bonus, referral reward, check-in reward, …) and expiry date.

Daily check-in (Profile, when enabled): calendar-style check-in with randomized credit rewards.

Activity / trial codes (in Wallet, when enabled): enter a code to temporarily unlock hidden groups (e.g. premium model pools) until it expires; qualifying top-ups may automatically grant you a personal trial code.

Referral rewards (in Wallet, when enabled): generate your invite link; rewards can be transferred into usable account credit.

### Usage & observability

Usage Logs (General → Usage Logs): separate tabs for regular, drawing, and async-task logs; drawing logs preview generated images inline; log details can show the full request content; live spend/RPM/TPM stats on top.

Performance board (General → Performance): per-group health and latency over 24h / 7d computed from real traffic, expandable to per-model detail — check it before choosing a group.

Model marketplace (top bar Models): every model's context window, capability tags, available groups, per-group prices (including tiered-pricing breakdowns), plus measured latency and success rate.

Rankings (top bar Hot): usage-based model leaderboard, vendor market share, and trending-up/down movers.

Overview dashboard (General → Overview): announcements, FAQ, upstream provider status, uptime monitoring, and the API endpoint card — pair it with the local speed test.

Notification center (top-bar bell): system notices and announcements tabs, with unread badges.

### Account & security

Passkey sign-in (Profile): log in with fingerprint / face / security key, no password typed.

Two-factor authentication (Profile): TOTP codes plus backup codes.

Login session management (Profile): see every active device and login method; revoke a single device or sign out all others.

Notification channels (Profile → Notifications): low-balance alerts and more via email / webhook / Bark / Gotify, with a dedicated notification address and threshold.

Account bindings (Profile): bind or unbind email, We Chat, Telegram, and OAuth providers such as GitHub.

Sidebar customization (Profile): hide menu entries you never use and keep your console lean.

### When something breaks

Start with Error codes & troubleshooting and work through the self-serve fixes.
