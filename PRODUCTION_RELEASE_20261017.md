# New API production release — 2026-10-17

## Scope and source

User requested finalization of the upstream/custom migration, creation of a production branch, and updating the production website.

- Release branch: `codex/production-release-20261017`.
- Starting migration commit: `4b3785f1bdc9b865617eb417dd001bf141c5a404`.
- Custom production source: `67a1872575c3a566ed32c7870d5f9ffe093c772f` (`main`).
- Fetched official upstream: `3e2ce2c88` (21 commits after the migration base).
- Existing activity-bar changes are carried forward without discarding user work.
- Production: `hly-HK-44` / `154.201.69.160`, `/opt/new-api`, MySQL and Redis.
- Observed current production image: `new-api:reliability-tierfix-20261008`.

## Execution plan

- [x] Verify worktree, upstream, fork, production host and running image.
- [x] Run activity-bar regression, typecheck and affected lint; commit existing verified fix.
- [x] Merge the fetched official upstream and resolve only integration conflicts.
- [x] Review custom-feature preservation and upstream changes in billing, plugins, rate limits and security.
- [ ] Run frontend tests/typecheck/build, backend tests/build, independent module builds/tests and affected lint/format checks.
- [x] Run fresh and production-main upgrade/idempotency checks on real SQLite, MySQL and PostgreSQL, including separate log DB.
- [ ] Rehearse a fresh production snapshot in an isolated database; compare original row hashes, balances, constraints and read APIs.
- [ ] Build a versioned linux/amd64 container and test it against isolated data/Redis; verify frontend assets and authentication behavior.
- [ ] Review the final diff and document material gaps before release.
- [ ] Prepare server-only backup and concrete rollback commands, stage the image, then switch production and verify public website/API behavior.

## Release constraints

Preserve existing custom wallet, bonus/RPM, redemption, commission, promotions, model marketplace/home and OIDC/SFT behavior. Do not reset production or local data, overwrite the main checkout's dirty files, expose secrets, or create documentation under `docs/` or plugin directories. Any failure affecting billing, data preservation or authentication blocks deployment until resolved. Health/status alone does not prove relay, payment or browser success.

Authentication review uses OWASP ASVS 5.0.0 and the Authentication, Session Management, OAuth2 and CSRF Cheat Sheets. Verify secure cookies, trusted public URL, current session validity, PKCE, expiry/replay protection and scoped management-token behavior. Existing browser sessions require re-login; legacy management-token transition and external integrations must be checked before release.

## Observed validation and deployment

Historical results in `CUSTOM_MIGRATION_PLAN.md` are context, not fresh release evidence. Current private execution logs are in `/tmp/new-api-release-20261017` (mode 700); no credentials or production record bodies belong in this document.

- `bun run test -- --maxWorkers=4` with bundled Node 24.19.0: 212 files / 2368 tests passed.
- `bun run typecheck`: passed. Custom frontend lint: 132 files, zero errors, two existing setup-guide test warnings. Protected-header-safe format check on 135 custom code/style paths: passed. Repository-wide format check has pre-existing unrelated failures; it is not claimed clean.
- Plugin lint on the four merged upstream plugins: zero errors, two upstream caught-error warnings. Numeric usage descriptions were reviewed for subject/unit-price wording.
- `GOWORK=off go build ./...` in `relaykit/` and `tokenkit/`: passed. `GOWORK=off go test -p 2 ./...` in `relaykit/`: passed.
- `python3 /tmp/new-api-release-20261017/run-matrix.py`: six fresh/production-main-upgrade runs passed on SQLite 3.50.4, MySQL 8.4.11 and PostgreSQL 16.15. Each runs `NEW_API_MIGRATION_MATRIX=1 go test -p 2 ./model -run '^TestCustomMigrationDatabase$' -count=1 -v`, with `NEW_API_MIGRATION_UPGRADE=1` for the old-main fixtures, repeated migrations and separately configured log databases. Only isolated `release_20261017_*` databases are used. Minimum supported database versions and ClickHouse were not tested; no database driver dependency changes were introduced.
- Found and reproduced a release-blocking bonus-wallet settlement error: bonus 100 / paid 0 / reservation 60 / final 120 rejected the final difference after usage was delivered. Commit `46c7cf336` separates strict request admission from completed-usage settlement, keeps transaction/CAS/cache compensation, consumes remaining bonus and records paid shortfall. Failing tests were observed before repair; the new real three-database matrix passed synchronous/task settlement with Redis on/off, strict admission, replay protection, cache consistency and debt limits. `TEST_BONUS_SETTLEMENT_mysql_DSN` and `TEST_BONUS_SETTLEMENT_postgres_DSN` point to isolated databases; command: `go test -p 2 ./service -run '^TestBonusWalletSettlementDatabaseMatrix$' -count=1 -v`.
- Independent read-only review found the settlement blocker and subsequently reviewed the four-file fix with no remaining confirmed P0–P2 findings. This was a focused review, not a certification of the whole application.
- Production at snapshot: MySQL 8.4.10, 2492 users, 2151 API tokens, 884681 log records, one OIDC client and 169 legacy management tokens. `ServerAddress` already equals `https://api.aiwanai.cc`; secure-cookie environment variables are absent and must be explicitly configured at rollout.

## Authentication and external integration transition

The current local `newapi-lottery` client uses `/api/user/self`, `/api/user/:id`, `/api/user/topup` and `POST /api/redemption` with management access tokens. The new gateway retains existing management tokens for the server-managed 30-day transition. Their retirement date is first persisted at migration. The worker and users must switch to scoped personal access tokens before that date (`profile:read` for user identity; administrative user/billing/redemption permissions for the worker). API `sk-` tokens are a separate mechanism. Do not weaken session/security requirements or extend the retirement deadline to disguise an unmigrated integration.

Payment callbacks, an actual lottery award, OIDC client login through the external site, and live model generation require distinct runtime evidence. Do not infer those from status or model-list success. Production backup, image staging, rollout and public validation remain pending.
