package middleware

import (
	"fayhub/pkg/config"
	"fayhub/pkg/domains"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowedOrigin := ""

		if config.GlobalConfig != nil {
			cfg := &config.GlobalConfig.Server

			allowedOrigins := domains.GetAllCORSOrigins()

			if cfg.CORSAllowAll {
				allowedOrigin = "*"
			} else if len(allowedOrigins) > 0 {
				for _, o := range allowedOrigins {
					if o == "*" {
						allowedOrigin = "*"
						break
					}
					if strings.EqualFold(o, origin) {
						allowedOrigin = origin
						break
					}
				}
			}

			if allowedOrigin == "" && origin != "" && cfg.Mode == "debug" {
				if strings.HasPrefix(origin, "http://localhost") || strings.HasPrefix(origin, "http://127.0.0.1") || strings.Contains(origin, ".fayhub.com") {
					allowedOrigin = origin
				}
			}

			if allowedOrigin != "" {
				c.Header("Access-Control-Allow-Origin", allowedOrigin)
			}

			if cfg.CORSAllowCredentials && allowedOrigin != "" && allowedOrigin != "*" {
				c.Header("Access-Control-Allow-Credentials", "true")
			}
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, Authorization, X-Request-ID")
		c.Header("Access-Control-Expose-Headers", "Content-Length, X-Request-ID")
		c.Header("Access-Control-Max-Age", "86400")

		if allowedOrigin != "" && allowedOrigin != "*" {
			c.Header("Vary", "Origin")
		}

		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=()")

		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
