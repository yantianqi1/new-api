package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type pricingResponse struct {
	Success bool            `json:"success"`
	Data    []model.Pricing `json:"data"`
}

func TestGetPricingMarksBonusQuotaModels(t *testing.T) {
	db := setupModelListControllerTestDB(t)
	require.NoError(t, ratio_setting.UpdateBonusQuotaModelsByJSONString(`{"bonus-model":true,"paid-model":false}`))
	t.Cleanup(func() {
		_ = ratio_setting.UpdateBonusQuotaModelsByJSONString(`{}`)
		model.InvalidatePricingCache()
	})
	model.InvalidatePricingCache()

	require.NoError(t, db.Create(&[]model.Channel{
		{Id: 3101, Type: 1, Key: "sk-test", Status: common.ChannelStatusEnabled, Name: "pricing-bonus-channel"},
	}).Error)
	require.NoError(t, db.Create(&[]model.Ability{
		{Group: "default", Model: "bonus-model", ChannelId: 3101, Enabled: true},
		{Group: "default", Model: "paid-model", ChannelId: 3101, Enabled: true},
	}).Error)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/pricing", nil)

	GetPricing(ctx)

	require.Equal(t, http.StatusOK, recorder.Code)
	var payload pricingResponse
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &payload))
	require.True(t, payload.Success)

	byName := pricingByModelName(payload.Data)
	require.Contains(t, byName, "bonus-model")
	require.Contains(t, byName, "paid-model")
	assert.True(t, byName["bonus-model"].BonusQuotaEnabled)
	assert.False(t, byName["paid-model"].BonusQuotaEnabled)
}
