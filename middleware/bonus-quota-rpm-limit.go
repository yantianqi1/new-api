package middleware

import (
	"fmt"

	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/setting"
)

// bonusQuotaModelRpmKeyPrefix 是用户级免费积分模型 RPM 合计计数在 Redis / 内存限流器中的 key 前缀。
const bonusQuotaModelRpmKeyPrefix = "rateLimit:bonusQuotaModelRpm:"

// bonusQuotaModelRpmByModelKeyPrefix 是按模型的免费积分模型 RPM 计数 key 前缀。
const bonusQuotaModelRpmByModelKeyPrefix = "rateLimit:bonusQuotaModelRpmByModel:"

// ResolveBonusQuotaModelRpmLimit 解析用户实际生效的免费积分模型合计 RPM 上限，0 表示不限。
// 用户在 setting 里有显式覆盖值（含 0）时优先使用，否则回落到全局默认设置。
func ResolveBonusQuotaModelRpmLimit(userSetting dto.UserSetting) int {
	if userSetting.BonusQuotaModelRpm != nil {
		return *userSetting.BonusQuotaModelRpm
	}
	return setting.BonusQuotaModelRpmLimit
}

// ResolveBonusQuotaModelRpmByModelLimit 解析用户对指定模型实际生效的 RPM 上限，0 表示不限。
// 优先级：用户按模型覆盖 -> 全局按模型默认 -> 不限。
func ResolveBonusQuotaModelRpmByModelLimit(userSetting dto.UserSetting, modelName string) int {
	if limit, ok := userSetting.BonusQuotaModelRpmByModel[modelName]; ok {
		return limit
	}
	if limit, found := setting.GetBonusQuotaModelRpmByModel(modelName); found {
		return limit
	}
	return 0
}

// AllowBonusQuotaModelRpmRequest 检查并记录一次针对该用户的免费积分模型合计请求。
func AllowBonusQuotaModelRpmRequest(userId int, limit int) bool {
	return slidingWindowAllows(fmt.Sprintf("%s%d", bonusQuotaModelRpmKeyPrefix, userId), limit)
}

// AllowBonusQuotaModelRpmByModelRequest 检查并记录一次针对该用户该模型的请求。
func AllowBonusQuotaModelRpmByModelRequest(userId int, modelName string, limit int) bool {
	key := fmt.Sprintf("%s%d:%s", bonusQuotaModelRpmByModelKeyPrefix, userId, modelName)
	return slidingWindowAllows(key, limit)
}
