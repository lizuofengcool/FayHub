package router

import (
	"fayhub/internal/controller"
	"fayhub/internal/middleware"

	"github.com/gin-gonic/gin"
)

type SystemSettingRouter struct{}

func (r *SystemSettingRouter) Init(router *gin.Engine) {
	settingGroup := router.Group("/api/system/settings")
	settingGroup.Use(middleware.JwtAuthMiddleware())
	settingGroup.Use(middleware.TenantMiddleware())
	{
		settingGroup.GET("", controller.ControllerGroupApp.SystemSettingController.GetSettings)
		settingGroup.PUT("", controller.ControllerGroupApp.SystemSettingController.UpdateSettings)
	}
}
