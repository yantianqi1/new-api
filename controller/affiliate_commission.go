package controller

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

type userAffiliateCommissionRecord struct {
	Id                  int    `json:"id"`
	TradeNo             string `json:"trade_no"`
	MaskedInvitee       string `json:"masked_invitee"`
	PaymentProvider     string `json:"payment_provider"`
	PaymentMethod       string `json:"payment_method"`
	ActualPaidAmount    string `json:"actual_paid_amount"`
	Currency            string `json:"currency"`
	CommissionBaseQuota int    `json:"commission_base_quota"`
	RateBasisPoints     int    `json:"rate_basis_points"`
	CommissionQuota     int    `json:"commission_quota"`
	Status              string `json:"status"`
	SkippedReason       string `json:"skipped_reason,omitempty"`
	CreateTime          int64  `json:"create_time"`
}

type transferAffiliateCommissionRequest struct {
	Quota *int `json:"quota"`
	All   bool `json:"all"`
}

func GetAffiliateCommissionSummary(c *gin.Context) {
	summary, err := model.GetAffiliateCommissionSummary(c.GetInt("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, summary)
}

func GetUserAffiliateCommissions(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	records, total, err := model.GetUserAffiliateCommissions(c.GetInt("id"), pageInfo)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	items := make([]userAffiliateCommissionRecord, 0, len(records))
	for _, record := range records {
		items = append(items, userAffiliateCommissionRecord{
			Id:                  record.Id,
			TradeNo:             record.TradeNo,
			MaskedInvitee:       maskAffiliateInviteeID(record.InviteeId),
			PaymentProvider:     record.PaymentProvider,
			PaymentMethod:       record.PaymentMethod,
			ActualPaidAmount:    record.ActualPaidAmount,
			Currency:            record.Currency,
			CommissionBaseQuota: record.CommissionBaseQuota,
			RateBasisPoints:     record.RateBasisPoints,
			CommissionQuota:     record.CommissionQuota,
			Status:              record.Status,
			SkippedReason:       record.SkippedReason,
			CreateTime:          record.CreateTime,
		})
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(items)
	common.ApiSuccess(c, pageInfo)
}

func GetUserAffiliateCommissionTransfers(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	transfers, total, err := model.GetUserAffiliateCommissionTransfers(c.GetInt("id"), pageInfo)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(transfers)
	common.ApiSuccess(c, pageInfo)
}

func TransferAffiliateCommission(c *gin.Context) {
	request := transferAffiliateCommissionRequest{}
	if err := c.ShouldBindJSON(&request); err != nil {
		common.ApiError(c, err)
		return
	}
	transfer, err := model.TransferAffiliateCommission(c.GetInt("id"), request.Quota, request.All)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, transfer)
}

func GetAdminAffiliateCommissions(c *gin.Context) {
	filters, err := affiliateCommissionFiltersFromContext(c)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo := common.GetPageQuery(c)
	records, total, err := model.GetAdminAffiliateCommissions(filters, pageInfo)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(records)
	common.ApiSuccess(c, pageInfo)
}

func affiliateCommissionFiltersFromContext(c *gin.Context) (model.AffiliateCommissionFilters, error) {
	filters := model.AffiliateCommissionFilters{
		TradeNo:         strings.TrimSpace(c.Query("trade_no")),
		PaymentProvider: strings.TrimSpace(c.Query("provider")),
		Status:          strings.TrimSpace(c.Query("status")),
	}
	var err error
	filters.InviterId, err = optionalPositiveInt(c.Query("inviter_id"))
	if err != nil {
		return filters, fmt.Errorf("invalid inviter_id")
	}
	filters.InviteeId, err = optionalPositiveInt(c.Query("invitee_id"))
	if err != nil {
		return filters, fmt.Errorf("invalid invitee_id")
	}
	if filters.Status != "" && filters.Status != model.AffiliateCommissionStatusAwarded && filters.Status != model.AffiliateCommissionStatusSkipped {
		return filters, fmt.Errorf("invalid commission status")
	}
	if value := strings.TrimSpace(c.Query("start_timestamp")); value != "" {
		filters.StartTime, err = strconv.ParseInt(value, 10, 64)
		if err != nil || filters.StartTime < 0 {
			return filters, fmt.Errorf("invalid start_timestamp")
		}
	}
	if value := strings.TrimSpace(c.Query("end_timestamp")); value != "" {
		filters.EndTime, err = strconv.ParseInt(value, 10, 64)
		if err != nil || filters.EndTime < 0 {
			return filters, fmt.Errorf("invalid end_timestamp")
		}
	}
	if filters.StartTime > 0 && filters.EndTime > 0 && filters.StartTime > filters.EndTime {
		return filters, fmt.Errorf("invalid commission time range")
	}
	return filters, nil
}

func optionalPositiveInt(value string) (*int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return nil, fmt.Errorf("invalid positive integer")
	}
	return &parsed, nil
}

func maskAffiliateInviteeID(userID int) string {
	digits := strconv.Itoa(userID)
	if len(digits) <= 2 {
		return "User #**"
	}
	return "User #" + digits[:1] + "***" + digits[len(digits)-1:]
}
