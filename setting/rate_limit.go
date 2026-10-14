package setting

import (
	"fmt"
	"math"
	"sync"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/types"
)

// maxRateLimitDurationSeconds is the largest window the count cap is computed
// against (24h). Token-bucket capacity is count*duration; this keeps that
// product inside int64 when the window is at most a day.
const maxRateLimitDurationSeconds = 24 * 60 * 60

// maxModelRequestRateLimitCount is math.MaxInt64 / maxRateLimitDurationSeconds.
// It is the largest count that cannot overflow int64(count)*duration for a
// window of at most 24 hours.
const maxModelRequestRateLimitCount int64 = math.MaxInt64 / maxRateLimitDurationSeconds

var ModelRequestRateLimitEnabled = false
var ModelRequestRateLimitDurationMinutes = 1
var ModelRequestRateLimitCount = 0
var ModelRequestRateLimitSuccessCount = 1000
var ModelRequestRateLimitGroup = map[string][2]int{}
var ModelRequestRateLimitMutex sync.RWMutex

// BonusQuotaModelRpmLimit 是免费积分模型（BonusQuotaModels 白名单）的全局默认每分钟请求数上限，
// 0 表示不限。单个用户可在自己的 setting 里覆盖该值。
var BonusQuotaModelRpmLimit = 0

// bonusQuotaModelRpmByModelMap 是「模型名 -> 每分钟请求数上限」的全局默认表，
// 只对 BonusQuotaModels 白名单内的模型生效。单个用户可按模型覆盖其中的值。
var bonusQuotaModelRpmByModelMap = types.NewRWMap[string, int]()

func BonusQuotaModelRpmByModel2JSONString() string {
	return bonusQuotaModelRpmByModelMap.MarshalJSONString()
}

func UpdateBonusQuotaModelRpmByModelByJSONString(jsonStr string) error {
	if jsonStr == "" {
		jsonStr = "{}"
	}
	return types.LoadFromJsonString(bonusQuotaModelRpmByModelMap, jsonStr)
}

// GetBonusQuotaModelRpmByModel 返回该模型配置的全局默认 RPM 上限，未配置时 found 为 false。
func GetBonusQuotaModelRpmByModel(modelName string) (limit int, found bool) {
	return bonusQuotaModelRpmByModelMap.Get(modelName)
}

// CheckBonusQuotaModelRpmByModel 校验按模型 RPM 表的 JSON 文本。
func CheckBonusQuotaModelRpmByModel(jsonStr string) error {
	rpmByModel := make(map[string]int)
	if err := common.Unmarshal([]byte(jsonStr), &rpmByModel); err != nil {
		return fmt.Errorf("invalid rpm by model json: %w", err)
	}
	return CheckBonusQuotaModelRpmByModelMap(rpmByModel)
}

// CheckBonusQuotaModelRpmByModelMap 校验按模型 RPM 表的取值。
// 模型名不做白名单校验：白名单会变动，硬校验会让配置在模型被移出白名单时变成非法。
func CheckBonusQuotaModelRpmByModelMap(rpmByModel map[string]int) error {
	for modelName, limit := range rpmByModel {
		if modelName == "" {
			return fmt.Errorf("model name cannot be empty in rpm by model")
		}
		if limit < 0 {
			return fmt.Errorf("model %s has negative rpm limit: %d", modelName, limit)
		}
		if limit > math.MaxInt32 {
			return fmt.Errorf("model %s has max rpm limit value %d", modelName, math.MaxInt32)
		}
	}
	return nil
}

func ModelRequestRateLimitGroup2JSONString() string {
	ModelRequestRateLimitMutex.RLock()
	defer ModelRequestRateLimitMutex.RUnlock()

	jsonBytes, err := common.Marshal(ModelRequestRateLimitGroup)
	if err != nil {
		common.SysLog(common.LogText("error marshalling model ratio: %s", err.Error()))
	}
	return string(jsonBytes)
}

func UpdateModelRequestRateLimitGroupByJSONString(jsonStr string) error {
	ModelRequestRateLimitMutex.RLock()
	defer ModelRequestRateLimitMutex.RUnlock()

	ModelRequestRateLimitGroup = make(map[string][2]int)
	return common.Unmarshal([]byte(jsonStr), &ModelRequestRateLimitGroup)
}

func GetGroupRateLimit(group string) (totalCount, successCount int, found bool) {
	ModelRequestRateLimitMutex.RLock()
	defer ModelRequestRateLimitMutex.RUnlock()

	if ModelRequestRateLimitGroup == nil {
		return 0, 0, false
	}

	limits, found := ModelRequestRateLimitGroup[group]
	if !found {
		return 0, 0, false
	}
	return limits[0], limits[1], true
}

func CheckModelRequestRateLimitGroup(jsonStr string) error {
	checkModelRequestRateLimitGroup := make(map[string][2]int)
	err := common.Unmarshal([]byte(jsonStr), &checkModelRequestRateLimitGroup)
	if err != nil {
		return err
	}
	for group, limits := range checkModelRequestRateLimitGroup {
		if limits[0] < 0 || limits[1] < 1 {
			return fmt.Errorf("group %s has negative rate limit values: [%d, %d]", group, limits[0], limits[1])
		}
		if int64(limits[0]) > maxModelRequestRateLimitCount || int64(limits[1]) > maxModelRequestRateLimitCount {
			return fmt.Errorf("group %s [%d, %d] exceeds max rate limit %d", group, limits[0], limits[1], maxModelRequestRateLimitCount)
		}
	}

	return nil
}
