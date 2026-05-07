// Copyright (c) 2026 FayHub Team
// SPDX-License-Identifier: MIT

package controller

import (
	"fayhub/internal/service"
	"fayhub/pkg/config"
	"fayhub/pkg/errors"
	"fayhub/pkg/response"

	"github.com/gin-gonic/gin"
)

type SSOController struct{}

// GetAuthorizationCode godoc
// @Summary 获取授权码
// @Description 生成OAuth2.0授权码，用于跳转到市场
// @Tags SSO
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response "获取成功"
// @Router /api/sso/authorize [get]
func (sc *SSOController) GetAuthorizationCode(c *gin.Context) {
	ctx := c.Request.Context()

	code, err := service.ServiceGroupApp.SSOService.GenerateAuthorizationCode(ctx)
	if err != nil {
		response.GinError(c, errors.ErrUnauthorized, err.Error())
		return
	}

	marketURL := config.GlobalConfig.Domains.MarketURL + "/auth/callback"
	redirectURL := marketURL + "?code=" + code

	response.GinSuccess(c, gin.H{
		"code":         code,
		"redirect_url": redirectURL,
	})
}

// ExchangeToken godoc
// @Summary 授权码换令牌
// @Description 市场调用此接口用授权码换取用户信息
// @Tags SSO
// @Accept json
// @Produce json
// @Param request body service.SSOTokenExchangeRequest true "授权码"
// @Success 200 {object} response.Response "换取成功"
// @Failure 400 {object} response.Response "授权码无效"
// @Router /api/sso/token [post]
func (sc *SSOController) ExchangeToken(c *gin.Context) {
	var req service.SSOTokenExchangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, errors.ErrParamValidation, "参数错误")
		return
	}

	ctx := c.Request.Context()
	resp, err := service.ServiceGroupApp.SSOService.ExchangeToken(ctx, req.Code, req.ClientID, req.ClientSecret)
	if err != nil {
		response.GinError(c, errors.ErrParamValidation, err.Error())
		return
	}

	response.GinSuccess(c, resp)
}

// VerifyToken godoc
// @Summary 验证令牌
// @Description 市场调用此接口验证用户Token有效性
// @Tags SSO
// @Accept json
// @Produce json
// @Param request body object true "Token"
// @Success 200 {object} response.Response "验证成功"
// @Router /api/sso/verify [post]
func (sc *SSOController) VerifyToken(c *gin.Context) {
	var req struct {
		Token string `json:"token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, errors.ErrParamValidation, "参数错误")
		return
	}

	ctx := c.Request.Context()
	valid, err := service.ServiceGroupApp.SSOService.VerifyToken(ctx, req.Token, "", "")
	if err != nil {
		response.GinError(c, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccess(c, gin.H{
		"valid": valid,
	})
}
