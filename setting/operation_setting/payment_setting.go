package operation_setting

import "github.com/QuantumNous/new-api/setting/config"

type SpecialTier struct {
	Id          string  `json:"id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`                // 自定义充值价位（支付金额）
	Quota       int64   `json:"quota"`                // 赠送额度：在实付金额折算额度之外额外赠送（系统内部Quota单位）
	TotalLimit  int     `json:"total_limit"`          // 全站限购总次数（0表示不限制）
	UserLimit   int     `json:"user_limit"`           // 每个用户限购次数（默认1次）
	Description string  `json:"description"`          // 档位描述说明
	Enabled     bool    `json:"enabled"`              // 是否启用
	SortOrder   int     `json:"sort_order"`           // 排序
	ShowStock   *bool   `json:"show_stock,omitempty"` // 是否在用户界面显示全站限量及剩余库存（管理员始终可见），默认 true
}

func (t *SpecialTier) IsStockVisible() bool {
	if t.ShowStock == nil {
		return true
	}
	return *t.ShowStock
}

type PaymentSetting struct {
	AmountOptions                      []int           `json:"amount_options"`
	AmountDiscount                     map[int]float64 `json:"amount_discount"` // 充值金额对应的折扣，例如 100 元 0.9 表示 100 元充值享受 9 折优惠
	AffiliateCommissionRateBasisPoints int             `json:"affiliate_commission_rate_basis_points"`
	SpecialTiers                       []SpecialTier   `json:"special_tiers"`

	ComplianceConfirmed    bool   `json:"compliance_confirmed"`
	ComplianceTermsVersion string `json:"compliance_terms_version"`
	ComplianceConfirmedAt  int64  `json:"compliance_confirmed_at"`
	ComplianceConfirmedBy  int    `json:"compliance_confirmed_by"`
	ComplianceConfirmedIP  string `json:"compliance_confirmed_ip"`
}

const CurrentComplianceTermsVersion = "v1"

// 默认配置
var paymentSetting = PaymentSetting{
	AmountOptions:  []int{10, 20, 50, 100, 200, 500},
	AmountDiscount: map[int]float64{},
	SpecialTiers:   []SpecialTier{},
}

func init() {
	// 注册到全局配置管理器
	config.GlobalConfig.Register("payment_setting", &paymentSetting)
}

func GetPaymentSetting() *PaymentSetting {
	return &paymentSetting
}

func IsPaymentComplianceConfirmed() bool {
	return paymentSetting.ComplianceConfirmed &&
		paymentSetting.ComplianceTermsVersion == CurrentComplianceTermsVersion
}

func GetAffiliateCommissionRateBasisPoints() int {
	rate := paymentSetting.AffiliateCommissionRateBasisPoints
	if rate < 0 || rate > 10000 {
		return 0
	}
	return rate
}

func GetSpecialTiers() []SpecialTier {
	if paymentSetting.SpecialTiers == nil {
		return []SpecialTier{}
	}
	return paymentSetting.SpecialTiers
}

func GetSpecialTierById(id string) *SpecialTier {
	for i := range paymentSetting.SpecialTiers {
		if paymentSetting.SpecialTiers[i].Id == id {
			return &paymentSetting.SpecialTiers[i]
		}
	}
	return nil
}
