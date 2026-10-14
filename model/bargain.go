package model

import (
	"errors"
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	BargainStatusActive    = "active"
	BargainStatusCompleted = "completed"
)

type BargainActivity struct {
	Id           int     `json:"id" gorm:"primaryKey;autoIncrement"`
	UserId       int     `json:"user_id" gorm:"index;not null"`
	CreatorIp    string  `json:"creator_ip" gorm:"type:varchar(64);default:''"`
	OriginPrice  float64 `json:"origin_price" gorm:"not null"`
	CurrentPrice float64 `json:"current_price" gorm:"not null"`
	CutAmount    float64 `json:"cut_amount" gorm:"not null"`
	MinPrice     float64 `json:"min_price" gorm:"not null"`
	MaxSlashes   int     `json:"max_slashes" gorm:"not null"`
	CutCount     int     `json:"cut_count" gorm:"default:0"`
	Quota        int64   `json:"quota" gorm:"type:bigint;not null"`
	Status       string  `json:"status" gorm:"type:varchar(32);index;not null"`
	TopUpId      int     `json:"top_up_id" gorm:"default:0"`
	CreatedAt    int64   `json:"created_at" gorm:"not null"`
	CompletedAt  int64   `json:"completed_at" gorm:"default:0"`
}

func (BargainActivity) TableName() string {
	return "bargain_activities"
}

type BargainRecord struct {
	Id         int     `json:"id" gorm:"primaryKey;autoIncrement"`
	ActivityId int     `json:"activity_id" gorm:"index;not null"`
	InviterId  int     `json:"inviter_id" gorm:"index;not null"`
	InviteeId  int     `json:"invitee_id" gorm:"uniqueIndex;not null"`
	CutAmount  float64 `json:"cut_amount" gorm:"not null"`
	HelperIp   string  `json:"helper_ip" gorm:"type:varchar(64);index;not null"`
	CreatedAt  int64   `json:"created_at" gorm:"not null"`
}

func (BargainRecord) TableName() string {
	return "bargain_records"
}

type BargainRecordWithUser struct {
	Id        int     `json:"id"`
	Username  string  `json:"username"`
	CutAmount float64 `json:"cut_amount"`
	CreatedAt int64   `json:"created_at"`
}

type BargainInfoResponse struct {
	Enabled           bool                    `json:"enabled"`
	Title             string                  `json:"title"`
	Description       string                  `json:"description"`
	OriginPrice       float64                 `json:"origin_price"`
	CutAmount         float64                 `json:"cut_amount"`
	MaxSlashes        int                     `json:"max_slashes"`
	MinPrice          float64                 `json:"min_price"`
	Quota             int64                   `json:"quota"`
	UserLimit         int                     `json:"user_limit"`
	PurchasedCount    int64                   `json:"purchased_count"`
	UserCanPurchase   bool                    `json:"user_can_purchase"`
	HasActiveActivity bool                    `json:"has_active_activity"`
	Activity          *BargainActivity        `json:"activity,omitempty"`
	Records           []BargainRecordWithUser `json:"records"`
}

