package router

import (
	"fayhub/internal/controller"
	"fayhub/internal/middleware"

	"github.com/gin-gonic/gin"
)

type TenantRouter struct{}

func (s *TenantRouter) Init(router *gin.Engine) {
	tenantGroup := router.Group("/api/tenants")
	tenantGroup.Use(middleware.TenantMiddleware())
	tenantGroup.Use(middleware.JwtAuthMiddleware())
	{
		tenantGroup.POST("", controller.ControllerGroupApp.TenantController.CreateTenant)
		tenantGroup.GET("", controller.ControllerGroupApp.TenantController.GetTenantList)
		tenantGroup.GET("/:id", controller.ControllerGroupApp.TenantController.GetTenant)
		tenantGroup.PUT("/:id", controller.ControllerGroupApp.TenantController.UpdateTenant)
		tenantGroup.DELETE("/:id", controller.ControllerGroupApp.TenantController.DeleteTenant)
	}
}
