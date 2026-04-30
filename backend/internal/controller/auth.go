package controller

import (
	"context"
	"fayhub/internal/middleware"
	"fayhub/internal/service"
	"fayhub/pkg/config"
	"fayhub/pkg/errors"
	"fayhub/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthController struct{}

func setTokenCookie(ctx *gin.Context, token string) {
	cfg := getConfig()
	maxAge := 86400
	if cfg != nil && cfg.JWT.Expire > 0 {
		maxAge = cfg.JWT.Expire
	}

	secure := false
	sameSite := http.SameSiteLaxMode
	if ctx.Request.TLS != nil {
		secure = true
		sameSite = http.SameSiteStrictMode
	}

	ctx.SetCookie("fayhub_token", token, maxAge, "/", "", secure, true)
	ctx.SetSameSite(sameSite)
}

func clearTokenCookie(ctx *gin.Context) {
	ctx.SetCookie("fayhub_token", "", -1, "/", "", false, true)
}

func getConfig() *config.Config {
	if config.GlobalConfig != nil {
		return config.GlobalConfig
	}
	return nil
}

// Login godoc
// @Summary 用户登录
// @Description 使用用户名和密码进行登录，返回JWT令牌
// @Tags 认证管理
// @Accept json
// @Produce json
// @Param body body service.LoginRequest true "登录请求参数"
// @Success 200 {object} response.Response "登录成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "认证失败"
// @Router /api/auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	var req service.LoginRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "请求参数错误")
		return
	}

	resp, err := service.ServiceGroupApp.AuthService.Login(ctx.Request.Context(), req)
	if err != nil {
		response.GinError(ctx, errors.ErrUnauthorized, err.Error())
		return
	}

	setTokenCookie(ctx, resp.Token)

	response.GinSuccessWithMessage(ctx, "登录成功", resp)
}

// Logout godoc
// @Summary 用户登出
// @Description 退出当前登录状态
// @Tags 认证管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response "登出成功"
// @Failure 500 {object} response.Response "登出失败"
// @Router /api/auth/logout [post]
func (c *AuthController) Logout(ctx *gin.Context) {
	tokenString, _ := middleware.GetTokenString(ctx)

	reqCtx := ctx.Request.Context()
	reqCtx = context.WithValue(reqCtx, "token_string", tokenString)

	err := service.ServiceGroupApp.AuthService.Logout(reqCtx)
	if err != nil {
		response.GinError(ctx, errors.ErrInternalServer, "登出失败")
		return
	}

	clearTokenCookie(ctx)

	response.GinSuccessWithMessage(ctx, "登出成功", nil)
}

// RefreshToken godoc
// @Summary 刷新令牌
// @Description 使用当前令牌刷新获取新的JWT令牌
// @Tags 认证管理
// @Accept json
// @Produce json
// @Param body body service.RefreshTokenRequest true "刷新令牌请求"
// @Success 200 {object} response.Response "刷新成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "令牌无效"
// @Router /api/auth/refresh [post]
func (c *AuthController) RefreshToken(ctx *gin.Context) {
	var req service.RefreshTokenRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "请求参数错误")
		return
	}

	newToken, err := service.ServiceGroupApp.AuthService.RefreshToken(ctx.Request.Context(), req.Token)
	if err != nil {
		response.GinError(ctx, errors.ErrUnauthorized, err.Error())
		return
	}

	setTokenCookie(ctx, newToken)

	response.GinSuccessWithMessage(ctx, "Token刷新成功", gin.H{
		"token": newToken,
	})
}

// GetCurrentUser godoc
// @Summary 获取当前用户信息
// @Description 获取当前登录用户的详细信息
// @Tags 认证管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response "获取成功"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/auth/me [get]
func (c *AuthController) GetCurrentUser(ctx *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(ctx)
	if !exists {
		response.GinError(ctx, errors.ErrUnauthorized, "未获取到用户信息")
		return
	}

	user, err := service.ServiceGroupApp.AuthService.GetCurrentUser(ctx.Request.Context(), userID)
	if err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "获取成功", user)
}

// Register godoc
// @Summary 用户注册
// @Description 注册新用户账号
// @Tags 认证管理
// @Accept json
// @Produce json
// @Param body body service.RegisterRequest true "注册请求参数"
// @Success 200 {object} response.Response "注册成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 409 {object} response.Response "用户名已存在"
// @Router /api/auth/register [post]
func (c *AuthController) Register(ctx *gin.Context) {
	var req service.RegisterRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "请求参数错误")
		return
	}

	resp, err := service.ServiceGroupApp.AuthService.Register(ctx.Request.Context(), req)
	if err != nil {
		response.GinError(ctx, errors.ErrConflict, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "注册成功", resp)
}
