package controller

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type affiliateCommissionListResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Total int `json:"total"`
		Items []struct {
			TradeNo       string `json:"trade_no"`
			MaskedInvitee string `json:"masked_invitee"`
			InviteeId     *int   `json:"invitee_id"`
			InviterId     *int   `json:"inviter_id"`
		} `json:"items"`
	} `json:"data"`
}

func setupAffiliateCommissionControllerTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	gin.SetMode(gin.TestMode)
	common.SetDatabaseTypes(common.DatabaseTypeSQLite, common.DatabaseTypeSQLite)
	common.RedisEnabled = false
	originalDB := model.DB
	originalLogDB := model.LOG_DB
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	model.DB = db
	model.LOG_DB = db
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.AffiliateCommission{}, &model.AffiliateCommissionTransfer{}))
	t.Cleanup(func() {
		model.DB = originalDB
		model.LOG_DB = originalLogDB
		sqlDB, dbErr := db.DB()
		if dbErr == nil {
			_ = sqlDB.Close()
		}
	})
	return db
}

func TestAffiliateCommissionAPIIsolatesUserRecords(t *testing.T) {
	db := setupAffiliateCommissionControllerTestDB(t)
	require.NoError(t, db.Create(&model.AffiliateCommission{TopUpId: 8001, TradeNo: "mine", InviterId: 1, InviteeId: 1234, Status: model.AffiliateCommissionStatusAwarded}).Error)
	require.NoError(t, db.Create(&model.AffiliateCommission{TopUpId: 8002, TradeNo: "other", InviterId: 2, InviteeId: 5678, Status: model.AffiliateCommissionStatusAwarded}).Error)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest("GET", "/api/user/affiliate_commission/records?p=1&page_size=20", nil)
	context.Set("id", 1)
	GetUserAffiliateCommissions(context)

	var response affiliateCommissionListResponse
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	require.True(t, response.Success)
	assert.Equal(t, 1, response.Data.Total)
	require.Len(t, response.Data.Items, 1)
	assert.Equal(t, "mine", response.Data.Items[0].TradeNo)
	assert.Equal(t, "User #1***4", response.Data.Items[0].MaskedInvitee)
	assert.Nil(t, response.Data.Items[0].InviteeId)
	assert.Nil(t, response.Data.Items[0].InviterId)
}

func TestAffiliateCommissionAdminAPIFiltersRecords(t *testing.T) {
	db := setupAffiliateCommissionControllerTestDB(t)
	require.NoError(t, db.Create(&model.AffiliateCommission{TopUpId: 8101, TradeNo: "stripe-award", InviterId: 1, InviteeId: 10, PaymentProvider: model.PaymentProviderStripe, Status: model.AffiliateCommissionStatusAwarded}).Error)
	require.NoError(t, db.Create(&model.AffiliateCommission{TopUpId: 8102, TradeNo: "epay-skip", InviterId: 2, InviteeId: 20, PaymentProvider: model.PaymentProviderEpay, Status: model.AffiliateCommissionStatusSkipped}).Error)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest("GET", "/api/user/affiliate_commission/admin/records?inviter_id=2&provider=epay&status=skipped", nil)
	GetAdminAffiliateCommissions(context)

	var response affiliateCommissionListResponse
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	require.True(t, response.Success)
	assert.Equal(t, 1, response.Data.Total)
	require.Len(t, response.Data.Items, 1)
	assert.Equal(t, "epay-skip", response.Data.Items[0].TradeNo)
	require.NotNil(t, response.Data.Items[0].InviteeId)
	assert.Equal(t, 20, *response.Data.Items[0].InviteeId)
}
