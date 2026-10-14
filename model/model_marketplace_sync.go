package model

import (
	"errors"
	"sort"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

// MarketplaceSyncResult summarizes a model marketplace synchronization.
type MarketplaceSyncResult struct {
	CreatedModels  int `json:"created_models"`
	ExistingModels int `json:"existing_models"`
	TotalModels    int `json:"total_models"`
}

// SyncModelsFromMarketplace registers every currently enabled marketplace model.
// Existing metadata is never modified, so pricing and status choices are preserved.
func SyncModelsFromMarketplace() (MarketplaceSyncResult, error) {
	names := GetEnabledModels()
	sort.Strings(names)
	result := MarketplaceSyncResult{TotalModels: len(names)}
	if len(names) == 0 {
		return result, nil
	}

	tx := DB.Begin()
	if tx.Error != nil {
		return result, tx.Error
	}
	for _, name := range names {
		var existing Model
		err := tx.Where("model_name = ?", name).First(&existing).Error
		if err == nil {
			result.ExistingModels++
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return result, err
		}
		created := &Model{ModelName: name, Status: 1, SyncOfficial: 0, CreatedTime: common.GetTimestamp(), UpdatedTime: common.GetTimestamp()}
		if err = tx.Create(created).Error; err != nil {
			tx.Rollback()
			return result, err
		}
		if err = tx.Model(&Model{}).Where("id = ?", created.Id).Updates(map[string]any{"status": 1, "sync_official": 0}).Error; err != nil {
			tx.Rollback()
			return result, err
		}
		result.CreatedModels++
	}
	if err := tx.Commit().Error; err != nil {
		return result, err
	}
	RefreshPricing()
	return result, nil
}
