package router

import (
	"fayhub/internal/controller"
	"fayhub/internal/middleware"

	"github.com/gin-gonic/gin"
)

type OnlineUserRouter struct{}

func (r *OnlineUserRouter) Init(router *gin.Engine) {
	ctrl := &controller.OnlineUserController{}

	onlineUserGroup := router.Group("/api/online-users")
	onlineUserGroup.Use(middleware.JwtAuthMiddleware())
	onlineUserGroup.Use(middleware.TenantMiddleware())
	{
		onlineUserGroup.GET("", ctrl.GetOnlineUsers)
		onlineUserGroup.POST("/force-logout", ctrl.ForceLogout)
		onlineUserGroup.GET("/count", ctrl.GetOnlineCount)
	}
}
