package middleware

import (
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"sync"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllowChannelRpmRequest(t *testing.T) {
	redisEnabled := common.RedisEnabled
	common.RedisEnabled = false
	t.Cleanup(func() {
		common.RedisEnabled = redisEnabled
	})

	t.Run("rpm limit disabled always allows", func(t *testing.T) {
		setting := dto.ChannelSettings{RpmLimit: 0}
		for range 10 {
			require.True(t, AllowChannelRpmRequest(900001, setting))
		}
	})

	t.Run("error action rejects all over-limit requests", func(t *testing.T) {
		setting := dto.ChannelSettings{RpmLimit: 3, RpmOverLimitAction: dto.ChannelRpmOverLimitActionError}
		for range 3 {
			require.True(t, AllowChannelRpmRequest(900002, setting))
		}
		assert.False(t, AllowChannelRpmRequest(900002, setting))
		assert.False(t, AllowChannelRpmRequest(900002, setting))
	})

	t.Run("random drop with 100 percent drop rate rejects all over-limit requests", func(t *testing.T) {
		setting := dto.ChannelSettings{
			RpmLimit:           2,
			RpmOverLimitAction: dto.ChannelRpmOverLimitActionRandomDrop,
			RpmDropRate:        100,
		}
		for range 2 {
			require.True(t, AllowChannelRpmRequest(900003, setting))
		}
		assert.False(t, AllowChannelRpmRequest(900003, setting))
	})

	t.Run("counting is per channel", func(t *testing.T) {
		setting := dto.ChannelSettings{RpmLimit: 1, RpmOverLimitAction: dto.ChannelRpmOverLimitActionError}
		require.True(t, AllowChannelRpmRequest(900004, setting))
		require.False(t, AllowChannelRpmRequest(900004, setting))
		// 另一个渠道有独立的窗口
		require.True(t, AllowChannelRpmRequest(900005, setting))
	})
}

func TestChannelRpmRedisWindowReservesAtomically(t *testing.T) {
	server := miniredis.RunT(t)
	oldEnabled, oldRedis := common.RedisEnabled, common.RDB
	common.RedisEnabled = true
	common.RDB = redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = common.RDB.Close(); common.RedisEnabled = oldEnabled; common.RDB = oldRedis })
	settings := dto.ChannelSettings{RpmLimit: 1}
	start := make(chan struct{})
	outcomes := make(chan bool, 2)
	var requests sync.WaitGroup
	for range 2 {
		requests.Go(func() { <-start; outcomes <- AllowChannelRpmRequest(919001, settings) })
	}
	close(start)
	requests.Wait()
	close(outcomes)
	allowed := 0
	for outcome := range outcomes {
		if outcome {
			allowed++
		}
	}
	assert.Equal(t, 1, allowed, "one concurrent request can claim a single remaining RPM slot")
	assert.True(t, AllowChannelRpmRequest(919002, settings), "separate channels have separate slots")
}
