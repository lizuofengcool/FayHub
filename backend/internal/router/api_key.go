package router

import (
	"fayhub/internal/controller"
	"fayhub/internal/middleware"

	"github.com/gin-gonic/gin"
)

type APIKeyRouter struct{}

func (r *APIKeyRouter) Init(router *gin.Engine) {
	apiKeyGroup := router.Group("/api/api-keys")
	apiKeyGroup.Use(middleware.JwtAuthMiddleware())
	apiKeyGroup.Use(middleware.TenantMiddleware())
	{
		apiKeyGroup.POST("", controller.ControllerGroupApp.APIKeyController.CreateAPIKey)
		apiKeyGroup.GET("", controller.ControllerGroupApp.APIKeyController.ListAPIKeys)
		apiKeyGroup.DELETE("/:id", controller.ControllerGroupApp.APIKeyController.DeleteAPIKey)
	}
}
