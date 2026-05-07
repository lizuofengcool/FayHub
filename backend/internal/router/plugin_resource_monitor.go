package router

import (
	"fayhub/internal/controller"
	"fayhub/internal/middleware"

	"github.com/gin-gonic/gin"
)

type PluginResourceMonitorRouter struct{}

func (r *PluginResourceMonitorRouter) Init(router *gin.Engine) {
	group := router.Group("/api/plugin-monitor")
	group.Use(middleware.JwtAuthMiddleware())
	group.Use(middleware.TenantMiddleware())
	{
		group.GET("/runtime", controller.ControllerGroupApp.PluginResourceMonitorController.GetRuntimeStats)
		group.GET("/db", controller.ControllerGroupApp.PluginResourceMonitorController.GetDBStats)
		group.GET("/alerts", controller.ControllerGroupApp.PluginResourceMonitorController.GetAlerts)
		group.POST("/reset", controller.ControllerGroupApp.PluginResourceMonitorController.ResetStats)
	}
}
