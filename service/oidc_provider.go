package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"math/big"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

const (
	OIDCDefaultScopes           = "openid profile email"
	OIDCAuthorizationCodeTTL    = 5 * time.Minute
	OIDCAccessTokenTTL          = 15 * time.Minute
	OIDCSigningAlgorithm        = "RS256"
	OIDCCodeChallengeMethodS256 = "S256"
	OIDCCodeChallengePlain      = "plain" // rejected; retained as a named protocol value
)

type OIDCProtocolError struct {
	Code        string
	Description string
	StatusCode  int
}

func (e *OIDCProtocolError) Error() string {
	if e.Description == "" {
		return e.Code
	}
	return e.Code + ": " + e.Description
}

type CreateOIDCClientInput struct {
	Name         string
	RedirectURIs []string
	Scopes       []string
	Trusted      bool
	Enabled      bool
}

type UpdateOIDCClientInput struct {
	Name         string
	RedirectURIs []string
	Scopes       []string
	Trusted      bool
	Enabled      bool
}

type AuthorizeRequest struct {
	ClientId            string
	RedirectURI         string
	ResponseType        string
	Scope               string
	State               string
	Nonce               string
	CodeChallenge       string
	CodeChallengeMethod string
}

type AuthorizationCodeInput struct {
	Identity            AuthIdentity
	AuthTime            int64
	ClientId            string
	UserId              int
	RedirectURI         string
	Scope               string
	Nonce               string
	CodeChallenge       string
	CodeChallengeMethod string
}

