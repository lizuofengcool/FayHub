package router

import (
	"fayhub/internal/controller"
	"fayhub/internal/middleware"

	"github.com/gin-gonic/gin"
)

type AuditRouter struct{}

func (r *AuditRouter) Init(router *gin.Engine) {
	auditGroup := router.Group("/api/audit")
	auditGroup.Use(middleware.JwtAuthMiddleware())
	auditGroup.Use(middleware.TenantMiddleware())
	{
		auditGroup.GET("/logs", controller.ControllerGroupApp.AuditController.ListAuditLogs)
		auditGroup.GET("/logs/:id", controller.ControllerGroupApp.AuditController.GetAuditLog)
		auditGroup.GET("/stats", controller.ControllerGroupApp.AuditController.GetAuditStats)
		auditGroup.GET("/export", controller.ControllerGroupApp.AuditController.ExportAuditLogs)
		auditGroup.POST("/cleanup", controller.ControllerGroupApp.AuditController.CleanupAuditLogs)
	}
}
