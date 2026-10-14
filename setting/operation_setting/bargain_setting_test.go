package operation_setting

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBargainSettingDefaults(t *testing.T) {
	setting := GetBargainSetting()
	require.NotNil(t, setting)
	assert.False(t, IsBargainEnabled())
	assert.Equal(t, 5.0, setting.OriginPrice)
	assert.Equal(t, 0.5, setting.CutAmount)
	assert.Equal(t, 6, setting.MaxSlashes)
	assert.Equal(t, 2.0, setting.MinPrice)
	assert.Equal(t, int64(500000), setting.Quota)
	assert.Equal(t, 1, setting.UserLimit)
	assert.Equal(t, 24, setting.IpLimitHours)
}
