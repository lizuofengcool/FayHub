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
		userGroup.POST("", controller.ControllerGroupApp.UserController.CreateUser)
		userGroup.GET("", controller.ControllerGroupApp.UserController.GetUserList)
		userGroup.GET("/profile", controller.ControllerGroupApp.UserController.GetProfile)
		userGroup.PUT("/change-password", controller.ControllerGroupApp.UserController.ChangePassword)
		userGroup.GET("/:id", controller.ControllerGroupApp.UserController.GetUser)
		userGroup.PUT("/:id", controller.ControllerGroupApp.UserController.UpdateUser)
		userGroup.DELETE("/:id", controller.ControllerGroupApp.UserController.DeleteUser)
		userGroup.PUT("/:id/reset-password", middleware.SuperAdminMiddleware(), controller.ControllerGroupApp.UserController.ResetPassword)
	}
}
