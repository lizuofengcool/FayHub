package router

import (
	"fayhub/internal/controller"
	"fayhub/internal/middleware"

	"github.com/gin-gonic/gin"
)

type MenuRouter struct{}

func (s *MenuRouter) Init(router *gin.Engine) {
	menuGroup := router.Group("/api/menus")
	menuGroup.Use(middleware.JwtAuthMiddleware())
	menuGroup.Use(middleware.TenantMiddleware())

	menuGroup.POST("", middleware.SuperAdminMiddleware(), controller.ControllerGroupApp.MenuController.CreateMenu)
	menuGroup.GET("", controller.ControllerGroupApp.MenuController.GetMenuList)
	menuGroup.GET("/tree", controller.ControllerGroupApp.MenuController.GetMenuTree)
	menuGroup.GET("/:menuID", controller.ControllerGroupApp.MenuController.GetMenuByID)
	menuGroup.PUT("/:menuID", middleware.SuperAdminMiddleware(), controller.ControllerGroupApp.MenuController.UpdateMenu)
	menuGroup.DELETE("/:menuID", middleware.SuperAdminMiddleware(), controller.ControllerGroupApp.MenuController.DeleteMenu)

	menuGroup.POST("/assign-role", middleware.SuperAdminMiddleware(), controller.ControllerGroupApp.MenuController.AssignRoleMenus)
	menuGroup.GET("/role/:roleID", controller.ControllerGroupApp.MenuController.GetRoleMenus)
}
