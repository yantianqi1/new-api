package model

import (
	"errors"
	"fmt"
	"maps"
	"math"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type TopUp struct {
	Id                  int     `json:"id"`
	UserId              int     `json:"user_id" gorm:"index"`
	Amount              int64   `json:"amount"`
	Money               float64 `json:"money"`
	ActualPaidAmount    string  `json:"actual_paid_amount" gorm:"type:varchar(64);column:actual_paid_amount"`
	Currency            string  `json:"currency" gorm:"type:varchar(16)"`
	CommissionBaseQuota int     `json:"commission_base_quota" gorm:"type:bigint;column:commission_base_quota"`
	TradeNo             string  `json:"trade_no" gorm:"unique;type:varchar(255);index"`
	PaymentMethod       string  `json:"payment_method" gorm:"type:varchar(50)"`
	PaymentProvider     string  `json:"payment_provider" gorm:"type:varchar(50);default:''"`
	SpecialTierId       string  `json:"special_tier_id" gorm:"type:varchar(64);index;default:''"`
	SpecialTierQuota    int64   `json:"special_tier_quota" gorm:"type:bigint;default:0"`
	BargainActivityId   int     `json:"bargain_activity_id" gorm:"index;default:0"`
	BargainQuota        int64   `json:"bargain_quota" gorm:"type:bigint;default:0"`
	CreateTime          int64   `json:"create_time"`
	CompleteTime        int64   `json:"complete_time"`
	Status              string  `json:"status"`
}

const (
	PaymentMethodStripe       = "stripe"
	PaymentMethodCreem        = "creem"
	PaymentMethodWaffo        = "waffo"
	PaymentMethodWaffoPancake = "waffo_pancake"
	PaymentMethodBalance      = "balance"
)

const (
	PaymentProviderEpay         = "epay"
	PaymentProviderStripe       = "stripe"
	PaymentProviderCreem        = "creem"
	PaymentProviderWaffo        = "waffo"
	PaymentProviderWaffoPancake = "waffo_pancake"
	PaymentProviderBalance      = "balance"
)

var (
	ErrPaymentMethodMismatch = errors.New("payment method mismatch")
	ErrTopUpNotFound         = errors.New("topup not found")
	ErrTopUpStatusInvalid    = errors.New("topup status invalid")
	// The quota errors reach the web console, so they carry message keys.
	ErrInvalidTopUpQuota        error = common.NewMessage("Invalid top-up quota")
	ErrTopUpQuotaLimitExceeded  error = common.NewMessage("Top-up quota limit exceeded")
	ErrWalletQuotaLimitExceeded error = common.NewMessage("Wallet quota limit exceeded")
)

func (topUp *TopUp) Insert() error {
	var err error
	err = DB.Create(topUp).Error
	return err
}

func topUpQuotaMaxCurrent(creditedQuota int) (int, error) {
	if creditedQuota <= 0 || creditedQuota > common.MaxWalletQuota {
		return 0, ErrInvalidTopUpQuota
	}
	return common.MaxWalletQuota - creditedQuota, nil
}

// ValidateTopUpQuotaCapacity performs the user-facing pre-payment check. The
// settlement path repeats the same invariant with an atomic conditional
// update, because the wallet balance can change after checkout creation.
func ValidateTopUpQuotaCapacity(userId int, creditedQuota int) error {
	maxCurrentQuota, err := topUpQuotaMaxCurrent(creditedQuota)
	if err != nil {
		return err
	}

	var user User
	if err := DB.Select("quota").Where("id = ?", userId).First(&user).Error; err != nil {
		return err
	}
	if user.Quota > maxCurrentQuota {
		return ErrTopUpQuotaLimitExceeded
	}
	return nil
}

// creditTopUpQuota atomically enforces the wallet ceiling while adding quota.
// Keeping the predicate and increment in one UPDATE prevents two
// concurrent callbacks from both passing a separate read/check.
func creditTopUpQuota(tx *gorm.DB, userId int, creditedQuota int, updates map[string]any) error {
	maxCurrentQuota, err := topUpQuotaMaxCurrent(creditedQuota)
	if err != nil {
		return err
	}

	updateFields := make(map[string]any, len(updates)+1)
	maps.Copy(updateFields, updates)
	updateFields["quota"] = gorm.Expr("quota + ?", creditedQuota)

	result := tx.Model(&User{}).
		Where("id = ? AND quota <= ?", userId, maxCurrentQuota).
		Updates(updateFields)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 1 {
		return nil
	}

	var count int64
	if err := tx.Model(&User{}).Where("id = ?", userId).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return gorm.ErrRecordNotFound
	}
	return ErrTopUpQuotaLimitExceeded
}

func (topUp *TopUp) Update() error {
	var err error
	err = DB.Save(topUp).Error
	return err
}

func GetTopUpById(id int) *TopUp {
	var topUp *TopUp
	var err error
	err = DB.Where("id = ?", id).First(&topUp).Error
	if err != nil {
		return nil
	}
	return topUp
}

