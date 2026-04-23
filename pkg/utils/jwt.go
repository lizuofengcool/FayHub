package utils

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// JWT配置
var (
	JwtSecret = []byte("fayhub-secret-key-change-in-production") // JWT密钥（生产环境必须修改）
	TokenExpireDuration = time.Hour * 24 * 7 // Token过期时间（7天）
)

// CustomClaims 自定义Claims
type CustomClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.StandardClaims
}

// GenerateToken 生成JWT Token
// @Summary 生成JWT Token
// @Description 根据用户信息生成JWT Token
// @Tags JWT工具
// @Param userID uint true "用户ID"
// @Param username string true "用户名"
// @Param role string true "用户角色"
// @Return string "JWT Token"
// @Return error "错误信息"
func GenerateToken(userID uint, username, role string) (string, error) {
	// 设置Token过期时间
	expireTime := time.Now().Add(TokenExpireDuration)
	
	// 创建Claims
	claims := CustomClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(),
			Issuer:    "fayhub",
		},
	}

	// 创建Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	// 签名Token
	return token.SignedString(JwtSecret)
}

// ParseToken 解析JWT Token
// @Summary 解析JWT Token
// @Description 解析JWT Token并验证有效性
// @Tags JWT工具
// @Param token string true "JWT Token"
// @Return *CustomClaims "Claims信息"
// @Return error "错误信息"
func ParseToken(tokenString string) (*CustomClaims, error) {
	// 解析Token
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return JwtSecret, nil
	})
	
	if err != nil {
		return nil, err
	}
	
	// 验证Token
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}
	
	return nil, errors.New("无效的Token")
}

// RefreshToken 刷新JWT Token
// @Summary 刷新JWT Token
// @Description 刷新即将过期的JWT Token
// @Tags JWT工具
// @Param tokenString string true "原JWT Token"
// @Return string "新的JWT Token"
// @Return error "错误信息"
func RefreshToken(tokenString string) (string, error) {
	// 解析原Token
	claims, err := ParseToken(tokenString)
	if err != nil {
		return "", err
	}
	
	// 生成新Token
	return GenerateToken(claims.UserID, claims.Username, claims.Role)
}