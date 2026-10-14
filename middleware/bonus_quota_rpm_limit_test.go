package middleware

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/setting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveBonusQuotaModelRpmLimit(t *testing.T) {
	original := setting.BonusQuotaModelRpmLimit
	t.Cleanup(func() {
		setting.BonusQuotaModelRpmLimit = original
	})

	zero := 0
	three := 3
	five := 5

	tests := []struct {
		name   string
		global int
		user   *int
		want   int
	}{
		{name: "no global default and no override means unlimited", global: 0, user: nil, want: 0},
		{name: "global default applies when user has no override", global: 3, user: nil, want: 3},
		{name: "user override wins over global default", global: 3, user: &five, want: 5},
		{name: "explicit zero override lifts the global default", global: 3, user: &zero, want: 0},
		{name: "user override applies without a global default", global: 0, user: &three, want: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setting.BonusQuotaModelRpmLimit = tt.global
			got := ResolveBonusQuotaModelRpmLimit(dto.UserSetting{BonusQuotaModelRpm: tt.user})
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAllowBonusQuotaModelRpmRequest(t *testing.T) {
	redisEnabled := common.RedisEnabled
	common.RedisEnabled = false
	t.Cleanup(func() {
		common.RedisEnabled = redisEnabled
	})

	t.Run("disabled limit always allows", func(t *testing.T) {
		for range 10 {
			require.True(t, AllowBonusQuotaModelRpmRequest(910001, 0))
		}
	})

	t.Run("rejects requests over the limit within the window", func(t *testing.T) {
		for range 3 {
			require.True(t, AllowBonusQuotaModelRpmRequest(910002, 3))
		}
		assert.False(t, AllowBonusQuotaModelRpmRequest(910002, 3))
		assert.False(t, AllowBonusQuotaModelRpmRequest(910002, 3))
	})

	t.Run("counting is per user", func(t *testing.T) {
		require.True(t, AllowBonusQuotaModelRpmRequest(910003, 1))
		require.False(t, AllowBonusQuotaModelRpmRequest(910003, 1))
		// 另一个用户有独立的窗口
		require.True(t, AllowBonusQuotaModelRpmRequest(910004, 1))
	})
}

func TestResolveBonusQuotaModelRpmByModelLimit(t *testing.T) {
	require.NoError(t, setting.UpdateBonusQuotaModelRpmByModelByJSONString(`{"frame":3,"free":1}`))
	t.Cleanup(func() {
		_ = setting.UpdateBonusQuotaModelRpmByModelByJSONString(`{}`)
	})

	tests := []struct {
		name    string
		userMap map[string]int
		model   string
		want    int
	}{
		{
			name:  "global default applies when user has no override",
			model: "frame",
			want:  3,
		},
		{
			name:    "user override wins over the global default",
			userMap: map[string]int{"frame": 5},
			model:   "frame",
			want:    5,
		},
		{
			name:    "explicit zero override lifts the global default",
			userMap: map[string]int{"frame": 0},
			model:   "frame",
			want:    0,
		},
		{
			name:    "each model resolves independently",
			userMap: map[string]int{"frame": 5},
			model:   "free",
			want:    1,
		},
		{
			name:    "model absent from both layers is unlimited",
			userMap: map[string]int{"frame": 5},
			model:   "other",
			want:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveBonusQuotaModelRpmByModelLimit(
				dto.UserSetting{BonusQuotaModelRpmByModel: tt.userMap},
				tt.model,
			)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAllowBonusQuotaModelRpmByModelRequest(t *testing.T) {
	redisEnabled := common.RedisEnabled
	common.RedisEnabled = false
	t.Cleanup(func() {
		common.RedisEnabled = redisEnabled
	})

	t.Run("disabled limit always allows", func(t *testing.T) {
		for range 10 {
			require.True(t, AllowBonusQuotaModelRpmByModelRequest(910101, "frame", 0))
		}
	})

	t.Run("each model has its own window", func(t *testing.T) {
		require.True(t, AllowBonusQuotaModelRpmByModelRequest(910102, "frame", 1))
		assert.False(t, AllowBonusQuotaModelRpmByModelRequest(910102, "frame", 1))
		// frame 打满不应影响同一用户的 free
		assert.True(t, AllowBonusQuotaModelRpmByModelRequest(910102, "free", 1))
	})

	t.Run("each user has its own window", func(t *testing.T) {
		require.True(t, AllowBonusQuotaModelRpmByModelRequest(910103, "frame", 1))
		require.False(t, AllowBonusQuotaModelRpmByModelRequest(910103, "frame", 1))
		assert.True(t, AllowBonusQuotaModelRpmByModelRequest(910104, "frame", 1))
	})

	t.Run("per-model and total gates count independently", func(t *testing.T) {
		require.True(t, AllowBonusQuotaModelRpmByModelRequest(910105, "frame", 1))
		require.False(t, AllowBonusQuotaModelRpmByModelRequest(910105, "frame", 1))
		// 被按模型闸拒绝的那次不应占用合计闸的名额
		for range 3 {
			require.True(t, AllowBonusQuotaModelRpmRequest(910105, 3))
		}
		assert.False(t, AllowBonusQuotaModelRpmRequest(910105, 3))
	})
}

func TestCheckBonusQuotaModelRpmByModelMap(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]int
		wantErr bool
	}{
		{name: "nil map is valid", input: nil},
		{name: "empty map is valid", input: map[string]int{}},
		{name: "non negative limits are valid", input: map[string]int{"frame": 3, "free": 0}},
		{name: "negative limit is rejected", input: map[string]int{"frame": -1}, wantErr: true},
		{name: "empty model name is rejected", input: map[string]int{"": 3}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := setting.CheckBonusQuotaModelRpmByModelMap(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
