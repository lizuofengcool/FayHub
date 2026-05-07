package pluginsign

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"sync"
)

var (
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
	keyMu      sync.Mutex
	keyLoaded  bool
)

func InitPublicKey(publicKeyPath string) error {
	keyMu.Lock()
	defer keyMu.Unlock()

	if keyLoaded {
		return nil
	}

	if publicKeyPath == "" {
		return fmt.Errorf("公钥路径为空")
	}

	data, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return fmt.Errorf("读取公钥文件失败: %w", err)
	}

	return initPublicKeyFromBytes(data)
}

func InitPublicKeyFromBytes(pemData []byte) error {
	keyMu.Lock()
	defer keyMu.Unlock()

	if keyLoaded {
		return nil
	}

	return initPublicKeyFromBytes(pemData)
}

func initPublicKeyFromBytes(data []byte) error {
	block, _ := pem.Decode(data)
	if block == nil {
		return fmt.Errorf("解析PEM公钥失败")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("解析公钥失败: %w", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("公钥类型不是RSA")
	}

	publicKey = rsaPub
	keyLoaded = true
	return nil
}

func InitPrivateKey(privateKeyPath string) error {
	data, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return fmt.Errorf("读取私钥文件失败: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return fmt.Errorf("解析PEM私钥失败")
	}

	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		priv, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("解析私钥失败: %w", err)
		}
	}

	rsaPriv, ok := priv.(*rsa.PrivateKey)
	if !ok {
		return fmt.Errorf("私钥类型不是RSA")
	}

	privateKey = rsaPriv
	return nil
}

func Sign(data []byte) (string, error) {
	if privateKey == nil {
		return "", fmt.Errorf("私钥未初始化")
	}

	hash := sha256.Sum256(data)
	signature, err := rsa.SignPKCS1v15(nil, privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", fmt.Errorf("签名失败: %w", err)
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

func Verify(data []byte, signatureBase64 string) error {
	if publicKey == nil {
		return fmt.Errorf("公钥未初始化，跳过签名校验")
	}

	signature, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return fmt.Errorf("签名Base64解码失败: %w", err)
	}

	hash := sha256.Sum256(data)
	if err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash[:], signature); err != nil {
		return fmt.Errorf("签名校验失败: %w", err)
	}

	return nil
}

func IsInitialized() bool {
	return publicKey != nil
}

func VerifyPlugin(wasmBytes []byte, signature string) error {
	if !IsInitialized() {
		return nil
	}

	if signature == "" {
		return fmt.Errorf("插件缺少签名")
	}

	return Verify(wasmBytes, signature)
}

func VerifyPluginHash(wasmBytes []byte, expectedHash string) error {
	if expectedHash == "" {
		return nil
	}

	hash := sha256.Sum256(wasmBytes)
	actualHash := fmt.Sprintf("%x", hash)

	if actualHash != expectedHash {
		return fmt.Errorf("插件完整性校验失败: 期望 %s, 实际 %s", expectedHash, actualHash)
	}

	return nil
}
