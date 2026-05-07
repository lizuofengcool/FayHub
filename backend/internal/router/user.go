package router

import (
	"fayhub/internal/controller"
	"fayhub/internal/middleware"

	"github.com/gin-gonic/gin"
)

type UserRouter struct{}

func (s *UserRouter) Init(router *gin.Engine) {
	userGroup := router.Group("/api/users")
	userGroup.Use(middleware.JwtAuthMiddleware())
	userGroup.Use(middleware.TenantMiddleware())
	{
		userGroup.POST("", middleware.OperLog("用户管理", "新增用户"), controller.ControllerGroupApp.UserController.CreateUser)
		userGroup.GET("", controller.ControllerGroupApp.UserController.GetUserList)
		userGroup.GET("/profile", controller.ControllerGroupApp.UserController.GetProfile)
		userGroup.PUT("/change-password", middleware.OperLog("用户管理", "修改密码"), controller.ControllerGroupApp.UserController.ChangePassword)
		userGroup.GET("/:id", controller.ControllerGroupApp.UserController.GetUser)
		userGroup.PUT("/:id", middleware.OperLog("用户管理", "编辑用户"), controller.ControllerGroupApp.UserController.UpdateUser)
		userGroup.DELETE("/:id", middleware.OperLog("用户管理", "删除用户"), controller.ControllerGroupApp.UserController.DeleteUser)
		userGroup.PUT("/:id/reset-password", middleware.SuperAdminMiddleware(), middleware.OperLog("用户管理", "重置密码"), controller.ControllerGroupApp.UserController.ResetPassword)
	}
}
