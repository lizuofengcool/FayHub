package router

import (
	"fayhub/internal/controller"
	"fayhub/internal/middleware"

	"github.com/gin-gonic/gin"
)

type PaymentRouter struct{}

func (r *PaymentRouter) Init(router *gin.Engine) {
	paymentGroup := router.Group("/api/payment")
	paymentGroup.Use(middleware.JwtAuthMiddleware())
	paymentGroup.Use(middleware.TenantMiddleware())
	{
		paymentGroup.GET("/config", controller.ControllerGroupApp.PaymentController.GetConfig)
		paymentGroup.PUT("/config", controller.ControllerGroupApp.PaymentController.UpdateConfig)
		paymentGroup.POST("/orders", controller.ControllerGroupApp.PaymentController.CreateOrder)
		paymentGroup.GET("/transactions", controller.ControllerGroupApp.PaymentController.ListTransactions)
		paymentGroup.GET("/transactions/stats", controller.ControllerGroupApp.PaymentController.GetStats)
		paymentGroup.POST("/refund", controller.ControllerGroupApp.PaymentController.Refund)
	}

	notifyGroup := router.Group("/api/payment/notify")
	{
		notifyGroup.POST("/wechat", controller.ControllerGroupApp.PaymentController.WechatNotify)
		notifyGroup.POST("/alipay", controller.ControllerGroupApp.PaymentController.AlipayNotify)
	}
}
