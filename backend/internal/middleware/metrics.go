package middleware

import (
	"fayhub/pkg/metrics"
	"time"

	"github.com/gin-gonic/gin"
)

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		metrics.IncrementActiveRequests()
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		metrics.DecrementActiveRequests()
		metrics.RecordRequest(c.Request.Method, c.FullPath(), duration, c.Writer.Status())
	}
}
