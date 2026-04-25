package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"fayhub/pkg/errors"
	"fayhub/pkg/logger"
	"fayhub/pkg/response"
)

// BusinessError 业务异常
type BusinessError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error 实现error接口
func (e *BusinessError) Error() string {
	return fmt.Sprintf("业务异常[%d]: %s", e.Code, e.Message)
}

// NewBusinessError 创建业务异常
func NewBusinessError(code int, message string) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
	}
}

// NewBusinessErrorWithData 创建带数据的业务异常
func NewBusinessErrorWithData(code int, message string, data interface{}) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// GlobalExceptionMiddleware 全局异常处理中间件
func GlobalExceptionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				handleException(c, err)
			}
		}()
		
		c.Next()
		
		// 处理业务层抛出的错误
		if len(c.Errors) > 0 {
			for _, ginErr := range c.Errors {
				handleGinError(c, ginErr.Err)
			}
		}
	}
}

// handleException 处理异常
func handleException(c *gin.Context, err interface{}) {
	ctx := c.Request.Context()
	
	// 记录异常堆栈
	stack := string(debug.Stack())
	logger.Error(ctx, "全局异常捕获", 
		zap.Any("error", err),
		zap.String("stack", stack),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
	)
	
	// 根据异常类型返回不同的错误响应
	switch e := err.(type) {
	case *BusinessError:
		// 业务异常
		if e.Data != nil {
			response.GinErrorWithData(c, e.Code, e.Message, e.Data)
		} else {
			response.GinError(c, e.Code, e.Message)
		}
		
	case error:
		// 标准错误
		response.GinError(c, errors.ErrInternalServer, "服务器内部错误")
		
	default:
		// 其他类型异常
		response.GinError(c, errors.ErrInternalServer, "未知异常")
	}
	
	// 终止请求处理
	c.Abort()
}

// handleGinError 处理Gin框架错误
func handleGinError(c *gin.Context, err error) {
	ctx := c.Request.Context()
	
	// 记录错误日志
	logger.Error(ctx, "Gin框架错误", 
		zap.Error(err),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
	)
	
	// 根据错误类型返回不同的响应
	switch e := err.(type) {
	case *BusinessError:
		// 业务异常
		if e.Data != nil {
			response.GinErrorWithData(c, e.Code, e.Message, e.Data)
		} else {
			response.GinError(c, e.Code, e.Message)
		}
		
	default:
		// 其他错误
		response.GinError(c, errors.ErrInternalServer, "服务器内部错误")
	}
}

// RecoveryMiddleware 恢复中间件（兼容旧版本Gin）
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			logger.Error(c.Request.Context(), "Panic recovered", 
				zap.String("error", err),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
			)
		}
		
		response.GinError(c, errors.ErrInternalServer, "服务器内部错误")
	})
}

// NotFoundHandler 404处理
func NotFoundHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		response.GinError(c, errors.ErrResourceNotFound, "请求的资源不存在")
	}
}

// MethodNotAllowedHandler 405处理
func MethodNotAllowedHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		response.GinError(c, errors.ErrOperationFailed, "请求方法不被允许")
	}
}

// ValidationErrorHandler 参数验证错误处理
func ValidationErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取第一个验证错误
		if len(c.Errors) > 0 {
			err := c.Errors[0]
			if err.Type == gin.ErrorTypeBind {
				response.GinError(c, errors.ErrParamValidation, "参数验证失败: "+err.Error())
				return
			}
		}
		
		response.GinError(c, errors.ErrParamValidation, "参数验证失败")
	}
}

// ErrorHandler 统一错误处理
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		
		// 检查HTTP状态码
		if c.Writer.Status() >= http.StatusBadRequest {
			switch c.Writer.Status() {
			case http.StatusBadRequest:
				response.GinError(c, errors.ErrParamValidation, "请求参数错误")
			case http.StatusUnauthorized:
				response.GinError(c, errors.ErrUnauthorized, "未授权访问")
			case http.StatusForbidden:
				response.GinError(c, errors.ErrPermissionDenied, "权限不足")
			case http.StatusNotFound:
				response.GinError(c, errors.ErrResourceNotFound, "资源不存在")
			case http.StatusMethodNotAllowed:
				response.GinError(c, errors.ErrOperationFailed, "请求方法不被允许")
			case http.StatusInternalServerError:
				response.GinError(c, errors.ErrInternalServer, "服务器内部错误")
			default:
				response.GinError(c, errors.ErrOperationFailed, "请求处理失败")
			}
		}
	}
}