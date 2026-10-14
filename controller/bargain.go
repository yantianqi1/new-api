package controller

import (
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

// GetBargainInfo 获取当前用户的砍价大礼包信息
func GetBargainInfo(c *gin.Context) {
	userId := c.GetInt("id")
	info, err := model.GetUserBargainInfo(userId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, info)
}

// StartBargainActivity 发起/获取当前用户的砍价活动
func StartBargainActivity(c *gin.Context) {
	userId := c.GetInt("id")
	if userId <= 0 {
		common.ApiErrorT(c, "User is not signed in")
		return
	}

	var clientIp string
	if c.Request != nil {
		clientIp = c.ClientIP()
	}
	activity, err := model.CreateBargainActivity(userId, clientIp)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, activity)
}
