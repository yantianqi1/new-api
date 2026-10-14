package model

import (
	"errors"
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	AffiliateCommissionStatusAwarded = "awarded"
	AffiliateCommissionStatusSkipped = "skipped"
)

const (
	AffiliateCommissionSkipNoInviter      = "no_inviter"
	AffiliateCommissionSkipInviterDeleted = "inviter_deleted"
	AffiliateCommissionSkipMissingBase    = "missing_commission_base"
	AffiliateCommissionSkipRoundedZero    = "rounded_zero"
)

var (
	ErrTopUpSettledAmountMismatch   = errors.New("topup settled amount mismatch")
	ErrTopUpSettledCurrencyMismatch = errors.New("topup settled currency mismatch")
)

type AffiliateCommission struct {
	Id                  int    `json:"id"`
	TopUpId             int    `json:"topup_id" gorm:"column:topup_id;uniqueIndex"`
	TradeNo             string `json:"trade_no" gorm:"type:varchar(255);index"`
	InviterId           int    `json:"inviter_id" gorm:"column:inviter_id;index"`
	InviteeId           int    `json:"invitee_id" gorm:"column:invitee_id;index"`
	PaymentProvider     string `json:"payment_provider" gorm:"type:varchar(50);index"`
	PaymentMethod       string `json:"payment_method" gorm:"type:varchar(50)"`
	ActualPaidAmount    string `json:"actual_paid_amount" gorm:"type:varchar(64)"`
	Currency            string `json:"currency" gorm:"type:varchar(16)"`
	CommissionBaseQuota int    `json:"commission_base_quota" gorm:"type:bigint"`
	RateBasisPoints     int    `json:"rate_basis_points" gorm:"type:bigint"`
	CommissionQuota     int    `json:"commission_quota" gorm:"type:bigint"`
	Status              string `json:"status" gorm:"type:varchar(20);index"`
	SkippedReason       string `json:"skipped_reason" gorm:"type:varchar(50)"`
	CreateTime          int64  `json:"create_time" gorm:"index"`
}

type AffiliateCommissionTransfer struct {
	Id                 int   `json:"id"`
	UserId             int   `json:"user_id" gorm:"column:user_id;index"`
	Quota              int   `json:"quota" gorm:"type:bigint"`
	PendingQuotaBefore int   `json:"pending_quota_before" gorm:"type:bigint"`
	PendingQuotaAfter  int   `json:"pending_quota_after" gorm:"type:bigint"`
	PaidQuotaBefore    int   `json:"paid_quota_before" gorm:"type:bigint"`
	PaidQuotaAfter     int   `json:"paid_quota_after" gorm:"type:bigint"`
	CreateTime         int64 `json:"create_time" gorm:"index"`
}

type CompleteOnlineTopUpInput struct {
	TradeNo                         string
	PaymentProvider                 string
	PaymentMethod                   string
	SettledAmount                   string
	SettledCurrency                 string
	AllowSettledAmountBelowSnapshot bool
	StripeCustomer                  string
	CustomerEmail                   string
	SkipCommission                  bool
}

type CompleteOnlineTopUpResult struct {
	TopUp            *TopUp
	Commission       *AffiliateCommission
	QuotaAdded       int
	AlreadyCompleted bool
}

type AffiliateCommissionSummary struct {
	PendingQuota         int   `json:"pending_quota"`
	HistoryQuota         int   `json:"history_quota"`
	EligibleInviteeCount int64 `json:"eligible_invitee_count"`
}

type AffiliateCommissionFilters struct {
	InviterId       *int
	InviteeId       *int
	TradeNo         string
	PaymentProvider string
	Status          string
	StartTime       int64
	EndTime         int64
}

const affiliateCommissionQueryHardLimit = 10000

func CalculateAffiliateCommission(baseQuota int, rateBasisPoints int) int {
	if baseQuota <= 0 || baseQuota > common.MaxWalletQuota || rateBasisPoints <= 0 || rateBasisPoints > 10000 {
		return 0
	}
	quota, err := common.WalletQuotaFromDecimalStrict(decimal.NewFromInt(int64(baseQuota)).Mul(decimal.NewFromInt(int64(rateBasisPoints))).Div(decimal.NewFromInt(10000)).Truncate(0))
	if err != nil {
		return 0
	}
	return quota
}

