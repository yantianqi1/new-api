package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestNewUserRegistrationAwardGoesToBonusQuota(t *testing.T) {
	truncateTables(t)
	originalQuotaForNewUser := common.QuotaForNewUser
	t.Cleanup(func() {
		common.QuotaForNewUser = originalQuotaForNewUser
	})
	common.QuotaForNewUser = 321

	user := &User{
		Username: "bonus_registration_user",
		Password: "password123",
	}

	require.NoError(t, user.Insert(0))

	var created User
	require.NoError(t, DB.Where("username = ?", user.Username).First(&created).Error)
	assert.Zero(t, created.Quota)
	assert.Equal(t, 321, created.BonusQuota)
}

func TestNewOAuthUserRegistrationAwardGoesToBonusQuota(t *testing.T) {
	truncateTables(t)
	originalQuotaForNewUser := common.QuotaForNewUser
	t.Cleanup(func() {
		common.QuotaForNewUser = originalQuotaForNewUser
	})
	common.QuotaForNewUser = 654

	user := &User{
		Username: "bonus_oauth_registration_user",
	}

	require.NoError(t, DB.Transaction(func(tx *gorm.DB) error {
		return user.InsertWithTx(tx, 0)
	}))

	var created User
	require.NoError(t, DB.Where("username = ?", user.Username).First(&created).Error)
	assert.Zero(t, created.Quota)
	assert.Equal(t, 654, created.BonusQuota)
}
