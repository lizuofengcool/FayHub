package router

import (
	"fayhub/internal/controller"
	"fayhub/internal/middleware"

	"github.com/gin-gonic/gin"
)

type ExcelRouter struct{}

func (r *ExcelRouter) Init(router *gin.Engine) {
	excelGroup := router.Group("/api/excel")
	excelGroup.Use(middleware.JwtAuthMiddleware())
	excelGroup.Use(middleware.TenantMiddleware())
	{
		excelGroup.POST("/import", controller.ControllerGroupApp.ExcelController.Import)
		excelGroup.POST("/preview", controller.ControllerGroupApp.ExcelController.Preview)
		excelGroup.GET("/template", controller.ControllerGroupApp.ExcelController.DownloadTemplate)
		excelGroup.GET("/export", controller.ControllerGroupApp.ExcelController.ExportGeneric)
		excelGroup.POST("/export", controller.ControllerGroupApp.ExcelController.ExportData)
	}
}
