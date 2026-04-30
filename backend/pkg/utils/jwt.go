package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	jwtSecret     []byte
	jwtExpire     time.Duration
	jwtIssuer     string
	jwtAlgorithm  string
	jwtPrivateKey *rsa.PrivateKey
	jwtPublicKey  *rsa.PublicKey
)

func InitJWTConfig(secret string, expireHours int, issuer string) {
	jwtSecret = []byte(secret)
	jwtExpire = time.Hour * time.Duration(expireHours)
	jwtIssuer = issuer
	jwtAlgorithm = "HS256"
}

func InitJWTConfigRS256(secret string, expireHours int, issuer string, algorithm string, privateKeyPath string, publicKeyPath string) error {
	jwtSecret = []byte(secret)
	jwtExpire = time.Hour * time.Duration(expireHours)
	jwtIssuer = issuer
	jwtAlgorithm = "HS256"

	if algorithm == "RS256" {
		if privateKeyPath == "" || publicKeyPath == "" {
			return fmt.Errorf("RS256模式需要配置private_key_path和public_key_path")
		}

		privData, err := os.ReadFile(privateKeyPath)
		if err != nil {
			return fmt.Errorf("读取私钥文件失败: %w", err)
		}

		privKey, err := jwt.ParseRSAPrivateKeyFromPEM(privData)
		if err != nil {
			return fmt.Errorf("解析私钥失败: %w", err)
		}

		pubData, err := os.ReadFile(publicKeyPath)
		if err != nil {
			return fmt.Errorf("读取公钥文件失败: %w", err)
		}

		pubKey, err := jwt.ParseRSAPublicKeyFromPEM(pubData)
		if err != nil {
			return fmt.Errorf("解析公钥失败: %w", err)
		}

		jwtPrivateKey = privKey
		jwtPublicKey = pubKey
		jwtAlgorithm = "RS256"
	}

	return nil
}

func GetJWTExpire() time.Duration {
	return jwtExpire
}

func GetJWTAlgorithm() string {
	return jwtAlgorithm
}

type CustomClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	TenantID uint   `json:"tenant_id"`
	jwt.RegisteredClaims
}

func GenerateToken(userID uint, username, role string, tenantID uint) (string, error) {
	expireTime := time.Now().Add(jwtExpire)

	claims := CustomClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		TenantID: tenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			Issuer:    jwtIssuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        fmt.Sprintf("%d-%d", userID, time.Now().UnixNano()),
		},
	}

	if jwtAlgorithm == "RS256" && jwtPrivateKey != nil {
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		return token.SignedString(jwtPrivateKey)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ParseToken(tokenString string) (*CustomClaims, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if jwtAlgorithm == "RS256" && jwtPublicKey != nil {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("非预期的签名方法: %v", token.Header["alg"])
			}
			return jwtPublicKey, nil
		}
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("非预期的签名方法: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	}

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, keyFunc)
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("无效的Token")
}

func RefreshToken(tokenString string) (string, error) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return "", err
	}

	return GenerateToken(claims.UserID, claims.Username, claims.Role, claims.TenantID)
}

func GenerateRSAKeyPair(privateKeyPath, publicKeyPath string) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("生成RSA密钥对失败: %w", err)
	}

	privFile, err := os.Create(privateKeyPath)
	if err != nil {
		return fmt.Errorf("创建私钥文件失败: %w", err)
	}
	defer privFile.Close()

	privPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	if err := pem.Encode(privFile, privPEM); err != nil {
		return fmt.Errorf("写入私钥文件失败: %w", err)
	}

	pubFile, err := os.Create(publicKeyPath)
	if err != nil {
		return fmt.Errorf("创建公钥文件失败: %w", err)
	}
	defer pubFile.Close()

	pubPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&privateKey.PublicKey),
	}
	if err := pem.Encode(pubFile, pubPEM); err != nil {
		return fmt.Errorf("写入公钥文件失败: %w", err)
	}

	return nil
}