func TransferAffiliateCommission(userID int, amount *int, transferAll bool) (*AffiliateCommissionTransfer, error) {
	if userID <= 0 {
		return nil, common.NewMessage("Invalid user")
	}
	if transferAll == (amount != nil) {
		return nil, common.NewMessage("Choose either a quota amount or transfer all")
	}
	minimum, err := common.WalletQuotaFromDecimalStrict(decimal.NewFromFloat(common.QuotaPerUnit))
	if err != nil || minimum <= 0 {
		return nil, ErrInvalidTopUpQuota
	}
	if amount != nil && *amount < minimum {
		return nil, common.NewMessage("Minimum commission transfer quota is {{min}}", map[string]any{"min": minimum})
	}

	transfer := &AffiliateCommissionTransfer{}
	err = DB.Transaction(func(tx *gorm.DB) error {
		user := &User{}
		if err := lockForUpdate(tx).Where("id = ?", userID).First(user).Error; err != nil {
			return err
		}
		quotaToTransfer := user.AffCommissionQuota
		if amount != nil {
			quotaToTransfer = *amount
		}
		if quotaToTransfer < minimum {
			return common.NewMessage("Minimum commission transfer quota is {{min}}", map[string]any{"min": minimum})
		}
		if quotaToTransfer > user.AffCommissionQuota {
			return common.NewMessage("Insufficient affiliate commission quota")
		}

		transfer.UserId = user.Id
		transfer.Quota = quotaToTransfer
		transfer.PendingQuotaBefore = user.AffCommissionQuota
		transfer.PaidQuotaBefore = user.Quota
		user.AffCommissionQuota -= quotaToTransfer
		if quotaToTransfer > common.MaxWalletQuota || user.Quota > common.MaxWalletQuota-quotaToTransfer {
			return ErrWalletQuotaLimitExceeded
		}
		user.Quota += quotaToTransfer
		transfer.PendingQuotaAfter = user.AffCommissionQuota
		transfer.PaidQuotaAfter = user.Quota
		transfer.CreateTime = common.GetTimestamp()

		if err := tx.Save(user).Error; err != nil {
			return err
		}
		return tx.Create(transfer).Error
	})
	if err != nil {
		return nil, err
	}

	syncCreditUserQuotaCache(userID, transfer.Quota, "affiliate commission transfer")

	RecordLog(userID, LogTypeSystem, common.NewMessage("Affiliate commission transferred to paid quota: {{quota}}", map[string]any{"quota": transfer.Quota}))
	return transfer, nil
}

func GetAffiliateCommissionSummary(userID int) (*AffiliateCommissionSummary, error) {
	user := &User{}
	if err := DB.Select("aff_commission_quota", "aff_commission_history_quota").Where("id = ?", userID).First(user).Error; err != nil {
		return nil, err
	}
	summary := &AffiliateCommissionSummary{
		PendingQuota: user.AffCommissionQuota,
		HistoryQuota: user.AffCommissionHistoryQuota,
	}
	if err := DB.Model(&User{}).
		Where("inviter_id = ? AND aff_commission_eligible = ?", userID, true).
		Count(&summary.EligibleInviteeCount).Error; err != nil {
		return nil, err
	}
	return summary, nil
}

func GetUserAffiliateCommissions(userID int, pageInfo *common.PageInfo) ([]*AffiliateCommission, int64, error) {
	query := DB.Model(&AffiliateCommission{}).Where("inviter_id = ?", userID)
	return paginateAffiliateCommissions(query, pageInfo)
}

func GetAdminAffiliateCommissions(filters AffiliateCommissionFilters, pageInfo *common.PageInfo) ([]*AffiliateCommission, int64, error) {
	query := DB.Model(&AffiliateCommission{})
	if filters.InviterId != nil {
		query = query.Where("inviter_id = ?", *filters.InviterId)
	}
	if filters.InviteeId != nil {
		query = query.Where("invitee_id = ?", *filters.InviteeId)
	}
	if filters.TradeNo != "" {
		query = query.Where("trade_no = ?", filters.TradeNo)
	}
	if filters.PaymentProvider != "" {
		query = query.Where("payment_provider = ?", filters.PaymentProvider)
	}
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.StartTime > 0 {
		query = query.Where("create_time >= ?", filters.StartTime)
	}
	if filters.EndTime > 0 {
		query = query.Where("create_time <= ?", filters.EndTime)
	}
	return paginateAffiliateCommissions(query, pageInfo)
}

