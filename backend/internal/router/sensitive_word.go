package router

import (
	"fayhub/internal/controller"
	"fayhub/internal/middleware"

	"github.com/gin-gonic/gin"
)

type SensitiveWordRouter struct{}

func (r *SensitiveWordRouter) Init(router *gin.Engine) {
	group := router.Group("/api/sensitive-words")
	group.Use(middleware.JwtAuthMiddleware())
	group.Use(middleware.PermissionMiddleware("sensitive_word:list"))
	{
		group.GET("", controller.ControllerGroupApp.SensitiveWordController.List)
		group.POST("", controller.ControllerGroupApp.SensitiveWordController.Create)
		group.PUT("/:id", controller.ControllerGroupApp.SensitiveWordController.Update)
		group.DELETE("/:id", controller.ControllerGroupApp.SensitiveWordController.Delete)
		group.POST("/batch", controller.ControllerGroupApp.SensitiveWordController.BatchCreate)
		group.POST("/rebuild", controller.ControllerGroupApp.SensitiveWordController.Rebuild)
		group.POST("/check", controller.ControllerGroupApp.SensitiveWordController.Check)
	}
}
