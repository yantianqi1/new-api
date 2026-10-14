package controller

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupOIDCProviderTest(t *testing.T) (*model.User, service.AuthIdentity, *model.OIDCClient, string, string, service.AuthorizeRequest) {
	t.Helper()
	user, identity := setupSecurityEnrollmentTest(t)
	require.NoError(t, model.DB.AutoMigrate(&model.OIDCClient{}, &model.OIDCAuthorizationCode{}, &model.OIDCAccessToken{}, &model.OIDCSigningKey{}))
	oldAddress := system_setting.ServerAddress
	system_setting.ServerAddress = "https://issuer.example"
	t.Cleanup(func() { system_setting.ServerAddress = oldAddress })
	client, secret, err := service.CreateOIDCClient(service.CreateOIDCClientInput{Name: "Test client", RedirectURIs: []string{"https://client.example/callback"}, Scopes: []string{"openid", "email", "profile"}, Enabled: true})
	require.NoError(t, err)
	verifier := strings.Repeat("a", 43)
	sum := sha256.Sum256([]byte(verifier))
	req := service.AuthorizeRequest{ClientId: client.ClientId, RedirectURI: "https://client.example/callback", ResponseType: "code", Scope: "openid email", State: "state", Nonce: "nonce", CodeChallenge: base64.RawURLEncoding.EncodeToString(sum[:]), CodeChallengeMethod: "S256"}
	return user, identity, client, secret, verifier, req
}

func oidcAuthorizeRequest(t *testing.T, method string, req service.AuthorizeRequest, body string, identity service.AuthIdentity) *httptest.ResponseRecorder {
	t.Helper()
	query := url.Values{"client_id": {req.ClientId}, "redirect_uri": {req.RedirectURI}, "response_type": {req.ResponseType}, "scope": {req.Scope}, "state": {req.State}, "nonce": {req.Nonce}, "code_challenge": {req.CodeChallenge}, "code_challenge_method": {req.CodeChallengeMethod}, "consent": {"allow"}}
	response := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(response)
	c.Request = httptest.NewRequest(method, "/oauth/authorize?"+query.Encode(), strings.NewReader(body))
	c.Request.Header.Set("Accept", "application/json")
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("id", identity.UserID)
	c.Set("session_id", identity.SessionID)
	c.Set("auth_version", identity.UserAuthVersion)
	c.Set("session_version", identity.SessionVersion)
	OIDCAuthorize(c)
	return response
}

func oidcResponseData(t *testing.T, response *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	var envelope struct {
		Success bool           `json:"success"`
		Data    map[string]any `json:"data"`
	}
	require.NoError(t, common.Unmarshal(response.Body.Bytes(), &envelope))
	require.True(t, envelope.Success, response.Body.String())
	return envelope.Data
}

func TestOIDCAuthorizationConsentAndRedirectGuards(t *testing.T) {
	_, identity, _, _, _, req := setupOIDCProviderTest(t)
	for _, test := range []struct {
		name   string
		mutate func(*service.AuthorizeRequest)
	}{
		{"unregistered redirect", func(r *service.AuthorizeRequest) { r.RedirectURI = "https://attacker.example/callback" }},
		{"missing PKCE", func(r *service.AuthorizeRequest) { r.CodeChallenge = "" }},
		{"plain PKCE", func(r *service.AuthorizeRequest) { r.CodeChallengeMethod = "plain" }},
	} {
		t.Run(test.name, func(t *testing.T) {
			invalid := req
			test.mutate(&invalid)
			response := oidcAuthorizeRequest(t, http.MethodGet, invalid, "", identity)
			assert.Equal(t, http.StatusBadRequest, response.Code)
			assert.Empty(t, response.Header().Get("Location"))
		})
	}
	response := oidcAuthorizeRequest(t, http.MethodGet, req, "", identity)
	data := oidcResponseData(t, response)
	require.NotEmpty(t, data["consent_token"])
	assert.Nil(t, data["authorization_url"])
	assert.Equal(t, "no-store", response.Header().Get("Cache-Control"))
	token := data["consent_token"].(string)
	body := fmt.Sprintf(`{"consent_token":%q,"decision":"allow","redirect_uri":"https://attacker.example"}`, token)
	result := oidcResponseData(t, oidcAuthorizeRequest(t, http.MethodPost, req, body, identity))
	callback, err := url.Parse(result["authorization_url"].(string))
	require.NoError(t, err)
	assert.Equal(t, "client.example", callback.Host)
	assert.NotEmpty(t, callback.Query().Get("code"))
	assert.Equal(t, "state", callback.Query().Get("state"))
	assert.Equal(t, http.StatusBadRequest, oidcAuthorizeRequest(t, http.MethodPost, req, body, identity).Code)
	data = oidcResponseData(t, oidcAuthorizeRequest(t, http.MethodGet, req, "", identity))
	token = data["consent_token"].(string)
	otherSession, err := service.CreateLoginSession(identity.UserID, "password", "127.0.0.2", "other-browser")
	require.NoError(t, err)
	wrongIdentity, err := service.ParseAccessToken(otherSession.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, oidcAuthorizeRequest(t, http.MethodPost, req, fmt.Sprintf(`{"consent_token":%q,"decision":"allow"}`, token), wrongIdentity).Code)
	require.NoError(t, model.DB.Model(&model.AuthFlow{}).Where("purpose = ? AND session_id = ? AND consumed_at IS NULL", oidcConsentPurpose, identity.SessionID).Update("expires_at", time.Now().Add(-time.Minute)).Error)
	assert.Equal(t, http.StatusBadRequest, oidcAuthorizeRequest(t, http.MethodPost, req, fmt.Sprintf(`{"consent_token":%q,"decision":"allow"}`, token), identity).Code)
}

