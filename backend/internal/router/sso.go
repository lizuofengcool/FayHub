// Copyright (c) 2026 FayHub Team
// SPDX-License-Identifier: MIT

package router

import (
	"fayhub/internal/controller"
	"fayhub/internal/middleware"

	"github.com/gin-gonic/gin"
)

type SSORouter struct{}

func (r *SSORouter) Init(router *gin.Engine) {
	ssoGroup := router.Group("/api/sso")
	{
		// 获取授权码（需要登录）
		ssoGroup.GET("/authorize", middleware.JwtAuthMiddleware(), controller.ControllerGroupApp.SSOController.GetAuthorizationCode)
		
		// 授权码换令牌（市场调用，不需要JWT）
		ssoGroup.POST("/token", controller.ControllerGroupApp.SSOController.ExchangeToken)
		
		// 验证令牌（市场调用，不需要JWT）
		ssoGroup.POST("/verify", controller.ControllerGroupApp.SSOController.VerifyToken)
	}
}