func GetTopUpByTradeNo(tradeNo string) *TopUp {
	var topUp *TopUp
	var err error
	err = DB.Where("trade_no = ?", tradeNo).First(&topUp).Error
	if err != nil {
		return nil
	}
	return topUp
}

func UpdatePendingTopUpStatus(tradeNo string, expectedPaymentProvider string, targetStatus string) error {
	if tradeNo == "" {
		return errors.New("payment order number not provided")
	}

	refCol := "`trade_no`"
	if common.UsingMainDatabase(common.DatabaseTypePostgreSQL) {
		refCol = `"trade_no"`
	}

	return DB.Transaction(func(tx *gorm.DB) error {
		topUp := &TopUp{}
		if err := lockForUpdate(tx).Where(refCol+" = ?", tradeNo).First(topUp).Error; err != nil {
			return ErrTopUpNotFound
		}
		if expectedPaymentProvider != "" && topUp.PaymentProvider != expectedPaymentProvider {
			return ErrPaymentMethodMismatch
		}
		if topUp.Status != common.TopUpStatusPending {
			return ErrTopUpStatusInvalid
		}

		topUp.Status = targetStatus
		return tx.Save(topUp).Error
	})
}

func RechargeEpay(tradeNo string, actualPaymentMethod string, callerIp string) (bool, error) {
	result, err := CompleteOnlineTopUp(CompleteOnlineTopUpInput{TradeNo: tradeNo, PaymentProvider: PaymentProviderEpay, PaymentMethod: actualPaymentMethod})
	if err != nil {
		return false, err
	}
	if !result.AlreadyCompleted {
		recordCompletedTopUp(result, callerIp, PaymentProviderEpay)
	}
	return result.AlreadyCompleted, nil
}

func Recharge(referenceId string, customerId string, callerIp string) error {
	result, err := CompleteOnlineTopUp(CompleteOnlineTopUpInput{TradeNo: referenceId, PaymentProvider: PaymentProviderStripe, StripeCustomer: customerId})
	if err != nil {
		return err
	}
	if !result.AlreadyCompleted {
		recordCompletedTopUp(result, callerIp, PaymentProviderStripe)
	}
	return nil
}

// topUpQueryWindowSeconds 限制充值记录查询的时间窗口（秒）。
const topUpQueryWindowSeconds int64 = 30 * 24 * 60 * 60

// topUpQueryCutoff 返回允许查询的最早 create_time（秒级 Unix 时间戳）。
func topUpQueryCutoff() int64 {
	return common.GetTimestamp() - topUpQueryWindowSeconds
}

func GetUserTopUps(userId int, pageInfo *common.PageInfo) (topups []*TopUp, total int64, err error) {
	// Start transaction
	tx := DB.Begin()
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	cutoff := topUpQueryCutoff()

	// Get total count within transaction
	err = tx.Model(&TopUp{}).Where("user_id = ? AND create_time >= ?", userId, cutoff).Count(&total).Error
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	// Get paginated topups within same transaction
	err = tx.Where("user_id = ? AND create_time >= ?", userId, cutoff).Order("id desc").Limit(pageInfo.GetPageSize()).Offset(pageInfo.GetStartIdx()).Find(&topups).Error
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	// Commit transaction
	if err = tx.Commit().Error; err != nil {
		return nil, 0, err
	}

	return topups, total, nil
}

// GetAllTopUps 获取全平台的充值记录（管理员使用，不限制时间窗口）
func GetAllTopUps(pageInfo *common.PageInfo) (topups []*TopUp, total int64, err error) {
	tx := DB.Begin()
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err = tx.Model(&TopUp{}).Count(&total).Error; err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	if err = tx.Order("id desc").Limit(pageInfo.GetPageSize()).Offset(pageInfo.GetStartIdx()).Find(&topups).Error; err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	if err = tx.Commit().Error; err != nil {
		return nil, 0, err
	}

	return topups, total, nil
}

// searchTopUpCountHardLimit 搜索充值记录时 COUNT 的安全上限，
// 防止对超大表执行无界 COUNT 触发 DoS。
const searchTopUpCountHardLimit = 10000

// SearchUserTopUps 按订单号搜索某用户的充值记录
func SearchUserTopUps(userId int, keyword string, pageInfo *common.PageInfo) (topups []*TopUp, total int64, err error) {
	tx := DB.Begin()
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	query := tx.Model(&TopUp{}).Where("user_id = ? AND create_time >= ?", userId, topUpQueryCutoff())
	if keyword != "" {
		pattern, perr := sanitizeLikePattern(keyword)
		if perr != nil {
			tx.Rollback()
			return nil, 0, perr
		}
		query = query.Where("trade_no LIKE ? ESCAPE '!'", pattern)
	}

	if err = query.Limit(searchTopUpCountHardLimit).Count(&total).Error; err != nil {
		tx.Rollback()
		common.SysError(common.LogText("failed to count search topups: %s", err.Error()))
		return nil, 0, common.NewMessage("Failed to search top-up records")
	}

	if err = query.Order("id desc").Limit(pageInfo.GetPageSize()).Offset(pageInfo.GetStartIdx()).Find(&topups).Error; err != nil {
		tx.Rollback()
		common.SysError(common.LogText("failed to search topups: %s", err.Error()))
		return nil, 0, common.NewMessage("Failed to search top-up records")
	}

	if err = tx.Commit().Error; err != nil {
		return nil, 0, err
	}
	return topups, total, nil
}