func TestOIDCExchangeAndReferenceTokenRevocation(t *testing.T) {
	user, identity, client, secret, verifier, req := setupOIDCProviderTest(t)
	issue := func() string {
		t.Helper()
		code, err := service.IssueAuthorizationCode(service.AuthorizationCodeInput{Identity: identity, AuthTime: time.Now().Unix(), ClientId: client.ClientId, UserId: user.Id, RedirectURI: req.RedirectURI, Scope: req.Scope, Nonce: req.Nonce, CodeChallenge: req.CodeChallenge, CodeChallengeMethod: req.CodeChallengeMethod})
		require.NoError(t, err)
		return code
	}
	input := service.TokenExchangeInput{ClientId: client.ClientId, ClientSecret: secret, Code: issue(), RedirectURI: req.RedirectURI, CodeVerifier: verifier}
	for _, test := range []struct {
		name   string
		mutate func(*service.TokenExchangeInput)
	}{
		{"wrong secret", func(i *service.TokenExchangeInput) { i.ClientSecret = "bad" }},
		{"wrong redirect", func(i *service.TokenExchangeInput) { i.RedirectURI = "https://client.example/other" }},
		{"wrong verifier", func(i *service.TokenExchangeInput) { i.CodeVerifier = strings.Repeat("b", 43) }},
	} {
		t.Run(test.name, func(t *testing.T) {
			invalid := input
			test.mutate(&invalid)
			_, err := service.ExchangeAuthorizationCode(invalid)
			require.Error(t, err)
		})
	}
	token, err := service.ExchangeAuthorizationCode(input)
	require.NoError(t, err)
	info, err := service.BuildUserInfo("Bearer " + token.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, service.OIDCSubject(user.Id), info["sub"])
	assert.Nil(t, info["name"], "profile claims require profile scope")
	_, err = service.ExchangeAuthorizationCode(input)
	require.Error(t, err)
	_, err = service.BuildUserInfo("Bearer " + token.AccessToken)
	require.Error(t, err, "replaying a valid code revokes the originally issued token")
	input.Code = issue()
	expired, err := model.GetOIDCAuthorizationCodeByHash(service.HashOIDCToken(input.Code))
	require.NoError(t, err)
	require.NoError(t, model.DB.Model(expired).Update("expires_at", time.Now().Add(-time.Minute)).Error)
	_, err = service.ExchangeAuthorizationCode(input)
	require.Error(t, err)
	input.Code = issue()
	token, err = service.ExchangeAuthorizationCode(input)
	require.NoError(t, err)
	require.NoError(t, model.DB.Model(client).Update("enabled", false).Error)
	_, err = service.BuildUserInfo("Bearer " + token.AccessToken)
	require.Error(t, err)
	require.NoError(t, model.DB.Model(client).Update("enabled", true).Error)
	_, err = service.BuildUserInfo("Bearer " + token.AccessToken)
	require.NoError(t, err)
	_, err = model.RevokeUserSession(user.Id, identity.SessionID, "test")
	require.NoError(t, err)
	_, err = service.BuildUserInfo("Bearer " + token.AccessToken)
	require.Error(t, err)
}

