package router

import (
	"fayhub/internal/controller"
	"fayhub/internal/middleware"

	"github.com/gin-gonic/gin"
)

type BackupRouter struct{}

func (r *BackupRouter) Init(router *gin.Engine) {
	backupGroup := router.Group("/api/backups")
	backupGroup.Use(middleware.JwtAuthMiddleware())
	backupGroup.Use(middleware.SuperAdminMiddleware())
	{
		backupGroup.POST("", controller.ControllerGroupApp.BackupController.CreateBackup)
		backupGroup.GET("", controller.ControllerGroupApp.BackupController.ListBackups)
		backupGroup.DELETE("/:id", controller.ControllerGroupApp.BackupController.DeleteBackup)
		backupGroup.GET("/:id/download", controller.ControllerGroupApp.BackupController.DownloadBackup)
		backupGroup.POST("/restore", controller.ControllerGroupApp.BackupController.RestoreBackup)
	}
}