func GetActiveBargainActivity(userId int) (*BargainActivity, error) {
	var activity BargainActivity
	err := DB.Where("user_id = ? AND status = ?", userId, BargainStatusActive).First(&activity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &activity, nil
}

func GetBargainActivityById(id int) (*BargainActivity, error) {
	var activity BargainActivity
	err := DB.Where("id = ?", id).First(&activity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &activity, nil
}

func CountUserCompletedBargain(userId int) (int64, error) {
	var count int64
	err := DB.Model(&BargainActivity{}).Where("user_id = ? AND status = ?", userId, BargainStatusCompleted).Count(&count).Error
	return count, err
}

func CreateBargainActivity(userId int, clientIp ...string) (*BargainActivity, error) {
	setting := operation_setting.GetBargainSetting()
	if !setting.Enabled {
		return nil, common.NewMessage("Bargain activity is not enabled")
	}
	if userId <= 0 {
		return nil, common.NewMessage("Invalid user")
	}
	if !isFinitePositiveTopUpMoney(setting.OriginPrice) || !isFinitePositiveTopUpMoney(setting.MinPrice) || !isFinitePositiveTopUpMoney(setting.CutAmount) || setting.MaxSlashes <= 0 {
		return nil, common.NewMessage("Invalid bargain configuration")
	}
	if _, err := topUpQuotaToInt(setting.Quota); err != nil {
		return nil, err
	}
	var activity *BargainActivity
	err := DB.Transaction(func(tx *gorm.DB) error {
		var user User
		if err := lockForUpdate(tx).Where("id = ?", userId).First(&user).Error; err != nil {
			return err
		}
		var completedCount int64
		if err := tx.Model(&BargainActivity{}).Where("user_id = ? AND status = ?", userId, BargainStatusCompleted).Count(&completedCount).Error; err != nil {
			return err
		}
		if setting.UserLimit > 0 && completedCount >= int64(setting.UserLimit) {
			return common.NewMessage("Bargain purchase limit reached")
		}
		var existing BargainActivity
		err := tx.Where("user_id = ? AND status = ?", userId, BargainStatusActive).First(&existing).Error
		if err == nil {
			activity = &existing
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		var creatorIp string
		if len(clientIp) > 0 {
			creatorIp = strings.TrimSpace(clientIp[0])
		}
		activity = &BargainActivity{
			UserId: userId, CreatorIp: creatorIp, OriginPrice: setting.OriginPrice, CurrentPrice: setting.OriginPrice,
			CutAmount: setting.CutAmount, MinPrice: min(setting.MinPrice, setting.OriginPrice), MaxSlashes: setting.MaxSlashes,
			Quota: setting.Quota, Status: BargainStatusActive, CreatedAt: common.GetTimestamp(),
		}
		return tx.Create(activity).Error
	})
	return activity, err
}

func MaskUsername(name string) string {
	name = strings.TrimSpace(name)
	runes := []rune(name)
	if len(runes) == 0 {
		return "用户***"
	}
	if len(runes) == 1 {
		return string(runes[0]) + "***"
	}
	if len(runes) == 2 {
		return string(runes[0]) + "***"
	}
	return string(runes[0]) + "***" + string(runes[len(runes)-1])
}

func GetBargainRecords(activityId int) ([]BargainRecordWithUser, error) {
	var records []BargainRecord
	err := DB.Where("activity_id = ?", activityId).Order("id desc").Find(&records).Error
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return []BargainRecordWithUser{}, nil
	}

	inviteeIds := make([]int, len(records))
	for i, r := range records {
		inviteeIds[i] = r.InviteeId
	}

	var users []User
	_ = DB.Select("id, username, display_name").Where("id IN ?", inviteeIds).Find(&users).Error
	userMap := make(map[int]string, len(users))
	for _, u := range users {
		name := u.DisplayName
		if name == "" {
			name = u.Username
		}
		userMap[u.Id] = MaskUsername(name)
	}

	result := make([]BargainRecordWithUser, len(records))
	for i, r := range records {
		name := userMap[r.InviteeId]
		if name == "" {
			name = "新用户***"
		}
		result[i] = BargainRecordWithUser{
			Id:        r.Id,
			Username:  name,
			CutAmount: r.CutAmount,
			CreatedAt: r.CreatedAt,
		}
	}
	return result, nil
}

func GetUserBargainInfo(userId int) (*BargainInfoResponse, error) {
	setting := operation_setting.GetBargainSetting()
	completedCount, err := CountUserCompletedBargain(userId)
	if err != nil {
		return nil, err
	}

	userCanPurchase := true
	if setting.UserLimit > 0 && completedCount >= int64(setting.UserLimit) {
		userCanPurchase = false
	}

	resp := &BargainInfoResponse{
		Enabled:           setting.Enabled,
		Title:             setting.Title,
		Description:       setting.Description,
		OriginPrice:       setting.OriginPrice,
		CutAmount:         setting.CutAmount,
		MaxSlashes:        setting.MaxSlashes,
		MinPrice:          setting.MinPrice,
		Quota:             setting.Quota,
		UserLimit:         setting.UserLimit,
		PurchasedCount:    completedCount,
		UserCanPurchase:   userCanPurchase,
		HasActiveActivity: false,
		Records:           []BargainRecordWithUser{},
	}

	if userId > 0 {
		activity, err := GetActiveBargainActivity(userId)
		if err == nil && activity != nil {
			resp.HasActiveActivity = true
			resp.Activity = activity
			records, _ := GetBargainRecords(activity.Id)
			resp.Records = records
		}
	}

	return resp, nil
}

func TriggerBargainCut(inviterId int, newUserId int, clientIp string) (bool, error) {
	if inviterId <= 0 || newUserId <= 0 || inviterId == newUserId {
		return false, nil
	}

	setting := operation_setting.GetBargainSetting()
	if !setting.Enabled {
		return false, nil
	}

	cleanIp := strings.TrimSpace(clientIp)

	// 1. 单人终身限助力一次检查
	var helpedCount int64
	if err := DB.Model(&BargainRecord{}).Where("invitee_id = ?", newUserId).Count(&helpedCount).Error; err != nil {
		common.SysError(fmt.Sprintf("TriggerBargainCut: 查询助力记录失败 invitee=%d: %v", newUserId, err))
		return false, err
	}
	if helpedCount > 0 {
		return false, nil
	}

	// 2. IP 风控检查
	if cleanIp != "" {
		// 规则A: 同一 IP 限制时间窗口内（默认24小时）只允许有效助力一次
		ipHours := setting.IpLimitHours
		if ipHours <= 0 {
			ipHours = 24
		}
		cutoff := common.GetTimestamp() - int64(ipHours*3600)
		var ipUsedCount int64
		if err := DB.Model(&BargainRecord{}).Where("helper_ip = ? AND created_at >= ?", cleanIp, cutoff).Count(&ipUsedCount).Error; err == nil {
			if ipUsedCount > 0 {
				common.SysLog(fmt.Sprintf("TriggerBargainCut: 拦截同IP多账号助力 ip=%s invitee=%d inviter=%d", cleanIp, newUserId, inviterId))
				return false, nil
			}
		}

		// 规则B: 防自刷（新用户注册 IP 与邀请人近期登录/操作 IP 一致则拦截）
		var inviterIpLogCount int64
		logDb := LOG_DB
		if logDb == nil {
			logDb = DB
		}
		if err := logDb.Model(&Log{}).Where("user_id = ? AND ip = ?", inviterId, cleanIp).Limit(1).Count(&inviterIpLogCount).Error; err == nil {
			if inviterIpLogCount > 0 {
				common.SysLog(fmt.Sprintf("TriggerBargainCut: 拦截邀请人同IP自刷助力 ip=%s invitee=%d inviter=%d", cleanIp, newUserId, inviterId))
				return false, nil
			}
		}

		// 规则C: 防自刷（新用户注册 IP 与邀请人发起活动记录 IP 一致则拦截）
		var inviterActivityIpCount int64
		if err := DB.Model(&BargainActivity{}).Where("user_id = ? AND creator_ip = ?", inviterId, cleanIp).Limit(1).Count(&inviterActivityIpCount).Error; err == nil {
			if inviterActivityIpCount > 0 {
				common.SysLog(fmt.Sprintf("TriggerBargainCut: 拦截邀请人发起活动同IP自刷助力 ip=%s invitee=%d inviter=%d", cleanIp, newUserId, inviterId))
				return false, nil
			}
		}
	}

	// 3. 事务处理：加行级排他锁，扣减价格与插入记录
	var actualCut float64
	err := DB.Transaction(func(tx *gorm.DB) error {
		var activity BargainActivity
		if err := lockForUpdate(tx).
			Where("user_id = ? AND status = ?", inviterId, BargainStatusActive).
			First(&activity).Error; err != nil {
			return err
		}

		// 规则D: 再次确认当前活动的 CreatorIp 不是助力人的 cleanIp
		if cleanIp != "" && activity.CreatorIp != "" && strings.EqualFold(activity.CreatorIp, cleanIp) {
			common.SysLog(fmt.Sprintf("TriggerBargainCut: 拦截当前活动同IP自刷助力 ip=%s invitee=%d inviter=%d", cleanIp, newUserId, inviterId))
			return common.NewMessage("Bargain assistance from the same IP is not allowed")
		}

		// 检查是否已达到最大砍刀数，或当前价格已经达到或低于保底最低价
		if activity.CutCount >= activity.MaxSlashes {
			return common.NewMessage("Maximum bargain assists reached")
		}
		dCurrent := decimal.NewFromFloat(activity.CurrentPrice)
		dMin := decimal.NewFromFloat(activity.MinPrice)
		if dCurrent.LessThanOrEqual(dMin) {
			return common.NewMessage("Minimum bargain price reached")
		}

		dCut := decimal.NewFromFloat(activity.CutAmount)
		dNewPrice := dCurrent.Sub(dCut)
		if dNewPrice.LessThan(dMin) {
			dNewPrice = dMin
		}

		actualCutDecimal := dCurrent.Sub(dNewPrice)
		actualCut = actualCutDecimal.Round(2).InexactFloat64()
		if actualCut <= 0 {
			return common.NewMessage("Bargain discount must be positive")
		}

		record := &BargainRecord{
			ActivityId: activity.Id,
			InviterId:  inviterId,
			InviteeId:  newUserId,
			CutAmount:  actualCut,
			HelperIp:   cleanIp,
			CreatedAt:  common.GetTimestamp(),
		}
		if err := tx.Create(record).Error; err != nil {
			return err
		}

		activity.CurrentPrice = dNewPrice.Round(2).InexactFloat64()
		activity.CutCount++
		if err := tx.Save(&activity).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		// 业务无活动、已满刀等均属于正常跳过
		return false, nil
	}

	RecordLog(inviterId, LogTypeSystem, common.NewMessage("New referral reduced the bargain price by {{amount}}", map[string]any{"amount": fmt.Sprintf("%.2f", actualCut)}))
	return true, nil
}