func TestOIDCClientAdministrationProofAndSecretProtection(t *testing.T) {
	user, identity, client, _, _, _ := setupOIDCProviderTest(t)
	require.NoError(t, model.DB.Model(user).Update("role", common.RoleRootUser).Error)
	require.NoError(t, model.PublishUserAuthCache(user.Id))
	params := gin.Params{{Key: "id", Value: fmt.Sprint(client.Id)}}
	updateBody := `{"name":"Updated","redirect_uris":["https://client.example/callback"],"scopes":["openid"],"enabled":true}`
	response := adminUserRequest(http.MethodPut, "/api/oidc/client", updateBody, "", identity, common.RoleRootUser, params, UpdateOIDCClient)
	assert.Equal(t, http.StatusForbidden, response.Code)
	context, err := common.Marshal(service.OIDCClientContext{ClientID: client.Id})
	require.NoError(t, err)
	proof := issueSecurityEnrollmentProof(t, identity, service.VerificationOperation{Scope: service.VerificationScopeOIDCClientSecretRotate, Context: context}, "password")
	response = adminUserRequest(http.MethodPost, "/api/oidc/client/rotate", "", proof, identity, common.RoleRootUser, params, RotateOIDCClientSecret)
	data := oidcResponseData(t, response)
	secret := data["client_secret"].(string)
	assert.Equal(t, "no-store", response.Header().Get("Cache-Control"))
	assert.Equal(t, http.StatusForbidden, adminUserRequest(http.MethodPost, "/api/oidc/client/rotate", "", proof, identity, common.RoleRootUser, params, RotateOIDCClientSecret).Code)
	// A previously loaded client must not overwrite the rotated secret while
	// metadata is updated by a concurrent administrator.
	require.NoError(t, service.UpdateOIDCClient(client, service.UpdateOIDCClientInput{Name: "Updated", RedirectURIs: []string{"https://client.example/callback"}, Scopes: []string{"openid"}, Enabled: true}))
	stored, err := model.GetOIDCClientById(client.Id)
	require.NoError(t, err)
	assert.True(t, common.ValidatePasswordAndHash(secret, stored.ClientSecretHash))
	response = adminUserRequest(http.MethodGet, "/api/oidc/client", "", "", identity, common.RoleRootUser, nil, ListOIDCClients)
	assert.NotContains(t, response.Body.String(), secret)
	assert.NotContains(t, response.Body.String(), stored.ClientSecretHash)
	var audit model.AuditLog
	require.NoError(t, model.LOG_DB.Where("action = ?", "oidc.client.secret.rotate").First(&audit).Error)
	encoded, err := common.Marshal(audit)
	require.NoError(t, err)
	assert.NotContains(t, string(encoded), secret)
}

func TestOIDCTrustedAuthorizationAndSigningKeyDiscovery(t *testing.T) {
	_, identity, client, _, _, req := setupOIDCProviderTest(t)
	require.NoError(t, model.DB.Model(client).Update("trusted", true).Error)
	data := oidcResponseData(t, oidcAuthorizeRequest(t, http.MethodGet, req, "", identity))
	require.NotEmpty(t, data["authorization_url"])
	assert.Nil(t, data["consent_token"])
	jwks, err := service.BuildJWKS()
	require.NoError(t, err)
	require.Len(t, jwks["keys"], 1)
	key, err := model.GetActiveOIDCSigningKey()
	require.NoError(t, err)
	key.Id = 0
	key.Kid += "-another-node"
	require.NoError(t, model.CreateOIDCSigningKey(key))
	jwks, err = service.BuildJWKS()
	require.NoError(t, err)
	assert.Len(t, jwks["keys"], 2, "all simultaneously active keys must remain discoverable")
	system_setting.ServerAddress = ""
	_, err = service.OIDCIssuer()
	require.Error(t, err)
}

