package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
// @Summary 统一响应结构
// @Description 所有API接口返回的统一JSON格式
// @Tags 响应工具
type Response struct {
	Code    int         `json:"code"`    // 状态码
	Message string      `json:"message"` // 消息
	Data    interface{} `json:"data"`    // 数据
}

// 常用状态码
const (
	SuccessCode     = 200 // 成功
	BadRequestCode  = 400 // 请求参数错误
	UnauthorizedCode = 401 // 未授权
	ForbiddenCode   = 403 // 禁止访问
	NotFoundCode    = 404 // 资源不存在
	InternalErrorCode = 500 // 服务器内部错误
)

// Ok 成功响应
// @Summary 成功响应
// @Description 返回成功的统一响应
// @Tags 响应工具
// @Param c *gin.Context true "Gin上下文"
func Ok(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code:    SuccessCode,
		Message: "操作成功",
		Data:    nil,
	})
}

// OkWithMessage 带消息的成功响应
// @Summary 带消息的成功响应
// @Description 返回带自定义消息的成功响应
// @Tags 响应工具
// @Param c *gin.Context true "Gin上下文"
// @Param message string true "自定义消息"
func OkWithMessage(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    SuccessCode,
		Message: message,
		Data:    nil,
	})
}

// OkWithData 带数据的成功响应
// @Summary 带数据的成功响应
// @Description 返回带数据的成功响应
// @Tags 响应工具
// @Param c *gin.Context true "Gin上下文"
// @Param data interface{} true "响应数据"
func OkWithData(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    SuccessCode,
		Message: "操作成功",
		Data:    data,
	})
}

// OkWithDetailed 带详细信息的成功响应
// @Summary 带详细信息的成功响应
// @Description 返回带自定义消息和数据的成功响应
// @Tags 响应工具
// @Param c *gin.Context true "Gin上下文"
// @Param message string true "自定义消息"
// @Param data interface{} true "响应数据"
func OkWithDetailed(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    SuccessCode,
		Message: message,
		Data:    data,
	})
}

// Fail 失败响应
// @Summary 失败响应
// @Description 返回失败的统一响应
// @Tags 响应工具
// @Param c *gin.Context true "Gin上下文"
// @Param code int true "错误码"
func Fail(c *gin.Context, code int) {
	var message string
	switch code {
	case BadRequestCode:
		message = "请求参数错误"
	case UnauthorizedCode:
		message = "未授权访问"
	case ForbiddenCode:
		message = "禁止访问"
	case NotFoundCode:
		message = "资源不存在"
	case InternalErrorCode:
		message = "服务器内部错误"
	default:
		message = "操作失败"
	}

	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// FailWithMessage 带消息的失败响应
// @Summary 带消息的失败响应
// @Description 返回带自定义消息的失败响应
// @Tags 响应工具
// @Param c *gin.Context true "Gin上下文"
// @Param code int true "错误码"
// @Param message string true "自定义消息"
func FailWithMessage(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// ValidationError 参数验证错误响应
// @Summary 参数验证错误响应
// @Description 返回参数验证失败的错误响应
// @Tags 响应工具
// @Param c *gin.Context true "Gin上下文"
// @Param message string true "验证错误消息"
func ValidationError(c *gin.Context, message string) {
	FailWithMessage(c, BadRequestCode, message)
}