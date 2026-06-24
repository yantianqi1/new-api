package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type selfUserResponse struct {
	Success bool `json:"success"`
	Data    struct {
		ID         int `json:"id"`
		Quota      int `json:"quota"`
		BonusQuota int `json:"bonus_quota"`
		UsedQuota  int `json:"used_quota"`
	} `json:"data"`
}

func TestGetSelfIncludesBonusQuota(t *testing.T) {
	setupModelListControllerTestDB(t)
	require.NoError(t, model.DB.Create(&model.User{
		Id:         901,
		Username:   "self_bonus_user",
		Status:     common.UserStatusEnabled,
		Quota:      100,
		BonusQuota: 456,
		UsedQuota:  12,
	}).Error)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("id", 901)
	ctx.Set("role", common.RoleCommonUser)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/user/self", nil)

	GetSelf(ctx)

	require.Equal(t, http.StatusOK, recorder.Code)
	var payload selfUserResponse
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &payload))
	require.True(t, payload.Success)
	assert.Equal(t, 901, payload.Data.ID)
	assert.Equal(t, 100, payload.Data.Quota)
	assert.Equal(t, 456, payload.Data.BonusQuota)
	assert.Equal(t, 12, payload.Data.UsedQuota)
}
