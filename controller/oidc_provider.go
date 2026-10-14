package controller

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OIDCClientResponse struct {
	Id           int    `json:"id"`
	Name         string `json:"name"`
	ClientId     string `json:"client_id"`
	RedirectURIs string `json:"redirect_uris"`
	Scopes       string `json:"scopes"`
	Trusted      bool   `json:"trusted"`
	Enabled      bool   `json:"enabled"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`
}

type OIDCClientWithSecretResponse struct {
	*OIDCClientResponse
	ClientSecret string `json:"client_secret"`
}

type OIDCClientRequest struct {
	Name         string   `json:"name"`
	RedirectURIs []string `json:"redirect_uris"`
	Scopes       []string `json:"scopes"`
	Trusted      bool     `json:"trusted"`
	Enabled      bool     `json:"enabled"`
}

type oidcConsentState struct {
	Identity service.AuthIdentity     `json:"identity"`
	Request  service.AuthorizeRequest `json:"request"`
}

const oidcConsentPurpose = "oidc_consent"

func OIDCDiscovery(c *gin.Context) {
	issuer, err := service.OIDCIssuer()
	if err != nil {
		status, body := service.OIDCErrorResponse(err)
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, service.BuildOIDCDiscovery(issuer))
}

func OIDCJWKS(c *gin.Context) {
	jwks, err := service.BuildJWKS()
	if err != nil {
		status, body := service.OIDCErrorResponse(err)
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, jwks)
}

func OIDCAuthorize(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")
	c.Header("Referrer-Policy", "no-referrer")
	identity, ok := middleware.GetSessionAuthIdentity(c)
	if !ok {
		if c.Request.Method == http.MethodGet && !strings.Contains(c.GetHeader("Accept"), "application/json") {
			c.Redirect(http.StatusFound, "/oidc/authorize?"+c.Request.URL.RawQuery)
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "login_required", "error_description": "A browser sign-in is required"})
		return
	}
	if _, pat := model.ParseAccessTokenSessionID(identity.SessionID); pat {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	session, _, err := service.ValidateLoginSession(identity)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "login_required", "error_description": "The browser sign-in is no longer valid"})
		return
	}
	var req service.AuthorizeRequest
	if c.Request.Method == http.MethodPost {
		var input struct {
			ConsentToken string `json:"consent_token"`
			Decision     string `json:"decision"`
		}
		if c.ShouldBindJSON(&input) != nil || (input.Decision != "allow" && input.Decision != "deny") || input.ConsentToken == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "Invalid consent request"})
			return
		}
		var state oidcConsentState
		_, err := model.ConsumeAuthFlowWithAction(input.ConsentToken, model.AuthFlowMatch{Purpose: oidcConsentPurpose, UserId: identity.UserID, SessionId: identity.SessionID}, func(tx *gorm.DB, flow *model.AuthFlow) error {
			if common.UnmarshalJsonStr(flow.Payload, &state) != nil || state.Identity != identity {
				return model.ErrAuthFlowInvalid
			}
			return model.ValidateAuthSessionWithTx(tx, identity)
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "Consent has expired or was already used"})
			return
		}
		req = state.Request
		if input.Decision == "deny" {
			if _, _, err := service.ValidateAuthorizeRequest(req); err != nil {
				status, body := service.OIDCErrorResponse(err)
				c.JSON(status, body)
				return
			}
			redirectOIDCAuthorizeError(c, req.RedirectURI, req.State, &service.OIDCProtocolError{Code: "access_denied", Description: "The user denied consent", StatusCode: http.StatusForbidden})
			return
		}
	} else {
		query := c.Request.URL.Query()
		for _, key := range []string{"client_id", "redirect_uri", "response_type", "scope", "state", "nonce", "code_challenge", "code_challenge_method"} {
			if len(query[key]) > 1 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "Duplicate authorization parameter"})
				return
			}
		}
		req = service.AuthorizeRequest{ClientId: c.Query("client_id"), RedirectURI: c.Query("redirect_uri"), ResponseType: c.Query("response_type"), Scope: c.Query("scope"), State: c.Query("state"), Nonce: c.Query("nonce"), CodeChallenge: c.Query("code_challenge"), CodeChallengeMethod: c.Query("code_challenge_method")}
	}
	client, scope, err := service.ValidateAuthorizeRequest(req)
	if err != nil {
		status, body := service.OIDCErrorResponse(err)
		c.JSON(status, body)
		return
	}
	if !client.Trusted && c.Request.Method != http.MethodPost {
		payload, err := common.Marshal(oidcConsentState{Identity: identity, Request: req})
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		token, _, err := model.CreateAuthFlow(model.AuthFlowCreate{Purpose: oidcConsentPurpose, UserId: identity.UserID, SessionId: identity.SessionID, Payload: string(payload), ExpiresAt: time.Now().Add(5 * time.Minute)})
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		common.ApiSuccess(c, gin.H{"consent_token": token, "client_name": client.Name, "scope": scope, "expires_in": int(service.OIDCAccessTokenTTL.Seconds())})
		return
	}
	code, err := service.IssueAuthorizationCode(service.AuthorizationCodeInput{Identity: identity, AuthTime: session.CreatedAt, ClientId: client.ClientId, UserId: identity.UserID, RedirectURI: req.RedirectURI, Scope: scope, Nonce: req.Nonce, CodeChallenge: req.CodeChallenge, CodeChallengeMethod: req.CodeChallengeMethod})
	if err != nil {
		redirectOIDCAuthorizeError(c, req.RedirectURI, req.State, err)
		return
	}
	callback, _ := url.Parse(req.RedirectURI)
	query := callback.Query()
	query.Set("code", code)
	if req.State != "" {
		query.Set("state", req.State)
	}
	callback.RawQuery = query.Encode()
	if strings.Contains(c.GetHeader("Accept"), "application/json") {
		common.ApiSuccess(c, gin.H{"authorization_url": callback.String()})
		return
	}
	c.Redirect(http.StatusFound, callback.String())
}