// SearchAllTopUps 按订单号搜索全平台充值记录（管理员使用，不限制时间窗口）
func SearchAllTopUps(keyword string, pageInfo *common.PageInfo) (topups []*TopUp, total int64, err error) {
	tx := DB.Begin()
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	query := tx.Model(&TopUp{})
	if keyword != "" {
		pattern, perr := sanitizeLikePattern(keyword)
		if perr != nil {
			tx.Rollback()
			return nil, 0, perr
		}
		query = query.Where("trade_no LIKE ? ESCAPE '!'", pattern)
	}

	if err = query.Limit(searchTopUpCountHardLimit).Count(&total).Error; err != nil {
		tx.Rollback()
		common.SysError(common.LogText("failed to count search topups: %s", err.Error()))
		return nil, 0, common.NewMessage("Failed to search top-up records")
	}

	if err = query.Order("id desc").Limit(pageInfo.GetPageSize()).Offset(pageInfo.GetStartIdx()).Find(&topups).Error; err != nil {
		tx.Rollback()
		common.SysError(common.LogText("failed to search topups: %s", err.Error()))
		return nil, 0, common.NewMessage("Failed to search top-up records")
	}

	if err = tx.Commit().Error; err != nil {
		return nil, 0, err
	}
	return topups, total, nil
}

func ManualCompleteTopUp(tradeNo string, callerIp string) error {
	topUp := GetTopUpByTradeNo(tradeNo)
	if topUp == nil {
		return common.NewMessage("Top-up order does not exist")
	}
	result, err := CompleteOnlineTopUp(CompleteOnlineTopUpInput{TradeNo: tradeNo, PaymentProvider: topUp.PaymentProvider, SkipCommission: true})
	if err != nil {
		return err
	}
	if !result.AlreadyCompleted {
		recordCompletedTopUp(result, callerIp, "admin")
	}
	return nil
}

func RechargeCreem(referenceId string, customerEmail string, customerName string, callerIp string) error {
	result, err := CompleteOnlineTopUp(CompleteOnlineTopUpInput{TradeNo: referenceId, PaymentProvider: PaymentProviderCreem, CustomerEmail: customerEmail})
	if err != nil {
		return err
	}
	if !result.AlreadyCompleted {
		recordCompletedTopUp(result, callerIp, PaymentProviderCreem)
	}
	return nil
}

func RechargeWaffo(tradeNo string, callerIp string) error {
	result, err := CompleteOnlineTopUp(CompleteOnlineTopUpInput{TradeNo: tradeNo, PaymentProvider: PaymentProviderWaffo})
	if err != nil {
		return err
	}
	if !result.AlreadyCompleted {
		recordCompletedTopUp(result, callerIp, PaymentProviderWaffo)
	}
	return nil
}

func RechargeWaffoPancake(tradeNo string) error {
	result, err := CompleteOnlineTopUp(CompleteOnlineTopUpInput{TradeNo: tradeNo, PaymentProvider: PaymentProviderWaffoPancake})
	if err != nil {
		return err
	}
	if !result.AlreadyCompleted {
		recordCompletedTopUp(result, "", PaymentProviderWaffoPancake)
	}
	return nil
}

func recordCompletedTopUp(result *CompleteOnlineTopUpResult, callerIp string, provider string) {
	messageKey := "Online top-up succeeded, amount added: {{quota}}, payment amount: {{amount}}"
	if provider == "admin" {
		messageKey = "Administrator completed the order, amount added: {{quota}}, payment amount: {{amount}}"
	}
	RecordTopupLog(result.TopUp.UserId, common.NewMessage(messageKey, map[string]any{
		"quota": logger.FormatQuota(result.QuotaAdded), "amount": fmt.Sprintf("%f", result.TopUp.Money),
	}), callerIp, result.TopUp.PaymentMethod, provider)
}

func topUpQuotaToInt(quota int64) (int, error) {
	value, err := common.WalletQuotaFromDecimalStrict(decimal.NewFromInt(quota))
	if err != nil || value <= 0 {
		return 0, ErrInvalidTopUpQuota
	}
	return value, nil
}

func specialTierQuotaToAdd(topUp *TopUp) (int, error) {
	if topUp.SpecialTierQuota < 0 || !isFinitePositiveTopUpMoney(topUp.Money) {
		return 0, ErrInvalidTopUpQuota
	}
	quota := decimal.NewFromFloat(topUp.Money).Mul(decimal.NewFromFloat(common.QuotaPerUnit)).Add(decimal.NewFromInt(topUp.SpecialTierQuota))
	value, err := common.WalletQuotaFromDecimalStrict(quota)
	if err != nil || value <= 0 {
		return 0, ErrInvalidTopUpQuota
	}
	return value, nil
}

func isFinitePositiveTopUpMoney(money float64) bool {
	return money > 0 && !math.IsNaN(money) && !math.IsInf(money, 0)
}
