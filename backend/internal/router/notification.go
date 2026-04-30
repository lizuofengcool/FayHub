package router

import (
	"fayhub/internal/controller"
	"fayhub/internal/middleware"

	"github.com/gin-gonic/gin"
)

type NotificationRouter struct{}

func (r *NotificationRouter) Init(router *gin.Engine) {
	notifyGroup := router.Group("/api/notifications")
	notifyGroup.Use(middleware.JwtAuthMiddleware())
	notifyGroup.Use(middleware.TenantMiddleware())
	{
		notifyGroup.GET("", controller.ControllerGroupApp.NotificationController.ListNotifications)
		notifyGroup.GET("/unread-count", controller.ControllerGroupApp.NotificationController.GetUnreadCount)
		notifyGroup.PUT("/read", controller.ControllerGroupApp.NotificationController.MarkAsRead)
		notifyGroup.PUT("/read-all", controller.ControllerGroupApp.NotificationController.MarkAllAsRead)
		notifyGroup.DELETE("", controller.ControllerGroupApp.NotificationController.DeleteNotifications)
		notifyGroup.POST("/send", controller.ControllerGroupApp.NotificationController.SendNotification)
	}
}