func OIDCToken(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")
	if err := c.Request.ParseForm(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": "invalid form body",
		})
		return
	}
	for _, key := range []string{"grant_type", "client_id", "client_secret", "code", "redirect_uri", "code_verifier"} {
		if len(c.Request.PostForm[key]) > 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "Duplicate token request parameter"})
			return
		}
	}
	if _, _, basic := c.Request.BasicAuth(); basic && (c.PostForm("client_id") != "" || c.PostForm("client_secret") != "") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "Multiple client authentication methods are not allowed"})
		return
	}
	if c.PostForm("grant_type") != "authorization_code" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "unsupported_grant_type",
			"error_description": "grant_type must be authorization_code",
		})
		return
	}
	clientId, clientSecret := oidcClientCredentials(c)
	token, err := service.ExchangeAuthorizationCode(service.TokenExchangeInput{
		ClientId:     clientId,
		ClientSecret: clientSecret,
		Code:         c.PostForm("code"),
		RedirectURI:  c.PostForm("redirect_uri"),
		CodeVerifier: c.PostForm("code_verifier"),
	})
	if err != nil {
		status, body := service.OIDCErrorResponse(err)
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, token)
}

func OIDCUserInfo(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	info, err := service.BuildUserInfo(c.GetHeader("Authorization"))
	if err != nil {
		status, body := service.OIDCErrorResponse(err)
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, info)
}