func paginateAffiliateCommissions(query *gorm.DB, pageInfo *common.PageInfo) ([]*AffiliateCommission, int64, error) {
	var ids []int
	if err := query.Select("id").Limit(affiliateCommissionQueryHardLimit+1).Pluck("id", &ids).Error; err != nil {
		return nil, 0, err
	}
	total := len(ids)
	if total > affiliateCommissionQueryHardLimit {
		total = affiliateCommissionQueryHardLimit
	}
	var records []*AffiliateCommission
	if err := query.Select("*").Order("id desc").Limit(pageInfo.GetPageSize()).Offset(pageInfo.GetStartIdx()).Find(&records).Error; err != nil {
		return nil, 0, err
	}
	return records, int64(total), nil
}

func GetUserAffiliateCommissionTransfers(userID int, pageInfo *common.PageInfo) ([]*AffiliateCommissionTransfer, int64, error) {
	query := DB.Model(&AffiliateCommissionTransfer{}).Where("user_id = ?", userID)
	var ids []int
	if err := query.Select("id").Limit(affiliateCommissionQueryHardLimit+1).Pluck("id", &ids).Error; err != nil {
		return nil, 0, err
	}
	total := len(ids)
	if total > affiliateCommissionQueryHardLimit {
		total = affiliateCommissionQueryHardLimit
	}
	var transfers []*AffiliateCommissionTransfer
	if err := query.Select("*").Order("id desc").Limit(pageInfo.GetPageSize()).Offset(pageInfo.GetStartIdx()).Find(&transfers).Error; err != nil {
		return nil, 0, err
	}
	return transfers, int64(total), nil
}

