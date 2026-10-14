package operation_setting

import "github.com/QuantumNous/new-api/setting/config"

type BargainSetting struct {
	Enabled      bool    `json:"enabled"`        // 是否启用砍价大礼包
	Title        string  `json:"title"`          // 礼包活动标题
	Description  string  `json:"description"`    // 礼包活动说明
	OriginPrice  float64 `json:"origin_price"`   // 礼包原价（元）
	CutAmount    float64 `json:"cut_amount"`     // 单刀减免金额（元）
	MaxSlashes   int     `json:"max_slashes"`    // 最多砍刀数
	MinPrice     float64 `json:"min_price"`      // 最低保底到手价（元）
	Quota        int64   `json:"quota"`          // 购买后可获得的礼包额度
	UserLimit    int     `json:"user_limit"`     // 每个用户限购次数
	IpLimitHours int     `json:"ip_limit_hours"` // 同一IP助力限制时间窗口（小时）
}

var bargainSetting = BargainSetting{
	Enabled:      false,
	Title:        "好友助力·超值砍价大礼包",
	Description:  "邀请新好友注册即可帮砍，最低优惠带走超值额度！",
	OriginPrice:  5.0,
	CutAmount:    0.5,
	MaxSlashes:   6,
	MinPrice:     2.0,
	Quota:        500000,
	UserLimit:    1,
	IpLimitHours: 24,
}

func init() {
	config.GlobalConfig.Register("bargain_setting", &bargainSetting)
}

func GetBargainSetting() *BargainSetting {
	return &bargainSetting
}

func IsBargainEnabled() bool {
	return bargainSetting.Enabled
}