type TokenExchangeInput struct {
	ClientId     string
	ClientSecret string
	Code         string
	RedirectURI  string
	CodeVerifier string
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	IDToken     string `json:"id_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

func oidcProtocolError(code, description string, status int) *OIDCProtocolError {
	return &OIDCProtocolError{
		Code:        code,
		Description: description,
		StatusCode:  status,
	}
}

func OIDCErrorResponse(err error) (int, map[string]any) {
	var protocolErr *OIDCProtocolError
	if errors.As(err, &protocolErr) {
		status := protocolErr.StatusCode
		if status == 0 {
			status = http.StatusBadRequest
		}
		return status, map[string]any{
			"error":             protocolErr.Code,
			"error_description": protocolErr.Description,
		}
	}
	return http.StatusInternalServerError, map[string]any{
		"error":             "server_error",
		"error_description": "OIDC server error",
	}
}

func OIDCConsentRequiredError() error {
	return oidcProtocolError("consent_required", "client requires user consent", http.StatusForbidden)
}

func HashOIDCToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func NormalizeOIDCScope(scope string) string {
	seen := make(map[string]bool)
	var scopes []string
	for _, item := range strings.Fields(scope) {
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		scopes = append(scopes, item)
	}
	slices.Sort(scopes)
	return strings.Join(scopes, " ")
}

func OIDCSubject(userId int) string {
	return "user_" + strconv.Itoa(userId)
}

func OIDCUserRole(role int) string {
	if role >= common.RoleRootUser {
		return "root"
	}
	if role >= common.RoleAdminUser {
		return "admin"
	}
	return "user"
}

func VerifyPKCE(challenge, method, verifier string) bool {
	if method != OIDCCodeChallengeMethodS256 || len(challenge) != 43 || len(verifier) < 43 || len(verifier) > 128 {
		return false
	}
	for _, c := range verifier {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '.' || c == '_' || c == '~') {
			return false
		}
	}
	sum := sha256.Sum256([]byte(verifier))
	return subtle.ConstantTimeCompare([]byte(base64.RawURLEncoding.EncodeToString(sum[:])), []byte(challenge)) == 1
}

func BuildOIDCDiscovery(issuer string) map[string]any {
	issuer = strings.TrimRight(issuer, "/")
	return map[string]any{
		"issuer":                                issuer,
		"authorization_endpoint":                issuer + "/oauth/authorize",
		"token_endpoint":                        issuer + "/oauth/token",
		"userinfo_endpoint":                     issuer + "/oauth/userinfo",
		"jwks_uri":                              issuer + "/oauth/jwks",
		"response_types_supported":              []string{"code"},
		"grant_types_supported":                 []string{"authorization_code"},
		"subject_types_supported":               []string{"public"},
		"id_token_signing_alg_values_supported": []string{OIDCSigningAlgorithm},
		"scopes_supported":                      []string{"openid", "profile", "email"},
		"claims_supported":                      []string{"sub", "preferred_username", "name", "email", "role", "role_level", "is_admin"},
		"code_challenge_methods_supported":      []string{OIDCCodeChallengeMethodS256},
		"token_endpoint_auth_methods_supported": []string{"client_secret_basic", "client_secret_post"},
	}
}

func CreateOIDCClient(input CreateOIDCClientInput, authorization ...*model.AuthFlowAuthorization) (*model.OIDCClient, string, error) {
	name := strings.TrimSpace(input.Name)
	redirectURIs := normalizeOIDCList(input.RedirectURIs)
	scopes := normalizeOIDCList(input.Scopes)
	if len(scopes) == 0 {
		scopes = strings.Fields(OIDCDefaultScopes)
	}
	if name == "" || len(name) > 128 {
		return nil, "", common.NewMessage("OIDC client name must contain between 1 and 128 characters")
	}
	if len(redirectURIs) == 0 {
		return nil, "", common.NewMessage("At least one OIDC redirect URI is required")
	}
	if err := validateOIDCClientScopes(scopes); err != nil {
		return nil, "", err
	}
	if err := validateOIDCRedirectURIs(redirectURIs); err != nil {
		return nil, "", err
	}
	clientId, err := generateOIDCRandom("oidc_client_", 24)
	if err != nil {
		return nil, "", err
	}
	clientSecret, err := generateOIDCRandom("oidc_secret_", 40)
	if err != nil {
		return nil, "", err
	}
	secretHash, err := common.Password2Hash(clientSecret)
	if err != nil {
		return nil, "", err
	}
	client := &model.OIDCClient{
		Name:             name,
		ClientId:         clientId,
		ClientSecretHash: secretHash,
		RedirectURIs:     strings.Join(redirectURIs, "\n"),
		Scopes:           NormalizeOIDCScope(strings.Join(scopes, " ")),
		Trusted:          input.Trusted,
		Enabled:          input.Enabled,
	}
	if err := model.CreateOIDCClient(client, authorization...); err != nil {
		return nil, "", err
	}
	return client, clientSecret, nil
}

func UpdateOIDCClient(client *model.OIDCClient, input UpdateOIDCClientInput, authorization ...*model.AuthFlowAuthorization) error {
	name := strings.TrimSpace(input.Name)
	redirectURIs := normalizeOIDCList(input.RedirectURIs)
	scopes := normalizeOIDCList(input.Scopes)
	if len(scopes) == 0 {
		scopes = strings.Fields(OIDCDefaultScopes)
	}
	if name == "" || len(name) > 128 {
		return common.NewMessage("OIDC client name must contain between 1 and 128 characters")
	}
	if len(redirectURIs) == 0 {
		return common.NewMessage("At least one OIDC redirect URI is required")
	}
	if err := validateOIDCClientScopes(scopes); err != nil {
		return err
	}
	if err := validateOIDCRedirectURIs(redirectURIs); err != nil {
		return err
	}
	client.Name = name
	client.RedirectURIs = strings.Join(redirectURIs, "\n")
	client.Scopes = NormalizeOIDCScope(strings.Join(scopes, " "))
	client.Trusted = input.Trusted
	client.Enabled = input.Enabled
	return model.UpdateOIDCClient(client, authorization...)
}

func RotateOIDCClientSecret(client *model.OIDCClient, authorization ...*model.AuthFlowAuthorization) (string, error) {
	clientSecret, err := generateOIDCRandom("oidc_secret_", 40)
	if err != nil {
		return "", err
	}
	secretHash, err := common.Password2Hash(clientSecret)
	if err != nil {
		return "", err
	}
	client.ClientSecretHash = secretHash
	if err := model.RotateOIDCClientSecret(client.Id, secretHash, authorization...); err != nil {
		return "", err
	}
	return clientSecret, nil
}

func ValidateAuthorizeRequest(req AuthorizeRequest) (*model.OIDCClient, string, error) {
	if len(req.ClientId) > 128 || len(req.RedirectURI) > 2048 || len(req.Scope) > 255 || len(req.State) > 1024 || len(req.Nonce) > 255 {
		return nil, "", oidcProtocolError("invalid_request", "authorization parameters are too long", http.StatusBadRequest)
	}
	if req.ResponseType != "code" {
		return nil, "", oidcProtocolError("unsupported_response_type", "response_type must be code", http.StatusBadRequest)
	}
	if strings.TrimSpace(req.ClientId) == "" {
		return nil, "", oidcProtocolError("invalid_request", "client_id is required", http.StatusBadRequest)
	}
	client, err := model.GetOIDCClientByClientId(req.ClientId)
	if err != nil {
		return nil, "", oidcProtocolError("invalid_client", "client is unknown", http.StatusBadRequest)
	}
	if !client.Enabled {
		return nil, "", oidcProtocolError("invalid_client", "client is disabled", http.StatusBadRequest)
	}
	if !oidcListContains(splitOIDCList(client.RedirectURIs), req.RedirectURI) {
		return nil, "", oidcProtocolError("invalid_request", "redirect_uri is not allowed", http.StatusBadRequest)
	}
	requestedScope := NormalizeOIDCScope(req.Scope)
	if requestedScope == "" {
		requestedScope = "openid"
	}
	if !oidcListContains(strings.Fields(requestedScope), "openid") {
		return nil, "", oidcProtocolError("invalid_scope", "scope must include openid", http.StatusBadRequest)
	}
	for _, scope := range strings.Fields(requestedScope) {
		if !oidcListContains(strings.Fields(client.Scopes), scope) {
			return nil, "", oidcProtocolError("invalid_scope", "scope is not allowed: "+scope, http.StatusBadRequest)
		}
	}
	if req.CodeChallengeMethod != OIDCCodeChallengeMethodS256 || len(req.CodeChallenge) != 43 {
		return nil, "", oidcProtocolError("invalid_request", "S256 PKCE is required", http.StatusBadRequest)
	}
	decoded, err := base64.RawURLEncoding.DecodeString(req.CodeChallenge)
	if err != nil || len(decoded) != 32 {
		return nil, "", oidcProtocolError("invalid_request", "invalid S256 code challenge", http.StatusBadRequest)
	}

	return client, requestedScope, nil
}

func IssueAuthorizationCode(input AuthorizationCodeInput) (string, error) {
	var code string
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		var err error
		code, err = IssueAuthorizationCodeWithTx(tx, input)
		return err
	})
	return code, err
}

func IssueAuthorizationCodeWithTx(tx *gorm.DB, input AuthorizationCodeInput) (string, error) {
	if input.UserId <= 0 || input.UserId != input.Identity.UserID || input.CodeChallengeMethod != OIDCCodeChallengeMethodS256 || len(input.CodeChallenge) != 43 {
		return "", model.ErrAuthFlowInvalid
	}
	rawCode, err := generateOIDCRandom("oidc_code_", 40)
	if err != nil {
		return "", err
	}
	code := &model.OIDCAuthorizationCode{
		CodeHash: HashOIDCToken(rawCode), ClientId: input.ClientId, UserId: input.UserId, RedirectURI: input.RedirectURI,
		Scope: NormalizeOIDCScope(input.Scope), Nonce: input.Nonce, CodeChallenge: input.CodeChallenge, CodeChallengeMethod: input.CodeChallengeMethod,
		SessionId: input.Identity.SessionID, UserAuthVersion: input.Identity.UserAuthVersion, SessionVersion: input.Identity.SessionVersion,
		AuthTime: input.AuthTime, ExpiresAt: time.Now().Add(OIDCAuthorizationCodeTTL),
	}
	if err := model.CreateOIDCAuthorizationCodeWithTx(tx, code); err != nil {
		return "", err
	}
	return rawCode, nil
}

func ExchangeAuthorizationCode(input TokenExchangeInput) (*TokenResponse, error) {
	client, err := model.GetOIDCClientByClientId(input.ClientId)
	if err != nil {
		return nil, oidcProtocolError("invalid_client", "client authentication failed", http.StatusUnauthorized)
	}
	if !client.Enabled || len(input.ClientSecret) > 72 || !common.ValidatePasswordAndHash(input.ClientSecret, client.ClientSecretHash) {
		return nil, oidcProtocolError("invalid_client", "client authentication failed", http.StatusUnauthorized)
	}
	codeHash := HashOIDCToken(input.Code)
	code, err := model.GetOIDCAuthorizationCodeByHash(codeHash)
	if err != nil {
		return nil, oidcProtocolError("invalid_grant", "authorization code is invalid", http.StatusBadRequest)
	}
	now := time.Now()
	if code.ClientId != client.ClientId || code.RedirectURI != input.RedirectURI {
		return nil, oidcProtocolError("invalid_grant", "authorization code does not match this client", http.StatusBadRequest)
	}
	if !VerifyPKCE(code.CodeChallenge, code.CodeChallengeMethod, input.CodeVerifier) {
		return nil, oidcProtocolError("invalid_grant", "code_verifier is invalid", http.StatusBadRequest)
	}
	if code.UsedAt != nil {
		if err := revokeOIDCCodeTokens(code.Id); err != nil {
			return nil, err
		}
		return nil, oidcProtocolError("invalid_grant", "authorization code was already used", http.StatusBadRequest)
	}
	if !code.ExpiresAt.After(now) {
		return nil, oidcProtocolError("invalid_grant", "authorization code expired", http.StatusBadRequest)
	}

	user, err := model.GetUserById(code.UserId, true)
	if err != nil || user.Status != common.UserStatusEnabled {
		return nil, oidcProtocolError("invalid_grant", "user is unavailable", http.StatusBadRequest)
	}
	accessToken, err := generateOIDCRandom("oidc_at_", 48)
	if err != nil {
		return nil, err
	}
	idToken, err := signOIDCIDToken(client.ClientId, user, code.Scope, code.Nonce, code.AuthTime, now)
	if err != nil {
		return nil, err
	}

	err = model.DB.Transaction(func(tx *gorm.DB) error {
		if err := model.ValidateAuthSessionWithTx(tx, AuthIdentity{UserID: code.UserId, SessionID: code.SessionId, UserAuthVersion: code.UserAuthVersion, SessionVersion: code.SessionVersion}); err != nil {
			return oidcProtocolError("invalid_grant", "authorization session is unavailable", http.StatusBadRequest)
		}
		var currentClient model.OIDCClient
		if err := tx.Where("id = ? AND enabled = ? AND client_secret_hash = ?", client.Id, true, client.ClientSecretHash).First(&currentClient).Error; err != nil {
			return oidcProtocolError("invalid_client", "client authentication failed", http.StatusUnauthorized)
		}
		result := tx.Model(&model.OIDCAuthorizationCode{}).
			Where("id = ? AND scope = ? AND used_at IS NULL AND expires_at > ?", code.Id, code.Scope, time.Now()).
			Update("used_at", now)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return model.ErrAuthFlowConsumed
		}
		return model.CreateOIDCAccessToken(tx, &model.OIDCAccessToken{
			AuthorizationCodeId: code.Id,
			TokenHash:           HashOIDCToken(accessToken),
			ClientId:            client.ClientId,
			UserId:              user.Id,
			Scope:               code.Scope,
			SessionId:           code.SessionId, UserAuthVersion: code.UserAuthVersion, SessionVersion: code.SessionVersion,
			ExpiresAt: now.Add(OIDCAccessTokenTTL),
		})
	})
	if errors.Is(err, model.ErrAuthFlowConsumed) {
		if revokeErr := revokeOIDCCodeTokens(code.Id); revokeErr != nil {
			return nil, revokeErr
		}
		return nil, oidcProtocolError("invalid_grant", "authorization code was already used", http.StatusBadRequest)
	}
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken: accessToken,
		IDToken:     idToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(OIDCAccessTokenTTL.Seconds()),
		Scope:       code.Scope,
	}, nil
}

// Revoke every reference token derived from a replayed authorization code.
// Client authentication, redirect binding, and PKCE are checked before this
// operation so an unrelated request cannot revoke another client's tokens.
func revokeOIDCCodeTokens(codeID int) error {
	return model.DB.Model(&model.OIDCAccessToken{}).Where("authorization_code_id = ? AND revoked_at IS NULL", codeID).Update("revoked_at", time.Now()).Error
}

func BuildUserInfo(accessToken string) (map[string]any, error) {
	parts := strings.Fields(accessToken)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return nil, oidcProtocolError("invalid_token", "Bearer access token is required", http.StatusUnauthorized)
	}
	accessToken = parts[1]
	if accessToken == "" {
		return nil, oidcProtocolError("invalid_token", "access token is required", http.StatusUnauthorized)
	}
	token, err := model.GetOIDCAccessTokenByHash(HashOIDCToken(accessToken))
	if err != nil {
		return nil, oidcProtocolError("invalid_token", "access token is invalid", http.StatusUnauthorized)
	}
	if token.RevokedAt != nil || !token.ExpiresAt.After(time.Now()) {
		return nil, oidcProtocolError("invalid_token", "access token is expired or revoked", http.StatusUnauthorized)
	}
	client, err := model.GetOIDCClientByClientId(token.ClientId)
	if err != nil || !client.Enabled {
		return nil, oidcProtocolError("invalid_token", "client is unavailable", http.StatusUnauthorized)
	}
	if _, _, err := ValidateLoginSession(AuthIdentity{UserID: token.UserId, SessionID: token.SessionId, UserAuthVersion: token.UserAuthVersion, SessionVersion: token.SessionVersion}); err != nil {
		return nil, oidcProtocolError("invalid_token", "authorization session is unavailable", http.StatusUnauthorized)
	}
	user, err := model.GetUserById(token.UserId, true)
	if err != nil || user.Status != common.UserStatusEnabled {
		return nil, oidcProtocolError("invalid_token", "user is unavailable", http.StatusUnauthorized)
	}
	info := map[string]any{"sub": OIDCSubject(user.Id)}
	if slices.Contains(strings.Fields(token.Scope), "profile") {
		info["preferred_username"] = user.Username
		info["name"] = user.DisplayName
		info["role"] = OIDCUserRole(user.Role)
		info["role_level"] = user.Role
		info["is_admin"] = user.Role >= common.RoleAdminUser
	}

	if oidcListContains(strings.Fields(token.Scope), "email") && user.Email != "" {
		info["email"] = user.Email
	}
	return info, nil
}

func BuildJWKS() (map[string]any, error) {
	if _, err := ensureOIDCSigningKey(); err != nil {
		return nil, err
	}
	var signingKeys []model.OIDCSigningKey
	if err := model.DB.Where("active = ?", true).Order("id asc").Find(&signingKeys).Error; err != nil {
		return nil, err
	}
	keys := make([]map[string]any, 0, len(signingKeys))
	for _, signingKey := range signingKeys {
		publicKey, err := parseOIDCPublicKey(signingKey.PublicKeyPEM)
		if err != nil {
			return nil, err
		}
		keys = append(keys, map[string]any{"kty": "RSA", "use": "sig", "kid": signingKey.Kid, "alg": signingKey.Algorithm,
			"n": base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes()), "e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(publicKey.E)).Bytes())})
	}
	return map[string]any{"keys": keys}, nil
}

func OIDCIssuer() (string, error) {
	issuer := strings.TrimRight(system_setting.ServerAddress, "/")
	parsed, err := url.Parse(issuer)
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || strings.ContainsAny(issuer, "\r\n") {
		return "", errors.New("OIDC issuer is not configured correctly")
	}
	if parsed.Scheme != "https" && !(parsed.Scheme == "http" && (parsed.Hostname() == "localhost" || parsed.Hostname() == "127.0.0.1")) {
		return "", errors.New("OIDC issuer must use HTTPS")
	}
	return issuer, nil
}

func normalizeOIDCList(values []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, value := range values {
		for _, part := range strings.FieldsFunc(value, func(r rune) bool {
			return r == '\n' || r == '\r' || r == ','
		}) {
			part = strings.TrimSpace(part)
			if part == "" || seen[part] {
				continue
			}
			seen[part] = true
			result = append(result, part)
		}
	}
	return result
}

func splitOIDCList(value string) []string {
	return normalizeOIDCList([]string{value})
}

func oidcListContains(values []string, target string) bool { return slices.Contains(values, target) }

func validateOIDCRedirectURIs(redirectURIs []string) error {
	for _, redirectURI := range redirectURIs {
		parsed, err := url.Parse(redirectURI)
		if err != nil || parsed.Host == "" || parsed.User != nil || parsed.Fragment != "" || len(redirectURI) > 2048 || strings.ContainsAny(redirectURI, "\r\n") {
			return common.NewMessage("Invalid OIDC redirect URI: {{uri}}", map[string]any{"uri": redirectURI})
		}
		if parsed.Scheme == "https" {
			continue
		}
		if parsed.Scheme == "http" && (parsed.Hostname() == "localhost" || parsed.Hostname() == "127.0.0.1") {
			continue
		}
		return common.NewMessage("OIDC redirect URI must use HTTPS outside localhost: {{uri}}", map[string]any{"uri": redirectURI})
	}
	return nil
}

func generateOIDCRandom(prefix string, length int) (string, error) {
	raw, err := common.GenerateRandomKey(length)
	if err != nil {
		return "", err
	}
	return prefix + raw, nil
}

func signOIDCIDToken(clientId string, user *model.User, scope string, nonce string, authTime int64, now time.Time) (string, error) {
	signingKey, err := ensureOIDCSigningKey()
	if err != nil {
		return "", err
	}
	privateKey, err := parseOIDCPrivateKey(signingKey.PrivateKeyPEM)
	if err != nil {
		return "", err
	}
	issuer, err := OIDCIssuer()
	if err != nil {
		return "", err
	}
	claims := jwt.MapClaims{
		"iss": issuer, "sub": OIDCSubject(user.Id), "aud": clientId, "exp": now.Add(OIDCAccessTokenTTL).Unix(), "iat": now.Unix(), "auth_time": authTime,
	}
	if slices.Contains(strings.Fields(scope), "profile") {
		claims["preferred_username"] = user.Username
		claims["name"] = user.DisplayName
		claims["role"] = OIDCUserRole(user.Role)
		claims["role_level"] = user.Role
		claims["is_admin"] = user.Role >= common.RoleAdminUser
	}

	if nonce != "" {
		claims["nonce"] = nonce
	}
	if oidcListContains(strings.Fields(scope), "email") && user.Email != "" {
		claims["email"] = user.Email
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = signingKey.Kid
	return token.SignedString(privateKey)
}

var oidcSigningKeyLock sync.Mutex

func ensureOIDCSigningKey() (*model.OIDCSigningKey, error) {
	oidcSigningKeyLock.Lock()
	defer oidcSigningKeyLock.Unlock()
	signingKey, err := model.GetActiveOIDCSigningKey()
	if err == nil {
		return signingKey, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	privatePEM, publicPEM, err := encodeOIDCKeyPair(privateKey)
	if err != nil {
		return nil, err
	}
	kid, err := generateOIDCRandom("oidc_kid_", 16)
	if err != nil {
		return nil, err
	}
	signingKey = &model.OIDCSigningKey{
		Kid:           kid,
		PrivateKeyPEM: privatePEM,
		PublicKeyPEM:  publicPEM,
		Algorithm:     OIDCSigningAlgorithm,
		Active:        true,
	}
	if err := model.CreateOIDCSigningKey(signingKey); err != nil {
		return nil, err
	}
	return signingKey, nil
}

func encodeOIDCKeyPair(privateKey *rsa.PrivateKey) (string, string, error) {
	privateDER := x509.MarshalPKCS1PrivateKey(privateKey)
	privatePEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privateDER})
	publicDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", err
	}
	publicPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicDER})
	return string(privatePEM), string(publicPEM), nil
}

func parseOIDCPrivateKey(privateKeyPEM string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, errors.New("invalid private key PEM")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func parseOIDCPublicKey(publicKeyPEM string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, errors.New("invalid public key PEM")
	}
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("public key is not RSA")
	}
	return rsaKey, nil
}

func validateOIDCClientScopes(scopes []string) error {
	if len(scopes) > 3 || !slices.Contains(scopes, "openid") {
		return common.NewMessage("OIDC scopes must include openid and use only supported scopes")
	}
	for _, scope := range scopes {
		if !slices.Contains([]string{"openid", "profile", "email"}, scope) {
			return common.NewMessage("OIDC scopes must include openid and use only supported scopes")
		}
	}
	return nil
}

// OIDCGrant contains no usable code, token, credential, or session identifier.
type OIDCGrant struct {
	Id         int    `json:"id"`
	ClientName string `json:"client_name"`
	ClientId   string `json:"client_id"`
	Scope      string `json:"scope"`
	ExpiresAt  int64  `json:"expires_at"`
}

func ListOIDCGrants(identity AuthIdentity) ([]OIDCGrant, error) {
	if _, pat := model.ParseAccessTokenSessionID(identity.SessionID); pat || identity.SessionID == "" {
		return nil, ErrLoginSessionInvalid
	}
	if _, _, err := ValidateLoginSession(identity); err != nil {
		return nil, err
	}
	now := time.Now()
	codes, err := model.ListOIDCGrantCodes(identity.UserID, now)
	if err != nil {
		return nil, err
	}
	grants := make([]OIDCGrant, 0, len(codes))
	for _, code := range codes {
		client, err := model.GetOIDCClientByClientId(code.ClientId)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if !client.Enabled {
			continue
		}
		originSession, _, err := ValidateLoginSession(AuthIdentity{UserID: code.UserId, SessionID: code.SessionId, UserAuthVersion: code.UserAuthVersion, SessionVersion: code.SessionVersion})
		if err != nil {
			continue
		}
		expiresAt := code.ExpiresAt
		if code.UsedAt != nil {
			var token model.OIDCAccessToken
			err := model.DB.Where("authorization_code_id = ? AND user_id = ? AND revoked_at IS NULL AND expires_at > ?", code.Id, identity.UserID, now).Order("expires_at desc").First(&token).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			if err != nil {
				return nil, err
			}
			expiresAt = token.ExpiresAt
		}
		if expiresAt.Unix() > originSession.ExpiresAt {
			expiresAt = time.Unix(originSession.ExpiresAt, 0)
		}
		if !expiresAt.After(now) {
			continue
		}
		grants = append(grants, OIDCGrant{Id: code.Id, ClientName: client.Name, ClientId: client.ClientId, Scope: code.Scope, ExpiresAt: expiresAt.Unix()})
	}
	return grants, nil
}

func UpdateOIDCGrantScopes(identity AuthIdentity, codeID int, scope string) error {
	normalized := NormalizeOIDCScope(scope)
	return model.ChangeOIDCGrant(identity, codeID, &normalized)
}
