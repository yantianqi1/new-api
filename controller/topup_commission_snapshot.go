package controller

import (
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/shopspring/decimal"
)

type topUpCommissionSnapshot struct {
	ActualPaidAmount    string
	Currency            string
	CommissionBaseQuota int
}

func buildTopUpCommissionSnapshotFromString(actualAmount string, currency string, unitPrice decimal.Decimal) (topUpCommissionSnapshot, error) {
	normalizedAmount, err := decimal.NewFromString(strings.TrimSpace(actualAmount))
	if err != nil {
		return topUpCommissionSnapshot{}, fmt.Errorf("invalid topup commission pricing snapshot")
	}
	return buildTopUpCommissionSnapshot(normalizedAmount, currency, unitPrice)
}

func buildTopUpCommissionSnapshot(actualAmount decimal.Decimal, currency string, unitPrice decimal.Decimal) (topUpCommissionSnapshot, error) {
	if !actualAmount.IsPositive() || !unitPrice.IsPositive() {
		return topUpCommissionSnapshot{}, fmt.Errorf("invalid topup commission pricing snapshot")
	}
	baseQuota, err := common.WalletQuotaFromDecimalStrict(actualAmount.Div(unitPrice).Mul(decimal.NewFromFloat(common.QuotaPerUnit)).Truncate(0))
	if err != nil || baseQuota <= 0 {
		return topUpCommissionSnapshot{}, fmt.Errorf("invalid topup commission base quota")
	}

	return topUpCommissionSnapshot{
		ActualPaidAmount:    actualAmount.String(),
		Currency:            strings.ToUpper(strings.TrimSpace(currency)),
		CommissionBaseQuota: int(baseQuota),
	}, nil
}

func minorUnitAmountString(amount int64, currency string) string {
	if zeroDecimalCurrencies[strings.ToUpper(strings.TrimSpace(currency))] {
		return decimal.NewFromInt(amount).String()
	}
	return decimal.NewFromInt(amount).Shift(-2).String()
}
