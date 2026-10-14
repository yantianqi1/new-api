# Main customizations on latest upstream

User-approved approach: base the entire project on official upstream main, then move the production main branch's custom behavior onto that architecture.

- Upstream base: 7de10f2c602b50122017556e8a09cb7c488c273c
- Production source: 67a1872575c3a566ed32c7870d5f9ffe093c772f
- Customization ancestor: 6dc4030fdf5070815941a0cd777e986bbc33a6c6

Preserve latest authentication, pricing APIs, relaykit/tokenkit boundaries, plugins and upstream fixes. No legacy authentication bridge or fake compatibility endpoint. No deployment or production writes.

Migration inventory:

- [x] Paid/bonus wallet accounting, free-model whitelist, check-in and invitation bonus credits; task/refund/log integrations.
- [x] Multi-use paid/bonus redemption codes and legacy backfill.
- [x] Affiliate commissions and transfers, payment-provider snapshots, legacy pending settlement.
- [x] Promotional recharge tiers, purchase/stock limits, bonus plus base quota, bargain/referral packages.
- [x] Per-user/per-model free-credit RPM and per-channel RPM policies.
- [x] Personalized home, activity marquee, custom model-square links (including HeaderNavModules.pricing.link), model marketplace synchronization, separate paid/points overview and wallet displays, custom admin views, translations.
- [x] OIDC provider and client management with latest session/auth security.
- [x] SFT collector, model marketplace custom controls and privacy/redaction behavior; review provider changes against latest plugins.
- [x] Fresh/production-schema upgrade/idempotency verification with real SQLite, MySQL and PostgreSQL, including separately configured log DB where affected.
- [x] Frontend typecheck/test/lint/build; relevant backend regression tests; independent modules build; independent migration review.

## Observed verification (2026-10-14)

- Root: `go test -p 2 ./...` passed. After the final configuration fix, `go test -p 2 ./setting/config ./setting/operation_setting ./controller ./model ./service` passed again.
- Independent modules: `GOWORK=off go build ./...` passed inside both `relaykit/` and `tokenkit/`; `GOWORK=off go test -p 2 ./...` passed inside `relaykit/`; `GOWORK=off go test ./...` passed inside `sft-collector/`.
- Frontend: frozen upstream `web/bun.lock` dependencies, `bun run typecheck` and `bun run build` passed. `bun run test -- --maxWorkers=4` passed: 203 test files, 2308 tests. All 137 changed frontend paths pass lint with zero errors; two existing prefer-query-selector warnings remain in the touched setup-guide test. The subsequent scoped code-sample color correction was checked with the home regression and a new production build.
- Full frontend lint has pre-existing upstream errors. Pristine upstream baseline: 117 unique diagnostics; migration: 116, with no newly introduced diagnostics at that comparison. This is not a claim that repository-wide lint passes.
- Existing standalone `docs-site/` was preserved verbatim from production main, including package files. `bun run build` passed: 39 pages. Its pre-existing npm lock omits declared Wrangler development dependencies, so `npm ci` fails; validation installed production dependencies with `npm install --omit=dev --ignore-scripts --package-lock=false`, without editing the manifests/lock.
- Independent review confirmed and then verified fixes for points indicators in pricing, source-specific log costs, mixed wallet reservations with paid debt, and replacement of slice configuration with omitted optional fields. Final review: no remaining confirmed P0–P2 findings. A second independent OIDC grant review also found none; TestOIDCGrant passed with -race, along with all TestOIDC and focused PAT route tests.
- Real browser checks against an isolated test MySQL app: personalized desktop/mobile home, latest session login, paid/points wallet, authorization scope reduction and revocation, the model marketplace sync entry, and root OIDC client management; no uncaught browser errors. Home visual styling and code-sample colors were inspected against the built frontend.

## Real database matrix

Fixtures were initialized with the unchanged production-main backend at `67a1872575c3a566ed32c7870d5f9ffe093c772f`, including a user with paid/bonus/commission balances, an already-used multi-use redemption code and custom options. Production data and containers were not used.

