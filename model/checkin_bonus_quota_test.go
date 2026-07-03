package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserCheckinAwardsBonusQuotaInsteadOfPaidQuota(t *testing.T) {
	truncateTables(t)
	checkinSetting := operation_setting.GetCheckinSetting()
	originalEnabled := checkinSetting.Enabled
	originalMinQuota := checkinSetting.MinQuota
	originalMaxQuota := checkinSetting.MaxQuota
	t.Cleanup(func() {
		checkinSetting.Enabled = originalEnabled
		checkinSetting.MinQuota = originalMinQuota
		checkinSetting.MaxQuota = originalMaxQuota
	})

	checkinSetting.Enabled = true
	checkinSetting.MinQuota = 100
	checkinSetting.MaxQuota = 100

	require.NoError(t, DB.Create(&User{Id: 701, Username: "checkin_user", Quota: 50}).Error)

	checkin, err := UserCheckin(701)

	require.NoError(t, err)
	require.Equal(t, 100, checkin.QuotaAwarded)
	var user User
	require.NoError(t, DB.First(&user, 701).Error)
	assert.Equal(t, 50, user.Quota)
	assert.Equal(t, 100, user.BonusQuota)
}

func TestTransferAffQuotaAwardsBonusQuotaInsteadOfPaidQuota(t *testing.T) {
	truncateTables(t)
	transferQuota := int(common.QuotaPerUnit)
	require.NoError(t, DB.Create(&User{
		Id:         702,
		Username:   "affiliate_user",
		Quota:      50,
		BonusQuota: 20,
		AffQuota:   transferQuota * 2,
	}).Error)

	user, err := GetUserById(702, true)
	require.NoError(t, err)

	require.NoError(t, user.TransferAffQuotaToQuota(transferQuota))

	var updated User
	require.NoError(t, DB.First(&updated, 702).Error)
	assert.Equal(t, 50, updated.Quota)
	assert.Equal(t, 20+transferQuota, updated.BonusQuota)
	assert.Equal(t, transferQuota, updated.AffQuota)
}
