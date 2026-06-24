package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedeemBonusQuotaCodeAwardsBonusQuota(t *testing.T) {
	truncateTables(t)
	require.NoError(t, DB.Create(&User{
		Id:         801,
		Username:   "bonus_redeem_user",
		Status:     common.UserStatusEnabled,
		Quota:      50,
		BonusQuota: 20,
	}).Error)
	require.NoError(t, DB.Create(&Redemption{
		UserId:    1,
		Name:      "bonus",
		Key:       "bonus-code",
		Status:    common.RedemptionCodeStatusEnabled,
		Quota:     300,
		QuotaType: RedemptionQuotaTypeBonus,
	}).Error)

	quota, err := Redeem("bonus-code", 801)

	require.NoError(t, err)
	assert.Equal(t, 300, quota)
	var user User
	require.NoError(t, DB.First(&user, 801).Error)
	assert.Equal(t, 50, user.Quota)
	assert.Equal(t, 320, user.BonusQuota)

	var redemption Redemption
	require.NoError(t, DB.Where("key = ?", "bonus-code").First(&redemption).Error)
	assert.Equal(t, common.RedemptionCodeStatusUsed, redemption.Status)
	assert.Equal(t, 801, redemption.UsedUserId)
}

func TestRedeemDefaultQuotaTypeAwardsPaidQuota(t *testing.T) {
	truncateTables(t)
	require.NoError(t, DB.Create(&User{
		Id:         802,
		Username:   "paid_redeem_user",
		Status:     common.UserStatusEnabled,
		Quota:      50,
		BonusQuota: 20,
	}).Error)
	require.NoError(t, DB.Create(&Redemption{
		UserId: 1,
		Name:   "paid",
		Key:    "paid-code",
		Status: common.RedemptionCodeStatusEnabled,
		Quota:  300,
	}).Error)

	quota, err := Redeem("paid-code", 802)

	require.NoError(t, err)
	assert.Equal(t, 300, quota)
	var user User
	require.NoError(t, DB.First(&user, 802).Error)
	assert.Equal(t, 350, user.Quota)
	assert.Equal(t, 20, user.BonusQuota)
}

func TestRedeemDoesNotConsumeCodeWhenUserIsMissing(t *testing.T) {
	truncateTables(t)
	require.NoError(t, DB.Create(&Redemption{
		UserId:    1,
		Name:      "missing-user",
		Key:       "missing-user-code",
		Status:    common.RedemptionCodeStatusEnabled,
		Quota:     300,
		QuotaType: RedemptionQuotaTypeBonus,
	}).Error)

	quota, err := Redeem("missing-user-code", 999999)

	require.Error(t, err)
	assert.Zero(t, quota)
	var redemption Redemption
	require.NoError(t, DB.Where("key = ?", "missing-user-code").First(&redemption).Error)
	assert.Equal(t, common.RedemptionCodeStatusEnabled, redemption.Status)
	assert.Zero(t, redemption.UsedUserId)
	assert.Zero(t, redemption.RedeemedTime)
}
