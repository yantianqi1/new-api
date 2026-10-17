package model

import (
	"context"
	"errors"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

type cacheQuotaResult int

const (
	cacheQuotaInsufficient cacheQuotaResult = iota
	cacheQuotaOK
	cacheQuotaMiss
)

const userQuotaReserveScript = `
if tonumber(redis.call('HGET', KEYS[1], 'Id') or '0') ~= tonumber(ARGV[2])
  or tonumber(redis.call('HGET', KEYS[1], 'CacheSchema') or '0') ~= tonumber(ARGV[3])
  or redis.call('HEXISTS', KEYS[1], 'Quota') == 0 then
  return -1
end
local quota = tonumber(redis.call('HGET', KEYS[1], 'Quota'))
if quota == nil or quota < tonumber(ARGV[1]) then
  return 0
end
redis.call('HINCRBY', KEYS[1], 'Quota', -tonumber(ARGV[1]))
return 1`

const userQuotaDeltaScript = `
if tonumber(redis.call('HGET', KEYS[1], 'Id') or '0') ~= tonumber(ARGV[2])
  or tonumber(redis.call('HGET', KEYS[1], 'CacheSchema') or '0') ~= tonumber(ARGV[3])
  or redis.call('HEXISTS', KEYS[1], 'Quota') == 0 then
  return -1
end
redis.call('HINCRBY', KEYS[1], 'Quota', ARGV[1])
return 1`

const tokenQuotaReserveScript = `
if tonumber(redis.call('HGET', KEYS[1], 'Id') or '0') ~= tonumber(ARGV[2])
  or redis.call('HEXISTS', KEYS[1], 'RemainQuota') == 0
  or redis.call('HEXISTS', KEYS[1], 'UsedQuota') == 0 then
  return -1
end
local remain = tonumber(redis.call('HGET', KEYS[1], 'RemainQuota'))
if remain == nil or remain < tonumber(ARGV[1]) then
  return 0
end
redis.call('HINCRBY', KEYS[1], 'RemainQuota', -tonumber(ARGV[1]))
redis.call('HINCRBY', KEYS[1], 'UsedQuota', tonumber(ARGV[1]))
redis.call('HSET', KEYS[1], 'AccessedTime', ARGV[3])
return 1`

const tokenQuotaDeltaScript = `
if tonumber(redis.call('HGET', KEYS[1], 'Id') or '0') ~= tonumber(ARGV[2])
  or redis.call('HEXISTS', KEYS[1], 'RemainQuota') == 0
  or redis.call('HEXISTS', KEYS[1], 'UsedQuota') == 0 then
  return -1
end
redis.call('HINCRBY', KEYS[1], 'RemainQuota', tonumber(ARGV[1]))
redis.call('HINCRBY', KEYS[1], 'UsedQuota', -tonumber(ARGV[1]))
redis.call('HSET', KEYS[1], 'AccessedTime', ARGV[3])
return 1`

func quotaResultFromLua(result int, err error) (cacheQuotaResult, error) {
	if err != nil {
		return cacheQuotaMiss, err
	}
	switch result {
	case 1:
		return cacheQuotaOK, nil
	case 0:
		return cacheQuotaInsufficient, nil
	default:
		return cacheQuotaMiss, nil
	}
}

func cacheTryReserveUserQuota(userID int, amount int64) (cacheQuotaResult, error) {
	result, err := common.RDB.Eval(context.Background(), userQuotaReserveScript,
		[]string{getUserCacheKey(userID)}, amount, userID, userCacheSchemaVersion).Int()
	return quotaResultFromLua(result, err)
}

func cacheApplyUserQuotaDelta(userID int, delta int64) (cacheQuotaResult, error) {
	result, err := common.RDB.Eval(context.Background(), userQuotaDeltaScript,
		[]string{getUserCacheKey(userID)}, delta, userID, userCacheSchemaVersion).Int()
	return quotaResultFromLua(result, err)
}

func cacheTryReserveTokenQuota(id int, key string, amount int64) (cacheQuotaResult, error) {
	result, err := common.RDB.Eval(context.Background(), tokenQuotaReserveScript,
		[]string{getTokenCacheKey(key)}, amount, id, common.GetTimestamp()).Int()
	return quotaResultFromLua(result, err)
}

func cacheApplyTokenQuotaDelta(id int, key string, delta int64) (cacheQuotaResult, error) {
	result, err := common.RDB.Eval(context.Background(), tokenQuotaDeltaScript,
		[]string{getTokenCacheKey(key)}, delta, id, common.GetTimestamp()).Int()
	return quotaResultFromLua(result, err)
}

// persistUserQuotaDelta 把已在缓存侧预扣成功的增量落库；批量模式下入队，
// 直写模式下要求行存在（用户已删除时报错，交由调用方补偿缓存）。
func persistUserQuotaDelta(id int, delta int) error {
	if common.BatchUpdateEnabled {
		addNewRecord(BatchUpdateTypeUserQuota, id, delta)
		return nil
	}
	result := DB.Model(&User{}).Where("id = ?", id).Update("quota", gorm.Expr("quota + ?", delta))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func persistTokenQuotaDelta(id int, delta int) error {
	if common.BatchUpdateEnabled {
		addNewRecord(BatchUpdateTypeTokenQuota, id, delta)
		return nil
	}
	result := DB.Model(&Token{}).Where("id = ?", id).Updates(
		map[string]any{
			"remain_quota":  gorm.Expr("remain_quota + ?", delta),
			"used_quota":    gorm.Expr("used_quota - ?", delta),
			"accessed_time": common.GetTimestamp(),
		},
	)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func reserveUserQuotaDB(id int, quota int) (bool, error) {
	result := DB.Model(&User{}).
		Where("id = ? AND quota >= ?", id, quota).
		Update("quota", gorm.Expr("quota - ?", quota))
	return result.RowsAffected == 1, result.Error
}

func reserveTokenQuotaDB(id int, quota int) (bool, error) {
	result := DB.Model(&Token{}).
		Where("id = ? AND remain_quota >= ?", id, quota).
		Updates(map[string]any{
			"remain_quota":  gorm.Expr("remain_quota - ?", quota),
			"used_quota":    gorm.Expr("used_quota + ?", quota),
			"accessed_time": common.GetTimestamp(),
		})
	return result.RowsAffected == 1, result.Error
}

// TryReserveUserQuota atomically checks and deducts a user's wallet quota.
// 缓存命中时以缓存余额为准（避免批量模式下过期的数据库余额放大并发超扣）；
// Redis 异常或水合失败时降级为数据库条件更新，保证服务可用。
func TryReserveUserQuota(id int, quota int) (bool, error) {
	if quota < 0 {
		return false, errors.New("quota cannot be negative")
	}
	if quota == 0 {
		return true, nil
	}
	if !common.RedisEnabled {
		return reserveUserQuotaDB(id, quota)
	}

	result, err := cacheTryReserveUserQuota(id, int64(quota))
	if err == nil && result == cacheQuotaMiss {
		if _, hydrateErr := GetUserCache(id); hydrateErr == nil {
			result, err = cacheTryReserveUserQuota(id, int64(quota))
		}
	}
	if err != nil || result == cacheQuotaMiss {
		if err != nil {
			common.SysLog(common.LogText("user quota cache reserve unavailable, falling back to database: %s", err.Error()))
		}
		return reserveUserQuotaDB(id, quota)
	}
	if result == cacheQuotaInsufficient {
		return false, nil
	}
	if err = persistUserQuotaDelta(id, -quota); err != nil {
		compensated, compensateErr := cacheApplyUserQuotaDelta(id, int64(quota))
		if compensateErr != nil || compensated != cacheQuotaOK {
			common.SysError(common.LogText("failed to compensate reserved user quota: result=%d error=%v", compensated, compensateErr))
		}
		return false, err
	}
	return true, nil
}

// TryReserveTokenQuota atomically checks and deducts a token quota. Unlimited
// tokens skip the balance check but still update remain/used accounting.
func TryReserveTokenQuota(id int, key string, quota int, unlimited bool) (bool, error) {
	if quota < 0 {
		return false, errors.New("quota cannot be negative")
	}
	if quota == 0 {
		return true, nil
	}
	if unlimited {
		return true, DecreaseTokenQuota(id, key, quota)
	}
	if !common.RedisEnabled {
		return reserveTokenQuotaDB(id, quota)
	}

	result, err := cacheTryReserveTokenQuota(id, key, int64(quota))
	if err == nil && result == cacheQuotaMiss {
		if _, hydrateErr := GetTokenByKey(key, true); hydrateErr == nil {
			result, err = cacheTryReserveTokenQuota(id, key, int64(quota))
		}
	}
	if err != nil || result == cacheQuotaMiss {
		if err != nil {
			common.SysLog(common.LogText("token quota cache reserve unavailable, falling back to database: %s", err.Error()))
		}
		return reserveTokenQuotaDB(id, quota)
	}
	if result == cacheQuotaInsufficient {
		return false, nil
	}
	if err = persistTokenQuotaDelta(id, -quota); err != nil {
		compensated, compensateErr := cacheApplyTokenQuotaDelta(id, key, int64(quota))
		if compensateErr != nil || compensated != cacheQuotaOK {
			common.SysError(common.LogText("failed to compensate reserved token quota: result=%d error=%v", compensated, compensateErr))
		}
		return false, err
	}
	return true, nil
}

// TryReserveUserBonusAndPaidQuota reserves promotional credit first for eligible
// models, then paid credit, in one operation. A cache hit is authoritative even
// when older paid-wallet deductions are still waiting in the batch queue.
func TryReserveUserBonusAndPaidQuota(id, quota int, allowBonus bool) (bonus, paid int, ok bool, err error) {
	return consumeUserBonusAndPaidQuota(id, quota, allowBonus, false, nil)
}

// SettleUserBonusAndPaidQuota charges completed usage, using remaining bonus
// credit before recording any shortfall as paid debt, like ordinary settlement.
func SettleUserBonusAndPaidQuota(id, quota int) (bonus, paid int, err error) {
	bonus, paid, ok, err := consumeUserBonusAndPaidQuota(id, quota, true, true, nil)
	if err == nil && !ok {
		err = errors.New("wallet settlement was not applied")
	}
	return bonus, paid, err
}

// consumeUserBonusAndPaidQuota may commit a durable billing snapshot in the
// same transaction as the wallet deduction, after its atomic cache reservation.
// Only settlement of completed usage may allow paid debt; admission stays strict.
func consumeUserBonusAndPaidQuota(id, quota int, allowBonus, allowPaidDebt bool, persist func(*gorm.DB, int, int) error) (bonus, paid int, ok bool, err error) {
	if quota < 0 {
		return 0, 0, false, errors.New("quota cannot be negative")
	}
	if err := common.ValidateWalletQuota(quota); err != nil {
		return 0, 0, false, err
	}
	if quota == 0 {
		return 0, 0, true, nil
	}
	if !allowBonus {
		ok, err := TryReserveUserQuota(id, quota)
		if !ok {
			return 0, 0, false, err
		}
		return 0, quota, true, nil
	}
	if common.RedisEnabled {
		const script = `
if tonumber(redis.call('HGET', KEYS[1], 'Id') or '0') ~= tonumber(ARGV[2])
 or tonumber(redis.call('HGET', KEYS[1], 'CacheSchema') or '0') ~= tonumber(ARGV[3])
 or redis.call('HEXISTS', KEYS[1], 'Quota') == 0
 or redis.call('HEXISTS', KEYS[1], 'BonusQuota') == 0 then return {-1, 0, 0} end
local paid = tonumber(redis.call('HGET', KEYS[1], 'Quota'))
local bonus = tonumber(redis.call('HGET', KEYS[1], 'BonusQuota'))
local amount = tonumber(ARGV[1])
if paid == nil or bonus == nil then return {-1, 0, 0} end
local bonusSpend = math.min(math.max(bonus, 0), amount)
local paidSpend = amount - bonusSpend
if ARGV[4] ~= 'true' and paid < amount - math.max(bonus, 0) then return {0, 0, 0} end
if ARGV[4] == 'true' and paid < -tonumber(ARGV[5]) + paidSpend then return {-2, 0, 0} end
redis.call('HINCRBY', KEYS[1], 'BonusQuota', -bonusSpend)
redis.call('HINCRBY', KEYS[1], 'Quota', -paidSpend)
return {1, bonusSpend, paidSpend}`
		for attempt := range 2 {
			values, cacheErr := common.RDB.Eval(context.Background(), script, []string{getUserCacheKey(id)}, quota, id, userCacheSchemaVersion, strconv.FormatBool(allowPaidDebt), common.MaxWalletQuota).Int64Slice()
			if cacheErr != nil {
				common.SysLog(common.LogText("mixed wallet reserve cache unavailable: %s", cacheErr.Error()))
				break
			}
			if len(values) != 3 {
				break
			}
			if values[0] == 0 {
				return 0, 0, false, nil
			}
			if values[0] == -2 {
				return 0, 0, false, ErrWalletQuotaLimitExceeded
			}
			if values[0] == -1 {
				if attempt == 0 {
					if _, hydrateErr := GetUserCache(id); hydrateErr == nil {
						continue
					}
				}
				break
			}
			bonus, paid = int(values[1]), int(values[2])
			err = DB.Transaction(func(tx *gorm.DB) error {
				query := tx.Model(&User{}).Where("id = ?", id)
				if allowPaidDebt {
					query = query.Where("quota >= ?", -common.MaxWalletQuota+paid)
				}
				result := query.Updates(map[string]any{"bonus_quota": gorm.Expr("bonus_quota - ?", bonus), "quota": gorm.Expr("quota - ?", paid)})
				if result.Error != nil {
					return result.Error
				}
				if result.RowsAffected != 1 {
					return gorm.ErrRecordNotFound
				}
				if persist != nil {
					return persist(tx, bonus, paid)
				}
				return nil
			})
			if err != nil {
				if refundErr := cacheApplyUserWalletDelta(id, int64(bonus), int64(paid)); refundErr != nil {
					common.SysError(common.LogText("failed to compensate mixed wallet reservation: %s", refundErr.Error()))
				}
				return 0, 0, false, err
			}
			return bonus, paid, true, nil
		}
	}
	err = DB.Transaction(func(tx *gorm.DB) error {
		var user User
		if err := lockForUpdate(tx).Select("id", "bonus_quota", "quota").First(&user, id).Error; err != nil {
			return err
		}
		bonus = min(max(user.BonusQuota, 0), quota)
		paid = quota - bonus
		if !allowPaidDebt && user.Quota < quota-max(user.BonusQuota, 0) {
			bonus, paid = 0, 0
			return nil
		}
		if allowPaidDebt && user.Quota < -common.MaxWalletQuota+paid {
			return ErrWalletQuotaLimitExceeded
		}
		result := tx.Model(&User{}).Where("id = ? AND bonus_quota = ? AND quota = ?", id, user.BonusQuota, user.Quota).Updates(map[string]any{"bonus_quota": gorm.Expr("bonus_quota - ?", bonus), "quota": gorm.Expr("quota - ?", paid)})
		if result.Error != nil {
			return result.Error
		}
		ok = result.RowsAffected == 1
		if !ok {
			bonus, paid = 0, 0
			return nil
		}
		if persist != nil {
			return persist(tx, bonus, paid)
		}
		return nil
	})
	if err != nil {
		return 0, 0, false, err
	}
	return bonus, paid, ok, err
}

// RefundUserBonusAndPaidQuota restores each wallet to the source of its
// reservation; atomic ceiling checks prevent partially credited refunds.
func RefundUserBonusAndPaidQuota(id, bonus, paid int) error {
	if bonus < 0 || paid < 0 {
		return errors.New("quota cannot be negative")
	}
	if err := common.ValidateWalletQuota(bonus); err != nil {
		return err
	}
	if err := common.ValidateWalletQuota(paid); err != nil {
		return err
	}
	if bonus == 0 && paid == 0 {
		return nil
	}
	result := DB.Model(&User{}).Where("id = ? AND bonus_quota <= ? AND quota <= ?", id, common.MaxWalletQuota-bonus, common.MaxWalletQuota-paid).Updates(map[string]any{"bonus_quota": gorm.Expr("bonus_quota + ?", bonus), "quota": gorm.Expr("quota + ?", paid)})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return ErrWalletQuotaLimitExceeded
	}
	if err := cacheApplyUserWalletDelta(id, int64(bonus), int64(paid)); err != nil {
		common.SysError(common.LogText("failed to sync mixed wallet refund: %s", err.Error()))
	}
	return nil
}

func cacheApplyUserWalletDelta(id int, bonus, paid int64) error {
	if !common.RedisEnabled {
		return nil
	}
	const script = `
if tonumber(redis.call('HGET', KEYS[1], 'Id') or '0') ~= tonumber(ARGV[3])
 or tonumber(redis.call('HGET', KEYS[1], 'CacheSchema') or '0') ~= tonumber(ARGV[4])
 or redis.call('HEXISTS', KEYS[1], 'Quota') == 0
 or redis.call('HEXISTS', KEYS[1], 'BonusQuota') == 0 then return -1 end
redis.call('HINCRBY', KEYS[1], 'BonusQuota', ARGV[1])
redis.call('HINCRBY', KEYS[1], 'Quota', ARGV[2])
return 1`
	return common.RDB.Eval(context.Background(), script, []string{getUserCacheKey(id)}, bonus, paid, id, userCacheSchemaVersion).Err()
}
