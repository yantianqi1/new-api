package model

import (
	"fmt"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBargainActivityFlowAndRiskControl(t *testing.T) {
	// 启用砍价配置
	setting := operation_setting.GetBargainSetting()
	origSetting := *setting
	setting.Enabled = true
	setting.OriginPrice = 5.0
	setting.CutAmount = 0.5
	setting.MaxSlashes = 6
	setting.MinPrice = 2.0
	setting.Quota = 500000
	setting.UserLimit = 1
	setting.IpLimitHours = 24
	t.Cleanup(func() {
		*setting = origSetting
	})

	// 1. 创建发起人用户
	inviter := &User{
		Username: fmt.Sprintf("bargain_inviter_%d", common.GetTimestamp()),
		AffCode:  common.GetRandomString(6),
		Quota:    0,
	}
	require.NoError(t, DB.Create(inviter).Error)

	// 2. 模拟邀请人曾使用某 IP
	inviterIp := "192.168.1.100"
	require.NoError(t, DB.Create(&Log{
		UserId: inviter.Id,
		Ip:     inviterIp,
		Type:   LogTypeSystem,
	}).Error)

	// 3. 发起砍价活动
	activity, err := CreateBargainActivity(inviter.Id, inviterIp)
	require.NoError(t, err)
	require.NotNil(t, activity)
	assert.Equal(t, inviterIp, activity.CreatorIp)
	assert.Equal(t, 5.0, activity.CurrentPrice)
	assert.Equal(t, 0, activity.CutCount)
	assert.Equal(t, BargainStatusActive, activity.Status)

	// 4. 重复发起应返回同一进行中活动
	activity2, err := CreateBargainActivity(inviter.Id)
	require.NoError(t, err)
	assert.Equal(t, activity.Id, activity2.Id)

	// 5. 新人1注册，正常助力砍一刀 (不同IP)
	userHelper1 := &User{
		Username:  fmt.Sprintf("helper_1_%d", common.GetTimestamp()),
		AffCode:   common.GetRandomString(6),
		InviterId: inviter.Id,
	}
	require.NoError(t, DB.Create(userHelper1).Error)

	cutSuccess, err := TriggerBargainCut(inviter.Id, userHelper1.Id, "10.0.0.1")
	require.NoError(t, err)
	assert.True(t, cutSuccess)

	// 验证价格扣减
	updatedAct, err := GetBargainActivityById(activity.Id)
	require.NoError(t, err)
	assert.Equal(t, 4.5, updatedAct.CurrentPrice)
	assert.Equal(t, 1, updatedAct.CutCount)

	// 6. 新人1重复助力应被拦截 (单人终身限助力1次)
	cutAgain, err := TriggerBargainCut(inviter.Id, userHelper1.Id, "10.0.0.2")
	require.NoError(t, err)
	assert.False(t, cutAgain)

	// 7. 新人2注册，但使用同 IP "10.0.0.1" (同IP频控拦截)
	userHelper2 := &User{
		Username:  fmt.Sprintf("helper_2_%d", common.GetTimestamp()),
		AffCode:   common.GetRandomString(6),
		InviterId: inviter.Id,
	}
	require.NoError(t, DB.Create(userHelper2).Error)

	cutIpBlocked, err := TriggerBargainCut(inviter.Id, userHelper2.Id, "10.0.0.1")
	require.NoError(t, err)
	assert.False(t, cutIpBlocked)

	// 8. 新人3注册，使用邀请人的历史 IP "192.168.1.100" (自刷拦截)
	userHelper3 := &User{
		Username:  fmt.Sprintf("helper_3_%d", common.GetTimestamp()),
		AffCode:   common.GetRandomString(6),
		InviterId: inviter.Id,
	}
	require.NoError(t, DB.Create(userHelper3).Error)

	cutSelfBlocked, err := TriggerBargainCut(inviter.Id, userHelper3.Id, inviterIp)
	require.NoError(t, err)
	assert.False(t, cutSelfBlocked)

	// 9. 连续邀请有效好友直至保底价 (当前 4.5，再砍 5 刀到 2.0，再试一刀到保底)
	for i := 4; i <= 9; i++ {
		h := &User{
			Username:  fmt.Sprintf("helper_%d_%d", i, common.GetTimestamp()),
			AffCode:   common.GetRandomString(6),
			InviterId: inviter.Id,
		}
		require.NoError(t, DB.Create(h).Error)
		ip := fmt.Sprintf("10.0.%d.1", i)
		_, _ = TriggerBargainCut(inviter.Id, h.Id, ip)
	}

	actAtFloor, err := GetBargainActivityById(activity.Id)
	require.NoError(t, err)
	// 4.5 减 5 刀 * 0.5 = 2.0，第 6 刀已被达到底价或满刀
	assert.Equal(t, 2.0, actAtFloor.CurrentPrice)
	assert.Equal(t, 6, actAtFloor.CutCount)

	// 再来一个好友也不得再降价或超刀数
	hExtra := &User{
		Username:  fmt.Sprintf("helper_extra_%d", common.GetTimestamp()),
		AffCode:   common.GetRandomString(6),
		InviterId: inviter.Id,
	}
	require.NoError(t, DB.Create(hExtra).Error)
	cutExtra, _ := TriggerBargainCut(inviter.Id, hExtra.Id, "10.0.99.1")
	assert.False(t, cutExtra)

	// 10. 测试随时以当前价购买 (在 2.0 元时购买)
	tradeNo := fmt.Sprintf("BGT_%d", common.GetTimestamp())
	topUp := &TopUp{
		UserId:            inviter.Id,
		Amount:            2,
		Money:             2.0,
		ActualPaidAmount:  "2.00",
		Currency:          "CNY",
		TradeNo:           tradeNo,
		PaymentProvider:   PaymentProviderEpay,
		PaymentMethod:     "alipay",
		BargainActivityId: activity.Id,
		BargainQuota:      activity.Quota,
		Status:            common.TopUpStatusPending,
		CreateTime:        common.GetTimestamp(),
	}
	require.NoError(t, topUp.Insert())

	res, err := CompleteOnlineTopUp(CompleteOnlineTopUpInput{
		TradeNo:         tradeNo,
		PaymentProvider: PaymentProviderEpay,
		PaymentMethod:   "alipay",
		SettledAmount:   "2.00",
		SettledCurrency: "CNY",
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, int(setting.Quota), res.QuotaAdded)

	// 验证活动状态已被标记为 completed
	completedAct, err := GetBargainActivityById(activity.Id)
	require.NoError(t, err)
	assert.Equal(t, BargainStatusCompleted, completedAct.Status)
	assert.Equal(t, topUp.Id, completedAct.TopUpId)

	// 验证限购：达到限购次数后再次发起应被拒绝
	_, err = CreateBargainActivity(inviter.Id)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "purchase limit")
}
