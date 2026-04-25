package router

import (
	"fayhub/internal/controller"
	"fayhub/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RBACRouter RBAC权限路由
// @Summary RBAC权限路由
// @Description 注册角色、权限相关的路由
// @Tags 路由管理
type RBACRouter struct{}

// Init 初始化RBAC路由
// @Summary 初始化RBAC路由
// @Description 注册角色管理、权限分配等接口
// @Tags 路由管理
func (s *RBACRouter) Init(router *gin.Engine) {
	// 创建RBAC API分组
	rbacGroup := router.Group("/api/rbac")

	// 所有RBAC接口都需要JWT认证
	rbacGroup.Use(middleware.JwtAuthMiddleware())

	// 角色管理接口
	rbacGroup.POST("/roles", middleware.SuperAdminMiddleware(), controller.ControllerGroupApp.RBACController.CreateRole)
	rbacGroup.POST("/assign-role", middleware.SuperAdminMiddleware(), controller.ControllerGroupApp.RBACController.AssignRoleToUser)
	rbacGroup.POST("/remove-role", middleware.SuperAdminMiddleware(), controller.ControllerGroupApp.RBACController.RemoveRoleFromUser)
	rbacGroup.GET("/users/:userID/roles", middleware.SuperAdminMiddleware(), controller.ControllerGroupApp.RBACController.GetUserRoles)
	rbacGroup.GET("/users/:userID/permissions", middleware.SuperAdminMiddleware(), controller.ControllerGroupApp.RBACController.GetUserPermissions)
}
