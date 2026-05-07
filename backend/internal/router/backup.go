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
		backupLimiter := middleware.RateLimitMiddleware("backup")
		backupGroup.POST("", backupLimiter, controller.ControllerGroupApp.BackupController.CreateBackup)
		backupGroup.GET("", controller.ControllerGroupApp.BackupController.ListBackups)
		backupGroup.DELETE("/:id", backupLimiter, controller.ControllerGroupApp.BackupController.DeleteBackup)
		backupGroup.GET("/:id/download", controller.ControllerGroupApp.BackupController.DownloadBackup)
		backupGroup.POST("/restore", backupLimiter, controller.ControllerGroupApp.BackupController.RestoreBackup)
		backupGroup.POST("/restore/:id", backupLimiter, controller.ControllerGroupApp.BackupController.RestoreBackupByID)
		backupGroup.GET("/tables", controller.ControllerGroupApp.BackupController.ListTables)
		backupGroup.POST("/tables", backupLimiter, controller.ControllerGroupApp.BackupController.CreateBackupForTables)
		backupGroup.POST("/execute", backupLimiter, controller.ControllerGroupApp.BackupController.ExecuteSQL)
		backupGroup.GET("/processes", controller.ControllerGroupApp.BackupController.ListProcesses)
		backupGroup.DELETE("/processes/:pid", backupLimiter, controller.ControllerGroupApp.BackupController.KillProcess)
		backupGroup.GET("/verify", controller.ControllerGroupApp.BackupController.VerifyFields)
		backupGroup.POST("/replace", backupLimiter, controller.ControllerGroupApp.BackupController.ReplaceData)
		backupGroup.GET("/export", controller.ControllerGroupApp.BackupController.ExportTable)
		backupGroup.POST("/import", backupLimiter, controller.ControllerGroupApp.BackupController.ImportData)
		backupGroup.GET("/tables/:name/fields", controller.ControllerGroupApp.BackupController.GetTableFields)
		backupGroup.GET("/tables/:name/count", controller.ControllerGroupApp.BackupController.GetTableCount)
		backupGroup.GET("/tables/:name/preview", controller.ControllerGroupApp.BackupController.PreviewTable)
		backupGroup.POST("/replace/advanced", backupLimiter, controller.ControllerGroupApp.BackupController.AdvancedReplace)
		backupGroup.POST("/transfer", backupLimiter, controller.ControllerGroupApp.BackupController.DataTransfer)
		backupGroup.GET("/export/advanced", controller.ControllerGroupApp.BackupController.AdvancedExport)
		backupGroup.POST("/sql/write", backupLimiter, controller.ControllerGroupApp.BackupController.ExecuteWriteSQL)
		backupGroup.POST("/:id/notes", backupLimiter, controller.ControllerGroupApp.BackupController.UpdateBackupNotes)
	}
}
