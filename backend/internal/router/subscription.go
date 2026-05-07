package router

import (
	"fayhub/internal/controller"
	"fayhub/internal/middleware"

	"github.com/gin-gonic/gin"
)

type SubscriptionRouter struct{}

func (r *SubscriptionRouter) Init(router *gin.Engine) {
	ctrl := &controller.SubscriptionController{}

	group := router.Group("/api/subscriptions")
	group.Use(middleware.JwtAuthMiddleware())
	group.Use(middleware.TenantMiddleware())
	{
		group.GET("", ctrl.List)
		group.GET("/:id", ctrl.GetByID)
		group.POST("", ctrl.Create)
		group.PUT("/:id", ctrl.Update)
		group.DELETE("/:id", ctrl.Delete)
		group.POST("/:id/cancel", ctrl.Cancel)
		group.POST("/:id/renew", ctrl.Renew)
		group.GET("/:id/invoices", ctrl.GetInvoices)
	}
}
