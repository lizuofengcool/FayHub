package router

import (
	"fayhub/internal/controller"
	"fayhub/internal/middleware"

	"github.com/gin-gonic/gin"
)

type CronJobRouter struct{}

func (r *CronJobRouter) Init(router *gin.Engine) {
	ctrl := &controller.CronJobController{}

	group := router.Group("/api/cron-jobs")
	group.Use(middleware.JwtAuthMiddleware())
	group.Use(middleware.TenantMiddleware())
	{
		group.GET("", ctrl.List)
		group.GET("/:id", ctrl.GetByID)
		group.POST("", ctrl.Create)
		group.PUT("/:id", ctrl.Update)
		group.DELETE("/:id", ctrl.Delete)
		group.PUT("/:id/toggle", ctrl.ToggleStatus)
		group.POST("/:id/execute", ctrl.ExecuteOnce)
		group.GET("/:id/logs", ctrl.GetLogs)
	}
}
