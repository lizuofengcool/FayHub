package router

import (
	"fayhub/internal/controller"
	"fayhub/internal/middleware"

	"github.com/gin-gonic/gin"
)

type ErrorCodeRouter struct{}

func (r *ErrorCodeRouter) Init(router *gin.Engine) {
	ecGroup := router.Group("/api/error-codes")
	ecGroup.Use(middleware.JwtAuthMiddleware())
	ecGroup.Use(middleware.TenantMiddleware())
	{
		ecGroup.GET("", controller.ControllerGroupApp.ErrorCodeController.ListErrorCodes)
		ecGroup.POST("", controller.ControllerGroupApp.ErrorCodeController.CreateErrorCode)
		ecGroup.PUT("/:id", controller.ControllerGroupApp.ErrorCodeController.UpdateErrorCode)
		ecGroup.DELETE("/:id", controller.ControllerGroupApp.ErrorCodeController.DeleteErrorCode)
		ecGroup.POST("/refresh-cache", controller.ControllerGroupApp.ErrorCodeController.RefreshCache)
	}
}
