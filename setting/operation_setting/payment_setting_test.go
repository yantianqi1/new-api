package operation_setting

import (
	"github.com/QuantumNous/new-api/setting/config"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpecialTiersConfig(t *testing.T) {
	orig := paymentSetting
	t.Cleanup(func() { paymentSetting = orig })

	paymentSetting.SpecialTiers = []SpecialTier{
		{
			Id:         "tier-1",
			Name:       "新人特惠",
			Price:      9.9,
			Quota:      5000000,
			TotalLimit: 100,
			UserLimit:  1,
			Enabled:    true,
			SortOrder:  1,
		},
		{
			Id:         "tier-2",
			Name:       "已下线活动",
			Price:      19.9,
			Quota:      12000000,
			TotalLimit: 50,
			UserLimit:  1,
			Enabled:    false,
			SortOrder:  2,
		},
	}

	tiers := GetSpecialTiers()
	require.Len(t, tiers, 2)
	assert.Equal(t, "tier-1", tiers[0].Id)
	assert.Equal(t, 9.9, tiers[0].Price)
	assert.Equal(t, int64(5000000), tiers[0].Quota)

	// Test GetSpecialTierById
	t1 := GetSpecialTierById("tier-1")
	require.NotNil(t, t1)
	assert.Equal(t, "新人特惠", t1.Name)

	t2 := GetSpecialTierById("tier-2")
	require.NotNil(t, t2)
	assert.False(t, t2.Enabled)

	t3 := GetSpecialTierById("non-existent")
	assert.Nil(t, t3)
}

func TestSpecialTier_IsStockVisible(t *testing.T) {
	// default nil -> visible
	tierDefault := SpecialTier{Id: "t1"}
	assert.True(t, tierDefault.IsStockVisible())

	// explicitly true -> visible
	showTrue := true
	tierTrue := SpecialTier{Id: "t2", ShowStock: &showTrue}
	assert.True(t, tierTrue.IsStockVisible())

	// explicitly false -> hidden
	showFalse := false
	tierFalse := SpecialTier{Id: "t3", ShowStock: &showFalse}
	assert.False(t, tierFalse.IsStockVisible())
}

func TestReplacingSpecialTiersRestoresOmittedStockVisibility(t *testing.T) {
	settings := PaymentSetting{}
	require.NoError(t, config.UpdateConfigFromMap(&settings, map[string]string{"special_tiers": `[{"id":"old","show_stock":false}]`}))
	require.Len(t, settings.SpecialTiers, 1)
	assert.False(t, settings.SpecialTiers[0].IsStockVisible())
	require.NoError(t, config.UpdateConfigFromMap(&settings, map[string]string{"special_tiers": `[{"id":"new"}]`}))
	require.Len(t, settings.SpecialTiers, 1)
	assert.Nil(t, settings.SpecialTiers[0].ShowStock)
	assert.True(t, settings.SpecialTiers[0].IsStockVisible())
	require.NoError(t, config.UpdateConfigFromMap(&settings, map[string]string{"special_tiers": `[{"id":"invalid"},`}))
	assert.Equal(t, "new", settings.SpecialTiers[0].Id, "invalid JSON must preserve the previously accepted configuration")
}
