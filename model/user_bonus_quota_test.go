package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"
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

func TestInviteeRegistrationAwardGoesToBonusQuota(t *testing.T) {
	truncateTables(t)
	originalQuotaForNewUser := common.QuotaForNewUser
	originalQuotaForInvitee := common.QuotaForInvitee
	originalQuotaForInviter := common.QuotaForInviter
	paymentSetting := operation_setting.GetPaymentSetting()
	originalComplianceConfirmed := paymentSetting.ComplianceConfirmed
	originalComplianceTermsVersion := paymentSetting.ComplianceTermsVersion
	t.Cleanup(func() {
		common.QuotaForNewUser = originalQuotaForNewUser
		common.QuotaForInvitee = originalQuotaForInvitee
		common.QuotaForInviter = originalQuotaForInviter
		paymentSetting.ComplianceConfirmed = originalComplianceConfirmed
		paymentSetting.ComplianceTermsVersion = originalComplianceTermsVersion
	})
	common.QuotaForNewUser = 0
	common.QuotaForInvitee = 222
	common.QuotaForInviter = 0
	paymentSetting.ComplianceConfirmed = true
	paymentSetting.ComplianceTermsVersion = operation_setting.CurrentComplianceTermsVersion

	user := &User{
		Username: "bonus_invitee_registration_user",
		Password: "password123",
	}

	require.NoError(t, user.Insert(9001))

	var created User
	require.NoError(t, DB.Where("username = ?", user.Username).First(&created).Error)
	assert.Zero(t, created.Quota)
	assert.Equal(t, 222, created.BonusQuota)
}

func TestInviteeOAuthRegistrationAwardGoesToBonusQuota(t *testing.T) {
	truncateTables(t)
	originalQuotaForNewUser := common.QuotaForNewUser
	originalQuotaForInvitee := common.QuotaForInvitee
	originalQuotaForInviter := common.QuotaForInviter
	paymentSetting := operation_setting.GetPaymentSetting()
	originalComplianceConfirmed := paymentSetting.ComplianceConfirmed
	originalComplianceTermsVersion := paymentSetting.ComplianceTermsVersion
	t.Cleanup(func() {
		common.QuotaForNewUser = originalQuotaForNewUser
		common.QuotaForInvitee = originalQuotaForInvitee
		common.QuotaForInviter = originalQuotaForInviter
		paymentSetting.ComplianceConfirmed = originalComplianceConfirmed
		paymentSetting.ComplianceTermsVersion = originalComplianceTermsVersion
	})
	common.QuotaForNewUser = 0
	common.QuotaForInvitee = 333
	common.QuotaForInviter = 0
	paymentSetting.ComplianceConfirmed = true
	paymentSetting.ComplianceTermsVersion = operation_setting.CurrentComplianceTermsVersion

	user := &User{
		Username: "bonus_invitee_oauth_registration_user",
	}

	require.NoError(t, DB.Transaction(func(tx *gorm.DB) error {
		return user.InsertWithTx(tx, 9002)
	}))
	user.FinalizeOAuthUserCreation(9002)

	var created User
	require.NoError(t, DB.Where("username = ?", user.Username).First(&created).Error)
	assert.Zero(t, created.Quota)
	assert.Equal(t, 333, created.BonusQuota)
}