func TestOIDCTokenRejectsAmbiguousRequests(t *testing.T) {
	for _, test := range []struct {
		name, body string
		basic      bool
	}{
		{name: "duplicate grant", body: "grant_type=authorization_code&grant_type=authorization_code"},
		{name: "duplicate code", body: "grant_type=authorization_code&code=one&code=two"},
		{name: "mixed authentication", body: "grant_type=authorization_code&client_id=client", basic: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(response)
			c.Request = httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(test.body))
			c.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			if test.basic {
				c.Request.SetBasicAuth("client", "secret")
			}
			OIDCToken(c)
			assert.Equal(t, http.StatusBadRequest, response.Code)
			var body map[string]any
			require.NoError(t, common.Unmarshal(response.Body.Bytes(), &body))
			assert.Equal(t, "invalid_request", body["error"])
			assert.Equal(t, "no-store", response.Header().Get("Cache-Control"))
		})
	}
}

func TestOIDCTokenAcceptsEncodedAndExistingBasicCredentials(t *testing.T) {
	user, identity, client, _, verifier, req := setupOIDCProviderTest(t)
	// Represent a production credential hash containing the punctuation emitted
	// by the original generator. Keep that stored hash unchanged during exchange.
	secret := "existing+secret/with=punctuation"
	hash, err := common.Password2Hash(secret)
	require.NoError(t, err)
	require.NoError(t, model.RotateOIDCClientSecret(client.Id, hash))
	for _, encoded := range []bool{true, false} {
		t.Run(fmt.Sprintf("form_encoded=%t", encoded), func(t *testing.T) {
			code, err := service.IssueAuthorizationCode(service.AuthorizationCodeInput{Identity: identity, AuthTime: time.Now().Unix(), ClientId: client.ClientId, UserId: user.Id, RedirectURI: req.RedirectURI, Scope: req.Scope, CodeChallenge: req.CodeChallenge, CodeChallengeMethod: req.CodeChallengeMethod})
			require.NoError(t, err)
			form := url.Values{"grant_type": {"authorization_code"}, "code": {code}, "redirect_uri": {req.RedirectURI}, "code_verifier": {verifier}}
			response := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(response)
			c.Request = httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(form.Encode()))
			c.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			id, password := client.ClientId, secret
			if encoded {
				id, password = url.QueryEscape(id), url.QueryEscape(password)
			}
			c.Request.SetBasicAuth(id, password)
			OIDCToken(c)
			require.Equal(t, http.StatusOK, response.Code, response.Body.String())
			var token service.TokenResponse
			require.NoError(t, common.Unmarshal(response.Body.Bytes(), &token))
			info, err := service.BuildUserInfo("Bearer " + token.AccessToken)
			require.NoError(t, err)
			assert.Equal(t, service.OIDCSubject(user.Id), info["sub"])
			stored, err := model.GetOIDCClientById(client.Id)
			require.NoError(t, err)
			assert.Equal(t, hash, stored.ClientSecretHash)
		})
	}
}

func oidcGrantRequest(t *testing.T, method, path, body, accessToken string) *httptest.ResponseRecorder {
	t.Helper()
	engine := gin.New()
	group := engine.Group("/api/user/oidc/grants", middleware.UserAuth())
	group.GET("", ListOIDCGrants)
	group.PATCH("/:id", UpdateOIDCGrant)
	group.DELETE("/:id", RevokeOIDCGrant)
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+accessToken)
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	return response
}

