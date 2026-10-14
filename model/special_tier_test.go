package model

import (
	"fmt"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpecialTierPurchaseLimitAndQuota(t *testing.T) {
	// Configure test special tier
	testTierId := "test_tier_promo_1"
	promoTier := operation_setting.SpecialTier{
		Id:          testTierId,
		Name:        "新人首充特惠",
		Price:       9.9,
		Quota:       5000000, // 赠送额度 5M：实付金额折算额度之外额外赠送
		TotalLimit:  2,       // Site-wide total limit = 2
		UserLimit:   1,       // Per-user limit = 1
		Description: "限购1次",
		Enabled:     true,
	}

	// 实付 9.9 元按 500000/元折算 4,950,000，再加赠送额度 5,000,000
	const expectedQuotaAdded = 9950000

	paymentSetting := operation_setting.GetPaymentSetting()
	origTiers := paymentSetting.SpecialTiers
	paymentSetting.SpecialTiers = []operation_setting.SpecialTier{promoTier}
	t.Cleanup(func() {
		paymentSetting.SpecialTiers = origTiers
	})

	// Create test users
	user1 := &User{
		Username: "promo_user_1",
		AffCode:  "paff1",
		Quota:    100,
	}
	require.NoError(t, DB.Create(user1).Error)

	user2 := &User{
		Username: "promo_user_2",
		AffCode:  "paff2",
		Quota:    200,
	}
	require.NoError(t, DB.Create(user2).Error)

	user3 := &User{
		Username: "promo_user_3",
		AffCode:  "paff3",
		Quota:    300,
	}
	require.NoError(t, DB.Create(user3).Error)

	// User 1 purchases testTier
	order1 := &TopUp{
		UserId:           user1.Id,
		Amount:           10,
		Money:            9.9,
		ActualPaidAmount: "9.90",
		Currency:         "CNY",
		TradeNo:          fmt.Sprintf("promo-trade-%d-1", user1.Id),
		PaymentMethod:    "alipay",
		PaymentProvider:  PaymentProviderEpay,
		SpecialTierId:    testTierId,
		SpecialTierQuota: promoTier.Quota,
		Status:           common.TopUpStatusPending,
		CreateTime:       common.GetTimestamp(),
	}
	require.NoError(t, order1.Insert())

	res1, err := CompleteOnlineTopUp(CompleteOnlineTopUpInput{
		TradeNo:         order1.TradeNo,
		PaymentProvider: PaymentProviderEpay,
	})
	require.NoError(t, err)
	assert.Equal(t, expectedQuotaAdded, res1.QuotaAdded)

	// Verify User 1 quota increased by paid amount + tier bonus
	var updatedUser1 User
	require.NoError(t, DB.First(&updatedUser1, user1.Id).Error)
	assert.Equal(t, 100+expectedQuotaAdded, updatedUser1.Quota)

	// Verify User 1 purchase count is 1
	countU1, err := CountUserSpecialTierPurchases(user1.Id, testTierId)
	require.NoError(t, err)
	assert.Equal(t, int64(1), countU1)

	// User 1 tries to purchase AGAIN (should fail because UserLimit = 1)
	order1Dup := &TopUp{
		UserId:           user1.Id,
		Amount:           10,
		Money:            9.9,
		ActualPaidAmount: "9.90",
		Currency:         "CNY",
		TradeNo:          fmt.Sprintf("promo-trade-%d-2", user1.Id),
		PaymentMethod:    "alipay",
		PaymentProvider:  PaymentProviderEpay,
		SpecialTierId:    testTierId,
		SpecialTierQuota: promoTier.Quota,
		Status:           common.TopUpStatusPending,
		CreateTime:       common.GetTimestamp(),
	}
	require.NoError(t, order1Dup.Insert())

	_, err = CompleteOnlineTopUp(CompleteOnlineTopUpInput{
		TradeNo:         order1Dup.TradeNo,
		PaymentProvider: PaymentProviderEpay,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "purchase limit")

	// User 2 purchases testTier (Total purchases becomes 2, reaching TotalLimit)
	order2 := &TopUp{
		UserId:           user2.Id,
		Amount:           10,
		Money:            9.9,
		ActualPaidAmount: "9.90",
		Currency:         "CNY",
		TradeNo:          fmt.Sprintf("promo-trade-%d-1", user2.Id),
		PaymentMethod:    "alipay",
		PaymentProvider:  PaymentProviderEpay,
		SpecialTierId:    testTierId,
		SpecialTierQuota: promoTier.Quota,
		Status:           common.TopUpStatusPending,
		CreateTime:       common.GetTimestamp(),
	}
	require.NoError(t, order2.Insert())

	res2, err := CompleteOnlineTopUp(CompleteOnlineTopUpInput{
		TradeNo:         order2.TradeNo,
		PaymentProvider: PaymentProviderEpay,
	})
	require.NoError(t, err)
	assert.Equal(t, expectedQuotaAdded, res2.QuotaAdded)

	// Verify total purchases is 2
	totalCount, err := CountTotalSpecialTierPurchases(testTierId)
	require.NoError(t, err)
	assert.Equal(t, int64(2), totalCount)

	// User 3 tries to purchase (should fail because TotalLimit = 2 reached / sold out)
	order3 := &TopUp{
		UserId:           user3.Id,
		Amount:           10,
		Money:            9.9,
		ActualPaidAmount: "9.90",
		Currency:         "CNY",
		TradeNo:          fmt.Sprintf("promo-trade-%d-1", user3.Id),
		PaymentMethod:    "alipay",
		PaymentProvider:  PaymentProviderEpay,
		SpecialTierId:    testTierId,
		SpecialTierQuota: promoTier.Quota,
		Status:           common.TopUpStatusPending,
		CreateTime:       common.GetTimestamp(),
	}
	require.NoError(t, order3.Insert())

	_, err = CompleteOnlineTopUp(CompleteOnlineTopUpInput{
		TradeNo:         order3.TradeNo,
		PaymentProvider: PaymentProviderEpay,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sold out")
}

func TestManualCompleteTopUpSpecialTier(t *testing.T) {
	testTierId := "manual_tier_1"
	promoTier := operation_setting.SpecialTier{
		Id:          testTierId,
		Name:        "管理员手动补单测试档位",
		Price:       15.0,
		Quota:       8000000,
		TotalLimit:  10,
		UserLimit:   1,
		Description: "测试补单",
		Enabled:     true,
	}

	paymentSetting := operation_setting.GetPaymentSetting()
	origTiers := paymentSetting.SpecialTiers
	paymentSetting.SpecialTiers = []operation_setting.SpecialTier{promoTier}
	t.Cleanup(func() {
		paymentSetting.SpecialTiers = origTiers
	})

	user := &User{
		Username: "promo_manual_user",
		AffCode:  "paff_manual",
		Quota:    500,
	}
	require.NoError(t, DB.Create(user).Error)

	order := &TopUp{
		UserId:           user.Id,
		Amount:           15,
		Money:            15.0,
		ActualPaidAmount: "15.00",
		Currency:         "CNY",
		TradeNo:          fmt.Sprintf("promo-manual-%d", user.Id),
		PaymentMethod:    "alipay",
		PaymentProvider:  PaymentProviderEpay,
		SpecialTierId:    testTierId,
		SpecialTierQuota: promoTier.Quota,
		Status:           common.TopUpStatusPending,
		CreateTime:       common.GetTimestamp(),
	}
	require.NoError(t, order.Insert())

	require.NoError(t, ManualCompleteTopUp(order.TradeNo, "127.0.0.1"))

	var updatedUser User
	require.NoError(t, DB.First(&updatedUser, user.Id).Error)
	// 实付 15 元折算 7,500,000，再加赠送额度 8,000,000
	assert.Equal(t, 500+15500000, updatedUser.Quota)

	count, err := CountUserSpecialTierPurchases(user.Id, testTierId)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestSpecialTierOrderCreditsPaidAmountWithoutBonus(t *testing.T) {
	testTierId := "tier_without_bonus"

	paymentSetting := operation_setting.GetPaymentSetting()
	origTiers := paymentSetting.SpecialTiers
	paymentSetting.SpecialTiers = []operation_setting.SpecialTier{{
		Id:         testTierId,
		Name:       "无赠送档位",
		Price:      10,
		TotalLimit: 0,
		UserLimit:  1,
		Enabled:    true,
	}}
	t.Cleanup(func() {
		paymentSetting.SpecialTiers = origTiers
	})

	user := &User{
		Username: "promo_no_bonus_user",
		AffCode:  "paff_no_bonus",
		Quota:    100,
	}
	require.NoError(t, DB.Create(user).Error)

	order := &TopUp{
		UserId:           user.Id,
		Amount:           10,
		Money:            10,
		ActualPaidAmount: "10.00",
		Currency:         "CNY",
		TradeNo:          fmt.Sprintf("promo-no-bonus-%d", user.Id),
		PaymentMethod:    "alipay",
		PaymentProvider:  PaymentProviderEpay,
		SpecialTierId:    testTierId,
		SpecialTierQuota: 0,
		Status:           common.TopUpStatusPending,
		CreateTime:       common.GetTimestamp(),
	}
	require.NoError(t, order.Insert())

	// 赠送额度为 0 时仍要按实付金额折算入账，不能整单失败或入账 0。
	result, err := CompleteOnlineTopUp(CompleteOnlineTopUpInput{
		TradeNo:         order.TradeNo,
		PaymentProvider: PaymentProviderEpay,
	})
	require.NoError(t, err)
	assert.Equal(t, 5000000, result.QuotaAdded)

	var updatedUser User
	require.NoError(t, DB.First(&updatedUser, user.Id).Error)
	assert.Equal(t, 100+5000000, updatedUser.Quota)

	// 实付金额与赠送额度都为 0 的异常订单必须报错，而不是静默入账 0。
	zeroUser := &User{
		Username: "promo_zero_quota_user",
		AffCode:  "paff_zero_quota",
		Quota:    100,
	}
	require.NoError(t, DB.Create(zeroUser).Error)

	zeroOrder := &TopUp{
		UserId:          zeroUser.Id,
		TradeNo:         fmt.Sprintf("promo-zero-order-%d", zeroUser.Id),
		PaymentMethod:   "alipay",
		PaymentProvider: PaymentProviderEpay,
		SpecialTierId:   testTierId,
		Status:          common.TopUpStatusPending,
		CreateTime:      common.GetTimestamp(),
	}
	require.NoError(t, zeroOrder.Insert())

	_, err = CompleteOnlineTopUp(CompleteOnlineTopUpInput{
		TradeNo:         zeroOrder.TradeNo,
		PaymentProvider: PaymentProviderEpay,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid top-up quota")

	var unchangedZeroUser User
	require.NoError(t, DB.First(&unchangedZeroUser, zeroUser.Id).Error)
	assert.Equal(t, 100, unchangedZeroUser.Quota)
}
