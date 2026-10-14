package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/QuantumNous/new-api/common"
)

// rpmWindowSeconds 是 RPM 滑动窗口长度（秒）。
const rpmWindowSeconds = 60

// slidingWindowAllows 检查并记录一次针对 key 的请求。
// limit <= 0 时直接放行；Redis 可用时走 Redis 滑动窗口，否则退化为进程内限流器。
// Redis 出错时按放行处理并记录错误日志，避免 Redis 故障导致被限流方整体不可用。
func slidingWindowAllows(key string, limit int) bool {
	if limit <= 0 {
		return true
	}
	if common.RedisEnabled {
		return slidingWindowAllowsRedis(key, limit)
	}
	// It's safe to call multi times.
	inMemoryRateLimiter.Init(2 * rpmWindowSeconds * time.Second)
	return inMemoryRateLimiter.Request(key, limit, rpmWindowSeconds)
}

func slidingWindowAllowsRedis(key string, limit int) bool {
	ctx := context.Background()
	now := time.Now().UnixMilli()
	// Pruning, counting and recording happen atomically, so concurrent
	// requests cannot all pass the same stale list-length check.
	const script = `
local cutoff = tonumber(ARGV[1]) - tonumber(ARGV[3])
while redis.call('LLEN', KEYS[1]) > 0 do
 local oldest = tonumber(redis.call('LINDEX', KEYS[1], -1))
 if oldest == nil then redis.call('DEL', KEYS[1]); break end
 if oldest > cutoff then break end
 redis.call('RPOP', KEYS[1])
end
if redis.call('LLEN', KEYS[1]) >= tonumber(ARGV[2]) then return 0 end
redis.call('LPUSH', KEYS[1], ARGV[1])
redis.call('PEXPIRE', KEYS[1], tonumber(ARGV[3]) * 2)
return 1`
	allowed, err := common.RDB.Eval(ctx, script, []string{key}, now, limit, rpmWindowSeconds*1000).Int()
	if err != nil {
		common.SysError(fmt.Sprintf("rpm window limiter redis error: key=%s, error=%v", key, err))
		return true
	}
	return allowed == 1
}