func TestOIDCGrantsOwnerScopesExpiryAndRevocation(t *testing.T) {
	user, identity, client, secret, verifier, req := setupOIDCProviderTest(t)
	require.NoError(t, model.DB.Model(user).Update("email", "grant-owner@example.com").Error)
	session, err := service.CreateLoginSession(user.Id, "password", "127.0.0.2", "grant-owner-other-browser")
	require.NoError(t, err)
	req.Scope = "openid profile email"
	code, err := service.IssueAuthorizationCode(service.AuthorizationCodeInput{Identity: identity, AuthTime: time.Now().Unix(), ClientId: client.ClientId, UserId: user.Id, RedirectURI: req.RedirectURI, Scope: req.Scope, CodeChallenge: req.CodeChallenge, CodeChallengeMethod: req.CodeChallengeMethod})
	require.NoError(t, err)
	grant, err := model.GetOIDCAuthorizationCodeByHash(service.HashOIDCToken(code))
	require.NoError(t, err)
	input := service.TokenExchangeInput{ClientId: client.ClientId, ClientSecret: secret, Code: code, RedirectURI: req.RedirectURI, CodeVerifier: verifier}
	token, err := service.ExchangeAuthorizationCode(input)
	require.NoError(t, err)
	before, err := service.BuildUserInfo("Bearer " + token.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, "grant-owner@example.com", before["email"])
	list := oidcGrantRequest(t, http.MethodGet, "/api/user/oidc/grants", "", session.AccessToken)
	require.Equal(t, http.StatusOK, list.Code)
	var envelope struct {
		Success bool                `json:"success"`
		Data    []service.OIDCGrant `json:"data"`
	}
	require.NoError(t, common.Unmarshal(list.Body.Bytes(), &envelope))
	require.True(t, envelope.Success)
	require.Len(t, envelope.Data, 1)
	assert.Equal(t, grant.Id, envelope.Data[0].Id)
	assert.Equal(t, client.Name, envelope.Data[0].ClientName)
	assert.Greater(t, envelope.Data[0].ExpiresAt, time.Now().Unix())
	assert.NotContains(t, list.Body.String(), token.AccessToken)
	assert.NotContains(t, list.Body.String(), code)
	path := fmt.Sprintf("/api/user/oidc/grants/%d", grant.Id)
	patRaw := model.AccessTokenPrefix + "grant-pat-test"
	pat := &model.UserAccessToken{UserId: user.Id, Name: "grant-test", TokenHash: model.AccessTokenFingerprint(patRaw), CreatedAt: time.Now().Unix()}
	require.NoError(t, pat.SetScopes([]string{"profile:read", "account_security:write"}))
	require.NoError(t, model.CreateUserAccessToken(user.Id, pat, 20))
	for _, method := range []string{http.MethodGet, http.MethodPatch, http.MethodDelete} {
		target := path
		if method == http.MethodGet {
			target = "/api/user/oidc/grants"
		}
		response := oidcGrantRequest(t, method, target, `{"scope":"openid"}`, patRaw)
		assert.Equal(t, http.StatusForbidden, response.Code, "OIDC grants require browser authentication")
	}

	reduced := oidcGrantRequest(t, http.MethodPatch, path, `{"scope":"openid profile"}`, session.AccessToken)
	oidcResponseData(t, reduced)
	info, err := service.BuildUserInfo("Bearer " + token.AccessToken)
	require.NoError(t, err)
	assert.Contains(t, info, "name")
	assert.NotContains(t, info, "email")
	for _, scope := range []string{"openid email profile", "profile", "openid admin"} {
		invalid := oidcGrantRequest(t, http.MethodPatch, path, fmt.Sprintf(`{"scope":%q}`, scope), session.AccessToken)
		var result struct {
			Success bool `json:"success"`
		}
		require.NoError(t, common.Unmarshal(invalid.Body.Bytes(), &result))
		assert.False(t, result.Success, scope)
	}
	other := &model.User{Username: "oidc-grant-other", AffCode: "oidc-go", Status: common.UserStatusEnabled, Role: common.RoleCommonUser, AuthVersion: 1}
	require.NoError(t, model.DB.Create(other).Error)
	require.NoError(t, model.PublishUserAuthCache(other.Id))
	otherSession, err := service.CreateLoginSession(other.Id, "password", "127.0.0.3", "another-owner")
	require.NoError(t, err)
	otherList := oidcGrantRequest(t, http.MethodGet, "/api/user/oidc/grants", "", otherSession.AccessToken)
	require.NoError(t, common.Unmarshal(otherList.Body.Bytes(), &envelope))
	assert.Empty(t, envelope.Data, "grant lists are owner scoped")

	for _, method := range []string{http.MethodPatch, http.MethodDelete} {
		response := oidcGrantRequest(t, method, path, `{"scope":"openid"}`, otherSession.AccessToken)
		var result struct {
			Success bool `json:"success"`
		}
		require.NoError(t, common.Unmarshal(response.Body.Bytes(), &result))
		assert.False(t, result.Success, "another user's grant cannot be changed")
	}
	require.NoError(t, model.DB.Model(&model.OIDCAccessToken{}).Where("authorization_code_id = ?", grant.Id).Update("expires_at", time.Now().Add(-time.Minute)).Error)
	list = oidcGrantRequest(t, http.MethodGet, "/api/user/oidc/grants", "", session.AccessToken)
	require.NoError(t, common.Unmarshal(list.Body.Bytes(), &envelope))
	assert.Empty(t, envelope.Data)
	_, err = service.BuildUserInfo("Bearer " + token.AccessToken)
	require.Error(t, err)
	// Revocation is still owner permitted after expiry, including across sessions.
	oidcResponseData(t, oidcGrantRequest(t, http.MethodDelete, path, "", session.AccessToken))
	_, err = service.ExchangeAuthorizationCode(input)
	require.Error(t, err)
	pending, err := service.IssueAuthorizationCode(service.AuthorizationCodeInput{Identity: identity, AuthTime: time.Now().Unix(), ClientId: client.ClientId, UserId: user.Id, RedirectURI: req.RedirectURI, Scope: req.Scope, CodeChallenge: req.CodeChallenge, CodeChallengeMethod: req.CodeChallengeMethod})
	require.NoError(t, err)
	pendingGrant, err := model.GetOIDCAuthorizationCodeByHash(service.HashOIDCToken(pending))
	require.NoError(t, err)
	oidcResponseData(t, oidcGrantRequest(t, http.MethodDelete, fmt.Sprintf("/api/user/oidc/grants/%d", pendingGrant.Id), "", session.AccessToken))
	input.Code = pending
	_, err = service.ExchangeAuthorizationCode(input)
	require.Error(t, err)
	// A consumed, still live reference token is invalidated immediately too.
	fresh, err := service.IssueAuthorizationCode(service.AuthorizationCodeInput{Identity: identity, AuthTime: time.Now().Unix(), ClientId: client.ClientId, UserId: user.Id, RedirectURI: req.RedirectURI, Scope: req.Scope, CodeChallenge: req.CodeChallenge, CodeChallengeMethod: req.CodeChallengeMethod})
	require.NoError(t, err)
	freshGrant, err := model.GetOIDCAuthorizationCodeByHash(service.HashOIDCToken(fresh))
	require.NoError(t, err)
	input.Code = fresh
	live, err := service.ExchangeAuthorizationCode(input)
	require.NoError(t, err)
	oidcResponseData(t, oidcGrantRequest(t, http.MethodDelete, fmt.Sprintf("/api/user/oidc/grants/%d", freshGrant.Id), "", session.AccessToken))
	_, err = service.BuildUserInfo("Bearer " + live.AccessToken)
	require.Error(t, err)
}