func ListOIDCClients(c *gin.Context) {
	clients, err := model.ListOIDCClients()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	response := make([]*OIDCClientResponse, 0, len(clients))
	for _, client := range clients {
		response = append(response, toOIDCClientResponse(client))
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": response})
}

func GetOIDCClient(c *gin.Context) {
	client, ok := loadOIDCClientByParam(c)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": toOIDCClientResponse(client)})
}

func CreateOIDCClient(c *gin.Context) {
	var req OIDCClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorT(c, "Invalid parameters")
		return
	}
	authorization := middleware.RequireSecurityProof(c, service.VerificationOperation{Scope: service.VerificationScopeOIDCClientCreate})
	if authorization == nil {
		return
	}
	c.Header("Cache-Control", "no-store")
	client, secret, err := service.CreateOIDCClient(service.CreateOIDCClientInput{
		Name:         req.Name,
		RedirectURIs: req.RedirectURIs,
		Scopes:       req.Scopes,
		Trusted:      req.Trusted,
		Enabled:      req.Enabled,
	}, authorization)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	recordManageAuditFor(c, c.GetInt("id"), "oidc.client.create", map[string]any{"client_id": client.Id})
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": &OIDCClientWithSecretResponse{
			OIDCClientResponse: toOIDCClientResponse(client),
			ClientSecret:       secret,
		},
	})
}

func UpdateOIDCClient(c *gin.Context) {
	client, ok := loadOIDCClientByParam(c)
	if !ok {
		return
	}
	var req OIDCClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorT(c, "Invalid parameters")
		return
	}
	context, _ := common.Marshal(service.OIDCClientContext{ClientID: client.Id})
	authorization := middleware.RequireSecurityProof(c, service.VerificationOperation{Scope: service.VerificationScopeOIDCClientUpdate, Context: context})
	if authorization == nil {
		return
	}
	if err := service.UpdateOIDCClient(client, service.UpdateOIDCClientInput{
		Name:         req.Name,
		RedirectURIs: req.RedirectURIs,
		Scopes:       req.Scopes,
		Trusted:      req.Trusted,
		Enabled:      req.Enabled,
	}, authorization); err != nil {
		common.ApiError(c, err)
		return
	}
	recordManageAuditFor(c, c.GetInt("id"), "oidc.client.update", map[string]any{"client_id": client.Id})
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": toOIDCClientResponse(client)})
}

func DeleteOIDCClient(c *gin.Context) {
	client, ok := loadOIDCClientByParam(c)
	if !ok {
		return
	}
	context, _ := common.Marshal(service.OIDCClientContext{ClientID: client.Id})
	authorization := middleware.RequireSecurityProof(c, service.VerificationOperation{Scope: service.VerificationScopeOIDCClientDelete, Context: context})
	if authorization == nil {
		return
	}
	if err := model.DeleteOIDCClient(client.Id, authorization); err != nil {
		common.ApiError(c, err)
		return
	}
	recordManageAuditFor(c, c.GetInt("id"), "oidc.client.delete", map[string]any{"client_id": client.Id})
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": nil})
}

func RotateOIDCClientSecret(c *gin.Context) {
	client, ok := loadOIDCClientByParam(c)
	if !ok {
		return
	}
	context, _ := common.Marshal(service.OIDCClientContext{ClientID: client.Id})
	authorization := middleware.RequireSecurityProof(c, service.VerificationOperation{Scope: service.VerificationScopeOIDCClientSecretRotate, Context: context})
	if authorization == nil {
		return
	}
	c.Header("Cache-Control", "no-store")
	secret, err := service.RotateOIDCClientSecret(client, authorization)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	recordManageAuditFor(c, c.GetInt("id"), "oidc.client.secret.rotate", map[string]any{"client_id": client.Id})
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": &OIDCClientWithSecretResponse{
			OIDCClientResponse: toOIDCClientResponse(client),
			ClientSecret:       secret,
		},
	})
}

func toOIDCClientResponse(client *model.OIDCClient) *OIDCClientResponse {
	return &OIDCClientResponse{
		Id:           client.Id,
		Name:         client.Name,
		ClientId:     client.ClientId,
		RedirectURIs: client.RedirectURIs,
		Scopes:       client.Scopes,
		Trusted:      client.Trusted,
		Enabled:      client.Enabled,
		CreatedAt:    client.CreatedAt.Unix(),
		UpdatedAt:    client.UpdatedAt.Unix(),
	}
}

