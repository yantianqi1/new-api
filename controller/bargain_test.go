package controller

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupBargainControllerTestDB(t *testing.T) *gorm.DB {
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
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.BargainActivity{}, &model.BargainRecord{}, &model.TopUp{}))
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

func TestBargainAPI(t *testing.T) {
	db := setupBargainControllerTestDB(t)

	setting := operation_setting.GetBargainSetting()
	origSetting := *setting
	setting.Enabled = true
	setting.OriginPrice = 5.0
	setting.CutAmount = 0.5
	setting.MinPrice = 2.0
	setting.MaxSlashes = 6
	setting.Quota = 500000
	setting.UserLimit = 1
	t.Cleanup(func() {
		*setting = origSetting
	})

	user := &model.User{
		Username: "bargain_api_user",
		AffCode:  "bapi01",
	}
	require.NoError(t, db.Create(user).Error)

	// 1. 测试 GET /api/user/bargain (未开启活动)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("id", user.Id)
	GetBargainInfo(ctx)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var infoResp struct {
		Success bool                      `json:"success"`
		Data    model.BargainInfoResponse `json:"data"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &infoResp))
	assert.True(t, infoResp.Success)
	assert.True(t, infoResp.Data.Enabled)
	assert.False(t, infoResp.Data.HasActiveActivity)
	assert.True(t, infoResp.Data.UserCanPurchase)

	// 2. 测试 POST /api/user/bargain/start (开启活动)
	startRec := httptest.NewRecorder()
	startCtx, _ := gin.CreateTestContext(startRec)
	startCtx.Request = httptest.NewRequest("POST", "/api/user/bargain/start", nil)
	startCtx.Request.RemoteAddr = "192.168.1.100:12345"
	startCtx.Set("id", user.Id)
	StartBargainActivity(startCtx)
	assert.Equal(t, http.StatusOK, startRec.Code)

	var startResp struct {
		Success bool                  `json:"success"`
		Data    model.BargainActivity `json:"data"`
	}
	require.NoError(t, common.Unmarshal(startRec.Body.Bytes(), &startResp))
	assert.True(t, startResp.Success)
	assert.Equal(t, 5.0, startResp.Data.OriginPrice)
	assert.Equal(t, 5.0, startResp.Data.CurrentPrice)
	assert.Equal(t, 0, startResp.Data.CutCount)
	assert.Equal(t, "192.168.1.100", startResp.Data.CreatorIp)
	assert.Equal(t, model.BargainStatusActive, startResp.Data.Status)

	// 3. 再次 GET /api/user/bargain 检查已有活动
	getRec2 := httptest.NewRecorder()
	getCtx2, _ := gin.CreateTestContext(getRec2)
	getCtx2.Set("id", user.Id)
	GetBargainInfo(getCtx2)
	assert.Equal(t, http.StatusOK, getRec2.Code)

	var infoResp2 struct {
		Success bool                      `json:"success"`
		Data    model.BargainInfoResponse `json:"data"`
	}
	require.NoError(t, common.Unmarshal(getRec2.Body.Bytes(), &infoResp2))
	assert.True(t, infoResp2.Data.HasActiveActivity)
	require.NotNil(t, infoResp2.Data.Activity)
	assert.Equal(t, startResp.Data.Id, infoResp2.Data.Activity.Id)
}
