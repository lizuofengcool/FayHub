package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Response 统一响应结构体
type Response struct {
	Code      int         `json:"code"`
	Message   string      `json:"msg"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
	RequestID string      `json:"request_id"` // 全链路追踪ID
}

// Success 成功响应
func Success(data interface{}) *Response {
	return &Response{
		Code:      200,
		Message:   "操作成功",
		Data:      data,
		Timestamp: time.Now().Unix(),
		RequestID: generateRequestID(),
	}
}

// SuccessWithMessage 带自定义消息的成功响应
func SuccessWithMessage(message string, data interface{}) *Response {
	return &Response{
		Code:      200,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
		RequestID: generateRequestID(),
	}
}

// Error 错误响应
func Error(code int, message string) *Response {
	return &Response{
		Code:      code,
		Message:   message,
		Data:      nil,
		Timestamp: time.Now().Unix(),
		RequestID: generateRequestID(),
	}
}

// ErrorWithData 带数据的错误响应
func ErrorWithData(code int, message string, data interface{}) *Response {
	return &Response{
		Code:      code,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
		RequestID: generateRequestID(),
	}
}

// PageResponse 分页响应
func PageResponse(data interface{}, total int64, page, pageSize int) *Response {
	pageData := map[string]interface{}{
		"list":       data,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"total_page": calculateTotalPage(total, pageSize),
	}

	return Success(pageData)
}

// GinSuccess Gin框架成功响应
func GinSuccess(c *gin.Context, data interface{}) {
	requestID := getRequestID(c)
	c.JSON(200, &Response{
		Code:      200,
		Message:   "操作成功",
		Data:      data,
		Timestamp: time.Now().Unix(),
		RequestID: requestID,
	})
}

// GinSuccessWithMessage Gin框架带消息的成功响应
func GinSuccessWithMessage(c *gin.Context, message string, data interface{}) {
	requestID := getRequestID(c)
	c.JSON(200, &Response{
		Code:      200,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
		RequestID: requestID,
	})
}

func codeToHTTPStatus(code int) int {
	switch {
	case code == 200:
		return http.StatusOK
	case code >= 40000 && code < 40100:
		return http.StatusBadRequest
	case code >= 40100 && code < 41000:
		return http.StatusUnauthorized
	case code >= 41000 && code < 42000:
		return http.StatusUnauthorized
	case code >= 42000 && code < 42100:
		return http.StatusForbidden
	case code >= 42100 && code < 42200:
		return http.StatusForbidden
	case code >= 42200 && code < 43000:
		return http.StatusForbidden
	case code == 42900:
		return http.StatusTooManyRequests
	case code >= 43000 && code < 44000:
		return http.StatusBadRequest
	case code >= 44000 && code < 45000:
		return http.StatusBadRequest
	case code == 40400:
		return http.StatusNotFound
	case code == 40300:
		return http.StatusForbidden
	case code >= 50000 && code < 60000:
		return http.StatusInternalServerError
	default:
		return http.StatusBadRequest
	}
}

// GinError Gin框架错误响应
func GinError(c *gin.Context, code int, message string) {
	requestID := getRequestID(c)
	httpStatus := codeToHTTPStatus(code)
	c.JSON(httpStatus, &Response{
		Code:      code,
		Message:   message,
		Data:      nil,
		Timestamp: time.Now().Unix(),
		RequestID: requestID,
	})
}

// GinErrorWithData Gin框架带数据的错误响应
func GinErrorWithData(c *gin.Context, code int, message string, data interface{}) {
	requestID := getRequestID(c)
	httpStatus := codeToHTTPStatus(code)
	c.JSON(httpStatus, &Response{
		Code:      code,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
		RequestID: requestID,
	})
}

// GinPageResponse Gin框架分页响应
func GinPageResponse(c *gin.Context, data interface{}, total int64, page, pageSize int) {
	requestID := getRequestID(c)
	pageData := map[string]interface{}{
		"list":       data,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"total_page": calculateTotalPage(total, pageSize),
	}

	c.JSON(200, &Response{
		Code:      200,
		Message:   "操作成功",
		Data:      pageData,
		Timestamp: time.Now().Unix(),
		RequestID: requestID,
	})
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	return "req_" + uuid.New().String()
}

// getRequestID 从Gin上下文获取请求ID
func getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return generateRequestID()
}

// calculateTotalPage 计算总页数
func calculateTotalPage(total int64, pageSize int) int {
	if pageSize <= 0 {
		return 0
	}

	totalPage := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPage++
	}

	return totalPage
}
