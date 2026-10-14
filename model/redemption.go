package model

import (
	"errors"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	RedemptionQuotaTypePaid  = "paid"
	RedemptionQuotaTypeBonus = "bonus"
	RedemptionMaxUsesLimit   = 100000
)

type Redemption struct {
	Id           int            `json:"id"`
	UserId       int            `json:"user_id"`
	Key          string         `json:"key" gorm:"type:char(32);uniqueIndex"`
	Status       int            `json:"status" gorm:"default:1"`
	Name         string         `json:"name" gorm:"index"`
	Quota        int            `json:"quota" gorm:"type:bigint;default:100"`
	QuotaType    string         `json:"quota_type" gorm:"type:varchar(16);default:'paid'"`
	CreatedTime  int64          `json:"created_time" gorm:"bigint"`
	RedeemedTime int64          `json:"redeemed_time" gorm:"bigint"`
	Count        int            `json:"count" gorm:"-:all"` // only for api request
	MaxUses      int            `json:"max_uses"`
	UsedCount    int            `json:"used_count"`
	UsedUserId   int            `json:"used_user_id"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	ExpiredTime  int64          `json:"expired_time" gorm:"bigint"` // 过期时间，0 表示不过期
}

type RedemptionRecord struct {
	Id           int    `json:"id"`
	RedemptionId int    `json:"redemption_id" gorm:"uniqueIndex:idx_redemption_user"`
	UserId       int    `json:"user_id" gorm:"uniqueIndex:idx_redemption_user;index"`
	Quota        int    `json:"quota" gorm:"type:bigint"`
	QuotaType    string `json:"quota_type" gorm:"type:varchar(16)"`
	RedeemedTime int64  `json:"redeemed_time" gorm:"bigint"`
}

func (redemption *Redemption) NormalizeQuotaType() {
	if redemption.QuotaType == RedemptionQuotaTypeBonus {
		return
	}
	redemption.QuotaType = RedemptionQuotaTypePaid
}

func (redemption *Redemption) NormalizeUsageLimits() {
	if redemption.MaxUses <= 0 {
		redemption.MaxUses = 1
	}
	if redemption.Status == common.RedemptionCodeStatusUsed && redemption.UsedCount == 0 {
		redemption.UsedCount = 1
	}
}

func backfillRedemptionUsageLimits() error {
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&Redemption{}).Where("max_uses IS NULL OR max_uses <= ?", 0).Update("max_uses", 1).Error; err != nil {
			return err
		}
		if err := tx.Model(&Redemption{}).Where("(used_count IS NULL OR used_count = ?) AND status = ? AND max_uses = ?", 0, common.RedemptionCodeStatusUsed, 1).Update("used_count", 1).Error; err != nil {
			return err
		}
		if err := tx.Model(&Redemption{}).Where("used_count IS NULL").Update("used_count", 0).Error; err != nil {
			return err
		}
		// Legacy single-use codes only stored their last redeemer. Preserve that
		// user's claim when an administrator later increases the code's capacity.
		var redemptions []Redemption
		return tx.Where("used_user_id > ? AND used_count > ? AND quota > ?", 0, 0, 0).FindInBatches(&redemptions, 100, func(batch *gorm.DB, _ int) error {
			records := make([]RedemptionRecord, 0, len(redemptions))
			for _, redemption := range redemptions {
				redemption.NormalizeQuotaType()
				records = append(records, RedemptionRecord{RedemptionId: redemption.Id, UserId: redemption.UsedUserId, Quota: redemption.Quota, QuotaType: redemption.QuotaType, RedeemedTime: redemption.RedeemedTime})
			}
			if len(records) == 0 {
				return nil
			}
			return batch.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "redemption_id"}, {Name: "user_id"}}, DoNothing: true}).Create(&records).Error
		}).Error
	})
}

func GetAllRedemptions(startIdx int, num int) (redemptions []*Redemption, total int64, err error) {
	// 开始事务
	tx := DB.Begin()
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 获取总数
	err = tx.Model(&Redemption{}).Count(&total).Error
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	// 获取分页数据
	err = tx.Order("id desc").Limit(num).Offset(startIdx).Find(&redemptions).Error
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	// 提交事务
	if err = tx.Commit().Error; err != nil {
		return nil, 0, err
	}

	for _, redemption := range redemptions {
		redemption.NormalizeUsageLimits()
		redemption.NormalizeQuotaType()
	}
	return redemptions, total, nil
}

func SearchRedemptions(keyword string, status string, startIdx int, num int) (redemptions []*Redemption, total int64, err error) {
	tx := DB.Begin()
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	query := tx.Model(&Redemption{})

	if keyword != "" {
		if id, err := strconv.Atoi(keyword); err == nil {
			query = query.Where("id = ? OR name LIKE ?", id, keyword+"%")
		} else {
			query = query.Where("name LIKE ?", keyword+"%")
		}
	}

	if status != "" {
		now := common.GetTimestamp()
		switch status {
		case "expired":
			query = query.Where(
				"status = ? AND expired_time != 0 AND expired_time < ?",
				common.RedemptionCodeStatusEnabled,
				now,
			)
		case strconv.Itoa(common.RedemptionCodeStatusEnabled):
			query = query.Where(
				"status = ? AND (expired_time = 0 OR expired_time >= ?)",
				common.RedemptionCodeStatusEnabled,
				now,
			)
		case strconv.Itoa(common.RedemptionCodeStatusDisabled):
			query = query.Where("status = ?", common.RedemptionCodeStatusDisabled)
		case strconv.Itoa(common.RedemptionCodeStatusUsed):
			query = query.Where("status = ?", common.RedemptionCodeStatusUsed)
		}
	}

	// Get total count
	err = query.Count(&total).Error
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	// Get paginated data
	err = query.Order("id desc").Limit(num).Offset(startIdx).Find(&redemptions).Error
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	if err = tx.Commit().Error; err != nil {
		return nil, 0, err
	}

	for _, redemption := range redemptions {
		redemption.NormalizeUsageLimits()
		redemption.NormalizeQuotaType()
	}
	return redemptions, total, nil
}

func GetRedemptionById(id int) (*Redemption, error) {
	if id == 0 {
		return nil, common.NewMessage("ID is empty")
	}
	redemption := Redemption{Id: id}
	var err error = nil
	err = DB.First(&redemption, "id = ?", id).Error
	redemption.NormalizeUsageLimits()
	redemption.NormalizeQuotaType()
	return &redemption, err
}

func Redeem(key string, userId int) (quota int, err error) {
	if key == "" {
		return 0, errors.New("redemption code not provided")
	}
	if userId == 0 {
		return 0, errors.New("invalid user id")
	}
	redemption := &Redemption{}

	keyCol := "`key`"
	if common.UsingMainDatabase(common.DatabaseTypePostgreSQL) {
		keyCol = `"key"`
	}
	common.RandomSleep()
	err = DB.Transaction(func(tx *gorm.DB) error {
		err := lockForUpdate(tx).Where(keyCol+" = ?", key).First(redemption).Error
		if err != nil {
			return errors.New("invalid redemption code")
		}
		redemption.NormalizeUsageLimits()
		redemption.NormalizeQuotaType()
		if redemption.Status != common.RedemptionCodeStatusEnabled || redemption.UsedCount >= redemption.MaxUses {
			return errors.New("redemption code already used")
		}
		if redemption.ExpiredTime != 0 && redemption.ExpiredTime < common.GetTimestamp() {
			return errors.New("redemption code expired")
		}
		now := common.GetTimestamp()
		if err := tx.Create(&RedemptionRecord{RedemptionId: redemption.Id, UserId: userId, Quota: redemption.Quota, QuotaType: redemption.QuotaType, RedeemedTime: now}).Error; err != nil {
			return err
		}
		nextStatus := common.RedemptionCodeStatusEnabled
		if redemption.UsedCount+1 >= redemption.MaxUses {
			nextStatus = common.RedemptionCodeStatusUsed
		}
		// Compare-and-swap protects capacity even on databases without row locks.
		result := tx.Model(&Redemption{}).Where("id = ? AND status = ? AND COALESCE(used_count, 0) = ?", redemption.Id, common.RedemptionCodeStatusEnabled, redemption.UsedCount).Updates(map[string]any{"redeemed_time": now, "status": nextStatus, "used_user_id": userId, "max_uses": redemption.MaxUses, "used_count": redemption.UsedCount + 1})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return errors.New("redemption code already used")
		}
		if redemption.QuotaType == RedemptionQuotaTypeBonus {
			return creditUserBonusQuota(tx, userId, redemption.Quota)
		}
		return creditTopUpQuota(tx, userId, redemption.Quota, nil)
	})
	if err != nil {
		common.SysError(common.LogText("redemption failed: %s", err.Error()))
		return 0, ErrRedeemFailed
	}
	if redemption.QuotaType == RedemptionQuotaTypeBonus {
		syncCreditUserBonusQuotaCache(userId, redemption.Quota, "redemption")
	} else {
		syncCreditUserQuotaCache(userId, redemption.Quota, common.LogText("redemption"))
	}
	RecordLog(userId, LogTypeTopup, common.NewMessage("Topped up {{quota}} with a redemption code, code ID {{id}}", map[string]any{"quota": logger.FormatQuota(redemption.Quota), "id": redemption.Id}))
	return redemption.Quota, nil
}

func (redemption *Redemption) Insert() error {
	if redemption.Quota <= 0 {
		return errors.New("redemption quota must be positive")
	}
	if err := common.ValidateWalletQuota(redemption.Quota); err != nil {
		return err
	}
	redemption.NormalizeQuotaType()
	redemption.NormalizeUsageLimits()
	if redemption.MaxUses > RedemptionMaxUsesLimit {
		return errors.New("invalid redemption max uses")
	}
	var err error
	err = DB.Create(redemption).Error
	return err
}

func (redemption *Redemption) SelectUpdate() error {
	// This can update zero values
	return DB.Model(redemption).Select("redeemed_time", "status").Updates(redemption).Error
}

// Update Make sure your token's fields is completed, because this will update non-zero values
func (redemption *Redemption) Update() error {
	if redemption.Quota <= 0 {
		return errors.New("redemption quota must be positive")
	}
	if err := common.ValidateWalletQuota(redemption.Quota); err != nil {
		return err
	}
	redemption.NormalizeQuotaType()
	redemption.NormalizeUsageLimits()
	if redemption.MaxUses > RedemptionMaxUsesLimit || redemption.MaxUses < redemption.UsedCount {
		return errors.New("invalid redemption max uses")
	}
	var err error
	err = DB.Model(redemption).Select("name", "status", "quota", "quota_type", "redeemed_time", "expired_time", "max_uses").Updates(redemption).Error
	return err
}

func (redemption *Redemption) Delete() error {
	var err error
	err = DB.Delete(redemption).Error
	return err
}

func DeleteRedemptionById(id int) (err error) {
	if id == 0 {
		return common.NewMessage("ID is empty")
	}
	redemption := Redemption{Id: id}
	err = DB.Where(redemption).First(&redemption).Error
	if err != nil {
		return err
	}
	return redemption.Delete()
}

func DeleteInvalidRedemptions() (int64, error) {
	now := common.GetTimestamp()
	result := DB.Where("status IN ? OR (status = ? AND expired_time != 0 AND expired_time < ?)", []int{common.RedemptionCodeStatusUsed, common.RedemptionCodeStatusDisabled}, common.RedemptionCodeStatusEnabled, now).Delete(&Redemption{})
	return result.RowsAffected, result.Error
}

// BatchDeleteRedemptions soft-deletes the selected codes in one statement.
func BatchDeleteRedemptions(ids []int) (int64, error) {
	if len(ids) == 0 || len(ids) > 1000 {
		return 0, errors.New("select between 1 and 1000 redemption codes")
	}
	for _, id := range ids {
		if id <= 0 {
			return 0, errors.New("redemption IDs must be positive")
		}
	}
	result := DB.Where("id IN ?", ids).Delete(&Redemption{})
	return result.RowsAffected, result.Error
}
