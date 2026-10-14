package controller

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Stripe checkout is priced in StripeUnitPrice units, so a tier order must be
// refused instead of silently charging the untiered amount.
func TestStripePayRejectsSpecialTier(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/api/user/stripe/pay", RequestStripePay)

	body := bytes.NewBufferString(`{"amount":10,"payment_method":"stripe","tier_id":"tier_bonus_5"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/stripe/pay", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var res struct {
		Message string `json:"message"`
		Data    string `json:"data"`
	}
	require.NoError(t, common.Unmarshal(w.Body.Bytes(), &res))
	assert.Equal(t, "error", res.Message)
	assert.Contains(t, res.Data, "Special tiers")
}
