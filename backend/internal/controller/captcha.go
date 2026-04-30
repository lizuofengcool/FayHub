package controller

import (
	"fayhub/internal/service"
	"fayhub/pkg/errors"
	"fayhub/pkg/response"

	"github.com/gin-gonic/gin"
)

type CaptchaController struct{}

type CaptchaRequest struct {
	Key   string `json:"key" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

func (c *CaptchaController) GetCaptcha(ctx *gin.Context) {
	key, code, err := service.ServiceGroupApp.CaptchaService.Generate(ctx.Request.Context())
	if err != nil {
		response.GinError(ctx, errors.ErrInternalServer, "生成验证码失败")
		return
	}

	response.GinSuccess(ctx, gin.H{
		"captcha_key":   key,
		"captcha_code":  code,
		"expires_in":    300,
	})
}