func CompleteOnlineTopUp(input CompleteOnlineTopUpInput) (*CompleteOnlineTopUpResult, error) {
	if strings.TrimSpace(input.TradeNo) == "" {
		return nil, common.NewMessage("Order number not provided")
	}

	result := &CompleteOnlineTopUpResult{}
	err := DB.Transaction(func(tx *gorm.DB) error {
		topUp := &TopUp{}
		if err := lockForUpdate(tx).Where("trade_no = ?", input.TradeNo).First(topUp).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTopUpNotFound
			}
			return err
		}
		result.TopUp = topUp
		if topUp.PaymentProvider != input.PaymentProvider {
			return ErrPaymentMethodMismatch
		}
		if topUp.Status == common.TopUpStatusSuccess {
			result.AlreadyCompleted = true
			return nil
		}
		if topUp.Status != common.TopUpStatusPending {
			return ErrTopUpStatusInvalid
		}

		commissionBaseQuota, err := validateTopUpSettlement(topUp, input)
		if err != nil {
			return err
		}
		if input.PaymentMethod != "" {
			topUp.PaymentMethod = input.PaymentMethod
		}

		quotaToAdd, err := onlineTopUpQuotaToAdd(topUp)
		if err != nil {
			return err
		}

		// 先锁定用户行：同一用户并发补单时，档位限购检查必须在数据库层面串行化，
		// 仅靠进程内互斥锁在多实例部署下不成立。
		purchaser := &User{}
		if err := lockForUpdate(tx).Where("id = ?", topUp.UserId).First(purchaser).Error; err != nil {
			return err
		}

		if topUp.BargainActivityId > 0 {
			var activity BargainActivity
			if err := lockForUpdate(tx).Where("id = ?", topUp.BargainActivityId).First(&activity).Error; err != nil {
				return err
			}
			if activity.Status != BargainStatusActive || activity.UserId != topUp.UserId || activity.Quota != topUp.BargainQuota {
				return common.NewMessage("Bargain activity is no longer active")
			}
			setting := operation_setting.GetBargainSetting()
			if setting.UserLimit > 0 {
				var completedCount int64
				if err := tx.Model(&BargainActivity{}).Where("user_id = ? AND status = ?", topUp.UserId, BargainStatusCompleted).Count(&completedCount).Error; err != nil {
					return err
				}
				if completedCount >= int64(setting.UserLimit) {
					return common.NewMessage("Bargain purchase limit reached")
				}
			}
			activity.Status = BargainStatusCompleted
			activity.TopUpId = topUp.Id
			activity.CompletedAt = common.GetTimestamp()
			if err := tx.Save(&activity).Error; err != nil {
				return err
			}
		}

		if topUp.SpecialTierId != "" {
			stock := &SpecialTierStock{TierId: topUp.SpecialTierId}
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(stock).Error; err != nil {
				return err
			}
			if err := lockForUpdate(tx).Where("tier_id = ?", topUp.SpecialTierId).First(stock).Error; err != nil {
				return err
			}

			tier := operation_setting.GetSpecialTierById(topUp.SpecialTierId)
			userLimit := 0
			totalLimit := 0
			if tier != nil {
				userLimit = tier.UserLimit
				totalLimit = tier.TotalLimit
			}

			if userLimit > 0 {
				var userCount int64
				if err := tx.Model(&SpecialTierPurchase{}).Where("user_id = ? AND tier_id = ?", topUp.UserId, topUp.SpecialTierId).Count(&userCount).Error; err != nil {
					return err
				}
				if userCount >= int64(userLimit) {
					return common.NewMessage("Special tier purchase limit reached")
				}
			}

			if totalLimit > 0 {
				var totalCount int64
				if err := tx.Model(&SpecialTierPurchase{}).Where("tier_id = ?", topUp.SpecialTierId).Count(&totalCount).Error; err != nil {
					return err
				}
				if totalCount >= int64(totalLimit) {
					return common.NewMessage("Special tier is sold out")
				}
			}

			purchase := &SpecialTierPurchase{
				TierId:     topUp.SpecialTierId,
				UserId:     topUp.UserId,
				TopUpId:    topUp.Id,
				TradeNo:    topUp.TradeNo,
				Price:      topUp.Money,
				Quota:      topUp.SpecialTierQuota,
				CreateTime: common.GetTimestamp(),
			}
			if err := tx.Create(purchase).Error; err != nil {
				return err
			}
		}

		updates := map[string]any{}
		if input.PaymentProvider == PaymentProviderStripe && input.StripeCustomer != "" {
			updates["stripe_customer"] = input.StripeCustomer
		}
		if input.CustomerEmail != "" && purchaser.Email == "" {
			updates["email"] = input.CustomerEmail
		}
		if err := creditTopUpQuota(tx, topUp.UserId, quotaToAdd, updates); err != nil {
			return err
		}

		topUp.Status = common.TopUpStatusSuccess
		topUp.CompleteTime = common.GetTimestamp()
		if err := tx.Save(topUp).Error; err != nil {
			return err
		}
		result.QuotaAdded = quotaToAdd

		rateBasisPoints := operation_setting.GetAffiliateCommissionRateBasisPoints()
		if input.SkipCommission || !purchaser.AffCommissionEligible || rateBasisPoints == 0 {
			return nil
		}

		commission := &AffiliateCommission{
			TopUpId:             topUp.Id,
			TradeNo:             topUp.TradeNo,
			InviterId:           purchaser.InviterId,
			InviteeId:           purchaser.Id,
			PaymentProvider:     topUp.PaymentProvider,
			PaymentMethod:       topUp.PaymentMethod,
			ActualPaidAmount:    topUp.ActualPaidAmount,
			Currency:            topUp.Currency,
			CommissionBaseQuota: commissionBaseQuota,
			RateBasisPoints:     rateBasisPoints,
			CommissionQuota:     CalculateAffiliateCommission(commissionBaseQuota, rateBasisPoints),
			Status:              AffiliateCommissionStatusAwarded,
			CreateTime:          common.GetTimestamp(),
		}
		if purchaser.InviterId == 0 {
			commission.Status = AffiliateCommissionStatusSkipped
			commission.SkippedReason = AffiliateCommissionSkipNoInviter
		} else if commissionBaseQuota <= 0 {
			commission.Status = AffiliateCommissionStatusSkipped
			commission.SkippedReason = AffiliateCommissionSkipMissingBase
		} else if commission.CommissionQuota == 0 {
			commission.Status = AffiliateCommissionStatusSkipped
			commission.SkippedReason = AffiliateCommissionSkipRoundedZero
		} else {
			inviter := &User{}
			err := lockForUpdate(tx).Where("id = ?", purchaser.InviterId).First(inviter).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				commission.Status = AffiliateCommissionStatusSkipped
				commission.SkippedReason = AffiliateCommissionSkipInviterDeleted
				commission.CommissionQuota = 0
			} else if err != nil {
				return err
			} else {
				if inviter.Id == purchaser.Id {
					return errors.New("invalid self-referral commission")
				}
				if inviter.AffCommissionQuota > common.MaxWalletQuota-commission.CommissionQuota || inviter.AffCommissionHistoryQuota > common.MaxWalletQuota-commission.CommissionQuota {
					return ErrWalletQuotaLimitExceeded
				}
				inviter.AffCommissionQuota += commission.CommissionQuota
				inviter.AffCommissionHistoryQuota += commission.CommissionQuota
				if err := tx.Save(inviter).Error; err != nil {
					return err
				}
			}
		}
		if err := tx.Create(commission).Error; err != nil {
			return err
		}
		result.Commission = commission
		return nil
	})
	if err != nil {
		return nil, err
	}
	if result.AlreadyCompleted {
		return result, nil
	}

	syncCreditUserQuotaCache(result.TopUp.UserId, result.QuotaAdded, "online topup")

	if result.Commission != nil {
		if result.Commission.Status == AffiliateCommissionStatusAwarded {
			RecordLog(result.Commission.InviterId, LogTypeSystem, common.NewMessage("Affiliate top-up commission awarded: {{quota}}, order: {{order}}", map[string]any{"quota": result.Commission.CommissionQuota, "order": result.Commission.TradeNo}))
		} else {
			common.SysLog(fmt.Sprintf(
				"topup commission skipped order=%s inviter=%d invitee=%d reason=%s",
				result.Commission.TradeNo,
				result.Commission.InviterId,
				result.Commission.InviteeId,
				result.Commission.SkippedReason,
			))
		}
	}
	return result, nil
}

