package middleware

import (
	"bytes"
	"encoding/json"
	"fayhub/pkg/utils"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupRouter() *gin.Engine {
	r := gin.New()
	return r
}

func TestJwtAuthMiddleware_NoAuthHeader(t *testing.T) {
	r := setupRouter()
	r.Use(JwtAuthMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("无Authorization头应返回401, got %d", w.Code)
	}
}

func TestJwtAuthMiddleware_InvalidFormat(t *testing.T) {
	r := setupRouter()
	r.Use(JwtAuthMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "InvalidFormat token123")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("无效Token格式应返回401, got %d", w.Code)
	}
}

func TestJwtAuthMiddleware_NoBearerPrefix(t *testing.T) {
	r := setupRouter()
	r.Use(JwtAuthMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "token123")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("无Bearer前缀应返回401, got %d", w.Code)
	}
}

func TestJwtAuthMiddleware_InvalidToken(t *testing.T) {
	r := setupRouter()
	r.Use(JwtAuthMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.string")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("无效Token应返回401, got %d", w.Code)
	}
}

func TestJwtAuthMiddleware_ValidToken(t *testing.T) {
	utils.InitJWTConfig("test-secret-key-for-middleware-test", 60, "fayhub-test")

	token, err := utils.GenerateToken(1, "admin", "admin", 1)
	if err != nil {
		t.Fatalf("生成Token失败: %v", err)
	}

	r := setupRouter()
	r.Use(JwtAuthMiddleware())
	r.GET("/test", func(c *gin.Context) {
		userID, _ := GetUserIDFromContext(c)
		username, _ := GetUsernameFromContext(c)
		role, _ := GetRoleFromContext(c)
		c.JSON(200, gin.H{
			"user_id":  userID,
			"username": username,
			"role":     role,
		})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("有效Token应返回200, got %d, body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["username"] != "admin" {
		t.Errorf("username不匹配: got %v", resp["username"])
	}
}

func TestInputSanitizationMiddleware_PathTraversal(t *testing.T) {
	r := setupRouter()
	r.Use(InputSanitizationMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/../../../etc/passwd", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("路径遍历应返回400, got %d", w.Code)
	}
}

func TestInputSanitizationMiddleware_SQLInjection(t *testing.T) {
	r := setupRouter()
	r.Use(InputSanitizationMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test?q=1 UNION SELECT * FROM users", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("SQL注入应返回400, got %d", w.Code)
	}
}

func TestInputSanitizationMiddleware_XSS(t *testing.T) {
	r := setupRouter()
	r.Use(InputSanitizationMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test?q=<script>alert(1)</script>", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("XSS应返回400, got %d", w.Code)
	}
}

func TestInputSanitizationMiddleware_CleanRequest(t *testing.T) {
	r := setupRouter()
	r.Use(InputSanitizationMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test?name=hello&age=25", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("正常请求应返回200, got %d", w.Code)
	}
}

func TestInputSanitizationMiddleware_BodyInjection(t *testing.T) {
	r := setupRouter()
	r.Use(InputSanitizationMiddleware())
	r.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	body := `{"name": "'; DROP TABLE users; --"}`
	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("请求体SQL注入应返回400, got %d", w.Code)
	}
}

func TestInputSanitizationMiddleware_BodyXSS(t *testing.T) {
	r := setupRouter()
	r.Use(InputSanitizationMiddleware())
	r.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	body := `{"content": "<script>document.cookie</script>"}`
	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("请求体XSS应返回400, got %d", w.Code)
	}
}

func TestInputSanitizationMiddleware_CleanBody(t *testing.T) {
	r := setupRouter()
	r.Use(InputSanitizationMiddleware())
	r.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	body := `{"name": "Alice", "age": 30}`
	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("正常请求体应返回200, got %d", w.Code)
	}
}
