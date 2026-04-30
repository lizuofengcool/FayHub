package router

import (
	"fayhub/internal/controller"
	"fayhub/internal/middleware"

	"github.com/gin-gonic/gin"
)

type AuthRouter struct{}

func (s *AuthRouter) Init(router *gin.Engine) {
	authGroup := router.Group("/api/auth")

	loginLimiter := middleware.RateLimitMiddleware("login")
	authGroup.POST("/login", loginLimiter, controller.ControllerGroupApp.AuthController.Login)
	authGroup.POST("/register", loginLimiter, controller.ControllerGroupApp.AuthController.Register)
	authGroup.GET("/captcha", controller.ControllerGroupApp.CaptchaController.GetCaptcha)

	authGroup.Use(middleware.JwtAuthMiddleware())
	authGroup.Use(middleware.TenantMiddleware())
	{
		authGroup.POST("/logout", controller.ControllerGroupApp.AuthController.Logout)
		authGroup.POST("/refresh", controller.ControllerGroupApp.AuthController.RefreshToken)
		authGroup.GET("/me", controller.ControllerGroupApp.AuthController.GetCurrentUser)
	}
}
