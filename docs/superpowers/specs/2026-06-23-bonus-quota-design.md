# Bonus Quota Design

## Goal

Introduce a separate promotional points balance, displayed as "积分额度", so sign-in rewards can attract users without letting free rewards consume paid-only models.

The system will keep paid balance and points balance separate:

- Paid balance: recharge, redemption code, and admin paid top-up.
- Points balance: sign-in/check-in rewards.

Admins can choose which models allow points deduction. Points-eligible models consume points first and fall back to paid balance when points are insufficient. Other models never consume points.

## Admin Placement

The setting belongs in the model pricing area, not the sign-in settings area.

Reasoning:

- Sign-in settings decide how many points users receive.
- Model pricing decides how each model can be paid for.
- Keeping "积分抵扣" beside model price makes activity configuration model-centric and easier to audit.

Default frontend placement:

- `系统设置 -> 模型 -> Pricing Ratios / 模型定价`
- Add a visible "积分抵扣" signal in the model pricing table.
- Add an "允许积分抵扣" switch in the per-model pricing editor.
- Add a batch entry in the same pricing module for searchable multi-select enable/disable.

Classic frontend placement:

- `设置 -> 倍率/模型定价`
- Add the same per-model switch in the visual pricing editor.
- Keep manual JSON editing available for advanced operators.

## Data Model

Add `bonus_quota` to users.

- Type: integer, default `0`.
- Meaning: remaining promotional points quota.
- Existing `quota` remains paid balance.
- Existing `used_quota` continues to represent total consumed quota for aggregate reporting.

Add option-backed model eligibility setting:

- Key: `BonusQuotaModels`
- Value: JSON object keyed by model name, for example:

```json
{
  "gpt-4o-mini": true,
  "deepseek-chat": true
}
```

This avoids changing the meaning of existing `quota_type`, which currently means token-based versus per-request pricing.

## Billing Rules

For every request, compute price exactly as today. Only the funding split changes.

If the request model is not enabled in `BonusQuotaModels`:

- Pre-consume and settle only from paid `quota`.
- Ignore `bonus_quota` completely.

If the request model is enabled in `BonusQuotaModels`:

- Pre-consume from `bonus_quota` first.
- If points are insufficient, pre-consume the remaining amount from paid `quota`.
- If both balances together are insufficient, reject with insufficient quota.
- During settlement, adjust the same two balances based on actual cost.
- If estimated pre-consumption is higher than actual cost, refund paid quota first, then points. This keeps paid balance from being unnecessarily locked.
- If actual cost is higher than pre-consumption, consume remaining points first, then paid balance.

Token-level quota limits continue to apply to the total request cost. This feature only splits user wallet funding.

## Logs And Visibility

Usage logs should keep total quota unchanged and add split metadata:

- `bonus_quota`: amount consumed from points.
- `paid_quota`: amount consumed from paid balance.
- `bonus_quota_model`: whether this model allowed points deduction.

User-facing wallet/profile views should show both balances:

- "付费额度"
- "积分额度"

Check-in UI should describe rewards as points, not paid balance.

## Migration And Compatibility

Migration must work on SQLite, MySQL, and PostgreSQL.

- Add `bonus_quota` through GORM model migration and existing cross-database migration patterns.
- Existing users start with `0` points.
- Existing paid balances remain unchanged.
- Existing recharge, redemption, and payment behavior remains unchanged.

## Tests

Backend tests should cover billing invariants:

- Check-in increases `bonus_quota`, not paid `quota`.
- Redemption/top-up continue increasing paid `quota`.
- Non-points model consumes only paid quota.
- Points model consumes points first.
- Points model falls back to paid quota when points are insufficient.
- Insufficient combined quota rejects the request.
- Settlement refund returns paid quota before points when actual cost is lower than estimate.

Frontend verification should cover:

- Admin can see and edit points eligibility in model pricing.
- User wallet/profile displays both balances.
- Check-in success copy refers to points.