func validateTopUpSettlement(topUp *TopUp, input CompleteOnlineTopUpInput) (int, error) {
	if strings.TrimSpace(topUp.ActualPaidAmount) == "" && topUp.CommissionBaseQuota == 0 {
		if input.SettledAmount != "" {
			settledAmount, err := decimal.NewFromString(input.SettledAmount)
			if err != nil || !settledAmount.IsPositive() {
				return 0, ErrTopUpSettledAmountMismatch
			}
		}
		if input.SettledCurrency != "" {
			topUp.Currency = strings.ToUpper(strings.TrimSpace(input.SettledCurrency))
		}
		return 0, nil
	}

	commissionBaseQuota := topUp.CommissionBaseQuota
	if commissionBaseQuota < 0 || commissionBaseQuota > common.MaxWalletQuota {
		return 0, ErrInvalidTopUpQuota
	}
	if input.SettledAmount != "" {
		expectedAmount, err := decimal.NewFromString(topUp.ActualPaidAmount)
		if err != nil || !expectedAmount.IsPositive() {
			return 0, ErrTopUpSettledAmountMismatch
		}
		settledAmount, err := decimal.NewFromString(input.SettledAmount)
		if err != nil || !settledAmount.IsPositive() {
			return 0, ErrTopUpSettledAmountMismatch
		}
		if !settledAmount.Equal(expectedAmount) {
			if !input.AllowSettledAmountBelowSnapshot || settledAmount.GreaterThan(expectedAmount) {
				return 0, ErrTopUpSettledAmountMismatch
			}
			commissionBaseQuota, err = common.WalletQuotaFromDecimalStrict(decimal.NewFromInt(int64(commissionBaseQuota)).Mul(settledAmount).Div(expectedAmount).Truncate(0))
			if err != nil {
				return 0, ErrInvalidTopUpQuota
			}
			topUp.ActualPaidAmount = settledAmount.String()
			topUp.CommissionBaseQuota = commissionBaseQuota
		}
	}
	if input.SettledCurrency != "" {
		settledCurrency := strings.ToUpper(strings.TrimSpace(input.SettledCurrency))
		if topUp.Currency != "" && !strings.EqualFold(topUp.Currency, settledCurrency) {
			return 0, ErrTopUpSettledCurrencyMismatch
		}
		topUp.Currency = settledCurrency
	}
	return commissionBaseQuota, nil
}

func onlineTopUpQuotaToAdd(topUp *TopUp) (int, error) {
	if topUp.BargainActivityId > 0 {
		return topUpQuotaToInt(topUp.BargainQuota)
	}
	if topUp.SpecialTierId != "" {
		return specialTierQuotaToAdd(topUp)
	}
	if common.QuotaPerUnit <= 0 {
		return 0, ErrInvalidTopUpQuota
	}
	var quota decimal.Decimal
	switch topUp.PaymentProvider {
	case PaymentProviderStripe:
		if !isFinitePositiveTopUpMoney(topUp.Money) {
			return 0, ErrInvalidTopUpQuota
		}
		quota = decimal.NewFromFloat(topUp.Money).Mul(decimal.NewFromFloat(common.QuotaPerUnit))
	case PaymentProviderCreem:
		quota = decimal.NewFromInt(topUp.Amount)
	case PaymentProviderEpay, PaymentProviderWaffo, PaymentProviderWaffoPancake:
		quota = decimal.NewFromInt(topUp.Amount).Mul(decimal.NewFromFloat(common.QuotaPerUnit))
	default:
		return 0, ErrPaymentMethodMismatch
	}
	value, err := common.WalletQuotaFromDecimalStrict(quota)
	if err != nil || value <= 0 {
		return 0, ErrInvalidTopUpQuota
	}
	return value, nil
}
