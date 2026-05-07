package router

import (
	"fayhub/internal/controller"
	"fayhub/internal/middleware"

	"github.com/gin-gonic/gin"
)

type TenantChannelRouter struct{}

func (r *TenantChannelRouter) Init(router *gin.Engine) {
	channelGroup := router.Group("/api/tenant-channel")
	channelGroup.Use(middleware.JwtAuthMiddleware())
	channelGroup.Use(middleware.TenantMiddleware())
	{
		channelGroup.GET("/configs", controller.ControllerGroupApp.TenantChannelController.ListConfigs)
		channelGroup.GET("/configs/:id", controller.ControllerGroupApp.TenantChannelController.GetConfig)
		channelGroup.GET("/configs/type/:channel_type", controller.ControllerGroupApp.TenantChannelController.GetConfigByChannelType)
		channelGroup.POST("/configs", controller.ControllerGroupApp.TenantChannelController.CreateConfig)
		channelGroup.PUT("/configs/:id", controller.ControllerGroupApp.TenantChannelController.UpdateConfig)
		channelGroup.DELETE("/configs/:id", controller.ControllerGroupApp.TenantChannelController.DeleteConfig)

		channelGroup.GET("/bindings", controller.ControllerGroupApp.TenantChannelController.GetUserBindings)
		channelGroup.DELETE("/bindings/:id", controller.ControllerGroupApp.TenantChannelController.DeleteUserBinding)
	}
}
