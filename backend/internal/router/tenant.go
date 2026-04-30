package router

import (
	"fayhub/internal/controller"
	"fayhub/internal/middleware"

	"github.com/gin-gonic/gin"
)

type TenantRouter struct{}

func (s *TenantRouter) Init(router *gin.Engine) {
	tenantGroup := router.Group("/api/tenants")
	tenantGroup.Use(middleware.JwtAuthMiddleware())
	tenantGroup.Use(middleware.TenantMiddleware())
	{
		tenantGroup.POST("", controller.ControllerGroupApp.TenantController.CreateTenant)
		tenantGroup.GET("", controller.ControllerGroupApp.TenantController.GetTenantList)
		tenantGroup.GET("/:id", controller.ControllerGroupApp.TenantController.GetTenant)
		tenantGroup.PUT("/:id", controller.ControllerGroupApp.TenantController.UpdateTenant)
		tenantGroup.DELETE("/:id", controller.ControllerGroupApp.TenantController.DeleteTenant)

		quotaGroup := tenantGroup.Group("/:id/quota")
		{
			quotaGroup.GET("", controller.ControllerGroupApp.TenantQuotaController.GetQuota)
			quotaGroup.PUT("", controller.ControllerGroupApp.TenantQuotaController.UpdateQuota)
			quotaGroup.GET("/check/users", controller.ControllerGroupApp.TenantQuotaController.CheckUserQuota)
			quotaGroup.GET("/check/storage", controller.ControllerGroupApp.TenantQuotaController.CheckStorageQuota)
			quotaGroup.GET("/check/plugins", controller.ControllerGroupApp.TenantQuotaController.CheckPluginQuota)
			quotaGroup.POST("/sync", controller.ControllerGroupApp.TenantQuotaController.SyncUsage)
		}
	}
}
