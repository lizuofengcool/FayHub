package router

import (
	"fayhub/internal/controller"
	"fayhub/internal/middleware"

	"github.com/gin-gonic/gin"
)

type LoginLogRouter struct{}

func (r *LoginLogRouter) Init(router *gin.Engine) {
	loginLogGroup := router.Group("/api/login-logs")
	loginLogGroup.Use(middleware.JwtAuthMiddleware())
	loginLogGroup.Use(middleware.TenantMiddleware())
	{
		loginLogGroup.GET("", controller.ControllerGroupApp.LoginLogController.ListLoginLogs)
		loginLogGroup.POST("/cleanup", controller.ControllerGroupApp.LoginLogController.CleanupLoginLogs)
	}
}
