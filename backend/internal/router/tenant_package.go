package router

import (
	"fayhub/internal/controller"
	"fayhub/internal/middleware"

	"github.com/gin-gonic/gin"
)

type TenantPackageRouter struct{}

func (s *TenantPackageRouter) Init(router *gin.Engine) {
	group := router.Group("/api/tenant-packages")
	group.Use(middleware.JwtAuthMiddleware())
	group.Use(middleware.TenantMiddleware())
	{
		group.GET("", controller.TenantPackageControllerApp.List)
		group.GET("/all", controller.TenantPackageControllerApp.GetAll)
		group.GET("/:id", controller.TenantPackageControllerApp.GetByID)
		group.POST("", middleware.SuperAdminMiddleware(), middleware.OperLog("套餐管理", "新增套餐"), controller.TenantPackageControllerApp.Create)
		group.PUT("/:id", middleware.SuperAdminMiddleware(), middleware.OperLog("套餐管理", "编辑套餐"), controller.TenantPackageControllerApp.Update)
		group.DELETE("/:id", middleware.SuperAdminMiddleware(), middleware.OperLog("套餐管理", "删除套餐"), controller.TenantPackageControllerApp.Delete)
	}
}
