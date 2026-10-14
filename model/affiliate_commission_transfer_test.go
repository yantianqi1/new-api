package model

import (
	"sync"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAffiliateCommissionTransferPartialAndAll(t *testing.T) {
	testCases := []struct {
		name        string
		amount      *int
		transferAll bool
		pending     int
		wantAmount  int
	}{
		{name: "partial", amount: func() *int { value := int(common.QuotaPerUnit); return &value }(), pending: 3 * int(common.QuotaPerUnit), wantAmount: int(common.QuotaPerUnit)},
		{name: "all", transferAll: true, pending: 3 * int(common.QuotaPerUnit), wantAmount: 3 * int(common.QuotaPerUnit)},
	}

	for index, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			truncateTables(t)
			userID := 3000 + index
			require.NoError(t, DB.Create(&User{
				Id:                 userID,
				Username:           tc.name + "_commission_transfer",
				AffCode:            tc.name + "_transfer",
				Quota:              25,
				BonusQuota:         40,
				AffCommissionQuota: tc.pending,
			}).Error)

			transfer, err := TransferAffiliateCommission(userID, tc.amount, tc.transferAll)
			require.NoError(t, err)
			assert.Equal(t, tc.wantAmount, transfer.Quota)
			assert.Equal(t, tc.pending, transfer.PendingQuotaBefore)
			assert.Equal(t, tc.pending-tc.wantAmount, transfer.PendingQuotaAfter)
			assert.Equal(t, 25, transfer.PaidQuotaBefore)
			assert.Equal(t, 25+tc.wantAmount, transfer.PaidQuotaAfter)

			var user User
			require.NoError(t, DB.First(&user, userID).Error)
			assert.Equal(t, tc.pending-tc.wantAmount, user.AffCommissionQuota)
			assert.Equal(t, 25+tc.wantAmount, user.Quota)
			assert.Equal(t, 40, user.BonusQuota)
		})
	}
}

func TestAffiliateCommissionTransferValidation(t *testing.T) {
	truncateTables(t)
	minimum := int(common.QuotaPerUnit)
	require.NoError(t, DB.Create(&User{Id: 3100, Username: "transfer_validation", AffCode: "transfer_validation", AffCommissionQuota: minimum}).Error)
	tooSmall := minimum - 1
	tooLarge := minimum + 1
	valid := minimum

	testCases := []struct {
		name        string
		amount      *int
		transferAll bool
	}{
		{name: "missing amount"},
		{name: "ambiguous", amount: &valid, transferAll: true},
		{name: "below minimum", amount: &tooSmall},
		{name: "insufficient", amount: &tooLarge},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := TransferAffiliateCommission(3100, tc.amount, tc.transferAll)
			require.Error(t, err)
		})
	}

	var user User
	require.NoError(t, DB.First(&user, 3100).Error)
	assert.Equal(t, minimum, user.AffCommissionQuota)
	assert.Zero(t, user.Quota)
	var count int64
	require.NoError(t, DB.Model(&AffiliateCommissionTransfer{}).Count(&count).Error)
	assert.Zero(t, count)
}

func TestAffiliateCommissionTransferConcurrentRequestsDoNotOverdraw(t *testing.T) {
	truncateTables(t)
	minimum := int(common.QuotaPerUnit)
	require.NoError(t, DB.Create(&User{Id: 3200, Username: "transfer_concurrent", AffCode: "transfer_concurrent", AffCommissionQuota: minimum}).Error)

	start := make(chan struct{})
	errorsByRequest := make(chan error, 2)
	var waitGroup sync.WaitGroup
	for range 2 {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-start
			_, err := TransferAffiliateCommission(3200, &minimum, false)
			errorsByRequest <- err
		}()
	}
	close(start)
	waitGroup.Wait()
	close(errorsByRequest)

	successes := 0
	for err := range errorsByRequest {
		if err == nil {
			successes++
		}
	}
	assert.Equal(t, 1, successes)

	var user User
	require.NoError(t, DB.First(&user, 3200).Error)
	assert.Zero(t, user.AffCommissionQuota)
	assert.Equal(t, minimum, user.Quota)
	var count int64
	require.NoError(t, DB.Model(&AffiliateCommissionTransfer{}).Count(&count).Error)
	assert.EqualValues(t, 1, count)
}

func TestAffiliateCommissionQueriesAreOwnerScoped(t *testing.T) {
	truncateTables(t)
	require.NoError(t, DB.Create(&User{Id: 11, Username: "query_owner", AffCode: "query_owner", AffCommissionQuota: 25, AffCommissionHistoryQuota: 50}).Error)
	require.NoError(t, DB.Create(&User{Id: 3301, Username: "eligible_one", AffCode: "eligible_one", InviterId: 11, AffCommissionEligible: true}).Error)
	require.NoError(t, DB.Create(&User{Id: 3302, Username: "legacy_one", AffCode: "legacy_one", InviterId: 11, AffCommissionEligible: false}).Error)
	require.NoError(t, DB.Create(&AffiliateCommission{TopUpId: 7001, TradeNo: "owner-11", InviterId: 11, InviteeId: 3301, Status: AffiliateCommissionStatusAwarded}).Error)
	require.NoError(t, DB.Create(&AffiliateCommission{TopUpId: 7002, TradeNo: "owner-12", InviterId: 12, InviteeId: 3302, Status: AffiliateCommissionStatusAwarded}).Error)

	summary, err := GetAffiliateCommissionSummary(11)
	require.NoError(t, err)
	assert.EqualValues(t, 1, summary.EligibleInviteeCount)

	records, total, err := GetUserAffiliateCommissions(11, &common.PageInfo{Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.EqualValues(t, 1, total)
	require.Len(t, records, 1)
	assert.Equal(t, "owner-11", records[0].TradeNo)
}
