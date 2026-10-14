package model

import (
	"fmt"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestSyncModelsFromMarketplaceIsIdempotentAndPreservesStatus(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	oldDB := DB
	DB = db
	t.Cleanup(func() { DB = oldDB })
	require.NoError(t, db.AutoMigrate(&Ability{}, &Model{}, &Channel{}, &Vendor{}))
	require.NoError(t, db.Create(&Ability{Group: "default", Model: "new-model", ChannelId: 1, Enabled: true}).Error)
	require.NoError(t, db.Create(&Ability{Group: "default", Model: "old-model", ChannelId: 1, Enabled: true}).Error)
	require.NoError(t, db.Create(&Model{ModelName: "old-model", Status: 0}).Error)
	require.NoError(t, db.Model(&Model{}).Where("model_name = ?", "old-model").Update("status", 0).Error)

	result, err := SyncModelsFromMarketplace()
	require.NoError(t, err)
	require.Equal(t, 1, result.CreatedModels)
	require.Equal(t, 1, result.ExistingModels)
	require.Equal(t, 2, result.TotalModels)

	var added, existing Model
	require.NoError(t, db.Where("model_name = ?", "new-model").First(&added).Error)
	require.Equal(t, 1, added.Status)
	require.NoError(t, db.Where("model_name = ?", "old-model").First(&existing).Error)
	require.Equal(t, 0, existing.Status)

	result, err = SyncModelsFromMarketplace()
	require.NoError(t, err)
	require.Equal(t, 0, result.CreatedModels)
	require.Equal(t, 2, result.ExistingModels)
}
