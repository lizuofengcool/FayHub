package middleware

import (
	"bytes"
	"encoding/json"
	"fayhub/pkg/response"
	"fayhub/pkg/sanitizer"
	"fmt"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
)

func InputSanitizationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if sanitizer.HasPathTraversal(path) {
			response.GinError(c, 40003, "请求路径包含非法字符")
			c.Abort()
			return
		}

		for key, values := range c.Request.URL.Query() {
			for _, v := range values {
				if sanitizer.HasSQLInjection(v) {
					response.GinError(c, 40003, fmt.Sprintf("查询参数 '%s' 包含非法内容", key))
					c.Abort()
					return
				}
				if sanitizer.HasXSS(v) {
					response.GinError(c, 40003, fmt.Sprintf("查询参数 '%s' 包含非法内容", key))
					c.Abort()
					return
				}
			}
		}

		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			contentType := c.GetHeader("Content-Type")
			if strings.Contains(contentType, "application/json") && !strings.HasPrefix(path, "/api/plugin-data/") && !strings.HasPrefix(path, "/api/backups/") {
				bodyBytes, err := io.ReadAll(io.LimitReader(c.Request.Body, 64*1024))
				if err != nil {
					c.Next()
					return
				}
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				var bodyMap map[string]interface{}
				if err := json.Unmarshal(bodyBytes, &bodyMap); err == nil {
					if checkMapForInjection(bodyMap) {
						response.GinError(c, 40003, "请求体包含非法内容")
						c.Abort()
						return
					}
				}
			}
		}

		c.Next()
	}
}

func checkMapForInjection(data map[string]interface{}) bool {
	for _, v := range data {
		switch val := v.(type) {
		case string:
			if sanitizer.HasSQLInjection(val) || sanitizer.HasXSS(val) {
				return true
			}
		case map[string]interface{}:
			if checkMapForInjection(val) {
				return true
			}
		}
	}
	return false
}
