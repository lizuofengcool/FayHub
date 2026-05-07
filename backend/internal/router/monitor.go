package router

import (
	"fayhub/internal/controller"
	"fayhub/internal/middleware"

	"github.com/gin-gonic/gin"
)

type MonitorRouter struct{}

func (r *MonitorRouter) Init(router *gin.Engine) {
	monitorGroup := router.Group("/api/monitor")
	monitorGroup.Use(middleware.JwtAuthMiddleware())
	monitorGroup.Use(middleware.TenantMiddleware())
	{
		monitorGroup.GET("/health", controller.ControllerGroupApp.MonitorController.GetHealth)
		monitorGroup.GET("/metrics", controller.ControllerGroupApp.MonitorController.GetSystemMetrics)
		monitorGroup.GET("/metrics/prometheus", controller.ControllerGroupApp.MonitorController.GetPrometheusMetrics)
	}
}
