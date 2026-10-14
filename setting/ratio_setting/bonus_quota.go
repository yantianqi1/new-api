package ratio_setting

import "github.com/QuantumNous/new-api/types"

var bonusQuotaModelMap = types.NewRWMap[string, bool]()

func BonusQuotaModels2JSONString() string {
	return bonusQuotaModelMap.MarshalJSONString()
}

func UpdateBonusQuotaModelsByJSONString(jsonStr string) error {
	if jsonStr == "" {
		jsonStr = "{}"
	}
	return types.LoadFromJsonString(bonusQuotaModelMap, jsonStr)
}

func IsBonusQuotaModel(modelName string) bool {
	enabled, ok := bonusQuotaModelMap.Get(modelName)
	return ok && enabled
}

func GetBonusQuotaModelMapCopy() map[string]bool {
	return bonusQuotaModelMap.ReadAll()
}
