package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fayhub/pkg/redisclient"
	"fayhub/pkg/response"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type idempotentCacheEntry struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

func IdempotencyMiddleware(ttl time.Duration) gin.HandlerFunc {
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}

	return func(c *gin.Context) {
		if c.Request.Method != http.MethodPost && c.Request.Method != http.MethodPut && c.Request.Method != http.MethodPatch {
			c.Next()
			return
		}

		idempotencyKey := c.GetHeader("Idempotency-Key")
		if idempotencyKey == "" {
			c.Next()
			return
		}

		if len(idempotencyKey) > 256 {
			response.GinError(c, 40000, "Idempotency-Key 长度不能超过256字符")
			c.Abort()
			return
		}

		hash := sha256.Sum256([]byte(idempotencyKey))
		cacheKey := fmt.Sprintf("idempotent:%s", hex.EncodeToString(hash[:]))

		rdb := redisclient.GetRawClient()
		if rdb != nil {
			existing, err := rdb.Get(c.Request.Context(), cacheKey).Result()
			if err == nil && existing != "" {
				var entry idempotentCacheEntry
				if json.Unmarshal([]byte(existing), &entry) == nil {
					for k, v := range entry.Headers {
						c.Header(k, v)
					}
					c.Header("X-Idempotency-Replayed", "true")
					c.String(entry.StatusCode, entry.Body)
					c.Abort()
					return
				}
			}
		}

		writer := &idempotentResponseWriter{
			ResponseWriter: c.Writer,
			body:           new(bytes.Buffer),
		}
		c.Writer = writer

		c.Next()

		if writer.statusCode == 0 {
			writer.statusCode = 200
		}

		if isCacheableStatus(writer.statusCode) {
			entry := idempotentCacheEntry{
				StatusCode: writer.statusCode,
				Headers:    make(map[string]string),
				Body:       writer.body.String(),
			}

			for k, v := range writer.Header() {
				if len(v) > 0 {
					entry.Headers[k] = v[0]
				}
			}

			entryJSON, err := json.Marshal(entry)
			if err == nil {
				if rdb != nil {
					rdb.Set(c.Request.Context(), cacheKey, entryJSON, ttl)
				}
			}
		}
	}
}

type idempotentResponseWriter struct {
	gin.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (w *idempotentResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *idempotentResponseWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func (w *idempotentResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func isCacheableStatus(code int) bool {
	return code >= 200 && code < 300
}

func IdempotencyRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodPost && c.Request.Method != http.MethodPut && c.Request.Method != http.MethodPatch {
			c.Next()
			return
		}

		idempotencyKey := c.GetHeader("Idempotency-Key")
		if idempotencyKey == "" {
			response.GinError(c, 40000, "缺少 Idempotency-Key 请求头，请生成唯一幂等键后重试")
			c.Abort()
			return
		}

		c.Next()
	}
}

func GenerateIdempotencyKey(parts ...string) string {
	combined := strings.Join(parts, ":")
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])[:32]
}
