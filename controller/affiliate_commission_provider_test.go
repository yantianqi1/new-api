package controller

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAffiliateCommissionProviderSnapshotUsesActualCharge(t *testing.T) {
	snapshot, err := buildTopUpCommissionSnapshot(
		decimal.RequireFromString("90.00"),
		"cny",
		decimal.RequireFromString("1.00"),
	)
	require.NoError(t, err)
	assert.Equal(t, "90", snapshot.ActualPaidAmount)
	assert.Equal(t, "CNY", snapshot.Currency)
	assert.Equal(t, 90*int(common.QuotaPerUnit), snapshot.CommissionBaseQuota)
}

func TestAffiliateCommissionProviderSnapshotUsesNormalizedAmountString(t *testing.T) {
	testCases := []struct {
		name       string
		amount     string
		currency   string
		wantAmount string
	}{
		{name: "two decimal currency", amount: "9.99", currency: "USD", wantAmount: "9.99"},
		{name: "zero decimal currency", amount: "10", currency: "JPY", wantAmount: "10"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			snapshot, err := buildTopUpCommissionSnapshotFromString(tc.amount, tc.currency, decimal.NewFromInt(1))
			require.NoError(t, err)
			assert.Equal(t, tc.wantAmount, snapshot.ActualPaidAmount)
		})
	}
}

func TestAffiliateCommissionProviderSnapshotRejectsInvalidAmountString(t *testing.T) {
	_, err := buildTopUpCommissionSnapshotFromString("not-money", "USD", decimal.NewFromInt(1))
	require.Error(t, err)
}

func TestAffiliateCommissionProviderSnapshotRejectsInvalidPricing(t *testing.T) {
	testCases := []struct {
		name      string
		amount    decimal.Decimal
		unitPrice decimal.Decimal
	}{
		{name: "zero amount", amount: decimal.Zero, unitPrice: decimal.NewFromInt(1)},
		{name: "negative amount", amount: decimal.NewFromInt(-1), unitPrice: decimal.NewFromInt(1)},
		{name: "zero unit price", amount: decimal.NewFromInt(1), unitPrice: decimal.Zero},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := buildTopUpCommissionSnapshot(tc.amount, "USD", tc.unitPrice)
			require.Error(t, err)
		})
	}
}

func TestAffiliateCommissionProviderMinorUnitAmount(t *testing.T) {
	assert.Equal(t, "19.99", minorUnitAmountString(1999, "USD"))
	assert.Equal(t, "1999", minorUnitAmountString(1999, "JPY"))
}

func TestCreemCommissionBaseQuota(t *testing.T) {
	testCases := []struct {
		name    string
		product CreemProduct
		want    int
	}{
		{name: "explicit paid base", product: CreemProduct{Quota: 5_000_000, CommissionBaseQuota: 4_000_000}, want: 4_000_000},
		{name: "missing base", product: CreemProduct{Quota: 5_000_000}, want: 0},
		{name: "negative base", product: CreemProduct{Quota: 5_000_000, CommissionBaseQuota: -1}, want: 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.product.normalizedCommissionBaseQuota())
		})
	}
}
