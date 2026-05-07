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
		tenantGroup.POST("", middleware.OperLog("租户管理", "新增租户"), controller.ControllerGroupApp.TenantController.CreateTenant)
		tenantGroup.GET("", controller.ControllerGroupApp.TenantController.GetTenantList)
		tenantGroup.GET("/:id", controller.ControllerGroupApp.TenantController.GetTenant)
		tenantGroup.PUT("/:id", middleware.OperLog("租户管理", "编辑租户"), controller.ControllerGroupApp.TenantController.UpdateTenant)
		tenantGroup.POST("/:id/soft-delete", middleware.OperLog("租户管理", "移入回收站"), controller.ControllerGroupApp.TenantController.SoftDeleteTenant)
		tenantGroup.POST("/:id/restore", middleware.OperLog("租户管理", "恢复租户"), controller.ControllerGroupApp.TenantController.RestoreTenant)
		tenantGroup.DELETE("/:id/permanent", middleware.OperLog("租户管理", "永久删除租户"), controller.ControllerGroupApp.TenantController.PermanentDeleteTenant)
		tenantGroup.POST("/:id/impersonate", middleware.OperLog("租户管理", "模拟登录租户"), controller.ControllerGroupApp.TenantController.ImpersonateTenant)

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
