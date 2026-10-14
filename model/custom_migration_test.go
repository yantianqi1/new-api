package model

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The environment opts into real database startup and upgrade verification.
// Run against isolated fresh and production-main fixture databases for every dialect.
func TestCustomMigrationDatabase(t *testing.T) {
	if os.Getenv("NEW_API_MIGRATION_MATRIX") != "1" {
		t.Skip("isolated database matrix is opt-in")
	}
	common.IsMasterNode = true
	if path := os.Getenv("SQLITE_PATH"); path != "" {
		common.SQLitePath = path
	}
	require.NoError(t, InitDB())
	if common.UsingMainDatabase(common.DatabaseTypeSQLite) && os.Getenv("LOG_SQL_DSN") != "" {
		common.SQLitePath += ".logs"
	}
	require.NoError(t, InitLogDB())
	for range 2 {
		require.NoError(t, ensureUserQuotaColumns(DB, common.MainDatabaseType()))
		require.NoError(t, migrateDB())
		require.NoError(t, migrateLOGDB())
	}
	if os.Getenv("NEW_API_MIGRATION_UPGRADE") == "1" {
		var user User
		require.NoError(t, DB.Where("username = ?", "migration-fixture").First(&user).Error)
		assert.Equal(t, 1200, user.Quota)
		assert.Equal(t, 3400, user.BonusQuota)
		assert.Equal(t, 50, user.AffCommissionQuota)
		assert.Equal(t, "迁移测试", user.DisplayName)
		var code Redemption
		require.NoError(t, DB.Where(commonKeyCol+" = ?", "legacy-paid-used-fixture").First(&code).Error)
		assert.Equal(t, 3, code.MaxUses)
		assert.Equal(t, 1, code.UsedCount)
		var legacyRecord RedemptionRecord
		require.NoError(t, DB.Where("redemption_id = ? AND user_id = ?", code.Id, user.Id).First(&legacyRecord).Error)
		var option Option
		require.NoError(t, DB.Where(commonKeyCol+" = ?", "BonusQuotaModels").First(&option).Error)
		assert.JSONEq(t, `{"fixture-model":true}`, option.Value)
	}
	for _, entity := range []any{&RedemptionRecord{}, &AffiliateCommission{}, &AffiliateCommissionTransfer{}, &SpecialTierPurchase{}, &SpecialTierStock{}, &BargainActivity{}, &BargainRecord{}, &OIDCClient{}, &OIDCAuthorizationCode{}, &OIDCAccessToken{}, &OIDCSigningKey{}} {
		assert.True(t, DB.Migrator().HasTable(entity))
	}
	var existingCode Redemption
	if DB.Where(commonKeyCol+" = ?", "matrix-bonus-code").First(&existingCode).Error == nil {
		require.NoError(t, DB.Where("redemption_id = ?", existingCode.Id).Delete(&RedemptionRecord{}).Error)
		require.NoError(t, DB.Unscoped().Delete(&existingCode).Error)
	}
	var previousUser User
	if DB.Where("username = ?", "matrix-behavior-user").First(&previousUser).Error == nil {
		require.NoError(t, DB.Where("user_id = ?", previousUser.Id).Delete(&Checkin{}).Error)
		require.NoError(t, DB.Unscoped().Delete(&previousUser).Error)
	}
	require.NoError(t, DB.Where("client_id = ?", "matrix-client").Delete(&OIDCClient{}).Error)
	user := User{Username: "matrix-behavior-user", AffCode: "matrix-user-aff", Status: common.UserStatusEnabled, Quota: 100, BonusQuota: 350}
	require.NoError(t, DB.Create(&user).Error)
	bonus, paid, ok, err := TryReserveUserBonusAndPaidQuota(user.Id, 400, true)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, 350, bonus)
	assert.Equal(t, 50, paid)
	require.NoError(t, RefundUserBonusAndPaidQuota(user.Id, bonus, paid))
	require.NoError(t, DB.First(&user, user.Id).Error)
	assert.Equal(t, 100, user.Quota)
	assert.Equal(t, 350, user.BonusQuota)
	_, _, ok, err = TryReserveUserBonusAndPaidQuota(user.Id, 101, false)
	require.NoError(t, err)
	assert.False(t, ok)
	require.NoError(t, DB.Model(&User{}).Where("id = ?", user.Id).Updates(map[string]any{"quota": -10, "bonus_quota": 100}).Error)
	bonus, paid, ok, err = TryReserveUserBonusAndPaidQuota(user.Id, 20, true)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, 20, bonus)
	assert.Zero(t, paid)
	_, _, ok, err = TryReserveUserBonusAndPaidQuota(user.Id, 71, true)
	require.NoError(t, err)
	assert.False(t, ok)
	require.NoError(t, DB.Model(&User{}).Where("id = ?", user.Id).Updates(map[string]any{"quota": 100, "bonus_quota": 350}).Error)
	code := Redemption{Key: "matrix-bonus-code", Name: "matrix", Status: common.RedemptionCodeStatusEnabled, Quota: 20, QuotaType: RedemptionQuotaTypeBonus, MaxUses: 2}
	require.NoError(t, DB.Create(&code).Error)
	quota, err := Redeem(code.Key, user.Id)
	require.NoError(t, err)
	assert.Equal(t, 20, quota)
	_, err = Redeem(code.Key, user.Id)
	require.Error(t, err)
	require.NoError(t, DB.First(&user, user.Id).Error)
	assert.Equal(t, 100, user.Quota)
	assert.Equal(t, 370, user.BonusQuota)
	setting := operation_setting.GetCheckinSetting()
	old := *setting
	t.Cleanup(func() { *setting = old })
	setting.Enabled = true
	setting.MinQuota = 10
	setting.MaxQuota = 10
	_, err = UserCheckin(user.Id)
	require.NoError(t, err)
	require.NoError(t, DB.First(&user, user.Id).Error)
	assert.Equal(t, 100, user.Quota)
	assert.Equal(t, 380, user.BonusQuota)
	require.True(t, DB.Migrator().HasIndex(&RedemptionRecord{}, "idx_redemption_user"))
	require.True(t, DB.Migrator().HasIndex(&Checkin{}, "idx_user_checkin_date"))
	client := OIDCClient{Name: "matrix-app", ClientId: "matrix-client", ClientSecretHash: "fixture", RedirectURIs: "https://example.com/callback", Scopes: "openid", Enabled: true}
	require.NoError(t, CreateOIDCClient(&client))
	duplicate := client
	duplicate.Id = 0
	require.Error(t, CreateOIDCClient(&duplicate))
	// Insert in existing DB fixture, common.QuotaPerUnit=500000, operation settings restored via cleanup.
	// Uses GORM generated IDs and unique names/order strings derived from these IDs.
	payment := operation_setting.GetPaymentSetting()
	oldPayment := *payment
	bargain := operation_setting.GetBargainSetting()
	oldBargain := *bargain
	t.Cleanup(func() { *payment = oldPayment; *bargain = oldBargain })
	payment.AffiliateCommissionRateBasisPoints = 1000
	inviter := &User{Username: fmt.Sprintf("matrix_ci_%d", user.Id), AffCode: fmt.Sprintf("matrix-ci-%d", user.Id), Quota: 25, BonusQuota: 40}
	require.NoError(t, DB.Create(inviter).Error)
	invitee := &User{Username: fmt.Sprintf("matrix_ce_%d", user.Id), AffCode: fmt.Sprintf("matrix-ce-%d", user.Id), InviterId: inviter.Id, AffCommissionEligible: true}
	require.NoError(t, DB.Create(invitee).Error)
	order := &TopUp{UserId: invitee.Id, Amount: 100, Money: 10, ActualPaidAmount: "10.00", Currency: "CNY", CommissionBaseQuota: 10 * int(common.QuotaPerUnit), TradeNo: fmt.Sprintf("matrix-c-%d", invitee.Id), PaymentProvider: PaymentProviderEpay, PaymentMethod: "alipay", Status: common.TopUpStatusPending}
	require.NoError(t, order.Insert())
	result, err := CompleteOnlineTopUp(CompleteOnlineTopUpInput{TradeNo: order.TradeNo, PaymentProvider: PaymentProviderEpay, SettledAmount: "10.00", SettledCurrency: "CNY"})
	require.NoError(t, err)
	require.NotNil(t, result.Commission)
	assert.Equal(t, int(common.QuotaPerUnit), result.Commission.CommissionQuota)
	repeat, err := CompleteOnlineTopUp(CompleteOnlineTopUpInput{TradeNo: order.TradeNo, PaymentProvider: PaymentProviderEpay})
	require.NoError(t, err)
	assert.True(t, repeat.AlreadyCompleted)
	transfer, err := TransferAffiliateCommission(inviter.Id, nil, true)
	require.NoError(t, err)
	assert.Equal(t, int(common.QuotaPerUnit), transfer.Quota)
	var stored User
	require.NoError(t, DB.First(&stored, inviter.Id).Error)
	assert.Equal(t, 25+int(common.QuotaPerUnit), stored.Quota)
	assert.Equal(t, 40, stored.BonusQuota)
	assert.Zero(t, stored.AffCommissionQuota)
	assert.Equal(t, int(common.QuotaPerUnit), stored.AffCommissionHistoryQuota)

	// Special tier paid(1*500000) + gift(100) and stock1; second buyer rejected, untouched pending order.
	tierID := fmt.Sprintf("matrix-tier-%d", invitee.Id)
	payment.SpecialTiers = []operation_setting.SpecialTier{{Id: tierID, Name: "Matrix tier", Price: 1, Quota: 100, Enabled: true, UserLimit: 1, TotalLimit: 1}}
	tierOrder := &TopUp{UserId: invitee.Id, Amount: 1, Money: 1, ActualPaidAmount: "1.00", Currency: "CNY", TradeNo: fmt.Sprintf("matrix-t-%d", invitee.Id), PaymentProvider: PaymentProviderEpay, PaymentMethod: "alipay", Status: common.TopUpStatusPending, SpecialTierId: tierID, SpecialTierQuota: 100}
	require.NoError(t, tierOrder.Insert())
	tierResult, err := CompleteOnlineTopUp(CompleteOnlineTopUpInput{TradeNo: tierOrder.TradeNo, PaymentProvider: PaymentProviderEpay})
	require.NoError(t, err)
	assert.Equal(t, int(common.QuotaPerUnit)+100, tierResult.QuotaAdded)
	secondTier := *tierOrder
	secondTier.Id = 0
	secondTier.UserId = inviter.Id
	secondTier.TradeNo = fmt.Sprintf("matrix-t2-%d", inviter.Id)
	secondTier.Status = common.TopUpStatusPending
	secondTier.CompleteTime = 0
	require.NoError(t, secondTier.Insert())
	_, err = CompleteOnlineTopUp(CompleteOnlineTopUpInput{TradeNo: secondTier.TradeNo, PaymentProvider: PaymentProviderEpay})
	require.Error(t, err)
	count, err := CountTotalSpecialTierPurchases(tierID)
	require.NoError(t, err)
	assert.EqualValues(t, 1, count)

	// One deterministic referral cut and one replay; no log IP history needed, CreatorIp avoids LOG_DB query.
	*bargain = operation_setting.BargainSetting{Enabled: true, OriginPrice: 5, MinPrice: 2, CutAmount: .5, MaxSlashes: 6, Quota: 500000, UserLimit: 1, IpLimitHours: 24}
	activity, err := CreateBargainActivity(inviter.Id, fmt.Sprintf("2001:db8:1::%x", inviter.Id))
	require.NoError(t, err)
	cut, err := TriggerBargainCut(inviter.Id, invitee.Id, fmt.Sprintf("2001:db8:2::%x", invitee.Id))
	require.NoError(t, err)
	assert.True(t, cut)
	again, err := TriggerBargainCut(inviter.Id, invitee.Id, fmt.Sprintf("2001:db8:3::%x", invitee.Id))
	require.NoError(t, err)
	assert.False(t, again)
	storedActivity, err := GetBargainActivityById(activity.Id)
	require.NoError(t, err)
	assert.Equal(t, 4.5, storedActivity.CurrentPrice)
	assert.Equal(t, 1, storedActivity.CutCount)

	// Authorizations can be reduced and revoked from another active browser session.
	owner := &User{Username: fmt.Sprintf("matrix-oidc-owner-%d", user.Id), AffCode: fmt.Sprintf("matrix-oo-%d", user.Id), Role: common.RoleCommonUser, Status: common.UserStatusEnabled, AuthVersion: 1}
	require.NoError(t, DB.Create(owner).Error)
	now := time.Now()
	origin := &UserSession{SID: fmt.Sprintf("matrix-oidc-origin-%d", owner.Id), UserID: owner.Id, Version: 1, UserAuthVersion: 1, Status: UserSessionStatusActive, RefreshHash: strings.Repeat("a", 64), LoginMethod: "password", CreatedAt: now.Unix(), LastActiveAt: now.Unix(), ExpiresAt: now.Add(time.Hour).Unix()}
	current := *origin
	current.SID = fmt.Sprintf("matrix-oidc-current-%d", owner.Id)
	current.RefreshHash = strings.Repeat("b", 64)
	require.NoError(t, DB.Create(origin).Error)
	require.NoError(t, DB.Create(&current).Error)
	identity := AuthSessionIdentity{UserID: owner.Id, SessionID: current.SID, UserAuthVersion: 1, SessionVersion: 1}
	grantClient := &OIDCClient{Name: "Matrix app", ClientId: fmt.Sprintf("matrix-oidc-client-%d", owner.Id), ClientSecretHash: strings.Repeat("h", 64), RedirectURIs: "https://client.example/callback", Scopes: "openid profile email", Enabled: true}
	require.NoError(t, DB.Create(grantClient).Error)
	grantCode := &OIDCAuthorizationCode{CodeHash: fmt.Sprintf("%064d", owner.Id), ClientId: grantClient.ClientId, UserId: owner.Id, RedirectURI: grantClient.RedirectURIs, Scope: grantClient.Scopes, SessionId: origin.SID, UserAuthVersion: 1, SessionVersion: 1, AuthTime: now.Unix(), CodeChallenge: strings.Repeat("a", 43), CodeChallengeMethod: "S256", ExpiresAt: now.Add(5 * time.Minute)}
	require.NoError(t, CreateOIDCAuthorizationCode(grantCode))
	var listed []OIDCAuthorizationCode
	listed, err = ListOIDCGrantCodes(owner.Id, now)
	require.NoError(t, err)
	require.Len(t, listed, 1)
	reduced := "openid"
	require.NoError(t, ChangeOIDCGrant(identity, grantCode.Id, &reduced))
	var storedCode OIDCAuthorizationCode
	require.NoError(t, DB.First(&storedCode, grantCode.Id).Error)
	assert.Equal(t, "openid", storedCode.Scope)
	expanded := "openid email"
	require.Error(t, ChangeOIDCGrant(identity, grantCode.Id, &expanded))
	// A stale scope snapshot cannot claim an authorization code after reduction.
	cas := DB.Model(&OIDCAuthorizationCode{}).Where("id = ? AND scope = ? AND used_at IS NULL AND expires_at > ?", grantCode.Id, grantCode.Scope, now).Update("used_at", now)
	require.NoError(t, cas.Error)
	assert.Zero(t, cas.RowsAffected)
	used := now
	require.NoError(t, DB.Model(&storedCode).Update("used_at", used).Error)
	token := &OIDCAccessToken{TokenHash: fmt.Sprintf("%064d", owner.Id+1000000), AuthorizationCodeId: grantCode.Id, ClientId: grantClient.ClientId, UserId: owner.Id, Scope: reduced, SessionId: origin.SID, UserAuthVersion: 1, SessionVersion: 1, ExpiresAt: now.Add(15 * time.Minute)}
	require.NoError(t, DB.Create(token).Error)
	listed, err = ListOIDCGrantCodes(owner.Id, now)
	require.NoError(t, err)
	require.Len(t, listed, 1)
	require.NoError(t, ChangeOIDCGrant(identity, grantCode.Id, nil))
	require.NoError(t, DB.First(token, token.Id).Error)
	require.NotNil(t, token.RevokedAt)
	require.NoError(t, DB.First(&storedCode, grantCode.Id).Error)
	require.NotNil(t, storedCode.UsedAt)
	assert.False(t, storedCode.ExpiresAt.After(time.Now()))
	listed, err = ListOIDCGrantCodes(owner.Id, time.Now())
	require.NoError(t, err)
	assert.Empty(t, listed)
	// A revoked browser session can no longer modify a grant.
	_, err = RevokeUserSession(owner.Id, current.SID, "matrix")
	require.NoError(t, err)
	require.Error(t, ChangeOIDCGrant(identity, grantCode.Id, &reduced))

	var databaseVersion string
	switch common.MainDatabaseType() {
	case common.DatabaseTypeSQLite:
		require.NoError(t, DB.Raw("SELECT sqlite_version()").Scan(&databaseVersion).Error)
	case common.DatabaseTypeMySQL:
		require.NoError(t, DB.Raw("SELECT VERSION()").Scan(&databaseVersion).Error)
	case common.DatabaseTypePostgreSQL:
		require.NoError(t, DB.Raw("SHOW server_version").Scan(&databaseVersion).Error)
	}
	t.Logf("database version=%s", databaseVersion)
	t.Logf("verified dialect=%s fresh/upgrade=%s; migrations repeated; split reserve/refund, redemption, checkin, uniqueness", common.MainDatabaseType(), os.Getenv("NEW_API_MIGRATION_UPGRADE"))
}
