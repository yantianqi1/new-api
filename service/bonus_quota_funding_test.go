package service

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func seedBonusUser(t *testing.T, id int, paidQuota int, bonusQuota int) {
	t.Helper()
	user := &model.User{
		Id:         id,
		Username:   fmt.Sprintf("bonus_user_%d", id),
		Quota:      paidQuota,
		BonusQuota: bonusQuota,
		Status:     common.UserStatusEnabled,
	}
	require.NoError(t, model.DB.Create(user).Error)
}

func getBonusQuota(t *testing.T, id int) int {
	t.Helper()
	var user model.User
	require.NoError(t, model.DB.Select("bonus_quota").Where("id = ?", id).First(&user).Error)
	return user.BonusQuota
}

func bonusRelayInfo(userId int, modelName string, tokenId int, tokenKey string) *relaycommon.RelayInfo {
	return &relaycommon.RelayInfo{
		UserId:          userId,
		TokenId:         tokenId,
		TokenKey:        tokenKey,
		OriginModelName: modelName,
		RequestId:       fmt.Sprintf("bonus-quota-%d-%s", userId, modelName),
		IsPlayground:    true,
		UserSetting: dto.UserSetting{
			BillingPreference: "wallet_only",
		},
	}
}

func ginContextWithTokenQuota(tokenQuota int) *gin.Context {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest("POST", "/v1/chat/completions", nil)
	c.Set("token_quota", tokenQuota)
	return c
}

func resetBonusQuotaModels(t *testing.T) {
	t.Helper()
	require.NoError(t, ratio_setting.UpdateBonusQuotaModelsByJSONString(`{}`))
}

func TestBonusWalletFundingConsumesBonusBeforePaid(t *testing.T) {
	truncate(t)
	resetBonusQuotaModels(t)
	seedBonusUser(t, 90, 1000, 1000)

	funding := &BonusWalletFunding{userId: 90}
	require.NoError(t, funding.PreConsume(800))

	assert.Equal(t, 200, getBonusQuota(t, 90))
	assert.Equal(t, 1000, getUserQuota(t, 90))

	require.NoError(t, funding.Settle(400))
	assert.Equal(t, 0, getBonusQuota(t, 90))
	assert.Equal(t, 800, getUserQuota(t, 90))
}

func TestBonusWalletFundingFallsBackToPaidQuota(t *testing.T) {
	truncate(t)
	resetBonusQuotaModels(t)
	seedBonusUser(t, 91, 500, 300)

	funding := &BonusWalletFunding{userId: 91}
	require.NoError(t, funding.PreConsume(400))

	assert.Equal(t, 0, getBonusQuota(t, 91))
	assert.Equal(t, 400, getUserQuota(t, 91))
}

func TestBonusWalletFundingRefundsPaidBeforeBonus(t *testing.T) {
	truncate(t)
	resetBonusQuotaModels(t)
	seedBonusUser(t, 92, 500, 300)

	funding := &BonusWalletFunding{userId: 92}
	require.NoError(t, funding.PreConsume(400))
	require.NoError(t, funding.Settle(-250))

	assert.Equal(t, 150, getBonusQuota(t, 92))
	assert.Equal(t, 500, getUserQuota(t, 92))
}

func TestBonusWalletFundingRefundRestoresPreConsumedQuota(t *testing.T) {
	truncate(t)
	resetBonusQuotaModels(t)
	seedBonusUser(t, 95, 500, 300)

	funding := &BonusWalletFunding{userId: 95}
	require.NoError(t, funding.PreConsume(400))
	require.NoError(t, funding.Refund())

	assert.Equal(t, 300, getBonusQuota(t, 95))
	assert.Equal(t, 500, getUserQuota(t, 95))
}

func TestBonusQuotaModelSettingControlsEligibility(t *testing.T) {
	resetBonusQuotaModels(t)
	t.Cleanup(func() {
		_ = ratio_setting.UpdateBonusQuotaModelsByJSONString(`{}`)
	})

	require.NoError(t, ratio_setting.UpdateBonusQuotaModelsByJSONString(`{"free-model":true,"paid-model":false}`))

	assert.True(t, ratio_setting.IsBonusQuotaModel("free-model"))
	assert.False(t, ratio_setting.IsBonusQuotaModel("paid-model"))
	assert.False(t, ratio_setting.IsBonusQuotaModel("missing-model"))
}

func TestNewBillingSessionUsesPaidQuotaForNonBonusModel(t *testing.T) {
	truncate(t)
	resetBonusQuotaModels(t)
	seedBonusUser(t, 93, 500, 1000)
	seedToken(t, 930, 93, "sk-paid-model", 5000)

	info := bonusRelayInfo(93, "paid-model", 930, "sk-paid-model")
	session, apiErr := NewBillingSession(ginContextWithTokenQuota(5000), info, 400)

	require.Nil(t, apiErr)
	require.NotNil(t, session)
	assert.Equal(t, 1000, getBonusQuota(t, 93))
	assert.Equal(t, 100, getUserQuota(t, 93))
	assert.False(t, info.BonusQuotaModel)
	assert.Equal(t, 0, info.BonusQuotaPreConsumed)
	assert.Equal(t, 400, info.PaidQuotaPreConsumed)
}

func TestNewBillingSessionUsesBonusAndPaidForEligibleModel(t *testing.T) {
	truncate(t)
	resetBonusQuotaModels(t)
	require.NoError(t, ratio_setting.UpdateBonusQuotaModelsByJSONString(`{"free-model":true}`))
	seedBonusUser(t, 94, 300, 250)
	seedToken(t, 940, 94, "sk-free-model", 5000)

	info := bonusRelayInfo(94, "free-model", 940, "sk-free-model")
	session, apiErr := NewBillingSession(ginContextWithTokenQuota(5000), info, 400)

	require.Nil(t, apiErr)
	require.NotNil(t, session)
	assert.Equal(t, 0, getBonusQuota(t, 94))
	assert.Equal(t, 150, getUserQuota(t, 94))
	assert.True(t, info.BonusQuotaModel)
	assert.Equal(t, 250, info.BonusQuotaPreConsumed)
	assert.Equal(t, 150, info.PaidQuotaPreConsumed)
}

func TestAppendBillingInfoMarksFreePointBilling(t *testing.T) {
	truncate(t)
	resetBonusQuotaModels(t)
	require.NoError(t, ratio_setting.UpdateBonusQuotaModelsByJSONString(`{"free-model":true}`))
	seedBonusUser(t, 96, 300, 250)
	seedToken(t, 960, 96, "sk-free-marker", 5000)

	info := bonusRelayInfo(96, "free-model", 960, "sk-free-marker")
	session, apiErr := NewBillingSession(ginContextWithTokenQuota(5000), info, 400)

	require.Nil(t, apiErr)
	require.NotNil(t, session)
	require.NoError(t, session.Settle(300))

	other := map[string]interface{}{}
	appendBillingInfo(info, other)

	assert.Equal(t, BillingSourceWallet, other["billing_source"])
	assert.Equal(t, true, other["bonus_quota_model"])
	assert.Equal(t, 250, other["bonus_quota"])
	assert.Equal(t, 50, other["paid_quota"])
	assert.Equal(t, "free", other["billing_marker"])
}
