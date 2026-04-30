package router

import (
	"fayhub/internal/controller"
	"fayhub/internal/middleware"

	"github.com/gin-gonic/gin"
)

type FileRouter struct{}

func (r *FileRouter) Init(router *gin.Engine) {
	fileGroup := router.Group("/api/files")
	fileGroup.Use(middleware.JwtAuthMiddleware())
	fileGroup.Use(middleware.TenantMiddleware())
	{
		fileGroup.POST("/upload", controller.ControllerGroupApp.FileController.Upload)
		fileGroup.GET("/list", controller.ControllerGroupApp.FileController.List)
		fileGroup.GET("/:id", controller.ControllerGroupApp.FileController.Download)
		fileGroup.DELETE("/:id", controller.ControllerGroupApp.FileController.Delete)
	}
}
