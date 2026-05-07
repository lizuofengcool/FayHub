package router

import (
	"fayhub/internal/controller"
	"fayhub/internal/middleware"

	"github.com/gin-gonic/gin"
)

type NotificationChannelRouter struct{}

func (r *NotificationChannelRouter) Init(router *gin.Engine) {
	ctrl := &controller.NotificationChannelController{}

	channelGroup := router.Group("/api/notification-channels")
	channelGroup.Use(middleware.JwtAuthMiddleware())
	channelGroup.Use(middleware.TenantMiddleware())
	{
		channelGroup.GET("", ctrl.ListChannels)
		channelGroup.GET("/:id", ctrl.GetChannel)
		channelGroup.POST("", ctrl.CreateChannel)
		channelGroup.PUT("/:id", ctrl.UpdateChannel)
		channelGroup.DELETE("/:id", ctrl.DeleteChannel)
	}

	templateGroup := router.Group("/api/notification-templates")
	templateGroup.Use(middleware.JwtAuthMiddleware())
	templateGroup.Use(middleware.TenantMiddleware())
	{
		templateGroup.GET("", ctrl.ListTemplates)
		templateGroup.POST("", ctrl.CreateTemplate)
		templateGroup.PUT("/:id", ctrl.UpdateTemplate)
		templateGroup.DELETE("/:id", ctrl.DeleteTemplate)
	}

	recordGroup := router.Group("/api/notification-records")
	recordGroup.Use(middleware.JwtAuthMiddleware())
	recordGroup.Use(middleware.TenantMiddleware())
	{
		recordGroup.GET("", ctrl.GetRecords)
		recordGroup.POST("/send", ctrl.Send)
	}
}
