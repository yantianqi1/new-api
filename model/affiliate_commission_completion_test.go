package model

import (
	"fmt"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setAffiliateCommissionRateForTest(t *testing.T, rate int) {
	t.Helper()
	paymentSetting := operation_setting.GetPaymentSetting()
	original := paymentSetting.AffiliateCommissionRateBasisPoints
	paymentSetting.AffiliateCommissionRateBasisPoints = rate
	t.Cleanup(func() {
		paymentSetting.AffiliateCommissionRateBasisPoints = original
	})
}

func createCommissionCompletionFixture(t *testing.T, eligible bool, inviterID int, baseQuota int) *TopUp {
	t.Helper()
	if inviterID != 0 {
		require.NoError(t, DB.Create(&User{
			Id:       inviterID,
			Username: fmt.Sprintf("inviter_%d", inviterID),
			AffCode:  fmt.Sprintf("ir%d", inviterID),
			Quota:    50,
		}).Error)
	}
	inviteeID := inviterID + 1000
	if inviterID == 0 {
		inviteeID = 1000
	}
	require.NoError(t, DB.Create(&User{
		Id:                    inviteeID,
		Username:              fmt.Sprintf("invitee_%d", inviteeID),
		AffCode:               fmt.Sprintf("ie%d", inviteeID),
		InviterId:             inviterID,
		AffCommissionEligible: eligible,
		Quota:                 25,
	}).Error)
	topUp := &TopUp{
		UserId:              inviteeID,
		Amount:              100,
		Money:               90,
		ActualPaidAmount:    "90.00",
		Currency:            "CNY",
		CommissionBaseQuota: baseQuota,
		TradeNo:             fmt.Sprintf("commission-completion-%d", inviteeID),
		PaymentMethod:       "alipay",
		PaymentProvider:     PaymentProviderEpay,
		Status:              common.TopUpStatusPending,
		CreateTime:          common.GetTimestamp(),
	}
	require.NoError(t, topUp.Insert())
	return topUp
}

func TestCalculateAffiliateCommission(t *testing.T) {
	testCases := []struct {
		name string
		base int
		rate int
		want int
	}{
		{name: "disabled", base: 9000, rate: 0, want: 0},
		{name: "decimal percent", base: 10000, rate: 1050, want: 1050},
		{name: "full rate", base: 12345, rate: 10000, want: 12345},
		{name: "round down", base: 9, rate: 1111, want: 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, CalculateAffiliateCommission(tc.base, tc.rate))
		})
	}
}

func TestCompleteOnlineTopUpAwardsCommissionExactlyOnce(t *testing.T) {
	truncateTables(t)
	setAffiliateCommissionRateForTest(t, 1050)
	baseQuota := 9 * int(common.QuotaPerUnit)
	topUp := createCommissionCompletionFixture(t, true, 101, baseQuota)

	first, err := CompleteOnlineTopUp(CompleteOnlineTopUpInput{
		TradeNo:         topUp.TradeNo,
		PaymentProvider: PaymentProviderEpay,
		PaymentMethod:   "wechat",
		SettledAmount:   "90.00",
		SettledCurrency: "cny",
	})
	require.NoError(t, err)
	require.False(t, first.AlreadyCompleted)
	require.NotNil(t, first.Commission)
	assert.Equal(t, AffiliateCommissionStatusAwarded, first.Commission.Status)
	assert.Equal(t, CalculateAffiliateCommission(baseQuota, 1050), first.Commission.CommissionQuota)
	assert.Equal(t, 1050, first.Commission.RateBasisPoints)
	assert.Equal(t, "wechat", first.TopUp.PaymentMethod)

	second, err := CompleteOnlineTopUp(CompleteOnlineTopUpInput{
		TradeNo:         topUp.TradeNo,
		PaymentProvider: PaymentProviderEpay,
		SettledAmount:   "90.00",
		SettledCurrency: "CNY",
	})
	require.NoError(t, err)
	assert.True(t, second.AlreadyCompleted)

	var invitee User
	require.NoError(t, DB.First(&invitee, topUp.UserId).Error)
	assert.Equal(t, 25+100*int(common.QuotaPerUnit), invitee.Quota)

	var inviter User
	require.NoError(t, DB.First(&inviter, 101).Error)
	assert.Equal(t, first.Commission.CommissionQuota, inviter.AffCommissionQuota)
	assert.Equal(t, first.Commission.CommissionQuota, inviter.AffCommissionHistoryQuota)

	var ledgerCount int64
	require.NoError(t, DB.Model(&AffiliateCommission{}).Where("topup_id = ?", topUp.Id).Count(&ledgerCount).Error)
	assert.EqualValues(t, 1, ledgerCount)
}

func TestCompleteOnlineTopUpDoesNotAwardLegacyInvitation(t *testing.T) {
	truncateTables(t)
	setAffiliateCommissionRateForTest(t, 1000)
	topUp := createCommissionCompletionFixture(t, false, 102, int(common.QuotaPerUnit))

	result, err := CompleteOnlineTopUp(CompleteOnlineTopUpInput{TradeNo: topUp.TradeNo, PaymentProvider: PaymentProviderEpay})
	require.NoError(t, err)
	assert.Nil(t, result.Commission)

	var count int64
	require.NoError(t, DB.Model(&AffiliateCommission{}).Count(&count).Error)
	assert.Zero(t, count)
}

