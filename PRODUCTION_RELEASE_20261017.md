# New API production release — 2026-10-17

## Scope and source

User requested finalization of the upstream/custom migration and creation of a production branch. The later instruction explicitly defers production replacement: run a second independent instance, expose it by IP without a domain, and copy current production data into its own database.

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
- [x] Prepare server-only backup and rollback commands; stage the image without changing production.
- [ ] Start the independent preview stack at `http://154.201.69.160:3002`, copy production into its separate MySQL, migrate the copy, and verify IP access and original-stack isolation.

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

Payment callbacks, an actual lottery award, OIDC client login through the external site, and live model generation require distinct runtime evidence. Do not infer those from status or model-list success.

## Independent second-site rollout

The original `new-api` container remains on `new-api:reliability-tierfix-20261008`, port 3000, original MySQL/Redis, `/opt/new-api`, and its existing domain. Original `.env` and compose files were compared byte-for-byte with the pre-release backup and are unchanged. A production-switch script was staged but never executed; its staged scripts/configuration were withdrawn to a `deferred-switch/` backup subdirectory after the user's second-site instruction.

Candidate image `new-api:yantianqi1-release-20261017-8c5aa8417` is linux/amd64, image ID `sha256:ce7217a71fde4577555931958a18e4110fa1348af76a8bd7ecf6a445005964fa`, source revision `8c5aa84174bc5dbd5ba7f54510c41cb71ecf0072`, runtime version `release-20261017-46c7cf336`. It was built locally, transferred, and its ID/revision verified on the server. The isolated candidate container passed 13 login/cookie/origin/refresh/logout/management checks. Obscura loaded its built frontend. Focused auth/billing race tests, tokenkit tests and SFT collector tests passed.

The earlier production snapshot rehearsal preserved all 1,091,516 original rows by per-row hashes (only two option additions), all 62 original unique-constraint signatures, and identical schema on repeated migration. Four unique indexes were renamed by GORM; their column/uniqueness signatures were preserved. Existing API key model-list checks and old management-token reads passed; lottery-equivalent user lookup and redemption creation passed in the isolated copy only.

The checked-in `docker-compose.preview.yml` records the deployed second-site stack. The second site lives at `/opt/new-api-preview`, Compose project `new-api-preview`, containers `new-api-preview`, `new-api-preview-mysql`, `new-api-preview-redis`, and network `new-api-preview-network`. MySQL data, app files, logs and fresh cache/session secrets are independent. Only port 3002 is published. No proxy, certificate or domain change is made. `NODE_TYPE=slave`, `UPDATE_TASK=false` and disabled scheduled model updates prevent copied queued jobs from executing in parallel with production. The preview is limited to 0.5 CPU/256 MiB for the app, 0.5 CPU/384 MiB for MySQL, and 0.1 CPU/64 MiB for Redis to constrain its impact on the shared host. SQL pools are limited to 15 open / 5 idle connections. For the explicitly requested HTTP IP preview, secure cookies are disabled; TLS-only OIDC/passkey behavior requires a later HTTPS deployment.

An explicitly authorized fresh single-transaction production dump is copied into the new MySQL only. The preview's `ServerAddress` alone is adjusted to its IP URL. The original database, running container and domain are not switched or migrated.
