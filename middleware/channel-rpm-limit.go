package middleware

import (
	"fmt"
	"math/rand"

	"github.com/QuantumNous/new-api/relaykit/dto"
)

// channelRpmKeyPrefix 是渠道 RPM 计数在 Redis / 内存限流器中的 key 前缀。
const channelRpmKeyPrefix = "rateLimit:channelRpm:"

// AllowChannelRpmRequest 检查并记录一次针对该渠道的请求。
// 渠道未配置 RpmLimit（<= 0）时直接放行；
// 超限时按渠道设置处理：error 策略拒绝全部超限请求，
// random_drop 策略按 RpmDropRate 概率随机丢弃（429），其余放行。
func AllowChannelRpmRequest(channelId int, setting dto.ChannelSettings) bool {
	if setting.RpmLimit <= 0 {
		return true
	}
	if slidingWindowAllows(fmt.Sprintf("%s%d", channelRpmKeyPrefix, channelId), setting.RpmLimit) {
		return true
	}
	return overLimitDecision(setting)
}

// overLimitDecision 在滑动窗口已满时根据渠道配置决定放行与否。
func overLimitDecision(setting dto.ChannelSettings) bool {
	if setting.RpmOverLimitAction != dto.ChannelRpmOverLimitActionRandomDrop {
		return false
	}
	dropRate := setting.RpmDropRate
	if dropRate <= 0 {
		dropRate = dto.DefaultChannelRpmDropRate
	}
	if dropRate >= 100 {
		return false
	}
	// 命中丢弃概率则拒绝，否则放行（概率限流）
	return rand.Intn(100) >= dropRate
}
