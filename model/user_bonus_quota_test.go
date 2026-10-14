package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBonusWalletReservationAndRefund(t *testing.T) {
	truncateTables(t)
	resetBatchUpdateTestState(t)
	user := createReserveTestUser(t, 30)
	require.NoError(t, DB.Model(&User{}).Where("id = ?", user.Id).Update("bonus_quota", 20).Error)
	bonus, paid, ok, err := TryReserveUserBonusAndPaidQuota(user.Id, 40, true)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, 20, bonus)
	assert.Equal(t, 20, paid)
	var balance User
	require.NoError(t, DB.First(&balance, user.Id).Error)
	assert.Equal(t, 0, balance.BonusQuota)
	assert.Equal(t, 10, balance.Quota)
	_, _, ok, err = TryReserveUserBonusAndPaidQuota(user.Id, 11, true)
	require.NoError(t, err)
	assert.False(t, ok)
	require.NoError(t, RefundUserBonusAndPaidQuota(user.Id, bonus, paid))
	require.NoError(t, DB.First(&balance, user.Id).Error)
	assert.Equal(t, 20, balance.BonusQuota)
	assert.Equal(t, 30, balance.Quota)
	_, _, ok, err = TryReserveUserBonusAndPaidQuota(user.Id, 31, false)
	require.NoError(t, err)
	assert.False(t, ok, "paid-only models must not spend bonus credit")
}

func TestBonusCreditCeiling(t *testing.T) {
	truncateTables(t)
	user := createReserveTestUser(t, 0)
	require.NoError(t, DB.Model(&User{}).Where("id = ?", user.Id).Update("bonus_quota", common.MaxWalletQuota-1).Error)
	require.ErrorIs(t, IncreaseUserBonusQuota(user.Id, 2, true), ErrWalletQuotaLimitExceeded)
	require.NoError(t, IncreaseUserBonusQuota(user.Id, 1, true))
	quota, err := GetUserBonusQuota(user.Id, true)
	require.NoError(t, err)
	assert.Equal(t, common.MaxWalletQuota, quota)
}

func TestBonusWalletCacheKeepsReservationsAcrossProfileRefresh(t *testing.T) {
	truncateTables(t)
	resetBatchUpdateTestState(t)
	useUserCacheMiniRedis(t)
	common.BatchUpdateEnabled = true
	user := createReserveTestUser(t, 10)
	require.NoError(t, DB.Model(&User{}).Where("id = ?", user.Id).Update("bonus_quota", 5).Error)
	reserved, err := TryReserveUserQuota(user.Id, 8)
	require.NoError(t, err)
	require.True(t, reserved)
	bonus, paid, ok, err := TryReserveUserBonusAndPaidQuota(user.Id, 6, true)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, 5, bonus)
	assert.Equal(t, 1, paid)
	staleUser, err := GetUserById(user.Id, false)
	require.NoError(t, err)
	require.NoError(t, updateUserCache(*staleUser))
	cached, err := GetUserCache(user.Id)
	require.NoError(t, err)
	assert.Equal(t, 0, cached.BonusQuota)
	assert.Equal(t, 1, cached.Quota)
	_, _, ok, err = TryReserveUserBonusAndPaidQuota(user.Id, 2, true)
	require.NoError(t, err)
	assert.False(t, ok)
	require.NoError(t, RefundUserBonusAndPaidQuota(user.Id, bonus, paid))
	batchUpdate()
	var balance User
	require.NoError(t, DB.First(&balance, user.Id).Error)
	assert.Equal(t, 5, balance.BonusQuota)
	assert.Equal(t, 2, balance.Quota)
}

func TestUserUpdateCannotOverwriteBonusFinancialFields(t *testing.T) {
	truncateTables(t)
	user := createReserveTestUser(t, 10)
	require.NoError(t, DB.Model(&User{}).Where("id = ?", user.Id).Updates(map[string]any{"bonus_quota": 20, "aff_commission_quota": 5, "aff_commission_history_quota": 7, "aff_commission_eligible": true}).Error)
	require.NoError(t, DB.First(&user, user.Id).Error)
	user.BonusQuota = 999
	user.AffCommissionQuota = 999
	user.AffCommissionHistoryQuota = 999
	require.NoError(t, user.Update(false))
	assert.Equal(t, 20, user.BonusQuota)
	assert.Equal(t, 5, user.AffCommissionQuota)
	assert.Equal(t, 7, user.AffCommissionHistoryQuota)
}

func TestCheckinCreditsBonusAndRollsBackOverflow(t *testing.T) {
	truncateTables(t)
	require.NoError(t, DB.AutoMigrate(&Checkin{}))
	t.Cleanup(func() { require.NoError(t, DB.Where("1 = 1").Delete(&Checkin{}).Error) })
	setting := operation_setting.GetCheckinSetting()
	previous := *setting
	t.Cleanup(func() { *setting = previous })
	*setting = operation_setting.CheckinSetting{Enabled: true, MinQuota: 10, MaxQuota: 10}
	user := createReserveTestUser(t, 3)
	checkin, err := UserCheckin(user.Id)
	require.NoError(t, err)
	assert.Equal(t, 10, checkin.QuotaAwarded)
	var balance User
	require.NoError(t, DB.First(&balance, user.Id).Error)
	assert.Equal(t, 10, balance.BonusQuota)
	assert.Equal(t, 3, balance.Quota)
	_, err = UserCheckin(user.Id)
	require.Error(t, err)
	overflow := createReserveTestUser(t, 0)
	require.NoError(t, DB.Model(&User{}).Where("id = ?", overflow.Id).Update("bonus_quota", common.MaxWalletQuota-1).Error)
	_, err = UserCheckin(overflow.Id)
	require.Error(t, err)
	var count int64
	require.NoError(t, DB.Model(&Checkin{}).Where("user_id = ?", overflow.Id).Count(&count).Error)
	assert.Zero(t, count)
}

func TestBonusWalletReservationWithPaidDebt(t *testing.T) {
	for _, redis := range []bool{false, true} {
		t.Run(map[bool]string{false: "database", true: "redis"}[redis], func(t *testing.T) {
			truncateTables(t)
			resetBatchUpdateTestState(t)
			if redis {
				useUserCacheMiniRedis(t)
			}
			user := createReserveTestUser(t, 0)
			require.NoError(t, DB.Model(&User{}).Where("id = ?", user.Id).Updates(map[string]any{"quota": -10, "bonus_quota": 100}).Error)
			bonus, paid, ok, err := TryReserveUserBonusAndPaidQuota(user.Id, 20, true)
			require.NoError(t, err)
			require.True(t, ok)
			assert.Equal(t, 20, bonus)
			assert.Zero(t, paid)
			_, _, ok, err = TryReserveUserBonusAndPaidQuota(user.Id, 71, true)
			require.NoError(t, err)
			assert.False(t, ok, "net balance of 70 must not fund 71")
			var stored User
			require.NoError(t, DB.First(&stored, user.Id).Error)
			assert.Equal(t, -10, stored.Quota)
			assert.Equal(t, 80, stored.BonusQuota)
			require.NoError(t, RefundUserBonusAndPaidQuota(user.Id, bonus, paid))
		})
	}
}