| Engine | Exact version | Fresh | Production-main upgrade | Repeat startup/migration | Separate log database |
| --- | --- | --- | --- | --- | --- |
| SQLite | 3.50.4 | Passed | Passed | Passed, twice | Passed |
| MySQL | 8.4.11 | Passed | Passed | Passed, twice | Passed |
| PostgreSQL | 16.15 | Passed | Passed | Passed, twice | Passed |

Each run used `NEW_API_MIGRATION_MATRIX=1`, and upgrade runs also used `NEW_API_MIGRATION_UPGRADE=1`, with `go test -p 2 ./model -run '^TestCustomMigrationDatabase$' -count=1 -v` and an isolated `SQL_DSN` / `SQLITE_PATH`. Separate-log fresh runs additionally set `LOG_SQL_DSN`. The fixture test exercises real migrations, preservation of paid/bonus/commission values and options, legacy redemption claimant backfill, unique constraints, mixed reserve/refund, negative paid balance, paid-only rejection, bonus redemption and daily check-in, idempotent commission settlement/transfer, one-stock promotional tier limits, referral-cut replay rejection, and OIDC grant listing, reduction, stale-scope exchange prevention, cross-session revocation and expired/revoked session rejection. Minimum engine versions and ClickHouse were not exercised; no new driver/dialect dependencies were introduced.

Logs and temporary fixtures: `/tmp/new-api-full-migration/matrix-{mysql,postgres,sqlite}-{fresh,upgrade}.log`. The isolated test containers have names `new-api-migration-test-mysql` and `new-api-migration-test-postgres`.

## Authentication transition

The migration keeps the latest JWT/browser-session architecture. Legacy login cookies and old OIDC tokens without a current browser session cannot be reused; users sign in again. Existing accounts and OIDC client credential hashes are preserved. OIDC clients require S256 PKCE, exact allowlisted redirects, persisted session-bound single-use consent, code replay revocation, current session/account/client checks, and root step-up verification for client mutations. Secrets are returned once and excluded from list responses/audits. The configured `ServerAddress` is the issuer; Host headers cannot supply it.

References: [OWASP ASVS 5.0.0](https://owasp.org/www-project-application-security-verification-standard/), [Authentication](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html), [Session Management](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html), [OAuth2](https://cheatsheetseries.owasp.org/cheatsheets/OAuth2_Cheat_Sheet.html), [CSRF](https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html). Relevant OIDC controls reviewed: 10.4.1–3, 10.4.6, 10.4.10–11, 10.7.1–3, 10.4.9, 7.2.1, 7.4.1–2. Users can review their authorized applications in Security & Access, remove profile/email scopes, and revoke pending authorizations and their issued reference tokens from any active browser session. Ownership, session validity, scope subsets, PAT rejection and concurrent code-exchange fencing are enforced server-side and covered by regression tests. This is a scoped verification, not full-application ASVS certification. Before a future production deployment, configure the latest session requirements (`SESSION_COOKIE_SECURE=true`, `SESSION_COOKIE_TRUSTED_URL=https://api.aiwanai.cc`) and the actual issuer `ServerAddress`, then validate production cookie/login behavior.

## Architecture and reuse

Latest pricing APIs, provider plugins, dashboard authentication, audit/system/task pages, and standalone conversion/counting modules remain the foundation. Current official Hailuo plugin already covers MiniMax-H3 and v2 requests; the obsolete Go adaptor was not restored. Unsupported Gemini media errors retain the production privacy/400/no-retry behavior in both latest conversion entry points. The custom dual-old/new frontend build was replaced by the latest single frontend.

Custom features compose existing business components: settings sections, secure verification, confirmation dialogs, data tables, copy controls, StatCard, TitledCard, StaticDataTable, CodeBlock, Tabs, the latest cost display/billing breakdown, and pricing badges. Feature-local home presentation, promotion schemas/cards, and OIDC consent/client forms provide the custom business capabilities absent upstream; shared interactions are reused.

No commit on production `main`, push, production query/write, PR, or deployment has been performed by this migration.