func loadOIDCClientByParam(c *gin.Context) (*model.OIDCClient, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		common.ApiErrorT(c, "Invalid ID")
		return nil, false
	}
	client, err := model.GetOIDCClientById(id)
	if err != nil {
		common.ApiErrorT(c, "OIDC client was not found")
		return nil, false
	}
	return client, true
}

func oidcClientCredentials(c *gin.Context) (string, string) {
	clientId, clientSecret, ok := c.Request.BasicAuth()
	if ok {
		// OAuth Basic credentials are form encoded before base64 encoding. Existing
		// generated credentials can contain literal '+' but never '%', so decode
		// percent escaped values without rewriting the previously supported raw form.
		credentials := []string{clientId, clientSecret}
		for i, value := range credentials {
			if strings.Contains(value, "%") {
				decoded, err := url.QueryUnescape(value)
				if err != nil {
					return "", ""
				}
				credentials[i] = decoded
			}
		}
		return credentials[0], credentials[1]
	}
	return c.PostForm("client_id"), c.PostForm("client_secret")
}

func redirectOIDCAuthorizeError(c *gin.Context, redirectURI string, state string, err error) {
	status, body := service.OIDCErrorResponse(err)
	if redirectURI == "" {
		c.JSON(status, body)
		return
	}
	callback, parseErr := url.Parse(redirectURI)
	if parseErr != nil || callback.Host == "" {
		c.JSON(status, body)
		return
	}
	query := callback.Query()
	if value, ok := body["error"].(string); ok {
		query.Set("error", value)
	}
	if value, ok := body["error_description"].(string); ok {
		query.Set("error_description", value)
	}
	if state != "" {
		query.Set("state", state)
	}
	callback.RawQuery = query.Encode()
	if strings.Contains(c.GetHeader("Accept"), "application/json") {
		common.ApiSuccess(c, gin.H{"authorization_url": callback.String()})
		return
	}
	c.Redirect(http.StatusFound, callback.String())
}

func ListOIDCGrants(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	identity, ok := middleware.GetAuthIdentity(c)
	if !ok || identity.SessionID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "A browser sign-in is required"})
		return
	}
	if _, pat := model.ParseAccessTokenSessionID(identity.SessionID); pat {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	grants, err := service.ListOIDCGrants(identity)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, grants)
}

func UpdateOIDCGrant(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	identity, ok := middleware.GetAuthIdentity(c)
	if !ok || identity.SessionID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "A browser sign-in is required"})
		return
	}
	if _, pat := model.ParseAccessTokenSessionID(identity.SessionID); pat {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		common.ApiErrorT(c, "Invalid ID")
		return
	}
	var input struct {
		Scope string `json:"scope"`
	}
	if c.ShouldBindJSON(&input) != nil {
		common.ApiErrorT(c, "Invalid parameters")
		return
	}
	if err := service.UpdateOIDCGrantScopes(identity, id, input.Scope); err != nil {
		if err == gorm.ErrRecordNotFound {
			common.ApiErrorT(c, "OIDC grant was not found")
		} else {
			common.ApiError(c, err)
		}
		return
	}
	recordUserSecurityAudit(c, identity.UserID, "oidc.grant.scope_reduce", map[string]any{"grant_id": id, "scope": service.NormalizeOIDCScope(input.Scope)})
	common.ApiSuccess(c, nil)
}

func RevokeOIDCGrant(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	identity, ok := middleware.GetAuthIdentity(c)
	if !ok || identity.SessionID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "A browser sign-in is required"})
		return
	}
	if _, pat := model.ParseAccessTokenSessionID(identity.SessionID); pat {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		common.ApiErrorT(c, "Invalid ID")
		return
	}
	if err := model.ChangeOIDCGrant(identity, id, nil); err != nil {
		if err == gorm.ErrRecordNotFound {
			common.ApiErrorT(c, "OIDC grant was not found")
		} else {
			common.ApiError(c, err)
		}
		return
	}
	recordUserSecurityAudit(c, identity.UserID, "oidc.grant.revoke", map[string]any{"grant_id": id})
	common.ApiSuccess(c, nil)
}