func TestCompleteOnlineTopUpCompletesHistoricalOrderWithoutCommission(t *testing.T) {
	truncateTables(t)
	setAffiliateCommissionRateForTest(t, 1000)
	topUp := createCommissionCompletionFixture(t, false, 107, int(common.QuotaPerUnit))
	require.NoError(t, DB.Model(topUp).Updates(map[string]any{
		"actual_paid_amount":    "",
		"currency":              "",
		"commission_base_quota": 0,
	}).Error)

	result, err := CompleteOnlineTopUp(CompleteOnlineTopUpInput{
		TradeNo:         topUp.TradeNo,
		PaymentProvider: PaymentProviderEpay,
		SettledAmount:   "90.00",
		SettledCurrency: "CNY",
	})
	require.NoError(t, err)
	require.NotNil(t, result.TopUp)
	assert.Equal(t, common.TopUpStatusSuccess, result.TopUp.Status)
	assert.Nil(t, result.Commission)

	var invitee User
	require.NoError(t, DB.First(&invitee, topUp.UserId).Error)
	assert.Equal(t, 25+100*int(common.QuotaPerUnit), invitee.Quota)
	var inviter User
	require.NoError(t, DB.First(&inviter, 107).Error)
	assert.Zero(t, inviter.AffCommissionQuota)
	assert.Zero(t, inviter.AffCommissionHistoryQuota)
	var ledgerCount int64
	require.NoError(t, DB.Model(&AffiliateCommission{}).Count(&ledgerCount).Error)
	assert.Zero(t, ledgerCount)
}

func TestCompleteOnlineTopUpUsesCompletionRateSnapshot(t *testing.T) {
	truncateTables(t)
	setAffiliateCommissionRateForTest(t, 500)
	firstTopUp := createCommissionCompletionFixture(t, true, 103, 10000)
	first, err := CompleteOnlineTopUp(CompleteOnlineTopUpInput{TradeNo: firstTopUp.TradeNo, PaymentProvider: PaymentProviderEpay})
	require.NoError(t, err)
	assert.Equal(t, 500, first.Commission.RateBasisPoints)

	operation_setting.GetPaymentSetting().AffiliateCommissionRateBasisPoints = 1250
	secondTopUp := createCommissionCompletionFixture(t, true, 104, 10000)
	second, err := CompleteOnlineTopUp(CompleteOnlineTopUpInput{TradeNo: secondTopUp.TradeNo, PaymentProvider: PaymentProviderEpay})
	require.NoError(t, err)
	assert.Equal(t, 1250, second.Commission.RateBasisPoints)
	assert.Equal(t, 1250, second.Commission.CommissionQuota)
}

func TestCompleteOnlineTopUpAuditsDeletedInviterWithoutBlockingWallet(t *testing.T) {
	truncateTables(t)
	setAffiliateCommissionRateForTest(t, 1000)
	topUp := createCommissionCompletionFixture(t, true, 105, 10000)
	require.NoError(t, DB.Delete(&User{}, 105).Error)

	result, err := CompleteOnlineTopUp(CompleteOnlineTopUpInput{TradeNo: topUp.TradeNo, PaymentProvider: PaymentProviderEpay})
	require.NoError(t, err)
	require.NotNil(t, result.Commission)
	assert.Equal(t, AffiliateCommissionStatusSkipped, result.Commission.Status)
	assert.Equal(t, AffiliateCommissionSkipInviterDeleted, result.Commission.SkippedReason)

	var invitee User
	require.NoError(t, DB.First(&invitee, topUp.UserId).Error)
	assert.Equal(t, 25+100*int(common.QuotaPerUnit), invitee.Quota)
}

func TestCompleteOnlineTopUpAuditsRoundedZero(t *testing.T) {
	truncateTables(t)
	setAffiliateCommissionRateForTest(t, 1)
	topUp := createCommissionCompletionFixture(t, true, 106, 1)

	result, err := CompleteOnlineTopUp(CompleteOnlineTopUpInput{TradeNo: topUp.TradeNo, PaymentProvider: PaymentProviderEpay})
	require.NoError(t, err)
	require.NotNil(t, result.Commission)
	assert.Equal(t, AffiliateCommissionStatusSkipped, result.Commission.Status)
	assert.Equal(t, AffiliateCommissionSkipRoundedZero, result.Commission.SkippedReason)
}

func TestCompleteOnlineTopUpRejectsProviderAndSettlementMismatchAtomically(t *testing.T) {
	testCases := []struct {
		name  string
		input CompleteOnlineTopUpInput
	}{
		{name: "provider", input: CompleteOnlineTopUpInput{PaymentProvider: PaymentProviderStripe}},
		{name: "amount", input: CompleteOnlineTopUpInput{PaymentProvider: PaymentProviderEpay, SettledAmount: "89.99"}},
		{name: "currency", input: CompleteOnlineTopUpInput{PaymentProvider: PaymentProviderEpay, SettledCurrency: "USD"}},
	}

	for index, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			truncateTables(t)
			setAffiliateCommissionRateForTest(t, 1000)
			topUp := createCommissionCompletionFixture(t, true, 200+index, 10000)
			tc.input.TradeNo = topUp.TradeNo

			_, err := CompleteOnlineTopUp(tc.input)
			require.Error(t, err)

			var storedTopUp TopUp
			require.NoError(t, DB.First(&storedTopUp, topUp.Id).Error)
			assert.Equal(t, common.TopUpStatusPending, storedTopUp.Status)
			var invitee User
			require.NoError(t, DB.First(&invitee, topUp.UserId).Error)
			assert.Equal(t, 25, invitee.Quota)
			var count int64
			require.NoError(t, DB.Model(&AffiliateCommission{}).Count(&count).Error)
			assert.Zero(t, count)
		})
	}
}
