package router

import (
	"github.com/QuantumNous/new-api/controller"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/gin-gonic/gin"
)

func SetOIDCRouter(router *gin.Engine) {
	router.GET("/.well-known/openid-configuration", middleware.DisableCache(), controller.OIDCDiscovery)
	oauth := router.Group("/oauth", middleware.RouteTag("api"), middleware.DisableCache(), middleware.CriticalRateLimit(), middleware.BodyStorageCleanup())
	oauth.GET("/jwks", controller.OIDCJWKS)
	oauth.GET("/authorize", middleware.TryUserAuth(), controller.OIDCAuthorize)
	oauth.POST("/authorize", middleware.TryUserAuth(), middleware.AnonymousRequestBodyLimit(), controller.OIDCAuthorize)
	oauth.POST("/token", middleware.AnonymousRequestBodyLimit(), controller.OIDCToken)
	oauth.GET("/userinfo", controller.OIDCUserInfo)
}
