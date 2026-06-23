package model

import (
	"testing"

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
