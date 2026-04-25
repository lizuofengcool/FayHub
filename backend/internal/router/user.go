package router

import (
	"fayhub/internal/controller"
	"fayhub/internal/middleware"

	"github.com/gin-gonic/gin"
)

type UserRouter struct{}

func (s *UserRouter) Init(router *gin.Engine) {
	userGroup := router.Group("/api/users")
	userGroup.Use(middleware.TenantMiddleware())
	userGroup.Use(middleware.JwtAuthMiddleware())
	{
		userGroup.POST("", controller.ControllerGroupApp.UserController.CreateUser)
		userGroup.GET("", controller.ControllerGroupApp.UserController.GetUserList)
		userGroup.GET("/:id", controller.ControllerGroupApp.UserController.GetUser)
		userGroup.PUT("/:id", controller.ControllerGroupApp.UserController.UpdateUser)
		userGroup.DELETE("/:id", controller.ControllerGroupApp.UserController.DeleteUser)
	}
}
