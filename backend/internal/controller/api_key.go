package controller

import (
	"fayhub/internal/service"
	errs "fayhub/pkg/errors"
	"fayhub/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type APIKeyController struct{}

func (akc *APIKeyController) CreateAPIKey(c *gin.Context) {
	var req service.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, errs.ErrParamValidation, "参数错误")
		return
	}

	ctx := c.Request.Context()
	keyService := &service.APIKeyService{}
	result, err := keyService.CreateAPIKey(ctx, req)
	if err != nil {
		response.GinError(c, errs.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(c, "API密钥创建成功", result)
}

func (akc *APIKeyController) ListAPIKeys(c *gin.Context) {
	ctx := c.Request.Context()
	keyService := &service.APIKeyService{}
	keys, err := keyService.ListAPIKeys(ctx)
	if err != nil {
		response.GinError(c, errs.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccess(c, keys)
}

func (akc *APIKeyController) DeleteAPIKey(c *gin.Context) {
	keyID := c.Param("id")
	if keyID == "" {
		response.GinError(c, errs.ErrParamValidation, "缺少密钥ID")
		return
	}

	id, err := strconv.ParseUint(keyID, 10, 32)
	if err != nil {
		response.GinError(c, errs.ErrParamValidation, "密钥ID格式错误")
		return
	}

	ctx := c.Request.Context()
	keyService := &service.APIKeyService{}
	if err := keyService.DeleteAPIKey(ctx, uint(id)); err != nil {
		response.GinError(c, errs.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(c, "API密钥删除成功", nil)
}
