package model

import (
	"github.com/QuantumNous/new-api/common"
	"slices"
	"strings"
	"time"

	"gorm.io/gorm"
)

type OIDCClient struct {
	Id               int       `json:"id" gorm:"primaryKey"`
	Name             string    `json:"name" gorm:"type:varchar(128);not null"`
	ClientId         string    `json:"client_id" gorm:"type:varchar(128);uniqueIndex;not null"`
	ClientSecretHash string    `json:"-" gorm:"type:varchar(255);not null"`
	RedirectURIs     string    `json:"redirect_uris" gorm:"type:text;not null"`
	Scopes           string    `json:"scopes" gorm:"type:varchar(255);not null"`
	Trusted          bool      `json:"trusted"`
	Enabled          bool      `json:"enabled"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type OIDCAuthorizationCode struct {
	Id                  int        `json:"id" gorm:"primaryKey"`
	CodeHash            string     `json:"-" gorm:"type:char(64);uniqueIndex;not null"`
	ClientId            string     `json:"client_id" gorm:"type:varchar(128);index;not null"`
	UserId              int        `json:"user_id" gorm:"index;not null"`
	RedirectURI         string     `json:"redirect_uri" gorm:"type:text;not null"`
	Scope               string     `json:"scope" gorm:"type:varchar(255);not null"`
	SessionId           string     `json:"-" gorm:"type:varchar(64);index"`
	UserAuthVersion     int64      `json:"-"`
	SessionVersion      int64      `json:"-"`
	AuthTime            int64      `json:"-"`
	Nonce               string     `json:"nonce" gorm:"type:varchar(255)"`
	CodeChallenge       string     `json:"code_challenge" gorm:"type:varchar(255)"`
	CodeChallengeMethod string     `json:"code_challenge_method" gorm:"type:varchar(16)"`
	ExpiresAt           time.Time  `json:"expires_at" gorm:"index;not null"`
	UsedAt              *time.Time `json:"used_at"`
	CreatedAt           time.Time  `json:"created_at"`
}

type OIDCAccessToken struct {
	AuthorizationCodeId int        `json:"-" gorm:"index"`
	Id                  int        `json:"id" gorm:"primaryKey"`
	TokenHash           string     `json:"-" gorm:"type:char(64);uniqueIndex;not null"`
	ClientId            string     `json:"client_id" gorm:"type:varchar(128);index;not null"`
	UserId              int        `json:"user_id" gorm:"index;not null"`
	Scope               string     `json:"scope" gorm:"type:varchar(255);not null"`
	SessionId           string     `json:"-" gorm:"type:varchar(64);index"`
	UserAuthVersion     int64      `json:"-"`
	SessionVersion      int64      `json:"-"`
	ExpiresAt           time.Time  `json:"expires_at" gorm:"index;not null"`
	RevokedAt           *time.Time `json:"revoked_at"`
	CreatedAt           time.Time  `json:"created_at"`
}

type OIDCSigningKey struct {
	Id            int       `json:"id" gorm:"primaryKey"`
	Kid           string    `json:"kid" gorm:"type:varchar(128);uniqueIndex;not null"`
	PrivateKeyPEM string    `json:"-" gorm:"type:text;not null"`
	PublicKeyPEM  string    `json:"-" gorm:"type:text;not null"`
	Algorithm     string    `json:"algorithm" gorm:"type:varchar(16);not null"`
	Active        bool      `json:"active" gorm:"index"`
	CreatedAt     time.Time `json:"created_at"`
}

func (OIDCClient) TableName() string {
	return "oidc_clients"
}

func (OIDCAuthorizationCode) TableName() string {
	return "oidc_authorization_codes"
}

func (OIDCAccessToken) TableName() string {
	return "oidc_access_tokens"
}

func (OIDCSigningKey) TableName() string {
	return "oidc_signing_keys"
}

func ListOIDCClients() ([]*OIDCClient, error) {
	var clients []*OIDCClient
	err := DB.Order("id asc").Find(&clients).Error
	return clients, err
}

func GetOIDCClientById(id int) (*OIDCClient, error) {
	var client OIDCClient
	if err := DB.First(&client, id).Error; err != nil {
		return nil, err
	}
	return &client, nil
}

func GetOIDCClientByClientId(clientId string) (*OIDCClient, error) {
	var client OIDCClient
	if err := DB.Where("client_id = ?", clientId).First(&client).Error; err != nil {
		return nil, err
	}
	return &client, nil
}

func CreateOIDCClient(client *OIDCClient, authorization ...*AuthFlowAuthorization) error {
	if client.Name == "" || client.ClientId == "" || client.ClientSecretHash == "" {
		return common.NewMessage("Missing required OIDC client fields")
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := validateOIDCClientAdministration(tx, authorization); err != nil {
			return err
		}
		return tx.Create(client).Error
	})
}

func UpdateOIDCClient(client *OIDCClient, authorization ...*AuthFlowAuthorization) error {
	if client.Name == "" || client.ClientId == "" {
		return common.NewMessage("Missing required OIDC client fields")
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := validateOIDCClientAdministration(tx, authorization); err != nil {
			return err
		}
		result := tx.Model(&OIDCClient{}).Where("id = ?", client.Id).Updates(map[string]any{
			"name": client.Name, "redirect_uris": client.RedirectURIs, "scopes": client.Scopes, "trusted": client.Trusted, "enabled": client.Enabled,
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func DeleteOIDCClient(id int, authorization ...*AuthFlowAuthorization) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := validateOIDCClientAdministration(tx, authorization); err != nil {
			return err
		}
		return tx.Delete(&OIDCClient{}, id).Error
	})
}

func RotateOIDCClientSecret(id int, hash string, authorization ...*AuthFlowAuthorization) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := validateOIDCClientAdministration(tx, authorization); err != nil {
			return err
		}
		result := tx.Model(&OIDCClient{}).Where("id = ?", id).Update("client_secret_hash", hash)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func validateOIDCClientAdministration(tx *gorm.DB, authorizations []*AuthFlowAuthorization) error {
	if len(authorizations) == 0 {
		return nil
	}
	authorization := authorizations[0]
	if authorization == nil || authorization.ProofID <= 0 {
		return ErrAuthFlowInvalid
	}
	if err := ValidateAuthSessionWithTx(tx, authorization.AuthSessionIdentity); err != nil {
		return err
	}
	var user User
	if err := tx.First(&user, authorization.UserID).Error; err != nil {
		return err
	}
	if user.Role != common.RoleRootUser {
		return ErrAuthFlowInvalid
	}
	return nil
}

func CreateOIDCAuthorizationCode(code *OIDCAuthorizationCode) error {
	return DB.Transaction(func(tx *gorm.DB) error { return CreateOIDCAuthorizationCodeWithTx(tx, code) })
}

// CreateOIDCAuthorizationCodeWithTx rechecks both the dashboard session and
// the registered client before an authorization code can be committed.
func CreateOIDCAuthorizationCodeWithTx(tx *gorm.DB, code *OIDCAuthorizationCode) error {
	identity := AuthSessionIdentity{UserID: code.UserId, SessionID: code.SessionId, UserAuthVersion: code.UserAuthVersion, SessionVersion: code.SessionVersion}
	if _, pat := ParseAccessTokenSessionID(code.SessionId); pat {
		return ErrAuthFlowInvalid
	}
	if err := ValidateAuthSessionWithTx(tx, identity); err != nil {
		return err
	}
	var client OIDCClient
	if err := lockForUpdate(tx).Where("client_id = ? AND enabled = ?", code.ClientId, true).First(&client).Error; err != nil {
		return err
	}
	uris := strings.FieldsFunc(client.RedirectURIs, func(r rune) bool { return r == '\n' || r == '\r' || r == ',' })
	allowed := false
	for _, uri := range uris {
		if strings.TrimSpace(uri) == code.RedirectURI {
			allowed = true
			break
		}
	}
	if !allowed {
		return ErrAuthFlowInvalid
	}
	for _, scope := range strings.Fields(code.Scope) {
		if !slices.Contains(strings.Fields(client.Scopes), scope) {
			return ErrAuthFlowInvalid
		}
	}
	return tx.Create(code).Error
}

func GetOIDCAuthorizationCodeByHash(codeHash string) (*OIDCAuthorizationCode, error) {
	var code OIDCAuthorizationCode
	if err := DB.Where("code_hash = ?", codeHash).First(&code).Error; err != nil {
		return nil, err
	}
	return &code, nil
}

func MarkOIDCAuthorizationCodeUsed(tx *gorm.DB, id int, usedAt time.Time) error {
	result := tx.Model(&OIDCAuthorizationCode{}).Where("id = ? AND used_at IS NULL AND expires_at > ?", id, usedAt).Update("used_at", usedAt)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return ErrAuthFlowConsumed
	}
	return nil
}

func CreateOIDCAccessToken(tx *gorm.DB, token *OIDCAccessToken) error {
	return tx.Create(token).Error
}

func GetOIDCAccessTokenByHash(tokenHash string) (*OIDCAccessToken, error) {
	var token OIDCAccessToken
	if err := DB.Where("token_hash = ?", tokenHash).First(&token).Error; err != nil {
		return nil, err
	}
	return &token, nil
}

func GetActiveOIDCSigningKey() (*OIDCSigningKey, error) {
	var key OIDCSigningKey
	if err := DB.Where("active = ?", true).Order("id desc").First(&key).Error; err != nil {
		return nil, err
	}
	return &key, nil
}

func CreateOIDCSigningKey(key *OIDCSigningKey) error {
	return DB.Create(key).Error
}

// ListOIDCGrantCodes returns pending authorization requests and authorizations
// with a live reference token. Browser session validity is checked by service.
func ListOIDCGrantCodes(userID int, now time.Time) ([]OIDCAuthorizationCode, error) {
	activeTokens := DB.Model(&OIDCAccessToken{}).Select("authorization_code_id").Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", userID, now)
	var codes []OIDCAuthorizationCode
	err := DB.Where("user_id = ?", userID).Where("(used_at IS NULL AND expires_at > ?) OR id IN (?)", now, activeTokens).Order("id desc").Find(&codes).Error
	return codes, err
}

// ChangeOIDCGrant reduces an owner's scopes or revokes the entire grant. The
// current browser session may differ from the session that authorized it.
func ChangeOIDCGrant(identity AuthSessionIdentity, codeID int, scope *string) error {
	if _, pat := ParseAccessTokenSessionID(identity.SessionID); pat || identity.SessionID == "" {
		return ErrAuthFlowInvalid
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := ValidateAuthSessionWithTx(tx, identity); err != nil {
			return err
		}
		var code OIDCAuthorizationCode
		if err := lockForUpdate(tx).Where("id = ? AND user_id = ?", codeID, identity.UserID).First(&code).Error; err != nil {
			return err
		}
		now := time.Now()
		if scope == nil {
			if err := tx.Model(&code).Updates(map[string]any{"used_at": now, "expires_at": now}).Error; err != nil {
				return err
			}
			return tx.Model(&OIDCAccessToken{}).Where("authorization_code_id = ? AND user_id = ? AND revoked_at IS NULL", code.Id, identity.UserID).Update("revoked_at", now).Error
		}
		scopes := strings.Fields(*scope)
		if !slices.Contains(scopes, "openid") {
			return common.NewMessage("OIDC grant scopes can only be reduced and must include openid")
		}
		allowed := strings.Fields(code.Scope)
		for _, value := range scopes {
			if !slices.Contains([]string{"openid", "profile", "email"}, value) || !slices.Contains(allowed, value) {
				return common.NewMessage("OIDC grant scopes can only be reduced and must include openid")
			}
		}
		if code.Scope != *scope {
			if err := tx.Model(&code).Update("scope", *scope).Error; err != nil {
				return err
			}
		}
		return tx.Model(&OIDCAccessToken{}).Where("authorization_code_id = ? AND user_id = ? AND revoked_at IS NULL", code.Id, identity.UserID).Update("scope", *scope).Error
	})
}