func TestOIDCGrantScopeReductionPreventsStaleExchange(t *testing.T) {
	user, identity, client, secret, verifier, req := setupOIDCProviderTest(t)
	req.Scope = "openid profile email"
	code, err := service.IssueAuthorizationCode(service.AuthorizationCodeInput{Identity: identity, AuthTime: time.Now().Unix(), ClientId: client.ClientId, UserId: user.Id, RedirectURI: req.RedirectURI, Scope: req.Scope, CodeChallenge: req.CodeChallenge, CodeChallengeMethod: req.CodeChallengeMethod})
	require.NoError(t, err)
	grant, err := model.GetOIDCAuthorizationCodeByHash(service.HashOIDCToken(code))
	require.NoError(t, err)
	_, err = service.BuildJWKS()
	require.NoError(t, err)
	entered, resume := make(chan struct{}), make(chan struct{})
	var once sync.Once
	require.NoError(t, model.DB.Callback().Query().Before("gorm:query").Register("test:oidc-stale-scope", func(tx *gorm.DB) {
		if tx.Statement.Table == "oidc_signing_keys" {
			once.Do(func() { close(entered); <-resume })
		}
	}))
	t.Cleanup(func() { _ = model.DB.Callback().Query().Remove("test:oidc-stale-scope") })
	input := service.TokenExchangeInput{ClientId: client.ClientId, ClientSecret: secret, Code: code, RedirectURI: req.RedirectURI, CodeVerifier: verifier}
	result := make(chan error, 1)
	go func() { _, err := service.ExchangeAuthorizationCode(input); result <- err }()
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("exchange did not reach signing after reading its authorization scope")
	}
	err = service.UpdateOIDCGrantScopes(identity, grant.Id, "openid")
	close(resume)
	require.NoError(t, err)
	select {
	case err := <-result:
		require.Error(t, err, "stale scopes cannot be committed")
	case <-time.After(5 * time.Second):
		t.Fatal("exchange did not finish")
	}
	token, err := service.ExchangeAuthorizationCode(input)
	require.NoError(t, err)
	assert.Equal(t, "openid", token.Scope)
	info, err := service.BuildUserInfo("Bearer " + token.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"sub": service.OIDCSubject(user.Id)}, info)
}
