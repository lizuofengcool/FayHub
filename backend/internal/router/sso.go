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
		ssoGroup.GET("/authorize", middleware.JwtAuthMiddleware(), controller.ControllerGroupApp.SSOController.GetAuthorizationCode)

		ssoLimiter := middleware.RateLimitMiddleware("sso")
		ssoGroup.POST("/token", ssoLimiter, controller.ControllerGroupApp.SSOController.ExchangeToken)
		ssoGroup.POST("/verify", ssoLimiter, controller.ControllerGroupApp.SSOController.VerifyToken)
	}
}