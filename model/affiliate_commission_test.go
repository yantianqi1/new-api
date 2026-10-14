package model

import (
	"testing"

	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestAffiliateCommissionMigrationKeepsExistingUsersIneligible(t *testing.T) {
	type legacyUser struct {
		Id        int `gorm:"primaryKey"`
		Username  string
		InviterId int
	}
	type migratedUser struct {
		Id                        int `gorm:"primaryKey"`
		Username                  string
		InviterId                 int
		AffCommissionEligible     bool
		AffCommissionQuota        int
		AffCommissionHistoryQuota int
	}

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Table("users").AutoMigrate(&legacyUser{}))
	require.NoError(t, db.Table("users").Create(&legacyUser{Id: 1, Username: "legacy", InviterId: 42}).Error)

	require.NoError(t, db.Table("users").AutoMigrate(&migratedUser{}))

	var migrated migratedUser
	require.NoError(t, db.Table("users").First(&migrated, 1).Error)
	assert.False(t, migrated.AffCommissionEligible)
	assert.Zero(t, migrated.AffCommissionQuota)
	assert.Zero(t, migrated.AffCommissionHistoryQuota)
}

func TestAffiliateCommissionEligibilityIsSetWhenUserIsCreated(t *testing.T) {
	truncateTables(t)
	paymentSetting := operation_setting.GetPaymentSetting()
	originalConfirmed := paymentSetting.ComplianceConfirmed
	originalVersion := paymentSetting.ComplianceTermsVersion
	paymentSetting.ComplianceConfirmed = false
	paymentSetting.ComplianceTermsVersion = ""
	t.Cleanup(func() {
		paymentSetting.ComplianceConfirmed = originalConfirmed
		paymentSetting.ComplianceTermsVersion = originalVersion
	})

	invited := &User{Username: "new_invited", InviterId: 77}
	require.NoError(t, invited.Insert(77))
	plain := &User{Username: "new_plain"}
	require.NoError(t, plain.Insert(0))

	var invitedFromDB User
	require.NoError(t, DB.First(&invitedFromDB, invited.Id).Error)
	assert.True(t, invitedFromDB.AffCommissionEligible)

	var plainFromDB User
	require.NoError(t, DB.First(&plainFromDB, plain.Id).Error)
	assert.False(t, plainFromDB.AffCommissionEligible)
}

func TestAffiliateCommissionLedgerTopUpIsUnique(t *testing.T) {
	truncateTables(t)
	first := &AffiliateCommission{TopUpId: 901, TradeNo: "commission-unique", Status: AffiliateCommissionStatusAwarded}
	require.NoError(t, DB.Create(first).Error)

	err := DB.Create(&AffiliateCommission{TopUpId: 901, TradeNo: "commission-duplicate", Status: AffiliateCommissionStatusSkipped}).Error
	require.Error(t, err)
}
