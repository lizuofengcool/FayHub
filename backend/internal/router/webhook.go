package router

import (
	"fayhub/internal/controller"
	"fayhub/internal/middleware"

	"github.com/gin-gonic/gin"
)

type WebhookRouter struct{}

func (r *WebhookRouter) Init(router *gin.Engine) {
	webhookGroup := router.Group("/api/webhooks")
	webhookGroup.Use(middleware.JwtAuthMiddleware())
	webhookGroup.Use(middleware.TenantMiddleware())
	webhookGroup.Use(middleware.RateLimitMiddleware("webhook"))
	{
		webhookGroup.POST("/subscriptions", controller.ControllerGroupApp.WebhookController.CreateSubscription)
		webhookGroup.GET("/subscriptions", controller.ControllerGroupApp.WebhookController.ListSubscriptions)
		webhookGroup.GET("/subscriptions/:id", controller.ControllerGroupApp.WebhookController.GetSubscription)
		webhookGroup.PUT("/subscriptions/:id", controller.ControllerGroupApp.WebhookController.UpdateSubscription)
		webhookGroup.DELETE("/subscriptions/:id", controller.ControllerGroupApp.WebhookController.DeleteSubscription)
		webhookGroup.POST("/subscriptions/:id/test", controller.ControllerGroupApp.WebhookController.TestDelivery)

		webhookGroup.GET("/deliveries", controller.ControllerGroupApp.WebhookController.ListDeliveries)
		webhookGroup.POST("/deliveries/:id/redeliver", controller.ControllerGroupApp.WebhookController.Redeliver)
		webhookGroup.GET("/deliveries/stats", controller.ControllerGroupApp.WebhookController.GetDeliveryStats)
	}
}
