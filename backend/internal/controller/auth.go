package controller

import (
	"fayhub/internal/middleware"
	"fayhub/internal/service"
	"fayhub/pkg/errors"
	"fayhub/pkg/response"

	"github.com/gin-gonic/gin"
)

// AuthController 认证控制器
// @Summary 认证控制器
// @Description 处理用户认证相关的HTTP请求
// @Tags 认证接口
type AuthController struct{}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录接口，验证用户名密码并返回JWT Token
// @Tags 认证接口
// @Accept json
// @Produce json
// @Param loginRequest body service.LoginRequest true "登录请求"
// @Success 200 {object} map[string]interface{} "登录成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "用户名或密码错误"
// @Router /api/auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	var req service.LoginRequest

	// 绑定请求参数
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.GinError(ctx, errors.ParamValidationError.Code, "请求参数错误")
		return
	}

	// 调用Service层登录逻辑
	resp, err := service.ServiceGroupApp.AuthService.Login(ctx.Request.Context(), req)
	if err != nil {
		response.GinError(ctx, errors.UnauthorizedError.Code, err.Error())
		return
	}

	// 返回登录成功响应
	response.GinSuccessWithMessage(ctx, "登录成功", resp)
}

// Logout 用户登出
// @Summary 用户登出
// @Description 用户登出接口，清除认证信息
// @Tags 认证接口
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{} "登出成功"
// @Router /api/auth/logout [post]
func (c *AuthController) Logout(ctx *gin.Context) {
	// 调用Service层登出逻辑
	err := service.ServiceGroupApp.AuthService.Logout(ctx.Request.Context())
	if err != nil {
		response.GinError(ctx, errors.InternalServerError.Code, "登出失败")
		return
	}

	// 返回登出成功响应
	response.GinSuccessWithMessage(ctx, "登出成功", nil)
}

// RefreshToken 刷新Token
// @Summary 刷新Token
// @Description 刷新即将过期的JWT Token
// @Tags 认证接口
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param refreshTokenRequest body map[string]string true "刷新Token请求"
// @Success 200 {object} map[string]interface{} "刷新成功"
// @Failure 401 {object} map[string]interface{} "Token无效"
// @Router /api/auth/refresh [post]
func (c *AuthController) RefreshToken(ctx *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}

	// 绑定请求参数
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.GinError(ctx, errors.ParamValidationError.Code, "请求参数错误")
		return
	}

	// 调用Service层刷新Token逻辑
	newToken, err := service.ServiceGroupApp.AuthService.RefreshToken(ctx.Request.Context(), req.Token)
	if err != nil {
		response.GinError(ctx, errors.UnauthorizedError.Code, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "Token刷新成功", gin.H{
		"token": newToken,
	})
}

// GetCurrentUser 获取当前用户信息
// @Summary 获取当前用户信息
// @Description 获取当前登录用户的信息
// @Tags 认证接口
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{} "获取成功"
// @Router /api/auth/me [get]
func (c *AuthController) GetCurrentUser(ctx *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(ctx)
	if !exists {
		response.GinError(ctx, errors.UnauthorizedError.Code, "未获取到用户信息")
		return
	}

	user, err := service.ServiceGroupApp.AuthService.GetCurrentUser(ctx.Request.Context(), userID)
	if err != nil {
		response.GinError(ctx, errors.InternalServerError.Code, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "获取成功", user)
}
