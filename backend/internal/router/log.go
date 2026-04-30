package router

import (
	"fayhub/internal/controller"
	"fayhub/internal/middleware"

	"github.com/gin-gonic/gin"
)

type LogRouter struct{}

func (r *LogRouter) Init(router *gin.Engine) {
	logGroup := router.Group("/api/logs")
	logGroup.Use(middleware.JwtAuthMiddleware())
	logGroup.Use(middleware.TenantMiddleware())
	{
		logGroup.GET("/tenant/:tenantId", controller.ControllerGroupApp.LogController.QueryTenantLogs)
		logGroup.GET("/tenant/:tenantId/count", controller.ControllerGroupApp.LogController.GetTenantLogCount)
	}
}
