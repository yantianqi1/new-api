package model

type SpecialTierPurchase struct {
	Id         int     `json:"id"`
	TierId     string  `json:"tier_id" gorm:"type:varchar(64);index;not null"`
	UserId     int     `json:"user_id" gorm:"index;not null"`
	TopUpId    int     `json:"top_up_id" gorm:"uniqueIndex;not null"`
	TradeNo    string  `json:"trade_no" gorm:"type:varchar(255);not null"`
	Price      float64 `json:"price"`
	Quota      int64   `json:"quota"`
	CreateTime int64   `json:"create_time"`
}

func CountUserSpecialTierPurchases(userId int, tierId string) (int64, error) {
	var count int64
	err := DB.Model(&SpecialTierPurchase{}).Where("user_id = ? AND tier_id = ?", userId, tierId).Count(&count).Error
	return count, err
}

func CountTotalSpecialTierPurchases(tierId string) (int64, error) {
	var count int64
	err := DB.Model(&SpecialTierPurchase{}).Where("tier_id = ?", tierId).Count(&count).Error
	return count, err
}

func GetUserSpecialTierPurchaseCounts(userId int) (map[string]int64, error) {
	type Result struct {
		TierId string `gorm:"column:tier_id"`
		Count  int64  `gorm:"column:count"`
	}
	var results []Result
	err := DB.Model(&SpecialTierPurchase{}).
		Select("tier_id, count(*) as count").
		Where("user_id = ?", userId).
		Group("tier_id").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int64, len(results))
	for _, r := range results {
		counts[r.TierId] = r.Count
	}
	return counts, nil
}

func GetAllSpecialTierPurchaseCounts() (map[string]int64, error) {
	type Result struct {
		TierId string `gorm:"column:tier_id"`
		Count  int64  `gorm:"column:count"`
	}
	var results []Result
	err := DB.Model(&SpecialTierPurchase{}).
		Select("tier_id, count(*) as count").
		Group("tier_id").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int64, len(results))
	for _, r := range results {
		counts[r.TierId] = r.Count
	}
	return counts, nil
}

// SpecialTierStock gives each tier a stable database row to lock while checking
// the global purchase count, including across independent gateway instances.
type SpecialTierStock struct {
	TierId string `json:"tier_id" gorm:"type:varchar(64);primaryKey"`
}
